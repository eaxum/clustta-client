<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span v-for="tag in viewTags" class="filter-menu-item" @click="toggleFilter(tag)">
      <img class="small-icons" src="/icons/tags.svg">
      <div class="horizontal-flex">
        <div> {{ utils.capitalizeStr(tag.name) }} </div>
        <ToggleSwitch :switchValueProp="isFilterActive(tag)" />
      </div>
    </span>

  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import utils from '@/services/utils';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCommonStore } from '@/stores/common';
import { useMenu } from '@/stores/menu';
import { useTagStore } from '@/stores/tags';

const assetStore = useAssetStore();
const commonStore = useCommonStore();
const menu = useMenu();
const tagStore = useTagStore();

// refs
const collectionMenu = ref(null);

// computed properties
// Returns list of tags that are used by filtered assets.
const viewTags = computed(() => {
  let tags = tagStore.tags;
  let viewTags = [];
  let filteredAssetResults = assetStore.getFilteredAssets;

  for (const asset of filteredAssetResults) {
    let assetTags = asset.tags;
    for (let t = 0; t < assetTags.length; t++) {
      if (!viewTags.includes(assetTags[t])) {
        viewTags.push(assetTags[t]);
      }
    }
  }

  for (let i = 0; i < tags.length; i++) {
    tags[i].name = tags[i].name;
    tags[i].type = 'tags';
  }
  const availableTags = tags;
  const filteredTags = availableTags.filter((item) => viewTags.includes(item.name));
  return filteredTags;
});

// methods
// Adds a filter to the asset filters list.
const addFilter = (filter) => {
  commonStore.assetFilters.push(filter);
};

// Checks if a filter is currently active.
const isFilterActive = (filter) => {
  return commonStore.assetFilters.includes(filter);
};

// Removes a filter from the asset filters list.
const removeFilter = (filter) => {
  commonStore.assetFilters = commonStore.assetFilters.filter((item) => item !== filter);
};

// Toggles a filter on or off.
const toggleFilter = (filter) => {
  if (commonStore.assetFilters.includes(filter)) {
    removeFilter(filter);
  } else {
    addFilter(filter);
  }
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




