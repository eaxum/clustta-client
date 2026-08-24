package repository

import (
	"clustta/internal/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	PreLaunchHooksConfig  = "dcc_prelaunch_hooks_v1"
	PreLaunchHooksVersion = 1

	PreLaunchFailureBlock = "block"
	PreLaunchFailureWarn  = "warn"
	PreLaunchDCCMaya      = "maya"
	PreLaunchDCCBlender   = "blender"
)

var environmentVariableNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
var applicationVersionPattern = regexp.MustCompile(`^[0-9]+(?:\.[0-9]+){0,2}$`)

var SyncableProjectConfigNames = []string{
	ProjectScriptSettingsConfig,
	PreLaunchHooksConfig,
}

type PreLaunchEnvironmentVariable struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PreLaunchHook struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	Enabled                bool     `json:"enabled"`
	Extensions             []string `json:"extensions"`
	ApplicationVersion     string   `json:"application_version"`
	ScriptAssetIDs         []string `json:"script_asset_ids"`
	EnvironmentVariableIDs []string `json:"environment_variable_ids"`
	FailurePolicy          string   `json:"failure_policy"`
}

type PreLaunchHookSettings struct {
	Version              int                            `json:"version"`
	EnvironmentVariables []PreLaunchEnvironmentVariable `json:"environment_variables"`
	Hooks                []PreLaunchHook                `json:"hooks"`
}

func GetSyncableProjectConfigs(tx *sqlx.Tx, changedOnly bool) ([]ProjectConfig, error) {
	query, args, err := sqlx.In("SELECT name, value, mtime, synced FROM config WHERE name IN (?)", SyncableProjectConfigNames)
	if err != nil {
		return nil, err
	}
	if changedOnly {
		query += " AND synced = 0"
	}
	configs := []ProjectConfig{}
	if err := tx.Select(&configs, tx.Rebind(query), args...); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return configs, nil
}

func ApplySyncableProjectConfigs(tx *sqlx.Tx, configs []ProjectConfig) error {
	allowed := map[string]bool{}
	for _, name := range SyncableProjectConfigNames {
		allowed[name] = true
	}
	for _, config := range configs {
		if !allowed[config.Name] {
			return fmt.Errorf("project config %q is not syncable", config.Name)
		}
		if _, err := tx.Exec(`
			INSERT INTO config (name, value, mtime, synced)
			VALUES (?, ?, ?, 1)
			ON CONFLICT (name) DO UPDATE SET
				value = EXCLUDED.value,
				mtime = EXCLUDED.mtime,
				synced = 1
			WHERE EXCLUDED.mtime >= config.mtime
		`, config.Name, config.Value, config.Mtime); err != nil {
			return err
		}
	}
	return nil
}

func MarkSyncableProjectConfigsSynced(tx *sqlx.Tx) error {
	query, args, err := sqlx.In("UPDATE config SET synced = 1 WHERE name IN (?)", SyncableProjectConfigNames)
	if err != nil {
		return err
	}
	_, err = tx.Exec(tx.Rebind(query), args...)
	return err
}

func GetPreLaunchHookSettings(tx *sqlx.Tx) (PreLaunchHookSettings, error) {
	settings := PreLaunchHookSettings{Version: PreLaunchHooksVersion, Hooks: []PreLaunchHook{}}
	var settingsJSON string
	if err := tx.Get(&settingsJSON, "SELECT value FROM config WHERE name = ?", PreLaunchHooksConfig); err != nil {
		if err == sql.ErrNoRows {
			return settings, nil
		}
		return PreLaunchHookSettings{}, err
	}
	if err := json.Unmarshal([]byte(settingsJSON), &settings); err != nil {
		return PreLaunchHookSettings{}, fmt.Errorf("invalid pre-launch hook settings: %w", err)
	}
	return NormalizePreLaunchHookSettings(settings)
}

