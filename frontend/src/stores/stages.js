import { defineStore } from "pinia";
import { useMenu } from "@/stores/menu";
import { usePaneStore } from "@/stores/panes";
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from "@/stores/collections";
import { useCommonStore } from "@/stores/common";
import { useDndStore } from "@/stores/dnd";
import { useNotificationStore } from "@/stores/notifications";
import { useProjectStore } from "@/stores/projects";
import { AppService, AssetService, CollectionService, FSService } from '@/services';

export const useStageStore = defineStore("stages", {
  state: () => ({
    os: '',
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
    cmdOrCtrlKey(event) {
      // Use cmd key on macOS, ctrl key on other platforms
      return this.os === 'darwin' ? event.metaKey : event.ctrlKey;
    },
    
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

      if (this.cmdOrCtrlKey(event)) {
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

      if (this.cmdOrCtrlKey(event)) {
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
      } else if (itemType === "task") {
        assetStore.selectAsset(item);
        this.selectedItem = item;
      } else if (itemType === "resource") {
        assetStore.selectAsset(item);
        this.selectedItem = item;
      } else {
        projectStore.selectUntrackedItem(item);
        this.selectedItem = item;
      }
    },

    deselectAllItems() {
      const assetStore = useAssetStore();
      const collectionStore = useCollectionStore();
      const projectStore = useProjectStore();

      assetStore.selectedAsset = null;
      collectionStore.selectedCollection = null;
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

    // Pastes cut or copied items to the specified target collection.
    // If targetCollectionId is not provided, uses the current navigated collection or project root.
    async pasteItems(targetCollectionId = null, targetDirectory = null) {
      console.log('here')
      const collectionStore = useCollectionStore();
      const commonStore = useCommonStore();
      const notificationStore = useNotificationStore();
      const projectStore = useProjectStore();

      const hasCutItems = this.cutItems.length > 0;
      const hasCopiedItems = this.copiedItems.length > 0;
      if (!(hasCutItems || hasCopiedItems)) return { success: false, needsRefresh: false };

      this.operationActive = true;
      let needsRefresh = false;

      // Determine target location
      const finalTargetCollectionId = targetCollectionId !== null
        ? targetCollectionId
        : (commonStore.navigatorMode && collectionStore.navigatedCollection
          ? collectionStore.navigatedCollection.id
          : '');
          console.log(targetCollectionId)
      const finalTargetDirectory = targetDirectory !== null
        ? targetDirectory
        : (commonStore.navigatorMode && collectionStore.navigatedCollection
          ? collectionStore.navigatedCollection.file_path
          : projectStore.activeProject.working_directory);

      try {
        await FSService.MakeDirs(finalTargetDirectory);

        if (hasCutItems) {
          // CUT: Move items to target location
          const entityIdsToMove = [];
          const taskIdsToMove = [];
          const renameOperations = [];

          for (const item of this.cutItems) {
            if (item.type === 'entity') {
              entityIdsToMove.push(item.id);
            } else if (item.type === 'task') {
              taskIdsToMove.push(item.id);
            } else if (item.type === 'untracked_task' || item.type === 'untracked_entity') {
              const extension = item.type === 'untracked_task' ? item.extension : '';
              const fullName = item.name + extension;
              const newPath = await FSService.JoinPath(finalTargetDirectory, fullName);
              renameOperations.push({ oldPath: item.file_path, newPath });
            }
          }

          if (entityIdsToMove.length) {
            try {
              await CollectionService.ChangeCollectionParent(projectStore.activeProject.uri, entityIdsToMove, finalTargetCollectionId);
              notificationStore.addNotification('Moved successfully.', '', 'success');
              needsRefresh = true;
            } catch (error) {
              notificationStore.errorNotification('Error changing entity parent', error);
            }
          }
          if (taskIdsToMove.length) {
            try {
              await AssetService.ChangeAssetCollection(projectStore.activeProject.uri, taskIdsToMove, finalTargetCollectionId);
              notificationStore.addNotification('Moved successfully.', '', 'success');
              needsRefresh = true;
            } catch (error) {
              notificationStore.errorNotification('Error moving assets', error);
            }
          }
          if (renameOperations.length) {
            try {
              await FSService.RenameBatch(renameOperations);
              needsRefresh = true;
            } catch (error) {
              notificationStore.errorNotification('Error moving files', error);
            }
          }

          this.cutItems = [];
        } else if (hasCopiedItems) {
          // COPY: Duplicate items to target location
          let successCount = 0;
          let failureCount = 0;

          for (const item of this.copiedItems) {
            try {
              if (item.type === 'task') {
                // Duplicate tracked task to target collection
                const duplicatedTask = await AssetService.DuplicateAsset(
                  projectStore.activeProject.uri,
                  item.id,
                  finalTargetCollectionId
                );
                // Copy the physical file if it exists
                if (item.file_path && duplicatedTask.file_path) {
                  const sourceExists = await FSService.Exists(item.file_path);
                  if (sourceExists) {
                    await FSService.DuplicateFile(item.file_path, duplicatedTask.file_path);
                  }
                }
                successCount++;
              } else if (item.type === 'untracked_task') {
                // Copy untracked file
                const fullName = item.name + (item.extension || '');
                const destinationPath = await this.generateUniqueDestinationPath(finalTargetDirectory, fullName);
                await FSService.DuplicateFile(item.file_path, destinationPath);
                successCount++;
              } else if (item.type === 'untracked_entity') {
                // Copy untracked folder
                const destinationPath = await this.generateUniqueDestinationPath(finalTargetDirectory, item.name);
                await FSService.DuplicateFolder(item.file_path, destinationPath);
                successCount++;
              }
            } catch (error) {
              console.error('Error copying item:', item.name, error);
              failureCount++;
            }
          }

          if (successCount > 0) {
            needsRefresh = true;
            notificationStore.addNotification(
              successCount === 1 ? '1 item pasted' : `${successCount} items pasted`,
              '',
              'success'
            );
          }
          if (failureCount > 0) {
            notificationStore.errorNotification(
              failureCount === 1 ? '1 item failed to paste' : `${failureCount} items failed to paste`,
              ''
            );
          }

          this.copiedItems = [];
        }

        return { success: true, needsRefresh };
      } catch (error) {
        notificationStore.errorNotification('Error pasting items', error.message || error);
        return { success: false, needsRefresh };
      } finally {
        this.operationActive = false;
      }
    },

    // Generates a unique file path by appending a counter if the file already exists.
    async generateUniqueDestinationPath(directory, fileName) {
      const originalPath = await FSService.JoinPath(directory, fileName);
      const exists = await FSService.Exists(originalPath);
      if (!exists) return originalPath;
      const baseName = fileName.includes('.') ? fileName.substring(0, fileName.lastIndexOf('.')) : fileName;
      const extension = fileName.includes('.') ? fileName.substring(fileName.lastIndexOf('.')) : '';
      let counter = 1;
      let newPath;
      do {
        const newFileName = `${baseName} (${counter})${extension}`;
        newPath = await FSService.JoinPath(directory, newFileName);
        const pathExists = await FSService.Exists(newPath);
        if (!pathExists) return newPath;
        counter++;
      } while (counter < 100);
      const timestamp = Date.now();
      return await FSService.JoinPath(directory, `${baseName}_${timestamp}${extension}`);
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