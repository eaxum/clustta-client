<template>

  <div class="general-pane-header">
    <SearchBar v-model="searchQuery" :placeholder="$t('panes.searchDependencies')" :isLoading="isLoadingData"
      @input="debouncedUpdateSearch" @clear="clearSearch" />

    <FilterButton v-if="searchQuery" :icon="getAppIcon('filter')" v-tooltip="$t('panes.filter')"
      :showLabel="false" :alert="isFilterActive" @click="showFilterMenu($event, 'dependencySearchFilterMenu')" />
  </div>

  <div class="general-pane-root">

    <div v-if="isSearching" class="sidebar-scroll" >
      <div v-if="dependencyPresets.length" class="presets-section">
        <div class="section-header">{{ $t('panes.dependencyPresets') }}</div>
        <DependencyPresetItem v-for="preset in dependencyPresets" :key="preset.name" 
          :preset="preset" @apply="applyPreset" @delete="deletePreset" @update="updatePreset" />
      </div>

      <PageState v-if="!availableDependencies.length && !dependencyPresets.length && !isLoadingData" :message="message()" :illustration="illustration()" />
      <ItemsList v-else-if="availableDependencies.length" :items="availableDependencies" :isDependency="true" :showAdd="true" :forList="true" />
    </div>
    
    <div v-else-if="assetDependencies.length" class="sidebar-scroll">
      <ItemsList :items="assetDependencies" :isDependency="true" :showRemove="true" :forList="true"/>
      
      <div class="bottom-bar">
        <ActionButton :icon="getAppIcon('floppy-disk')" v-tooltip="$t('panes.saveAsPreset')" :buttonFunction="openSavePresetModal" />
        <ActionButton :icon="getAppIcon('square-arrow-right-up')"  v-tooltip="$t('panes.viewInGraph')" :buttonFunction="goToDependencyGraph" />
      </div>
    </div>

    <PageState v-else :message="message()" :illustration="illustration()" />

  </div>

</template>

<script setup>
// imports
import { computed, ref, watch, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useDebounce } from '@/lib/debounce';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';
import { isValidWeblink } from '@/lib/pointer';

// services
import { AssetService, CollectionService, SettingsService } from "@/services";

// states/store imports
import { useCommonStore } from '@/stores/common';
import { useStageStore } from '@/stores/stages';
import { useNotificationStore } from '@/stores/notifications';
import { useAssetStore } from '@/stores/assets';
import { useIconStore } from '@/stores/icons';
import { useProjectStore } from '@/stores/projects';
import { useMenu } from '@/stores/menu';
import { useDependencyStore } from '@/stores/dependency';
import { useDesktopModalStore } from '@/stores/desktopModals';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DependencyPresetItem from '@/instances/desktop/components/DependencyPresetItem.vue';
import FilterButton from '@/instances/desktop/components/FilterButton.vue';
import ItemsList from '@/instances/desktop/components/ItemsList.vue';
import PageState from '@/instances/common/components/PageState.vue';
import SearchBar from '@/instances/desktop/components/SearchBar.vue';

// states/stores
const commonStore = useCommonStore();
const stage = useStageStore();
const notificationStore = useNotificationStore();
const assetStore = useAssetStore();
const projectStore = useProjectStore();
const iconStore = useIconStore();
const menu = useMenu();
const dependencyStore = useDependencyStore();
const modals = useDesktopModalStore();

// i18n
const { t } = useI18n();

// refs
const dependencyPresets = ref([]);
const isLoadingData = ref(false);
const searchQuery = ref('');
const allDependencies = ref([]);
const availableDependencies = ref([]);

// computed props
const selectedAsset = computed(() => {
  return assetStore.selectedAsset
});

const isSearching = computed(() => {
  return searchQuery.value;
});

const isFilterActive = computed(() => {
  return !commonStore.filterDependencyAssets || !commonStore.filterDependencyCollections || !commonStore.filterDependencyResources
});


// methods
const showFilterMenu = (event, menuName) => {
	if (menu.activeMenu === menuName && menu.contextMenuVisible) {
		menu.disableAllMenus();
		menu.activeMenu = null;
	} else {
		menu.showContextMenu(event, menuName, true, true);
	}
};

const clearSearch = () => {
  searchQuery.value = '';
  availableDependencies.value = [];
  menu.disableAllMenus();
  menu.activeMenu = null;
}

