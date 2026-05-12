<template>
	<div ref="breadcrumbRoot" class="breadcrumb-root">
		<template v-if="commonStore.viewMode === 'kanban'">
			<div class="kanban-indicator">
				<ActionButton v-if="isDefaultWorkspace" :icon="getAppIcon('arrow-left')" v-tooltip="$t('components.breadcrumbs.exitKanban')" :buttonFunction="exitKanbanView" />
				<ActionButton v-else :icon="getAppIcon('kanban')" :allowDeactivate="true" />
				<span class="kanban-indicator-label">{{ $t('settings.kanban') }}</span>
			</div>
			<div class="kanban-path-separator"></div>
		</template>

		<template v-else>
			<ActionButton v-if="commonStore.navigatorMode" :icon="getAppIcon(commonStore.navigatorMode ? 'home' : 'forward-slash')" v-tooltip="$t('components.breadcrumbs.home')" :buttonFunction="goHome" />
			<ActionButton :icon="getAppIcon('refresh')" v-tooltip="$t('components.breadcrumbs.refresh')" :buttonFunction="refresh" />

			<ActionButton v-if="commonStore.navigatorMode" :icon="getAppIcon('arrow-back-ramp')"
				:allowDeactivate="true" v-tooltip="$t('components.breadcrumbs.upALevel')" :buttonFunction="goUpALevel" />

			<ActionButton v-if="showProjectChip" :icon="getAppIcon('forward-slash')" v-tooltip="commonStore.navigatorMode ? 'Home' : ''"
				:label="projectStore.activeProject?.name" :buttonFunction="goHome" />

			<div v-if="showRootFilteredSeparator" class="kanban-path-separator"></div>
		</template>

		<div ref="breadcrumbWrapper" class="breadcrumb-wrapper">
			<div ref="breadcrumbContainer" class="breadcrumb-container">
				<span v-if="showFilteredLabel" class="kanban-filter-text">
					{{ $t('components.breadcrumbs.filterResults') }}
				</span>
				<nav v-else-if="path" ref="breadcrumbContent" class="nav">
					<ActionButton v-if="showEllipsis" :icon="getAppIcon('dots')" :allowDeactivate="true" @click="toggleOverflowList" />
					<div v-for="(segment, index) in visibleSegments" :key="`${segment}-${index}`" class="breadcrumb-segment">
						<ActionButton v-if="path !== 'Home'" :icon="getAppIcon('forward-slash')" :allowDeactivate="true"
							:label="segment.split('/').pop()" @click="goToCollection(segment)" />
					</div>
				</nav>
			</div>
		</div>

		<template v-if="commonStore.viewMode !== 'kanban'">
			<ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('copy')" :showLabel="false" :fullWidth="false"
				v-tooltip="$t('components.breadcrumbs.copyPath')" @click="copyDirectoryPath" />

			<ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('folder-arrow-up-right')" :showLabel="false" :fullWidth="false"
				v-tooltip="$t('components.breadcrumbs.showInExplorer')" @click="revealInExplorer" />
		</template>
	</div>

	<Teleport to="#app">
		<div v-if="displayOverflowItems && commonStore.viewMode !== 'kanban'" :style="{ top: listItemsAnchor + 'px', left: listItemsLeft + 'px' }" class="breadcrumb-list-container">
			<div class="breadcrumb-instance-container">
				<div v-for="(overflowPath, index) in overflowPaths" :key="index" class="breadcrumb-instance" @click="goToCollection(overflowPath)">
					<div class="breadcrumb-instance-meta">
						<div>{{ overflowPath.split('/').pop() }}</div>
					</div>
				</div>
			</div>
		</div>
	</Teleport>
</template>

