<template>
  <div ref="collectionMenu" class="filter-menu-container">

    <ActionButton :icon="getAppIcon('edit')" :showLabel="true" :fullWidth="true" label="Rename Collection"
      v-if="userStore.canDo('update_entity')" :buttonFunction="renameEntity" />

    <ActionButton :icon="getAppIcon('parameters')" :showLabel="true" :fullWidth="true" label="Edit Collection"
      v-if="userStore.canDo('update_entity')" :buttonFunction="editEntity" />

    <ActionButton v-if="canSelectContent" :icon="getAppIcon('checkbox-selected')" :showLabel="true" :fullWidth="true"
      label="Select Content" :buttonFunction="selectContent" />

    <span v-if="userStore.canDo('update_entity')" class="menu-divider"></span>

    <!-- Create -->
    <ActionButton :icon="getAppIcon('brush-plus')" :showLabel="true" :fullWidth="true" label="Add Asset"
      v-if="templateStore.getTemplates.length && (userStore.canDo('create_task') || collectionStore.selectedCollection.can_modify)" :buttonFunction="createTask" />

    <ActionButton :icon="getAppIcon('folder-plus')" :showLabel="true" :fullWidth="true" label="Add Collection"
      v-if="userStore.canDo('create_entity') || collectionStore.selectedCollection.can_modify" :buttonFunction="createEntity" />

    <ActionButton :icon="getAppIcon('workflow-plus')" :showLabel="true" :fullWidth="true" label="Add Workflow"
      v-if="workflowStore.workflows.length && userStore.canDo('create_task')" :buttonFunction="addWorkflow" />

    

    <!-- <ActionButton :icon="getAppIcon('website-link')" :showLabel="true" :fullWidth="true" label="New Link"
      v-if="userStore.canDo('create_task') || collectionStore.selectedCollection.can_modify" :buttonFunction="createLink" /> -->

    <ActionButton :icon="getAppIcon('arrow-down-ramp')" :showLabel="true" :fullWidth="true" label="Import Items"
      v-if="userStore.canDo('create_task')" :buttonFunction="importItems" />


    <span v-if="userStore.canDo('update_entity') || collectionStore.selectedCollection.can_modify" class="menu-divider"></span>

    <!-- Reveal in Explorer -->
    <span class="horizontal-flex">
      <ActionButton :icon="getAppIcon('folder-arrow-up-right')" :showLabel="true" :fullWidth="true" label="Show in Explorer"
        :buttonFunction="revealInExplorer" />
      <ActionButton :icon="getAppIcon('copy')" :showLabel="false" :fullWidth="false" @click="copyEntityPath('entity')"
        v-tooltip="'Copy Path'" />
    </span>

    <!-- Revert Contents -->
    <ActionButton v-if="hasModifiedContents" :noFilter="true" :icon="getAppIcon('revert-alert')" :showLabel="true" :fullWidth="true" 
      label="Revert Contents" :buttonFunction="prepRevertContentsPopUpModal" />

    <!-- Rebuild -->
    <ActionButton :icon="getAppIcon('jigsaw')" :showLabel="true" :fullWidth="true" label="Rebuild Collection"
      :buttonFunction="rebuildCollection" />

    <!-- Free space -->
    <ActionButton :icon="getAppIcon('broom')" :showLabel="true" :fullWidth="true" label="Free Up space"
      :buttonFunction="prepFreeUpSpacePopUpModal" />

    <!-- Delete Task -->
    <ActionButton :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true" label="Delete Collection"
      v-if="userStore.canDo('delete_entity')" :buttonFunction="deleteEntity" />

  </div>

</template>

<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};
// imports
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import utils from '@/services/utils';

// services
import { EntityService, SyncService, TaskService, TrashService, CheckpointService } from "@/../bindings/clustta/services";

// states/store imports
import { useTrayStates } from '@/stores/TrayStates';
import { useMenu } from '@/stores/menu';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useUserStore } from '@/stores/users';
import { useModalStore } from '@/stores/modals';
import { useCollectionStore } from '@/stores/collections';
import { useAssetStore } from '@/stores/assets';
import { useCommonStore } from '@/stores/common';
import { useProjectStore } from '@/stores/projects';
import { useWorkflowStore } from '@/stores/workflow';
import { useTemplateStore } from '@/stores/template';
import emitter from '@/lib/mitt';
// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import { ClipboardService, FSService, DialogService } from '@/../bindings/clustta/services/index';

// states/stores
const trayStates = useTrayStates();
const templateStore = useTemplateStore();
const userStore = useUserStore();
const menu = useMenu();
const panes = usePaneStore();
const stage = useStageStore();
const modals = useDesktopModalStore();
const modalStore = useModalStore();
const notificationStore = useNotificationStore();
const collectionStore = useCollectionStore();
const assetStore = useAssetStore();
const projectStore = useProjectStore();
const workflowStore = useWorkflowStore();
const commonStore = useCommonStore();

// computed
const canSelectContent = computed(() => {
  const entityId = collectionStore.selectedCollection.id;
  return entityId in stage.expandedEntities && stage.entityDataIds.length
})

