package store

import (
	"errors"
	"sync"
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
	SaveMetrics(metrics models.DeviceMetricsRecord)
	GetLatestMetrics(deviceID string) (models.DeviceMetricsRecord, error)
}

type MemoryStore struct {
	mu      sync.RWMutex
	frames  map[string]models.FrameRecord
	metrics map[string]models.DeviceMetricsRecord
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		frames:  make(map[string]models.FrameRecord),
		metrics: make(map[string]models.DeviceMetricsRecord),
	}
}

func (store *MemoryStore) SaveFrame(frame models.FrameRecord) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.frames[frame.DeviceID] = frame
	return nil
}

func (store *MemoryStore) GetLastFrame(deviceID string) (models.FrameRecord, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	frame, exists := store.frames[deviceID]
	if !exists {
		return models.FrameRecord{}, ErrFrameNotFound
	}

	return frame, nil
}

func (store *MemoryStore) SaveMetrics(metrics models.DeviceMetricsRecord) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.metrics[metrics.DeviceID] = metrics
}

func (store *MemoryStore) GetLatestMetrics(deviceID string) (models.DeviceMetricsRecord, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	metrics, exists := store.metrics[deviceID]
	if !exists {
		return models.DeviceMetricsRecord{}, ErrMetricsNotFound
	}

	return metrics, nil
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
