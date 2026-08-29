package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"clustta/internal/repository/models"
	"clustta/internal/utils"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	CheckpointGroupTypeSingle = "single"
	CheckpointGroupTypeMulti  = "multi"
)

var ErrCheckpointGroupNotFound = errors.New("checkpoint group not found")

func validateCheckpointGroupType(groupType string) error {
	if groupType != CheckpointGroupTypeSingle && groupType != CheckpointGroupTypeMulti {
		return fmt.Errorf("invalid checkpoint group type: %s", groupType)
	}
	return nil
}

// GetCheckpointGroup retrieves a checkpoint group by ID.
func GetCheckpointGroup(tx *sqlx.Tx, groupId string) (models.CheckpointGroup, error) {
	group := models.CheckpointGroup{}
	err := tx.Get(&group, "SELECT * FROM checkpoint_group WHERE id = ?", groupId)
	if errors.Is(err, sql.ErrNoRows) {
		return group, ErrCheckpointGroupNotFound
	}
	return group, err
}

func createCheckpointGroup(tx *sqlx.Tx, groupId, groupType string) (models.CheckpointGroup, error) {
	now := utils.GetEpochTime()
	_, err := tx.Exec(`
		INSERT INTO checkpoint_group (id, mtime, created_at, group_type, finalized, synced)
		VALUES (?, ?, ?, ?, 0, 0)
	`, groupId, now, now, groupType)
	if err != nil {
		return models.CheckpointGroup{}, err
	}
	return GetCheckpointGroup(tx, groupId)
}

// EnsureCheckpointGroup creates a single checkpoint group when it does not exist.
func EnsureCheckpointGroup(tx *sqlx.Tx, groupId string) (models.CheckpointGroup, bool, error) {
	if strings.TrimSpace(groupId) == "" {
		return models.CheckpointGroup{}, false, errors.New("group_id can't be empty")
	}

	group, err := GetCheckpointGroup(tx, groupId)
	if err == nil {
		return group, false, nil
	}
	if !errors.Is(err, ErrCheckpointGroupNotFound) {
		return models.CheckpointGroup{}, false, err
	}

	group, err = createCheckpointGroup(tx, groupId, CheckpointGroupTypeSingle)
	return group, true, err
}

// BeginCheckpointGroup creates an unfinalized group with an explicit type.
func BeginCheckpointGroup(tx *sqlx.Tx, groupId, groupType string) (models.CheckpointGroup, error) {
	if err := validateCheckpointGroupType(groupType); err != nil {
		return models.CheckpointGroup{}, err
	}
	if strings.TrimSpace(groupId) == "" {
		return models.CheckpointGroup{}, errors.New("group_id can't be empty")
	}

	group, err := GetCheckpointGroup(tx, groupId)
	if err == nil {
		return models.CheckpointGroup{}, fmt.Errorf("checkpoint group already exists: %s", group.Id)
	}
	if !errors.Is(err, ErrCheckpointGroupNotFound) {
		return models.CheckpointGroup{}, err
	}
	return createCheckpointGroup(tx, groupId, groupType)
}

// ShouldAutoFinalizeCheckpointGroup reports whether a caller owns finalization.
func ShouldAutoFinalizeCheckpointGroup(tx *sqlx.Tx, groupId string) (bool, error) {
	group, err := GetCheckpointGroup(tx, groupId)
	if err != nil {
		if errors.Is(err, ErrCheckpointGroupNotFound) {
			return true, nil
		}
		return false, err
	}
	if group.Finalized {
		var tagCount int
		if err = tx.Get(&tagCount, "SELECT COUNT(*) FROM checkpoint_group_tag WHERE group_id = ?", groupId); err != nil {
			return false, err
		}
		if tagCount > 0 {
			return false, errors.New("tagged checkpoint group membership is immutable")
		}
		return true, nil
	}
	return false, nil
}

// FinalizeCheckpointGroup validates members and makes a group tag eligible when appropriate.
func FinalizeCheckpointGroup(tx *sqlx.Tx, groupId string) (models.CheckpointGroup, error) {
	group, err := GetCheckpointGroup(tx, groupId)
	if err != nil {
		return models.CheckpointGroup{}, err
	}

	memberCounts := struct {
		CheckpointCount int `db:"checkpoint_count"`
		AssetCount      int `db:"asset_count"`
	}{}
	err = tx.Get(&memberCounts, `
		SELECT COUNT(*) AS checkpoint_count, COUNT(DISTINCT asset_id) AS asset_count
		FROM asset_checkpoint
		WHERE group_id = ? AND trashed = 0
	`, groupId)
	if err != nil {
		return models.CheckpointGroup{}, err
	}
	if memberCounts.AssetCount == 0 {
		return models.CheckpointGroup{}, errors.New("checkpoint group has no active checkpoints")
	}
	if group.GroupType == CheckpointGroupTypeMulti && memberCounts.AssetCount < 2 {
		return models.CheckpointGroup{}, errors.New("multi checkpoint group requires at least two assets")
	}
	if memberCounts.AssetCount >= 2 && memberCounts.CheckpointCount != memberCounts.AssetCount {
		return models.CheckpointGroup{}, errors.New("multi checkpoint group cannot contain multiple active checkpoints for one asset")
	}

	groupType := CheckpointGroupTypeSingle
	if memberCounts.AssetCount >= 2 {
		groupType = CheckpointGroupTypeMulti
	}
	now := utils.GetEpochTime()
	_, err = tx.Exec(`
		UPDATE checkpoint_group
		SET group_type = ?, finalized = 1,
			mtime = CASE WHEN mtime >= ? THEN mtime + 1 ELSE ? END,
			synced = 0
		WHERE id = ?
	`, groupType, now, now, groupId)
	if err != nil {
		return models.CheckpointGroup{}, err
	}
	return GetCheckpointGroup(tx, groupId)
}

