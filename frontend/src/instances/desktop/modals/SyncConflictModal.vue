<template>
  <div class="modal-container" ref="modalContainer" v-stop-propagation>
    <HeaderArea title="Sync Conflict Detected" icon="alert-triangle" :showSearch="false" />
    <div class="general-container">
      
      <div class="conflict-message">
        <p>The following items already exist on the server (created by another user).</p>
        <p>Your local versions will be merged with the server versions.</p>
      </div>

      <div class="conflict-list" v-if="conflicts.length > 0">
        <div class="conflict-section" v-if="entityConflicts.length > 0">
          <div class="section-header">
            <img :src="getAppIcon('folder')" alt="entity" class="section-icon" />
            <span>Entities ({{ entityConflicts.length }})</span>
          </div>
          <div class="conflict-item" v-for="conflict in entityConflicts" :key="conflict.local_id">
            <span class="conflict-name">{{ conflict.name }}</span>
            <span class="conflict-arrow">→</span>
            <span class="conflict-action">will merge with server version</span>
          </div>
        </div>

        <div class="conflict-section" v-if="taskConflicts.length > 0">
          <div class="section-header">
            <img :src="getAppIcon('file')" alt="task" class="section-icon" />
            <span>Tasks ({{ taskConflicts.length }})</span>
          </div>
          <div class="conflict-item" v-for="conflict in taskConflicts" :key="conflict.local_id">
            <span class="conflict-name">{{ conflict.name }}{{ conflict.extension ? '.' + conflict.extension : '' }}</span>
            <span class="conflict-arrow">→</span>
            <span class="conflict-action">checkpoints will be added to existing task</span>
          </div>
        </div>
      </div>

      <div class="conflict-actions">
        <GeneralButton 
          label="Cancel" 
          :fullWidth="true" 
          :buttonFunction="handleCancel" 
          :colored="false" 
        />
        <GeneralButton 
          label="Merge & Sync" 
          :fullWidth="true" 
          :buttonFunction="handleMerge" 
          :isActive="true" 
          :loading="isLoading" 
        />
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useSyncConflictStore } from '@/stores/syncConflict';
import { useIconStore } from '@/stores/icons';

import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';

import { SyncService } from '@/services';

const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const syncConflictStore = useSyncConflictStore();
const iconStore = useIconStore();

const isLoading = ref(false);

const conflicts = computed(() => syncConflictStore.conflicts || []);
const projectPath = computed(() => syncConflictStore.projectPath);
const remoteURL = computed(() => syncConflictStore.remoteURL);

const entityConflicts = computed(() => 
  conflicts.value.filter(c => c.type === 'entity')
);

const taskConflicts = computed(() => 
  conflicts.value.filter(c => c.type === 'task')
);

const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

const handleCancel = () => {
  syncConflictStore.clearConflicts();
  modals.disableAllModals();
};

const handleMerge = async () => {
  if (isLoading.value) return;
  
  isLoading.value = true;
  
  try {
    // Pass conflicts as JSON string to avoid Wails binding issues
    const conflictsJSON = JSON.stringify(conflicts.value);
    await SyncService.ResolveConflicts(projectPath.value, conflictsJSON);
    
    notificationStore.addNotification(
      'Conflicts Resolved', 
      `${conflicts.value.length} item(s) merged successfully. Retrying sync...`
    );
    
    // Close modal and clear conflicts
    syncConflictStore.clearConflicts();
    modals.disableAllModals();
    
    // Retry sync
    let syncOptions = {
      only_latest_checkpoints: false,
      task_dependencies: false,
      tasks: false,
      templates: false,
    };
    
    // await SyncService.SyncData(projectPath.value, remoteURL.value, false, syncOptions);
    
    // notificationStore.successNotification('Sync Complete', 'Data synced successfully');
    
  } catch (error) {
    console.error('Failed to resolve conflicts:', error);
    notificationStore.errorNotification('Merge Failed', error.message || 'Failed to resolve conflicts');
  } finally {
    isLoading.value = false;
  }
};

onMounted(() => {
  console.log('SyncConflictModal mounted with conflicts:', conflicts.value);
});

onBeforeUnmount(() => {
  // Don't clear conflicts here in case user wants to retry
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

.conflict-list {
  max-height: 300px;
  overflow-y: auto;
  padding: 0 0.5rem;
  margin-bottom: 1rem;
}

.conflict-section {
  margin-bottom: 1rem;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  background-color: var(--charcoal);
  border-radius: 4px;
  margin-bottom: 0.5rem;
  font-size: 12px;
  font-weight: 600;
  color: var(--white);
}

.section-icon {
  width: 16px;
  height: 16px;
}

.conflict-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  background-color: var(--charcoal-light);
  border-radius: 4px;
  margin-bottom: 0.25rem;
  font-size: 12px;
}

.conflict-name {
  color: var(--amber);
  font-weight: 500;
  flex-shrink: 0;
}

.conflict-arrow {
  color: var(--gray);
  flex-shrink: 0;
}

.conflict-action {
  color: var(--gray);
  font-size: 11px;
}

.conflict-actions {
  display: flex;
  gap: 0.5rem;
  padding: 0.5rem;
}
</style>
