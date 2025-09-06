<template>
  <div ref="filterBarRoot" class="filter-bar-root">
  	<div ref="filterOptions" class="filter-options">
			<FilterButton v-if="!kanbanView" :icon="getAppIcon('clock')" v-tooltip="barIsOverflowing ? 'Status' : ''" label='Status'
				:alert="isFilterActive('status')" :showLabel="!barIsOverflowing" @mouseenter="flashFilterMenu($event, 'statusFilterMenu')"
				@click="showFilterMenu($event, 'statusFilterMenu')" />
			<FilterButton v-if="!kanbanView" :icon="getAppIcon('circle-check')" v-tooltip="barIsOverflowing ? 'State' : ''" :alert="isFilterActive('state')"
				label='State' :showLabel="!barIsOverflowing" @mouseenter="flashFilterMenu($event, 'stateFilterMenu')"
				@click="showFilterMenu($event, 'stateFilterMenu')" />
			<FilterButton :icon="getAppIcon('extension')" v-tooltip="barIsOverflowing ? 'Extension' : ''" :alert="isFilterActive('extension')"
				label='Extension' :showLabel="!barIsOverflowing" @mouseenter="flashFilterMenu($event, 'extensionFilterMenu')"
				@click="showFilterMenu($event, 'extensionFilterMenu')" />
			<FilterButton :icon="getAppIcon('man-running')" v-tooltip="barIsOverflowing ? 'Asset Type' : ''" :alert="isFilterActive('asset-type')"
				label='Asset Type' :showLabel="!barIsOverflowing" @mouseenter="flashFilterMenu($event, 'assetTypeFilterMenu')"
				@click="showFilterMenu($event, 'assetTypeFilterMenu')" />
			<FilterButton v-if="showTagsFilter && viewTags.length" :icon="getAppIcon('tag')" v-tooltip="barIsOverflowing ? 'Tags' : ''"
				label='Tags' :alert="isFilterActive('tags')" :showLabel="!barIsOverflowing" @mouseenter="flashFilterMenu($event, 'tagsFilterMenu')"
				@click="showFilterMenu($event, 'tagsFilterMenu')" />
			<FilterButton :icon="getAppIcon('person')" v-tooltip="barIsOverflowing ? 'Assignation' : ''" label='Assignees'
				:alert="isFilterActive('assignation')" :showLabel="!barIsOverflowing" @mouseenter="flashFilterMenu($event, 'assigneeFilterMenu')"
				@click="showFilterMenu($event, 'assigneeFilterMenu')" />
			<FilterButton v-if="!kanbanView" :icon="getAppIcon('shapes')" v-tooltip="barIsOverflowing ? 'Type' : ''" :alert="isFilterActive('general')"
			 :showLabel="!barIsOverflowing"	@mouseenter="flashFilterMenu($event, 'typeFilterMenu')"
				@click="showFilterMenu($event, 'typeFilterMenu')" />
			<ActionButton v-if="filtersActive" :icon="getAppIcon('close')" :allowDeactivate="true"
				v-tooltip="'Reset Filters'" :buttonFunction="clearFilters" />
		</div>
	</div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, nextTick, onBeforeUnmount } from 'vue';

//components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import FilterButton from '@/instances/desktop/components/FilterButton.vue'

//stores
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';

// states
const iconStore = useIconStore();
const menu = useMenu();

// refs
const filterBarRoot = ref(null);
const filterOptions = ref(null);
const barIsOverflowing = ref(false);
let resizeObserver = null;

// computed properties

// props
const props = defineProps({
  filtersActive: { type: Boolean, default: true },
  kanbanView: { type: Boolean, default: false },
  showTagsFilter: { type: Boolean, default: true },
  viewTags: { type: Array, default: () => [] },
  isFilterActive: {
	type: Function,
	default: () => {
	  return () => false;
	}
  },
  clearFilters: {
	type: Function,
	default: () => {
	  return () => {};
	}
  },
});

const emit = defineEmits([ 'selectCrumb'])
// methods

const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

const showFilterMenu = (event, menuName) => {
	if (menu.activeMenu === menuName && menu.contextMenuVisible) {
		menu.disableAllMenus();
		menu.activeMenu = null;
	} else {
		menu.showContextMenu(event, menuName, true, true);
	}
};

const flashFilterMenu = (event, menuName) => {
	if (menu.contextMenuVisible && !menu.nonFilterMenus.includes(menu.activeMenu)) {
		menu.showContextMenu(event, menuName, true, true);
	}
};

const checkOverflow = async() => {
	await nextTick();
	if (filterBarRoot.value && filterOptions.value) {
		const filterBarWidth = filterBarRoot.value.getBoundingClientRect().width;
		const filterOptionsWidth = filterOptions.value.clientWidth;
		barIsOverflowing.value = filterBarWidth <= filterOptionsWidth;
		const isOverflowing = filterBarWidth <= filterOptionsWidth;
		if (isOverflowing) {
			// filterOptions.value.style.overflowX = 'auto';
		} else {
			// filterOptions.value.style.overflowX = 'hidden';
		}
	}
	// console.log(filterBarRoot.value.getBoundingClientRect().width)
	// console.log(filterOptions.value.clientWidth)
};


// onMounted hook
onMounted(() => {
	if (filterBarRoot.value) {
    resizeObserver = new ResizeObserver(() => {
      checkOverflow()
    })
    resizeObserver.observe(filterBarRoot.value)
  }
});

onBeforeUnmount(() => {
  if (resizeObserver) {
    resizeObserver.disconnect()
  }
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.filter-bar-root {
	display: flex;
	align-items: center;
	padding: .2rem;
	height: max-content;
	justify-content: flex-start;
	box-sizing: border-box;
	min-height: 35px;
	overflow: hidden;
	/* overflow-x: scroll; */
	width: 100%;
}

.filter-options {
	display: flex;
	gap: .4rem;
	align-items: center;
	/* padding: .2rem 0; */
	height: max-content;
	justify-content: flex-end;
	box-sizing: border-box;
	width: min-content;
	/* width: min-content; */
	/* width: 100%; */
	/* min-height: 35px; */
	/* min-width: min-content; */
	/* flex: 0 0 auto; */
	/* background-color: crimson; */
}
</style>


