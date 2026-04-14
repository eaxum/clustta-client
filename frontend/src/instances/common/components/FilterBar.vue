<template>
  <div ref="filterBarRoot" class="filter-bar-root">
  	<div ref="filterOptions" class="filter-options">
			<FilterButton v-if="!kanbanView" :icon="getAppIcon('clock')" v-tooltip="barIsOverflowing ? $t('components.filterBar.status') : ''" :label="$t('components.filterBar.status')"
				:alert="isFilterActive('status')" :showLabel="!barIsOverflowing" @mouseenter="flashFilterMenu($event, 'statusFilterMenu')"
				@click="showFilterMenu($event, 'statusFilterMenu')" />
			<FilterButton v-if="!kanbanView" :icon="getAppIcon('circle-check')" v-tooltip="barIsOverflowing ? $t('components.filterBar.state') : ''" :alert="isFilterActive('state')"
				:label="$t('components.filterBar.state')" :showLabel="!barIsOverflowing" @mouseenter="flashFilterMenu($event, 'stateFilterMenu')"
				@click="showFilterMenu($event, 'stateFilterMenu')" />
			<FilterButton :icon="getAppIcon('extension')" v-tooltip="barIsOverflowing ? $t('components.filterBar.extension') : ''" :alert="isFilterActive('extension')"
				:label="$t('components.filterBar.extension')" :showLabel="!barIsOverflowing" @mouseenter="flashFilterMenu($event, 'extensionFilterMenu')"
				@click="showFilterMenu($event, 'extensionFilterMenu')" />
			<FilterButton :icon="getAppIcon('man-running')" v-tooltip="barIsOverflowing ? $t('components.filterBar.assetType') : ''" :alert="isFilterActive('asset-type')"
				:label="$t('components.filterBar.assetType')" :showLabel="!barIsOverflowing" @mouseenter="flashFilterMenu($event, 'assetTypeFilterMenu')"
				@click="showFilterMenu($event, 'assetTypeFilterMenu')" />
			<FilterButton v-if="showTagsFilter" :icon="getAppIcon('tag')" v-tooltip="barIsOverflowing ? $t('components.filterBar.tags') : ''"
				:label="$t('components.filterBar.tags')" :alert="isFilterActive('tags')" :showLabel="!barIsOverflowing" @mouseenter="flashFilterMenu($event, 'tagsFilterMenu')"
				@click="showFilterMenu($event, 'tagsFilterMenu')" />
			<FilterButton :icon="getAppIcon('person')" v-tooltip="barIsOverflowing ? $t('components.filterBar.assignation') : ''" :label="$t('components.filterBar.assignees')"
				:alert="isFilterActive('assignation')" :showLabel="!barIsOverflowing" @mouseenter="flashFilterMenu($event, 'assigneeFilterMenu')"
				@click="showFilterMenu($event, 'assigneeFilterMenu')" />
			<FilterButton v-if="!kanbanView" :icon="getAppIcon('shapes')" v-tooltip="barIsOverflowing ? $t('components.filterBar.type') : ''" :alert="isFilterActive('general')"
			 :showLabel="!barIsOverflowing"	@mouseenter="flashFilterMenu($event, 'typeFilterMenu')"
				@click="showFilterMenu($event, 'typeFilterMenu')" />
			<ActionButton v-if="filtersActive" :icon="getAppIcon('close')" :allowDeactivate="true"
				v-tooltip="$t('components.filterBar.resetFilters')" :buttonFunction="clearFilters" />
		</div>
	</div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, nextTick, onBeforeUnmount } from 'vue';
import emitter from '@/lib/mitt';

//components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import FilterButton from '@/instances/desktop/components/FilterButton.vue'

//stores
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useTagStore } from '@/stores/tags';

// states
const commonStore = useCommonStore();
const iconStore = useIconStore();
const menu = useMenu();
const tagStore = useTagStore();

// refs
const filterBarRoot = ref(null);
const filterOptions = ref(null);
const barIsOverflowing = ref(true);
let resizeObserver = null;

// props
const props = defineProps({
	kanbanView: { type: Boolean, default: false },
});

// computed properties
const filtersActive = computed(() => {
	const assigneeFilters = commonStore.hasAssignees || commonStore.noAssignees;
	const collectionFilters = commonStore.collectionFilters.length > 0;
	const assetFilters = commonStore.assetFilters.length > 0;
	const resourceFilters = commonStore.resourceFilters.length > 0;
	const generalFilter = isFilterActive('general');
	return assigneeFilters || collectionFilters || assetFilters || resourceFilters || generalFilter;
});

const showTagsFilter = computed(() => !!tagStore.tags.length && (commonStore.showAssets || commonStore.showResources));

const emit = defineEmits(['selectCrumb']);

// methods

// Clears all filters and triggers a browser refresh.
const clearFilters = () => { commonStore.resetFilters(); emitter.emit('refresh-browser'); };

// Checks if a specific filter type is currently active.
const isFilterActive = (filter) => {
	if (filter.includes('general')) {
		const isActive = commonStore.showCollections && commonStore.showAssets && commonStore.showResources && commonStore.showChildCollections && commonStore.showChildAssets && commonStore.showDependencies && !commonStore.onlyAssets;
		return !isActive;
	} else if (filter.includes('collection')) return commonStore.collectionFilters.some((item) => item.type === filter);
	else if (filter.includes('assignation')) {
		const assigneeFilters = commonStore.hasAssignees || commonStore.noAssignees;
		return assigneeFilters || commonStore.assetFilters.some((item) => item.type === filter);
	} else return commonStore.assetFilters.some((item) => item.type === filter);
};

// Shows a filter menu for the selected filter button.
const flashFilterMenu = (event, menuName) => {
	if (menu.contextMenuVisible && !menu.nonFilterMenus.includes(menu.activeMenu)) {
		menu.showContextMenu(event, menuName, true, true);
	}
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Toggles or shows a filter menu on click.
const showFilterMenu = (event, menuName) => {
	if (menu.activeMenu === menuName && menu.contextMenuVisible) {
		menu.disableAllMenus();
		menu.activeMenu = null;
	} else {
		menu.showContextMenu(event, menuName, true, true);
	}
};

// Checks if the filter bar is overflowing.
const checkOverflow = async() => {
	return
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
};

// lifecycle hooks

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
	height: max-content;
	justify-content: flex-start;
	box-sizing: border-box;
	min-height: 30px;
	overflow: hidden;
	width: 100%;
	/* background-color: crimson; */
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




