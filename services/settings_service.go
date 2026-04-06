package services

import (
	"clustta/internal/bridge"
	"clustta/internal/settings"
	"os"
	"strings"
)

type SettingsService struct{}

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

// GetSyncAfterCheckpoint returns whether auto-sync after checkpoint is enabled.
func (s *SettingsService) GetSyncAfterCheckpoint() (bool, error) {
	return settings.GetSyncAfterCheckpoint()
}

// SetSyncAfterCheckpoint sets the auto-sync after checkpoint preference.
func (s *SettingsService) SetSyncAfterCheckpoint(enabled bool) error {
	return settings.SetSyncAfterCheckpoint(enabled)
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
	if enabled {
		bridge.Start()
	} else {
		bridge.Stop()
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
