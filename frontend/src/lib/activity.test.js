import { activityTransferProgress } from './activity.js';
import { sentenceCase, activityTransferSizes, activitySavedSize } from './activity.js';
import { formatError } from './errors.js';
import assert from 'node:assert/strict';
import test from 'node:test';
import { buildPendingTransfers, isTransferPending, activityDisplayName, applyActivitySnapshot, filterActivityOperations, sortActivityOperations } from './activity.js';

test('activity summaries show the requested name', () => {
  const assetDownload = { title: 'Downloading checkpoint', assets: [{ name: 'Scene', extension: '.blend' }], extra_message: 'Data saved: 127.22 MB (22.39%)' };
  assert.equal(activityDisplayName(assetDownload), 'Scene.blend');
  assert.equal(activityDisplayName({ title: 'Fetching ajani', assets: [] }), 'Ajani');
});

test('activity search includes requested assets, projects and status without changing order', () => {
  const items = [
    { title: 'Fetching characters', project_name: 'Meji', status: 'running', assets: [{ name: 'Scene', extension: '.blend' }] },
    { title: 'Fetching textures', project_name: 'Other', status: 'completed', assets: [] },
  ];
  assert.deepEqual(filterActivityOperations(items, ' SCENE.BLEND '), [items[0]]);
  assert.deepEqual(filterActivityOperations(items, 'meji'), [items[0]]);
  assert.deepEqual(filterActivityOperations(items, 'completed'), [items[1]]);
  assert.deepEqual(filterActivityOperations(items, 'fetching'), items);
  assert.deepEqual(filterActivityOperations(items, 'missing'), []);
});

test('running operations appear first with newest first in each group', () => {
  const operations = [
    { operation_id: 'old-running', status: 'running' },
    { operation_id: 'old-finished', status: 'completed' },
    { operation_id: 'new-running', status: 'running' },
    { operation_id: 'new-finished', status: 'cancelled' },
  ];
  assert.deepEqual(sortActivityOperations(operations).map((item) => item.operation_id),
    ['new-running', 'old-running', 'new-finished', 'old-finished']);
  assert.equal(operations[0].operation_id, 'old-running');
});

const state = () => ({ operations: [], revision: -1, batchIDs: [], expanded: false });

test('queued actions remain pending, preserve FIFO order, and do not trigger completion effects', () => {
  const activity = state();
  const queued = (id) => ({ operation_id: id, status: 'queued', project_uri: 'project-a', assets: [{ id: 'asset' }] });
  const snapshot = { revision: 1, operations: [{ operation_id: 'a', status: 'running' }, queued('b'), queued('c')] };
  assert.deepEqual(applyActivitySnapshot(activity, snapshot), []);
  assert.deepEqual(sortActivityOperations(snapshot.operations).map((item) => item.operation_id), ['a', 'b', 'c']);
  assert.equal(isTransferPending(buildPendingTransfers([queued('b')]), 'project-a', { id: 'asset', type: 'asset' }), true);
});
const operation = (id, unused = false, status = 'running') => ({ operation_id: id, status });

test('new actions keep the panel collapsed and batches advance', () => {
  const activity = state();
  applyActivitySnapshot(activity, { revision: 1, operations: [operation('a')] });
  assert.equal(activity.expanded, false);
  activity.expanded = false;
  applyActivitySnapshot(activity, { revision: 2, operations: [operation('a')] });
  assert.equal(activity.expanded, false);
  applyActivitySnapshot(activity, { revision: 3, operations: [operation('a'), operation('b', false, 'queued')] });
  assert.equal(activity.expanded, false);
  assert.deepEqual(activity.batchIDs, ['a', 'b']);
  applyActivitySnapshot(activity, { revision: 4, operations: [operation('a', false, 'completed'), operation('b')] });
  assert.equal(activity.batchIDs.indexOf('b') + 1, 2);
  applyActivitySnapshot(activity, { revision: 5, operations: [operation('b', false, 'completed')] });
  applyActivitySnapshot(activity, { revision: 6, operations: [operation('b', false, 'completed'), operation('c')] });
  assert.deepEqual(activity.batchIDs, ['c']);
});

test('completion effects run once and an older snapshot cannot resurrect finished work', () => {
  const activity = state();
  applyActivitySnapshot(activity, { revision: 1, operations: [operation('a', true), operation('b')] });
  const completed = { revision: 2, operations: [operation('a', true, 'completed'), operation('b')] };
  assert.deepEqual(applyActivitySnapshot(activity, completed).map((item) => item.operation_id), ['a']);
  assert.deepEqual(applyActivitySnapshot(activity, completed), []);
  applyActivitySnapshot(activity, { revision: 1, operations: [operation('a', true)] });
  assert.equal(activity.operations[0].status, 'completed');
});

