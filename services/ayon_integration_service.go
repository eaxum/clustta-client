package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"clustta/internal/auth_service"
	"clustta/internal/error_service"
	"clustta/internal/integrations"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	ayonIntegrationID        = "ayon"
	ayonDCCIntegrationID     = "ayon-dcc"
	ayonWorkfileKeySeparator = "::"
)

var (
	ErrAYONIntegrationNotConfigured = errors.New("AYON integration is not configured")
	ErrAYONMappingNotConfigured     = errors.New("AYON mapping is not configured")
	ErrAYONPermissionDenied         = errors.New("AYON workfile operation is not permitted")
	ayonWorkfileMu                  sync.Mutex
)

// AYONWorkfileContext identifies one AYON task workfile.
type AYONWorkfileContext struct {
	ProjectID   string
	FolderID    string
	FolderPath  string
	Folders     []AYONFolderContext
	TaskID      string
	TaskName    string
	TaskTypeID  string
	TaskType    string
	Application string
	Extension   string
}

// AYONFolderContext identifies one folder in the selected AYON hierarchy.
type AYONFolderContext struct {
	ID       string
	ParentID string
	Name     string
	Path     string
	Type     string
}

// AYONWorkfilePreparation describes how the DCC should open a workfile.
type AYONWorkfilePreparation struct {
	Asset          *models.Asset
	TargetPath     string
	RequiresImport bool
	FileExists     bool
}

// AYONWorkfileImportResult contains the imported asset and initial checkpoint.
type AYONWorkfileImportResult struct {
	Asset      models.Asset
	Checkpoint models.Checkpoint
	Created    bool
}

// PrepareAYONWorkfile creates missing hierarchy and resolves the launch target.
func (s *IntegrationService) PrepareAYONWorkfile(projectPath string, context AYONWorkfileContext) (AYONWorkfilePreparation, error) {
	if err := validateAYONPreparationContext(context); err != nil {
		return AYONWorkfilePreparation{}, err
	}

	ayonWorkfileMu.Lock()
	defer ayonWorkfileMu.Unlock()

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return AYONWorkfilePreparation{}, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return AYONWorkfilePreparation{}, err
	}
	defer tx.Rollback()

	userID, role, err := ayonProjectUserAndRole(tx)
	if err != nil {
		return AYONWorkfilePreparation{}, err
	}
	existing, found, err := resolveAnyMappedAYONAsset(tx, context, userID, role)
	if err != nil {
		return AYONWorkfilePreparation{}, err
	}
	if found {
		if err = tx.Commit(); err != nil {
			return AYONWorkfilePreparation{}, err
		}
		return AYONWorkfilePreparation{Asset: &existing}, nil
	}

	workingDirectory, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return AYONWorkfilePreparation{}, err
	}
	targetPath, relativePath, err := ayonTargetPaths(workingDirectory, context)
	if err != nil {
		return AYONWorkfilePreparation{}, err
	}
	collectionID, err := resolveTrackAssetCollection(
		tx,
		relativePath,
		models.User{Id: userID},
		role,
	)
	if err != nil {
		return AYONWorkfilePreparation{}, err
	}
	if err = ensureAYONFolderMappings(tx, context, collectionID); err != nil {
		return AYONWorkfilePreparation{}, err
	}
	if err = tx.Commit(); err != nil {
		return AYONWorkfilePreparation{}, err
	}
	if err = os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return AYONWorkfilePreparation{}, err
	}
	_, statErr := os.Stat(targetPath)
	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return AYONWorkfilePreparation{}, statErr
	}
	return AYONWorkfilePreparation{
		TargetPath:     targetPath,
		RequiresImport: true,
		FileExists:     statErr == nil,
	}, nil
}

