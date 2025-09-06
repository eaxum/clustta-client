package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"clustta/internal/base_service"
	"clustta/internal/error_service"
	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

type Dependency struct {
	Id       string `json:"id"`
	TypeId   string `json:"type_id"`
	TypeName string `json:"type_name"`
}
type TaskTags struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func CreateTaskFast(
	tx *sqlx.Tx, id, name, taskTypeId, entityId string, isResource bool, description, template_file_path string, previewId, userId, comment, checkpointGroupId, taskPath, statusId string,
	callback func(int, int, string, string),
) error {
	if id == "" {
		id = uuid.New().String()
	}
	if checkpointGroupId == "" {
		checkpointGroupId = uuid.New().String()
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("task name cannot be empty")
	}

	if entityId != "" {
		parent := models.Entity{}
		err := base_service.Get(tx, "entity", entityId, &parent)
		if err != nil {
			return err
		}
	}

	extension := filepath.Ext(template_file_path)

	if template_file_path == "" {
		return errors.New("template file does not exist")
	}

	chunkSequence, err := StoreFileChunks(tx, template_file_path, callback)
	if err != nil {
		return err
	}
	checksum, err := utils.GenerateXXHashChecksum(template_file_path)
	if err != nil {
		return err
	}
	fileInfo, err := os.Stat(template_file_path)
	if err != nil {
		return err
	}
	timeModified := int(fileInfo.ModTime().Unix())
	fileSize := int(fileInfo.Size())

	params := map[string]any{
		"id":           id,
		"created_at":   utils.GetCurrentTime(),
		"name":         name,
		"description":  description,
		"extension":    extension,
		"task_type_id": taskTypeId,
		"entity_id":    entityId,
		"is_resource":  isResource,
		"status_id":    statusId,
		"pointer":      "",
		"is_link":      false,
		"preview_id":   previewId,
	}
	err = base_service.Create(tx, "task", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return error_service.ErrTaskExists
			}
		}
		return err
	}

	author_id := userId
	if comment == "" {
		comment = "new file"
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return err
	}
	taskFilePath := filepath.Join(rootFolder, taskPath+extension)

	err = CreateNewTaskCheckpoint(tx, id, comment, chunkSequence, checksum, timeModified, fileSize, taskFilePath, author_id, "", checkpointGroupId, func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		return err
	}
	return nil
}

func CreateTask(
	tx *sqlx.Tx, id, name, taskTypeId, entityId string, isResource bool,
	templateId, description, template_file_path string, tags []string, pointer string, isLink bool, previewId, userId, comment, checkpointGroupId string,
	callback func(int, int, string, string),
) (models.Task, error) {
	if checkpointGroupId == "" {
		checkpointGroupId = uuid.New().String()
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return models.Task{}, errors.New("task name cannot be empty")
	}

	if isLink && !utils.IsValidPointer(pointer) {
		return models.Task{}, errors.New("invalid pointer, path does not exist")
	}

	if !isLink && templateId == "" && template_file_path == "" {
		return models.Task{}, errors.New("template not provided")
	}

	if entityId != "" {
		parent := models.Entity{}
		err := base_service.Get(tx, "entity", entityId, &parent)
		if err != nil {
			return models.Task{}, err
		}
	}

	task := models.Task{}

	chunkSequence := ""
	checksum := ""
	timeModified := 0
	fileSize := 0
	extension := ""
	if !isLink {
		if template_file_path == "" {
			if templateId == "" {
				return models.Task{}, errors.New("no template provided")
			}
			template, err := GetTemplate(tx, templateId)
			if err != nil {
				return models.Task{}, err
			}

			// if !utils.NonCaseSensitiveContains(tags, template.Name) {
			// 	tags = append(tags, template.Name)
			// }
			extension = template.Extension

			conditions := map[string]interface{}{
				"name":      name,
				"entity_id": entityId,
				"extension": extension,
			}
			err = base_service.GetBy(tx, "task", conditions, &task)
			if err == nil {
				if task.Trashed {
					return models.Task{}, error_service.ErrTaskExistsInTrash
				} else {
					return models.Task{}, error_service.ErrTaskExists
				}
			}

			chunkSequence = template.Chunks
			checksum = template.XxhashChecksum
			fileSize = template.FileSize
			timeModified = int(utils.GetEpochTime())
		} else {
			if template_file_path == "" {
				return models.Task{}, errors.New("template file does not exist")
			}
			extension = filepath.Ext(template_file_path)

			conditions := map[string]interface{}{
				"name":      name,
				"entity_id": entityId,
				"extension": extension,
			}
			err := base_service.GetBy(tx, "task", conditions, &task)
			if err == nil {
				if task.Trashed {
					return models.Task{}, error_service.ErrTaskExistsInTrash
				} else {
					return models.Task{}, error_service.ErrTaskExists
				}
			}

			Sequence, err := StoreFileChunks(tx, template_file_path, callback)
			chunkSequence = Sequence
			if err != nil {
				return models.Task{}, err
			}
			checksum, err = utils.GenerateXXHashChecksum(template_file_path)
			if err != nil {
				return models.Task{}, err
			}
			fileInfo, err := os.Stat(template_file_path)
			if err != nil {
				return models.Task{}, err
			}
			timeModified = int(fileInfo.ModTime().Unix())
			fileSize = int(fileInfo.Size())
		}
	} else {
		conditions := map[string]interface{}{
			"name":      name,
			"entity_id": entityId,
		}
		err := base_service.GetBy(tx, "task", conditions, &task)
		if err == nil {
			if task.Trashed {
				return models.Task{}, error_service.ErrTaskExistsInTrash
			} else {
				return models.Task{}, error_service.ErrTaskExists
			}
		}
	}
	status, err := GetStatusByShortName(tx, "todo")
	if err != nil {
		return models.Task{}, err
	}
	statusId := status.Id
	params := map[string]any{
		"id":           id,
		"created_at":   utils.GetCurrentTime(),
		"name":         name,
		"description":  description,
		"extension":    extension,
		"task_type_id": taskTypeId,
		"entity_id":    entityId,
		"is_resource":  isResource,
		"status_id":    statusId,
		"pointer":      pointer,
		"is_link":      isLink,
		"preview_id":   previewId,
	}
	err = base_service.Create(tx, "task", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return models.Task{}, error_service.ErrTaskExists
			}
		}
		return models.Task{}, err
	}
	task, err = GetTaskByName(tx, name, entityId, extension)
	if err != nil {
		return models.Task{}, err
	}
	// tx.Get(&task, "SELECT * FROM task WHERE name = ?", name)

	// if isLink {
	// 	if !utils.NonCaseSensitiveContains(tags, "link") {
	// 		tags = append(tags, "link")
	// 	}
	// }
	// if !isLink && pointer != "" {
	// 	if !utils.NonCaseSensitiveContains(tags, "tracked") {
	// 		tags = append(tags, "tracked")
	// 	}
	// }

	if len(tags) > 0 {
		for _, tag := range tags {
			err = AddTagToTask(tx, task.Id, tag)
			if err != nil {
				return models.Task{}, err
			}
			task.Tags = append(task.Tags, tag)
		}
	}
	if !isLink {
		author_id := userId
		if comment == "" {
			comment = "new file"
		}
		_, err = CreateCheckpoint(tx, task.Id, comment, chunkSequence, checksum, timeModified, fileSize, task.FilePath, author_id, "", checkpointGroupId, func(i1, i2 int, s1, s2 string) {})
		if err != nil {
			return models.Task{}, err
		}
	}
	return task, nil
}

