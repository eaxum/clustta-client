<template>
  <div ref="collectionMenu" class="filter-menu-container">

    <ActionButton :icon="getAppIcon('edit')" :showLabel="true" :fullWidth="true" :label="$t('common.rename')"
      v-if="userStore.canDo('update_collection')" :buttonFunction="renameCollection" />

    <ActionButton :icon="getAppIcon('switches')" :showLabel="true" :fullWidth="true" :label="$t('common.edit')"
      v-if="userStore.canDo('update_collection')" :buttonFunction="editCollection" />

    <ActionButton v-if="canSelectContent" :icon="getAppIcon('checkbox-selected')" :showLabel="true" :fullWidth="true"
      :label="$t('menus.selectContent')" :buttonFunction="selectContent" />

    <span v-if="userStore.canDo('update_collection')" class="menu-divider"></span>

    <!-- Create -->
    <ActionButton :icon="getAppIcon('file-plus')" :showLabel="true" :fullWidth="true" :label="$t('menus.addAsset')"
      v-if="templateStore.getTemplates.length && (userStore.canDo('create_asset') || collectionStore.selectedCollection.can_modify)" :buttonFunction="createAsset" />

    <ActionButton :icon="getAppIcon('folder-plus')" :showLabel="true" :fullWidth="true" :label="$t('menus.addCollection')"
      v-if="userStore.canDo('create_collection') || collectionStore.selectedCollection.can_modify" :buttonFunction="createCollection" />

    <ActionButton :icon="getAppIcon('workflow-plus')" :showLabel="true" :fullWidth="true" :label="$t('menus.addWorkflow')"
      v-if="workflowStore.workflows.length && userStore.canDo('create_asset')" :buttonFunction="addWorkflow" />

    

    <ActionButton :icon="getAppIcon('web-plus')" :showLabel="true" :fullWidth="true" :label="$t('menus.newLink')"
      v-if="userStore.canDo('create_asset') || collectionStore.selectedCollection.can_modify" :buttonFunction="createLink" />

    <ActionButton :icon="getAppIcon('data-download')" :showLabel="true" :fullWidth="true" :label="$t('modals.importItems')"
      v-if="!platformStore.isWeb && userStore.canDo('create_asset')" :buttonFunction="importItems" />

    <ActionButton :icon="getAppIcon('arrow-up-ramp')" :showLabel="true" :fullWidth="true" :label="$t('menus.uploadItems')"
      v-if="platformStore.isWeb && userStore.canDo('create_asset')" :buttonFunction="uploadItems" />

    <ActionButton :icon="getAppIcon('clipboard')" :showLabel="true" :fullWidth="true" :label="$t('common.paste')"
      v-if="hasClipboardItems && userStore.canDo('update_collection')" :buttonFunction="pasteItems" />

    
    <!-- Collection State Actions -->
    <span v-if="collectionStateFlags.has_untracked || collectionStateFlags.has_modified || collectionStateFlags.has_outdated || collectionStateFlags.has_rebuildable" class="menu-divider"></span>

    <ActionButton v-if="collectionStateFlags.has_untracked || collectionStateFlags.has_modified" :icon="getAppIcon('plus-stone')" :useAlert="collectionStateFlags.has_modified" :useDanger="collectionStateFlags.has_untracked" :showLabel="true" :fullWidth="true" :label="$t('modals.createCheckpoints')"
      :buttonFunction="prepCreateCheckpointsModal" />

    <ActionButton v-if="!platformStore.isWeb && collectionStateFlags.has_rebuildable" :icon="getAppIcon('jigsaw')" :showLabel="true" :fullWidth="true" :label="$t('menus.rebuildContents')"
      :buttonFunction="rebuildCollection" />

    <ActionButton v-if="!platformStore.isWeb && collectionStateFlags.has_outdated" :icon="getAppIcon('circle-check')" :useAlert="true" :noFilter="true" :showLabel="true" :fullWidth="true" :label="$t('menus.updateContents')"
      :buttonFunction="updateContents" />

    <!-- Revert Contents -->
    <ActionButton v-if="!platformStore.isWeb && collectionStateFlags.has_modified" :noFilter="true" :icon="getAppIcon('revert')" :useAlert="true" :showLabel="true" :fullWidth="true" 
      :label="$t('menus.revertContents')" :buttonFunction="prepRevertContentsPopUpModal" />



    <span v-if="userStore.canDo('update_collection') || collectionStore.selectedCollection.can_modify" class="menu-divider"></span>

    <!-- Reveal in Explorer -->
    <span v-if="!platformStore.isWeb" class="horizontal-flex">
      <ActionButton :icon="getAppIcon('folder-arrow-up-right')" :showLabel="true" :fullWidth="true" :label="$t('common.showInExplorer')"
        :buttonFunction="revealInExplorer" />
      <ActionButton :icon="getAppIcon('copy')" :showLabel="false" :fullWidth="false" @click="copyCollectionPath('collection')"
        v-tooltip="$t('common.copyPath')" />
    </span>

    <!-- Copy Clustta deep link (dev/test) -->
    <ActionButton :icon="getAppIcon('copy')" :showLabel="true" :fullWidth="true"
      :label="$t('menus.copyClusttaLink')" :buttonFunction="copyDeepLink" />

    <!-- Free space -->
    <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('broom')" :showLabel="true" :fullWidth="true" :label="$t('common.freeUpSpace')"
      :buttonFunction="prepFreeUpSpacePopUpModal" />

    <!-- Delete Asset -->
    <ActionButton :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true" :label="$t('common.delete')"
      v-if="userStore.canDo('delete_collection')" :buttonFunction="deleteCollection" />

  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { Clipboard } from '@wailsio/runtime';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// services
