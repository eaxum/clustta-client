package repository

import (
	"database/sql"
	"errors"
	"strings"

	"clustta/internal/auth_service"
	"clustta/internal/base_service"
	"clustta/internal/error_service"
	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

func CreateWorkflow(tx *sqlx.Tx, id, name string, workflowTasks []models.WorkflowTask, workflowEntities []models.WorkflowEntity, workflowLinks []models.WorkflowLink) (models.Workflow, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.Workflow{}, errors.New("workflow name cannot be empty")
	}

	params := map[string]any{
		"id":   id,
		"name": name,
	}
	err := base_service.Create(tx, "workflow", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return models.Workflow{}, error_service.ErrTaskExists
			}
		}
		return models.Workflow{}, err
	}
	workflow, err := GetWorkflowByName(tx, name)
	if err != nil {
		return models.Workflow{}, err
	}

	for _, workflowTask := range workflowTasks {
		_, err := CreateWorkflowTask(tx, "", workflowTask.Name, workflow.Id, workflowTask.TaskTypeId, workflowTask.IsResource, workflowTask.TemplateId, workflowTask.Pointer, workflowTask.IsLink)
		if err != nil {
			return models.Workflow{}, err
		}
	}
	for _, workflowEntity := range workflowEntities {
		_, err := CreateWorkflowEntity(tx, "", workflowEntity.Name, workflow.Id, workflowEntity.EntityTypeId)
		if err != nil {
			return models.Workflow{}, err
		}
	}

	for _, workflowLink := range workflowLinks {
		err := LinkWorkflow(tx, workflowLink.Name, workflowLink.EntityTypeId, workflow.Id, workflowLink.LinkedWorkflowId)
		if err != nil {
			return models.Workflow{}, err
		}
	}

	return GetWorkflow(tx, workflow.Id)
}
func AddWorkflow(tx *sqlx.Tx, workflowId, name, entityTypeId, parentId string, user auth_service.User) error {
	workflow, err := GetWorkflow(tx, workflowId)
	if err != nil {
		return err
	}
	entity, err := CreateEntity(tx, "", name, "", entityTypeId, parentId, "", false)
	if err != nil {
		return err
	}

	for _, workflowEntity := range workflow.Entities {
		_, err := CreateEntity(tx, "", workflowEntity.Name, "", workflowEntity.EntityTypeId, entity.Id, "", false)
		if err != nil {
			return err
		}
	}
	for _, workflowTask := range workflow.Tasks {
		_, err := CreateTask(tx, "", workflowTask.Name, workflowTask.TaskTypeId, entity.Id, workflowTask.IsResource, workflowTask.TemplateId, "", "", []string{}, workflowTask.Pointer, workflowTask.IsLink, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
		if err != nil {
			return err
		}
	}
	for _, workflowLink := range workflow.Links {
		err = AddWorkflow(tx, workflowLink.LinkedWorkflowId, workflowLink.Name, workflowLink.EntityTypeId, entity.Id, user)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetWorkflowByName(tx *sqlx.Tx, name string) (models.Workflow, error) {
	workflow := models.Workflow{}

	err := base_service.GetByName(tx, "workflow", name, &workflow)
	if err != nil && err == sql.ErrNoRows {
		return models.Workflow{}, error_service.ErrWorkflowNotFound
	} else if err != nil {
		return models.Workflow{}, err
	}

	return workflow, nil
}

func GetWorkflow(tx *sqlx.Tx, workflowId string) (models.Workflow, error) {
	workflow := models.Workflow{}

	err := base_service.Get(tx, "workflow", workflowId, &workflow)
	if err != nil && err == sql.ErrNoRows {
		return models.Workflow{}, error_service.ErrWorkflowNotFound
	} else if err != nil {
		return models.Workflow{}, err
	}

	workflowTasks, err := GetWorkflowTasks(tx, workflowId)
	if err != nil {
		return workflow, err
	}

	workflowEntities, err := GetWorkflowEntities(tx, workflowId)
	if err != nil {
		return workflow, err
	}

	workflowLinks, err := GetWorkflowLinks(tx, workflowId)
	if err != nil {
		return workflow, err
	}

	workflow.Tasks = workflowTasks
	workflow.Entities = workflowEntities
	workflow.Links = workflowLinks

	return workflow, nil
}

func GetWorkflows(tx *sqlx.Tx) ([]models.Workflow, error) {
	workflows := []models.Workflow{}

	err := base_service.GetAll(tx, "workflow", &workflows)
	if err != nil {
		return workflows, err
	}

	allWorkflowsMap := map[string]models.Workflow{}
	for _, workflow := range workflows {
		allWorkflowsMap[workflow.Id] = workflow
	}

	allWorkflowTasks := []models.WorkflowTask{}
	err = base_service.GetAll(tx, "workflow_task", &allWorkflowTasks)
	if err != nil {
		return workflows, err
	}

	allWorkflowTasksMap := map[string][]models.WorkflowTask{}
	for _, task := range allWorkflowTasks {
		if _, ok := allWorkflowTasksMap[task.WorkflowId]; !ok {
			allWorkflowTasksMap[task.WorkflowId] = []models.WorkflowTask{}
		}

		allWorkflowTasksMap[task.WorkflowId] = append(allWorkflowTasksMap[task.WorkflowId], task)
	}

	allWorkflowEntities := []models.WorkflowEntity{}
	err = base_service.GetAll(tx, "workflow_entity", &allWorkflowEntities)
	if err != nil {
		return workflows, err
	}

	allWorkflowEntitiesMap := map[string][]models.WorkflowEntity{}
	for _, entity := range allWorkflowEntities {
		if _, ok := allWorkflowEntitiesMap[entity.WorkflowId]; !ok {
			allWorkflowEntitiesMap[entity.WorkflowId] = []models.WorkflowEntity{}
		}

		allWorkflowEntitiesMap[entity.WorkflowId] = append(allWorkflowEntitiesMap[entity.WorkflowId], entity)
	}

	allWorkflowLinks := []models.WorkflowLink{}
	err = base_service.GetAll(tx, "workflow_link", &allWorkflowLinks)
	if err != nil {
		return workflows, err
	}

	allWorkflowLinksMap := map[string][]models.WorkflowLink{}
	for _, link := range allWorkflowLinks {
		if _, ok := allWorkflowLinksMap[link.WorkflowId]; !ok {
			allWorkflowLinksMap[link.WorkflowId] = []models.WorkflowLink{}
		}
		linkedWorkflow := allWorkflowsMap[link.LinkedWorkflowId]
		link.LinkedWorkflowName = linkedWorkflow.Name
		allWorkflowLinksMap[link.WorkflowId] = append(allWorkflowLinksMap[link.WorkflowId], link)
	}

	for i, workflow := range workflows {
		if tasks, ok := allWorkflowTasksMap[workflow.Id]; ok {
			workflows[i].Tasks = tasks
		} else {
			workflows[i].Tasks = []models.WorkflowTask{}
		}

		if entities, ok := allWorkflowEntitiesMap[workflow.Id]; ok {
			workflows[i].Entities = entities
		} else {
			workflows[i].Entities = []models.WorkflowEntity{}
		}

		if links, ok := allWorkflowLinksMap[workflow.Id]; ok {
			workflows[i].Links = links
		} else {
			workflows[i].Links = []models.WorkflowLink{}
		}
	}

	return workflows, nil
}

func RenameWorkflow(tx *sqlx.Tx, id, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("workflow name cannot be empty")
	}

	err := base_service.Rename(tx, "workflow", id, name)
	if err != nil {
		return err
	}

	err = base_service.UpdateMtime(tx, "workflow", id, utils.GetEpochTime())
	if err != nil {
		return err
	}

	return nil
}

func UpdateWorkflow(tx *sqlx.Tx, workflowId, name string, workflowTasks []models.WorkflowTask, workflowEntities []models.WorkflowEntity, workflowLinks []models.WorkflowLink) (models.Workflow, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.Workflow{}, errors.New("workflow name cannot be empty")
	}

	workflow, err := GetWorkflow(tx, workflowId)
	if err != nil {
		return models.Workflow{}, err
	}

	if name != workflow.Name {
		err := RenameWorkflow(tx, workflowId, name)
		if err != nil {
			return models.Workflow{}, err
		}
		workflow.Name = name
		err = base_service.UpdateMtime(tx, "workflow", workflowId, utils.GetEpochTime())
		if err != nil {
			return models.Workflow{}, err
		}
	}

	originalWorkflowTaskIds := []string{}
	workflowTaskQuery := "SELECT id FROM workflow_task WHERE workflow_id = ?"
	err = tx.Select(&originalWorkflowTaskIds, workflowTaskQuery, workflowId)
	if err != nil {
		return models.Workflow{}, err
	}

	listOfWorkFlowTaskIds := []string{}
	for _, workflowTask := range workflowTasks {
		listOfWorkFlowTaskIds = append(listOfWorkFlowTaskIds, workflowTask.Id)
	}

	for _, originalWorkflowTaskId := range originalWorkflowTaskIds {
		if !utils.Contains(listOfWorkFlowTaskIds, originalWorkflowTaskId) {
			err = DeleteWorkflowTask(tx, originalWorkflowTaskId)
			if err != nil {
				return models.Workflow{}, err
			}
		}
	}

	originalWorkflowEntityIds := []string{}
	workflowEntityQuery := "SELECT id FROM workflow_entity WHERE workflow_id = ?"
	err = tx.Select(&originalWorkflowEntityIds, workflowEntityQuery, workflowId)
	if err != nil {
		return models.Workflow{}, err
	}
	listOfWorkFlowEntityIds := []string{}
	for _, workflowEntity := range workflowEntities {
		listOfWorkFlowEntityIds = append(listOfWorkFlowEntityIds, workflowEntity.Id)
	}

	for _, originalWorkflowEntityId := range originalWorkflowEntityIds {
		if !utils.Contains(listOfWorkFlowEntityIds, originalWorkflowEntityId) {
			err = DeleteWorkflowEntity(tx, originalWorkflowEntityId)
			if err != nil {
				return models.Workflow{}, err
			}
		}
	}

	originalWorkflowLinkIds := []string{}
	workflowLinkQuery := "SELECT id FROM workflow_link WHERE workflow_id = ?"
	err = tx.Select(&originalWorkflowLinkIds, workflowLinkQuery, workflowId)
	if err != nil {
		return models.Workflow{}, err
	}

	listOfWorkFlowLinkIds := []string{}
	for _, workflowLink := range workflowLinks {
		listOfWorkFlowLinkIds = append(listOfWorkFlowLinkIds, workflowLink.Id)
	}

	for _, originalWorkflowLinkId := range originalWorkflowLinkIds {
		if !utils.Contains(listOfWorkFlowLinkIds, originalWorkflowLinkId) {
			err = DeleteWorkflowLink(tx, originalWorkflowLinkId)
			if err != nil {
				return models.Workflow{}, err
			}
		}
	}

	for _, workflowTask := range workflowTasks {
		_, err = GetWorkflowTask(tx, workflowTask.Id)
		if err != nil && err == error_service.ErrWorkflowTaskNotFound {
			_, err := CreateWorkflowTask(tx, "", workflowTask.Name, workflow.Id, workflowTask.TaskTypeId, workflowTask.IsResource, workflowTask.TemplateId, workflowTask.Pointer, workflowTask.IsLink)
			if err != nil {
				return models.Workflow{}, err
			}
		} else {
			_, err := UpdateWorkflowTask(tx, workflowTask.Id, workflowTask.Name, workflowTask.TaskTypeId, workflowTask.IsResource, workflowTask.TemplateId, workflowTask.Pointer, workflowTask.IsLink)
			if err != nil {
				return models.Workflow{}, err
			}
		}
	}
	for _, workflowEntity := range workflowEntities {
		_, err = GetWorkflowEntity(tx, workflowEntity.Id)
		if err != nil && err == error_service.ErrWorkflowEntityNotFound {
			_, err := CreateWorkflowEntity(tx, "", workflowEntity.Name, workflow.Id, workflowEntity.EntityTypeId)
			if err != nil {
				return models.Workflow{}, err
			}
		} else {
			_, err := UpdateWorkflowEntity(tx, workflowEntity.Id, workflowEntity.Name, workflowEntity.EntityTypeId)
			if err != nil {
				return models.Workflow{}, err
			}
		}
	}

	for _, workflowLink := range workflowLinks {
		err := RenameLinkedWorkflow(tx, workflowLink.Id, workflowLink.Name)
		if err != nil {
			return models.Workflow{}, err
		}
	}

	return GetWorkflow(tx, workflow.Id)
}

func RenameLinkedWorkflow(tx *sqlx.Tx, id, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("workflow link name cannot be empty")
	}

	err := base_service.Rename(tx, "workflow_link", id, name)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "workflow_link", id, utils.GetEpochTime())
	if err != nil {
		return err
	}

	return nil
}

