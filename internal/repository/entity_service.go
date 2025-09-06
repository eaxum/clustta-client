package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"clustta/internal/base_service"
	"clustta/internal/error_service"
	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

func TopologicalSortOld(entities []models.Entity) ([]models.Entity, error) {

	// Create maps for easy lookup
	idToEntity := make(map[string]models.Entity)
	fmt.Println("idToEntity: ", idToEntity)
	dependencyCount := make(map[string]int)
	fmt.Println("Dependency count: ", dependencyCount)
	children := make(map[string][]string)
	fmt.Println("Children: ", children)

	// Build maps
	for _, e := range entities {
		idToEntity[e.Id] = e
		if e.ParentId != "" {
			dependencyCount[e.Id]++
			children[e.ParentId] = append(children[e.ParentId], e.Id)
		}
	}

	// Start with nodes that have no dependencies (root nodes)
	var queue []string
	for _, e := range entities {
		if dependencyCount[e.Id] == 0 {
			queue = append(queue, e.Id)
		}
	}

	var result []models.Entity
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		result = append(result, idToEntity[id])

		for _, childId := range children[id] {
			dependencyCount[childId]--
			if dependencyCount[childId] == 0 {
				queue = append(queue, childId)
			}
		}
	}
	fmt.Println("Entities: ", entities)
	fmt.Println("Result: ", result)

	if len(result) != len(entities) {
		fmt.Println("cycle detected or missing parents: ", len(result)-len(entities))
		// return nil, fmt.Errorf("cycle detected or missing parents")
	}

	return result, nil
}

func TopologicalSort(entities []models.Entity) ([]models.Entity, error) {
	idToEntity := make(map[string]models.Entity)
	dependencyCount := make(map[string]int)
	children := make(map[string][]string)

	for _, e := range entities {
		idToEntity[e.Id] = e
		if e.ParentId != "" {
			dependencyCount[e.Id]++
			children[e.ParentId] = append(children[e.ParentId], e.Id)
		}
	}

	entityIds := make(map[string]struct{})
	for _, e := range entities {
		entityIds[e.Id] = struct{}{}
	}

	var queue []string
	for _, e := range entities {
		_, parentExists := entityIds[e.ParentId]
		if dependencyCount[e.Id] == 0 || (e.ParentId != "" && !parentExists) {
			queue = append(queue, e.Id)
		}
	}

	processed := make(map[string]bool)
	var result []models.Entity
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		result = append(result, idToEntity[id])
		processed[id] = true

		for _, childId := range children[id] {
			dependencyCount[childId]--
			if dependencyCount[childId] == 0 {
				queue = append(queue, childId)
			}
		}
	}

	for _, e := range entities {
		if !processed[e.Id] {
			result = append(result, e)
		}
	}

	if len(result) != len(entities) {
		fmt.Println("cycle detected or missing parents: ", len(result)-len(entities))
		// return nil, fmt.Errorf("cycle detected or missing parents")
	}

	return result, nil
}

func CreateEntityFast(
	tx *sqlx.Tx, id string, name, description, entity_type_id, parent_id, previewId string, isLibrary bool,
) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("entity name cannot be empty")
	}

	// if parent_id != "" {
	// 	parent := models.Entity{}
	// 	err := base_service.Get(tx, "entity", parent_id, &parent)
	// 	if err != nil {
	// 		return models.Entity{}, errors.New("parent entity not found")
	// 	}
	// }

	params := map[string]any{
		"id":             id,
		"created_at":     utils.GetCurrentTime(),
		"name":           name,
		"description":    description,
		"entity_type_id": entity_type_id,
		"parent_id":      parent_id,
		"preview_id":     previewId,
		"is_library":     isLibrary,
	}
	err := base_service.Create(tx, "entity", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return error_service.ErrEntityExists
			}
		}
		return err
	}
	return nil
}

