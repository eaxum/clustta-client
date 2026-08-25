package repository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"clustta/internal/auth_service"
	"clustta/internal/utils"

	"github.com/stretchr/testify/require"
)

func TestThreeDimensionalAnimationTemplateIncludesLaunchScripts(t *testing.T) {
	templates, err := loadTemplateDefinitions()
	require.NoError(t, err)

	for _, template := range templates.Templates {
		if template.Name != "3D Animation" {
			continue
		}
		require.Contains(t, template.CollectionTypes, TemplateType{Name: "Scripts", Icon: "code-bracket"})
		require.Len(t, template.PreLaunchHooks, 2)
		for _, hook := range template.PreLaunchHooks {
			content, err := templatePreLaunchScript(hook.Script)
			require.NoError(t, err)
			require.NotEmpty(t, content)
		}
		return
	}
	t.Fatal("3D Animation template not found")
}

func TestMayaProductionTemplateIncludesWorkspaceSetup(t *testing.T) {
	templates, err := loadTemplateDefinitions()
	require.NoError(t, err)

	for _, template := range templates.Templates {
		if template.Name != "Maya 3D Animation Production" {
			continue
		}
		require.Equal(t, []TemplatePreLaunchHook{
			{
				Name:       "Maya - Configure Production Workspace",
				Extensions: []string{".ma", ".mb"},
				Script:     "maya_production_prelaunch.py",
			},
		}, template.PreLaunchHooks)
		require.Equal(t, []TemplateProjectFile{
			{
				Name:      "workspace.mel",
				Source:    "maya_workspace.mel",
				AssetType: "Configuration",
			},
		}, template.ProjectFiles)
		require.Contains(t, template.CollectionTypes, TemplateType{Name: "Maps", Icon: "texture"})
		require.NotEmpty(t, template.Collections)

		workspaceContent, err := templateProjectFile(template.ProjectFiles[0].Source)
		require.NoError(t, err)
		require.Contains(t, string(workspaceContent), `workspace -fr "sourceImages" "Assets";`)

		scriptContent, err := templatePreLaunchScript(template.PreLaunchHooks[0].Script)
		require.NoError(t, err)
		require.NotContains(t, strings.ToLower(string(scriptContent)), "makedirs")
		return
	}
	t.Fatal("Maya 3D Animation Production template not found")
}

func TestApplyMayaProductionTemplateSetup(t *testing.T) {
	templates, err := loadTemplateDefinitions()
	require.NoError(t, err)

	var definition ProjectTemplateDefinition
	for _, template := range templates.Templates {
		if template.Name == "Maya 3D Animation Production" {
			definition = template
			break
		}
	}
	require.NotEmpty(t, definition.Name)

	workingDirectory := filepath.Join(t.TempDir(), "working")
	projectPath := filepath.Join(t.TempDir(), "project.clst")
	user := auth_service.OfflineUser()
	setupDB, err := utils.OpenDb(projectPath)
	require.NoError(t, err)
	defer setupDB.Close()
	_, err = setupDB.Exec(ProjectSchema)
	require.NoError(t, err)
	tx, err := setupDB.Beginx()
	require.NoError(t, err)
	require.NoError(t, utils.SetProjectWorkingDir(tx, workingDirectory))
	require.NoError(t, initData(tx))
	role, err := GetRoleByName(tx, "admin")
	require.NoError(t, err)
	_, err = AddKnownUser(
		tx, user.Id, user.Email, user.Username, user.FirstName, user.LastName, role.Id, nil, false,
	)
	require.NoError(t, err)
	for _, assetType := range definition.AssetTypes {
		_, err = GetOrCreateAssetType(tx, assetType.Name, assetType.Icon)
		require.NoError(t, err)
	}
	for _, collectionType := range definition.CollectionTypes {
		_, err = GetOrCreateCollectionType(tx, collectionType.Name, collectionType.Icon)
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit())
	require.NoError(t, setupDB.Close())

	templatePath := filepath.Join(t.TempDir(), definition.Name+".clst")
	require.NoError(t, applyTemplateProjectSetup(projectPath, templatePath, user.Id))
	require.FileExists(t, filepath.Join(workingDirectory, "workspace.mel"))
	require.FileExists(t, filepath.Join(workingDirectory, "Scripts", "maya_production_prelaunch.py"))
	require.DirExists(t, filepath.Join(workingDirectory, "Assets", "Characters"))
	require.DirExists(t, filepath.Join(workingDirectory, "Production", "Cache", "diskCache"))

	projectDB, err := utils.OpenDb(projectPath)
	require.NoError(t, err)
	defer projectDB.Close()
	tx, err = projectDB.Beginx()
	require.NoError(t, err)
	defer tx.Rollback()
	workspaceAsset, err := GetAssetByName(tx, "workspace", "", ".mel")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(workingDirectory, "workspace.mel"), workspaceAsset.GetFilePath())
	hookSettings, err := GetPreLaunchHookSettings(tx)
	require.NoError(t, err)
	require.Len(t, hookSettings.Hooks, 1)
	require.Equal(t, []string{".ma", ".mb"}, hookSettings.Hooks[0].Extensions)

	content, err := os.ReadFile(filepath.Join(workingDirectory, "workspace.mel"))
	require.NoError(t, err)
	require.Equal(t, mayaWorkspaceFile, content)
}
