<template>
  <!-- Grid View Task Item -->
  <div v-if="commonStore.useGrid" 
    ref="taskItem" 
    class="task-item-main task-item-grid" 
    v-return="launchSelectedTask" 
    v-esc="handleEscKey" 
    v-stop-propagation
    :style="gridStyles" 
    :class="{
      'task-item-grid-selected': stage.markedItems.includes(task.id) && !isGhost,
      'task-item-grid-cut': stage.cutItems.map((item) => item.id).includes(task.id) && !isGhost,
      'task-item-grid-only-selected': stage.markedItems.length === 1 && stage.firstSelectedItemId === task.id && !isGhost,
      'task-item-grid-last-selected': stage.lastSelectedItemId === task.id && !isGhost,
      'task-item-child': task.parent_id,
      'drop-zone-hovered': isHovered,
      'task-item-untracked': isUntracked
    }" 
    @dblclick="launchTaskCommand()">

    <div class="main-task-item-grid">
      
      <!-- Grid Status Menu Overlay -->
      <GridStatusMenu 
        v-if="gridStatusMenuVisible && !isUntracked && task.status" 
        @statusSelected="handleGridStatusSelected"
        @close="closeGridStatusMenu"
      />
      
      <div class="main-task-item-grid-thumb-container">

        <div v-if="task.preview || osThumbnail" class="task-item-preview-container">

          <div class="task-item-preview-image">
            <img class="screenshot-thumb" :src="displayThumbnail">
          </div>

          <!-- Icon container at bottom left when preview is present -->
          <div class="task-item-icon-container task-item-icon-overlay">
            <img v-if="task.icon" class="small-icons no-filter overlay-icons" :src="task.icon">
            <img v-else-if="isUntracked" class="small-icons overlay-icons" :src="getAppIcon(getFileTypeIcon(task))" @error="$event.target.src = getAppIcon('file')">
            <span v-else class="app-ext">
            </span>
          </div>
        </div>

        <div v-else class="task-item-icon-container">
          <img class="gigantic-icons no-filter " :src="displayThumbnail">
        </div>

        <!-- Task assignee overlay in top right corner -->
        <div v-if="!isUntracked && (!task.is_resource || isCurrentUser) && !isEditing" class="task-item-grid-assignee-overlay-top-right">
          <!-- Show assignee profile picture if assigned -->
          <div v-if="task.assignee_id" @click="prepAssignTask(index, task, $event)" v-stop-propagation class="task-item-assignee">
            <span v-tooltip="userFullName" class="single-action-button">
              <div class="profile-picture-grid" :style="{ backgroundColor: profileColor(task.assignee_id) }">
                <img v-if="userPhoto" class="profile-img-grid" :src="userPhoto">
                <img v-else class="profile-img-grid" :src="getAppIcon('person')">
              </div>
            </span>
          </div>
        </div>
        
      </div>
      
      <!-- Bottom bar with task type icon, name, and file status -->
      <div class="main-task-item-grid-bottom-bar">
        
        <!-- Outermost container with relative positioning -->
        <div class="task-item-grid-bottom-bar-wrapper">
          
          <!-- Middle container with slide-up transition and padding for file state -->
          <div v-if="!isEditing" class="task-item-grid-slide-container">
            
            <!-- Row 1: Name/Meta (always visible) -->
            <div class="task-item-grid-meta-row">
              <div class="task-item-grid-type-icon" >
                <img v-if="isUntracked" class="small-icons" :src="getAppIcon('generic')">
                <img v-else class="small-icons" :src="getAppIcon(task.task_type_icon)" v-tooltip="assetTypeName">
              </div>
              
              <div class="main-task-item-grid-meta">
                {{ taskName }}
              </div>
            </div>
            
            <!-- Row 2: Action Buttons (shows on hover) -->
            <div class="task-item-grid-actions-row">
              
              <!-- Untracked label for untracked items -->
              <div v-if="isUntracked" class="task-item-grid-untracked-label">
                <span>Untracked</span>
              </div>
              
              <!-- Task Status -->
              <div v-if="!isUntracked && task.status" @click="openGridStatusMenu" class="task-item-grid-status-display">
                <div class="task-item-status-grid" :style="{ backgroundColor: task.status.color }">
                  {{ task.status.short_name }}
                </div>
                
              </div>
              
              <!-- View Checkpoints button -->
              <div v-if="!task.is_link && !isUntracked && userStore.canDo('view_checkpoint')" class="task-item-grid-checkpoints-button">
                <ActionButton :icon="getAppIcon('layers')" v-tooltip="'View Checkpoints'" @click="viewCheckpoints(index, task, $event)" />
              </div>
              
              <!-- Assign Task button -->
              <div v-if="!isUntracked && userStore.canDo('assign_task') && (!task.is_resource || isCurrentUser)" class="task-item-grid-assign-task-button">
                <ActionButton :icon="getAppIcon('person-plus')" v-tooltip="'Assign Task'" @click="prepAssignTask(index, task, $event)" />
              </div>
              
            </div>
            
          </div>

          <!-- Editing mode -->
          <div v-else class="task-item-grid-slide-container">
            <div class="task-item-grid-meta-row rename-input-grid">
              <RenameInput 
                v-model="editableTaskName"
                :originalValue="task.name || ''"
                placeholder="Task name"
                @confirm="confirmRename"
                @cancel="cancelRename"
              />
            </div>
          </div>
          
          <!-- File state section (absolute positioned, always visible) -->
          <div v-if="!isEditing" class="task-item-grid-file-state-absolute">

            <div v-if="loadingAssetState" class="file-state">
              <ActionButton :isLoading="true" :icon="getAppIcon('loading')"  
                v-tooltip="'Loading...'" />
            </div>

            <div v-else-if="!isUntracked && userStore.canDo('pull_chunk')" class="file-state">
              <ActionButton v-if="task.is_link" :icon="getAppIcon('launch-box')" 
                v-tooltip="'Visit link'" @click="openLink()" />
              <ActionButton v-else-if="task.file_status == 'normal'" :icon="getAppIcon('circle-check-go')" :noFilter="true" 
                v-tooltip="'No changes'"  />
              <ActionButton :icon="getAppIcon('circle-check')" :useAlert="true" :noFilter="true" 
                v-tooltip="'Outdated - Click to update'" v-else-if="task.file_status == 'outdated'" 
                @click="revertTask(index, task, $event)" />
              <ActionButton :icon="getAppIcon('layers-plus')" :useAlert="true" :noFilter="true" 
                v-tooltip="'Modified - Assigned to someone else'" 
                v-else-if="task.file_status == 'modified' && !canModify" @click="canModifyPopUpModal()" />
              <ActionButton :icon="getAppIcon('layers-plus')" :useAlert="true" :noFilter="true" 
                v-tooltip="'Modified - Click to add Checkpoint'" 
                v-else-if="task.file_status == 'modified' && userStore.canDo('create_checkpoint')"
                @click="prepCreateCheckpoint(index, task, $event)" />
              <ActionButton :icon="getAppIcon('jigsaw')" v-tooltip="'File missing - Click to build'"
                v-else-if="task.file_status == 'rebuildable'" @click="revertTask(index, task, $event)" />
              <ActionButton :icon="getAppIcon('alert')" :noFilter="true" 
                v-tooltip="'Task missing - Resync your project'" v-else-if="task.file_status == 'missing'" />
            </div>

            <div v-else-if="isUntracked">
              <ActionButton v-if="userStore.canDo('create_task') || canImport" 
                @click="prepCreateCheckpoint(index, task, $event)" :icon="getAppIcon('layers-plus')" :useDanger="true" 
                :noFilter="true" v-tooltip="'File untracked, click to add.'" />
              <ActionButton v-else :icon="getAppIcon('dot-big')" :useDanger="true" :noFilter="true" 
                v-tooltip="'File untracked'" />
            </div>
          </div>
          
        </div>
      </div>
    </div>
  </div>

  <!-- List View Task Item -->
  <div v-else 
    ref="taskItem" 
    class="task-item-main" 
    v-return="launchSelectedTask" 
    v-esc="handleEscKey" 
    v-stop-propagation
    :style="itemHeightStyles" 
    :class="{
      'task-item-selected': stage.markedItems.includes(task.id) && !isGhost,
      'task-item-cut': stage.cutItems.map((item) => item.id).includes(task.id) && !isGhost,
      'task-item-only-selected': stage.markedItems.length === 1 && stage.firstSelectedItemId === task.id && !isGhost,
      'task-item-last-selected': stage.lastSelectedItemId === task.id && !isGhost,
      'task-item-child': task.parent_id,
      'drop-zone-hovered': isHovered
    }" 
    @dblclick="launchTaskCommand()">

    <div class="task-spacer" v-tooltip="assetTypeName" @click="console.log(task)">
      <span v-if="isUntracked" class="single-action-button single-action-button-disabled">
        <img class="small-icons entity-collapsed" :src="getAppIcon('generic')">
      </span>
      <span v-else class="single-action-button single-action-button-disabled">
        <img class="small-icons entity-collapsed" :src="getAppIcon(task.task_type_icon)">
      </span>
    </div>

    <div class="main-task-item-root">

      <div class="task-item-container drop-zone">

        <!-- Thumbnail preview for list view (optional) -->
        <!-- <div v-if="commonStore.showThumbs && (task.preview || osThumbnail)" class="task-item-preview-container">
          <div class="task-item-preview-image">
            <img class="screenshot-thumb" :src="displayThumbnail">
          </div>
        </div> -->

        <div class="task-item-icon-container">
          <img v-if="task.icon" class="large-icons no-filter" :src="task.icon">
          <img v-else-if="isUntracked" class="large-icons " :src="getAppIcon(getFileTypeIcon(task))" @error="$event.target.src = getAppIcon('file')">
          <span v-else class="app-ext">
          </span>
        </div>

        <div class="task-item-content selection-area">
          <div v-if="!isEditing" class="task-item-details">
            {{ taskName }}
          </div>

          <RenameInput 
            v-else
            v-model="editableTaskName"
            :originalValue="task.name || ''"
            placeholder="Task name"
            @confirm="confirmRename"
            @cancel="cancelRename"
          />

          

          <div v-if="!isEditing && task.is_link" class="weblink-pointer-container">
              <div class="weblink-pointer">
                {{ task.pointer }}  
              </div>
          </div>

        </div>

        <template v-if="!isEditing && !task.is_link">
          <!-- task assignation -->
          <div v-if="!isUntracked && (!task.is_resource || isCurrentUser)" class="task-item-assignee-container">
            <ActionButton class="task-item-assignee-button" v-if="userStore.canDo('assign_task') && !statusMenuDisplayed && !task.assignee_id"
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

          <div v-else-if="!isEditing" class="task-item-assignee-container">
            <ActionButton class="task-item-assignee-button" v-if="userStore.canDo('assign_task') && !statusMenuDisplayed && !task.assignee_id && !isUntracked"
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
          <div v-if="!isEditing && !isUntracked && !statusMenuDisplayed" class="task-item-actions">
            <ActionButton v-if="userStore.canDo('view_checkpoint')" :icon="getAppIcon('layers')" v-tooltip="'Checkpoints'"
              v-stop-propagation @click="viewCheckpoints(index, task, $event)" />

            <div v-if="loadingAssetState" class="file-state">
                <ActionButton :isLoading="true" :icon="getAppIcon('loading')" 
                  v-tooltip="'Loading...'" />
            </div>

            <div v-else-if="userStore.canDo('pull_chunk')" class="file-state">
              <ActionButton :icon="getAppIcon('circle-check-go')" :noFilter="true" @click="handleClick(index, task, $event)"
                v-tooltip="'No changes'" v-if="task.file_status == 'normal'" />
              <ActionButton :icon="getAppIcon('circle-check')" :useAlert="true" :noFilter="true" v-tooltip="'Outdated - Click to update'"
                v-else-if="task.file_status == 'outdated'" @click="revertTask(index, task, $event)" />
              <ActionButton :icon="getAppIcon('layers-plus')" :useAlert="true" :noFilter="true" v-tooltip="'Modified - Assigned to someone else'"
                v-else-if="task.file_status == 'modified' && !canModify" @click="canModifyPopUpModal()" />
              <ActionButton :icon="getAppIcon('layers-plus')" :useAlert="true" :noFilter="true" v-tooltip="'Modified - Click to add Checkpoint'"
                v-else-if="task.file_status == 'modified' && userStore.canDo('create_checkpoint')"
                @click="prepCreateCheckpoint(index, task, $event)" />
              <ActionButton :icon="getAppIcon('jigsaw')" v-tooltip="'File missing - Click to build'"
                v-else-if="task.file_status == 'rebuildable'" @click="revertTask(index, task, $event)" />
              <ActionButton :icon="getAppIcon('alert')" :noFilter="true" v-tooltip="'Task missing - Resync your project'"
                v-else-if="task.file_status == 'missing'" />
            </div>
          </div>
        </template>

        <div v-if="task.is_link" class="task-item-actions link-item-actions" >
          <ActionButton :icon="getAppIcon('launch-box')" v-tooltip="'Visit link'" v-stop-propagation @click="openLink()" />
        </div>

        <div v-else-if="isUntracked" class="task-item-actions">
          <ActionButton v-if="userStore.canDo('create_task') || canImport" @click="prepCreateCheckpoint(index, task, $event)"
            :icon="getAppIcon('layers-plus')" :useDanger="true" :noFilter="true" v-tooltip="'File untracked, click to add.'" />
          <ActionButton v-else :icon="getAppIcon('dot-big')" :useDanger="true" :noFilter="true" v-tooltip="'File untracked'" />
        </div>


      </div>

    </div>
  </div>
