import { defineStore } from "pinia";
import router from "@/router";

export const useModalStore = defineStore("modals", {
  state: () => ({
    modalContainer: null,
    modalIcon: "",
    modalMaskVisible: false,
    modalStates: {
      tagsFilterMenu: false,
      assetTypeFilterMenu: false,
      statusFilterMenu: false,
      stateFilterMenu: false,
      extensionFilterMenu: false,

      appInfoModal: false,
      assetMenu: false,
      collectionMenu: false,
      resourceItemMenu: false,
      workspaceMenu: false,
      createMenu: false,
      checkPointMenu: false,
      projectMenu: false,
      popUpModal: false,
      loginModal: false,
      createCheckpointModal: false,
      createMultipleCheckpointsModal: false,
      createTaskModal: false,
      editTaskModal: false,
      editResourceModal: false,
      importResourcesModal: false,
      editWorkspaceModal: false,
      addProjectModal: false,
      addFileLinkModal: false,
      addWebLinkModal: false,
      addTemplateModal: false,
      importFileModal: false,
      trackFileModal: false,
      editTemplateModal: false,
      manageCollaboratorModal: false,
      editCollaboratorModal: false,
      assignTaskModal: false,
      testModal: false,
    },
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
    triggerMenuItem(menu, page) {
      this.setModalVisibility(menu, false);
      this.showTraySearch = false;
      router.push({ name: page });
    },
  },
});
