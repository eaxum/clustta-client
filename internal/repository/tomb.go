package repository

import (
	"clustta/internal/base_service"
	"database/sql"
	"fmt"

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

func AddItemsToTomb(tx *sqlx.Tx, tombs []Tomb) error {
	for _, tomb := range tombs {
		query := fmt.Sprintf("DELETE FROM %s WHERE id = '%s';", tomb.TableName, tomb.Id)
		_, err := tx.Exec(query)
		if err != nil {
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
