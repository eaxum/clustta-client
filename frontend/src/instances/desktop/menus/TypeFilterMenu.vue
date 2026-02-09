<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span :class="{ 'disabled' : commonStore.onlyAssets }" class="filter-menu-item" @click="toggleShowEntities()">
      <img class="small-icons" :src="getAppIcon('folder')">
      <div class="horizontal-flex">
        <div class="menu-item-text" >Collections</div>
        <ToggleSwitch :switchValueProp="commonStore.showEntities" />
      </div>
    </span>

    <span :class="{ 'disabled' : commonStore.onlyAssets }" class="filter-menu-item" @click="toggleShowTasks()">
      <img class="small-icons" :src="getAppIcon('brush')">
      <div class="horizontal-flex">
        <div class="menu-item-text">Assets </div>
        <ToggleSwitch :switchValueProp="commonStore.showTasks" />
      </div>
    </span>

    <span class="menu-divider"></span>

     <span :class="{ 'disabled' : !commonStore.showTasks }" v-if="!commonStore.navigatorMode && stage.activeStage === 'browser'" class="filter-menu-item" @click="toggleOnlyAssets()">
      <img class="small-icons" :src="getAppIcon('shapes')">
      <div class="horizontal-flex">
        <div class="menu-item-text">Only Project Assets</div>
        <ToggleSwitch :switchValueProp="commonStore.onlyAssets" />
      </div>
    </span>

    <span :class="{ 'disabled' : !commonStore.showTasks }" class="filter-menu-item" @click="toggleShowResources()">
      <img class="small-icons" :src="getAppIcon('paperclip')">
      <div class="horizontal-flex">
        <div class="menu-item-text">Resources</div>
        <ToggleSwitch :switchValueProp="commonStore.showResources" />
      </div>
    </span>

  </div>

</template>

<script setup>
// imports
import { onBeforeUnmount, onMounted, ref } from 'vue';
import emitter from '@/lib/mitt';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// stores
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useStageStore } from '@/stores/stages';

const commonStore = useCommonStore();
const iconStore = useIconStore();
const menu = useMenu();
const stage = useStageStore();

// refs
const collectionMenu = ref(null);

// methods
// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Toggles only assets filter and refreshes browser.
const toggleOnlyAssets = () => {
  commonStore.onlyAssets = !commonStore.onlyAssets;
  emitter.emit('refresh-browser');
};

// Toggles show entities filter and refreshes browser.
const toggleShowEntities = () => {
  commonStore.showEntities = !commonStore.showEntities;
  emitter.emit('refresh-browser');
};

// Toggles show resources filter and refreshes browser.
const toggleShowResources = () => {
  commonStore.showResources = !commonStore.showResources;
  emitter.emit('refresh-browser');
};

// Toggles show tasks filter and refreshes browser.
const toggleShowTasks = () => {
  commonStore.showTasks = !commonStore.showTasks;
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

