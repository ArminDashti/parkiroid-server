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
	db              *sql.DB
	retentionPeriod time.Duration
}

func OpenSQLite(databasePath string, retentionPeriod time.Duration) (*SQLiteStore, error) {
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	database.SetMaxOpenConns(1)

	if _, err := database.Exec(schemaDDL); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("apply sqlite schema: %w", err)
	}

	return &SQLiteStore{db: database, retentionPeriod: retentionPeriod}, nil
}

func (store *SQLiteStore) Close() error {
	return store.db.Close()
}

func (store *SQLiteStore) retentionCutoff() time.Time {
	return time.Now().UTC().Add(-store.retentionPeriod)
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
	cutoff := store.retentionCutoff().Format(time.RFC3339)

	err = store.db.QueryRow(
		`SELECT f.id, f.path, f.captured_at
		 FROM frames f
		 WHERE f.device_id = ? AND f.captured_at >= ?
		 ORDER BY f.captured_at DESC, f.id DESC
		 LIMIT 1`,
		deviceRowID,
		cutoff,
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

func (store *SQLiteStore) SaveMetrics(metrics models.DeviceMetricsRecord) error {
	deviceRowID, err := store.resolveDeviceRowID(metrics.DeviceID)
	if err != nil {
		return err
	}

	recordedAt := NormalizeRecordedAt(metrics.RecordedAt)
	receivedAt := metrics.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = time.Now().UTC()
	}

	_, err = store.db.Exec(
		`INSERT INTO device_metrics (
			device_id, cpu_usage, memory_usage, disk_usage,
			battery_level, temperature_c, signal_strength,
			recorded_at, received_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		deviceRowID,
		metrics.CPUUsage,
		metrics.MemoryUsage,
		metrics.DiskUsage,
		nullableFloat(metrics.BatteryLevel),
		nullableFloat(metrics.TemperatureC),
		nullableInt(metrics.SignalStrength),
		recordedAt.Format(time.RFC3339),
		receivedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("insert device metrics: %w", err)
	}

	return nil
}

func (store *SQLiteStore) GetLatestMetrics(deviceID string) (models.DeviceMetricsRecord, error) {
	deviceRowID, err := store.resolveDeviceRowID(deviceID)
	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return models.DeviceMetricsRecord{}, ErrMetricsNotFound
		}
		return models.DeviceMetricsRecord{}, err
	}

	var metrics models.DeviceMetricsRecord
	var recordedAtRaw string
	var receivedAtRaw string
	var batteryLevel sql.NullFloat64
	var temperatureC sql.NullFloat64
	var signalStrength sql.NullInt64
	cutoff := store.retentionCutoff().Format(time.RFC3339)

	err = store.db.QueryRow(
		`SELECT cpu_usage, memory_usage, disk_usage,
		        battery_level, temperature_c, signal_strength,
		        recorded_at, received_at
		 FROM device_metrics
		 WHERE device_id = ? AND recorded_at >= ?
		 ORDER BY recorded_at DESC, id DESC
		 LIMIT 1`,
		deviceRowID,
		cutoff,
	).Scan(
		&metrics.CPUUsage,
		&metrics.MemoryUsage,
		&metrics.DiskUsage,
		&batteryLevel,
		&temperatureC,
		&signalStrength,
		&recordedAtRaw,
		&receivedAtRaw,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.DeviceMetricsRecord{}, ErrMetricsNotFound
		}
		return models.DeviceMetricsRecord{}, fmt.Errorf("query latest metrics: %w", err)
	}

	metrics.DeviceID = deviceID
	metrics.BatteryLevel = floatPointer(batteryLevel)
	metrics.TemperatureC = floatPointer(temperatureC)
	metrics.SignalStrength = intPointer(signalStrength)

	metrics.RecordedAt, err = time.Parse(time.RFC3339, recordedAtRaw)
	if err != nil {
		return models.DeviceMetricsRecord{}, fmt.Errorf("parse recorded_at: %w", err)
	}

	metrics.ReceivedAt, err = time.Parse(time.RFC3339, receivedAtRaw)
	if err != nil {
		return models.DeviceMetricsRecord{}, fmt.Errorf("parse received_at: %w", err)
	}

	return metrics, nil
}

func (store *SQLiteStore) PurgeExpiredFrames(cutoff time.Time) ([]string, error) {
	cutoffRaw := cutoff.UTC().Format(time.RFC3339)

	rows, err := store.db.Query(
		`SELECT path FROM frames WHERE captured_at < ?`,
		cutoffRaw,
	)
	if err != nil {
		return nil, fmt.Errorf("query expired frames: %w", err)
	}
	defer rows.Close()

	framePaths := make([]string, 0)
	for rows.Next() {
		var framePath string
		if err := rows.Scan(&framePath); err != nil {
			return nil, fmt.Errorf("scan expired frame path: %w", err)
		}
		framePaths = append(framePaths, framePath)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expired frames: %w", err)
	}

	if _, err := store.db.Exec(`DELETE FROM frames WHERE captured_at < ?`, cutoffRaw); err != nil {
		return nil, fmt.Errorf("delete expired frames: %w", err)
	}

	return framePaths, nil
}

func (store *SQLiteStore) PurgeExpiredMetrics(cutoff time.Time) error {
	cutoffRaw := cutoff.UTC().Format(time.RFC3339)

	if _, err := store.db.Exec(`DELETE FROM device_metrics WHERE recorded_at < ?`, cutoffRaw); err != nil {
		return fmt.Errorf("delete expired metrics: %w", err)
	}

	return nil
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

func nullableFloat(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func floatPointer(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	result := value.Float64
	return &result
}

func intPointer(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	result := int(value.Int64)
	return &result
}
