<template>

  <div class="kanban-root" @mousemove="onDrag" 
  @mouseup="onDragStop(draggedItemRefs[dndStore.draggedItemId])"
  @mouseleave="stopScrolling" ref="scrollContainerRef">
  
  <div class="kanban-container" >

    <div v-for="column in filledColumns" :key="column.id" :ref="el => el && (targetRefs[column.id] = el)" 
      class="column" :class="{ 'column-minimized' : minimizedColumns.includes(column.id)}" >

      <div  class="column-header" :style="{ outline :  '2px solid' + column.color }" @dblclick="minimizeColumn(column.id)" 
      :class="{ 'column-header-minimized' : minimizedColumns.includes(column.id)}" @click="loadAssetTasks">

        <div class="column-header-icon" >
            <img class="small-icons no-filter" :src="getStatusIcon(column.short_name)">
        </div>

        <div class="column-header-meta" >
          <div class="column-header-text" >
            {{ utils.capitalizeStr(column.short_name) }}
          </div>

          <div v-if="column.cards.length" class="column-header-count" >
            {{ column.cards.length }}
          </div>
        </div>

      </div>

      <div class="column-cards-container" :class="{ 'column-cards-container-minimized' : minimizedColumns.includes(column.id)}">

        <template v-for="(card, index) in column.cards" :key="card.id">
          <!-- Insert indicator at the top or between cards -->
          <div v-if="hoveredColumnId === column.id && hoveredCardIndex === index" 
               class="card-insert-indicator"></div>
          
          <div :class="{ 'column-card': true, 'dragging': dndStore.draggedItemId === card.id }"
               :ref="el => el && (draggedItemRefs[card.id] = el)" 
               @mousedown="onDragStart($event, card.id, minimizedColumns.includes(column.id))">
            <TaskItemCard v-bind="card" :task="card" :isTask="isTask(card.task_name)" :isMinimized="minimizedColumns.includes(column.id)" />
          </div>
          
        </template>
        
        <!-- Insert indicator at the end of the column -->
        <div v-if="hoveredColumnId === column.id && hoveredCardIndex >= column.cards.length" 
             class="card-insert-indicator"></div>
        
      </div>

    </div>

    <div id="ghost-card" ref="ghostCard"
        :style="ghostCardStyles"
        :class="{ 'column-card': true, 'active': dndStore.draggedItemId !== null, leaving: dndStore.ghostCardStyle.leaving }">
        <TaskItemCard v-bind="draggedCard" :task="draggedCard"  :isTask="isTask(draggedCard.task_name)"/>
    </div>

  </div>
  </div>
</template>

<script setup>
// imports
import { ProjectService, EntityService, TaskService, CheckpointService, TrashService } from "@/../bindings/clustta/services";

import { reactive, computed, ref, onMounted, onUnmounted, watch } from 'vue';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// stores/state imports
import { useCommonStore } from '@/stores/common';
import { useEntityStore } from '@/stores/entity';
import { useAssetStore } from '@/stores/assets';
import { useIconStore } from '@/stores/icons';

import { useStatusStore } from '@/stores/status';
import { useDndStore } from '@/stores/dnd';
import { useProjectStore } from '@/stores/projects';

// components
import TaskItemCard from '@/instances/desktop/components/TaskItemCard.vue'

// stores/state 
const dndStore = useDndStore();
const assetStore = useAssetStore();
const statusStore = useStatusStore();
const commonStore = useCommonStore();
const entityStore = useEntityStore();
const projectStore = useProjectStore();
const iconStore = useIconStore();

// props
const props = defineProps({
  
  tasks: Array,
  filtersActive: Boolean,
  showThumbs: {
    type: Boolean,
    default: false
  },


  // Size of the edge area that triggers scrolling (in pixels)
  initialEdgeSize: {
    type: Number,
    default: 50
  },
  // Base scroll speed in pixels per animation frame
  initialScrollSpeed: {
    type: Number,
    default: 10
  }
});

// refs
const targetRefs = ref({});
const draggedItemRefs = ref({});
const minimizedColumns = ref([]);
const cards = ref([]);
const filteredCards = ref([]);
const hoveredCardIndex = ref(-1);
const hoveredColumnId = ref(null);

