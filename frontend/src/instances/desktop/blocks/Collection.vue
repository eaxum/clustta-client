<template>
  <!-- Grid View Collection Item -->
  <div v-if="commonStore.useGrid" 
    data-file-drop-target
    :id="'drop-' + collection.id"
    class="collection-item-main collection-item-grid" 
    v-return="revealSelectedCollection" 
    v-esc="cancelRename" 
    v-stop-propagation
    v-right-click="cacheCollectionDataIds" 
    @dblclick="exploreCollection(collection)"
    :style="gridStyles"
    :class="{
      'collection-item-grid-selected': stage.markedItems.includes(collection.id) && !isGhost,
      'collection-item-grid-cut': stage.cutItems.map((item) => item.id).includes(collection.id) && !isGhost,
      'collection-item-grid-only-selected': stage.markedItems.length === 1 && stage.firstSelectedItemId === collection.id && !isGhost,
      'collection-item-grid-last-selected': stage.lastSelectedItemId === collection.id && !isGhost,
      'file-drop-target-active': isHovered
    }">
    
    <div class="main-collection-item-grid">

      <div class="main-collection-item-grid-bottom-bar">
        <div v-if="!isEditing && settingsStore.showTypeIcons" class="collection-item-grid-type-icon">
          <img class="small-icons" :src="getAppIcon(collectionTypeIcon)" v-tooltip="collectionTypeName">
        </div>
        
        <div v-if="!isEditing" class="main-collection-item-grid-meta">
          {{collectionName}}
        </div>

        <div v-else class="collection-item-grid-left-section rename-input-grid">
          <RenameInput 
            v-model="editableCollectionName"
            :originalValue="collection.name"
            :placeholder="$t('placeholders.collectionName')"
            @confirm="confirmRename"
            @cancel="cancelRename"
          />
        </div>
        
        <div v-if="!isEditing" class="collection-item-grid-status">
          <ActionButton v-if="loadingCollectionState" :isLoading="true" :icon="getAppIcon('loading')" v-tooltip="$t('blocks.loadingState')" />
          <template v-else-if="!isUntracked">
            <ActionButton v-if="collectionStateFlags.has_outdated" 
              @click="updateCollectionAssets" 
              :icon="getAppIcon('dot-big')" :useAlert="true" :noFilter="true" v-tooltip="$t('blocks.outdatedClickToUpdate')" />
            <ActionButton v-else-if="collectionStateFlags.has_modified" 
              @click="prepAllCheckpointModal(props.collection.collection_path)" 
              :icon="getAppIcon('dot-big')" :useAlert="true" :noFilter="true" v-tooltip="$t('blocks.untrackedModifiedClickCheckpoint')" />
            <ActionButton v-else-if="collectionStateFlags.has_untracked" 
              @click="prepAllCheckpointModal(props.collection.collection_path)" 
              :icon="getAppIcon('dot-big')" :useDanger="true" :noFilter="true" v-tooltip="$t('blocks.untrackedClickCheckpoint')" />
          </template>
          <template v-else-if="collection.type === 'untracked_collection' && props.hasChildren">
            <ActionButton @click="prepAllCheckpointModal(props.collection.collection_path)" 
              v-if="userStore.canDo('create_collection') || canImport || isAssigned"
              :icon="getAppIcon('plus-stone')" :useDanger="true" :noFilter="true" v-tooltip="$t('blocks.addCheckpoints')" />
          </template>
          <template v-else-if="collection.type === 'untracked_collection' && !props.hasChildren">
            <ActionButton @click="" :icon="getAppIcon('dot-big')" :useDanger="true" :noFilter="true"
              v-tooltip="$t('blocks.untrackedCollection')" />
          </template>
        </div>
      </div>
    </div>
  </div>

  <!-- List View Collection Item -->
  <div v-else
    data-file-drop-target
    :id="'drop-' + collection.id"
    class="collection-item-main" 
    v-return="revealSelectedCollection" 
    v-esc="cancelRename" 
    v-stop-propagation
    v-right-click="cacheCollectionDataIds" 
    @dblclick="exploreCollection(collection)"
    :style="itemHeightStyles"
    :class="{
      'collection-item-selected': stage.markedItems.includes(collection.id) && !isGhost,
      'collection-item-cut': stage.cutItems.map((item) => item.id).includes(collection.id) && !isGhost,
      'collection-item-only-selected': stage.markedItems.length === 1 && stage.firstSelectedItemId === collection.id && !isGhost,
      'collection-item-last-selected': stage.lastSelectedItemId === collection.id && !isGhost,
      'file-drop-target-active': isHovered
    }">

    <div v-if="loadingChildren && !isGhost" class="collection-spacer">
      <ActionButton :isLoading="true" :icon="getAppIcon('loading')" 
        v-tooltip="$t('common.loading')" />
    </div>

    <div v-else class="collection-spacer" :class="{ 'collection-spacer-inactive': !!!props.hasChildren }">
      <span @click="expandCollection()" class="single-action-button">
        <img class="small-icons collection-collapsed" :class="{ 'collection-expanded': collection.id in stage.expandedCollections }"
          :src="getAppIcon('chevron-down')">
      </span>
    </div>

    <div class="collection-item-root">
      <div class="collection-item-container ">

        <!-- <div v-if="commonStore.showThumbs && collection.preview" class="collection-item-preview-container">
          <div class="collection-item-preview-image">
            <img v-if="collection.preview" class="screenshot-thumb" :src="collection.preview">
            <img v-else class="screenshot-thumb" src='/page-states/no_image.png'>
          </div>
        </div> -->

        <div v-if="settingsStore.showTypeIcons" @click="console.log(entity)" class="entity-item-icon-container">
          <img class="large-icons" :src="getAppIcon(collectionTypeIcon)" v-tooltip="collectionTypeName">
        </div>

        <div class="collection-item-content selection-area">
          <div v-if="!isEditing" class="collection-item-details">
            {{ collectionName }}
          </div>

          <RenameInput 
            v-else
            v-model="editableCollectionName"
            :originalValue="collection.name"
            :placeholder="$t('placeholders.collectionName')"
            @confirm="confirmRename"
            @cancel="cancelRename"
          />

        </div>

        <div v-if="!isEditing" class="collection-item-meta-container">
          <div v-if="collectionChildren" class="collection-item-meta">
            {{ collectionMeta }}
          </div>
        </div>

        <div v-if="collaboratorsList.length && !isEditing && !isGhost" class="collection-item-assignees">
          <div v-for="(collaborator, index) in collaboratorsList.slice(0, 5)" class="collection-item-assignee-container"
            v-tooltip="collaborator.full_name"
            :class="{ 'collection-item-assignee-container-selected': stage.markedItems.length === 1 && stage.firstSelectedItemId === collection.id && !isGhost }"
            :style="{ zIndex: collaboratorsList.length - index }">
            <ProfilePhoto :assigneeId="collaborator.id" :userPhoto="collaborator.photo"
              :avatarColor="collaborator.avatarColor" />
          </div>
        </div>
        
        
        <div v-if="collaboratorsList.length && collection.is_library && !isEditing && !isGhost" class="horizontal-divider">
        </div>

        <!-- Optimized collection-item-actions using GetCollectionStateFlags -->
        <div v-if="!isEditing && !isUntracked" class="collection-item-actions">
          <ActionButton v-if="loadingCollectionState" :isLoading="true" :icon="getAppIcon('loading')" v-tooltip="$t('blocks.loadingState')" />
          <template v-else>
            <ActionButton v-if="collectionStateFlags.has_modified && !(collection.id in stage.expandedCollections)" 
              @click="prepAllCheckpointModal(props.collection.collection_path)" 
              :icon="getAppIcon('plus-stone')" :useAlert="true" :noFilter="true" v-tooltip="$t('blocks.untrackedModifiedClickCheckpoint')" />
            <ActionButton v-else-if="collectionStateFlags.has_untracked && !(collection.id in stage.expandedCollections)" 
              @click="prepAllCheckpointModal(props.collection.collection_path)" 
              :icon="getAppIcon('plus-stone')" :useDanger="true" :noFilter="true" v-tooltip="$t('blocks.untrackedClickCheckpoint')" />
            <ActionButton v-if="collectionStateFlags.has_outdated && !(collection.id in stage.expandedCollections)" 
              @click="updateCollectionAssets" 
              :icon="getAppIcon('circle-check')" :useAlert="true" :noFilter="true" v-tooltip="$t('blocks.outdatedClickToUpdate')" />
            <ActionButton v-if="collectionStateFlags.has_rebuildable && !(collection.id in stage.expandedCollections)" 
              @click="rebuildCollection" 
              :icon="getAppIcon('jigsaw')" v-tooltip="$t('blocks.itemsMissingClickRebuild')" />
              <ActionButton v-if="collection.is_library" :icon="getAppIcon('library')" v-tooltip="$t('blocks.thisIsALibrary')" />
          </template>
        </div>

        <div v-else-if="!isEditing && collection.type === 'untracked_collection' && props.hasChildren" class="collection-item-actions">
            <ActionButton @click="prepAllCheckpointModal(props.collection.collection_path)" v-if="userStore.canDo('create_collection') || canImport || isAssigned"
              :icon="getAppIcon('plus-stone')" :useDanger="true" :noFilter="true" v-tooltip="$t('blocks.addCheckpoints')" />
        </div>

        <div v-else-if="!isEditing && collection.type === 'untracked_collection' && !props.hasChildren" class="collection-item-actions">
            <ActionButton @click="" :icon="getAppIcon('dot-big')" :useDanger="true" :noFilter="true"
              v-tooltip="$t('blocks.untrackedCollection')" />
        </div>

      </div>


    </div>

  </div>
