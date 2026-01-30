<template>
  <div ref="collectionMenu" class="filter-menu-container">

    <ActionButton :icon="getAppIcon('edit')" :showLabel="true" :fullWidth="true" label="Rename"
      v-if="userStore.canDo('update_entity')" :buttonFunction="renameEntity" />

    <ActionButton :icon="getAppIcon('switches')" :showLabel="true" :fullWidth="true" label="Edit"
      v-if="userStore.canDo('update_entity')" :buttonFunction="editEntity" />

    <ActionButton v-if="canSelectContent" :icon="getAppIcon('checkbox-selected')" :showLabel="true" :fullWidth="true"
      label="Select Content" :buttonFunction="selectContent" />

    <span v-if="userStore.canDo('update_entity')" class="menu-divider"></span>

    <!-- Create -->
    <ActionButton :icon="getAppIcon('file-plus')" :showLabel="true" :fullWidth="true" label="Add Asset"
      v-if="templateStore.getTemplates.length && (userStore.canDo('create_task') || collectionStore.selectedCollection.can_modify)" :buttonFunction="createTask" />

    <ActionButton :icon="getAppIcon('folder-plus')" :showLabel="true" :fullWidth="true" label="Add Collection"
      v-if="userStore.canDo('create_entity') || collectionStore.selectedCollection.can_modify" :buttonFunction="createEntity" />

    <ActionButton :icon="getAppIcon('workflow-plus')" :showLabel="true" :fullWidth="true" label="Add Workflow"
      v-if="workflowStore.workflows.length && userStore.canDo('create_task')" :buttonFunction="addWorkflow" />

    

    <ActionButton :icon="getAppIcon('web-plus')" :showLabel="true" :fullWidth="true" label="New Link"
      v-if="userStore.canDo('create_task') || collectionStore.selectedCollection.can_modify" :buttonFunction="createLink" />

    <ActionButton :icon="getAppIcon('arrow-down-ramp')" :showLabel="true" :fullWidth="true" label="Import Items"
      v-if="!platformStore.isWeb && userStore.canDo('create_task')" :buttonFunction="importItems" />

    <ActionButton :icon="getAppIcon('arrow-up-ramp')" :showLabel="true" :fullWidth="true" label="Upload Items"
      v-if="platformStore.isWeb && userStore.canDo('create_task')" :buttonFunction="uploadItems" />

    
    <!-- Collection State Actions -->
    <span v-if="collectionStateFlags.has_untracked || collectionStateFlags.has_modified || collectionStateFlags.has_outdated || collectionStateFlags.has_rebuildable" class="menu-divider"></span>

    <ActionButton v-if="collectionStateFlags.has_untracked || collectionStateFlags.has_modified" :icon="getAppIcon('layers-plus')" :useAlert="collectionStateFlags.has_modified" :useDanger="collectionStateFlags.has_untracked" :showLabel="true" :fullWidth="true" label="Create Checkpoints"
      :buttonFunction="prepCreateCheckpointsModal" />

    <ActionButton v-if="!platformStore.isWeb && collectionStateFlags.has_rebuildable" :icon="getAppIcon('jigsaw')" :showLabel="true" :fullWidth="true" label="Rebuild Contents"
      :buttonFunction="rebuildCollection" />

    <ActionButton v-if="!platformStore.isWeb && collectionStateFlags.has_outdated" :icon="getAppIcon('circle-check')" :useAlert="true" :noFilter="true" :showLabel="true" :fullWidth="true" label="Update Contents"
      :buttonFunction="updateContents" />

    <!-- Revert Contents -->
    <ActionButton v-if="!platformStore.isWeb && collectionStateFlags.has_modified" :noFilter="true" :icon="getAppIcon('revert')" :useAlert="true" :showLabel="true" :fullWidth="true" 
      label="Revert Contents" :buttonFunction="prepRevertContentsPopUpModal" />



    <span v-if="userStore.canDo('update_entity') || collectionStore.selectedCollection.can_modify" class="menu-divider"></span>

    <!-- Reveal in Explorer -->
    <span v-if="!platformStore.isWeb" class="horizontal-flex">
      <ActionButton :icon="getAppIcon('folder-arrow-up-right')" :showLabel="true" :fullWidth="true" label="Show in Explorer"
        :buttonFunction="revealInExplorer" />
      <ActionButton :icon="getAppIcon('copy')" :showLabel="false" :fullWidth="false" @click="copyEntityPath('entity')"
        v-tooltip="'Copy Path'" />
    </span>

    <!-- Free space -->
    <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('broom')" :showLabel="true" :fullWidth="true" label="Free Up space"
      :buttonFunction="prepFreeUpSpacePopUpModal" />

    <!-- Delete Task -->
    <ActionButton :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true" label="Delete"
      v-if="userStore.canDo('delete_entity')" :buttonFunction="deleteEntity" />

  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
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

