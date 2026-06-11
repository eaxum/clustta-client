package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/error_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/sync_service"
	"clustta/internal/utils"
	"clustta/output"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type CheckpointService struct{}

// isImageFile checks whether a file path has an image extension.
func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".tiff", ".webp":
		return true
	}
	return false
}

// DeleteCheckpoint removes a checkpoint from the project.
// Returns an error if the deletion fails.
func (c *CheckpointService) DeleteCheckpoint(projectPath, checkpointId string) error {
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

	err = repository.DeleteCheckpoint(
		tx,
		checkpointId,
		true,
		true,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// RevertToCheckpoint reverts a asset to a specific checkpoint state.
// Downloads missing chunks if needed and supports cancellation.
func (c *CheckpointService) RevertToCheckpoint(projectPath, remoteUrl, assetId, checkpointId string) error {
	defer reset()

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

	// Initial progress
	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      "Reverting Asset",
		Message:    "Preparing to Revert",
		Percentage: 0,
		Current:    1,
		Total:      2,
	}:
	}

	asset, err := repository.GetAsset(tx, assetId)
	if err != nil {
		return err
	}

	checkpoint, err := repository.GetCheckpoint(tx, checkpointId)
	if err != nil {
		return err
	}
	isMisssingChunks, err := checkpoint.HasMissingChunks(tx)
	if err != nil {
		return err
	}

	err = tx.Rollback()
	if err != nil {
		return err
	}

	if isMisssingChunks {
		callBack := func(current int, total int, message string, extraMessage string) {
			if ctx.Err() != nil {
				return
			}
			currentSize := utils.BytesToHumanReadable(current)
			totalSize := utils.BytesToHumanReadable(total)
			select {
			case <-ctx.Done():
				return
			case progressChan <- output.ProgressReport{
				Title:      "Downloading Checkpoint",
				Message:    fmt.Sprintf("Receiving %s/%s", currentSize, totalSize),
				Percentage: (float64(current) / float64(total) * 99),
				Current:    1,
				Total:      1,
			}:
			default: // Skip progress update if channel is full
			}
		}

		go func() {
			err := sync_service.DownloadCheckpoint(ctx, projectPath, remoteUrl, checkpointId, user.Id, callBack)
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
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	tx, err = dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	callBack := func(current int, total int, message string, extraMessage string) {
		progress := output.ProgressReport{
			Title:      "Reverting Asset",
			Message:    asset.Name,
			Percentage: float64(current) / float64(total) * 100,
			Current:    1,
			Total:      1,
		}
		app.Event.Emit("progress-update", progress)
	}
	err = repository.RevertToCheckpoint(tx, checkpointId, asset.FilePath, callBack)
	if err != nil {
		return err
	}

	close(progressChan)
	progress := output.ProgressReport{
		Title:      "Reverting Asset",
		Message:    asset.Name,
		Percentage: 100,
		Current:    1,
		Total:      1,
	}
	app.Event.Emit("progress-update", progress)
	return nil
}

// AddCheckpoint creates new checkpoints for multiple assets.
// Returns the created checkpoints or an error if the operation fails.
func (c *CheckpointService) AddCheckpoint(projectPath string, assetPaths, extensions []string, message, previewPath, groupId string, useAsThumbnail, sendToIntegration bool) ([]models.Checkpoint, error) {
	app := application.Get()
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Checkpoint{}, err
	}
	defer dbConn.Close()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return []models.Checkpoint{}, err
	}
	authorId := user.Id

	totalAssets := len(assetPaths)
	previewId := ""
	if previewPath != "" && isImageFile(previewPath) {
		tx, err := dbConn.Beginx()
		if err != nil {
			return []models.Checkpoint{}, err
		}

		preview, err := repository.CreatePreview(tx, previewPath)
		if err != nil {
			tx.Rollback()
			return []models.Checkpoint{}, err
		}
		previewId = preview.Hash
		err = tx.Commit()
		if err != nil {
			return []models.Checkpoint{}, err
		}
	}
	checkpoints := []models.Checkpoint{}
	for i, assetPath := range assetPaths {
		tx, err := dbConn.Beginx()
		if err != nil {
			return []models.Checkpoint{}, err
		}

		asset, err := repository.GetAssetByPath(tx, assetPath, extensions[i])
		if err != nil {
			return []models.Checkpoint{}, err
		}
		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Creating Checkpoint",
				Message:    asset.Name,
				Percentage: float64(current) / float64(total) * 99,
				Current:    i + 1,
				Total:      totalAssets,
			}
			app.Event.Emit("progress-update", progress)
		}

		checkpoint, err := repository.CreateCheckpoint(
			tx,
			asset.Id,
			message,
			"",
			"",
			0,
			0,
			asset.FilePath,
			authorId,
			previewId,
			groupId,
			callBack,
		)
		if err != nil {
			tx.Rollback()
			return []models.Checkpoint{}, err
		}
		if previewId != "" && useAsThumbnail {
			err = repository.SetCollectionPreview(tx, asset.Id, "asset", previewId)
			if err != nil {
				tx.Rollback()
				return []models.Checkpoint{}, err
			}
		}
		checkpoints = append(checkpoints, checkpoint)
		err = tx.Commit()
		if err != nil {
			return []models.Checkpoint{}, err
		}
	}

	// Push to external integration in background (preview upload + status sync)
	// Only for single-asset checkpoints to avoid flooding the external system
	willPushToIntegration := sendToIntegration && len(checkpoints) == 1 && previewPath != ""

	if !willPushToIntegration {
		progress := output.ProgressReport{
			Title:      "Creating Checkpoint",
			Message:    "finishing up",
			Percentage: 100,
			Current:    totalAssets,
			Total:      totalAssets,
			EntityData: checkpoints,
		}
		app.Event.Emit("progress-update", progress)
	}

	if sendToIntegration && len(checkpoints) == 1 {
		go func() {
			integrationSvc := &IntegrationService{}
			if err := integrationSvc.PushToIntegration(projectPath, []string{checkpoints[0].AssetId}, checkpoints[0].Id, previewPath, message); err != nil {
				log.Printf("integration push failed (checkpoint still created): %v", err)
				app.Event.Emit("integration-push-failed", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}()
	}

	return checkpoints, nil
}

// AddUntrackedAsset tracks previously untracked files and creates checkpoints for them.
// Returns the newly tracked assets or an error if the operation fails.
func (c *CheckpointService) AddUntrackedAsset(projectPath, projectWorkingDir string, assetPaths []string, completed, totalAssets int, message, previewPath, groupId string) ([]models.Asset, error) {
	app := application.Get()
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Asset{}, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Asset{}, err
	}
	defer tx.Rollback()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return []models.Asset{}, err
	}
	userData, err := repository.GetUser(tx, user.Id)
	if err != nil {
		return []models.Asset{}, err
	}
	userRole, err := repository.GetRole(tx, userData.RoleId)
	if err != nil {
		return []models.Asset{}, err
	}
	if !userRole.CreateAsset {
		return []models.Asset{}, error_service.ErrNotUnauthorized
	}

	collections := []models.Collection{}
	err = tx.Select(&collections, "SELECT id, collection_path FROM full_collection")
	if err != nil {
		return []models.Asset{}, err
	}
	collectionType, err := repository.GetCollectionTypeByName(tx, "generic")
	if err != nil {
		return []models.Asset{}, err
	}
	assetType, err := repository.GetAssetTypeByName(tx, "generic")
	if err != nil {
		return []models.Asset{}, err
	}
	status, err := repository.GetStatusByShortName(tx, "todo")
	if err != nil {
		return []models.Asset{}, err
	}
	statusId := status.Id
	tx.Rollback()

	collectionPathsIndex := map[string]string{}
	for _, collection := range collections {
		collectionPathsIndex[collection.CollectionPath] = collection.Id
	}

	previewId := ""
	if previewPath != "" {
		tx, err := dbConn.Beginx()
		if err != nil {
			return []models.Asset{}, err
		}

		preview, err := repository.CreatePreview(tx, previewPath)
		if err != nil {
			tx.Rollback()
			return []models.Asset{}, err
		}
		previewId = preview.Hash
		err = tx.Commit()
		if err != nil {
			return []models.Asset{}, err
		}
	}
	assets := []models.Asset{}
	state := completed
	for i, assetPath := range assetPaths {
		tx, err := dbConn.Beginx()
		if err != nil {
			return []models.Asset{}, err
		}
		defer tx.Rollback()

		assetFilePath := filepath.Join(projectWorkingDir, assetPath)
		collectionPath := utils.GetParent(assetPath)
		assetCollectionId := ""

		for _, curentCollectionPath := range utils.GetCollectionPaths(collectionPath) {
			collectionId, exists := collectionPathsIndex[curentCollectionPath]
			if !exists {
				collectionName := filepath.Base(curentCollectionPath)
				parentCollectionId := collectionPathsIndex[utils.GetParent(curentCollectionPath)]
				collectionId = uuid.New().String()
				err = repository.CreateCollectionFast(tx, collectionId, collectionName, "", collectionType.Id, parentCollectionId, "", false)
				if err != nil {
					return []models.Asset{}, err
				}
				collectionPathsIndex[curentCollectionPath] = collectionId
			}
			assetCollectionId = collectionId
		}

		assetName := strings.TrimSuffix(filepath.Base(assetPath), filepath.Ext(assetPath))

		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Creating Checkpoint",
				Message:    assetName,
				Percentage: float64(current) / float64(total) * 99,
				Current:    completed + (i + 1),
				Total:      totalAssets,
			}
			app.Event.Emit("progress-update", progress)
		}
		err = repository.CreateAssetFast(tx, "", assetName, assetType.Id, assetCollectionId, true, "", assetFilePath, previewId, user.Id, message, groupId, assetPath, statusId, callBack)
		if err != nil {
			tx.Rollback()
			return []models.Asset{}, err
		}

		err = tx.Commit()
		if err != nil {
			return []models.Asset{}, err
		}
		state = completed + (i + 1)
	}

	if state == totalAssets {
		progress := output.ProgressReport{
			Title:      "Creating Checkpoint",
			Message:    "finishing up",
			Percentage: 100,
			Current:    totalAssets,
			Total:      totalAssets,
		}
		app.Event.Emit("progress-update", progress)
	}
	return assets, nil
}

