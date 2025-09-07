<template>
  <div ref="taskItem" class="task-item-main" v-return="launchSelectedTask" v-esc="handleEscKey" v-stop-propagation
    :style="commonStore.useGrid ? gridStyles : itemHeightStyles" :class="{
      'task-item-grid': commonStore.useGrid,
      'task-item-selected': stage.markedItems.includes(task.id) && !isGhost,
      'task-item-cut': stage.cutItems.map((item) => item.id).includes(task.id) && !isGhost,
      'task-item-only-selected': stage.markedItems.length === 1 && stage.firstSelectedItemId === task.id && !isGhost,
      'task-item-last-selected': stage.lastSelectedItemId === task.id && !isGhost,
      'task-item-child': task.parent_id,
      'drop-zone-hovered': isHovered
    }" @dblclick="launchTaskCommand()">

    <div v-if="!commonStore.useGrid" class="task-spacer" v-tooltip="taskTypeName" @click="console.log(task)">
      <span v-if="isUntracked" class="single-action-button single-action-button-disabled">
        <img class="small-icons entity-collapsed" :src="getAppIcon('generic')">
      </span>
      <span v-else class="single-action-button single-action-button-disabled">
        <img class="small-icons entity-collapsed" :src="getAppIcon(task.task_type_icon)">
      </span>
    </div>

    <div v-if="commonStore.useGrid" class="main-task-item-grid">
      <div class="main-task-item-grid-icon">

        <div v-if="commonStore.showThumbs && task.preview" class="task-item-preview-container">
          <div class="task-item-preview-image">
            <img v-if="task.preview" class="screenshot-thumb" :src="task.preview">
            <img v-else class="screenshot-thumb" src='/page-states/no_image.png'>
          </div>
        </div>

        <div v-else class="task-item-icon-container">
          <img v-if="task.icon && !isUntracked" class="gigantic-icons no-filter " :src="task.icon">
          <img v-else-if="isUntracked" class="gigantic-icons " :src="getAppIcon(getFileTypeIcon(task))" @error="$event.target.src = getAppIcon('file')">
          <span v-else class="app-ext">
          </span>
        </div>

      </div>
      <div class="main-task-item-grid-meta">
        {{ taskName }}
      </div>
    </div>

    <div v-else class="main-task-item-root">

      <div class="task-item-container drop-zone">

        <div v-if="commonStore.showThumbs && task.preview" class="task-item-preview-container">
          <div class="task-item-preview-image">
            <img v-if="task.preview" class="screenshot-thumb" :src="task.preview">
            <img v-else class="screenshot-thumb" src='/page-states/no_image.png'>
          </div>
        </div>

        <div class="task-item-icon-container">
          <img v-if="task.icon &&!isUntracked" class="large-icons no-filter" :src="task.icon">
          <img v-else-if="isUntracked" class="large-icons " :src="getAppIcon(getFileTypeIcon(task))" @error="$event.target.src = getAppIcon('file')">
          <span v-else class="app-ext">
          </span>
        </div>

        <div class="task-item-content selection-area">
          <div v-if="!isEditing" class="task-item-details">
            {{ taskName }}
          </div>

          <div v-else class="rename-input">
            <input spellcheck="false" v-model="editableTaskName" class="input-short" type="text" placeholder="Task name"
              v-focus @keydown.enter="handleEnterKey" @keydown.esc="handleEscKey" />
            <ActionButton :isDisabled="!isNameChanged" :icon="getAppIcon('check')" v-tooltip="'Confirm'"
              @click="confirmRename" />
            <ActionButton :icon="getAppIcon('close')" v-tooltip="'Cancel'" @click="cancelRename" />
          </div>
        </div>

        <!-- task assignation -->
        <div v-if="!isEditing && !isUntracked && (!task.is_resource || isCurrentUser)"
          class="task-item-assignee-container">
          <ActionButton v-if="userStore.canDo('assign_task') && !statusMenuDisplayed && !task.assignee_id"
            :icon="getAppIcon('person-plus')" v-tooltip="'Assign Task'" @click="prepAssignTask(index, task, $event)" />

          <div v-else-if="task.assignee_id" @click="prepAssignTask(index, task, $event)" v-stop-propagation
            class="task-item-assignee">
            <span v-tooltip="userFullName" class="single-action-button">
              <div class="profile-picture" :style="{ backgroundColor: profileColor(task.assignee_id) }">
                <img v-if="userPhoto" class="profile-img" :src="userPhoto">
                <img v-else class="profile-img" :src="getAppIcon('person')">
              </div>
            </span>
          </div>
        </div>

        <div v-else class="task-item-assignee-container">
          <ActionButton v-if="userStore.canDo('assign_task') && !statusMenuDisplayed && !task.assignee_id && !isUntracked"
            :icon="getAppIcon('person-plus')" v-tooltip="'Assign Task'" @click="prepAssignTask(index, task, $event)" />
        </div>

        <!-- task status -->
        <div v-if="!isEditing && !isUntracked && (!task.is_resource || isCurrentUser)" class="task-item-status-root">
          <StatusMenu @statusSelected="closeStatusMenu" v-if="statusMenuDisplayed" />

          <div :class="{ 'is-disabled': stage.operationActive }" v-else class="task-item-status-container"
            v-stop-propagation @click="toggleDisplayStatusMenu(index, task, $event)">
            <div class="task-item-status" :style="{ backgroundColor: task.status.color }">
              {{ task.status.short_name }}
            </div>
          </div>
        </div>

        <div v-else-if="!isEditing && !isUntracked" class="task-item-status-root">

          <div class="task-item-status-container" v-stop-propagation>
            <div class="task-item-status" :style="{ backgroundColor: task.status.color, padding: 3 + 'px' }">
            </div>
          </div>
        </div>

        <!-- task actions -->
        <div v-if="!isEditing && !isUntracked && !statusMenuDisplayed && !task.is_link" class="task-item-actions">
          <ActionButton v-if="userStore.canDo('view_checkpoint')" :icon="getAppIcon('layers')" v-tooltip="'Checkpoints'"
            v-stop-propagation @click="viewCheckpoints(index, task, $event)" />
          <div v-if="userStore.canDo('pull_chunk')" class="file-state">
            <ActionButton :icon="getAppIcon('circle-check-go')" :noFilter="true" @click="handleClick(index, task, $event)"
              v-tooltip="'No changes'" v-if="task.file_status == 'normal'" />
            <ActionButton :icon="getAppIcon('circle-check-alert')" :noFilter="true" v-tooltip="'Outdated - Click to update'"
              v-else-if="task.file_status == 'outdated'" @click="revertTask(index, task, $event)" />
            <ActionButton :icon="getAppIcon('layers-plus-alert')" :noFilter="true" v-tooltip="'Modified - Assigned to someone else'"
              v-else-if="task.file_status == 'modified' && !canModify" @click="canModifyPopUpModal()" />
            <ActionButton :icon="getAppIcon('layers-plus-alert')" :noFilter="true" v-tooltip="'Modified - Click to add Checkpoint'"
              v-else-if="task.file_status == 'modified' && userStore.canDo('create_checkpoint')"
              @click="prepCreateCheckpoint(index, task, $event)" />
            <ActionButton :icon="getAppIcon('jigsaw')" v-tooltip="'File missing - Click to build'"
              v-else-if="task.file_status == 'rebuildable'" @click="revertTask(index, task, $event)" />
            <ActionButton :icon="getAppIcon('alert')" :noFilter="true" v-tooltip="'Task missing - Resync your project'"
              v-else-if="task.file_status == 'missing'" />
          </div>
        </div>

        <div v-if="task.is_link" class="task-item-actions">
          <ActionButton :icon="getAppIcon('website')" v-tooltip="'Visit link'" v-stop-propagation @click="openLink()" />
        </div>


        <div v-else-if="isUntracked" class="task-item-actions">
          <ActionButton v-if="userStore.canDo('create_task') || canImport" @click="prepCreateCheckpoint(index, task, $event)"
            :icon="getAppIcon('layers-plus-danger')" :noFilter="true" v-tooltip="'File untracked, click to add.'" />
          <ActionButton v-else :icon="getAppIcon('dot-big-danger')" :noFilter="true" v-tooltip="'File untracked'" />
        </div>


      </div>

    </div>
  </div>
