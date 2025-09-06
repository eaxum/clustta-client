<template>
  <div ref="virtuaItemRef" class="virtua-item" >
    <div class="virtua-item-header drop-zone" :data-id="child.id" :data-other="JSON.stringify(child)" :id="child.id"
      :class="{ 'drop-zone-hovered': isHovered }">
      <Collection ref="entityItemRef" v-if="child.type == 'entity'" @toggleEditMode="toggleEditMode" v-right-click="openCollectionMenu"
        :isGhost="isGhost" :entity="child" :index="index" />
      <Asset v-if="child.type == 'task'" @toggleEditMode="toggleEditMode" v-right-click="openAssetMenu" :task="child" :index="index" />
      <Collection ref="entityItemRef" v-if="child.type == 'untracked_entity'" @toggleEditMode="toggleEditMode" v-right-click="openUntrackedItemMenu"
        :isUntracked="true" :entity="child" :index="index" />
      <Asset v-if="child.type == 'untracked_task'" @toggleEditMode="toggleEditMode" v-right-click="openUntrackedItemMenu" :isUntracked="true"
        :task="child" :index="index" />
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue';

import { useMenu } from '@/stores/menu';
import { useDndStore } from '@/stores/dnd';
import { useTaskStore } from '@/stores/task';
import { useStageStore } from '@/stores/stages';
import { useEntityStore } from '@/stores/entity';
import { useScrollStore } from '@/stores/scroll';
import { useUntrackedItemStore } from '@/stores/untracked';
import { useProjectStore } from '@/stores/projects';
import { useDesktopModalStore } from '@/stores/desktopModals';

const menu = useMenu();
const stage = useStageStore();
const dndStore = useDndStore();
const taskStore = useTaskStore();
const scrollStore = useScrollStore();
const entityStore = useEntityStore();
const untrackedItemStore = useUntrackedItemStore();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();

// refs
const virtuaItemRef = ref(null);
const virtuaChildrenRef = ref(null);
const entityItemRef = ref(null);
const entityChildren = ref(null);

// components
import Asset from '@/instances/desktop/blocks/Asset.vue'
import Collection from '@/instances/desktop/blocks/Collection.vue'

// props
const props = defineProps({
  index: { type: Number, required: true },
  isGhost: { type: Boolean, default: false },
  child: { type: Object, default: null }
});

// menu methods
const openCollectionMenu = (event) => {
  const id = props.child.id;
  const entity = props.child;
  stage.markedEntities = [id];
  entityStore.selectEntity(entity);
  menu.showContextMenu(event, 'collectionMenu', true);
};

const openAssetMenu = (event) => {
  const id = props.child.id;
  const task = props.child;
  taskStore.selectTask(task);
  stage.markedTasks = [id];
  menu.showContextMenu(event, 'assetMenu', true);
};

const openUntrackedItemMenu = (event) => {
  const id = props.child.id;
  stage.markedEntities = [id];
  untrackedItemStore.selectUntrackedItem(props.child)
  menu.showContextMenu(event, 'untrackedItemMenu', true);
};

// computed
const isExpanded = computed(() => {
  return props.child.id in stage.expandedEntities;
});

const calculateRelativePosition = () => {
  const itemRect = virtuaItemRef.value.getBoundingClientRect();
  const scrollRect = scrollStore.scrollRoot.getBoundingClientRect();
  return itemRect.top - scrollRect.top + scrollStore.scrollTop;
};

const updateItemHeight = async () => {
  if (!virtuaItemRef.value) return;
  await nextTick();
  const height = virtuaItemRef.value.offsetHeight;
  props.onHeightChange(props.index, height);
};

const handleScrollPosition = () => {
  if (!isExpanded.value) return;

  // Get full position including parent offsets
  const relativePosition = calculateRelativePosition();

  // Request scroll with calculated position
  scrollStore.requestScroll(relativePosition);
};

const handleToggle = async () => {
  await props.onToggle(props.index);
  if (stage.firstSelectedItemId === props.child.id) {
    // handleScrollPosition();
  }
  await updateItemHeight();
};

// Compute if current item has children
const hasChildren = computed(() => {
  return entityChildren.value?.length > 0;
});

