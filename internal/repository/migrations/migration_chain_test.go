package migrations

import (
	"path/filepath"
	"testing"

	"clustta/internal/settings"
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
)

func TestRunMigrationsFromV1_2(t *testing.T) {
	t.Setenv("USERPROFILE", t.TempDir())
	if err := settings.InitializeServer(); err != nil {
		t.Fatal(err)
	}

	db, err := utils.OpenDb(filepath.Join(t.TempDir(), "project.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	createV1_2Database(t, db)

	if err = RunMigrations(db, "1.2", checkpointTagMigrationSchema); err != nil {
		t.Fatal(err)
	}

	assertMigrationValue(t, db, "SELECT value FROM config WHERE name = 'version'", LatestVersion)
	assertMigrationValue(t, db, "SELECT name FROM asset WHERE id = 'asset-1'", "Asset")
	assertMigrationValue(t, db, "SELECT name FROM collection WHERE id = 'collection-1'", "Collection")
	assertMigrationValue(t, db, "SELECT resolution_mode FROM asset_dependency WHERE id = 'dependency-1'", "floating")

	var checkpointCount int
	if err = db.Get(&checkpointCount, "SELECT count(*) FROM asset_checkpoint WHERE id = 'checkpoint-1' AND group_id != ''"); err != nil {
		t.Fatal(err)
	}
	if checkpointCount != 1 {
		t.Fatal("checkpoint was not preserved and grouped")
	}

	if err = RunMigrations(db, LatestVersion, checkpointTagMigrationSchema); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}

func createV1_2Database(t *testing.T, db *sqlx.DB) {
	t.Helper()
	_, err := db.Exec(`
		CREATE TABLE config (name TEXT PRIMARY KEY, value TEXT, mtime INTEGER NOT NULL);
		CREATE TABLE tomb (id TEXT, mtime INTEGER, table_name TEXT, synced BOOLEAN DEFAULT 0 NOT NULL);
		CREATE TABLE role (
			id TEXT PRIMARY KEY, name TEXT NOT NULL,
			view_entity BOOLEAN DEFAULT 0 NOT NULL,
			create_entity BOOLEAN DEFAULT 0 NOT NULL,
			update_entity BOOLEAN DEFAULT 0 NOT NULL,
			delete_entity BOOLEAN DEFAULT 0 NOT NULL,
			view_task BOOLEAN DEFAULT 0 NOT NULL,
			create_task BOOLEAN DEFAULT 0 NOT NULL,
			update_task BOOLEAN DEFAULT 0 NOT NULL,
			delete_task BOOLEAN DEFAULT 0 NOT NULL,
			assign_task BOOLEAN DEFAULT 0 NOT NULL,
			unassign_task BOOLEAN DEFAULT 0 NOT NULL,
			set_done_task BOOLEAN DEFAULT 0 NOT NULL,
			set_retake_task BOOLEAN DEFAULT 0 NOT NULL,
			view_done_task BOOLEAN DEFAULT 0 NOT NULL
		);
		CREATE TABLE entity_type (id TEXT PRIMARY KEY, name TEXT, icon TEXT);
		CREATE TABLE task_type (id TEXT PRIMARY KEY, name TEXT, icon TEXT);
		CREATE TABLE entity (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, parent_id TEXT,
			entity_type_id TEXT NOT NULL
		);
		CREATE TABLE task (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, entity_id TEXT NOT NULL,
			task_type_id TEXT NOT NULL
		);
		CREATE TABLE entity_assignee (id TEXT PRIMARY KEY, entity_id TEXT NOT NULL);
		CREATE TABLE entity_dependency (id TEXT PRIMARY KEY, task_id TEXT NOT NULL);
		CREATE TABLE task_dependency (
			id TEXT PRIMARY KEY, mtime INTEGER NOT NULL, task_id TEXT NOT NULL,
			dependency_id TEXT NOT NULL, dependency_type_id TEXT NOT NULL,
			synced BOOLEAN DEFAULT 0 NOT NULL
		);
		CREATE TABLE task_tag (id TEXT PRIMARY KEY, task_id TEXT NOT NULL);
		CREATE TABLE task_checkpoint (
			id TEXT PRIMARY KEY, entity_id TEXT NOT NULL, created_at TEXT NOT NULL,
			comment TEXT DEFAULT '' NOT NULL, author_id TEXT NOT NULL
		);
		CREATE TABLE workflow_entity (id TEXT PRIMARY KEY, entity_type_id TEXT NOT NULL);
		CREATE TABLE workflow_task (id TEXT PRIMARY KEY, task_type_id TEXT NOT NULL);
		CREATE TABLE workflow_link (id TEXT PRIMARY KEY, entity_type_id TEXT NOT NULL);
		CREATE VIEW entity_hierarchy AS
		SELECT id, '/' || name || '/' AS entity_path FROM entity;

		INSERT INTO config (name, value, mtime) VALUES
			('version', '1.2', 1),
			('server_name', 'Test Studio', 1),
			('name', 'Test Project', 1);
		INSERT INTO role (id, name) VALUES ('admin-role', 'Admin');
		INSERT INTO entity_type (id, name, icon) VALUES ('collection-type', 'Collection', 'scene');
		INSERT INTO task_type (id, name, icon) VALUES ('asset-type', 'Asset', 'modeling');
		INSERT INTO entity (id, name, parent_id, entity_type_id)
		VALUES ('collection-1', 'Collection', '', 'collection-type');
		INSERT INTO task (id, name, entity_id, task_type_id)
		VALUES ('asset-1', 'Asset', 'collection-1', 'asset-type');
		INSERT INTO task_dependency (id, mtime, task_id, dependency_id, dependency_type_id)
		VALUES ('dependency-1', 1, 'asset-1', 'asset-2', 'dependency-type');
		INSERT INTO task_checkpoint (id, entity_id, created_at, comment, author_id)
		VALUES ('checkpoint-1', 'asset-1', '2026-01-01T00:00:00Z', 'Initial', 'author');
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func assertMigrationValue(t *testing.T, db *sqlx.DB, query, expected string) {
	t.Helper()
	var actual string
	if err := db.Get(&actual, query); err != nil {
		t.Fatal(err)
	}
	if actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}
