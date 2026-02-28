<template>
  <!-- Grid View Entity Item -->
  <div v-if="commonStore.useGrid" 
    data-file-drop-target
    :id="'drop-' + entity.id"
    class="entity-item-main entity-item-grid" 
    v-return="revealSelectedEntity" 
    v-esc="cancelRename" 
    v-stop-propagation
    v-right-click="cacheEntityDataIds" 
    @dblclick="exploreEntity(entity)"
    :style="gridStyles"
    :class="{
      'entity-item-grid-selected': stage.markedItems.includes(entity.id) && !isGhost,
      'entity-item-grid-cut': stage.cutItems.map((item) => item.id).includes(entity.id) && !isGhost,
      'entity-item-grid-only-selected': stage.markedItems.length === 1 && stage.firstSelectedItemId === entity.id && !isGhost,
      'entity-item-grid-last-selected': stage.lastSelectedItemId === entity.id && !isGhost,
      'file-drop-target-active': isHovered
    }">
    
    <div class="main-entity-item-grid">

      <div class="main-entity-item-grid-bottom-bar">
        <div v-if="!isEditing && settingsStore.showTypeIcons" class="entity-item-grid-type-icon">
          <img class="small-icons" :src="getAppIcon(collectionTypeIcon)" v-tooltip="collectionTypeName">
        </div>
        
        <div v-if="!isEditing" class="main-entity-item-grid-meta">
          {{entityName}}
        </div>

        <div v-else class="entity-item-grid-left-section rename-input-grid">
          <RenameInput 
            v-model="editableEntityName"
            :originalValue="entity.name"
            :placeholder="$t('placeholders.collectionName')"
            @confirm="confirmRename"
            @cancel="cancelRename"
          />
        </div>
        
        <div v-if="!isEditing" class="entity-item-grid-status">
          <ActionButton v-if="loadingCollectionState" :isLoading="true" :icon="getAppIcon('loading')" v-tooltip="$t('blocks.loadingState')" />
          <template v-else-if="!isUntracked">
            <ActionButton v-if="collectionStateFlags.has_outdated" 
              @click="updateEntityAssets" 
              :icon="getAppIcon('dot-big')" :useAlert="true" :noFilter="true" v-tooltip="$t('blocks.outdatedClickToUpdate')" />
            <ActionButton v-else-if="collectionStateFlags.has_modified" 
              @click="prepAllCheckpointModal(props.entity.entity_path)" 
              :icon="getAppIcon('dot-big')" :useAlert="true" :noFilter="true" v-tooltip="$t('blocks.untrackedModifiedClickCheckpoint')" />
            <ActionButton v-else-if="collectionStateFlags.has_untracked" 
              @click="prepAllCheckpointModal(props.entity.entity_path)" 
              :icon="getAppIcon('dot-big')" :useDanger="true" :noFilter="true" v-tooltip="$t('blocks.untrackedClickCheckpoint')" />
          </template>
          <template v-else-if="entity.type === 'untracked_entity' && props.hasChildren">
            <ActionButton @click="prepAllCheckpointModal(props.entity.entity_path)" 
              v-if="userStore.canDo('create_entity') || canImport || isAssigned"
              :icon="getAppIcon('layers-plus')" :useDanger="true" :noFilter="true" v-tooltip="$t('blocks.addCheckpoints')" />
          </template>
          <template v-else-if="entity.type === 'untracked_entity' && !props.hasChildren">
            <ActionButton @click="" :icon="getAppIcon('dot-big')" :useDanger="true" :noFilter="true"
              v-tooltip="$t('blocks.untrackedCollection')" />
          </template>
        </div>
      </div>
    </div>
  </div>

  <!-- List View Entity Item -->
  <div v-else
    data-file-drop-target
    :id="'drop-' + entity.id"
    class="entity-item-main" 
    v-return="revealSelectedEntity" 
    v-esc="cancelRename" 
    v-stop-propagation
    v-right-click="cacheEntityDataIds" 
    @dblclick="exploreEntity(entity)"
    :style="itemHeightStyles"
    :class="{
      'entity-item-selected': stage.markedItems.includes(entity.id) && !isGhost,
      'entity-item-cut': stage.cutItems.map((item) => item.id).includes(entity.id) && !isGhost,
      'entity-item-only-selected': stage.markedItems.length === 1 && stage.firstSelectedItemId === entity.id && !isGhost,
      'entity-item-last-selected': stage.lastSelectedItemId === entity.id && !isGhost,
      'file-drop-target-active': isHovered
    }">

    <div v-if="loadingChildren && !isGhost" class="entity-spacer">
      <ActionButton :isLoading="true" :icon="getAppIcon('loading')" 
        v-tooltip="$t('common.loading')" />
    </div>

    <div v-else class="entity-spacer" :class="{ 'entity-spacer-inactive': !!!props.hasChildren }">
      <span @click="expandEntity()" class="single-action-button">
        <img class="small-icons entity-collapsed" :class="{ 'entity-expanded': entity.id in stage.expandedEntities }"
          :src="getAppIcon('chevron-down')">
      </span>
    </div>

    <div class="entity-item-root">
      <div class="entity-item-container ">

        <!-- <div v-if="commonStore.showThumbs && entity.preview" class="entity-item-preview-container">
          <div class="entity-item-preview-image">
            <img v-if="entity.preview" class="screenshot-thumb" :src="entity.preview">
            <img v-else class="screenshot-thumb" src='/page-states/no_image.png'>
          </div>
        </div> -->

        <div v-if="settingsStore.showTypeIcons" class="entity-item-icon-container">
          <img class="large-icons" :src="getAppIcon(collectionTypeIcon)" v-tooltip="collectionTypeName">
        </div>

        <div class="entity-item-content selection-area">
          <div v-if="!isEditing" class="entity-item-details">
            {{ entityName }}
          </div>

          <RenameInput 
            v-else
            v-model="editableEntityName"
            :originalValue="entity.name"
            :placeholder="$t('placeholders.collectionName')"
            @confirm="confirmRename"
            @cancel="cancelRename"
          />

        </div>

        <div v-if="!isEditing" class="entity-item-meta-container">
          <div v-if="entityChildren" class="entity-item-meta">
            {{ entityMeta }}
          </div>
        </div>

        <div v-if="collaboratorsList.length && !isEditing && !isGhost" class="entity-item-assignees">
          <div v-for="(collaborator, index) in collaboratorsList.slice(0, 5)" class="entity-item-assignee-container"
            v-tooltip="collaborator.full_name"
            :class="{ 'entity-item-assignee-container-selected': stage.markedItems.length === 1 && stage.firstSelectedItemId === entity.id && !isGhost }"
            :style="{ zIndex: collaboratorsList.length - index }">
            <ProfilePhoto :assigneeId="collaborator.id" :userPhoto="collaborator.photo"
              :avatarColor="collaborator.avatarColor" />
          </div>
        </div>
        
        
        <div v-if="collaboratorsList.length && entity.is_library && !isEditing && !isGhost" class="horizontal-divider">
        </div>

        <!-- Optimized entity-item-actions using GetCollectionStateFlags -->
        <div v-if="!isEditing && !isUntracked" class="entity-item-actions">
          <ActionButton v-if="loadingCollectionState" :isLoading="true" :icon="getAppIcon('loading')" v-tooltip="$t('blocks.loadingState')" />
          <template v-else>
            <ActionButton v-if="collectionStateFlags.has_modified && !(entity.id in stage.expandedEntities)" 
              @click="prepAllCheckpointModal(props.entity.entity_path)" 
              :icon="getAppIcon('layers-plus')" :useAlert="true" :noFilter="true" v-tooltip="$t('blocks.untrackedModifiedClickCheckpoint')" />
            <ActionButton v-else-if="collectionStateFlags.has_untracked && !(entity.id in stage.expandedEntities)" 
              @click="prepAllCheckpointModal(props.entity.entity_path)" 
              :icon="getAppIcon('layers-plus')" :useDanger="true" :noFilter="true" v-tooltip="$t('blocks.untrackedClickCheckpoint')" />
            <ActionButton v-if="collectionStateFlags.has_outdated && !(entity.id in stage.expandedEntities)" 
              @click="updateEntityAssets" 
              :icon="getAppIcon('circle-check')" :useAlert="true" :noFilter="true" v-tooltip="$t('blocks.outdatedClickToUpdate')" />
            <ActionButton v-if="collectionStateFlags.has_rebuildable && !(entity.id in stage.expandedEntities)" 
              @click="rebuildEntity" 
              :icon="getAppIcon('jigsaw')" v-tooltip="$t('blocks.itemsMissingClickRebuild')" />
              <ActionButton v-if="entity.is_library" :icon="getAppIcon('library')" v-tooltip="$t('blocks.thisIsALibrary')" />
          </template>
        </div>

        <div v-else-if="!isEditing && entity.type === 'untracked_entity' && props.hasChildren" class="entity-item-actions">
            <ActionButton @click="prepAllCheckpointModal(props.entity.entity_path)" v-if="userStore.canDo('create_entity') || canImport || isAssigned"
              :icon="getAppIcon('layers-plus')" :useDanger="true" :noFilter="true" v-tooltip="$t('blocks.addCheckpoints')" />
        </div>

        <div v-else-if="!isEditing && entity.type === 'untracked_entity' && !props.hasChildren" class="entity-item-actions">
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
  entity: Object,
  entityChildren: { type: Array, default: [] },
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
  if (isEntityInFocus.value && userStore.canDo('update_entity')) {
    startRename();
  }
});

