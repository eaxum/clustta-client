package repository

import (
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func openDependencyTestDB(t *testing.T) (*sqlx.DB, *sqlx.Tx) {
	t.Helper()
	db, err := sqlx.Open("sqlite3", filepath.Join(t.TempDir(), "project.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ProjectSchema); err != nil {
		db.Close()
		t.Fatal(err)
	}
	tx, err := db.Beginx()
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		tx.Rollback()
		db.Close()
	})
	return db, tx
}

func insertTestCheckpoint(t *testing.T, tx *sqlx.Tx, id, assetId, groupId string, createdAt int) {
	t.Helper()
	_, err := tx.Exec(`
		INSERT INTO asset_checkpoint (
			id, created_at, mtime, asset_id, xxhash_checksum, time_modified,
			file_size, chunks, author_id, group_id
		) VALUES (?, ?, 1, ?, ?, 1, 1, '', 'author', ?)
	`, id, createdAt, assetId, id, groupId)
	if err != nil {
		t.Fatal(err)
	}
}
