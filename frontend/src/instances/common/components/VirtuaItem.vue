<template>
  <div ref="virtuaItemRef" class="virtua-item" :style="{ '--depth': depth, '--lightness': `${65 + (depth * 5)}%` }">
    <div class="virtua-item-header drop-zone" :data-id="child.id" :data-other="JSON.stringify(child)" :id="child.id"
      :style="{ height: `${itemHeight}px` }" :class="{ 'drop-zone-hovered': isHovered }">
      <Collection ref="collectionItemRef" v-if="child.type == 'collection'" @toggleEditMode="toggleEditMode"
        v-right-click="openCollectionMenu" :hasChildren="hasChildren" :loadingChildren="loadingChildren" :isGhost="isGhost" @toggle="handleToggle"
        :collection="child" :index="index" :collectionChildren="collectionChildren" :loadingCollectionState="loadingCollectionState && child.type === 'collection'" />
      <Asset v-if="child.type == 'asset'" @refreshData="emit('refreshData')" @toggleEditMode="toggleEditMode"
        v-right-click="openAssetMenu" :asset="child" :index="index" :loadingAssetState="loadingAssetState && child.type === 'asset'" />
      <Collection ref="collectionItemRef" v-if="child.type == 'untracked_collection'" @toggleEditMode="toggleEditMode"
        v-right-click="openUntrackedItemMenu" :hasChildren="hasChildren" :loadingChildren="loadingChildren" :isUntracked="true" :collection="child"
        @toggle="handleToggle" :index="index" :collectionChildren="collectionChildren" :loadingCollectionState="loadingCollectionState && child.type === 'collection'"/>
      <Asset v-if="child.type == 'untracked_asset'" @toggleEditMode="toggleEditMode"
        v-right-click="openUntrackedItemMenu" :isUntracked="true" :asset="child" :index="index" />
    </div>
    <template v-if="isExpanded">
      <ListSkeleton v-if="loadingChildrenSkeleton" :itemHeight="commonStore.listItemHeight" :height="virtuaIndentHeight" :depth="depth + 1" />
      <div ref="virtuaChildrenRef" v-else-if="collectionChildren.length" class="virtua-item-children">
        <div class="indent-guide" :style="{ height: `${indentHeight}px` }"
          :class="{ 'indent-guide-selected': stage.markedItems.length === 1 && stage.firstSelectedItemId === child.id }">
        </div>
        <VirtuaList @refreshData="loadCollectionChildren" @updateChildren="handleUpdateChildren" @updateChildrenUntrackedItems="handleUpdateUntrackedItems" 
          :collectionId="child.id" :collectionType="child.type" :collectionPath="child.file_path" @shiftParents="handleToggle" :items="collectionChildren"
          :containerHeight="scrollStore.scrollRootHeight || 0" :depth="depth + 1" :parentOffset="totalOffset"
          :itemHeight="commonStore.listItemHeight" />
      </div>
    </template>
  </div>
</template>

<script setup>
// imports
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';

// components
import Asset from '@/instances/desktop/blocks/Asset.vue';
import Collection from '@/instances/desktop/blocks/Collection.vue';
import ListSkeleton from '@/instances/desktop/components/ListSkeleton.vue';
import VirtuaList from '@/instances/common/components/VirtuaList.vue';

// services
import { AssetService, CollectionService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useDndStore } from '@/stores/dnd';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useProjectStore } from '@/stores/projects';
import { useScrollStore } from '@/stores/scroll';
import { useStageStore } from '@/stores/stages';
import { useUntrackedItemStore } from '@/stores/untracked';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const dndStore = useDndStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const scrollStore = useScrollStore();
const stage = useStageStore();
const untrackedItemStore = useUntrackedItemStore();

// props
const props = defineProps({
  child: { type: Object, default: null },
  depth: { type: Number, default: 0 },
  getItemPosition: { type: Function, required: true },
  index: { type: Number, required: true },
  isGhost: { type: Boolean, default: false },
  itemHeight: { type: Number, default: 60 },
  offsetY: { type: Number, default: 0 },
  onHeightChange: { type: Function, default: () => { } },
  parentOffset: { type: Number, default: 0 },
  parentScrollContainer: { type: Object, default: null },
  totalHeight: { type: Number, default: 60 }
});

