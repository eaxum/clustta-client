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
    getTagNames: (state) => {
      return state.tags.map((tag) => tag.name);
    },
  },
  actions: {
    async reloadTags() {
      const projectStore = useProjectStore();
      if (!projectStore.activeProject?.uri) return;
      const tags = await TagService.GetTags(projectStore.activeProject.uri);
      this.tags = tags.map((tag) => ({ ...tag, type: "tags" }));
    },

    // Adds a tag to an asset and returns the updated tag names.
    async addTagToAsset(assetId, tagName) {
      const projectStore = useProjectStore();
      if (!projectStore.activeProject?.uri) return [];
      const tags = await TagService.AddTagToAsset(projectStore.activeProject.uri, assetId, tagName);
      await this.reloadTags();
      return tags.map((tag) => tag.name);
    },

    // Removes a tag from an asset and returns the updated tag names.
    async removeTagFromAsset(assetId, tagId) {
      const projectStore = useProjectStore();
      if (!projectStore.activeProject?.uri) return [];
      const tags = await TagService.RemoveTagFromAsset(projectStore.activeProject.uri, assetId, tagId);
      return tags.map((tag) => tag.name);
    },
  },
});
