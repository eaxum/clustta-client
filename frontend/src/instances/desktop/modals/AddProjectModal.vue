<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>

    <HeaderArea :title="title" :icon="getAppIcon('briefcase-plus')" :showSearch="false" />

    <div class="general-container">
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
// imports
import { computed, onMounted, ref, watchEffect } from 'vue';

// components
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';

// services
import { DialogService, ProjectService, SettingsService, SyncService } from '@/services';

// stores
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const projectTemplateStore = useProjectTemplateStore();
const stage = useStageStore();

import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useProjectTemplateStore } from '@/stores/project_template';
import { useStageStore } from '@/stores/stages';

// refs
const isAwaitingResponse = ref(false);
const isLoadingLocations = ref(false);
const modalContainer = ref(null);
const projectIsCreated = ref(false);
const projectLocations = ref([]);
const projectName = ref('');
const projectNameInput = ref(null);
const selectedLocation = ref(null);
const selectedProjectTemplate = ref('No Template');

// constants
const title = 'Add Project';

// computed
// Returns whether form values are valid for submission.
const isValueChanged = computed(() => {
  return !projectNameEmpty.value && !projectNameInUse.value;
});

// Returns display names for location dropdown.
const locationDisplayNames = computed(() => {
  return projectLocations.value.map(loc => `${loc.name} - [${loc.path}]`);
});

// Returns whether the project name field is empty.
const projectNameEmpty = computed(() => {
  return projectName.value === '';
});

// Returns whether the project name is already in use.
const projectNameInUse = computed(() => {
  return restrictedNames.value.includes(projectName.value.toLowerCase());
});

// Returns list of project template names with 'No Template' option.
const projectTemplateNames = computed(() => {
  return ['No Template', ...projectTemplateStore.projectTemplateNames];
});

// Returns list of restricted project names (lowercase).
const restrictedNames = computed(() => {
  return projectStore.projects.map((project) => project.name.toLowerCase());
});

// Returns display string for currently selected location.
const selectedLocationDisplay = computed(() => {
  if (!selectedLocation.value) return '';
  return `${selectedLocation.value.name} - [${selectedLocation.value.path}]`;
});

// Returns the computed working directory path.
const workingDirectory = computed(() => {
  if (!selectedLocation.value || !projectName.value) return '';
  const studioName = projectStore.selectedStudio.name;
  if (studioName === 'Personal') {
    return `${selectedLocation.value.path}/${projectName.value}`;
  }
  return `${selectedLocation.value.path}/${studioName}/${projectName.value}`;
});

// methods
// Adds a new project location via folder dialog.
const addNewLocation = async () => {
  const userDirectory = await SettingsService.GetUserDirectory();
  const documentsPath = userDirectory + 'Documents';
  const result = await DialogService.SelectSpecificFolderDialog("Select New Location Folder", documentsPath);
  if (!result) return;
  const path = result.replace(/\\/g, '/');
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

// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles enter key press to submit form.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    createProject();
  }
};

// Loads available project locations from settings.
const loadProjectLocations = async () => {
  isLoadingLocations.value = true;
  try {
    const locations = await SettingsService.GetAllLocationPaths();
    projectLocations.value = locations;
    const defaultLoc = locations.find(loc => loc.is_default);
    selectedLocation.value = defaultLoc || locations[0];
  } catch (error) {
    notificationStore.errorNotification('Error loading locations', error);
  } finally {
    isLoadingLocations.value = false;
  }
};

// Selects a location from the dropdown by display name.
const selectLocation = (displayName) => {
  const location = projectLocations.value.find(loc =>
    `${loc.name} - [${loc.path}]` === displayName
  );
  if (location) {
    selectedLocation.value = location;
  }
};

// Selects a project template from the dropdown.
const selectProjectTemplate = (selectedTemplateName) => {
  selectedProjectTemplate.value = selectedTemplateName;
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
  
  console.log(path)
  console.log(studio.name)
  console.log(workingDirectory.value)
  console.log(selectedProjectTemplate.value)

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

// Clones the project from server after creation.
const cloneProject = async () => {
  const project = projectStore.activeProject;
  const studioDisplayName = projectStore.selectedStudio.name;
  const projectName = project.name;
  const projectUrl = projectStore.getStudioUrl + '/' + projectName;
  const syncOptions = {
    only_latest_checkpoints: true,
    task_dependencies: true,
    tasks: false,
    templates: true,
  };
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  await SyncService.CloneProject(projectUrl, studioDisplayName, workingDirectory.value, syncOptions)
    .then(async () => {
      projectStore.projects.find(p => p.name === projectName).working_directory = workingDirectory.value;
      projectStore.activeProject.working_directory = workingDirectory.value;
      await projectStore.refreshProjects();
      const updatedProject = projectStore.projects.find(p => p.name === projectName);
      if (updatedProject) {
        projectStore.activeProject = updatedProject;
      }
      if (selectedProjectTemplate.value && selectedProjectTemplate.value !== 'No Template') {
        const localProjectPath = projectStore.activeProject.uri;
        await ProjectService.ApplyTemplate(localProjectPath, selectedProjectTemplate.value)
          .then(async () => {
            const templateSyncOptions = {
              only_latest_checkpoints: false,
              task_dependencies: false,
              tasks: false,
              templates: false,
            };
            await SyncService.SyncData(localProjectPath, projectUrl, false, templateSyncOptions)
              .catch((error) => {
                console.error('Failed to sync template changes:', error);
                notificationStore.addNotification('Template applied locally but sync failed', 'warning');
              });
          })
          .catch((error) => {
            console.error('Failed to apply template:', error);
            notificationStore.errorNotification('Failed to apply template', error);
          });
      }
      await projectStore.refreshProjectsPreview();
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification("Error Cloning Project", error);
    });
  closeModal();
};

// Resets project data in stores after creation.
const resetProjectData = () => {
  commonStore.activeWorkspace = 'Default';
  commonStore.resetFilters();
  collectionStore.collections = [];
  assetStore.assets = [];
  collectionStore.selectedCollection = null;
  assetStore.selectedAsset = null;
  stage.expandedEntities = {};
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// lifecycle hooks
onMounted(async () => {
  await loadProjectLocations();
  await projectTemplateStore.loadProjectTemplates();
});

</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/modals.css";

.computed-path-display {
  font-size: 12px;
  color: var(--text-secondary);
  padding: 0.5rem;
  background: var(--bg-secondary);
  border-radius: 4px;
  margin-top: 0.5rem;
  word-break: break-all;
}

.general-container {
  gap: 1rem;
}

.input-section {
  width: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  gap: .4px;
  color: var(--white);
}

.input-short {
  flex: 1;
  width: 100%;
}

.location-dropdown-wrapper {
  flex: 1;
  width: 100%;
}
</style>






