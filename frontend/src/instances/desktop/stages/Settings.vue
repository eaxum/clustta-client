<template>
	<div ref="pageListRoot" class="page-list-root absolute-pane">
		<div class="settings-stage-root">
			<div class="settings-stage-header">
				<HeaderTabs :dataTypes="settingsItems" @filter="filterList" :fullWidth="true" :useSelected="true" :selectedTab="selectedSettingsTab" />
			</div>
			<div class="settings-stage-body">
				<div class="settings-stage-body-container">
					<component v-for="page in visiblePages" :key="page.name" :is="page.component" />
				</div>
			</div>
		</div>
	</div>
</template>

<script setup>
// imports
import { computed, onMounted, onUnmounted, ref, watchEffect } from 'vue';

// state imports
import { useSettingsStore } from '@/stores/settings';
import { useMenu } from '@/stores/menu';

// states/stores
const settings = useSettingsStore();
const menu = useMenu();

const pageListRoot = ref(null);
const selectedSettingsTab = ref('general');

// components
import HeaderTabs from '@/instances/common/components/HeaderTabs.vue';
import General from '@/instances/desktop/settings/General.vue';
import Collaborators from '@/instances/desktop/settings/Collaborators.vue';
import ProjectTemplates from '@/instances/desktop/settings/ProjectTemplates.vue';
import Directories from '@/instances/desktop/settings/Directories.vue';
import UserAdvanced from '@/instances/desktop/settings/UserAdvanced.vue';


// refs
const settingsComponents = {
	general: General,
	collaborators: Collaborators,
	projecttemplates: ProjectTemplates,
	directories: Directories,
	advanced: UserAdvanced,
};

// computed props
const settingsItems = computed(() => {
	
	const userSettingsIds = ['general', 'directories', 'projecttemplates', 'advanced'];
	const generalSettings = settings.settingsItems.filter((item) => userSettingsIds.includes(item.id));
	return generalSettings
});

const visiblePages = computed(() => {
	return Object.entries(settings.modalStates)
		.filter(([name, isVisible]) => isVisible)
		.map(([name]) => ({
			name,
			component: settingsComponents[name],
		}));
});

// methods
const filterList = (selectedTab) => {
	selectedSettingsTab.value = selectedTab;
	settings.setModalVisibility(selectedTab, true);
};

watchEffect(() => {
  if (pageListRoot.value) {
    menu.clickOutsideMask = pageListRoot.value;
  }
});

// lifecycle hooks
onMounted(() => {
	const tab = settings.pendingTab || 'general';
	settings.pendingTab = null;
	selectedSettingsTab.value = tab;
	settings.setModalVisibility(tab, true);
});

onUnmounted(() => {
	settings.pendingTab = null;
});

onUnmounted(() => {
	settings.disableAllModals();
	settings.activeModal = null;
})
</script>

<style>
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

