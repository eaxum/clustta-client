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

func TestPreserveLocalVersionedDependencyChanges(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	t.Cleanup(func() { db.Close() })
	db.MustExec(`
		CREATE TABLE asset_checkpoint_tag (id TEXT PRIMARY KEY, synced BOOLEAN);
		CREATE TABLE tomb (id TEXT, table_name TEXT, synced BOOLEAN);
		CREATE TABLE asset_dependency (
			id TEXT PRIMARY KEY,
			resolution_mode TEXT,
			checkpoint_id TEXT,
			asset_checkpoint_tag_id TEXT,
			synced BOOLEAN
		);
		INSERT INTO asset_checkpoint_tag (id, synced) VALUES ('assignment', 1);
		INSERT INTO tomb (id, table_name, synced) VALUES ('deleted-assignment', 'asset_checkpoint_tag', 1);
		INSERT INTO asset_dependency VALUES ('floating', 'floating', NULL, NULL, 1);
		INSERT INTO asset_dependency VALUES ('pinned', 'pinned', 'checkpoint', NULL, 1);
	`)
	tx := db.MustBegin()

	require.NoError(t, PreserveLocalVersionedDependencyChanges(tx))

	var assignmentSynced bool
	require.NoError(t, tx.Get(&assignmentSynced, "SELECT synced FROM asset_checkpoint_tag WHERE id = 'assignment'"))
	require.False(t, assignmentSynced)
	var assignmentTombSynced bool
	require.NoError(t, tx.Get(&assignmentTombSynced, "SELECT synced FROM tomb WHERE id = 'deleted-assignment'"))
	require.False(t, assignmentTombSynced)
	var floatingSynced bool
	require.NoError(t, tx.Get(&floatingSynced, "SELECT synced FROM asset_dependency WHERE id = 'floating'"))
	require.True(t, floatingSynced)
	var pinnedSynced bool
	require.NoError(t, tx.Get(&pinnedSynced, "SELECT synced FROM asset_dependency WHERE id = 'pinned'"))
	require.False(t, pinnedSynced)
}
