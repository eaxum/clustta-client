<template>
	<div ref="browserRoot" v-esc="cancelOps" v-right-click="openMenu" class="dash-board-root absolute-pane">
		<div ref="browserFilters" class="dash-board-filter">
			<Breadcrumbs />
			<SearchBar ref="searchBar" v-model="commonStore.viewSearchQuery" :placeholder="$t('common.search')" :isLoading="!assetStore.assetsLoaded"
				@input="debouncedUpdateSearch" @clear="clearSearch" />
			<ActionButton v-if="!kanbanView" :icon="getAppIcon('filter')" :buttonFunction="toggleShowFilters" :isActive="showFilters" :showIndicator="filtersActive" v-tooltip="$t('stages.filters')" />
		</div>

		<div class="dash-board-header">
			<FilterBar v-if="showFilters || kanbanView" :kanbanView="kanbanView" />
			<CreateMenu v-else-if="!kanbanView" :kanbanView="kanbanView" :importItems="importItems" :disabled="!canCreateInWorkspace" />
			<StateBar v-if="!showFilters && !kanbanView" :hasData="!!rootData.length" />
			<div v-if="(rootData.length || commonStore.viewSearchQuery.length || commonStore.showUntracked)"
				class="view-options">
				<ViewOptions />
				<ActionButton v-if="!kanbanView" :icon="getAppIcon('arrows-sort')" v-tooltip="$t('stages.sort')" :buttonFunction="openSortMenu" />
				<ActionButton v-if="!kanbanView" :icon="getAppIcon('eye-cog')" v-tooltip="$t('stages.viewOptions')" :buttonFunction="openViewMenu" />
				<ActionButton v-if="!kanbanView && isWideScreen" :icon="panes.showDetailsPane ? getAppIcon('collapse-right') : getAppIcon('collapse-left')"
					v-tooltip="panes.showDetailsPane ? $t('stages.closePane') : $t('stages.openPane')" :buttonFunction="toggleDetailsPane" />
			</div>
		</div>
		
		<div v-if="!kanbanView" ref="assetListContainer" class="browser-root-container" @mousemove="onDrag($event)"
			:class="{ 'browser-root-container-hover-drop': isHovered }" @mouseup="onDragStop($event)" @scroll="disableMenus">
			<GhostItem :data="draggedCard" :index="0" />
			<div class="browser-root-content">
				<div class="left-column" data-file-drop-target>
					<VirtuaScroll v-if="(!assetStore.assetsLoaded || rootData.length) && !commonStore.useGrid" :items="rootData" />
					<GridView v-else-if="!assetStore.assetsLoaded || rootData.length" :rootItems="rootData" />
					<PageState v-else :message="message()" :prompt="prompt()" :illustration="illustration()" />
				</div>
				<DetailsPane v-if="projectStore.getProjects.length && isWideScreen" :isVisible="panes.showDetailsPane" />
			</div>
		</div>
		<div v-else ref="assetListContainer" class="browser-root-container kanban-container">
			<Kanban :filtersActive="filtersActive" :assets="rootData" />
		</div>
	</div>
</template>

<script setup>
// imports
import { computed, nextTick, onBeforeUnmount, onMounted, onUnmounted, ref, watch } from 'vue';
import { Events } from '@wailsio/runtime';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import { getRelativePath } from '@/lib/pathlib';
import { useDebounce } from '@/lib/debounce';
import { useFsWatch } from '@/composables/useFsWatch';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import Breadcrumbs from '@/instances/common/components/Breadcrumbs.vue';
import CreateMenu from '@/instances/desktop/components/CreateMenu.vue';
import DetailsPane from '@/instances/desktop/components/DetailsPane.vue';
import FilterBar from '@/instances/common/components/FilterBar.vue';
import GhostItem from '@/instances/desktop/blocks/GhostItem.vue';
import GridView from '@/instances/desktop/components/GridView.vue';
import Kanban from '@/instances/desktop/components/Kanban.vue';
import PageState from '@/instances/common/components/PageState.vue';
import SearchBar from '@/instances/desktop/components/SearchBar.vue';
import StateBar from '@/instances/common/components/StateBar.vue';
import ViewOptions from '@/instances/common/components/ViewOptions.vue';
import VirtuaScroll from '@/instances/common/components/VirtuaScroll.vue';

// services
import { AssetService, CollectionService, DialogService, FSService, TrashService } from '@/services';

// store imports
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useDependencyStore } from '@/stores/dependency';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useDndStore } from '@/stores/dnd';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { usePaneStore } from '@/stores/panes';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';
import { useScrollStore } from '@/stores/scroll';
import { useStageStore } from '@/stores/stages';
import { useStudioStore } from '@/stores/studio';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';
import { useWorkflowStore } from '@/stores/workflow';
import { useStatusStore } from '@/stores/status';

// stores
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const dependencyStore = useDependencyStore();
const dndStore = useDndStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const panes = usePaneStore();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const scrollStore = useScrollStore();
const stage = useStageStore();
const studioStore = useStudioStore();
const trayStates = useTrayStates();
const userStore = useUserStore();
const workflowStore = useWorkflowStore();
const statusStore = useStatusStore();
const { t } = useI18n();

// refs
const browserFilters = ref(null);
const browserRoot = ref(null);
const observer = ref(null);
const rootData = ref([]);
const screenWidth = ref(window.innerWidth);
const searchBar = ref(null);
const showFilters = ref(false);
const assetListContainer = ref(null);

// computed properties
const draggedCard = computed(() => dndStore.allViewItems?.find(card => card.id === dndStore.draggedItemId));

const collectionExpanded = computed(() => Object.keys(stage.expandedCollections).length);

const filtersActive = computed(() => {
	const assigneeFilters = commonStore.hasAssignees || commonStore.noAssignees;
	const collectionFilters = commonStore.collectionFilters.length > 0;
	const assetFilters = commonStore.assetFilters.length > 0;
	const resourceFilters = commonStore.resourceFilters.length > 0;
	const generalFilterActive = !(commonStore.showCollections && commonStore.showAssets && commonStore.showResources && commonStore.showChildCollections && commonStore.showChildAssets && commonStore.showDependencies && !commonStore.onlyAssets);
	return assigneeFilters || collectionFilters || assetFilters || resourceFilters || generalFilterActive;
});

const isDefaultWorkspace = computed(() => commonStore.activeWorkspace === 'Default');

// Whether the active workspace allows creating items (Default or a collection-based workspace).
const canCreateInWorkspace = computed(() => {
	if (isDefaultWorkspace.value) return true;
	const workspace = commonStore.workspaces.find(w => w.name === commonStore.activeWorkspace);
	return !!(workspace && workspace.collection);
});

