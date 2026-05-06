<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span :class="{ 'disabled' : commonStore.onlyAssets }" class="filter-menu-item" @click="toggleShowCollections()">
      <CiFolder class="small-icons" :size="20" />
      <div class="horizontal-flex">
        <div class="menu-item-text" >{{ $t('menus.collections') }}</div>
        <ToggleSwitch :switchValueProp="commonStore.showCollections" />
      </div>
    </span>

    <span :class="{ 'disabled' : commonStore.onlyAssets }" class="filter-menu-item" @click="toggleShowAssets()">
      <CiBrush class="small-icons" :size="20" />
      <div class="horizontal-flex">
        <div class="menu-item-text">{{ $t('menus.assets') }}</div>
        <ToggleSwitch :switchValueProp="commonStore.showAssets" />
      </div>
    </span>

    <span class="menu-divider"></span>

     <span :class="{ 'disabled' : !commonStore.showAssets }" v-if="!commonStore.navigatorMode && stage.activeStage === 'browser'" class="filter-menu-item" @click="toggleOnlyAssets()">
      <CiShapes class="small-icons" :size="20" />
      <div class="horizontal-flex">
        <div class="menu-item-text">{{ $t('menus.onlyProjectAssets') }}</div>
        <ToggleSwitch :switchValueProp="commonStore.onlyAssets" />
      </div>
    </span>

    <span :class="{ 'disabled' : !commonStore.showAssets }" class="filter-menu-item" @click="toggleShowResources()">
      <CiPaperclip class="small-icons" :size="20" />
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
import { CiBrush, CiFolder, CiPaperclip, CiShapes } from '@clustta/icons-vue';
import emitter from '@/lib/mitt';
import { useI18n } from 'vue-i18n';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// stores
import { useCommonStore } from '@/stores/common';
import { useMenu } from '@/stores/menu';
import { useStageStore } from '@/stores/stages';

const commonStore = useCommonStore();
const menu = useMenu();
const stage = useStageStore();

const { t } = useI18n();

// refs
const collectionMenu = ref(null);

// methods
// Toggles only assets filter and refreshes browser.
const toggleOnlyAssets = () => {
  commonStore.onlyAssets = !commonStore.onlyAssets;
  emitter.emit('refresh-browser');
};

// Toggles show collections filter and refreshes browser.
const toggleShowCollections = () => {
  commonStore.showCollections = !commonStore.showCollections;
  emitter.emit('refresh-browser');
};

// Toggles show resources filter and refreshes browser.
const toggleShowResources = () => {
  commonStore.showResources = !commonStore.showResources;
  emitter.emit('refresh-browser');
};

// Toggles show assets filter and refreshes browser.
const toggleShowAssets = () => {
  commonStore.showAssets = !commonStore.showAssets;
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

