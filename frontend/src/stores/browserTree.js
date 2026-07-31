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

    beginRootRefresh(projectUri) {
      this.setProject(projectUri);
      return this.beginParentRefresh(browserRootParentKey);
    },

    isCurrentRootRefresh(refreshVersion) {
      return this.isCurrentRefresh(browserRootParentKey, refreshVersion);
    },

    replaceRootItemsIfCurrent(projectUri, refreshVersion, incomingItems = []) {
      if (this.projectUri !== projectUri) return null;
      return this.replaceChildrenIfCurrent(
        browserRootParentKey,
        refreshVersion,
        incomingItems
      );
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

    patchItemsById(itemId, updates = {}) {
      if (!itemId) return 0;

      let updatedItemCount = 0;
      for (const item of Object.values(this.itemsByKey)) {
        if (item.id !== itemId) continue;

        Object.assign(item, updates);
        updatedItemCount++;
      }
      return updatedItemCount;
    },

    markAssetsAvailable(assetIds = []) {
      let updatedItemCount = 0;
      for (const assetId of new Set(assetIds || [])) {
        updatedItemCount += this.patchItemsById(assetId, {
          file_status: 'normal',
        });
      }
      return updatedItemCount;
    },

    markCollectionFetched(collectionId, assetIds = []) {
      const updatedItemCount = this.markAssetsAvailable(assetIds);
      const collection = this.itemsByKey[`collection:${collectionId}`];
      if (collection) {
        collection.collectionStateFlags = {
          ...(collection.collectionStateFlags || {}),
          has_fetchable: false,
        };
      }
      return updatedItemCount;
    },

    markCollectionTreeFetched(collectionId, assetIds = []) {
      const updatedItemCount = this.markAssetsAvailable(assetIds);
      const targetCollection = collectionId
        ? this.itemsByKey[`collection:${collectionId}`]
        : null;
      const targetPath = targetCollection?.collection_path || '';

      for (const [itemKey, item] of Object.entries(this.itemsByKey)) {
        if (!itemKey.startsWith('collection:')) continue;

        const isRootFetch = !collectionId || collectionId === 'root';
        const isTarget = item.id === collectionId;
        const itemPath = item.collection_path || '';
        const isDescendant = targetPath
          && itemPath.startsWith(`${targetPath}/`);
        if (!isRootFetch && !isTarget && !isDescendant) continue;

        item.collectionStateFlags = {
          ...(item.collectionStateFlags || {}),
          has_fetchable: false,
        };
      }

      return updatedItemCount;
    },

    applyItemUpdates(eventData) {
      const itemUpdates = Array.isArray(eventData) ? eventData : [eventData];
      let updatedItemCount = 0;

      for (const itemUpdate of itemUpdates) {
        if (!itemUpdate?.itemId) continue;

        const updates = {};
        if (itemUpdate.property && itemUpdate.value !== undefined) {
          updates[itemUpdate.property] = itemUpdate.value;
        }
        if (Array.isArray(itemUpdate.updates)) {
          for (const update of itemUpdate.updates) {
            if (update?.property && update.value !== undefined) {
              updates[update.property] = update.value;
            }
          }
        }
        updatedItemCount += this.patchItemsById(itemUpdate.itemId, updates);
      }

      return updatedItemCount;
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
