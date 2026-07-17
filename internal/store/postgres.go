package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dogan/dogan-server/internal/models"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStore struct {
	db              *sql.DB
	retentionPeriod time.Duration
}

func OpenPostgres(databaseURL string, retentionPeriod time.Duration) (*PostgresStore, error) {
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres database: %w", err)
	}

	database.SetMaxOpenConns(10)

	if err := database.Ping(); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping postgres database: %w", err)
	}

	if _, err := database.Exec(schemaDDL); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("apply postgres schema: %w", err)
	}

	if _, err := database.Exec(schemaMigrationDDL); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("apply postgres schema migrations: %w", err)
	}

	return &PostgresStore{db: database, retentionPeriod: retentionPeriod}, nil
}

func (store *PostgresStore) Close() error {
	return store.db.Close()
}

func (store *PostgresStore) retentionCutoff() time.Time {
	return time.Now().UTC().Add(-store.retentionPeriod)
}

func (store *PostgresStore) SaveLoginLog(ip, username, browserInfo string, success bool) error {
	_, err := store.db.Exec(
		`INSERT INTO login_logs (ip, username, browser_info, attempted_at, success)
		 VALUES ($1, $2, $3, $4, $5)`,
		ip,
		username,
		browserInfo,
		time.Now().UTC(),
		success,
	)
	if err != nil {
		return fmt.Errorf("insert login log: %w", err)
	}
	return nil
}

func (store *PostgresStore) SaveFrame(frame models.FrameRecord) error {
	deviceRowID, err := store.resolveDeviceRowID(frame.DeviceID)
	if err != nil {
		return err
	}

	capturedAt := NormalizeCapturedAt(frame.CapturedAt)
	receivedAt := frame.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = time.Now().UTC()
	}

	_, err = store.db.Exec(
		`INSERT INTO stored_images (device_id, path, captured_at, received_at)
		 VALUES ($1, $2, $3, $4)`,
		deviceRowID,
		frame.Path,
		capturedAt,
		receivedAt,
	)
	if err != nil {
		return fmt.Errorf("insert stored image: %w", err)
	}

	return nil
}

func (store *PostgresStore) GetLastFrame(deviceID string) (models.FrameRecord, error) {
	deviceRowID, err := store.resolveDeviceRowID(deviceID)
	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return models.FrameRecord{}, ErrFrameNotFound
		}
		return models.FrameRecord{}, err
	}

	var frame models.FrameRecord
	cutoff := store.retentionCutoff()

	err = store.db.QueryRow(
		`SELECT id, path, captured_at, received_at
		 FROM stored_images
		 WHERE device_id = $1 AND captured_at >= $2
		 ORDER BY captured_at DESC, id DESC
		 LIMIT 1`,
		deviceRowID,
		cutoff,
	).Scan(&frame.ID, &frame.Path, &frame.CapturedAt, &frame.ReceivedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.FrameRecord{}, ErrFrameNotFound
		}
		return models.FrameRecord{}, fmt.Errorf("query last frame: %w", err)
	}

	frame.DeviceID = deviceID
	return frame, nil
}

