import { defineStore } from "pinia";
import {
  SettingsService,
} from "@/services";
import { useCollectionStore } from "@/stores/collections";

let defaultViewMode = 'dense';
await SettingsService.GetDefaultViewMode()
  .then((response) => {
    defaultViewMode = response;
  })
  .catch((error) => console.log(error));

let defaultShowUntracked = true;
await SettingsService.GetUntrackedVisibility()
  .then((response) => {
    defaultShowUntracked = response;
  })
  .catch((error) => console.log(error));

export const useCommonStore = defineStore("common", {
  state: () => ({
    activeFilters: [],
    resourceFilters: [],
    assetFilters: [],
    collectionFilters: [],
    hasAssignees: false,
    noAssignees: false,
    reloadFilters: false,
    showFullPath: false,
    hideExtensions: true,
    showThumbs: true,
    showUntracked: defaultShowUntracked,
    showCollections: true,
    showAssets: true,
    onlyAssets: false,
    showTasks: true,
    showResources: true,
    showChildCollections: true,
    showChildAssets: true,
    showChildResources: true,
    showDependencies: true,
    useDeep: false,
    navigatorMode: false,
    useGrid: defaultViewMode === 'grid',
    viewMode: defaultViewMode,
    previousViewMode: null,
    sortBy: 'name',
    sortOrder: 'asc',
    gridSize: 200,
    listItemHeight: defaultViewMode === 'dense' ? 42 : 60,
    listItemGap: defaultViewMode === 'dense' ? 2 : 4,

    filterDependencyAssets: true,
    filterDependencyCollections: true,
    filterDependencyResources: true,

    viewSearchQuery: "",
    workspaceSearchQuery: "",
    fileStates: [
      { name: "normal", type: "state", icon: "circle-check" },
      { name: "modified", type: "state", icon: "layers-plus-alert" },
      { name: "outdated", type: "state", icon: "circle-check-alert" },
      { name: "fetchable", type: "state", icon: "fetch" },
      { name: "missing", type: "state", icon: "alert" },
    ],
    syncOptions: [
      { name: "All", active: false, icon: "four-squares" },
      {
        name: "Only Latest Checkpoints",
        active: true,
        icon: "layers",
      },
      { name: "Dependencies", active: true, icon: "dependency" },
      { name: "All Assets/Resources", active: false, icon: "brush" },
      { name: "Templates", active: false, icon: "file" },
    ],
    workspaces: [],
    projectWorkflows: [],
    activeWorkspace: "Default",
    savedWorkspaceSnapshot: null,
    ghostCardStyle: {
      leaving: false,
      pos: { x: 0, y: 0 },
      width: 0,
      cursorDistance: { x: 0, y: 0 },
      transform: "",
    },
  }),
  getters: {
    getCollections: (state) => {
      return state.collections;
    },
    getAssets: (state) => {
      return state.assets;
    },
    getResources: (state) => {
      return state.resources;
    },
    isWorkspaceDirty: (state) => {
      if (!state.savedWorkspaceSnapshot) return false;
      const snap = state.savedWorkspaceSnapshot;
      const collectionStore = useCollectionStore();
      const currentPath = collectionStore.navigatedCollection?.collection_path
        || collectionStore.navigatedCollection?.item_path
        || null;
      return (
        JSON.stringify(state.assetFilters) !== JSON.stringify(snap.assetFilters) ||
        JSON.stringify(state.collectionFilters) !== JSON.stringify(snap.collectionFilters) ||
        JSON.stringify(state.resourceFilters) !== JSON.stringify(snap.resourceFilters) ||
        state.showCollections !== snap.showCollections ||
        state.showAssets !== snap.showAssets ||
        state.onlyAssets !== snap.onlyAssets ||
        state.showTasks !== snap.showTasks ||
        state.showResources !== snap.showResources ||
        state.showChildCollections !== snap.showChildCollections ||
        state.showChildAssets !== snap.showChildAssets ||
        state.showChildResources !== snap.showChildResources ||
        state.showDependencies !== snap.showDependencies ||
        state.useDeep !== snap.useDeep ||
        state.hasAssignees !== snap.hasAssignees ||
        state.noAssignees !== snap.noAssignees ||
        state.viewSearchQuery !== snap.workspaceSearchQuery ||
        state.viewMode !== snap.viewMode ||
        currentPath !== snap.collectionPath
      );
    },
  },
  actions: {
    setActiveWorkspace(workspace) {
      this.activeWorkspace = workspace.name;
      this.assetFilters = workspace.filters.assetFilters;
      this.collectionFilters = workspace.filters.collectionFilters;
      this.resourceFilters = workspace.filters.resourceFilters;

      this.showCollections = workspace.filters.showCollections;
      this.showAssets = workspace.filters.showAssets;
      this.onlyAssets = workspace.filters.onlyAssets;
      this.showTasks = workspace.filters.showTasks ?? true;
      this.showResources = workspace.filters.showResources;
      this.showChildCollections = workspace.filters.showChildCollections;
      this.showChildAssets = workspace.filters.showChildAssets;
      this.showChildResources = workspace.filters.showChildResources;
      this.showDependencies = workspace.filters.showDependencies;
      this.useDeep = workspace.filters.useDeep;

      this.hasAssignees = workspace.filters.hasAssignees;
      this.noAssignees = workspace.filters.noAssignees;

      this.workspaceSearchQuery = workspace.workspaceSearchQuery;
      this.viewSearchQuery = workspace.workspaceSearchQuery || '';
      this.applyViewMode(workspace.viewMode || defaultViewMode);
      this.snapshotWorkspace();
    },
    applyViewMode(mode) {
      this.viewMode = mode;
      this.useGrid = mode === 'grid';
      if (mode === 'dense') {
        this.listItemGap = 2;
        this.listItemHeight = 42;
      } else if (mode === 'compact' || mode === 'list') {
        this.listItemGap = 4;
        this.listItemHeight = 60;
      }
    },
    snapshotWorkspace() {
      const collectionStore = useCollectionStore();
      this.savedWorkspaceSnapshot = {
        assetFilters: JSON.parse(JSON.stringify(this.assetFilters)),
        collectionFilters: JSON.parse(JSON.stringify(this.collectionFilters)),
        resourceFilters: JSON.parse(JSON.stringify(this.resourceFilters)),
        showCollections: this.showCollections,
        showAssets: this.showAssets,
        onlyAssets: this.onlyAssets,
        showTasks: this.showTasks,
        showResources: this.showResources,
        showChildCollections: this.showChildCollections,
        showChildAssets: this.showChildAssets,
        showChildResources: this.showChildResources,
        showDependencies: this.showDependencies,
        useDeep: this.useDeep,
        hasAssignees: this.hasAssignees,
        noAssignees: this.noAssignees,
        workspaceSearchQuery: this.viewSearchQuery,
        viewMode: this.viewMode,
        collectionPath: collectionStore.navigatedCollection?.collection_path
          || collectionStore.navigatedCollection?.item_path
          || null,
      };
    },
    getCurrentWorkspaceState() {
      return {
        assetFilters: this.assetFilters,
        collectionFilters: this.collectionFilters,
        resourceFilters: this.resourceFilters,
        showCollections: this.showCollections,
        showAssets: this.showAssets,
        onlyAssets: this.onlyAssets,
        showTasks: this.showTasks,
        showResources: this.showResources,
        showChildCollections: this.showChildCollections,
        showChildAssets: this.showChildAssets,
        showChildResources: this.showChildResources,
        showDependencies: this.showDependencies,
        useDeep: this.useDeep,
        hasAssignees: this.hasAssignees,
        noAssignees: this.noAssignees,
      };
    },
    resetFilters() {
      (this.showCollections = true),
        (this.showAssets = true),
        (this.onlyAssets = false),
        (this.showTasks = true),
        (this.showResources = true),
        (this.showChildCollections = true),
        (this.showChildAssets = true),
        (this.showChildResources = true),
        (this.useDeep = false);
      this.hasAssignees = false;
      this.noAssignees = false;
      this.showDependencies = true;
      this.assetFilters = [];
      this.collectionFilters = [];
      this.resourceFilters = [];
      this.workspaceSearchQuery = "";
      this.viewSearchQuery = "";
      this.applyViewMode(defaultViewMode);
    },
    async setUntrackedVisibility(value) {
      this.showUntracked = value;
      defaultShowUntracked = value;
      await SettingsService.SetUntrackedVisibility(value);
    },
    setCompactView() {
      this.viewMode = 'compact';
      this.useGrid = false;
      this.listItemGap = 4;
      this.listItemHeight = 60;
      defaultViewMode = 'compact';
      SettingsService.SetDefaultViewMode('compact');
    },
    setGridView() {
      this.viewMode = 'grid';
      this.useGrid = true;
      defaultViewMode = 'grid';
      SettingsService.SetDefaultViewMode('grid');
    },
    setKanbanView() {
      if (this.viewMode !== 'kanban') {
        this.previousViewMode = this.viewMode;
      }
      this.viewMode = 'kanban';
      this.useGrid = false;
      defaultViewMode = 'kanban';
      SettingsService.SetDefaultViewMode('kanban');
    },
    restorePreviousView() {
      const target = this.previousViewMode && this.previousViewMode !== 'kanban'
        ? this.previousViewMode
        : 'dense';
      this.previousViewMode = null;
      switch (target) {
        case 'grid': this.setGridView(); break;
        case 'compact': this.setCompactView(); break;
        case 'list': this.setListView(); break;
        case 'dense':
        default: this.setDenseView(); break;
      }
    },
    setDenseView() {
      this.viewMode = 'dense';
      this.useGrid = false;
      this.listItemGap = 2;
      this.listItemHeight = 42;
      defaultViewMode = 'dense';
      SettingsService.SetDefaultViewMode('dense');
    },
    setListView() {
      this.viewMode = 'compact';
      this.useGrid = false;
      this.listItemGap = 4;
      this.listItemHeight = 60;
      defaultViewMode = 'compact';
      SettingsService.SetDefaultViewMode('compact');
    },
  },
});
