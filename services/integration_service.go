package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/constants"
	error_service "clustta/internal/error_service"
	"clustta/internal/integrations"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/settings"
	"clustta/internal/utils"
	"clustta/output"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/wailsapp/wails/v3/pkg/application"
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
// Also deletes all collection and asset mappings, and removes the stored
// integration credential for this user so the unlink fully revokes access.
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

	if err := tx.Commit(); err != nil {
		return err
	}

	// Best-effort credential cleanup. Credentials are user-scoped and shared
	// across any other project this user has linked to the same integration;
	// the user accepted that trade-off in the unlink confirmation dialog.
	_ = settings.DeleteIntegrationCredential(integration.IntegrationId)

	return nil
}

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
	if syncOptions.CollectionTypeMappings == nil {
		syncOptions.CollectionTypeMappings = make(map[string]integrations.TypeMapping)
	}
	if syncOptions.AssetTypeMappings == nil {
		syncOptions.AssetTypeMappings = make(map[string]integrations.TypeMapping)
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
					"template": "Assets/<CollectionType>/<Asset>/<AssetType><TemplateExtension>",
				},
				"shot": map[string]interface{}{
					"name":     "Shots",
					"icon":     "clapperboard",
					"template": "Episodes/<Episode>/<Sequence>/<Shot>/<AssetType><TemplateExtension>",
				},
			},
		}
	}

	// Fetch external hierarchy
	externalCollections, err := integration.GetProjectCollections(token, integrationProject.ApiUrl, integrationProject.ExternalProjectId)
	if err != nil {
		return integrations.SyncPreview{}, err
	}

	externalAssets, err := integration.GetProjectAssets(token, integrationProject.ApiUrl, integrationProject.ExternalProjectId)
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

	// Build collection ID → collection map for path building
	collectionByID := make(map[string]integrations.ExternalCollection)
	for _, e := range externalCollections {
		collectionByID[e.ID] = e
	}

	// Build full paths for all external collections using DirectoryStructure templates
	collectionPaths := make(map[string]string) // external ID → full path
	for _, collection := range externalCollections {
		path := resolveCollectionPath(collection, collectionByID, syncOptions.DirectoryStructure)
		collectionPaths[collection.ID] = path
	}

	// Build sync preview - only includes items to CREATE
	preview := integrations.SyncPreview{
		IntegrationID: integrationProject.IntegrationId,
		Collections:   []integrations.SyncCollection{},
		Assets:        []integrations.SyncAsset{},
		MissingTypes:  []integrations.MissingType{},
	}

	// Track missing types
	missingCollectionTypes := make(map[string]bool)
	missingAssetTypes := make(map[string]bool)

	// Process collections (collections) - only add if needs to be created
	for _, collection := range externalCollections {
		// Skip virtual folder collections (asset type containers) - they're only for path building
		if strings.HasPrefix(collection.ID, "asset-type-") {
			continue
		}

		// Skip if already mapped
		if _, exists := existingCollectionMap[collection.ID]; exists {
			continue
		}

		fullPath := collectionPaths[collection.ID]

		// Try to auto-match by path in Clustta
		existingCollection, err := repository.GetCollectionByPath(tx, fullPath)
		if err == nil && existingCollection.Id != "" {
			// Found matching collection - skip (not a new item)
			continue
		}

		// Get type mapping
		typeMapping, hasMaping := syncOptions.CollectionTypeMappings[collection.Type]
		collectionTypeName := ""
		collectionTypeIcon := ""
		if hasMaping {
			collectionTypeName = typeMapping.ClustttaName
			collectionTypeIcon = typeMapping.ClustttaIcon
		} else {
			// Track missing type
			missingCollectionTypes[collection.Type] = true
		}

		// This is a new collection to create
		preview.Collections = append(preview.Collections, integrations.SyncCollection{
			TempID:             collection.ID,
			ExternalID:         collection.ID,
			ExternalType:       collection.Type,
			ExternalName:       collection.Name,
			ExternalParentID:   collection.ParentID,
			ExternalPath:       fullPath,
			CollectionPath:     fullPath,
			Action:             "create",
			CollectionTypeName: collectionTypeName,
			CollectionTypeIcon: collectionTypeIcon,
			Selected:           true,
		})
	}

	// Process assets (assets) - only add if needs to be created
	for _, asset := range externalAssets {
		// Skip if already mapped
		if _, exists := existingAssetMap[asset.ID]; exists {
			continue
		}

		// Get parent collection path
		parentPath := ""
		if parentCollection, exists := collectionByID[asset.ParentID]; exists {
			parentPath = collectionPaths[parentCollection.ID]
		}

		// Get template mapping
		templateID := ""
		templateExtension := ""
		if syncOptions.AssetTypeTemplates != nil {
			if tmplID, ok := syncOptions.AssetTypeTemplates[asset.AssetTypeID]; ok {
				templateID = tmplID
				if tmpl, exists := templateByID[tmplID]; exists {
					templateExtension = tmpl.Extension
				}
			}
		}

		assetName := asset.Name
		if parentCollection, exists := collectionByID[asset.ParentID]; exists {
			assetName = resolveAssetNameFromTemplate(assetName, parentCollection, asset, collectionByID, syncOptions.DirectoryStructure, templateExtension)
		}

		// Try to auto-match by name+parent in Clustta
		if parentPath != "" {
			parentCollection, err := repository.GetCollectionByPath(tx, parentPath)
			if err == nil && parentCollection.Id != "" {
				// Check if asset exists in this collection
				existingAsset, err := repository.GetAssetByName(tx, assetName, parentCollection.Id, templateExtension)
				if err == nil && existingAsset.Id != "" {
					// Found matching asset - skip (not a new item)
					continue
				}
			}
		}

		// Get type mapping
		typeMapping, hasMapping := syncOptions.AssetTypeMappings[asset.AssetType]
		assetTypeName := ""
		assetTypeIcon := ""
		if hasMapping {
			assetTypeName = typeMapping.ClustttaName
			assetTypeIcon = typeMapping.ClustttaIcon
		} else if asset.AssetType != "" {
			// Track missing type
			missingAssetTypes[asset.AssetType] = true
		}

		// This is a new asset to create
		preview.Assets = append(preview.Assets, integrations.SyncAsset{
			TempID:            asset.ID,
			ExternalID:        asset.ID,
			ExternalName:      assetName,
			ExternalParentID:  asset.ParentID,
			ExternalType:      asset.AssetType,
			ExternalTypeID:    asset.AssetTypeID,
			ExternalStatus:    asset.Status,
			ExternalAssignees: asset.Assignees,
			CollectionPath:    parentPath,
			Action:            "create",
			AssetTypeName:     assetTypeName,
			AssetTypeIcon:     assetTypeIcon,
			Selected:          true,
			TemplateID:        templateID,
			TemplateExtension: templateExtension,
		})
	}

	// Build missing types list
	collectionIcons := constants.CollectionTypeIcons
	assetIcons := constants.AssetTypeIcons

	i := 0
	for typeName := range missingCollectionTypes {
		preview.MissingTypes = append(preview.MissingTypes, integrations.MissingType{
			ExternalName:  typeName,
			ExternalID:    typeName,
			TypeCategory:  "collection",
			SuggestedName: strings.ToLower(strings.TrimSpace(typeName)),
			SuggestedIcon: collectionIcons[i%len(collectionIcons)],
		})
		i++
	}

	j := 0
	for typeName := range missingAssetTypes {
		preview.MissingTypes = append(preview.MissingTypes, integrations.MissingType{
			ExternalName:  typeName,
			ExternalID:    typeName,
			TypeCategory:  "asset",
			SuggestedName: strings.ToLower(strings.TrimSpace(typeName)),
			SuggestedIcon: assetIcons[j%len(assetIcons)],
		})
		j++
	}

	// Build summary
	preview.Summary = integrations.SyncPreviewSummary{
		TotalCollections:    len(externalCollections),
		TotalAssets:         len(externalAssets),
		CollectionsToCreate: len(preview.Collections),
		AssetsToCreate:      len(preview.Assets),
	}

	// Generate unified PreviewItems from collections and assets
	preview.PreviewItems = buildPreviewItems(preview.Collections, preview.Assets)

	return preview, nil
}

