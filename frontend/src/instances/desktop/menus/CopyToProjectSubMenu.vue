<template>
  <div ref="subMenu" class="filter-menu-container sub-menu-container" v-stop-propagation @click.stop>
    
    <!-- Header with back button -->
    <div class="sub-menu-header" @click.stop>
      <ActionButton :icon="getAppIcon('chevron-left')" :showLabel="true" :fullWidth="true"
        :label="headerTitle" :buttonFunction="goBack" />
    </div>

    <span class="menu-divider"></span>

    <!-- Copy here option -->
    <ActionButton v-if="currentView === 'entities'" :icon="getAppIcon('arrow-down-ramp')" :showLabel="true"
      :fullWidth="true" label="Copy Here" :buttonFunction="() => copyToLocation(currentParentId)" />

    <!-- <span v-if="filteredEntities.length > 0" class="menu-divider"></span> -->

    <!-- Search input -->
    <div v-if="showSearch" class="input-section">
      <input ref="searchInput" v-stop-propagation v-model="searchTerm" class="input-short" type="text"
        :placeholder="currentView === 'projects' ? 'Search projects' : 'Search collections'" />
    </div>

    <span v-if="showSearch" class="menu-divider"></span>

    <!-- Loading state -->
    <div v-if="isLoading" class="sub-menu-loading">
      <span class="menu-item-text">Loading...</span>
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
          <span class="menu-item-text">{{ searchTerm ? 'No matching projects' : 'No other projects available' }}</span>
        </div>
      </template>

      <!-- Entity list (when navigating inside a project) -->
      <template v-else-if="currentView === 'entities'">

        <!-- Child entities -->
        <template v-for="entity in filteredEntities" :key="entity.id">
          <ActionButton :customIconUrl="getAppIcon(entity.entity_type_icon || 'folder')" :icon="getAppIcon('chevron-right')" 
            :showLabel="true" :fullWidth="true" :iconAfter="true" :label="entity.name" 
            :buttonFunction="() => navigateIntoEntity(entity)" />
        </template>

        <div v-if="filteredEntities.length === 0 && childEntities.length > 0" class="sub-menu-empty">
          <span class="menu-item-text subtle">No matching collections</span>
        </div>

        <div v-if="childEntities.length === 0" class="sub-menu-empty">
          <span class="menu-item-text subtle">No sub-collections</span>
        </div>
      </template>

    </div>

  </div>
</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
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

const assetStore = useAssetStore();
const iconStore = useIconStore();
const menu = useMenu();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();

// refs
const childEntities = ref([]);
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

// Returns the current parent entity ID.
const currentParentId = computed(() => {
  return currentNavItem.value?.parentId || null;
});

// Returns the current view type.
const currentView = computed(() => {
  return currentNavItem.value?.type || 'projects';
});

// Returns entities filtered by search term.
const filteredEntities = computed(() => {
  if (!searchTerm.value) {
    return childEntities.value;
  }
  const term = searchTerm.value.toLowerCase();
  return childEntities.value.filter(entity => 
    entity.name.toLowerCase().includes(term)
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
    return 'Select Project';
  }
  return currentNavItem.value?.title || 'Select Location';
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
  return childEntities.value.length > 10;
});

// methods
// Copies the asset to the specified location.
const copyToLocation = async (targetEntityId, projectOverride = null) => {
  const targetProject = projectOverride || selectedProject.value;
  
  if (!targetProject) {
    notificationStore.errorNotification('Error', 'No target project selected');
    return;
  }
  
  const asset = assetStore.selectedAsset;
  if (!asset) {
    notificationStore.errorNotification('Error', 'No asset selected');
    return;
  }
  
  try {
    stage.operationActive = true;
    
    await AssetService.CopyAssetToProject(
      projectStore.activeProject.uri,
      asset.id,
      targetProject.uri,
      targetEntityId || '',
      false
    );
    
    notificationStore.addNotification('Asset Copied', `${asset.name} copied to ${targetProject.name}`, 'success');
    menu.hideContextMenu();
    menu.resetSubMenu();
  } catch (error) {
    console.error('Error copying asset:', error);
    notificationStore.errorNotification('Failed to Copy Asset', error);
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

// Loads child entities for a project and parent.
const loadEntities = async (project, parentId, entityFilePath = null) => {
  isLoading.value = true;
  childEntities.value = [];
  
  try {
    const entityId = parentId || 'root';
    const folderPath = entityFilePath || project.working_directory;
    
    const children = await CollectionService.GetCollectionChildren(
      project.uri,
      entityId,
      project.working_directory,
      folderPath,
      project.ignore_list || [],
      false
    );
    
    childEntities.value = (children.entities || []).sort((a, b) => 
      a.name.localeCompare(b.name)
    );
  } catch (error) {
    console.error('Error loading entities:', error);
    notificationStore.errorNotification('Failed to load collections', error);
  } finally {
    isLoading.value = false;
  }
};

// Navigates into an entity to show its children.
const navigateIntoEntity = async (entity) => {
  const project = selectedProject.value;
  searchTerm.value = '';
  
  menu.navigateSubMenuForward({
    type: 'entities',
    projectUri: project.uri,
    parentId: entity.id,
    entityFilePath: entity.file_path,
    title: entity.name
  });
  
  await loadEntities(project, entity.id, entity.file_path);
};

// Navigates into a project to show its root entities.
const navigateIntoProject = async (project) => {
  searchTerm.value = '';
  menu.subMenuState.selectedProject = project;
  
  menu.navigateSubMenuForward({
    type: 'entities',
    projectUri: project.uri,
    parentId: null,
    entityFilePath: project.working_directory,
    title: project.name
  });
  
  await loadEntities(project, null, null);
};

// watchers
watch(
  () => menu.subMenuState.navigationStack.length,
  async (newLength) => {
    if (newLength > 0) {
      const navItem = currentNavItem.value;
      const project = selectedProject.value;
      if (navItem?.type === 'entities' && project) {
        await loadEntities(project, navItem.parentId, navItem.entityFilePath);
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
  border-radius: 8px;
  background-color: var(--light-steel);
}

.scrollable-list-container::-webkit-scrollbar-track {
  margin-top: 5px;
  border-radius: 10px;
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
  color: var(--white);
}
</style>
