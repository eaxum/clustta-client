<template>
  <div class="modal-container" ref="modalContainer" v-stop-propagation>
    <HeaderArea :title="$t('modals.resolveConflicts')" :icon="getAppIcon('cloud-error')" :showSearch="false" />

    <div class="general-container">
      <div class="conflict-message">
        <p>{{ $t('modals.conflictsDescription') }}</p>

        <p>{{ $t('modals.conflictsInstruction') }} <span class="learn-more-link" @click="openLearnMore">{{ $t('modals.learnMore') }} <ActionButton :icon="getAppIcon('square-arrow-right-up')" :allowDeactivate="true" :isMini="true" /></span></p>
      </div>

      <div class="conflict-tabs-header">
        <div class="conflict-tabs">
          <PaneHeaderTabs :dataTypes="filterTabs" :selectedTab="selectedFilter" @filter="handleFilterChange" />
        </div>

        <div class="conflict-tabs-options">
          <ActionButton :icon="hideExtensions ? getAppIcon('extension-cancel') : getAppIcon('extension')" v-tooltip="hideExtensions ? $t('modals.showExtensions') : $t('modals.hideExtensions')" :buttonFunction="toggleHideExtensions" />

          <ActionButton :icon="showFullPath ? getAppIcon('file-name') : getAppIcon('file-path')" v-tooltip="showFullPath ? $t('modals.nameColumn') : $t('modals.pathColumn')" :buttonFunction="toggleShowFullPath" />
        </div>
      </div>

      <div class="conflict-list-container conflict-list-empty" v-if="isEnriching">
        <div class="conflict-loading">
          <img :src="getAppIcon('loading')" alt="loading" class="loading-icon" />

          <span>{{ $t('modals.loadingConflicts') }}</span>
        </div>
      </div>

      <div class="conflict-list-container conflict-list-empty" v-else-if="!filteredConflicts.length">
        <PageState :message="$t('modals.noConflicts')" :illustration="'/page-states/resources.png'" />
      </div>

      <div class="conflict-list-container" v-else-if="filteredConflicts.length > 0">
        <div class="conflict-list">
          <ConflictItem v-for="conflict in filteredConflicts" :key="conflict.local_id" :conflict="conflict" :hideExtensions="hideExtensions" :showFullPath="showFullPath" @resolved="handleRenameResolved" @merge="handleSingleMerge" />
        </div>
      </div>

      <div class="pop-up-actions" :class="{ 'pop-up-actions-end': !hasConflicts }">
        <GeneralButton v-if="hasConflicts" :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="handleCancel" :colored="false" />

        <GeneralButton :label="hasConflicts ? $t('modals.mergeAll') : $t('common.done')" :fullWidth="true" :buttonFunction="hasConflicts ? handleMergeAll : handleDone" :isActive="true" :loading="isLoading" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import {
  filterTopLevelConflicts,
  getResolutionSummary,
  prepareRecursiveMergeConflicts,
} from '@/lib/conflictUtils';
import { Browser } from "@wailsio/runtime";

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

const { t } = useI18n();
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
  { name: 'all', nameKey: 'common.all', icon: 'list' },
  { name: 'assets', nameKey: 'panes.assets', icon: 'file' },
  { name: 'collections', nameKey: 'panes.collections', icon: 'folder' },
];

// computed properties (dependencies first, then alphabetically)
const allEnrichedConflicts = computed(() => 
  enrichedConflicts.value.length > 0 ? enrichedConflicts.value : syncConflictStore.allConflicts || []
);

const topLevelConflicts = computed(() => filterTopLevelConflicts(allEnrichedConflicts.value));

const collectionConflicts = computed(() => topLevelConflicts.value.filter(c => c.type === 'collection'));

const filteredConflicts = computed(() => {
  if (selectedFilter.value === 'all') {
    return topLevelConflicts.value;
  } else if (selectedFilter.value === 'assets') {
    return assetConflicts.value;
  } else if (selectedFilter.value === 'collections') {
    return collectionConflicts.value;
  }
  return topLevelConflicts.value;
});

const hasConflicts = computed(() => enrichedConflicts.value.length > 0);

const projectPath = computed(() => syncConflictStore.projectPath);

const assetConflicts = computed(() => topLevelConflicts.value.filter(c => c.type === 'asset'));

