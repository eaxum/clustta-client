<template>
	<div class="header-bar">

		<div class="header-bar-breadcrumbs-parent">

			<div v-if="projectIsActive && stage.activeStage === 'browser'" class="header-bar-tabs">
				<WorkspaceTabs />
			</div>

			<div v-if="stage.activeStage === 'projects'" class="header-bar-dependencies">
				<ActionButton :icon="getAppIcon('home')" :isInactive="true" />
				<div class="header-area-container">
					<HeaderArea :title="'Projects'" :miniDisplay="true" />
				</div>
			</div>

			<div v-if="stage.activeStage === 'dependencies'" class="header-bar-dependencies">
				<ActionButton :icon="getAppIcon('arrow-left')" @click="goToList()" v-tooltip="'Back'" />
				<div class="header-area-container" @click="toggleFullTaskPath()">
					<HeaderArea :title="taskName" :miniDisplay="true" :customIcon="assetStore.selectedTask.icon" />
				</div>
			</div>

			<div v-if="stage.activeStage === 'trash'" class="header-bar-dependencies">
				<ActionButton :icon="getAppIcon('arrow-left')" @click="goToList()" v-tooltip="'Back'" />
				<div class="header-area-container">
					<HeaderArea :title="'Trash'" :miniDisplay="true" />
				</div>
			</div>

			<div v-if="stage.activeStage === 'projectSettings'" class="header-bar-dependencies">
				<ActionButton :icon="getAppIcon('arrow-left')" @click="goToList()" v-tooltip="'Back'" />
				<div class="header-area-container">
					<HeaderArea :title="'Project Settings'" :miniDisplay="true" />
				</div>
			</div>

			<div v-if="stage.activeStage === 'studioSettings'" class="header-bar-dependencies">
				<ActionButton :icon="getAppIcon('arrow-left')" @click="goToProjects()" v-tooltip="'Back'" />
				<div class="header-area-container">
					<HeaderArea :title="'Studio Settings'" :miniDisplay="true" />
				</div>
			</div>

			<div v-if="stage.activeStage === 'settings'" class="header-bar-dependencies">
				<ActionButton :icon="getAppIcon('arrow-left')" @click="goToProjects()" v-tooltip="'Back'" />
				<div class="header-area-container">
					<HeaderArea :title="'Clustta Settings'" :miniDisplay="true" />
				</div>
			</div>

			<div v-if="stage.activeStage === 'account'" class="header-bar-dependencies">
				<ActionButton :icon="getAppIcon('arrow-left')" @click="goToProjects()" v-tooltip="'Back'" />
				<div class="header-area-container">
					<HeaderArea :title="'Account Settings'" :miniDisplay="true" />
				</div>
			</div>


		</div>
		<div class="header-bar-actions">

			<div class="local-project-actions" v-if="stage.selectedStage === 'browser'">
				<!-- <ActionButton v-if="userStore.canDo('view_checkpoint')" :icon="getAppIcon('layers')"
					@click="showProjectCheckpoints()" v-tooltip="'Project Checkpoints'" /> -->
				<ActionButton v-if="userStore.canDo('delete_task')" :icon="getAppIcon('trash')" @click="goToTrash()"
					v-tooltip="'Trash'" />
				<ActionButton v-if="userStore.canDo('create_task')" :icon="getAppIcon('briefcase-cog')"
					@click="goToSettings()" v-tooltip="'Project Settings'" />

			</div>

			<div class="remote-project-actions"
				v-if="projectStore.getActiveProject?.has_remote && projectStore.getActiveProject.is_downloaded && enabledStages.includes(stage.selectedStage)">

				<ActionButton v-if="unSynced" @click="prepResetPopUpModal()" :icon="getAppIcon('undo')"
					:iconAfter="true" v-tooltip="'Revert local changes'" :useBackground="true" label="Revert" />

				<ActionButton v-if="unSynced" :icon="getAppIcon('cloud-up')" :iconAfter="true" :isAlert="unSynced"
					:useBackground="true" label="Save" @click="syncData" v-tooltip="'Save'" />

				<!-- <ActionButton :icon="getAppIcon('bell')" @click="panes.setPaneVisibility('notifications', true)" v-tooltip="'Notifications'"  /> -->
			</div>

		</div>

		<div class="header-bar-actions" v-if="stage.selectedStage === 'trash' && trayStates.trashables.length">
			<ActionButton :icon="getAppIcon('trash')" label="Empty" :showLabel="true" @click="prepEmptyTrashPopUpModal"
				v-tooltip="'Empty trash'" :useBackground="true" :color="'var(--danger)'" />
		</div>
	</div>

