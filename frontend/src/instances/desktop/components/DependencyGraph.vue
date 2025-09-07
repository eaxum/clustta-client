<template>
  <div class="page-list-root absolute-pane">
    <div class="page-list-container">
      <div class="dependency-graph-header">
        <div class="dependency-count"> {{ message }}</div>
        <div class="dependency-toggle-container">
          <div class="input-label"> Full graph</div>
          <ToggleSwitch :switchValueProp="useMaxDepth" @click="changeDepth()" />
        </div>
        <div v-if="false" class="node-filters">
          <ActionButton :icon="commonStore.showThumbs ? '/icons/hide_thumbs.svg' : '/icons/show_thumbs.svg'"
            v-tooltip="commonStore.showThumbs ? 'Hide Thumbnails' : 'Show Thumbnails'"
            :buttonFunction="toggleShowThumbs" />
          <ActionButton :icon="'/icons/new_task.svg'" :isActive="showTasks" v-tooltip="'Toggle Tasks Display'"
            :buttonFunction="toggleShowTasks" />
          <ActionButton :icon="'/entity-icons/other.svg'" :isActive="showEntities" v-tooltip="'Toggle Entity Display'"
            :buttonFunction="toggleShowEntities" />
          <ActionButton :icon="'/icons/resources.svg'" :isActive="showResources" v-tooltip="'Toggle Resources Display'"
            :buttonFunction="toggleShowResources" />
        </div>
      </div>
      <div class="task-graph-container">
        <div class="graph-container">
          <div class="fit-view-button">
            <ActionButton :icon="getAppIcon('arrows-expand')" v-tooltip="'Fit View'" @click="fitViewToAllNodes()" />
          </div>
          <VueFlow v-model="graphElements" :default-viewport="{ zoom: 1 }" :fit-view-on-init="true"
            :no-drag-class-name="noDragClassName">
            <Background :size="1" :gap="20" pattern-color="#BDBDBD" />
            <!-- <MiniMap /> -->
            <!-- <Controls />  -->
            <template #node-custom="props">
              <VirtualNode :isNode="true" :id="props.id" :data="props.data" />
            </template>
          </VueFlow>
        </div>
      </div>
    </div>

    <div class="sidebar-outer">
      <div class="sidebar">
        <input v-model="commonStore.viewSearchQuery" class="desktop-search-bar" type="text" placeholder="Search"
          @input="updateSearch" />

        <div class="deps-graph-filter">
          <div class="filter-options">
            <FilterButton :icon="getAppIcon('folder')" v-tooltip="'Collection Type'"
              :alert="isFilterActive('entity-type')" @mouseenter="flashFilterMenu($event, 'collectionTypeFilterMenu')"
              @click="showFilterMenu($event, 'collectionTypeFilterMenu')" />
            <FilterButton :icon="getAppIcon('brush')" v-tooltip="'Task Type'" :alert="isFilterActive('task-type')"
              @mouseenter="flashFilterMenu($event, 'assetTypeFilterMenu')"
              @click="showFilterMenu($event, 'assetTypeFilterMenu')" />
            <!-- <FilterButton  icon="/types-icons/layout.svg"  v-tooltip="'Extension'" :alert="isFilterActive('extension')"
                @mouseenter="flashFilterMenu($event, 'extensionFilterMenu')" @click="showFilterMenu($event, 'extensionFilterMenu')" /> -->
            <!-- <FilterButton v-if="showTagsFilter" :icon="getAppIcon('tag')"  v-tooltip="'Tags'" :alert="isFilterActive('tags')"
                @mouseenter="flashFilterMenu($event, 'tagsFilterMenu')" @click="showFilterMenu($event, 'tagsFilterMenu')" /> -->
            <FilterButton :icon="getAppIcon('filter')" v-tooltip="'Type'" :alert="isFilterActive('general')"
              @mouseenter="flashFilterMenu($event, 'typeFilterMenu')"
              @click="showFilterMenu($event, 'typeFilterMenu')" />
          </div>
          <ActionButton v-if="filtersActive" :icon="'/icons/close.svg'" v-tooltip="'Reset Filters'"
            :buttonFunction="clearFilters" />
        </div>

        <div class="sidebar-scroll">
          <DependencyVirtualScroll :forList="true" :itemComponent="VirtualNode" :items="projectData" :useRef="false" />
        </div>
      </div>
    </div>

  </div>
