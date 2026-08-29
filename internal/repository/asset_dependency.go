package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"clustta/internal/base_service"
	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/jmoiron/sqlx"
)

const (
	DependencyResolutionFloating = "floating"
	DependencyResolutionPinned   = "pinned"
	DependencyResolutionTagged   = "tagged"
)

func normalizeSelectorReference(reference *string) *string {
	if reference == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*reference)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// ValidateDependencySelector validates selector shape and asset ownership.
func ValidateDependencySelector(tx *sqlx.Tx, dependencyId, resolutionMode string, checkpointId, checkpointGroupTagId *string) error {
	checkpointId = normalizeSelectorReference(checkpointId)
	checkpointGroupTagId = normalizeSelectorReference(checkpointGroupTagId)

	switch resolutionMode {
	case DependencyResolutionFloating:
		if checkpointId != nil || checkpointGroupTagId != nil {
			return errors.New("floating dependencies cannot reference a checkpoint or tag")
		}
	case DependencyResolutionPinned:
		if checkpointId == nil || checkpointGroupTagId != nil {
			return errors.New("pinned dependencies require only a checkpoint_id")
		}
		var checkpointCount int
		err := tx.Get(&checkpointCount, `
			SELECT COUNT(*) FROM asset_checkpoint
			WHERE id = ? AND asset_id = ? AND trashed = 0
		`, *checkpointId, dependencyId)
		if err != nil {
			return err
		}
		if checkpointCount == 0 {
			return errors.New("checkpoint does not belong to the dependency asset")
		}
	case DependencyResolutionTagged:
		if checkpointId != nil || checkpointGroupTagId == nil {
			return errors.New("tagged dependencies require only a checkpoint_group_tag_id")
		}
		var tagCount int
		err := tx.Get(&tagCount, `
			SELECT COUNT(*)
			FROM checkpoint_group_tag cgt
			JOIN checkpoint_group cg ON cg.id = cgt.group_id
			JOIN asset_checkpoint ac ON ac.group_id = cg.id
			WHERE cgt.id = ?
				AND cg.finalized = 1
				AND cg.group_type = ?
				AND ac.asset_id = ?
				AND ac.trashed = 0
		`, *checkpointGroupTagId, CheckpointGroupTypeMulti, dependencyId)
		if err != nil {
			return err
		}
		if tagCount == 0 {
			return errors.New("tag does not contain the dependency asset")
		}
	default:
		return fmt.Errorf("invalid dependency resolution mode: %s", resolutionMode)
	}
	return nil
}

func validateDependencyCycle(tx *sqlx.Tx, assetId, dependencyId string) error {
	if assetId == dependencyId {
		return errors.New("an asset cannot depend on itself")
	}
	var directCycleCount int
	err := tx.Get(&directCycleCount, `
		WITH RECURSIVE reachable(id) AS (
			SELECT ?
			UNION
			SELECT ad.dependency_id
			FROM asset_dependency ad
			JOIN reachable r ON r.id = ad.asset_id
		)
		SELECT COUNT(*) FROM reachable WHERE id = ?
	`, dependencyId, assetId)
	if err != nil {
		return err
	}
	if directCycleCount > 0 {
		return errors.New("dependency would create a cycle")
	}
	dependencyAssetIds, err := ResolveBuildDependencies(tx, dependencyId)
	if err != nil {
		return err
	}
	for _, dependencyAssetId := range dependencyAssetIds {
		if dependencyAssetId == assetId {
			return errors.New("dependency would create a cycle")
		}
	}
	return nil
}

func validateDependencyAssets(tx *sqlx.Tx, assetId, dependencyId string) error {
	var assetCount int
	err := tx.Get(&assetCount, `
		SELECT COUNT(*) FROM asset
		WHERE id IN (?, ?) AND trashed = 0
	`, assetId, dependencyId)
	if err != nil {
		return err
	}
	if assetCount != 2 {
		return errors.New("dependency assets must exist and be active")
	}
	return nil
}

// AddDependency creates a floating dependency edge.
func AddDependency(tx *sqlx.Tx, id, assetId, dependencyId, dependencyTypeId string) (models.AssetDependencyEdge, error) {
	return AddDependencyWithSelector(
		tx,
		id,
		assetId,
		dependencyId,
		dependencyTypeId,
		DependencyResolutionFloating,
		nil,
		nil,
	)
}

// AddDependencyWithSelector creates a validated versioned dependency edge.
func AddDependencyWithSelector(
	tx *sqlx.Tx,
	id, assetId, dependencyId, dependencyTypeId, resolutionMode string,
	checkpointId, checkpointGroupTagId *string,
) (models.AssetDependencyEdge, error) {
	assetDependency := models.AssetDependencyEdge{}
	if err := validateDependencyAssets(tx, assetId, dependencyId); err != nil {
		return assetDependency, err
	}
	if _, err := GetDependencyType(tx, dependencyTypeId); err != nil {
		return assetDependency, err
	}
	if err := validateDependencyCycle(tx, assetId, dependencyId); err != nil {
		return assetDependency, err
	}
	if err := ValidateDependencySelector(tx, dependencyId, resolutionMode, checkpointId, checkpointGroupTagId); err != nil {
		return assetDependency, err
	}

	checkpointId = normalizeSelectorReference(checkpointId)
	checkpointGroupTagId = normalizeSelectorReference(checkpointGroupTagId)
	params := map[string]any{
		"id":                      id,
		"asset_id":                assetId,
		"dependency_id":           dependencyId,
		"dependency_type_id":      dependencyTypeId,
		"resolution_mode":         resolutionMode,
		"checkpoint_id":           checkpointId,
		"checkpoint_group_tag_id": checkpointGroupTagId,
	}
	if err := base_service.Create(tx, "asset_dependency", params); err != nil {
		return assetDependency, err
	}
	edges, err := GetAssetDependencyEdges(tx, assetId)
	if err != nil {
		return assetDependency, err
	}
	for _, edge := range edges {
		if edge.DependencyId == dependencyId {
			return edge, nil
		}
	}
	return assetDependency, errors.New("dependency edge not found after creation")
}

