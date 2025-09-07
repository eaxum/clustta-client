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
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import utils from '@/services/utils';

// states/store imports
import { useMenu } from '@/stores/menu';
import { useCommonStore } from '@/stores/common';
import { useAssetStore } from '@/stores/assets';
import { useTagStore } from '@/stores/tags';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// states/stores
const menu = useMenu();
const commonStore = useCommonStore();
const tagStore = useTagStore();
const assetStore = useAssetStore();

// refs
const collectionMenu = ref(null);

const viewTags = computed(() => {
  let tags = tagStore.tags;
		let viewTags = [];
		let filteredTaskResults = assetStore.getFilteredTasks;

		for (const task of filteredTaskResults){
			let taskTags = task.tags;
			for(let t = 0; t < taskTags.length; t++){
				if(!viewTags.includes(taskTags[t])){
					viewTags.push(taskTags[t])
				}
			}
		}

		for(let i = 0; i < tags.length; i++){
			tags[i].name = tags[i].name;
			tags[i].type = 'tags'
		}
		const availableTags = tags;
		const filteredTags = availableTags.filter((item) => viewTags.includes(item.name));
		return filteredTags
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


