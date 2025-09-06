package settings

import (
	"clustta/internal/auth_service"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Project struct {
	Id   int    `json:"id,omitempty"`
	Path string `json:"path,omitempty"`
}

type Mode string

const (
	ServerMode Mode = "server"
	ClientMode Mode = "client"
)

type RuntimeConfig struct {
	Mode Mode
	// Add other configuration fields as needed
}

var (
	currentMode = ClientMode // Default to client mode
	initialized bool
	mu          sync.RWMutex
)

func InitializeServer() error {
	mu.Lock()
	defer mu.Unlock()

	if initialized {
		return fmt.Errorf("runtime already initialized")
	}

	currentMode = ServerMode
	initialized = true
	return nil
}

// IsServer returns true if running in server mode
func IsServer() bool {
	mu.RLock()
	defer mu.RUnlock()
	return currentMode == ServerMode
}

// IsClient returns true if running in client mode
func IsClient() bool {
	mu.RLock()
	defer mu.RUnlock()
	return currentMode == ClientMode
}

// GetMode returns the current mode
func GetMode() Mode {
	mu.RLock()
	defer mu.RUnlock()
	return currentMode
}

func GetRoamingPath() (string, error) {
	roamingPath, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return roamingPath, nil
}

func GetLogPath() (string, error) {
	roamingPath, err := GetRoamingPath()
	if err != nil {
		return "", err
	}
	settingsPath := filepath.Join(roamingPath, "Clustta", "clustta.log")
	os.MkdirAll(filepath.Join(roamingPath, "Clustta"), os.ModePerm)
	return settingsPath, nil
}
func GetSettingsPath() (string, error) {
	roamingPath, err := GetRoamingPath()
	if err != nil {
		return "", err
	}
	settingsPath := filepath.Join(roamingPath, "Clustta", ".settings.json")
	return settingsPath, nil
}
func GetUserSettingsPath() (string, error) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return "", err
	}
	roamingPath, err := GetRoamingPath()
	if err != nil {
		return "", err
	}
	settingsPath := filepath.Join(roamingPath, "Clustta", user.Id+".json")
	return settingsPath, nil
}

func GetUserProjectTemplatesPath() (string, error) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return "", err
	}
	roamingPath, err := GetRoamingPath()
	if err != nil {
		return "", err
	}
	projectTemplatesPath := filepath.Join(roamingPath, "Clustta", user.Id, "project_templates")
	os.MkdirAll(projectTemplatesPath, os.ModePerm)
	return projectTemplatesPath, nil
}

func GetUserDataFolder(user auth_service.User) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	userDataDir := filepath.Join(homeDir, "Documents", "clustta", user.Username)
	os.MkdirAll(userDataDir, os.ModePerm)
	return userDataDir, nil
}

func GetLocalWorkingDirectory() (string, error) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return "", err
	}

	userDataFolder, err := GetUserDataFolder(user)
	if err != nil {
		return "", err
	}
	// workingDir := filepath.Join(userDataFolder, "working_directory")
	// os.MkdirAll(workingDir, os.ModePerm)
	return userDataFolder, nil
}
