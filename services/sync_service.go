package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/error_service"
	"clustta/internal/repository/sync_service"
	"clustta/internal/settings"
	"clustta/internal/utils"
	"clustta/output"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type SyncService struct{}

func (s *SyncService) CancelSync() {
	mu.Lock()
	if cancel != nil {
		cancel()
	}
	mu.Unlock()
}

func (s *SyncService) CloneProject(projectUri, studioName, workingDir string, syncOptions sync_service.SyncOptions) error {
	defer reset() // Ensure context is reset when we're done

	ctx := getContext()
	if ctx.Err() != nil {
		return errors.New("operation cancelled before starting")
	}

	app := application.Get()
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	var projectName string
	if utils.IsValidURL(projectUri) {
		projectName = path.Base(projectUri)
	} else if utils.FileExists(projectUri) {
		projectName = strings.Split(filepath.Base(projectUri), ".")[0]

	}
	projectsDir, err := settings.GetSharedProjectDirectory()
	fmt.Println(projectsDir)
	if err != nil {
		return err
	}
	studioProjectsDir := filepath.Join(projectsDir, studioName)
	projectPath := filepath.Join(studioProjectsDir, projectName) + ".clst"

	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		err = os.MkdirAll(workingDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Create buffered channels to prevent blocking
	errChan := make(chan error, 1)
	progressChan := make(chan output.ProgressReport, 10)

	// Start progress update goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return // Exit immediately on cancellation
			case progress, ok := <-progressChan:
				if !ok {
					return
				}
				app.Event.Emit("progress-update", progress)
			}
		}
	}()

	// Initial progress
	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:         fmt.Sprintf("Downloading %s  Project", projectName),
		Message:       fmt.Sprintf("Downloading %s  Project", projectName),
		Percentage:    0,
		Current:       1,
		Total:         1,
		OperationType: "write", // Sync operations modify database
	}:
	}

	callBack := func(current int, total int, message string, extraMessage string) {
		if ctx.Err() != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case progressChan <- output.ProgressReport{
			Title:         fmt.Sprintf("Downloading %s  Project", projectName),
			Message:       message,
			Percentage:    float64(current) / float64(total) * 98,
			Current:       1,
			Total:         1,
			ExtraMessage:  extraMessage,
			OperationType: "write",
		}:
		default: // Skip progress update if channel is full
		}
	}

	// Push data with cancellation
	cloneDone := make(chan struct{})
	go func() {
		defer close(cloneDone)
		err := sync_service.CloneProject(ctx, projectUri, projectPath, studioName, workingDir, user, syncOptions, callBack)
		if ctx.Err() == nil { // Only send error if not cancelled
			errChan <- err
		}
	}()

	select {
	case err = <-errChan:
		if err != nil {
			if utils.FileExists(projectPath) {
				journal := projectPath + "-journal"
				err := os.Remove(projectPath)
				if err != nil {
					return err
				}
				if utils.FileExists(journal) {
					err = os.Remove(journal)
					if err != nil {
						return err
					}
				}
			}
			close(progressChan)
			return err
		}
	case <-ctx.Done():
		<-cloneDone
		if utils.FileExists(projectPath) {
			journal := projectPath + "-journal"
			err := os.Remove(projectPath)
			if err != nil {
				return err
			}
			if utils.FileExists(journal) {
				err = os.Remove(journal)
				if err != nil {
					return err
				}
			}

		}
		close(progressChan) // Stop progress updates
		return errors.New("operation cancelled")
	}

	close(progressChan)
	progress := output.ProgressReport{
		Title:         fmt.Sprintf("Downloading %s  Project", projectName),
		Message:       fmt.Sprintf("Downloading %s  Project", projectName),
		Percentage:    100,
		Current:       1,
		Total:         1,
		OperationType: "write",
	}
	app.Event.Emit("progress-update", progress)
	return nil
}

