package repository

import (
	"clustta/internal/repository"
	"clustta/internal/testutils"
	"testing"
)

func TestCreateStatus(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	newstatus, err := repository.CreateStatus(testutils.Tx, "", "new status", "pink", "nst")
	if err != nil {
		t.Errorf(err.Error())
	}
	status, err := repository.GetStatus(testutils.Tx, newstatus.Id)
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = repository.GetStatusByShortName(testutils.Tx, status.ShortName)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestGetStatuses(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.GetStatuses(testutils.Tx)
	if err != nil {
		t.Errorf(err.Error())
	}
}
func TestUpdatestatus(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTaskType()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureEntityType()
	testutils.GenerateFixtureEntity()
	testutils.GenerateFixtureTask()
	defer testutils.Teardown()
	newstatus, err := repository.CreateStatus(testutils.Tx, "", "new status", "pink", "nst")
	if err != nil {
		t.Errorf(err.Error())
	}
	err = repository.Updatestatus(testutils.Tx, testutils.Task.Id, newstatus.Id)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestGetOrCreateStatus(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.GetOrCreateStatus(testutils.Tx, "new status", "st", "yellow")
	if err != nil {
		t.Errorf(err.Error())
	}
}
