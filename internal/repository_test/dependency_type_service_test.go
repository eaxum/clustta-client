package repository

import (
	"clustta/internal/error_service"
	"clustta/internal/repository"
	"clustta/internal/testutils"
	"testing"
)

func TestCreateDependencyType(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.CreateDependencyType(testutils.Tx, "", "dependency type 1")
	if err != nil {
		t.Errorf(err.Error())
	}
}
func TestGetDependencyType(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.GetDependencyType(testutils.Tx, "not found")
	if err == nil {
		t.Error("expected error of task dependency type not found, but got none")
	} else {
		if err != error_service.ErrDependencyTypeNotFound {
			t.Error(err.Error())
		}
	}
}

func TestGetDependencyTypes(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.GetDependencyTypes(testutils.Tx)
	if err != nil {
		t.Errorf(err.Error())
	}
}
func TestRenameDependencyType(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	dependencyType, err := repository.CreateDependencyType(testutils.Tx, "", "dependency type 1")
	if err != nil {
		t.Errorf(err.Error())
	}
	err = repository.RenameDependencyType(testutils.Tx, dependencyType.Id, "renamed")
	if err != nil {
		t.Errorf(err.Error())
	}
	renamedependencyType, err := repository.GetDependencyTypeByName(testutils.Tx, "renamed")
	if err != nil {
		t.Errorf(err.Error())
	} else {
		if renamedependencyType.Name != "renamed" {
			t.Errorf("expected dependency type name to be 'renamed', but got '%s'", renamedependencyType.Name)
		}
	}
}

func TestDeleteDependencyType(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	dependencyType, err := repository.CreateDependencyType(testutils.Tx, "", "dependency type 1")
	if err != nil {
		t.Errorf(err.Error())
	}
	err = repository.DeleteDependencyType(testutils.Tx, dependencyType.Id)
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = repository.GetDependencyTypeByName(testutils.Tx, dependencyType.Name)
	if err == nil {
		t.Error("expected error of task dependency type not found, but got none")
	} else {
		if err != error_service.ErrDependencyTypeNotFound {
			t.Error(err.Error())
		}
	}
}

func TestGetOrCreateDependencyType(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.GetOrCreateDependencyType(testutils.Tx, "new")
	if err != nil {
		t.Errorf("unexpected error %v", err.Error())
	} else {
		if err == error_service.ErrDependencyTypeNotFound {
			t.Errorf("expected no error, if not found , it should create it")
		}
	}
}
