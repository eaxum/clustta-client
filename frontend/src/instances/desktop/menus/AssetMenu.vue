<template>
  <div ref="popUpMenu" class="filter-menu-container">

    <ActionButton v-if="!platformStore.isWeb && userStore.canDo('pull_chunk')" :icon="getAppIcon('launch')" :showLabel="true" :fullWidth="true"
      :label="$t('common.openWith')" :buttonFunction="launchAssetWithCommand" />

    <span v-if="!platformStore.isWeb && userStore.canDo('pull_chunk')" class="menu-divider"></span>

    <ActionButton v-if="userStore.canDo('update_asset')" :icon="getAppIcon('edit')" :showLabel="true" :fullWidth="true"
      :label="$t('common.rename')" :buttonFunction="renameAsset" />

    <ActionButton v-if="userStore.canDo('update_asset')" :icon="getAppIcon('switches')" :showLabel="true"
      :fullWidth="true" :label="$t('common.edit')" :buttonFunction="editAsset" />

    <ActionButton v-if="userStore.canDo('create_asset')" :icon="getAppIcon('duplicate')" :showLabel="true"
      :fullWidth="true" :label="$t('common.duplicate')" :buttonFunction="duplicateAsset" />

    <!-- Copy to Project -->
    <ActionButton v-if="!platformStore.isWeb && userStore.canDo('create_asset') && canCopyToOtherProject" 
      :icon="getAppIcon('briefcase')" :showLabel="true"
      :fullWidth="true" :label="$t('menus.copyToProject')" :buttonFunction="copyToProject" />

    <!-- Move to Collection -->
    <ActionButton v-if="!platformStore.isWeb && userStore.canDo('update_asset')" 
      :icon="getAppIcon('folder-arrow-in')" :showLabel="true"
      :fullWidth="true" :label="$t('common.move')" :buttonFunction="moveToCollection" />

    <ActionButton v-if="!platformStore.isWeb && (asset.dependencies.length || asset.collection_dependencies.length)" :icon="getAppIcon('jigsaw')" :showLabel="true"
      :fullWidth="true" :label="$t('menus.buildWithDependencies')" :buttonFunction="buildWithDependencies" />

    <ActionButton v-if="isRemoteProject && userStore.canDo('manage_dependencies')" :icon="getAppIcon('dependency')" :showLabel="true"
      :fullWidth="true" :label="$t('menus.dependencyGraph')" :buttonFunction="goToDependencyGraph" />

    <!-- Go to Location -->
    <ActionButton v-if="commonStore.viewSearchQuery || filtersActive" :icon="getAppIcon('file-search')" :showLabel="true" :fullWidth="true"
      :label="$t('menus.goToAsset')" :buttonFunction="goToLocation" />

    <!-- Reveal in Explorer -->
    <span v-if="!platformStore.isWeb" class="horizontal-flex">
      <ActionButton :icon="getAppIcon('folder-arrow-up-right')" :showLabel="true" :fullWidth="true" :label="$t('common.showInExplorer')"
        :buttonFunction="revealInExplorer" />
      <ActionButton :icon="getAppIcon('copy')" :showLabel="false" :fullWidth="false" @click="copyAssetPath('asset')"
        v-tooltip="$t('common.copyPath')" />
    </span>

    <!-- Extract Archive -->
    <ActionButton v-if="!platformStore.isWeb && isArchive" :icon="getAppIcon('unarchive')" :showLabel="true" :fullWidth="true" 
      :label="$t('common.extract')" :buttonFunction="extractArchive" />

    <!-- Checkpoints -->
    <ActionButton v-if="!platformStore.isWeb && isAssetModified" :noFilter="true" :icon="getAppIcon('revert')" :useAlert="true" :showLabel="true" :fullWidth="true"
      :label="$t('menus.revertFile')" :buttonFunction="revertAsset" />

    <!-- Sync Asset -->
    <ActionButton v-if="isRemoteProject && !asset.synced" :icon="getAppIcon('cloud-up')" :showLabel="true" :fullWidth="true"
      :label="$t('menus.syncAsset')" :buttonFunction="syncAsset" />

    <span v-if="userStore.canDo('delete_asset') || !isNotOnDisk" class="menu-divider"></span>

    <!-- Free space -->
    <ActionButton :icon="getAppIcon('broom')" v-if="!platformStore.isWeb && !isNotOnDisk" :showLabel="true" :fullWidth="true"
      :label="$t('common.freeUpSpace')" :buttonFunction="prepFreeUpSpacePopUpModal" />

    <!-- Delete Asset -->
    <ActionButton :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true" :label="$t('common.delete')"
      v-if="userStore.canDo('delete_asset')" :buttonFunction="deleteAsset" />

  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { Clipboard } from '@wailsio/runtime';
