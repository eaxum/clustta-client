import { getBrowserItemKey } from './browserTree.js';

const LOCAL_FILE_STATES = new Set(['normal', 'modified', 'outdated']);

export function getExportDragSelection(asset, selectedItems, itemsByKey) {
  const selection = selectedItems.some((item) => item.id === asset.id) ? selectedItems : [asset];
  return selection.map((item) => itemsByKey[getBrowserItemKey(item)] || item);
}

export function canDisplayExportDrag(items) {
  return items.length > 0 && items.every((item) => item.type === 'asset'
    && !item.is_link && !item.pointer && !item.trashed);
}

export function canExportFiles(items) {
  return canDisplayExportDrag(items) && items.every((item) =>
    Boolean(item.local_path || item.file_path) && LOCAL_FILE_STATES.has(item.file_status));
}
