package commands

import (
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"testing"
)

func TestPairAllToEachCreatesEveryLink(t *testing.T) {
	targets := []scope.Entity{
		{Type: scope.TypeAsset, ID: "house-1"},
		{Type: scope.TypeAsset, ID: "house-2"},
	}
	dependencies := []scope.Entity{
		{Type: scope.TypeAsset, ID: "ben-1"},
		{Type: scope.TypeAsset, ID: "ben-2"},
		{Type: scope.TypeAsset, ID: "ben-3"},
	}

	pairs := pairAllToEach(targets, dependencies)

	if len(pairs) != 6 {
		t.Fatalf("pairAllToEach() created %d pairs, want 6", len(pairs))
	}
	if pairs[0].Target.ID != "house-1" || pairs[0].Dependency.ID != "ben-1" {
		t.Fatalf("first pair = %s -> %s, want house-1 -> ben-1", pairs[0].Target.ID, pairs[0].Dependency.ID)
	}
	if pairs[5].Target.ID != "house-2" || pairs[5].Dependency.ID != "ben-3" {
		t.Fatalf("last pair = %s -> %s, want house-2 -> ben-3", pairs[5].Target.ID, pairs[5].Dependency.ID)
	}
}

func TestPairSameNameSiblingsUsesNameAndParent(t *testing.T) {
	targets := []scope.Entity{
		{Type: scope.TypeAsset, ID: "hero-task", Name: "Hero", ParentID: "characters"},
		{Type: scope.TypeAsset, ID: "hero-other", Name: "Hero", ParentID: "props"},
		{Type: scope.TypeAsset, ID: "missing", Name: "Missing", ParentID: "characters"},
	}
	dependencies := []scope.Entity{
		{Type: scope.TypeCollection, ID: "hero-collection", Name: "hero", ParentID: "characters"},
		{Type: scope.TypeCollection, ID: "hero-props", Name: "Hero", ParentID: "props"},
	}

	pairs := pairSameNameSiblings(targets, dependencies)

	if pairs[0].Dependency.ID != "hero-collection" {
		t.Fatalf("first dependency = %q, want hero-collection", pairs[0].Dependency.ID)
	}
	if pairs[1].Dependency.ID != "hero-props" {
		t.Fatalf("second dependency = %q, want hero-props", pairs[1].Dependency.ID)
	}
	if pairs[2].Dependency.ID != "" || pairs[2].Warning == "" {
		t.Fatalf("missing sibling pair = %#v, want warning without dependency", pairs[2])
	}
}

func TestSelectDependenciesPreservesIndividualPairs(t *testing.T) {
	first := planning.Change{
		Key: "asset:house->asset:ben-1", Entity: scope.Entity{Type: scope.TypeAsset, ID: "house"}, Valid: true,
		After: map[string]interface{}{"dependency_id": "ben-1", "dependency_entity_type": "asset"},
	}
	second := planning.Change{
		Key: "asset:house->asset:ben-2", Entity: scope.Entity{Type: scope.TypeAsset, ID: "house"}, Valid: true,
		After: map[string]interface{}{"dependency_id": "ben-2", "dependency_entity_type": "asset"},
	}
	args := map[string]interface{}{}
	approved := planning.Plan{
		Options: map[string]interface{}{"mode": dependencyModePaired},
		Changes: []planning.Change{first, second},
	}

	err := selectDependencies(args, approved, []string{second.Key})
	if err != nil {
		t.Fatal(err)
	}
	pairs, err := parseRequestedDependencyPairs(args[dependencyPairs])
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 1 || pairs[0].DependencyID != "ben-2" {
		t.Fatalf("selected pairs = %#v, want only ben-2", pairs)
	}
}
