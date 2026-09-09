package repository

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"clustta/internal/error_service"
	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
)

type buildRequirement struct {
	assetId            string
	checkpointId       string
	resolutionMode     string
	dependencyEdgeId   string
	requestedByAssetId string
	path               []string
}

type dependencyBuildResolver struct {
	tx                  *sqlx.Tx
	checkBuildReadiness bool
	requirements        map[string][]buildRequirement
	visitState          map[string]int
	orderedAssets       []string
	conflicts           []models.DependencyBuildConflict
}

// ResolveDependencyBuildPlan freezes the dependency graph to exact checkpoints.
func ResolveDependencyBuildPlan(tx *sqlx.Tx, rootAssetId string) (models.DependencyBuildPlan, error) {
	return resolveDependencyPlan(tx, rootAssetId, true)
}

// ResolveDependencyGraphPlan resolves checkpoints and conflicts without local readiness checks.
func ResolveDependencyGraphPlan(tx *sqlx.Tx, rootAssetId string) (models.DependencyGraphPlan, error) {
	plan, err := resolveDependencyPlan(tx, rootAssetId, false)
	if err != nil {
		return models.DependencyGraphPlan{}, err
	}
	graphPlan := models.DependencyGraphPlan{
		Entries:   make([]models.DependencyGraphPlanEntry, 0, len(plan.Entries)),
		Conflicts: plan.Conflicts,
	}
	for _, entry := range plan.Entries {
		graphPlan.Entries = append(graphPlan.Entries, models.DependencyGraphPlanEntry{
			AssetId:          entry.AssetId,
			CheckpointId:     entry.CheckpointId,
			DependencyEdgeId: entry.DependencyEdgeId,
		})
	}
	return graphPlan, nil
}

func resolveDependencyPlan(tx *sqlx.Tx, rootAssetId string, checkBuildReadiness bool) (models.DependencyBuildPlan, error) {
	plan := models.DependencyBuildPlan{
		RootAssetId: rootAssetId,
		ResolvedAt:  utils.GetEpochTime(),
		Entries:     []models.DependencyBuildPlanEntry{},
		Warnings:    []string{},
		Conflicts:   []models.DependencyBuildConflict{},
	}
	var rootExists bool
	if err := tx.Get(&rootExists, "SELECT EXISTS(SELECT 1 FROM asset WHERE id = ?)", rootAssetId); err != nil {
		return plan, err
	}
	if !rootExists {
		return plan, error_service.ErrAssetNotFound
	}

	resolver := dependencyBuildResolver{
		tx:                  tx,
		checkBuildReadiness: checkBuildReadiness,
		requirements:        map[string][]buildRequirement{},
		visitState:          map[string]int{},
		conflicts:           []models.DependencyBuildConflict{},
	}
	rootCheckpointId, err := resolver.latestCheckpointId(rootAssetId)
	if err != nil {
		return plan, err
	}
	resolver.addRequirement(buildRequirement{
		assetId:        rootAssetId,
		checkpointId:   rootCheckpointId,
		resolutionMode: DependencyResolutionFloating,
		path:           []string{rootAssetId},
	})
	if err = resolver.walkAsset(rootAssetId, []string{rootAssetId}); err != nil {
		return plan, err
	}

	plan.Conflicts = append(plan.Conflicts, resolver.conflicts...)
	for _, assetId := range resolver.orderedAssets {
		entry, conflicts, err := resolver.resolveAssetRequirements(assetId)
		if err != nil {
			return plan, err
		}
		plan.Conflicts = append(plan.Conflicts, conflicts...)
		if entry.CheckpointId != "" {
			plan.Entries = append(plan.Entries, entry)
		}
	}
	plan.Fingerprint = buildPlanFingerprint(plan)
	return plan, nil
}

