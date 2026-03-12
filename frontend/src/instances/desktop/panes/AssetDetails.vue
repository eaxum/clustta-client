<template>
  
  <div class="general-pane-root">
    <div class="general-pane-container">

      <div v-if="assetStore.selectedAsset?.preview" class="entity-thumb-container">
        <div class="entity-thumb">
          <img v-if="assetStore.selectedAsset.preview" class="screenshot-thumb" :src="assetStore.selectedAsset.preview">
          <img v-else class="screenshot-thumb" src="/page-states/no_image.png">
        </div>
      </div>

      <div class="pane-parameter-section">
        <div class="action-bar" v-if="userStore.canDo('update_task')">

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('file-plus')" :label="$t('panes.type')" />
            <DropDownBox :items="assetStore.getAssetTypesNames" :selectedItem="assetStore.selectedAsset?.task_type_name"
              :onSelect="changeTaskType" :fixedWidth="true" />
          </div>

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('clock')" :label="$t('panes.status')" />
            <DropDownBox :items="projectStatuses" :selectedItem="assetStore.selectedAsset.status_short_name"
              :onSelect="setStatus" :fixedWidth="true" />
          </div>

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('shapes')" :label="$t('panes.task')" />

            <ToggleSwitch v-tooltip="!assetStore.selectedAsset.is_resource ? $t('panes.unsetAsTask') : $t('panes.setAsTask')"
              @click="toggleIsTask" :switchValueProp="!assetStore.selectedAsset.is_resource" />
          </div>

        </div>

        <span v-if="userStore.canDo('update_task')" class="menu-divider"></span>

        <div class="task-details">
          <div class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.parent') }}
            </div>
            <div class="simple-text-value">
              {{ assetStore.selectedAsset.entity_name }}
            </div>
          </div>

          <div v-if="!assetStore.selectedAsset.is_link" class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.extension') }}
            </div>
            <div class="simple-text-value">
              {{ assetStore.selectedAsset.extension }}
            </div>
          </div>

          <div class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.assignedTo') }}
            </div>
            <ActionButton v-if="assetStore.selectedAsset.assignee_id" :iconAfter="true" :label="userFullName" v-tooltip="$t('panes.seeAllTasks')" :buttonFunction="showAllTasks"/>
            <div v-else class="simple-text-value">
              {{ userFullName }}
            </div>
          </div>

          <div v-if="!assetStore.selectedAsset.is_link" class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.checkpointComment') }}
            </div>
            <div class="simple-text-value">
              {{ lastCheckpoint.comment }}
            </div>
          </div>

          <div v-if="lastCheckpoint?.comment !== $t('panes.noCheckpoints') && !assetStore.selectedAsset.is_link"
            class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.checkpointDate') }}
            </div>
            <div class="simple-text-value">
              {{ formatMtime(lastCheckpoint.created_at) }}
            </div>
          </div>
          

          <div v-if="!assetStore.selectedAsset.is_link" class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.location') }}
            </div>
              <div class="simple-text-value">
                {{ assetStore.selectedAsset.file_path }}
              </div>
              <div v-if="!platformStore.isWeb" class="pane-parameter-actions">
                <ActionButton :icon="getAppIcon('copy')" v-tooltip="$t('common.copyPath')" @click="copyTaskPath('task')"/>
                <ActionButton :icon="getAppIcon('folder-arrow-up-right')" v-tooltip="$t('common.revealInExplorer')" :buttonFunction="revealInExplorer"/>
              </div>
          </div>

          <div v-if="!assetStore.selectedAsset.is_link" class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.fileState') }}
            </div>
            <div class="simple-text-value">
              {{ assetStore.selectedAsset.file_status }}
            </div>
          </div>

          <div v-if="!assetStore.selectedAsset.is_link" class="pane-parameter-detail">
          <div class="simple-text-key">
          {{ $t('panes.size') }}
          </div>
          <div class="simple-text-value">
            {{  assetSize }}
          </div>
        </div>

          <div v-if="assetStore.selectedAsset.tags.length" class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.tags') }}
            </div>
          </div>
        </div>

      </div>

      <div class="pane-parameter-section">
        <TagContainer :tags="assetStore.selectedAsset.tags" :displayOnly="true" />
      </div>

    </div>
  </div>



</template>

<script setup>
// imports
import { ref, computed, onMounted, watch, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { FSService } from '@/services';
import { Clipboard } from '@wailsio/runtime';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// store imports
import { useProjectStore } from '@/stores/projects';
import { useUserStore } from '@/stores/users';
import { useStageStore } from '@/stores/stages';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useAssetStore } from '@/stores/assets';
import { useStatusStore } from '@/stores/status';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useCommonStore } from '@/stores/common';
import { usePlatformStore } from '@/stores/platform';

// services
import { AssetService, CheckpointService } from "@/services";

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import TagContainer from '@/instances/common/components/TagContainer.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import AssigneeItem from '@/instances/common/components/AssigneeItem.vue';

