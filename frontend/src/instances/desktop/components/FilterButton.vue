<template>
	<span v-stop-propagation @click="buttonFunction" class="filter-button"
		:class="{ 'button-background': alert, 'full-width': fullWidth, 'icon-after': iconAfter, 'centered': centered }">
		<component v-if="isIconComponent" :is="icon" :size="16" class="ci-btn-icon no-cursor" :class="{ 'inverted-icon': alert }" />
		<img v-else class="small-icons no-cursor" :class="{ 'inverted-icon': alert }" :src="icon">
		<div v-if="showLabel && label" class="label-text no-cursor" :class="{ 'label-text-inverted': alert }">{{ label }}</div>
		<img class="small-icons chevron no-cursor" :class="{ 'inverted-icon': alert }" src="/icons/chevron_down_white.svg">
	</span>
</template>

<script setup>
import { computed } from 'vue';
// props
const props = defineProps({
	alert: { type: Boolean, default: false },
	buttonFunction: { type: Function, default: () => {} },
	centered: { type: Boolean, default: false },
	fullWidth: { type: Boolean, default: false },
	icon: { type: [String, Object, Function], default: '' },
	iconAfter: { type: Boolean, default: false },
	label: { type: String, default: '' },
	showLabel: { type: Boolean, default: false },
});

const isIconComponent = computed(() => props.icon && typeof props.icon !== 'string');
</script>

<style scoped>
@import "@/assets/desktop.css";

.centered {
	justify-content: space-around;
}

.chevron {
	pointer-events: none;
	height: 10px;
	min-width: 10px;
	transition: all .1s ease-in;
}

.filter-button {
	overflow: hidden;
	background-color: var(--black-steel);
	text-align: center;
	font-size: 14px;
	line-height: 14px;
	color: var(--white);
	position: relative;
	border-radius: var(--large-radius);
	box-sizing: border-box;
	cursor: pointer;
	display: flex;
	align-items: center;
	padding: .2rem;
	padding-right: .5rem;
	gap: .2rem;
	height: max-content;
	width: max-content;
	min-width: max-content;
	min-height: max-content;
	outline-offset: -1px;
}

.filter-button:hover {
	outline: var(--transparent-line);
}

.filter-button:active {
	background-color: #00000013;
}

.full-width {
	width: 100%;
}

.icon-after {
	justify-content: space-between;
}

.inverted-icon {
	filter: none;
}

[data-theme="dark"] .inverted-icon {
	filter: invert(1);
}

.button-background {
	background-color: var(--white);
	color: var(--black);
}

.label-text {
	font-size: 14px;
	font-weight: 400;
}

.label-text-inverted {
	font-size: 14px;
	font-weight: 400;
}

.no-cursor {
	pointer-events: none;
}

.ci-btn-icon {
	stroke: var(--light-steel);
}
</style>  
  

