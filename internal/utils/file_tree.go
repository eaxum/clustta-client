package utils

import (
	"path/filepath"
)

func BuildEntityPath(rootFolder, entityPath string) (string, error) {
	taskFilePath := filepath.Join(rootFolder, entityPath)
	return taskFilePath, nil
}
func BuildTaskPath(rootFolder, entityPath, taskName, extension string) (string, error) {
	taskFileName := taskName + extension
	taskFilePath := filepath.Join(rootFolder, entityPath, taskFileName)
	return taskFilePath, nil
}
