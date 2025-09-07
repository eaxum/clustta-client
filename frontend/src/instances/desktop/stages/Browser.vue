<template>
	<div ref="browserRoot" v-esc="cancelOps" v-right-click="openMenu" class="dash-board-root absolute-pane">
		
		<div v-if="isDefaultWorkspace" ref="browserFilters" class="dash-board-filter">
			<Breadcrumbs />
			<div class="searchbar-container">
				<input ref="searchBar" v-model="commonStore.viewSearchQuery" class="desktop-search-bar" type="text"
					:placeholder="'Search'" @input="debouncedUpdateSearch" />
				<ActionButton v-if="commonStore.viewSearchQuery.length" :icon="getAppIcon('close')"
					:allowDeactivate="true" v-tooltip="'Clear search'" :buttonFunction="clearSearch" />
			</div>
			<span class="single-action-button filter-button" @click="toggleShowFilters" v-tooltip="'Filters'">
				<div v-if="filtersActive" class="filter-button-indicator"></div>
				<img class="small-icons" :src="getAppIcon('filter')">
			</span>
		</div>

		<FilterBar v-if="showFilters && isDefaultWorkspace" 
		:filtersActive="filtersActive" :kanbanView="kanbanView" 
		:showTagsFilter="showTagsFilter" :viewTags="viewTags" 
		:isFilterActive="isFilterActive" :clearFilters="clearFilters" />


		<div class="menu-divider" v-if="showFilters && isDefaultWorkspace" ></div>

		<div class="dash-board-header">
				<div v-if="isDefaultWorkspace" class="create-menu">
					<ActionButton :icon="getAppIcon('refresh')" v-tooltip="'Refresh'" :buttonFunction="refresh" />
					<ActionButton :icon="getAppIcon('brush-plus')"
						v-if=" !kanbanView && templateStore.getTemplates.length && (userStore.canDo('create_task') || canModifyEntity)"
						@click="createTask" v-tooltip="'Add Asset'" />
					<ActionButton :icon="getAppIcon('folder-plus')"
						v-if=" !kanbanView && userStore.canDo('create_entity') || canModifyEntity" @click="createEntity"
						v-tooltip="'Add Collection'" />
					<ActionButton :icon="getAppIcon('workflow-plus')"
						v-if=" !kanbanView && workflowStore.workflows.length && userStore.canDo('create_entity')" @click="createWorkflow"
						v-tooltip="'Add Workflow'" />
					<ActionButton :icon="getAppIcon('arrow-down-ramp')"
						v-if=" !kanbanView && userStore.canDo('create_entity')" @click="importItems"
						v-tooltip="'Import Items'" />
			</div>

			<div class="action-bar-container">
				<div v-if="!kanbanView && assetStore.loadingAssetStates && rootData.length" class="action-bar">
					<span class="single-action-button" v-tooltip="'Refreshing Asset states'">
						<img class="small-icons loading-children-icon" :src="getAppIcon('loading')">
					</span>
				</div>

				<div v-else-if="!kanbanView && rootData.length" class="action-bar">
					<ActionButton v-if="isTasksRebuildable" :icon="getAppIcon('jigsaw')" v-tooltip="'Rebuild All'"
						:buttonFunction="rebuildAll" />
					<ActionButton v-if="isTasksUntracked && userStore.canDo('create_checkpoint')"
						:icon="getAppIcon('layers-plus-danger')" :noFilter="true" v-tooltip="'Create Checkpoints'"
						:buttonFunction="prepAllCheckpointModal" />
					<ActionButton v-else-if="isTasksModified && userStore.canDo('create_checkpoint')"
						:icon="getAppIcon('layers-plus-alert')" :noFilter="true" v-tooltip="'Create Checkpoints'"
						:buttonFunction="prepAllCheckpointModal" />
					<ActionButton v-if="isTasksModified" :icon="getAppIcon('revert-alert')" :noFilter="true" v-tooltip="'Revert All'"
						:buttonFunction="prepResetPopUpModal" />
					<ActionButton v-if="isTasksOutdated" :icon="getAppIcon('circle-check-alert')" :noFilter="true" v-tooltip="'Update all'"
						:buttonFunction="updateAll" />
				</div>
			</div>


			<div v-if="rootData.length || commonStore.viewSearchQuery.length || commonStore.showUntracked"
				class="view-options">
				<ViewOptions v-if="!kanbanView" :kanbanView="kanbanView"  />
				<ActionButton v-if="isDefaultWorkspace" :icon="kanbanView ? getAppIcon('list') : getAppIcon('kanban')"
					v-tooltip="kanbanView ? 'List' : 'Kanban'" :buttonFunction="toggleViewMode" />
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
				<ActionButton :icon="panes.showDetailsPane ? getAppIcon('collapse-right') : getAppIcon('collapse-left')"
					v-tooltip="panes.showDetailsPane ? 'Close pane' : 'Open pane'"
					:buttonFunction="toggleDetailsPane" />
			</div>

			<div v-else class="view-options">
				<ActionButton :icon="panes.showDetailsPane ? getAppIcon('collapse-right') : getAppIcon('collapse-left')"
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

			<TaskListSkeleton v-if="!assetStore.tasksLoaded" />
			<VirtuaScroll v-else-if="rootData.length && !commonStore.useGrid" :items="rootData" />
			<GridView v-else-if="rootData.length" :rootItems="rootData" />
			<PageState v-else :message="message()" :prompt="prompt()" :illustration="illustration()" />
		</div>

		<div v-else ref="taskListContainer" class="browser-root-container">
			<Kanban :filtersActive="filtersActive" :tasks="rootData" />
		</div>
	</div>
</template>

<script setup>
// imports
import { computed, ref, onMounted, onUpdated, watch, nextTick, onUnmounted, onBeforeUnmount } from 'vue';
import { useDebounce } from '@/lib/debounce';
import emitter from '@/lib/mitt';

// services
import { EntityService, TaskService, CheckpointService, TrashService } from "@/../bindings/clustta/services";
import { FSService, SyncService, DialogService } from '@/../bindings/clustta/services/index';

// state imports
import { useCommonStore } from '@/stores/common';
import { useCollectionStore } from '@/stores/collections';
import { useAssetStore } from '@/stores/assets';
import { useTrayStates } from '@/stores/TrayStates';
import { useStageStore } from '@/stores/stages';
import { useUserStore } from '@/stores/users';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { usePaneStore } from '@/stores/panes';
import { useProjectStore } from '@/stores/projects';
import { useNotificationStore } from '@/stores/notifications';
import { useMenu } from '@/stores/menu';
import { useTagStore } from '@/stores/tags';
import { useDndStore } from '@/stores/dnd';
import { useDependencyStore } from '@/stores/dependency';
import { useScrollStore } from '@/stores/scroll';
import { getRelativePath } from '@/lib/pathlib';
import { useWorkflowStore } from '@/stores/workflow';
import { useTemplateStore } from '@/stores/template';
import { syncData } from '@/lib/sync';

// states/stores
const trayStates = useTrayStates();
const stage = useStageStore();
const userStore = useUserStore();
const projectStore = useProjectStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const panes = usePaneStore();
const commonStore = useCommonStore();
const collectionStore = useCollectionStore();
const assetStore = useAssetStore();
const menu = useMenu();
const tagStore = useTagStore();
const dependencyStore = useDependencyStore();
const iconStore = useIconStore();
const scrollStore = useScrollStore();
const workflowStore = useWorkflowStore();
const templateStore = useTemplateStore();
const dndStore = useDndStore();

// components
import Breadcrumbs from '@/instances/common/components/Breadcrumbs.vue';
import FilterBar from '@/instances/common/components/FilterBar.vue';
import ViewOptions from '@/instances/common/components/ViewOptions.vue';
import VirtuaScroll from '@/instances/common/components/VirtuaScroll.vue';
import TaskListSkeleton from '@/instances/desktop/components/TaskListSkeleton.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import GhostItem from '@/instances/desktop/blocks/GhostItem.vue';
import GridView from '@/instances/desktop/components/GridView.vue';
import Kanban from '@/instances/desktop/components/Kanban.vue';
import PageState from '@/instances/common/components/PageState.vue';
import utils from '@/services/utils';

