<template>
  <div class="modal-container large-modal" v-esc="closeModal">
    <HeaderArea :title="title" :icon="'sync'" />

    <div class="general-container">
      <!-- Loading State -->
      <div v-if="isLoading" class="loading-state">
        <span>Fetching data from {{ integrationName }}...</span>
      </div>

      <!-- Sync Preview Content -->
      <div v-else-if="syncPreview" class="sync-content">
        <!-- Summary -->
        <div class="sync-summary">
          <div class="summary-item">
            <span class="summary-count">{{ collectionsToCreate }}</span>
            <span class="summary-label">New Collections</span>
          </div>
          <div class="summary-item">
            <span class="summary-count">{{ assetsToCreate }}</span>
            <span class="summary-label">New Assets</span>
          </div>
          <div class="summary-item">
            <span class="summary-count">{{ unchangedCount }}</span>
            <span class="summary-label">Unchanged</span>
          </div>
        </div>

        <!-- Tabs -->
        <div class="sync-tabs">
          <button class="tab-button" :class="{ active: activeTab === 'collections' }" @click="activeTab = 'collections'">
            Collections ({{ collections.length }})
          </button>
          <button class="tab-button" :class="{ active: activeTab === 'assets' }" @click="activeTab = 'assets'">
            Assets ({{ assets.length }})
          </button>
        </div>

        <!-- Collections List -->
        <div v-if="activeTab === 'collections'" class="sync-list">
          <div v-for="item in collections" :key="item.external_id" class="sync-item"
            :class="{ 'sync-create': item.action === 'create', 'sync-update': item.action === 'update' }">
            <input type="checkbox" v-model="selectedCollections" :value="item.external_id" :disabled="item.action === 'unchanged'" />
            <div class="item-info">
              <span class="item-name">{{ item.external_name }}</span>
              <span class="item-path">{{ item.external_path }}</span>
            </div>
            <span class="item-action" :class="item.action">{{ item.action }}</span>
          </div>
          <div v-if="collections.length === 0" class="empty-list">No collections to sync</div>
        </div>

        <!-- Assets List -->
        <div v-if="activeTab === 'assets'" class="sync-list">
          <div v-for="item in assets" :key="item.external_id" class="sync-item"
            :class="{ 'sync-create': item.action === 'create', 'sync-update': item.action === 'update' }">
            <input type="checkbox" v-model="selectedAssets" :value="item.external_id" :disabled="item.action === 'unchanged'" />
            <div class="item-info">
              <span class="item-name">{{ item.external_name }}</span>
              <span class="item-type">{{ item.external_type }}</span>
            </div>
            <span class="item-action" :class="item.action">{{ item.action }}</span>
          </div>
          <div v-if="assets.length === 0" class="empty-list">No assets to sync</div>
        </div>

        <!-- Selection Controls -->
        <div class="selection-controls">
          <ActionButton :icon="getAppIcon('select-all')" :label="'Select All New'" :buttonFunction="selectAllNew" />
          <ActionButton :icon="getAppIcon('deselect')" :label="'Clear Selection'" :buttonFunction="clearSelection" />
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="error-state">
        <p>{{ error }}</p>
        <GeneralButton :label="'Retry'" :buttonFunction="loadPreview" />
      </div>

      <!-- Actions -->
      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Sync Selected'" :fullWidth="true" @click="executeSync"
          :isActive="hasSelection" :loading="isSyncing" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useNotificationStore } from '@/stores/notifications';

const { t } = useI18n();
const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();

// refs
const activeTab = ref('collections');
const error = ref(null);
const isLoading = ref(false);
const isSyncing = ref(false);
const selectedAssets = ref([]);
const selectedCollections = ref([]);

// constants
const title = 'Sync Preview';

// computed
// Returns all assets from sync preview.
const assets = computed(() => integrationStore.assetsToSync);

// Returns count of assets to create.
const assetsToCreate = computed(() => assets.value.filter(a => a.action === 'create').length);

// Returns all collections from sync preview.
const collections = computed(() => integrationStore.collectionsToSync);