const isHovered = computed(() => dndStore.isDropHovering && dndStore.targetItemId === null);

const kanbanView = computed(() => commonStore.viewMode === 'kanban');

const isWideScreen = computed(() => screenWidth.value >= 1000);

const operationsActive = computed(() => {
	return stage.operationActive || !!modals.activeModal || !!menu.activeMenu || !assetStore.assetsLoaded || stage.activeStage !== 'browser'
});

// methods

// Adds an collection dependency to a asset. Returns true on success.
const addCollectionDependency = async (assetId, dependencyId, dependencyTypeId) => {
	try {
		await AssetService.AddCollectionDependency(projectStore.activeProject.uri, assetId, dependencyId, dependencyTypeId);
		notificationStore.addNotification(t('stages.dependencyAdded'), "", "success");
		return true;
	} catch (error) {
		console.log(error);
		notificationStore.errorNotification(t('stages.errorAddingDependencies'), error);
		return false;
	}
};

// Adds a asset dependency between two assets. Returns true on success.
const addDependency = async (assetId, dependencyId, dependencyTypeId) => {
	try {
		await AssetService.AddAssetDependency(projectStore.activeProject.uri, assetId, dependencyId, dependencyTypeId);
		notificationStore.addNotification(t('stages.dependencyAdded'), "", "success");
		return true;
	} catch (error) {
		console.log(error);
		notificationStore.errorNotification(t('stages.errorAddingDependencies'), error);
		return false;
	}
};

// Cancels active operations like search or drag-and-drop.
const cancelOps = () => {
	if (commonStore.viewSearchQuery) clearSearch();
	if (!dndStore.altKeyActive) dndStore.resetValues();
};

// Changes the parent collection of one or more collections. Returns true on success.
const changeCollectionParent = async (collectionIds, parentId) => {
	try {
		await CollectionService.ChangeCollectionParent(projectStore.activeProject.uri, collectionIds, parentId);
		notificationStore.addNotification(t('stages.movedSuccessfully'), "", "success");
		return true;
	} catch (error) {
		console.error(error);
		notificationStore.errorNotification(t('stages.errorChangingCollectionParent'), error);
		return false;
	}
};

// Moves one or more assets to a different collection. Returns true on success.
// Duplicate name+extension validation is handled at the service layer.
const changeAssetCollection = async (assetIds, collectionId) => {
	try {
		await AssetService.ChangeAssetCollection(projectStore.activeProject.uri, assetIds, collectionId);
		notificationStore.addNotification(t('stages.movedSuccessfully'), "", "success");
		return true;
	} catch (error) {
		notificationStore.errorNotification(t('stages.errorMovingAssets'), error);
		return false;
	}
};

// Clears the search query and refreshes the view.
const clearSearch = async () => { commonStore.viewSearchQuery = ""; await softRefresh(); };

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

// Collapses all expanded collections and clears selection.
const collapseAll = () => {
	stage.expandedCollections = {};
	stage.markedItems = [];
	stage.firstSelectedCollectionId = '';
	collectionStore.selectedCollection = null;
};

// Opens the application selection modal to create a new asset.
const createAsset = () => { clearSelection(); modals.setModalVisibility('selectAppModal', true); };

// Opens the create collection modal.
const createCollection = () => { if (!stage.groupItems) clearSelection(); modals.setModalVisibility('createCollectionModal', true); };

// Opens the add web link modal.
const createWebLink = () => { clearSelection(); modals.setModalVisibility('addWebLinkModal', true); };

// Deletes all selected items including assets, collections, and untracked files.
const deleteMultipleItems = async () => {
	panes.setPaneVisibility('projectDetails', true);
	stage.operationActive = true;
	stage.firstSelectedItemId = '';
	stage.lastSelectedItemId = '';
	assetStore.selectedAsset = null;
	collectionStore.selectedCollection = null;
	const allItemsToDelete = dndStore.allViewItems.filter((item) => stage.markedItems.includes(item.id));
	const assetsToDelete = allItemsToDelete.filter((item) => item.type === 'asset');
	const collectionsToDelete = allItemsToDelete.filter((item) => item.type === 'collection');
	const untrackedAssetsToDelete = allItemsToDelete.filter((item) => item.type === 'untracked_asset');
	const untrackedCollectionsToDelete = allItemsToDelete.filter((item) => item.type === 'untracked_collection');
	await deleteMultipleAssets(assetsToDelete.map((item) => item.id));
	await deleteMultipleCollections(collectionsToDelete.map((item) => item.id));
	await deleteMultipleUntrackedAssets(untrackedAssetsToDelete);
	await deleteMultipleUntrackedCollections(untrackedCollectionsToDelete);
	stage.markedItems = [];
	stage.selectedItems = [];
	stage.markedAssets = [];
	stage.markedCollections = [];
	stage.operationActive = false;
	modals.setModalVisibility('popUpModal', false);
};

// Moves multiple collections to trash.
const deleteMultipleCollections = async (collectionIds) => {
	for (let collectionId of collectionIds) {
		await CollectionService.DeleteCollection(projectStore.activeProject.uri, collectionId, true)
			.then(async () => { await collectionStore.markCollectionAsDeleted(collectionId); notificationStore.addNotification(t('stages.collectionMovedToTrash'), '', "success", false); })
			.catch((error) => { console.log(error); notificationStore.errorNotification(t('stages.collectionsFailedToDelete'), error); });
	}
};

// Moves multiple assets to trash.
const deleteMultipleAssets = async (assetIds) => {
	for (let assetId of assetIds) {
		await AssetService.DeleteAsset(projectStore.activeProject.uri, assetId, true)
			.then(async () => { softRefresh(); notificationStore.addNotification(t('stages.assetsMovedToTrash'), '', "success", false); })
			.catch((error) => { console.log(error); notificationStore.errorNotification(t('stages.assetsFailedToDelete'), error); });
	}
};

// Permanently deletes multiple untracked collection folders.
const deleteMultipleUntrackedCollections = async (untrackedCollections) => {
	for (let untrackedCollection of untrackedCollections) {
		FSService.DeleteFolder(untrackedCollection.file_path);
		projectStore.removeUntrackedCollection(untrackedCollection.id);
	}
};

// Permanently deletes multiple untracked asset files.
const deleteMultipleUntrackedAssets = async (untrackedAssets) => {
	for (let untrackedAsset of untrackedAssets) {
		await FSService.DeleteFile(untrackedAsset.file_path);
		projectStore.removeUntrackedAsset(untrackedAsset.id);
	}
};

// Detects Alt key state for drag-and-drop modifier behavior.
const detectModifier = (event) => { dndStore.altKeyActive = event.getModifierState('Alt'); };