// refs
const taskListContainer = ref(null);
const browserRoot = ref(null);
const kanbanView = ref(false);
const showFilters = ref(false);
const browserFilters = ref(null);
const searchBar = ref(null);
const observer = ref(null);
const rootData = ref([]);

// events
import { Events } from "@wailsio/runtime";

const operationsActive = computed(() => {
	return stage.operationActive || !!modals.activeModal || !!menu.activeMenu || !assetStore.tasksLoaded || stage.activeStage !== 'browser'
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
	createTask();
});

Events.On('new-web-link', async () => {
	if (operationsActive.value) return
	createWebLink();
});

Events.On('sync-project', async () => {
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

Events.On('paste-items', async () => {
	if (operationsActive.value) return
	if (!!stage.cutItems && userStore.canDo('update_entity')) {
		console.log(stage.cutItems);
		stage.cutItems = [];
	}
});

Events.On('free-item-space', async () => {
	if (operationsActive.value) return
	if (stage.markedItems.length > 1) {
		prepFreeUpSpacePopUpModal();
	}
});

Events.On('window-focused', async () => {
	if (operationsActive.value) return
	// console.log('focused')
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
	duplicateTask();
});

const duplicateTask = async () => {

	const selectedItemId = stage.markedItems[0]
	const selectedItem = dndStore.allViewItems.find(item => item.id === selectedItemId)
	
	if (!selectedItem || selectedItem.type !== 'task') return
	
	try {
		stage.operationActive = true
		
		stage.markedItems = [];
		assetStore.selectedTask = null;

		await TaskService.DuplicateTask(projectStore.activeProject.uri, selectedItemId)
		.then((duplicatedTask) => {
			softRefresh();
			assetStore.selectTask(duplicatedTask);
			stage.selectedItem = duplicatedTask;
			stage.markedItems = [duplicatedTask.id];
			stage.lastSelectedItemId = "";
			stage.firstSelectedItemId = duplicatedTask.id;
			notificationStore.addNotification(`Asset duplicated`, '', "success", false)
		})
		
		
	} catch (error) {
		console.error('Error duplicating task:', error)
		notificationStore.errorNotification("Failed to duplicate task", error.message || error)
	} finally {
		stage.operationActive = false
	}

}

Events.On("clustta-drag-over", (event) => {
	return
	console.log('dragged')
	if (stage.selectedStage !== 'browser') {
		dndStore.isDragging = false;
		dndStore.isDropHovering = false;
		return
	}
	dndStore.dragPosition = {
		x: event.data[0].position.x,
		y: event.data[0].position.y,
	}
	dndStore.isDragging = true;
	dndStore.isDropHovering = true;
	nextTick(checkHoverState)
});

Events.On('clustta-drag-drop', async (event) => {
	return
	if (stage.selectedStage !== 'browser') {
		dndStore.isDragging = false;
		dndStore.isDropHovering = false;
		return
	}
	const paths = event.data[0].paths;
	const droppedItems = await categorizePaths(paths);
	dndStore.droppedFolders = droppedItems.folders;
	dndStore.droppedFiles = droppedItems.files;
	modals.setModalVisibility('importItemsModal', true);

	// reset
	dndStore.isDragging = false;
	dndStore.isDropHovering = false;
});

const checkHoverState = () => {

	if (!dndStore.isDragging) return
	dndStore.targetItemId = null;

	const allTargets = collectionStore.getEntities;
	for (let target of allTargets) {
		let targetEl = dndStore.itemRefs[target.id];
		if (!targetEl) {
			continue;
		}

		let dropZone = targetEl.querySelector('.drop-zone');
		const dropZoneRect = dropZone.getBoundingClientRect();

		dndStore.isOverlapping = dndStore.checkOverlap(dndStore.dragPosition, dropZoneRect);

		if (dndStore.isOverlapping && dndStore.targetItemId === target.id) {
			return requestAnimationFrame(checkHoverState);
		}
		else if (dndStore.isOverlapping) {
			dndStore.targetItemId = target.id;
			dndStore.targetItem = target;
			break;
		}
	}
}

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

const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

// computed properties
const isHovered = computed(() => { return dndStore.isDropHovering && dndStore.targetItemId === null });

const isDefaultWorkspace = computed(() => {
	return commonStore.activeWorkspace === 'Default';
});

const filtersActive = computed(() => {
	const assigneeFilters = commonStore.hasAssignees || commonStore.noAssignees;
	const entityFilters = commonStore.entityFilters.length > 0;
	const taskFilters = commonStore.taskFilters.length > 0;
	const resourceFilters = commonStore.resourceFilters.length > 0;
	const generalFilter = isFilterActive('general');
	return assigneeFilters || entityFilters || taskFilters || resourceFilters || generalFilter
});

const viewTags = computed(() => {
	let tags = tagStore.tags;
	let viewTags = [];
	let filteredTaskResults = assetStore.getFilteredTasks;

	for (const task of filteredTaskResults) {
		let taskTags = task.tags;
		for (let t = 0; t < taskTags.length; t++) {
			if (!viewTags.includes(taskTags[t])) {
				viewTags.push(taskTags[t])
			}
		}
	}

	for (let i = 0; i < tags.length; i++) {
		tags[i].name = tags[i].name;
		tags[i].type = 'tags'
	}
	const availableTags = tags;
	const filteredTags = availableTags.filter((item) => viewTags.includes(item.name));
	return filteredTags
});

const hideExtensionsFilter = computed(() => {
	let viewExtensions = [];
	let filteredTaskResults = assetStore.getFilteredTasks;

	for (const task of filteredTaskResults) {
		let taskExtension = task.extension;
		if (!viewExtensions.includes(taskExtension)) {
			viewExtensions.push(taskExtension)
		}
	}
	return viewExtensions.length && (commonStore.showTasks || commonStore.showResources)
});

const canModifyEntity = computed(() => {
	if (!collectionStore.selectedEntity) {
		return false
	}
	let selectedIsMarked = stage.markedItems.includes(collectionStore.selectedEntity.id)
	if (selectedIsMarked && stage.markedItems.length === 1) {
		return collectionStore.selectedEntity.can_modify
	} else {
		return false
	}
});

const showTagsFilter = computed(() => {
	return !!tagStore.tags.length && (commonStore.showTasks || commonStore.showResources)
});

const entityExpanded = computed(() => {
	return Object.keys(stage.expandedEntities).length;
});

const navigatedEntity = computed(() => {
	return collectionStore.navigatedEntity;
});

// computed getters
const projectEntities = computed(() => {

	const allEntities = collectionStore.getFilteredEntities;
	const trashedItems = trayStates.trashables;
	const trashedItemIds = trashedItems.map(item => item.id);
	const nestedEntities = allEntities.filter((item) => !item.parent_id);
	const showEntities = commonStore.showEntities;
	const searching = commonStore.viewSearchQuery;
	const useDeep = commonStore.useDeep;

	if (searching || useDeep) {
		if (showEntities) {
			return allEntities.filter((item) => !item.trashed && !trashedItemIds.includes(item.entity_id));
		} else {
			return []
		}
	} else {
		if (showEntities) {
			return nestedEntities.filter((item) => !item.trashed && !trashedItemIds.includes(item.entity_id));
		} else {
			return []
		}
	}
});

const projectTasks = computed(() => {
	// return []
	const allTasks = assetStore.getFilteredTasks.filter((item) => !item.is_resource);
	const trashedItems = trayStates.trashables;
	const trashedItemIds = trashedItems.map(item => item.id);
	const viewEntities = projectEntities.value;
	let viewEntityIds = [];

	for (let i = 0; i < viewEntities.length; i++) {
		viewEntityIds.push(viewEntities[i].id);
	}

	const viewTasks = allTasks.filter((item) => viewEntityIds.includes(item.entity_id));

	const nestedTasks = allTasks.filter((item) => !item.entity_id);
	const showEntities = commonStore.showEntities;
	const showTasks = commonStore.showTasks;
	const searching = commonStore.viewSearchQuery;

	if (searching) {
		if (showTasks) {
			// return tasks who's parent is an entity in the viewer
			return allTasks.filter((item) => !item.trashed && !trashedItemIds.includes(item.entity_id));
		} else {
			return []
		}
	} else {
		if (showTasks) {
			if (showEntities) {
				return nestedTasks.filter((item) => !item.trashed && !trashedItemIds.includes(item.entity_id));
			} else {
				return allTasks.filter((item) => !item.trashed && !trashedItemIds.includes(item.entity_id))
			}
		} else {
			return []
		}
	}
});

const projectUntrackedFiles = computed(() => {
	if (!commonStore.showChildTasks || !commonStore.showUntracked) {
		return []
	}

	const projectUntrackedFiles = projectStore.untrackedFiles;
	const rootUntrackedFolders = projectUntrackedFiles.filter((item) => item.entity_path === "");

	const showEntities = commonStore.showEntities;
	const showTasks = commonStore.showTasks;
	const searching = commonStore.viewSearchQuery;

	if (searching) {
		if (showTasks) {
			return projectUntrackedFiles.filter((item) => item.file_path.toLowerCase().replace(/\\/g, '/').includes(searching));;
		} else {
			return []
		}
	} else {
		if (showTasks) {
			if (showEntities) {
				return rootUntrackedFolders;
			} else {
				return projectUntrackedFiles;
			}
		} else {
			return []
		}
	}
});

const projectUntrackedFolders = computed(() => {
	if (!commonStore.showUntracked) {
		return []
	}

	const projectUntrackedFolders = projectStore.untrackedFolders;
	const rootUntrackedFolders = projectUntrackedFolders.filter((item) => item.entity_path === "");

	const showEntities = commonStore.showEntities;
	const searching = commonStore.viewSearchQuery;
	const useDeep = commonStore.useDeep;

	if (searching || useDeep) {
		if (showEntities) {
			return projectUntrackedFolders.filter((item) => item.file_path.toLowerCase().replace(/\\/g, '/').includes(searching));
		} else {
			return []
		}
	} else {
		if (showEntities) {
			return rootUntrackedFolders;
		} else {
			return []
		}
	}
});

const selectedEntity = computed(() => {
	return collectionStore.navigatedEntity
});

const isUntracked = computed(() => {
	return selectedEntity.value?.type !== 'entity'
})

const entityEntities = computed(() => {
	const entityId = selectedEntity.value?.id;
	const childEntities = collectionStore.getEntityChildren(entityId);
	return childEntities

});

const entityTasks = computed(() => {
	const entityId = selectedEntity.value?.id;
	return assetStore.getEntityTasks(entityId)?.filter((item) => !item?.is_resource)

});

const entityResources = computed(() => {

	const entityId = selectedEntity.value?.id;
	return assetStore.getEntityTasks(entityId)?.filter((item) => item.is_resource)

});

const entityUntrackedFiles = computed(() => {
	if (!commonStore.showChildTasks || !commonStore.showUntracked) {
		return []
	}
	let entityPath
	if (isUntracked.value) {
		entityPath = selectedEntity.value?.item_path;
	} else {
		entityPath = selectedEntity.value?.entity_path;
	}
	const projectUntrackedFiles = projectStore.untrackedFiles;

	let childUntrackedFiles
	if (isUntracked.value) {
		childUntrackedFiles = projectUntrackedFiles.filter((item) => item.entity_path && item.entity_path === selectedEntity.value?.item_path);
	} else {

		childUntrackedFiles = projectUntrackedFiles.filter((item) => item.entity_path && item.entity_path === selectedEntity.value?.entity_path);
	}

	return childUntrackedFiles;

});

const entityUntrackedFolders = computed(() => {
	if (!commonStore.showChildEntities || !commonStore.showUntracked) {
		return []
	}

	const entityUntrackedFolders = projectStore.untrackedFolders;
	let childUntrackedFolders
	if (isUntracked.value) {
		childUntrackedFolders = entityUntrackedFolders.filter((item) => item.entity_path && item.entity_path === selectedEntity.value?.item_path);
	} else {
		childUntrackedFolders = entityUntrackedFolders.filter((item) => item.entity_path && item.entity_path === selectedEntity.value?.entity_path);
	}

	return childUntrackedFolders;

});

const entityData = computed(() => {
	const allEntityData = [...entityEntities.value, ...entityUntrackedFolders.value, ...entityTasks.value, ...entityResources.value, ...entityUntrackedFiles.value];
	const validEntityData = allEntityData.filter((item) => !item?.trashed);
	const noDependencyData = validEntityData.filter((item) => !item?.is_dependency)
	return commonStore.showDependencies
		? validEntityData
		: noDependencyData;
});

const canModify = (task) => {
	let assigneeId = task.assignee_id
	if (assigneeId == "") {
		return true
	} else if (assigneeId == userStore.user.id) {
		return true
	} else {
		return false
	}
};

// state compute props
const isTasksModified = computed(() => {
	let path;
	path = collectionStore.navigatedEntity?.type === 'entity'
		? collectionStore.navigatedEntity?.entity_path
		: collectionStore.navigatedEntity?.item_path;


	const modifiedTasksPath = assetStore.modifiedTasksPath;
	let filteredPaths;

	filteredPaths = modifiedTasksPath.filter(item => item.startsWith(path));

	if (path) {
		filteredPaths = modifiedTasksPath.filter(item => item.startsWith(path));
	} else {
		filteredPaths = modifiedTasksPath;
	}

	return filteredPaths.length > 0;
});

const isTasksOutdated = computed(() => {

	let path = collectionStore.navigatedEntity?.entity_path;
	const outdatedTasksPath = assetStore.outdatedTasksPath;

	let filteredPaths;

	if (path) {
		filteredPaths = outdatedTasksPath.filter(item => item.startsWith(path));
	} else {
		filteredPaths = outdatedTasksPath;
	}

	return filteredPaths.length > 0;
});

const isTasksRebuildable = computed(() => {

	let path = collectionStore.navigatedEntity?.entity_path;
	const rebuildableTasksPath = assetStore.rebuildableTasksPath;

	let filteredPaths;

	if (path) {
		filteredPaths = rebuildableTasksPath.filter(item => item.startsWith(path));
	} else {
		filteredPaths = rebuildableTasksPath;
	}

	return filteredPaths.length > 0;
});

const isTasksUntracked = computed(() => {
	let path;
	path = collectionStore.navigatedEntity?.type === 'entity'
		? collectionStore.navigatedEntity?.entity_path
		: collectionStore.navigatedEntity?.item_path;


	const untrackedTasksPath = assetStore.untrackedTasksPath;
	let filteredPaths;

	filteredPaths = untrackedTasksPath.filter(item => item.startsWith(path));

	if (path) {
		filteredPaths = untrackedTasksPath.filter(item => item.startsWith(path));
	} else {
		filteredPaths = untrackedTasksPath;
	}

	return filteredPaths.length > 0;

});

// drag and drop
const draggedCard = computed(() => {
	return dndStore.allViewItems?.find(card => card.id === dndStore.draggedItemId);
});

const ghostCardStyles = computed(() => {
	return {
		width: `${dndStore.ghostCardStyle.width}px`,
		left: `${dndStore.ghostCardStyle.pos.x}px`,
		top: `${dndStore.ghostCardStyle.pos.y}px`,
		transform: dndStore.ghostCardStyle.transform,
	};
});

// methods
const onDrag = (e) => {
	dndStore.onDrag(e)
};

const onDragStop = async (event) => {
	let refresh = true;
	if (kanbanView.value) { return };
	if (dndStore.draggedItemId === null) return;

	document.documentElement.style.cursor = 'default';
	const dropTarget = dndStore.itemRefs[dndStore.targetItemId];
	const draggedItem = dndStore.itemRefs[dndStore.draggedItemId];
	let cardRect
	const targetEntity = dndStore.allViewItems.find((item) => item.id === dndStore.targetItemId);

	stage.operationActive = true;

	let draggedEntity;
	const draggedItemIds = stage.markedItems;
	// const draggedEntity = dndStore.allViewItems.find((item) => item.id === dndStore.draggedItemId);

	for (const draggedItemId of draggedItemIds) {

		draggedEntity = dndStore.allViewItems.find((item) => item.id === draggedItemId);

		if(!draggedEntity) return

		// if alt is pressed, to move the item to root
		if (event.altKey) {
			cardRect = draggedItem.getBoundingClientRect();
			if (draggedEntity.entity_type_id) {
				const entityId = draggedItemId;
				const parentId = '';
				await changeEntityParent(entityId, parentId);
			} else if (draggedEntity.task_type_id) {
				const taskId = draggedItemId;
				const entityId = '';
				await changeTaskEntity(taskId, entityId);

			} else {

				let extension = draggedEntity.type === 'untracked_task' ? draggedEntity.extension : '';
				let fullName = draggedEntity.name + extension
				
				await FSService.MakeDirs(projectStore.activeProject.working_directory)
				let newPath = await FSService.JoinPath(projectStore.activeProject.working_directory, fullName)
				await FSService.Rename(draggedEntity.file_path, newPath);
			}
		}

		// if it overlaps another item
		else if (dndStore.isOverlapping && dropTarget) {

			cardRect = dropTarget.getBoundingClientRect();

			// if target item is not the same as dragged
			if (draggedItemId !== dndStore.targetItemId) {

				// if this is an entity make the selected items children
				if (targetEntity.type === 'entity') {
					if (draggedEntity.type === 'entity') {
						const entityId = draggedItemId;
						const parentId = dndStore.targetItemId;
						await changeEntityParent(entityId, parentId);
					} else if (draggedEntity.type === 'task') {
						const taskId = draggedItemId;
						const entityId = dndStore.targetItemId;
						await changeTaskEntity(taskId, entityId);
					} else {

  						let entity = await EntityService.GetEntityByID(projectStore.activeProject.uri, dndStore.targetItemId)
						await FSService.MakeDirs(entity.file_path)
						let extension = draggedEntity.type === 'untracked_task' ? draggedEntity.extension : '';
						let fullName = draggedEntity.name + extension
						let newPath = await FSService.JoinPath(entity.file_path, fullName)
						const untrackedPath = newPath.replace(/^\/+|\/+$/g, "").replace(/\\/g, "/");
						const workingDir = projectStore.activeProject.working_directory.replace(/^\/+|\/+$/g, "").replace(/\\/g, "/");

						const itemPath = getRelativePath(workingDir, untrackedPath)

						let entityPath = "";
						const itemPathEntities = itemPath.split("/");
						if (itemPathEntities.length > 1) {
							// Take all elements except the last one
							const pathWithoutLast = itemPathEntities.slice(0, -1);
							entityPath = pathWithoutLast.join("/");
						}
						await FSService.Rename(draggedEntity.file_path, newPath);
					}
				}
				// if this is a task make the selected items dependencies
				else if (targetEntity.task_type_id) {
					// refresh = false;
					let dependencyTypeId = dependencyStore.dependency_types.find(item => item.name === "linked").id;
					if (draggedEntity.entity_type_id) {
						const taskId = dndStore.targetItemId;
						const dependencyId = draggedItemId;
						await addEntityDependency(taskId, dependencyId, dependencyTypeId);
					} else if (draggedEntity.task_type_id) {
						const taskId = dndStore.targetItemId;
						const dependencyId = draggedItemId;
						await addDependency(taskId, dependencyId, dependencyTypeId);
					} else if (draggedEntity.item_type) {
						console.log('cant drop here')
					} else {
						const taskId = dndStore.targetItemId;
						const dependencyId = draggedItemId;
						await addDependency(taskId, dependencyId, dependencyTypeId);
					}
				}
				// if this is an untracked folder
				else if (targetEntity.type === 'untracked_entity') {
					if (draggedEntity.entity_type_id || draggedEntity.task_type_id) {
						console.log('cant drop here')
					} else if (draggedEntity.type === 'untracked_task' || draggedEntity.type === 'untracked_entity') {
						console.log('untracked')

						let entity = targetEntity;
						// await FSService.MakeDirs(entity.file_path)
						let extension = draggedEntity.type === 'untracked_task' ? draggedEntity.extension : '';
						let fullName = draggedEntity.name + extension
						
						let newPath = await FSService.JoinPath(entity.file_path, fullName)

						console.log(draggedEntity)

						await FSService.Rename(draggedEntity.file_path, newPath)

						console.log('moved untracked file into untracked Collection')
					}
				}
				// if this is an untracked file
				else if (targetEntity.item_type === 'file') {
					console.log('cant drop here');
				}
			} else {
				// target is the same
				console.log('same item - reset');
				cardRect = draggedItem.getBoundingClientRect();

			}

		} else {
			// if no target, go on to reset
			console.log('no target')
			cardRect = draggedItem.getBoundingClientRect();
		}
	}

	setTimeout(() => {
		dndStore.resetValues();
	}, 100);

	dndStore.ghostCardStyle.leaving = true;
	let xOffset = cardRect.x - dndStore.ghostCardStyle.pos.x;
	let yOffset = cardRect.y - dndStore.ghostCardStyle.pos.y;
	dndStore.ghostCardStyle.transform = `scale(1) translate(${xOffset}px, ${yOffset}px)`;

	if(refresh) softRefresh();
	
	stage.operationActive = false;

};

const changeEntityParent = async (entityId, parentId) => {

	await EntityService.ChangeEntityParent(projectStore.activeProject.uri, entityId, parentId)
		.then((response) => {
			const successMessage = 'Moved successfully.'
			notificationStore.addNotification(successMessage, "", "success")
		})
		.catch((error) => {
			console.error(error);
			notificationStore.errorNotification("Error changing entity parent", error)
		});
};

const changeTaskEntity = async (taskId, entityId) => {
	await TaskService.ChangeTaskEntity(projectStore.activeProject.uri, taskId, entityId)
		.then((response) => {
			const successMessage = 'Moved successfully.'
			notificationStore.addNotification(successMessage, "", "success")
		})
		.catch((error) => {
			notificationStore.errorNotification("Error changing task entity", error)
		});
};

const addDependency = async (taskId, dependencyId, dependencyTypeId) => {
	await TaskService.AddTaskDependency(projectStore.activeProject.uri, taskId, dependencyId, dependencyTypeId)
		.then((response) => {
			// assetStore.addDependency(taskId, dependencyId, "task");
			const successMessage = 'Dependency Added.'
			notificationStore.addNotification(successMessage, "", "success")
		})
		.catch((error) => {
			console.log(error)
			notificationStore.errorNotification("Error adding dependencies", error);
		});
};

const addEntityDependency = async (taskId, dependencyId, dependencyTypeId) => {
	await TaskService.AddEntityDependency(projectStore.activeProject.uri, taskId, dependencyId, dependencyTypeId)
		.then((response) => {
			// assetStore.addDependency(taskId, dependencyId, "entity");
			const successMessage = 'Dependency Added.'
			notificationStore.addNotification(successMessage, "", "success")
		})
		.catch((error) => {
			console.log(error)
			notificationStore.errorNotification("Error adding dependencies", error);
		});
};

const clearSearch = async () => {
	commonStore.viewSearchQuery = "";
	await softRefresh();
};

const toggleShowFilters = () => {
	showFilters.value = !showFilters.value;
};

const clearFilters = () => {
	commonStore.resetFilters();
	softRefresh();
};

const isFilterActive = (filter) => {
	if (filter.includes('general')) {
		const isActive = commonStore.showEntities && commonStore.showTasks
			&& commonStore.showResources && commonStore.showChildEntities
			&& commonStore.showChildTasks && commonStore.showDependencies && !commonStore.onlyAssets;
		return !isActive;
	} else
		if (filter.includes('entity')) {
			return commonStore.entityFilters.some((item) => item.type === filter);
		} else if (filter.includes('assignation')) {
			const assigneeFilters = commonStore.hasAssignees || commonStore.noAssignees;
			const assignationFilters = commonStore.taskFilters.some((item) => item.type === filter);
			return assigneeFilters || assignationFilters;
		} else {
			return commonStore.taskFilters.some((item) => item.type === filter);
		}
};

const updateSearch = async (event) => {

	if (scrollStore.scrollTop > 0) {
		scrollStore.requestScroll(0);
	}
	if (entityExpanded.value) {
		collapseAll();
	}

	const searchQuery = event.target.value;
	commonStore.viewSearchQuery = searchQuery.toLowerCase();

	await softRefresh();

	const searchContext = commonStore.navigatorMode;

};

const debouncedUpdateSearch = useDebounce(updateSearch, 300);

const prepFreeUpSpacePopUpModal = () => {
	trayStates.popUpModalTitle = "Free up Space";
	trayStates.popUpModalMessage = "Are you sure you want to delete these items? This will permanently remove all uncheckpointed resources and all task outputs. Please confirm if you wish to proceed.";
	trayStates.popUpModalIcon = 'broom';
	trayStates.popUpModalFunction = freeUpSpace;
	modals.setModalVisibility('popUpModal', true);
};

const freeUpSpace = async () => {
	panes.setPaneVisibility('projectDetails', true);
	stage.operationActive = true;

	stage.firstSelectedItemId = '';
	stage.lastSelectedItemId = '';
	assetStore.selectedTask = null;
	collectionStore.selectedEntity = null;

	const allItemsToDelete = dndStore.allViewItems.filter((item) => stage.markedItems.includes(item.id))

	const tasksToFreeUp = allItemsToDelete.filter((item) => item.type === 'task');

	const entitiesToFreeUp = allItemsToDelete.filter((item) => item.type === 'entity');

	const untrackedTasksToDelete = allItemsToDelete.filter((item) => item.type === 'untracked_task');

	const untrackedEntitiesToDelete = allItemsToDelete.filter((item) => item.type === 'untracked_entity');

	await freeUpMultipleTaskSpace(tasksToFreeUp);
	await freeUpMultipleEntitySpace(entitiesToFreeUp);
	await deleteMultipleUntrackedTasks(untrackedTasksToDelete);
	await deleteMultipleUntrackedEntities(untrackedEntitiesToDelete);

	stage.markedItems = [];
	stage.selectedItems = [];
	stage.markedTasks = [];
	stage.markedEntities = [];
	stage.operationActive = false;
	modals.setModalVisibility('popUpModal', false);
};

const freeUpMultipleTaskSpace = async (selectedTasks) => {
	const fileStatus = ['missing', 'rebuildable'];
	let taskIds = [];
	for (let task of selectedTasks) {
		if (!fileStatus.includes(task.file_status)) {
			let taskId = task.id
			taskIds.push(taskId)
		}
	};
	for (const taskId of taskIds) {
		let task = assetStore.getTasks.find((item) => item.id === taskId);
		let taskPath = task.file_path.replace(/\\/g, '/')
		await FSService.DeleteFile(taskPath)
			.then((response) => {
				task.file_status = 'rebuildable';
			})
			.catch((error) => {
				console.error(error);
			});
	}
};

const freeUpMultipleEntitySpace = async (entities) => {
	for (const entity of entities) {
		let entityDir = entity.file_path.replace(/\\/g, '/');
		await FSService.DeleteFolder(entityDir)
			.then((response) => {
				assetStore.refreshEntityFilesStatus(entity.id)
			})
			.catch((error) => {
				console.error(error);
			});
	}

};

const prepDeleteMultipleItemsPopUpModal = () => {
	const numberOfItems = stage.markedItems.length;
	trayStates.popUpModalTitle = "Delete " + numberOfItems + " items";
	trayStates.popUpModalMessage = "You have selected some untracked/modified items and they will be permanently deleted. Continue?";
	trayStates.popUpModalIcon = 'trash';
	trayStates.popUpModalFunction = deleteMultipleItems;
	modals.setModalVisibility('popUpModal', true);
};

const deleteMultipleItems = async () => {
	panes.setPaneVisibility('projectDetails', true);
	stage.operationActive = true;

	stage.firstSelectedItemId = '';
	stage.lastSelectedItemId = '';
	assetStore.selectedTask = null;
	collectionStore.selectedEntity = null;

	const allItemsToDelete = dndStore.allViewItems.filter((item) => stage.markedItems.includes(item.id))

	const tasksToDelete = allItemsToDelete.filter((item) => item.type === 'task');
	const taskIds = tasksToDelete.map((item) => item.id)

	const entitiesToDelete = allItemsToDelete.filter((item) => item.type === 'entity');
	const entityIds = entitiesToDelete.map((item) => item.id);

	const untrackedTasksToDelete = allItemsToDelete.filter((item) => item.type === 'untracked_task');

	const untrackedEntitiesToDelete = allItemsToDelete.filter((item) => item.type === 'untracked_entity');

	await deleteMultipleTasks(taskIds);
	await deleteMultipleEntities(entityIds);
	await deleteMultipleUntrackedTasks(untrackedTasksToDelete);
	await deleteMultipleUntrackedEntities(untrackedEntitiesToDelete);

	stage.markedItems = [];
	stage.selectedItems = [];
	stage.markedTasks = [];
	stage.markedEntities = [];
	stage.operationActive = false;
	modals.setModalVisibility('popUpModal', false);
};

const deleteMultipleTasks = async (taskIds) => {
	for (let taskId of taskIds) {
		await TaskService.DeleteTask(projectStore.activeProject.uri, taskId, true)
			.then(async (response) => {
				softRefresh()
				notificationStore.addNotification("Tasks moved to Trash.", '', "success", false);
			})
			.catch((error) => {
				console.log(error)
				notificationStore.errorNotification("Tasks failed to delete.", error)
			});
	}
};

const deleteMultipleEntities = async (entityIds) => {
	for (let entityId of entityIds) {
		await EntityService.DeleteEntity(projectStore.activeProject.uri, entityId, true)
			.then(async (response) => {
				await collectionStore.markEntityAsDeleted(entityId);
				notificationStore.addNotification("Entity moved to Trash.", '', "success", false);
			})
			.catch((error) => {
				console.log(error)
				notificationStore.errorNotification("Entities failed to delete.", error)
			});
	}
};

const deleteMultipleUntrackedTasks = async (untrackedTasks) => {
	for (let untrackedTask of untrackedTasks) {
		await FSService.DeleteFile(untrackedTask.file_path);
		projectStore.removeUntrackedTask(untrackedTask.id);
	}
};

const deleteMultipleUntrackedEntities = async (untrackedEntities) => {
	for (let untrackedEntity of untrackedEntities) {
		FSService.DeleteFolder(untrackedEntity.file_path);
		projectStore.removeUntrackedEntity(untrackedEntity.id);
	}
};

// page state fns
const message = () => {

	const searching = commonStore.viewSearchQuery;
	// const filtersActive = commonStore.activeFilters.length;
	const myTasksWorkspace = commonStore.activeWorkspace === 'My Tasks';

	if (searching) {
		return 'No results found.'
	} else if (isDefaultWorkspace.value && filtersActive.value) {
		return 'No results match your filters.'
	} else if (myTasksWorkspace) {
		return 'You have no assets assigned to you.'
	} else if (!isDefaultWorkspace.value) {
		return 'Nothing in this workspace.'
	} else {
		return 'Nothing to see here.'
	}

};

const prompt = () => {

	const searching = commonStore.viewSearchQuery;

	if (searching) {
		return ''
	} else if (!isDefaultWorkspace.value || filtersActive.value) {
		return ''
	} else {
		return 'Right click to create a new Collection or Asset.'
	}

};

const illustration = () => {

	const searching = commonStore.viewSearchQuery;

	if (searching) {
		return '/page-states/resources.png'
	} else {
		return '/page-states/tasks.png'
	}
};

const openMenu = (event) => {
	assetStore.selectedTask = null;
	collectionStore.selectedEntity = null;
	projectStore.selectedUntrackedItem = null;
	handleClickOutside(event, true)
	if (kanbanView.value) {
		return
	}
	menu.showContextMenu(event, 'projectMenu', true);
};

const cancelOps = () => {
	stage.cutItems = [];
	if (!dndStore.altKeyActive) {
		dndStore.resetValues();
	}
}

const disableMenus = () => {
	menu.disableAllMenus();
};

const createEntity = () => {
	clearSelection();
	modals.setModalVisibility('createCollectionModal', true);
};

const createWorkflow = () => {
	clearSelection();
	modals.setModalVisibility('selectWorkflowModal', true);
};

const createTask = () => {
	clearSelection();
	modals.setModalVisibility('selectAppModal', true);
};

const createWebLink = () => {
	clearSelection();
	modals.setModalVisibility('addWebLinkModal', true);
};

const importItems = async () => {

	try {

		let selectedPaths;
		try {
			selectedPaths = await DialogService.SelectFilesDialog();
		} catch (error) {
			// User cancelled the dialog
			return;
		}

		if (!selectedPaths || selectedPaths.length === 0) {
			// No files selected (but dialog wasn't cancelled)
			return;
		}

		// Get current directory for copying
		const currentDirectory = getCurrentDirectory();
		if (!currentDirectory) {
			notificationStore.errorNotification("Could not determine current directory", "");
			return;
		}

		
		// Show operation in progress
		stage.operationActive = true;
		await FSService.MakeDirs(currentDirectory);
		
		let successCount = 0;
		let failureCount = 0;
		const errors = [];

		// Process each selected path
		for (const sourcePath of selectedPaths) {
			try {
				const isFile = await FSService.IsFile(sourcePath);
				const itemName = await FSService.BaseName(sourcePath);
				
				// Generate unique destination path
				const destinationPath = await generateUniqueDestinationPath(currentDirectory, itemName);
				
				if (isFile) {
					// Copy file
					await FSService.DuplicateFile(sourcePath, destinationPath);
				} else {
					// Copy folder
					await FSService.DuplicateFolder(sourcePath, destinationPath);
				}
				
				successCount++;
			} catch (error) {
				failureCount++;
				const itemName = await FSService.BaseName(sourcePath).catch(() => sourcePath);
				errors.push(`${itemName}: ${error.message || error}`);
			}
		}

		// Show results
		if (successCount > 0) {
			const message = successCount === 1 
				? "1 item imported successfully" 
				: `${successCount} items imported successfully`;
			notificationStore.addNotification(message, "", "success");
		}

		if (failureCount > 0) {
			const message = failureCount === 1 
				? "1 item failed to import" 
				: `${failureCount} items failed to import`;
			notificationStore.errorNotification(message, errors.join("\n"));
		}

		// Refresh the view to show imported items
		if (successCount > 0) {
			await softRefresh();
		}

	} catch (error) {
		notificationStore.errorNotification("Error importing items", error.message || error);
	} finally {
		stage.operationActive = false;
	}
};

const getCurrentDirectory = () => {
	// If in navigator mode (viewing inside an entity/folder)
	if (commonStore.navigatorMode && collectionStore.navigatedEntity) {
		const navigated = collectionStore.navigatedEntity;
		// Return the file path of the current entity or folder
		return navigated.file_path;
	}
	
	// If at project root
	return projectStore.activeProject?.working_directory;
};

const generateUniqueDestinationPath = async (directory, fileName) => {
	const originalPath = await FSService.JoinPath(directory, fileName);
	
	// Check if file/folder already exists
	const exists = await FSService.Exists(originalPath);
	if (!exists) {
		return originalPath;
	}
	
	// Generate unique name with counter
	const baseName = fileName.includes('.') 
		? fileName.substring(0, fileName.lastIndexOf('.'))
		: fileName;
	const extension = fileName.includes('.') 
		? fileName.substring(fileName.lastIndexOf('.'))
		: '';
	
	let counter = 1;
	let newPath;
	
	do {
		const newFileName = `${baseName} (${counter})${extension}`;
		newPath = await FSService.JoinPath(directory, newFileName);
		const pathExists = await FSService.Exists(newPath);
		if (!pathExists) {
			return newPath;
		}
		counter++;
	} while (counter < 100); // Safety limit
	
	// Fallback with timestamp if we hit the limit
	const timestamp = Date.now();
	const timestampFileName = `${baseName}_${timestamp}${extension}`;
	return await FSService.JoinPath(directory, timestampFileName);
};

const toggleDetailsPane = () => {
	panes.showDetailsPane = !panes.showDetailsPane;
};

const toggleViewMode = () => {
	kanbanView.value = !kanbanView.value;
	softRefresh();
	panes.showDetailsPane = !kanbanView.value;
};

const toggleHideExtensions = () => {
	commonStore.hideExtensions = !commonStore.hideExtensions;
};

const toggleShowFullPath = () => {
	commonStore.showFullPath = !commonStore.showFullPath;
};

const toggleExpandEntities = () => {
	if (entityExpanded.value) {
		collapseAll()
	} else {
		expandAll()
	}
};

const collapseAll = () => {
	stage.expandedEntities = {};
	stage.markedEntities = [];
	stage.firstSelectedEntityId = '';
	collectionStore.selectedEntity = null;
};

const expandAll = () => {
	const entities = collectionStore.getEntities;
	const expandedEntities = {};

	for (let i = 0; i < entities.length; i++) {
		expandedEntities[entities[i].id] = {
			"height": 0,
			"entity_path": entities[i].entity_path
		};
	}

	stage.expandedEntities = expandedEntities;
};

const rebuildAll = async () => { 

	const path = collectionStore.navigatedEntity?.entity_path;
	const navigatedEntityId = collectionStore.navigatedEntity?.id;
	const rebuildableTasksPath = assetStore.rebuildableTasksPath;

	notificationStore.cancleFunction = SyncService.CancelSync;
	notificationStore.canCancel = true;

	if (commonStore.activeWorkspace === 'My Tasks' && rootData.value.length) {

		const userTasks = rootData.value;
		const userTaskIds = userTasks.map((task) => task.id)
		const userTaskPaths = userTasks.map(task => task.file_path);

		await CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, userTaskIds)
		.then(() => {
			
			assetStore.rebuildableTasksPath = rebuildableTasksPath.filter(taskPath => 
				!userTaskPaths.some(userTaskPath => taskPath.startsWith(userTaskPath))
			);
			softRefresh();
			return;
		})
		.catch((error) => {
			console.error(`Error rebuilding task ${task.id}:`, error);
		});

	} else {

		await EntityService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, navigatedEntityId)
			.then((data) => {

				if(path){
					assetStore.rebuildableTasksPath = rebuildableTasksPath.filter(item => !item.startsWith(path))
				} else {
					assetStore.rebuildableTasksPath = [];
				}

				softRefresh();
			}).catch(async (error) => {
				notificationStore.errorNotification("Error Rebuilding All", error)
			})
	}
};

