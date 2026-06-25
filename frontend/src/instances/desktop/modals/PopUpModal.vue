<template>

  <div class="modal-container" ref="modalContainer" v-stop-propagation v-return="handleEnterKey">
    <HeaderArea :title="title" :icon="icon" :showSearch="showSearch" />
    <div class="general-container">

      <div v-if="trayStates.popUpModalMessage" class="pop-up-text-container">
        <div class="pop-up-body">
          {{ trayStates.popUpModalMessage }}
        </div>
      </div>

      <div v-if="trayStates.usePopUpModalInput" class="input-section">
        <div class="horizontal-flex">
          <input ref="popUpInput" class="input-short" type="text" :placeholder="trayStates.popUpModalPlaceholder"
            v-model="trayStates.popUpModalInputValue" v-focus />
        </div>
      </div>


      <div class="pop-up-actions">
        <GeneralButton v-if="leftButton" :label="leftButton" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton v-if="trayStates.popUpModalFunction && rightButton" :label="rightButton" :fullWidth="true"
          @click="trayStates.popUpModalFunction" :isActive="true" :loading="trayStates.popUpModalLoading" />
      </div>

    </div>
  </div>


</template>

<script setup>
// imports
import { computed, onBeforeUnmount, ref } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useTrayStates } from '@/stores/TrayStates';

const modals = useDesktopModalStore();
const trayStates = useTrayStates();
const { t } = useI18n();

// refs
const popUpInput = ref(null);

// constants
const icon = trayStates.popUpModalIcon;
const showSearch = false;
const title = trayStates.popUpModalTitle;

// computed
// Returns the left button label.
const leftButton = computed(() => {
  const label = trayStates.popUpModalButtons[0];
  return label === 'Cancel' ? t('common.cancel') : label;
});

// Returns the right button label.
const rightButton = computed(() => {
  const label = trayStates.popUpModalButtons[1];
  return label === 'Confirm' ? t('common.confirm') : label;
});

// methods
// Closes the modal and resets input value.
const closeModal = () => {
  modals.disableAllModals();
  trayStates.popUpModalInputValue = '';
};

// Handles enter key press to execute modal function.
const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    trayStates.popUpModalFunction();
  }
};

// lifecycle hooks
onBeforeUnmount(() => {
  trayStates.usePopUpModalInput = false;
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.pop-up-text-container {
  /* background-color: crimson; */
  padding: .5rem;
}

.pop-up-info {
  padding: 0rem 1rem;
  margin-bottom: .75rem;
}

.pop-up-body {
  font-size: 13px;
  color: var(--text);
}

.input-short {
  width: 100%;
}
</style>

