package services

import (
	"path/filepath"
	"testing"
)

func TestPendingNamesInFolderIncludesMovedEntities(t *testing.T) {
	db := openMetadataTestDB(t)
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	root := t.TempDir()
	assetPath := filepath.Join(root, "OldAsset.psd")
	collectionPath := filepath.Join(root, "OldCollection")
	if _, err := tx.Exec(`INSERT INTO pending_path_update (entity_type, entity_id, current_local_path)
		VALUES ('asset', 'asset-id', ?), ('collection', 'collection-id', ?)`, assetPath, collectionPath); err != nil {
		t.Fatal(err)
	}

	folders, files, err := pendingNamesInFolder(tx, root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != "OldAsset.psd" {
		t.Fatalf("unexpected pending files: %v", files)
	}
	if len(folders) != 1 || folders[0] != "OldCollection" {
		t.Fatalf("unexpected pending folders: %v", folders)
	}
}
