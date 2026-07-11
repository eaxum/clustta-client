import { defineStore } from "pinia";

import { CollectionService } from "@/services";

import { useCommonStore } from "@/stores/common";
import { useAssetStore } from "@/stores/assets";
import { useProjectStore } from "./projects";

import utils from "@/services/utils";

const normalizeSearchValue = (value) => String(value || "").toLowerCase();

const matchesCollectionSearch = (collection, viewSearchQuery, workspaceSearchQuery) => {
  const name = normalizeSearchValue(collection.name);
  const collectionPath = normalizeSearchValue(collection.collection_path);
  const viewSearch = normalizeSearchValue(viewSearchQuery);
  const workspaceSearch = normalizeSearchValue(workspaceSearchQuery);

  const viewMatch = !viewSearch || name.includes(viewSearch) || collectionPath.includes(viewSearch);
  const workspaceMatch = !workspaceSearch || name.includes(workspaceSearch) || collectionPath.includes(workspaceSearch);

  return viewMatch && workspaceMatch;
};

const getCollectionAssigneeIds = (collection) => {
  return Array.isArray(collection.assignee_ids) ? collection.assignee_ids : [];
};

const matchesCollectionAssigneeFilters = (collection, commonStore, selectedAssigneeIds) => {
  const assigneeIds = getCollectionAssigneeIds(collection);

  if (commonStore.hasAssignees) return assigneeIds.length > 0;
  if (commonStore.noAssignees) return assigneeIds.length === 0;
  if (!selectedAssigneeIds.size) return true;

  return assigneeIds.some((assigneeId) => selectedAssigneeIds.has(String(assigneeId)));
};

const matchesCollectionSharedFilters = (collection, selectedSharedValues) => {
  return !selectedSharedValues.size || selectedSharedValues.has(!!collection.is_shared);
};

const matchesCollectionTypeFilters = (collection, collectionTypes, selectedTypeIds, selectedTypeNames) => {
  if (!selectedTypeIds.size && !selectedTypeNames.size) return true;

  if (selectedTypeIds.has(String(collection.collection_type_id))) return true;

  const collectionType = collectionTypes.find((item) => item.id === collection.collection_type_id);
  const collectionTypeName = normalizeSearchValue(collection.collection_type || collectionType?.name);

  return selectedTypeNames.has(collectionTypeName);
};

const filterCollectionsByCommonFilters = (collections, commonStore, collectionTypes) => {
  const selectedTypeFilters = commonStore.collectionFilters.filter((filter) => filter.type === "collection-type");
  const selectedTypeIds = new Set(selectedTypeFilters.map((filter) => String(filter.id)));
  const selectedTypeNames = new Set(selectedTypeFilters.map((filter) => normalizeSearchValue(filter.name)));
  const selectedAssigneeIds = new Set(
    commonStore.assetFilters
      .filter((filter) => filter.type === "assignation")
      .map((filter) => String(filter.id))
  );
  const selectedSharedValues = new Set(
    commonStore.collectionFilters
      .filter((filter) => filter.type === "shared")
      .map((filter) => !!filter.value)
  );

  return (collections || []).filter((collection) => {
    return matchesCollectionTypeFilters(collection, collectionTypes, selectedTypeIds, selectedTypeNames)
      && matchesCollectionAssigneeFilters(collection, commonStore, selectedAssigneeIds)
      && matchesCollectionSharedFilters(collection, selectedSharedValues)
      && matchesCollectionSearch(collection, commonStore.viewSearchQuery, commonStore.workspaceSearchQuery);
  });
};

