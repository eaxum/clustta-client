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
// 	testutils.GenerateFixtureEntityType()
// 	code := m.Run()
// 	testutils.Teardown()
// 	os.Exit(code)
// }

func TestCreateEntityType(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.CreateEntityType(testutils.Tx, "", "Test Entity Type 2", "test")
	if err != nil {
		t.Errorf("didnt expect any error but got: %s", err.Error())
	}
}

func TestGetEntityTypes(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.GetEntityTypes(testutils.Tx)
	if err != nil {
		t.Errorf("didnt expect any error but got: %s", err.Error())
	}
}

func TestCreateEntity(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureEntityType()
	defer testutils.Teardown()
	entity, err := repository.CreateEntity(testutils.Tx, "", "Test Entity 2", "", testutils.EntityType.Id, "", "", true)
	if err != nil {
		t.Error(err.Error())
	}

	_, err = repository.CreateEntity(testutils.Tx, "", "Test Entity 2", "", testutils.EntityType.Id, entity.Id, "", true)
	if err != nil {
		t.Error(err.Error())
	}

	_, err = repository.CreateEntity(testutils.Tx, "", "Test Entity 2", "", testutils.EntityType.Id, entity.Id, "", true)
	if err == nil {
		t.Error("expected error of entity exist, but got none")
	} else {
		if err != error_service.ErrEntityExists {
			t.Error(err.Error())
		}
	}

}
func TestRenameEntity(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureEntityType()
	defer testutils.Teardown()
	entity, err := repository.CreateEntity(testutils.Tx, "", "Test", "", testutils.EntityType.Id, "", "", true)
	if err != nil {
		t.Error(err.Error())
	}

	entity, err = repository.RenameEntity(testutils.Tx, entity.Id, "Renamed")
	if err != nil {
		t.Error(err.Error())
	}

}
func TestDeleteEntity(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureEntityType()
	defer testutils.Teardown()
	entity, err := repository.CreateEntity(testutils.Tx, "", "Test", "", testutils.EntityType.Id, "", "", true)
	if err != nil {
		t.Error(err.Error())
	}

	err = repository.DeleteEntity(testutils.Tx, entity.Id, true, false)
	if err != nil {
		t.Error(err.Error())
	}

	trashs, err := repository.GetDeletedEntities(testutils.Tx)
	if err != nil {
		t.Error(err.Error())
	}
	if len(trashs) > 0 {
		t.Errorf("expected trashs to be none, but got: %v", trashs)
	}

}

func TestGetEntities(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureEntityType()
	defer testutils.Teardown()
	entity, err := repository.CreateEntity(testutils.Tx, "", "Test Entity 2", "", testutils.EntityType.Id, "", "", true)
	if err != nil {
		t.Error(err.Error())
	}

	_, err = repository.CreateEntity(testutils.Tx, "", "Test Entity 2", "", testutils.EntityType.Id, entity.Id, "", true)
	if err != nil {
		t.Error(err.Error())
	}
	entities, err := repository.GetEntities(testutils.Tx, false)
	if err != nil {
		t.Error(err.Error())
	}
	expectedEntityCount := 2
	if len(entities) != expectedEntityCount {
		t.Errorf("expected %d entities, but got %d", expectedEntityCount, len(entities))
	}

}

func TestGetEntityByFilePath(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureEntityType()
	defer testutils.Teardown()
	entity, err := repository.CreateEntity(testutils.Tx, "", "Test Entity 2", "", testutils.EntityType.Id, "", "", true)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.GetEntityByPath(testutils.Tx, entity.EntityPath)
	if err != nil {
		t.Error(err.Error())
	}

}

func TestGetUserEntities(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureEntityType()
	defer testutils.Teardown()
	_, err := repository.CreateEntity(testutils.Tx, "", "Test Entity 2", "", testutils.EntityType.Id, "", "", true)
	if err != nil {
		t.Error(err.Error())
	}
	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	userTaskInfo, err := repository.GetUserTasksMinimal(testutils.Tx, user.Id)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.GetUserEntities(testutils.Tx, userTaskInfo, user.Id)
	if err != nil {
		t.Error(err.Error())
	}

}
