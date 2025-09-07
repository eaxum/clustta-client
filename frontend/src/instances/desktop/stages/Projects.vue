<template>
	<div ref="projectListRoot" class="project-stage-root absolute-pane">

		<div class="task-header">
			<div class="create-menu" >
				<ActionButton v-if="userStore.userCanCreateProject" :icon="getAppIcon('briefcase-plus')" :label="!panes.showDetailsPane ? 'New Project' : ''"
					@click="createProject" v-tooltip="'New Project'" :buttonFunction="doNothing" />
				<ActionButton v-if="projectStore.projectsLoaded" :icon="getAppIcon('refresh')" :label="!panes.showDetailsPane ? 'Refresh' : ''"
					v-tooltip="'Refresh'" :buttonFunction="refresh" />
			</div>
			<div class="action-bar" v-if="projects.length && projectStore.projectsLoaded || projectStore.projectSearchQuery">
				<input ref="searchBar" v-model="projectStore.projectSearchQuery" class="desktop-search-bar" type="text"
					:placeholder="'Search projects'" @input="updateSearch" />
			</div>
			<div class="view-options">
				<ActionButton v-if="cardView && projects.length" :icon="getAppIcon('list')" v-tooltip="'List'"
					:buttonFunction="switchViewLayout" />
				<ActionButton v-else-if="projects.length" :icon="getAppIcon('four-squares')" v-tooltip="'Cards'"
					:buttonFunction="switchViewLayout" />
				<ActionButton v-if="projects.length" :isDisabled="!projects.length"
					:icon="panes.showDetailsPane ? getAppIcon('collapse-right') : getAppIcon('collapse-left')"
					v-tooltip="panes.showDetailsPane ? 'Close pane' : 'Open pane'"
					:buttonFunction="toggleDetailsPane" />
			</div>
		</div>

		<div ref="projectListContainer" class="project-list-root" 
		:class="{ 'project-list-root-hover-drop': isHovered }">

			<ProjectListSkeleton :cardView="cardView" v-if="!projectStore.projectsLoaded" />

			<div v-else-if="projects.length" class="project-list-container" ref="openProjectsContainer" @scroll="disableAllMenus">
				<div v-if="openProjects.length" class="project-list" :class="{ 'project-list-cards': cardView }">
					<ProjectItem class="project-item" v-for="(project, index) in pinnedProjects" :key="project.id"
						:project="project" :index="index" :cardView="cardView"
						:style="{ animationDelay: index < 12 ? `${(index - 1) * 0.03}s` : '0s' }" />
					<ProjectItem class="project-item" v-for="(project, index) in unpinnedProjects" :key="project.id"
						:project="project" :index="index" :cardView="cardView"
						:style="{ animationDelay: index < 12 ? `${(index - 1) * 0.03}s` : '0s' }" />
					<ProjectItem class="project-item" v-for="(project, index) in undownloadedProjects" :key="project.id"
						:project="project" :index="index" :cardView="cardView"
						:style="{ animationDelay: index < 12 ? `${(index - 1) * 0.03}s` : '0s' }" />
				</div>

				<div v-if="closedProjects.length" class="project-list-divider" ref="projectListDivider">
					<TabButton
						:icon="closedProjectsVisible ? '/icons/chevron_up_white_slim.svg' : '/icons/chevron_down_white_slim.svg'"
						:label="closedProjectsVisible ? 'Hide archived projects' : 'Show archived projects'"
						:showLabel="true" @click="toggleExpandClosedProjects" />
					<div class="menu-divider"></div>
				</div>

				<div v-if="closedProjects.length && closedProjectsVisible" class="project-list"
					:class="{ 'project-list-cards': cardView }">
					<ProjectItem class="project-item" v-for="(project, index) in closedProjects" :key="project.id"
						:project="project" :index="index" :cardView="cardView"
						:style="{ animationDelay: index < 12 ? `${(index - 1) * 0.03}s` : '0s' }" />
				</div>
			</div>

			<PageState v-else :message="message()" :illustration="illustration()"
				:secondaryIcon="getAppIcon('plus-circle')" :secondaryActionMessage="secondaryActionMessage()"
				:secondaryActionFunction="secondaryActionFunction" />

		</div>
	</div>
