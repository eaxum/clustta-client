package repository

import (
	"clustta/internal/auth_service"
	"clustta/internal/repository"
	"clustta/internal/testutils"
	"testing"

	"github.com/google/uuid"
)

// func TestMain(m *testing.M) {
// 	testutils.Setup()
// 	testutils.GenerateFixtureTemplate()
// 	testutils.GenerateFixtureCollectionType()
// 	testutils.GenerateFixtureCollection()
// 	testutils.GenerateFixtureAssetType()
// 	code := m.Run()
// 	testutils.Teardown()
// 	os.Exit(code)
// }

func TestCreateAssetType(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.CreateAssetType(testutils.Tx, "", "asset type 1", "test")
	if err != nil {
		t.Errorf("didnt expect any error but got: %s", err.Error())
	}
}
func TestDeleteAssetType(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	assetType, err := repository.CreateAssetType(testutils.Tx, "", "asset type 1", "test")
	if err != nil {
		t.Errorf("didnt expect any error but got: %s", err.Error())
	}
	err = repository.DeleteAssetType(testutils.Tx, assetType.Id)
	if err != nil {
		t.Errorf("didnt expect any error but got: %s", err.Error())
	}

}
func TestGetAssetTypes(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.GetAssetTypes(testutils.Tx)
	if err != nil {
		t.Errorf("didnt expect any error but got: %s", err.Error())
	}
}

