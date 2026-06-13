package store

const schemaDDL = `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS devices (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	device_name TEXT NOT NULL,
	mac_address TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS frames (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	path TEXT NOT NULL,
	device_id INTEGER NOT NULL,
	captured_at TEXT NOT NULL,
	FOREIGN KEY (device_id) REFERENCES devices(id)
);

CREATE INDEX IF NOT EXISTS idx_frames_device_captured_at
	ON frames (device_id, captured_at DESC);

CREATE TABLE IF NOT EXISTS device_metrics (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	device_id INTEGER NOT NULL,
	cpu_usage REAL NOT NULL,
	memory_usage REAL NOT NULL,
	disk_usage REAL NOT NULL,
	battery_level REAL,
	temperature_c REAL,
	signal_strength INTEGER,
	recorded_at TEXT NOT NULL,
	received_at TEXT NOT NULL,
	FOREIGN KEY (device_id) REFERENCES devices(id)
);

CREATE INDEX IF NOT EXISTS idx_device_metrics_device_recorded_at
	ON device_metrics (device_id, recorded_at DESC);
`
