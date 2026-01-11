<template>
  <div ref="collectionMenu" class="filter-menu-container">

    <!-- Create -->
     <ActionButton :icon="getAppIcon('file-plus')" :showLabel="true" :fullWidth="true" label="Add Asset"
      v-if="templateStore.getTemplates.length && userStore.canDo('create_task')" :buttonFunction="createTask" />

    <ActionButton :icon="getAppIcon('folder-plus')" :showLabel="true" :fullWidth="true" label="Add Collection"
      v-if="userStore.canDo('create_entity')" :buttonFunction="createEntity" />

    <ActionButton :icon="getAppIcon('workflow-plus')" :showLabel="true" :fullWidth="true" label="Add Workflow"
      v-if="workflowStore.workflows.length && userStore.canDo('create_task')" :buttonFunction="addWorkflow" />

    <ActionButton :icon="getAppIcon('arrow-down-ramp')" :showLabel="true" :fullWidth="true" label="Import Items"
      v-if="!platformStore.isWeb && userStore.canDo('create_task')" :buttonFunction="importItems" />

    <ActionButton :icon="getAppIcon('arrow-up-ramp')" :showLabel="true" :fullWidth="true" label="Upload Items"
      v-if="platformStore.isWeb && userStore.canDo('create_task')" :buttonFunction="uploadItems" />

    <span v-if="userStore.canDo('create_entity') && !platformStore.isWeb" class="menu-divider"></span>

    <!-- Reveal in Explorer -->
    <span v-if="!platformStore.isWeb" class="horizontal-flex">
      <ActionButton :icon="getAppIcon('folder-arrow-up-right')" :showLabel="true" :fullWidth="true" label="Show in Explorer"
        :buttonFunction="revealInExplorer" />
      <ActionButton :icon="getAppIcon('copy')" :showLabel="false" :fullWidth="false" @click="copyDirectoryPath()"
        v-tooltip="'Copy Path'" />
    </span>

    <!-- Relocate Working Directory -->
    <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('folder-arrow-in')" :showLabel="true" :fullWidth="true" label="Relocate"
      :buttonFunction="relocateWorkingDirectory" />

    <!-- Rebuild -->
    <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('jigsaw')" :showLabel="true" :fullWidth="true" label="Build Project"
      :buttonFunction="rebuildAll" />

    <!-- Free space -->
    <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('broom')" :showLabel="true" :fullWidth="true" label="Free Up space"
      :buttonFunction="prepFreeUpSpacePopUpModal" />

    <!-- Clear Trash -->
    <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true" label="Empty Trash"
      :buttonFunction="prepEmptyTrashPopUpModal" />


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
import emitter from '@/lib/mitt';

// services
import { CollectionService, AssetService } from "@/services";
import { TrashService } from "@/services";

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
import { useTemplateStore } from '@/stores/template';
import { useWorkflowStore } from '@/stores/workflow';
import { usePlatformStore } from '@/stores/platform';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import { FSService, DialogService, ProjectService } from '@/services';
import { SyncService } from '@/services';
import { Clipboard } from '@wailsio/runtime';

// states/stores
const workflowStore = useWorkflowStore();
const trayStates = useTrayStates();
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
const commonStore = useCommonStore();
const templateStore = useTemplateStore();
const platformStore = usePlatformStore();

// refs
const collectionMenu = ref(null);

const renameEntity = () => {
  modals.setModalVisibility('renameEntityModal', true);
  menu.hideContextMenu();
};

const viewCollectionDetails = () => {
  panes.setPaneVisibility('collectionDetails', true)
  menu.hideContextMenu();
};

