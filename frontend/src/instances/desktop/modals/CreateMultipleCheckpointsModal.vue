<template>
  <div class="modal-container" v-stop-propagation>
    <HeaderArea :title="$t('modals.createCheckpoints')" :icon="getAppIcon('plus-stone')" />
    
    <div class="general-container" :class="{ 'with-modified-items-history': showModifiedItemsSidePane }">
      <div class="checkpoint-create-layout">
        <div class="checkpoint-create-form">
          <textarea v-model="message" class="desktop-input-long" type="text" :placeholder="$t('placeholders.writeAComment')" v-focus
            @keydown.enter="handleEnterKey" />

          <div class="checkpoint-create-controls">
            <InputAlert :show="!isValueChanged" :message="validationMessage" />

            <div v-if="assetStore.loadingAssetStates" class="horizontal-flex input-alert loading-items-count">
              <ActionButton :isLoading="true" :icon="getAppIcon('loading')"
                v-tooltip="$t('modals.loadingCollectionStates')" />

              <div class="refresh-label">
                {{ $t('modals.refreshingModifiedItems') }}
              </div>
            </div>

            <div v-else class="horizontal-flex input-alert modified-items-count"
              :class="{ 'modified-items-count-expanded' : showCheckpointItems && !showModifiedItemsSidePane}"
              @click="toggleShowCheckpointItems()">

              {{ $t('modals.itemsModified', { count: totalCheckpointItems }) }}

              <ActionButton :isInactive="true" :label="showCheckpointItems ? $t('common.hide') : $t('common.show')"
                :icon="getAppIcon(showCheckpointItems ? 'eye-cancel' : 'eye')" />
            </div>

            <div v-if="showUntrackedToggle" class="checkpoint-untracked-toggle" @click="toggleIncludeUntracked">
              <div class="checkpoint-untracked-toggle-label">{{ $t('menus.untracked') }}</div>
              <ToggleSwitch :switchValueProp="includeUntrackedInCheckpoint" />
            </div>

            <div v-if="showCheckpointItems && !showModifiedItemsSidePane" class="modified-items-inline">
              <div class="modified-items-tabs-header">
                <div class="modified-items-tabs" :class="{ 'modified-items-tabs-empty': !showCheckpointTabs }">
                  <PaneHeaderTabs v-if="showCheckpointTabs" :dataTypes="modifiedItemTabs" :selectedTab="selectedModifiedItemsFilter" @filter="handleModifiedItemsFilterChange" />
                </div>

                <div class="modified-items-tabs-options">
                  <ActionButton :icon="hideExtensions ? getAppIcon('extension-cancel') : getAppIcon('extension')" v-tooltip="hideExtensions ? $t('modals.showExtensions') : $t('modals.hideExtensions')" :buttonFunction="toggleHideExtensions" />
                  <ActionButton :icon="showFullPath ? getAppIcon('file-name') : getAppIcon('file-path')" v-tooltip="showFullPath ? $t('modals.nameColumn') : $t('modals.pathColumn')" :buttonFunction="toggleShowFullPath" />
                </div>
              </div>

              <div class="modified-items">
                <PageState v-if="!filteredCheckpointItems.length" class="modified-items-empty-state" :message="modifiedItemsEmptyMessage" :illustration="'/page-states/resources.png'" />
                <template v-for="item in filteredCheckpointItems" :key="item.key">
                  <AssetItem
                    :item="item"
                    :hideExtension="hideExtensions"
                    :showFullPath="showFullPath"
                    :showNavigate="true"
                    :navigateTooltip="$t('components.changeItem.goToItem')"
                    @navigate="goToItem(item)"
                  >
                    <template #actions>
                      <ActionButton :icon="getAppIcon('close')" v-tooltip="$t('common.remove')" :buttonFunction="() => removeItem(item.key)" />
                    </template>
                  </AssetItem>
                </template>
              </div>
            </div>
          </div>
        </div>

        <div v-if="showModifiedItemsSidePane" class="modified-items-history">
          <div class="modified-items-tabs-header">
            <div class="modified-items-tabs" :class="{ 'modified-items-tabs-empty': !showCheckpointTabs }">
              <PaneHeaderTabs v-if="showCheckpointTabs" :dataTypes="modifiedItemTabs" :selectedTab="selectedModifiedItemsFilter" @filter="handleModifiedItemsFilterChange" />
            </div>

            <div class="modified-items-tabs-options">
              <ActionButton :icon="hideExtensions ? getAppIcon('extension-cancel') : getAppIcon('extension')" v-tooltip="hideExtensions ? $t('modals.showExtensions') : $t('modals.hideExtensions')" :buttonFunction="toggleHideExtensions" />
              <ActionButton :icon="showFullPath ? getAppIcon('file-name') : getAppIcon('file-path')" v-tooltip="showFullPath ? $t('modals.nameColumn') : $t('modals.pathColumn')" :buttonFunction="toggleShowFullPath" />
            </div>
          </div>

          <div class="modified-items">
            <PageState v-if="!filteredCheckpointItems.length" class="modified-items-empty-state" :message="modifiedItemsEmptyMessage" :illustration="'/page-states/resources.png'" />
            <template v-for="item in filteredCheckpointItems" :key="item.key">
              <AssetItem
                :item="item"
                :hideExtension="hideExtensions"
                :showFullPath="showFullPath"
                :showNavigate="true"
                :navigateTooltip="$t('components.changeItem.goToItem')"
                @navigate="goToItem(item)"
              >
                <template #actions>
                  <ActionButton :icon="getAppIcon('close')" v-tooltip="$t('common.remove')" :buttonFunction="() => removeItem(item.key)" />
                </template>
              </AssetItem>
            </template>
          </div>
        </div>
      </div>

    <div class="pop-up-actions">
      <GeneralButton :label="$t('common.close')" :fullWidth="true" :buttonFunction="closeModal" :isActive="!isAwaitingResponse" :colored="false" />
      <GeneralButton :label="$t('common.confirm')" :fullWidth="true" @click="createCheckPoints" :isActive="isValueChanged"
        :loading="isAwaitingResponse" />
    </div>
    
    </div>


  </div>
