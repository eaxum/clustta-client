<template>
	<aside
		class="sidebar-frame"
		:class="{ 'sidebar-frame-expanded': expanded }"
		:style="sidebarStyle"
	>
		<div class="sidebar-frame-content">
			<slot />
		</div>
	</aside>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
	expanded: {
		type: Boolean,
		default: false,
	},
	expandedWidth: {
		type: Number,
		default: 240,
	},
	collapsedWidth: {
		type: Number,
		default: 50,
	},
});

const sidebarStyle = computed(() => ({
	'--sidebar-width': `${props.expanded ? props.expandedWidth : props.collapsedWidth}px`,
}));
</script>

<style scoped>
.sidebar-frame {
	display: flex;
	flex: 0 0 var(--sidebar-width);
	width: var(--sidebar-width);
	min-width: var(--sidebar-width);
	height: 100%;
	overflow: hidden;
	background-color: var(--surface-1);
	box-sizing: border-box;
	transition:
		width 180ms cubic-bezier(0.6, 0.05, 0.01, 0.99),
		min-width 180ms cubic-bezier(0.6, 0.05, 0.01, 0.99),
		flex-basis 180ms cubic-bezier(0.6, 0.05, 0.01, 0.99);
}

.sidebar-frame-content {
	display: flex;
	width: 100%;
	min-width: 0;
	height: 100%;
	overflow: hidden;
}

@media (prefers-reduced-motion: reduce) {
	.sidebar-frame {
		transition: none;
	}
}
</style>