func DeleteWorkflow(tx *sqlx.Tx, workflowId string) error {
	err := base_service.Delete(tx, "workflow", workflowId)
	if err != nil {
		return err
	}
	return nil
}

func CreateWorkflowTask(
	tx *sqlx.Tx, id, name, workflowId, taskTypeId string,
	isResource bool, templateId, pointer string, isLink bool,
) (models.WorkflowTask, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.WorkflowTask{}, errors.New("task name cannot be empty")
	}

	if isLink && !utils.IsValidPointer(pointer) {
		return models.WorkflowTask{}, errors.New("invalid pointer, path does not exist")
	}

	params := map[string]any{
		"id":           id,
		"name":         name,
		"workflow_id":  workflowId,
		"task_type_id": taskTypeId,
		"is_resource":  isResource,
		"template_id":  templateId,
		"pointer":      pointer,
		"is_link":      isLink,
	}
	err := base_service.Create(tx, "workflow_task", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return models.WorkflowTask{}, error_service.ErrWorkflowTaskExists
			}
		}
		return models.WorkflowTask{}, err
	}
	workflow, err := GetWorkflowTaskByName(tx, name, workflowId)
	if err != nil {
		return models.WorkflowTask{}, err
	}

	return workflow, nil
}

func UpdateWorkflowTask(
	tx *sqlx.Tx, id string, name, taskTypeId string,
	isResource bool, templateId, pointer string, isLink bool,
) (models.WorkflowTask, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.WorkflowTask{}, errors.New("entity name cannot be empty")
	}

	params := map[string]any{
		"name":         name,
		"task_type_id": taskTypeId,
		"is_resource":  isResource,
		"template_id":  templateId,
		"pointer":      pointer,
		"is_link":      isLink,
	}
	err := base_service.Update(tx, "workflow_task", id, params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return models.WorkflowTask{}, error_service.ErrWorkflowTaskExists
			}
		}
		return models.WorkflowTask{}, err
	}

	err = base_service.UpdateMtime(tx, "workflow_task", id, utils.GetEpochTime())
	if err != nil {
		return models.WorkflowTask{}, err
	}

	workflowTask, err := GetWorkflowTask(tx, id)
	if err != nil {
		return models.WorkflowTask{}, err
	}
	return workflowTask, nil
}

