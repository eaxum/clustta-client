package repository

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestNormalizeProjectScriptSettings(t *testing.T) {
	settings, err := NormalizeProjectScriptSettings(ProjectScriptSettings{
		Directory:  "tools/scripts",
		Extensions: []string{"PY", "*.bat", ".py", " .MEL "},
	})

	require.NoError(t, err)
	require.Equal(t, ProjectScriptSettingsVersion, settings.Version)
	require.Equal(t, "tools/scripts", settings.Directory)
	require.Equal(t, []string{".bat", ".mel", ".py"}, settings.Extensions)
}

func TestNormalizeProjectScriptSettingsRejectsOutsideDirectory(t *testing.T) {
	_, err := NormalizeProjectScriptSettings(ProjectScriptSettings{Directory: "../scripts"})
	require.ErrorContains(t, err, "relative to the project root")
}

func TestNormalizeProjectScriptSettingsRejectsAbsoluteDirectory(t *testing.T) {
	_, err := NormalizeProjectScriptSettings(ProjectScriptSettings{Directory: `C:\scripts`})
	require.ErrorContains(t, err, "relative to the project root")
}

func TestProjectScriptSettingsUseSingleConfigRow(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	t.Cleanup(func() { db.Close() })
	db.MustExec(`CREATE TABLE config (name TEXT PRIMARY KEY, value CLOB, mtime INTEGER NOT NULL, synced BOOLEAN DEFAULT 0 NOT NULL)`)
	tx := db.MustBegin()

	err := SetProjectScriptSettings(tx, ProjectScriptSettings{
		Directory: "Pipeline/Scripts", Extensions: []string{".py", ".mel"},
	})
	require.NoError(t, err)

	var names []string
	require.NoError(t, tx.Select(&names, "SELECT name FROM config"))
	require.Equal(t, []string{ProjectScriptSettingsConfig}, names)

	settings, err := GetProjectScriptSettings(tx)
	require.NoError(t, err)
	require.Equal(t, "Pipeline/Scripts", settings.Directory)
	require.Equal(t, []string{".mel", ".py"}, settings.Extensions)
}
