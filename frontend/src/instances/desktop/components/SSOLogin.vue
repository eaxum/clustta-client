<template>
  <div class="sso-section">
    <div class="sso-divider">
      <span class="sso-divider-line"></span>
      <span class="sso-divider-text">or</span>
      <span class="sso-divider-line"></span>
    </div>

    <div class="sso-buttons">
      <button class="sso-button" @click="handleSSO('google')" :disabled="isLoading">
        <img v-if="!isLoading || activeProvider !== 'google'" class="sso-icon" src="/brand-logos/google.svg" alt="Google" />
        <ActionButton v-else :icon="getAppIcon('loading')" :isLoading="true" :showLabel="false" :noFilter="true" />
        <span>Google</span>
      </button>

      <button class="sso-button" @click="handleSSO('microsoft')" :disabled="isLoading">
        <img v-if="!isLoading || activeProvider !== 'microsoft'" class="sso-icon" src="/brand-logos/microsoft.svg" alt="Microsoft" />
        <ActionButton v-else :icon="getAppIcon('loading')" :isLoading="true" :showLabel="false" :noFilter="true" />
        <span>Microsoft</span>
      </button>

      <button class="sso-button" @click="handleSSO('apple')" :disabled="isLoading">
        <img v-if="!isLoading || activeProvider !== 'apple'" class="sso-icon apple-icon" src="/brand-logos/apple.svg" alt="Apple" />
        <ActionButton v-else :icon="getAppIcon('loading')" :isLoading="true" :showLabel="false" :noFilter="true" />
        <span>Apple</span>
      </button>
    </div>
  </div>
</template>

<script setup>

// imports
import { ref } from 'vue';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// services
import { AuthService } from '@/services';

// stores
const iconStore = useIconStore();

// store imports
import { useIconStore } from '@/stores/icons';

// props
const props = defineProps({
  mode: { type: String, default: 'login' },
});

// emits
const emit = defineEmits(['success', 'error']);

// refs
const activeProvider = ref('');
const isLoading = ref(false);

// computed properties

// methods

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Initiates SSO by opening the system browser for the specified provider.
const handleSSO = async (provider) => {
  if (isLoading.value) return;
  isLoading.value = true;
  activeProvider.value = provider;

  try {
    const data = await AuthService.LoginWithSSO('', provider);
    emit('success', data);
  } catch (err) {
    console.log(err);
    emit('error', err.message || `${provider} sign-in failed. Please try again.`);
  } finally {
    isLoading.value = false;
    activeProvider.value = '';
  }
};
</script>

<style scoped>
.sso-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  gap: 0.75rem;
}

.sso-buttons {
  display: flex;
  flex-direction: row;
  width: 100%;
  gap: 0.75rem;
}

.sso-divider {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 0.75rem;
}

.sso-divider-line {
  flex: 1;
  height: 1px;
  background-color: var(--surface-4);
}

.sso-divider-text {
  font-size: 0.8rem;
  color: var(--surface-4);
  text-transform: lowercase;
}

.sso-button {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.1rem;
  border: none;
  border-radius: var(--large-radius);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s, border-radius 0.2s;
  background-color: var(--surface-inverse);
  color: var(--text-inverse);
  height: 40px;
  min-height: 40px;
  max-height: 40px;
}

.sso-button:hover {
  border-radius: var(--normal-radius);
}

.sso-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.sso-icon {
  flex-shrink: 0;
  width: 20px;
  height: 20px;
}

.apple-icon {
  filter: invert(1);
}

[data-theme="dark"] .apple-icon {
  filter: none;
}

@media (max-width: 768px) {
  .sso-button span {
    display: none;
  }

  .sso-button {
    gap: 0;
  }
}
</style>
