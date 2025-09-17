<template>
  <div class="page-root login-page-root">

    <!-- responsive root -->
    <div class="auth-root">

      <!-- header -->
      <div class="header-container">
        <ClusttaLogo :colored="true" :inverted="true" :boldText="true" />
        <div class="auth-header">
          Login to Clustta
        </div>
      </div>

      <div class="auth-container">

        <!-- form container -->
        <div class="auth-form-container">

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
                <input autocomplete="off" class="form-input-mini" placeholder="Password" v-model="loginForm.password"
                  :type="isPasswordVisible ? 'text' : 'password'" required  @keydown.enter="handleEnterKey">
                <div v-if="loginForm.password" class="form-input-icon" @mousedown="showPassword"
                  @mouseup="hidePassword" @mouseleave="hidePassword">
                  <img class="alert-icons" :src="isPasswordVisible ? getAppIcon('eye-cancel') : getAppIcon('eye')" />
                </div>
              </div>
            </div>

            <!-- submit button -->
            <button type="submit" class="submit-button display-font" :class="{ 'button-inactive': !isLoginFormFilled }">
              <div v-if="!isAwaitingResponse">
                Login
              </div>
              <div v-else class="submit-button-icon loading-icon">
                <img src="/icons/loading.svg" />
              </div>
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
              SignUp 🚀
            </div>
        </div>


      </div>
      
    </div>
  </div>
</template>


<script setup>
import { ref, reactive, computed, onMounted, onBeforeMount } from 'vue';
import { useTrayStates } from '@/stores/TrayStates';
import { useProjectStore } from '@/stores/projects';
import { AuthService } from "@/../bindings/clustta/services";
import { useNotificationStore } from '@/stores/notifications';
import { useUserStore } from '@/stores/users';
import { useIconStore } from '@/stores/icons';
import { useThemeStore } from '@/stores/theme';
import { useDesktopModalStore } from '@/stores/desktopModals';
import utils from "@/services/utils";

import { SettingsService } from '@/../bindings/clustta/services/index';

const trayStates = useTrayStates();
const projectStore = useProjectStore();
const userStore = useUserStore();
const notificationStore = useNotificationStore();
const iconStore = useIconStore();
const themeStore = useThemeStore();
const modals = useDesktopModalStore();

// refs
const error = ref('');
const isAwaitingResponse = ref(false);
const isPasswordVisible = ref(false)

// components
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';

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

// emits
const emit = defineEmits(['toggle-login']);

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

const eulaAccepted = ref(false);
const projectDirectoryExists = ref(false);

const showEula = async () => {

  projectDirectoryExists.value = await SettingsService.GetProjectDirectory();
  projectStore.projectsLoaded = !projectDirectoryExists.value;
  eulaAccepted.value = await SettingsService.GetEulaAccepted();

  if(eulaAccepted.value) return
	modals.setModalVisibility('eulaModal', true);

};

const toggleLogin = () => {
  emit('toggle-login')
};

const handleLogin = async () => {
  isAwaitingResponse.value = true;
  await AuthService.Login(loginForm.email, loginForm.password)
    .then(async (data) => {
      userStore.user = data.user;
      userStore.isUserAuthenticated = true;

      // await showEula();

      await themeStore.initializeTheme();
      await projectStore.loadStudios();

      projectDirectoryExists.value = await SettingsService.GetProjectDirectory();

      if(projectDirectoryExists.value){
        await projectStore.loadProjects();
        trayStates.refreshData();
      } else {
        setDirectories();
      }

    })
    .catch((error) => {
      console.log(error)
      isAwaitingResponse.value = false;
      notificationStore.errorNotification("Error Loggin In", error)
    });
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

</style>

