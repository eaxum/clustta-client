<template>
  <div ref="collectionMenu" class="filter-menu-container">

    <!-- Create -->
     <ActionButton :icon="getAppIcon('file-plus')" :showLabel="true" :fullWidth="true" :label="$t('menus.addAsset')"
      v-if="templateStore.getTemplates.length && userStore.canDo('create_task')" :buttonFunction="createTask" />

    <ActionButton :icon="getAppIcon('folder-plus')" :showLabel="true" :fullWidth="true" :label="$t('menus.addCollection')"
      v-if="userStore.canDo('create_entity')" :buttonFunction="createEntity" />

    <ActionButton :icon="getAppIcon('workflow-plus')" :showLabel="true" :fullWidth="true" :label="$t('menus.addWorkflow')"
      v-if="workflowStore.workflows.length && userStore.canDo('create_task')" :buttonFunction="addWorkflow" />

    <ActionButton :icon="getAppIcon('arrow-down-ramp')" :showLabel="true" :fullWidth="true" :label="$t('modals.importItems')"
      v-if="!platformStore.isWeb && userStore.canDo('create_task')" :buttonFunction="importItems" />

    <ActionButton :icon="getAppIcon('arrow-up-ramp')" :showLabel="true" :fullWidth="true" :label="$t('menus.uploadItems')"
      v-if="platformStore.isWeb && userStore.canDo('create_task')" :buttonFunction="uploadItems" />

    <ActionButton :icon="getAppIcon('clipboard')" :showLabel="true" :fullWidth="true" :label="$t('common.paste')"
      v-if="hasClipboardItems && userStore.canDo('update_entity')" :buttonFunction="pasteItems" />

    <span v-if="userStore.canDo('create_entity') && !platformStore.isWeb" class="menu-divider"></span>

    <!-- Reveal in Explorer -->
    <span v-if="!platformStore.isWeb" class="horizontal-flex">
      <ActionButton :icon="getAppIcon('folder-arrow-up-right')" :showLabel="true" :fullWidth="true" :label="$t('common.showInExplorer')"
        :buttonFunction="revealInExplorer" />
      <ActionButton :icon="getAppIcon('copy')" :showLabel="false" :fullWidth="false" @click="copyDirectoryPath()"
        v-tooltip="$t('common.copyPath')" />
    </span>

    <!-- Relocate Working Directory -->
    <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('folder-arrow-in')" :showLabel="true" :fullWidth="true" :label="$t('menus.relocate')"
      :buttonFunction="relocateWorkingDirectory" />

    <!-- Rebuild -->
    <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('jigsaw')" :showLabel="true" :fullWidth="true" :label="$t('menus.buildProject')"
      :buttonFunction="rebuildAll" />

    <!-- Free space -->
    <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('broom')" :showLabel="true" :fullWidth="true" :label="$t('common.freeUpSpace')"
      :buttonFunction="prepFreeUpSpacePopUpModal" />

    <!-- Clear Trash -->
    <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true" :label="$t('common.emptyTrash')"
      :buttonFunction="prepEmptyTrashPopUpModal" />


  </div>

</template>

<script setup>
// imports
import { onBeforeUnmount, onMounted, ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { Clipboard } from '@wailsio/runtime';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// services
import { AssetService, CollectionService, DialogService, FSService, ProjectService, SyncService } from "@/services";

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTemplateStore } from '@/stores/template';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';
import { useWorkflowStore } from '@/stores/workflow';

const { t } = useI18n();
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const templateStore = useTemplateStore();
const trayStates = useTrayStates();
const userStore = useUserStore();
const workflowStore = useWorkflowStore();

// refs
const collectionMenu = ref(null);

// computed properties
// Returns true if there are items in the clipboard.
const hasClipboardItems = computed(() => stage.cutItems.length > 0 || stage.copiedItems.length > 0);

// methods
// Opens the workflow selection modal.
const addWorkflow = () => {
  modals.setModalVisibility('selectWorkflowModal', true);
  menu.hideContextMenu();
};

