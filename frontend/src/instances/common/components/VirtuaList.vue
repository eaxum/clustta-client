<template>
  <div class="virtua-scroll-viewport" ref="scrollContainerRef" :style="{ height: `${totalHeight}px` }"
    :data-visibility="containerVisibility">
    <div class="virtua-scroll-conveyor" :style="{ transform: `translateY(${offsetY}px)` }">
      <VirtuaItem v-for="child in visibleChildren" @refreshData="emit('refreshData')" :child="items[child.index]" :key="child.index" :index="child.index"
        :itemHeight="itemHeight" :isExpanded="child.isExpanded" :onHeightChange="onHeightChange"
        :depth="depth" :getItemPosition="getItemPosition" :parentOffset="props.parentOffset" :offsetY="offsetY"
        :totalHeight="totalHeight" @mousedown="onMouseDown($event, child, index)"
        @mouseup="onMouseUp($event, child, index)" :ref="el => handleRef(child.id, el?.$el || el)" />
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, inject, nextTick, onBeforeUnmount, onMounted, onUpdated, ref, watch, watchEffect } from 'vue';
import emitter from '@/lib/mitt';
import { useFsWatch } from '@/composables/useFsWatch';

// components
import VirtuaItem from '@/instances/common/components/VirtuaItem.vue';

// services
import { CollectionService } from "@/services";

// stores
import { useAssetStore } from '@/stores/assets';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useDndStore } from '@/stores/dnd';
import { useMenu } from '@/stores/menu';
import { useProjectStore } from '@/stores/projects';
import { useScrollStore } from '@/stores/scroll';
import { useStageStore } from '@/stores/stages';

const assetStore = useAssetStore();
const dndStore = useDndStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const scrollStore = useScrollStore();
const stage = useStageStore();

// props
const props = defineProps({
  collectionType: { type: String, default: ''},
  collectionId: { type: String, default: '' },
  collectionPath: { type: String, default: '' },
  containerHeight: { type: Number, required: true },
  depth: { type: Number, default: 0 },
  isRoot: { type: Boolean, default: false },
  itemHeight: { type: Number, default: 60 },
  items: { type: Array, default: [] },
  parentOffset: { type: Number, default: 0 },
  renderAhead: { type: Number, default: 10 }
});

// emits
const emit = defineEmits(['shift-parents', 'refreshData', 'update-children', 'update-children-untracked-items']);

// refs
const childPositions = ref([0]);
const currentState = ref(null);
const dragTimer = ref(null);
const intersectionObserver = ref(null);
const intersectionRatio = ref(0);
const intersectionRect = ref(null);
const isVisible = ref(false);
const pixelsAboveRoot = ref(0);
const scrollContainerRef = ref(null);

let refreshToken = 0;

// injects
const rootScrollContainer = inject('rootScrollContainer', null);

// computed
// Returns all item IDs from the dnd store.
const allItemsIds = computed(() => {
  const allItemsIds = dndStore.allElements.map((item) => item.id);
  return allItemsIds;
});

// Determines container visibility state based on intersection.
const containerVisibility = computed(() => {
  if (!isVisible.value) return 'hidden';
  if (intersectionRatio.value === 1) return 'fully-visible';
  return 'partially-visible';
});

// Calculates end node index with render ahead buffer.
const endNode = computed(() =>
  Math.min(itemCount.value - 1, lastVisibleNode.value + props.renderAhead)
);

// Finds the first visible node index using binary search.
const firstVisibleNode = computed(() => {
  return findStartNode(relativeScrollTop.value, childPositions.value, itemCount.value);
});

// Returns the total number of items in the list.
const itemCount = computed(() => {
  return props.items.length;
});

// Finds the last visible node index within container height.
const lastVisibleNode = computed(() =>
  findEndNode(childPositions.value, firstVisibleNode.value, itemCount.value, props.containerHeight)
);

// Calculates vertical offset for scroll positioning.
const offsetY = computed(() => {
  return childPositions.value[startNode.value];
});

// Calculates relative scroll position based on root or nested context.
const relativeScrollTop = computed(() => {
  if (props.isRoot) return scrollStore.scrollTop;
  return pixelsAboveRoot.value;
});

// Gets the currently selected item ID.
const selectedId = computed(() => {
  return stage.firstSelectedItemId;
});

