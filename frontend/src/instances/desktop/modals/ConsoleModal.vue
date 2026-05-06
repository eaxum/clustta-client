<template>
  <div class="modal-container console-modal-container" v-stop-propagation>
    <div class="console-modal-header">
      <div class="console-modal-title">
        <CiConsole :size="20" class="console-modal-icon" />
        <span>{{ t('panes.consoleTab') }}</span>
      </div>
      <ActionButton :icon="CiClose" :showLabel="false" v-tooltip="$t('common.close')" :buttonFunction="closeModal" />
    </div>

    <div class="console-modal-body">
      <Console :isModal="true" />
    </div>
  </div>
</template>

<script setup>
// imports
import { useI18n } from 'vue-i18n';
import { CiClose, CiConsole } from '@clustta/icons-vue';
import { resolveIcon } from '@/lib/icon-map';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import Console from '@/instances/desktop/panes/Console.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();
const modals = useDesktopModalStore();
const { t } = useI18n();

// methods

// Closes the console modal.
const closeModal = () => {
  modals.setModalVisibility('consoleModal', false);
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.resolveIcon(iconName);
</script>

<style scoped>
@import "@/assets/desktop.css";

.console-modal-container {
  width: 80vw;
  max-width: 800px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  padding: 0 .5rem;
}

.console-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0.5rem;
  background-color: blue;
  border-radius: var(--small-radius);
  background-color: var(--midnight-steel);
  outline: var(--transparent-line);
}

.console-modal-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 18px;
  font-weight: 500;
  color: var(--white);
}

.console-modal-icon {
  width: 16px;
  height: 16px;
}

.console-modal-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
</style>
