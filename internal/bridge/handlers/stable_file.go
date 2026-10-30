package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"clustta/internal/utils"
	"github.com/mattn/go-sqlite3"
)

var errSourceFileChanged = errors.New("file changed while preparing the operation - wait for the host save to finish and retry")

const (
	audacityApplicationID = 0x41554459
	backupPageCount       = 1024
	backupStallTimeout    = 30 * time.Second
	backupRetryDelay      = 10 * time.Millisecond
)

var audacityProjectTables = []string{
	"autosave",
	"project",
	"project_history",
	"sampleblocks",
}

func createStableFileSnapshot(sourcePath string) (string, error) {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return "", err
	}
	if !sourceInfo.Mode().IsRegular() || sourceInfo.Size() == 0 {
		return "", errors.New("source file must be a non-empty regular file")
	}
	if strings.EqualFold(filepath.Ext(sourcePath), ".aup4") {
		return createAudacityProjectSnapshot(sourcePath, sourceInfo)
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		return "", err
	}
	defer source.Close()

	extension := filepath.Ext(sourcePath)
	snapshotFile, err := os.CreateTemp("", "clustta-bridge-*"+extension)
	if err != nil {
		return "", err
	}
	snapshotPath := snapshotFile.Name()
	removeSnapshot := true
	defer func() {
		if removeSnapshot {
			os.Remove(snapshotPath)
		}
	}()

	if _, err = io.Copy(snapshotFile, source); err != nil {
		snapshotFile.Close()
		return "", err
	}
	if err = snapshotFile.Close(); err != nil {
		return "", err
	}
	if err = os.Chtimes(snapshotPath, sourceInfo.ModTime(), sourceInfo.ModTime()); err != nil {
		return "", err
	}
	if err = verifyStableFileSnapshot(sourcePath, sourceInfo, snapshotPath); err != nil {
		return "", err
	}

	removeSnapshot = false
	return snapshotPath, nil
}

func createAudacityProjectSnapshot(sourcePath string, sourceInfo os.FileInfo) (string, error) {
	snapshotFile, err := os.CreateTemp("", "clustta-bridge-*.aup4")
	if err != nil {
		return "", err
	}
	snapshotPath := snapshotFile.Name()
	if err = snapshotFile.Close(); err != nil {
		os.Remove(snapshotPath)
		return "", err
	}
	removeSnapshot := true
	defer func() {
		if removeSnapshot {
			os.Remove(snapshotPath)
		}
	}()

	if err = backupSQLiteDatabase(sourcePath, snapshotPath); err != nil {
		return "", fmt.Errorf("prepare Audacity project snapshot: %w", err)
	}
	if err = validateAudacityProjectSnapshot(snapshotPath); err != nil {
		return "", err
	}
	if err = os.Chtimes(snapshotPath, sourceInfo.ModTime(), sourceInfo.ModTime()); err != nil {
		return "", err
	}

	removeSnapshot = false
	return snapshotPath, nil
}

func backupSQLiteDatabase(sourcePath, destinationPath string) error {
	sourceDB, err := sql.Open("sqlite3", sqliteFileURI(sourcePath, "mode=ro&_busy_timeout=5000"))
	if err != nil {
		return err
	}
	defer sourceDB.Close()
	destinationDB, err := sql.Open("sqlite3", sqliteFileURI(destinationPath, "mode=rwc&_busy_timeout=5000"))
	if err != nil {
		return err
	}
	defer destinationDB.Close()

	sourceConnection, err := sourceDB.Conn(context.Background())
	if err != nil {
		return err
	}
	defer sourceConnection.Close()
	destinationConnection, err := destinationDB.Conn(context.Background())
	if err != nil {
		return err
	}
	defer destinationConnection.Close()

	return destinationConnection.Raw(func(destinationDriver any) error {
		destination, ok := destinationDriver.(*sqlite3.SQLiteConn)
		if !ok {
			return errors.New("Audacity snapshot requires SQLite")
		}
		return sourceConnection.Raw(func(sourceDriver any) error {
			source, ok := sourceDriver.(*sqlite3.SQLiteConn)
			if !ok {
				return errors.New("Audacity project requires SQLite")
			}
			backup, backupErr := destination.Backup("main", source, "main")
			if backupErr != nil {
				return backupErr
			}
			lastRemaining := -1
			lastProgress := time.Now()
			for {
				done, stepErr := backup.Step(backupPageCount)
				if stepErr != nil {
					backup.Close()
					return stepErr
				}
				if done {
					return backup.Close()
				}
				remaining := backup.Remaining()
				if remaining != lastRemaining {
					lastRemaining = remaining
					lastProgress = time.Now()
					continue
				}
				if time.Since(lastProgress) >= backupStallTimeout {
					backup.Close()
					return errors.New("Audacity project remained busy during snapshot")
				}
				time.Sleep(backupRetryDelay)
			}
		})
	})
}

func validateAudacityProjectSnapshot(snapshotPath string) error {
	db, err := sql.Open("sqlite3", sqliteFileURI(snapshotPath, "mode=ro"))
	if err != nil {
		return err
	}
	defer db.Close()

	var applicationID int
	if err = db.QueryRow("PRAGMA application_id").Scan(&applicationID); err != nil {
		return err
	}
	if applicationID != audacityApplicationID {
		return errors.New("selected file is not an Audacity project")
	}
	var integrityResult string
	if err = db.QueryRow("PRAGMA quick_check").Scan(&integrityResult); err != nil {
		return err
	}
	if integrityResult != "ok" {
		return fmt.Errorf("Audacity project snapshot failed integrity check: %s", integrityResult)
	}
	for _, table := range audacityProjectTables {
		var count int
		err = db.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?",
			table,
		).Scan(&count)
		if err != nil {
			return err
		}
		if count != 1 {
			return fmt.Errorf("Audacity project is missing the %s table", table)
		}
	}
	var autosaveCount int
	if err = db.QueryRow("SELECT COUNT(*) FROM autosave").Scan(&autosaveCount); err != nil {
		return err
	}
	if autosaveCount != 0 {
		return errors.New("save the current Audacity project before creating a checkpoint")
	}
	return nil
}

func sqliteFileURI(filePath, rawQuery string) string {
	uriPath := filepath.ToSlash(filePath)
	if !strings.HasPrefix(uriPath, "/") {
		uriPath = "/" + uriPath
	}
	return (&url.URL{Scheme: "file", Path: uriPath, RawQuery: rawQuery}).String()
}

func verifyStableFileSnapshot(sourcePath string, sourceInfo os.FileInfo, snapshotPath string) error {
	currentInfo, err := os.Stat(sourcePath)
	if err != nil {
		return errSourceFileChanged
	}
	if !os.SameFile(sourceInfo, currentInfo) ||
		sourceInfo.Size() != currentInfo.Size() ||
		!sourceInfo.ModTime().Equal(currentInfo.ModTime()) {
		return errSourceFileChanged
	}

	sourceChecksum, err := utils.GenerateXXHashChecksum(sourcePath)
	if err != nil {
		return err
	}
	snapshotChecksum, err := utils.GenerateXXHashChecksum(snapshotPath)
	if err != nil {
		return err
	}
	if sourceChecksum != snapshotChecksum {
		return errSourceFileChanged
	}
	return nil
}
