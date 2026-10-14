<template>
  <div
    ref="columnBrowser"
    class="column-browser"
    tabindex="0"
    @keydown="handleKeyDown"
  >
    <div class="column-browser-grid" :style="columnGridStyle">
      <ColumnPane
        v-for="column in visibleColumns"
        :key="column.key"
        :error="browserTreeStore.parentErrors[column.key]"
        :items="column.items"
        :loading="isColumnLoading(column)"
        :selectedKey="selectedKeys[column.index] || ''"
        :title="column.title"
        @contextmenu="openContextMenu"
        @open="openItem"
        @retry="retryColumn(column)"
        @select="selectItem($event, column.index)"
        @stateAction="handleAssetStateAction"
      />

      <div v-for="placeholder in placeholderCount" :key="`placeholder-${placeholder}`" class="column-placeholder"></div>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAssetLauncher } from '@/composables/useAssetLauncher';
import ColumnPane from '@/instances/desktop/components/ColumnPane.vue';
import { browserRootParentKey, getBrowserItemKey } from '@/lib/browserTree';
import { canCreateCheckpointForItem, isCheckpointBlockedByAssignment } from '@/lib/permissions';
import { CheckpointService } from '@/services';
import { useAssetStore } from '@/stores/assets';
import { useBrowserTreeStore } from '@/stores/browserTree';
import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { usePaneStore } from '@/stores/panes';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTrayStates } from '@/stores/TrayStates';
import { useUntrackedItemStore } from '@/stores/untracked';

const MINIMUM_COLUMN_COUNT = 3;
const MAXIMUM_COLUMN_COUNT = 5;
const MINIMUM_COLUMN_WIDTH = 230;

const props = defineProps({
  flatMode: { type: Boolean, default: false },
  loadChildren: { type: Function, required: true },
  rootItems: { type: Array, default: () => [] },
  rootLabel: { type: String, default: '' },
  rootLoading: { type: Boolean, default: false },
});

const assetStore = useAssetStore();
const browserTreeStore = useBrowserTreeStore();
const collectionStore = useCollectionStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const panes = usePaneStore();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const trayStates = useTrayStates();
const untrackedItemStore = useUntrackedItemStore();
const { t } = useI18n();
const { launchAsset } = useAssetLauncher();

const activeColumnIndex = ref(0);
const columnBrowser = ref(null);
const columnCapacity = ref(MINIMUM_COLUMN_COUNT);
const selectedKeys = ref([]);
const selectedPath = ref([]);
let resizeObserver = null;

const columns = computed(() => {
  const result = [{
    index: 0,
    items: props.rootItems,
    key: browserRootParentKey,
    parent: null,
    title: props.rootLabel,
  }];

  selectedPath.value.forEach((parent, pathIndex) => {
    const parentKey = getBrowserItemKey(parent);
    result.push({
      index: pathIndex + 1,
      items: browserTreeStore.getChildren(parentKey),
      key: parentKey,
      parent,
      title: parent.name || props.rootLabel,
    });
  });

  return result;
});

const visibleColumns = computed(() => {
  const startIndex = Math.max(0, columns.value.length - columnCapacity.value);
  return columns.value.slice(startIndex);
});

const placeholderCount = computed(() => {
  return Math.max(0, columnCapacity.value - visibleColumns.value.length);
});

const columnGridStyle = computed(() => ({
  gridTemplateColumns: `repeat(${columnCapacity.value}, minmax(0, 1fr))`,
}));

const isCollection = (item) => {
  const itemType = item?.type || item?.item_type || '';
  return itemType.includes('collection') || !!item?.collection_type_id;
};

const getSelectionType = (item) => {
  if (item.type === 'untracked_collection' || item.type === 'untracked_asset') return item.type;
  if (isCollection(item)) return 'collection';
  if (item.type === 'asset' || item.asset_type_id) return item.is_resource ? 'resource' : 'asset';
  return item.type || item.item_type;
};

const focusItem = (item) => {
  const selectionType = getSelectionType(item);
  stage.markedItems = [item.id];
  stage.markedAssets = selectionType === 'asset' ? [item.id] : [];
  stage.markedCollections = selectionType === 'collection' ? [item.id] : [];
  stage.markedResources = selectionType === 'resource' ? [item.id] : [];
  stage.selectedItems = [item];
  stage.firstSelectedItemId = item.id;
  stage.lastSelectedItemId = '';
  stage.selectItem(item, selectionType, true);

  if (selectionType === 'collection') {
    panes.setPaneVisibility('collectionDetails', true);
  } else if (selectionType === 'asset' || selectionType === 'resource') {
    panes.setPaneVisibility('assetDetails', true);
  } else {
    panes.setPaneVisibility('untrackedItemDetails', true);
  }
};