</template>

<script setup>

// imports
import { computed, ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue';
import { isValidWeblink } from '@/lib/pointer';
import { AssetService, CheckpointService, FSService, SyncService } from "@/../bindings/clustta/services";
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
import RenameInput from '@/instances/desktop/components/RenameInput.vue'
import StatusMenu from '@/instances/desktop/menus/StatusMenu.vue'
import GridStatusMenu from '@/instances/desktop/menus/GridStatusMenu.vue'
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
  loadingAssetState: { type: Boolean, default: false },
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
const gridStatusMenuVisible = ref(false);

// OS Thumbnail caching
const osThumbnail = ref('');
const thumbnailLoading = ref(false);
const thumbnailCache = new Map();

// Determines the thumbnail to display with priority: user preview > OS thumbnail > task icon > file type icon
const displayThumbnail = computed(() => {
  if (props.task.preview) {
    return props.task.preview;
  }
  
  if (osThumbnail.value) {
    return `data:image/png;base64,${osThumbnail.value}`;
  }
  
  if (props.task.icon) {
    return props.task.icon;
  }
  
  return getAppIcon(getFileTypeIcon(props.task));
});

// Load OS-generated thumbnail for the file with memory caching and async generation fallback
const loadOSThumbnail = async () => {
  const filePath = props.task.file_path;

  const fileExists = await FSService.Exists(filePath)
  if ( !commonStore.useGrid || props.task.preview || !props.task.file_path || !fileExists || thumbnailLoading.value || props.task.is_link) {
    return;
  }

  const cacheKey = filePath;
  
  if (thumbnailCache.has(cacheKey)) {
    osThumbnail.value = thumbnailCache.get(cacheKey);
    return;
  }
  
  thumbnailLoading.value = true;
  
  try {
    const size = 512;
    
    let thumbnail = await FSService.GetCachedOSThumbnail(filePath, size);
    
    if (thumbnail && thumbnail.length > 0) {
      osThumbnail.value = thumbnail;
      thumbnailCache.set(cacheKey, thumbnail);
    } else {
      // Asynchronously generate thumbnail if not cached
      setTimeout(async () => {
        try {
          thumbnail = await FSService.GetOSThumbnail(filePath, size);
          if (thumbnail && thumbnail.length > 0) {
            osThumbnail.value = thumbnail;
            thumbnailCache.set(cacheKey, thumbnail);
          }
        } catch (error) {
          console.debug('Thumbnail generation failed:', error);
        } finally {
          thumbnailLoading.value = false;
        }
      }, 0);
    }
  } catch (error) {
    console.debug('Thumbnail loading failed:', error);
  } finally {
    if (!osThumbnail.value) {
      thumbnailLoading.value = false;
    }
  }
};

const openGridStatusMenu = (event) => {
  console.log('llllllll')
  // event.stopPropagation();
  const id = props.task.id;
  const task = props.task;
  assetStore.selectAsset(task);
  stage.markedTasks = [id];
  gridStatusMenuVisible.value = !gridStatusMenuVisible.value;
};

const closeGridStatusMenu = () => {
  gridStatusMenuVisible.value = false;
};

const handleGridStatusSelected = () => {
  props.task.status = assetStore.selectedAsset.status;
  props.task.status_id = assetStore.selectedAsset.status_id;
  props.task.status_short_name = assetStore.selectedAsset.status_short_name;
  closeGridStatusMenu();
};


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

const assetTypeName = computed(() => {
  return utils.capitalizeStr(props.task?.task_type_name);
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
  await updateAssetName();
  toggleEditMode();
};

const handleEscKey = () => {
  if (isEditing.value) {
    cancelRename();
  }
  if (statusMenuDisplayed.value) {
    statusMenuDisplayed.value = false;
  }
  if (gridStatusMenuVisible.value) {
    gridStatusMenuVisible.value = false;
  }
};

const updateAssetName = async () => {

  isAwaitingResponse.value = true;

  let taskId = props.task.id;
  let task = props.task;

  if (props.task.type === 'task') {

    await AssetService.RenameAsset(projectStore.activeProject.uri, taskId, editableTaskName.value)
      .then((data) => {
        task.name = editableTaskName.value;
        console.log(props.task.file_status)
        emitTaskUpdates(taskId, [
          { property: 'name', value: editableTaskName.value },
          { property: 'file_status', value: 'outdated' },
        ]);

        props.task.file_status = 'outdated'
        
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

        emitTaskUpdates(taskId, [
          { property: 'name', value: editableTaskName.value },
          { property: 'file_path', value: newPath }
        ]);
        
        isAwaitingResponse.value = false;
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        console.error('Error:', error);
      });

  }
}

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
    AssetService.DeleteAsset(projectStore.activeProject.uri, taskId, true)
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
  if (gridStatusMenuVisible.value) {
    gridStatusMenuVisible.value = false;
  }
}, { deep: true });

