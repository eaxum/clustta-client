<template>
  <div class="modal-container large-modal" v-esc="closeModal">
    <HeaderArea :title="$t('kitsu.importTitle')" icon="kitsu" />

    <div class="general-container">
      <!-- Loading State -->
      <div v-if="isLoading" class="loading-state">
        <span>{{ loadingMessage }}</span>
      </div>

      <!-- Syncing Progress -->
      <ProgressSection v-else-if="isSyncing" />

      <div v-else-if="readinessState" class="readiness-state">
        <p>{{ readinessMessage }}</p>
        <GeneralButton :label="readinessActionLabel" :buttonFunction="resolveReadiness" />
      </div>

      <!-- Sync Preview -->
      <div v-else-if="!error" class="step-content">
        <div class="preview-header">
          <div class="preview-summary">
            <span>{{ $t('kitsu.newCount', { count: selectedCreateCount }) }}</span>
            <span>{{ $t('kitsu.linkCount', { count: selectedLinkCount }) }}</span>
            <span>{{ $t('kitsu.noActionCount', { count: noActionCount }) }}</span>
          </div>
          <div class="preview-actions">
            <button type="button" @click="selectAll">{{ $t('kitsu.selectAll') }}</button>
            <button type="button" @click="unselectAll">{{ $t('kitsu.clearSelection') }}</button>
          </div>
          <ActionButton :icon="getAppIcon('refresh')" v-tooltip="$t('kitsu.refresh')" :buttonFunction="loadSyncPreview" />
        </div>

        <div class="preview-divider"></div>

        <!-- Tree View -->
        <div class="sync-preview-scroll">
          <div v-if="syncPreviewTree.length === 0" class="empty-preview">
            <img :src="getAppIcon('check-circle')" alt="" class="empty-icon" />
            <span class="empty-text">{{ $t('kitsu.emptyImport') }}</span>
          </div>
          <div v-else class="preview-tree-content">
            <PreviewVirtuaItem v-for="item in syncPreviewTree" :key="item.id" :item="item" 
              :depth="0" :itemHeight="48" :expandedItems="expandedItems" 
              :selectedItems="selectedKeys" :canSelectItem="canSelectItem" @toggle-expand="toggleExpand"
              @toggle-selection="toggleSelection" />
          </div>
        </div>

        <MissingTypesAlert v-if="selectedMissingTypes.length" :missingTypes="selectedMissingTypes"
          integrationName="Kitsu" :expanded="showMissingTypes" @toggle="showMissingTypes = !showMissingTypes" />
        <p v-if="permissionMessage" class="permission-message">{{ permissionMessage }}</p>
        <p v-if="executionError" class="inline-error">{{ executionError }}</p>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="error-state">
        <p>{{ error }}</p>
        <GeneralButton :label="$t('kitsu.retry')" :buttonFunction="loadSyncPreview" />
      </div>

      <!-- Actions -->
      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" :isActive="!isSyncing" />
        <GeneralButton :label="$t('kitsu.importSelected', { count: selectedCount })" :fullWidth="true" :buttonFunction="executeSync" :isActive="hasSelectedItems && !isLoading && !isSyncing" :loading="isSyncing" />
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
import MissingTypesAlert from '@/instances/desktop/components/MissingTypesAlert.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useStageStore } from '@/stores/stages';
import { useTemplateStore } from '@/stores/template';
import { useUserStore } from '@/stores/users';

const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const modals = useDesktopModalStore();
const stageStore = useStageStore();
const templateStore = useTemplateStore();
const userStore = useUserStore();
const { t } = useI18n();

// refs
const error = ref(null);
const expandedItems = ref(new Set());
const isLoading = ref(false);
const isSyncing = ref(false);
const loadingMessage = ref('');
const readinessState = ref(null);
const selectedKeys = ref(new Set());
const executionError = ref(null);
const showMissingTypes = ref(false);

// computed
// Returns the hierarchical tree for sync preview.
const syncPreviewTree = computed(() => integrationStore.syncPreviewTree);

const allTreeItems = computed(() => flattenTree(syncPreviewTree.value));