func DeleteWorkflowTask(tx *sqlx.Tx, id string) error {
	err := base_service.Delete(tx, "workflow_task", id)
	if err != nil {
		return err
	}

	return nil
}

func GetWorkflowTasks(tx *sqlx.Tx, workflowId string) ([]models.WorkflowTask, error) {
	workflowTasks := []models.WorkflowTask{}

	conditions := map[string]any{"workflow_id": workflowId}

	err := base_service.GetAllBy(tx, "workflow_task", conditions, &workflowTasks)
	if err != nil {
		return workflowTasks, err
	}
	return workflowTasks, nil
}

func GetWorkflowTask(tx *sqlx.Tx, id string) (models.WorkflowTask, error) {
	task := models.WorkflowTask{}
	err := base_service.Get(tx, "workflow_task", id, &task)
	if err != nil && err == sql.ErrNoRows {
		return models.WorkflowTask{}, error_service.ErrTaskNotFound
	} else if err != nil {
		return models.WorkflowTask{}, err
	}
	return task, nil
}

func GetWorkflowTaskByName(tx *sqlx.Tx, name, workflowId string) (models.WorkflowTask, error) {
	workflow := models.WorkflowTask{}
	query := "SELECT * FROM workflow_task WHERE name = ? AND workflow_id = ?"

	err := tx.Get(&workflow, query, name, workflowId)
	if err != nil && err == sql.ErrNoRows {
		return models.WorkflowTask{}, error_service.ErrWorkflowNotFound
	} else if err != nil {
		return models.WorkflowTask{}, err
	}
	return workflow, nil
}

