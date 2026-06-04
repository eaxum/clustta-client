<template>
  <div v-if="debugging" class="general-pane-header">
    <HeaderArea :title="utils.capitalizeStr(projectStore.selectedUntrackedItem?.name)" :notModal="true" :icon="getAppIcon(itemIcon)" />
  </div>

  <div v-if="debugging" class="general-pane-root">
    <div class="general-pane-container">
      <div class="pane-parameter-section">
        <div class="action-bar">

          <div class="action-bar-section">
            <ActionButton @click="deleteItem" :icon="getAppIcon('trash')" :label="$t('panes.deleteItem')" />
          </div>
        </div>

        <div class="asset-details">

          <div v-if="projectStore.selectedUntrackedItem.type === 'untracked_asset'" class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.extension') }}
            </div>
            <div class="simple-text-value">
              {{ projectStore.selectedUntrackedItem.extension }}
            </div>
          </div>

          <div v-if="projectStore.selectedUntrackedItem.type === 'untracked_asset'" class="pane-parameter-detail">
            <div class="simple-text-key">
            {{ $t('panes.size') }}
            </div>
            <div class="simple-text-value">
              {{  itemSize }}
            </div>
          </div>

          <div v-if="projectStore.selectedUntrackedItem.type === 'untracked_collection'" class="pane-parameter-detail">
            <div class="simple-text-key">
            {{ $t('panes.size') }}
            </div>
            <div class="simple-text-value">
              {{  collectionSize }}
            </div>
        </div>

          <div class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.location') }}
            </div>
            <div class="simple-text-value truncate-path" v-tooltip="itemPath">
              {{ itemPath }}
            </div>
            <div v-if="!platformStore.isWeb" class="pane-parameter-actions">
              <ActionButton :icon="getAppIcon('copy')" v-tooltip="$t('common.copyPath')" @click="copyItemPath" />
              <ActionButton :icon="getAppIcon('folder-arrow-up-right')" v-tooltip="revealLabel" :buttonFunction="revealInExplorer" />
            </div>
          </div>

        </div>

      </div>
    </div>
  </div>
</template>

<script setup>


import { FSService, AssetService, CollectionService } from "@/services";
import { Clipboard } from '@wailsio/runtime';

// imports
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRevealLabel } from '@/composables/useRevealLabel';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// store imports
import { useIconStore } from '@/stores/icons';
import { useUserStore } from '@/stores/users';
import { useModalStore } from '@/stores/modals';
import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useStageStore } from '@/stores/stages';
import { useProjectStore } from '@/stores/projects';
import { useDndStore } from '@/stores/dnd';
import { usePaneStore } from '@/stores/panes';
import { useAssetStore } from '@/stores/assets';
import { usePlatformStore } from '@/stores/platform';
import { useNotificationStore } from '@/stores/notifications';

// state imports
import { useTrayStates } from '@/stores/TrayStates';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// states
const trayStates = useTrayStates();

// stores
const iconStore = useIconStore();
const userStore = useUserStore();
const modalStore = useModalStore();
const collectionStore = useCollectionStore();
const modals = useDesktopModalStore();
const stage = useStageStore();
const projectStore = useProjectStore();
const dndStore = useDndStore();
const panes = usePaneStore();
const assetStore = useAssetStore();
const notificationStore = useNotificationStore();
const platformStore = usePlatformStore();

const { t } = useI18n();
const { revealLabel } = useRevealLabel();

// vars
const debugging = ref(true);

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const parentName = computed(() => {
  const parentId = projectStore.selectedUntrackedItem.parent_id
  const parent = collectionStore.getCollections.find((item) => item.id === parentId)
  return parent ? parent.name : 'None'
});

const itemIcon = computed(() => {
  const item = projectStore.selectedUntrackedItem;
  if (item?.type === 'untracked_asset') {
    return 'file'
  } else if (item?.type === 'untracked_collection') {
    return 'folder'
  }
});

const untrackedItem = computed(() => {
  return projectStore.selectedUntrackedItem
});

const deleteItem = () => {
  prepDeleteUntrackedItemPopUpModal();
};

const deleteUntrackedFolder = () => {
  FSService.DeleteFolder(untrackedItem.value.file_path);
  projectStore.removeUntrackedCollection(untrackedItem.value.id);
  panes.setPaneVisibility('projectDetails', true)
  collectionStore.selectedCollection = null;
  stage.markedItems = [];
  emitter.emit('refresh-browser')
  modals.disableAllModals();
};

const deleteUntrackedFile = () => {
  FSService.DeleteFile(untrackedItem.value.file_path);
  projectStore.removeUntrackedAsset(untrackedItem.value.id);
  panes.setPaneVisibility('projectDetails', true)
  assetStore.selectedAsset = null;
  stage.markedItems = [];
  emitter.emit('refresh-browser')
  modals.disableAllModals();
};

const prepDeleteUntrackedItemPopUpModal = () => {
  const untrackedItemType = untrackedItem.value?.type;
  trayStates.popUpModalTitle = t('common.delete');
  trayStates.popUpModalMessage = t('confirmations.deleteItemPermanently');
  trayStates.popUpModalIcon = 'trash';
  trayStates.popUpModalFunction = untrackedItemType === 'untracked_collection' ? deleteUntrackedFolder : deleteUntrackedFile;
  modals.setModalVisibility('popUpModal', true);
};

