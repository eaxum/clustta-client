package services

import (
	"os"
	"path/filepath"
	"testing"

	"clustta/internal/auth_service"
	"clustta/internal/repository"

	"github.com/jmoiron/sqlx"
	"github.com/zalando/go-keyring"
)

func TestTrackAssetCreatesAssignedAssetAndCheckpoint(t *testing.T) {
	projectPath, workingDirectory := setupTrackAssetProject(t)
	filePath := filepath.Join(workingDirectory, "characters", "hero.blend")
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filePath, []byte("blend-data"), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := trackAsset(projectPath, TrackAssetOptions{FilePath: filePath}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Created {
		t.Fatal("expected the asset to be created")
	}
	if result.Asset.AssigneeId != auth_service.OfflineUserID {
		t.Fatalf("expected asset assignment to active user, got %q", result.Asset.AssigneeId)
	}
	if result.Asset.CollectionPath != "/characters/" {
		t.Fatalf("unexpected collection path %q", result.Asset.CollectionPath)
	}
	if result.Checkpoint.AssetId != result.Asset.Id || result.Checkpoint.Comment != initialCheckpointComment {
		t.Fatalf("unexpected initial checkpoint: %#v", result.Checkpoint)
	}

	repeated, err := trackAsset(projectPath, TrackAssetOptions{FilePath: filePath}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if repeated.Created || repeated.Asset.Id != result.Asset.Id || repeated.Checkpoint.Id != result.Checkpoint.Id {
		t.Fatalf("expected existing tracked asset, got %#v", repeated)
	}
}

func TestTrackAssetRejectsFileOutsideWorkingDirectory(t *testing.T) {
	projectPath, _ := setupTrackAssetProject(t)
	outsideDirectory := createTrackAssetTestDirectory(t)
	filePath := filepath.Join(outsideDirectory, "outside.blend")
	if err := os.WriteFile(filePath, []byte("blend-data"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := trackAsset(projectPath, TrackAssetOptions{FilePath: filePath}, nil)
	if err == nil || err.Error() != "asset file must be inside the project working directory" {
		t.Fatalf("expected outside-path rejection, got %v", err)
	}
}

func setupTrackAssetProject(t *testing.T) (string, string) {
	t.Helper()
	keyring.MockInit()
	t.Cleanup(keyring.MockInit)
	if err := auth_service.AddAccountToken(auth_service.NewOfflineAccountToken()); err != nil {
		t.Fatal(err)
	}

	rootDirectory := createTrackAssetTestDirectory(t)
	workingDirectory := filepath.Join(rootDirectory, "workspace")
	if err := os.MkdirAll(workingDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	projectPath := filepath.Join(rootDirectory, "project.clst")
	db, err := sqlx.Open("sqlite3", projectPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(repository.ProjectSchema); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`
		INSERT INTO role (
			id, mtime, name, create_collection, create_asset,
			create_checkpoint, assign_asset, synced
		) VALUES ('admin-role', 1, 'admin', 1, 1, 1, 1, 1);
		INSERT INTO user (
			id, mtime, added_at, first_name, last_name, username,
			email, role_id, synced
		) VALUES (
			'offline-user', 1, 1, 'Offline', 'User', 'offline',
			'offline@local', 'admin-role', 1
		);
		INSERT INTO collection_type (id, mtime, name, icon, synced)
		VALUES ('generic-collection', 1, 'generic', 'folder', 1);
		INSERT INTO asset_type (id, mtime, name, icon, synced)
		VALUES ('generic-asset', 1, 'generic', 'file', 1);
		INSERT INTO status (id, mtime, name, short_name, color, synced)
		VALUES ('todo-status', 1, 'todo', 'todo', '#cccccc', 1);
		INSERT INTO config (name, value, mtime, synced)
		VALUES ('working_dir', ?, 1, 1);
	`, workingDirectory); err != nil {
		t.Fatal(err)
	}
	return projectPath, workingDirectory
}

func createTrackAssetTestDirectory(t *testing.T) string {
	t.Helper()
	directory, err := os.MkdirTemp(".", ".track-asset-test-")
	if err != nil {
		t.Fatal(err)
	}
	directory, err = filepath.Abs(directory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(directory); err != nil {
			t.Error(err)
		}
	})
	return directory
}