</template>

<script setup>
// imports
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch, watchEffect } from 'vue';
import { Events } from "@wailsio/runtime";
import emitter from '@/lib/mitt';
import { getParentPath } from '@/lib/pathlib';
import { useI18n } from 'vue-i18n';
import utils from '@/services/utils';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ProfilePhoto from '@/instances/common/components/ProfilePhoto.vue';
import RenameInput from '@/instances/desktop/components/RenameInput.vue';

// services
import { CheckpointService, CollectionService, FSService, SyncService, AssetService } from "@/services";

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useDndStore } from '@/stores/dnd';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { usePaneStore } from '@/stores/panes';
import { useProjectStore } from '@/stores/projects';
import { useSettingsStore } from '@/stores/settings';
import { useStageStore } from '@/stores/stages';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const dndStore = useDndStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const panes = usePaneStore();
const projectStore = useProjectStore();
const settingsStore = useSettingsStore();
const stage = useStageStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

const { t } = useI18n();

// props
const props = defineProps({
  collection: Object,
  collectionChildren: { type: Array, default: [] },
  hasChildren: { type: Boolean, default: false },
  index: Number,
  isGhost: { type: Boolean, default: false },
  isUntracked: { type: Boolean, default: false },
  loadingChildren: { type: Boolean, default: false },
  loadingCollectionState: { type: Boolean, default: false },
});