</template>


<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};
// imports
import { ref, computed, onMounted, watch, nextTick, markRaw, onUnmounted, reactive } from 'vue'
import dagre from '@dagrejs/dagre'
import { TaskService, EntityService } from "@/../bindings/clustta/services";
import emitter from '@/lib/mitt';
import utils from '@/services/utils';

// vue flow
import { VueFlow, useVueFlow, Position } from '@vue-flow/core'
import { Background } from '@vue-flow/background';
import { MiniMap } from '@vue-flow/minimap';
import { Controls } from '@vue-flow/controls';
import { useNodesInitialized } from '@vue-flow/core'
import { v4 as uuidv4 } from 'uuid'

// styles
import '@vue-flow/core/dist/style.css';
import '@vue-flow/core/dist/theme-default.css';

// state imports
import { useCommonStore } from '@/stores/common';
import { useEntityStore } from '@/stores/entity';
import { useAssetStore } from '@/stores/assets';
import { useNotificationStore } from '@/stores/notifications';
import { useDependencyStore } from '@/stores/dependency';
import { useTagStore } from '@/stores/tags';
import { useMenu } from '@/stores/menu';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import FilterButton from '@/instances/desktop/components/FilterButton.vue';
import VirtualNode from '@/instances/desktop/components/VirtualNode.vue'
import DependencyVirtualScroll from '@/instances/desktop/components/DependencyVirtualScroll.vue'
import { useProjectStore } from '@/stores/projects';

// states
const notificationStore = useNotificationStore();
const dependencyStore = useDependencyStore();
const tagStore = useTagStore();
const menu = useMenu();
const commonStore = useCommonStore();
const entityStore = useEntityStore();
const assetStore = useAssetStore();
const projectStore = useProjectStore();

// refs
const useMaxDepth = ref(false);
const graphElements = ref([]);
const noDragClassName = 'no-drag';

// New reactive data for refactored approach
const sidebarTasks = ref([]);
const sidebarEntities = ref([]);
const filteredTasks = ref([]);
const filteredEntities = ref([]);
const graphData = ref({ nodes: [], edges: [] });
const isLoadingGraph = ref(false);
const isLoadingSidebar = ref(false);

// vars
const { fitView, getNodes, addNodes, addEdges, setNodes, setEdges } = useVueFlow();
const smoothNode = false;
const nodeStyle = smoothNode ? 'smoothstep' : '';

// computed props
const maxDepth = computed(() => { return useMaxDepth.value ? 4 : 1 });

// filter
const filtersActive = computed(() => {
  return commonStore.entityFilters.length || commonStore.taskFilters.length || isFilterActive('general') || commonStore.resourceFilters.length;
});

const updateSearch = (event) => {
  const searchQuery = event.target.value;
  commonStore.viewSearchQuery = searchQuery.toLowerCase();
};

const isFilterActive = (filter) => {
  if (filter.includes('general')) {
    const isActive = commonStore.showEntities && commonStore.showTasks && commonStore.showResources;
    return !isActive;
  } else
    if (filter.includes('entity')) {
      return commonStore.entityFilters.some((item) => item.type === filter);
    } else {
      return commonStore.taskFilters.some((item) => item.type === filter);
    }
};

const showFilterMenu = (event, menuName) => {
  if (menu.activeMenu === menuName && menu.contextMenuVisible) {
    menu.disableAllMenus();
    menu.activeMenu = null;
  } else {
    menu.showContextMenu(event, menuName, true, true);
  }
};

const flashFilterMenu = (event, menuName) => {
  if (menu.contextMenuVisible && !menu.nonFilterMenus.includes(menu.activeMenu)) {
    menu.showContextMenu(event, menuName, true, true);
  }
};

const clearFilters = () => {
  commonStore.resetFilters();

};

const toggleShowThumbs = () => {
  commonStore.showThumbs = !commonStore.showThumbs;
};

const toggleShowTasks = async () => {
  showTasks.value = !showTasks.value;
  await buildGraphFromDependencies();
};

const toggleShowEntities = async () => {
  showEntities.value = !showEntities.value;
  await buildGraphFromDependencies();
};

const toggleShowResources = async () => {
  showResources.value = !showResources.value;
  await buildGraphFromDependencies();
};