// refs
const collectionMenu = ref(null);

// computed
// Checks if content can be selected.
const canSelectContent = computed(() => {
  const entityId = collectionStore.selectedCollection.id;
  return entityId in stage.expandedEntities && stage.entityDataIds.length;
});

// Returns the collection state flags.
const collectionStateFlags = computed(() => {
  const entity = collectionStore.selectedCollection;
  if (!entity) return {
    has_untracked: false,
    has_modified: false,
    has_outdated: false,
    has_rebuildable: false
  };
  
  return entity.collectionStateFlags || {
    has_untracked: false,
    has_modified: false,
    has_outdated: false,
    has_rebuildable: false
  };
});

// methods
// Opens the workflow selection modal.
const addWorkflow = () => {
  modals.setModalVisibility('selectWorkflowModal', true);
  menu.hideContextMenu();
};

// Copies the entity path to clipboard.
const copyEntityPath = async () => {
  let entity = collectionStore.selectedCollection;
  let entityDir = entity.file_path;
  entityDir = entityDir.replace(/\\/g, '/');
  FSService.MakeDirs(entityDir);
  await Clipboard.SetText(entityDir);
  notificationStore.addNotification('Path copied to clipboard', "", "success");
  menu.hideContextMenu();
};

// Opens the create collection modal.
const createEntity = () => {
  stage.expandEntity(collectionStore.selectedCollection);
  modals.setModalVisibility('createCollectionModal', true);
  menu.hideContextMenu();
};

// Opens the add web link modal.
const createLink = () => {
  stage.expandEntity(collectionStore.selectedCollection);
  modals.setModalVisibility('addWebLinkModal', true);
  menu.hideContextMenu();
};

// Opens the select app modal to create an asset.
const createTask = () => {
  stage.expandEntity(collectionStore.selectedCollection);
  modals.setModalVisibility('selectAppModal', true);
  menu.hideContextMenu();
};

// Deletes the selected collection.
const deleteEntity = async () => {
  let entity = collectionStore.selectedCollection;
  panes.setPaneVisibility('projectDetails', true);
  CollectionService.DeleteCollection(projectStore.activeProject.uri, entity.id)
    .then(async () => {
      stage.markedItems = [];
      collectionStore.selectedCollection = null;
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      console.error(error);
    });
  let longMessage = `Collection of name: ${entity.name} was moved to Trash.`;
  notificationStore.addNotification("Collection moved to Trash.", longMessage, "success", true);
  menu.hideContextMenu();
};

// Opens the edit collection modal.
const editEntity = () => {
  modals.setModalVisibility('editCollectionModal', true);
  menu.hideContextMenu();
};

// Frees up space by deleting the entity's working files.
const freeUpSpace = async () => {
  let entity = collectionStore.selectedCollection;
  let entityDir = entity.file_path.replace(/\\/g, '/');
  await FSService.DeleteFolder(entityDir)
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
      notificationStore.errorNotification("Could not determine entity directory", "");
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
      const message = successCount === 1 ? "1 item imported successfully" : `${successCount} items imported successfully`;
      notificationStore.addNotification(message, "", "success");
    }

    if (failureCount > 0) {
      const message = failureCount === 1 ? "1 item failed to import" : `${failureCount} items failed to import`;
      notificationStore.errorNotification(message, errors.join("\n"));
    }

    if (successCount > 0) {
      stage.expandEntity(collectionStore.selectedCollection);
      emitter.emit('refresh-browser');
    }
  } catch (error) {
    notificationStore.errorNotification("Error importing items", error.message || error);
  } finally {
    stage.operationActive = false;
    menu.hideContextMenu();
  }
};

// Prepares and shows the create checkpoints modal.
const prepCreateCheckpointsModal = () => {
  const entity = collectionStore.selectedCollection;
  trayStates.createMultipleCheckpointsEntityPath = entity.entity_path;
  modals.setModalVisibility('createMultipleCheckpointsModal', true);
  menu.hideContextMenu();
};