// emits
const emit = defineEmits(['toggle', 'toggle-edit-mode']);

// refs
const isAwaitingResponse = ref(false);
const isEditing = ref(false);
const renameInput = ref(null);

// events

Events.On('rename-item', async () => {
  if (operationsActive.value) return;
  if (isCollectionInFocus.value && userStore.canDo('update_collection')) {
    startRename();
  }
});

Events.On('edit-item', async () => {
  if (operationsActive.value) return;
  if (isCollectionInFocus.value && userStore.canDo('update_collection')) {
    modals.setModalVisibility('editCollectionModal', true);
  }
});

Events.On('delete-item', async () => {
  if (operationsActive.value) return;
  if (isCollectionInFocus.value && userStore.canDo('delete_collection')) {
    panes.setPaneVisibility('projectDetails', true);
    deleteCollection();
  }
});

Events.On('free-item-space', async () => {
  if (operationsActive.value) return;
  if (isCollectionInFocus.value) {
    if (props.collection.type === 'collection') {
      prepFreeUpSpacePopUpModal();
    } else if (props.collection.type === 'untracked_collection') {
      prepDeleteUntrackedCollectionPopUpModal();
    }
  }
});

// computed
// Checks if the user can import into this untracked collection.
const canImport = computed(() => {
  let trackedParent = utils.getUntrackedCollectionparent(props.collection);
  if (props.collection.collection_path === "") {
    return false;
  }
  return trackedParent && trackedParent.can_modify;
});

// Returns list of collaborators assigned to this collection.
const collaboratorsList = computed(() => {
  if (props.isUntracked) {
    return [];
  }
  const collection = props.collection;
  if (!collection.assignee_ids?.length) {
    return [];
  }
  const projectCollaborators = userStore.getProjectCollaborators
    .map(user => ({
      ...user,
      id: user.id,
      full_name: `${user.first_name} ${user.last_name}`,
      avatarColor: userStore.userProfileColor(user.id)
    }));
  return projectCollaborators.filter((user) => collection.assignee_ids.includes(user.id));
});

