package repository

import (
	"clustta/internal/auth_service"
	"clustta/internal/repository"
	"clustta/internal/utils"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const testIntegrationID = "kitsu"

type integrationTestData struct {
	tx           *sqlx.Tx
	assetID      string
	collectionID string
}

func TestDeleteAssetRemovesIntegrationMapping(t *testing.T) {
	testData := setupMappedAsset(t)

	if err := repository.DeleteAsset(testData.tx, testData.assetID, false, true); err != nil {
		t.Fatalf("delete asset: %v", err)
	}

	mappings, err := repository.GetAssetMappingsByAssetId(testData.tx, testData.assetID)
	if err != nil {
		t.Fatalf("get asset mappings: %v", err)
	}
	if len(mappings) != 0 {
		t.Fatalf("expected asset mapping to be removed, got %d", len(mappings))
	}
}

func TestDeleteCollectionRemovesIntegrationMapping(t *testing.T) {
	testData := setupIntegrationTest(t)
	collectionID := createCollection(t, testData.tx)
	createCollectionMapping(t, testData.tx, collectionID)

	if err := repository.DeleteCollection(testData.tx, collectionID, false, true); err != nil {
		t.Fatalf("delete collection: %v", err)
	}

	_, err := repository.GetCollectionMappingByCollectionId(testData.tx, collectionID)
	if err == nil {
		t.Fatal("expected collection mapping to be removed")
	}
}

func TestDeleteIntegrationProjectClearsMappings(t *testing.T) {
	testData := setupMappedAsset(t)
	createCollectionMapping(t, testData.tx, testData.collectionID)
	createStaleProviderMappings(t, testData)

	integration, err := repository.CreateIntegrationProject(
		testData.tx,
		uuid.New().String(),
		testIntegrationID,
		"external-project",
		"External Project",
		"https://kitsu.example.com",
		"{}",
		"user-id",
		"2026-11-07T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("create integration project: %v", err)
	}
	if err := repository.DeleteIntegrationProject(testData.tx, integration.Id); err != nil {
		t.Fatalf("delete integration project: %v", err)
	}

	assetMappings, err := repository.GetAllAssetMappings(testData.tx)
	if err != nil {
		t.Fatalf("get asset mappings: %v", err)
	}
	collectionMappings, err := repository.GetAllCollectionMappings(testData.tx)
	if err != nil {
		t.Fatalf("get collection mappings: %v", err)
	}
	if len(assetMappings) != 0 || len(collectionMappings) != 0 {
		t.Fatalf("expected all mappings to be removed, got %d asset and %d collection mappings", len(assetMappings), len(collectionMappings))
	}
	if _, err := repository.GetIntegrationProject(testData.tx); err == nil {
		t.Fatal("expected integration project to be removed")
	}
}

func createStaleProviderMappings(t *testing.T, testData integrationTestData) {
	t.Helper()
	_, err := repository.CreateAssetMapping(
		testData.tx, uuid.New().String(), "ayon", "stale-asset", "Stale Asset",
		"stale-parent", "Modeling", "wip", "[]", "{}", testData.assetID, "2026-11-07T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("create stale asset mapping: %v", err)
	}
	_, err = repository.CreateCollectionMapping(
		testData.tx, uuid.New().String(), "ayon", "stale-collection", "folder",
		"Stale Collection", "", "/stale-collection/", "{}", testData.collectionID, "2026-11-07T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("create stale collection mapping: %v", err)
	}
}

func setupMappedAsset(t *testing.T) integrationTestData {
	t.Helper()
	testData := setupIntegrationTest(t)
	testData.collectionID = createCollection(t, testData.tx)
	templatePath := filepath.Join(t.TempDir(), "template.blend")
	if err := os.WriteFile(templatePath, []byte("template"), 0o600); err != nil {
		t.Fatalf("create template file: %v", err)
	}
	template, err := repository.CreateTemplate(testData.tx, "Basic", templatePath)
	if err != nil {
		t.Fatalf("create template: %v", err)
	}
	assetType, err := repository.CreateAssetType(testData.tx, "", "Test Asset Type", "test")
	if err != nil {
		t.Fatalf("create asset type: %v", err)
	}
	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Fatalf("get active user: %v", err)
	}
	asset, err := repository.CreateAsset(
		testData.tx, "", "Test Asset", assetType.Id, testData.collectionID, false,
		template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created",
		uuid.New().String(), func(int, int, string, string) {},
	)
	if err != nil {
		t.Fatalf("create asset: %v", err)
	}
	testData.assetID = asset.Id

	_, err = repository.CreateAssetMapping(
		testData.tx, uuid.New().String(), testIntegrationID, "external-asset", "External Asset",
		"external-parent", "Modeling", "wip", "[]", "{}", asset.Id, "2026-11-07T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("create asset mapping: %v", err)
	}
	return testData
}

func setupIntegrationTest(t *testing.T) integrationTestData {
	t.Helper()
	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Fatalf("get active user: %v", err)
	}
	projectPath := filepath.Join(t.TempDir(), "integration-test.clst")
	project, err := repository.CreateProject(projectPath, "integration-test", "", "No Template", "", user)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	dbConn, err := utils.OpenDb(project.Uri)
	if err != nil {
		t.Fatalf("open project: %v", err)
	}
	tx, err := dbConn.Beginx()
	if err != nil {
		_ = dbConn.Close()
		t.Fatalf("begin transaction: %v", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback()
		_ = dbConn.Close()
	})
	return integrationTestData{tx: tx}
}

func createCollection(t *testing.T, tx *sqlx.Tx) string {
	t.Helper()
	collectionType, err := repository.CreateCollectionType(tx, "", "Test Collection Type", "test")
	if err != nil {
		t.Fatalf("create collection type: %v", err)
	}
	collection, err := repository.CreateCollection(tx, "", "Test Collection", "", collectionType.Id, "", "", true)
	if err != nil {
		t.Fatalf("create collection: %v", err)
	}
	return collection.Id
}

func createCollectionMapping(t *testing.T, tx *sqlx.Tx, collectionID string) {
	t.Helper()
	_, err := repository.CreateCollectionMapping(
		tx, uuid.New().String(), testIntegrationID, "external-collection", "shot",
		"External Collection", "", "/external-collection/", "{}", collectionID, "2026-11-07T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("create collection mapping: %v", err)
	}
}
