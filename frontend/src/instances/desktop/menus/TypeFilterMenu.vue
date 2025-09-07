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
  commonStore.showEntities = !commonStore.showEntities;
  emitter.emit('refresh-browser');
};

const toggleShowTasks = () => {
  commonStore.showTasks = !commonStore.showTasks;
  emitter.emit('refresh-browser');
};

const toggleShowResources = () => {
  commonStore.showResources = !commonStore.showResources;
  emitter.emit('refresh-browser');
};

const toggleOnlyAssets = () => {
  commonStore.onlyAssets = !commonStore.onlyAssets;
  emitter.emit('refresh-browser');
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