// Copies the current directory path to clipboard.
const copyDirectoryPath = async () => {
  const isNavigated = commonStore.navigatorMode;
  let project = projectStore.getActiveProject;

  if (!isNavigated) {
    let projectDir = project.working_directory;
    projectDir = projectDir.replace(/\\/g, '/');
    FSService.MakeDirs(projectDir);
    await Clipboard.SetText(projectDir);
  } else {
    let path = collectionStore.navigatedCollection?.type === 'entity' 
      ? collectionStore.navigatedCollection.entity_path
      : collectionStore.navigatedCollection.item_path;

    let explorerPath = `${project.working_directory}${path}`;
    explorerPath = explorerPath.replace(/\\/g, '/');
    await Clipboard.SetText(explorerPath);
  }

  notificationStore.addNotification(t('notifications.pathCopied'), "", "success");
  menu.hideContextMenu();
};

// Opens the create collection modal.
const createEntity = () => {
  modals.setModalVisibility('createCollectionModal', true);
  menu.hideContextMenu();
};

// Opens the create asset modal.
const createTask = () => {
  modals.setModalVisibility('selectAppModal', true);
  menu.hideContextMenu();
};

// Empties the project trash.
const emptyTrash = async () => {
  await ProjectService.Purge(projectStore.activeProject.uri)
    .then(() => {
      trayStates.trashables = [];
      modals.disableAllModals();
    })
    .catch((error) => {
      console.error(error.message);
      notificationStore.addNotification(t('notifications.errorSyncingData'), error.message, "error", false);
      modals.disableAllModals();
    });
};

// Frees up space by deleting the navigated entity's files.
const freeUpEntitySpace = async () => {
  let entity = collectionStore.navigatedCollection;
  let entityDir = entity.file_path.replace(/\\/g, '/');
  await FSService.DeleteFolder(entityDir)
    .then(() => {
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      console.error(error);
    });
  modals.disableAllModals();
};

// Frees up space by deleting the project's working directory.
const freeUpProjectSpace = async () => {
  let project = projectStore.getActiveProject;
  await FSService.DeleteFolder(project.working_directory)
    .then(() => {
      projectStore.refreshProjects();
      
      AssetService.GetAssetsStates(project.uri, project.working_directory, project.ignore_list).then((assetsStates) => {
        assetStore.modifiedAssetsPath = assetsStates.modified;
        assetStore.outdatedAssetsPath = assetsStates.outdated;
        assetStore.rebuildableAssetsPath = assetsStates.rebuildable;
      });

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

// Returns the current directory path based on navigation state.
const getCurrentDirectory = () => {
  if (commonStore.navigatorMode && collectionStore.navigatedCollection) {
    return collectionStore.navigatedCollection.file_path;
  }
  return projectStore.activeProject?.working_directory;
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
      notificationStore.addNotification(t('notifications.itemsImported', successCount), "", "success");
    }

    if (failureCount > 0) {
      notificationStore.errorNotification(t('notifications.itemsFailedImport', failureCount), errors.join("\n"));
    }

    if (successCount > 0) {
      emitter.emit('refresh-browser');
    }
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorImportingItems'), error.message || error);
  } finally {
    stage.operationActive = false;
    menu.hideContextMenu();
  }
};

// Prepares and shows the empty trash confirmation modal.
const prepEmptyTrashPopUpModal = () => {
  menu.hideContextMenu();
  trayStates.popUpModalIcon = 'trash';
  trayStates.popUpModalTitle = t('common.emptyTrash');
  trayStates.popUpModalMessage = t('confirmations.emptyTrash');
  trayStates.popUpModalFunction = emptyTrash;
  modals.setModalVisibility('popUpModal', true);
};

// Prepares and shows the free up space confirmation modal.
const prepFreeUpSpacePopUpModal = () => {
  menu.hideContextMenu();
  let project = projectStore.getActiveProject;
  if (commonStore.navigatorMode) {
    const navigatedEntity = collectionStore.navigatedCollection;
    trayStates.popUpModalTitle = t('confirmations.deleteFilesIn', { name: navigatedEntity.name });
    trayStates.popUpModalMessage = t('confirmations.clearContentsEntity', { name: navigatedEntity.name });
    trayStates.popUpModalFunction = freeUpEntitySpace;
  } else {
    trayStates.popUpModalTitle = t('confirmations.deleteFilesIn', { name: project.name });
    trayStates.popUpModalMessage = t('confirmations.clearContentsProject');
    trayStates.popUpModalFunction = freeUpProjectSpace;
  }
  trayStates.popUpModalIcon = 'broom';
  modals.setModalVisibility('popUpModal', true);
};

// Pastes clipboard items to the current location.
const pasteItems = async () => {
  menu.hideContextMenu();
  const result = await stage.pasteItems();
  if (result.needsRefresh) {
    emitter.emit('refresh-browser');
  }
};

// Rebuilds all assets in the current context.
const rebuildAll = async () => { 
  menu.hideContextMenu();
  const path = collectionStore.navigatedCollection?.entity_path;
  const navigatedEntityId = collectionStore.navigatedCollection?.id;
  const rebuildableTasksPath = assetStore.rebuildableAssetsPath;

  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;

  await CollectionService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, navigatedEntityId)
    .then(() => {
      if (path) {
        assetStore.rebuildableAssetsPath = rebuildableTasksPath.filter(item => !item.startsWith(path));
      } else {
        assetStore.rebuildableAssetsPath = [];
      }
      emitter.emit('refresh-browser');
    })
    .catch(async (error) => {
      notificationStore.errorNotification(t('notifications.errorRebuildingAll'), error);
    });
};

