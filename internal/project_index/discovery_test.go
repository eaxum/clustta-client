package projectindex

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestInspectProjectReadsIndexMetadata(t *testing.T) {
	root := t.TempDir()
	workingDirectory := filepath.Join(root, "working")
	if err := os.MkdirAll(workingDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	databasePath := filepath.Join(root, "project.clst")
	createProjectIndexFixture(t, databasePath, workingDirectory)

	entry, err := inspectProject(databasePath)
	if err != nil {
		t.Fatal(err)
	}
	if entry.ProjectID != "project-id" || entry.ProjectName != "project" {
		t.Fatalf("unexpected project identity: %#v", entry)
	}
	if entry.StudioName != "Test Studio" {
		t.Fatalf("unexpected Studio: %q", entry.StudioName)
	}
	if entry.WorkingDirectory != workingDirectory {
		t.Fatalf("unexpected working directory: %q", entry.WorkingDirectory)
	}
}

func createProjectIndexFixture(
	t *testing.T,
	databasePath string,
	workingDirectory string,
) {
	t.Helper()
	database, err := sqlx.Connect("sqlite3", databasePath)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	_, err = database.Exec(`
		CREATE TABLE config (name TEXT PRIMARY KEY, value TEXT);
		INSERT INTO config (name, value) VALUES
			('project_id', 'project-id'),
			('project_name', 'project'),
			('studio_name', 'Test Studio'),
			('working_dir', ?);
	`, workingDirectory)
	if err != nil {
		t.Fatal(err)
	}
}
