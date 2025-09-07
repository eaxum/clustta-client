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
        <div class="horizontal-flex">
          <input v-model="projectName" @input="updateWorkingDirectory" class="input-short" type="text" placeholder="Project Name" ref="projectNameInput"
            @keydown.enter="handleEnterKey" v-focus />
        </div>
        <div v-if="!projectIsCreated && projectNameInUse" class="horizontal-flex input-alert">
          A project with this name already exists.
        </div>
      </div>

      <div class="input-section">
        <span class="regular">Project Folder</span>
        <div class="horizontal-flex">
          <input v-model="workingDirectory" class="input-short" type="text"
            placeholder="Project Folder" ref="workingDirectoryInput" />
          <span @click="resetDefaultPath()" class="single-action-button" v-tooltip="'Reset'"><img
              class="small-icons" :src="getAppIcon('refresh')"></span>
          <span @click="selectDirectoryPath('workingDir')" class="single-action-button" v-tooltip="'Browse Path'"><img
              class="small-icons" :src="getAppIcon('explorer')"></span>
        </div>
      </div>

      <div v-if="projectStore.selectedStudio.name === 'Personal' && projectTemplateStore.projectTemplates.length" class="input-section drop-down-box-section">
        <DropDownBox :items="projectTemplateNames" :selectedItem="selectedProjectTemplate"
          :onSelect="selectProjectTemplate" />
      </div>



      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Create'" :fullWidth="true" @click="createProject" :isActive="isValueChanged"
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
import { ref, onMounted, computed, watchEffect } from 'vue';

// services
import { ProjectService, SyncService } from "@/../bindings/clustta/services";

//stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useProjectStore } from '@/stores/projects';
import { useNotificationStore } from '@/stores/notifications';
import { useEntityStore } from '@/stores/entity';
import { useStageStore } from '@/stores/stages';
import { useAssetStore } from '@/stores/assets';
import { useCommonStore } from '@/stores/common';
import { useMenu } from '@/stores/menu';

//components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import { useProjectTemplateStore } from '@/stores/project_template';
import { ClipboardService, SettingsService, DialogService } from '@/../bindings/clustta/services/index';

//header vars
let title = 'Add Project';

// stores/states
const modals = useDesktopModalStore();
const projectStore = useProjectStore()
const projectTemplateStore = useProjectTemplateStore();
const notificationStore = useNotificationStore();
const entityStore = useEntityStore();
const stage = useStageStore();
const assetStore = useAssetStore();
const commonStore = useCommonStore();
const menu = useMenu();

//refs
const isAwaitingResponse = ref(false);
const projectIsCreated = ref(false);
const modalContainer = ref(null);

const projectName = ref('');
const projectNameInput = ref(null);
const workingDirectory = ref('');
const defaultWorkingDirectory = ref('');
const selectedPath = ref('');
const workingDirectoryInput = ref(null);

const updateWorkingDirectory = () => {
  if (projectName.value !== '') {
    workingDirectory.value = selectedPath.value + '/' + projectName.value;
  } else {
    workingDirectory.value = selectedPath.value + '/';
  }
}

const selectDirectoryPath = async (context) => {
  const result = await DialogService.SelectFolderDialog("Select Folder File");
  if (result) {
    let fileDir = result.replace(/\\/g, '/');
    selectedPath.value = fileDir;
    if(projectName.value){
      workingDirectory.value = selectedPath.value + '/' + projectName.value;
    } else {
      workingDirectory.value = fileDir + '/';
    }

    projectNameInput.value.focus();
  }
};

const resetDefaultPath = () => {
  workingDirectory.value = defaultWorkingDirectory.value + '/' + projectName.value;
  selectedPath.value = defaultWorkingDirectory.value ;
};

const projectTemplateNames = computed(() => {
  return ['No Template', ...projectTemplateStore.projectTemplateNames]
});

const selectedProjectTemplate = ref(projectTemplateNames.value[0])

const selectProjectTemplate = (selectedTemplateName) => {
  selectedProjectTemplate.value = selectedTemplateName;
};

const restrictedNames = computed(() => {
  const projectNames = projectStore.projects.map((project) => project.name);
  let restrictedNames = [];
  for (let i = 0; i < projectNames.length; i++) {
    restrictedNames.push(projectNames[i].toLowerCase())
  }
  return restrictedNames;
});

const projectNameInUse = computed(() => {
  return restrictedNames.value.includes(projectName.value.toLowerCase());
});

const projectNameEmpty = computed(() => {
  return projectName.value === ''
});

const isValueChanged = computed(() => {
  return !projectNameEmpty.value && !projectNameInUse.value
});

const closeModal = () => {
  modals.disableAllModals();
};

const addCoverImage = () => {
  // TODO dialogue to select and add project cover image
};

const removeCoverImage = () => {
  // TODO dialogue to select and add project cover image
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    createProject();
  }
};

const createProject = async () => {

  isAwaitingResponse.value = true;

  let studio = projectStore.selectedStudio
  let projectFilepath = studio.url
  let name = projectName.value;
  let path = projectFilepath + '/' + name;
  path = path.replace(/\\/g, '/');

  if (studio.name === 'Personal') {
    path = path + ".clst"
  }

  ProjectService.CreateProject(path, studio.name, workingDirectory.value, selectedProjectTemplate.value).then(async (project) => {

    projectIsCreated.value = true;

    resetProjectData();
    await projectStore.loadProjects();
    projectStore.activeProject = project;

        
    if(studio.name !== 'Personal'){
      await cloneProject()
    }
    
    closeModal();
    isAwaitingResponse.value = false;

  }).catch((error) => {
    isAwaitingResponse.value = false
    console.log(error)
    notificationStore.errorNotification('Error creating project', error);
  });

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

const resetProjectData = () => {

  commonStore.activeWorkspace = 'Default';
  commonStore.resetFilters();

  entityStore.entities = [];
  assetStore.tasks = [];

  entityStore.selectedEntity = null;
  assetStore.selectedTask = null;

  stage.expandedEntities = {};

};

watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

onMounted(async () => {
  await SettingsService.GetWorkingDirectory()
    .then((response) => {
      workingDirectory.value = response.replace(/\\/g, '/') + '/' + projectStore.selectedStudio.name + '/';
      defaultWorkingDirectory.value = response.replace(/\\/g, '/') + '/' + projectStore.selectedStudio.name;
      selectedPath.value = response.replace(/\\/g, '/') + '/' + projectStore.selectedStudio.name;
    })
    .catch((error) => {
      notificationStore.addNotification(
        "Error Loading Settings",
        error.message,
        "error",
        false
      )
    });

  await projectTemplateStore.loadProjectTemplates()
});

onMounted(async () => {
  return
  await SettingsService.GetWorkingDirectory()
    .then((response) => {
      workingDirectory.value = response.replace(/\\/g, '/') + '/' + projectStore.selectedStudio.name + '/' + projectStore.activeProject?.name;
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

.regular{
  color: var(--white);
}

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
  color: white;
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
  color: white;
}

.modal-subtitle {
  /* background-color: beige; */
  /* max-width: 100%; */
  align-self: stretch;
  width: 464px;
  color: rgba(16, 24, 40, 1);
  color: white;
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
  color: white;
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



