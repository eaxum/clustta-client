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

func TestResolveCollectionPathUsesCategoryMetadata(t *testing.T) {
	directoryStructure := integrations.DirectoryStructure{
		Style: "lowercase",
		Paths: map[string]interface{}{
			"asset": map[string]interface{}{
				"template": "Assets/<CollectionType>/<Category>/<Asset>/<OutputName><TemplateExtension>",
			},
		},
	}
	collection := integrations.ExternalCollection{
		Name: "Chair",
		Type: "Props",
		Metadata: map[string]interface{}{
			"Category": "Furniture",
		},
	}

	path := resolveCollectionPath(collection, map[string]integrations.ExternalCollection{}, directoryStructure)

	if path != "/Assets/props/furniture/chair/" {
		t.Fatalf("expected category collection path, got %q", path)
	}
}

func TestResolveCollectionPathOmitsEmptyCategory(t *testing.T) {
	directoryStructure := integrations.DirectoryStructure{
		Style: "lowercase",
		Paths: map[string]interface{}{
			"asset": map[string]interface{}{
				"template": "Assets/<CollectionType>/<Category>/<Asset>/<OutputName><TemplateExtension>",
			},
		},
	}
	collection := integrations.ExternalCollection{
		Name: "Ball",
		Type: "Props",
		Metadata: map[string]interface{}{
			"category": "",
		},
	}

	path := resolveCollectionPath(collection, map[string]integrations.ExternalCollection{}, directoryStructure)

	if path != "/Assets/props/ball/" {
		t.Fatalf("expected flat collection path, got %q", path)
	}
}

func TestResolveAssetParentPathUsesCategoryAfterAsset(t *testing.T) {
	directoryStructure := integrations.DirectoryStructure{
		Style: "kebab-case",
		Paths: map[string]interface{}{
			"asset": map[string]interface{}{
				"template": "Assets/<CollectionType>s/<Asset>/<Category>/<OutputName><TemplateExtension>",
			},
		},
	}
	collection := integrations.ExternalCollection{
		Name: "Broom",
		Type: "Prop",
		Metadata: map[string]interface{}{
			"Category": "Interactive Props",
		},
	}
	asset := integrations.ExternalAsset{Name: "Modeling", AssetType: "Modeling"}

	path := resolveAssetParentPathFromTemplate("/Assets/props/broom/", collection, asset, nil, directoryStructure, ".blend")

	if path != "/Assets/props/broom/interactive-props/" {
		t.Fatalf("expected category task-output path, got %q", path)
	}
}

func TestResolveAssetParentPathOmitsMissingCategory(t *testing.T) {
	directoryStructure := integrations.DirectoryStructure{
		Style: "lowercase",
		Paths: map[string]interface{}{
			"asset": map[string]interface{}{
				"template": "Assets/<CollectionType>s/<Asset>/<Category>/<OutputName><TemplateExtension>",
			},
		},
	}
	collection := integrations.ExternalCollection{Name: "Broom Stick", Type: "Prop"}
	asset := integrations.ExternalAsset{Name: "Modeling", AssetType: "Modeling"}

	path := resolveAssetParentPathFromTemplate("/Assets/props/broom-stick/", collection, asset, nil, directoryStructure, ".blend")

	if path != "/Assets/props/broom stick/" {
		t.Fatalf("expected flat task-output path, got %q", path)
	}
}

func TestReplaceCategoryVariableSupportsStructuredMetadata(t *testing.T) {
	collection := integrations.ExternalCollection{
		Metadata: map[string]interface{}{
			"Category": map[string]interface{}{"name": "Interactive Props"},
		},
	}

	resolved := replaceCategoryVariable("<Category>", collection, "kebab-case")

	if resolved != "interactive-props" {
		t.Fatalf("expected structured category metadata, got %q", resolved)
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

func TestBuildPreviewItemsSortsAlphabetically(t *testing.T) {
	collections := []integrations.SyncCollection{
		{ExternalID: "zebra", ExternalName: "Zebra", CollectionPath: "/zebra/"},
		{ExternalID: "alpha", ExternalName: "alpha", CollectionPath: "/alpha/"},
		{ExternalID: "item-10", ExternalName: "item-10", CollectionPath: "/item-10/"},
		{ExternalID: "item-2", ExternalName: "item-2", CollectionPath: "/item-2/"},
	}

	items := buildPreviewItems(collections, nil)
	names := make([]string, 0, len(items))
	for _, item := range items {
		if !item.IsVirtual {
			names = append(names, item.Name)
		}
	}

	want := []string{"alpha", "item-10", "item-2", "zebra"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("unexpected preview order: got %#v want %#v", names, want)
	}
}
