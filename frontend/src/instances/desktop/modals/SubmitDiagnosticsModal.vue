<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>
    <HeaderArea :title="$t('modals.submitDiagnosticReport')" :icon="'bug'" />
    <div class="general-container">

      <p class="description-text">{{ $t('modals.diagnosticDescription') }}</p>

      <FormInput v-model="email" :labelTop="true" :placeholder="$t('placeholders.yourEmail')" type="email"
        :needsValidation="true" :showValidation="email.length > 0" :valid="isEmailValid" :error="emailError" />

      <textarea v-model="message" class="desktop-input-long" type="text" :placeholder="$t('placeholders.describeProblem')"
        @keydown.enter="handleEnterKey" />

      <InputAlert :show="!isFormValid && message.length > 0" :message="validationMessage" />

      <div class="system-info-section">
        <div class="system-info-header">
          <img class="small-icons" :src="getAppIcon('monitor')">
          <span>{{ $t('modals.systemInformation') }}</span>
        </div>
        <div class="system-info-content">
          <div class="info-row"><span class="info-label">{{ $t('modals.osLabel') }}</span><span class="info-value">{{ systemInfo.os }}</span></div>
          <div class="info-row"><span class="info-label">{{ $t('modals.osVersionLabel') }}</span><span class="info-value">{{ systemInfo.osVersion }}</span></div>
          <div class="info-row"><span class="info-label">{{ $t('modals.architectureLabel') }}</span><span class="info-value">{{ systemInfo.arch }}</span></div>
          <div class="info-row"><span class="info-label">{{ $t('modals.clusttaVersionLabel') }}</span><span class="info-value">{{ systemInfo.clusttaVersion }}</span></div>
        </div>
      </div>

      <div class="attachment-section">
        <img class="small-icons" :src="getAppIcon('paperclip')">
        <span class="attachment-label">{{ logFileName || $t('modals.noLogFile') }}</span>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.send')" :fullWidth="true" @click="submitDiagnostics" :isActive="isFormValid"
          :loading="isAwaitingResponse" />
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import utils from '@/services/utils';

// components
import FormInput from '@/instances/desktop/components/FormInput.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';

// services
import { AppService, AuthService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';

const { t } = useI18n();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();

// refs
const email = ref('');
const isAwaitingResponse = ref(false);
const logFileName = ref('');
const message = ref('');
const modalContainer = ref(null);
const systemInfo = ref({
  os: '',
  osVersion: '',
  arch: '',
  clusttaVersion: '',
});

// computed

// Returns the email validation error message.
const emailError = computed(() => {
  if (email.value.length === 0) return '';
  if (!isEmailValid.value) return t('notifications.invalidEmail');
  return '';
});

// Returns whether the email format is valid.
const isEmailValid = computed(() => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email.value);
});

// Returns whether the form is valid for submission.
const isFormValid = computed(() => {
  // Email is optional, but must be valid if provided
  const emailValid = email.value.length === 0 || isEmailValid.value;
  return emailValid && message.value.trim().length > 10;
});

// Returns the validation message for the form.
const validationMessage = computed(() => {
  if (message.value.trim().length <= 10) {
    return t('notifications.moreDetailedDescription');
  }
  return '';
});

// methods

// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles enter key press to submit form.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && !event.shiftKey && isFormValid.value) {
    event.preventDefault();
    submitDiagnostics();
  }
};

// Loads system information on mount.
const loadSystemInfo = async () => {
  try {
    const sysInfo = await AppService.GetSystemInfo();
    const version = await utils.getRawClusttaVersion();

    // Format OS name
    const osName = sysInfo.os.charAt(0).toUpperCase() + sysInfo.os.slice(1);
    
    systemInfo.value = {
      os: osName === 'Darwin' ? 'macOS' : osName,
      osVersion: sysInfo.os_version || 'Unknown',
      arch: sysInfo.arch || 'Unknown',
      clusttaVersion: version,
    };

    // Generate log file name with human-readable timestamp
    const now = new Date();
    const options = { month: 'long', day: 'numeric', hour: 'numeric', minute: '2-digit', hour12: true, timeZoneName: 'short' };
    const humanReadableDate = now.toLocaleDateString('en-US', options).replace(',', '').replace(/:/g, '-');
    logFileName.value = `clustta-diagnostics-${humanReadableDate}.log`;
  } catch (error) {
    console.error('Failed to load system info:', error);
  }
};

// Submits diagnostics to support.
const submitDiagnostics = async () => {
  if (!isFormValid.value) return;

  isAwaitingResponse.value = true;
  try {
    await AuthService.SubmitDiagnostics(
      email.value,
      message.value,
      `${systemInfo.value.os} ${systemInfo.value.osVersion}`,
      systemInfo.value.arch,
      systemInfo.value.clusttaVersion
    );
    notificationStore.addNotification(t('notifications.reportSent'), t('notifications.reportSentDescription'), 'success');
    closeModal();
  } catch (error) {
    notificationStore.errorNotification(t('notifications.failedToSendReport'), error);
  } finally {
    isAwaitingResponse.value = false;
  }
};

// lifecycle hooks
onMounted(async () => {
  await loadSystemInfo();
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.description-text {
  font-family: Inter, sans-serif;
  font-size: 14px;
  color: var(--silver);
  line-height: 1.5;
  margin-bottom: 0.5rem;
}

.desktop-input-long {
  margin-top: 0px;
  font-weight: 200;
  color: var(--white);
  min-height: 100px;
}

.general-container {
  gap: .5rem;
  align-items: center;
  justify-content: flex-start;
}

.system-info-section {
  display: flex;
  flex-direction: column;
  width: 100%;
  padding: 0.8rem;
  background-color: var(--midnight-steel);
  border-radius: var(--normal-radius);
  box-sizing: border-box;
  margin-top: 0.5rem;
}

.system-info-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 13px;
  font-weight: 600;
  color: var(--white);
  margin-bottom: 0.5rem;
}

.system-info-content {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

.info-row {
  display: flex;
  font-size: 12px;
}

.info-label {
  color: var(--silver);
  min-width: 120px;
}

.info-value {
  color: var(--white);
}

.attachment-section {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.6rem 0.8rem;
  background-color: var(--dark-steel);
  border-radius: var(--normal-radius);
  box-sizing: border-box;
  margin-top: 0.3rem;
}

.attachment-label {
  font-size: 12px;
  color: var(--silver);
}
</style>