// Returns count of collections to create.
const collectionsToCreate = computed(() => collections.value.filter(c => c.action === 'create').length);

// Checks if any items are selected.
const hasSelection = computed(() => selectedCollections.value.length > 0 || selectedAssets.value.length > 0);

// Returns the integration name.
const integrationName = computed(() => {
  const id = integrationStore.linkedIntegrationId;
  const integration = integrationStore.getIntegration(id);
  return integration?.name || id;
});

// Returns the sync preview data.
const syncPreview = computed(() => integrationStore.syncPreview);

// Returns count of unchanged items.
const unchangedCount = computed(() => {
  const unchangedCollections = collections.value.filter(c => c.action === 'unchanged').length;
  const unchangedAssets = assets.value.filter(a => a.action === 'unchanged').length;
  return unchangedCollections + unchangedAssets;
});

// methods
// Clears all selections.
const clearSelection = () => {
  selectedCollections.value = [];
  selectedAssets.value = [];
};

// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Executes sync for selected items.
const executeSync = async () => {
  if (!hasSelection.value) return;

  isSyncing.value = true;
  try {
    await integrationStore.executeSync(selectedCollections.value, selectedAssets.value);
    closeModal();
  } catch (err) {
    // Error handled by store
  } finally {
    isSyncing.value = false;
  }
};

// Returns the app icon path.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Loads the sync preview.
const loadPreview = async () => {
  isLoading.value = true;
  error.value = null;

  try {
    await integrationStore.getSyncPreview();
    // Pre-select items to create
    selectAllNew();
  } catch (err) {
    error.value = err.message || 'Failed to load sync preview';
  } finally {
    isLoading.value = false;
  }
};

// Selects all items that need to be created.
const selectAllNew = () => {
  selectedCollections.value = collections.value
    .filter(c => c.action === 'create')
    .map(c => c.external_id);
  selectedAssets.value = assets.value
    .filter(a => a.action === 'create')
    .map(a => a.external_id);
};

// lifecycle
onMounted(() => {
  loadPreview();
});
</script>

<style scoped>
.large-modal {
  width: 600px;
  max-height: 80vh;
}

.sync-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.sync-summary {
  display: flex;
  justify-content: space-around;
  padding: 16px;
  background: var(--surface-primary);
  border-radius: var(--small-radius);
}

.summary-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.summary-count {
  font-size: 24px;
  font-weight: 600;
  color: var(--accent-primary);
}

.summary-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.sync-tabs {
  display: flex;
  gap: 4px;
  padding: 4px;
  background: var(--surface-primary);
  border-radius: var(--small-radius);
}

.tab-button {
  flex: 1;
  padding: 8px 16px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  border-radius: var(--small-radius);
  cursor: pointer;
  transition: all 0.15s;
}

.tab-button.active {
  background: var(--surface-secondary);
  color: var(--text-primary);
}

.sync-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-height: 300px;
  overflow-y: auto;
}

.sync-list::-webkit-scrollbar {
  width: 4px;
}

.sync-list::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.sync-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  background: var(--surface-primary);
  border-radius: var(--small-radius);
  border-left: 3px solid transparent;
}

.sync-item.sync-create {
  border-left-color: var(--color-success);
}

.sync-item.sync-update {
  border-left-color: var(--color-warning);
}

.sync-item input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.sync-item input[type="checkbox"]:disabled {
  opacity: 0.5;
  cursor: default;
}

.item-info {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}

.item-name {
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-path,
.item-type {
  font-size: 11px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-action {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: var(--small-radius);
  text-transform: capitalize;
}

.item-action.create {
  background: var(--color-success-subtle);
  color: var(--color-success);
}

.item-action.update {
  background: var(--color-warning-subtle);
  color: var(--color-warning);
}

.item-action.unchanged {
  background: var(--surface-secondary);
  color: var(--text-tertiary);
}

.selection-controls {
  display: flex;
  justify-content: center;
  gap: 16px;
  padding-top: 8px;
}

.empty-list {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 32px;
  color: var(--text-secondary);
}

.loading-state,
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px;
  gap: 16px;
  color: var(--text-secondary);
}
</style>