</template>

<script setup>

// imports
import { computed, ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue';
import { isValidWeblink } from '@/lib/pointer';
import { TaskService, CheckpointService, FSService } from "@/../bindings/clustta/services";
import { SyncService } from "@/../bindings/clustta/services";
import utils from '@/services/utils';
import { Events } from "@wailsio/runtime";

// states/store imports
import { useTrayStates } from '@/stores/TrayStates';
import { useAssetStore } from '@/stores/assets';
import { useMenu } from '@/stores/menu';
import { useIconStore } from '@/stores/icons';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { useUserStore } from '@/stores/users';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useCommonStore } from '@/stores/common';
import { useCollectionStore } from '@/stores/collections';
import { useProjectStore } from '@/stores/projects';
import { useDndStore } from '@/stores/dnd';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import StatusMenu from '@/instances/desktop/menus/StatusMenu.vue'
import { Browser } from "@wailsio/runtime";
import { getParentPath } from '@/lib/pathlib';

// states/stores
const userStore = useUserStore();
const iconStore = useIconStore();
const trayStates = useTrayStates();
const menu = useMenu();
const panes = usePaneStore();
const stage = useStageStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const assetStore = useAssetStore();
const commonStore = useCommonStore();
const collectionStore = useCollectionStore();
const projectStore = useProjectStore();
const dndStore = useDndStore();