const revertAllChanges = async () => {
	modals.setModalVisibility('popUpModal', false);
	
	// Get current navigation context
	let path;
	path = collectionStore.navigatedEntity?.type === 'entity'
		? collectionStore.navigatedEntity?.entity_path
		: collectionStore.navigatedEntity?.item_path;

	const modifiedTasksPath = assetStore.modifiedTasksPath;
	let filteredPaths;

	// Filter paths based on current context
	if (path) {
		filteredPaths = modifiedTasksPath.filter(item => item.startsWith(path));
	} else {
		filteredPaths = modifiedTasksPath;
	}

	await CheckpointService.RevertTaskPaths(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, filteredPaths)
		.then((response) => {
			assetStore.modifiedTasksPath = assetStore.modifiedTasksPath.filter(item => !filteredPaths.includes(item));
			softRefresh();
		})
		.catch((error) => {
			notificationStore.errorNotification("Failed to Revert Tasks", error)
			console.error(error);
		});
};

const updateAllModified = async () => {
	modals.setModalVisibility('popUpModal', false);
	
	// Get current navigation context
	let path;
	path = collectionStore.navigatedEntity?.type === 'entity'
		? collectionStore.navigatedEntity?.entity_path
		: collectionStore.navigatedEntity?.item_path;

	const outdatedTasksPath = assetStore.outdatedTasksPath;
	let filteredPaths;

	// Filter paths based on current context
	if (path) {
		filteredPaths = outdatedTasksPath.filter(item => item.startsWith(path));
	} else {
		filteredPaths = outdatedTasksPath;
	}

	await CheckpointService.RevertTaskPaths(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, filteredPaths)
		.then((response) => {
			assetStore.outdatedTasksPath = assetStore.outdatedTasksPath.filter(item => !filteredPaths.includes(item));
			softRefresh();
		})
		.catch((error) => {
			notificationStore.errorNotification("Failed to Revert Tasks", error)
			console.error(error);
		});
};

