<template>
	<nav class="sidebar-navigation-list" :aria-label="ariaLabel">
		<section v-for="group in visibleGroups" :key="group.id" class="sidebar-navigation-group">
			<div v-if="group.name" class="sidebar-navigation-group-title">{{ group.name }}</div>
			<div v-if="group.name" class="menu-divider"></div>
			<ActionButton
				v-for="page in group.pages"
				:key="page.id"
				:ref="(element) => setNavigationItem(page.id, element)"
				:icon="getAppIcon(page.icon)"
				:label="page.name"
				:show-label="true"
				:full-width="true"
				:is-active="page.id === activePageId"
				:button-function="() => selectPage(page.id)"
				:aria-current="page.id === activePageId ? 'page' : undefined"
				role="button"
				tabindex="0"
				@keydown="handleNavigationKeydown($event, page.id)"
			>
				<template v-if="page.badge" #trailing>
					<span class="sidebar-navigation-badge">{{ page.badge }}</span>
				</template>
			</ActionButton>
		</section>
	</nav>
</template>

<script setup>
import { computed, ref } from 'vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
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

const setNavigationItem = (pageId, component) => {
	if (component) {
		navigationItems.value[pageId] = component.$el;
		return;
	}
	delete navigationItems.value[pageId];
};

const handleNavigationKeydown = (event, pageId) => {
	if (['Enter', ' '].includes(event.key)) {
		event.preventDefault();
		selectPage(pageId);
		return;
	}

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
	gap: .1rem;
	padding: .5rem;
}

.sidebar-navigation-group-title {
	margin: .45rem .45rem .5rem;
	color: var(--text-muted);
	font-size: .85rem;
	font-weight: 450;
	line-height: 1.25;
}

.menu-divider {
	margin-bottom: .2rem;
	width: 100%;
}

.sidebar-navigation-badge {
	margin-left: auto;
	padding: .15rem .35rem;
	border-radius: 999px;
	color: var(--text-muted);
	background-color: var(--surface-2);
	font-size: .58rem;
	font-weight: 500;
	line-height: 1.2;
	white-space: nowrap;
}
</style>
