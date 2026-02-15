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

    <span v-for="entityType in entityTypes" class="filter-menu-item" @click="toggleFilter(entityType)">
      <img class="small-icons" :src="getAppIcon(entityType.icon)">
      <div class="horizontal-flex">
        <div class="menu-item-text"> {{ utils.capitalizeStr(entityType.name) }} </div>
        <ToggleSwitch :switchValueProp="isFilterActive(entityType)" />
      </div>
    </span>

  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import utils from '@/services/utils';
import { useI18n } from 'vue-i18n';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// stores
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useStageStore } from '@/stores/stages';

const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const menu = useMenu();
const stage = useStageStore();

const { t } = useI18n();

// refs
const collectionMenu = ref(null);
const useExclusive = ref(false);

// computed properties
// Returns list of collection/entity types available in the project.
const entityTypes = computed(() => {
  return collectionStore.getCollectionTypes;
});

// methods
// Adds a filter to the entity filters list.
const addFilter = (filter) => {
  commonStore.entityFilters.push(filter);
};

// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Checks if a filter is currently active.
const isFilterActive = (filter) => {
  return commonStore.entityFilters.includes(filter);
};

// Removes a filter from the entity filters list.
const removeFilter = (filter) => {
  commonStore.entityFilters = commonStore.entityFilters.filter((item) => item !== filter);
};

// Toggles a filter on or off with exclusive mode support.
const toggleFilter = (filter) => {
  const existingFilter = commonStore.entityFilters.find((item) => item.type = 'entity-type');

  if (commonStore.entityFilters.includes(filter)) {
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
    commonStore.entityFilters = commonStore.entityFilters.filter((item) => item.type !== 'entity-type');
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