const updateSearch = async () => {
  if (!searchQuery.value) {
    return;
  }
  if(!allDependencies.value.length){
    await fetchProjectData();
  }

  const dependencies = allDependencies.value;
  const assets = commonStore.filterDependencyAssets ? dependencies.filter((item) => item.type === 'asset' && !item.is_resource) : [];
  const resources = commonStore.filterDependencyResources ? dependencies.filter((item) => item.type === 'asset' && item.is_resource ) : [];
  const collections = commonStore.filterDependencyCollections ? dependencies.filter((item) => item.type === 'collection') : [];
  const query = searchQuery.value?.toLowerCase();
  const projectData = [ ...assets, ...resources, ...collections  ]
  availableDependencies.value = projectData?.filter((item) => item.name.toLowerCase().includes(query?.toLowerCase()))
  
};

const debouncedUpdateSearch = useDebounce(updateSearch, 300);


const fetchProjectData = async () => {

  const assetId = selectedAsset.value?.id;
  isLoadingData.value = true;

  try {
    const projectPath = projectStore.activeProject.uri;
    
    const [assetsResult, collectionsResult] = await Promise.all([
      AssetService.GetAssets(projectPath),
      CollectionService.GetCollections(projectPath)
    ]);
    
    const result = [ ...(assetsResult ?? []), ...(collectionsResult ?? [])];
    allDependencies.value = result.filter((item) => item. id !== assetId);
  } catch (error) {
    console.error("Error fetching sidebar data:", error);
    notificationStore.errorNotification(t('notifications.errorLoadingProjectData'), error);
  } finally {
    isLoadingData.value = false;
  }
};

const message = () => {
  if(isSearching.value){
    return t('panes.noItemsMatchSearch');
  } else {
    return t('panes.noDependencies');
  }
};

const illustration = () => {
  return '/page-states/resources.png';
};

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const goToDependencyGraph = () => {
  stage.setStageVisibility('dependencies', true);
};

const fetchDependencyPresets = async () => {
  try {
    const presets = await SettingsService.GetProjectDependencyPresets(projectStore.getActiveProject.id);
    dependencyPresets.value = presets || [];
  } catch (error) {
    console.error("Error fetching dependency presets:", error);
  }
};

const openSavePresetModal = () => {
  assetStore.dependencyPresetModalData = {
    dependencies: assetDependencies.value,
    existingPresets: dependencyPresets.value
  };
  modals.setModalVisibility('addDependencyPresetModal', true);
};

const applyPreset = async (preset) => {
  const asset = assetStore.selectedAsset;
  if (!asset) return;

  const existingAssetDeps = asset.dependencies || [];
  const existingCollectionDeps = asset.collection_dependencies || [];
  
  // Filter out dependencies that already exist
  const depsToAdd = preset.dependencies.filter(dep => {
    if (dep.type === 'asset') {
      return !existingAssetDeps.includes(dep.id);
    } else {
      return !existingCollectionDeps.includes(dep.id);
    }
  });

  const skippedCount = preset.dependencies.length - depsToAdd.length;
  
  for (const dep of depsToAdd) {
    await addDependency(dep.id, dep.type);
  }

  if (skippedCount > 0) {
    notificationStore.addNotification(
      t('notifications.presetAppliedWithSkipped', { skipped: skippedCount }),
      "",
      "warning"
    );
  }
};

const updatePreset = async (updatedPreset) => {
  try {
    if (updatedPreset.dependencies.length === 0) {
      await SettingsService.RemoveDependencyPreset(projectStore.getActiveProject.id, updatedPreset.name);
      dependencyPresets.value = dependencyPresets.value.filter(p => p.name !== updatedPreset.name);
    } else {
      await SettingsService.UpdateDependencyPreset(projectStore.getActiveProject.id, updatedPreset.name, updatedPreset);
      const index = dependencyPresets.value.findIndex(p => p.name === updatedPreset.name);
      if (index !== -1) {
        dependencyPresets.value[index] = updatedPreset;
      }
    }
  } catch (error) {
    console.error("Error updating preset:", error);
    notificationStore.errorNotification(t('notifications.errorUpdatingPreset'), error);
  }
};

// Deletes a dependency preset.
const deletePreset = async (preset) => {
  try {
    await SettingsService.RemoveDependencyPreset(projectStore.getActiveProject.id, preset.name);
    dependencyPresets.value = dependencyPresets.value.filter(p => p.name !== preset.name);
    notificationStore.addNotification(t('notifications.dependencyPresetDeleted'), '', 'success');
  } catch (error) {
    console.error("Error deleting preset:", error);
    notificationStore.errorNotification(t('notifications.errorDeletingDependencyPreset'), error);
  }
};