// //////////////////////////////
const message = computed(() => {
  if (!useMaxDepth.value) {
    if (totalTaskDeps.value > 1) {
      return 'This task has ' + totalTaskDeps.value + ' direct dependencies';
    } else if (totalTaskDeps.value === 1) {
      return 'This task has ' + totalTaskDeps.value + ' direct dependency';
    } else {
      return 'This task has no dependencies';
    }
  } else {
    if (totalTaskDeps.value > 1) {
      return 'This task has ' + totalTaskDeps.value + ' total dependencies';
    } else if (totalTaskDeps.value === 1) {
      return 'This task has ' + totalTaskDeps.value + ' total dependency';
    } else {
      return 'This task has no dependencies';
    }
  }
});



const dependencies = ref([]);
const totalTaskDepsCount = ref(0);
const totalTaskDeps = computed(() => { return totalTaskDepsCount.value });
const showTasks = ref(true);
const showEntities = ref(true);
const showResources = ref(true);

// computed getters - refactored to use service data
const projectEntities = computed(() => {
  if (!commonStore.showEntities) return [];
  return filteredEntities.value;
});

const projectTasks = computed(() => {
  if (!commonStore.showTasks) return [];
  return filteredTasks.value;
});

const projectData = computed(() => {
  const selectedTask = assetStore.selectedTask;
  if (!selectedTask) return [];
  
  const allData = [...projectTasks.value, ...projectEntities.value]
  const currentDependencies = dependencies.value;
  const filteredData = allData.filter((item) => {
    if (currentDependencies.includes(item.id) || item.id === selectedTask.id) {
      return false;
    }
    
    const itemDependencies = [...(item.dependencies || []), ...(item.entity_dependencies || [])];
    if (itemDependencies.includes(selectedTask.id)) {
      return false;
    }
    
    return true;
  });
  return filteredData
});

// graph methods
const changeDepth = async () => {
  useMaxDepth.value = !useMaxDepth.value;
  await buildGraphFromDependencies();
  nextTick(() => {
    fitViewToAllNodes(true);
  })
};

const fitViewToAllNodes = (useDelay = false) => {
  const timeOut = useDelay ? 400 : 0;
  setTimeout(() => {
    fitView({ padding: 0.1, includeHiddenNodes: false, duration: 200 })
  }, timeOut);
};

// New async methods for refactored approach
const fetchSidebarData = async () => {
  isLoadingSidebar.value = true;
  try {
    const projectPath = projectStore.activeProject.uri;
    
    // Fetch all tasks and entities for the sidebar
    const [tasksResult, entitiesResult] = await Promise.all([
      TaskService.GetTasks(projectPath),
      EntityService.GetEntities(projectPath)
    ]);
    
    sidebarTasks.value = tasksResult || [];
    sidebarEntities.value = entitiesResult || [];
    
    // Update filtered tasks and entities after fetching
    await updateFilteredTasks();
    await updateFilteredEntities();
  } catch (error) {
    console.error("Error fetching sidebar data:", error);
    notificationStore.errorNotification("Error loading project data", error);
  } finally {
    isLoadingSidebar.value = false;
  }
};

const updateFilteredTasks = async () => {
  try {
    filteredTasks.value = await assetStore.filterTasks(sidebarTasks.value);
  } catch (error) {
    console.error("Error filtering tasks:", error);
    filteredTasks.value = [];
  }
};

const updateFilteredEntities = async () => {
  try {
    filteredEntities.value = await entityStore.filterEntities(sidebarEntities.value);
  } catch (error) {
    console.error("Error filtering entities:", error);
    filteredEntities.value = [];
  }
};

