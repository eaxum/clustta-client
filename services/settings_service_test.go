package services

import (
	"os"
	"path/filepath"
	"testing"

	"clustta/internal/agent"
	"clustta/internal/repository"
	"clustta/internal/utils"

	"github.com/stretchr/testify/require"
)

func TestValidatePreLaunchHookScriptsAllowsModifiedScript(t *testing.T) {
	projectPath, scriptID, scriptPath := createProjectWithTrackedScript(t)
	require.NoError(t, os.WriteFile(scriptPath, []byte("print('modified')"), 0o600))

	_, err := agent.ResolveTrackedScriptPaths(projectPath, []string{scriptID})
	require.ErrorContains(t, err, "must be current before launch, status is modified")

	err = validatePreLaunchHookScripts(projectPath, []repository.PreLaunchHook{{
		Name:           "Blender - Initialize Scene",
		Enabled:        true,
		Extensions:     []string{".blend"},
		ScriptAssetIDs: []string{scriptID},
		FailurePolicy:  repository.PreLaunchFailureBlock,
	}})

	require.NoError(t, err)
}

func createProjectWithTrackedScript(t *testing.T) (string, string, string) {
	t.Helper()
	root := t.TempDir()
	workingDirectory := filepath.Join(root, "working")
	scriptDirectory := filepath.Join(workingDirectory, repository.DefaultProjectScriptDirectory)
	require.NoError(t, os.MkdirAll(scriptDirectory, 0o755))
	projectPath := filepath.Join(root, "project.clst")
	db, err := utils.OpenDb(projectPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	_, err = db.Exec(repository.ProjectSchema)
	require.NoError(t, err)
	tx, err := db.Beginx()
	require.NoError(t, err)
	defer tx.Rollback()
	require.NoError(t, utils.SetProjectWorkingDir(tx, workingDirectory))
	_, err = tx.Exec(`
		INSERT INTO collection_type (id, mtime, name, icon) VALUES ('scripts-type', 1, 'Scripts', 'code-bracket');
		INSERT INTO asset_type (id, mtime, name, icon) VALUES ('script-type', 1, 'Script', 'script');
		INSERT INTO status (id, mtime, name, short_name) VALUES ('todo', 1, 'To Do', 'todo');
		INSERT INTO collection (
			id, created_at, mtime, name, collection_type_id, parent_id
		) VALUES ('scripts', '2026-01-01', 1, 'Scripts', 'scripts-type', '');
	`)
	require.NoError(t, err)

	scriptID := "blender-prelaunch"
	scriptPath := filepath.Join(scriptDirectory, "blender_prelaunch.py")
	require.NoError(t, os.WriteFile(scriptPath, []byte("print('current')"), 0o600))
	checksum, err := utils.GenerateXXHashChecksum(scriptPath)
	require.NoError(t, err)
	fileInfo, err := os.Stat(scriptPath)
	require.NoError(t, err)
	_, err = tx.Exec(`
		INSERT INTO asset (
			id, created_at, mtime, name, extension, status_id, asset_type_id, collection_id
		) VALUES (?, '2026-01-01', 1, 'blender_prelaunch', '.py', 'todo', 'script-type', 'scripts')
	`, scriptID)
	require.NoError(t, err)
	_, err = tx.Exec(`
		INSERT INTO asset_checkpoint (
			id, created_at, mtime, asset_id, xxhash_checksum, time_modified,
			file_size, chunks, author_id
		) VALUES ('checkpoint', '2026-01-01', 1, ?, ?, ?, ?, '', 'author')
	`, scriptID, checksum, fileInfo.ModTime().Unix(), fileInfo.Size())
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	return projectPath, scriptID, scriptPath
}
