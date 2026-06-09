<template>
  <div class="page-root login-page-root">

    <!-- responsive root -->
    <div class="auth-root">
      <!-- header -->
      <div class="header-container">
        <ClusttaLogo :colored="true" :inverted="true" />
        <div class="auth-header">
          {{ $t('auth.login.title') }}
        </div>
      </div>

      <div class="auth-container">

        <!-- Cloud login -->
        <template v-if="loginMode === 'cloud'">
          <AuthLoader v-if="isInitializing" :status="loadingStatus" />

          <template v-else>
          <div class="auth-form-container">
            <form @submit.prevent="handleLogin" class="auth-form">
              <FormInput v-model="loginForm.email" :placeholder="$t('auth.login.usernamePlaceholder')" />

              <FormInput v-model="loginForm.password" :placeholder="$t('auth.login.passwordPlaceholder')" isSecret @keydown.enter="handleEnterKey" />

              <div @click="showResetPassword" class="forgot-password-link">
                {{ $t('auth.login.forgotPassword') }}
              </div>

              <button type="submit" class="submit-button display-font" :class="{ 'button-inactive': !isLoginFormFilled }">
                <div v-if="!isAwaitingResponse">{{ $t('auth.login.loginButton') }}</div>
                <ActionButton v-else :icon="getAppIcon('loading')" :isLoading="true" :showLabel="false" :noFilter="true" />
              </button>

              <div v-if="loadingStatus" class="loading-status">{{ loadingStatus }}</div>
            </form>

            <SSOLogin v-if="!isAwaitingResponse" mode="login" @success="handleSSOSuccess" @error="handleSSOError" />

            <div v-if="error" class="error-message">{{ error }}</div>
          </div>

          <div v-if="!platformStore.isWeb && !isAwaitingResponse" class="additional-actions">
            <div @click="setLoginMode('studio-connect')" class="studio-reveal-link">{{ $t('auth.login.connectToStudio') }}</div>
          </div>

          <div v-if="!isAwaitingResponse" class="additional-actions">
            <div @click="toggleLogin" class="signup-toggle">
              {{ $t('auth.login.noAccount') }}&nbsp;<span class="bold">{{ $t('auth.login.signUpLink') }}</span>
            </div>
          </div>
          </template>
        </template>

        <!-- Studio URL connect -->
        <template v-if="loginMode === 'studio-connect'">
          <div class="auth-form-container">
            <FormInput v-model="studioUrl" :placeholder="$t('auth.login.studioUrl')" :error="studioUrlError" :info="!studioUrlError ? $t('auth.login.studioUrlInfo') : ''" @input="validateStudioUrl" />

            <button class="submit-button display-font" :class="{ 'button-inactive': !isUrlValid }" @click="connectToServer">
              <div v-if="!isConnecting">{{ $t('auth.login.connectButton') }}</div>
              <ActionButton v-else :icon="getAppIcon('loading')" :isLoading="true" :showLabel="false" />
            </button>

            <div v-if="connectionError" class="error-message">{{ connectionError }}</div>
          </div>

          <div class="additional-actions">
            <div @click="setLoginMode('cloud')" class="back-link">{{ $t('auth.login.loginToCloud') }}</div>
          </div>
        </template>

        <!-- Studio login (connected) -->
        <template v-if="loginMode === 'studio-login'">
          <div class="auth-form-container">
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

            <form @submit.prevent="handleLogin" class="auth-form">
              <FormInput v-model="loginForm.email" :placeholder="$t('auth.login.usernamePlaceholder')" />

              <FormInput v-model="loginForm.password" :placeholder="$t('auth.login.passwordPlaceholder')" isSecret @keydown.enter="handleEnterKey" />

              <button type="submit" class="submit-button display-font" :class="{ 'button-inactive': !isLoginFormFilled }">
                <div v-if="!isAwaitingResponse">{{ $t('auth.login.loginButton') }}</div>
                <ActionButton v-else :icon="getAppIcon('loading')" :isLoading="true" :showLabel="false" :noFilter="true" />
              </button>

              <div v-if="loadingStatus" class="loading-status">{{ loadingStatus }}</div>
            </form>

            <div v-if="error" class="error-message">{{ error }}</div>
          </div>

          <div v-if="!isAwaitingResponse" class="additional-actions">
            <div @click="goToStudioSignUp" class="back-link">
              {{ $t('auth.login.signUpToStudio') }}
            </div>

            <div @click="backToCloudLogin" class="back-link">{{ $t('auth.login.loginToCloud') }}</div>
          </div>
        </template>

      </div>

    </div>
  </div>
