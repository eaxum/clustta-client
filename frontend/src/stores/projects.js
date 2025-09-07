import { defineStore } from "pinia";
import {
  SettingsService,
  ProjectService,
  SyncService,
  FSService,
} from "@/../bindings/clustta/services";
import { useNotificationStore } from "./notifications";
import { useCommonStore } from "./common";
import { useUserStore } from "./users";
import { useEntityStore } from "./entity";
import { useAssetStore } from '@/stores/assets';
import { useStageStore } from "./stages";
import { usePaneStore } from "./panes";
import { useTrayStates } from "./TrayStates";

let useAltUrl = false;
await SettingsService.GetUseAltUrl()
  .then((response) => {
    useAltUrl = response;
  })
  .catch((error) => console.log(error));

let isProjectGridView = true;
await SettingsService.IsProjectGridView()
  .then((response) => {
    isProjectGridView = response;
  })
  .catch((error) => console.log(error));

let lastStudio = "";
await SettingsService.GetLastStudio()
  .then((response) => {
    lastStudio = response;
  })
  .catch((error) => console.log(error));

export const useProjectStore = defineStore("projects", {
  state: () => ({
    activeProject: null,
    projectSearchQuery: "",
    activeProjectCover: "",
    pinnedProjects: [],
    recentProjects: [],
    studioUrl: "",
    useAltUrl: useAltUrl,
    isProjectGridView: isProjectGridView,
    lastStudio: lastStudio,
    projects: [],
    projectsLoaded: false,
    selectedStudio: null,
    studios: [],
    untrackedFiles: [],
    untrackedFilesIndex: {},
    untrackedFolders: [],
    selectedUntrackedItem: null,
    untrackedFoldersIndex: {},
  }),
  getters: {
    getActiveProjectName: (state) => {
      //get base name of the project
      if (state.projects.length) {
        if (!state.activeProject) {
          state.activeProject = state.projects[0];
        }
        let project = state.activeProject;
        return project.name;
      }
      return "";
    },
    getSelectedStudioName: (state) => {
      if (state.selectedStudio) {
        return state.selectedStudio.name;
      }
      return "";
    },
    getProjects: (state) => {
      return state.projects;
    },
    getStudiosNames: (state) => {
      console.log(state.studios);
      if (state.studios) {
        return state.studios.map((studio) => studio.name);
      }
      return [];
    },
    getActiveProject: (state) => {
      if (!state.activeProject) {
        state.activeProject = state.projects[0];
        if (state.activeProject) {
          const commonStore = useCommonStore();

          SettingsService.GetProjectWorkspaces(state.activeProject.id).then(
            (response) => {
              commonStore.workspaces = response;
            }
          );
        }

        return state.activeProject;
      }
      return state.activeProject;
    },
    getActiveProjectUrl: (state) => {
      let projectName = state.activeProject.name;
      const projectUrl = state.getStudioUrl + "/" + projectName;
      return projectUrl;
    },
    getStudioUrl: (state) => {
      let studio = state.selectedStudio;
      if (studio) {
        if (studio.alt_url == "") {
          return studio.url;
        }
        return !state.useAltUrl ? studio.url : studio.alt_url;
      }
      return "";
    },
  },
  actions: {
    async setActiveProject(project) {
      const commonStore = useCommonStore();
      this.activeProject = project;
      commonStore.workspaces = await SettingsService.GetProjectWorkspaces(
        project.id
      );
    },
    async gotoProject(project) {
      const commonStore = useCommonStore();
      const entityStore = useEntityStore();
      const assetStore = useAssetStore();
      const stage = useStageStore();
      const panes = usePaneStore();
      const trayStates = useTrayStates();

      
 
      await this.setActiveProject(project);

      if (!project.is_downloaded) {
        return;
      }

      commonStore.activeWorkspace = "Default";
      commonStore.resetFilters();

      entityStore.entities = [];
      assetStore.tasks = [];

      commonStore.navigatorMode = false;
      entityStore.navigatedEntity = null;
      entityStore.selectedEntity = null;
      assetStore.selectedTask = null;

      stage.expandedEntities = {};
      stage.setStageVisibility("browser", true);
      panes.setPaneVisibility("projectDetails", true);
      const studioName = this.getSelectedStudioName;
      SettingsService.AddRecentProject(studioName, project.id).then(
        (recentProjects) => {
          this.recentProjects = recentProjects;
        }
      );
    },
    async loadProjects() {
      const notificationStore = useNotificationStore();
      this.projectsLoaded = false;

      let studio = this.selectedStudio;
      const studioUrl = !(studio.alt_url && this.useAltUrl)
        ? studio.url
        : studio.alt_url;
      this.studioUrl = studioUrl;

      SettingsService.GetPinnedProjects(studio.name).then((response) => {
        this.pinnedProjects = response;
      });
      SettingsService.GetRecentProjects(studio.name).then((response) => {
        this.recentProjects = response;
      });
      await ProjectService.GetStudioProjects(studioUrl, studio.name)
        .then(async (response) => {
          this.projects = response;
        })
        .catch((error) => {
          console.error(error);
          notificationStore.errorNotification("Error loading projects", error);
        });

      await this.refreshProjects();
      await this.refreshProjectsPreview();
    },
    async reloadActiveProject() {
      if (this.activeProject) {
        await ProjectService.ProjectInfo(this.activeProject.uri)
          .then((response) => {
            this.activeProject.sync_token = response.sync_token;
            // find the project in the projects array and update it
            let projectIndex = this.projects.findIndex((project) => {
              return project.id === this.activeProject.id;
            });
            if (projectIndex !== -1) {
              this.projects[projectIndex] = this.activeProject;
            }

            this.refreshActiveProject();
          })
          .catch((error) => {
            console.error(error);
          });
      }
    },
    async refreshProjects() {
      
      const stage = useStageStore();

      this.projectsLoaded = false;
      for (let i = 0; i < this.projects.length; i++) {
        //TODO check the importance if exist for projects
        if (await FSService.Exists(this.projects[i].uri)) {
          this.projects[i].is_downloaded = true;
          await SyncService.IsUnsynced(this.projects[i].uri)
            .then(async (isUnsynced) => {
              this.projects[i].is_unsynced = isUnsynced;
            })
            .catch((error) => {
              console.log(error);
              // notificationStore.errorNotification("Error Loading Data", error)
            });
        } else {
          this.projects[i].is_downloaded = false;
          this.projects[i].is_unsynced = false;
        }
      }
        console.log(this.activeProject)
      if (this.activeProject && stage.activeStage !== 'projects') {
        console.log('reloading')
        await this.refreshActiveProject();
      }

      this.projectsLoaded = true;
    },
    async refreshProjectsPreview() {
      for (let i = 0; i < this.projects.length; i++) {
        if (await FSService.Exists(this.projects[i].uri)) {
          await ProjectService.GetPreview(this.projects[i].uri)
            .then(async (preview) => {
              if (preview) {
                this.projects[i].preview = "data:image/png;base64," + preview;
              }
            })
            .catch((error) => {
              console.log(error);
              // notificationStore.errorNotification("Error Loading Data", error)
            });
        }
      }
    },
    async refreshProjectPreview(projectId) {
      let projectIndex = this.projects.findIndex((project) => {
        return project.id === projectId;
      });
      let project = this.projects[projectIndex];
      if (await FSService.Exists(project.uri)) {
        await ProjectService.GetPreview(project.uri)
          .then(async (preview) => {
            this.projects[projectIndex].preview =
              "data:image/png;base64," + preview;
          })
          .catch((error) => {
            console.log(error);
            // notificationStore.errorNotification("Error Loading Data", error)
          });
      }
    },
    async refreshActiveProject() {
      let project = this.getActiveProject;
      if (project) {
        if (await FSService.Exists(project.uri)) {
          project.is_downloaded = true;
          await SyncService.IsUnsynced(project.uri)
            .then(async (isUnsynced) => {
              project.is_unsynced = isUnsynced;
            })
            .catch((error) => {
              console.log(error);
              // notificationStore.errorNotification("Error Loading Data", error)
            });
        } else {
          project.is_downloaded = false;
          project.is_unsynced = false;
        }
      }
    },
    async loadStudios() {
      const notificationStore = useNotificationStore();
      await SettingsService.GetStudios()
        .then(async (data) => {
          this.studios = data;
          let lastSelectedStudio = this.studios.find((item) => item.name === lastStudio)
          this.selectedStudio = lastSelectedStudio ? lastSelectedStudio: data[0] ;
        })
        .catch((error) => {
          console.log(error);
          notificationStore.errorNotification("Loading Studios", error);
        });
    },
    async selectStudio(studio) {
      const commonStore = useCommonStore();
      this.activeProject = null;
      this.projects = [];
      this.selectedStudio = studio;
      commonStore.resetFilters();
      await this.loadProjects();
      SettingsService.SetLastStudio(studio.name)
    },
    // async reloadUntrackedItems() {
    //   const stage = useStageStore();
    //   // if (stage.activeStage !== "browser") {
    //   //   return;
    //   // }
    //   const userStore = useUserStore();
    //   const entityStore = useEntityStore();
    //   const assetStore = useAssetStore();
    //   let project = this.activeProject;
    //   if (project) {
    //     let untrackedItems = await ProjectService.GetUntrackedItems(
    //       project.working_directory,
    //       project.uri,
    //       project.ignore_list
    //     );

    //     console.log(untrackedItems);

    //     let untrackedFiles = [];
    //     let untrackedFolders = [];
    //     let projectWorkingDir = project.working_directory;
    //     let tasks = untrackedItems.tasks;
    //     let entities = untrackedItems.entities;
    //     for (let task of tasks) {
    //       const itemPath = task.task_path
    //         .replace(/^\/+|\/+$/g, "")
    //         .replace(/\\/g, "/");
    //       let entityPath = "";
    //       const itemPathEntities = itemPath.split("/");

    //       if (itemPathEntities.length > 1) {
    //         // Take all elements except the last one
    //         const pathWithoutLast = itemPathEntities.slice(0, -1);
    //         entityPath = pathWithoutLast.join("/");
    //       }
    //       let untrackedFile = {
    //         ...task,
    //         task_type_icon: "generic",
    //         item_path: itemPath,
    //         entity_path: entityPath,
    //         item_type: "file",
    //         type: "untracked_task",
    //       };
    //       untrackedFiles.push(untrackedFile);
    //     }
    //     for (let entity of entities) {
    //       if (entity.file_path == "") {
    //         continue;
    //       }
    //       const itemPath = entity.entity_path
    //         .replace(/^\/+|\/+$/g, "")
    //         .replace(/\\/g, "/");
    //       // Handle entity path calculation
    //       let entityPath = "";
    //       const itemPathEntities = itemPath.split("/");

    //       if (itemPathEntities.length > 1) {
    //         // Take all elements except the last one
    //         const pathWithoutLast = itemPathEntities.slice(0, -1);
    //         entityPath = pathWithoutLast.join("/");
    //       }
    //       let untrackedFolder = {
    //         ...entity,
    //         entity_type_icon: "folder",
    //         item_path: itemPath,
    //         entity_path: entityPath,
    //         item_type: "folder",
    //         type: "untracked_entity",
    //       };
    //       untrackedFolders.push(untrackedFolder);
    //     }

    //     // Sort function for both files and folders
    //     const sortByPathAndName = (a, b) => {
    //       // First compare by entity_path
    //       if (a.entity_path !== b.entity_path) {
    //         return a.entity_path.localeCompare(b.entity_path);
    //       }
    //       // Then compare by name
    //       return a.name.localeCompare(b.name);
    //     };

    //     // Sort both arrays
    //     untrackedFiles.sort(sortByPathAndName);
    //     untrackedFolders.sort(sortByPathAndName);

    //     this.untrackedFiles = untrackedFiles;
    //     this.untrackedFolders = untrackedFolders;
    //     await this.rebuildUntrackedTaskIndex();
    //     await this.rebuildUntrackedEntityIndex();
    //   }
    // },
    async refreshUntrackedItems() {
      let untrackedFiles = [];
      let untrackedFolders = [];
      for (let untrackedFile of this.untrackedFiles) {
        let isIgnored = await ProjectService.IsIgnored(
          untrackedFile.item_path + untrackedFile.extension,
          this.activeProject.ignore_list
        );
        if (isIgnored) {
          continue;
        }
        untrackedFiles.push(untrackedFile);
      }
      for (let untrackedFolder of this.untrackedFolders) {
        let isIgnored = await ProjectService.IsIgnored(
          untrackedFolder.item_path,
          this.activeProject.ignore_list
        );
        if (isIgnored) {
          continue;
        }
        untrackedFolders.push(untrackedFolder);
      }
      this.untrackedFiles = untrackedFiles;
      this.untrackedFolders = untrackedFolders;
      await this.rebuildUntrackedTaskIndex();
      await this.rebuildUntrackedEntityIndex();
    },
    async rebuildUntrackedTaskIndex() {
      let untrackedFilesIndex = {};
      for (let i = 0; i < this.untrackedFiles.length; i++) {
        let untrackedTaskId = this.untrackedFiles[i].id;
        untrackedFilesIndex[untrackedTaskId] = i;
      }
      this.untrackedFilesIndex = untrackedFilesIndex;
    },
    async rebuildUntrackedEntityIndex() {
      let untrackedFoldersIndex = {};
      for (let i = 0; i < this.untrackedFolders.length; i++) {
        let untrackedEntityId = this.untrackedFolders[i].id;
        untrackedFoldersIndex[untrackedEntityId] = i;
      }
      this.untrackedFoldersIndex = untrackedFoldersIndex;
    },
    findUntrackedTask(id) {
      let untrackedTaskIndex = this.untrackedFilesIndex[id];
      return this.untrackedFiles[untrackedTaskIndex];
    },
    findUntrackedEntity(id) {
      let untrackedFolderIndex = this.untrackedFoldersIndex[id];
      return this.untrackedFolders[untrackedFolderIndex];
    },
    removeUntrackedTask(id) {
      let untrackedTaskIndex = this.untrackedFilesIndex[id];
      this.untrackedFiles.splice(untrackedTaskIndex, 1);
    },
    removeUntrackedEntity(id) {
      let untrackedFolderIndex = this.untrackedFoldersIndex[id];
      this.untrackedFolders.splice(untrackedFolderIndex, 1);
    },
    async getUntrackedItems() {
      const stage = useStageStore();
      if (stage.activeStage !== "browser") {
        return;
      }
      const userStore = useUserStore();
      const entityStore = useEntityStore();
      const assetStore = useAssetStore();
      const ignoredExtensions = ["blend1", ".tif", "tmp"];
      let project = this.activeProject;
      if (project) {

        // let untrackedItems = await ProjectService.GetUntrackedItems(
        //   project.working_directory,
        //   project.uri,
        //   ignoredExtensions
        // );

        // let untrackedFiles = [];
        // let untrackedFolders = [];
        // let projectWorkingDir = project.working_directory;
        // let tasks = untrackedItems.tasks;
        // let entities = untrackedItems.entities;
        // for (let task of tasks) {
        //   const itemPath = task.task_path
        //     .replace(/^\/+|\/+$/g, "")
        //     .replace(/\\/g, "/");
        //   let entityPath = "";
        //   const itemPathEntities = itemPath.split("/");

        //   if (itemPathEntities.length > 1) {
        //     // Take all elements except the last one
        //     const pathWithoutLast = itemPathEntities.slice(0, -1);
        //     entityPath = pathWithoutLast.join("/");
        //   }
        //   let untrackedFile = {
        //     ...task,
        //     task_type_icon: "generic",
        //     item_path: itemPath,
        //     entity_path: entityPath,
        //     item_type: "file",
        //     type: "untracked_task",
        //   };
        //   untrackedFiles.push(untrackedFile);
        // }
        // for (let entity of entities) {
        //   if (entity.file_path == "") {
        //     continue;
        //   }
        //   const itemPath = entity.entity_path
        //     .replace(/^\/+|\/+$/g, "")
        //     .replace(/\\/g, "/");
        //   // Handle entity path calculation
        //   let entityPath = "";
        //   const itemPathEntities = itemPath.split("/");

        //   if (itemPathEntities.length > 1) {
        //     // Take all elements except the last one
        //     const pathWithoutLast = itemPathEntities.slice(0, -1);
        //     entityPath = pathWithoutLast.join("/");
        //   }
        //   let untrackedFolder = {
        //     ...entity,
        //     entity_type_icon: "folder",
        //     item_path: itemPath,
        //     entity_path: entityPath,
        //     item_type: "folder",
        //     type: "untracked_entity",
        //   };
        //   untrackedFolders.push(untrackedFolder);
        // }

        // // Sort function for both files and folders
        // const sortByPathAndName = (a, b) => {
        //   // First compare by entity_path
        //   if (a.entity_path !== b.entity_path) {
        //     return a.entity_path.localeCompare(b.entity_path);
        //   }
        //   // Then compare by name
        //   return a.name.localeCompare(b.name);
        // };

        // // Sort both arrays
        // untrackedFiles.sort(sortByPathAndName);
        // untrackedFolders.sort(sortByPathAndName);

        // this.untrackedFiles = untrackedFiles;
        // this.untrackedFolders = untrackedFolders;

      }
    },
    selectUntrackedItem(item) {
      this.selectedUntrackedItem = item;
    },
  },
});

