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
        <!-- form container -->
        <div class="auth-form-container">
          <!-- actual-form -->
          <form @submit.prevent="handleLogin" class="auth-form">
            <!-- email -->
            <FormInput v-model="loginForm.email" :placeholder="$t('auth.login.usernamePlaceholder')" />
            <!-- password -->
            <FormInput
              v-model="loginForm.password"
              :placeholder="$t('auth.login.passwordPlaceholder')"
              isSecret
              @keydown.enter="handleEnterKey"
            />
            <!-- forgot password -->
            <div @click="showResetPassword" class="forgot-password-link">
              {{ $t('auth.login.forgotPassword') }}
            </div>
            <!-- submit button -->
            <button type="submit" class="submit-button display-font" :class="{ 'button-inactive': !isLoginFormFilled }">
              <div v-if="!isAwaitingResponse">
                {{ $t('auth.login.loginButton') }}
              </div>
              <ActionButton
                v-else
                :icon="getAppIcon('loading')"
                :isLoading="true"
                :showLabel="false"
                :noFilter="true"
              />
            </button>
            <div v-if="loadingStatus" class="loading-status">
              {{ loadingStatus }}
            </div>
          </form>
          <!-- form error -->
          <div v-if="error" class="error-message">
            {{ error }}
          </div>
        </div>
        <!-- studio URL (contextual) -->
        <div v-if="!platformStore.isWeb" class="studio-section">
          <div @click="toggleStudioLogin" class="studio-reveal-link">
            {{ $t('auth.login.connectingToStudio') }}
          </div>
          <div v-if="showStudioLogin" class="studio-url-container">
            <FormInput v-model="studioUrl" :placeholder="$t('auth.login.studioUrl')" :error="studioUrlError" :info="!studioUrlError ? $t('auth.login.studioUrlInfo') : ''" @input="validateStudioUrl" />
          </div>
        </div>

        <!-- toggle -->
        <div v-if="!isAwaitingResponse" class="additional-actions">
          <div @click="toggleLogin" class="signup-toggle">
            {{ $t('auth.login.noAccount') }}&nbsp;<span class="bold">{{ $t('auth.login.signUpLink') }}</span>
          </div>
        </div>
      </div>
      
    </div>
  </div>
</template>


