<template>
  <div ref="selectionMenu" class="filter-menu-container">

    <!-- Selection summary -->
    <div v-if="itemCounts.collection" class="pane-parameter-detail">
      {{ itemCounts.collection + ' ' + $t('components.detailsPane.collections') }}
    </div>

    <div v-if="itemCounts.asset" class="pane-parameter-detail">
      {{ itemCounts.asset + ' ' + $t('components.detailsPane.assets') }}
    </div>

    <div v-if="itemsIsUntracked" class="pane-parameter-detail">
      {{ (itemCounts.untracked_asset + itemCounts.untracked_collection) + ' ' + $t('components.detailsPane.untrackedItems') }}
    </div>

    <!-- Context-sensitive actions (based on active item) -->
    <ActionButton v-if="activeIsAsset" :icon="getAppIcon('dependency')" :showLabel="true" :fullWidth="true"
      :label="$t('components.detailsPane.makeDependencies')" :buttonFunction="makeDependenciesOfActive" />

    <ActionButton v-if="activeIsCollection" :icon="getAppIcon('folder-arrow-in')" :showLabel="true" :fullWidth="true"
      :label="$t('components.detailsPane.moveIntoCollection')" :buttonFunction="moveIntoFolder" />

    <!-- Asset-only actions -->
    <template v-if="onlyAssets">
      <span class="menu-divider"></span>

      <ActionButton v-if="!platformStore.isWeb && userStore.canDo('update_asset')" :icon="getAppIcon('folder-arrow-in')"
        :showLabel="true" :fullWidth="true" :label="$t('components.detailsPane.moveToCollection')" :buttonFunction="prepMoveToCollection" />

      <ActionButton v-if="!platformStore.isWeb && assetsCanRebuild" :icon="getAppIcon('jigsaw')" :showLabel="true"
        :fullWidth="true" :label="$t('components.detailsPane.rebuildAssets')" :buttonFunction="revertAllChanges" />

      <ActionButton v-if="assetsModified" :noFilter="true" :icon="getAppIcon('plus-stone')" :useAlert="true" :showLabel="true"
        :fullWidth="true" :label="$t('components.detailsPane.createCheckpoints')" :buttonFunction="prepAllCheckpointModal" />

      <ActionButton v-if="!platformStore.isWeb && assetsModified" :noFilter="true" :icon="getAppIcon('revert')" :useAlert="true"
        :showLabel="true" :fullWidth="true" :label="$t('components.detailsPane.revertAssets')" :buttonFunction="prepResetPopUpModal" />

      <ActionButton :icon="getAppIcon('person-plus')" :showLabel="true" :fullWidth="true"
        :label="$t('components.detailsPane.assignAssets')" :buttonFunction="prepAssignAsset" />

      <ActionButton :icon="getAppIcon('person-minus')" :showLabel="true" :fullWidth="true"
        :label="$t('components.detailsPane.unassignAssets')" :buttonFunction="unassignAssets" />

      <ActionButton :icon="getAppIcon('tag')" :showLabel="true" :fullWidth="true"
        :label="$t('components.detailsPane.manageTags')" :buttonFunction="prepManageTags" />

      <ActionButton v-if="!platformStore.isWeb && assetsOnDisk" :icon="getAppIcon('broom')" :showLabel="true" :fullWidth="true"
        :label="$t('components.detailsPane.freeUpSpace')" :buttonFunction="prepFreeUpSpacePopUpModal" />

      <span v-if="userStore.canDo('delete_asset')" class="menu-divider"></span>

      <ActionButton v-if="userStore.canDo('delete_asset')" :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true"
        :label="$t('components.detailsPane.deleteSelectedAssets')" :buttonFunction="deleteMultipleAssets" />
    </template>

    <!-- Collection-only actions -->
    <template v-else-if="onlyCollections">
      <span class="menu-divider"></span>

      <ActionButton :icon="getAppIcon('person-minus')" :showLabel="true" :fullWidth="true"
        :label="$t('components.detailsPane.unassignCollections')" :buttonFunction="unassignCollections" />

      <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('jigsaw')" :showLabel="true" :fullWidth="true"
        :label="$t('components.detailsPane.rebuildCollections')" :buttonFunction="rebuildCollections" />

      <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('broom')" :showLabel="true" :fullWidth="true"
        :label="$t('components.detailsPane.freeUpSpace')" :buttonFunction="freeUpCollectionSpacePopUpModal" />

      <span v-if="userStore.canDo('delete_collection')" class="menu-divider"></span>

      <ActionButton v-if="userStore.canDo('delete_collection')" :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true"
        :label="$t('components.detailsPane.deleteCollections')" :buttonFunction="deleteMultipleCollections" />
    </template>

    <!-- Untracked-only actions -->
    <template v-else-if="onlyUntracked">
      <span class="menu-divider"></span>

      <ActionButton v-if="userStore.canDo('create_asset') && onlyUntrackedAssets" :icon="getAppIcon('plus-stone')" :useDanger="true"
        :noFilter="true" :showLabel="true" :fullWidth="true" :label="$t('components.detailsPane.createCheckpoints')" :buttonFunction="prepAllCheckpointModal" />

      <ActionButton v-if="squashEnabled" :icon="getAppIcon('squash')" :showLabel="true" :fullWidth="true"
        :label="$t('components.detailsPane.squashAssets')" :buttonFunction="prepSquashModal" />

      <ActionButton :icon="getAppIcon('file-watch')" :showLabel="true" :fullWidth="true"
        :label="$t('components.detailsPane.ignoreItems')" :buttonFunction="ignoreItems" />

      <ActionButton :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true"
        :label="$t('components.detailsPane.deleteItems')" :buttonFunction="deleteMultipleUntrackedAssets" />
    </template>

    <!-- Mixed selection fallback -->
    <template v-else>
      <span class="menu-divider"></span>

      <ActionButton :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true"
        :label="$t('components.detailsPane.deleteItems')" :buttonFunction="deleteMultipleItems" />
    </template>

  </div>
