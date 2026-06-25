package services

import (
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