import emitter from '@/lib/mitt';
import { isValidWeblink } from '@/lib/pointer';
import { useI18n } from 'vue-i18n';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// services
import { AssetService, CheckpointService, CollectionService, FSService, SyncService } from "@/services";

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
import { useStudioStore } from '@/stores/studio';

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
const studioStore = useStudioStore();
const { t } = useI18n();

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
  return hasOtherDownloadedProjects && assetIsNormal && studioStore.isStudioAdmin;
});

// Checks if any filters are active.
const filtersActive = computed(() => {
  const assigneeFilters = commonStore.hasAssignees || commonStore.noAssignees;
  const collectionFilters = commonStore.collectionFilters.length > 0;
  const assetFilters = commonStore.assetFilters.length > 0;
  const resourceFilters = commonStore.resourceFilters.length > 0;
  const generalFilter = isFilterActive('general');
  return assigneeFilters || collectionFilters || assetFilters || resourceFilters || generalFilter;
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

// Checks if the active project is remote.
const isRemoteProject = computed(() => {
  return projectStore.activeProject?.has_remote;
});

// methods
// Emits asset data updates to Browser and VirtuaItem components.
const emitAssetUpdates = (assetId, updates) => {
  const updateData = { itemId: assetId, updates };
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

// Builds the asset with all its transitive dependencies (assets and collections),
// resolved on the backend so the full graph is expanded and de-duplicated.
const buildWithDependencies = async () => {
  menu.hideContextMenu();
  try {
    const assetIds = await AssetService.ResolveBuildDependencies(
      projectStore.activeProject.uri,
      asset.value.id,
    );
    await CheckpointService.Revert(
      projectStore.activeProject.uri,
      projectStore.getActiveProjectUrl,
      assetIds,
    );
    emitter.emit('refresh-browser');
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorRevertingAssets'), error);
    console.error(error);
  }
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
  notificationStore.addNotification(t('notifications.pathCopied'), "", "success");
  menu.hideContextMenu();
};

// Shows the copy to project sub-menu.
const copyToProject = () => {
  menu.showSubMenu('assetMenu', {
    type: 'projects',
    title: t('menus.selectProject')
  });
};

// Deletes the selected asset.
const deleteAsset = async () => {
  let assetId = assetStore.selectedAsset.id;
  let longMessage = t('notifications.movedToTrash', { item: assetStore.selectedAsset.name });
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
      notificationStore.errorNotification(t('notifications.assetFailedToDelete'), error);
    });
  notificationStore.addNotification(t('notifications.movedToTrash', { item: 'Asset' }), longMessage, "success", true);
};

// Duplicates the selected asset.
const duplicateAsset = async () => {
  menu.hideContextMenu();
  
  try {
    stage.operationActive = true;
    let selectedAsset = assetStore.selectedAsset;
    
    await AssetService.DuplicateAsset(projectStore.activeProject.uri, selectedAsset.id, '')
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
        
        notificationStore.addNotification(t('notifications.assetDuplicated'), t('notifications.assetDuplicated'), 'success');
      });
  } catch (error) {
    console.error('Error duplicating asset:', error);
    notificationStore.errorNotification(t('notifications.failedToDuplicateAsset'), error);
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
      notificationStore.errorNotification(t('notifications.cannotExtract'), t('notifications.fileMustBeDownloaded'));
      return;
    }
    
    const filePath = selectedAsset.file_path;
    
    if (!await FSService.Exists(filePath)) {
      notificationStore.errorNotification(t('notifications.cannotExtract'), t('notifications.archiveNotFound'));
      return;
    }
    
    await FSService.ExtractAll(filePath)
      .then(() => {
        notificationStore.addNotification(t('notifications.archiveExtracted'), t('notifications.archiveExtracted', { name: selectedAsset.name }), 'success');
      })
      .catch((error) => {
        console.error('Error extracting archive:', error);
        notificationStore.errorNotification(t('notifications.failedToExtractArchive'), error);
      });
  } catch (error) {
    console.error('Error extracting archive:', error);
    notificationStore.errorNotification(t('notifications.failedToExtractArchive'), error);
  }
};

