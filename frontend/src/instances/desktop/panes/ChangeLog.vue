<template>
  <div class="general-pane-root">
      <div class="changelog-actions">
        <ActionButton :icon="getAppIcon('revert')" :label="$t('panes.discardAll')" :showLabel="true" :buttonFunction="prepDiscardAll" :useDanger="true" :useBackground="true" :isDisabled="isLoading" />
        <ActionButton :icon="getAppIcon(getCloudIcon)" :label="$t('panes.syncNow')" :showLabel="true" :buttonFunction="syncNow" :useBackground="true" :isDisabled="isLoading" />
      </div>
    <div v-if="hasChanges" class="changelog-scroll-container">

      <div v-if="summary.assets.length" class="changelog-group">
        <div class="changelog-group-header" @click="toggleGroup('assets')">
          <ActionButton :icon="getAppIcon('chevron-right')" :isMini="true" :isInactive="true" :class="{ 'chevron-expanded': expandedGroups.assets }" />
          <span class="changelog-group-title">{{ $t('panes.assets') }}</span>
          <span class="changelog-group-count">{{ summary.assets.length }}</span>
          <div class="menu-divider"></div>
        </div>

        <div v-if="expandedGroups.assets" class="changelog-group-items">
          <ChangeLogItem v-for="item in summary.assets" :key="item.id" :item="item" itemType="asset" :isLoading="isLoading" @find="(id) => findItem(id, 'asset')" @discard="(id) => discardItem(id, 'asset')" />
        </div>
      </div>

      <div v-if="summary.collections.length" class="changelog-group">
        <div class="changelog-group-header" @click="toggleGroup('collections')">
          <ActionButton :icon="getAppIcon('chevron-right')" :isMini="true" :isInactive="true" :class="{ 'chevron-expanded': expandedGroups.collections }" />
          <span class="changelog-group-title">{{ $t('panes.collections') }}</span>
          <span class="changelog-group-count">{{ summary.collections.length }}</span>
          <div class="menu-divider"></div>
        </div>

        <div v-if="expandedGroups.collections" class="changelog-group-items">
          <ChangeLogItem v-for="item in summary.collections" :key="item.id" :item="item" itemType="collection" :isLoading="isLoading" @find="(id) => findItem(id, 'collection')" @discard="(id) => discardItem(id, 'collection')" />
        </div>
      </div>

      <div v-if="summary.other.length" class="changelog-group">
        <div class="changelog-group-header" @click="toggleGroup('other')">
          <ActionButton :icon="getAppIcon('chevron-right')" :isMini="true" :isInactive="true" :class="{ 'chevron-expanded': expandedGroups.other }" />
          <span class="changelog-group-title">{{ $t('panes.other') }}</span>
          <span class="changelog-group-count">{{ summary.other.length }}</span>
          <div class="menu-divider"></div>
        </div>

        <div v-if="expandedGroups.other" class="changelog-group-items">
          <ChangeLogItem v-for="item in summary.other" :key="item.id + item.source" :item="item" itemType="other" :isLoading="isLoading" />
        </div>
      </div>
    </div>

    <PageState v-else :message="emptyMessage" :illustration="illustration" />
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import { syncData } from '@/lib/sync';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ChangeLogItem from '@/instances/desktop/components/ChangeLogItem.vue';
import PageState from '@/instances/common/components/PageState.vue';

// services
import { AssetService, CollectionService, SyncService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useStudioStore } from '@/stores/studio';
import { useTrayStates } from '@/stores/TrayStates';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const studioStore = useStudioStore();
const trayStates = useTrayStates();

const { t } = useI18n();

// refs
const expandedGroups = reactive({ assets: true, collections: true, other: false });
const isLoading = ref(false);
const summary = ref({ assets: [], collections: [], other: [], total_count: 0 });

// computed properties
const emptyMessage = computed(() => isLoading.value ? t('panes.loadingChanges') : t('panes.noPendingChanges'));

const getCloudIcon = computed(() => {
  if (!studioStore.appOnline || projectStore.getActiveProject?.is_offline) return 'cloud-cancel';
  return 'cloud-up';
});

const hasChanges = computed(() => summary.value.total_count > 0);

const illustration = computed(() => '/page-states/resources.png');

// methods

// Discards all pending changes after confirmation.
// const discardAll = async () => {
//   isLoading.value = true;
//   try {
//     await SyncService.DiscardAllChanges(projectStore.activeProject.uri, projectStore.getActiveProjectUrl);
//     projectStore.activeProject.is_unsynced = false;
//     notificationStore.addNotification(t('notifications.allChangesDiscarded'), '', 'success', false);
//     emitter.emit('refresh-browser');
//     modals.disableAllModals();
//     await loadChanges();
//   } catch (error) {
//     console.error(error);
//     notificationStore.errorNotification(t('notifications.errorDiscardingChanges'), error);
//   }
//   isLoading.value = false;
// };

