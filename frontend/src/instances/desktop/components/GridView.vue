<template>
	<div class="navigator-root-viewport" @scroll="disableMenu()">

    <GridSkeleton v-if="!assetStore.assetsLoaded" :height="containerHeight" />

    <div v-else ref="navigatorRoot" class="navigator-root">
      
      <div v-if="entityItems.length > 0" class="navigator-item-container" :style="gridStyles">
          <GridItem v-for="(child, index) in entityItems" :child="child" :key="child.index" :index="index" 
          @mousedown="onMouseDown($event, child, index)"
          @mouseup="onMouseUp($event, child, index)" :ref="el => handleRef(child.id, el?.$el || el)" />
      </div>
      
      <div v-if="taskItems.length > 0" class="navigator-item-container" :style="gridStyles">
          <GridItem v-for="(child, index) in taskItems" :child="child" :key="child.index" :index="index" 
          @mousedown="onMouseDown($event, child, index)"
          @mouseup="onMouseUp($event, child, index)" :ref="el => handleRef(child.id, el?.$el || el)" />
      </div>
  </div>

</div>
</template>

<script setup>
// imports
import { computed, ref, nextTick, onUnmounted, watch } from 'vue';
import { Events } from '@wailsio/runtime';

// components
import GridItem from '@/instances/common/components/GridItem.vue';
import GridSkeleton from '@/instances/desktop/components/GridSkeleton.vue';

// state imports
import { useMenu } from '@/stores/menu';
import { useStageStore } from '@/stores/stages';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useAssetStore } from '@/stores/assets';
import { useProjectStore } from '@/stores/projects';
import { useDndStore } from '@/stores/dnd';
import { CollectionService, FSService } from '@/../bindings/clustta/services';
import emitter from '@/lib/mitt';

// states/stores
const menu = useMenu();
const stage = useStageStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const assetStore = useAssetStore();
const projectStore = useProjectStore();
const dndStore = useDndStore();

// props
const props = defineProps({
  rootItems: { type: Array, default: [] }
});

// refs
const navigatorRoot = ref(null);
const refreshDebounceTimer = ref(null);
const currentWatchedPath = ref(null);
const dragTimer = ref(null);

// computed props
// Grid styling based on commonStore grid size
const gridStyles = computed(() => ({
  display: 'grid',
  boxSizing: 'border-box',
  gridTemplateColumns: `repeat(auto-fill, minmax(${commonStore.gridSize}px, 1fr))`,
  gap: '10px',
  width: '100%'
}));

// Filter root items to get only entity type items
const entityItems = computed(() => {
  return props.rootItems.filter(item => 
    item.type === 'entity' || item.type === 'untracked_entity'
  );
});

// Filter root items to get only task type items
const taskItems = computed(() => {
  return props.rootItems.filter(item => 
    item.type !== 'entity' && item.type !== 'untracked_entity'
  );
});

// Get container height from navigator root element
const containerHeight = computed(() => {
  return navigatorRoot.value?.getBoundingClientRect().height || 500;
});

// Get all untracked items from root items
const previousUntracked = computed(() => {
  const allUntracked = props.rootItems.filter((item) => item.type === 'untracked_task' || item.type === 'untracked_entity');
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

// Disable all context menus
const disableMenu = () => {
  menu.disableAllMenus();
};

// Handle file system change events by refreshing view
const handleFSChange = (event) => {
  debouncedRefreshView();
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
    
    collectionStore.loadCollectionStateFlags();
    
  } catch (error) {
    console.error('Error getting collection children state:', error);
  }
};

// Emit updates to item data across components
const emitItemUpdates = (itemId, updates) => {
  const updateData = { itemId, updates };
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

// Handle mouse down event for item selection and drag initiation
const onMouseDown = (event, item, index) => {
  const id = item.id;
  const allItems = props.rootItems;
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
  const allItems = props.rootItems;
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

// watchers
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
    console.log(newPath);
    try {
      const exists = await FSService.DirExists(newPath);
      console.log(exists)
      if (exists) {
        await FSService.AddWatcherFolder(newPath);
        currentWatchedPath.value = newPath;
      }
    } catch (error) {
      console.error('Error adding watcher:', error);
    }
  }
}, { immediate: true });

// lifecycle hooks
Events.On('fs-change', handleFSChange);

onUnmounted(async () => {
  Events.Off('fs-change', handleFSChange);
  
  if (refreshDebounceTimer.value) {
    clearTimeout(refreshDebounceTimer.value);
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.navigator-root-viewport {
	padding-right: .4rem;
	position: relative;
	height: 100%;
	width: 100%;
  flex-direction: column;
	box-sizing: border-box;
	align-items: flex-start;
	justify-content: center;
	overflow: hidden;
  overflow-y: scroll;
  gap: 10px;
  border-radius: var(--small-radius);
}

.navigator-root-viewport::-webkit-scrollbar {
  width: 6px;
}

.virtua-scroll-tray::-webkit-scrollbar {
  width: 4px;
}

.navigator-root-viewport::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--light-steel);
}

.navigator-root-viewport::-webkit-scrollbar-track {
  border-radius: 10px;
}

.navigator-root {
	position: relative;
  height: min-content;
	width: 100%;
	display: flex;
  flex-direction: column;
	box-sizing: border-box;
	align-items: flex-start;
	justify-content: center;
	overflow: hidden;
  gap: 10px;
  border-radius: var(--small-radius);
}


.navigator-item-container {
  position: relative;
  height: 100%;
  height: min-content;
  width: 100%;
  box-sizing: border-box;
  align-items: flex-start;
  justify-content: flex-start;
  overflow: hidden;
  flex-direction: column;
  gap: .5rem;
}
</style>







