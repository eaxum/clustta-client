import { defineStore } from "pinia";

export const usePaneStore = defineStore("panes", {
  state: () => ({
    detailPanes: {
      dependencies: false,
      collaborators: false,
      projectDetails: false,
      collectionDetails: false,
      assetDetails: false,
      untrackedItemDetails: false,
      resourceDetails: false,
      checkpoints: false,
      projectCheckpoints: false,
    },

    activeModal: null,
    modalMaskVisible: false,
    enabledPanes: ['browser', 'projects'],
    showDetailsPane: false,
  }),
  getters: {},
  actions: {
    setPaneVisibility(paneName, value) {
      if (this.detailPanes.hasOwnProperty(paneName)) {
        // Check if the modal is already active
        if (
          value &&
          this.activeModal !== null &&
          this.activeModal !== paneName
        ) {
          // Disable the currently active modal
          this.detailPanes[this.activeModal] = false;
        }

        this.detailPanes[paneName] = value;
        this.activeModal = value ? paneName : null;
        this.modalMaskVisible = this.isAnyModalActive();
      }
    },
    disableAllModals() {
      for (const paneName in this.detailPanes) {
        this.detailPanes[paneName] = false;
      }
      this.activeModal = null;
      this.modalMaskVisible = false;
    },
    isAnyModalActive() {
      return Object.values(this.detailPanes).some((isVisible) => isVisible);
    },
  },
});
