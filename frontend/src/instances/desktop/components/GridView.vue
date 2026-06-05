<template>
	<div class="navigator-root-viewport" @scroll="disableMenu()">

    <GridSkeleton v-if="!assetStore.assetsLoaded" :height="containerHeight" />

    <div v-else ref="navigatorRoot" class="navigator-root">
      
      <div v-if="collectionItems.length > 0" class="navigator-item-container" :style="gridStyles">
          <GridItem v-for="(child, index) in collectionItems" :child="child" :key="child.index" :index="index" 
          @mousedown="onMouseDown($event, child, index)"
          @mouseup="onMouseUp($event, child, index)" :ref="el => handleRef(child.id, el?.$el || el)" />
      </div>
      
      <div v-if="assetItems.length > 0" class="navigator-item-container" :style="gridStyles">
          <GridItem v-for="(child, index) in assetItems" :child="child" :key="child.index" :index="index" 
          @mousedown="onMouseDown($event, child, index)"
          @mouseup="onMouseUp($event, child, index)" :ref="el => handleRef(child.id, el?.$el || el)" />
      </div>
  </div>

</div>
</template>

<script setup>
// imports
import { computed, ref, nextTick } from 'vue';

// components
import GridItem from '@/instances/common/components/GridItem.vue';
import GridSkeleton from '@/instances/desktop/components/GridSkeleton.vue';

// state imports
import { useMenu } from '@/stores/menu';
import { useStageStore } from '@/stores/stages';
import { useCommonStore } from '@/stores/common';
import { useAssetStore } from '@/stores/assets';
import { useDndStore } from '@/stores/dnd';

// states/stores
const menu = useMenu();
const stage = useStageStore();
const commonStore = useCommonStore();
const assetStore = useAssetStore();
const dndStore = useDndStore();

// props
const props = defineProps({
  rootItems: { type: Array, default: [] }
});

// refs
const navigatorRoot = ref(null);
const dragTimer = ref(null);

// computed props
// Grid styling based on commonStore grid size
const gridStyles = computed(() => ({
  display: 'grid',
  boxSizing: 'border-box',
  gridTemplateColumns: `repeat(auto-fill, minmax(${commonStore.gridSize}px, 1fr))`,
  gap: '8px',
  width: '100%'
}));

// Filter root items to get only collection type items
const collectionItems = computed(() => {
  return props.rootItems.filter(item => 
    item.type === 'collection' || item.type === 'untracked_collection'
  );
});

// Filter root items to get only asset type items
const assetItems = computed(() => {
  return props.rootItems.filter(item => 
    item.type !== 'collection' && item.type !== 'untracked_collection'
  );
});

// Get container height from navigator root element
const containerHeight = computed(() => {
  return navigatorRoot.value?.getBoundingClientRect().height || 500;
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

// Handle mouse down event for item selection and drag initiation
const onMouseDown = (event, item, index) => {
  const id = item.id;
  const allItems = props.rootItems;
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
  if (!stage.isContextMenuClick(event)) dragItem(event, id);
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

  if (stage.markedItems.includes(id) && !stage.cmdOrCtrlKey(event) && !stage.isContextMenuClick(event)) {
    stage.handleClick(event, item, itemType, allItems);
  }

  clearTimeout(dragTimer.value);
};

// watchers

// lifecycle hooks
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
  background-color: var(--surface-4);
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
  gap: 8px;
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







