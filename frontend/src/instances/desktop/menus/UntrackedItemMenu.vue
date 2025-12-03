<template>
  <div ref="popUpMenu" class="filter-menu-container">

    <!-- Launch -->
    <ActionButton
      v-if="userStore.canDo('pull_chunk') && untrackedItemStore.selectedUntrackedItem.type == 'untracked_task'"
      :icon="getAppIcon('launch')" :showLabel="true" :fullWidth="true" label="Open With"
      :buttonFunction="launchTaskWithCommand" />

    <!-- <span v-if="userStore.canDo('pull_chunk')" class="menu-divider"></span> -->

    <!-- Rename -->
    <ActionButton v-if="userStore.canDo('update_task')" :icon="getAppIcon('edit')" :showLabel="true" :fullWidth="true"
      label="Rename" :buttonFunction="renameItem" />

    <!-- Ignore -->
    <ActionButton :icon="getAppIcon('file-watch')" :showLabel="true" :fullWidth="true" label="Ignore this file/folder"
      :buttonFunction="ignoreItem" />

    <ActionButton v-if="untrackedItemStore.selectedUntrackedItem.type == 'untracked_task'"
      :icon="getAppIcon('file-watch')" :showLabel="true" :fullWidth="true" label="Ignore extension type"
      :buttonFunction="ignoreExtensionType" />

    <!-- Reveal in Explorer -->
    <span class="horizontal-flex">
      <ActionButton :icon="getAppIcon('folder-arrow-up-right')" :showLabel="true" :fullWidth="true" label="Show in Explorer"
        :buttonFunction="revealInExplorer" />
      <ActionButton :icon="getAppIcon('copy')" :showLabel="false" :fullWidth="false" @click="copyItemPath('task')"
        v-tooltip="'Copy Path'" />
    </span>

    <!-- Extract Archive -->
    <ActionButton v-if="isArchive" :icon="getAppIcon('unarchive')" :showLabel="true" :fullWidth="true" 
      label="Extract" :buttonFunction="extractArchive" />

    <!-- Delete -->
    <ActionButton :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true" label="Delete "
      :buttonFunction="prepDeleteItemPopUpModal" />

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
import { AssetService, CheckpointService } from "@/../bindings/clustta/services";
import { TrashService } from "@/../bindings/clustta/services";

// states/store imports
import { useTrayStates } from '@/stores/TrayStates';
import { useMenu } from '@/stores/menu';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useUserStore } from '@/stores/users';
import { useModalStore } from '@/stores/modals';
import { useAssetStore } from '@/stores/assets';
import { useTemplateStore } from '@/stores/template';
import { useProjectStore } from '@/stores/projects';
import { useUntrackedItemStore } from '@/stores/untracked';
import { useCollectionStore } from '@/stores/collections';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import { FSService, ProjectService } from '@/../bindings/clustta/services';
import { Clipboard } from '@wailsio/runtime';
import { addIgnoredItem } from '@/lib/untracked';

// states/stores
const trayStates = useTrayStates();
const userStore = useUserStore();
const menu = useMenu();
const panes = usePaneStore();
const stage = useStageStore();
const modals = useDesktopModalStore();
const modalStore = useModalStore();
const notificationStore = useNotificationStore();
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const templateStore = useTemplateStore();
const projectStore = useProjectStore();
const untrackedItemStore = useUntrackedItemStore();

// refs
const popUpMenu = ref(null);

// computed properties
// Check if the selected untracked item is an archive
const isArchive = computed(() => {
  const archiveFormats = ['.zip', '.rar', '.7z', '.tar', '.gz', '.bz2'];
  const item = untrackedItemStore.selectedUntrackedItem;
  if (item?.type !== 'untracked_task') {
    return false;
  }
  const extension = item?.extension?.toLowerCase() || '';
  return archiveFormats.includes(extension);
});

// props
const props = defineProps({
  top: Number,
  left: Number,
});

const emit = defineEmits(['clicked']);

const launchTaskWithCommand = async () => {
  let item = untrackedItemStore.selectedUntrackedItem
  let file_path = item.file_path
  if (await FSService.Exists(file_path)) {
    FSService.LaunchFileWith(file_path)
  } else {
    notificationStore.addNotification("File Not On Disk, Rebuild", "File not found on disk, rebuild task", "error")
  }
  menu.hideContextMenu();
};