func CreateWorkflowEntity(
	tx *sqlx.Tx, id string, name, workflow_id, entity_type_id string,
) (models.WorkflowEntity, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.WorkflowEntity{}, errors.New("entity name cannot be empty")
	}

	conditions := map[string]any{
		"name":        name,
		"workflow_id": workflow_id,
	}
	workflowEntity := models.WorkflowEntity{}
	err := base_service.GetBy(tx, "workflow_entity", conditions, &workflowEntity)
	if err == nil {
		return models.WorkflowEntity{}, error_service.ErrWorkflowEntityExists

	}

	params := map[string]any{
		"id":             id,
		"name":           name,
		"workflow_id":    workflow_id,
		"entity_type_id": entity_type_id,
	}
	err = base_service.Create(tx, "workflow_entity", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return models.WorkflowEntity{}, error_service.ErrWorkflowEntityExists
			}
		}
		return models.WorkflowEntity{}, err
	}
	workflowEntity, err = GetWorkflowEntityByName(tx, name, workflow_id)
	if err != nil {
		return models.WorkflowEntity{}, err
	}
	return workflowEntity, nil
}

func UpdateWorkflowEntity(
	tx *sqlx.Tx, id string, name, entity_type_id string,
) (models.WorkflowEntity, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.WorkflowEntity{}, errors.New("entity name cannot be empty")
	}

	params := map[string]any{
		"name":           name,
		"entity_type_id": entity_type_id,
	}
	err := base_service.Update(tx, "workflow_entity", id, params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return models.WorkflowEntity{}, error_service.ErrWorkflowEntityExists
			}
		}
		return models.WorkflowEntity{}, err
	}
	err = base_service.UpdateMtime(tx, "workflow_entity", id, utils.GetEpochTime())
	if err != nil {
		return models.WorkflowEntity{}, err
	}
	workflowEntity, err := GetWorkflowEntity(tx, id)
	if err != nil {
		return models.WorkflowEntity{}, err
	}
	return workflowEntity, nil
}

