package store

import (
	"errors"
	"time"

	"github.com/dogan/dogan-server/internal/models"
)

var (
	ErrFrameNotFound    = errors.New("frame not found for device")
	ErrMetricsNotFound  = errors.New("device metrics not found")
	ErrActionNotFound   = errors.New("action not found")
	ErrDeviceNotFound   = errors.New("device not found")
	ErrAIModelNotFound  = errors.New("ai model not found")
)

type FrameStore interface {
	SaveFrame(frame models.FrameRecord) error
	GetLastFrame(deviceID string) (models.FrameRecord, error)
	ListFrames(limit int) ([]models.FrameRecord, error)
	GetFrameByID(imageID int64) (models.FrameRecord, error)
}

type MetricsStore interface {
	SaveMetrics(metrics models.DeviceMetricsRecord) error
	GetLatestMetrics(deviceID string) (models.DeviceMetricsRecord, error)
	ListMetricsHistory(deviceID string, limit int) ([]models.DeviceMetricsRecord, error)
}

type LoginLogStore interface {
	SaveLoginLog(ip, username, browserInfo string, success bool) error
}

type ActionStore interface {
	CreateAction(action models.PhoneActionRecord) (models.PhoneActionRecord, error)
	GetPendingActions(deviceID string) ([]models.PhoneActionRecord, error)
	AcknowledgeAction(actionID int64, status string) error
}

type SettingsStore interface {
	UpsertSetting(setting models.AppSettingRecord) error
	GetSettings(platform string) ([]models.AppSettingRecord, error)
}

type AIModelStore interface {
	UpsertAIModel(model models.AIModelRecord) (models.AIModelRecord, error)
	ListAIModels() ([]models.AIModelRecord, error)
	GetAIModelByName(modelName string) (models.AIModelRecord, error)
}

type WebRTCStore interface {
	SaveConnection(connection models.WebRTCConnectionRecord) (models.WebRTCConnectionRecord, error)
	ListConnections(deviceID string, limit int) ([]models.WebRTCConnectionRecord, error)
}

type DeviceStore interface {
	ListDevices() ([]models.DeviceListItem, error)
	GetDeviceName(deviceID string) (string, error)
}

type DiagnosticAudioStore interface {
	SaveDiagnosticAudio(deviceID, segmentID, path string, startMs, endMs int64, rmsPeak float64, linkedAlertID, mode string) error
}

type RetentionStore interface {
	PurgeExpiredFrames(cutoff time.Time) ([]string, error)
	PurgeExpiredMetrics(cutoff time.Time) error
}

func NormalizeCapturedAt(capturedAt time.Time) time.Time {
	if capturedAt.IsZero() {
		return time.Now().UTC()
	}
	return capturedAt.UTC()
}

func NormalizeRecordedAt(recordedAt time.Time) time.Time {
	if recordedAt.IsZero() {
		return time.Now().UTC()
	}
	return recordedAt.UTC()
}
