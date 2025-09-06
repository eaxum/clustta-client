<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>
    <HeaderArea :title="'Create Checkpoint'" :icon="'layers-plus'" />
    <div class="general-container">

      <textarea v-model="message" class="desktop-input-long" type="text" placeholder="make a comment..." v-focus
        @keydown.enter="handleEnterKey" />

      <div v-if="!isValueChanged" class="horizontal-flex input-alert">
        Your message is too short.
      </div>

      <div v-if="!statusMenuDisplayed" class="attachment-area">
        <div class="task-item-status-container" v-stop-propagation>
          <div class="task-item-status" @click="toggleDisplayStatusMenu()"
            :style="{ backgroundColor: taskStatus.color }">
            {{ taskStatus.short_name }}
          </div>
        </div>
        <ActionButton :icon="getAppIcon('paperclip')" v-tooltip="'Attach Snapshot'" v-stop-propagation
          :buttonFunction="selectPreviewFile" />
        <ActionButton :icon="getAppIcon('clipboard')" v-tooltip="'Paste Snapshot'" v-stop-propagation
          :buttonFunction="addImageFromClipBoard" />
        <ActionButton v-if="trayStates.screenshot" :icon="getAppIcon('trash')" v-tooltip="'Delete Snapshot'"
          v-stop-propagation :buttonFunction="removePreveiw" />
      </div>

      <div v-else class="status-section">
        <div class="task-item-status-container status-displayed">
          <StatusMenu @statusSelected="closeStatusMenu" />
        </div>
      </div>

      <span v-if="trayStates.screenshot" class="screenshot-preview">
        <img class="screenshot-thumb" :src="trayStates.screenshot">
      </span>

      <div class="horizontal-flex">
        <div class="input-label"> Use Image as Task Cover</div>
        <ToggleSwitch :switchValueProp="useImageAsCover" @click="useAsCover()" />
      </div>

    </div>

    <div class="pop-up-actions">
      <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
      <GeneralButton :label="'Create'" :fullWidth="true" @click="createCheckPoint" :isActive="isValueChanged"
        :loading="isAwaitingResponse" />
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
// utils/services
import { onMounted, watchEffect, onUnmounted, ref, computed } from 'vue';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';
import { CheckpointService, ClipboardService } from "@/../bindings/clustta/services";

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import StatusMenu from '@/instances/desktop/menus/StatusMenu.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import { v4 as uuidv4 } from 'uuid';

// state imports
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useUserStore } from '@/stores/users';
import { useTrayStates } from '@/stores/TrayStates';
import { useNotificationStore } from '@/stores/notifications';
import { useTaskStore } from '@/stores/task';
import { useProjectStore } from '@/stores/projects';
import { DialogService } from '@/../bindings/clustta/services/index';
import { useStatusStore } from '@/stores/status';

// refs
const userStore = useUserStore();
const trayStates = useTrayStates();
const message = ref('');
const taskStore = useTaskStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const statusStore = useStatusStore();
const projectStore = useProjectStore();
const useImageAsCover = ref(true);
const isAwaitingResponse = ref(false);
const displayStatusMenu = ref(false);
const modalContainer = ref(null);

// computed
const isValueChanged = computed(() => {
  const forbiddenComments = [
    'wip',
    'wfa',
    'retake',
    'retook',
    'todo',
    'fmf'
  ]
  return message.value.trim().length > 6 && !forbiddenComments.some(comment =>
    message.value.toLowerCase().includes(comment.toLowerCase())
  );
});

const statusMenuDisplayed = computed(() => { return taskStore.selectedTask.type !== "untracked_task" && displayStatusMenu.value });

const taskStatus = computed(() => {
  if (taskStore.selectedTask.type === "untracked_task") {
    return statusStore.statuses.find((item) => item.name === 'todo')
  }
  return taskStore.selectedTask.status;
});

// methods
const closeStatusMenu = () => {
  displayStatusMenu.value = false;
};

const toggleDisplayStatusMenu = () => {
  if (!userStore.canDo('change_status')) {
    return
  }
  taskStore.isTaskStatus = true;
  displayStatusMenu.value = true;
};

const useAsCover = () => {
  useImageAsCover.value = !useImageAsCover.value;
};

const escape = () => {
  closeModal('createCheckpointModal');
};

const closeModal = () => {
  modals.disableAllModals();
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    createCheckPoint();
  }
};

