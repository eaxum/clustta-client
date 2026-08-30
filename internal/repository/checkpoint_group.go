package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"clustta/internal/repository/models"
	"clustta/internal/utils"
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
		return true, nil
	}
	return false, nil
}

// FinalizeCheckpointGroup validates members and completes the group.
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
