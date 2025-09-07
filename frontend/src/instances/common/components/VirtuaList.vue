<template>
  <div class="virtua-scroll-viewport" ref="scrollContainerRef" :style="{ height: `${totalHeight}px` }"
    @mouseenter="updateAssetState" @mouseleave="onMouseLeave" :data-visibility="containerVisibility">
    <div class="virtua-scroll-conveyor" :style="{ transform: `translateY(${offsetY}px)` }">
      <VirtuaItem @refreshData="emit('refreshData')" v-for="child in visibleChildren" :child="items[child.index]" :key="child.index" :index="child.index"
        :itemHeight="itemHeight" :isExpanded="child.isExpanded" :onHeightChange="onHeightChange"
        :depth="depth" :getItemPosition="getItemPosition" :parentOffset="props.parentOffset" :offsetY="offsetY"
        :totalHeight="totalHeight" @mousedown="onMouseDown($event, child, index)"
        @mouseup="onMouseUp($event, child, index)" :ref="el => handleRef(child.id, el?.$el || el)" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watchEffect, watch, onUpdated, provide, inject, nextTick, onMounted, onBeforeUnmount } from 'vue';

import { useScrollStore } from '@/stores/scroll';
import { useStageStore } from '@/stores/stages';
import { useMenu } from '@/stores/menu';
import { useDndStore } from '@/stores/dnd';
import { useAssetStore } from '@/stores/assets';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useCommonStore } from '@/stores/common';

import VirtuaItem from '@/instances/common/components/VirtuaItem.vue';
import { TaskService } from '@/../bindings/clustta/services';
import { useProjectStore } from '@/stores/projects';

const stage = useStageStore();
const menu = useMenu();
const dndStore = useDndStore();
const assetStore = useAssetStore();
const projectStore = useProjectStore();
const scrollStore = useScrollStore();
const modals = useDesktopModalStore();
const commonStore = useCommonStore();

const scrollContainerRef = ref(null);
const intersectionObserver = ref(null);
const isVisible = ref(false);
const intersectionRatio = ref(0);
const intersectionRect = ref(null);

const rootScrollContainer = inject('rootScrollContainer', null);

const props = defineProps({
  items: { type: Array, default: [] },
  isRoot: { type: Boolean, default: false },
  // child: { type: Object, default: null },
  containerHeight: { type: Number, required: true },
  renderAhead: { type: Number, default: 10 },
  depth: { type: Number, default: 0 },
  parentOffset: { type: Number, default: 0 },
  itemHeight: { type: Number, default: 60 }
});

// emits
const emit = defineEmits(['shift-parents', 'refreshData']);

// drag and drop
const handleRef = async (id, el) => {
  
  if (!el) {
    // removeLocalRef(id);
    dndStore.removeRef(id);
    dndStore.removeVisibleItemsRef(id);
    return
  }

  await nextTick()

  // Get the actual DOM element
  const domElement = el instanceof HTMLElement ? el : el.$el

  if (domElement) {
    dndStore.addRef(id, domElement);
  }
};

const dragTimer = ref(null);

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

  if(!stage.markedItems.includes(id) || event.ctrlKey){
    stage.handleClick(event, item, itemType, allItems);
  }

  menu.disableAllMenus();
  event.stopPropagation();
  dragItem(event, id);
};

const dragItem = (event, id) => {
  if(stage.operationActive) return
  dragTimer.value = setTimeout(() => {
    onDragStart(event, id);
  }, dndStore.dragDelay);

}

const onMouseUp = (event, item) => {

  if(dndStore.draggedItemId) return 
  
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

  if (stage.markedItems.includes(id) && !event.ctrlKey) {
    stage.handleClick(event, item, itemType, allItems);
  }

  // console.log('mouse-up')
  clearTimeout(dragTimer.value);
};

const onDragStart = (e, id) => {
  stage.firstSelectedItemId = '';

  if (!id) {
    return
  }
  dndStore.onDragStart(e, id);
};

const itemCount = computed(() => {
  return  props.items.length
});

const getChildHeight = (index) => {
  const item = props.items[index];
  return item && item.id in stage.expandedEntities ?
    stage.expandedEntities[item.id]["height"] || props.itemHeight : props.itemHeight;
};

const onHeightChange = (index, height) => {
  if (height > props.itemHeight) {
    const item = props.items[index];
    if (item && item.id) {
      // Store the height directly in expandedEntities
      stage.expandedEntities[item.id]["height"] = height;
    }
  }
  calculateChildPositions();
  // emit('shift-parents')
};

watch(() => props.items, (newItems, oldItems) => {
  // Recalculate positions and notify parent components
  calculateChildPositions();
  // emit('shift-parents');
}, { deep: true });


onUpdated(() => {
  dndStore.triggerDomUpdate()
})

const childPositions = ref([0]);

const calculateChildPositions = () => {
  emit('shift-parents');
  const positions = [0];
  for (let i = 1; i < itemCount.value; i++) {
    positions.push(positions[i - 1] + getChildHeight(i - 1));
  }
  childPositions.value = positions;
};

