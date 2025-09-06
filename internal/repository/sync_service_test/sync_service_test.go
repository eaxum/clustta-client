package sync_service

import (
	"clustta/internal/auth_service"
	"clustta/internal/error_service"
	"clustta/internal/repository"
	"clustta/internal/repository/sync_service"
	"clustta/internal/testutils"
	"context"
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

func TestSyncProject(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
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

}

func TestGetLocalStudioProject(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	testutils.Tx.Commit()
	_, err = sync_service.GetStudioProjects(user, testutils.TestFolder, "Personal")
	if err != nil {
		t.Error(err.Error())
	}

}
func TestGetStudioProject(t *testing.T) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}

	_, err = sync_service.GetStudioProjects(user, "http://localhost:8080", "Test")
	if err != nil {
		t.Error(err.Error())
	}

}

func TestGetCloneProject(t *testing.T) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}

	projects, err := sync_service.GetStudioProjects(user, "http://localhost:7774", "Eaxum")
	if err != nil {
		t.Error(err.Error())
	}
	project := projects[0]
	syncOptions := sync_service.SyncOptions{
		OnlyLatestCheckpoints: true,
		TaskDependencies:      true,
		Tasks:                 true,
		Resources:             true,
		Templates:             true,
	}
	err = sync_service.CloneProject(context.TODO(), project.Remote, project.Uri, "Eaxum", "", user, syncOptions, func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

}
