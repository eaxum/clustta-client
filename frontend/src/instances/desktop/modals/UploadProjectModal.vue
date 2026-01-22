<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>

    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="getAppIcon('arrow-up-ramp')" :showSearch="false" />
    </div>

    <div class="general-container">

      <div class="input-section">
        <span class="input-label">Project File</span>
        <div class="horizontal-flex">
          <input v-model="selectedFilePath" class="input-short" type="text" placeholder="Select .clst file..." readonly />
          <span @click="selectProjectFile" class="single-action-button" v-tooltip="'Browse'">
            <img class="small-icons" :src="getAppIcon('clustta')">
          </span>
        </div>
        <InputAlert :show="fileError" :message="fileErrorMessage" />
      </div>

      <div class="input-section">
        <span class="input-label">Project Name</span>
        <div class="horizontal-flex">
          <input v-model="projectName" class="input-short" type="text" placeholder="Project Name"
            @keydown.enter="handleEnterKey" />
        </div>
        <InputAlert :show="!projectIsUploaded && projectNameInUse" message="A project with this name already exists." />
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

      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Upload'" :fullWidth="true" @click="uploadProject" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { ref, onMounted, computed, watchEffect } from 'vue';

// components
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';

// services
import { DialogService, ProjectService, SettingsService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();

// refs
const fileError = ref(false);
const fileErrorMessage = ref('');
const isAwaitingResponse = ref(false);
const isLoadingLocations = ref(false);
const modalContainer = ref(null);
const projectIsUploaded = ref(false);
const projectLocations = ref([]);
const projectName = ref('');
const selectedFilePath = ref('');
const selectedLocation = ref(null);

// header vars
const title = 'Upload Project';

// methods
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// computed properties
const locationDisplayNames = computed(() => {
  return projectLocations.value.map(loc => `${loc.name} - [${loc.path}]`);
});

const projectNameEmpty = computed(() => {
  return projectName.value === '';
});

const projectNameInUse = computed(() => {
  const projectNames = projectStore.projects.map((project) => project.name.toLowerCase());
  return projectNames.includes(projectName.value.toLowerCase());
});

const selectedLocationDisplay = computed(() => {
  if (!selectedLocation.value) return '';
  return `${selectedLocation.value.name} - [${selectedLocation.value.path}]`;
});

const workingDirectory = computed(() => {
  if (!selectedLocation.value || !projectName.value) return '';
  const studioName = projectStore.selectedStudio.name;
  return `${selectedLocation.value.path}/${studioName}/${projectName.value}`;
});

const isValueChanged = computed(() => {
  return !projectNameEmpty.value && !projectNameInUse.value && selectedFilePath.value !== '' && !fileError.value;
});

// methods
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

const closeModal = () => {
  modals.disableAllModals();
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    uploadProject();
  }
};

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

const resetProjectData = () => {
  commonStore.activeWorkspace = 'Default';
  commonStore.resetFilters();
  collectionStore.collections = [];
  assetStore.assets = [];
  collectionStore.selectedCollection = null;
  assetStore.selectedAsset = null;
  stage.expandedEntities = {};
};

const selectLocation = (displayName) => {
  const location = projectLocations.value.find(loc => 
    `${loc.name} - [${loc.path}]` === displayName
  );
  if (location) {
    selectedLocation.value = location;
  }
};

const selectProjectFile = async () => {
  try {
    const result = await DialogService.SelectFileDialog("Select Project File", "*.clst");
    if (!result) return;
    
    selectedFilePath.value = result.replace(/\\/g, '/');
    fileError.value = false;
    fileErrorMessage.value = '';
    
    // Extract project name from file path
    const fileName = selectedFilePath.value.split('/').pop();
    const extractedName = fileName.replace('.clst', '');
    projectName.value = extractedName;
    
    // Validate the file
    const isValid = await ProjectService.ValidateProjectFile(selectedFilePath.value);
    if (!isValid) {
      fileError.value = true;
      fileErrorMessage.value = 'Invalid or corrupted project file.';
    }
  } catch (error) {
    fileError.value = true;
    fileErrorMessage.value = 'Error selecting file: ' + (error.message || error);
  }
};

const uploadProject = async () => {
  if (!selectedLocation.value) {
    notificationStore.addNotification(
      'No location selected',
      'Please select or add a project location',
      'error',
      false
    );
    return;
  }

  if (!workingDirectory.value) {
    notificationStore.addNotification(
      'Invalid working directory',
      'Working directory cannot be empty',
      'error',
      false
    );
    return;
  }

  if (!selectedFilePath.value) {
    notificationStore.addNotification(
      'No file selected',
      'Please select a .clst project file',
      'error',
      false
    );
    return;
  }

  isAwaitingResponse.value = true;

  const studio = projectStore.selectedStudio;
  const studioUrl = projectStore.getStudioUrl;
  const remoteProjectUrl = studioUrl + '/' + projectName.value;

  try {
    const project = await ProjectService.UploadProject(
      selectedFilePath.value,
      studio.name,
      workingDirectory.value,
      projectName.value,
      remoteProjectUrl
    );

    projectIsUploaded.value = true;

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

    notificationStore.addNotification(
      'Project uploaded successfully',
      'Trigger sync to push all data to the remote studio.',
      'success',
      false
    );

    closeModal();
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification('Error uploading project', error);
  } finally {
    isAwaitingResponse.value = false;
  }
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
  justify-content: flex-start;
  gap: .4px;
  color: var(--white);
}

.general-container {
  gap: 1rem;
}

.input-short {
  flex: 1;
  width: 100%;
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
