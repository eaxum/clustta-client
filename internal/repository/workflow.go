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

func CreateWorkflow(tx *sqlx.Tx, id, name string, workflowAssets []models.WorkflowAsset, workflowCollections []models.WorkflowCollection, workflowLinks []models.WorkflowLink) (models.Workflow, error) {
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
				return models.Workflow{}, error_service.ErrAssetExists
			}
		}
		return models.Workflow{}, err
	}
	workflow, err := GetWorkflowByName(tx, name)
	if err != nil {
		return models.Workflow{}, err
	}

	for _, workflowAsset := range workflowAssets {
		_, err := CreateWorkflowAsset(tx, "", workflowAsset.Name, workflow.Id, workflowAsset.AssetTypeId, workflowAsset.IsResource, workflowAsset.TemplateId, workflowAsset.Pointer, workflowAsset.IsLink)
		if err != nil {
			return models.Workflow{}, err
		}
	}
	for _, workflowCollection := range workflowCollections {
		_, err := CreateWorkflowCollection(tx, "", workflowCollection.Name, workflow.Id, workflowCollection.CollectionTypeId)
		if err != nil {
			return models.Workflow{}, err
		}
	}

	for _, workflowLink := range workflowLinks {
		err := LinkWorkflow(tx, workflowLink.Name, workflowLink.CollectionTypeId, workflow.Id, workflowLink.LinkedWorkflowId)
		if err != nil {
			return models.Workflow{}, err
		}
	}

	return GetWorkflow(tx, workflow.Id)
}
func AddWorkflow(tx *sqlx.Tx, workflowId, name, collectionTypeId, parentId string, user auth_service.User) error {
	workflow, err := GetWorkflow(tx, workflowId)
	if err != nil {
		return err
	}
	collection, err := CreateCollection(tx, "", name, "", collectionTypeId, parentId, "", false)
	if err != nil {
		return err
	}

	for _, workflowCollection := range workflow.Collections {
		_, err := CreateCollection(tx, "", workflowCollection.Name, "", workflowCollection.CollectionTypeId, collection.Id, "", false)
		if err != nil {
			return err
		}
	}
	for _, workflowAsset := range workflow.Assets {
		_, err := CreateAsset(tx, "", workflowAsset.Name, workflowAsset.AssetTypeId, collection.Id, workflowAsset.IsResource, workflowAsset.TemplateId, "", "", []string{}, workflowAsset.Pointer, workflowAsset.IsLink, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
		if err != nil {
			return err
		}
	}
	for _, workflowLink := range workflow.Links {
		err = AddWorkflow(tx, workflowLink.LinkedWorkflowId, workflowLink.Name, workflowLink.CollectionTypeId, collection.Id, user)
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

	workflowAssets, err := GetWorkflowAssets(tx, workflowId)
	if err != nil {
		return workflow, err
	}

	workflowCollections, err := GetWorkflowCollections(tx, workflowId)
	if err != nil {
		return workflow, err
	}

	workflowLinks, err := GetWorkflowLinks(tx, workflowId)
	if err != nil {
		return workflow, err
	}

	workflow.Assets = workflowAssets
	workflow.Collections = workflowCollections
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

	allWorkflowAssets := []models.WorkflowAsset{}
	err = base_service.GetAll(tx, "workflow_asset", &allWorkflowAssets)
	if err != nil {
		return workflows, err
	}

	allWorkflowAssetsMap := map[string][]models.WorkflowAsset{}
	for _, asset := range allWorkflowAssets {
		if _, ok := allWorkflowAssetsMap[asset.WorkflowId]; !ok {
			allWorkflowAssetsMap[asset.WorkflowId] = []models.WorkflowAsset{}
		}

		allWorkflowAssetsMap[asset.WorkflowId] = append(allWorkflowAssetsMap[asset.WorkflowId], asset)
	}

	allWorkflowCollections := []models.WorkflowCollection{}
	err = base_service.GetAll(tx, "workflow_collection", &allWorkflowCollections)
	if err != nil {
		return workflows, err
	}

	allWorkflowCollectionsMap := map[string][]models.WorkflowCollection{}
	for _, collection := range allWorkflowCollections {
		if _, ok := allWorkflowCollectionsMap[collection.WorkflowId]; !ok {
			allWorkflowCollectionsMap[collection.WorkflowId] = []models.WorkflowCollection{}
		}

		allWorkflowCollectionsMap[collection.WorkflowId] = append(allWorkflowCollectionsMap[collection.WorkflowId], collection)
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
		if assets, ok := allWorkflowAssetsMap[workflow.Id]; ok {
			workflows[i].Assets = assets
		} else {
			workflows[i].Assets = []models.WorkflowAsset{}
		}

		if collections, ok := allWorkflowCollectionsMap[workflow.Id]; ok {
			workflows[i].Collections = collections
		} else {
			workflows[i].Collections = []models.WorkflowCollection{}
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

func UpdateWorkflow(tx *sqlx.Tx, workflowId, name string, workflowAssets []models.WorkflowAsset, workflowCollections []models.WorkflowCollection, workflowLinks []models.WorkflowLink) (models.Workflow, error) {
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

	originalWorkflowAssetIds := []string{}
	workflowAssetQuery := "SELECT id FROM workflow_asset WHERE workflow_id = ?"
	err = tx.Select(&originalWorkflowAssetIds, workflowAssetQuery, workflowId)
	if err != nil {
		return models.Workflow{}, err
	}

	listOfWorkFlowAssetIds := []string{}
	for _, workflowAsset := range workflowAssets {
		listOfWorkFlowAssetIds = append(listOfWorkFlowAssetIds, workflowAsset.Id)
	}

	for _, originalWorkflowAssetId := range originalWorkflowAssetIds {
		if !utils.Contains(listOfWorkFlowAssetIds, originalWorkflowAssetId) {
			err = DeleteWorkflowAsset(tx, originalWorkflowAssetId)
			if err != nil {
				return models.Workflow{}, err
			}
		}
	}

	originalWorkflowCollectionIds := []string{}
	workflowCollectionQuery := "SELECT id FROM workflow_collection WHERE workflow_id = ?"
	err = tx.Select(&originalWorkflowCollectionIds, workflowCollectionQuery, workflowId)
	if err != nil {
		return models.Workflow{}, err
	}
	listOfWorkFlowCollectionIds := []string{}
	for _, workflowCollection := range workflowCollections {
		listOfWorkFlowCollectionIds = append(listOfWorkFlowCollectionIds, workflowCollection.Id)
	}

	for _, originalWorkflowCollectionId := range originalWorkflowCollectionIds {
		if !utils.Contains(listOfWorkFlowCollectionIds, originalWorkflowCollectionId) {
			err = DeleteWorkflowCollection(tx, originalWorkflowCollectionId)
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

	for _, workflowAsset := range workflowAssets {
		_, err = GetWorkflowAsset(tx, workflowAsset.Id)
		if err != nil && err == error_service.ErrWorkflowAssetNotFound {
			_, err := CreateWorkflowAsset(tx, "", workflowAsset.Name, workflow.Id, workflowAsset.AssetTypeId, workflowAsset.IsResource, workflowAsset.TemplateId, workflowAsset.Pointer, workflowAsset.IsLink)
			if err != nil {
				return models.Workflow{}, err
			}
		} else {
			_, err := UpdateWorkflowAsset(tx, workflowAsset.Id, workflowAsset.Name, workflowAsset.AssetTypeId, workflowAsset.IsResource, workflowAsset.TemplateId, workflowAsset.Pointer, workflowAsset.IsLink)
			if err != nil {
				return models.Workflow{}, err
			}
		}
	}
	for _, workflowCollection := range workflowCollections {
		_, err = GetWorkflowCollection(tx, workflowCollection.Id)
		if err != nil && err == error_service.ErrWorkflowCollectionNotFound {
			_, err := CreateWorkflowCollection(tx, "", workflowCollection.Name, workflow.Id, workflowCollection.CollectionTypeId)
			if err != nil {
				return models.Workflow{}, err
			}
		} else {
			_, err := UpdateWorkflowCollection(tx, workflowCollection.Id, workflowCollection.Name, workflowCollection.CollectionTypeId)
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

func CreateWorkflowAsset(
	tx *sqlx.Tx, id, name, workflowId, assetTypeId string,
	isResource bool, templateId, pointer string, isLink bool,
) (models.WorkflowAsset, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.WorkflowAsset{}, errors.New("asset name cannot be empty")
	}

	if isLink && !utils.IsValidPointer(pointer) {
		return models.WorkflowAsset{}, errors.New("invalid pointer, path does not exist")
	}

	params := map[string]any{
		"id":           id,
		"name":         name,
		"workflow_id":  workflowId,
		"asset_type_id": assetTypeId,
		"is_resource":  isResource,
		"template_id":  templateId,
		"pointer":      pointer,
		"is_link":      isLink,
	}
	err := base_service.Create(tx, "workflow_asset", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return models.WorkflowAsset{}, error_service.ErrWorkflowAssetExists
			}
		}
		return models.WorkflowAsset{}, err
	}
	workflow, err := GetWorkflowAssetByName(tx, name, workflowId)
	if err != nil {
		return models.WorkflowAsset{}, err
	}

	return workflow, nil
}

func UpdateWorkflowAsset(
	tx *sqlx.Tx, id string, name, assetTypeId string,
	isResource bool, templateId, pointer string, isLink bool,
) (models.WorkflowAsset, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.WorkflowAsset{}, errors.New("collection name cannot be empty")
	}

	params := map[string]any{
		"name":         name,
		"asset_type_id": assetTypeId,
		"is_resource":  isResource,
		"template_id":  templateId,
		"pointer":      pointer,
		"is_link":      isLink,
	}
	err := base_service.Update(tx, "workflow_asset", id, params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return models.WorkflowAsset{}, error_service.ErrWorkflowAssetExists
			}
		}
		return models.WorkflowAsset{}, err
	}

	err = base_service.UpdateMtime(tx, "workflow_asset", id, utils.GetEpochTime())
	if err != nil {
		return models.WorkflowAsset{}, err
	}

	workflowAsset, err := GetWorkflowAsset(tx, id)
	if err != nil {
		return models.WorkflowAsset{}, err
	}
	return workflowAsset, nil
}

func DeleteWorkflowAsset(tx *sqlx.Tx, id string) error {
	err := base_service.Delete(tx, "workflow_asset", id)
	if err != nil {
		return err
	}

	return nil
}

func GetWorkflowAssets(tx *sqlx.Tx, workflowId string) ([]models.WorkflowAsset, error) {
	workflowAssets := []models.WorkflowAsset{}

	conditions := map[string]any{"workflow_id": workflowId}

	err := base_service.GetAllBy(tx, "workflow_asset", conditions, &workflowAssets)
	if err != nil {
		return workflowAssets, err
	}
	return workflowAssets, nil
}

func GetWorkflowAsset(tx *sqlx.Tx, id string) (models.WorkflowAsset, error) {
	asset := models.WorkflowAsset{}
	err := base_service.Get(tx, "workflow_asset", id, &asset)
	if err != nil && err == sql.ErrNoRows {
		return models.WorkflowAsset{}, error_service.ErrAssetNotFound
	} else if err != nil {
		return models.WorkflowAsset{}, err
	}
	return asset, nil
}

func GetWorkflowAssetByName(tx *sqlx.Tx, name, workflowId string) (models.WorkflowAsset, error) {
	workflow := models.WorkflowAsset{}
	query := "SELECT * FROM workflow_asset WHERE name = ? AND workflow_id = ?"

	err := tx.Get(&workflow, query, name, workflowId)
	if err != nil && err == sql.ErrNoRows {
		return models.WorkflowAsset{}, error_service.ErrWorkflowNotFound
	} else if err != nil {
		return models.WorkflowAsset{}, err
	}
	return workflow, nil
}

func CreateWorkflowCollection(
	tx *sqlx.Tx, id string, name, workflow_id, collection_type_id string,
) (models.WorkflowCollection, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.WorkflowCollection{}, errors.New("collection name cannot be empty")
	}

	conditions := map[string]any{
		"name":        name,
		"workflow_id": workflow_id,
	}
	workflowCollection := models.WorkflowCollection{}
	err := base_service.GetBy(tx, "workflow_collection", conditions, &workflowCollection)
	if err == nil {
		return models.WorkflowCollection{}, error_service.ErrWorkflowCollectionExists

	}

	params := map[string]any{
		"id":             id,
		"name":           name,
		"workflow_id":    workflow_id,
		"collection_type_id": collection_type_id,
	}
	err = base_service.Create(tx, "workflow_collection", params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return models.WorkflowCollection{}, error_service.ErrWorkflowCollectionExists
			}
		}
		return models.WorkflowCollection{}, err
	}
	workflowCollection, err = GetWorkflowCollectionByName(tx, name, workflow_id)
	if err != nil {
		return models.WorkflowCollection{}, err
	}
	return workflowCollection, nil
}

func UpdateWorkflowCollection(
	tx *sqlx.Tx, id string, name, collection_type_id string,
) (models.WorkflowCollection, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.WorkflowCollection{}, errors.New("collection name cannot be empty")
	}

	params := map[string]any{
		"name":           name,
		"collection_type_id": collection_type_id,
	}
	err := base_service.Update(tx, "workflow_collection", id, params)
	if err != nil {
		//FIXME check back here for error handling
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return models.WorkflowCollection{}, error_service.ErrWorkflowCollectionExists
			}
		}
		return models.WorkflowCollection{}, err
	}
	err = base_service.UpdateMtime(tx, "workflow_collection", id, utils.GetEpochTime())
	if err != nil {
		return models.WorkflowCollection{}, err
	}
	workflowCollection, err := GetWorkflowCollection(tx, id)
	if err != nil {
		return models.WorkflowCollection{}, err
	}
	return workflowCollection, nil
}

