package repository

import "testing"

func TestCheckpointTagMovesWithinOneAsset(t *testing.T) {
	_, tx := openCheckpointGroupTestDB(t)

	for _, groupId := range []string{"boy-v1", "boy-v2"} {
		if _, err := BeginCheckpointGroup(tx, groupId, CheckpointGroupTypeSingle); err != nil {
			t.Fatal(err)
		}
		insertCheckpointGroupMember(t, tx, groupId+"-cp", "boy", groupId, 1)
		if _, err := FinalizeCheckpointGroup(tx, groupId); err != nil {
			t.Fatal(err)
		}
	}

	tag, err := SetCheckpointTag(tx, "", "animation-approved", "boy-v1-cp")
	if err != nil {
		t.Fatal(err)
	}
	movedTag, err := SetCheckpointTag(tx, "", tag.Name, "boy-v2-cp")
	if err != nil {
		t.Fatal(err)
	}
	if movedTag.Id != tag.Id || movedTag.CheckpointId != "boy-v2-cp" {
		t.Fatalf("expected the existing tag to move, got %+v", movedTag)
	}
	if movedTag.MTime <= tag.MTime {
		t.Fatalf("expected moved tag mtime to advance from %d, got %d", tag.MTime, movedTag.MTime)
	}

	var tagCount int
	if err = tx.Get(&tagCount, "SELECT COUNT(*) FROM checkpoint_tag WHERE asset_id = ? AND name = ?", "boy", tag.Name); err != nil {
		t.Fatal(err)
	}
	if tagCount != 1 {
		t.Fatalf("expected one tag assignment in the asset history, got %d", tagCount)
	}
	if _, err = tx.Exec("UPDATE asset_checkpoint SET trashed = 1 WHERE id = ?", movedTag.CheckpointId); err == nil {
		t.Fatal("expected a tagged checkpoint to reject trashing")
	}

	if err = DeleteCheckpointTag(tx, tag.Id); err != nil {
		t.Fatal(err)
	}
	var tombCount int
	if err = tx.Get(&tombCount, "SELECT COUNT(*) FROM tomb WHERE id = ? AND table_name = 'checkpoint_tag'", tag.Id); err != nil {
		t.Fatal(err)
	}
	if tombCount != 1 {
		t.Fatalf("expected one checkpoint tag tombstone, got %d", tombCount)
	}
}

func TestSetCheckpointTagsForSingleAndMultiGroups(t *testing.T) {
	_, tx := openCheckpointGroupTestDB(t)

	if _, err := BeginCheckpointGroup(tx, "single", CheckpointGroupTypeSingle); err != nil {
		t.Fatal(err)
	}
	insertCheckpointGroupMember(t, tx, "single-cp", "boy", "single", 1)
	if _, err := FinalizeCheckpointGroup(tx, "single"); err != nil {
		t.Fatal(err)
	}
	singleTags, err := SetCheckpointTagsForGroup(tx, "approved", "single")
	if err != nil {
		t.Fatal(err)
	}
	if len(singleTags) != 1 || singleTags[0].CheckpointId != "single-cp" {
		t.Fatalf("unexpected single group tags: %+v", singleTags)
	}

	if _, err = BeginCheckpointGroup(tx, "multi", CheckpointGroupTypeMulti); err != nil {
		t.Fatal(err)
	}
	insertCheckpointGroupMember(t, tx, "boy-cp", "boy", "multi", 2)
	insertCheckpointGroupMember(t, tx, "prop-cp", "prop", "multi", 2)
	if _, err = FinalizeCheckpointGroup(tx, "multi"); err != nil {
		t.Fatal(err)
	}
	multiTags, err := SetCheckpointTagsForGroup(tx, "approved", "multi")
	if err != nil {
		t.Fatal(err)
	}
	if len(multiTags) != 2 {
		t.Fatalf("expected one tag per asset, got %+v", multiTags)
	}

	boyTags, err := GetCheckpointTagsForAsset(tx, "boy")
	if err != nil {
		t.Fatal(err)
	}
	if len(boyTags) != 1 || boyTags[0].CheckpointId != "boy-cp" {
		t.Fatalf("expected the boy tag to move to the multi checkpoint, got %+v", boyTags)
	}
}
