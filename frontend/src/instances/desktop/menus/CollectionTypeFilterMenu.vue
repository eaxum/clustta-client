<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span v-if="stage.activeStage === 'browser'" class="filter-menu-item" @click="toggleUseExclusive">
      <img class="small-icons" src="/icons/parameters.svg">
      <div class="horizontal-flex">
        <div>{{ $t('menus.useExclusive') }}</div>
        <ToggleSwitch :switchValueProp="useExclusive" />
      </div>
    </span>

    <span v-if="stage.activeStage === 'browser'" class="filter-menu-item" @click="toggleUseDeep">
      <img class="small-icons" src="/icons/deep.svg">
      <div class="horizontal-flex">
        <div>{{ $t('menus.deep') }}</div>
        <ToggleSwitch :switchValueProp="commonStore.useDeep" />
      </div>
    </span>

    <span v-if="stage.activeStage === 'browser'" class="menu-divider"></span>

    <span v-for="collectionType in collectionTypes" class="filter-menu-item" @click="toggleFilter(collectionType)">
      <component :is="resolveIcon(collectionType.icon)" class="small-icons" :size="20" />
      <div class="horizontal-flex">
        <div class="menu-item-text"> {{ utils.capitalizeStr(collectionType.name) }} </div>
        <ToggleSwitch :switchValueProp="isFilterActive(collectionType)" />
      </div>
    </span>

  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { resolveIcon } from '@/lib/icon-map';
import utils from '@/services/utils';
import { useI18n } from 'vue-i18n';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// stores
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useMenu } from '@/stores/menu';
import { useStageStore } from '@/stores/stages';

const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const menu = useMenu();
const stage = useStageStore();

const { t } = useI18n();

// refs
const collectionMenu = ref(null);
const useExclusive = ref(false);

// computed properties
// Returns list of collection/collection types available in the project.
const collectionTypes = computed(() => {
  return collectionStore.getCollectionTypes;
});

// methods
// Adds a filter to the collection filters list.
const addFilter = (filter) => {
  commonStore.collectionFilters.push(filter);
};

// Checks if a filter is currently active.
const isFilterActive = (filter) => {
  return commonStore.collectionFilters.includes(filter);
};

// Removes a filter from the collection filters list.
const removeFilter = (filter) => {
  commonStore.collectionFilters = commonStore.collectionFilters.filter((item) => item !== filter);
};

// Toggles a filter on or off with exclusive mode support.
const toggleFilter = (filter) => {
  const existingFilter = commonStore.collectionFilters.find((item) => item.type = 'collection-type');

  if (commonStore.collectionFilters.includes(filter)) {
    removeFilter(filter);
  } else {
    addFilter(filter);
    if (useExclusive.value) {
      removeFilter(existingFilter);
    }
  }
};

// Toggles deep filtering mode.
const toggleUseDeep = () => {
  commonStore.useDeep = !commonStore.useDeep;
};

// Toggles exclusive filtering mode.
const toggleUseExclusive = () => {
  useExclusive.value = !useExclusive.value;
  if (useExclusive.value) {
    commonStore.collectionFilters = commonStore.collectionFilters.filter((item) => item.type !== 'collection-type');
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