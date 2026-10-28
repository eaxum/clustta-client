package services

import (
	"path/filepath"
	"testing"

	"clustta/internal/repository"

	"github.com/jmoiron/sqlx"
)

func TestLinkedDependencyTypeID(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "dependency-type.db")
	db, err := sqlx.Open("sqlite3", projectPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.MustExec(repository.ProjectSchema)
	db.MustExec(`
		INSERT INTO dependency_type (id, mtime, name)
		VALUES ('blocking-id', 1, 'blocking'), ('linked-id', 1, 'linked')
	`)
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	dependencyTypeID, err := linkedDependencyTypeID(tx)
	if err != nil {
		t.Fatal(err)
	}
	if dependencyTypeID != "linked-id" {
		t.Fatalf("expected linked dependency type, got %q", dependencyTypeID)
	}
}
