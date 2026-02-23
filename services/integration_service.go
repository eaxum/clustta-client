package services

import (
	"clustta/internal/integrations"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"errors"
	"strings"
	"time"
)

// IntegrationService handles external integration operations (Kitsu, ClickUp, etc.)
// Exposed to frontend via Wails bindings.
type IntegrationService struct {
}

// ═══════════════════════════════════════════════════════════════════════════
// INTEGRATION REGISTRY
// ═══════════════════════════════════════════════════════════════════════════

// GetAvailableIntegrations returns all registered integrations.
// Each integration contains id, name, description, and auth requirements.
func (s *IntegrationService) GetAvailableIntegrations() []integrations.IntegrationInfo {
	return integrations.GetAllInfo()
}

// GetIntegration returns info for a specific integration by ID.
func (s *IntegrationService) GetIntegration(integrationId string) (integrations.IntegrationInfo, error) {
	integration, err := integrations.Get(integrationId)
	if err != nil {
		return integrations.IntegrationInfo{}, err
	}
	return integration.GetInfo(), nil
}

// ═══════════════════════════════════════════════════════════════════════════
// AUTHENTICATION
// ═══════════════════════════════════════════════════════════════════════════

// Authenticate authenticates with an external integration.
// Returns auth result with user info and token on success.
func (s *IntegrationService) Authenticate(integrationId string, credentials map[string]string) (integrations.AuthResult, error) {
	integration, err := integrations.Get(integrationId)
	if err != nil {
		return integrations.AuthResult{}, err
	}
	return integration.Authenticate(credentials)
}

// ValidateToken validates an existing token is still valid.
func (s *IntegrationService) ValidateToken(integrationId, token, apiUrl string) (bool, error) {
	integration, err := integrations.Get(integrationId)
	if err != nil {
		return false, err
	}
	return integration.ValidateToken(token, apiUrl)
}

// ═══════════════════════════════════════════════════════════════════════════
// PROJECT LINKING
// ═══════════════════════════════════════════════════════════════════════════

// GetLinkedIntegration returns the integration linked to a project.
// Returns empty struct if no integration is linked.
func (s *IntegrationService) GetLinkedIntegration(projectPath string) (models.IntegrationProject, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.IntegrationProject{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.IntegrationProject{}, err
	}
	defer tx.Rollback()

	return repository.GetIntegrationProject(tx)
}

// GetExternalProjects fetches available projects from an external integration.
// Requires valid token for the integration.
func (s *IntegrationService) GetExternalProjects(integrationId, token, apiUrl string) ([]integrations.ExternalProject, error) {
	integration, err := integrations.Get(integrationId)
	if err != nil {
		return nil, err
	}
	return integration.GetProjects(token, apiUrl)
}

// LinkProject links a Clustta project to an external project.
// Returns error if project already has an integration.
func (s *IntegrationService) LinkProject(projectPath, integrationId, externalProjectId, externalProjectName, apiUrl, syncOptions, userId string) (models.IntegrationProject, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.IntegrationProject{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.IntegrationProject{}, err
	}
	defer tx.Rollback()

	linkedAt := time.Now().UTC().Format(time.RFC3339)
	integration, err := repository.CreateIntegrationProject(
		tx,
		"",
		integrationId,
		externalProjectId,
		externalProjectName,
		apiUrl,
		syncOptions,
		userId,
		linkedAt,
	)
	if err != nil {
		return models.IntegrationProject{}, err
	}

	err = tx.Commit()
	if err != nil {
		return models.IntegrationProject{}, err
	}
	return integration, nil
}

