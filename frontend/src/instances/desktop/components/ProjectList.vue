<template>
	<div class="project-list" ref="projectListRef" 
       :class="{ 
         'is-disabled': stage.operationActive, 
         'top-gradient': showTopGradient && !showCenterGradient, 
         'bottom-gradient': showBottomGradient && !showCenterGradient,
         'center-gradient': showCenterGradient 
       }" 
       @scroll="handleScroll">
		
	   <div v-if="projects.length" v-tooltip="'Pinned projects'" class="pinned-indicator">
			<div class="menu-divider"></div>
			<img class="tiny-icons" :src="getAppIcon('pin')">
			<div class="menu-divider"></div>
		</div>

		<span class="compound-list-item" v-for="(project, index) in projects" :key="project.uri">


			<div ref="listItem" class="project-avatar-item" v-tooltip="sidePaneActive ? '' : project.name"
				@click="projectStore.gotoProject(project)"
				:class="{ 'project-avatar-item-centered': !sidePaneActive, 'project-avatar-item-active': isActiveProject(project) }">

				<span v-if="project.icon.length < 10" class="project-icon" v-html="project.icon"></span>
				<span v-else class="project-icon">
					<img class="screenshot-thumb" :src="project.icon">
				</span>

			</div>

		</span>

		<div v-if="recents.length" v-tooltip="'Recent projects'" class="pinned-indicator">
			<div class="menu-divider"></div>
			<img class="tiny-icons" :src="getAppIcon('clock')">
			<div class="menu-divider"></div>
		</div>

		<span class="compound-list-item" v-for="(project, index) in recents" :key="project.uri">


			<div ref="listItem" class="project-avatar-item" v-tooltip="sidePaneActive ? '' : project.name"
				@click="projectStore.gotoProject(project)"
				:class="{ 'project-avatar-item-centered': !sidePaneActive, 'project-avatar-item-active': isActiveProject(project) }">

				<span v-if="project.icon.length < 10" class="project-icon" v-html="project.icon"></span>
				<span v-else class="project-icon">
					<img class="screenshot-thumb" :src="project.icon">
				</span>

				<div v-if="project.is_remote && activeProjectIndex === index && !sidePaneActive && project.is_unsynced"
					class="critical-items" :style="{ top: anchor.top + 'px', left: anchor.left + 'px' }">
				</div>

				<div v-if="sidePaneActive && project.is_unsynced" class="critical-items critical-items-static">
				</div>

			</div>

		</span>
	</div>
	
</template>

<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

// imports
import { computed, ref, onMounted } from 'vue';

// stores/state imports
import { useStageStore } from '@/stores/stages';
import { useProjectStore } from '@/stores/projects';

// refs
const stage = useStageStore();
const projectStore = useProjectStore();
const listItem = ref(null);

const props = defineProps({
	sidePaneActive: Boolean,
})

// computed properties
const activeProjectIndex = computed(() => {
	const projects = projectStore.projects;
	const activeProject = projectStore.getActiveProject;
	return projects.indexOf(activeProject);
});

const projects = computed(() => {
	const pinnedProjects = projectStore.pinnedProjects;
	return projectStore.projects.filter((project) => project.is_downloaded && pinnedProjects?.includes(project.id));
});

const recents = computed(() => {
	const pinnedProjects = projectStore.pinnedProjects;
	const recentProjects = projectStore.recentProjects;
	let recentProjectsAvailable = projectStore.projects.filter((project) => project.is_downloaded && recentProjects?.includes(project.id) && !pinnedProjects?.includes(project.id));

	// sort recent projects by most recent
	recentProjectsAvailable.sort((a, b) => {
		return recentProjects.indexOf(a.id) - recentProjects.indexOf(b.id);
	});

	// Calculate how many recent projects we can show
	const pinnedCount = projects.value.length || 0;
	const maxRecentCount = 10 - pinnedCount;

	return recentProjectsAvailable.slice(0, Math.max(0, maxRecentCount));
});

const criticalItemsDot = computed(() => {
	return projectStore.getActiveProject.has_remote && projectStore.getActiveProject.is_unsynced && !props.sidePaneActive
});

const anchor = computed(() => {
	if (listItem.value) {
		const anchor = listItem.value[activeProjectIndex.value].getBoundingClientRect();
		const anchorData = { top: (anchor.top - 3), left: (anchor.right - 7) }
		return anchorData
	} else {
		return { top: 0, left: 0 }
	}
});

// methods
const isActiveProject = (project) => {
	if (!project) {
		return false
	}
	return projectStore.getActiveProjectName === project.name
};

const dynamicName = (string) => {
	return props.sidePaneActive ? string : string[0].toUpperCase();
};


const projectListRef = ref(null);
const showTopGradient = ref(false);
const showBottomGradient = ref(false);
const showCenterGradient = ref(false);


onMounted(() => {
  handleScroll();
});