</template>

<script setup>


// imports
import { computed, ref, onMounted, toRaw } from 'vue';
import { SyncService } from "@/../bindings/clustta/services";
import { ProjectService } from '@/../bindings/clustta/services/index';
import { syncData } from '@/lib/sync';
import utils from '@/services/utils';

import emitter from '@/lib/mitt';

// state imports
import { useIconStore } from '@/stores/icons';
import { useTrayStates } from '@/stores/TrayStates';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { useProjectStore } from '@/stores/projects';
import { useCollectionStore } from '@/stores/collections';
import { useAssetStore } from '@/stores/assets';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useUserStore } from '@/stores/users';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import WorkspaceTabs from '@/instances/desktop/components/WorkspaceTabs.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// states/stores
const iconStore = useIconStore();
const trayStates = useTrayStates();
const projectStore = useProjectStore();
const panes = usePaneStore();
const stage = useStageStore();
const collectionStore = useCollectionStore();
const assetStore = useAssetStore();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const userStore = useUserStore();

const emits = defineEmits(["update-search", "toggle-search"]);
const enabledStages = ref(['browser', 'projectSettings']);

// computed props
const projectIsActive = computed(() => { return projectStore.getActiveProject && projectStore.getActiveProject.is_downloaded });

// refs
const fullTaskPath = ref(true);

const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

