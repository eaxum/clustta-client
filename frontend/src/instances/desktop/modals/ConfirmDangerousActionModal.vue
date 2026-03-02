<template>
  <div class="modal-container" ref="modalContainer" v-stop-propagation>
    <HeaderArea :title="trayStates.dangerousActionTitle" :icon="trayStates.dangerousActionIcon" :showSearch="false" />

    <div class="general-container">
      <div class="message-container">
        <p class="message-body">{{ trayStates.dangerousActionMessage }} <span v-if="activeToggleHint" :class="{ 'hint-destructive': deleteWorkingFiles }">{{ activeToggleHint }}</span></p>
      </div>

      <FormInput v-if="trayStates.dangerousActionShowInput" v-model="inputValue" :placeholder="trayStates.dangerousActionConfirmText" :error="inputError"
        :labelTop="true" :autofocus="true" @input="clearError" />

    <span v-if="trayStates.dangerousActionShowToggle" class="toggle-row" @click="toggleDeleteFiles">
        <div class="toggle-label">{{ trayStates.dangerousActionToggleLabel }}</div>
        <ToggleSwitch :switchValueProp="deleteWorkingFiles" />
      </span>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.confirm')" :fullWidth="true" :buttonFunction="confirmAction"
          :isActive="isInputValid" :isDisabled="!isInputValid" :loading="isLoading" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onBeforeUnmount, ref } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import FormInput from '@/instances/desktop/components/FormInput.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useTrayStates } from '@/stores/TrayStates';

const { t } = useI18n();
const modals = useDesktopModalStore();
const trayStates = useTrayStates();

// refs
const deleteWorkingFiles = ref(false);
const inputError = ref('');
const inputValue = ref('');
const isLoading = ref(false);

// computed
// Returns the appropriate toggle hint based on toggle state.
const activeToggleHint = computed(() => {
  if (!trayStates.dangerousActionShowToggle) return '';
  return deleteWorkingFiles.value
    ? trayStates.dangerousActionToggleOnHint
    : trayStates.dangerousActionToggleOffHint;
});

// Checks if the input value matches the confirmation text.
const isInputValid = computed(() => {
  if (!trayStates.dangerousActionShowInput) return true;
  return inputValue.value === trayStates.dangerousActionConfirmText;
});

// methods
// Clears any error message when input changes.
const clearError = () => {
  inputError.value = '';
};

// Closes the modal and resets state.
const closeModal = () => {
  deleteWorkingFiles.value = false;
  inputValue.value = '';
  inputError.value = '';
  isLoading.value = false;
  modals.setModalVisibility('confirmDangerousActionModal', false);
};

// Executes the dangerous action after validation.
const confirmAction = async () => {
  if (!isInputValid.value) {
    inputError.value = t('modals.confirmDangerousAction.nameMismatch');
    return;
  }

  isLoading.value = true;
  inputError.value = '';

  try {
    await trayStates.dangerousActionFunction({ deleteWorkingFiles: deleteWorkingFiles.value });
    closeModal();
  } catch (error) {
    console.error(error);
    inputError.value = error.message || t('notifications.errorOccurred');
    isLoading.value = false;
  }
};

// Toggles the delete working files switch.
const toggleDeleteFiles = () => {
  deleteWorkingFiles.value = !deleteWorkingFiles.value;
};

// lifecycle hooks
onBeforeUnmount(() => {
  deleteWorkingFiles.value = false;
  inputValue.value = '';
  inputError.value = '';
  isLoading.value = false;
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.general-container{
    gap: 0px !important;
    padding-top: 0px;
}

.message-body {
  font-size: 13px;
  color: var(--white);
  line-height: 1.5;
}

.hint-destructive {
  color: var(--danger);
  font-weight: 500;
}

.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
  cursor: pointer;
  width: 100%;
}

.toggle-label {
  font-size: 13px;
  color: var(--white);
}

</style>
