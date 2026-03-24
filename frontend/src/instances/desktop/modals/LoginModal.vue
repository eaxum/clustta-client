<template>
  <div class="modal-container" v-stop-propagation v-esc="closeModal" v-return="handleEnterKey">
    <HeaderArea :title="title" :icon="'login'" />
    <div class="general-container">

      <!-- Cloud login -->
      <template v-if="loginMode === 'cloud'">
        <div class="login-form">
          <FormInput v-model="username" :placeholder="$t('placeholders.emailAddress')" needsValidation :error="emailError" :valid="isEmailValid" :showValidation="!!username" @input="validateEmail" />

          <FormInput v-model="password" :placeholder="$t('placeholders.password')" isSecret />
        </div>

        <div class="pop-up-actions">
          <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
          <GeneralButton :label="$t('common.logIn')" :fullWidth="true" @click="logUserIn(username, password)" :isActive="isLoginFormValid" :loading="isAwaitingResponse" />
        </div>

        <SSOLogin v-if="!isAwaitingResponse" mode="login" @success="handleSSOSuccess" @error="handleSSOError" />

        <div v-if="error" class="error-message">{{ error }}</div>

        <div v-if="!isAwaitingResponse" class="additional-actions">
          <div @click="setLoginMode('studio-connect')" class="studio-reveal-link">{{ $t('auth.login.connectToStudio') }}</div>
        </div>
      </template>

      <!-- Studio URL connect -->
      <template v-if="loginMode === 'studio-connect'">
        <div class="login-form">
          <FormInput v-model="studioUrl" :placeholder="$t('placeholders.studioUrl')" needsValidation :error="studioUrlError" :valid="isStudioUrlValid" :showValidation="!!studioUrl" @input="validateStudioUrl" />
        </div>

        <div class="pop-up-actions">
          <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
          <GeneralButton :label="$t('auth.login.connectButton')" :fullWidth="true" @click="connectToServer" :isActive="isStudioUrlValid" :loading="isConnecting" />
        </div>

        <div v-if="connectionError" class="error-message">{{ connectionError }}</div>

        <div class="additional-actions">
          <div @click="setLoginMode('cloud')" class="back-link">{{ $t('auth.login.loginToCloud') }}</div>
        </div>
      </template>

      <!-- Studio login (connected) -->
      <template v-if="loginMode === 'studio-login'">
        <div class="login-form">
          <div class="server-badge">
            <div class="server-badge-info">
              <div class="status-dot"></div>
              <div class="server-badge-details">
                <span class="server-badge-name">{{ connectedServerName }}</span>
                <span class="server-badge-url">{{ studioUrl }}</span>
              </div>
            </div>
            <div @click="disconnectServer" class="server-change-link">{{ $t('auth.login.changeServer') }}</div>
          </div>

          <FormInput v-model="username" :placeholder="$t('placeholders.emailAddress')" needsValidation :error="emailError" :valid="isEmailValid" :showValidation="!!username" @input="validateEmail" />

          <FormInput v-model="password" :placeholder="$t('placeholders.password')" isSecret />
        </div>

        <div class="pop-up-actions">
          <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
          <GeneralButton :label="$t('common.logIn')" :fullWidth="true" @click="logUserIn(username, password)" :isActive="isLoginFormValid" :loading="isAwaitingResponse" />
        </div>

        <div v-if="error" class="error-message">{{ error }}</div>

        <div v-if="!isAwaitingResponse" class="additional-actions">
          <div @click="backToCloudLogin" class="back-link">{{ $t('auth.login.loginToCloud') }}</div>
        </div>
      </template>

    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import FormInput from '@/instances/desktop/components/FormInput.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import SSOLogin from '@/instances/desktop/components/SSOLogin.vue';

// services
import { AuthService, SettingsService, StudioService } from '@/services';

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
const { t } = useI18n();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stageStore = useStageStore();
const themeStore = useThemeStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

// refs
const connectedServerName = ref('');
const connectionError = ref('');
const emailError = ref('');
const error = ref('');
const isAwaitingResponse = ref(false);
const isConnecting = ref(false);
const loginMode = ref('cloud');
const password = ref('');
const projectDirectoryExists = ref(false);
const studioUrl = ref('');
const studioUrlError = ref('');
const username = ref('');

// constants
const title = t('modals.loginTitle');

// computed
// Returns whether the email is valid.
const isEmailValid = computed(() => {
  if (!username.value) return false;
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(username.value);
});

// Returns whether the login form has valid values.
const isLoginFormValid = computed(() => {
  return isEmailValid.value && password.value !== '';
});

// Returns whether the studio URL is valid.
const isStudioUrlValid = computed(() => {
  if (!studioUrl.value) return false;
  const urlRegex = /^https?:\/\/[a-zA-Z0-9][-a-zA-Z0-9]*(\.[a-zA-Z0-9][-a-zA-Z0-9]*)+(:\d+)?(\/.*)?$/;
  return urlRegex.test(studioUrl.value.trim());
});

