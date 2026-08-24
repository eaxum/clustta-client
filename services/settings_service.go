package services

import (
	"clustta/internal/agent"
	"clustta/internal/repository"
	"clustta/internal/settings"
	"clustta/internal/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
)

type SettingsService struct{}

func requireChangeRolePermission(tx *sqlx.Tx) error {
	_, role, err := activeAssetRole(tx)
	if err != nil {
		return err
	}
	if !role.ChangeRole {
		return fmt.Errorf("user does not have change_role permission")
	}
	return nil
}

func (s *SettingsService) GetProjectScriptSettings(projectPath string) (repository.ProjectScriptSettings, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return repository.ProjectScriptSettings{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return repository.ProjectScriptSettings{}, err
	}
	defer tx.Rollback()
	return repository.GetProjectScriptSettings(tx)
}

func (s *SettingsService) SetProjectScriptSettings(projectPath, directory string, extensions []string) error {
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
	if err := requireChangeRolePermission(tx); err != nil {
		return err
	}
	if err := repository.SetProjectScriptSettings(tx, repository.ProjectScriptSettings{
		Directory: directory, Extensions: extensions,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SettingsService) GetPreLaunchHookSettings(projectPath string) (repository.PreLaunchHookSettings, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return repository.PreLaunchHookSettings{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return repository.PreLaunchHookSettings{}, err
	}
	defer tx.Rollback()
	return repository.GetPreLaunchHookSettings(tx)
}

func (s *SettingsService) SetPreLaunchHookSettings(projectPath string, settings repository.PreLaunchHookSettings) (repository.PreLaunchHookSettings, error) {
	normalized, err := repository.NormalizePreLaunchHookSettings(settings)
	if err != nil {
		return repository.PreLaunchHookSettings{}, err
	}
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return repository.PreLaunchHookSettings{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return repository.PreLaunchHookSettings{}, err
	}
	defer tx.Rollback()
	if err := requireChangeRolePermission(tx); err != nil {
		return repository.PreLaunchHookSettings{}, err
	}
	for _, hook := range normalized.Hooks {
		scriptPaths, err := agent.ResolveTrackedScriptPaths(projectPath, hook.ScriptAssetIDs)
		if err != nil {
			return repository.PreLaunchHookSettings{}, err
		}
		for _, scriptPath := range scriptPaths {
			extension := strings.ToLower(filepath.Ext(scriptPath))
			for _, targetExtension := range hook.Extensions {
				dcc := repository.PreLaunchDCCForExtension(targetExtension)
				if dcc == "" {
					return repository.PreLaunchHookSettings{}, fmt.Errorf("script hooks do not support %s files", targetExtension)
				}
				if dcc == repository.PreLaunchDCCBlender && extension != ".py" {
					return repository.PreLaunchHookSettings{}, fmt.Errorf("Blender hooks only support Python scripts")
				}
				if dcc == repository.PreLaunchDCCMaya && extension != ".py" && extension != ".mel" {
					return repository.PreLaunchHookSettings{}, fmt.Errorf("Maya hooks only support Python and MEL scripts")
				}
			}
		}
	}
	if err := repository.SetPreLaunchHookSettings(tx, normalized); err != nil {
		return repository.PreLaunchHookSettings{}, err
	}
	if err := tx.Commit(); err != nil {
		return repository.PreLaunchHookSettings{}, err
	}
	return s.GetPreLaunchHookSettings(projectPath)
}

func (s *SettingsService) SetProjectEnvironmentVariables(projectPath string, requestedSettings repository.PreLaunchHookSettings) (repository.PreLaunchHookSettings, error) {
	environmentVariables := requestedSettings.EnvironmentVariables
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return repository.PreLaunchHookSettings{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return repository.PreLaunchHookSettings{}, err
	}
	defer tx.Rollback()
	if err := requireChangeRolePermission(tx); err != nil {
		return repository.PreLaunchHookSettings{}, err
	}
	settings, err := repository.GetPreLaunchHookSettings(tx)
	if err != nil {
		return repository.PreLaunchHookSettings{}, err
	}
	availableIDs := make(map[string]bool, len(environmentVariables))
	for _, environmentVariable := range environmentVariables {
		availableIDs[strings.TrimSpace(environmentVariable.ID)] = true
	}
	for index := range settings.Hooks {
		selectedIDs := settings.Hooks[index].EnvironmentVariableIDs[:0]
		for _, environmentVariableID := range settings.Hooks[index].EnvironmentVariableIDs {
			if availableIDs[environmentVariableID] {
				selectedIDs = append(selectedIDs, environmentVariableID)
			}
		}
		settings.Hooks[index].EnvironmentVariableIDs = selectedIDs
	}
	settings.EnvironmentVariables = environmentVariables
	if err := repository.SetPreLaunchHookSettings(tx, settings); err != nil {
		return repository.PreLaunchHookSettings{}, err
	}
	if err := tx.Commit(); err != nil {
		return repository.PreLaunchHookSettings{}, err
	}
	return s.GetPreLaunchHookSettings(projectPath)
}

var (
	bridgeLifecycleMu sync.RWMutex
	startBridge       = func() {}
	stopBridge        = func() {}
)

func ConfigureBridgeLifecycle(start, stop func()) {
	bridgeLifecycleMu.Lock()
	defer bridgeLifecycleMu.Unlock()
	startBridge = start
	stopBridge = stop
}

// GetStudios retrieves all configured studios from user settings.
func (s *SettingsService) GetStudios(path string) ([]settings.Studio, error) {
	studios, err := settings.GetStudios()
	if err != nil {
		return studios, err
	}
	return studios, nil
}

// PinProject pins a project to the studio's favorites list.
func (s *SettingsService) PinProject(studioName, projectId string) ([]string, error) {
	projectsId, err := settings.PinProject(studioName, projectId)
	if err != nil {
		return projectsId, err
	}
	return projectsId, nil
}

// GetUserDirectory returns the current user's home directory path.
func (s *SettingsService) GetUserDirectory() (string, error) {
	return settings.GetUserDirectory()
}

// GetLogPath returns the path to the Clustta log file.
func (s *SettingsService) GetLogPath() (string, error) {
	return settings.GetLogPath()
}

// GetLogContents reads and returns the last N lines from the log file.
func (s *SettingsService) GetLogContents(maxLines int) ([]string, error) {
	logPath, err := settings.GetLogPath()
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	lines := strings.Split(string(content), "\n")

	// Filter out empty lines
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			filtered = append(filtered, line)
		}
	}

	// Return only the last maxLines
	if maxLines > 0 && len(filtered) > maxLines {
		filtered = filtered[len(filtered)-maxLines:]
	}

	return filtered, nil
}