func AddTask(
	tx *sqlx.Tx, id string, createdAt string, name string, taskTypeId, entityId string,
	statusId string, extension string, description string, tags []string,
	pointer string, isLink bool, assignee_id string, previewId string,
) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("task name cannot be empty")
	}

	if isLink && !utils.IsValidPointer(pointer) {
		return errors.New("invalid pointer, path does not exist")
	}

	params := map[string]interface{}{
		"id":           id,
		"created_at":   createdAt,
		"name":         name,
		"description":  description,
		"extension":    extension,
		"task_type_id": taskTypeId,
		"entity_id":    entityId,
		"status_id":    statusId,
		"pointer":      pointer,
		"is_link":      isLink,
		"assignee_id":  assignee_id,
		"preview_id":   previewId,
	}
	err := base_service.Create(tx, "task", params)
	if err != nil {
		return err
	}
	return nil
}

func GetTask(tx *sqlx.Tx, id string) (models.Task, error) {
	task := models.Task{}

	query := "SELECT * FROM full_task WHERE id = ?"

	err := tx.Get(&task, query, id)
	if err != nil && err == sql.ErrNoRows {
		return models.Task{}, error_service.ErrTaskNotFound
	} else if err != nil {
		return models.Task{}, err
	}

	if task.TagsRaw != "[]" {
		taskTags := []TaskTags{}
		err = json.Unmarshal([]byte(task.TagsRaw), &taskTags)
		if err != nil {
			return task, err
		}
		for _, taskTag := range taskTags {
			task.Tags = append(task.Tags, taskTag.Name)
		}
	} else {
		task.Tags = []string{} // Ensure it's initialized as an empty slice
	}

	if task.EntityDependenciesRaw != "[]" {
		entityDependencies := []Dependency{}
		err = json.Unmarshal([]byte(task.EntityDependenciesRaw), &entityDependencies)
		if err != nil {
			return task, err
		}
		for _, entityDependency := range entityDependencies {
			task.EntityDependencies = append(task.EntityDependencies, entityDependency.Id)
		}
	} else {
		task.EntityDependencies = []string{} // Ensure it's initialized as an empty slice
	}

	if task.DependenciesRaw != "[]" {
		taskDependencies := []Dependency{}
		err = json.Unmarshal([]byte(task.DependenciesRaw), &taskDependencies)
		if err != nil {
			return task, err
		}
		for _, taskDependency := range taskDependencies {
			task.Dependencies = append(task.Dependencies, taskDependency.Id)
		}
	} else {
		task.Dependencies = []string{} // Ensure it's initialized as an empty slice
	}

	status, err := GetStatus(tx, task.StatusId)
	if err != nil {
		return task, err
	}
	task.Status = status
	task.StatusShortName = status.ShortName

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return task, err
	}

	taskFilePath, err := utils.BuildTaskPath(rootFolder, task.EntityPath, task.Name, task.Extension)
	if err != nil {
		return task, err
	}
	task.FilePath = taskFilePath
	checkpoints, err := GetCheckpoints(tx, task.Id, false)
	if err != nil {
		return task, err
	}
	task.Checkpoints = checkpoints
	fileStatus, err := GetTaskFileStatus(&task, checkpoints)
	if err != nil {
		return task, err
	}
	task.FileStatus = fileStatus
	return task, nil
}

func GetSimpleTasks(tx *sqlx.Tx) ([]models.Task, error) {
	tasks := []models.Task{}

	query := "SELECT * FROM task"

	err := tx.Select(&tasks, query)
	if err != nil && err == sql.ErrNoRows {
		return []models.Task{}, nil
	} else if err != nil {
		return []models.Task{}, err
	}
	return tasks, nil
}
func GetSimpleTask(tx *sqlx.Tx, id string) (models.Task, error) {
	task := models.Task{}

	query := "SELECT * FROM task WHERE id = ?"

	err := tx.Get(&task, query, id)
	if err != nil && err == sql.ErrNoRows {
		return models.Task{}, error_service.ErrTaskNotFound
	} else if err != nil {
		return models.Task{}, err
	}
	return task, nil
}

func GetTaskByName(tx *sqlx.Tx, name, entityId string, extension string) (models.Task, error) {
	task := models.Task{}
	query := "SELECT * FROM full_task WHERE name = ? AND entity_id = ? AND extension = ?"

	err := tx.Get(&task, query, name, entityId, extension)
	if err != nil && err == sql.ErrNoRows {
		return models.Task{}, error_service.ErrTaskNotFound
	} else if err != nil {
		return models.Task{}, err
	}

	if task.TagsRaw != "[]" {
		taskTags := []TaskTags{}
		err = json.Unmarshal([]byte(task.TagsRaw), &taskTags)
		if err != nil {
			return task, err
		}
		for _, taskTag := range taskTags {
			task.Tags = append(task.Tags, taskTag.Name)
		}
	} else {
		task.Tags = []string{} // Ensure it's initialized as an empty slice
	}

	if task.EntityDependenciesRaw != "[]" {
		entityDependencies := []Dependency{}
		err = json.Unmarshal([]byte(task.EntityDependenciesRaw), &entityDependencies)
		if err != nil {
			return task, err
		}
		for _, entityDependency := range entityDependencies {
			task.EntityDependencies = append(task.EntityDependencies, entityDependency.Id)
		}
	} else {
		task.EntityDependencies = []string{} // Ensure it's initialized as an empty slice
	}
	if task.DependenciesRaw != "[]" {
		taskDependencies := []Dependency{}
		err = json.Unmarshal([]byte(task.DependenciesRaw), &taskDependencies)
		if err != nil {
			return task, err
		}
		for _, taskDependency := range taskDependencies {
			task.Dependencies = append(task.Dependencies, taskDependency.Id)
		}
	} else {
		task.Dependencies = []string{} // Ensure it's initialized as an empty slice
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return task, err
	}

	taskFilePath, err := utils.BuildTaskPath(rootFolder, task.EntityPath, task.Name, task.Extension)
	if err != nil {
		return task, err
	}
	checkpoints, err := GetCheckpoints(tx, task.Id, false)
	if err != nil {
		return task, err
	}
	task.Checkpoints = checkpoints
	task.FilePath = taskFilePath
	fileStatus, err := GetTaskFileStatus(&task, checkpoints)
	if err != nil {
		return task, err
	}
	task.FileStatus = fileStatus
	return task, nil
}