const handleScroll = () => {
  if (projectListRef.value) {
    const element = projectListRef.value;
    const hasTopOverflow = element.scrollTop > 0;
    
    const hasBottomOverflow = element.scrollTop + element.clientHeight < element.scrollHeight - 10;
    
    showTopGradient.value = hasTopOverflow;
    showBottomGradient.value = hasBottomOverflow;
    
    showCenterGradient.value = hasTopOverflow && hasBottomOverflow;
  }
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.pinned-indicator{
	display: flex;
	align-items: center;
	/* background-color: hotpink; */
}

.tiny-icons{
	padding-top: .3rem;
	height: 20px;
	opacity: .3;
}

.alert-items {
	z-index: 10000;
	overflow: hidden;
	width: 10px;
	height: 10px;
	background-color: #ecb603;
	border-radius: 5px;
	position: absolute;
	position: fixed;
	display: flex;
	align-items: center;
	justify-content: center;
	top: 15px;
	right: 15px;
	border-radius: 10px;
	padding: 3px;
	font-size: 12px;
	color: white;
	/* outline-offset: -1px; */
	outline: solid 1px rgb(236, 182, 3);
}

.critical-items {
	/* z-index: 10000; */
	overflow: hidden;
	width: min-content;
	max-width: 30px;
	min-width: 5px;
	height: 5px;
	background-color: #ecb603;
	border-radius: 5px;
	position: absolute;
	position: fixed;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 10px;
	padding: 3px;
	font-size: 12px;
	color: white;
	outline: solid 1px #bd2d2d;
	background-color: #bd2d2d;
}

.critical-items-static {
	position: relative;
}


.project-avatar-item {
	border-radius: 8px;
	box-sizing: border-box;
	cursor: pointer;
	display: flex;
	gap: 10px;
	justify-content: flex-start;
	align-items: center;
	padding: 3px 3px;
	width: 100%;
	color: #fff;
	/* background-color: coral; */
	overflow: hidden;
	text-wrap: nowrap;
	/* transition: all 0.3s ease; */
	box-sizing: border-box;
	height: 35px;
	width: 35px;
	aspect-ratio: 1;
	position: relative;
	justify-content: space-between;
	/* background-color: crimson; */

}

.project-avatar-item:hover {
	background-color: rgb(121, 121, 121);
	background-color: #ffffff15;
	outline: var(--transparent-line);
	outline-offset: -1px;
}

.project-avatar-item:active {
	background-color: rgb(70, 70, 70);
	background-color: #00000013;
}

.project-avatar-item-text {
	overflow: hidden;
	/* background-color: chocolate; */
	text-overflow: ellipsis;
}

.project-avatar-item {
	/* flex-direction: row; */
}

.project-avatar-item-centered {
	justify-content: center;
}

.project-item-preview-image {
	display: flex;
	box-sizing: border-box;
	align-items: center;
	justify-content: center;
	overflow: hidden;
	height: 100%;
	width: 100%;
	aspect-ratio: 1;
	/* background-color: var(--light-steel); */
	border-radius: 5px;
	filter: blur(0px);
}

.project-icon {
	font-size: x-large;
	font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI Emoji",
		"Segoe UI", "Apple Color Emoji", sans-serif;
	font-size: 1.8rem;
}

.project-avatar-item-active {
	background-color: rgb(173, 173, 173);
	background-color: var(--light-steel);
	color: var(--black-steel);
	outline: var(--transparent-line);
	outline-offset: -1px;
}

.project-avatar-item-active:hover {
	/* background-color: white; */
	color: var(--black-steel);
	outline: var(--solid-line);
	outline-offset: -1px;
}

.project-list {
	display: flex;
	box-sizing: border-box;
	overflow: hidden;
	flex-direction: column;
	gap: 1rem;
	padding: 2px 2px;
	width: 100%;
	height: max-content;
	/* background-color: var(--black-steel); */
	border-radius: 10px;
	transition: all 0.3s cubic-bezier(0.6, 0.05, 0.01, 0.99);
	height: 100%;
	overflow-y: scroll;
	position: relative; 
}

.project-list::-webkit-scrollbar {
	width: 0px;
}

/* Gradient overlay elements */
.top-gradient {
	background: linear-gradient(to bottom, rgba(255, 255, 255, 0.1) 0%, transparent 10%);
}

.bottom-gradient {
	background: linear-gradient(to top, rgba(255, 255, 255, 0.1) 0%, transparent 10%);
}

.center-gradient {
	background: 
        linear-gradient(to bottom, rgba(255, 255, 255, 0.1) 0%, transparent 10%),
        linear-gradient(to top, rgba(255, 255, 255, 0.1) 0%, transparent 10%);

}

.gradient-overlay.visible {
	opacity: 1;
}

.project-list-closed {
	height: 0px;
	padding: 0px .2rem;
}

.project-list-minimized {
	overflow: hidden;

	align-items: center;
}

.compound-list-item {
	/* background-color: hotpink; */
	align-items: center;
	justify-content: center;
}

.menu-divider{
	height: 5px;
	margin-top: 10px;
	margin-bottom: 10px;
}

.chevron-container {
	display: flex;
	box-sizing: border-box;
	overflow: hidden;
	/* flex-direction: column; */
	gap: .4rem;
	align-items: center;
	justify-content: center;
	padding: .2rem;
	width: 100%;
	height: max-content;
	height: 44px;
	/* background-color: darkorange; */
	transition: all 0.3s cubic-bezier(0.6, 0.05, 0.01, 0.99);
}

.chevron-icon {
	display: flex;
	box-sizing: border-box;
	overflow: hidden;
	width: min-content;
	height: min-content;
	/* background-color: teal; */
	transition: all 0.3s ease;
}
</style>

