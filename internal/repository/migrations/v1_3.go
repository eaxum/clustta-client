package migrations

import (
	"clustta/internal/auth_service"
	"clustta/internal/settings"
	"clustta/internal/utils"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
)

// MigrateV1_3 sets the default working directory for the project.
func MigrateV1_3(db *sqlx.DB, _ string) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	projectWorkingDir := ""
	if !settings.IsServer() {
		user, err := auth_service.GetActiveUser()
		if err != nil {
			return err
		}

		workingDir, err := settings.GetUserDataFolder(user)
		if err != nil {
			return err
		}
		studioName, err := utils.GetStudioName(tx)
		if err != nil {
			return err
		}
		projectName, err := utils.GetProjectName(tx)
		if err != nil {
			return err
		}
		projectWorkingDir = filepath.Join(workingDir, studioName, projectName)
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		workingDir := filepath.Join(homeDir, "Documents", "clustta")
		os.MkdirAll(workingDir, os.ModePerm)

		projectName, err := utils.GetProjectName(tx)
		if err != nil {
			return err
		}
		projectWorkingDir = filepath.Join(workingDir, projectName)
	}

	err = utils.SetProjectWorkingDir(tx, projectWorkingDir)
	if err != nil {
		return err
	}

	return tx.Commit()
}