Events.On('edit-item', async () => {
  if (operationsActive.value) return;
  if (isEntityInFocus.value && userStore.canDo('update_entity')) {
    modals.setModalVisibility('editCollectionModal', true);
  }
});

Events.On('delete-item', async () => {
  if (operationsActive.value) return;
  if (isEntityInFocus.value && userStore.canDo('delete_entity')) {
    panes.setPaneVisibility('projectDetails', true);
    deleteEntity();
  }
});

Events.On('free-item-space', async () => {
  if (operationsActive.value) return;
  if (isEntityInFocus.value) {
    if (props.entity.type === 'entity') {
      prepFreeUpSpacePopUpModal();
    } else if (props.entity.type === 'untracked_entity') {
      prepDeleteUntrackedEntityPopUpModal();
    }
  }
});

// computed
// Checks if the user can import into this untracked entity.
const canImport = computed(() => {
  let trackedParent = utils.getUntrackedEntityparent(props.entity);
  if (props.entity.entity_path === "") {
    return false;
  }
  return trackedParent && trackedParent.can_modify;
});

// Returns list of collaborators assigned to this entity.
const collaboratorsList = computed(() => {
  if (props.isUntracked) {
    return [];
  }
  const entity = props.entity;
  if (!entity.assignee_ids?.length) {
    return [];
  }
  const projectCollaborators = userStore.getProjectCollaborators
    .map(user => ({
      ...user,
      id: user.id,
      full_name: `${user.first_name} ${user.last_name}`,
      avatarColor: userStore.userProfileColor(user.id)
    }));
  return projectCollaborators.filter((user) => entity.assignee_ids.includes(user.id));
});

