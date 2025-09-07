<template>
	<div ref="stageContainer" class="center-stage" @scroll="disableMenu()">
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

// states/stores
const menu = useMenu();
const stage = useStageStore();
const stageContainer = ref(null);

// components
import Dependencies from '@/instances/desktop/stages/Dependencies.vue'
import Browser from '@/instances/desktop/stages/Browser.vue'
import Projects from '@/instances/desktop/stages/Projects.vue'
import Dashboard from '@/instances/desktop/stages/Dashboard.vue'
import TrashList from '@/instances/desktop/stages/TrashList.vue'
import Account from '@/instances/desktop/stages/Account.vue'
import Settings from '@/instances/desktop/stages/Settings.vue'
import ProjectSettings from '@/instances/desktop/stages/ProjectSettings.vue'
import StudioSettings from '@/instances/desktop/stages/StudioSettings.vue'
import ContextMenu from '@/instances/desktop/menus/ContextMenu.vue'

const pageComponents = {
	projects: Projects,
	dependencies: Dependencies,
	dashboard: Dashboard,
	browser: Browser,
	trash: TrashList,
	account: Account,
	settings: Settings,
	projectSettings: ProjectSettings,
	studioSettings: StudioSettings,
};

const visibleStages = computed(() => {
	return Object.entries(stage.stages)
		.filter(([name, isVisible]) => isVisible)
		.map(([name]) => ({
			name,
			component: pageComponents[name],
		}));
});



// methods
const disableMenu = () => {
	menu.disableAllMenus();
};

onMounted(() => {
	stage.setStageVisibility('projects', true);
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
	background-color: var(--steel);
}

.absolute-pane{
	padding-bottom: 0;
}
</style>

