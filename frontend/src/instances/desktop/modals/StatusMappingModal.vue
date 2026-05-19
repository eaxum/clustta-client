<template>
  <div class="modal-container large-modal" v-esc="closeModal">
    <HeaderArea :title="'Status Mapping'" :icon="'clock'" />

    <div class="general-container">
      <!-- Loading State -->
      <div v-if="isLoading" class="loading-state">
        <span>Loading...</span>
      </div>

      <!-- No Statuses Warning -->
      <div v-else-if="localStatuses.length === 0" class="empty-state">
        <img :src="getAppIcon('warning')" alt="" class="empty-icon" />
        <span class="empty-title">No Statuses Found</span>
        <span class="empty-description">Create statuses in your project before mapping them to the external integration.</span>
      </div>

      <div v-else class="mapping-content">
        <div class="section-header">
          <p class="section-description">
            Map each Clustta status to a {{ integrationName }} status. When a checkpoint is created, the mapped status will be pushed automatically.
          </p>
          <ActionButton :icon="getAppIcon('sparkles')" :label="'Auto'" :buttonFunction="autoAssign" :showLabel="true" :useBackground="true" />
        </div>

        <!-- Mapping Table -->
        <div class="mapping-table">
          <div class="table-header">
            <span class="col-local">Clustta Status</span>
            <span class="col-external">{{ integrationName }} Status</span>
          </div>

          <div class="table-body">
            <div v-for="status in localStatuses" :key="status.id" class="mapping-row">
              <div class="col-local">
                <span class="status-dot" :style="{ backgroundColor: status.color }"></span>
                <span class="type-name">{{ status.name }}</span>
              </div>
              <div class="col-external">
                <DropDownBox :items="externalStatusOptions" :selectedItem="getSelectedExternalName(status.id)"
                  :onSelect="(val) => setMapping(status.id, val)" :placeHolder="'None'"
                  :useFilter="true" :fullWidth="true" />
              </div>
            </div>
          </div>
        </div>

        <!-- Unmapped Info -->
        <div v-if="unmappedCount > 0" class="warning-banner">
          <img :src="getAppIcon('alert')" alt="" class="warning-icon" />
          <span>{{ unmappedCount }} status{{ unmappedCount > 1 ? 'es' : '' }} not mapped. Unmapped statuses won't be pushed on checkpoint.</span>
        </div>
      </div>
    </div>

    <!-- Actions -->
    <div class="pop-up-actions">
      <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
      <GeneralButton :label="'Save'" :fullWidth="true" :buttonFunction="saveMapping" :isActive="true" :loading="isSaving" />
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useNotificationStore } from '@/stores/notifications';

const { t } = useI18n();
const desktopModals = useDesktopModalStore();
const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const notificationStore = useNotificationStore();

// refs
const isLoading = ref(false);
const isSaving = ref(false);
const mappings = ref({});

// computed
// Returns the display name of the linked integration.
const integrationName = computed(() => {
  const id = integrationStore.linkedIntegration?.integration_id || '';
  return id.charAt(0).toUpperCase() + id.slice(1);
});

// Returns local Clustta statuses.
const localStatuses = computed(() => {
  return integrationStore.localStatuses || [];
});

// Returns external statuses formatted for the dropdown.
const externalStatusOptions = computed(() => {
  return (integrationStore.externalStatuses || []).map(s => ({
    id: s.id,
    name: s.name,
  }));
});

// Returns the count of unmapped statuses.
const unmappedCount = computed(() => {
  return localStatuses.value.filter(s => !mappings.value[s.id]).length;
});

// methods
// Attempts to auto-assign external statuses by matching names.
const autoAssign = () => {
  for (const status of localStatuses.value) {
    if (mappings.value[status.id]) continue;

    const localName = status.name.toLowerCase().trim();
    const match = externalStatusOptions.value.find(ext =>
      ext.name.toLowerCase().includes(localName) || localName.includes(ext.name.toLowerCase())
    );
    if (match) {
      mappings.value[status.id] = match.id;
    }
  }
};