// emits
const emit = defineEmits(['refreshData']);

// refs
const collectionChildren = ref([]);
const collectionItemRef = ref(null);
const hasChildren = ref(false);
const indentPadding = ref(4);
const isEditing = ref(false);
const loadingAssetState = ref(false);
const loadingChildren = ref(false);
const loadingChildrenSkeleton = ref(true);
const loadingCollectionState = ref(false);

// timers
const loadingDelay = 250;
let loadingChildrenTimer = null;
const virtuaChildrenRef = ref(null);
const virtuaItemRef = ref(null);

// computed
// Checks if any filters are currently active.
const filtersActive = computed(() => {
  const assigneeFilters = commonStore.hasAssignees || commonStore.noAssignees;
  const collectionFilters = commonStore.collectionFilters.length > 0;
  const assetFilters = commonStore.assetFilters.length > 0;
  const resourceFilters = commonStore.resourceFilters.length > 0;
  const generalFilter = !(
    commonStore.showCollections &&
    commonStore.showAssets &&
    commonStore.showTasks &&
    commonStore.showResources &&
    commonStore.showChildCollections &&
    commonStore.showChildAssets &&
    commonStore.showDependencies &&
    !commonStore.onlyAssets &&
    !commonStore.onlyCollections
  );
  return assigneeFilters || collectionFilters || assetFilters || resourceFilters || generalFilter;
});

// Search and filters present collections as flat results, so nested expansion is disabled.
const isFilteredView = computed(() => !!commonStore.viewSearchQuery || filtersActive.value);

const shouldLoadAssetState = computed(() => !isFilteredView.value);

// Calculates the indent guide height for nested items.
const indentHeight = computed(() => {
  const itemHeight = stage.expandedCollections[props.child.id]["height"];
  collectionChildren.value;
  const height = virtuaChildrenRef.value?.getBoundingClientRect().height;
  return height - indentPadding.value;
});

// Checks if the current item has focus for selection.
const isItemInFocus = computed(() => {
  return stage.markedItems.length === 1 && stage.firstSelectedItemId === props.child.id && !dndStore.draggedItem;
});

// Checks if the item is expanded in the tree view.
const isExpanded = computed(() => {
  return props.child.id in stage.expandedCollections;
});

// Checks if the item is currently hovered for drag and drop.
const isHovered = computed(() => { return dndStore.targetItemId === props.child.id; });

// Calculates the total offset for nested scroll positioning.
const totalOffset = computed(() => {
  const itemPosition = props.getItemPosition(props.index);
  return props.offsetY + itemPosition;
});

// Calculates the height for the virtua indent skeleton.
const virtuaIndentHeight = computed(() => {
  const rootHeight = stage.expandedCollections[props.child.id]["height"];
  return rootHeight - props.itemHeight - indentPadding.value;
});

// methods
// Applies tracked task/resource visibility to assets that have already passed metadata filters.
const applyTrackedAssetVisibility = (assets = []) => {
  if (commonStore.onlyCollections) return [];
  if (!commonStore.showAssets) return [];
  return (assets || []).filter((asset) =>
    (asset.is_resource && commonStore.showResources) ||
    (!asset.is_resource && commonStore.showTasks)
  );
};

// Filters untracked assets by broad asset/untracked toggles plus search and extension filters.
const filterUntrackedAssets = (assets = []) => {
  if (commonStore.onlyCollections) return [];
  if (!commonStore.showAssets || !commonStore.showUntracked) return [];
  const viewSearchQuery = commonStore.viewSearchQuery?.toLowerCase() || "";
  const workspaceSearchQuery = commonStore.workspaceSearchQuery?.toLowerCase() || "";
  const selectedExtensions = commonStore.assetFilters
    .filter((filter) => filter.type === "extension")
    .map((filter) => filter.extension.toLowerCase());

  return (assets || []).filter((item) => {
    const extensionMatch = selectedExtensions.length === 0 || selectedExtensions.includes(item.extension?.toLowerCase());
    const name = item.name?.toLowerCase() || "";
    const filePath = item.file_path?.toLowerCase().replace(/\\/g, "/") || "";
    const itemPath = item.item_path?.toLowerCase().replace(/\\/g, "/") || "";
    const viewMatch = !viewSearchQuery || name.includes(viewSearchQuery) || filePath.includes(viewSearchQuery) || itemPath.includes(viewSearchQuery);
    const workspaceMatch = !workspaceSearchQuery || name.includes(workspaceSearchQuery) || filePath.includes(workspaceSearchQuery) || itemPath.includes(workspaceSearchQuery);
    return extensionMatch && viewMatch && workspaceMatch;
  });
};