// Hides all context menus.
const disableMenus = () => { menu.disableAllMenus(); };

// Duplicates the selected asset in the database and copies the physical file.
const duplicateAsset = async () => {
	const selectedItemId = stage.markedItems[0];
	const selectedItem = dndStore.allViewItems.find(item => item.id === selectedItemId);
	if (!selectedItem || selectedItem.type !== 'asset') return;
	try {
		stage.operationActive = true;
		stage.markedItems = [];
		assetStore.selectedAsset = null;
		await AssetService.DuplicateAsset(projectStore.activeProject.uri, selectedItemId, '')
			.then(async (duplicatedAsset) => {
				try { await FSService.DuplicateFile(selectedItem.file_path, duplicatedAsset.file_path); }
				catch (fileError) { console.warn('Physical file duplication failed (asset may be rebuildable):', fileError); }
				await refresh();
				assetStore.selectAsset(duplicatedAsset);
				stage.selectedItem = duplicatedAsset;
				stage.markedItems = [duplicatedAsset.id];
				stage.lastSelectedItemId = "";
				stage.firstSelectedItemId = duplicatedAsset.id;
				notificationStore.addNotification(t('stages.assetDuplicated'), '', "success", false);
			});
	} catch (error) {
		console.error('Error duplicating asset:', error);
		notificationStore.errorNotification(t('stages.failedToDuplicateAsset'), error.message || error);
	} finally { stage.operationActive = false; }
};

// Expands all collections in the view.
const expandAll = () => {
	const collections = collectionStore.getCollections;
	const expandedCollections = {};
	for (let i = 0; i < collections.length; i++) {
		expandedCollections[collections[i].id] = { "height": 0, "collection_path": collections[i].collection_path };
	}
	stage.expandedCollections = expandedCollections;
};

// Frees up disk space by removing files for selected assets and collections.
const freeUpSpace = async () => {
	panes.setPaneVisibility('projectDetails', true);
	stage.operationActive = true;
	stage.firstSelectedItemId = '';
	stage.lastSelectedItemId = '';
	assetStore.selectedAsset = null;
	collectionStore.selectedCollection = null;
	const allItemsToDelete = dndStore.allViewItems.filter((item) => stage.markedItems.includes(item.id));
	await freeUpMultipleAssetSpace(allItemsToDelete.filter((item) => item.type === 'asset'));
	await freeUpMultipleCollectionSpace(allItemsToDelete.filter((item) => item.type === 'collection'));
	await deleteMultipleUntrackedAssets(allItemsToDelete.filter((item) => item.type === 'untracked_asset'));
	await deleteMultipleUntrackedCollections(allItemsToDelete.filter((item) => item.type === 'untracked_collection'));
	stage.markedItems = [];
	stage.selectedItems = [];
	stage.markedAssets = [];
	stage.markedCollections = [];
	stage.operationActive = false;
	modals.setModalVisibility('popUpModal', false);
};

// Deletes collection folders to free up disk space.
const freeUpMultipleCollectionSpace = async (collections) => {
	for (const collection of collections) {
		let collectionDir = collection.file_path.replace(/\\/g, '/');
		await FSService.DeleteFolder(collectionDir)
			.then(() => assetStore.refreshCollectionFilesStatus(collection.id))
			.catch((error) => console.error(error));
	}
};

// Deletes physical files for assets to free up disk space.
const freeUpMultipleAssetSpace = async (selectedAssets) => {
	const fileStatus = ['missing', 'rebuildable'];
	let assetIds = [];
	for (let asset of selectedAssets) { if (!fileStatus.includes(asset.file_status)) assetIds.push(asset.id); }
	for (const assetId of assetIds) {
		let asset = assetStore.getAssets.find((item) => item.id === assetId);
		let assetPath = asset.file_path.replace(/\\/g, '/');
		await FSService.DeleteFile(assetPath)
			.then(() => { asset.file_status = 'rebuildable'; })
			.catch((error) => console.error(error));
	}
};

