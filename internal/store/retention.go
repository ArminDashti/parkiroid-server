package store

import (
	"log"
	"os"
	"time"
)

const retentionCleanupInterval = time.Hour

type RetentionCleaner struct {
	store           RetentionStore
	framesDir       string
	retentionPeriod time.Duration
}

func NewRetentionCleaner(store RetentionStore, framesDir string, retentionPeriod time.Duration) *RetentionCleaner {
	return &RetentionCleaner{
		store:           store,
		framesDir:       framesDir,
		retentionPeriod: retentionPeriod,
	}
}

func (cleaner *RetentionCleaner) Start() {
	go cleaner.run()
}

func (cleaner *RetentionCleaner) run() {
	cleaner.cleanupOnce()

	ticker := time.NewTicker(retentionCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		cleaner.cleanupOnce()
	}
}

func (cleaner *RetentionCleaner) cleanupOnce() {
	cutoff := time.Now().UTC().Add(-cleaner.retentionPeriod)

	framePaths, err := cleaner.store.PurgeExpiredFrames(cutoff)
	if err != nil {
		log.Printf("retention cleanup failed for frames: %v", err)
	} else {
		cleaner.removeFrameFiles(framePaths)
	}

	if err := cleaner.store.PurgeExpiredMetrics(cutoff); err != nil {
		log.Printf("retention cleanup failed for metrics: %v", err)
	}
}

func (cleaner *RetentionCleaner) removeFrameFiles(framePaths []string) {
	for _, framePath := range framePaths {
		if err := os.Remove(framePath); err != nil && !os.IsNotExist(err) {
			log.Printf("retention cleanup failed to remove frame file %s: %v", framePath, err)
		}
	}
}
