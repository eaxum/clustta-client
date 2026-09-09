package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"clustta/internal/repository/models"

	"github.com/jmoiron/sqlx"
)

func addBuildPlanCheckpoint(t *testing.T, tx *sqlx.Tx, assetId, checkpointId string, createdAt int) {
	t.Helper()
	groupId := checkpointId + "-group"
	insertTestCheckpoint(t, tx, checkpointId, assetId, groupId, createdAt)
}

func TestDependencyBuildPlanUsesExactSelectorsAndDetectsConflicts(t *testing.T) {
	_, tx := openDependencyTestDB(t)
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
	insertTestCheckpoint(t, tx, "release-boy", "boy", "release", 6)
	insertTestCheckpoint(t, tx, "release-prop", "prop", "release", 7)
	tag, err := SetCheckpointTag(tx, "", "animation-approved", "release-boy")
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
	assertGraphMatchesBuildPlan(t, tx, plan)

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
	assertGraphMatchesBuildPlan(t, tx, conflictingPlan)
}

func assertGraphMatchesBuildPlan(t *testing.T, tx *sqlx.Tx, buildPlan models.DependencyBuildPlan) {
	t.Helper()
	graphPlan, err := ResolveDependencyGraphPlan(tx, buildPlan.RootAssetId)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(graphPlan.Conflicts, buildPlan.Conflicts) {
		t.Fatalf("graph conflicts differ from build: %+v", graphPlan.Conflicts)
	}
	if len(graphPlan.Entries) != len(buildPlan.Entries) {
		t.Fatalf("graph and build entry counts differ: %d != %d", len(graphPlan.Entries), len(buildPlan.Entries))
	}
	for i, entry := range graphPlan.Entries {
		buildEntry := buildPlan.Entries[i]
		if entry.AssetId != buildEntry.AssetId || entry.CheckpointId != buildEntry.CheckpointId || entry.DependencyEdgeId != buildEntry.DependencyEdgeId {
			t.Fatalf("graph checkpoint selection differs from build: %+v != %+v", entry, buildEntry)
		}
	}
}

func TestDependencyGraphPlanSkipsBuildReadiness(t *testing.T) {
	_, tx := openDependencyTestDB(t)
	workingDir := t.TempDir()
	if _, err := tx.Exec("INSERT INTO config (name, value, mtime) VALUES ('working_dir', ?, 1)", workingDir); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`
		INSERT INTO status (id, mtime, name, short_name) VALUES ('status', 1, 'Todo', 'todo');
		INSERT INTO asset_type (id, mtime, name, icon) VALUES ('asset-type', 1, 'Asset', 'asset');
	`); err != nil {
		t.Fatal(err)
	}
	insertDependencyAsset(t, tx, "root")
	addBuildPlanCheckpoint(t, tx, "root", "root-cp", 1)
	assetPath := filepath.Join(workingDir, "root.blend")
	if err := os.WriteFile(assetPath, []byte("modified asset"), 0600); err != nil {
		t.Fatal(err)
	}
	buildPlan, err := ResolveDependencyBuildPlan(tx, "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(buildPlan.Entries) != 1 {
		t.Fatalf("expected one build entry, got %+v", buildPlan.Entries)
	}
	entry := buildPlan.Entries[0]
	if !entry.MissingChunks || entry.FileStatus != "modified" || !entry.RequiresOverwrite {
		t.Fatalf("build readiness checks were skipped: %+v", entry)
	}
	assertGraphMatchesBuildPlan(t, tx, buildPlan)

	// A directory at the asset path makes any attempt to hash it fail.
	if err = os.Remove(assetPath); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(assetPath, 0700); err != nil {
		t.Fatal(err)
	}
	assertGraphMatchesBuildPlan(t, tx, buildPlan)
	if _, err = ResolveDependencyBuildPlan(tx, "root"); err == nil {
		t.Fatal("expected build readiness to reject an unreadable asset")
	}
	if _, err = tx.Exec("DROP TABLE chunk"); err != nil {
		t.Fatal(err)
	}
	assertGraphMatchesBuildPlan(t, tx, buildPlan)
	if _, err = ResolveDependencyBuildPlan(tx, "root"); err == nil || !strings.Contains(err.Error(), "chunk") {
		t.Fatalf("expected build readiness to query chunks, got %v", err)
	}
	graphPlan, err := ResolveDependencyGraphPlan(tx, "root")
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(graphPlan)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"missing_chunks", "file_status", "requires_overwrite"} {
		if strings.Contains(string(payload), field) {
			t.Fatalf("graph response exposes unchecked build field %s", field)
		}
	}
}
