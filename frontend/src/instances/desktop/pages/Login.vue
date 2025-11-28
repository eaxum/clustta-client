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
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

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
const emit = defineEmits(['toggle-login', 'show-verification', 'show-reset-password']);

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
  emit('toggle-login')
};

const showResetPassword = () => {
  emit('show-reset-password')
};

const handleLogin = async () => {
  isAwaitingResponse.value = true;
  error.value = '';
  
  await AuthService.Login(loginForm.email, loginForm.password)
    .then(async (data) => {
      userStore.user = data.user;
      userStore.isUserAuthenticated = true;

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
      console.log(error);
      isAwaitingResponse.value = false;
      
      // Check if error indicates user needs verification
      const errorMessage = error.message || error.toString();
      const isUnverifiedUser = errorMessage.toLowerCase().includes('please verify your email before logging in') || 
                               errorMessage.toLowerCase().includes('account not verified');
      
      if (isUnverifiedUser) {
        notificationStore.addNotification("Verification Required", "Please check your email for a verification code.", "info");
        emit('show-verification', { email: loginForm.email, password: loginForm.password });
      } else {
        // Handle other login errors normally
        // console.log(error)
        // error.message = errorMessage;
        notificationStore.errorNotification("Error Logging In", 'Please check your credentials and try again');
      }
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