</template>

<script setup>
// imports
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { v4 as uuidv4 } from 'uuid';
import emitter from '@/lib/mitt';
import { canCreateCheckpointForItem } from '@/lib/permissions';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import AssetItem from '@/instances/desktop/components/AssetItem.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';
import PageState from '@/instances/common/components/PageState.vue';
import PaneHeaderTabs from '@/instances/common/components/PaneHeaderTabs.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// services
import { AssetService, CheckpointService, CollectionService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useDndStore } from '@/stores/dnd';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTrayStates } from '@/stores/TrayStates';

const { t } = useI18n();
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const dndStore = useDndStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const trayStates = useTrayStates();

// refs
const isAwaitingResponse = ref(false);
const hideExtensions = ref(true);
const includeUntrackedInCheckpoint = ref(commonStore.showUntracked);
const message = ref('');
const removedPaths = ref([]);
const selectedModifiedItemsFilter = ref('all');
const showCheckpointItems = ref(false);
const showFullPath = ref(false);
const useImageAsCover = ref(true);

// constants
const forbiddenComments = ['wip', 'wfa', 'retake', 'retook', 'todo', 'fmf'];
const modifiedItemTabs = [
  { name: 'all', icon: 'plus-stone' },
  { name: 'modified', icon: 'plus-stone', iconClass: 'modified-items-tab-alert' },
  { name: 'new', icon: 'plus-stone', iconClass: 'modified-items-tab-danger' },
];

const getModifiedAssetKey = (assetState) => `${assetState.asset_path}${assetState.extension || ''}`;

// computed
// Returns modified asset display paths after filtering.
const currentModifiedDisplayPaths = computed(() => {
  let filteredAssets = assetStore.modifiedAssets.modified || [];
  filteredAssets = filteredAssets.filter((assetState) => !removedPaths.value.includes(getModifiedAssetKey(assetState)));
  filteredAssets = filteredAssets.filter((assetState) => canCreateCheckpointForItem(assetState));
  if (trayStates.createMultipleCheckpointsCollectionPath) {
    filteredAssets = filteredAssets.filter((assetState) => assetState.asset_path.startsWith(trayStates.createMultipleCheckpointsCollectionPath));
  }
  if (!trayStates.createMultipleCheckpoints) {
    filteredAssets = filteredAssets.filter((assetState) => stage.markedItems.includes(assetState.asset_id));
  }
  return filteredAssets;
});