import { CheckpointService, CollectionService, DialogService, FSService, SyncService } from "@/services";

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { usePaneStore } from '@/stores/panes';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTemplateStore } from '@/stores/template';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';
import { useWorkflowStore } from '@/stores/workflow';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const panes = usePaneStore();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const templateStore = useTemplateStore();
const trayStates = useTrayStates();
const userStore = useUserStore();
const workflowStore = useWorkflowStore();
const { t } = useI18n();

// refs
const collectionMenu = ref(null);

// computed
// Checks if content can be selected.
const canSelectContent = computed(() => {
  const collectionId = collectionStore.selectedCollection.id;
  return collectionId in stage.expandedCollections && stage.collectionDataIds.length;
});

// Returns the collection state flags.
const collectionStateFlags = computed(() => {
  const collection = collectionStore.selectedCollection;
  if (!collection) return {
    has_untracked: false,
    has_modified: false,
    has_outdated: false,
    has_rebuildable: false
  };
  
  return collection.collectionStateFlags || {
    has_untracked: false,
    has_modified: false,
    has_outdated: false,
    has_rebuildable: false
  };
});

// Checks if there are items in the clipboard.
const hasClipboardItems = computed(() => {
  return stage.cutItems.length > 0 || stage.copiedItems.length > 0;
});

// methods
// Opens the workflow selection modal.
const addWorkflow = () => {
  modals.setModalVisibility('selectWorkflowModal', true);
  menu.hideContextMenu();
};

// Copies the collection path to clipboard.
const copyCollectionPath = async () => {
  let collection = collectionStore.selectedCollection;
  let collectionDir = collection.file_path;
  collectionDir = collectionDir.replace(/\\/g, '/');
  FSService.MakeDirs(collectionDir);
  await Clipboard.SetText(collectionDir);
  notificationStore.addNotification(t('notifications.pathCopied'), "", "success");
  menu.hideContextMenu();
};

// Builds clustta://open?studio=...&project=...&collection=... and copies it to the clipboard.
const copyDeepLink = async () => {
  const selectedCollection = collectionStore.selectedCollection;
  const project = projectStore.activeProject;
  if (!selectedCollection || !project) {
    menu.hideContextMenu();
    return;
  }

  const params = new URLSearchParams();
  const studioName = projectStore.selectedStudio?.name;
  if (studioName) params.set('studio', studioName);
  params.set('project', project.id);
  params.set('collection', selectedCollection.id);

  const deepLink = `clustta://open?${params.toString()}`;
  await Clipboard.SetText(deepLink);
  notificationStore.addNotification(t('notifications.pathCopied'), deepLink, 'success');
  menu.hideContextMenu();
};

// Opens the create collection modal.
const createCollection = () => {
  stage.expandCollection(collectionStore.selectedCollection);
  modals.setModalVisibility('createCollectionModal', true);
  menu.hideContextMenu();
};

// Opens the add web link modal.
const createLink = () => {
  stage.expandCollection(collectionStore.selectedCollection);
  modals.setModalVisibility('addWebLinkModal', true);
  menu.hideContextMenu();
};

// Opens the select app modal to create an asset.
const createAsset = () => {
  stage.expandCollection(collectionStore.selectedCollection);
  modals.setModalVisibility('selectAppModal', true);
  menu.hideContextMenu();
};