func DeleteWorkflowCollection(tx *sqlx.Tx, id string) error {
	err := base_service.Delete(tx, "workflow_collection", id)
	if err != nil {
		return err
	}

	return nil
}

func GetWorkflowCollections(tx *sqlx.Tx, workflowId string) ([]models.WorkflowCollection, error) {
	collections := []models.WorkflowCollection{}

	conditions := map[string]any{"workflow_id": workflowId}

	err := base_service.GetAllBy(tx, "workflow_collection", conditions, &collections)
	if err != nil {
		return collections, err
	}
	return collections, nil
}

func GetWorkflowCollection(tx *sqlx.Tx, id string) (models.WorkflowCollection, error) {
	collection := models.WorkflowCollection{}
	err := base_service.Get(tx, "workflow_collection", id, &collection)
	if err != nil && err == sql.ErrNoRows {
		return models.WorkflowCollection{}, error_service.ErrCollectionNotFound
	} else if err != nil {
		return models.WorkflowCollection{}, err
	}
	return collection, nil
}
func GetWorkflowCollectionByName(tx *sqlx.Tx, name, workflowId string) (models.WorkflowCollection, error) {
	collection := models.WorkflowCollection{}
	query := "SELECT * FROM workflow_collection WHERE name = ? AND workflow_id = ?"

	err := tx.Get(&collection, query, name, workflowId)
	if err != nil && err == sql.ErrNoRows {
		return models.WorkflowCollection{}, error_service.ErrCollectionNotFound
	} else if err != nil {
		return models.WorkflowCollection{}, err
	}
	return collection, nil
}

func LinkWorkflow(tx *sqlx.Tx, name, collectionTypeId, workflow_id, linked_workflow_id string) error {
	params := map[string]any{
		"name":               name,
		"collection_type_id":     collectionTypeId,
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

func AddLinkWorkflow(tx *sqlx.Tx, id, name, collectionTypeId, workflow_id, linked_workflow_id string) error {
	params := map[string]any{
		"id":                 id,
		"name":               name,
		"collection_type_id":     collectionTypeId,
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