// GetUsername returns the current system username.
func (s *SettingsService) GetUsername() (string, error) {
	return settings.GetUsername()
}

// UnpinProject removes a project from the studio's favorites list.
func (s *SettingsService) UnpinProject(studioName, projectId string) ([]string, error) {
	projectsId, err := settings.UnpinProject(studioName, projectId)
	if err != nil {
		return projectsId, err
	}
	return projectsId, nil
}

// GetPinnedProjects retrieves all pinned projects for the specified studio.
func (s *SettingsService) GetPinnedProjects(studioName string) ([]string, error) {
	projectsId, err := settings.GetPinnedProjects(studioName)
	if err != nil {
		return projectsId, err
	}
	return projectsId, nil
}

// GetRecentProjects retrieves recently accessed projects for the specified studio.
func (s *SettingsService) GetRecentProjects(studioName string) ([]string, error) {
	projectsId, err := settings.GetRecentProjects(studioName)
	if err != nil {
		return projectsId, err
	}
	return projectsId, nil
}

// AddRecentProject adds a project to the recent projects list.
func (s *SettingsService) AddRecentProject(studioName, projectId string) ([]string, error) {
	projectsId, err := settings.AddRecentProject(studioName, projectId)
	if err != nil {
		return projectsId, err
	}
	return projectsId, nil
}

// ClearRecentProject clears the recent projects list for the specified studio.
func (s *SettingsService) ClearRecentProject(studioName string) error {
	return settings.ClearRecentProject()
}

