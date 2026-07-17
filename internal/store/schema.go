package store

const schemaDDL = `
CREATE TABLE IF NOT EXISTS devices (
	id BIGSERIAL PRIMARY KEY,
	device_name TEXT NOT NULL,
	mac_address TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS login_logs (
	id BIGSERIAL PRIMARY KEY,
	ip TEXT NOT NULL,
	username TEXT NOT NULL,
	browser_info TEXT NOT NULL DEFAULT '',
	attempted_at TIMESTAMPTZ NOT NULL,
	success BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_login_logs_attempted_at ON login_logs (attempted_at DESC);

CREATE TABLE IF NOT EXISTS stored_images (
	id BIGSERIAL PRIMARY KEY,
	device_id BIGINT NOT NULL REFERENCES devices(id),
	path TEXT NOT NULL,
	captured_at TIMESTAMPTZ NOT NULL,
	received_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_stored_images_device_captured_at
	ON stored_images (device_id, captured_at DESC);

CREATE TABLE IF NOT EXISTS phone_actions (
	id BIGSERIAL PRIMARY KEY,
	device_id BIGINT NOT NULL REFERENCES devices(id),
	action_type TEXT NOT NULL,
	payload JSONB NOT NULL DEFAULT '{}',
	sent_at TIMESTAMPTZ NOT NULL,
	status TEXT NOT NULL DEFAULT 'pending'
);

CREATE INDEX IF NOT EXISTS idx_phone_actions_device_status
	ON phone_actions (device_id, status, sent_at DESC);

CREATE TABLE IF NOT EXISTS webrtc_connections (
	id BIGSERIAL PRIMARY KEY,
	device_id BIGINT NOT NULL REFERENCES devices(id),
	room TEXT NOT NULL,
	identity TEXT NOT NULL,
	role TEXT NOT NULL,
	connected_at TIMESTAMPTZ NOT NULL,
	disconnected_at TIMESTAMPTZ,
	status TEXT NOT NULL DEFAULT 'active'
);

CREATE INDEX IF NOT EXISTS idx_webrtc_connections_device_connected_at
	ON webrtc_connections (device_id, connected_at DESC);

CREATE TABLE IF NOT EXISTS app_settings (
	id BIGSERIAL PRIMARY KEY,
	platform TEXT NOT NULL,
	key TEXT NOT NULL,
	value TEXT NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,
	UNIQUE (platform, key)
);

CREATE TABLE IF NOT EXISTS ai_model_paths (
	id BIGSERIAL PRIMARY KEY,
	model_name TEXT NOT NULL UNIQUE,
	param_sha256 TEXT NOT NULL DEFAULT '',
	bin_sha256 TEXT NOT NULL DEFAULT '',
	labels JSONB NOT NULL DEFAULT '[]',
	format TEXT NOT NULL DEFAULT 'ncnn',
	version TEXT NOT NULL DEFAULT '',
	updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS android_telemetry (
	id BIGSERIAL PRIMARY KEY,
	device_id BIGINT NOT NULL REFERENCES devices(id),
	battery_level DOUBLE PRECISION,
	signal_strength INTEGER,
	network_type TEXT,
	temperature_c DOUBLE PRECISION,
	latitude DOUBLE PRECISION,
	longitude DOUBLE PRECISION,
	cabin_noise_rms DOUBLE PRECISION,
	gps_signal_quality TEXT,
	speed_kmh DOUBLE PRECISION,
	ambient_light_lux DOUBLE PRECISION,
	server_latency_ms INTEGER,
	device_ip_address TEXT,
	jolt DOUBLE PRECISION,
	cpu_usage_percent DOUBLE PRECISION,
	ram_usage_percent DOUBLE PRECISION,
	recorded_at TIMESTAMPTZ NOT NULL,
	received_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_android_telemetry_device_recorded_at
	ON android_telemetry (device_id, recorded_at DESC);

CREATE TABLE IF NOT EXISTS diagnostic_audio (
	id BIGSERIAL PRIMARY KEY,
	device_id BIGINT NOT NULL REFERENCES devices(id),
	segment_id TEXT NOT NULL,
	path TEXT NOT NULL,
	start_ms BIGINT NOT NULL DEFAULT 0,
	end_ms BIGINT NOT NULL DEFAULT 0,
	rms_peak DOUBLE PRECISION,
	linked_alert_id TEXT,
	mode TEXT,
	received_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_diagnostic_audio_device_received_at
	ON diagnostic_audio (device_id, received_at DESC);
`

const schemaMigrationDDL = `
ALTER TABLE ai_model_paths ADD COLUMN IF NOT EXISTS param_sha256 TEXT NOT NULL DEFAULT '';
ALTER TABLE ai_model_paths ADD COLUMN IF NOT EXISTS bin_sha256 TEXT NOT NULL DEFAULT '';
ALTER TABLE ai_model_paths ADD COLUMN IF NOT EXISTS labels JSONB NOT NULL DEFAULT '[]';
ALTER TABLE ai_model_paths ADD COLUMN IF NOT EXISTS format TEXT NOT NULL DEFAULT 'ncnn';
ALTER TABLE ai_model_paths DROP COLUMN IF EXISTS path;
ALTER TABLE android_telemetry ADD COLUMN IF NOT EXISTS cabin_noise_rms DOUBLE PRECISION;
ALTER TABLE android_telemetry ADD COLUMN IF NOT EXISTS gps_signal_quality TEXT;
ALTER TABLE android_telemetry ADD COLUMN IF NOT EXISTS speed_kmh DOUBLE PRECISION;
ALTER TABLE android_telemetry ADD COLUMN IF NOT EXISTS ambient_light_lux DOUBLE PRECISION;
ALTER TABLE android_telemetry ADD COLUMN IF NOT EXISTS server_latency_ms INTEGER;
ALTER TABLE android_telemetry ADD COLUMN IF NOT EXISTS device_ip_address TEXT;
ALTER TABLE android_telemetry ADD COLUMN IF NOT EXISTS jolt DOUBLE PRECISION;
ALTER TABLE android_telemetry ADD COLUMN IF NOT EXISTS cpu_usage_percent DOUBLE PRECISION;
ALTER TABLE android_telemetry ADD COLUMN IF NOT EXISTS ram_usage_percent DOUBLE PRECISION;
`
