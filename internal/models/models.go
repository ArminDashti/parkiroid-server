package models

import "time"

type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
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
	ID         int64  `json:"id"`
	DeviceName string `json:"device_name"`
	MACAddress string `json:"mac_address"`
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
	DeviceID       string    `json:"device_id" binding:"required"`
	BatteryLevel   *float64  `json:"battery_level_percent,omitempty"`
	SignalStrength *int      `json:"signal_strength_dbm,omitempty"`
	NetworkType    string    `json:"network_type,omitempty"`
	TemperatureC   *float64  `json:"temperature_celsius,omitempty"`
	Latitude       *float64  `json:"latitude,omitempty"`
	Longitude      *float64  `json:"longitude,omitempty"`
	RecordedAt     time.Time `json:"recorded_at"`
}

type DeviceMetricsRecord struct {
	DeviceID       string    `json:"device_id"`
	BatteryLevel   *float64  `json:"battery_level_percent,omitempty"`
	SignalStrength *int      `json:"signal_strength_dbm,omitempty"`
	NetworkType    string    `json:"network_type,omitempty"`
	TemperatureC   *float64  `json:"temperature_celsius,omitempty"`
	Latitude       *float64  `json:"latitude,omitempty"`
	Longitude      *float64  `json:"longitude,omitempty"`
	RecordedAt     time.Time `json:"recorded_at"`
	ReceivedAt     time.Time `json:"received_at"`
}

type PhoneActionPayload struct {
	DeviceID   string         `json:"device_id" binding:"required"`
	ActionType string         `json:"action_type" binding:"required"`
	Payload    map[string]any `json:"payload"`
}

type PhoneActionRecord struct {
	ID         int64          `json:"id"`
	DeviceID   string         `json:"device_id"`
	ActionType string         `json:"action_type"`
	Payload    map[string]any `json:"payload"`
	SentAt     time.Time      `json:"sent_at"`
	Status     string         `json:"status"`
}

type PhoneActionAckPayload struct {
	Status string `json:"status"`
}

type AppSettingPayload struct {
	Platform string `json:"platform" binding:"required"`
	Key      string `json:"key" binding:"required"`
	Value    string `json:"value" binding:"required"`
}

type AppSettingRecord struct {
	Platform  string    `json:"platform"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AIModelPayload struct {
	ModelName string `json:"model_name" binding:"required"`
	Path      string `json:"path" binding:"required"`
	Version   string `json:"version"`
}

type AIModelRecord struct {
	ID        int64     `json:"id"`
	ModelName string    `json:"model_name"`
	Path      string    `json:"path"`
	Version   string    `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WebRTCConnectionRecord struct {
	ID             int64      `json:"id"`
	DeviceID       string     `json:"device_id"`
	Room           string     `json:"room"`
	Identity       string     `json:"identity"`
	Role           string     `json:"role"`
	ConnectedAt    time.Time  `json:"connected_at"`
	DisconnectedAt *time.Time `json:"disconnected_at,omitempty"`
	Status         string     `json:"status"`
}

type LiveKitTokenRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
	Identity string `json:"identity"`
	Role     string `json:"role"`
}

type LiveKitTokenResponse struct {
	Token     string    `json:"token"`
	URL       string    `json:"url"`
	Room      string    `json:"room"`
	Identity  string    `json:"identity"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
