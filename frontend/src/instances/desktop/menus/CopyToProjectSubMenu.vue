<template>
  <div ref="subMenu" class="filter-menu-container sub-menu-container" v-stop-propagation @click.stop>
    
    <!-- Header with back button -->
    <div class="sub-menu-header" @click.stop>
      <ActionButton :icon="getAppIcon('chevron-left')" :showLabel="true" :fullWidth="true"
        :label="headerTitle" :buttonFunction="goBack" />
    </div>

    <span class="menu-divider"></span>

    <!-- Copy here option -->
    <ActionButton v-if="currentView === 'collections'" :icon="getAppIcon('arrow-down-ramp')" :showLabel="true"
      :fullWidth="true" :label="$t('menus.copyHere')" :buttonFunction="() => copyToLocation(currentParentId)" />

    <!-- <span v-if="filteredCollections.length > 0" class="menu-divider"></span> -->

    <!-- Search input -->
    <div v-if="showSearch" class="input-section">
      <input ref="searchInput" v-stop-propagation v-model="searchTerm" class="input-short" type="text"
        :placeholder="currentView === 'projects' ? $t('placeholders.searchProjects') : $t('placeholders.searchCollections')" />
    </div>

    <span v-if="showSearch" class="menu-divider"></span>

    <!-- Loading state -->
    <div v-if="isLoading" class="sub-menu-loading">
      <span class="menu-item-text">{{ $t('common.loading') }}...</span>
    </div>

    <!-- Scrollable list container -->
    <div v-else class="scrollable-list-container">
      
      <!-- Project list (initial view) -->
      <template v-if="currentView === 'projects'">
        <template v-for="project in filteredProjects" :key="project.uri">
          <ActionButton :emoji="project.icon && project.icon.length < 10 ? project.icon : ''"
            :customIconUrl="project.icon && project.icon.length >= 10 ? project.icon : ''" :icon="getAppIcon('chevron-right')" 
            :showLabel="true" :fullWidth="true" :iconAfter="true" :label="project.name"
            :buttonFunction="() => navigateIntoProject(project)" />
        </template>

        <div v-if="filteredProjects.length === 0" class="sub-menu-empty">
          <span class="menu-item-text">{{ searchTerm ? $t('menus.noMatchingProjects') : $t('menus.noOtherProjects') }}</span>
        </div>
      </template>

      <!-- Collection list (when navigating inside a project) -->
      <template v-else-if="currentView === 'collections'">

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
      </template>

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
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';

const { t } = useI18n();
const assetStore = useAssetStore();
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
// Returns projects that are downloaded and not the current project.
const availableProjects = computed(() => {
  return projectStore.projects.filter(project => 
    project.is_downloaded && 
    project.uri !== projectStore.activeProject?.uri
  );
});

// Returns the current navigation item from the stack.
const currentNavItem = computed(() => {
  const stack = menu.subMenuState.navigationStack;
  return stack.length > 0 ? stack[stack.length - 1] : null;
});

// Returns the current parent collection ID.
const currentParentId = computed(() => {
  return currentNavItem.value?.parentId || null;
});

// Returns the current view type.
const currentView = computed(() => {
  return currentNavItem.value?.type || 'projects';
});

// Returns collections filtered by search term.
const filteredCollections = computed(() => {
  if (!searchTerm.value) {
    return childCollections.value;
  }
  const term = searchTerm.value.toLowerCase();
  return childCollections.value.filter(collection => 
    collection.name.toLowerCase().includes(term)
  );
});

// Returns projects filtered by search term.
const filteredProjects = computed(() => {
  if (!searchTerm.value) {
    return availableProjects.value;
  }
  const term = searchTerm.value.toLowerCase();
  return availableProjects.value.filter(project => 
    project.name.toLowerCase().includes(term)
  );
});

// Returns the header title based on navigation depth.
const headerTitle = computed(() => {
  if (navigationDepth.value <= 1) {
    return t('menus.selectProject');
  }
  return currentNavItem.value?.title || t('menus.selectLocation');
});

// Returns the current navigation depth.
const navigationDepth = computed(() => {
  return menu.subMenuState.navigationStack.length;
});