</template>

<script setup>
// imports
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import { getRelativePath } from '@/lib/pathlib';
import { addIgnoredItem } from '@/lib/untracked';
import { canSquash } from '@/utils/squash';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// services
import { AssetService, CheckpointService, CollectionService, FSService, SyncService } from "@/services";

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useDependencyStore } from '@/stores/dependency';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { usePaneStore } from '@/stores/panes';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const dependencyStore = useDependencyStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const panes = usePaneStore();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const trayStates = useTrayStates();
const userStore = useUserStore();
const { t } = useI18n();

// refs
const selectionMenu = ref(null);

// computed properties
// Returns whether the active (last-clicked) item is an asset.
const activeIsAsset = computed(() => {
  const activeAsset = stage.selectedItems.find((item) => item.id === stage.lastSelectedItemId);
  return activeAsset?.type === 'asset';
});

// Returns whether the active (last-clicked) item is a collection.
const activeIsCollection = computed(() => {
  const activeCollection = stage.selectedItems.find((item) => item.id === stage.lastSelectedItemId);
  return activeCollection?.type === 'collection';
});

// Counts the selected items by type.
const itemCounts = computed(() => {
  const counts = { collection: 0, asset: 0, untracked_asset: 0, untracked_collection: 0, resource: 0 };
  stage.selectedItems.forEach(item => { if (item.type in counts) counts[item.type]++; });
  return counts;
});

// Returns whether any selected item is untracked.
const itemsIsUntracked = computed(() => stage.selectedItems.some((item) => item.type === 'untracked_asset' || item.type === 'untracked_collection'));

// Returns whether all selected items are assets.
const onlyAssets = computed(() => stage.selectedItems.every((item) => item.type === 'asset'));

