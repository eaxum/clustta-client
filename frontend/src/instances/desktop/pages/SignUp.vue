<template>
  <div class="page-root sign-up-page-root">

    <!-- responsive root -->
    <div class="auth-root">

      <!-- header -->
      <div class="header-container">
        <ClusttaLogo :colored="true" :inverted="true" :boldText="true" />
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
                <div v-if="registerForm.username" class="form-input-icon">
                  <img v-if="errors.username || !usernameValid" class="alert-icons" :src="getAppIcon('alert')" />
                  <img v-else-if="checkingUsernameAvailability" class="alert-icons" src="/icons/loading.svg"
                    :class="{ 'loading-icon': checkingUsernameAvailability }" />
                  <img v-else class="alert-icons" :src="getAppIcon('circle-check')" />
                </div>
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
                <div v-if="registerForm.email" class="form-input-icon">
                  <img v-if="errors.email || !emailValid" class="alert-icons" :src="getAppIcon('alert')" />
                  <img v-else-if="checkingEmailAvailability" class="alert-icons" src="/icons/loading.svg"
                    :class="{ 'loading-icon': checkingEmailAvailability }" />
                  <img v-else class="alert-icons" :src="getAppIcon('circle-check')" />
                </div>
              </div>
              <span v-if="errors.email" class="error-message">{{ errors.email }}</span>
            </div>

            <!-- password -->
            <div class="form-group">
              <div class="compound-form-input">
                <input class="form-input-mini" placeholder="Password" v-model="registerForm.password"
                  :type="isPasswordVisible ? 'text' : 'password'" required :class="{ 'error': errors.password }" autocomplete="new-password" name="new-password">
                <div v-if="registerForm.password" class="form-input-icon" @mousedown="showPassword"
                  @mouseup="hidePassword" @mouseleave="hidePassword">
                  <img class="alert-icons" :src="isPasswordVisible ? getAppIcon('eye-cancel') : getAppIcon('eye')" />
                </div>
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
import { useTrayStates } from '@/stores/TrayStates';
import { useProjectStore } from '@/stores/projects';
import { AuthService } from "@/../bindings/clustta/services";
import { useNotificationStore } from '@/stores/notifications';
import { useUserStore } from '@/stores/users';
import { useIconStore } from '@/stores/icons';
import { useDesktopModalStore } from '@/stores/desktopModals';

// components
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';

const isAwaitingResponse = ref(false);
const trayStates = useTrayStates();
const projectStore = useProjectStore();
const userStore = useUserStore();
const notificationStore = useNotificationStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();

const emit = defineEmits(['toggle-login', 'signup-success']);

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
  emit('toggle-login')
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

const escapeRegexChars = (string) => {
  return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
};

const checkUsername = async () => {
  if (!registerForm.username) return
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
  try {
    if (registerForm.password !== registerForm.confirm_password) {
      error.value = 'Passwords do not match'
      isAwaitingResponse.value = false;
      return
    }

    await AuthService.Register(registerForm.first_name, registerForm.last_name, registerForm.username, registerForm.email, registerForm.password, registerForm.confirm_password)
    .then(async (data) => {
      // Registration successful, emit signup-success to trigger verification flow
      notificationStore.addNotification("Registration Successful", "Please check your email for a verification code.", "success");
      emit('signup-success', registerForm.email);
      isAwaitingResponse.value = false;
    }).catch((error) => {
      console.log(error);
      isAwaitingResponse.value = false;
      notificationStore.errorNotification("Registration Failed", error.message || "Registration failed. Please try again.");
      return
    })

  } catch (err) {
    error.value = err.response?.data?.message || 'Registration failed'
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

</style>

