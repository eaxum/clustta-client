<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span v-for="state in allStates" class="filter-menu-item" @click="toggleFilter(state)">
      <img class="small-icons" :class="{ 'no-filter' : isColored(state?.name)}" :src="getAppIcon(state.icon)">
      <div class="horizontal-flex">
        <div class="menu-item-text"> {{ utils.capitalizeStr(state.name) }} </div>
        <ToggleSwitch :switchValueProp="isFilterActive(state)" />
      </div>
    </span>

  </div>

</template>

<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
    const icon = iconStore.getAppIcon(iconName);
    return icon
};

// imports
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// states/store imports
import { useMenu } from '@/stores/menu';
import { useCommonStore } from '@/stores/common';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// states/stores
const menu = useMenu();
const commonStore = useCommonStore();

// refs
const collectionMenu = ref(null);

const allStates = computed(() => {
  let fileStates = commonStore.fileStates;
  return fileStates;
});

// methods
const isFilterActive = (filter) => {
  const isActive = commonStore.taskFilters.includes(filter) || commonStore.resourceFilters.includes(filter)
  return isActive;
};

const toggleFilter = (filter) => {
  if (isFilterActive(filter)) {
    removeFilter(filter);
  } else {
    addFilter(filter);
  }
  emitter.emit('refresh-browser');
};

const isColored = (stateName) => {
  const coloredItems = ['modified', 'outdated']
  return coloredItems.includes(stateName) 
};

const addFilter = (filter) => {
  commonStore.taskFilters.push(filter);
  commonStore.resourceFilters.push(filter);
};

const removeFilter = (filter) => {
  commonStore.taskFilters = commonStore.taskFilters.filter((item) => item !== filter)
  commonStore.resourceFilters = commonStore.resourceFilters.filter((item) => item !== filter)
};



// methods

// onMounted hook
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

