<template>
	<div ref="pageListRoot" class="page-list-root absolute-pane">
		<SettingsShell
			:pages="userSettingsPages"
			:activePageId="selectedSettingsTab"
		/>
	</div>
</template>

<script setup>
import { onMounted, onUnmounted, ref, watch, watchEffect } from 'vue';
import SettingsShell from '@/instances/desktop/settings/components/SettingsShell.vue';
import {
	userSettingsPages,
} from '@/instances/desktop/settings/settingsNavigation';
import { useMenu } from '@/stores/menu';
import { useSettingsStore } from '@/stores/settings';

const DEFAULT_PAGE_ID = 'general';

const menu = useMenu();
const settings = useSettingsStore();
const pageListRoot = ref(null);
const selectedSettingsTab = ref(DEFAULT_PAGE_ID);

const isAvailablePage = (pageId) => {
	return userSettingsPages.some((page) => page.id === pageId);
};

const filterList = (pageId) => {
	if (!isAvailablePage(pageId)) return;
	selectedSettingsTab.value = pageId;
	settings.activeModalName = pageId;
	settings.setModalVisibility(pageId, true);
};

watch(
	() => settings.activeModal,
	(pageId) => {
		if (isAvailablePage(pageId)) selectedSettingsTab.value = pageId;
	},
);

watchEffect(() => {
	if (pageListRoot.value) menu.clickOutsideMask = pageListRoot.value;
});

onMounted(() => {
	const requestedPageId = settings.pendingTab;
	settings.pendingTab = null;
	filterList(isAvailablePage(requestedPageId) ? requestedPageId : DEFAULT_PAGE_ID);
});

onUnmounted(() => {
	settings.pendingTab = null;
	settings.disableAllModals();
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.page-list-root {
	box-sizing: border-box;
	/* padding: .4rem; */
	display: flex;
	align-items: center;
	justify-content: center;
	color: white;
	background-color: var(--surface-3);
	/* background-color: crimson; */
}

.settings-stage-root {
	display: flex;
	flex-direction: column;
	box-sizing: border-box;
	width: 100%;
	height: 100%;
	gap: .5rem;
}

.settings-stage-header {
	width: 100%;
	display: flex;
	box-sizing: border-box;
	align-items: flex-start;
	justify-content: flex-start;
}

.settings-stage-body {
	width: 100%;
	height: 100%;
	display: flex;
	box-sizing: border-box;
	align-items: flex-start;
	justify-content: center;
	overflow: hidden;
	padding: .5rem;
}

.settings-stage-body-container {
	width: 100%;
	max-width: 960px;
	height: 100%;
	display: flex;
	box-sizing: border-box;
	align-items: center;
	justify-content: center;
	overflow: hidden;
	padding: .5rem;
}

.settings-component-container {
  border-radius: var(--gigantic-radius) !important;
}

/* Shared Card Styles for all settings components */
.settings-section-card {
	display: flex;
	flex-direction: column;
	background-color: var(--surface-1);
	border-radius: var(--very-large-radius);
	overflow: hidden;
	box-sizing: border-box;
	width: 100%;
	outline: var(--transparent-line);
	outline-offset: -1px;
	height: min-content;
}

.settings-section-card-header {
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: var(--transparent-line);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.settings-section-card-title {
  font-size: 1rem;
  font-weight: 300;
  color: var(--text);
  margin: 0;
  flex: 1;
}

.settings-section-card-content {
  display: flex;
  flex-direction: column;
  /* gap: 0.5rem; */
  background-color: var(--surface-2);
  border-radius: var(--normal-radius);
  overflow: hidden;
  height: min-content;
}

/* Shared action divider */
.actions-divider {
	display: flex;
	background-color: var(--surface-4);
	height: 16px;
	width: 1.5px;
}

/* Shared horizontal flex utility */
.horizontal-flex {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex: 1;
}
</style>

