<template>
  <div ref="virtuaItemRef" class="virtua-item" :style="{ '--depth': depth, '--lightness': `${65 + (depth * 5)}%` }">
    <div class="virtua-item-header drop-zone" :data-id="child.id" :data-other="JSON.stringify(child)" :id="child.id"
      :style="{ height: `${itemHeight}px` }" :class="{ 'drop-zone-hovered': isHovered }">
      <Collection ref="entityItemRef" v-if="child.type == 'entity'" @toggleEditMode="toggleEditMode"
        v-right-click="openCollectionMenu" :hasChildren="hasChildren" :loadingChildren="loadingChildren" :isGhost="isGhost" @toggle="handleToggle"
        :entity="child" :index="index" :entityChildren="entityChildren" :loadingCollectionState="loadingCollectionState && child.type === 'entity'" />
      <Asset v-if="child.type == 'task'" @refreshData="emit('refreshData')" @toggleEditMode="toggleEditMode"
        v-right-click="openAssetMenu" :task="child" :index="index" :loadingAssetState="loadingAssetState && child.type === 'task'" />
      <Collection ref="entityItemRef" v-if="child.type == 'untracked_entity'" @toggleEditMode="toggleEditMode"
        v-right-click="openUntrackedItemMenu" :hasChildren="hasChildren" :loadingChildren="loadingChildren" :isUntracked="true" :entity="child"
        @toggle="handleToggle" :index="index" :entityChildren="entityChildren" :loadingCollectionState="loadingCollectionState && child.type === 'entity'"/>
      <Asset v-if="child.type == 'untracked_task'" @toggleEditMode="toggleEditMode"
        v-right-click="openUntrackedItemMenu" :isUntracked="true" :task="child" :index="index" />
    </div>
    <template v-if="isExpanded">
      <ListSkeleton v-if="loadingChildren" :itemHeight="commonStore.listItemHeight" :height="virtuaIndentHeight" :depth="depth + 1" />
      <div ref="virtuaChildrenRef" v-else-if="entityChildren.length" class="virtua-item-children">
        <div class="indent-guide" :style="{ height: `${indentHeight}px` }"
          :class="{ 'indent-guide-selected': stage.markedItems.length === 1 && stage.firstSelectedItemId === child.id }">
        </div>
        <VirtuaList @refreshData="loadEntityChildren" @shiftParents="handleToggle" :items="entityChildren"
          :containerHeight="scrollStore.scrollRootHeight || 0" :depth="depth + 1" :parentOffset="totalOffset"
          :itemHeight="commonStore.listItemHeight" />
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue';
import emitter from '@/lib/mitt';

import { useMenu } from '@/stores/menu';
import { useDndStore } from '@/stores/dnd';
import { useAssetStore } from '@/stores/assets';
import { useStageStore } from '@/stores/stages';
import { useCollectionStore } from '@/stores/collections';
import { useScrollStore } from '@/stores/scroll';
import { useUntrackedItemStore } from '@/stores/untracked';
import { useProjectStore } from '@/stores/projects';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useCommonStore } from '@/stores/common';
import { CollectionService, AssetService } from '@/../bindings/clustta/services';
import { useIconStore } from '@/stores/icons';

const menu = useMenu();
const stage = useStageStore();
const dndStore = useDndStore();
const assetStore = useAssetStore();
const scrollStore = useScrollStore();
const iconStore = useIconStore();
const collectionStore = useCollectionStore();
const untrackedItemStore = useUntrackedItemStore();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const commonStore = useCommonStore();

// refs
const virtuaItemRef = ref(null);
const virtuaChildrenRef = ref(null);
const entityItemRef = ref(null);
const entityChildren = ref([]);
const hasChildren = ref(false);
const loadingChildren = ref(true);
const loadingAssetState = ref(false);
const loadingCollectionState = ref(false);
const indentPadding = ref(4);

const emit = defineEmits(['refreshData']);

// components
import VirtuaList from '@/instances/common/components/VirtuaList.vue';
import Asset from '@/instances/desktop/blocks/Asset.vue'
import Collection from '@/instances/desktop/blocks/Collection.vue';
import ListSkeleton from '@/instances/desktop/components/ListSkeleton.vue';

// props
const props = defineProps({
  index: { type: Number, required: true },
  isGhost: { type: Boolean, default: false },
  child: { type: Object, default: null },
  itemHeight: { type: Number, default: 60 },
  totalHeight: { type: Number, default: 60 },
  depth: { type: Number, default: 0 },
  onHeightChange: { type: Function, default: () => { } },
  getItemPosition: { type: Function, required: true },
  parentScrollContainer: { type: Object, default: null },
  parentOffset: { type: Number, default: 0 },
  offsetY: { type: Number, default: 0 },
});

