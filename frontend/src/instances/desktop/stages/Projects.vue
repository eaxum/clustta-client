<template>
	<div ref="projectListRoot" class="project-stage-root absolute-pane">
		<div class="task-header">
			<div class="create-menu" >
				<ActionButton :isDisabled="operationsActive" v-if="userStore.userCanCreateProject" :icon="getAppIcon('briefcase-plus')" 
					@click="createProject" v-tooltip="$t('stages.newProject')" :buttonFunction="doNothing" />
				<ActionButton :isDisabled="operationsActive" v-if="projectStore.selectedStudio?.name === 'Personal'" :icon="getAppIcon('arrow-down-ramp')" 
					v-tooltip="$t('stages.importProject')" :buttonFunction="importProject" />
				<ActionButton :isDisabled="operationsActive" v-if="projectStore.selectedStudio?.name !== 'Personal' && userStore.userCanCreateProject" :icon="getAppIcon('arrow-up-ramp')" 
					v-tooltip="$t('stages.uploadProject')" :buttonFunction="uploadProject" />
				<ActionButton :isDisabled="operationsActive" :icon="getAppIcon('refresh')" 
					v-tooltip="$t('common.refresh')" :buttonFunction="refresh" />
			</div>
			<div class="action-bar" v-if="projects.length && projectStore.projectsLoaded || projectStore.projectSearchQuery">
				<SearchBar ref="searchBar" v-model="projectStore.projectSearchQuery" :placeholder="$t('stages.searchProjects')" :isLoading="!projectStore.projectsLoaded" @input="updateSearch" @clear="clearSearch" />
			</div>
		<div class="view-options">
			<ActionButton v-if="projectStore.selectedStudio?.name === 'Personal'" :isDisabled="!untrackedProjects.length || operationsActive" :icon="getAppIcon(projectStore.showUntrackedProjects ? 'eye-cancel' : 'eye')" v-tooltip="projectStore.showUntrackedProjects ? $t('stages.hideUntrackedProjects') : $t('stages.showUntrackedProjects')"
				:buttonFunction="toggleShowUntrackedProjects" />
			<ActionButton :isDisabled="!projects.length || operationsActive" :icon="getAppIcon(cardView ? 'list' : 'four-squares')" :v-tooltip="cardView ? $t('stages.list') : $t('stages.cards')"
				:buttonFunction="switchViewLayout" />
		</div>
	</div>		<div ref="projectListContainer" class="project-list-root" 
		:class="{ 'project-list-root-hover-drop': isHovered }">

			<ProjectListSkeleton :cardView="cardView" v-if="!projectStore.projectsLoaded" />

			<div v-else-if="(openProjects.length || closedProjects.length) || (projectStore.showUntrackedProjects && untrackedProjects.length)" class="project-list-container" ref="openProjectsContainer" @scroll="disableAllMenus">
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
						:label="closedProjectsVisible ? $t('stages.hideArchivedProjects') : $t('stages.showArchivedProjects')"
						:smallIcons="true" :showLabel="true" @click="toggleExpandClosedProjects" />
					<div class="menu-divider"></div>
				</div>

				<div v-if="closedProjects.length && closedProjectsVisible" class="project-list"
					:class="{ 'project-list-cards': cardView }">
					<ProjectItem class="project-item" v-for="(project, index) in closedProjects" :key="project.id"
						:project="project" :index="index" :cardView="cardView"
						:style="{ animationDelay: index < 12 ? `${(index - 1) * 0.03}s` : '0s' }" />
				</div>

				<div v-if="untrackedProjects.length && projectStore.showUntrackedProjects" class="project-list-divider" ref="untrackedProjectsDivider">
					<TabButton
						:icon="untrackedProjectsVisible ? '/icons/chevron_up_white_slim.svg' : '/icons/chevron_down_white_slim.svg'"
						:label="untrackedProjectsVisible ? $t('stages.collapseUntrackedProjects') : $t('stages.expandUntrackedProjects')"
						:smallIcons="true" :showLabel="true" @click="toggleExpandUntrackedProjects" />
					<div class="menu-divider"></div>
				</div>

				<div v-if="untrackedProjects.length && untrackedProjectsVisible && projectStore.showUntrackedProjects" class="project-list"
					:class="{ 'project-list-cards': cardView }">
					<ProjectItem class="project-item" v-for="(project, index) in untrackedProjects" :key="project.id"
						:project="project" :index="index" :cardView="cardView"
						:style="{ animationDelay: index < 12 ? `${(index - 1) * 0.03}s` : '0s' }" />
				</div>
			</div>

			<PageState v-else :message="message()" :illustration="illustration()"
				:secondaryIcon="secondaryActionIcon()" :secondaryActionMessage="secondaryActionMessage()"
				:secondaryActionFunction="secondaryActionFunction" />

			

		</div>
	</div>
</template>