const copyItemPath = async () => {
  if (!itemPath.value) return;
  await Clipboard.SetText(itemPath.value);
  const message = t('notifications.pathCopied');
  notificationStore.addNotification(message, "", "success");
};

const revealInExplorer = () => {
  if (!itemPath.value) return;
  FSService.RevealInExplorer(untrackedItem.value.file_path);
};

const importItem = () => {
  if (untrackedItem.value.type === 'untracked_asset') {
    importAsset();
  } else if (untrackedItem.value.type === 'untracked_collection') {
    importFolder();
  }
};

const importFolder = () => {
  const inRoot = untrackedItem.value.collection_path === ""
  const folderPath = untrackedItem.value.file_path;
  let parentId = ""
  let untrackedParents = []
  let parentPaths = utils.getParentPaths(untrackedItem.value.collection_path)
  if (!inRoot) {
    for (let parent of parentPaths) {
      parentId = collectionStore.collections.find((item) => item.collection_path === parent)?.id;
      if (parentId !== undefined) {
        break
      }
      untrackedParents.unshift(parent)
    }
  }

  dndStore.untrackedParents = untrackedParents
  dndStore.targetItemId = parentId;
  dndStore.droppedFolders.push(folderPath);
  panes.setPaneVisibility('projectDetails', true);
  modals.setModalVisibility('importItemsModal', true);
};

const importAsset = () => {
  const inRoot = untrackedItem.value.collection_path === ""
  const assetPath = untrackedItem.value.file_path;
  let parentId = ""
  let untrackedParents = []
  let parentPaths = utils.getParentPaths(untrackedItem.value.collection_path)
  if (!inRoot) {
    for (let parent of parentPaths) {
      parentId = collectionStore.collections.find((item) => item.collection_path === parent)?.id;
      if (parentId !== undefined) {
        break
      }
      untrackedParents.unshift(parent)
    }
  }
  dndStore.untrackedParents = untrackedParents
  dndStore.targetItemId = parentId;
  dndStore.droppedFiles.push(assetPath);
  panes.setPaneVisibility('projectDetails', true);
  modals.setModalVisibility('importItemsModal', true);
};

const itemSize = ref(0);
const collectionSize = ref(0);

const itemType = computed(() => {
  return projectStore.selectedUntrackedItem?.type;
})

const itemPath = computed(() => {
  const path = projectStore.selectedUntrackedItem?.file_path;
  if (!path) return '';
  return path.replace(/\\/g, '/')
});

const getItemSize = async() => {
  if (!itemPath.value) return;
  try {
    const size = await FSService.FileStat(itemPath.value);
    itemSize.value = size.formattedSize;
  } catch (error) {
    itemSize.value = 'Not on disk';
  }
}

const getCollectionSize = async() => {
  if (!itemPath.value) return;
  try {
    const size = await FSService.FolderSize(itemPath.value);
    collectionSize.value = size;
  } catch (error) {
    collectionSize.value = 'Not on disk';
  }
}

const getProjectData = async () => {
  if (!itemPath.value) return;
  
  if(itemType.value === 'untracked_asset'){
    if (!await FSService.Exists(itemPath.value)){
      itemSize.value = 'Not on disk'
      return
    }
    getItemSize();
  } else if (itemType.value === 'untracked_collection') {
    getCollectionSize();
  }
}

watch(() => projectStore.selectedUntrackedItem, () => {
  itemSize.value = 0;
  collectionSize.value = 0;
  getProjectData();
});


// onMounted
onMounted(() => {
  stage.markedAssets = [];
  
  getProjectData();
	emitter.on('get-project-data', getProjectData);
});

onBeforeUnmount(() => {
	emitter.off('get-project-data', getProjectData);
});


</script>
<style scoped>
@import "@/assets/desktop.css";

.compound-input-section {
  /* background-color: royalblue; */
  /* flex: 1; */
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .4rem;
  width: 100%;
  justify-content: space-between;
  justify-content: space-around;
}

.pane-parameter-section {
  flex: 1;
  height: 200px;
}

.asset-details{
  overflow: hidden;
  overflow-y: scroll;
  padding-right: .5rem;
}

.asset-details::-webkit-scrollbar {
  width: 4px;
}

.asset-details::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--surface-4);
}

.asset-details::-webkit-scrollbar-track {
  border-radius: 10px;
}
.pane-parameter-detail {
  display: flex;
  font-size: 14px;
  height: max-content;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  height: 30px;
  border-bottom: var(--transparent-line);
}

.simple-text-key {
  white-space: nowrap;
}

.truncate-path {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
  flex: 1;
}

.action-bar {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: .6rem;
  width: max-content;
  width: 100%;
  /* justify-content: space-around; */
  height: max-content;
  padding: .2rem;
  /* background-color: black; */
  /* background-color: tomato; */
  align-items: flex-start;
  box-sizing: border-box;
}

.action-bar-section {
  display: flex;
  align-items: center;
  gap: .5rem;
  justify-content: space-between;
  width: 100%;
}
</style>