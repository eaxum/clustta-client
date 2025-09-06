import { defineStore } from "pinia";

export const useSettingsStore = defineStore("settings", {
  state: () => ({
    modalStates: {
      general: false,
      templates: false,
      collaborators: false,
      workflows: false,
      roles: false,
      collectiontypes: false,
      assettypes: false,
      
      ignorelist: false,
      usertemplates: false,
      projecttemplates: false,
      
      studio: false,
      studiocollaborators: false,
    },

    activeModal: null,
    activeModalName: "",
    selectedStage: null,
    modalMaskVisible: false,
    showTaskCheckboxes: false,
    expandAllSubtasks: false,
    firstSelectedTaskId: "",
    lastSelectedTaskId: "",

    sidePaneActive: false,

    allTasks: [],
    allTasksCollapsed: true,

    markedTasks: [],
    allTasksMarked: true,

    settingsItems: [
      { name: "General", icon: "monitor" },
      { name: "Templates", icon: "file" },
      { name: "Collaborators", icon: "person" },
      { name: "Roles", icon: "scale" },
      
      { name: "Asset types", icon: "brush" },
      { name: "Collection types", icon: "folder" },
      { name: "Ignore List", icon: "file-watch" },
      { name: "Project Templates", icon: "briefcase" },
      { name: "Workflows", icon: "workflow-arrow" },
      
      { name: "Studio", icon: "storefront" },
      { name: "Studio Collaborators", icon: "person" },
    ],

    templateContexts: [
      { name: "Templates", icon: "file" },
      { name: "Asset types", icon: "brush" },
      { name: "Collection types", icon: "folder" },
      { name: "Ignore List", icon: "file-watch" },
    ],
  }),
  getters: {},
  actions: {
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
    toggleTaskCheckboxVisibility() {
      this.showTaskCheckboxes = !this.showTaskCheckboxes;
    },

    toggleSubtasks(id) {
      //check if active sequence's index is included and then add it if not
      if (this.allTasks.includes(id)) {
        this.allTasks = this.allTasks.filter((i) => i !== id);
      } else {
        this.allTasks.push(id);
      }

      //collapse/expand all sequences based on whether the allTasks is empty
      if (this.allTasks.length >= 0) {
        this.allTasksCollapsed = false;
      }
      if (this.allTasks.length == 0) {
        this.allTasksCollapsed = true;
      }
    },

    toggleSubtaskVisibility(fullData) {
      //ifthere are open sequences collapse all
      if (this.allTasks.length != 0) {
        this.allTasksCollapsed = !this.allTasksCollapsed;
        this.allTasks = [];
      } else {
        //if there are none, expand all
        const idArray = fullData.map((task) => task.id);
        this.allTasksCollapsed = !this.allTasksCollapsed;
        this.allTasks = idArray;
      }
    },

    markTask(id) {
      //check if active sequence's id is included and then add it if not
      if (this.markedTasks.includes(id)) {
        this.markedTasks = this.markedTasks.filter((i) => i !== id);
        console.log(this.markedTasks);
      } else {
        this.markedTasks.push(id);
        console.log(this.markedTasks);
      }

      //collapse/expand all sequences based on whether the markedTasks is empty

      if (this.markedTasks.length >= 2) {
        this.allTasksMarked = false;
      } else {
        this.allTasksMarked = true;
        console.log(this.allTasksMarked);
      }
    },
  },
});
