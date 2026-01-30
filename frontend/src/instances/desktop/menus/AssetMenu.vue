<template>
  <div ref="popUpMenu" class="filter-menu-container">

    <ActionButton v-if="!platformStore.isWeb && userStore.canDo('pull_chunk')" :icon="getAppIcon('launch')" :showLabel="true" :fullWidth="true"
      label="Open With" :buttonFunction="launchAssetWithCommand" />

    <span v-if="!platformStore.isWeb && userStore.canDo('pull_chunk')" class="menu-divider"></span>

    <ActionButton v-if="userStore.canDo('update_task')" :icon="getAppIcon('edit')" :showLabel="true" :fullWidth="true"
      label="Rename" :buttonFunction="renameAsset" />

    <ActionButton v-if="userStore.canDo('update_task')" :icon="getAppIcon('switches')" :showLabel="true"
      :fullWidth="true" label="Edit" :buttonFunction="editAsset" />

    <ActionButton v-if="userStore.canDo('create_task')" :icon="getAppIcon('duplicate')" :showLabel="true"
      :fullWidth="true" label="Duplicate" :buttonFunction="duplicateAsset" />

    <!-- Copy to Project -->
    <ActionButton v-if="!platformStore.isWeb && userStore.canDo('create_task') && canCopyToOtherProject" 
      :icon="getAppIcon('briefcase')" :showLabel="true"
      :fullWidth="true" label="Copy to Project" :buttonFunction="copyToProject" />

    <!-- Move to Collection -->
    <ActionButton v-if="!platformStore.isWeb && userStore.canDo('update_task')" 
      :icon="getAppIcon('folder-arrow-in')" :showLabel="true"
      :fullWidth="true" label="Move" :buttonFunction="moveToCollection" />

    <ActionButton v-if="!platformStore.isWeb && (asset.dependencies.length || asset.entity_dependencies.length)" :icon="getAppIcon('jigsaw')" :showLabel="true"
      :fullWidth="true" label="Build with dependencies" :buttonFunction="buildWithDependencies" />

    <ActionButton v-if="userStore.canDo('manage_dependencies')" :icon="getAppIcon('dependency')" :showLabel="true"
      :fullWidth="true" label="Dependency Graph" :buttonFunction="goToDependencyGraph" />

    <!-- Go to Location -->
    <ActionButton v-if="commonStore.viewSearchQuery || filtersActive" :icon="getAppIcon('file-search')" :showLabel="true" :fullWidth="true"
      label="Go to Asset" :buttonFunction="goToLocation" />

    <!-- Reveal in Explorer -->
    <span v-if="!platformStore.isWeb" class="horizontal-flex">
      <ActionButton :icon="getAppIcon('folder-arrow-up-right')" :showLabel="true" :fullWidth="true" label="Show in Explorer"
        :buttonFunction="revealInExplorer" />
      <ActionButton :icon="getAppIcon('copy')" :showLabel="false" :fullWidth="false" @click="copyAssetPath('asset')"
        v-tooltip="'Copy Path'" />
    </span>

    <!-- Extract Archive -->
    <ActionButton v-if="!platformStore.isWeb && isArchive" :icon="getAppIcon('unarchive')" :showLabel="true" :fullWidth="true" 
      label="Extract" :buttonFunction="extractArchive" />

    <!-- Checkpoints -->
    <ActionButton v-if="!platformStore.isWeb && isAssetModified" :noFilter="true" :icon="getAppIcon('revert')" :useAlert="true" :showLabel="true" :fullWidth="true"
      label="Revert File" :buttonFunction="revertAsset" />

    <span v-if="userStore.canDo('delete_task') || !isNotOnDisk" class="menu-divider"></span>

    <!-- Free space -->
    <ActionButton :icon="getAppIcon('broom')" v-if="!platformStore.isWeb && !isNotOnDisk" :showLabel="true" :fullWidth="true"
      label="Free Up space" :buttonFunction="prepFreeUpSpacePopUpModal" />

    <!-- Delete Task -->
    <ActionButton :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true" label="Delete"
      v-if="userStore.canDo('delete_task')" :buttonFunction="deleteAsset" />

  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { Clipboard } from '@wailsio/runtime';
