package utils

import (
	"path/filepath"
)

func BuildCollectionPath(rootFolder, collectionPath string) (string, error) {
	assetFilePath := filepath.Join(rootFolder, collectionPath)
	return assetFilePath, nil
}
func BuildAssetPath(rootFolder, collectionPath, assetName, extension string) (string, error) {
	assetFileName := assetName + extension
	assetFilePath := filepath.Join(rootFolder, collectionPath, assetFileName)
	return assetFilePath, nil
}