func GetTaskByPath(tx *sqlx.Tx, taskPath string) (models.Task, error) {
	task := models.Task{}
	query := "SELECT * FROM full_task WHERE task_path = ?"

	err := tx.Get(&task, query, taskPath)
	if err != nil && err == sql.ErrNoRows {
		return models.Task{}, error_service.ErrTaskNotFound
	} else if err != nil {
		return models.Task{}, err
	}

	if task.TagsRaw != "[]" {
		taskTags := []TaskTags{}
		err = json.Unmarshal([]byte(task.TagsRaw), &taskTags)
		if err != nil {
			return task, err
		}
		for _, taskTag := range taskTags {
			task.Tags = append(task.Tags, taskTag.Name)
		}
	} else {
		task.Tags = []string{} // Ensure it's initialized as an empty slice
	}

	if task.EntityDependenciesRaw != "[]" {
		entityDependencies := []Dependency{}
		err = json.Unmarshal([]byte(task.EntityDependenciesRaw), &entityDependencies)
		if err != nil {
			return task, err
		}
		for _, entityDependency := range entityDependencies {
			task.EntityDependencies = append(task.EntityDependencies, entityDependency.Id)
		}
	} else {
		task.EntityDependencies = []string{} // Ensure it's initialized as an empty slice
	}
	if task.DependenciesRaw != "[]" {
		taskDependencies := []Dependency{}
		err = json.Unmarshal([]byte(task.DependenciesRaw), &taskDependencies)
		if err != nil {
			return task, err
		}
		for _, taskDependency := range taskDependencies {
			task.Dependencies = append(task.Dependencies, taskDependency.Id)
		}
	} else {
		task.Dependencies = []string{} // Ensure it's initialized as an empty slice
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return task, err
	}
	taskFilePath, err := utils.BuildTaskPath(rootFolder, task.EntityPath, task.Name, task.Extension)
	if err != nil {
		return task, err
	}
	checkpoints, err := GetCheckpoints(tx, task.Id, false)
	if err != nil {
		return task, err
	}
	task.Checkpoints = checkpoints
	task.FilePath = taskFilePath
	fileStatus, err := GetTaskFileStatus(&task, checkpoints)
	if err != nil {
		return task, err
	}
	task.FileStatus = fileStatus
	return task, nil
}

func GetTasks(tx *sqlx.Tx, withDeleted bool) ([]models.Task, error) {
	tasks := []models.Task{}
	queryWhereClause := ""
	if !withDeleted {
		queryWhereClause = "WHERE trashed = 0"
	}
	query := fmt.Sprintf("SELECT * FROM full_task %s ORDER BY task_path", queryWhereClause)

	err := tx.Select(&tasks, query)
	if err != nil {
		return tasks, err
	}

	statuses, err := GetStatuses(tx)
	if err != nil {
		return tasks, err
	}
	statusesMap := map[string]models.Status{}
	for _, status := range statuses {
		statusesMap[status.Id] = status
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
		return tasks, err
	}

	for i, task := range tasks {
		status := statusesMap[task.StatusId]
		tasks[i].Status = status
		tasks[i].StatusShortName = status.ShortName

		if tasks[i].TagsRaw != "[]" {
			taskTags := []TaskTags{}
			err = json.Unmarshal([]byte(task.TagsRaw), &taskTags)
			if err != nil {
				return tasks, err
			}
			for _, taskTag := range taskTags {
				tasks[i].Tags = append(tasks[i].Tags, taskTag.Name)
			}
		} else {
			tasks[i].Tags = []string{} // Ensure it's initialized as an empty slice
		}

		if task.EntityDependenciesRaw != "[]" {
			entityDependencies := []Dependency{}
			err = json.Unmarshal([]byte(task.EntityDependenciesRaw), &entityDependencies)
			if err != nil {
				return tasks, err
			}
			for _, entityDependency := range entityDependencies {
				tasks[i].EntityDependencies = append(tasks[i].EntityDependencies, entityDependency.Id)
			}
		} else {
			tasks[i].EntityDependencies = []string{} // Ensure it's initialized as an empty slice
		}
		if task.DependenciesRaw != "[]" {
			taskDependencies := []Dependency{}
			err = json.Unmarshal([]byte(task.DependenciesRaw), &taskDependencies)
			if err != nil {
				return tasks, err
			}
			for _, taskDependency := range taskDependencies {
				tasks[i].Dependencies = append(tasks[i].Dependencies, taskDependency.Id)
			}
		} else {
			tasks[i].Dependencies = []string{} // Ensure it's initialized as an empty slice
		}

		taskFilePath, err := utils.BuildTaskPath(rootFolder, task.EntityPath, task.Name, task.Extension)
		if err != nil {
			return tasks, err
		}
		tasks[i].FilePath = taskFilePath
		tasks[i].Checkpoints = taskCheckpoints[task.Id]

		// fileStatus, err := GetTaskFileStatus(&tasks[i], taskCheckpoints[task.Id])
		// if err != nil {
		// 	return tasks, err
		// }
		// tasks[i].FileStatus = fileStatus
	}
	return tasks, nil
}

