<template>
  <div class="page-root sign-up-page-root">

    <!-- responsive root -->
    <div class="auth-root">

      <!-- header -->
      <div class="header-container">
        <ClusttaLogo :colored="true" :inverted="true" />
        <div class="auth-header">
          Sign up for Clustta
        </div>
      </div>

      <div class="auth-container">
        <!-- form container -->
        <div class="auth-form-container">
          <!-- studio server toggle -->
          <div class="horizontal-flex studio-toggle-row">
            <ActionButton :isInactive="true" :icon="getAppIcon('two-drives')" :label="'Private Server'" />
            <ToggleSwitch @click="toggleStudioSignup" :switchValueProp="showStudioSignup" />
          </div>
          <!-- studio URL input (shown when toggled) -->
          <div v-if="showStudioSignup" class="studio-url-container">
            <FormInput
              v-model="studioUrl"
              placeholder="Studio URL"
              :error="studioUrlError"
              :info="!studioUrlError ? 'Enter the URL of your studio server to create an account there.' : ''"
              @input="validateStudioUrl"
            />
          </div>
          <!-- actual-form -->
          <form @submit.prevent="handleRegister" class="auth-form" autocomplete="off">
            <!-- first and last names -->
            <div class="form-row">
              <FormInput v-model="registerForm.first_name" placeholder="First Name" :error="errors.first_name" />
              <FormInput v-model="registerForm.last_name" placeholder="Last Name" />
            </div>
            <!-- username -->
            <FormInput
              v-model="registerForm.username"
              placeholder="Username"
              needsValidation
              :error="errors.username || (!usernameValid && registerForm.username ? 'Username must be at least 3 characters long and can only contain letters, numbers, and underscores (_).' : '')"
              :loading="checkingUsernameAvailability"
              :valid="usernameValid && !isUsernameTaken"
              :showValidation="!!registerForm.username"
              @input="checkUsername"
            />
            <!-- email -->
            <FormInput
              v-model="registerForm.email"
              placeholder="Email address"
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
              placeholder="Password"
              isSecret
              :error="passwordValidation"
            />
            <!-- confirm password -->
            <FormInput
              v-model="registerForm.confirm_password"
              placeholder="Confirm password"
              isSecret
              :error="!passwordsMatch && registerForm.confirm_password ? errors.confirm_password : ''"
            />
            <!-- submit button -->
            <button type="submit" class="submit-button display-font" :class="{ 'button-inactive': !isRegisterFormFilled }">
              <div v-if="!isAwaitingResponse">
                Sign Up
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


        </div>

        <!-- toggle -->
        <div class="additional-actions">
          <div @click="toggleLogin" class="login-toggle">
            Have an account?&nbsp;<span class="bold">Login</span>
          </div>
        </div>


      </div>
      
    </div>
  </div>
</template>

<script setup>

// imports
import { ref, reactive, computed, onMounted, onBeforeMount } from 'vue'
import { useRouter } from 'vue-router'
import { useTrayStates } from '@/stores/TrayStates';
import { useProjectStore } from '@/stores/projects';
import { AuthService } from "@/services";
import { useNotificationStore } from '@/stores/notifications';
import { useUserStore } from '@/stores/users';
import { useIconStore } from '@/stores/icons';
import { useDesktopModalStore } from '@/stores/desktopModals';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';
import FormInput from '@/instances/desktop/components/FormInput.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

const router = useRouter();
const trayStates = useTrayStates();
const projectStore = useProjectStore();
const userStore = useUserStore();
const notificationStore = useNotificationStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();

// refs
const checkingEmailAvailability = ref(false);
const checkingUsernameAvailability = ref(false);
const delay = ref(0)
const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const error = ref('');
const isAwaitingResponse = ref(false);
const isEmailTaken = ref(false);
const isUsernameTaken = ref(false);
const showStudioSignup = ref(false);
const studioUrl = ref('');
const studioUrlError = ref('');
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
      errorMessage: 'Password cannot contain your Email or Username'
    },
    {
      value: email.split('@')[0],
      errorMessage: 'Password cannot contain your Email or Username'
    },
    {
      value: firstName,
      errorMessage: 'Password cannot contain your First Name'
    },
    {
      value: lastName,
      errorMessage: 'Password cannot contain your Last Name'
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
      errorMessage: 'Password must be at least 8 characters long'
    },
    {
      regex: /[A-Z]/,
      errorMessage: 'Password must include at least one uppercase letter (A-Z)'
    },
    {
      regex: /[a-z]/,
      errorMessage: 'Password must include at least one lowercase letter (a-z)'
    },
    {
      regex: /\d/,
      errorMessage: 'Password must include at least one number'
    },
    {
      regex: /[@$!%*?&]/,
      errorMessage: 'Password must include at least one special character (@, $, !, %, *, ?, &)'
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
  errors.confirm_password = passwordsMatch ? '' : 'Passwords do not match';
  return passwordsMatch && registerForm.password.length
});
const isRegisterFormFilled = computed(() => {
  return detailsInputed.value && credentialsValid.value && !passwordValidation.value && passwordsMatch.value
});

