package repository

import (
	"clustta/internal/auth_service"
	"clustta/internal/settings"
	"clustta/internal/utils"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

//go:embed project_templates.json
var templateJSONData []byte

//go:embed template_scripts/maya_prelaunch.py
var mayaPreLaunchScript []byte

//go:embed template_scripts/blender_prelaunch.py
var blenderPreLaunchScript []byte

type ProjectTemplateDefinition struct {
	Name            string                  `json:"name"`
	Icon            string                  `json:"icon"`
	Description     string                  `json:"description"`
	AssetTypes      []TemplateType          `json:"assetTypes"`
	CollectionTypes []TemplateType          `json:"collectionTypes"`
	IgnoreList      []string                `json:"ignoreList"`
	PreLaunchHooks  []TemplatePreLaunchHook `json:"preLaunchHooks"`
}

type TemplatePreLaunchHook struct {
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"`
	Script     string   `json:"script"`
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
		log.Printf("Failed to get templates path: %v", err)
		return fmt.Errorf("failed to get templates path: %w", err)
	}

	if err := os.MkdirAll(templatesPath, os.ModePerm); err != nil {
		log.Printf("Failed to create templates directory: %v", err)
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	hasTemplates, err := hasDefaultTemplates(templatesPath)
	if err != nil {
		log.Printf("Failed to check existing templates: %v", err)
		return fmt.Errorf("failed to check existing templates: %w", err)
	}

	if hasTemplates {
		log.Printf("Default templates already exist at %s", templatesPath)
		return nil
	}

	log.Printf("No default templates found, creating them at %s", templatesPath)

	var activeUser auth_service.User
	if user != nil {
		activeUser = *user
	} else {
		activeUser, err = auth_service.GetActiveUser()
		if err != nil {
			log.Printf("Failed to get active user: %v", err)
			return fmt.Errorf("failed to get active user: %w", err)
		}
	}

	templateDefs, err := loadTemplateDefinitions()
	if err != nil {
		log.Printf("Failed to load template definitions: %v", err)
		return fmt.Errorf("failed to load template definitions: %w", err)
	}

	for _, templateDef := range templateDefs.Templates {
		templatePath := filepath.Join(templatesPath, templateDef.Name+".clst")

		if utils.FileExists(templatePath) {
			log.Printf("Template %s already exists, skipping", templateDef.Name)
			continue
		}

		log.Printf("Creating template: %s", templateDef.Name)
		err = createDefaultTemplate(templatePath, templateDef, activeUser)
		if err != nil {
			log.Printf("Failed to create template %s: %v", templateDef.Name, err)
			return fmt.Errorf("failed to create template %s: %w", templateDef.Name, err)
		}
		log.Printf("Successfully created template: %s", templateDef.Name)
	}

	return nil
}

// hasDefaultTemplates reports whether any valid .clst template exists in the directory.
// Corrupt or partially-created templates are quarantined so the defaults can be rebuilt.
func hasDefaultTemplates(templatesPath string) (bool, error) {
	entries, err := os.ReadDir(templatesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".clst") {
			continue
		}

		path := filepath.Join(templatesPath, entry.Name())
		ok, err := VerifyProjectIntegrity(path)
		if err != nil || !ok {
			log.Printf("Removing corrupt template %s (integrity check failed: %v)", path, err)
			os.Remove(path)
			os.Remove(path + "-wal")
			os.Remove(path + "-shm")
			continue
		}

		return true, nil
	}

	return false, nil
}

// loadTemplateDefinitions reads and parses the embedded project_templates.json file
func loadTemplateDefinitions() (*ProjectTemplatesConfig, error) {
	var config ProjectTemplatesConfig
	err := json.Unmarshal(templateJSONData, &config)
	if err != nil {
		log.Printf("Failed to parse templates JSON: %v", err)
		return nil, fmt.Errorf("failed to parse templates JSON: %w", err)
	}

	return &config, nil
}

// createDefaultTemplate creates a new template .clst file with the specified metadata.
// It builds in a temp directory and atomically renames once verified, so a partial template is never published.
func createDefaultTemplate(templatePath string, templateDef ProjectTemplateDefinition, user auth_service.User) error {
	// Build the template inside a temporary sibling directory using its final
	// base name, so the derived project name stays correct (e.g. "Animation",
	// not "Animation.clst.tmp-..."). The final rename within the same templates
	// directory is atomic.
	tmpDir := filepath.Join(filepath.Dir(templatePath), ".tmp-"+uuid.NewString())
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		log.Printf("Failed to create temp template directory: %v", err)
		return fmt.Errorf("failed to create temp template directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpPath := filepath.Join(tmpDir, filepath.Base(templatePath))

	_, err := CreateProject(tmpPath, "Personal", "", "", "", user)
	if err != nil {
		log.Printf("Failed to create template project: %v", err)
		return fmt.Errorf("failed to create template project: %w", err)
	}

	db, err := utils.OpenDb(tmpPath)
	if err != nil {
		log.Printf("Failed to open template database: %v", err)
		return fmt.Errorf("failed to open template database: %w", err)
	}
	defer db.Close()

	tx, err := db.Beginx()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	err = utils.SetProjectIcon(tx, templateDef.Icon)
	if err != nil {
		log.Printf("Failed to set template icon: %v", err)
		return fmt.Errorf("failed to set template icon: %w", err)
	}

	for _, assetType := range templateDef.AssetTypes {
		_, err = GetOrCreateAssetType(tx, assetType.Name, assetType.Icon)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				continue
			}
			log.Printf("Failed to create asset type %s: %v", assetType.Name, err)
			return fmt.Errorf("failed to create asset type %s: %w", assetType.Name, err)
		}
	}

	for _, collectionType := range templateDef.CollectionTypes {
		_, err = GetOrCreateCollectionType(tx, collectionType.Name, collectionType.Icon)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				continue
			}
			log.Printf("Failed to create collection type %s: %v", collectionType.Name, err)
			return fmt.Errorf("failed to create collection type %s: %w", collectionType.Name, err)
		}
	}

	err = utils.SetProjectIgnoreList(tx, templateDef.IgnoreList)
	if err != nil {
		log.Printf("Failed to set ignore list: %v", err)
		return fmt.Errorf("failed to set ignore list: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Failed to commit template: %v", err)
		return fmt.Errorf("failed to commit template: %w", err)
	}

	// Ensure the WAL is checkpointed into the main file before publishing.
	if err := db.Close(); err != nil {
		log.Printf("Failed to close template database: %v", err)
		return fmt.Errorf("failed to close template database: %w", err)
	}
	if err := applyTemplatePreLaunchHooks(tmpPath, tmpPath, user.Id); err != nil {
		return fmt.Errorf("failed to add template launch hooks: %w", err)
	}

	// Verify the template is fully initialized before publishing it.
	ok, err := VerifyProjectIntegrity(tmpPath)
	if err != nil || !ok {
		log.Printf("Template %s failed integrity check: %v", templateDef.Name, err)
		return fmt.Errorf("template %s failed integrity check: %w", templateDef.Name, err)
	}

	// Atomically publish the completed template into the templates directory.
	if err := os.Rename(tmpPath, templatePath); err != nil {
		log.Printf("Failed to publish template %s: %v", templateDef.Name, err)
		return fmt.Errorf("failed to publish template %s: %w", templateDef.Name, err)
	}

	return nil
}

