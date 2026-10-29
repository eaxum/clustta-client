package services

import (
	"errors"
	"testing"

	"clustta/internal/auth_service"
	"clustta/internal/error_service"

	"github.com/jmoiron/sqlx"
)

func TestGetVisibleAssetByIDEnforcesUserScope(t *testing.T) {
	projectPath, workingDirectory := setupTrackAssetProject(t)
	filePath := createTrackAssetTestFile(t, workingDirectory, "private.blend")
	result, err := trackAsset(projectPath, TrackAssetOptions{FilePath: filePath}, nil)
	if err != nil {
		t.Fatal(err)
	}

	db, err := sqlx.Open("sqlite3", projectPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("UPDATE role SET name = 'artist', view_asset = 0 WHERE id = 'admin-role'"); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}

	assetService := &AssetService{}
	_, err = assetService.GetVisibleAssetByID(projectPath, result.Asset.Id)
	if !errors.Is(err, error_service.ErrNotUnauthorized) {
		t.Fatalf("expected hidden asset rejection, got %v", err)
	}

	db, err = sqlx.Open("sqlite3", projectPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(
		"UPDATE asset SET assignee_id = ?, is_resource = 0 WHERE id = ?",
		auth_service.OfflineUserID,
		result.Asset.Id,
	); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}

	visible, err := assetService.GetVisibleAssetByID(projectPath, result.Asset.Id)
	if err != nil {
		t.Fatal(err)
	}
	if visible.Id != result.Asset.Id {
		t.Fatalf("unexpected visible asset %q", visible.Id)
	}
}

func TestValidateRestoreFileStatusRequiresModifiedConfirmation(t *testing.T) {
	if err := validateRestoreFileStatus("modified", false); err == nil {
		t.Fatal("expected modified asset to require confirmation")
	}
	if err := validateRestoreFileStatus("modified", true); err != nil {
		t.Fatalf("expected confirmed overwrite to succeed, got %v", err)
	}
	if err := validateRestoreFileStatus("outdated", false); err != nil {
		t.Fatalf("expected outdated asset to restore without overwrite confirmation, got %v", err)
	}
}
