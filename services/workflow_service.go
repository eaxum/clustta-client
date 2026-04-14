package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
)

type WorkflowService struct{}

// Retrieves all workflows from the project database
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

// Creates a new workflow with assets, collections, and links
func (t *WorkflowService) CreateWorkflow(projectPath, name string, workflowAssets []models.WorkflowAsset, workflowCollections []models.WorkflowCollection, workflowLinks []models.WorkflowLink) (models.Workflow, error) {
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

	workflow, err := repository.CreateWorkflow(tx, "", name, workflowAssets, workflowCollections, workflowLinks)
	if err != nil {
		tx.Rollback()
		return models.Workflow{}, err
	}
	tx.Commit()

	return workflow, nil
}

// Adds a workflow to an collection with specified parent and type
func (t *WorkflowService) AddWorkflow(projectPath, workflow_id, name, collectionTypeId, parentId string) error {
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

	err = repository.AddWorkflow(tx, workflow_id, name, collectionTypeId, parentId, user)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	return nil
}

// Updates an existing workflow including assets, collections, and links
func (t *WorkflowService) UpdateWorkflow(projectPath, workflowId, name string, workflowAssets []models.WorkflowAsset, workflowCollections []models.WorkflowCollection, workflowLinks []models.WorkflowLink) (models.Workflow, error) {
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

	workflow, err := repository.UpdateWorkflow(tx, workflowId, name, workflowAssets, workflowCollections, workflowLinks)
	if err != nil {
		tx.Rollback()
		return models.Workflow{}, err
	}
	tx.Commit()

	return workflow, nil
}
