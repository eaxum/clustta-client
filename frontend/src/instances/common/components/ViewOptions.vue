<template>
  <div class="view-options-root" @mouseenter="expanded = true" @mouseleave="expanded = false">
    <ActionButton v-if="!expanded" :icon="resolveIcon(activeIcon)" v-tooltip="activeTooltip" />

    <template v-else>
      <ActionButton :icon="CiList" v-tooltip="$t('menus.listView')" :buttonFunction="setListView" />
      <ActionButton :icon="CiListCompact" v-tooltip="$t('menus.compactView')" :buttonFunction="setDenseView" />
      <ActionButton :icon="CiFourSquares" v-tooltip="$t('menus.gridView')" :buttonFunction="setGridView" />
      <ActionButton v-if="isDefaultWorkspace" :icon="CiKanban" v-tooltip="$t('menus.kanbanView')" :buttonFunction="setKanbanView" />
    </template>
  </div>
</template>

<script setup>
// imports
import { computed, ref } from 'vue';
import emitter from '@/lib/mitt';
import { CiFourSquares, CiKanban, CiList, CiListCompact } from '@clustta/icons-vue';
import { resolveIcon } from '@/lib/icon-map';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// stores
import { useCommonStore } from '@/stores/common';

const commonStore = useCommonStore();

// refs
const expanded = ref(false);

// computed properties
const isDefaultWorkspace = computed(() => commonStore.activeWorkspace === 'Default');
const isDenseActive = computed(() => commonStore.viewMode === 'dense');
const isGridActive = computed(() => commonStore.viewMode === 'grid');
const isKanbanActive = computed(() => commonStore.viewMode === 'kanban');
const isListActive = computed(() => commonStore.viewMode === 'compact');

// Returns the icon name for the currently active view mode.
const activeIcon = computed(() => {
  if (isDenseActive.value) return 'list-compact';
  if (isGridActive.value) return 'four-squares';
  if (isKanbanActive.value) return 'kanban';
  return 'list';
});

// Returns the tooltip for the currently active view mode.
const activeTooltip = computed(() => {
  return commonStore.viewMode;
});

// methods

// Sets the view to dense list mode.
const setDenseView = () => {
  commonStore.setDenseView();
  emitter.emit('refresh-browser');
};

// Sets the view to grid mode.
const setGridView = () => {
  commonStore.setGridView();
  emitter.emit('refresh-browser');
};

// Sets the view to kanban mode.
const setKanbanView = () => {
  commonStore.setKanbanView();
  emitter.emit('refresh-browser');
};

// Sets the view to list mode.
const setListView = () => {
  commonStore.setListView();
  emitter.emit('refresh-browser');
};
</script>

<style scoped>
.view-options-root {
  display: flex;
  gap: .1rem;
  align-items: center;
  justify-content: flex-end;
  box-sizing: border-box;
  height: min-content;
  overflow: hidden;
}

.view-options-root:hover {
  /* padding: .2rem; */
  /* background-color: var(--black-steel); */
  border-radius: var(--large-radius);
}
</style>