// Deletes the selected collection.
const deleteCollection = async () => {
  let collection = collectionStore.selectedCollection;
  panes.setPaneVisibility('projectDetails', true);
  CollectionService.DeleteCollection(projectStore.activeProject.uri, collection.id)
    .then(async () => {
      stage.markedItems = [];
      collectionStore.selectedCollection = null;
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      console.error(error);
    });
  let longMessage = t('notifications.movedToTrash', { item: collection.name });
  notificationStore.addNotification(t('notifications.movedToTrash', { item: 'Collection' }), longMessage, "success", true);
  menu.hideContextMenu();
};

// Opens the edit collection modal.
const editCollection = () => {
  modals.setModalVisibility('editCollectionModal', true);
  menu.hideContextMenu();
};

// Frees up space by deleting the collection's working files.
const freeUpSpace = async () => {
  let collection = collectionStore.selectedCollection;
  let collectionDir = collection.file_path.replace(/\\/g, '/');
  await FSService.DeleteFolder(collectionDir)
    .then(() => {
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      console.error(error);
    });
  modals.disableAllModals();
  menu.hideContextMenu();
};

// Generates a unique destination path for imports.
const generateUniqueDestinationPath = async (directory, fileName) => {
  const originalPath = await FSService.JoinPath(directory, fileName);
  const exists = await FSService.Exists(originalPath);
  if (!exists) return originalPath;
  
  const baseName = fileName.includes('.') 
    ? fileName.substring(0, fileName.lastIndexOf('.'))
    : fileName;
  const extension = fileName.includes('.') 
    ? fileName.substring(fileName.lastIndexOf('.'))
    : '';
  
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
  const timestampFileName = `${baseName}_${timestamp}${extension}`;
  return await FSService.JoinPath(directory, timestampFileName);
};

// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Returns the current directory path.
const getCurrentDirectory = () => {
  return collectionStore.selectedCollection?.file_path;
};

// Imports files and folders from the file system.
const importItems = async () => {
  try {
    const selectedPaths = await DialogService.SelectFilesDialog();
    if (!selectedPaths || selectedPaths.length === 0) {
      menu.hideContextMenu();
      return;
    }

    const currentDirectory = getCurrentDirectory();
    if (!currentDirectory) {
      notificationStore.errorNotification(t('notifications.couldNotDetermineDirectory'), "");
      menu.hideContextMenu();
      return;
    }

    await FSService.MakeDirs(currentDirectory);
    stage.operationActive = true;
    
    let successCount = 0;
    let failureCount = 0;
    const errors = [];

    for (const sourcePath of selectedPaths) {
      try {
        const isFile = await FSService.IsFile(sourcePath);
        const itemName = await FSService.BaseName(sourcePath);
        const destinationPath = await generateUniqueDestinationPath(currentDirectory, itemName);
        
        if (isFile) {
          await FSService.DuplicateFile(sourcePath, destinationPath);
        } else {
          await FSService.DuplicateFolder(sourcePath, destinationPath);
        }
        successCount++;
      } catch (error) {
        failureCount++;
        const itemName = await FSService.BaseName(sourcePath).catch(() => sourcePath);
        errors.push(`${itemName}: ${error.message || error}`);
      }
    }

    if (successCount > 0) {
      const message = t('notifications.itemsImported', successCount);
      notificationStore.addNotification(message, "", "success");
    }

    if (failureCount > 0) {
      const message = t('notifications.itemsFailedImport', failureCount);
      notificationStore.errorNotification(message, errors.join("\n"));
    }

    if (successCount > 0) {
      stage.expandCollection(collectionStore.selectedCollection);
      emitter.emit('refresh-browser');
    }
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorImportingItems'), error.message || error);
  } finally {
    stage.operationActive = false;
    menu.hideContextMenu();
  }
};

// Pastes clipboard items into the selected collection.
const pasteItems = async () => {
  const collection = collectionStore.selectedCollection;
  if (!collection) return;
  
  menu.hideContextMenu();
  
  const result = await stage.pasteItems(collection.id, collection.file_path);
  if (result.needsRefresh) {
    stage.expandCollection(collection);
    emitter.emit('refresh-browser');
  }
};

// Prepares and shows the create checkpoints modal.
const prepCreateCheckpointsModal = () => {
  const collection = collectionStore.selectedCollection;
  trayStates.createMultipleCheckpointsCollectionPath = collection.collection_path;
  modals.setModalVisibility('createMultipleCheckpointsModal', true);
  menu.hideContextMenu();
};

// Prepares and shows the free up space confirmation modal.
const prepFreeUpSpacePopUpModal = () => {
  trayStates.popUpModalTitle = t('menus.freeUpCollectionSpace');
  trayStates.popUpModalMessage = t('confirmations.deleteWorkingFiles', { item: 'collection' });
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpSpace;
  modals.setModalVisibility('popUpModal', true);
  menu.hideContextMenu();
};