// Returns the state flags for this collection.
const collectionStateFlags = computed(() => {
  return props.entity.collectionStateFlags || {
    has_untracked: false,
    has_modified: false,
    has_outdated: false,
    has_rebuildable: false
  };
});

// Returns the icon name for the collection type.
const collectionTypeIcon = computed(() => {
  if (props.isUntracked) return 'folder';
  if (props.entity.entity_type_icon === 'generic') return 'folder';
  return props.entity.entity_type_icon;
});

// Returns the capitalized collection type name.
const collectionTypeName = computed(() => {
  return utils.capitalizeStr(props.entity?.entity_type_name);
});

// Returns the editable entity name for renaming.
const editableEntityName = ref(props.entity.name || '')

// Returns the display name for the entity.
const entityName = computed(() => {
  const isUntracked = props.isUntracked;
  const entity = props.entity;
  const entityName = entity.name;
  const isDirectParent = props.entity.id === entity.entity_id;
  const itemPath = isUntracked ? entity.item_path : entity.entity_path;
  const entityPath = itemPath.replace(/\//g, ' / ');

  if (commonStore.showFullPath) {
    return entityPath;
  }
  if (props.isChild) {
    if (commonStore.showChildEntities) {
      return entityName;
    } else {
      return isDirectParent ? (entityName) : entityPath;
    }
  } else {
    if (commonStore.viewSearchQuery) {
      return entityPath;
    } else {
      return entityName;
    }
  }
});

// Returns the meta information for the entity (item count).
const entityMeta = computed(() => {
  const noOfItems = props.entityChildren?.length;
  return t('blocks.itemCount', noOfItems);
});

// Returns the grid styles for the entity item.
const gridStyles = computed(() => ({
  minWidth: commonStore.gridSize + 'px',
}));

// Checks if the user is assigned to this entity.
const isAssigned = computed(() => {
  const user = userStore.user;
  if (!user) {
    return false;
  }
  let currentUserId = user.id;
  return props.entity.assignee_ids?.includes(currentUserId);
});

// Checks if the entity is currently focused for selection.
const isEntityInFocus = computed(() => {
  return stage.markedItems.length === 1 && stage.firstSelectedItemId === props.entity.id && !dndStore.draggedItem;
});

// Checks if the entity is hovered for drag and drop.
const isHovered = computed(() => { return dndStore.targetItemId === props.entity.id; });

// Checks if the entity name has been changed.
const isNameChanged = computed(() => {
  const restrictedEntries = [entityName.value, ''];

  const lowerCaseEditableName = editableEntityName.value.toLowerCase();
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
// Caches entity data IDs (currently placeholder).
const cacheEntityDataIds = () => {
  // stage.entityDataIds = entityDataIds.value;
};

// Cancels the current rename operation.
const cancelRename = () => {
  editableEntityName.value = props.entity.name;
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

// Deletes the entity or prepares to delete untracked entity.
const deleteEntity = async () => {
  if (props.entity.type === 'entity') {
    let entity = collectionStore.selectedCollection;
    CollectionService.DeleteCollection(projectStore.activeProject.uri, entity.id, true)
      .then(async (response) => {
        emitter.emit('refresh-browser');
        collectionStore.selectedCollection = null;
        stage.markedItems = [];
      })
      .catch((error) => {
        console.error(error);
      });
    let longMessage = t('notifications.movedToTrash', { item: entity.name });
    notificationStore.addNotification(t('notifications.movedToTrash', { item: 'Collection' }), longMessage, "success", true);

  } else if (props.entity.type === 'untracked_entity') {
    prepDeleteUntrackedEntityPopUpModal();
  }
};

// Deletes an untracked item from the file system.
const deleteUntrackedItem = () => {
  FSService.DeleteFolder(props.entity.file_path);
  projectStore.removeUntrackedEntity(props.entity.id);
  emitter.emit('refresh-browser');
  modals.disableAllModals();
};

// Emits entity data updates to related components.
const emitEntityUpdates = (entityId, updates) => {
  const updateData = { itemId: entityId, updates };
  
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

// Expands or collapses the entity in the tree view.
const expandEntity = () => {
  const entity = props.entity;
  stage.expandEntity(entity, props.isUntracked);
  cancelRename();
  emit('toggle', entity.name);
};

// Navigates into the entity to explore its contents.
const exploreEntity = (entity) => {
  collectionStore.navigateToCollection(entity);
  commonStore.navigatorMode = true;
};

// Frees up disk space by deleting working files.
const freeUpSpace = async () => {
  let entity = collectionStore.selectedCollection;
  let entityDir = entity.file_path.replace(/\\/g, '/');
  await FSService.DeleteFolder(entityDir)
    .then((response) => {
      emitter.emit('refresh-browser');
      assetStore.refreshEntityFilesStatus(entity.id);
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
const prepAllCheckpointModal = (entityPath) => {
  trayStates.createMultipleCheckpointsEntityPath = entityPath;
  modals.setModalVisibility('createMultipleCheckpointsModal', true);
};

// Prepares the delete untracked entity popup modal.
const prepDeleteUntrackedEntityPopUpModal = () => {
  trayStates.popUpModalTitle = t('common.delete');
  trayStates.popUpModalMessage = t('confirmations.deleteItemPermanently');
  trayStates.popUpModalIcon = 'trash';
  trayStates.popUpModalFunction = deleteUntrackedItem;
  modals.setModalVisibility('popUpModal', true);
};

// Prepares the free up space popup modal.
const prepFreeUpSpacePopUpModal = () => {
  trayStates.popUpModalTitle = t('menus.freeUpEntitySpace');
  trayStates.popUpModalMessage = t('confirmations.deleteWorkingFiles', { item: 'entity' });
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpSpace;
  modals.setModalVisibility('popUpModal', true);
};

// Rebuilds all rebuildable items in the entity.
const rebuildEntity = async () => {
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  await CollectionService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, props.entity.id)
    .then((data) => {
      assetStore.rebuildableAssetsPath = assetStore.rebuildableAssetsPath.filter(taskPath => !taskPath.startsWith(props.entity.entity_path));
      emitter.emit('refresh-browser');
    }).catch(async (error) => {
      notificationStore.errorNotification(t('notifications.errorRebuildingAll'), error);
    });
};

// Reveals the selected entity in the explorer.
const revealSelectedEntity = () => {
  if (isEditing.value) return;
  if (isEntityInFocus.value && !modals.activeModal) {
    exploreEntity(props.entity);
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
  if (isEntityInFocus.value && userStore.canDo('update_entity')) {
    startRename();
  }
};

// Updates the collection name in the backend.
const updateCollectionName = async () => {
  if (props.entity.type === 'entity') {
    let entity = props.entity;
    let entityId = entity.id;
    await CollectionService.RenameCollection(projectStore.activeProject.uri, entityId, editableEntityName.value)
      .then((data) => {
        const newEntityPath = getParentPath(entity.entity_path) 
          ? getParentPath(entity.entity_path) + "/" + editableEntityName.value 
          : editableEntityName.value;
        const newFilePath = getParentPath(entity.file_path.replace(/\\/g, '/')) + "/" + editableEntityName.value;
        
        entity.name = editableEntityName.value;
        entity.entity_path = newEntityPath;
        entity.file_path = newFilePath;
        
        emitEntityUpdates(entityId, [
          { property: 'name', value: editableEntityName.value },
          { property: 'entity_path', value: newEntityPath },
          { property: 'file_path', value: newFilePath }
        ]);
        
        if (collectionStore.selectedCollection?.id === entityId) {
          collectionStore.selectedCollection.name = editableEntityName.value;
          collectionStore.selectedCollection.entity_path = newEntityPath;
          collectionStore.selectedCollection.file_path = newFilePath;
        }
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        notificationStore.errorNotification(t('notifications.renameFailed'), error.message || t('notifications.failedToRenameCollection'));
        console.error('Error:', error);
      });
    } else if (props.entity.type === 'untracked_entity') {
    let oldPath = props.entity.file_path.replace(/\\/g, '/');
    let newPath = getParentPath(oldPath) + "/" + editableEntityName.value;
    let entityId = props.entity.id;
    await FSService.Rename(oldPath, newPath)
      .then((data) => {
        emitEntityUpdates(entityId, [
          { property: 'name', value: editableEntityName.value },
          { property: 'file_path', value: newPath }
        ]);
        
        if (projectStore.selectedUntrackedItem?.id === entityId) {
          projectStore.selectedUntrackedItem.name = editableEntityName.value;
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

// Updates all outdated assets in this entity.
const updateEntityAssets = async () => {
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  
  const outdatedTasks = await collectionStore.getOutdatedItems(props.entity.id);
  const entityOutdatedAssets = outdatedTasks.map(task => task.task_path);
  
  if (entityOutdatedAssets.length === 0) {
    return;
  }
  
  await CheckpointService.RevertTaskPaths(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, entityOutdatedAssets)
    .then((data) => {
      emitter.emit('refresh-browser');
    }).catch(async (error) => {
      notificationStore.errorNotification(t('notifications.errorUpdatingItems'), error);
    });
};

// watchers
watch(() => isEntityInFocus.value, (newItems, oldItems) => {
  if (isEditing.value) {
    isEditing.value = false;
    editableEntityName.value = props.entity.name;
  }
}, { deep: true });

watch(() => props.entityChildren, () => {
  if (props.entityChildren.length === 0) {
    const entity = props.entity;
    if (entity.id in stage.expandedEntities) {
      // expandEntity()
    }
  }
});

watchEffect(() => {
  if (!props.hasChildren) {
    if (props.entity.id in stage.expandedEntities) {
      // expandEntity()
    }
  }
});

// lifecycle hooks
onMounted(async () => {
  emitter.on('renameEntity', triggerRename);
});

onBeforeUnmount(() => {
  emitter.off('renameEntity', triggerRename);
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
  /* background-color: crimson; */
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

.entity-collapsed {
  transform: rotate(-90deg);
}

.entity-expanded {
  transform: rotate(0deg);
}

.chevron-inactive {
  opacity: .2;
}

.entity-item-main {
  display: flex;
  gap: .2rem;
  color: var(--white);
  align-items: center;
  padding-left: .5rem;
  box-sizing: border-box;
  width: 100%;
  justify-content: flex-end;
  align-items: center;
  border-radius: 10px;
  overflow: hidden;
  padding-right: 0px;
  background-color: var(--dark-steel);
  outline: var(--transparent-line);
  outline-offset: -1px;
  border-radius: var(--large-radius);
  transition: all .2s ease-out;
}

.entity-item-main:hover {
  background-color: var(--steel);
  border-radius: var(--small-radius);
  outline: 1px solid var(--light-steel);
}

.entity-item-grid:hover {
  background-color: var(--steel);
  border-radius: var(--small-radius);
} 


.entity-item-main-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--black-steel);
  background-color: var(--entity-item-selected);
  background-color: crimson;
}


.entity-item-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--blue-steel);
}

.entity-item-selected:hover {
  background-color: var(--solid-blue-steel);
}

.entity-item-cut{
  opacity: .5;
}

.entity-item-grid{
  align-items: flex-end;
  padding-left: 0px;
  padding-left: .5rem;
  /* padding: .5rem; */
  align-items: center;
  justify-content: center;
  background-color: var(--dark-steel);
  outline: var(--transparent-line);
  outline-offset: -1px;
  height: min-content;
  min-height: 50px;
  max-height: 50px;
  box-sizing: border-box;
}

.entity-item-grid-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--blue-steel);
}

.entity-item-grid-selected:hover {
  background-color: var(--solid-blue-steel);
}

.entity-item-grid-cut {
  opacity: .5;
}

.entity-item-grid-last-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.entity-item-grid-only-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.entity-item-grid-only-selected:hover {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.main-entity-item-grid{
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.main-entity-item-grid-bottom-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .2rem;
  padding: .5rem .5rem;
  /* padding-top: 1rem; */
  min-height: 50px;
  box-sizing: border-box;
}

.entity-item-grid-type-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.entity-item-grid-status {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.main-entity-item-grid-icon{
  display: flex;
  overflow: hidden;
  height: 100%;
  width: 100%;
  background-color: rgba(0, 0, 0, 0.2);
  border-radius: 8px;
  align-items: center;
  justify-content: center;
}

.main-entity-item-grid-meta{
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

.entity-item-last-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.entity-item-only-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.entity-item-only-selected:hover {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.entity-item-root {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: var(--white);
  align-items: center;
  padding: .1rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  border-radius: 10px;
  overflow: hidden;
  padding-right: 0px;
}

.entity-item-container {
  display: flex;
  gap: .5rem;
  color: var(--white);
  align-items: center;
  padding: .2rem .4rem;
  box-sizing: border-box;
  width: 100%;
  height: 50px;
  justify-content: space-between;
  transition: all .3s ease-out;
}

.entity-child-root {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: var(--white);
  align-items: center;
  box-sizing: border-box;
  width: 100%;
  overflow: hidden;
}


.entity-child-root-collapsed {
  height: 0px;
}

.entity-spacer {
  position: relative;
  width: min-content;
  width: 36px;
  height: 60px;
  height: 100%;
  /* background-color: royalblue; */
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.entity-spacer-inactive {
  opacity: .2;
  pointer-events: none;
}

.entity-spacer-empty {
  background-color: moccasin;
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

.entity-item-preview-container {
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

.entity-item-preview-image {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  height: 100%;
  aspect-ratio: 16 / 9;
  background-color: var(--light-steel);
  border-radius: 5px;
}

.entity-item-icon-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  min-width: min-content;
  padding: .1rem;
  overflow: hidden;
  height: 100%;
}

.entity-item-content {
  gap: .4rem;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.entity-item-meta-container {
  width: 100%;
  width: max-content;
  display: none;
  flex-wrap: nowrap;
  text-wrap: nowrap;
  justify-content: flex-end;
  background-color: crimson;
  width: min-content;
  
  color: var(--white);
  padding: .2rem .6rem;
  border-radius: var(--tiny-radius);
  font-size: 12px;
  overflow: hidden;
  outline: var(--transparent-line);
  background-color: var(--black-steel);
  /* display: flex; */
  align-items: center;
  justify-content: center;
  height: max-content;
  box-sizing: border-box;
  /* overflow: hidden; */
}

.entity-item-meta {
  display: flex;
  padding: 0 .4rem;
  box-sizing: border-box;
  align-items: center;
  height: 100%;
  font-weight: 400;
}

.entity-item-main:hover .entity-item-meta-container {
  display: flex;
  align-items: center;
  gap: .5rem;
}

.untracked-item-action {
  width: 100%;
  display: none;
  justify-content: flex-end;
}

.entity-item-main:hover .untracked-item-action {
  display: flex;
  align-items: center;
  gap: .5rem;
}

.untracked-item-alert {
  width: 100%;
  display: flex;
  justify-content: flex-end;
}

.entity-item-main:hover .untracked-item-alert {
  display: none;
  align-items: center;
  gap: .5rem;
}


.entity-item-details {
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


.entity-item-tag {
  display: flex;
  box-sizing: border-box;
  overflow: hidden;
  padding: .1rem .4rem;
  font-size: 12px;
  background-color: black;
  border-radius: 20px;
}


.entity-item-status-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  padding: .4rem;
  height: 100%;
}

.entity-item-status {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  width: 60px;
  padding: .4rem .4rem;
  height: max-content;
  height: 100%;
  background-color: firebrick;
  font-size: 12px;
  text-transform: uppercase;
  font-weight: 700;
  color: black;
  transition: all 0.2s ease-out;
}

.entity-item-actions {
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

.entity-item-assignees {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: flex-end;
  align-items: center;
  height: 100%;
  /* background-color: crimson; */
  padding-right: .4rem;
}

.entity-item-assignee-container {
  height: 26px;
  min-width: 26px;
  overflow: hidden;
  display: flex;
  align-items: center;
  border-radius: 26px;
  outline: 3px solid var(--dark-steel);
  outline-offset: -2px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: black;
  margin-right: -0.5rem;
}

.entity-item-assignee-container-selected {
  outline: 3px solid var(--solid-blue-steel);
}

.profile-spacer {
  /* background-color: red; */
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
  border-radius: 12px;
  height: 100%;
  color: var(--white);
}

.entity-item-grid-left-section {
  display: flex;
  align-items: center;
  gap: .3rem;
  width: 100%;
  overflow: hidden;
}
</style>