import emitter from '@/lib/mitt';
import { isValidWeblink } from '@/lib/pointer';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// services
import { AssetService, CheckpointService, CollectionService, FSService } from "@/services";

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
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
const commonStore = useCommonStore();
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

// refs
const popUpMenu = ref(null);

// props
const props = defineProps({
  top: Number,
  left: Number,
});

const emit = defineEmits(['clicked']);

// computed
// Returns the selected asset.
const asset = computed(() => {
  return assetStore.selectedAsset;
});

// Checks if the asset can be copied to other projects.
const canCopyToOtherProject = computed(() => {
  const hasOtherDownloadedProjects = projectStore.projects.filter(project => 
    project.is_downloaded && 
    project.uri !== projectStore.activeProject?.uri
  ).length > 0;
  const assetIsNormal = asset.value?.file_status === 'normal';
  return hasOtherDownloadedProjects && assetIsNormal && userStore.userCanCreateProject;
});

// Checks if any filters are active.
const filtersActive = computed(() => {
  const assigneeFilters = commonStore.hasAssignees || commonStore.noAssignees;
  const entityFilters = commonStore.entityFilters.length > 0;
  const taskFilters = commonStore.taskFilters.length > 0;
  const resourceFilters = commonStore.resourceFilters.length > 0;
  const generalFilter = isFilterActive('general');
  return assigneeFilters || entityFilters || taskFilters || resourceFilters || generalFilter;
});

// Checks if there are collections to move to or the asset is in a collection (can move to root).
const hasCollections = computed(() => {
  const collectionsExist = collectionStore.getCollections.length > 0;
  const assetHasParent = !!asset.value?.parent_id;
  return collectionsExist || assetHasParent;
});

// Checks if the selected asset is an archive.
const isArchive = computed(() => {
  const archiveFormats = ['.zip', '.rar', '.7z', '.tar', '.gz', '.bz2'];
  const extension = asset.value?.extension?.toLowerCase() || '';
  return archiveFormats.includes(extension);
});

// Checks if the asset has been modified.
const isAssetModified = computed(() => {
  return assetStore.selectedAsset.file_status === 'modified';
});

// Checks if the asset is not on disk.
const isNotOnDisk = computed(() => {
  return asset.value?.file_status === 'rebuildable';
});

// methods
// Builds the asset with all its dependencies.
const buildWithDependencies = async () => {
  menu.hideContextMenu();
  let assetIds = [asset.value.id, ...asset.value.dependencies];
  for (let entityId of asset.value.entity_dependencies) {
    let entityAssets = assetStore.getEntityAssets(entityId, true);
    for (let entityAsset of entityAssets) {
      if (!assetIds.includes(entityAsset.id)) {
        assetIds.push(entityAsset.id);
      }
    }
  }
  await CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, assetIds)
    .then(() => {
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      notificationStore.errorNotification("Error Reverting Assets", error);
      console.error(error);
    });
};

// Copies the asset path to clipboard.
const copyAssetPath = async (pathType) => {
  let asset = assetStore.selectedAsset;
  let assetPath = asset.file_path;
  assetPath = assetPath.replace(/\\/g, '/');
  let assetDir = assetPath.split('/').slice(0, -1).join('/');
  let resourcesFolder = assetDir + '/resources';
  let outputPath = assetDir + '/output';
  if (pathType === 'resources') {
    assetPath = resourcesFolder;
  } else if (pathType === 'output') {
    assetPath = outputPath;
  }
  await Clipboard.SetText(assetPath);
  notificationStore.addNotification('Path copied to clipboard', "", "success");
  menu.hideContextMenu();
};

// Shows the copy to project sub-menu.
const copyToProject = () => {
  menu.showSubMenu('assetMenu', {
    type: 'projects',
    title: 'Select Project'
  });
};

// Deletes the selected asset.
const deleteAsset = async () => {
  let assetId = assetStore.selectedAsset.id;
  let longMessage = `Asset of name: ${assetStore.selectedAsset.name} was moved to Trash.`;
  panes.setPaneVisibility('projectDetails', true);
  menu.hideContextMenu();
  assetStore.selectedAsset = null;
  AssetService.DeleteAsset(projectStore.activeProject.uri, assetId, true)
    .then(async () => {
      assetStore.selectedAsset = null;
      stage.markedItems = [];
      projectStore.refreshActiveProject();
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      notificationStore.errorNotification("Asset failed to delete.", error);
    });
  notificationStore.addNotification("Asset moved to Trash.", longMessage, "success", true);
};

