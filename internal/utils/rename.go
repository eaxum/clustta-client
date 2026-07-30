package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// RenamePathCaseSafe renames a path, using a temporary sibling for case-only
// changes on case-insensitive filesystems.
func RenamePathCaseSafe(oldPath, newPath string) error {
	if oldPath == newPath {
		return nil
	}
	oldInfo, oldErr := os.Stat(oldPath)
	if oldErr != nil {
		return oldErr
	}
	if targetInfo, err := os.Stat(newPath); err == nil && !os.SameFile(oldInfo, targetInfo) {
		return fmt.Errorf("rename target already exists: %s", newPath)
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	if !strings.EqualFold(filepath.Clean(oldPath), filepath.Clean(newPath)) {
		return os.Rename(oldPath, newPath)
	}
	tempPath := filepath.Join(filepath.Dir(oldPath), ".clustta-rename-"+uuid.NewString())
	if err := os.Rename(oldPath, tempPath); err != nil {
		return err
	}
	if err := os.Rename(tempPath, newPath); err != nil {
		if rollbackErr := os.Rename(tempPath, oldPath); rollbackErr != nil {
			return fmt.Errorf("case-only rename failed: %v; rollback failed: %v", err, rollbackErr)
		}
		return err
	}
	return nil
}