// Returns the state flags for this collection.
const collectionStateFlags = computed(() => {
  return props.collection.collectionStateFlags || {
    has_untracked: false,
    has_modified: false,
    has_outdated: false,
    has_rebuildable: false
  };
});

// Returns the icon name for the collection type.
const collectionTypeIcon = computed(() => {
  if (props.isUntracked) return 'folder';
  if (props.collection.collection_type_icon === 'generic') return 'folder';
  return props.collection.collection_type_icon;
});

// Returns the capitalized collection type name.
const collectionTypeName = computed(() => {
  return utils.capitalizeStr(props.collection?.collection_type_name);
});

// Returns the editable collection name for renaming.
const editableCollectionName = ref(props.collection.name || '')

// Returns the display name for the collection.
const collectionName = computed(() => {
  const isUntracked = props.isUntracked;
  const collection = props.collection;
  const collectionName = collection.name;
  const isDirectParent = props.collection.id === collection.collection_id;
  const itemPath = isUntracked ? collection.item_path : collection.collection_path;
  const collectionPath = itemPath.replace(/\//g, ' / ');

  if (commonStore.showFullPath) {
    return collectionPath;
  }
  if (props.isChild) {
    if (commonStore.showChildCollections) {
      return collectionName;
    } else {
      return isDirectParent ? (collectionName) : collectionPath;
    }
  } else {
    if (commonStore.viewSearchQuery) {
      return collectionPath;
    } else {
      return collectionName;
    }
  }
});

// Returns the meta information for the collection (item count).
const collectionMeta = computed(() => {
  const noOfItems = props.collectionChildren?.length;
  return t('blocks.itemCount', noOfItems);
});

// Returns the grid styles for the collection item.
const gridStyles = computed(() => ({
  minWidth: commonStore.gridSize + 'px',
}));

// Checks if the user is assigned to this collection.
const isAssigned = computed(() => {
  const user = userStore.user;
  if (!user) {
    return false;
  }
  let currentUserId = user.id;
  return props.collection.assignee_ids?.includes(currentUserId);
});

// Checks if the collection is currently focused for selection.
const isCollectionInFocus = computed(() => {
  return stage.markedItems.length === 1 && stage.firstSelectedItemId === props.collection.id && !dndStore.draggedItem;
});

// Checks if the collection is hovered for drag and drop.
const isHovered = computed(() => { return dndStore.targetItemId === props.collection.id; });

// Checks if the collection name has been changed.
const isNameChanged = computed(() => {
  const restrictedEntries = [collectionName.value, ''];

  const lowerCaseEditableName = editableCollectionName.value.toLowerCase();
  const lowerCaseRestrictedEntries = restrictedEntries.map(entry =>
    typeof entry === 'string' ? entry.toLowerCase() : entry
  );

  return !lowerCaseRestrictedEntries.includes(lowerCaseEditableName);
});

// Returns the height styles for the item in list view.
const itemHeightStyles = computed(() => ({
  height: `calc(100% - ${commonStore.listItemGap}px)`,
}));

// Checks if any operations are currently active.
const operationsActive = computed(() => {
  return stage.operationActive || !!modals.activeModal || !!menu.activeMenu || isEditing.value || stage.activeStage !== 'browser';
});

// methods
// Caches collection data IDs (currently placeholder).
const cacheCollectionDataIds = () => {
  // stage.collectionDataIds = collectionDataIds.value;
};

// Cancels the current rename operation.
const cancelRename = () => {
  editableCollectionName.value = props.collection.name;
  if (isEditing.value) {
    toggleEditMode();
  }
};

// Confirms and applies the rename.
const confirmRename = async () => {
  isAwaitingResponse.value = true;
  await updateCollectionName();
  toggleEditMode();
};

// Deletes the collection or prepares to delete untracked collection.
const deleteCollection = async () => {
  if (props.collection.type === 'collection') {
    let collection = collectionStore.selectedCollection;
    CollectionService.DeleteCollection(projectStore.activeProject.uri, collection.id, true)
      .then(async (response) => {
        emitter.emit('refresh-browser');
        collectionStore.selectedCollection = null;
        stage.markedItems = [];
      })
      .catch((error) => {
        console.error(error);
      });
    let longMessage = t('notifications.movedToTrash', { item: collection.name });
    notificationStore.addNotification(t('notifications.movedToTrash', { item: 'Collection' }), longMessage, "success", true);

  } else if (props.collection.type === 'untracked_collection') {
    prepDeleteUntrackedCollectionPopUpModal();
  }
};

// Deletes an untracked item from the file system.
const deleteUntrackedItem = () => {
  FSService.DeleteFolder(props.collection.file_path);
  projectStore.removeUntrackedCollection(props.collection.id);
  emitter.emit('refresh-browser');
  modals.disableAllModals();
};

// Emits collection data updates to related components.
const emitCollectionUpdates = (collectionId, updates) => {
  const updateData = { itemId: collectionId, updates };
  
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

// Expands or collapses the collection in the tree view.
const expandCollection = () => {
  const collection = props.collection;
  stage.expandCollection(collection, props.isUntracked);
  cancelRename();
  emit('toggle', collection.name);
};

// Navigates into the collection to explore its contents.
const exploreCollection = (collection) => {
  collectionStore.navigateToCollection(collection);
  commonStore.navigatorMode = true;
};

// Frees up disk space by deleting working files.
const freeUpSpace = async () => {
  let collection = collectionStore.selectedCollection;
  let collectionDir = collection.file_path.replace(/\\/g, '/');
  await FSService.DeleteFolder(collectionDir)
    .then((response) => {
      emitter.emit('refresh-browser');
      assetStore.refreshCollectionFilesStatus(collection.id);
    })
    .catch((error) => {
      console.error(error);
    });
  modals.disableAllModals();
};

// Returns the icon path for the given icon name.
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon;
};

// Prepares the modal for adding checkpoints to all items.
const prepAllCheckpointModal = (collectionPath) => {
  trayStates.createMultipleCheckpointsCollectionPath = collectionPath;
  modals.setModalVisibility('createMultipleCheckpointsModal', true);
};

// Prepares the delete untracked collection popup modal.
const prepDeleteUntrackedCollectionPopUpModal = () => {
  trayStates.popUpModalTitle = t('common.delete');
  trayStates.popUpModalMessage = t('confirmations.deleteItemPermanently');
  trayStates.popUpModalIcon = 'trash';
  trayStates.popUpModalFunction = deleteUntrackedItem;
  modals.setModalVisibility('popUpModal', true);
};

// Prepares the free up space popup modal.
const prepFreeUpSpacePopUpModal = () => {
  trayStates.popUpModalTitle = t('menus.freeUpCollectionSpace');
  trayStates.popUpModalMessage = t('confirmations.deleteWorkingFiles', { item: 'collection' });
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpSpace;
  modals.setModalVisibility('popUpModal', true);
};

// Rebuilds all rebuildable items in the collection.
const rebuildCollection = async () => {
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  await CollectionService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, props.collection.id)
    .then((data) => {
      assetStore.rebuildableAssetsPath = assetStore.rebuildableAssetsPath.filter(assetPath => !assetPath.startsWith(props.collection.collection_path));
      emitter.emit('refresh-browser');
    }).catch(async (error) => {
      notificationStore.errorNotification(t('notifications.errorRebuildingAll'), error);
    });
};

