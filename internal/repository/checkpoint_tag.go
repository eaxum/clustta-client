package repository

import (
	"database/sql"
	"errors"
	"strings"

	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// GetCheckpointTag retrieves a checkpoint tag by ID.
func GetCheckpointTag(tx *sqlx.Tx, tagId string) (models.CheckpointTag, error) {
	tag := models.CheckpointTag{}
	err := tx.Get(&tag, "SELECT * FROM checkpoint_tag WHERE id = ?", tagId)
	if errors.Is(err, sql.ErrNoRows) {
		return tag, errors.New("checkpoint tag not found")
	}
	return tag, err
}

// GetCheckpointTagsForAsset returns the moving tags defined for an asset.
func GetCheckpointTagsForAsset(tx *sqlx.Tx, assetId string) ([]models.CheckpointTag, error) {
	tags := []models.CheckpointTag{}
	err := tx.Select(&tags, `
		SELECT checkpoint_tag.*
		FROM checkpoint_tag
		JOIN asset_checkpoint ON asset_checkpoint.id = checkpoint_tag.checkpoint_id
		WHERE checkpoint_tag.asset_id = ? AND asset_checkpoint.trashed = 0
		ORDER BY checkpoint_tag.name COLLATE NOCASE
	`, assetId)
	return tags, err
}

// SetCheckpointTag creates, renames, or moves an asset-scoped tag.
func SetCheckpointTag(tx *sqlx.Tx, tagId, name, checkpointId string) (models.CheckpointTag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.CheckpointTag{}, errors.New("checkpoint tag name cannot be empty")
	}

	checkpoint := models.Checkpoint{}
	err := tx.Get(&checkpoint, `
		SELECT * FROM asset_checkpoint
		WHERE id = ? AND trashed = 0
	`, checkpointId)
	if errors.Is(err, sql.ErrNoRows) {
		return models.CheckpointTag{}, errors.New("checkpoint not found")
	}
	if err != nil {
		return models.CheckpointTag{}, err
	}

	if tagId != "" {
		existingTag := models.CheckpointTag{}
		getErr := tx.Get(&existingTag, "SELECT * FROM checkpoint_tag WHERE id = ?", tagId)
		if getErr != nil && !errors.Is(getErr, sql.ErrNoRows) {
			return models.CheckpointTag{}, getErr
		}
		if getErr == nil && existingTag.AssetId != checkpoint.AssetId {
			return models.CheckpointTag{}, errors.New("checkpoint tag belongs to another asset")
		}
	} else {
		err = tx.Get(&tagId, `
			SELECT id FROM checkpoint_tag
			WHERE asset_id = ? AND name = ? COLLATE NOCASE
		`, checkpoint.AssetId, name)
		if errors.Is(err, sql.ErrNoRows) {
			tagId = uuid.New().String()
		} else if err != nil {
			return models.CheckpointTag{}, err
		}
	}

	now := utils.GetEpochTime()
	_, err = tx.Exec(`
		INSERT INTO checkpoint_tag (id, mtime, name, asset_id, checkpoint_id, synced)
		VALUES (?, ?, ?, ?, ?, 0)
		ON CONFLICT(id) DO UPDATE SET
			mtime = CASE
				WHEN checkpoint_tag.mtime >= excluded.mtime THEN checkpoint_tag.mtime + 1
				ELSE excluded.mtime
			END,
			name = excluded.name,
			checkpoint_id = excluded.checkpoint_id,
			synced = 0
	`, tagId, now, name, checkpoint.AssetId, checkpointId)
	if err != nil {
		return models.CheckpointTag{}, err
	}
	return GetCheckpointTag(tx, tagId)
}

// SetCheckpointTagsForGroup applies one tag name to every active group member.
func SetCheckpointTagsForGroup(tx *sqlx.Tx, name, groupId string) ([]models.CheckpointTag, error) {
	group, err := GetCheckpointGroup(tx, groupId)
	if err != nil {
		return nil, err
	}
	if !group.Finalized {
		return nil, errors.New("checkpoint group is not finalized")
	}

	checkpointIds := []string{}
	if err = tx.Select(&checkpointIds, `
		SELECT id FROM asset_checkpoint
		WHERE group_id = ? AND trashed = 0
		ORDER BY asset_id
	`, groupId); err != nil {
		return nil, err
	}
	if len(checkpointIds) == 0 {
		return nil, errors.New("checkpoint group has no active checkpoints")
	}

	tags := make([]models.CheckpointTag, 0, len(checkpointIds))
	for _, checkpointId := range checkpointIds {
		tag, setErr := SetCheckpointTag(tx, "", name, checkpointId)
		if setErr != nil {
			return nil, setErr
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// DeleteCheckpointTag removes an unreferenced checkpoint tag.
func DeleteCheckpointTag(tx *sqlx.Tx, tagId string) error {
	var referenceCount int
	if err := tx.Get(&referenceCount, "SELECT COUNT(*) FROM asset_dependency WHERE checkpoint_tag_id = ?", tagId); err != nil {
		return err
	}
	if referenceCount > 0 {
		return errors.New("checkpoint tag is referenced by a dependency")
	}

	result, err := tx.Exec("DELETE FROM checkpoint_tag WHERE id = ?", tagId)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("checkpoint tag not found")
	}
	return nil
}
