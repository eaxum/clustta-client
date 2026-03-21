<template>
  <div class="sso-section">
    <div class="sso-divider">
      <span class="sso-divider-line"></span>
      <span class="sso-divider-text">or</span>
      <span class="sso-divider-line"></span>
    </div>

    <button class="sso-button google" @click="handleGoogleSSO" :disabled="isLoading">
      <img class="sso-icon" src="/brand-logos/google.svg" alt="Google" />
      <span v-if="!isLoading">{{ buttonLabel }}</span>
      <ActionButton v-else :icon="getAppIcon('loading')" :isLoading="true" :showLabel="false" :noFilter="true" />
    </button>
  </div>
</template>

<script setup>

// imports
import { computed, ref } from 'vue';

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
const isLoading = ref(false);

// computed properties
const buttonLabel = computed(() => {
  return props.mode === 'signup' ? 'Sign up with Google' : 'Login with Google';
});

// methods

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Initiates Google SSO by opening the system browser.
const handleGoogleSSO = async () => {
  if (isLoading.value) return;
  isLoading.value = true;

  try {
    const data = await AuthService.LoginWithSSO('');
    emit('success', data);
  } catch (err) {
    console.log(err);
    emit('error', err.message || 'Google sign-in failed. Please try again.');
  } finally {
    isLoading.value = false;
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

.sso-divider {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 0.75rem;
}

.sso-divider-line {
  flex: 1;
  height: 1px;
  background-color: var(--light-steel);
}

.sso-divider-text {
  font-size: 0.8rem;
  color: var(--light-steel);
  text-transform: lowercase;
}

.sso-button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.7rem 1rem;
  border: none;
  border-radius: var(--large-radius);
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s, border-radius 0.2s;
  background-color: var(--silver);
  color: var(--black-steel);
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
  width: 18px;
  height: 18px;
}
</style>
