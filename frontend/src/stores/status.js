import { defineStore } from "pinia";
import { StatusService } from "@/services";
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
      if (!projectStore.activeProject?.uri) return;
      this.statuses = await StatusService.GetStatuses(
        projectStore.activeProject.uri
      );
    },
  },
});
