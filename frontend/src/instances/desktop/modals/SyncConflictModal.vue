<template>
  <div class="modal-container" ref="modalContainer" v-stop-propagation>
    <HeaderArea title="Sync Conflict Detected" :icon="getAppIcon('cloud-error')" :showSearch="false" />

    <div class="general-container">
      <div class="conflict-message">
        <p>The following items already exist on the server (created by another user).</p>

        <p>Your local versions will be merged with the server versions.</p>
      </div>

      <div class="conflict-tabs-header">
        <div class="conflict-tabs">
          <PaneHeaderTabs :dataTypes="filterTabs" :selectedTab="selectedFilter" @filter="handleFilterChange" />
        </div>

        <div class="conflict-tabs-options">
          <ActionButton :icon="hideExtensions ? getAppIcon('extension-cancel') : getAppIcon('extension')" v-tooltip="hideExtensions ? 'Show extensions' : 'Hide extensions'" :buttonFunction="toggleHideExtensions" />

          <ActionButton :icon="showFullPath ? getAppIcon('file-name') : getAppIcon('file-path')" v-tooltip="showFullPath ? 'Name' : 'Path'" :buttonFunction="toggleShowFullPath" />
        </div>
      </div>

      <div class="conflict-list-container conflict-list-empty" v-if="isEnriching">
        <div class="conflict-loading">
          <img :src="getAppIcon('loading')" alt="loading" class="loading-icon" />

          <span>Loading conflict details...</span>
        </div>
      </div>

      <div class="conflict-list-container conflict-list-empty" v-else-if="!filteredConflicts.length">
        <PageState :message="'No conflicts to resolve'" :illustration="'/page-states/resources.png'" />
      </div>

      <div class="conflict-list-container" v-else-if="filteredConflicts.length > 0">
        <div class="conflict-list">
          <ConflictItem v-for="conflict in filteredConflicts" :key="conflict.local_id" :conflict="conflict" :hideExtensions="hideExtensions" :showFullPath="showFullPath" @resolved="handleConflictResolved" @merge="handleSingleMerge" />
        </div>
      </div>

      <div class="pop-up-actions" :class="{ 'pop-up-actions-end': !hasConflicts }">
        <GeneralButton v-if="hasConflicts" label="Cancel" :fullWidth="true" :buttonFunction="handleCancel" :colored="false" />

        <GeneralButton :label="hasConflicts ? 'Merge All' : 'Done'" :fullWidth="true" :buttonFunction="hasConflicts ? handleMerge : handleDone" :isActive="true" :loading="isLoading" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ConflictItem from '@/instances/desktop/components/ConflictItem.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import PageState from '@/instances/common/components/PageState.vue';
import PaneHeaderTabs from '@/instances/common/components/PaneHeaderTabs.vue';

// services
import { AssetService, CollectionService, SyncService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useSyncConflictStore } from '@/stores/syncConflict';

const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const syncConflictStore = useSyncConflictStore();

// refs
const enrichedConflicts = ref([]);
const hideExtensions = ref(true);
const isEnriching = ref(false);
const isLoading = ref(false);
const selectedFilter = ref('all');
const showFullPath = ref(false);

// constants
const filterTabs = [
  { name: 'all', icon: 'list' },
  { name: 'assets', icon: 'file' },
  { name: 'collections', icon: 'folder' },
];

// computed properties (dependencies first, then alphabetically)
const conflicts = computed(() => enrichedConflicts.value.length > 0 ? enrichedConflicts.value : syncConflictStore.conflicts || []);

const entityConflicts = computed(() => conflicts.value.filter(c => c.type === 'entity'));

const taskConflicts = computed(() => conflicts.value.filter(c => c.type === 'task'));

const filteredConflicts = computed(() => {
  if (selectedFilter.value === 'all') {
    return conflicts.value;
  } else if (selectedFilter.value === 'assets') {
    return taskConflicts.value;
  } else if (selectedFilter.value === 'collections') {
    return entityConflicts.value;
  }
  return conflicts.value;
});

const hasConflicts = computed(() => enrichedConflicts.value.length > 0);

const projectPath = computed(() => syncConflictStore.projectPath);

// methods
// Fetches full data for conflicts from DB using parallel requests.
const enrichConflicts = async () => {
  const rawConflicts = syncConflictStore.conflicts || [];
  if (rawConflicts.length === 0) return;

  isEnriching.value = true;
  const projectUri = projectStore.activeProject?.uri || projectPath.value;

  try {
    const entityIds = rawConflicts.filter(c => c.type === 'entity').map(c => c.local_id);
    const taskIds = rawConflicts.filter(c => c.type === 'task').map(c => c.local_id);

    const [entityResults, taskResults] = await Promise.all([
      Promise.all(entityIds.map(id => 
        CollectionService.GetCollectionByID(projectUri, id).catch(err => {
          console.warn(`Failed to fetch entity ${id}:`, err);
          return null;
        })
      )),
      Promise.all(taskIds.map(id => 
        AssetService.GetAssetByID(projectUri, id).catch(err => {
          console.warn(`Failed to fetch task ${id}:`, err);
          return null;
        })
      ))
    ]);

    const entityMap = new Map();
    entityResults.forEach((entity, idx) => {
      if (entity && entity.id) {
        entityMap.set(entityIds[idx], entity);
      }
    });

    const taskMap = new Map();
    taskResults.forEach((task, idx) => {
      if (task && task.id) {
        taskMap.set(taskIds[idx], task);
      }
    });

    enrichedConflicts.value = rawConflicts.map(conflict => {
      if (conflict.type === 'entity') {
        const entityData = entityMap.get(conflict.local_id);
        if (entityData) {
          return {
            ...conflict,
            ...entityData,
            local_id: conflict.local_id,
            server_id: conflict.server_id,
            type: conflict.type,
          };
        }
      } else if (conflict.type === 'task') {
        const taskData = taskMap.get(conflict.local_id);
        if (taskData) {
          return {
            ...conflict,
            ...taskData,
            local_id: conflict.local_id,
            server_id: conflict.server_id,
            type: conflict.type,
          };
        }
      }
      return conflict;
    });
  } catch (error) {
    console.error('Failed to enrich conflicts:', error);
    enrichedConflicts.value = rawConflicts;
  } finally {
    isEnriching.value = false;
  }
};

