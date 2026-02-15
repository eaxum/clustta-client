<template>
  <div ref="sortMenu" class="filter-menu-container" v-stop-propagation>

    <!-- Sort Options Section -->
    <ActionButton :icon="getAppIcon('sort-a-z')" :showLabel="true" :fullWidth="true" :label="$t('menus.sortAlphabetically')"
      :color="isAlphabeticalActive ? 'var(--steel)' : undefined" :buttonFunction="setSortByName" />

    <ActionButton :icon="getAppIcon('clock')" :showLabel="true" :fullWidth="true" :label="$t('menus.sortByStatus')"
      :color="isStatusActive ? 'var(--steel)' : undefined" :buttonFunction="setSortByStatus" />

    <span class="menu-divider"></span>

    <!-- Sort Order Section -->
    <ActionButton :icon="getAppIcon(ascendingIcon)" :showLabel="true" :fullWidth="true" :label="$t('menus.ascending')"
      :color="isAscending ? 'var(--steel)' : undefined" :buttonFunction="setSortAscending" />

    <ActionButton :icon="getAppIcon(descendingIcon)" :showLabel="true" :fullWidth="true" :label="$t('menus.descending')"
      :color="!isAscending ? 'var(--steel)' : undefined" :buttonFunction="setSortDescending" />

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
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';

const commonStore = useCommonStore();
const iconStore = useIconStore();
const menu = useMenu();

const { t } = useI18n();

// refs
const sortMenu = ref(null);

// computed properties
const ascendingIcon = computed(() => commonStore.sortBy === 'status' ? 'sort-ascending-shapes' : 'sort-ascending-letters');
const descendingIcon = computed(() => commonStore.sortBy === 'status' ? 'sort-descending-shapes' : 'sort-descending-letters');
const isAlphabeticalActive = computed(() => commonStore.sortBy === 'name');
const isAscending = computed(() => commonStore.sortOrder === 'asc');
const isStatusActive = computed(() => commonStore.sortBy === 'status');

// methods

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Sets the sort order to ascending.
const setSortAscending = () => {
  commonStore.sortOrder = 'asc';
  emitter.emit('refresh-browser');
  menu.hideContextMenu();
};

// Sets the sort by name (alphabetical).
const setSortByName = () => {
  commonStore.sortBy = 'name';
  emitter.emit('refresh-browser');
  menu.hideContextMenu();
};

// Sets the sort by status.
const setSortByStatus = () => {
  commonStore.sortBy = 'status';
  emitter.emit('refresh-browser');
  menu.hideContextMenu();
};

// Sets the sort order to descending.
const setSortDescending = () => {
  commonStore.sortOrder = 'desc';
  emitter.emit('refresh-browser');
  menu.hideContextMenu();
};

// Updates menu dimensions when mounted.
const updateMenuDimensions = () => {
  if (!sortMenu.value) return;
  menu.assetMenuWidth = sortMenu.value.getBoundingClientRect().width;
  menu.assetMenuHeight = sortMenu.value.getBoundingClientRect().height;
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