func SetPreLaunchHookSettings(tx *sqlx.Tx, settings PreLaunchHookSettings) error {
	normalized, err := NormalizePreLaunchHookSettings(settings)
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
	`, PreLaunchHooksConfig, string(settingsJSON), utils.GetEpochTime())
	return err
}

func NormalizePreLaunchHookSettings(settings PreLaunchHookSettings) (PreLaunchHookSettings, error) {
	settings.Version = PreLaunchHooksVersion
	variables, err := normalizeEnvironmentVariables(settings.EnvironmentVariables)
	if err != nil {
		return PreLaunchHookSettings{}, err
	}
	settings.EnvironmentVariables = variables
	variableIDs := make(map[string]bool, len(variables))
	for _, variable := range variables {
		variableIDs[variable.ID] = true
	}
	seenIDs := map[string]bool{}
	seenExtensions := map[string]string{}
	for index := range settings.Hooks {
		hook, err := normalizePreLaunchHook(settings.Hooks[index], variableIDs)
		if err != nil {
			return PreLaunchHookSettings{}, fmt.Errorf("hook %d: %w", index+1, err)
		}
		if hook.ID == "" {
			hook.ID = uuid.NewString()
		}
		if seenIDs[hook.ID] {
			return PreLaunchHookSettings{}, fmt.Errorf("duplicate hook ID %q", hook.ID)
		}
		seenIDs[hook.ID] = true
		if hook.Enabled {
			for _, extension := range hook.Extensions {
				if existingName, exists := seenExtensions[extension]; exists {
					return PreLaunchHookSettings{}, fmt.Errorf("%s and %s both match %s", existingName, hook.Name, extension)
				}
				seenExtensions[extension] = hook.Name
			}
		}
		settings.Hooks[index] = hook
	}
	return settings, nil
}

func normalizePreLaunchHook(hook PreLaunchHook, environmentVariableIDs map[string]bool) (PreLaunchHook, error) {
	hook.ID = strings.TrimSpace(hook.ID)
	hook.Name = strings.TrimSpace(hook.Name)
	hook.ApplicationVersion = strings.TrimSpace(hook.ApplicationVersion)
	hook.FailurePolicy = strings.ToLower(strings.TrimSpace(hook.FailurePolicy))
	if hook.Name == "" {
		return PreLaunchHook{}, fmt.Errorf("name is required")
	}
	if hook.FailurePolicy == "" {
		hook.FailurePolicy = PreLaunchFailureBlock
	}
	if hook.FailurePolicy != PreLaunchFailureBlock && hook.FailurePolicy != PreLaunchFailureWarn {
		return PreLaunchHook{}, fmt.Errorf("failure policy must be block or warn")
	}
	extensions, err := normalizeHookExtensions(hook.Extensions)
	if err != nil {
		return PreLaunchHook{}, err
	}
	if len(extensions) == 0 {
		return PreLaunchHook{}, fmt.Errorf("at least one file extension is required")
	}
	hook.Extensions = extensions
	if hook.ApplicationVersion != "" {
		if !applicationVersionPattern.MatchString(hook.ApplicationVersion) {
			return PreLaunchHook{}, fmt.Errorf("application version must contain only numeric version components")
		}
		if err := validateVersionedHookExtensions(hook.Extensions); err != nil {
			return PreLaunchHook{}, err
		}
	}
	hook.ScriptAssetIDs = normalizeUniqueStrings(hook.ScriptAssetIDs)
	if len(hook.ScriptAssetIDs) > 1 {
		return PreLaunchHook{}, fmt.Errorf("only one script asset can be attached")
	}
	hook.EnvironmentVariableIDs = normalizeUniqueStrings(hook.EnvironmentVariableIDs)
	for _, variableID := range hook.EnvironmentVariableIDs {
		if !environmentVariableIDs[variableID] {
			return PreLaunchHook{}, fmt.Errorf("environment variable %q does not exist", variableID)
		}
	}
	if len(hook.ScriptAssetIDs) == 0 && len(hook.EnvironmentVariableIDs) == 0 {
		return PreLaunchHook{}, fmt.Errorf("at least one script asset or environment variable is required")
	}
	return hook, nil
}

func validateVersionedHookExtensions(extensions []string) error {
	dcc := ""
	for _, extension := range extensions {
		extensionDCC := PreLaunchDCCForExtension(extension)
		if extensionDCC == "" {
			return fmt.Errorf("application version requires Maya or Blender file extensions")
		}
		if dcc != "" && extensionDCC != dcc {
			return fmt.Errorf("application version cannot apply to multiple DCC applications")
		}
		dcc = extensionDCC
	}
	return nil
}

func PreLaunchDCCForExtension(extension string) string {
	switch strings.ToLower(extension) {
	case ".ma", ".mb":
		return PreLaunchDCCMaya
	case ".blend":
		return PreLaunchDCCBlender
	default:
		return ""
	}
}

func normalizeHookExtensions(values []string) ([]string, error) {
	extensions := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		extension := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(value, "*")))
		if extension == "" {
			continue
		}
		if !strings.HasPrefix(extension, ".") {
			extension = "." + extension
		}
		if extension == "." || strings.ContainsAny(extension, `/\\`) {
			return nil, fmt.Errorf("invalid file extension %q", value)
		}
		if !seen[extension] {
			seen[extension] = true
			extensions = append(extensions, extension)
		}
	}
	sort.Strings(extensions)
	return extensions, nil
}

func normalizeUniqueStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func normalizeEnvironmentVariables(values []PreLaunchEnvironmentVariable) ([]PreLaunchEnvironmentVariable, error) {
	result := make([]PreLaunchEnvironmentVariable, 0, len(values))
	seenIDs := map[string]bool{}
	seenNames := map[string]bool{}
	for _, variable := range values {
		variable.ID = strings.TrimSpace(variable.ID)
		variable.Name = strings.TrimSpace(variable.Name)
		variable.Value = strings.TrimSpace(variable.Value)
		if variable.ID == "" && variable.Name == "" && variable.Value == "" {
			continue
		}
		if variable.ID == "" {
			variable.ID = uuid.NewString()
		}
		if seenIDs[variable.ID] {
			return nil, fmt.Errorf("duplicate environment variable ID %q", variable.ID)
		}
		if !environmentVariableNamePattern.MatchString(variable.Name) {
			return nil, fmt.Errorf("invalid environment variable name %q", variable.Name)
		}
		if variable.Value == "" {
			return nil, fmt.Errorf("environment variable %s requires a value", variable.Name)
		}
		normalizedName := strings.ToUpper(variable.Name)
		if seenNames[normalizedName] {
			return nil, fmt.Errorf("duplicate environment variable %q", variable.Name)
		}
		seenIDs[variable.ID] = true
		seenNames[normalizedName] = true
		result = append(result, variable)
	}
	return result, nil
}