// Generates a unique file path by appending a counter if the file already exists.
const generateUniqueDestinationPath = async (directory, fileName) => {
	const originalPath = await FSService.JoinPath(directory, fileName);
	const exists = await FSService.Exists(originalPath);
	if (!exists) return originalPath;
	const baseName = fileName.includes('.') ? fileName.substring(0, fileName.lastIndexOf('.')) : fileName;
	const extension = fileName.includes('.') ? fileName.substring(fileName.lastIndexOf('.')) : '';
	let counter = 1;
	let newPath;
	do {
		const newFileName = `${baseName} (${counter})${extension}`;
		newPath = await FSService.JoinPath(directory, newFileName);
		const pathExists = await FSService.Exists(newPath);
		if (!pathExists) return newPath;
		counter++;
	} while (counter < 100);
	const timestamp = Date.now();
	return await FSService.JoinPath(directory, `${baseName}_${timestamp}${extension}`);
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Returns the current directory path based on navigation context.
const getCurrentDirectory = () => {
	if (commonStore.navigatorMode && collectionStore.navigatedCollection) return collectionStore.navigatedCollection.file_path;
	return projectStore.activeProject?.working_directory;
};

// Handles clicks outside of items to clear selection.
const handleClickOutside = (event, rightClick = false) => {
	if (event) {
		if (!event.shiftKey || !event.ctrlKey) {
			if (!event.target.closest('.collection-item-main')) {
				stage.markedItems = []; stage.selectedItems = []; stage.markedCollections = []; stage.firstSelectedItemId = ''; stage.lastSelectedItemId = '';
				stage.selectedItem = null; assetStore.selectedAsset = null; collectionStore.selectedCollection = null; projectStore.selectedUntrackedItem = null;
			}
			if (!event.target.closest('.asset-item-main')) {
				stage.markedItems = []; stage.selectedItems = []; stage.markedCollections = []; stage.firstSelectedItemId = ''; stage.lastSelectedItemId = '';
				stage.selectedItem = null; assetStore.selectedAsset = null; projectStore.selectedUntrackedItem = null;
			}
		}
	}
};

// Handles root data updates from emitter events for individual item property changes.
const handleUpdateRootData = (eventData) => {
	if (Array.isArray(eventData)) {
		eventData.forEach(({ itemId, updates }) => {
			const itemIndex = rootData.value.findIndex(item => item.id === itemId);
			if (itemIndex !== -1 && updates && Array.isArray(updates)) {
				updates.forEach(update => { if (update.property && update.value !== undefined) rootData.value[itemIndex][update.property] = update.value; });
			}
		});
	} else {
		const { itemId, property, value, updates } = eventData;
		const itemIndex = rootData.value.findIndex(item => item.id === itemId);
		if (itemIndex !== -1) {
			if (property && value !== undefined) rootData.value[itemIndex][property] = value;
			if (updates && Array.isArray(updates)) updates.forEach(update => { rootData.value[itemIndex][update.property] = update.value; });
		}
	}
	emitter.emit('get-project-data');
	collectionStore.loadCollectionStateFlags();
};

// Folder watched at the root level. Falls back to the project working dir.
const rootWatchPath = computed(() =>
	collectionStore.navigatedCollection?.file_path
		?? projectStore.activeProject?.working_directory
		?? null
);

let rootFsToken = 0;
let rootFsState = null;

// Refreshes the navigated collection in response to fs changes.
// Tracked collections fetch state diffs; untracked rescan the folder.
const handleRootFsChange = async () => {
	if (!projectStore.activeProject) return;
	const project = projectStore.activeProject;
	const nav = collectionStore.navigatedCollection;

	const myToken = ++rootFsToken;

	if (nav?.type === 'untracked_collection') {
		try {
			const children = await CollectionService.GetCollectionChildren(
				project.uri,
				nav.id,
				project.working_directory,
				nav.file_path,
				project.ignore_list,
				true
			);
			if (myToken !== rootFsToken) return;

			const untracked = [
				...(children.untracked_collections || children.untracked_folders || []),
				...(children.untracked_assets || children.untracked_files || [])
			];
			await assetStore.processUntrackedAssetsIcons(untracked);
			emitter.emit('update-untracked-items', untracked);
		} catch (err) {
			console.error('handleRootFsChange (untracked) failed', err);
		}
		return;
	}

	try {
		const state = await CollectionService.GetCollectionChildrenState(
			project.uri,
			nav?.id ?? '',
			project.working_directory,
			project.ignore_list
		);
		if (myToken !== rootFsToken) return;

		const snapshot = {
			modified: (state.modified_assets || []).map(a => a.id).sort(),
			normal: (state.normal_assets || []).map(a => a.id).sort(),
			outdated: (state.outdated_assets || []).map(a => a.id).sort(),
			rebuildable: (state.rebuildable_assets || []).map(a => a.id).sort(),
			untracked_files: (state.untracked_files || []).map(f => f.id).sort(),
			untracked_folders: (state.untracked_folders || []).map(f => f.id).sort()
		};
		if (JSON.stringify(rootFsState) === JSON.stringify(snapshot)) return;
		rootFsState = snapshot;

		const statusUpdates = [
			...(state.normal_assets || []).map(a => ({ itemId: a.id, updates: [{ property: 'file_status', value: 'normal' }] })),
			...(state.modified_assets || []).map(a => ({ itemId: a.id, updates: [{ property: 'file_status', value: 'modified' }] })),
			...(state.outdated_assets || []).map(a => ({ itemId: a.id, updates: [{ property: 'file_status', value: 'outdated' }] })),
			...(state.rebuildable_assets || []).map(a => ({ itemId: a.id, updates: [{ property: 'file_status', value: 'rebuildable' }] }))
		];
		if (statusUpdates.length) emitter.emit('update-root-data', statusUpdates);

		const untracked = [
			...(state.untracked_folders || []),
			...(state.untracked_files || [])
		];
		await assetStore.processUntrackedAssetsIcons(untracked);
		emitter.emit('update-untracked-items', untracked);
	} catch (err) {
		console.error('handleRootFsChange failed', err);
	}
};

useFsWatch(rootWatchPath, handleRootFsChange);

// Replaces untracked items in root data with updated list from emitter events.
// Rebuilds rootData through sortItems so order matches the initial load and item
// identities stay put — otherwise reordering would force VirtuaList to remount
// rows and leave stale skeletons on expanded untracked collections.
const handleUpdateUntrackedItems = (untrackedItems) => {
	if (!untrackedItems) return;
	const trackedCollections = rootData.value.filter(item => item.type === 'collection');
	const trackedAssets = rootData.value.filter(item => item.type === 'asset');
	const untrackedCollections = untrackedItems.filter(item => item.type === 'untracked_collection');
	const untrackedAssets = untrackedItems.filter(item => item.type === 'untracked_asset');
	rootData.value = sortItems(trackedCollections, trackedAssets, untrackedCollections, untrackedAssets);
	emitter.emit('get-project-data');
	collectionStore.loadCollectionStateFlags();
};

// Returns the empty state illustration path.
const illustration = () => commonStore.viewSearchQuery ? '/page-states/resources.png' : '/page-states/assets.png';

// Checks if an editable element (input, textarea, contenteditable) is currently focused.
const isEditableElementFocused = () => {
	const activeElement = document.activeElement;
	if (!activeElement) return false;
	const tagName = activeElement.tagName;
	if (tagName === 'INPUT' || tagName === 'TEXTAREA') return true;
	if (activeElement.isContentEditable) return true;
	return false;
};

// Imports files or folders from the file system into the current directory.
const importItems = async () => {
	try {
		let selectedPaths;
		try { selectedPaths = await DialogService.SelectFilesDialog(); } catch (error) { return; }
		if (!selectedPaths || selectedPaths.length === 0) return;
		const currentDirectory = getCurrentDirectory();
		if (!currentDirectory) { notificationStore.errorNotification(t('stages.couldNotDetermineCurrentDirectory'), ""); return; }
		stage.operationActive = true;
		await FSService.MakeDirs(currentDirectory);
		let successCount = 0, failureCount = 0;
		const errors = [];
		for (const sourcePath of selectedPaths) {
			try {
				const isFile = await FSService.IsFile(sourcePath);
				const itemName = await FSService.BaseName(sourcePath);
				const destinationPath = await generateUniqueDestinationPath(currentDirectory, itemName);
				if (isFile) await FSService.DuplicateFile(sourcePath, destinationPath);
				else await FSService.DuplicateFolder(sourcePath, destinationPath);
				successCount++;
			} catch (error) {
				failureCount++;
				const itemName = await FSService.BaseName(sourcePath).catch(() => sourcePath);
				errors.push(`${itemName}: ${error.message || error}`);
			}
		}
		if (successCount > 0) notificationStore.addNotification(successCount === 1 ? t('stages.itemImportedSuccessfully') : t('stages.itemsImportedSuccessfully', { count: successCount }), "", "success");
		if (failureCount > 0) notificationStore.errorNotification(failureCount === 1 ? t('stages.itemFailedToImport') : t('stages.itemsFailedToImport', { count: failureCount }), errors.join("\n"));
		if (successCount > 0) await softRefresh();
	} catch (error) { notificationStore.errorNotification(t('stages.errorImportingItems'), error.message || error); }
	finally { stage.operationActive = false; }
};

// Returns the empty state message based on current view context.
const message = () => {
	const searching = commonStore.viewSearchQuery;
	const myAssetsWorkspace = commonStore.activeWorkspace === 'My Assets';
	if (searching) return t('stages.noResultsFound');
	if (isDefaultWorkspace.value && filtersActive.value) return t('stages.noResultsMatchFilters');
	if (myAssetsWorkspace) return t('stages.noAssetsAssigned');
	if (!isDefaultWorkspace.value) return t('stages.nothingInWorkspace');
	return t('stages.nothingToSeeHere');
};

// Handles drag movement events.
const onDrag = (e) => dndStore.onDrag(e);

// Handles drag stop events and processes item moves, parent changes, or dependency additions.
const onDragStop = async (event) => {
	if (kanbanView.value) return;
	if (dndStore.draggedItemId === null) return;
	document.documentElement.style.cursor = 'default';
	const dropTarget = dndStore.itemRefs[dndStore.targetItemId];
	const draggedItem = dndStore.itemRefs[dndStore.draggedItemId];
	const targetCollection = dndStore.allViewItems.find((item) => item.id === dndStore.targetItemId);
	stage.operationActive = true;

	// Initialize cardRect with fallback to ghost position
	let cardRect = draggedItem?.getBoundingClientRect() ?? { x: dndStore.ghostCardStyle.pos.x, y: dndStore.ghostCardStyle.pos.y };

	const draggedItemIds = stage.markedItems;
	const draggedItems = draggedItemIds.map(id => dndStore.allViewItems.find(item => item.id === id)).filter(Boolean);

	// Collect items by type for batch operations
	const collectionIdsToMove = [];
	const assetIdsToMove = [];
	const renameOperations = [];
	const dependencyUpdates = { assetId: null, dependencies: [], collectionDependencies: [] };
	let needsRefresh = false;

	for (const draggedCollection of draggedItems) {
		if (event.altKey) {
			if (draggedItem) cardRect = draggedItem.getBoundingClientRect();
			if (draggedCollection.collection_type_id) collectionIdsToMove.push(draggedCollection.id);
			else if (draggedCollection.asset_type_id) assetIdsToMove.push(draggedCollection.id);
			else {
				let extension = draggedCollection.type === 'untracked_asset' ? draggedCollection.extension : '';
				let fullName = draggedCollection.name + extension;
				await FSService.MakeDirs(projectStore.activeProject.working_directory);
				let newPath = await FSService.JoinPath(projectStore.activeProject.working_directory, fullName);
				renameOperations.push({ oldPath: draggedCollection.file_path, newPath });
			}
		} else if (dndStore.isOverlapping && dropTarget) {
			cardRect = dropTarget.getBoundingClientRect();
			if (draggedCollection.id !== dndStore.targetItemId) {
				if (targetCollection.type === 'collection') {
					if (draggedCollection.type === 'collection') collectionIdsToMove.push(draggedCollection.id);
					else if (draggedCollection.type === 'asset') assetIdsToMove.push(draggedCollection.id);
					else {
						let collection = await CollectionService.GetCollectionByID(projectStore.activeProject.uri, dndStore.targetItemId);
						await FSService.MakeDirs(collection.file_path);
						let extension = draggedCollection.type === 'untracked_asset' ? draggedCollection.extension : '';
						let fullName = draggedCollection.name + extension;
						let newPath = await FSService.JoinPath(collection.file_path, fullName);
						renameOperations.push({ oldPath: draggedCollection.file_path, newPath });
					}
				} else if (targetCollection?.asset_type_id) {
					let dependencyTypeId = dependencyStore.dependency_types.find(item => item.name === "linked").id;
					dependencyUpdates.assetId = dndStore.targetItemId;
					if (draggedCollection.collection_type_id) {
						const success = await addCollectionDependency(dndStore.targetItemId, draggedCollection.id, dependencyTypeId);
						if (success) dependencyUpdates.collectionDependencies.push(draggedCollection.id);
					} else if (draggedCollection.asset_type_id) {
						const success = await addDependency(dndStore.targetItemId, draggedCollection.id, dependencyTypeId);
						if (success) dependencyUpdates.dependencies.push(draggedCollection.id);
					} else if (!draggedCollection.item_type) {
						const success = await addDependency(dndStore.targetItemId, draggedCollection.id, dependencyTypeId);
						if (success) dependencyUpdates.dependencies.push(draggedCollection.id);
					}
				} else if (targetCollection?.type === 'untracked_collection') {
					if (!draggedCollection.collection_type_id && !draggedCollection.asset_type_id && (draggedCollection.type === 'untracked_asset' || draggedCollection.type === 'untracked_collection')) {
						let extension = draggedCollection.type === 'untracked_asset' ? draggedCollection.extension : '';
						let fullName = draggedCollection.name + extension;
						let newPath = await FSService.JoinPath(targetCollection.file_path, fullName);
						renameOperations.push({ oldPath: draggedCollection.file_path, newPath });
					}
				}
			} else if (draggedItem) cardRect = draggedItem.getBoundingClientRect();
		} else if (draggedItem) cardRect = draggedItem.getBoundingClientRect();
	}

	// Execute batch operations for file moves (requires refresh)
	const targetParentId = event.altKey ? '' : dndStore.targetItemId;
	if (collectionIdsToMove.length) {
		const success = await changeCollectionParent(collectionIdsToMove, targetParentId);
		if (success) needsRefresh = true;
	}
	if (assetIdsToMove.length) {
		const success = await changeAssetCollection(assetIdsToMove, targetParentId);
		if (success) needsRefresh = true;
	}
	if (renameOperations.length) {
		try {
			await FSService.RenameBatch(JSON.stringify(renameOperations));
			needsRefresh = true;
		} catch (error) {
			notificationStore.errorNotification(t('stages.errorMovingFiles'), error);
		}
	}

	// Emit dependency updates (no refresh needed, just update item data)
	if (dependencyUpdates.assetId && (dependencyUpdates.dependencies.length || dependencyUpdates.collectionDependencies.length)) {
		const targetAsset = dndStore.allViewItems.find(item => item.id === dependencyUpdates.assetId);
		if (targetAsset) {
			const currentDeps = targetAsset.dependencies || [];
			const currentCollectionDeps = targetAsset.collection_dependencies || [];
			const updates = [
				{ property: 'dependencies', value: [...currentDeps, ...dependencyUpdates.dependencies] },
				{ property: 'collection_dependencies', value: [...currentCollectionDeps, ...dependencyUpdates.collectionDependencies] }
			];
			emitter.emit('update-root-data', { itemId: dependencyUpdates.assetId, updates });
		}
	}

	setTimeout(() => dndStore.resetValues(), 100);
	dndStore.ghostCardStyle.leaving = true;
	let xOffset = cardRect.x - dndStore.ghostCardStyle.pos.x;
	let yOffset = cardRect.y - dndStore.ghostCardStyle.pos.y;
	dndStore.ghostCardStyle.transform = `scale(1) translate(${xOffset}px, ${yOffset}px)`;
	if (needsRefresh) softRefresh();
	stage.operationActive = false;
};

// Opens the project context menu on right-click.
const openMenu = (event) => {
	assetStore.selectedAsset = null;
	collectionStore.selectedCollection = null;
	projectStore.selectedUntrackedItem = null;
	handleClickOutside(event, true);
	if (kanbanView.value) return;
	menu.showContextMenu(event, 'projectMenu', true);
};

// Opens the sort options menu.
const openSortMenu = (event) => {
	menu.showContextMenu(event, 'sortMenu', true, true);
};

// Opens the view options menu.
const openViewMenu = (event) => {
	menu.showContextMenu(event, 'viewMenu', true, true);
};

// Prepares and shows the create multiple checkpoints modal.
const prepAllCheckpointModal = () => {
	clearSelection();
	trayStates.createMultipleCheckpoints = true;
	trayStates.createMultipleCheckpointsCollectionPath = "";
	modals.setModalVisibility('createMultipleCheckpointsModal', true);
};

// Prepares and shows the delete multiple items confirmation modal.
const prepDeleteMultipleItemsPopUpModal = () => {
	const numberOfItems = stage.markedItems.length;
	trayStates.popUpModalTitle = t('stages.deleteNItems', { count: numberOfItems });
	trayStates.popUpModalMessage = t('stages.deleteUntrackedItemsConfirmation');
	trayStates.popUpModalIcon = 'trash';
	trayStates.popUpModalFunction = deleteMultipleItems;
	modals.setModalVisibility('popUpModal', true);
};

// Prepares and shows the free up space confirmation modal.
const prepFreeUpSpacePopUpModal = () => {
	trayStates.popUpModalTitle = t('stages.freeUpSpace');
	trayStates.popUpModalMessage = t('stages.freeUpSpaceConfirmation');
	trayStates.popUpModalIcon = 'broom';
	trayStates.popUpModalFunction = freeUpSpace;
	modals.setModalVisibility('popUpModal', true);
};

// Returns the empty state prompt text.
const prompt = () => {
	if (commonStore.viewSearchQuery) return '';
	if (!isDefaultWorkspace.value || filtersActive.value) return '';
	return t('stages.rightClickToCreate');
};

// Status priority map for sorting (lower number = higher priority).
const statusPriority = { 'retake': 1, 'wip': 2, 'wfa': 3, 'ready': 4, 'todo': 5, 'done': 6 };

// Sorts items based on the current sort settings in commonStore.
// Collections and assets are sorted separately to maintain grouping.
const sortItems = (collections, assets, untrackedCollections, untrackedAssets) => {
	const sortBy = commonStore.sortBy;
	const sortOrder = commonStore.sortOrder;
	const multiplier = sortOrder === 'asc' ? 1 : -1;

	const sortByName = (a, b) => {
		const nameA = (a.name || '').toLowerCase();
		const nameB = (b.name || '').toLowerCase();
		return multiplier * nameA.localeCompare(nameB);
	};

	const sortByStatus = (a, b) => {
		const statusA = (a.status_short_name || '').toLowerCase();
		const statusB = (b.status_short_name || '').toLowerCase();
		const priorityA = statusPriority[statusA] ?? 99;
		const priorityB = statusPriority[statusB] ?? 99;
		if (priorityA !== priorityB) return multiplier * (priorityA - priorityB);
		return sortByName(a, b);
	};

	const sortFn = sortBy === 'status' ? sortByStatus : sortByName;

	return [
		...(collections ? [...collections].sort(sortByName) : []),
		...(untrackedCollections ? [...untrackedCollections].sort(sortByName) : []),
		...(assets ? [...assets].sort(sortFn) : []),
		...(untrackedAssets ? [...untrackedAssets].sort(sortByName) : [])
	];
};

// Full refresh: reloads project data, fetches all children, processes icons/previews, and updates state flags.
const refresh = async () => {
	if (kanbanView.value){
		await trayStates.refreshData();
		return;
	} 
	assetStore.assetsLoaded = false;
	await projectStore.refreshActiveProject();
	await trayStates.refreshData();
	let children;
	let project = projectStore.activeProject;
	if (!commonStore.navigatorMode) children = await CollectionService.GetCollectionChildren(project.uri, "root", project.working_directory, project.working_directory, project.ignore_list, false);
	else {
		const navigatedCollectionId = collectionStore.navigatedCollection?.id;
		const collection_file_path = collectionStore.navigatedCollection?.file_path;
		children = await CollectionService.GetCollectionChildren(project.uri, navigatedCollectionId, project.working_directory, collection_file_path, project.ignore_list, false);
	}
	await assetStore.processAssetsIconsAndPreviews(children.assets);
	await assetStore.processUntrackedAssetsIcons(children.untracked_assets);
	rootData.value = sortItems(children.collections, children.assets, children.untracked_collections, children.untracked_assets);
	assetStore.assetsLoaded = true;
	collectionStore.loadCollectionStateFlags();
	await nextTick();
	dndStore.triggerDomUpdate();
};

// Lightweight refresh: fetches children with search/filter support, processes icons, updates root data and state flags.
const softRefresh = async () => {
	assetStore.assetsLoaded = false;
	let children = {};
	let project = projectStore.activeProject;
	const searching = commonStore.viewSearchQuery.toLowerCase();
	if (searching || filtersActive.value) {
		let collections, assets;
		if (!commonStore.navigatorMode) {
			if (!searching) {
				const rootItems = await CollectionService.GetCollectionChildren(project.uri, "root", project.working_directory, project.working_directory, project.ignore_list, false);
				collections = rootItems['collections'];
				assets = commonStore.onlyAssets ? await AssetService.GetAssets(project.uri) : rootItems['assets'];
			} else {
				collections = await CollectionService.GetCollections(project.uri);
				assets = await AssetService.GetAssets(project.uri);
			}
			collections = commonStore.onlyAssets ? [] : collections;
		} else {
			const navigatedCollectionId = collectionStore.navigatedCollection?.id;
			const collection_file_path = collectionStore.navigatedCollection?.file_path;
			const collectionItems = await CollectionService.GetCollectionChildren(project.uri, navigatedCollectionId, project.working_directory, collection_file_path, project.ignore_list, false);
			collections = collectionItems['collections'];
			assets = collectionItems['assets'];
		}
		children['collections'] = await collectionStore.filterCollections(collections);
		children['assets'] = await assetStore.filterAssets(assets);
	} else {
		if (!commonStore.navigatorMode) children = await CollectionService.GetCollectionChildren(project.uri, "root", project.working_directory, project.working_directory, project.ignore_list, false);
		else {
			const navigatedCollectionId = collectionStore.navigatedCollection?.id;
			const collection_file_path = collectionStore.navigatedCollection?.file_path;
			children = await CollectionService.GetCollectionChildren(project.uri, navigatedCollectionId, project.working_directory, collection_file_path, project.ignore_list, false);
		}
	}
	if (children.assets) await assetStore.processAssetsIconsAndPreviews(children.assets);
	if (children.untracked_assets) await assetStore.processUntrackedAssetsIcons(children.untracked_assets);
	const allCollections = commonStore.showCollections ? children.collections?.filter((item) => !item.is_trashed) : [];
	const allAssets = commonStore.showAssets ? children.assets : [];
	rootData.value = sortItems(allCollections, allAssets, children.untracked_collections, children.untracked_assets);
	assetStore.assetsLoaded = true;
	collectionStore.loadCollectionStateFlags();
	await nextTick();
	dndStore.triggerDomUpdate();
};

// Toggles the details pane visibility.
const toggleDetailsPane = () => { panes.showDetailsPane = !panes.showDetailsPane; };

// Toggles between showing file name or full path.
const toggleShowFilters = () => { showFilters.value = !showFilters.value; };

// Callback for ResizeObserver to track container width changes.
const trackWidthChange = (entries) => { /* Reserved for future responsive layout calculations */ };

// Updates the search query, resets scroll position, and refreshes results.
const updateSearch = async (event) => {
	if (scrollStore.scrollTop > 0) scrollStore.requestScroll(0);
	if (collectionExpanded.value) collapseAll();
	commonStore.viewSearchQuery = event.target.value.toLowerCase();
	await softRefresh();
};

// Updates the screen width and hides details pane on smaller screens.
const updateScreenWidth = () => { screenWidth.value = window.innerWidth; if (screenWidth.value < 1000) panes.showDetailsPane = false; };

const debouncedUpdateSearch = useDebounce(updateSearch, 300);

// Copies files dropped from the OS file explorer into the target directory.
const handleFileDrop = async (files, details) => {
	if (!files || files.length === 0) return;

	const elementId = details?.id || '';
	if (elementId === 'agent-console-drop-zone') return;
	let destinationDir;
	if (elementId.startsWith('drop-')) {
		const collectionId = elementId.replace('drop-', '');
		const collection = dndStore.allViewItems?.find(item => item.id === collectionId && (item.type === 'collection' || item.type === 'untracked_collection'));
		if (collection?.file_path) destinationDir = collection.file_path;
	}
	if (!destinationDir) destinationDir = getCurrentDirectory();
	if (!destinationDir) { notificationStore.errorNotification(t('stages.couldNotDetermineCurrentDirectory'), ''); return; }

	stage.operationActive = true;
	try {
		await FSService.MakeDirs(destinationDir);
		let successCount = 0, failureCount = 0;
		const errors = [];
		for (const sourcePath of files) {
			try {
				const isFile = await FSService.IsFile(sourcePath);
				const itemName = await FSService.BaseName(sourcePath);
				const destinationPath = await generateUniqueDestinationPath(destinationDir, itemName);
				if (isFile) await FSService.DuplicateFile(sourcePath, destinationPath);
				else await FSService.DuplicateFolder(sourcePath, destinationPath);
				successCount++;
			} catch (error) {
				failureCount++;
				const itemName = await FSService.BaseName(sourcePath).catch(() => sourcePath);
				errors.push(`${itemName}: ${error.message || error}`);
			}
		}
		if (successCount > 0) notificationStore.addNotification(successCount === 1 ? t('stages.itemImportedSuccessfully') : t('stages.itemsImportedSuccessfully', { count: successCount }), '', 'success');
		if (failureCount > 0) notificationStore.errorNotification(failureCount === 1 ? t('stages.itemFailedToImport') : t('stages.itemsFailedToImport', { count: failureCount }), errors.join('\n'));
		if (successCount > 0) await softRefresh();
	} catch (error) { notificationStore.errorNotification(t('stages.errorImportingItems'), error.message || error); }
	finally { stage.operationActive = false; }
};

// events

Events.On('files-dropped', async (event) => {
	if (operationsActive.value) return;
	handleFileDrop(event.data?.files, event.data?.details);
});

Events.On('reload-view', async () => {
	if (operationsActive.value) return
	refresh();
});

Events.On('search', async () => {
	if (operationsActive.value) return
	if (searchBar.value) {
		searchBar.value.focus();
	}
});

Events.On('new-collection', async () => {
	if (operationsActive.value) return
	createCollection();
});

Events.On('new-asset', async () => {
	if (operationsActive.value) return
	createAsset();
});

Events.On('new-web-link', async () => {
	if (operationsActive.value) return
	createWebLink();
});

Events.On('sync-project', async () => {
	return 
	if (operationsActive.value) return
	//TODO prevent multiple syncs
	stage.operationActive = true;
	await syncData()
		.then(() => {
			stage.operationActive = false;
		})
		.catch((err) => {
			console.log(err)
			stage.operationActive = false;
		});
});

Events.On('group-items', async () => {
	if (operationsActive.value) return
	if (stage.markedItems.length > 1 && userStore.canDo('update_collection')) {
		stage.groupItems = true;
		createCollection();
	}
});

Events.On('cut-items', async () => {
	if (operationsActive.value) return;
	if (isEditableElementFocused()) return;
	if (!!stage.markedItems.length && userStore.canDo('update_collection')) {
		stage.copiedItems = [];
		const viewItems = dndStore.allViewItems;
		stage.cutItems = viewItems.filter((item) => stage.markedItems.includes(item.id));
		stage.cutItems = stage.cutItems.filter((item) => !stage.markedItems.includes(item.parent_id || item.collection_id));
		clearSelection();
	}
});

Events.On('copy-items', async () => {
	if (operationsActive.value) return;
	if (isEditableElementFocused()) return;
	if (!!stage.markedItems.length && userStore.canDo('update_collection')) {
		stage.cutItems = [];
		const viewItems = dndStore.allViewItems;
		let copiedItems = viewItems.filter((item) => stage.markedItems.includes(item.id));
		copiedItems = copiedItems.filter((item) => !stage.markedItems.includes(item.parent_id || item.collection_id));
		// Filter out tracked collections - copying collections is not yet supported
		stage.copiedItems = copiedItems.filter((item) => item.type !== 'collection');
		if (copiedItems.length > stage.copiedItems.length) {
			notificationStore.addNotification(t('stages.collectionsCannotBeCopied'), t('stages.onlyAssetsAddedToClipboard'), 'info');
		}
		clearSelection();
	}
});

Events.On('paste-items', async () => {
	if (operationsActive.value) return;
	if (isEditableElementFocused()) return;
	if (!userStore.canDo('update_collection')) return;
	const result = await stage.pasteItems();
	if (result.needsRefresh) await refresh();
});

Events.On('free-item-space', async () => {
	if (operationsActive.value) return
	if (stage.markedItems.length > 1) {
		prepFreeUpSpacePopUpModal();
	}
});

// Events.On('window-focused', async () => {
// 	if (operationsActive.value) return
// 	// await refresh();
// });

Events.On('delete-item', async () => {
	if (operationsActive.value) return
	if (stage.markedItems.length > 1 && userStore.canDo('delete_asset') && userStore.canDo('delete_collection')) {

		const allItemsToDelete = dndStore.allViewItems.filter((item) => stage.markedItems.includes(item.id));
		const hasUntrackedItems = allItemsToDelete.some((item) => item.type.includes('untracked'));
		const hasModifiedItems = allItemsToDelete.some((item) => item.file_status === 'modified');
		console.log(hasUntrackedItems)
		if (hasUntrackedItems || hasModifiedItems) {
			prepDeleteMultipleItemsPopUpModal();
		} else {
			deleteMultipleItems();
		}
	}
});

Events.On('duplicate-asset', async () => {
	if (operationsActive.value) return
	if (stage.markedItems.length !== 1) return
	if (!userStore.canDo('create_asset')) return
	await duplicateAsset();
});

Events.On('toggle-agent-console', async () => {
	const isOpen = !!modals.modalStates?.consoleModal;
	modals.setModalVisibility('consoleModal', !isOpen);
});

// watchers

watch(() => assetStore.assetsLoaded, async () => {
	if (assetStore.assetsLoaded) {
		const scrollTop = scrollStore.scrollTop;
		await nextTick();
		scrollStore.requestScroll(scrollTop);
	}
});

watch(() => projectStore.activeProject, async () => {
	if (projectStore.activeProject) {
		await refresh();
		stage.copiedItems = [];
		stage.cutItems = [];
	}
});

watch(() => collectionStore.navigatedCollection, async () => {
	await softRefresh();
});

watch(() => commonStore.showAssets, async () => {
	stage.operationActive = true
	await softRefresh();
	stage.operationActive = false
});

watch(() => commonStore.navigatorMode, async () => {
	stage.operationActive = true
	await softRefresh();
	stage.operationActive = false
});

// lifecycle hooks

onMounted(async () => {
	commonStore.resetFilters();
	dndStore.lockUI = true;
	commonStore.activeWorkspace = 'Default';
	commonStore.snapshotWorkspace();
	panes.showDetailsPane = screenWidth.value >= 1000;
	window.addEventListener('resize', updateScreenWidth);
	observer.value = new ResizeObserver(trackWidthChange);
	observer.value.observe(browserRoot.value);
	document.addEventListener('click', handleClickOutside);
	window.addEventListener('keydown', detectModifier);
	window.addEventListener('keyup', detectModifier);
	emitter.on('refresh-browser', softRefresh);
	emitter.on('update-root-data', handleUpdateRootData);
	emitter.on('update-untracked-items', handleUpdateUntrackedItems);
	emitter.on('expand-all', expandAll);
	emitter.on('collapse-all', collapseAll);

	if (projectStore.activeProject.is_remote) {
		await studioStore.getStudioUsers();
	}

	await refresh();
	dndStore.triggerDomUpdate();

	trayStates.trashables = await TrashService.GetTrashs(projectStore.activeProject.uri);
});

onUnmounted(() => {
	assetStore.assetsLoaded = false;
	emitter.off('refresh-browser', softRefresh);
	emitter.off('update-root-data', handleUpdateRootData);
	emitter.off('update-untracked-items', handleUpdateUntrackedItems);
	emitter.off('expand-all', expandAll);
	emitter.off('collapse-all', collapseAll);
	disableMenus();
});

onBeforeUnmount(() => {
	stage.expandedCollections = {};
	stage.markedCollections = [];
	panes.showDetailsPane = false;
	document.removeEventListener('click', handleClickOutside);
	window.removeEventListener('keydown', detectModifier);
	window.removeEventListener('keyup', detectModifier);
	window.removeEventListener('resize', updateScreenWidth);
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.dash-board-root {
	/* padding: .4rem; */
	display: flex;
	gap: .4rem;
	flex-direction: column;
	box-sizing: border-box;
	height: 100%;
	width: 100%;
	height: max-content;
}

.browser-root-container {
	position: relative;
	overflow: hidden;
	height: 100%;
	width: 100%;
	box-sizing: border-box;
	display: flex;
	flex-direction: row;
}

.kanban-container {
	z-index: 1;
	position: relative;
	flex-direction: column;
	padding: .5rem;
	overflow: hidden;
	height: 100%;
	border-radius: var(--very-large-radius);
	background-color: var(--surface-1);
	width: 100%;
	box-sizing: border-box;
	display: flex;
	flex-direction: row;
}

.browser-root-content {
	position: relative;
	flex-direction: column;
	overflow: hidden;
	height: 100%;
	width: 100%;
	box-sizing: border-box;
	display: flex;
	gap: .5rem;
	flex-direction: row;
}

.left-column {
	display: flex;
	position: relative;
	padding: .5rem;
	overflow: hidden;
	height: 100%;
	border-radius: var(--very-large-radius);
	background-color: var(--surface-1);
	width: 100%;
	min-width: 550px;
	box-sizing: border-box;
}

.browser-root-container-hover-drop {
	background-color: #1e7fee6c;
	outline: 1px solid rgb(255, 255, 255);
	outline-offset: -1px;
}

.dash-board-header {
	position: relative;
	display: flex;
	width: 100%;
	align-items: center;
	height: max-content;
	gap: 1rem;
	justify-content: space-between;
	box-sizing: border-box;
	min-width: max-content;
	min-height: 30px;
}

.view-options {
	display: flex;
	gap: .4rem;
	align-items: center;
	width: max-content;
	height: max-content;
	min-width: min-content;
}

.dash-board-filter {
	position: relative;
	display: flex;
	width: 100%;
	align-items: center;
	height: max-content;
	gap: .5rem;
	justify-content: space-between;
	padding: .5rem 0;
	box-sizing: border-box;
	overflow: hidden;
	min-height: 50px;
}
</style>