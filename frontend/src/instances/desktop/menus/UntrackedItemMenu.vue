<template>
  <div ref="popUpMenu" class="filter-menu-container">

    <!-- Launch -->
    <ActionButton
      v-if="userStore.canDo('pull_chunk') && isUntrackedAsset"
      :icon="getAppIcon('launch')" :showLabel="true" :fullWidth="true" :label="$t('common.openWith')"
      :buttonFunction="launchAssetWithCommand" />

    <!-- <span v-if="userStore.canDo('pull_chunk')" class="menu-divider"></span> -->

    <!-- Rename -->
    <ActionButton v-if="canRenameUntracked" :icon="getAppIcon('edit')" :showLabel="true" :fullWidth="true"
      :label="$t('common.rename')" :buttonFunction="renameItem" />

    <!-- Ignore -->
    <ActionButton :icon="getAppIcon('file-watch')" :showLabel="true" :fullWidth="true" :label="$t('menus.ignoreFileFolder')"
      :buttonFunction="ignoreItem" />

    <ActionButton v-if="isUntrackedAsset"
      :icon="getAppIcon('file-watch')" :showLabel="true" :fullWidth="true" :label="$t('menus.ignoreExtensionType')"
      :buttonFunction="ignoreExtensionType" />

    <!-- Reveal in Explorer -->
    <span class="horizontal-flex">
      <ActionButton :icon="getAppIcon('folder-arrow-up-right')" :showLabel="true" :fullWidth="true" :label="showLabel"
        :buttonFunction="revealInExplorer" />
      <ActionButton :icon="getAppIcon('copy')" :showLabel="false" :fullWidth="false" @click="copyItemPath('asset')"
        v-tooltip="$t('common.copyPath')" />
    </span>

    <!-- Extract Archive -->
    <ActionButton v-if="isArchive" :icon="getAppIcon('unarchive')" :showLabel="true" :fullWidth="true" 
      :label="$t('common.extract')" :buttonFunction="extractArchive" />

    <!-- Delete -->
    <ActionButton :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true" :label="$t('common.delete')"
      :buttonFunction="prepDeleteItemPopUpModal" />

  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { Clipboard } from '@wailsio/runtime';
import emitter from '@/lib/mitt';
import { addIgnoredItem } from '@/lib/untracked';
import { canActInNavigatedCollection } from '@/lib/permissions';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// composables
import { useRevealLabel } from '@/composables/useRevealLabel';

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

const { t } = useI18n();
const { showLabel } = useRevealLabel();

// refs
const popUpMenu = ref(null);

// props
const props = defineProps({
  top: Number,
  left: Number,
});

const emit = defineEmits(['clicked']);

// computed
// Checks if the selected untracked item is an untracked asset.
const isUntrackedAsset = computed(() => {
  return untrackedItemStore.selectedUntrackedItem?.type === 'untracked_asset';
});

// Whether the user can rename the selected untracked item.
const canRenameUntracked = computed(() => canActInNavigatedCollection('update_asset'));

// Checks if the selected untracked item is an archive.
const isArchive = computed(() => {
  const archiveFormats = ['.zip', '.rar', '.7z', '.tar', '.gz', '.bz2'];
  const item = untrackedItemStore.selectedUntrackedItem;
  if (item?.type !== 'untracked_asset') {
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
  notificationStore.addNotification(t('notifications.pathCopied'), "", "success");
  menu.hideContextMenu();
};

// Deletes the selected untracked item.
const deleteItem = async () => {
  panes.setPaneVisibility('projectDetails', true);
  let item = untrackedItemStore.selectedUntrackedItem;
  if (item.type == 'untracked_asset') {
    assetStore.selectedAsset = null;
    FSService.DeleteFile(item.file_path);
    projectStore.removeUntrackedAsset(item.id);
  } else if (item.type == 'untracked_collection') {
    collectionStore.selectedCollection = null;
    FSService.DeleteFolder(item.file_path);
    projectStore.removeUntrackedCollection(item.id);
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
    
    if (selectedItem.type !== 'untracked_asset') {
      notificationStore.errorNotification(t('notifications.cannotExtract'), t('notifications.onlyFilesExtracted'));
      return;
    }
    
    const filePath = selectedItem.file_path;
    
    if (!await FSService.Exists(filePath)) {
      notificationStore.errorNotification(t('notifications.cannotExtract'), t('notifications.archiveNotFound'));
      return;
    }
    
    await FSService.ExtractAll(filePath)
      .then(() => {
        notificationStore.addNotification(t('notifications.archiveExtracted'), `Successfully extracted ${selectedItem.name || 'archive'}`, 'success');
        emitter.emit('refresh-browser');
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
  if (item.type == "untracked_asset") {
    await addIgnoredItem(item.asset_path);
  } else {
    const untrackedCollection = removeLastSlash(item.item_path);
    await addIgnoredItem(untrackedCollection);
  }
  emitter.emit('refresh-browser');
  menu.hideContextMenu();
};

// Launches the item with the system's default application.
const launchAssetWithCommand = async () => {
  let item = untrackedItemStore.selectedUntrackedItem;
  let file_path = item.file_path;
  if (await FSService.Exists(file_path)) {
    FSService.LaunchFileWith(file_path);
  } else {
    notificationStore.addNotification(t('notifications.fileNotOnDisk'), t('notifications.fileNotOnDiskDesc'), "error");
  }
  menu.hideContextMenu();
};

// Prepares and shows the delete confirmation modal.
const prepDeleteItemPopUpModal = () => {
  trayStates.popUpModalTitle = t('common.delete');
  trayStates.popUpModalMessage = t('confirmations.deleteItemPermanently');
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
  emitter.emit(isUntrackedAsset.value ? 'renameAsset' : 'renameCollection');
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






