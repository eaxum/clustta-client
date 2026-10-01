package migrations

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestMigrateV2_3AddsProjectManagementPermissions(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", filepath.Join(t.TempDir(), "project.db"))
	defer db.Close()
	db.MustExec(`
		CREATE TABLE role (id TEXT PRIMARY KEY, name TEXT NOT NULL);
		INSERT INTO role (id, name) VALUES ('admin-role', 'Admin'), ('artist-role', 'artist');
	`)

	if err := MigrateV2_3(db, ""); err != nil {
		t.Fatal(err)
	}

	for _, permission := range projectManagementPermissions {
		var adminAllowed bool
		if err := db.Get(&adminAllowed, fmt.Sprintf("SELECT %s FROM role WHERE id = 'admin-role'", permission)); err != nil {
			t.Fatal(err)
		}
		if !adminAllowed {
			t.Fatalf("expected admin to receive %s", permission)
		}

		var artistAllowed bool
		if err := db.Get(&artistAllowed, fmt.Sprintf("SELECT %s FROM role WHERE id = 'artist-role'", permission)); err != nil {
			t.Fatal(err)
		}
		if artistAllowed {
			t.Fatalf("expected %s to default to false", permission)
		}
	}
}
