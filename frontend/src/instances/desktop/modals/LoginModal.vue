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
import { ref, computed, onBeforeMount } from 'vue';
import { AuthService, SettingsService } from "@/services";

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// state imports
import { useTrayStates } from '@/stores/TrayStates';
import { useProjectStore } from '@/stores/projects';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useUserStore } from '@/stores/users';
import { useThemeStore } from '@/stores/theme';
import { useAccountStore } from '@/stores/accounts';
import { useStageStore } from '@/stores/stages';
import { useIconStore } from '@/stores/icons';

let username = ref('');
let password = ref('');
const isAwaitingResponse = ref(false);
const isCheckingAuth = ref(true);
const eulaAccepted = ref(false);
const projectDirectoryExists = ref(false);
const isPasswordVisible = ref(false);
const showStudioLogin = ref(false);
const studioUrl = ref('');

// stores/states
const modals = useDesktopModalStore();
const trayStates = useTrayStates();
const projectStore = useProjectStore();
const userStore = useUserStore();
const notificationStore = useNotificationStore();
const themeStore = useThemeStore();
const accountStore = useAccountStore();
const stageStore = useStageStore();
const iconStore = useIconStore();

// computed props
const isValueChanged = computed(() => {
  const usernameValid = username.value !== '';
  const passwordValid = password.value !== '';
  // If studio login is enabled, also require studio URL
  if (showStudioLogin.value) {
    return usernameValid && passwordValid && studioUrl.value.trim() !== '';
  }
  return usernameValid && passwordValid;
});

// methods
const closeModal = () => {
  modals.disableAllModals();
};

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon;
};

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

// Normalizes studio URL to ensure proper format.
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

const showEula = async () => {
  projectDirectoryExists.value = await SettingsService.GetProjectDirectory();
  projectStore.projectsLoaded = !projectDirectoryExists.value;
  eulaAccepted.value = await SettingsService.GetEulaAccepted();

  if(eulaAccepted.value) return
  modals.setModalVisibility('eulaModal', true);
};

const setDirectories = async () => {
  modals.setModalVisibility('dirOnboardModal', true);
};

const loadProjects = async () => {
  await projectStore.loadProjects();
  trayStates.refreshData();
};

const logUserIn = async (username, password) => {
  isAwaitingResponse.value = true;
  
  // Determine if this is a studio login
  const isStudioLogin = showStudioLogin.value && studioUrl.value.trim();
  const normalizedStudioUrl = isStudioLogin ? normalizeStudioUrl(studioUrl.value) : '';
  
  // Use appropriate login method based on mode
  const loginPromise = isStudioLogin 
    ? AuthService.LoginWithHost(username, password, normalizedStudioUrl, 'studio', '')
    : AuthService.Login(username, password);
    
  await loginPromise
    .then(async (data) => {
      // Store user in userStore (existing behavior)
      userStore.user = data.user
      userStore.isUserAuthenticated = true
      
      // Add account to multi-account system  
      // Note: data already contains the token structure with session_id and user
      // The AuthService.Login automatically adds it via SetToken -> AddAccount
      
      // Reset stores after successful login
      userStore.$reset();
      projectStore.$reset();
      trayStates.$reset();
      
      // Close the modal after successful account switch
      modals.setModalVisibility("loginModal", false);
      
      // Set the user data again after reset
      userStore.user = data.user;
      userStore.isUserAuthenticated = true;
      
      // Refresh account store to pick up the newly added account
      await accountStore.refreshAccounts();
      projectDirectoryExists.value = await SettingsService.GetProjectDirectory();

      console.log(accountStore.accounts)
      
      // Check if this is an additional account login
      if (accountStore.isAdditionalAccount) {
        // For additional accounts, switch to the newly added account
        await accountStore.switchToAccount(data.user.id, {
          userStore,
          projectStore,
          trayStates,
          themeStore,
          notificationStore,
          stageStore
        });
        
        if(!projectDirectoryExists.value){
            setDirectories();
        }

      } else {
        
        await themeStore.initializeTheme();
        await projectStore.loadStudios();
        
        // Conditionally load projects based on directory setup
        if(projectDirectoryExists.value){
          await loadProjects();
          trayStates.refreshData();
          modals.setModalVisibility("loginModal", false);
        } else {
          setDirectories();
        }
      }
    })
    .catch((error) => {
      console.log(error)
      isAwaitingResponse.value = false;
      notificationStore.errorNotification("Error Loggin In", error)
    });
}
const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    logUserIn(username.value, password.value);
  }
};
let title = 'Login';


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

.login-button-icon {
  width: 30px;
  height: 30px;
}

.login-button-text {
  font-size: 18px;
}

.login-page-root {
  background-color: var(--black);
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
}

.clustta-logo {
  width: 80px;
  aspect-ratio: 1/1;
}

.logo-container {
  display: flex;
  width: 40%;
  width: min-content;
  align-items: center;
  justify-content: center;
  /* background-color: red; */
  box-sizing: border-box;
  overflow: hidden;
  padding: 1rem;
}

.login-large-text-container {
  color: white;
  width: min-content;
  display: flex;
  width: max-content;
  align-items: center;
  justify-content: center;
  /* background-color: red; */
  font-size: 32px;
  font-weight: 400;
  box-sizing: border-box;
  overflow: hidden;
}

.login-small-text-container {
  color: white;
  width: min-content;
  display: flex;
  width: max-content;
  align-items: center;
  justify-content: center;
  /* background-color: red; */
  font-size: 14px;
  font-weight: 100;
  box-sizing: border-box;
  overflow: hidden;
}

.login-form-container {
  flex-direction: column;
  box-sizing: border-box;
  overflow: hidden;
  /* background-color: royalblue; */
  /* padding-bottom: 5rem; */
  display: flex;
  height: 100%;
  width: 100%;
  align-items: center;
  justify-content: center;
  gap: 1rem;
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
  box-sizing: border-box;
}

.input-short {
  width: 90%;
  height: 50px;
}

.login-button {
  margin-top: 1rem;
  height: 50px;
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

.toggle-icon {
  width: 18px;
  height: 18px;
  filter: var(--icon-filter);
}

.toggle-label {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.toggle-label .bold {
  font-weight: 500;
  font-size: 0.875rem;
  color: var(--white);
}

.toggle-hint {
  font-size: 0.75rem;
  color: var(--light-grey);
}

.studio-url-group {
  margin-top: 0.5rem;
}

.studio-url-hint {
  font-size: 0.75rem;
  color: var(--light-grey);
  margin-top: 0.25rem;
  padding-left: 0.5rem;
}
</style>

