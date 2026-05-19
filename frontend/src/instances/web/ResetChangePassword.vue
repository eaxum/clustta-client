<template>
  <div class="page-root reset-change-password-page-root">

    <!-- responsive root -->
    <div class="auth-root">

      <!-- header -->
      <div class="header-container">
        <ClusttaLogo :colored="true" :inverted="true" />
        <div class="auth-header">
          Change Password
        </div>
        <div class="auth-subheader">
          Enter your new password below.
        </div>
      </div>

      <div class="auth-container">

        <!-- form container -->
        <div class="auth-form-container">

          <!-- actual-form -->
          <form @submit.prevent="handleResetPassword" class="auth-form">

            <!-- new password -->
            <FormInput
              v-model="resetForm.new_password"
              placeholder="New Password"
              isSecret
              :error="passwordValidation"
            />

            <!-- confirm password -->
            <FormInput
              v-model="resetForm.confirm_password"
              placeholder="Confirm New Password"
              isSecret
              needsValidation
              :error="!passwordsMatch && resetForm.confirm_password ? 'Passwords do not match' : ''"
              :valid="passwordsMatch && !!resetForm.confirm_password"
              :showValidation="!!resetForm.confirm_password"
            />

            <!-- submit button -->
            <button type="submit" class="submit-button display-font" :class="{ 'button-inactive': !isFormValid }">
              <div v-if="!isAwaitingResponse">
                Change Password
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
        <div @click="backToLogin" class="toggle-container">
            <div class="bold">
              Back to Login
            </div>
        </div>

      </div>
      
    </div>
  </div>
</template>


<script setup>
import { ref, reactive, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { AuthService } from "@/services/adapters/authservice.js";
import { useNotificationStore } from '@/stores/notifications';
import { useIconStore } from '@/stores/icons';

const route = useRoute();
const router = useRouter();
const notificationStore = useNotificationStore();
const iconStore = useIconStore();

// refs
const error = ref('');
const isAwaitingResponse = ref(false);

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';
import FormInput from '@/instances/desktop/components/FormInput.vue';

// vars
const resetForm = reactive({
  new_password: '',
  confirm_password: '',
});

// computed props
const passwordsMatch = computed(() => {
  return resetForm.new_password === resetForm.confirm_password;
});

const passwordValidation = computed(() => {
  const password = resetForm.new_password;

  if (!password) {
    return null;
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
  ];

  for (const rule of validationRules) {
    if (!rule.regex.test(password)) {
      return rule.errorMessage;
    }
  }

  return null;
});

const isFormValid = computed(() => {
  return resetForm.new_password && 
         resetForm.confirm_password && 
         !passwordValidation.value && 
         passwordsMatch.value;
});

// methods

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Navigates back to the login page.
const backToLogin = () => {
  router.push('/');
};

const handleResetPassword = async () => {
  if (!isFormValid.value) {
    return;
  }

  // Get email and token from query params
  const email = route.query.email;
  const token = route.query.token;

  if (!email || !token) {
    error.value = 'Invalid password reset link. Please request a new one.';
    return;
  }

  isAwaitingResponse.value = true;
  error.value = '';
  
  try {
    await AuthService.ResetChangePassword(
      email,
      token,
      resetForm.new_password,
      resetForm.confirm_password
    );

    notificationStore.addNotification(
      "Password Changed", 
      "Your password has been changed successfully. Please login with your new password.", 
      "success"
    );

    // Reset form
    resetForm.new_password = '';
    resetForm.confirm_password = '';

    // Redirect to login after short delay
    setTimeout(() => {
      router.push('/');
    }, 1500);

  } catch (err) {
    console.error(err);
    const errorMessage = err.message || err.toString();
    error.value = errorMessage || 'Failed to change password. Please try again.';
    notificationStore.errorNotification("Error", errorMessage || 'Failed to change password. Please try again.');
  } finally {
    isAwaitingResponse.value = false;
  }
};

onMounted(() => {
  // Check if we have the required query params
  if (!route.query.email || !route.query.token) {
    error.value = 'Invalid password reset link. Please request a new one.';
  }
});
</script>


<style scoped>
@import "@/assets/desktop.css";

.reset-change-password-page-root {
  display: flex;
  box-sizing: border-box;
  width: 100%;
  align-items: center;
  height: 100%;
  flex-direction: column;
  background-color: var(--surface-1);
  overflow: hidden;
  overflow-y: auto;
}

.auth-subheader {
  color: var(--text-muted);
  font-size: 14px;
  text-align: center;
  margin-top: 0.5rem;
  opacity: 0.8;
}

.auth-root {
  display: flex;
  flex-direction: row;
  align-items: center;
  height: 100%;
  min-height: min-content;
  max-width: min-content;
  width: 100%;
  justify-content: space-around;
  gap: 2rem;
  box-sizing: border-box;
}

.header-container {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 400px;
  box-sizing: border-box;
  overflow: hidden;
  height: max-content;
  min-width: 300px;
}

.auth-container {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  width: min-content;
  min-width: 400px;
  box-sizing: border-box;
  overflow: hidden;
}

.auth-header {
  font-family: 'Bricolage Grotesque', sans-serif;
  font-size: 4rem;
  font-weight: 600;
  line-height: 90%;
  width: max-content;
  text-align: left;
  color: var(--text);
  height: max-content;
  min-width: 330px;
  width: 100%;
  text-wrap: wrap;
}

.auth-form-container {
  padding: 1rem;
  box-sizing: border-box;
  width: 100%;
  max-width: 480px;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0px;
  width: 100%;
  align-items: center;
  box-sizing: border-box;
  overflow: hidden;
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
  background-color: var(--bg);
  align-items: center;
}

.form-input-mini {
  color: var(--text);
  box-sizing: border-box;
  border: 0px;
  border-radius: 4px;
  font-size: 1rem;
  width: 100%;
  height: 100%;
  padding: 0.75rem;
  background-color: var(--bg);
  font-family: 'Inter', sans-serif;
  box-sizing: border-box;
  font-size: 16px;
  border-radius: 12px;
  padding: 10px;
  border: 0px;
  border-style: solid;
  outline: none;
  background-color: var(--bg);
  color: var(--text);
}

.error-message {
  margin-top: .5rem;
  color: #dc2626;
  text-align: center;
  font-size: 0.875rem;
  font-weight: 300;
  width: 100%;
  display: flex;
  padding: .3rem;
}

.submit-button {
  font-size: x-large;
  background-color: var(--grape);
  color: var(--text);
  color: white;
  padding: 0.75rem;
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  border: none;
  border-radius: var(--large-radius);
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
}

.submit-button:hover {
  background-color: var(--bright-grape);
}

.button-inactive {
  opacity: .5;
  cursor: not-allowed;
}

.button-inactive:hover {
  background-color: var(--grape);
}

input.error {
  border-color: #dc3545;
}

.toggle-container {
  width: 90%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text);
  cursor: pointer;
  gap: 1rem;
}

@media (max-width: 768px) {
  .auth-form-container {
    padding: 1.5rem;
  }

  .auth-root {
    flex-direction: column;
    padding: .5rem;
    gap: 0px;
    min-width: 100%;
  }
  
  .auth-container {
    width: 100%;
    flex: 1;
    min-width: 300px;
    padding: 0 2rem;
    margin-bottom: 2rem;
  }
  
  .auth-header {
    font-size: 3rem;
    text-wrap: wrap;
    width: 100%;
  }
  
  .header-container {
    justify-content: flex-start;
    width: 100%;
    height: max-content;
    max-width: 500px;
    padding: 1.5rem;
    padding-bottom: .5rem;
  }
}
</style>
