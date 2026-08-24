package repository

import (
	"clustta/internal/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

const (
	ProjectScriptSettingsConfig   = "project_script_settings_v1"
	ProjectScriptSettingsVersion  = 1
	DefaultProjectScriptDirectory = "Scripts"
)

var DefaultProjectScriptExtensions = []string{".py"}

type ProjectScriptSettings struct {
	Version    int      `json:"version"`
	Directory  string   `json:"directory"`
	Extensions []string `json:"extensions"`
}

func GetProjectScriptSettings(tx *sqlx.Tx) (ProjectScriptSettings, error) {
	settings := ProjectScriptSettings{
		Version:    ProjectScriptSettingsVersion,
		Directory:  DefaultProjectScriptDirectory,
		Extensions: append([]string(nil), DefaultProjectScriptExtensions...),
	}
	var settingsJSON string
	if err := tx.Get(&settingsJSON, "SELECT value FROM config WHERE name = ?", ProjectScriptSettingsConfig); err != nil {
		if err == sql.ErrNoRows {
			return settings, nil
		}
		return ProjectScriptSettings{}, err
	}
	if err := json.Unmarshal([]byte(settingsJSON), &settings); err != nil {
		return ProjectScriptSettings{}, fmt.Errorf("invalid project script settings: %w", err)
	}
	return NormalizeProjectScriptSettings(settings)
}

func SetProjectScriptSettings(tx *sqlx.Tx, settings ProjectScriptSettings) error {
	normalized, err := NormalizeProjectScriptSettings(settings)
	if err != nil {
		return err
	}
	settingsJSON, err := json.Marshal(normalized)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO config (name, value, mtime, synced)
		VALUES (?, ?, ?, 0)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime, synced = 0
	`, ProjectScriptSettingsConfig, string(settingsJSON), utils.GetEpochTime())
	return err
}

func NormalizeProjectScriptSettings(settings ProjectScriptSettings) (ProjectScriptSettings, error) {
	directory := filepath.Clean(filepath.FromSlash(strings.TrimSpace(settings.Directory)))
	if directory == "." || directory == "" {
		return ProjectScriptSettings{}, fmt.Errorf("scripts directory is required")
	}
	if !filepath.IsLocal(directory) {
		return ProjectScriptSettings{}, fmt.Errorf("scripts directory must be relative to the project root")
	}

	extensions := make([]string, 0, len(settings.Extensions))
	seen := map[string]bool{}
	for _, extension := range settings.Extensions {
		extension = strings.ToLower(strings.TrimSpace(extension))
		extension = strings.TrimPrefix(extension, "*")
		if extension == "" {
			continue
		}
		if !strings.HasPrefix(extension, ".") {
			extension = "." + extension
		}
		if strings.ContainsAny(extension, `/\`) || extension == "." {
			return ProjectScriptSettings{}, fmt.Errorf("invalid script extension %q", extension)
		}
		if !seen[extension] {
			seen[extension] = true
			extensions = append(extensions, extension)
		}
	}
	sort.Strings(extensions)
	return ProjectScriptSettings{
		Version: ProjectScriptSettingsVersion, Directory: filepath.ToSlash(directory), Extensions: extensions,
	}, nil
}
