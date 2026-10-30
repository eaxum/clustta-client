package handlers

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateStableFileSnapshotPreservesContentAndExtension(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "scene.blend")
	content := []byte("stable blend content")
	if err := os.WriteFile(sourcePath, content, 0o600); err != nil {
		t.Fatal(err)
	}

	snapshotPath, err := createStableFileSnapshot(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Remove(snapshotPath); err != nil && !os.IsNotExist(err) {
			t.Error(err)
		}
	})

	if filepath.Ext(snapshotPath) != ".blend" {
		t.Fatalf("unexpected snapshot extension %q", filepath.Ext(snapshotPath))
	}
	snapshotContent, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(snapshotContent) != string(content) {
		t.Fatalf("unexpected snapshot content %q", snapshotContent)
	}
}

func TestVerifyStableFileSnapshotRejectsChangedSource(t *testing.T) {
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "scene.blend")
	snapshotPath := filepath.Join(directory, "snapshot.blend")
	if err := os.WriteFile(sourcePath, []byte("version-one"), 0o600); err != nil {
		t.Fatal(err)
	}
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(snapshotPath, []byte("version-one"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(sourcePath, []byte("version-two"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.Chtimes(sourcePath, sourceInfo.ModTime(), sourceInfo.ModTime()); err != nil {
		t.Fatal(err)
	}

	err = verifyStableFileSnapshot(sourcePath, sourceInfo, snapshotPath)
	if !errors.Is(err, errSourceFileChanged) {
		t.Fatalf("expected changed-source error, got %v", err)
	}
}

func TestCreateStableFileSnapshotBacksUpAudacityWAL(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "session.aup4")
	db := createAudacityTestProject(t, sourcePath)
	defer db.Close()
	if _, err := db.Exec("INSERT INTO sampleblocks VALUES (42)"); err != nil {
		t.Fatal(err)
	}

	snapshotPath, err := createStableFileSnapshot(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Remove(snapshotPath); err != nil && !os.IsNotExist(err) {
			t.Error(err)
		}
	})

	snapshot, err := sql.Open("sqlite3", snapshotPath)
	if err != nil {
		t.Fatal(err)
	}
	var blockID int
	if err = snapshot.QueryRow("SELECT id FROM sampleblocks").Scan(&blockID); err != nil {
		snapshot.Close()
		t.Fatal(err)
	}
	if blockID != 42 {
		snapshot.Close()
		t.Fatalf("unexpected sample block %d", blockID)
	}
	if err = snapshot.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(snapshotPath + "-wal"); !os.IsNotExist(err) {
		t.Fatalf("snapshot should not require a WAL file: %v", err)
	}
}

func TestCreateStableFileSnapshotRejectsUnsavedAudacityProject(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "session.aup4")
	db := createAudacityTestProject(t, sourcePath)
	defer db.Close()
	if _, err := db.Exec("INSERT INTO autosave VALUES (1)"); err != nil {
		t.Fatal(err)
	}

	_, err := createStableFileSnapshot(sourcePath)
	if err == nil || err.Error() != "save the current Audacity project before creating a checkpoint" {
		t.Fatalf("expected unsaved-project error, got %v", err)
	}
}

func createAudacityTestProject(t *testing.T, filePath string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		t.Fatal(err)
	}
	statements := []string{
		"PRAGMA application_id = 1096107097",
		"PRAGMA journal_mode = WAL",
		"PRAGMA wal_autocheckpoint = 0",
		"CREATE TABLE autosave (id INTEGER)",
		"CREATE TABLE project (id INTEGER)",
		"CREATE TABLE project_history (id INTEGER)",
		"CREATE TABLE sampleblocks (id INTEGER)",
	}
	for _, statement := range statements {
		if _, err = db.Exec(statement); err != nil {
			db.Close()
			t.Fatal(err)
		}
	}
	return db
}
