package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/error_service"
	"clustta/internal/ignore"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/sync_service"
	"clustta/internal/utils"
	"clustta/output"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/jmoiron/sqlx"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type EntityItems struct {
	Tasks            []models.Task            `json:"tasks"`
	Entities         []models.Entity          `json:"entities"`
	UntrackedFiles   []models.UntrackedTask   `json:"untracked_tasks"`
	UntrackedFolders []models.UntrackedEntity `json:"untracked_entities"`
}

type CollectionService struct {
}

func (t *CollectionService) GetEntityCount(projectPath string) (int, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return 0, err
	}
	defer dbConn.Close()

	var count int
	query := "SELECT COUNT(*) FROM full_entity"

	err = dbConn.Get(&count, query)
	if err != nil {
		return 0, err
	}

	return count, nil
}
func (e *CollectionService) CreateEntity(projectPath, name, description, entityTypeId, parentId, previewPath string, isLibrary bool) (models.Entity, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Entity{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Entity{}, err
	}
	defer tx.Rollback()

	previewId := ""
	if previewPath != "" {
		preview, err := repository.CreatePreview(tx, previewPath)
		if err != nil {
			tx.Rollback()
			return models.Entity{}, err
		}
		previewId = preview.Hash
	}

	createdEntity, err := repository.CreateEntity(
		tx,
		"",
		name,
		description,
		entityTypeId,
		parentId,
		previewId,
		isLibrary,
	)
	if err != nil {
		tx.Rollback()
		return models.Entity{}, err
	}
	err = tx.Commit()
	if err != nil {
		return models.Entity{}, err
	}
	return createdEntity, nil
}

func (e *CollectionService) RenameEntity(projectPath, entityId, newName string) (models.Entity, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Entity{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Entity{}, err
	}
	defer tx.Rollback()

	updatedEntity, err := repository.RenameEntity(tx, entityId, newName)
	if err != nil {
		tx.Rollback()
		return models.Entity{}, err
	}

	err = tx.Commit()
	if err != nil {
		return models.Entity{}, err
	}
	return updatedEntity, nil
}

func (e *CollectionService) CreateEntities(projectPath, name, description, entityTypeId, parentId string) ([]models.Entity, error) {
	//TODO implement CreateEntities
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Entity{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Entity{}, err
	}
	defer tx.Rollback()

	return []models.Entity{}, nil
}

func (e *CollectionService) DeleteEntity(projectPath, entityId string, removeFiles bool) error {

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

	err = repository.DeleteEntity(tx, entityId, removeFiles, true)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (e *CollectionService) GetEntities(projectPath string) ([]models.Entity, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.Entity{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Entity{}, err
	}
	defer tx.Rollback()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return []models.Entity{}, err
	}
	userData, err := repository.GetUser(tx, user.Id)
	if err != nil {
		return []models.Entity{}, err
	}
	userRole, err := repository.GetRole(tx, userData.RoleId)
	if err != nil {
		return []models.Entity{}, err
	}

	if userRole.ViewTask {
		entities, err := repository.GetEntities(tx, true)
		if err != nil {
			return []models.Entity{}, err
		}
		return entities, err
	} else {
		userTaskInfo, err := repository.GetUserTasksMinimal(tx, user.Id)
		if err != nil {
			return []models.Entity{}, err
		}

		entities, err := repository.GetUserEntities(tx, userTaskInfo, user.Id)
		if err != nil {
			return []models.Entity{}, err
		}
		return entities, err
	}
}