// Returns untracked file paths after filtering.
const availableUntrackedPaths = computed(() => {
  let filteredAssets = assetStore.modifiedAssets.untracked || [];
  filteredAssets = filteredAssets.filter((untrackedAsset) => !removedPaths.value.includes(getUntrackedCandidatePath(untrackedAsset)));
  filteredAssets = filteredAssets.filter((untrackedAsset) => {
    const untrackedAssetPath = getUntrackedCandidatePath(untrackedAsset);
    const resolvedItem = resolveUntrackedItemByPath(untrackedAssetPath) || (typeof untrackedAsset === 'object' ? untrackedAsset : null) || {
      type: 'untracked_asset',
      asset_path: normalizePath(untrackedAssetPath),
      item_path: normalizePath(untrackedAssetPath),
      collection_path: getParentPath(untrackedAssetPath),
    };
    return canCreateCheckpointForItem(resolvedItem);
  });
  if (trayStates.createMultipleCheckpointsCollectionPath) {
    filteredAssets = filteredAssets.filter((untrackedAsset) => getUntrackedCandidatePath(untrackedAsset).startsWith(trayStates.createMultipleCheckpointsCollectionPath));
  }
  if (trayStates.createMultipleCheckpoints) {
    return filteredAssets;
  } else {
    const selectedUntrackedAssets = stage.selectedItems
      .filter(item => item.type === 'untracked_asset' || item.type === 'untracked_collection')
      .map(item => item.asset_path || item.item_path)
      .filter(path => path && filteredAssets.some((untrackedAsset) => getUntrackedCandidatePath(untrackedAsset) === path));
    return selectedUntrackedAssets;
  }
});

// Returns untracked file paths when the modal-local untracked toggle is enabled.
const currentUntrackedPaths = computed(() => {
  if (!includeUntrackedInCheckpoint.value) return [];
  return availableUntrackedPaths.value;
});

// Returns a normalized list of modified and untracked checkpoint candidates.
const checkpointItems = computed(() => {
  const modifiedItems = currentModifiedDisplayPaths.value.map((assetState) => {
    const key = getModifiedAssetKey(assetState);
    const fullPath = normalizePath(assetState.display_path || `${assetState.asset_path}${assetState.extension || ''}`);
    const extension = assetState.extension || getExtension(fullPath);
    const name = assetState.name || stripExtension(getFileName(fullPath), extension);

    return {
      key,
      id: assetState.asset_id,
      kind: 'modified',
      kindLabel: 'Modified',
      name,
      extension,
      fullPath,
      fallbackIcon: assetState.icon || '',
      source: assetState,
    };
  });

  const untrackedItems = currentUntrackedPaths.value.map((untrackedAsset) => {
    const assetPath = getUntrackedCandidatePath(untrackedAsset);
    const fullPath = normalizePath(assetPath);
    const extension = getExtension(fullPath);
    const name = stripExtension(getFileName(fullPath), extension);
    const resolvedItem = resolveUntrackedItemByPath(fullPath) || (typeof untrackedAsset === 'object' ? untrackedAsset : null);

    return {
      key: assetPath,
      id: resolvedItem?.id || assetPath,
      kind: 'untracked',
      kindLabel: 'New',
      name: resolvedItem?.name || name,
      extension: resolvedItem?.extension || extension,
      fullPath: getUntrackedDisplayPath(resolvedItem) || fullPath,
      fallbackIcon: resolvedItem?.icon || '',
      source: resolvedItem || {
        id: assetPath,
        type: 'untracked_asset',
        asset_path: fullPath,
        file_path: getAbsoluteProjectPath(fullPath),
        name,
        extension,
      },
    };
  });

  return [...modifiedItems, ...untrackedItems];
});

// Returns checkpoint candidates matching the selected tab.
const filteredCheckpointItems = computed(() => {
  if (!showCheckpointTabs.value) return checkpointItems.value;
  if (selectedModifiedItemsFilter.value === 'modified') {
    return checkpointItems.value.filter((item) => item.kind === 'modified');
  }
  if (selectedModifiedItemsFilter.value === 'new') {
    return checkpointItems.value.filter((item) => item.kind === 'untracked');
  }
  return checkpointItems.value;
});

