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
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/zalando/go-keyring"
)

var (
	activeProjectWorkingDir   string
	activeProjectWorkingDirMu sync.RWMutex
)

// SetActiveProjectWorkingDir registers the current project's working directory as an allowed path.
func SetActiveProjectWorkingDir(path string) {
	cleaned, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return
	}
	activeProjectWorkingDirMu.Lock()
	defer activeProjectWorkingDirMu.Unlock()
	activeProjectWorkingDir = strings.ToLower(cleaned)
}

// flexBool tolerates legacy persisted values where Studio.active was a string ("", "0", "false", "1", "true").
// It always marshals back as a real JSON boolean.
type flexBool bool

func (b *flexBool) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		switch strings.ToLower(s) {
		case "1", "true":
			*b = true
		default:
			*b = false
		}
		return nil
	}
	var v bool
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*b = flexBool(v)
	return nil
}

func (b flexBool) MarshalJSON() ([]byte, error) {
	return json.Marshal(bool(b))
}

type Studio struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Active      flexBool `json:"active"`
	AltUrl      string   `json:"alt_url"`
	Url         string   `json:"url"`
	Usage       string   `json:"usage"`
	HostingMode string   `json:"hosting_mode"`
	Users       []models.StudioUserInfo
}

type ProjectLocation struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	Bookmark   []byte   `json:"bookmark,omitempty"`
	IsDefault  bool     `json:"is_default"`
	ProjectIDs []string `json:"project_ids"`
}

