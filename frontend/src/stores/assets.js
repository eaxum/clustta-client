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
    getAssetTypes: (state) => {
      return state.taskTypes;
    },
    getAssetTypesNames: (state) => {
      let assetTypes = state.taskTypes;
      let assetTypesNames = [];
      for (let i = 0; i < assetTypes.length; i++) {
        let assetType = assetTypes[i];
        assetTypesNames.push(assetType.name);
      }
      return assetTypesNames;
    },
    getAssets: (state) => {
      const commonStore = useCommonStore();

      let assets = [];
      for (let i = 0; i < state.tasks.length; i++) {
        let asset = state.tasks[i];
        if (
          asset.trashed === false &&
          (state.showDoneTasks || asset.status.name !== "done") &&
          (!state.showTraySearch ||
            state.taskListTags.every((tag) => asset.tags.includes(tag))) &&
          (commonStore.viewSearchQuery === "" ||
            !state.showTraySearch ||
            asset.name
              .toLowerCase()
              .includes(commonStore.viewSearchQuery.toLowerCase()))
        ) {
          assets.push(asset);
        }
      }

      // Sort entitys with closed entitys at the bottom
      // assets.sort((a, b) => {
      //   return a.name.localeCompare(b.name);
      // });
      return assets;
    },

    getAssignedAssets: (state) => {
      const commonStore = useCommonStore();
      const userStore = useUserStore();
      let user = userStore.user;
      let assets = [];
      for (let i = 0; i < state.tasks.length; i++) {
        let asset = state.tasks[i];
        if (asset.assignee_id === user.id && asset.trashed === false) {
          assets.push(asset);
        }
      }

      // Sort entitys with closed entitys at the bottom
      assets.sort((a, b) => {
        return a.name.localeCompare(b.name);
      });
      return state.tasks;
    },

    getFilteredAssets: (state) => {
      const commonStore = useCommonStore();

      let assets = [];
      for (let i = 0; i < state.tasks.length; i++) {
        let asset = state.tasks[i];
        if (
          asset.trashed === false &&
          (state.showDoneTasks || asset.status.name !== "done") &&
          (!state.showTraySearch ||
            state.taskListTags.every((tag) => asset.tags.includes(tag))) &&
          (commonStore.viewSearchQuery === "" ||
            !state.showTraySearch ||
            asset.name
              .toLowerCase()
              .includes(commonStore.viewSearchQuery.toLowerCase()))
        ) {
          assets.push(asset);
        }
      }

      // Sort entitys with closed entitys at the bottom
      // assets.sort((a, b) => {
      //   return a.name.localeCompare(b.name);
      // });

      const viewSearchQuery = commonStore.viewSearchQuery;
      const workspaceSearchQuery = commonStore.workspaceSearchQuery;

      let filteredAsset;

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
        const selectedAssetTypes = commonStore.taskFilters
          .filter((filter) => filter.type === "task-type")
          .map((filter) => filter.name.toLowerCase());

        filteredAsset = assets
          .filter((asset) => {
            // matched status
            const statusMatch =
              selectedStatus.length === 0 ||
              selectedStatus.includes(asset.status_short_name.toLowerCase());

            // matched tags
            let tagMatch;
            const matchAll = true;
            if (matchAll) {
              tagMatch =
                selectedTags.length === 0 ||
                selectedTags.some((tag) =>
                  asset.tags.some(
                    (assetTag) =>
                      assetTag.replace(/\s+/g, "").toLowerCase() ===
                      tag.replace(/\s+/g, "").toLowerCase()
                  )
                );
            } else {
              tagMatch =
                selectedTags.length === 0 ||
                selectedTags.every((tag) =>
                  asset.tags.some(
                    (assetTag) =>
                      assetTag.replace(/\s+/g, "").toLowerCase() ===
                      tag.replace(/\s+/g, "").toLowerCase()
                  )
                );
            }

            // matched assignees
            let assigneeMatch;
            if (commonStore.hasAssignees) {
              assigneeMatch = asset.assignee_id;
            } else if (commonStore.noAssignees) {
              assigneeMatch = !asset.assignee_id;
            } else {
              assigneeMatch =
                selectedAssignees.length === 0 ||
                selectedAssignees.includes(asset.assignee_id);
            }

            // matched states
            const stateMatch =
              selectedState.length === 0 ||
              selectedState.includes(asset.file_status);

            // matched extensions
            const extensionMatch =
              selectedExtensions.length === 0 ||
              selectedExtensions.includes(asset.extension.toLowerCase());

            // matched asset types
            const assetType = state.taskTypes.find(
              (item) => item.id === asset.task_type_id
            );
            const assetTypeMatch =
              selectedAssetTypes.length === 0 ||
              selectedAssetTypes.includes(assetType.name.toLowerCase());

            return (
              statusMatch &&
              assigneeMatch &&
              tagMatch &&
              stateMatch &&
              extensionMatch &&
              assetTypeMatch
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
        // const sortedAssets = utils.sortAlphabetically(filteredAsset);
        // const sortedAssets = utils.sortPathAlphabetically(filteredAsset, "asset");
        return commonStore.showDependencies
          ? filteredAsset
          : filteredAsset.filter((item) => !item.is_dependency);
      } else {
        filteredAsset = assets.filter(
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
        // const sortedAssets = utils.sortAlphabetically(filteredAsset);
        // const sortedAssets = utils.sortPathAlphabetically(filteredAsset, "asset");
        return commonStore.showDependencies
          ? filteredAsset
          : filteredAsset.filter((item) => !item.is_dependency);
      }
    },

    getDisplayedAssets: (state) => {
      const stageStore = useStageStore();
      const commonStore = useCommonStore();
      const collectionStore = useCollectionStore();
      let filteredAssets = state.getFilteredAssets;
      let showEntities = commonStore.showEntities;

      let expandedEntities = stageStore.expandedEntities;
      let displayedAssets = [];

      for (let i = 0; i < filteredAssets.length; i++) {
        let asset = filteredAssets[i];
        if (showEntities) {
          if (
            asset.entity_id === "" ||
            asset.entity_id in expandedEntities // Changed from expandedEntities.includes(asset.entity_id)
          ) {
            displayedAssets.push(asset);
          }
        } else {
          displayedAssets.push(asset);
        }
      }
      return displayedAssets;
    },

    // getEntityTasks: (state) => {
    //   const collectionStore = useCollectionStore();
    //   let tasks = {};
    //   let taskIds = [];
    //   if (collectionStore.selectedCollection) {
    //     for (let i = 0; i < state.tasks.length; i++) {
    //       let task = state.tasks[i];
    //       if (
    //         task.trashed === false &&
    //         task.entity_id === collectionStore.selectedCollection.id
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

    doneAssetsExist: (state) => {
      const collectionStore = useCollectionStore();
      for (let i = 0; i < state.tasks.length; i++) {
        let asset = state.tasks[i];
        if (asset.trashed) {
          continue;
        }
        if (state.showDoneTasks || asset.status.name === "done") {
          return true;
        }
      }
      return false;
    },

    getAssetDependencies: (state) => {
      return state.tasks;
    },

    getDisplayAssetsFileStatus: (state) => {
      //return the most important file status of the assets
      let statusPiority = ["modified", "missing", "outdated", "normal"];
      let status = "normal";
      for (let i = 0; i < state.getDisplayedAssets.length; i++) {
        let asset = state.getDisplayedAssets[i];
        if (
          asset.file_status &&
          statusPiority.indexOf(asset.file_status) <
            statusPiority.indexOf(status)
        ) {
          status = asset.file_status;
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

    async filterAssets (assets){

      const commonStore = useCommonStore();
      const viewSearchQuery = commonStore.viewSearchQuery || "";
      const workspaceSearchQuery = commonStore.workspaceSearchQuery || "";

      let filteredAsset;

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
        const selectedAssetTypes = commonStore.taskFilters
          .filter((filter) => filter.type === "asset-type")
          .map((filter) => filter.name.toLowerCase());

        filteredAsset = assets
          .filter((asset) => {
            // matched status
            const statusMatch =
              selectedStatus.length === 0 ||
              selectedStatus.includes(asset.status_short_name.toLowerCase());

            // matched tags
            let tagMatch;
            const matchAll = true;
            if (matchAll) {
              tagMatch =
                selectedTags.length === 0 ||
                selectedTags.some((tag) =>
                  asset.tags.some(
                    (assetTag) =>
                      assetTag.replace(/\s+/g, "").toLowerCase() ===
                      tag.replace(/\s+/g, "").toLowerCase()
                  )
                );
            } else {
              tagMatch =
                selectedTags.length === 0 ||
                selectedTags.every((tag) =>
                  asset.tags.some(
                    (assetTag) =>
                      assetTag.replace(/\s+/g, "").toLowerCase() ===
                      tag.replace(/\s+/g, "").toLowerCase()
                  )
                );
            }

            // matched assignees
            let assigneeMatch;
            if (commonStore.hasAssignees) {
              assigneeMatch = asset.assignee_id;
            } else if (commonStore.noAssignees) {
              assigneeMatch = !asset.assignee_id;
            } else {
              assigneeMatch =
                selectedAssignees.length === 0 ||
                selectedAssignees.includes(asset.assignee_id);
            }

            // matched states
            const stateMatch =
              selectedState.length === 0 ||
              selectedState.includes(asset.file_status);

            // matched extensions
            const extensionMatch =
              selectedExtensions.length === 0 ||
              selectedExtensions.includes(asset.extension.toLowerCase());

            // matched asset types
            const assetType = this.taskTypes.find(
              (item) => item.id === asset.task_type_id
            );
            const assetTypeMatch =
              selectedAssetTypes.length === 0 ||
              selectedAssetTypes.includes(assetType.name.toLowerCase());

            return (
              statusMatch &&
              assigneeMatch &&
              tagMatch &&
              stateMatch &&
              extensionMatch &&
              assetTypeMatch
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
          ? filteredAsset
          : filteredAsset.filter((item) => !item.is_resource);
      } else {
        filteredAsset = assets.filter(
          (item) => {
            const nameMatch = !viewSearchQuery || item.name?.toLowerCase().includes(viewSearchQuery.toLowerCase());
            const filePathMatch = !viewSearchQuery || item.file_path?.toLowerCase().replace(/\\/g, "/").includes(viewSearchQuery.toLowerCase());
            const workspaceNameMatch = !workspaceSearchQuery || item.name?.toLowerCase().includes(workspaceSearchQuery.toLowerCase());
            const workspaceFilePathMatch = !workspaceSearchQuery || item.file_path?.toLowerCase().replace(/\\/g, "/").includes(workspaceSearchQuery.toLowerCase());
            
            return (nameMatch || filePathMatch) && (workspaceNameMatch || workspaceFilePathMatch);
          }
        );
        return commonStore.showResources
          ? filteredAsset
          : filteredAsset.filter((item) => !item.is_resource);
      }
    },
    async addModifiedAssetPath(assetPath) {
      if (!(assetPath in this.modifiedTasksPath)) {
        this.modifiedTasksPath.push(assetPath);
      }
    },
    async removeModifiedAssetPath(assetPath) {
        this.modifiedTasksPath.filter((item)=>item !== assetPath)
    },
    async refreshDisplayedFilesStatus() {
      if (!this.getDisplayedAssets.length) {
        return;
      }
      for (let i = 0; i < this.getDisplayedAssets.length; i++) {
        let asset_id = this.getDisplayedAssets[i].id;
        let assetIndex = this.tasks_index[asset_id];
        let status = await this.getAssetFileStatus(this.tasks[assetIndex]);
        this.tasks[assetIndex].file_status = status;
      }
    },

    async refreshEntityFilesStatus(entity_id) {
      let assetIds = [];
      for (let asset of this.tasks) {
        if (entity_id == asset.entity_id || entity_id === "") {
          assetIds.push(asset.id);
        }
      }

      for (let i = 0; i < assetIds.length; i++) {
        let asset_id = assetIds[i];
        let assetIndex = this.tasks_index[asset_id];
        let status = await this.getAssetFileStatus(this.tasks[assetIndex]);
        this.tasks[assetIndex].file_status = status;
      }
    },

    async refreshAllFilesStatus() {
      let assetIds = [];
      for (let asset of this.tasks) {
        assetIds.push(asset.id);
      }

      for (let i = 0; i < assetIds.length; i++) {
        let asset_id = assetIds[i];
        let assetIndex = this.tasks_index[asset_id];
        let status = await this.getAssetFileStatus(this.tasks[assetIndex]);
        this.tasks[assetIndex].file_status = status;
      }
    },

    async markAssetAsDeleted(assetId) {
      let assetIndex = this.tasks_index[assetId];
      this.tasks[assetIndex].trashed = true;
    },

    async unmarkAssetAsDeleted(assetId) {
      let assetIndex = this.tasks_index[assetId];
      this.tasks[assetIndex].trashed = false;
    },

    // here
    async markMultipleAssetsAsDeleted(assetIds) {
      for (const assetId of assetIds) {
        let assetIndex = this.tasks_index[assetId];
        if (assetIndex) {
          this.tasks[assetIndex].trashed = true;
        }
      }
    },

    async unmarkMultipleAssetsAsDeleted(assetIds) {
      for (const assetId of assetIds) {
        let assetIndex = this.tasks_index[assetId];
        if (assetIndex) {
          this.tasks[assetIndex].trashed = false;
        }
      }
    },

    async reloadAssetTypes() {
      const projectStore = useProjectStore();
      let assetTypes = await TaskService.GetTaskTypes(
        projectStore.activeProject.uri
      );
      this.taskTypes = assetTypes.map(type => ({
        ...type,
        type: 'asset-type',
      }));
    },

    async reloadAssets(refresh = false) {
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

      console.time("assets_loading");
      let btasks = await TaskService.GetTasksPB(projectStore.activeProject.uri);
      console.timeEnd("assets_loading");

      console.time("assets_decoding");
      let assetBinary = pako.inflate(utils.base64ToUint8Array(btasks));

      let fullAssetList = repository.FullTaskList.decode(assetBinary);
      let assets = fullAssetList.full_tasks;
      console.timeEnd("assets_decoding");

      // let assets = await TaskService.GetTasks(projectStore.activeProject.uri);
      await this.processAssetsIconsAndPreviews(assets);
      this.tasks = assets;
      await this.rebuildAssetsIndex();
      this.tasksLoaded = true;
    },

    async rebuildAssetsIndex() {
      let assetIndex = {};
      let entityAssetsIndex = {};
      let modifiedAssetsPath = [];
      for (let i = 0; i < this.tasks.length; i++) {
        if (this.tasks[i].file_status === "modified") {
          modifiedAssetsPath.push(this.tasks[i].task_path);
        }
        let entityId = this.tasks[i].entity_id;
        let assetId = this.tasks[i].id;
        assetIndex[assetId] = i;
        if (!entityAssetsIndex[entityId]) {
          entityAssetsIndex[entityId] = [assetId];
        } else {
          entityAssetsIndex[entityId].push(assetId);
        }
      }
      this.tasks_index = assetIndex;
      this.entity_tasks_index = entityAssetsIndex;
      this.modifiedTasksPath = modifiedAssetsPath;
    },

    findAsset(id) {
      let assetIndex = this.tasks_index[id];
      return this.tasks[assetIndex];
    },

    getEntityAssets(entityId, recursive = false, entityMap = {}) {
      let entityAssetIds = this.entity_tasks_index[entityId] || [];
      let assets = entityAssetIds.map((assetId) => this.findAsset(assetId));

      if (recursive && entityMap[entityId]) {
        for (let childId of entityMap[entityId]) {
          assets = assets.concat(this.getEntityAssets(childId, true, entityMap));
        }
      }

      return assets;
    },
    selectAsset(asset) {
      this.selectedTask = asset;
    },

    getAssetTypeIcon(assetTypeId) {
      let assetTypeIcon = "";
      for (let i = 0; i < this.taskTypes.length; i++) {
        let type = this.taskTypes[i];
        if (type.id === assetTypeId) {
          assetTypeIcon = type.icon;
          break;
        }
      }
      return assetTypeIcon;
    },

    async setStatus(status, assetId) {
      const projectStore = useProjectStore();
      await TaskService.ChangeStatus(
        projectStore.activeProject.uri,
        assetId,
        status.id
      )
        .then((data) => {
        })
        .catch((error) => {
          console.error("Error:", error);
        });
      emitter.emit("refresh-browser");
    },
    async setMultipleStatus(status, assetIds) {
      const projectStore = useProjectStore();
      for (const assetId of assetIds) {
        await TaskService.ChangeStatus(
          projectStore.activeProject.uri,
          assetId,
          status.id
        )
          .then((data) => {})
          .catch((error) => {
            console.error("Error:", error);
          });
      }
      emitter.emit("refresh-browser");
    },
    toggleShowDoneAssets() {
      this.showDoneTasks = !this.showDoneTasks;
    },
    addDependency(asset_id, dependency_id, itemType) {
      let assetIndex = this.tasks_index[asset_id];
      if (itemType === "asset") {
        this.tasks[assetIndex].dependencies.push(dependency_id);
      } else {
        this.tasks[assetIndex].entity_dependencies.push(dependency_id);
      }
    },
    removeDependency(asset_id, dependency_id, itemType) {
      let assetIndex = this.tasks_index[asset_id];
      if (itemType === "asset") {
        let dependencies = this.tasks[assetIndex].dependencies;
        this.tasks[assetIndex].dependencies = dependencies.filter(
          (dep) => dep !== dependency_id
        );
      } else {
        let entity_dependencies = this.tasks[assetIndex].entity_dependencies;
        this.tasks[assetIndex].entity_dependencies = entity_dependencies.filter(
          (dep) => dep !== dependency_id
        );
      }
    },
    changeAssetEntity(asset_id, entityId) {
      let assetIndex = this.tasks_index[asset_id];
      this.tasks[assetIndex].entity_id = entityId;
      this.rebuildAssetsIndex();
    },
    addAssetsCheckpoint(checkpoints) {
      for (let checkpoint of checkpoints) {
        let assetId = checkpoint.task_id;
        let assetIndex = this.tasks_index[assetId];
        console.log(this.tasks[assetIndex]);
        this.tasks[assetIndex].checkpoints.unshift(checkpoint);
      }
    },
    async getAssetFileStatus(asset) {
      if (!asset) {
        return "normal";
      }
      let filePath = asset.file_path;
      let checkpoints = asset.checkpoints;
      if (asset.is_link && (await isValidPointer(asset.pointer))) {
        return "normal";
      }
      if (asset.IsLink && (await isValidPointer(asset.pointer))) {
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
    getAssetEntity(assetId, recursive = false) {
      const collectionStore = useCollectionStore();
      let assetIndex = this.tasks_index[assetId];
      let asset = this.tasks[assetIndex];
      let entityId = asset.entity_id;
      let parentEntities = [];

      if (entityId) {
        let parentEntity = collectionStore.findCollection(entityId);
        if (parentEntity) {
          parentEntities.push(parentEntity);
        }
        if (recursive) {
          let parentId = parentEntity.parent_id;
          let depth = 0;
          while (parentId && depth < 20) {
            let parentEntity = collectionStore.findCollection(parentId);
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

    async processAssetsIconsAndPreviews(assets) {
      const iconStore = useIconStore();
      
      if (!assets || !Array.isArray(assets)) {
        return assets;
      }

      for (let i = 0; i < assets.length; i++) {
        let asset = assets[i];
        let extension = "";
        
        // Set file_path from pointer if it exists
        if (asset.pointer) {
          asset.file_path = asset.pointer;
        }
        
        // Determine extension
        if (asset.is_link && !isValidWeblink(asset.pointer)) {
          extension = utils.getFileExtension(asset.pointer).substring(1);
        } else if (!asset.is_link) {
          extension = asset.extension.toLowerCase().substring(1);
          
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
        if (asset.is_link && isValidWeblink(asset.pointer)) {
          iconPath = await iconStore.getWebIcon(asset.pointer);
        } else {
          iconPath = (await iconStore.getIcon(extension)) || "";
        }
        asset.icon = iconPath;
        
        // Set preview
        let preview = null;
        if (asset.preview) {
          preview = "data:image/png;base64," + asset.preview;
        }
        asset.preview = preview;
      }
      
      return assets;
    },

    async processUntrackedAssetsIcons(untrackedAssets) {
      const iconStore = useIconStore();
      
      if (!untrackedAssets || !Array.isArray(untrackedAssets)) {
        return untrackedAssets;
      }

      for (let i = 0; i < untrackedAssets.length; i++) {
        let extension = "";
        extension = untrackedAssets[i].extension.toLowerCase().substring(1);
        
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
        untrackedAssets[i].icon = iconPath;
      }
      
      return untrackedAssets;
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