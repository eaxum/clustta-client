package repository

import (
	"clustta/internal/auth_service"
	"clustta/internal/error_service"
	"clustta/internal/repository"
	"clustta/internal/testutils"
	"testing"
)

// func TestMain(m *testing.M) {
// 	testutils.Setup()
// 	testutils.GenerateFixtureCollectionType()
// 	code := m.Run()
// 	testutils.Teardown()
// 	os.Exit(code)
// }

func TestCreateCollectionType(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.CreateCollectionType(testutils.Tx, "", "Test Collection Type 2", "test")
	if err != nil {
		t.Errorf("didnt expect any error but got: %s", err.Error())
	}
}

func TestGetCollectionTypes(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.GetCollectionTypes(testutils.Tx)
	if err != nil {
		t.Errorf("didnt expect any error but got: %s", err.Error())
	}
}

func TestCreateCollection(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureCollectionType()
	defer testutils.Teardown()
	collection, err := repository.CreateCollection(testutils.Tx, "", "Test Collection 2", "", testutils.CollectionType.Id, "", "", true)
	if err != nil {
		t.Error(err.Error())
	}

	_, err = repository.CreateCollection(testutils.Tx, "", "Test Collection 2", "", testutils.CollectionType.Id, collection.Id, "", true)
	if err != nil {
		t.Error(err.Error())
	}

	_, err = repository.CreateCollection(testutils.Tx, "", "Test Collection 2", "", testutils.CollectionType.Id, collection.Id, "", true)
	if err == nil {
		t.Error("expected error of collection exist, but got none")
	} else {
		if err != error_service.ErrCollectionExists {
			t.Error(err.Error())
		}
	}

}
func TestRenameCollection(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureCollectionType()
	defer testutils.Teardown()
	collection, err := repository.CreateCollection(testutils.Tx, "", "Test", "", testutils.CollectionType.Id, "", "", true)
	if err != nil {
		t.Error(err.Error())
	}

	collection, err = repository.RenameCollection(testutils.Tx, collection.Id, "Renamed")
	if err != nil {
		t.Error(err.Error())
	}

}
func TestDeleteCollection(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureCollectionType()
	defer testutils.Teardown()
	collection, err := repository.CreateCollection(testutils.Tx, "", "Test", "", testutils.CollectionType.Id, "", "", true)
	if err != nil {
		t.Error(err.Error())
	}

	err = repository.DeleteCollection(testutils.Tx, collection.Id, true, false)
	if err != nil {
		t.Error(err.Error())
	}

	trashs, err := repository.GetDeletedCollections(testutils.Tx)
	if err != nil {
		t.Error(err.Error())
	}
	if len(trashs) > 0 {
		t.Errorf("expected trashs to be none, but got: %v", trashs)
	}

}

func TestGetCollections(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureCollectionType()
	defer testutils.Teardown()
	collection, err := repository.CreateCollection(testutils.Tx, "", "Test Collection 2", "", testutils.CollectionType.Id, "", "", true)
	if err != nil {
		t.Error(err.Error())
	}

	_, err = repository.CreateCollection(testutils.Tx, "", "Test Collection 2", "", testutils.CollectionType.Id, collection.Id, "", true)
	if err != nil {
		t.Error(err.Error())
	}
	collections, err := repository.GetCollections(testutils.Tx, false)
	if err != nil {
		t.Error(err.Error())
	}
	expectedCollectionCount := 2
	if len(collections) != expectedCollectionCount {
		t.Errorf("expected %d collections, but got %d", expectedCollectionCount, len(collections))
	}

}

func TestGetCollectionByFilePath(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureCollectionType()
	defer testutils.Teardown()
	collection, err := repository.CreateCollection(testutils.Tx, "", "Test Collection 2", "", testutils.CollectionType.Id, "", "", true)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.GetCollectionByPath(testutils.Tx, collection.CollectionPath)
	if err != nil {
		t.Error(err.Error())
	}

}

func TestGetUserCollections(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureCollectionType()
	defer testutils.Teardown()
	_, err := repository.CreateCollection(testutils.Tx, "", "Test Collection 2", "", testutils.CollectionType.Id, "", "", true)
	if err != nil {
		t.Error(err.Error())
	}
	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	userAssetInfo, err := repository.GetUserAssetsMinimal(testutils.Tx, user.Id)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.GetUserCollections(testutils.Tx, userAssetInfo, user.Id)
	if err != nil {
		t.Error(err.Error())
	}

}