</template>


<script setup>

// imports
import { ref, reactive, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import AuthLoader from '@/instances/desktop/components/AuthLoader.vue';
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';
import FormInput from '@/instances/desktop/components/FormInput.vue';
import SSOLogin from '@/instances/desktop/components/SSOLogin.vue';

// services
import { AuthService, SettingsService, StudioService } from '@/services';

// store imports
import { useAccountStore } from '@/stores/accounts';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';
import { useThemeStore } from '@/stores/theme';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';
import { markStoresInitialized } from '@/router';

// stores
const accountStore = useAccountStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const themeStore = useThemeStore();
const userStore = useUserStore();

const route = useRoute();
const router = useRouter();
const trayStates = useTrayStates();
const { t } = useI18n();

// refs
const connectedServerName = ref('');
const connectionError = ref('');
const error = ref('');
const isAwaitingResponse = ref(false);
const isConnecting = ref(false);
const isInitializing = ref(false);
const loadingStatus = ref('');
const loginMode = ref('cloud');
const studioUrl = ref('');
const studioUrlError = ref('');

// vars
const loginForm = reactive({
  email: '',
  password: '',
});

const errors = reactive({
  email: '',
  password: '',
});

// computed properties
const isLoginFormFilled = computed(() => {
  return loginForm.email && loginForm.password;
});

const isUrlValid = computed(() => {
  return studioUrl.value.trim() && !studioUrlError.value;
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

// Connects to the studio server and retrieves its info.
const connectToServer = async () => {
  if (!isUrlValid.value || isConnecting.value) return;

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

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles SSO error from the SSOLogin component.
const handleSSOError = (message) => {
  error.value = message;
};

// Handles SSO success from the SSOLogin component.
const handleSSOSuccess = async (data) => {
  error.value = '';
  isInitializing.value = true;

  try {
    userStore.user = data.user;
    userStore.isUserAuthenticated = true;

    const savedIntent = accountStore.onboardingIntent;
    loadingStatus.value = t('auth.login.loadingAccount');
    await accountStore.initialize();

    loadingStatus.value = t('auth.login.applyingTheme');
    await themeStore.initializeTheme();

    loadingStatus.value = t('auth.login.loadingStudios');
    await projectStore.loadStudios();

    const projectDirectoryExists = await SettingsService.GetProjectDirectory();
    if (projectDirectoryExists) {
      loadingStatus.value = t('auth.login.loadingProjects');
      await projectStore.loadProjects();
      trayStates.refreshData();
    } else {
      modals.setModalVisibility('dirOnboardModal', true);
    }

    markStoresInitialized();

    // Route based on onboarding intent
    if (savedIntent === 'studio') {
      router.push('/auth/studio-setup');
    } else {
      router.push('/');
    }
  } catch (err) {
    console.log(err);
    isInitializing.value = false;
    loadingStatus.value = '';
    error.value = 'Failed to initialize after sign-in. Please try again.';
  }
};

// Navigates to the studio setup page with the connected server URL.
const goToStudioSignUp = () => {
  router.push({
    path: '/auth/studio-setup',
    query: { type: 'self-hosted', url: studioUrl.value, name: connectedServerName.value }
  });
};

// Handles the login form submission.
const handleLogin = async () => {
  isAwaitingResponse.value = true;
  error.value = '';
  loadingStatus.value = t('auth.login.authenticating');

  const isStudioLogin = loginMode.value === 'studio-login';
  const normalizedUrl = isStudioLogin ? normalizeStudioUrl(studioUrl.value) : '';

  try {
    let data;
    if (isStudioLogin) {
      data = await AuthService.LoginWithHost(loginForm.email, loginForm.password, normalizedUrl, 'studio', '');
    } else {
      data = await AuthService.Login(loginForm.email, loginForm.password);
    }

    userStore.user = data.user;
    userStore.isUserAuthenticated = true;

    isAwaitingResponse.value = false;
    isInitializing.value = true;

    const savedIntent = accountStore.onboardingIntent;
    loadingStatus.value = t('auth.login.loadingAccount');
    await accountStore.initialize();

    loadingStatus.value = t('auth.login.applyingTheme');
    await themeStore.initializeTheme();

    loadingStatus.value = t('auth.login.loadingStudios');
    await projectStore.loadStudios();

    const projectDirectoryExists = await SettingsService.GetProjectDirectory();

    if (projectDirectoryExists) {
      loadingStatus.value = t('auth.login.loadingProjects');
      await projectStore.loadProjects();
      trayStates.refreshData();
    } else {
      modals.setModalVisibility('dirOnboardModal', true);
    }

    markStoresInitialized();

    if (isStudioLogin) {
      notificationStore.addNotification(t('auth.login.studioLoginTitle'), t('auth.login.studioLoginSuccess', { url: normalizedUrl }), '●');
    }

    // Route based on onboarding intent
    if (savedIntent === 'studio') {
      router.push('/auth/studio-setup');
    } else {
      router.push(platformStore.isWeb ? '/profile' : '/');
    }
  } catch (err) {
    console.log(err);
    isAwaitingResponse.value = false;
    isInitializing.value = false;
    loadingStatus.value = '';

    const errorMessage = err.message || err.toString();
    const isUnverifiedUser = errorMessage.toLowerCase().includes('please verify your email before logging in') ||
                             errorMessage.toLowerCase().includes('account not verified');

    if (isUnverifiedUser) {
      notificationStore.addNotification(t('auth.login.verificationRequired'), t('auth.login.checkEmailForCode'), 'info');
      userStore.setPendingVerification(loginForm.email, loginForm.password);
      AuthService.ResendToken(loginForm.email).catch(() => {});
      router.push('/verify-email');
    } else if (isStudioLogin) {
      notificationStore.errorNotification(t('auth.login.studioLoginFailed'), errorMessage || t('auth.login.studioConnectionError'));
      error.value = errorMessage || t('auth.login.studioConnectionError');
    } else {
      notificationStore.errorNotification(t('auth.login.errorLoggingIn'), t('auth.login.checkCredentials'));
    }
  }
};

// Handles enter key press on the password field.
const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    handleLogin();
  }
};

// Normalizes a studio URL by ensuring protocol and removing trailing slash.
const normalizeStudioUrl = (url) => {
  if (!url) return '';
  let normalized = url.trim();
  normalized = normalized.replace(/\/+$/, '');
  if (!normalized.startsWith('http://') && !normalized.startsWith('https://')) {
    normalized = 'https://' + normalized;
  }
  return normalized;
};

// Sets the login mode and resets errors.
const setLoginMode = (mode) => {
  loginMode.value = mode;
  error.value = '';
  connectionError.value = '';
};

// Navigates to the forgot password page.
const showResetPassword = () => {
  router.push('/auth/forgot-password');
};

// Navigates to the sign up page.
const toggleLogin = () => {
  router.push('/auth/signup');
};

// Validates the studio URL format.
const validateStudioUrl = () => {
  if (!studioUrl.value) {
    studioUrlError.value = '';
    return;
  }

  const urlPattern = /^https?:\/\/[a-zA-Z0-9][-a-zA-Z0-9]*(\.[a-zA-Z0-9][-a-zA-Z0-9]*)+(:\d+)?(\/.*)?$/;

  if (!studioUrl.value.startsWith('http://') && !studioUrl.value.startsWith('https://')) {
    studioUrlError.value = t('auth.login.urlMustStartWith');
  } else if (!urlPattern.test(studioUrl.value)) {
    studioUrlError.value = t('auth.login.invalidUrl');
  } else {
    studioUrlError.value = '';
  }
};

// lifecycle hooks
onMounted(async () => {
  const queryUrl = route.query.studioUrl;
  const queryName = route.query.name;
  if (queryUrl) {
    studioUrl.value = queryUrl;
    connectedServerName.value = queryName || queryUrl;
    loginMode.value = 'studio-login';
    await connectToServer();
  }
})
</script>


<style scoped>
@import "@/assets/desktop.css";

.additional-actions {
  display: flex;
  box-sizing: border-box;
  padding: 0.5rem;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  width: 100%;
  font-weight: 300;
  font-size: 14px;
  justify-content: center;
}

.back-link {
  color: hsl(var(--foreground));
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

.button-inactive {
  opacity: 0.5;
  pointer-events: none;
}

.forgot-password-link {
  display: flex;
  justify-content: flex-end;
  font-size: 0.875rem;
  color: hsl(var(--muted-foreground));
  opacity: 0.6;
  cursor: pointer;
  transition: opacity 0.2s;
  margin-bottom: 1rem;
}

.forgot-password-link:hover {
  opacity: 1;
}

.loading-status {
  text-align: center;
  font-size: 0.85rem;
  color: hsl(var(--muted-foreground));
  opacity: 0.7;
  margin-top: 0.5rem;
  animation: pulse 1.5s ease-in-out infinite;
}

.server-badge {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.6rem 1rem;
  margin-bottom: 0.5rem;
  border-radius: var(--radius);
  background-color: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  gap: 1rem;
  transition: border-radius 0.2s;
}

.server-badge:hover {
  border-color: hsl(var(--ring));
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
  color: hsl(var(--foreground));
}

.server-badge-url {
  font-size: 0.7rem;
  color: hsl(var(--muted-foreground));
  font-weight: 300;
}

.server-change-link {
  font-size: 0.75rem;
  color: hsl(var(--muted-foreground));
  cursor: pointer;
  transition: opacity 0.2s;
}

.server-change-link:hover {
  opacity: 1;
}

.signup-toggle {
  color: hsl(var(--foreground));
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.3rem;
  opacity: 0.6;
  cursor: pointer;
  transition: opacity 0.2s;
}

.signup-toggle:hover {
  opacity: 1;
}

.status-dot {
  position: relative;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: hsl(var(--success));
  animation: dot-entrance 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
}

.status-dot::before,
.status-dot::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background-color: hsl(var(--success));
  animation: ripple 2.5s cubic-bezier(0.4, 0, 0.2, 1) infinite;
}

.status-dot::after {
  animation-delay: 1.25s;
}

.studio-reveal-link {
  font-size: 0.8rem;
  color: hsl(var(--muted-foreground));
  cursor: pointer;
  transition: opacity 0.2s;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.studio-reveal-link:hover {
  opacity: 0.9;
}

@keyframes dot-entrance {
  0% { transform: scale(0); }
  60% { transform: scale(1.3); }
  80% { transform: scale(0.9); }
  100% { transform: scale(1); }
}

@keyframes pulse {
  0%, 100% { opacity: 0.5; }
  50% { opacity: 0.9; }
}

@keyframes ripple {
  0% { opacity: 0.5; transform: scale(1); }
  100% { opacity: 0; transform: scale(3); }
}

</style>

