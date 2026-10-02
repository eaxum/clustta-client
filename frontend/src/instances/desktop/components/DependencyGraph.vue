<template>
  <div class="page-list-root">
    <div class="page-list-container">
      <div class="dependency-graph-header">
        <div class="dependency-count"> {{ message }}</div>
        <div v-if="isLoadingGraph" role="status">{{ $t('common.loading') }}</div>
        <div v-else-if="graphLoadFailed" class="dependency-conflict-count">{{ $t('common.error') }}</div>
        <div v-if="graphConflictCount" class="dependency-conflict-count">
          {{ $t('components.dependencyGraph.conflicts', { count: graphConflictCount }) }}
        </div>
        <div class="dependency-toggles">
          <div class="dependency-toggle-container">
            <div class="input-label"> {{ $t('components.dependencyGraph.collectionContents') }}</div>
            <ToggleSwitch :switchValueProp="showCollectionContents" @click="changeCollectionContents()" />
          </div>
          <div class="dependency-toggle-container">
            <div class="input-label"> {{ $t('components.dependencyGraph.fullGraph') }}</div>
            <ToggleSwitch :switchValueProp="useMaxDepth" @click="changeDepth()" />
          </div>
        </div>
      </div>
      <div class="asset-graph-container">
        <div class="graph-container">
          <div class="fit-view-button">
            <ActionButton :icon="getAppIcon('arrows-expand')" v-tooltip="$t('components.dependencyGraph.fitView')" @click="fitViewToAllNodes()" />
          </div>
          <VueFlow v-model="graphElements" :default-viewport="{ zoom: 1 }" :fit-view-on-init="true"
            :max-zoom="1" :nodes-draggable="false" :no-drag-class-name="noDragClassName">
            <Background :size="1" :gap="20" pattern-color="#BDBDBD" />
            <!-- <MiniMap /> -->
            <template #node-custom="props">
              <DependencyGraphNode :data="props.data" @assign="openAssignmentMenu"
                @add="openDependencyPane" @navigate="goToGraphItem" @remove="removeGraphDependency"
                @selectorUpdated="handleGraphSelectorUpdated" />
            </template>
          </VueFlow>
        </div>
      </div>
    </div>

    <Transition name="dependency-picker" @after-enter="fitViewToAllNodes" @after-leave="fitViewToAllNodes">
      <aside v-if="showDependencyPicker" class="dependency-picker expandable-panel">
        <ExpandablePanelHeader v-model="dependencySearchQuery"
          title="" closeIcon="chevron-right" :showMaximize="false" :showTitle="false"
          :filterPlaceholder="$t('panes.searchDependencies')"
          @close="showDependencyPicker = false" />
        <div class="dependency-picker-list">
          <div v-if="isLoadingSidebar" class="dependency-picker-state">{{ $t('common.loading') }}</div>
          <ItemsList v-else-if="availableDependencies.length" :items="availableDependencies"
            :isDependency="true" :forList="true" :showAdd="canManageDependencies" />
          <PageState v-else class="dependency-picker-empty"
            :message="dependencySearchQuery ? $t('panes.noItemsMatchSearch') : $t('components.dependencyGraph.noMoreDependenciesToAdd')"
            illustration="/page-states/template.png" />
        </div>
      </aside>
    </Transition>

  </div>
</template>