// Duplicates the selected asset.
const duplicateAsset = async () => {
  menu.hideContextMenu();
  
  try {
    stage.operationActive = true;
    let selectedAsset = assetStore.selectedAsset;
    
    await AssetService.DuplicateAsset(projectStore.activeProject.uri, selectedAsset.id)
      .then(async (duplicatedAsset) => {
        if (selectedAsset.file_path && duplicatedAsset.file_path) {
          try {
            await FSService.DuplicateFile(selectedAsset.file_path, duplicatedAsset.file_path);
          } catch (fileError) {
            console.error('Error duplicating physical file:', fileError);
          }
        }
        
        emitter.emit('refresh-browser');
        assetStore.selectAsset(duplicatedAsset);
        stage.selectedItem = duplicatedAsset;
        stage.markedItems = [duplicatedAsset.id];
        stage.lastSelectedItemId = "";
        stage.firstSelectedItemId = duplicatedAsset.id;
        
        notificationStore.addNotification('Asset Duplicated', `Asset duplicated`, 'success');
      });
  } catch (error) {
    console.error('Error duplicating asset:', error);
    notificationStore.errorNotification('Failed to Duplicate Asset', error);
  } finally {
    stage.operationActive = false;
  }
};

// Opens the edit asset modal.
const editAsset = () => {
  modals.setModalVisibility('editAssetModal', true);
  menu.hideContextMenu();
};

// Extracts the archive file.
const extractArchive = async () => {
  menu.hideContextMenu();
  
  try {
    const selectedAsset = assetStore.selectedAsset;
    
    if (selectedAsset.file_status === 'rebuildable') {
      notificationStore.errorNotification('Cannot Extract', 'File must be downloaded first');
      return;
    }
    
    const filePath = selectedAsset.file_path;
    
    if (!await FSService.Exists(filePath)) {
      notificationStore.errorNotification('Cannot Extract', 'Archive file not found');
      return;
    }
    
    await FSService.ExtractAll(filePath)
      .then(() => {
        notificationStore.addNotification('Archive Extracted', `Successfully extracted ${selectedAsset.name}`, 'success');
      })
      .catch((error) => {
        console.error('Error extracting archive:', error);
        notificationStore.errorNotification('Failed to Extract Archive', error);
      });
  } catch (error) {
    console.error('Error extracting archive:', error);
    notificationStore.errorNotification('Failed to Extract Archive', error);
  }
};

// Frees up space by deleting the asset file.
const freeUpSpace = async () => {
  let asset = assetStore.selectedAsset;
  let assetDir = asset.file_path.replace(/\\/g, '/');
  await FSService.DeleteFile(assetDir)
    .then(() => {
      asset.file_status = 'rebuildable';
      assetStore.rebuildableAssetsPath.push(asset.task_path);
      assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(assetPath => assetPath !== asset.task_path);
      assetStore.modifiedAssetsPath = assetStore.modifiedAssetsPath.filter(assetPath => assetPath !== asset.task_path);
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      console.error(error);
    });
  modals.disableAllModals();
};

// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Navigates to the dependency graph view.
const goToDependencyGraph = () => {
  stage.setStageVisibility('dependencies', true);
  menu.hideContextMenu();
};

// Navigates to the asset's location in the browser.// Navigates to the asset's location in the browser.
const goToLocation = async () => {
  commonStore.activeWorkspace = 'Default';
  menu.hideContextMenu();
  commonStore.viewSearchQuery = '';
  commonStore.resetFilters();
  
  try {
    commonStore.navigatorMode = true;
    const selectedAsset = assetStore.selectedAsset;
    if (selectedAsset && selectedAsset.entity_id) {
      const parentEntity = await CollectionService.GetCollectionByID(projectStore.activeProject.uri, selectedAsset.entity_id);
      if (parentEntity) {
        collectionStore.navigatedCollection = parentEntity;
        collectionStore.selectedCollection = parentEntity;
      }
    }
  } catch (error) {
    console.error('Error navigating to location:', error);
    notificationStore.errorNotification('Failed to navigate to location', error);
  }
};

