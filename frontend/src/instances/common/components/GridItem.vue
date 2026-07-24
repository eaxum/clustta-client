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
import { getBrowserItemKey } from '@/lib/browserTree';
import emitter from '@/lib/mitt';

import { useMenu } from '@/stores/menu';
import { useDndStore } from '@/stores/dnd';
import { useAssetStore } from '@/stores/assets';
import { useBrowserTreeStore } from '@/stores/browserTree';
import { useStageStore } from '@/stores/stages';
import { useCollectionStore } from '@/stores/collections';
import { useScrollStore } from '@/stores/scroll';
import { useUntrackedItemStore } from '@/stores/untracked';
import { useProjectStore } from '@/stores/projects';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { AssetService, CollectionService } from '@/services';

const menu = useMenu();
const stage = useStageStore();
const dndStore = useDndStore();
const assetStore = useAssetStore();
const browserTreeStore = useBrowserTreeStore();
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
  child: { type: Object, default: null },
  isFilteredView: { type: Boolean, default: false }
});

// menu methods
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

const openCollectionMenu = (event) => {
  const id = props.child.id;
  if (tryOpenSelectionMenu(event, id)) return;
  const collection = props.child;
  stage.markedCollections = [id];
  collectionStore.selectCollection(collection);
  focusSingleItem(collection);
  menu.showContextMenu(event, 'collectionMenu', true);
};

const openAssetMenu = (event) => {
  const id = props.child.id;
  if (tryOpenSelectionMenu(event, id)) return;
  const asset = props.child;
  assetStore.selectAsset(asset);
  stage.markedAssets = [id];
  focusSingleItem(asset);
  menu.showContextMenu(event, 'assetMenu', true);
};

const openUntrackedItemMenu = (event) => {
  const id = props.child.id;
  if (tryOpenSelectionMenu(event, id)) return;
  stage.markedCollections = [id];
  untrackedItemStore.selectUntrackedItem(props.child)
  focusSingleItem(props.child);
  menu.showContextMenu(event, 'untrackedItemMenu', true);
};

// computed
const isExpanded = computed(() => {
  return props.child.id in stage.expandedCollections;
});

const itemKey = computed(() => getBrowserItemKey(props.child));

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

const emptyCollectionStateFlags = () => ({
  has_untracked: false,
  has_modified: false,
  has_outdated: false,
  has_fetchable: false
});

// Loads the file status state for an asset.
const loadAssetState = async (options = {}) => {
  const asset = props.child;
  const silent = options.silent === true;
  
  if (asset.type !== 'asset' || asset.is_link) return;
  if (props.isFilteredView) {
    loadingAssetState.value = false;
    return;
  }

  const loadingTimer = silent ? null : setTimeout(() => {
    loadingAssetState.value = true;
  }, loadingDelay);
  
  try {
    const fileStatus = await AssetService.GetAssetState(
      projectStore.activeProject.uri,
      asset.id
    );

    if (props.isFilteredView) return;
    browserTreeStore.patchItem(itemKey.value, { file_status: fileStatus });
  } catch (error) {
    if (props.isFilteredView) return;
    console.error(`Error loading asset state for ${asset.id}:`, error);
    browserTreeStore.patchItem(itemKey.value, { file_status: 'fetchable' });
  } finally {
    clearTimeout(loadingTimer);
    loadingAssetState.value = false;
  }
};

// Loads the state flags for a collection.
const loadCollectionState = async (options = {}) => {
  const collection = props.child;
  const silent = options.silent === true;
  
  if (collection.type !== 'collection') return;
  if (props.isFilteredView) {
    loadingCollectionState.value = false;
    browserTreeStore.patchItem(itemKey.value, {
      collectionStateFlags: emptyCollectionStateFlags()
    });
    return;
  }

  const loadingTimer = silent ? null : setTimeout(() => {
    loadingCollectionState.value = true;
  }, loadingDelay);

  try {
    const flags = await CollectionService.GetCollectionStateFlags(
      projectStore.activeProject.uri,
      collection.id,
      projectStore.activeProject.working_directory,
      projectStore.activeProject.ignore_list
    );

    if (props.isFilteredView) {
      browserTreeStore.patchItem(itemKey.value, {
        collectionStateFlags: emptyCollectionStateFlags()
      });
      return;
    }
    browserTreeStore.patchItem(itemKey.value, { collectionStateFlags: flags });
  } catch (error) {
    if (props.isFilteredView) return;
    console.error(`Error loading collection state for ${collection.id}:`, error);
    browserTreeStore.patchItem(itemKey.value, {
      collectionStateFlags: emptyCollectionStateFlags()
    });
  } finally {
    clearTimeout(loadingTimer);
    loadingCollectionState.value = false;
  }
};

watch(() => props.isFilteredView, async (filtered) => {
  loadingAssetState.value = false;
  loadingCollectionState.value = false;

  if (filtered) {
    if (props.child.type === 'collection') {
      browserTreeStore.patchItem(itemKey.value, {
        collectionStateFlags: emptyCollectionStateFlags()
      });
    }
    return;
  }

  if (props.child.type === 'asset') await loadAssetState();
  if (props.child.type === 'collection') await loadCollectionState();
});

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

const refreshCachedItem = async () => {
  if (props.child.type === 'asset') {
    await loadAssetState({ silent: true });
    return;
  }
  if (props.child.type === 'collection') {
    await loadCollectionState({ silent: true });
  }
};

const handleSilentRefresh = (eventData = {}) => {
  const refreshTask = refreshCachedItem();
  if (Array.isArray(eventData.tasks)) {
    eventData.tasks.push(refreshTask);
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
  emitter.on('silent-refresh-browser-items', handleSilentRefresh);
});

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeyArrowKeys);
  emitter.off('silent-refresh-browser-items', handleSilentRefresh);
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
