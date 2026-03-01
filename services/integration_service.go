package services

import (
	"clustta/internal/integrations"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// IntegrationService handles external integration operations (Kitsu, etc.)
// Exposed to frontend via Wails bindings.
type IntegrationService struct {
}

// GetAvailableIntegrations returns all registered integrations.
// Each integration contains id, name, description, and auth requirements.
func (s *IntegrationService) GetAvailableIntegrations() []integrations.IntegrationInfo {
	return integrations.GetAllInfo()
}

// Authenticate authenticates with an external integration.
// Returns auth result with user info and token on success.
func (s *IntegrationService) Authenticate(integrationId string, credentials map[string]string) (integrations.AuthResult, error) {
	integration, err := integrations.Get(integrationId)
	if err != nil {
		return integrations.AuthResult{}, err
	}
	return integration.Authenticate(credentials)
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
// Applies type mappings, auto-matches by path/name, returns only NEW items.
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

	// Load type mappings from sync_options
	var syncOptions integrations.SyncOptions
	if integrationProject.SyncOptions != "" && integrationProject.SyncOptions != "{}" {
		json.Unmarshal([]byte(integrationProject.SyncOptions), &syncOptions)
	}
	if syncOptions.EntityTypeMappings == nil {
		syncOptions.EntityTypeMappings = make(map[string]integrations.TypeMapping)
	}
	if syncOptions.TaskTypeMappings == nil {
		syncOptions.TaskTypeMappings = make(map[string]integrations.TypeMapping)
	}

	// Apply default directory structure if none configured
	if len(syncOptions.DirectoryStructure.Paths) == 0 {
		syncOptions.DirectoryStructure = integrations.DirectoryStructure{
			Preset: "3d-animation",
			Style:  "lowercase",
			Paths: map[string]interface{}{
				"asset": map[string]interface{}{
					"name":     "Assets",
					"icon":     "package",
					"template": "Assets/<CollectionType>/<Asset>",
				},
				"shot": map[string]interface{}{
					"name":     "Shots",
					"icon":     "clapperboard",
					"template": "Episodes/<Episode>/<Sequence>/<Shot>",
				},
			},
		}
	}

	// Fetch external hierarchy
	externalEntities, err := integration.GetProjectEntities(token, integrationProject.ApiUrl, integrationProject.ExternalProjectId)
	if err != nil {
		return integrations.SyncPreview{}, err
	}

	externalTasks, err := integration.GetProjectTasks(token, integrationProject.ApiUrl, integrationProject.ExternalProjectId)
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

	// Load templates for extension lookup
	templates, err := repository.GetTemplates(tx, false)
	if err != nil {
		return integrations.SyncPreview{}, err
	}
	templateByID := make(map[string]models.Template)
	for _, t := range templates {
		templateByID[t.Id] = t
	}

	// Build lookup maps for existing mappings (by external ID)
	existingCollectionMap := make(map[string]models.IntegrationCollectionMapping)
	for _, m := range existingCollections {
		existingCollectionMap[m.ExternalId] = m
	}
	existingAssetMap := make(map[string]models.IntegrationAssetMapping)
	for _, m := range existingAssets {
		existingAssetMap[m.ExternalId] = m
	}

	// Build entity ID → entity map for path building
	entityByID := make(map[string]integrations.ExternalEntity)
	for _, e := range externalEntities {
		entityByID[e.ID] = e
	}

	// Build full paths for all external entities using DirectoryStructure templates
	entityPaths := make(map[string]string) // external ID → full path
	for _, entity := range externalEntities {
		path := resolveEntityPath(entity, entityByID, syncOptions.DirectoryStructure)
		entityPaths[entity.ID] = path
	}

	// Build sync preview - only includes items to CREATE
	preview := integrations.SyncPreview{
		IntegrationID: integrationProject.IntegrationId,
		Collections:   []integrations.SyncCollection{},
		Assets:        []integrations.SyncAsset{},
		MissingTypes:  []integrations.MissingType{},
	}

	// Track missing types
	missingEntityTypes := make(map[string]bool)
	missingTaskTypes := make(map[string]bool)

	// Process entities (collections) - only add if needs to be created
	for _, entity := range externalEntities {
		// Skip if already mapped
		if _, exists := existingCollectionMap[entity.ID]; exists {
			continue
		}

		fullPath := entityPaths[entity.ID]

		// Try to auto-match by path in Clustta
		existingEntity, err := repository.GetEntityByPath(tx, fullPath)
		if err == nil && existingEntity.Id != "" {
			// Found matching collection - skip (not a new item)
			continue
		}

		// Get type mapping
		typeMapping, hasMaping := syncOptions.EntityTypeMappings[entity.Type]
		entityTypeName := ""
		entityTypeIcon := ""
		if hasMaping {
			entityTypeName = typeMapping.ClustttaName
			entityTypeIcon = typeMapping.ClustttaIcon
		} else {
			// Track missing type
			missingEntityTypes[entity.Type] = true
		}

		// This is a new collection to create
		preview.Collections = append(preview.Collections, integrations.SyncCollection{
			TempID:           entity.ID,
			ExternalID:       entity.ID,
			ExternalType:     entity.Type,
			ExternalName:     entity.Name,
			ExternalParentID: entity.ParentID,
			ExternalPath:     fullPath,
			CollectionPath:   fullPath,
			Action:           "create",
			EntityTypeName:   entityTypeName,
			EntityTypeIcon:   entityTypeIcon,
			Selected:         true,
		})
	}

	// Process tasks (assets) - only add if needs to be created
	for _, task := range externalTasks {
		// Skip if already mapped
		if _, exists := existingAssetMap[task.ID]; exists {
			continue
		}

		// Get parent entity path
		parentPath := ""
		if parentEntity, exists := entityByID[task.ParentID]; exists {
			parentPath = entityPaths[parentEntity.ID]
		}

		// Try to auto-match by name+parent in Clustta
		if parentPath != "" {
			parentCollection, err := repository.GetEntityByPath(tx, parentPath)
			if err == nil && parentCollection.Id != "" {
				// Check if asset exists in this collection
				existingTask, err := repository.GetTaskByName(tx, task.Name, parentCollection.Id, "")
				if err == nil && existingTask.Id != "" {
					// Found matching asset - skip (not a new item)
					continue
				}
			}
		}

		// Get type mapping
		typeMapping, hasMapping := syncOptions.TaskTypeMappings[task.TaskType]
		taskTypeName := ""
		taskTypeIcon := ""
		if hasMapping {
			taskTypeName = typeMapping.ClustttaName
			taskTypeIcon = typeMapping.ClustttaIcon
		} else if task.TaskType != "" {
			// Track missing type
			missingTaskTypes[task.TaskType] = true
		}

		// Get template mapping
		templateID := ""
		templateExtension := ""
		if syncOptions.TaskTypeTemplates != nil {
			if tmplID, ok := syncOptions.TaskTypeTemplates[task.TaskTypeID]; ok {
				templateID = tmplID
				if tmpl, exists := templateByID[tmplID]; exists {
					templateExtension = tmpl.Extension
				}
			}
		}

		// This is a new asset to create
		preview.Assets = append(preview.Assets, integrations.SyncAsset{
			TempID:            task.ID,
			ExternalID:        task.ID,
			ExternalName:      task.Name,
			ExternalParentID:  task.ParentID,
			ExternalType:      task.TaskType,
			ExternalTypeID:    task.TaskTypeID,
			ExternalStatus:    task.Status,
			ExternalAssignees: task.Assignees,
			CollectionPath:    parentPath,
			Action:            "create",
			TaskTypeName:      taskTypeName,
			TaskTypeIcon:      taskTypeIcon,
			Selected:          true,
			TemplateID:        templateID,
			TemplateExtension: templateExtension,
		})
	}

	// Build missing types list
	entityIcons := []string{"folder", "episode", "sequence", "shot", "character", "prop", "environment", "scene"}
	taskIcons := []string{"animation", "lighting", "compositing", "modeling", "rigging", "texturing", "fx", "rendering", "concept art", "layout"}

	i := 0
	for typeName := range missingEntityTypes {
		preview.MissingTypes = append(preview.MissingTypes, integrations.MissingType{
			ExternalName:  typeName,
			ExternalID:    typeName,
			TypeCategory:  "entity",
			SuggestedName: strings.ToLower(strings.TrimSpace(typeName)),
			SuggestedIcon: entityIcons[i%len(entityIcons)],
		})
		i++
	}

	j := 0
	for typeName := range missingTaskTypes {
		preview.MissingTypes = append(preview.MissingTypes, integrations.MissingType{
			ExternalName:  typeName,
			ExternalID:    typeName,
			TypeCategory:  "task",
			SuggestedName: strings.ToLower(strings.TrimSpace(typeName)),
			SuggestedIcon: taskIcons[j%len(taskIcons)],
		})
		j++
	}

	// Build summary
	preview.Summary = integrations.SyncPreviewSummary{
		TotalCollections:    len(externalEntities),
		TotalAssets:         len(externalTasks),
		CollectionsToCreate: len(preview.Collections),
		AssetsToCreate:      len(preview.Assets),
	}

	// Generate unified PreviewItems from collections and assets
	preview.PreviewItems = buildPreviewItems(preview.Collections, preview.Assets)

	return preview, nil
}

// resolveEntityPath resolves the path for an entity using DirectoryStructure templates.
// Falls back to raw hierarchy path if no matching template is found.
func resolveEntityPath(entity integrations.ExternalEntity, entityByID map[string]integrations.ExternalEntity, dirStructure integrations.DirectoryStructure) string {
	// Find matching template based on entity type
	template := findMatchingTemplate(entity.Type, dirStructure)
	if template == "" {
		// No matching template - use raw hierarchy
		fallback := buildExternalEntityPath(entity, entityByID)
		return normalizeCollectionPath(fallback)
	}

	// Resolve template variables
	resolved := resolveTemplateVariables(template, entity, entityByID, dirStructure.Style)
	return normalizeCollectionPath(resolved)
}

// normalizeCollectionPath ensures the path is lowercase and has leading/trailing slashes.
// Example: "Episodes/EP01/Sequences" -> "/episodes/ep01/sequences/"
func normalizeCollectionPath(path string) string {
	// Convert to lowercase
	path = strings.ToLower(path)

	// Add leading slash if missing
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Add trailing slash if missing
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	return path
}

// findMatchingTemplate finds a template that matches the given entity type.
// Templates match based on the variable names they contain.
// Uses frontend format: <Episode>, <Sequence>, <Shot>, <Asset>, <CollectionType>
func findMatchingTemplate(entityType string, dirStructure integrations.DirectoryStructure) string {
	if dirStructure.Paths == nil {
		return ""
	}

	entityTypeLower := strings.ToLower(entityType)

	// Standard entity types and their variable names (matching frontend format)
	standardTypes := map[string]string{
		"episode":  "<Episode>",
		"sequence": "<Sequence>",
		"shot":     "<Shot>",
	}

	// Check if this is a standard type
	if varName, ok := standardTypes[entityTypeLower]; ok {
		// Find template containing this variable
		for _, pathData := range dirStructure.Paths {
			if data, ok := pathData.(map[string]interface{}); ok {
				if tmpl, ok := data["template"].(string); ok {
					if strings.Contains(tmpl, varName) {
						return tmpl
					}
				}
			}
		}
		return ""
	}

	// For assets (non-standard types like "Character", "Prop"), find template with <Asset>
	for _, pathData := range dirStructure.Paths {
		if data, ok := pathData.(map[string]interface{}); ok {
			if tmpl, ok := data["template"].(string); ok {
				if strings.Contains(tmpl, "<Asset>") {
					return tmpl
				}
			}
		}
	}

	return ""
}

// resolveTemplateVariables resolves variables in a template path.
// Variables like <Episode>, <Sequence>, <Shot>, <Asset>, <CollectionType> are replaced with actual values.
// The template is truncated at the entity's level to avoid unresolved variables.
func resolveTemplateVariables(template string, entity integrations.ExternalEntity, entityByID map[string]integrations.ExternalEntity, style string) string {
	entityTypeLower := strings.ToLower(entity.Type)

	// Standard variable mappings by entity type (matching frontend format)
	typeToVar := map[string]string{
		"episode":  "<Episode>",
		"sequence": "<Sequence>",
		"shot":     "<Shot>",
	}

	// Truncate template at the entity's variable level
	// For a sequence, "Episodes/<Episode>/<Sequence>/<Shot>" becomes "Episodes/<Episode>/<Sequence>"
	if varName, ok := typeToVar[entityTypeLower]; ok {
		// Find the position of this entity's variable and truncate after it
		idx := strings.Index(template, varName)
		if idx != -1 {
			template = template[:idx+len(varName)]
		}
	} else {
		// For assets (non-standard types), truncate at <Asset>
		idx := strings.Index(template, "<Asset>")
		if idx != -1 {
			template = template[:idx+len("<Asset>")]
		}
	}

	result := template

	// Build entity hierarchy for variable resolution
	hierarchy := buildEntityHierarchy(entity, entityByID)

	// Resolve standard type variables from hierarchy
	for _, e := range hierarchy {
		typeLower := strings.ToLower(e.Type)
		if varName, ok := typeToVar[typeLower]; ok {
			result = strings.ReplaceAll(result, varName, applyNamingStyle(e.Name, style))
		}
	}

	// For assets, resolve <Asset> and <CollectionType>
	if entityTypeLower != "episode" && entityTypeLower != "sequence" && entityTypeLower != "shot" {
		// This is an asset - resolve <Asset> with entity name
		result = strings.ReplaceAll(result, "<Asset>", applyNamingStyle(entity.Name, style))
		// Resolve <CollectionType> with entity type (like "Character", "Prop")
		result = strings.ReplaceAll(result, "<CollectionType>", applyNamingStyle(entity.Type, style))
	}

	// Remove any path segments that still contain unresolved variables
	// This handles cases where parent entities are missing (e.g., sequence without episode)
	segments := strings.Split(result, "/")
	var cleanedSegments []string
	for _, seg := range segments {
		if !strings.Contains(seg, "<") {
			cleanedSegments = append(cleanedSegments, seg)
		}
	}
	result = strings.Join(cleanedSegments, "/")

	return result
}

// buildEntityHierarchy returns the entity and all its ancestors from leaf to root.
func buildEntityHierarchy(entity integrations.ExternalEntity, entityByID map[string]integrations.ExternalEntity) []integrations.ExternalEntity {
	var hierarchy []integrations.ExternalEntity
	current := entity

	for {
		hierarchy = append(hierarchy, current)
		if current.ParentID == "" {
			break
		}
		parent, exists := entityByID[current.ParentID]
		if !exists {
			break
		}
		current = parent
	}

	return hierarchy
}

// applyNamingStyle applies the configured naming style to a string.
func applyNamingStyle(name, style string) string {
	switch style {
	case "lowercase":
		return strings.ToLower(name)
	case "uppercase":
		return strings.ToUpper(name)
	case "kebab-case":
		// Convert spaces and underscores to hyphens, lowercase
		result := strings.ReplaceAll(name, " ", "-")
		result = strings.ReplaceAll(result, "_", "-")
		return strings.ToLower(result)
	case "capitalize":
		// Capitalize first letter of each word
		return strings.Title(strings.ToLower(name))
	default:
		return name
	}
}

// buildExternalEntityPath builds the full path for an external entity.
func buildExternalEntityPath(entity integrations.ExternalEntity, entityByID map[string]integrations.ExternalEntity) string {
	var parts []string
	current := entity

	for {
		parts = append([]string{current.Name}, parts...)
		if current.ParentID == "" {
			break
		}
		parent, exists := entityByID[current.ParentID]
		if !exists {
			break
		}
		current = parent
	}

	return strings.Join(parts, "/")
}

// buildPreviewItems generates a unified list of PreviewItems from collections and assets.
// It also creates virtual folder items for any path segments that don't have real items.
func buildPreviewItems(collections []integrations.SyncCollection, assets []integrations.SyncAsset) []integrations.PreviewItem {
	items := make([]integrations.PreviewItem, 0)
	pathToItem := make(map[string]bool) // Track which paths have real items
	allPaths := make(map[string]bool)   // Track all paths including parent segments

	// First pass: convert collections to PreviewItems and track paths
	for _, c := range collections {
		parentPath := getParentPath(c.CollectionPath)
		items = append(items, integrations.PreviewItem{
			ID:             c.ExternalID,
			Name:           c.ExternalName,
			ItemType:       "entity",
			CollectionPath: c.CollectionPath,
			ParentPath:     parentPath,
			ExternalID:     c.ExternalID,
			ExternalType:   c.ExternalType,
			ExternalName:   c.ExternalName,
			TypeName:       c.EntityTypeName,
			TypeIcon:       c.EntityTypeIcon,
			Action:         c.Action,
			Selected:       c.Selected,
			IsVirtual:      false,
			HasChildren:    false, // Will be updated below
		})
		pathToItem[c.CollectionPath] = true
		allPaths[c.CollectionPath] = true

		// Track all parent paths
		addParentPaths(c.CollectionPath, allPaths)
	}

	// Second pass: convert assets to PreviewItems
	for _, a := range assets {
		// Assets use their parent's collection_path as their parent
		// Their own "path" is parent + /asset-{id}/
		assetPath := a.CollectionPath + "asset-" + a.ExternalID + "/"
		items = append(items, integrations.PreviewItem{
			ID:                a.ExternalID,
			Name:              a.ExternalName,
			ItemType:          "task",
			CollectionPath:    assetPath,
			ParentPath:        a.CollectionPath,
			ExternalID:        a.ExternalID,
			ExternalType:      a.ExternalType,
			ExternalTypeID:    a.ExternalTypeID,
			ExternalName:      a.ExternalName,
			TypeName:          a.TaskTypeName,
			TypeIcon:          a.TaskTypeIcon,
			Action:            a.Action,
			Selected:          a.Selected,
			IsVirtual:         false,
			HasChildren:       false,
			TemplateID:        a.TemplateID,
			TemplateExtension: a.TemplateExtension,
		})
		pathToItem[assetPath] = true // Mark asset path as having a real item
		allPaths[assetPath] = true

		// Track parent paths of the asset's collection
		addParentPaths(a.CollectionPath, allPaths)
	}

	// Third pass: create virtual folder items for path segments without real items
	virtualFolders := make([]integrations.PreviewItem, 0)
	for path := range allPaths {
		if pathToItem[path] {
			continue // Real item exists at this path
		}
		// This is a path segment that needs a virtual folder
		name := getPathSegmentName(path)
		if name == "" {
			continue // Root path, skip
		}
		parentPath := getParentPath(path)
		virtualFolders = append(virtualFolders, integrations.PreviewItem{
			ID:             "virtual-" + path,
			Name:           name,
			ItemType:       "virtual",
			CollectionPath: path,
			ParentPath:     parentPath,
			ExternalID:     "",
			ExternalType:   "folder",
			ExternalName:   name,
			TypeName:       "Folder",
			TypeIcon:       "folder",
			Action:         "virtual",
			Selected:       false,
			IsVirtual:      true,
			HasChildren:    true,
		})
	}
	items = append(items, virtualFolders...)

	// Fourth pass: update HasChildren flag for all items
	childCount := make(map[string]int)
	for _, item := range items {
		if item.ParentPath != "" && item.ParentPath != "/" {
			childCount[item.ParentPath]++
		}
	}
	for i := range items {
		if childCount[items[i].CollectionPath] > 0 {
			items[i].HasChildren = true
		}
	}

	return items
}

// addParentPaths adds all parent path segments to the map.
func addParentPaths(path string, allPaths map[string]bool) {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	currentPath := "/"
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		currentPath = currentPath + seg + "/"
		allPaths[currentPath] = true
	}
}

