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
		t.Error(err.Error())
	}
	status, err := repository.GetStatus(testutils.Tx, newstatus.Id)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.GetStatusByShortName(testutils.Tx, status.ShortName)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetStatuses(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.GetStatuses(testutils.Tx)
	if err != nil {
		t.Error(err.Error())
	}
}
func TestUpdatestatus(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureAssetType()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureCollectionType()
	testutils.GenerateFixtureCollection()
	testutils.GenerateFixtureAsset()
	defer testutils.Teardown()
	newstatus, err := repository.CreateStatus(testutils.Tx, "", "new status", "pink", "nst")
	if err != nil {
		t.Error(err.Error())
	}
	err = repository.Updatestatus(testutils.Tx, testutils.Asset.Id, newstatus.Id)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetOrCreateStatus(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.GetOrCreateStatus(testutils.Tx, "new status", "st", "yellow")
	if err != nil {
		t.Error(err.Error())
	}
}
