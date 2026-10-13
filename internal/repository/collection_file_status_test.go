package repository

import (
	"os"
	"path/filepath"
	"testing"

	"clustta/internal/utils"
)

func TestGetCollectionAssetsReturnsAuthoritativeFileStatuses(t *testing.T) {
	_, tx := openDependencyTestDB(t)
	workingDirectory := t.TempDir()
	currentFilePath := filepath.Join(workingDirectory, "current.py")
	if err := os.WriteFile(currentFilePath, []byte("current"), 0o600); err != nil {
		t.Fatal(err)
	}
	currentHash, err := utils.GenerateXXHashChecksum(currentFilePath)
	if err != nil {
		t.Fatal(err)
	}

	tx.MustExec("INSERT INTO config (name, value, mtime) VALUES ('working_dir', ?, 1)", workingDirectory)
	tx.MustExec("INSERT INTO status (id, mtime, name, short_name) VALUES ('todo', 1, 'Todo', 'todo')")
	tx.MustExec("INSERT INTO asset_type (id, mtime, name, icon) VALUES ('script', 1, 'Script', 'script-icon')")
	tx.MustExec(`
		INSERT INTO asset (
			id, created_at, mtime, name, extension, status_id, asset_type_id
		) VALUES
			('current', 1, 1, 'current', '.py', 'todo', 'script'),
			('remote', 1, 1, 'remote', '.py', 'todo', 'script')
	`)
	tx.MustExec(`
		INSERT INTO asset_checkpoint (
			id, created_at, mtime, asset_id, xxhash_checksum, time_modified,
			file_size, chunks, author_id, group_id
		) VALUES
			('current-checkpoint', 2, 1, 'current', ?, 1, 7, '', 'author', 'group'),
			('remote-checkpoint', 2, 1, 'remote', 'remote-hash', 1, 7, '', 'author', 'group')
	`, currentHash)

	assets, err := GetCollectionAssets(tx, "")
	if err != nil {
		t.Fatal(err)
	}

	statuses := make(map[string]string, len(assets))
	for _, asset := range assets {
		statuses[asset.Id] = asset.FileStatus
	}
	if statuses["current"] != "normal" {
		t.Fatalf("expected current asset to be normal, got %q", statuses["current"])
	}
	if statuses["remote"] != "fetchable" {
		t.Fatalf("expected remote asset to be fetchable, got %q", statuses["remote"])
	}
}
