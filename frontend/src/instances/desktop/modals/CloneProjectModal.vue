<template>
  <div ref="modalContainer" class="modal-container">

    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="getAppIcon('briefcase-plus')" :showSearch="false" />
    </div>

    <div class="general-container">

      <span v-if="projectStore.activeProjectCover" class="screenshot-preview">
        <img class="screenshot-thumb" :src="projectStore.activeProjectCover">
      </span>
      <div class="input-section">
        <span class="regular">Working Folder</span>
        <div class="horizontal-flex">
          <input v-model="workingDirectory" class="input-short" type="text"
            placeholder="Working Directory" ref="workingDirectoryInput" v-focus />
          <span @click="pasteDirectoryPath('workingDir')" class="single-action-button" v-tooltip="'Paste link'"><img
              class="small-icons" :src="getAppIcon('clipboard')"></span>
          <span @click="selectDirectoryPath('workingDir')" class="single-action-button" v-tooltip="'Browse Path'"><img
              class="small-icons" :src="getAppIcon('explorer')"></span>
        </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Download'" :fullWidth="true" @click="cloneProject" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
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
import { useTrayStates } from '@/stores/TrayStates';
import { ref, onMounted, computed, watchEffect } from 'vue';

// services
import { SyncService } from "@/../bindings/clustta/services";

//stores
import { useModalStore } from '@/stores/modals';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useProjectStore } from '@/stores/projects';
import { useNotificationStore } from '@/stores/notifications';
import { useEntityStore } from '@/stores/entity';
import { useStageStore } from '@/stores/stages';
import { usePaneStore } from '@/stores/panes';
import { useTaskStore } from '@/stores/task';
import { useCommonStore } from '@/stores/common';
import { useMenu } from '@/stores/menu';
import { useSettingsStore } from '@/stores/settings';

//components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import { useProjectTemplateStore } from '@/stores/project_template';
import { ClipboardService, SettingsService, DialogService } from '@/../bindings/clustta/services/index';

//header vars


// stores/states
const modals = useDesktopModalStore();
const projectStore = useProjectStore()
const notificationStore = useNotificationStore();
const menu = useMenu();

//refs
const isAwaitingResponse = ref(false);
const modalContainer = ref(null);

const workingDirectory = ref('');
const workingDirectoryInput = ref(null);

let title = `Download "${projectStore.activeProject.name}"`;

const pasteDirectoryPath = async (context) => {
  ClipboardService.ReadText()
    .then(path => {
      if (context === 'shared') {
        sharedProjectsDirectory.value = path;
      } else if (context === 'personal') {
        projectsDirectory.value = path;
      }else if (context === 'workingDir') {
        workingDirectory.value = path;
      }
    }).catch(err => {
      console.error("Failed to paste from clipboard:", err);
    });
};

const selectDirectoryPath = async (context) => {
  const result = await DialogService.SelectFolderDialog("Select Folder File");
  if (result) {
    let fileDir = result.replace(/\\/g, '/');
    workingDirectory.value = fileDir;
    workingDirectoryInput.value.focus();
    
  }
};

const cloneProject = async () => {
  let project = projectStore.activeProject;
  let studioDisplayName = projectStore.selectedStudio.name;
  const projectName = project.name;
  const projectUrl = projectStore.getStudioUrl + '/' + projectName;
  let syncOptions = {
    only_latest_checkpoints: true,
    task_dependencies: true,
    tasks: false,
    templates: true,
  };
  notificationStore.cancleFunction = SyncService.CancelSync
  notificationStore.canCancel = true
  await SyncService.CloneProject(projectUrl, studioDisplayName, workingDirectory.value, syncOptions)
    .then(async () => {
      projectStore.projects.find(p => p.name === projectName).working_directory = workingDirectory.value;
      projectStore.activeProject.working_directory = workingDirectory.value;
      console.log('Project cloned successfully')
      closeModal()
      await projectStore.refreshProjects()
      await projectStore.refreshProjectsPreview()
    }).catch((error) => {
      console.error(error)
      notificationStore.errorNotification(
        "Error Cloning Project",
        error
      )
    })
  closeModal()
}

const isValueChanged = computed(() => {
  return workingDirectory .value !== ''
});

const closeModal = () => {
  modals.disableAllModals();
};

watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

onMounted(async () => {
  await SettingsService.GetWorkingDirectory()
    .then((response) => {
      workingDirectory.value = response.replace(/\\/g, '/') + '/' + projectStore.selectedStudio.name + '/' + projectStore.activeProject.name;
    })
    .catch((error) => {
      notificationStore.addNotification(
        "Error Loading Settings",
        error.message,
        "error",
        false
      )
    });
});

</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/modals.css";

.input-section {
  width: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  justify-content: flex-start;
  gap: .4px;
  color: white;
}

.regular{
  color: var(--white);
}
.general-container {
  gap: 1rem;
}

.modal-info {
  display: flex;
  flex-direction: column;
  max-width: 100%;
  justify-content: flex-start;
  align-self: stretch;
  width: 464px;
  align-items: flex-start;
  box-sizing: border-box;

}

.modal-text-container {
  display: flex;
  flex-direction: column;
  max-width: 100%;
  justify-content: flex-start;
  align-self: stretch;
  width: 464px;
  align-items: flex-start;
  /* margin-top: 20px; */
}

.modal-title {
  max-width: 100%;
  align-self: stretch;
  width: 464px;
  color: rgba(16, 24, 40, 1);
  color: var(--white);
  font-size: 18px;
  line-height: 28px;
  letter-spacing: 0%;
  text-align: left;
}

.input-header {
  /* background-color: lightblue; */
  width: 100%;
  display: flex;
  align-items: center;
  margin: 10px 0px;
}

.input-count {
  background-color: none;
  font-size: 14px;
  color: var(--white);
}

.modal-subtitle {
  /* background-color: beige; */
  /* max-width: 100%; */
  align-self: stretch;
  width: 464px;
  color: rgba(16, 24, 40, 1);
  color: var(--white);
  font-size: 14px;
  /* line-height: 28px; */
  letter-spacing: 0%;
  text-align: left;
}



.modal-body {
  box-sizing: border-box;
  max-width: 100%;
  align-self: stretch;
  width: 464px;
  margin: 8px 0px;
  font-size: 14px;
  color: rgba(16, 24, 40, 1);
  line-height: 20px;
  letter-spacing: 0%;
  text-align: left;
}

.modal-actions {
  box-sizing: border-box;
  padding: 1rem 2rem;
  gap: 2rem;
  display: flex;
  flex-direction: row;
  max-width: 100%;
  align-self: stretch;
  align-items: center;
  justify-content: space-evenly;
  width: 464px;
  margin-top: 32px;
}

.div-10 {
  display: flex;
}

.task-options-container {
  position: relative;
  box-sizing: border-box;
  width: 100%;
  height: max-content;
  height: 40px;
  transition: all .2s ease-in-out;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0;
}

.task-options-container-closed {
  height: 0px;
  padding: 0;
  margin-bottom: -1rem;
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
  color: var(--white);
  font-size: 16px;
  white-space: nowrap;
  flex: 1;
}

.pop-up-prompt {
  gap: 10px;
  align-items: center;
  justify-content: center;
}
</style>
