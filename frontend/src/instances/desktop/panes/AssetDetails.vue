<template>
  <div class="general-pane-header">
    <HeaderArea :title="utils.capitalizeStr(selectedTaskName)" :icon="selectedTaskIcon" :useIconBlob="true" />
    <ActionButton v-if="userStore.canDo('update_task')" :icon="getAppIcon('parameters')" :showLabel="false"
      v-tooltip="'Edit Task'" :buttonFunction="editTask" />
  </div>

  <div class="general-pane-root">
    <div class="general-pane-container">

      <div v-if="taskStore.selectedTask?.preview" class="entity-thumb-container">
        <div class="entity-thumb">
          <img v-if="taskStore.selectedTask.preview" class="screenshot-thumb" :src="taskStore.selectedTask.preview">
          <img v-else class="screenshot-thumb" src="/page-states/no_image.png">
        </div>
      </div>

      <div class="pane-parameter-section">
        <div class="action-bar" v-if="userStore.canDo('update_task')">

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('brush-plus')" :label="'Type'" />
            <DropDownBox :items="taskStore.getTaskTypesNames" :selectedItem="taskStore.selectedTask.task_type_name"
              :onSelect="changeTaskType" :fixedWidth="true" />
          </div>

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('clock')" :label="'Status'" />
            <DropDownBox :items="projectStatuses" :selectedItem="taskStore.selectedTask.status_short_name"
              :onSelect="setStatus" :fixedWidth="true" />
          </div>

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('shapes')" :label="'Task'" />

            <ToggleSwitch v-tooltip="!taskStore.selectedTask.is_resource ? 'Unset as Task' : 'Set as Task'"
              @click="toggleIsTask" :switchValueProp="!taskStore.selectedTask.is_resource" />
          </div>

        </div>

        <span v-if="userStore.canDo('update_task')" class="menu-divider"></span>

        <div class="task-details">
          <div class="pane-parameter-detail">
            <div class="simple-text-key">
              Parent
            </div>
            <div class="simple-text-value">
              {{ taskStore.selectedTask.entity_name }}
            </div>
          </div>

          <div v-if="!taskStore.selectedTask.is_link" class="pane-parameter-detail">
            <div class="simple-text-key">
              Extension
            </div>
            <div class="simple-text-value">
              {{ taskStore.selectedTask.extension }}
            </div>
          </div>

          <div class="pane-parameter-detail">
            <div class="simple-text-key">
              Assigned to
            </div>
            <ActionButton v-if="taskStore.selectedTask.assignee_id" :iconAfter="true" :label="userFullName" v-tooltip="'See all tasks'" :buttonFunction="showAllTasks"/>
            <div v-else class="simple-text-value">
              {{ userFullName }}
            </div>
          </div>

          <div v-if="!taskStore.selectedTask.is_link" class="pane-parameter-detail">
            <div class="simple-text-key">
              Checkpoint Comment
            </div>
            <div class="simple-text-value">
              {{ lastCheckpoint.comment }}
            </div>
          </div>

          <div v-if="lastCheckpoint?.comment !== 'No checkpoints' && !taskStore.selectedTask.is_link"
            class="pane-parameter-detail">
            <div class="simple-text-key">
              Checkpoint Date
            </div>
            <div class="simple-text-value">
              {{ formatMtime(lastCheckpoint.created_at) }}
            </div>
          </div>
          

          <div class="pane-parameter-detail">
            <div class="simple-text-key">
              Location
            </div>
              <div class="simple-text-value">
                {{ taskStore.selectedTask.file_path }}
              </div>
              <div class="pane-parameter-actions">
                <ActionButton :icon="getAppIcon('copy')" v-tooltip="'Copy Path'" @click="copyTaskPath('task')"/>
                <ActionButton :icon="getAppIcon('folder-arrow-up-right')" v-tooltip="'Reveal in Explorer'" :buttonFunction="revealInExplorer"/>
              </div>
          </div>

          <div v-if="!taskStore.selectedTask.is_link" class="pane-parameter-detail">
            <div class="simple-text-key">
              File State
            </div>
            <div class="simple-text-value">
              {{ taskStore.selectedTask.file_status }}
            </div>
          </div>

          <div class="pane-parameter-detail">
          <div class="simple-text-key">
          Size 
          </div>
          <div class="simple-text-value">
            {{  assetSize }}
          </div>
        </div>

          <div v-if="taskStore.selectedTask.tags.length" class="pane-parameter-detail">
            <div class="simple-text-key">
              Tags
            </div>
          </div>
        </div>

      </div>

      <div class="pane-parameter-section">
        <TagContainer :tags="taskStore.selectedTask.tags" :displayOnly="true" />
      </div>

    </div>
  </div>



</template>

<script setup>
// imports
import { ref, computed, onMounted, watch, onBeforeUnmount } from 'vue';
import { ClipboardService, FSService } from '@/../bindings/clustta/services/index';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// store imports
import { useProjectStore } from '@/stores/projects';
import { useUserStore } from '@/stores/users';
import { useStageStore } from '@/stores/stages';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useTaskStore } from '@/stores/task';
import { useStatusStore } from '@/stores/status';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useCommonStore } from '@/stores/common';