func (e *CollectionService) GetEntityChildren(projectPath, entityId, projectWorkingDir, entityFolderPath string, ignoreList []string, isUntracked bool) (EntityItems, error) {
	children := EntityItems{
		Tasks:            make([]models.Task, 0),
		Entities:         make([]models.Entity, 0),
		UntrackedFiles:   make([]models.UntrackedTask, 0),
		UntrackedFolders: make([]models.UntrackedEntity, 0),
	}
	if entityId == "root" {
		entityId = ""
	}

	entityTrackFolders := []string{}
	entityTrackFiles := []string{}
	if !isUntracked {
		dbConn, err := utils.OpenDb(projectPath)
		if err != nil {
			return children, err
		}
		defer dbConn.Close()
		tx, err := dbConn.Beginx()
		if err != nil {
			return children, err
		}
		defer tx.Rollback()

		if entityId == "root" {
			entityId = ""
		}

		entities, err := repository.GetEntityChildren(tx, entityId)
		if err != nil {
			return children, err
		}
		children.Entities = entities

		tasks, err := repository.GetEntityTasks(tx, entityId)
		if err != nil {
			return children, err
		}
		children.Tasks = tasks

		for _, child := range entities {
			entityTrackFolders = append(entityTrackFolders, child.Name)
		}

		for _, child := range tasks {
			entityTrackFiles = append(entityTrackFiles, child.Name+child.Extension)
		}
	}

	if !utils.DirExists(entityFolderPath) {
		return children, nil // No untracked items if the entity folder does not exist
	}

	// Get the absolute path of the entity folder
	absoluteEntityFolderPath, err := filepath.Abs(entityFolderPath)
	if err != nil {
		return children, err
	}
	// Get the relative path of the entity folder
	relativeEntityFolderPath, err := filepath.Rel(projectWorkingDir, absoluteEntityFolderPath)
	if err != nil {
		return children, err
	}
	relativeEntityFolderPath = utils.NormalizePath(relativeEntityFolderPath)
	// Compile the ignore patterns
	clusttaIgnore := ignore.CompileIgnoreLines(ignoreList...)
	// Read the directory entries
	entries, err := os.ReadDir(absoluteEntityFolderPath)
	if err != nil {
		return children, err
	}

	// Iterate over the entries in the entity folder
	for _, entry := range entries {
		entryPath := filepath.Join(absoluteEntityFolderPath, entry.Name())
		relativePath := utils.NormalizePath(filepath.Join(relativeEntityFolderPath, entry.Name()))
		parentId := entityId
		if parentId == "root" {
			parentId = ""
		}
		if entry.IsDir() {
			//check if it is tracked
			if slices.Contains(entityTrackFolders, entry.Name()) {
				continue
			}
			if !clusttaIgnore.MatchesPath(relativePath) {
				children.UntrackedFolders = append(children.UntrackedFolders, models.UntrackedEntity{
					Id:         utils.GetMD5Hash(entryPath),
					Name:       entry.Name(),
					FilePath:   entryPath,
					EntityPath: "/" + relativePath + "/",
					ItemPath:   "/" + relativePath + "/",
					ParentId:   parentId,
				})
			}
		} else {
			//check if it is tracked
			if slices.Contains(entityTrackFiles, entry.Name()) {
				continue
			}

			// If the entry is a file, check if it is tracked
			if !clusttaIgnore.MatchesPath(relativePath) {
				taskName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
				children.UntrackedFiles = append(children.UntrackedFiles, models.UntrackedTask{
					Id:           utils.GetMD5Hash(entryPath),
					Name:         taskName,
					FilePath:     entryPath,
					TaskPath:     "/" + relativePath,
					EntityId:     parentId,
					EntityPath:   "/" + relativeEntityFolderPath + "/",
					Extension:    filepath.Ext(entry.Name()),
					ItemPath:     "/" + relativePath + "/",
					TaskTypeIcon: "generic",
				})
			}
		}
	}

	return children, nil
}

func (e *CollectionService) GetEntityTasks(projectPath, entityId string) ([]models.Task, error) {
	if entityId == "root" {
		entityId = ""
	}
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
	return repository.GetEntityTasks(tx, entityId)
}

func (e *CollectionService) GetEntityByID(projectPath, entityId string) (models.Entity, error) {
	if entityId == "root" {
		entityId = ""
	}
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Entity{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Entity{}, err
	}
	defer tx.Rollback()
	return repository.GetEntity(tx, entityId)
}

