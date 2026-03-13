package migrations

import (
	"database/sql"
	_ "embed"
	"time"

	"clustta/internal/utils"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

//go:embed sql/v1_8_drop.sql
var v1_8DropSQL string

// MigrateV1_8 renames all task/entity tables and columns to asset/collection.
func MigrateV1_8(db *sqlx.DB, schema string) error {
	// Drop old views, triggers, and indexes FIRST so that SQLite's
	// ALTER TABLE RENAME COLUMN doesn't try to update stale references.
	stmts := utils.SplitStatements(v1_8DropSQL)
	for _, stmt := range stmts {
		_, err := db.Exec(stmt)
		if err != nil {
			return err
		}
	}

	// Rename tables from old naming to new naming.
	// CreateSchema may have created empty tables with new names,
	// so drop those first, then rename the old tables.
	tableRenames := [][2]string{
		{"workflow_entity", "workflow_collection"},
		{"workflow_task", "workflow_asset"},
		{"entity_assignee", "collection_assignee"},
		{"entity_dependency", "collection_dependency"},
		{"task_dependency", "asset_dependency"},
		{"task_tag", "asset_tag"},
		{"task_checkpoint", "asset_checkpoint"},
		{"task", "asset"},
		{"entity", "collection"},
		{"task_type", "asset_type"},
		{"entity_type", "collection_type"},
	}

	for _, rename := range tableRenames {
		oldName, newName := rename[0], rename[1]
		oldExists, err := utils.TableExists(db, oldName)
		if err != nil {
			return err
		}
		if oldExists {
			_, err = db.Exec("DROP TABLE IF EXISTS " + newName)
			if err != nil {
				return err
			}
			err = utils.RenameTable(db, oldName, newName)
			if err != nil {
				return err
			}
		}
	}

	// Rename columns within renamed tables.
	columnRenames := [][3]string{
		{"collection", "entity_path", "collection_path"},
		{"collection", "entity_type_id", "collection_type_id"},
		{"collection_assignee", "entity_id", "collection_id"},
		{"collection_dependency", "task_id", "asset_id"},
		{"asset", "task_type_id", "asset_type_id"},
		{"asset", "entity_id", "collection_id"},
		{"asset_dependency", "task_id", "asset_id"},
		{"asset_tag", "task_id", "asset_id"},
		{"asset_checkpoint", "task_id", "asset_id"},
		{"workflow_collection", "entity_type_id", "collection_type_id"},
		{"workflow_asset", "task_type_id", "asset_type_id"},
		{"workflow_link", "entity_type_id", "collection_type_id"},
	}

	for _, rename := range columnRenames {
		err := utils.RenameColumn(db, rename[0], rename[1], rename[2])
		if err != nil {
			return err
		}
	}

	// Rename role permission columns.
	roleRenames := [][2]string{
		{"view_entity", "view_collection"},
		{"create_entity", "create_collection"},
		{"update_entity", "update_collection"},
		{"delete_entity", "delete_collection"},
		{"view_task", "view_asset"},
		{"create_task", "create_asset"},
		{"update_task", "update_asset"},
		{"delete_task", "delete_asset"},
		{"assign_task", "assign_asset"},
		{"unassign_task", "unassign_asset"},
		{"set_done_task", "set_done_asset"},
		{"set_retake_task", "set_retake_asset"},
		{"view_done_task", "view_done_asset"},
	}

	for _, rename := range roleRenames {
		err := utils.RenameColumn(db, "role", rename[0], rename[1])
		if err != nil {
			return err
		}
	}

	// Update tomb table entries to use new table names.
	tombRenames := [][2]string{
		{"entity", "collection"},
		{"entity_type", "collection_type"},
		{"entity_assignee", "collection_assignee"},
		{"entity_dependency", "collection_dependency"},
		{"task", "asset"},
		{"task_type", "asset_type"},
		{"task_dependency", "asset_dependency"},
		{"task_tag", "asset_tag"},
		{"task_checkpoint", "asset_checkpoint"},
		{"workflow_entity", "workflow_collection"},
		{"workflow_task", "workflow_asset"},
	}

	for _, rename := range tombRenames {
		_, err := db.Exec("UPDATE tomb SET table_name = ? WHERE table_name = ?", rename[1], rename[0])
		if err != nil {
			return err
		}
	}

	// Re-apply schema to create views, triggers, and indexes with new names.
	err := utils.CreateSchema(db, schema)
	if err != nil {
		return err
	}

	// Run icon migration on the now-renamed tables.
	if err := remapIcons(db); err != nil {
		return err
	}

	// Auto-group any ungrouped checkpoints on the now-renamed table.
	return autoGroupCheckpointsNew(db)
}

// remapIcons updates legacy icon names in asset_type and collection_type.
func remapIcons(db *sqlx.DB) error {
	iconMap := map[string]string{
		"hdri": "image", "character creation": "masks", "prop creation": "drum",
		"environment creation": "stall", "concept art": "palette", "modeling": "cube",
		"rigging": "bone", "texturing": "texture", "lookdev": "mystery-ball",
		"editing": "scissors", "previz": "video-camera", "layout": "shapes",
		"animation": "man-running", "fx": "fire", "lighting": "bulb",
		"rendering": "camera-flash", "compositing": "flow-chart",
		"character": "masks", "prop": "drum", "environment": "stall",
		"scene": "tree", "shot": "clapboard", "sequence": "film-strip", "episode": "film-reel",
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	type typeRow struct {
		Id   string `db:"id"`
		Icon string `db:"icon"`
	}

	var assetTypes []typeRow
	if err := tx.Select(&assetTypes, "SELECT id, icon FROM asset_type"); err == nil {
		for _, at := range assetTypes {
			if newIcon, exists := iconMap[at.Icon]; exists {
				_, _ = tx.Exec("UPDATE asset_type SET icon = ? WHERE id = ?", newIcon, at.Id)
			}
		}
	}

	var collectionTypes []typeRow
	if err := tx.Select(&collectionTypes, "SELECT id, icon FROM collection_type"); err == nil {
		for _, ct := range collectionTypes {
			if newIcon, exists := iconMap[ct.Icon]; exists {
				_, _ = tx.Exec("UPDATE collection_type SET icon = ? WHERE id = ?", newIcon, ct.Id)
			}
		}
	}

	return tx.Commit()
}

// autoGroupCheckpointsNew groups ungrouped checkpoints using the new table name (asset_checkpoint).
func autoGroupCheckpointsNew(db *sqlx.DB) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var ungroupedCount int
	err = tx.Get(&ungroupedCount, "SELECT COUNT(*) FROM asset_checkpoint WHERE group_id = '' OR group_id IS NULL")
	if err != nil || ungroupedCount == 0 {
		return nil
	}

	type miniCheckpoint struct {
		Id        string `db:"id"`
		CreatedAt string `db:"created_at"`
		Comment   string `db:"comment"`
		AuthorUID string `db:"author_id"`
	}

	var checkpoints []miniCheckpoint
	err = tx.Select(&checkpoints, `SELECT id, created_at, comment, author_id
		FROM asset_checkpoint
		WHERE group_id = '' OR group_id IS NULL
		ORDER BY created_at DESC`)
	if err != nil && err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}

	type group struct {
		CreatedAt     string
		CheckpointIds []string
		Comment       string
		AuthorUID     string
		GroupId       string
	}

	var timeline []group
	prev := group{}
	for i, cp := range checkpoints {
		if prev.CreatedAt == "" {
			prev = group{
				CreatedAt:     cp.CreatedAt,
				CheckpointIds: []string{cp.Id},
				Comment:       cp.Comment,
				AuthorUID:     cp.AuthorUID,
				GroupId:       uuid.New().String(),
			}
			if i == len(checkpoints)-1 {
				timeline = append(timeline, prev)
			}
			continue
		}

		cpTime, err := time.Parse(time.RFC3339, cp.CreatedAt)
		if err != nil {
			return err
		}
		prevTime, err := time.Parse(time.RFC3339, prev.CreatedAt)
		if err != nil {
			return err
		}

		diff := prevTime.Sub(cpTime)
		if prev.AuthorUID == cp.AuthorUID && prev.Comment == cp.Comment && diff.Seconds() <= 120 {
			prev.CheckpointIds = append(prev.CheckpointIds, cp.Id)
		} else {
			timeline = append(timeline, prev)
			prev = group{
				CreatedAt:     cp.CreatedAt,
				CheckpointIds: []string{cp.Id},
				Comment:       cp.Comment,
				AuthorUID:     cp.AuthorUID,
				GroupId:       uuid.New().String(),
			}
		}
		if i == len(checkpoints)-1 {
			timeline = append(timeline, prev)
		}
	}

	for _, g := range timeline {
		for _, id := range g.CheckpointIds {
			_, err := tx.Exec("UPDATE asset_checkpoint SET group_id = ? WHERE id = ?", g.GroupId, id)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}
