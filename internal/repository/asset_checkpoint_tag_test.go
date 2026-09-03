package repository

import "testing"

func TestCheckpointTagMovesWithinOneAssetAndMaintainsAssetTag(t *testing.T) {
	_, tx := openDependencyTestDB(t)
	insertDependencyAsset(t, tx, "boy")
	insertTestCheckpoint(t, tx, "boy-v1-cp", "boy", "boy-v1", 1)
	insertTestCheckpoint(t, tx, "boy-v2-cp", "boy", "boy-v2", 2)

	assignment, err := SetCheckpointTag(tx, "", "animation-approved", "boy-v1-cp")
	if err != nil {
		t.Fatal(err)
	}
	movedAssignment, err := SetCheckpointTag(tx, assignment.TagId, assignment.Name, "boy-v2-cp")
	if err != nil {
		t.Fatal(err)
	}
	if movedAssignment.Id != assignment.Id || movedAssignment.CheckpointId != "boy-v2-cp" {
		t.Fatalf("expected the existing assignment to move, got %+v", movedAssignment)
	}
	if movedAssignment.MTime <= assignment.MTime {
		t.Fatalf("expected moved assignment mtime to advance from %d, got %d", assignment.MTime, movedAssignment.MTime)
	}

	var assetTagCount int
	if err = tx.Get(&assetTagCount, `
		SELECT COUNT(*) FROM asset_tag WHERE asset_id = ? AND tag_id = ?
	`, "boy", assignment.TagId); err != nil {
		t.Fatal(err)
	}
	if assetTagCount != 1 {
		t.Fatalf("expected the checkpoint tag to add one asset tag, got %d", assetTagCount)
	}
	if _, err = tx.Exec("UPDATE asset_checkpoint SET trashed = 1 WHERE id = ?", movedAssignment.CheckpointId); err == nil {
		t.Fatal("expected a tagged checkpoint to reject trashing")
	}

	if err = DeleteCheckpointTag(tx, assignment.Id); err != nil {
		t.Fatal(err)
	}
	if err = tx.Get(&assetTagCount, `
		SELECT COUNT(*) FROM asset_tag WHERE asset_id = ? AND tag_id = ?
	`, "boy", assignment.TagId); err != nil {
		t.Fatal(err)
	}
	if assetTagCount != 0 {
		t.Fatalf("expected removing the checkpoint assignment to remove the asset tag, got %d", assetTagCount)
	}

	var tombCount int
	if err = tx.Get(&tombCount, `
		SELECT COUNT(*) FROM tomb WHERE id = ? AND table_name = 'asset_checkpoint_tag'
	`, assignment.Id); err != nil {
		t.Fatal(err)
	}
	if tombCount != 0 {
		t.Fatalf("expected an unsynced assignment removal to cancel its tombstone, got %d", tombCount)
	}
}

func TestDeleteSyncedCheckpointTagCreatesTombstones(t *testing.T) {
	_, tx := openDependencyTestDB(t)
	insertDependencyAsset(t, tx, "boy")
	insertTestCheckpoint(t, tx, "boy-cp", "boy", "boy-v1", 1)

	assignment, err := SetCheckpointTag(tx, "", "approved", "boy-cp")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec("UPDATE asset_checkpoint_tag SET synced = 1 WHERE id = ?", assignment.Id); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec("UPDATE asset_tag SET synced = 1 WHERE asset_id = ? AND tag_id = ?", assignment.AssetId, assignment.TagId); err != nil {
		t.Fatal(err)
	}

	if err = DeleteCheckpointTag(tx, assignment.Id); err != nil {
		t.Fatal(err)
	}
	var tombCount int
	if err = tx.Get(&tombCount, `
		SELECT COUNT(*) FROM tomb
		WHERE (id = ? AND table_name = 'asset_checkpoint_tag') OR table_name = 'asset_tag'
	`, assignment.Id); err != nil {
		t.Fatal(err)
	}
	if tombCount != 2 {
		t.Fatalf("expected checkpoint and asset tag tombstones, got %d", tombCount)
	}
}

func TestSetCheckpointTagsForGroupReusesProjectTag(t *testing.T) {
	_, tx := openDependencyTestDB(t)
	insertDependencyAsset(t, tx, "boy")
	insertDependencyAsset(t, tx, "prop")
	insertTestCheckpoint(t, tx, "boy-cp", "boy", "multi", 2)
	insertTestCheckpoint(t, tx, "prop-cp", "prop", "multi", 2)

	assignments, err := SetCheckpointTagsForGroup(tx, "", "approved", "multi")
	if err != nil {
		t.Fatal(err)
	}
	if len(assignments) != 2 {
		t.Fatalf("expected one assignment per asset, got %+v", assignments)
	}
	if assignments[0].TagId != assignments[1].TagId {
		t.Fatalf("expected both assets to reuse one project tag, got %+v", assignments)
	}

	boyTags, err := GetCheckpointTagsForAsset(tx, "boy")
	if err != nil {
		t.Fatal(err)
	}
	if len(boyTags) != 1 || boyTags[0].CheckpointId != "boy-cp" {
		t.Fatalf("unexpected boy checkpoint tag: %+v", boyTags)
	}
}