func TestCreateAsset(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureCollectionType()
	testutils.GenerateFixtureCollection()
	testutils.GenerateFixtureAssetType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.CreateAsset(
		testutils.Tx, "", "asset 2", testutils.AssetType.Id, testutils.Collection.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

}
func TestCreateRootAsset(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureCollectionType()
	testutils.GenerateFixtureCollection()
	testutils.GenerateFixtureAssetType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.CreateAsset(
		testutils.Tx, "", "asset 2", testutils.AssetType.Id, "", false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

}
func TestGetAsset(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureCollectionType()
	testutils.GenerateFixtureCollection()
	testutils.GenerateFixtureAssetType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	asset, err := repository.CreateAsset(
		testutils.Tx, "", "asset", testutils.AssetType.Id, testutils.Collection.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.GetAsset(testutils.Tx, asset.Id)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.GetAssetByName(testutils.Tx, asset.Name, asset.CollectionId, asset.Extension)
	if err != nil {
		t.Error(err.Error())
	}

}
func TestGetUserAssets(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureCollectionType()
	testutils.GenerateFixtureCollection()
	testutils.GenerateFixtureAssetType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	users, err := repository.GetUsers(testutils.Tx)
	if err != nil {
		t.Error(err.Error())
	}
	asset, err := repository.CreateAsset(
		testutils.Tx, "", "asset", testutils.AssetType.Id, testutils.Collection.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}
	err = repository.AssignAsset(testutils.Tx, asset.Id, users[0].Id)
	if err != nil {
		t.Error(err.Error())
	}
	userAssets, err := repository.GetUserAssets(testutils.Tx, users[0].Id)
	if err != nil {
		t.Error(err.Error())
	}
	if len(userAssets) < 1 {
		t.Errorf("Expected 1 asset, got %d", len(userAssets))
	}
	err = repository.UnAssignAsset(testutils.Tx, asset.Id)
	if err != nil {
		t.Error(err.Error())
	}

	userAssets, err = repository.GetUserAssets(testutils.Tx, users[0].Id)
	if err != nil {
		t.Error(err.Error())
	}
	if len(userAssets) != 0 {
		t.Errorf("Expected 0 asset, got %d", len(userAssets))
	}

}
func TestGetAssets(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureCollectionType()
	testutils.GenerateFixtureCollection()
	testutils.GenerateFixtureAssetType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}

	assets, err := repository.GetAssets(testutils.Tx, true)
	if err != nil {
		t.Error(err.Error())
	}
	expectedCount := 0
	if len(assets) != expectedCount {
		t.Errorf("Expected %d assets, got %d", expectedCount, len(assets))
	}

	_, err = repository.CreateAsset(
		testutils.Tx, "", "asset", testutils.AssetType.Id, testutils.Collection.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

	assets, err = repository.GetAssets(testutils.Tx, true)
	if err != nil {
		t.Error(err.Error())
	}
	expectedCount = 1
	if len(assets) != expectedCount {
		t.Errorf("Expected %d assets, got %d", expectedCount, len(assets))
	}

}
func TestDeleteAsset(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureCollectionType()
	testutils.GenerateFixtureCollection()
	testutils.GenerateFixtureAssetType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}

	asset, err := repository.CreateAsset(
		testutils.Tx, "", "asset 2", testutils.AssetType.Id, testutils.Collection.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}
	err = repository.DeleteAsset(testutils.Tx, asset.Id, true, true)
	if err != nil {
		t.Error(err.Error())
	}

}
func TestUpdateAsset(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureCollectionType()
	testutils.GenerateFixtureCollection()
	testutils.GenerateFixtureAssetType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}

	asset, err := repository.CreateAsset(
		testutils.Tx, "", "asset 2", testutils.AssetType.Id, testutils.Collection.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}
	updatedAsset, err := repository.UpdateAsset(testutils.Tx, asset.Id, "renamed", asset.AssetTypeId, asset.IsResource, asset.Pointer, []string{})
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.GetAssetByName(testutils.Tx, updatedAsset.Name, updatedAsset.CollectionId, updatedAsset.Extension)
	if err != nil {
		t.Error(err.Error())
	}

}

func TestUpdateAssetStatus(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureCollectionType()
	testutils.GenerateFixtureCollection()
	testutils.GenerateFixtureAssetType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}

	asset, err := repository.CreateAsset(
		testutils.Tx, "", "asset 2", testutils.AssetType.Id, testutils.Collection.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}
	status, err := repository.GetStatusByShortName(testutils.Tx, "wip")
	if err != nil {
		t.Error(err.Error())
	}
	err = repository.UpdateStatus(testutils.Tx, asset.Id, status.Id)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestDeleteCollectionAssets(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureCollectionType()
	testutils.GenerateFixtureCollection()
	testutils.GenerateFixtureAssetType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.CreateAsset(
		testutils.Tx, "", "asset 2", testutils.AssetType.Id, testutils.Collection.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

	_, err = repository.CreateAsset(
		testutils.Tx, "", "asset 3", testutils.AssetType.Id, testutils.Collection.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

	err = repository.DeleteCollectionAssets(testutils.Tx, testutils.Collection.Id, true)
	if err != nil {
		t.Error(err.Error())
	}
	collectionAssets, err := repository.GetAssetsByCollectionId(testutils.Tx, testutils.Collection.Id)
	if err != nil {
		t.Error(err.Error())
	}
	if len(collectionAssets) != 0 {
		t.Errorf("Expected 0 assets of collection, got %d", len(collectionAssets))
	}

}

func TestAddAssetDependency(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureCollectionType()
	testutils.GenerateFixtureCollection()
	testutils.GenerateFixtureAssetType()
	testutils.GenerateFixtureDependencyType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	assetA, err := repository.CreateAsset(
		testutils.Tx, "", "asset A", testutils.AssetType.Id, testutils.Collection.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}
	assetB, err := repository.CreateAsset(
		testutils.Tx, "", "asset B", testutils.AssetType.Id, testutils.Collection.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

	_, err = repository.AddDependency(testutils.Tx, "", assetA.Id, assetB.Id, testutils.DependencyType.Id)
	if err != nil {
		t.Error(err.Error())
	}

}
func TestRemoveAssetDependency(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureCollectionType()
	testutils.GenerateFixtureCollection()
	testutils.GenerateFixtureAssetType()
	testutils.GenerateFixtureDependencyType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	assetA, err := repository.CreateAsset(
		testutils.Tx, "", "asset A", testutils.AssetType.Id, testutils.Collection.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}
	assetB, err := repository.CreateAsset(
		testutils.Tx, "", "asset B", testutils.AssetType.Id, testutils.Collection.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

	_, err = repository.AddDependency(testutils.Tx, "", assetA.Id, assetB.Id, testutils.DependencyType.Id)
	if err != nil {
		t.Error(err.Error())
	}
	err = repository.RemoveAssetDependency(testutils.Tx, assetA.Id, assetB.Id)
	if err != nil {
		t.Error(err.Error())
	}

}
func TestReAddAssetDependency(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureCollectionType()
	testutils.GenerateFixtureCollection()
	testutils.GenerateFixtureAssetType()
	testutils.GenerateFixtureDependencyType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	assetA, err := repository.CreateAsset(
		testutils.Tx, "", "asset A", testutils.AssetType.Id, testutils.Collection.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}
	assetB, err := repository.CreateAsset(
		testutils.Tx, "", "asset B", testutils.AssetType.Id, testutils.Collection.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

	_, err = repository.AddDependency(testutils.Tx, "", assetA.Id, assetB.Id, testutils.DependencyType.Id)
	if err != nil {
		t.Error(err.Error())
	}
	err = repository.RemoveAssetDependency(testutils.Tx, assetA.Id, assetB.Id)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.AddDependency(testutils.Tx, "", assetA.Id, assetB.Id, testutils.DependencyType.Id)
	if err != nil {
		t.Error(err.Error())
	}

}
