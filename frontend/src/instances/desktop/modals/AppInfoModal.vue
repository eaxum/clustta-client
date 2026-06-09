<template>

  <div class="modal-container" v-esc="closeModal">
    <div class="single-action-button close" @click="closeModal()">
      <img class="small-icons" src="/icons/close.svg">
    </div>
    <HeaderArea :title="title" :icon="'clustta'" :showSearch="showSearch" />

    <div class="general-container">
      <div class="version-info">
        <!-- <div> Clustta for Windows (64 bit) </div> -->
        <div> {{ $t('modals.versionLabel') }} {{ clusttaVersion }} </div>
        <div> {{ $t('modals.copyrightNotice') }} </div>
        <div> {{ $t('modals.trademarkNotice') }} </div>
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import utils from '@/services/utils';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';

const { t } = useI18n();
const modals = useDesktopModalStore();

// refs
const clusttaVersion = ref('');

// constants
const showSearch = false;
const title = t('modals.clusttaApp');

// methods
// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// lifecycle hooks
onMounted(async () => {
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
  background-color: hsl(var(--destructive));
  transform: scale(1.02);
}

.version-info {
  gap: .2rem;
  font-size: 14px;
  width: 100%;
  display: flex;
  flex-direction: column;
  padding: .2rem;
  color: hsl(var(--foreground));
}

.general-container {
  gap: .3rem;
  padding-top: .5rem;
  align-items: flex-start;
}
</style>