// Returns whether all selected items are collections.
const onlyCollections = computed(() => stage.selectedItems.every((item) => item.type === 'collection'));

// Returns whether all selected items are untracked assets.
const onlyUntrackedAssets = computed(() => stage.selectedItems.every((item) => item.type === 'untracked_asset'));

// Returns whether all selected items are untracked collections.
const onlyUntrackedCollections = computed(() => stage.selectedItems.every((item) => item.type === 'untracked_collection'));

// Returns whether all selected items are untracked (assets or collections).
const onlyUntracked = computed(() => onlyUntrackedAssets.value || onlyUntrackedCollections.value);

// Returns whether any selected asset can be rebuilt.
const assetsCanRebuild = computed(() => stage.selectedItems.filter((item) => item.type === 'asset').some((item) => item.file_status === 'rebuildable'));

// Returns whether any selected asset has been modified.
const assetsModified = computed(() => {
  const modifiedAssetsState = assetStore.getModifiedDisplayPaths;
  return modifiedAssetsState.some((assetState) => stage.markedItems.includes(assetState.asset_id));
});

// Returns whether any selected asset is present on disk.
const assetsOnDisk = computed(() => stage.selectedItems.filter((item) => item.type === 'asset').some((item) => item.file_status !== 'rebuildable'));

// Returns whether the selected items can be squashed.
const squashEnabled = computed(() => {
  if (!userStore.canDo('create_asset')) return false;
  return canSquash(stage.selectedItems).valid;
});

// methods
// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Adds a collection dependency to an asset.
const addCollectionDependency = async (asset, dependencyId) => {
  let dependencyTypeID = dependencyStore.dependency_types.find(item => item.name === "linked").id;
  await AssetService.AddCollectionDependency(projectStore.activeProject.uri, asset.id, dependencyId, dependencyTypeID)
    .then(() => {
      if (!asset.collection_dependencies) asset.collection_dependencies = [];
      asset.collection_dependencies.push(dependencyId);
      notificationStore.addNotification(t('components.detailsPane.dependencyAdded'), "", "success");
    })
    .catch((error) => { console.log(error); notificationStore.errorNotification(t('components.detailsPane.errorAddingDependencies'), error); });
};

// Adds an asset dependency to an asset.
const addAssetDependency = async (asset, dependencyId) => {
  let dependencyTypeID = dependencyStore.dependency_types.find(item => item.name === "linked").id;
  await AssetService.AddAssetDependency(projectStore.activeProject.uri, asset.id, dependencyId, dependencyTypeID)
    .then(() => {
      if (!asset.dependencies) asset.dependencies = [];
      asset.dependencies.push(dependencyId);
      notificationStore.addNotification(t('components.detailsPane.dependencyAdded'), "", "success");
    })
    .catch((error) => { console.log(error); notificationStore.errorNotification(t('components.detailsPane.errorAddingDependencies'), error); });
};

// Moves one or more assets to a different collection.
const changeAssetCollection = async (assetIds, collectionId) => {
  await AssetService.ChangeAssetCollection(projectStore.activeProject.uri, assetIds, collectionId)
    .then(() => notificationStore.addNotification(t('components.detailsPane.movedSuccessfully'), "", "success"))
    .catch((error) => { console.error(error); notificationStore.errorNotification(t('components.detailsPane.errorChangingCollection'), error); });
};

// Changes the parent collection of one or more collections.
const changeCollectionParent = async (collectionIds, parentId) => {
  await CollectionService.ChangeCollectionParent(projectStore.activeProject.uri, collectionIds, parentId)
    .then(() => notificationStore.addNotification(t('components.detailsPane.movedSuccessfully'), "", "success"))
    .catch((error) => { console.error(error); notificationStore.errorNotification(t('components.detailsPane.errorChangingParent'), error); });
};

// Clears all item selections.
const clearSelection = () => {
  stage.markedItems = [];
  stage.selectedItems = [];
  stage.firstSelectedItemId = '';
  stage.lastSelectedItemId = '';
  assetStore.selectedAsset = null;
  collectionStore.selectedCollection = null;
};