// ImportAYONWorkfile tracks a DCC-created file and creates its AYON mapping.
func (s *IntegrationService) ImportAYONWorkfile(
	projectPath string,
	context AYONWorkfileContext,
	filePath string,
	progress func(int),
) (AYONWorkfileImportResult, error) {
	if err := validateAYONPreparationContext(context); err != nil {
		return AYONWorkfileImportResult{}, err
	}

	ayonWorkfileMu.Lock()
	defer ayonWorkfileMu.Unlock()

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return AYONWorkfileImportResult{}, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return AYONWorkfileImportResult{}, err
	}
	defer tx.Rollback()

	userID, role, err := ayonProjectUserAndRole(tx)
	if err != nil {
		return AYONWorkfileImportResult{}, err
	}
	existing, found, err := resolveAnyMappedAYONAsset(tx, context, userID, role)
	if err != nil {
		return AYONWorkfileImportResult{}, err
	}
	if found {
		checkpoint, checkpointErr := repository.GetLatestCheckpoint(tx, existing.Id)
		if checkpointErr != nil {
			return AYONWorkfileImportResult{}, checkpointErr
		}
		if err = tx.Commit(); err != nil {
			return AYONWorkfileImportResult{}, err
		}
		return AYONWorkfileImportResult{Asset: existing, Checkpoint: checkpoint}, nil
	}
	if err = authorizeTrackAssetRole(role, true); err != nil {
		return AYONWorkfileImportResult{}, err
	}

	workingDirectory, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return AYONWorkfileImportResult{}, err
	}
	targetPath, relativePath, err := ayonTargetPaths(workingDirectory, context)
	if err != nil {
		return AYONWorkfileImportResult{}, err
	}
	if canonicalTrackPath(filePath) != canonicalTrackPath(targetPath) {
		return AYONWorkfileImportResult{}, errors.New("AYON workfile does not match the prepared target path")
	}
	fileInfo, err := os.Stat(targetPath)
	if err != nil {
		return AYONWorkfileImportResult{}, fmt.Errorf("AYON workfile is not available: %w", err)
	}
	if !fileInfo.Mode().IsRegular() || fileInfo.Size() == 0 {
		return AYONWorkfileImportResult{}, errors.New("AYON workfile must be a non-empty regular file")
	}

	collection, err := repository.GetCollectionByPath(tx, utils.GetParent(relativePath))
	if err != nil {
		return AYONWorkfileImportResult{}, err
	}
	assetType, err := repository.GetAssetTypeByName(tx, "generic")
	if err != nil {
		return AYONWorkfileImportResult{}, err
	}
	status, err := repository.GetStatusByShortName(tx, "todo")
	if err != nil {
		return AYONWorkfileImportResult{}, err
	}
	assetID := uuid.NewString()
	checkpointGroupID := uuid.NewString()
	callback := func(current, total int, _, _ string) {
		if progress != nil && total > 0 {
			progress(trackChunkProgressStart + current*trackChunkProgressRange/total)
		}
	}
	if err = repository.CreateAssetFast(
		tx,
		assetID,
		context.TaskName,
		assetType.Id,
		collection.Id,
		false,
		"",
		targetPath,
		"",
		userID,
		"Created from AYON task",
		checkpointGroupID,
		relativePath,
		status.Id,
		callback,
	); err != nil {
		return AYONWorkfileImportResult{}, err
	}
	if err = repository.AssignAsset(tx, assetID, userID); err != nil {
		return AYONWorkfileImportResult{}, err
	}
	if err = createAYONAssetMapping(tx, context, assetID, userID); err != nil {
		return AYONWorkfileImportResult{}, err
	}
	asset, err := repository.GetAsset(tx, assetID)
	if err != nil {
		return AYONWorkfileImportResult{}, err
	}
	checkpoint, err := repository.GetLatestCheckpoint(tx, assetID)
	if err != nil {
		return AYONWorkfileImportResult{}, err
	}
	if err = tx.Commit(); err != nil {
		return AYONWorkfileImportResult{}, err
	}
	return AYONWorkfileImportResult{
		Asset:      asset,
		Checkpoint: checkpoint,
		Created:    true,
	}, nil
}

// AYONWorkfileResult contains the resolved Clustta asset.
type AYONWorkfileResult struct {
	Asset   models.Asset
	Created bool
}