// watch ( () => props.task, (newItems, oldItems) => {
//   let fileStatus = await assetStore.getAssetFileStatus(task)
// }, { deep: true });


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

  notificationStore.cancleFunction = SyncService.CancelSync
  notificationStore.canCancel = true

  CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [taskId])
    .then(async (response) => {
      emitTaskUpdates(taskId, [
        { property: 'file_status', value: 'normal' }
      ]);
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
    return 'Removed User'
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
  if (gridStatusMenuVisible.value) {
    gridStatusMenuVisible.value = false;
  }
};

onMounted(async () => {
  emitter.on('renameAsset', menuRename);
  document.addEventListener('click', handleClickOutside);
  
  // Load OS thumbnail for this asset
  await loadOSThumbnail();
});

onBeforeUnmount(() => {
  emitter.off('renameAsset', menuRename);
  document.removeEventListener('click', handleClickOutside);
});

// Watch for file path changes (e.g., after rename)
watch(() => props.task.file_path, async (newPath, oldPath) => {
  if (newPath && newPath !== oldPath) {
    osThumbnail.value = '';
    await loadOSThumbnail();
  }
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
  outline-offset: -1px;
  border-radius: var(--large-radius);
  transition: all .2s ease-out;
}

.task-item-main:hover {
  background-color: var(--blue-steel);
  border-radius: var(--small-radius);
}

.task-item-main:hover {
  background-color: var(--steel);
  border-radius: var(--small-radius);
  outline: 1px solid var(--light-steel);
}

.task-item-main:hover  .main-task-item-grid-thumb-container{
  border-radius: var(--tiny-radius);
}

.task-item-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--blue-steel);
}

