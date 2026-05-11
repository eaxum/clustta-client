<template>
  <div class="modal-container">
    <HeaderArea :title="title" :icon="'floppy-disk'" :showSearch="false" />
    <div class="general-container">
      <div class="input-section">
        <FormInput v-model="presetName" :placeholder="$t('placeholders.presetName')" :info="infoMessage" :autofocus="true" @keydown.enter="handleEnterKey" />
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.save')" :fullWidth="true" @click="savePreset" :isActive="isValidName" :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import { ignoreTemplateNames } from '@/lib/ignoreTemplates';

// components
import FormInput from '@/instances/desktop/components/FormInput.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { SettingsService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

const { t } = useI18n();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

// refs
const existingNames = ref([]);
const isAwaitingResponse = ref(false);
const presetName = ref('');

// constants
const builtInNamesLower = ignoreTemplateNames.map((n) => n.toLowerCase());
const title = t('modals.saveIgnoreListPreset');

// computed
// Returns the current ignore list entries to be saved.
const entries = computed(() => projectStore.activeProject?.ignore_list || []);

// Returns an info message for overwrite or built-in name collisions.
const infoMessage = computed(() => {
  const trimmed = presetName.value.trim();
  if (!trimmed) return '';
  const lower = trimmed.toLowerCase();
  if (builtInNamesLower.includes(lower)) {
    return t('info.ignorePresetBuiltInCollision');
  }
  if (existingNames.value.map((n) => n.toLowerCase()).includes(lower)) {
    return t('info.presetWillOverwrite');
  }
  return '';
});

// Returns whether the preset name is valid and savable.
const isValidName = computed(() => {
  const trimmed = presetName.value.trim();
  if (!trimmed) return false;
  if (builtInNamesLower.includes(trimmed.toLowerCase())) return false;
  return entries.value.length > 0;
});

// methods
// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Submits the form when Enter is pressed.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValidName.value) {
    savePreset();
  }
};

// Persists the current ignore list as a named preset.
const savePreset = async () => {
  if (!isValidName.value) return;
  isAwaitingResponse.value = true;
  const name = presetName.value.trim();
  try {
    await SettingsService.AddIgnoreListPreset(name, [...entries.value]);
    notificationStore.addNotification(t('notifications.ignorePresetSaved'), '', 'success');
    emitter.emit('ignore-preset-added', { name });
    closeModal();
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorSavingIgnorePreset'), error);
  } finally {
    isAwaitingResponse.value = false;
  }
};

// lifecycle hooks
onMounted(async () => {
  try {
    const presets = await SettingsService.GetIgnoreListPresets();
    existingNames.value = Object.keys(presets || {});
  } catch (error) {
    existingNames.value = [];
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.general-container {
  gap: .5rem;
}
</style>
