package repository

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func stringReference(value string) *string {
	return &value
}

func insertDependencyAsset(t *testing.T, tx *sqlx.Tx, id string) {
	t.Helper()
	_, err := tx.Exec(`
		INSERT INTO asset (
			id, created_at, mtime, name, extension, status_id, asset_type_id
		) VALUES (?, 1, 1, ?, '.blend', 'status', 'asset-type')
	`, id, id)
	if err != nil {
		t.Fatal(err)
	}
}

func insertDependencyType(t *testing.T, tx *sqlx.Tx) {
	t.Helper()
	if _, err := tx.Exec("INSERT INTO dependency_type (id, mtime, name) VALUES ('reference', 1, 'Reference')"); err != nil {
		t.Fatal(err)
	}
}

func TestDependencySelectorsResolveFloatingPinnedAndTagged(t *testing.T) {
	_, tx := openDependencyTestDB(t)
	insertDependencyType(t, tx)
	for _, assetId := range []string{"shot", "boy", "prop"} {
		insertDependencyAsset(t, tx, assetId)
	}

	insertTestCheckpoint(t, tx, "boy-cp-1", "boy", "boy-v1", 1)

	edge, err := AddDependencyWithSelector(
		tx,
		"edge",
		"shot",
		"boy",
		"reference",
		DependencyResolutionFloating,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if edge.ResolutionStatus != "ready" || edge.ResolvedCheckpointId == nil || *edge.ResolvedCheckpointId != "boy-cp-1" {
		t.Fatalf("unexpected floating edge: %+v", edge)
	}

	insertTestCheckpoint(t, tx, "boy-cp-2", "boy", "boy-v2", 2)
	edges, err := GetAssetDependencyEdges(tx, "shot")
	if err != nil {
		t.Fatal(err)
	}
	if edges[0].ResolvedCheckpointId == nil || *edges[0].ResolvedCheckpointId != "boy-cp-2" {
		t.Fatalf("expected floating edge to follow boy-cp-2, got %+v", edges[0])
	}

	edge, err = UpdateDependencySelector(
		tx,
		"edge",
		DependencyResolutionPinned,
		stringReference("boy-cp-1"),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if edge.ResolvedCheckpointId == nil || *edge.ResolvedCheckpointId != "boy-cp-1" {
		t.Fatalf("expected pinned edge to resolve boy-cp-1, got %+v", edge)
	}
	referenceCounts, err := GetCheckpointDependencyReferenceCounts(tx, "boy")
	if err != nil {
		t.Fatal(err)
	}
	if referenceCounts["boy-cp-1"] != 1 {
		t.Fatalf("expected one exact-pin follower, got %+v", referenceCounts)
	}

	insertTestCheckpoint(t, tx, "release-boy", "boy", "release", 3)
	insertTestCheckpoint(t, tx, "release-prop", "prop", "release", 4)
	tag, err := SetCheckpointTag(tx, "", "animation-approved", "release-boy")
	if err != nil {
		t.Fatal(err)
	}
	edge, err = UpdateDependencySelector(
		tx,
		"edge",
		DependencyResolutionTagged,
		nil,
		&tag.Id,
	)
	if err != nil {
		t.Fatal(err)
	}
	if edge.ResolvedCheckpointId == nil || *edge.ResolvedCheckpointId != "release-boy" || edge.TagName != tag.Name {
		t.Fatalf("unexpected tagged edge: %+v", edge)
	}
	insertTestCheckpoint(t, tx, "incompatible-prop", "prop", "incompatible-release", 5)
	insertTestCheckpoint(t, tx, "incompatible-other", "shot", "incompatible-release", 6)
	otherAssignment, err := SetCheckpointTag(tx, tag.TagId, tag.Name, "incompatible-prop")
	if err != nil {
		t.Fatal(err)
	}
	if otherAssignment.AssetId != "prop" || otherAssignment.TagId != tag.TagId {
		t.Fatalf("expected the global tag to be reusable across assets, got %+v", otherAssignment)
	}
	if err = DeleteCheckpointTag(tx, tag.Id); err == nil {
		t.Fatal("expected referenced tag deletion to fail")
	}
}

func TestDependencySelectorValidationRejectsInvalidOwnershipAndCycles(t *testing.T) {
	_, tx := openDependencyTestDB(t)
	insertDependencyType(t, tx)
	for _, assetId := range []string{"shot", "boy", "other", "asset-a", "asset-b", "asset-c", "asset-d"} {
		insertDependencyAsset(t, tx, assetId)
	}

	insertTestCheckpoint(t, tx, "other-cp", "other", "other-v1", 1)
	if _, err := AddDependencyWithSelector(
		tx,
		"invalid-pin",
		"shot",
		"boy",
		"reference",
		DependencyResolutionPinned,
		stringReference("other-cp"),
		nil,
	); err == nil {
		t.Fatal("expected checkpoint ownership validation to fail")
	}
	insertTestCheckpoint(t, tx, "other-release-cp", "other", "other-release", 2)
	insertTestCheckpoint(t, tx, "asset-c-release-cp", "asset-c", "other-release", 3)
	otherTag, err := SetCheckpointTag(tx, "", "other-approved", "other-release-cp")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = AddDependencyWithSelector(
		tx,
		"invalid-tag",
		"shot",
		"boy",
		"reference",
		DependencyResolutionTagged,
		nil,
		&otherTag.Id,
	); err == nil {
		t.Fatal("expected tag asset ownership validation to fail")
	}

	if _, err := AddDependency(tx, "a-to-b", "asset-a", "asset-b", "reference"); err != nil {
		t.Fatal(err)
	}
	if _, err := AddDependency(tx, "b-to-a", "asset-b", "asset-a", "reference"); err == nil {
		t.Fatal("expected dependency cycle validation to fail")
	}

	if _, err := tx.Exec(`
		INSERT INTO asset_dependency (
			id, mtime, asset_id, dependency_id, dependency_type_id,
			resolution_mode, checkpoint_id
		) VALUES ('invalid-shape', 1, 'asset-c', 'asset-d', 'reference', 'floating', 'other-cp')
	`); err == nil {
		t.Fatal("expected database selector constraint to fail")
	}
}
