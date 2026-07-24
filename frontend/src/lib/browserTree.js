export const browserRootParentKey = 'browser:root';

export const getBrowserItemKey = (item) => {
  if (!item?.id) return '';

  const itemType = item.type
    || item.item_type
    || (item.asset_type_id ? 'asset' : '')
    || (item.collection_type_id ? 'collection' : '');
  if (!itemType) return '';

  return `${itemType}:${item.id}`;
};

export const reconcileBrowserItems = (currentItems = [], incomingItems = []) => {
  const currentItemsByKey = new Map();

  for (const item of currentItems) {
    const itemKey = getBrowserItemKey(item);
    if (itemKey) currentItemsByKey.set(itemKey, item);
  }

  const reconciledItems = [];
  const includedKeys = new Set();

  for (const incomingItem of incomingItems) {
    const itemKey = getBrowserItemKey(incomingItem);
    if (!itemKey || includedKeys.has(itemKey)) continue;

    const currentItem = currentItemsByKey.get(itemKey);
    if (currentItem && currentItem !== incomingItem) {
      Object.assign(currentItem, incomingItem);
    }

    reconciledItems.push(currentItem || incomingItem);
    includedKeys.add(itemKey);
  }

  return reconciledItems;
};