// ViewCheckpoint creates a temporary file from a checkpoint and opens it.
// The temp file is placed in the same directory as the original asset so relative dependencies resolve correctly.
func (c *CheckpointService) ViewCheckpoint(projectPath, checkpointId, assetId, collectionName, extension string) error {
	app := application.Get()
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

	asset, err := repository.GetAsset(tx, assetId)
	if err != nil {
		return err
	}
	assetDir := filepath.Dir(asset.GetFilePath())
	assetName := strings.TrimSuffix(filepath.Base(asset.GetFilePath()), extension)
	tempFile := filepath.Join(assetDir, fmt.Sprintf("%s-checkpoint-%s%s", assetName, checkpointId[:4], extension))
	callBack := func(current int, total int, message string, extraMessage string) {
		progress := output.ProgressReport{
			Title:      "Preparing Checkpoint",
			Message:    collectionName,
			Percentage: float64(current) / float64(total) * 100,
			Current:    1,
			Total:      1,
		}
		app.Event.Emit("progress-update", progress)
	}
	err = repository.RevertToCheckpoint(tx, checkpointId, tempFile, callBack)
	if err != nil {
		return err
	}

	err = utils.LaunchFile(tempFile)
	if err != nil {
		return err
	}
	return nil
}

// GetCheckpoints retrieves all checkpoints for a specific asset.
// Returns the list of checkpoints or an error if the operation fails.
func (c *CheckpointService) GetCheckpoints(projectPath, assetId string) ([]models.Checkpoint, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Checkpoint{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Checkpoint{}, err
	}
	defer tx.Rollback()

	checkPoints, err := repository.GetCheckpoints(tx, assetId, false)
	if err != nil {
		return []models.Checkpoint{}, err
	}
	return checkPoints, nil
}