// Returns icon path from icon store.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Cancels conflict resolution and closes modal.
const handleCancel = () => {
  syncConflictStore.clearConflicts();
  modals.disableAllModals();
};

// Removes resolved conflict and any child conflicts from list.
const handleConflictResolved = (resolvedConflict) => {
  enrichedConflicts.value = enrichedConflicts.value.filter(c => c.local_id !== resolvedConflict.local_id);
  
  if (resolvedConflict.type === 'entity' && resolvedConflict.entity_path) {
    const parentPath = resolvedConflict.entity_path;
    enrichedConflicts.value = enrichedConflicts.value.filter(c => !c.entity_path?.startsWith(parentPath));
  }
};

// Closes modal and refreshes browser view.
const handleDone = () => {
  syncConflictStore.clearConflicts();
  modals.disableAllModals();
  emitter.emit('refresh-browser');
};

// Updates selected filter tab.
const handleFilterChange = (filter) => {
  selectedFilter.value = filter;
};

// Merges all remaining conflicts with server versions.
const handleMerge = async () => {
  if (isLoading.value) return;
  
  isLoading.value = true;
  
  try {
    const conflictsJSON = JSON.stringify(conflicts.value);
    await SyncService.ResolveConflicts(projectPath.value, conflictsJSON);
    
    notificationStore.addNotification(
      'Conflicts Resolved', 
      `${conflicts.value.length} item(s) merged successfully.`,
      'success'
    );
    
    enrichedConflicts.value = [];
    syncConflictStore.clearConflicts();
  } catch (error) {
    console.error('Failed to resolve conflicts:', error);
    notificationStore.errorNotification('Merge Failed', error.message || 'Failed to resolve conflicts');
  } finally {
    isLoading.value = false;
  }
};

// Merges a single conflict item with server version.
const handleSingleMerge = async (conflict) => {
  try {
    const conflictsJSON = JSON.stringify([conflict]);
    await SyncService.ResolveConflicts(projectPath.value, conflictsJSON);
    
    enrichedConflicts.value = enrichedConflicts.value.filter(c => c.local_id !== conflict.local_id);
    
    notificationStore.addNotification(
      'Merged Successfully',
      `${conflict.type === 'entity' ? 'Collection' : 'Asset'} "${conflict.name}" merged`,
      'success'
    );
  } catch (error) {
    console.error('Failed to merge conflict:', error);
    notificationStore.errorNotification('Merge Failed', error.message || 'Failed to merge item');
  }
};

// Toggles file extension visibility in conflict list.
const toggleHideExtensions = () => {
  hideExtensions.value = !hideExtensions.value;
};

// Toggles between showing name only or full path.
const toggleShowFullPath = () => {
  showFullPath.value = !showFullPath.value;
};

// lifecycle hooks
onMounted(async () => {
  await enrichConflicts();
});

onBeforeUnmount(() => {
  enrichedConflicts.value = [];
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.conflict-message {
  padding: 0.5rem 1rem;
  margin-bottom: 1rem;
}

.conflict-message p {
  font-size: 13px;
  color: var(--white);
  margin: 0.25rem 0;
}

.conflict-list-container {
  max-height: 300px;
  min-height: 300px;
  box-sizing: border-box;
  width: 100%;
  overflow: hidden;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  margin-bottom: 1rem;
  background-color: var(--midnight-steel);
  border-radius: var(--very-large-radius);
}

.conflict-list-empty {
  align-items: center;
  justify-content: center;
}
.conflict-list-container::-webkit-scrollbar {
  width: 4px;
}

.conflict-list-container::-webkit-scrollbar-thumb {
  border-radius: 8px;
  background-color: var(--light-steel);
}

.conflict-list-container::-webkit-scrollbar-track {
  margin-top: 5px;
  border-radius: 10px;
}

.conflict-list {
  width: 100%;
  box-sizing: border-box;
  padding: 0 0.5rem;
  display: flex;
  flex-direction: column;
  gap: .5rem;
  padding: .5rem;
}

.conflict-tabs-header {
  margin-bottom: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  border-radius: var(--very-large-radius);
}

.conflict-tabs {
  padding: 0.3rem 0.5rem;
  display: flex;
  background-color: var(--midnight-steel);
  width: 100%;
  max-width: 250px;
  border-radius: var(--very-large-radius);
}

.conflict-tabs-options {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0 0.5rem;
}

.conflict-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 2rem;
  color: var(--gray);
  font-size: 13px;
}

.loading-icon {
  width: 16px;
  height: 16px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.pop-up-actions {
  padding: 0px;
  margin-top: 0;
}

.pop-up-actions-end {
  justify-content: flex-end;
}
</style>
