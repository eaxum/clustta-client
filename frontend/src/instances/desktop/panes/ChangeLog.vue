<template>
  <div class="general-pane-root">
    <div v-if="hasChanges" class="changelog-scroll-container">
      <div class="changelog-actions">
        <ActionButton :icon="getAppIcon('revert')" label="Discard All" :showLabel="true" :buttonFunction="prepDiscardAll" :useDanger="true" :useBackground="true" :isDisabled="isLoading" />
        <ActionButton :icon="getAppIcon(getCloudIcon)" label="Sync Now" :showLabel="true" :buttonFunction="syncNow" :useBackground="true" :isDisabled="isLoading" />
      </div>

      <div v-if="summary.tasks.length" class="changelog-group">
        <div class="changelog-group-header" @click="toggleGroup('tasks')">
          <ActionButton :icon="getAppIcon('chevron-right')" :isMini="true" :isInactive="true" :class="{ 'chevron-expanded': expandedGroups.tasks }" />
          <span class="changelog-group-title">Tasks</span>
          <span class="changelog-group-count">{{ summary.tasks.length }}</span>
        </div>

        <div v-if="expandedGroups.tasks" class="changelog-group-items">
          <ChangeLogItem v-for="item in summary.tasks" :key="item.id" :item="item" itemType="task" :isLoading="isLoading" @find="(id) => findItem(id, 'task')" @discard="(id) => discardItem(id, 'task')" />
        </div>
      </div>

      <div v-if="summary.entities.length" class="changelog-group">
        <div class="changelog-group-header" @click="toggleGroup('entities')">
          <ActionButton :icon="getAppIcon('chevron-right')" :isMini="true" :isInactive="true" :class="{ 'chevron-expanded': expandedGroups.entities }" />
          <span class="changelog-group-title">Collections</span>
          <span class="changelog-group-count">{{ summary.entities.length }}</span>
        </div>

        <div v-if="expandedGroups.entities" class="changelog-group-items">
          <ChangeLogItem v-for="item in summary.entities" :key="item.id" :item="item" itemType="entity" :isLoading="isLoading" @find="(id) => findItem(id, 'entity')" @discard="(id) => discardItem(id, 'entity')" />
        </div>
      </div>

      <div v-if="summary.other.length" class="changelog-group">
        <div class="changelog-group-header" @click="toggleGroup('other')">
          <ActionButton :icon="getAppIcon('chevron-right')" :isMini="true" :isInactive="true" :class="{ 'chevron-expanded': expandedGroups.other }" />
          <span class="changelog-group-title">Other</span>
          <span class="changelog-group-count">{{ summary.other.length }}</span>
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

// refs
const expandedGroups = reactive({ tasks: true, entities: true, other: false });
const isLoading = ref(false);
const summary = ref({ tasks: [], entities: [], other: [], total_count: 0 });

// computed properties
const emptyMessage = computed(() => isLoading.value ? 'Loading changes...' : 'No pending changes');

const getCloudIcon = computed(() => {
  if (!studioStore.appOnline || projectStore.getActiveProject?.is_offline) return 'cloud-cancel';
  return 'cloud-up';
});

const hasChanges = computed(() => summary.value.total_count > 0);

const illustration = computed(() => '/page-states/resources.png');

// methods

// Discards all pending changes after confirmation.
const discardAll = async () => {
  isLoading.value = true;
  try {
    await SyncService.DiscardAllChanges(projectStore.activeProject.uri, projectStore.getActiveProjectUrl);
    projectStore.activeProject.is_unsynced = false;
    notificationStore.addNotification('All changes discarded', '', 'success', false);
    emitter.emit('refresh-browser');
    modals.disableAllModals();
    await loadChanges();
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification('Error discarding changes', error);
  }
  isLoading.value = false;
};

// Discards changes for a single item.
const discardItem = async (itemId, itemType) => {
  isLoading.value = true;
  try {
    await SyncService.DiscardChanges(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [itemId], itemType);
    notificationStore.addNotification('Change discarded', '', 'success', false);
    emitter.emit('refresh-browser');
    await loadChanges();
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification('Error discarding change', error);
  }
  isLoading.value = false;
};

// Navigates to the item in the browser view.
const findItem = async (itemId, itemType) => {
  try {
    if (itemType === 'task') {
      const task = await AssetService.GetAssetByID(projectStore.activeProject.uri, itemId);
      if (!task?.id) return;
      const taskParent = await CollectionService.GetCollectionByID(projectStore.activeProject.uri, task.entity_id);
      if (taskParent) {
        collectionStore.navigateToCollection(taskParent);
        commonStore.navigatorMode = true;
      }
      stage.deselectAllItems();
      assetStore.selectAsset(task.id);
      stage.firstSelectedItemId = task.id;
      stage.markedItems = [task.id];
    } else if (itemType === 'entity') {
      const entity = await CollectionService.GetCollectionByID(projectStore.activeProject.uri, itemId);
      if (entity) {
        collectionStore.navigateToCollection(entity);
        commonStore.navigatorMode = true;
      }
    }
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification('Error navigating to item', error);
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
    summary.value = { tasks: [], entities: [], other: [], total_count: 0 };
  }
  isLoading.value = false;
};

// Shows the discard all confirmation modal.
const prepDiscardAll = () => {
  trayStates.popUpModalIcon = 'revert';
  trayStates.popUpModalTitle = 'Discard All Changes';
  trayStates.popUpModalMessage = 'All unsynced changes will be reverted to the remote version. This cannot be undone. Continue?';
  trayStates.popUpModalFunction = discardAll;
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
}

.changelog-group {
  display: flex;
  flex-direction: column;
}

.changelog-group-count {
  font-size: 12px;
  opacity: .5;
  margin-left: auto;
}

.changelog-group-header {
  display: flex;
  align-items: center;
  gap: .5rem;
  padding: .4rem .2rem;
  cursor: pointer;
  border-radius: var(--small-radius);
}

.changelog-group-header:hover {
  background-color: var(--light-steel);
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
  color: var(--white);
}

.changelog-scroll-container::-webkit-scrollbar {
  width: 4px;
}

.changelog-scroll-container::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.changelog-scroll-container::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}
</style>
