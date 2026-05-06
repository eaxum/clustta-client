<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span v-for="assetType in projectAssetTypes" class="filter-menu-item" @click="toggleFilter(assetType)">
      <component :is="resolveIcon(assetType.icon)" class="small-icons" :size="20" />
      <div class="horizontal-flex">
        <div class="menu-item-text"> {{ utils.capitalizeStr(assetType.name) }} </div>
        <ToggleSwitch :switchValueProp="isFilterActive(assetType)" />
      </div>
    </span>

  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import emitter from '@/lib/mitt';
import { resolveIcon } from '@/lib/icon-map';
import utils from '@/services/utils';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

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
// Returns list of asset types available in the project.
const projectAssetTypes = computed(() => {
  return assetStore.getAssetTypes;
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
  commonStore.assetFilters = commonStore.assetFilters.filter((item) => item.id !== filter.id);
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