// stores
const assetStore = useAssetStore();
const stage = useStageStore();
const userStore = useUserStore();
const modals = useDesktopModalStore();
const statusStore = useStatusStore();
const projectStore = useProjectStore();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const commonStore = useCommonStore();
const platformStore = usePlatformStore();
const { t } = useI18n();

// refs
const numberOfSelectedTasks = ref(0);
const multiStatusChange = ref(false);
const latestCheckpoint = ref(null);

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
  const isSingleTask = stage.markedTasks.length <= 1 && assetStore.selectedAsset;
  return isSingleTask
});

const selectedTaskName = computed(() => {
  if (assetStore.selectedAsset) {
    return singleTask.value ? assetStore.selectedAsset.name : t('panes.multipleTasksSelected')
  }
});

const selectedTaskIcon = computed(() => {
  if (assetStore.selectedAsset) {
    return singleTask.value ? assetStore.selectedAsset.icon : '/icons/categories.svg'
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
  let task = assetStore.selectedAsset;
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
  await Clipboard.SetText(taskPath);
  const message = t('notifications.pathCopied');
  notificationStore.addNotification(message, "", "success");
};

const showAllTasks = () => {
  commonStore.activeWorkspace = 'Default'
  commonStore.resetFilters();
  commonStore.navigatorMode = false;
  if (assetStore.selectedAsset?.assignee_id) {
    const assignee = userStore.getUserData(assetStore.selectedAsset.assignee_id);
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
  const taskId = assetStore.selectedAsset.id;
  if(assetStore.selectedAsset.file_status == "rebuildable"){
    await CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [taskId])
    .then( async (response) => {
      assetStore.rebuildableAssetsPath = assetStore.rebuildableAssetsPath.filter(taskPath => taskPath !== task.task_path)
      assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(taskPath => taskPath !== task.task_path);
      emitter.emit('get-project-data')
    })
    .catch((error) => {
      notificationStore.errorNotification(t('notifications.errorDownloadingTask'), error);
      console.error(error);
    });

  } 
  AssetService.RevealAsset(projectStore.activeProject.uri, assetStore.selectedAsset.id);
};

const toggleIsTask = async () => {
  stage.operationActive = true;
  const projectPath = projectStore.activeProject.uri;
  let isTask = assetStore.selectedAsset.is_resource;
  let task = assetStore.selectedAsset;
    
  await AssetService.ToggleIsTask(projectPath, task.id,  isTask)
    .then((data) => {

      assetStore.selectedAsset.is_resource = !isTask;
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
  const taskTypes = assetStore.getAssetTypes;
  newTaskType = taskTypes.find((item) => item.name === taskTypeName);

  const projectPath = projectStore.activeProject.uri;
  let task = assetStore.selectedAsset;

  await AssetService.UpdateAsset(projectPath, task.id, task.name, newTaskType.id, task.is_resource, '', task.tags)
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
  let task = assetStore.selectedAsset;
  
  await AssetService.ChangeStatus(projectPath, task.id, status.id)
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
  let assigneeId = assetStore.selectedAsset.assignee_id
  let user = userStore.getUserData(assigneeId);
  if (assigneeId && user) {
    let fullname = `${user.first_name} ${user.last_name}`;
    return fullname
  } else if(!assigneeId) {
    return t('panes.nobody')
  } else {
    return t('notifications.removedUser')
  }
});

const lastCheckpoint = computed(() => {
  let task = assetStore.selectedAsset;
  if (task.is_link) return { comment: '', created_at: task.created_at };
  
  if (latestCheckpoint.value) {
    return { 
      comment: latestCheckpoint.value.comment, 
      created_at: latestCheckpoint.value.created_at 
    };
  }
  
  return { comment: t('panes.noCheckpoints'), created_at: task.created_at };
});

const loadLatestCheckpoint = async () => {
  latestCheckpoint.value = null;
  
  if (!assetStore.selectedAsset || assetStore.selectedAsset.is_link) {
    return;
  }

  try {
    const checkpoint = await CheckpointService.GetLatestCheckpoint(
      projectStore.activeProject.uri,
      assetStore.selectedAsset.id
    );
    latestCheckpoint.value = checkpoint;
  } catch (error) {
    console.log('No checkpoint found or error:', error);
    latestCheckpoint.value = null;
  }
};

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
  const path = assetStore.selectedAsset?.file_path;
  return path?.replace(/\\/g, '/')
});

const getAssetSize = async() => {
  const size = await FSService.FileStat(assetPath.value);
  assetSize.value = size.formattedSize;
}

const getProjectData = async () => {
  if (!await FSService.Exists(assetPath.value)){
    assetSize.value = t('panes.notOnDisk')
    return
  }
  getAssetSize();
  loadLatestCheckpoint();
}

watch(() => assetStore.selectedAsset, () => {
  assetSize.value = 0;
  getProjectData();
  loadLatestCheckpoint();
});


// onMounted
onMounted(() => {
  stage.markedTasks = [];
  
  getProjectData();
  loadLatestCheckpoint();
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
  font-size: 14px;
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