const revealInExplorer = async () => {

  menu.hideContextMenu();
  const isNavigated = commonStore.navigatorMode;
  let project = projectStore.getActiveProject;

  if(!isNavigated){
    await FSService.MakeDirs(project.working_directory)
    FSService.RevealInExplorer(project.working_directory)
  } else {
    let path = collectionStore.navigatedCollection?.type === 'entity' 
      ? collectionStore.navigatedCollection.entity_path
      : collectionStore.navigatedCollection.item_path;

    const trimmedPath = path.endsWith('/') ? path.slice(0, -1) : path;
    let explorerPath = `${project.working_directory}${trimmedPath}`
    explorerPath = explorerPath.replace(/\\/g, '/');

    await FSService.MakeDirs(explorerPath)
    FSService.RevealInExplorer(explorerPath)
  }
};

const copyDirectoryPath = async () => {

  const isNavigated = commonStore.navigatorMode;
  let project = projectStore.getActiveProject;


  if(!isNavigated){
    let projectDir = project.working_directory;
    projectDir = projectDir.replace(/\\/g, '/');
    FSService.MakeDirs(projectDir);
    await Clipboard.SetText(projectDir);
  } else {

    let path = collectionStore.navigatedCollection?.type === 'entity' 
      ? collectionStore.navigatedCollection.entity_path
      : collectionStore.navigatedCollection.item_path;

    let explorerPath = `${project.working_directory}${path}`
    explorerPath = explorerPath.replace(/\\/g, '/');
    await Clipboard.SetText(explorerPath);
  }

  const message = 'Path copied to clipboard';
  notificationStore.addNotification(message, "", "success");
  menu.hideContextMenu();
};

const relocateWorkingDirectory = async () => {
  menu.hideContextMenu();
  
  const project = projectStore.getActiveProject;
  const currentWorkingDir = project.working_directory;
  
  try {
    // Open folder dialog to select new working directory
    const result = await DialogService.SelectFolderDialog("Select New Working Directory");
    
    if (!result) {
      return; // User cancelled
    }
    
    let newWorkingDir = result.replace(/\\/g, '/');
    
    // Confirm with user
    trayStates.popUpModalTitle = 'Relocate Working Directory?';
    trayStates.popUpModalMessage = `Change working directory from:\n${currentWorkingDir}\n\nTo:\n${newWorkingDir}\n\nNote: Files will NOT be moved. Only the path will be updated.`;
    trayStates.popUpModalIcon = 'folder';
    trayStates.popUpModalFunction = async () => {
      try {
        stage.operationActive = true;
        
        // Update working directory
        await ProjectService.UpdateWorkingDirectory(
          project.has_remote ? projectStore.getActiveProjectUrl : project.uri,
          projectStore.selectedStudio.name,
          newWorkingDir
        );
        
        // Update the project in memory
        project.working_directory = newWorkingDir;
        
        // Refresh project list
        await projectStore.refreshProjects();
        
        notificationStore.addNotification(
          'Working directory updated',
          `New location: ${newWorkingDir}`,
          'success',
          false
        );
        
      } catch (error) {
        notificationStore.errorNotification('Error updating working directory', error);
      } finally {
        stage.operationActive = false;
        modals.setModalVisibility('popUpModal', false);
        emitter.emit('refresh-browser');
      }
    };
    
    modals.setModalVisibility('popUpModal', true);
    
  } catch (error) {
    notificationStore.errorNotification('Error selecting directory', error);
  }
};

const createEntity = () => {
  modals.setModalVisibility('createCollectionModal', true);
  menu.hideContextMenu();
};

const addWorkflow = () => {
  modals.setModalVisibility('selectWorkflowModal', true);
  menu.hideContextMenu();
};

const createTask = () => {
  modals.setModalVisibility('selectAppModal', true);
  menu.hideContextMenu();
};

const createResources = () => {
  modals.setModalVisibility('addResourcesModal', true);
  menu.hideContextMenu();
};

