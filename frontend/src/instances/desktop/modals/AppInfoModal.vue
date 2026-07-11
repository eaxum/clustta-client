<template>

  <div class="modal-container" v-esc="closeModal">
    <div class="single-action-button close" @click="closeModal()">
      <img class="small-icons" src="/icons/close.svg">
    </div>
    <HeaderArea :title="title" :icon="'clustta'" :showSearch="showSearch" />

    <div class="general-container">
      <div class="version-info">
        <div> {{ osAppLabel }} </div>
        <div> {{ $t('modals.versionLabel') }} {{ clusttaVersion }} </div>
        <div> {{ $t('modals.copyrightNotice', { year: currentYear }) }} </div>
        <div> {{ $t('modals.trademarkNotice') }} </div>
      </div>

      <div class="update-row">
        <span v-if="updateStore.checked" class="update-status">{{ updateStatus }}</span>
        <GeneralButton :colored="false" :label="$t('modals.checkForUpdate')" :fullWidth="false" :buttonFunction="checkForUpdate" :loading="updateStore.isChecking" />
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
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { usePlatformStore } from '@/stores/platform';
import { useUpdateStore } from '@/stores/updates';

const { t } = useI18n();
const modals = useDesktopModalStore();
const platformStore = usePlatformStore();
const updateStore = useUpdateStore();

// refs
const clusttaVersion = ref('');

// constants
const showSearch = false;
const title = t('modals.clusttaApp');
const currentYear = new Date().getFullYear();

// computed properties
const osAppLabel = computed(() => {
  if (platformStore.isMac) return t('modals.clusttaForMac');
  if (platformStore.isLinux) return t('modals.clusttaForLinux');
  return t('modals.clusttaForWindows');
});

const updateStatus = computed(() => {
  if (updateStore.isUpdateAvailable) {
    return t('modals.updateAvailable', { version: updateStore.latestVersion });
  }
  return t('modals.upToDate');
});

// methods
// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Checks for an available update.
const checkForUpdate = async () => {
  await updateStore.checkForUpdate();
};

// lifecycle hooks
onMounted(async () => {
  await platformStore.initialize();
  clusttaVersion.value = await utils.getRawClusttaVersion();
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.close {
  opacity: .5;
  border-radius: 60px;
  position: absolute;
  top: 8px;
  right: 8px;
}

.close:hover {
  opacity: 1;
  background-color: crimson;
  transform: scale(1.02);
}

.version-info {
  gap: .2rem;
  font-size: 14px;
  width: 100%;
  display: flex;
  flex-direction: column;
  padding: .2rem;
  color: var(--text);
}

.general-container {
  gap: .3rem;
  padding-top: .5rem;
  align-items: flex-start;
}

.update-row {
  gap: .5rem;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding-top: .5rem;
}

.update-status {
  font-size: 13px;
  color: var(--text);
  margin-right: auto;
}
</style>



