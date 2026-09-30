package utils

import (
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestResolveProjectRemoteURLDoesNotRequireSchemaCompatibility(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "legacy.clst")
	db, err := OpenDb(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
		CREATE TABLE config (name TEXT PRIMARY KEY, value TEXT);
		INSERT INTO config (name, value) VALUES ('remote', 'https://studio.example/project');
		INSERT INTO config (name, value) VALUES ('version', '2');
	`)
	db.Close()
	if err != nil {
		t.Fatal(err)
	}

	remoteURL, err := ResolveProjectRemoteURL(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	if remoteURL != "https://studio.example/project" {
		t.Fatalf("unexpected remote URL %q", remoteURL)
	}
}