const handlePresetAdded = (newPreset) => {
  dependencyPresets.value.push(newPreset);
};

// refs
const assetDependencies = ref([]);

const emitUpdates = (assetId, updates) => {
  const updateData = { itemId: assetId, updates };
  
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

const getAssetDependencies = async() => {
	let project = projectStore.activeProject
  let allDependencies;
  const selectedAssetDependencies = assetStore.selectedAsset?.dependencies || [] ;
  const selectedAssetCollectionDependencies = assetStore.selectedAsset?.collection_dependencies || [];
  allDependencies = [ ...selectedAssetDependencies, ...selectedAssetCollectionDependencies];
  const children = await AssetService.GetAssetDependencies(project.uri, allDependencies);

  for (let i = 0; i < children.length; i++) {
      let item = children[i];
      let extension = "";
      if (item.pointer) {
        item.file_path = item.pointer;
      }
      if (item.is_link && !isValidWeblink(item.pointer)) {
        extension = utils.getFileExtension(item.pointer).substring(1);
      } else if (!item.is_link) {
        extension = children[i].extension?.toLowerCase().substring(1);
        if (!assetStore.projectExtensionsFlat.includes(extension)) {
          assetStore.projectExtensionsFlat.push(extension);
          let extensionData = {
            name: extension,
            type: "extension",
            extension: "." + extension,
            icon: (await iconStore.getIcon(extension)) || "",
          };
          assetStore.projectExtensions.push(extensionData);
        }
      }
      let iconPath = "";
      if(item.type === "asset"){
        if (item.is_link && isValidWeblink(item.pointer)) {
          iconPath = await iconStore.getWebIcon(item.pointer);
        } else {
          iconPath = (await iconStore.getIcon(extension)) || "";
        }
      } else{
          // iconPath = getAppIcon('folder')
      }
      children[i].icon = iconPath;
      let preview = null;
      if (item.preview) {
        preview = "data:image/png;base64," + item.preview;
      }
      children[i].preview = preview;
    }

    assetDependencies.value = children;
}

const handleRemoveDependency = (payload) => {
  removeDependency(payload.id, payload.itemType);
};

const handleAddDependency = (payload) => {
  addDependency(payload.id, payload.itemType);
};

const removeDependency = async (dependencyId, itemType) => {
  const asset = assetStore.selectedAsset;
  let selectedAssetDependencies;

  if (itemType === "asset") {
    selectedAssetDependencies = asset.dependencies;
    await AssetService.RemoveAssetDependency(projectStore.activeProject.uri, asset.id, dependencyId)
      .then((response) => {
        notificationStore.addNotification(t('notifications.dependencyRemoved'), "", "success");
        assetDependencies.value = assetDependencies.value.filter((item) => item.id !== dependencyId)
        emitUpdates(asset.id, [
          { property: 'dependencies', value: selectedAssetDependencies.filter(dep => dep !== dependencyId) }
        ])
      })
      .catch((error) => {
        notificationStore.errorNotification(t('notifications.errorRemovingDependency'), error);
      });
  } else {
    selectedAssetDependencies = asset.collection_dependencies;
    await AssetService.RemoveCollectionDependency(projectStore.activeProject.uri, asset.id, dependencyId)
      .then((response) => {
        notificationStore.addNotification(t('notifications.dependencyRemoved'), "", "success");
        assetDependencies.value = assetDependencies.value.filter((item) => item.id !== dependencyId)
        emitUpdates(asset.id, [
          { property: 'collection_dependencies', value: selectedAssetDependencies.filter(dep => dep !== dependencyId) }
        ])
      })
      .catch((error) => {
        notificationStore.errorNotification(t('notifications.errorRemovingDependency'), error);
      });
  }
  
};

const addDependency = async (dependencyId, itemType) => {
  const asset = assetStore.selectedAsset;
  let selectedAssetDependencies;
  let dependencyTypeID = dependencyStore.dependency_types.find(item => item.name === "linked")?.id;
  
  if (!dependencyTypeID) {
    notificationStore.errorNotification(t('notifications.errorAddingDependency'), "Default dependency type not found");
    return;
  }

  if (itemType === "asset") {
    selectedAssetDependencies = asset.dependencies || [];
    
    // Check if dependency already exists
    if (selectedAssetDependencies.includes(dependencyId)) {
      notificationStore.addNotification(t('notifications.dependencyAlreadyExists'), "", "warning");
      return;
    }
    
    await AssetService.AddAssetDependency(projectStore.activeProject.uri, asset.id, dependencyId, dependencyTypeID)
      .then((response) => {
        notificationStore.addNotification(t('notifications.dependencyAdded'), "", "success");
        
        // Update local asset dependencies
        if (!asset.dependencies) {
          asset.dependencies = [];
        }
        asset.dependencies.push(dependencyId);
        
        // Remove from available dependencies
        availableDependencies.value = availableDependencies.value.filter(dep => dep.id !== dependencyId);
        
        // Refresh the dependency list
        getAssetDependencies();
        
        emitUpdates(asset.id, [
          { property: 'dependencies', value: asset.dependencies }
        ]);
      })
      .catch((error) => {
        notificationStore.errorNotification(t('notifications.errorAddingDependency'), error);
      });
  } else {
    selectedAssetDependencies = asset.collection_dependencies || [];
    
    // Check if dependency already exists
    if (selectedAssetDependencies.includes(dependencyId)) {
      notificationStore.addNotification(t('notifications.dependencyAlreadyExists'), "", "warning");
      return;
    }
    
    await AssetService.AddCollectionDependency(projectStore.activeProject.uri, asset.id, dependencyId, dependencyTypeID)
      .then((response) => {
        notificationStore.addNotification(t('notifications.dependencyAdded'), "", "success");
        
        // Update local asset collection dependencies
        if (!asset.collection_dependencies) {
          asset.collection_dependencies = [];
        }
        asset.collection_dependencies.push(dependencyId);
        
        // Remove from available dependencies
        availableDependencies.value = availableDependencies.value.filter(dep => dep.id !== dependencyId);
        
        // Refresh the dependency list
        getAssetDependencies();
        
        emitUpdates(asset.id, [
          { property: 'collection_dependencies', value: asset.collection_dependencies }
        ]);
      })
      .catch((error) => {
        notificationStore.errorNotification(t('notifications.errorAddingDependency'), error);
      });
  }
};

// watchers
watch(
  () => [
    commonStore.filterDependencyResources,
    commonStore.filterDependencyAssets,
    commonStore.filterDependencyCollections
  ],
  async () => {
    await updateSearch();
  },
  { deep: true }
);

watch(() => assetStore.selectedAsset, async () => {
    getAssetDependencies();
}, { deep: true });

// onMounted hook
onMounted( async () => {
  await getAssetDependencies();
  await fetchDependencyPresets();
  emitter.on('addDependency', handleAddDependency);
  emitter.on('removeDependency', handleRemoveDependency);
  emitter.on('dependency-preset-added', handlePresetAdded);
});

onUnmounted(() => {
  emitter.off('addDependency', handleAddDependency);
  emitter.off('removeDependency', handleRemoveDependency);
  emitter.off('dependency-preset-added', handlePresetAdded);
});


</script>
<style scoped>
@import "@/assets/desktop.css";

@keyframes loadingRotate {
  from {
      transform: rotate(0deg);
  }
  to {
      transform: rotate(360deg);
  }
}

.single-action-button{
  align-content: center;
  justify-content: center;
  pointer-events: none;
}

.loading-children-icon {
  width: 20px;
  height: 20px;
  overflow: hidden;
  padding: 0px;
  animation: loadingRotate .5s linear infinite;
}

.bottom-bar{
  display: flex;
  justify-content: flex-end;
}

.general-pane-header{
  width: 96%;
}

.sidebar-scroll {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-sizing: border-box;
  gap: .4rem;
  padding: .5rem;
  position: relative;
  height: 100%;
  width: 100%;
  justify-content: flex-start;
  padding: 5px;
}

.sidebar-scroll::-webkit-scrollbar {
  width: 4px;
}

.sidebar-scroll::-webkit-scrollbar-thumb {
  border-radius: var(--large-radius);
  background-color: hsl(var(--foreground));
}

.sidebar::-webkit-scrollbar-track {
  border-radius: var(--large-radius);
}

.action-bar {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: .6rem;
  width: max-content;
  width: 100%;
  height: max-content;
  padding: .2rem;
  align-items: flex-start;
}

.presets-section {
  display: flex;
  flex-direction: column;
  gap: .25rem;
  margin-bottom: .5rem;
  padding-bottom: .5rem;
  border-bottom: 1px solid var(--border-color);
  color: hsl(var(--foreground));
}

.section-header {
  font-size: .7rem;
  font-weight: 600;
  color: var(--subtle-text);
  text-transform: uppercase;
  padding: .25rem .5rem;
}
</style>


