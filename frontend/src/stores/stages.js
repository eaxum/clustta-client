import { defineStore } from "pinia";
import { useMenu } from "@/stores/menu";
import { usePaneStore } from "@/stores/panes";
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from "@/stores/collections";
import { useDndStore } from "@/stores/dnd";
import { useProjectStore } from "@/stores/projects";

export const useStageStore = defineStore("stages", {
  state: () => ({
    stages: {
      projects: false,
      browser: false,
      dependencies: false,
      trash: false,
      settings: false,
      projectSettings: false,
      studioSettings: false,
      account: false,
    },

    operationActive: false,
    groupItems: false,
    activeStage: null,
    selectedStage: null,
    modalMaskVisible: false,
    showTaskCheckboxes: false,
    showEntityCheckboxes: false,
    expandAllSubtasks: false,

    firstSelectedTaskId: "",
    lastSelectedTaskId: "",

    firstSelectedEntityId: "",
    lastSelectedEntityId: "",

    firstSelectedItemId: "",
    lastSelectedItemId: "",

    sidePaneActive: false,

    navigationBreadCrumbs: ["browser", "dependencies"],
    allTasks: [],
    expandedTask: null,
    expandedTasks: [],
    expandedEntity: null,
    expandedEntities: {},
    selectedItem: null,
    cutItems: [],
    copiedItems: [],
    allTasksCollapsed: true,
    visibleTasks: 0,
    visibleSubtasks: 0,
    visibleEntities: 0,
    entityDataIds: [],
    selectedTypes: "none",
    markedTasks: [],
    selectdProject: [],
    markedProjects: [],
    markedResources: [],
    markedEntities: [],
    markedItems: [],
    selectedItems: [],
    allTasksMarked: true,
    allSubtasksMarked: true,
    allEntitiesMarked: true,
    allResourcesMarked: true,
  }),
  getters: {
    typeTracker() {
      const counts = {
        entity: 0,
        task: 0,
        untracked_task: 0,
        untracked_entity: 0,
      };
      
      this.selectedItems.forEach(item => {
        if (item.type in counts) {
          counts[item.type]++;
        } else if (item.type === 'resource') {
          counts.task++; // Resources are counted as tasks
        }
      });
      
      return counts;
    }
  },
  actions: {
    setStageVisibility(stageName, value) {
      if (this.stages.hasOwnProperty(stageName)) {
        // Check if the modal is already active
        if (
          value &&
          this.activeStage !== null &&
          this.activeStage !== stageName
        ) {
          // Disable the currently active modal
          this.stages[this.activeStage] = false;
        }

        this.stages[stageName] = value;
        this.activeStage = value ? stageName : null;
        this.modalMaskVisible = this.isAnyModalActive();
      }
      this.selectedStage = stageName;
    },
    disableAllModals() {
      for (const stageName in this.stages) {
        this.stages[stageName] = false;
      }
      this.activeStage = null;
      this.modalMaskVisible = false;
    },
    isAnyModalActive() {
      return Object.values(this.stages).some((isVisible) => isVisible);
    },
    toggleTaskCheckboxVisibility() {
      this.showTaskCheckboxes = !this.showTaskCheckboxes;
    },
    toggleEntityCheckboxVisibility() {
      this.showEntityCheckboxes = !this.showEntityCheckboxes;
    },

    toggleSubtasks(id) {
      //check if active sequence's index is included and then add it if not
      if (this.allTasks.includes(id)) {
        this.allTasks = this.allTasks.filter((i) => i !== id);
      } else {
        this.allTasks.push(id);
      }

      //collapse/expand all sequences based on whether the allTasks is empty
      if (this.allTasks.length >= 0) {
        this.allTasksCollapsed = false;
      }
      if (this.allTasks.length == 0) {
        this.allTasksCollapsed = true;
      }
    },

    expandTask(taskId) {
      if (this.expandedTasks.includes(taskId)) {
        this.expandedTasks = this.expandedTasks.filter(
          (item) => item !== taskId
        );
      } else {
        this.expandedTasks.push(taskId);
      }
    },

    expandEntity(entity, untracked = false) {
      let entityId = entity.id;
      if (entityId in this.expandedEntities) {
        const childrenIds = Object.entries(this.expandedEntities)
          .filter(([key, value]) => value.entity_path.startsWith(entity.entity_path))
          .map(([key]) => key);
        let entitiesToClose = [entityId, ...childrenIds];

        const newExpandedEntities = { ...this.expandedEntities };
        for (const id of entitiesToClose) {
          delete newExpandedEntities[id];
        }
        this.expandedEntities = newExpandedEntities;
      } else {
        const assetStore = useAssetStore();
        // Initialize with 0 initially, the actual height will be set by onHeightChange
        this.expandedEntities = {
          ...this.expandedEntities,
          [entityId]: {
            height: 0,
            entity_path: entity.entity_path,
          },
        };
      }
    },

    handleClick(event, item, itemType, allItems) {
      const id = item.id;
      const untrackedTypes = ["folder", "file"];

      if (event.ctrlKey) {
        if (!this.markedItems.includes(id)) {
          this.markedItems.push(id);
          this.selectedItems.push(item);
          this.lastSelectedItemId = id;
        } else {
          if (this.lastSelectedItemId === id) {
            this.lastSelectedItemId = "";
            const filteredItemIds = this.markedItems.filter(
              (item) => item !== id
            );
            this.markedItems = filteredItemIds;
            this.selectedItems = this.selectedItems.filter(
              (selectedItem) => selectedItem.id !== id
            );
            console.log(item.type)
          } else {
            this.lastSelectedItemId = id;
          }
        }
      } else if (event.shiftKey) {
        this.lastSelectedItemId = id;
        const firstIndex = allItems.findIndex(
          (i) => i.id === this.firstSelectedItemId
        );
        const lastIndex = allItems.findIndex(
          (i) => i.id === this.lastSelectedItemId
        );

        if (firstIndex === -1) {
          this.firstSelectedItemId = id;
          this.lastSelectedItemId = "";
          this.markedItems = [id];
          this.selectedItems = [item];
          if (!untrackedTypes.includes(itemType)) {
            this.selectItem(item, itemType, true);
          }
        }

        const start = Math.min(firstIndex, lastIndex);
        const end = Math.max(firstIndex, lastIndex);

        const selectedRange = allItems.slice(start, end + 1);

        this.markedItems = selectedRange.map((i) => i.id);
        this.selectedItems = [...selectedRange];
      } else {
        this.lastSelectedItemId = "";
        if (this.firstSelectedItemId !== id) {
          this.firstSelectedItemId = id;
          this.markedItems = [id];
          this.selectedItems = [item];
          this.selectItem(item, itemType, true);
        } else if (this.markedItems.length && this.firstSelectedItemId === id) {
          this.markedItems = [id];
          this.selectedItems = [item];
          this.selectItem(item, itemType, true);
        } else {
          this.markedItems = [];
          this.selectedItems = [];
          this.firstSelectedItemId = "";
        }
      }

    },

    handlePreviewClick(event, item, allItems) {
      const dndStore = useDndStore();
      const key = this.pluralize(item.type);

      let itemType;
      const id = item.id;

      if (item.entity_type_id) {
        itemType = "entity";
      } else {
        if (item.is_resource) {
          itemType = "resource";
        } else {
          itemType = "task";
        }
      }

      if (event.ctrlKey) {
        if (this.markedItems.length === 1 && this.markedItems.includes(id)) {
          return;
        }
        this.selectPreviewItem(item);
      } else if (event.shiftKey) {
        this.lastSelectedItemId = id;
        const firstIndex = allItems.findIndex(
          (i) => i.id === this.firstSelectedItemId
        );
        const lastIndex = allItems.findIndex(
          (i) => i.id === this.lastSelectedItemId
        );

        if (firstIndex === -1) {
          this.firstSelectedItemId = id;
          this.lastSelectedItemId = "";
          this.markedItems = [item.id];

          Object.keys(dndStore.previewDataSelectedItems).forEach((key) => {
            delete dndStore.previewDataSelectedItems[key];
          });

          this.markedItems = [item.id];
          dndStore.previewDataSelectedItems[key] = [item];
          return;
        }

        const start = Math.min(firstIndex, lastIndex);
        const end = Math.max(firstIndex, lastIndex);

        const selectedRange = allItems.slice(start, end + 1);
        this.markedItems = selectedRange.map((i) => i.id);

        for (const rangeItem of selectedRange) {
          rangeItem.type = rangeItem.entity_type_id ? "entity" : "task";

          const entities = selectedRange.filter(
            (item) => item.type === "entity"
          );
          const tasks = selectedRange.filter((item) => item.type === "task");

          dndStore.previewDataSelectedItems["entities"] = entities;
          dndStore.previewDataSelectedItems["tasks"] = tasks;
        }
      } else {
        const key = this.pluralize(item.type);

        Object.keys(dndStore.previewDataSelectedItems).forEach((key) => {
          delete dndStore.previewDataSelectedItems[key];
        });

        this.firstSelectedItemId = id;
        this.markedItems = [id];
        dndStore.previewDataSelectedItems[key] = [item];
        return;
      }
    },

    pluralize(word) {
      const pluralRules = {
        entity: "entities",
        task: "tasks",
        resource: "resources",
      };

      return pluralRules[word] || `${word}s`;
    },

    selectPreviewItem(item) {
      const dndStore = useDndStore();
      const key = this.pluralize(item.type);
      dndStore.previewDataActiveItem = item;

      if (!dndStore.previewDataSelectedItems) {
        dndStore.previewDataSelectedItems = {};
      }

      if (!dndStore.previewDataSelectedItems[key]) {
        dndStore.previewDataSelectedItems[key] = [];
      }

      const itemId = item.id;
      const selectedItems = dndStore.previewDataSelectedItems[key];

      const existingIndex = selectedItems.findIndex(
        (selectedItem) => selectedItem.id === itemId
      );

      if (existingIndex === -1) {
        this.markedItems.push(itemId);
        dndStore.previewDataSelectedItems[key].push(item);
      } else {
        this.markedItems = this.markedItems.filter((id) => id !== itemId);
        dndStore.previewDataSelectedItems[key].splice(existingIndex, 1);
      }
    },

    selectItem(item, itemType, solo = false) {
      const panes = usePaneStore();
      const assetStore = useAssetStore();
      const collectionStore = useCollectionStore();
      const projectStore = useProjectStore();

      if (solo) {
        this.deselectAllItems();
      }

      if (itemType === "entity") {
        collectionStore.selectCollection(item);
        this.selectedItem = item;
        // panes.setPaneVisibility("collectionDetails", true);
      } else if (itemType === "task") {
        assetStore.selectAsset(item);
        this.selectedItem = item;
        // panes.setPaneVisibility("assetDetails", true);
      } else if (itemType === "resource") {
        assetStore.selectAsset(item);
        this.selectedItem = item;
        // panes.setPaneVisibility("assetDetails", true);
      } else {
        projectStore.selectUntrackedItem(item);
        this.selectedItem = item;
        // console.log(item)
        // panes.setPaneVisibility("untrackedItemDetails", true);
      }
    },

    deselectAllItems() {
      const assetStore = useAssetStore();
      const collectionStore = useCollectionStore();
      const projectStore = useProjectStore();

      assetStore.selectedTask = null;
      collectionStore.selectedEntity = null;
      projectStore.selectedUntrackedItem = null;
    },

    checkIntersections() {
      const dndStore = useDndStore();
      const visibleItemIds = Object.keys(dndStore.visibleItemRefs);
      const noItems = visibleItemIds.length <= 0;
      if (noItems) {
        dndStore.intersectingItemIds = [];
        return;
      } else {
        dndStore.intersectingItemIds = dndStore.intersectingItemIds.filter(
          (item) => visibleItemIds.includes(item)
        );
      }

      const intersectorEl = document.querySelector(".intersector");
      if (!intersectorEl) return;

      const intersectorRect = intersectorEl.getBoundingClientRect();

      Object.entries(dndStore.visibleItemRefs).forEach(([id, item]) => {
        if (item) {
          const itemRect = item.getBoundingClientRect();
          const isWithinBounds =
            itemRect.top >= intersectorRect.top &&
            itemRect.bottom <= intersectorRect.bottom;
          if (isWithinBounds) {
            if (!dndStore.intersectingItemIds.includes(id)) {
              dndStore.intersectingItemIds.push(id);
            }
          } else {
            dndStore.intersectingItemIds = dndStore.intersectingItemIds.filter(
              (item) => item !== id
            );
          }
        }
      });
    },

    debounce(func, wait) {
      let timeout;
      return function executedFunction(...args) {
        const later = () => {
          clearTimeout(timeout);
          func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
      };
    },
  },
});