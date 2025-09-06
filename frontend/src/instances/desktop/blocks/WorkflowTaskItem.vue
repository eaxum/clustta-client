<template>
  <div ref="taskItem" class="task-item-main" v-stop-propagation :class="{
    'task-item-selected': stage.markedItems.includes(task.id),
    'task-item-only-selected': stage.markedItems.length === 1 && stage.firstSelectedItemId === task.id,
    'task-item-last-selected': stage.lastSelectedItemId === task.id,
    'task-item-child': task.parent_id,
    'is-intersecting': dndStore.intersectingItemIds.includes(task.id)
  }" @dblclick="launchTaskCommand(index, task, $event)">
    <!-- @click="handleClick(index, task, $event)" -->
    <div class="task-spacer" v-tooltip="taskTypeName" @click="console.log(task)">
      <span class="single-action-button single-action-button-disabled">
        <img class="small-icons entity-collapsed" :src="getAppIcon(task.task_type_icon)">
      </span>
    </div>

    <div class="main-task-item-root">

      <div class="task-item-container drop-zone">

        <div v-if="commonStore.showThumbs" class="task-item-preview-container">
          <div class="task-item-preview-image">
            <img v-if="task.preview" class="screenshot-thumb" :src="task.preview">
            <img v-else class="screenshot-thumb" src='/page-states/no_image.png'>
          </div>
        </div>

        <div class="task-item-icon-container">
          <img v-if="task.icon" class="large-icons" :src="task.icon">
          <span v-else class="app-ext">
          </span>
        </div>

        <div class="task-item-content selection-area">
          <div class="task-item-details">
            {{ utils.capitalizeStr(taskName) }}
            <!-- {{ dndStore.intersectingItemIds.includes(task.id) }} -->
            <!-- {{ task.is_resource }} -->
          </div>
        </div>

        <!-- <div class="task-item-meta-container">
          <div class="task-item-meta">
            {{ taskName }}
          </div>
        </div> -->

        <!-- task assignation -->
        <div v-if="!isUntracked && (!task.is_resource || isCurrentUser)" class="task-item-assignee-container">
          <ActionButton v-if="userStore.canDo('assign_task') && !statusMenuDisplayed && !task.assignee_id"
            :icon="getAppIcon('person-plus')" v-tooltip="'Assign Task'" @click="prepAssignTask(index, task, $event)" />

          <div v-else-if="task.assignee_id" @click="prepAssignTask(index, task, $event)" v-stop-propagation
            class="task-item-assignee">
            <span v-tooltip="userFullName" class="single-action-button">
              <div class="profile-picture" :style="{ backgroundColor: profileColor(task.assignee_id) }">
                <img class="profile-img" :src="getAppIcon('person')">
              </div>
            </span>
          </div>
        </div>

        <div v-else class="task-item-assignee-container">
        </div>

        <!-- task status -->
        <div v-if="!isUntracked && (!task.is_resource || isCurrentUser)" class="task-item-status-root">
          <StatusMenu @statusSelected="closeStatusMenu" v-if="statusMenuDisplayed" />

          <div v-else class="task-item-status-container" v-stop-propagation
            @click="toggleDisplayStatusMenu(index, task, $event)">
            <div class="task-item-status" :style="{ backgroundColor: task.status.color }">
              {{ task.status.short_name }}
            </div>
          </div>
        </div>

        <div v-else-if="!isUntracked" class="task-item-status-root">

          <div class="task-item-status-container" v-stop-propagation>
            <div class="task-item-status" :style="{ backgroundColor: task.status.color }">
              <!-- {{ task.status.short_name }} -->
            </div>
          </div>
        </div>

        <!-- task actions -->
        <div v-if="!isUntracked && !statusMenuDisplayed" class="task-item-actions">
          <ActionButton v-if="userStore.canDo('view_checkpoint')" :icon="getAppIcon('layers')" v-tooltip="'Checkpoints'"
            v-stop-propagation @click="viewCheckpoints(index, task, $event)" />
          <div v-if="userStore.canDo('pull_chunk')" class="file-state">
            <ActionButton :icon="getAppIcon('circle-check-go')" @click="handleClick(index, task, $event)"
              v-tooltip="'No changes'" v-if="task.file_status == 'normal'" />
            <ActionButton :icon="getAppIcon('circle-check-alert')" v-tooltip="'Outdated - Click to update'"
              v-else-if="task.file_status == 'outdated'" @click="revertTask(index, task, $event)" />
            <ActionButton :icon="getAppIcon('layers-plus-alert')" v-tooltip="'Modified - Click to add Checkpoint'"
              v-else-if="task.file_status == 'modified' && userStore.canDo('create_checkpoint')"
              @click="prepCreateCheckpoint(index, task, $event)" />
            <ActionButton :icon="getAppIcon('jigsaw')" v-tooltip="'File missing - Click to rebuild'"
              v-else-if="task.file_status == 'rebuildable'" @click="revertTask(index, task, $event)" />
            <ActionButton :icon="getAppIcon('alert')" v-tooltip="'Task missing - Resync your project'"
              v-else-if="task.file_status == 'missing'" />
          </div>
        </div>

        <ActionButton v-if="isUntracked" :icon="getAppIcon('plus-circle')" v-tooltip="'Add File'"
          @click="importItem(index, task, $event)" />
      </div>

    </div>
  </div>