// Closes the modal.
const closeModal = () => {
  desktopModals.setModalVisibility('statusMappingModal', false);
};

// Returns the app icon path.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Gets the currently selected external status name for a local status.
const getSelectedExternalName = (localStatusId) => {
  const externalId = mappings.value[localStatusId];
  if (!externalId) return null;
  const ext = externalStatusOptions.value.find(s => s.id === externalId);
  return ext ? ext.name : null;
};

// Sets a mapping between a local status and an external status.
const setMapping = (localStatusId, selectedName) => {
  const ext = externalStatusOptions.value.find(s => s.name === selectedName);
  if (ext) {
    mappings.value[localStatusId] = ext.id;
  } else {
    delete mappings.value[localStatusId];
  }
};

// Saves the status mappings.
const saveMapping = async () => {
  isSaving.value = true;
  try {
    await integrationStore.saveStatusMappings(mappings.value);
    closeModal();
  } catch (error) {
    notificationStore.addNotification(error.message || 'Failed to save', '', 'error');
  } finally {
    isSaving.value = false;
  }
};

// lifecycle hooks
onMounted(async () => {
  isLoading.value = true;
  try {
    await integrationStore.loadLinkedIntegration();
    await integrationStore.loadTypeMappings();

    // Load both local and external statuses
    await Promise.all([
      integrationStore.getLocalStatuses(),
      integrationStore.getExternalStatuses(),
    ]);

    // Load existing mappings from sync options
    const existing = integrationStore.typeMappings?.status_mappings || {};
    mappings.value = { ...existing };
  } catch (error) {
    console.error('Failed to load status mappings:', error);
    notificationStore.addNotification(error.message || 'Failed to load statuses', '', 'error');
  } finally {
    isLoading.value = false;
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.modal-container {
  max-height: 80vh;
  max-width: 500px;
}

.general-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1rem;
  overflow-y: auto;
  width: 500px;
  max-width: 500px;
}

.general-container::-webkit-scrollbar {
  width: 4px;
}

.general-container::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-5);
}

.general-container::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  color: var(--text);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 2rem;
  text-align: center;
}

.empty-icon {
  width: 48px;
  height: 48px;
  opacity: 0.5;
}

.empty-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
}

.empty-description {
  font-size: 0.875rem;
  color: var(--surface-5);
}

.mapping-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.section-description {
  font-size: 0.875rem;
  color: var(--surface-5);
  margin: 0;
}

.mapping-table {
  display: flex;
  flex-direction: column;
  outline: var(--transparent-line);
  outline-offset: -1px;
  border-radius: var(--large-radius);
  overflow: hidden;
}

.table-header {
  display: flex;
  padding: 0.75rem 1rem;
  background-color: var(--bg);
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--surface-5);
  text-transform: uppercase;
}

.table-body {
  display: flex;
  flex-direction: column;
  max-height: 300px;
  overflow-y: auto;
}

.table-body::-webkit-scrollbar {
  width: 4px;
}

.table-body::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-5);
}

.mapping-row {
  display: flex;
  padding: 0.75rem 1rem;
  border-top: 1px solid var(--surface-3);
  align-items: center;
}

.mapping-row:hover {
  background-color: var(--surface-3);
}

.col-local {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.col-external {
  flex: 1;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.type-name {
  font-size: 0.875rem;
  color: var(--text);
}

.warning-banner {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background-color: rgba(255, 193, 7, 0.1);
  border: 1px solid var(--attention);
  border-radius: var(--small-radius);
  font-size: 0.875rem;
  color: var(--attention);
}

.warning-icon {
  width: 16px;
  height: 16px;
}

.pop-up-actions {
  display: flex;
  gap: 0.5rem;
  padding: 0.5rem 1rem 1rem 1rem;
  width: 500px;
  box-sizing: border-box;
}
</style>