// UnlinkProject removes the integration link from a project.
// Also deletes all collection and asset mappings.
func (s *IntegrationService) UnlinkProject(projectPath string) error {
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

	integration, err := repository.GetIntegrationProject(tx)
	if err != nil {
		return err
	}

	err = repository.DeleteIntegrationProject(tx, integration.Id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ═══════════════════════════════════════════════════════════════════════════
// SYNC PREVIEW
// ═══════════════════════════════════════════════════════════════════════════

// GetSyncPreview fetches external hierarchy and compares with local state.
// Returns preview of what will be created/updated/unchanged.
func (s *IntegrationService) GetSyncPreview(projectPath, token string) (integrations.SyncPreview, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return integrations.SyncPreview{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return integrations.SyncPreview{}, err
	}
	defer tx.Rollback()

	// Get linked integration
	integrationProject, err := repository.GetIntegrationProject(tx)
	if err != nil {
		return integrations.SyncPreview{}, errors.New("no integration linked to this project")
	}

	// Get integration client
	integration, err := integrations.Get(integrationProject.IntegrationId)
	if err != nil {
		return integrations.SyncPreview{}, err
	}

	// Fetch external hierarchy
	entities, err := integration.GetProjectEntities(token, integrationProject.ApiUrl, integrationProject.ExternalProjectId)
	if err != nil {
		return integrations.SyncPreview{}, err
	}

	tasks, err := integration.GetProjectTasks(token, integrationProject.ApiUrl, integrationProject.ExternalProjectId)
	if err != nil {
		return integrations.SyncPreview{}, err
	}

	// Get existing mappings
	existingCollections, err := repository.GetCollectionMappings(tx, integrationProject.IntegrationId)
	if err != nil {
		return integrations.SyncPreview{}, err
	}
	existingAssets, err := repository.GetAssetMappings(tx, integrationProject.IntegrationId)
	if err != nil {
		return integrations.SyncPreview{}, err
	}

	// Build lookup maps for existing mappings
	existingCollectionMap := make(map[string]models.IntegrationCollectionMapping)
	for _, m := range existingCollections {
		existingCollectionMap[m.ExternalId] = m
	}
	existingAssetMap := make(map[string]models.IntegrationAssetMapping)
	for _, m := range existingAssets {
		existingAssetMap[m.ExternalId] = m
	}

	// Build sync preview
	preview := integrations.SyncPreview{
		Collections: []integrations.SyncCollection{},
		Assets:      []integrations.SyncAsset{},
	}

	// Process entities (collections)
	for _, entity := range entities {
		syncColl := integrations.SyncCollection{
			ExternalID:   entity.ID,
			ExternalName: entity.Name,
			ExternalPath: entity.Path,
			ExternalType: entity.Type,
		}
		if existing, ok := existingCollectionMap[entity.ID]; ok {
			syncColl.Action = "unchanged"
			syncColl.CollectionID = existing.CollectionId
			// Check if name changed
			if existing.ExternalName != entity.Name {
				syncColl.Action = "update"
			}
		} else {
			syncColl.Action = "create"
		}
		preview.Collections = append(preview.Collections, syncColl)
	}

	// Process tasks (assets)
	for _, task := range tasks {
		syncAsset := integrations.SyncAsset{
			ExternalID:   task.ID,
			ExternalName: task.Name,
			ExternalType: task.TaskType,
		}
		if existing, ok := existingAssetMap[task.ID]; ok {
			syncAsset.Action = "unchanged"
			syncAsset.AssetID = existing.AssetId
			// Check if name or status changed
			if existing.ExternalName != task.Name || existing.ExternalStatus != task.Status {
				syncAsset.Action = "update"
			}
		} else {
			syncAsset.Action = "create"
		}
		preview.Assets = append(preview.Assets, syncAsset)
	}

	return preview, nil
}

// ═══════════════════════════════════════════════════════════════════════════
// SYNC EXECUTION
// ═══════════════════════════════════════════════════════════════════════════

// ExecuteSync stores mappings for selected external items.
// Collections and assets are not automatically created - users create them manually
// and the mappings help track what's been imported.
func (s *IntegrationService) ExecuteSync(projectPath, token string, selectedCollections, selectedAssets []string) error {
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

	// Get linked integration
	integrationProject, err := repository.GetIntegrationProject(tx)
	if err != nil {
		return errors.New("no integration linked to this project")
	}

	// Get integration client
	integration, err := integrations.Get(integrationProject.IntegrationId)
	if err != nil {
		return err
	}

	// Fetch fresh hierarchy data
	entities, err := integration.GetProjectEntities(token, integrationProject.ApiUrl, integrationProject.ExternalProjectId)
	if err != nil {
		return err
	}

	tasks, err := integration.GetProjectTasks(token, integrationProject.ApiUrl, integrationProject.ExternalProjectId)
	if err != nil {
		return err
	}

	// Build maps for quick lookup
	selectedCollectionSet := make(map[string]bool)
	for _, id := range selectedCollections {
		selectedCollectionSet[id] = true
	}
	selectedAssetSet := make(map[string]bool)
	for _, id := range selectedAssets {
		selectedAssetSet[id] = true
	}

	syncedAt := time.Now().UTC().Format(time.RFC3339)

	// Process collections - create mappings (without creating actual collections)
	for _, entity := range entities {
		if !selectedCollectionSet[entity.ID] {
			continue
		}

		// Check if already mapped
		_, err := repository.GetCollectionMappingByExternalId(tx, integrationProject.IntegrationId, entity.ID)
		if err == nil {
			// Already mapped, skip
			continue
		}

		// Create mapping without linking to a collection yet
		_, err = repository.CreateCollectionMapping(
			tx,
			"",
			integrationProject.IntegrationId,
			entity.ID,
			entity.Type,
			entity.Name,
			entity.ParentID,
			entity.Path,
			"{}",
			"", // collection_id empty - will be linked when user creates/selects collection
			syncedAt,
		)
		if err != nil {
			return err
		}
	}

	// Process assets - create mappings (without creating actual assets)
	for _, task := range tasks {
		if !selectedAssetSet[task.ID] {
			continue
		}

		// Check if already mapped
		_, err := repository.GetAssetMappingByExternalId(tx, integrationProject.IntegrationId, task.ID)
		if err == nil {
			// Already mapped, skip
			continue
		}

		// Convert assignees to JSON string
		assigneesStr := "[]"
		if len(task.Assignees) > 0 {
			assigneesStr = "[\"" + strings.Join(task.Assignees, "\",\"") + "\"]"
		}

		// Create mapping without linking to an asset yet
		_, err = repository.CreateAssetMapping(
			tx,
			"",
			integrationProject.IntegrationId,
			task.ID,
			task.Name,
			task.ParentID,
			task.Type,
			task.Status,
			assigneesStr,
			"{}",
			"", // asset_id empty - will be linked when user creates/selects asset
			syncedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ═══════════════════════════════════════════════════════════════════════════
// MAPPING QUERIES
// ═══════════════════════════════════════════════════════════════════════════

// GetCollectionMappings returns all collection mappings for the project.
func (s *IntegrationService) GetCollectionMappings(projectPath string) ([]models.IntegrationCollectionMapping, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return nil, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	integration, err := repository.GetIntegrationProject(tx)
	if err != nil {
		return nil, err
	}

	return repository.GetCollectionMappings(tx, integration.IntegrationId)
}

// GetAssetMappings returns all asset mappings for the project.
func (s *IntegrationService) GetAssetMappings(projectPath string) ([]models.IntegrationAssetMapping, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return nil, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	integration, err := repository.GetIntegrationProject(tx)
	if err != nil {
		return nil, err
	}

	return repository.GetAssetMappings(tx, integration.IntegrationId)
}

// GetAssetExternalInfo returns the external info for a synced asset.
func (s *IntegrationService) GetAssetExternalInfo(projectPath, assetId string) (models.IntegrationAssetMapping, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.IntegrationAssetMapping{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.IntegrationAssetMapping{}, err
	}
	defer tx.Rollback()

	return repository.GetAssetMappingByAssetId(tx, assetId)
}

// GetCollectionExternalInfo returns the external info for a synced collection.
func (s *IntegrationService) GetCollectionExternalInfo(projectPath, collectionId string) (models.IntegrationCollectionMapping, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.IntegrationCollectionMapping{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.IntegrationCollectionMapping{}, err
	}
	defer tx.Rollback()

	return repository.GetCollectionMappingByCollectionId(tx, collectionId)
}