func OLDGetUserTasks(tx *sqlx.Tx, userId string) ([]models.Task, error) {
	tasks := []models.Task{}
	// query := "SELECT * FROM full_task WHERE trashed = 0 AND assignee_id = ?"
	query := fmt.Sprintf(`
		WITH RECURSIVE 
			-- First get all entity dependencies recursively
			entity_dependencies AS (
				-- Base case: Get direct entity dependencies
				SELECT 
					ed.task_id,
					e.id as entity_id,
					0 as entity_depth,
					CAST(e.id AS TEXT) as entity_path
				FROM entity_dependency ed
				JOIN entity e ON ed.dependency_id = e.id
				WHERE e.trashed = 0

				UNION ALL

				-- Recursive case: Get child entities
				SELECT 
					ed.task_id,
					e.id as entity_id,
					ed.entity_depth + 1,
					ed.entity_path || ',' || e.id
				FROM entity e
				JOIN entity_dependencies ed ON e.parent_id = ed.entity_id
				WHERE e.trashed = 0
				AND e.id NOT IN (
					SELECT value 
					FROM json_each('["' || REPLACE(ed.entity_path, ',', '","') || '"]')
				)
			),
			
			-- Then get all tasks and their dependencies
			task_dependencies AS (
				-- Base case: Get all tasks directly assigned to the user
				SELECT 
					t.*,
					0 as dependency_level,
					0 as is_dependency,  -- Directly assigned tasks
					t.id as root_task_id,  -- Track the original assigned task
					CAST(t.id AS TEXT) as dependency_path,  -- Track the path to prevent cycles
					NULL as via_entity  -- Track if task came through entity dependency
				FROM full_task t
				WHERE t.assignee_id = '%s' 
				AND t.trashed = 0

				UNION ALL

				-- Get tasks through entity dependencies
				SELECT 
					t.*,
					td.dependency_level,
					1 as is_dependency,
					td.root_task_id,
					td.dependency_path || ',' || t.id,
					ed.entity_id as via_entity
				FROM full_task t
				JOIN entity_dependencies ed ON t.entity_id = ed.entity_id
				JOIN task_dependencies td ON ed.task_id = td.id
				WHERE t.trashed = 0
				AND t.id NOT IN (
					SELECT value 
					FROM json_each('["' || REPLACE(td.dependency_path, ',', '","') || '"]')
				)

				UNION ALL

				-- Get direct task dependencies
				SELECT 
					t.*,
					td.dependency_level + 1,
					1 as is_dependency,
					td.root_task_id,
					td.dependency_path || ',' || t.id,
					NULL as via_entity
				FROM full_task t
				JOIN task_dependency dep ON t.id = dep.dependency_id
				JOIN task_dependencies td ON dep.task_id = td.id
				WHERE t.trashed = 0
				AND t.id NOT IN (
					SELECT value 
					FROM json_each('["' || REPLACE(td.dependency_path, ',', '","') || '"]')
				)
			),
			
			-- Select the highest-level occurrence of each task
			ranked_dependencies AS (
				SELECT 
					*,
					ROW_NUMBER() OVER (
						PARTITION BY id 
						ORDER BY 
							-- Prefer directly assigned tasks
							CASE WHEN assignee_id = '%s' THEN 0 ELSE 1 END,
							-- Then prefer the shortest dependency chain
							dependency_level,
							-- For equal levels, prefer alphabetical order of root task
							root_task_id
					) as rank
				FROM task_dependencies
			)
		SELECT 
			id,
			created_at,
			mtime,
			name,
			description,
			extension,
			is_resource,
			is_link,
			pointer,
			status_id,
			task_type_id,
			entity_id,
			assignee_id,
			assigner_id,
			preview_id,
			trashed,
			synced,
			task_type_icon,
			task_type_name,
			entity_name,
			preview_extension,
			preview,
			COALESCE(entity_path, '') as entity_path, -- Handle NULL
			task_path,
			assignee_name,
			assignee_email,
			assigner_name,
			assigner_email,
			tags,
			dependencies,
			entity_dependencies,
			dependency_level,
			is_dependency
		FROM ranked_dependencies
		WHERE rank = 1
		ORDER BY 
			dependency_level,
			name;
	`, userId, userId)

	err := tx.Select(&tasks, query)
	if err != nil {
		return tasks, err
	}

	statuses, err := GetStatuses(tx)
	if err != nil {
		return tasks, err
	}
	statusesMap := map[string]models.Status{}
	for _, status := range statuses {
		statusesMap[status.Id] = status
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
		return tasks, err
	}
	for i, task := range tasks {
		status := statusesMap[task.StatusId]
		tasks[i].StatusShortName = status.ShortName
		tasks[i].Status = status

		if tasks[i].TagsRaw != "[]" {
			taskTags := []TaskTags{}
			err = json.Unmarshal([]byte(task.TagsRaw), &taskTags)
			if err != nil {
				return tasks, err
			}
			for _, taskTag := range taskTags {
				tasks[i].Tags = append(tasks[i].Tags, taskTag.Name)
			}
		} else {
			tasks[i].Tags = []string{} // Ensure it's initialized as an empty slice
		}

		if task.EntityDependenciesRaw != "[]" {
			entityDependencies := []Dependency{}
			err = json.Unmarshal([]byte(task.EntityDependenciesRaw), &entityDependencies)
			if err != nil {
				return tasks, err
			}
			for _, entityDependency := range entityDependencies {
				tasks[i].EntityDependencies = append(tasks[i].EntityDependencies, entityDependency.Id)
			}
		} else {
			tasks[i].EntityDependencies = []string{} // Ensure it's initialized as an empty slice
		}
		if task.DependenciesRaw != "[]" {
			taskDependencies := []Dependency{}
			err = json.Unmarshal([]byte(task.DependenciesRaw), &taskDependencies)
			if err != nil {
				return tasks, err
			}
			for _, taskDependency := range taskDependencies {
				tasks[i].Dependencies = append(tasks[i].Dependencies, taskDependency.Id)
			}
		} else {
			tasks[i].Dependencies = []string{} // Ensure it's initialized as an empty slice
		}

		taskFilePath, err := utils.BuildTaskPath(rootFolder, task.EntityPath, task.Name, task.Extension)
		if err != nil {
			return tasks, err
		}
		tasks[i].FilePath = taskFilePath
		tasks[i].Checkpoints = taskCheckpoints[task.Id]

		// fileStatus := "normal"
		fileStatus, err := GetTaskFileStatus(&tasks[i], taskCheckpoints[task.Id])
		if err != nil {
			return tasks, err
		}
		tasks[i].FileStatus = fileStatus
	}

	return tasks, nil
}

func OLDGetUserTasksX(tx *sqlx.Tx, userId string) ([]models.Task, error) {
	// Get all tasks assigned to the user
	assignedTaskIds := []string{}
	query := "SELECT id FROM task WHERE assignee_id = ? AND trashed = 0"
	err := tx.Select(&assignedTaskIds, query, userId)
	if err != nil {
		return nil, err
	}

	// Get all tasks that are dependencies of the assigned tasks and dependencies of those tasks recursively
	dependencyTaskIds := []string{}
	query = `
		WITH RECURSIVE task_dependencies AS (
			-- Base case: Get direct dependencies
			SELECT 
				td.dependency_id, 
				0 AS dependency_level, 
				CAST(td.dependency_id AS TEXT) AS dependency_path
			FROM task_dependency td
			JOIN task t ON td.task_id = t.id
			WHERE t.assignee_id = ? AND t.trashed = 0

			UNION ALL

			-- Recursive case: Get dependencies of dependencies
			SELECT 
				td.dependency_id, 
				td2.dependency_level + 1 AS dependency_level, 
				td2.dependency_path || ',' || td.dependency_id AS dependency_path
			FROM task_dependency td
			JOIN task_dependencies td2 ON td.task_id = td2.dependency_id
			WHERE td.dependency_id NOT IN (
				SELECT value
				FROM json_each('["' || REPLACE(td2.dependency_path, ',', '","') || '"]')
			)
		)
		SELECT DISTINCT t.id
		FROM task t
		JOIN task_dependencies td ON t.id = td.dependency_id
	`
	err = tx.Select(&dependencyTaskIds, query, userId)
	if err != nil {
		return nil, err
	}

	// Get all tasks that are dependencies of the assigned tasks through entity dependencies
	entityTasksDependencyIds := []string{}
	query = `
		WITH RECURSIVE entity_dependencies AS (
			SELECT ed.task_id, e.id as entity_id, 0 as entity_depth, CAST(e.id AS TEXT) as entity_path
			FROM entity_dependency ed
			JOIN entity e ON ed.dependency_id = e.id
			WHERE e.trashed = 0

			UNION ALL

			SELECT ed.task_id, e.id as entity_id, ed.entity_depth + 1, ed.entity_path || ',' || e.id
			FROM entity e
			JOIN entity_dependencies ed ON e.parent_id = ed.entity_id
			WHERE e.trashed = 0
			AND e.id NOT IN (
				SELECT value
				FROM json_each('["' || REPLACE(ed.entity_path, ',', '","') || '"]')
			)
		)
		SELECT DISTINCT t.id
		FROM task t
		JOIN entity_dependencies ed ON t.entity_id = ed.entity_id
		JOIN task_dependency td ON t.id = td.dependency_id
		WHERE td.task_id IN (SELECT id FROM task WHERE assignee_id = ? AND trashed = 0)
		AND t.trashed = 0
	`
	err = tx.Select(&entityTasksDependencyIds, query, userId)
	if err != nil {
		return nil, err
	}

	// Combine all task IDs
	allTaskIds := make(map[string]struct{})
	for _, id := range assignedTaskIds {
		allTaskIds[id] = struct{}{}
	}
	for _, id := range dependencyTaskIds {
		allTaskIds[id] = struct{}{}
	}
	for _, id := range entityTasksDependencyIds {
		allTaskIds[id] = struct{}{}
	}

	// Convert map keys to slice
	taskIds := make([]string, 0, len(allTaskIds))
	for id := range allTaskIds {
		taskIds = append(taskIds, id)
	}

	// Get all tasks with the collected IDs
	query = `
		SELECT * FROM full_task WHERE id IN (SELECT value FROM json_each(?)) AND trashed = 0
	`
	tasks := []models.Task{}
	jsonTaskIds, err := json.Marshal(taskIds)
	if err != nil {
		return nil, err
	}
	err = tx.Select(&tasks, query, jsonTaskIds)
	if err != nil {
		return nil, err
	}

	statuses, err := GetStatuses(tx)
	if err != nil {
		return tasks, err
	}
	statusesMap := map[string]models.Status{}
	for _, status := range statuses {
		statusesMap[status.Id] = status
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
		return tasks, err
	}
	for i, task := range tasks {
		status := statusesMap[task.StatusId]
		tasks[i].StatusShortName = status.ShortName
		tasks[i].Status = status

		if tasks[i].TagsRaw != "[]" {
			taskTags := []TaskTags{}
			err = json.Unmarshal([]byte(task.TagsRaw), &taskTags)
			if err != nil {
				return tasks, err
			}
			for _, taskTag := range taskTags {
				tasks[i].Tags = append(tasks[i].Tags, taskTag.Name)
			}
		} else {
			tasks[i].Tags = []string{} // Ensure it's initialized as an empty slice
		}

		if task.EntityDependenciesRaw != "[]" {
			entityDependencies := []Dependency{}
			err = json.Unmarshal([]byte(task.EntityDependenciesRaw), &entityDependencies)
			if err != nil {
				return tasks, err
			}
			for _, entityDependency := range entityDependencies {
				tasks[i].EntityDependencies = append(tasks[i].EntityDependencies, entityDependency.Id)
			}
		} else {
			tasks[i].EntityDependencies = []string{} // Ensure it's initialized as an empty slice
		}
		if task.DependenciesRaw != "[]" {
			taskDependencies := []Dependency{}
			err = json.Unmarshal([]byte(task.DependenciesRaw), &taskDependencies)
			if err != nil {
				return tasks, err
			}
			for _, taskDependency := range taskDependencies {
				tasks[i].Dependencies = append(tasks[i].Dependencies, taskDependency.Id)
			}
		} else {
			tasks[i].Dependencies = []string{} // Ensure it's initialized as an empty slice
		}

		taskFilePath, err := utils.BuildTaskPath(rootFolder, task.EntityPath, task.Name, task.Extension)
		if err != nil {
			return tasks, err
		}
		tasks[i].FilePath = taskFilePath
		tasks[i].Checkpoints = taskCheckpoints[task.Id]

		// fileStatus := "normal"
		fileStatus, err := GetTaskFileStatus(&tasks[i], taskCheckpoints[task.Id])
		if err != nil {
			return tasks, err
		}
		tasks[i].FileStatus = fileStatus
	}

	return tasks, nil
}

