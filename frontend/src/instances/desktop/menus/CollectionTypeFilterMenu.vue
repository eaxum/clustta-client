<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span v-if="stage.activeStage === 'browser'" class="filter-menu-item" @click="toggleUseExclusive">
      <img class="small-icons" src="/icons/parameters.svg">
      <div class="horizontal-flex">
        <div> Use Exclusive </div>
        <ToggleSwitch :switchValueProp="useExclusive" />
      </div>
    </span>

    <span v-if="stage.activeStage === 'browser'" class="filter-menu-item" @click="toggleUseDeep">
      <img class="small-icons" src="/icons/deep.svg">
      <div class="horizontal-flex">
        <div> Deep </div>
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
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};


// imports
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import utils from '@/services/utils';

// states/store imports
import { useMenu } from '@/stores/menu';
import { useCommonStore } from '@/stores/common';
import { useStatusStore } from '@/stores/status';
import { useEntityStore } from '@/stores/entity';
import { useStageStore } from '@/stores/stages';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// states/stores
const menu = useMenu();
const commonStore = useCommonStore();
const statusStore = useStatusStore();
const entityStore = useEntityStore();
const stage = useStageStore();

// refs
const collectionMenu = ref(null);
const useExclusive = ref(false);
const useDeep = ref(false);


const entityTypes = computed(() => {
  console.log(entityStore.getEntityTypes)
  return entityStore.getEntityTypes
});


// methods
const isFilterActive = (filter) => {
  return commonStore.entityFilters.includes(filter);
};

const toggleUseExclusive = () => {
  useExclusive.value = !useExclusive.value;
  if (useExclusive.value) {
    commonStore.entityFilters = commonStore.entityFilters.filter((item) => item.type !== 'entity-type')
  }
};

const toggleUseDeep = () => {
  commonStore.useDeep = !commonStore.useDeep;
};

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

const addFilter = (filter) => {
  commonStore.entityFilters.push(filter);
};

const removeFilter = (filter) => {
  commonStore.entityFilters = commonStore.entityFilters.filter((item) => item !== filter)
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