// Fetch asset tasks from the backend
const loadAssetTasks = async () => {
  try {
    const projectPath = projectStore.activeProject?.uri;
    if (projectPath) {
      const tasks = await TaskService.GetAssetTasks(projectPath);
      await assetStore.processTasksIconsAndPreviews(tasks);
      cards.value = tasks; // Update cards ref with the fetched tasks
      await updateFilteredCards(); // Update filtered cards
    }
  } catch (error) {
    console.error('Error loading asset tasks:', error);
  }
};

// Update filtered cards using the async filterTasks method
const updateFilteredCards = async () => {
  try {
    
    if (cards.value && cards.value.length > 0) {
      const filtered = await assetStore.filterTasks(cards.value);
      filteredCards.value = filtered || [];
    } else {
      filteredCards.value = [];
    }
  } catch (error) {
    console.error('Error filtering cards:', error);
    filteredCards.value = cards.value || []; // Fallback to unfiltered cards
  }
};

const statuses = computed(() => { 
  const taskStatuses = statusStore.statuses;
  const order = { 'todo': 1, 'ready': 2, 'wip': 3, 'wfa': 4, 'retake': 5, 'done': 6 };
  return taskStatuses.sort((a, b) => {
    const orderA = order[a.short_name] || Infinity;
    const orderB = order[b.short_name] || Infinity;
    return orderA - orderB;
  });
});

const statusColumns = reactive(statuses.value);


const filledColumns = computed(() => {
  const result = statusColumns.map(column => ({
    ...column,
    cards: filteredCards.value
      .filter(card => card.status_id === column.id)
      // .sort((a, b) => priorityMap.value[b.priority] - priorityMap.value[a.priority])
  }));
  // console.log(result)
  return result
});

const getStatusIcon = (status) => {
  return '/status-icons/status_' + status + '.svg'; 
}

// auto scroll on edge /////////////////////////////////////////////////////////////////////////
// Refs
const scrollContainerRef = ref(null)
const edgeSize = ref(props.initialEdgeSize)
const scrollSpeed = ref(props.initialScrollSpeed)

// Scroll tracking state
const isScrolling = ref(false)
const scrollDirection = ref(0)
let animationFrameId = null

// Cursor debug info
const cursorPosition = ref({ x: 0, edge: 'none' })


// Start the scrolling animation
const startScrolling = () => {
  if (isScrolling.value) {
    const el = scrollContainerRef.value
    if (!el) return
    
    // Apply the scroll
    const scrollAmount = scrollDirection.value * scrollSpeed.value
    el.scrollLeft += scrollAmount
    
    // Continue animation
    animationFrameId = requestAnimationFrame(startScrolling)
  }
}

// Stop scrolling and clean up
const stopScrolling = () => {
  isScrolling.value = false
  scrollDirection.value = 0
  
  if (animationFrameId !== null) {
    cancelAnimationFrame(animationFrameId)
    animationFrameId = null
  }
}

// Handle mouse movement inside the container
const handleMouseMove = (e) => {
  const el = scrollContainerRef.value
  // console.log(el)
  if (!el) return
  
  const rect = el.getBoundingClientRect()
  const containerWidth = rect.width
  const cursorX = e.clientX - rect.left
  
  // Update cursor position for debug display
  cursorPosition.value = { x: cursorX, edge: 'none' }
  
  // If we're already scrolling, cancel the animation
  if (isScrolling.value) {
    if (animationFrameId !== null) {
      cancelAnimationFrame(animationFrameId)
      animationFrameId = null
    }
    isScrolling.value = false
  }
  
  // Check if cursor is near left edge
  if (cursorX <= edgeSize.value) {
    const speedFactor = 1 - (cursorX / edgeSize.value)
    scrollDirection.value = -speedFactor
    cursorPosition.value.edge = 'left'
    isScrolling.value = true
    startScrolling()
  }
  // Check if cursor is near right edge
  else if (cursorX >= containerWidth - edgeSize.value) {
    const speedFactor = (cursorX - (containerWidth - edgeSize.value)) / edgeSize.value
    scrollDirection.value = speedFactor
    cursorPosition.value.edge = 'right'
    isScrolling.value = true
    startScrolling()
  }
};

////////////////////////////////////////////////////////////////////////////////////////////////