// Calculates start node index with render ahead buffer.
const startNode = computed(() =>
  Math.max(0, firstVisibleNode.value - props.renderAhead)
);

// Calculates total height of all items.
const totalHeight = computed(() => {
  if (itemCount.value === 0) return 0;
  return childPositions.value[itemCount.value - 1] + getChildHeight(itemCount.value - 1);
});

// Gets visible children based on scroll position.
const visibleChildren = computed(() =>
  Array(visibleNodeCount.value)
    .fill(null)
    .map((_, index) => {
      const actualIndex = index + startNode.value;
      const item = props.items[actualIndex];

      return item ? {
        ...item,
        index: actualIndex,
        isExpanded: false
      } : null;
    })
    .filter(Boolean)
);

// Calculates number of visible nodes.
const visibleNodeCount = computed(() =>
  endNode.value - startNode.value + 1
);

// methods
// Calculates cumulative positions of all child items.
const calculateChildPositions = () => {
  emit('shift-parents');
  const positions = [0];
  for (let i = 1; i < itemCount.value; i++) {
    positions.push(positions[i - 1] + getChildHeight(i - 1));
  }
  childPositions.value = positions;
};

// Starts drag timer for delayed drag operation.
const dragItem = (event, id) => {
  if (stage.operationActive) return;
  dragTimer.value = setTimeout(() => {
    onDragStart(event, id);
  }, dndStore.dragDelay);
};

// Emits updates to item data across components.
const emitItemUpdates = (updates) => {
  if (Array.isArray(updates) && updates.length > 0) {
    if (props.isRoot) {
      emitter.emit('update-root-data', updates);
    } else {
      emit('update-children', updates);
    }
  }
};

// Emits untracked items updates.
const emitUntrackedUpdates = (allUntrackedItems) => {
  if (props.isRoot) {
    emitter.emit('update-untracked-items', allUntrackedItems);
  } else {
    emit('update-children-untracked-items', allUntrackedItems);
  }
};

// Finds end node based on container height.
const findEndNode = (nodePositions, startNode, itemCount, height) => {
  let endNode;
  for (endNode = startNode; endNode < itemCount; endNode++) {
    if (nodePositions[endNode] > nodePositions[startNode] + height) {
      return endNode;
    }
  }
  return endNode;
};

// Binary search to find start node based on scroll position.
const findStartNode = (scrollTop, nodePositions, itemCount) => {
  let startRange = 0;
  let endRange = itemCount - 1;
  while (endRange !== startRange) {
    const middle = Math.floor((endRange - startRange) / 2 + startRange);
    if (nodePositions[middle] <= scrollTop && nodePositions[middle + 1] > scrollTop) {
      return middle;
    }
    if (middle === startRange) {
      return endRange;
    } else {
      if (nodePositions[middle] <= scrollTop) {
        startRange = middle;
      } else {
        endRange = middle;
      }
    }
  }
  return itemCount;
};

// Gets height of child item considering expanded state.
const getChildHeight = (index) => {
  const item = props.items[index];
  return item && item.id in stage.expandedCollections ?
    stage.expandedCollections[item.id]["height"] || props.itemHeight : props.itemHeight;
};

// Gets position of item by index.
const getItemPosition = (index) => {
  return childPositions.value[index] || 0;
};

// Handles keyboard navigation with arrow keys.
const handleKeyDown = (event) => {
  if (modals.activeModal) {
    return;
  }
  if (event.key === 'ArrowUp' || event.key === 'ArrowDown') {
    event.preventDefault();
    const currentIndex = allItemsIds.value.indexOf(selectedId.value);
    if (currentIndex !== undefined) {
      let newIndex;
      if (event.key === 'ArrowUp') {
        newIndex = currentIndex > 0 ? currentIndex - 1 : currentIndex;
      } else {
        newIndex = currentIndex < allItemsIds.value.length - 1 ? currentIndex + 1 : currentIndex;
      }
      selectItem(newIndex);
    }
  }
};

// Adds or removes element reference to dnd store.
const handleRef = async (id, el) => {
  if (!el) {
    dndStore.removeRef(id);
    dndStore.removeVisibleItemsRef(id);
    return;
  }

  await nextTick();

  const domElement = el instanceof HTMLElement ? el : el.$el;

  if (domElement) {
    dndStore.addRef(id, domElement);
  }
};

