package repository

import (
	"os"
	"path/filepath"
	"testing"

	"clustta/internal/utils"

	"github.com/stretchr/testify/require"
)

func TestJackOfAllTradesTemplateRoundTrip(t *testing.T) {
	definitions, err := loadTemplateDefinitions()
	require.NoError(t, err)
	var definition ProjectTemplateDefinition
	for _, candidate := range definitions.Templates {
		if candidate.Name == "Jack of all Trades" {
			definition = candidate
			break
		}
	}
	require.NotEmpty(t, definition.Name)
	require.Len(t, definition.AssetTemplates, 4)

	templatePath := filepath.Join(t.TempDir(), definition.Name+".clst")
	createAssetTemplateTestProject(t, templatePath)
	templateDB, err := utils.OpenDb(templatePath)
	require.NoError(t, err)
	templateTx, err := templateDB.Beginx()
	require.NoError(t, err)
	require.NoError(t, createBundledAssetTemplates(templateTx, t.TempDir(), definition.AssetTemplates))
	require.NoError(t, createTemplateTags(templateTx, definition.Tags))
	for _, assetType := range definition.AssetTypes {
		_, err := GetOrCreateAssetType(templateTx, assetType.Name, assetType.Icon)
		require.NoError(t, err)
	}
	for _, collectionType := range definition.CollectionTypes {
		_, err := GetOrCreateCollectionType(templateTx, collectionType.Name, collectionType.Icon)
		require.NoError(t, err)
	}
	require.NoError(t, templateTx.Commit())
	require.NoError(t, templateDB.Close())
	projectPath := filepath.Join(t.TempDir(), "project.clst")
	createAssetTemplateTestProject(t, projectPath)
	require.NoError(t, LoadProjectTemplateData(projectPath, templatePath))
	metadata, err := extractTemplateData(templatePath)
	require.NoError(t, err)
	_, err = copyTemplateMetadata(projectPath, metadata)
	require.NoError(t, err)

	db, err := utils.OpenDb(projectPath)
	require.NoError(t, err)
	defer db.Close()
	tx, err := db.Beginx()
	require.NoError(t, err)
	defer tx.Rollback()
	starters, err := GetTemplates(tx, false)
	require.NoError(t, err)
	require.Len(t, starters, 4)
	tags, err := GetTags(tx)
	require.NoError(t, err)
	tagNames := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}
	require.ElementsMatch(t, definition.Tags, tagNames)
	assetTypes, err := GetAssetTypes(tx)
	require.NoError(t, err)
	assetTypeNames := make([]string, 0, len(assetTypes))
	for _, assetType := range assetTypes {
		assetTypeNames = append(assetTypeNames, assetType.Name)
	}
	require.ElementsMatch(t, []string{"generic", "weblink", "3D", "Artwork", "Audio"}, assetTypeNames)
	collectionTypes, err := GetCollectionTypes(tx)
	require.NoError(t, err)
	collectionTypeNames := make([]string, 0, len(collectionTypes))
	for _, collectionType := range collectionTypes {
		collectionTypeNames = append(collectionTypeNames, collectionType.Name)
	}
	require.ElementsMatch(t, []string{"generic", "Asset", "Reference"}, collectionTypeNames)
	for _, starter := range definition.AssetTemplates {
		stored, err := GetTemplateByName(tx, starter.Name)
		require.NoError(t, err)
		require.Equal(t, filepath.Ext(starter.Source), stored.Extension)
		outputPath := filepath.Join(t.TempDir(), filepath.Base(starter.Source))
		require.NoError(t, RestoreFileFromChunks(tx, stored.Chunks, outputPath, 0, func(int, int, string, string) {}))
		actual, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		expected, err := bundledAssetTemplates.ReadFile("template_files/" + starter.Source)
		require.NoError(t, err)
		require.Equal(t, expected, actual, starter.Name)
	}
	_, err = GetTemplateByName(tx, "Blockbench")
	require.Error(t, err)
	_, err = GetTemplateByName(tx, "Blender Lighting")
	require.Error(t, err)
}

func TestHasDefaultTemplatesRequiresEveryBundledTemplate(t *testing.T) {
	directory := t.TempDir()
	definitions, err := loadTemplateDefinitions()
	require.NoError(t, err)
	for _, definition := range definitions.Templates {
		complete, err := hasDefaultTemplates(directory)
		require.NoError(t, err)
		require.False(t, complete)
		createAssetTemplateTestProject(t, filepath.Join(directory, definition.Name+".clst"))
	}
	complete, err := hasDefaultTemplates(directory)
	require.NoError(t, err)
	require.True(t, complete)
}

func createAssetTemplateTestProject(t *testing.T, path string) {
	t.Helper()
	db, err := utils.OpenDb(path)
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec(ProjectSchema)
	require.NoError(t, err)
	tx, err := db.Beginx()
	require.NoError(t, err)
	defer tx.Rollback()
	require.NoError(t, initData(tx))
	require.NoError(t, tx.Commit())
}