const showCheckpointWarning = (messageKey) => {
  trayStates.popUpModalTitle = t('common.warning');
  trayStates.popUpModalMessage = t(messageKey);
  trayStates.popUpModalIcon = 'help';
  trayStates.popUpModalFunction = null;
  modals.setModalVisibility('popUpModal', true);
};

const openCheckpointModal = (asset) => {
  if (isCheckpointBlockedByAssignment(asset)) {
    showCheckpointWarning('notifications.cannotModifyAsset');
    return;
  }
  if (!canCreateCheckpointForItem(asset)) {
    showCheckpointWarning('notifications.cannotCreateCheckpoint');
    return;
  }

  assetStore.selectAsset(asset);
  modals.setModalVisibility('createCheckpointModal', true);
};

const downloadAsset = async (asset) => {
  try {
    const { CheckpointService: WebCheckpointService } = await import('@/services/adapters/checkpointservice.js');
    await WebCheckpointService.DownloadAsset(projectStore.activeProject.uri, asset.id, null);
    notificationStore.addNotification(
      t('notifications.downloadComplete'),
      t('notifications.downloadedSuccessfully', { fileName: `${asset.name}${asset.extension}` }),
      'success',
      true
    );
  } catch (error) {
    notificationStore.errorNotification(t('notifications.downloadFailed'), error.message || error);
  }
};

const restoreAsset = async (asset) => {
  try {
    const response = await CheckpointService.Revert(
      projectStore.activeProject.uri,
      projectStore.getActiveProjectUrl,
      [asset.id]
    );
    browserTreeStore.markAssetsAvailable(response?.restored_asset_ids);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorRevertingAsset'), error);
  }
};

const handleAssetStateAction = async (asset) => {
  focusItem(asset);
  if (platformStore.isWeb && asset.type === 'asset') {
    await downloadAsset(asset);
    return;
  }
  if (asset.type === 'untracked_asset' || asset.file_status === 'modified') {
    openCheckpointModal(asset);
    return;
  }
  if (asset.file_status === 'outdated' || asset.file_status === 'fetchable') {
    await restoreAsset(asset);
  }
};

const loadCollection = async (collection, force = false) => {
  const parentKey = getBrowserItemKey(collection);
  if (!force && browserTreeStore.loadedParents[parentKey]) return;
  await props.loadChildren(collection, { force });
};

const syncExpandedCollectionPath = () => {
  const expandedCollections = {};
  for (const collection of selectedPath.value) {
    expandedCollections[collection.id] = {
      height: 0,
      collection_path: collection.collection_path || collection.item_path || '',
    };
  }
  stage.expandedCollections = expandedCollections;
};

const restoreNavigationPath = async () => {
  if (props.flatMode) return;

  if (!collectionStore.navigatedCollection) {
    const selectedItemKey = getBrowserItemKey(stage.selectedItem);
    if (selectedItemKey && props.rootItems.some((item) => getBrowserItemKey(item) === selectedItemKey)) {
      selectedKeys.value = [selectedItemKey];
      activeColumnIndex.value = 0;
    }
    return;
  }

  const navigatedCollection = collectionStore.navigatedCollection;
  const targetPath = navigatedCollection.collection_path || navigatedCollection.item_path || '';
  const normalizedTargetPath = targetPath.replace(/\\/g, '/');
  const availableCollections = Object.values(browserTreeStore.itemsByKey)
    .filter((item) => {
      if (!isCollection(item)) return false;
      const itemPath = (item.collection_path || item.item_path || '').replace(/\\/g, '/');
      return itemPath && normalizedTargetPath.startsWith(itemPath);
    });

  if (!availableCollections.some((item) => getBrowserItemKey(item) === getBrowserItemKey(navigatedCollection))) {
    availableCollections.push(navigatedCollection);
  }

  availableCollections.sort((left, right) => {
    const leftPath = left.collection_path || left.item_path || '';
    const rightPath = right.collection_path || right.item_path || '';
    return leftPath.length - rightPath.length;
  });

  const rootKeys = new Set(props.rootItems.map(getBrowserItemKey));
  const rootIndex = availableCollections.findIndex((item) => rootKeys.has(getBrowserItemKey(item)));
  if (rootIndex < 0) return;

  selectedPath.value = availableCollections.slice(rootIndex);
  selectedKeys.value = selectedPath.value.map(getBrowserItemKey);

  const selectedItem = stage.selectedItem;
  const selectedItemKey = getBrowserItemKey(selectedItem);
  if (selectedItemKey && !isCollection(selectedItem)) {
    selectedKeys.value[selectedPath.value.length] = selectedItemKey;
    activeColumnIndex.value = selectedPath.value.length;
  } else {
    activeColumnIndex.value = Math.max(0, selectedPath.value.length - 1);
  }

  syncExpandedCollectionPath();
  for (const collection of selectedPath.value) await loadCollection(collection);
};