const uploadItems = () => {
  // TODO: Implement web upload functionality
  modals.setModalVisibility('uploadItemsModal', true);
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

    // Get current directory for copying
    const currentDirectory = getCurrentDirectory();
    if (!currentDirectory) {
      notificationStore.errorNotification("Could not determine current directory", "");
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

    // Refresh the view to show imported items
    if (successCount > 0) {
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
  // Check if we're in navigator mode (inside an entity/folder)
  if (commonStore.navigatorMode && collectionStore.navigatedCollection) {
    const navigated = collectionStore.navigatedCollection;
    // Return the file path of the current entity or folder
    return navigated.file_path;
  }
  
  // If at project root
  return projectStore.activeProject?.working_directory;
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

const freeUpProjectSpace = async () => {
  let project = projectStore.getActiveProject;
  console.log(project.working_directory)
  await FSService.DeleteFolder(project.working_directory)
    .then((response) => {
      projectStore.refreshProjects()
      
      AssetService.GetAssetsStates(project.uri, project.working_directory, project.ignore_list).then((assetsStates)=>{
        console.log(assetsStates)
        assetStore.modifiedAssetsPath = assetsStates.modified
        assetStore.outdatedAssetsPath = assetsStates.outdated
        assetStore.rebuildableAssetsPath = assetsStates.rebuildable
      })

      if (projectStore.activeProject.id == project.id) {
        trayStates.$reset();
      }

      emitter.emit('refresh-browser');

    })
    .catch((error) => {
      console.error(error);
    });
  modals.disableAllModals();
};

const freeUpEntitySpace = async () => {
  let entity = collectionStore.navigatedCollection;
  let entityDir = entity.file_path.replace(/\\/g, '/');
  await FSService.DeleteFolder(entityDir)
    .then((response) => {
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      console.error(error);
    });
  modals.disableAllModals();
};

const rebuildAll = async () => { 
  menu.hideContextMenu();

	const path = collectionStore.navigatedCollection?.entity_path;
	const navigatedEntityId = collectionStore.navigatedCollection?.id;
	const rebuildableTasksPath = assetStore.rebuildableAssetsPath;

	notificationStore.cancleFunction = SyncService.CancelSync;
	notificationStore.canCancel = true;

	await CollectionService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, navigatedEntityId)
		.then((data) => {

			if(path){
				assetStore.rebuildableAssetsPath = rebuildableTasksPath.filter(item => !item.startsWith(path))
			} else {
				assetStore.rebuildableAssetsPath = [];
			}

			softRefresh();
		}).catch(async (error) => {
			notificationStore.errorNotification("Error Rebuilding All", error)
		})
};

const prepFreeUpSpacePopUpModal = () => {
  menu.hideContextMenu();
  let project = projectStore.getActiveProject;
  if(commonStore.navigatorMode){
    const navigatedEntity = collectionStore.navigatedCollection;
    trayStates.popUpModalTitle = `Delete the files in \"${navigatedEntity.name}\"? `;
    trayStates.popUpModalMessage = `This will clear the contents of \"${navigatedEntity.name}\". Please save your checkpoints before proceeding to avoid losing your work`;
    trayStates.popUpModalFunction = freeUpEntitySpace;
  } else {
    trayStates.popUpModalTitle = `Delete the files in \"${project.name}\"? `;
    trayStates.popUpModalMessage = "This will clear the current project directory. Please save your checkpoints before proceeding to avoid losing your work";
    trayStates.popUpModalFunction = freeUpProjectSpace;
  }
  trayStates.popUpModalIcon = 'broom';
  modals.setModalVisibility('popUpModal', true);
};

const prepEmptyTrashPopUpModal = () => {
  menu.hideContextMenu();
	trayStates.popUpModalIcon = 'trash'
	trayStates.popUpModalTitle = "Empty Trash";
	trayStates.popUpModalMessage = "This will irreversibly delete all items in trash. Continue?";
	trayStates.popUpModalFunction = emptyTrash;
	modals.setModalVisibility('popUpModal', true);
};

const emptyTrash = async () => {
	await ProjectService.Purge(projectStore.activeProject.uri)
		.then(() => {
			trayStates.trashables = [];
			modals.disableAllModals();
		}).catch((error) => {
			console.error(error.message)
			notificationStore.addNotification(
				"Error Syncing Data",
				error.message,
				"error",
				false
			)
			modals.disableAllModals();
		})
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

.horizontal-flex{
  padding: 0;
}

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