// emits
const emit = defineEmits(['toggle-edit-mode', 'expand', 'refreshData']);

// props
const props = defineProps({
  task: Object,
  entityId: { type: String, default: '' },
  index: Number,
  isChild: { type: Boolean, default: false },
  isUntracked: { type: Boolean, default: false },
  isGhost: { type: Boolean, default: false },
});

const gridStyles = computed(() => ({
  minWidth: commonStore.gridSize + 'px',
  height: commonStore.gridSize + 'px',
}));

const itemHeightStyles = computed(() => ({
  height: `calc(100% - ${commonStore.listItemGap}px)`,
}));

// refs
const taskItem = ref(null);
const isExpanded = ref(false);
const displayStatusMenu = ref(false);
const taskTypeName = ref('');


const getFileTypeIcon = (task) => {
  const extension = task.extension?.toLowerCase() || '';

  const imageFormats = ['.png', '.exr', '.jpg', '.jpeg', '.gif', '.bmp', '.tiff', '.webp', '.svg'];
  const videoFormats = ['.mp4', '.mkv', '.avi', '.mov', '.wmv', '.flv', '.webm'];
  const audioFormats = ['.mp3', '.wav', '.flac', '.aac', '.ogg'];
  const archiveFormats = ['.zip', '.rar', '.7z', '.tar', '.gz', '.bz2'];
  const textFormats = ['.txt', '.md', '.rtf'];
  const codeFormats = ['.js', '.ts', '.css', '.html', '.vue', '.py', '.java', '.cpp', '.c', '.go', '.rs'];
  const spreadsheetFormats = ['.xls', '.xlsx', '.csv'];
  const presentationFormats = ['.ppt', '.pptx'];
  const wordFormats = ['.doc', '.docx'];

  if (imageFormats.includes(extension)) {
    return 'image'
  } else if (videoFormats.includes(extension)) {
    return 'video-camera'
  } else if (audioFormats.includes(extension)) {
    return 'music'
  } else if (extension === '.pdf') {
    return 'file-pdf'
  } else if (archiveFormats.includes(extension)) {
    return 'file-zip'
  } else if (textFormats.includes(extension)) {
    return 'file-text'
  } else if (codeFormats.includes(extension)) {
    return 'file-code'
  } else if (spreadsheetFormats.includes(extension)) {
    return 'file-excel'
  } else if (presentationFormats.includes(extension)) {
    return 'file-powerpoint'
  } else if (wordFormats.includes(extension)) {
    return 'file-word'
  } else {
    return 'file'
  }
};