<script setup>
// imports
import { ref, computed, onMounted, watch, nextTick, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n';
import dagre from '@dagrejs/dagre'
import { AssetService, CollectionService } from "@/services";
import emitter from '@/lib/mitt';
import utils from '@/services/utils';
import { canActOnAsset } from '@/lib/permissions';

// vue flow
import { VueFlow, useVueFlow } from '@vue-flow/core'
import { Background } from '@vue-flow/background';

// styles
import '@vue-flow/core/dist/style.css';
import '@vue-flow/core/dist/theme-default.css';

// state imports
import { useCommonStore } from '@/stores/common';
import { useCollectionStore } from '@/stores/collections';
import { useAssetStore } from '@/stores/assets';
import { useNotificationStore } from '@/stores/notifications';
import { useDependencyStore } from '@/stores/dependency';
import { useMenu } from '@/stores/menu';
import { useIconStore } from '@/stores/icons';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useStatusStore } from '@/stores/status';
import { useDesktopModalStore } from '@/stores/desktopModals';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ItemsList from '@/instances/desktop/components/ItemsList.vue';
import DependencyGraphNode from '@/instances/desktop/components/DependencyGraphNode.vue';
import ExpandablePanelHeader from '@/instances/desktop/components/ExpandablePanelHeader.vue';
import PageState from '@/instances/common/components/PageState.vue';

// states
const notificationStore = useNotificationStore();
const dependencyStore = useDependencyStore();
const commonStore = useCommonStore();
const collectionStore = useCollectionStore();
const assetStore = useAssetStore();
const projectStore = useProjectStore();
const iconStore = useIconStore();
const menu = useMenu();
const stage = useStageStore();
const statusStore = useStatusStore();
const modals = useDesktopModalStore();

const { t } = useI18n();

// refs
const useMaxDepth = ref(false);
const showCollectionContents = ref(true);
const graphElements = ref([]);
const noDragClassName = 'no-drag';
const sidebarAssets = ref([]);
const sidebarCollections = ref([]);
const dependencySearchQuery = ref('');
const showDependencyPicker = ref(false);
const graphData = ref({ nodes: [], edges: [] });
const isLoadingGraph = ref(false);
const isLoadingSidebar = ref(false);
const graphConflictCount = ref(0);
const graphLoadFailed = ref(false);
const graphRootAsset = ref(assetStore.selectedAsset);
let graphRequestId = 0;
let graphDisposed = false;

// vars
const { fitView, setNodes, setEdges } = useVueFlow();
const smoothNode = false;
const nodeStyle = smoothNode ? 'smoothstep' : '';

// computed props
const DIRECT_DEPENDENCY_DEPTH = 1;
const FULL_DEPENDENCY_DEPTH = 0;
const COLLECTION_ASSET_COLLAPSE_THRESHOLD = 3;
const maxDepth = computed(() => useMaxDepth.value ? FULL_DEPENDENCY_DEPTH : DIRECT_DEPENDENCY_DEPTH);
const canManageDependencies = computed(() => canActOnAsset('manage_dependencies', graphRootAsset.value));

// methods
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

// //////////////////////////////
const message = computed(() => {
  if (!useMaxDepth.value) {
    if (totalAssetDeps.value > 1) {
      return t('components.dependencyGraph.directDependencies', { count: totalAssetDeps.value });
    } else if (totalAssetDeps.value === 1) {
      return t('components.dependencyGraph.directDependency', { count: totalAssetDeps.value });
    } else {
      return t('components.dependencyGraph.noDependencies');
    }
  } else {
    if (totalAssetDeps.value > 1) {
      return t('components.dependencyGraph.totalDependencies', { count: totalAssetDeps.value });
    } else if (totalAssetDeps.value === 1) {
      return t('components.dependencyGraph.totalDependency', { count: totalAssetDeps.value });
    } else {
      return t('components.dependencyGraph.noDependencies');
    }
  }
});



const dependencies = ref([]);
const totalAssetDepsCount = ref(0);
const totalAssetDeps = computed(() => { return totalAssetDepsCount.value });
const availableDependencies = computed(() => {
  const rootAsset = graphRootAsset.value;
  if (!rootAsset) return [];
  const query = dependencySearchQuery.value.trim().toLowerCase();
  const currentDependencies = new Set(dependencies.value);
  const items = [
    ...sidebarAssets.value.map(item => ({ ...item, type: 'asset' })),
    ...sidebarCollections.value.map(item => ({ ...item, type: 'collection' })),
  ];
  return items
    .filter((item) => {
      if (item.id === rootAsset.id || currentDependencies.has(item.id)) return false;
      const itemDependencies = [...(item.dependencies || []), ...(item.collection_dependencies || [])];
      if (itemDependencies.includes(rootAsset.id)) return false;
      if (!query) return true;
      const searchableText = `${item.name || ''} ${item.asset_path || item.collection_path || ''}`.toLowerCase();
      return searchableText.includes(query);
    })
    .sort((first, second) => String(first.name || '').localeCompare(String(second.name || '')));
});