watchEffect(() => {
  calculateChildPositions();
});

const totalHeight = computed(() =>
  childPositions.value[itemCount.value - 1] + getChildHeight(itemCount.value - 1)
);

// binary search
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

const findEndNode = (nodePositions, startNode, itemCount, height) => {
  let endNode;
  for (endNode = startNode; endNode < itemCount; endNode++) {
    if (nodePositions[endNode] > nodePositions[startNode] + height) {
      // console.log(startNode)
      return endNode;
    }
  }
  return endNode;
};
const relativeScrollTop = computed(() => {
  if (props.isRoot) return scrollStore.scrollTop;
  return pixelsAboveRoot.value
});

const firstVisibleNode = computed(() => {
  return findStartNode(relativeScrollTop.value, childPositions.value, itemCount.value)
});

const startNode = computed(() =>
  Math.max(0, firstVisibleNode.value - props.renderAhead)
);

const lastVisibleNode = computed(() =>
  findEndNode(childPositions.value, firstVisibleNode.value, itemCount.value, props.containerHeight)
);

const endNode = computed(() =>
  Math.min(itemCount.value - 1, lastVisibleNode.value + props.renderAhead)
);

const visibleNodeCount = computed(() =>
  endNode.value - startNode.value + 1
);

const offsetY = computed(() => {
  return childPositions.value[startNode.value];
});

const getItemPosition = (index) => {
  return childPositions.value[index] || 0;
};

const visibleChildren = computed(() =>
  Array(visibleNodeCount.value)
    .fill(null)
    .map((_, index) => {
      const actualIndex = index + startNode.value;
      const item = props.items[actualIndex];

      return item ? {
        ...item,
        index: actualIndex,
        // isExpanded: !!expandedItems.value[actualIndex]
        isExpanded: false
      } : null;
    })
    .filter(Boolean)
);

// Add a ref to store the precise pixels above
const pixelsAboveRoot = ref(0);

// Update the setupIntersectionObserver function to calculate pixels above
const setupIntersectionObserver = () => {
  if (!scrollContainerRef.value || !rootScrollContainer.value) return;

  // Clean up existing observer if any
  if (intersectionObserver.value) {
    intersectionObserver.value.disconnect();
  }

  intersectionObserver.value = new IntersectionObserver(
    (entries) => {
      const entry = entries[0];
      isVisible.value = entry.isIntersecting;
      intersectionRatio.value = entry.intersectionRatio;
      intersectionRect.value = entry.intersectionRect;

      // Calculate exact pixels above root container
      const rootRect = rootScrollContainer.value.getBoundingClientRect();
      const targetRect = scrollContainerRef.value.getBoundingClientRect();

      // If the target's top is above the root's top, calculate pixels above
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

// Also add a continuous calculation during scrolling to get more accurate updates
const updatePixelsAbove = () => {
  if (!scrollContainerRef.value || !rootScrollContainer.value) return;

  const rootRect = rootScrollContainer.value.getBoundingClientRect();
  const targetRect = scrollContainerRef.value.getBoundingClientRect();

  pixelsAboveRoot.value = rootRect.top > targetRect.top ?
    Math.round(rootRect.top - targetRect.top) : 0;
};

// Call this on scroll events
const onRootScroll = () => {
  updatePixelsAbove();
};

const selectedId = computed(() => {
  return stage.firstSelectedItemId;
});

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

const allItemsIds = computed(() => {
  const allItemsIds = dndStore.allElements.map((item) => item.id);
  return allItemsIds
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
});

onBeforeUnmount(() => {
  if (intersectionObserver.value) {
    intersectionObserver.value.disconnect();
  }

  // Remove scroll event listener
  if (rootScrollContainer.value) {
    rootScrollContainer.value.removeEventListener('scroll', onRootScroll);
  }
  if (props.isRoot) {
    window.removeEventListener('keydown', handleKeyDown);
  }
});

// We can use the intersection data to optimize visibility calculations
const containerVisibility = computed(() => {
  if (!isVisible.value) return 'hidden';
  if (intersectionRatio.value === 1) return 'fully-visible';
  return 'partially-visible';
});

const updateAssetState = async () => {
  for (let i = 0; i < visibleChildren.value.length; i++) {
    let visibleChild = visibleChildren.value[i];
    if (visibleChild.type == "task") {
      let fileStatus = await assetStore.getAssetFileStatus(visibleChild)
      if (fileStatus === "modified") {
        if (!assetStore.modifiedTasksPath.includes(visibleChild.task_path)) {
          assetStore.addModifiedAssetPath(visibleChild.task_path)
        }
      }
      props.items[visibleChild.index].file_status = fileStatus;
    }
  }
};

const onMouseLeave = () => {
};

</script>

<style scoped>
.selected {
  outline: 1px solid red;
  outline-offset: -1px;
}

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

.relative-scroll-top {
  position: fixed;
  left: 0;
  top: 300px;
  background-color: white;
  padding: .5rem;
  z-index: 1000;
  font-weight: 300;
  color: black;
}

.data {
  position: absolute;
  color: wheat;
}
</style>




