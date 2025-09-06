// testutils/setup.go
package testutils

import (
	"clustta/internal/auth_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var TestFolder = getAbsolutePath("../../tests")
var TestFile = getAbsolutePath("../testutils/place_holder.blend")
var TestFileModified = getAbsolutePath("../testutils/place_holder_modified.blend")
var TestPreviewFile = getAbsolutePath("../testutils/image.png")

var DbConn *sqlx.DB
var Tx *sqlx.Tx

var Project repository.ProjectInfo
var Template models.Template
var DependencyType models.DependencyType
var EntityType models.EntityType
var Entity models.Entity
var TaskType models.TaskType
var Task models.Task

func getAbsolutePath(file_path string) string {
	file, err := filepath.Abs(file_path)
	if err != nil {
		panic(err)
	}
	return file
}

func GenerateFixtureProject() {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		panic(err)
	}
	projectUri := filepath.Join(TestFolder, "test.clst")
	projectInfo, err := repository.CreateProject(projectUri, "test", "", "No Template", user)
	if err != nil {
		panic(err)
	}
	Project = projectInfo
}
func GenerateFixtureTemplate() {
	template, err := repository.CreateTemplate(Tx, "Basic", TestFile)
	if err != nil {
		panic(err)
	}
	Template = template
}
func GenerateFixtureEntityType() {
	entityType, err := repository.CreateEntityType(Tx, "", "Test Entity Type", "test")
	if err != nil {
		panic(err)
	}
	EntityType = entityType
}
func GenerateFixtureEntity() {
	entity, err := repository.CreateEntity(Tx, "", "Test Entity", "", EntityType.Id, "", "", true)
	if err != nil {
		panic(err)
	}
	Entity = entity
}
func GenerateFixtureTaskType() {
	taskType, err := repository.CreateTaskType(Tx, "", "Test Task Type", "test")
	if err != nil {
		panic(err)
	}
	TaskType = taskType
}
func GenerateFixtureTask() {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		panic(err)
	}
	task, err := repository.CreateTask(Tx, "", "Test Task", TaskType.Id, Entity.Id, false, Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(int, int, string, string) {})
	if err != nil {
		panic(err)
	}
	Task = task
}
func GenerateFixtureDependencyType() {
	dependencyType, err := repository.CreateDependencyType(Tx, "", "Test")
	if err != nil {
		panic(err)
	}
	DependencyType = dependencyType
}

func Setup() {
	err := os.RemoveAll(TestFolder)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(TestFolder, os.ModePerm)
	if err != nil {
		panic(err)
	}
	GenerateFixtureProject()
	DbConn, err = utils.OpenDb(Project.Uri)
	if err != nil {
		panic(err)
	}
	Tx, err = DbConn.Beginx()
	if err != nil {
		panic(err)
	}
}

func Teardown() {
	Tx.Rollback()
	DbConn.Close()

	err := os.RemoveAll(TestFolder)
	if err != nil {
		panic(err)
	}
}