// Initializes drag operation through dnd store.
const onDragStart = (e, id) => {
  stage.firstSelectedItemId = '';

  if (!id) {
    return;
  }
  dndStore.onDragStart(e, id);
};

// Handles height change of expanded items.
const onHeightChange = (index, height) => {
  const item = props.items[index];
  if (item && item.id && item.id in stage.expandedCollections) {
    stage.expandedCollections[item.id].height = Math.max(props.itemHeight, height || 0);
  }
  calculateChildPositions();
};

// Handles mouse down event for item selection and drag initiation.
const onMouseDown = (event, item, index) => {
  const id = item.id;
  const allItems = props.items;
  let itemType;

  if (item.collection_type_id) {
    itemType = 'collection';
  } else if (item.asset_type_id) {
    if (item.is_resource) {
      itemType = 'resource';
    } else {
      itemType = 'asset';
    }
  } else if (item.item_type) {
    itemType = item.item_type;
  }

  if (!stage.markedItems.includes(id) || stage.cmdOrCtrlKey(event)) {
    stage.handleClick(event, item, itemType, allItems);
  }

  menu.disableAllMenus();
  event.stopPropagation();
  dragItem(event, id);
};

// Handles mouse up event for item selection.
const onMouseUp = (event, item) => {
  if (dndStore.draggedItemId) return;
  
  const id = item.id;
  const allItems = props.items;
  let itemType;

  if (item.collection_type_id) {
    itemType = 'collection';
  } else if (item.asset_type_id) {
    if (item.is_resource) {
      itemType = 'resource';
    } else {
      itemType = 'asset';
    }
  } else if (item.item_type) {
    itemType = item.item_type;
  }

  if (stage.markedItems.includes(id) && !stage.cmdOrCtrlKey(event)) {
    stage.handleClick(event, item, itemType, allItems);
  }

  clearTimeout(dragTimer.value);
};

// Handles root container scroll events.
const onRootScroll = () => {
  updatePixelsAbove();
};

// Parses collection state into simpler structure with just IDs.
const parseCollectionState = (state) => {
  return {
    modified: state.modified_assets?.map(t => t.id).sort() || [],
    normal: state.normal_assets?.map(t => t.id).sort() || [],
    outdated: state.outdated_assets?.map(t => t.id).sort() || [],
    rebuildable: state.rebuildable_assets?.map(t => t.id).sort() || [],
    untracked_files: state.untracked_files?.map(f => f.id).sort() || [],
    untracked_folders: state.untracked_folders?.map(f => f.id).sort() || []
  };
};

// Refreshes this collection in response to fs changes.
// Tracked collections fetch state diffs; untracked rescan the folder.
const refreshView = async () => {
  if (!projectStore.activeProject) return;

  const project = projectStore.activeProject;
  const myToken = ++refreshToken;

  if (props.collectionType === 'untracked_collection') {
    try {
      const children = await CollectionService.GetCollectionChildren(
        project.uri,
        props.collectionId,
        project.working_directory,
        props.collectionPath,
        project.ignore_list,
        true
      );
      if (myToken !== refreshToken) return;

      const untracked = [
        ...(children.untracked_collections || children.untracked_folders || []),
        ...(children.untracked_assets || children.untracked_files || [])
      ];
      await assetStore.processUntrackedAssetsIcons(untracked);
      emitUntrackedUpdates(untracked);
    } catch (error) {
      console.error('Error scanning untracked folder:', error);
    }
    return;
  }

  try {
    const state = await CollectionService.GetCollectionChildrenState(
      project.uri,
      props.collectionId,
      project.working_directory,
      project.ignore_list
    );
    if (myToken !== refreshToken) return;

    const parsedState = parseCollectionState(state);
    const dataIsUnchanged = JSON.stringify(currentState.value) === JSON.stringify(parsedState);
    if (dataIsUnchanged) return;
    currentState.value = parsedState;

    const statusUpdates = [
      ...(state.normal_assets || []).map(a => ({ itemId: a.id, updates: [{ property: 'file_status', value: 'normal' }] })),
      ...(state.modified_assets || []).map(a => ({ itemId: a.id, updates: [{ property: 'file_status', value: 'modified' }] })),
      ...(state.outdated_assets || []).map(a => ({ itemId: a.id, updates: [{ property: 'file_status', value: 'outdated' }] })),
      ...(state.rebuildable_assets || []).map(a => ({ itemId: a.id, updates: [{ property: 'file_status', value: 'rebuildable' }] }))
    ];
    if (statusUpdates.length) emitItemUpdates(statusUpdates);

    const untracked = [
      ...(state.untracked_folders || []),
      ...(state.untracked_files || [])
    ];
    await assetStore.processUntrackedAssetsIcons(untracked);
    emitUntrackedUpdates(untracked);
  } catch (error) {
    console.error('Error getting collection children state:', error);
  }
};

