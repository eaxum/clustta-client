<template>
  <div class="modal-container large-modal" v-esc="closeModal">
    <HeaderArea :title="'Sync Preview'" :icon="'cloud-sync'" />

    <div class="general-container">
      <!-- Loading State -->
      <div v-if="isLoading" class="loading-state">
        <span>{{ loadingMessage }}</span>
      </div>

      <!-- Syncing Progress -->
      <ProgressSection v-else-if="isSyncing" />

      <!-- Sync Preview -->
      <div v-else-if="!error" class="step-content">
        <div class="preview-header">
          <span class="preview-summary">{{ assetsToCreate }} Assets in {{ collectionsToCreate }} Collections</span>
          <ActionButton :icon="getAppIcon('refresh')" v-tooltip="'Refresh'" :buttonFunction="loadSyncPreview" />
        </div>

        <div class="preview-divider"></div>

        <!-- Tree View -->
        <div class="sync-preview-scroll">
          <div v-if="syncPreviewTree.length === 0" class="empty-preview">
            <img :src="getAppIcon('check-circle')" alt="" class="empty-icon" />
            <span class="empty-text">Everything is up to date</span>
          </div>
          <div v-else class="preview-tree-content">
            <PreviewVirtuaItem v-for="item in syncPreviewTree" :key="item.id" :item="item" 
              :depth="0" :itemHeight="48" :expandedItems="expandedItems" 
              @toggle-expand="toggleExpand" />
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
        <GeneralButton :label="'Create'" :fullWidth="true" :buttonFunction="executeSync" :isActive="hasItemsToCreate && !isLoading" :loading="isSyncing" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import PreviewVirtuaItem from '@/instances/common/components/PreviewVirtuaItem.vue';
import ProgressSection from '@/instances/common/components/ProgressSection.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useNotificationStore } from '@/stores/notifications';
import { useStageStore } from '@/stores/stages';
import { useTemplateStore } from '@/stores/template';

const { t } = useI18n();
const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const stageStore = useStageStore();
const templateStore = useTemplateStore();

// refs
const error = ref(null);
const expandedItems = ref(new Set());
const isLoading = ref(false);
const isSyncing = ref(false);
const loadingMessage = ref('');

// computed
// Returns all assets from sync preview.
const assets = computed(() => integrationStore.assetsToSync);

// Returns count of assets to create (excludes existing).
const assetsToCreate = computed(() => assets.value.filter(a => a.action === 'create').length);

// Returns all collections from sync preview.
const collections = computed(() => integrationStore.collectionsToSync);

// Returns count of collections to create (excludes existing).
const collectionsToCreate = computed(() => collections.value.filter(c => c.action === 'create').length);

// Checks if there are items to create.
const hasItemsToCreate = computed(() => collectionsToCreate.value > 0 || assetsToCreate.value > 0);

// Returns the integration name.
const integrationName = computed(() => {
  const id = integrationStore.linkedIntegrationId;
  const integration = integrationStore.getIntegration(id);
  return integration?.name || id;
});



// Returns the hierarchical tree for sync preview.
const syncPreviewTree = computed(() => integrationStore.syncPreviewTree);

// methods
// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Executes sync - creates all items from the preview.
const executeSync = async () => {
  if (!hasItemsToCreate.value) return;

  isSyncing.value = true;
  stageStore.operationActive = true;
  try {
    await integrationStore.executeSync();
    emitter.emit('refresh-browser');
    closeModal();
  } catch (err) {
    // Error handled by store
  } finally {
    stageStore.operationActive = false;
    isSyncing.value = false;
  }
};

// Returns the app icon path.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Loads sync preview, auto-creating type mappings.
const loadSyncPreview = async () => {
  isLoading.value = true;
  loadingMessage.value = 'Loading...';
  error.value = null;

  try {
    // Load external types from the integration
    loadingMessage.value = 'Fetching types from ' + integrationName.value + '...';
    await integrationStore.getExternalTypes();

    // Load existing type mappings
    await integrationStore.loadTypeMappings();

    // Auto-generate 1:1 type mappings for any unmapped types
    const collectionTypeMappingsMap = { ...(integrationStore.typeMappings?.collection_type_mappings || {}) };
    for (const type of integrationStore.externalCollectionTypes) {
      if (!collectionTypeMappingsMap[type.name]) {
        collectionTypeMappingsMap[type.name] = {
          external_name: type.name,
          external_id: type.id,
          clustta_name: type.name,
          clustta_icon: 'folder',
        };
      }
    }

    const assetTypeMappingsMap = { ...(integrationStore.typeMappings?.asset_type_mappings || {}) };
    for (const type of integrationStore.externalAssetTypes) {
      if (!assetTypeMappingsMap[type.name]) {
        assetTypeMappingsMap[type.name] = {
          external_name: type.name,
          external_id: type.id,
          clustta_name: type.name,
          clustta_icon: 'generic',
        };
      }
    }

    // Save auto-generated mappings (preserving existing directory_structure and asset_type_templates)
    loadingMessage.value = 'Saving type mappings...';
    await integrationStore.saveTypeMappings({
      ...integrationStore.typeMappings,
      collection_type_mappings: collectionTypeMappingsMap,
      asset_type_mappings: assetTypeMappingsMap,
    });

    // Load the sync preview (missing types are auto-created during ExecuteSync)
    loadingMessage.value = 'Fetching data from ' + integrationName.value + '...';
    await integrationStore.getSyncPreview();

    // Load templates for extension display
    await templateStore.reloadTemplates();

    
  } catch (err) {
    console.error('loadSyncPreview error:', err);
    error.value = err.message || 'Failed to load sync preview';
  } finally {
    isLoading.value = false;
  }
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
  width: 50vw;
  min-width: 600px;
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

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.preview-summary {
  font-size: 15px;
}

.preview-divider {
  width: 100%;
  height: 1px;
  background-color: var(--steel);
  flex-shrink: 0;
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
