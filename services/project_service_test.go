package services

import (
	"clustta/internal/utils"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveIgnoreListProjectURIUsesConfiguredRemote(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "project.clst")
	db, err := utils.OpenDb(projectPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	_, err = db.Exec("CREATE TABLE config (name TEXT PRIMARY KEY, value TEXT)")
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO config (name, value) VALUES ('remote', ?)", "https://studio.example/projects/demo")
	require.NoError(t, err)

	projectURI, err := resolveIgnoreListProjectURI(projectPath)

	require.NoError(t, err)
	require.Equal(t, "https://studio.example/projects/demo", projectURI)
}

func TestResolveIgnoreListProjectURIFallsBackToLocalPath(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "project.clst")
	db, err := utils.OpenDb(projectPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	_, err = db.Exec("CREATE TABLE config (name TEXT PRIMARY KEY, value TEXT)")
	require.NoError(t, err)

	projectURI, err := resolveIgnoreListProjectURI(projectPath)

	require.NoError(t, err)
	require.Equal(t, projectPath, projectURI)
}