const prepResetPopUpModal = () => {
	clearSelection();
	trayStates.popUpModalIcon = 'revert'
	trayStates.popUpModalTitle = "Revert All Changes";
	trayStates.popUpModalMessage = "All Modified tasks will be reverted to their last saved state. Are you sure you want to continue?";
	trayStates.popUpModalFunction = revertAllChanges;
	modals.setModalVisibility('popUpModal', true);
};

const prepAllCheckpointModal = () => {
	clearSelection();
	trayStates.createMultipleCheckpoints = true;
	trayStates.createMultipleCheckpointsEntityPath = ""
	modals.setModalVisibility('createMultipleCheckpointsModal', true);
};

const updateAll = () => {
	clearSelection();
	updateAllModified();
};

const clearSelection = () => {
	stage.markedItems = [];
	stage.selectedItems = [];
	stage.firstSelectedItemId = '';
	stage.lastSelectedItemId = '';
	assetStore.selectedTask = null;
	collectionStore.selectedEntity = null;
}

const handleClickOutside = (event, rightClick = false) => {
	// return
	if (!rightClick) {
	}
	if (event) {
		if (!event.shiftKey || !event.ctrlKey) {
			if (!event.target.closest('.entity-item-main')) {
				stage.markedItems = [];
				stage.selectedItems = [];
				stage.markedEntities = [];
				stage.firstSelectedItemId = '';
				stage.lastSelectedItemId = '';
				stage.selectedItem = null;
				assetStore.selectedTask = null;
				collectionStore.selectedEntity = null;
				stage.cutItems = [];
				projectStore.selectedUntrackedItem = null;
			}
			if (!event.target.closest('.task-item-main')) {
				stage.markedItems = [];
				stage.selectedItems = [];
				stage.markedEntities = [];
				stage.firstSelectedItemId = '';
				stage.lastSelectedItemId = '';
				stage.selectedItem = null;
				assetStore.selectedTask = null;
				stage.cutItems = [];
				projectStore.selectedUntrackedItem = null;
			}
		}
	}
};

