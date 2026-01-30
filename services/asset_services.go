package services

import (
	"bytes"
	"clustta/internal/auth_service"
	"clustta/internal/error_service"
	"clustta/internal/ignore"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/repositorypb"
	"clustta/internal/utils"
	"clustta/output"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"compress/zlib"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/wailsapp/wails/v3/pkg/application"
	"google.golang.org/protobuf/proto"
)

type AssetService struct {
}

type ChangedFiles struct {
	Modifieds  []string `json:"modified"`
	Untrackeds []string `json:"untracked"`
}

type AssetStateItem struct {
	TaskId      string `json:"task_id"`      // task ID for filtering
	TaskPath    string `json:"task_path"`    // for checkpoints: "path/to/file"
	DisplayPath string `json:"display_path"` // for UI: "path/to/file.blend"
}

type AssetsStates struct {
	Modifieds   []AssetStateItem `json:"modified"`
	Rebuildable []AssetStateItem `json:"rebuildable"`
	Outdated    []AssetStateItem `json:"outdated"`
}

type ModifiedAssets struct {
	Modified  []AssetStateItem `json:"modified"`
	Untracked []string         `json:"untracked"`
}

func (t *AssetService) GetAssetCount(projectPath string) (int, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return 0, err
	}
	defer dbConn.Close()

	var count int
	query := "SELECT COUNT(*) FROM full_task"

	err = dbConn.Get(&count, query)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (t *AssetService) GetAssetByID(projectPath, assetId string) (models.Task, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Task{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Task{}, err
	}
	defer tx.Rollback()
	return repository.GetTask(tx, assetId)
}

func (t *AssetService) CreateAsset(projectPath, name, description, taskTypeId, entityId string, isResource bool, templateId, templateFilePath, pointer string, isLink bool, tags []string, previewPath, comment string) (models.Task, error) {
	app := application.Get()
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Task{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Task{}, err
	}
	defer tx.Rollback()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return models.Task{}, err
	}

	previewId := ""
	if previewPath != "" {
		preview, err := repository.CreatePreview(tx, previewPath)
		if err != nil {
			tx.Rollback()
			return models.Task{}, err
		}
		previewId = preview.Hash
	}
	callBack := func(current int, total int, message string, extraMessage string) {
		progress := output.ProgressReport{
			Title:         "Creating Tasks for Entity",
			Message:       name,
			Percentage:    float64(current) / float64(total) * 99,
			Current:       1,
			Total:         2,
			OperationType: "write", // Write operation - creates database records
		}
		app.Event.Emit("progress-update", progress)
	}

	createdTask, err := repository.CreateTask(
		tx,
		"",
		name,
		taskTypeId,
		entityId,
		isResource,
		templateId,
		description,
		templateFilePath,
		tags,
		pointer,
		isLink,
		previewId,
		user.Id,
		comment,
		uuid.New().String(),
		callBack,
	)
	if err != nil {
		tx.Rollback()
		return models.Task{}, err
	}

	createdTask, err = repository.GetTask(tx, createdTask.Id)
	if err != nil {
		return models.Task{}, err
	}

	err = tx.Commit()
	if err != nil {
		return models.Task{}, err
	}

	if templateFilePath == "" {
		return createdTask, nil
	} else {
		progress := output.ProgressReport{
			Title:         "Creating Tasks for Entity",
			Message:       "Task",
			Percentage:    100,
			Current:       1,
			Total:         1,
			OperationType: "write",
		}
		app.Event.Emit("progress-update", progress)
	}
	return createdTask, nil
}

func (t *AssetService) DuplicateAsset(projectPath, sourceTaskId string) (models.Task, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Task{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Task{}, err
	}
	defer tx.Rollback()

	// Get the source task
	sourceTask, err := repository.GetTask(tx, sourceTaskId)
	if err != nil {
		return models.Task{}, err
	}

	// Generate unique name by checking for conflicts
	baseName := sourceTask.Name + "-duplicate"
	newName := baseName
	counter := 1

	// Check for name conflicts in the same entity
	for {
		_, err := repository.GetTaskByName(tx, newName, sourceTask.EntityId, sourceTask.Extension)
		if err != nil {
			// Task with this name doesn't exist, so we can use it
			break
		}
		// Task exists, try with number suffix
		newName = fmt.Sprintf("%s-%d", baseName, counter)
		counter++
	}

	// Generate new ID
	newTaskId := uuid.New().String()

	// Create the duplicate task using AddTask (simpler than CreateTask)
	err = repository.AddTask(
		tx,
		newTaskId,
		utils.GetCurrentTime(),
		newName,
		sourceTask.TaskTypeId,
		sourceTask.EntityId,
		sourceTask.StatusId,
		sourceTask.Extension,
		sourceTask.Description,
		sourceTask.Tags,
		sourceTask.Pointer,
		sourceTask.IsLink,
		"", // No assignee for duplicated task
		sourceTask.PreviewId,
	)
	if err != nil {
		return models.Task{}, err
	}

	// Copy tags from source task
	if len(sourceTask.Tags) > 0 {
		for _, tag := range sourceTask.Tags {
			err = repository.AddTagToTask(tx, newTaskId, tag)
			if err != nil {
				return models.Task{}, err
			}
		}
	}

	// Duplicate the most recent checkpoint if it exists
	latestCheckpoint, err := repository.GetLatestCheckpoint(tx, sourceTaskId)
	if err != nil && err.Error() != "no checkpoints" {
		return models.Task{}, err
	}

	if latestCheckpoint.Id != "" {
		// Generate new checkpoint ID and group ID
		newCheckpointId := uuid.New().String()
		newGroupId := uuid.New().String()

		// Duplicate the latest checkpoint for the new task
		err = repository.AddCheckpoint(
			tx,
			newCheckpointId,
			utils.GetEpochTime(),
			newTaskId, // Link to the new duplicated task
			latestCheckpoint.XXHashChecksum,
			latestCheckpoint.TimeModified,
			latestCheckpoint.FileSize,
			latestCheckpoint.Comment,
			latestCheckpoint.Chunks,
			latestCheckpoint.AuthorUID,
			latestCheckpoint.PreviewId,
			false, // Not synced initially
		)
		if err != nil {
			return models.Task{}, err
		}

		// Update the checkpoint with the new group ID
		_, err = tx.Exec("UPDATE task_checkpoint SET group_id = ? WHERE id = ?", newGroupId, newCheckpointId)
		if err != nil {
			return models.Task{}, err
		}
	}

	// Get the newly created task
	duplicatedTask, err := repository.GetTask(tx, newTaskId)
	if err != nil {
		return models.Task{}, err
	}

	err = tx.Commit()
	if err != nil {
		return models.Task{}, err
	}

	return duplicatedTask, nil
}

