<template>
  <div class="virtua-scroll-viewport" ref="scrollContainerRef" :style="{ height: `${totalHeight}px` }"
    :data-visibility="containerVisibility">
    <div class="virtua-scroll-conveyor" :style="{ transform: `translateY(${offsetY}px)` }">
      <VirtuaItem v-for="child in visibleChildren" :child="items[child.index]" :key="child.index" :index="child.index"
        :itemHeight="itemHeight" :isExpanded="child.isExpanded" :onHeightChange="onHeightChange"
        :depth="depth" :getItemPosition="getItemPosition" :parentOffset="props.parentOffset" :offsetY="offsetY"
        :totalHeight="totalHeight" @mousedown="onMouseDown($event, child, index)"
        @mouseup="onMouseUp($event, child, index)" :ref="el => handleRef(child.id, el?.$el || el)" />
    </div>
  </div>
</template>

<script setup>
// imports
import { ref, computed, watchEffect, watch, onUpdated, inject, nextTick, onMounted, onBeforeUnmount } from 'vue';
import { Events } from '@wailsio/runtime';

// components
import VirtuaItem from '@/instances/common/components/VirtuaItem.vue';

// state imports
import { useScrollStore } from '@/stores/scroll';
import { useStageStore } from '@/stores/stages';
import { useAssetStore } from '@/stores/assets';
import { useMenu } from '@/stores/menu';
import { useDndStore } from '@/stores/dnd';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useProjectStore } from '@/stores/projects';
import { useCollectionStore } from '@/stores/collections';
import { CollectionService, FSService } from "@/../bindings/clustta/services";
import emitter from '@/lib/mitt';

// states/stores
const stage = useStageStore();
const menu = useMenu();
const dndStore = useDndStore();
const projectStore = useProjectStore();
const scrollStore = useScrollStore();
const modals = useDesktopModalStore();
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();

// props
const props = defineProps({
  items: { type: Array, default: [] },
  isRoot: { type: Boolean, default: false },
  containerHeight: { type: Number, required: true },
  renderAhead: { type: Number, default: 10 },
  depth: { type: Number, default: 0 },
  parentOffset: { type: Number, default: 0 },
  itemHeight: { type: Number, default: 60 }
});

// emits
const emit = defineEmits(['shift-parents', 'refreshData']);

// refs
const scrollContainerRef = ref(null);
const intersectionObserver = ref(null);
const isVisible = ref(false);
const intersectionRatio = ref(0);
const intersectionRect = ref(null);
const refreshDebounceTimer = ref(null);
const currentWatchedPath = ref(null);
const dragTimer = ref(null);
const childPositions = ref([0]);
const pixelsAboveRoot = ref(0);

const rootScrollContainer = inject('rootScrollContainer', null);

// computed props
// Total number of items in the list
const itemCount = computed(() => {
  return props.items.length;
});

// Calculate total height of all items
const totalHeight = computed(() =>
  childPositions.value[itemCount.value - 1] + getChildHeight(itemCount.value - 1)
);

// Calculate relative scroll position based on root or nested context
const relativeScrollTop = computed(() => {
  if (props.isRoot) return scrollStore.scrollTop;
  return pixelsAboveRoot.value;
});

// Find first visible node using binary search
const firstVisibleNode = computed(() => {
  return findStartNode(relativeScrollTop.value, childPositions.value, itemCount.value);
});

// Calculate start node with render ahead buffer
const startNode = computed(() =>
  Math.max(0, firstVisibleNode.value - props.renderAhead)
);

// Find last visible node within container height
const lastVisibleNode = computed(() =>
  findEndNode(childPositions.value, firstVisibleNode.value, itemCount.value, props.containerHeight)
);

// Calculate end node with render ahead buffer
const endNode = computed(() =>
  Math.min(itemCount.value - 1, lastVisibleNode.value + props.renderAhead)
);

// Calculate number of visible nodes
const visibleNodeCount = computed(() =>
  endNode.value - startNode.value + 1
);

// Calculate vertical offset for scroll position
const offsetY = computed(() => {
  return childPositions.value[startNode.value];
});

// Get visible children based on scroll position
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

// Get currently selected item id
const selectedId = computed(() => {
  return stage.firstSelectedItemId;
});