</template>

<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};
// imports
import { computed, ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue';
import { isValidWeblink } from '@/lib/pointer';
import { TaskService, CheckpointService, FSService } from "@/../bindings/clustta/services";
import utils from '@/services/utils';

// states/store imports
import { useTrayStates } from '@/stores/TrayStates';
import { useTaskStore } from '@/stores/task';
import { useMenu } from '@/stores/menu';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { useUserStore } from '@/stores/users';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useCommonStore } from '@/stores/common';
import { useEntityStore } from '@/stores/entity';
import { useProjectStore } from '@/stores/projects';
import { useDndStore } from '@/stores/dnd';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import StatusMenu from '@/instances/desktop/menus/StatusMenu.vue'
import { Browser } from "@wailsio/runtime";

// states/stores
const userStore = useUserStore();
const trayStates = useTrayStates();
const menu = useMenu();
const panes = usePaneStore();
const stage = useStageStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const taskStore = useTaskStore();
const commonStore = useCommonStore();
const entityStore = useEntityStore();
const projectStore = useProjectStore();
const dndStore = useDndStore();

// emits
const emit = defineEmits(['update:height', 'expand']);

// props
const props = defineProps({
  task: Object,
  entityId: { type: String, default: '' },
  index: Number,
  isChild: { type: Boolean, default: false },
  isUntracked: { type: Boolean, default: false },
});

// refs
const taskItem = ref(null);
const isExpanded = ref(false);
const displayStatusMenu = ref(false);
const taskTypeName = ref('');