func getAllDependencies(taskId string, depth int, allTaskDependencies []models.TaskDependency) []string {
	if depth > 20 {
		return []string{}
	}
	dependencies := []string{}
	for _, taskDependency := range allTaskDependencies {
		if taskDependency.TaskId == taskId {
			dependencies = append(dependencies, taskDependency.DependencyId)
			dependencies = append(dependencies, getAllDependencies(taskDependency.DependencyId, depth+1, allTaskDependencies)...)
		}
	}
	return dependencies
}

func getAllEntityTasks(entityId string, allTasks []models.Task, allEntities []models.Entity) []string {
	dependencies := []string{}
	for _, task := range allTasks {
		if task.EntityId == entityId {
			dependencies = append(dependencies, task.Id)
		}
	}
	for _, entity := range allEntities {
		if entity.ParentId == entityId {
			dependencies = append(dependencies, getAllEntityTasks(entity.Id, allTasks, allEntities)...)
		}
	}
	return dependencies
}

func GetUserTasks(tx *sqlx.Tx, userId string) ([]models.Task, error) {
	// Get all tasks assigned to the user
	assignedTaskIds := []string{}
	query := "SELECT id FROM task WHERE assignee_id = ? AND trashed = 0"
	err := tx.Select(&assignedTaskIds, query, userId)
	if err != nil {
		return nil, err
	}

	// Get all tasks id
	allTasksId := []string{}
	query = `SELECT id FROM task WHERE trashed = 0`
	err = tx.Select(&allTasksId, query)
	if err != nil {
		return nil, err
	}

	allTaskInfo := []models.Task{}
	query = `SELECT id, entity_id FROM task WHERE trashed = 0`
	err = tx.Select(&allTaskInfo, query)
	if err != nil {
		return nil, err
	}

	allEntityInfo := []models.Entity{}
	query = `SELECT id, parent_id FROM entity WHERE trashed = 0`
	err = tx.Select(&allEntityInfo, query)
	if err != nil {
		return nil, err
	}

	// All task dependencies records
	allTaskDependencies := []models.TaskDependency{}
	query = `SELECT task_id, dependency_id FROM task_dependency`
	err = tx.Select(&allTaskDependencies, query)
	if err != nil {
		return nil, err
	}

	// All entity dependencies records
	allEntityDependencies := []models.EntityDependency{}
	query = `SELECT task_id, dependency_id FROM entity_dependency`
	err = tx.Select(&allEntityDependencies, query)
	if err != nil {
		return nil, err
	}

	// Get all assigned entities for the user
	assignedEntitysIds := []string{}
	query = "select entity_id from entity_assignee where assignee_id = ?"
	err = tx.Select(&assignedEntitysIds, query, userId)
	if err != nil {
		return nil, err
	}

	libraryEntities := []string{}
	query = "select id from entity where is_library = 1"
	err = tx.Select(&libraryEntities, query)
	if err != nil {
		return nil, err
	}

	// process all user tasks and their dependencies recursively
	dependencies := map[string]struct{}{}
	for _, taskId := range assignedTaskIds {
		dependencies[taskId] = struct{}{}
		for _, dependency := range getAllDependencies(taskId, 0, allTaskDependencies) {
			dependencies[dependency] = struct{}{}
		}
		for _, entityDependency := range allEntityDependencies {
			if entityDependency.TaskId == taskId {
				entityTasks := getAllEntityTasks(entityDependency.DependencyId, allTaskInfo, allEntityInfo)
				for _, entityTask := range entityTasks {
					dependencies[entityTask] = struct{}{}
				}
			}
		}
	}

	for _, entityId := range assignedEntitysIds {
		entityTasks := getAllEntityTasks(entityId, allTaskInfo, allEntityInfo)
		for _, entityTask := range entityTasks {
			dependencies[entityTask] = struct{}{}
		}
	}

	for _, entityId := range libraryEntities {
		entityTasks := getAllEntityTasks(entityId, allTaskInfo, allEntityInfo)
		for _, entityTask := range entityTasks {
			dependencies[entityTask] = struct{}{}
		}
	}

	// Convert map keys to slice
	taskIds := make([]string, 0, len(dependencies))
	for id := range dependencies {
		taskIds = append(taskIds, id)
	}

	// Get all tasks with the collected IDs
	query = `
		SELECT * FROM full_task WHERE id IN (SELECT value FROM json_each(?)) AND trashed = 0
	`
	tasks := []models.Task{}
	jsonTaskIds, err := json.Marshal(taskIds)
	if err != nil {
		return nil, err
	}
	err = tx.Select(&tasks, query, jsonTaskIds)
	if err != nil {
		return nil, err
	}

	statuses, err := GetStatuses(tx)
	if err != nil {
		return tasks, err
	}
	statusesMap := map[string]models.Status{}
	for _, status := range statuses {
		statusesMap[status.Id] = status
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
		return tasks, err
	}
	for i, task := range tasks {
		status := statusesMap[task.StatusId]
		tasks[i].StatusShortName = status.ShortName
		tasks[i].Status = status

		if tasks[i].TagsRaw != "[]" {
			taskTags := []TaskTags{}
			err = json.Unmarshal([]byte(task.TagsRaw), &taskTags)
			if err != nil {
				return tasks, err
			}
			for _, taskTag := range taskTags {
				tasks[i].Tags = append(tasks[i].Tags, taskTag.Name)
			}
		} else {
			tasks[i].Tags = []string{} // Ensure it's initialized as an empty slice
		}

		if task.EntityDependenciesRaw != "[]" {
			entityDependencies := []Dependency{}
			err = json.Unmarshal([]byte(task.EntityDependenciesRaw), &entityDependencies)
			if err != nil {
				return tasks, err
			}
			for _, entityDependency := range entityDependencies {
				tasks[i].EntityDependencies = append(tasks[i].EntityDependencies, entityDependency.Id)
			}
		} else {
			tasks[i].EntityDependencies = []string{} // Ensure it's initialized as an empty slice
		}
		if task.DependenciesRaw != "[]" {
			taskDependencies := []Dependency{}
			err = json.Unmarshal([]byte(task.DependenciesRaw), &taskDependencies)
			if err != nil {
				return tasks, err
			}
			for _, taskDependency := range taskDependencies {
				tasks[i].Dependencies = append(tasks[i].Dependencies, taskDependency.Id)
			}
		} else {
			tasks[i].Dependencies = []string{} // Ensure it's initialized as an empty slice
		}

		taskFilePath, err := utils.BuildTaskPath(rootFolder, task.EntityPath, task.Name, task.Extension)
		if err != nil {
			return tasks, err
		}
		tasks[i].FilePath = taskFilePath
		tasks[i].Checkpoints = taskCheckpoints[task.Id]

		// fileStatus := "normal"
		fileStatus, err := GetTaskFileStatus(&tasks[i], taskCheckpoints[task.Id])
		if err != nil {
			return tasks, err
		}
		tasks[i].FileStatus = fileStatus
	}

	return tasks, nil
}