func CreateEntity(
	tx *sqlx.Tx, id string, name, description, entity_type_id, parent_id, previewId string, isLibrary bool,
) (models.Entity, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.Entity{}, errors.New("entity name cannot be empty")
	}

	if parent_id != "" {
		parent := models.Entity{}
		err := base_service.Get(tx, "entity", parent_id, &parent)
		if err != nil {
			return models.Entity{}, errors.New("parent entity not found")
		}
	}

	conditions := map[string]any{
		"parent_id": parent_id,
		"name":      name,
	}
	entity := models.Entity{}
	err := base_service.GetBy(tx, "entity", conditions, &entity)
	if err == nil {
		if entity.Trashed {
			return models.Entity{}, error_service.ErrEntityExistsInTrash
		} else {
			return models.Entity{}, error_service.ErrEntityExists
		}
	}

	params := map[string]any{
		"id":             id,
		"created_at":     utils.GetCurrentTime(),
		"name":           name,
		"description":    description,
		"entity_type_id": entity_type_id,
		"parent_id":      parent_id,
		"preview_id":     previewId,
		"is_library":     isLibrary,
	}
	err = base_service.Create(tx, "entity", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return models.Entity{}, error_service.ErrEntityExists
			}
		}
		return models.Entity{}, err
	}
	entity, err = GetEntityByName(tx, name, parent_id)
	if err != nil {
		return models.Entity{}, err
	}
	return entity, nil
}

func AddEntity(
	tx *sqlx.Tx, id string, name, description, entity_type_id, parent_id, previewId string, isLibrary bool,
) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("entity name cannot be empty")
	}

	// if parent_id != "" {
	// 	parent := models.Entity{}
	// 	err := base_service.Get(tx, "entity", parent_id, &parent)
	// 	if err != nil {
	// 		return models.Entity{}, errors.New("parent entity not found")
	// 	}
	// }

	params := map[string]any{
		"id":             id,
		"created_at":     utils.GetCurrentTime(),
		"name":           name,
		"description":    description,
		"entity_type_id": entity_type_id,
		"parent_id":      parent_id,
		"preview_id":     previewId,
		"is_library":     isLibrary,
	}
	err := base_service.Create(tx, "entity", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return error_service.ErrEntityExists
			}
		}
		return err
	}
	return nil
}

func GetSimpleEntities(tx *sqlx.Tx) ([]models.Entity, error) {
	entities := []models.Entity{}

	query := "SELECT * FROM entity"

	err := tx.Select(&entities, query)
	if err != nil && err == sql.ErrNoRows {
		return []models.Entity{}, nil
	} else if err != nil {
		return []models.Entity{}, err
	}
	fmt.Println("SimpleEntities: ", entities)
	return entities, nil
}
func GetEntity(tx *sqlx.Tx, id string) (models.Entity, error) {
	entity := models.Entity{}

	query := "SELECT * FROM full_entity WHERE id = ?"

	err := tx.Get(&entity, query, id)
	if err != nil && err == sql.ErrNoRows {
		return models.Entity{}, error_service.ErrEntityNotFound
	} else if err != nil {
		return models.Entity{}, err
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return entity, err
	}
	entityFilePath, err := utils.BuildEntityPath(rootFolder, entity.EntityPath)
	if err != nil {
		return entity, err
	}
	entity.FilePath = entityFilePath

	if entity.AssigneeIdsRaw != "[]" {
		assigneeIds := []string{}
		err = json.Unmarshal([]byte(entity.AssigneeIdsRaw), &assigneeIds)
		if err != nil {
			return entity, err
		}
		entity.AssigneeIds = assigneeIds
	} else {
		entity.AssigneeIds = []string{} // Ensure it's initialized as an empty slice
	}
	fmt.Println("Final entity: ", entity.ParentId)
	return entity, nil
}

func GetEntityChildren(tx *sqlx.Tx, id string) ([]models.Entity, error) {
	entities := []models.Entity{}

	query := "SELECT * FROM full_entity WHERE parent_id = ? AND trashed = 0 ORDER BY name"

	err := tx.Select(&entities, query, id)
	if err != nil && err == sql.ErrNoRows {
		return entities, nil
	} else if err != nil {
		return []models.Entity{}, err
	}
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return entities, err
	}
	for i, entity := range entities {
		entityFilePath, err := utils.BuildEntityPath(rootFolder, entity.EntityPath)
		if err != nil {
			return entities, err
		}
		entities[i].FilePath = entityFilePath

		if entity.AssigneeIdsRaw != "[]" {
			assigneeIds := []string{}
			err = json.Unmarshal([]byte(entity.AssigneeIdsRaw), &assigneeIds)
			if err != nil {
				return entities, err
			}
			entities[i].AssigneeIds = assigneeIds
		} else {
			entities[i].AssigneeIds = []string{} // Ensure it's initialized as an empty slice
		}
	}
	return entities, nil
}