const trackWidthChange = (entries) => {
	calculateSpace();
};

const calculateSpace = () => {
	let filterWidth;
	let rootWidth;
	if (browserFilters.value) {
		filterWidth = browserFilters.value.getBoundingClientRect().width;
	}
	if (browserRoot.value) {
		rootWidth = browserRoot.value.getBoundingClientRect().width;
	}
};

const detectModifier = (event) => {
	if (event.getModifierState('Alt')) {
		dndStore.altKeyActive = true;
	} else {
		dndStore.altKeyActive = false;
	}
};

const refresh = async () => {
	if(kanbanView.value){
		return
	}
	assetStore.tasksLoaded = false;
	assetStore.modifiedTasksPath = []
	assetStore.untrackedTasksPath = [];
	stage.cutItems = [];
	await projectStore.refreshActiveProject();
	await trayStates.refreshData();
	let children;
	let project = projectStore.activeProject

	if (!commonStore.navigatorMode) {
		children = await EntityService.GetEntityChildren(project.uri, "root", project.working_directory, project.working_directory, project.ignore_list, false)
	} else {
		const navigatedEntityId = collectionStore.navigatedEntity?.id;
		const entity_file_path = collectionStore.navigatedEntity?.file_path;
		children = await EntityService.GetEntityChildren(project.uri, navigatedEntityId, project.working_directory, entity_file_path, project.ignore_list, false)
	}

	await assetStore.processTasksIconsAndPreviews(children.tasks);

	await assetStore.processUntrackedTasksIcons(children.untracked_tasks);

	rootData.value = [...children.entities, ...children.untracked_entities, ...children.tasks, ...children.untracked_tasks];
	assetStore.tasksLoaded = true;
	console.log(children);

	assetStore.reloadAssetStates();

};