// Filters untracked collections by broad collection/untracked toggles plus search.
const filterUntrackedCollections = (collections = []) => {
  if (!commonStore.showCollections || !commonStore.showUntracked || commonStore.onlyAssets) return [];
  const viewSearchQuery = commonStore.viewSearchQuery?.toLowerCase() || "";
  const workspaceSearchQuery = commonStore.workspaceSearchQuery?.toLowerCase() || "";

  return (collections || []).filter((item) => {
    const name = item.name?.toLowerCase() || "";
    const collectionPath = item.collection_path?.toLowerCase().replace(/\\/g, "/") || "";
    const itemPath = item.item_path?.toLowerCase().replace(/\\/g, "/") || "";
    const viewMatch = !viewSearchQuery || name.includes(viewSearchQuery) || collectionPath.includes(viewSearchQuery) || itemPath.includes(viewSearchQuery);
    const workspaceMatch = !workspaceSearchQuery || name.includes(workspaceSearchQuery) || collectionPath.includes(workspaceSearchQuery) || itemPath.includes(workspaceSearchQuery);
    return viewMatch && workspaceMatch;
  });
};

// Composes nested children using the same bucket visibility rules as the root browser.
const composeVisibleChildren = (children = {}) => {
  const collections = commonStore.showCollections && !commonStore.onlyAssets
    ? (children.collections || []).filter((item) => !item.is_trashed)
    : [];
  const assets = applyTrackedAssetVisibility(children.assets || []);
  const untrackedCollections = filterUntrackedCollections(children.untracked_collections || children.untracked_folders || []);
  const untrackedAssets = filterUntrackedAssets(children.untracked_assets || children.untracked_files || []);

  return [...collections, ...untrackedCollections, ...assets, ...untrackedAssets];
};

// Handles keyboard arrow key navigation for expanding/collapsing items.
const handleKeyArrowKeys = async (event) => {
  if (modals.activeModal) {
    return;
  }
  if (isItemInFocus.value && !isEditing.value) {
    if (event.key === 'ArrowLeft' || event.key === 'ArrowRight') {
      event.preventDefault();

      const item = props.child;
      const type = item.type;

      if (event.key === 'ArrowRight') {
        if (type.includes('collection') && collectionChildren.value.length && !(item.id in stage.expandedCollections)) {
          if (type === 'untracked_collection') {
            stage.expandCollection(item, true);
          } else {
            stage.expandCollection(item);
          }
          const firstChild = collectionChildren.value[0];
          const firstChildId = firstChild.id;
          handleToggle();
          stage.markedItems = [firstChildId];
          stage.firstSelectedItemId = firstChildId;
          stage.selectItem(firstChild, firstChild.type, true);
        }
      } else {
        let parent;
        const allItems = dndStore.allViewItems;

        if (type === 'collection') {
          parent = allItems.find((collection) => collection.id === item.parent_id);
        } else if (type === 'asset' || type === 'untracked_asset') {
          // Both tracked and untracked assets have collection_id pointing to parent
          parent = allItems.find((collection) => collection.id === item.collection_id);
        } else if (type === 'untracked_collection') {
          // For untracked collections, derive parent from path
          const currentPath = item.collection_path || item.item_path;
          const pathParts = currentPath.split('/').filter(part => part.trim() !== '');
          
          if (pathParts.length >= 2) {
            const parentPath = '/' + pathParts.slice(0, -1).join('/') + '/';
            // First try to find a tracked collection with this path
            parent = allItems.find((collection) => collection.collection_path === parentPath);
            
            // If not found, generate an untracked collection reference
            if (!parent) {
              const projectPath = projectStore.activeProject.working_directory.replace(/\\/g, '/').replace(/\/$/, '');
              const parentAbsPath = projectPath + parentPath.slice(0, -1);
              const parentId = utils.getMD5Hash(parentAbsPath);
              parent = allItems.find((collection) => collection.id === parentId);
            }
          }
        }

        if (parent) {
          const isParentUntracked = parent.type === 'untracked_collection';
          stage.expandCollection(parent, isParentUntracked);
          handleToggle();
          stage.markedItems = [parent.id];
          stage.firstSelectedItemId = parent.id;
          stage.selectItem(parent, parent.type, true);
        }
      }
    }
  }
};

