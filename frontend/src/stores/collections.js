import { defineStore } from "pinia";
import { useCommonStore } from "@/stores/common";
import { useAssetStore } from "@/stores/assets";
import utils from "@/services/utils";
import { EntityService } from "@/../bindings/clustta/services";
import { useProjectStore } from "./projects";

export const useCollectionStore = defineStore("collection", {
  state: () => ({
    entities: [],
    entities_index: {},
    entity_children_index: {},
    entityTypes: [],
    entityNameIndex: {},
    selectedEntity: null,
    selectedEntityType: null,
    navigatedEntity: null,
  }),
  getters: {
    getCollectionTypes: (state) => {
      return state.entityTypes;
    },
    getCollectionTypesNames: (state) => {
      let collectionTypes = state.entityTypes;
      let collectionTypesNames = [];
      for (let i = 0; i < collectionTypes.length; i++) {
        let collectionType = collectionTypes[i];
        collectionTypesNames.push(collectionType.name);
      }
      return collectionTypesNames;
    },
    getCollections: (state) => {
      return state.entities;
    },
    getFilteredCollections: (state) => {
      const commonStore = useCommonStore();
      const assetStore = useAssetStore();

      const collections = state.entities;
      const viewSearchQuery = commonStore.viewSearchQuery;
      const workspaceSearchQuery = commonStore.workspaceSearchQuery;

      let filteredCollections;

      if (commonStore.entityFilters.length) {
        const selectedCollectionTypes = commonStore.entityFilters
          .filter((filter) => filter.type === "entity-type")
          .map((filter) => filter.name.toLowerCase());

        filteredCollections = collections
          .filter((collection) => {
            const collectionTypeMatch =
              selectedCollectionTypes.length === 0 ||
              selectedCollectionTypes.includes(collection.entity_type.toLowerCase());

            return collectionTypeMatch;
          })
          .filter(
            (item) =>
              item.entity_path
                .toLowerCase()
                .includes(viewSearchQuery) &&
              item.entity_path
                .toLowerCase()
                .includes(workspaceSearchQuery)
          );
      } else {
        filteredCollections = collections.filter(
          (item) =>
            item.entity_path
              .toLowerCase()
              .includes(viewSearchQuery) &&
            item.entity_path
              .toLowerCase()
              .includes(workspaceSearchQuery)
        );
      }

      const sortedCollections = utils.sortPathAlphabetically(
        filteredCollections,
        "entity"
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
      const assetStore = useAssetStore();

      const viewSearchQuery = commonStore.viewSearchQuery;
      const workspaceSearchQuery = commonStore.workspaceSearchQuery;

      let filteredCollections;

      if (commonStore.entityFilters.length) {
        const selectedCollectionTypes = commonStore.entityFilters
          .filter((filter) => filter.type === "entity-type")
          .map((filter) => filter.name.toLowerCase());

        filteredCollections = collections
          .filter((collection) => {
            const collectionTypeMatch =
              selectedCollectionTypes.length === 0 ||
              selectedCollectionTypes.includes(collection.entity_type.toLowerCase());

            return collectionTypeMatch;
          })
          .filter(
            (item) => {
              const nameMatch = !viewSearchQuery || item.name?.toLowerCase().includes(viewSearchQuery.toLowerCase());
              const entityPathMatch = !viewSearchQuery || item.entity_path?.toLowerCase().includes(viewSearchQuery.toLowerCase());
              const workspaceNameMatch = !workspaceSearchQuery || item.name?.toLowerCase().includes(workspaceSearchQuery.toLowerCase());
              const workspaceEntityPathMatch = !workspaceSearchQuery || item.entity_path?.toLowerCase().includes(workspaceSearchQuery.toLowerCase());
              
              return (nameMatch || entityPathMatch) && (workspaceNameMatch || workspaceEntityPathMatch);
            }
          );
      } else {
        filteredCollections = collections.filter(
          (item) => {
            const nameMatch = !viewSearchQuery || item.name?.toLowerCase().includes(viewSearchQuery.toLowerCase());
            const entityPathMatch = !viewSearchQuery || item.entity_path?.toLowerCase().includes(viewSearchQuery.toLowerCase());
            const workspaceNameMatch = !workspaceSearchQuery || item.name?.toLowerCase().includes(workspaceSearchQuery.toLowerCase());
            const workspaceEntityPathMatch = !workspaceSearchQuery || item.entity_path?.toLowerCase().includes(workspaceSearchQuery.toLowerCase());
            
            return (nameMatch || entityPathMatch) && (workspaceNameMatch || workspaceEntityPathMatch);
          }
        );
      }

      const sortedCollections = utils.sortPathAlphabetically(
        filteredCollections,
        "entity"
      );
      return sortedCollections;
    },
    async markCollectionAsDeleted(collectionId) {
      let collectionIndex = this.entities_index[collectionId];
      this.entities[collectionIndex].trashed = true;
    },
    
    async unmarkCollectionAsDeleted(collectionId) {
      let collectionIndex = this.entities_index[collectionId];
      this.entities[collectionIndex].trashed = false;
    },

    // @deprecated - This method is not currently used
    async markMultipleCollectionsAsDeleted(collectionIds) {
      for (const collectionId of collectionIds) {
        let collectionIndex = this.entities_index[collectionId];
        if (collectionIndex !== undefined) {
          this.entities[collectionIndex].trashed = true;
        }
      }
    },

    // @deprecated - This method is not currently used
    async unmarkMultipleCollectionsAsDeleted(collectionIds) {
      for (const collectionId of collectionIds) {
        let collectionIndex = this.entities_index[collectionId];
        if (collectionIndex !== undefined) {
          this.entities[collectionIndex].trashed = false;
        }
      }
    },

    async reloadCollectionTypes() {
      const projectStore = useProjectStore();
      let collectionTypes = await EntityService.GetEntityTypes(
        projectStore.activeProject.uri
      );
      this.entityTypes = collectionTypes.map(type => ({
        ...type,
        type: 'entity-type',
      }));
    },

    async reloadCollections() {
      const projectStore = useProjectStore();
      let collections = await EntityService.GetEntities(
        projectStore.activeProject.uri
      );
      this.entities = collections;
      await this.rebuildCollectionsIndex();
    },

    async rebuildCollectionsIndex() {
      let collectionIndex = {};
      let collectionChildrenIndex = {};
      let collectionNameIndex = {};
      for (let i = 0; i < this.entities.length; i++) {
        let collectionId = this.entities[i].id;
        let parentId = this.entities[i].parent_id;
        collectionIndex[collectionId] = i;
        collectionNameIndex[this.entities[i].name] = this.entities[i];
        if (!collectionChildrenIndex[parentId]) {
          collectionChildrenIndex[parentId] = [collectionId];
        } else {
          collectionChildrenIndex[parentId].push(collectionId);
        }
      }
      this.entities_index = collectionIndex;
      this.entity_children_index = collectionChildrenIndex;
      this.entityNameIndex = collectionNameIndex;
    },

    findCollection(id) {
      let collectionIndex = this.entities_index[id];
      return this.entities[collectionIndex];
    },

    findCollectionByName(name) {
      return this.entityNameIndex[name];
    },

    selectCollection(collection) {
      this.selectedEntity = collection;
    },

    getChildCollections(collectionId, recursive = false) {
      let childCollectionIds = this.entity_children_index[collectionId] || [];
      let collections = childCollectionIds.map((collectionId) => this.findCollection(collectionId));

      if (recursive) {
        for (let childId of childCollectionIds) {
          collections = collections.concat(this.getChildCollections(childId, true));
        }
      }

      return collections;
    },

    // @deprecated - This method is not currently used
    getAllCollectionChildren(collectionId) {
      let allChildren = {};
      let directChildren = this.entity_children_index[collectionId] || [];
      
      for (let childId of directChildren) {
        allChildren[childId] = this.getAllCollectionChildren(childId);
      }
      
      return allChildren;
    },

    getCollectionTypeIcon(collectionTypeId) {
      let collectionTypeIcon = "";
      for (let i = 0; i < this.entityTypes.length; i++) {
        let type = this.entityTypes[i];
        if (type.id === collectionTypeId) {
          collectionTypeIcon = type.icon;
          break;
        }
      }
      return collectionTypeIcon;
    },
    navigateToCollection(collection) {
      this.navigatedEntity = collection;
    },
  },
});