const untrackedItemIcon = computed(() => {
  return getFileTypeIcon(props.task);
});


const isHovered = computed(() => { return dndStore.targetItemId === props.task.id });
const canModify = computed(() => {
  let assigneeId = props.task.assignee_id
  if (assigneeId == "") {
    return true
  } else if (assigneeId == userStore.user.id) {
    return true
  } else {
    return false
  }
});

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const canImport = computed(() => {
  let trackedParent = utils.getUntrackedEntityparent(props.task)
  if (props.task.entity_path === "") {
    return false
  }
  return trackedParent && trackedParent.can_modify
});

// Helper function to emit task data updates
const emitTaskUpdates = (taskId, updates) => {
  const updateData = { itemId: taskId, updates };
  
  // Emit to both Browser and VirtuaItem components
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};


const importItem = (index, task, event) => {
  handleClick(index, task, event);
  const inRoot = props.task.entity_path === ""
  const taskPath = props.task.file_path;
  let parentId = ""
  let untrackedParents = []
  let parentPaths = utils.getParentPaths(props.task.entity_path)
  if (!inRoot) {
    for (let parent of parentPaths) {
      parentId = collectionStore.collections.find((item) => item.entity_path === parent)?.id;
      if (parentId !== undefined) {
        break
      }
      untrackedParents.unshift(parent)
    }
  }
  dndStore.untrackedParents = untrackedParents
  dndStore.targetItemId = parentId;
  dndStore.droppedFiles.push(taskPath);
  modals.setModalVisibility('importItemsModal', true);
};


// Toggle the accordion expansion
const toggle = () => {
  isExpanded.value = !isExpanded.value
  emit('expand', props.item)
  updateHeight()
}