// graph methods
const changeDepth = async () => {
  useMaxDepth.value = !useMaxDepth.value;
  await buildGraphFromDependencies();
  nextTick(() => {
    fitViewToAllNodes(true);
  })
};

const changeCollectionContents = async () => {
  showCollectionContents.value = !showCollectionContents.value;
  await buildGraphFromDependencies();
  nextTick(() => {
    fitViewToAllNodes(true);
  });
};

const fitViewToAllNodes = (useDelay = false) => {
  const timeOut = useDelay ? 400 : 0;
  setTimeout(() => {
    fitView({ padding: 0.1, includeHiddenNodes: false, duration: 200, maxZoom: 1 })
  }, timeOut);
};

const fetchSidebarData = async () => {
  isLoadingSidebar.value = true;
  try {
    const projectPath = projectStore.activeProject.uri;
    
    const [assetsResult, collectionsResult] = await Promise.all([
      AssetService.GetAssets(projectPath),
      CollectionService.GetCollections(projectPath)
    ]);
    
    sidebarAssets.value = assetsResult || [];
    sidebarCollections.value = collectionsResult || [];
  } catch (error) {
    console.error("Error fetching sidebar data:", error);
    notificationStore.errorNotification(t('components.dependencyGraph.errorLoadingProject'), error);
  } finally {
    isLoadingSidebar.value = false;
  }
};