// ResolveOrCreateAYONWorkfile resolves or lazily creates an AYON task workfile.
func (s *IntegrationService) ResolveOrCreateAYONWorkfile(projectPath string, context AYONWorkfileContext) (AYONWorkfileResult, error) {
	if err := validateAYONWorkfileContext(context); err != nil {
		return AYONWorkfileResult{}, err
	}

	ayonWorkfileMu.Lock()
	defer ayonWorkfileMu.Unlock()

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return AYONWorkfileResult{}, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return AYONWorkfileResult{}, err
	}
	defer tx.Rollback()

	userID, role, err := ayonProjectUserAndRole(tx)
	if err != nil {
		return AYONWorkfileResult{}, err
	}
	integrationProject, err := repository.GetIntegrationProject(tx)
	if err != nil || !integrationProject.Enabled || integrationProject.IntegrationId != ayonIntegrationID {
		return AYONWorkfileResult{}, ErrAYONIntegrationNotConfigured
	}
	if integrationProject.ExternalProjectId != context.ProjectID {
		return AYONWorkfileResult{}, fmt.Errorf("%w: AYON project is not linked", ErrAYONMappingNotConfigured)
	}

	externalID := strings.TrimSpace(context.TaskID)
	existing, found, err := resolveMappedAYONAsset(tx, externalID, userID, role)
	if err != nil {
		return AYONWorkfileResult{}, err
	}
	if !found {
		legacyExternalID := ayonWorkfileExternalID(context.TaskID, context.Application)
		existing, found, err = resolveMappedAYONAsset(tx, legacyExternalID, userID, role)
		if err != nil {
			return AYONWorkfileResult{}, err
		}
	}
	if found {
		if err := tx.Commit(); err != nil {
			return AYONWorkfileResult{}, err
		}
		return AYONWorkfileResult{Asset: existing}, nil
	}
	if !role.CreateAsset || !role.AssignAsset {
		return AYONWorkfileResult{}, fmt.Errorf("%w: create_asset and assign_asset are required", ErrAYONPermissionDenied)
	}

	collectionMapping, err := repository.GetCollectionMappingByExternalId(tx, ayonIntegrationID, context.FolderID)
	if err != nil {
		if errors.Is(err, repository.ErrCollectionMappingNotFound) {
			return AYONWorkfileResult{}, fmt.Errorf("%w: AYON folder has no Clustta collection", ErrAYONMappingNotConfigured)
		}
		return AYONWorkfileResult{}, err
	}

	syncOptions := integrations.SyncOptions{}
	if integrationProject.SyncOptions != "" {
		if err := json.Unmarshal([]byte(integrationProject.SyncOptions), &syncOptions); err != nil {
			return AYONWorkfileResult{}, fmt.Errorf("invalid AYON sync options: %w", err)
		}
	}
	templateID := ayonTemplateID(syncOptions, context.TaskTypeID, context.TaskType)
	if templateID == "" {
		return AYONWorkfileResult{}, fmt.Errorf("%w: task type %q has no workfile template", ErrAYONMappingNotConfigured, context.TaskType)
	}
	template, err := repository.GetTemplate(tx, templateID)
	if err != nil {
		return AYONWorkfileResult{}, fmt.Errorf("mapped workfile template is unavailable: %w", err)
	}
	if context.Extension != "" && !strings.EqualFold(template.Extension, normalizedExtension(context.Extension)) {
		return AYONWorkfileResult{}, fmt.Errorf("mapped template extension %q does not match %q", template.Extension, context.Extension)
	}

	assetTypeID, err := ayonAssetTypeID(tx, syncOptions, context.TaskTypeID, context.TaskType)
	if err != nil {
		return AYONWorkfileResult{}, err
	}
	asset, err := repository.CreateAsset(
		tx,
		"",
		context.TaskName,
		assetTypeID,
		collectionMapping.CollectionId,
		false,
		templateID,
		"",
		"",
		[]string{},
		"",
		false,
		"",
		userID,
		"Created from AYON task",
		uuid.NewString(),
		nil,
	)
	if err != nil {
		if !errors.Is(err, error_service.ErrAssetExists) {
			return AYONWorkfileResult{}, err
		}
		asset, err = repository.GetAssetByName(tx, context.TaskName, collectionMapping.CollectionId, template.Extension)
		if err != nil {
			return AYONWorkfileResult{}, err
		}
	}
	if asset.AssigneeId == "" {
		if err := repository.AssignAsset(tx, asset.Id, userID); err != nil {
			return AYONWorkfileResult{}, err
		}
	} else if asset.AssigneeId != userID {
		return AYONWorkfileResult{}, fmt.Errorf("%w: matching workfile is assigned to another user", ErrAYONPermissionDenied)
	}

	metadata, err := json.Marshal(map[string]string{
		"application": context.Application,
		"task_id":     context.TaskID,
	})
	if err != nil {
		return AYONWorkfileResult{}, err
	}
	assignees, err := json.Marshal([]string{userID})
	if err != nil {
		return AYONWorkfileResult{}, err
	}
	_, err = repository.CreateAssetMapping(
		tx,
		uuid.NewString(),
		ayonIntegrationID,
		externalID,
		context.TaskName,
		context.FolderID,
		context.TaskType,
		"",
		string(assignees),
		string(metadata),
		asset.Id,
		time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return AYONWorkfileResult{}, err
	}
	asset, err = repository.GetAsset(tx, asset.Id)
	if err != nil {
		return AYONWorkfileResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return AYONWorkfileResult{}, err
	}
	return AYONWorkfileResult{Asset: asset, Created: true}, nil
}

