<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>


    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="'folder-arrow-in'" :showSearch="false" />
      <!-- <ActionButton :icon="getAppIcon('edit')" :showLabel="false" :isActive="dndStore.importEditMode"
        v-tooltip="'Edit Items'" :buttonFunction="toggleEditItems" /> -->
    </div>

    <div class="general-container general-container-wide">

      <div class="selected-folder">
        <ImportPreview />
      </div>

      <div class="pop-up-actions" ref="popUpActions">
        <GeneralButton :label="$t('common.close')" :fullWidth="false" :buttonFunction="closeModal" :isActive="!isAwaitingResponse" :colored="false" />
        <GeneralButton :label="$t('common.confirm')" :fullWidth="false" @click="importItems()" :isActive="storeHasData"
          :loading="isAwaitingResponse" />
      </div>

    </div>

  </div>

</template>

<script setup>
// imports
import { computed, onMounted, onUnmounted, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { v4 as uuidv4 } from 'uuid';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ImportPreview from '@/instances/desktop/components/ImportPreview.vue';

// services
import { ImportService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useDndStore } from '@/stores/dnd';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useStatusStore } from '@/stores/status';
import { useTrayStates } from '@/stores/TrayStates';

const { t } = useI18n();
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const dndStore = useDndStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const statusStore = useStatusStore();
const trayStates = useTrayStates();

// refs
const isAwaitingResponse = ref(false);
const modalContainer = ref(null);
const popUpActions = ref(null);

// constants
const title = t('modals.importItems');

// computed
// Returns whether the preview data store has any items.
const storeHasData = computed(() => {
  const rawData = dndStore.previewData;
  return Object.values(rawData).some(arr => arr.length > 0);
});