func GetEntityTasks(tx *sqlx.Tx, id string) ([]models.Task, error) {
	tasks := []models.Task{}

	query := "SELECT * FROM full_task WHERE entity_id = ? AND trashed = 0 ORDER BY name"

	err := tx.Select(&tasks, query, id)
	if err != nil && err == sql.ErrNoRows {
		return tasks, nil
	} else if err != nil {
		return []models.Task{}, err
	}

	statuses, err := GetStatuses(tx)
	if err != nil {
		return tasks, err
	}
	statusesMap := map[string]models.Status{}
	for _, status := range statuses {
		statusesMap[status.Id] = status
	}

	qoutedTaskIds := make([]string, len(tasks))
	for i, task := range tasks {
		qoutedTaskIds[i] = fmt.Sprintf("\"%s\"", task.Id)
	}
	tasksCheckpointQuery := fmt.Sprintf("SELECT * FROM task_checkpoint WHERE task_id IN (%s) AND trashed = 0 ORDER BY created_at DESC", strings.Join(qoutedTaskIds, ","))

	// checkpointQuery := "SELECT * FROM task_checkpoint WHERE trashed = 0 ORDER BY created_at DESC"
	tasksCheckpoints := []models.Checkpoint{}
	tx.Select(&tasksCheckpoints, tasksCheckpointQuery)

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

		fileStatus, err := GetTaskFileStatus(&tasks[i], taskCheckpoints[task.Id])
		if err != nil {
			return tasks, err
		}
		tasks[i].FileStatus = fileStatus
	}

	return tasks, nil
}

func GetEntityByPath(tx *sqlx.Tx, entityPath string) (models.Entity, error) {
	entity := models.Entity{}

	query := "SELECT * FROM full_entity WHERE entity_path = ?"

	err := tx.Get(&entity, query, entityPath)
	if err != nil && err == sql.ErrNoRows {
		return models.Entity{}, error_service.ErrEntityNotFound
	} else if err != nil {
		return models.Entity{}, err
	}
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return entity, err
	}
	entityFilePath, err := utils.BuildEntityPath(rootFolder, entity.EntityPath)
	if err != nil {
		return entity, err
	}
	entity.FilePath = entityFilePath
	if entity.AssigneeIdsRaw != "[]" {
		assigneeIds := []string{}
		err = json.Unmarshal([]byte(entity.AssigneeIdsRaw), &assigneeIds)
		if err != nil {
			return entity, err
		}
		entity.AssigneeIds = assigneeIds
	} else {
		entity.AssigneeIds = []string{} // Ensure it's initialized as an empty slice
	}
	return entity, nil
}

func GetOrCreateEntity(tx *sqlx.Tx, entityPath string) (models.Entity, []models.Entity, error) {
	parts := strings.Split(entityPath, "/")
	var entityPaths []string
	var current string

	if entityPath == "" {
		return models.Entity{}, []models.Entity{}, nil
	}

	for _, part := range parts {
		if current == "" {
			current = part
		} else {
			current = path.Join(current, part)
		}
		entityPaths = append(entityPaths, current)
	}

	entityType, err := GetEntityTypeByName(tx, "generic")
	if err != nil {
		return models.Entity{}, []models.Entity{}, err
	}

	newEntities := []models.Entity{}
	prevEntity := models.Entity{}
	for _, curentEntityPath := range entityPaths {
		entity, err := GetEntityByPath(tx, curentEntityPath)
		if err == nil {
			prevEntity = entity
			continue
		}
		entityName := filepath.Base(curentEntityPath)
		entity, err = CreateEntity(tx, "", entityName, "", entityType.Id, prevEntity.Id, "", false)
		if err != nil {
			return entity, newEntities, err
		}
		newEntities = append(newEntities, entity)
		prevEntity = entity
	}

	return prevEntity, newEntities, nil
}

