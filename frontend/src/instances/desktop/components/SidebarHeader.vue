<template>
	<header class="sidebar-header">
		<ActionButton
			:icon="getAppIcon('chevron-left')"
			:buttonFunction="backFunction"
			v-tooltip="$t('components.headerBar.back')"
		/>
		<div class="sidebar-header-title">{{ title }}</div>
		<div class="sidebar-header-actions">
			<slot />
		</div>
	</header>
</template>

<script setup>
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import { useIconStore } from '@/stores/icons';

defineProps({
	backFunction: {
		type: Function,
		required: true,
	},
	title: {
		type: String,
		required: true,
	},
});

const iconStore = useIconStore();
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);
</script>

<style scoped>
.sidebar-header {
	display: grid;
	grid-template-columns: min-content minmax(0, 1fr) min-content;
	align-items: center;
	gap: .3rem;
	min-height: 36px;
	margin: 0 .2rem .65rem;
}

.sidebar-header-title {
	min-width: 0;
	margin: 0;
	overflow: hidden;
	color: var(--text);
	font-size: 1rem;
	font-weight: 400;
	line-height: 1.2;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.sidebar-header-actions {
	display: flex;
	align-items: center;
	gap: .2rem;
}
</style>