const renameItem = () => {
  emitter.emit('renameAsset');
  menu.hideContextMenu();
};

function removeLastSlash(text) {
    const lastSlashIndex = text.lastIndexOf('/');
    if (lastSlashIndex !== -1) {
        return text.slice(0, lastSlashIndex) + text.slice(lastSlashIndex + 1);
    }
    return text;
}

const ignoreItem = async () => {
  let item = untrackedItemStore.selectedUntrackedItem
  if (item.type == "untracked_task") {
    await addIgnoredItem(item.task_path)
  } else {
    const untrackedEntity = removeLastSlash(item.item_path)
    await addIgnoredItem(untrackedEntity)
  }
  emitter.emit('refresh-browser');
  menu.hideContextMenu();
};

const ignoreExtensionType = async () => {
  let item = untrackedItemStore.selectedUntrackedItem
  await addIgnoredItem("*" + item.extension);
  emitter.emit('refresh-browser');
  menu.hideContextMenu();
};

const revealInExplorer = () => {
  let item = untrackedItemStore.selectedUntrackedItem
  FSService.RevealInExplorer(item.file_path)
  menu.hideContextMenu();
};

const extractArchive = async () => {
  menu.hideContextMenu();
  
  try {
    const selectedItem = untrackedItemStore.selectedUntrackedItem;
    
    // Only extract files (untracked_task), not folders
    if (selectedItem.type !== 'untracked_task') {
      notificationStore.errorNotification('Cannot Extract', 'Only files can be extracted');
      return;
    }
    
    // Get the file path
    const filePath = selectedItem.file_path;
    
    if (!await FSService.Exists(filePath)) {
      notificationStore.errorNotification('Cannot Extract', 'Archive file not found');
      return;
    }
    
    // Call the extraction service
    await FSService.ExtractAll(filePath)
      .then((response) => {
        notificationStore.addNotification(
          'Archive Extracted', 
          `Successfully extracted ${selectedItem.name || 'archive'}`, 
          'success'
        );
        // Refresh the browser to show the extracted folder
        emitter.emit('refresh-browser');
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

const copyItemPath = async (pathType) => {
  let item = untrackedItemStore.selectedUntrackedItem;
  console.log(item)
  let itemPath = item.file_path;
  itemPath = itemPath.replace(/\\/g, '/');
  let itemDir = itemPath.split('/').slice(0, -1).join('/');
  let resourcesFolder = itemDir + '/resources';
  let outputPath = itemDir + '/output';
  if (pathType === 'resources') {
    itemPath = resourcesFolder;
  } else if (pathType === 'output') {
    itemPath = outputPath;
  }
  await Clipboard.SetText(itemPath);
  const message = 'Path copied to clipboard';
  notificationStore.addNotification(message, "", "success");
  menu.hideContextMenu();
};

const deleteItem = async () => {
  panes.setPaneVisibility('projectDetails', true);
  let item = untrackedItemStore.selectedUntrackedItem
  if (item.type == 'untracked_task') {
    assetStore.selectedAsset = null;
    FSService.DeleteFile(item.file_path);
    projectStore.removeUntrackedTask(item.id);
  } else if (item.type == 'untracked_entity') {
    collectionStore.selectedCollection = null;
    FSService.DeleteFolder(item.file_path);
    projectStore.removeUntrackedEntity(item.id)
  }
  stage.markedItems = [];
  emitter.emit('refresh-browser')
  menu.hideContextMenu();
  modals.disableAllModals();
};

const prepDeleteItemPopUpModal = () => {
  trayStates.popUpModalTitle = "Delete";
  trayStates.popUpModalMessage = "Are you sure you want to delete this item? This will permanently remove this item. Please confirm if you wish to proceed.";
  trayStates.popUpModalIcon = 'trash';
  trayStates.popUpModalFunction = deleteItem;
  modals.setModalVisibility('popUpModal', true);
  menu.hideContextMenu();
};



// onMounted hook
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

.task-item-menu-container {
  z-index: 10;
  display: flex;
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

.task-item-menu-visible {
  /* display: flex; */
  opacity: 1;
  visibility: visible;
}
</style>






