package repository

import (
	"clustta/internal/auth_service"
	"clustta/internal/error_service"
	"clustta/internal/repository"
	"clustta/internal/testutils"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	testutils.Setup()
	code := m.Run()
	testutils.Teardown()
	os.Exit(code)
}

func TestCreateProject(t *testing.T) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	projectUri := filepath.Join(testutils.TestFolder, "test_2.clst")
	_, err = repository.CreateProject(projectUri, "test", "", "No Template", user)
	if err != nil {
		t.Errorf("didnt expect any error but got: %s", err.Error())
	}

	verify, err := repository.VerifyProjectIntegrity(projectUri)
	if err != nil {
		t.Error(err.Error())
	}
	if !verify {
		t.Error("Project failed integrity test")
	}

	_, err = repository.CreateProject(projectUri, "test", "", "No Template", user)
	if err == nil {
		t.Error("expected error of project exist, but got none")
	} else {
		if err != error_service.ErrProjectExists {
			t.Error(err.Error())
		}
	}
	println(testutils.TestPreviewFile)
	err = repository.SetProjectPreview(testutils.Tx, testutils.TestPreviewFile)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.GetProjectPreview(testutils.Tx)
	if err != nil {
		t.Error(err.Error())
	}

}
