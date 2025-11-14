package repository

import (
	"clustta/internal/auth_service"
	"clustta/internal/settings"
	"clustta/internal/utils"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed project_templates.json
var templateJSONData []byte

type ProjectTemplateDefinition struct {
	Name            string         `json:"name"`
	Icon            string         `json:"icon"`
	Description     string         `json:"description"`
	AssetTypes      []TemplateType `json:"assetTypes"`
	CollectionTypes []TemplateType `json:"collectionTypes"`
	IgnoreList      []string       `json:"ignoreList"`
}

type TemplateType struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type ProjectTemplatesConfig struct {
	Templates []ProjectTemplateDefinition `json:"templates"`
}

// Creates default project templates if they don't exist.
// Accepts an optional user; if nil, attempts to get the active user.
func InitializeDefaultTemplates(user *auth_service.User) error {
	templatesPath, err := settings.GetUserProjectTemplatesPath()
	if err != nil {
		return fmt.Errorf("failed to get templates path: %w", err)
	}

	if err := os.MkdirAll(templatesPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	hasTemplates, err := hasDefaultTemplates(templatesPath)
	if err != nil {
		return fmt.Errorf("failed to check existing templates: %w", err)
	}

	if hasTemplates {
		return nil
	}

	var activeUser auth_service.User
	if user != nil {
		activeUser = *user
	} else {
		activeUser, err = auth_service.GetActiveUser()
		if err != nil {
			return fmt.Errorf("failed to get active user: %w", err)
		}
	}

	templateDefs, err := loadTemplateDefinitions()
	if err != nil {
		return fmt.Errorf("failed to load template definitions: %w", err)
	}

	for _, templateDef := range templateDefs.Templates {
		templatePath := filepath.Join(templatesPath, templateDef.Name+".clst")

		if utils.FileExists(templatePath) {
			continue
		}

		err = createDefaultTemplate(templatePath, templateDef, activeUser)
		if err != nil {
			return fmt.Errorf("failed to create template %s: %w", templateDef.Name, err)
		}
	}

	return nil
}

// hasDefaultTemplates checks if any .clst template files exist in the directory
func hasDefaultTemplates(templatesPath string) (bool, error) {
	entries, err := os.ReadDir(templatesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".clst") {
			return true, nil
		}
	}

	return false, nil
}

// loadTemplateDefinitions reads and parses the embedded project_templates.json file
func loadTemplateDefinitions() (*ProjectTemplatesConfig, error) {
	var config ProjectTemplatesConfig
	err := json.Unmarshal(templateJSONData, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates JSON: %w", err)
	}

	return &config, nil
}

// createDefaultTemplate creates a new template .clst file with the specified metadata
func createDefaultTemplate(templatePath string, templateDef ProjectTemplateDefinition, user auth_service.User) error {
	_, err := CreateProject(templatePath, "Personal", "", "", user)
	if err != nil {
		return fmt.Errorf("failed to create template project: %w", err)
	}

	db, err := utils.OpenDb(templatePath)
	if err != nil {
		return fmt.Errorf("failed to open template database: %w", err)
	}
	defer db.Close()

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	err = utils.SetProjectIcon(tx, templateDef.Icon)
	if err != nil {
		return fmt.Errorf("failed to set template icon: %w", err)
	}

	for _, assetType := range templateDef.AssetTypes {
		_, err = GetOrCreateTaskType(tx, assetType.Name, assetType.Icon)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				continue
			}
			return fmt.Errorf("failed to create asset type %s: %w", assetType.Name, err)
		}
	}

	for _, collectionType := range templateDef.CollectionTypes {
		_, err = GetOrCreateEntityType(tx, collectionType.Name, collectionType.Icon)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				continue
			}
			return fmt.Errorf("failed to create collection type %s: %w", collectionType.Name, err)
		}
	}

	err = utils.SetProjectIgnoreList(tx, templateDef.IgnoreList)
	if err != nil {
		return fmt.Errorf("failed to set ignore list: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit template: %w", err)
	}

	return nil
}

// ResetDefaultTemplates removes all existing templates and recreates the defaults.
// This can be used as a "Reset to Factory Defaults" feature.
func ResetDefaultTemplates() error {
	templatesPath, err := settings.GetUserProjectTemplatesPath()
	if err != nil {
		return fmt.Errorf("failed to get templates path: %w", err)
	}

	templateDefs, err := loadTemplateDefinitions()
	if err != nil {
		return fmt.Errorf("failed to load template definitions: %w", err)
	}

	for _, templateDef := range templateDefs.Templates {
		templatePath := filepath.Join(templatesPath, templateDef.Name+".clst")
		if utils.FileExists(templatePath) {
			err = os.Remove(templatePath)
			if err != nil {
				return fmt.Errorf("failed to remove template %s: %w", templateDef.Name, err)
			}
		}
	}

	return InitializeDefaultTemplates(nil)
}
