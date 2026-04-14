<template>
  <div class="page-root welcome-page-root">

    <div class="auth-root">

      <!-- header -->
      <div class="header-container">
        <ClusttaLogo :colored="true" :inverted="true" />
        <div class="auth-header">{{ $t('auth.welcome.title') }}</div>
      </div>

      <div class="auth-container">

        <!-- path cards -->
        <div class="welcome-cards-container">

          <div class="welcome-subheader">{{ $t('auth.welcome.subtitle') }}</div>

          <!-- Path A: Work Locally (desktop only) -->
          <OptionCard v-if="!platformStore.isWeb" :icon="getAppIcon('monitor')" :title="$t('auth.welcome.localTitle')" :description="$t('auth.welcome.localDescription')" :loading="isEnablingOffline" @select="enableOfflineMode" />

          <!-- Path B: Personal + Collaborate -->
          <OptionCard :icon="getAppIcon('website')" :title="$t('auth.welcome.personalTitle')" :description="$t('auth.welcome.personalDescription')" @select="goToSignUp" />

          <!-- Path C: Team / Studio (ClusttaCloud) -->
          <OptionCard :icon="getAppIcon('clustta')" :title="$t('auth.welcome.teamTitle')" :description="$t('auth.welcome.teamDescription')" @select="goToStudioSetup" />

          <!-- Path D: Studio Server (self-hosted) -->
          <OptionCard :icon="getAppIcon('stall')" :title="$t('auth.welcome.studioTitle')" :description="$t('auth.welcome.studioDescription')" @select="goToSelfHosted" />

        </div>

        <!-- existing account link -->
        <div class="welcome-footer">
          <div @click="goToLogin" class="signin-link">
            {{ $t('auth.welcome.haveAccount') }}&nbsp;<span class="bold">{{ $t('auth.welcome.signInLink') }}</span>
          </div>
        </div>

        <!-- loading status -->
        <div v-if="loadingStatus" class="loading-status">{{ loadingStatus }}</div>

        <!-- error -->
        <div v-if="error" class="error-message">{{ error }}</div>

      </div>

    </div>

  </div>
</template>

<script setup>
// imports
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

// components
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';
import OptionCard from '@/instances/common/components/OptionCard.vue';

// services
import { AuthService, SettingsService } from '@/services';

// stores
const accountStore = useAccountStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const themeStore = useThemeStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

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

// refs
const error = ref('');
const isAwaitingResponse = ref(false);
const isEnablingOffline = ref(false);
const loadingStatus = ref('');

// methods/functions

// Enables offline mode and navigates to the home screen.
const enableOfflineMode = async () => {
  if (isAwaitingResponse.value) return;

  isAwaitingResponse.value = true;
  isEnablingOffline.value = true;
  error.value = '';
  loadingStatus.value = t('auth.login.enablingOfflineMode');

  try {
    await AuthService.EnableOfflineMode();

    userStore.user = {
      id: 'offline-user',
      username: 'offline',
      email: 'offline@local',
      first_name: 'Offline',
      last_name: 'User',
      photo: null
    };
    userStore.isUserAuthenticated = true;

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

    notificationStore.addNotification(
      t('auth.login.offlineModeTitle'),
      t('auth.login.offlineModeMessage'),
      'success'
    );
    router.push('/');
  } catch (err) {
    console.error('Failed to enable offline mode:', err);
    error.value = t('auth.login.offlineModeFailed');
    loadingStatus.value = '';
    notificationStore.errorNotification(
      t('auth.login.offlineModeErrorTitle'),
      t('auth.login.offlineModeErrorMessage')
    );
  } finally {
    isAwaitingResponse.value = false;
    isEnablingOffline.value = false;
  }
};

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Navigates to the login page.
const goToLogin = () => {
  router.push('/auth/login');
};

// Navigates to the sign up page.
const goToSignUp = () => {
  accountStore.onboardingIntent = 'personal';
  router.push('/auth/signup');
};

// Navigates to signup with studio intent for ClusttaCloud team creation.
const goToStudioSetup = () => {
  accountStore.onboardingIntent = 'studio';
  router.push('/auth/signup');
};

// Navigates directly to studio setup for self-hosted server registration.
const goToSelfHosted = () => {
  accountStore.onboardingIntent = 'self-hosted';
  router.push('/auth/studio-setup');
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.welcome-page-root {
  display: flex;
  box-sizing: border-box;
  width: 100%;
  align-items: center;
  height: 100%;
  flex-direction: column;
  background-color: var(--black);
  overflow: hidden;
  overflow-y: auto;
}

.welcome-subheader {
  font-size: 0.9rem;
  color: var(--white);
  opacity: 0.6;
  text-align: left;
  font-weight: 300;
}

.welcome-cards-container {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  width: 100%;
  min-width: 400px;
}



.welcome-footer {
  display: flex;
  box-sizing: border-box;
  padding: 0.5rem;
  align-items: center;
  justify-content: center;
  width: 100%;
  font-weight: 300;
  font-size: 14px;
}

.signin-link {
  color: var(--white);
  font-size: 14px;
  font-weight: 300;
  opacity: 0.6;
  cursor: pointer;
  transition: opacity 0.2s;
}

.signin-link:hover {
  opacity: 1;
}

.button-inactive {
  opacity: 0.5;
  pointer-events: none;
}

.loading-status {
  text-align: center;
  font-size: 0.85rem;
  color: var(--white);
  opacity: 0.7;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 0.5; }
  50% { opacity: 0.9; }
}

@media (max-width: 768px) {
  .welcome-cards-container {
    min-width: 300px;
  }
}
</style>