export const useCollectionStore = defineStore("collection", {
  state: () => ({
    collections: [],
    collections_index: {},
    collection_children_index: {},
    collectionTypes: [],
    collectionNameIndex: {},
    selectedCollection: null,
    selectedCollectionType: null,
    navigatedCollection: null,
    collectionStateFlags: {
      has_untracked: false,
      has_modified: false,
      has_outdated: false,
      has_fetchable: false
    },
    loadingCollectionStates: false,
  }),
  getters: {
    getCollectionTypes: (state) => {
      return state.collectionTypes;
    },
    getCollectionTypesNames: (state) => {
      let collectionTypes = state.collectionTypes;
      let collectionTypesNames = [];
      for (let i = 0; i < collectionTypes.length; i++) {
        let collectionType = collectionTypes[i];
        collectionTypesNames.push(collectionType.name);
      }
      return collectionTypesNames;
    },
    getCollections: (state) => {
      return state.collections;
    },
    getFilteredCollections: (state) => {
      const commonStore = useCommonStore();

      const collections = state.collections;
      const filteredCollections = filterCollectionsByCommonFilters(collections, commonStore, state.collectionTypes);

      const sortedCollections = utils.sortPathAlphabetically(
        filteredCollections,
        "collection"
      );
      return sortedCollections;
    },

    getDisplayedCollections: (state) => {
      const commonStore = useCommonStore();
      let filteredCollections = state.getFilteredCollections;
      let displayedCollections = [];

      for (let i = 0; i < filteredCollections.length; i++) {
        let collection = filteredCollections[i];
        if (collection.trashed === false) {
          displayedCollections.push(collection);
        }
      }

      return displayedCollections;
    },

    getDisplayedCollectionsNames: (state) => {
      const sortedCollections = state.getDisplayedCollections.slice().sort((a, b) => {
        return a.name.localeCompare(b.name);
      });
      return sortedCollections.map((collection) => collection.name);
    },
  },
  actions: {
    filterCollections(collections){
      const commonStore = useCommonStore();
      const filteredCollections = filterCollectionsByCommonFilters(collections, commonStore, this.collectionTypes);

      const sortedCollections = utils.sortPathAlphabetically(
        filteredCollections,
        "collection"
      );
      return sortedCollections;
    },
    async markCollectionAsDeleted(collectionId) {
      let collectionIndex = this.collections_index[collectionId];
      this.collections[collectionIndex].trashed = true;
    },
    
    async unmarkCollectionAsDeleted(collectionId) {
      let collectionIndex = this.collections_index[collectionId];
      this.collections[collectionIndex].trashed = false;
    },

    async reloadCollectionTypes() {
      const projectStore = useProjectStore();
      if (!projectStore.activeProject?.uri) return;
      let collectionTypes = await CollectionService.GetCollectionTypes(
        projectStore.activeProject.uri
      );
      this.collectionTypes = collectionTypes.map(type => ({
        ...type,
        type: 'collection-type',
      }));
    },

    async reloadCollections() {
      const projectStore = useProjectStore();
      let collections = await CollectionService.GetCollections(
        projectStore.activeProject.uri
      );
      this.collections = collections;
      await this.rebuildCollectionsIndex();
    },

    async rebuildCollectionsIndex() {
      let collectionIndex = {};
      let collectionChildrenIndex = {};
      let collectionNameIndex = {};
      for (let i = 0; i < this.collections.length; i++) {
        let collectionId = this.collections[i].id;
        let parentId = this.collections[i].parent_id;
        collectionIndex[collectionId] = i;
        collectionNameIndex[this.collections[i].name] = this.collections[i];
        if (!collectionChildrenIndex[parentId]) {
          collectionChildrenIndex[parentId] = [collectionId];
        } else {
          collectionChildrenIndex[parentId].push(collectionId);
        }
      }
      this.collections_index = collectionIndex;
      this.collection_children_index = collectionChildrenIndex;
      this.collectionNameIndex = collectionNameIndex;
    },

    findCollection(id) {
      let collectionIndex = this.collections_index[id];
      return this.collections[collectionIndex];
    },

    findCollectionByName(name) {
      return this.collectionNameIndex[name];
    },

    selectCollection(collection) {
      this.selectedCollection = collection;
    },

    getChildCollections(collectionId, recursive = false) {
      let childCollectionIds = this.collection_children_index[collectionId] || [];
      let collections = childCollectionIds.map((collectionId) => this.findCollection(collectionId));

      if (recursive) {
        for (let childId of childCollectionIds) {
          collections = collections.concat(this.getChildCollections(childId, true));
        }
      }

      return collections;
    },

    getCollectionTypeIcon(collectionTypeId) {
      let collectionTypeIcon = "";
      for (let i = 0; i < this.collectionTypes.length; i++) {
        let type = this.collectionTypes[i];
        if (type.id === collectionTypeId) {
          collectionTypeIcon = type.icon;
          break;
        }
      }
      return collectionTypeIcon;
    },
    navigateToCollection(collection) {
      this.navigatedCollection = collection;
    },

    /**
     * Loads modified and untracked items for checkpoint creation.
     * 
     * Supports three modes:
     * - Tracked collection: Pass collectionId to scan tracked collection and its hierarchy
     * - Untracked path: Pass targetPath to scan a filesystem location
     * - Root: Pass neither to scan entire project
     * 
     * Updates assetStore.modifiedAssets with checkpoint candidates.
     */
    async reloadItemsForCheckpoint(collectionId = null, targetPath = null) {
      const assetStore = useAssetStore();
      assetStore.loadingAssetStates = true;
      const projectStore = useProjectStore();
      let project = projectStore.activeProject;
      
      collectionId = collectionId || "";
      let scanPath = targetPath || "";
      
      try {
        const items = await CollectionService.GetItemsForCheckpoint(
          project.uri,
          collectionId,
          scanPath,
          project.working_directory,
          project.ignore_list
        );
        const modifiedAssets = items.modified_assets.map(asset => ({
          asset_id: asset.id,
          type: 'asset',
          name: asset.name,
          icon: asset.icon,
          asset_path: asset.asset_path,
          assignee_id: asset.assignee_id,
          collection_id: asset.collection_id,
          can_modify: true,
          extension: asset.extension,
          display_path: asset.asset_path + asset.extension
        }));

        const untrackedItems = items.untracked_files.map(file => ({
          ...file,
          display_path: file.asset_path || file.item_path
        }));

        assetStore.modifiedAssets = {
          modified: modifiedAssets,
          untracked: untrackedItems
        };
      } catch (error) {
        console.error('Error loading items for checkpoint:', error);
        assetStore.modifiedAssets = {
          modified: [],
          untracked: []
        };
      } finally {
        assetStore.loadingAssetStates = false;
      }
    },

    /**
     * Fetches outdated assets recursively for a collection.
     * @param {string|null} collectionId - Collection ID to scan (null for root)
     * @returns {Promise<Array>} Array of outdated asset objects
     */
    async getOutdatedItems(collectionId = null) {
      const assetStore = useAssetStore();
      assetStore.loadingAssetStates = true;
      const projectStore = useProjectStore();
      let project = projectStore.activeProject;
      
      collectionId = collectionId || "";
      
      try {
        const result = await CollectionService.GetOutdatedItemsInCollection(
          project.uri,
          collectionId,
          project.working_directory,
          project.ignore_list
        );

        return result.outdated_assets || [];
      } catch (error) {
        console.error('Error loading outdated items:', error);
        return [];
      } finally {
        assetStore.loadingAssetStates = false;
      }
    },

    /**
     * Loads optimized state flags (untracked/modified/outdated/fetchable) for current collection context.
     * Updates collectionStateFlags with boolean flags indicating presence of items in each state.
     */
    async loadCollectionStateFlags() {
      this.loadingCollectionStates = true;
      const projectStore = useProjectStore();
      const commonStore = useCommonStore();
      
      try {
        const project = projectStore.activeProject;
        if (!project) return;

        let collectionId = "root";
        
        if (commonStore.navigatorMode && this.navigatedCollection) {
          collectionId = this.navigatedCollection.id || "root";
        }

        if (this.navigatedCollection) {
          if (this.navigatedCollection?.type !== 'collection') {
            return;
          }
        }

        const flags = await CollectionService.GetCollectionStateFlags(
          project.uri,
          collectionId,
          project.working_directory,
          project.ignore_list
        );
        
        this.collectionStateFlags = flags;
      } catch (error) {
        console.error('Error loading collection state flags:', error);
        this.collectionStateFlags = {
          has_untracked: false,
          has_modified: false,
          has_outdated: false,
          has_fetchable: false
        };
      } finally {
        this.loadingCollectionStates = false;
      }
    },
  },
});