const buildGraphFromDependencies = async () => {
  isLoadingGraph.value = true;
  const selectedTask = assetStore.selectedTask;
  
  if (!selectedTask) {
    graphData.value = { nodes: [], edges: [] };
    isLoadingGraph.value = false;
    return;
  }

  try {
    // Use the new recursive dependencies service with proper depth control
    const dependencyItems = await TaskService.GetRecursiveDependencies(
      projectStore.activeProject.uri, 
      selectedTask.id, 
      maxDepth.value
    );

    console.log("Dependency items with parent info:", dependencyItems);
    
    // Build the graph nodes and edges
    const nodes = [];
    const edges = [];
    const nodeIdMap = new Map(); // Map from item IDs to generated node IDs
    
    // First, create the parent task node
    const parentNodeId = uuidv4();
    nodeIdMap.set(selectedTask.id, parentNodeId);
    
    const parentNode = {
      id: parentNodeId,
      label: selectedTask.name,
      position: { x: 0, y: 0 },
      type: 'custom',
      data: { ...selectedTask, nodeId: parentNodeId, parentId: null, depth: 0 },
      sourcePosition: Position.Left,
      targetPosition: Position.Right,
    };
    nodes.push(parentNode);

    totalTaskDepsCount.value = dependencyItems.length;

    // Create nodes for all dependencies
    dependencyItems.forEach(item => {
      const nodeId = uuidv4();
      
      // Extract the actual item data based on the service response structure
      let actualItem;
      let itemDepth = 1;
      let parentTaskId = selectedTask.id; // Default parent
      
      if (item.task) {
        actualItem = item.task;
        itemDepth = item.depth || 1;
        parentTaskId = item.parentId || selectedTask.id;
        nodeIdMap.set(actualItem.id, nodeId);
      } else if (item.entity) {
        actualItem = item.entity;
        itemDepth = item.depth || 1;
        parentTaskId = item.parentId || selectedTask.id;
        nodeIdMap.set(actualItem.id, nodeId);
      } else {
        // Fallback for direct item (if not wrapped in depth structure)
        actualItem = item;
        nodeIdMap.set(actualItem.id, nodeId);
      }

      const node = {
        id: nodeId,
        label: `${actualItem.name}${itemDepth === maxDepth.value ? ' (...)' : ''}`,
        position: { x: 0, y: 0 },
        type: 'custom',
        data: { ...actualItem, nodeId: nodeId, parentId: parentTaskId, depth: itemDepth },
        sourcePosition: Position.Left,
        targetPosition: Position.Right,
      };
      nodes.push(node);
    });

    // Create edges based on parent-child relationships
    dependencyItems.forEach(item => {
      let actualItem;
      let parentTaskId = selectedTask.id;
      
      if (item.task) {
        actualItem = item.task;
        parentTaskId = item.parentId || selectedTask.id;
      } else if (item.entity) {
        actualItem = item.entity;
        parentTaskId = item.parentId || selectedTask.id;
      } else {
        actualItem = item;
      }

      const childNodeId = nodeIdMap.get(actualItem.id);
      const parentNodeId = nodeIdMap.get(parentTaskId);
      
      if (childNodeId && parentNodeId && childNodeId !== parentNodeId) {
        edges.push({
          id: `e${parentNodeId}-${childNodeId}`,
          source: parentNodeId,
          target: childNodeId,
          type: nodeStyle,
        });
      }
    });

    // Update dependencies ref for other computed properties
    const allDependencyIds = dependencyItems.map(item => {
      if (item.task) return item.task.id;
      if (item.entity) return item.entity.id;
      return item.id;
    });
    dependencies.value = allDependencyIds;

    graphData.value = { nodes, edges };
  } catch (error) {
    console.error("Error building graph:", error);
    notificationStore.errorNotification("Error building dependency graph", error);
  } finally {
    isLoadingGraph.value = false;
  }
};

const applyDagreLayout = (nodes, edges) => {
  const g = new dagre.graphlib.Graph()
  g.setGraph({
    rankdir: 'LR',
    nodesep: 30,
    ranksep: 200,
    edgesep: 20,
  })
  g.setDefaultEdgeLabel(() => ({}))

  nodes.forEach(node => {
    g.setNode(node.id, { width: 150, height: 50 })
  })

  edges.forEach(edge => {
    g.setEdge(edge.source, edge.target)
  })

  dagre.layout(g)

  return nodes.map(node => {
    const nodeWithPosition = g.node(node.id)
    node.position = { x: nodeWithPosition.x, y: nodeWithPosition.y }
    return node
  })
};

// Replace buildGraph computed with reactive graph updates
const updateGraphLayout = (newGraph) => {
  const nodesWithLayout = applyDagreLayout(newGraph.nodes, newGraph.edges)
  graphElements.value = [...nodesWithLayout, ...newGraph.edges]
  nextTick(() => {
    setNodes(nodesWithLayout)
    setEdges(newGraph.edges)
  })
};

// Watch for graph data changes and update layout
watch(graphData, (newGraphData) => {
  if (newGraphData.nodes.length > 0 || newGraphData.edges.length > 0) {
    updateGraphLayout(newGraphData);
  }
}, { immediate: true });

