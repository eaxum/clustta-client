<template>
  <div class="entity-item-main" v-return="revealSelectedEntity" v-esc="cancelRename" v-stop-propagation
    v-right-click="cacheEntityDataIds" @dblclick="exploreEntity(entity)"
    :style="commonStore.useGrid ? gridStyles : itemHeightStyles"
    :class="{
      'entity-item-grid': commonStore.useGrid,
      'entity-item-selected': stage.markedItems.includes(entity.id) && !isGhost,
      'entity-item-cut': stage.cutItems.map((item) => item.id).includes(entity.id) && !isGhost,
      'entity-item-only-selected': stage.markedItems.length === 1 && stage.firstSelectedItemId === entity.id && !isGhost,
      'entity-item-last-selected': stage.lastSelectedItemId === entity.id && !isGhost,
      'drop-zone-hovered': isHovered
    }">

    <div v-if="!commonStore.useGrid && props.loadingChildren" class="entity-spacer">
      <span class="single-action-button">
        <img class="small-icons loading-children-icon" :src="getAppIcon('loading')">
      </span>
    </div>

    <div v-else-if="!commonStore.useGrid" class="entity-spacer" :class="{ 'entity-spacer-inactive': !!!props.hasChildren }">
      <span @click="expandEntity()" class="single-action-button">
        <img class="small-icons entity-collapsed" :class="{ 'entity-expanded': entity.id in stage.expandedEntities }"
          :src="getAppIcon('chevron-down')">
      </span>
    </div>

    <div v-if="commonStore.useGrid" class="main-entity-item-grid">
      <div class="main-entity-item-grid-icon">

        <div class="entity-item-icon-container">
          <img  class="gigantic-icons" :src="getAppIcon('folder-grid')">
        </div>

      </div>
      <div class="main-entity-item-grid-meta">
        {{entityName}}
      </div>
    </div>

    <div v-else class="entity-item-root">
      <div class="entity-item-container ">

        <div v-if="commonStore.showThumbs && entity.preview" class="entity-item-preview-container">
          <div class="entity-item-preview-image">
            <img v-if="entity.preview" class="screenshot-thumb" :src="entity.preview">
            <img v-else class="screenshot-thumb" src='/page-states/no_image.png'>
          </div>
        </div>

        <div class="entity-item-icon-container">
          <img v-if="isUntracked" class="large-icons" :src="getAppIcon('folder')">
          <img v-else class="large-icons" :src="getAppIcon(entity.entity_type_icon)">
        </div>

        <div class="entity-item-content selection-area">
          <div v-if="!isEditing" class="entity-item-details">
            {{ entityName }}
          </div>

          <div v-else ref="renameInput" class="rename-input">
            <input spellcheck="false" v-model="editableEntityName" class="input-short" type="text"
              placeholder="Collection name" v-focus @keydown.enter="handleEnterKey" @keydown.esc="handleEscKey" />
            <ActionButton :isDisabled="!isNameChanged" :icon="getAppIcon('check')" v-tooltip="'Confirm'"
              @click="confirmRename" />
            <ActionButton :icon="getAppIcon('close')" v-tooltip="'Cancel'" @click="cancelRename" />
          </div>

        </div>

        <div v-if="!isEditing" class="entity-item-meta-container">
          <div v-if="entityChildren" class="entity-item-meta">
            {{ entityMeta }}
          </div>
        </div>

        <div v-if="entity.is_library && !isEditing && !isGhost" class="entity-item-actions">
          <ActionButton v-if="entity.is_library" :icon="getAppIcon('bookmark')" v-tooltip="'This is a Library'" />
        </div>

        <div v-if="collaboratorsList.length && entity.is_library && !isEditing && !isGhost" class="horizontal-divider">
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

        <!-- <div class="entity-item-status-container" v-stop-propagation>
          <div class="entity-item-status" :style="{ backgroundColor: 'transparent' }">
            STATUS
          </div>
        </div> -->

        <div v-if="!isEditing && itemsUntracked && !(entity.id in stage.expandedEntities)"  class="entity-item-actions">
            <ActionButton @click="prepAllCheckpointModal(props.entity.entity_path)" v-if="userStore.canDo('create_entity') || canImport || isAssigned"
              :icon="getAppIcon('layers-plus-danger')" :noFilter="true" v-tooltip="'Add checkpoints.'" />
        </div>

        <div v-else-if="!isEditing && itemsModified && !(entity.id in stage.expandedEntities)"
          class="entity-item-actions">
            <ActionButton @click="prepAllCheckpointModal(props.entity.entity_path)" :icon="getAppIcon('layers-plus-alert')" :noFilter="true"
              v-tooltip="'Add checkpoints.'" />
        
        </div>

        <div v-else-if="!isEditing && itemsOutdated && !(entity.id in stage.expandedEntities)" class="entity-item-actions">
            <ActionButton @click="updateEntityAssets" :icon="getAppIcon('circle-check-alert')" :noFilter="true"
              v-tooltip="'update to latest'" />
        </div>

        <div v-else-if="!isEditing && itemsRebuildable && !(entity.id in stage.expandedEntities)" class="entity-item-actions">
            <ActionButton @click="rebuildEntity" :icon="getAppIcon('jigsaw')"
              v-tooltip="'Rebuild'" />
        </div>

        <div v-else-if="!isEditing && entity.type === 'untracked_entity' && props.hasChildren" class="entity-item-actions">
            <ActionButton @click="prepAllCheckpointModal(props.entity.entity_path)" v-if="userStore.canDo('create_entity') || canImport || isAssigned"
              :icon="getAppIcon('layers-plus-danger')" :noFilter="true" v-tooltip="'Add checkpoints.'" />
        </div>

        <div v-else-if="!isEditing && entity.type === 'untracked_entity' && !props.hasChildren" class="entity-item-actions">
            <ActionButton @click="" :icon="getAppIcon('dot-big-danger')" :noFilter="true"
              v-tooltip="'Untracked Collection'" />
        </div>

      </div>


    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref, onMounted, onBeforeUnmount, nextTick, watch, watchEffect } from 'vue';
