package migrations

import (
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const checkpointGroupMigrationSchema = `
CREATE TABLE IF NOT EXISTS config (
    name TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    mtime INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS asset_dependency (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    asset_id TEXT NOT NULL,
    dependency_id TEXT NOT NULL,
    dependency_type_id TEXT NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL
);
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
);
CREATE TRIGGER IF NOT EXISTS asset_dependency_selector_insert BEFORE INSERT ON asset_dependency
FOR EACH ROW
WHEN NEW.resolution_mode = 'floating' AND NEW.checkpoint_id IS NULL
BEGIN
    SELECT RAISE(ABORT, 'invalid dependency selector');
END;
CREATE TRIGGER IF NOT EXISTS asset_dependency_selector_update BEFORE UPDATE OF resolution_mode, checkpoint_id ON asset_dependency
FOR EACH ROW
WHEN NEW.resolution_mode = 'floating' AND NEW.checkpoint_id IS NULL
BEGIN
    SELECT RAISE(ABORT, 'invalid dependency selector');
END;
DROP VIEW IF EXISTS asset_dependencies;
CREATE VIEW asset_dependencies AS
SELECT
    asset_id,
    checkpoint_id,
    checkpoint_group_tag_id
FROM asset_dependency;
DROP VIEW IF EXISTS full_asset;
CREATE VIEW full_asset AS
SELECT asset_id FROM asset_dependencies;`

func TestMigrateV2_2BackfillsSingleAndMultiGroups(t *testing.T) {
	db, err := sqlx.Open("sqlite3", filepath.Join(t.TempDir(), "project.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(checkpointGroupMigrationSchema); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`
		DROP TRIGGER asset_dependency_selector_insert;
		DROP TRIGGER asset_dependency_selector_update;
		INSERT INTO asset_dependency (
			id, mtime, asset_id, dependency_id, dependency_type_id
		) VALUES ('dependency-edge', 1, 'asset-1', 'asset-2', 'dependency-type');
	`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`
		CREATE TRIGGER asset_dependency_selector_insert BEFORE INSERT ON asset_dependency
		FOR EACH ROW
		WHEN NEW.resolution_mode = 'floating' AND NEW.checkpoint_id IS NULL
		BEGIN
			SELECT RAISE(ABORT, 'invalid dependency selector');
		END;
		CREATE TRIGGER asset_dependency_selector_update BEFORE UPDATE OF resolution_mode, checkpoint_id ON asset_dependency
		FOR EACH ROW
		WHEN NEW.resolution_mode = 'floating' AND NEW.checkpoint_id IS NULL
		BEGIN
			SELECT RAISE(ABORT, 'invalid dependency selector');
		END;
	`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO config (name, value, mtime) VALUES ('version', '2.1', 1)"); err != nil {
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

	if err = RunMigrations(db, 2.1, checkpointGroupMigrationSchema); err != nil {
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

	var resolutionMode string
	if err = db.Get(&resolutionMode, "SELECT resolution_mode FROM asset_dependency WHERE id = 'dependency-edge'"); err != nil {
		t.Fatal(err)
	}
	if resolutionMode != "floating" {
		t.Fatalf("expected existing dependency to migrate to floating, got %s", resolutionMode)
	}

	var projectVersion string
	if err = db.Get(&projectVersion, "SELECT value FROM config WHERE name = 'version'"); err != nil {
		t.Fatal(err)
	}
	if projectVersion != "2.2" {
		t.Fatalf("expected project version 2.2, got %s", projectVersion)
	}
}