func (e *CollectionService) Rebuild(projectPath, remoteUrl, entityIds, userId string) error {
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
		Title:      "Rebuilding",
		Message:    "Preparing to Rebuild",
		Percentage: 0,
		Current:    1,
		Total:      2,
	}:
	}

	// Parse entityIds - could be single ID, comma-separated IDs, or empty string for all
	var entityIdList []string
	if entityIds == "" {
		entityIdList = []string{""}
	} else {
		entityIdList = strings.Split(entityIds, ",")
		// Trim whitespace from each ID
		for i, id := range entityIdList {
			entityIdList[i] = strings.TrimSpace(id)
		}
	}

	entityEntitiesQuery := `
	SELECT full_entity.*
	FROM full_entity
	WHERE full_entity.entity_path LIKE ? OR full_entity.entity_path LIKE ?;
	`

	entities := []models.Entity{}
	allTasks := []models.Task{}

	// Process each entity ID
	for _, entityId := range entityIdList {
		if entityId == "" {
			// Get all entities and tasks for root rebuild
			rootEntities, err := repository.GetEntities(tx, false)
			if err != nil {
				return err
			}
			entities = append(entities, rootEntities...)

			rootTasks, err := repository.GetTasks(tx, false)
			if err != nil {
				return err
			}
			allTasks = append(allTasks, rootTasks...)
		} else {
			// Get specific entity and its children
			parentEntity, err := repository.GetEntity(tx, entityId)
			if err != nil {
				return err
			}
			err = os.MkdirAll(parentEntity.FilePath, os.ModePerm)
			if err != nil {
				return err
			}
			pathLike := parentEntity.EntityPath + "%"
			var entityChildren []models.Entity
			err = tx.Select(&entityChildren, entityEntitiesQuery, parentEntity.EntityPath, pathLike)
			if err != nil {
				return err
			}
			entities = append(entities, entityChildren...)

			// Get tasks for this entity
			entityTasksQuery := `
			SELECT full_task.*
			FROM full_task
			WHERE full_task.entity_path LIKE ? OR full_task.entity_path LIKE ?;
			`

			var entityTasks []models.Task
			err = tx.Select(&entityTasks, entityTasksQuery, parentEntity.EntityPath, pathLike)
			if err != nil {
				return err
			}
			allTasks = append(allTasks, entityTasks...)
		}
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return err
	}

	// Create entity directories
	for _, entity := range entities {
		entityPath, err := utils.BuildEntityPath(rootFolder, entity.EntityPath)
		if err != nil {
			return err
		}
		err = os.MkdirAll(entityPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Get all task IDs for batch checkpoint retrieval
	taskIds := []string{}
	for _, task := range allTasks {
		taskIds = append(taskIds, task.Id)
	}

	// Batch fetch checkpoints for all tasks
	if len(taskIds) > 0 {
		quotedTaskIds := make([]string, len(taskIds))
		for i, id := range taskIds {
			quotedTaskIds[i] = fmt.Sprintf("\"%s\"", id)
		}

		checkpoints := []models.Checkpoint{}
		err = tx.Select(&checkpoints, fmt.Sprintf("SELECT * FROM task_checkpoint WHERE trashed = 0 AND task_id IN (%s) ORDER BY created_at DESC", strings.Join(quotedTaskIds, ",")))
		if err != nil {
			return err
		}

		taskCheckpoints := map[string][]models.Checkpoint{}
		for _, taskCheckpoint := range checkpoints {
			taskCheckpoints[taskCheckpoint.TaskId] = append(taskCheckpoints[taskCheckpoint.TaskId], taskCheckpoint)
		}

		for i, task := range allTasks {
			allTasks[i].Checkpoints = taskCheckpoints[task.Id]
		}
	}

	tasksToRebuild := []models.Task{}
	for _, task := range allTasks {
		taskFilePath, err := utils.BuildTaskPath(rootFolder, task.EntityPath, task.Name, task.Extension)
		if err != nil {
			return err
		}
		task.FilePath = taskFilePath
		if _, err := os.Stat(task.GetFilePath()); os.IsNotExist(err) {
			tasksToRebuild = append(tasksToRebuild, task)
		}
	}

	checkpointIdsToDownload := []string{}
	for _, task := range tasksToRebuild {
		latestCheckpoint := task.Checkpoints[0]
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

	tx, err = dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	totalItems := len(tasksToRebuild)
	for i, task := range tasksToRebuild {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Rebuilding Files",
				Message:    task.Name,
				Percentage: float64(current) / float64(total) * 100,
				Current:    i + 1,
				Total:      totalItems,
			}
			app.EmitEvent("progress-update", progress)
		}
		err = repository.RevertToLatestCheckpoint(tx, task.Id, task.FilePath, callBack)
		if err != nil {
			return err
		}
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

func (e *CollectionService) RevealEntity(projectPath, entityId string) error {
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

	entity, err := repository.GetEntity(tx, entityId)
	if err != nil {
		return err
	}
	utils.RevealInExplorer(entity.GetFilePath())
	return nil
}

func (e *CollectionService) RevertEntities(projectPath string, entityIds []string) error {
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

	totalEntities := len(entityIds)
	for i, entityId := range entityIds {
		tx, err := dbConn.Beginx()
		if err != nil {
			return err
		}
		entity, err := repository.GetEntity(tx, entityId)
		if err != nil {
			return err
		}
		callBack := func(current int, total int, message string, extraMessage string) {
			progress := output.ProgressReport{
				Title:      "Reverting",
				Message:    entity.EntityTypeName,
				Percentage: float64(current) / float64(total) * 100,
				Current:    i + 1,
				Total:      totalEntities,
			}
			app.EmitEvent("progress-update", progress)
		}

		err = repository.RevertToLatestCheckpoint(tx, entityId, entity.FilePath, callBack)
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
	}
	return nil
}

func (e *CollectionService) ChangeEntityParent(projectPath, entityId, parentId string) error {
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

	err = repository.ChangeParent(tx, entityId, parentId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (e *CollectionService) ChangeType(projectPath, entityId, entityTypeId string) error {
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

	err = repository.ChangeEntityType(tx, entityId, entityTypeId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (e *CollectionService) ChangeIsLibrary(projectPath, entityId string, isLibrary bool) error {
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

	err = repository.ChangeIsLibrary(tx, entityId, isLibrary)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (e *CollectionService) Assign(projectPath, entityId, userId string) error {
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

	err = repository.AssignEntity(tx, entityId, userId)
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
func (e *CollectionService) Unassign(projectPath, entityId, userId string) error {
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

	err = repository.UnAssignEntity(tx, entityId, userId)
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

// entity types
func (e *CollectionService) CreateEntityType(projectPath, entityTypeName, entityTypeIcon string) (models.EntityType, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.EntityType{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.EntityType{}, err
	}
	defer tx.Rollback()

	entityType, err := repository.CreateEntityType(tx, "", entityTypeName, entityTypeIcon)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: entity_type.name" {
			tx.Rollback()
			return models.EntityType{}, error_service.ErrEntityTypeExists
		}
		tx.Rollback()
		return models.EntityType{}, err
	}
	tx.Commit()
	return entityType, nil
}
func (e *CollectionService) UpdateEntityType(projectPath, id, entityTypeName, entityTypeIcon string) (models.EntityType, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.EntityType{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.EntityType{}, err
	}
	defer tx.Rollback()

	entityType, err := repository.UpdateEntityType(tx, id, entityTypeName, entityTypeIcon)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: entity_type.name" {
			tx.Rollback()
			return models.EntityType{}, error_service.ErrEntityTypeExists
		}
		tx.Rollback()
		return models.EntityType{}, err
	}
	err = tx.Commit()
	if err != nil {
		return models.EntityType{}, err
	}
	return entityType, nil
}

func (e *CollectionService) DeleteEntityType(projectPath, id string) error {
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

	err = repository.DeleteEntityType(tx, id)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (e *CollectionService) GetEntityTypes(projectPath string) ([]models.EntityType, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.EntityType{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.EntityType{}, err
	}
	defer tx.Rollback()

	entityTypes, err := repository.GetEntityTypes(tx)
	if err != nil {
		return entityTypes, err
	}
	return entityTypes, nil
}

func (p *CollectionService) UpdatePreview(projectPath, entityId, previewPath string) error {
	if !utils.FileExists(projectPath) {
		return error_service.ErrProjectNotFound
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

	err = repository.UpdateEntityPreview(tx, entityId, previewPath)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
