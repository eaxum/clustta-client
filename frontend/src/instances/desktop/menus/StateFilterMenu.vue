<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span v-for="state in allStates" class="filter-menu-item" @click="toggleFilter(state)">
      <img class="small-icons" :class="{ 'no-filter' : isColored(state?.name)}" :src="getAppIcon(state.icon)">
      <div class="horizontal-flex">
        <div class="menu-item-text"> {{ utils.capitalizeStr(state.name) }} </div>
        <CheckBox :modelValue="isFilterActive(state)" :ariaLabel="`Filter by ${state.name}`" />
      </div>
    </span>

  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';

// components
import CheckBox from '@/instances/common/components/CheckBox.vue';

// stores
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';

const commonStore = useCommonStore();
const iconStore = useIconStore();
const menu = useMenu();

// refs
const collectionMenu = ref(null);

// computed properties
// Returns list of file states available for filtering.
const allStates = computed(() => {
  return commonStore.fileStates;
});

// methods
// Adds a filter to both asset and resource filters lists.
const addFilter = (filter) => {
  commonStore.assetFilters.push(filter);
  commonStore.resourceFilters.push(filter);
};

// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Checks if a state should display in color.
const isColored = (stateName) => {
  const coloredItems = ['modified', 'outdated'];
  return coloredItems.includes(stateName);
};

// Checks if a filter is currently active in asset or resource filters.
const isFilterActive = (filter) => {
  return commonStore.assetFilters.includes(filter) || commonStore.resourceFilters.includes(filter);
};

// Removes a filter from both asset and resource filters lists.
const removeFilter = (filter) => {
  commonStore.assetFilters = commonStore.assetFilters.filter((item) => item !== filter);
  commonStore.resourceFilters = commonStore.resourceFilters.filter((item) => item !== filter);
};

// Toggles a filter on or off and refreshes browser.
const toggleFilter = (filter) => {
  if (isFilterActive(filter)) {
    removeFilter(filter);
  } else {
    addFilter(filter);
  }
  emitter.emit('refresh-browser');
};

// lifecycle hooks
onMounted(() => {
  menu.assetMenuWidth = collectionMenu.value.getBoundingClientRect().width;
  menu.collectionMenu = collectionMenu.value;
});

onBeforeUnmount(() => {
  menu.assetMenuWidth = collectionMenu.value.getBoundingClientRect().width;
  menu.assetMenuHeight = collectionMenu.value.getBoundingClientRect().height;
});
</script>

<style scoped>

@import "@/assets/desktop.css";
@import "@/assets/menu.css";
</style>

