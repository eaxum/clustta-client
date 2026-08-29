package migrations

import (
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const checkpointGroupMigrationSchema = `
CREATE TABLE IF NOT EXISTS asset_checkpoint (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    asset_id TEXT NOT NULL,
    group_id TEXT DEFAULT '' NOT NULL,
    trashed BOOLEAN DEFAULT 0 NOT NULL
);
CREATE TABLE IF NOT EXISTS checkpoint_group (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    group_type TEXT NOT NULL CHECK (group_type IN ('single', 'multi')),
    finalized BOOLEAN DEFAULT 0 NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL
);`

func TestMigrateV2_2BackfillsSingleAndMultiGroups(t *testing.T) {
	db, err := sqlx.Open("sqlite3", filepath.Join(t.TempDir(), "project.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(checkpointGroupMigrationSchema); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
		INSERT INTO asset_checkpoint (id, created_at, asset_id, group_id) VALUES
		('cp-1', 1, 'asset-1', 'single-group'),
		('cp-2', 2, 'asset-1', 'multi-group'),
		('cp-3', 3, 'asset-2', 'multi-group'),
		('cp-4', 4, 'asset-3', '')
	`)
	if err != nil {
		t.Fatal(err)
	}

	if err = MigrateV2_2(db, checkpointGroupMigrationSchema); err != nil {
		t.Fatal(err)
	}

	groups := []struct {
		Id        string `db:"id"`
		GroupType string `db:"group_type"`
		Finalized bool   `db:"finalized"`
	}{}
	if err = db.Select(&groups, "SELECT id, group_type, finalized FROM checkpoint_group"); err != nil {
		t.Fatal(err)
	}
	if len(groups) != 3 {
		t.Fatalf("expected three backfilled groups, got %d", len(groups))
	}

	typesById := map[string]string{}
	for _, group := range groups {
		if !group.Finalized {
			t.Fatalf("expected backfilled group %s to be finalized", group.Id)
		}
		typesById[group.Id] = group.GroupType
	}
	if typesById["single-group"] != "single" || typesById["multi-group"] != "multi" {
		t.Fatalf("unexpected backfilled group types: %+v", typesById)
	}

	var emptyGroupIds int
	if err = db.Get(&emptyGroupIds, "SELECT COUNT(*) FROM asset_checkpoint WHERE group_id = ''"); err != nil {
		t.Fatal(err)
	}
	if emptyGroupIds != 0 {
		t.Fatalf("expected empty group IDs to be replaced, got %d", emptyGroupIds)
	}
}