test('restoring a session does not expand Activity', () => {
  const activity = state();
  assert.deepEqual(applyActivitySnapshot(activity, { revision: 4, operations: [operation('a', true, 'completed')] }, true), []);
  activity.expanded = true;
  applyActivitySnapshot(activity, { revision: 5, operations: [operation('a', true, 'completed'), operation('b', true)] });
});

const asset = { id: 'asset', collection_id: 'aremu', collection_path: '/Assets/characters/aremu/' };
const pendingOperation = (overrides = {}) => ({ status: 'running', project_uri: 'project-a', assets: [asset], ...overrides });

test('pending identifies assets and collapsed ancestors without loaded children', () => {
  const pending = buildPendingTransfers([pendingOperation()]);
  assert.equal(isTransferPending(pending, 'project-a', { id: 'asset', type: 'asset' }), true);
  assert.equal(isTransferPending(pending, 'project-b', { id: 'asset', type: 'asset' }), false);
  assert.equal(isTransferPending(pending, 'project-a', { id: 'characters', type: 'collection', collection_path: '\\Assets\\characters\\' }), true);
  assert.equal(isTransferPending(pending, 'project-a', { id: 'other', type: 'collection', collection_path: '/Assets/characters/aremu-other/' }), false);
});

test('restored and terminal work clears pending unless another action still needs the asset', () => {
  const item = { id: 'asset', type: 'asset' };
  for (const status of ['completed', 'cancelled', 'failed']) {
    assert.equal(isTransferPending(buildPendingTransfers([pendingOperation({ status })]), 'project-a', item), false);
  }
  const restored = pendingOperation({ restored_asset_ids: ['asset'] });
  assert.equal(isTransferPending(buildPendingTransfers([restored]), 'project-a', item), false);
  assert.equal(isTransferPending(buildPendingTransfers([restored, pendingOperation()]), 'project-a', item), true);
});

test('requested collections stay pending during preparation before assets are resolved', () => {
  const pending = buildPendingTransfers([pendingOperation({ assets: [], collections: [{ id: 'aremu', collection_path: '/Assets/characters/aremu/' }] })]);
  assert.equal(isTransferPending(pending, 'project-a', { id: 'aremu', type: 'collection' }), true);
  assert.equal(isTransferPending(pending, 'project-a', { id: 'characters', type: 'collection', collection_path: '/Assets/characters/' }), true);
});

test('errors preserve actionable details and remove signed download URLs', () => {
 assert.match(formatError(new Error('Get "https://server/chunk?signature=secret": connectex: A socket operation was attempted to an unreachable network.')), /Check your connection/);
 assert.equal(formatError('{"error":{"message":"Access denied to local file"}}'), 'Access denied to local file');
 assert.equal(formatError('Get "https://server/chunk?signature=secret": checksum mismatch'), 'Get "[server]": checksum mismatch');
 assert.match(formatError('context deadline exceeded'), /timed out/);
 assert.equal(formatError({}), 'Something went wrong. Please try again.');
});

test('activity labels preserve names and compact byte summaries preserve values', () => {
  assert.equal(sentenceCase('aremu'), 'Aremu');
  assert.equal(sentenceCase('HDR textures'), 'HDR textures');
  assert.equal(sentenceCase(''), '');
  const item = { download_message: 'Receiving 568.53 MB/724.24 MB', extra_message: 'Data saved: 421.98 MB (74.22%)' };
  assert.equal(activityTransferSizes(item), '568.53 MB / 724.24 MB');
  assert.equal(activitySavedSize(item), '421.98 MB');
  assert.equal(activityTransferSizes({ message: 'Preparing' }), '');
  assert.equal(activitySavedSize({}), '');
});

test('transfer segments partition progress without counting saved data as downloaded', () => {
  const operation = { percentage: 60, message: 'Receiving 300 MB/500 MB', download_message: 'Receiving 300 MB/500 MB', extra_message: 'Data saved: 100 MB (33.33%)' };
  const progress = activityTransferProgress(operation);
  assert.equal(progress.savedPercentage, 20);
  const rebuilding = activityTransferProgress({ ...operation, message: 'Rebuilding asset' });
  assert.equal(rebuilding.savedPercentage, 0);
  assert.equal(rebuilding.downloadedPercentage, 60);
  assert.equal(progress.downloadedPercentage, 40);
  assert.equal(progress.downloaded, 200 * 1024 ** 2);
  assert.equal(activityTransferProgress({ ...operation, extra_message: 'Data saved: 300 MB (100.00%)' }).downloadedPercentage, 0);
  assert.equal(activityTransferProgress({ ...operation, extra_message: '' }).downloadedPercentage, 60);
  assert.deepEqual(activityTransferProgress({}), { downloading: false, saved: 0, downloaded: 0, savedPercentage: 0, downloadedPercentage: 0 });
});
