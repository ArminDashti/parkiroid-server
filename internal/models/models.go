package models

import "time"

type AuthRequest struct {
	APIKey string `json:"api_key" binding:"required"`
}

type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type EndpointDescriptor struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Auth        bool   `json:"auth_required"`
}

type FramePayload struct {
	DeviceID   string    `json:"device_id" binding:"required"`
	ImageData  string    `json:"image_data" binding:"required"`
	CapturedAt time.Time `json:"captured_at"`
}

type Device struct {
	ID          int64  `json:"id"`
	DeviceName  string `json:"device_name"`
	MACAddress  string `json:"mac_address"`
}

type FrameRecord struct {
	ID         int64     `json:"id,omitempty"`
	DeviceID   string    `json:"device_id"`
	Path       string    `json:"path,omitempty"`
	ImageData  string    `json:"image_data,omitempty"`
	CapturedAt time.Time `json:"captured_at"`
	ReceivedAt time.Time `json:"received_at"`
}

type DeviceMetricsPayload struct {
	DeviceID      string    `json:"device_id" binding:"required"`
	CPUUsage      float64   `json:"cpu_usage_percent"`
	MemoryUsage   float64   `json:"memory_usage_percent"`
	DiskUsage     float64   `json:"disk_usage_percent"`
	BatteryLevel  *float64  `json:"battery_level_percent,omitempty"`
	TemperatureC  *float64  `json:"temperature_celsius,omitempty"`
	SignalStrength *int     `json:"signal_strength_dbm,omitempty"`
	RecordedAt    time.Time `json:"recorded_at"`
}

type DeviceMetricsRecord struct {
	DeviceID       string    `json:"device_id"`
	CPUUsage       float64   `json:"cpu_usage_percent"`
	MemoryUsage    float64   `json:"memory_usage_percent"`
	DiskUsage      float64   `json:"disk_usage_percent"`
	BatteryLevel   *float64  `json:"battery_level_percent,omitempty"`
	TemperatureC   *float64  `json:"temperature_celsius,omitempty"`
	SignalStrength *int      `json:"signal_strength_dbm,omitempty"`
	RecordedAt     time.Time `json:"recorded_at"`
	ReceivedAt     time.Time `json:"received_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
