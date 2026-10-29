package handlers

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"clustta/internal/utils"
)

var errSourceFileChanged = errors.New("file changed while preparing the operation - wait for the host save to finish and retry")

func createStableFileSnapshot(sourcePath string) (string, error) {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return "", err
	}
	if !sourceInfo.Mode().IsRegular() || sourceInfo.Size() == 0 {
		return "", errors.New("source file must be a non-empty regular file")
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
