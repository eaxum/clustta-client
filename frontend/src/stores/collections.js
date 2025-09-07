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
      let entityTypes = state.entityTypes;
      let entityTypesNames = [];
      for (let i = 0; i < entityTypes.length; i++) {
        let entityType = entityTypes[i];
        entityTypesNames.push(entityType.name);
      }
      return entityTypesNames;
    },
    getCollections: (state) => {
      return state.entities;
    },
    getFilteredCollections: (state) => {
      const commonStore = useCommonStore();
      const assetStore = useAssetStore();

      const entities = state.entities;
      const viewSearchQuery = commonStore.viewSearchQuery;
      const workspaceSearchQuery = commonStore.workspaceSearchQuery;

      let filteredEntities;

      if (commonStore.entityFilters.length) {
        const selectedEntityTypes = commonStore.entityFilters
          .filter((filter) => filter.type === "entity-type")
          .map((filter) => filter.name.toLowerCase());

        filteredEntities = entities
          .filter((entity) => {
            const entityTypeMatch =
              selectedEntityTypes.length === 0 ||
              selectedEntityTypes.includes(entity.entity_type.toLowerCase());

            return entityTypeMatch;
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
        filteredEntities = entities.filter(
          (item) =>
            item.entity_path
              .toLowerCase()
              .includes(viewSearchQuery) &&
            item.entity_path
              .toLowerCase()
              .includes(workspaceSearchQuery)
        );
      }

      const sortedEntities = utils.sortPathAlphabetically(
        filteredEntities,
        "entity"
      );
      return sortedEntities;
    },

    getDisplayedCollections: (state) => {
      const commonStore = useCommonStore();
      let filteredEntities = state.getFilteredCollections;
      let displayedEntities = [];

      for (let i = 0; i < filteredEntities.length; i++) {
        let entity = filteredEntities[i];
        if (entity.trashed === false) {
          displayedEntities.push(entity);
        }
      }

      return displayedEntities;
    },

    getDisplayedCollectionsNames: (state) => {
      const sortedEntities = state.getDisplayedCollections.slice().sort((a, b) => {
        return a.name.localeCompare(b.name);
      });
      return sortedEntities.map((entity) => entity.name);
    },
  },
  actions: {
    filterCollections(entities){
      const commonStore = useCommonStore();
      const assetStore = useAssetStore();

      const viewSearchQuery = commonStore.viewSearchQuery;
      const workspaceSearchQuery = commonStore.workspaceSearchQuery;

      let filteredEntities;

      if (commonStore.entityFilters.length) {
        const selectedEntityTypes = commonStore.entityFilters
          .filter((filter) => filter.type === "entity-type")
          .map((filter) => filter.name.toLowerCase());

        filteredEntities = entities
          .filter((entity) => {
            const entityTypeMatch =
              selectedEntityTypes.length === 0 ||
              selectedEntityTypes.includes(entity.entity_type.toLowerCase());

            return entityTypeMatch;
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
        filteredEntities = entities.filter(
          (item) => {
            const nameMatch = !viewSearchQuery || item.name?.toLowerCase().includes(viewSearchQuery.toLowerCase());
            const entityPathMatch = !viewSearchQuery || item.entity_path?.toLowerCase().includes(viewSearchQuery.toLowerCase());
            const workspaceNameMatch = !workspaceSearchQuery || item.name?.toLowerCase().includes(workspaceSearchQuery.toLowerCase());
            const workspaceEntityPathMatch = !workspaceSearchQuery || item.entity_path?.toLowerCase().includes(workspaceSearchQuery.toLowerCase());
            
            return (nameMatch || entityPathMatch) && (workspaceNameMatch || workspaceEntityPathMatch);
          }
        );
      }

      const sortedEntities = utils.sortPathAlphabetically(
        filteredEntities,
        "entity"
      );
      return sortedEntities;
    },
    async markCollectionAsDeleted(entityId) {
      let entityIndex = this.entities_index[entityId];
      this.entities[entityIndex].trashed = true;
    },
    
    async unmarkCollectionAsDeleted(entityId) {
      let entityIndex = this.entities_index[entityId];
      this.entities[entityIndex].trashed = false;
    },

    // @deprecated - This method is not currently used
    async markMultipleCollectionsAsDeleted(entityIds) {
      for (const entityId of entityIds) {
        let entityIndex = this.entities_index[entityId];
        if (entityIndex !== undefined) {
          this.entities[entityIndex].trashed = true;
        }
      }
    },

    // @deprecated - This method is not currently used
    async unmarkMultipleCollectionsAsDeleted(entityIds) {
      for (const entityId of entityIds) {
        let entityIndex = this.entities_index[entityId];
        if (entityIndex !== undefined) {
          this.entities[entityIndex].trashed = false;
        }
      }
    },

    async reloadCollectionTypes() {
      const projectStore = useProjectStore();
      let entityTypes = await EntityService.GetEntityTypes(
        projectStore.activeProject.uri
      );
      this.entityTypes = entityTypes.map(type => ({
        ...type,
        type: 'entity-type',
      }));
    },

    async reloadCollections() {
      const projectStore = useProjectStore();
      let entities = await EntityService.GetEntities(
        projectStore.activeProject.uri
      );
      this.entities = entities;
      await this.rebuildCollectionsIndex();
    },

    async rebuildCollectionsIndex() {
      let entityIndex = {};
      let entityChildrenIndex = {};
      let entityNameIndex = {};
      for (let i = 0; i < this.entities.length; i++) {
        let entityId = this.entities[i].id;
        let parentId = this.entities[i].parent_id;
        entityIndex[entityId] = i;
        entityNameIndex[this.entities[i].name] = this.entities[i];
        if (!entityChildrenIndex[parentId]) {
          entityChildrenIndex[parentId] = [entityId];
        } else {
          entityChildrenIndex[parentId].push(entityId);
        }
      }
      this.entities_index = entityIndex;
      this.entity_children_index = entityChildrenIndex;
      this.entityNameIndex = entityNameIndex;
    },

    findCollection(id) {
      let entityIndex = this.entities_index[id];
      return this.entities[entityIndex];
    },

    findCollectionByName(name) {
      return this.entityNameIndex[name];
    },

    selectCollection(entity) {
      this.selectedEntity = entity;
    },

    getChildCollections(entityId, recursive = false) {
      let childEntityIds = this.entity_children_index[entityId] || [];
      let entities = childEntityIds.map((entityId) => this.findCollection(entityId));

      if (recursive) {
        for (let childId of childEntityIds) {
          entities = entities.concat(this.getChildCollections(childId, true));
        }
      }

      return entities;
    },

    // @deprecated - This method is not currently used
    getAllCollectionChildren(entityId) {
      let allChildren = {};
      let directChildren = this.entity_children_index[entityId] || [];
      
      for (let childId of directChildren) {
        allChildren[childId] = this.getAllCollectionChildren(childId);
      }
      
      return allChildren;
    },

    getCollectionTypeIcon(entityTypeId) {
      let entityTypeIcon = "";
      for (let i = 0; i < this.entityTypes.length; i++) {
        let type = this.entityTypes[i];
        if (type.id === entityTypeId) {
          entityTypeIcon = type.icon;
          break;
        }
      }
      return entityTypeIcon;
    },
    navigateToCollection(entity) {
      this.navigatedEntity = entity;
    },
  },
});