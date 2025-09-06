package services

import (
	"clustta/internal/settings"
)

type SettingsService struct{}

func (s *SettingsService) GetStudios(path string) ([]settings.Studio, error) {
	studios, err := settings.GetStudios()
	if err != nil {
		return studios, err
	}
	return studios, nil
}

func (s *SettingsService) PinProject(studioName, projectId string) ([]string, error) {
	projectsId, err := settings.PinProject(studioName, projectId)
	if err != nil {
		return projectsId, err
	}
	return projectsId, nil
}

func (s *SettingsService) GetUserDirectory() (string, error) {
	return settings.GetUserDirectory()
}

func (s *SettingsService) GetUsername() (string, error) {
	return settings.GetUsername()
}

func (s *SettingsService) UnpinProject(studioName, projectId string) ([]string, error) {
	projectsId, err := settings.UnpinProject(studioName, projectId)
	if err != nil {
		return projectsId, err
	}
	return projectsId, nil
}

func (s *SettingsService) GetPinnedProjects(studioName string) ([]string, error) {
	projectsId, err := settings.GetPinnedProjects(studioName)
	if err != nil {
		return projectsId, err
	}
	return projectsId, nil
}

func (s *SettingsService) GetRecentProjects(studioName string) ([]string, error) {
	projectsId, err := settings.GetRecentProjects(studioName)
	if err != nil {
		return projectsId, err
	}
	return projectsId, nil
}

func (s *SettingsService) AddRecentProject(studioName, projectId string) ([]string, error) {
	projectsId, err := settings.AddRecentProject(studioName, projectId)
	if err != nil {
		return projectsId, err
	}
	return projectsId, nil
}
func (s *SettingsService) ClearRecentProject(studioName string) error {
	return settings.ClearRecentProject()
}

func (s *SettingsService) AddProjectWorkspace(projectId string, workspaceData interface{}) error {
	err := settings.AddProjectWorkspace(projectId, workspaceData)
	if err != nil {
		return err
	}
	return nil
}
func (s *SettingsService) RemoveProjectWorkspace(projectId string, workspaceName string) error {
	err := settings.RemoveProjectWorkspace(projectId, workspaceName)
	if err != nil {
		return err
	}
	return nil
}

func (s *SettingsService) GetProjectWorkspaces(projectId string) ([]interface{}, error) {
	projectWorkspaces, err := settings.GetProjectWorkspaces(projectId)
	if err != nil {
		return projectWorkspaces, err
	}
	return projectWorkspaces, nil
}

func (s *SettingsService) GetUseAltUrl() (bool, error) {
	useAltUrl, err := settings.GetUseAltUrl()
	if err != nil {
		return useAltUrl, err
	}
	return useAltUrl, nil
}

func (s *SettingsService) SetUseAltUrl(useAltUrl bool) error {
	err := settings.SetUseAltUrl(useAltUrl)
	if err != nil {
		return err
	}
	return nil
}

func (s *SettingsService) GetEulaAccepted() (bool, error) {
	eulaAccepted, err := settings.GetEulaAccepted()
	if err != nil {
		return eulaAccepted, err
	}
	return eulaAccepted, nil
}

func (s *SettingsService) SetEulaAccepted(eulaAccepted bool) error {
	err := settings.SetEulaAccepted(eulaAccepted)
	if err != nil {
		return err
	}
	return nil
}

func (s *SettingsService) GetLastStudio() (string, error) {
	lastStudioName, err := settings.GetLastStudio()
	if err != nil {
		return lastStudioName, err
	}
	return lastStudioName, nil
}

func (s *SettingsService) SetLastStudio(lastStudioName string) error {
	err := settings.SetLastStudio(lastStudioName)
	if err != nil {
		return err
	}
	return nil
}

func (s *SettingsService) GetCurrentVersion() (string, error) {
	versionNumber, err := settings.GetCurrentVersion()
	if err != nil {
		return versionNumber, err
	}
	return versionNumber, nil
}

func (s *SettingsService) SetCurrentVersion(versionNumber string) error {
	err := settings.SetCurrentVersion(versionNumber)
	if err != nil {
		return err
	}
	return nil
}

func (s *SettingsService) IsProjectGridView() (bool, error) {
	return settings.IsProjectGridView()
}

func (s *SettingsService) ToggleProjectGridView() error {
	return settings.ToggleProjectGridView()
}

func (s *SettingsService) GetIconScheme() (string, error) {
	iconScheme, err := settings.GetIconScheme()
	if err != nil {
		return iconScheme, err
	}
	return iconScheme, nil
}

func (s *SettingsService) SetIconScheme(iconScheme string) error {
	err := settings.SetIconScheme(iconScheme)
	if err != nil {
		return err
	}
	return nil
}

func (s *SettingsService) GetTheme() (string, error) {
	theme, err := settings.GetTheme()
	if err != nil {
		return theme, err
	}
	return theme, nil
}

func (s *SettingsService) SetTheme(theme string) error {
	err := settings.SetTheme(theme)
	if err != nil {
		return err
	}
	return nil
}

func (s *SettingsService) GetProjectDirectory() (string, error) {
	projectDir, err := settings.GetProjectDirectory()
	if err != nil {
		return projectDir, err
	}
	return projectDir, nil
}

func (s *SettingsService) SetProjectDirectory(dir string) error {
	err := settings.SetProjectDirectory(dir)
	if err != nil {
		return err
	}
	return nil
}

func (s *SettingsService) GetSharedProjectDirectory() (string, error) {
	projectDir, err := settings.GetSharedProjectDirectory()
	if err != nil {
		return projectDir, err
	}
	return projectDir, nil
}

func (s *SettingsService) SetSharedProjectDirectory(dir string) error {
	err := settings.SetSharedProjectDirectory(dir)
	if err != nil {
		return err
	}
	return nil
}

func (s *SettingsService) GetWorkingDirectory() (string, error) {
	projectDir, err := settings.GetWorkingDirectory()
	if err != nil {
		return projectDir, err
	}
	return projectDir, nil
}

func (s *SettingsService) SetWorkingDirectory(dir string) error {
	err := settings.SetWorkingDirectory(dir)
	if err != nil {
		return err
	}
	return nil
}
