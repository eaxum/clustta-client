package services

import (
	"os"
	"path/filepath"
	"testing"

	"clustta/internal/integrations"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
)

func TestAYONWorkfileExternalIDNormalizesApplication(t *testing.T) {
	actual := ayonWorkfileExternalID(" task-id ", " Blender ")
	if actual != "task-id::blender" {
		t.Fatalf("expected normalized workfile ID, got %q", actual)
	}
}

func TestAYONPrepareAndImportCreatesNativeWorkfileMapping(t *testing.T) {
	projectPath, workingDirectory := setupTrackAssetProject(t)
	context := AYONWorkfileContext{
		ProjectID:    "ayon-project",
		FolderID:     "hero-folder",
		FolderPath:   "/assets/hero/",
		WorkfilePath: filepath.Join(workingDirectory, "assets", "hero", "aremu_modeling.blend"),
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
	if imported.Asset.Name != "aremu_modeling" {
		t.Fatalf("expected AYON filename as asset name, got %q", imported.Asset.Name)
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

func TestAYONPrepareReplacesStaleWorkfileMapping(t *testing.T) {
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

	db, err := utils.OpenDb(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(
		"UPDATE integration_asset_mapping SET asset_id = ? WHERE asset_id = ?",
		"missing-asset",
		imported.Asset.Id,
	); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}

	repaired, err := service.PrepareAYONWorkfile(projectPath, context)
	if err != nil {
		t.Fatal(err)
	}
	if !repaired.RequiresImport || repaired.Asset != nil {
		t.Fatalf("expected stale mapping to be replaced, got %#v", repaired)
	}

	db, err = utils.OpenDb(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	_, err = repository.GetAssetMappingByExternalId(
		tx,
		ayonDCCIntegrationID,
		ayonScopedWorkfileExternalID(context),
	)
	if err != repository.ErrAssetMappingNotFound {
		t.Fatalf("expected stale mapping to be removed, got %v", err)
	}
}

func TestAYONPrepareReplacesMappingToTrashedAsset(t *testing.T) {
	projectPath, _ := setupTrackAssetProject(t)
	context := AYONWorkfileContext{
		ProjectID:   "ayon-project",
		FolderID:    "hero-folder",
		FolderPath:  "/assets/hero/",
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

	db, err := utils.OpenDb(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(
		"UPDATE asset SET trashed = 1 WHERE id = ?",
		imported.Asset.Id,
	); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}

	repaired, err := service.PrepareAYONWorkfile(projectPath, context)
	if err != nil {
		t.Fatal(err)
	}
	if !repaired.RequiresImport || repaired.Asset != nil {
		t.Fatalf("expected trashed mapping to be replaced, got %#v", repaired)
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

func TestAYONTargetPathsPreferAuthoritativeWorkfilePath(t *testing.T) {
	workingDirectory := t.TempDir()
	workfilePath := filepath.Join(
		workingDirectory,
		"assets",
		"hero",
		"Hero_Modeling.blend",
	)
	context := AYONWorkfileContext{
		FolderPath:   "/ignored/",
		WorkfilePath: workfilePath,
		TaskName:     "Modeling",
		Extension:    ".blend",
	}

	targetPath, relativePath, err := ayonTargetPaths(workingDirectory, context)
	if err != nil {
		t.Fatal(err)
	}
	if targetPath != workfilePath {
		t.Fatalf("expected target %q, got %q", workfilePath, targetPath)
	}
	if relativePath != "assets/hero/Hero_Modeling.blend" {
		t.Fatalf("unexpected relative path %q", relativePath)
	}
}

func TestAYONTargetPathsRejectWorkfileOutsideProject(t *testing.T) {
	workingDirectory := t.TempDir()
	context := AYONWorkfileContext{
		FolderPath:   "/assets/hero/",
		WorkfilePath: filepath.Join(filepath.Dir(workingDirectory), "outside.blend"),
		TaskName:     "Modeling",
		Extension:    ".blend",
	}

	if _, _, err := ayonTargetPaths(workingDirectory, context); err == nil {
		t.Fatal("expected workfile outside project to be rejected")
	}
}

func TestValidateAYONAssetPathRejectsDifferentPath(t *testing.T) {
	asset := models.Asset{FilePath: filepath.Join("project", "Hero_Modeling.blend")}
	requestedPath := filepath.Join("project", "Hero_Modeling_v002.blend")

	if err := validateAYONAssetPath(asset, requestedPath); err == nil {
		t.Fatal("expected mismatched AYON workfile path to be rejected")
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
