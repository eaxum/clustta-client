package handlers

import (
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