</template>


<script setup>
// imports
import { computed, ref, onMounted, onUnmounted, nextTick } from 'vue';

// stores/state imports
import { useStageStore } from '@/stores/stages';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { usePaneStore } from '@/stores/panes';
import { useMenu } from '@/stores/menu';
import { useUserStore } from '@/stores/users';
import { useIconStore } from '@/stores/icons';
import { useDndStore } from '@/stores/dnd';


import ProjectItem from '@/instances/desktop/blocks/ProjectItem.vue'
import PageState from '@/instances/common/components/PageState.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import TabButton from '@/instances/desktop/components/TabButton.vue'
import ProjectListSkeleton from '@/instances/desktop/components/ProjectListSkeleton.vue'
import { FSService, SettingsService } from '@/../bindings/clustta/services/index';
import { Events } from "@wailsio/runtime";

// refs
const trayStates = useTrayStates();
const stage = useStageStore();
const projectStore = useProjectStore();
const modals = useDesktopModalStore();
const projectToRemove = ref(null);
const userStore = useUserStore();
const panes = usePaneStore();
const menu = useMenu();
const iconStore = useIconStore();
const dndStore = useDndStore();

const projectListContainer = ref(null);
const projectListRoot = ref(null);
const projectListDivider = ref(null);
const openProjectsContainer = ref(null);
const closedProjectsVisible = ref(false);
const searchBar = ref(null);

const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

const cardView = computed(() => {
	return projectStore.isProjectGridView
});

const operationsActive = computed(() => {
	return stage.operationActive || !!modals.activeModal || !!menu.activeMenu || !projectStore.projectsLoaded
});

Events.On('search', async () => {
	if (operationsActive.value) return
	if (searchBar.value) {
		searchBar.value.focus();
	}
});

Events.On('reload-view', async () => {
	if (operationsActive.value) return
	refresh();
});

Events.On('new-project', async () => {
	if (operationsActive.value) return
	if(userStore.userCanCreateProject){
		createProject();
	}
});

// Events.On("clustta-drag-over", (event) => {
// 	if (!userStore.userCanCreateProject || stage.selectedStage !== 'projects') {
// 		dndStore.isDragging = false;
// 		dndStore.isDropHovering = false;
// 		return
// 	}
// 	dndStore.dragPosition = {
// 		x: event.data[0].position.x,
// 		y: event.data[0].position.y,
// 	}
// 	dndStore.isDragging = true;
// 	dndStore.isDropHovering = true;
// });

// Events.On('clustta-drag-drop', async (event) => {
// 	if (!userStore.userCanCreateProject || stage.selectedStage !== 'projects') {
// 		dndStore.isDragging = false;
// 		dndStore.isDropHovering = false;
// 		return
// 	}
// 	const paths = event.data[0].paths;
// 	const droppedItems = await categorizePaths(paths);
// 	dndStore.droppedFolders = droppedItems.folders;
// 	dndStore.droppedFiles = droppedItems.files;
// 	createProject();

// 	// reset
// 	dndStore.isDragging = false;
// 	dndStore.isDropHovering = false;
// });

const categorizePaths = async (paths) => {
	let files = []
	let folders = []
	for (let fullPath of paths) {
		let isFile = await FSService.IsFile(fullPath)
		if (isFile) {
			files.push(fullPath)
		} else {
			folders.push(fullPath)
		}
	}
	return { folders: folders, files: files }
}

// computed properties
const isHovered = computed(() => { 
	return 
	return dndStore.isDropHovering
})
const projects = computed(() => {
	return projectStore.getProjects.filter((item) => item.name.toLowerCase().includes(projectStore.projectSearchQuery))
});

const openProjects = computed(() => {
	return projects.value.filter((project) => !project.is_closed);
});

const downloadedProjects = computed(() => {
	return openProjects.value.filter((project) => project.is_downloaded);
});

const pinnedProjects = computed(() => {
	const pinnedProjects = projectStore.pinnedProjects;
	return downloadedProjects.value.filter((project) => pinnedProjects?.includes(project.id));
});

