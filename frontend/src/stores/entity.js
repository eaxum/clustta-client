import { defineStore } from "pinia";
import { useCommonStore } from "@/stores/common";
import { useTaskStore } from "@/stores/task";
import utils from "@/services/utils";
import { EntityService } from "@/../bindings/clustta/services";
import { useProjectStore } from "./projects";

export const useEntityStore = defineStore("entity", {
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
    getEntityTypes: (state) => {
      return state.entityTypes;
    },
    getEntityTypesNames: (state) => {
      let entityTypes = state.entityTypes;
      let entityTypesNames = [];
      for (let i = 0; i < entityTypes.length; i++) {
        let entityType = entityTypes[i];
        entityTypesNames.push(entityType.name);
      }
      return entityTypesNames;
    },
    getEntities: (state) => {
      return state.entities;
    },
    getFilteredEntities: (state) => {
      const commonStore = useCommonStore();
      const taskStore = useTaskStore();

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
            // matched entity types
            const entityTypeMatch =
              selectedEntityTypes.length === 0 ||
              selectedEntityTypes.includes(
                entity.entity_type_name.toLowerCase()
              );

            return entityTypeMatch;
          })
          .filter(
            (item) =>
              item.file_path.toLowerCase().replace(/\\/g, '/').includes(viewSearchQuery) &&
              item.file_path.toLowerCase().replace(/\\/g, '/').includes(workspaceSearchQuery)
          );
        return utils.sortPathAlphabetically(filteredEntities, "entity");
      } else {
        filteredEntities = entities.filter(
          (item) =>
            item.file_path.toLowerCase().replace(/\\/g, '/').includes(viewSearchQuery) &&
            item.file_path.toLowerCase().replace(/\\/g, '/').includes(workspaceSearchQuery)
        );
        return utils.sortPathAlphabetically(filteredEntities, "entity");
      }
    },
    getEntitiesNames: (state) => {
      const sortedEntities = state.entities.sort((a, b) => {
        return a.name.localeCompare(b.name);
      });
      return sortedEntities.map((entity) => entity.name);
    },
  },
  actions: {
    filterEntities(entities){
      const commonStore = useCommonStore();
      const taskStore = useTaskStore();

      const viewSearchQuery = commonStore.viewSearchQuery;
      const workspaceSearchQuery = commonStore.workspaceSearchQuery;

      let filteredEntities;

      if (commonStore.entityFilters.length) {
        const selectedEntityTypes = commonStore.entityFilters
          .filter((filter) => filter.type === "entity-type")
          .map((filter) => filter.name.toLowerCase());

        filteredEntities = entities
          .filter((entity) => {
            // matched entity types
            const entityTypeMatch =
              selectedEntityTypes.length === 0 ||
              selectedEntityTypes.includes(
                entity.entity_type_name.toLowerCase()
              );

            return entityTypeMatch;
          })
          .filter(
            (item) =>
              item.file_path.toLowerCase().replace(/\\/g, '/').includes(viewSearchQuery) &&
              item.file_path.toLowerCase().replace(/\\/g, '/').includes(workspaceSearchQuery)
          );
        return utils.sortPathAlphabetically(filteredEntities, "entity");
      } else {
        filteredEntities = entities.filter(
          (item) =>
            item.file_path.toLowerCase().replace(/\\/g, '/').includes(viewSearchQuery) &&
            item.file_path.toLowerCase().replace(/\\/g, '/').includes(workspaceSearchQuery)
        );
        return utils.sortPathAlphabetically(filteredEntities, "entity");
      }
    },
    async reloadEntityTypes() {
      const projectStore = useProjectStore();
      let entityTypes = await EntityService.GetEntityTypes(
        projectStore.activeProject.uri
      );
      this.entityTypes = entityTypes;
    },
    async reloadEntities() {
      const projectStore = useProjectStore();
      console.time("entities_loading");
      let entities = await EntityService.GetEntities(
        projectStore.activeProject.uri
      );
      console.timeEnd("entities_loading");
      this.entities = [];

      for (let i = 0; i < entities.length; i++) {
        let entity = entities[i];
        let preview = null;
        if (entity.preview) {
          preview = "data:image/png;base64," + entity.preview;
        }
        entity.preview = preview;

        this.entities.push(entity);
      }
      this.rebuildEntitiesNameIndex();
      this.rebuildEntitiesIndex();
    },

    async rebuildEntitiesNameIndex() {
      let entityNameIndex = {};
      for (let i = 0; i < this.entities.length; i++) {
        entityNameIndex[this.entities[i].name] = i;
      }
      this.entityNameIndex = entityNameIndex;
    },
    async rebuildEntitiesIndex() {
      let entities_index = {};
      let entityChildrenIndex = {};
      for (let i = 0; i < this.entities.length; i++) {
        let parentId = this.entities[i].parent_id;
        let entityId = this.entities[i].id;
        entities_index[entityId] = i;
        if (!entityChildrenIndex[parentId]) {
          entityChildrenIndex[parentId] = [entityId];
        } else {
          entityChildrenIndex[parentId].push(entityId);
        }
      }
      this.entities_index = entities_index;
      this.entity_children_index = entityChildrenIndex;
    },
    findEntity(id) {
      let entityIndex = this.entities_index[id];
      return this.entities[entityIndex];
    },
    getEntityChildren(entityId, recursive = false) {
      let entityChildrenIds = this.entity_children_index[entityId] || [];
      let children = entityChildrenIds.map((childId) =>
        this.findEntity(childId)
      );

      if (recursive) {
        for (let i = 0; i < children.length; i++) {
          let child = children[i];
          let childChildren = this.getEntityChildren(child.id, true);
          children = children.concat(childChildren);
        }
      }

      return children;
    },
    getEntityByName(name) {
      const allEntities = this.entities;
      return allEntities.find((item) => item.name === name);
      // return this.entities.find((entity) => entity.name === name);
    },
    getEntity(id) {
      return this.entities.find((entity) => entity.id === id);
    },
    async markEntityAsDeleted(entityId) {
      let entity = this.entities.find((entity) => entity.id === entityId);
      ////console.log(entity);
      let entityIndex = this.entityNameIndex[entity.name];
      this.entities[entityIndex].trashed = true;
    },
    async unmarkEntityAsDeleted(entityId) {
      let entity = this.entities.find((entity) => entity.id === entityId);
      let entityIndex = this.entityNameIndex[entity.name];
      this.entities[entityIndex].trashed = false;
    },
    reloadSelectedEntity() {
      for (let entity of this.entities) {
        if (!entity.trashed) {
          this.selectedEntity = entity;
          break;
        }
      }
    },
    changeEntityParent(entity_id, parentId) {
      let entityIndex = this.entities_index[entity_id];
      this.entities[entityIndex].parent_id = parentId;
      this.rebuildEntitiesIndex();
    },
    selectEntity(entity) {
      this.selectedEntity = entity;
    },
    getEntityTypeIcon(entityTypeId) {
      let entityType = this.entityTypes.find(
        (entityType) => entityType.id === entityTypeId
      );
      return entityType.icon;
    },
    navigateToEntity(entity) {
      this.navigatedEntity = entity;
    },
  },
});
