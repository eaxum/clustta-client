package repository

import (
	"clustta/internal/base_service"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// CreateIntegrationProject creates or updates an integration link for the project.
// If an integration already exists, it will be updated with the new values.
func CreateIntegrationProject(tx *sqlx.Tx, id, integrationId, externalProjectId, externalProjectName, apiUrl, syncOptions, linkedByUserId, linkedAt string) (models.IntegrationProject, error) {
	integration := models.IntegrationProject{}

	if id == "" {
		id = uuid.New().String()
	}

	// Check if integration already exists - update instead of error
	existing, err := GetIntegrationProject(tx)
	if err == nil && existing.Id != "" {
		err = UpdateIntegrationProject(tx, existing.Id, map[string]interface{}{
			"integration_id":        integrationId,
			"external_project_id":   externalProjectId,
			"external_project_name": externalProjectName,
			"api_url":               apiUrl,
			"sync_options":          syncOptions,
			"linked_by_user_id":     linkedByUserId,
			"linked_at":             linkedAt,
			"enabled":               true,
		})
		if err != nil {
			return integration, err
		}
		return GetIntegrationProject(tx)
	}

	params := map[string]interface{}{
		"id":                    id,
		"integration_id":        integrationId,
		"external_project_id":   externalProjectId,
		"external_project_name": externalProjectName,
		"api_url":               apiUrl,
		"sync_options":          syncOptions,
		"linked_by_user_id":     linkedByUserId,
		"linked_at":             linkedAt,
		"enabled":               true,
	}
	err = base_service.Create(tx, "integration_project", params)
	if err != nil {
		return integration, err
	}
	err = base_service.Get(tx, "integration_project", id, &integration)
	if err != nil {
		return integration, err
	}
	return integration, nil
}

// GetIntegrationProject retrieves the integration link for the project.
// Returns empty struct and error if no integration exists.
func GetIntegrationProject(tx *sqlx.Tx) (models.IntegrationProject, error) {
	integration := models.IntegrationProject{}
	integrations := []models.IntegrationProject{}
	err := base_service.GetAll(tx, "integration_project", &integrations)
	if err != nil {
		return integration, err
	}
	if len(integrations) == 0 {
		return integration, errors.New("no integration found")
	}
	return integrations[0], nil
}

// GetIntegrationProjects retrieves all integration projects.
func GetIntegrationProjects(tx *sqlx.Tx) ([]models.IntegrationProject, error) {
	integrations := []models.IntegrationProject{}
	err := base_service.GetAll(tx, "integration_project", &integrations)
	return integrations, err
}

// UpdateIntegrationProject updates the integration project settings.
func UpdateIntegrationProject(tx *sqlx.Tx, id string, params map[string]interface{}) error {
	err := base_service.Update(tx, "integration_project", id, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, "integration_project", id, utils.GetEpochTime())
	return err
}

// DeleteIntegrationProject removes the integration link and all mappings.
func DeleteIntegrationProject(tx *sqlx.Tx, id string) error {
	// Get integration to know the integration_id for cleanup
	integration := models.IntegrationProject{}
	err := base_service.Get(tx, "integration_project", id, &integration)
	if err != nil {
		return err
	}

	// Delete all collection mappings for this integration
	err = deleteCollectionMappingsByIntegration(tx, integration.IntegrationId)
	if err != nil {
		return err
	}

	// Delete all asset mappings for this integration
	err = deleteAssetMappingsByIntegration(tx, integration.IntegrationId)
	if err != nil {
		return err
	}

	// Delete the integration project itself
	return base_service.Delete(tx, "integration_project", id)
}

// CreateCollectionMapping creates a mapping between external collection and Clustta collection.
func CreateCollectionMapping(tx *sqlx.Tx, id, integrationId, externalId, externalType, externalName, externalParentId, externalPath, externalMetadata, collectionId, syncedAt string) (models.IntegrationCollectionMapping, error) {
	mapping := models.IntegrationCollectionMapping{}
	params := map[string]interface{}{
		"id":                 id,
		"integration_id":     integrationId,
		"external_id":        externalId,
		"external_type":      externalType,
		"external_name":      externalName,
		"external_parent_id": externalParentId,
		"external_path":      externalPath,
		"external_metadata":  externalMetadata,
		"collection_id":      collectionId,
		"synced_at":          syncedAt,
	}
	err := base_service.Create(tx, "integration_collection_mapping", params)
	if err != nil {
		return mapping, err
	}
	err = base_service.Get(tx, "integration_collection_mapping", id, &mapping)
	return mapping, err
}

// GetCollectionMapping retrieves a collection mapping by ID.
func GetCollectionMapping(tx *sqlx.Tx, id string) (models.IntegrationCollectionMapping, error) {
	mapping := models.IntegrationCollectionMapping{}
	err := base_service.Get(tx, "integration_collection_mapping", id, &mapping)
	return mapping, err
}

// GetCollectionMappingByExternalId retrieves mapping by external ID.
func GetCollectionMappingByExternalId(tx *sqlx.Tx, integrationId, externalId string) (models.IntegrationCollectionMapping, error) {
	mapping := models.IntegrationCollectionMapping{}
	conditions := map[string]interface{}{
		"integration_id": integrationId,
		"external_id":    externalId,
	}
	mappings := []models.IntegrationCollectionMapping{}
	err := base_service.GetAllBy(tx, "integration_collection_mapping", conditions, &mappings)
	if err != nil {
		return mapping, err
	}
	if len(mappings) == 0 {
		return mapping, errors.New("collection mapping not found")
	}
	return mappings[0], nil
}

// GetCollectionMappingByCollectionId retrieves mapping by Clustta collection ID.
func GetCollectionMappingByCollectionId(tx *sqlx.Tx, collectionId string) (models.IntegrationCollectionMapping, error) {
	mapping := models.IntegrationCollectionMapping{}
	conditions := map[string]interface{}{
		"collection_id": collectionId,
	}
	mappings := []models.IntegrationCollectionMapping{}
	err := base_service.GetAllBy(tx, "integration_collection_mapping", conditions, &mappings)
	if err != nil {
		return mapping, err
	}
	if len(mappings) == 0 {
		return mapping, errors.New("collection mapping not found")
	}
	return mappings[0], nil
}

// GetCollectionMappings retrieves all collection mappings for an integration.
func GetCollectionMappings(tx *sqlx.Tx, integrationId string) ([]models.IntegrationCollectionMapping, error) {
	mappings := []models.IntegrationCollectionMapping{}
	conditions := map[string]interface{}{
		"integration_id": integrationId,
	}
	err := base_service.GetAllBy(tx, "integration_collection_mapping", conditions, &mappings)
	return mappings, err
}

// GetAllCollectionMappings retrieves all collection mappings.
func GetAllCollectionMappings(tx *sqlx.Tx) ([]models.IntegrationCollectionMapping, error) {
	mappings := []models.IntegrationCollectionMapping{}
	err := base_service.GetAll(tx, "integration_collection_mapping", &mappings)
	return mappings, err
}

// UpdateCollectionMapping updates a collection mapping.
func UpdateCollectionMapping(tx *sqlx.Tx, id string, params map[string]interface{}) error {
	err := base_service.Update(tx, "integration_collection_mapping", id, params)
	if err != nil {
		return err
	}
	return base_service.UpdateMtime(tx, "integration_collection_mapping", id, utils.GetEpochTime())
}

// DeleteCollectionMapping deletes a single collection mapping.
func DeleteCollectionMapping(tx *sqlx.Tx, id string) error {
	return base_service.Delete(tx, "integration_collection_mapping", id)
}

// deleteCollectionMappingsByIntegration deletes all collection mappings for an integration.
func deleteCollectionMappingsByIntegration(tx *sqlx.Tx, integrationId string) error {
	query := `DELETE FROM integration_collection_mapping WHERE integration_id = ?`
	_, err := tx.Exec(query, integrationId)
	return err
}

// CreateAssetMapping creates a mapping between external asset and Clustta asset.
func CreateAssetMapping(tx *sqlx.Tx, id, integrationId, externalId, externalName, externalParentId, externalType, externalStatus, externalAssignees, externalMetadata, assetId, syncedAt string) (models.IntegrationAssetMapping, error) {
	mapping := models.IntegrationAssetMapping{}
	params := map[string]interface{}{
		"id":                 id,
		"integration_id":     integrationId,
		"external_id":        externalId,
		"external_name":      externalName,
		"external_parent_id": externalParentId,
		"external_type":      externalType,
		"external_status":    externalStatus,
		"external_assignees": externalAssignees,
		"external_metadata":  externalMetadata,
		"asset_id":           assetId,
		"synced_at":          syncedAt,
	}
	err := base_service.Create(tx, "integration_asset_mapping", params)
	if err != nil {
		return mapping, err
	}
	err = base_service.Get(tx, "integration_asset_mapping", id, &mapping)
	return mapping, err
}

// GetAssetMapping retrieves an asset mapping by ID.
func GetAssetMapping(tx *sqlx.Tx, id string) (models.IntegrationAssetMapping, error) {
	mapping := models.IntegrationAssetMapping{}
	err := base_service.Get(tx, "integration_asset_mapping", id, &mapping)
	return mapping, err
}

// GetAssetMappingByExternalId retrieves mapping by external ID.
func GetAssetMappingByExternalId(tx *sqlx.Tx, integrationId, externalId string) (models.IntegrationAssetMapping, error) {
	mapping := models.IntegrationAssetMapping{}
	conditions := map[string]interface{}{
		"integration_id": integrationId,
		"external_id":    externalId,
	}
	mappings := []models.IntegrationAssetMapping{}
	err := base_service.GetAllBy(tx, "integration_asset_mapping", conditions, &mappings)
	if err != nil {
		return mapping, err
	}
	if len(mappings) == 0 {
		return mapping, errors.New("asset mapping not found")
	}
	return mappings[0], nil
}

// GetAssetMappingByAssetId retrieves mapping by Clustta asset ID.
func GetAssetMappingByAssetId(tx *sqlx.Tx, assetId string) (models.IntegrationAssetMapping, error) {
	mapping := models.IntegrationAssetMapping{}
	conditions := map[string]interface{}{
		"asset_id": assetId,
	}
	mappings := []models.IntegrationAssetMapping{}
	err := base_service.GetAllBy(tx, "integration_asset_mapping", conditions, &mappings)
	if err != nil {
		return mapping, err
	}
	if len(mappings) == 0 {
		return mapping, errors.New("asset mapping not found")
	}
	return mappings[0], nil
}

// GetAssetMappingsByAssetId retrieves all mappings for a Clustta asset ID.
func GetAssetMappingsByAssetId(tx *sqlx.Tx, assetId string) ([]models.IntegrationAssetMapping, error) {
	mappings := []models.IntegrationAssetMapping{}
	conditions := map[string]interface{}{
		"asset_id": assetId,
	}
	err := base_service.GetAllBy(tx, "integration_asset_mapping", conditions, &mappings)
	return mappings, err
}

// GetAssetMappings retrieves all asset mappings for an integration.
func GetAssetMappings(tx *sqlx.Tx, integrationId string) ([]models.IntegrationAssetMapping, error) {
	mappings := []models.IntegrationAssetMapping{}
	conditions := map[string]interface{}{
		"integration_id": integrationId,
	}
	err := base_service.GetAllBy(tx, "integration_asset_mapping", conditions, &mappings)
	return mappings, err
}

// GetAllAssetMappings retrieves all asset mappings.
func GetAllAssetMappings(tx *sqlx.Tx) ([]models.IntegrationAssetMapping, error) {
	mappings := []models.IntegrationAssetMapping{}
	err := base_service.GetAll(tx, "integration_asset_mapping", &mappings)
	return mappings, err
}

// GetAssetMappingsByCollectionMapping returns all asset mappings whose external_parent_id
// matches the external_id of the given collection mapping.
func GetAssetMappingsByCollectionMapping(tx *sqlx.Tx, integrationId, externalParentId string) ([]models.IntegrationAssetMapping, error) {
	mappings := []models.IntegrationAssetMapping{}
	conditions := map[string]interface{}{
		"integration_id":     integrationId,
		"external_parent_id": externalParentId,
	}
	err := base_service.GetAllBy(tx, "integration_asset_mapping", conditions, &mappings)
	return mappings, err
}

// UpdateAssetMapping updates an asset mapping.
func UpdateAssetMapping(tx *sqlx.Tx, id string, params map[string]interface{}) error {
	err := base_service.Update(tx, "integration_asset_mapping", id, params)
	if err != nil {
		return err
	}
	return base_service.UpdateMtime(tx, "integration_asset_mapping", id, utils.GetEpochTime())
}

// DeleteAssetMapping deletes a single asset mapping.
func DeleteAssetMapping(tx *sqlx.Tx, id string) error {
	return base_service.Delete(tx, "integration_asset_mapping", id)
}

// deleteAssetMappingsByIntegration deletes all asset mappings for an integration.
func deleteAssetMappingsByIntegration(tx *sqlx.Tx, integrationId string) error {
	query := `DELETE FROM integration_asset_mapping WHERE integration_id = ?`
	_, err := tx.Exec(query, integrationId)
	return err
}
