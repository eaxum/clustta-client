<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span v-for="tag in tagStore.tags" class="filter-menu-item" @click="toggleFilter(tag)">
      <img class="small-icons" src="/icons/tags.svg">
      <div class="horizontal-flex">
        <div> {{ utils.capitalizeStr(tag.name) }} </div>
        <ToggleSwitch :switchValueProp="isFilterActive(tag)" />
      </div>
    </span>

  </div>

</template>

<script setup>
// imports
import { onBeforeUnmount, onMounted, ref } from 'vue';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// stores
import { useCommonStore } from '@/stores/common';
import { useMenu } from '@/stores/menu';
import { useTagStore } from '@/stores/tags';

const commonStore = useCommonStore();
const menu = useMenu();
const tagStore = useTagStore();

// refs
const collectionMenu = ref(null);

// methods
// Adds a filter to the asset filters list.
const addFilter = (filter) => {
  commonStore.assetFilters.push(filter);
};

// Checks if a filter is currently active by name.
const isFilterActive = (filter) => {
  return commonStore.assetFilters.some((f) => f.type === 'tags' && f.name === filter.name);
};

// Removes a filter from the asset filters list by name.
const removeFilter = (filter) => {
  commonStore.assetFilters = commonStore.assetFilters.filter((f) => !(f.type === 'tags' && f.name === filter.name));
};

// Toggles a filter on or off.
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