const selectedCount = computed(() => allTreeItems.value.filter(item => {
  if (!isActionable(item)) return false;
  const keys = item.selection_keys || [];
  return keys.length > 0 && keys.every(key => selectedKeys.value.has(key));
}).length);

const hasSelectedItems = computed(() => selectedCount.value > 0);

const selectedCreateCount = computed(() => countSelectedActions('create'));

const selectedLinkCount = computed(() => countSelectedActions('link'));

const noActionCount = computed(() => allTreeItems.value.filter(item => item.action === 'skip').length);

const missingTypes = computed(() => integrationStore.syncPreview?.missing_types || []);

const selectedMissingTypes = computed(() => missingTypes.value.filter(type => {
  return allTreeItems.value.some(item => {
    const selected = (item.selection_keys || []).some(key => selectedKeys.value.has(key));
    const externalTypes = item.external_types || [item.external_type];
    return selected && item.action === 'create' && externalTypes.includes(type.external_name) && item.type === type.type_category;
  });
}));

const blockedItemCount = computed(() => allTreeItems.value.filter(item => isActionable(item) && !canSelectItem(item)).length);

const permissionMessage = computed(() => {
  if (!blockedItemCount.value) return '';
  return t('kitsu.permissionBlocked', { count: blockedItemCount.value });
});

const readinessMessage = computed(() => {
  if (readinessState.value === 'credentialRequired') return t('kitsu.connectToPreview');
  if (readinessState.value === 'unlinked') return t('kitsu.linkToPreview');
  return t('kitsu.openProjectToImport');
});

const readinessActionLabel = computed(() => readinessState.value === 'credentialRequired' ? t('kitsu.connectKitsu') : t('kitsu.manageIntegration'));

// methods
// Closes the modal.
const closeModal = () => {
  if (isSyncing.value) return;
  modals.disableAllModals();
};

// Executes sync - creates all items from the preview.
const executeSync = async () => {
  if (!hasSelectedItems.value) return;

  isSyncing.value = true;
  executionError.value = null;
  stageStore.operationActive = true;
  try {
    await integrationStore.executeSync(selectedKeys.value);
    emitter.emit('refresh-browser');
    modals.disableAllModals();
  } catch (err) {
    executionError.value = err.message || t('kitsu.importFailed');
  } finally {
    stageStore.operationActive = false;
    isSyncing.value = false;
  }
};

// Returns the app icon path.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

const flattenTree = (items) => {
  const flattened = [];
  for (const item of items) {
    flattened.push(item);
    flattened.push(...flattenTree(item.children || []));
  }
  return flattened;
};

const isActionable = (item) => item.action === 'create' || item.action === 'link';

const itemNeedsMissingType = (item) => {
  const externalTypes = item.external_types || [item.external_type];
  return missingTypes.value.some(type => {
    return externalTypes.includes(type.external_name) && type.type_category === item.type;
  });
};

const canSelectItem = (item) => {
  if (!isActionable(item)) return false;
  if (item.action === 'create' && item.type === 'asset' && !userStore.canDo('create_asset')) return false;
  if (item.action === 'create' && item.type === 'collection' && !userStore.canDo('create_collection')) return false;
  if (item.action === 'create' && itemNeedsMissingType(item) && !userStore.canDo('change_role')) return false;

  let parentPath = item.parent_path;
  while (parentPath && parentPath !== '/') {
    const parent = allTreeItems.value.find(candidate => candidate.collection_path === parentPath);
    if (!parent) break;
    if (parent.action === 'create' && !userStore.canDo('create_collection')) return false;
    if (parent.action === 'create' && itemNeedsMissingType(parent) && !userStore.canDo('change_role')) return false;
    parentPath = parent.parent_path;
  }
  return true;
};

const countSelectedActions = (action) => {
  return allTreeItems.value.filter(item => {
    if (item.action !== action) return false;
    return (item.selection_keys || []).some(key => selectedKeys.value.has(key));
  }).length;
};

