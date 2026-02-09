<template>
  <div ref="viewMenu" class="filter-menu-container" v-stop-propagation>

    <!-- View Mode Section -->
    <ActionButton :icon="getAppIcon('list-compact')" :showLabel="true" :fullWidth="true" label="List View"
      :color="isListActive ? 'var(--steel)' : undefined" :buttonFunction="setListView" />

    <ActionButton :icon="getAppIcon('four-squares')" :showLabel="true" :fullWidth="true" label="Grid View"
      :color="isGridActive ? 'var(--steel)' : undefined" :buttonFunction="setGridView" />

    <ActionButton v-if="isDefaultWorkspace" :icon="getAppIcon('kanban')" :showLabel="true" :fullWidth="true" label="Kanban View"
      :color="isKanbanActive ? 'var(--steel)' : undefined" :buttonFunction="setKanbanView" />

    <span  v-if="isDefaultWorkspace && !isKanbanActive && userStore.canDo('update_entity')" class="menu-divider"></span>

    <!-- Display Options Section -->
    <ActionButton v-if="isDefaultWorkspace && !isKanbanActive && userStore.canDo('update_entity')"
      :icon="dndStore.lockUI ? getAppIcon('lock-closed') : getAppIcon('lock-open')" :showLabel="true" :fullWidth="true"
      :label="dndStore.lockUI ? 'Unlock UI' : 'Lock UI'" :buttonFunction="toggleLockUI" />

    <ActionButton v-if="!isKanbanActive"
      :icon="commonStore.hideExtensions ? getAppIcon('extension-cancel') : getAppIcon('extension')" :showLabel="true" :fullWidth="true"
      :label="commonStore.hideExtensions ? 'Show Extensions' : 'Hide Extensions'" :buttonFunction="toggleHideExtensions" />

    <ActionButton v-if="!isKanbanActive"
      :icon="commonStore.showFullPath ? getAppIcon('file-name') : getAppIcon('file-path')" :showLabel="true" :fullWidth="true"
      :label="commonStore.showFullPath ? 'Show Name Only' : 'Show Full Path'" :buttonFunction="toggleShowFullPath" />

    <span v-if="!isKanbanActive && !commonStore.useGrid" class="menu-divider"></span>

    <!-- Collapse Section -->
    <ActionButton v-if="!isKanbanActive && !commonStore.useGrid"
      :icon="getAppIcon('collapse-up')" :showLabel="true" :fullWidth="true" label="Collapse All"
      :buttonFunction="collapseAll" />

  </div>
</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// stores
import { useCommonStore } from '@/stores/common';
import { useDndStore } from '@/stores/dnd';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useUserStore } from '@/stores/users';

const commonStore = useCommonStore();
const dndStore = useDndStore();
const iconStore = useIconStore();
const menu = useMenu();
const userStore = useUserStore();

// refs
const viewMenu = ref(null);

// computed properties
const isDefaultWorkspace = computed(() => commonStore.activeWorkspace === 'Default');
const isGridActive = computed(() => commonStore.viewMode === 'grid');
const isKanbanActive = computed(() => commonStore.viewMode === 'kanban');
const isListActive = computed(() => commonStore.viewMode === 'compact');

// methods

// Collapses all expanded entities.
const collapseAll = () => {
  emitter.emit('collapse-all');
  menu.hideContextMenu();
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Sets the view to grid mode.
const setGridView = () => {
  commonStore.setGridView();
  emitter.emit('refresh-browser');
  menu.hideContextMenu();
};

// Sets the view to kanban mode.
const setKanbanView = () => {
  commonStore.setKanbanView();
  emitter.emit('refresh-browser');
  menu.hideContextMenu();
};

// Sets the view to list mode.
const setListView = () => {
  commonStore.setListView();
  emitter.emit('refresh-browser');
  menu.hideContextMenu();
};

// Toggles the hide extensions option.
const toggleHideExtensions = () => {
  commonStore.hideExtensions = !commonStore.hideExtensions;
  menu.hideContextMenu();
};

// Toggles the lock UI option.
const toggleLockUI = () => {
  dndStore.lockUI = !dndStore.lockUI;
  menu.hideContextMenu();
};

// Toggles the show full path option.
const toggleShowFullPath = () => {
  commonStore.showFullPath = !commonStore.showFullPath;
  menu.hideContextMenu();
};

// Updates menu dimensions when mounted.
const updateMenuDimensions = () => {
  if (!viewMenu.value) return;
  menu.assetMenuWidth = viewMenu.value.getBoundingClientRect().width;
  menu.assetMenuHeight = viewMenu.value.getBoundingClientRect().height;
};

// lifecycle hooks
onMounted(() => {
  updateMenuDimensions();
});

onBeforeUnmount(() => {
  menu.hideContextMenu();
});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/menu.css";
</style>