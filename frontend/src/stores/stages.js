import { defineStore } from "pinia";
import { useMenu } from "@/stores/menu";
import { usePaneStore } from "@/stores/panes";
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from "@/stores/collections";
import { useCommonStore } from "@/stores/common";
import { useDndStore } from "@/stores/dnd";
import { useNotificationStore } from "@/stores/notifications";
import { useProjectStore } from "@/stores/projects";
import { usePlatformStore } from "@/stores/platform";
import { AppService, AssetService, CollectionService, FSService } from '@/services';

export const useStageStore = defineStore("stages", {
  state: () => ({
    stages: {
      projects: false,
      browser: false,
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
    showAssetCheckboxes: false,
    showCollectionCheckboxes: false,
    expandAllSubassets: false,

    firstSelectedAssetId: "",
    lastSelectedAssetId: "",

    firstSelectedCollectionId: "",
    lastSelectedCollectionId: "",

    firstSelectedItemId: "",
    lastSelectedItemId: "",

    sidePaneActive: false,

    navigationBreadCrumbs: ["browser"],
    allAssets: [],
    expandedAsset: null,
    expandedAssets: [],
    expandedCollection: null,
    expandedCollections: {},
    selectedItem: null,
    cutItems: [],
    copiedItems: [],
    allAssetsCollapsed: true,
    visibleAssets: 0,
    visibleSubassets: 0,
    visibleCollections: 0,
    collectionDataIds: [],
    selectedTypes: "none",
    markedAssets: [],
    selectdProject: [],
    markedProjects: [],
    markedResources: [],
    markedCollections: [],
    markedItems: [],
    selectedItems: [],
    allAssetsMarked: true,
    allSubassetsMarked: true,
    allCollectionsMarked: true,
    allResourcesMarked: true,
  }),
  getters: {
    typeTracker() {
      const counts = {
        collection: 0,
        asset: 0,
        untracked_asset: 0,
        untracked_collection: 0,
      };
      
      this.selectedItems.forEach(item => {
        if (item.type in counts) {
          counts[item.type]++;
        } else if (item.type === 'resource') {
          counts.asset++; // Resources are counted as assets
        }
      });
      
      return counts;
    }
  },
  actions: {
    cmdOrCtrlKey(event) {
      // Use cmd key on macOS, ctrl key on other platforms
      return usePlatformStore().isMac ? event.metaKey : event.ctrlKey;
    },

    // Returns true when the press is a context-menu gesture (right button,
    // or Ctrl+Click on macOS) rather than a primary selection/drag press.
    isContextMenuClick(event) {
      if (event.button === 2) return true;
      return usePlatformStore().isMac && event.button === 0 && event.ctrlKey;
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
    toggleAssetCheckboxVisibility() {
      this.showAssetCheckboxes = !this.showAssetCheckboxes;
    },
    toggleCollectionCheckboxVisibility() {
      this.showCollectionCheckboxes = !this.showCollectionCheckboxes;
    },

    toggleSubassets(id) {
      //check if active sequence's index is included and then add it if not
      if (this.allAssets.includes(id)) {
        this.allAssets = this.allAssets.filter((i) => i !== id);
      } else {
        this.allAssets.push(id);
      }

      //collapse/expand all sequences based on whether the allAssets is empty
      if (this.allAssets.length >= 0) {
        this.allAssetsCollapsed = false;
      }
      if (this.allAssets.length == 0) {
        this.allAssetsCollapsed = true;
      }
    },

    expandAsset(assetId) {
      if (this.expandedAssets.includes(assetId)) {
        this.expandedAssets = this.expandedAssets.filter(
          (item) => item !== assetId
        );
      } else {
        this.expandedAssets.push(assetId);
      }
    },

    expandCollection(collection, untracked = false) {
      let collectionId = collection.id;
      if (collectionId in this.expandedCollections) {
        const childrenIds = Object.entries(this.expandedCollections)
          .filter(([key, value]) => value.collection_path.startsWith(collection.collection_path))
          .map(([key]) => key);
        let collectionsToClose = [collectionId, ...childrenIds];

        const newExpandedCollections = { ...this.expandedCollections };
        for (const id of collectionsToClose) {
          delete newExpandedCollections[id];
        }
        this.expandedCollections = newExpandedCollections;
      } else {
        const assetStore = useAssetStore();
        // Initialize with 0 initially, the actual height will be set by onHeightChange
        this.expandedCollections = {
          ...this.expandedCollections,
          [collectionId]: {
            height: 0,
            collection_path: collection.collection_path,
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

      if (item.collection_type_id) {
        itemType = "collection";
      } else {
        if (item.is_resource) {
          itemType = "resource";
        } else {
          itemType = "asset";
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
          rangeItem.type = rangeItem.collection_type_id ? "collection" : "asset";

          const collections = selectedRange.filter(
            (item) => item.type === "collection"
          );
          const assets = selectedRange.filter((item) => item.type === "asset");

          dndStore.previewDataSelectedItems["collections"] = collections;
          dndStore.previewDataSelectedItems["assets"] = assets;
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
        collection: "collections",
        asset: "assets",
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

      if (itemType === "collection") {
        collectionStore.selectCollection(item);
        this.selectedItem = item;
      } else if (itemType === "asset") {
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
          const collectionIdsToMove = [];
          const assetIdsToMove = [];
          const renameOperations = [];

          for (const item of this.cutItems) {
            if (item.type === 'collection') {
              collectionIdsToMove.push(item.id);
            } else if (item.type === 'asset') {
              assetIdsToMove.push(item.id);
            } else if (item.type === 'untracked_asset' || item.type === 'untracked_collection') {
              const extension = item.type === 'untracked_asset' ? item.extension : '';
              const fullName = item.name + extension;
              const newPath = await FSService.JoinPath(finalTargetDirectory, fullName);
              renameOperations.push({ oldPath: item.file_path, newPath });
            }
          }

          if (collectionIdsToMove.length) {
            try {
              await CollectionService.ChangeCollectionParent(projectStore.activeProject.uri, collectionIdsToMove, finalTargetCollectionId);
              notificationStore.addNotification('Moved successfully.', '', 'success');
              needsRefresh = true;
            } catch (error) {
              notificationStore.errorNotification('Error changing collection parent', error);
            }
          }
          if (assetIdsToMove.length) {
            try {
              await AssetService.ChangeAssetCollection(projectStore.activeProject.uri, assetIdsToMove, finalTargetCollectionId);
              notificationStore.addNotification('Moved successfully.', '', 'success');
              needsRefresh = true;
            } catch (error) {
              notificationStore.errorNotification('Error moving assets', error);
            }
          }
          if (renameOperations.length) {
            try {
              await FSService.RenameBatch(JSON.stringify(renameOperations));
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
              if (item.type === 'asset') {
                // Duplicate tracked asset to target collection
                const duplicatedAsset = await AssetService.DuplicateAsset(
                  projectStore.activeProject.uri,
                  item.id,
                  finalTargetCollectionId
                );
                // Copy the physical file if it exists
                if (item.file_path && duplicatedAsset.file_path) {
                  const sourceExists = await FSService.Exists(item.file_path);
                  if (sourceExists) {
                    await FSService.DuplicateFile(item.file_path, duplicatedAsset.file_path);
                  }
                }
                successCount++;
              } else if (item.type === 'untracked_asset') {
                // Copy untracked file
                const fullName = item.name + (item.extension || '');
                const destinationPath = await this.generateUniqueDestinationPath(finalTargetDirectory, fullName);
                await FSService.DuplicateFile(item.file_path, destinationPath);
                successCount++;
              } else if (item.type === 'untracked_collection') {
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