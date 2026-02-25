<template>
  <div class="modal-container">
    <HeaderArea :title="title" :icon="'floppy-disk'" :showSearch="false" />
    <div class="general-container">
      <div class="input-section">
        <FormInput v-model="presetName" :placeholder="$t('placeholders.presetName')" :info="infoMessage"
          :autofocus="true" @keydown.enter="handleEnterKey" />
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.create')" :fullWidth="true" @click="savePreset" :isActive="isValidName"
          :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';

// components
import FormInput from '@/instances/desktop/components/FormInput.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { SettingsService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

const { t } = useI18n();
const assetStore = useAssetStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

// refs
const isAwaitingResponse = ref(false);
const presetName = ref('');

// constants
const title = t('modals.saveDependencyPreset');

// computed
// Returns the dependencies from the store.
const dependencies = computed(() => assetStore.dependencyPresetModalData.dependencies);

// Returns the existing presets from the store.
const existingPresets = computed(() => assetStore.dependencyPresetModalData.existingPresets);

// Returns the info message for the preset name input when overwriting.
const infoMessage = computed(() => {
  if (presetName.value.trim() === '') return '';
  const existingNames = existingPresets.value.map((preset) => preset.name.toLowerCase());
  if (existingNames.includes(presetName.value.toLowerCase())) {
    return t('info.presetWillOverwrite');
  }
  return '';
});

// Returns whether the preset name is valid.
const isValidName = computed(() => {
  return presetName.value.trim() !== '';
});

// methods
// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Handles enter key press to submit form.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValidName.value) {
    savePreset();
  }
};

// Saves the dependency preset.
const savePreset = async () => {
  isAwaitingResponse.value = true;
  
  const dependencyData = dependencies.value.map((dep) => ({
    id: dep.id,
    type: dep.type
  }));

  const newPreset = {
    name: presetName.value,
    dependencies: dependencyData,
    createdAt: Date.now()
  };

  try {
    await SettingsService.AddDependencyPreset(projectStore.getActiveProject.id, newPreset);
    notificationStore.addNotification(t('notifications.dependencyPresetSaved'), "", "success");
    emitter.emit('dependency-preset-added', newPreset);
    closeModal();
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorSavingDependencyPreset'), error);
  } finally {
    isAwaitingResponse.value = false;
  }
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.general-container {
  gap: .5rem;
}
</style>
