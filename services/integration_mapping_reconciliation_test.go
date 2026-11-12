package services

import (
	"testing"

	"clustta/internal/integrations"
	"clustta/internal/repository/models"
)

func TestReconcileMappingsRepairsChangedCollectionAndAssetIDs(t *testing.T) {
	externalCollections := []integrations.ExternalCollection{
		{ID: "new-parent", Name: "Edit", Type: "edit"},
	}
	externalAssets := []integrations.ExternalAsset{
		{ID: "new-task", ParentID: "new-parent", Name: "Edit", AssetType: "Edit"},
	}
	existingCollections := []models.IntegrationCollectionMapping{
		{
			Id:           "collection-mapping",
			ExternalId:   "old-parent",
			ExternalName: "Edit",
			ExternalType: "edit",
			ExternalPath: "/Edits/Edit/",
			CollectionId: "local-collection",
		},
	}
	existingAssets := []models.IntegrationAssetMapping{
		{
			Id:               "asset-mapping",
			ExternalId:       "old-task",
			ExternalName:     "Edit",
			ExternalParentId: "old-parent",
			ExternalType:     "Edit",
			AssetId:          "local-asset",
		},
	}

	result := reconcileMappings(
		externalCollections,
		map[string]string{"new-parent": "/Edits/Edit/"},
		externalAssets,
		existingCollections,
		existingAssets,
	)

	if result.collectionStatuses["new-parent"] != mappingStatusRepairable {
		t.Fatalf("expected collection to be repairable, got %q", result.collectionStatuses["new-parent"])
	}
	if result.assetStatuses["new-task"] != mappingStatusRepairable {
		t.Fatalf("expected asset to be repairable, got %q", result.assetStatuses["new-task"])
	}
	if len(result.issues) != 0 {
		t.Fatalf("expected no unresolved issues, got %d", len(result.issues))
	}
}

func TestReconcileMappingsReportsAmbiguousAndMissingAssets(t *testing.T) {
	externalAssets := []integrations.ExternalAsset{
		{ID: "candidate-a", ParentID: "parent", AssetType: "Edit"},
		{ID: "candidate-b", ParentID: "parent", AssetType: "Edit"},
	}
	existingAssets := []models.IntegrationAssetMapping{
		{Id: "ambiguous", ExternalId: "old-a", ExternalParentId: "parent", ExternalType: "Edit"},
		{Id: "missing", ExternalId: "old-b", ExternalParentId: "other", ExternalType: "Edit"},
	}

	result := reconcileMappings(nil, nil, externalAssets, nil, existingAssets)

	if countMappingIssues(result.issues, mappingStatusAmbiguous) != 1 {
		t.Fatalf("expected one ambiguous issue, got %#v", result.issues)
	}
	if countMappingIssues(result.issues, mappingStatusMissing) != 1 {
		t.Fatalf("expected one missing issue, got %#v", result.issues)
	}
}

func TestReconcileMappingsMarksMetadataChanges(t *testing.T) {
	externalAssets := []integrations.ExternalAsset{
		{ID: "task", ParentID: "parent", AssetType: "Edit", Status: "done"},
	}
	existingAssets := []models.IntegrationAssetMapping{
		{Id: "mapping", ExternalId: "task", ExternalParentId: "parent", ExternalType: "Edit", ExternalStatus: "todo"},
	}

	result := reconcileMappings(nil, nil, externalAssets, nil, existingAssets)

	if result.assetStatuses["task"] != mappingStatusChanged {
		t.Fatalf("expected changed status, got %q", result.assetStatuses["task"])
	}
	if mappingAction(result.assetStatuses["task"]) != "update" {
		t.Fatalf("expected update action, got %q", mappingAction(result.assetStatuses["task"]))
	}
}

func TestConsumeUniqueMappingIssueUsesLocalItemAndType(t *testing.T) {
	issues := []*mappingIssue{
		{itemType: "asset", mappingID: "mapping", localID: "local", externalType: "Lighting", status: mappingStatusMissing},
		{itemType: "asset", mappingID: "other", localID: "other", externalType: "Lighting", status: mappingStatusMissing},
	}

	reconciliation := mappingReconciliation{issuesByLocalType: make(map[string][]*mappingIssue)}
	for _, issue := range issues {
		reconciliation.addIssue(issue)
	}
	issue, found := reconciliation.consumeUniqueIssue("asset", "local", "lighting")

	if !found || issue.mappingID != "mapping" {
		t.Fatalf("expected the matching issue, got %#v", issue)
	}
	if countMappingIssues(reconciliation.issues, mappingStatusMissing) != 1 {
		t.Fatalf("expected the unrelated issue to remain unresolved, got %#v", reconciliation.issues)
	}
}