func (r *dependencyBuildResolver) walkAsset(assetId string, path []string) error {
	if r.visitState[assetId] == 1 {
		r.conflicts = append(r.conflicts, models.DependencyBuildConflict{
			AssetId: assetId,
			Paths:   [][]string{append([]string{}, path...)},
			Message: fmt.Sprintf("dependency cycle reaches asset %s", assetId),
		})
		return nil
	}
	if r.visitState[assetId] == 2 {
		return nil
	}
	r.visitState[assetId] = 1

	edges, err := GetAssetDependencyEdges(r.tx, assetId)
	if err != nil {
		return err
	}
	for _, edge := range edges {
		dependencyPath := appendPath(path, edge.DependencyId)
		checkpointId := ""
		if edge.ResolvedCheckpointId != nil {
			checkpointId = *edge.ResolvedCheckpointId
		}
		r.addRequirement(buildRequirement{
			assetId:            edge.DependencyId,
			checkpointId:       checkpointId,
			resolutionMode:     edge.ResolutionMode,
			dependencyEdgeId:   edge.Id,
			requestedByAssetId: assetId,
			path:               dependencyPath,
		})
		if edge.ResolutionStatus != "ready" {
			r.conflicts = append(r.conflicts, models.DependencyBuildConflict{
				AssetId: edge.DependencyId,
				Paths:   [][]string{dependencyPath},
				Message: fmt.Sprintf("dependency edge %s cannot resolve: %s", edge.Id, edge.ResolutionStatus),
			})
		}
		if err = r.walkAsset(edge.DependencyId, dependencyPath); err != nil {
			return err
		}
	}

	collectionIds := []string{}
	if err = r.tx.Select(&collectionIds, `
		SELECT cd.dependency_id
		FROM collection_dependency cd
		JOIN collection c ON c.id = cd.dependency_id
		WHERE cd.asset_id = ? AND c.trashed = 0
		ORDER BY cd.id
	`, assetId); err != nil {
		return err
	}
	for _, collectionId := range collectionIds {
		if err = r.walkCollection(assetId, collectionId, path, map[string]bool{}); err != nil {
			return err
		}
	}

	r.visitState[assetId] = 2
	r.orderedAssets = append(r.orderedAssets, assetId)
	return nil
}

func (r *dependencyBuildResolver) walkCollection(ownerAssetId, collectionId string, path []string, visited map[string]bool) error {
	if visited[collectionId] {
		return nil
	}
	visited[collectionId] = true

	assetIds := []string{}
	if err := r.tx.Select(&assetIds, `
		SELECT id FROM asset
		WHERE collection_id = ? AND trashed = 0
		ORDER BY id
	`, collectionId); err != nil {
		return err
	}
	for _, dependencyAssetId := range assetIds {
		dependencyPath := appendPath(path, dependencyAssetId)
		checkpointId, err := r.latestCheckpointId(dependencyAssetId)
		if err != nil {
			return err
		}
		r.addRequirement(buildRequirement{
			assetId:            dependencyAssetId,
			checkpointId:       checkpointId,
			resolutionMode:     DependencyResolutionFloating,
			requestedByAssetId: ownerAssetId,
			path:               dependencyPath,
		})
		if err = r.walkAsset(dependencyAssetId, dependencyPath); err != nil {
			return err
		}
	}

	childCollectionIds := []string{}
	if err := r.tx.Select(&childCollectionIds, `
		SELECT id FROM collection
		WHERE parent_id = ? AND trashed = 0
		ORDER BY id
	`, collectionId); err != nil {
		return err
	}
	for _, childCollectionId := range childCollectionIds {
		if err := r.walkCollection(ownerAssetId, childCollectionId, path, visited); err != nil {
			return err
		}
	}
	return nil
}