// Selects item by index and scrolls into view.
const selectItem = (index) => {
  const allItems = allItemsIds.value;
  const newSelectedId = allItems[index];

  stage.firstSelectedItemId = newSelectedId;
  stage.markedItems = [newSelectedId];

  const viewItems = dndStore.allViewItems;
  const selectedItem = viewItems.find((item) => item.id === newSelectedId);

  stage.selectItem(selectedItem, selectedItem.type, true);

  dndStore.allElements[index].scrollIntoView({
    behavior: 'smooth',
    block: 'nearest'
  });
};

// Sets up intersection observer to track visibility and position.
const setupIntersectionObserver = () => {
  if (!scrollContainerRef.value || !rootScrollContainer.value) return;

  if (intersectionObserver.value) {
    intersectionObserver.value.disconnect();
  }

  intersectionObserver.value = new IntersectionObserver(
    (entries) => {
      const entry = entries[0];
      isVisible.value = entry.isIntersecting;
      intersectionRatio.value = entry.intersectionRatio;
      intersectionRect.value = entry.intersectionRect;

      const rootRect = rootScrollContainer.value.getBoundingClientRect();
      const targetRect = scrollContainerRef.value.getBoundingClientRect();

      pixelsAboveRoot.value = rootRect.top > targetRect.top ?
        Math.round(rootRect.top - targetRect.top) : 0;
    },
    {
      root: rootScrollContainer.value,
      threshold: [0, 0.1, 0.5, 1.0]
    }
  );

  intersectionObserver.value.observe(scrollContainerRef.value);
};

// Updates pixels above during scroll for accurate positioning.
const updatePixelsAbove = () => {
  if (!scrollContainerRef.value || !rootScrollContainer.value) return;

  const rootRect = rootScrollContainer.value.getBoundingClientRect();
  const targetRect = scrollContainerRef.value.getBoundingClientRect();

  pixelsAboveRoot.value = rootRect.top > targetRect.top ?
    Math.round(rootRect.top - targetRect.top) : 0;
};

// watchers
watch(() => props.items, (newItems, oldItems) => {
  calculateChildPositions();
}, { deep: true });

// Subscribes nested lists to fs events for their own folder.
// The root list is handled by Browser.vue.
const nestedWatchPath = computed(() =>
  props.isRoot ? null : (props.collectionPath || null)
);
useFsWatch(nestedWatchPath, refreshView);

watchEffect(() => {
  calculateChildPositions();
});

// lifecycle hooks
onUpdated(() => {
  dndStore.triggerDomUpdate();
});

onMounted(() => {
  nextTick(() => {
    // setupIntersectionObserver();
    updatePixelsAbove();

    if (rootScrollContainer.value) {
      rootScrollContainer.value.addEventListener('scroll', onRootScroll);
    }
    if (props.isRoot) {
      window.addEventListener('keydown', handleKeyDown);
    }
  });
});

onBeforeUnmount(async () => {
  if (intersectionObserver.value) {
    intersectionObserver.value.disconnect();
  }

  if (rootScrollContainer.value) {
    rootScrollContainer.value.removeEventListener('scroll', onRootScroll);
  }
  if (props.isRoot) {
    window.removeEventListener('keydown', handleKeyDown);
  }
});
</script>

<style scoped>
.virtua-scroll-conveyor {
  will-change: transform;
}

.virtua-scroll-viewport {
  box-sizing: border-box;
  overflow: hidden;
  position: relative;
  will-change: transform;
}
</style>