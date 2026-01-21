<template>
  <div class="page-root login-page-root">

    <!-- responsive root -->
    <div class="auth-root">

      <!-- header -->
      <div class="header-container">
        <ClusttaLogo :colored="true" :inverted="true" />
        <div class="auth-header">
          Login to Clustta
        </div>
      </div>

      <div class="auth-container">

        <!-- form container -->
        <div class="auth-form-container">

          <div class="horizontal-flex studio-toggle-row">
            <ActionButton :isInactive="true" :icon="getAppIcon('two-drives')" :label="'Private Server'" />
            <ToggleSwitch @click="toggleStudioLogin" :switchValueProp="showStudioLogin" />
          </div>

          <div v-if="showStudioLogin" class="studio-url-container">
            <div class="form-group">
              <div class="compound-form-input">
                <input autocomplete="off" class="form-input-mini" placeholder="Studio URL (e.g., https://studio.mycompany.com)" v-model="studioUrl" type="text" />
              </div>
            </div>
            <div class="studio-url-hint">
              Enter the URL of your studio server, then login with your studio credentials below.
            </div>
          </div>

          <!-- actual-form -->
          <form @submit.prevent="handleLogin" class="auth-form">

            <!-- email -->
            <div class="form-group">
              <div class="compound-form-input">
                <input autocomplete="off" class="form-input-mini" placeholder="Username or Email address" v-model="loginForm.email" type="text"
                  required />
              </div>
            </div>

            <!-- password -->
            <div class="form-group">
              <div class="compound-form-input">
                <input autocomplete="new-password" class="form-input-mini" placeholder="Password" v-model="loginForm.password"
                  :type="isPasswordVisible ? 'text' : 'password'" required  @keydown.enter="handleEnterKey">
                <ActionButton 
                  v-if="loginForm.password"
                  v-tooltip="isPasswordVisible ? 'Hide Password' : 'Show Password'"
                  :icon="isPasswordVisible ? getAppIcon('eye-cancel') : getAppIcon('eye')"
                  @click="togglePasswordVisibility"
                  :showLabel="false"
                />
              </div>
            </div>

            <!-- submit button -->
            <button type="submit" class="submit-button display-font" :class="{ 'button-inactive': !isLoginFormFilled }">
              <div v-if="!isAwaitingResponse">
                Login
              </div>
              <ActionButton
                v-else
                :icon="getAppIcon('loading')"
                :isLoading="true"
                :showLabel="false"
                :noFilter="true"
              />
            </button>

          </form>

          <!-- form error -->
          <div v-if="error" class="error-message">
            {{ error }}
          </div>

        </div>

        <!-- toggle -->
        <div @click="toggleLogin" class=" toggle-container">
            No account?
            <div class="bold" >
              SignUp
            </div>
        </div>

        <!-- forgot password -->
        <div @click="showResetPassword" class="toggle-container">
            <div class="bold" >
              Forgot password?
            </div>
        </div>

        <div @click="enableOfflineMode" class="toggle-container offline-toggle" :class="{ 'button-inactive': isAwaitingResponse }">
          <div class="bold">
            Use without an account
          </div>
        </div>

      </div>
      
    </div>
  </div>
</template>


<script setup>
import { ref, reactive, computed, onMounted, onBeforeMount } from 'vue';
import { useRouter } from 'vue-router';
import { useTrayStates } from '@/stores/TrayStates';
import { useProjectStore } from '@/stores/projects';
import { AuthService } from "@/services";
import { useNotificationStore } from '@/stores/notifications';
import { useUserStore } from '@/stores/users';
import { useIconStore } from '@/stores/icons';
import { useThemeStore } from '@/stores/theme';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useAccountStore } from '@/stores/accounts';
import { markStoresInitialized } from '@/router';
import utils from "@/services/utils";

import { SettingsService } from '@/services';

const router = useRouter();
const trayStates = useTrayStates();
const projectStore = useProjectStore();
const userStore = useUserStore();
const notificationStore = useNotificationStore();
const iconStore = useIconStore();
const themeStore = useThemeStore();
const modals = useDesktopModalStore();
const accountStore = useAccountStore();

// refs
const error = ref('');
const isAwaitingResponse = ref(false);
const isPasswordVisible = ref(false)
const showStudioLogin = ref(false)
const studioUrl = ref('')

// components
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

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
const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

const showPassword = () => {
  isPasswordVisible.value = true
};

const hidePassword = () => {
  isPasswordVisible.value = false
};

