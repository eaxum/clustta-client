<template>
  <div ref="subMenu" class="filter-menu-container sub-menu-container" v-stop-propagation @click.stop>
    
    <!-- Header with back button -->
    <div class="sub-menu-header" @click.stop>
      <ActionButton :icon="isAtRoot ? null : getAppIcon('chevron-left')" :showLabel="true" :fullWidth="true"
        :label="headerTitle" :buttonFunction="isAtRoot ? null : goBack" :isInactive="isAtRoot" />
    </div>

    <span class="menu-divider"></span>

    <!-- Move here option -->
    <ActionButton v-if="canMoveHere" :icon="getAppIcon('circle-check')" :showLabel="true"
      :fullWidth="true" :label="$t('menus.moveHere')" :buttonFunction="() => moveToLocation(currentParentId)" />

    <!-- Administrators can move to project root when the assets started in a collection. -->
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
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import { isProjectAdmin } from '@/lib/permissions';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// services
import { AssetService, CollectionService } from '@/services';

// stores
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';

const { t } = useI18n();
const iconStore = useIconStore();
const menu = useMenu();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();

// refs
const allowedCollections = ref([]);
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
  return !isAtRoot.value && currentParentId.value !== startingId;
});

// Only project administrators can move assets outside collection assignment scope.
const canMoveToRoot = computed(() => {
  const startingId = menu.subMenuState.startingCollectionId || '';
  return isProjectAdmin() && startingId !== '';
});

// Returns the current navigation item from the stack.
const currentNavItem = computed(() => {
  const stack = menu.subMenuState.navigationStack;
  return stack.length > 0 ? stack[stack.length - 1] : null;
});

// Returns the current parent collection ID.
const currentParentId = computed(() => currentNavItem.value?.parentId ?? '');

// Returns the allowed children at the current location. At the virtual root,
// collections whose physical parent is outside the allowed set become roots.
const childCollections = computed(() => {
  const allowedIds = new Set(allowedCollections.value.map(collection => collection.id));
  const collections = isAtRoot.value
    ? allowedCollections.value.filter(collection => !allowedIds.has(collection.parent_id))
    : allowedCollections.value.filter(collection => collection.parent_id === currentParentId.value);

  return collections.slice().sort((a, b) => a.name.localeCompare(b.name));
});

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

// methods

// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Navigates back in the sub-menu (to parent collection).
const goBack = () => {
  searchTerm.value = '';
  if (isAtRoot.value) return;
  menu.subMenuState.slideDirection = 'right';
  menu.subMenuState.navigationStack.pop();
};

// Loads collections in the active user's direct or inherited assignment scope.
const loadCollections = async () => {
  isLoading.value = true;
  allowedCollections.value = [];
  
  try {
    const collections = await CollectionService.GetCollections(projectStore.activeProject.uri);
    allowedCollections.value = (collections || []).filter(collection => collection.can_modify);
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
const navigateIntoCollection = (collection) => {
  searchTerm.value = '';
  
  menu.navigateSubMenuForward({
    type: 'move-to-collection',
    parentId: collection.id,
    collectionFilePath: collection.file_path,
    title: collection.name
  });
  
  menu.subMenuState.slideDirection = 'left';
};

// lifecycle hooks
onMounted(async () => {
  if (subMenu.value) {
    menu.popUpMenuWidth = subMenu.value.getBoundingClientRect().width;
    menu.popUpMenu = subMenu.value;
  }
  if (searchInput.value) {
    searchInput.value.focus();
  }
  
  // Start at a virtual root containing only the user's assignment scopes.
  menu.subMenuState.navigationStack = [];
  await loadCollections();
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
  background-color: var(--surface-4);
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
  color: var(--text);
}
</style>
