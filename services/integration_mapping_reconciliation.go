package services

import (
	"path"
	"strings"

	"clustta/internal/integrations"
	"clustta/internal/repository/models"
)

const (
	mappingStatusLinked     = "linked"
	mappingStatusChanged    = "changed"
	mappingStatusRepairable = "repairable"
	mappingStatusMissing    = "missing"
	mappingStatusAmbiguous  = "ambiguous"
)

type mappingReconciliation struct {
	collectionMappings           map[string]models.IntegrationCollectionMapping
	collectionStatuses           map[string]string
	obsoleteCollectionMappingIDs map[string][]string
	assetMappings                map[string]models.IntegrationAssetMapping
	assetStatuses                map[string]string
	obsoleteAssetMappingIDs      map[string][]string
	issues                       []*mappingIssue
	issuesByLocalType            map[string][]*mappingIssue
}

type mappingIssue struct {
	itemType     string
	mappingID    string
	localID      string
	externalType string
	status       string
	resolved     bool
}

func reconcileMappings(
	externalCollections []integrations.ExternalCollection,
	collectionPaths map[string]string,
	externalAssets []integrations.ExternalAsset,
	existingCollections []models.IntegrationCollectionMapping,
	existingAssets []models.IntegrationAssetMapping,
) mappingReconciliation {
	result := mappingReconciliation{
		collectionMappings:           make(map[string]models.IntegrationCollectionMapping),
		collectionStatuses:           make(map[string]string),
		obsoleteCollectionMappingIDs: make(map[string][]string),
		assetMappings:                make(map[string]models.IntegrationAssetMapping),
		assetStatuses:                make(map[string]string),
		obsoleteAssetMappingIDs:      make(map[string][]string),
		issues:                       []*mappingIssue{},
		issuesByLocalType:            make(map[string][]*mappingIssue),
	}

	collectionByID := make(map[string]integrations.ExternalCollection)
	for _, collection := range externalCollections {
		collectionByID[collection.ID] = collection
	}
	assetByID := make(map[string]integrations.ExternalAsset)
	for _, asset := range externalAssets {
		assetByID[asset.ID] = asset
	}

	staleCollections := make([]models.IntegrationCollectionMapping, 0)
	exactCollectionsByLocalType := make(map[string][]string)
	for _, mapping := range existingCollections {
		collection, exists := collectionByID[mapping.ExternalId]
		if !exists {
			staleCollections = append(staleCollections, mapping)
			continue
		}
		result.collectionMappings[collection.ID] = mapping
		key := localTypeKey(mapping.CollectionId, collection.Type)
		exactCollectionsByLocalType[key] = append(exactCollectionsByLocalType[key], collection.ID)
		status := mappingStatusLinked
		if collectionMappingChanged(mapping, collection, collectionPaths[collection.ID]) {
			status = mappingStatusChanged
		}
		result.collectionStatuses[collection.ID] = status
	}
	resolvedParentIDs := make(map[string]string)
	for externalID := range collectionByID {
		resolvedParentIDs[externalID] = externalID
	}
	remainingStaleCollections := make([]models.IntegrationCollectionMapping, 0, len(staleCollections))
	for _, mapping := range staleCollections {
		externalIDs := exactCollectionsByLocalType[localTypeKey(mapping.CollectionId, mapping.ExternalType)]
		if len(externalIDs) > 1 {
			result.addIssue(collectionMappingIssue(mapping, mappingStatusAmbiguous))
			continue
		}
		if len(externalIDs) == 0 {
			remainingStaleCollections = append(remainingStaleCollections, mapping)
			continue
		}
		externalID := externalIDs[0]
		result.collectionStatuses[externalID] = mappingStatusRepairable
		result.obsoleteCollectionMappingIDs[externalID] = append(result.obsoleteCollectionMappingIDs[externalID], mapping.Id)
		resolvedParentIDs[mapping.ExternalId] = externalID
	}
	staleCollections = remainingStaleCollections

	collectionCandidates := make(map[string][]integrations.ExternalCollection)
	for _, collection := range externalCollections {
		if strings.HasPrefix(collection.ID, "asset-type-") {
			continue
		}
		if _, mapped := result.collectionMappings[collection.ID]; mapped {
			continue
		}
		key := collectionReconciliationKey(collectionPaths[collection.ID], collection.Type)
		collectionCandidates[key] = append(collectionCandidates[key], collection)
	}
	staleCollectionCounts := make(map[string]int)
	for _, mapping := range staleCollections {
		key := collectionReconciliationKey(mapping.ExternalPath, mapping.ExternalType)
		staleCollectionCounts[key]++
	}

	for _, mapping := range staleCollections {
		key := collectionReconciliationKey(mapping.ExternalPath, mapping.ExternalType)
		candidates := collectionCandidates[key]
		if len(candidates) == 1 && staleCollectionCounts[key] == 1 {
			candidate := candidates[0]
			result.collectionMappings[candidate.ID] = mapping
			result.collectionStatuses[candidate.ID] = mappingStatusRepairable
			resolvedParentIDs[mapping.ExternalId] = candidate.ID
			delete(collectionCandidates, key)
			continue
		}
		status := mappingStatusMissing
		if len(candidates) > 0 {
			status = mappingStatusAmbiguous
		}
		result.addIssue(collectionMappingIssue(mapping, status))
	}

	staleAssets := make([]models.IntegrationAssetMapping, 0)
	exactAssetsByLocalType := make(map[string][]string)
	for _, mapping := range existingAssets {
		asset, exists := assetByID[mapping.ExternalId]
		if !exists {
			staleAssets = append(staleAssets, mapping)
			continue
		}
		result.assetMappings[asset.ID] = mapping
		key := localTypeKey(mapping.AssetId, asset.AssetType)
		exactAssetsByLocalType[key] = append(exactAssetsByLocalType[key], asset.ID)
		status := mappingStatusLinked
		if assetMappingChanged(mapping, asset) {
			status = mappingStatusChanged
		}
		result.assetStatuses[asset.ID] = status
	}
	remainingStaleAssets := make([]models.IntegrationAssetMapping, 0, len(staleAssets))
	for _, mapping := range staleAssets {
		externalIDs := exactAssetsByLocalType[localTypeKey(mapping.AssetId, mapping.ExternalType)]
		if len(externalIDs) > 1 {
			result.addIssue(assetMappingIssue(mapping, mappingStatusAmbiguous))
			continue
		}
		if len(externalIDs) == 0 {
			remainingStaleAssets = append(remainingStaleAssets, mapping)
			continue
		}
		externalID := externalIDs[0]
		result.assetStatuses[externalID] = mappingStatusRepairable
		result.obsoleteAssetMappingIDs[externalID] = append(result.obsoleteAssetMappingIDs[externalID], mapping.Id)
	}
	staleAssets = remainingStaleAssets

	assetCandidates := make(map[string][]integrations.ExternalAsset)
	for _, asset := range externalAssets {
		if _, mapped := result.assetMappings[asset.ID]; mapped {
			continue
		}
		key := assetReconciliationKey(asset.ParentID, asset.AssetType)
		assetCandidates[key] = append(assetCandidates[key], asset)
	}
	staleAssetCounts := make(map[string]int)
	for _, mapping := range staleAssets {
		parentID := mapping.ExternalParentId
		if resolvedID, exists := resolvedParentIDs[parentID]; exists {
			parentID = resolvedID
		}
		key := assetReconciliationKey(parentID, mapping.ExternalType)
		staleAssetCounts[key]++
	}
	for _, mapping := range staleAssets {
		parentID := mapping.ExternalParentId
		if resolvedID, exists := resolvedParentIDs[parentID]; exists {
			parentID = resolvedID
		}
		key := assetReconciliationKey(parentID, mapping.ExternalType)
		candidates := assetCandidates[key]
		if len(candidates) == 1 && staleAssetCounts[key] == 1 {
			candidate := candidates[0]
			result.assetMappings[candidate.ID] = mapping
			result.assetStatuses[candidate.ID] = mappingStatusRepairable
			delete(assetCandidates, key)
			continue
		}
		status := mappingStatusMissing
		if len(candidates) > 0 {
			status = mappingStatusAmbiguous
		}
		result.addIssue(assetMappingIssue(mapping, status))
	}

	return result
}

