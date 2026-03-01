<template>
  <div class="modal-container large-modal" v-esc="closeModal">
    <HeaderArea :title="'Sync Preview'" :icon="'cloud-sync'" />

    <div class="general-container">
      <!-- Loading State -->
      <div v-if="isLoading" class="loading-state">
        <span>{{ loadingMessage }}</span>
      </div>

      <!-- Sync Preview -->
      <div v-else-if="!error" class="step-content">
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
        </div>

        <!-- Tree View -->
        <div class="sync-preview-scroll">
          <div v-if="syncPreviewTree.length === 0" class="empty-preview">
            <img :src="getAppIcon('check-circle')" alt="" class="empty-icon" />
            <span class="empty-text">Everything is up to date</span>
          </div>
          <div v-else class="preview-tree-content">
            <PreviewVirtuaItem v-for="item in syncPreviewTree" :key="item.id" :item="item" 
              :depth="0" :itemHeight="36" :expandedItems="expandedItems" :selectedItems="selectedItemsSet" 
              @toggle-expand="toggleExpand" @toggle-selection="toggleSelection" />
          </div>
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="error-state">
        <p>{{ error }}</p>
        <GeneralButton :label="'Retry'" :buttonFunction="loadSyncPreview" />
      </div>

      <!-- Actions -->
      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Sync Selected'" :fullWidth="true" :buttonFunction="executeSync" :isActive="hasSelection" :loading="isSyncing" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import PreviewVirtuaItem from '@/instances/common/components/PreviewVirtuaItem.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useNotificationStore } from '@/stores/notifications';
import { useTemplateStore } from '@/stores/template';

const { t } = useI18n();
const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const templateStore = useTemplateStore();

// refs
const error = ref(null);
const expandedItems = ref(new Set());
const isLoading = ref(false);
const isSyncing = ref(false);
const loadingMessage = ref('');
const selectedItems = ref([]);

// computed
// Returns all assets from sync preview.
const assets = computed(() => integrationStore.assetsToSync);

// Returns count of assets to create (excludes existing).
const assetsToCreate = computed(() => assets.value.filter(a => a.action === 'create').length);

// Returns all collections from sync preview.
const collections = computed(() => integrationStore.collectionsToSync);

// Returns count of collections to create (excludes existing).
const collectionsToCreate = computed(() => collections.value.filter(c => c.action === 'create').length);

// Checks if any items are selected.
const hasSelection = computed(() => selectedItems.value.length > 0);

// Returns the integration name.
const integrationName = computed(() => {
  const id = integrationStore.linkedIntegrationId;
  const integration = integrationStore.getIntegration(id);
  return integration?.name || id;
});

// Returns selected items as a Set for efficient lookup.
const selectedItemsSet = computed(() => new Set(selectedItems.value));

// Returns the hierarchical tree for sync preview.
const syncPreviewTree = computed(() => integrationStore.syncPreviewTree);