// computed properties
const taskName = computed(() => {
  const task = props.task;
  const extension = commonStore.hideExtensions ? '' : task.name ? task.extension : '';
  const taskName = task.name ? task.name : task.extension;
  const isDirectParent = props.task.id === task.entity_id;
  const taskPath = task.task_path.replace(/\//g, ' / ').replace(/^( \/ )?/, '');

  if (commonStore.showFullPath) {
    return taskPath + extension
  }
  if (props.isChild) {
    if (commonStore.showChildEntities) {
      return taskName + extension
    } else {
      return isDirectParent ? (taskName + extension) : taskPath
    }
  } else {
    if (commonStore.viewSearchQuery) {
      return taskPath + extension
    } else {
      return taskName + extension
    }
  }
});

const isEditing = ref(false);
const isAwaitingResponse = ref(false);
const editableTaskName = ref(props.task.name || '');
const statusMenuDisplayed = ref(false);

const toggleEditMode = (event) => {
  if (statusMenuDisplayed.value) {
    statusMenuDisplayed.value = false;
  }
  isEditing.value = !isEditing.value;
  emit('toggle-edit-mode', isEditing.value);
  
  if (isEditing.value) {
    nextTick(() => {
      const inputElement = document.querySelector('.input-short');
      if (inputElement) {
        inputElement.focus();
        inputElement.select();
      }
    });
  }
};

const cancelRename = () => {
  editableTaskName.value = props.task.name || '';
  toggleEditMode();
};

const startRename = () => {
  toggleEditMode();
};

const confirmRename = async () => {
  isAwaitingResponse.value = true;
  await updateTaskName();
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
  console.log('canceled');
  if (isEditing.value) {
    cancelRename();
  }
  if (statusMenuDisplayed.value) {
    statusMenuDisplayed.value = false;
  }
};

const updateTaskName = async () => {

  isAwaitingResponse.value = true;

  let taskId = props.task.id;
  let task = props.task;

  if (props.task.type === 'task') {

    await TaskService.RenameTask(projectStore.activeProject.uri, taskId, editableTaskName.value)
      .then((data) => {
        task.name = editableTaskName.value;
        
        emitTaskUpdates(taskId, [
          { property: 'name', value: editableTaskName.value }
        ]);
        
        isAwaitingResponse.value = false;
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        console.error('Error:', error);
      });
  } else if (props.task.type === 'untracked_task') {
    let oldPath = props.task.file_path
    let newPath = getParentPath(props.task.file_path) + "/" + editableTaskName.value + props.task.extension
    let task = projectStore.findUntrackedTask(props.task.id);
    await FSService.Rename(oldPath, newPath)
      .then((data) => {
        // Update local task data

        emitTaskUpdates(taskId, [
          { property: 'name', value: editableTaskName.value },
          { property: 'file_path', value: newPath }
        ]);
        
        // For untracked tasks, emit 'refresh-browser' as they may not be in the same data structure
        // emitter.emit('refresh-browser');
        
        isAwaitingResponse.value = false;
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        console.error('Error:', error);
      });

  }
}

const isNameChanged = computed(() => {
  const restrictedEntries = [taskName.value, ''];

  const lowerCaseEditableName = editableTaskName.value.toLowerCase();
  const lowerCaseRestrictedEntries = restrictedEntries.map(entry =>
    typeof entry === 'string' ? entry.toLowerCase() : entry
  );

  return !lowerCaseRestrictedEntries.includes(lowerCaseEditableName);
});

const isTaskInFocus = computed(() => {
  return stage.markedItems.length === 1 && stage.firstSelectedItemId === props.task.id && !dndStore.draggedItem
});

const operationsActive = computed(() => {
  return stage.operationActive || !!modals.activeModal || !!menu.activeMenu || isEditing.value || stage.activeStage !== 'browser'
});

Events.On('rename-item', async () => {
  if (operationsActive.value) return
  if (isTaskInFocus.value && userStore.canDo('update_task')) {
    startRename();
  }
});

Events.On('edit-item', async () => {
  if (operationsActive.value) return
  if (isTaskInFocus.value && userStore.canDo('update_task')) {
    modals.setModalVisibility('editAssetModal', true);
  }
});

const triggerRename = () => {
  if (operationsActive.value) return
  if (isTaskInFocus.value && userStore.canDo('update_task')) {
    startRename();
  }
}

Events.On('add-checkpoint', async () => {
  if (operationsActive.value) return
  if (isTaskInFocus.value && userStore.canDo('create_checkpoint')) {
    prepCreateCheckpoint();
  }
});

Events.On('free-item-space', async () => {
  if (operationsActive.value) return
  if (isTaskInFocus.value) {
    if (props.task.type === 'task') {
      prepFreeUpSpacePopUpModal();
    } else if (props.task.type === 'untracked_task') {
      prepDeleteUntrackedTaskPopUpModal();
    }
  }
});

Events.On('delete-item', async () => {
  if (operationsActive.value) return
  if (isTaskInFocus.value && userStore.canDo('delete_task')) {
    panes.setPaneVisibility('projectDetails', true);
    deleteTask();
  }
});

const menuRename = () => {
  if (isTaskInFocus.value && userStore.canDo('update_task')) {
    startRename();
  }
}

const launchSelectedTask = () => {
  if (isEditing.value) return
  if (isTaskInFocus.value && !modals.activeModal) {
    launchTaskCommand();
  }
}

const isOnDisk = computed(() => { return props.task.file_status === 'rebuildable' });

const prepFreeUpSpacePopUpModal = () => {
  trayStates.popUpModalTitle = "Free Up Task Space";
  trayStates.popUpModalMessage = "Are you sure you want to delete this task working files? This will permanently remove all uncheckpointed resources and all task outputs. Please confirm if you wish to proceed.";
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpSpace;
  modals.setModalVisibility('popUpModal', true);
};

const freeUpSpace = async () => {
  let task = assetStore.selectedAsset;
  let taskDir = task.file_path.replace(/\\/g, '/');
  await FSService.DeleteFile(taskDir)
    .then((response) => {
      task.file_status = 'rebuildable';
      assetStore.rebuildableAssetsPath.push(task.task_path)
      assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(taskPath => taskPath !== task.task_path)
      assetStore.modifiedAssetsPath = assetStore.modifiedAssetsPath.filter(taskPath => taskPath !== task.task_path);
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      console.error(error);
    });
  modals.disableAllModals();
};

const deleteTask = async () => {
  if (props.task.type === 'task') {
    let taskId = assetStore.selectedAsset.id;
    TaskService.DeleteTask(projectStore.activeProject.uri, taskId, true)
      .then(async (response) => {
        assetStore.selectedAsset = null;
        stage.markedItems = [];
        emitter.emit('refresh-browser');
      })
      .catch((error) => {
        notificationStore.errorNotification("Task failed to delete.", error)
      });
    let longMessage = `Task of name: ${assetStore.selectedAsset.name} was moved to Trash.`
    notificationStore.addNotification("Task moved to Trash.", longMessage, "success", true);
  } else if (props.task.type === 'untracked_task') {
    prepDeleteUntrackedTaskPopUpModal();
  }
};

const deleteUntrackedItem = () => {
  FSService.DeleteFile(props.task.file_path);
  projectStore.removeUntrackedTask(props.task.id);
  emitter.emit('refresh-browser');
  modals.disableAllModals();
};

const prepDeleteUntrackedTaskPopUpModal = () => {
  trayStates.popUpModalTitle = "Delete";
  trayStates.popUpModalMessage = "Are you sure you want to delete this item? This will permanently remove this item. Please confirm if you wish to proceed.";
  trayStates.popUpModalIcon = 'trash';
  trayStates.popUpModalFunction = deleteUntrackedItem;
  modals.setModalVisibility('popUpModal', true);
};

watch(() => isTaskInFocus.value, (newItems, oldItems) => {
  if (isEditing.value) {
    isEditing.value = false;
    editableTaskName.value = props.task.name || '';
  }
  if (statusMenuDisplayed.value) {
    statusMenuDisplayed.value = false;
  }
}, { deep: true });

const taskTypeIcon = computed(() => {
  const icon = '/types-icons/' + props.task.task_type_icon + '.svg';
  if (icon) {
    return icon
  } else {
    return '/types-icons/other.svg';
  }
});

// methods
const closeStatusMenu = () => {
  props.task.status = assetStore.selectedAsset.status
  props.task.status_id = assetStore.selectedAsset.status_id
  props.task.status_short_name = assetStore.selectedAsset.status_short_name
  statusMenuDisplayed.value = false;
};

const openLink = () => {
  const task = props.task;
  if (task.is_link && isValidWeblink(task.pointer)) {
    Browser.OpenURL(task.pointer)
  }
};
const launchTaskCommand = async () => {
  if (!userStore.canDo('pull_chunk')) {
    return
  }
  const task = props.task;
  if (task.is_link && isValidWeblink(task.pointer)) {
    Browser.OpenURL(task.pointer)
  } else {

    let file_path = task.pointer ? task.pointer : task.file_path
    if (await FSService.Exists(file_path)) {
      FSService.LaunchFile(file_path)
    } else {
      CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [task.id])
        .then(async (response) => {
          let fileStatus = await assetStore.getAssetFileStatus(task)
          props.task.file_status = fileStatus
          FSService.LaunchFile(file_path)
        })
        .catch((error) => {
          console.log(error)
          notificationStore.errorNotification("Error Rebuilding Task", error)
        });
    }

  }
};

