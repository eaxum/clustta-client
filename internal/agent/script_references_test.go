package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResolveScriptDirectoryFromProjectRoot(t *testing.T) {
	projectRoot := t.TempDir()
	scriptDirectory := filepath.Join(projectRoot, "tools", "scripts")
	require.NoError(t, os.MkdirAll(scriptDirectory, 0o755))

	resolved, err := resolveScriptDirectory(projectRoot, "tools/scripts")

	require.NoError(t, err)
	require.Equal(t, scriptDirectory, resolved)
}

func TestResolveScriptDirectoryAcceptsForwardSlashProjectRoot(t *testing.T) {
	projectRoot := t.TempDir()
	scriptDirectory := filepath.Join(projectRoot, "Scripts")
	require.NoError(t, os.MkdirAll(scriptDirectory, 0o755))

	resolved, err := resolveScriptDirectory(filepath.ToSlash(projectRoot), "Scripts")

	require.NoError(t, err)
	require.Equal(t, scriptDirectory, resolved)
}

func TestTrackedScriptsByPathUsesAssetSchemaPaths(t *testing.T) {
	projectRoot := t.TempDir()
	scriptDirectory := filepath.Join(projectRoot, "Scripts")
	require.NoError(t, os.MkdirAll(scriptDirectory, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(scriptDirectory, "remap_shots.py"), []byte("print('ok')"), 0o644))
	db, err := sqlx.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE collection (id TEXT PRIMARY KEY, collection_path TEXT NOT NULL);
		CREATE TABLE asset (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			extension TEXT NOT NULL,
			pointer TEXT NOT NULL,
			collection_id TEXT NOT NULL,
			trashed BOOLEAN NOT NULL
		);
		INSERT INTO collection (id, collection_path) VALUES ('scripts', '/Scripts/');
		INSERT INTO asset (id, name, extension, pointer, collection_id, trashed)
		VALUES ('script-1', 'remap_shots', '.py', '', 'scripts', 0);
	`)
	require.NoError(t, err)
	tx, err := db.Beginx()
	require.NoError(t, err)
	defer tx.Rollback()

	tracked, err := trackedScriptsByPath(tx, projectRoot)

	require.NoError(t, err)
	expectedPath := normalizedFilePath(filepath.Join(projectRoot, "Scripts", "remap_shots.py"))
	require.Equal(t, trackedScript{ID: "script-1", Name: "remap_shots"}, tracked[expectedPath])
}
