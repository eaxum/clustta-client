<template>
	<div ref="browserRoot" v-esc="cancelOps" v-right-click="openMenu" class="dash-board-root absolute-pane">
		
		<div v-if="isDefaultWorkspace" ref="browserFilters" class="dash-board-filter">
			<Breadcrumbs />
			<div class="searchbar-container">
				<input ref="searchBar" v-model="commonStore.viewSearchQuery" class="desktop-search-bar" type="text"
					:placeholder="'Search'" @input="debouncedUpdateSearch" spellcheck="false" />
				
				<ActionButton v-if="commonStore.viewSearchQuery.length && !assetStore.assetsLoaded" :isLoading="true" :icon="getAppIcon('loading')"  
					v-tooltip="'Loading...'" />

				<ActionButton v-else-if="commonStore.viewSearchQuery.length" :icon="getAppIcon('close')"
					:allowDeactivate="true" v-tooltip="'Clear search'" :buttonFunction="clearSearch" />
			</div>
			<ActionButton :icon="getAppIcon('filter')" :buttonFunction="toggleShowFilters" :isActive="showFilters" :showIndicator="filtersActive" v-tooltip="'Filters'" />
		</div>

		<div class="dash-board-header">

			<FilterBar v-if="showFilters && isDefaultWorkspace" 
			:filtersActive="filtersActive" :kanbanView="kanbanView" 
			:showTagsFilter="showTagsFilter" :viewTags="viewTags" 
			:isFilterActive="isFilterActive" :clearFilters="clearFilters" />

			<div v-else-if="isDefaultWorkspace" class="create-menu">
				<ActionButton :icon="getAppIcon('refresh')" v-tooltip="'Refresh'" :buttonFunction="refresh" />
				<ActionButton :icon="getAppIcon('file-plus')"
					v-if=" !kanbanView && (userStore.canDo('create_task') || canModifyEntity)"
					@click="createAsset" v-tooltip="'Add Asset'" />
				<ActionButton :icon="getAppIcon('folder-plus')"
					v-if=" !kanbanView && userStore.canDo('create_entity') || canModifyEntity" @click="createEntity"
					v-tooltip="'Add Collection'" />
				<ActionButton :icon="getAppIcon('workflow-plus')"
					v-if=" !kanbanView && workflowStore.workflows.length && userStore.canDo('create_entity')" @click="createWorkflow"
					v-tooltip="'Add Workflow'" />
				<ActionButton :icon="getAppIcon('web-plus')"
					v-if=" !kanbanView && userStore.canDo('create_task')" @click="createWebLink"
					v-tooltip="'Add Weblink'" />
				<ActionButton :icon="getAppIcon('arrow-down-ramp')"
					v-if="!platformStore.isWeb && !kanbanView && userStore.canDo('create_entity')" @click="importItems"
					v-tooltip="'Import Items'" />
				<!-- <ActionButton :icon="getAppIcon('arrow-up-ramp')"
					v-if="platformStore.isWeb && !kanbanView && userStore.canDo('create_entity')" @click="uploadItems"
					v-tooltip="'Upload Items'" /> -->
			</div>

			<div v-if="!showFilters || !isDefaultWorkspace" class="action-bar-container">
				
				<div v-if="!kanbanView && (collectionStore.loadingCollectionStates || assetStore.loadingAssetStates) && rootData.length" class="action-bar">
					<ActionButton :isLoading="true" :icon="getAppIcon('loading')"  
					v-tooltip="'Loading collection states'" />
				</div>

				<div v-else-if="!kanbanView && rootData.length" class="action-bar">
					<ActionButton v-if="collectionStore.collectionStateFlags.has_rebuildable" :icon="getAppIcon('jigsaw')" v-tooltip="'Rebuild All'"
						:buttonFunction="rebuildAll" />
					<ActionButton v-if="collectionStore.collectionStateFlags.has_untracked && userStore.canDo('create_checkpoint')"
						:icon="getAppIcon('layers-plus')" :useDanger="true" :noFilter="true" v-tooltip="'Create Checkpoints'"
						:buttonFunction="prepAllCheckpointModal" />
					<ActionButton v-else-if="collectionStore.collectionStateFlags.has_modified && userStore.canDo('create_checkpoint')"
						:icon="getAppIcon('layers-plus')" :useAlert="true"  :noFilter="true" v-tooltip="'Create Checkpoints'"
						:buttonFunction="prepAllCheckpointModal" />
					<ActionButton v-if="collectionStore.collectionStateFlags.has_modified" :icon="getAppIcon('revert')" :useAlert="true" :noFilter="true" v-tooltip="'Revert All'"
						:buttonFunction="prepResetPopUpModal" />
					<ActionButton v-if="collectionStore.collectionStateFlags.has_outdated" :icon="getAppIcon('circle-check')" :useAlert="true" :noFilter="true" v-tooltip="'Update all'"
						:buttonFunction="updateAll" />
				</div>
			</div>


			<div v-if="rootData.length || commonStore.viewSearchQuery.length || commonStore.showUntracked"
				class="view-options">
				<ViewOptions v-if="!kanbanView" :kanbanView="kanbanView"  />
				<ActionButton v-if="isDefaultWorkspace" :icon="kanbanView ? getAppIcon('list') : getAppIcon('kanban')"
					v-tooltip="kanbanView ? 'List' : 'Kanban'" :buttonFunction="toggleViewMode" />
				<ActionButton v-if="isDefaultWorkspace && !kanbanView && userStore.canDo('update_entity')" :icon="dndStore.lockUI ? getAppIcon('lock-closed') : getAppIcon('lock-open')"
					v-tooltip="dndStore.lockUI ? 'Unlock UI' : 'Lock UI'" :buttonFunction="toggleLockUI" />
				<ActionButton v-if="!kanbanView"
					:icon="commonStore.hideExtensions ? getAppIcon('extension-cancel') : getAppIcon('extension')"
					v-tooltip="commonStore.hideExtensions ? 'Show extensions' : 'Hide extensions'"
					:buttonFunction="toggleHideExtensions" />
				<ActionButton v-if="!kanbanView"
					:icon="commonStore.showFullPath ? getAppIcon('file-name') : getAppIcon('file-path')"
					v-tooltip="commonStore.showFullPath ? 'Name' : 'Path'"
					:buttonFunction="toggleShowFullPath" />
				<ActionButton v-if="!kanbanView && commonStore.showEntities && entityExpanded"
					:icon="entityExpanded ? getAppIcon('collapse-up') : getAppIcon('collapse-down')"
					v-tooltip="entityExpanded ? 'Collapse all' : 'Expand all'" :buttonFunction="toggleExpandEntities" />
				<ActionButton v-if="!kanbanView && isWideScreen" :icon="panes.showDetailsPane ? getAppIcon('collapse-right') : getAppIcon('collapse-left')"
					v-tooltip="panes.showDetailsPane ? 'Close pane' : 'Open pane'"
					:buttonFunction="toggleDetailsPane" />
			</div>

			<div v-else class="view-options">
				<ActionButton v-if="isWideScreen" :icon="panes.showDetailsPane ? getAppIcon('collapse-right') : getAppIcon('collapse-left')"
					v-tooltip="panes.showDetailsPane ? 'Close pane' : 'Open pane'"
					:buttonFunction="toggleDetailsPane" />
			</div>
			
		</div>

		<div v-if="!kanbanView" ref="taskListContainer" class="browser-root-container" @mousemove="onDrag($event)"
			:class="{ 'browser-root-container-hover-drop': isHovered }" @mouseup="onDragStop($event)"
			@scroll="disableMenus">

			<div v-if="draggedCard" id="ghost-card" ref="ghostCard" :style="ghostCardStyles"
				:class="{ 'active': dndStore.draggedItemId !== null, 'single-ghost': !commonStore.useGrid && stage.markedItems.length === 1, 'leaving': dndStore.ghostCardStyle.leaving }">
				<GhostItem v-bind="draggedCard" :data="draggedCard" />
			</div>

			<div class="browser-root-content">
				<div class="left-column">
					<VirtuaScroll v-if="(!assetStore.assetsLoaded || rootData.length) && !commonStore.useGrid" :items="rootData" />
					<GridView v-else-if="!assetStore.assetsLoaded || rootData.length" :rootItems="rootData" />
					<PageState v-else :message="message()" :prompt="prompt()" :illustration="illustration()" />
				</div>
				<DetailsPane v-if="projectStore.getProjects.length && isWideScreen" :isVisible="panes.showDetailsPane" />
			</div>


		</div>

		<div v-else ref="taskListContainer" class="browser-root-container kanban-container">
			<Kanban :filtersActive="filtersActive" :tasks="rootData" />
		</div>
	</div>