func (s *SyncService) SyncData(projectPath, remoteURL string, pullChunk bool, syncOptions sync_service.SyncOptions) error {
	defer reset() // Ensure context is reset when we're done

	ctx := getContext()
	if ctx.Err() != nil {
		return errors.New("operation cancelled before starting")
	}

	app := application.Get()
	activeUser, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	// Create buffered channels to prevent blocking
	errChan := make(chan error, 1)
	progressChan := make(chan output.ProgressReport, 10)

	// Start progress update goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return // Exit immediately on cancellation
			case progress, ok := <-progressChan:
				if !ok {
					return
				}
				app.Event.Emit("progress-update", progress)
			}
		}
	}()

	// Initial progress
	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      "Syncing",
		Message:    "Sending",
		Percentage: 0,
		Current:    1,
		Total:      2,
	}:
	}

	pushCallBack := func(current int, total int, message string, extraMessage string) {
		if ctx.Err() != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case progressChan <- output.ProgressReport{
			Title:        "Syncing",
			Message:      message,
			Percentage:   (float64(current) / float64(total) * 99),
			Current:      1,
			Total:        2,
			ExtraMessage: extraMessage,
		}:
		default: // Skip progress update if channel is full
		}
	}

	// Push data with cancellation
	go func() {
		err := sync_service.PushData(projectPath, remoteURL, activeUser.Id, pushCallBack)
		if ctx.Err() == nil { // Only send error if not cancelled
			errChan <- err
		}
	}()

	// Wait for push completion or cancellation
	select {
	case err = <-errChan:
		if err != nil {
			// Check if it's a sync conflict error
			if conflictErr, ok := err.(*sync_service.SyncConflictError); ok {
				close(progressChan)
				// Emit conflict event to frontend with details
				app.Event.Emit("sync-conflict", map[string]interface{}{
					"projectPath": projectPath,
					"remoteURL":   remoteURL,
					"conflicts":   conflictErr.Conflicts,
				})
				return errors.New("sync_conflict")
			}

			if errors.Is(err, syscall.ECONNREFUSED) {
				println(err.Error())
				return errors.New("syncing failed, connection refused")
			}
			return err
		}
	case <-ctx.Done():
		close(progressChan) // Stop progress updates
		return errors.New("cancelled")
	}

	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      "Syncing",
		Message:    "Receiving",
		Percentage: 0,
		Current:    2,
		Total:      2,
	}:
	}

	pullCallBack := func(current int, total int, message string, extraMessage string) {
		if ctx.Err() != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case progressChan <- output.ProgressReport{
			Title:        "Syncing",
			Message:      message,
			Percentage:   (float64(current) / float64(total) * 99),
			Current:      2,
			Total:        2,
			ExtraMessage: extraMessage,
		}:
		default: // Skip progress update if channel is full
		}
	}

	// Signal start of pull operation
	select {
	case <-ctx.Done():
		close(progressChan)
		return errors.New("sync operation cancelled before pull")
	case progressChan <- output.ProgressReport{
		Message:    "Receiving",
		Current:    2,
		Total:      2,
		Percentage: 0,
	}:
	}

	// Pull data with cancellation
	go func() {
		err := sync_service.PullData(ctx, projectPath, remoteURL, activeUser.Id, pullChunk, syncOptions, pullCallBack)
		if ctx.Err() == nil { // Only send error if not cancelled
			errChan <- err
		}
	}()

	// Wait for pull completion or cancellation
	select {
	case err = <-errChan:
		if err != nil {
			if errors.Is(err, syscall.ECONNREFUSED) {
				return errors.New("syncing failed, connection refused")
			}
			return err
		}
	case <-ctx.Done():
		close(progressChan) // Stop progress updates
		return errors.New("cancelled")
	}

	// Final progress update
	select {
	case <-ctx.Done():
		return errors.New("sync operation cancelled during completion")
	case progressChan <- output.ProgressReport{
		Message:    "Completing Sync",
		Current:    2,
		Total:      2,
		Percentage: 100,
	}:
	}

	close(progressChan)
	InvalidateRemoteCache(projectPath)
	return nil
}