// AddProjectWorkspace adds a workspace configuration to a project.
func (s *SettingsService) AddProjectWorkspace(projectId string, workspaceData interface{}) error {
	err := settings.AddProjectWorkspace(projectId, workspaceData)
	if err != nil {
		return err
	}
	return nil
}

// RemoveProjectWorkspace removes a workspace configuration from a project.
func (s *SettingsService) RemoveProjectWorkspace(projectId string, workspaceName string) error {
	err := settings.RemoveProjectWorkspace(projectId, workspaceName)
	if err != nil {
		return err
	}
	return nil
}

// UpdateProjectWorkspace replaces an existing workspace configuration by name.
func (s *SettingsService) UpdateProjectWorkspace(projectId string, workspaceName string, workspaceData interface{}) error {
	err := settings.UpdateProjectWorkspace(projectId, workspaceName, workspaceData)
	if err != nil {
		return err
	}
	return nil
}

// GetProjectWorkspaces retrieves all workspace configurations for a project.
func (s *SettingsService) GetProjectWorkspaces(projectId string) ([]interface{}, error) {
	projectWorkspaces, err := settings.GetProjectWorkspaces(projectId)
	if err != nil {
		return projectWorkspaces, err
	}
	return projectWorkspaces, nil
}

// AddDependencyPreset adds a dependency preset to a project.
func (s *SettingsService) AddDependencyPreset(projectId string, presetData interface{}) error {
	err := settings.AddDependencyPreset(projectId, presetData)
	if err != nil {
		return err
	}
	return nil
}

// RemoveDependencyPreset removes a dependency preset from a project.
func (s *SettingsService) RemoveDependencyPreset(projectId string, presetName string) error {
	err := settings.RemoveDependencyPreset(projectId, presetName)
	if err != nil {
		return err
	}
	return nil
}

// UpdateDependencyPreset updates an existing dependency preset.
func (s *SettingsService) UpdateDependencyPreset(projectId string, presetName string, updatedPreset interface{}) error {
	err := settings.UpdateDependencyPreset(projectId, presetName, updatedPreset)
	if err != nil {
		return err
	}
	return nil
}

// GetProjectDependencyPresets retrieves all dependency presets for a project.
func (s *SettingsService) GetProjectDependencyPresets(projectId string) ([]interface{}, error) {
	presets, err := settings.GetProjectDependencyPresets(projectId)
	if err != nil {
		return presets, err
	}
	return presets, nil
}

// GetIgnoreListPresets retrieves all user-defined ignore list presets.
func (s *SettingsService) GetIgnoreListPresets() (map[string][]string, error) {
	presets, err := settings.GetIgnoreListPresets()
	if err != nil {
		return presets, err
	}
	return presets, nil
}

// AddIgnoreListPreset stores an ignore list preset under the given name.
// If a preset with the same name already exists, it is overwritten.
func (s *SettingsService) AddIgnoreListPreset(name string, entries []string) error {
	return settings.AddIgnoreListPreset(name, entries)
}

// RemoveIgnoreListPreset deletes a user-defined ignore list preset by name.
func (s *SettingsService) RemoveIgnoreListPreset(name string) error {
	return settings.RemoveIgnoreListPreset(name)
}

// GetEulaAccepted retrieves whether the user has accepted the EULA.
func (s *SettingsService) GetEulaAccepted() (bool, error) {
	eulaAccepted, err := settings.GetEulaAccepted()
	if err != nil {
		return eulaAccepted, err
	}
	return eulaAccepted, nil
}

// SetEulaAccepted sets the EULA acceptance status.
func (s *SettingsService) SetEulaAccepted(eulaAccepted bool) error {
	err := settings.SetEulaAccepted(eulaAccepted)
	if err != nil {
		return err
	}
	return nil
}

// GetLastStudio retrieves the last active studio name.
func (s *SettingsService) GetLastStudio() (string, error) {
	lastStudioName, err := settings.GetLastStudio()
	if err != nil {
		return lastStudioName, err
	}
	return lastStudioName, nil
}

// SetLastStudio sets the last active studio name.
func (s *SettingsService) SetLastStudio(lastStudioName string) error {
	err := settings.SetLastStudio(lastStudioName)
	if err != nil {
		return err
	}
	return nil
}