// Reveals the selected collection in the explorer.
const revealSelectedCollection = () => {
  if (isEditing.value) return;
  if (isCollectionInFocus.value && !modals.activeModal) {
    exploreCollection(props.collection);
  }
};

// Starts the rename operation.
const startRename = () => {
  toggleEditMode();
};

// Toggles the edit mode for renaming.
const toggleEditMode = (event) => {
  isEditing.value = !isEditing.value;
  emit('toggle-edit-mode', isEditing.value);
  
  if (isEditing.value) {
    nextTick(() => {
      const inputElement = renameInput.value?.querySelector('.input-short');
      if (inputElement) {
        inputElement.focus();
        inputElement.select();
      }
    });
  }
};

// Triggers the rename operation if conditions are met.
const triggerRename = () => {
  if (isCollectionInFocus.value && userStore.canDo('update_collection')) {
    startRename();
  }
};

// Updates the collection name in the backend.
const updateCollectionName = async () => {
  if (props.collection.type === 'collection') {
    let collection = props.collection;
    let collectionId = collection.id;
    await CollectionService.RenameCollection(projectStore.activeProject.uri, collectionId, editableCollectionName.value)
      .then((data) => {
        const newCollectionPath = getParentPath(collection.collection_path) 
          ? getParentPath(collection.collection_path) + "/" + editableCollectionName.value 
          : editableCollectionName.value;
        const newFilePath = getParentPath(collection.file_path.replace(/\\/g, '/')) + "/" + editableCollectionName.value;
        
        collection.name = editableCollectionName.value;
        collection.collection_path = newCollectionPath;
        collection.file_path = newFilePath;
        
        emitCollectionUpdates(collectionId, [
          { property: 'name', value: editableCollectionName.value },
          { property: 'collection_path', value: newCollectionPath },
          { property: 'file_path', value: newFilePath }
        ]);
        
        if (collectionStore.selectedCollection?.id === collectionId) {
          collectionStore.selectedCollection.name = editableCollectionName.value;
          collectionStore.selectedCollection.collection_path = newCollectionPath;
          collectionStore.selectedCollection.file_path = newFilePath;
        }
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        notificationStore.errorNotification(t('notifications.renameFailed'), error.message || t('notifications.failedToRenameCollection'));
        console.error('Error:', error);
      });
    } else if (props.collection.type === 'untracked_collection') {
    let oldPath = props.collection.file_path.replace(/\\/g, '/');
    let newPath = getParentPath(oldPath) + "/" + editableCollectionName.value;
    let collectionId = props.collection.id;
    await FSService.Rename(oldPath, newPath)
      .then((data) => {
        emitCollectionUpdates(collectionId, [
          { property: 'name', value: editableCollectionName.value },
          { property: 'file_path', value: newPath }
        ]);
        
        if (projectStore.selectedUntrackedItem?.id === collectionId) {
          projectStore.selectedUntrackedItem.name = editableCollectionName.value;
          projectStore.selectedUntrackedItem.file_path = newPath;
        }
        
        isAwaitingResponse.value = false;
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        notificationStore.errorNotification(t('notifications.renameFailedInUse'), error.message || t('notifications.failedToRenameItem'));
        console.error('Error:', error);
      });
  }
};