const indentHeight = computed(() => {
  const itemHeight = stage.expandedEntities[props.child.id]
  entityChildren.value;
  const height = virtuaChildrenRef.value?.getBoundingClientRect().height;
  return height - 10
});

// Calculate the total offset for nested scroll
const totalOffset = computed(() => {
  const itemPosition = props.getItemPosition(props.index);
  return props.offsetY + itemPosition;
});

const isHovered = computed(() => { return dndStore.targetItemId === props.child.id })

// Watch for changes to the computed property
watch(() => entityItemRef.value?.entityData, (newValue) => {
  entityChildren.value = newValue
});

const isEditing = ref(false);

const toggleEditMode = (value) => {
  isEditing.value = value;
};

const isItemInFocus = computed(() => {
  return stage.markedItems.length === 1 && stage.firstSelectedItemId === props.child.id && !dndStore.draggedItem
});

const handleKeyArrowKeys = (event) => {
  if(modals.activeModal){
    return
  }
  if(isItemInFocus.value && !isEditing.value){
    if (event.key === 'ArrowLeft' || event.key === 'ArrowRight') {
      event.preventDefault();
      
      const item = props.child;
      const type = item.type;
      
      if(event.key === 'ArrowRight'){
        if(type.includes('entity')  && entityChildren.value.length && !(item.id in stage.expandedEntities)){
          stage.expandEntity(item);
          const firstChildId = entityChildren.value[0].id;
          handleToggle();
          stage.markedItems = [firstChildId]
          stage.firstSelectedItemId = firstChildId;
        } else {

        }
      } else {
        
        let parent; 
        
        const allEntities = entityStore.getEntities;
        const alluntrackedFolders = projectStore.untrackedFolders;
        const allItems = [ ...allEntities, ...alluntrackedFolders];

        if(type === 'entity'){
          parent = allEntities.find((entity) => entity.id === item.parent_id);
        } else if ( type === 'task'){
          parent = allEntities.find((entity) => entity.id === item.entity_id);
        } else {
          const entityPath = item.entity_path;
          const parentEntity = allEntities.find((entity) => entity.entity_path === entityPath);
          const parentUntrackedEntity = alluntrackedFolders.find((entity) => entity.item_path === entityPath);
          parent = parentEntity ? parentEntity : parentUntrackedEntity;
        }

        if(parent){
          stage.expandEntity(parent);
          handleToggle();
          stage.markedItems = [parent.id]
          stage.firstSelectedItemId = parent.id;
        }

      }

    }
  }
};

onMounted(() => {
  if (props.child.entity_type_id) {
    entityChildren.value = entityItemRef.value.entityData
  };
  window.addEventListener('keydown', handleKeyArrowKeys);
});

onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleKeyArrowKeys);
});



</script>

<style scoped>
.virtua-item {
  transition: all 0.3s ease;
  overflow: hidden;
  margin: 0 auto;
  cursor: pointer;
  border-radius: 8px;
  box-sizing: border-box;
  width: 100%;
}

.indent-guide {
  position: absolute;
  width: 100%;
  box-sizing: border-box;
  border-left: var(--transparent-line);
  left: 15px;
}

.indent-guide-hovered {
  border-left: var(--solid-line);
}

.virtua-item-header:hover+.virtua-item-children>.indent-guide {
  border-left: var(--solid-line);
}

.indent-guide-selected {
  border-left: var(--solid-line);
}

.entity-drop-zone-hovered {
  width: 100%;
  height: 100%;
  position: absolute;
  opacity: .3;
  background-color: var(--drop-hover);
}

.drop-zone-hovered {
  /* background-color: crimson; */
}

.virtua-item-header {
  overflow: hidden;
  box-sizing: border-box;
  /* background-color: hsl(330, 100%, var(--lightness)); */
  /* padding: 0 10px; */
  /* display: flex; */
  /* justify-content: space-between; */
  /* align-items: center; */
  /* height: 60px; */
  /* border-radius: 8px; */
  /* outline: 1px solid black; */
  /* outline-offset: -1px; */
  /* padding-bottom: 5px; */
}

.virtua-item-children {
  padding-left: 30px;
  box-sizing: border-box;
  overflow: hidden;
  /* background-color: red; */
  /* padding-right: 1px; */
}
</style>