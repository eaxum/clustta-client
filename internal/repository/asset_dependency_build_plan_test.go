package repository

import (
	"testing"

	"clustta/internal/repository/models"

	"github.com/jmoiron/sqlx"
)

func addBuildPlanCheckpoint(t *testing.T, tx *sqlx.Tx, assetId, checkpointId string, createdAt int) {
	t.Helper()
	groupId := checkpointId + "-group"
	if _, err := BeginCheckpointGroup(tx, groupId, CheckpointGroupTypeSingle); err != nil {
		t.Fatal(err)
	}
	insertCheckpointGroupMember(t, tx, checkpointId, assetId, groupId, createdAt)
	if _, err := FinalizeCheckpointGroup(tx, groupId); err != nil {
		t.Fatal(err)
	}
}

func TestDependencyBuildPlanUsesExactSelectorsAndDetectsConflicts(t *testing.T) {
	_, tx := openCheckpointGroupTestDB(t)
	if _, err := tx.Exec(
		"INSERT INTO config (name, value, mtime) VALUES ('working_dir', ?, 1)",
		t.TempDir(),
	); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`
		INSERT INTO status (id, mtime, name, short_name) VALUES ('status', 1, 'Todo', 'todo');
		INSERT INTO asset_type (id, mtime, name, icon) VALUES ('asset-type', 1, 'Asset', 'asset');
	`); err != nil {
		t.Fatal(err)
	}
	insertDependencyType(t, tx)
	for _, assetId := range []string{"root", "left", "right", "boy", "prop"} {
		insertDependencyAsset(t, tx, assetId)
	}
	addBuildPlanCheckpoint(t, tx, "root", "root-cp", 1)
	addBuildPlanCheckpoint(t, tx, "left", "left-cp", 2)
	addBuildPlanCheckpoint(t, tx, "right", "right-cp", 3)
	addBuildPlanCheckpoint(t, tx, "boy", "boy-cp-1", 4)
	addBuildPlanCheckpoint(t, tx, "boy", "boy-cp-2", 5)
	if _, err := BeginCheckpointGroup(tx, "release", CheckpointGroupTypeMulti); err != nil {
		t.Fatal(err)
	}
	insertCheckpointGroupMember(t, tx, "release-boy", "boy", "release", 6)
	insertCheckpointGroupMember(t, tx, "release-prop", "prop", "release", 7)
	if _, err := FinalizeCheckpointGroup(tx, "release"); err != nil {
		t.Fatal(err)
	}
	tag, err := SetCheckpointGroupTag(tx, "release-tag", "animation-approved", "release")
	if err != nil {
		t.Fatal(err)
	}

	edges := []struct {
		id           string
		owner        string
		dependency   string
		mode         string
		checkpointId *string
	}{
		{id: "root-left", owner: "root", dependency: "left", mode: DependencyResolutionFloating},
		{id: "root-right", owner: "root", dependency: "right", mode: DependencyResolutionFloating},
		{id: "left-boy", owner: "left", dependency: "boy", mode: DependencyResolutionPinned, checkpointId: stringReference("boy-cp-1")},
		{id: "right-boy", owner: "right", dependency: "boy", mode: DependencyResolutionFloating},
	}
	for _, edge := range edges {
		if _, err := AddDependencyWithSelector(
			tx,
			edge.id,
			edge.owner,
			edge.dependency,
			"reference",
			edge.mode,
			edge.checkpointId,
			nil,
		); err != nil {
			t.Fatal(err)
		}
	}

	plan, err := ResolveDependencyBuildPlan(tx, "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Conflicts) != 0 {
		t.Fatalf("expected a compatible plan, got %+v", plan.Conflicts)
	}
	if len(plan.Entries) != 4 || plan.Entries[len(plan.Entries)-1].AssetId != "root" {
		t.Fatalf("expected dependency-first entries ending in root, got %+v", plan.Entries)
	}
	boyEntry := models.DependencyBuildPlanEntry{}
	for _, entry := range plan.Entries {
		if entry.AssetId == "boy" {
			boyEntry = entry
		}
	}
	if boyEntry.CheckpointId != "boy-cp-1" || boyEntry.ResolutionMode != DependencyResolutionPinned {
		t.Fatalf("expected the pin to override floating, got %+v", boyEntry)
	}
	secondPlan, err := ResolveDependencyBuildPlan(tx, "root")
	if err != nil {
		t.Fatal(err)
	}
	if secondPlan.Fingerprint != plan.Fingerprint {
		t.Fatal("expected unchanged dependency metadata to produce the same fingerprint")
	}

	if _, err = UpdateDependencySelector(
		tx,
		"right-boy",
		DependencyResolutionTagged,
		nil,
		&tag.Id,
	); err != nil {
		t.Fatal(err)
	}
	conflictingPlan, err := ResolveDependencyBuildPlan(tx, "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(conflictingPlan.Conflicts) != 1 || conflictingPlan.Conflicts[0].AssetId != "boy" {
		t.Fatalf("expected one boy checkpoint conflict, got %+v", conflictingPlan.Conflicts)
	}
}
