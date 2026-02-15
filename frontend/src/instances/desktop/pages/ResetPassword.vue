<template>
  <div class="page-root reset-password-page-root">

    <!-- responsive root -->
    <div class="auth-root">

      <!-- header -->
      <div class="header-container">
        <ClusttaLogo :colored="true" :inverted="true" />
        <div class="auth-header">
          {{ $t('auth.resetPassword.title') }}
        </div>
        <div class="auth-subheader">
          {{ $t('auth.resetPassword.subheader') }}
        </div>
      </div>

      <div class="auth-container">

        <!-- form container -->
        <div class="auth-form-container">

          <!-- actual-form -->
          <form @submit.prevent="handleResetPassword" class="auth-form">

            <!-- email -->
            <FormInput
              v-model="resetForm.email"
              :placeholder="$t('auth.resetPassword.emailPlaceholder')"
              needsValidation
              :error="errors.email"
              :valid="emailValid"
              :showValidation="!!resetForm.email"
              @input="validateEmail"
            />

            <!-- submit button -->
            <button type="submit" class="submit-button display-font" :class="{ 'button-inactive': !isResetFormFilled }">
              <div v-if="!isAwaitingResponse">
                {{ $t('auth.resetPassword.resetButton') }}
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
              {{ $t('auth.resetPassword.backToLogin') }}
            </div>
        </div>

      </div>
      
    </div>
  </div>
</template>


<script setup>
import { ref, reactive, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { AuthService } from "@/services";
import { useNotificationStore } from '@/stores/notifications';
import { useIconStore } from '@/stores/icons';

const router = useRouter();
const { t } = useI18n();
const notificationStore = useNotificationStore();
const iconStore = useIconStore();

// refs
const error = ref('');
const isAwaitingResponse = ref(false);
const emailValid = ref(false);

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';
import FormInput from '@/instances/desktop/components/FormInput.vue';

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

// methods
const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

const validateEmail = () => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  emailValid.value = emailRegex.test(resetForm.email);
  
  if (resetForm.email && !emailValid.value) {
    errors.email = t('auth.resetPassword.invalidEmail');
  } else {
    errors.email = '';
  }
};

const backToLogin = () => {
  router.push('/auth/login');
};

const handleResetPassword = async () => {
  if (!emailValid.value) {
    error.value = t('auth.resetPassword.invalidEmail');
    return;
  }

  isAwaitingResponse.value = true;
  error.value = '';
  
  await AuthService.ResetPassword(resetForm.email)
    .then(() => {
      notificationStore.addNotification(
        t('auth.resetPassword.passwordResetSent'), 
        t('auth.resetPassword.checkEmailForInstructions'), 
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
      notificationStore.errorNotification(t('auth.resetPassword.errorTitle'), errorMessage || t('auth.resetPassword.resetFailed'));
      error.value = errorMessage || t('auth.resetPassword.resetFailed');
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
