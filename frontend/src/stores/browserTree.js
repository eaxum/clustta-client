import { defineStore } from 'pinia';
import {
  browserRootParentKey,
  getBrowserItemKey,
  reconcileBrowserItems,
} from '@/lib/browserTree';

export const useBrowserTreeStore = defineStore('browserTree', {
  state: () => ({
    projectUri: '',
    itemsByKey: {},
    childKeysByParent: {},
    loadedParents: {},
    loadingParents: {},
    parentErrors: {},
    refreshVersions: {},
  }),

  getters: {
    getItem: (state) => (itemKey) => state.itemsByKey[itemKey] || null,
    getChildren: (state) => (parentKey) => {
      const childKeys = state.childKeysByParent[parentKey] || [];
      return childKeys
        .map((itemKey) => state.itemsByKey[itemKey])
        .filter(Boolean);
    },
    rootItems() {
      return this.getChildren(browserRootParentKey);
    },
  },

  actions: {
    setProject(projectUri) {
      const nextProjectUri = projectUri || '';
      if (this.projectUri === nextProjectUri) return;

      this.reset();
      this.projectUri = nextProjectUri;
    },

    replaceRootItems(projectUri, incomingItems = []) {
      this.setProject(projectUri);
      return this.replaceChildren(browserRootParentKey, incomingItems);
    },

    replaceChildren(parentKey, incomingItems = []) {
      if (!parentKey) return [];

      const currentItems = this.getChildren(parentKey);
      const reconciledItems = reconcileBrowserItems(currentItems, incomingItems);
      const childKeys = [];

      for (const item of reconciledItems) {
        const itemKey = getBrowserItemKey(item);
        if (!itemKey) continue;

        this.itemsByKey[itemKey] = item;
        childKeys.push(itemKey);
      }

      this.childKeysByParent[parentKey] = childKeys;
      this.loadedParents[parentKey] = true;
      this.loadingParents[parentKey] = false;
      delete this.parentErrors[parentKey];

      return reconciledItems;
    },

    patchItem(itemKey, updates = {}) {
      const item = this.itemsByKey[itemKey];
      if (!item) return null;

      Object.assign(item, updates);
      return item;
    },

    beginParentRefresh(parentKey) {
      const refreshVersion = (this.refreshVersions[parentKey] || 0) + 1;
      this.refreshVersions[parentKey] = refreshVersion;
      this.loadingParents[parentKey] = true;
      delete this.parentErrors[parentKey];
      return refreshVersion;
    },

    isCurrentRefresh(parentKey, refreshVersion) {
      return this.refreshVersions[parentKey] === refreshVersion;
    },

    replaceChildrenIfCurrent(parentKey, refreshVersion, incomingItems = []) {
      if (!this.isCurrentRefresh(parentKey, refreshVersion)) return null;
      return this.replaceChildren(parentKey, incomingItems);
    },

    failParentRefresh(parentKey, refreshVersion, error) {
      if (!this.isCurrentRefresh(parentKey, refreshVersion)) return;

      this.loadingParents[parentKey] = false;
      this.parentErrors[parentKey] = error;
    },

    reset() {
      this.projectUri = '';
      this.itemsByKey = {};
      this.childKeysByParent = {};
      this.loadedParents = {};
      this.loadingParents = {};
      this.parentErrors = {};
      this.refreshVersions = {};
    },
  },
});