// getParentPath returns the parent path of a collection_path.
// Example: "/episodes/ep01/seq01/" -> "/episodes/ep01/"
func getParentPath(path string) string {
	path = strings.TrimSuffix(path, "/")
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash <= 0 {
		return "/"
	}
	return path[:lastSlash] + "/"
}

// getPathSegmentName returns the last segment name from a path.
// Example: "/episodes/ep01/" -> "ep01"
func getPathSegmentName(path string) string {
	path = strings.TrimSuffix(path, "/")
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		return path
	}
	return path[lastSlash+1:]
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
// ═══════════════════════════════════════════════════════════════════════════
// TYPE MAPPING
// ═══════════════════════════════════════════════════════════════════════════

// GetTypeMappings retrieves the type mappings from sync_options for a project.
// Returns empty SyncOptions if no mappings configured yet.
func (s *IntegrationService) GetTypeMappings(projectPath string) (integrations.SyncOptions, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return integrations.SyncOptions{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return integrations.SyncOptions{}, err
	}
	defer tx.Rollback()

	integrationProject, err := repository.GetIntegrationProject(tx)
	if err != nil {
		return integrations.SyncOptions{}, errors.New("no integration linked to this project")
	}

	var syncOptions integrations.SyncOptions
	if integrationProject.SyncOptions != "" && integrationProject.SyncOptions != "{}" {
		if err := json.Unmarshal([]byte(integrationProject.SyncOptions), &syncOptions); err != nil {
			// Return empty options if parse fails
			return integrations.SyncOptions{
				EntityTypeMappings: make(map[string]integrations.TypeMapping),
				TaskTypeMappings:   make(map[string]integrations.TypeMapping),
			}, nil
		}
	}

	// Initialize maps if nil
	if syncOptions.EntityTypeMappings == nil {
		syncOptions.EntityTypeMappings = make(map[string]integrations.TypeMapping)
	}
	if syncOptions.TaskTypeMappings == nil {
		syncOptions.TaskTypeMappings = make(map[string]integrations.TypeMapping)
	}
	if syncOptions.TaskTypeTemplates == nil {
		syncOptions.TaskTypeTemplates = make(map[string]string)
	}

	return syncOptions, nil
}

