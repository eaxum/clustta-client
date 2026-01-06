<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>

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
        <InputAlert :show="!projectIsCreated && projectNameInUse" message="A project with this name already exists." />
      </div>

      <div class="input-section">
        <span class="input-label">Location</span>
        <div class="horizontal-flex">
          <div class="location-dropdown-wrapper">
            <DropDownBox 
              :items="locationDisplayNames" 
              :selectedItem="selectedLocationDisplay"
              :onSelect="selectLocation" 
            />
          </div>
          <span @click="addNewLocation" class="single-action-button" v-tooltip="'Add New Location'">
            <img class="small-icons" :src="getAppIcon('plus-circle')">
          </span>
        </div>
        <div v-if="workingDirectory" class="computed-path-display">
          Final path: {{ workingDirectory }}
        </div>
      </div>

      <div v-if="projectTemplateStore.projectTemplates.length" class="input-section drop-down-box-section">
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
import { ProjectService, SyncService } from "@/services";

//stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useProjectStore } from '@/stores/projects';
import { useNotificationStore } from '@/stores/notifications';
import { useCollectionStore } from '@/stores/collections';
import { useStageStore } from '@/stores/stages';
import { useAssetStore } from '@/stores/assets';
import { useCommonStore } from '@/stores/common';
import { useMenu } from '@/stores/menu';

//components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';
import { useProjectTemplateStore } from '@/stores/project_template';
import { ClipboardService, SettingsService, DialogService } from '@/services';

//header vars
let title = 'Add Project';

// stores/states
const modals = useDesktopModalStore();
const projectStore = useProjectStore()
const projectTemplateStore = useProjectTemplateStore();
const notificationStore = useNotificationStore();
const collectionStore = useCollectionStore();
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
const projectLocations = ref([]);
const selectedLocation = ref(null);
const isLoadingLocations = ref(false);

const workingDirectory = computed(() => {
  if (!selectedLocation.value || !projectName.value) return '';
  const studioName = projectStore.selectedStudio.name;
  
  // For personal projects, don't include studio name in working directory
  if (studioName === 'Personal') {
    return `${selectedLocation.value.path}/${projectName.value}`;
  }
  
  // For studio projects, include studio name
  return `${selectedLocation.value.path}/${studioName}/${projectName.value}`;
});

// Computed properties for DropDownBox
const locationDisplayNames = computed(() => {
  return projectLocations.value.map(loc => `${loc.name} - [${loc.path}]`);
});

const selectedLocationDisplay = computed(() => {
  if (!selectedLocation.value) return '';
  return `${selectedLocation.value.name} - [${selectedLocation.value.path}]`;
});

const selectLocation = (displayName) => {
  const location = projectLocations.value.find(loc => 
    `${loc.name} - [${loc.path}]` === displayName
  );
  if (location) {
    selectedLocation.value = location;
  }
};

const loadProjectLocations = async () => {
  isLoadingLocations.value = true;
  try {
    const locations = await SettingsService.GetAllLocationPaths();
    projectLocations.value = locations;
    
    // Set default location as selected
    const defaultLoc = locations.find(loc => loc.is_default);
    selectedLocation.value = defaultLoc || locations[0];
  } catch (error) {
    notificationStore.errorNotification('Error loading locations', error);
  } finally {
    isLoadingLocations.value = false;
  }
};

const addNewLocation = async () => {
  // Get user's Documents folder as default starting point
  const userDirectory = await SettingsService.GetUserDirectory();
  const documentsPath = userDirectory + 'Documents';
  
  const result = await DialogService.SelectSpecificFolderDialog("Select New Location Folder", documentsPath);
  if (!result) return;
  
  const path = result.replace(/\\/g, '/');
  
  // Extract folder name from path
  const pathParts = path.split('/');
  const folderName = pathParts[pathParts.length - 1] || `Location ${projectLocations.value.length + 1}`;
  
  try {
    const newLocation = await SettingsService.AddProjectLocation(folderName, path);
    projectLocations.value.push(newLocation);
    selectedLocation.value = newLocation;
    notificationStore.addNotification('Location added successfully', '', 'success', false);
  } catch (error) {
    notificationStore.errorNotification('Error adding location', error);
  }
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

  // Validate that a location is selected
  if (!selectedLocation.value) {
    notificationStore.addNotification(
      'No location selected',
      'Please select or add a project location',
      'error',
      false
    );
    return;
  }

  // Validate working directory is not empty
  if (!workingDirectory.value) {
    notificationStore.addNotification(
      'Invalid working directory',
      'Working directory cannot be empty',
      'error',
      false
    );
    return;
  }

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

    // Assign project to selected location
    if (selectedLocation.value) {
      try {
        await SettingsService.AssignProjectToLocation(project.id, selectedLocation.value.id);
      } catch (error) {
        console.error('Error assigning project to location:', error);
      }
    }

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
      
      // Refresh project info to get local file path after cloning
      await projectStore.refreshProjects()
      
      // Get updated project info with local path
      const updatedProject = projectStore.projects.find(p => p.name === projectName);
      if (updatedProject) {
        projectStore.activeProject = updatedProject;
      }
      
      // Apply template if one was selected
      if (selectedProjectTemplate.value && selectedProjectTemplate.value !== 'No Template') {
        console.log('Applying template:', selectedProjectTemplate.value)
        // Now use the updated project info which has the local .clst path
        const localProjectPath = projectStore.activeProject.uri;
        console.log('Local project path:', localProjectPath)
        
        await ProjectService.ApplyTemplate(localProjectPath, selectedProjectTemplate.value)
          .then(async () => {

            // Sync template changes back to server
            const syncOptions = {
              only_latest_checkpoints: false,
              task_dependencies: false,
              tasks: false,
              templates: false,
            };
            await SyncService.SyncData(localProjectPath, projectUrl, false, syncOptions)
              .then(() => {
              })
              .catch((error) => {
                console.error('Failed to sync template changes:', error)
                notificationStore.addNotification('Template applied locally but sync failed', 'warning')
              })
          })
          .catch((error) => {
            console.error('Failed to apply template:', error)
            notificationStore.errorNotification('Failed to apply template', error)
          })
      }
      
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

  collectionStore.collections = [];
  assetStore.assets = [];

  collectionStore.selectedCollection = null;
  assetStore.selectedAsset = null;

  stage.expandedEntities = {};

};

watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

onMounted(async () => {
  await loadProjectLocations();
  await projectTemplateStore.loadProjectTemplates();
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

.pop-up-prompt {
  gap: 10px;
  align-items: center;
  justify-content: center;
}

.location-dropdown-wrapper {
  flex: 1;
  width: 100%;
}

.computed-path-display {
  font-size: 12px;
  color: var(--text-secondary);
  padding: 0.5rem;
  background: var(--bg-secondary);
  border-radius: 4px;
  margin-top: 0.5rem;
  word-break: break-all;
}
</style>






