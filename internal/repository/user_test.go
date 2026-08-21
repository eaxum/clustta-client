package repository

import (
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
)

func TestAddKnownUserClearsExistingUserTomb(t *testing.T) {
	db, err := sqlx.Open("sqlite3", filepath.Join(t.TempDir(), "project.clst"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(ProjectSchema); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO role(id, mtime, name, synced) VALUES('artist-role', 1, 'Artist', 1)"); err != nil {
		t.Fatal(err)
	}

	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	const userID = "collaborator-user"
	if _, err = AddKnownUser(tx, userID, "artist@example.com", "artist", "Project", "Artist", "artist-role", nil, false); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec("DELETE FROM user WHERE id = ?", userID); err != nil {
		t.Fatal(err)
	}
	if _, err = AddKnownUser(tx, userID, "artist@example.com", "artist", "Project", "Artist", "artist-role", nil, false); err != nil {
		t.Fatal(err)
	}

	var tombCount int
	if err = tx.Get(&tombCount, "SELECT COUNT(*) FROM tomb WHERE id = ? AND table_name = 'user'", userID); err != nil {
		t.Fatal(err)
	}
	if tombCount != 0 {
		t.Fatalf("expected user tomb to be cleared, found %d", tombCount)
	}
}
