import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { runInNewContext } from 'node:vm';
import { setImmediate } from 'node:timers/promises';
import { ref, watch, nextTick } from 'vue';

const component = readFileSync(new URL('./DependencyGraph.vue', import.meta.url), 'utf8');
const loader = component.slice(
  component.indexOf('const buildGraphFromDependencies ='),
  component.indexOf('const handleGraphSelectorUpdated ='),
);

function deferred() {
  let resolve;
  let reject;
  const promise = new Promise((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
}

function createHarness() {
  const calls = [];
  const notifications = [];
  const renders = [];
  const relationships = new Map();
  const versions = new Map();
  const edges = new Map();
  const request = (requests, id) => {
    if (!requests.has(id)) requests.set(id, deferred());
    return requests.get(id);
  };
  const context = {
    graphRequestId: 0,
    graphDisposed: false,
    projectStore: { activeProject: { uri: 'project' } },
    assetStore: { selectedAsset: { id: 'root', name: 'Root', type: 'asset' } },
    graphRootAsset: ref({ id: 'root', name: 'Root', type: 'asset' }),
    statusStore: { statuses: [] },
    utils: { capitalizeStr: value => value },
    getAppIcon: value => value,
    canActOnAsset: () => true,
    maxDepth: ref(0),
    FULL_DEPENDENCY_DEPTH: 0,
    isLoadingGraph: ref(false),
    graphLoadFailed: ref(false),
    graphConflictCount: ref(0),
    graphData: ref({ nodes: [], edges: [] }),
    getNodes: ref([]),
    dependencies: ref([]),
    totalAssetDepsCount: ref(0),
    canManageDependencies: ref(true),
    Position: { Left: 'left', Right: 'right' },
    nodeStyle: '',
    nextTick,
    t: key => key,
    console: { error() {} },
    notificationStore: { errorNotification: (...args) => notifications.push(args) },
    AssetService: {
      GetRecursiveDependencies: (_, id) => {
        calls.push(`relationships:${id}`);
        return request(relationships, id).promise;
      },
      ResolveDependencyGraphPlan: (_, id) => {
        calls.push(`versions:${id}`);
        return request(versions, id).promise;
      },
      GetAssetDependencyEdges: (_, id) => {
        calls.push(`edges:${id}`);
        return request(edges, id).promise;
      },
    },
  };
  watch(context.graphData, graph => {
    renders.push(graph);
    context.getNodes.value = graph.nodes.map(node => ({ ...node, data: { ...node.data } }));
  });
  runInNewContext(`${loader}\nthis.load = buildGraphFromDependencies;`, context);
  return { context, calls, notifications, relationships, versions, edges, renders };
}

const child = { asset: { id: 'child', name: 'Child', type: 'asset' }, parentId: 'root', depth: 1 };
const edge = { id: 'root-child', asset_id: 'root', dependency_id: 'child', resolution_mode: 'pinned' };
const plan = { entries: [{ asset_id: 'child', checkpoint_id: 'checkpoint', dependency_edge_id: edge.id }], conflicts: [] };

async function showRelationships(harness) {
  harness.relationships.get('root').resolve([child]);
  await setImmediate();
}

async function finishEdges(harness) {
  harness.edges.get('root').resolve([edge]);
  harness.edges.get('child').resolve([]);
  await setImmediate();
}

test('draws once with selectors and conflicts after parallel requests finish', async () => {
  const harness = createHarness();
  const pending = harness.context.load();
  assert.deepEqual(harness.calls, ['versions:root', 'relationships:root']);
  await showRelationships(harness);
  assert.equal(harness.renders.length, 0);
  assert.equal(harness.context.isLoadingGraph.value, true);
  assert.deepEqual(harness.calls.slice(2), ['edges:root', 'edges:child']);
  await finishEdges(harness);
  assert.equal(harness.renders.length, 0);
  harness.versions.get('root').resolve({ ...plan, conflicts: [{ asset_id: 'child' }] });
  await pending;
  await nextTick();
  assert.equal(harness.renders.length, 1);
  assert.equal(harness.context.getNodes.value.length, 2);
  assert.equal(harness.context.graphData.value.edges.length, 1);
  assert.equal(harness.context.getNodes.value[1].data.hasConflict, true);
  assert.equal(harness.context.graphConflictCount.value, 1);
  assert.equal(harness.context.getNodes.value[1].data.dependencyEdge.id, edge.id);
  assert.equal(harness.context.getNodes.value[0].data.canAdd, true);
  assert.equal(harness.context.getNodes.value[1].data.canAdd, false);
  assert.equal(harness.context.isLoadingGraph.value, false);
});

test('an early version failure preserves the previous graph without rendering partial results', async () => {
  const harness = createHarness();
  harness.context.graphData.value = { nodes: [{ id: 'previous', data: { id: 'previous' } }], edges: [] };
  await nextTick();
  const pending = harness.context.load();
  harness.versions.get('root').reject(new Error('version lookup failed'));
  await setImmediate();
  await showRelationships(harness);
  await finishEdges(harness);
  await pending;
  assert.equal(harness.renders.length, 1);
  assert.equal(harness.context.getNodes.value[0].id, 'previous');
  assert.equal(harness.context.graphLoadFailed.value, true);
  assert.equal(harness.context.isLoadingGraph.value, false);
  assert.equal(harness.notifications.length, 1);
});

test('late versions cannot overwrite a newer graph', async () => {
  const harness = createHarness();
  const oldRequest = harness.context.load();
  await showRelationships(harness);
  await finishEdges(harness);
  harness.context.assetStore.selectedAsset = { id: 'new-root', name: 'New', type: 'asset' };
  harness.context.graphRootAsset.value = harness.context.assetStore.selectedAsset;
  const newRequest = harness.context.load();
  harness.relationships.get('new-root').resolve([]);
  await setImmediate();
  harness.edges.get('new-root').resolve([]);
  harness.versions.get('new-root').resolve({ entries: [], conflicts: [] });
  await newRequest;
  harness.versions.get('root').resolve({ ...plan, conflicts: [{ asset_id: 'new-root' }] });
  await oldRequest;
  assert.equal(harness.context.getNodes.value.length, 1);
  assert.equal(harness.context.getNodes.value[0].data.entityId, 'new-root');
  assert.equal(harness.context.graphConflictCount.value, 0);
  assert.equal(harness.notifications.length, 0);
});

test('closing the graph ignores late failures', async () => {
  const harness = createHarness();
  const pending = harness.context.load();
  await showRelationships(harness);
  await finishEdges(harness);
  harness.context.graphDisposed = true;
  harness.versions.get('root').reject(new Error('late failure'));
  await pending;
  assert.equal(harness.notifications.length, 0);
});

for (const fullGraph of [false, true]) {
  test(fullGraph ? 'full graph retains direct links alongside nested links' : 'direct view shows three root links and skips full conflict resolution', async () => {
    const harness = createHarness();
    harness.context.maxDepth.value = fullGraph ? 0 : 1;
    const pending = harness.context.load();
    const dependencyIds = ['animation', 'fx', 'fx2'];
    harness.relationships.get('root').resolve(dependencyIds.map(id => ({
      asset: { id, name: id, type: 'asset' },
      parentId: 'root',
      parentIds: !fullGraph
        ? ['root']
        : id === 'animation' ? ['root', 'fx'] : id === 'fx2' ? ['root', 'animation'] : ['root'],
      depth: 1,
    })));
    await setImmediate();
    const directEdges = dependencyIds.map(id => ({ id: `root-${id}`, asset_id: 'root', dependency_id: id, resolution_mode: 'floating' }));
    harness.edges.get('root').resolve(directEdges);
    if (fullGraph) {
      harness.edges.get('fx').resolve([{ id: 'fx-animation', asset_id: 'fx', dependency_id: 'animation' }]);
      harness.edges.get('animation').resolve([{ id: 'animation-fx2', asset_id: 'animation', dependency_id: 'fx2' }]);
      harness.edges.get('fx2').resolve([]);
      harness.versions.get('root').resolve({ entries: [], conflicts: [] });
    } else {
      assert.deepEqual(harness.calls, ['relationships:root', 'edges:root']);
      assert.equal(harness.versions.size, 0);
    }
    await pending;
    await nextTick();
    const drawnEdges = harness.context.graphData.value.edges;
    assert.equal(drawnEdges.filter(item => item.source === 'root-root').length, 3);
    assert.equal(drawnEdges.length, fullGraph ? 6 : 3);
    assert.equal(
      drawnEdges.every(item => item.sourceHandle === 'output' && item.targetHandle === 'input'),
      true
    );
    assert.equal(harness.context.graphConflictCount.value, 0);
    if (fullGraph) {
      const animationNodes = harness.context.graphData.value.nodes
        .filter(item => item.data.entityId === 'animation');
      assert.equal(animationNodes.length, 2);
      assert.deepEqual(
        Array.from(animationNodes, item => item.data.dependencyEdge.id).sort(),
        ['fx-animation', 'root-animation'],
      );
    }
  });
}

test('full graph preserves every owner of a shared collection', async () => {
  const harness = createHarness();
  const pending = harness.context.load();
  harness.relationships.get('root').resolve([
    child,
    { collection: { id: 'group', name: 'Group', type: 'collection' }, parentId: 'root', parentIds: ['root', 'child'], depth: 1 },
  ]);
  await setImmediate();
  await finishEdges(harness);
  harness.versions.get('root').resolve(plan);
  await pending;
  const collectionNodes = harness.context.graphData.value.nodes.filter(item => item.data.entityId === 'group');
  assert.equal(collectionNodes.length, 2);
  assert.equal(collectionNodes.every(item => item.data.versionLabel === ''), true);
  const collectionNodeIds = new Set(collectionNodes.map(item => item.id));
  const collectionEdges = harness.context.graphData.value.edges.filter(item => collectionNodeIds.has(item.target));
  assert.equal(collectionEdges.length, 2);
  assert.equal(collectionNodes.every(item => item.data.hasIncoming && !item.data.hasOutgoing), true);
});