// methods

// Navigates back to cloud login, clearing studio state.
const backToCloudLogin = () => {
  loginMode.value = 'cloud';
  studioUrl.value = '';
  studioUrlError.value = '';
  connectedServerName.value = '';
  connectionError.value = '';
  error.value = '';
};

// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Connects to the studio server and retrieves its info.
const connectToServer = async () => {
  if (!isStudioUrlValid.value || isConnecting.value) return;

  isConnecting.value = true;
  connectionError.value = '';

  const normalizedUrl = normalizeStudioUrl(studioUrl.value);

  try {
    const info = await StudioService.GetStudioInfo(normalizedUrl);
    connectedServerName.value = info.name || normalizedUrl;
    studioUrl.value = normalizedUrl;
    loginMode.value = 'studio-login';
  } catch (err) {
    console.log(err);
    connectionError.value = t('auth.login.connectionFailed');
  } finally {
    isConnecting.value = false;
  }
};

// Disconnects from the connected server and returns to URL input.
const disconnectServer = () => {
  loginMode.value = 'studio-connect';
  connectedServerName.value = '';
  connectionError.value = '';
  error.value = '';
};

// Handles enter key press to submit login.
const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    if (loginMode.value === 'studio-connect') {
      connectToServer();
    } else {
      logUserIn(username.value, password.value);
    }
  }
};

// Handles SSO error from the SSOLogin component.
const handleSSOError = (message) => {
  error.value = message;
};

// Handles SSO success from the SSOLogin component.
const handleSSOSuccess = async (data) => {
  error.value = '';
  isAwaitingResponse.value = true;

  try {
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
        modals.setModalVisibility('loginModal', false);
      } else {
        setDirectories();
      }
    }
  } catch (err) {
    console.log(err);
    isAwaitingResponse.value = false;
    error.value = 'Failed to initialize after sign-in. Please try again.';
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
  error.value = '';
  const isStudioLogin = loginMode.value === 'studio-login';
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
    .catch((err) => {
      console.log(err);
      isAwaitingResponse.value = false;
      notificationStore.errorNotification(t('notifications.errorLoggingIn'), err);
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

// Sets the login mode and resets errors.
const setLoginMode = (mode) => {
  loginMode.value = mode;
  error.value = '';
  connectionError.value = '';
};

// Validates the email format.
const validateEmail = () => {
  if (!username.value) {
    emailError.value = '';
    return;
  }
  emailError.value = isEmailValid.value ? '' : 'Please enter a valid email address';
};

// Validates the studio URL format.
const validateStudioUrl = () => {
  if (!studioUrl.value) {
    studioUrlError.value = '';
    return;
  }
  studioUrlError.value = isStudioUrlValid.value ? '' : 'Please enter a valid URL';
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
  width: 100%;
  flex-direction: column;
  align-items: center;
  justify-content: space-around;
  overflow: hidden;
}

.additional-actions {
  display: flex;
  box-sizing: border-box;
  padding: 0.5rem;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  font-weight: 300;
  font-size: 13px;
  justify-content: center;
}

.back-link {
  color: var(--white);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.3rem;
  opacity: 0.6;
  cursor: pointer;
  transition: opacity 0.2s;
}

.back-link:hover {
  opacity: 1;
}

.error-message {
  text-align: center;
  font-size: 0.8rem;
  color: var(--error-red, #ef4444);
  padding: 0.3rem 0.5rem;
}

.server-badge {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.6rem 1rem;
  margin-bottom: 0.5rem;
  border-radius: var(--large-radius);
  background-color: var(--midnight-steel);
  outline: var(--transparent-line);
  outline-offset: -1px;
  gap: 1rem;
  width: 100%;
  box-sizing: border-box;
  transition: border-radius 0.2s;
}

.server-badge:hover {
  border-radius: var(--normal-radius);
}

.server-badge-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.server-badge-details {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
}

.server-badge-name {
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--white);
  opacity: 0.9;
}

.server-badge-url {
  font-size: 0.7rem;
  color: var(--white);
  opacity: 0.4;
  font-weight: 300;
}

.server-change-link {
  font-size: 0.75rem;
  color: var(--white);
  opacity: 0.5;
  cursor: pointer;
  transition: opacity 0.2s;
}

.server-change-link:hover {
  opacity: 1;
}

.status-dot {
  position: relative;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: #22c55e;
}

.studio-reveal-link {
  font-size: 0.8rem;
  color: var(--white);
  opacity: 0.5;
  cursor: pointer;
  transition: opacity 0.2s;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.studio-reveal-link:hover {
  opacity: 0.9;
}
</style>

