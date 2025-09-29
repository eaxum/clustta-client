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
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

// imports
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';

// states/store imports
import { useMenu } from '@/stores/menu';
import { useCommonStore } from '@/stores/common';
import { useStageStore } from '@/stores/stages';
import emitter from '@/lib/mitt';
// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// states/stores
const menu = useMenu();
const commonStore = useCommonStore();
const stage = useStageStore();

// refs
const collectionMenu = ref(null);

// computed properties
const toggleShowEntities = () => {
  commonStore.filterDependencyCollections = !commonStore.filterDependencyCollections;
};

const toggleShowTasks = () => {
  commonStore.filterDependencyAssets = !commonStore.filterDependencyAssets;
};

const toggleShowResources = () => {
  commonStore.filterDependencyResources = !commonStore.filterDependencyResources;
};

const toggleOnlyAssets = () => {
  commonStore.onlyAssets = !commonStore.onlyAssets;
};

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