// This function is meant for the get user entities to process collections proper
func GetUserTasksMinimal(tx *sqlx.Tx, userId string) ([]models.Task, error) {
	// Get all tasks assigned to the user
	assignedTaskIds := []string{}
	query := "SELECT id FROM task WHERE assignee_id = ? AND trashed = 0"
	err := tx.Select(&assignedTaskIds, query, userId)
	if err != nil {
		return nil, err
	}

	// Get all tasks id
	allTasksId := []string{}
	query = `SELECT id FROM task WHERE trashed = 0`
	err = tx.Select(&allTasksId, query)
	if err != nil {
		return nil, err
	}

	allTaskInfo := []models.Task{}
	query = `SELECT id, entity_id FROM task WHERE trashed = 0`
	err = tx.Select(&allTaskInfo, query)
	if err != nil {
		return nil, err
	}

	allEntityInfo := []models.Entity{}
	query = `SELECT id, parent_id FROM entity WHERE trashed = 0`
	err = tx.Select(&allEntityInfo, query)
	if err != nil {
		return nil, err
	}

	// All task dependencies records
	allTaskDependencies := []models.TaskDependency{}
	query = `SELECT task_id, dependency_id FROM task_dependency`
	err = tx.Select(&allTaskDependencies, query)
	if err != nil {
		return nil, err
	}

	// All entity dependencies records
	allEntityDependencies := []models.EntityDependency{}
	query = `SELECT task_id, dependency_id FROM entity_dependency`
	err = tx.Select(&allEntityDependencies, query)
	if err != nil {
		return nil, err
	}

	// Get all assigned entities for the user
	assignedEntitysIds := []string{}
	query = "select entity_id from entity_assignee where assignee_id = ?"
	err = tx.Select(&assignedEntitysIds, query, userId)
	if err != nil {
		return nil, err
	}

	libraryEntities := []string{}
	query = "select id from entity where is_library = 1"
	err = tx.Select(&libraryEntities, query, userId)
	if err != nil {
		return nil, err
	}

	// process all user tasks and their dependencies recursively
	dependencies := map[string]struct{}{}
	for _, taskId := range assignedTaskIds {
		dependencies[taskId] = struct{}{}
		for _, dependency := range getAllDependencies(taskId, 0, allTaskDependencies) {
			dependencies[dependency] = struct{}{}
		}
		for _, entityDependency := range allEntityDependencies {
			if entityDependency.TaskId == taskId {
				entityTasks := getAllEntityTasks(entityDependency.DependencyId, allTaskInfo, allEntityInfo)
				for _, entityTask := range entityTasks {
					dependencies[entityTask] = struct{}{}
				}
			}
		}

		for _, entityId := range assignedEntitysIds {
			entityTasks := getAllEntityTasks(entityId, allTaskInfo, allEntityInfo)
			for _, entityTask := range entityTasks {
				dependencies[entityTask] = struct{}{}
			}
		}

		for _, entityId := range libraryEntities {
			entityTasks := getAllEntityTasks(entityId, allTaskInfo, allEntityInfo)
			for _, entityTask := range entityTasks {
				dependencies[entityTask] = struct{}{}
			}
		}
	}

	// Convert map keys to slice
	taskIds := make([]string, 0, len(dependencies))
	for id := range dependencies {
		taskIds = append(taskIds, id)
	}

	// Get all tasks with the collected IDs
	query = `
		SELECT id, entity_id FROM task WHERE id IN (SELECT value FROM json_each(?)) AND trashed = 0
	`
	tasks := []models.Task{}
	jsonTaskIds, err := json.Marshal(taskIds)
	if err != nil {
		return nil, err
	}
	err = tx.Select(&tasks, query, jsonTaskIds)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func GetDeletedTasks(tx *sqlx.Tx) ([]models.Task, error) {
	tasks := []models.Task{}

	entities, err := GetEntities(tx, true)
	if err != nil {
		return tasks, err
	}
	entitiesMap := map[string]models.Entity{}
	for _, entity := range entities {
		entitiesMap[entity.Id] = entity
	}

	statuses, err := GetStatuses(tx)
	if err != nil {
		return tasks, err
	}
	statusesMap := map[string]models.Status{}
	for _, status := range statuses {
		statusesMap[status.Id] = status
	}

	tags, err := GetTags(tx)
	if err != nil {
		return tasks, err
	}
	tagsMap := map[string]models.Tag{}
	for _, tag := range tags {
		tagsMap[tag.Id] = tag
	}

	query := "SELECT task_id, GROUP_CONCAT(tag_id) FROM task_tag GROUP BY task_id"
	tasksTags := map[string][]models.Tag{}
	tx.Select(&tasksTags, query)
	for taskId, tagIds := range tasksTags {
		tasksTags[taskId] = []models.Tag{}
		for _, tagId := range tagIds {
			tasksTags[taskId] = append(tasksTags[taskId], tagsMap[tagId.Id])
		}
	}

	conditions := map[string]interface{}{
		"trashed": 1,
	}
	err = base_service.GetAllBy(tx, "task", conditions, &tasks)
	if err != nil {
		return tasks, err
	}
	// tx.Select(&tasks, "SELECT * FROM task WHERE trashed = 0")
	newTasks := []models.Task{}
	//TODO Investigate why for loop does not update data
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return tasks, err
	}
	for _, task := range tasks {
		entity := entitiesMap[task.EntityId]
		task.EntityName = entity.Name

		status := statusesMap[task.StatusId]
		task.StatusShortName = status.ShortName

		taskTags := tasksTags[task.Id]
		task.Tags = []string{}
		for _, tag := range taskTags {
			task.Tags = append(task.Tags, tag.Name)
		}

		taskFilePath, err := utils.BuildTaskPath(rootFolder, task.EntityPath, task.Name, task.Extension)
		if err != nil {
			return tasks, err
		}
		task.FilePath = taskFilePath
		newTasks = append(newTasks, task)
	}

	return newTasks, nil
}