const selectItem = async (item, columnIndex) => {
  activeColumnIndex.value = columnIndex;
  selectedKeys.value = selectedKeys.value.slice(0, columnIndex + 1);
  selectedKeys.value[columnIndex] = getBrowserItemKey(item);
  focusItem(item);

  selectedPath.value = selectedPath.value.slice(0, columnIndex);
  if (props.flatMode) return;

  if (!isCollection(item)) {
    collectionStore.navigatedCollection = selectedPath.value.at(-1) || null;
    syncExpandedCollectionPath();
    return;
  }

  selectedPath.value.push(item);
  collectionStore.navigateToCollection(item);
  syncExpandedCollectionPath();
  await loadCollection(item);
};

const openItem = async (item) => {
  if (isCollection(item)) return;
  await launchAsset(item);
};

const openContextMenu = (event, item) => {
  focusItem(item);
  const itemType = item.type || '';

  if (itemType === 'collection') {
    stage.markedCollections = [item.id];
    collectionStore.selectCollection(item);
    menu.showContextMenu(event, 'collectionMenu', true);
    return;
  }

  if (itemType === 'asset') {
    stage.markedAssets = [item.id];
    assetStore.selectAsset(item);
    menu.showContextMenu(event, 'assetMenu', true);
    return;
  }

  untrackedItemStore.selectUntrackedItem(item);
  menu.showContextMenu(event, 'untrackedItemMenu', true);
};

const retryColumn = async (column) => {
  if (!column.parent) return;
  await loadCollection(column.parent, true);
};

const getColumnItemIndex = (column, selectedKey) => {
  return column.items.findIndex((item) => getBrowserItemKey(item) === selectedKey);
};

const selectAdjacentItem = async (direction) => {
  const column = columns.value[activeColumnIndex.value];
  if (!column?.items.length) return;
  const selectedKey = selectedKeys.value[activeColumnIndex.value] || '';
  const currentIndex = getColumnItemIndex(column, selectedKey);
  const nextIndex = Math.min(
    column.items.length - 1,
    Math.max(0, currentIndex < 0 ? 0 : currentIndex + direction)
  );
  await selectItem(column.items[nextIndex], activeColumnIndex.value);
};

const moveToNextColumn = async () => {
  const selectedKey = selectedKeys.value[activeColumnIndex.value];
  const column = columns.value[activeColumnIndex.value];
  const selectedItem = column?.items.find((item) => getBrowserItemKey(item) === selectedKey);
  if (!selectedItem || !isCollection(selectedItem) || props.flatMode) return;

  await loadCollection(selectedItem);
  activeColumnIndex.value = Math.min(activeColumnIndex.value + 1, columns.value.length - 1);
  const nextColumn = columns.value[activeColumnIndex.value];
  if (nextColumn?.items.length && !selectedKeys.value[activeColumnIndex.value]) {
    await selectItem(nextColumn.items[0], activeColumnIndex.value);
  }
};

const moveToPreviousColumn = () => {
  if (activeColumnIndex.value === 0) return;
  activeColumnIndex.value -= 1;
  selectedPath.value = selectedPath.value.slice(0, activeColumnIndex.value);
  selectedKeys.value = selectedKeys.value.slice(0, activeColumnIndex.value + 1);

  const column = columns.value[activeColumnIndex.value];
  const selectedItem = column?.items.find((item) => (
    getBrowserItemKey(item) === selectedKeys.value[activeColumnIndex.value]
  ));
  if (!selectedItem) return;
  focusItem(selectedItem);
  collectionStore.navigateToCollection(selectedItem);
  syncExpandedCollectionPath();
};

