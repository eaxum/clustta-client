import { formatError } from './errors.js';

export const isActivityPending = (item) => item.status === 'running' || item.status === 'queued';

export function sentenceCase(value = '') {
  return value.replace(/^(\s*)(\S)/u, (_, space, initial) => space + initial.toLocaleUpperCase());
}

export function activityTransferSizes(operation) {
  const message = operation.download_message || operation.message || '';
  return message.match(/^Receiving (.+)$/)?.[1].replace(/\s*\/\s*/, ' / ') || '';
}

export function activitySavedSize(operation) {
  return operation.extra_message?.match(/^Data saved: (.+) \([\d.]+%\)$/)?.[1] || '';
}

const byteUnits = ['B', 'KB', 'MB', 'GB', 'TB'];
const bytesPerUnit = 1024;

function parseByteSize(value = '') {
  const match = value.trim().match(/^([\d.]+)\s*(B|KB|MB|GB|TB)$/i);
  if (!match) return 0;
  return Number(match[1]) * bytesPerUnit ** byteUnits.indexOf(match[2].toUpperCase());
}

export function activityTransferProgress(operation) {
  const received = parseByteSize(activityTransferSizes(operation).split('/')[0]);
  const saved = Math.min(received, parseByteSize(activitySavedSize(operation)));
  const percentage = Math.min(100, Math.max(0, Number(operation.percentage) || 0));
  const downloading = !!operation.download_message && operation.message === operation.download_message;
  const savedPercentage = downloading && received > 0 ? percentage * saved / received : 0;
  return {
    downloading,
    saved,
    downloaded: Math.max(0, received - saved),
    savedPercentage,
    downloadedPercentage: percentage - savedPercentage,
  };
}

export function activityDisplayName(operation) {
  if (['Fetching assets', 'Downloading checkpoint'].includes(operation.title)) {
    return sentenceCase(operation.assets.map((asset) => `${asset.name}${asset.extension || ''}`).join(', ') || operation.title);
  }
  return sentenceCase(operation.title.replace(/^(Fetching|Downloading) /, ''));
}

export function filterActivityOperations(operations, searchQuery) {
  const query = searchQuery.trim().toLowerCase();
  if (!query) return operations;
  return operations.filter((item) => [item.title, item.project_name, item.status, ...item.assets.map((asset) => `${asset.name}${asset.extension || ''}`)]
    .some((value) => value.toLowerCase().includes(query)));
}

export function sortActivityOperations(operations) {
  const latestFirst = [...operations].reverse();
  return [
    ...latestFirst.filter((item) => item.status === 'running'),
    ...operations.filter((item) => item.status === 'queued'),
    ...latestFirst.filter((item) => !isActivityPending(item)),
  ];
}

export function applyActivitySnapshot(state, snapshot, restore = false) {
  if (!snapshot || snapshot.revision <= state.revision) return [];
  const previous = new Map(state.operations.map((item) => [item.operation_id, item]));
  state.revision = snapshot.revision;
  const hadPending = state.operations.some(isActivityPending);
  state.operations = snapshot.operations.map((item) => ({ ...item, error: item.error ? formatError(item.error) : '' }));
  const pending = state.operations.filter(isActivityPending);
  if (restore || !hadPending) state.batchIDs = pending.map((item) => item.operation_id);
  for (const item of pending) {
    if (!state.batchIDs.includes(item.operation_id)) state.batchIDs.push(item.operation_id);
  }
  const finished = [];
  for (const item of state.operations) {
    const before = previous.get(item.operation_id);
    if (!restore && !isActivityPending(item) && before?.status !== item.status) finished.push(item);
  }
  return finished;
}

const normalizePath = (path = '') => path.replace(/\\/g, '/').split('/').filter(Boolean).join('/');

function includeAncestors(paths, collectionPath) {
  const parts = normalizePath(collectionPath).split('/').filter(Boolean);
  while (parts.length) {
    paths.add(parts.join('/'));
    parts.pop();
  }
}

export function buildPendingTransfers(operations) {
  const projects = {};
  for (const operation of operations) {
    if (!isActivityPending(operation)) continue;
    const pending = projects[operation.project_uri] ||= { assets: new Set(), collections: new Set(), paths: new Set() };
    const restored = new Set(operation.restored_asset_ids || []);
    for (const asset of operation.assets || []) {
      if (restored.has(asset.id)) continue;
      pending.assets.add(asset.id);
      if (asset.collection_id) pending.collections.add(asset.collection_id);
      includeAncestors(pending.paths, asset.collection_path);
    }
    for (const collection of operation.collections || []) {
      if (collection.id) pending.collections.add(collection.id);
      includeAncestors(pending.paths, collection.collection_path);
    }
  }
  return projects;
}

export function isTransferPending(projects, projectUri, item) {
  const pending = projects[projectUri];
  if (!pending || !item) return false;
  if (item.type === 'asset') return pending.assets.has(item.id);
  if (item.type !== 'collection') return false;
  return pending.collections.has(item.id) || pending.paths.has(normalizePath(item.collection_path));
}