func (s *SyncService) PullLatestCheckpoints(projectPath, remoteURL string) error {
	defer reset() // Ensure context is reset when we're done

	ctx := getContext()
	if ctx.Err() != nil {
		return errors.New("operation cancelled before starting")
	}

	app := application.Get()
	activeUser, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	// Create buffered channels to prevent blocking
	errChan := make(chan error, 1)
	progressChan := make(chan output.ProgressReport, 10)

	// Start progress update goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return // Exit immediately on cancellation
			case progress, ok := <-progressChan:
				if !ok {
					return
				}
				app.Event.Emit("progress-update", progress)
			}
		}
	}()

	// Initial progress
	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      "Syncing",
		Message:    "Sending",
		Percentage: 0,
		Current:    1,
		Total:      1,
	}:
	}

	pullCallBack := func(current int, total int, message string, extraMessage string) {
		if ctx.Err() != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case progressChan <- output.ProgressReport{
			Title:        "Downloading Latest Checkpoints",
			Message:      message,
			Percentage:   (float64(current) / float64(total) * 99),
			Current:      1,
			Total:        1,
			ExtraMessage: extraMessage,
		}:
		default: // Skip progress update if channel is full
		}
	}

	// Signal start of pull operation
	select {
	case <-ctx.Done():
		close(progressChan)
		return errors.New("sync operation cancelled before pull")
	case progressChan <- output.ProgressReport{
		Message:    "Receiving",
		Current:    1,
		Total:      1,
		Percentage: 0,
	}:
	}

	// Pull data with cancellation
	go func() {
		err := sync_service.PullLatestCheckpoints(ctx, projectPath, remoteURL, activeUser.Id, pullCallBack)
		if ctx.Err() == nil { // Only send error if not cancelled
			errChan <- err
		}
	}()

	// Wait for pull completion or cancellation
	select {
	case err = <-errChan:
		if err != nil {
			if errors.Is(err, syscall.ECONNREFUSED) {
				return errors.New("syncing failed, connection refused")
			}
			return err
		}
	case <-ctx.Done():
		close(progressChan) // Stop progress updates
		return errors.New("cancelled")
	}

	// Final progress update
	select {
	case <-ctx.Done():
		return errors.New("sync operation cancelled during completion")
	case progressChan <- output.ProgressReport{
		Message:    "Completing Sync",
		Current:    1,
		Total:      1,
		Percentage: 100,
	}:
	}

	close(progressChan)
	return nil
}

func (s *SyncService) PullData(projectPath string, remoteURL string, pullChunk bool, syncOptions sync_service.SyncOptions) error {
	defer reset() // Ensure context is reset when we're done

	ctx := getContext()
	if ctx.Err() != nil {
		return errors.New("operation cancelled before starting")
	}
	app := application.Get()

	activeUser, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}
	progress := output.ProgressReport{
		Title:      "Syncing",
		Message:    "Receiving",
		Percentage: 0,
		Current:    1,
		Total:      1,
	}
	app.Event.Emit("progress-update", progress)

	pullCallBack := func(current int, total int, message string, extraMessage string) {
		progress := output.ProgressReport{
			Title:        progress.Title,
			Message:      message,
			Percentage:   (float64(current) / float64(total) * 99),
			Current:      1,
			Total:        1,
			ExtraMessage: extraMessage,
		}
		app.Event.Emit("progress-update", progress)
	}
	err = sync_service.PullData(ctx, projectPath, remoteURL, activeUser.Id, pullChunk, syncOptions, pullCallBack)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			return errors.New("syncing failed, connection refused")
		}
		return err
	}

	progress.Message = "Completing Sync"
	progress.Current = 1
	progress.Percentage = 100
	app.Event.Emit("progress-update", progress)
	InvalidateRemoteCache(projectPath)
	return nil
}

func (s *SyncService) PushCheckpoints(projectPath string, remoteURL string, pullChunk bool, syncOptions sync_service.SyncOptions) error {
	app := application.Get()

	activeUser, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}
	progress := output.ProgressReport{
		Title:      "Sending",
		Message:    "Sending",
		Percentage: 0,
		Current:    1,
		Total:      1,
	}
	app.Event.Emit("progress-update", progress)

	pushCallBack := func(current int, total int, message string, extraMessage string) {

		progress := output.ProgressReport{
			Title:        progress.Title,
			Message:      message,
			Percentage:   (float64(current) / float64(total) * 99),
			Current:      1,
			Total:        1,
			ExtraMessage: extraMessage,
		}
		app.Event.Emit("progress-update", progress)
	}
	err = sync_service.PushData(projectPath, remoteURL, activeUser.Id, pushCallBack)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			return errors.New("pushing failed, connection refused")
		}
		return errors.New("pushing failed, check your connection")
	}

	progress.Message = "Completing Sync"
	progress.Current = 1
	progress.Percentage = 100
	app.Event.Emit("progress-update", progress)
	return nil
}