// Opens the relocate working directory dialog.
const relocateWorkingDirectory = async () => {
  menu.hideContextMenu();
  
  const project = projectStore.getActiveProject;
  const currentWorkingDir = project.working_directory;
  
  try {
    const result = await DialogService.SelectFolderDialog(t('menus.selectNewWorkingDirectory'));
    if (!result) return;
    
    let newWorkingDir = result.replace(/\\/g, '/');
    
    trayStates.popUpModalTitle = t('menus.relocateWorkingDirectory');
    trayStates.popUpModalMessage = t('confirmations.relocateWorkingDirectory', { from: currentWorkingDir, to: newWorkingDir });
    trayStates.popUpModalIcon = 'folder';
    trayStates.popUpModalFunction = async () => {
      try {
        stage.operationActive = true;
        await ProjectService.UpdateWorkingDirectory(
          project.has_remote ? projectStore.getActiveProjectUrl : project.uri,
          projectStore.selectedStudio.name,
          newWorkingDir
        );
        project.working_directory = newWorkingDir;
        await projectStore.refreshProjects();
        notificationStore.addNotification(t('notifications.workingDirUpdated'), `New location: ${newWorkingDir}`, 'success', false);
      } catch (error) {
        notificationStore.errorNotification(t('notifications.errorUpdatingDirectory'), error);
      } finally {
        stage.operationActive = false;
        modals.setModalVisibility('popUpModal', false);
        emitter.emit('refresh-browser');
      }
    };
    
    modals.setModalVisibility('popUpModal', true);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorSelectingDirectory'), error);
  }
};

// Reveals the current directory in the file explorer.
const revealInExplorer = async () => {
  menu.hideContextMenu();
  const isNavigated = commonStore.navigatorMode;
  let project = projectStore.getActiveProject;

  if (!isNavigated) {
    await FSService.MakeDirs(project.working_directory);
    FSService.RevealInExplorer(project.working_directory);
  } else {
    let path = collectionStore.navigatedCollection?.type === 'entity' 
      ? collectionStore.navigatedCollection.entity_path
      : collectionStore.navigatedCollection.item_path;

    const trimmedPath = path.endsWith('/') ? path.slice(0, -1) : path;
    let explorerPath = `${project.working_directory}${trimmedPath}`;
    explorerPath = explorerPath.replace(/\\/g, '/');

    await FSService.MakeDirs(explorerPath);
    FSService.RevealInExplorer(explorerPath);
  }
};

// Opens the upload items modal for web platform.
const uploadItems = () => {
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

.horizontal-flex {
  padding: 0;
}
</style>