<script setup>
import { ref, reactive, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

// services
import { AuthService, SettingsService } from '@/services';

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

const router = useRouter();
const { t } = useI18n();
const trayStates = useTrayStates();
const projectStore = useProjectStore();
const userStore = useUserStore();
const notificationStore = useNotificationStore();
const iconStore = useIconStore();
const themeStore = useThemeStore();
const modals = useDesktopModalStore();
const accountStore = useAccountStore();
const platformStore = usePlatformStore();

// refs
const error = ref('');
const isAwaitingResponse = ref(false);
const loadingStatus = ref('');
const showStudioLogin = ref(false);
const studioUrl = ref('');
const studioUrlError = ref('');

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';
import FormInput from '@/instances/desktop/components/FormInput.vue';

// vars
const loginForm = reactive({
  email: '',
  password: '',
});

const errors = reactive({
  email: '',
  password: '',
});

// computed props
const isLoginFormFilled = computed(() => {
  return loginForm.email && loginForm.password
});

// methods

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

const projectDirectoryExists = ref(false);

const toggleLogin = () => {
  router.push('/auth/signup')
};

const showResetPassword = () => {
  router.push('/auth/forgot-password')
};

const toggleStudioLogin = () => {
  showStudioLogin.value = !showStudioLogin.value;
  if (!showStudioLogin.value) {
    studioUrl.value = '';
    studioUrlError.value = '';
  }
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

const normalizeStudioUrl = (url) => {
  if (!url) return '';
  let normalized = url.trim();
  // Remove trailing slash
  normalized = normalized.replace(/\/+$/, '');
  // Ensure https:// prefix if no protocol
  if (!normalized.startsWith('http://') && !normalized.startsWith('https://')) {
    normalized = 'https://' + normalized;
  }
  return normalized;
};

const handleLogin = async () => {
  isAwaitingResponse.value = true;
  error.value = '';
  loadingStatus.value = t('auth.login.authenticating');

  // Determine if this is a studio login
  const isStudioLogin = showStudioLogin.value && studioUrl.value.trim();
  const normalizedStudioUrl = isStudioLogin ? normalizeStudioUrl(studioUrl.value) : '';
  
  try {
    let data;
    if (isStudioLogin) {
      // Login to studio server - authMode 'studio', studioId can be empty
      data = await AuthService.LoginWithHost(loginForm.email, loginForm.password, normalizedStudioUrl, 'studio', '');
    } else {
      // Regular global login
      data = await AuthService.Login(loginForm.email, loginForm.password);
    }
    
    userStore.user = data.user;
    userStore.isUserAuthenticated = true;

    // Initialize stores that require authentication
    loadingStatus.value = t('auth.login.loadingAccount');
    await accountStore.initialize();
    
    loadingStatus.value = t('auth.login.applyingTheme');
    await themeStore.initializeTheme();
    
    loadingStatus.value = t('auth.login.loadingStudios');
    await projectStore.loadStudios();

    projectDirectoryExists.value = await SettingsService.GetProjectDirectory();

    if(projectDirectoryExists.value){
      loadingStatus.value = t('auth.login.loadingProjects');
      await projectStore.loadProjects();
      trayStates.refreshData();
    } else {
      setDirectories();
    }

    // Mark stores as initialized so router doesn't re-init
    markStoresInitialized();
    
    // Navigate to home after successful login
    if (isStudioLogin) {
      notificationStore.addNotification(t('auth.login.studioLoginTitle'), t('auth.login.studioLoginSuccess', { url: normalizedStudioUrl }), "●");
    }
    router.push(platformStore.isWeb ? '/profile' : '/');
  } catch (err) {
    console.log(err);
    isAwaitingResponse.value = false;
    loadingStatus.value = '';
    
    // Check if error indicates user needs verification
    const errorMessage = err.message || err.toString();
    const isUnverifiedUser = errorMessage.toLowerCase().includes('please verify your email before logging in') || 
                             errorMessage.toLowerCase().includes('account not verified');
    
    if (isUnverifiedUser) {
      notificationStore.addNotification(t('auth.login.verificationRequired'), t('auth.login.checkEmailForCode'), "info");
      // Store credentials for verification page and navigate
      userStore.setPendingVerification(loginForm.email, loginForm.password);
      // Resend verification token
      AuthService.ResendToken(loginForm.email).catch(() => {});
      router.push('/verify-email');
    } else if (isStudioLogin) {
      // Handle studio login errors
      notificationStore.errorNotification(t('auth.login.studioLoginFailed'), errorMessage || t('auth.login.studioConnectionError'));
      error.value = errorMessage || t('auth.login.studioConnectionError');
    } else {
      // Handle other login errors normally
      notificationStore.errorNotification(t('auth.login.errorLoggingIn'), t('auth.login.checkCredentials'));
    }
  }
};

const setDirectories = async () => {
	  modals.setModalVisibility('dirOnboardModal', true);
};

const loadProjects = async () => {
      await projectStore.loadProjects();
      trayStates.refreshData();
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    handleLogin();
  }
};

onMounted( async () => {
})
</script>


<style scoped>
@import "@/assets/desktop.css";

.studio-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0 1rem;
  box-sizing: border-box;
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

.additional-actions{
  display: flex;
  box-sizing: border-box;
  padding: .5rem;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  width: 100%;
  font-weight: 300;
  font-size: 14px;
  justify-content: center;
}

.signup-toggle {
  color: var(--white);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: .3rem;
  opacity: 0.6;
  cursor: pointer;
  transition: opacity 0.2s;
}

.signup-toggle:hover {
  opacity: 1;
}

.studio-url-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  width: 100%;
}

.button-inactive {
  opacity: 0.5;
  pointer-events: none;
}

.forgot-password-link {
  display: flex;
  justify-content: flex-end;
  font-size: 0.875rem;
  color: var(--white);
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
  color: var(--white);
  opacity: 0.7;
  margin-top: 0.5rem;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 0.5; }
  50% { opacity: 0.9; }
}

</style>