// resolveCollectionPath resolves the path for an collection using DirectoryStructure templates.
// Falls back to raw hierarchy path if no matching template is found.
func resolveCollectionPath(collection integrations.ExternalCollection, collectionByID map[string]integrations.ExternalCollection, dirStructure integrations.DirectoryStructure) string {
	// Find matching template based on collection type
	template := findMatchingTemplate(collection.Type, dirStructure)
	if template == "" {
		// No matching template - use raw hierarchy
		fallback := buildExternalCollectionPath(collection, collectionByID)
		return normalizeCollectionPath(fallback)
	}

	// Resolve template variables
	resolved := resolveTemplateVariables(template, collection, collectionByID, dirStructure.Style)
	return normalizeCollectionPath(resolved)
}

// resolveAssetNameFromTemplate returns the filename stem implied by the asset directory template.
func resolveAssetNameFromTemplate(fallback string, parentCollection integrations.ExternalCollection, asset integrations.ExternalAsset, collectionByID map[string]integrations.ExternalCollection, dirStructure integrations.DirectoryStructure, templateExtension string) string {
	template := findMatchingTemplate(parentCollection.Type, dirStructure)
	if template == "" {
		return fallback
	}

	resolved := resolveAssetTemplateVariables(template, parentCollection, asset, collectionByID, dirStructure.Style, templateExtension)
	resolved = strings.Trim(strings.TrimSpace(resolved), "/")
	if resolved == "" {
		return fallback
	}

	segments := strings.Split(resolved, "/")
	fileName := strings.TrimSpace(segments[len(segments)-1])
	if fileName == "" || strings.Contains(fileName, "<") {
		return fallback
	}

	if templateExtension != "" && strings.HasSuffix(fileName, templateExtension) {
		fileName = strings.TrimSuffix(fileName, templateExtension)
	}
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return fallback
	}

	return fileName
}

