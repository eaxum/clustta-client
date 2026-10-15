<template>
  <div class="view-options-root">
    <ActionButton :icon="getAppIcon('list')" :isActive="isListActive" v-tooltip="$t('menus.listView')" :buttonFunction="setListView" />
    <ActionButton :icon="getAppIcon('four-squares')" :isActive="isGridActive" v-tooltip="$t('menus.gridView')" :buttonFunction="setGridView" />
    <ActionButton :icon="getAppIcon('columns')" :isActive="isColumnsActive" v-tooltip="$t('menus.columnsView')" :buttonFunction="setColumnsView" />
    <ActionButton :icon="getAppIcon('kanban')" :isActive="isKanbanActive" v-tooltip="$t('menus.kanbanView')" :buttonFunction="setKanbanView" />
  </div>
</template>

<script setup>
// imports
import { computed } from 'vue';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// stores
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';

const commonStore = useCommonStore();
const iconStore = useIconStore();

// computed properties
const isListActive = computed(() => commonStore.viewMode === 'list');
const isColumnsActive = computed(() => commonStore.viewMode === 'columns');
const isGridActive = computed(() => commonStore.viewMode === 'grid');
const isKanbanActive = computed(() => commonStore.viewMode === 'kanban');

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

// Sets the view to columns mode.
const setColumnsView = () => {
  commonStore.setColumnsView();
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
</style>
