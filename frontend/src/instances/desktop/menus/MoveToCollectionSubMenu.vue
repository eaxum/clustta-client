<template>
  <div ref="subMenu" class="filter-menu-container sub-menu-container" v-stop-propagation @click.stop>
    
    <!-- Header with back button -->
    <div class="sub-menu-header" @click.stop>
      <ActionButton :icon="isAtRoot ? null : getAppIcon('chevron-left')" :showLabel="true" :fullWidth="true"
        :label="headerTitle" :buttonFunction="isAtRoot ? null : goBack" :isInactive="isAtRoot" />
    </div>

    <span class="menu-divider"></span>

    <!-- Move here option -->
    <ActionButton v-if="canMoveHere" :icon="getAppIcon('arrow-down-ramp')" :showLabel="true"
      :fullWidth="true" :label="$t('menus.moveHere')" :buttonFunction="() => moveToLocation(currentParentId)" />

    <!-- Move to root option (when not at root and didn't start at root) -->
    <ActionButton v-if="canMoveToRoot" :icon="getAppIcon('home')" :showLabel="true"
      :fullWidth="true" :label="$t('menus.moveToRoot')" :buttonFunction="() => moveToLocation('')" />

    <!-- Search input -->
    <div v-if="childCollections.length > 10" class="input-section">
      <input ref="searchInput" v-stop-propagation v-model="searchTerm" class="input-short" type="text"
        :placeholder="$t('placeholders.searchCollections')" />
    </div>

    <span v-if="childCollections.length > 10" class="menu-divider"></span>

    <!-- Loading state -->
    <div v-if="isLoading" class="sub-menu-loading">
      <span class="menu-item-text">{{ $t('common.loading') }}...</span>
    </div>

    <!-- Scrollable list container -->
    <div v-else class="scrollable-list-container">
      
      <!-- Child collections -->
      <template v-for="collection in filteredCollections" :key="collection.id">
        <ActionButton :customIconUrl="getAppIcon(collection.collection_type_icon || 'folder')" :icon="getAppIcon('chevron-right')" 
          :showLabel="true" :fullWidth="true" :iconAfter="true" :label="collection.name" 
          :buttonFunction="() => navigateIntoCollection(collection)" />
      </template>

      <div v-if="filteredCollections.length === 0 && childCollections.length > 0" class="sub-menu-empty">
        <span class="menu-item-text subtle">{{ $t('menus.noMatchingCollections') }}</span>
      </div>

      <div v-if="childCollections.length === 0" class="sub-menu-empty">
        <span class="menu-item-text subtle">{{ $t('menus.noSubCollections') }}</span>
      </div>

    </div>

  </div>
</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// services
import { AssetService, CollectionService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';

const { t } = useI18n();
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const iconStore = useIconStore();
const menu = useMenu();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();

// refs
const childCollections = ref([]);
const isLoading = ref(false);
const searchInput = ref(null);
const searchTerm = ref('');
const subMenu = ref(null);

// computed

// Returns the asset IDs being moved.
const assetIds = computed(() => menu.subMenuState.selectedAssetIds || []);

// Checks if user can move to the current location (not the starting location).
const canMoveHere = computed(() => {
  const startingId = menu.subMenuState.startingCollectionId || '';
  return currentParentId.value !== startingId;
});

// Checks if user can move to root (not at root and didn't start at root).
const canMoveToRoot = computed(() => {
  const startingId = menu.subMenuState.startingCollectionId || '';
  return !isAtRoot.value && startingId !== '';
});

// Returns the current navigation item from the stack.
const currentNavItem = computed(() => {
  const stack = menu.subMenuState.navigationStack;
  return stack.length > 0 ? stack[stack.length - 1] : null;
});

// Returns the current parent collection ID.
const currentParentId = computed(() => currentNavItem.value?.parentId ?? '');

// Returns collections filtered by search term.
const filteredCollections = computed(() => {
  if (!searchTerm.value) return childCollections.value;
  const term = searchTerm.value.toLowerCase();
  return childCollections.value.filter(collection => collection.name.toLowerCase().includes(term));
});

// Returns the header title based on current location.
const headerTitle = computed(() => {
  if (isAtRoot.value) return projectStore.activeProject?.name || t('menus.projectRoot');
  return currentNavItem.value?.title || t('menus.collection');
});

// Checks if currently at project root.
const isAtRoot = computed(() => currentParentId.value === '');

// Returns the current navigation depth.
const navigationDepth = computed(() => menu.subMenuState.navigationStack.length);

// methods

// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Navigates back in the sub-menu (to parent collection).
const goBack = async () => {
  searchTerm.value = '';
  
  // If already at root, do nothing (can't go further back)
  if (isAtRoot.value) return;
  
  const currentId = currentParentId.value;
  
  try {
    // Get the parent of the current collection
    const currentCollection = await CollectionService.GetCollectionByID(projectStore.activeProject.uri, currentId);
    const parentId = currentCollection?.collection_id || '';
    
    // Get parent name for the header
    let parentName = projectStore.activeProject?.name || t('menus.projectRoot');
    if (parentId) {
      const parentCollection = await CollectionService.GetCollectionByID(projectStore.activeProject.uri, parentId);
      parentName = parentCollection?.name || t('menus.collection');
    }
    
    // Update navigation stack to go to parent
    menu.subMenuState.slideDirection = 'right';
    menu.subMenuState.navigationStack = [{
      type: 'move-to-collection',
      parentId: parentId,
      title: parentName
    }];
    
    await loadCollections(parentId);
  } catch (error) {
    console.error('Error navigating back:', error);
  }
};

// Loads child collections for a parent.
const loadCollections = async (parentId) => {
  isLoading.value = true;
  childCollections.value = [];
  
  try {
    const project = projectStore.activeProject;
    const collectionId = parentId || 'root';
    
    // Get the folder path for the parent
    let folderPath = project.working_directory;
    if (parentId) {
      const parentCollection = await CollectionService.GetCollectionByID(project.uri, parentId);
      if (parentCollection) folderPath = parentCollection.file_path;
    }
    
    const children = await CollectionService.GetCollectionChildren(
      project.uri,
      collectionId,
      project.working_directory,
      folderPath,
      project.ignore_list || [],
      false
    );
    
    // Filter out the assets being moved (if they're collections - shouldn't happen but safety check)
    childCollections.value = (children.collections || [])
      .filter(collection => !assetIds.value.includes(collection.id))
      .sort((a, b) => a.name.localeCompare(b.name));
  } catch (error) {
    console.error('Error loading collections:', error);
    notificationStore.errorNotification(t('notifications.failedToLoadCollections'), error);
  } finally {
    isLoading.value = false;
  }
};

// Moves the assets to the specified collection.
const moveToLocation = async (targetCollectionId) => {
  if (!assetIds.value.length) {
    notificationStore.errorNotification(t('common.error'), t('notifications.noAssetsSelected'));
    return;
  }
  
  try {
    stage.operationActive = true;
    
    await AssetService.MoveAssetsToCollection(
      projectStore.activeProject.uri,
      assetIds.value,
      targetCollectionId
    );
    
    const count = assetIds.value.length;
    notificationStore.addNotification(
      t('notifications.assetsMoved'),
      t('notifications.assetsMovedDesc', { count }),
      'success'
    );
    
    menu.hideContextMenu();
    menu.resetSubMenu();
    emitter.emit('refresh-browser');
  } catch (error) {
    console.error('Error moving assets:', error);
    notificationStore.errorNotification(t('notifications.failedToMoveAssets'), error);
  } finally {
    stage.operationActive = false;
  }
};

// Navigates into an collection to show its children.
const navigateIntoCollection = async (collection) => {
  searchTerm.value = '';
  
  menu.navigateSubMenuForward({
    type: 'move-to-collection',
    parentId: collection.id,
    collectionFilePath: collection.file_path,
    title: collection.name
  });
  
  await loadCollections(collection.id);
};

// watchers
watch(
  () => menu.subMenuState.navigationStack.length,
  async (newLength) => {
    if (newLength > 0) {
      const navItem = currentNavItem.value;
      if (navItem?.type === 'move-to-collection') {
        await loadCollections(navItem.parentId || '');
      }
    }
  }
);

// lifecycle hooks
onMounted(async () => {
  if (subMenu.value) {
    menu.popUpMenuWidth = subMenu.value.getBoundingClientRect().width;
    menu.popUpMenu = subMenu.value;
  }
  if (searchInput.value) {
    searchInput.value.focus();
  }
  
  // Initialize navigation stack with starting collection info
  const startingId = menu.subMenuState.startingCollectionId || '';
  
  if (startingId) {
    // Starting from a collection - get its info to initialize the stack
    try {
      const startingCollection = await CollectionService.GetCollectionByID(projectStore.activeProject.uri, startingId);
      if (startingCollection) {
        menu.subMenuState.navigationStack = [{
          type: 'move-to-collection',
          parentId: startingId,
          title: startingCollection.name
        }];
      }
    } catch (error) {
      console.error('Error loading starting collection:', error);
    }
  }
  // If startingId is empty, stack stays empty = at root
  
  await loadCollections(startingId);
});

onBeforeUnmount(() => {
  if (subMenu.value) {
    menu.popUpMenuWidth = subMenu.value.getBoundingClientRect().width;
    menu.popUpMenuHeight = subMenu.value.getBoundingClientRect().height;
  }
  menu.resetSubMenu();
});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/menu.css";

.sub-menu-container {
  max-height: 400px;
  display: flex;
  flex-direction: column;
}

.sub-menu-header {
  width: 100%;
}

.sub-menu-header :deep(.action-button) {
  justify-content: flex-start;
}

.input-section {
  min-height: min-content;
  width: 100%;
}

.input-short {
  flex: 1;
  width: 100%;
  font-size: 14px;
}

.scrollable-list-container {
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  overflow-x: hidden;
  flex: 1;
  min-height: 0;
  min-width: 200px;
  max-width: 200px;
}

.scrollable-list-container :deep(.action-button) {
  min-width: 0;
  max-width: 100%;
}

.scrollable-list-container :deep(.button-label) {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  min-width: 0;
}

.scrollable-list-container::-webkit-scrollbar {
  width: 4px;
}

.scrollable-list-container::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: hsl(var(--border));
}

.scrollable-list-container::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.sub-menu-loading,
.sub-menu-empty {
  padding: 0.8rem 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.sub-menu-empty .menu-item-text.subtle {
  opacity: 0.6;
  font-style: italic;
  color: hsl(var(--foreground));
}
</style>
