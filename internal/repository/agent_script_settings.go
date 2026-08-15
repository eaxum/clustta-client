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
	AgentScriptDirectoryConfig  = "agent_script_directory"
	AgentScriptExtensionsConfig = "agent_script_extensions"
	DefaultAgentScriptDirectory = "Scripts"
)

var DefaultAgentScriptExtensions = []string{".py"}

type AgentScriptSettings struct {
	Directory  string   `json:"directory"`
	Extensions []string `json:"extensions"`
}

func GetAgentScriptSettings(tx *sqlx.Tx) (AgentScriptSettings, error) {
	settings := AgentScriptSettings{
		Directory:  DefaultAgentScriptDirectory,
		Extensions: append([]string(nil), DefaultAgentScriptExtensions...),
	}
	var directory string
	if err := tx.Get(&directory, "SELECT value FROM config WHERE name = ?", AgentScriptDirectoryConfig); err == nil {
		settings.Directory = directory
	} else if err != sql.ErrNoRows {
		return AgentScriptSettings{}, err
	}
	var extensionsJSON string
	if err := tx.Get(&extensionsJSON, "SELECT value FROM config WHERE name = ?", AgentScriptExtensionsConfig); err == nil {
		if err := json.Unmarshal([]byte(extensionsJSON), &settings.Extensions); err != nil {
			return AgentScriptSettings{}, fmt.Errorf("invalid agent script extensions: %w", err)
		}
	} else if err != sql.ErrNoRows {
		return AgentScriptSettings{}, err
	}
	return NormalizeAgentScriptSettings(settings)
}

func SetAgentScriptSettings(tx *sqlx.Tx, settings AgentScriptSettings) error {
	normalized, err := NormalizeAgentScriptSettings(settings)
	if err != nil {
		return err
	}
	extensionsJSON, err := json.Marshal(normalized.Extensions)
	if err != nil {
		return err
	}
	for name, value := range map[string]string{
		AgentScriptDirectoryConfig:  normalized.Directory,
		AgentScriptExtensionsConfig: string(extensionsJSON),
	} {
		if _, err := tx.Exec(`
			INSERT INTO config (name, value, mtime, synced)
			VALUES (?, ?, ?, 0)
			ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime, synced = 0
		`, name, value, utils.GetEpochTime()); err != nil {
			return err
		}
	}
	return nil
}

func NormalizeAgentScriptSettings(settings AgentScriptSettings) (AgentScriptSettings, error) {
	directory := filepath.Clean(filepath.FromSlash(strings.TrimSpace(settings.Directory)))
	if directory == "." || directory == "" {
		return AgentScriptSettings{}, fmt.Errorf("scripts directory is required")
	}
	if !filepath.IsLocal(directory) {
		return AgentScriptSettings{}, fmt.Errorf("scripts directory must be relative to the project root")
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
		if strings.ContainsAny(extension, `/\\`) || extension == "." {
			return AgentScriptSettings{}, fmt.Errorf("invalid script extension %q", extension)
		}
		if !seen[extension] {
			seen[extension] = true
			extensions = append(extensions, extension)
		}
	}
	sort.Strings(extensions)
	return AgentScriptSettings{Directory: filepath.ToSlash(directory), Extensions: extensions}, nil
}