const hasModifiedContents = computed(() => {
  const entity = collectionStore.selectedCollection;
  if (!entity) return false;
  
  const entityPath = entity.entity_path;
  const modifiedTasksPath = assetStore.modifiedAssetsPath;
  
  if (!entityPath || !modifiedTasksPath.length) return false;
  
  // Filter tasks that are within this entity's path recursively
  const filteredPaths = modifiedTasksPath.filter(taskPath => taskPath.startsWith(entityPath));
  
  return filteredPaths.length > 0;
})

// refs
const collectionMenu = ref(null);

const editEntity = () => {
  modals.setModalVisibility('editCollectionModal', true);
  menu.hideContextMenu();
};

const renameEntity = () => {
  emitter.emit('renameEntity');
  menu.hideContextMenu();
};

const selectContent = () => {
  stage.markedItems = stage.entityDataIds;
  menu.hideContextMenu();
};

const revealInExplorer = async () => {
  await FSService.MakeDirs(collectionStore.selectedCollection.file_path)
  FSService.RevealInExplorer(collectionStore.selectedCollection.file_path)
  menu.hideContextMenu();
};

const createEntity = () => {
  stage.expandEntity(collectionStore.selectedCollection);
  modals.setModalVisibility('createCollectionModal', true);
  menu.hideContextMenu();
};

const createTask = () => {
  stage.expandEntity(collectionStore.selectedCollection);
  modals.setModalVisibility('selectAppModal', true);
  menu.hideContextMenu();
};

const createLink = () => {
  stage.expandEntity(collectionStore.selectedCollection);
  modals.setModalVisibility('addWebLinkModal', true);
  menu.hideContextMenu();
};

const addWorkflow = () => {
  modals.setModalVisibility('selectWorkflowModal', true);
  menu.hideContextMenu();
};

const importItems = async () => {
  try {
    // Show items picker dialog (files and folders)
    const selectedPaths = await DialogService.SelectFilesDialog();
    if (!selectedPaths || selectedPaths.length === 0) {
      menu.hideContextMenu();
      return; // User cancelled or no items selected
    }

    // Get current entity directory for copying
    const currentDirectory = getCurrentDirectory();
    if (!currentDirectory) {
      notificationStore.errorNotification("Could not determine entity directory", "");
      menu.hideContextMenu();
      return;
    }

    await FSService.MakeDirs(currentDirectory);

    // Show operation in progress
    stage.operationActive = true;
    
    let successCount = 0;
    let failureCount = 0;
    const errors = [];

    // Process each selected path
    for (const sourcePath of selectedPaths) {
      try {
        const isFile = await FSService.IsFile(sourcePath);
        const itemName = await FSService.BaseName(sourcePath);
        
        // Generate unique destination path
        const destinationPath = await generateUniqueDestinationPath(currentDirectory, itemName);
        
        if (isFile) {
          // Copy file
          await FSService.DuplicateFile(sourcePath, destinationPath);
        } else {
          // Copy folder
          await FSService.DuplicateFolder(sourcePath, destinationPath);
        }
        
        successCount++;
      } catch (error) {
        failureCount++;
        const itemName = await FSService.BaseName(sourcePath).catch(() => sourcePath);
        errors.push(`${itemName}: ${error.message || error}`);
      }
    }

    // Show results
    if (successCount > 0) {
      const message = successCount === 1 
        ? "1 item imported successfully" 
        : `${successCount} items imported successfully`;
      notificationStore.addNotification(message, "", "success");
    }

    if (failureCount > 0) {
      const message = failureCount === 1 
        ? "1 item failed to import" 
        : `${failureCount} items failed to import`;
      notificationStore.errorNotification(message, errors.join("\n"));
    }

    // Expand entity and refresh the view to show imported items
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

const getCurrentDirectory = () => {
  // Return the file path of the currently selected entity
  return collectionStore.selectedCollection?.file_path;
};

const generateUniqueDestinationPath = async (directory, fileName) => {
  const originalPath = await FSService.JoinPath(directory, fileName);
  
  // Check if file/folder already exists
  const exists = await FSService.Exists(originalPath);
  if (!exists) {
    return originalPath;
  }
  
  // Generate unique name with counter
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
    if (!pathExists) {
      return newPath;
    }
    counter++;
  } while (counter < 100); // Safety limit
  
  // Fallback with timestamp if we hit the limit
  const timestamp = Date.now();
  const timestampFileName = `${baseName}_${timestamp}${extension}`;
  return await FSService.JoinPath(directory, timestampFileName);
};

// methods
const deleteEntity = async () => {
  let entity = collectionStore.selectedCollection;
  panes.setPaneVisibility('projectDetails', true);
  EntityService.DeleteEntity(projectStore.activeProject.uri, entity.id)
    .then(async (response) => {
      stage.markedItems = [];
      collectionStore.selectedCollection = null;
      emitter.emit('refresh-browser')
    })
    .catch((error) => {
      console.error(error);
    });
  let longMessage = `Collection of name: ${entity.name} was moved to Trash.`
  notificationStore.addNotification("Collection moved to Trash.", longMessage, "success", true);
  menu.hideContextMenu();
};

const freeUpSpace = async () => {
  let entity = collectionStore.selectedCollection;
  let entityDir = entity.file_path.replace(/\\/g, '/');
  await FSService.DeleteFolder(entityDir)
    .then((response) => {
      emitter.emit('refresh-browser');
      
      // let project = projectStore.activeProject
      // assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(taskPath => !taskPath.startsWith(collectionStore.selectedCollection.entity_path))
      // assetStore.modifiedAssetsPath = assetStore.modifiedAssetsPath.filter(taskPath => !taskPath.startsWith(collectionStore.selectedCollection.entity_path))
      // TaskService.GetAssetsStates(project.uri, project.working_directory, project.ignore_list).then((assetsStates)=>{
      //   assetStore.modifiedAssetsPath = assetsStates.modified
      //   assetStore.outdatedAssetsPath = assetsStates.outdated
      //   assetStore.rebuildableAssetsPath = assetsStates.rebuildable
      // })

    })
    .catch((error) => {
      console.error(error);
    });
  modals.disableAllModals();
  menu.hideContextMenu();

};

const rebuildCollection = () => {
  menu.hideContextMenu();
  let entity = collectionStore.selectedCollection;
  notificationStore.cancleFunction = SyncService.CancelSync
  notificationStore.canCancel = true
  EntityService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, entity.id)
    .then((data) => {
      assetStore.refreshEntityFilesStatus(entity.id)
      emitter.emit('refresh-browser');
    }).catch(error => {
      console.log(error)
    })
  notificationStore.canCancel = false;
};