// GetLatestCheckpoint retrieves the most recent checkpoint for a asset.
// Returns the latest checkpoint or an error if not found.
func (c *CheckpointService) GetLatestCheckpoint(projectPath, assetId string) (models.Checkpoint, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Checkpoint{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Checkpoint{}, err
	}
	defer tx.Rollback()

	checkpoint, err := repository.GetLatestCheckpoint(tx, assetId)
	if err != nil {
		return models.Checkpoint{}, err
	}
	return checkpoint, nil
}

// GetTimeline retrieves the project timeline showing checkpoint history.
// Returns the timeline data or an error if the operation fails.
func (c *CheckpointService) GetTimeline(projectPath string) ([]repository.CompatTimeline, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []repository.CompatTimeline{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []repository.CompatTimeline{}, err
	}
	defer tx.Rollback()

	timeline, err := repository.GetTimeline(tx)
	if err != nil {
		return []repository.CompatTimeline{}, err
	}
	return timeline, nil
}

// Revert reverts multiple assets to their latest checkpoints.
// Downloads missing chunks if needed and supports cancellation.
func (c *CheckpointService) Revert(projectPath, remoteUrl string, assetIds []string) error {
	defer reset()

	ctx := getContext()
	if ctx.Err() != nil {
		return errors.New("operation cancelled before starting")
	}

	app := application.Get()
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	errChan := make(chan error, 1)
	progressChan := make(chan output.ProgressReport, 10)

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

	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      "Reverting",
		Message:    "Preparing to Revert",
		Percentage: 0,
		Current:    1,
		Total:      1,
	}:
	}

	if len(assetIds) == 0 {
		return nil
	}

	checkpointQuery, checkpointArgs, err := sqlx.In(
		"SELECT * FROM asset_checkpoint WHERE trashed = 0 AND asset_id IN (?) ORDER BY created_at DESC",
		assetIds,
	)
	if err != nil {
		return err
	}
	checkpointQuery = tx.Rebind(checkpointQuery)
	checkpoints := []models.Checkpoint{}
	err = tx.Select(&checkpoints, checkpointQuery, checkpointArgs...)
	if err != nil {
		return err
	}
	assetCheckpoints := map[string][]models.Checkpoint{}
	for _, assetCheckpoint := range checkpoints {
		assetCheckpoints[assetCheckpoint.AssetId] = append(assetCheckpoints[assetCheckpoint.AssetId], assetCheckpoint)
	}

	revertableAssetIds := make([]string, 0, len(assetIds))
	checkpointIdsToDownload := []string{}
	for _, assetId := range assetIds {
		assetCheckpointList, ok := assetCheckpoints[assetId]
		if !ok || len(assetCheckpointList) == 0 {
			// Skip assets with no checkpoints (e.g. brand-new placeholder dependencies).
			continue
		}
		revertableAssetIds = append(revertableAssetIds, assetId)
		latestCheckpoint := assetCheckpointList[0]
		isMisssingChunks, err := latestCheckpoint.HasMissingChunks(tx)
		if err != nil {
			return err
		}
		if isMisssingChunks {
			checkpointIdsToDownload = append(checkpointIdsToDownload, latestCheckpoint.Id)
		}
	}

	err = tx.Rollback()
	if err != nil {
		return err
	}

	if len(checkpointIdsToDownload) != 0 {
		callBack := func(current int, total int, message string, extraMessage string) {
			if ctx.Err() != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			case progressChan <- output.ProgressReport{
				Title:        "Downloading files",
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
			err := sync_service.DownloadCheckpoints(ctx, projectPath, remoteUrl, checkpointIdsToDownload, user.Id, callBack)
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
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	totalAssets := len(revertableAssetIds)
	for i, assetId := range revertableAssetIds {
		tx, err := dbConn.Beginx()
		if err != nil {
			return err
		}
		asset, err := repository.GetAsset(tx, assetId)
		if err != nil {
			return err
		}
		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Reverting",
				Message:    asset.Name,
				Percentage: float64(current) / float64(total) * 100,
				Current:    i + 1,
				Total:      totalAssets,
			}
			app.Event.Emit("progress-update", progress)
		}

		err = repository.RevertToLatestCheckpoint(tx, assetId, asset.FilePath, callBack)
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Rollback()
	}

	close(progressChan)
	progress := output.ProgressReport{
		Title:      "Reverting",
		Message:    "Reverting",
		Percentage: 100,
		Current:    1,
		Total:      1,
	}
	app.Event.Emit("progress-update", progress)
	return nil
}