func collectionReconciliationKey(collectionPath, collectionType string) string {
	normalizedPath := strings.ToLower(path.Clean("/" + strings.TrimSpace(collectionPath)))
	return normalizedPath + "|" + strings.ToLower(strings.TrimSpace(collectionType))
}

func assetReconciliationKey(parentID, assetType string) string {
	return strings.TrimSpace(parentID) + "|" + strings.ToLower(strings.TrimSpace(assetType))
}

func localTypeKey(localID, externalType string) string {
	return strings.TrimSpace(localID) + "|" + strings.ToLower(strings.TrimSpace(externalType))
}

func collectionMappingChanged(mapping models.IntegrationCollectionMapping, collection integrations.ExternalCollection, collectionPath string) bool {
	return mapping.ExternalName != collection.Name ||
		mapping.ExternalType != collection.Type ||
		mapping.ExternalParentId != collection.ParentID ||
		mapping.ExternalPath != collectionPath
}

func assetMappingChanged(mapping models.IntegrationAssetMapping, asset integrations.ExternalAsset) bool {
	return mapping.ExternalType != asset.AssetType ||
		mapping.ExternalParentId != asset.ParentID ||
		mapping.ExternalStatus != asset.Status
}

func mappingAction(status string) string {
	switch status {
	case mappingStatusChanged:
		return "update"
	case mappingStatusRepairable:
		return "relink"
	default:
		return "skip"
	}
}

