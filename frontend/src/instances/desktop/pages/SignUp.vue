<template>
  <div class="page-root sign-up-page-root">
    
    <!-- responsive root -->
    <div class="auth-root">

      <!-- header -->
      <div class="header-container">
        <ClusttaLogo :colored="true" :inverted="true" />
        <div class="auth-header">
          {{ $t('auth.signUp.title') }}
        </div>
      </div>

      <div class="auth-container">
        <AuthLoader v-if="isInitializing" :status="loadingStatus" />

        <!-- form container -->
        <div v-if="!isInitializing" class="auth-form-container">
          <!-- actual-form -->
          <form @submit.prevent="handleRegister" class="auth-form" autocomplete="off">
            <!-- first and last names -->
            <div class="form-row">
              <FormInput v-model="registerForm.first_name" :placeholder="$t('auth.signUp.firstName')" :error="errors.first_name" />
              <FormInput v-model="registerForm.last_name" :placeholder="$t('auth.signUp.lastName')" />
            </div>
            <!-- username -->
            <FormInput
              v-model="registerForm.username"
              :placeholder="$t('auth.signUp.username')"
              needsValidation
              :error="errors.username || (!usernameValid && registerForm.username ? $t('auth.signUp.usernameValidation') : '')"
              :loading="checkingUsernameAvailability"
              :valid="usernameValid && !isUsernameTaken"
              :showValidation="!!registerForm.username"
              @input="checkUsername"
            />
            <!-- email -->
            <FormInput
              v-model="registerForm.email"
              :placeholder="$t('auth.signUp.emailAddress')"
              needsValidation
              :error="errors.email"
              :loading="checkingEmailAvailability"
              :valid="emailValid && !isEmailTaken"
              :showValidation="!!registerForm.email"
              @input="checkEmail"
            />
            <!-- password -->
            <FormInput
              v-model="registerForm.password"
              :placeholder="$t('auth.signUp.password')"
              isSecret
              :error="passwordValidation"
            />
            <!-- confirm password -->
            <FormInput
              v-model="registerForm.confirm_password"
              :placeholder="$t('auth.signUp.confirmPassword')"
              isSecret
              :error="!passwordsMatch && registerForm.confirm_password ? errors.confirm_password : ''"
            />
            <!-- submit button -->
            <button type="submit" class="submit-button display-font" :class="{ 'button-inactive': !isRegisterFormFilled }">
              <div v-if="!isAwaitingResponse">
                {{ $t('auth.signUp.signUpButton') }}
              </div>
              <ActionButton
                v-else
                :icon="getAppIcon('loading')"
                :isLoading="true"
                :showLabel="false"
              />
            </button>
          </form>

          <!-- form error -->
          <div v-if="error" class="error-message">
            {{ error }}
          </div>

          <SSOLogin v-if="!isAwaitingResponse" mode="signup" @success="handleSSOSuccess" @error="handleSSOError" />

        </div>

        <!-- toggle -->
        <div v-if="!isInitializing" class="additional-actions">
          <div v-if="!platformStore.isWeb && !accountStore.onboardingIntent" @click="goToStudioSignUp" class="login-toggle">{{ $t('auth.signUp.signUpToStudio') }}</div>

          <div @click="toggleLogin" class="login-toggle">
            {{ $t('auth.signUp.haveAccount') }}&nbsp;<span class="bold">{{ $t('auth.signUp.loginLink') }}</span>
          </div>
        </div>

        <!-- legal agreement -->
        <div v-if="!isInitializing" class="legal-agreement">
          <p>{{ $t('auth.signUp.legalPrefix') }} <span class="legal-link" @click="openPrivacyPolicy">{{ $t('auth.signUp.privacyPolicy') }} <ActionButton :icon="getAppIcon('square-arrow-right-up')" :allowDeactivate="true" :isMini="true" /></span> {{ $t('auth.signUp.legalMiddle') }} <span class="legal-link" @click="openTermsOfService">{{ $t('auth.signUp.termsOfService') }} <ActionButton :icon="getAppIcon('square-arrow-right-up')" :allowDeactivate="true" :isMini="true" /></span>.</p>
        </div>

      </div>
      
    </div>
  </div>