func GetEntityByName(tx *sqlx.Tx, name string, parentId string) (models.Entity, error) {
	entity := models.Entity{}
	query := "SELECT * FROM full_entity WHERE name = ? AND parent_id = ?"

	err := tx.Get(&entity, query, name, parentId)
	if err != nil && err == sql.ErrNoRows {
		return models.Entity{}, error_service.ErrEntityNotFound
	} else if err != nil {
		return models.Entity{}, err
	}
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return entity, err
	}
	entityFilePath, err := utils.BuildEntityPath(rootFolder, entity.EntityPath)
	if err != nil {
		return entity, err
	}
	entity.FilePath = entityFilePath
	if entity.AssigneeIdsRaw != "[]" {
		assigneeIds := []string{}
		err = json.Unmarshal([]byte(entity.AssigneeIdsRaw), &assigneeIds)
		if err != nil {
			return entity, err
		}
		entity.AssigneeIds = assigneeIds
	} else {
		entity.AssigneeIds = []string{} // Ensure it's initialized as an empty slice
	}
	return entity, nil
}

func GetEntities(tx *sqlx.Tx, withDeleted bool) ([]models.Entity, error) {
	entities := []models.Entity{}
	queryWhereClause := ""
	if !withDeleted {
		queryWhereClause = "WHERE trashed = 0"
	}
	query := fmt.Sprintf("SELECT * FROM full_entity %s", queryWhereClause)

	err := tx.Select(&entities, query)
	if err != nil {
		return entities, err
	}
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return entities, err
	}
	for i, entity := range entities {
		entityFilePath, err := utils.BuildEntityPath(rootFolder, entity.EntityPath)
		if err != nil {
			return entities, err
		}
		entities[i].FilePath = entityFilePath

		if entity.AssigneeIdsRaw != "[]" {
			assigneeIds := []string{}
			err = json.Unmarshal([]byte(entity.AssigneeIdsRaw), &assigneeIds)
			if err != nil {
				return entities, err
			}
			entities[i].AssigneeIds = assigneeIds
		} else {
			entities[i].AssigneeIds = []string{} // Ensure it's initialized as an empty slice
		}
	}

	return entities, nil
}

func OLDGetUserEntities(tx *sqlx.Tx, userId string) ([]models.Entity, error) {
	entities := []models.Entity{}
	// query := "SELECT * FROM full_entity WHERE trashed = 0"
	query := fmt.Sprintf(`
		WITH RECURSIVE 
		-- First get all tasks and their dependencies
		task_dependencies AS (
			-- Base case: directly assigned tasks
			SELECT 
				id,
				entity_id
			FROM task
			WHERE assignee_id = '%s' 
			AND trashed = 0
			
			UNION
			
			-- Recursive case: task dependencies
			SELECT 
				t.id,
				t.entity_id
			FROM task t
			JOIN task_dependency dep ON t.id = dep.dependency_id
			JOIN task_dependencies td ON dep.task_id = td.id
			WHERE t.trashed = 0
		),
		
		-- Get entity dependencies for all tasks
		entity_dependencies AS (
			-- Base case: direct entity dependencies from tasks
			SELECT 
				ed.dependency_id as id
			FROM entity_dependency ed
			JOIN task_dependencies td ON ed.task_id = td.id
			WHERE ed.dependency_id != ''
			
			UNION
			
			-- Include original task entities
			SELECT DISTINCT entity_id as id
			FROM task_dependencies
			WHERE entity_id != ''
		),
		
		-- Get full entity hierarchy (both parents and children) for all relevant entities
		entity_hierarchy_full AS (
			-- Base case: entities from tasks and entity dependencies
			SELECT 
				e.id,
				e.parent_id,
				e.name,
				0 as level,
				CAST(e.id AS TEXT) as hierarchy_path
			FROM entity e
			LEFT JOIN entity_dependencies ed ON e.id = ed.id
			WHERE e.trashed = 0
			AND (ed.id IS NOT NULL OR e.is_library = 1)
			
			UNION ALL
			
			-- Recursive case upward: get parents
			SELECT 
				e.id,
				e.parent_id,
				e.name,
				ehf.level - 1,
				ehf.hierarchy_path || ',' || e.id
			FROM entity e
			JOIN entity_hierarchy_full ehf ON e.id = ehf.parent_id
			WHERE e.trashed = 0
			AND e.id != ''
			AND e.id NOT IN (
				SELECT value 
				FROM json_each('["' || REPLACE(ehf.hierarchy_path, ',', '","') || '"]')
			)
			
			UNION ALL
			
			-- Recursive case downward: get children
			SELECT 
				e.id,
				e.parent_id,
				e.name,
				ehf.level + 1,
				ehf.hierarchy_path || ',' || e.id
			FROM entity e
			JOIN entity_hierarchy_full ehf ON e.parent_id = ehf.id
			WHERE e.trashed = 0
			AND e.is_library = 1
			AND e.id NOT IN (
				SELECT value 
				FROM json_each('["' || REPLACE(ehf.hierarchy_path, ',', '","') || '"]')
			)
		)
		-- Final select with entity details
		SELECT DISTINCT
			e.*,
			et.name AS entity_type_name,
			et.icon AS entity_type_icon,
			p.preview AS preview,
			COALESCE(eh.entity_path, '') AS entity_path, -- Ensure no NULL values
			CASE 
				WHEN ed.id IS NOT NULL THEN true 
				ELSE false 
			END as is_dependency
		FROM entity_hierarchy_full ehf
		JOIN entity e ON ehf.id = e.id
		JOIN entity_type et ON e.entity_type_id = et.id
		LEFT JOIN preview p ON e.preview_id = p.hash
		LEFT JOIN entity_hierarchy eh ON e.id = eh.id
		LEFT JOIN entity_dependencies ed ON e.id = ed.id
		ORDER BY 
			eh.entity_path;
	`, userId)

	err := tx.Select(&entities, query)
	if err != nil {
		return entities, err
	}
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return entities, err
	}
	for i, entity := range entities {

		entityFilePath, err := utils.BuildEntityPath(rootFolder, entity.EntityPath)
		if err != nil {
			return entities, err
		}
		entities[i].FilePath = entityFilePath
	}

	return entities, nil
}

