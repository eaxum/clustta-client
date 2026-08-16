package repository

import (
	"os"
	"path/filepath"
	"testing"

	"clustta/internal/repository/models"
)

func TestGetAssetFileStatusPreservesModifiedStatusDuringPendingRename(t *testing.T) {
	localPath := filepath.Join(t.TempDir(), "old-name.psd")
	if err := os.WriteFile(localPath, []byte("modified"), 0o600); err != nil {
		t.Fatal(err)
	}

	asset := models.Asset{
		FilePath:  filepath.Join(filepath.Dir(localPath), "new-name.psd"),
		LocalPath: localPath,
	}
	status, err := GetAssetFileStatus(&asset, nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != "modified" {
		t.Fatalf("expected modified status, got %q", status)
	}
}
