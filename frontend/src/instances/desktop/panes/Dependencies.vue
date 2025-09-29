<template>

  <div class="general-pane-header">
      <div class="searchbar-container" v-esc="clearSearch">
        <input ref="searchBar"  v-model="searchQuery" class="pane-search-bar" type="text" spellcheck="false"
          :placeholder="'Search for dependencies to add'" @input="debouncedUpdateSearch" />

        
        <span v-if="searchQuery && isLoadingData" class="single-action-button">
          <img class="small-icons loading-children-icon" :src="getAppIcon('loading')">
        </span>
        <ActionButton v-else-if="searchQuery" :icon="getAppIcon('close')"
          :allowDeactivate="true" v-tooltip="'Clear search'" :buttonFunction="clearSearch" />
      </div>
      
			<FilterButton v-if="searchQuery" :icon="getAppIcon('filter')" v-tooltip="'Filter'"
			 :showLabel="false" :alert="isFilterActive"	 @click="showFilterMenu($event, 'dependencySearchFilterMenu')" />

  </div>

  <div class="general-pane-root">

    <div v-if="isSearching" class="sidebar-scroll" >

      <PageState v-if="!availableDependencies.length && !isLoadingData" :message="message()" :illustration="illustration()" />
      <DependencyList v-else :items="availableDependencies" :isDependency="true" :showAdd="true" :forList="true" />
    </div>
    
    <div v-else-if="assetDependencies.length" class="sidebar-scroll">
      <DependencyList :items="assetDependencies" :isDependency="true" :showRemove="true" :forList="true"/>
      
      <div class="bottom-bar">
        <ActionButton :icon="getAppIcon('square-arrow-right-up')" :showLabel="true" :iconAfter="true" :fullWidth="false" label="View in Graph"
        :buttonFunction="goToDependencyGraph" />
      </div>
    </div>

    <PageState v-else :message="message()" :illustration="illustration()" />

  </div>

</template>

<script setup>
// imports
import { computed, ref, watch, onMounted, onUnmounted } from 'vue';
import { useDebounce } from '@/lib/debounce';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';
import { isValidWeblink } from '@/lib/pointer';

// services
import { AssetService, CollectionService } from "@/../bindings/clustta/services";

// states/store imports
import { useCommonStore } from '@/stores/common';
import { useStageStore } from '@/stores/stages';
import { useNotificationStore } from '@/stores/notifications';
import { useAssetStore } from '@/stores/assets';
import { useIconStore } from '@/stores/icons';
import { useProjectStore } from '@/stores/projects';
import { useMenu } from '@/stores/menu';
import { useDependencyStore } from '@/stores/dependency';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DependencyList from '@/instances/desktop/components/DependencyList.vue';
import PageState from '@/instances/common/components/PageState.vue';
import FilterButton from '@/instances/desktop/components/FilterButton.vue';

// states/stores
const stage = useStageStore();
const notificationStore = useNotificationStore();
const assetStore = useAssetStore();
const projectStore = useProjectStore();
const iconStore = useIconStore();
const menu = useMenu();
const commonStore = useCommonStore();
const dependencyStore = useDependencyStore();

// refs
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
  const assets = commonStore.filterDependencyAssets ? dependencies.filter((item) => item.type === 'task' && !item.is_resource) : [];
  const resources = commonStore.filterDependencyResources ? dependencies.filter((item) => item.type === 'task' && item.is_resource ) : [];
  const collections = commonStore.filterDependencyCollections ? dependencies.filter((item) => item.type === 'entity') : [];
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
    notificationStore.errorNotification("Error loading project data", error);
  } finally {
    isLoadingData.value = false;
  }
};