// IntegrationCredential stores user credentials for external integrations.
// Stored per user locally, keyed by "projectId_integrationId".
type IntegrationCredential struct {
	IntegrationId string `json:"integration_id"`
	UserId        string `json:"user_id"`
	UserName      string `json:"user_name"`
	UserEmail     string `json:"user_email"`
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	ExpiresAt     int64  `json:"expires_at"`
	ApiUrl        string `json:"api_url"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type Settings struct {
	IconScheme            string `json:"icon_scheme"`
	Theme                 string `json:"theme"`
	ThemeTint             string `json:"theme_tint"`
	Language              string `json:"language"` // User's language preference (e.g., "en", "es", "fr")
	EulaAccepted          bool   `json:"eula_accepted"`
	ProjectGridView       bool   `json:"project_grid_view"`
	UseGrid               bool   `json:"use_grid"`
	DefaultViewMode       string `json:"default_view_mode"`
	ShowUntrackedProjects bool   `json:"show_untracked_projects"`
	ShowTypeIcons         *bool  `json:"show_type_icons,omitempty"`

	ProjectsDir         string `json:"projects_dir"`
	ProjectsDirBookmark []byte `json:"projects_dir_bookmark,omitempty"`

	SharedProjectsDir         string `json:"shared_projects_dir"`
	SharedProjectsDirBookmark []byte `json:"shared_projects_dir_bookmark,omitempty"`

	WorkingDir         string `json:"working_dir,omitempty"`          // Deprecated: Use ProjectLocations instead
	WorkingDirBookmark []byte `json:"working_dir_bookmark,omitempty"` // Deprecated: Use ProjectLocations instead

	ProjectLocations  []ProjectLocation `json:"project_locations"`
	DefaultLocationID string            `json:"default_location_id"`

	SyncAfterCheckpoint bool  `json:"sync_after_checkpoint"`
	UseUpdateSync       bool  `json:"use_update_sync"`
	BridgeEnabled       bool  `json:"bridge_enabled"`
	MinimizeOnClose     *bool `json:"minimize_on_close,omitempty"`

	AgentAutoApproveDestructive bool `json:"agent_auto_approve_destructive"`

	AgentProviderModels map[string]string `json:"agent_provider_models,omitempty"`

	PinnedProjects    map[string][]string              `json:"pinned_projects"`
	RecentProjects    map[string][]string              `json:"recent_projects"`
	Studios           []Studio                         `json:"studios"`
	WorkSpaces        map[string][]interface{}         `json:"workspaces"`
	DependencyPresets map[string][]interface{}         `json:"dependency_presets"`
	IgnoreListPresets map[string][]string              `json:"ignore_list_presets"`
	IntegrationCreds  map[string]IntegrationCredential `json:"integration_credentials"`
	LastStudio        string                           `json:"last_studio"`
	CurrentVersion    string                           `json:"current_version"`
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

	if migrateLocationIDsToUUID(&settings) {
		if saveErr := saveSettings(settings); saveErr != nil {
			log.Printf("Failed to persist project location ID migration: %v", saveErr)
		}
	}

	return settings, nil
}

// migrateLocationIDsToUUID rewrites any legacy non-UUID project location IDs
// to UUIDs and updates DefaultLocationID to match. Returns true if any change
// was made.
func migrateLocationIDsToUUID(settings *Settings) bool {
	changed := false
	idRemap := make(map[string]string)
	for i := range settings.ProjectLocations {
		loc := &settings.ProjectLocations[i]
		if _, err := uuid.Parse(loc.ID); err == nil {
			continue
		}
		oldID := loc.ID
		newID := uuid.NewString()
		loc.ID = newID
		idRemap[oldID] = newID
		changed = true
	}
	if newDefault, ok := idRemap[settings.DefaultLocationID]; ok {
		settings.DefaultLocationID = newDefault
		changed = true
	}
	return changed
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

	homeDir := currentUser.HomeDir
	if homeDir == "" {
		return "", fmt.Errorf("user home directory is empty")
	}

	// Normalize path separators and ensure trailing slash
	homeDir = strings.ReplaceAll(homeDir, "\\", "/")
	if !strings.HasSuffix(homeDir, "/") {
		homeDir += "/"
	}

	return homeDir, nil
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

// GetThemeTint returns the current theme tint name, defaulting to "neutral".
func GetThemeTint() (string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return "", err
	}
	if settings.ThemeTint == "" {
		settings.ThemeTint = "neutral"
	}
	return settings.ThemeTint, nil
}

// SetThemeTint persists the chosen theme tint.
func SetThemeTint(tint string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.ThemeTint = tint
	return saveSettings(settings)
}

// GetLanguage returns the user's language preference or defaults to "en".
func GetLanguage() (string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return "", err
	}
	if settings.Language == "" {
		settings.Language = "en"
	}
	return settings.Language, nil
}

// SetLanguage updates the user's language preference.
func SetLanguage(language string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.Language = language
	return saveSettings(settings)
}

func GetUseGrid() (bool, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return false, err
	}
	// Default to true (grid view)
	if !settings.UseGrid {
		// Check if this is first time (no setting saved yet)
		// If UseGrid field doesn't exist in JSON, it defaults to false
		// For new users, we want default to be true
		settingsFile, _ := GetUserSettingsPath()
		file, _ := os.ReadFile(settingsFile)
		if !contains(string(file), "use_grid") {
			settings.UseGrid = true
			saveSettings(settings)
		}
	}
	return settings.UseGrid, nil
}

func SetUseGrid(useGrid bool) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.UseGrid = useGrid
	return saveSettings(settings)
}

// GetDefaultViewMode retrieves the default view mode setting.
// Returns "compact" (list), "dense" (compact), or "grid".
func GetDefaultViewMode() (string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return "compact", err
	}
	if settings.DefaultViewMode == "" {
		if settings.UseGrid {
			return "grid", nil
		}
		return "compact", nil
	}
	return settings.DefaultViewMode, nil
}

// SetDefaultViewMode sets the default view mode.
func SetDefaultViewMode(viewMode string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.DefaultViewMode = viewMode
	settings.UseGrid = viewMode == "grid"
	return saveSettings(settings)
}

// GetSyncAfterCheckpoint returns whether auto-sync after checkpoint is enabled.
// Defaults to false if not set.
func GetSyncAfterCheckpoint() (bool, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return false, err
	}
	return settings.SyncAfterCheckpoint, nil
}

// SetSyncAfterCheckpoint sets the auto-sync after checkpoint preference.
func SetSyncAfterCheckpoint(enabled bool) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.SyncAfterCheckpoint = enabled
	return saveSettings(settings)
}

// GetUseUpdateSync returns whether the experimental non-destructive update
// sync is enabled. Defaults to false.
func GetUseUpdateSync() (bool, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return false, err
	}
	return settings.UseUpdateSync, nil
}

// SetUseUpdateSync sets whether the experimental non-destructive update sync
// is enabled in place of the destructive pull during polling.
func SetUseUpdateSync(enabled bool) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.UseUpdateSync = enabled
	return saveSettings(settings)
}

// GetMinimizeOnClose returns whether the app should minimize to tray on close.
// Defaults to true if not explicitly set.
func GetMinimizeOnClose() (bool, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return true, err
	}
	if settings.MinimizeOnClose == nil {
		return true, nil
	}
	return *settings.MinimizeOnClose, nil
}

// SetMinimizeOnClose sets the minimize-to-tray on close preference.
func SetMinimizeOnClose(enabled bool) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.MinimizeOnClose = &enabled
	return saveSettings(settings)
}

// GetBridgeEnabled returns whether the bridge HTTP server is enabled.
// Defaults to false if not set.
func GetBridgeEnabled() (bool, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return false, err
	}
	return settings.BridgeEnabled, nil
}

// SetBridgeEnabled sets the bridge HTTP server enabled preference.
func SetBridgeEnabled(enabled bool) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.BridgeEnabled = enabled
	return saveSettings(settings)
}

// GetAgentAutoApproveDestructive returns whether the agent should skip approval
// prompts for destructive tool calls. Defaults to false.
func GetAgentAutoApproveDestructive() (bool, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return false, err
	}
	return settings.AgentAutoApproveDestructive, nil
}

// SetAgentAutoApproveDestructive sets the auto-approve preference for the agent.
func SetAgentAutoApproveDestructive(enabled bool) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.AgentAutoApproveDestructive = enabled
	return saveSettings(settings)
}

// GetAgentModel returns the user-selected model for the given provider.
// Returns empty string if no override is set (caller should fall back to provider default).
func GetAgentModel(provider string) (string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return "", err
	}
	if settings.AgentProviderModels == nil {
		return "", nil
	}
	return settings.AgentProviderModels[provider], nil
}

// SetAgentModel persists the selected model for the given provider.
// Pass an empty model string to clear the override.
func SetAgentModel(provider, model string) error {
	if provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	if settings.AgentProviderModels == nil {
		settings.AgentProviderModels = map[string]string{}
	}
	if model == "" {
		delete(settings.AgentProviderModels, provider)
	} else {
		settings.AgentProviderModels[provider] = model
	}
	return saveSettings(settings)
}

// GetShowTypeIcons returns whether type icons are shown in the browser.
// Defaults to true if not set.
func GetShowTypeIcons() (bool, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return true, err
	}
	if settings.ShowTypeIcons == nil {
		return true, nil
	}
	return *settings.ShowTypeIcons, nil
}

// SetShowTypeIcons sets the show type icons preference.
func SetShowTypeIcons(enabled bool) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.ShowTypeIcons = &enabled
	return saveSettings(settings)
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (len(s) >= len(substr)) && (s == substr || len(s) > len(substr) && (s[0:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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

// GetWorkingDirectory returns the default location path (for backward compatibility)
// Deprecated: Use GetDefaultLocation() instead
func GetWorkingDirectory() (string, error) {
	defaultLoc, err := GetDefaultLocation()
	if err != nil {
		// Fallback to old behavior if no locations configured
		settings, err := loadUserSettings()
		if err != nil {
			return "", err
		}

		if runtime.GOOS == "darwin" {
			if len(settings.WorkingDirBookmark) > 0 && !IsBookmarkStale(settings.WorkingDirBookmark) {
				resolvedPath, err := ResolveBookmark(settings.WorkingDirBookmark)
				if err == nil {
					if _, err := os.Stat(resolvedPath); err == nil {
						return resolvedPath, nil
					}
				}
			}
		}

		return settings.WorkingDir, nil
	}
	return defaultLoc.Path, nil
}

// SetWorkingDirectory is deprecated but kept for compatibility during migration
// Deprecated: Use AddProjectLocation() and SetDefaultLocation() instead
func SetWorkingDirectory(dir string) error {
	// This now adds/updates the default location
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}

	if len(settings.ProjectLocations) == 0 {
		// No locations yet, create first one
		_, err := AddProjectLocation("Default", dir)
		return err
	}

	// Update default location
	defaultLoc, err := GetDefaultLocation()
	if err != nil {
		return err
	}

	return UpdateProjectLocation(defaultLoc.ID, defaultLoc.Name, dir)
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

func ToggleShowUntrackedProjects() error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	settings.ShowUntrackedProjects = !settings.ShowUntrackedProjects
	return saveSettings(settings)
}

func IsShowUntrackedProjects() (bool, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return true, err
	}
	return settings.ShowUntrackedProjects, nil
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

	// Handle based on auth mode
	authMode := auth_service.GetActiveAuthMode()

	switch authMode {
	case auth_service.AuthModeOffline:
		// Offline mode: only Personal studio, no network calls
		settings.Studios = []Studio{personal}
		err = saveSettings(settings)
		if err != nil {
			return settings.Studios, err
		}
		return settings.Studios, nil

	case auth_service.AuthModeStudio:
		// Studio mode: only the private server studio (no Personal)
		accountToken, err := auth_service.GetActiveAccountToken()
		if err != nil {
			return []Studio{}, fmt.Errorf("failed to get account token: %w", err)
		}

		// Fetch studio info from the private server
		studioInfo, err := studio_service.GetStudioInfo(accountToken.AuthHost)
		if err != nil {
			// If we can't get studio info, use fallback with auth host
			privateStudio := Studio{
				Id:          accountToken.StudioId,
				Name:        "Private Studio",
				Url:         accountToken.AuthHost,
				AltUrl:      "",
				Active:      true,
				HostingMode: "private",
			}
			settings.Studios = []Studio{privateStudio}
		} else {
			privateStudio := Studio{
				Id:          studioInfo.Id,
				Name:        studioInfo.Name,
				Url:         studioInfo.Url,
				AltUrl:      studioInfo.AltUrl,
				Active:      true,
				HostingMode: "private",
			}
			// If name is empty, use a default
			if privateStudio.Name == "" {
				privateStudio.Name = "Private Studio"
			}
			settings.Studios = []Studio{privateStudio}
		}

		err = saveSettings(settings)
		if err != nil {
			return settings.Studios, err
		}
		return settings.Studios, nil

	default: // AuthModeGlobal
		// Global mode: fetch studios from api.clustta.com (existing behavior)
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
				Id:          userStudio.Id,
				Name:        userStudio.Name,
				Url:         userStudio.URL,
				AltUrl:      userStudio.AltURL,
				Active:      flexBool(userStudio.Active),
				HostingMode: userStudio.HostingMode,
				Users:       studioUsers,
			}
			settings.Studios = append(settings.Studios, studio)
		}

		err = saveSettings(settings)
		if err != nil {
			return settings.Studios, err
		}
		return settings.Studios, nil
	}
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
		"filters":              map[string]interface{}{"assetFilters": []interface{}{}, "collectionFilters": []interface{}{}, "resourceFilters": []interface{}{}},
		"workspaceSearchQuery": "",
		"collection":           nil,
	}

	assetFilter := map[string]interface{}{
		"email":      user.Email,
		"first_name": user.FirstName,
		"id":         user.Id,
		"last_name":  user.LastName,
		"type":       "assignation",
		"username":   user.Username,
	}
	assignedAssetsWorkspace := map[string]interface{}{
		"name":                 "My Assets",
		"filters":              map[string]interface{}{"assetFilters": []interface{}{assetFilter}, "collectionFilters": []interface{}{}, "resourceFilters": []interface{}{}, "showAssets": true, "onlyAssets": true},
		"workspaceSearchQuery": "",
		"collection":           nil,
	}

	projectWorkspaces, exists := settings.WorkSpaces[projectId]
	if !exists {
		projectWorkspaces = append(projectWorkspaces, defaultWorkspace)
		projectWorkspaces = append(projectWorkspaces, assignedAssetsWorkspace)
		return projectWorkspaces, nil
	}
	projectWorkspaces = append([]interface{}{defaultWorkspace, assignedAssetsWorkspace}, projectWorkspaces...)
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

// UpdateProjectWorkspace replaces an existing workspace configuration by name.
func UpdateProjectWorkspace(projectId string, workspaceName string, workspaceData interface{}) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}

	if settings.WorkSpaces == nil {
		return fmt.Errorf("no workspaces found for project %s", projectId)
	}

	projectWorkspaces, exists := settings.WorkSpaces[projectId]
	if !exists {
		return fmt.Errorf("no workspaces found for project %s", projectId)
	}

	for i, workspace := range projectWorkspaces {
		if workspaceName == workspace.(map[string]interface{})["name"] {
			projectWorkspaces[i] = workspaceData
			settings.WorkSpaces[projectId] = projectWorkspaces
			return saveSettings(settings)
		}
	}

	return fmt.Errorf("workspace %s not found in project %s", workspaceName, projectId)
}

// ========== Integration Credentials Management ==========

// integrationKeyringService is the OS-keyring service name under which
// integration access/refresh tokens are stored. Storing them in the keyring
// (Windows Credential Manager / macOS Keychain / libsecret) keeps API keys
// off disk in plaintext.
const integrationKeyringService = "clustta-integration"

type integrationKeyringPayload struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// loadIntegrationSecrets reads access/refresh tokens for an integration from
// the OS keyring. Returns empty values (no error) if nothing is stored.
func loadIntegrationSecrets(integrationId string) (string, string, error) {
	raw, err := keyring.Get(integrationKeyringService, integrationId)
	if err != nil {
		// keyring.ErrNotFound is the typical case for "never stored".
		return "", "", nil
	}
	var payload integrationKeyringPayload
	if jsonErr := json.Unmarshal([]byte(raw), &payload); jsonErr != nil {
		return "", "", jsonErr
	}
	return payload.AccessToken, payload.RefreshToken, nil
}

// storeIntegrationSecrets writes access/refresh tokens to the OS keyring.
// If both are empty, the entry is removed.
func storeIntegrationSecrets(integrationId, accessToken, refreshToken string) error {
	if accessToken == "" && refreshToken == "" {
		_ = keyring.Delete(integrationKeyringService, integrationId)
		return nil
	}
	payload, err := json.Marshal(integrationKeyringPayload{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
	if err != nil {
		return err
	}
	return keyring.Set(integrationKeyringService, integrationId, string(payload))
}

// GetIntegrationCredential retrieves integration credentials for an integration.
// Credentials are stored per user per integration (not per project).
// Secrets (access/refresh tokens) live in the OS keyring; the rest of the
// metadata lives in settings.json. This function rejoins the two halves.
func GetIntegrationCredential(integrationId string) (IntegrationCredential, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return IntegrationCredential{}, err
	}

	cred, exists := IntegrationCredential{}, false
	if settings.IntegrationCreds != nil {
		cred, exists = settings.IntegrationCreds[integrationId]
	}
	if !exists {
		// Some integrations (e.g. agent ollama provider) only store the provider
		// name in ApiUrl with no token. Surface "not found" only if we genuinely
		// have nothing in either location.
		access, refresh, _ := loadIntegrationSecrets(integrationId)
		if access == "" && refresh == "" {
			return IntegrationCredential{}, fmt.Errorf("no credentials found for %s", integrationId)
		}
		cred.IntegrationId = integrationId
		cred.AccessToken = access
		cred.RefreshToken = refresh
		return cred, nil
	}

	// Hydrate secrets from keyring. Fall back to whatever is in settings (legacy
	// installs may still have the token stored in JSON until next SaveIntegrationCredential).
	access, refresh, _ := loadIntegrationSecrets(integrationId)
	if access != "" {
		cred.AccessToken = access
	}
	if refresh != "" {
		cred.RefreshToken = refresh
	}
	return cred, nil
}

// SaveIntegrationCredential saves or updates integration credentials.
// Secrets go to the OS keyring; the rest of the metadata is persisted in
// settings.json with the secret fields blanked out.
func SaveIntegrationCredential(cred IntegrationCredential) error {
	if err := storeIntegrationSecrets(cred.IntegrationId, cred.AccessToken, cred.RefreshToken); err != nil {
		return fmt.Errorf("failed to store integration secrets in keyring: %w", err)
	}

	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	if settings.IntegrationCreds == nil {
		settings.IntegrationCreds = make(map[string]IntegrationCredential)
	}

	stored := cred
	stored.AccessToken = ""
	stored.RefreshToken = ""
	settings.IntegrationCreds[cred.IntegrationId] = stored
	return saveSettings(settings)
}

// DeleteIntegrationCredential deletes integration credentials for an integration.
// Removes both the keyring entry and the settings.json metadata.
func DeleteIntegrationCredential(integrationId string) error {
	_ = keyring.Delete(integrationKeyringService, integrationId)

	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	if settings.IntegrationCreds == nil {
		return nil
	}
	delete(settings.IntegrationCreds, integrationId)
	return saveSettings(settings)
}

// ========== Dependency Preset Management ==========

// GetProjectDependencyPresets retrieves all dependency presets for a project.
func GetProjectDependencyPresets(projectId string) ([]interface{}, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return []interface{}{}, err
	}

	presets, exists := settings.DependencyPresets[projectId]
	if !exists {
		return []interface{}{}, nil
	}
	return presets, nil
}

// AddDependencyPreset adds a dependency preset to a project.
// If a preset with the same name exists, it will be overwritten.
func AddDependencyPreset(projectId string, presetData interface{}) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}

	if settings.DependencyPresets == nil {
		settings.DependencyPresets = make(map[string][]interface{})
	}

	if _, exists := settings.DependencyPresets[projectId]; !exists {
		settings.DependencyPresets[projectId] = []interface{}{}
	}

	projectPresets := settings.DependencyPresets[projectId]
	presetName := presetData.(map[string]interface{})["name"]

	// Remove existing preset with the same name if it exists
	for i, preset := range projectPresets {
		if presetName == preset.(map[string]interface{})["name"] {
			projectPresets = append(projectPresets[:i], projectPresets[i+1:]...)
			break
		}
	}

	projectPresets = append(projectPresets, presetData)
	settings.DependencyPresets[projectId] = projectPresets
	return saveSettings(settings)
}

// RemoveDependencyPreset removes a dependency preset from a project.
func RemoveDependencyPreset(projectId string, presetName string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	projectPresets := settings.DependencyPresets[projectId]
	for i, preset := range projectPresets {
		if presetName == preset.(map[string]interface{})["name"] {
			projectPresets = append(projectPresets[:i], projectPresets[i+1:]...)
			break
		}
	}
	settings.DependencyPresets[projectId] = projectPresets
	return saveSettings(settings)
}

// UpdateDependencyPreset updates an existing dependency preset.
func UpdateDependencyPreset(projectId string, presetName string, updatedPreset interface{}) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	projectPresets := settings.DependencyPresets[projectId]
	for i, preset := range projectPresets {
		if presetName == preset.(map[string]interface{})["name"] {
			projectPresets[i] = updatedPreset
			break
		}
	}
	settings.DependencyPresets[projectId] = projectPresets
	return saveSettings(settings)
}

// ========== Ignore List Preset Management ==========

// GetIgnoreListPresets retrieves all user-defined ignore list presets.
// Returns a map of preset name to its list of ignore patterns.
func GetIgnoreListPresets() (map[string][]string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return map[string][]string{}, err
	}
	if settings.IgnoreListPresets == nil {
		return map[string][]string{}, nil
	}
	return settings.IgnoreListPresets, nil
}

// AddIgnoreListPreset stores an ignore list preset under the given name.
// If a preset with the same name already exists, it is overwritten.
func AddIgnoreListPreset(name string, entries []string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("preset name cannot be empty")
	}
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	if settings.IgnoreListPresets == nil {
		settings.IgnoreListPresets = make(map[string][]string)
	}
	if entries == nil {
		entries = []string{}
	}
	settings.IgnoreListPresets[name] = entries
	return saveSettings(settings)
}

// RemoveIgnoreListPreset deletes an ignore list preset by name.
func RemoveIgnoreListPreset(name string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}
	if settings.IgnoreListPresets == nil {
		return nil
	}
	delete(settings.IgnoreListPresets, name)
	return saveSettings(settings)
}

// IsPathAllowed checks whether the given path falls within a registered project directory.
// Allowed roots: ProjectsDir, SharedProjectsDir, all ProjectLocation paths, and OS temp dir.
func IsPathAllowed(path string) bool {
	cleaned := filepath.Clean(path)
	absPath, err := filepath.Abs(cleaned)
	if err != nil {
		return false
	}

	var roots []string

	s, err := loadUserSettings()
	if err != nil {
		return false
	}

	if s.ProjectsDir != "" {
		roots = append(roots, s.ProjectsDir)
	}
	if s.SharedProjectsDir != "" {
		roots = append(roots, s.SharedProjectsDir)
	}
	for _, loc := range s.ProjectLocations {
		if loc.Path != "" {
			roots = append(roots, loc.Path)
		}
	}
	roots = append(roots, os.TempDir())

	for _, root := range roots {
		cleanRoot := filepath.Clean(root)
		absRoot, err := filepath.Abs(cleanRoot)
		if err != nil {
			continue
		}
		if strings.EqualFold(absPath, absRoot) || strings.HasPrefix(strings.ToLower(absPath), strings.ToLower(absRoot)+string(os.PathSeparator)) {
			return true
		}
	}

	activeProjectWorkingDirMu.RLock()
	activeDir := activeProjectWorkingDir
	activeProjectWorkingDirMu.RUnlock()
	if activeDir != "" {
		lowerAbs := strings.ToLower(absPath)
		if lowerAbs == activeDir || strings.HasPrefix(lowerAbs, activeDir+string(os.PathSeparator)) {
			return true
		}
	}

	return false
}

// ========== Project Location Management ==========

// GetAllLocationPaths returns all configured project locations
func GetAllLocationPaths() ([]ProjectLocation, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return []ProjectLocation{}, err
	}

	// Auto-migrate if needed
	if len(settings.ProjectLocations) == 0 && settings.WorkingDir != "" {
		err = migrateWorkingDirectory()
		if err != nil {
			return []ProjectLocation{}, err
		}
		// Reload settings after migration
		settings, err = loadUserSettings()
		if err != nil {
			return []ProjectLocation{}, err
		}
	}

	// Resolve bookmarks on macOS
	if runtime.GOOS == "darwin" {
		for i := range settings.ProjectLocations {
			if len(settings.ProjectLocations[i].Bookmark) > 0 && !IsBookmarkStale(settings.ProjectLocations[i].Bookmark) {
				resolvedPath, err := ResolveBookmark(settings.ProjectLocations[i].Bookmark)
				if err == nil {
					if _, err := os.Stat(resolvedPath); err == nil {
						settings.ProjectLocations[i].Path = resolvedPath
					}
				}
			}
		}
	}

	return settings.ProjectLocations, nil
}

// GetDefaultLocation returns the default project location
func GetDefaultLocation() (ProjectLocation, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return ProjectLocation{}, err
	}

	// Auto-migrate if needed
	if len(settings.ProjectLocations) == 0 && settings.WorkingDir != "" {
		err = migrateWorkingDirectory()
		if err != nil {
			return ProjectLocation{}, err
		}
		settings, err = loadUserSettings()
		if err != nil {
			return ProjectLocation{}, err
		}
	}

	// Find default location
	for _, loc := range settings.ProjectLocations {
		if loc.IsDefault {
			// Resolve bookmark on macOS
			if runtime.GOOS == "darwin" {
				if len(loc.Bookmark) > 0 && !IsBookmarkStale(loc.Bookmark) {
					resolvedPath, err := ResolveBookmark(loc.Bookmark)
					if err == nil {
						if _, err := os.Stat(resolvedPath); err == nil {
							loc.Path = resolvedPath
						}
					}
				}
			}
			return loc, nil
		}
	}

	// If no default found but locations exist, return first one
	if len(settings.ProjectLocations) > 0 {
		return settings.ProjectLocations[0], nil
	}

	return ProjectLocation{}, fmt.Errorf("no project locations configured")
}

// AddProjectLocation adds a new project location
func AddProjectLocation(name, path string) (ProjectLocation, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return ProjectLocation{}, err
	}

	// Initialize if needed
	if settings.ProjectLocations == nil {
		settings.ProjectLocations = []ProjectLocation{}
	}

	// Check for duplicate paths
	for _, loc := range settings.ProjectLocations {
		if loc.Path == path {
			return ProjectLocation{}, fmt.Errorf("location with path %s already exists", path)
		}
	}

	// Create new location
	newLocation := ProjectLocation{
		ID:         uuid.NewString(),
		Name:       name,
		Path:       path,
		IsDefault:  len(settings.ProjectLocations) == 0, // First location is default
		ProjectIDs: []string{},
	}

	// Create bookmark on macOS
	if runtime.GOOS == "darwin" {
		bookmarkData, err := CreateBookmarkFromPath(path)
		if err != nil {
			log.Printf("Failed to create bookmark for location %s: %v", name, err)
		}
		newLocation.Bookmark = bookmarkData
	}

	settings.ProjectLocations = append(settings.ProjectLocations, newLocation)

	// Set as default if it's the first location
	if newLocation.IsDefault {
		settings.DefaultLocationID = newLocation.ID
	}

	err = saveSettings(settings)
	if err != nil {
		return ProjectLocation{}, err
	}

	return newLocation, nil
}

// RemoveProjectLocation removes a project location by ID
func RemoveProjectLocation(locationID string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}

	// Can't remove if only one location
	if len(settings.ProjectLocations) <= 1 {
		return fmt.Errorf("cannot remove the last project location")
	}

	// Check if location has projects assigned
	for i, loc := range settings.ProjectLocations {
		if loc.ID == locationID {
			if len(loc.ProjectIDs) > 0 {
				return fmt.Errorf("cannot remove location: %d project(s) are using it", len(loc.ProjectIDs))
			}

			// Remove location
			settings.ProjectLocations = append(settings.ProjectLocations[:i], settings.ProjectLocations[i+1:]...)

			// If removing default, set first remaining location as default
			if loc.IsDefault && len(settings.ProjectLocations) > 0 {
				settings.ProjectLocations[0].IsDefault = true
				settings.DefaultLocationID = settings.ProjectLocations[0].ID
			}

			return saveSettings(settings)
		}
	}

	return fmt.Errorf("location with ID %s not found", locationID)
}

// UpdateProjectLocation updates an existing project location
func UpdateProjectLocation(locationID, name, path string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}

	for i, loc := range settings.ProjectLocations {
		if loc.ID == locationID {
			// Update name
			if name != "" {
				settings.ProjectLocations[i].Name = name
			}

			// Update path and bookmark
			if path != "" {
				// Check for duplicate paths (excluding current location) only when the path changes
				if path != loc.Path {
					for j, otherLoc := range settings.ProjectLocations {
						if j != i && otherLoc.Path == path {
							return fmt.Errorf("location with path %s already exists", path)
						}
					}

					settings.ProjectLocations[i].Path = path
				}

				// Always (re)create the bookmark on macOS so that re-selecting the same
				// folder re-establishes security-scoped access for stale bookmarks.
				if runtime.GOOS == "darwin" {
					bookmarkData, err := CreateBookmarkFromPath(path)
					if err != nil {
						log.Printf("Failed to create bookmark for location %s: %v", name, err)
					}
					settings.ProjectLocations[i].Bookmark = bookmarkData
				}
			}

			return saveSettings(settings)
		}
	}

	return fmt.Errorf("location with ID %s not found", locationID)
}

// SetDefaultLocation sets a location as the default
func SetDefaultLocation(locationID string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}

	found := false
	for i := range settings.ProjectLocations {
		if settings.ProjectLocations[i].ID == locationID {
			settings.ProjectLocations[i].IsDefault = true
			settings.DefaultLocationID = locationID
			found = true
		} else {
			settings.ProjectLocations[i].IsDefault = false
		}
	}

	if !found {
		return fmt.Errorf("location with ID %s not found", locationID)
	}

	return saveSettings(settings)
}

// AssignProjectToLocation assigns a project to a specific location
func AssignProjectToLocation(projectID, locationID string) error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}

	// Remove project from all locations first
	for i := range settings.ProjectLocations {
		for j, pid := range settings.ProjectLocations[i].ProjectIDs {
			if pid == projectID {
				settings.ProjectLocations[i].ProjectIDs = append(
					settings.ProjectLocations[i].ProjectIDs[:j],
					settings.ProjectLocations[i].ProjectIDs[j+1:]...,
				)
				break
			}
		}
	}

	// Add project to specified location
	for i := range settings.ProjectLocations {
		if settings.ProjectLocations[i].ID == locationID {
			// Check if already assigned
			for _, pid := range settings.ProjectLocations[i].ProjectIDs {
				if pid == projectID {
					return nil // Already assigned
				}
			}
			settings.ProjectLocations[i].ProjectIDs = append(settings.ProjectLocations[i].ProjectIDs, projectID)
			return saveSettings(settings)
		}
	}

	return fmt.Errorf("location with ID %s not found", locationID)
}

// GetProjectLocation returns the location ID for a project
func GetProjectLocation(projectID string) (string, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return "", err
	}

	for _, loc := range settings.ProjectLocations {
		for _, pid := range loc.ProjectIDs {
			if pid == projectID {
				return loc.ID, nil
			}
		}
	}

	return "", fmt.Errorf("project not assigned to any location")
}

// GetLocationUsage returns the count of projects using a location
func GetLocationUsage(locationID string) (int, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return 0, err
	}

	for _, loc := range settings.ProjectLocations {
		if loc.ID == locationID {
			return len(loc.ProjectIDs), nil
		}
	}

	return 0, fmt.Errorf("location with ID %s not found", locationID)
}

// CanDeleteLocation returns true if location can be safely deleted
func CanDeleteLocation(locationID string) (bool, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return false, err
	}

	// Can't remove if only one location
	if len(settings.ProjectLocations) <= 1 {
		return false, nil
	}

	for _, loc := range settings.ProjectLocations {
		if loc.ID == locationID {
			return len(loc.ProjectIDs) == 0, nil
		}
	}

	return false, fmt.Errorf("location with ID %s not found", locationID)
}

// LocationHealth represents the health status of a location
type LocationHealth struct {
	ID        string `json:"id"`
	Exists    bool   `json:"exists"`
	Writable  bool   `json:"writable"`
	FreeSpace int64  `json:"free_space"`
	Stale     bool   `json:"stale"` // darwin: security-scoped bookmark needs re-selection
}

// CheckLocationHealth checks the health of a specific location
func CheckLocationHealth(locationID string) (LocationHealth, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return LocationHealth{}, err
	}

	for _, loc := range settings.ProjectLocations {
		if loc.ID == locationID {
			health := LocationHealth{
				ID:        locationID,
				Exists:    false,
				Writable:  false,
				FreeSpace: 0,
			}

			// On macOS, a security-scoped bookmark can go stale even when the
			// path still exists, leaving the sandbox without access to the folder.
			if runtime.GOOS == "darwin" && len(loc.Bookmark) > 0 {
				health.Stale = IsBookmarkStale(loc.Bookmark)
			}

			// Check if path exists
			if _, err := os.Stat(loc.Path); err == nil {
				health.Exists = true

				// Check if writable
				testFile := filepath.Join(loc.Path, ".clustta_write_test")
				if err := os.WriteFile(testFile, []byte("test"), 0644); err == nil {
					health.Writable = true
					os.Remove(testFile)
				}

				// TODO: Get free space if needed
			}

			return health, nil
		}
	}

	return LocationHealth{}, fmt.Errorf("location with ID %s not found", locationID)
}

// CheckAllLocationsHealth checks the health of all locations
func CheckAllLocationsHealth() ([]LocationHealth, error) {
	settings, err := loadUserSettings()
	if err != nil {
		return []LocationHealth{}, err
	}

	healthStatuses := []LocationHealth{}
	for _, loc := range settings.ProjectLocations {
		health, err := CheckLocationHealth(loc.ID)
		if err != nil {
			log.Printf("Error checking health for location %s: %v", loc.Name, err)
			continue
		}
		healthStatuses = append(healthStatuses, health)
	}

	return healthStatuses, nil
}

// SystemBookmarksHealth reports staleness of the top-level projects directory bookmarks.
type SystemBookmarksHealth struct {
	ProjectsDirStale       bool `json:"projects_dir_stale"`
	SharedProjectsDirStale bool `json:"shared_projects_dir_stale"`
}

// CheckSystemBookmarksHealth reports whether the projects or shared projects
// directory security-scoped bookmarks have gone stale on macOS.
func CheckSystemBookmarksHealth() (SystemBookmarksHealth, error) {
	health := SystemBookmarksHealth{}
	if runtime.GOOS != "darwin" {
		return health, nil
	}

	settings, err := loadUserSettings()
	if err != nil {
		return health, err
	}

	if len(settings.ProjectsDirBookmark) > 0 {
		health.ProjectsDirStale = IsBookmarkStale(settings.ProjectsDirBookmark)
	}
	if len(settings.SharedProjectsDirBookmark) > 0 {
		health.SharedProjectsDirStale = IsBookmarkStale(settings.SharedProjectsDirBookmark)
	}

	return health, nil
}

// migrateWorkingDirectory migrates old WorkingDirectory setting to ProjectLocations
func migrateWorkingDirectory() error {
	settings, err := loadUserSettings()
	if err != nil {
		return err
	}

	// Check if migration needed
	if len(settings.ProjectLocations) == 0 && settings.WorkingDir != "" {
		log.Printf("Migrating WorkingDirectory to ProjectLocations: %s", settings.WorkingDir)

		location := ProjectLocation{
			ID:         "1",
			Name:       "Default",
			Path:       settings.WorkingDir,
			IsDefault:  true,
			ProjectIDs: []string{},
		}

		// Create bookmark if macOS
		if runtime.GOOS == "darwin" {
			bookmarkData, err := CreateBookmarkFromPath(settings.WorkingDir)
			if err != nil {
				log.Printf("Failed to create bookmark during migration: %v", err)
			}
			location.Bookmark = bookmarkData
		}

		settings.ProjectLocations = []ProjectLocation{location}
		settings.DefaultLocationID = location.ID

		// Clear old fields
		settings.WorkingDir = ""
		settings.WorkingDirBookmark = nil

		return saveSettings(settings)
	}

	return nil
}
