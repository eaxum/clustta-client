<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span v-for="extension in allExtensions" class="filter-menu-item" @click="toggleFilter(extension)">
      <img class="small-icons no-filter" :src="extension.icon">
      <div class="horizontal-flex">
        <div class="menu-item-text" > {{ extension?.name?.toUpperCase()}} </div>
        <CheckBox :modelValue="isFilterActive(extension)" :ariaLabel="`Filter by ${extension?.name?.toUpperCase()}`"
          @click.stop @change="toggleFilter(extension)" />
      </div>
    </span>

  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import emitter from '@/lib/mitt';

// components
import CheckBox from '@/instances/common/components/CheckBox.vue';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCommonStore } from '@/stores/common';
import { useMenu } from '@/stores/menu';

const assetStore = useAssetStore();
const commonStore = useCommonStore();
const menu = useMenu();

// refs
const collectionMenu = ref(null);

// computed properties
// Returns list of file extensions used in the project.
const allExtensions = computed(() => {
  return assetStore.projectExtensions;
});

// methods
// Adds a filter to the asset filters list.
const addFilter = (filter) => {
  commonStore.assetFilters.push(filter);
};

// Checks if a filter is currently active.
const isFilterActive = (filter) => {
  return commonStore.assetFilters.includes(filter);
};

// Removes a filter from the asset filters list.
const removeFilter = (filter) => {
  commonStore.assetFilters = commonStore.assetFilters.filter((item) => item !== filter);
};

// Toggles a filter on or off and refreshes browser.
const toggleFilter = (filter) => {
  if (commonStore.assetFilters.includes(filter)) {
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


