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
	issues                       []integrations.MappingIssue
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
		issues:                       []integrations.MappingIssue{},
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
	exactCollectionsByLocalType := make(map[string]string)
	for _, mapping := range existingCollections {
		collection, exists := collectionByID[mapping.ExternalId]
		if !exists {
			staleCollections = append(staleCollections, mapping)
			continue
		}
		result.collectionMappings[collection.ID] = mapping
		exactCollectionsByLocalType[localTypeKey(mapping.CollectionId, collection.Type)] = collection.ID
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
		externalID, exists := exactCollectionsByLocalType[localTypeKey(mapping.CollectionId, mapping.ExternalType)]
		if !exists {
			remainingStaleCollections = append(remainingStaleCollections, mapping)
			continue
		}
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
		result.issues = append(result.issues, collectionMappingIssue(mapping, status, len(candidates)))
	}

	staleAssets := make([]models.IntegrationAssetMapping, 0)
	exactAssetsByLocalType := make(map[string]string)
	for _, mapping := range existingAssets {
		asset, exists := assetByID[mapping.ExternalId]
		if !exists {
			staleAssets = append(staleAssets, mapping)
			continue
		}
		result.assetMappings[asset.ID] = mapping
		exactAssetsByLocalType[localTypeKey(mapping.AssetId, asset.AssetType)] = asset.ID
		status := mappingStatusLinked
		if assetMappingChanged(mapping, asset) {
			status = mappingStatusChanged
		}
		result.assetStatuses[asset.ID] = status
	}
	remainingStaleAssets := make([]models.IntegrationAssetMapping, 0, len(staleAssets))
	for _, mapping := range staleAssets {
		externalID, exists := exactAssetsByLocalType[localTypeKey(mapping.AssetId, mapping.ExternalType)]
		if !exists {
			remainingStaleAssets = append(remainingStaleAssets, mapping)
			continue
		}
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
		result.issues = append(result.issues, assetMappingIssue(mapping, status, len(candidates)))
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

func countMappingIssues(issues []integrations.MappingIssue, status string) int {
	count := 0
	for _, issue := range issues {
		if issue.Status == status {
			count++
		}
	}
	return count
}

func consumeUniqueMappingIssue(
	issues []integrations.MappingIssue,
	itemType string,
	localID string,
	externalType string,
) (integrations.MappingIssue, []integrations.MappingIssue, bool) {
	matchingIndex := -1
	for index, issue := range issues {
		if issue.ItemType != itemType || issue.LocalID != localID {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(issue.ExternalType), strings.TrimSpace(externalType)) {
			continue
		}
		if matchingIndex != -1 {
			return integrations.MappingIssue{}, issues, false
		}
		matchingIndex = index
	}
	if matchingIndex == -1 {
		return integrations.MappingIssue{}, issues, false
	}

	issue := issues[matchingIndex]
	remaining := make([]integrations.MappingIssue, 0, len(issues)-1)
	remaining = append(remaining, issues[:matchingIndex]...)
	remaining = append(remaining, issues[matchingIndex+1:]...)
	return issue, remaining, true
}

func applyMappingIssueStatuses(preview *integrations.SyncPreview) {
	collectionStatuses := make(map[string]string)
	assetStatuses := make(map[string]string)
	for _, issue := range preview.MappingIssues {
		statuses := assetStatuses
		if issue.ItemType == "collection" {
			statuses = collectionStatuses
		}
		existingStatus := statuses[issue.LocalID]
		if existingStatus == mappingStatusAmbiguous {
			continue
		}
		statuses[issue.LocalID] = issue.Status
	}

	for index := range preview.Collections {
		status := collectionStatuses[preview.Collections[index].CollectionID]
		if status != "" {
			preview.Collections[index].MappingStatus = status
		}
	}
	for index := range preview.Assets {
		status := assetStatuses[preview.Assets[index].AssetID]
		if status != "" {
			preview.Assets[index].MappingStatus = status
		}
	}
}

func collectionMappingIssue(mapping models.IntegrationCollectionMapping, status string, candidateCount int) integrations.MappingIssue {
	return integrations.MappingIssue{
		ItemType:       "collection",
		MappingID:      mapping.Id,
		LocalID:        mapping.CollectionId,
		ExternalID:     mapping.ExternalId,
		ExternalName:   mapping.ExternalName,
		ExternalType:   mapping.ExternalType,
		Status:         status,
		CandidateCount: candidateCount,
	}
}

func assetMappingIssue(mapping models.IntegrationAssetMapping, status string, candidateCount int) integrations.MappingIssue {
	return integrations.MappingIssue{
		ItemType:       "asset",
		MappingID:      mapping.Id,
		LocalID:        mapping.AssetId,
		ExternalID:     mapping.ExternalId,
		ExternalName:   mapping.ExternalName,
		ExternalType:   mapping.ExternalType,
		Status:         status,
		CandidateCount: candidateCount,
	}
}
