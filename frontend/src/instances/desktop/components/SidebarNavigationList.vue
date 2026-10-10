<template>
	<nav class="sidebar-navigation-list" :aria-label="ariaLabel">
		<section v-for="group in visibleGroups" :key="group.id" class="sidebar-navigation-group">
			<h2 v-if="group.name" class="sidebar-navigation-group-title">{{ group.name }}</h2>
			<button
				v-for="page in group.pages"
				:key="page.id"
				:ref="(element) => setNavigationItem(page.id, element)"
				class="sidebar-navigation-item"
				:class="{ 'sidebar-navigation-item-selected': page.id === activePageId }"
				:aria-current="page.id === activePageId ? 'page' : undefined"
				type="button"
				@click="selectPage(page.id)"
				@keydown="handleNavigationKeydown($event, page.id)"
			>
				<img class="small-icons sidebar-navigation-icon" :src="getAppIcon(page.icon)" alt="">
				<span>{{ page.name }}</span>
			</button>
		</section>
	</nav>
</template>

<script setup>
import { computed, ref } from 'vue';
import { useIconStore } from '@/stores/icons';

const emit = defineEmits(['select']);

const props = defineProps({
	activePageId: {
		type: String,
		default: '',
	},
	ariaLabel: {
		type: String,
		required: true,
	},
	groups: {
		type: Array,
		default: () => [],
	},
});

const iconStore = useIconStore();
const navigationItems = ref({});

const visibleGroups = computed(() => props.groups.filter((group) => group.pages.length));
const orderedPageIds = computed(() => {
	return visibleGroups.value.flatMap((group) => group.pages.map((page) => page.id));
});

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);
const selectPage = (pageId) => emit('select', pageId);

const setNavigationItem = (pageId, element) => {
	if (element) {
		navigationItems.value[pageId] = element;
		return;
	}
	delete navigationItems.value[pageId];
};

const handleNavigationKeydown = (event, pageId) => {
	if (!['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(event.key)) return;
	event.preventDefault();

	const currentIndex = orderedPageIds.value.indexOf(pageId);
	if (currentIndex === -1) return;

	let targetIndex = currentIndex;
	if (event.key === 'ArrowDown') targetIndex = (currentIndex + 1) % orderedPageIds.value.length;
	if (event.key === 'ArrowUp') targetIndex = (currentIndex - 1 + orderedPageIds.value.length) % orderedPageIds.value.length;
	if (event.key === 'Home') targetIndex = 0;
	if (event.key === 'End') targetIndex = orderedPageIds.value.length - 1;

	const targetPageId = orderedPageIds.value[targetIndex];
	navigationItems.value[targetPageId]?.focus();
	selectPage(targetPageId);
};
</script>

<style scoped>
.sidebar-navigation-list {
	display: flex;
	flex-direction: column;
	gap: 1rem;
	min-height: 0;
	overflow-x: hidden;
	overflow-y: auto;
	padding: 0 .2rem .75rem 0;
	scrollbar-color: var(--surface-4) transparent;
	scrollbar-width: thin;
}

.sidebar-navigation-list::-webkit-scrollbar {
	width: 6px;
}

.sidebar-navigation-list::-webkit-scrollbar-track {
	background-color: transparent;
}

.sidebar-navigation-list::-webkit-scrollbar-thumb {
	min-height: 36px;
	border-radius: 999px;
	background-color: var(--surface-4);
}

.sidebar-navigation-group {
	display: flex;
	flex-direction: column;
	gap: .2rem;
}

.sidebar-navigation-group-title {
	margin: 0 .55rem .25rem;
	color: var(--text-muted);
	font-size: .72rem;
	font-weight: 500;
	line-height: 1.4;
}

.sidebar-navigation-item {
	display: flex;
	align-items: center;
	width: 100%;
	min-height: 36px;
	gap: .65rem;
	padding: .45rem .6rem;
	border: 0;
	border-radius: var(--large-radius);
	color: var(--text-muted);
	background-color: transparent;
	font: inherit;
	font-size: .82rem;
	font-weight: 400;
	text-align: left;
	cursor: pointer;
	box-sizing: border-box;
	transition:
		background-color 150ms ease,
		border-radius 150ms ease,
		color 150ms ease;
}

.sidebar-navigation-item:hover {
	color: var(--text);
	background-color: var(--surface-3);
	outline: var(--transparent-line);
	outline-offset: -1px;
}

.sidebar-navigation-item:focus-visible {
	outline: 2px solid var(--accent);
	outline-offset: -2px;
}

.sidebar-navigation-item-selected {
	color: var(--text);
	background-color: var(--surface-4);
	font-weight: 500;
}

.sidebar-navigation-item-selected:hover {
	border-radius: var(--small-radius);
}

.sidebar-navigation-icon {
	flex: 0 0 auto;
	opacity: .8;
}

.sidebar-navigation-item-selected .sidebar-navigation-icon {
	opacity: 1;
}
</style>
