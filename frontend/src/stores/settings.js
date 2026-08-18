import { defineStore } from "pinia";
import { SettingsService } from '@/services';

export const useSettingsStore = defineStore("settings", {
  state: () => ({
    bridgeEnabled: false,
    minimizeOnClose: true,
    overwriteDroppedFiles: true,
    pendingTab: null,
    showTypeIcons: true,
    locationsStale: false,
    systemBookmarksHealth: { projects_dir_stale: false, shared_projects_dir_stale: false },
    modalStates: {
      general: false,
      templates: false,
      collaborators: false,
      tags: false,
      workflows: false,
      roles: false,
      collectiontypes: false,
      assettypes: false,
      
      ignorelist: false,
      usertemplates: false,
      projecttemplates: false,
      directories: false,
      advanced: false,

      
      studio: false,
      studioprojects: false,
      studiocollaborators: false,
      studiointegrations: false,
    },

    activeModal: null,
    activeModalName: "",
    selectedStage: null,
    modalMaskVisible: false,
    showAssetCheckboxes: false,
    expandAllSubassets: false,
    firstSelectedAssetId: "",
    lastSelectedAssetId: "",

    sidePaneActive: false,

    allAssets: [],
    allAssetsCollapsed: true,

    markedAssets: [],
    allAssetsMarked: true,

    settingsItems: [
      { id: "general", nameKey: "settings.general", name: "General", icon: "monitor" },
      { id: "directories", nameKey: "settings.directories", name: "Directories", icon: "explorer" },
      { id: "templates", nameKey: "settings.templates", name: "Templates", icon: "file" },
      { id: "collaborators", nameKey: "settings.collaborators", name: "Collaborators", icon: "person" },
      { id: "roles", nameKey: "settings.roles", name: "Roles", icon: "scale" },

      { id: "assettypes", nameKey: "settings.assetTypes", name: "Asset types", icon: "brush" },
      { id: "collectiontypes", nameKey: "settings.collectionTypes", name: "Collection types", icon: "folder" },
      { id: "tags", nameKey: "settings.tags", name: "Tags", icon: "tag" },
      { id: "ignorelist", nameKey: "settings.ignoreList", name: "Ignore List", icon: "file-watch" },
      { id: "projecttemplates", nameKey: "settings.projectTemplates", name: "Project Templates", icon: "briefcase" },
      { id: "workflows", nameKey: "settings.workflows", name: "Workflows", icon: "workflow-arrow" },

      { id: "advanced", nameKey: "settings.advanced", name: "Advanced", icon: "skull" },

      { id: "studio", nameKey: "settings.studio", name: "Studio", icon: "stall" },
      { id: "studioprojects", nameKey: "settings.projectStorage", name: "Project Storage", icon: "briefcase" },
      { id: "studiocollaborators", nameKey: "settings.studioCollaborators", name: "Studio Collaborators", icon: "person" },
      { id: "studiointegrations", nameKey: "settings.studioIntegrations", name: "Integrations", icon: "plug" },
    ],

    templateContexts: [
      { id: "templates", nameKey: "settings.templates", name: "Templates", icon: "file" },
      { id: "assettypes", nameKey: "settings.assetTypes", name: "Asset types", icon: "brush" },
      { id: "collectiontypes", nameKey: "settings.collectionTypes", name: "Collection types", icon: "folder" },
      { id: "ignorelist", nameKey: "settings.ignoreList", name: "Ignore List", icon: "file-watch" },
    ],
  }),
  getters: {},
  actions: {
    // Refreshes the stale state of project locations (macOS security-scoped bookmarks).
    // Sets locationsStale to true when any configured location's bookmark needs re-selection.
    async refreshLocationsHealth() {
      try {
        const healthStatuses = await SettingsService.CheckAllLocationsHealth();
        const systemHealth = await SettingsService.CheckSystemBookmarksHealth();
        this.systemBookmarksHealth = systemHealth;
        const locationsStale = healthStatuses.some((h) => h.stale);
        const systemStale = systemHealth.projects_dir_stale || systemHealth.shared_projects_dir_stale;
        this.locationsStale = locationsStale || systemStale;
      } catch (error) {
        console.log(error);
      }
    },
    // Loads the bridge enabled state from user settings.
    async initializeBridge() {
      try {
        this.bridgeEnabled = await SettingsService.GetBridgeEnabled();
      } catch (error) {
        console.log(error);
      }
    },

    // Toggles the bridge on or off and persists the setting.
    async toggleBridge() {
      const newValue = !this.bridgeEnabled;
      await SettingsService.SetBridgeEnabled(newValue);
      this.bridgeEnabled = newValue;
    },

    // Loads the minimize on close state from user settings.
    async initializeMinimizeOnClose() {
      try {
        this.minimizeOnClose = await SettingsService.GetMinimizeOnClose();
      } catch (error) {
        console.log(error);
      }
    },

    // Toggles minimize on close and persists the setting.
    async toggleMinimizeOnClose() {
      const newValue = !this.minimizeOnClose;
      await SettingsService.SetMinimizeOnClose(newValue);
      this.minimizeOnClose = newValue;
    },

    // Loads the overwrite dropped files state from user settings.
    async initializeOverwriteDroppedFiles() {
      try {
        this.overwriteDroppedFiles = await SettingsService.GetOverwriteDroppedFiles();
      } catch (error) {
        console.log(error);
      }
    },

    // Toggles overwrite dropped files and persists the setting.
    async toggleOverwriteDroppedFiles() {
      const newValue = !this.overwriteDroppedFiles;
      await SettingsService.SetOverwriteDroppedFiles(newValue);
      this.overwriteDroppedFiles = newValue;
    },

    // Loads the show type icons state from user settings.
    async initializeShowTypeIcons() {
      try {
        this.showTypeIcons = await SettingsService.GetShowTypeIcons();
      } catch (error) {
        console.log(error);
      }
    },

    // Toggles show type icons and persists the setting.
    async toggleShowTypeIcons() {
      const newValue = !this.showTypeIcons;
      await SettingsService.SetShowTypeIcons(newValue);
      this.showTypeIcons = newValue;
    },

    setModalVisibility(modalName, value) {
      if (this.modalStates.hasOwnProperty(modalName)) {
        // Check if the modal is already active
        if (
          value &&
          this.activeModal !== null &&
          this.activeModal !== modalName
        ) {
          // Disable the currently active modal
          this.modalStates[this.activeModal] = false;
        }

        this.modalStates[modalName] = value;
        this.activeModal = value ? modalName : null;
        this.modalMaskVisible = this.isAnyModalActive();
      }
      this.selectedStage = modalName;
      // console.log(modalName);
    },
    disableAllModals() {
      for (const modalName in this.modalStates) {
        this.modalStates[modalName] = false;
      }
      this.activeModal = null;
      this.modalMaskVisible = false;
    },
    isAnyModalActive() {
      return Object.values(this.modalStates).some((isVisible) => isVisible);
    },
    toggleAssetCheckboxVisibility() {
      this.showAssetCheckboxes = !this.showAssetCheckboxes;
    },

    toggleSubassets(id) {
      //check if active sequence's index is included and then add it if not
      if (this.allAssets.includes(id)) {
        this.allAssets = this.allAssets.filter((i) => i !== id);
      } else {
        this.allAssets.push(id);
      }

      //collapse/expand all sequences based on whether the allAssets is empty
      if (this.allAssets.length >= 0) {
        this.allAssetsCollapsed = false;
      }
      if (this.allAssets.length == 0) {
        this.allAssetsCollapsed = true;
      }
    },

    toggleSubassetVisibility(fullData) {
      //ifthere are open sequences collapse all
      if (this.allAssets.length != 0) {
        this.allAssetsCollapsed = !this.allAssetsCollapsed;
        this.allAssets = [];
      } else {
        //if there are none, expand all
        const idArray = fullData.map((asset) => asset.id);
        this.allAssetsCollapsed = !this.allAssetsCollapsed;
        this.allAssets = idArray;
      }
    },

    markAsset(id) {
      //check if active sequence's id is included and then add it if not
      if (this.markedAssets.includes(id)) {
        this.markedAssets = this.markedAssets.filter((i) => i !== id);
        console.log(this.markedAssets);
      } else {
        this.markedAssets.push(id);
        console.log(this.markedAssets);
      }

      //collapse/expand all sequences based on whether the markedAssets is empty

      if (this.markedAssets.length >= 2) {
        this.allAssetsMarked = false;
      } else {
        this.allAssetsMarked = true;
        console.log(this.allAssetsMarked);
      }
    },
  },
});