func GetTasksByEntityId(tx *sqlx.Tx, entityId string) ([]models.Task, error) {
	tasks := []models.Task{}

	statuses, err := GetStatuses(tx)
	if err != nil {
		return tasks, err
	}
	statusesMap := map[string]models.Status{}
	for _, status := range statuses {
		statusesMap[status.Id] = status
	}

	tags, err := GetTags(tx)
	if err != nil {
		return tasks, err
	}
	tagsMap := map[string]models.Tag{}
	for _, tag := range tags {
		tagsMap[tag.Id] = tag
	}

	query := "SELECT * FROM task_tag"
	tasksTags := []models.TaskTag{}
	tx.Select(&tasksTags, query)

	taskTags := map[string][]models.Tag{}
	for _, taskTag := range tasksTags {
		taskTags[taskTag.TaskId] = append(taskTags[taskTag.TaskId], tagsMap[taskTag.TagId])
	}

	checkpointQuery := "SELECT * FROM task_checkpoint WHERE trashed = 0 ORDER BY created_at DESC"
	tasksCheckpoints := []models.Checkpoint{}
	tx.Select(&tasksCheckpoints, checkpointQuery)

	taskCheckpoints := map[string][]models.Checkpoint{}
	for _, taskCheckpoint := range tasksCheckpoints {
		taskCheckpoints[taskCheckpoint.TaskId] = append(taskCheckpoints[taskCheckpoint.TaskId], taskCheckpoint)
	}

	conditions := map[string]interface{}{
		"trashed":   0,
		"entity_id": entityId,
	}
	err = base_service.GetAllBy(tx, "full_task", conditions, &tasks)
	if err != nil {
		return tasks, err
	}
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return tasks, err
	}
	for i, task := range tasks {
		status := statusesMap[task.StatusId]
		tasks[i].StatusShortName = status.ShortName

		taskTags := taskTags[task.Id]
		tasks[i].Tags = []string{}
		for _, tag := range taskTags {
			tasks[i].Tags = append(tasks[i].Tags, tag.Name)
		}
		taskFilePath, err := utils.BuildTaskPath(rootFolder, task.EntityPath, task.Name, task.Extension)
		if err != nil {
			return tasks, err
		}
		tasks[i].FilePath = taskFilePath

		// fileStatus := "normal"
		fileStatus, err := GetTaskFileStatus(&tasks[i], taskCheckpoints[task.Id])
		if err != nil {
			return tasks, err
		}
		tasks[i].FileStatus = fileStatus
	}

	return tasks, nil
}

func GetTasksByTagId(tx *sqlx.Tx, tagId string) []models.Task {
	tasks := []models.Task{}
	tx.Select(&tasks, "SELECT * FROM task WHERE id IN (SELECT task_id FROM task_tag WHERE tag_id = ?)", tagId)
	return tasks
}

func GetTasksByTagName(tx *sqlx.Tx, tagName string) []models.Task {
	tasks := []models.Task{}
	tx.Select(&tasks, "SELECT * FROM task WHERE id IN (SELECT task_id FROM task_tag WHERE tag_id IN (SELECT id FROM tag WHERE name = ?))", tagName)
	return tasks
}