func (store *PostgresStore) SaveMetrics(metrics models.DeviceMetricsRecord) error {
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
		`INSERT INTO android_telemetry (
			device_id, battery_level, signal_strength, network_type,
			temperature_c, latitude, longitude,
			cabin_noise_rms, gps_signal_quality, speed_kmh, ambient_light_lux,
			server_latency_ms, device_ip_address, jolt,
			cpu_usage_percent, ram_usage_percent,
			recorded_at, received_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
		deviceRowID,
		nullableFloat(metrics.BatteryLevel),
		nullableInt(metrics.SignalStrength),
		nullIfEmpty(metrics.NetworkType),
		nullableFloat(metrics.TemperatureC),
		nullableFloat(metrics.Latitude),
		nullableFloat(metrics.Longitude),
		nullableFloat(metrics.CabinNoiseRMS),
		nullIfEmpty(metrics.GPSSignalQuality),
		nullableFloat(metrics.SpeedKmh),
		nullableFloat(metrics.AmbientLightLux),
		nullableInt(metrics.ServerLatencyMs),
		nullIfEmpty(metrics.DeviceIPAddress),
		nullableFloat(metrics.Jolt),
		nullableFloat(metrics.CPUUsagePercent),
		nullableFloat(metrics.RAMUsagePercent),
		recordedAt,
		receivedAt,
	)
	if err != nil {
		return fmt.Errorf("insert android telemetry: %w", err)
	}

	return nil
}

func (store *PostgresStore) GetLatestMetrics(deviceID string) (models.DeviceMetricsRecord, error) {
	deviceRowID, err := store.resolveDeviceRowID(deviceID)
	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return models.DeviceMetricsRecord{}, ErrMetricsNotFound
		}
		return models.DeviceMetricsRecord{}, err
	}

	var metrics models.DeviceMetricsRecord
	var batteryLevel sql.NullFloat64
	var temperatureC sql.NullFloat64
	var latitude sql.NullFloat64
	var longitude sql.NullFloat64
	var signalStrength sql.NullInt64
	var networkType sql.NullString
	var cabinNoiseRMS sql.NullFloat64
	var gpsSignalQuality sql.NullString
	var speedKmh sql.NullFloat64
	var ambientLightLux sql.NullFloat64
	var serverLatencyMs sql.NullInt64
	var deviceIPAddress sql.NullString
	var jolt sql.NullFloat64
	var cpuUsagePercent sql.NullFloat64
	var ramUsagePercent sql.NullFloat64
	cutoff := store.retentionCutoff()

	err = store.db.QueryRow(
		`SELECT battery_level, signal_strength, network_type,
		        temperature_c, latitude, longitude,
		        cabin_noise_rms, gps_signal_quality, speed_kmh, ambient_light_lux,
		        server_latency_ms, device_ip_address, jolt,
		        cpu_usage_percent, ram_usage_percent,
		        recorded_at, received_at
		 FROM android_telemetry
		 WHERE device_id = $1 AND recorded_at >= $2
		 ORDER BY recorded_at DESC, id DESC
		 LIMIT 1`,
		deviceRowID,
		cutoff,
	).Scan(
		&batteryLevel,
		&signalStrength,
		&networkType,
		&temperatureC,
		&latitude,
		&longitude,
		&cabinNoiseRMS,
		&gpsSignalQuality,
		&speedKmh,
		&ambientLightLux,
		&serverLatencyMs,
		&deviceIPAddress,
		&jolt,
		&cpuUsagePercent,
		&ramUsagePercent,
		&metrics.RecordedAt,
		&metrics.ReceivedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.DeviceMetricsRecord{}, ErrMetricsNotFound
		}
		return models.DeviceMetricsRecord{}, fmt.Errorf("query latest metrics: %w", err)
	}

	metrics.DeviceID = deviceID
	metrics.BatteryLevel = floatPointer(batteryLevel)
	metrics.SignalStrength = intPointer(signalStrength)
	metrics.NetworkType = stringPointer(networkType)
	metrics.TemperatureC = floatPointer(temperatureC)
	metrics.Latitude = floatPointer(latitude)
	metrics.Longitude = floatPointer(longitude)
	metrics.CabinNoiseRMS = floatPointer(cabinNoiseRMS)
	metrics.GPSSignalQuality = stringValue(gpsSignalQuality)
	metrics.SpeedKmh = floatPointer(speedKmh)
	metrics.AmbientLightLux = floatPointer(ambientLightLux)
	metrics.ServerLatencyMs = intPointer(serverLatencyMs)
	metrics.DeviceIPAddress = stringValue(deviceIPAddress)
	metrics.Jolt = floatPointer(jolt)
	metrics.CPUUsagePercent = floatPointer(cpuUsagePercent)
	metrics.RAMUsagePercent = floatPointer(ramUsagePercent)

	return metrics, nil
}

func (store *PostgresStore) ListMetricsHistory(deviceID string, limit int) ([]models.DeviceMetricsRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	deviceRowID, err := store.resolveDeviceRowID(deviceID)
	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return nil, ErrMetricsNotFound
		}
		return nil, err
	}

	cutoff := store.retentionCutoff()
	rows, err := store.db.Query(
		`SELECT battery_level, signal_strength, network_type,
		        temperature_c, latitude, longitude,
		        cabin_noise_rms, gps_signal_quality, speed_kmh, ambient_light_lux,
		        server_latency_ms, device_ip_address, jolt,
		        cpu_usage_percent, ram_usage_percent,
		        recorded_at, received_at
		 FROM android_telemetry
		 WHERE device_id = $1 AND recorded_at >= $2
		 ORDER BY recorded_at DESC, id DESC
		 LIMIT $3`,
		deviceRowID,
		cutoff,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query metrics history: %w", err)
	}
	defer rows.Close()

	history := make([]models.DeviceMetricsRecord, 0)
	for rows.Next() {
		var metrics models.DeviceMetricsRecord
		var batteryLevel sql.NullFloat64
		var temperatureC sql.NullFloat64
		var latitude sql.NullFloat64
		var longitude sql.NullFloat64
		var signalStrength sql.NullInt64
		var networkType sql.NullString
		var cabinNoiseRMS sql.NullFloat64
		var gpsSignalQuality sql.NullString
		var speedKmh sql.NullFloat64
		var ambientLightLux sql.NullFloat64
		var serverLatencyMs sql.NullInt64
		var deviceIPAddress sql.NullString
		var jolt sql.NullFloat64
		var cpuUsagePercent sql.NullFloat64
		var ramUsagePercent sql.NullFloat64

		if err := rows.Scan(
			&batteryLevel,
			&signalStrength,
			&networkType,
			&temperatureC,
			&latitude,
			&longitude,
			&cabinNoiseRMS,
			&gpsSignalQuality,
			&speedKmh,
			&ambientLightLux,
			&serverLatencyMs,
			&deviceIPAddress,
			&jolt,
			&cpuUsagePercent,
			&ramUsagePercent,
			&metrics.RecordedAt,
			&metrics.ReceivedAt,
		); err != nil {
			return nil, fmt.Errorf("scan metrics history: %w", err)
		}

		metrics.DeviceID = deviceID
		metrics.BatteryLevel = floatPointer(batteryLevel)
		metrics.SignalStrength = intPointer(signalStrength)
		metrics.NetworkType = stringPointer(networkType)
		metrics.TemperatureC = floatPointer(temperatureC)
		metrics.Latitude = floatPointer(latitude)
		metrics.Longitude = floatPointer(longitude)
		metrics.CabinNoiseRMS = floatPointer(cabinNoiseRMS)
		metrics.GPSSignalQuality = stringValue(gpsSignalQuality)
		metrics.SpeedKmh = floatPointer(speedKmh)
		metrics.AmbientLightLux = floatPointer(ambientLightLux)
		metrics.ServerLatencyMs = intPointer(serverLatencyMs)
		metrics.DeviceIPAddress = stringValue(deviceIPAddress)
		metrics.Jolt = floatPointer(jolt)
		metrics.CPUUsagePercent = floatPointer(cpuUsagePercent)
		metrics.RAMUsagePercent = floatPointer(ramUsagePercent)
		history = append(history, metrics)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate metrics history: %w", err)
	}
	if len(history) == 0 {
		return nil, ErrMetricsNotFound
	}

	return history, nil
}

func (store *PostgresStore) ListFrames(limit int) ([]models.FrameRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	cutoff := store.retentionCutoff()
	rows, err := store.db.Query(
		`SELECT si.id, d.mac_address, si.path, si.captured_at, si.received_at
		 FROM stored_images si
		 JOIN devices d ON d.id = si.device_id
		 WHERE si.captured_at >= $1
		 ORDER BY si.captured_at DESC, si.id DESC
		 LIMIT $2`,
		cutoff,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list frames: %w", err)
	}
	defer rows.Close()

	frames := make([]models.FrameRecord, 0)
	for rows.Next() {
		var frame models.FrameRecord
		if err := rows.Scan(&frame.ID, &frame.DeviceID, &frame.Path, &frame.CapturedAt, &frame.ReceivedAt); err != nil {
			return nil, fmt.Errorf("scan frame: %w", err)
		}
		frames = append(frames, frame)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate frames: %w", err)
	}
	return frames, nil
}

func (store *PostgresStore) GetFrameByID(imageID int64) (models.FrameRecord, error) {
	var frame models.FrameRecord
	cutoff := store.retentionCutoff()
	err := store.db.QueryRow(
		`SELECT si.id, d.mac_address, si.path, si.captured_at, si.received_at
		 FROM stored_images si
		 JOIN devices d ON d.id = si.device_id
		 WHERE si.id = $1 AND si.captured_at >= $2`,
		imageID,
		cutoff,
	).Scan(&frame.ID, &frame.DeviceID, &frame.Path, &frame.CapturedAt, &frame.ReceivedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.FrameRecord{}, ErrFrameNotFound
		}
		return models.FrameRecord{}, fmt.Errorf("query frame by id: %w", err)
	}
	return frame, nil
}

func (store *PostgresStore) GetDeviceName(deviceID string) (string, error) {
	var name string
	err := store.db.QueryRow(
		`SELECT device_name FROM devices
		 WHERE mac_address = $1 OR device_name = $1 OR CAST(id AS TEXT) = $1
		 LIMIT 1`,
		deviceID,
	).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrDeviceNotFound
		}
		return "", fmt.Errorf("query device name: %w", err)
	}
	return name, nil
}

func (store *PostgresStore) SaveDiagnosticAudio(deviceID, segmentID, path string, startMs, endMs int64, rmsPeak float64, linkedAlertID, mode string) error {
	deviceRowID, err := store.resolveDeviceRowID(deviceID)
	if err != nil {
		return err
	}

	_, err = store.db.Exec(
		`INSERT INTO diagnostic_audio (
			device_id, segment_id, path, start_ms, end_ms, rms_peak, linked_alert_id, mode, received_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		deviceRowID,
		segmentID,
		path,
		startMs,
		endMs,
		rmsPeak,
		nullIfEmpty(linkedAlertID),
		nullIfEmpty(mode),
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert diagnostic audio: %w", err)
	}
	return nil
}