// menu methods
const openCollectionMenu = (event) => {
  const id = props.child.id;
  const entity = props.child;
  stage.markedEntities = [id];
  collectionStore.selectCollection(entity);
  menu.showContextMenu(event, 'collectionMenu', true);
};

const openAssetMenu = (event) => {
  const id = props.child.id;
  const task = props.child;
  assetStore.selectAsset(task);
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
const filtersActive = computed(() => {
	const assigneeFilters = commonStore.hasAssignees || commonStore.noAssignees;
	const entityFilters = commonStore.entityFilters.length > 0;
	const taskFilters = commonStore.taskFilters.length > 0;
  let generalFilter
	return assigneeFilters || entityFilters || taskFilters || generalFilter
});

const isExpanded = computed(() => {
  return props.child.id in stage.expandedEntities;
});

const updateItemHeight = async () => {
  if (!virtuaItemRef.value) return;
  await nextTick();
  const height = virtuaItemRef.value?.offsetHeight;
  props.onHeightChange(props.index, height);
};

const handleToggle = async () => {
  await updateItemHeight();
};

const indentHeight = computed(() => {
  const itemHeight = stage.expandedEntities[props.child.id]["height"];
  entityChildren.value;
  const height = virtuaChildrenRef.value?.getBoundingClientRect().height;
  return height - indentPadding.value
});

const virtuaIndentHeight = computed(() => {
  const rootHeight = stage.expandedEntities[props.child.id]["height"];
  return rootHeight - props.itemHeight - indentPadding.value
});

// Calculate the total offset for nested scroll
const totalOffset = computed(() => {
  const itemPosition = props.getItemPosition(props.index);
  return props.offsetY + itemPosition;
});

const isHovered = computed(() => { return dndStore.targetItemId === props.child.id })

const isEditing = ref(false);

const toggleEditMode = (value) => {
  isEditing.value = value;
};

const isItemInFocus = computed(() => {
  return stage.markedItems.length === 1 && stage.firstSelectedItemId === props.child.id && !dndStore.draggedItem
});

const handleKeyArrowKeys = async (event) => {
  if (modals.activeModal) {
    return
  }
  if (isItemInFocus.value && !isEditing.value) {
    if (event.key === 'ArrowLeft' || event.key === 'ArrowRight') {
      event.preventDefault();

      const item = props.child;
      const type = item.type;

      if (event.key === 'ArrowRight') {

        if (type.includes('entity') && entityChildren.value.length && !(item.id in stage.expandedEntities)) {
          if (type === 'untracked_entity') {
            stage.expandEntity(item, true);
          } else {
            stage.expandEntity(item);
          }
          const firstChildId = entityChildren.value[0].id;
          handleToggle();
          stage.markedItems = [firstChildId]
          stage.firstSelectedItemId = firstChildId;
        } else {

        }
      } else {

        let parent;
        const allEntities = await       CollectionService.GetCollections(projectStore.activeProject.uri)
        const alluntrackedFolders = projectStore.untrackedFolders;

        // const allItems = [...allEntities, ...alluntrackedFolders];
        const allItems = dndStore.allViewItems;

        if (type === 'entity') {
          parent = allItems.find((entity) => entity.id === item.parent_id);
        } else if (type === 'task') {
          parent = allItems.find((entity) => entity.id === item.entity_id);
        } else {
          const entityPath = item.entity_path;
          const parentEntity = allItems.find((entity) => entity.entity_path === entityPath);
          const parentUntrackedEntity = alluntrackedFolders.find((entity) => entity.item_path === entityPath);
          parent = parentEntity ? parentEntity : parentUntrackedEntity;
        }

        if (parent) {
          if (type === 'untracked_entity') {
            stage.expandEntity(parent, true);
          } else {
            stage.expandEntity(parent);
          }
          handleToggle();
          stage.markedItems = [parent.id]
          stage.firstSelectedItemId = parent.id;
        }

      }

    }
  }
};

const loadAssetState = async () => {

  const task = props.child;
  
  if (task.type !== 'task' || task.is_link) return;

  loadingAssetState.value = true;
  
  try {
    const projectStore = useProjectStore();
    const fileStatus = await AssetService.GetAssetState(
      projectStore.activeProject.uri,
      task.id
    );

    props.child.file_status = fileStatus;

  } catch (error) {
    console.error(`Error loading asset state for ${task.id}:`, error);
    task.file_status = 'rebuildable';
  } finally {
    loadingAssetState.value = false;
  }
};

const loadCollectionState = async () => {

  const entity = props.child;
  
  if (entity.type !== 'entity') return;
  
  loadingCollectionState.value = true;
  
  try {
    const projectStore = useProjectStore();
    const flags = await CollectionService.GetCollectionStateFlags(
      projectStore.activeProject.uri,
      entity.id,
      projectStore.activeProject.working_directory,
      projectStore.activeProject.ignore_list
    );

    // Store the flags on the entity object
    props.child.collectionStateFlags = flags;

  } catch (error) {
    console.error(`Error loading collection state for ${entity.id}:`, error);
    // Reset flags on error
    props.child.collectionStateFlags = {
      has_untracked: false,
      has_modified: false,
      has_outdated: false,
      has_rebuildable: false
    };
  } finally {
    loadingCollectionState.value = false;
  }
};

const loadEntityChildren = async () => {
  if(stage.operationActive){
    loadingChildren.value = true;
    return
  }
  if (props.child.type == "entity" || props.child.type == 'untracked_entity') {
    let isUntracked = props.child.type == 'untracked_entity'
    let project = projectStore.activeProject
    let children = await CollectionService.GetCollectionChildren(project.uri, props.child.id, project.working_directory, props.child.file_path, project.ignore_list, isUntracked)
    await assetStore.processAssetsIconsAndPreviews(children.tasks);
    await assetStore.processUntrackedAssetsIcons(children.untracked_tasks);

    let childrenEntities = filtersActive.value ? await collectionStore.filterCollections(children.entities) : children.entities ;
		let childrenTasks = filtersActive.value ? await assetStore.filterAssets(children.tasks) : children.tasks ;

    entityChildren.value = [...childrenEntities, ...children.untracked_entities, ...childrenTasks,  ...children.untracked_tasks]
    hasChildren.value = entityChildren.value.length > 0
    
    // If this entity has no children, collapse it
    if (!hasChildren.value && props.child.id in stage.expandedEntities) {
      stage.expandEntity(props.child, isUntracked);
    }
  }
  loadingChildren.value = false;
};

const handleUpdateChildren = (eventData) => {
  const { itemId, property, value, updates } = eventData;
  
  // Find the item in entityChildren
  const itemIndex = entityChildren.value.findIndex(item => item.id === itemId);
  if (itemIndex !== -1) {
    // Handle single property update (backward compatibility)
    if (property && value !== undefined) {
      entityChildren.value[itemIndex][property] = value;
    }
    // Handle multiple property updates
    if (updates && Array.isArray(updates)) {
      updates.forEach(update => {
        entityChildren.value[itemIndex][update.property] = update.value;
      });
    }
  };
  
  emitter.emit('get-project-data');
  collectionStore.loadCollectionStateFlags();
};

// Watch for stage.operationActive changes
watch(() => stage.operationActive, (newValue, oldValue) => {
  if (oldValue && !newValue && loadingChildren.value) {
    loadEntityChildren();
  }
});

onMounted(async () => {
  if (props.child.type === 'entity' || props.child.type === 'untracked_entity') {
    await loadEntityChildren();
  }
  
  if (props.child.type === 'entity') {
    await loadCollectionState();
  }
  
  if (props.child.type === 'task') {
    await loadAssetState();
  }
  
  window.addEventListener('keydown', handleKeyArrowKeys);
  emitter.on('update-children', handleUpdateChildren);
});

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeyArrowKeys);
  emitter.off('update-children', handleUpdateChildren);
});



</script>

<style scoped>
.virtua-item-children-wrapper{
  width: 100%;
  background-color: indigo;
  padding-right: 20px;
  overflow: hidden;
}

.virtua-item {
  transition: all 0.3s ease;
  overflow: hidden;
  margin: 0 auto;
  cursor: pointer;
  box-sizing: border-box;
  /* background-color: crimson; */
}

.indent-guide {
  position: absolute;
  width: 100%;
  height: 300px;
  /* background-color: forestgreen; */
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
}

.virtua-item-children {
  padding-left: 30px;
  box-sizing: border-box;
  overflow: hidden;
  /* padding-right: 1px; */
  width: 100%;
  /* width: min-content; */
  /* background-color: royalblue; */
  /* padding: .5rem; */
}

.debugger{
  width: 200px;
  height: 200px;
  height: min-content;
  display: flex;
  width: min-content;
  padding: .5rem;
  position: fixed;
  top: 0;
  left: 0;
  background-color: forestgreen;
}
</style>