// services
import { TaskService } from "@/../bindings/clustta/services";

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import TagContainer from '@/instances/common/components/TagContainer.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import AssigneeItem from '@/instances/common/components/AssigneeItem.vue';

// stores
const taskStore = useTaskStore();
const stage = useStageStore();
const userStore = useUserStore();
const modals = useDesktopModalStore();
const statusStore = useStatusStore();
const projectStore = useProjectStore();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const commonStore = useCommonStore();

// refs
const numberOfSelectedTasks = ref(0);
const multiStatusChange = ref(false);

// computed properties
const projectStatuses = computed(() => {
  const allStatuses = statusStore.statuses;
  if (!userStore.canDo('set_done_task')) {
    const limitedStatus = ['done', 'retake']
    return allStatuses.filter((item) => !limitedStatus.includes(item.short_name))
  } else {
    return allStatuses.map((status) => status.short_name.toUpperCase())
  }
});

const singleTask = computed(() => {
  numberOfSelectedTasks.value = stage.markedTasks.length;
  const isSingleTask = stage.markedTasks.length <= 1 && taskStore.selectedTask;
  return isSingleTask
});

const selectedTaskName = computed(() => {
  if (taskStore.selectedTask) {
    return singleTask.value ? taskStore.selectedTask.name : 'Multiple tasks selected'
  }
});

const selectedTaskIcon = computed(() => {
  if (taskStore.selectedTask) {
    return singleTask.value ? taskStore.selectedTask.icon : '/icons/categories.svg'
  }
});

//methods
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const emitTaskUpdates = (taskId, updates) => {
  const updateData = { itemId: taskId, updates };
  
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

const copyTaskPath = async (pathType) => {
  let task = taskStore.selectedTask;
  let taskPath = task.file_path;
  taskPath = taskPath.replace(/\\/g, '/');
  let taskDir = taskPath.split('/').slice(0, -1).join('/');
  let resourcesFolder = taskDir + '/resources';
  let outputPath = taskDir + '/output';
  if (pathType === 'resources') {
    taskPath = resourcesFolder;
  } else if (pathType === 'output') {
    taskPath = outputPath;
  }
  await ClipboardService.WriteText(taskPath);
  const message = 'Path copied to clipboard';
  notificationStore.addNotification(message, "", "success");
};

const showAllTasks = () => {
  commonStore.activeWorkspace = 'Default'
  commonStore.resetFilters();
  commonStore.navigatorMode = false;
  if (taskStore.selectedTask?.assignee_id) {
    const assignee = userStore.getUserData(taskStore.selectedTask.assignee_id);
    if (assignee) {
      const assigneeFilter = {
        name: `${assignee.first_name} ${assignee.last_name}`,
        id: assignee.id,
        type: 'assignation',
        avatarColor: userStore.userProfileColor(assignee.id)
      };
      commonStore.taskFilters.push(assigneeFilter);
    }
  }
  commonStore.onlyAssets = true;
  emitter.emit('refresh-browser');
};

const revealInExplorer = async () => {
  const taskId = taskStore.selectedTask.id;
  if(taskStore.selectedTask.file_status == "rebuildable"){
    await CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [taskId])
    .then( async (response) => {
      taskStore.rebuildableTasksPath = taskStore.rebuildableTasksPath.filter(taskPath => taskPath !== task.task_path)
      taskStore.outdatedTasksPath = taskStore.outdatedTasksPath.filter(taskPath => taskPath !== task.task_path);
      emitter.emit('get-project-data')
    })
    .catch((error) => {
      notificationStore.errorNotification("Error downloading Task", error);
      console.error(error);
    });

  } 
  TaskService.RevealTask(projectStore.activeProject.uri, taskStore.selectedTask.id);
};

const toggleIsTask = async () => {
  stage.operationActive = true;
  const projectPath = projectStore.activeProject.uri;
  let isTask = taskStore.selectedTask.is_resource;
  let task = taskStore.selectedTask;
    
  await TaskService.ToggleIsTask(projectPath, task.id,  isTask)
    .then((data) => {

      taskStore.selectedTask.is_resource = !isTask;
      emitTaskUpdates(task.id, [
        { property: 'is_resource', value: !isTask }
      ]);
      
    })
    .catch((error) => {
      console.error('Error:', error);
    });
    
    stage.operationActive = false;
};