// CopyAssetToProject copies an asset from one project to another, including metadata, checkpoints, chunks, and previews.
// If targetEntityId is empty, the asset is copied to the root of the target project.
// If copyAllCheckpoints is false, only the latest checkpoint is copied.
func (t *AssetService) CopyAssetToProject(sourceProjectPath, sourceTaskId, targetProjectPath, targetEntityId string, copyAllCheckpoints bool) (models.Task, error) {
	app := application.Get()

	// Open source database
	sourceDbConn, err := utils.OpenDb(sourceProjectPath)
	if err != nil {
		return models.Task{}, fmt.Errorf("failed to open source project: %w", err)
	}
	defer sourceDbConn.Close()

	sourceTx, err := sourceDbConn.Beginx()
	if err != nil {
		return models.Task{}, err
	}
	defer sourceTx.Rollback()

	// Open target database
	targetDbConn, err := utils.OpenDb(targetProjectPath)
	if err != nil {
		return models.Task{}, fmt.Errorf("failed to open target project: %w", err)
	}
	defer targetDbConn.Close()

	targetTx, err := targetDbConn.Beginx()
	if err != nil {
		return models.Task{}, err
	}
	defer targetTx.Rollback()

	// Get the source task
	sourceTask, err := repository.GetTask(sourceTx, sourceTaskId)
	if err != nil {
		return models.Task{}, fmt.Errorf("failed to get source task: %w", err)
	}

	// Emit progress
	progress := output.ProgressReport{
		Title:         "Copying Asset",
		Message:       sourceTask.Name,
		Percentage:    10,
		OperationType: "write",
	}
	app.Event.Emit("progress-update", progress)

	// Map task type - find equivalent in target project by name, or use "Generic"
	var targetTaskTypeId string
	if sourceTask.TaskTypeName != "" {
		targetTaskType, err := repository.GetTaskTypeByName(targetTx, sourceTask.TaskTypeName)
		if err == nil {
			targetTaskTypeId = targetTaskType.Id
		}
	}
	if targetTaskTypeId == "" {
		// Fall back to Generic
		genericTaskType, err := repository.GetTaskTypeByName(targetTx, "Generic")
		if err == nil {
			targetTaskTypeId = genericTaskType.Id
		} else {
			// Create Generic if it doesn't exist
			genericTaskType, err = repository.GetOrCreateTaskType(targetTx, "Generic", "")
			if err != nil {
				return models.Task{}, fmt.Errorf("failed to get or create Generic task type: %w", err)
			}
			targetTaskTypeId = genericTaskType.Id
		}
	}

	// Map status - find equivalent in target project or use first available
	var targetStatusId string
	if sourceTask.StatusShortName != "" {
		targetStatus, err := repository.GetStatusByShortName(targetTx, sourceTask.StatusShortName)
		if err == nil {
			targetStatusId = targetStatus.Id
		}
	}
	if targetStatusId == "" {
		// Get first available status
		statuses, err := repository.GetStatuses(targetTx)
		if err == nil && len(statuses) > 0 {
			targetStatusId = statuses[0].Id
		}
	}

	// Generate unique name by checking for conflicts in target project
	baseName := sourceTask.Name
	newName := baseName
	counter := 1

	for {
		_, err := repository.GetTaskByName(targetTx, newName, targetEntityId, sourceTask.Extension)
		if err != nil {
			// Task with this name doesn't exist, we can use it
			break
		}
		// Task exists, try with number suffix
		newName = fmt.Sprintf("%s-%d", baseName, counter)
		counter++
	}

	progress.Percentage = 20
	progress.Message = "Creating asset in target project"
	app.Event.Emit("progress-update", progress)

	// Generate new task ID
	newTaskId := uuid.New().String()

	// Copy preview if exists
	var newPreviewId string
	if sourceTask.PreviewId != "" {
		sourcePreview, err := repository.GetPreview(sourceTx, sourceTask.PreviewId)
		if err == nil {
			// Add preview to target project if it doesn't exist
			err = repository.AddPreview(targetTx, sourcePreview.Hash, sourcePreview.Preview, sourcePreview.Extension)
			if err == nil {
				newPreviewId = sourcePreview.Hash
			}
		}
	}

	// Create the task in target project
	err = repository.AddTask(
		targetTx,
		newTaskId,
		utils.GetCurrentTime(),
		newName,
		targetTaskTypeId,
		targetEntityId,
		targetStatusId,
		sourceTask.Extension,
		sourceTask.Description,
		[]string{}, // Don't copy tags (they may not exist in target project)
		"",         // No pointer (file path will be different in target project)
		false,      // Not a link
		"",         // No assignee
		newPreviewId,
	)
	if err != nil {
		return models.Task{}, fmt.Errorf("failed to create task in target project: %w", err)
	}

	progress.Percentage = 30
	progress.Message = "Copying checkpoints"
	app.Event.Emit("progress-update", progress)

	// Get checkpoints to copy
	var checkpointsToCopy []models.Checkpoint
	if copyAllCheckpoints {
		checkpointsToCopy, err = repository.GetCheckpoints(sourceTx, sourceTaskId, false)
		if err != nil && err.Error() != "no checkpoints" {
			return models.Task{}, fmt.Errorf("failed to get checkpoints: %w", err)
		}
	} else {
		latestCheckpoint, err := repository.GetLatestCheckpoint(sourceTx, sourceTaskId)
		if err != nil && err.Error() != "no checkpoints" {
			return models.Task{}, fmt.Errorf("failed to get latest checkpoint: %w", err)
		}
		if latestCheckpoint.Id != "" {
			checkpointsToCopy = []models.Checkpoint{latestCheckpoint}
		}
	}

	// Copy each checkpoint
	totalCheckpoints := len(checkpointsToCopy)
	for i, checkpoint := range checkpointsToCopy {
		progressPercent := 30 + (50 * (i + 1) / max(totalCheckpoints, 1))
		progress.Percentage = float64(progressPercent)
		progress.Message = fmt.Sprintf("Copying checkpoint %d of %d", i+1, totalCheckpoints)
		app.Event.Emit("progress-update", progress)

		// Copy checkpoint preview if exists
		var checkpointPreviewId string
		if checkpoint.PreviewId != "" {
			checkpointPreview, err := repository.GetPreview(sourceTx, checkpoint.PreviewId)
			if err == nil {
				err = repository.AddPreview(targetTx, checkpointPreview.Hash, checkpointPreview.Preview, checkpointPreview.Extension)
				if err == nil {
					checkpointPreviewId = checkpointPreview.Hash
				}
			}
		}

		// Copy chunks for this checkpoint
		if checkpoint.Chunks != "" {
			chunkHashes := strings.Split(checkpoint.Chunks, ",")
			for _, chunkHash := range chunkHashes {
				if chunkHash == "" {
					continue
				}
				// Check if chunk already exists in target
				var existingHash string
				targetTx.Get(&existingHash, "SELECT hash FROM chunk WHERE hash = ?", chunkHash)
				if existingHash != "" {
					continue // Chunk already exists
				}

				// Get chunk from source
				var chunk struct {
					Hash string `db:"hash"`
					Data []byte `db:"data"`
					Size int    `db:"size"`
				}
				err := sourceTx.Get(&chunk, "SELECT hash, data, size FROM chunk WHERE hash = ?", chunkHash)
				if err != nil {
					continue // Skip if chunk not found
				}

				// Insert chunk into target
				_, err = targetTx.Exec("INSERT INTO chunk (hash, data, size) VALUES (?, ?, ?)",
					chunk.Hash, chunk.Data, chunk.Size)
				if err != nil {
					// Ignore duplicate errors
					continue
				}
			}
		}

		// Generate new checkpoint ID and group ID
		newCheckpointId := uuid.New().String()
		newGroupId := uuid.New().String()

		// Add checkpoint to target project
		err = repository.AddCheckpoint(
			targetTx,
			newCheckpointId,
			utils.GetEpochTime(),
			newTaskId,
			checkpoint.XXHashChecksum,
			checkpoint.TimeModified,
			checkpoint.FileSize,
			checkpoint.Comment,
			checkpoint.Chunks,
			checkpoint.AuthorUID,
			checkpointPreviewId,
			false, // Not synced
		)
		if err != nil {
			return models.Task{}, fmt.Errorf("failed to add checkpoint: %w", err)
		}

		// Update the checkpoint with the group ID
		_, err = targetTx.Exec("UPDATE task_checkpoint SET group_id = ? WHERE id = ?", newGroupId, newCheckpointId)
		if err != nil {
			return models.Task{}, fmt.Errorf("failed to update checkpoint group: %w", err)
		}
	}

	progress.Percentage = 90
	progress.Message = "Finalizing"
	app.Event.Emit("progress-update", progress)

	// Get the newly created task
	newTask, err := repository.GetTask(targetTx, newTaskId)
	if err != nil {
		return models.Task{}, fmt.Errorf("failed to get created task: %w", err)
	}

	// Commit target transaction
	err = targetTx.Commit()
	if err != nil {
		return models.Task{}, fmt.Errorf("failed to commit changes: %w", err)
	}

	progress.Percentage = 100
	progress.Message = "Complete"
	app.Event.Emit("progress-update", progress)

	return newTask, nil
}

