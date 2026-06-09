package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/parkiroid/parkiroid-server/internal/models"
	_ "modernc.org/sqlite"
)

var ErrDeviceNotFound = errors.New("device not found")

type SQLiteStore struct {
	db *sql.DB
}

func OpenSQLite(databasePath string) (*SQLiteStore, error) {
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	database.SetMaxOpenConns(1)

	if _, err := database.Exec(schemaDDL); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("apply sqlite schema: %w", err)
	}

	return &SQLiteStore{db: database}, nil
}

func (store *SQLiteStore) Close() error {
	return store.db.Close()
}

func (store *SQLiteStore) SaveFrame(frame models.FrameRecord) error {
	deviceRowID, err := store.resolveDeviceRowID(frame.DeviceID)
	if err != nil {
		return err
	}

	capturedAt := NormalizeCapturedAt(frame.CapturedAt)
	_, err = store.db.Exec(
		`INSERT INTO frames (path, device_id, captured_at) VALUES (?, ?, ?)`,
		frame.Path,
		deviceRowID,
		capturedAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("insert frame: %w", err)
	}

	return nil
}

func (store *SQLiteStore) GetLastFrame(deviceID string) (models.FrameRecord, error) {
	deviceRowID, err := store.resolveDeviceRowID(deviceID)
	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return models.FrameRecord{}, ErrFrameNotFound
		}
		return models.FrameRecord{}, err
	}

	var frame models.FrameRecord
	var capturedAtRaw string

	err = store.db.QueryRow(
		`SELECT f.id, f.path, f.captured_at
		 FROM frames f
		 WHERE f.device_id = ?
		 ORDER BY f.captured_at DESC, f.id DESC
		 LIMIT 1`,
		deviceRowID,
	).Scan(&frame.ID, &frame.Path, &capturedAtRaw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.FrameRecord{}, ErrFrameNotFound
		}
		return models.FrameRecord{}, fmt.Errorf("query last frame: %w", err)
	}

	frame.DeviceID = deviceID
	frame.CapturedAt, err = time.Parse(time.RFC3339, capturedAtRaw)
	if err != nil {
		return models.FrameRecord{}, fmt.Errorf("parse captured_at: %w", err)
	}

	return frame, nil
}

func (store *SQLiteStore) UpsertDevice(device models.Device) (models.Device, error) {
	if device.MACAddress == "" {
		return models.Device{}, fmt.Errorf("mac_address is required")
	}
	if device.DeviceName == "" {
		device.DeviceName = device.MACAddress
	}

	result, err := store.db.Exec(
		`INSERT INTO devices (device_name, mac_address)
		 VALUES (?, ?)
		 ON CONFLICT(mac_address) DO UPDATE SET device_name = excluded.device_name`,
		device.DeviceName,
		device.MACAddress,
	)
	if err != nil {
		return models.Device{}, fmt.Errorf("upsert device: %w", err)
	}

	rowID, err := result.LastInsertId()
	if err != nil || rowID == 0 {
		return store.GetDeviceByMACAddress(device.MACAddress)
	}

	device.ID = rowID
	return device, nil
}

func (store *SQLiteStore) GetDeviceByMACAddress(macAddress string) (models.Device, error) {
	var device models.Device
	err := store.db.QueryRow(
		`SELECT id, device_name, mac_address FROM devices WHERE mac_address = ?`,
		macAddress,
	).Scan(&device.ID, &device.DeviceName, &device.MACAddress)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Device{}, ErrDeviceNotFound
		}
		return models.Device{}, fmt.Errorf("query device: %w", err)
	}

	return device, nil
}

func (store *SQLiteStore) ListDevices() ([]models.Device, error) {
	rows, err := store.db.Query(
		`SELECT id, device_name, mac_address FROM devices ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	devices := make([]models.Device, 0)
	for rows.Next() {
		var device models.Device
		if err := rows.Scan(&device.ID, &device.DeviceName, &device.MACAddress); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		devices = append(devices, device)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate devices: %w", err)
	}

	return devices, nil
}

func (store *SQLiteStore) resolveDeviceRowID(deviceIdentifier string) (int64, error) {
	var deviceRowID int64
	err := store.db.QueryRow(
		`SELECT id FROM devices
		 WHERE mac_address = ? OR device_name = ? OR CAST(id AS TEXT) = ?
		 LIMIT 1`,
		deviceIdentifier,
		deviceIdentifier,
		deviceIdentifier,
	).Scan(&deviceRowID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			insertedDevice, upsertErr := store.UpsertDevice(models.Device{
				DeviceName: deviceIdentifier,
				MACAddress: deviceIdentifier,
			})
			if upsertErr != nil {
				return 0, upsertErr
			}
			return insertedDevice.ID, nil
		}
		return 0, fmt.Errorf("resolve device: %w", err)
	}

	return deviceRowID, nil
}
