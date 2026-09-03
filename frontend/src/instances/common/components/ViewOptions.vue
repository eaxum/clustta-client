<template>
  <div class="view-options-root" @mouseenter="expanded = true" @mouseleave="expanded = false">
    <ActionButton v-if="!isExpanded" :icon="getAppIcon(activeIcon)" v-tooltip="activeTooltip" />

    <template v-else>
      <ActionButton :icon="getAppIcon('list')" v-tooltip="$t('menus.listView')" :buttonFunction="setListView" />
      <ActionButton :icon="getAppIcon('four-squares')" v-tooltip="$t('menus.gridView')" :buttonFunction="setGridView" />
      <ActionButton :icon="getAppIcon('kanban')" v-tooltip="$t('menus.kanbanView')" :buttonFunction="setKanbanView" />
    </template>
  </div>
</template>

<script setup>
// imports
import { computed, ref } from 'vue';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// stores
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';

const commonStore = useCommonStore();
const iconStore = useIconStore();

// refs
const expanded = ref(false);

// computed properties
const isExpanded = computed(() => expanded.value || commonStore.viewMode === 'kanban');
const isGridActive = computed(() => commonStore.viewMode === 'grid');
const isKanbanActive = computed(() => commonStore.viewMode === 'kanban');

// Returns the icon name for the currently active view mode.
const activeIcon = computed(() => {
  if (isGridActive.value) return 'four-squares';
  if (isKanbanActive.value) return 'kanban';
  return 'list';
});

// Returns the tooltip for the currently active view mode.
const activeTooltip = computed(() => {
  return commonStore.viewMode;
});

// methods

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

const refreshView = () => emitter.emit('refresh-browser', {
  hardRefresh: true
});

// Sets the view to grid mode.
const setGridView = () => {
  commonStore.setGridView();
  refreshView();
};

// Sets the view to kanban mode.
const setKanbanView = () => {
  commonStore.setKanbanView();
  refreshView();
};

// Sets the view to list mode.
const setListView = () => {
  commonStore.setListView();
  refreshView();
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
  /* background-color: var(--surface-1); */
  border-radius: var(--large-radius);
}
</style>
