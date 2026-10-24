package services

import (
	"os"
	"path/filepath"
	"testing"

	"clustta/internal/integrations"
)

func TestAYONWorkfileExternalIDNormalizesApplication(t *testing.T) {
	actual := ayonWorkfileExternalID(" task-id ", " Blender ")
	if actual != "task-id::blender" {
		t.Fatalf("expected normalized workfile ID, got %q", actual)
	}
}

func TestAYONPrepareAndImportCreatesNativeWorkfileMapping(t *testing.T) {
	projectPath, _ := setupTrackAssetProject(t)
	context := AYONWorkfileContext{
		ProjectID:  "ayon-project",
		FolderID:   "hero-folder",
		FolderPath: "/assets/hero/",
		Folders: []AYONFolderContext{
			{
				ID:   "assets-folder",
				Name: "assets",
				Path: "/assets/",
				Type: "folder",
			},
			{
				ID:       "hero-folder",
				ParentID: "assets-folder",
				Name:     "hero",
				Path:     "/assets/hero/",
				Type:     "asset",
			},
		},
		TaskID:      "modeling-task",
		TaskName:    "Modeling",
		TaskTypeID:  "modeling",
		TaskType:    "Modeling",
		Application: "blender",
		Extension:   ".blend",
	}
	service := &IntegrationService{}
	preparation, err := service.PrepareAYONWorkfile(projectPath, context)
	if err != nil {
		t.Fatal(err)
	}
	if !preparation.RequiresImport || preparation.Asset != nil {
		t.Fatalf("unexpected preparation: %#v", preparation)
	}
	if err = os.WriteFile(preparation.TargetPath, []byte("blend-data"), 0o600); err != nil {
		t.Fatal(err)
	}
	imported, err := service.ImportAYONWorkfile(
		projectPath,
		context,
		preparation.TargetPath,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !imported.Created || imported.Asset.Extension != ".blend" {
		t.Fatalf("unexpected imported asset: %#v", imported)
	}
	if imported.Checkpoint.AssetId != imported.Asset.Id {
		t.Fatalf("unexpected initial checkpoint: %#v", imported.Checkpoint)
	}

	repeated, err := service.PrepareAYONWorkfile(projectPath, context)
	if err != nil {
		t.Fatal(err)
	}
	if repeated.Asset == nil || repeated.Asset.Id != imported.Asset.Id {
		t.Fatalf("expected existing mapped asset, got %#v", repeated)
	}
}

func TestAYONScopedWorkfileExternalIDIncludesProjectAndApplication(t *testing.T) {
	context := AYONWorkfileContext{
		ProjectID:   "project-id",
		TaskID:      "task-id",
		Application: "Blender",
	}
	actual := ayonScopedWorkfileExternalID(context)
	if actual != "project-id::task-id::blender" {
		t.Fatalf("unexpected scoped workfile ID %q", actual)
	}
}

func TestAYONTemplateIDPrefersTaskTypeID(t *testing.T) {
	options := integrations.SyncOptions{
		AssetTypeTemplates: map[string]string{
			"modeling-id": "id-template",
			"Modeling":    "name-template",
		},
	}

	actual := ayonTemplateID(options, "modeling-id", "Modeling")
	if actual != "id-template" {
		t.Fatalf("expected ID template, got %q", actual)
	}
}

func TestValidateAYONWorkfileContextRequiresTaskType(t *testing.T) {
	context := AYONWorkfileContext{
		ProjectID:   "project",
		FolderID:    "folder",
		TaskID:      "task",
		TaskName:    "modeling",
		Application: "blender",
	}

	if err := validateAYONWorkfileContext(context); err == nil {
		t.Fatal("expected missing task type error")
	}
}

func TestAYONTargetPathsUseFolderHierarchy(t *testing.T) {
	workingDirectory := t.TempDir()
	context := AYONWorkfileContext{
		FolderPath: "/assets/characters/hero/",
		TaskName:   "Modeling",
		Extension:  ".blend",
	}
	targetPath, relativePath, err := ayonTargetPaths(workingDirectory, context)
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(
		workingDirectory,
		"assets",
		"characters",
		"hero",
		"Modeling.blend",
	)
	if targetPath != expected {
		t.Fatalf("expected target %q, got %q", expected, targetPath)
	}
	if relativePath != "assets/characters/hero/Modeling.blend" {
		t.Fatalf("unexpected relative path %q", relativePath)
	}
}

func TestNormalizedAYONFoldersUseSelectedFolderID(t *testing.T) {
	context := AYONWorkfileContext{
		ProjectID:  "project-id",
		FolderID:   "hero-id",
		FolderPath: "/assets/hero/",
	}
	folders := normalizedAYONFolders(context)
	if len(folders) != 2 {
		t.Fatalf("expected two folders, got %d", len(folders))
	}
	if folders[1].ID != "hero-id" || folders[1].ParentID != folders[0].ID {
		t.Fatalf("unexpected folder hierarchy: %#v", folders)
	}
}