// Handles toggle event and updates item height.
const handleToggle = async () => {
  await updateItemHeight();
};

// Handles updates to children from events.
const handleUpdateChildren = (eventData) => {
  if (!isExpanded.value) return;
  let loadFlags = true;

  if (Array.isArray(eventData)) {
    loadFlags = false;
    eventData.forEach(({ itemId, updates }) => {
      const itemIndex = collectionChildren.value.findIndex(item => item.id === itemId);
      if (itemIndex !== -1 && updates && Array.isArray(updates)) {
        updates.forEach(update => {
          if (update.property && update.value !== undefined) {
            collectionChildren.value[itemIndex][update.property] = update.value;
          }
        });
      }
    });
  } else {
    const { itemId, property, value, updates } = eventData;
    const itemIndex = collectionChildren.value.findIndex(item => item.id === itemId);
    if (itemIndex !== -1) {
      if (property && value !== undefined) {
        collectionChildren.value[itemIndex][property] = value;
      }
      if (updates && Array.isArray(updates)) {
        updates.forEach(update => {
          collectionChildren.value[itemIndex][update.property] = update.value;
        });
      }
    }
  }
  
  emitter.emit('get-project-data');
  if (loadFlags) {
    collectionStore.loadCollectionStateFlags();
  }
};

// Handles updates to untracked items from events.
// Rebuilds children in the original [tracked collections, untracked collections,
// tracked assets, untracked assets] order so VirtuaList keeps each row's identity
// stable across fs-driven refreshes.
const handleUpdateUntrackedItems = (untrackedItems) => {
  if (!untrackedItems) return;

  const trackedCollections = collectionChildren.value.filter(item => item.type === 'collection');
  const trackedAssets = collectionChildren.value.filter(item => item.type === 'asset');
  const untrackedCollections = untrackedItems.filter(item => item.type === 'untracked_collection');
  const untrackedAssets = untrackedItems.filter(item => item.type === 'untracked_asset');

  collectionChildren.value = composeVisibleChildren({
    collections: trackedCollections,
    assets: trackedAssets,
    untracked_collections: untrackedCollections,
    untracked_assets: untrackedAssets
  });
  hasChildren.value = collectionChildren.value.length > 0;

  if (!hasChildren.value && isExpanded.value) {
    const isUntracked = props.child.type === 'untracked_collection';
    stage.expandCollection(props.child, isUntracked);
  } else {
    updateItemHeight();
  }

  emitter.emit('get-project-data');
  collectionStore.loadCollectionStateFlags();
};

// Clears nested collection state when search/filter results are being shown.
const clearCollectionChildren = () => {
  clearTimeout(loadingChildrenTimer);
  collectionChildren.value = [];
  hasChildren.value = false;
  loadingChildren.value = false;
  loadingChildrenSkeleton.value = false;

  if (isExpanded.value) {
    const isUntracked = props.child.type === 'untracked_collection';
    stage.expandCollection(props.child, isUntracked);
  }
};

const emptyCollectionStateFlags = () => ({
  has_untracked: false,
  has_modified: false,
  has_outdated: false,
  has_fetchable: false
});

