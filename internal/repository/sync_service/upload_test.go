package sync_service

import (
	"clustta/internal/repository"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestMarkAllTablesUnsyncedIncludesProjectConfigs(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	t.Cleanup(func() { db.Close() })
	db.MustExec(`CREATE TABLE config (name TEXT PRIMARY KEY, value TEXT, mtime INTEGER, synced BOOLEAN)`)
	db.MustExec(`INSERT INTO config (name, value, mtime, synced) VALUES
		('project_preview', '', 1, 1),
		('dcc_prelaunch_hooks', '{}', 1, 1),
		('project_script_settings', '{}', 1, 1),
		('unrelated_setting', '', 1, 1)`)
	tx := db.MustBegin()

	require.NoError(t, MarkAllTablesUnsynced(tx))

	for _, name := range append([]string{projectPreviewConfigName}, repository.SyncableProjectConfigNames...) {
		var synced bool
		require.NoError(t, tx.Get(&synced, "SELECT synced FROM config WHERE name = ?", name))
		require.False(t, synced)
	}
	var unrelatedSynced bool
	require.NoError(t, tx.Get(&unrelatedSynced, "SELECT synced FROM config WHERE name = 'unrelated_setting'"))
	require.True(t, unrelatedSynced)
}
