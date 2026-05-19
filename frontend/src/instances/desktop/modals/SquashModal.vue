<template>
  <div class="modal-container large-modal" v-stop-propagation v-esc="closeModal">
    <HeaderArea :title="$t('components.squashModal.title')" :icon="'squash'" />

    <div class="general-container">
      <ProgressSection v-if="isSquashing" :title="$t('components.squashModal.squashingAssets')" />

      <div v-else class="step-content">
        <FormInput :label="$t('components.squashModal.assetName')" :modelValue="assetName" @update:modelValue="updateAssetName"
          :placeholder="$t('components.squashModal.assetNamePlaceholder')" :error="nameError" :needsValidation="true"
          :showValidation="assetName.length > 0" :valid="!nameError && assetName.length > 0" :autofocus="true" />

        <div class="preview-divider"></div>

        <div class="preview-header">
          <span class="preview-summary">{{ checkpointItems.length }} {{ $t('components.squashModal.filesToCheckpoints') }}</span>
          <span v-if="commonExtension" class="preview-extension">{{ commonExtension }}</span>
        </div>

        <div class="squash-preview-scroll">
          <div v-if="checkpointItems.length === 0" class="empty-preview">
            <span class="empty-text">{{ $t('components.squashModal.noItems') }}</span>
          </div>
          <div v-else class="preview-list-content">
            <SquashPreviewItem v-for="entry in checkpointItems" :key="entry.item.id" :item="entry.item"
              :label="entry.label" :index="entry.index" />
          </div>
        </div>

        <div class="squash-option-row" @click="deleteSourceFiles = !deleteSourceFiles">
          <ToggleSwitch :switchValueProp="deleteSourceFiles" />
          <span class="option-label">{{ $t('components.squashModal.deleteSourceFiles') }}</span>
        </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('components.squashModal.squash')" :fullWidth="true" :buttonFunction="executeSquash"
          :isActive="canConfirm && !isSquashing" :loading="isSquashing" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import { analyzeFileNames, generateCheckpointLabels } from '@/utils/squash';

// components
import FormInput from '@/instances/desktop/components/FormInput.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ProgressSection from '@/instances/common/components/ProgressSection.vue';
import SquashPreviewItem from '@/instances/common/components/SquashPreviewItem.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// services
import { AssetService, CheckpointService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';

const { t } = useI18n();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stageStore = useStageStore();

// refs
const assetName = ref('');
const deleteSourceFiles = ref(false);
const excludedNames = ref(new Set());
const isSquashing = ref(false);
const orderedItems = ref([]);
const patternFound = ref(false);

// computed
// Returns the checkpoint preview items with labels.
const checkpointItems = computed(() => generateCheckpointLabels(orderedItems.value, patternFound.value));

// Returns whether the confirm button should be active.
const canConfirm = computed(() => {
  return assetName.value.trim().length > 0 && !nameError.value && orderedItems.value.length >= 2;
});

// Returns the common extension of all selected files.
const commonExtension = computed(() => {
  const items = stageStore.selectedItems.filter(i => i.type === 'untracked_asset');
  if (items.length === 0) return '';
  return items[0].extension || '';
});

// Returns the collection ID shared by all selected items.
const collectionId = computed(() => {
  const items = stageStore.selectedItems.filter(i => i.type === 'untracked_asset');
  if (items.length === 0) return '';
  return items[0].collection_id || '';
});

// Returns validation error for the asset name.
const nameError = computed(() => {
  const name = assetName.value.trim();
  if (name.length === 0) return '';
  if (name.length > 255) return t('components.squashModal.nameTooLong');
  if (excludedNames.value.has(name.toLowerCase())) return t('components.squashModal.nameAlreadyExists');
  return '';
});

// methods
// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Executes the squash operation.
const executeSquash = async () => {
  if (!canConfirm.value) return;

  isSquashing.value = true;
  stageStore.operationActive = true;

  try {
    const filePaths = orderedItems.value.map(item => item.file_path);
    const comments = checkpointItems.value.map(entry => entry.label);
    await CheckpointService.SquashAssets(
      projectStore.activeProject.uri,
      projectStore.activeProject.working_directory,
      filePaths,
      assetName.value.trim(),
      collectionId.value,
      deleteSourceFiles.value,
      comments,
    );
    notificationStore.addNotification(t('components.squashModal.squashSuccess'), '', 'success', false);
    emitter.emit('refresh-browser');
    closeModal();
  } catch (err) {
    console.error('Squash error:', err);
    notificationStore.errorNotification(t('components.squashModal.squashFailed'), err);
  } finally {
    stageStore.operationActive = false;
    isSquashing.value = false;
  }
};

// Loads sibling asset names for validation.
const loadExcludedNames = async () => {
  try {
    const ext = commonExtension.value;
    const eid = collectionId.value;
    const names = await AssetService.GetSiblingAssetNames(projectStore.activeProject.uri, eid, ext);
    excludedNames.value = new Set((names || []).map(n => n.toLowerCase()));
  } catch (err) {
    console.error('Failed to load sibling names:', err);
    excludedNames.value = new Set();
  }
};

// Updates the asset name from input.
const updateAssetName = (value) => {
  assetName.value = value;
};

// lifecycle
onMounted(async () => {
  const selected = stageStore.selectedItems.filter(i => i.type === 'untracked_asset');
  const analysis = analyzeFileNames(selected);
  orderedItems.value = analysis.orderedItems;
  patternFound.value = analysis.patternFound;
  assetName.value = analysis.commonName;
  await loadExcludedNames();
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
  width: 45vw;
  min-width: 500px;
  max-width: 750px;
  box-sizing: border-box;
}

.step-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
  box-sizing: border-box;
  overflow: hidden;
  width: 100%;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.preview-summary {
  font-size: 14px;
  color: var(--text-muted);
}

.preview-extension {
  font-size: 12px;
  color: var(--text-muted);
  opacity: 0.7;
  background-color: var(--surface-3);
  padding: 2px 8px;
  border-radius: var(--small-radius);
}

.preview-divider {
  width: 100%;
  height: 1px;
  background-color: var(--surface-3);
  flex-shrink: 0;
}

.squash-preview-scroll {
  display: flex;
  flex-direction: column;
  max-height: 320px;
  overflow-y: auto;
  border-radius: var(--small-radius);
  background: var(--surface-secondary);
}

.squash-preview-scroll::-webkit-scrollbar {
  width: 4px;
}

.squash-preview-scroll::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-4);
}

.squash-preview-scroll::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.preview-list-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 4px;
}

.empty-preview {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
}

.empty-text {
  font-size: 14px;
  color: var(--text-muted);
  opacity: 0.7;
}

.squash-option-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.25rem 0;
}

.option-label {
  font-size: 13px;
  color: var(--text-muted);
  cursor: pointer;
  user-select: none;
}

.option-label:hover {
  color: var(--text);
}
</style>
