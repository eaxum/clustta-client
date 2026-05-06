<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span class="filter-menu-item" @click="toggleShowCollections()">
      <CiFolder class="small-icons" :size="20" />
      <div class="horizontal-flex">
        <div class="menu-item-text" >{{ $t('menus.collections') }}</div>
        <ToggleSwitch :switchValueProp="commonStore.filterDependencyCollections" />
      </div>
    </span>

    <span class="filter-menu-item" @click="toggleShowAssets()">
      <CiBrush class="small-icons" :size="20" />
      <div class="horizontal-flex">
        <div class="menu-item-text">{{ $t('menus.assets') }}</div>
        <ToggleSwitch :switchValueProp="commonStore.filterDependencyAssets" />
      </div>
    </span>


    <span class="filter-menu-item" @click="toggleShowResources()">
      <CiPaperclip class="small-icons" :size="20" />
      <div class="horizontal-flex">
        <div class="menu-item-text">{{ $t('menus.resources') }}</div>
        <ToggleSwitch :switchValueProp="commonStore.filterDependencyResources" />
      </div>
    </span>

  </div>

</template>

<script setup>
// imports
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { CiBrush, CiFolder, CiPaperclip } from '@clustta/icons-vue';
import { useI18n } from 'vue-i18n';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// stores
import { useCommonStore } from '@/stores/common';
import { useMenu } from '@/stores/menu';

const commonStore = useCommonStore();
const menu = useMenu();

const { t } = useI18n();

// refs
const collectionMenu = ref(null);

// methods
// Toggles the filter for showing collections in dependency search.
const toggleShowCollections = () => {
  commonStore.filterDependencyCollections = !commonStore.filterDependencyCollections;
};

// Toggles the filter for showing resources in dependency search.
const toggleShowResources = () => {
  commonStore.filterDependencyResources = !commonStore.filterDependencyResources;
};

// Toggles the filter for showing assets in dependency search.
const toggleShowAssets = () => {
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