<script setup>
// imports
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Clipboard } from '@wailsio/runtime';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// services
import { CollectionService, FSService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const { t } = useI18n();

// refs
const breadcrumbContainer = ref(null);
const breadcrumbContent = ref(null);
const breadcrumbRoot = ref(null);
const displayOverflowItems = ref(false);
const overflowPaths = ref([]);
const resizeObserver = ref(null);
const showEllipsis = ref(false);
const visibleSegments = ref([]);

// computed properties
const listItemsAnchor = computed(() => {
	const rootHeight = breadcrumbRoot.value?.getBoundingClientRect().height || 0;
	const rootTop = breadcrumbRoot.value?.getBoundingClientRect().top || 0;
	return rootTop + rootHeight;
});

const listItemsLeft = computed(() => breadcrumbContent.value?.getBoundingClientRect().left || 0);

const navigatedCollection = computed(() => collectionStore.navigatedCollection);

const isDefaultWorkspace = computed(() => commonStore.activeWorkspace === 'Default');

// True when any filter, search or visibility toggle is active (ignores view-mode changes).
const hasActiveFilters = computed(() => {
	return !!(
		commonStore.viewSearchQuery
		|| commonStore.assetFilters.length
		|| commonStore.collectionFilters.length
		|| commonStore.resourceFilters.length
		|| commonStore.hasAssignees
		|| commonStore.noAssignees
		|| !commonStore.showCollections
		|| !commonStore.showAssets
		|| commonStore.onlyAssets
		|| !commonStore.showResources
		|| !commonStore.showChildCollections
		|| !commonStore.showChildAssets
		|| !commonStore.showChildResources
		|| !commonStore.showDependencies
		|| commonStore.useDeep
	);
});

// Show the project-name chip only on Default workspace at the root with no active filters/search.
const showProjectChip = computed(() => {
	return !commonStore.navigatorMode
		&& commonStore.viewMode !== 'kanban'
		&& isDefaultWorkspace.value
		&& !hasActiveFilters.value;
});

const showFilteredLabel = computed(() => {
	if (commonStore.navigatorMode) return false;
	if (commonStore.viewMode === 'kanban') return true;
	return !showProjectChip.value;
});

const showRootFilteredSeparator = computed(() => {
	return commonStore.viewMode !== 'kanban' && showFilteredLabel.value;
});

const path = computed(() => {
	if (commonStore.navigatorMode) {
		return navigatedCollection.value?.type === 'collection'
			? navigatedCollection.value?.collection_path
			: navigatedCollection.value?.item_path;
	}
	return 'Home';
});

const segments = computed(() => {
	const pathParts = path.value.split('/').filter(segment => segment.trim() !== '');
	return pathParts.map((_, index) => pathParts.slice(0, index + 1).join('/'));
});

// methods

// Checks if the breadcrumb bar is overflowing and adjusts visible segments.
const checkOverflow = async () => {
	if (!path.value) return;
	await nextTick();
	if (!breadcrumbContainer.value || segments.value.length === 0) return;

	const nav = breadcrumbContainer.value.querySelector('.nav');
	const container = breadcrumbContainer.value;
	if (!nav) return;

	const testFit = async (numSegments) => {
		const testSegments = segments.value.slice(-numSegments);
		const hiddenSegments = segments.value.slice(0, segments.value.length - numSegments);
		const needsEllipsis = numSegments < segments.value.length;

		overflowPaths.value = hiddenSegments;
		visibleSegments.value = testSegments;
		showEllipsis.value = needsEllipsis;

		await nextTick();
		return nav.scrollWidth <= container.clientWidth;
	};

	if (await testFit(segments.value.length)) return;

	let left = 1;
	let right = segments.value.length - 1;
	let bestFit = 1;

	while (left <= right) {
		const mid = Math.floor((left + right) / 2);
		if (await testFit(mid)) {
			bestFit = mid;
			left = mid + 1;
		} else {
			right = mid - 1;
		}
	}

	await testFit(bestFit);
};

// Clears all item selections and resets selection state.
const clearAllSelections = () => {
	stage.markedItems = [];
	stage.selectedItems = [];
	stage.firstSelectedItemId = '';
	stage.lastSelectedItemId = '';
	stage.selectedItem = null;
	assetStore.selectedAsset = null;
	collectionStore.selectedCollection = null;
	projectStore.selectedUntrackedItem = null;
};

// Copies the current directory path to clipboard.
const copyDirectoryPath = async () => {
	const project = projectStore.getActiveProject;
	let explorerPath;

	if (!commonStore.navigatorMode) {
		explorerPath = project.working_directory.replace(/\\/g, '/');
		await FSService.MakeDirs(explorerPath);
	} else {
		const navPath = collectionStore.navigatedCollection?.type === 'collection'
			? collectionStore.navigatedCollection.collection_path
			: collectionStore.navigatedCollection.item_path;
		explorerPath = `${project.working_directory}${navPath}`.replace(/\\/g, '/');
		await FSService.MakeDirs(explorerPath);
	}

	await Clipboard.SetText(explorerPath);
	notificationStore.addNotification(t('components.breadcrumbs.pathCopiedToClipboard'), '', 'success');
};

// Generates an untracked collection object from a given path.
const generateUntrackedCollectionFromPath = (targetPath, projectPath) => {
	const pathParts = targetPath.split('/').filter(part => part.trim() !== '');
	if (pathParts.length === 0) return null;

	const collectionName = pathParts[pathParts.length - 1];
	const normalizedProjectPath = projectPath.replace(/\\/g, '/').replace(/\/$/, '');
	const absPath = normalizedProjectPath + targetPath.slice(0, -1);

	let parentId = '';
	if (pathParts.length >= 2) {
		const parentPath = '/' + pathParts.slice(0, -1).join('/') + '/';
		const parentAbsPath = normalizedProjectPath + parentPath.slice(0, -1);
		parentId = utils.getMD5Hash(parentAbsPath);
	}

	return {
		id: utils.getMD5Hash(absPath),
		name: collectionName,
		collection_path: targetPath,
		item_path: targetPath,
		file_path: absPath,
		parent_id: parentId,
		type: 'untracked_collection'
	};
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Returns the parent collection for an untracked collection.
const getUntrackedCollectionParent = () => {
	const currentPath = path.value;
	const projectPath = projectStore.activeProject.working_directory;
	const pathParts = currentPath.split('/').filter(part => part.trim() !== '');

	if (pathParts.length < 2) {
		commonStore.navigatorMode = false;
		return null;
	}

	const parentCollectionPath = '/' + pathParts.slice(0, -1).join('/') + '/';
	return generateUntrackedCollectionFromPath(parentCollectionPath, projectPath);
};

// Navigates to a specific collection by path.
const goToCollection = async (selectedPath) => {
	displayOverflowItems.value = false;
	const clickedPath = `/${selectedPath}/`;

	if (clickedPath === path.value) return;

	const navigatedCollectionType = collectionStore.navigatedCollection?.type;
	const projectPath = projectStore.activeProject.working_directory;

	let targetCollection = null;
	if (navigatedCollectionType === 'collection') {
		targetCollection = await CollectionService.GetCollectionByPath(projectStore.activeProject.uri, clickedPath);
	} else {
		targetCollection = generateUntrackedCollectionFromPath(clickedPath, projectPath);
	}

	if (targetCollection) {
		collectionStore.navigatedCollection = targetCollection;
		collectionStore.selectedCollection = targetCollection;
	} else {
		notificationStore.addNotification(t('components.breadcrumbs.navigationFailed'), t('components.breadcrumbs.couldNotFindSelectedPath'), 'error');
	}
};

// Navigates to the home/root view.
const goHome = () => {
	commonStore.navigatorMode = false;
	collectionStore.navigatedCollection = null;
	clearAllSelections();
};

// Exits kanban view and restores the previously active view mode.
const exitKanbanView = () => {
	commonStore.restorePreviousView();
	emitter.emit('refresh-browser');
};

// Navigates up one level in the breadcrumb hierarchy.
const goUpALevel = async () => {
	const collection = collectionStore.navigatedCollection;
	const collectionType = collection.type;
	let parentCollectionId = collection.parent_id;

	if (!parentCollectionId) {
		commonStore.navigatorMode = false;
		collectionStore.navigatedCollection = null;
		clearAllSelections();
		return;
	}

	let parentCollection;
	if (collectionType === 'untracked_collection') {
		parentCollection = getUntrackedCollectionParent();
		if (parentCollection?.collection_path) {
			try {
				const trackedParent = await CollectionService.GetCollectionByPath(projectStore.activeProject.uri, parentCollection.collection_path);
				if (trackedParent) parentCollection = trackedParent;
			} catch (error) {
				// console.log('Parent collection not found in DB, using untracked collection');
			}
		}
	} else {
		parentCollection = await CollectionService.GetCollectionByID(projectStore.activeProject.uri, parentCollectionId);
	}

	if (parentCollection) {
		collectionStore.navigatedCollection = parentCollection;
		stage.lastSelectedItemId = '';
		stage.firstSelectedItemId = parentCollection.id;
		stage.markedItems = [parentCollection.id];
		stage.selectedItems = [parentCollection];
		stage.selectItem(parentCollection, parentCollection.type, true);
	} else {
		commonStore.navigatorMode = false;
	}
};

// Handles clicks outside to close overflow menu.
const handleClickOutside = () => { if (displayOverflowItems.value) displayOverflowItems.value = false; };

// Emits refresh event to reload the browser view.
const refresh = () => emitter.emit('refresh-browser');

// Opens the current directory in the system file explorer.
const revealInExplorer = async () => {
	const project = projectStore.getActiveProject;

	if (!commonStore.navigatorMode) {
		await FSService.MakeDirs(project.working_directory);
		FSService.RevealInExplorer(project.working_directory);
	} else {
		const navPath = collectionStore.navigatedCollection?.type === 'collection'
			? collectionStore.navigatedCollection.collection_path
			: collectionStore.navigatedCollection.item_path;
		const trimmedPath = navPath.endsWith('/') ? navPath.slice(0, -1) : navPath;
		let explorerPath = `${project.working_directory}${trimmedPath}`.replace(/\\/g, '/');
		await FSService.MakeDirs(explorerPath);
		FSService.RevealInExplorer(explorerPath);
	}
};

// Toggles the overflow items dropdown visibility.
const toggleOverflowList = () => { displayOverflowItems.value = !displayOverflowItems.value; };

// watchers

watch(path, () => {
	if (path.value === 'Home') showEllipsis.value = false;
	checkOverflow();
});

watch(() => projectStore.activeProject?.uri, () => {
	commonStore.navigatorMode = false;
	collectionStore.navigatedCollection = null;
	collectionStore.selectedCollection = null;
});

// lifecycle hooks

onMounted(() => {
	checkOverflow();
	if (breadcrumbRoot.value) {
		resizeObserver.value = new ResizeObserver(() => checkOverflow());
		resizeObserver.value.observe(breadcrumbRoot.value);
	}
	document.addEventListener('click', handleClickOutside);
});

onBeforeUnmount(() => {
	if (resizeObserver.value) resizeObserver.value.disconnect();
	document.removeEventListener('click', handleClickOutside);
});
</script>

<style scoped>
.breadcrumb-root {
	display: flex;
	align-items: center;
	font-size: 0.875rem;
	font-weight: 500;
	background-color: var(--black-steel);
	border-radius: var(--large-radius);
	overflow: hidden;
	max-width: 70%;
	min-width: 65%;
	padding: .2rem;
	width: 100%;
	box-sizing: border-box;
}

.kanban-indicator {
	display: flex;
	align-items: center;
	gap: .5rem;
	flex-shrink: 0;
}

.kanban-indicator-label {
	color: var(--white);
	font-size: 0.85rem;
	font-weight: 500;
	white-space: nowrap;
}

.kanban-path-separator {
	width: 1.5px;
	height: 18px;
	background-color: var(--light-steel);
	margin: 0 1rem;
	flex-shrink: 0;
}

.kanban-static-icon {
	width: 18px;
	height: 18px;
	object-fit: contain;
	opacity: 0.8;
}

.kanban-filter-text {
	color: var(--white);
	opacity: 0.5;
	font-size: 0.875rem;
	font-weight: 300;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.kanban-path-text {
	color: var(--white);
	opacity: 0.5;
	font-size: 0.875rem;
	font-weight: 300;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
	flex: 1;
	min-width: 0;
}

.breadcrumb-root::-webkit-scrollbar {
	display: none;
}

.breadcrumb-wrapper {
	display: flex;
	align-items: center;
	width: 100%;
	overflow: hidden;
}

.breadcrumb-container {
	display: flex;
	align-items: center;
	width: min-content;
	overflow-x: auto;
	scrollbar-width: none;
	border-radius: var(--small-radius);
	justify-content: flex-end;
}

.nav {
	display: flex;
	align-items: center;
	font-size: 14px;
	white-space: nowrap;
}

.breadcrumb-segment {
	display: flex;
	align-items: center;
	width: min-content;
}

.breadcrumb-list-container {
	position: absolute;
	z-index: 10000;
	min-width: 160px;
	width: max-content;
	height: max-content;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 1rem;
	border-radius: var(--small-radius);
	background-color: var(--black);
	color: var(--white);
	outline: var(--transparent-line);
	outline-offset: -1px;
	overflow: hidden;
	box-sizing: border-box;
}

.breadcrumb-instance-container {
	width: 100%;
	height: max-content;
	gap: .5rem;
	display: flex;
	padding: .3rem;
	box-sizing: border-box;
	overflow: hidden;
	flex-direction: column;
}

.breadcrumb-instance {
	overflow: hidden;
	background-color: transparent;
	text-align: center;
	font-size: 14px;
	line-height: 14px;
	color: var(--white);
	position: relative;
	border-radius: var(--small-radius);
	box-sizing: border-box;
	cursor: pointer;
	display: flex;
	gap: 5px;
	align-items: center;
	padding: .1rem;
	height: max-content;
	width: 100%;
	min-width: max-content;
	min-height: max-content;
	transition: all 0.3s ease;
	justify-content: space-between;
}

.breadcrumb-instance:hover {
	background-color: #ffffff15;
}

.breadcrumb-instance:active {
	background-color: #00000013;
}

.breadcrumb-instance-meta {
	display: flex;
	align-items: center;
	box-sizing: border-box;
	width: 100%;
	height: 30px;
	padding: .2rem .5rem;
	gap: 10px;
}
</style>