const handleKeyDown = async (event) => {
  if (event.key === 'ArrowDown') {
    event.preventDefault();
    await selectAdjacentItem(1);
  } else if (event.key === 'ArrowUp') {
    event.preventDefault();
    await selectAdjacentItem(-1);
  } else if (event.key === 'ArrowRight') {
    event.preventDefault();
    await moveToNextColumn();
  } else if (event.key === 'ArrowLeft') {
    event.preventDefault();
    moveToPreviousColumn();
  } else if (event.key === 'Enter') {
    const column = columns.value[activeColumnIndex.value];
    const selectedItem = column?.items.find((item) => (
      getBrowserItemKey(item) === selectedKeys.value[activeColumnIndex.value]
    ));
    if (selectedItem) openItem(selectedItem);
  }
};

const isColumnLoading = (column) => {
  if (column.key === browserRootParentKey) return props.rootLoading;
  return !!browserTreeStore.loadingParents[column.key];
};

const updateColumnCapacity = () => {
  const width = columnBrowser.value?.getBoundingClientRect().width || 0;
  if (!width) return;
  columnCapacity.value = Math.min(
    MAXIMUM_COLUMN_COUNT,
    Math.max(MINIMUM_COLUMN_COUNT, Math.floor(width / MINIMUM_COLUMN_WIDTH))
  );
};

const refreshOpenColumns = async () => {
  for (const collection of selectedPath.value) {
    await loadCollection(collection, true);
  }
};

watch(() => props.flatMode, (flatMode) => {
  if (!flatMode) return;
  selectedPath.value = [];
  selectedKeys.value = [];
  activeColumnIndex.value = 0;
});

watch(() => getBrowserItemKey(collectionStore.navigatedCollection), async () => {
  if (props.flatMode) return;

  const collection = collectionStore.navigatedCollection;
  if (!collection) {
    selectedPath.value = [];
    selectedKeys.value = [];
    activeColumnIndex.value = 0;
    syncExpandedCollectionPath();
    return;
  }

  const collectionKey = getBrowserItemKey(collection);
  const pathIndex = selectedPath.value.findIndex((item) => getBrowserItemKey(item) === collectionKey);
  if (pathIndex < 0) {
    await restoreNavigationPath();
    return;
  }
  if (pathIndex === selectedPath.value.length - 1) return;

  selectedPath.value = selectedPath.value.slice(0, pathIndex + 1);
  selectedKeys.value = selectedKeys.value.slice(0, pathIndex + 1);
  activeColumnIndex.value = pathIndex;
  focusItem(collection);
  syncExpandedCollectionPath();
});

watch(() => props.rootItems.map(getBrowserItemKey).join('|'), async () => {
  if (!selectedPath.value.length) {
    await restoreNavigationPath();
    return;
  }
  const firstKey = getBrowserItemKey(selectedPath.value[0]);
  if (props.rootItems.some((item) => getBrowserItemKey(item) === firstKey)) return;
  selectedPath.value = [];
  selectedKeys.value = [];
  activeColumnIndex.value = 0;
  await restoreNavigationPath();
});

onMounted(async () => {
  resizeObserver = new ResizeObserver(updateColumnCapacity);
  resizeObserver.observe(columnBrowser.value);
  updateColumnCapacity();
  await restoreNavigationPath();
  nextTick(() => columnBrowser.value?.focus());
});

onBeforeUnmount(() => {
  resizeObserver?.disconnect();
});

defineExpose({ refreshOpenColumns });
</script>

<style scoped>
.column-browser {
  display: flex;
  width: 100%;
  height: 100%;
  min-width: 0;
  flex-direction: column;
  overflow: hidden;
  outline: none;
}

.column-browser-grid {
  display: grid;
  min-height: 0;
  flex: 1;
  gap: .8rem;
  box-sizing: border-box;
  padding: 1px;
}

.column-browser:focus-visible {
  outline: 1px solid var(--accent);
  outline-offset: -1px;
}

.column-placeholder {
  min-width: 0;
  box-sizing: border-box;
  background-color: color-mix(in oklch, var(--surface-1) 82%, transparent);
  border-radius: var(--very-large-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
}
</style>