// Loads the file status state for an asset.
const loadAssetState = async () => {
  const asset = props.child;
  
  if (asset.type !== 'asset' || asset.is_link) return;
  if (!shouldLoadAssetState.value) {
    loadingAssetState.value = false;
    return;
  }

  const loadingTimer = setTimeout(() => {
    loadingAssetState.value = true;
  }, loadingDelay);
  
  try {
    const fileStatus = await AssetService.GetAssetState(
      projectStore.activeProject.uri,
      asset.id
    );

    if (!shouldLoadAssetState.value) return;

    props.child.file_status = fileStatus;
  } catch (error) {
    console.error(`Error loading asset state for ${asset.id}:`, error);
    asset.file_status = 'fetchable';
  } finally {
    clearTimeout(loadingTimer);
    loadingAssetState.value = false;
  }
};

// Loads the state flags for a collection.
const loadCollectionState = async () => {
  const collection = props.child;
  
  if (collection.type !== 'collection') return;

  if (isFilteredView.value) {
    loadingCollectionState.value = false;
    props.child.collectionStateFlags = emptyCollectionStateFlags();
    return;
  }

  const loadingTimer = setTimeout(() => {
    loadingCollectionState.value = true;
  }, loadingDelay);

  try {
    const flags = await CollectionService.GetCollectionStateFlags(
      projectStore.activeProject.uri,
      collection.id,
      projectStore.activeProject.working_directory,
      projectStore.activeProject.ignore_list
    );

    if (isFilteredView.value) {
      props.child.collectionStateFlags = emptyCollectionStateFlags();
      return;
    }

    props.child.collectionStateFlags = flags;
  } catch (error) {
    console.error(`Error loading collection state for ${collection.id}:`, error);
    props.child.collectionStateFlags = emptyCollectionStateFlags();
  } finally {
    clearTimeout(loadingTimer);
    loadingCollectionState.value = false;
  }
};

// Loads children for an collection or untracked collection.
const loadCollectionChildren = async () => {
  if (isFilteredView.value) {
    clearCollectionChildren();
    return;
  }

  if (stage.operationActive) {
    loadingChildrenSkeleton.value = true;
    return;
  }

  loadingChildrenSkeleton.value = true;
  clearTimeout(loadingChildrenTimer);
  loadingChildrenTimer = setTimeout(() => {
    loadingChildren.value = true;
  }, loadingDelay);

  if (props.child.type == "collection" || props.child.type == 'untracked_collection') {
    let isUntracked = props.child.type == 'untracked_collection';
    let project = projectStore.activeProject;
    let children = await CollectionService.GetCollectionChildren(project.uri, props.child.id, project.working_directory, props.child.file_path, project.ignore_list, isUntracked);
    if (isFilteredView.value) {
      clearCollectionChildren();
      return;
    }

    await assetStore.processAssetsIconsAndPreviews(children.assets);
    await assetStore.processUntrackedAssetsIcons(children.untracked_assets);
    if (isFilteredView.value) {
      clearCollectionChildren();
      return;
    }

    let childrenCollections = filtersActive.value ? await collectionStore.filterCollections(children.collections) : children.collections;
    let childrenAssets = filtersActive.value ? await assetStore.filterAssets(children.assets) : children.assets;
    if (isFilteredView.value) {
      clearCollectionChildren();
      return;
    }

    collectionChildren.value = composeVisibleChildren({
      collections: childrenCollections,
      assets: childrenAssets,
      untracked_collections: children.untracked_collections,
      untracked_assets: children.untracked_assets
    });
    hasChildren.value = collectionChildren.value.length > 0;
    
    if (!hasChildren.value && props.child.id in stage.expandedCollections) {
      stage.expandCollection(props.child, isUntracked);
    }
  }
  clearTimeout(loadingChildrenTimer);
  loadingChildren.value = false;
  loadingChildrenSkeleton.value = false;
};

// Focuses a single item so right-click menu actions (e.g. rename) work.
// Preserves an existing multi-selection that already contains the item.
const focusSingleItem = (item) => {
  if (stage.markedItems.includes(item.id)) return;
  stage.markedItems = [item.id];
  stage.firstSelectedItemId = item.id;
  stage.lastSelectedItemId = "";
  stage.selectedItems = [item];
};

