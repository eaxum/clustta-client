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
          @click="trayStates.popUpModalFunction" :isActive="true" :loading="false" />
      </div>

    </div>
  </div>


</template>

<script setup>
import { ref, computed, onMounted, nextTick, onBeforeUnmount } from 'vue';
import { useTrayStates } from '@/stores/TrayStates';
import { useDesktopModalStore } from '@/stores/desktopModals';

import HeaderArea from '@/instances/common/components/HeaderArea.vue'
import GeneralButton from '@/instances/common/components/GeneralButton.vue';

const trayStates = useTrayStates();
const modals = useDesktopModalStore();

let title = trayStates.popUpModalTitle;
let icon = trayStates.popUpModalIcon;
let showSearch = false;

const popUpInput = ref(null);

const leftButton = computed(() => {
  return trayStates.popUpModalButtons[0]
});

const rightButton = computed(() => {
  return trayStates.popUpModalButtons[1]
});

const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    trayStates.popUpModalFunction();
  }
};

const closeModal = (modalName) => {
  modals.disableAllModals();
  trayStates.popUpModalInputValue = '';
};

onMounted(() => {

});

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
  font-size: 14px;
  color: var(--white);
}

.input-short {
  width: 100%;
}
</style>