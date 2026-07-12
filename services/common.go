package services

import (
	"clustta/internal/repository"
	"clustta/internal/settings"
	"clustta/output"
	"context"
	"log"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/wailsapp/wails/v3/pkg/application"
)

var mu sync.Mutex
var cancel context.CancelFunc
var ctx context.Context

func init() {
	ctx, cancel = context.WithCancel(context.Background())
}

// cancelSync cancels the current context in a thread-safe manner.
// Acquires a lock before canceling to ensure thread safety.
func cancelSync() {
	mu.Lock()
	if cancel != nil {
		cancel()
	}
	mu.Unlock()
}

// reset cancels the current context and creates a new one.
// Used to reset the context state after an operation completes or is cancelled.
func reset() {
	mu.Lock()
	defer mu.Unlock()
	if cancel != nil {
		cancel()
	}
	ctx, cancel = context.WithCancel(context.Background())
}

// getContext returns the current context in a thread-safe way.
// Acquires a lock to safely read the context variable.
func getContext() context.Context {
	mu.Lock()
	defer mu.Unlock()
	return ctx
}

// clearChunkCacheIfEnabled removes cached chunks after a completed transfer.
func clearChunkCacheIfEnabled(projectPath string, dbConn *sqlx.DB) error {
	enabled, err := settings.GetMetadataOnlyStorage()
	if err != nil || !enabled {
		return err
	}

	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err = repository.ClearSyncedChunkCache(tx); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}

	if err = repository.VacuumIfNeeded(dbConn, projectPath, func() {
		application.Get().Event.Emit("progress-update", output.ProgressReport{
			Title:      "Reclaiming archive space",
			Message:    "Compacting project archive",
			Percentage: 99,
			Current:    1,
			Total:      1,
		})
	}); err != nil {
		log.Printf("Failed to reclaim project archive space: %v", err)
	}
	return nil
}
