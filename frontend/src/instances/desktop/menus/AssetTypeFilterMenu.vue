<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span v-for="taskType in projectTaskTypes" class="filter-menu-item" @click="toggleFilter(taskType)">
      <img class="small-icons" :src="getAppIcon(taskType.icon)">
      <div class="horizontal-flex">
        <div class="menu-item-text"> {{ utils.capitalizeStr(taskType.name) }} </div>
        <ToggleSwitch :switchValueProp="isFilterActive(taskType)" />
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
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';

const assetStore = useAssetStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const menu = useMenu();

// refs
const collectionMenu = ref(null);

// computed properties
// Returns list of asset types available in the project.
const projectTaskTypes = computed(() => {
  return assetStore.getAssetTypes;
});

// methods
// Adds a filter to the task filters list.
const addFilter = (filter) => {
  commonStore.taskFilters.push(filter);
};

// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Checks if a filter is currently active.
const isFilterActive = (filter) => {
  return commonStore.taskFilters.includes(filter);
};

// Removes a filter from the task filters list.
const removeFilter = (filter) => {
  commonStore.taskFilters = commonStore.taskFilters.filter((item) => item.id !== filter.id);
};

// Toggles a filter on or off and refreshes browser.
const toggleFilter = (filter) => {
  if (commonStore.taskFilters.includes(filter)) {
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




