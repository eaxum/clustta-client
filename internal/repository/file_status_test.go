package repository

import (
	"os"
	"path/filepath"
	"testing"

	"clustta/internal/repository/models"
)

func TestGetAssetFileStatusCachesUnchangedFile(t *testing.T) {
	filePath := writeStatusTestFile(t, "asset.blend", "content")
	asset := models.Asset{FilePath: filePath}
	checkpoints := []models.Checkpoint{{Id: "checkpoint", XXHashChecksum: "hash"}}
	hashCalls := 0
	hasher := func(string) (string, error) {
		hashCalls++
		return "hash", nil
	}

	for range 2 {
		status, err := getAssetFileStatus(&asset, checkpoints, hasher)
		if err != nil {
			t.Fatal(err)
		}
		if status != "normal" {
			t.Fatalf("expected normal status, got %q", status)
		}
	}
	if hashCalls != 1 {
		t.Fatalf("expected one hash calculation, got %d", hashCalls)
	}
}

func TestGetAssetFileStatusInvalidatesChangedFile(t *testing.T) {
	filePath := writeStatusTestFile(t, "asset.blend", "first")
	asset := models.Asset{FilePath: filePath}
	checkpoints := []models.Checkpoint{{Id: "checkpoint", XXHashChecksum: "hash"}}
	hashCalls := 0
	hasher := func(string) (string, error) {
		hashCalls++
		return "hash", nil
	}

	if _, err := getAssetFileStatus(&asset, checkpoints, hasher); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filePath, []byte("changed-content"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := getAssetFileStatus(&asset, checkpoints, hasher); err != nil {
		t.Fatal(err)
	}
	if hashCalls != 2 {
		t.Fatalf("expected changed file to be rehashed, got %d calls", hashCalls)
	}
}

func TestGetAssetFileStatusInvalidatesChangedCheckpoints(t *testing.T) {
	filePath := writeStatusTestFile(t, "asset.blend", "content")
	asset := models.Asset{FilePath: filePath}
	hashCalls := 0
	hasher := func(string) (string, error) {
		hashCalls++
		return "second-hash", nil
	}

	first := []models.Checkpoint{{Id: "first", XXHashChecksum: "first-hash"}}
	status, err := getAssetFileStatus(&asset, first, hasher)
	if err != nil {
		t.Fatal(err)
	}
	if status != "modified" {
		t.Fatalf("expected modified status, got %q", status)
	}

	second := []models.Checkpoint{{Id: "second", XXHashChecksum: "second-hash"}}
	status, err = getAssetFileStatus(&asset, second, hasher)
	if err != nil {
		t.Fatal(err)
	}
	if status != "normal" {
		t.Fatalf("expected normal status, got %q", status)
	}
	if hashCalls != 2 {
		t.Fatalf("expected checkpoint change to rehash, got %d calls", hashCalls)
	}
}

func writeStatusTestFile(t *testing.T, name string, content string) string {
	t.Helper()
	filePath := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return filePath
}
