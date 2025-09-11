package server_service

import (
	"clustta/internal/utils"
	"database/sql"
	"strconv"

	"github.com/jmoiron/sqlx"
)

func GetServerName(tx *sqlx.Tx) (string, error) {
	var serverName string
	err := tx.Get(&serverName, "SELECT value FROM config WHERE name='server_name'")
	if err != nil {
		return serverName, err
	}
	return serverName, nil
}

func SetServerName(tx *sqlx.Tx, serverName string) error {
	_, err := tx.Exec(`
		INSERT INTO config (name, value, mtime)
		VALUES ('server_name', $1, $2)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime
	`, serverName, utils.GetEpochTime())
	return err
}

func GetServerVersion(tx *sqlx.Tx) (float64, error) {
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

func SetServerVersion(tx *sqlx.Tx, version float64) error {
	versionStr := strconv.FormatFloat(version, 'f', -1, 64)
	_, err := tx.Exec(`
		INSERT INTO config (name, value, mtime)
		VALUES ('version', ?, ?)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime
	`, versionStr, utils.GetEpochTime())
	return err
}
