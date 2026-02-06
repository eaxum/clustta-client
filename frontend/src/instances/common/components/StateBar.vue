<template>
	<div v-if="isLoading" class="state-bar">
		<ActionButton :isLoading="true" :icon="getAppIcon('loading')" v-tooltip="'Loading collection states'" />
	</div>

	<div v-else-if="hasData" class="state-bar">
		<ActionButton v-if="collectionStore.collectionStateFlags.has_rebuildable" :icon="getAppIcon('jigsaw')" 
			v-tooltip="'Rebuild All'" :buttonFunction="rebuildAll" />

		<ActionButton v-if="collectionStore.collectionStateFlags.has_untracked && userStore.canDo('create_checkpoint')"
			:icon="getAppIcon('layers-plus')" :useDanger="true" :noFilter="true" v-tooltip="'Create Checkpoints'"
			:buttonFunction="prepAllCheckpointModal" />

		<ActionButton v-else-if="collectionStore.collectionStateFlags.has_modified && userStore.canDo('create_checkpoint')"
			:icon="getAppIcon('layers-plus')" :useAlert="true" :noFilter="true" v-tooltip="'Create Checkpoints'"
			:buttonFunction="prepAllCheckpointModal" />

		<ActionButton v-if="collectionStore.collectionStateFlags.has_modified" :icon="getAppIcon('revert')" 
			:useAlert="true" :noFilter="true" v-tooltip="'Revert All'" :buttonFunction="prepResetPopUpModal" />

		<ActionButton v-if="collectionStore.collectionStateFlags.has_outdated" :icon="getAppIcon('circle-check')" 
			:useAlert="true" :noFilter="true" v-tooltip="'Update all'" :buttonFunction="updateAll" />
	</div>
</template>

<script setup>
// imports
import { computed } from 'vue';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// services
import { CheckpointService, CollectionService, SyncService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

// props
const props = defineProps({
	hasData: { type: Boolean, default: false },
});

// computed properties
const isLoading = computed(() => 
	(collectionStore.loadingCollectionStates || assetStore.loadingAssetStates) && props.hasData
);

// methods

// Clears all item selections and resets selection state.
const clearSelection = () => {
	stage.markedItems = [];
	stage.selectedItem = [];
	stage.selectedItems = [];
	stage.firstSelectedItemId = '';
	stage.lastSelectedItemId = '';
	assetStore.selectedAsset = null;
	collectionStore.selectedCollection = null;
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Prepares and shows the create multiple checkpoints modal.
const prepAllCheckpointModal = () => {
	clearSelection();
	trayStates.createMultipleCheckpoints = true;
	trayStates.createMultipleCheckpointsEntityPath = "";
	modals.setModalVisibility('createMultipleCheckpointsModal', true);
};

// Prepares and shows the revert all changes confirmation modal.
const prepResetPopUpModal = () => {
	clearSelection();
	trayStates.popUpModalIcon = 'revert';
	trayStates.popUpModalTitle = "Revert All Changes";
	trayStates.popUpModalMessage = "All Modified tasks will be reverted to their last saved state. Are you sure you want to continue?";
	trayStates.popUpModalFunction = revertAllChanges;
	modals.setModalVisibility('popUpModal', true);
};

// Rebuilds all rebuildable assets in the current view.
const rebuildAll = async () => {
	const path = collectionStore.navigatedCollection?.entity_path;
	const navigatedEntityId = collectionStore.navigatedCollection?.id;
	notificationStore.cancleFunction = SyncService.CancelSync;
	notificationStore.canCancel = true;
	if (commonStore.activeWorkspace === 'My Tasks') {
		const userTaskIds = assetStore.getAssets.filter(task => task.assignee_id === userStore.user?.id && !task.trashed).map(task => task.id);
		if (userTaskIds.length) {
			await CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, userTaskIds)
				.then(() => emitter.emit('refresh-browser'))
				.catch((error) => console.error(`Error rebuilding tasks:`, error));
		}
	} else {
		await CollectionService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, navigatedEntityId)
			.then(() => { if (!path) assetStore.rebuildableAssetsPath = []; emitter.emit('refresh-browser'); })
			.catch((error) => notificationStore.errorNotification("Error Rebuilding All", error));
	}
};

// Reverts all modified tasks to their last checkpointed state.
const revertAllChanges = async () => {
	modals.setModalVisibility('popUpModal', false);
	const navigated = collectionStore.navigatedCollection;
	let collectionId = null, targetPath = null;
	if (navigated?.type === 'entity') collectionId = navigated.id;
	else if (navigated?.type === 'untracked_entity') targetPath = navigated.file_path;
	await collectionStore.reloadItemsForCheckpoint(collectionId, targetPath);
	const filteredPaths = assetStore.modifiedAssets.modified.map(asset => asset.task_path);
	if (filteredPaths.length === 0) return;
	await CheckpointService.RevertTaskPaths(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, filteredPaths)
		.then(() => { 
			assetStore.modifiedAssets.modified = assetStore.modifiedAssets.modified.filter((item) => !filteredPaths.includes(item.task_path)); 
			emitter.emit('refresh-browser'); 
		})
		.catch((error) => { notificationStore.errorNotification("Failed to Revert Tasks", error); console.error(error); });
};

// Updates all outdated tasks to their latest server version.
const updateAll = async () => {
	clearSelection();
	notificationStore.cancleFunction = SyncService.CancelSync;
	notificationStore.canCancel = true;
	const navigated = collectionStore.navigatedCollection;
	let collectionId = null;
	if (navigated?.type === 'entity') collectionId = navigated.id;
	const outdatedTasks = await collectionStore.getOutdatedItems(collectionId);
	const filteredPaths = outdatedTasks.map(task => task.task_path);
	if (filteredPaths.length === 0) return;
	await CheckpointService.RevertTaskPaths(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, filteredPaths)
		.then(() => emitter.emit('refresh-browser'))
		.catch((error) => { notificationStore.errorNotification("Failed to Revert Tasks", error); console.error(error); });
};
</script>

<style scoped>
.state-bar {
	position: relative;
	display: flex;
	align-items: center;
	gap: .4rem;
	width: max-content;
	height: max-content;
}
</style>