// Watch for depth changes and rebuild graph
watch(maxDepth, () => {
  buildGraphFromDependencies();
});

// Watch for show toggles and rebuild graph  
watch([showTasks, showEntities], () => {
  buildGraphFromDependencies();
});

// Watch for selected task changes and rebuild graph
watch(() => assetStore.selectedTask, async (newTask) => {
  if (newTask) {
    await buildGraphFromDependencies();
  }
}, { immediate: false });

// Watch for filter changes and update filtered tasks and entities
watch(
  () => [
    commonStore.viewSearchQuery,
    commonStore.workspaceSearchQuery,
    commonStore.taskFilters,
    commonStore.entityFilters,
    commonStore.showResources
  ],
  async () => {
    await updateFilteredTasks();
    await updateFilteredEntities();
  },
  { deep: true }
);

// Watch for sidebar data changes and update filtered data
watch(sidebarTasks, async () => {
  await updateFilteredTasks();
}, { deep: true });

watch(sidebarEntities, async () => {
  await updateFilteredEntities();
}, { deep: true });

const selectTask = async (taskId) => {
  // Find task in our cached sidebar data first
  let task = sidebarTasks.value.find(item => item.id === taskId);
  
  if (!task) {
    // If not found in sidebar, try to fetch it from the service
    try {
      const allTasks = await TaskService.GetTasks(projectStore.activeProject.uri);
      task = allTasks.find(item => item.id === taskId);
    } catch (error) {
      console.error("Error fetching task:", error);
      notificationStore.errorNotification("Error selecting task", error);
      return;
    }
  }
  
  if (task) {
    assetStore.selectTask(task);
    await fetchSidebarData()
    await buildGraphFromDependencies();
    
    nextTick(() => {
      fitViewToAllNodes(true);
    });
  }
};

const addDependency = async (dependencyId, itemType) => {
  const task = assetStore.selectedTask;
  const allDependencies = [...sidebarTasks.value, ...sidebarEntities.value];

  let dependencyTypeID = dependencyStore.dependency_types.find(item => item.name === "linked").id;
  if (itemType === "task") {
    await TaskService.AddTaskDependency(projectStore.activeProject.uri, task.id, dependencyId, dependencyTypeID)
      .then( async(response) => {
        notificationStore.addNotification("Dependency Added", "", "success");
        const addedDependency = allDependencies.find((newDependency) => newDependency.id === dependencyId);
        if (addedDependency) {
          dependencies.value.push(addedDependency.id);
          
          await buildGraphFromDependencies();
          nextTick(() => {
            fitViewToAllNodes();
          })
        }
      })
      .catch((error) => {
        console.log(error)
        notificationStore.errorNotification("Error adding dependencies", error);
      });
  } else {
    await TaskService.AddEntityDependency(projectStore.activeProject.uri, task.id, dependencyId, dependencyTypeID)
      .then( async(response) => {
        notificationStore.addNotification("Dependency Added", "", "success");
        const addedDependency = allDependencies.find((newDependency) => newDependency.id === dependencyId);
        if (addedDependency) {
          dependencies.value.push(addedDependency.id);
          
          await buildGraphFromDependencies();
          nextTick(() => {
            fitViewToAllNodes();
          })
        }
      })
      .catch((error) => {
        console.log(error)
        notificationStore.errorNotification("Error adding dependencies", error);
      });
  }
};

const removeDependency = async (dependencyId, itemType) => {
  const task = assetStore.selectedTask;
  if (itemType === "task") {
    await TaskService.RemoveTaskDependency(projectStore.activeProject.uri, task.id, dependencyId)
      .then(async(response) => {
        notificationStore.addNotification("Dependency Removed", "", "success");
        dependencies.value = dependencies.value.filter(id => id !== dependencyId);
        await buildGraphFromDependencies();
        nextTick(() => {
          fitViewToAllNodes();
        })
      })
      .catch((error) => {
        notificationStore.errorNotification("Error removing dependencies", error);
      });
  } else {
    await TaskService.RemoveEntityDependency(projectStore.activeProject.uri, task.id, dependencyId)
      .then(async(response) => {
        notificationStore.addNotification("Dependency Removed", "", "success");
        dependencies.value = dependencies.value.filter(id => id !== dependencyId);
        buildGraphFromDependencies();
        nextTick(() => {
          fitViewToAllNodes();
        })
      })
      .catch((error) => {
        notificationStore.errorNotification("Error removing dependencies", error);
      });
  }
};