// GetCurrentVersion retrieves the current application version number.
func (s *SettingsService) GetCurrentVersion() (string, error) {
	versionNumber, err := settings.GetCurrentVersion()
	if err != nil {
		return versionNumber, err
	}
	return versionNumber, nil
}

// SetCurrentVersion sets the current application version number.
func (s *SettingsService) SetCurrentVersion(versionNumber string) error {
	err := settings.SetCurrentVersion(versionNumber)
	if err != nil {
		return err
	}
	return nil
}

// IsProjectGridView returns whether project grid view is enabled.
func (s *SettingsService) IsProjectGridView() (bool, error) {
	return settings.IsProjectGridView()
}

// ToggleProjectGridView toggles between grid and list view for projects.
func (s *SettingsService) ToggleProjectGridView() error {
	return settings.ToggleProjectGridView()
}

// ToggleShowUntrackedProjects toggles visibility of untracked projects.
func (s *SettingsService) ToggleShowUntrackedProjects() error {
	return settings.ToggleShowUntrackedProjects()
}

// IsShowUntrackedProjects returns whether untracked projects are visible.
func (s *SettingsService) IsShowUntrackedProjects() (bool, error) {
	return settings.IsShowUntrackedProjects()
}

// GetIconScheme retrieves the current icon scheme name.
func (s *SettingsService) GetIconScheme() (string, error) {
	iconScheme, err := settings.GetIconScheme()
	if err != nil {
		return iconScheme, err
	}
	return iconScheme, nil
}

// SetIconScheme sets the icon scheme for the application.
func (s *SettingsService) SetIconScheme(iconScheme string) error {
	err := settings.SetIconScheme(iconScheme)
	if err != nil {
		return err
	}
	return nil
}

// GetTheme retrieves the current theme name.
func (s *SettingsService) GetTheme() (string, error) {
	theme, err := settings.GetTheme()
	if err != nil {
		return theme, err
	}
	return theme, nil
}

// SetTheme sets the application theme.
func (s *SettingsService) SetTheme(theme string) error {
	err := settings.SetTheme(theme)
	if err != nil {
		return err
	}
	return nil
}

// GetThemeTint retrieves the current theme tint name.
func (s *SettingsService) GetThemeTint() (string, error) {
	return settings.GetThemeTint()
}

// SetThemeTint sets the application theme tint.
func (s *SettingsService) SetThemeTint(tint string) error {
	return settings.SetThemeTint(tint)
}

// GetLanguage retrieves the user's language preference.
func (s *SettingsService) GetLanguage() (string, error) {
	language, err := settings.GetLanguage()
	if err != nil {
		return "", err
	}
	return language, nil
}

// SetLanguage sets the user's language preference.
func (s *SettingsService) SetLanguage(language string) error {
	err := settings.SetLanguage(language)
	if err != nil {
		return err
	}
	return nil
}

// GetUseGrid retrieves whether grid view is enabled.
func (s *SettingsService) GetUseGrid() (bool, error) {
	useGrid, err := settings.GetUseGrid()
	if err != nil {
		return useGrid, err
	}
	return useGrid, nil
}

// SetUseGrid sets whether to use grid view.
func (s *SettingsService) SetUseGrid(useGrid bool) error {
	err := settings.SetUseGrid(useGrid)
	if err != nil {
		return err
	}
	return nil
}

// GetDefaultViewMode retrieves the default view mode setting.
// Returns "compact" (list), "dense" (compact), or "grid".
func (s *SettingsService) GetDefaultViewMode() (string, error) {
	viewMode, err := settings.GetDefaultViewMode()
	if err != nil {
		return viewMode, err
	}
	return viewMode, nil
}

// SetDefaultViewMode sets the default view mode.
func (s *SettingsService) SetDefaultViewMode(viewMode string) error {
	err := settings.SetDefaultViewMode(viewMode)
	if err != nil {
		return err
	}
	return nil
}

// GetUntrackedVisibility retrieves whether untracked browser items are visible.
func (s *SettingsService) GetUntrackedVisibility() (bool, error) {
	return settings.GetUntrackedVisibility()
}

// SetUntrackedVisibility sets whether untracked browser items are visible.
func (s *SettingsService) SetUntrackedVisibility(enabled bool) error {
	return settings.SetUntrackedVisibility(enabled)
}

