<template>

  <div class="modal-container" v-esc="closeModal" v-return="handleEnterKey">
    <HeaderArea :title="title" :icon="'login'" />

    <div class="general-container">

      <div class="login-form">
        <input v-model="username" class="input-short" type="text" placeholder="Email address" autocomplete="off"/>
        <input v-model="password" class="input-short" type="password" placeholder="Password"
          @keydown.enter="handleEnterKey" />
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
import { AuthService, SettingsService } from "@/../bindings/clustta/services";

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';

// state imports
import { useTrayStates } from '@/stores/TrayStates';
import { useProjectStore } from '@/stores/projects';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useUserStore } from '@/stores/users';
import { useThemeStore } from '@/stores/theme';
import { useAccountStore } from '@/stores/accounts';
import { useStageStore } from '@/stores/stages';

let username = ref('');
let password = ref('');
const isAwaitingResponse = ref(false);
const isCheckingAuth = ref(true);
const eulaAccepted = ref(false);
const projectDirectoryExists = ref(false);

// stores/states
const modals = useDesktopModalStore();
const trayStates = useTrayStates();
const projectStore = useProjectStore();
const userStore = useUserStore();
const notificationStore = useNotificationStore();
const themeStore = useThemeStore();
const accountStore = useAccountStore();
const stageStore = useStageStore();

// computed props
const isValueChanged = computed(() => {
  const usernameValid = username.value !== '';
  const passwordValid = password.value !== '';
  return usernameValid && passwordValid;
});

// methods
const closeModal = () => {
  modals.disableAllModals();
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
  await AuthService.Login(username, password)
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
  width: 50%;
  width: 500px;
  padding: 1rem 2rem;
  /* width: max-content; */
  flex-direction: column;
  align-items: center;
  justify-content: space-around;
  overflow: hidden;
  box-sizing: border-box;
  /* background-color: firebrick; */
}

.input-short {
  width: 90%;
  height: 50px;
}

.login-button {
  margin-top: 1rem;
  height: 50px;
}
</style>

