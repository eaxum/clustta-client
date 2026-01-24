<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span class="filter-menu-item" @click="toggleShowEntities()">
      <img class="small-icons" :src="getAppIcon('folder')">
      <div class="horizontal-flex">
        <div class="menu-item-text" >Collections</div>
        <ToggleSwitch :switchValueProp="commonStore.filterDependencyCollections" />
      </div>
    </span>

    <span class="filter-menu-item" @click="toggleShowTasks()">
      <img class="small-icons" :src="getAppIcon('brush')">
      <div class="horizontal-flex">
        <div class="menu-item-text">Assets </div>
        <ToggleSwitch :switchValueProp="commonStore.filterDependencyAssets" />
      </div>
    </span>


    <span class="filter-menu-item" @click="toggleShowResources()">
      <img class="small-icons" :src="getAppIcon('paperclip')">
      <div class="horizontal-flex">
        <div class="menu-item-text">Resources</div>
        <ToggleSwitch :switchValueProp="commonStore.filterDependencyResources" />
      </div>
    </span>

  </div>

</template>

<script setup>
// imports
import { onBeforeUnmount, onMounted, ref } from 'vue';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// stores
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';

const commonStore = useCommonStore();
const iconStore = useIconStore();
const menu = useMenu();

// refs
const collectionMenu = ref(null);

// methods
// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Toggles the filter for showing collections in dependency search.
const toggleShowEntities = () => {
  commonStore.filterDependencyCollections = !commonStore.filterDependencyCollections;
};

// Toggles the filter for showing resources in dependency search.
const toggleShowResources = () => {
  commonStore.filterDependencyResources = !commonStore.filterDependencyResources;
};

// Toggles the filter for showing assets in dependency search.
const toggleShowTasks = () => {
  commonStore.filterDependencyAssets = !commonStore.filterDependencyAssets;
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