const selectItem = (item, selected, nextSelection) => {
  if (canSelectItem(item)) {
    for (const key of item.selection_keys || []) {
      if (selected) nextSelection.add(key);
      else nextSelection.delete(key);
    }
  }
  for (const child of item.children || []) {
    selectItem(child, selected, nextSelection);
  }
};

const selectRequiredParents = (item, nextSelection) => {
  let parentPath = item.parent_path;
  while (parentPath && parentPath !== '/') {
    const parent = allTreeItems.value.find(candidate => candidate.collection_path === parentPath);
    if (!parent) break;
    if (canSelectItem(parent)) {
      for (const key of parent.selection_keys || []) nextSelection.add(key);
    }
    parentPath = parent.parent_path;
  }
};

const toggleSelection = (item) => {
  if (!canSelectItem(item)) return;
  const keys = item.selection_keys || [];
  const selected = !keys.every(key => selectedKeys.value.has(key));
  const nextSelection = new Set(selectedKeys.value);
  selectItem(item, selected, nextSelection);
  if (selected) selectRequiredParents(item, nextSelection);
  selectedKeys.value = nextSelection;
};

const selectAll = () => {
  const nextSelection = new Set();
  for (const item of allTreeItems.value) {
    if (!canSelectItem(item)) continue;
    for (const key of item.selection_keys || []) nextSelection.add(key);
  }
  selectedKeys.value = nextSelection;
};

const unselectAll = () => {
  selectedKeys.value = new Set();
};

// Loads the sync preview without changing project data.
const loadSyncPreview = async () => {
  isLoading.value = true;
  loadingMessage.value = t('kitsu.loadingImport');
  error.value = null;
  executionError.value = null;
  readinessState.value = null;

  try {
    await integrationStore.initializeForActiveProject();
    if (!integrationStore.activeProjectUri) {
      readinessState.value = 'noProject';
      return;
    }
    if (!integrationStore.linkedIntegration) {
      readinessState.value = 'unlinked';
      return;
    }
    const integrationId = integrationStore.linkedIntegration.integration_id;
    if (!integrationStore.isAuthenticated(integrationId)) {
      readinessState.value = 'credentialRequired';
      return;
    }

    loadingMessage.value = t('kitsu.fetchingData');
    await integrationStore.getSyncPreview();

    // Load templates for extension display
    await templateStore.reloadTemplates();
    selectAll();
    expandedItems.value = new Set(
      allTreeItems.value
        .filter(item => item.children?.length)
        .map(item => item.id)
    );
  } catch (err) {
    console.error('loadSyncPreview error:', err);
    if (err.code === 'INTEGRATION_CREDENTIAL_REQUIRED' || /unauthor|forbidden|token|session|401|403/i.test(err.message || '')) {
      readinessState.value = 'credentialRequired';
      return;
    }
    error.value = err.message || t('kitsu.previewFailed');
  } finally {
    isLoading.value = false;
  }
};

const resolveReadiness = () => {
  if (readinessState.value === 'credentialRequired') {
    const integrationId = integrationStore.linkedIntegration?.integration_id || 'kitsu';
    integrationStore.requestAuthentication(integrationId, 'integrationSyncModal');
    modals.setModalVisibility('integrationAuthModal', true);
    return;
  }
  modals.setModalVisibility('integrationLinkModal', true);
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
  background-color: var(--surface-4);
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
  display: flex;
  gap: 12px;
  font-size: 15px;
}

.preview-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-left: auto;
}

.preview-actions button {
  border: 0;
  padding: 3px 7px;
  color: var(--text-muted);
  background: transparent;
  cursor: pointer;
}

.preview-actions button:hover {
  color: var(--text);
  background-color: var(--hover);
  border-radius: var(--tiny-radius);
}

.preview-divider {
  width: 100%;
  height: 1px;
  background-color: var(--surface-3);
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
.error-state,
.readiness-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px;
  gap: 16px;
  color: var(--text-secondary);
}

.permission-message,
.inline-error {
  margin: 0;
  font-size: 12px;
}

.permission-message {
  color: var(--text-secondary);
}

.inline-error {
  color: var(--error);
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
  background-color: var(--surface-4);
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
