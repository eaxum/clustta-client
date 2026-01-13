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
    <div class="input-section">
      <input 
        ref="searchInput" 
        v-stop-propagation 
        v-model="searchTerm" 
        class="input-short" 
        type="text"
        :placeholder="currentView === 'projects' ? 'Search projects' : 'Search collections'" 
      />
    </div>

    <span class="menu-divider"></span>

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
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue';

// Stores
import { useMenu } from '@/stores/menu';
import { useProjectStore } from '@/stores/projects';
import { useAssetStore } from '@/stores/assets';
import { useNotificationStore } from '@/stores/notifications';
import { useStageStore } from '@/stores/stages';
import { useIconStore } from '@/stores/icons';

// Services
import { AssetService, CollectionService } from '@/services';

// Components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

import emitter from '@/lib/mitt';

// Store instances
const menu = useMenu();
const projectStore = useProjectStore();
const assetStore = useAssetStore();
const notificationStore = useNotificationStore();
const stage = useStageStore();
const iconStore = useIconStore();

// Refs
const subMenu = ref(null);
const searchInput = ref(null);
const isLoading = ref(false);
const childEntities = ref([]);
const searchTerm = ref('');

// Icon helper
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Computed: Available projects (downloaded, not current)
const availableProjects = computed(() => {
  return projectStore.projects.filter(project => 
    project.is_downloaded && 
    project.uri !== projectStore.activeProject?.uri
  );
});

// Computed: Filtered projects based on search
const filteredProjects = computed(() => {
  if (!searchTerm.value) {
    return availableProjects.value;
  }
  const term = searchTerm.value.toLowerCase();
  return availableProjects.value.filter(project => 
    project.name.toLowerCase().includes(term)
  );
});

// Computed: Filtered entities based on search
const filteredEntities = computed(() => {
  if (!searchTerm.value) {
    return childEntities.value;
  }
  const term = searchTerm.value.toLowerCase();
  return childEntities.value.filter(entity => 
    entity.name.toLowerCase().includes(term)
  );
});

// Computed: Current navigation state
const currentNavItem = computed(() => {
  const stack = menu.subMenuState.navigationStack;
  return stack.length > 0 ? stack[stack.length - 1] : null;
});

const navigationDepth = computed(() => {
  return menu.subMenuState.navigationStack.length;
});

const currentView = computed(() => {
  return currentNavItem.value?.type || 'projects';
});

const currentParentId = computed(() => {
  return currentNavItem.value?.parentId || null;
});

const selectedProject = computed(() => {
  return menu.subMenuState.selectedProject;
});

// Computed: Header title
const headerTitle = computed(() => {
  if (navigationDepth.value <= 1) {
    return 'Select Project';
  }
  return currentNavItem.value?.title || 'Select Location';
});

// Navigation methods
const goBack = () => {
  searchTerm.value = '';
  menu.navigateSubMenuBack();
};

const navigateIntoProject = async (project) => {
  searchTerm.value = '';
  menu.subMenuState.selectedProject = project;
  
  // Navigate into project root
  menu.navigateSubMenuForward({
    type: 'entities',
    projectUri: project.uri,
    parentId: null,
    entityFilePath: project.working_directory,
    title: project.name
  });
  
  // Load root entities
  await loadEntities(project, null, null);
};

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
  
  // Load child entities
  await loadEntities(project, entity.id, entity.file_path);
};

// Load entities for a given project and parent
const loadEntities = async (project, parentId, entityFilePath = null) => {
  isLoading.value = true;
  childEntities.value = [];
  
  try {
    // Use GetCollectionChildren with proper project parameters
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
    
    // Extract only entities (collections) from the result
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

// Copy action
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
    
    // Call the CopyAssetToProject service
    await AssetService.CopyAssetToProject(
      projectStore.activeProject.uri,  // sourceProjectPath
      asset.id,                          // assetId
      targetProject.uri,                 // targetProjectPath
      targetEntityId || '',              // targetEntityId (empty string for root)
      false                              // copyAllCheckpoints - set to false as requested
    );
    
    notificationStore.addNotification(
      'Asset Copied',
      `${asset.name} copied to ${targetProject.name}`,
      'success'
    );
    
    // Close the menu
    menu.hideContextMenu();
    menu.resetSubMenu();
    
  } catch (error) {
    console.error('Error copying asset:', error);
    notificationStore.errorNotification('Failed to Copy Asset', error);
  } finally {
    stage.operationActive = false;
  }
};

// Watch for navigation changes to reload entities
watch(
  () => menu.subMenuState.navigationStack.length,
  async (newLength, oldLength) => {
    if (newLength > 0) {
      const navItem = currentNavItem.value;
      const project = selectedProject.value;
      if (navItem?.type === 'entities' && project) {
        await loadEntities(project, navItem.parentId, navItem.entityFilePath);
      }
    }
  }
);

// Lifecycle
onMounted(() => {
  if (subMenu.value) {
    menu.popUpMenuWidth = subMenu.value.getBoundingClientRect().width;
    menu.popUpMenu = subMenu.value;
  }
  // Focus search input
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