</template>

<script setup>

// imports
import { ref, reactive, computed, onMounted, onBeforeMount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Browser } from "@wailsio/runtime";

// services
import { AuthService, SettingsService } from "@/services";

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import AuthLoader from '@/instances/desktop/components/AuthLoader.vue';
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';
import FormInput from '@/instances/desktop/components/FormInput.vue';
import SSOLogin from '@/instances/desktop/components/SSOLogin.vue';

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
const trayStates = useTrayStates();
const userStore = useUserStore();

const router = useRouter();
const { t } = useI18n();

// refs
const checkingEmailAvailability = ref(false);
const checkingUsernameAvailability = ref(false);
const delay = ref(0)
const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const error = ref('');
const isAwaitingResponse = ref(false);
const isEmailTaken = ref(false);
const isInitializing = ref(false);
const isUsernameTaken = ref(false);
const loadingStatus = ref('');
const userNameRegex = /^[a-zA-Z0-9_]{3,}$/

const registerForm = reactive({
  first_name: '',
  last_name: '',
  username: '',
  email: '',
  password: '',
  confirm_password: ''
});

const errors = reactive({
  first_name: '',
  last_name: '',
  username: '',
  email: '',
  password: '',
  confirm_password: ''
});

// computed props
const passwordValidation = computed(() => {
  const password = registerForm.password
  const username = registerForm.username.toLowerCase()
  const email = registerForm.email.toLowerCase()
  const firstName = registerForm.first_name.toLowerCase()
  const lastName = registerForm.last_name.toLowerCase()

  if (!password) {
    return null
  }

  const lowerPassword = password.toLowerCase()

  const patterns = [
    {
      value: username,
      errorMessage: t('auth.signUp.passwordContainsEmailOrUsername')
    },
    {
      value: email.split('@')[0],
      errorMessage: t('auth.signUp.passwordContainsEmailOrUsername')
    },
    {
      value: firstName,
      errorMessage: t('auth.signUp.passwordContainsFirstName')
    },
    {
      value: lastName,
      errorMessage: t('auth.signUp.passwordContainsLastName')
    }
  ]

  for (const pattern of patterns) {
    if (!pattern.value) continue

    const escapedPattern = escapeRegexChars(pattern.value)
    const regex = new RegExp(escapedPattern, 'i')

    if (regex.test(lowerPassword)) {
      return pattern.errorMessage
    }
  }

  const validationRules = [
    {
      regex: /.{8,}/,
      errorMessage: t('auth.signUp.passwordMinLength')
    },
    {
      regex: /[A-Z]/,
      errorMessage: t('auth.signUp.passwordUppercase')
    },
    {
      regex: /[a-z]/,
      errorMessage: t('auth.signUp.passwordLowercase')
    },
    {
      regex: /\d/,
      errorMessage: t('auth.signUp.passwordNumber')
    },
    {
      regex: /[@$!%*?&]/,
      errorMessage: t('auth.signUp.passwordSpecialChar')
    }
  ]

  for (const rule of validationRules) {
    if (!rule.regex.test(password)) {
      return rule.errorMessage
    }
  }

  return null
})
const usernameValid = computed(() => { return userNameRegex.test(registerForm.username) });
const emailValid = computed(() => { return emailRegex.test(registerForm.email) });
const credentialsValid = computed(() => { return emailValid.value && !isEmailTaken.value && usernameValid.value && !isUsernameTaken.value });
const detailsInputed = computed(() => { return registerForm.first_name && registerForm.last_name && registerForm.username });
const passwordsMatch = computed(() => {
  const passwordsMatch = registerForm.password === registerForm.confirm_password
  errors.confirm_password = passwordsMatch ? '' : t('auth.signUp.passwordsDoNotMatch');
  return passwordsMatch && registerForm.password.length
});
const isRegisterFormFilled = computed(() => {
  return !!detailsInputed.value && !!credentialsValid.value && !passwordValidation.value && !!passwordsMatch.value
});

// methods

