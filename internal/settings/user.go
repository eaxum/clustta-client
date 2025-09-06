package settings

import (
	"clustta/internal/auth_service"
	"clustta/internal/server/models"
	"clustta/internal/studio_service"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

type Studio struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Active string `json:"active"`
	AltUrl string `json:"alt_url"`
	Url    string `json:"url"`
	Usage  string `json:"usage"`
	Users  []models.StudioUserInfo
}

type Settings struct {
	IconScheme      string `json:"icon_scheme"`
	Theme           string `json:"theme"`
	UseAltUrl       bool   `json:"use_alt_url"`
	EulaAccepted    bool   `json:"eula_accepted"`
	ProjectGridView bool   `json:"project_grid_view"`

	ProjectsDir         string `json:"projects_dir"`
	ProjectsDirBookmark []byte `json:"projects_dir_bookmark,omitempty"`

	SharedProjectsDir         string `json:"shared_projects_dir"`
	SharedProjectsDirBookmark []byte `json:"shared_projects_dir_bookmark,omitempty"`

	WorkingDir         string `json:"working_dir"`
	WorkingDirBookmark []byte `json:"working_dir_bookmark,omitempty"`

	PinnedProjects map[string][]string      `json:"pinned_projects"`
	RecentProjects map[string][]string      `json:"recent_projects"`
	Studios        []Studio                 `json:"studios"`
	WorkSpaces     map[string][]interface{} `json:"workspaces"`
	LastStudio     string                   `json:"last_studio"`
	CurrentVersion string                   `json:"current_version"`
}

func loadUserSettings() (Settings, error) {
	settings := Settings{}
	settingsFile, err := GetUserSettingsPath()
	if err != nil {
		return settings, err
	}

	err = os.MkdirAll(filepath.Dir(settingsFile), os.ModePerm)
	if err != nil {
		return settings, err
	}

	_, err = os.Stat(settingsFile)
	if os.IsNotExist(err) {
		err = saveSettings(Settings{})
		if err != nil {
			return settings, err
		}
	}

	file, err := os.ReadFile(settingsFile)
	if err != nil {
		return settings, err
	}
	err = json.Unmarshal(file, &settings)
	if err != nil {
		return settings, err
	}
	return settings, nil
}

func saveSettings(settings Settings) error {
	settingsFile, err := GetUserSettingsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsFile, data, 0644)
}

func GetUserDirectory() (string, error) {

	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	username := extractUsername(currentUser.Username)

	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("C:/Users/%s/", username), nil

	case "darwin":

		return fmt.Sprintf("/Users/%s/", username), nil

	case "linux":
		return fmt.Sprintf("/home/%s/", username), nil

	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func extractUsername(rawUsername string) string {

	if runtime.GOOS == "windows" {
		for i := len(rawUsername) - 1; i >= 0; i-- {
			if rawUsername[i] == '\\' {
				return rawUsername[i+1:]
			}
		}
	}
	return rawUsername
}

func GetUsername() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	return extractUsername(currentUser.Username), nil
}

func GetUseAltUrl() (bool, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return false, err
	}
	return settings.UseAltUrl, nil
}

func SetUseAltUrl(useAltUrl bool) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.UseAltUrl = useAltUrl
	return saveSettings(settings)
}

func GetEulaAccepted() (bool, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return false, err
	}
	return settings.EulaAccepted, nil
}

func SetEulaAccepted(eulaAccepted bool) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.EulaAccepted = eulaAccepted
	return saveSettings(settings)
}

func GetCurrentVersion() (string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return "", err
	}
	return settings.CurrentVersion, nil
}

func SetCurrentVersion(versionNumber string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.CurrentVersion = versionNumber
	return saveSettings(settings)
}

func GetLastStudio() (string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return "", err
	}
	return settings.LastStudio, nil
}

func SetLastStudio(name string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.LastStudio = name
	return saveSettings(settings)
}

func ClearLastStudio() error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.LastStudio = ""
	return saveSettings(settings)
}

func GetIconScheme() (string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return "", err
	}
	if settings.IconScheme == "" {
		settings.IconScheme = "solid"
	}
	return settings.IconScheme, nil
}

