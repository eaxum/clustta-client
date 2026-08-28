package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareDirectoryCreatesNestedDirectory(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "nested", "projects")

	if err := prepareDirectory(directory); err != nil {
		t.Fatalf("prepareDirectory() error = %v", err)
	}
	if info, err := os.Stat(directory); err != nil || !info.IsDir() {
		t.Fatalf("prepareDirectory() did not create %s", directory)
	}

	if err := prepareDirectory(directory); err != nil {
		t.Fatalf("prepareDirectory() should be idempotent, error = %v", err)
	}
}

func TestPrepareDirectoryRejectsEmptyPath(t *testing.T) {
	if err := prepareDirectory("  "); err == nil {
		t.Fatal("prepareDirectory() expected an error for an empty path")
	}
}
