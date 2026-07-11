import { defineStore } from "pinia";
import {
  SettingsService,
  ProjectService,
  SyncService,
  FSService,
  StudioService,
} from "@/services";
import { useNotificationStore } from "./notifications";
import { useCommonStore } from "./common";
import { useUserStore } from "./users";
import { useCollectionStore } from "./collections";
import { useAssetStore } from '@/stores/assets';
import { useStageStore } from "./stages";
import { usePaneStore } from "./panes";
import { useTrayStates } from "./TrayStates";
import { useSettingsStore } from "./settings";
import { t } from "@/i18n";

let isProjectGridView = true;
await SettingsService.IsProjectGridView()
  .then((response) => {
    isProjectGridView = response;
  })
  .catch((error) => console.log(error));

let showUntrackedProjects = true;
await SettingsService.IsShowUntrackedProjects()
  .then((response) => {
    showUntrackedProjects = response;
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
    isProjectGridView: isProjectGridView,
    showUntrackedProjects: showUntrackedProjects,
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
    newUsers: {}, // Map of projectUri -> array of new user emails
    isProjectStatsExpanded: false,
    markedProjectIds: [],
    selectedProjects: [],
    firstSelectedProjectId: "",
    lastSelectedProjectId: "",
  }),
  getters: {
    getActiveProjectName: (state) => {
      if (state.projects.length && state.activeProject) {
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
      if (state.studios) {
        return state.studios.map((studio) => studio.name);
      }
      return [];
    },
    getActiveProject: (state) => {
      return state.activeProject;
    },
    getActiveProjectUrl: (state) => {
      if (state.activeProject?.has_remote && state.activeProject?.remote) {
        return state.activeProject.remote;
      }
      let projectName = state.activeProject.name;
      const projectUrl = state.getStudioUrl + "/" + projectName;
      return projectUrl;
    },
    getStudioUrl: (state) => {
      return state.studioUrl;
    },
    isCloudHosted: (state) => {
      if (state.selectedStudio?.name === 'Personal' && !!state.activeProject?.has_remote && !!state.activeProject?.remote) return true;
      if (state.selectedStudio?.hosting_mode === 'cloud') return true;
      return false;
    },
    isPersonalRemote: (state) => {
      return state.selectedStudio?.name === 'Personal'
        && !!state.activeProject?.has_remote
        && !!state.activeProject?.remote;
    },
    supportsIntegrations: (state) => {
      return state.selectedStudio?.name !== 'Personal';
    },

  },
  actions: {
    async setActiveProject(project) {
      const commonStore = useCommonStore();
      this.activeProject = project;
      FSService.SetProjectContext(project.uri);
      commonStore.workspaces = await SettingsService.GetProjectWorkspaces(
        project.id
      );
    },
    // Selectable means a tracked, downloaded project (untracked and undownloaded
    // projects are excluded from multi-select operations).
    isProjectSelectable(project) {
      return !!(project && project.is_tracked !== false && project.is_downloaded);
    },
    // Resets all multi-select state for projects.
    clearProjectSelection() {
      this.markedProjectIds = [];
      this.selectedProjects = [];
      this.firstSelectedProjectId = "";
      this.lastSelectedProjectId = "";
    },
    // Handles a click on a project row, applying ctrl/cmd toggle, shift range,
    // or single-select semantics. `allTrackedProjects` is the ordered, flat list
    // of currently-selectable projects used to compute shift-ranges.
    handleProjectClick(event, project, allTrackedProjects) {
      const stage = useStageStore();

      // Untracked or undownloaded projects bypass multi-select entirely.
      if (!this.isProjectSelectable(project)) {
        this.clearProjectSelection();
        this.setActiveProject(project);
        return;
      }

      const id = project.id;
      const isCmdOrCtrl = stage.cmdOrCtrlKey ? stage.cmdOrCtrlKey(event) : (event.ctrlKey || event.metaKey);

      if (isCmdOrCtrl) {
        if (!this.markedProjectIds.includes(id)) {
          this.markedProjectIds.push(id);
          this.selectedProjects.push(project);
          this.lastSelectedProjectId = id;
          if (!this.firstSelectedProjectId) this.firstSelectedProjectId = id;
          this.setActiveProject(project);
        } else {
          this.markedProjectIds = this.markedProjectIds.filter((i) => i !== id);
          this.selectedProjects = this.selectedProjects.filter((p) => p.id !== id);
          if (this.firstSelectedProjectId === id) {
            this.firstSelectedProjectId = this.markedProjectIds[0] || "";
          }
          if (this.lastSelectedProjectId === id) {
            this.lastSelectedProjectId = this.markedProjectIds[this.markedProjectIds.length - 1] || "";
          }
          if (this.activeProject?.id === id) {
            const fallback = this.selectedProjects[this.selectedProjects.length - 1];
            if (fallback) this.setActiveProject(fallback);
          }
        }
      } else if (event.shiftKey && this.firstSelectedProjectId) {
        const firstIndex = allTrackedProjects.findIndex((p) => p.id === this.firstSelectedProjectId);
        const lastIndex = allTrackedProjects.findIndex((p) => p.id === id);

        if (firstIndex === -1 || lastIndex === -1) {
          this.firstSelectedProjectId = id;
          this.lastSelectedProjectId = "";
          this.markedProjectIds = [id];
          this.selectedProjects = [project];
          this.setActiveProject(project);
          return;
        }

        const start = Math.min(firstIndex, lastIndex);
        const end = Math.max(firstIndex, lastIndex);
        const range = allTrackedProjects.slice(start, end + 1);

        this.markedProjectIds = range.map((p) => p.id);
        this.selectedProjects = [...range];
        this.lastSelectedProjectId = id;
        this.setActiveProject(project);
      } else {
        this.firstSelectedProjectId = id;
        this.lastSelectedProjectId = "";
        this.markedProjectIds = [id];
        this.selectedProjects = [project];
        this.setActiveProject(project);
      }
    },
    // Pins all selected projects that are not already pinned.
    async bulkPinProjects() {
      const studioName = this.getSelectedStudioName;
      const targets = this.selectedProjects.filter(
        (p) => p.is_downloaded && !this.pinnedProjects.includes(p.id)
      );
      for (const project of targets) {
        try {
          await SettingsService.PinProject(studioName, project.id);
          if (!this.pinnedProjects.includes(project.id)) {
            this.pinnedProjects.push(project.id);
          }
        } catch (error) {
          console.error("Error pinning project:", project.id, error);
        }
      }
    },
    // Unpins all selected projects that are currently pinned.
    async bulkUnpinProjects() {
      const studioName = this.getSelectedStudioName;
      const targets = this.selectedProjects.filter((p) => this.pinnedProjects.includes(p.id));
      for (const project of targets) {
        try {
          await SettingsService.UnpinProject(studioName, project.id);
          this.pinnedProjects = this.pinnedProjects.filter((id) => id !== project.id);
        } catch (error) {
          console.error("Error unpinning project:", project.id, error);
        }
      }
    },
    // Toggles the closed/archived state of all selected projects matching the
    // requested target state. `targetClosed=true` archives, `false` unarchives.
    async bulkToggleClosedProjects(targetClosed) {
      const notificationStore = useNotificationStore();
      const studioName = this.selectedStudio?.name || "";
      const targets = this.selectedProjects.filter((p) => !!p.is_closed !== !!targetClosed);
      for (const project of targets) {
        const uri = project.has_remote && project.remote ? project.remote : project.uri;
        try {
          await ProjectService.ToggleCloseProject(uri, studioName);
          const idx = this.projects.findIndex((p) => p.id === project.id);
          if (idx !== -1) this.projects[idx].is_closed = targetClosed;
          if (this.activeProject?.id === project.id) {
            this.activeProject.is_closed = targetClosed;
          }
        } catch (error) {
          console.error("Error toggling project closed state:", project.id, error);
          notificationStore.errorNotification("Error updating project", error);
        }
      }
    },
    // Removes the local .clst file for each selected project (server copy stays).
    // Only acts on projects that are downloaded and have a remote.
    async bulkRemoveProjects({ deleteWorkingFiles } = {}) {
      const targets = this.selectedProjects.filter(
        (p) => p.has_remote && p.is_downloaded
      );
      for (const project of targets) {
        try {
          await FSService.DeleteFile(project.uri);
          if (deleteWorkingFiles && project.working_directory) {
            await FSService.DeleteFolder(project.working_directory);
          }
          this.removeProjectFromList(project.uri);
        } catch (error) {
          console.error("Error removing project:", project.id, error);
        }
      }
    },
    async gotoProject(project) {
      const commonStore = useCommonStore();
      const collectionStore = useCollectionStore();
      const assetStore = useAssetStore();
      const stage = useStageStore();
      const panes = usePaneStore();
      const trayStates = useTrayStates();

      const settingsStore = useSettingsStore();
      if (settingsStore.locationsStale) {
        const notificationStore = useNotificationStore();
        notificationStore.addNotification(
          t("notifications.locationStaleTitle"),
          t("notifications.locationStaleMessage"),
          "danger",
          false
        );
        return;
      }

      await this.setActiveProject(project);
      commonStore.activeWorkspace = "Default";
      commonStore.resetFilters();
      commonStore.snapshotWorkspace();

      collectionStore.collections = [];
      assetStore.assets = [];

      commonStore.navigatorMode = false;
      collectionStore.navigatedCollection = null;
      collectionStore.selectedCollection = null;
      assetStore.selectedAsset = null;

      stage.expandedCollections = {};
      stage.setStageVisibility("browser", true);
      panes.setPaneVisibility("projectDetails", true);
      const studioName = this.getSelectedStudioName;
      SettingsService.AddRecentProject(studioName, project.id).then(
        (recentProjects) => {
          this.recentProjects = recentProjects;
        }
      );
    },
    addProjectToList(project) {
      if (!project) return;
      if (this.projects.some((p) => p.id === project.id || p.uri === project.uri)) return;
      const name = (project.name || "").toLowerCase();
      const insertAt = this.projects.findIndex(
        (p) => (p.name || "").toLowerCase().localeCompare(name) > 0
      );
      if (insertAt === -1) {
        this.projects.push(project);
      } else {
        this.projects.splice(insertAt, 0, project);
      }
    },
    // Removes a project from the in-memory list, or for remote projects marks it
    // as not-downloaded so it stays visible and can be re-downloaded.
    // Pass { force: true } to fully remove a remote project (e.g. deleted from the
    // server or left as a collaborator).
    removeProjectFromList(uri, { force = false } = {}) {
      if (!uri) return;
      const removed = this.projects.find((p) => p.uri === uri);

      if (removed && removed.has_remote && !force) {
        removed.is_downloaded = false;
        removed.is_unsynced = false;
        removed.preview = null;
        this.pinnedProjects = this.pinnedProjects.filter((p) => p.uri !== uri);
        this.recentProjects = this.recentProjects.filter((p) => p.uri !== uri);
        if (this.activeProject?.uri === uri) {
          this.activeProject = null;
        }
        this.markedProjectIds = this.markedProjectIds.filter((id) => id !== removed.id);
        this.selectedProjects = this.selectedProjects.filter((p) => p.id !== removed.id);
        if (this.firstSelectedProjectId === removed.id) this.firstSelectedProjectId = this.markedProjectIds[0] || "";
        if (this.lastSelectedProjectId === removed.id) this.lastSelectedProjectId = "";
        return;
      }

      this.projects = this.projects.filter((p) => p.uri !== uri);
      this.pinnedProjects = this.pinnedProjects.filter((p) => p.uri !== uri);
      this.recentProjects = this.recentProjects.filter((p) => p.uri !== uri);
      if (this.activeProject?.uri === uri) {
        this.activeProject = null;
      }
      if (removed) {
        this.markedProjectIds = this.markedProjectIds.filter((id) => id !== removed.id);
        this.selectedProjects = this.selectedProjects.filter((p) => p.id !== removed.id);
        if (this.firstSelectedProjectId === removed.id) this.firstSelectedProjectId = this.markedProjectIds[0] || "";
        if (this.lastSelectedProjectId === removed.id) this.lastSelectedProjectId = "";
      }
    },
    async loadProjects() {
      const notificationStore = useNotificationStore();
      this.projectsLoaded = false;

      let studio = this.selectedStudio;
      let studioUrl;
      try {
        studioUrl = await StudioService.ResolveStudioUrl(studio.url, studio.alt_url || "");
      } catch {
        studioUrl = studio.url;
      }
      this.studioUrl = studioUrl;

      // Update studio reachability after URL resolution
      const { useStudioStore } = await import('./studio');
      const studioStore = useStudioStore();
      if (studio.name !== 'Personal') {
        studioStore.checkStudioReachability();
      }

      SettingsService.GetPinnedProjects(studio.name).then((response) => {
        this.pinnedProjects = response;
      });
      SettingsService.GetRecentProjects(studio.name).then((response) => {
        this.recentProjects = response;
      });
      await ProjectService.GetStudioProjects(studioUrl, studio.name, studio.hosting_mode || '', studio.id || '')
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
      
      // Parallelize project status checks
      await Promise.all(this.projects.map(async (project, i) => {
        //TODO check the importance if exist for projects
        if (await FSService.Exists(project.uri)) {
          this.projects[i].is_downloaded = true;
          try {
            const isUnsynced = await SyncService.IsUnsynced(project.uri);
            this.projects[i].is_unsynced = isUnsynced;
          } catch (error) {
            console.log(error);
            // notificationStore.errorNotification("Error Loading Data", error)
          }
        } else {
          this.projects[i].is_downloaded = false;
          this.projects[i].is_unsynced = false;
        }
      }));
      
      if (this.activeProject && stage.activeStage !== 'projects') {
        await this.refreshActiveProject();
      }

      this.projectsLoaded = true;
    },
    async refreshProjectsPreview() {
      // Parallelize preview fetching
      await Promise.all(this.projects.map(async (project, i) => {
        if (await FSService.Exists(project.uri)) {
          try {
            const preview = await ProjectService.GetPreview(project.uri);
            if (preview) {
              this.projects[i].preview = "data:image/png;base64," + preview;
            }
          } catch (error) {
            console.log(error);
            // notificationStore.errorNotification("Error Loading Data", error)
          }
        }
      }));
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

          // Fetch studio users for non-Personal studios so permissions are available immediately
          if (this.selectedStudio?.name !== 'Personal') {
            const { useStudioStore } = await import('./studio');
            const studioStore = useStudioStore();
            await studioStore.getStudioUsers();
          }
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
      this.clearProjectSelection();
      commonStore.resetFilters();
      commonStore.snapshotWorkspace();
      await this.loadProjects();
      SettingsService.SetLastStudio(studio.name)
    },
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
      await this.rebuildUntrackedAssetIndex();
      await this.rebuildUntrackedCollectionIndex();
    },
    async rebuildUntrackedAssetIndex() {
      let untrackedFilesIndex = {};
      for (let i = 0; i < this.untrackedFiles.length; i++) {
        let untrackedAssetId = this.untrackedFiles[i].id;
        untrackedFilesIndex[untrackedAssetId] = i;
      }
      this.untrackedFilesIndex = untrackedFilesIndex;
    },
    async rebuildUntrackedCollectionIndex() {
      let untrackedFoldersIndex = {};
      for (let i = 0; i < this.untrackedFolders.length; i++) {
        let untrackedCollectionId = this.untrackedFolders[i].id;
        untrackedFoldersIndex[untrackedCollectionId] = i;
      }
      this.untrackedFoldersIndex = untrackedFoldersIndex;
    },
    findUntrackedAsset(id) {
      let untrackedAssetIndex = this.untrackedFilesIndex[id];
      return this.untrackedFiles[untrackedAssetIndex];
    },
    findUntrackedCollection(id) {
      let untrackedFolderIndex = this.untrackedFoldersIndex[id];
      return this.untrackedFolders[untrackedFolderIndex];
    },
    removeUntrackedAsset(id) {
      let untrackedAssetIndex = this.untrackedFilesIndex[id];
      this.untrackedFiles.splice(untrackedAssetIndex, 1);
    },
    removeUntrackedCollection(id) {
      let untrackedFolderIndex = this.untrackedFoldersIndex[id];
      this.untrackedFolders.splice(untrackedFolderIndex, 1);
    },
    async getUntrackedItems() {
      const stage = useStageStore();
      if (stage.activeStage !== "browser") {
        return;
      }
      const userStore = useUserStore();
      const collectionStore = useCollectionStore();
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
        // let assets = untrackedItems.assets;
        // let collections = untrackedItems.collections;
        // for (let asset of assets) {
        //   const itemPath = asset.asset_path
        //     .replace(/^\/+|\/+$/g, "")
        //     .replace(/\\/g, "/");
        //   let collectionPath = "";
        //   const itemPathCollections = itemPath.split("/");

        //   if (itemPathCollections.length > 1) {
        //     // Take all elements except the last one
        //     const pathWithoutLast = itemPathCollections.slice(0, -1);
        //     collectionPath = pathWithoutLast.join("/");
        //   }
        //   let untrackedFile = {
        //     ...asset,
        //     asset_type_icon: "generic",
        //     item_path: itemPath,
        //     collection_path: collectionPath,
        //     item_type: "file",
        //     type: "untracked_asset",
        //   };
        //   untrackedFiles.push(untrackedFile);
        // }
        // for (let collection of collections) {
        //   if (collection.file_path == "") {
        //     continue;
        //   }
        //   const itemPath = collection.collection_path
        //     .replace(/^\/+|\/+$/g, "")
        //     .replace(/\\/g, "/");
        //   // Handle collection path calculation
        //   let collectionPath = "";
        //   const itemPathCollections = itemPath.split("/");

        //   if (itemPathCollections.length > 1) {
        //     // Take all elements except the last one
        //     const pathWithoutLast = itemPathCollections.slice(0, -1);
        //     collectionPath = pathWithoutLast.join("/");
        //   }
        //   let untrackedFolder = {
        //     ...collection,
        //     collection_type_icon: "folder",
        //     item_path: itemPath,
        //     collection_path: collectionPath,
        //     item_type: "folder",
        //     type: "untracked_collection",
        //   };
        //   untrackedFolders.push(untrackedFolder);
        // }

        // // Sort function for both files and folders
        // const sortByPathAndName = (a, b) => {
        //   // First compare by collection_path
        //   if (a.collection_path !== b.collection_path) {
        //     return a.collection_path.localeCompare(b.collection_path);
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
    
    // New user management methods
    addNewUsersToProject(projectUri, userEmails) {
      if (!this.newUsers[projectUri]) {
        this.newUsers[projectUri] = [];
      }
      
      // Add only unique emails that aren't already tracked
      const uniqueEmails = userEmails.filter(email => 
        !this.newUsers[projectUri].includes(email)
      );
      
      this.newUsers[projectUri].push(...uniqueEmails);
    },
    
    getNewUsersForProject(projectUri) {
      return this.newUsers[projectUri] || [];
    },
    
    clearNewUsersForProject(projectUri) {
      delete this.newUsers[projectUri];
    },
    
    hasNewUsersForProject(projectUri) {
      return this.newUsers[projectUri] && this.newUsers[projectUri].length > 0;
    },
    
    updateProjectName(projectId, newName, oldName) {
      // Find and update the project in the projects array
      const projectIndex = this.projects.findIndex(p => p.id === projectId);
      if (projectIndex !== -1) {
        const project = this.projects[projectIndex];
        project.name = newName;
        
        // Update URI: replace old name with new name in the path
        if (project.uri) {
          project.uri = project.uri.replace(`${oldName}.clst`, `${newName}.clst`);
        }
        
        // Update remote URL if it exists
        if (project.remote) {
          project.remote = project.remote.replace(`${oldName}`, `${newName}`);
        }
      }
      
      // Update activeProject if it's the one being renamed
      if (this.activeProject && this.activeProject.id === projectId) {
        this.activeProject.name = newName;
        
        // Update URI for active project
        if (this.activeProject.uri) {
          this.activeProject.uri = this.activeProject.uri.replace(`${oldName}.clst`, `${newName}.clst`);
        }
        
        // Update remote URL for active project
        if (this.activeProject.remote) {
          this.activeProject.remote = this.activeProject.remote.replace(`${oldName}`, `${newName}`);
        }
      }
    },

    // Opens a .clst file from an arbitrary location.
    // If the project belongs to a known studio and is already listed, navigates to it.
    // Otherwise, temporarily adds it to the Personal studio project list.
    async openClusttaFile(fileInfo) {
      const notificationStore = useNotificationStore();
      const stage = useStageStore();

      const studioName = fileInfo.studio_name || "Personal";

      // Check if the project's studio matches the currently selected studio
      const matchingStudio = this.studios.find(s => s.name === studioName);

      if (matchingStudio && matchingStudio.name === this.selectedStudio?.name) {
        // Same studio - check if project is already in the list
        const existingProject = this.projects.find(p => p.id === fileInfo.id);
        if (existingProject) {
          await this.gotoProject(existingProject);
          stage.setStageVisibility("browser", true);
          return;
        }
      }

      if (matchingStudio && matchingStudio.name !== "Personal") {
        // Project belongs to a different known studio - switch to it and look for the project
        await this.selectStudio(matchingStudio);
        const existingProject = this.projects.find(p => p.id === fileInfo.id);
        if (existingProject) {
          await this.gotoProject(existingProject);
          stage.setStageVisibility("browser", true);
          return;
        }
      }

      // Project is not in any known studio directory - open temporarily in first studio
      const fallbackStudio = this.studios[0];
      if (fallbackStudio && this.selectedStudio?.name !== fallbackStudio.name) {
        await this.selectStudio(fallbackStudio);
      }

      // Build a temporary ProjectInfo from the file info and add to list
      let projectInfo;
      try {
        projectInfo = await ProjectService.ProjectInfo(fileInfo.file_path);
      } catch (error) {
        notificationStore.errorNotification("Failed to read project file", error);
        return;
      }

      projectInfo.uri = fileInfo.file_path;
      projectInfo.is_downloaded = true;
      projectInfo.is_external = true;

      // Avoid duplicates
      const alreadyAdded = this.projects.find(p => p.id === projectInfo.id);
      if (alreadyAdded) {
        await this.gotoProject(alreadyAdded);
      } else {
        this.projects.unshift(projectInfo);
        await this.gotoProject(projectInfo);
      }

      stage.setStageVisibility("browser", true);
    },
  },
});




