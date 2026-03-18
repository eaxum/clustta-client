<template>
  <div class="modal-container" v-stop-propagation>
    <HeaderArea :title="'Configure AI Agent'" :icon="getAppIcon('brain')" :showSearch="false" />
    <div class="general-container">

      <div class="input-section">
        <label class="input-label">LLM Provider</label>
        <DropDownBox :items="providerOptions" :onSelect="onProviderSelect" :selectedItem="selectedProvider" :placeHolder="'Select Provider'" :fullWidth="true" />
      </div>

      <div v-if="selectedProvider !== 'Ollama'" class="input-section">
        <label class="input-label">API Key</label>
        <input v-model="apiKey" type="password" class="input-short" :placeholder="agentKeyConfigured ? 'Enter new key to replace...' : 'Paste your API key...'" @keydown.enter="saveConfig" />
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Save'" :fullWidth="true" :buttonFunction="saveConfig" :isActive="canSave" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';

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
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

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
    notificationStore.addNotification('AI Agent', 'LLM provider saved successfully.', 'success');
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
