package repository

import (
	"clustta/internal/repository"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestClearSyncedChunkCache(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "project.clst")
	db, err := sqlx.Open("sqlite3", projectPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE chunk (hash TEXT PRIMARY KEY);
		CREATE TABLE asset_checkpoint (chunks TEXT, synced INTEGER);
		CREATE TABLE template (chunks TEXT, synced INTEGER);
		INSERT INTO chunk (hash) VALUES ('cached'), ('checkpoint'), ('template');
		INSERT INTO asset_checkpoint (chunks, synced) VALUES ('checkpoint', 0);
		INSERT INTO template (chunks, synced) VALUES ('template', 0);
	`)
	if err != nil {
		t.Fatal(err)
	}

	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	if err = repository.ClearSyncedChunkCache(tx); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}

	var hashes []string
	if err = db.Select(&hashes, "SELECT hash FROM chunk ORDER BY hash"); err != nil {
		t.Fatal(err)
	}
	if len(hashes) != 2 || hashes[0] != "checkpoint" || hashes[1] != "template" {
		t.Fatalf("unexpected retained chunks: %v", hashes)
	}
}

func TestClearChunkCache(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "project.clst")
	db, err := sqlx.Open("sqlite3", projectPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err = db.Exec("CREATE TABLE chunk (hash TEXT PRIMARY KEY); INSERT INTO chunk (hash) VALUES ('one'), ('two')"); err != nil {
		t.Fatal(err)
	}
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	if err = repository.ClearChunkCache(tx); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}

	var count int
	if err = db.Get(&count, "SELECT COUNT(*) FROM chunk"); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected empty chunk cache, got %d rows", count)
	}
}

func TestShouldVacuum(t *testing.T) {
	const megabyte = int64(1 << 20)
	tests := []struct {
		name      string
		fileSize  int64
		freeSpace int64
		expected  bool
	}{
		{name: "below minimum space", fileSize: 500 * megabyte, freeSpace: 99 * megabyte, expected: false},
		{name: "below minimum percentage", fileSize: 1000 * megabyte, freeSpace: 100 * megabyte, expected: false},
		{name: "meets both thresholds", fileSize: 500 * megabyte, freeSpace: 100 * megabyte, expected: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := repository.ShouldVacuum(test.fileSize, 4096, test.freeSpace/4096); actual != test.expected {
				t.Fatalf("expected %t, got %t", test.expected, actual)
			}
		})
	}
}
