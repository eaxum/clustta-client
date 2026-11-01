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
import { onMounted, ref } from 'vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import { useDesktopModalStore } from '@/stores/desktopModals';

const LICENSE_NOTICES_PATH = 'licenses/THIRD-PARTY-NOTICES.txt';
const LICENSE_NOTICES_LOADING_MESSAGE = 'Loading third-party notices...';
const LICENSE_NOTICES_ERROR_MESSAGE = 'Third-party notices could not be loaded.';

const modals = useDesktopModalStore();
const noticesText = ref(LICENSE_NOTICES_LOADING_MESSAGE);
const closeModal = () => modals.disableAllModals();

const loadNotices = async () => {
  try {
    const noticesUrl = `${import.meta.env.BASE_URL}${LICENSE_NOTICES_PATH}`;
    const response = await fetch(noticesUrl);
    if (!response.ok) {
      throw new Error(`Failed to load third-party notices: ${response.status}`);
    }
    noticesText.value = await response.text();
  } catch (error) {
    console.error(LICENSE_NOTICES_ERROR_MESSAGE, error);
    noticesText.value = LICENSE_NOTICES_ERROR_MESSAGE;
  }
};

onMounted(loadNotices);
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
  border-radius: 10px 10px 18px 18px;
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
