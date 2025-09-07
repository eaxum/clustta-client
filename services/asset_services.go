package services

import (
	"bytes"
	"clustta/internal/auth_service"
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

func (t *AssetService) GetTaskCount(projectPath string) (int, error) {
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
func (t *AssetService) CreateTask(projectPath, name, description, taskTypeId, entityId string, isResource bool, templateId, templateFilePath, pointer string, isLink bool, tags []string, previewPath, comment string) (models.Task, error) {
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
			Title:      "Creating Tasks for Entity",
			Message:    name,
			Percentage: float64(current) / float64(total) * 99,
			Current:    1,
			Total:      2,
		}
		app.EmitEvent("progress-update", progress)
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
			Title:      "Creating Tasks for Entity",
			Message:    "Task",
			Percentage: 100,
			Current:    1,
			Total:      1,
		}
		app.EmitEvent("progress-update", progress)
	}
	return createdTask, nil
}

func (t *AssetService) DuplicateTask(projectPath, sourceTaskId string) (models.Task, error) {
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

func (t *AssetService) ChangeTaskEntity(projectPath, taskId, entityId string) error {
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

	repository.ChangeEntity(tx, taskId, entityId)
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
func (t *AssetService) DeleteTask(projectPath, taskId string, removeFiles bool) error {
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

func (t *AssetService) UpdateTask(projectPath, taskId, name, taskTypeId string, isResource bool, pointer string, tags []string) (models.Task, error) {
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

func (t *AssetService) ChangeTaskType(projectPath, taskId, taskTypeId string) error {
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

func (t *AssetService) RenameTask(projectPath, taskId, name string) (models.Task, error) {
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

func (t *AssetService) AssignTask(projectPath, taskId, userId string) error {
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

	err = repository.AssignTask(tx, taskId, userId)
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
func (t *AssetService) UnassignTask(projectPath, taskId string) error {
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
func (t *AssetService) TaskFileStatus(projectPath, taskId string) (string, error) {
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
func (t *AssetService) TaskFilesStatus(projectPath string, taskIds []string) (map[string]string, error) {
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

func (t *AssetService) RevealTask(projectPath, taskId string) error {
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
func (t *AssetService) AddTaskDependency(projectPath, taskId, dependencyId, dependencyTypeId string) (models.TaskDependency, error) {
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
func (t *AssetService) RemoveTaskDependency(projectPath, taskId, dependencyId string) error {
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
func (t *AssetService) GetTaskDependencies2(projectPath string, taskIds []string) ([]models.Task, error) {
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
func (t *AssetService) GetTaskDependencies(projectPath string, taskIds []string) ([]interface{}, error) {
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

func (t *AssetService) GetTasks(projectPath string) ([]models.Task, error) {
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

func (t *AssetService) GetTasksPB(projectPath string) ([]byte, error) {
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

// task types
func (t *AssetService) GetTaskTypes(projectPath string) ([]models.TaskType, error) {
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

func (t *AssetService) DeleteTaskType(projectPath, id string) error {
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

func (t *AssetService) CreateTaskType(projectPath, name, icon string) (models.TaskType, error) {
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

func (t *AssetService) UpdateTaskType(projectPath, id, name, icon string) (models.TaskType, error) {
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
