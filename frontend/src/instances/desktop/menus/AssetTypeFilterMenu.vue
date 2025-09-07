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
import { useAssetStore } from '@/stores/assets';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// states/stores
const menu = useMenu();
const commonStore = useCommonStore();
const assetStore = useAssetStore();

// refs
const collectionMenu = ref(null);

const projectTaskTypes = computed(() => {
  console.log(assetStore.getTaskTypes)
  return assetStore.getTaskTypes
});

// methods
const isFilterActive = (filter) => {
  return commonStore.taskFilters.includes((filter));
};

const toggleFilter = (filter) => {
  if (commonStore.taskFilters.includes(filter)) {
    removeFilter(filter);
  } else {
    addFilter(filter);
  }
  emitter.emit('refresh-browser');
};

const addFilter = (filter) => {
  commonStore.taskFilters.push(filter);
};

const removeFilter = (filter) => {
  commonStore.taskFilters = commonStore.taskFilters.filter((item) => item.id !== filter.id)
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


