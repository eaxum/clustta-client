import { defineStore } from "pinia";
import { TagService } from "@/services";
import { useProjectStore } from "./projects";
import emitter from "@/lib/mitt";

export const useTagStore = defineStore("tags", {
  state: () => ({
    tags: [],
    selectedTag: null,
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

    async createTag(name) {
      const projectStore = useProjectStore();
      if (!projectStore.activeProject?.uri) return null;
      const tag = await TagService.CreateTag(projectStore.activeProject.uri, name);
      await this.reloadTags();
      return tag;
    },

    async updateTag(tagId, name, mergeOnCollision = false) {
      const projectStore = useProjectStore();
      if (!projectStore.activeProject?.uri) return null;
      const tag = await TagService.UpdateTag(projectStore.activeProject.uri, tagId, name, mergeOnCollision);
      await this.reloadTags();
      emitter.emit("refresh-browser");
      return tag;
    },

    async deleteTag(tagId) {
      const projectStore = useProjectStore();
      if (!projectStore.activeProject?.uri) return;
      await TagService.DeleteTag(projectStore.activeProject.uri, tagId);
      await this.reloadTags();
      emitter.emit("refresh-browser");
    },

    async getTagUsageCount(tagId) {
      const projectStore = useProjectStore();
      if (!projectStore.activeProject?.uri) return 0;
      return TagService.GetTagUsageCount(projectStore.activeProject.uri, tagId);
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

    // Adds a tag (by name) to multiple assets in a single transaction.
    async addTagToMultipleAssets(assetIds, tagName) {
      const projectStore = useProjectStore();
      if (!projectStore.activeProject?.uri || !assetIds.length) return;
      await TagService.AddTagToAssets(projectStore.activeProject.uri, assetIds, tagName);
      await this.reloadTags();
    },

    // Removes a tag (by id) from multiple assets in a single transaction.
    async removeTagFromMultipleAssets(assetIds, tagId) {
      const projectStore = useProjectStore();
      if (!projectStore.activeProject?.uri || !assetIds.length) return;
      await TagService.RemoveTagFromAssets(projectStore.activeProject.uri, assetIds, tagId);
    },
  },
});
