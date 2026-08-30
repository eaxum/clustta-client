package migrations

import (
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const checkpointTagMigrationSchema = `
CREATE TABLE IF NOT EXISTS config (name TEXT PRIMARY KEY, value TEXT NOT NULL, mtime INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS tag (
    id TEXT PRIMARY KEY, mtime INTEGER NOT NULL, name TEXT UNIQUE NOT NULL COLLATE NOCASE,
    synced BOOLEAN DEFAULT 0 NOT NULL
);
CREATE TABLE IF NOT EXISTS asset_dependency (
    id TEXT PRIMARY KEY, mtime INTEGER NOT NULL, asset_id TEXT NOT NULL,
    dependency_id TEXT NOT NULL, dependency_type_id TEXT NOT NULL,
    resolution_mode TEXT DEFAULT 'floating' NOT NULL, checkpoint_id TEXT NULL,
    asset_checkpoint_tag_id TEXT NULL, synced BOOLEAN DEFAULT 0 NOT NULL
);
CREATE TABLE IF NOT EXISTS asset_checkpoint (
    id TEXT PRIMARY KEY, created_at DATETIME NOT NULL, asset_id TEXT NOT NULL,
    group_id TEXT DEFAULT '' NOT NULL, trashed BOOLEAN DEFAULT 0 NOT NULL
);
CREATE TABLE IF NOT EXISTS asset_checkpoint_tag (
    id TEXT PRIMARY KEY, mtime INTEGER NOT NULL, asset_id TEXT NOT NULL,
    tag_id TEXT NOT NULL, checkpoint_id TEXT NOT NULL, synced BOOLEAN DEFAULT 0 NOT NULL,
    UNIQUE (asset_id, tag_id)
);
CREATE TRIGGER IF NOT EXISTS asset_dependency_selector_insert BEFORE INSERT ON asset_dependency
FOR EACH ROW WHEN NOT (
    (NEW.resolution_mode = 'floating' AND NEW.checkpoint_id IS NULL AND NEW.asset_checkpoint_tag_id IS NULL) OR
    (NEW.resolution_mode = 'pinned' AND NEW.checkpoint_id IS NOT NULL AND NEW.asset_checkpoint_tag_id IS NULL) OR
    (NEW.resolution_mode = 'tagged' AND NEW.checkpoint_id IS NULL AND NEW.asset_checkpoint_tag_id IS NOT NULL)
) BEGIN SELECT RAISE(ABORT, 'invalid dependency selector'); END;
CREATE TRIGGER IF NOT EXISTS asset_dependency_selector_update BEFORE UPDATE OF resolution_mode, checkpoint_id, asset_checkpoint_tag_id ON asset_dependency
FOR EACH ROW WHEN NOT (
    (NEW.resolution_mode = 'floating' AND NEW.checkpoint_id IS NULL AND NEW.asset_checkpoint_tag_id IS NULL) OR
    (NEW.resolution_mode = 'pinned' AND NEW.checkpoint_id IS NOT NULL AND NEW.asset_checkpoint_tag_id IS NULL) OR
    (NEW.resolution_mode = 'tagged' AND NEW.checkpoint_id IS NULL AND NEW.asset_checkpoint_tag_id IS NOT NULL)
) BEGIN SELECT RAISE(ABORT, 'invalid dependency selector'); END;
DROP VIEW IF EXISTS asset_dependencies;
CREATE VIEW asset_dependencies AS SELECT asset_id, checkpoint_id, asset_checkpoint_tag_id FROM asset_dependency;
DROP VIEW IF EXISTS full_asset;
CREATE VIEW full_asset AS SELECT asset_id FROM asset_dependencies;`

func TestMigrateV2_2AddsVersionedDependencyAndCheckpointTagSchema(t *testing.T) {
	db, err := sqlx.Open("sqlite3", filepath.Join(t.TempDir(), "project.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(`
		CREATE TABLE config (name TEXT PRIMARY KEY, value TEXT NOT NULL, mtime INTEGER NOT NULL);
		CREATE TABLE asset_dependency (
			id TEXT PRIMARY KEY, mtime INTEGER NOT NULL, asset_id TEXT NOT NULL,
			dependency_id TEXT NOT NULL, dependency_type_id TEXT NOT NULL,
			synced BOOLEAN DEFAULT 0 NOT NULL
		);
		CREATE TABLE asset_checkpoint (
			id TEXT PRIMARY KEY, created_at DATETIME NOT NULL, asset_id TEXT NOT NULL,
			group_id TEXT DEFAULT '' NOT NULL, trashed BOOLEAN DEFAULT 0 NOT NULL
		);
		CREATE VIEW asset_dependencies AS SELECT asset_id FROM asset_dependency;
		CREATE VIEW full_asset AS SELECT asset_id FROM asset_dependencies;
		INSERT INTO config (name, value, mtime) VALUES ('version', '2.1', 1);
		INSERT INTO asset_dependency (id, mtime, asset_id, dependency_id, dependency_type_id)
		VALUES ('dependency-edge', 1, 'asset-1', 'asset-2', 'dependency-type');
		INSERT INTO asset_checkpoint (id, created_at, asset_id, group_id)
		VALUES ('checkpoint', 1, 'asset-1', 'existing-group');
	`); err != nil {
		t.Fatal(err)
	}

	if err = RunMigrations(db, 2.1, checkpointTagMigrationSchema); err != nil {
		t.Fatal(err)
	}

	var resolutionMode string
	if err = db.Get(&resolutionMode, "SELECT resolution_mode FROM asset_dependency WHERE id = 'dependency-edge'"); err != nil {
		t.Fatal(err)
	}
	if resolutionMode != "floating" {
		t.Fatalf("expected floating dependency, got %s", resolutionMode)
	}

	var assignmentColumnCount int
	if err = db.Get(&assignmentColumnCount, `
		SELECT COUNT(*) FROM pragma_table_info('asset_dependency')
		WHERE name = 'asset_checkpoint_tag_id'
	`); err != nil {
		t.Fatal(err)
	}
	if assignmentColumnCount != 1 {
		t.Fatal("expected asset_checkpoint_tag_id to be added")
	}

	var assignmentTableCount int
	if err = db.Get(&assignmentTableCount, `
		SELECT COUNT(*) FROM sqlite_master
		WHERE type = 'table' AND name = 'asset_checkpoint_tag'
	`); err != nil {
		t.Fatal(err)
	}
	if assignmentTableCount != 1 {
		t.Fatal("expected asset_checkpoint_tag table to exist")
	}

	var checkpointGroupTableCount int
	if err = db.Get(&checkpointGroupTableCount, `
		SELECT COUNT(*) FROM sqlite_master
		WHERE type = 'table' AND name = 'checkpoint_group'
	`); err != nil {
		t.Fatal(err)
	}
	if checkpointGroupTableCount != 0 {
		t.Fatal("checkpoint_group should not be created")
	}

	var groupId string
	if err = db.Get(&groupId, "SELECT group_id FROM asset_checkpoint WHERE id = 'checkpoint'"); err != nil {
		t.Fatal(err)
	}
	if groupId != "existing-group" {
		t.Fatalf("expected group ID to remain unchanged, got %s", groupId)
	}

	var projectVersion string
	if err = db.Get(&projectVersion, "SELECT value FROM config WHERE name = 'version'"); err != nil {
		t.Fatal(err)
	}
	if projectVersion != "2.2" {
		t.Fatalf("expected project version 2.2, got %s", projectVersion)
	}
}