</template>

<script setup>
// imports
import { computed, nextTick, onBeforeUnmount, onMounted, onUnmounted, ref, watch } from 'vue';
import { Events } from '@wailsio/runtime';
import emitter from '@/lib/mitt';
import { getRelativePath } from '@/lib/pathlib';
import { useDebounce } from '@/lib/debounce';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import Breadcrumbs from '@/instances/common/components/Breadcrumbs.vue';
import DetailsPane from '@/instances/desktop/components/DetailsPane.vue';
import FilterBar from '@/instances/common/components/FilterBar.vue';
import GhostItem from '@/instances/desktop/blocks/GhostItem.vue';
import GridView from '@/instances/desktop/components/GridView.vue';
import Kanban from '@/instances/desktop/components/Kanban.vue';
import PageState from '@/instances/common/components/PageState.vue';
import ViewOptions from '@/instances/common/components/ViewOptions.vue';
import VirtuaScroll from '@/instances/common/components/VirtuaScroll.vue';

// services
import { AssetService, CheckpointService, CollectionService, DialogService, FSService, SyncService, TrashService } from '@/services';

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
import { useTagStore } from '@/stores/tags';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';
import { useWorkflowStore } from '@/stores/workflow';

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
const tagStore = useTagStore();
const trayStates = useTrayStates();
const userStore = useUserStore();
const workflowStore = useWorkflowStore();

// refs
const browserFilters = ref(null);
const browserRoot = ref(null);
const kanbanView = ref(false);
const observer = ref(null);
const rootData = ref([]);
const screenWidth = ref(window.innerWidth);
const searchBar = ref(null);
const showFilters = ref(false);
const taskListContainer = ref(null);

// computed properties
const canModifyEntity = computed(() => {
	if (!collectionStore.selectedCollection) return false
	let selectedIsMarked = stage.markedItems.includes(collectionStore.selectedCollection.id)
	if (selectedIsMarked && stage.markedItems.length === 1) return collectionStore.selectedCollection.can_modify
	return false
});

const draggedCard = computed(() => dndStore.allViewItems?.find(card => card.id === dndStore.draggedItemId));

const entityExpanded = computed(() => Object.keys(stage.expandedEntities).length);