const softRefresh = async () => {
	assetStore.tasksLoaded = false;
	stage.cutItems = [];

	let children = {};
	let project = projectStore.activeProject

	const searching = commonStore.viewSearchQuery.toLowerCase();

	if (searching || filtersActive.value) {

		let entities;
		let tasks;

		if (!commonStore.navigatorMode) {
			if(!searching){

				// fetch only root items
				const rootItems = await EntityService.GetEntityChildren(project.uri, "root", project.working_directory, project.working_directory, project.ignore_list, false);

				entities = rootItems['entities'];
				tasks = commonStore.onlyAssets ? await TaskService.GetTasks(project.uri) : rootItems['tasks'];

			} else {

				// fetch everything
				entities = await EntityService.GetEntities(project.uri);
				tasks = await TaskService.GetTasks(project.uri);
			}

			entities = commonStore.onlyAssets ? [] : entities;

		} else {

			const navigatedEntityId = collectionStore.navigatedEntity?.id;
			const entity_file_path = collectionStore.navigatedEntity?.file_path;
			
			const entityItems = await EntityService.GetEntityChildren(project.uri, navigatedEntityId, project.working_directory, entity_file_path, project.ignore_list, false)

			entities = entityItems['entities'];
			tasks = entityItems['tasks'];
		}	
		
			children['entities'] = await collectionStore.filterEntities(entities);
			children['tasks'] = await assetStore.filterTasks(tasks);

	} else {
		if (!commonStore.navigatorMode) {

			children = await EntityService.GetEntityChildren(project.uri, "root", project.working_directory, project.working_directory, project.ignore_list, false)

		} else {
			const navigatedEntityId = collectionStore.navigatedEntity?.id;
			const entity_file_path = collectionStore.navigatedEntity?.file_path;
			children = await EntityService.GetEntityChildren(project.uri, navigatedEntityId, project.working_directory, entity_file_path, project.ignore_list, false);
		}
	}

	if (children.tasks) {
		await assetStore.processTasksIconsAndPreviews(children.tasks);
	}

	if (children.untracked_tasks) {
		await assetStore.processUntrackedTasksIcons(children.untracked_tasks);
	}



	const allEntities = commonStore.showEntities ? children.entities?.filter((item) => !item.is_trashed) : [];
	const allTasks = commonStore.showTasks ? children.tasks : [];

	rootData.value = [...(allEntities ?? []), ...(allTasks ?? []), ...(children.untracked_entities ?? []), ...(children.untracked_tasks ?? [])];

	assetStore.tasksLoaded = true;
	assetStore.reloadAssetStates();
};

