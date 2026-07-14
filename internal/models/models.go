package models

import "time"

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	APIKey   string `json:"api_key"`
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
	ModelName   string   `json:"model_name" binding:"required"`
	Version     string   `json:"version"`
	ParamSHA256 string   `json:"param_sha256"`
	BinSHA256   string   `json:"bin_sha256"`
	Labels      []string `json:"labels"`
	Format      string   `json:"format"`
}

type AIModelRecord struct {
	ID          int64     `json:"id"`
	ModelName   string    `json:"model_name"`
	ParamSHA256 string    `json:"param_sha256"`
	BinSHA256   string    `json:"bin_sha256"`
	Labels      []string  `json:"labels"`
	Format      string    `json:"format"`
	Version     string    `json:"version"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AIModelManifestEntry struct {
	ID          string   `json:"id"`
	ParamURL    string   `json:"param_url"`
	BinURL      string   `json:"bin_url"`
	ParamSHA256 string   `json:"param_sha256"`
	BinSHA256   string   `json:"bin_sha256"`
	Format      string   `json:"format"`
	Labels      []string `json:"labels"`
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

type WebRTCSessionRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
}

type IceServerConfig struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

type WebRTCSessionResponse struct {
	SessionID  string            `json:"session_id"`
	Token      string            `json:"token"`
	URL        string            `json:"url"`
	Room       string            `json:"room"`
	Identity   string            `json:"identity"`
	ExpiresAt  time.Time         `json:"expires_at"`
	IceServers []IceServerConfig `json:"ice_servers"`
}

type DeviceListItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Location string `json:"location,omitempty"`
}

type DeviceStreamResponse struct {
	DeviceID  string    `json:"device_id"`
	Token     string    `json:"token"`
	URL       string    `json:"url"`
	Room      string    `json:"room"`
	Identity  string    `json:"identity"`
	ExpiresAt time.Time `json:"expires_at"`
}

type WebLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type WebUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type WebAuthResponse struct {
	Token string  `json:"token"`
	User  WebUser `json:"user"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