<script setup>
// imports
import { computed, ref, onMounted, onUnmounted, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';

// stores/state imports
import { useStageStore } from '@/stores/stages';
import { useProjectStore } from '@/stores/projects';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { usePaneStore } from '@/stores/panes';
import { useMenu } from '@/stores/menu';
import { useUserStore } from '@/stores/users';
import { useIconStore } from '@/stores/icons';
import { useDndStore } from '@/stores/dnd';


import ProjectItem from '@/instances/desktop/blocks/ProjectItem.vue'
import PageState from '@/instances/common/components/PageState.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import ProjectListSkeleton from '@/instances/desktop/components/ProjectListSkeleton.vue'
import SearchBar from '@/instances/desktop/components/SearchBar.vue'
import TabButton from '@/instances/desktop/components/TabButton.vue'
import { FSService, SettingsService } from '@/services';
import { Events } from "@wailsio/runtime";

// refs
const stage = useStageStore();
const projectStore = useProjectStore();
const modals = useDesktopModalStore();
const userStore = useUserStore();
const panes = usePaneStore();
const menu = useMenu();
const iconStore = useIconStore();
const dndStore = useDndStore();
const { t } = useI18n();

const projectListContainer = ref(null);
const projectListRoot = ref(null);
const projectListDivider = ref(null);
const untrackedProjectsDivider = ref(null);
const openProjectsContainer = ref(null);
const closedProjectsVisible = ref(false);
const untrackedProjectsVisible = ref(true);
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
	searchBar.value?.focus();
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
	let filteredProjects = projectStore.getProjects.filter((item) => 
		item.name.toLowerCase().includes(projectStore.projectSearchQuery)
	);
	
	return filteredProjects;
});

const trackedProjects = computed(() => {
	return projects.value.filter((project) => project.is_tracked !== false);
});

const untrackedProjects = computed(() => {
	return projects.value.filter((project) => project.is_tracked === false);
});

const openProjects = computed(() => {
	return trackedProjects.value.filter((project) => !project.is_closed);
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
	return trackedProjects.value.filter((project) => project.is_closed);
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

const toggleExpandUntrackedProjects = () => {
	const element = untrackedProjectsVisible.value ? openProjectsContainer.value : untrackedProjectsDivider.value;
	untrackedProjectsVisible.value = !untrackedProjectsVisible.value;
	nextTick(() => {
		if (element) {
			element.scrollIntoView({ behavior: 'smooth', block: 'start' });
		}
	});
};

const resetAll = () => {
	if (projectStore.projectSearchQuery) {
		projectStore.projectSearchQuery = '';
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
	await projectStore.loadProjects();
	// await projectStore.refreshProjects();
};

const updateSearch = (event) => {
	projectStore.projectSearchQuery = event.target.value.toLowerCase();
	closedProjectsVisible.value = projectStore.projectSearchQuery.length > 0;
};

const clearSearch = () => {
	projectStore.projectSearchQuery = '';
	closedProjectsVisible.value = false;
};


const message = () => {
	const searching = !!projectStore.projectSearchQuery.length;

	if (searching) {
		return t('stages.noProjectsMatchSearch')
	} else {
		if (userStore.userCanCreateProject) {
			return t('stages.noProjects')
		} else {
			return t('stages.noProjectAccess')
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
	const hasUntrackedProjects = untrackedProjects.value.length > 0;
	const hasTrackedProjects = trackedProjects.value.length > 0;

	if (searching) {
		return ''
	} else if (!hasTrackedProjects && hasUntrackedProjects) {
		return t('stages.displayUntrackedProjects')
	} else if (userStore.userCanCreateProject) {
		return t('stages.createNewProject')
	} else {
		return ''
	}
};

const secondaryActionIcon = () => {
	const hasUntrackedProjects = untrackedProjects.value.length > 0;
	const hasTrackedProjects = trackedProjects.value.length > 0;

	if (!hasTrackedProjects && hasUntrackedProjects) {
		return getAppIcon('eye');
	} else {
		return getAppIcon('plus-circle');
	}
};

const secondaryActionFunction = () => {
	const hasUntrackedProjects = untrackedProjects.value.length > 0;
	const hasTrackedProjects = trackedProjects.value.length > 0;
	
	if (!hasTrackedProjects && hasUntrackedProjects) {
		return toggleShowUntrackedProjects();
	} else if (userStore.userCanCreateProject) {
		return createProject();
	} else {
		return 
	}
};

const createProject = () => {
	modals.setModalVisibility('addProjectModal', true)
};

const importProject = () => {
	modals.setModalVisibility('importProjectModal', true)
};

const uploadProject = () => {
	modals.setModalVisibility('uploadProjectModal', true)
};

const switchViewLayout = () => {
	SettingsService.ToggleProjectGridView().then(() => {
		projectStore.isProjectGridView = !projectStore.isProjectGridView
	})
};

const toggleShowUntrackedProjects = () => {
	SettingsService.ToggleShowUntrackedProjects().then(() => {
		projectStore.showUntrackedProjects = !projectStore.showUntrackedProjects
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
	padding: 1rem 0rem;
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
	border-radius: var(--very-large-radius);
	background-color: var(--black-steel);
	width: 100%;
	box-sizing: border-box;
	min-width: 300px;
}

.project-list-root::-webkit-scrollbar {
	width: 4px;
}

.project-list-root::-webkit-scrollbar-thumb {
	border-radius: 4px;
	background-color: var(--dark-steel);
}

.project-list-root::-webkit-scrollbar-track {
	border-radius: 4px;
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
	width: 6px;
}

.project-list-container::-webkit-scrollbar-thumb {
	border-radius: 8px;
	background-color: var(--dark-steel);
}

.project-list-container::-webkit-scrollbar-track {
	border-radius: 8px;
	margin: 10px;
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
}

.action-bar {
	position: relative;
	display: flex;
	align-items: center;
	gap: .4rem;
	width: max-content;
	height: max-content;
	padding: .2rem;
	width: 500px;
	max-width: 400px;
}


.view-options {
	display: flex;
	gap: .4rem;
	align-items: center;
	padding: .2rem;
	width: max-content;
	height: max-content;
}
</style>

