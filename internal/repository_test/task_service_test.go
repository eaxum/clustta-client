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
// 	testutils.GenerateFixtureEntityType()
// 	testutils.GenerateFixtureEntity()
// 	testutils.GenerateFixtureTaskType()
// 	code := m.Run()
// 	testutils.Teardown()
// 	os.Exit(code)
// }

func TestCreateTaskType(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.CreateTaskType(testutils.Tx, "", "task type 1", "test")
	if err != nil {
		t.Errorf("didnt expect any error but got: %s", err.Error())
	}
}
func TestDeleteTaskType(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	taskType, err := repository.CreateTaskType(testutils.Tx, "", "task type 1", "test")
	if err != nil {
		t.Errorf("didnt expect any error but got: %s", err.Error())
	}
	err = repository.DeleteTaskType(testutils.Tx, taskType.Id)
	if err != nil {
		t.Errorf("didnt expect any error but got: %s", err.Error())
	}

}
func TestGetTaskTypes(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.GetTaskTypes(testutils.Tx)
	if err != nil {
		t.Errorf("didnt expect any error but got: %s", err.Error())
	}
}

func TestCreateTask(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureEntityType()
	testutils.GenerateFixtureEntity()
	testutils.GenerateFixtureTaskType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.CreateTask(
		testutils.Tx, "", "task 2", testutils.TaskType.Id, testutils.Entity.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

}
func TestCreateRootTask(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureEntityType()
	testutils.GenerateFixtureEntity()
	testutils.GenerateFixtureTaskType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.CreateTask(
		testutils.Tx, "", "task 2", testutils.TaskType.Id, "", false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

}
func TestGetTask(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureEntityType()
	testutils.GenerateFixtureEntity()
	testutils.GenerateFixtureTaskType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	task, err := repository.CreateTask(
		testutils.Tx, "", "task", testutils.TaskType.Id, testutils.Entity.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.GetTask(testutils.Tx, task.Id)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.GetTaskByName(testutils.Tx, task.Name, task.EntityId, task.Extension)
	if err != nil {
		t.Error(err.Error())
	}

}
func TestGetUserTasks(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureEntityType()
	testutils.GenerateFixtureEntity()
	testutils.GenerateFixtureTaskType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	users, err := repository.GetUsers(testutils.Tx)
	if err != nil {
		t.Error(err.Error())
	}
	task, err := repository.CreateTask(
		testutils.Tx, "", "task", testutils.TaskType.Id, testutils.Entity.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}
	err = repository.AssignTask(testutils.Tx, task.Id, users[0].Id)
	if err != nil {
		t.Error(err.Error())
	}
	userTasks, err := repository.GetUserTasks(testutils.Tx, users[0].Id)
	if err != nil {
		t.Error(err.Error())
	}
	if len(userTasks) < 1 {
		t.Errorf("Expected 1 task, got %d", len(userTasks))
	}
	err = repository.UnAssignTask(testutils.Tx, task.Id)
	if err != nil {
		t.Error(err.Error())
	}

	userTasks, err = repository.GetUserTasks(testutils.Tx, users[0].Id)
	if err != nil {
		t.Error(err.Error())
	}
	if len(userTasks) != 0 {
		t.Errorf("Expected 0 task, got %d", len(userTasks))
	}

}
func TestGetTasks(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureEntityType()
	testutils.GenerateFixtureEntity()
	testutils.GenerateFixtureTaskType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}

	tasks, err := repository.GetTasks(testutils.Tx, true)
	if err != nil {
		t.Error(err.Error())
	}
	expectedCount := 0
	if len(tasks) != expectedCount {
		t.Errorf("Expected %d tasks, got %d", expectedCount, len(tasks))
	}

	_, err = repository.CreateTask(
		testutils.Tx, "", "task", testutils.TaskType.Id, testutils.Entity.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

	tasks, err = repository.GetTasks(testutils.Tx, true)
	if err != nil {
		t.Error(err.Error())
	}
	expectedCount = 1
	if len(tasks) != expectedCount {
		t.Errorf("Expected %d tasks, got %d", expectedCount, len(tasks))
	}

}
func TestDeleteTask(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureEntityType()
	testutils.GenerateFixtureEntity()
	testutils.GenerateFixtureTaskType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}

	task, err := repository.CreateTask(
		testutils.Tx, "", "task 2", testutils.TaskType.Id, testutils.Entity.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}
	err = repository.DeleteTask(testutils.Tx, task.Id, true, true)
	if err != nil {
		t.Error(err.Error())
	}

}
func TestUpdateTask(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureEntityType()
	testutils.GenerateFixtureEntity()
	testutils.GenerateFixtureTaskType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}

	task, err := repository.CreateTask(
		testutils.Tx, "", "task 2", testutils.TaskType.Id, testutils.Entity.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}
	updatedTask, err := repository.UpdateTask(testutils.Tx, task.Id, "renamed", task.TaskTypeId, task.IsResource, task.Pointer, []string{})
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = repository.GetTaskByName(testutils.Tx, updatedTask.Name, updatedTask.EntityId, updatedTask.Extension)
	if err != nil {
		t.Error(err.Error())
	}

}

func TestUpdateTaskStatus(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureEntityType()
	testutils.GenerateFixtureEntity()
	testutils.GenerateFixtureTaskType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}

	task, err := repository.CreateTask(
		testutils.Tx, "", "task 2", testutils.TaskType.Id, testutils.Entity.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}
	status, err := repository.GetStatusByShortName(testutils.Tx, "wip")
	if err != nil {
		t.Errorf(err.Error())
	}
	err = repository.UpdateStatus(testutils.Tx, task.Id, status.Id)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestDeleteEntityTasks(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureEntityType()
	testutils.GenerateFixtureEntity()
	testutils.GenerateFixtureTaskType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.CreateTask(
		testutils.Tx, "", "task 2", testutils.TaskType.Id, testutils.Entity.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

	_, err = repository.CreateTask(
		testutils.Tx, "", "task 3", testutils.TaskType.Id, testutils.Entity.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

	err = repository.DeleteEntityTasks(testutils.Tx, testutils.Entity.Id, true)
	if err != nil {
		t.Error(err.Error())
	}
	entityTasks, err := repository.GetTasksByEntityId(testutils.Tx, testutils.Entity.Id)
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(entityTasks) != 0 {
		t.Errorf("Expected 0 tasks of entity, got %d", len(entityTasks))
	}

}

func TestAddTaskDependency(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureEntityType()
	testutils.GenerateFixtureEntity()
	testutils.GenerateFixtureTaskType()
	testutils.GenerateFixtureDependencyType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	taskA, err := repository.CreateTask(
		testutils.Tx, "", "task A", testutils.TaskType.Id, testutils.Entity.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}
	taskB, err := repository.CreateTask(
		testutils.Tx, "", "task B", testutils.TaskType.Id, testutils.Entity.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

	_, err = repository.AddDependency(testutils.Tx, "", taskA.Id, taskB.Id, testutils.DependencyType.Id)
	if err != nil {
		t.Error(err.Error())
	}

}
func TestRemoveTaskDependency(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureEntityType()
	testutils.GenerateFixtureEntity()
	testutils.GenerateFixtureTaskType()
	testutils.GenerateFixtureDependencyType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	taskA, err := repository.CreateTask(
		testutils.Tx, "", "task A", testutils.TaskType.Id, testutils.Entity.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}
	taskB, err := repository.CreateTask(
		testutils.Tx, "", "task B", testutils.TaskType.Id, testutils.Entity.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

	_, err = repository.AddDependency(testutils.Tx, "", taskA.Id, taskB.Id, testutils.DependencyType.Id)
	if err != nil {
		t.Error(err.Error())
	}
	err = repository.RemoveTaskDependency(testutils.Tx, taskA.Id, taskB.Id)
	if err != nil {
		t.Error(err.Error())
	}

}
func TestReAddTaskDependency(t *testing.T) {
	testutils.Setup()
	testutils.GenerateFixtureTemplate()
	testutils.GenerateFixtureEntityType()
	testutils.GenerateFixtureEntity()
	testutils.GenerateFixtureTaskType()
	testutils.GenerateFixtureDependencyType()
	defer testutils.Teardown()

	user, err := auth_service.GetActiveUser()
	if err != nil {
		t.Error(err.Error())
	}
	taskA, err := repository.CreateTask(
		testutils.Tx, "", "task A", testutils.TaskType.Id, testutils.Entity.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}
	taskB, err := repository.CreateTask(
		testutils.Tx, "", "task B", testutils.TaskType.Id, testutils.Entity.Id, false,
		testutils.Template.Id, "", "", []string{}, "", false, "", user.Id, "new file", uuid.New().String(), func(i1, i2 int, s1, s2 string) {})
	if err != nil {
		t.Error(err.Error())
	}

	_, err = repository.AddDependency(testutils.Tx, "", taskA.Id, taskB.Id, testutils.DependencyType.Id)
	if err != nil {
		t.Error(err.Error())
	}
	err = repository.RemoveTaskDependency(testutils.Tx, taskA.Id, taskB.Id)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = repository.AddDependency(testutils.Tx, "", taskA.Id, taskB.Id, testutils.DependencyType.Id)
	if err != nil {
		t.Error(err.Error())
	}

}