func DeleteWorkflowEntity(tx *sqlx.Tx, id string) error {
	err := base_service.Delete(tx, "workflow_entity", id)
	if err != nil {
		return err
	}

	return nil
}

func GetWorkflowEntities(tx *sqlx.Tx, workflowId string) ([]models.WorkflowEntity, error) {
	entities := []models.WorkflowEntity{}

	conditions := map[string]any{"workflow_id": workflowId}

	err := base_service.GetAllBy(tx, "workflow_entity", conditions, &entities)
	if err != nil {
		return entities, err
	}
	return entities, nil
}

func GetWorkflowEntity(tx *sqlx.Tx, id string) (models.WorkflowEntity, error) {
	entity := models.WorkflowEntity{}
	err := base_service.Get(tx, "workflow_entity", id, &entity)
	if err != nil && err == sql.ErrNoRows {
		return models.WorkflowEntity{}, error_service.ErrEntityNotFound
	} else if err != nil {
		return models.WorkflowEntity{}, err
	}
	return entity, nil
}
func GetWorkflowEntityByName(tx *sqlx.Tx, name, workflowId string) (models.WorkflowEntity, error) {
	entity := models.WorkflowEntity{}
	query := "SELECT * FROM workflow_entity WHERE name = ? AND workflow_id = ?"

	err := tx.Get(&entity, query, name, workflowId)
	if err != nil && err == sql.ErrNoRows {
		return models.WorkflowEntity{}, error_service.ErrEntityNotFound
	} else if err != nil {
		return models.WorkflowEntity{}, err
	}
	return entity, nil
}

func LinkWorkflow(tx *sqlx.Tx, name, entityTypeId, workflow_id, linked_workflow_id string) error {
	params := map[string]any{
		"name":               name,
		"entity_type_id":     entityTypeId,
		"workflow_id":        workflow_id,
		"linked_workflow_id": linked_workflow_id,
	}
	err := base_service.Create(tx, "workflow_link", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return error_service.ErrWorkflowLinkExists
			}
		}
		return err
	}

	return nil
}

func AddLinkWorkflow(tx *sqlx.Tx, id, name, entityTypeId, workflow_id, linked_workflow_id string) error {
	params := map[string]any{
		"id":                 id,
		"name":               name,
		"entity_type_id":     entityTypeId,
		"workflow_id":        workflow_id,
		"linked_workflow_id": linked_workflow_id,
	}
	err := base_service.Create(tx, "workflow_link", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return error_service.ErrWorkflowLinkExists
			}
		}
		return err
	}

	return nil
}

func DeleteWorkflowLink(tx *sqlx.Tx, id string) error {
	err := base_service.Delete(tx, "workflow_link", id)
	if err != nil {
		return err
	}

	return nil
}

func GetWorkflowLink(tx *sqlx.Tx, id string) (models.WorkflowLink, error) {
	workflowLink := models.WorkflowLink{}

	err := base_service.Get(tx, "workflow_link", id, &workflowLink)
	if err != nil && err == sql.ErrNoRows {
		return models.WorkflowLink{}, error_service.ErrWorkflowLinkNotFound
	} else if err != nil {
		return models.WorkflowLink{}, err
	}
	return workflowLink, err
}

func GetWorkflowLinks(tx *sqlx.Tx, workflowId string) ([]models.WorkflowLink, error) {
	links := []models.WorkflowLink{}
	conditions := map[string]any{"workflow_id": workflowId}

	err := base_service.GetAllBy(tx, "workflow_link", conditions, &links)
	if err != nil {
		return []models.WorkflowLink{}, err
	}

	for i, link := range links {
		workflow, err := GetWorkflow(tx, link.LinkedWorkflowId)
		if err != nil {
			return []models.WorkflowLink{}, err
		}
		links[i].LinkedWorkflowName = workflow.Name
	}
	return links, nil
}