// Returns the empty-state message for the selected modified-items tab.
const modifiedItemsEmptyMessage = computed(() => {
  if (selectedModifiedItemsFilter.value === 'modified') return 'No modified items';
  if (selectedModifiedItemsFilter.value === 'new') return 'No new items';
  return 'No items';
});

// Returns whether the modified items should render as a side pane.
const showModifiedItemsSidePane = computed(() => {
  return showCheckpointItems.value && totalCheckpointItems.value > 2;
});

// Returns whether both tracked and untracked candidates are currently visible.
const hasMixedCheckpointItems = computed(() => currentModifiedDisplayPaths.value.length > 0 && currentUntrackedPaths.value.length > 0);

// Returns whether the untracked toggle is useful for the current candidate set.
const showUntrackedToggle = computed(() => {
  return trayStates.createMultipleCheckpoints
    && currentModifiedDisplayPaths.value.length > 0
    && availableUntrackedPaths.value.length > 0;
});

// Returns whether tab filters are useful for the current candidate set.
const showCheckpointTabs = computed(() => hasMixedCheckpointItems.value);

// Returns total checkpoint candidate count.
const totalCheckpointItems = computed(() => {
  return currentModifiedDisplayPaths.value.length + currentUntrackedPaths.value.length;
});

// Returns whether the message is valid for submission.
const isValueChanged = computed(() => {
  const messageWords = message.value.toLowerCase().split(/\s+/);
  const hasForbiddenWord = forbiddenComments.some(comment =>
    messageWords.includes(comment.toLowerCase())
  );
  return !assetStore.loadingAssetStates && message.value.trim().length > 6 && !hasForbiddenWord;
});

// Returns the validation message for the comment field.
const validationMessage = computed(() => {
  if (message.value.trim().length <= 6) {
    return t('notifications.messageTooShort');
  }
  const messageWords = message.value.toLowerCase().split(/\s+/);
  const foundForbidden = forbiddenComments.find(comment =>
    messageWords.includes(comment.toLowerCase())
  );
  if (foundForbidden) {
    return t('notifications.avoidForbiddenWord', { word: foundForbidden.toUpperCase() });
  }
  return '';
});

// methods
// Closes the modal.
const closeModal = () => {
  trayStates.createMultipleCheckpoints = true;
  modals.disableAllModals();
};