const unpinnedProjects = computed(() => {
	const pinnedProjects = projectStore.pinnedProjects;
	return downloadedProjects.value.filter((project) => !pinnedProjects?.includes(project.id));
});

const undownloadedProjects = computed(() => {
	return openProjects.value.filter((project) => !project.is_downloaded);
});

const closedProjects = computed(() => {
	return projects.value.filter((project) => project.is_closed);
});

// methods
const toggleExpandClosedProjects = () => {
	const element = closedProjectsVisible.value ? openProjectsContainer.value : projectListDivider.value;
	console.log(element)
	closedProjectsVisible.value = !closedProjectsVisible.value;
	// return
	nextTick(() => {
		element.scrollIntoView({ behavior: 'smooth', block: 'start' });
	});
};

const resetAll = () => {
	if (projectStore.projectSearchQuery) {
		projectStore.projectSearchQuery = '';
		// if (searchBar.value) {
		// 	searchBar.value.focus();
		// }
	}
};

const disableAllMenus = () => {
	menu.disableAllMenus();
};

const toggleDetailsPane = () => {
	if (!projects.value.length) {
		return
	}
	panes.showDetailsPane = !panes.showDetailsPane;
};

const doNothing = () => {
	console.log('nothing');
};

const refresh = async () => {
	await projectStore.refreshProjects();
};

const updateSearch = (event) => {
	projectStore.projectSearchQuery = event.target.value.toLowerCase();
};


const message = () => {
	const searching = !!projectStore.projectSearchQuery.length;

	if (searching) {
		return 'No projects match your search.'
	} else {
		if (userStore.userCanCreateProject) {
			return 'You have no projects.'
		} else {
			return 'You dont have access to any projects. Please check in with your supervisor'
		}
	}

};

const illustration = () => {

	const searching = !!projectStore.projectSearchQuery.length;

	if (searching) {
		return '/page-states/project.png'
	} else {
		return '/page-states/project.png'
	}

};

const secondaryActionMessage = () => {
	const searching = !!projectStore.projectSearchQuery.length;

	if (searching) {
		return ''
	} else if (userStore.userCanCreateProject) {
		return 'Create New Project.'
	} else {
		return ''
	}
};

const secondaryActionFunction = () => {
	
	
	if (userStore.userCanCreateProject) {
			return createProject();
		} else {
			return 
		}
};

const createProject = () => {
	modals.setModalVisibility('addProjectModal', true)
};

const switchViewLayout = () => {
	SettingsService.ToggleProjectGridView().then(() => {
		projectStore.isProjectGridView = !projectStore.isProjectGridView
	})
};

// Handle escape key globally
const handleEscapeKey = (event) => {
	if (event.key === 'Escape' && projectStore.projectSearchQuery) {
		resetAll();
	}
};

onMounted(() => {
	projectStore.projectSearchQuery = ''
	panes.showDetailsPane = false;
	// projectStore.activeProject = null;
	// console.log(projectStore.activeProject)

	// Add event listener for escape key
	document.addEventListener('keydown', handleEscapeKey);
});