func SetIconScheme(iconScheme string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.IconScheme = iconScheme
	return saveSettings(settings)
}

func GetTheme() (string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return "", err
	}
	if settings.Theme == "" {
		settings.Theme = "light"
	}
	return settings.Theme, nil
}

func SetTheme(theme string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.Theme = theme
	return saveSettings(settings)
}

// InitializeBookmarks should be called at app startup to resolve all stored bookmarks
func InitializeBookmarks() error {
	settings, err := loadUserSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	var errors []error

	// Initialize Projects Directory bookmark
	if len(settings.ProjectsDirBookmark) > 0 {
		if IsBookmarkStale(settings.ProjectsDirBookmark) {
			log.Printf("Projects directory bookmark is stale, will need reselection")
		} else {
			resolvedPath, err := ResolveBookmark(settings.ProjectsDirBookmark)
			if err != nil {
				log.Printf("Failed to resolve projects directory bookmark: %v", err)
				errors = append(errors, fmt.Errorf("projects directory bookmark resolution failed: %w", err))
			} else {
				log.Printf("Successfully resolved projects directory bookmark: %s", resolvedPath)
			}
		}
	}

	// Initialize Shared Projects Directory bookmark
	if len(settings.SharedProjectsDirBookmark) > 0 {
		if IsBookmarkStale(settings.SharedProjectsDirBookmark) {
			log.Printf("Shared projects directory bookmark is stale, will need reselection")
		} else {
			resolvedPath, err := ResolveBookmark(settings.SharedProjectsDirBookmark)
			if err != nil {
				log.Printf("Failed to resolve shared projects directory bookmark: %v", err)
				errors = append(errors, fmt.Errorf("shared projects directory bookmark resolution failed: %w", err))
			} else {
				log.Printf("Successfully resolved shared projects directory bookmark: %s", resolvedPath)
			}
		}
	}

	// Initialize Working Directory bookmark
	if len(settings.WorkingDirBookmark) > 0 {
		if IsBookmarkStale(settings.WorkingDirBookmark) {
			log.Printf("Working directory bookmark is stale, will need reselection")
		} else {
			resolvedPath, err := ResolveBookmark(settings.WorkingDirBookmark)
			if err != nil {
				log.Printf("Failed to resolve working directory bookmark: %v", err)
				errors = append(errors, fmt.Errorf("working directory bookmark resolution failed: %w", err))
			} else {
				log.Printf("Successfully resolved working directory bookmark: %s", resolvedPath)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("bookmark initialization had %d errors: %v", len(errors), errors)
	}

	return nil
}

func GetProjectDirectory() (string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "darwin" {
		if len(settings.ProjectsDirBookmark) > 0 && !IsBookmarkStale(settings.ProjectsDirBookmark) {
			resolvedPath, err := ResolveBookmark(settings.ProjectsDirBookmark)
			if err == nil {
				// Verify the resolved path still exists
				if _, err := os.Stat(resolvedPath); err == nil {
					return resolvedPath, nil
				}
			}
			log.Printf("Failed to resolve projects directory bookmark, falling back to stored path: %v", err)
		}
	}

	return settings.ProjectsDir, nil
}

func SetProjectDirectory(dir string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}

	if runtime.GOOS == "darwin" {
		bookmarkData, err := CreateBookmarkFromPath(dir)
		if err != nil {
			log.Printf("Failed to create bookmark for projects directory %s: %v", dir, err)
			// Continue without bookmark - store path only
		}
		settings.ProjectsDirBookmark = bookmarkData
	}

	settings.ProjectsDir = dir
	return saveSettings(settings)
}

func GetSharedProjectDirectory() (string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "darwin" {
		if len(settings.SharedProjectsDirBookmark) > 0 && !IsBookmarkStale(settings.SharedProjectsDirBookmark) {
			resolvedPath, err := ResolveBookmark(settings.SharedProjectsDirBookmark)
			if err == nil {
				// Verify the resolved path still exists
				if _, err := os.Stat(resolvedPath); err == nil {
					return resolvedPath, nil
				}
			}
			log.Printf("Failed to resolve shared projects directory bookmark, falling back to stored path: %v", err)
		}
	}

	return settings.SharedProjectsDir, nil
}

func SetSharedProjectDirectory(dir string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}

	if runtime.GOOS == "darwin" {
		bookmarkData, err := CreateBookmarkFromPath(dir)
		if err != nil {
			log.Printf("Failed to create bookmark for shared projects directory %s: %v", dir, err)
		}
		settings.SharedProjectsDirBookmark = bookmarkData
	}
	settings.SharedProjectsDir = dir

	return saveSettings(settings)
}