// Returns the currently selected project.
const selectedProject = computed(() => {
  return menu.subMenuState.selectedProject;
});

// Determines whether to show the search input (more than 10 items).
const showSearch = computed(() => {
  if (currentView.value === 'projects') return availableProjects.value.length > 10;
  return childCollections.value.length > 10;
});

// methods
// Copies the asset to the specified location.
const copyToLocation = async (targetCollectionId, projectOverride = null) => {
  const targetProject = projectOverride || selectedProject.value;
  
  if (!targetProject) {
    notificationStore.errorNotification(t('common.error'), t('notifications.noTargetProject'));
    return;
  }
  
  const asset = assetStore.selectedAsset;
  if (!asset) {
    notificationStore.errorNotification(t('common.error'), t('notifications.noAssetSelected'));
    return;
  }
  
  try {
    stage.operationActive = true;
    
    await AssetService.CopyAssetToProject(
      projectStore.activeProject.uri,
      asset.id,
      targetProject.uri,
      targetCollectionId || '',
      false
    );
    
    notificationStore.addNotification(t('notifications.assetCopied'), t('notifications.assetCopiedTo', { name: asset.name, project: targetProject.name }), 'success');
    menu.hideContextMenu();
    menu.resetSubMenu();
  } catch (error) {
    console.error('Error copying asset:', error);
    notificationStore.errorNotification(t('notifications.failedToCopyAsset'), error);
  } finally {
    stage.operationActive = false;
  }
};

// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Navigates back in the sub-menu.
const goBack = () => {
  searchTerm.value = '';
  menu.navigateSubMenuBack();
};

// Loads child collections for a project and parent.
const loadCollections = async (project, parentId, collectionFilePath = null) => {
  isLoading.value = true;
  childCollections.value = [];
  
  try {
    const collectionId = parentId || 'root';
    const folderPath = collectionFilePath || project.working_directory;
    
    const children = await CollectionService.GetCollectionChildren(
      project.uri,
      collectionId,
      project.working_directory,
      folderPath,
      project.ignore_list || [],
      false
    );
    
    childCollections.value = (children.collections || []).sort((a, b) => 
      a.name.localeCompare(b.name)
    );
  } catch (error) {
    console.error('Error loading collections:', error);
    notificationStore.errorNotification(t('notifications.failedToLoadCollections'), error);
  } finally {
    isLoading.value = false;
  }
};

// Navigates into an collection to show its children.
const navigateIntoCollection = async (collection) => {
  const project = selectedProject.value;
  searchTerm.value = '';
  
  menu.navigateSubMenuForward({
    type: 'collections',
    projectUri: project.uri,
    parentId: collection.id,
    collectionFilePath: collection.file_path,
    title: collection.name
  });
  
  await loadCollections(project, collection.id, collection.file_path);
};

// Navigates into a project to show its root collections.
const navigateIntoProject = async (project) => {
  searchTerm.value = '';
  menu.subMenuState.selectedProject = project;
  
  menu.navigateSubMenuForward({
    type: 'collections',
    projectUri: project.uri,
    parentId: null,
    collectionFilePath: project.working_directory,
    title: project.name
  });
  
  await loadCollections(project, null, null);
};

// watchers
watch(
  () => menu.subMenuState.navigationStack.length,
  async (newLength) => {
    if (newLength > 0) {
      const navItem = currentNavItem.value;
      const project = selectedProject.value;
      if (navItem?.type === 'collections' && project) {
        await loadCollections(project, navItem.parentId, navItem.collectionFilePath);
      }
    }
  }
);

// lifecycle hooks
onMounted(() => {
  if (subMenu.value) {
    menu.popUpMenuWidth = subMenu.value.getBoundingClientRect().width;
    menu.popUpMenu = subMenu.value;
  }
  if (searchInput.value) {
    searchInput.value.focus();
  }
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
  border-radius: var(--normal-radius);
  background-color: hsl(var(--border));
}

.scrollable-list-container::-webkit-scrollbar-track {
  margin-top: 5px;
  border-radius: var(--large-radius);
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
