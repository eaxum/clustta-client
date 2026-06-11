<template>
	<div v-if="isLoading" class="state-bar">
		<ActionButton :isLoading="true" :icon="getAppIcon('loading')" v-tooltip="$t('components.stateBar.loadingCollectionStates')" />
	</div>

	<div v-else-if="hasData" class="state-bar">
		<ActionButton v-if="collectionStore.collectionStateFlags.has_rebuildable" :icon="getAppIcon('jigsaw')" 
			v-tooltip="$t('components.stateBar.rebuildAll')" :buttonFunction="rebuildAll" />

		<ActionButton v-if="collectionStore.collectionStateFlags.has_untracked && canCreateFromUntrackedHere"
			:icon="getAppIcon('plus-stone')" :useDanger="true" :noFilter="true" v-tooltip="$t('components.stateBar.createCheckpoints')"
			:buttonFunction="prepAllCheckpointModal" />

		<ActionButton v-else-if="collectionStore.collectionStateFlags.has_modified && canCreateCheckpointHere"
			:icon="getAppIcon('plus-stone')" :useAlert="true" :noFilter="true" v-tooltip="$t('components.stateBar.createCheckpoints')"
			:buttonFunction="prepAllCheckpointModal" />

		<ActionButton v-if="collectionStore.collectionStateFlags.has_modified" :icon="getAppIcon('revert')" 
			:useAlert="true" :noFilter="true" v-tooltip="$t('components.stateBar.revertAll')" :buttonFunction="prepResetPopUpModal" />

		<ActionButton v-if="collectionStore.collectionStateFlags.has_outdated" :icon="getAppIcon('circle-check')" 
			:useAlert="true" :noFilter="true" v-tooltip="$t('components.stateBar.updateAll')" :buttonFunction="updateAll" />
	</div>
</template>

<script setup>
// imports
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import { canActInNavigatedCollection, canCreateAssetHere } from '@/lib/permissions';

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
const { t } = useI18n();

// props
const props = defineProps({
	hasData: { type: Boolean, default: false },
});

// computed properties
const isLoading = computed(() => 
	(collectionStore.loadingCollectionStates || assetStore.loadingAssetStates) && props.hasData
);

// Whether the user can create checkpoints on existing modified assets in the
// navigated collection.
const canCreateCheckpointHere = computed(() => {
	collectionStore.navigatedCollection;
	return canActInNavigatedCollection('create_checkpoint');
});

// Whether the user can materialize untracked items as new assets. Requires the
// create_asset role on top of collection scope.
const canCreateFromUntrackedHere = computed(() => {
	collectionStore.navigatedCollection;
	return canCreateAssetHere();
});

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
	trayStates.createMultipleCheckpointsCollectionPath = "";
	modals.setModalVisibility('createMultipleCheckpointsModal', true);
};

// Prepares and shows the revert all changes confirmation modal.
const prepResetPopUpModal = () => {
	clearSelection();
	trayStates.popUpModalIcon = 'revert';
	trayStates.popUpModalTitle = t('components.stateBar.revertAllChangesTitle');
	trayStates.popUpModalMessage = t('components.stateBar.revertAllChangesMessage');
	trayStates.popUpModalFunction = revertAllChanges;
	modals.setModalVisibility('popUpModal', true);
};

// Rebuilds all rebuildable assets in the current view.
const rebuildAll = async () => {
	const path = collectionStore.navigatedCollection?.collection_path;
	const navigatedCollectionId = collectionStore.navigatedCollection?.id;
	notificationStore.cancleFunction = SyncService.CancelSync;
	notificationStore.canCancel = true;
	if (commonStore.activeWorkspace === 'My Assets') {
		const userAssetIds = assetStore.getAssets.filter(asset => asset.assignee_id === userStore.user?.id && !asset.trashed).map(asset => asset.id);
		if (userAssetIds.length) {
			await CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, userAssetIds)
				.then(() => emitter.emit('refresh-browser'))
				.catch((error) => console.error(`Error rebuilding assets:`, error));
		}
	} else {
		await CollectionService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, navigatedCollectionId)
			.then(() => { if (!path) assetStore.rebuildableAssetsPath = []; emitter.emit('refresh-browser'); })
			.catch((error) => notificationStore.errorNotification(t('components.stateBar.errorRebuildingAll'), error));
	}
};

// Reverts all modified assets to their last checkpointed state.
const revertAllChanges = async () => {
	modals.setModalVisibility('popUpModal', false);
	const navigated = collectionStore.navigatedCollection;
	let collectionId = null, targetPath = null;
	if (navigated?.type === 'collection') collectionId = navigated.id;
	else if (navigated?.type === 'untracked_collection') targetPath = navigated.file_path;
	await collectionStore.reloadItemsForCheckpoint(collectionId, targetPath);
	const filteredPaths = assetStore.modifiedAssets.modified.map(asset => asset.display_path);
	if (filteredPaths.length === 0) return;
	await CheckpointService.RevertAssetPaths(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, filteredPaths)
		.then(() => { 
			assetStore.modifiedAssets.modified = assetStore.modifiedAssets.modified.filter((item) => !filteredPaths.includes(item.display_path)); 
			emitter.emit('refresh-browser'); 
		})
		.catch((error) => { notificationStore.errorNotification(t('components.stateBar.failedToRevertAssets'), error); console.error(error); });
};

// Updates all outdated assets to their latest server version.
const updateAll = async () => {
	clearSelection();
	notificationStore.cancleFunction = SyncService.CancelSync;
	notificationStore.canCancel = true;
	const navigated = collectionStore.navigatedCollection;
	let collectionId = null;
	if (navigated?.type === 'collection') collectionId = navigated.id;
	const outdatedAssets = await collectionStore.getOutdatedItems(collectionId);
	const filteredPaths = outdatedAssets.map(asset => asset.asset_path + asset.extension);
	if (filteredPaths.length === 0) return;
	await CheckpointService.RevertAssetPaths(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, filteredPaths)
		.then(() => emitter.emit('refresh-browser'))
		.catch((error) => { notificationStore.errorNotification(t('components.stateBar.failedToRevertAssets'), error); console.error(error); });
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