watch(() => assetStore.tasksLoaded, async () => {
	if (assetStore.tasksLoaded) {
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

watch(() => collectionStore.navigatedEntity, async () => {
	await softRefresh();
});

const assigneeFilters = computed(() => {
	return commonStore.hasAssignees || commonStore.noAssignees
});

const filterItems = ref([commonStore.taskFilters.length, assigneeFilters.value, commonStore.showTasks, commonStore.showEntities])

watch(() => commonStore.showTasks, async () => {
	stage.operationActive = true
	await softRefresh();
	stage.operationActive = false
});

watch(() => commonStore.navigatorMode, async () => {
	await softRefresh();
});

const handleUpdateRootData = (eventData) => {
	const { itemId, property, value, updates } = eventData;
	
	// Find the item in rootData
	const itemIndex = rootData.value.findIndex(item => item.id === itemId);
	if (itemIndex !== -1) {
		// Handle single property update (backward compatibility)
		if (property && value !== undefined) {
			rootData.value[itemIndex][property] = value;
		}
		// Handle multiple property updates
		if (updates && Array.isArray(updates)) {
			updates.forEach(update => {
				rootData.value[itemIndex][update.property] = update.value;
			});
		}
	}
};

onMounted(async () => {
	commonStore.resetFilters();
	commonStore.activeWorkspace = 'Default';
	panes.showDetailsPane = true;
	observer.value = new ResizeObserver(trackWidthChange);
	observer.value.observe(browserRoot.value);
	document.addEventListener('click', handleClickOutside);
	window.addEventListener('keydown', detectModifier);
	window.addEventListener('keyup', detectModifier);
	emitter.on('refresh-browser', softRefresh);
	emitter.on('update-root-data', handleUpdateRootData);
	// emitter.on('reload-asset-states', reloadAssetStates);

	await refresh();
	dndStore.triggerDomUpdate()

	trayStates.trashables = await TrashService.GetTrashs(projectStore.activeProject.uri);


});

onUnmounted(() => {
	emitter.off('refresh-browser', refresh);
	emitter.off('update-root-data', handleUpdateRootData);
	// emitter.off('reload-asset-states', reloadAssetStates);
	disableMenus();
});



onBeforeUnmount(() => {
	stage.expandedEntities = {};
	stage.markedEntities = [];
	panes.showDetailsPane = false;
	document.removeEventListener('click', handleClickOutside);
	window.removeEventListener('keydown', detectModifier);
	window.removeEventListener('keyup', detectModifier);
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.intersector {
	z-index: 1000;
	position: fixed;
	left: 0;
	top: 0;
	top: 40vh;
	width: 100vw;
	height: 40vh;
	/* height: 90vh; */
	background: rgba(255, 0, 0, 0.1);
	pointer-events: none;
	/* border: 1px solid rgba(255, 0, 0, 0.3); */
}

.dash-board {
	padding: .4rem;
	display: flex;
	gap: .4rem;
	flex-direction: column;
	box-sizing: border-box;
	height: 100%;
	width: 100%;
	height: max-content;
}

.filter-root {
	width: 100%;
	display: flex;
	/* background-color: firebrick; */
	background-color: var(--black-steel);
	border-radius: var(--normal-radius);
	align-items: center;
	box-sizing: border-box;
	padding: .2rem;
	flex-direction: column;
}

.filter-header {
	width: 100%;
	display: flex;
	background-color: var(--steel);
	border-radius: var(--small-radius);
	/* background-color: green; */
	align-items: center;
	box-sizing: border-box;
	padding: 0rem .2rem;
}

.filter-header-tabs {
	box-sizing: border-box;
	width: 100%;
	width: min-content;
	/* flex: 1; */
	height: 100%;
	height: min-content;
	display: flex;
	/* background-color: purple; */
	align-items: center;
}

.selected-filters {
	box-sizing: border-box;
	width: 100%;
	overflow: hidden;
	/* flex: 1; */
	height: 100%;
	display: flex;
	/* background-color: goldenrod; */
	align-items: center;
}

.dash-board-container {
	z-index: 5;
	position: relative;
	display: flex;
	padding-right: .4rem;
	overflow: hidden;
	height: 100%;
	width: 100%;
	box-sizing: border-box;
	min-width: 300px;
}

.dash-board-container::-webkit-scrollbar {
	width: 8px;
}

.dash-board-container::-webkit-scrollbar-thumb {
	border-radius: 10px;
	background-color: var(--dark-steel);
}

.dash-board-container::-webkit-scrollbar-track {
	border-radius: 10px;
}

.browser-root-container {
	z-index: 5;
	position: relative;
	flex-direction: column;
	padding: .5rem;
	overflow: hidden;
	height: 100%;
	background-color: tomato;
	border-radius: var(--large-radius);
	background-color: var(--black-steel);
	width: 100%;
	box-sizing: border-box;
	min-width: 300px;
}

.browser-root-container::-webkit-scrollbar {
	width: 8px;
}

.browser-root-container::-webkit-scrollbar-thumb {
	border-radius: 10px;
	background-color: var(--dark-steel);
}

.browser-root-container::-webkit-scrollbar-track {
	border-radius: 10px;
}

.browser-root-container-hover-drop {
	background-color: #1e7fee6c;
	outline: 1px solid rgb(255, 255, 255);
	outline-offset: -1px;
}

.entity-list-scroll {
	z-index: 5;
	position: relative;
	display: flex;
	flex-direction: column;
	gap: .5rem;
	overflow: hidden;
	height: min-content;
	/* background-color: green; */
	width: 100%;
	box-sizing: border-box;
}

.dash-board-root {
	z-index: 1;
	position: relative;
	display: flex;
	flex-direction: column;
	color: var(--white);
	width: 100%;
	height: 100%;
	box-sizing: border-box;
	overflow: hidden;
	padding: .4rem;
	gap: .4rem;
	padding-top: .8rem;
}

.dash-board-header {
	position: relative;
	display: flex;
	width: 100%;
	align-items: center;
	height: max-content;
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

@keyframes loadingRotate {
  from {
      transform: rotate(0deg);
  }
  to {
      transform: rotate(360deg);
  }
}

.single-action-button{
  align-content: center;
  justify-content: center;
}

.loading-children-icon {
  width: 20px;
  height: 20px;
  overflow: hidden;
  padding: 0px;
  animation: loadingRotate .5s linear infinite;
}

.action-bar {
	position: relative;
	display: flex;
	align-items: center;
	gap: .4rem;
	width: max-content;
	height: max-content;
	padding: .2rem;
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
	padding: .2rem;
	width: max-content;
	height: max-content;
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
	padding-top: .5rem;
	box-sizing: border-box;
	overflow: hidden;
	min-height: 50px;
}

.navigator-mode-header {
	position: relative;
	display: flex;
	width: 100%;
	align-items: center;
	height: max-content;
	gap: 1rem;
	justify-content: flex-start;
	padding: .5rem;
	box-sizing: border-box;
	border-radius: var(--normal-radius);
	overflow: hidden;
}

.filter-options {
	display: flex;
	gap: .4rem;
	align-items: center;
	padding: .2rem 0;
	height: max-content;
	justify-content: flex-end;
	box-sizing: border-box;
	width: max-content;
	width: 100%;
	min-height: 35px;
	min-width: min-content;
	flex: 0 0 auto;
}

.menu-divider {
	width: 100%;
	margin-top: .5rem;
}

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
	max-width: 400px;
}

.desktop-search-bar::-ms-reveal {
	filter: invert(100%);
}

.filter-button{
	position: relative;
	overflow: visible;
}

.filter-button-indicator {
	overflow: hidden;
	width: 7px;
	height: 7px;
	border-radius: 5px;
	position: absolute;
	border-radius: 10px;
	outline: solid 1px var(--attention);
	background-color: var(--attention);
	top: 0;
	right: 1px;
}

.searchbar-container {
	display: flex;
	align-items: center;
	border: 0px;
	border-style: solid;
	outline: none;
	background-color: var(--midnight-steel);
	/* background-color: indigo; */
	border-radius: var(--normal-radius);
	width: 100%;
	/* flex: 1 1 30%; */
	max-width: 30%;
	/* min-width: 30%; */
	padding-right: .4rem;
	box-sizing: border-box;
}

.searchbar-container:hover {
	outline: var(--transparent-line);
	outline-offset: -1px;
}

.desktop-search-bar:focus .searchbar-container {
	background-color: red;
	outline: var(--solid-line);
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