func (r *dependencyBuildResolver) resolveAssetRequirements(assetId string) (models.DependencyBuildPlanEntry, []models.DependencyBuildConflict, error) {
	requirements := r.requirements[assetId]
	conflicts := []models.DependencyBuildConflict{}
	validRequirements := []buildRequirement{}
	for _, requirement := range requirements {
		if requirement.checkpointId != "" {
			validRequirements = append(validRequirements, requirement)
		}
	}
	if len(validRequirements) == 0 {
		for _, requirement := range requirements {
			if requirement.dependencyEdgeId != "" {
				return models.DependencyBuildPlanEntry{}, conflicts, nil
			}
		}
		paths := make([][]string, 0, len(requirements))
		for _, requirement := range requirements {
			paths = append(paths, requirement.path)
		}
		conflicts = append(conflicts, models.DependencyBuildConflict{
			AssetId: assetId,
			Paths:   paths,
			Message: fmt.Sprintf("asset %s has no active checkpoint", assetId),
		})
		return models.DependencyBuildPlanEntry{}, conflicts, nil
	}

	selected := validRequirements[0]
	exactCheckpoints := map[string]bool{}
	for _, requirement := range validRequirements {
		if requirement.resolutionMode == DependencyResolutionFloating {
			continue
		}
		exactCheckpoints[requirement.checkpointId] = true
		selected = requirement
	}
	if len(exactCheckpoints) > 1 {
		checkpointIds := make([]string, 0, len(exactCheckpoints))
		paths := make([][]string, 0, len(validRequirements))
		for checkpointId := range exactCheckpoints {
			checkpointIds = append(checkpointIds, checkpointId)
		}
		for _, requirement := range validRequirements {
			if requirement.resolutionMode != DependencyResolutionFloating {
				paths = append(paths, requirement.path)
			}
		}
		sort.Strings(checkpointIds)
		conflicts = append(conflicts, models.DependencyBuildConflict{
			AssetId:       assetId,
			CheckpointIds: checkpointIds,
			Paths:         paths,
			Message:       fmt.Sprintf("asset %s resolves to incompatible checkpoints", assetId),
		})
		return models.DependencyBuildPlanEntry{}, conflicts, nil
	}

	checkpoint := models.Checkpoint{}
	err := r.tx.Get(&checkpoint, `
		SELECT id, asset_id, trashed FROM asset_checkpoint WHERE id = ?
	`, selected.checkpointId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return models.DependencyBuildPlanEntry{}, nil, err
	}
	if errors.Is(err, sql.ErrNoRows) || checkpoint.Trashed || checkpoint.AssetId != assetId {
		conflicts = append(conflicts, models.DependencyBuildConflict{
			AssetId:       assetId,
			CheckpointIds: []string{selected.checkpointId},
			Paths:         [][]string{selected.path},
			Message:       fmt.Sprintf("checkpoint %s is not active for asset %s", selected.checkpointId, assetId),
		})
		return models.DependencyBuildPlanEntry{}, conflicts, nil
	}
	entry := models.DependencyBuildPlanEntry{
		AssetId:            assetId,
		CheckpointId:       selected.checkpointId,
		ResolutionMode:     selected.resolutionMode,
		DependencyEdgeId:   selected.dependencyEdgeId,
		RequestedByAssetId: selected.requestedByAssetId,
		ResolutionPath:     selected.path,
	}
	if !r.checkBuildReadiness {
		return entry, conflicts, nil
	}
	if err = r.tx.Get(&checkpoint.Chunks, "SELECT chunks FROM asset_checkpoint WHERE id = ?", checkpoint.Id); err != nil {
		return models.DependencyBuildPlanEntry{}, nil, err
	}
	entry.MissingChunks, err = checkpoint.HasMissingChunks(r.tx)
	if err != nil {
		return models.DependencyBuildPlanEntry{}, nil, err
	}
	entry.FileStatus, err = GetAssetState(r.tx, assetId)
	if err != nil {
		return models.DependencyBuildPlanEntry{}, nil, err
	}
	entry.RequiresOverwrite = entry.FileStatus == "modified"
	return entry, conflicts, nil
}

func (r *dependencyBuildResolver) latestCheckpointId(assetId string) (string, error) {
	var checkpointId string
	err := r.tx.Get(&checkpointId, `
		SELECT id FROM asset_checkpoint
		WHERE asset_id = ? AND trashed = 0 ORDER BY created_at DESC LIMIT 1
	`, assetId)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return checkpointId, err
}