// methods
// Fetches full data for conflicts from DB using parallel requests.
const enrichConflicts = async () => {
  const rawConflicts = syncConflictStore.allConflicts || [];
  if (rawConflicts.length === 0) return;

  isEnriching.value = true;
  const projectUri = projectStore.activeProject?.uri || projectPath.value;

  try {
    const collectionIds = rawConflicts.filter(c => c.type === 'collection').map(c => c.local_id);
    const assetIds = rawConflicts.filter(c => c.type === 'asset').map(c => c.local_id);

    const [collectionResults, assetResults] = await Promise.all([
      Promise.all(collectionIds.map(id => 
        CollectionService.GetCollectionByID(projectUri, id).catch(err => {
          console.warn(`Failed to fetch collection ${id}:`, err);
          return null;
        })
      )),
      Promise.all(assetIds.map(id => 
        AssetService.GetAssetByID(projectUri, id).catch(err => {
          console.warn(`Failed to fetch asset ${id}:`, err);
          return null;
        })
      ))
    ]);

    const collectionMap = new Map();
    collectionResults.forEach((collection, idx) => {
      if (collection && collection.id) {
        collectionMap.set(collectionIds[idx], collection);
      }
    });

    const assetMap = new Map();
    assetResults.forEach((asset, idx) => {
      if (asset && asset.id) {
        assetMap.set(assetIds[idx], asset);
      }
    });

    enrichedConflicts.value = rawConflicts.map(conflict => {
      if (conflict.type === 'collection') {
        const collectionData = collectionMap.get(conflict.local_id);
        if (collectionData) {
          return {
            ...conflict,
            ...collectionData,
            local_id: conflict.local_id,
            server_id: conflict.server_id,
            type: conflict.type,
          };
        }
      } else if (conflict.type === 'asset') {
        const assetData = assetMap.get(conflict.local_id);
        if (assetData) {
          return {
            ...conflict,
            ...assetData,
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

// Opens the sync documentation page.
const openLearnMore = () => {
  Browser.OpenURL('https://docs.clustta.com/guide/assignations-dependencies-and-syncing.html#sync-conflicts');
};

// Merges all remaining conflicts with server versions.
const handleMergeAll = async () => {
  if (isLoading.value) return;
  
  isLoading.value = true;
  
  try {
    const conflictsJSON = JSON.stringify(allEnrichedConflicts.value);
    await SyncService.ResolveConflicts(projectPath.value, conflictsJSON);
    
    notificationStore.addNotification(
      t('notifications.conflictsResolved'), 
      t('notifications.itemsMergedSuccessfully', { count: allEnrichedConflicts.value.length }),
      'success'
    );
    
    enrichedConflicts.value = [];
    syncConflictStore.clearConflicts();
  } catch (error) {
    console.error('Failed to resolve conflicts:', error);
    notificationStore.errorNotification(t('notifications.mergeFailed'), error.message || t('notifications.failedToResolveConflicts'));
  } finally {
    isLoading.value = false;
  }
};

// Handles rename resolution for a conflict (including recursive child resolution).
const handleRenameResolved = (conflict) => {
  const summary = getResolutionSummary(conflict, allEnrichedConflicts.value, 'rename');
  
  // Remove the resolved conflict and all its children from enriched list
  const conflictsToRemove = prepareRecursiveMergeConflicts(conflict, allEnrichedConflicts.value);
  const idsToRemove = conflictsToRemove.map(c => c.local_id);
  
  enrichedConflicts.value = enrichedConflicts.value.filter(c => !idsToRemove.includes(c.local_id));
  syncConflictStore.removeConflicts(idsToRemove);
  
  notificationStore.addNotification(t('notifications.renamedSuccessfully'), summary, 'success');
};

// Merges a single conflict item with server version (including children recursively).
const handleSingleMerge = async (conflict) => {
  try {
    // Get parent + all children for recursive merge
    const conflictsToMerge = prepareRecursiveMergeConflicts(conflict, allEnrichedConflicts.value);
    const conflictsJSON = JSON.stringify(conflictsToMerge);
    
    await SyncService.ResolveConflicts(projectPath.value, conflictsJSON);
    
    // Remove merged conflicts from local state
    const idsToRemove = conflictsToMerge.map(c => c.local_id);
    enrichedConflicts.value = enrichedConflicts.value.filter(c => !idsToRemove.includes(c.local_id));
    syncConflictStore.removeConflicts(idsToRemove);
    
    const summary = getResolutionSummary(conflict, allEnrichedConflicts.value, 'merge');
    notificationStore.addNotification(t('notifications.mergedSuccessfully'), summary, 'success');
  } catch (error) {
    console.error('Failed to merge conflict:', error);
    notificationStore.errorNotification(t('notifications.mergeFailed'), error.message || t('notifications.failedToMergeItem'));
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
  padding: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: .5rem;
  font-weight: 300;
}

.conflict-message p {
  font-size: 13px;
  color: hsl(var(--foreground));
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
  background-color: hsl(var(--card));
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
  border-radius: var(--normal-radius);
  background-color: hsl(var(--border));
}

.conflict-list-container::-webkit-scrollbar-track {
  margin-top: 5px;
  border-radius: var(--large-radius);
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
  background-color: hsl(var(--card));
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

.learn-more-link {
  color: var(--blue);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
}

.learn-more-link:hover {
  text-decoration: underline;
}
</style>
