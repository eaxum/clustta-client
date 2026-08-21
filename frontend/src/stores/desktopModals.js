import { defineStore } from "pinia";

export const useDesktopModalStore = defineStore("desktopModals", {
  state: () => ({
    modalStates: {
      popUpModal: false,
      shareModal: false,
      confirmDangerousActionModal: false,
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
      exportModal: false,
      addWorkspaceModal: false,
      addDependencyPresetModal: false,
      saveIgnorePresetModal: false,

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

      addTagModal: false,
      editTagModal: false,

      manageCollaboratorModal: false,
      addCollaboratorModal: false,

      addRoleModal: false,
      duplicateRoleModal: false,
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
      configClusttaCloudStudioModal: false,
      updateStudioModal: false,

      // sync
      syncConflictModal: false,

      // plans
      clusttaCloudModal: false,

      // integrations
      integrationAuthModal: false,
      integrationLinkModal: false,
      integrationSyncModal: false,
      directoryMappingModal: false,
      assetTypeMappingModal: false,
      statusMappingModal: false,
      studioIntegrationConfigModal: false,

      // squash
      squashModal: false,

      // agent
      configAgentModal: false,
      consoleModal: false,
      agentApprovalModal: false,

      // dependency
      dependencyGraphModal: false,

      // image viewer
      imageViewerModal: false,

      // diagnostics
      submitDiagnosticsModal: false,
    },

    activeModal: null,
    modalMaskVisible: false,

    // Image viewer payload (thumbnail src + title + source file) for imageViewerModal.
    imageViewer: { src: "", title: "", filePath: "", extension: "" },
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

    // Opens the image viewer modal with a fallback thumbnail src, title, and source file details.
    openImageViewer(src, title = "", filePath = "", extension = "") {
      this.imageViewer = { src, title, filePath, extension };
      this.setModalVisibility("imageViewerModal", true);
    },
  },
});