// methods

// Checks if the email is already registered.
const checkEmail = async () => {
  if (!registerForm.email || !emailValid.value) return
  
  // For studio signup, skip live availability check (validated on submit)
  if (showStudioSignup.value) {
    isEmailTaken.value = false;
    errors.email = '';
    return;
  }
  
  checkingEmailAvailability.value = true;

  try {
    const emailExist = await AuthService.CheckEmailExists(registerForm.email)
    if (emailExist) {
      isEmailTaken.value = true;
      errors.email = 'Email is already registered'
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
  
  // For studio signup, skip live availability check (validated on submit)
  if (showStudioSignup.value) {
    isUsernameTaken.value = false;
    errors.username = '';
    return;
  }
  
  checkingUsernameAvailability.value = true;

  try {
    const usernameExist = await AuthService.CheckUsernameExists(registerForm.username.toLowerCase())
    if (usernameExist) {
      errors.username = 'Username is already taken'
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

// Handles the registration form submission.
const handleRegister = async () => {
  isAwaitingResponse.value = true;
  error.value = '';
  
  try {
    if (registerForm.password !== registerForm.confirm_password) {
      error.value = 'Passwords do not match';
      isAwaitingResponse.value = false;
      return;
    }

    // Determine if this is a studio registration
    const isStudioSignup = showStudioSignup.value && studioUrl.value.trim();
    const normalizedStudioUrl = isStudioSignup ? normalizeStudioUrl(studioUrl.value) : '';

    if (isStudioSignup) {
      // Register against studio server
      await AuthService.RegisterWithHost(
        registerForm.first_name,
        registerForm.last_name,
        registerForm.username,
        registerForm.email,
        registerForm.password,
        registerForm.confirm_password,
        normalizedStudioUrl
      );
      
      // Studio registration is auto-activated, go directly to login
      notificationStore.addNotification(
        "Registration Successful",
        `Account created on ${normalizedStudioUrl}. You can now login.`,
        "success"
      );
      router.push('/auth/login');
    } else {
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
        "Registration Successful",
        "Please check your email for a verification code.",
        "success"
      );
      userStore.setPendingVerification(registerForm.email, registerForm.password);
      router.push('/auth/verify-email');
    }
  } catch (err) {
    console.log(err);
    const errorMessage = err.message || err.response?.data?.message || 'Registration failed';
    error.value = errorMessage;
    notificationStore.errorNotification("Registration Failed", errorMessage);
  } finally {
    isAwaitingResponse.value = false;
  }
};

// Normalizes the studio URL by ensuring it has a protocol and no trailing slash.
const normalizeStudioUrl = (url) => {
  if (!url) return '';
  let normalized = url.trim();
  normalized = normalized.replace(/\/+$/, '');
  if (!normalized.startsWith('http://') && !normalized.startsWith('https://')) {
    normalized = 'https://' + normalized;
  }
  return normalized;
};

// Navigates to the login page.
const toggleLogin = () => {
  router.push('/auth/login')
};

// Toggles the studio signup mode.
const toggleStudioSignup = () => {
  showStudioSignup.value = !showStudioSignup.value;
  if (!showStudioSignup.value) {
    studioUrl.value = '';
    studioUrlError.value = '';
  }
  isEmailTaken.value = false;
  isUsernameTaken.value = false;
  errors.email = '';
  errors.username = '';
};

// Validates the studio URL format.
const validateStudioUrl = () => {
  if (!studioUrl.value) {
    studioUrlError.value = '';
    return;
  }
  
  const urlPattern = /^https?:\/\/[a-zA-Z0-9][-a-zA-Z0-9]*(\.[a-zA-Z0-9][-a-zA-Z0-9]*)+(:\d+)?(\/.*)?$/;
  
  if (!studioUrl.value.startsWith('http://') && !studioUrl.value.startsWith('https://')) {
    studioUrlError.value = 'URL must start with http:// or https://';
  } else if (!urlPattern.test(studioUrl.value)) {
    studioUrlError.value = 'Please enter a valid URL';
  } else {
    studioUrlError.value = '';
  }
};

onMounted(() => {
  
});

onBeforeMount(async () => {
  
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

.studio-toggle-row {
  justify-content: space-between;
  padding: 0.5rem 0;
}

.studio-url-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.studio-url-hint {
  font-size: 0.75rem;
  color: var(--white-60);
  text-align: center;
}

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
  color: var(--white);
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

</style>

