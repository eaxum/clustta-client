<template>
  <div class="page-root login-page-root">

    <!-- responsive root -->
    <div class="auth-root">

      <!-- header -->
      <div class="header-container">
        <ClusttaLogo :colored="true" :inverted="true" :boldText="true" />
        <div class="auth-header">
          Verify Your Account
        </div>
        <div class="auth-subheader">
          Enter the 6-character verification code sent to your email
        </div>
      </div>

      <div class="auth-container">

        <!-- form container -->
        <div class="auth-form-container">

          <!-- actual-form -->
          <form @submit.prevent="handleVerification" class="auth-form">

            <!-- verification token -->
            <div class="form-group">
              <div class="token-input-container">
                <input 
                  v-for="(digit, index) in tokenDigits" 
                  :key="index"
                  :ref="el => tokenInputs[index] = el"
                  class="token-digit-input" 
                  type="text"
                  inputmode="numeric"
                  pattern="[0-9]*"
                  maxlength="1"
                  v-model="tokenDigits[index]"
                  @input="handleDigitInput(index, $event)"
                  @keydown="handleKeyDown(index, $event)"
                  @paste="handlePaste($event)"
                  required 
                />
              </div>
            </div>

            <!-- submit button -->
              <div v-if="isAwaitingResponse" class="loading-section">
                <div class="submit-button-icon loading-icon">
                  <img src="/icons/loading.svg" />
                </div>
              </div>

          </form>

          <!-- form error -->
          <div v-if="error" class="error-message">
            {{ error }}
          </div>

        </div>

        <!-- toggle -->
        <div class="verification-actions">
          
          
          <div @click="handleResendToken" class="resend-container" :class="{ 'disabled': isResendDisabled }">
              <span v-if="!isResendingToken">Resend Code</span>
              <div v-else class="submit-button-icon loading-icon">
                <img src="/icons/loading.svg" />
              </div>
          </div>
          

          <div @click="toggleLogin" class="toggle-container bold">
              Back to Login 🔐
          </div>

        </div>

      </div>
      
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue';
import { AuthService, SettingsService } from "@/../bindings/clustta/services";
import { useNotificationStore } from '@/stores/notifications';
import { useUserStore } from '@/stores/users';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useThemeStore } from '@/stores/theme';
import { useDesktopModalStore } from '@/stores/desktopModals';

const notificationStore = useNotificationStore();
const userStore = useUserStore();
const projectStore = useProjectStore();
const trayStates = useTrayStates();
const themeStore = useThemeStore();
const modals = useDesktopModalStore();

// refs
const error = ref('');
const isAwaitingResponse = ref(false);
const isResendingToken = ref(false);
const resendCooldown = ref(0);
const tokenDigits = ref(['', '', '', '', '', '']);
const tokenInputs = ref([]);

// computed
const isResendDisabled = computed(() => {
  return isResendingToken.value || resendCooldown.value > 0;
});

// components
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';


// computed props
const isVerificationFormFilled = computed(() => {
  return tokenDigits.value.every(digit => digit.length === 1);
});

const fullToken = computed(() => {
  return tokenDigits.value.join('');
});

// emits
const emit = defineEmits(['toggle-login', 'verification-success']);

// props
const props = defineProps({
  userEmail: {
    type: String,
    required: true
  },
  userPassword: {
    type: String,
    required: true
  }
});

// methods
const handleDigitInput = (index, event) => {
  const value = event.target.value;
  
  // Only allow numeric input
  if (!/^\d*$/.test(value)) {
    event.target.value = '';
    tokenDigits.value[index] = '';
    return;
  }
  
  tokenDigits.value[index] = value;
  
  // Auto-focus next input if current is filled
  if (value && index < 5) {
    const nextInput = tokenInputs.value[index + 1];
    if (nextInput) {
      nextInput.focus();
    }
  }
  
  // Auto-trigger verification if this is the last digit and form is complete
  if (value && index === 5 && isVerificationFormFilled.value) {
    // Small delay to ensure UI updates before verification
    setTimeout(() => {
      handleVerification();
    }, 100);
  }
};

const handleKeyDown = (index, event) => {
  // Handle backspace to move to previous input
  if (event.key === 'Backspace' && !tokenDigits.value[index] && index > 0) {
    const prevInput = tokenInputs.value[index - 1];
    if (prevInput) {
      prevInput.focus();
    }
  }
  
  // Handle enter key
  if (event.key === 'Enter') {
    handleVerification();
  }
};

const handlePaste = (event) => {
  event.preventDefault();
  const pastedText = (event.clipboardData || window.clipboardData).getData('text');
  const digits = pastedText.replace(/\D/g, '').slice(0, 6).split('');
  
  // Fill the digits array
  for (let i = 0; i < 6; i++) {
    tokenDigits.value[i] = digits[i] || '';
  }
  
  // Focus the appropriate input
  const lastFilledIndex = digits.length - 1;
  if (lastFilledIndex >= 0 && lastFilledIndex < 6) {
    const targetInput = tokenInputs.value[Math.min(lastFilledIndex + 1, 5)];
    if (targetInput) {
      targetInput.focus();
    }
  }
};