const buildGraphFromDependencies = async () => {
  const requestId = ++graphRequestId;
  const projectPath = projectStore.activeProject?.uri;
  const depth = maxDepth.value;
  const fullGraph = depth === FULL_DEPENDENCY_DEPTH;
  const selectedAsset = graphRootAsset.value;
  const isCurrentRequest = () => !graphDisposed && requestId === graphRequestId
    && projectStore.activeProject?.uri === projectPath
    && graphRootAsset.value?.id === selectedAsset?.id;
  isLoadingGraph.value = true;
  graphLoadFailed.value = false;
  graphConflictCount.value = 0;
  
  if (!selectedAsset || !projectPath) {
    graphData.value = { nodes: [], edges: [] };
    dependencies.value = [];
    totalAssetDepsCount.value = 0;
    isLoadingGraph.value = false;
    return;
  }

  // Capture rejection immediately, even while the relationship request is pending.
  const loadGraphPlan = async () => {
    try {
      return { plan: await AssetService.ResolveDependencyGraphPlan(projectPath, selectedAsset.id) };
    } catch (error) {
      return { error };
    }
  };
  const graphPlanPromise = fullGraph ? loadGraphPlan() : Promise.resolve({ plan: { entries: [], conflicts: [] } });

  try {
    const dependencyItems = await AssetService.GetRecursiveDependencies(
      projectPath,
      selectedAsset.id,
      depth,
      showCollectionContents.value,
    );
    if (!isCurrentRequest()) return;

    const entitiesById = new Map([[selectedAsset.id, selectedAsset]]);
    const entityTypesById = new Map([[selectedAsset.id, 'asset']]);
    const directAssetCountsByCollectionId = new Map();
    let collapsedAssetCount = 0;
    const relationshipsByParentId = new Map();
    dependencyItems.forEach(item => {
      const entity = item.asset || item.collection || item;
      const entityType = item.collection ? 'collection' : 'asset';
      entitiesById.set(entity.id, entity);
      entityTypesById.set(entity.id, entityType);
      if (entityType === 'collection') {
        directAssetCountsByCollectionId.set(entity.id, item.directAssetCount || 0);
        collapsedAssetCount += item.collapsedAssetCount || 0;
      }
      const parentIds = item.parentIds?.length ? item.parentIds : [item.parentId || selectedAsset.id];
      parentIds.forEach(parentId => {
        if (!relationshipsByParentId.has(parentId)) relationshipsByParentId.set(parentId, []);
        relationshipsByParentId.get(parentId).push({ childId: entity.id, entityType });
      });
    });

    const assetIds = [selectedAsset.id];
    if (fullGraph) {
      for (const [entityId, entityType] of entityTypesById) {
        if (entityType === 'asset' && entityId !== selectedAsset.id) assetIds.push(entityId);
      }
    }
    const edgeGroups = await Promise.all(assetIds.map(assetId => (
      AssetService.GetAssetDependencyEdges(projectPath, assetId)
    )));
    if (!isCurrentRequest()) return;
    const selectorEdges = edgeGroups.flat();
    const selectorEdgesByPair = new Map(selectorEdges.map(edge => [
      `${edge.asset_id}:${edge.dependency_id}`,
      edge,
    ]));

    const graphPlanResult = await graphPlanPromise;
    if (!isCurrentRequest()) return;
    if ('error' in graphPlanResult) throw graphPlanResult.error;
    const graphPlan = graphPlanResult.plan;
    const conflictingAssetIds = new Set(graphPlan.conflicts.map(conflict => conflict.asset_id));
    const conflictMessagesByAssetId = new Map(graphPlan.conflicts.map(conflict => [
      conflict.asset_id,
      conflict.message,
    ]));
    const statusesById = new Map(statusStore.statuses.map(status => [status.id, status]));
    const nodes = [];
    const edges = [];
    let occurrenceIndex = 0;

    const createNodeData = (entity, entityType, relationship, occurrenceDepth) => {
      const dependencyEdge = relationship?.dependencyEdge || null;
      const resolutionWarning = dependencyEdge?.resolution_status !== 'ready'
        ? dependencyEdge?.resolution_status?.replace(/_/g, ' ')
        : '';
      const status = statusesById.get(entity.status_id);
      return {
        rawEntity: entity,
        entityId: entity.id,
        entityType,
        name: entity.name,
        path: entity.asset_path || entity.collection_path || entity.name,
        extension: entity.extension || '',
        icon: getAppIcon(entity.collection_type_icon || entity.asset_type_icon || (entityType === 'collection' ? 'folder' : 'file')),
        statusLabel: entityType === 'asset' ? utils.capitalizeStr(entity.status_short_name || status?.short_name || '') : '',
        statusColor: entity.status?.color || status?.color || '',
        assigneeId: entity.assignee_id || '',
        assigneeName: entity.assignee_name || '',
        dependencyEdge,
        versionLabel: '',
        itemCountLabel: entityType === 'collection'
          ? t('blocks.itemCount', directAssetCountsByCollectionId.get(entity.id) || 0)
          : '',
        depth: occurrenceDepth,
        hasIncoming: occurrenceDepth > 0,
        hasOutgoing: false,
        hasConflict: conflictingAssetIds.has(entity.id),
        warning: conflictMessagesByAssetId.get(entity.id) || resolutionWarning,
        canAssign: entityType === 'asset'
          && (canActOnAsset('assign_asset', entity) || canActOnAsset('unassign_asset', entity)),
        canAdd: occurrenceDepth === 0 && canManageDependencies.value,
        canEditSelector: !!dependencyEdge && canManageDependencies.value
          && dependencyEdge.asset_id === selectedAsset.id,
        canRemove: !!relationship && occurrenceDepth === 1 && canManageDependencies.value,
        ownerAssetId: relationship?.parentEntityId || '',
      };
    };

    const isDirectCollectionContent = (collectionId, relationship) => {
      const child = entitiesById.get(relationship.childId);
      if (!child) return false;
      if (relationship.entityType === 'collection') return child.parent_id === collectionId;
      return child.collection_id === collectionId;
    };

    const addCollectionAssetStack = (collection, collectionNodeId, occurrenceDepth, assetCount) => {
      const nodeId = `collection-assets-${occurrenceIndex++}-${collection.id}`;
      nodes.push({
        id: nodeId,
        label: t('components.dependencyGraph.assetCount', { count: assetCount }),
        position: { x: 0, y: 0 },
        type: 'custom',
        data: {
          rawEntity: collection,
          entityId: collection.id,
          entityType: 'collection-assets',
          name: t('components.dependencyGraph.assetCount', { count: assetCount }),
          path: collection.collection_path || collection.name,
          extension: '',
          icon: getAppIcon('file'),
          statusLabel: '',
          statusColor: '',
          assigneeId: '',
          assigneeName: '',
          dependencyEdge: null,
          versionLabel: '',
          itemCountLabel: '',
          depth: occurrenceDepth + 1,
          hasIncoming: true,
          hasOutgoing: false,
          hasConflict: false,
          warning: '',
          canAssign: false,
          canAdd: false,
          canEditSelector: false,
          canRemove: false,
          ownerAssetId: collection.id,
          isAssetStack: true,
        },
      });
      edges.push({
        id: `collection-assets-${collectionNodeId}-${nodeId}`,
        source: collectionNodeId,
        target: nodeId,
        sourceHandle: 'output',
        targetHandle: 'input',
        type: nodeStyle,
      });
    };

    const addOccurrence = (entityId, parentNodeId, relationship, pathIds, occurrenceDepth) => {
      const entity = entitiesById.get(entityId);
      if (!entity) return null;
      const entityType = entityTypesById.get(entityId) || 'asset';
      const nodeId = occurrenceDepth === 0 ? `root-${entityId}` : `occurrence-${occurrenceIndex++}-${entityId}`;
      const node = {
        id: nodeId,
        label: entity.name,
        position: { x: 0, y: 0 },
        type: 'custom',
        data: createNodeData(entity, entityType, relationship, occurrenceDepth),
      };
      nodes.push(node);
      if (parentNodeId) {
        edges.push({
          id: `${relationship.dependencyEdge?.id || 'relationship'}-${parentNodeId}-${nodeId}`,
          source: parentNodeId,
          target: nodeId,
          sourceHandle: 'output',
          targetHandle: 'input',
          type: nodeStyle,
        });
      }

      const childRelationships = relationshipsByParentId.get(entityId) || [];
      const directAssetCount = entityType === 'collection'
        ? directAssetCountsByCollectionId.get(entityId) || 0
        : 0;
      const shouldCollapseAssets = showCollectionContents.value
        && directAssetCount > COLLECTION_ASSET_COLLAPSE_THRESHOLD;
      if (shouldCollapseAssets) {
        addCollectionAssetStack(entity, nodeId, occurrenceDepth, directAssetCount);
      }
      for (const childRelationship of childRelationships) {
        const isCollectionContent = entityType === 'collection'
          && isDirectCollectionContent(entityId, childRelationship);
        if (isCollectionContent && childRelationship.entityType === 'asset' && shouldCollapseAssets) {
          continue;
        }
        if (pathIds.has(childRelationship.childId)) continue;
        const dependencyEdge = selectorEdgesByPair.get(`${entityId}:${childRelationship.childId}`) || null;
        const nextPathIds = new Set(pathIds);
        nextPathIds.add(childRelationship.childId);
        addOccurrence(
          childRelationship.childId,
          nodeId,
          { ...childRelationship, dependencyEdge, parentEntityId: entityId },
          nextPathIds,
          occurrenceDepth + 1,
        );
      }
      node.data.hasOutgoing = edges.some(edge => edge.source === nodeId);
      return node;
    };

    addOccurrence(selectedAsset.id, '', null, new Set([selectedAsset.id]), 0);
    dependencies.value = dependencyItems.map(item => (item.asset || item.collection || item).id);
    totalAssetDepsCount.value = dependencyItems.length + collapsedAssetCount;
    graphConflictCount.value = graphPlan.conflicts.length;
    graphData.value = { nodes, edges };
  } catch (error) {
    if (!isCurrentRequest()) return;
    graphLoadFailed.value = true;
    console.error("Error building graph:", error);
    notificationStore.errorNotification(t('components.dependencyGraph.errorBuildingGraph'), error);
  } finally {
    if (isCurrentRequest()) {
      isLoadingGraph.value = false;
    }
  }
};