// Closes all modals.
const closeModals = () => modals.disableAllModals();

// Deletes multiple assets.
const deleteMultipleAssets = async () => {
  menu.hideContextMenu();
  stage.operationActive = true;
  for (let assetId of stage.markedItems) {
    await AssetService.DeleteAsset(projectStore.activeProject.uri, assetId, true)
      .then(() => { emitter.emit('refresh-browser'); notificationStore.addNotification(t('components.detailsPane.assetsMovedToTrash'), '', "success", false); })
      .catch((error) => { if (onlyAssets.value) { console.log(error); notificationStore.errorNotification(t('components.detailsPane.assetsFailedToDelete'), error); } });
  }
  stage.operationActive = false;
};

// Deletes multiple collections.
const deleteMultipleCollections = async () => {
  menu.hideContextMenu();
  stage.operationActive = true;
  for (let collectionId of stage.markedItems) {
    await CollectionService.DeleteCollection(projectStore.activeProject.uri, collectionId, true)
      .then(() => { if (onlyCollections.value) { stage.markedItems = []; collectionStore.selectedCollection = null; } })
      .catch((error) => { console.log(error); notificationStore.errorNotification(t('components.detailsPane.collectionsFailedToDelete'), error); });
  }
  clearSelection();
  notificationStore.addNotification(t('components.detailsPane.collectionsMovedToTrash'), '', "success", false);
  emitter.emit('refresh-browser');
  stage.operationActive = false;
};

// Deletes multiple items (assets and collections).
const deleteMultipleItems = async () => {
  menu.hideContextMenu();
  panes.setPaneVisibility('projectDetails', true);
  await deleteMultipleCollections();
  await deleteMultipleAssets();
  stage.markedItems = [];
  collectionStore.selectedCollection = null;
};

// Deletes multiple untracked items.
const deleteMultipleUntrackedAssets = async () => {
  menu.hideContextMenu();
  stage.operationActive = true;
  try {
    for (let untrackedItem of stage.selectedItems) {
      if (untrackedItem.type === 'untracked_asset') {
        await FSService.DeleteFile(untrackedItem.file_path);
        projectStore.removeUntrackedAsset(untrackedItem.id);
      } else if (untrackedItem.type === 'untracked_collection') {
        await FSService.DeleteFolder(untrackedItem.file_path);
        projectStore.removeUntrackedCollection(untrackedItem.id);
      }
    }
    if (onlyUntracked.value) { stage.markedItems = []; projectStore.selectedUntrackedItem = null; }
    emitter.emit('refresh-browser');
    notificationStore.addNotification(t('components.detailsPane.untrackedItemsDeleted'), '', "success", false);
  } catch (error) { console.error(error); notificationStore.errorNotification(t('components.detailsPane.failedToDeleteUntracked'), error); }
  stage.operationActive = false;
};