// RevertAssetPaths reverts assets by their file paths to latest checkpoints.
// Downloads missing chunks if needed and supports cancellation.
func (c *CheckpointService) RevertAssetPaths(projectPath, remoteUrl string, assetPaths []string) error {
	defer reset()

	ctx := getContext()
	if ctx.Err() != nil {
		return errors.New("operation cancelled before starting")
	}

	app := application.Get()
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	errChan := make(chan error, 1)
	progressChan := make(chan output.ProgressReport, 10)

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

	quotedAssetPaths := make([]string, len(assetPaths))
	for i, assetPath := range assetPaths {
		quotedAssetPaths[i] = fmt.Sprintf("\"%s\"", assetPath)
	}

	assetIds := []string{}
	err = tx.Select(&assetIds, fmt.Sprintf("SELECT id FROM full_asset WHERE trashed = 0 AND (asset_path || extension) IN (%s) ORDER BY created_at DESC", strings.Join(quotedAssetPaths, ",")))
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      "Reverting",
		Message:    "Preparing to Revert",
		Percentage: 0,
		Current:    1,
		Total:      1,
	}:
	}

	quotedAssetIds := make([]string, len(assetIds))
	for i, id := range assetIds {
		quotedAssetIds[i] = fmt.Sprintf("\"%s\"", id)
	}
	checkpoints := []models.Checkpoint{}
	err = tx.Select(&checkpoints, fmt.Sprintf("SELECT * FROM asset_checkpoint WHERE trashed = 0 AND asset_id IN (%s) ORDER BY created_at DESC", strings.Join(quotedAssetIds, ",")))
	if err != nil {
		return err
	}
	assetCheckpoints := map[string][]models.Checkpoint{}
	for _, assetCheckpoint := range checkpoints {
		assetCheckpoints[assetCheckpoint.AssetId] = append(assetCheckpoints[assetCheckpoint.AssetId], assetCheckpoint)
	}

	checkpointIdsToDownload := []string{}
	for _, assetId := range assetIds {
		latestCheckpoint := assetCheckpoints[assetId][0]
		isMisssingChunks, err := latestCheckpoint.HasMissingChunks(tx)
		if err != nil {
			return err
		}
		if isMisssingChunks {
			checkpointIdsToDownload = append(checkpointIdsToDownload, latestCheckpoint.Id)
		}
	}

	err = tx.Rollback()
	if err != nil {
		return err
	}

	if len(checkpointIdsToDownload) != 0 {
		callBack := func(current int, total int, message string, extraMessage string) {
			if ctx.Err() != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			case progressChan <- output.ProgressReport{
				Title:        "Downloading files",
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
			err := sync_service.DownloadCheckpoints(ctx, projectPath, remoteUrl, checkpointIdsToDownload, user.Id, callBack)
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
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	totalAssets := len(assetIds)
	for i, assetId := range assetIds {
		tx, err := dbConn.Beginx()
		if err != nil {
			return err
		}
		asset, err := repository.GetAsset(tx, assetId)
		if err != nil {
			return err
		}
		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Reverting",
				Message:    asset.Name,
				Percentage: float64(current) / float64(total) * 100,
				Current:    i + 1,
				Total:      totalAssets,
			}
			app.Event.Emit("progress-update", progress)
		}

		err = repository.RevertToLatestCheckpoint(tx, assetId, asset.FilePath, callBack)
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Rollback()
	}

	close(progressChan)
	progress := output.ProgressReport{
		Title:      "Reverting",
		Message:    "Reverting",
		Percentage: 100,
		Current:    1,
		Total:      1,
	}
	app.Event.Emit("progress-update", progress)
	return nil
}

