<template>
  <div class="general-pane-header">
    <HeaderArea :notModal="true" title="Change Log" />
    <ActionButton :icon="getAppIcon('refresh')" :showLabel="false" v-tooltip="'Refresh'" :buttonFunction="loadChanges" />
  </div>

  <div class="general-pane-root">
    <div v-if="hasChanges" class="changelog-scroll-container">
      <div class="changelog-actions">
        <ActionButton :icon="getAppIcon('revert')" label="Discard All" :showLabel="true" :buttonFunction="prepDiscardAll" :useDanger="true" :useBackground="true" :isDisabled="isLoading" />
        <ActionButton :icon="getAppIcon(getCloudIcon)" label="Sync Now" :showLabel="true" :buttonFunction="syncNow" :useBackground="true" :isDisabled="isLoading" />
      </div>

      <div v-if="summary.tasks.length" class="changelog-group">
        <div class="changelog-group-header" @click="toggleGroup('tasks')">
          <img class="changelog-chevron" :class="{ 'chevron-expanded': expandedGroups.tasks }" :src="getAppIcon('chevron-right')" />
          <span class="changelog-group-title">Tasks</span>
          <span class="changelog-group-count">{{ summary.tasks.length }}</span>
        </div>

        <div v-if="expandedGroups.tasks" class="changelog-group-items">
          <div v-for="item in summary.tasks" :key="item.id" class="changelog-item">
            <div class="changelog-item-info">
              <span class="changelog-change-type" :class="'change-' + item.change_type">{{ item.change_type }}</span>
              <span class="changelog-item-name">{{ item.name || item.id }}</span>
            </div>
            <ActionButton :icon="getAppIcon('revert')" :showLabel="false" v-tooltip="'Discard'" :buttonFunction="() => discardItem(item.id, 'task')" :isDisabled="isLoading" />
          </div>
        </div>
      </div>

      <div v-if="summary.entities.length" class="changelog-group">
        <div class="changelog-group-header" @click="toggleGroup('entities')">
          <img class="changelog-chevron" :class="{ 'chevron-expanded': expandedGroups.entities }" :src="getAppIcon('chevron-right')" />
          <span class="changelog-group-title">Collections</span>
          <span class="changelog-group-count">{{ summary.entities.length }}</span>
        </div>

        <div v-if="expandedGroups.entities" class="changelog-group-items">
          <div v-for="item in summary.entities" :key="item.id" class="changelog-item">
            <div class="changelog-item-info">
              <span class="changelog-change-type" :class="'change-' + item.change_type">{{ item.change_type }}</span>
              <span class="changelog-item-name">{{ item.name || item.id }}</span>
            </div>
            <ActionButton :icon="getAppIcon('revert')" :showLabel="false" v-tooltip="'Discard'" :buttonFunction="() => discardItem(item.id, 'entity')" :isDisabled="isLoading" />
          </div>
        </div>
      </div>

      <div v-if="summary.other.length" class="changelog-group">
        <div class="changelog-group-header" @click="toggleGroup('other')">
          <img class="changelog-chevron" :class="{ 'chevron-expanded': expandedGroups.other }" :src="getAppIcon('chevron-right')" />
          <span class="changelog-group-title">Other</span>
          <span class="changelog-group-count">{{ summary.other.length }}</span>
        </div>

        <div v-if="expandedGroups.other" class="changelog-group-items">
          <div v-for="item in summary.other" :key="item.id + item.source" class="changelog-item">
            <div class="changelog-item-info">
              <span class="changelog-change-type" :class="'change-' + item.change_type">{{ item.change_type }}</span>
              <span class="changelog-item-name">{{ item.source }}: {{ item.name || item.id }}</span>
            </div>
          </div>
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
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import PageState from '@/instances/common/components/PageState.vue';

// services
import { SyncService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStudioStore } from '@/stores/studio';
import { useTrayStates } from '@/stores/TrayStates';

const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
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

.changelog-chevron {
  width: 12px;
  height: 12px;
  transition: transform 0.2s ease;
  filter: brightness(0) invert(1);
  opacity: .5;
}

.changelog-change-type {
  font-size: 11px;
  font-weight: 500;
  padding: 1px 6px;
  border-radius: 4px;
  text-transform: uppercase;
  white-space: nowrap;
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
  padding-left: .5rem;
}

.changelog-group-title {
  font-size: 13px;
  font-weight: 500;
}

.changelog-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: .3rem .4rem;
  border-radius: var(--small-radius);
}

.changelog-item:hover {
  background-color: var(--light-steel);
}

.changelog-item-info {
  display: flex;
  align-items: center;
  gap: .5rem;
  overflow: hidden;
}

.changelog-item-name {
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.changelog-scroll-container {
  display: flex;
  flex-direction: column;
  gap: .3rem;
  overflow-y: auto;
  padding: 0 .5rem;
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

.change-deleted {
  background-color: rgba(220, 50, 50, 0.2);
  color: #f87171;
}

.change-modified {
  background-color: rgba(59, 130, 246, 0.2);
  color: #60a5fa;
}

.chevron-expanded {
  transform: rotate(90deg);
}
</style>