// Updates all outdated assets in this collection.
const updateCollectionAssets = async () => {
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  
  const outdatedAssets = await collectionStore.getOutdatedItems(props.collection.id);
  const collectionOutdatedAssets = outdatedAssets.map(asset => asset.asset_path);
  
  if (collectionOutdatedAssets.length === 0) {
    return;
  }
  
  await CheckpointService.RevertAssetPaths(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, collectionOutdatedAssets)
    .then((data) => {
      emitter.emit('refresh-browser');
    }).catch(async (error) => {
      notificationStore.errorNotification(t('notifications.errorUpdatingItems'), error);
    });
};

// watchers
watch(() => isCollectionInFocus.value, (newItems, oldItems) => {
  if (isEditing.value) {
    isEditing.value = false;
    editableCollectionName.value = props.collection.name;
  }
}, { deep: true });

watch(() => props.collectionChildren, () => {
  if (props.collectionChildren.length === 0) {
    const collection = props.collection;
    if (collection.id in stage.expandedCollections) {
      // expandCollection()
    }
  }
});

watchEffect(() => {
  if (!props.hasChildren) {
    if (props.collection.id in stage.expandedCollections) {
      // expandCollection()
    }
  }
});

// lifecycle hooks
onMounted(async () => {
  emitter.on('renameCollection', triggerRename);
});

onBeforeUnmount(() => {
  emitter.off('renameCollection', triggerRename);
});
</script>

<style scoped>
@import "@/assets/desktop.css";

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
  /* background-color: hsl(var(--destructive)); */
  border-radius: var(--small-radius);
}

.loading-children-icon {
  width: 20px;
  height: 20px;
  overflow: hidden;
  padding: 0px;
  animation: loadingRotate .5s linear infinite;
}

.single-action-button-disabled {
  pointer-events: none;
}

.collection-collapsed {
  transform: rotate(-90deg);
}

.collection-expanded {
  transform: rotate(0deg);
}

.chevron-inactive {
  opacity: .2;
}

.collection-item-main {
  display: flex;
  gap: .2rem;
  color: hsl(var(--foreground));
  align-items: center;
  padding-left: .5rem;
  box-sizing: border-box;
  width: 100%;
  justify-content: flex-end;
  align-items: center;
  border-radius: var(--large-radius);
  overflow: hidden;
  padding-right: 0px;
  background-color: hsl(var(--muted));
  border: 1px solid hsl(var(--border));
  
  border-radius: var(--large-radius);
  transition: all .2s ease-out;
}