const createCheckPoint = async () => {
  isAwaitingResponse.value = true;
  let taskPath = taskStore.selectedTask.task_path
  let comment = message.value
  let previewPath = trayStates.previewFullPath;
  let groupId = uuidv4()
  if (taskStore.selectedTask.type === "task") {
    CheckpointService.AddCheckpoint(projectStore.activeProject.uri, [taskPath], comment, previewPath, groupId, useImageAsCover.value)
      .then((response) => {
        emitter.emit('refresh-browser');
        taskStore.modifiedTasksPath = taskStore.modifiedTasksPath.filter((modifiedTaskPath) => modifiedTaskPath !== taskPath)
        taskStore.selectedTask.file_status = "normal"
        projectStore.refreshProjects();
        isAwaitingResponse.value = false;
        closeModal();
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        notificationStore.errorNotification("Error Creating Checkpoint", error)
      });
  } else {
    await CheckpointService.AddUntrackedTask(projectStore.activeProject.uri, projectStore.activeProject.working_directory, [taskPath], 0, 1, comment, previewPath, groupId)
      .then((response) => {
        taskStore.untrackedTasksPath = taskStore.untrackedTasksPath.filter((path) => path !== taskPath)
        emitter.emit('refresh-browser');
        projectStore.refreshProjects();
        isAwaitingResponse.value = false;
        closeModal();
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        notificationStore.errorNotification("Error Creating Checkpoint", error)
      });
  }

};

const selectPreviewFile = async () => {
  if (!trayStates.userPin) {
    await trayStates.togglePin()
  }
  const result = await DialogService.SelectFileDialog("Select Image File", "*.png; *.jpg; *.jpeg; *.gif; *.bmp; *.tiff; *.webp");
  if (result) {
    //console.log(result);
    let filePath = result.replace(/\\/g, '/');
    let fileName = filePath.split('/').pop();
    let base64Image = await utils.base64FromFile(filePath)
    trayStates.previewFile = fileName;
    trayStates.previewFullPath = filePath;
    // trayStates.screenshot = await utils.base64FromFile(filePath)
    trayStates.screenshot = base64Image
  }

  if (!trayStates.userPin) {
    await trayStates.togglePin()
  }
}

const removePreveiw = () => {
  trayStates.screenshot = ""
  trayStates.previewFile = ""
  trayStates.previewFullPath = ""
}

const addImageFromClipBoard = () => {
  ClipboardService.ReadImageBase64()
    .then(async (base64Img) => {
      let imageStr = `data:image/png;base64, ${base64Img}`;
      let width;
      let height;

      // getImageDimensions(imageStr, function(dimensions) {
      //   console.log(`Width: ${dimensions.width}, Height: ${dimensions.height}`);
      // });

      // width = dimensions.width;
      // console.log(getImageDimensions());
      trayStates.screenshot = imageStr
      trayStates.previewFullPath = await utils.base64ToFile(base64Img);
    })
    .catch((err) => {
      console.log(err)
      //pass
    });
}

onMounted(
  async () => {
    trayStates.screenshot = null
    trayStates.previewFile = ""
    trayStates.previewFullPath = ""
  }
);

watchEffect(() => {
  if (modalContainer.value) {
    modalContainer.value.addEventListener('click', closeStatusMenu);
  }
});

onUnmounted(() => {
  if (modalContainer.value) {
    modalContainer.value.removeEventListener('click', closeStatusMenu);
  }
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.attachment-area {
  display: flex;
  align-items: center;
  gap: .5rem;
  width: 100%;
  padding: .1rem .3rem;
  /* background-color: deepskyblue; */
  box-sizing: border-box;
}

.status-section {
  display: flex;
  /* background-color: forestgreen; */
  width: 100%;
}

.task-item-status-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  width: 100%;
  padding: .4rem .2rem;
  height: 3.2rem;
  overflow: hidden;
  /* background-color: darkorange; */
  /* flex: 1; */
}

.status-displayed {
  justify-content: center;
  /* background-color: crimson; */

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

.general-container {
  gap: .5rem;
  align-items: center;
  justify-content: flex-start;
  /* background-color: hotpink; */
}

.desktop-input-long {
  margin-top: 20px;
  font-weight: 100;
  color: var(--white);
}

.input-alert {
  color: var(--attention);
}
</style>