// Opens the multi-selection menu when right-clicking an item that is part of
// an existing multi-selection. Returns true when the selection menu was shown.
const tryOpenSelectionMenu = (event, id) => {
  if (stage.markedItems.length > 1 && stage.markedItems.includes(id)) {
    menu.showContextMenu(event, 'selectionMenu', true);
    return true;
  }
  return false;
};

// Opens the asset context menu.
const openAssetMenu = (event) => {
  const id = props.child.id;
  if (tryOpenSelectionMenu(event, id)) return;
  const asset = props.child;
  assetStore.selectAsset(asset);
  stage.markedAssets = [id];
  focusSingleItem(asset);
  menu.showContextMenu(event, 'assetMenu', true);
};

// Opens the collection context menu.
const openCollectionMenu = (event) => {
  const id = props.child.id;
  if (tryOpenSelectionMenu(event, id)) return;
  const collection = props.child;
  stage.markedCollections = [id];
  collectionStore.selectCollection(collection);
  focusSingleItem(collection);
  menu.showContextMenu(event, 'collectionMenu', true);
};

// Opens the untracked item context menu.
const openUntrackedItemMenu = (event) => {
  const id = props.child.id;
  if (tryOpenSelectionMenu(event, id)) return;
  stage.markedCollections = [id];
  untrackedItemStore.selectUntrackedItem(props.child);
  focusSingleItem(props.child);
  menu.showContextMenu(event, 'untrackedItemMenu', true);
};

// Toggles the edit mode for renaming.
const toggleEditMode = (value) => {
  isEditing.value = value;
};

// Updates the item height after DOM changes.
const updateItemHeight = async () => {
  if (!virtuaItemRef.value) return;
  await nextTick();
  const height = virtuaItemRef.value?.offsetHeight;
  props.onHeightChange(props.index, height);
};

// watchers
watch(() => stage.operationActive, (newValue, oldValue) => {
  if (oldValue && !newValue && loadingChildrenSkeleton.value) {
    loadCollectionChildren();
  }
});

watch(isFilteredView, async (filtered) => {
  if (props.child.type === 'asset') {
    if (filtered) {
      loadingAssetState.value = false;
      return;
    }
    await loadAssetState();
    return;
  }

  if (props.child.type !== 'collection' && props.child.type !== 'untracked_collection') return;

  if (filtered) {
    clearCollectionChildren();
    if (props.child.type === 'collection') {
      props.child.collectionStateFlags = emptyCollectionStateFlags();
      loadingCollectionState.value = false;
    }
    return;
  }

  await loadCollectionChildren();
  if (props.child.type === 'collection') {
    await loadCollectionState();
  }
});

// lifecycle hooks
onMounted(async () => {
  if (props.child.type === 'collection' || props.child.type === 'untracked_collection') {
    await loadCollectionChildren();
  }
  
  if (props.child.type === 'collection') {
    await loadCollectionState();
  }
  
  if (props.child.type === 'asset') {
    await loadAssetState();
  }
  
  window.addEventListener('keydown', handleKeyArrowKeys);
  emitter.on('update-children', handleUpdateChildren);
});

onBeforeUnmount(() => {
  clearTimeout(loadingChildrenTimer);
  window.removeEventListener('keydown', handleKeyArrowKeys);
  emitter.off('update-children', handleUpdateChildren);
});
</script>

<style scoped>
.indent-guide {
  border-left: var(--transparent-line);
  box-sizing: border-box;
  height: 300px;
  left: 8px;
  position: absolute;
  width: 100%;
}

.indent-guide-selected {
  border-left: var(--medium-transparent-line);
}

.virtua-item {
  box-sizing: border-box;
  cursor: pointer;
  margin: 0 auto;
  overflow: hidden;
  transition: all 0.3s ease;
}

.virtua-item-children {
  box-sizing: border-box;
  overflow: hidden;
  padding-left: 15px;
  width: 100%;
}

.virtua-item-header {
  box-sizing: border-box;
  overflow: hidden;
}

.virtua-item-header:hover + .virtua-item-children > .indent-guide {
  border-left: var(--medium-transparent-line);
}
</style>