const copyEntityPath = async () => {
  let entity = collectionStore.selectedCollection;
  let entityDir = entity.file_path;
  entityDir = entityDir.replace(/\\/g, '/');
  await ClipboardService.WriteText(entityDir);
  const message = 'Path copied to clipboard';
  notificationStore.addNotification(message, "", "success");
  menu.hideContextMenu();
};

const prepRevertContentsPopUpModal = () => {
  trayStates.popUpModalIcon = 'revert';
  trayStates.popUpModalTitle = "Revert Contents";
  trayStates.popUpModalMessage = "All modified contents within this collection will be reverted to their last saved state. Are you sure you want to continue?";
  trayStates.popUpModalFunction = revertContents;
  modals.setModalVisibility('popUpModal', true);
  menu.hideContextMenu();
};

const revertContents = async () => {
  modals.setModalVisibility('popUpModal', false);
  
  const entity = collectionStore.selectedCollection;
  if (!entity) return;
  
  const entityPath = entity.entity_path;
  const modifiedTasksPath = assetStore.modifiedAssetsPath;
  
  // Filter only the modified tasks within this entity's path recursively
  const entityModifiedPaths = modifiedTasksPath.filter(taskPath => taskPath.startsWith(entityPath));
  
  if (entityModifiedPaths.length === 0) {
    notificationStore.addNotification("No modified contents found in this collection", "", "info");
    return;
  }
  
  try {
    await CheckpointService.RevertTaskPaths(
      projectStore.activeProject.uri, 
      projectStore.getActiveProjectUrl, 
      entityModifiedPaths
    );
    
    // Update the global modified tasks list by removing the reverted paths
    assetStore.modifiedAssetsPath = assetStore.modifiedAssetsPath.filter(
      taskPath => !entityModifiedPaths.includes(taskPath)
    );
    
    emitter.emit('refresh-browser');
    
    const message = entityModifiedPaths.length === 1 
      ? "1 item reverted successfully" 
      : `${entityModifiedPaths.length} items reverted successfully`;
    notificationStore.addNotification(message, "", "success");
    
  } catch (error) {
    notificationStore.errorNotification("Failed to revert contents", error);
    console.error(error);
  }
};

const prepFreeUpSpacePopUpModal = () => {
  trayStates.popUpModalTitle = "Free Up Entity Space";
  trayStates.popUpModalMessage = "Are you sure you want to delete this entity working files? This will permanently remove all uncheckpointed resources and all entity outputs. Please confirm if you wish to proceed.";
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpSpace;
  modals.setModalVisibility('popUpModal', true);
  menu.hideContextMenu();
};

const viewCheckPoints = () => {
  modalStore.triggerMenuItem('collectionMenu', 'CheckPoints');
};



// onMounted hook
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

.entity-item-menu-container {
  z-index: 10;
  display: flex;
  /* opacity: 0;
  visibility : hidden;
  position: absolute; */
  top: 0;
  left: 0;
  flex-direction: column;
  color: white;
  align-items: center;
  gap: .3rem;
  padding: .6rem;
  box-sizing: border-box;
  width: max-content;
  width: 250px;
  height: max-content;
  border-radius: 16px;
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--light-steel);

}

.entity-item-menu-visible {
  /* display: flex; */
  opacity: 1;
  visibility: visible;
}
</style>





