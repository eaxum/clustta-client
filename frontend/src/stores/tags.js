import { defineStore } from "pinia";
import { TagService } from "@/services";
import { useProjectStore } from "./projects";

export const useTagStore = defineStore("tags", {
  state: () => ({
    tags: [],
  }),
  getters: {
    getTags: (state) => {
      return state.tags;
    },
  },
  actions: {
    async reloadTags() {
      const projectStore = useProjectStore();
      if (!projectStore.activeProject?.uri) return;
      this.tags = await TagService.GetTags(projectStore.activeProject.uri);
    },
  },
});
