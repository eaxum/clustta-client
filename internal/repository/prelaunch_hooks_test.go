package repository

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestNormalizePreLaunchHookSettings(t *testing.T) {
	settings, err := NormalizePreLaunchHookSettings(PreLaunchHookSettings{Hooks: []PreLaunchHook{{
		Name: " Blender setup ", Enabled: true, Extensions: []string{"blend", "*.BLEND"},
		ScriptAssetIDs: []string{"script-id"}, FailurePolicy: "",
	}}})

	require.NoError(t, err)
	require.Equal(t, PreLaunchHooksVersion, settings.Version)
	require.NotEmpty(t, settings.Hooks[0].ID)
	require.Equal(t, "Blender setup", settings.Hooks[0].Name)
	require.Equal(t, []string{".blend"}, settings.Hooks[0].Extensions)
	require.Equal(t, PreLaunchFailureBlock, settings.Hooks[0].FailurePolicy)
}

func TestNormalizePreLaunchHookSettingsRejectsAmbiguousMatches(t *testing.T) {
	_, err := NormalizePreLaunchHookSettings(PreLaunchHookSettings{Hooks: []PreLaunchHook{
		{Name: "One", Enabled: true, Extensions: []string{".ma"}, ScriptAssetIDs: []string{"one"}},
		{Name: "Two", Enabled: true, Extensions: []string{".ma"}, ScriptAssetIDs: []string{"two"}},
	}})

	require.ErrorContains(t, err, "both match .ma")
}

func TestNormalizePreLaunchHookSettingsValidatesEnvironmentVariables(t *testing.T) {
	settings, err := NormalizePreLaunchHookSettings(PreLaunchHookSettings{
		EnvironmentVariables: []PreLaunchEnvironmentVariable{{
			ID: "ocio", Name: "OCIO", Value: "<ProjectRoot>/config.ocio",
		}},
		Hooks: []PreLaunchHook{{
			Name: "OCIO", Enabled: true, Extensions: []string{".blend"},
			EnvironmentVariableIDs: []string{"ocio"},
		}},
	})

	require.NoError(t, err)
	require.Equal(t, "ocio", settings.Hooks[0].EnvironmentVariableIDs[0])
}

func TestNormalizePreLaunchHookSettingsRejectsMissingEnvironmentReference(t *testing.T) {
	_, err := NormalizePreLaunchHookSettings(PreLaunchHookSettings{Hooks: []PreLaunchHook{{
		Name: "OCIO", Enabled: true, Extensions: []string{".blend"},
		EnvironmentVariableIDs: []string{"missing"},
	}}})

	require.ErrorContains(t, err, "does not exist")
}

func TestNormalizePreLaunchHookSettingsRequiresTrackedScriptReference(t *testing.T) {
	_, err := NormalizePreLaunchHookSettings(PreLaunchHookSettings{Hooks: []PreLaunchHook{{
		Name: "Custom", Enabled: true, Extensions: []string{".blend"},
	}}})

	require.ErrorContains(t, err, "at least one script asset or environment variable")
}

func TestPreLaunchDCCForExtension(t *testing.T) {
	require.Equal(t, PreLaunchDCCMaya, PreLaunchDCCForExtension(".MA"))
	require.Equal(t, PreLaunchDCCBlender, PreLaunchDCCForExtension(".blend"))
	require.Empty(t, PreLaunchDCCForExtension(".usd"))
}

func TestNormalizePreLaunchHookSettingsRejectsMultipleScripts(t *testing.T) {
	_, err := NormalizePreLaunchHookSettings(PreLaunchHookSettings{Hooks: []PreLaunchHook{{
		Name: "Blender", Enabled: true, Extensions: []string{".blend"},
		ScriptAssetIDs: []string{"one", "two"},
	}}})

	require.ErrorContains(t, err, "only one script asset")
}

func TestNormalizePreLaunchHookSettingsAcceptsApplicationVersion(t *testing.T) {
	settings, err := NormalizePreLaunchHookSettings(PreLaunchHookSettings{Hooks: []PreLaunchHook{{
		Name: "Blender", Enabled: true, Extensions: []string{".blend"},
		ApplicationVersion: " 5.2.0 ", ScriptAssetIDs: []string{"script"},
	}}})

	require.NoError(t, err)
	require.Equal(t, "5.2.0", settings.Hooks[0].ApplicationVersion)
}

func TestNormalizePreLaunchHookSettingsRejectsVersionAcrossDCCs(t *testing.T) {
	_, err := NormalizePreLaunchHookSettings(PreLaunchHookSettings{Hooks: []PreLaunchHook{{
		Name: "Mixed", Enabled: true, Extensions: []string{".blend", ".ma"},
		ApplicationVersion: "2025", ScriptAssetIDs: []string{"script"},
	}}})

	require.ErrorContains(t, err, "cannot apply to multiple DCC")
}

func TestApplySyncableProjectConfigsUsesExistingConfigTable(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	t.Cleanup(func() { db.Close() })
	db.MustExec(`CREATE TABLE config (name TEXT PRIMARY KEY, value CLOB, mtime INTEGER NOT NULL, synced BOOLEAN DEFAULT 0 NOT NULL)`)
	tx := db.MustBegin()

	err := ApplySyncableProjectConfigs(tx, []ProjectConfig{{
		Name: PreLaunchHooksConfig, Value: `{"version":1,"hooks":[]}`, Mtime: 42,
	}})
	require.NoError(t, err)
	configs, err := GetSyncableProjectConfigs(tx, false)
	require.NoError(t, err)
	require.Len(t, configs, 1)
	require.Equal(t, 42, configs[0].Mtime)
	require.True(t, configs[0].Synced)
}

func TestApplySyncableProjectConfigsRejectsUnknownKeys(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	t.Cleanup(func() { db.Close() })
	db.MustExec(`CREATE TABLE config (name TEXT PRIMARY KEY, value CLOB, mtime INTEGER NOT NULL, synced BOOLEAN DEFAULT 0 NOT NULL)`)
	tx := db.MustBegin()

	err := ApplySyncableProjectConfigs(tx, []ProjectConfig{{Name: "remote", Value: "unsafe", Mtime: 1}})
	require.ErrorContains(t, err, "not syncable")
}