func validateAYONWorkfileContext(context AYONWorkfileContext) error {
	required := map[string]string{
		"project_id":  context.ProjectID,
		"folder_id":   context.FolderID,
		"task_id":     context.TaskID,
		"task_name":   context.TaskName,
		"task_type":   context.TaskType,
		"application": context.Application,
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	return nil
}

func ayonProjectUserAndRole(tx *sqlx.Tx) (string, models.Role, error) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return "", models.Role{}, err
	}
	projectUser, err := repository.GetUser(tx, user.Id)
	if err != nil {
		return "", models.Role{}, err
	}
	role, err := repository.GetRole(tx, projectUser.RoleId)
	if err != nil {
		return "", models.Role{}, err
	}
	return user.Id, role, nil
}

func resolveMappedAYONAsset(tx *sqlx.Tx, externalID, userID string, role models.Role) (models.Asset, bool, error) {
	return resolveMappedAYONAssetForIntegration(
		tx,
		ayonIntegrationID,
		externalID,
		userID,
		role,
	)
}

func resolveMappedAYONAssetForIntegration(
	tx *sqlx.Tx,
	integrationID string,
	externalID string,
	userID string,
	role models.Role,
) (models.Asset, bool, error) {
	mapping, err := repository.GetAssetMappingByExternalId(tx, integrationID, externalID)
	if err != nil {
		if errors.Is(err, repository.ErrAssetMappingNotFound) {
			return models.Asset{}, false, nil
		}
		return models.Asset{}, false, err
	}
	asset, err := repository.GetAsset(tx, mapping.AssetId)
	if err != nil {
		return models.Asset{}, false, err
	}
	if asset.AssigneeId != "" && asset.AssigneeId != userID {
		return models.Asset{}, false, fmt.Errorf("%w: workfile is assigned to another user", ErrAYONPermissionDenied)
	}
	if asset.AssigneeId == "" {
		if !role.AssignAsset {
			return models.Asset{}, false, fmt.Errorf("%w: assign_asset is required", ErrAYONPermissionDenied)
		}
		if err := repository.AssignAsset(tx, asset.Id, userID); err != nil {
			return models.Asset{}, false, err
		}
		asset, err = repository.GetAsset(tx, asset.Id)
		if err != nil {
			return models.Asset{}, false, err
		}
	}
	return asset, true, nil
}

func ayonWorkfileExternalID(taskID, application string) string {
	return strings.TrimSpace(taskID) + ayonWorkfileKeySeparator + strings.ToLower(strings.TrimSpace(application))
}

