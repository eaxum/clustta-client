<template>
  <div class="modal-container" v-stop-propagation v-esc="closeModal" v-return="handleEnterKey">
    <HeaderArea :title="title" :icon="headerIcon" />

    <div class="general-container">
      <!-- Integration Selection -->
      <div v-if="!selectedIntegration" class="integration-list">
        <div v-for="integration in availableIntegrations" :key="integration.id" class="integration-item"
          :class="{ 'authenticated': isAuthenticated(integration.id) }" v-stop-propagation @click="selectIntegration(integration)">
          <img :src="getAppIcon(integration.icon)" :alt="integration.name" class="integration-icon" />
          <div class="integration-info">
            <span class="integration-name">{{ integration.name }}</span>
            <span class="integration-desc">{{ integration.description }}</span>
          </div>
          <span v-if="isAuthenticated(integration.id)" class="auth-status">Connected</span>
        </div>
      </div>

      <!-- Authentication Form -->
      <div v-else class="auth-form">
        <!-- Kitsu Login Form -->
        <div v-if="selectedIntegration.auth_type === 'password'" class="form-fields">
          <FormInput v-model="apiUrl" placeholder="Server URL (e.g., https://kitsu.mystudio.com)" needsValidation
            :valid="isApiUrlValid" :showValidation="!!apiUrl" :error="apiUrlError" />
          <FormInput v-model="email" placeholder="Email address" needsValidation :valid="isEmailValid"
            :showValidation="!!email" />
          <FormInput v-model="password" placeholder="Password" isSecret />
        </div>

        <!-- OAuth Info (ClickUp) -->
        <div v-else-if="selectedIntegration.auth_type === 'oauth'" class="oauth-info">
          <p>Click the button below to authorize with {{ selectedIntegration.name }}.</p>
          <p class="oauth-note">You'll be redirected to {{ selectedIntegration.name }} to grant access.</p>
        </div>

        <div class="pop-up-actions">
          <GeneralButton :label="$t('common.back')" :fullWidth="true" :buttonFunction="clearSelection" :colored="false" />
          <GeneralButton v-if="selectedIntegration.auth_type === 'password'" :label="isAuthenticated(selectedIntegration.id) ? 'Reconnect' : 'Connect'"
            :fullWidth="true" @click="authenticate" :isActive="canAuthenticate" :loading="isAuthenticating" />
          <GeneralButton v-else :label="'Authorize'" :fullWidth="true" @click="startOAuth" :isActive="true" />
        </div>

        <!-- Disconnect Option -->
        <div v-if="isAuthenticated(selectedIntegration.id)" class="disconnect-section">
          <ActionButton :icon="getAppIcon('disconnect')" :label="'Disconnect'" :buttonFunction="disconnect" />
        </div>
      </div>

      <!-- Close Button (when showing list) -->
      <div v-if="!selectedIntegration" class="pop-up-actions">
        <GeneralButton :label="$t('common.close')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
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
import FormInput from '@/instances/desktop/components/FormInput.vue';
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
const apiUrl = ref('');
const apiUrlError = ref('');
const email = ref('');
const isAuthenticating = ref(false);
const password = ref('');
const selectedIntegration = ref(null);

// computed
const title = computed(() => {
  if (selectedIntegration.value) {
    return selectedIntegration.value.name;
  }
  return 'Connect Integration';
});

const availableIntegrations = computed(() => integrationStore.availableIntegrations);

const headerIcon = computed(() => {
  if (selectedIntegration.value) {
    return selectedIntegration.value.icon;
  }
  return 'plug';
});

// Validates the API URL format.
const isApiUrlValid = computed(() => {
  if (!apiUrl.value) return false;
  const urlRegex = /^https?:\/\/.+/;
  return urlRegex.test(apiUrl.value.trim());
});

// Validates the email format.
const isEmailValid = computed(() => {
  if (!email.value) return false;
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email.value);
});

// Checks if form can be submitted.
const canAuthenticate = computed(() => {
  return isApiUrlValid.value && isEmailValid.value && password.value.length > 0;
});

// methods
// Authenticates with the selected integration.
const authenticate = async () => {
  if (!canAuthenticate.value) return;

  isAuthenticating.value = true;
  try {
    const result = await integrationStore.authenticate(selectedIntegration.value.id, {
      api_url: apiUrl.value.trim(),
      email: email.value,
      password: password.value,
    });

    if (result.success) {
      notificationStore.addNotification('Connected to ' + selectedIntegration.value.name, '', 'success');
      clearSelection();
    }
  } finally {
    isAuthenticating.value = false;
  }
};

// Clears the selected integration.
const clearSelection = () => {
  selectedIntegration.value = null;
  apiUrl.value = '';
  email.value = '';
  password.value = '';
  apiUrlError.value = '';
};

// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Disconnects from the selected integration.
const disconnect = () => {
  integrationStore.disconnect(selectedIntegration.value.id);
  notificationStore.addNotification('Disconnected from ' + selectedIntegration.value.name, '', 'success');
};

// Returns the app icon path.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles enter key press.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && selectedIntegration.value && canAuthenticate.value) {
    authenticate();
  }
};

// Checks if authenticated with integration.
const isAuthenticated = (integrationId) => {
  return integrationStore.isAuthenticated(integrationId);
};

// Selects an integration for authentication.
const selectIntegration = (integration) => {
  selectedIntegration.value = integration;
  // Pre-fill API URL if already authenticated
  const existingUrl = integrationStore.getApiUrl(integration.id);
  if (existingUrl) {
    apiUrl.value = existingUrl;
  }
};

// Starts OAuth flow.
const startOAuth = () => {
  // TODO: Implement OAuth redirect for ClickUp
  notificationStore.addNotification('OAuth integration coming soon','', 'warning');
};

// lifecycle
onMounted(async () => {
  await integrationStore.loadAvailableIntegrations();
});
</script>

<style scoped>
.integration-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.integration-item {
  display: flex;
  align-items: center;
  gap: .5rem;
  padding: 1rem 1rem;
  border-radius: var(--large-radius);
  background-color: var(--dark-steel);
  outline: var(--transparent-line);
  outline-offset: -1px;
  cursor: pointer;
  transition: all 0.2s ease-in-out;
}

.integration-item:hover {
  background-color: var(--steel);
  border-radius: var(--small-radius);
}

.integration-item.authenticated {
  outline: 1px solid var(--accent-primary);
}

.integration-icon {
  width: 32px;
  height: 32px;
  object-fit: contain;
}

.integration-info {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.integration-name {
  font-weight: 500;
  color: var(--white);
}

.integration-desc {
  font-size: 12px;
  color: var(--white);
}

.auth-status {
  font-size: 11px;
  color: var(--accent-primary);
  padding: 2px 8px;
  background: var(--accent-primary-subtle);
  border-radius: var(--small-radius);
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
}

.form-fields {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.oauth-info {
  text-align: center;
  padding: 20px;
  color: var(--white);
}

.oauth-note {
  font-size: 12px;
  margin-top: 8px;
}

.disconnect-section {
  display: flex;
  justify-content: center;
  padding-top: 12px;
  border-top: 1px solid var(--border-primary);
}
</style>
