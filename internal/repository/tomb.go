package repository

import (
	"clustta/internal/base_service"
	"database/sql"
	"fmt"
	"sort"

	"github.com/jmoiron/sqlx"
)

type Tomb struct {
	Id        string `db:"id" json:"id"`
	Mtime     int    `db:"mtime" json:"mtime"`
	TableName string `db:"table_name" json:"table_name"`
	Synced    bool   `db:"synced" json:"synced"`
}

func GetTombs(tx *sqlx.Tx) ([]Tomb, error) {
	tombs := []Tomb{}
	err := base_service.GetAll(tx, "tomb", &tombs)
	if err != nil {
		return tombs, err
	}
	return tombs, nil
}

// GetTomb returns a single tomb entry by its ID.
func GetTomb(tx *sqlx.Tx, id string) (Tomb, error) {
	var tomb Tomb
	err := tx.Get(&tomb, "SELECT id, mtime, table_name, synced FROM tomb WHERE id = ?", id)
	return tomb, err
}

func AddItemsToTomb(tx *sqlx.Tx, tombs []Tomb) error {
	orderedTombs := append([]Tomb(nil), tombs...)
	sort.SliceStable(orderedTombs, func(i, j int) bool {
		return tombDeletePriority(orderedTombs[i].TableName) < tombDeletePriority(orderedTombs[j].TableName)
	})
	for _, tomb := range orderedTombs {
		switch tomb.TableName {
		case "asset", "asset_checkpoint", "asset_checkpoint_tag", "asset_dependency", "collection_dependency",
			"asset_tag", "tag", "asset_type", "dependency_type", "collection", "collection_type",
			"collection_assignee", "template", "status", "user", "role", "workflow", "workflow_link",
			"workflow_asset", "workflow_collection", "integration_project",
			"integration_asset_mapping", "integration_collection_mapping":
		default:
			return fmt.Errorf("invalid tomb table: %s", tomb.TableName)
		}
		if _, err := tx.Exec("DELETE FROM "+tomb.TableName+" WHERE id = ?", tomb.Id); err != nil {
			return err
		}
		if _, err := tx.Exec(`
			INSERT INTO tomb (id, mtime, table_name, synced) VALUES (?, ?, ?, 0)
			ON CONFLICT(id) DO UPDATE SET mtime = MAX(tomb.mtime, excluded.mtime), synced = 0
		`, tomb.Id, tomb.Mtime, tomb.TableName); err != nil {
			return err
		}
	}
	return nil
}

func GetTombedItems(tx *sqlx.Tx) ([]string, error) {
	var tombedIds []string
	err := tx.Select(&tombedIds, "SELECT id FROM tomb")
	return tombedIds, err
}

func IsItemInTomb(tx *sqlx.Tx, itemID, tableName string) (bool, error) {
	var isItemInTomb bool
	query := `
		SELECT COUNT(*) > 0 AS item_in_tomb
		FROM tomb
		WHERE id = ?
		  AND table_name = ?
		  AND synced = 0
	`
	err := tx.Get(&isItemInTomb, query, itemID, tableName)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return isItemInTomb, nil
}

// RemoveItemFromTomb removes a matching tombstone.
func RemoveItemFromTomb(tx *sqlx.Tx, itemID, tableName string) error {
	_, err := tx.Exec("DELETE FROM tomb WHERE id = ? AND table_name = ?", itemID, tableName)
	return err
}

func tombDeletePriority(tableName string) int {
	switch tableName {
	case "asset_dependency", "collection_dependency":
		return 0
	case "asset_checkpoint_tag":
		return 1
	case "asset":
		return 3
	default:
		return 2
	}
}