func GetWorkingDirectory() (string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "darwin" {
		if len(settings.WorkingDirBookmark) > 0 && !IsBookmarkStale(settings.WorkingDirBookmark) {
			resolvedPath, err := ResolveBookmark(settings.WorkingDirBookmark)
			if err == nil {
				// Verify the resolved path still exists
				if _, err := os.Stat(resolvedPath); err == nil {
					return resolvedPath, nil
				}
			}
			log.Printf("Failed to resolve working directory bookmark, falling back to stored path: %v", err)
		}
	}

	return settings.WorkingDir, nil
}

func SetWorkingDirectory(dir string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}

	if runtime.GOOS == "darwin" {
		bookmarkData, err := CreateBookmarkFromPath(dir)
		if err != nil {
			log.Printf("Failed to create bookmark for working directory %s: %v", dir, err)
		}
		settings.WorkingDirBookmark = bookmarkData
	}

	settings.WorkingDir = dir
	return saveSettings(settings)
}

func ToggleProjectGridView() error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.ProjectGridView = !settings.ProjectGridView
	return saveSettings(settings)
}

func IsProjectGridView() (bool, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return true, err
	}
	return settings.ProjectGridView, nil
}

func GetPinnedProjects(studioName string) ([]string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return []string{}, err
	}
	projects, exists := settings.PinnedProjects[studioName]
	if !exists {
		return []string{}, nil
	}
	return projects, nil
}

func PinProject(studioName string, projectId string) ([]string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return []string{}, err
	}

	if settings.PinnedProjects == nil {
		settings.PinnedProjects = make(map[string][]string)
	}

	if _, exists := settings.PinnedProjects[studioName]; !exists {
		settings.PinnedProjects[studioName] = []string{}
	}
	settings.PinnedProjects[studioName] = append(settings.PinnedProjects[studioName], projectId)
	err = saveSettings(settings)
	if err != nil {
		return []string{}, err
	}
	return settings.PinnedProjects[studioName], err
}

func GetRecentProjects(studioName string) ([]string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return []string{}, err
	}
	projects, exists := settings.RecentProjects[studioName]
	if !exists {
		return []string{}, nil
	}
	return projects, nil
}

func AddRecentProject(studioName, projectId string) ([]string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return []string{}, err
	}

	// Initialize the map if it doesn't exist
	if settings.RecentProjects == nil {
		settings.RecentProjects = make(map[string][]string)
	}

	// Initialize the studio's project list if it doesn't exist
	if _, exists := settings.RecentProjects[studioName]; !exists {
		settings.RecentProjects[studioName] = []string{}
	}

	recentProjects := settings.RecentProjects[studioName]

	// Check if the project already exists in the list
	foundIndex := -1
	for i, id := range recentProjects {
		if id == projectId {
			foundIndex = i
			break
		}
	}

	// If project exists, remove it from its current position
	if foundIndex != -1 {
		recentProjects = append(recentProjects[:foundIndex], recentProjects[foundIndex+1:]...)
	}

	// Add the project to the top of the list
	recentProjects = append([]string{projectId}, recentProjects...)

	// Optional: Limit the size of the recent projects list (e.g., to 10 items)
	const maxRecentProjects = 10
	if len(recentProjects) > maxRecentProjects {
		recentProjects = recentProjects[:maxRecentProjects]
	}

	// Update the settings
	settings.RecentProjects[studioName] = recentProjects

	// Save settings
	err = saveSettings(settings)
	if err != nil {
		return []string{}, err
	}

	return settings.RecentProjects[studioName], nil
}

func ClearRecentProject() error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.RecentProjects = make(map[string][]string)
	return saveSettings(settings)
}