func ayonTemplateID(options integrations.SyncOptions, taskTypeID, taskType string) string {
	if templateID := options.AssetTypeTemplates[taskTypeID]; templateID != "" {
		return templateID
	}
	return options.AssetTypeTemplates[taskType]
}

func ayonAssetTypeID(tx *sqlx.Tx, options integrations.SyncOptions, taskTypeID, taskType string) (string, error) {
	for _, key := range []string{taskTypeID, taskType} {
		if mapping, ok := options.AssetTypeMappings[key]; ok && mapping.ClustttaTypeID != "" {
			return mapping.ClustttaTypeID, nil
		}
	}
	genericType, err := repository.GetAssetTypeByName(tx, "generic")
	if err != nil {
		return "", errors.New("generic asset type not found")
	}
	return genericType.Id, nil
}

func normalizedExtension(extension string) string {
	extension = strings.TrimSpace(extension)
	if extension == "" || strings.HasPrefix(extension, ".") {
		return extension
	}
	return "." + extension
}

func validateAYONPreparationContext(context AYONWorkfileContext) error {
	if err := validateAYONWorkfileContext(context); err != nil {
		return err
	}
	if strings.TrimSpace(context.FolderPath) == "" {
		return errors.New("folder_path is required")
	}
	if strings.TrimSpace(context.Extension) == "" {
		return errors.New("extension is required")
	}
	name := strings.TrimSpace(context.TaskName)
	if name == "." || name == ".." || strings.ContainsAny(name, `/\\`) {
		return errors.New("task_name cannot contain path separators")
	}
	return nil
}

func ayonTargetPaths(workingDirectory string, context AYONWorkfileContext) (string, string, error) {
	folderPath := normalizeAYONFolderPath(context.FolderPath)
	if folderPath == "/" {
		return "", "", errors.New("AYON workfiles cannot be created at the project root")
	}
	fileName := strings.TrimSpace(context.TaskName) + normalizedExtension(context.Extension)
	relativePath := path.Join(strings.TrimPrefix(folderPath, "/"), fileName)
	targetPath := filepath.Join(workingDirectory, filepath.FromSlash(relativePath))
	resolvedRelativePath, resolvedTargetPath, err := trackAssetPaths(workingDirectory, targetPath)
	if err != nil {
		return "", "", err
	}
	return resolvedTargetPath, resolvedRelativePath, nil
}

func normalizeAYONFolderPath(folderPath string) string {
	folderPath = strings.ReplaceAll(strings.TrimSpace(folderPath), "\\", "/")
	folderPath = "/" + strings.Trim(folderPath, "/")
	if folderPath == "/" {
		return folderPath
	}
	return path.Clean(folderPath) + "/"
}

func resolveAnyMappedAYONAsset(
	tx *sqlx.Tx,
	context AYONWorkfileContext,
	userID string,
	role models.Role,
) (models.Asset, bool, error) {
	candidates := []struct {
		integrationID string
		externalID    string
	}{
		{ayonDCCIntegrationID, ayonScopedWorkfileExternalID(context)},
		{ayonIntegrationID, ayonScopedWorkfileExternalID(context)},
		{ayonIntegrationID, strings.TrimSpace(context.TaskID)},
		{ayonIntegrationID, ayonWorkfileExternalID(context.TaskID, context.Application)},
	}
	seen := map[string]bool{}
	for _, candidate := range candidates {
		key := candidate.integrationID + "\x00" + candidate.externalID
		if candidate.externalID == "" || seen[key] {
			continue
		}
		seen[key] = true
		asset, found, err := resolveMappedAYONAssetForIntegration(
			tx,
			candidate.integrationID,
			candidate.externalID,
			userID,
			role,
		)
		if err != nil || found {
			return asset, found, err
		}
	}
	return models.Asset{}, false, nil
}