const importItem = (index, task, event) => {
  handleClick(index, task, event);
  const inRoot = props.task.entity_path === ""
  const taskPath = props.task.file_path;
  let parentId = ""
  let untrackedParents = []
  let parentPaths = utils.getParentPaths(props.task.entity_path)
  if (!inRoot) {
    for (let parent of parentPaths) {
      parentId = entityStore.entities.find((item) => item.entity_path === parent)?.id;
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

// Update the item's height and emit the new height to the virtual list
const updateHeight = () => {
  nextTick(() => {
    emit('update:height', taskItem.value.offsetHeight)
  })
}

// Toggle the accordion expansion
const toggle = () => {
  isExpanded.value = !isExpanded.value
  emit('expand', props.item)
  updateHeight()
}

// Watch for changes in expansion and adjust height
watch(isExpanded, updateHeight)

// computed properties
const taskName = computed(() => {
  const task = props.task;
  const taskName = task.name;
  const isDirectParent = props.entityId === task.entity_id;
  const taskPath = task.task_path.replace(/\//g, ' / ');

  if (commonStore.showFullPath) {
    return taskPath
  }
  if (props.isChild) {
    if (commonStore.showChildEntities) {
      return taskName
    } else {
      return isDirectParent ? taskName : taskPath
    }
  } else {
    if (commonStore.viewSearchQuery) {
      return taskPath
    } else {
      return taskName
    }
  }
});

const taskTypeIcon = computed(() => {
  const icon = '/types-icons/' + props.task.task_type_icon + '.svg';
  if (icon) {
    return icon
  } else {
    return '/types-icons/other.svg';
  }
});

const statusMenuDisplayed = computed(() => { return displayStatusMenu.value && isSelectedTask.value });
const isSelectedTask = computed((taskId) => {
  if (!taskStore.selectedTask) {
    return
  }
  return taskStore.selectedTask.id === props.task.id
});

// methods
const closeStatusMenu = () => {
  displayStatusMenu.value = false;
};

const launchTaskCommand = async (index, task, event) => {
  if (!userStore.canDo('pull_chunk')) {
    return
  }
  // handleClick(index, task, event);
  if (task.is_link && isValidWeblink(task.pointer)) {
    Browser.OpenURL(task.pointer)
  } else {
    let file_path = task.pointer ? task.pointer : task.file_path
    if (await FSService.Exists(file_path)) {
      FSService.LaunchFile(file_path)
    } else {
      notificationStore.addNotification("File Not On Disk, Rebuild", "File not found on disk, rebuild task", "error")
    }
  }
};

const prepCreateCheckpoint = (index, task, event) => {
  handleClick(index, task, event);
  taskStore.selectedTask = task;
  modals.setModalVisibility('createCheckpointModal', true);
};

const revertTask = async (index, task, event) => {
  handleClick(index, task, event);
  const taskId = task.id;
  CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [taskId])
    .then(async (response) => {
      let fileStatus = await taskStore.getTaskFileStatus(task)
      props.task.file_status = fileStatus
    })
    .catch((error) => {
      console.log(error)
      notificationStore.errorNotification("Error Reverting Task", error)
    });
};

const editParams = () => {
  console.log('escape')
};

const toggleFlyout = (event, task) => {
  // if(userStore.canDo('change_status')){
  taskStore.isTaskStatus = true;
  taskStore.selectTask(task);
  if (event) {
    const top = event.target.getBoundingClientRect().top;
    const right = event.target.getBoundingClientRect().right;
    const width = event.target.offsetWidth + 20;
    menu.flyoutTop = top;
    menu.flyoutLeft = right;
    console.log(right)
    menu.flyoutWidth = width;
    trayStates.flyoutTop = top;
    trayStates.flyoutLeft = 0;
  };

  menu.showStatusOptions = true;
  // }
};

const userFullName = computed(() => {
  let user = userStore.getUserData(props.task.assignee_id);
  if (!user) {
    return 'No User'
  } else {
    return `${user.first_name} ${user.last_name}`;
  }
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
  // menu.contextMenuVisible = false;
  closeStatusMenu();
  const id = task.id
  // taskStore.selectTask(task);
  // const id = task.id;


  if (stage.firstSelectedItemId !== id) {
    stage.markedItems = [id];
    stage.firstSelectedItemId = id;
    taskStore.selectTask(task);
    panes.setPaneVisibility('assetDetails', true);
    console.log('1')
  } else if (stage.markedItems.length && stage.firstSelectedItemId === id) {

    stage.markedItems = [id];
    taskStore.selectTask(task);
    console.log('2')

  } else {
    stage.markedItems = [];
    stage.firstSelectedItemId = '';
    console.log('3')
  }

};

const toggleDisplayStatusMenu = (index, task, event) => {
  handleClick(index, task, event);
  if (!userStore.canDo('change_status')) {
    return
  }
  taskStore.isTaskStatus = true;
  taskStore.selectTask(task);
  displayStatusMenu.value = true;
};

const goToDependencies = (index, task, event) => {
  handleClick(index, task, event);
  taskStore.selectTask(task);
  stage.setStageVisibility('dependencies', true);
};

const viewCheckpoints = (index, task, event) => {
  // stage.handleClick(event, task);
  // handleClick(index, task, event);
  stage.markedItems = [task.id];
  taskStore.selectTask(task);
  panes.setPaneVisibility('checkpoints', true);
  panes.showDetailsPane = true;
};

const prepAssignTask = (index, task, event) => {
  if (!userStore.canDo('assign_task')) {
    return
  }
  // if (!isCurrentUser.value) {
  //   return
  // }
  handleClick(index, task, event);


  const id = task.id;
  taskStore.selectTask(task);
  stage.markedTasks = [id];
  menu.showContextMenu(event, 'assignMenu', true);
};

const handleClickOutside = (event) => {
  if (displayStatusMenu.value) {
    displayStatusMenu.value = false;
  }
};

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
});

onBeforeUnmount(() => {
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
  display: flex;
  gap: .2rem;
  color: white;
  align-items: center;
  padding-left: .5rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  align-items: flex-start;
  background-color: var(--dark-steel);
  border-radius: 10px;
  overflow: hidden;
  padding-right: 0px;

  outline: var(--transparent-line);
  outline-offset: -1px;

}

.task-item-main:hover {
  outline: var(--transparent-line);
  outline: 1px solid rgb(255, 255, 255);
  outline-offset: -1.5px;
}

.task-item-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--blue-steel);
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

.main-task-item-root {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: white;
  align-items: center;
  padding: .3rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  /* background-color: blue; */
  border-radius: 10px;
  overflow: hidden;
  padding-right: 0px;

}

.task-item-container {
  display: flex;
  gap: .5rem;
  color: white;
  align-items: center;
  padding: .4rem;
  box-sizing: border-box;
  width: 100%;
  height: 50px;
  justify-content: space-between;
  transition: all .3s ease-out;
}

.task-spacer {
  /* display: flex;
  align-items: center;
  justify-content: center;
  max-width: 40px;
  min-width: 40px;
  aspect-ratio: 1/1;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box; */
  position: relative;
  width: min-content;
  width: 25px;
  height: 60px;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  /* background-color: purple; */
  /* flex: 1; */
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
  /* background-color: indigo; */
  width: 100%;
  overflow: hidden;
}

.task-item-meta-container {
  background-color: blue;
  width: 100%;
  display: flex;
  justify-content: flex-end;
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

.task-item-details {
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
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
  /* background-color: darkcyan; */
  min-width: 150px;
  min-width: var(--actions-width);

  /* flex: 1; */
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