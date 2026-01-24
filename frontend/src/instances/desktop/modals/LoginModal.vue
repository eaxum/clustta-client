<template>

  <div class="modal-container" v-esc="closeModal" v-return="handleEnterKey">
    <HeaderArea :title="title" :icon="'login'" />

    <div class="general-container">

      <div class="login-form">

        <div v-if="showStudioLogin" class="form-group studio-url-group">
          <div class="compound-form-input">
            <input autocomplete="off" class="form-input-mini" placeholder="Studio URL (e.g., https://studio.mycompany.com)" v-model="studioUrl" type="text" />
          </div>
        </div>

        <div class="form-group">
          <div class="compound-form-input">
            <input v-model="username" class="form-input-mini" type="text" placeholder="Username or Email address" autocomplete="off"/>
          </div>
        </div>
        
        <div class="form-group">
          <div class="compound-form-input">
            <input v-model="password" class="form-input-mini" :type="isPasswordVisible ? 'text' : 'password'" placeholder="Password"
              autocomplete="new-password" @keydown.enter="handleEnterKey" />
            <ActionButton 
              v-if="password"
              v-tooltip="isPasswordVisible ? 'Hide Password' : 'Show Password'"
              :icon="isPasswordVisible ? getAppIcon('eye-cancel') : getAppIcon('eye')"
              @click="togglePasswordVisibility"
              :showLabel="false"
            />
          </div>
        </div>

        

        <div class="horizontal-flex">
          <ActionButton :isInactive="true" :icon="getAppIcon('two-drives')" :label="'Private Server'" />
          <ToggleSwitch  @click="toggleStudioLogin" :switchValueProp="showStudioLogin" />
        </div>

      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Log in'" :fullWidth="true" @click="logUserIn(username, password)"
          :isActive="isValueChanged" :loading="isAwaitingResponse" />
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref } from 'vue';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// services
import { AuthService, SettingsService } from '@/services';

// stores
import { useAccountStore } from '@/stores/accounts';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useThemeStore } from '@/stores/theme';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

const accountStore = useAccountStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stageStore = useStageStore();
const themeStore = useThemeStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

// refs
const isAwaitingResponse = ref(false);
const isPasswordVisible = ref(false);
const password = ref('');
const projectDirectoryExists = ref(false);
const showStudioLogin = ref(false);
const studioUrl = ref('');
const username = ref('');

// constants
const title = 'Login';

// computed
// Returns whether the form has valid values.
const isValueChanged = computed(() => {
  const usernameValid = username.value !== '';
  const passwordValid = password.value !== '';
  if (showStudioLogin.value) {
    return usernameValid && passwordValid && studioUrl.value.trim() !== '';
  }
  return usernameValid && passwordValid;
});

// methods
// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles enter key press to submit login.
const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    logUserIn(username.value, password.value);
  }
};

// Loads projects for the user.
const loadProjects = async () => {
  await projectStore.loadProjects();
  trayStates.refreshData();
};

// Logs the user in with provided credentials.
const logUserIn = async (usernameValue, passwordValue) => {
  isAwaitingResponse.value = true;
  const isStudioLogin = showStudioLogin.value && studioUrl.value.trim();
  const normalizedStudioUrl = isStudioLogin ? normalizeStudioUrl(studioUrl.value) : '';
  const loginPromise = isStudioLogin
    ? AuthService.LoginWithHost(usernameValue, passwordValue, normalizedStudioUrl, 'studio', '')
    : AuthService.Login(usernameValue, passwordValue);

  await loginPromise
    .then(async (data) => {
      userStore.user = data.user;
      userStore.isUserAuthenticated = true;
      userStore.$reset();
      projectStore.$reset();
      trayStates.$reset();
      modals.setModalVisibility('loginModal', false);
      userStore.user = data.user;
      userStore.isUserAuthenticated = true;
      await accountStore.refreshAccounts();
      projectDirectoryExists.value = await SettingsService.GetProjectDirectory();

      if (accountStore.isAdditionalAccount) {
        await accountStore.switchToAccount(data.user.id, {
          userStore,
          projectStore,
          trayStates,
          themeStore,
          notificationStore,
          stageStore
        });
        if (!projectDirectoryExists.value) {
          setDirectories();
        }
      } else {
        await themeStore.initializeTheme();
        await projectStore.loadStudios();
        if (projectDirectoryExists.value) {
          await loadProjects();
          trayStates.refreshData();
          modals.setModalVisibility('loginModal', false);
        } else {
          setDirectories();
        }
      }
    })
    .catch((error) => {
      console.log(error);
      isAwaitingResponse.value = false;
      notificationStore.errorNotification('Error Loggin In', error);
    });
};

// Normalizes studio URL to ensure proper format.
const normalizeStudioUrl = (url) => {
  if (!url) return '';
  let normalized = url.trim();
  normalized = normalized.replace(/\/+$/, '');
  if (!normalized.startsWith('http://') && !normalized.startsWith('https://')) {
    normalized = 'https://' + normalized;
  }
  return normalized;
};

// Opens the directory onboarding modal.
const setDirectories = async () => {
  modals.setModalVisibility('dirOnboardModal', true);
};

// Toggles password visibility.
const togglePasswordVisibility = () => {
  isPasswordVisible.value = !isPasswordVisible.value;
};

// Toggles studio login mode on/off.
const toggleStudioLogin = () => {
  showStudioLogin.value = !showStudioLogin.value;
  if (!showStudioLogin.value) {
    studioUrl.value = '';
  }
};
</script>


<style scoped>
@import "@/assets/desktop.css";

.modal-container {
  justify-content: flex-start;
  align-items: flex-start;
  max-height: 90vh;
}

.general-container {
  padding-top: 1rem;
}

.login-form {
  display: flex;
  box-sizing: border-box;
  height: max-content;
  gap: 1rem;
  width: 100%;
  flex-direction: column;
  align-items: center;
  justify-content: space-around;
  overflow: hidden;
}

.form-group {
  width: 100%;
}

.compound-form-input {
  box-sizing: border-box;
  border-radius: 4px;
  font-size: 1rem;
  transition: border-color 0.2s;
  width: 100%;
  height: 50px;
  border-radius: var(--normal-radius);
  padding-right: .5rem;
  display: flex;
  overflow: hidden;
  gap: .2rem;
  background-color: var(--midnight-steel);
  align-items: center;
}

.form-input-mini {
  color: var(--white);
  box-sizing: border-box;
  border: 0px;
  border-radius: 4px;
  font-size: 1rem;
  width: 100%;
  height: 100%;
  padding: 0.75rem;
  background-color: var(--midnight-steel);
  font-family: 'Inter', sans-serif;
  font-size: 16px;
  border-radius: 12px;
  padding: 10px;
  border-style: solid;
  outline: none;
}
</style>

