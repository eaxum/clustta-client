import { defineStore } from "pinia";

export const useDesktopModalStore = defineStore("desktopModals", {
  state: () => ({
    modalStates: {
      popUpModal: false,
      loginModal: false,

      ignoreConfigModal: false,
      directoryConfigModal: false,
      appInfoModal: false,
      eulaModal: false,
      dirOnboardModal: false,

      // browser
      addProjectModal: false,
      cloneProjectModal: false,
      editProjectModal: false,

      configWorkflowModal: false,
      selectWorkflowModal: false,

      createEntityModal: false,
      editEntityModal: false,

      createTaskModal: false,
      selectAppModal: false,
      editTaskModal: false,

      addWebLinkModal: false,

      createCheckpointModal: false,
      createMultipleCheckpointsModal: false,

      importItemsModal: false,
      timelineModal: false,
      addWorkspaceModal: false,

      // settings
      addProjectTemplateModal: false,
      editProjectTemplateModal: false,
      duplicateProjectTemplateModal: false,

      addTemplateModal: false,
      editTemplateModal: false,

      addAssetTypeModal: false,
      editTaskTypeModal: false,

      addCollectionTypeModal: false,
      editEntityTypeModal: false,

      manageCollaboratorModal: false,
      addCollaboratorModal: false,
      editCollaboratorModal: false,

      addRoleModal: false,
      editRoleModal: false,

      composeWorkflowModal: false,

      // user
      addUserTemplateModal: false,
      editUserTemplateModal: false,

      addUserAssetTypeModal: false,
      editUserAssetTypeModal: false,

      addUserCollectionTypeModal: false,
      editUserCollectionTypeModal: false,
    },

    activeModal: null,
    modalMaskVisible: false,
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
  },
});