// Emits item data updates to notify components.
const emitItemUpdates = (assetId, updates) => {
  const selectedItemIndex = stage.selectedItems.findIndex(item => item.id === assetId);
  if (selectedItemIndex !== -1) {
    if (typeof updates === 'object' && !Array.isArray(updates)) {
      Object.assign(stage.selectedItems[selectedItemIndex], updates);
    } else if (Array.isArray(updates)) {
      updates.forEach(update => { if (update.property && update.value !== undefined) stage.selectedItems[selectedItemIndex][update.property] = update.value; });
    }
  }
  const updateData = { itemId: assetId, updates };
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

// Frees up collection space by deleting contents.
const freeUpCollectionSpace = async () => {
  const selectedCollections = stage.selectedItems.filter(item => item.type === 'collection');
  for (const collection of selectedCollections) {
    let collectionPath = collection.file_path.replace(/\\/g, '/');
    await FSService.DeleteFolder(collectionPath).catch((error) => { console.error(error); notificationStore.errorNotification('Error freeing collection space', error); });
  }
  closeModals();
  emitter.emit('refresh-browser');
};

// Shows the free up collection space confirmation modal.
const freeUpCollectionSpacePopUpModal = () => {
  menu.hideContextMenu();
  trayStates.popUpModalTitle = t('components.detailsPane.freeUpCollectionSpaceTitle');
  trayStates.popUpModalMessage = t('components.detailsPane.freeUpCollectionSpaceMessage');
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpCollectionSpace;
  modals.setModalVisibility('popUpModal', true);
};

// Frees up asset space by deleting working files.
const freeUpSpace = async () => {
  const selectedAssets = stage.selectedItems.filter(item => item.type === 'asset');
  const fileStatus = ['missing', 'rebuildable'];
  const assetsToProcess = selectedAssets.filter(asset => !fileStatus.includes(asset.file_status));
  for (const asset of assetsToProcess) {
    let assetPath = asset.file_path.replace(/\\/g, '/');
    await FSService.DeleteFile(assetPath)
      .then(() => {
        asset.file_status = 'rebuildable';
        assetStore.rebuildableAssetsPath.push(asset.asset_path + asset.extension);
        assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(assetPath => assetPath !== asset.asset_path + asset.extension);
        assetStore.modifiedAssetsPath = assetStore.modifiedAssetsPath.filter(assetPath => assetPath !== asset.asset_path + asset.extension);
        emitItemUpdates(asset.id, [{ property: 'file_status', value: 'rebuildable' }]);
      })
      .catch((error) => console.error(error));
  }
  closeModals();
};

// Adds items to the ignore list.
const ignoreItems = async () => {
  menu.hideContextMenu();
  stage.operationActive = true;
  try {
    for (let untrackedItem of stage.selectedItems) {
      if (untrackedItem.type == "untracked_asset") {
        await addIgnoredItem(untrackedItem.asset_path);
      } else {
        const untrackedCollection = removeLastSlash(untrackedItem.item_path);
        await addIgnoredItem(untrackedCollection);
      }
    }
    panes.setPaneVisibility('projectDetails', true);
    clearSelection();
    emitter.emit('refresh-browser');
    notificationStore.addNotification(t('components.detailsPane.updatedIgnoreList'), '', "success", false);
  } catch (error) {
    notificationStore.addNotification(t('components.detailsPane.failedToUpdateIgnoreList'), "error");
  }
  stage.operationActive = false;
};

// Makes selected items dependencies of the active asset.
const makeDependenciesOfActive = async () => {
  menu.hideContextMenu();
  stage.operationActive = true;
  const activeItemId = stage.lastSelectedItemId;
  const selectedItems = stage.selectedItems.filter((item) => item.id !== activeItemId);
  const asset = stage.selectedItems.find((item) => item.id === activeItemId);
  for (const item of selectedItems) {
    if (item.type === 'collection') {
      if (item.id !== asset.collection_id && !asset.collection_dependencies?.includes(item.id)) {
        await addCollectionDependency(asset, item.id);
      }
    } else {
      if (!asset.dependencies?.includes(item.id)) {
        await addAssetDependency(asset, item.id);
      }
    }
  }
  emitItemUpdates(asset.id, [
    { property: 'dependencies', value: asset.dependencies },
    { property: 'collection_dependencies', value: asset.collection_dependencies }
  ]);
  stage.operationActive = false;
};

// Moves selected items into the active collection.
const moveIntoFolder = async () => {
  menu.hideContextMenu();
  stage.operationActive = true;
  const activeItemId = stage.lastSelectedItemId;
  const selectedItems = stage.selectedItems.filter((item) => item.id !== activeItemId);

  const collectionIds = [];
  const assetIds = [];
  const untrackedItems = [];

  for (const item of selectedItems) {
    if (item.type === 'collection') collectionIds.push(item.id);
    else if (item.type === 'asset') assetIds.push(item.id);
    else untrackedItems.push(item);
  }

  if (collectionIds.length) await changeCollectionParent(collectionIds, activeItemId);
  if (assetIds.length) await changeAssetCollection(assetIds, activeItemId);

  if (untrackedItems.length) {
    let collection = collectionStore.findCollection(activeItemId);
    await FSService.MakeDirs(collection.file_path);
    const renameOperations = [];
    const itemUpdates = [];

    for (const item of untrackedItems) {
      let newPath = await FSService.JoinPath(collection.file_path, item.name);
      const untrackedPath = newPath.replace(/^\/+|\/+$/g, "").replace(/\\/g, "/");
      const workingDir = projectStore.activeProject.working_directory.replace(/^\/+|\/+$/g, "").replace(/\\/g, "/");
      const itemPath = getRelativePath(workingDir, untrackedPath);
      let collectionPath = "";
      const itemPathCollections = itemPath.split("/");
      if (itemPathCollections.length > 1) collectionPath = itemPathCollections.slice(0, -1).join("/");
      renameOperations.push({ oldPath: item.file_path, newPath });
      itemUpdates.push({ item, itemPath, newPath, collectionPath });
    }

    await FSService.RenameBatch(JSON.stringify(renameOperations));

    for (const { item, itemPath, newPath, collectionPath } of itemUpdates) {
      if (item.type == "untracked_asset") {
        let untrackedAssetIndex = projectStore.untrackedFilesIndex[item.id];
        projectStore.untrackedFiles[untrackedAssetIndex].item_path = itemPath;
        projectStore.untrackedFiles[untrackedAssetIndex].file_path = newPath;
        projectStore.untrackedFiles[untrackedAssetIndex].collection_path = collectionPath;
      } else if (item.type == "untracked_collection") {
        let untrackedFolderIndex = projectStore.untrackedFoldersIndex[item.id];
        projectStore.untrackedFolders[untrackedFolderIndex].item_path = itemPath;
        projectStore.untrackedFolders[untrackedFolderIndex].file_path = newPath;
        projectStore.untrackedFolders[untrackedFolderIndex].collection_path = collectionPath;
      }
    }
  }

  emitter.emit('refresh-browser');
  stage.operationActive = false;
};

// Shows the create checkpoints modal.
const prepAllCheckpointModal = () => {
  menu.hideContextMenu();
  trayStates.createMultipleCheckpoints = false;
  modals.setModalVisibility('createMultipleCheckpointsModal', true);
};

// Opens the assign menu for the currently selected assets.
const prepAssignAsset = (event) => menu.showContextMenu(event, 'assignMenu', true);

// Shows the free up asset space confirmation modal.
const prepFreeUpSpacePopUpModal = () => {
  menu.hideContextMenu();
  trayStates.popUpModalTitle = t('components.detailsPane.freeUpAssetSpaceTitle');
  trayStates.popUpModalMessage = t('components.detailsPane.freeUpAssetSpaceMessage');
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpSpace;
  modals.setModalVisibility('popUpModal', true);
};

// Opens the manage tags menu for the currently selected assets.
const prepManageTags = (event) => menu.showContextMenu(event, 'manageTagsMenu', true);

// Prepares and opens the Move to Collection sub-menu for multi-selection.
const prepMoveToCollection = (event) => {
  const selectedAssetIds = stage.markedItems.filter(id => stage.selectedItems.find(item => item.id === id && item.type === 'asset'));
  if (selectedAssetIds.length === 0) return;
  const firstAsset = stage.selectedItems.find(item => item.id === selectedAssetIds[0] && item.type === 'asset');
  menu.subMenuState.selectedAssetIds = selectedAssetIds;
  menu.subMenuState.startingCollectionId = firstAsset?.parent_id || '';
  menu.position = { x: event.clientX, y: event.clientY };
  menu.showSubMenu('move-to-collection');
};

// Shows the revert assets confirmation modal.
const prepResetPopUpModal = () => {
  menu.hideContextMenu();
  trayStates.popUpModalIcon = 'revert';
  trayStates.popUpModalTitle = t('components.detailsPane.revertSelectedAssetsTitle');
  trayStates.popUpModalMessage = t('components.detailsPane.revertSelectedAssetsMessage');
  trayStates.popUpModalFunction = revertAllChanges;
  modals.setModalVisibility('popUpModal', true);
};

// Shows the squash modal.
const prepSquashModal = () => {
  menu.hideContextMenu();
  modals.setModalVisibility('squashModal', true);
};

// Rebuilds multiple collections.
const rebuildCollections = async () => {
  menu.hideContextMenu();
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  try {
    const collectionIdsString = stage.markedItems.join(',');
    await CollectionService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, collectionIdsString)
      .then(() => {
        assetStore.refreshCollectionFilesStatus();
        notificationStore.addNotification(t('components.detailsPane.collectionsRebuiltSuccessfully', { count: stage.markedItems.length }), '', "success", false);
      })
      .catch((error) => { console.error('Error rebuilding collections:', error); notificationStore.errorNotification(t('components.detailsPane.errorRebuildingCollections'), error); });
  } catch (error) {
    console.error('Error rebuilding collections:', error);
    notificationStore.errorNotification(t('components.detailsPane.errorRebuildingCollections'), error);
  } finally {
    emitter.emit('refresh-browser');
    notificationStore.canCancel = false;
  }
};

