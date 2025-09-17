<template>
	<div ref="pageListRoot" class="page-list-root absolute-pane">
		<div class="settings-stage-root">
			<div class="settings-stage-header">
				<HeaderTabs :dataTypes="settingsItems" @filter="filterList" :fullWidth="true" />
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
import { computed, ref, onMounted, onUnmounted, watchEffect } from 'vue';

// state imports
import { useSettingsStore } from '@/stores/settings';
import { useMenu } from '@/stores/menu';

// states/stores
const settings = useSettingsStore();
const menu = useMenu();

const pageListRoot = ref(null);

// components
import HeaderTabs from '@/instances/common/components/HeaderTabs.vue';
import General from '@/instances/desktop/settings/General.vue';
import Collaborators from '@/instances/desktop/settings/Collaborators.vue';
import ProjectTemplates from '@/instances/desktop/settings/ProjectTemplates.vue';


// refs
const settingsComponents = {
	general: General,
	collaborators: Collaborators,
	projecttemplates: ProjectTemplates,
};

// computed props
const settingsItems = computed(() => {
	
	const userSettingsItems = ['General', 'Project Templates'];
	const generalSettings = settings.settingsItems.filter((item) => userSettingsItems.includes(item.name));
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
	let modalName;
	if(selectedTab === 'Project Templates'){
		modalName = 'ProjectTemplates'
	} else {
		modalName = selectedTab;
	}
	const selectedTabName = modalName.toLowerCase();
	settings.setModalVisibility(selectedTabName, true);
};

watchEffect(() => {
  if (pageListRoot.value) {
    menu.clickOutsideMask = pageListRoot.value;
  }
});

// onmounted hook
onMounted(() => {
	settings.setModalVisibility('general', true);
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
	align-items: center;
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

