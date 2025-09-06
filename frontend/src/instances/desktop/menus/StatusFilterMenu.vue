<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span v-for="status in allStatuses" class="filter-menu-item" @click="toggleFilter(status)">
      <img class="small-icons no-filter" :src="getStatusIcon(status)">
      <div class="horizontal-flex">
        <div class="menu-item-text"> {{ status.name.toUpperCase() }} </div>
        <ToggleSwitch :switchValueProp="isFilterActive(status)" />
      </div>
    </span>

  </div>

</template>

<script setup>
// imports
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import utils from '@/services/utils';

// states/store imports
import { useMenu } from '@/stores/menu';
import { useCommonStore } from '@/stores/common';
import { useStatusStore } from '@/stores/status';
import emitter from '@/lib/mitt';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// states/stores
const menu = useMenu();
const commonStore = useCommonStore();
const statusStore = useStatusStore();

// refs
const collectionMenu = ref(null);

const allStatuses = computed(() => {
  let statuses = statusStore.statuses;
    for(let i = 0; i < statuses.length; i++){
        statuses[i].name = statuses[i].short_name.toLowerCase();
        statuses[i].backgroundColor = statuses[i].color
        statuses[i].type = 'status'
        statuses[i].textColor = 'black'
    }
    return statuses
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

const getStatusIcon = (status) => {
  return '/status-icons/status_' + status.short_name + '.svg'; 
}

const addFilter = (filter) => {
	commonStore.taskFilters.push(filter);
};

const removeFilter = (filter) => {
	commonStore.taskFilters = commonStore.taskFilters.filter((item) => item !== filter )
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