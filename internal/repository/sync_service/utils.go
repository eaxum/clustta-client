package sync_service

import (
	"clustta/internal/repository"
	"clustta/internal/utils"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type SyncOptions struct {
	OnlyLatestCheckpoints bool `json:"only_latest_checkpoints"`
	AssetDependencies      bool `json:"asset_dependencies"`
	Assets                 bool `json:"assets"`
	Resources             bool `json:"resources"`
	Templates             bool `json:"templates"`
	Force                 bool `json:"force"`
}

var ProjectTables = []string{
	"role", "user", "status", "tag",
	"asset_type", "asset", "dependency_type", "asset_dependency", "collection_dependency",
	"collection_type", "collection", "collection_assignee", "template",
	"workflow", "workflow_link", "workflow_collection", "workflow_asset",
	"asset_tag", "asset_checkpoint", "tomb",
	"integration_project", "integration_collection_mapping", "integration_asset_mapping",
}

func clearTables(tx *sqlx.Tx, tables []string) error {
	for _, table := range tables {
		query := "DELETE FROM " + table
		_, err := tx.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func dropTables(tx *sqlx.Tx, tables []string) error {
	for _, table := range tables {
		query := "DROP TABLE IF EXISTS " + table
		_, err := tx.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}
	return nil
}

func ClearLocalData(tx *sqlx.Tx) error {
	// Clear the tables
	err := clearTables(tx, ProjectTables)
	if err != nil {
		return err
	}
	return nil
}

func ClearLocalDataDrop(tx *sqlx.Tx) error {
	// Clear the tables
	err := dropTables(tx, ProjectTables)
	if err != nil {
		return err
	}

	err = utils.CreateSchemaTx(tx, repository.ProjectSchema)
	if err != nil {
		return err
	}

	return nil
}
