package services

import (
	"encoding/json"
	"errors"
	"fmt"
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
	TaskID      string
	TaskName    string
	TaskTypeID  string
	TaskType    string
	Application string
	Extension   string
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

	externalID := ayonWorkfileExternalID(context.TaskID, context.Application)
	existing, found, err := resolveMappedAYONAsset(tx, externalID, userID, role)
	if err != nil {
		return AYONWorkfileResult{}, err
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
	mapping, err := repository.GetAssetMappingByExternalId(tx, ayonIntegrationID, externalID)
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