// methods
// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Executes sync for selected items.
const executeSync = async () => {
  if (!hasSelection.value) return;

  // Split selected items into collections and assets
  const collectionIds = new Set(collections.value.map(c => c.external_id));
  const selectedCollectionIds = selectedItems.value.filter(id => collectionIds.has(id));
  const selectedAssetIds = selectedItems.value.filter(id => !collectionIds.has(id));

  isSyncing.value = true;
  try {
    await integrationStore.executeSync(selectedCollectionIds, selectedAssetIds);
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

// Loads sync preview, auto-creating type mappings and missing types.
const loadSyncPreview = async () => {
  isLoading.value = true;
  loadingMessage.value = 'Loading...';
  error.value = null;

  try {
    // Load external types from the integration
    loadingMessage.value = 'Fetching types from ' + integrationName.value + '...';
    await integrationStore.getExternalTypes();

    // Load local types from Clustta
    await integrationStore.getLocalTypes();

    // Load existing type mappings
    await integrationStore.loadTypeMappings();

    // Auto-generate 1:1 type mappings for any unmapped types
    const entityTypeMappingsMap = { ...(integrationStore.typeMappings?.entity_type_mappings || {}) };
    for (const type of integrationStore.externalEntityTypes) {
      if (!entityTypeMappingsMap[type.name]) {
        entityTypeMappingsMap[type.name] = {
          external_name: type.name,
          external_id: type.id,
          clustta_name: type.name,
          clustta_icon: 'folder',
        };
      }
    }

    const taskTypeMappingsMap = { ...(integrationStore.typeMappings?.task_type_mappings || {}) };
    for (const type of integrationStore.externalTaskTypes) {
      if (!taskTypeMappingsMap[type.name]) {
        taskTypeMappingsMap[type.name] = {
          external_name: type.name,
          external_id: type.id,
          clustta_name: type.name,
          clustta_icon: 'generic',
        };
      }
    }

    // Save auto-generated mappings (preserving existing directory_structure and task_type_templates)
    loadingMessage.value = 'Saving type mappings...';
    await integrationStore.saveTypeMappings({
      ...integrationStore.typeMappings,
      entity_type_mappings: entityTypeMappingsMap,
      task_type_mappings: taskTypeMappingsMap,
    });

    // Get and auto-create missing types
    await integrationStore.getMissingTypes();
    const missingTypes = integrationStore.missingTypes || { entity_types: [], task_types: [] };
    const hasMissingTypes = (missingTypes.entity_types?.length || 0) > 0 || (missingTypes.task_types?.length || 0) > 0;

    if (hasMissingTypes) {
      loadingMessage.value = 'Creating missing types...';
      await integrationStore.createMissingTypes();
    }

    // Load the sync preview
    loadingMessage.value = 'Fetching data from ' + integrationName.value + '...';
    await integrationStore.getSyncPreview();

    // Load templates for extension display
    await templateStore.reloadTemplates();

    // Pre-select all new items
    selectAllNew();
  } catch (err) {
    console.error('loadSyncPreview error:', err);
    error.value = err.message || 'Failed to load sync preview';
  } finally {
    isLoading.value = false;
  }
};

// Selects all items that will be created (excludes existing items).
const selectAllNew = () => {
  const collectionIds = collections.value.filter(c => c.action === 'create').map(c => c.external_id);
  const assetIds = assets.value.filter(a => a.action === 'create').map(a => a.external_id);
  selectedItems.value = [...collectionIds, ...assetIds];
};

// Toggles item expand state.
const toggleExpand = (itemId) => {
  const newSet = new Set(expandedItems.value);
  if (newSet.has(itemId)) {
    newSet.delete(itemId);
  } else {
    newSet.add(itemId);
  }
  expandedItems.value = newSet;
};

// Toggles item selection state.
const toggleSelection = (itemId) => {
  const index = selectedItems.value.indexOf(itemId);
  if (index === -1) {
    selectedItems.value = [...selectedItems.value, itemId];
  } else {
    selectedItems.value = selectedItems.value.filter(id => id !== itemId);
  }
};

// lifecycle
onMounted(() => {
  loadSyncPreview();
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.modal-container {
  max-height: 80vh;
  max-width: 90vw;
}

.general-container {
  overflow: hidden;
  display: flex;
  flex-direction: column;
  width: 90vw;
  max-width: 900px;
  box-sizing: border-box;
}

.step-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
  box-sizing: border-box;
  overflow: hidden;
  overflow-y: auto;
  width: 100%;
}

.step-content::-webkit-scrollbar {
  width: 4px;
}

.step-content::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.step-content::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.sync-summary {
  display: flex;
  justify-content: space-around;
  padding: 16px;
  background: var(--surface-primary);
  border-radius: var(--small-radius);
  flex-shrink: 0;
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

.sync-preview-scroll {
  display: flex;
  flex-direction: column;
  max-height: 350px;
  overflow-y: auto;
  border-radius: var(--small-radius);
  background: var(--surface-secondary);
}

.sync-preview-scroll::-webkit-scrollbar {
  width: 4px;
}

.sync-preview-scroll::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.sync-preview-scroll::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.preview-tree-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 4px;
}

.empty-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  gap: 12px;
}

.empty-icon {
  width: 48px;
  height: 48px;
  opacity: 0.4;
}

.empty-text {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}
</style>
