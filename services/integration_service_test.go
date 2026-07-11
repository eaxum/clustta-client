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