func getAllEntityChildren(entityId string, allEntities []models.Entity) []string {
	children := []string{}
	for _, entity := range allEntities {
		if entity.ParentId == entityId {
			children = append(children, entity.Id)
			// Recursively get children of the current child
			children = append(children, getAllEntityChildren(entity.Id, allEntities)...)
		}
	}
	return children
}

func getEntityParents(entityId string, allEntities []models.Entity) []string {
	parents := []string{}
	for _, entity := range allEntities {
		if entity.Id == entityId && entity.ParentId != "" {
			parents = append(parents, entity.ParentId)
			// Recursively get parents of the current parent
			parents = append(parents, getEntityParents(entity.ParentId, allEntities)...)
		}
	}
	return parents
}

func getTaskEntities(task models.Task, allEntities []models.Entity) []string {
	parents := []string{}
	parents = append(parents, task.EntityId)
	// Recursively get parent entities
	parents = append(parents, getEntityParents(task.EntityId, allEntities)...)
	return parents
}

func GetUserEntities(tx *sqlx.Tx, userTaskInfos []models.Task, userId string) ([]models.Entity, error) {
	// Get all assigned entities for the user
	assignedEntitysIds := []string{}
	query := "select entity_id from entity_assignee where assignee_id = ?"
	err := tx.Select(&assignedEntitysIds, query, userId)
	if err != nil {
		return nil, err
	}

	libraryEntities := []string{}
	query = "select id from entity where is_library = 1"
	err = tx.Select(&libraryEntities, query, userId)
	if err != nil {
		return nil, err
	}

	allEntityInfo := []models.Entity{}
	query = `SELECT id, parent_id FROM entity WHERE trashed = 0`
	err = tx.Select(&allEntityInfo, query)
	if err != nil {
		return nil, err
	}

	userTaskEntitiesIds := []string{}
	for _, userTask := range userTaskInfos {
		userTaskEntitiesIds = append(userTaskEntitiesIds, getTaskEntities(userTask, allEntityInfo)...)
	}

	//process all user entities
	canModifyEntitiesIds := map[string]struct{}{}
	userEntitiesIds := map[string]struct{}{}
	for _, entityId := range assignedEntitysIds {
		userEntitiesIds[entityId] = struct{}{}
		canModifyEntitiesIds[entityId] = struct{}{}
		for _, parent := range getEntityParents(entityId, allEntityInfo) {
			userEntitiesIds[parent] = struct{}{}
		}
		for _, child := range getAllEntityChildren(entityId, allEntityInfo) {
			userEntitiesIds[child] = struct{}{}
			canModifyEntitiesIds[child] = struct{}{}
		}
	}
	for _, entityId := range libraryEntities {
		userEntitiesIds[entityId] = struct{}{}
		for _, parent := range getEntityParents(entityId, allEntityInfo) {
			userEntitiesIds[parent] = struct{}{}
		}
		for _, child := range getAllEntityChildren(entityId, allEntityInfo) {
			userEntitiesIds[child] = struct{}{}
		}
	}
	for _, entityId := range userTaskEntitiesIds {
		userEntitiesIds[entityId] = struct{}{}
	}

	entitiesIds := make([]string, 0, len(userEntitiesIds))
	for id := range userEntitiesIds {
		entitiesIds = append(entitiesIds, id)
	}

	// Get all entities with the collected IDs
	query = `
		SELECT * FROM full_entity WHERE id IN (SELECT value FROM json_each(?)) AND trashed = 0
	`
	entities := []models.Entity{}
	jsonEntityIds, err := json.Marshal(entitiesIds)
	if err != nil {
		return nil, err
	}
	err = tx.Select(&entities, query, jsonEntityIds)
	if err != nil {
		return nil, err
	}

	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return entities, err
	}
	for i, entity := range entities {
		entityFilePath, err := utils.BuildEntityPath(rootFolder, entity.EntityPath)
		if err != nil {
			return entities, err
		}
		entities[i].FilePath = entityFilePath
		if entity.AssigneeIdsRaw != "[]" {
			assigneeIds := []string{}
			err = json.Unmarshal([]byte(entity.AssigneeIdsRaw), &assigneeIds)
			if err != nil {
				return entities, err
			}
			entities[i].AssigneeIds = assigneeIds
		} else {
			entities[i].AssigneeIds = []string{} // Ensure it's initialized as an empty slice
		}

		// Add CanModify to entity if in the canModifyEntitiesIds map
		if _, exists := canModifyEntitiesIds[entity.Id]; exists {
			entities[i].CanModify = true
		} else {
			entities[i].CanModify = false
		}
	}

	return entities, nil
}

