package repository

import (
	"database/sql"
	"errors"
	"strings"

	"clustta/internal/base_service"
	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// GetAssetCheckpointTag retrieves one checkpoint tag assignment.
func GetAssetCheckpointTag(tx *sqlx.Tx, assignmentId string) (models.AssetCheckpointTag, error) {
	assignment := models.AssetCheckpointTag{}
	err := tx.Get(&assignment, `
		SELECT act.*, t.name
		FROM asset_checkpoint_tag act
		JOIN tag t ON t.id = act.tag_id
		WHERE act.id = ?
	`, assignmentId)
	if errors.Is(err, sql.ErrNoRows) {
		return assignment, errors.New("checkpoint tag assignment not found")
	}
	return assignment, err
}

// GetCheckpointTagsForAsset returns active checkpoint tag assignments for an asset.
func GetCheckpointTagsForAsset(tx *sqlx.Tx, assetId string) ([]models.AssetCheckpointTag, error) {
	assignments := []models.AssetCheckpointTag{}
	err := tx.Select(&assignments, `
		SELECT act.*, t.name
		FROM asset_checkpoint_tag act
		JOIN tag t ON t.id = act.tag_id
		JOIN asset_checkpoint ac ON ac.id = act.checkpoint_id
		WHERE act.asset_id = ? AND ac.trashed = 0
		ORDER BY t.name COLLATE NOCASE
	`, assetId)
	return assignments, err
}

func resolveCheckpointTag(tx *sqlx.Tx, tagId, name string) (models.Tag, error) {
	if tagId != "" {
		return GetTag(tx, tagId)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return models.Tag{}, errors.New("checkpoint tag name cannot be empty")
	}
	return GetOrCreateTag(tx, name)
}

func ensureAssetTag(tx *sqlx.Tx, assetId string, tag models.Tag) error {
	var assignmentCount int
	if err := tx.Get(&assignmentCount, `
		SELECT COUNT(*) FROM asset_tag WHERE asset_id = ? AND tag_id = ?
	`, assetId, tag.Id); err != nil {
		return err
	}
	if assignmentCount > 0 {
		return nil
	}
	return AddTagToAssetById(tx, uuid.New().String(), assetId, tag.Id)
}

// SetCheckpointTag creates or moves a project tag assignment within an asset.
func SetCheckpointTag(tx *sqlx.Tx, tagId, name, checkpointId string) (models.AssetCheckpointTag, error) {
	checkpoint := models.Checkpoint{}
	err := tx.Get(&checkpoint, `
		SELECT * FROM asset_checkpoint
		WHERE id = ? AND trashed = 0
	`, checkpointId)
	if errors.Is(err, sql.ErrNoRows) {
		return models.AssetCheckpointTag{}, errors.New("checkpoint not found")
	}
	if err != nil {
		return models.AssetCheckpointTag{}, err
	}

	tag, err := resolveCheckpointTag(tx, tagId, name)
	if err != nil {
		return models.AssetCheckpointTag{}, err
	}
	if err = ensureAssetTag(tx, checkpoint.AssetId, tag); err != nil {
		return models.AssetCheckpointTag{}, err
	}

	assignmentId := ""
	err = tx.Get(&assignmentId, `
		SELECT id FROM asset_checkpoint_tag
		WHERE asset_id = ? AND tag_id = ?
	`, checkpoint.AssetId, tag.Id)
	if errors.Is(err, sql.ErrNoRows) {
		assignmentId = uuid.New().String()
	} else if err != nil {
		return models.AssetCheckpointTag{}, err
	}

	now := utils.GetEpochTime()
	_, err = tx.Exec(`
		INSERT INTO asset_checkpoint_tag (id, mtime, asset_id, tag_id, checkpoint_id, synced)
		VALUES (?, ?, ?, ?, ?, 0)
		ON CONFLICT(asset_id, tag_id) DO UPDATE SET
			mtime = CASE
				WHEN asset_checkpoint_tag.mtime >= excluded.mtime THEN asset_checkpoint_tag.mtime + 1
				ELSE excluded.mtime
			END,
			checkpoint_id = excluded.checkpoint_id,
			synced = 0
	`, assignmentId, now, checkpoint.AssetId, tag.Id, checkpointId)
	if err != nil {
		return models.AssetCheckpointTag{}, err
	}
	return GetAssetCheckpointTag(tx, assignmentId)
}

// SetCheckpointTagsForGroup applies one project tag to each asset's latest group checkpoint.
func SetCheckpointTagsForGroup(tx *sqlx.Tx, tagId, name, groupId string) ([]models.AssetCheckpointTag, error) {
	checkpointIds := []string{}
	if err := tx.Select(&checkpointIds, `
		SELECT id FROM (
			SELECT id, asset_id,
				ROW_NUMBER() OVER (PARTITION BY asset_id ORDER BY created_at DESC, id DESC) AS row_number
			FROM asset_checkpoint
			WHERE group_id = ? AND trashed = 0
		)
		WHERE row_number = 1
		ORDER BY asset_id
	`, groupId); err != nil {
		return nil, err
	}
	if len(checkpointIds) == 0 {
		return nil, errors.New("checkpoint group has no active checkpoints")
	}

	assignments := make([]models.AssetCheckpointTag, 0, len(checkpointIds))
	for _, checkpointId := range checkpointIds {
		assignment, err := SetCheckpointTag(tx, tagId, name, checkpointId)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}
	return assignments, nil
}

// DeleteCheckpointTag removes the checkpoint and asset tag assignments.
func DeleteCheckpointTag(tx *sqlx.Tx, assignmentId string) error {
	assignment, err := GetAssetCheckpointTag(tx, assignmentId)
	if err != nil {
		return err
	}
	assetTag := models.AssetTag{}
	assetTagErr := tx.Get(&assetTag, `
		SELECT * FROM asset_tag WHERE asset_id = ? AND tag_id = ?
	`, assignment.AssetId, assignment.TagId)
	if assetTagErr != nil && !errors.Is(assetTagErr, sql.ErrNoRows) {
		return assetTagErr
	}

	var referenceCount int
	if err = tx.Get(&referenceCount, `
		SELECT COUNT(*) FROM asset_dependency WHERE asset_checkpoint_tag_id = ?
	`, assignmentId); err != nil {
		return err
	}
	if referenceCount > 0 {
		return errors.New("checkpoint tag is referenced by a dependency")
	}

	if err = base_service.Delete(tx, "asset_checkpoint_tag", assignmentId); err != nil {
		return err
	}
	if err = RemoveTagFromAsset(tx, assignment.AssetId, assignment.TagId); err != nil {
		return err
	}
	if !assignment.Synced {
		if _, err = tx.Exec("DELETE FROM tomb WHERE id = ? AND table_name = 'asset_checkpoint_tag'", assignment.Id); err != nil {
			return err
		}
	}
	if assetTagErr == nil && !assetTag.Synced {
		_, err = tx.Exec("DELETE FROM tomb WHERE id = ? AND table_name = 'asset_tag'", assetTag.Id)
	}
	return err
}