// GetProjectDirectory retrieves the default project directory path.
func (s *SettingsService) GetProjectDirectory() (string, error) {
	projectDir, err := settings.GetProjectDirectory()
	if err != nil {
		return projectDir, err
	}
	return projectDir, nil
}

// SetProjectDirectory sets the default project directory path.
func (s *SettingsService) SetProjectDirectory(dir string) error {
	err := settings.SetProjectDirectory(dir)
	if err != nil {
		return err
	}
	return nil
}

// GetSharedProjectDirectory retrieves the shared project directory path.
func (s *SettingsService) GetSharedProjectDirectory() (string, error) {
	projectDir, err := settings.GetSharedProjectDirectory()
	if err != nil {
		return projectDir, err
	}
	return projectDir, nil
}

// SetSharedProjectDirectory sets the shared project directory path.
func (s *SettingsService) SetSharedProjectDirectory(dir string) error {
	err := settings.SetSharedProjectDirectory(dir)
	if err != nil {
		return err
	}
	return nil
}

// GetWorkingDirectory retrieves the working directory path.
func (s *SettingsService) GetWorkingDirectory() (string, error) {
	projectDir, err := settings.GetWorkingDirectory()
	if err != nil {
		return projectDir, err
	}
	return projectDir, nil
}

// SetWorkingDirectory sets the working directory path.
func (s *SettingsService) SetWorkingDirectory(dir string) error {
	err := settings.SetWorkingDirectory(dir)
	if err != nil {
		return err
	}
	return nil
}

// GetAllLocationPaths retrieves all configured project locations.
func (s *SettingsService) GetAllLocationPaths() ([]settings.ProjectLocation, error) {
	return settings.GetAllLocationPaths()
}

// GetDefaultLocation retrieves the default project location.
func (s *SettingsService) GetDefaultLocation() (settings.ProjectLocation, error) {
	return settings.GetDefaultLocation()
}

// AddProjectLocation adds a new project location with name and path.
func (s *SettingsService) AddProjectLocation(name, path string) (settings.ProjectLocation, error) {
	return settings.AddProjectLocation(name, path)
}

// RemoveProjectLocation removes a project location by ID.
func (s *SettingsService) RemoveProjectLocation(locationID string) error {
	return settings.RemoveProjectLocation(locationID)
}

// UpdateProjectLocation updates a project location's name and path.
func (s *SettingsService) UpdateProjectLocation(locationID, name, path string) error {
	return settings.UpdateProjectLocation(locationID, name, path)
}

// SetDefaultLocation sets the default project location by ID.
func (s *SettingsService) SetDefaultLocation(locationID string) error {
	return settings.SetDefaultLocation(locationID)
}

// AssignProjectToLocation assigns a project to a specific location.
func (s *SettingsService) AssignProjectToLocation(projectID, locationID string) error {
	return settings.AssignProjectToLocation(projectID, locationID)
}

// GetProjectLocation retrieves the location ID for a specific project.
func (s *SettingsService) GetProjectLocation(projectID string) (string, error) {
	return settings.GetProjectLocation(projectID)
}

// GetLocationUsage returns the number of projects using a location.
func (s *SettingsService) GetLocationUsage(locationID string) (int, error) {
	return settings.GetLocationUsage(locationID)
}

// CanDeleteLocation checks if a location can be safely deleted.
func (s *SettingsService) CanDeleteLocation(locationID string) (bool, error) {
	return settings.CanDeleteLocation(locationID)
}

// CheckLocationHealth verifies the health status of a specific location.
func (s *SettingsService) CheckLocationHealth(locationID string) (settings.LocationHealth, error) {
	return settings.CheckLocationHealth(locationID)
}

// CheckAllLocationsHealth verifies the health status of all project locations.
func (s *SettingsService) CheckAllLocationsHealth() ([]settings.LocationHealth, error) {
	return settings.CheckAllLocationsHealth()
}

// CheckSystemBookmarksHealth reports staleness of the projects and shared projects directory bookmarks (macOS).
func (s *SettingsService) CheckSystemBookmarksHealth() (settings.SystemBookmarksHealth, error) {
	return settings.CheckSystemBookmarksHealth()
}

