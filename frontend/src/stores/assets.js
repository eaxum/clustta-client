import { defineStore } from "pinia";
import { isValidWeblink, isValidPointer } from "@/lib/pointer";
import { metadata } from "@/lib/metadata";
import { useCollectionStore } from "./collections";
import { TaskService, TagService } from "@/../bindings/clustta/services";
import { useIconStore } from "./icons";
import utils from "@/services/utils";
import { useCommonStore } from "@/stores/common";
import { useUserStore } from "./users";
import { FSService } from "@/../bindings/clustta/services";
import { useProjectStore } from "./projects";
import { useStageStore } from "./stages";
import { repository } from "@/lib/repositorypb";
import pako from "pako";
import emitter from "@/lib/mitt";

export const useAssetStore = defineStore("asset", {
  state: () => ({
    tasks: [],
    taskTypes: [],
    filteredTasks: [],
    tasks_index: {},
    entity_tasks_index: {},
    isTaskStatus: false,
    isTaskListFiltered: false,
    selectedTask: null,
    selectedTaskType: null,
    tasksLoaded: false,
    projectTags: [],
    tasksSearchQuery: "",
    showDoneTasks: true,
    loadingAssetStates: false,
    taskListTags: [],
    projectExtensions: [],
    projectExtensionsFlat: [],
    modifiedTasksPath: [],
    outdatedTasksPath: [],
    rebuildableTasksPath: [],
    untrackedTasksPath: [],
    modifiedAssetsState: [],
    outdatedAssetsState: [],
    rebuildableAssetsState: [],
  }),
  getters: {
    getTaskTypes: (state) => {
      return state.taskTypes;
    },
    getTaskTypesNames: (state) => {
      let taskTypes = state.taskTypes;
      let taskTypesNames = [];
      for (let i = 0; i < taskTypes.length; i++) {
        let taskType = taskTypes[i];
        taskTypesNames.push(taskType.name);
      }
      return taskTypesNames;
    },
    getTasks: (state) => {
      const commonStore = useCommonStore();

      let tasks = [];
      for (let i = 0; i < state.tasks.length; i++) {
        let task = state.tasks[i];
        if (
          task.trashed === false &&
          (state.showDoneTasks || task.status.name !== "done") &&
          (!state.showTraySearch ||
            state.taskListTags.every((tag) => task.tags.includes(tag))) &&
          (commonStore.viewSearchQuery === "" ||
            !state.showTraySearch ||
            task.name
              .toLowerCase()
              .includes(commonStore.viewSearchQuery.toLowerCase()))
        ) {
          tasks.push(task);
        }
      }

      // Sort entitys with closed entitys at the bottom
      // tasks.sort((a, b) => {
      //   return a.name.localeCompare(b.name);
      // });
      return tasks;
    },

    getAssignedTasks: (state) => {
      const commonStore = useCommonStore();
      const userStore = useUserStore();
      let user = userStore.user;
      let tasks = [];
      for (let i = 0; i < state.tasks.length; i++) {
        let task = state.tasks[i];
        if (task.assignee_id === user.id && task.trashed === false) {
          tasks.push(task);
        }
      }

      // Sort entitys with closed entitys at the bottom
      tasks.sort((a, b) => {
        return a.name.localeCompare(b.name);
      });
      return state.tasks;
    },

    getFilteredTasks: (state) => {
      const commonStore = useCommonStore();

      let tasks = [];
      for (let i = 0; i < state.tasks.length; i++) {
        let task = state.tasks[i];
        if (
          task.trashed === false &&
          (state.showDoneTasks || task.status.name !== "done") &&
          (!state.showTraySearch ||
            state.taskListTags.every((tag) => task.tags.includes(tag))) &&
          (commonStore.viewSearchQuery === "" ||
            !state.showTraySearch ||
            task.name
              .toLowerCase()
              .includes(commonStore.viewSearchQuery.toLowerCase()))
        ) {
          tasks.push(task);
        }
      }

      // Sort entitys with closed entitys at the bottom
      // tasks.sort((a, b) => {
      //   return a.name.localeCompare(b.name);
      // });

      const viewSearchQuery = commonStore.viewSearchQuery;
      const workspaceSearchQuery = commonStore.workspaceSearchQuery;

      let filteredTask;

      const assigneeFilters =
        commonStore.hasAssignees || commonStore.noAssignees;
      const filtersActive = commonStore.taskFilters.length || assigneeFilters;

      if (filtersActive) {
        console.log("Filters active");
        const selectedStatus = commonStore.taskFilters
          .filter((filter) => filter.type === "status")
          .map((filter) => filter.name.toLowerCase());
        const selectedTags = commonStore.taskFilters
          .filter((filter) => filter.type === "tags")
          .map((filter) => filter.name.toLowerCase());
        const selectedAssignees = commonStore.taskFilters
          .filter((filter) => filter.type === "assignation")
          .map((filter) => filter.id);
        const selectedState = commonStore.taskFilters
          .filter((filter) => filter.type === "state")
          .map((filter) => filter.name.toLowerCase());
        const selectedExtensions = commonStore.taskFilters
          .filter((filter) => filter.type === "extension")
          .map((filter) => filter.extension.toLowerCase());
        const selectedTaskTypes = commonStore.taskFilters
          .filter((filter) => filter.type === "task-type")
          .map((filter) => filter.name.toLowerCase());

        filteredTask = tasks
          .filter((task) => {
            // matched status
            const statusMatch =
              selectedStatus.length === 0 ||
              selectedStatus.includes(task.status_short_name.toLowerCase());

            // matched tags
            let tagMatch;
            const matchAll = true;
            if (matchAll) {
              tagMatch =
                selectedTags.length === 0 ||
                selectedTags.some((tag) =>
                  task.tags.some(
                    (taskTag) =>
                      taskTag.replace(/\s+/g, "").toLowerCase() ===
                      tag.replace(/\s+/g, "").toLowerCase()
                  )
                );
            } else {
              tagMatch =
                selectedTags.length === 0 ||
                selectedTags.every((tag) =>
                  task.tags.some(
                    (taskTag) =>
                      taskTag.replace(/\s+/g, "").toLowerCase() ===
                      tag.replace(/\s+/g, "").toLowerCase()
                  )
                );
            }

            // matched assignees
            let assigneeMatch;
            if (commonStore.hasAssignees) {
              assigneeMatch = task.assignee_id;
            } else if (commonStore.noAssignees) {
              assigneeMatch = !task.assignee_id;
            } else {
              assigneeMatch =
                selectedAssignees.length === 0 ||
                selectedAssignees.includes(task.assignee_id);
            }

            // matched states
            const stateMatch =
              selectedState.length === 0 ||
              selectedState.includes(task.file_status);

            // matched extensions
            const extensionMatch =
              selectedExtensions.length === 0 ||
              selectedExtensions.includes(task.extension.toLowerCase());

            // matched task types
            const taskType = state.taskTypes.find(
              (item) => item.id === task.task_type_id
            );
            const taskTypeMatch =
              selectedTaskTypes.length === 0 ||
              selectedTaskTypes.includes(taskType.name.toLowerCase());

            return (
              statusMatch &&
              assigneeMatch &&
              tagMatch &&
              stateMatch &&
              extensionMatch &&
              taskTypeMatch
            );
          })
          .filter(
            (item) =>
              item.file_path
                .toLowerCase()
                .replace(/\\/g, "/")
                .includes(viewSearchQuery) &&
              item.file_path
                .toLowerCase()
                .replace(/\\/g, "/")
                .includes(workspaceSearchQuery)
          );
        // const sortedTasks = utils.sortAlphabetically(filteredTask);
        // const sortedTasks = utils.sortPathAlphabetically(filteredTask, "task");
        return commonStore.showDependencies
          ? filteredTask
          : filteredTask.filter((item) => !item.is_dependency);
      } else {
        filteredTask = tasks.filter(
          (item) =>
            item.file_path
              .toLowerCase()
              .replace(/\\/g, "/")
              .includes(viewSearchQuery) &&
            item.file_path
              .toLowerCase()
              .replace(/\\/g, "/")
              .includes(workspaceSearchQuery)
        );
        // const sortedTasks = utils.sortAlphabetically(filteredTask);
        // const sortedTasks = utils.sortPathAlphabetically(filteredTask, "task");
        return commonStore.showDependencies
          ? filteredTask
          : filteredTask.filter((item) => !item.is_dependency);
      }
    },

    getDisplayedTasks: (state) => {
      const stageStore = useStageStore();
      const commonStore = useCommonStore();
      const collectionStore = useCollectionStore();
      let filteredTasks = state.getFilteredTasks;
      let showEntities = commonStore.showEntities;

      let expandedEntities = stageStore.expandedEntities;
      let displayedTasks = [];

      for (let i = 0; i < filteredTasks.length; i++) {
        let task = filteredTasks[i];
        if (showEntities) {
          if (
            task.entity_id === "" ||
            task.entity_id in expandedEntities // Changed from expandedEntities.includes(task.entity_id)
          ) {
            displayedTasks.push(task);
          }
        } else {
          displayedTasks.push(task);
        }
      }
      return displayedTasks;
    },

    // getEntityTasks: (state) => {
    //   const collectionStore = useCollectionStore();
    //   let tasks = {};
    //   let taskIds = [];
    //   if (collectionStore.selectedEntity) {
    //     for (let i = 0; i < state.tasks.length; i++) {
    //       let task = state.tasks[i];
    //       if (
    //         task.trashed === false &&
    //         task.entity_id === collectionStore.selectedEntity.id
    //       ) {
    //         tasks[task.id] = task;
    //         taskIds.push(task.id);
    //       }
    //     }
    //   }
    //   let task_list = Object.values(tasks);
    //   const statusOrder = { wip: 0, todo: 1, done: 2 };
    //   task_list.sort((a, b) => {
    //     // First, sort by status
    //     const statusComparison =
    //       statusOrder[a.status.short_name] - statusOrder[b.status.short_name];
    //     // statusOrder[b.status_short_name];
    //     if (statusComparison !== 0) {
    //       return statusComparison;
    //     }
    //     // If status is the same, sort alphabetically
    //     return a.name.localeCompare(b.name);
    //   });

    //   state.displayedTasks = task_list;
    //   return task_list;
    // },

    doneTasksExist: (state) => {
      const collectionStore = useCollectionStore();
      for (let i = 0; i < state.tasks.length; i++) {
        let task = state.tasks[i];
        if (task.trashed) {
          continue;
        }
        if (state.showDoneTasks || task.status.name === "done") {
          return true;
        }
      }
      return false;
    },

    getTaskDependencies: (state) => {
      return state.tasks;
    },

    getDisplaytasksFileStatus: (state) => {
      //return the most important file status of the tasks
      let statusPiority = ["modified", "missing", "outdated", "normal"];
      let status = "normal";
      for (let i = 0; i < state.getDisplayedTasks.length; i++) {
        let task = state.getDisplayedTasks[i];
        if (
          task.file_status &&
          statusPiority.indexOf(task.file_status) <
            statusPiority.indexOf(status)
        ) {
          status = task.file_status;
        }
      }
      return status;
    },

    // Getters for display paths (with extensions)
    getModifiedDisplayPaths: (state) => {
      return state.modifiedAssetsState;
    },
    getOutdatedDisplayPaths: (state) => {
      return state.outdatedAssetsState;
    },
    getRebuildableDisplayPaths: (state) => {
      return state.rebuildableAssetsState;
    },
  },
  actions: {

    async filterTasks (tasks){

      const commonStore = useCommonStore();
      const viewSearchQuery = commonStore.viewSearchQuery || "";
      const workspaceSearchQuery = commonStore.workspaceSearchQuery || "";

      let filteredTask;

      const assigneeFilters =
        commonStore.hasAssignees || commonStore.noAssignees;
      const filtersActive = commonStore.taskFilters.length || assigneeFilters;

      if (filtersActive) {
        const selectedStatus = commonStore.taskFilters
          .filter((filter) => filter.type === "status")
          .map((filter) => filter.name.toLowerCase());
        const selectedTags = commonStore.taskFilters
          .filter((filter) => filter.type === "tags")
          .map((filter) => filter.name.toLowerCase());
        const selectedAssignees = commonStore.taskFilters
          .filter((filter) => filter.type === "assignation")
          .map((filter) => filter.id);
        const selectedState = commonStore.taskFilters
          .filter((filter) => filter.type === "state")
          .map((filter) => filter.name.toLowerCase());
        const selectedExtensions = commonStore.taskFilters
          .filter((filter) => filter.type === "extension")
          .map((filter) => filter.extension.toLowerCase());
        const selectedTaskTypes = commonStore.taskFilters
          .filter((filter) => filter.type === "asset-type")
          .map((filter) => filter.name.toLowerCase());

        filteredTask = tasks
          .filter((task) => {
            // matched status
            const statusMatch =
              selectedStatus.length === 0 ||
              selectedStatus.includes(task.status_short_name.toLowerCase());

            // matched tags
            let tagMatch;
            const matchAll = true;
            if (matchAll) {
              tagMatch =
                selectedTags.length === 0 ||
                selectedTags.some((tag) =>
                  task.tags.some(
                    (taskTag) =>
                      taskTag.replace(/\s+/g, "").toLowerCase() ===
                      tag.replace(/\s+/g, "").toLowerCase()
                  )
                );
            } else {
              tagMatch =
                selectedTags.length === 0 ||
                selectedTags.every((tag) =>
                  task.tags.some(
                    (taskTag) =>
                      taskTag.replace(/\s+/g, "").toLowerCase() ===
                      tag.replace(/\s+/g, "").toLowerCase()
                  )
                );
            }

            // matched assignees
            let assigneeMatch;
            if (commonStore.hasAssignees) {
              assigneeMatch = task.assignee_id;
            } else if (commonStore.noAssignees) {
              assigneeMatch = !task.assignee_id;
            } else {
              assigneeMatch =
                selectedAssignees.length === 0 ||
                selectedAssignees.includes(task.assignee_id);
            }

            // matched states
            const stateMatch =
              selectedState.length === 0 ||
              selectedState.includes(task.file_status);

            // matched extensions
            const extensionMatch =
              selectedExtensions.length === 0 ||
              selectedExtensions.includes(task.extension.toLowerCase());

            // matched task types
            const taskType = this.taskTypes.find(
              (item) => item.id === task.task_type_id
            );
            const taskTypeMatch =
              selectedTaskTypes.length === 0 ||
              selectedTaskTypes.includes(taskType.name.toLowerCase());

            return (
              statusMatch &&
              assigneeMatch &&
              tagMatch &&
              stateMatch &&
              extensionMatch &&
              taskTypeMatch
            );
          })
          .filter(
            (item) => {
              const nameMatch = !viewSearchQuery ||  item.name?.toLowerCase().includes(viewSearchQuery.toLowerCase());
              const filePathMatch = !viewSearchQuery || item.file_path?.toLowerCase().replace(/\\/g, "/").includes(viewSearchQuery.toLowerCase());
              const workspaceNameMatch = !workspaceSearchQuery || item.name?.toLowerCase().includes(workspaceSearchQuery.toLowerCase());
              const workspaceFilePathMatch = !workspaceSearchQuery || item.file_path?.toLowerCase().replace(/\\/g, "/").includes(workspaceSearchQuery.toLowerCase());
              
              return (nameMatch || filePathMatch) && (workspaceNameMatch || workspaceFilePathMatch);
            }
          );
        return commonStore.showResources
          ? filteredTask
          : filteredTask.filter((item) => !item.is_resource);
      } else {
        filteredTask = tasks.filter(
          (item) => {
            const nameMatch = !viewSearchQuery || item.name?.toLowerCase().includes(viewSearchQuery.toLowerCase());
            const filePathMatch = !viewSearchQuery || item.file_path?.toLowerCase().replace(/\\/g, "/").includes(viewSearchQuery.toLowerCase());
            const workspaceNameMatch = !workspaceSearchQuery || item.name?.toLowerCase().includes(workspaceSearchQuery.toLowerCase());
            const workspaceFilePathMatch = !workspaceSearchQuery || item.file_path?.toLowerCase().replace(/\\/g, "/").includes(workspaceSearchQuery.toLowerCase());
            
            return (nameMatch || filePathMatch) && (workspaceNameMatch || workspaceFilePathMatch);
          }
        );
        return commonStore.showResources
          ? filteredTask
          : filteredTask.filter((item) => !item.is_resource);
      }
    },
    async addModifiedTaskPath(taskPath) {
      if (!(taskPath in this.modifiedTasksPath)) {
        this.modifiedTasksPath.push(taskPath);
      }
    },
    async removeModifiedTaskPath(taskPath) {
        this.modifiedTasksPath.filter((item)=>item !== taskPath)
    },
    async refreshDisplayedFilesStatus() {
      if (!this.getDisplayedTasks.length) {
        return;
      }
      for (let i = 0; i < this.getDisplayedTasks.length; i++) {
        let task_id = this.getDisplayedTasks[i].id;
        let taskIndex = this.tasks_index[task_id];
        let status = await this.getTaskFileStatus(this.tasks[taskIndex]);
        this.tasks[taskIndex].file_status = status;
      }
    },

    async refreshEntityFilesStatus(entity_id) {
      let taskIds = [];
      for (let task of this.tasks) {
        if (entity_id == task.entity_id || entity_id === "") {
          taskIds.push(task.id);
        }
      }

      for (let i = 0; i < taskIds.length; i++) {
        let task_id = taskIds[i];
        let taskIndex = this.tasks_index[task_id];
        let status = await this.getTaskFileStatus(this.tasks[taskIndex]);
        this.tasks[taskIndex].file_status = status;
      }
    },

    async refreshAllFilesStatus() {
      let taskIds = [];
      for (let task of this.tasks) {
        taskIds.push(task.id);
      }

      for (let i = 0; i < taskIds.length; i++) {
        let task_id = taskIds[i];
        let taskIndex = this.tasks_index[task_id];
        let status = await this.getTaskFileStatus(this.tasks[taskIndex]);
        this.tasks[taskIndex].file_status = status;
      }
    },

    async markTaskAsDeleted(taskId) {
      let taskIndex = this.tasks_index[taskId];
      this.tasks[taskIndex].trashed = true;
    },

    async unmarkTaskAsDeleted(taskId) {
      let taskIndex = this.tasks_index[taskId];
      this.tasks[taskIndex].trashed = false;
    },

    // here
    async markMultipleTasksAsDeleted(taskIds) {
      for (const taskId of taskIds) {
        let taskIndex = this.tasks_index[taskId];
        if (taskIndex) {
          this.tasks[taskIndex].trashed = true;
        }
      }
    },

    async unmarkMultipleTasksAsDeleted(taskIds) {
      for (const taskId of taskIds) {
        let taskIndex = this.tasks_index[taskId];
        if (taskIndex) {
          this.tasks[taskIndex].trashed = false;
        }
      }
    },

    async reloadTaskTypes() {
      const projectStore = useProjectStore();
      let taskTypes = await TaskService.GetTaskTypes(
        projectStore.activeProject.uri
      );
      this.taskTypes = taskTypes.map(type => ({
        ...type,
        type: 'asset-type',
      }));
    },

    async reloadTasks(refresh = false) {
      const projectStore = useProjectStore();
      this.projectExtensions = [];
      if (!refresh) {
        this.tasksLoaded = false;
      }
      const iconStore = useIconStore();
      this.projectTags = [];
      let tags = await TagService.GetTags(projectStore.activeProject.uri);
      for (let i = 0; i < tags.length; i++) {
        this.projectTags.push(tags[i].name);
      }

      console.time("test_loading");
      await TaskService.TestData();
      console.timeEnd("test_loading");

      console.time("tasks_loading");
      let btasks = await TaskService.GetTasksPB(projectStore.activeProject.uri);
      console.timeEnd("tasks_loading");

      console.time("tasks_decoding");
      let taskBinary = pako.inflate(utils.base64ToUint8Array(btasks));

      let fullTaskList = repository.FullTaskList.decode(taskBinary);
      let tasks = fullTaskList.full_tasks;
      console.timeEnd("tasks_decoding");

      // let tasks = await TaskService.GetTasks(projectStore.activeProject.uri);
      await this.processTasksIconsAndPreviews(tasks);
      this.tasks = tasks;
      await this.rebuildTasksIndex();
      this.tasksLoaded = true;
    },

    async rebuildTasksIndex() {
      let taskIndex = {};
      let entityTasksIndex = {};
      let modifiedTasksPath = [];
      for (let i = 0; i < this.tasks.length; i++) {
        if (this.tasks[i].file_status === "modified") {
          modifiedTasksPath.push(this.tasks[i].task_path);
        }
        let entityId = this.tasks[i].entity_id;
        let taskId = this.tasks[i].id;
        taskIndex[taskId] = i;
        if (!entityTasksIndex[entityId]) {
          entityTasksIndex[entityId] = [taskId];
        } else {
          entityTasksIndex[entityId].push(taskId);
        }
      }
      this.tasks_index = taskIndex;
      this.entity_tasks_index = entityTasksIndex;
      this.modifiedTasksPath = modifiedTasksPath;
    },

    findTask(id) {
      let taskIndex = this.tasks_index[id];
      return this.tasks[taskIndex];
    },

    getEntityTasks(entityId, recursive = false, entityMap = {}) {
      let entityTaskIds = this.entity_tasks_index[entityId] || [];
      let tasks = entityTaskIds.map((taskId) => this.findTask(taskId));

      if (recursive && entityMap[entityId]) {
        for (let childId of entityMap[entityId]) {
          tasks = tasks.concat(this.getEntityTasks(childId, true, entityMap));
        }
      }

      return tasks;
    },
    selectTask(task) {
      this.selectedTask = task;
    },

    getTaskTypeIcon(taskTypeId) {
      let taskTypeIcon = "";
      for (let i = 0; i < this.taskTypes.length; i++) {
        let type = this.taskTypes[i];
        if (type.id === taskTypeId) {
          taskTypeIcon = type.icon;
          break;
        }
      }
      return taskTypeIcon;
    },

    async setStatus(status, taskId) {
      const projectStore = useProjectStore();
      await TaskService.ChangeStatus(
        projectStore.activeProject.uri,
        taskId,
        status.id
      )
        .then((data) => {
        })
        .catch((error) => {
          console.error("Error:", error);
        });
      emitter.emit("refresh-browser");
    },
    async setMultipleStatus(status, taskIds) {
      const projectStore = useProjectStore();
      for (const taskId of taskIds) {
        await TaskService.ChangeStatus(
          projectStore.activeProject.uri,
          taskId,
          status.id
        )
          .then((data) => {})
          .catch((error) => {
            console.error("Error:", error);
          });
      }
      emitter.emit("refresh-browser");
    },
    toggleShowDoneTasks() {
      this.showDoneTasks = !this.showDoneTasks;
    },
    addDependency(task_id, dependency_id, itemType) {
      let taskIndex = this.tasks_index[task_id];
      if (itemType === "task") {
        this.tasks[taskIndex].dependencies.push(dependency_id);
      } else {
        this.tasks[taskIndex].entity_dependencies.push(dependency_id);
      }
    },
    removeDependency(task_id, dependency_id, itemType) {
      let taskIndex = this.tasks_index[task_id];
      if (itemType === "task") {
        let dependencies = this.tasks[taskIndex].dependencies;
        this.tasks[taskIndex].dependencies = dependencies.filter(
          (dep) => dep !== dependency_id
        );
      } else {
        let entity_dependencies = this.tasks[taskIndex].entity_dependencies;
        this.tasks[taskIndex].entity_dependencies = entity_dependencies.filter(
          (dep) => dep !== dependency_id
        );
      }
    },
    changeTaskEntity(task_id, entityId) {
      let taskIndex = this.tasks_index[task_id];
      this.tasks[taskIndex].entity_id = entityId;
      this.rebuildTasksIndex();
    },
    addTasksCheckpoint(checkpoints) {
      for (let checkpoint of checkpoints) {
        let taskId = checkpoint.task_id;
        let taskIndex = this.tasks_index[taskId];
        console.log(this.tasks[taskIndex]);
        this.tasks[taskIndex].checkpoints.unshift(checkpoint);
      }
    },
    async getTaskFileStatus(task) {
      if (!task) {
        return "normal";
      }
      let filePath = task.file_path;
      let checkpoints = task.checkpoints;
      if (task.is_link && (await isValidPointer(task.pointer))) {
        return "normal";
      }
      if (task.IsLink && (await isValidPointer(task.pointer))) {
        return "missing";
      }

      if (!checkpoints || checkpoints.length == 0) {
        return "missing";
      }

      let isMissing = !(await FSService.Exists(filePath));
      if (isMissing && (!checkpoints || checkpoints.length == 0)) {
        return "missing";
      } else if (isMissing && checkpoints.length != 0) {
        return "rebuildable";
      }
      // let fileStat = await FSService.FileStat(filePath);
      let hash = await FSService.FileHash(filePath);
      // let fileSize = fileStat.size;
      // let fileTimeModified = fileStat.modTime;
      for (let i = 0; i < checkpoints.length; i++) {
        let checkpoint = checkpoints[i];
        let checkpointFileSize = checkpoint.file_size;
        // let checkpointTimeModified = checkpoint.time_modified;
        // if (fileTimeModified > checkpointTimeModified) {
        //   return "modified";
        // }
        // if (fileSize != checkpointFileSize) {
        //   return "modified";
        // }
        if (
          hash == checkpoint.xxhash_checksum
          // hash == checkpoint.xxhash_checksum &&
          // fileSize == checkpointFileSize &&
          // fileTimeModified == checkpointTimeModified
        ) {
          if (i == 0) {
            return "normal";
          } else {
            return "outdated";
          }
        }
      }

      return "modified";
    },
    getTaskEntity(taskId, recursive = false) {
      const collectionStore = useCollectionStore();
      let taskIndex = this.tasks_index[taskId];
      let task = this.tasks[taskIndex];
      let entityId = task.entity_id;
      let parentEntities = [];

      if (entityId) {
        let parentEntity = collectionStore.findEntity(entityId);
        if (parentEntity) {
          parentEntities.push(parentEntity);
        }
        if (recursive) {
          let parentId = parentEntity.parent_id;
          let depth = 0;
          while (parentId && depth < 20) {
            let parentEntity = collectionStore.findEntity(parentId);
            if (parentEntity) {
              parentEntities.push(parentEntity);
              parentId = parentEntity.parent_id;
            } else {
              break;
            }
            depth++;
          }
        }
      }
      return parentEntities;
    },

    async processTasksIconsAndPreviews(tasks) {
      const iconStore = useIconStore();
      
      if (!tasks || !Array.isArray(tasks)) {
        return tasks;
      }

      for (let i = 0; i < tasks.length; i++) {
        let task = tasks[i];
        let extension = "";
        
        // Set file_path from pointer if it exists
        if (task.pointer) {
          task.file_path = task.pointer;
        }
        
        // Determine extension
        if (task.is_link && !isValidWeblink(task.pointer)) {
          extension = utils.getFileExtension(task.pointer).substring(1);
        } else if (!task.is_link) {
          extension = task.extension.toLowerCase().substring(1);
          
          // Add to project extensions if not already present
          if (!this.projectExtensionsFlat.includes(extension)) {
            this.projectExtensionsFlat.push(extension);
            let extensionData = {
              name: extension,
              type: "extension",
              extension: "." + extension,
              icon: (await iconStore.getIcon(extension)) || "",
            };
            this.projectExtensions.push(extensionData);
          }
        }
        
        // Set icon
        let iconPath = "";
        if (task.is_link && isValidWeblink(task.pointer)) {
          iconPath = await iconStore.getWebIcon(task.pointer);
        } else {
          iconPath = (await iconStore.getIcon(extension)) || "";
        }
        task.icon = iconPath;
        
        // Set preview
        let preview = null;
        if (task.preview) {
          preview = "data:image/png;base64," + task.preview;
        }
        task.preview = preview;
      }
      
      return tasks;
    },

    async processUntrackedTasksIcons(untrackedTasks) {
      const iconStore = useIconStore();
      
      if (!untrackedTasks || !Array.isArray(untrackedTasks)) {
        return untrackedTasks;
      }

      for (let i = 0; i < untrackedTasks.length; i++) {
        let extension = "";
        extension = untrackedTasks[i].extension.toLowerCase().substring(1);
        
        // Add to project extensions if not already present
        if (!this.projectExtensionsFlat.includes(extension)) {
          this.projectExtensionsFlat.push(extension);
          let extensionData = {
            name: extension,
            type: "extension",
            extension: "." + extension,
            icon: (await iconStore.getIcon(extension)) || "",
          };
          this.projectExtensions.push(extensionData);
        }

        let iconPath = (await iconStore.getIcon(extension)) || "";
        untrackedTasks[i].icon = iconPath;
      }
      
      return untrackedTasks;
    },

    async reloadAssetStates(){
      this.loadingAssetStates = true;
      const projectStore = useProjectStore();
      let project = projectStore.activeProject
      await TaskService.GetAssetsStates(project.uri, project.working_directory, project.ignore_list).then((assetsStates) => {
        this.modifiedTasksPath = assetsStates.modified.map(item => item.task_path)
        this.outdatedTasksPath = assetsStates.outdated.map(item => item.task_path)
        this.rebuildableTasksPath = assetsStates.rebuildable.map(item => item.task_path)

        // Store the full asset state data for UI display
        this.modifiedAssetsState = assetsStates.modified
        this.outdatedAssetsState = assetsStates.outdated
        this.rebuildableAssetsState = assetsStates.rebuildable
      })
      await TaskService.GetUntrackedFiles(project.uri, project.working_directory, project.ignore_list).then((untrackedFiles) => {
        this.untrackedTasksPath = untrackedFiles
      })
      // stage.operationActive = false
      this.loadingAssetStates = false;
  }

  },
});



