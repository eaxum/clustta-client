import { defineStore } from "pinia";

export const useCommonStore = defineStore("common", {
  state: () => ({
    activeFilters: [],
    resourceFilters: [],
    taskFilters: [],
    entityFilters: [],
    hasAssignees: false,
    noAssignees: false,
    reloadFilters: false,
    showFullPath: false,
    hideExtensions: true,
    showThumbs: false,
    showUntracked: true,
    showEntities: true,
    showTasks: true,
    onlyAssets: false,
    showResources: true,
    showChildEntities: true,
    showChildTasks: true,
    showChildResources: true,
    showDependencies: true,
    useDeep: false,
    navigatorMode: false,
    useGrid: false,
    viewMode: 'compact',
    gridSize: 100,
    listItemHeight: 60,
    listItemGap: 4,
    viewSearchQuery: "",
    workspaceSearchQuery: "",
    fileStates: [
      { name: "normal", type: "state", icon: "circle-check" },
      { name: "modified", type: "state", icon: "layers-plus-alert" },
      { name: "outdated", type: "state", icon: "circle-check-alert" },
      { name: "rebuildable", type: "state", icon: "jigsaw" },
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
      { name: "All Tasks/Resources", active: false, icon: "brush" },
      { name: "Templates", active: false, icon: "file" },
    ],
    workspaces: [],
    projectWorkflows: [],
    activeWorkspace: "Default",
    ghostCardStyle: {
      leaving: false,
      pos: { x: 0, y: 0 },
      width: 0,
      cursorDistance: { x: 0, y: 0 },
      transform: "",
    },
  }),
  getters: {
    getEntities: (state) => {
      return state.entities;
    },
    getTasks: (state) => {
      return state.tasks;
    },
    getResources: (state) => {
      return state.resources;
    },
  },
  actions: {
    setActiveWorkspace(workspace) {
      this.activeWorkspace = workspace.name;
      this.taskFilters = workspace.filters.taskFilters;
      this.entityFilters = workspace.filters.entityFilters;
      this.resourceFilters = workspace.filters.resourceFilters;

      this.showEntities = workspace.filters.showEntities;
      this.showTasks = workspace.filters.showTasks;
      this.onlyAssets = workspace.filters.onlyAssets;
      this.showResources = workspace.filters.showResources;
      this.showChildEntities = workspace.filters.showChildEntities;
      this.showChildTasks = workspace.filters.showChildTasks;
      this.showChildResources = workspace.filters.showChildResources;
      this.showDependencies = workspace.filters.showDependencies;
      this.useDeep = workspace.filters.useDeep;

      this.hasAssignees = workspace.filters.hasAssignees;
      this.noAssignees = workspace.filters.noAssignees;

      this.workspaceSearchQuery = workspace.workspaceSearchQuery;
    },
    resetFilters() {
      (this.showEntities = true),
        (this.showTasks = true),
        (this.onlyAssets = false),
        (this.showResources = true),
        (this.showChildEntities = true),
        (this.showChildTasks = true),
        (this.showChildResources = true),
        (this.useDeep = false);
      this.hasAssignees = false;
      this.noAssignees = false;
      this.showDependencies = true;
      this.taskFilters = [];
      this.entityFilters = [];
      this.resourceFilters = [];
      this.workspaceSearchQuery = "";
      this.viewSearchQuery = "";
    },
    setCompactView() {
      this.viewMode = 'compact';
      this.useGrid = false;
      this.listItemGap = 4;
      this.listItemHeight = 60;
    },
    setLargerView() {
      this.viewMode = 'larger';
      this.useGrid = false;
      this.listItemGap = 10;
      this.listItemHeight = 70;
    },
    setGridView() {
      this.viewMode = 'grid';
      this.useGrid = true;
      // Keep existing grid settings
    },
  },
});