// computed properties
const taskName = computed(() => {
	const task = assetStore.selectedTask;
	if (!task) {
		return
	}
	if (!fullTaskPath.value) {
		return utils.capitalizeStr(task.name)
	} else {
		const fullPath = task.task_path;
		return fullPath.replace(/\//g, ' / ');
	}
});

const toggleFullTaskPath = () => {
	fullTaskPath.value = !fullTaskPath.value;
}

const unSynced = computed(() => { return projectStore.getActiveProject.is_unsynced });

// methods

const revertChanges = async () => {
	let syncOptions = {
		only_latest_checkpoints: false,
		task_dependencies: false,
		tasks: false,
		templates: false,
		force: true,
	};
	await SyncService.PullData(
		projectStore.activeProject.uri, projectStore.getActiveProjectUrl, false, syncOptions
	)
		.then(() => {
			projectStore.activeProject.is_unsynced = false;
			trayStates.refreshData();
			emitter.emit('refresh-browser');
			modals.disableAllModals();
		}).catch((error) => {
			console.error(error.message)
			notificationStore.addNotification(
				"Error Syncing Data",
				error.message,
				"error",
				false
			)
			modals.disableAllModals();
		})
};

const prepResetPopUpModal = () => {
	trayStates.popUpModalIcon = 'revert'
	trayStates.popUpModalTitle = "Revert project";
	trayStates.popUpModalMessage = "Your project will be reverted to the remote version as of the last sync. Continue?";
	trayStates.popUpModalFunction = revertChanges;
	modals.setModalVisibility('popUpModal', true);
};

const prepEmptyTrashPopUpModal = () => {
	trayStates.popUpModalIcon = 'trash'
	trayStates.popUpModalTitle = "Empty Trash";
	trayStates.popUpModalMessage = "This will irreversibly delete all items in trash. Continue?";
	trayStates.popUpModalFunction = emptyTrash;
	modals.setModalVisibility('popUpModal', true);
};

const emptyTrash = async () => {
	await ProjectService.Purge(projectStore.activeProject.uri)
		.then(() => {
			trayStates.trashables = [];
			// trayStates.refreshData();
			modals.disableAllModals();
		}).catch((error) => {
			console.error(error.message)
			notificationStore.addNotification(
				"Error Syncing Data",
				error.message,
				"error",
				false
			)
			modals.disableAllModals();
		})
};

const goToList = () => {
	if (assetStore.selectedTask) {
		const taskId = assetStore.selectedTask.id;
		stage.markedTasks = [taskId];
	}
	stage.setStageVisibility('browser', true);
};

const goToProjects = () => {
	stage.setStageVisibility('projects', true);
	return
	if (projectStore.activeProject) {
		stage.setStageVisibility('browser', true);
	} else {
		stage.setStageVisibility('projects', true);
	}
};

const showProjectCheckpoints = () => {
	
	collectionStore.selectedCollection  = null ;
	assetStore.selectedTask = null ;
	projectStore.selectedUntrackedItem = null;

	if (panes.activeModal !== 'projectCheckpoints' || !panes.showDetailsPane) {
		panes.setPaneVisibility('projectCheckpoints', true);
		panes.showDetailsPane = true;
	} else {
		panes.setPaneVisibility('projectDetails', true);
	}
};

const goToTrash = () => {
	stage.setStageVisibility('trash', true);
};

const goToSettings = () => {
	stage.setStageVisibility('projectSettings', true);
};

</script>

<style scoped>
@import "@/assets/desktop.css";

.desktop-search-bar {
	font-family: 'Inter', sans-serif;
	font-weight: 200;
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
}

.desktop-search-bar::-ms-reveal {
	filter: invert(100%);
	/* color: var(--white); */
}

.desktop-search-bar:hover {
	outline: var(--transparent-line);
	outline-offset: -1px;
}

.desktop-search-bar:focus {
	outline: var(--solid-line);
	outline-offset: -1px;
}

.sync-button {
	/* background-color: ; */
}

.project-unsynced {
	background-color: #bd2d2d;
}

.revert-button {
	color: #E6CC49;
}

.workspace-section {
	display: flex;
	flex-direction: row;
	align-items: center;
	gap: .5rem;
	box-sizing: border-box;
	width: 100%;
}

.breadcrumb-active {
	transform: rotate(90deg);
}

.chevron-icon {
	display: flex;
	box-sizing: border-box;
	overflow: hidden;
	width: min-content;
	height: min-content;
	background-color: teal;
	transition: all 0.3s ease;
	/* transform: rotate(-90deg); */
}

.workspace-name {
	background-color: darkred;
	height: min-content;
	overflow: hidden;
	padding: .2rem;
	display: flex;
	align-items: center;
	width: max-content;
	font-size: 18px;
}

.category-name {
	background-color: darkgrey;
	height: min-content;
	overflow: hidden;
	padding: .2rem;
	display: flex;
	align-items: center;
	width: max-content;
	font-size: 18px;
}

.header-bar {
	color: var(--white);
	padding-top: .3rem;
	width: 100%;
	height: 50px;
	display: flex;
	overflow: hidden;
	box-sizing: border-box;
	align-items: center;
	justify-content: space-between;
	overflow: hidden;
	/* border-bottom: var(--transparent-line); */
	background-color: var(--black);
	/* background-color: rebeccapurple; */
}

.header-bar-overlay {}

.header-bar-breadcrumbs-parent {
	/* padding: .2rem; */
	display: flex;
	overflow: hidden;
	box-sizing: border-box;
	width: max-content;
	width: 100%;
	height: min-content;
	height: 100%;
	/* background-color: firebrick; */
	align-items: center;
	/* gap: .1rem; */

}

.header-bar-tabs {
	/* padding: .2rem; */
	display: flex;
	overflow: hidden;
	box-sizing: border-box;
	width: max-content;
	height: min-content;
	height: 100%;
	/* background-color: forestgreen; */
	align-items: center;
	gap: .1rem;
}

.header-bar-dependencies {
	width: 100%;
	gap: .5rem;
	overflow: hidden;
	display: flex;
	height: 100%;
	align-items: center;
	padding: .5rem;
	box-sizing: border-box;
	/* background-color: goldenrod; */
}

.header-area-container {
	width: 100%;
	/* background-color: forestgreen; */
}

.header-bar-actions {
	align-items: center;
	justify-content: flex-end;
	padding: .5rem;
	display: flex;
	overflow: hidden;
	box-sizing: border-box;
	width: max-content;
	height: min-content;
	height: 100%;
	/* gap: .5rem; */
	min-width: max-content;
	/* background-color: forestgreen; */
}

.search-container {
	display: flex;
	box-sizing: border-box;
	padding: 0 1rem;
	height: 100%;
	align-items: flex-start;
	justify-content: center;
}

.remote-project-actions {
	align-items: center;
	justify-content: flex-end;
	padding: 0px .2rem;
	display: flex;
	overflow: hidden;
	box-sizing: border-box;
	width: max-content;
	height: min-content;
	gap: .4rem;
	min-width: max-content;
	/* background-color: hotpink; */
}

.local-project-actions {
	align-items: center;
	justify-content: flex-end;
	padding: 0px .2rem;
	display: flex;
	overflow: hidden;
	box-sizing: border-box;
	width: max-content;
	height: min-content;
	gap: .4rem;
	min-width: max-content;
	/* background-color: hotpink; */
}

.desktop-search-bar {
	/* height: 35px; */
	width: 100%;
	transition: width 0.2s ease-out;
	border-radius: var(--large-radius);
}

.desktop-search-bar:focus {
	width: 400px;
}
</style>





