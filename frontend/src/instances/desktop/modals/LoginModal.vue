<template>
  <div class="modal-container" v-esc="closeModal" v-return="handleEnterKey">
    <HeaderArea :title="title" :icon="'login'" />
    <div class="general-container">
      <div class="login-form">
        <FormInput
          v-if="showStudioLogin"
          v-model="studioUrl"
          :placeholder="$t('placeholders.studioUrl')"
          needsValidation
          :error="studioUrlError"
          :valid="isStudioUrlValid"
          :showValidation="!!studioUrl"
          @input="validateStudioUrl"
        />
        <FormInput
          v-model="username"
          :placeholder="$t('placeholders.emailAddress')"
          needsValidation
          :error="emailError"
          :valid="isEmailValid"
          :showValidation="!!username"
          @input="validateEmail"
        />
        <FormInput v-model="password" :placeholder="$t('placeholders.password')" isSecret />
        <div class="horizontal-flex">
          <ActionButton :isInactive="true" :icon="getAppIcon('two-drives')" :label="$t('modals.privateServer')" />
          <ToggleSwitch  @click="toggleStudioLogin" :switchValueProp="showStudioLogin" />
        </div>
      </div>
      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.logIn')" :fullWidth="true" @click="logUserIn(username, password)"
          :isActive="isValueChanged" :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import FormInput from '@/instances/desktop/components/FormInput.vue';
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
const emailError = ref('');
const isAwaitingResponse = ref(false);
const password = ref('');
const projectDirectoryExists = ref(false);
const showStudioLogin = ref(false);
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

// Returns whether the studio URL is valid.
const isStudioUrlValid = computed(() => {
  if (!studioUrl.value) return false;
  const urlRegex = /^(https?:\/\/)?([\w-]+\.)+[\w-]+(\/[\w-./?%&=]*)?$/;
  return urlRegex.test(studioUrl.value.trim());
});

// Returns whether the form has valid values.
const isValueChanged = computed(() => {
  const passwordValid = password.value !== '';
  if (showStudioLogin.value) {
    return isEmailValid.value && passwordValid && isStudioUrlValid.value;
  }
  return isEmailValid.value && passwordValid;
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
      notificationStore.errorNotification(t('notifications.errorLoggingIn'), error);
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

// Toggles studio login mode on/off.
const toggleStudioLogin = () => {
  showStudioLogin.value = !showStudioLogin.value;
  if (!showStudioLogin.value) {
    studioUrl.value = '';
    studioUrlError.value = '';
  }
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

</style>

