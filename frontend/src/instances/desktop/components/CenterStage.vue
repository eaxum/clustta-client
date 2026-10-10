<template>
	<div ref="stageContainer" :class="['center-stage', { 'web-mode': platformStore.isWeb }]" @scroll="disableMenu()">
		<ContextMenu />
		<component v-for="stage in visibleStages" :key="stage.name" :is="stage.component" />
	</div>
</template>

<script setup>
// imports
import { computed, ref, onMounted, onUnmounted } from 'vue';

// state imports
import { useMenu } from '@/stores/menu';
import { useStageStore } from '@/stores/stages';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';

// states/stores
const menu = useMenu();
const stage = useStageStore();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const stageContainer = ref(null);

// components
import Browser from '@/instances/desktop/stages/Browser.vue'
import Projects from '@/instances/desktop/stages/Projects.vue'
import Dashboard from '@/instances/desktop/stages/Dashboard.vue'
import TrashList from '@/instances/desktop/stages/TrashList.vue'
import Account from '@/instances/desktop/stages/Account.vue'
import UserProfile from '@/instances/desktop/stages/UserProfile.vue'
import Settings from '@/instances/desktop/stages/Settings.vue'
import ProjectSettings from '@/instances/desktop/stages/ProjectSettings.vue'
import StudioSettings from '@/instances/desktop/stages/StudioSettings.vue'
import ContextMenu from '@/instances/desktop/menus/ContextMenu.vue'

const pageComponents = {
	projects: Projects,
	dashboard: Dashboard,
	browser: Browser,
	trash: TrashList,
	account: UserProfile,
	settings: Settings,
	projectSettings: ProjectSettings,
	studioSettings: StudioSettings,
};

const projectDatabaseStages = new Set(['browser', 'projectSettings', 'trash']);

const visibleStages = computed(() => {
	const stages = Object.entries(stage.stages)
		.filter(([name, isVisible]) => isVisible)
		.map(([name]) => ({
			name,
			component: pageComponents[name],
		}));
	if (projectStore.activeProjectCanQuery) return stages;
	if (!stages.some(({ name }) => projectDatabaseStages.has(name))) return stages;
	return [{ name: 'projects', component: Projects }];
});



// methods
const disableMenu = () => {
	menu.disableAllMenus();
};

onMounted(() => {
	if (platformStore.isWeb) {
		console.log('web')
		stage.setStageVisibility('account', true);
	} else {
		stage.setStageVisibility('projects', true);
	}
});


</script>

<style scoped>
@import "@/assets/desktop.css";

.center-stage {
	padding: .4rem;
	position: relative;
	height: 100%;
	width: 100%;
	display: flex;
	box-sizing: border-box;
	align-items: center;
	justify-content: center;
	overflow: hidden;
	min-width: 550px;
	background-color: firebrick;
	background-color: var(--surface-2-5);
	/* background-color: forestgreen; */
}

.center-stage.web-mode {
	min-width: unset;
}

.absolute-pane{
	padding-bottom: 0;
}
</style>