func DeleteTask(tx *sqlx.Tx, taskId string, removeFromDir bool, recycle bool) error {
	task, err := GetTask(tx, taskId)
	if err != nil {
		return err
	}
	if recycle {
		err = base_service.MarkAsDeleted(tx, "task", taskId)
		if err != nil {
			return err
		}
		err = base_service.UpdateMtime(tx, "task", taskId, utils.GetEpochTime())
		if err != nil {
			return err
		}
	} else {
		err = DeleteCheckpoints(tx, taskId)
		if err != nil {
			return err
		}
		RemoveAllTagsFromTask(tx, taskId)
		err = base_service.Delete(tx, "task", taskId)
		if err != nil {
			return err
		}
	}
	if removeFromDir {
		err := os.RemoveAll(task.FilePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteEntityTasks(tx *sqlx.Tx, entityId string, removeFromDir bool) error {
	tasks, err := GetTasksByEntityId(tx, entityId)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		err := DeleteTask(tx, task.Id, removeFromDir, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateStatus(tx *sqlx.Tx, taskId string, statusId string) error {
	params := map[string]interface{}{
		"status_id": statusId,
	}
	err := base_service.Update(tx, "task", taskId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "task", taskId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func ChangeEntity(tx *sqlx.Tx, taskId string, entityId string) error {
	oldTask, err := GetTask(tx, taskId)
	if err != nil {
		return err
	}

	params := map[string]interface{}{
		"entity_id": entityId,
	}
	err = base_service.Update(tx, "task", taskId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "task", taskId, utils.GetEpochTime())
	if err != nil {
		return err
	}

	task, err := GetTask(tx, taskId)
	if err != nil {
		return err
	}

	if oldTask.FilePath != task.FilePath && utils.FileExists(oldTask.FilePath) {
		newFilePath := task.FilePath

		newTaskFolder := filepath.Dir(newFilePath)

		err := os.MkdirAll(newTaskFolder, os.ModePerm)
		if err != nil {
			return err
		}

		err = os.Rename(oldTask.FilePath, task.FilePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func ChangeTaskType(tx *sqlx.Tx, taskId string, taskTypeId string) error {
	params := map[string]any{
		"task_type_id": taskTypeId,
	}
	err := base_service.Update(tx, "task", taskId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "task", taskId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func ToggleIsTask(tx *sqlx.Tx, taskId string, isTask bool) error {
	params := map[string]any{
		"is_resource": !isTask,
	}
	err := base_service.Update(tx, "task", taskId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "task", taskId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func UpdateTask(tx *sqlx.Tx, taskId string, name, taskTypeId string, isResource bool, pointer string, tags []string) (models.Task, error) {
	name = strings.TrimSpace(name)
	oldTask, err := GetTask(tx, taskId)
	if err != nil {
		return models.Task{}, err
	}

	newTaskName := oldTask.Name

	if (name != "") && (name != oldTask.Name) {
		newTaskName = name
	}

	params := map[string]interface{}{
		"name":         newTaskName,
		"pointer":      pointer,
		"is_resource":  isResource,
		"task_type_id": taskTypeId,
	}
	err = base_service.Update(tx, "task", taskId, params)
	if err != nil {
		return models.Task{}, err
	}
	err = base_service.UpdateMtime(tx, "task", taskId, utils.GetEpochTime())
	if err != nil {
		return models.Task{}, err
	}
	task, err := GetTask(tx, taskId)
	if err != nil {
		return models.Task{}, err
	}

	err = RemoveAllTagsFromTask(tx, taskId)
	if err != nil {
		return models.Task{}, err
	}
	if task.IsLink {
		if !utils.NonCaseSensitiveContains(tags, "link") {
			tags = append(tags, "link")
		}
	}
	if !task.IsLink && task.Pointer != "" {
		if !utils.NonCaseSensitiveContains(tags, "tracked") {
			tags = append(tags, "tracked")
		}
	}

	task.Tags = []string{}
	if len(tags) > 0 {
		for _, tag := range tags {
			err = AddTagToTask(tx, task.Id, tag)
			if err != nil {
				return models.Task{}, err
			}
			task.Tags = append(task.Tags, tag)
		}
	}

	if oldTask.FilePath != task.FilePath && utils.FileExists(oldTask.FilePath) {
		newFilePath := task.FilePath

		newTaskFolder := filepath.Dir(newFilePath)

		err := os.MkdirAll(newTaskFolder, os.ModePerm)
		if err != nil {
			return models.Task{}, err
		}

		err = os.Rename(oldTask.FilePath, task.FilePath)
		if err != nil {
			return models.Task{}, err
		}
	}

	return task, nil
}

func UpdateSyncTask(tx *sqlx.Tx, taskId string, name, entityId, taskTypeId, assigneeId, assignerId, statusId, previewId string, isResource, isLink bool, pointer string, tags []string) error {
	name = strings.TrimSpace(name)
	oldTask, err := GetTask(tx, taskId)
	if err != nil {
		return err
	}

	newTaskName := oldTask.Name

	if (name != "") && (name != oldTask.Name) {
		newTaskName = name
	}

	params := map[string]any{
		"name":         newTaskName,
		"is_resource":  isResource,
		"is_link":      isLink,
		"pointer":      pointer,
		"task_type_id": taskTypeId,
		"assignee_id":  assigneeId,
		"assigner_id":  assignerId,
		"entity_id":    entityId,
		"status_id":    statusId,
		"preview_id":   previewId,
	}
	err = base_service.Update(tx, "task", taskId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "task", taskId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	task, err := GetSimpleTask(tx, taskId)
	if err != nil {
		return err
	}

	err = RemoveAllTagsFromTask(tx, taskId)
	if err != nil {
		return err
	}
	if task.IsLink {
		if !utils.NonCaseSensitiveContains(tags, "link") {
			tags = append(tags, "link")
		}
	}
	if !task.IsLink && task.Pointer != "" {
		if !utils.NonCaseSensitiveContains(tags, "tracked") {
			tags = append(tags, "tracked")
		}
	}

	if len(tags) > 0 {
		for _, tag := range tags {
			err = AddTagToTask(tx, task.Id, tag)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func RenameTask(tx *sqlx.Tx, taskId, name string) (models.Task, error) {
	name = strings.TrimSpace(name)
	oldTask, err := GetTask(tx, taskId)
	if err != nil {
		return models.Task{}, err
	}

	newTaskName := oldTask.Name

	if (name != "") && (name != oldTask.Name) {
		newTaskName = name
	}

	err = base_service.Rename(tx, "task", taskId, newTaskName)
	if err != nil {
		return models.Task{}, err
	}
	err = base_service.UpdateMtime(tx, "task", taskId, utils.GetEpochTime())
	if err != nil {
		return models.Task{}, err
	}
	task, err := GetTask(tx, taskId)
	if err != nil {
		return models.Task{}, err
	}

	if oldTask.FilePath != task.FilePath && utils.FileExists(oldTask.FilePath) {
		newFilePath := task.FilePath

		newTaskFolder := filepath.Dir(newFilePath)

		err := os.MkdirAll(newTaskFolder, os.ModePerm)
		if err != nil {
			return models.Task{}, err
		}

		err = os.Rename(oldTask.FilePath, task.FilePath)
		if err != nil {
			return models.Task{}, err
		}
	}

	return task, nil
}

func ToggleIsResource(tx *sqlx.Tx, taskId string, isResource bool) error {
	params := map[string]interface{}{
		"is_resource": isResource,
	}
	err := base_service.Update(tx, "task", taskId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "task", taskId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func ToggleIsResourceM(tx *sqlx.Tx, taskIds []string, isResource bool) error {
	for _, taskId := range taskIds {
		err := ToggleIsResource(tx, taskId, isResource)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateAssignation(tx *sqlx.Tx, taskId string, assigneeId, assignerId string) error {
	params := map[string]interface{}{
		"assignee_id": assigneeId,
		"assigner_id": assignerId,
	}
	err := base_service.Update(tx, "task", taskId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "task", taskId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func AssignTask(tx *sqlx.Tx, taskId string, userId string) error {
	params := map[string]interface{}{
		"assignee_id": userId,
	}
	err := base_service.Update(tx, "task", taskId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "task", taskId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func UnAssignTask(tx *sqlx.Tx, taskId string) error {
	params := map[string]interface{}{
		"assignee_id": "",
	}
	err := base_service.Update(tx, "task", taskId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "task", taskId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func UpdateTaskPointer(tx *sqlx.Tx, taskId string, pointer string) (models.Task, error) {
	if !utils.FileExists(pointer) {
		return models.Task{}, errors.New("invalid pointer, path does not exist")
	}
	params := map[string]interface{}{
		"pointer": pointer,
	}
	err := base_service.Update(tx, "task", taskId, params)
	if err != nil {
		return models.Task{}, err
	}
	err = base_service.UpdateMtime(tx, "task", taskId, utils.GetEpochTime())
	if err != nil {
		return models.Task{}, err
	}
	task, err := GetTask(tx, taskId)
	if err != nil {
		return models.Task{}, err
	}

	return task, nil
}

// GetAssetTasks gets all tasks where is_resource is false with minimal fields for UI display
func GetAssetTasks(tx *sqlx.Tx) ([]models.Task, error) {
	tasks := []models.Task{}
	queryWhereClause := "WHERE is_resource = 0 AND trashed = 0"

	query := fmt.Sprintf(`
		SELECT 
			id,
			name,
			task_type_icon,
			assignee_id,
			preview,
			status_id,
			extension
		FROM full_task %s ORDER BY name`, queryWhereClause)

	err := tx.Select(&tasks, query)
	if err != nil {
		return tasks, err
	}
	return tasks, nil
}