// Checks if the email is already registered.
const checkEmail = async () => {
  if (!registerForm.email || !emailValid.value) return
  
  checkingEmailAvailability.value = true;

  try {
    const emailExist = await AuthService.CheckEmailExists(registerForm.email)
    if (emailExist) {
      isEmailTaken.value = true;
      errors.email = t('auth.signUp.emailAlreadyRegistered')
    } else {
      isEmailTaken.value = false;
      errors.email = ''
    }
    checkingEmailAvailability.value = false;
  } catch (error) {
    errors.email = ''
    console.error('Error checking email:', error);
    checkingEmailAvailability.value = false;
  }
};

// Checks if the username is already taken.
const checkUsername = async () => {
  if (!registerForm.username) return
  
  checkingUsernameAvailability.value = true;

  try {
    const usernameExist = await AuthService.CheckUsernameExists(registerForm.username.toLowerCase())
    if (usernameExist) {
      errors.username = t('auth.signUp.usernameAlreadyTaken')
      isUsernameTaken.value = true;
    } else {
      errors.username = ''
      isUsernameTaken.value = false;
    }
    checkingUsernameAvailability.value = false;
  } catch (error) {
    errors.username = ''
    console.error('Error checking username:', error)
    checkingUsernameAvailability.value = false;
  }
};

// Escapes special regex characters in a string.
const escapeRegexChars = (string) => {
  return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
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
    } else if (savedIntent !== 'studio') {
      modals.setModalVisibility('dirOnboardModal', true);
    }

    markStoresInitialized();

    if (savedIntent === 'studio') {
      router.push('/auth/studio-setup');
    } else {
      router.push('/');
    }
  } catch (err) {
    console.log(err);
    isInitializing.value = false;
    loadingStatus.value = '';
    error.value = 'Failed to initialize after sign-up. Please try again.';
  }
};

// Navigates to the studio setup page for self-hosted registration.
const goToStudioSignUp = () => {
  router.push({ path: '/auth/studio-setup', query: { type: 'self-hosted' } });
};

// Handles the registration form submission.
const handleRegister = async () => {
  isAwaitingResponse.value = true;
  error.value = '';
  
  try {
    if (registerForm.password !== registerForm.confirm_password) {
      error.value = t('auth.signUp.passwordsDoNotMatch');
      isAwaitingResponse.value = false;
      return;
    }

    // Register against Clustta Cloud (requires email verification)
    await AuthService.Register(
      registerForm.first_name,
      registerForm.last_name,
      registerForm.username,
      registerForm.email,
      registerForm.password,
      registerForm.confirm_password
    );

    notificationStore.addNotification(
      t('auth.signUp.registrationSuccessful'),
      t('auth.signUp.checkEmailForCode'),
      'success'
    );
    userStore.setPendingVerification(registerForm.email, registerForm.password);
    router.push('/auth/verify-email');
  } catch (err) {
    console.log(err);
    const errorMessage = err.message || err.response?.data?.message || t('auth.signUp.registrationFailedDefault');
    error.value = errorMessage;
    notificationStore.errorNotification(t('auth.signUp.registrationFailed'), errorMessage);
  } finally {
    isAwaitingResponse.value = false;
  }
};

// Opens the privacy policy page in the browser.
const openPrivacyPolicy = () => {
  Browser.OpenURL('https://clustta.com/privacy-policy');
};

// Opens the terms of service page in the browser.
const openTermsOfService = () => {
  Browser.OpenURL('https://clustta.com/terms-of-service');
};

// Navigates to the login page.
const toggleLogin = () => {
  router.push('/auth/login')
};

onMounted(() => {
  
});

onBeforeMount(async () => {
  
})
</script>

<style scoped>
@import "@/assets/desktop.css";

.additional-actions {
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

.login-toggle {
  color: var(--text);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: .3rem;
  opacity: 0.6;
  cursor: pointer;
  transition: opacity 0.2s;
}

.login-toggle:hover {
  opacity: 1;
}

.legal-agreement {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 1rem 0.5rem;
  font-size: 12px;
  color: var(--text);
  font-weight: 300;
  gap: 0.25rem;
}

.legal-agreement p {
  margin: 0;
}

.legal-link {
  color: var(--text);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  text-decoration: underline;
}

.legal-link:hover {
  color: var(--blue);
}

</style>