.task-item-selected:hover {
  background-color: var(--solid-blue-steel);
}

.task-item-cut{
  opacity: .5;
}

.task-item-grid {
  align-items: flex-end;
  padding-left: 0px;
  padding: .5rem;
  background-color: var(--dark-steel);
  outline: var(--transparent-line);
  outline-offset: -1.5px;
}

.task-item-grid-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--blue-steel);
}

.task-item-grid-cut {
  opacity: .5;
}

.task-item-grid-last-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.task-item-grid-only-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.task-item-grid-only-selected:hover {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.main-task-item-grid {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.main-task-item-grid-bottom-bar {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: .2rem;
  padding-top: .5rem;
  box-sizing: border-box;
  overflow: visible;
  position: relative;
  flex-shrink: 0;
  transition: all 0.2s ease-out;
  /* background-color: forestgreen; */
}


.task-item-grid-bottom-bar-wrapper {
  position: relative;
  display: flex;
  flex-direction: column;
  width: 100%;
  box-sizing: border-box;
  gap: .3rem;
  height: 32px;
  transition: all 0.2s ease-out;
  /* background-color: crimson; */
}

/* Expand height on hover, but not in edit mode */
.task-item-grid:hover .task-item-grid-bottom-bar-wrapper:not(:has(.rename-input-grid)) {
  height: 70px;
  /* background-color: black; */
}

.task-item-grid-slide-container {
  display: flex;
  flex-direction: column;
  width: 80%;
  box-sizing: border-box;
  gap: .3rem;
  flex: 1;
  justify-content: space-between;
  transition: all 0.2s ease-in-out;
  /* background-color: royalblue; */
}

.task-item-grid-slide-container:has(.rename-input-grid) {
  width: 100%;
}

.task-item-grid:hover .task-item-grid-slide-container {
  width: 100%;
  /* transition: all 0.2s ease-out; */
}

.task-item-grid-meta-row {
  display: flex;
  align-items: center;
  gap: .3rem;
  width: 100%;
  min-height: 32px;
  height: 32px;
  flex-shrink: 0;
  box-sizing: border-box;
  overflow: hidden;
  /* background-color: darkslateblue; */
}

.task-item-grid-actions-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  width: 100%;
  min-height: 32px;
  height: 32px;
  flex-shrink: 0;
  box-sizing: border-box;
  overflow: hidden;
  width: 80%;
  /* background-color: hotpink; */
}

.task-item-grid:hover .task-item-grid-actions-row {
  display: flex;
}

.task-item-grid-type-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.task-item-grid-file-state-absolute {
  position: absolute;
  bottom: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 32px;
  z-index: 10;
}

.main-task-item-grid-thumb-container {
  position: relative;
  display: flex;
  overflow: hidden;
  height: 100%;
  width: 100%;
  background-color: rgba(0, 0, 0, 0.2);
  border-radius: var(--normal-radius);
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease-out;
  flex: 1; 
}

.main-task-item-grid-meta {
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

.task-item-last-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.task-item-only-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.task-item-only-selected:hover {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
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
  position: relative;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  /* padding: .1rem; */
  overflow: hidden;
  min-width: 60px;
  height: 100%;
  width: 100%;
}

.task-item-preview-image img,
.task-item-icon-container img {
  transition: opacity 0.2s ease-in-out;
}

.screenshot-thumb{
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.task-item-preview-image img[src*="data:image"],
.task-item-icon-container img[src*="data:image"] {
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.task-item-preview-image {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  height: 100%;
  width: 100%;
  /* background-color: forestgreen; */
  /* aspect-ratio: 16 / 9; */
  /* background-color: var(--light-steel); */
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

.task-item-icon-overlay {
  position: absolute;
  /* background-color: forestgreen; */
  bottom: 0;
  left: 0;
  width: 32px;
  height: 32px;
  /* background-color: rgba(0, 0, 0, 0.7); */
  border-radius: 6px;
  /* padding: 4px; */
  min-width: unset;
}

.overlay-icons{
  width: 100%;
  height: 100%;
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
  /* direction: rtl; */
  text-align: left;
  font-size: 14px;
}

.weblink-pointer-container {
  /* width: 100%; */
  /* width: max-content; */
  display: none;
  flex-wrap: nowrap;
  text-wrap: nowrap;
  justify-content: flex-end;
  /* background-color: crimson; */
  /* width: min-content; */
  flex: 1;
  color: var(--white);
  padding: .2rem .2rem;
  border-radius: var(--tiny-radius);
  font-size: 12px;
  overflow: hidden;
  /* outline: var(--transparent-line); */
  /* background-color: var(--black-steel); */
  /* display: flex; */
  align-items: center;
  justify-content: flex-start;
  height: max-content;
  box-sizing: border-box;
  overflow: hidden;
  opacity: .5;
}

.weblink-pointer {
  /* display: flex; */
  width: 100%;
  overflow: hidden;
  /* padding: 0 .4rem; */
  box-sizing: border-box;
  align-items: flex-start;
  height: 100%;
  font-weight: 300;
  text-overflow: ellipsis;
  /* background-color: forestgreen; */
}

.task-item-main:hover .weblink-pointer-container {
  /* text-decoration: underline; */
  display: flex;
}

.task-item-assignee-button {
  display: none;
}

.task-item-main:hover .task-item-assignee-button {
  /* text-decoration: underline; */
  display: flex;
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
  border-radius: 6px;
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
  /* min-width: 150px; */
  min-width: var(--actions-width);
}

.link-item-actions{
  min-width: max-content;
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

.single-action-button{
  align-content: center;
  justify-content: center;
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

.task-item-grid-status-display {
  display: flex;
  align-items: center;
  justify-content: center;
}

.task-item-status-grid {
  display: flex;
  border-radius: var(--normal-radius);
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  min-width: 60px;
  padding: .4rem .4rem;
  height: max-content;
  font-size: 12px;
  text-transform: uppercase;
  font-weight: 700;
  color: black;
  transition: all 0.2s ease-out;
}

.task-item-status-grid:hover {
  border-radius: 6px;
}

.task-item-grid-untracked-label {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  flex: 1;
  padding: 0 .5rem ;
  /* background-color: forestgreen; */
}

.task-item-grid-untracked-label span {
  font-style: italic;
  font-size: 14px;
  color: var(--white);
  opacity: 0.7;
}

.task-item-grid-checkpoints-button,
.task-item-grid-assign-task-button {
  display: flex;
  align-items: center;
  justify-content: center;
}

.task-item-grid-assignee {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}



.task-item-grid-assignee-overlay-top-right {
  position: absolute;
  bottom: 8px;
  right: 8px;
  z-index: 10;
}

.task-item-grid-assign-button {
  opacity: 0;
  transition: opacity 0.2s ease-in-out;
}

.task-item-grid:hover .task-item-grid-assign-button {
  opacity: 1;
}

.profile-picture-grid {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border: 2px solid rgba(255, 255, 255, 0.8);
}

.profile-img-grid {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
}

.profile-picture-grid-small {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border: 1px solid rgba(255, 255, 255, 0.6);
}

.profile-img-grid-small {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
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

@keyframes loadingRotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.loading-asset-state-icon {
  width: 20px;
  height: 20px;
  overflow: hidden;
  animation: loadingRotate .5s linear infinite;
}

</style>