func ayonScopedWorkfileExternalID(context AYONWorkfileContext) string {
	parts := []string{
		strings.TrimSpace(context.ProjectID),
		strings.TrimSpace(context.TaskID),
		strings.ToLower(strings.TrimSpace(context.Application)),
	}
	return strings.Join(parts, ayonWorkfileKeySeparator)
}

func ensureAYONFolderMappings(
	tx *sqlx.Tx,
	context AYONWorkfileContext,
	leafCollectionID string,
) error {
	folders := normalizedAYONFolders(context)
	if len(folders) == 0 {
		return errors.New("AYON folder hierarchy is required")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	for index, folder := range folders {
		if folder.ID == "" || folder.Name == "" || folder.Path == "/" {
			return errors.New("AYON folder hierarchy contains an incomplete folder")
		}
		collection, err := repository.GetCollectionByPath(tx, normalizeAYONFolderPath(folder.Path))
		if err != nil {
			return err
		}
		if index == len(folders)-1 && collection.Id != leafCollectionID {
			return errors.New("AYON folder hierarchy did not resolve to the target collection")
		}
		if index == len(folders)-1 && folder.ID != context.FolderID {
			return errors.New("AYON folder hierarchy does not end at the selected folder")
		}
		externalID := ayonScopedFolderExternalID(context.ProjectID, folder.ID)
		if _, err = repository.GetCollectionMappingByExternalId(tx, ayonDCCIntegrationID, externalID); err == nil {
			continue
		} else if !errors.Is(err, repository.ErrCollectionMappingNotFound) {
			return err
		}
		externalParentID := ""
		if folder.ParentID != "" {
			externalParentID = ayonScopedFolderExternalID(context.ProjectID, folder.ParentID)
		}
		_, err = repository.CreateCollectionMapping(
			tx,
			uuid.NewString(),
			ayonDCCIntegrationID,
			externalID,
			folder.Type,
			folder.Name,
			externalParentID,
			normalizeAYONFolderPath(folder.Path),
			"{}",
			collection.Id,
			now,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func normalizedAYONFolders(context AYONWorkfileContext) []AYONFolderContext {
	if len(context.Folders) > 0 {
		folders := make([]AYONFolderContext, 0, len(context.Folders))
		for _, folder := range context.Folders {
			folder.Path = normalizeAYONFolderPath(folder.Path)
			folder.Name = strings.TrimSpace(folder.Name)
			folders = append(folders, folder)
		}
		return folders
	}

	paths := utils.GetCollectionPaths(normalizeAYONFolderPath(context.FolderPath))
	folders := make([]AYONFolderContext, 0, len(paths))
	for index, folderPath := range paths {
		folderID := "path:" + folderPath
		if index == len(paths)-1 {
			folderID = context.FolderID
		}
		parentID := ""
		if index > 0 {
			parentID = folders[index-1].ID
		}
		folders = append(folders, AYONFolderContext{
			ID:       folderID,
			ParentID: parentID,
			Name:     path.Base(strings.TrimSuffix(folderPath, "/")),
			Path:     folderPath,
			Type:     "folder",
		})
	}
	return folders
}

func ayonScopedFolderExternalID(projectID, folderID string) string {
	return strings.Join(
		[]string{strings.TrimSpace(projectID), "folder", strings.TrimSpace(folderID)},
		ayonWorkfileKeySeparator,
	)
}

func createAYONAssetMapping(tx *sqlx.Tx, context AYONWorkfileContext, assetID, userID string) error {
	metadata, err := json.Marshal(map[string]string{
		"application": context.Application,
		"project_id":  context.ProjectID,
		"task_id":     context.TaskID,
	})
	if err != nil {
		return err
	}
	assignees, err := json.Marshal([]string{userID})
	if err != nil {
		return err
	}
	_, err = repository.CreateAssetMapping(
		tx,
		uuid.NewString(),
		ayonDCCIntegrationID,
		ayonScopedWorkfileExternalID(context),
		context.TaskName,
		ayonScopedFolderExternalID(context.ProjectID, context.FolderID),
		context.TaskType,
		"",
		string(assignees),
		string(metadata),
		assetID,
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}
