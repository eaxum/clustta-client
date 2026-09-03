<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span v-for="status in allStatuses" class="filter-menu-item" @click="toggleFilter(status)">
      <img class="small-icons no-filter" :src="getStatusIcon(status)">
      <div class="horizontal-flex">
        <div class="menu-item-text"> {{ status.name.toUpperCase() }} </div>
        <CheckBox :modelValue="isFilterActive(status)" :ariaLabel="`Filter by ${status.name}`"
          @click.stop @change="toggleFilter(status)" />
      </div>
    </span>

  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import emitter from '@/lib/mitt';

// components
import CheckBox from '@/instances/common/components/CheckBox.vue';

// stores
import { useCommonStore } from '@/stores/common';
import { useMenu } from '@/stores/menu';
import { useStatusStore } from '@/stores/status';

const commonStore = useCommonStore();
const menu = useMenu();
const statusStore = useStatusStore();

// refs
const collectionMenu = ref(null);

// computed properties
// Returns list of statuses with formatted properties.
const allStatuses = computed(() => {
  let statuses = statusStore.statuses;
  for (let i = 0; i < statuses.length; i++) {
    statuses[i].name = statuses[i].short_name.toLowerCase();
    statuses[i].backgroundColor = statuses[i].color;
    statuses[i].type = 'status';
    statuses[i].textColor = 'black';
  }
  return statuses;
});

// methods
// Adds a filter to the asset filters list.
const addFilter = (filter) => {
  commonStore.assetFilters.push(filter);
};

// Returns the status icon path for a given status.
const getStatusIcon = (status) => {
  return '/status-icons/status_' + status.short_name + '.svg';
};

// Checks if a filter is currently active.
const isFilterActive = (filter) => {
  return commonStore.assetFilters.includes(filter);
};

// Removes a filter from the asset filters list.
const removeFilter = (filter) => {
  commonStore.assetFilters = commonStore.assetFilters.filter((item) => item !== filter);
};

// Toggles a filter on or off and refreshes browser.
const toggleFilter = (filter) => {
  if (commonStore.assetFilters.includes(filter)) {
    removeFilter(filter);
  } else {
    addFilter(filter);
  }
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