// Get all item ids from dnd store
const allItemsIds = computed(() => {
  const allItemsIds = dndStore.allElements.map((item) => item.id);
  return allItemsIds;
});

// Determine container visibility state
const containerVisibility = computed(() => {
  if (!isVisible.value) return 'hidden';
  if (intersectionRatio.value === 1) return 'fully-visible';
  return 'partially-visible';
});

// Get all untracked items from root items
const previousUntracked = computed(() => {
  const allUntracked = props.items.filter((item) => item.type === 'untracked_task' || item.type === 'untracked_entity');
  return allUntracked;
});

// Get current file path location for watching
const location = computed(() => {
  return collectionStore.navigatedCollection ? collectionStore.navigatedCollection.file_path : projectStore.activeProject.working_directory;
});

// functions
// Add or remove element reference to dnd store
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

// Handle mouse down event for item selection and drag initiation
const onMouseDown = (event, item, index) => {
  const id = item.id;
  const allItems = props.items;
  let itemType;

  if (item.entity_type_id) {
    itemType = 'entity';
  } else if (item.task_type_id) {
    if (item.is_resource) {
      itemType = 'resource';
    } else {
      itemType = 'task';
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

// Initialize drag operation through dnd store
const onDragStart = (e, id) => {
  stage.firstSelectedItemId = '';

  if (!id) {
    return;
  }
  dndStore.onDragStart(e, id);
};

// Start drag timer for delayed drag operation
const dragItem = (event, id) => {
  if (stage.operationActive) return;
  dragTimer.value = setTimeout(() => {
    onDragStart(event, id);
  }, dndStore.dragDelay);
};

// Handle mouse up event for item selection
const onMouseUp = (event, item) => {
  if (dndStore.draggedItemId) return;
  
  const id = item.id;
  const allItems = props.items;
  let itemType;

  if (item.entity_type_id) {
    itemType = 'entity';
  } else if (item.task_type_id) {
    if (item.is_resource) {
      itemType = 'resource';
    } else {
      itemType = 'task';
    }
  } else if (item.item_type) {
    itemType = item.item_type;
  }

  if (stage.markedItems.includes(id) && !stage.cmdOrCtrlKey(event)) {
    stage.handleClick(event, item, itemType, allItems);
  }

  clearTimeout(dragTimer.value);
};

// Get height of child item considering expanded state
const getChildHeight = (index) => {
  const item = props.items[index];
  return item && item.id in stage.expandedEntities ?
    stage.expandedEntities[item.id]["height"] || props.itemHeight : props.itemHeight;
};

// Handle height change of expanded items
const onHeightChange = (index, height) => {
  if (height > props.itemHeight) {
    const item = props.items[index];
    if (item && item.id) {
      stage.expandedEntities[item.id]["height"] = height;
    }
  }
  calculateChildPositions();
};

// Calculate cumulative positions of all child items
const calculateChildPositions = () => {
  emit('shift-parents');
  const positions = [0];
  for (let i = 1; i < itemCount.value; i++) {
    positions.push(positions[i - 1] + getChildHeight(i - 1));
  }
  childPositions.value = positions;
};

// Get position of item by index
const getItemPosition = (index) => {
  return childPositions.value[index] || 0;
};

// Binary search to find start node based on scroll position
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

// Find end node based on container height
const findEndNode = (nodePositions, startNode, itemCount, height) => {
  let endNode;
  for (endNode = startNode; endNode < itemCount; endNode++) {
    if (nodePositions[endNode] > nodePositions[startNode] + height) {
      return endNode;
    }
  }
  return endNode;
};

// Setup intersection observer to track visibility and position
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

// Update pixels above during scroll for accurate positioning
const updatePixelsAbove = () => {
  if (!scrollContainerRef.value || !rootScrollContainer.value) return;

  const rootRect = rootScrollContainer.value.getBoundingClientRect();
  const targetRect = scrollContainerRef.value.getBoundingClientRect();

  pixelsAboveRoot.value = rootRect.top > targetRect.top ?
    Math.round(rootRect.top - targetRect.top) : 0;
};

// Handle root container scroll events
const onRootScroll = () => {
  updatePixelsAbove();
};

// Handle keyboard navigation with arrow keys
const handleKeyDown = (event) => {
  if (modals.activeModal) {
    return
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

// Select item by index and scroll into view
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

// Emit updates to item data across components
const emitItemUpdates = (itemId, updates) => {
  const updateData = { itemId, updates };
  
  // Emit to both Browser and VirtuaItem components
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

// Fetch and update collection children state from backend
const refreshView = async () => {
  const project = projectStore.activeProject;
  const entityId = collectionStore.navigatedCollection?.id;
  
  try {
    const state = await CollectionService.GetCollectionChildrenState(
      project.uri,
      entityId,
      project.working_directory,
      project.ignore_list
    );
    
    if (state.normal_tasks && state.normal_tasks.length > 0) {
      state.normal_tasks.forEach(task => {
        emitItemUpdates(task.id, [
          { property: 'file_status', value: 'normal' }
        ]);
      });
    }

    if (state.modified_tasks && state.modified_tasks.length > 0) {
      state.modified_tasks.forEach(task => {
        emitItemUpdates(task.id, [
          { property: 'file_status', value: 'modified' }
        ]);
      });
    }
    
    if (state.outdated_tasks && state.outdated_tasks.length > 0) {
      state.outdated_tasks.forEach(task => {
        emitItemUpdates(task.id, [
          { property: 'file_status', value: 'outdated' }
        ]);
      });
    }
    
    if (state.rebuildable_tasks && state.rebuildable_tasks.length > 0) {
      state.rebuildable_tasks.forEach(task => {
        emitItemUpdates(task.id, [
          { property: 'file_status', value: 'rebuildable' }
        ]);
      });
    }
    
    const currentUntrackedFolders = state.untracked_folders || [];
    const currentUntrackedFiles = state.untracked_files || [];
    const currentUntracked = [...currentUntrackedFolders, ...currentUntrackedFiles];
    
    if (currentUntracked !== previousUntracked.value) {
      const allUntrackedItems = [...currentUntrackedFolders, ...currentUntrackedFiles];
      await assetStore.processUntrackedAssetsIcons(allUntrackedItems);
      emitter.emit('update-untracked-items', allUntrackedItems);
    }
    
  } catch (error) {
    console.error('Error getting collection children state:', error);
  }
};

// Debounce refresh view to prevent rapid consecutive calls
const debouncedRefreshView = () => {
  if (refreshDebounceTimer.value) {
    clearTimeout(refreshDebounceTimer.value);
  }
  refreshDebounceTimer.value = setTimeout(() => {
    refreshView();
  }, 200);
};

// Handle file system change events by refreshing view
const handleFSChange = (event) => {
  console.log(event);
  debouncedRefreshView();
};

// watchers
// Recalculate child positions when items change
watch(() => props.items, (newItems, oldItems) => {
  calculateChildPositions();
}, { deep: true });

// Watch location changes and update file system watcher
watch(() => location.value, async (newPath, oldPath) => {
  if (oldPath && currentWatchedPath.value) {
    try {
      await FSService.RemoveWatcherFolder(oldPath);
    } catch (error) {
      console.error('Error removing watcher:', error);
    }
  }
  
  if (newPath) {
    try {
      const exists = await FSService.DirExists(newPath);
      if (exists) {
        await FSService.AddWatcherFolder(newPath);
        currentWatchedPath.value = newPath;
      }
    } catch (error) {
      console.error('Error adding watcher:', error);
    }
  }
}, { immediate: true });

watchEffect(() => {
  calculateChildPositions();
});

// lifecycle hooks
onUpdated(() => {
  dndStore.triggerDomUpdate();
});

onMounted(() => {
  nextTick(() => {
    setupIntersectionObserver();
    updatePixelsAbove();

    if (rootScrollContainer.value) {
      rootScrollContainer.value.addEventListener('scroll', onRootScroll);
    }
    if (props.isRoot) {
      window.addEventListener('keydown', handleKeyDown);
    }
  });
  
  Events.On('fs-change', handleFSChange);
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
  
  Events.Off('fs-change', handleFSChange);
  
  if (refreshDebounceTimer.value) {
    clearTimeout(refreshDebounceTimer.value);
  }
});
</script>

<style scoped>
.virtua-scroll-viewport {
  /* margin-top: 1px; */
  position: relative;
  overflow: hidden;
  will-change: transform;
  box-sizing: border-box;
}

.virtua-scroll-conveyor {
  will-change: transform;
}
</style>




