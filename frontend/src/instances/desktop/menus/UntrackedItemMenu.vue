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
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { Clipboard } from '@wailsio/runtime';
import emitter from '@/lib/mitt';
import { addIgnoredItem } from '@/lib/untracked';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// services
import { FSService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { usePaneStore } from '@/stores/panes';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTrayStates } from '@/stores/TrayStates';
import { useUntrackedItemStore } from '@/stores/untracked';
import { useUserStore } from '@/stores/users';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const panes = usePaneStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const trayStates = useTrayStates();
const untrackedItemStore = useUntrackedItemStore();
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
// Checks if the selected untracked item is an archive.
const isArchive = computed(() => {
  const archiveFormats = ['.zip', '.rar', '.7z', '.tar', '.gz', '.bz2'];
  const item = untrackedItemStore.selectedUntrackedItem;
  if (item?.type !== 'untracked_task') {
    return false;
  }
  const extension = item?.extension?.toLowerCase() || '';
  return archiveFormats.includes(extension);
});

// methods
// Copies the item path to clipboard.
const copyItemPath = async (pathType) => {
  let item = untrackedItemStore.selectedUntrackedItem;
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
  notificationStore.addNotification('Path copied to clipboard', "", "success");
  menu.hideContextMenu();
};

// Deletes the selected untracked item.
const deleteItem = async () => {
  panes.setPaneVisibility('projectDetails', true);
  let item = untrackedItemStore.selectedUntrackedItem;
  if (item.type == 'untracked_task') {
    assetStore.selectedAsset = null;
    FSService.DeleteFile(item.file_path);
    projectStore.removeUntrackedTask(item.id);
  } else if (item.type == 'untracked_entity') {
    collectionStore.selectedCollection = null;
    FSService.DeleteFolder(item.file_path);
    projectStore.removeUntrackedEntity(item.id);
  }
  stage.markedItems = [];
  emitter.emit('refresh-browser');
  menu.hideContextMenu();
  modals.disableAllModals();
};

// Extracts the archive file.
const extractArchive = async () => {
  menu.hideContextMenu();
  
  try {
    const selectedItem = untrackedItemStore.selectedUntrackedItem;
    
    if (selectedItem.type !== 'untracked_task') {
      notificationStore.errorNotification('Cannot Extract', 'Only files can be extracted');
      return;
    }
    
    const filePath = selectedItem.file_path;
    
    if (!await FSService.Exists(filePath)) {
      notificationStore.errorNotification('Cannot Extract', 'Archive file not found');
      return;
    }
    
    await FSService.ExtractAll(filePath)
      .then(() => {
        notificationStore.addNotification('Archive Extracted', `Successfully extracted ${selectedItem.name || 'archive'}`, 'success');
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

// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Adds the item's extension to the ignore list.
const ignoreExtensionType = async () => {
  let item = untrackedItemStore.selectedUntrackedItem;
  await addIgnoredItem("*" + item.extension);
  emitter.emit('refresh-browser');
  menu.hideContextMenu();
};

// Adds the item to the ignore list.
const ignoreItem = async () => {
  let item = untrackedItemStore.selectedUntrackedItem;
  if (item.type == "untracked_task") {
    await addIgnoredItem(item.task_path);
  } else {
    const untrackedEntity = removeLastSlash(item.item_path);
    await addIgnoredItem(untrackedEntity);
  }
  emitter.emit('refresh-browser');
  menu.hideContextMenu();
};

// Launches the item with the system's default application.
const launchTaskWithCommand = async () => {
  let item = untrackedItemStore.selectedUntrackedItem;
  let file_path = item.file_path;
  if (await FSService.Exists(file_path)) {
    FSService.LaunchFileWith(file_path);
  } else {
    notificationStore.addNotification("File Not On Disk, Rebuild", "File not found on disk, rebuild task", "error");
  }
  menu.hideContextMenu();
};

// Prepares and shows the delete confirmation modal.
const prepDeleteItemPopUpModal = () => {
  trayStates.popUpModalTitle = "Delete";
  trayStates.popUpModalMessage = "Are you sure you want to delete this item? This will permanently remove this item. Please confirm if you wish to proceed.";
  trayStates.popUpModalIcon = 'trash';
  trayStates.popUpModalFunction = deleteItem;
  modals.setModalVisibility('popUpModal', true);
  menu.hideContextMenu();
};

// Removes the last slash from a path string.
const removeLastSlash = (text) => {
  const lastSlashIndex = text.lastIndexOf('/');
  if (lastSlashIndex !== -1) {
    return text.slice(0, lastSlashIndex) + text.slice(lastSlashIndex + 1);
  }
  return text;
};

// Emits event to rename the item.
const renameItem = () => {
  emitter.emit('renameAsset');
  menu.hideContextMenu();
};

// Reveals the item in the file explorer.
const revealInExplorer = () => {
  let item = untrackedItemStore.selectedUntrackedItem;
  FSService.RevealInExplorer(item.file_path);
  menu.hideContextMenu();
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






