<template>
  <div class="modal-container" v-stop-propagation>
    <HeaderArea :title="t('modals.configureAiAgent')" :icon="CiBrain" :showSearch="false" />
    <div class="general-container">

      <div class="input-section">
        <label class="input-label">{{ $t('settings.llmProvider') }}</label>
        <DropDownBox :items="providerOptions" :onSelect="onProviderSelect" :selectedItem="selectedProvider" :placeHolder="t('modals.selectProvider')" :fullWidth="true" />
      </div>

      <div v-if="selectedProvider !== 'Ollama'" class="input-section">
        <label class="input-label">{{ $t('modals.apiKey') }}</label>
        <input v-model="apiKey" type="password" class="input-short" :placeholder="agentKeyConfigured ? t('modals.enterNewKey') : t('modals.pasteApiKey')" @keydown.enter="saveConfig" />
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.save')" :fullWidth="true" :buttonFunction="saveConfig" :isActive="canSave" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { CiBrain } from '@clustta/icons-vue';
import { resolveIcon } from '@/lib/icon-map';

// components
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { AgentService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';

const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const { t } = useI18n();

// refs
const agentKeyConfigured = ref(false);
const apiKey = ref('');
const providerMap = { 'OpenAI': 'openai', 'Anthropic': 'anthropic', 'Google Gemini': 'gemini', 'Groq': 'groq', 'Ollama': 'ollama' };
const providerMapReverse = Object.fromEntries(Object.entries(providerMap).map(([k, v]) => [v, k]));
const providerOptions = ref(['OpenAI', 'Anthropic', 'Google Gemini', 'Groq', 'Ollama']);
const selectedProvider = ref('OpenAI');

// computed
// Returns whether the save action should be enabled.
const canSave = computed(() => {
  if (selectedProvider.value === 'Ollama') return true;
  return apiKey.value.trim().length > 0 || agentKeyConfigured.value;
});

// methods
// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility('configAgentModal', false);
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.resolveIcon(iconName);

// Handles provider dropdown selection.
const onProviderSelect = (displayName) => {
  selectedProvider.value = displayName;
};

// Saves the agent LLM provider and API key.
const saveConfig = async () => {
  const providerKey = providerMap[selectedProvider.value];
  if (!providerKey) return;
  try {
    await AgentService.SetAPIKey(providerKey, apiKey.value.trim());
    agentKeyConfigured.value = true;
    apiKey.value = '';
    notificationStore.addNotification(t('settings.aiAgent'), t('modals.llmProviderSaved'), 'success');
    closeModal();
  } catch (err) {
    notificationStore.addNotification('Error', `Failed to save: ${err}`, 'error');
  }
};

// lifecycle hooks
onMounted(async () => {
  try {
    const status = await AgentService.GetAPIKeyStatus();
    agentKeyConfigured.value = status.configured;
    if (status.provider) selectedProvider.value = providerMapReverse[status.provider] || 'OpenAI';
  } catch {
    // ignore
  }
});
</script>
