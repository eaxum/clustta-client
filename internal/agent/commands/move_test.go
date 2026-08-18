package commands

import (
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"testing"
)

func TestPairSameNameMoveSourcesUsesNameAndParent(t *testing.T) {
	sources := []scope.Entity{
		{Type: scope.TypeAsset, ID: "hero-task", Name: "Hero", ParentID: "characters"},
		{Type: scope.TypeAsset, ID: "hero-prop", Name: "Hero", ParentID: "props"},
		{Type: scope.TypeAsset, ID: "missing", Name: "Missing", ParentID: "characters"},
	}
	candidates := map[string][]scope.Entity{
		moveSiblingKey("characters", "hero"): {
			{Type: scope.TypeCollection, ID: "hero-character", Name: "hero", ParentID: "characters", Path: "characters/hero"},
		},
		moveSiblingKey("props", "Hero"): {
			{Type: scope.TypeCollection, ID: "hero-prop-collection", Name: "Hero", ParentID: "props", Path: "props/Hero"},
		},
	}

	pairs := pairSameNameMoveSources(sources, candidates)

	if pairs[0].Destination.CollectionID != "hero-character" {
		t.Fatalf("first destination = %q, want hero-character", pairs[0].Destination.CollectionID)
	}
	if pairs[1].Destination.CollectionID != "hero-prop-collection" {
		t.Fatalf("second destination = %q, want hero-prop-collection", pairs[1].Destination.CollectionID)
	}
	if pairs[2].Destination.CollectionID != "" || pairs[2].Warning == "" {
		t.Fatalf("missing destination pair = %#v, want warning without destination", pairs[2])
	}
}

func TestParseMoveMappingsRejectsDuplicateSources(t *testing.T) {
	raw := []map[string]interface{}{
		{"entity_id": "hero", "entity_type": "asset", "target_collection_id": "characters"},
		{"entity_id": "hero", "entity_type": "asset", "target_collection_id": "archive"},
	}

	if _, err := parseMoveMappings(raw); err == nil {
		t.Fatal("parseMoveMappings() accepted duplicate source mappings")
	}
}

func TestSelectMovesPreservesApprovedDestinations(t *testing.T) {
	first := planning.Change{
		Entity: scope.Entity{Type: scope.TypeAsset, ID: "hero"}, Valid: true,
		After: map[string]interface{}{"parent_id": "characters"},
	}
	second := planning.Change{
		Entity: scope.Entity{Type: scope.TypeAsset, ID: "axe"}, Valid: true,
		After: map[string]interface{}{"parent_id": "props"},
	}
	args := map[string]interface{}{}
	approved := planning.Plan{Changes: []planning.Change{first, second}}

	err := selectMoves(args, approved, []string{"asset:axe"})
	if err != nil {
		t.Fatal(err)
	}
	mappings, err := parseMoveMappings(args[moveMappings])
	if err != nil {
		t.Fatal(err)
	}
	if len(mappings) != 1 || mappings[0].EntityID != "axe" || mappings[0].TargetCollectionID != "props" {
		t.Fatalf("selected mappings = %#v, want axe -> props", mappings)
	}
}
