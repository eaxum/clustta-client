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
      backUpProjectModal: false,
      importProjectModal: false,

      // browser
      addProjectModal: false,
      cloneProjectModal: false,
      editProjectModal: false,
      projectDetailsModal: false,
      uploadProjectModal: false,

      configWorkflowModal: false,
      selectWorkflowModal: false,

      createCollectionModal: false,
      editCollectionModal: false,

      createAssetModal: false,
      selectAppModal: false,
      editAssetModal: false,

      addWebLinkModal: false,

      createCheckpointModal: false,
      createMultipleCheckpointsModal: false,

      importItemsModal: false,
      addWorkspaceModal: false,

      // settings
      addProjectTemplateModal: false,
      editProjectTemplateModal: false,
      duplicateProjectTemplateModal: false,

      addTemplateModal: false,
      editTemplateModal: false,

      addAssetTypeModal: false,
      editAssetTypeModal: false,

      addCollectionTypeModal: false,
      editCollectionTypeModal: false,

      manageCollaboratorModal: false,
      addCollaboratorModal: false,

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

      //studio
      configSelfManagedStudioModal: false,
      configClusttaCloudStudioModal: false,
      selectNewStudioTypeModal: false,
      updateStudioModal: false,

      // sync
      syncConflictModal: false,

      // diagnostics
      submitDiagnosticsModal: false,
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