func countMappingStatuses(reconciliation mappingReconciliation, status string) int {
	count := 0
	for _, mappingStatus := range reconciliation.collectionStatuses {
		if mappingStatus == status {
			count++
		}
	}
	for _, mappingStatus := range reconciliation.assetStatuses {
		if mappingStatus == status {
			count++
		}
	}
	return count
}

func countMappingIssues(issues []*mappingIssue, status string) int {
	count := 0
	for _, issue := range issues {
		if !issue.resolved && issue.status == status {
			count++
		}
	}
	return count
}

func (reconciliation *mappingReconciliation) addIssue(issue *mappingIssue) {
	reconciliation.issues = append(reconciliation.issues, issue)
	key := mappingIssueKey(issue.itemType, issue.localID, issue.externalType)
	reconciliation.issuesByLocalType[key] = append(reconciliation.issuesByLocalType[key], issue)
}

func (reconciliation *mappingReconciliation) consumeUniqueIssue(itemType, localID, externalType string) (*mappingIssue, bool) {
	key := mappingIssueKey(itemType, localID, externalType)
	issues := reconciliation.issuesByLocalType[key]
	if len(issues) != 1 || issues[0].resolved {
		return nil, false
	}
	issues[0].resolved = true
	return issues[0], true
}

func mappingIssueKey(itemType, localID, externalType string) string {
	return itemType + "|" + localTypeKey(localID, externalType)
}

func applyMappingIssueStatuses(preview *integrations.SyncPreview, issues []*mappingIssue) {
	collectionStatuses := make(map[string]string)
	assetStatuses := make(map[string]string)
	for _, issue := range issues {
		if issue.resolved {
			continue
		}
		statuses := assetStatuses
		if issue.itemType == "collection" {
			statuses = collectionStatuses
		}
		key := localTypeKey(issue.localID, issue.externalType)
		existingStatus := statuses[key]
		if existingStatus == mappingStatusAmbiguous {
			continue
		}
		statuses[key] = issue.status
	}

	for index := range preview.Collections {
		item := &preview.Collections[index]
		status := collectionStatuses[localTypeKey(item.CollectionID, item.ExternalType)]
		if status != "" {
			item.MappingStatus = status
		}
	}
	for index := range preview.Assets {
		item := &preview.Assets[index]
		status := assetStatuses[localTypeKey(item.AssetID, item.ExternalType)]
		if status != "" {
			item.MappingStatus = status
		}
	}
}

func collectionMappingIssue(mapping models.IntegrationCollectionMapping, status string) *mappingIssue {
	return &mappingIssue{
		itemType:     "collection",
		mappingID:    mapping.Id,
		localID:      mapping.CollectionId,
		externalType: mapping.ExternalType,
		status:       status,
	}
}

func assetMappingIssue(mapping models.IntegrationAssetMapping, status string) *mappingIssue {
	return &mappingIssue{
		itemType:     "asset",
		mappingID:    mapping.Id,
		localID:      mapping.AssetId,
		externalType: mapping.ExternalType,
		status:       status,
	}
}
