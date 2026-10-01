package repository

import (
	"clustta/internal/compatibility"
	"clustta/internal/repository/migrations"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
)

func TestProjectSchemaMatchesMigrationTarget(t *testing.T) {
	if compatibility.CurrentProjectSchema != migrations.LatestVersion {
		t.Fatal("project schema does not match migration target")
	}
}

func TestProjectScanMigratesReplicaWithoutClearingPendingWork(t *testing.T) {
	path := filepath.Join(t.TempDir(), "replica.clst")
	db := sqlx.MustOpen("sqlite3", path)
	defer db.Close()
	db.MustExec(ProjectSchema)
	db.MustExec(`INSERT OR REPLACE INTO config (name, value, mtime, synced)
		VALUES ('version', '2.1', 1, 1), ('remote', 'https://studio.test/project', 1, 1), ('sync_token', 'pending', 1, 0)`)

	if err := UpdateProject(path); err != nil {
		t.Fatal(err)
	}

	var version, token string
	var synced bool
	if err := db.Get(&version, "SELECT value FROM config WHERE name = 'version'"); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&token, "SELECT value FROM config WHERE name = 'sync_token'"); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&synced, "SELECT synced FROM config WHERE name = 'sync_token'"); err != nil {
		t.Fatal(err)
	}
	if version != migrations.LatestVersion || token != "pending" || synced {
		t.Fatalf("unexpected migrated state: version=%s token=%s synced=%v", version, token, synced)
	}
}
