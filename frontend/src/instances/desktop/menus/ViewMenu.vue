<template>
  <div ref="viewMenu" class="filter-menu-container" v-stop-propagation>

    <!-- View Mode Section -->
    <!-- <ActionButton :icon="getAppIcon('list')" :showLabel="true" :fullWidth="true" :label="$t('menus.listView')"
      :color="isListActive ? 'var(--surface-3)' : undefined" :buttonFunction="setListView" />

    <ActionButton :icon="getAppIcon('list-compact')" :showLabel="true" :fullWidth="true" :label="$t('menus.compactView')"
      :color="isDenseActive ? 'var(--surface-3)' : undefined" :buttonFunction="setDenseView" />

    <ActionButton :icon="getAppIcon('four-squares')" :showLabel="true" :fullWidth="true" :label="$t('menus.gridView')"
      :color="isGridActive ? 'var(--surface-3)' : undefined" :buttonFunction="setGridView" />

    <ActionButton v-if="isDefaultWorkspace" :icon="getAppIcon('kanban')" :showLabel="true" :fullWidth="true" :label="$t('menus.kanbanView')"
      :color="isKanbanActive ? 'var(--surface-3)' : undefined" :buttonFunction="setKanbanView" />

    <span  v-if="isDefaultWorkspace && !isKanbanActive && userStore.canDo('update_collection')" class="menu-divider"></span> -->

    <!-- Display Options Section -->
    <ActionButton v-if="isDefaultWorkspace && !isKanbanActive && userStore.canDo('update_collection')"
      :icon="dndStore.lockUI ? getAppIcon('lock-closed') : getAppIcon('lock-open')" :showLabel="true" :fullWidth="true"
      :label="dndStore.lockUI ? $t('menus.unlockUI') : $t('menus.lockUI')" :buttonFunction="toggleLockUI" />

    <ActionButton v-if="!isKanbanActive"
      :icon="commonStore.hideExtensions ? getAppIcon('extension-cancel') : getAppIcon('extension')" :showLabel="true" :fullWidth="true"
      :label="commonStore.hideExtensions ? $t('modals.showExtensions') : $t('modals.hideExtensions')" :buttonFunction="toggleHideExtensions" />

    <ActionButton v-if="!isKanbanActive"
      :icon="commonStore.showFullPath ? getAppIcon('file-name') : getAppIcon('file-path')" :showLabel="true" :fullWidth="true"
      :label="commonStore.showFullPath ? $t('menus.showNameOnly') : $t('menus.showFullPath')" :buttonFunction="toggleShowFullPath" />

    <ActionButton v-if="!isKanbanActive"
      :icon="commonStore.showUntracked ? getAppIcon('eye-cancel') : getAppIcon('eye')" :showLabel="true" :fullWidth="true"
      :label="untrackedVisibilityLabel" :buttonFunction="toggleShowUntracked" />

    <ActionButton v-if="!isKanbanActive"
      :icon="getAppIcon('shapes')" :showLabel="true" :fullWidth="true"
      :label="settingsStore.showTypeIcons ? $t('modals.hideTypeIcons') : $t('modals.showTypeIcons')" :buttonFunction="toggleShowTypeIcons" />

    <span v-if="!isKanbanActive && !commonStore.useGrid" class="menu-divider"></span>

    <!-- Collapse Section -->
    <ActionButton v-if="!isKanbanActive && !commonStore.useGrid"
      :icon="getAppIcon('collapse-up')" :showLabel="true" :fullWidth="true" :label="$t('menus.collapseAll')"
      :buttonFunction="collapseAll" />

  </div>
</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import emitter from '@/lib/mitt';
import { useI18n } from 'vue-i18n';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// stores
import { useCommonStore } from '@/stores/common';
import { useDndStore } from '@/stores/dnd';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useSettingsStore } from '@/stores/settings';
import { useUserStore } from '@/stores/users';

const commonStore = useCommonStore();
const dndStore = useDndStore();
const iconStore = useIconStore();
const menu = useMenu();
const settingsStore = useSettingsStore();
const userStore = useUserStore();

const { t } = useI18n();

// refs
const viewMenu = ref(null);

// computed properties
const isDefaultWorkspace = computed(() => commonStore.activeWorkspace === 'Default');
const isDenseActive = computed(() => commonStore.viewMode === 'dense');
const isGridActive = computed(() => commonStore.viewMode === 'grid');
const isKanbanActive = computed(() => commonStore.viewMode === 'kanban');
const isListActive = computed(() => commonStore.viewMode === 'compact');
const untrackedVisibilityLabel = computed(() => `${commonStore.showUntracked ? t('common.hide') : t('common.show')} ${t('menus.untracked')}`);

// methods

// Collapses all expanded collections.
const collapseAll = () => {
  emitter.emit('collapse-all');
  menu.hideContextMenu();
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

const refreshView = () => emitter.emit('refresh-browser', {
  hardRefresh: true
});

// Sets the view to dense list mode.
const setDenseView = () => {
  commonStore.setDenseView();
  refreshView();
  menu.hideContextMenu();
};

// Sets the view to grid mode.
const setGridView = () => {
  commonStore.setGridView();
  refreshView();
  menu.hideContextMenu();
};

// Sets the view to kanban mode.
const setKanbanView = () => {
  commonStore.setKanbanView();
  refreshView();
  menu.hideContextMenu();
};

// Sets the view to list mode.
const setListView = () => {
  commonStore.setListView();
  refreshView();
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

// Toggles untracked item visibility and refreshes the browser.
const toggleShowUntracked = async () => {
  await commonStore.setUntrackedVisibility(!commonStore.showUntracked);
  emitter.emit('refresh-browser');
  menu.hideContextMenu();
};

// Toggles the show type icons option.
const toggleShowTypeIcons = async () => {
  await settingsStore.toggleShowTypeIcons();
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