// SaveTypeMappings saves type mappings to sync_options for a project.
func (s *IntegrationService) SaveTypeMappings(projectPath string, syncOptions integrations.SyncOptions) error {
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

	integrationProject, err := repository.GetIntegrationProject(tx)
	if err != nil {
		return errors.New("no integration linked to this project")
	}

	syncOptionsJSON, err := json.Marshal(syncOptions)
	if err != nil {
		return err
	}

	err = repository.UpdateIntegrationProject(tx, integrationProject.Id, map[string]interface{}{
		"sync_options": string(syncOptionsJSON),
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetExternalTypes fetches entity and task types from the external integration.
// Requires valid token and linked integration.
func (s *IntegrationService) GetExternalTypes(projectPath, token string) ([]integrations.ExternalTypeInfo, []integrations.ExternalTypeInfo, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return nil, nil, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	integrationProject, err := repository.GetIntegrationProject(tx)
	if err != nil {
		return nil, nil, errors.New("no integration linked to this project")
	}

	integration, err := integrations.Get(integrationProject.IntegrationId)
	if err != nil {
		return nil, nil, err
	}

	entityTypes, err := integration.GetEntityTypes(token, integrationProject.ApiUrl, integrationProject.ExternalProjectId)
	if err != nil {
		return nil, nil, err
	}

	taskTypes, err := integration.GetTaskTypes(token, integrationProject.ApiUrl, integrationProject.ExternalProjectId)
	if err != nil {
		return nil, nil, err
	}

	return entityTypes, taskTypes, nil
}

// GetMissingTypes compares external types with local Clustta types.
// Returns types that don't have a mapping in sync_options and don't exist in Clustta.
func (s *IntegrationService) GetMissingTypes(projectPath, token string) ([]integrations.MissingType, error) {
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

	integrationProject, err := repository.GetIntegrationProject(tx)
	if err != nil {
		return nil, errors.New("no integration linked to this project")
	}

	// Get integration client
	integration, err := integrations.Get(integrationProject.IntegrationId)
	if err != nil {
		return nil, err
	}

	// Get current type mappings
	var syncOptions integrations.SyncOptions
	if integrationProject.SyncOptions != "" && integrationProject.SyncOptions != "{}" {
		json.Unmarshal([]byte(integrationProject.SyncOptions), &syncOptions)
	}
	if syncOptions.EntityTypeMappings == nil {
		syncOptions.EntityTypeMappings = make(map[string]integrations.TypeMapping)
	}
	if syncOptions.TaskTypeMappings == nil {
		syncOptions.TaskTypeMappings = make(map[string]integrations.TypeMapping)
	}

	// Get local types
	localEntityTypes, err := repository.GetEntityTypes(tx)
	if err != nil {
		return nil, err
	}
	localTaskTypes, err := repository.GetTaskTypes(tx)
	if err != nil {
		return nil, err
	}

	// Build lookup maps for local types (by lowercase name)
	localEntityTypeMap := make(map[string]models.EntityType)
	for _, et := range localEntityTypes {
		localEntityTypeMap[strings.ToLower(et.Name)] = et
	}
	localTaskTypeMap := make(map[string]models.TaskType)
	for _, tt := range localTaskTypes {
		localTaskTypeMap[strings.ToLower(tt.Name)] = tt
	}

	// Get external types
	externalEntityTypes, err := integration.GetEntityTypes(token, integrationProject.ApiUrl, integrationProject.ExternalProjectId)
	if err != nil {
		return nil, err
	}
	externalTaskTypes, err := integration.GetTaskTypes(token, integrationProject.ApiUrl, integrationProject.ExternalProjectId)
	if err != nil {
		return nil, err
	}

	missingTypes := []integrations.MissingType{}

	// Available icons for random selection
	entityIcons := []string{"folder", "episode", "sequence", "shot", "character", "prop", "environment", "scene"}
	taskIcons := []string{"animation", "lighting", "compositing", "modeling", "rigging", "texturing", "fx", "rendering", "concept art", "layout"}

	// Check entity types
	for i, et := range externalEntityTypes {
		// Skip if already mapped
		if _, exists := syncOptions.EntityTypeMappings[et.Name]; exists {
			continue
		}

		// Check if matching local type exists (case-insensitive)
		suggestedName := strings.ToLower(strings.TrimSpace(et.Name))
		if localType, exists := localEntityTypeMap[suggestedName]; exists {
			// Auto-map to existing type
			syncOptions.EntityTypeMappings[et.Name] = integrations.TypeMapping{
				ExternalName:   et.Name,
				ExternalID:     et.ID,
				ClustttaTypeID: localType.Id,
				ClustttaName:   localType.Name,
				ClustttaIcon:   localType.Icon,
			}
			continue
		}

		// Type is missing - add to list
		missingTypes = append(missingTypes, integrations.MissingType{
			ExternalName:  et.Name,
			ExternalID:    et.ID,
			TypeCategory:  "entity",
			SuggestedName: suggestedName,
			SuggestedIcon: entityIcons[i%len(entityIcons)],
		})
	}

	// Check task types
	for i, tt := range externalTaskTypes {
		// Skip if already mapped
		if _, exists := syncOptions.TaskTypeMappings[tt.Name]; exists {
			continue
		}

		// Check if matching local type exists (case-insensitive)
		suggestedName := strings.ToLower(strings.TrimSpace(tt.Name))
		if localType, exists := localTaskTypeMap[suggestedName]; exists {
			// Auto-map to existing type
			syncOptions.TaskTypeMappings[tt.Name] = integrations.TypeMapping{
				ExternalName:   tt.Name,
				ExternalID:     tt.ID,
				ClustttaTypeID: localType.Id,
				ClustttaName:   localType.Name,
				ClustttaIcon:   localType.Icon,
			}
			continue
		}

		// Type is missing - add to list
		missingTypes = append(missingTypes, integrations.MissingType{
			ExternalName:  tt.Name,
			ExternalID:    tt.ID,
			TypeCategory:  "task",
			SuggestedName: suggestedName,
			SuggestedIcon: taskIcons[i%len(taskIcons)],
		})
	}

	// Save auto-mapped types if any
	if len(syncOptions.EntityTypeMappings) > 0 || len(syncOptions.TaskTypeMappings) > 0 {
		syncOptionsJSON, _ := json.Marshal(syncOptions)
		repository.UpdateIntegrationProject(tx, integrationProject.Id, map[string]interface{}{
			"sync_options": string(syncOptionsJSON),
		})
		tx.Commit()
	}

	return missingTypes, nil
}

// CreateMissingTypes creates entity and task types in Clustta for missing external types.
// Updates sync_options with the new mappings.
func (s *IntegrationService) CreateMissingTypes(projectPath string, missingTypes []integrations.MissingType) error {
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

	integrationProject, err := repository.GetIntegrationProject(tx)
	if err != nil {
		return errors.New("no integration linked to this project")
	}

	// Get current sync options
	var syncOptions integrations.SyncOptions
	if integrationProject.SyncOptions != "" && integrationProject.SyncOptions != "{}" {
		json.Unmarshal([]byte(integrationProject.SyncOptions), &syncOptions)
	}
	if syncOptions.EntityTypeMappings == nil {
		syncOptions.EntityTypeMappings = make(map[string]integrations.TypeMapping)
	}
	if syncOptions.TaskTypeMappings == nil {
		syncOptions.TaskTypeMappings = make(map[string]integrations.TypeMapping)
	}

	// Create each missing type
	for _, mt := range missingTypes {
		if mt.TypeCategory == "entity" {
			// Create entity type
			entityType, err := repository.GetOrCreateEntityType(tx, mt.SuggestedName, mt.SuggestedIcon)
			if err != nil {
				return err
			}

			// Add to mappings
			syncOptions.EntityTypeMappings[mt.ExternalName] = integrations.TypeMapping{
				ExternalName:   mt.ExternalName,
				ExternalID:     mt.ExternalID,
				ClustttaTypeID: entityType.Id,
				ClustttaName:   entityType.Name,
				ClustttaIcon:   entityType.Icon,
			}
		} else if mt.TypeCategory == "task" {
			// Create task type
			taskType, err := repository.GetOrCreateTaskType(tx, mt.SuggestedName, mt.SuggestedIcon)
			if err != nil {
				return err
			}

			// Add to mappings
			syncOptions.TaskTypeMappings[mt.ExternalName] = integrations.TypeMapping{
				ExternalName:   mt.ExternalName,
				ExternalID:     mt.ExternalID,
				ClustttaTypeID: taskType.Id,
				ClustttaName:   taskType.Name,
				ClustttaIcon:   taskType.Icon,
			}
		}
	}

	// Save updated sync options
	syncOptionsJSON, err := json.Marshal(syncOptions)
	if err != nil {
		return err
	}

	err = repository.UpdateIntegrationProject(tx, integrationProject.Id, map[string]interface{}{
		"sync_options": string(syncOptionsJSON),
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetLocalTypes returns all entity and task types from the project.
// Used by frontend to populate mapping dropdowns.
func (s *IntegrationService) GetLocalTypes(projectPath string) ([]models.EntityType, []models.TaskType, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return nil, nil, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	entityTypes, err := repository.GetEntityTypes(tx)
	if err != nil {
		return nil, nil, err
	}

	taskTypes, err := repository.GetTaskTypes(tx)
	if err != nil {
		return nil, nil, err
	}

	return entityTypes, taskTypes, nil
}