// Removes the trailing slash from a path.
const removeLastSlash = (path) => path.replace(/\/+$/, '');

// Reverts all selected assets to their last checkpoint.
const revertAllChanges = async () => {
  menu.hideContextMenu();
  modals.setModalVisibility('popUpModal', false);
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  await CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, stage.markedItems)
    .then(() => {
      const revertedAssets = stage.selectedItems.filter(item => item.type === 'asset');
      for (const asset of revertedAssets) {
        asset.file_status = 'normal';
        emitItemUpdates(asset.id, [{ property: 'file_status', value: 'normal' }]);
      }
    })
    .catch((error) => { notificationStore.errorNotification(t('components.detailsPane.errorRevertingAssets'), error); console.error(error); });
};

// Unassigns all collaborators from assets.
const unassignAssets = async () => {
  menu.hideContextMenu();
  for (const assetId of stage.markedItems) {
    await AssetService.UnassignAsset(projectStore.activeProject.uri, assetId).catch((error) => { console.log(error); notificationStore.errorNotification(t('components.detailsPane.errorAssigningAsset'), error); });
  }
  emitter.emit('refresh-browser');
  notificationStore.addNotification(t('components.detailsPane.assetsUnassignedSuccessfully'), "", "success");
};

// Unassigns all collaborators from collections.
const unassignCollections = async () => {
  menu.hideContextMenu();
  stage.operationActive = true;
  for (const collection of stage.selectedItems) {
    const currentAssigneeIds = collection.assignee_ids || [];
    for (const assigneeId of currentAssigneeIds) {
      await CollectionService.Unassign(projectStore.activeProject.uri, collection.id, assigneeId)
        .then(() => {
          const itemIndex = stage.selectedItems.findIndex(item => item.id === collection.id);
          if (itemIndex !== -1) stage.selectedItems[itemIndex].assignee_ids = stage.selectedItems[itemIndex].assignee_ids.filter(id => id !== assigneeId);
        })
        .catch((error) => { notificationStore.errorNotification(t('components.detailsPane.errorRemovingUser'), error); console.error('Error removing user:', error); });
    }
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;
};
</script>

<style scoped>

@import "@/assets/desktop.css";
@import "@/assets/menu.css";

.pane-parameter-detail {
  width: 100%;
  padding: .2rem .4rem;
  box-sizing: border-box;
  font-size: .8rem;
  opacity: .7;
}
</style>
