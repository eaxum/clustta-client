<template>
	<div class="page-list-root absolute-pane">
		<div class="settings-stage-root">
			<div class="settings-stage-header">
				<HeaderTabs :useSelected="true" :selectedTab="selectedSettingsContext" :dataTypes="settingsItems" @filter="filterList" :fullWidth="true" />
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
import { computed, ref, onMounted, onUnmounted } from 'vue';

// state imports
import { useSettingsStore } from '@/stores/settings';
import { useUserStore } from '@/stores/users';
import { useProjectStore } from '@/stores/projects';
import { useEntitlementStore } from '@/stores/entitlements';

// states/stores
const userStore = useUserStore();
const settings = useSettingsStore();
const projectStore = useProjectStore();
const entitlementStore = useEntitlementStore();

// components
import HeaderTabs from '@/instances/common/components/HeaderTabs.vue';
import General from '@/instances/desktop/settings/General.vue';
import Collaborators from '@/instances/desktop/settings/Collaborators.vue';
import Templates from '@/instances/desktop/settings/Templates.vue';
import WorkflowTemplates from '@/instances/desktop/settings/WorkflowTemplates.vue';
import AssetTypes from '@/instances/desktop/settings/AssetTypes.vue';
import Roles from '@/instances/desktop/settings/Roles.vue';
import CollectionTypes from '@/instances/desktop/settings/CollectionTypes.vue';
import IgnoreList from '@/instances/desktop/settings/IgnoreList.vue';
import Advanced from '@/instances/desktop/settings/Advanced.vue';

const selectedSettingsContext = ref('');

// refs
const settingsComponents = {
	general: General,
	templates: Templates,
	collaborators: Collaborators,
	workflows: WorkflowTemplates,
	assettypes: AssetTypes,
	roles: Roles,
	collectiontypes: CollectionTypes,
	ignorelist: IgnoreList,
	advanced: Advanced,
};

// computed props
const settingsItems = computed(() => {
	const canChangeRole = userStore.canDo('change_role');
	const canViewTemplate = userStore.canDo('view_template');
	const isProjectRemote = projectStore.activeProject.has_remote;

	const canCollaborate = entitlementStore.canCollaborate;
	const hasCustomRoles = entitlementStore.hasCustomRoles;
	const hasIntegrations = entitlementStore.hasIntegrations;

	const userSettingsIds = ['general', 'directories', 'projecttemplates', 'studio', 'studiocollaborators', 'studiointegrations'];
	const remoteProjectIds = ['collaborators', 'roles', 'advanced'];

	const projectSettings = settings.settingsItems.filter((item) => 
		!userSettingsIds.includes(item.id) &&
		(canChangeRole || item.id !== 'roles') &&
		(canChangeRole || item.id !== 'advanced') &&
		(canViewTemplate || item.id !== 'projecttemplates') &&
		(canCollaborate || item.id !== 'collaborators') &&
		(hasCustomRoles || item.id !== 'roles') &&
		(hasIntegrations || item.id !== 'advanced')
	);
	
	const localProjectSettings = projectSettings.filter((item) => !remoteProjectIds.includes(item.id));
	return isProjectRemote ? projectSettings : localProjectSettings;
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
	selectedSettingsContext.value = selectedTab;
	settings.setModalVisibility(selectedTab, true);
};

// onmounted hook
onMounted(() => {
	if(!settings.activeModal){
		settings.setModalVisibility('templates', true);
		settings.activeModalName = 'templates';
		selectedSettingsContext.value = 'templates';
	} else {
		selectedSettingsContext.value = settings.activeModal;
	}
});

onUnmounted(() => {
	settings.disableAllModals();
	settings.activeModal = null;
})

</script>

<style scoped>
@import "@/assets/desktop.css";


.page-list-root {
	box-sizing: border-box;
	padding: .4rem;
	display: flex;
	align-items: center;
	justify-content: center;
	color: white;
	/* background-color: sienna; */
}

.settings-stage-root {
	display: flex;
	flex-direction: column;
	box-sizing: border-box;
	/* background-color: firebrick; */
	width: 100%;
	height: 100%;
	gap: .5rem;

}

.settings-stage-header {
	width: 100%;
	display: flex;
	box-sizing: border-box;
	/* background-color: rebeccapurple; */
	align-items: flex-start;
	justify-content: flex-start;
}

.settings-stage-body {
	width: 100%;
	/* max-width: 1200px; */
	height: 100%;
	display: flex;
	box-sizing: border-box;
	/* background-color: teal; */
	align-items: flex-start;
	justify-content: flex-start;
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
	/* background-color: tomato; */
	align-items: flex-start;
	justify-content: flex-start;
	justify-content: center;
	overflow: hidden;
	padding: .5rem;
}

.page-list {
	background-color: rgb(125, 192, 59);
	padding: .4rem;
	display: flex;
	gap: .4rem;
	flex-direction: column;
	box-sizing: border-box;
	height: 100%;
	width: 100%;
	/* overflow: hidden; */
	height: max-content;
}

.page-list-container {
	display: flex;
	padding-right: .4rem;
	overflow: hidden;
	overflow-y: scroll;
	height: 100%;
	width: 100%;
	box-sizing: border-box;
	/* max-width: 600px; */
	min-width: 300px;
}

.page-list-container::-webkit-scrollbar {
	width: 8px;
}

.page-list-container::-webkit-scrollbar-thumb {
	border-radius: 10px;
	background-color: rgb(36, 49, 59);
}

.page-list-container::-webkit-scrollbar-track {
	border-radius: 10px;
}

.page-header {
	position: relative;
	display: flex;
	width: 100%;
	align-items: center;
	height: max-content;
	gap: 1rem;
	justify-content: space-between;
	background-color: khaki;
	padding: .2rem;
	box-sizing: border-box;
	min-width: max-content;
}
</style>