const filtersActive = computed(() => {
	const assigneeFilters = commonStore.hasAssignees || commonStore.noAssignees;
	const entityFilters = commonStore.entityFilters.length > 0;
	const taskFilters = commonStore.taskFilters.length > 0;
	const resourceFilters = commonStore.resourceFilters.length > 0;
	const generalFilter = isFilterActive('general');
	return assigneeFilters || entityFilters || taskFilters || resourceFilters || generalFilter
});

const ghostCardStyles = computed(() => ({
	width: `${dndStore.ghostCardStyle.width}px`,
	left: `${dndStore.ghostCardStyle.pos.x}px`,
	top: `${dndStore.ghostCardStyle.pos.y}px`,
	transform: dndStore.ghostCardStyle.transform,
}));

const isDefaultWorkspace = computed(() => commonStore.activeWorkspace === 'Default');

const isHovered = computed(() => dndStore.isDropHovering && dndStore.targetItemId === null);

const isWideScreen = computed(() => screenWidth.value >= 1000);

const operationsActive = computed(() => {
	return stage.operationActive || !!modals.activeModal || !!menu.activeMenu || !assetStore.assetsLoaded || stage.activeStage !== 'browser'
});

const showTagsFilter = computed(() => !!tagStore.tags.length && (commonStore.showTasks || commonStore.showResources));

const viewTags = computed(() => {
	let tags = tagStore.tags;
	let viewTags = [];
	let filteredTaskResults = assetStore.getFilteredAssets;
	for (const task of filteredTaskResults) {
		let taskTags = task.tags;
		for (let t = 0; t < taskTags.length; t++) {
			if (!viewTags.includes(taskTags[t])) viewTags.push(taskTags[t])
		}
	}
	for (let i = 0; i < tags.length; i++) {
		tags[i].name = tags[i].name;
		tags[i].type = 'tags'
	}
	return tags.filter((item) => viewTags.includes(item.name));
});

// methods

// Adds an entity dependency to a task.
const addEntityDependency = async (taskId, dependencyId, dependencyTypeId) => {
	await AssetService.AddEntityDependency(projectStore.activeProject.uri, taskId, dependencyId, dependencyTypeId)
		.then(() => notificationStore.addNotification('Dependency Added.', "", "success"))
		.catch((error) => { console.log(error); notificationStore.errorNotification("Error adding dependencies", error); });
};

// Adds a task dependency between two assets.
const addDependency = async (taskId, dependencyId, dependencyTypeId) => {
	await AssetService.AddAssetDependency(projectStore.activeProject.uri, taskId, dependencyId, dependencyTypeId)
		.then(() => notificationStore.addNotification('Dependency Added.', "", "success"))
		.catch((error) => { console.log(error); notificationStore.errorNotification("Error adding dependencies", error); });
};

// Cancels active operations like search or drag-and-drop.
const cancelOps = () => {
	if (commonStore.viewSearchQuery) clearSearch();
	stage.cutItems = [];
	if (!dndStore.altKeyActive) dndStore.resetValues();
};

// Changes the parent collection of an entity.
const changeEntityParent = async (entityId, parentId) => {
	await CollectionService.ChangeCollectionParent(projectStore.activeProject.uri, entityId, parentId)
		.then(() => notificationStore.addNotification('Moved successfully.', "", "success"))
		.catch((error) => { console.error(error); notificationStore.errorNotification("Error changing entity parent", error); });
};

// Moves a task to a different collection.
const changeTaskEntity = async (taskId, entityId) => {
	await AssetService.ChangeAssetCollection(projectStore.activeProject.uri, taskId, entityId)
		.then(() => notificationStore.addNotification('Moved successfully.', "", "success"))
		.catch((error) => notificationStore.errorNotification("Error changing task entity", error));
};

// Clears all filters and refreshes the view.
const clearFilters = () => { commonStore.resetFilters(); softRefresh(); };

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

// Collapses all expanded entities and clears selection.
const collapseAll = () => {
	stage.expandedEntities = {};
	stage.markedItems = [];
	stage.firstSelectedEntityId = '';
	collectionStore.selectedCollection = null;
};

// Opens the application selection modal to create a new asset.
const createAsset = () => { clearSelection(); modals.setModalVisibility('selectAppModal', true); };

// Opens the create collection modal.
const createEntity = () => { if (!stage.groupItems) clearSelection(); modals.setModalVisibility('createCollectionModal', true); };

// Opens the add web link modal.
const createWebLink = () => { clearSelection(); modals.setModalVisibility('addWebLinkModal', true); };

// Opens the workflow selection modal.
const createWorkflow = () => { clearSelection(); modals.setModalVisibility('selectWorkflowModal', true); };

// Deletes all selected items including tasks, entities, and untracked files.
const deleteMultipleItems = async () => {
	panes.setPaneVisibility('projectDetails', true);
	stage.operationActive = true;
	stage.firstSelectedItemId = '';
	stage.lastSelectedItemId = '';
	assetStore.selectedAsset = null;
	collectionStore.selectedCollection = null;
	const allItemsToDelete = dndStore.allViewItems.filter((item) => stage.markedItems.includes(item.id));
	const tasksToDelete = allItemsToDelete.filter((item) => item.type === 'task');
	const entitiesToDelete = allItemsToDelete.filter((item) => item.type === 'entity');
	const untrackedTasksToDelete = allItemsToDelete.filter((item) => item.type === 'untracked_task');
	const untrackedEntitiesToDelete = allItemsToDelete.filter((item) => item.type === 'untracked_entity');
	await deleteMultipleTasks(tasksToDelete.map((item) => item.id));
	await deleteMultipleEntities(entitiesToDelete.map((item) => item.id));
	await deleteMultipleUntrackedTasks(untrackedTasksToDelete);
	await deleteMultipleUntrackedEntities(untrackedEntitiesToDelete);
	stage.markedItems = [];
	stage.selectedItems = [];
	stage.markedTasks = [];
	stage.markedEntities = [];
	stage.operationActive = false;
	modals.setModalVisibility('popUpModal', false);
};

