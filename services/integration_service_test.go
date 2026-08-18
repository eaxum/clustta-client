package services

import (
	"reflect"
	"testing"

	"clustta/internal/integrations"
)

func TestResolveAssetNameFromTemplateUsesFinalFilenameSegment(t *testing.T) {
	dirStructure := integrations.DirectoryStructure{
		Style: "lowercase",
		Paths: map[string]interface{}{
			"asset": map[string]interface{}{
				"template": "assets/<CollectionType>s/<Asset>/<Asset>_<AssetType><TemplateExtension>",
			},
		},
	}
	parentCollection := integrations.ExternalCollection{
		ID:   "asset-hero",
		Name: "Hero",
		Type: "Character",
	}
	asset := integrations.ExternalAsset{
		Name:      "Modeling",
		AssetType: "Modeling",
	}

	name := resolveAssetNameFromTemplate("Modeling", parentCollection, asset, map[string]integrations.ExternalCollection{}, dirStructure, ".etx")

	if name != "hero_modeling" {
		t.Fatalf("expected hero_modeling, got %q", name)
	}
}

func TestResolveAssetNameFromTemplateFallsBackWhenNoAssetTemplateMatches(t *testing.T) {
	dirStructure := integrations.DirectoryStructure{
		Style: "lowercase",
		Paths: map[string]interface{}{
			"shot": map[string]interface{}{
				"template": "shots/<Shot>/<AssetType><TemplateExtension>",
			},
		},
	}
	parentCollection := integrations.ExternalCollection{
		ID:   "asset-hero",
		Name: "Hero",
		Type: "Character",
	}
	asset := integrations.ExternalAsset{
		Name:      "Modeling",
		AssetType: "Modeling",
	}

	name := resolveAssetNameFromTemplate("Modeling", parentCollection, asset, map[string]integrations.ExternalCollection{}, dirStructure, ".etx")

	if name != "Modeling" {
		t.Fatalf("expected fallback Modeling, got %q", name)
	}
}

func TestBuildInboundStatusMappingsReversesOutboundMappings(t *testing.T) {
	got := buildInboundStatusMappings(map[string]string{
		"local-todo": "kitsu-todo",
		"local-wip":  "kitsu-wip",
		"":           "ignored-external",
		"ignored":    "",
	})

	want := map[string]string{
		"kitsu-todo": "local-todo",
		"kitsu-wip":  "local-wip",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected inbound status mappings: got %#v want %#v", got, want)
	}
}

func TestBuildInboundStatusMappingsExcludesAmbiguousExternalStatuses(t *testing.T) {
	got := buildInboundStatusMappings(map[string]string{
		"local-wip":    "kitsu-in-progress",
		"local-review": "kitsu-in-progress",
		"local-done":   "kitsu-done",
	})

	want := map[string]string{
		"kitsu-done": "local-done",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected inbound status mappings: got %#v want %#v", got, want)
	}
}

func TestIsTaskOutputEnabledDefaultsToEnabled(t *testing.T) {
	parent := integrations.ExternalCollection{Type: "Shot"}
	asset := integrations.ExternalAsset{AssetType: "Lighting"}
	directoryStructure := integrations.DirectoryStructure{
		Paths: map[string]interface{}{
			"shot": map[string]interface{}{
				"template": "Shots/<Shot>/<OutputName><TemplateExtension>",
			},
		},
	}

	if !isTaskOutputEnabled(parent, asset, directoryStructure) {
		t.Fatal("expected task output without an explicit selection to be enabled")
	}
}

func TestIsTaskOutputEnabledUsesCaseInsensitiveTaskSelection(t *testing.T) {
	parent := integrations.ExternalCollection{Type: "Shot"}
	asset := integrations.ExternalAsset{AssetType: "Lighting"}
	directoryStructure := integrations.DirectoryStructure{
		Paths: map[string]interface{}{
			"shot": map[string]interface{}{
				"template": "Shots/<Shot>/<OutputName><TemplateExtension>",
				"task_output_enabled": map[string]interface{}{
					"lighting": false,
				},
			},
		},
	}

	if isTaskOutputEnabled(parent, asset, directoryStructure) {
		t.Fatal("expected disabled task output to be excluded")
	}
}

func TestBuildPreviewItemsPreservesActionStates(t *testing.T) {
	collections := []integrations.SyncCollection{
		{ExternalID: "create-collection", ExternalName: "Create", CollectionPath: "/create/", Action: "create", Selected: true},
		{ExternalID: "link-collection", ExternalName: "Link", CollectionPath: "/link/", Action: "link", Selected: true},
		{ExternalID: "skip-collection", ExternalName: "Skip", CollectionPath: "/skip/", Action: "skip"},
	}
	assets := []integrations.SyncAsset{
		{ExternalID: "create-asset", ExternalName: "Create asset", CollectionPath: "/create/", Action: "create", Selected: true},
		{ExternalID: "link-asset", ExternalName: "Link asset", CollectionPath: "/link/", Action: "link", Selected: true},
		{ExternalID: "skip-asset", ExternalName: "Skip asset", CollectionPath: "/skip/", Action: "skip"},
	}

	items := buildPreviewItems(collections, assets)
	actions := make(map[string]string)
	for _, item := range items {
		if item.ExternalID != "" {
			actions[item.ExternalID] = item.Action
		}
	}

	want := map[string]string{
		"create-collection": "create",
		"link-collection":   "link",
		"skip-collection":   "skip",
		"create-asset":      "create",
		"link-asset":        "link",
		"skip-asset":        "skip",
	}
	if !reflect.DeepEqual(actions, want) {
		t.Fatalf("unexpected preview actions: got %#v want %#v", actions, want)
	}
}

func TestBuildPreviewItemsUsesResolvedCollectionName(t *testing.T) {
	collections := []integrations.SyncCollection{
		{
			ExternalID:     "collection-1",
			ExternalName:   "Broom Stick",
			CollectionPath: "/assets/props/broom-stick/",
			Action:         "create",
		},
	}

	items := buildPreviewItems(collections, nil)
	for _, item := range items {
		if item.ExternalID == "collection-1" {
			if item.Name != "broom-stick" {
				t.Fatalf("expected resolved collection name broom-stick, got %q", item.Name)
			}
			return
		}
	}
	t.Fatal("expected collection preview item")
}
