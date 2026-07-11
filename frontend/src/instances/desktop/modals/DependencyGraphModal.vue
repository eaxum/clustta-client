<template>
  <div class="modal-container dependency-graph-modal-container" v-stop-propagation>
    <div class="dependency-graph-modal-header">
      <div class="dependency-graph-modal-title">
        <img class="dependency-graph-modal-icon" :src="getAppIcon('dependency')">
        <span>{{ t('menus.dependencyGraph') }}</span>
      </div>
      <ActionButton :icon="getAppIcon('close')" :showLabel="false" v-tooltip="$t('common.close')" :buttonFunction="closeModal" />
    </div>

    <div class="dependency-graph-modal-body">
      <div class="dependency-graph-component">
        <DependencyGraph />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { useI18n } from 'vue-i18n';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DependencyGraph from '@/instances/desktop/components/DependencyGraph.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();
const modals = useDesktopModalStore();
const { t } = useI18n();

// methods

// Closes the dependency graph modal.
const closeModal = () => {
  modals.setModalVisibility('dependencyGraphModal', false);
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);
</script>

<style scoped>
@import "@/assets/desktop.css";

.dependency-graph-modal-container {
  display: flex;
  flex-direction: column;
}

.dependency-graph-modal-header {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0.5rem 1rem;
  border-radius: var(--small-radius);
  background-color: var(--bg);
  outline: var(--transparent-line);
}

.dependency-graph-modal-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 18px;
  font-weight: 500;
  color: var(--text);
}

.dependency-graph-modal-icon {
  width: 16px;
  height: 16px;
}

.dependency-graph-modal-body {
  flex: 1;
  width: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dependency-graph-component {
  display: flex;
  width: 90vw;
  height: 80vh;
  min-width: 480px;
  min-height: 480px;
  max-width: 1400px;
  max-height: 85vh;
}

</style>
