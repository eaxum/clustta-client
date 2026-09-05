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

// GetCheckpointDependencyReferenceCounts returns active exact-pin followers by checkpoint.
func GetCheckpointDependencyReferenceCounts(tx *sqlx.Tx, assetId string) (map[string]int, error) {
	rows := []struct {
		CheckpointId string `db:"checkpoint_id"`
		Count        int    `db:"reference_count"`
	}{}
	err := tx.Select(&rows, `
		SELECT checkpoint_id, COUNT(*) AS reference_count
		FROM asset_dependency
		WHERE dependency_id = ? AND resolution_mode = ?
		GROUP BY checkpoint_id
	`, assetId, DependencyResolutionPinned)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int, len(rows))
	for _, row := range rows {
		counts[row.CheckpointId] = row.Count
	}
	return counts, nil
}

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
func ValidateDependencySelector(tx *sqlx.Tx, dependencyId, resolutionMode string, checkpointId, assetCheckpointTagId *string) error {
	checkpointId = normalizeSelectorReference(checkpointId)
	assetCheckpointTagId = normalizeSelectorReference(assetCheckpointTagId)

	switch resolutionMode {
	case DependencyResolutionFloating:
		if checkpointId != nil || assetCheckpointTagId != nil {
			return errors.New("floating dependencies cannot reference a checkpoint or tag")
		}
	case DependencyResolutionPinned:
		if checkpointId == nil || assetCheckpointTagId != nil {
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
		if checkpointId != nil || assetCheckpointTagId == nil {
			return errors.New("tagged dependencies require only an asset_checkpoint_tag_id")
		}
		var tagCount int
		err := tx.Get(&tagCount, `
			SELECT COUNT(*)
			FROM asset_checkpoint_tag act
			JOIN asset_checkpoint ac ON ac.id = act.checkpoint_id
			WHERE act.id = ?
				AND act.asset_id = ?
				AND ac.trashed = 0
		`, *assetCheckpointTagId, dependencyId)
		if err != nil {
			return err
		}
		if tagCount == 0 {
			return errors.New("tag does not belong to the dependency asset")
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
	dependencyAssetIds, err := resolveReachableDependencyAssets(tx, dependencyId)
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
	checkpointId, assetCheckpointTagId *string,
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
	if err := ValidateDependencySelector(tx, dependencyId, resolutionMode, checkpointId, assetCheckpointTagId); err != nil {
		return assetDependency, err
	}

	checkpointId = normalizeSelectorReference(checkpointId)
	assetCheckpointTagId = normalizeSelectorReference(assetCheckpointTagId)
	params := map[string]any{
		"id":                      id,
		"asset_id":                assetId,
		"dependency_id":           dependencyId,
		"dependency_type_id":      dependencyTypeId,
		"resolution_mode":         resolutionMode,
		"checkpoint_id":           checkpointId,
		"asset_checkpoint_tag_id": assetCheckpointTagId,
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
					FROM asset_checkpoint_tag act
					JOIN asset_checkpoint ac ON ac.id = act.checkpoint_id
					WHERE act.id = ad.asset_checkpoint_tag_id
						AND act.asset_id = ad.dependency_id
						AND ac.trashed = 0
				)
			END AS resolved_checkpoint_id,
			COALESCE((
				SELECT t.name
				FROM asset_checkpoint_tag act
				JOIN tag t ON t.id = act.tag_id
				WHERE act.id = ad.asset_checkpoint_tag_id
			), '') AS tag_name
		FROM asset_dependency ad
	)
	SELECT
		de.*,
		COALESCE(ac.comment, '') AS resolved_checkpoint_label,
		CASE
			WHEN de.resolved_checkpoint_id IS NOT NULL THEN 'ready'
			WHEN de.resolution_mode = 'tagged' AND de.tag_name = '' THEN 'missing_tag'
			WHEN de.resolution_mode = 'tagged' THEN 'tag_checkpoint_missing'
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
	checkpointId, assetCheckpointTagId *string,
) (models.AssetDependencyEdge, error) {
	dependency, err := GetDependency(tx, edgeId)
	if err != nil {
		return models.AssetDependencyEdge{}, err
	}
	if err = ValidateDependencySelector(tx, dependency.DependencyId, resolutionMode, checkpointId, assetCheckpointTagId); err != nil {
		return models.AssetDependencyEdge{}, err
	}

	checkpointId = normalizeSelectorReference(checkpointId)
	assetCheckpointTagId = normalizeSelectorReference(assetCheckpointTagId)
	now := utils.GetEpochTime()
	_, err = tx.Exec(`
		UPDATE asset_dependency
		SET resolution_mode = ?, checkpoint_id = ?, asset_checkpoint_tag_id = ?,
			mtime = CASE WHEN mtime >= ? THEN mtime + 1 ELSE ? END,
			synced = 0
		WHERE id = ?
	`, resolutionMode, checkpointId, assetCheckpointTagId, now, now, edgeId)
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

// SaveDependency applies a newer dependency edge received through sync.
func SaveDependency(tx *sqlx.Tx, dependency models.AssetDependency) error {
	var existing models.AssetDependency
	err := tx.Get(&existing, "SELECT * FROM asset_dependency WHERE id = ?", dependency.Id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err == nil {
		if existing.AssetId != dependency.AssetId || existing.DependencyId != dependency.DependencyId {
			return errors.New("dependency endpoints cannot change")
		}
		if dependency.MTime <= existing.MTime {
			return nil
		}
		if dependency.ResolutionMode == "" {
			return errors.New("dependency resolution mode is required when updating an existing dependency")
		}
	}
	if dependency.ResolutionMode == "" {
		dependency.ResolutionMode = DependencyResolutionFloating
	}
	var referenceCount int
	if err := tx.Get(&referenceCount, `
		SELECT COUNT(*) FROM asset source, asset target, dependency_type
		WHERE source.id = ? AND target.id = ? AND dependency_type.id = ?
	`, dependency.AssetId, dependency.DependencyId, dependency.DependencyTypeId); err != nil {
		return err
	}
	if referenceCount != 1 {
		return errors.New("dependency requires existing assets and a dependency type")
	}
	dependency.CheckpointId = normalizeSelectorReference(dependency.CheckpointId)
	dependency.AssetCheckpointTagId = normalizeSelectorReference(dependency.AssetCheckpointTagId)
	if err := ValidateDependencySelector(tx, dependency.DependencyId, dependency.ResolutionMode, dependency.CheckpointId, dependency.AssetCheckpointTagId); err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO asset_dependency (
			id, mtime, asset_id, dependency_id, dependency_type_id,
			resolution_mode, checkpoint_id, asset_checkpoint_tag_id, synced
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			mtime = excluded.mtime,
			asset_id = excluded.asset_id,
			dependency_id = excluded.dependency_id,
			dependency_type_id = excluded.dependency_type_id,
			resolution_mode = excluded.resolution_mode,
			checkpoint_id = excluded.checkpoint_id,
			asset_checkpoint_tag_id = excluded.asset_checkpoint_tag_id,
			synced = excluded.synced
		WHERE excluded.mtime > asset_dependency.mtime
	`, dependency.Id, dependency.MTime, dependency.AssetId, dependency.DependencyId,
		dependency.DependencyTypeId, dependency.ResolutionMode, dependency.CheckpointId,
		dependency.AssetCheckpointTagId, dependency.Synced)
	return err
}