func (r *dependencyBuildResolver) addRequirement(requirement buildRequirement) {
	r.requirements[requirement.assetId] = append(r.requirements[requirement.assetId], requirement)
}

func appendPath(path []string, assetId string) []string {
	result := append([]string{}, path...)
	return append(result, assetId)
}

func buildPlanFingerprint(plan models.DependencyBuildPlan) string {
	parts := []string{plan.RootAssetId}
	for _, entry := range plan.Entries {
		parts = append(parts, strings.Join([]string{
			entry.AssetId,
			entry.CheckpointId,
			entry.ResolutionMode,
			entry.DependencyEdgeId,
			entry.RequestedByAssetId,
			strings.Join(entry.ResolutionPath, ">"),
		}, "|"))
	}
	for _, conflict := range plan.Conflicts {
		parts = append(parts, conflict.AssetId+"|"+strings.Join(conflict.CheckpointIds, ",")+"|"+conflict.Message)
	}
	digest := sha256.Sum256([]byte(strings.Join(parts, "\n")))
	return hex.EncodeToString(digest[:])
}

func resolveReachableDependencyAssets(tx *sqlx.Tx, rootAssetId string) ([]string, error) {
	visitedAssets := map[string]bool{}
	visitedCollections := map[string]bool{}
	result := []string{}
	assetQueue := []string{rootAssetId}
	visitedAssets[rootAssetId] = true

	for len(assetQueue) > 0 {
		assetId := assetQueue[0]
		assetQueue = assetQueue[1:]
		result = append(result, assetId)

		dependencyAssetIds := []string{}
		if err := tx.Select(&dependencyAssetIds, `
			SELECT ad.dependency_id
			FROM asset_dependency ad
			JOIN asset a ON a.id = ad.dependency_id
			WHERE ad.asset_id = ? AND a.trashed = 0
			ORDER BY ad.id
		`, assetId); err != nil {
			return nil, err
		}
		for _, dependencyAssetId := range dependencyAssetIds {
			if !visitedAssets[dependencyAssetId] {
				visitedAssets[dependencyAssetId] = true
				assetQueue = append(assetQueue, dependencyAssetId)
			}
		}

		collectionIds := []string{}
		if err := tx.Select(&collectionIds, `
			SELECT cd.dependency_id
			FROM collection_dependency cd
			JOIN collection c ON c.id = cd.dependency_id
			WHERE cd.asset_id = ? AND c.trashed = 0
			ORDER BY cd.id
		`, assetId); err != nil {
			return nil, err
		}
		for len(collectionIds) > 0 {
			collectionId := collectionIds[0]
			collectionIds = collectionIds[1:]
			if visitedCollections[collectionId] {
				continue
			}
			visitedCollections[collectionId] = true

			collectionAssetIds := []string{}
			if err := tx.Select(&collectionAssetIds, `
				SELECT id FROM asset
				WHERE collection_id = ? AND trashed = 0
				ORDER BY id
			`, collectionId); err != nil {
				return nil, err
			}
			for _, dependencyAssetId := range collectionAssetIds {
				if !visitedAssets[dependencyAssetId] {
					visitedAssets[dependencyAssetId] = true
					assetQueue = append(assetQueue, dependencyAssetId)
				}
			}

			childCollectionIds := []string{}
			if err := tx.Select(&childCollectionIds, `
				SELECT id FROM collection
				WHERE parent_id = ? AND trashed = 0
				ORDER BY id
			`, collectionId); err != nil {
				return nil, err
			}
			collectionIds = append(collectionIds, childCollectionIds...)
		}
	}
	return result, nil
}

// ResolveBuildDependencies returns dependency-first asset IDs from an exact plan.
func ResolveBuildDependencies(tx *sqlx.Tx, rootAssetId string) ([]string, error) {
	plan, err := ResolveDependencyBuildPlan(tx, rootAssetId)
	if err != nil {
		return nil, err
	}
	assetIds := make([]string, 0, len(plan.Entries))
	for _, entry := range plan.Entries {
		assetIds = append(assetIds, entry.AssetId)
	}
	return assetIds, nil
}