func (s *SyncService) DownloadCheckpoint(projectPath, remoteURL, checkpointId string) error {
	defer reset() // Ensure context is reset when we're done

	ctx := getContext()
	if ctx.Err() != nil {
		return errors.New("operation cancelled before starting")
	}

	app := application.Get()
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	// Create buffered channels to prevent blocking
	errChan := make(chan error, 1)
	progressChan := make(chan output.ProgressReport, 10)

	// Start progress update goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return // Exit immediately on cancellation
			case progress, ok := <-progressChan:
				if !ok {
					return
				}
				app.Event.Emit("progress-update", progress)
			}
		}
	}()

	callBack := func(current int, total int, message string, extraMessage string) {
		if ctx.Err() != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case progressChan <- output.ProgressReport{
			Title:        "Downloading Checkpoint",
			Message:      message,
			Percentage:   (float64(current) / float64(total) * 99),
			Current:      1,
			Total:        1,
			ExtraMessage: extraMessage,
		}:
		default: // Skip progress update if channel is full
		}
	}

	go func() {
		err := sync_service.DownloadCheckpoint(ctx, projectPath, remoteURL, checkpointId, user.Id, callBack)
		if ctx.Err() == nil { // Only send error if not cancelled
			errChan <- err
		}
	}()

	select {
	case err = <-errChan:
		if err != nil {
			if errors.Is(err, syscall.ECONNREFUSED) {
				return errors.New("download failed, connection refused")
			}
			return errors.New("download failed, check your connection")
		}
	case <-ctx.Done():
		close(progressChan) // Stop progress updates
		return errors.New("cancelled")
	}

	close(progressChan)
	progress := output.ProgressReport{
		Title:      "Downloading Checkpoint",
		Message:    "Receiving",
		Percentage: 100,
		Current:    1,
		Total:      1,
	}
	app.Event.Emit("progress-update", progress)
	return nil
}