const prepCreateCheckpoint = (index, mask, event) => {
  const task = props.task
  assetStore.selectedAsset = task;
  console.log(assetStore.selectedAsset)
  handleClick(index, task, event);
  modals.setModalVisibility('createCheckpointModal', true);
};

const canModifyPopUpModal = () => {
  trayStates.popUpModalTitle = "Warning";
  trayStates.popUpModalMessage = "You cannot modify this task because it is assigned to another user.";
  trayStates.popUpModalIcon = 'help';
  trayStates.popUpModalFunction = null;
  modals.setModalVisibility('popUpModal', true);
};

const downloadCheckpoint = (checkpointId) => {
  notificationStore.cancleFunction = SyncService.CancelSync
  notificationStore.canCancel = true
  SyncService.DownloadCheckpoint(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, checkpointId)
    .then((response) => {
      emit('refreshCheckpoints');
    })
    .catch((error) => {
      console.log(error)
      notificationStore.errorNotification("Error Downloading Checkpoint", error)
    });
};

const revertTask = async (index, task, event) => {
  handleClick(index, task, event);
  const taskId = task.id;

  CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [taskId])
    .then(async (response) => {
      assetStore.rebuildableAssetsPath = assetStore.rebuildableAssetsPath.filter(taskPath => taskPath !== task.task_path)
      assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(taskPath => taskPath !== task.task_path)
      let fileStatus = await assetStore.getAssetFileStatus(task)
      props.task.file_status = fileStatus;
      emitter.emit('get-project-data')
    })
    .catch((error) => {
      console.log(error)
      notificationStore.errorNotification("Error Reverting Task", error)
    });
};