func (t *AssetService) ChangeStatus(projectPath, taskId, statusId string) error {
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

	err = repository.UpdateStatus(tx, taskId, statusId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// ChangeAssetCollection moves one or more assets to a different collection.
// Checks for name+extension conflicts in the target collection before moving.
// Returns an error if any asset would conflict or if the operation fails.
func (t *AssetService) ChangeAssetCollection(projectPath string, assetIds []string, entityId string) error {
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

	var conflicts []string
	for _, assetId := range assetIds {
		asset, err := repository.GetTask(tx, assetId)
		if err != nil {
			return err
		}
		if asset.EntityId == entityId {
			continue
		}
		_, err = repository.GetTaskByName(tx, asset.Name, entityId, asset.Extension)
		if err == nil {
			conflicts = append(conflicts, asset.Name+asset.Extension)
		} else if err != error_service.ErrTaskNotFound {
			return err
		}
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("assets with the same name and extension already exist in the target collection: %s", strings.Join(conflicts, ", "))
	}

	for _, assetId := range assetIds {
		repository.ChangeEntity(tx, assetId, entityId)
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// MoveAssetsToCollection moves one or more assets to a different collection.
// Updates the database and moves the physical files if they exist on disk.
// Checks for name+extension conflicts in the target collection before moving.
// Returns an error if any asset would conflict or if the operation fails.
func (t *AssetService) MoveAssetsToCollection(projectPath string, assetIds []string, targetEntityId string) error {
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

	// Get target directory path
	var targetDir string
	if targetEntityId == "" {
		// Moving to root - use project's working directory
		targetDir, err = utils.GetProjectWorkingDir(tx)
		if err != nil {
			return err
		}
	} else {
		// Moving to a collection - get its file path
		targetEntity, err := repository.GetEntity(tx, targetEntityId)
		if err != nil {
			return err
		}
		targetDir = targetEntity.FilePath
	}

	// Check for name+extension conflicts in target collection
	var conflicts []string
	var assetsToMove []models.Task
	for _, assetId := range assetIds {
		asset, err := repository.GetTask(tx, assetId)
		if err != nil {
			return err
		}
		// Skip if already in target collection
		if asset.EntityId == targetEntityId {
			continue
		}
		// Check if a task with same name+extension exists in target collection
		_, err = repository.GetTaskByName(tx, asset.Name, targetEntityId, asset.Extension)
		if err == nil {
			// Task exists - this is a conflict
			conflicts = append(conflicts, asset.Name+asset.Extension)
		} else if err != error_service.ErrTaskNotFound {
			// Some other error occurred
			return err
		}
		assetsToMove = append(assetsToMove, asset)
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("assets with the same name and extension already exist in the target collection: %s", strings.Join(conflicts, ", "))
	}

	// No conflicts - proceed with move
	// Ensure target directory exists
	os.MkdirAll(targetDir, os.ModePerm)

	// Move files and update database
	for _, asset := range assetsToMove {
		// Calculate new file path
		fileName := asset.Name + asset.Extension
		newFilePath := filepath.Join(targetDir, fileName)

		// Move file if it exists on disk
		if asset.FilePath != "" {
			if _, err := os.Stat(asset.FilePath); err == nil {
				err = os.Rename(asset.FilePath, newFilePath)
				if err != nil {
					return fmt.Errorf("failed to move file %s: %w", asset.Name, err)
				}
			}
		}

		// Update database
		repository.ChangeEntity(tx, asset.Id, targetEntityId)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (t *AssetService) DeleteAsset(projectPath, taskId string, removeFiles bool) error {
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

	err = repository.DeleteTask(tx, taskId, removeFiles, true)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (t *AssetService) UpdateAsset(projectPath, taskId, name, taskTypeId string, isResource bool, pointer string, tags []string) (models.Task, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Task{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Task{}, err
	}
	defer tx.Rollback()

	updatedTask, err := repository.UpdateTask(tx, taskId, name, taskTypeId, isResource, pointer, tags)
	if err != nil {
		tx.Rollback()
		return models.Task{}, err
	}
	err = tx.Commit()
	if err != nil {
		return models.Task{}, err
	}
	return updatedTask, nil
}

func (t *AssetService) ChangeAssetType(projectPath, taskId, taskTypeId string) error {
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

	err = repository.ChangeTaskType(tx, taskId, taskTypeId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (t *AssetService) ToggleIsTask(projectPath, taskId string, isTask bool) error {
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

	err = repository.ToggleIsTask(tx, taskId, isTask)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (t *AssetService) RenameAsset(projectPath, taskId, name string) (models.Task, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Task{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Task{}, err
	}
	defer tx.Rollback()

	updatedTask, err := repository.RenameTask(tx, taskId, name)
	if err != nil {
		tx.Rollback()
		return models.Task{}, err
	}
	err = tx.Commit()
	if err != nil {
		return models.Task{}, err
	}
	return updatedTask, nil
}

func (t *AssetService) AddPreview(projectPath, taskId, previewPath string) (models.Task, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Task{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Task{}, err
	}
	defer tx.Rollback()

	preview, err := repository.CreatePreview(tx, previewPath)
	if err != nil {
		tx.Rollback()
		return models.Task{}, err
	}
	err = repository.SetEntityPreview(tx, taskId, "task", preview.Hash)
	if err != nil {
		tx.Rollback()
		return models.Task{}, err
	}
	updatedTask, err := repository.GetTask(tx, taskId)
	if err != nil {
		return models.Task{}, err
	}

	err = tx.Commit()
	if err != nil {
		return models.Task{}, err
	}

	return updatedTask, nil
}

// AssignAsset assigns a task to a user.
// If the task is a resource (is_resource == true), it will be converted to a task first.
func (t *AssetService) AssignAsset(projectPath, taskId, userId string) error {
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

	task, err := repository.GetTask(tx, taskId)
	if err != nil {
		return err
	}

	if task.IsResource {
		err = repository.ToggleIsTask(tx, taskId, true)
		if err != nil {
			return err
		}
	}

	err = repository.AssignTask(tx, taskId, userId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
func (t *AssetService) UnassignAsset(projectPath, taskId string) error {
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

	err = repository.UnAssignTask(tx, taskId)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
func (t *AssetService) UnassignAssets(projectPath string, taskIds []string) error {
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

	err = repository.UnAssignTasks(tx, taskIds)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
func (t *AssetService) AssetFileStatus(projectPath, taskId string) (string, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return "", err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	task, err := repository.GetTask(tx, taskId)
	if err != nil {
		return "", err
	}
	return task.FileStatus, nil
}
func (t *AssetService) AssetFilesStatus(projectPath string, taskIds []string) (map[string]string, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return map[string]string{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return map[string]string{}, err
	}
	defer tx.Rollback()

	filesStatus, err := repository.GetFilesStatus(tx, taskIds)
	if err != nil {
		return map[string]string{}, err
	}
	return filesStatus, nil
}

func (t *AssetService) GetAssetState(projectPath string, taskId string) (string, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return "", err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	assetState, err := repository.GetAssetState(tx, taskId)
	if err != nil {
		return "", err
	}
	return assetState, nil
}

func (t *AssetService) ToggleIsResource(projectPath string, taskIds []string, isResource bool) error {
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

	err = repository.ToggleIsResourceM(tx, taskIds, isResource)
	if err != nil {
		return err
	}
	return nil
}

func (t *AssetService) RevealAsset(projectPath, taskId string) error {
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

	task, err := repository.GetTask(tx, taskId)
	if err != nil {
		return err
	}
	utils.RevealInExplorer(task.GetFilePath())
	return nil
}

// dependencies
func (t *AssetService) AddEntityDependency(projectPath, taskId, dependencyId, dependencyTypeId string) (models.TaskDependency, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.TaskDependency{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.TaskDependency{}, err
	}
	defer tx.Rollback()
	entityDependency, err := repository.AddEntityDependency(tx, "", taskId, dependencyId, dependencyTypeId)
	if err != nil {
		return models.TaskDependency{}, err
	}
	err = tx.Commit()
	if err != nil {
		return models.TaskDependency{}, err
	}
	return entityDependency, nil
}
func (t *AssetService) RemoveEntityDependency(projectPath, taskId, dependencyId string) error {
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
	err = repository.RemoveEntityDependency(tx, taskId, dependencyId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
func (t *AssetService) AddAssetDependency(projectPath, taskId, dependencyId, dependencyTypeId string) (models.TaskDependency, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.TaskDependency{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.TaskDependency{}, err
	}
	defer tx.Rollback()

	taskDependency, err := repository.AddDependency(tx, "", taskId, dependencyId, dependencyTypeId)
	if err != nil {
		return models.TaskDependency{}, err
	}
	err = tx.Commit()
	if err != nil {
		return models.TaskDependency{}, err
	}

	return taskDependency, nil
}
func (t *AssetService) RemoveAssetDependency(projectPath, taskId, dependencyId string) error {
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
	err = repository.RemoveTaskDependency(tx, taskId, dependencyId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
func (t *AssetService) GetAssetDependencies2(projectPath string, taskIds []string) ([]models.Task, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Task{}, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Task{}, err
	}
	defer tx.Rollback()

	if len(taskIds) == 0 {
		return []models.Task{}, nil
	}

	tasks := []models.Task{}
	quotedTaskIds := make([]string, len(taskIds))
	for i, id := range taskIds {
		quotedTaskIds[i] = fmt.Sprintf("'%s'", id)
	}

	tasksQuery := fmt.Sprintf(` SELECT * FROM full_task  WHERE id IN (%s) AND trashed = 0 ORDER BY name `, strings.Join(quotedTaskIds, ","))

	err = tx.Select(&tasks, tasksQuery)
	if err != nil && err != sql.ErrNoRows {
		return []models.Task{}, err
	}

	err = tx.Commit()
	if err != nil {
		return tasks, err
	}

	return tasks, nil
}
func (t *AssetService) GetAssetDependencies(projectPath string, taskIds []string) ([]interface{}, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []interface{}{}, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return []interface{}{}, err
	}
	defer tx.Rollback()

	// If no task IDs provided, return empty result
	if len(taskIds) == 0 {
		return []interface{}{}, nil
	}

	result := []interface{}{}

	// Get the task records
	tasks := []models.Task{}
	quotedTaskIds := make([]string, len(taskIds))
	for i, id := range taskIds {
		quotedTaskIds[i] = fmt.Sprintf("'%s'", id)
	}

	tasksQuery := fmt.Sprintf(`
		SELECT * FROM full_task 
		WHERE id IN (%s) AND trashed = 0 
		ORDER BY name
	`, strings.Join(quotedTaskIds, ","))

	err = tx.Select(&tasks, tasksQuery)
	if err != nil && err != sql.ErrNoRows {
		return []interface{}{}, err
	}

	// Add tasks to result
	for _, task := range tasks {
		result = append(result, task)
	}

	// Find IDs that didn't match any tasks
	foundTaskIds := make(map[string]bool)
	for _, task := range tasks {
		foundTaskIds[task.Id] = true
	}

	missingIds := []string{}
	for _, id := range taskIds {
		if !foundTaskIds[id] {
			missingIds = append(missingIds, id)
		}
	}

	// Get entities for missing IDs
	if len(missingIds) > 0 {
		entities := []models.Entity{}
		quotedMissingIds := make([]string, len(missingIds))
		for i, id := range missingIds {
			quotedMissingIds[i] = fmt.Sprintf("'%s'", id)
		}

		entitiesQuery := fmt.Sprintf(`
			SELECT * FROM full_entity 
			WHERE id IN (%s) AND trashed = 0 
			ORDER BY name
		`, strings.Join(quotedMissingIds, ","))

		err = tx.Select(&entities, entitiesQuery)
		if err != nil && err != sql.ErrNoRows {
			return []interface{}{}, err
		}

		for _, entity := range entities {
			result = append(result, entity)
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return result, err
	}

	return result, nil
}

func (t *AssetService) GetRecursiveDependencies(projectPath string, taskId string, maxDepth int) ([]interface{}, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []interface{}{}, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return []interface{}{}, err
	}
	defer tx.Rollback()

	// Get all task dependencies records
	allTaskDependencies := []models.TaskDependency{}
	query := `SELECT task_id, dependency_id FROM task_dependency`
	err = tx.Select(&allTaskDependencies, query)
	if err != nil {
		return []interface{}{}, err
	}

	// Get all entity dependencies records
	allEntityDependencies := []models.EntityDependency{}
	query = `SELECT task_id, dependency_id FROM entity_dependency`
	err = tx.Select(&allEntityDependencies, query)
	if err != nil {
		return []interface{}{}, err
	}

	// Get all task info for checking task existence
	allTaskInfo := []models.Task{}
	query = `SELECT id FROM task WHERE trashed = 0`
	err = tx.Select(&allTaskInfo, query)
	if err != nil {
		return []interface{}{}, err
	}

	// Get all entity info for checking entity existence
	allEntityInfo := []models.Entity{}
	query = `SELECT id FROM entity WHERE trashed = 0`
	err = tx.Select(&allEntityInfo, query)
	if err != nil {
		return []interface{}{}, err
	}

	// Track dependencies with their depth and parent information
	type DependencyInfo struct {
		ID       string
		Depth    int
		ParentID string
	}

	result := []interface{}{}
	dependenciesMap := make(map[string]DependencyInfo) // track dependency info

	// Helper function to collect dependencies recursively
	var collectDependencies func(string, int, string)
	collectDependencies = func(currentTaskId string, currentDepth int, parentTaskId string) {
		if currentDepth >= maxDepth {
			return
		}

		// If we encounter the original taskId, skip collecting it or its dependencies
		if currentTaskId == taskId && currentDepth > 0 {
			return
		}

		// Get direct task dependencies
		for _, taskDep := range allTaskDependencies {
			if taskDep.TaskId == currentTaskId {
				depId := taskDep.DependencyId
				if depId == taskId {
					// Skip collecting the original taskId as a dependency
					continue
				}
				if existing, exists := dependenciesMap[depId]; !exists || currentDepth+1 < existing.Depth {
					dependenciesMap[depId] = DependencyInfo{
						ID:       depId,
						Depth:    currentDepth + 1,
						ParentID: currentTaskId, // The current task is the parent of this dependency
					}
					collectDependencies(depId, currentDepth+1, currentTaskId)
				}
			}
		}

		// Get entity dependencies (entities only, no child traversal)
		for _, entityDep := range allEntityDependencies {
			if entityDep.TaskId == currentTaskId {
				entityId := entityDep.DependencyId
				if entityId == taskId {
					// Skip collecting the original taskId as a dependency
					continue
				}
				if existing, exists := dependenciesMap[entityId]; !exists || currentDepth+1 < existing.Depth {
					dependenciesMap[entityId] = DependencyInfo{
						ID:       entityId,
						Depth:    currentDepth + 1,
						ParentID: currentTaskId, // The current task is the parent of this entity dependency
					}
				}
			}
		}
	}

	// Start recursive collection from the given task
	collectDependencies(taskId, 0, "")

	// Get all dependency IDs
	dependencyIds := make([]string, 0, len(dependenciesMap))
	for depId := range dependenciesMap {
		dependencyIds = append(dependencyIds, depId)
	}

	if len(dependencyIds) == 0 {
		return result, nil
	}

	// Fetch task objects
	tasks := []models.Task{}
	quotedTaskIds := make([]string, 0)

	for _, depId := range dependencyIds {
		// Check if this ID corresponds to a task
		for _, task := range allTaskInfo {
			if task.Id == depId {
				quotedTaskIds = append(quotedTaskIds, fmt.Sprintf("'%s'", depId))
				break
			}
		}
	}

	if len(quotedTaskIds) > 0 {
		tasksQuery := fmt.Sprintf(`
			SELECT * FROM full_task 
			WHERE id IN (%s) AND trashed = 0 
			ORDER BY name
		`, strings.Join(quotedTaskIds, ","))

		err = tx.Select(&tasks, tasksQuery)
		if err != nil && err != sql.ErrNoRows {
			return []interface{}{}, err
		}

		// Add depth and parent information to tasks
		for _, task := range tasks {
			depInfo := dependenciesMap[task.Id]
			taskWithDepth := map[string]interface{}{
				"task":     task,
				"name":     task.Name,
				"depth":    depInfo.Depth,
				"parentId": depInfo.ParentID,
				"type":     "task",
			}
			result = append(result, taskWithDepth)
		}
	}

	// Fetch entity objects
	entities := []models.Entity{}
	quotedEntityIds := make([]string, 0)

	for _, depId := range dependencyIds {
		// Check if this ID corresponds to an entity
		for _, entity := range allEntityInfo {
			if entity.Id == depId {
				quotedEntityIds = append(quotedEntityIds, fmt.Sprintf("'%s'", depId))
				break
			}
		}
	}

	if len(quotedEntityIds) > 0 {
		entitiesQuery := fmt.Sprintf(`
			SELECT * FROM full_entity 
			WHERE id IN (%s) AND trashed = 0 
			ORDER BY name
		`, strings.Join(quotedEntityIds, ","))

		err = tx.Select(&entities, entitiesQuery)
		if err != nil && err != sql.ErrNoRows {
			return []interface{}{}, err
		}

		// Add depth and parent information to entities
		for _, entity := range entities {
			depInfo := dependenciesMap[entity.Id]
			entityWithDepth := map[string]interface{}{
				"entity":   entity,
				"depth":    depInfo.Depth,
				"parentId": depInfo.ParentID,
				"type":     "entity",
			}
			result = append(result, entityWithDepth)
		}
	}

	err = tx.Commit()
	if err != nil {
		return result, err
	}

	return result, nil
}

func (t *AssetService) GetAssets(projectPath string) ([]models.Task, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return []models.Task{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Task{}, err
	}
	defer tx.Rollback()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return []models.Task{}, err
	}
	userData, err := repository.GetUser(tx, user.Id)
	if err != nil {
		return []models.Task{}, err
	}
	userRole, err := repository.GetRole(tx, userData.RoleId)
	if err != nil {
		return []models.Task{}, err
	}
	if userRole.ViewTask {
		start := time.Now()
		tasks, err := repository.GetTasks(tx, true)
		if err != nil {
			return []models.Task{}, err
		}
		elapsed := time.Since(start)
		fmt.Printf("GetTasks operation took %s\n", elapsed)
		return tasks, nil
	} else {
		tasks, err := repository.GetUserTasks(tx, user.Id)
		if err != nil {
			return []models.Task{}, err
		}
		return tasks, nil
	}
}

// GetAssetTasks gets all tasks where is_resource is false with minimal fields for UI display
func (t *AssetService) GetAssetTasks(projectPath string) ([]models.Task, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return []models.Task{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Task{}, err
	}
	defer tx.Rollback()

	tasks, err := repository.GetAssetTasks(tx)
	if err != nil {
		return []models.Task{}, err
	}
	return tasks, nil
}

func (t *AssetService) TestData() string {
	return "test"
}

func (t *AssetService) GetAssetsPB(projectPath string) ([]byte, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return []byte{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []byte{}, err
	}
	defer tx.Rollback()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return []byte{}, err
	}
	userData, err := repository.GetUser(tx, user.Id)
	if err != nil {
		return []byte{}, err
	}
	userRole, err := repository.GetRole(tx, userData.RoleId)
	if err != nil {
		return []byte{}, err
	}
	if userRole.ViewTask {
		start := time.Now()
		tasks, err := repository.GetTasks(tx, true)
		if err != nil {
			return []byte{}, err
		}

		pbTasks := repository.ToPbFullTasks(tasks)
		pbTasksList := &repositorypb.FullTaskList{FullTasks: pbTasks}
		pbTasksBytes, err := proto.Marshal(pbTasksList)
		if err != nil {
			return []byte{}, err
		}

		elapsed := time.Since(start)
		fmt.Printf("GetTasks operation took %s\n", elapsed)

		//zlib compression
		compressedData := bytes.NewBuffer(nil)
		writer := zlib.NewWriter(compressedData)
		_, err = writer.Write(pbTasksBytes)
		if err != nil {
			return []byte{}, err
		}
		err = writer.Close()
		if err != nil {
			return []byte{}, err
		}
		compressedBytes := compressedData.Bytes()

		return compressedBytes, nil
	} else {
		tasks, err := repository.GetUserTasks(tx, user.Id)
		if err != nil {
			return []byte{}, err
		}

		pbTasks := repository.ToPbFullTasks(tasks)
		pbTasksList := &repositorypb.FullTaskList{FullTasks: pbTasks}
		pbTasksBytes, err := proto.Marshal(pbTasksList)
		if err != nil {
			return []byte{}, err
		}

		compressedData := bytes.NewBuffer(nil)
		writer := zlib.NewWriter(compressedData)
		_, err = writer.Write(pbTasksBytes)
		if err != nil {
			return []byte{}, err
		}
		err = writer.Close()
		if err != nil {
			return []byte{}, err
		}
		compressedBytes := compressedData.Bytes()
		return compressedBytes, nil
	}
}

// asset types
func (t *AssetService) GetAssetTypes(projectPath string) ([]models.TaskType, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return []models.TaskType{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.TaskType{}, err
	}
	defer tx.Rollback()

	taskTypes, err := repository.GetTaskTypes(tx)
	if err != nil {
		return []models.TaskType{}, err
	}
	return taskTypes, nil
}

func (t *AssetService) DeleteAssetType(projectPath, id string) error {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = repository.DeleteTaskType(tx, id)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (t *AssetService) CreateAssetType(projectPath, name, icon string) (models.TaskType, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return models.TaskType{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.TaskType{}, err
	}
	defer tx.Rollback()

	taskTypes, err := repository.CreateTaskType(tx, "", name, icon)
	if err != nil {
		return models.TaskType{}, err
	}

	err = tx.Commit()
	if err != nil {
		return models.TaskType{}, err
	}
	return taskTypes, nil
}

func (t *AssetService) UpdateAssetType(projectPath, id, name, icon string) (models.TaskType, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return models.TaskType{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.TaskType{}, err
	}
	defer tx.Rollback()

	taskType, err := repository.UpdateTaskType(tx, id, name, icon)
	if err != nil {
		return models.TaskType{}, err
	}

	err = tx.Commit()
	if err != nil {
		return models.TaskType{}, err
	}
	return taskType, nil
}

func (t *AssetService) GetAssetsStates(projectPath, projectWorkingDir string, ignoreList []string) (AssetsStates, error) {
	assetsStates := AssetsStates{
		Modifieds:   []AssetStateItem{},
		Rebuildable: []AssetStateItem{},
		Outdated:    []AssetStateItem{},
	}

	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return assetsStates, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return assetsStates, err
	}
	defer tx.Rollback()

	if !utils.DirExists(projectWorkingDir) {
		tasks := []models.Task{}
		query := "SELECT task_path, extension FROM full_task WHERE trashed = 0 ORDER BY task_path"
		err = tx.Select(&tasks, query)
		if err != nil {
			return assetsStates, err
		}
		for _, task := range tasks {
			displayPath := task.TaskPath
			if task.Extension != "" {
				displayPath = task.TaskPath + task.Extension
			}
			assetsStates.Rebuildable = append(assetsStates.Rebuildable, AssetStateItem{
				TaskId:      task.Id,
				TaskPath:    task.TaskPath,
				DisplayPath: displayPath,
			})
		}
		return assetsStates, nil // No untracked items if the entity folder does not exist
	}

	tasks := []models.Task{}
	query := "SELECT * FROM full_task WHERE trashed = 0 ORDER BY task_path"

	err = tx.Select(&tasks, query)
	if err != nil {
		return assetsStates, err
	}

	checkpointQuery := "SELECT * FROM task_checkpoint WHERE trashed = 0 ORDER BY created_at DESC"
	tasksCheckpoints := []models.Checkpoint{}
	tx.Select(&tasksCheckpoints, checkpointQuery)

	taskCheckpoints := map[string][]models.Checkpoint{}
	for _, taskCheckpoint := range tasksCheckpoints {
		taskCheckpoints[taskCheckpoint.TaskId] = append(taskCheckpoints[taskCheckpoint.TaskId], taskCheckpoint)
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return assetsStates, err
	}

	for i, task := range tasks {
		taskFilePath, err := utils.BuildTaskPath(rootFolder, task.EntityPath, task.Name, task.Extension)
		if err != nil {
			return assetsStates, err
		}
		tasks[i].FilePath = taskFilePath
		tasks[i].Checkpoints = taskCheckpoints[task.Id]

		fileStatus, err := repository.GetTaskFileStatus(&tasks[i], taskCheckpoints[task.Id])
		if err != nil {
			return assetsStates, err
		}
		if fileStatus == "modified" {
			displayPath := task.TaskPath
			if task.Extension != "" {
				displayPath = task.TaskPath + task.Extension
			}
			assetsStates.Modifieds = append(assetsStates.Modifieds, AssetStateItem{
				TaskId:      task.Id,
				TaskPath:    task.TaskPath,
				DisplayPath: displayPath,
			})
		} else if fileStatus == "outdated" {
			displayPath := task.TaskPath
			if task.Extension != "" {
				displayPath = task.TaskPath + task.Extension
			}
			assetsStates.Outdated = append(assetsStates.Outdated, AssetStateItem{
				TaskId:      task.Id,
				TaskPath:    task.TaskPath,
				DisplayPath: displayPath,
			})
		} else if fileStatus == "rebuildable" {
			displayPath := task.TaskPath
			if task.Extension != "" {
				displayPath = task.TaskPath + task.Extension
			}
			assetsStates.Rebuildable = append(assetsStates.Rebuildable, AssetStateItem{
				TaskId:      task.Id,
				TaskPath:    task.TaskPath,
				DisplayPath: displayPath,
			})
		}
	}

	return assetsStates, nil
}

func (t *AssetService) GetUntrackedFiles(projectPath, projectWorkingDir string, ignoreList []string) ([]string, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return []string{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []string{}, err
	}
	defer tx.Rollback()

	untrackedFiles := []string{}
	if !utils.DirExists(projectWorkingDir) {
		return untrackedFiles, nil
	}

	clusttaIgnore := ignore.CompileIgnoreLines(ignoreList...)

	// Pre-process tracked items into maps for O(1) lookup
	absoluteTrackedFiles := make(map[string]bool)

	tasks := []models.Task{}
	query := "SELECT task_path, extension FROM full_task WHERE trashed = 0 ORDER BY task_path"

	err = tx.Select(&tasks, query)
	if err != nil {
		return untrackedFiles, err
	}

	for _, task := range tasks {
		absoluteTaskFilePath, err := filepath.Abs(filepath.Join(projectWorkingDir, task.TaskPath+task.Extension))
		if err != nil {
			return untrackedFiles, err
		}
		// absoluteTaskFilePath = utils.NormalizePath(absoluteTaskFilePath)
		absoluteTrackedFiles[absoluteTaskFilePath] = true
	}

	err = filepath.WalkDir(projectWorkingDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}
		} else {
			if strings.HasPrefix(filepath.Base(path), ".") {
				return nil
			}
			relativePath, err := filepath.Rel(projectWorkingDir, path)
			if err != nil {
				return err
			}
			relativePath = utils.NormalizePath(relativePath)
			if !absoluteTrackedFiles[path] && !clusttaIgnore.MatchesPath(relativePath) {
				untrackedFiles = append(untrackedFiles, "/"+relativePath)
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return untrackedFiles, nil
}
