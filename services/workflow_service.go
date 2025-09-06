package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
)

type WorkflowService struct{}

func (t *WorkflowService) GetWorkflows(projectPath string) ([]models.Workflow, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return []models.Workflow{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.Workflow{}, err
	}
	defer tx.Rollback()

	workflows, err := repository.GetWorkflows(tx)
	if err != nil {
		return []models.Workflow{}, err
	}
	return workflows, nil
}

func (t *WorkflowService) CreateWorkflow(projectPath, name string, workflowTasks []models.WorkflowTask, workflowEntities []models.WorkflowEntity, workflowLinks []models.WorkflowLink) (models.Workflow, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Workflow{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Workflow{}, err
	}
	defer tx.Rollback()

	workflow, err := repository.CreateWorkflow(tx, "", name, workflowTasks, workflowEntities, workflowLinks)
	if err != nil {
		tx.Rollback()
		return models.Workflow{}, err
	}
	tx.Commit()

	return workflow, nil
}

func (t *WorkflowService) AddWorkflow(projectPath, workflow_id, name, entityTypeId, parentId string) error {
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

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	err = repository.AddWorkflow(tx, workflow_id, name, entityTypeId, parentId, user)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	return nil
}

func (t *WorkflowService) UpdateWorkflow(projectPath, workflowId, name string, workflowTasks []models.WorkflowTask, workflowEntities []models.WorkflowEntity, workflowLinks []models.WorkflowLink) (models.Workflow, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.Workflow{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Workflow{}, err
	}
	defer tx.Rollback()

	workflow, err := repository.UpdateWorkflow(tx, workflowId, name, workflowTasks, workflowEntities, workflowLinks)
	if err != nil {
		tx.Rollback()
		return models.Workflow{}, err
	}
	tx.Commit()

	return workflow, nil
}