// Prepares and shows the revert contents confirmation modal.
const prepRevertContentsPopUpModal = () => {
  trayStates.popUpModalIcon = 'revert';
  trayStates.popUpModalTitle = t('menus.revertContents');
  trayStates.popUpModalMessage = t('confirmations.revertContents');
  trayStates.popUpModalFunction = revertContents;
  modals.setModalVisibility('popUpModal', true);
  menu.hideContextMenu();
};

// Rebuilds the collection contents.
const rebuildCollection = () => {
  menu.hideContextMenu();
  let collection = collectionStore.selectedCollection;
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  CollectionService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, collection.id)
    .then(() => {
      assetStore.refreshCollectionFilesStatus(collection.id);
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      console.error(error);
    });
  notificationStore.canCancel = false;
};

// Emits event to rename the collection.
const renameCollection = () => {
  emitter.emit('renameCollection');
  menu.hideContextMenu();
};

// Reverts all modified contents in the collection.
const revertContents = async () => {
  modals.setModalVisibility('popUpModal', false);
  
  const collection = collectionStore.selectedCollection;
  if (!collection) return;
  
  const collectionId = collection.id;
  await collectionStore.reloadItemsForCheckpoint(collectionId, null);
  const filteredPaths = assetStore.modifiedAssets.modified.map(asset => asset.asset_path);
  
  if (filteredPaths.length === 0) {
    notificationStore.addNotification(t('notifications.noModifiedContents'), "", "info");
    return;
  }
  
  try {
    await CheckpointService.RevertAssetPaths(
      projectStore.activeProject.uri, 
      projectStore.getActiveProjectUrl, 
      filteredPaths
    );
    
    assetStore.modifiedAssets.modified = assetStore.modifiedAssets.modified.filter(
      (item) => !filteredPaths.includes(item.asset_path)
    );
    
    emitter.emit('refresh-browser');
    
    const message = t('notifications.itemsRevertedSuccessfully', filteredPaths.length);
    notificationStore.addNotification(message, "", "success");
  } catch (error) {
    notificationStore.errorNotification(t('notifications.failedToRevertContents'), error);
    console.error(error);
  }
};

// Reveals the collection in the file explorer.
const revealInExplorer = async () => {
  await FSService.MakeDirs(collectionStore.selectedCollection.file_path);
  FSService.RevealInExplorer(collectionStore.selectedCollection.file_path);
  menu.hideContextMenu();
};

// Selects all content in the collection.
const selectContent = () => {
  stage.markedItems = stage.collectionDataIds;
  menu.hideContextMenu();
};

// Updates outdated contents in the collection.
const updateContents = async () => {
  menu.hideContextMenu();
  
  const collection = collectionStore.selectedCollection;
  if (!collection) return;
  
  const collectionPath = collection.collection_path;
  const outdatedAssetsPath = assetStore.outdatedAssetsPath;
  const collectionOutdatedPaths = outdatedAssetsPath.filter(assetPath => assetPath.startsWith(collectionPath));
  
  if (collectionOutdatedPaths.length === 0) {
    notificationStore.addNotification(t('notifications.noOutdatedContents'), "", "info");
    return;
  }
  
  try {
    notificationStore.cancleFunction = SyncService.CancelSync;
    notificationStore.canCancel = true;
    
    await CheckpointService.RevertAssetPaths(
      projectStore.activeProject.uri, 
      projectStore.getActiveProjectUrl, 
      collectionOutdatedPaths
    );
    
    assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(
      assetPath => !collectionOutdatedPaths.includes(assetPath)
    );
    
    emitter.emit('refresh-browser');
    
    const message = t('notifications.itemsUpdatedSuccessfully', collectionOutdatedPaths.length);
    notificationStore.addNotification(message, "", "success");
  } catch (error) {
    notificationStore.errorNotification(t('notifications.failedToUpdateContents'), error);
    console.error(error);
  } finally {
    notificationStore.canCancel = false;
  }
};

// Opens the upload items modal for web platform.
const uploadItems = () => {
  stage.expandCollection(collectionStore.selectedCollection);
  modals.setModalVisibility('uploadItemsModal', true);
  menu.hideContextMenu();
};

// lifecycle hooks
onMounted(() => {
  menu.assetMenuWidth = collectionMenu.value.getBoundingClientRect().width;
  menu.collectionMenu = collectionMenu.value;
});

onBeforeUnmount(() => {
  menu.assetMenuWidth = collectionMenu.value.getBoundingClientRect().width;
  menu.assetMenuHeight = collectionMenu.value.getBoundingClientRect().height;
});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/menu.css";
</style>






