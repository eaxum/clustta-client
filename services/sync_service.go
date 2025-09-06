package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/error_service"
	"clustta/internal/repository/sync_service"
	"clustta/internal/settings"
	"clustta/internal/utils"
	"clustta/output"
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
				app.EmitEvent("progress-update", progress)
			}
		}
	}()

	// Initial progress
	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      fmt.Sprintf("Downloading %s  Project", projectName),
		Message:    fmt.Sprintf("Downloading %s  Project", projectName),
		Percentage: 0,
		Current:    1,
		Total:      1,
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
			Title:        fmt.Sprintf("Downloading %s  Project", projectName),
			Message:      message,
			Percentage:   float64(current) / float64(total) * 98,
			Current:      1,
			Total:        1,
			ExtraMessage: extraMessage,
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
		Title:      fmt.Sprintf("Downloading %s  Project", projectName),
		Message:    fmt.Sprintf("Downloading %s  Project", projectName),
		Percentage: 100,
		Current:    1,
		Total:      1,
	}
	app.EmitEvent("progress-update", progress)
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
				app.EmitEvent("progress-update", progress)
			}
		}
	}()

	// Initial progress
	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      "Syncing",
		Message:    "Pushing Data",
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
		Message:    "Pulling Data",
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
		Message:    "Pulling Data",
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
				app.EmitEvent("progress-update", progress)
			}
		}
	}()

	// Initial progress
	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      "Syncing",
		Message:    "Pushing Data",
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
		Message:    "Pulling Data",
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
		Message:    "Pulling Data",
		Percentage: 0,
		Current:    1,
		Total:      1,
	}
	app.EmitEvent("progress-update", progress)

	pullCallBack := func(current int, total int, message string, extraMessage string) {
		progress := output.ProgressReport{
			Title:        progress.Title,
			Message:      message,
			Percentage:   (float64(current) / float64(total) * 99),
			Current:      1,
			Total:        1,
			ExtraMessage: extraMessage,
		}
		app.EmitEvent("progress-update", progress)
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
	app.EmitEvent("progress-update", progress)
	return nil
}

func (s *SyncService) PushCheckpoints(projectPath string, remoteURL string, pullChunk bool, syncOptions sync_service.SyncOptions) error {
	app := application.Get()

	activeUser, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}
	progress := output.ProgressReport{
		Title:      "Pushing",
		Message:    "Pushing Data",
		Percentage: 0,
		Current:    1,
		Total:      1,
	}
	app.EmitEvent("progress-update", progress)

	pushCallBack := func(current int, total int, message string, extraMessage string) {

		progress := output.ProgressReport{
			Title:        progress.Title,
			Message:      message,
			Percentage:   (float64(current) / float64(total) * 99),
			Current:      1,
			Total:        1,
			ExtraMessage: extraMessage,
		}
		app.EmitEvent("progress-update", progress)
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
	app.EmitEvent("progress-update", progress)
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
				app.EmitEvent("progress-update", progress)
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
		Message:    "Pulling Data",
		Percentage: 100,
		Current:    1,
		Total:      1,
	}
	app.EmitEvent("progress-update", progress)
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
