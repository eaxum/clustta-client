<template>
  <div class="page-root reset-password-page-root">

    <!-- responsive root -->
    <div class="auth-root">

      <!-- header -->
      <div class="header-container">
        <ClusttaLogo :colored="true" :inverted="true" :boldText="true" />
        <div class="auth-header">
          Reset Password
        </div>
        <div class="auth-subheader">
          Enter your email address and we'll send you instructions to reset your password.
        </div>
      </div>

      <div class="auth-container">

        <!-- form container -->
        <div class="auth-form-container">

          <!-- actual-form -->
          <form @submit.prevent="handleResetPassword" class="auth-form">

            <!-- email -->
            <div class="form-group">
              <div class="compound-form-input">
                <input autocomplete="off" class="form-input-mini" placeholder="Email address" v-model="resetForm.email" type="email"
                  required @input="validateEmail" :class="{ 'error': errors.email && resetForm.email }" />
                <ActionButton 
                  v-if="resetForm.email && !emailValid"
                  :icon="getAppIcon('alert')"
                  :showLabel="false"
                  :useAlert="true"
                  :isInactive="true"
                  :noFilter="true"
                />
                <ActionButton 
                  v-else-if="resetForm.email && emailValid"
                  :icon="getAppIcon('circle-check')"
                  :showLabel="false"
                  :useGo="true"
                  :noFilter="true"
                  :isInactive="true"
                />
              </div>
              <span v-if="errors.email && resetForm.email" class="error-message">{{ errors.email }}</span>
            </div>

            <!-- submit button -->
            <button type="submit" class="submit-button display-font" :class="{ 'button-inactive': !isResetFormFilled }">
              <div v-if="!isAwaitingResponse">
                Reset Password
              </div>
              <ActionButton
                v-else
                :icon="getAppIcon('loading')"
                :isLoading="true"
                :showLabel="false"
                :isInactive="true"
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
            <div class="bold" >
              Back to Login
            </div>
        </div>

      </div>
      
    </div>
  </div>
</template>


<script setup>
import { ref, reactive, computed } from 'vue';
import { AuthService } from "@/services";
import { useNotificationStore } from '@/stores/notifications';
import { useIconStore } from '@/stores/icons';

const notificationStore = useNotificationStore();
const iconStore = useIconStore();

// refs
const error = ref('');
const isAwaitingResponse = ref(false);
const emailValid = ref(false);

// components
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// vars
const resetForm = reactive({
  email: '',
});

const errors = reactive({
  email: '',
});

// computed props
const isResetFormFilled = computed(() => {
  return resetForm.email && emailValid.value && !errors.email
});

// emits
const emit = defineEmits(['back-to-login']);

// methods
const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

const validateEmail = () => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  emailValid.value = emailRegex.test(resetForm.email);
  
  if (resetForm.email && !emailValid.value) {
    errors.email = 'Please enter a valid email address';
  } else {
    errors.email = '';
  }
};

const backToLogin = () => {
  emit('back-to-login')
};

const handleResetPassword = async () => {
  if (!emailValid.value) {
    error.value = 'Please enter a valid email address';
    return;
  }

  isAwaitingResponse.value = true;
  error.value = '';
  
  await AuthService.ResetPassword(resetForm.email)
    .then(() => {
      notificationStore.addNotification(
        "Password Reset Sent", 
        "Please check your email for password reset instructions.", 
        "success"
      );
      // Reset form
      resetForm.email = '';
      emailValid.value = false;
      // Go back to login
      setTimeout(() => {
        backToLogin();
      }, 1500);
    })
    .catch((error) => {
      console.log(error);
      isAwaitingResponse.value = false;
      const errorMessage = error.message || error.toString();
      notificationStore.errorNotification("Error", errorMessage || 'Failed to send password reset email. Please try again.');
      error.value = errorMessage || 'Failed to send password reset email. Please try again.';
    })
    .finally(() => {
      isAwaitingResponse.value = false;
    });
};
</script>


<style scoped>
@import "@/assets/desktop.css";

.auth-subheader {
  color: var(--silver);
  font-size: 14px;
  text-align: center;
  margin-top: 0.5rem;
  opacity: 0.8;
}
</style>