.collection-item-main:hover {
  background-color: hsl(var(--accent));
  border-radius: var(--small-radius);
  border: 1px solid hsl(var(--border));
}

.collection-item-grid:hover {
  background-color: hsl(var(--accent));
  border-radius: var(--small-radius);
} 


.collection-item-main-selected {
  border: 1px solid hsl(var(--border));
  
  background-color: hsl(var(--background));
  background-color: var(--collection-item-selected);
  background-color: hsl(var(--destructive));
}


.collection-item-selected {
  outline: 1px solid hsl(var(--primary));
  border: 1px solid hsl(var(--border));
  
  background-color: hsl(var(--primary) / 0.15);
}

.collection-item-selected:hover {
  background-color: hsl(var(--primary) / 0.3);
}

.collection-item-cut{
  opacity: .5;
}

.collection-item-grid{
  align-items: flex-end;
  padding-left: 0px;
  padding-left: .5rem;
  /* padding: .5rem; */
  align-items: center;
  justify-content: center;
  background-color: hsl(var(--muted));
  border: 1px solid hsl(var(--border));
  
  height: min-content;
  min-height: 50px;
  max-height: 50px;
  box-sizing: border-box;
}

.collection-item-grid-selected {
  border: 1px solid hsl(var(--border));
  
  background-color: hsl(var(--primary) / 0.15);
}

.collection-item-grid-selected:hover {
  background-color: hsl(var(--primary) / 0.3);
}

.collection-item-grid-cut {
  opacity: .5;
}

.collection-item-grid-last-selected {
  border: 1px solid hsl(var(--border));
  
  background-color: hsl(var(--primary) / 0.3);
}

.collection-item-grid-only-selected {
  border: 1px solid hsl(var(--border));
  
  background-color: hsl(var(--primary) / 0.3);
}

.collection-item-grid-only-selected:hover {
  border: 1px solid hsl(var(--border));
  
  background-color: hsl(var(--primary) / 0.3);
}

.main-collection-item-grid{
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.main-collection-item-grid-bottom-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .2rem;
  padding: .5rem .5rem;
  /* padding-top: 1rem; */
  min-height: 50px;
  box-sizing: border-box;
}

.collection-item-grid-type-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.collection-item-grid-status {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.main-collection-item-grid-icon{
  display: flex;
  overflow: hidden;
  height: 100%;
  width: 100%;
  background-color: hsl(var(--muted));
  border-radius: var(--normal-radius);
  align-items: center;
  justify-content: center;
}

.main-collection-item-grid-meta{
  display: block;
  flex: 1;
  text-align: left;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 14px;
  font-weight: 300;
  min-width: 0;
  margin-left: .3rem;
  line-height: 1.2;
}

.collection-item-last-selected {
  border: 1px solid hsl(var(--border));
  
  background-color: hsl(var(--primary) / 0.3);
}

.collection-item-only-selected {
  border: 1px solid hsl(var(--border));
  
  background-color: hsl(var(--primary) / 0.3);
}

.collection-item-only-selected:hover {
  border: 1px solid hsl(var(--border));
  
  background-color: hsl(var(--primary) / 0.3);
}

.collection-item-root {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: hsl(var(--foreground));
  align-items: center;
  padding: .1rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  border-radius: var(--large-radius);
  overflow: hidden;
  padding-right: 0px;
}

.collection-item-container {
  display: flex;
  gap: .5rem;
  color: hsl(var(--foreground));
  align-items: center;
  padding: .2rem .4rem;
  box-sizing: border-box;
  width: 100%;
  height: 50px;
  justify-content: space-between;
  transition: all .3s ease-out;
}

.collection-child-root {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: hsl(var(--foreground));
  align-items: center;
  box-sizing: border-box;
  width: 100%;
  overflow: hidden;
}


.collection-child-root-collapsed {
  height: 0px;
}

.collection-spacer {
  position: relative;
  width: min-content;
  width: 36px;
  height: 60px;
  height: 100%;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.collection-spacer-inactive {
  opacity: .2;
  pointer-events: none;
}

.collection-spacer-empty {
}

.action-column {
  text-align: center;
  padding: 2px;
  box-sizing: border-box;
  transition: all .3s ease-in;
}

.checkbox-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  padding: .1rem;
  overflow: hidden;
  min-width: min-content;
  height: 100%;
}

.collection-item-preview-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  padding: .1rem;
  overflow: hidden;
  min-width: 60px;
  height: 100%;
  aspect-ratio: 16 / 9;
}

