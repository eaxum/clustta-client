<template>
  <div ref="virtuaItemRef" class="virtua-item" >
    <div class="virtua-item-header drop-zone" :data-id="child.id" :data-other="JSON.stringify(child)" :id="child.id"
      :class="{ 'drop-zone-hovered': isHovered }">
      <Collection ref="collectionItemRef" v-if="child.type == 'collection'" @toggleEditMode="toggleEditMode" v-right-click="openCollectionMenu"
        :isGhost="isGhost" :collection="child" :index="index" :loadingCollectionState="loadingCollectionState" />
      <Asset v-if="child.type == 'asset'" @toggleEditMode="toggleEditMode" v-right-click="openAssetMenu" :asset="child" :index="index" :loadingAssetState="loadingAssetState" />
      <Collection ref="collectionItemRef" v-if="child.type == 'untracked_collection'" @toggleEditMode="toggleEditMode" v-right-click="openUntrackedItemMenu"
        :isUntracked="true" :collection="child" :index="index" />
      <Asset v-if="child.type == 'untracked_asset'" @toggleEditMode="toggleEditMode" v-right-click="openUntrackedItemMenu" :isUntracked="true"
        :asset="child" :index="index" />
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue';

import { useMenu } from '@/stores/menu';
import { useDndStore } from '@/stores/dnd';
import { useAssetStore } from '@/stores/assets';
import { useStageStore } from '@/stores/stages';
import { useCollectionStore } from '@/stores/collections';
import { useScrollStore } from '@/stores/scroll';
import { useUntrackedItemStore } from '@/stores/untracked';
import { useProjectStore } from '@/stores/projects';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { AssetService, CollectionService } from '@/services';
import emitter from '@/lib/mitt';

const menu = useMenu();
const stage = useStageStore();
const dndStore = useDndStore();
const assetStore = useAssetStore();
const scrollStore = useScrollStore();
const collectionStore = useCollectionStore();
const untrackedItemStore = useUntrackedItemStore();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();

// refs
const loadingDelay = 250;
const virtuaItemRef = ref(null);
const virtuaChildrenRef = ref(null);
const collectionItemRef = ref(null);
const collectionChildren = ref(null);
const loadingAssetState = ref(false);
const loadingCollectionState = ref(false);

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
  const collection = props.child;
  stage.markedCollections = [id];
  collectionStore.selectCollection(collection);
  menu.showContextMenu(event, 'collectionMenu', true);
};

const openAssetMenu = (event) => {
  const id = props.child.id;
  const asset = props.child;
  assetStore.selectAsset(asset);
  stage.markedAssets = [id];
  menu.showContextMenu(event, 'assetMenu', true);
};

const openUntrackedItemMenu = (event) => {
  const id = props.child.id;
  stage.markedCollections = [id];
  untrackedItemStore.selectUntrackedItem(props.child)
  menu.showContextMenu(event, 'untrackedItemMenu', true);
};

// computed
const isExpanded = computed(() => {
  return props.child.id in stage.expandedCollections;
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

const handleToggle = async () => {
  await props.onToggle(props.index);
  if (stage.firstSelectedItemId === props.child.id) {
    // handleScrollPosition();
  }
  await updateItemHeight();
};

const isHovered = computed(() => { return dndStore.targetItemId === props.child.id })

// Watch for changes to the computed property
watch(() => collectionItemRef.value?.collectionData, (newValue) => {
  collectionChildren.value = newValue
});

const isEditing = ref(false);

const toggleEditMode = (value) => {
  isEditing.value = value;
};

// Loads the file status state for an asset.
const loadAssetState = async () => {
  const asset = props.child;
  
  if (asset.type !== 'asset' || asset.is_link) return;

  const loadingTimer = setTimeout(() => {
    loadingAssetState.value = true;
  }, loadingDelay);
  
  try {
    const fileStatus = await AssetService.GetAssetState(
      projectStore.activeProject.uri,
      asset.id
    );

    props.child.file_status = fileStatus;
  } catch (error) {
    console.error(`Error loading asset state for ${asset.id}:`, error);
    asset.file_status = 'rebuildable';
  } finally {
    clearTimeout(loadingTimer);
    loadingAssetState.value = false;
  }
};

// Loads the state flags for a collection.
const loadCollectionState = async () => {
  const collection = props.child;
  
  if (collection.type !== 'collection') return;

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

    props.child.collectionStateFlags = flags;
  } catch (error) {
    console.error(`Error loading collection state for ${collection.id}:`, error);
    props.child.collectionStateFlags = {
      has_untracked: false,
      has_modified: false,
      has_outdated: false,
      has_rebuildable: false
    };
  } finally {
    clearTimeout(loadingTimer);
    loadingCollectionState.value = false;
  }
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
        if(type.includes('collection')  && collectionChildren.value.length && !(item.id in stage.expandedCollections)){
          stage.expandCollection(item);
          const firstChildId = collectionChildren.value[0].id;
          handleToggle();
          stage.markedItems = [firstChildId]
          stage.firstSelectedItemId = firstChildId;
        } else {

        }
      } else {
        
        let parent; 
        
        const allCollections = collectionStore.getCollections;
        const alluntrackedFolders = projectStore.untrackedFolders;
        const allItems = [ ...allCollections, ...alluntrackedFolders];

        if(type === 'collection'){
          parent = allCollections.find((collection) => collection.id === item.parent_id);
        } else if ( type === 'asset'){
          parent = allCollections.find((collection) => collection.id === item.collection_id);
        } else {
          const collectionPath = item.collection_path;
          const parentCollection = allCollections.find((collection) => collection.collection_path === collectionPath);
          const parentUntrackedCollection = alluntrackedFolders.find((collection) => collection.item_path === collectionPath);
          parent = parentCollection ? parentCollection : parentUntrackedCollection;
        }

        if(parent){
          stage.expandCollection(parent);
          handleToggle();
          stage.markedItems = [parent.id]
          stage.firstSelectedItemId = parent.id;
        }

      }

    }
  }
};

onMounted(async () => {
  if (props.child.collection_type_id) {
    collectionChildren.value = collectionItemRef.value.collectionData
  }
  
  if (props.child.type === 'collection') {
    await loadCollectionState();
  }
  
  if (props.child.type === 'asset') {
    await loadAssetState();
  }
  
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

.collection-drop-zone-hovered {
  width: 100%;
  height: 100%;
  position: absolute;
  opacity: .3;
  background-color: var(--drop-hover);
}


.virtua-item-header {
  overflow: hidden;
  box-sizing: border-box;
}

.virtua-item-children {
  padding-left: 30px;
  box-sizing: border-box;
  overflow: hidden;
  /* background-color: red; */
  /* padding-right: 1px; */
}
</style>