// Creates checkpoints for all modified items.
const createCheckPoints = async () => {
  const startTime = performance.now();
  isAwaitingResponse.value = true;
  const comment = message.value;
  const previewPath = '';
  const groupId = uuidv4();
  const assetPathsForCheckpoints = currentModifiedDisplayPaths.value.map(assetState => assetState.asset_path);
  const extensionsForCheckpoints = currentModifiedDisplayPaths.value.map(assetState => assetState.extension);
  const modifiedAssetKeysForCheckpoints = currentModifiedDisplayPaths.value.map(getModifiedAssetKey);
  await CheckpointService.AddCheckpoint(projectStore.activeProject.uri, assetPathsForCheckpoints, extensionsForCheckpoints, comment, previewPath, groupId, useImageAsCover.value, false)
    .then(() => {
      assetStore.modifiedAssets.modified = assetStore.modifiedAssets.modified.filter(
        (item) => !modifiedAssetKeysForCheckpoints.includes(getModifiedAssetKey(item))
      );
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification(t('notifications.failedToCreateCheckpoints'), error);
      isAwaitingResponse.value = false;
    });
  const untracked = currentUntrackedPaths.value;
  try {
    for (let i = 0; i < untracked.length; i += 100) {
      const batch = untracked.slice(i, i + 100).map(getUntrackedCandidatePath);
      await CheckpointService.AddUntrackedAsset(projectStore.activeProject.uri, projectStore.activeProject.working_directory, batch, i, untracked.length, comment, previewPath, groupId);
    }
  } catch (error) {
    isAwaitingResponse.value = false;
    notificationStore.errorNotification(t('notifications.errorCreatingCheckpoint'), error);
  }
  assetStore.modifiedAssets.untracked = assetStore.modifiedAssets.untracked.filter(
    (untrackedAsset) => !currentUntrackedPaths.value.some((checkpointAsset) => getUntrackedCandidatePath(checkpointAsset) === getUntrackedCandidatePath(untrackedAsset))
  );
  emitter.emit('refresh-browser');
  isAwaitingResponse.value = false;
  modals.disableAllModals();
  const endTime = performance.now();
  const executionTime = endTime - startTime;
  const minutes = Math.floor(executionTime / 60000);
  const seconds = Math.floor((executionTime % 60000) / 1000);
  console.log(`createCheckPoints completed in: ${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`);
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Returns a normalized checkpoint path from either a path string or an untracked item.
const getUntrackedCandidatePath = (item) => {
  if (typeof item === 'string') return normalizePath(item);
  return normalizePath(item?.asset_path || item?.item_path || item?.display_path || '');
};

// Returns the file name portion from a normalized path.
const getFileName = (filePath = '') => {
  return normalizePath(filePath).split('/').filter(Boolean).pop() || filePath;
};

// Returns the extension from a path, including the leading dot.
const getExtension = (filePath = '') => {
  const fileName = getFileName(filePath);
  const dotIndex = fileName.lastIndexOf('.');
  if (dotIndex <= 0) return '';
  return fileName.slice(dotIndex);
};

// Returns the parent path for a normalized item path.
const getParentPath = (filePath = '') => {
  const parts = normalizePath(filePath).split('/').filter(Boolean);
  parts.pop();
  return parts.join('/');
};

// Returns an untracked item path suitable for display and path matching.
const getUntrackedDisplayPath = (item) => {
  if (!item) return '';
  return normalizePath(
    item.item_path
    || item.asset_path
    || getRelativeProjectPath(item.file_path)
    || ''
  );
};

// Returns a project-relative path for an absolute filesystem path.
const getRelativeProjectPath = (filePath = '') => {
  const normalizedPath = normalizePath(filePath);
  const workingDirectory = normalizePath(projectStore.activeProject?.working_directory || '');
  if (!normalizedPath || !workingDirectory) return normalizedPath;
  if (!normalizedPath.toLowerCase().startsWith(workingDirectory.toLowerCase())) return normalizedPath;
  return normalizedPath.slice(workingDirectory.length).replace(/^\/+/, '');
};

// Returns an absolute filesystem path inside the active project.
const getAbsoluteProjectPath = (filePath = '') => {
  const normalizedPath = normalizePath(filePath);
  const workingDirectory = normalizePath(projectStore.activeProject?.working_directory || '').replace(/\/+$/, '');
  if (!normalizedPath || !workingDirectory) return normalizedPath;
  if (normalizedPath.toLowerCase().startsWith(workingDirectory.toLowerCase())) return normalizedPath;
  return `${workingDirectory}/${normalizedPath.replace(/^\/+/, '')}`;
};

// Returns all normalized path variants that can identify an untracked item.
const getUntrackedPathCandidates = (item = {}) => {
  const paths = [
    item.item_path,
    item.asset_path,
    item.file_path,
    getRelativeProjectPath(item.file_path),
  ].filter(Boolean).map(normalizePath);

  if (item.item_path && item.extension && !normalizePath(item.item_path).toLowerCase().endsWith(item.extension.toLowerCase())) {
    paths.push(normalizePath(`${item.item_path}${item.extension}`));
  }

  return [...new Set(paths)];
};

// Resolves an untracked candidate path to the real browser/project item when possible.
const resolveUntrackedItemByPath = (filePath = '') => {
  const normalizedPath = normalizePath(filePath);
  const normalizedPathWithoutExtension = stripExtension(normalizedPath, getExtension(normalizedPath));
  const matchesPath = (candidate) => {
    return getUntrackedPathCandidates(candidate).some((candidatePath) => {
      const relativeCandidatePath = getRelativeProjectPath(candidatePath);
      const candidatePathWithoutExtension = stripExtension(candidatePath, getExtension(candidatePath));
      const relativeCandidatePathWithoutExtension = stripExtension(relativeCandidatePath, getExtension(relativeCandidatePath));
      return candidatePath === normalizedPath
        || relativeCandidatePath === normalizedPath
        || candidatePathWithoutExtension === normalizedPathWithoutExtension
        || relativeCandidatePathWithoutExtension === normalizedPathWithoutExtension;
    });
  };

  return stage.selectedItems.find(matchesPath)
    || dndStore.allViewItems.find((candidate) => candidate.type === 'untracked_asset' && matchesPath(candidate))
    || projectStore.untrackedFiles.find(matchesPath)
    || projectStore.findUntrackedAsset(filePath);
};

// Returns the tracked or untracked parent collection for an untracked item.
const resolveUntrackedParent = (item) => {
  const collectionPath = normalizePath(item.collection_path || getParentPath(getUntrackedDisplayPath(item)));
  if (!collectionPath) return null;

  return collectionStore.collections.find((collection) => normalizePath(collection.collection_path) === collectionPath)
    || projectStore.untrackedFolders.find((collection) => normalizePath(collection.item_path) === collectionPath)
    || {
      id: collectionPath,
      name: getFileName(collectionPath),
      item_path: collectionPath,
      collection_path: getParentPath(collectionPath),
      file_path: getAbsoluteProjectPath(collectionPath),
      item_type: 'folder',
      type: 'untracked_collection',
      collection_type_icon: 'folder',
    };
};

// Waits briefly for the browser to render the target element after navigation.
const waitForItemElement = async (itemId) => {
  for (let i = 0; i < 10; i++) {
    await new Promise((resolve) => setTimeout(resolve, 50));
    await nextTick();
    const el = dndStore.visibleItemRefs[itemId] || dndStore.itemRefs[itemId];
    if (el) return el;
  }
  return null;
};

// Waits for a path-only untracked candidate to resolve to a rendered browser item.
const waitForUntrackedItem = async (filePath, fallbackItem) => {
  for (let i = 0; i < 10; i++) {
    await new Promise((resolve) => setTimeout(resolve, 50));
    await nextTick();
    const resolvedItem = resolveUntrackedItemByPath(filePath);
    if (resolvedItem?.id && (dndStore.visibleItemRefs[resolvedItem.id] || dndStore.itemRefs[resolvedItem.id])) {
      return resolvedItem;
    }
  }
  return fallbackItem;
};

// Navigates to a checkpoint candidate in the browser.
const goToItem = async (item) => {
  try {
    commonStore.navigatorMode = true;
    if (item.kind === 'modified') {
      const asset = await AssetService.GetAssetByID(projectStore.activeProject.uri, item.id);
      if (!asset?.id) return;
      if (asset.collection_id) {
        const parent = await CollectionService.GetCollectionByID(projectStore.activeProject.uri, asset.collection_id);
        if (parent) {
          collectionStore.navigateToCollection(parent);
        }
      } else {
        collectionStore.navigatedCollection = null;
      }
      stage.deselectAllItems();
      assetStore.selectAsset(asset);
      stage.selectedItem = asset;
      stage.selectedItems = [asset];
      stage.firstSelectedItemId = asset.id;
      stage.markedItems = [asset.id];
      emitter.emit('refresh-browser');
      const el = await waitForItemElement(asset.id);
      if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' });
      modals.disableAllModals();
      return;
    }

    let untrackedItem = resolveUntrackedItemByPath(item.fullPath) || item.source;
    const parent = resolveUntrackedParent(untrackedItem);
    collectionStore.navigateToCollection(parent);
    stage.deselectAllItems();
    projectStore.selectUntrackedItem(untrackedItem);
    stage.selectedItem = untrackedItem;
    stage.selectedItems = [untrackedItem];
    stage.firstSelectedItemId = untrackedItem.id;
    stage.markedItems = [untrackedItem.id];
    emitter.emit('refresh-browser');
    untrackedItem = await waitForUntrackedItem(item.fullPath, untrackedItem);
    projectStore.selectUntrackedItem(untrackedItem);
    stage.selectedItem = untrackedItem;
    stage.selectedItems = [untrackedItem];
    stage.firstSelectedItemId = untrackedItem.id;
    stage.markedItems = [untrackedItem.id];
    const el = await waitForItemElement(untrackedItem.id);
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' });
    modals.disableAllModals();
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification(t('notifications.errorNavigatingToItem'), error);
  }
};

// Handles enter key press to submit form.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    createCheckPoints();
  }
};