const editParams = () => {
  console.log('escape')
};

const userFullName = computed(() => {
  let user = userStore.getUserData(props.task.assignee_id);
  if (!user) {
    return 'No User'
  } else {
    return `${user.first_name} ${user.last_name}`;
  }
});

const userPhoto = computed(() => {
  return userStore.userProfilePhoto(props.task.assignee_id);
});

const isCurrentUser = computed(() => {
  const user = userStore.user;
  if (!user) {
    return false
  }
  let currentUserId = user.id;
  return props.task.assignee_id === currentUserId
});

const profileColor = (uuid) => {
  const parts = uuid.split('-');
  return '#' + parts[0];
};

const handleClick = (index, task, event) => {
  closeStatusMenu();
  const id = task.id
};

const toggleDisplayStatusMenu = (index, task, event) => {
  handleClick(index, task, event);
  if (!userStore.canDo('change_status')) {
    return
  }
  assetStore.isAssetTaskStatus = true;
  assetStore.selectAsset(task);
  statusMenuDisplayed.value = true;
};

const goToDependencies = (index, task, event) => {
  handleClick(index, task, event);
  assetStore.selectAsset(task);
  stage.setStageVisibility('dependencies', true);
};

const viewCheckpoints = (index, task, event) => {
  // stage.handleClick(event, task);
  // handleClick(index, task, event);
  stage.markedItems = [task.id];
  assetStore.selectAsset(task);
  emitter.emit('view-checkpoints');
  // panes.setPaneVisibility('checkpoints', true);
  panes.showDetailsPane = true;
};

const prepAssignTask = (index, task, event) => {
  if (!userStore.canDo('assign_task')) {
    return
  }
  handleClick(index, task, event);


  const id = task.id;
  assetStore.selectAsset(task);
  stage.markedTasks = [id];
  menu.showContextMenu(event, 'assignMenu', true);
};

const handleClickOutside = (event) => {
  if (statusMenuDisplayed.value) {
    statusMenuDisplayed.value = false;
  }
};

onMounted(async () => {
  emitter.on('renameTask', menuRename);
  document.addEventListener('click', handleClickOutside);
});

onBeforeUnmount(() => {
  emitter.off('renameTask', menuRename);
  document.removeEventListener('click', handleClickOutside);
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.single-action-button-disabled {
  pointer-events: none;
}

.task-collapsed {
  transform: rotate(-90deg);
}

.task-expanded {
  transform: rotate(0deg);
}

.chevron-inactive {
  opacity: .2;
}

.task-item-main {
  z-index: 100000;
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
  outline-offset: -1.5px;
}

.task-item-main:hover {
  outline: var(--transparent-line);
  outline: 1px solid rgb(255, 255, 255);
  outline-offset: -1.5px;
}

.task-item-grid:hover{
  outline: 0px;
  background-color: rgba(255, 255, 255, 0.05);
} 

.task-item-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--blue-steel);
}

.task-item-cut{
  opacity: .5;
}

.task-item-grid {
  align-items: flex-end;
  padding-left: 0px;
  outline: none;
  background-color: transparent;
}

.main-task-item-grid {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  /* background-color: hotpink; */
  overflow: hidden;
}

.main-task-item-grid-icon {
  display: flex;
  /* flex: 1; */
  overflow: hidden;
  height: 100%;
  width: 100%;
  /* background-color: forestgreen; */
  align-items: center;
  justify-content: center;
}