const clearTokenInputs = () => {
  tokenDigits.value = ['', '', '', '', '', ''];
  if (tokenInputs.value[0]) {
    tokenInputs.value[0].focus();
  }
};

const toggleLogin = () => {
  emit('toggle-login')
};

const handleVerification = async () => {
  if (!isVerificationFormFilled.value) {
    error.value = 'Please enter all 6 digits of your verification code';
    return;
  }

  isAwaitingResponse.value = true;
  error.value = '';

  try {
    
    const token = fullToken.value;
    await AuthService.VerifyOTP(props.userEmail, token);
    notificationStore.addNotification("Account Verified", "Your account has been successfully verified!", "success");
    
    // Auto-login the user after successful verification
    try {
      const loginData = await AuthService.Login(props.userEmail, props.userPassword);
      userStore.user = loginData.user;
      userStore.isUserAuthenticated = true;

      await themeStore.initializeTheme();
      await projectStore.loadStudios();

      const projectDirectoryExists = await SettingsService.GetProjectDirectory();

      if(projectDirectoryExists){
        await projectStore.loadProjects();
        trayStates.refreshData();
      } else {
        modals.setModalVisibility('dirOnboardModal', true);
      }

      emit('verification-success');
      
    } catch (loginError) {
      console.log("Auto-login failed:", loginError);
      notificationStore.errorNotification("Login Failed", "Verification successful but auto-login failed. Please try logging in manually.");
      emit('verification-success');
    }
    
  } catch (error) {
    console.log(error);
    isAwaitingResponse.value = false;
    notificationStore.errorNotification("Verification Failed", error.message || "Invalid verification code. Please try again.");
    error.value = error.message || "Invalid verification code. Please try again.";
    clearTokenInputs();
  }
};

const handleResendToken = async () => {
  if (isResendDisabled.value) return;
  
  isResendingToken.value = true;
  error.value = '';
  
  try {
    // TODO: Replace with actual service call once bindings are regenerated
    await AuthService.ResendToken(props.userEmail);
    
    // Simulate API call
    // await new Promise(resolve => setTimeout(resolve, 1000));
    
    notificationStore.addNotification("Code Sent", "A new verification code has been sent to your email.", "success");
    
    // Start cooldown timer (30 seconds)
    resendCooldown.value = 30;
    const cooldownInterval = setInterval(() => {
      resendCooldown.value--;
      if (resendCooldown.value <= 0) {
        clearInterval(cooldownInterval);
      }
    }, 1000);
    
  } catch (error) {
    console.log(error);
    notificationStore.errorNotification("Resend Failed", error.message || "Failed to resend verification code. Please try again.");
    error.value = error.message || "Failed to resend verification code. Please try again.";
  } finally {
    isResendingToken.value = false;
  }
};


onMounted(async () => {
  // Focus on the first token input when component mounts
  setTimeout(() => {
    if (tokenInputs.value[0]) {
      tokenInputs.value[0].focus();
    }
  }, 100);
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.loading-section{
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.loading-icon{
  width: 30px;
  height: 30px;
}

.auth-subheader {
  font-family: 'Inter', sans-serif;
  font-size: 1rem;
  font-weight: 300;
  color: var(--silver);
  margin-top: 0.5rem;
  text-align: left;
  width: 100%;
}

.token-input-container {
  display: flex;
  gap: 0.5rem;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  box-sizing: border-box;
}

.token-digit-input {
  width: 100%;
  height: 60px;
  text-align: center;
  font-size: 1.5rem;
  font-weight: 600;
  border: 2px solid var(--midnight-steel);
  border-radius: var(--normal-radius);
  background-color: var(--midnight-steel);
  color: var(--white);
  outline: none;
  transition: border-color 0.2s ease, background-color 0.2s ease;
}

.token-digit-input:focus {
  border-color: var(--grape);
  background-color: var(--black-steel);
}

.token-digit-input:valid {
  border-color: var(--bright-grape);
}

.token-input {
  text-align: center;
  font-size: 1.2rem;
  font-weight: 600;
  letter-spacing: 0.2rem;
  text-transform: uppercase;
}

.token-input::placeholder {
  letter-spacing: normal;
  text-transform: none;
  font-weight: 300;
}

.verification-actions {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  align-items: center;
  width: 100%;
}

.resend-container {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--grape);
  cursor: pointer;
  font-weight: 500;
  transition: color 0.2s;
  padding: 0.5rem;
  border-radius: var(--normal-radius);
  border: 1px solid transparent;
}

.resend-container:hover:not(.disabled) {
  color: var(--bright-grape);
  border-color: var(--grape);
}

.resend-container.disabled {
  opacity: 0.5;
  cursor: not-allowed;
  color: var(--light-gray);
}

@media (max-width: 768px) {
  .auth-subheader {
    font-size: 0.9rem;
    text-align: center;
  }
  
  .token-input-container {
    gap: 0.5rem;
  }
  
  .token-digit-input {
    width: 40px;
    height: 50px;
    font-size: 1.2rem;
  }
  
  .verification-actions {
    gap: 0.75rem;
  }
}
</style>