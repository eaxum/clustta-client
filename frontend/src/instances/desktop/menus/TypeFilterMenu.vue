<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span :class="{ 'disabled' : commonStore.onlyAssets || commonStore.onlyCollections }" class="filter-menu-item" @click="toggleShowCollections()">
      <img class="small-icons" :src="getAppIcon('folder')">
      <div class="horizontal-flex">
        <div class="menu-item-text" >{{ $t('menus.collections') }}</div>
        <ToggleSwitch :switchValueProp="commonStore.showCollections" />
      </div>
    </span>

    <span :class="{ 'disabled' : commonStore.onlyAssets || commonStore.onlyCollections }" class="filter-menu-item" @click="toggleShowAssets()">
      <img class="small-icons" :src="getAppIcon('file')">
      <div class="horizontal-flex">
        <div class="menu-item-text">{{ $t('menus.assets') }}</div>
        <ToggleSwitch :switchValueProp="commonStore.showAssets" />
      </div>
    </span>

    <span class="menu-divider"></span>

    <span :class="{ 'disabled' : !commonStore.showAssets || commonStore.onlyCollections }" v-if="stage.activeStage === 'browser'" class="filter-menu-item" @click="toggleOnlyAssets()">
      <img class="small-icons" :src="getAppIcon('shapes')">
      <div class="horizontal-flex">
        <div class="menu-item-text">{{ commonStore.navigatorMode ? $t('menus.onlyAssets') : $t('menus.onlyProjectAssets') }}</div>
        <ToggleSwitch :switchValueProp="commonStore.onlyAssets" />
      </div>
    </span>

    <span :class="{ 'disabled' : !commonStore.showCollections || commonStore.onlyAssets }" v-if="stage.activeStage === 'browser'" class="filter-menu-item" @click="toggleOnlyCollections()">
      <img class="small-icons" :src="getAppIcon('folder')">
      <div class="horizontal-flex">
        <div class="menu-item-text">{{ commonStore.navigatorMode ? $t('menus.onlyCollections') : $t('menus.onlyProjectCollections') }}</div>
        <ToggleSwitch :switchValueProp="commonStore.onlyCollections" />
      </div>
    </span>

    <span :class="{ 'disabled' : !commonStore.showAssets || commonStore.onlyCollections }" class="filter-menu-item" @click="toggleShowTasks()">
      <img class="small-icons" :src="getAppIcon('kanban')">
      <div class="horizontal-flex">
        <div class="menu-item-text">{{ $t('menus.tasks') }}</div>
        <ToggleSwitch :switchValueProp="commonStore.showTasks" />
      </div>
    </span>

    <span :class="{ 'disabled' : !commonStore.showAssets || commonStore.onlyCollections }" class="filter-menu-item" @click="toggleShowResources()">
      <img class="small-icons" :src="getAppIcon('paperclip')">
      <div class="horizontal-flex">
        <div class="menu-item-text">{{ $t('menus.resources') }}</div>
        <ToggleSwitch :switchValueProp="commonStore.showResources" />
      </div>
    </span>

  </div>

</template>

<script setup>
// imports
import { onBeforeUnmount, onMounted, ref } from 'vue';
import emitter from '@/lib/mitt';
import { useI18n } from 'vue-i18n';

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

const { t } = useI18n();

// refs
const collectionMenu = ref(null);

// methods
// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Toggles only assets filter and refreshes browser.
const toggleOnlyAssets = () => {
  if (!commonStore.showAssets || commonStore.onlyCollections) return;
  commonStore.onlyAssets = !commonStore.onlyAssets;
  if (commonStore.onlyAssets) commonStore.onlyCollections = false;
  emitter.emit('refresh-browser');
};

// Toggles only collections filter and refreshes browser.
const toggleOnlyCollections = () => {
  if (!commonStore.showCollections || commonStore.onlyAssets) return;
  commonStore.onlyCollections = !commonStore.onlyCollections;
  if (commonStore.onlyCollections) commonStore.onlyAssets = false;
  emitter.emit('refresh-browser');
};

// Toggles show collections filter and refreshes browser.
const toggleShowCollections = () => {
  if (commonStore.onlyAssets || commonStore.onlyCollections) return;
  commonStore.showCollections = !commonStore.showCollections;
  if (!commonStore.showCollections) commonStore.onlyCollections = false;
  emitter.emit('refresh-browser');
};

// Toggles tracked task assets and refreshes browser.
const toggleShowTasks = () => {
  if (commonStore.onlyCollections) return;
  if (!commonStore.showAssets) return;
  commonStore.showTasks = !commonStore.showTasks;
  emitter.emit('refresh-browser');
};

// Toggles show resources filter and refreshes browser.
const toggleShowResources = () => {
  if (commonStore.onlyCollections) return;
  if (!commonStore.showAssets) return;
  commonStore.showResources = !commonStore.showResources;
  emitter.emit('refresh-browser');
};

// Toggles show assets filter and refreshes browser.
const toggleShowAssets = () => {
  if (commonStore.onlyAssets || commonStore.onlyCollections) return;
  commonStore.showAssets = !commonStore.showAssets;
  if (!commonStore.showAssets) commonStore.onlyAssets = false;
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

