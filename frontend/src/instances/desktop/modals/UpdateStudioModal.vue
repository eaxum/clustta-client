<template>
  <div ref="modalContainer" class="modal-container">

    <HeaderArea :title="title" :icon="getAppIcon('stall')" :showSearch="false" />

    <div class="general-container">

      <div class="input-section">
        <div class="input-label-row">
          <label class="input-label">{{ $t('modals.studioNameLabel') }}</label>
        </div>
        <div class="horizontal-flex">
          <input v-model="studioName" class="input-short" type="text" :placeholder="$t('placeholders.studioName')" disabled />
        </div>
      </div>

      <div class="input-section">
        <div class="input-label-row">
          <label class="input-label">{{ $t('modals.studioUrlLabel') }}</label>
        </div>
        <div class="horizontal-flex">
          <input v-model="studioUrl" class="input-short" type="text" :placeholder="$t('placeholders.studioUrl')" v-focus />
        </div>
      </div>

      <div class="input-section">
        <div class="input-label-row">
          <label class="input-label">{{ $t('modals.alternateUrlLabel') }}</label>
        </div>
        <div class="horizontal-flex">
          <input v-model="studioAltUrl" class="input-short" type="text" :placeholder="$t('placeholders.alternateUrlOptional')" />
        </div>
      </div>

      <!-- <div class="input-section">
        <div class="input-label-row">
          <label class="input-label">Port</label>
        </div>
        <div class="horizontal-flex">
          <input v-model="studioPort" class="input-short" type="text" placeholder="Port" />
        </div>
      </div> -->

      <div class="input-section">
        <div class="input-label-row">
          <label class="input-label">{{ $t('modals.studioKeyLabel') }}</label>
        </div>
        <div class="horizontal-flex">
          <input v-model="studioKey" class="input-short monospace-input" type="password" :placeholder="$t('placeholders.studioKey')" />
        </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.close')" :fullWidth="true" :buttonFunction="closeModal" :isActive="!isAwaitingResponse" :colored="false" />
        <GeneralButton :label="$t('common.update')" :fullWidth="true" @click="updateStudio" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>
    </div>
    
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { StudioService } from '@/services';

// stores
const { t } = useI18n();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

// constants
const title = t('modals.updateStudio');

// refs
const isAwaitingResponse = ref(false);
const modalContainer = ref(null);
const originalStudio = ref(null);
const studioAltUrl = ref('');
const studioKey = ref('');
const studioName = ref('');
const studioPort = ref('');
const studioUrl = ref('');

// computed (dependencies first)
const hasChanges = computed(() => {
  if (!originalStudio.value) return false;
  return (
    studioUrl.value !== originalStudio.value.url ||
    studioAltUrl.value !== (originalStudio.value.alt_url || '') ||
    studioPort.value !== (originalStudio.value.port || '')
  );
});

const keyProvided = computed(() => {
  return studioKey.value.trim() !== '';
});

const isValueChanged = computed(() => {
  return keyProvided.value && hasChanges.value;
});

// methods
// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Returns icon path from icon store.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles enter key press to trigger update.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    updateStudio();
  }
};

// Loads current studio data into form fields.
const loadStudioData = () => {
  const currentStudio = projectStore.selectedStudio;
  if (currentStudio) {
    originalStudio.value = { ...currentStudio };
    studioName.value = currentStudio.name || '';
    studioUrl.value = currentStudio.url || '';
    studioAltUrl.value = currentStudio.alt_url || '';
    studioPort.value = currentStudio.port || '';
    studioKey.value = '';
  }
};

// Submits studio update to server.
const updateStudio = async () => {
  if (!isValueChanged.value) return;
  
  isAwaitingResponse.value = true;

  try {
    await StudioService.UpdateStudio(
      studioName.value,
      studioUrl.value,
      studioAltUrl.value,
      studioPort.value,
      studioKey.value
    );

    notificationStore.addNotification(t('notifications.studioUpdated'), "", "success");
    
    await projectStore.loadStudios();
    let studio = projectStore.studios.find((item) => item.name === studioName.value);
    if (studio) {
      projectStore.selectedStudio = studio;
    }
    await projectStore.loadProjects();
    
    isAwaitingResponse.value = false;
    closeModal();
  } catch (error) {
    isAwaitingResponse.value = false;
    console.log(error);
    notificationStore.errorNotification(t('notifications.errorUpdatingStudio'), error);
  }
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// lifecycle
onMounted(async () => {
  loadStudioData();
});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/modals.css";

.general-container {
  gap: 1rem;
}

.input-section {
  width: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  gap: 0.5rem;
  color: var(--text);
}

.input-label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.input-label {
  font-family: Inter, sans-serif;
  color: var(--text);
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  opacity: 0.9;
}

.input-short {
  flex: 1;
  width: 100%;
}

.input-short:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.monospace-input {
  font-family: 'Courier New', monospace;
  font-size: 14px;
}

[data-theme="dark"] .input-short {
  font-weight: 200;
}
</style>
