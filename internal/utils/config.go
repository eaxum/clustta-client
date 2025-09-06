package utils

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/jmoiron/sqlx"
)

func SetTableToSynced(tx *sqlx.Tx, table string) error {
	query := fmt.Sprintf("UPDATE %s SET synced = 1 WHERE synced = 0;", table)
	_, err := tx.Exec(query)
	if err != nil {
		return err
	}
	return nil
}
func SetTablesToSynced(tx *sqlx.Tx, tables []string) error {
	for _, table := range tables {
		err := SetTableToSynced(tx, table)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetProjectVersion(tx *sqlx.Tx) (float64, error) {
	var version string
	err := tx.Get(&version, "SELECT value FROM config WHERE name = 'version'")
	if err != nil && err == sql.ErrNoRows {
		return 0.0, nil
	} else if err != nil {
		return 0.0, err
	}
	versionFloat, err := strconv.ParseFloat(version, 8)
	if err != nil {
		return 0.0, err
	}
	return versionFloat, nil
}

func SetProjectVersion(tx *sqlx.Tx, version float64) error {
	versionStr := strconv.FormatFloat(version, 'f', -1, 64)
	_, err := tx.Exec(`
		INSERT INTO config (name, value, mtime)
		VALUES ('version', ?, ?)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime
	`, versionStr, GetEpochTime())
	return err
}

func GetProjectWorkingDir(tx *sqlx.Tx) (string, error) {
	var projectIcon string
	err := tx.Get(&projectIcon, "SELECT value FROM config WHERE name='working_dir'")
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("invalid working directory")
		}
		return projectIcon, err
	}
	return projectIcon, nil
}

func SetProjectWorkingDir(tx *sqlx.Tx, workingDir string) error {
	_, err := tx.Exec(`
		INSERT INTO config (name, value, mtime)
		VALUES ('working_dir', $1, $2)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime
	`, workingDir, GetEpochTime())
	return err
}

func GetProjectId(tx *sqlx.Tx) (string, error) {
	var project_id string
	err := tx.Get(&project_id, "SELECT value FROM config WHERE name='project_id'")
	if err != nil {
		return "", err
	}
	return project_id, nil
}

func GetLastSyncTime(tx *sqlx.Tx) (int64, error) {
	var lastSyncTime int64
	err := tx.Get(&lastSyncTime, "SELECT value FROM config WHERE name='last_sync_time'")
	if err != nil {
		return lastSyncTime, err
	}
	return lastSyncTime, nil
}

func SetLastSyncTime(tx *sqlx.Tx, lastSyncTime int64) error {
	_, err := tx.Exec(`
		INSERT INTO config (name, value, mtime)
		VALUES ('last_sync_time', $1, $2)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime
	`, lastSyncTime, GetEpochTime())
	return err
}

func GetStudioName(tx *sqlx.Tx) (string, error) {
	var studioName string
	query := `
	SELECT COALESCE(
		(SELECT value FROM config WHERE name = 'studio_name'),
		(SELECT value FROM config WHERE name = 'server_name')
	) AS value;
	`
	// err := tx.Get(&studioName, "SELECT value FROM config WHERE name='studio_name'")
	err := tx.Get(&studioName, query)
	if err != nil {
		return studioName, err
	}
	return studioName, nil
}

func SetStudioName(tx *sqlx.Tx, studioName string) error {
	_, err := tx.Exec(`
		INSERT INTO config (name, value, mtime)
		VALUES ('studio_name', $1, $2)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime
	`, studioName, GetEpochTime())
	return err
}
func IsProjectDirty(tx *sqlx.Tx) (bool, error) {
	var isDirty bool
	err := tx.Get(&isDirty, "SELECT value FROM config WHERE name='is_dirty'")
	if err != nil {
		return isDirty, err
	}
	return isDirty, nil
}

func SetProjectDirty(tx *sqlx.Tx, isDirty bool) error {
	_, err := tx.Exec(`
		INSERT INTO config (name, value, mtime)
		VALUES ('is_dirty', $1, $2)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime
	`, isDirty, GetEpochTime())
	return err
}

func SetIsClosed(tx *sqlx.Tx, isClosed bool) error {
	_, err := tx.Exec(`
		INSERT INTO config (name, value, mtime)
		VALUES ('is_closed', $1, $2)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime
	`, isClosed, GetEpochTime())
	return err
}

func GetIsClosed(tx *sqlx.Tx) (bool, error) {
	var isClosed bool
	err := tx.Get(&isClosed, "SELECT value FROM config WHERE name='is_closed'")
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return isClosed, err
	}
	return isClosed, nil
}

func GetProjectIcon(tx *sqlx.Tx) (string, error) {
	var projectIcon string
	err := tx.Get(&projectIcon, "SELECT value FROM config WHERE name='project_icon'")
	if err != nil {
		if err == sql.ErrNoRows {
			return "&#128188", nil
		}
		return projectIcon, err
	}
	return projectIcon, nil
}

func SetProjectSyncToken(tx *sqlx.Tx, syncToken string) error {
	_, err := tx.Exec(`
		INSERT INTO config (name, value, mtime)
		VALUES ('sync_token', $1, $2)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime
	`, syncToken, GetEpochTime())
	return err
}

func GetProjectSyncToken(tx *sqlx.Tx) (string, error) {
	var projectIcon string
	err := tx.Get(&projectIcon, "SELECT value FROM config WHERE name='sync_token'")
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return projectIcon, err
	}
	return projectIcon, nil
}

func SetProjectIcon(tx *sqlx.Tx, projectIcon string) error {
	_, err := tx.Exec(`
		INSERT INTO config (name, value, mtime)
		VALUES ('project_icon', $1, $2)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime
	`, projectIcon, GetEpochTime())
	return err
}

func SetProjectIgnoreList(tx *sqlx.Tx, ignoreList []string) error {
	ignoreListJson, err := json.Marshal(ignoreList)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO config (name, value, mtime, synced)
		VALUES ('ignore_list', $1, $2, 0)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime, synced = 0
	`, ignoreListJson, GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}
