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

          <!-- actual-form -->
          <form @submit.prevent="handleRegister" class="auth-form" autocomplete="off">

            <!-- first and last names -->
            <div class="form-row">
              <div class="form-group">
                <input class="form-input input-short" placeholder="First Name" v-model="registerForm.first_name" type="text"
                  required :class="{ 'error': errors.first_name }" autocomplete="off" name="new-first-name" />
                <span v-if="errors.first_name" class="error-message">{{ errors.first_name }}</span>
              </div>
              <div class="form-group">
                <input class="form-input input-short" placeholder="Last Name" v-model="registerForm.last_name" type="text"
                  required autocomplete="off" name="new-last-name" />
              </div>
            </div>

            <!-- username -->
            <div class="form-group">
              <div class="compound-form-input">
                <input class="form-input-mini" placeholder="Username" v-model="registerForm.username" type="text"
                  required @input="checkUsername" autocomplete="off" name="new-username" />
                <ActionButton 
                  v-if="registerForm.username && (errors.username || !usernameValid)"
                  :icon="getAppIcon('alert')"
                  :showLabel="false"
                  :useAlert="true"
                  :isInactive="true"
                />
                <ActionButton 
                  v-else-if="registerForm.username && checkingUsernameAvailability"
                  :icon="getAppIcon('loading')"
                  :isLoading="true"
                  :showLabel="false"
                  :isInactive="true"
                />
                <ActionButton 
                  v-else-if="registerForm.username"
                  :icon="getAppIcon('circle-check')"
                  :showLabel="false"
                  :useGo="true"
                  :isInactive="true"
                />
              </div>
              <span v-if="errors.username" class="error-message">{{ errors.username }}</span>
              <span v-if="registerForm.username && !usernameValid" class="error-message"> Username must be at least 3
                characters long and can only contain letters, numbers, and underscores (_). </span>
            </div>

            <!-- email -->
            <div class="form-group">
              <div class="compound-form-input">
                <input class="form-input-mini" placeholder="Email address" v-model="registerForm.email" type="text"
                  required @input="checkEmail" autocomplete="off" name="new-email" />
                <ActionButton 
                  v-if="registerForm.email && (errors.email || !emailValid)"
                  :icon="getAppIcon('alert')"
                  :showLabel="false"
                  :useAlert="true"
                  :isInactive="true"
                  :noFilter="true"
                />
                <ActionButton 
                  v-else-if="registerForm.email && checkingEmailAvailability"
                  :icon="getAppIcon('loading')"
                  :isLoading="true"
                  :showLabel="false"
                  :isInactive="true"
                />
                <ActionButton 
                  v-else-if="registerForm.email"
                  :icon="getAppIcon('circle-check')"
                  :showLabel="false"
                  :useGo="true"
                  :noFilter="true"
                  :isInactive="true"
                />
              </div>
              <span v-if="errors.email" class="error-message">{{ errors.email }}</span>
            </div>

            <!-- password -->
            <div class="form-group">
              <div class="compound-form-input">
                <input class="form-input-mini" placeholder="Password" v-model="registerForm.password"
                  :type="isPasswordVisible ? 'text' : 'password'" required :class="{ 'error': errors.password }" autocomplete="new-password" name="new-password">
                <ActionButton 
                  v-if="registerForm.password"
                  v-tooltip="isPasswordVisible ? 'Hide Password' : 'Show Password'"
                  :icon="isPasswordVisible ? getAppIcon('eye-cancel') : getAppIcon('eye')"
                  @click="togglePasswordVisibility"
                  :showLabel="false"
                />
              </div>
              <span v-if="passwordValidation" class="error-message">{{ passwordValidation }}</span>
            </div>

            <!-- confirm password -->
            <div class="form-group">
              <input class="form-input input-short" placeholder="Confirm password" v-model="registerForm.confirm_password"
                type="password" required
                :class="{ 'error': errors.confirm_password && registerForm.confirm_password }" autocomplete="new-password" name="confirm-new-password" />
              <span v-if="!passwordsMatch && registerForm.confirm_password" class="error-message">{{
                errors.confirm_password }}</span>
            </div>

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

        <!-- studio server toggle -->
        <div class="horizontal-flex studio-toggle-row">
          <ActionButton :isInactive="true" :icon="getAppIcon('two-drives')" :label="'Private Server'" />
          <ToggleSwitch @click="toggleStudioSignup" :switchValueProp="showStudioSignup" />
        </div>

        <!-- studio URL input (shown when toggled) -->
        <div v-if="showStudioSignup" class="studio-url-container">
          <div class="form-group">
            <div class="compound-form-input">
              <input autocomplete="off" class="form-input-mini" placeholder="Studio URL (e.g., https://studio.mycompany.com)" v-model="studioUrl" type="text" />
            </div>
          </div>
          <div class="studio-url-hint">
            Enter the URL of your studio server to create an account there.
          </div>
        </div>

        <!-- toggle -->
        <div @click="toggleLogin" class="toggle-container">
            Have an account?
            <div class="bold" >
              Login 🚪
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
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

const router = useRouter();
const isAwaitingResponse = ref(false);
const trayStates = useTrayStates();
const projectStore = useProjectStore();
const userStore = useUserStore();
const notificationStore = useNotificationStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();

// refs
const isPasswordVisible = ref(false);
const error = ref('');
const checkingEmailAvailability = ref(false);
const checkingUsernameAvailability = ref(false);
const delay = ref(0)
const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const userNameRegex = /^[a-zA-Z0-9_]{3,}$/
const isEmailTaken = ref(false);
const isUsernameTaken = ref(false);
const showStudioSignup = ref(false);
const studioUrl = ref('');

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
const toggleLogin = () => {
  router.push('/auth/login')
};

const showPassword = () => {
  isPasswordVisible.value = true
};

const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

const hidePassword = () => {
  isPasswordVisible.value = false
};

const togglePasswordVisibility = () => {
  isPasswordVisible.value = !isPasswordVisible.value;
};

const escapeRegexChars = (string) => {
  return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
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

const toggleStudioSignup = () => {
  showStudioSignup.value = !showStudioSignup.value;
  if (!showStudioSignup.value) {
    studioUrl.value = '';
  }
  // Reset email/username validation when switching modes
  isEmailTaken.value = false;
  isUsernameTaken.value = false;
  errors.email = '';
  errors.username = '';
};

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

const showEula = async () => {
	modals.setModalVisibility('eulaModal', true);
};

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
  width: 90%;
  justify-content: space-between;
  padding: 0.5rem 0;
}

.studio-url-container {
  width: 90%;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.75rem;
  background: var(--white-05);
  border-radius: 8px;
  border: 1px solid var(--accent-color-30);
}

.studio-url-hint {
  font-size: 0.75rem;
  color: var(--white-60);
  text-align: center;
}

</style>