const handleGraphSelectorUpdated = async () => {
  await buildGraphFromDependencies();
};

const NODE_WIDTH = 310;
const ASSET_NODE_HEIGHT = 86;
const COLLECTION_NODE_HEIGHT = 52;
const COLLECTION_ASSET_STACK_NODE_HEIGHT = 72;

const applyDagreLayout = (nodes, edges) => {
  const g = new dagre.graphlib.Graph()
  g.setGraph({
    rankdir: 'LR',
    nodesep: 45,
    ranksep: 110,
    edgesep: 20,
  })
  g.setDefaultEdgeLabel(() => ({}))

  nodes.forEach(node => {
    const height = node.data.isAssetStack
      ? COLLECTION_ASSET_STACK_NODE_HEIGHT
      : node.data.entityType === 'collection' ? COLLECTION_NODE_HEIGHT : ASSET_NODE_HEIGHT;
    g.setNode(node.id, { width: NODE_WIDTH, height })
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
  setNodes(nodesWithLayout)
  setEdges(newGraph.edges)
};

// Watch for graph data changes and update layout
watch(graphData, (newGraphData) => {
  updateGraphLayout(newGraphData);
}, { immediate: true });

watch(() => projectStore.activeProject?.uri, () => {
  buildGraphFromDependencies();
});

const selectAsset = async (assetId) => {
  // Find asset in our cached sidebar data first
  let asset = sidebarAssets.value.find(item => item.id === assetId);
  
  if (!asset) {
    // If not found in sidebar, try to fetch it from the service
    try {
      const allAssets = await AssetService.GetAssets(projectStore.activeProject.uri);
      asset = allAssets.find(item => item.id === assetId);
    } catch (error) {
      console.error("Error fetching asset:", error);
      notificationStore.errorNotification("Error selecting asset", error);
      return;
    }
  }
  
  if (asset) {
    graphRootAsset.value = asset;
    assetStore.selectAsset(asset);
    await fetchSidebarData()
    await buildGraphFromDependencies();
    
    nextTick(() => {
      fitViewToAllNodes(true);
    });
  }
};

const addDependency = async (dependencyId, itemType) => {
  const asset = graphRootAsset.value;
  const allDependencies = [...sidebarAssets.value, ...sidebarCollections.value];

  let dependencyTypeID = dependencyStore.dependency_types.find(item => item.name === "linked").id;
  if (itemType === "asset") {
    await AssetService.AddAssetDependency(projectStore.activeProject.uri, asset.id, dependencyId, dependencyTypeID)
      .then( async(response) => {
        notificationStore.addNotification(t('components.dependencyGraph.dependencyAdded'), "", "success");
        const addedDependency = allDependencies.find((newDependency) => newDependency.id === dependencyId);
        if (addedDependency) {
          dependencies.value.push(addedDependency.id);
          graphRootAsset.value.dependencies = dependencies.value;
          await buildGraphFromDependencies();
          nextTick(() => {
            fitViewToAllNodes();
          })
        }
      })
      .catch((error) => {
        console.log(error)
        notificationStore.errorNotification(t('components.dependencyGraph.errorAddingDependencies'), error);
      });
  } else {
    await AssetService.AddCollectionDependency(projectStore.activeProject.uri, asset.id, dependencyId, dependencyTypeID)
      .then( async(response) => {
        notificationStore.addNotification(t('components.dependencyGraph.dependencyAdded'), "", "success");
        const addedDependency = allDependencies.find((newDependency) => newDependency.id === dependencyId);
        if (addedDependency) {
          dependencies.value.push(addedDependency.id);
          graphRootAsset.value.collection_dependencies = dependencies.value;
          
          await buildGraphFromDependencies();
          nextTick(() => {
            fitViewToAllNodes();
          })
        }
      })
      .catch((error) => {
        console.log(error)
        notificationStore.errorNotification(t('components.dependencyGraph.errorAddingDependencies'), error);
      });
  }
};

const removeDependency = async (dependencyId, itemType) => {
  const asset = graphRootAsset.value;
  if (itemType === "asset") {
    await AssetService.RemoveAssetDependency(projectStore.activeProject.uri, asset.id, dependencyId)
      .then(async(response) => {
        notificationStore.addNotification(t('components.dependencyGraph.dependencyRemoved'), "", "success");
        dependencies.value = dependencies.value.filter(id => id !== dependencyId);
        graphRootAsset.value.dependencies = dependencies.value;
        await buildGraphFromDependencies();
        nextTick(() => {
          fitViewToAllNodes();
        })
      })
      .catch((error) => {
        notificationStore.errorNotification(t('components.dependencyGraph.errorRemovingDependencies'), error);
      });
  } else {
    await AssetService.RemoveCollectionDependency(projectStore.activeProject.uri, asset.id, dependencyId)
      .then(async(response) => {
        notificationStore.addNotification(t('components.dependencyGraph.dependencyRemoved'), "", "success");
        dependencies.value = dependencies.value.filter(id => id !== dependencyId);
        graphRootAsset.value.collection_dependencies = dependencies.value;
        buildGraphFromDependencies();
        nextTick(() => {
          fitViewToAllNodes();
        })
      })
      .catch((error) => {
        notificationStore.errorNotification(t('components.dependencyGraph.errorRemovingDependencies'), error);
      });
  }
};

const handleSelectItem = (payload) => {
  const id = payload.message;
  selectAsset(id);
};

const handleAddDependency = (payload) => {
  console.log(payload)
  addDependency(payload.id, payload.itemType);
};

const handleRemoveDependency = (payload) => {
  removeDependency(payload.id, payload.itemType);
};

const removeGraphDependency = (nodeData) => {
  removeDependency(nodeData.entityId, nodeData.entityType);
};

const openDependencyPane = async () => {
  if (!graphRootAsset.value || !canManageDependencies.value) return;
  dependencySearchQuery.value = '';
  showDependencyPicker.value = true;
  if (!sidebarAssets.value.length && !sidebarCollections.value.length) {
    await fetchSidebarData();
  }
};

const openAssignmentMenu = (nodeData, event) => {
  if (!nodeData.canAssign || nodeData.entityType !== 'asset') return;
  assetStore.selectAsset(nodeData.rawEntity);
  stage.markedItems = [nodeData.entityId];
  stage.markedAssets = [nodeData.entityId];
  menu.showContextMenu(event, 'assignMenu', true);
};

const goToGraphItem = async (nodeData) => {
  try {
    const item = nodeData.rawEntity;
    const isCollection = nodeData.entityType === 'collection' || nodeData.entityType === 'collection-assets';
    const collection = isCollection ? item : item.collection_id
      ? await CollectionService.GetCollectionByID(projectStore.activeProject.uri, item.collection_id)
      : null;
    commonStore.activeWorkspace = 'Project';
    commonStore.viewSearchQuery = '';
    commonStore.resetFilters();
    commonStore.navigatorMode = true;
    stage.deselectAllItems();
    collectionStore.navigateToCollection(collection);
    if (isCollection) {
      collectionStore.selectCollection(item);
    } else {
      assetStore.selectAsset(item);
    }
    stage.firstSelectedItemId = item.id;
    stage.markedItems = [item.id];
    emitter.emit('view-details');
    emitter.emit('refresh-browser');
    modals.setModalVisibility('dependencyGraphModal', false);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.failedToNavigate'), error);
  }
};

const handleGraphDataUpdated = ({ itemId }) => {
  if (!graphData.value.nodes.some(node => node.data.entityId === itemId)) return;
  void buildGraphFromDependencies();
};

onMounted(async () => {
  emitter.on('selectItem', handleSelectItem);
  emitter.on('addDependency', handleAddDependency);
  emitter.on('removeDependency', handleRemoveDependency);
  emitter.on('dependency-selector-updated', handleGraphSelectorUpdated);
  emitter.on('update-root-data', handleGraphDataUpdated);
  if (!statusStore.statuses.length) {
    try {
      await statusStore.reloadStatuses();
    } catch (error) {
      notificationStore.errorNotification(t('notifications.errorLoadingProjectData'), error);
    }
  }
  void fetchSidebarData();
  void buildGraphFromDependencies();
});

onUnmounted(() => {
  graphDisposed = true;
  graphRequestId += 1;
  emitter.off('selectItem', handleSelectItem)
  emitter.off('addDependency', handleAddDependency)
  emitter.off('removeDependency', handleRemoveDependency)
  emitter.off('dependency-selector-updated', handleGraphSelectorUpdated)
  emitter.off('update-root-data', handleGraphDataUpdated)
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
  --dependency-graph-spacing: .5rem;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  color: var(--text);
  justify-content: flex-start;
  gap: var(--dependency-graph-spacing);
  height: 100%;
  overflow: hidden;
  padding: var(--dependency-graph-spacing);
  position: relative;
  width: 100%;
}

.page-list-container {
  flex: 1;
  min-width: 0;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
  flex-direction: column;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text);
  padding: 0;
}

.dependency-picker {
  flex: 0 0 360px;
  width: 360px;
  height: 100%;
  outline: var(--transparent-line);
  animation: none;
}

.dependency-picker-list {
  display: flex;
  flex: 1;
  min-height: 0;
  padding: .5rem;
  overflow: hidden;
}

.dependency-picker-list :deep(.virtual-scroll-container) {
  flex: 1;
  min-height: 0;
}

.dependency-picker-state {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  color: var(--text-muted);
  font-size: 12px;
  text-align: center;
}

.dependency-picker-empty {
  width: 100%;
  height: 100%;
}

.dependency-picker-empty :deep(.page-state-illustration) {
  width: 80%;
  max-width: 220px;
}

.dependency-picker-enter-active,
.dependency-picker-leave-active {
  transition: opacity .2s ease-out, transform .2s ease-out;
}

.dependency-picker-enter-from,
.dependency-picker-leave-to {
  opacity: 0;
  transform: translateX(2rem);
}

.asset-graph-container {
  background-color: var(--surface-1);
  border-radius: var(--large-radius);
  display: flex;
  width: 100%;
  height: 100%;
  border-radius: var(--very-large-radius);
  overflow: hidden;
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

.dependency-toggles {
  display: flex;
  align-items: center;
  gap: 1rem;
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
  color: var(--text);
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
  background-color: var(--surface-1);
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
  background-color: var(--surface-1);
  border-radius: var(--large-radius);
  border-radius: var(--very-large-radius);
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
  background-color: var(--surface-inverse);
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
  color: var(--text);
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
  background-color: var(--surface-3);
  color: var(--text);
  transition: width 0.2s ease-out;
  width: 100%;
  max-width: 400px;
  border-radius: var(--very-large-radius);
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
  background-color: var(--surface-1);
  border-radius: var(--normal-radius);
  align-items: center;
  box-sizing: border-box;
  padding: .2rem;
  flex-direction: column;
}

.filter-header {
  width: 100%;
  display: flex;
  background-color: var(--surface-3);
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
  color: var(--text);
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

.draggable-asset {
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

.dependency-conflict-count {
  color: var(--danger);
  font-size: .72rem;
  font-weight: 600;
}

:deep(.dependency-conflict-node) {
  border: 2px solid var(--danger);
  border-radius: .4rem;
}

:deep(.vue-flow__edge-path) {
  stroke: var(--text-muted);
  stroke-width: 1.5;
}
</style>