// Moves multiple entities to trash.
const deleteMultipleEntities = async (entityIds) => {
	for (let entityId of entityIds) {
		await CollectionService.DeleteCollection(projectStore.activeProject.uri, entityId, true)
			.then(async () => { await collectionStore.markCollectionAsDeleted(entityId); notificationStore.addNotification("Entity moved to Trash.", '', "success", false); })
			.catch((error) => { console.log(error); notificationStore.errorNotification("Entities failed to delete.", error); });
	}
};

// Moves multiple tasks to trash.
const deleteMultipleTasks = async (taskIds) => {
	for (let taskId of taskIds) {
		await AssetService.DeleteAsset(projectStore.activeProject.uri, taskId, true)
			.then(async () => { softRefresh(); notificationStore.addNotification("Tasks moved to Trash.", '', "success", false); })
			.catch((error) => { console.log(error); notificationStore.errorNotification("Tasks failed to delete.", error); });
	}
};

// Permanently deletes multiple untracked entity folders.
const deleteMultipleUntrackedEntities = async (untrackedEntities) => {
	for (let untrackedEntity of untrackedEntities) {
		FSService.DeleteFolder(untrackedEntity.file_path);
		projectStore.removeUntrackedEntity(untrackedEntity.id);
	}
};

// Permanently deletes multiple untracked task files.
const deleteMultipleUntrackedTasks = async (untrackedTasks) => {
	for (let untrackedTask of untrackedTasks) {
		await FSService.DeleteFile(untrackedTask.file_path);
		projectStore.removeUntrackedTask(untrackedTask.id);
	}
};

// Detects Alt key state for drag-and-drop modifier behavior.
const detectModifier = (event) => { dndStore.altKeyActive = event.getModifierState('Alt'); };

// Hides all context menus.
const disableMenus = () => { menu.disableAllMenus(); };

// Duplicates the selected task in the database and copies the physical file.
const duplicateTask = async () => {
	const selectedItemId = stage.markedItems[0];
	const selectedItem = dndStore.allViewItems.find(item => item.id === selectedItemId);
	if (!selectedItem || selectedItem.type !== 'task') return;
	try {
		stage.operationActive = true;
		stage.markedItems = [];
		assetStore.selectedAsset = null;
		await AssetService.DuplicateAsset(projectStore.activeProject.uri, selectedItemId)
			.then(async (duplicatedTask) => {
				try { await FSService.DuplicateFile(selectedItem.file_path, duplicatedTask.file_path); }
				catch (fileError) { console.warn('Physical file duplication failed (asset may be rebuildable):', fileError); }
				await refresh();
				assetStore.selectAsset(duplicatedTask);
				stage.selectedItem = duplicatedTask;
				stage.markedItems = [duplicatedTask.id];
				stage.lastSelectedItemId = "";
				stage.firstSelectedItemId = duplicatedTask.id;
				notificationStore.addNotification(`Asset duplicated`, '', "success", false);
			});
	} catch (error) {
		console.error('Error duplicating task:', error);
		notificationStore.errorNotification("Failed to duplicate task", error.message || error);
	} finally { stage.operationActive = false; }
};

// Expands all entities in the view.
const expandAll = () => {
	const entities = collectionStore.getCollections;
	const expandedEntities = {};
	for (let i = 0; i < entities.length; i++) {
		expandedEntities[entities[i].id] = { "height": 0, "entity_path": entities[i].entity_path };
	}
	stage.expandedEntities = expandedEntities;
};

// Frees up disk space by removing files for selected tasks and entities.
const freeUpSpace = async () => {
	panes.setPaneVisibility('projectDetails', true);
	stage.operationActive = true;
	stage.firstSelectedItemId = '';
	stage.lastSelectedItemId = '';
	assetStore.selectedAsset = null;
	collectionStore.selectedCollection = null;
	const allItemsToDelete = dndStore.allViewItems.filter((item) => stage.markedItems.includes(item.id));
	await freeUpMultipleTaskSpace(allItemsToDelete.filter((item) => item.type === 'task'));
	await freeUpMultipleEntitySpace(allItemsToDelete.filter((item) => item.type === 'entity'));
	await deleteMultipleUntrackedTasks(allItemsToDelete.filter((item) => item.type === 'untracked_task'));
	await deleteMultipleUntrackedEntities(allItemsToDelete.filter((item) => item.type === 'untracked_entity'));
	stage.markedItems = [];
	stage.selectedItems = [];
	stage.markedTasks = [];
	stage.markedEntities = [];
	stage.operationActive = false;
	modals.setModalVisibility('popUpModal', false);
};

// Deletes entity folders to free up disk space.
const freeUpMultipleEntitySpace = async (entities) => {
	for (const entity of entities) {
		let entityDir = entity.file_path.replace(/\\/g, '/');
		await FSService.DeleteFolder(entityDir)
			.then(() => assetStore.refreshEntityFilesStatus(entity.id))
			.catch((error) => console.error(error));
	}
};