func GetDeletedEntities(tx *sqlx.Tx) ([]models.Entity, error) {
	entities := []models.Entity{}

	conditions := map[string]any{
		"trashed": 1,
	}
	err := base_service.GetAllBy(tx, "full_entity", conditions, &entities)
	if err != nil {
		return entities, err
	}
	// tx.Select(&entities, "SELECT * FROM full_entity WHERE trashed = 0")
	newEntities := []models.Entity{}
	//TODO Investigate why for loop does not update data
	rootFolder, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return entities, err
	}
	for _, entity := range entities {
		entityFilePath, err := utils.BuildEntityPath(rootFolder, entity.EntityPath)
		if err != nil {
			return entities, err
		}
		entity.FilePath = entityFilePath
		newEntities = append(newEntities, entity)
	}

	return newEntities, nil
}

func DeleteEntity(tx *sqlx.Tx, entityId string, removeFromDir bool, recycle bool) error {
	entity, err := GetEntity(tx, entityId)
	if err != nil {
		return err
	}
	if recycle {
		err = base_service.MarkAsDeleted(tx, "entity", entityId)
		if err != nil {
			return err
		}
	} else {
		err = base_service.Delete(tx, "entity", entityId)
		if err != nil {
			return err
		}
	}
	if removeFromDir {
		err := os.RemoveAll(entity.FilePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateEntity(tx *sqlx.Tx, entityId string, name string, tags []string) (models.Entity, error) {
	name = strings.TrimSpace(name)
	oldEntity, err := GetEntity(tx, entityId)
	if err != nil {
		return models.Entity{}, err
	}

	newEntityName := oldEntity.Name

	if (name != "") && (name != oldEntity.Name) {
		newEntityName = name
	}

	params := map[string]any{
		"name": newEntityName,
	}
	err = base_service.Update(tx, "entity", entityId, params)
	if err != nil {
		return models.Entity{}, err
	}
	err = base_service.UpdateMtime(tx, "entity", entityId, utils.GetEpochTime())
	if err != nil {
		return models.Entity{}, err
	}
	entity, err := GetEntity(tx, entityId)
	if err != nil {
		return models.Entity{}, err
	}

	if oldEntity.FilePath != entity.FilePath && utils.DirExists(oldEntity.FilePath) {
		newEntityFolderDir := filepath.Dir(entity.FilePath)
		err := os.MkdirAll(newEntityFolderDir, os.ModePerm)
		if err != nil {
			return models.Entity{}, err
		}

		err = os.Rename(oldEntity.FilePath, entity.FilePath)
		if err != nil {
			return models.Entity{}, err
		}
	}

	return entity, nil
}

func RenameEntity(tx *sqlx.Tx, entityId string, name string) (models.Entity, error) {
	name = strings.TrimSpace(name)
	oldEntity, err := GetEntity(tx, entityId)
	if err != nil {
		return models.Entity{}, err
	}
	newEntityName := oldEntity.Name

	if (name != "") && (name != oldEntity.Name) {
		newEntityName = name
	}

	err = base_service.Rename(tx, "entity", entityId, newEntityName)
	if err != nil {
		return models.Entity{}, err
	}
	err = base_service.UpdateMtime(tx, "entity", entityId, utils.GetEpochTime())
	if err != nil {
		return models.Entity{}, err
	}
	entity, err := GetEntity(tx, entityId)
	if err != nil {
		return models.Entity{}, err
	}

	if oldEntity.FilePath != entity.FilePath && utils.DirExists(oldEntity.FilePath) {

		newEntityFolderDir := filepath.Dir(entity.FilePath)

		err := os.MkdirAll(newEntityFolderDir, os.ModePerm)
		if err != nil {
			return models.Entity{}, err
		}

		err = os.Rename(oldEntity.FilePath, entity.FilePath)
		if err != nil {
			return models.Entity{}, err
		}
	}
	return entity, nil
}

func ChangeParent(tx *sqlx.Tx, entityId string, parentId string) error {
	oldEntity, err := GetEntity(tx, entityId)
	if err != nil {
		return err
	}
	params := map[string]any{
		"parent_id": parentId,
	}
	err = base_service.Update(tx, "entity", entityId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "entity", entityId, utils.GetEpochTime())
	if err != nil {
		return err
	}

	entity, err := GetEntity(tx, entityId)
	if err != nil {
		return err
	}
	if oldEntity.FilePath != entity.FilePath && utils.DirExists(oldEntity.FilePath) {
		newEntityFolderDir := filepath.Dir(entity.FilePath)
		err := os.MkdirAll(newEntityFolderDir, os.ModePerm)
		if err != nil {
			return err
		}
		err = os.Rename(oldEntity.FilePath, entity.FilePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func ChangeEntityType(tx *sqlx.Tx, entityId string, entityTypeId string) error {
	params := map[string]any{
		"entity_type_id": entityTypeId,
	}
	err := base_service.Update(tx, "entity", entityId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "entity", entityId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func ChangeIsLibrary(tx *sqlx.Tx, entityId string, isLibrary bool) error {
	params := map[string]any{
		"is_library": isLibrary,
	}
	err := base_service.Update(tx, "entity", entityId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "entity", entityId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func UpdateEntityPreview(tx *sqlx.Tx, entityId string, previewPath string) error {
	_, err := GetEntity(tx, entityId)
	if err != nil {
		return err
	}
	preview, err := CreatePreview(tx, previewPath)
	if err != nil {
		return err
	}
	params := map[string]any{
		"preview_id": preview.Hash,
	}
	err = base_service.Update(tx, "entity", entityId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "entity", entityId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func AddAssignee(tx *sqlx.Tx, id, entityId, userId string) error {
	params := map[string]any{
		"id":          id,
		"entity_id":   entityId,
		"assignee_id": userId,
	}
	err := base_service.Create(tx, "entity_assignee", params)
	if err != nil {
		return err
	}
	return nil
}

func GetAssignee(tx *sqlx.Tx, id string) (models.EntityAssignee, error) {
	assignee := models.EntityAssignee{}
	err := base_service.Get(tx, "entity_assignee", id, &assignee)
	if err != nil {
		return assignee, err
	}
	return assignee, nil
}

func AssignEntity(tx *sqlx.Tx, entityId, userId string) error {
	params := map[string]any{
		"entity_id":   entityId,
		"assignee_id": userId,
	}
	err := base_service.Create(tx, "entity_assignee", params)
	if err != nil {
		return err
	}
	return nil
}

func UnAssignEntity(tx *sqlx.Tx, entityId, userId string) error {
	conditions := map[string]any{
		"entity_id":   entityId,
		"assignee_id": userId,
	}
	err := base_service.DeleteBy(tx, "entity_assignee", conditions)
	if err != nil {
		return err
	}
	return nil
}