// normalizeCollectionPath ensures the path has leading/trailing slashes.
// Example: "Episodes/EP01/Sequences" -> "/Episodes/EP01/Sequences/"
func normalizeCollectionPath(path string) string {
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

// findMatchingTemplate finds a template that matches the given collection type.
// Templates match based on the variable names they contain.
// Uses frontend format: <Episode>, <Sequence>, <Shot>, <Asset>, <CollectionType>
func findMatchingTemplate(collectionType string, dirStructure integrations.DirectoryStructure) string {
	if dirStructure.Paths == nil {
		return ""
	}

	collectionTypeLower := strings.ToLower(collectionType)

	// Standard collection types and their variable names (matching frontend format)
	standardTypes := map[string]string{
		"episode":  "<Episode>",
		"sequence": "<Sequence>",
		"shot":     "<Shot>",
	}

	// Check if this is a standard type
	if varName, ok := standardTypes[collectionTypeLower]; ok {
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
// The template is truncated at the collection's level to avoid unresolved variables.
func resolveTemplateVariables(template string, collection integrations.ExternalCollection, collectionByID map[string]integrations.ExternalCollection, style string) string {
	collectionTypeLower := strings.ToLower(collection.Type)

	// Standard variable mappings by collection type (matching frontend format)
	typeToVar := map[string]string{
		"episode":  "<Episode>",
		"sequence": "<Sequence>",
		"shot":     "<Shot>",
	}

	// Truncate template at the collection's variable level
	// For a sequence, "Episodes/<Episode>/<Sequence>/<Shot>" becomes "Episodes/<Episode>/<Sequence>"
	if varName, ok := typeToVar[collectionTypeLower]; ok {
		// Find the position of this collection's variable and truncate after it
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

	// Build collection hierarchy for variable resolution
	hierarchy := buildCollectionHierarchy(collection, collectionByID)

	// Resolve standard type variables from hierarchy
	for _, e := range hierarchy {
		typeLower := strings.ToLower(e.Type)
		if varName, ok := typeToVar[typeLower]; ok {
			result = strings.ReplaceAll(result, varName, applyNamingStyle(e.Name, style))
		}
	}

	// For assets, resolve <Asset> and <CollectionType>
	if collectionTypeLower != "episode" && collectionTypeLower != "sequence" && collectionTypeLower != "shot" {
		// This is an asset - resolve <Asset> with collection name
		result = strings.ReplaceAll(result, "<Asset>", applyNamingStyle(collection.Name, style))
		// Resolve <CollectionType> with collection type (like "Character", "Prop")
		result = strings.ReplaceAll(result, "<CollectionType>", applyNamingStyle(collection.Type, style))
	}

	// Remove any path segments that still contain unresolved variables
	// This handles cases where parent collections are missing (e.g., sequence without episode)
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

// resolveAssetTemplateVariables resolves the full asset template, including filename placeholders.
func resolveAssetTemplateVariables(template string, parentCollection integrations.ExternalCollection, asset integrations.ExternalAsset, collectionByID map[string]integrations.ExternalCollection, style, templateExtension string) string {
	result := template

	typeToVar := map[string]string{
		"episode":  "<Episode>",
		"sequence": "<Sequence>",
		"shot":     "<Shot>",
	}

	hierarchy := buildCollectionHierarchy(parentCollection, collectionByID)
	for _, e := range hierarchy {
		typeLower := strings.ToLower(e.Type)
		if varName, ok := typeToVar[typeLower]; ok {
			result = strings.ReplaceAll(result, varName, applyNamingStyle(e.Name, style))
		}
	}

	result = strings.ReplaceAll(result, "<Asset>", applyNamingStyle(parentCollection.Name, style))
	result = strings.ReplaceAll(result, "<CollectionType>", applyNamingStyle(parentCollection.Type, style))
	assetType := asset.AssetType
	if assetType == "" {
		assetType = asset.Name
	}
	result = strings.ReplaceAll(result, "<AssetType>", applyNamingStyle(assetType, style))
	result = strings.ReplaceAll(result, "<TemplateExtension>", templateExtension)

	return result
}

// buildCollectionHierarchy returns the collection and all its ancestors from leaf to root.
func buildCollectionHierarchy(collection integrations.ExternalCollection, collectionByID map[string]integrations.ExternalCollection) []integrations.ExternalCollection {
	var hierarchy []integrations.ExternalCollection
	current := collection

	for {
		hierarchy = append(hierarchy, current)
		if current.ParentID == "" {
			break
		}
		parent, exists := collectionByID[current.ParentID]
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

// buildExternalCollectionPath builds the full path for an external collection.
func buildExternalCollectionPath(collection integrations.ExternalCollection, collectionByID map[string]integrations.ExternalCollection) string {
	var parts []string
	current := collection

	for {
		parts = append([]string{current.Name}, parts...)
		if current.ParentID == "" {
			break
		}
		parent, exists := collectionByID[current.ParentID]
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
			ItemType:       "collection",
			CollectionPath: c.CollectionPath,
			ParentPath:     parentPath,
			ExternalID:     c.ExternalID,
			ExternalType:   c.ExternalType,
			ExternalName:   c.ExternalName,
			TypeName:       c.CollectionTypeName,
			TypeIcon:       c.CollectionTypeIcon,
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
			ItemType:          "asset",
			CollectionPath:    assetPath,
			ParentPath:        a.CollectionPath,
			ExternalID:        a.ExternalID,
			ExternalType:      a.ExternalType,
			ExternalTypeID:    a.ExternalTypeID,
			ExternalName:      a.ExternalName,
			TypeName:          a.AssetTypeName,
			TypeIcon:          a.AssetTypeIcon,
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

// ExecuteSync creates Clustta collections and assets from the provided sync preview data.
// Accepts collections and assets from frontend instead of re-fetching from integration.
// Creates all items - collections sorted by path depth (parents first), then assets with templates.
func (s *IntegrationService) ExecuteSync(projectPath string, collectionsJSON string, assetsJSON string) error {
	app := application.Get()

	// Get the currently signed-in user as the author for created assets
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return errors.New("must be signed in to sync")
	}

	// Parse collections and assets from JSON
	var collections []integrations.SyncCollection
	var assets []integrations.SyncAsset

	if collectionsJSON != "" {
		if err := json.Unmarshal([]byte(collectionsJSON), &collections); err != nil {
			return errors.New("invalid collections data")
		}
	}
	if assetsJSON != "" {
		if err := json.Unmarshal([]byte(assetsJSON), &assets); err != nil {
			return errors.New("invalid assets data")
		}
	}

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

	// Emit initial progress
	app.Event.Emit("progress-update", output.ProgressReport{
		Title:   "Integration Sync",
		Message: "Preparing sync...",
	})

	// Get linked integration
	integrationProject, err := repository.GetIntegrationProject(tx)
	if err != nil {
		return errors.New("no integration linked to this project")
	}

	// Count items to create for progress tracking
	collectionsToCreate := 0
	for _, c := range collections {
		if c.Action == "create" {
			collectionsToCreate++
		}
	}
	assetsToCreate := 0
	for _, a := range assets {
		if a.Action == "create" {
			assetsToCreate++
		}
	}
	totalItems := collectionsToCreate + assetsToCreate

	// Load current sync options
	syncOptions := integrations.SyncOptions{}
	if integrationProject.SyncOptions != "" {
		json.Unmarshal([]byte(integrationProject.SyncOptions), &syncOptions)
	}

	// Build preview struct for ensureTypesExist
	preview := integrations.SyncPreview{
		Collections: collections,
		Assets:      assets,
	}

	// Phase 0: Validate/auto-create types
	app.Event.Emit("progress-update", output.ProgressReport{
		Title:   "Integration Sync",
		Message: "Validating types...",
	})

	syncOptionsModified, err := s.ensureTypesExist(tx, preview, &syncOptions)
	if err != nil {
		return err
	}

	// Phase 1: Create collections (sorted by path depth - parents first)
	app.Event.Emit("progress-update", output.ProgressReport{
		Title:   "Integration Sync",
		Message: "Creating collections...",
		Total:   totalItems,
	})

	collectionMap, err := s.createCollectionsWithProgress(tx, collections, &syncOptions, app, totalItems)
	if err != nil {
		return err
	}

	// Phase 2: Create assets with templates
	app.Event.Emit("progress-update", output.ProgressReport{
		Title:   "Integration Sync",
		Message: "Creating assets...",
		Current: collectionsToCreate,
		Total:   totalItems,
	})

	assetMap, err := s.createAssetsWithProgress(tx, assets, &syncOptions, app, collectionsToCreate, totalItems, user.Id)
	if err != nil {
		return err
	}

	syncedAt := time.Now().UTC().Format(time.RFC3339)

	// Phase 3: Create mappings for collections
	app.Event.Emit("progress-update", output.ProgressReport{
		Title:      "Integration Sync",
		Message:    "Creating mappings...",
		Percentage: 90.0,
	})

	for _, coll := range collections {
		if coll.Action != "create" {
			continue
		}
		collectionID, exists := collectionMap[coll.ExternalID]
		if !exists {
			continue
		}

		mappingID := uuid.New().String()
		_, err = repository.CreateCollectionMapping(
			tx,
			mappingID,
			integrationProject.IntegrationId,
			coll.ExternalID,
			coll.ExternalType,
			coll.ExternalName,
			coll.ExternalParentID,
			coll.CollectionPath,
			"{}",
			collectionID,
			syncedAt,
		)
		if err != nil {
			return err
		}
	}

	// Phase 3b: Create mappings for assets
	for _, asset := range assets {
		if asset.Action != "create" {
			continue
		}
		assetID, exists := assetMap[asset.ExternalID]
		if !exists {
			continue
		}

		assigneesStr := "[]"
		if len(asset.ExternalAssignees) > 0 {
			assigneesStr = "[\"" + strings.Join(asset.ExternalAssignees, "\",\"") + "\"]"
		}

		mappingID := uuid.New().String()
		_, err = repository.CreateAssetMapping(
			tx,
			mappingID,
			integrationProject.IntegrationId,
			asset.ExternalID,
			asset.ExternalName,
			"", // parent_id not needed
			asset.ExternalType,
			asset.ExternalStatus,
			assigneesStr,
			"{}",
			assetID,
			syncedAt,
		)
		if err != nil {
			return err
		}
	}

	// Phase 4: Save updated sync_options if types were auto-created
	if syncOptionsModified {
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
	}

	// Final progress
	app.Event.Emit("progress-update", output.ProgressReport{
		Title:      "Integration Sync",
		Message:    "Sync complete!",
		Percentage: 100.0,
		Current:    totalItems,
		Total:      totalItems,
	})

	return tx.Commit()
}

// Available icons for collection types
var collectionTypeIcons = constants.CollectionTypeIcons

// Available icons for asset types
var assetTypeIcons = constants.AssetTypeIcons

// ensureTypesExist validates that all required types exist, auto-creating if needed.
// Returns true if syncOptions was modified.
func (s *IntegrationService) ensureTypesExist(tx *sqlx.Tx, preview integrations.SyncPreview, syncOptions *integrations.SyncOptions) (bool, error) {
	modified := false

	// Ensure collection type mappings map exists
	if syncOptions.CollectionTypeMappings == nil {
		syncOptions.CollectionTypeMappings = make(map[string]integrations.TypeMapping)
	}

	// Ensure asset type mappings map exists
	if syncOptions.AssetTypeMappings == nil {
		syncOptions.AssetTypeMappings = make(map[string]integrations.TypeMapping)
	}

	// Build set of used icons from existing collection types
	existingCollectionTypes, err := repository.GetCollectionTypes(tx)
	if err != nil {
		return false, err
	}
	usedCollectionIcons := make(map[string]bool)
	for _, et := range existingCollectionTypes {
		if et.Icon != "" {
			usedCollectionIcons[et.Icon] = true
		}
	}

	// Build set of used icons from existing asset types
	existingAssetTypes, err := repository.GetAssetTypes(tx)
	if err != nil {
		return false, err
	}
	usedAssetIcons := make(map[string]bool)
	for _, tt := range existingAssetTypes {
		if tt.Icon != "" {
			usedAssetIcons[tt.Icon] = true
		}
	}

	// Helper to get next available icon
	getNextAvailableIcon := func(iconPool []string, usedIcons map[string]bool) string {
		for _, icon := range iconPool {
			if !usedIcons[icon] {
				return icon
			}
		}
		// All icons used, generate a unique fallback
		return "type-" + uuid.New().String()[:8]
	}

	// Check collection types
	for _, coll := range preview.Collections {
		if coll.Action != "create" {
			continue
		}

		mapping, exists := syncOptions.CollectionTypeMappings[coll.ExternalType]
		if !exists || mapping.ClustttaName == "" {
			continue
		}

		// Has name but no ID - need to get or create the type
		if mapping.ClustttaTypeID == "" {
			// Try to find by name first (may already exist)
			existingType, err := repository.GetCollectionTypeByName(tx, mapping.ClustttaName)
			if err == nil {
				// Type exists, use its ID
				mapping.ClustttaTypeID = existingType.Id
				syncOptions.CollectionTypeMappings[coll.ExternalType] = mapping
				modified = true
				continue
			}

			// Need to create new type - get unique icon
			icon := getNextAvailableIcon(collectionTypeIcons, usedCollectionIcons)
			usedCollectionIcons[icon] = true

			collectionType, err := repository.CreateCollectionType(tx, "", mapping.ClustttaName, icon)
			if err != nil {
				return false, err
			}
			mapping.ClustttaTypeID = collectionType.Id
			syncOptions.CollectionTypeMappings[coll.ExternalType] = mapping
			modified = true
		}
	}

	// Check asset types
	for _, asset := range preview.Assets {
		if asset.Action != "create" {
			continue
		}

		mapping, exists := syncOptions.AssetTypeMappings[asset.ExternalType]
		if !exists || mapping.ClustttaName == "" {
			continue
		}

		// Has name but no ID - need to get or create the type
		if mapping.ClustttaTypeID == "" {
			// Try to find by name first (may already exist)
			existingType, err := repository.GetAssetTypeByName(tx, mapping.ClustttaName)
			if err == nil {
				// Type exists, use its ID
				mapping.ClustttaTypeID = existingType.Id
				syncOptions.AssetTypeMappings[asset.ExternalType] = mapping
				modified = true
				continue
			}

			// Need to create new type - get unique icon
			icon := getNextAvailableIcon(assetTypeIcons, usedAssetIcons)
			usedAssetIcons[icon] = true

			assetType, err := repository.CreateAssetType(tx, "", mapping.ClustttaName, icon)
			if err != nil {
				return false, err
			}
			mapping.ClustttaTypeID = assetType.Id
			syncOptions.AssetTypeMappings[asset.ExternalType] = mapping
			modified = true
		}
	}

	return modified, nil
}

// createCollectionsWithProgress creates collections and emits progress events grouped by type.
// Returns map of external_id -> created collection_id.
func (s *IntegrationService) createCollectionsWithProgress(tx *sqlx.Tx, collections []integrations.SyncCollection, syncOptions *integrations.SyncOptions, app *application.App, totalItems int) (map[string]string, error) {
	// Get generic type for intermediate folders
	genericType, err := repository.GetCollectionTypeByName(tx, "generic")
	if err != nil {
		return nil, errors.New("generic collection type not found")
	}

	// Phase 1: Create all missing intermediate path segments
	allPaths := make(map[string]bool)
	for _, coll := range collections {
		if coll.Action != "create" {
			continue
		}
		segments := strings.Split(strings.Trim(coll.CollectionPath, "/"), "/")
		currentPath := "/"
		for i := 0; i < len(segments)-1; i++ {
			currentPath = currentPath + segments[i] + "/"
			allPaths[currentPath] = true
		}
	}

	intermediatePaths := make([]string, 0, len(allPaths))
	for path := range allPaths {
		intermediatePaths = append(intermediatePaths, path)
	}
	sort.Slice(intermediatePaths, func(i, j int) bool {
		depthI := strings.Count(intermediatePaths[i], "/")
		depthJ := strings.Count(intermediatePaths[j], "/")
		return depthI < depthJ
	})

	for _, path := range intermediatePaths {
		existingCollection, err := repository.GetCollectionByPath(tx, path)
		if err == nil && existingCollection.Id != "" {
			continue
		}

		parentID := ""
		parentPath := getParentPath(path)
		if parentPath != "" && parentPath != "/" {
			parentCollection, err := repository.GetCollectionByPath(tx, parentPath)
			if err == nil && parentCollection.Id != "" {
				parentID = parentCollection.Id
			}
		}

		folderName := getPathSegmentName(path)
		newID := uuid.New().String()
		_, err = repository.CreateCollection(tx, newID, folderName, "", genericType.Id, parentID, "", false)
		if err != nil && err != error_service.ErrCollectionExists && err != error_service.ErrCollectionExistsInTrash {
			return nil, err
		}
	}

	// Phase 2: Group collections by type and sort by depth within each group
	typeGroups := make(map[string][]integrations.SyncCollection)
	for _, coll := range collections {
		if coll.Action != "create" {
			continue
		}
		typeGroups[coll.ExternalType] = append(typeGroups[coll.ExternalType], coll)
	}

	// Sort each group by path depth
	for typeName := range typeGroups {
		group := typeGroups[typeName]
		sort.Slice(group, func(i, j int) bool {
			depthI := strings.Count(group[i].CollectionPath, "/")
			depthJ := strings.Count(group[j].CollectionPath, "/")
			return depthI < depthJ
		})
		typeGroups[typeName] = group
	}

	result := make(map[string]string)
	currentItem := 0

	// Process each type group
	for typeName, group := range typeGroups {
		// Emit progress for this type group
		app.Event.Emit("progress-update", output.ProgressReport{
			Title:      "Integration Sync",
			Message:    "Creating " + typeName + "s...",
			Percentage: (float64(currentItem) / float64(totalItems)) * 50,
			Current:    currentItem,
			Total:      totalItems,
		})

		for _, coll := range group {
			currentItem++

			existingCollection, err := repository.GetCollectionByPath(tx, coll.CollectionPath)
			if err == nil && existingCollection.Id != "" {
				result[coll.ExternalID] = existingCollection.Id
				continue
			}

			parentID := ""
			parentPath := getParentPath(coll.CollectionPath)
			if parentPath != "" && parentPath != "/" {
				parentCollection, err := repository.GetCollectionByPath(tx, parentPath)
				if err == nil && parentCollection.Id != "" {
					parentID = parentCollection.Id
				}
			}

			collectionTypeID := ""
			if mapping, exists := syncOptions.CollectionTypeMappings[coll.ExternalType]; exists {
				collectionTypeID = mapping.ClustttaTypeID
			}
			if collectionTypeID == "" {
				collectionTypeID = genericType.Id
			}

			collectionName := getPathSegmentName(coll.CollectionPath)
			if collectionName == "" {
				collectionName = coll.ExternalName
			}

			newID := uuid.New().String()
			collection, err := repository.CreateCollection(tx, newID, collectionName, "", collectionTypeID, parentID, "", false)
			if err != nil {
				if err == error_service.ErrCollectionExists || err == error_service.ErrCollectionExistsInTrash {
					existingCollection, getErr := repository.GetCollectionByName(tx, collectionName, parentID)
					if getErr == nil {
						result[coll.ExternalID] = existingCollection.Id
						continue
					}
				}
				return nil, err
			}

			result[coll.ExternalID] = collection.Id
		}
	}

	return result, nil
}

// createAssetsWithProgress creates assets with templates and emits progress events grouped by type.
// Returns map of external_id -> created asset_id.
func (s *IntegrationService) createAssetsWithProgress(tx *sqlx.Tx, assets []integrations.SyncAsset, syncOptions *integrations.SyncOptions, app *application.App, startIndex int, totalItems int, userId string) (map[string]string, error) {
	// Group assets by type
	typeGroups := make(map[string][]integrations.SyncAsset)
	for _, asset := range assets {
		if asset.Action != "create" {
			continue
		}
		typeGroups[asset.ExternalType] = append(typeGroups[asset.ExternalType], asset)
	}

	result := make(map[string]string)
	currentItem := startIndex

	// Process each type group
	for typeName, group := range typeGroups {
		// Emit progress for this type group
		app.Event.Emit("progress-update", output.ProgressReport{
			Title:      "Integration Sync",
			Message:    "Creating " + typeName + " assets...",
			Percentage: 50.0 + (float64(currentItem-startIndex)/float64(totalItems-startIndex))*40,
			Current:    currentItem,
			Total:      totalItems,
		})

		for _, asset := range group {
			currentItem++

			if asset.TemplateID == "" {
				continue
			}

			parentID := ""
			if asset.CollectionPath != "" {
				parentCollection, err := repository.GetCollectionByPath(tx, asset.CollectionPath)
				if err == nil && parentCollection.Id != "" {
					parentID = parentCollection.Id
				}
			}

			assetTypeID := ""
			if mapping, exists := syncOptions.AssetTypeMappings[asset.ExternalType]; exists {
				assetTypeID = mapping.ClustttaTypeID
			}
			if assetTypeID == "" {
				genericType, err := repository.GetAssetTypeByName(tx, "generic")
				if err != nil {
					return nil, errors.New("generic asset type not found")
				}
				assetTypeID = genericType.Id
			}

			newID := uuid.New().String()
			checkpointGroupID := uuid.New().String()

			createdAsset, err := repository.CreateAsset(
				tx,
				newID,
				asset.ExternalName,
				assetTypeID,
				parentID,
				false,
				asset.TemplateID,
				"",
				"",
				[]string{},
				"",
				false,
				"",
				userId,
				"Synced from external integration",
				checkpointGroupID,
				nil,
			)
			if err != nil {
				if err == error_service.ErrAssetExists || err == error_service.ErrAssetExistsInTrash {
					template, _ := repository.GetTemplate(tx, asset.TemplateID)
					existingAsset, getErr := repository.GetAssetByName(tx, asset.ExternalName, parentID, template.Extension)
					if getErr == nil {
						result[asset.ExternalID] = existingAsset.Id
						continue
					}
				}
				return nil, err
			}

			result[asset.ExternalID] = createdAsset.Id
		}
	}

	return result, nil
}

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
				CollectionTypeMappings: make(map[string]integrations.TypeMapping),
				AssetTypeMappings:      make(map[string]integrations.TypeMapping),
			}, nil
		}
	}

	// Initialize maps if nil
	if syncOptions.CollectionTypeMappings == nil {
		syncOptions.CollectionTypeMappings = make(map[string]integrations.TypeMapping)
	}
	if syncOptions.AssetTypeMappings == nil {
		syncOptions.AssetTypeMappings = make(map[string]integrations.TypeMapping)
	}
	if syncOptions.AssetTypeTemplates == nil {
		syncOptions.AssetTypeTemplates = make(map[string]string)
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

// GetExternalTypes fetches collection and asset types from the external integration.
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

	collectionTypes, err := integration.GetCollectionTypes(token, integrationProject.ApiUrl, integrationProject.ExternalProjectId)
	if err != nil {
		return nil, nil, err
	}

	assetTypes, err := integration.GetAssetTypes(token, integrationProject.ApiUrl, integrationProject.ExternalProjectId)
	if err != nil {
		return nil, nil, err
	}

	return collectionTypes, assetTypes, nil
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
	if syncOptions.CollectionTypeMappings == nil {
		syncOptions.CollectionTypeMappings = make(map[string]integrations.TypeMapping)
	}
	if syncOptions.AssetTypeMappings == nil {
		syncOptions.AssetTypeMappings = make(map[string]integrations.TypeMapping)
	}

	// Get local types
	localCollectionTypes, err := repository.GetCollectionTypes(tx)
	if err != nil {
		return nil, err
	}
	localAssetTypes, err := repository.GetAssetTypes(tx)
	if err != nil {
		return nil, err
	}

	// Build lookup maps for local types (by lowercase name)
	localCollectionTypeMap := make(map[string]models.CollectionType)
	for _, et := range localCollectionTypes {
		localCollectionTypeMap[strings.ToLower(et.Name)] = et
	}
	localAssetTypeMap := make(map[string]models.AssetType)
	for _, tt := range localAssetTypes {
		localAssetTypeMap[strings.ToLower(tt.Name)] = tt
	}

	// Get external types
	externalCollectionTypes, err := integration.GetCollectionTypes(token, integrationProject.ApiUrl, integrationProject.ExternalProjectId)
	if err != nil {
		return nil, err
	}
	externalAssetTypes, err := integration.GetAssetTypes(token, integrationProject.ApiUrl, integrationProject.ExternalProjectId)
	if err != nil {
		return nil, err
	}

	missingTypes := []integrations.MissingType{}

	// Available icons for selection
	collectionIcons := constants.CollectionTypeIcons
	assetIcons := constants.AssetTypeIcons

	// Check collection types
	for i, et := range externalCollectionTypes {
		// Skip if already mapped
		if _, exists := syncOptions.CollectionTypeMappings[et.Name]; exists {
			continue
		}

		// Check if matching local type exists (case-insensitive)
		suggestedName := strings.ToLower(strings.TrimSpace(et.Name))
		if localType, exists := localCollectionTypeMap[suggestedName]; exists {
			// Auto-map to existing type
			syncOptions.CollectionTypeMappings[et.Name] = integrations.TypeMapping{
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
			TypeCategory:  "collection",
			SuggestedName: suggestedName,
			SuggestedIcon: collectionIcons[i%len(collectionIcons)],
		})
	}

	// Check asset types
	for i, tt := range externalAssetTypes {
		// Skip if already mapped
		if _, exists := syncOptions.AssetTypeMappings[tt.Name]; exists {
			continue
		}

		// Check if matching local type exists (case-insensitive)
		suggestedName := strings.ToLower(strings.TrimSpace(tt.Name))
		if localType, exists := localAssetTypeMap[suggestedName]; exists {
			// Auto-map to existing type
			syncOptions.AssetTypeMappings[tt.Name] = integrations.TypeMapping{
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
			TypeCategory:  "asset",
			SuggestedName: suggestedName,
			SuggestedIcon: assetIcons[i%len(assetIcons)],
		})
	}

	// Save auto-mapped types if any
	if len(syncOptions.CollectionTypeMappings) > 0 || len(syncOptions.AssetTypeMappings) > 0 {
		syncOptionsJSON, _ := json.Marshal(syncOptions)
		repository.UpdateIntegrationProject(tx, integrationProject.Id, map[string]interface{}{
			"sync_options": string(syncOptionsJSON),
		})
		tx.Commit()
	}

	return missingTypes, nil
}

// CreateMissingTypes creates collection and asset types in Clustta for missing external types.
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
	if syncOptions.CollectionTypeMappings == nil {
		syncOptions.CollectionTypeMappings = make(map[string]integrations.TypeMapping)
	}
	if syncOptions.AssetTypeMappings == nil {
		syncOptions.AssetTypeMappings = make(map[string]integrations.TypeMapping)
	}

	// Create each missing type
	for _, mt := range missingTypes {
		if mt.TypeCategory == "collection" {
			// Create collection type
			collectionType, err := repository.GetOrCreateCollectionType(tx, mt.SuggestedName, mt.SuggestedIcon)
			if err != nil {
				return err
			}

			// Add to mappings
			syncOptions.CollectionTypeMappings[mt.ExternalName] = integrations.TypeMapping{
				ExternalName:   mt.ExternalName,
				ExternalID:     mt.ExternalID,
				ClustttaTypeID: collectionType.Id,
				ClustttaName:   collectionType.Name,
				ClustttaIcon:   collectionType.Icon,
			}
		} else if mt.TypeCategory == "asset" {
			// Create asset type
			assetType, err := repository.GetOrCreateAssetType(tx, mt.SuggestedName, mt.SuggestedIcon)
			if err != nil {
				return err
			}

			// Add to mappings
			syncOptions.AssetTypeMappings[mt.ExternalName] = integrations.TypeMapping{
				ExternalName:   mt.ExternalName,
				ExternalID:     mt.ExternalID,
				ClustttaTypeID: assetType.Id,
				ClustttaName:   assetType.Name,
				ClustttaIcon:   assetType.Icon,
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

// GetLocalTypes returns all collection and asset types from the project.
// Used by frontend to populate mapping dropdowns.
func (s *IntegrationService) GetLocalTypes(projectPath string) ([]models.CollectionType, []models.AssetType, error) {
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

	collectionTypes, err := repository.GetCollectionTypes(tx)
	if err != nil {
		return nil, nil, err
	}

	assetTypes, err := repository.GetAssetTypes(tx)
	if err != nil {
		return nil, nil, err
	}

	return collectionTypes, assetTypes, nil
}

// GetExternalStatuses fetches task statuses from the external integration.
// Returns the list of available statuses for mapping.
func (s *IntegrationService) GetExternalStatuses(projectPath, token string) ([]integrations.ExternalStatusInfo, error) {
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

	integration, err := integrations.Get(integrationProject.IntegrationId)
	if err != nil {
		return nil, err
	}

	return integration.GetTaskStatuses(token, integrationProject.ApiUrl)
}

// SaveStatusMappings saves status mappings to sync_options (Clustta status ID → external status ID).
func (s *IntegrationService) SaveStatusMappings(projectPath string, statusMappings map[string]string) error {
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

	var syncOptions integrations.SyncOptions
	if integrationProject.SyncOptions != "" && integrationProject.SyncOptions != "{}" {
		json.Unmarshal([]byte(integrationProject.SyncOptions), &syncOptions)
	}

	syncOptions.StatusMappings = statusMappings

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

// PushToIntegration pushes a checkpoint's preview and status to the linked external integration.
// Returns an error if the push fails. A nil error means the push completed (or was skipped).
func (s *IntegrationService) PushToIntegration(projectPath string, assetIds []string, checkpointId, previewPath, message string) error {
	app := application.Get()

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return fmt.Errorf("failed to open db: %w", err)
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	// Check if project has a linked integration
	integrationProject, err := repository.GetIntegrationProject(tx)
	if err != nil || !integrationProject.Enabled {
		return nil // No integration linked — not an error
	}

	// Load integration client
	integration, err := integrations.Get(integrationProject.IntegrationId)
	if err != nil {
		return fmt.Errorf("unknown integration %s: %w", integrationProject.IntegrationId, err)
	}

	// Load stored credentials
	cred, err := settings.GetIntegrationCredential(integrationProject.IntegrationId)
	if err != nil || cred.AccessToken == "" {
		return nil // Not authenticated — not an error
	}

	// Validate token is still valid
	valid, err := integration.ValidateToken(cred.AccessToken, integrationProject.ApiUrl)
	if err != nil || !valid {
		return fmt.Errorf("token invalid for %s", integrationProject.IntegrationId)
	}

	// Load sync options for status mappings
	var syncOptions integrations.SyncOptions
	if integrationProject.SyncOptions != "" && integrationProject.SyncOptions != "{}" {
		json.Unmarshal([]byte(integrationProject.SyncOptions), &syncOptions)
	}

	pushedCount := 0
	failedCount := 0
	totalAssets := len(assetIds)
	showProgress := previewPath != ""
	integrationName := integration.Name()

	for i, assetId := range assetIds {

		// Look up integration mapping for this asset
		mapping, err := repository.GetAssetMappingByAssetId(tx, assetId)
		if err != nil {
			continue // Asset not mapped to integration — skip
		}

		// Resolve asset name and external status ID
		assetName := ""
		externalStatusId := ""
		asset, err := repository.GetAsset(tx, assetId)
		if err == nil {
			assetName = asset.Name
			if syncOptions.StatusMappings != nil && asset.StatusId != "" {
				if sid, ok := syncOptions.StatusMappings[asset.StatusId]; ok {
					externalStatusId = sid
				}
			}
		}

		// Emit initial progress for this asset
		if showProgress {
			app.Event.Emit("progress-update", output.ProgressReport{
				Title:      "Uploading to " + integrationName,
				Message:    "Updating preview for " + assetName,
				Percentage: float64(i) / float64(totalAssets) * 100,
				Current:    i + 1,
				Total:      totalAssets + 1,
			})
		}

		// Upload preview if provided — includes status change in the same comment when mapped
		statusPushedWithPreview := false
		if previewPath != "" {
			var onProgress integrations.UploadProgressFunc
			if showProgress {
				onProgress = func(bytesSent, totalBytes int64) {
					pct := (float64(i) + float64(bytesSent)/float64(totalBytes)) / float64(totalAssets) * 100
					app.Event.Emit("progress-update", output.ProgressReport{
						Title:        "Uploading to " + integrationName,
						Message:      "Updating preview for " + assetName,
						Percentage:   pct,
						Current:      i + 1,
						Total:        totalAssets + 1,
						ExtraMessage: formatBytes(bytesSent) + " / " + formatBytes(totalBytes),
					})
				}
			}
			err = integration.UploadPreview(cred.AccessToken, integrationProject.ApiUrl, mapping.ExternalId, previewPath, message, externalStatusId, onProgress)
			if err != nil {
				log.Printf("integration push: preview upload failed for asset %s: %v", assetId, err)
				failedCount++
				continue
			}
			if externalStatusId != "" {
				statusPushedWithPreview = true
			}
		}

		// Push status update only if not already included in the preview comment
		if externalStatusId != "" && !statusPushedWithPreview {
			err = integration.UpdateAssetStatus(cred.AccessToken, integrationProject.ApiUrl, mapping.ExternalId, externalStatusId)
			if err != nil {
				log.Printf("integration push: status update failed for asset %s: %v", assetId, err)
			}
		}

		// Update last_pushed_checkpoint_id on the mapping
		if checkpointId != "" {
			repository.UpdateAssetMapping(tx, mapping.Id, map[string]interface{}{
				"last_pushed_checkpoint_id": checkpointId,
				"synced_at":                 time.Now().UTC().Format(time.RFC3339),
			})
		}

		pushedCount++
	}

	// Dismiss progress bar
	if showProgress {
		app.Event.Emit("progress-update", output.ProgressReport{
			Title: "Uploading to " + integrationName, Message: "complete", Percentage: 100, Current: totalAssets + 1, Total: totalAssets + 1,
		})
	}

	tx.Commit()

	// Emit completion event for UI feedback
	if pushedCount > 0 {
		app.Event.Emit("integration-push-complete", map[string]interface{}{
			"pushed":      pushedCount,
			"failed":      failedCount,
			"integration": integrationProject.IntegrationId,
		})
	}

	if failedCount > 0 {
		app.Event.Emit("integration-push-failed", map[string]interface{}{
			"failed":      failedCount,
			"integration": integrationProject.IntegrationId,
			"error":       fmt.Sprintf("%d asset(s) failed to push", failedCount),
		})
		return fmt.Errorf("%d asset(s) failed to push to %s", failedCount, integrationProject.IntegrationId)
	}

	return nil
}

// formatBytes returns a human-readable byte size string.
func formatBytes(b int64) string {
	const mb = 1024 * 1024
	const kb = 1024
	switch {
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.0f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