const changeTaskType = async (taskTypeName) => {
  stage.operationActive = true;

  let newTaskType;
  const taskTypes = taskStore.getTaskTypes;
  newTaskType = taskTypes.find((item) => item.name === taskTypeName);

  const projectPath = projectStore.activeProject.uri;
  let task = taskStore.selectedTask;

  await TaskService.UpdateTask(projectPath, task.id, task.name, newTaskType.id, task.is_resource, '', task.tags)
    .then((data) => {
      task.task_type_name = newTaskType.name;
      task.task_type_icon = newTaskType.icon;
      task.task_type_id = newTaskType.id;
      
      emitTaskUpdates(task.id, [
        { property: 'task_type_name', value: newTaskType.name },
        { property: 'task_type_icon', value: newTaskType.icon },
        { property: 'task_type_id', value: newTaskType.id }
      ]);
      
    })
    .catch((error) => {
      console.error('Error:', error);
    });

  stage.operationActive = false;

};

const setStatus = async (statusName) => {
  stage.operationActive = true;
  const projectPath = projectStore.activeProject.uri;
  const status = statusStore.statuses.find(item => item.short_name === statusName.toLowerCase());
  let task = taskStore.selectedTask;
  
  await TaskService.ChangeStatus(projectPath, task.id, status.id)
    .then((data) => {
      task.status_short_name = status.short_name;
      task.status = status;
      
      emitTaskUpdates(task.id, [
        { property: 'status_short_name', value: status.short_name },
        { property: 'status', value: status }
      ]);
      
    })
    .catch((error) => {
      console.error('Error:', error);
    });
    
  multiStatusChange.value = false;
  stage.operationActive = false;
};

const userFullName = computed(() => {
  let user = userStore.getUserData(taskStore.selectedTask.assignee_id);
  if (taskStore.selectedTask.assignee_id) {
    let fullname = `${user.first_name} ${user.last_name}`;
    return fullname
  } else {
    return 'Nobody'
  }
});

const lastCheckpoint = computed(() => {
  let task = taskStore.selectedTask;
  if (task.is_link) return ''
  let checkpoint = task.checkpoints[0];
  return checkpoint ? { comment: checkpoint.comment, created_at: checkpoint.created_at } : { comment: 'No checkpoints', created_at: task.created_at };
});

const formatMtime = (mtime) => {
  const date = new Date(mtime);
  const day = date.getDate();
  const monthNames = ["Jan", "Feb", "Mar", "Apr", "May", "Jun",
    "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
  const month = monthNames[date.getMonth()];
  const year = date.getFullYear();

  return `${day} ${month} ${year}`;
};

const editTask = () => {
  modals.setModalVisibility('editAssetModal', true);
};

const assetSize = ref(0);

const assetPath = computed(() => {
  const path = taskStore.selectedTask?.file_path;
  return path?.replace(/\\/g, '/')
});

const getAssetSize = async() => {
  const size = await FSService.FileStat(assetPath.value);
  assetSize.value = size.formattedSize;
}

const getProjectData = async () => {
  if (!await FSService.Exists(assetPath.value)){
    assetSize.value = 'Not on disk'
    return
  }
  getAssetSize();
}

watch(() => taskStore.selectedTask, () => {
  assetSize.value = 0;
  getProjectData();
});


// onMounted
onMounted(() => {
  stage.markedTasks = [];
  
  getProjectData();
	emitter.on('get-project-data', getProjectData);
});

onBeforeUnmount(() => {
	emitter.off('get-project-data', getProjectData);
});


</script>


<style scoped>
@import "@/assets/desktop.css";
.task-details{
  overflow: hidden;
  overflow-y: scroll;
  padding-right: .5rem;
}

.task-details::-webkit-scrollbar {
  width: 4px;
}

.task-details::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--light-steel);
}

.task-details::-webkit-scrollbar-track {
  border-radius: 10px;
}
.pane-parameter-detail {
  display: flex;
  font-size: 14px;
  height: max-content;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  height: 30px;
  border-bottom: var(--transparent-line);
}

.menu-divider {
  height: 5px;
  margin-top: 10px;
  /* margin-bottom: 10px; */
  width: 100%
}

.simple-text-key {
  white-space: nowrap;
}

.status-box-container {
  width: 50%;
  height: 100%;
  height: max-content;
  /* height: 60px; */
}

.input-short {
  flex: 1;
  width: 100%;
}

.listbox-short {

  flex: 1;
  width: 130px;
}

.input-label {

  font-family: Inter, sans-serif;
  color: white;
  font-size: 16px;
  white-space: nowrap;
  flex: 1;

}

.pop-up-prompt {
  gap: 10px;
  /* background-color: bisque; */
  align-items: center;
  justify-content: center;
  /* height: 400px; */
  /* background-color: darkseagreen; */
}

.action-bar {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: .6rem;
  width: max-content;
  width: 100%;
  /* justify-content: space-around; */
  height: max-content;
  padding: .2rem;
  /* background-color: black; */
  /* background-color: tomato; */
  align-items: flex-start;
  box-sizing: border-box;
}

.action-bar-section {
  display: flex;
  align-items: center;
  gap: .5rem;
  justify-content: space-between;
  width: 100%;
}

.multi-assign {
  /* background-color: tomato; */
  width: 100%;
  display: flex;
  align-items: flex-start;
  flex-direction: column;
}
</style>