const togglePasswordVisibility = () => {
  isPasswordVisible.value = !isPasswordVisible.value;
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
    await accountStore.initialize();
    await themeStore.initializeTheme();
    await projectStore.loadStudios();

    projectDirectoryExists.value = await SettingsService.GetProjectDirectory();

    if(projectDirectoryExists.value){
      await projectStore.loadProjects();
      trayStates.refreshData();
    } else {
      setDirectories();
    }

    // Mark stores as initialized so router doesn't re-init
    markStoresInitialized();
    
    // Navigate to home after successful login
    if (isStudioLogin) {
      notificationStore.addNotification("Studio Login", `Successfully logged in to ${normalizedStudioUrl}`, "success");
    }
    router.push('/');
  } catch (err) {
    console.log(err);
    isAwaitingResponse.value = false;
    
    // Check if error indicates user needs verification
    const errorMessage = err.message || err.toString();
    const isUnverifiedUser = errorMessage.toLowerCase().includes('please verify your email before logging in') || 
                             errorMessage.toLowerCase().includes('account not verified');
    
    if (isUnverifiedUser) {
      notificationStore.addNotification("Verification Required", "Please check your email for a verification code.", "info");
      // Store credentials for verification page and navigate
      userStore.setPendingVerification(loginForm.email, loginForm.password);
      // Resend verification token
      AuthService.ResendToken(loginForm.email).catch(() => {});
      router.push('/verify-email');
    } else if (isStudioLogin) {
      // Handle studio login errors
      notificationStore.errorNotification("Studio Login Failed", errorMessage || 'Could not connect to studio server. Please check the URL and credentials.');
      error.value = errorMessage || 'Could not connect to studio server';
    } else {
      // Handle other login errors normally
      notificationStore.errorNotification("Error Logging In", 'Please check your credentials and try again');
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

const enableOfflineMode = async () => {
  if (isAwaitingResponse.value) return;
  
  isAwaitingResponse.value = true;
  error.value = '';

  try {
    await AuthService.EnableOfflineMode();
    
    // Set up the offline user in the user store
    userStore.user = {
      id: 'offline-user',
      username: 'offline',
      email: 'offline@local',
      first_name: 'Offline',
      last_name: 'User',
      photo: null
    };
    userStore.isUserAuthenticated = true;
    
    // Initialize stores
    await accountStore.initialize();
    await themeStore.initializeTheme();
    await projectStore.loadStudios();
    
    projectDirectoryExists.value = await SettingsService.GetProjectDirectory();

    if (projectDirectoryExists.value) {
      await projectStore.loadProjects();
      trayStates.refreshData();
    } else {
      setDirectories();
    }

    // Mark stores as initialized so router doesn't re-init
    markStoresInitialized();
    
    // Navigate to home after successful offline setup
    notificationStore.addNotification("Offline Mode", "You're now using Clustta in offline mode. Some features will be limited.", "info");
    router.push('/');
  } catch (err) {
    console.error('Failed to enable offline mode:', err);
    error.value = 'Failed to enable offline mode. Please try again.';
    notificationStore.errorNotification("Offline Mode Error", 'Failed to enable offline mode');
  } finally {
    isAwaitingResponse.value = false;
  }
};

onMounted( async () => {
})
</script>


<style scoped>
@import "@/assets/desktop.css";

.divider-container {
  width: 90%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  margin: 0.5rem 0;
}

.divider-line {
  flex: 1;
  height: 1px;
  background: var(--white-20);
}

.divider-text {
  color: var(--white-60);
  font-size: 0.85rem;
  text-transform: uppercase;
}

.offline-toggle {
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.75rem 1rem;
  border: 1px solid var(--white-20);
  border-radius: 8px;
  transition: all 0.2s ease;
}

.offline-toggle:hover {
  border-color: var(--white-40);
  background: var(--white-05);
}

.offline-hint {
  font-size: 0.75rem;
  color: var(--white-60);
  font-weight: normal;
}

.studio-toggle-row {
  justify-content: space-between;
  padding: 0.5rem 0;
}

.studio-url-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  /* padding: 0.75rem; */
  background: var(--white-05);
  border-radius: 8px;
  border: 1px solid var(--accent-color-30);
}

.studio-url-hint {
  font-size: 0.75rem;
  color: var(--white-60);
  text-align: center;
}

.button-inactive {
  opacity: 0.5;
  pointer-events: none;
}

</style>

