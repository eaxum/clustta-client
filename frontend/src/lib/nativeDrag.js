const LOCAL_FILE_STATES = new Set(['normal', 'modified', 'outdated']);

export function dragSelection(asset, selectedItems) {
  return selectedItems.some((item) => item.id === asset.id) ? selectedItems : [asset];
}

export function canDragFiles(items) {
  return items.length > 0 && items.every((item) => item.type === 'asset'
    && !item.is_link && !item.pointer && !item.trashed
    && Boolean(item.local_path || item.file_path) && LOCAL_FILE_STATES.has(item.file_status));
}