// Updates the visible modified item filter.
const handleModifiedItemsFilterChange = (filter) => {
  selectedModifiedItemsFilter.value = filter;
};

// Normalizes a filesystem path for display comparisons.
const normalizePath = (filePath = '') => {
  return filePath.replace(/\\/g, '/');
};

// Removes an item from the checkpoint list.
const removeItem = (itemPath) => {
  removedPaths.value.push(itemPath);
  if (currentModifiedDisplayPaths.value.length + currentUntrackedPaths.value.length < 1) {
    closeModal();
  }
};

// Toggles the visibility of checkpoint items list.
const toggleShowCheckpointItems = () => {
  showCheckpointItems.value = !showCheckpointItems.value;
};

// Toggles untracked checkpoint candidates for collection-level checkpoint creation.
const toggleIncludeUntracked = () => {
  includeUntrackedInCheckpoint.value = !includeUntrackedInCheckpoint.value;
};

// Toggles extension visibility in modified items.
const toggleHideExtensions = () => {
  hideExtensions.value = !hideExtensions.value;
};

// Toggles full path visibility in modified items.
const toggleShowFullPath = () => {
  showFullPath.value = !showFullPath.value;
};

// Removes an extension from a display string.
const stripExtension = (value = '', extension = '') => {
  if (!extension || !value.toLowerCase().endsWith(extension.toLowerCase())) return value;
  return value.slice(0, -extension.length);
};

