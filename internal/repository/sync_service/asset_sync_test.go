package sync_service

import (
	"path/filepath"
	"testing"

	"clustta/internal/repository"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestNestedAssetSyncIncludesAncestorCollections(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "project.clst")
	db := sqlx.MustOpen("sqlite3", projectPath)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	db.MustExec(repository.ProjectSchema)
	db.MustExec(`
		INSERT INTO collection_type (id, mtime, name, icon, synced)
		VALUES ('generic-collection', 1, 'generic', 'folder', 1);
		INSERT INTO asset_type (id, mtime, name, icon, synced)
		VALUES ('generic-asset', 1, 'generic', 'file', 1);
		INSERT INTO status (id, mtime, name, short_name, color, synced)
		VALUES ('todo-status', 1, 'todo', 'todo', '#cccccc', 1);
		INSERT INTO collection (
			id, created_at, mtime, name, description,
			collection_type_id, parent_id, synced
		) VALUES
			('assets', '2026-10-29', 1, 'Assets', '', 'generic-collection', '', 1),
			('character', '2026-10-29', 2, 'character', '', 'generic-collection', 'assets', 0),
			('billy', '2026-10-29', 3, 'billy', '', 'generic-collection', 'character', 0),
			('publish', '2026-10-29', 4, 'publish', '', 'generic-collection', 'billy', 0);
		INSERT INTO asset (
			id, created_at, mtime, name, extension, asset_type_id,
			status_id, collection_id, synced
		) VALUES (
			'billy-blend', '2026-10-29', 5, 'billy', '.blend',
			'generic-asset', 'todo-status', 'publish', 0
		);
	`)

	tx := db.MustBegin()
	defer tx.Rollback()

	var pendingCollections int
	require.NoError(t, tx.Get(
		&pendingCollections,
		"SELECT COUNT(*) FROM collection WHERE synced = 0",
	))
	t.Logf("before sync: assets=1 unsynced collections=%d", pendingCollections)
	t.Log("legacy asset-only payload: assets=[billy-blend] collections=[]")

	data, err := LoadAssetData(tx, "billy-blend")
	require.NoError(t, err)
	t.Logf(
		"outbound payload: assets=%v collections=%v",
		assetIDs(data),
		collectionIDs(data),
	)
	require.Equal(t, []string{"character", "billy", "publish"}, collectionIDs(data))

	require.NoError(t, markAssetDataSynced(tx, data))
	var unsyncedCollections int
	require.NoError(t, tx.Get(
		&unsyncedCollections,
		"SELECT COUNT(*) FROM collection WHERE synced = 0",
	))
	t.Logf("unsynced collections after simulated success: %d", unsyncedCollections)
	require.Zero(t, unsyncedCollections)
}

func assetIDs(data ProjectData) []string {
	ids := make([]string, 0, len(data.Assets))
	for _, asset := range data.Assets {
		ids = append(ids, asset.Id)
	}
	return ids
}

func collectionIDs(data ProjectData) []string {
	ids := make([]string, 0, len(data.Collections))
	for _, collection := range data.Collections {
		ids = append(ids, collection.Id)
	}
	return ids
}
