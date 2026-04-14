package migrations

import (
	"database/sql"
	"errors"
	"time"

	"clustta/internal/utils"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// MigrateV1_4 adds group_id column to task_checkpoint and auto-groups checkpoints.
func MigrateV1_4(db *sqlx.DB, _ string) error {
	err := utils.AddColumnIfNotExist(db, "task_checkpoint", "group_id", "TEXT", "", false)
	if err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = autoGroupCheckpointsLegacy(tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// autoGroupCheckpointsLegacy groups checkpoints using the old table name (task_checkpoint).
func autoGroupCheckpointsLegacy(tx *sqlx.Tx) error {
	type groupedTimeline struct {
		CreatedAt     string
		CheckpointIds []string
		Comment       string
		AuthorUID     string
		GroupId       string
	}
	type miniCheckpoint struct {
		Id        string `db:"id"`
		CreatedAt string `db:"created_at"`
		Comment   string `db:"comment"`
		AuthorUID string `db:"author_id"`
	}

	var checkpoints []miniCheckpoint
	query := `SELECT id, created_at, comment, author_id
		FROM task_checkpoint
		ORDER BY created_at DESC`
	err := tx.Select(&checkpoints, query)
	if err != nil && err == sql.ErrNoRows {
		return errors.New("no checkpoints")
	} else if err != nil {
		return err
	}

	var timeline []groupedTimeline
	previous := groupedTimeline{}
	for i, cp := range checkpoints {
		if previous.CreatedAt == "" {
			previous = groupedTimeline{
				CreatedAt:     cp.CreatedAt,
				CheckpointIds: []string{cp.Id},
				Comment:       cp.Comment,
				AuthorUID:     cp.AuthorUID,
				GroupId:       uuid.New().String(),
			}
			if i == len(checkpoints)-1 {
				timeline = append(timeline, previous)
			}
			continue
		}

		cpTime, err := time.Parse(time.RFC3339, cp.CreatedAt)
		if err != nil {
			return err
		}
		prevTime, err := time.Parse(time.RFC3339, previous.CreatedAt)
		if err != nil {
			return err
		}

		diff := prevTime.Sub(cpTime)
		if previous.AuthorUID == cp.AuthorUID && previous.Comment == cp.Comment && diff.Seconds() <= 120 {
			previous.CheckpointIds = append(previous.CheckpointIds, cp.Id)
		} else {
			timeline = append(timeline, previous)
			previous = groupedTimeline{
				CreatedAt:     cp.CreatedAt,
				CheckpointIds: []string{cp.Id},
				Comment:       cp.Comment,
				AuthorUID:     cp.AuthorUID,
				GroupId:       uuid.New().String(),
			}
		}
		if i == len(checkpoints)-1 {
			timeline = append(timeline, previous)
		}
	}

	for _, group := range timeline {
		for _, id := range group.CheckpointIds {
			_, err := tx.Exec("UPDATE task_checkpoint SET group_id = ? WHERE id = ?", group.GroupId, id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
