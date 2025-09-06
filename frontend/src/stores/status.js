import { defineStore } from "pinia";
import { StatusService } from "@/../bindings/clustta/services";
import { useProjectStore } from "./projects";

export const useStatusStore = defineStore("status", {
  state: () => ({
    statuses: [],
  }),
  getters: {
    getStatuses: (state) => {
      return state.statuses;
    },
  },
  actions: {
    async reloadStatuses() {
      const projectStore = useProjectStore();
      this.statuses = await StatusService.GetStatuses(
        projectStore.activeProject.uri
      );
    },
  },
});