const message = () => {
  if(isSearching.value){
    return 'No items match your search';
  } else {
    return 'This asset has no dependencies';
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
  console.log(assetStore.selectedAsset);
  const selectedAssetDependencies = assetStore.selectedAsset?.dependencies;
  const selectedAssetCollectionDependencies = assetStore.selectedAsset?.entity_dependencies;
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
      if(item.type === "task"){
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

    // await assetStore.processAssetsIconsAndPreviews(children.tasks);
    
    assetDependencies.value = children;
    // console.log(children);
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

  if (itemType === "task") {
    selectedAssetDependencies = asset.dependencies;
    await AssetService.RemoveAssetDependency(projectStore.activeProject.uri, asset.id, dependencyId)
      .then((response) => {
        notificationStore.addNotification("Dependency Removed", "", "success");
        assetDependencies.value = assetDependencies.value.filter((item) => item.id !== dependencyId)
        emitUpdates(asset.id, [
          { property: 'dependencies', value: selectedAssetDependencies.filter(dep => dep !== dependencyId) }
        ])
      })
      .catch((error) => {
        notificationStore.errorNotification("Error removing dependencies", error);
      });
  } else {
    selectedAssetDependencies = asset.entity_dependencies;
    await AssetService.RemoveEntityDependency(projectStore.activeProject.uri, asset.id, dependencyId)
      .then((response) => {
        notificationStore.addNotification("Dependency Removed", "", "success");
        assetDependencies.value = assetDependencies.value.filter((item) => item.id !== dependencyId)
        emitUpdates(asset.id, [
          { property: 'entity_dependencies', value: selectedAssetDependencies.filter(dep => dep !== dependencyId) }
        ])
      })
      .catch((error) => {
        notificationStore.errorNotification("Error removing dependencies", error);
      });
  }
  
};

const addDependency = async (dependencyId, itemType) => {
  const asset = assetStore.selectedAsset;
  let selectedAssetDependencies;
  let dependencyTypeID = dependencyStore.dependency_types.find(item => item.name === "linked")?.id;
  
  if (!dependencyTypeID) {
    notificationStore.errorNotification("Error adding dependency", "Default dependency type not found");
    return;
  }

  if (itemType === "task") {
    selectedAssetDependencies = asset.dependencies || [];
    
    // Check if dependency already exists
    if (selectedAssetDependencies.includes(dependencyId)) {
      notificationStore.addNotification("Dependency already exists", "", "info");
      return;
    }
    
    await AssetService.AddAssetDependency(projectStore.activeProject.uri, asset.id, dependencyId, dependencyTypeID)
      .then((response) => {
        notificationStore.addNotification("Dependency Added", "", "success");
        
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
        notificationStore.errorNotification("Error adding dependencies", error);
      });
  } else {
    selectedAssetDependencies = asset.entity_dependencies || [];
    
    // Check if dependency already exists
    if (selectedAssetDependencies.includes(dependencyId)) {
      notificationStore.addNotification("Dependency already exists", "", "info");
      return;
    }
    
    await AssetService.AddEntityDependency(projectStore.activeProject.uri, asset.id, dependencyId, dependencyTypeID)
      .then((response) => {
        notificationStore.addNotification("Dependency Added", "", "success");
        
        // Update local asset collection dependencies
        if (!asset.entity_dependencies) {
          asset.entity_dependencies = [];
        }
        asset.entity_dependencies.push(dependencyId);
        
        // Remove from available dependencies
        availableDependencies.value = availableDependencies.value.filter(dep => dep.id !== dependencyId);
        
        // Refresh the dependency list
        getAssetDependencies();
        
        emitUpdates(asset.id, [
          { property: 'entity_dependencies', value: asset.entity_dependencies }
        ]);
      })
      .catch((error) => {
        notificationStore.errorNotification("Error adding dependencies", error);
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
  emitter.on('addDependency', handleAddDependency);
  emitter.on('removeDependency', handleRemoveDependency);
});

onUnmounted(() => {
  console.log('unmounted')
  emitter.off('addDependency', handleAddDependency);
  emitter.off('removeDependency', handleRemoveDependency);
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

.pane-search-bar {
	font-family: 'Inter', sans-serif;
	box-sizing: border-box;
	font-size: 16px;
  font-weight: 300;
	border-radius: 8px;
	padding: 10px;
	border: 0px;
	border-style: solid;
	outline: none;
	background-color: var(--midnight-steel);
	color: var(--white);
	transition: width 0.2s ease-out;
	border-radius: var(--large-radius);
	width: 100%;
}
.general-pane-header{
  gap: .5rem;
}

.searchbar-container {
	display: flex;
	align-items: center;
	border: 0px;
	border-style: solid;
	outline: none;
	background-color: var(--midnight-steel);
	border-radius: var(--normal-radius);
	width: 100%;
  width: 98%;
	padding-right: .2rem;
	box-sizing: border-box;
  z-index: 2;
}

.searchbar-container:hover {
	outline: var(--transparent-line);
	outline-offset: -1px;
}

.pane-search-bar:focus .searchbar-container {
	background-color: red;
	outline: var(--solid-line);
	outline-offset: -1px;
}

.bottom-bar{
  display: flex;
  justify-content: flex-end;
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
  border-radius: 10px;
  background-color: var(--white);
}

.sidebar::-webkit-scrollbar-track {
  border-radius: 10px;
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
</style>