// GetSyncAfterCheckpoint returns whether auto-sync after checkpoint is enabled.
func (s *SettingsService) GetSyncAfterCheckpoint() (bool, error) {
	return settings.GetSyncAfterCheckpoint()
}

// SetSyncAfterCheckpoint sets the auto-sync after checkpoint preference.
func (s *SettingsService) SetSyncAfterCheckpoint(enabled bool) error {
	return settings.SetSyncAfterCheckpoint(enabled)
}

// GetUseUpdateSync returns whether the experimental non-destructive update
// sync is enabled for the polling loop.
func (s *SettingsService) GetUseUpdateSync() (bool, error) {
	return settings.GetUseUpdateSync()
}

// SetUseUpdateSync enables or disables the experimental non-destructive
// update sync used by the polling loop in place of the destructive pull.
func (s *SettingsService) SetUseUpdateSync(enabled bool) error {
	return settings.SetUseUpdateSync(enabled)
}

// GetMetadataOnlyStorage returns whether downloaded and uploaded chunks are discarded.
func (s *SettingsService) GetMetadataOnlyStorage() (bool, error) {
	return settings.GetMetadataOnlyStorage()
}

// SetMetadataOnlyStorage sets whether downloaded and uploaded chunks are discarded.
func (s *SettingsService) SetMetadataOnlyStorage(enabled bool) error {
	return settings.SetMetadataOnlyStorage(enabled)
}

// GetOverwriteDroppedFiles returns whether OS file drops overwrite matching files.
func (s *SettingsService) GetOverwriteDroppedFiles() (bool, error) {
	return settings.GetOverwriteDroppedFiles()
}

// SetOverwriteDroppedFiles sets whether OS file drops overwrite matching files.
func (s *SettingsService) SetOverwriteDroppedFiles(enabled bool) error {
	return settings.SetOverwriteDroppedFiles(enabled)
}

// GetMinimizeOnClose returns whether the app minimizes to tray on close.
func (s *SettingsService) GetMinimizeOnClose() (bool, error) {
	return settings.GetMinimizeOnClose()
}

// SetMinimizeOnClose sets the minimize-to-tray on close preference.
func (s *SettingsService) SetMinimizeOnClose(enabled bool) error {
	return settings.SetMinimizeOnClose(enabled)
}

// GetBridgeEnabled returns whether the bridge HTTP server is enabled.
func (s *SettingsService) GetBridgeEnabled() (bool, error) {
	return settings.GetBridgeEnabled()
}

// SetBridgeEnabled sets the bridge HTTP server enabled preference.
// Starts or stops the bridge server accordingly.
func (s *SettingsService) SetBridgeEnabled(enabled bool) error {
	err := settings.SetBridgeEnabled(enabled)
	if err != nil {
		return err
	}
	bridgeLifecycleMu.RLock()
	start := startBridge
	stop := stopBridge
	bridgeLifecycleMu.RUnlock()
	if enabled {
		start()
	} else {
		stop()
	}
	return nil
}

// GetShowTypeIcons returns whether type icons are shown in the browser.
func (s *SettingsService) GetShowTypeIcons() (bool, error) {
	return settings.GetShowTypeIcons()
}

// SetShowTypeIcons sets the show type icons preference.
func (s *SettingsService) SetShowTypeIcons(enabled bool) error {
	return settings.SetShowTypeIcons(enabled)
}

// GetIntegrationCredential retrieves integration credentials for an integration.
// Credentials are stored per user per integration (not per project).
func (s *SettingsService) GetIntegrationCredential(integrationId string) (settings.IntegrationCredential, error) {
	return settings.GetIntegrationCredential(integrationId)
}

// SaveIntegrationCredential saves or updates integration credentials.
// Credentials are stored per user per integration (not per project).
func (s *SettingsService) SaveIntegrationCredential(cred settings.IntegrationCredential) error {
	return settings.SaveIntegrationCredential(cred)
}

// DeleteIntegrationCredential deletes integration credentials for an integration.
func (s *SettingsService) DeleteIntegrationCredential(integrationId string) error {
	return settings.DeleteIntegrationCredential(integrationId)
}