import { CheckpointService, EntityService, FSService, SyncService, TaskService } from "@/../bindings/clustta/services";
import utils from '@/services/utils';
import emitter from '@/lib/mitt';
import { Events } from "@wailsio/runtime";

// states/store imports
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useTaskStore } from '@/stores/task';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { useCommonStore } from '@/stores/common';
import { useEntityStore } from '@/stores/entity';
import { useUserStore } from '@/stores/users';
import { useDndStore } from '@/stores/dnd';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useIconStore } from '@/stores/icons';
import { getParentPath } from '@/lib/pathlib';
import { useMenu } from '@/stores/menu';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import ProfilePhoto from '@/instances/common/components/ProfilePhoto.vue'

// states/stores
const userStore = useUserStore();
const iconStore = useIconStore();
const menu = useMenu();
const panes = usePaneStore();
const stage = useStageStore();
const projectStore = useProjectStore();
const entityStore = useEntityStore();
const taskStore = useTaskStore();
const commonStore = useCommonStore();
const dndStore = useDndStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const trayStates = useTrayStates();

// emits
const emit = defineEmits(['toggle', 'toggle-edit-mode']);

// props
const props = defineProps({
  entity: Object,
  index: Number,
  isGhost: { type: Boolean, default: false },
  entityChildren: { type: Array, default: []},
  hasChildren: { type: Boolean, default: false },
  isUntracked: { type: Boolean, default: false },
  loadingChildren: { type: Boolean, default: true },
});

// refs
const renameInput = ref(null);
const isEditing = ref(false);
const isAwaitingResponse = ref(false);

//events
Events.On('rename-item', async () => {
  if (operationsActive.value) return
  if (isEntityInFocus.value && userStore.canDo('update_entity')) {
    startRename();
  }
});

Events.On('edit-item', async () => {
  if (operationsActive.value) return
  if (isEntityInFocus.value && userStore.canDo('update_entity')) {
    modals.setModalVisibility('editEntityModal', true);
  }
});

Events.On('delete-item', async () => {
  if (operationsActive.value) return
  if (isEntityInFocus.value && userStore.canDo('delete_entity')) {
    panes.setPaneVisibility('projectDetails', true);
    deleteEntity();
  }
});