.main-task-item-grid-meta {
  display: flex;
  /* flex: 1; */
  height: 40px;
  min-height: 30px;
  width: 100%;
  /* background-color: royalblue; */
  overflow: hidden;
  align-items: center;
  justify-content: center;
  text-wrap: nowrap;
  text-overflow: ellipsis;
  font-size: 14px;
  font-weight: 300;
}

.task-item-last-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.task-item-only-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.task-item-selected:hover {
  outline: 1px solid rgb(255, 255, 255);
  outline-offset: -1px;
}

.task-item-child {
  padding-left: 0px;
}

.drop-zone-hovered {
  background-color: var(--drop-hover);
}

.main-task-item-root {
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

.task-item-container {
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

.task-spacer {

  position: relative;
  width: min-content;
  width: 36px;
  height: 60px;
  height: 100%;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  overflow: hidden;

  /* background-color: royalblue; */
}

.task-spacer-empty {
  background-color: moccasin;
}

.checkboxes {
  width: 18px;
  height: 18px;
  border-radius: 4px;
  border: 2px solid yellow;
  background: #FFF;
  padding: 10px;
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
  /* background-color: royalblue; */
}

.task-item-preview-container {

  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  padding: .1rem;
  overflow: hidden;
  min-width: 60px;
  height: 100%;
  aspect-ratio: 16 / 9;
  /* background-color: firebrick; */
}

.task-item-preview-image {
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

.task-item-icon-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  min-width: min-content;
  padding: .1rem;
  overflow: hidden;
  height: 100%;
  /* background-color: firebrick; */
}

.task-item-content {
  gap: .4rem;
  /* flex-direction: column; */
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.task-item-meta-container {
  width: 100%;
  display: none;
  justify-content: flex-end;
}

.task-item-main:hover .task-item-meta-container {
  display: flex;
  align-items: center;
  gap: .5rem;
}

.task-item-meta {
  display: flex;
  padding: .2rem;
  box-sizing: border-box;
  align-items: center;
  gap: .4rem;
  /* width: 100%; */
  height: 100%;
  overflow: hidden;
  background-color: rosybrown;
  font-weight: 100;
}

.task-item-details-old {
  /* display: flex; */
  padding: .2rem;
  flex-wrap: nowrap;
  overflow: hidden;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  /* width: 50%; */
  height: 100%;
  height: min-content;
  white-space: nowrap;
  text-overflow: ellipsis;

  /* background-color: forestgreen; */
}

.task-item-details {
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
  color: var(--white);
  justify-content: flex-end;
  direction: rtl;
  text-align: left;
}

.input-short {
  width: 100%;
  height: 100%;
}

.task-item-tag {
  display: flex;
  box-sizing: border-box;
  overflow: hidden;
  padding: .1rem .4rem;
  font-size: 12px;
  background-color: black;
  border-radius: 20px;
}


.task-item-status-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  padding: .4rem;
  height: 100%;
  /* background-color: darkorange; */
  /* flex: 1; */
}

.task-item-status {
  display: flex;
  border-radius: var(--normal-radius);
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  width: 60px;
  padding: .4rem .4rem;
  height: max-content;
  background-color: firebrick;
  font-size: 12px;
  text-transform: uppercase;
  font-weight: 700;
  color: black;
  transition: all 0.2s ease-out;
}

.task-item-status:hover {
  border-radius: 10px;
  transform: scale(1.03);
}

.task-item-actions {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  justify-content: flex-end;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
  /* background-color: darkcyan; */
  min-width: 150px;
  min-width: var(--actions-width);

  /* flex: 1; */
}

.untracked-item-action {
  width: 100%;
  display: none;
  justify-content: flex-end;
}

.untracked-item-alert {
  width: 100%;
  display: flex;
  justify-content: flex-end;
}

.task-item-main:hover .untracked-item-alert {
  display: none;
  align-items: center;
  gap: .5rem;
}

.task-item-main:hover .untracked-item-action {
  display: flex;
  align-items: center;
  gap: .5rem;
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

.task-item-assignee {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
  /* background-color: darkcyan; */
  /* flex: 1; */
}
</style>