// Reverts project to the remote version as of the last sync.
const revertChanges = async () => {
  isLoading.value = true;
  const syncOptions = {
    only_latest_checkpoints: false,
    asset_dependencies: false,
    assets: false,
    templates: false,
    force: true,
  };
  await SyncService.PullData(
    projectStore.activeProject.uri, projectStore.getActiveProjectUrl, false, syncOptions
  )
    .then(async () => {
      projectStore.activeProject.is_unsynced = false;
      trayStates.refreshData();
      emitter.emit('refresh-browser');
      modals.disableAllModals();
      await loadChanges();
    }).catch((error) => {
      console.error(error.message);
      notificationStore.errorNotification(t('notifications.errorDiscardingChanges'), error);
      modals.disableAllModals();
    });
  isLoading.value = false;
};

// Discards changes for a single item.
const discardItem = async (itemId, itemType) => {
  isLoading.value = true;
  try {
    await SyncService.DiscardChanges(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [itemId], itemType);
    notificationStore.addNotification(t('notifications.changeDiscarded'), '', 'success', false);
    emitter.emit('refresh-browser');
    await loadChanges();
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification(t('notifications.errorDiscardingChange'), error);
  }
  isLoading.value = false;
};

// Navigates to the item in the browser view.
const findItem = async (itemId, itemType) => {
  try {
    if (itemType === 'asset') {
      const asset = await AssetService.GetAssetByID(projectStore.activeProject.uri, itemId);
      if (!asset?.id) return;
      const assetParent = await CollectionService.GetCollectionByID(projectStore.activeProject.uri, asset.collection_id);
      if (assetParent) {
        collectionStore.navigateToCollection(assetParent);
        commonStore.navigatorMode = true;
      }
      stage.deselectAllItems();
      assetStore.selectAsset(asset.id);
      stage.firstSelectedItemId = asset.id;
      stage.markedItems = [asset.id];
    } else if (itemType === 'collection') {
      const collection = await CollectionService.GetCollectionByID(projectStore.activeProject.uri, itemId);
      if (collection) {
        collectionStore.navigateToCollection(collection);
        commonStore.navigatorMode = true;
      }
    }
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification(t('notifications.errorNavigatingToItem'), error);
  }
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Loads the pending change summary from the backend.
const loadChanges = async () => {
  if (!projectStore.activeProject?.uri) return;
  isLoading.value = true;
  try {
    summary.value = await SyncService.GetPendingChanges(projectStore.activeProject.uri);
  } catch (error) {
    console.error(error);
    summary.value = { assets: [], collections: [], other: [], total_count: 0 };
  }
  isLoading.value = false;
};

// Shows the discard all confirmation modal.
const prepDiscardAll = () => {
  trayStates.popUpModalIcon = 'revert';
  trayStates.popUpModalTitle = t('panes.discardAllChanges');
  trayStates.popUpModalMessage = t('confirmations.discardAllChanges');
  trayStates.popUpModalFunction = revertChanges;
  modals.setModalVisibility('popUpModal', true);
};

// Triggers a sync operation.
const syncNow = async () => {
  await syncData();
  await loadChanges();
};

// Toggles the expanded state of a changelog group.
const toggleGroup = (group) => {
  expandedGroups[group] = !expandedGroups[group];
};

// lifecycle hooks
onMounted(() => {
  loadChanges();
  emitter.on('refresh-browser', loadChanges);
});

onUnmounted(() => {
  emitter.off('refresh-browser', loadChanges);
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.changelog-actions {
  display: flex;
  gap: .5rem;
  padding: .5rem 0;
  width: 100%;
}

.changelog-group {
  display: flex;
  flex-direction: column;
}

.changelog-group-count {
  font-size: 12px;
  opacity: .5;
  white-space: nowrap;
}

.changelog-group-header {
  display: flex;
  align-items: center;
  gap: .5rem;
  padding: .4rem .2rem;
  cursor: pointer;
  opacity: .5;
  transition: opacity .2s ease;
}

.changelog-group-header:hover {
  opacity: 1;
}

.changelog-group-items {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  padding: .2rem 0;
}

.changelog-group-title {
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
}

.chevron-expanded {
  transform: rotate(90deg);
}

.changelog-scroll-container {
  display: flex;
  flex-direction: column;
  gap: .3rem;
  overflow-y: auto;
  padding: 0 .5rem;
  width: 100%;
  color: hsl(var(--foreground));
}

.changelog-scroll-container::-webkit-scrollbar {
  width: 4px;
}

.changelog-scroll-container::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: hsl(var(--border));
}

.changelog-scroll-container::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}
</style>
