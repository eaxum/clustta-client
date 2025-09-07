<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span v-for="extension in allExtensions" class="filter-menu-item" @click="toggleFilter(extension)">
      <img class="small-icons no-filter" :src="extension.icon">
      <div class="horizontal-flex">
        <div class="menu-item-text" > {{ extension.name.toUpperCase()}} </div>
        <ToggleSwitch :switchValueProp="isFilterActive(extension)" />
      </div>
    </span>

  </div>

</template>

<script setup>
// imports
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// states/store imports
import { useMenu } from '@/stores/menu';
import { useCommonStore } from '@/stores/common';
import { useTemplateStore } from '@/stores/template';
import { useAssetStore } from '@/stores/assets';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// states/stores
const menu = useMenu();
const commonStore = useCommonStore();
const templateStore = useTemplateStore();
const assetStore = useAssetStore();

// refs
const collectionMenu = ref(null);

const allExtensions = computed(() => {
  let templates = templateStore.getTemplates;
  return assetStore.projectExtensions

});

// methods
const isFilterActive = (filter) => {
    return commonStore.taskFilters.includes(filter);
};

const toggleFilter = (filter) => {
    if(commonStore.taskFilters.includes(filter)){
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
  commonStore.taskFilters = commonStore.taskFilters.filter((item) => item !== filter )
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