func (c *CheckpointService) RevertProject(projectPath, remoteUrl string, checkpointTime string) error {
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

	// Initial progress
	select {
	case <-ctx.Done():
		return errors.New("operation cancelled")
	case progressChan <- output.ProgressReport{
		Title:      "Reverting",
		Message:    "Preparing to Revert",
		Percentage: 0,
		Current:    1,
		Total:      1,
	}:
	}

	unixTime, err := time.Parse(time.RFC3339, checkpointTime)
	if err != nil {
		return err
	}
	checkpoints, err := repository.GetLatestCheckpointsByTime(tx, unixTime.Unix())
	if err != nil {
		return err
	}

	checkpointIdsToDownload := []string{}
	for _, checkpoint := range checkpoints {
		isMisssingChunks, err := checkpoint.HasMissingChunks(tx)
		if err != nil {
			return err
		}
		if isMisssingChunks {
			checkpointIdsToDownload = append(checkpointIdsToDownload, checkpoint.Id)
		}
	}

	err = tx.Rollback()
	if err != nil {
		return err
	}

	if len(checkpointIdsToDownload) != 0 {
		callBack := func(current int, total int, message string, extraMessage string) {
			if ctx.Err() != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			case progressChan <- output.ProgressReport{
				Title:        "Downloading files",
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
			err := sync_service.DownloadCheckpoints(ctx, projectPath, remoteUrl, checkpointIdsToDownload, user.Id, callBack)
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
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	totalAssets := len(checkpoints)
	for i, checkpoint := range checkpoints {
		tx, err := dbConn.Beginx()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		asset, err := repository.GetAsset(tx, checkpoint.AssetId)
		if err != nil {
			tx.Rollback()
			return err
		}
		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Reverting",
				Message:    asset.Name,
				Percentage: float64(current) / float64(total) * 100,
				Current:    i + 1,
				Total:      totalAssets,
			}
			app.Event.Emit("progress-update", progress)
		}
		if utils.FileExists(asset.FilePath) {
			fileXXHash, err := utils.GenerateXXHashChecksum(asset.FilePath)
			if err != nil {
				return err
			}
			if fileXXHash == checkpoint.XXHashChecksum {
				progress := output.ProgressReport{
					Title:      "Reverting",
					Message:    asset.Name,
					Percentage: 100,
					Current:    i + 1,
					Total:      totalAssets,
				}
				app.Event.Emit("progress-update", progress)
				tx.Rollback()
				continue // Skip if the file is already in the correct state
			}
		}
		err = repository.RevertToCheckpoint(tx, checkpoint.Id, asset.FilePath, callBack)
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Rollback()
	}

	close(progressChan)
	progress := output.ProgressReport{
		Title:      "Reverting",
		Message:    "Reverting",
		Percentage: 100,
		Current:    1,
		Total:      1,
	}
	app.Event.Emit("progress-update", progress)
	return nil
}

func (c *CheckpointService) AddMissingGroupIds(projectPath string) error {
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

	totalUpdated, totalGroups, err := repository.AddMissingGroupIds(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	output.Message(fmt.Sprintf("Migration completed: Updated %d checkpoints into %d groups (marked as not synced)", totalUpdated, totalGroups))

	tx.Commit()
	return nil
}

// SquashAssets combines multiple untracked files into a single asset with sequential checkpoints.
// The first file becomes the initial checkpoint, and subsequent files are added as additional checkpoints.
func (c *CheckpointService) SquashAssets(projectPath, projectWorkingDir string, filePaths []string, assetName, collectionId string, deleteSourceFiles bool, checkpointComments []string) (models.Asset, error) {
	if len(filePaths) < 2 {
		return models.Asset{}, errors.New("at least two files are required for squash")
	}
	if len(filePaths) > 99 {
		return models.Asset{}, errors.New("cannot squash more than 99 files")
	}
	if assetName == "" {
		return models.Asset{}, errors.New("asset name cannot be empty")
	}
	if len(checkpointComments) != len(filePaths) {
		return models.Asset{}, errors.New("checkpoint comments count must match file paths count")
	}

	app := application.Get()
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Asset{}, err
	}
	defer dbConn.Close()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return models.Asset{}, err
	}

	// Look up required types and status
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Asset{}, err
	}
	userData, err := repository.GetUser(tx, user.Id)
	if err != nil {
		tx.Rollback()
		return models.Asset{}, err
	}
	userRole, err := repository.GetRole(tx, userData.RoleId)
	if err != nil {
		tx.Rollback()
		return models.Asset{}, err
	}
	if !userRole.CreateAsset {
		tx.Rollback()
		return models.Asset{}, error_service.ErrNotUnauthorized
	}
	assetType, err := repository.GetAssetTypeByName(tx, "generic")
	if err != nil {
		tx.Rollback()
		return models.Asset{}, err
	}
	status, err := repository.GetStatusByShortName(tx, "todo")
	if err != nil {
		tx.Rollback()
		return models.Asset{}, err
	}
	statusId := status.Id
	tx.Rollback()

	totalFiles := len(filePaths)
	groupId := uuid.New().String()
	assetId := uuid.New().String()

	// Determine the asset path relative to working dir using the first file
	firstFilePath := filePaths[0]
	if projectWorkingDir == "" {
		return models.Asset{}, errors.New("project working directory cannot be empty")
	}
	firstRelPath, err := filepath.Rel(projectWorkingDir, firstFilePath)
	if err != nil {
		return models.Asset{}, fmt.Errorf("failed to compute relative path: %w", err)
	}

	// Rebuild the asset path with the new asset name
	extension := filepath.Ext(firstFilePath)
	assetDir := filepath.Dir(firstRelPath)
	assetRelPath := filepath.Join(assetDir, assetName+extension)
	assetAbsPath := filepath.Join(projectWorkingDir, assetRelPath)

	// Step 1: Create the asset + first checkpoint from the first file
	app.Event.Emit("progress-update", output.ProgressReport{
		Title:      "Squashing Assets",
		Message:    filepath.Base(filePaths[0]),
		Percentage: 0,
		Current:    1,
		Total:      totalFiles,
	})

	tx, err = dbConn.Beginx()
	if err != nil {
		return models.Asset{}, err
	}
	comment := checkpointComments[0]
	callBack := func(current int, total int, message string, extraMessage string) {
		progress := output.ProgressReport{
			Title:      "Squashing Assets",
			Message:    assetName,
			Percentage: float64(current) / float64(total) * 99,
			Current:    1,
			Total:      totalFiles,
		}
		app.Event.Emit("progress-update", progress)
	}
	err = repository.CreateAssetFast(tx, assetId, assetName, assetType.Id, collectionId, true, "", firstFilePath, "", user.Id, comment, groupId, assetRelPath, statusId, callBack)
	if err != nil {
		tx.Rollback()
		return models.Asset{}, fmt.Errorf("failed to create asset from first file: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return models.Asset{}, err
	}

	// Step 2: Create checkpoints from subsequent files
	for i := 1; i < len(filePaths); i++ {
		filePath := filePaths[i]
		checkpointComment := checkpointComments[i]

		tx, err = dbConn.Beginx()
		if err != nil {
			return models.Asset{}, err
		}

		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Squashing Assets",
				Message:    filepath.Base(filePath),
				Percentage: float64(current) / float64(total) * 99,
				Current:    i + 1,
				Total:      totalFiles,
			}
			app.Event.Emit("progress-update", progress)
		}

		_, err = repository.CreateCheckpoint(tx, assetId, checkpointComment, "", "", 0, 0, filePath, user.Id, "", groupId, callBack)
		if err != nil {
			tx.Rollback()
			if err.Error() == "file not modified" {
				continue
			}
			return models.Asset{}, fmt.Errorf("failed to create checkpoint %s: %w", checkpointComment, err)
		}

		err = tx.Commit()
		if err != nil {
			return models.Asset{}, err
		}
	}

	// Step 3: Fix checkpoint timestamps so they sort correctly (first=oldest, last=newest)
	tx, err = dbConn.Beginx()
	if err != nil {
		return models.Asset{}, err
	}
	checkpoints, err := repository.GetCheckpoints(tx, assetId, false)
	if err != nil {
		tx.Rollback()
		return models.Asset{}, fmt.Errorf("failed to fetch checkpoints for reordering: %w", err)
	}

	// Build a lookup from comment to desired order index
	commentOrder := map[string]int{}
	for i, c := range checkpointComments {
		commentOrder[c] = i
	}

	// Sort checkpoints by their intended order (matching checkpointComments array)
	sort.Slice(checkpoints, func(i, j int) bool {
		return commentOrder[checkpoints[i].Comment] < commentOrder[checkpoints[j].Comment]
	})

	baseEpoch := time.Now().Unix() - int64(len(checkpoints))
	for i, cp := range checkpoints {
		newTime := baseEpoch + int64(i)
		_, err = tx.Exec("UPDATE asset_checkpoint SET created_at = ? WHERE id = ?", newTime, cp.Id)
		if err != nil {
			tx.Rollback()
			return models.Asset{}, fmt.Errorf("failed to update checkpoint timestamp: %w", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return models.Asset{}, err
	}

	// Step 4: Rename/copy the latest file to the asset's working path
	latestFilePath := filePaths[len(filePaths)-1]
	if latestFilePath != assetAbsPath {
		assetDir := filepath.Dir(assetAbsPath)
		err = os.MkdirAll(assetDir, os.ModePerm)
		if err != nil {
			return models.Asset{}, fmt.Errorf("failed to create directory for working file: %w", err)
		}
		data, err := os.ReadFile(latestFilePath)
		if err != nil {
			return models.Asset{}, fmt.Errorf("failed to read latest file: %w", err)
		}
		err = os.WriteFile(assetAbsPath, data, 0644)
		if err != nil {
			return models.Asset{}, fmt.Errorf("failed to write working file: %w", err)
		}
	}

	// Step 5: Delete source files if requested
	if deleteSourceFiles {
		for _, filePath := range filePaths {
			if filePath == assetAbsPath {
				continue
			}
			os.Remove(filePath)
		}
	}

	// Step 6: Retrieve and return the created asset
	tx, err = dbConn.Beginx()
	if err != nil {
		return models.Asset{}, err
	}
	defer tx.Rollback()

	asset, err := repository.GetAsset(tx, assetId)
	if err != nil {
		return models.Asset{}, fmt.Errorf("failed to retrieve created asset: %w", err)
	}

	app.Event.Emit("progress-update", output.ProgressReport{
		Title:      "Squashing Assets",
		Message:    "Complete",
		Percentage: 100,
		Current:    totalFiles,
		Total:      totalFiles,
	})

	return asset, nil
}