func applyTemplatePreLaunchHooks(projectPath, templatePath, authorID string) error {
	templateName := strings.TrimSuffix(filepath.Base(templatePath), filepath.Ext(templatePath))
	templateDefs, err := loadTemplateDefinitions()
	if err != nil {
		return err
	}
	var hooks []TemplatePreLaunchHook
	for _, templateDef := range templateDefs.Templates {
		if templateDef.Name == templateName {
			hooks = templateDef.PreLaunchHooks
			break
		}
	}
	if len(hooks) == 0 {
		return nil
	}

	db, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	workingDir, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return err
	}
	scriptsDir := filepath.Join(workingDir, "Scripts")
	if err := os.MkdirAll(scriptsDir, os.ModePerm); err != nil {
		return err
	}
	collectionType, err := GetOrCreateCollectionType(tx, "Scripts", "code-bracket")
	if err != nil {
		return err
	}
	assetType, err := GetOrCreateAssetType(tx, "Script", "code-bracket")
	if err != nil {
		return err
	}
	scriptsCollection, err := CreateCollection(tx, "", "Scripts", "Project launch scripts", collectionType.Id, "", "", false)
	if err != nil {
		return err
	}
	if authorID == "" {
		user, err := auth_service.GetActiveUser()
		if err != nil {
			return err
		}
		authorID = user.Id
	}

	settings := PreLaunchHookSettings{Version: PreLaunchHooksVersion}
	for _, hookDef := range hooks {
		scriptContent, err := templatePreLaunchScript(hookDef.Script)
		if err != nil {
			return err
		}
		scriptPath := filepath.Join(scriptsDir, hookDef.Script)
		if err := os.WriteFile(scriptPath, scriptContent, 0o644); err != nil {
			return err
		}
		asset, err := CreateAsset(
			tx, "", strings.TrimSuffix(hookDef.Script, filepath.Ext(hookDef.Script)), assetType.Id,
			scriptsCollection.Id, false, "", "", scriptPath, nil, "", false, "", authorID,
			"Created from project template", uuid.NewString(), func(int, int, string, string) {},
		)
		if err != nil {
			return err
		}
		settings.Hooks = append(settings.Hooks, PreLaunchHook{
			Name: hookDef.Name, Enabled: true, Extensions: hookDef.Extensions,
			ScriptAssetIDs: []string{asset.Id}, FailurePolicy: PreLaunchFailureBlock,
		})
	}
	if err := SetPreLaunchHookSettings(tx, settings); err != nil {
		return err
	}
	return tx.Commit()
}

func templatePreLaunchScript(name string) ([]byte, error) {
	switch name {
	case "maya_prelaunch.py":
		return mayaPreLaunchScript, nil
	case "blender_prelaunch.py":
		return blenderPreLaunchScript, nil
	default:
		return nil, fmt.Errorf("unknown template pre-launch script %q", name)
	}
}

// ResetDefaultTemplates removes all existing templates and recreates the defaults.
// This can be used as a "Reset to Factory Defaults" feature.
func ResetDefaultTemplates() error {
	templatesPath, err := settings.GetUserProjectTemplatesPath()
	if err != nil {
		log.Printf("Failed to get templates path: %v", err)
		return fmt.Errorf("failed to get templates path: %w", err)
	}

	templateDefs, err := loadTemplateDefinitions()
	if err != nil {
		log.Printf("Failed to load template definitions: %v", err)
		return fmt.Errorf("failed to load template definitions: %w", err)
	}

	for _, templateDef := range templateDefs.Templates {
		templatePath := filepath.Join(templatesPath, templateDef.Name+".clst")
		if utils.FileExists(templatePath) {
			err = os.Remove(templatePath)
			if err != nil {
				log.Printf("Failed to remove template %s: %v", templateDef.Name, err)
				return fmt.Errorf("failed to remove template %s: %w", templateDef.Name, err)
			}
		}
	}

	return InitializeDefaultTemplates(nil)
}