// Loads checkpoint candidates directly from the current selection.
const loadSelectedItemsForCheckpoint = () => {
  assetStore.loadingAssetStates = true;
  try {
    const modifiedAssets = stage.selectedItems
      .filter((item) => item.type === 'asset' && item.file_status === 'modified')
      .map((asset) => ({
        asset_id: asset.id,
        type: 'asset',
        name: asset.name,
        icon: asset.icon,
        asset_path: asset.asset_path,
        assignee_id: asset.assignee_id,
        collection_id: asset.collection_id,
        extension: asset.extension,
        display_path: asset.asset_path + asset.extension
      }));

    const untrackedPaths = commonStore.showUntracked
      ? stage.selectedItems
        .filter((item) => item.type === 'untracked_asset' || item.type === 'untracked_collection')
        .filter((item) => canCreateCheckpointForItem(item))
        .map((asset) => ({
          ...asset,
          display_path: asset.asset_path || asset.item_path
        }))
        .filter((asset) => getUntrackedCandidatePath(asset))
      : [];

    assetStore.modifiedAssets = {
      modified: modifiedAssets,
      untracked: untrackedPaths
    };
  } finally {
    assetStore.loadingAssetStates = false;
  }
};

// Loads checkpoint candidates from the active collection context.
const loadCollectionItemsForCheckpoint = async () => {
  let collectionId = null;
  let targetPath = null;
  const selectedItem = stage.selectedItem;
  let selectedCollection;

  if (selectedItem?.type?.includes('collection')) {
    selectedCollection = selectedItem;
  } else {
    selectedCollection = collectionStore.navigatedCollection;
  }

  if (selectedCollection) {
    if (selectedCollection.type === 'collection') {
      collectionId = selectedCollection.id;
    } else if (selectedCollection.type === 'untracked_collection') {
      targetPath = selectedCollection.file_path;
    }
  }

  await collectionStore.reloadItemsForCheckpoint(collectionId, targetPath);
};

// lifecycle hooks
onMounted(async () => {
  if (!trayStates.createMultipleCheckpoints) {
    loadSelectedItemsForCheckpoint();
  } else {
    await loadCollectionItemsForCheckpoint();
  }
  trayStates.screenshot = null;
  trayStates.previewFile = '';
  trayStates.previewFullPath = '';
});

onBeforeUnmount(() => {
  trayStates.createMultipleCheckpoints = true;
  trayStates.createMultipleCheckpointsCollectionPath = '';
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.modified-items {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: .5rem;
  font-size: medium;
  width: 100%;
  box-sizing: border-box;
  padding: .5rem .75rem .5rem .5rem;
  border-radius: var(--small-radius);
  max-height: 40vh;
  overflow: hidden;
  overflow-y: scroll;
  scrollbar-gutter: stable;
}

.modified-items::-webkit-scrollbar {
  width: 4px;
}

.modified-items::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--surface-4);
}

.modified-items::-webkit-scrollbar-track {
  border-radius: 10px;
  margin: .5rem 0;
}

.modified-items-empty-state {
  min-height: 180px;
}