func stringValue(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

func (store *PostgresStore) CreateAction(action models.PhoneActionRecord) (models.PhoneActionRecord, error) {
	deviceRowID, err := store.resolveDeviceRowID(action.DeviceID)
	if err != nil {
		return models.PhoneActionRecord{}, err
	}

	payloadJSON, err := json.Marshal(action.Payload)
	if err != nil {
		return models.PhoneActionRecord{}, fmt.Errorf("marshal action payload: %w", err)
	}

	sentAt := action.SentAt
	if sentAt.IsZero() {
		sentAt = time.Now().UTC()
	}

	status := action.Status
	if status == "" {
		status = "pending"
	}

	err = store.db.QueryRow(
		`INSERT INTO phone_actions (device_id, action_type, payload, sent_at, status)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		deviceRowID,
		action.ActionType,
		payloadJSON,
		sentAt,
		status,
	).Scan(&action.ID)
	if err != nil {
		return models.PhoneActionRecord{}, fmt.Errorf("insert phone action: %w", err)
	}

	action.SentAt = sentAt
	action.Status = status
	return action, nil
}

func (store *PostgresStore) GetPendingActions(deviceID string) ([]models.PhoneActionRecord, error) {
	deviceRowID, err := store.resolveDeviceRowID(deviceID)
	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return []models.PhoneActionRecord{}, nil
		}
		return nil, err
	}

	rows, err := store.db.Query(
		`SELECT id, action_type, payload, sent_at, status
		 FROM phone_actions
		 WHERE device_id = $1 AND status = 'pending'
		 ORDER BY sent_at ASC, id ASC`,
		deviceRowID,
	)
	if err != nil {
		return nil, fmt.Errorf("query pending actions: %w", err)
	}
	defer rows.Close()

	actions := make([]models.PhoneActionRecord, 0)
	for rows.Next() {
		var action models.PhoneActionRecord
		var payloadJSON []byte
		if err := rows.Scan(&action.ID, &action.ActionType, &payloadJSON, &action.SentAt, &action.Status); err != nil {
			return nil, fmt.Errorf("scan pending action: %w", err)
		}
		action.DeviceID = deviceID
		if len(payloadJSON) > 0 {
			_ = json.Unmarshal(payloadJSON, &action.Payload)
		}
		if action.Payload == nil {
			action.Payload = map[string]any{}
		}
		actions = append(actions, action)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pending actions: %w", err)
	}

	return actions, nil
}

func (store *PostgresStore) AcknowledgeAction(actionID int64, status string) error {
	if status == "" {
		status = "done"
	}

	result, err := store.db.Exec(
		`UPDATE phone_actions SET status = $1 WHERE id = $2`,
		status,
		actionID,
	)
	if err != nil {
		return fmt.Errorf("acknowledge action: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("acknowledge action rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrActionNotFound
	}

	return nil
}

func (store *PostgresStore) UpsertSetting(setting models.AppSettingRecord) error {
	updatedAt := setting.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	_, err := store.db.Exec(
		`INSERT INTO app_settings (platform, key, value, updated_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (platform, key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`,
		setting.Platform,
		setting.Key,
		setting.Value,
		updatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert app setting: %w", err)
	}

	return nil
}

func (store *PostgresStore) GetSettings(platform string) ([]models.AppSettingRecord, error) {
	rows, err := store.db.Query(
		`SELECT platform, key, value, updated_at
		 FROM app_settings
		 WHERE platform = $1
		 ORDER BY key`,
		platform,
	)
	if err != nil {
		return nil, fmt.Errorf("query app settings: %w", err)
	}
	defer rows.Close()

	settings := make([]models.AppSettingRecord, 0)
	for rows.Next() {
		var setting models.AppSettingRecord
		if err := rows.Scan(&setting.Platform, &setting.Key, &setting.Value, &setting.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan app setting: %w", err)
		}
		settings = append(settings, setting)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate app settings: %w", err)
	}

	return settings, nil
}

func (store *PostgresStore) UpsertAIModel(model models.AIModelRecord) (models.AIModelRecord, error) {
	updatedAt := model.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	format := model.Format
	if format == "" {
		format = "ncnn"
	}

	labelsJSON, err := json.Marshal(model.Labels)
	if err != nil {
		return models.AIModelRecord{}, fmt.Errorf("marshal ai model labels: %w", err)
	}

	err = store.db.QueryRow(
		`INSERT INTO ai_model_paths (model_name, param_sha256, bin_sha256, labels, format, version, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (model_name) DO UPDATE SET
		   param_sha256 = EXCLUDED.param_sha256,
		   bin_sha256 = EXCLUDED.bin_sha256,
		   labels = EXCLUDED.labels,
		   format = EXCLUDED.format,
		   version = EXCLUDED.version,
		   updated_at = EXCLUDED.updated_at
		 RETURNING id, updated_at`,
		model.ModelName,
		model.ParamSHA256,
		model.BinSHA256,
		labelsJSON,
		format,
		model.Version,
		updatedAt,
	).Scan(&model.ID, &model.UpdatedAt)
	if err != nil {
		return models.AIModelRecord{}, fmt.Errorf("upsert ai model path: %w", err)
	}

	model.Format = format
	return model, nil
}

func (store *PostgresStore) ListAIModels() ([]models.AIModelRecord, error) {
	rows, err := store.db.Query(
		`SELECT id, model_name, param_sha256, bin_sha256, labels, format, version, updated_at
		 FROM ai_model_paths
		 ORDER BY model_name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list ai model paths: %w", err)
	}
	defer rows.Close()

	modelsList := make([]models.AIModelRecord, 0)
	for rows.Next() {
		var model models.AIModelRecord
		var labelsJSON []byte
		if err := rows.Scan(
			&model.ID,
			&model.ModelName,
			&model.ParamSHA256,
			&model.BinSHA256,
			&labelsJSON,
			&model.Format,
			&model.Version,
			&model.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ai model path: %w", err)
		}
		if err := json.Unmarshal(labelsJSON, &model.Labels); err != nil {
			return nil, fmt.Errorf("unmarshal ai model labels: %w", err)
		}
		modelsList = append(modelsList, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ai model paths: %w", err)
	}

	return modelsList, nil
}

func (store *PostgresStore) GetAIModelByName(modelName string) (models.AIModelRecord, error) {
	var model models.AIModelRecord
	var labelsJSON []byte
	err := store.db.QueryRow(
		`SELECT id, model_name, param_sha256, bin_sha256, labels, format, version, updated_at
		 FROM ai_model_paths
		 WHERE model_name = $1`,
		modelName,
	).Scan(
		&model.ID,
		&model.ModelName,
		&model.ParamSHA256,
		&model.BinSHA256,
		&labelsJSON,
		&model.Format,
		&model.Version,
		&model.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.AIModelRecord{}, ErrAIModelNotFound
		}
		return models.AIModelRecord{}, fmt.Errorf("get ai model path: %w", err)
	}
	if err := json.Unmarshal(labelsJSON, &model.Labels); err != nil {
		return models.AIModelRecord{}, fmt.Errorf("unmarshal ai model labels: %w", err)
	}
	return model, nil
}

func (store *PostgresStore) SaveConnection(connection models.WebRTCConnectionRecord) (models.WebRTCConnectionRecord, error) {
	deviceRowID, err := store.resolveDeviceRowID(connection.DeviceID)
	if err != nil {
		return models.WebRTCConnectionRecord{}, err
	}

	connectedAt := connection.ConnectedAt
	if connectedAt.IsZero() {
		connectedAt = time.Now().UTC()
	}

	status := connection.Status
	if status == "" {
		status = "active"
	}

	err = store.db.QueryRow(
		`INSERT INTO webrtc_connections (device_id, room, identity, role, connected_at, disconnected_at, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		deviceRowID,
		connection.Room,
		connection.Identity,
		connection.Role,
		connectedAt,
		connection.DisconnectedAt,
		status,
	).Scan(&connection.ID)
	if err != nil {
		return models.WebRTCConnectionRecord{}, fmt.Errorf("insert webrtc connection: %w", err)
	}

	connection.ConnectedAt = connectedAt
	connection.Status = status
	return connection, nil
}

func (store *PostgresStore) ListConnections(deviceID string, limit int) ([]models.WebRTCConnectionRecord, error) {
	if limit <= 0 {
		limit = 50
	}

	deviceRowID, err := store.resolveDeviceRowID(deviceID)
	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return []models.WebRTCConnectionRecord{}, nil
		}
		return nil, err
	}

	rows, err := store.db.Query(
		`SELECT id, room, identity, role, connected_at, disconnected_at, status
		 FROM webrtc_connections
		 WHERE device_id = $1
		 ORDER BY connected_at DESC, id DESC
		 LIMIT $2`,
		deviceRowID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list webrtc connections: %w", err)
	}
	defer rows.Close()

	connections := make([]models.WebRTCConnectionRecord, 0)
	for rows.Next() {
		var connection models.WebRTCConnectionRecord
		var disconnectedAt sql.NullTime
		if err := rows.Scan(
			&connection.ID,
			&connection.Room,
			&connection.Identity,
			&connection.Role,
			&connection.ConnectedAt,
			&disconnectedAt,
			&connection.Status,
		); err != nil {
			return nil, fmt.Errorf("scan webrtc connection: %w", err)
		}
		connection.DeviceID = deviceID
		if disconnectedAt.Valid {
			disconnectedAtValue := disconnectedAt.Time
			connection.DisconnectedAt = &disconnectedAtValue
		}
		connections = append(connections, connection)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate webrtc connections: %w", err)
	}

	return connections, nil
}

func (store *PostgresStore) PurgeExpiredFrames(cutoff time.Time) ([]string, error) {
	rows, err := store.db.Query(
		`SELECT path FROM stored_images WHERE captured_at < $1`,
		cutoff.UTC(),
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

	if _, err := store.db.Exec(`DELETE FROM stored_images WHERE captured_at < $1`, cutoff.UTC()); err != nil {
		return nil, fmt.Errorf("delete expired frames: %w", err)
	}

	return framePaths, nil
}

func (store *PostgresStore) PurgeExpiredMetrics(cutoff time.Time) error {
	if _, err := store.db.Exec(`DELETE FROM android_telemetry WHERE recorded_at < $1`, cutoff.UTC()); err != nil {
		return fmt.Errorf("delete expired metrics: %w", err)
	}
	return nil
}

func (store *PostgresStore) UpsertDevice(device models.Device) (models.Device, error) {
	if device.MACAddress == "" {
		return models.Device{}, fmt.Errorf("mac_address is required")
	}
	if device.DeviceName == "" {
		device.DeviceName = device.MACAddress
	}

	err := store.db.QueryRow(
		`INSERT INTO devices (device_name, mac_address)
		 VALUES ($1, $2)
		 ON CONFLICT (mac_address) DO UPDATE SET device_name = EXCLUDED.device_name
		 RETURNING id, device_name, mac_address`,
		device.DeviceName,
		device.MACAddress,
	).Scan(&device.ID, &device.DeviceName, &device.MACAddress)
	if err != nil {
		return models.Device{}, fmt.Errorf("upsert device: %w", err)
	}

	return device, nil
}

func (store *PostgresStore) GetDeviceByMACAddress(macAddress string) (models.Device, error) {
	var device models.Device
	err := store.db.QueryRow(
		`SELECT id, device_name, mac_address FROM devices WHERE mac_address = $1`,
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

func (store *PostgresStore) resolveDeviceRowID(deviceIdentifier string) (int64, error) {
	var deviceRowID int64
	err := store.db.QueryRow(
		`SELECT id FROM devices
		 WHERE mac_address = $1 OR device_name = $1 OR CAST(id AS TEXT) = $1
		 LIMIT 1`,
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

func (store *PostgresStore) ListDevices() ([]models.DeviceListItem, error) {
	rows, err := store.db.Query(
		`SELECT d.mac_address, d.device_name,
		        CASE
		            WHEN t.recorded_at IS NOT NULL AND t.recorded_at >= NOW() AT TIME ZONE 'UTC' - INTERVAL '5 minutes'
		            THEN 'online'
		            ELSE 'offline'
		        END AS status
		 FROM devices d
		 LEFT JOIN LATERAL (
		     SELECT recorded_at
		     FROM android_telemetry
		     WHERE device_id = d.id
		     ORDER BY recorded_at DESC
		     LIMIT 1
		 ) t ON TRUE
		 ORDER BY d.device_name ASC, d.mac_address ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	devices := make([]models.DeviceListItem, 0)
	for rows.Next() {
		var device models.DeviceListItem
		if err := rows.Scan(&device.ID, &device.Name, &device.Status); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		devices = append(devices, device)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate devices: %w", err)
	}

	return devices, nil
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

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
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

func stringPointer(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}