func TestApplyMappingIssueStatusesMarksExistingPreviewItems(t *testing.T) {
	preview := integrations.SyncPreview{
		Assets: []integrations.SyncAsset{
			{AssetID: "local-asset", ExternalType: "Lighting", Action: "skip", MappingStatus: mappingStatusLinked},
			{AssetID: "local-asset", ExternalType: "FX", Action: "skip", MappingStatus: mappingStatusLinked},
		},
		Collections: []integrations.SyncCollection{
			{CollectionID: "local-collection", ExternalType: "shot", Action: "skip", MappingStatus: mappingStatusLinked},
		},
	}
	issues := []*mappingIssue{
		{itemType: "asset", localID: "local-asset", externalType: "Lighting", status: mappingStatusMissing},
		{itemType: "collection", localID: "local-collection", externalType: "shot", status: mappingStatusAmbiguous},
	}

	applyMappingIssueStatuses(&preview, issues)

	if preview.Assets[0].MappingStatus != mappingStatusMissing {
		t.Fatalf("expected missing asset status, got %q", preview.Assets[0].MappingStatus)
	}
	if preview.Collections[0].MappingStatus != mappingStatusAmbiguous {
		t.Fatalf("expected ambiguous collection status, got %q", preview.Collections[0].MappingStatus)
	}
	if preview.Assets[1].MappingStatus != mappingStatusLinked {
		t.Fatalf("expected unrelated asset type to remain linked, got %q", preview.Assets[1].MappingStatus)
	}
}

func TestReconcileMappingsDoesNotCleanUpDuplicateWithMultipleLiveMatches(t *testing.T) {
	externalAssets := []integrations.ExternalAsset{
		{ID: "current-a", ParentID: "parent", AssetType: "Lighting"},
		{ID: "current-b", ParentID: "parent", AssetType: "Lighting"},
	}
	existingAssets := []models.IntegrationAssetMapping{
		{Id: "current-map-a", ExternalId: "current-a", ExternalParentId: "parent", ExternalType: "Lighting", AssetId: "local-asset"},
		{Id: "current-map-b", ExternalId: "current-b", ExternalParentId: "parent", ExternalType: "Lighting", AssetId: "local-asset"},
		{Id: "stale-map", ExternalId: "old", ExternalParentId: "parent", ExternalType: "Lighting", AssetId: "local-asset"},
	}

	result := reconcileMappings(nil, nil, externalAssets, nil, existingAssets)

	if len(result.obsoleteAssetMappingIDs) != 0 {
		t.Fatalf("expected no automatic cleanup for ambiguous live matches, got %#v", result.obsoleteAssetMappingIDs)
	}
	if countMappingIssues(result.issues, mappingStatusAmbiguous) != 1 {
		t.Fatalf("expected the stale mapping to remain ambiguous, got %#v", result.issues)
	}
}

func TestReconcileMappingsMarksStaleDuplicatesForCleanup(t *testing.T) {
	externalCollections := []integrations.ExternalCollection{
		{ID: "current-parent", Type: "shot"},
	}
	externalAssets := []integrations.ExternalAsset{
		{ID: "current-task", ParentID: "current-parent", AssetType: "Lighting"},
	}
	existingCollections := []models.IntegrationCollectionMapping{
		{Id: "current-collection-map", ExternalId: "current-parent", ExternalType: "shot", CollectionId: "local-shot"},
		{Id: "stale-collection-map", ExternalId: "old-parent", ExternalType: "shot", CollectionId: "local-shot"},
	}
	existingAssets := []models.IntegrationAssetMapping{
		{Id: "current-asset-map", ExternalId: "current-task", ExternalParentId: "current-parent", ExternalType: "Lighting", AssetId: "local-asset"},
		{Id: "stale-asset-map", ExternalId: "old-task", ExternalParentId: "old-parent", ExternalType: "Lighting", AssetId: "local-asset"},
	}

	result := reconcileMappings(
		externalCollections,
		map[string]string{"current-parent": "/Shots/0090/"},
		externalAssets,
		existingCollections,
		existingAssets,
	)

	if result.collectionStatuses["current-parent"] != mappingStatusRepairable {
		t.Fatalf("expected collection relink, got %q", result.collectionStatuses["current-parent"])
	}
	if result.assetStatuses["current-task"] != mappingStatusRepairable {
		t.Fatalf("expected asset relink, got %q", result.assetStatuses["current-task"])
	}
	if len(result.obsoleteCollectionMappingIDs["current-parent"]) != 1 {
		t.Fatalf("expected one obsolete collection mapping, got %#v", result.obsoleteCollectionMappingIDs)
	}
	if len(result.obsoleteAssetMappingIDs["current-task"]) != 1 {
		t.Fatalf("expected one obsolete asset mapping, got %#v", result.obsoleteAssetMappingIDs)
	}
	if len(result.issues) != 0 {
		t.Fatalf("expected duplicate mappings to be repairable, got %#v", result.issues)
	}
}