const handleSelectItem = (payload) => {
  const id = payload.message;
  selectTask(id);
};

const handleAddDependency = (payload) => {
  console.log(payload)
  addDependency(payload.id, payload.itemType);
};

const handleRemoveDependency = (payload) => {
  removeDependency(payload.id, payload.itemType);
};

onMounted(async () => {
  commonStore.resetFilters();
  
  // Fetch sidebar data first
  await fetchSidebarData();
  
  // Build initial graph if there's a selected task
  if (assetStore.selectedTask) {
    await buildGraphFromDependencies();
  }
  
  nextTick(() => {
    addNodes(graphElements.value.filter(el => el.position))
    addEdges(graphElements.value.filter(el => !el.position))
  })
  
  emitter.on('selectItem', handleSelectItem);
  emitter.on('addDependency', handleAddDependency);
  emitter.on('removeDependency', handleRemoveDependency);
});

onUnmounted(() => {
  emitter.off('selectItem', handleSelectItem)
  emitter.off('addDependency', handleAddDependency)
  emitter.off('removeDependency', handleRemoveDependency)
});
</script>

<style scoped>
@import "@/assets/desktop.css";

/* :deep(.vue-flow__handle) {
  width: 6px;
  height: 50%;
  background: #ffffff;
  border: 1px solid #fff;
  border-radius: 3px;
  cursor: pointer;
} */

/* :deep(.vue-flow__controls-button) {
  background: #fefefe;
  border: none;
  border-bottom: 1px solid #eee;
  box-sizing: content-box;
  display: flex;
  justify-content: center;
  align-items: center;
  width: 16px;
  height: 16px;
  cursor: pointer;
  user-select: none;
  padding: 5px;
} */

.page-list-root {
  box-sizing: border-box;
  display: flex;
  /* flex-direction: column; */
  align-items: center;
  justify-content: center;
  /* background-color: darkcyan; */
  color: var(--white);
  padding: 0px;
  /* padding: 1rem; */
  justify-content: flex-start;
}

.absolute-pane {
  padding: 0px;
}

.page-list-container {
  width: 100%;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
  flex-direction: column;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--white);
  padding: 1rem;
  padding-right: 0px;
  padding-top: 0px;
  /* background-color: tomato; */
}

.task-graph-container {
  background-color: var(--black-steel);
  border-radius: var(--large-radius);
  display: flex;
  width: 100%;
  height: 100%;
}

.temp-dependencies {
  width: 200px;
  overflow: hidden;
}

.dependency-graph-header {
  display: flex;
  justify-content: space-between;
  width: 100%;
  padding: .2rem 1rem;
  height: 60px;
  box-sizing: border-box;
  align-items: center;
  /* background-color: red; */
}

.dependency-listbox-container {
  display: flex;
  width: min-content;
  max-width: 50%;
  /* flex: 1; */
  width: 300px;
  padding: .2rem;
  justify-content: space-between;
  box-sizing: border-box;
  align-items: center;
  /* background-color: green; */
}

.dependency-toggle-container {
  display: flex;
  width: max-content;
  padding: .2rem;
  gap: 1rem;
  justify-content: space-between;
  box-sizing: border-box;
  align-items: center;
  /* background-color: green; */
}

.node-filters {
  /* background-color: firebrick; */
  display: flex;
  box-sizing: border-box;
  height: 100%;
  align-items: center;
  gap: .5rem;
  padding: .5rem;
}

.sidebar-outer {
  padding: 1rem;
  color: var(--white);
  /* border-left: var(--transparent-line); */
  position: relative;
  height: 100%;
  max-width: 600px;
  min-width: 250px;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex: 1 1 50%;
  /* transition: all .2s ease-out; */
  background-color: var(--black);
  background-color: transparent;
  /* background-color: red; */
}

.sidebar {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  /* overflow-y: scroll; */
  box-sizing: border-box;
  gap: .4rem;
  padding: .5rem;
  /* border-left: var(--transparent-line); */
  position: relative;
  height: 100%;
  max-width: 600px;
  min-width: 250px;
  width: 250px;
  justify-content: flex-start;
  padding: 10px;
  flex: 1 1 50%;
  background-color: tomato;
  background-color: var(--black-steel);
  border-radius: var(--large-radius);
}