.modified-items-empty-state :deep(.page-state-message) {
  height: auto;
}

.modal-container {
  max-width: min(860px, 90vw);
}

.general-container {
  gap: .5rem;
  padding-bottom: 1rem;
}

.general-container.with-modified-items-history {
  width: 860px;
  max-width: min(860px, 90vw);
}

.checkpoint-create-controls {
  display: flex;
  flex: 0 0 auto;
  flex-direction: column;
  gap: .5rem;
  width: 100%;
  min-width: 0;
}

.checkpoint-untracked-toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .75rem;
  height: 30px;
  padding: 0 .5rem;
  border-radius: var(--small-radius);
  background-color: var(--surface-2);
  color: var(--text);
  cursor: pointer;
  box-sizing: border-box;
}

.checkpoint-untracked-toggle:hover {
  background-color: var(--surface-3);
}

.checkpoint-untracked-toggle-label {
  font-size: 13px;
  font-weight: 200;
}

.checkpoint-create-form {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: .5rem;
  min-width: 0;
}

.checkpoint-create-layout {
  display: flex;
  align-items: stretch;
  gap: .75rem;
  width: 100%;
  min-height: 0;
  box-sizing: border-box;
}

.desktop-input-long {
  margin-top: 0px;
  font-weight: 200;
  color: var(--text);
  max-height: 180px;
}

.with-modified-items-history .desktop-input-long {
  flex: 1;
  height: 100%;
  max-height: none;
  max-height: 180px;
}

.modified-items-count {
  padding-left: .5rem;
  color: var(--text);
  /* background-color: forestgreen; */
  font-weight: 200;
  height: min-content;
  overflow: hidden;
  box-sizing: border-box;
  height: 35px;
  border-radius: var(--small-radius);
}

.modified-items-count-expanded {
  margin-bottom: 1rem;
}

[data-theme="dark"] .modified-items-count:hover{
  background-color: #ffffff15;
}

.modified-items-count:hover {
  background-color: rgba(0, 0, 0, 0.11);
}

.loading-items-count {
  padding-left: .5rem;
  color: var(--text);
  justify-content: flex-start;
}

.import-prompt {
  padding: 1rem .5rem;
}

.modified-items-history {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 300px;
  max-width: 390px;
  height: 420px;
  min-height: 0;
  overflow: hidden;
  border-left: var(--transparent-line);
  padding-left: .75rem;
  box-sizing: border-box;
}

.modified-items-history .modified-items {
  flex: 1;
  height: auto;
  max-height: none;
  min-height: 0;
  padding: .5rem .75rem .5rem .5rem;
}

.modified-items-inline {
  display: flex;
  flex-direction: column;
  width: 100%;
  min-width: 0;
}

.modified-items-tabs {
  padding: 0.3rem 0.5rem;
  display: flex;
  background-color: var(--bg);
  width: 100%;
  max-width: 250px;
  border-radius: var(--very-large-radius);
}

.modified-items-tabs-empty {
  background: none;
}

.modified-items-tabs-header {
  margin-bottom: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  border-radius: var(--very-large-radius);
}

.modified-items-tabs-options {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0 0.5rem;
}

.modified-items-tabs :deep(.modified-items-tab-alert) {
  filter: brightness(0) saturate(100%) invert(60%) sepia(72%) saturate(489%) hue-rotate(1deg) brightness(92%) contrast(90%);
}

[data-theme="dark"] .modified-items-tabs :deep(.modified-items-tab-alert) {
  filter: brightness(0) saturate(100%) invert(88%) sepia(45%) saturate(566%) hue-rotate(359deg) brightness(97%) contrast(92%);
}

.modified-items-tabs :deep(.modified-items-tab-danger),
[data-theme="dark"] .modified-items-tabs :deep(.modified-items-tab-danger) {
  filter: brightness(0) saturate(100%) invert(18%) sepia(95%) saturate(7471%) hue-rotate(347deg) brightness(88%) contrast(93%);
}

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
}

.loading-children-icon {
  width: 20px;
  height: 20px;
  overflow: hidden;
  padding: 0px;
  animation: loadingRotate .5s linear infinite;
}

.refresh-label{
  font-style: italic;
  font-size: 14px;
  color: var(--text);
  opacity: 0.7;
}
</style>