Events.On('free-item-space', async () => {
  if (operationsActive.value) return
  if (isEntityInFocus.value) {
    if (props.entity.type === 'entity') {
      prepFreeUpSpacePopUpModal();
    } else if (props.entity.type === 'untracked_entity') {
      prepDeleteUntrackedEntityPopUpModal();
    }
  }
});

// computed props
const isHovered = computed(() => { return dndStore.targetItemId === props.entity.id });

const entityName = computed(() => {
  const isUntracked = props.isUntracked;
  const entity = props.entity;
  const entityName = entity.name;
  const isDirectParent = props.entity.id === entity.entity_id;
  const itemPath = isUntracked ? entity.item_path : entity.entity_path;
  const entityPath = itemPath.replace(/\//g, ' / ');

  if (commonStore.showFullPath) {
    return entityPath
  }
  if (props.isChild) {
    if (commonStore.showChildEntities) {
      return entityName
    } else {
      return isDirectParent ? (entityName) : entityPath
    }
  } else {
    if (commonStore.viewSearchQuery) {
      return entityPath
    } else {
      return entityName
    }
  }
});

const editableEntityName = ref(entityName.value);

const gridStyles = computed(() => ({
  minWidth: commonStore.gridSize + 'px',
  height: commonStore.gridSize + 'px',
}));

const itemHeightStyles = computed(() => ({
  height: `calc(100% - ${commonStore.listItemGap}px)`,
}));

const collaboratorsList = computed(() => {
  if (props.isUntracked) {
    return []
  }
  const entity = props.entity;
  if (!entity.assignee_ids.length) {
    return []
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

const isNameChanged = computed(() => {
  const restrictedEntries = [entityName.value, ''];

  const lowerCaseEditableName = editableEntityName.value.toLowerCase();
  const lowerCaseRestrictedEntries = restrictedEntries.map(entry =>
    typeof entry === 'string' ? entry.toLowerCase() : entry
  );

  return !lowerCaseRestrictedEntries.includes(lowerCaseEditableName);
});

const isEntityInFocus = computed(() => {
  return stage.markedItems.length === 1 && stage.firstSelectedItemId === props.entity.id && !dndStore.draggedItem
});

const operationsActive = computed(() => {
  return stage.operationActive || !!modals.activeModal || !!menu.activeMenu || isEditing.value || stage.activeStage !== 'browser'
});

const itemsModified = computed(() => {
  return taskStore.modifiedTasksPath.some(taskPath => taskPath.startsWith(props.entity.entity_path));
});

const itemsUntracked = computed(() => {
  return taskStore.untrackedTasksPath.some(taskPath => taskPath.startsWith(props.entity.entity_path));
});

const itemsOutdated = computed(() => {
  return taskStore.outdatedTasksPath.some(taskPath => taskPath.startsWith(props.entity.entity_path));
});

const itemsRebuildable = computed(() => {
  return taskStore.rebuildableTasksPath.some(taskPath => taskPath.startsWith(props.entity.entity_path));
});

//methods
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

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

const cancelRename = () => {
  editableEntityName.value = props.entity.name;
  if (isEditing.value) {
    toggleEditMode();
  }
};

const startRename = () => {
  toggleEditMode();
};

const confirmRename = async () => {
  isAwaitingResponse.value = true;
  await updateEntityName();
  toggleEditMode();
};

const handleEnterKey = () => {
  const inputElement = document.querySelector('.input-short');
  if (!isNameChanged.value) {
    if (inputElement) {
      inputElement.classList.add('shake');
      setTimeout(() => {
        inputElement.classList.remove('shake');
      }, 300);
    }
  } else {
    confirmRename();
  }
};

const handleEscKey = () => {
  cancelRename();
};

const updateEntityName = async () => {

  if (props.entity.type === 'entity') {
    let entity = props.entity;
    let entityId = entity.id;
    
    await EntityService.RenameEntity(projectStore.activeProject.uri, entityId, editableEntityName.value)
      .then((data) => {
        entity.name = editableEntityName.value;
        emitEntityUpdates(entityId, [
          { property: 'name', value: editableEntityName.value }
        ]);
        
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        console.error('Error:', error);
      });
  } else if (props.entity.type === 'untracked_entity') {
    let oldPath = props.entity.file_path
    let newPath = getParentPath(props.entity.file_path) + "/" + editableEntityName.value
    await FSService.Rename(oldPath, newPath)
      .then((data) => {
        // Update local entity data
        let untrackedEntity = projectStore.findUntrackedEntity(props.entity.id);
        untrackedEntity.name = editableEntityName.value;
        untrackedEntity.file_path = newPath;
        
        // For untracked entities, emit 'refresh-browser' as they may not be in the same data structure
        emitter.emit('refresh-browser');
        
        isAwaitingResponse.value = false;
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        console.error('Error:', error);
      });

  }

};

const triggerRename = () => {
  if (isEntityInFocus.value && userStore.canDo('update_entity')) {
    startRename();
  }
};

const updateEntityAssets = async () => {
	notificationStore.cancleFunction = SyncService.CancelSync
	notificationStore.canCancel = true
  let entityOutdatedAssets = taskStore.outdatedTasksPath.filter(taskPath => taskPath.startsWith(props.entity.entity_path))
	await CheckpointService.RevertTaskPaths(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, entityOutdatedAssets)
		.then((data) => {
			taskStore.outdatedTasksPath = taskStore.outdatedTasksPath.filter(taskPath => !taskPath.startsWith(props.entity.entity_path))
			emitter.emit('refresh-browser');
		}).catch(async (error) => {
			notificationStore.errorNotification("Error Rebuilding All", error)
		})
};

const rebuildEntity = async () => {
	notificationStore.cancleFunction = SyncService.CancelSync
	notificationStore.canCancel = true
	await EntityService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, props.entity.id)
		.then((data) => {
			taskStore.rebuildableTasksPath = taskStore.rebuildableTasksPath.filter(taskPath => !taskPath.startsWith(props.entity.entity_path))
			emitter.emit('refresh-browser');
		}).catch(async (error) => {
			notificationStore.errorNotification("Error Rebuilding All", error)
		})
};

const revealSelectedEntity = () => {
  if (isEditing.value) return
  if (isEntityInFocus.value && !modals.activeModal) {
    exploreEntity(props.entity);
  }
}

const prepFreeUpSpacePopUpModal = () => {
  trayStates.popUpModalTitle = "Free Up Entity Space";
  trayStates.popUpModalMessage = "Are you sure you want to delete this entity working files? This will permanently remove all uncheckpointed resources and all entity outputs. Please confirm if you wish to proceed.";
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpSpace;
  modals.setModalVisibility('popUpModal', true);
};

const freeUpSpace = async () => {
  let entity = entityStore.selectedEntity;
  let entityDir = entity.file_path.replace(/\\/g, '/');
  await FSService.DeleteFolder(entityDir)
    .then((response) => {
      emitter.emit('refresh-browser');
      taskStore.refreshEntityFilesStatus(entity.id)
    })
    .catch((error) => {
      console.error(error);
    });
  modals.disableAllModals();
};

const deleteEntity = async () => {
  if (props.entity.type === 'entity') {
    let entity = entityStore.selectedEntity;
    EntityService.DeleteEntity(projectStore.activeProject.uri, entity.id)
      .then(async (response) => {
        emitter.emit('refresh-browser');
        entityStore.selectedEntity = null;
        stage.markedItems = [];
      })
      .catch((error) => {
        console.error(error);
      });
    let longMessage = `Collection of name: ${entity.name} was moved to Trash.`
    notificationStore.addNotification("Collection moved to Trash.", longMessage, "success", true);

  } else if (props.entity.type === 'untracked_entity') {
    prepDeleteUntrackedEntityPopUpModal();
  }
};

const deleteUntrackedItem = () => {
  FSService.DeleteFolder(props.entity.file_path);
  projectStore.removeUntrackedEntity(props.entity.id);
  emitter.emit('refresh-browser');
  modals.disableAllModals();
};

const prepDeleteUntrackedEntityPopUpModal = () => {
  trayStates.popUpModalTitle = "Delete";
  trayStates.popUpModalMessage = "Are you sure you want to delete this item? This will permanently remove this item. Please confirm if you wish to proceed.";
  trayStates.popUpModalIcon = 'trash';
  trayStates.popUpModalFunction = deleteUntrackedItem;
  modals.setModalVisibility('popUpModal', true);
};


const prepAllCheckpointModal = (entityPath) => {
	trayStates.createMultipleCheckpointsEntityPath = entityPath;
	modals.setModalVisibility('createMultipleCheckpointsModal', true);
};

watch(() => isEntityInFocus.value, (newItems, oldItems) => {
  if (isEditing.value) {
    isEditing.value = false;
    editableEntityName.value = props.entity.name;
  }
}, { deep: true });

const canImport = computed(() => {
  let trackedParent = utils.getUntrackedEntityparent(props.entity)
  if (props.entity.entity_path === "") {
    return false
  }
  return trackedParent && trackedParent.can_modify
});

// Helper function to emit entity data updates
const emitEntityUpdates = (entityId, updates) => {
  const updateData = { itemId: entityId, updates };
  
  // Emit to both Browser and VirtuaItem components
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

const isAssigned = computed(() => {
  const user = userStore.user;
  if (!user) {
    return false
  }
  let currentUserId = user.id;
  return props.entity.assignee_ids?.includes(currentUserId)
})

const entityMeta = computed(() => {
  const noOfItems = props.entityChildren?.length;
  const message = noOfItems === 1 ? ' item' : ' items'
  return noOfItems + message
});

const cacheEntityDataIds = () => {
  // stage.entityDataIds = entityDataIds.value;
};


// methods
const expandEntity = () => {
  const entity = props.entity;
  console.log(entity);
  stage.expandEntity(entity, props.isUntracked);
  cancelRename();
  emit('toggle', entity.name);
};

watchEffect(() => {
  if(!props.hasChildren){
    if(props.entity.id in stage.expandedEntities){
      // expandEntity()
    }
  }
});

const exploreEntity = (entity) => {
  entityStore.navigateToEntity(entity);
  commonStore.navigatorMode = true;
};



watch(() => props.entityChildren, () => {
  if (props.entityChildren.length === 0) {
    const entity = props.entity;
    if (entity.id in stage.expandedEntities) {
    }
  }
});

onMounted(async () => {
  emitter.on('renameEntity', triggerRename);
});

onBeforeUnmount(() => {
  emitter.off('renameEntity', triggerRename)
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
  /* background-color: forestgreen; */
  /* width: 100%; */
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
}

.entity-item-main:hover {
  outline: var(--transparent-line);
  outline: 1px solid rgb(255, 255, 255);
  outline-offset: -1.5px;
}


.entity-item-grid:hover{
  outline: 0px;
  background-color: rgba(255, 255, 255, 0.05);
} 


.entity-item-main-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--black-steel);
  background-color: var(--entity-item-selected);
}

.entity-item-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--blue-steel);
}

.entity-item-cut{
  opacity: .5;
}

.entity-item-grid{
  align-items: flex-end;
  padding-left: 0px;
  outline: none;
  background-color: transparent;
}

.main-entity-item-grid{
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.main-entity-item-grid-icon{
  display: flex;
  overflow: hidden;
  height: 100%;
  width: 100%;
  align-items: center;
  justify-content: center;
}

.main-entity-item-grid-meta{
  display: flex;
  height: 40px;
  min-height: 30px;
  width: 100%;
  overflow: hidden;
  align-items: center;
  justify-content: center;
  text-wrap: nowrap;
  text-overflow: ellipsis;
  font-size: 14px;
  font-weight: 300;
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

.entity-item-main-selected:hover {
  outline: 1px solid rgb(255, 255, 255);
  outline-offset: -1px;
}


.entity-drop-zone-hovered {
  width: 100%;
  height: 100%;
  position: absolute;
  opacity: .3;
  background-color: var(--drop-hover);
}

.drop-zone-hovered {
  background-color: var(--drop-hover);
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

.entity-item-meta {
  display: flex;
  padding: .2rem;
  box-sizing: border-box;
  align-items: center;
  gap: .4rem;
  height: 100%;
  overflow: hidden;
  font-weight: 100;
  font-size: 14px;
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
  padding-right: .8rem;
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
</style>