// Prepares and shows the free up space confirmation modal.
const prepFreeUpSpacePopUpModal = () => {
  trayStates.popUpModalTitle = "Free Up Entity Space";
  trayStates.popUpModalMessage = "Are you sure you want to delete this entity working files? This will permanently remove all uncheckpointed resources and all entity outputs. Please confirm if you wish to proceed.";
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpSpace;
  modals.setModalVisibility('popUpModal', true);
  menu.hideContextMenu();
};

// Prepares and shows the revert contents confirmation modal.
const prepRevertContentsPopUpModal = () => {
  trayStates.popUpModalIcon = 'revert';
  trayStates.popUpModalTitle = "Revert Contents";
  trayStates.popUpModalMessage = "All modified contents within this collection will be reverted to their last saved state. Are you sure you want to continue?";
  trayStates.popUpModalFunction = revertContents;
  modals.setModalVisibility('popUpModal', true);
  menu.hideContextMenu();
};

// Rebuilds the collection contents.
const rebuildCollection = () => {
  menu.hideContextMenu();
  let entity = collectionStore.selectedCollection;
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  CollectionService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, entity.id)
    .then(() => {
      assetStore.refreshEntityFilesStatus(entity.id);
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      console.error(error);
    });
  notificationStore.canCancel = false;
};

// Emits event to rename the collection.
const renameEntity = () => {
  emitter.emit('renameEntity');
  menu.hideContextMenu();
};

// Reverts all modified contents in the collection.
const revertContents = async () => {
  modals.setModalVisibility('popUpModal', false);
  
  const entity = collectionStore.selectedCollection;
  if (!entity) return;
  
  const collectionId = entity.id;
  await collectionStore.reloadItemsForCheckpoint(collectionId, null);
  const filteredPaths = assetStore.modifiedAssets.modified.map(asset => asset.task_path);
  
  if (filteredPaths.length === 0) {
    notificationStore.addNotification("No modified contents found in this collection", "", "info");
    return;
  }
  
  try {
    await CheckpointService.RevertTaskPaths(
      projectStore.activeProject.uri, 
      projectStore.getActiveProjectUrl, 
      filteredPaths
    );
    
    assetStore.modifiedAssets.modified = assetStore.modifiedAssets.modified.filter(
      (item) => !filteredPaths.includes(item.task_path)
    );
    
    emitter.emit('refresh-browser');
    
    const message = filteredPaths.length === 1 
      ? "1 item reverted successfully" 
      : `${filteredPaths.length} items reverted successfully`;
    notificationStore.addNotification(message, "", "success");
  } catch (error) {
    notificationStore.errorNotification("Failed to revert contents", error);
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
  stage.markedItems = stage.entityDataIds;
  menu.hideContextMenu();
};

// Updates outdated contents in the collection.
const updateContents = async () => {
  menu.hideContextMenu();
  
  const entity = collectionStore.selectedCollection;
  if (!entity) return;
  
  const entityPath = entity.entity_path;
  const outdatedTasksPath = assetStore.outdatedAssetsPath;
  const entityOutdatedPaths = outdatedTasksPath.filter(taskPath => taskPath.startsWith(entityPath));
  
  if (entityOutdatedPaths.length === 0) {
    notificationStore.addNotification("No outdated contents found in this collection", "", "info");
    return;
  }
  
  try {
    notificationStore.cancleFunction = SyncService.CancelSync;
    notificationStore.canCancel = true;
    
    await CheckpointService.RevertTaskPaths(
      projectStore.activeProject.uri, 
      projectStore.getActiveProjectUrl, 
      entityOutdatedPaths
    );
    
    assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(
      taskPath => !entityOutdatedPaths.includes(taskPath)
    );
    
    emitter.emit('refresh-browser');
    
    const message = entityOutdatedPaths.length === 1 
      ? "1 item updated successfully" 
      : `${entityOutdatedPaths.length} items updated successfully`;
    notificationStore.addNotification(message, "", "success");
  } catch (error) {
    notificationStore.errorNotification("Failed to update contents", error);
    console.error(error);
  } finally {
    notificationStore.canCancel = false;
  }
};

// Opens the upload items modal for web platform.
const uploadItems = () => {
  stage.expandEntity(collectionStore.selectedCollection);
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