func GetDependency(tx *sqlx.Tx, id string) (models.AssetDependency, error) {
	dependency := models.AssetDependency{}
	if err := base_service.Get(tx, "asset_dependency", id, &dependency); err != nil {
		return dependency, err
	}
	return dependency, nil
}

func GetAssetDependencies(tx *sqlx.Tx, assetId string) ([]models.AssetDependencyEdge, error) {
	return GetAssetDependencyEdges(tx, assetId)
}

const dependencyEdgeSelect = `
	WITH dependency_edges AS (
		SELECT
			ad.*,
			CASE ad.resolution_mode
				WHEN 'floating' THEN (
					SELECT ac.id FROM asset_checkpoint ac
					WHERE ac.asset_id = ad.dependency_id AND ac.trashed = 0
					ORDER BY ac.created_at DESC LIMIT 1
				)
				WHEN 'pinned' THEN (
					SELECT ac.id FROM asset_checkpoint ac
					WHERE ac.id = ad.checkpoint_id AND ac.asset_id = ad.dependency_id AND ac.trashed = 0
				)
				WHEN 'tagged' THEN (
					SELECT ac.id
					FROM checkpoint_group_tag cgt
					JOIN checkpoint_group cg ON cg.id = cgt.group_id AND cg.finalized = 1
					JOIN asset_checkpoint ac ON ac.group_id = cg.id
					WHERE cgt.id = ad.checkpoint_group_tag_id
						AND ac.asset_id = ad.dependency_id
						AND ac.trashed = 0
					LIMIT 1
				)
			END AS resolved_checkpoint_id,
			COALESCE((
				SELECT cgt.name FROM checkpoint_group_tag cgt
				WHERE cgt.id = ad.checkpoint_group_tag_id
			), '') AS tag_name
		FROM asset_dependency ad
	)
	SELECT
		de.*,
		COALESCE(ac.comment, '') AS resolved_checkpoint_label,
		CASE
			WHEN de.resolved_checkpoint_id IS NOT NULL THEN 'ready'
			WHEN de.resolution_mode = 'tagged' AND de.tag_name = '' THEN 'missing_tag'
			WHEN de.resolution_mode = 'tagged' THEN 'tag_asset_missing'
			ELSE 'missing_checkpoint'
		END AS resolution_status
	FROM dependency_edges de
	LEFT JOIN asset_checkpoint ac ON ac.id = de.resolved_checkpoint_id
`

// GetAssetDependencyEdges returns selector-aware edges owned by an asset.
func GetAssetDependencyEdges(tx *sqlx.Tx, assetId string) ([]models.AssetDependencyEdge, error) {
	edges := []models.AssetDependencyEdge{}
	query := dependencyEdgeSelect + " WHERE de.asset_id = ? ORDER BY de.id"
	if err := tx.Select(&edges, query, assetId); err != nil {
		return edges, err
	}
	return edges, nil
}

// GetAssetDependencyEdge returns one selector-aware edge.
func GetAssetDependencyEdge(tx *sqlx.Tx, edgeId string) (models.AssetDependencyEdge, error) {
	edge := models.AssetDependencyEdge{}
	query := dependencyEdgeSelect + " WHERE de.id = ?"
	err := tx.Get(&edge, query, edgeId)
	if errors.Is(err, sql.ErrNoRows) {
		return edge, errors.New("dependency edge not found")
	}
	return edge, err
}

// UpdateDependencySelector changes only the version selector on an edge.
func UpdateDependencySelector(
	tx *sqlx.Tx,
	edgeId, resolutionMode string,
	checkpointId, checkpointGroupTagId *string,
) (models.AssetDependencyEdge, error) {
	dependency, err := GetDependency(tx, edgeId)
	if err != nil {
		return models.AssetDependencyEdge{}, err
	}
	if err = ValidateDependencySelector(tx, dependency.DependencyId, resolutionMode, checkpointId, checkpointGroupTagId); err != nil {
		return models.AssetDependencyEdge{}, err
	}

	checkpointId = normalizeSelectorReference(checkpointId)
	checkpointGroupTagId = normalizeSelectorReference(checkpointGroupTagId)
	now := utils.GetEpochTime()
	_, err = tx.Exec(`
		UPDATE asset_dependency
		SET resolution_mode = ?, checkpoint_id = ?, checkpoint_group_tag_id = ?,
			mtime = CASE WHEN mtime >= ? THEN mtime + 1 ELSE ? END,
			synced = 0
		WHERE id = ?
	`, resolutionMode, checkpointId, checkpointGroupTagId, now, now, edgeId)
	if err != nil {
		return models.AssetDependencyEdge{}, err
	}
	return GetAssetDependencyEdge(tx, edgeId)
}

func RemoveAssetDependency(tx *sqlx.Tx, assetId, dependencyId string) error {
	assetDependency := models.AssetDependency{}
	conditions := map[string]interface{}{
		"asset_id":      assetId,
		"dependency_id": dependencyId,
	}
	if err := base_service.DeleteBy(tx, "asset_dependency", conditions); err != nil {
		return err
	}
	err := base_service.GetBy(tx, "asset_dependency", conditions, &assetDependency)
	if err == nil {
		return errors.New("dependency failed to remove")
	}
	if err != sql.ErrNoRows {
		return err
	}
	return nil
}