// Frees up space by deleting the asset file.
const freeUpSpace = async () => {
  let asset = assetStore.selectedAsset;
  let assetDir = asset.file_path.replace(/\\/g, '/');
  await FSService.DeleteFile(assetDir)
    .then(() => {
      asset.file_status = 'rebuildable';
      assetStore.rebuildableAssetsPath.push(asset.asset_path);
      assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(assetPath => assetPath !== asset.asset_path);
      assetStore.modifiedAssetsPath = assetStore.modifiedAssetsPath.filter(assetPath => assetPath !== asset.asset_path);
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

// Opens the dependency graph modal.
const goToDependencyGraph = () => {
  modals.setModalVisibility('dependencyGraphModal', true);
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
    if (selectedAsset && selectedAsset.collection_id) {
      const parentCollection = await CollectionService.GetCollectionByID(projectStore.activeProject.uri, selectedAsset.collection_id);
      if (parentCollection) {
        collectionStore.navigatedCollection = parentCollection;
        collectionStore.selectedCollection = parentCollection;
      }
    }
  } catch (error) {
    console.error('Error navigating to location:', error);
    notificationStore.errorNotification(t('notifications.failedToNavigate'), error);
  }
};

// Checks if a specific filter type is active.
const isFilterActive = (filter) => {
  if (filter.includes('general')) {
    const isActive = commonStore.showCollections && commonStore.showAssets
      && commonStore.showResources && commonStore.showChildCollections
      && commonStore.showChildAssets && commonStore.showDependencies && !commonStore.onlyAssets;
    return !isActive;
  } else if (filter.includes('collection')) {
    return commonStore.collectionFilters.some((item) => item.type === filter);
  } else if (filter.includes('assignation')) {
    const assigneeFilters = commonStore.hasAssignees || commonStore.noAssignees;
    const assignationFilters = commonStore.assetFilters.some((item) => item.type === filter);
    return assigneeFilters || assignationFilters;
  } else {
    return commonStore.assetFilters.some((item) => item.type === filter);
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
  menu.subMenuState.startingCollectionId = assetStore.selectedAsset.collection_id || '';
  menu.showSubMenu('assetMenu', {
    type: 'move-to-collection',
    title: t('menus.selectCollection')
  });
};

// Prepares and shows the free up space confirmation modal.
const prepFreeUpSpacePopUpModal = () => {
  trayStates.popUpModalTitle = t('menus.freeUpAssetSpace');
  trayStates.popUpModalMessage = t('confirmations.deleteWorkingFiles', { item: 'asset' });
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
      notificationStore.errorNotification(t('notifications.failedToRevertAsset'), error);
    });
  menu.hideContextMenu();
};

// Syncs the selected asset (including checkpoints and chunks) to the server.
const syncAsset = () => {
  const assetId = assetStore.selectedAsset.id;
  if (!assetId || !projectStore.activeProject?.has_remote) return;
  menu.hideContextMenu();
  SyncService.SyncAsset(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, assetId)
    .then(() => {
      assetStore.selectedAsset.synced = true;
      emitAssetUpdates(assetId, [
        { property: 'synced', value: true }
      ]);
      notificationStore.addNotification(t('common.sync'), t('notifications.assetSyncedSuccessfully'), 'success');
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification(t('notifications.errorSyncingAsset'), error);
    });
};

// Reveals the asset in the file explorer.
const revealInExplorer = async () => {
  menu.hideContextMenu();
  const assetId = assetStore.selectedAsset.id;

  if (assetStore.selectedAsset.file_status == "rebuildable") {
    await CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [assetId])
      .then(async () => {
        assetStore.rebuildableAssetsPath = assetStore.rebuildableAssetsPath.filter(assetPath => assetPath !== asset.asset_path);
        assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(assetPath => assetPath !== asset.asset_path);
        emitter.emit('get-project-data');
      })
      .catch((error) => {
        notificationStore.errorNotification(t('notifications.errorDownloadingAsset'), error);
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