func (s *SyncService) IsUnsynced(projectPath string) (bool, error) {
	if !utils.FileExists(projectPath) {
		return false, error_service.ErrProjectNotFound
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return false, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	isUnsynced, err := sync_service.IsUnsynced(tx)
	if err != nil {
		return false, err
	}
	return isUnsynced, nil
}

// ResolveConflicts resolves sync conflicts by remapping local IDs to match server IDs.
// This should be called after the user accepts the conflict resolution in the UI.
// Accepts conflicts as a JSON string since Wails binding may have issues with complex slice types.
func (s *SyncService) ResolveConflicts(projectPath string, conflictsJSON string) error {
	if !utils.FileExists(projectPath) {
		return error_service.ErrProjectNotFound
	}

	// Parse the JSON string into []ConflictInfo
	var conflicts []sync_service.ConflictInfo
	if err := json.Unmarshal([]byte(conflictsJSON), &conflicts); err != nil {
		return fmt.Errorf("failed to parse conflicts JSON: %w", err)
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = sync_service.ResolveConflicts(tx, conflicts)
	if err != nil {
		return fmt.Errorf("failed to resolve conflicts: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit conflict resolution: %w", err)
	}

	return nil
}

// GetPendingChanges returns a lightweight summary of all unsynced changes in the project.
// Used by the ChangeLog pane to display pending changes without loading full row data.
func (s *SyncService) GetPendingChanges(projectPath string) (sync_service.ChangeSummary, error) {
	if !utils.FileExists(projectPath) {
		return sync_service.ChangeSummary{}, error_service.ErrProjectNotFound
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return sync_service.ChangeSummary{}, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return sync_service.ChangeSummary{}, err
	}
	defer tx.Rollback()

	summary, err := sync_service.LoadChangeSummary(tx)
	if err != nil {
		return sync_service.ChangeSummary{}, err
	}
	return summary, nil
}

// DiscardChanges reverts specific items to their server state by fetching remote data
// and selectively replacing local rows. itemType should be "asset" or "collection".
func (s *SyncService) DiscardChanges(projectPath, remoteURL string, itemIds []string, itemType string) error {
	if !utils.FileExists(projectPath) {
		return error_service.ErrProjectNotFound
	}

	activeUser, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	serverData, err := sync_service.FetchData(remoteURL, activeUser.Id)
	if err != nil {
		return fmt.Errorf("failed to fetch server data: %w", err)
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, itemId := range itemIds {
		switch itemType {
		case "asset":
			err = sync_service.DiscardAssetChanges(tx, serverData, itemId)
		case "collection":
			err = sync_service.DiscardCollectionChanges(tx, serverData, itemId)
		default:
			return fmt.Errorf("unsupported item type: %s", itemType)
		}
		if err != nil {
			return fmt.Errorf("failed to discard changes for %s %s: %w", itemType, itemId, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit discard: %w", err)
	}
	return nil
}

// DiscardAllChanges reverts all unsynced changes to the server state.
// This replaces the nuclear PullData(force=true) approach with selective replacement.
func (s *SyncService) DiscardAllChanges(projectPath, remoteURL string) error {
	if !utils.FileExists(projectPath) {
		return error_service.ErrProjectNotFound
	}

	activeUser, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	serverData, err := sync_service.FetchData(remoteURL, activeUser.Id)
	if err != nil {
		return fmt.Errorf("failed to fetch server data: %w", err)
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = sync_service.DiscardAllChanges(tx, serverData)
	if err != nil {
		return fmt.Errorf("failed to discard all changes: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit discard all: %w", err)
	}
	return nil
}

// SyncAsset pushes a single asset and its checkpoints (including chunks and previews) to the server.
// This is a user-initiated action that bypasses the write-through gate.
func (s *SyncService) SyncAsset(projectPath, remoteURL, assetId string) error {
	defer reset()

	ctx := getContext()
	if ctx.Err() != nil {
		return errors.New("operation cancelled before starting")
	}

	app := application.Get()
	activeUser, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	errChan := make(chan error, 1)
	progressChan := make(chan output.ProgressReport, 10)

	// Progress update goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case progress, ok := <-progressChan:
				if !ok {
					return
				}
				app.Event.Emit("progress-update", progress)
			}
		}
	}()

	// Initial progress
	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      "Syncing Asset",
		Message:    "Sending",
		Percentage: 0,
		Current:    1,
		Total:      1,
	}:
	}

	pushCallback := func(current int, total int, message string, extraMessage string) {
		if ctx.Err() != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case progressChan <- output.ProgressReport{
			Title:        "Syncing Asset",
			Message:      message,
			Percentage:   (float64(current) / float64(total) * 99),
			Current:      1,
			Total:        1,
			ExtraMessage: extraMessage,
		}:
		default:
		}
	}

	go func() {
		err := sync_service.PushAssetData(projectPath, remoteURL, activeUser.Id, assetId, pushCallback)
		if ctx.Err() == nil {
			errChan <- err
		}
	}()

	select {
	case err = <-errChan:
		if err != nil {
			if conflictErr, ok := err.(*sync_service.SyncConflictError); ok {
				close(progressChan)
				app.Event.Emit("sync-conflict", map[string]interface{}{
					"projectPath": projectPath,
					"remoteURL":   remoteURL,
					"conflicts":   conflictErr.Conflicts,
				})
				return errors.New("sync_conflict")
			}
			close(progressChan)
			return err
		}
	case <-ctx.Done():
		close(progressChan)
		return errors.New("cancelled")
	}

	select {
	case <-ctx.Done():
		return errors.New("operation cancelled during completion")
	case progressChan <- output.ProgressReport{
		Title:      "Syncing Asset",
		Message:    "Complete",
		Percentage: 100,
		Current:    1,
		Total:      1,
	}:
	}

	close(progressChan)
	return nil
}
