<template>
	<div class="header-bar">

		<div class="header-bar-breadcrumbs-parent">

			<div v-if="projectIsActive && stage.activeStage === 'browser'" class="header-bar-tabs">
				<WorkspaceTabs />
			</div>

			<div v-if="activeHeaderConfig" class="header-bar-dependencies">
				<ActionButton :icon="getAppIcon(activeHeaderConfig.icon)" :isInactive="activeHeaderConfig.isInactive" @click="activeHeaderConfig.action?.()" v-tooltip="activeHeaderConfig.tooltip" />
				<div class="header-area-container" @click="activeHeaderConfig.containerClick?.()">
					<HeaderArea :notModal="true" :title="activeHeaderConfig.title" :miniDisplay="true" :customIcon="activeHeaderConfig.customIcon" />
				</div>
			</div>


		</div>
		<div class="header-bar-actions">

			<div class="local-project-actions" v-if="stage.selectedStage === 'browser'">
				<ActionButton v-if="userStore.canDo('delete_task')" :icon="getAppIcon('trash')" @click="goToTrash()"
					v-tooltip="'Trash'" />
				<ActionButton v-if="userStore.canDo('create_task')" :icon="getAppIcon('briefcase-cog')"
					@click="goToSettings()" v-tooltip="'Project Settings'" />

			</div>

			<div class="remote-project-actions" v-if="projectStore.getActiveProject?.has_remote && (projectStore.getActiveProject.is_downloaded || platformStore.isWeb) && enabledStages.includes(stage.selectedStage)">

				<div class="actions-divider" ></div>
				
				<ActionButton :isDisabled="revertButtonDisabled" @click="prepResetPopUpModal()" :icon="getAppIcon('revert')"
				 :noFilter="unSynced"	:iconAfter="true" v-tooltip="revertButtonTooltip"  :useDanger="unSynced"/>

				<ActionButton :isDisabled="syncButtonDisabled" @click="unSynced ? syncData() : pullData()" :icon="getAppIcon(getCloudIcon)"
				 :noFilter="unSynced"	:iconAfter="true" v-tooltip="cloudIconTooltip" :useAlert="unSynced" :useDanger="offline || !studioStore.appOnline" />
				
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
import { SyncService } from "@/services";
import { ProjectService } from '@/services';
import { syncData, pullData } from '@/lib/sync';
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
import { usePlatformStore } from '@/stores/platform';
import { useStudioStore } from '@/stores/studio';

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
const studioStore = useStudioStore();
const userStore = useUserStore();
const platformStore = usePlatformStore();

const emits = defineEmits(["update-search", "toggle-search"]);
const enabledStages = ref(['browser', 'projectSettings']);

// computed props
const projectIsActive = computed(() => { return projectStore.getActiveProject && (platformStore.isWeb || projectStore.getActiveProject.is_downloaded) });

// refs
const fullTaskPath = ref(true);

const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

const getCloudIcon = computed(() => {
	// Check if server is reachable
	if (!studioStore.appOnline || projectStore.getActiveProject?.is_offline) {
		return 'cloud-cancel';
	}
	// Check if any operations are active
	if (!!notificationStore.getProgress.running) {
		return 'cloud-clock';
	}
	// Server is available
	if (!unSynced.value) {
		return 'cloud-down';
	}
	return 'cloud-up';
});

// Returns the tooltip text for the cloud/sync icon.
const cloudIconTooltip = computed(() => {
	if (!studioStore.appOnline) return 'Server unreachable';
	if (projectStore.getActiveProject?.is_offline) return 'Project offline';
	if (!!notificationStore.getProgress.running) return 'Syncing...';
	if (!unSynced.value) return 'Up to date';
	return 'Unsynced changes';
});

// computed properties
const taskName = computed(() => {
	const task = assetStore.selectedAsset;
	if (!task) {
		return
	}
	if (!fullTaskPath.value) {
		return utils.capitalizeStr(task.name)
	} else {
		const fullPath = task.task_path;
		return fullPath?.replace(/\//g, ' / ');
	}
});

const toggleFullTaskPath = () => {
	fullTaskPath.value = !fullTaskPath.value;
}

// Returns the header configuration for the active stage.
const activeHeaderConfig = computed(() => {
	const configs = {
		projects: { icon: 'home', isInactive: true, title: 'Projects' },
		dependencies: { icon: 'chevron-left', action: goToList, title: taskName.value, customIcon: assetStore.selectedAsset?.icon, containerClick: toggleFullTaskPath, tooltip: 'Back' },
		trash: { icon: 'chevron-left', action: goToList, title: 'Trash', tooltip: 'Back' },
		projectSettings: { icon: 'chevron-left', action: goToList, title: 'Project Settings', tooltip: 'Back' },
		studioSettings: { icon: 'chevron-left', action: goToProjects, title: 'Studio Settings', tooltip: 'Back' },
		settings: { icon: 'chevron-left', action: goToProjects, title: 'Clustta Settings', tooltip: 'Back' },
		account: { icon: 'chevron-left', action: goToProjects, title: 'Account Settings', tooltip: 'Back' },
	};
	return configs[stage.activeStage] || null;
});

const unSynced = computed(() => { return projectStore.getActiveProject.is_unsynced });
const offline = computed(() => { return projectStore.getActiveProject?.is_offline });

const revertButtonDisabled = computed(() => {
	return !!notificationStore.getProgress.running || 
	       stage.operationActive || 
	       !projectStore.getActiveProject?.is_downloaded ||
	       !unSynced.value;
});

const revertButtonTooltip = computed(() => {
	if (projectStore.serverIsBusy) return 'Server is busy...';
	if (stage.operationActive) return 'Operation in progress...';
	if (!projectStore.getActiveProject?.is_downloaded) return 'Project not downloaded';
	if (!unSynced.value) return 'No changes to revert';
	return 'Revert local changes';
});

const syncButtonDisabled = computed(() => {
	return !!notificationStore.getProgress.running || 
	       stage.operationActive || 
	       !projectStore.getActiveProject?.is_downloaded 
	    //    !unSynced.value;
});

const syncButtonTooltip = computed(() => {
	if (projectStore.serverIsBusy) return 'Server is busy...';
	if (stage.operationActive) return 'Operation in progress...';
	if (projectStore.getActiveProject?.is_offline) return 'Server Unreachable';
	if (!projectStore.getActiveProject?.is_downloaded) return 'Project not downloaded';
	if (!unSynced.value) return 'Sync';
	return 'Sync';
});

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
	if (assetStore.selectedAsset) {
		const taskId = assetStore.selectedAsset.id;
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

.sync-button-alert :deep(img) {
	filter: brightness(0) saturate(100%) invert(60%) sepia(72%) saturate(489%) hue-rotate(1deg) brightness(92%) contrast(90%);
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
	min-height: 50px;
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
	/* gap: .5rem; */
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

.actions-divider {
	display: flex;
	background-color: var(--light-steel);
	height: 16px;
	width: 1.5px;
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