// methods
// Closes the modal and resets drag and drop values.
const closeModal = () => {
  dndStore.targetItemId = '';
  dndStore.trackedParents = [];
  dndStore.untrackedParents = [];
  dndStore.droppedFolders = [];
  dndStore.droppedFiles = [];
  modals.setModalVisibility('importItemsModal', false);
  resetDndValues();
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Returns the parent path from a given path.
const getPathParent = (path) => {
  if (!path.includes('/')) {
    return '';
  }
  const lastSlashIndex = path.lastIndexOf('/');
  return path.substring(0, lastSlashIndex);
};

// Imports the previewed items by creating collections and assets.
const importItems = async (comment = 'Asset created') => {
  const collections = dndStore.previewData.collections.filter(collection => !collection.is_tracked_parent);
  const assets = dndStore.previewData.assets;
  let success = false;
  let errorMessage;
  try {
    for (let i = 0; i < collections.length; i += 100) {
      const batch = collections.slice(i, i + 100);
      await ImportService.CreateCollections(projectStore.activeProject.uri, batch, i, collections.length);
    }
    for (let i = 0; i < assets.length; i += 100) {
      const batch = assets.slice(i, i + 100);
      await ImportService.CreateAssets(projectStore.activeProject.uri, batch, i, assets.length, comment);
    }
    success = true;
  } catch (error) {
    console.error('Error caught:', error);
    errorMessage = error;
  }

  if (success) {
    dndStore.previewData = {};
    isAwaitingResponse.value = false;
    refresh();
    closeModal();
  } else {
    isAwaitingResponse.value = false;
    notificationStore.resetProgress();
    notificationStore.errorNotification(t('notifications.errorCreatingItems'), errorMessage);
    closeModal();
  }
};

// Generates preview data for items to be imported.
const previewImportItems = async () => {
  isAwaitingResponse.value = true;
  const folders = dndStore.droppedFolders;
  const files = dndStore.droppedFiles;
  let parentId = dndStore.targetItemId;
  const parentPath = dndStore.targetItemPath;
  const collectionsId = {};
  collectionsId[parentPath] = parentId;

  if (parentId === undefined) {
    parentId = '';
  }
  const trackedParentData = [];
  const untrackedParentData = [];
  const workingDir = projectStore.activeProject.working_directory;
  await ImportService.ImportFolder(projectStore.activeProject.uri, parentId, folders, files, workingDir, projectStore.activeProject.ignore_list)
    .then((response) => {
      if (dndStore.trackedParents.length + dndStore.untrackedParents.length > 0) {
        const collectionTypeId = collectionStore.collectionTypes.find((item) => item.name === 'generic')?.id;
        for (const trackedParent of dndStore.trackedParents) {
          const collectionData = collectionStore.collections.find((item) => item.collection_path === trackedParent);
          collectionData.is_tracked_parent = true;
          collectionData.is_expanded = true;
          trackedParentData.push(collectionData);
          collectionsId[trackedParent] = collectionData.id;
        }
        for (const untrackedParent of dndStore.untrackedParents) {
          const untrackedParentPath = getPathParent(untrackedParent);
          const untrackedParentId = collectionsId[untrackedParentPath];
          const name = untrackedParent.split('/').pop();
          const uid = uuidv4();
          const data = {
            id: uid,
            created_at: '',
            description: '',
            collection_path: '',
            collection_type_icon: 'generic',
            collection_type_id: collectionTypeId,
            collection_type_name: 'generic',
            file_path: workingDir + '/' + untrackedParent,
            is_dependency: false,
            mtime: 0,
            name: name,
            parent_id: untrackedParentId,
            preview: '',
            preview_extension: '',
            preview_id: '',
            synced: false,
            trashed: false,
            is_expanded: true,
          };
          untrackedParentData.push(data);
          collectionsId[untrackedParent] = uid;
        }

        response.collections.forEach(collection => {
          const collectionParentPath = getPathParent(collection.collection_path);
          const collectionParentId = collectionsId[collectionParentPath];
          if (collectionParentId) {
            collection.parent_id = collectionParentId;
          }
        });

        response.assets.forEach(asset => {
          const assetParentPath = getPathParent(asset.asset_path);
          const assetParentId = collectionsId[assetParentPath];
          if (assetParentId) {
            asset.collection_id = assetParentId;
          }
        });

        response.collections = [...trackedParentData, ...untrackedParentData, ...response.collections];
      }
      dndStore.previewData = response;
      isAwaitingResponse.value = false;
    })
    .catch((error) => {
      isAwaitingResponse.value = false;
      console.log(error);
      notificationStore.errorNotification(t('notifications.errorGeneratingPreviews'), error);
    });
};

// Refreshes the project data after import.
const refresh = async () => {
  assetStore.assetsLoaded = false;
  await projectStore.refreshActiveProject();
  await statusStore.reloadStatuses();
  projectStore.getUntrackedItems();
  assetStore.assetsLoaded = true;
};

// Resets drag and drop values after a delay.
const resetDndValues = () => {
  setTimeout(() => {
    dndStore.resetValues();
  }, 100);
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// lifecycle hooks
onMounted(async () => {
  await previewImportItems();
  trayStates.listItemsBoundary = modalContainer.value;
  trayStates.tagSearchQuery = '';
});

onUnmounted(() => {
  dndStore.previewDataSelectedItems = {};
  menu.clickOutsideMask = null;
  dndStore.droppedItems = '';
  stage.markedItems = [];
  stage.firstSelectedItemId = '';
  stage.lastSelectedItemId = '';
  dndStore.importEditMode = false;
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.pop-up-actions {
  align-items: center;
  box-sizing: border-box;
}

.modal-container {
  justify-content: flex-start;
  align-items: flex-start;
  max-height: 90vh;
}

.general-container-wide {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  width: 50vw;
  min-width: 600px !important;
  max-width: 1000px;
  max-height: 80vh;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.selected-folder {
  width: 100%;
  padding: .2rem;
  overflow: hidden;
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  padding: 10px 20px;
  box-sizing: border-box;
}
</style>