.collection-item-preview-image {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  height: 100%;
  aspect-ratio: 16 / 9;
  background-color: hsl(var(--border));
  border-radius: 5px;
}

.collection-item-icon-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  min-width: min-content;
  padding: .1rem;
  overflow: hidden;
  height: 100%;
}

.collection-item-content {
  gap: .4rem;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.collection-item-meta-container {
  width: 100%;
  width: max-content;
  display: none;
  flex-wrap: nowrap;
  text-wrap: nowrap;
  justify-content: flex-end;
  background-color: hsl(var(--destructive));
  width: min-content;
  
  color: hsl(var(--foreground));
  padding: .2rem .6rem;
  border-radius: var(--tiny-radius);
  font-size: 12px;
  overflow: hidden;
  border: 1px solid hsl(var(--border));
  background-color: hsl(var(--background));
  /* display: flex; */
  align-items: center;
  justify-content: center;
  height: max-content;
  box-sizing: border-box;
  /* overflow: hidden; */
}

.collection-item-meta {
  display: flex;
  padding: 0 .4rem;
  box-sizing: border-box;
  align-items: center;
  height: 100%;
  font-weight: 400;
}

.collection-item-main:hover .collection-item-meta-container {
  display: flex;
  align-items: center;
  gap: .5rem;
}

.untracked-item-action {
  width: 100%;
  display: none;
  justify-content: flex-end;
}

.collection-item-main:hover .untracked-item-action {
  display: flex;
  align-items: center;
  gap: .5rem;
}

.untracked-item-alert {
  width: 100%;
  display: flex;
  justify-content: flex-end;
}

.collection-item-main:hover .untracked-item-alert {
  display: none;
  align-items: center;
  gap: .5rem;
}


.collection-item-details {
  padding: .2rem;
  flex-wrap: nowrap;
  overflow: hidden;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  height: min-content;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-size: 14px;
}

.input-short {
  width: 100%;
  height: 100%;
}


.collection-item-tag {
  display: flex;
  box-sizing: border-box;
  overflow: hidden;
  padding: .1rem .4rem;
  font-size: 12px;
  background-color: hsl(var(--background));
  border-radius: 20px;
}


.collection-item-status-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  padding: .4rem;
  height: 100%;
}

.collection-item-status {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  width: 60px;
  padding: .4rem .4rem;
  height: max-content;
  height: 100%;
  background-color: hsl(var(--destructive));
  font-size: 12px;
  text-transform: uppercase;
  font-weight: 700;
  color: hsl(var(--background));
  transition: all 0.2s ease-out;
}

.collection-item-actions {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
  justify-content: flex-end;
}

.file-state {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
}

.collection-item-assignees {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: flex-end;
  align-items: center;
  height: 100%;
  /* background-color: hsl(var(--destructive)); */
  padding-right: .4rem;
}

.collection-item-assignee-container {
  height: 26px;
  min-width: 26px;
  overflow: hidden;
  display: flex;
  align-items: center;
  border-radius: 26px;
  outline: 3px solid hsl(var(--muted));
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: hsl(var(--background));
  margin-right: -0.5rem;
}

.collection-item-assignee-container-selected {
  outline: 3px solid hsl(var(--primary) / 0.3);
}

.profile-spacer {
  /* background-color: hsl(var(--destructive)); */
  overflow: hidden;
  display: flex;
  align-items: center;
  border-radius: 20px;
  height: 24px;
  width: 24px;
  /* padding: 5px; */
}

.rename-input-grid {
  display: flex !important;
  align-items: center;
  gap: 0.3rem;
  width: 100%;
  box-sizing: border-box;
}

.rename-input-grid .input-grid {
  box-sizing: border-box;
  flex: 1;
  min-width: 0;
  font-size: 14px;
  border-radius: var(--very-large-radius);
  height: 100%;
  color: hsl(var(--foreground));
}

.collection-item-grid-left-section {
  display: flex;
  align-items: center;
  gap: .3rem;
  width: 100%;
  overflow: hidden;
}
</style>