const draggedCard = computed(() => {
  return cards.value.find(card => card.id === dndStore.draggedItemId) || { name: '' };
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
const minimizeColumn = (id) => {
  if(!minimizedColumns.value.includes(id)){
    minimizedColumns.value.push(id);
  } else if(minimizedColumns.value.includes(id)) {
    const idIndex =  minimizedColumns.value.indexOf(id);
    minimizedColumns.value.splice(idIndex, 1);
  }
};

const isTask = (taskName) => {
  if(taskName){
    return false
  } else {
    return true
  }
}

const findInsertionIndex = (cardElements, mouseY) => {
  if (cardElements.length === 0) return 0;
  
  for (let i = 0; i < cardElements.length; i++) {
    const cardRect = cardElements[i].getBoundingClientRect();
    const cardMiddle = cardRect.top + cardRect.height / 2;
    
    if (mouseY < cardMiddle) {
      return i;
    }
  }
  
  // If we've gone through all cards and mouse is below them all
  return cardElements.length;
};

const onDragStart = (e, id, isMinimized) => {
  if(isMinimized){
    return
  }
  // console.log(draggedItemRefs.value)
  let selectedCard = draggedItemRefs.value[id];
  let cardRect = selectedCard.getBoundingClientRect();
  
  document.documentElement.style.cursor = 'grabbing';

  let paddingLeft = parseFloat(getComputedStyle(selectedCard).paddingLeft);
  let paddingRight = parseFloat(getComputedStyle(selectedCard).paddingRight);

  dndStore.mousePos.x = e.pageX;
  dndStore.mousePos.y = e.pageY;

  dndStore.draggedItemId = id;

 
  const taskId = id;
  const task = assetStore.getTasks.find(item => item.id === taskId );
  assetStore.selectTask(task);

  dndStore.ghostCardStyle.width = selectedCard.clientWidth - paddingLeft - paddingRight;
  dndStore.ghostCardStyle.cursorDistance.x = e.pageX - cardRect.x;
  dndStore.ghostCardStyle.cursorDistance.y = e.pageY - cardRect.y;
  
  dndStore.setGhostCardStyle(e, true)
  updateUI();
};

// done
const onDrag = (e) => {
    handleMouseMove(e);
    dndStore.onDrag(e);
};

const updateUI = () => {
  let dragX = dndStore.mousePos.x,
      dragY = dndStore.mousePos.y;

  if (dndStore.draggedItemId === null || dndStore.ghostCardStyle.leaving) return;

  if (!dragX && !dragY) {
    return requestAnimationFrame(updateUI);
  }

  dndStore.setGhostCardStyle(true, true)

  let isOverlapping;
  let newHoveredColumnId = null;
  let newHoveredCardIndex = -1;

  for (let column of statusColumns) {
    let columnEl = targetRefs.value[column.id];
    isOverlapping = dndStore.checkOverlap(
      { x: dragX, y: dragY },
      columnEl.getBoundingClientRect()
    );

    if (isOverlapping) {
      newHoveredColumnId = column.id;
      
      // Find the insertion position within the column
      const cardsContainer = columnEl.querySelector('.column-cards-container');
      if (cardsContainer) {
        const cardElements = Array.from(cardsContainer.children).filter(el => 
          el.classList.contains('column-card') && !el.classList.contains('dragging')
        );
        
        newHoveredCardIndex = findInsertionIndex(cardElements, dragY);
      }
      
      dndStore.itemOverlappedId = column.id;
      break;
    }
  }

  // Update hover state
  hoveredColumnId.value = newHoveredColumnId;
  hoveredCardIndex.value = newHoveredCardIndex;

  if (!isOverlapping) {
    return requestAnimationFrame(updateUI);
  }
  
  putCardInColumn();
  return requestAnimationFrame(updateUI);
};

const setStatus = async () => {
  const taskId = dndStore.draggedItemId;
  const status = statusColumns.find(item => item.id === dndStore.itemOverlappedId);
  
  try {
    const projectPath = projectStore.activeProject?.uri;
    if (projectPath && taskId && status) {
      await TaskService.ChangeStatus(projectPath, taskId, status.id);
      
      // The card position has already been updated in putCardInColumn
      // No need to update local data again here since putCardInColumn handles positioning
    }
  } catch (error) {
    console.error('Error changing task status:', error);
    
    // On error, reload the tasks to ensure consistency
    await loadAssetTasks();
  }
};

// done
const onDragStop = (cardEl) => {
  if (dndStore.draggedItemId === null) return;
  document.documentElement.style.cursor = 'default';

  let cardRect = cardEl.getBoundingClientRect();

  setStatus();

  // Reset hover state
  hoveredColumnId.value = null;
  hoveredCardIndex.value = -1;

  setTimeout(() => {
    dndStore.resetValues();
  }, 100);

  dndStore.ghostCardStyle.leaving = true;
  let xOffset = cardRect.x - dndStore.ghostCardStyle.pos.x;
  let yOffset = cardRect.y - dndStore.ghostCardStyle.pos.y;
  dndStore.ghostCardStyle.transform = `scale(1) translate(${xOffset}px, ${yOffset}px)`;
};

const putCardInColumn = () => {
  let draggedCard = cards.value.find(card => card.id === dndStore.draggedItemId);
  if (!draggedCard) return;

  const targetColumnId = dndStore.itemOverlappedId;
  const insertIndex = hoveredCardIndex.value;
  
  // Remove the card from its current position
  const currentIndex = cards.value.findIndex(card => card.id === dndStore.draggedItemId);
  if (currentIndex !== -1) {
    cards.value.splice(currentIndex, 1);
  }

  // Update the card's status
  draggedCard.status_id = targetColumnId;

  // Find the target column's cards
  const targetColumnCards = cards.value.filter(card => card.status_id === targetColumnId);
  
  // Calculate the global insertion index
  let globalInsertIndex;
  if (insertIndex >= targetColumnCards.length) {
    // Insert at the end of the column
    const lastCardOfColumn = targetColumnCards[targetColumnCards.length - 1];
    if (lastCardOfColumn) {
      globalInsertIndex = cards.value.findIndex(card => card.id === lastCardOfColumn.id) + 1;
    } else {
      // Column is empty, find where to insert based on status order
      globalInsertIndex = findGlobalInsertIndexForEmptyColumn(targetColumnId);
    }
  } else if (insertIndex === 0) {
    // Insert at the beginning of the column
    const firstCardOfColumn = targetColumnCards[0];
    if (firstCardOfColumn) {
      globalInsertIndex = cards.value.findIndex(card => card.id === firstCardOfColumn.id);
    } else {
      globalInsertIndex = findGlobalInsertIndexForEmptyColumn(targetColumnId);
    }
  } else {
    // Insert between cards
    const targetCard = targetColumnCards[insertIndex];
    if (targetCard) {
      globalInsertIndex = cards.value.findIndex(card => card.id === targetCard.id);
    } else {
      globalInsertIndex = cards.value.length;
    }
  }

  // Insert the card at the calculated position
  cards.value.splice(globalInsertIndex, 0, draggedCard);
};

const findGlobalInsertIndexForEmptyColumn = (targetColumnId) => {
  // Find the position where this column should be based on status order
  const statusOrder = statuses.value;
  const targetStatusIndex = statusOrder.findIndex(status => status.id === targetColumnId);
  
  // Find the first card that belongs to a status that comes after the target status
  for (let i = targetStatusIndex + 1; i < statusOrder.length; i++) {
    const nextStatusId = statusOrder[i].id;
    const nextStatusCardIndex = cards.value.findIndex(card => card.status_id === nextStatusId);
    if (nextStatusCardIndex !== -1) {
      return nextStatusCardIndex;
    }
  }
  
  // If no cards found in later statuses, insert at the end
  return cards.value.length;
};


onMounted(async () => {
  // emitter.emit('refresh-browser')
  await loadAssetTasks();
  // Ensure filtered cards are initialized even if no filters are active
  if (filteredCards.value.length === 0 && cards.value.length > 0) {
    await updateFilteredCards();
  }
});

// Watch for changes in filter-related properties and update filtered cards
watch(() => cards.value, async () => {
  await updateFilteredCards();
}, { deep: true });

watch(() => commonStore.taskFilters, async () => {
  await updateFilteredCards();
}, { deep: true });

watch(() => commonStore.viewSearchQuery, async () => {
  await updateFilteredCards();
});

watch(() => commonStore.workspaceSearchQuery, async () => {
  await updateFilteredCards();
});

watch(() => commonStore.hasAssignees, async () => {
  await updateFilteredCards();
});

watch(() => commonStore.noAssignees, async () => {
  await updateFilteredCards();
});

onUnmounted(() => {
});


</script>

<style scoped>
@import "@/assets/desktop.css";


.kanban-root {
  height: 100%;
  overflow: hidden;
  overflow-x: scroll;
  width: 100%;
  display: flex;
  justify-content: flex-start;
  /* background-color: rebeccapurple; */
  padding: 0.5rem 0;
  box-sizing: border-box;
  gap: .8rem;
  margin-bottom: .2rem;
  /* position: relative; */
}

.kanban-root::-webkit-scrollbar {
  height: 4px;
  /* margin-left: 1rem; */
}

.kanban-root::-webkit-scrollbar-thumb {
  border-radius: 10px;
    background-color: var(--white);
}

.kanban-root::-webkit-scrollbar-track {
  border-radius: 10px;
  background-color: rgba(0, 0, 0, 0);
}

.kanban-container {
  height: 100%;
  width: 100%;
  /* overflow: hidden;
  overflow-y: scroll; */
  display: flex;
  justify-content: flex-start;
  align-items: flex-start;
  padding: 0rem .5rem;
  box-sizing: border-box;
  gap: .8rem;
  /* background-color: chocolate; */
  min-width: max-content;
  position: relative;
}

.column {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  border-radius: var(--normal-radius);
  padding: 5px;
  box-sizing: border-box;
  height: max-content;
  height: 100%;
  color: var(--white);
  align-items: center;
  background-color: #0000002c;
  outline: var(--transparent-line);
  outline-offset: -1.5px;
  transition: all .3s ease-out;
  border-radius: var(--large-radius);
}

.column-minimized{
  width: min-content;
  flex: 0;
  transition: all .3s ease-out;
}

.column-header {
  display: flex;
  width: 100%;
  height: 40px;
  /* width: min-content; */
  padding: .3rem .8rem;
  box-sizing: border-box;
  overflow: hidden;
  /* background-color: red; */
  align-items: center;
  justify-content: space-between;
  justify-content: flex-start;
  border-radius: var(--normal-radius);
  outline-offset: -1px;
  gap: .5rem;
}

.column-header-minimized{
  height: 40px;
}

.column-header-text{
  font-weight: 400;
  font-size: 14px;
  color: var(--white);
}

.column-header-meta{
  display: flex;
  width: 100%;
  /* background-color: royalblue; */
  justify-content: space-between;
}

.column-header-count{
  font-weight: 400;
  font-size: 14px;
  color: var(--white);
}

[data-theme="dark"] .column-header-text .column-header-count{
  font-weight: 300;
}


.column-header-icon{
  width: 24px;
  height: 24px;
}

.column-cards-container{
  width: min-content;
  min-width: 100px;
  align-items: center;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: .5rem;
  overflow: hidden;
  overflow-y: scroll;
  padding: .2rem;
  align-items: center;
  box-sizing: border-box;
  transition: all .3s ease-out;
  border-radius: var(--large-radius);
}

.column-cards-container-minimized{
  min-width: 100px;
}


.column-cards-container::-webkit-scrollbar {
  width: 4px;
  /* margin-left: 1rem; */
}

.column-cards-container::-webkit-scrollbar-thumb {
  border-radius: 10px;
    background-color: var(--white);
}

.column-cards-container::-webkit-scrollbar-track {
  border-radius: 10px;
  background-color: rgba(0, 0, 0, 0);
}

.column-card {
  /* background: tomato; */
  /* padding: 10px; */
  /* margin: 10px 0; */
  border-radius: 8px;
  cursor: grab;
}
/* .column-card:active {
} */

.column-card.dragging {
  color: transparent;
  background: none;
  box-shadow: none;
  user-select: none;
  -moz-user-select: none;
  -webkit-user-select: none;
  cursor: grabbing;
  opacity: .2;
}

.card-insert-indicator {
  width: 100%;
  height: 3px;
  background-color: var(--white);
  border-radius: 2px;
  margin: 2px 0;
  opacity: 0.8;
  /* box-shadow: 0 0 8px rgba(255, 255, 255, 0.6); */
}

#ghost-card {
  position: fixed;
  user-select: none;
  pointer-events: none;
  opacity: 0;
  transform-origin: center;
  transform: scale(1) rotate(0);
  box-shadow: 0 1px 0 rgba(9, 30, 66, 0.25);
  transition: transform 0.04s ease-in-out;
}
#ghost-card.animate {
  transition: box-shadow 0.1s ease-in-out;
  transition: transform 0.05s ease-in-out;
}
#ghost-card.active {
  outline: 1.5px solid rgb(255, 255, 255);
  outline-offset: -1.5px;
  opacity: 1;
  /* box-shadow: 0 12px 24px -6px rgba(9, 30, 66, 0.25), 0 0 0 1px rgba(9, 30, 66, 0.08); */
}
#ghost-card.leaving {
  transition: all 0.1s ease;
  box-shadow: 0 1px 0 rgba(9, 30, 66, 0.25);
}
</style>


