package repository

import (
	"os"
	"path/filepath"
	"testing"

	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestPendingCollectionPathUpdateRenamesRecursiveChildren(t *testing.T) {
	db, err := sqlx.Open("sqlite3", filepath.Join(t.TempDir(), "project.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(ProjectSchema); err != nil {
		t.Fatal(err)
	}

	workingDirectory := filepath.Join(t.TempDir(), "working")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if err := utils.SetProjectWorkingDir(tx, workingDirectory); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec("INSERT INTO collection_type (id, mtime, name, icon) VALUES ('ct', 1, 'Folder', 'folder')"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec("INSERT INTO asset_type (id, mtime, name, icon) VALUES ('at', 1, 'File', 'file')"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec("INSERT INTO status (id, mtime, name, short_name, color) VALUES ('status', 1, 'Ready', 'RDY', '#fff')"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO collection
		(id, created_at, mtime, name, collection_type_id, parent_id)
		VALUES ('parent', 1, 1, 'OldParent', 'ct', ''), ('child', 1, 1, 'OldChild', 'ct', 'parent')`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO asset
		(id, created_at, mtime, name, extension, status_id, asset_type_id, collection_id)
		VALUES ('root-asset', 1, 1, 'RootAsset', '.txt', 'status', 'at', 'parent'),
		('child-asset', 1, 1, 'OldAsset', '.txt', 'status', 'at', 'child')`); err != nil {
		t.Fatal(err)
	}

	oldParentPath, err := utils.BuildCollectionPath(workingDirectory, "/OldParent/")
	if err != nil {
		t.Fatal(err)
	}
	oldChildPath, err := utils.BuildCollectionPath(workingDirectory, "/OldParent/OldChild/")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(oldChildPath, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	rootAssetPath, err := utils.BuildAssetPath(workingDirectory, "/OldParent/", "RootAsset", ".txt")
	if err != nil {
		t.Fatal(err)
	}
	childAssetPath, err := utils.BuildAssetPath(workingDirectory, "/OldParent/OldChild/", "OldAsset", ".txt")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rootAssetPath, []byte("root"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(childAssetPath, []byte("child"), 0o600); err != nil {
		t.Fatal(err)
	}

	incomingCollections := []models.Collection{
		{Id: "parent", MTime: 2, Name: "NewParent", ParentId: ""},
		{Id: "child", MTime: 2, Name: "NewChild", ParentId: "parent"},
	}
	incomingAssets := []models.Asset{
		{Id: "root-asset", MTime: 2, Name: "RootAsset", Extension: ".txt", CollectionId: "parent"},
		{Id: "child-asset", MTime: 2, Name: "NewAsset", Extension: ".txt", CollectionId: "child"},
	}
	if err := RecordPendingPathUpdates(tx, incomingCollections, incomingAssets, true); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec("UPDATE collection SET name = 'NewParent', mtime = 2 WHERE id = 'parent'"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec("UPDATE collection SET name = 'NewChild', mtime = 2 WHERE id = 'child'"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec("UPDATE asset SET name = 'NewAsset', mtime = 2 WHERE id = 'child-asset'"); err != nil {
		t.Fatal(err)
	}
	if err := ReconcilePendingPathUpdates(tx); err != nil {
		t.Fatal(err)
	}

	bindingCount := 0
	if err := tx.Get(&bindingCount, "SELECT COUNT(*) FROM pending_path_update"); err != nil {
		t.Fatal(err)
	}
	if bindingCount != 4 {
		t.Fatalf("expected 4 pending paths, got %d", bindingCount)
	}
	if err := ApplyCollectionPathUpdate(tx, "child"); err != nil {
		t.Fatal(err)
	}

	newRootAssetPath, err := utils.BuildAssetPath(workingDirectory, "/NewParent/", "RootAsset", ".txt")
	if err != nil {
		t.Fatal(err)
	}
	newChildAssetPath, err := utils.BuildAssetPath(workingDirectory, "/NewParent/NewChild/", "NewAsset", ".txt")
	if err != nil {
		t.Fatal(err)
	}
	if !utils.FileExists(newRootAssetPath) || !utils.FileExists(newChildAssetPath) {
		t.Fatal("recursive path update did not move all assets")
	}
	if utils.DirExists(oldParentPath) {
		t.Fatal("old collection path still exists")
	}
	if err := tx.Get(&bindingCount, "SELECT COUNT(*) FROM pending_path_update"); err != nil {
		t.Fatal(err)
	}
	if bindingCount != 0 {
		t.Fatalf("expected pending paths to be cleared, got %d", bindingCount)
	}
}