func UnpinProject(studioName string, projectId string) ([]string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return []string{}, err
	}
	projects := settings.PinnedProjects[studioName]
	for i, project := range projects {
		if project == projectId {
			projects = append(projects[:i], projects[i+1:]...)
			break
		}
	}
	settings.PinnedProjects[studioName] = projects
	err = saveSettings(settings)
	if err != nil {
		return []string{}, err
	}
	return settings.PinnedProjects[studioName], err
}

func AddStudio(studio Studio) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.Studios = append(settings.Studios, studio)
	return saveSettings(settings)
}

func GetStudios() ([]Studio, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return []Studio{}, err
	}

	projectsPath, err := GetProjectDirectory()
	if err != nil {
		return settings.Studios, err
	}
	personal := Studio{
		Name:   "Personal",
		Url:    projectsPath,
		AltUrl: "",
	}

	if len(settings.Studios) == 0 {
		settings.Studios = append(settings.Studios, personal)
	}

	userStudios, err := studio_service.GetUserStudios()
	if err != nil {
		return settings.Studios, nil
	}

	settings.Studios = []Studio{personal}

	for _, userStudio := range userStudios {
		studioUsers, err := studio_service.GetStudioUsers(userStudio.Id)
		if err != nil {
			return settings.Studios, nil
		}
		studio := Studio{
			Id:     userStudio.Id,
			Name:   userStudio.Name,
			Url:    userStudio.URL,
			AltUrl: userStudio.AltURL,
			Users:  studioUsers,
		}
		settings.Studios = append(settings.Studios, studio)
	}

	err = saveSettings(settings)
	if err != nil {
		return settings.Studios, err
	}
	return settings.Studios, nil
}

func GetProjectWorkspaces(projectId string) ([]interface{}, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return []interface{}{}, err
	}

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return []interface{}{}, err
	}

	defaultWorkspace := map[string]interface{}{
		"name":                 "Default",
		"filters":              map[string]interface{}{"taskFilters": []interface{}{}, "entityFilters": []interface{}{}, "resourceFilters": []interface{}{}},
		"workspaceSearchQuery": "",
	}

	taskFilter := map[string]interface{}{
		"email":      user.Email,
		"first_name": user.FirstName,
		"id":         user.Id,
		"last_name":  user.LastName,
		"type":       "assignation",
		"username":   user.Username,
	}
	assignedTasksWorkspace := map[string]interface{}{
		"name":                 "My Tasks",
		"filters":              map[string]interface{}{"taskFilters": []interface{}{taskFilter}, "entityFilters": []interface{}{}, "resourceFilters": []interface{}{}, "showTasks": true, "onlyAssets": true},
		"workspaceSearchQuery": "",
	}

	projectWorkspaces, exists := settings.WorkSpaces[projectId]
	if !exists {
		projectWorkspaces = append(projectWorkspaces, defaultWorkspace)
		projectWorkspaces = append(projectWorkspaces, assignedTasksWorkspace)
		return projectWorkspaces, nil
	}
	projectWorkspaces = append([]interface{}{defaultWorkspace, assignedTasksWorkspace}, projectWorkspaces...)
	return projectWorkspaces, nil
}

func AddProjectWorkspace(projectId string, workspaceData interface{}) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}

	// check if settings.WorkSpaces is nil
	if settings.WorkSpaces == nil {
		settings.WorkSpaces = make(map[string][]interface{})
	}

	if _, exists := settings.WorkSpaces[projectId]; !exists {
		settings.WorkSpaces[projectId] = []interface{}{}
	}

	projectWorkspaces := settings.WorkSpaces[projectId]
	projectWorkspaces = append(projectWorkspaces, workspaceData)
	settings.WorkSpaces[projectId] = projectWorkspaces
	return saveSettings(settings)
}

func RemoveProjectWorkspace(projectId string, workspaceName string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	projectWorkspaces := settings.WorkSpaces[projectId]
	for i, workspace := range projectWorkspaces {
		if workspaceName == workspace.(map[string]interface{})["name"] {
			projectWorkspaces = append(projectWorkspaces[:i], projectWorkspaces[i+1:]...)
			break
		}
	}
	settings.WorkSpaces[projectId] = projectWorkspaces
	return saveSettings(settings)
}