// Checks if a specific filter type is active.
const isFilterActive = (filter) => {
  if (filter.includes('general')) {
    const isActive = commonStore.showEntities && commonStore.showTasks
      && commonStore.showResources && commonStore.showChildEntities
      && commonStore.showChildTasks && commonStore.showDependencies && !commonStore.onlyAssets;
    return !isActive;
  } else if (filter.includes('entity')) {
    return commonStore.entityFilters.some((item) => item.type === filter);
  } else if (filter.includes('assignation')) {
    const assigneeFilters = commonStore.hasAssignees || commonStore.noAssignees;
    const assignationFilters = commonStore.taskFilters.some((item) => item.type === filter);
    return assigneeFilters || assignationFilters;
  } else {
    return commonStore.taskFilters.some((item) => item.type === filter);
  }
};

// Launches the asset with the system's default application.
const launchAssetWithCommand = async () => {
  let asset = assetStore.selectedAsset;
  if (asset.is_link && isValidWeblink(asset.pointer)) {
    Browser.OpenURL(asset.pointer);
  } else {
    let file_path = asset.pointer ? asset.pointer : asset.file_path;
    if (await FSService.Exists(file_path)) {
      FSService.LaunchFileWith(file_path);
    } else {
      CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [asset.id])
        .then(async () => {
          let fileStatus = await assetStore.getAssetFileStatus(asset);
          asset.file_status = fileStatus;
          FSService.LaunchFileWith(file_path);
        })
        .catch((error) => {
          console.error(error);
          notificationStore.errorNotification("Error Rebuilding Asset", error);
        });
    }
  }
  menu.hideContextMenu();
};

// Shows the move to collection sub-menu.
const moveToCollection = () => {
  menu.subMenuState.selectedAssetIds = [assetStore.selectedAsset.id];
  menu.subMenuState.startingEntityId = assetStore.selectedAsset.entity_id || '';
  menu.showSubMenu('assetMenu', {
    type: 'move-to-collection',
    title: 'Select Collection'
  });
};

// Prepares and shows the free up space confirmation modal.
const prepFreeUpSpacePopUpModal = () => {
  trayStates.popUpModalTitle = "Free Up Asset Space";
  trayStates.popUpModalMessage = "Are you sure you want to delete this asset working files? This will permanently remove all uncheckpointed resources and all asset outputs. Please confirm if you wish to proceed.";
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpSpace;
  modals.setModalVisibility('popUpModal', true);
  menu.hideContextMenu();
};

// Emits event to rename the asset.
const renameAsset = () => {
  emitter.emit('renameAsset');
  menu.hideContextMenu();
};

// Reverts the asset to its last checkpointed state.
const revertAsset = async () => {
  let assetId = assetStore.selectedAsset.id;
  CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [assetId])
    .then(() => {
      assetStore.selectedAsset.file_status = "normal";
    })
    .catch((error) => {
      notificationStore.errorNotification("Failed to Revert Asset", error);
    });
  menu.hideContextMenu();
};

// Reveals the asset in the file explorer.
const revealInExplorer = async () => {
  menu.hideContextMenu();
  const assetId = assetStore.selectedAsset.id;

  if (assetStore.selectedAsset.file_status == "rebuildable") {
    await CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [assetId])
      .then(async () => {
        assetStore.rebuildableAssetsPath = assetStore.rebuildableAssetsPath.filter(assetPath => assetPath !== asset.task_path);
        assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(assetPath => assetPath !== asset.task_path);
        emitter.emit('get-project-data');
      })
      .catch((error) => {
        notificationStore.errorNotification("Error downloading Asset", error);
        console.error(error);
      });
  }
  AssetService.RevealAsset(projectStore.activeProject.uri, assetStore.selectedAsset.id);
};

// lifecycle hooks
onMounted(() => {
  menu.popUpMenuWidth = popUpMenu.value.getBoundingClientRect().width;
  menu.popUpMenu = popUpMenu.value;
});

onBeforeUnmount(() => {
  menu.popUpMenuWidth = popUpMenu.value.getBoundingClientRect().width;
  menu.popUpMenuHeight = popUpMenu.value.getBoundingClientRect().height;
});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/menu.css";
</style>