.sidebar-scroll {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  /* overflow-y: scroll; */
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

.filter-alert {
  overflow: hidden;
  width: 2px;
  height: 2px;
  background-color: #ecb603;
  border-radius: 5px;
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: center;
  top: 2px;
  right: 2px;
  border-radius: 10px;
  padding: 3px;
  font-size: 12px;
  color: var(--white);
}


.match-condition {
  display: flex;
  width: max-content;
  white-space: nowrap;
  /* background-color: tomato; */
  height: 100%;
  align-items: center;
  gap: .5rem;
  padding: .5rem;
  /* overflow: hidden; */
  box-sizing: border-box;
}

.desktop-search-bar {
  font-family: 'Inter', sans-serif;
  font-weight: 200;
  box-sizing: border-box;
  font-size: 16px;
  border-radius: 8px;
  padding: 10px;
  border: 0px;
  border-style: solid;
  outline: none;
  background-color: var(--steel);
  color: var(--white);
  transition: width 0.2s ease-out;
  border-radius: var(--large-radius);
  width: 100%;
  max-width: 400px;
}

.desktop-search-bar::-ms-reveal {
  filter: invert(100%);
  /* color: white; */
}

.desktop-search-bar:hover {
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.desktop-search-bar:focus {
  outline: var(--solid-line);
  outline-offset: -1px;
}


.deps-graph-filter {
  position: relative;
  display: flex;
  width: 100%;
  align-items: center;
  height: max-content;
  gap: 1rem;
  justify-content: space-between;
  /* background-color: firebrick; */
  padding: .5rem;
  box-sizing: border-box;
  border-radius: var(--normal-radius);
  overflow: hidden;
  /* min-width: max-content; */
}

.filter-options {
  display: flex;
  gap: .4rem;
  align-items: center;
  padding: .2rem;
  height: max-content;
  justify-content: flex-end;
  /* background-color: goldenrod; */
  width: max-content;
  min-width: max-content;
}

.filter-root {
  width: 100%;
  display: flex;
  /* background-color: firebrick; */
  background-color: var(--black-steel);
  border-radius: var(--normal-radius);
  align-items: center;
  box-sizing: border-box;
  padding: .2rem;
  flex-direction: column;
}

.filter-header {
  width: 100%;
  display: flex;
  background-color: var(--steel);
  border-radius: var(--small-radius);
  /* background-color: green; */
  align-items: center;
  box-sizing: border-box;
  padding: 0rem .2rem;
}

.filter-header-tabs {
  box-sizing: border-box;
  width: 100%;
  width: min-content;
  /* flex: 1; */
  height: 100%;
  height: min-content;
  display: flex;
  /* background-color: purple; */
  align-items: center;
}

.selected-filters {
  box-sizing: border-box;
  width: 100%;
  overflow: hidden;
  /* flex: 1; */
  height: 100%;
  display: flex;
  /* background-color: goldenrod; */
  align-items: center;
}

.filter-tabs {
  width: 100%;
  height: max-content;
  height: 55px;
  /* min-height: 30px; */
  /* background-color: royalblue; */
  display: flex;
  flex-wrap: wrap;
  box-sizing: border-box;
  align-items: center;
  padding: .2rem .5rem;
}

.no-available-filters {
  display: flex;
  width: max-content;
  font-style: italic;
  opacity: .4;
  /* background-color: red; */
}

.no-active-filters {
  display: flex;
  width: max-content;
  font-style: italic;
  opacity: .4;
  color: var(--white);
  font-size: 14px
    /* background-color: red; */
}

.relayout {
  background-color: rebeccapurple;
  display: flex;
  margin: 10px 0;
  padding: 10px;
  background-color: #e0e0e0;
  color: black;
  border: 1px solid #ccc;
  border-radius: 5px;
  cursor: pointer;

}

.graph-container {
  position: relative;
  flex-grow: 1;
}

.fit-view-button {
  width: max-content;
  height: max-content;
  position: absolute;
  top: .5rem;
  right: .5rem;
  cursor: pointer;
  z-index: 1000;
}

.draggable-task {
  display: flex;
  margin: 10px 0;
  padding: 10px;
  background-color: #e0e0e0;
  background: goldenrod;
  color: black;
  border: 1px solid #ccc;
  border-radius: 5px;
  /* cursor: move; */
  cursor: pointer;

}
</style>


