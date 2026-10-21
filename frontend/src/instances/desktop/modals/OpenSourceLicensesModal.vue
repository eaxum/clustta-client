<template>
  <div class="modal-container licenses-modal" v-esc="closeModal">
    <div class="single-action-button close" @click="closeModal">
      <img class="small-icons" src="/icons/close.svg">
    </div>
    <HeaderArea :title="$t('modals.openSourceLicenses')" icon="code-bracket" :showSearch="false" />

    <div class="general-container general-container-wide">
      <pre class="licenses-content">{{ noticesText }}</pre>
    </div>
  </div>
</template>

<script setup>
import noticesText from '../../../../public/licenses/THIRD-PARTY-NOTICES.txt?raw';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import { useDesktopModalStore } from '@/stores/desktopModals';

const modals = useDesktopModalStore();
const closeModal = () => modals.disableAllModals();
</script>

<style scoped>
@import "@/assets/desktop.css";

.licenses-modal {
  align-items: flex-start;
  justify-content: flex-start;
  width: min(900px, 90vw);
  max-height: 90vh;
}

.close {
  position: absolute;
  top: 8px;
  right: 8px;
  border-radius: 60px;
  opacity: .5;
}

.close:hover {
  border-radius: 60px;
  background-color: crimson;
  opacity: 1;
  transform: scale(1.02);
}

.general-container {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: .5rem;
  overflow: hidden;
  box-sizing: border-box;
}

.general-container-wide {
  /* min-width: 500px !important; */
  min-width: 80vw;
}

.licenses-content {
  box-sizing: border-box;
  width: 100%;
  max-height: 65vh;
  width: 100%;
  margin: 0;
  padding: 1rem;
  overflow: hidden;
  overflow-y: scroll;
  border-radius: var(--small-radius);
  color: var(--text);
  background-color: var(--surface-3);
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.licenses-content::-webkit-scrollbar {
  width: 4px;
}

.licenses-content::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-4);
}

.licenses-content::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}
</style>
