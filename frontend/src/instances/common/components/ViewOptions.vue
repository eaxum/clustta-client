<template>
  	<div class="view-options-root">
        <ActionButton :icon="getAppIcon('list-compact')" v-tooltip="'Compact'" :buttonFunction="setCompactView" :color="isCompactActive ? 'var(--steel)' : undefined" />
        <!-- <ActionButton :icon="getAppIcon('list')" v-tooltip="'Larger'" :buttonFunction="setLargerView" :color="isLargerActive ? 'var(--steel)' : undefined" /> -->
        <ActionButton :icon="getAppIcon('four-squares')" v-tooltip="'Grid'" :buttonFunction="setGridView" :color="isGridActive ? 'var(--steel)' : undefined" />
    </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, nextTick, onBeforeUnmount } from 'vue';

//components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'

//stores
import { useIconStore } from '@/stores/icons';
import { useCommonStore } from '@/stores/common';

// states
const commonStore = useCommonStore();
const iconStore = useIconStore();

// refs

// computed properties
const isCompactActive = computed(() => commonStore.viewMode === 'compact');
const isLargerActive = computed(() => commonStore.viewMode === 'larger');
const isGridActive = computed(() => commonStore.viewMode === 'grid');

// props
const props = defineProps({
  kanbanView: { type: Boolean, default: false },
});

const emit = defineEmits([ 'selectCrumb'])

// methods

const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

const setCompactView = () => {
	commonStore.setCompactView();
};

const setLargerView = () => {
	commonStore.setLargerView();
};

const setGridView = () => {
	commonStore.setGridView();
};

// onMounted hook
onMounted(() => {

});

onBeforeUnmount(() => {
    
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.view-options-root {
	display: flex;
	gap: .1rem;
	align-items: center;
	padding: .1rem;
	justify-content: flex-end;
	box-sizing: border-box;
	width: min-content;
    height: min-content;
    background-color: var(--black-steel);
    border-radius: 10px;
    overflow: hidden;
}
</style>