// SetCheckpointGroupTag creates or moves a named tag to an eligible group.
func SetCheckpointGroupTag(tx *sqlx.Tx, tagId, name, groupId string) (models.CheckpointGroupTag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.CheckpointGroupTag{}, errors.New("checkpoint group tag name cannot be empty")
	}

	group, err := GetCheckpointGroup(tx, groupId)
	if err != nil {
		return models.CheckpointGroupTag{}, err
	}
	if !group.Finalized || group.GroupType != CheckpointGroupTypeMulti {
		return models.CheckpointGroupTag{}, errors.New("checkpoint group is not taggable")
	}

	if tagId == "" {
		tagId = uuid.New().String()
	}
	now := utils.GetEpochTime()
	_, err = tx.Exec(`
		INSERT INTO checkpoint_group_tag (id, mtime, name, group_id, synced)
		VALUES (?, ?, ?, ?, 0)
		ON CONFLICT(id) DO UPDATE SET
			mtime = CASE
				WHEN checkpoint_group_tag.mtime >= excluded.mtime THEN checkpoint_group_tag.mtime + 1
				ELSE excluded.mtime
			END,
			name = excluded.name,
			group_id = excluded.group_id,
			synced = 0
	`, tagId, now, name, groupId)
	if err != nil {
		return models.CheckpointGroupTag{}, err
	}

	tag := models.CheckpointGroupTag{}
	err = tx.Get(&tag, "SELECT * FROM checkpoint_group_tag WHERE id = ?", tagId)
	return tag, err
}

// GetCheckpointGroupTag retrieves a tag by ID.
func GetCheckpointGroupTag(tx *sqlx.Tx, tagId string) (models.CheckpointGroupTag, error) {
	tag := models.CheckpointGroupTag{}
	err := tx.Get(&tag, "SELECT * FROM checkpoint_group_tag WHERE id = ?", tagId)
	if errors.Is(err, sql.ErrNoRows) {
		return tag, errors.New("checkpoint group tag not found")
	}
	return tag, err
}

// GetCheckpointGroupTagsForAsset returns tags whose groups contain the asset.
func GetCheckpointGroupTagsForAsset(tx *sqlx.Tx, assetId string) ([]models.CheckpointGroupTag, error) {
	tags := []models.CheckpointGroupTag{}
	err := tx.Select(&tags, `
		SELECT DISTINCT cgt.*
		FROM checkpoint_group_tag cgt
		JOIN checkpoint_group cg ON cg.id = cgt.group_id
		JOIN asset_checkpoint ac ON ac.group_id = cg.id
		WHERE cg.finalized = 1
			AND cg.group_type = ?
			AND ac.asset_id = ?
			AND ac.trashed = 0
		ORDER BY cgt.name COLLATE NOCASE
	`, CheckpointGroupTypeMulti, assetId)
	return tags, err
}

// GetCheckpointGroupAssetIds returns active assets represented by a group.
func GetCheckpointGroupAssetIds(tx *sqlx.Tx, groupId string) ([]string, error) {
	assetIds := []string{}
	err := tx.Select(&assetIds, `
		SELECT DISTINCT asset_id
		FROM asset_checkpoint
		WHERE group_id = ? AND trashed = 0
		ORDER BY asset_id
	`, groupId)
	return assetIds, err
}

// DeleteCheckpointGroupTag removes a tag and records its tombstone.
func DeleteCheckpointGroupTag(tx *sqlx.Tx, tagId string) error {
	var referenceCount int
	if err := tx.Get(&referenceCount, "SELECT COUNT(*) FROM asset_dependency WHERE checkpoint_group_tag_id = ?", tagId); err != nil {
		return err
	}
	if referenceCount > 0 {
		return errors.New("checkpoint group tag is referenced by a dependency")
	}
	result, err := tx.Exec("DELETE FROM checkpoint_group_tag WHERE id = ?", tagId)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("checkpoint group tag not found")
	}
	return nil
}
