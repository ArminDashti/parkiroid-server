package store

import (
	"errors"
	"time"

	"github.com/parkiroid/parkiroid-server/internal/models"
)

var ErrFrameNotFound = errors.New("frame not found for device")
var ErrMetricsNotFound = errors.New("device metrics not found")

type FrameStore interface {
	SaveFrame(frame models.FrameRecord) error
	GetLastFrame(deviceID string) (models.FrameRecord, error)
}

type MetricsStore interface {
	SaveMetrics(metrics models.DeviceMetricsRecord) error
	GetLatestMetrics(deviceID string) (models.DeviceMetricsRecord, error)
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
