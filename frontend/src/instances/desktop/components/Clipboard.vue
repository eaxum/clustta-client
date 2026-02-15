<template>
  <div v-if="hasClipboardItems" class="clipboard-toast">
    <div class="clipboard-header">
      <div class="header-content">
        <h2 class="clipboard-title">{{ clipboardTitle }}</h2>
      </div>
      <ActionButton :icon="getAppIcon('broom')" v-tooltip="$t('components.clipboard.clearClipboard')" :buttonFunction="clearClipboard" />
    </div>

    <div class="clipboard-list-container">
      <ItemsList :items="clipboardItems" :forList="true" />
    </div>
  </div>
</template>

<script setup>
// imports
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ItemsList from '@/instances/desktop/components/ItemsList.vue';

// stores
import { useIconStore } from '@/stores/icons';
import { useStageStore } from '@/stores/stages';

const iconStore = useIconStore();
const stage = useStageStore();

const { t } = useI18n();

// computed

// Returns the clipboard items (either copied or cut).
const clipboardItems = computed(() => {
  if (stage.cutItems.length > 0) return stage.cutItems;
  return stage.copiedItems;
});

// Returns the title for the clipboard pane.
const clipboardTitle = computed(() => {
  const count = clipboardItems.value.length;
  const action = isCutMode.value ? t('components.clipboard.cut') : t('components.clipboard.copied');
  const unit = count !== 1 ? 'items' : 'item';
  return t('components.clipboard.clipboardTitle', { action, count, unit });
});

// Checks if there are items in the clipboard.
const hasClipboardItems = computed(() => {
  return stage.cutItems.length > 0 || stage.copiedItems.length > 0;
});

// Determines if items are in cut mode vs copy mode.
const isCutMode = computed(() => stage.cutItems.length > 0);

// methods

// Clears all clipboard items.
const clearClipboard = () => {
  stage.cutItems = [];
  stage.copiedItems = [];
};

// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);
</script>

<style scoped>
@import "@/assets/desktop.css";

.clipboard-toast {
  position: absolute;
  bottom: .5rem;
  left: .5rem;
  right: .5rem;
  z-index: 1;
  display: flex;
  flex-direction: column;
  background-color: var(--steel);
  border-radius: var(--large-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  height: 200px;
  box-sizing: border-box;
  max-height: 250px;
  animation: slideUp 0.2s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.clipboard-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  background-color: var(--midnight-steel);
  border-radius: var(--normal-radius);
}

.header-content {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
}

.clipboard-title {
  font-size: 14px;
  font-weight: 400;
  color: var(--white);
  margin: 0;
}

.clipboard-description {
  font-size: 12px;
  color: var(--silver);
  opacity: 0.9;
}

.clipboard-list-container {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 0.5rem;
  min-height: 0;
}

.clipboard-list-container::-webkit-scrollbar {
  width: 4px;
}

.clipboard-list-container::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.clipboard-list-container::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}
</style>