onUnmounted(() => {
	projectStore.projectSearchQuery = ''
	stage.markedProjects = [];
	stage.selectdProject = '';

	// Remove event listener when component is unmounted
	document.removeEventListener('keydown', handleEscapeKey);
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.pickers {
	display: flex;
	width: 100%;
	height: 100%;
}

.project-stage-root {
	box-sizing: border-box;
	/* padding: .4rem; */
	display: flex;
	align-items: center;
	justify-content: center;
	/* background-color: khaki;
    background-color: firebrick; */
	color: var(--white);
	flex-direction: column;
}

.absolute-pane {
	/* padding: 0px; */
}

.project-item {
	opacity: 0;
	animation: fadeIn .2s ease-in-out forwards;
}

@keyframes fadeIn {
	from {
		opacity: 0;
	}

	to {
		opacity: 1;
	}
}

.all-tasks-collapsed {
	transform: rotate(90deg);
}

.all-tasks-expanded {
	transform: rotate(-90deg);
}

.project-list {
	box-sizing: border-box;
	height: max-content;
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(500px, 1fr));
	gap: 10px;
	width: 100%;
}

.project-list-divider {
	display: flex;
	/* background-color: hotpink; */
	width: 100%;
	align-items: center;
	justify-content: space-between;
	gap: 1rem;
	padding: 1rem;
	box-sizing: border-box;
}

.project-list-cards {
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
	gap: 10px;
	width: 100%;
}

.project-list-root {
	/* z-index: 5; */
	position: relative;
	padding: .5rem;
	overflow: hidden;
	/* overflow-y: scroll; */
	height: 100%;
	background-color: tomato;
	border-radius: var(--large-radius);
	background-color: var(--black-steel);
	width: 100%;
	box-sizing: border-box;
	min-width: 300px;
}

.project-list-root::-webkit-scrollbar {
	width: 8px;
}

.project-list-root::-webkit-scrollbar-thumb {
	border-radius: 10px;
	background-color: var(--dark-steel);
}

.project-list-root::-webkit-scrollbar-track {
	border-radius: 10px;
}

.project-list-root {
	z-index: 1;
	position: relative;
	display: flex;
	flex-direction: column;
	/* background-color: firebrick; */
	color: var(--white);
	width: 100%;
	/* min-width: max-content; */
	/* max-width: 300px; */
	height: 100%;
	box-sizing: border-box;
	overflow: hidden;
	padding: .4rem;
	gap: .4rem;
	/* flex: 1 1 50%; */
}

.project-list-root-hover-drop {
	background-color: #1e7fee6c;
	outline: 1px solid rgb(255, 255, 255);
	outline-offset: -1px;
}

.project-list-container {
	z-index: 1;
	position: relative;
	display: flex;
	flex-direction: column;
	/* background-color: firebrick; */
	color: var(--white);
	width: 100%;
	/* height: 100%; */
	box-sizing: border-box;
	overflow: hidden;
	padding: .4rem;
	gap: .4rem;
	overflow: hidden;
	overflow-y: scroll;
}

.project-list-container::-webkit-scrollbar {
	width: 8px;
}

.project-list-container::-webkit-scrollbar-thumb {
	border-radius: 10px;
	background-color: var(--dark-steel);
}

.project-list-container::-webkit-scrollbar-track {
	border-radius: 10px;
}

.task-header {
	position: relative;
	display: flex;
	width: 100%;
	align-items: center;
	height: max-content;
	min-height: 50px;
	gap: 1rem;
	justify-content: space-between;
	/* background-color: khaki; */
	padding: .2rem;
	box-sizing: border-box;
	min-width: max-content;
}

.create-menu {
	position: relative;
	display: flex;
	align-items: center;
	gap: .4rem;
	width: max-content;
	height: max-content;
	padding: .2rem;
	/* background-color: black; */
}

.action-bar {
	position: relative;
	display: flex;
	align-items: center;
	gap: .4rem;
	width: max-content;
	height: max-content;
	padding: .2rem;
	/* background-color: black; */
}

.desktop-search-bar {
	font-family: 'Inter', sans-serif;
	box-sizing: border-box;
	font-size: 16px;
	border-radius: 8px;
	padding: 10px;
	border: 0px;
	border-style: solid;
	outline: none;
	background-color: var(--midnight-steel);
	color: var(--white);
	transition: width 0.2s ease-out;
	border-radius: var(--large-radius);
	width: 100%;
	width: 500px;
	max-width: 400px;
	color: var(--white);
}

.desktop-search-bar::-ms-reveal {
	filter: invert(100%);
}

.desktop-search-bar:hover {
	outline: var(--transparent-line);
	outline-offset: -1px;
}

.desktop-search-bar:focus {
	outline: var(--transparent-line);
	outline-offset: -1px;
}

.view-options {
	display: flex;
	gap: .4rem;
	align-items: center;
	padding: .2rem;
	width: max-content;
	height: max-content;
	/* background-color: darkorange; */
}

.compound-list-item {
	/* background-color: blue; */
}
</style>