// Deletes physical files for tasks to free up disk space.
const freeUpMultipleTaskSpace = async (selectedTasks) => {
	const fileStatus = ['missing', 'rebuildable'];
	let taskIds = [];
	for (let task of selectedTasks) { if (!fileStatus.includes(task.file_status)) taskIds.push(task.id); }
	for (const taskId of taskIds) {
		let task = assetStore.getAssets.find((item) => item.id === taskId);
		let taskPath = task.file_path.replace(/\\/g, '/');
		await FSService.DeleteFile(taskPath)
			.then(() => { task.file_status = 'rebuildable'; })
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
			if (!event.target.closest('.entity-item-main')) {
				stage.markedItems = []; stage.selectedItems = []; stage.markedEntities = []; stage.firstSelectedItemId = ''; stage.lastSelectedItemId = '';
				stage.selectedItem = null; assetStore.selectedAsset = null; collectionStore.selectedCollection = null; stage.cutItems = []; projectStore.selectedUntrackedItem = null;
			}
			if (!event.target.closest('.task-item-main')) {
				stage.markedItems = []; stage.selectedItems = []; stage.markedEntities = []; stage.firstSelectedItemId = ''; stage.lastSelectedItemId = '';
				stage.selectedItem = null; assetStore.selectedAsset = null; stage.cutItems = []; projectStore.selectedUntrackedItem = null;
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

// Replaces untracked items in root data with updated list from emitter events.
const handleUpdateUntrackedItems = (untrackedItems) => {
	if (!untrackedItems) return;
	rootData.value = rootData.value.filter(item => item.type !== 'untracked_entity' && item.type !== 'untracked_task');
	rootData.value.push(...untrackedItems);
	emitter.emit('get-project-data');
	collectionStore.loadCollectionStateFlags();
};

// Returns the empty state illustration path.
const illustration = () => commonStore.viewSearchQuery ? '/page-states/resources.png' : '/page-states/tasks.png';

// Imports files or folders from the file system into the current directory.
const importItems = async () => {
	try {
		let selectedPaths;
		try { selectedPaths = await DialogService.SelectFilesDialog(); } catch (error) { return; }
		if (!selectedPaths || selectedPaths.length === 0) return;
		const currentDirectory = getCurrentDirectory();
		if (!currentDirectory) { notificationStore.errorNotification("Could not determine current directory", ""); return; }
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
		if (successCount > 0) notificationStore.addNotification(successCount === 1 ? "1 item imported successfully" : `${successCount} items imported successfully`, "", "success");
		if (failureCount > 0) notificationStore.errorNotification(failureCount === 1 ? "1 item failed to import" : `${failureCount} items failed to import`, errors.join("\n"));
		if (successCount > 0) await softRefresh();
	} catch (error) { notificationStore.errorNotification("Error importing items", error.message || error); }
	finally { stage.operationActive = false; }
};

// Checks if a specific filter type is currently active.
const isFilterActive = (filter) => {
	if (filter.includes('general')) {
		const isActive = commonStore.showEntities && commonStore.showTasks && commonStore.showResources && commonStore.showChildEntities && commonStore.showChildTasks && commonStore.showDependencies && !commonStore.onlyAssets;
		return !isActive;
	} else if (filter.includes('entity')) return commonStore.entityFilters.some((item) => item.type === filter);
	else if (filter.includes('assignation')) {
		const assigneeFilters = commonStore.hasAssignees || commonStore.noAssignees;
		return assigneeFilters || commonStore.taskFilters.some((item) => item.type === filter);
	} else return commonStore.taskFilters.some((item) => item.type === filter);
};

// Returns the empty state message based on current view context.
const message = () => {
	const searching = commonStore.viewSearchQuery;
	const myTasksWorkspace = commonStore.activeWorkspace === 'My Tasks';
	if (searching) return 'No results found.';
	if (isDefaultWorkspace.value && filtersActive.value) return 'No results match your filters.';
	if (myTasksWorkspace) return 'You have no assets assigned to you.';
	if (!isDefaultWorkspace.value) return 'Nothing in this workspace.';
	return 'Nothing to see here.';
};

// Handles drag movement events.
const onDrag = (e) => dndStore.onDrag(e);

// Handles drag stop events and processes item moves, parent changes, or dependency additions.
const onDragStop = async (event) => {
	let doRefresh = true;
	if (kanbanView.value) return;
	if (dndStore.draggedItemId === null) return;
	document.documentElement.style.cursor = 'default';
	const dropTarget = dndStore.itemRefs[dndStore.targetItemId];
	const draggedItem = dndStore.itemRefs[dndStore.draggedItemId];
	let cardRect;
	const targetEntity = dndStore.allViewItems.find((item) => item.id === dndStore.targetItemId);
	stage.operationActive = true;
	let draggedEntity;
	const draggedItemIds = stage.markedItems;
	for (const draggedItemId of draggedItemIds) {
		draggedEntity = dndStore.allViewItems.find((item) => item.id === draggedItemId);
		if (!draggedEntity) return;
		if (event.altKey) {
			cardRect = draggedItem.getBoundingClientRect();
			if (draggedEntity.entity_type_id) await changeEntityParent(draggedItemId, '');
			else if (draggedEntity.task_type_id) await changeTaskEntity(draggedItemId, '');
			else {
				let extension = draggedEntity.type === 'untracked_task' ? draggedEntity.extension : '';
				let fullName = draggedEntity.name + extension;
				await FSService.MakeDirs(projectStore.activeProject.working_directory);
				let newPath = await FSService.JoinPath(projectStore.activeProject.working_directory, fullName);
				await FSService.Rename(draggedEntity.file_path, newPath);
			}
		} else if (dndStore.isOverlapping && dropTarget) {
			cardRect = dropTarget.getBoundingClientRect();
			if (draggedItemId !== dndStore.targetItemId) {
				if (targetEntity.type === 'entity') {
					if (draggedEntity.type === 'entity') await changeEntityParent(draggedItemId, dndStore.targetItemId);
					else if (draggedEntity.type === 'task') await changeTaskEntity(draggedItemId, dndStore.targetItemId);
					else {
						let entity = await CollectionService.GetCollectionByID(projectStore.activeProject.uri, dndStore.targetItemId);
						await FSService.MakeDirs(entity.file_path);
						let extension = draggedEntity.type === 'untracked_task' ? draggedEntity.extension : '';
						let fullName = draggedEntity.name + extension;
						let newPath = await FSService.JoinPath(entity.file_path, fullName);
						await FSService.Rename(draggedEntity.file_path, newPath);
					}
				} else if (targetEntity.task_type_id) {
					let dependencyTypeId = dependencyStore.dependency_types.find(item => item.name === "linked").id;
					if (draggedEntity.entity_type_id) await addEntityDependency(dndStore.targetItemId, draggedItemId, dependencyTypeId);
					else if (draggedEntity.task_type_id) await addDependency(dndStore.targetItemId, draggedItemId, dependencyTypeId);
					else if (!draggedEntity.item_type) await addDependency(dndStore.targetItemId, draggedItemId, dependencyTypeId);
				} else if (targetEntity.type === 'untracked_entity') {
					if (!draggedEntity.entity_type_id && !draggedEntity.task_type_id && (draggedEntity.type === 'untracked_task' || draggedEntity.type === 'untracked_entity')) {
						let extension = draggedEntity.type === 'untracked_task' ? draggedEntity.extension : '';
						let fullName = draggedEntity.name + extension;
						let newPath = await FSService.JoinPath(targetEntity.file_path, fullName);
						await FSService.Rename(draggedEntity.file_path, newPath);
					}
				}
			} else cardRect = draggedItem.getBoundingClientRect();
		} else cardRect = draggedItem.getBoundingClientRect();
	}
	setTimeout(() => dndStore.resetValues(), 100);
	dndStore.ghostCardStyle.leaving = true;
	let xOffset = cardRect.x - dndStore.ghostCardStyle.pos.x;
	let yOffset = cardRect.y - dndStore.ghostCardStyle.pos.y;
	dndStore.ghostCardStyle.transform = `scale(1) translate(${xOffset}px, ${yOffset}px)`;
	if (doRefresh) softRefresh();
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

// Prepares and shows the create multiple checkpoints modal.
const prepAllCheckpointModal = () => {
	clearSelection();
	trayStates.createMultipleCheckpoints = true;
	trayStates.createMultipleCheckpointsEntityPath = "";
	modals.setModalVisibility('createMultipleCheckpointsModal', true);
};

// Prepares and shows the delete multiple items confirmation modal.
const prepDeleteMultipleItemsPopUpModal = () => {
	const numberOfItems = stage.markedItems.length;
	trayStates.popUpModalTitle = "Delete " + numberOfItems + " items";
	trayStates.popUpModalMessage = "You have selected some untracked/modified items and they will be permanently deleted. Continue?";
	trayStates.popUpModalIcon = 'trash';
	trayStates.popUpModalFunction = deleteMultipleItems;
	modals.setModalVisibility('popUpModal', true);
};

// Prepares and shows the free up space confirmation modal.
const prepFreeUpSpacePopUpModal = () => {
	trayStates.popUpModalTitle = "Free up Space";
	trayStates.popUpModalMessage = "Are you sure you want to delete these items? This will permanently remove all uncheckpointed resources and all task outputs. Please confirm if you wish to proceed.";
	trayStates.popUpModalIcon = 'broom';
	trayStates.popUpModalFunction = freeUpSpace;
	modals.setModalVisibility('popUpModal', true);
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

// Returns the empty state prompt text.
const prompt = () => {
	if (commonStore.viewSearchQuery) return '';
	if (!isDefaultWorkspace.value || filtersActive.value) return '';
	return 'Right click to create a new Collection or Asset.';
};

// Rebuilds all rebuildable assets in the current view.
const rebuildAll = async () => {
	const path = collectionStore.navigatedCollection?.entity_path;
	const navigatedEntityId = collectionStore.navigatedCollection?.id;
	notificationStore.cancleFunction = SyncService.CancelSync;
	notificationStore.canCancel = true;
	if (commonStore.activeWorkspace === 'My Tasks' && rootData.value.length) {
		const userTaskIds = rootData.value.map((task) => task.id);
		await CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, userTaskIds)
			.then(() => softRefresh())
			.catch((error) => console.error(`Error rebuilding tasks:`, error));
	} else {
		await CollectionService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, navigatedEntityId)
			.then(() => { if (!path) assetStore.rebuildableAssetsPath = []; softRefresh(); })
			.catch((error) => notificationStore.errorNotification("Error Rebuilding All", error));
	}
};

// Full refresh: reloads project data, fetches all children, processes icons/previews, and updates state flags.
const refresh = async () => {
	if (kanbanView.value) return;
	assetStore.assetsLoaded = false;
	stage.cutItems = [];
	await projectStore.refreshActiveProject();
	await trayStates.refreshData();
	let children;
	let project = projectStore.activeProject;
	if (!commonStore.navigatorMode) children = await CollectionService.GetCollectionChildren(project.uri, "root", project.working_directory, project.working_directory, project.ignore_list, false);
	else {
		const navigatedEntityId = collectionStore.navigatedCollection?.id;
		const entity_file_path = collectionStore.navigatedCollection?.file_path;
		children = await CollectionService.GetCollectionChildren(project.uri, navigatedEntityId, project.working_directory, entity_file_path, project.ignore_list, false);
	}
	await assetStore.processAssetsIconsAndPreviews(children.tasks);
	await assetStore.processUntrackedAssetsIcons(children.untracked_tasks);
	rootData.value = [...children.entities, ...children.untracked_entities, ...children.tasks, ...children.untracked_tasks];
	assetStore.assetsLoaded = true;
	collectionStore.loadCollectionStateFlags();
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
		.then(() => { assetStore.modifiedAssets.modified = assetStore.modifiedAssets.modified.filter((item) => !filteredPaths.includes(item.task_path)); softRefresh(); })
		.catch((error) => { notificationStore.errorNotification("Failed to Revert Tasks", error); console.error(error); });
};

// Lightweight refresh: fetches children with search/filter support, processes icons, updates root data and state flags.
const softRefresh = async () => {
	assetStore.assetsLoaded = false;
	stage.cutItems = [];
	let children = {};
	let project = projectStore.activeProject;
	const searching = commonStore.viewSearchQuery.toLowerCase();
	if (searching || filtersActive.value) {
		let entities, tasks;
		if (!commonStore.navigatorMode) {
			if (!searching) {
				const rootItems = await CollectionService.GetCollectionChildren(project.uri, "root", project.working_directory, project.working_directory, project.ignore_list, false);
				entities = rootItems['entities'];
				tasks = commonStore.onlyAssets ? await AssetService.GetAssets(project.uri) : rootItems['tasks'];
			} else {
				entities = await CollectionService.GetCollections(project.uri);
				tasks = await AssetService.GetAssets(project.uri);
			}
			entities = commonStore.onlyAssets ? [] : entities;
		} else {
			const navigatedEntityId = collectionStore.navigatedCollection?.id;
			const entity_file_path = collectionStore.navigatedCollection?.file_path;
			const entityItems = await CollectionService.GetCollectionChildren(project.uri, navigatedEntityId, project.working_directory, entity_file_path, project.ignore_list, false);
			entities = entityItems['entities'];
			tasks = entityItems['tasks'];
		}
		children['entities'] = await collectionStore.filterCollections(entities);
		children['tasks'] = await assetStore.filterAssets(tasks);
	} else {
		if (!commonStore.navigatorMode) children = await CollectionService.GetCollectionChildren(project.uri, "root", project.working_directory, project.working_directory, project.ignore_list, false);
		else {
			const navigatedEntityId = collectionStore.navigatedCollection?.id;
			const entity_file_path = collectionStore.navigatedCollection?.file_path;
			children = await CollectionService.GetCollectionChildren(project.uri, navigatedEntityId, project.working_directory, entity_file_path, project.ignore_list, false);
		}
	}
	if (children.tasks) await assetStore.processAssetsIconsAndPreviews(children.tasks);
	if (children.untracked_tasks) await assetStore.processUntrackedAssetsIcons(children.untracked_tasks);
	const allEntities = commonStore.showEntities ? children.entities?.filter((item) => !item.is_trashed) : [];
	const allTasks = commonStore.showTasks ? children.tasks : [];
	rootData.value = [...(allEntities ?? []), ...(allTasks ?? []), ...(children.untracked_entities ?? []), ...(children.untracked_tasks ?? [])];
	assetStore.assetsLoaded = true;
	collectionStore.loadCollectionStateFlags();
};

// Toggles the details pane visibility.
const toggleDetailsPane = () => { panes.showDetailsPane = !panes.showDetailsPane; };

// Toggles between expanding and collapsing all entities.
const toggleExpandEntities = () => { if (entityExpanded.value) collapseAll(); else expandAll(); };

// Toggles file extension visibility in the browser.
const toggleHideExtensions = () => { commonStore.hideExtensions = !commonStore.hideExtensions; };

// Toggles the UI lock state for drag-and-drop.
const toggleLockUI = () => { dndStore.lockUI = !dndStore.lockUI; };

// Toggles between showing file name or full path.
const toggleShowFilters = () => { showFilters.value = !showFilters.value; };

// Toggles between showing file name or full path.
const toggleShowFullPath = () => { commonStore.showFullPath = !commonStore.showFullPath; };

// Toggles between list and kanban view modes.
const toggleViewMode = () => { kanbanView.value = !kanbanView.value; softRefresh(); panes.showDetailsPane = !kanbanView.value; };

// Callback for ResizeObserver to track container width changes.
const trackWidthChange = (entries) => { /* Reserved for future responsive layout calculations */ };

// Updates all outdated tasks to their latest server version.
const updateAllOutdated = async () => {
	modals.setModalVisibility('popUpModal', false);
	notificationStore.cancleFunction = SyncService.CancelSync;
	notificationStore.canCancel = true;
	const navigated = collectionStore.navigatedCollection;
	let collectionId = null;
	if (navigated?.type === 'entity') collectionId = navigated.id;
	const outdatedTasks = await collectionStore.getOutdatedItems(collectionId);
	const filteredPaths = outdatedTasks.map(task => task.task_path);
	if (filteredPaths.length === 0) return;
	await CheckpointService.RevertTaskPaths(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, filteredPaths)
		.then(() => softRefresh())
		.catch((error) => { notificationStore.errorNotification("Failed to Revert Tasks", error); console.error(error); });
};

// Updates all outdated tasks after clearing selection.
const updateAll = () => { clearSelection(); updateAllOutdated(); };

// Updates the search query, resets scroll position, and refreshes results.
const updateSearch = async (event) => {
	if (scrollStore.scrollTop > 0) scrollStore.requestScroll(0);
	if (entityExpanded.value) collapseAll();
	commonStore.viewSearchQuery = event.target.value.toLowerCase();
	await softRefresh();
};

// Updates the screen width and hides details pane on smaller screens.
const updateScreenWidth = () => { screenWidth.value = window.innerWidth; if (screenWidth.value < 1000) panes.showDetailsPane = false; };

const debouncedUpdateSearch = useDebounce(updateSearch, 300);

// events

Events.On('clustta-drag-drop', async () => {
	console.log('clustta-drag-drop')
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
	createEntity();
});

Events.On('new-task', async () => {
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
	if (stage.markedItems.length > 1 && userStore.canDo('update_entity')) {
		stage.groupItems = true;
		createEntity();
	}
});

Events.On('cut-items', async () => {
	if (operationsActive.value) return
	if (!!stage.markedItems.length && userStore.canDo('update_entity')) {
		const viewItems = dndStore.allViewItems;
		stage.cutItems = viewItems.filter((item) => stage.markedItems.includes(item.id));
		stage.cutItems = stage.cutItems.filter((item) => !stage.markedItems.includes(item.parent_id || item.entity_id));
		stage.markedItems = stage.cutItems.map((item) => item.id)
		// console.log(stage.cutItems);
	}
});

Events.On('copy-items', async () => {
	if (operationsActive.value) return
	if (!!stage.markedItems.length && userStore.canDo('update_entity')) {
		const viewItems = dndStore.allViewItems;
		stage.copiedItems = viewItems.filter((item) => stage.markedItems.includes(item.id));
		stage.copiedItems = stage.copiedItems.filter((item) => !stage.markedItems.includes(item.parent_id || item.entity_id));
		stage.markedItems = stage.copiedItems.map((item) => item.id)
		// console.log(stage.copiedItems);
	}
});

Events.On('paste-items', async () => {
	if (operationsActive.value) return
	if (!!stage.copiedItems && userStore.canDo('update_entity')) {
		stage.copiedItems = [];
	}
});

Events.On('free-item-space', async () => {
	if (operationsActive.value) return
	if (stage.markedItems.length > 1) {
		prepFreeUpSpacePopUpModal();
	}
});

Events.On('window-focused', async () => {
	console.log('focused')
	if (operationsActive.value) return
	// await refresh();
});

Events.On('delete-item', async () => {
	if (operationsActive.value) return
	if (stage.markedItems.length > 1 && userStore.canDo('delete_task') && userStore.canDo('delete_entity')) {

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

Events.On('duplicate-task', async () => {
	if (operationsActive.value) return
	if (stage.markedItems.length !== 1) return
	if (!userStore.canDo('create_task')) return
	await duplicateTask();
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
		kanbanView.value = false;
		await refresh();
	}
});

watch(() => collectionStore.navigatedCollection, async () => {
	await softRefresh();
});

watch(() => commonStore.showTasks, async () => {
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

	if (!studioStore.studioUsers.length) {
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
	disableMenus();
});

onBeforeUnmount(() => {
	stage.expandedEntities = {};
	stage.markedEntities = [];
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
	padding: .4rem;
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
	z-index: 5;
	position: relative;
	flex-direction: column;
	padding: .5rem;
	overflow: hidden;
	height: 100%;
	border-radius: var(--large-radius);
	background-color: var(--black-steel);
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
	background-color: var(--black-steel);
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
}


.create-menu {
	position: relative;
	display: flex;
	align-items: center;
	gap: .4rem;
	width: max-content;
	height: max-content;
}

.action-bar {
	position: relative;
	display: flex;
	align-items: center;
	gap: .4rem;
	width: max-content;
	height: max-content;
}

.action-bar-container{
	position: relative;
	display: flex;
	align-items: center;
	gap: .4rem;
	width: max-content;
	height: max-content;
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

.desktop-search-bar {
	font-family: 'Inter', sans-serif;
	font-weight: 200;
	box-sizing: border-box;
	font-size: 16px;
	padding: 10px;
	border: 0px;
	border-style: solid;
	outline: none;
	background-color: var(--midnight-steel);
	color: var(--white);
	transition: width 0.2s ease-out;
	border-radius: var(--large-radius);
	width: 100%;
	max-width: 400px;
}

.desktop-search-bar::-ms-reveal {
	filter: invert(100%);
}

.searchbar-container {
	display: flex;
	align-items: center;
	border: 0px;
	border-style: solid;
	outline: none;
	background-color: var(--midnight-steel);
	width: 100%;
	max-width: 350px;
	padding-right: .2rem;
	box-sizing: border-box;
	border-radius: var(--large-radius);
}

.searchbar-container:hover {
	outline: var(--transparent-line);
	outline-offset: -1px;
}

#ghost-card {
	position: fixed;
	z-index: 100;
	user-select: none;
	pointer-events: none;
	opacity: 0;
	transform-origin: center;
	transform: scale(1) rotate(0);
	box-shadow: 0 1px 0 rgba(9, 30, 66, 0.25);
	transition: transform 0.04s ease-in-out;
	border-radius: 10px;
}

#ghost-card.animate {
	transition: box-shadow 0.1s ease-in-out;
	transition: transform 0.05s ease-in-out;
}

#ghost-card.active {
	opacity: 1;
}

#ghost-card.single-ghost {
	outline: 1.5px solid rgb(255, 255, 255);
}

#ghost-card.leaving {
	transition: all 0.1s ease;
	box-shadow: 0 1px 0 rgba(9, 30, 66, 0.25);
}
</style>