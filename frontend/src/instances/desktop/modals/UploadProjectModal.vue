<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>

    <HeaderArea :title="title" :icon="getAppIcon('arrow-up-ramp')" :showSearch="false" />

    <div class="general-container">

      <div class="input-section">
        <span class="input-label">{{ $t('modals.projectFileLabel') }}</span>
        <div class="horizontal-flex">
          <input v-model="selectedFilePath" class="input-short" type="text" :placeholder="$t('placeholders.selectClusttaFile')" readonly />
          <span @click="selectProjectFile" class="single-action-button" v-tooltip="$t('common.browse')">
            <img class="small-icons" :src="getAppIcon('clustta')">
          </span>
        </div>
        <InputAlert :show="fileError" :message="fileErrorMessage" />
      </div>

      <div class="input-section">
        <span class="input-label">{{ $t('placeholders.projectName') }}</span>
        <div class="horizontal-flex">
          <input v-model="projectName" class="input-short" type="text" :placeholder="$t('placeholders.projectName')"
            @keydown.enter="handleEnterKey" />
        </div>
        <InputAlert :show="!projectIsUploaded && projectNameInUse" :message="$t('notifications.projectNameInUse')" />
      </div>

      <div class="input-section">
        <span class="input-label">{{ $t('modals.locationLabel') }}</span>
        <div class="horizontal-flex">
          <div class="location-dropdown-wrapper">
            <DropDownBox 
              :items="locationDisplayNames" 
              :selectedItem="selectedLocationDisplay"
              :onSelect="selectLocation" 
            />
          </div>
          <span @click="addNewLocation" class="single-action-button" v-tooltip="$t('modals.addNewLocation')">
            <img class="small-icons" :src="getAppIcon('plus-circle')">
          </span>
        </div>
        <div v-if="workingDirectory" class="computed-path-display">
          {{ $t('modals.finalPath') }} {{ workingDirectory }}
        </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.upload')" :fullWidth="true" @click="uploadProject" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';

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
const { t } = useI18n();

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

// constants
const title = computed(() => t('modals.uploadProject'));

// computed
// Checks if form values are valid for upload.
const isValueChanged = computed(() => {
  return !projectNameEmpty.value && !projectNameInUse.value && selectedFilePath.value !== '' && !fileError.value;
});

// Returns display names for location dropdown.
const locationDisplayNames = computed(() => {
  return projectLocations.value.map(loc => `${loc.name} - [${loc.path}]`);
});

// Checks if project name field is empty.
const projectNameEmpty = computed(() => {
  return projectName.value === '';
});

// Checks if project name is already in use.
const projectNameInUse = computed(() => {
  const projectNames = projectStore.projects.map((project) => project.name.toLowerCase());
  return projectNames.includes(projectName.value.toLowerCase());
});

// Returns display string for selected location.
const selectedLocationDisplay = computed(() => {
  if (!selectedLocation.value) return '';
  return `${selectedLocation.value.name} - [${selectedLocation.value.path}]`;
});

// Computes the working directory path.
const workingDirectory = computed(() => {
  if (!selectedLocation.value || !projectName.value) return '';
  const studioName = projectStore.selectedStudio.name;
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
    notificationStore.addNotification(t('notifications.locationAdded'), '', 'success', false);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorAddingLocation'), error);
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

// Handles enter key press to trigger upload.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    uploadProject();
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
    notificationStore.errorNotification(t('notifications.errorLoadingLocations'), error);
  } finally {
    isLoadingLocations.value = false;
  }
};

// Resets project data stores to initial state.
const resetProjectData = () => {
  commonStore.activeWorkspace = 'Default';
  commonStore.resetFilters();
  collectionStore.collections = [];
  assetStore.assets = [];
  collectionStore.selectedCollection = null;
  assetStore.selectedAsset = null;
  stage.expandedCollections = {};
};

// Selects a location from the dropdown.
const selectLocation = (displayName) => {
  const location = projectLocations.value.find(loc => 
    `${loc.name} - [${loc.path}]` === displayName
  );
  if (location) {
    selectedLocation.value = location;
  }
};

// Opens dialog to select project file.
const selectProjectFile = async () => {
  try {
    const result = await DialogService.SelectFileDialog("Select Project File", "*.clst");
    if (!result) return;
    
    selectedFilePath.value = result.replace(/\\/g, '/');
    fileError.value = false;
    fileErrorMessage.value = '';
    
    const fileName = selectedFilePath.value.split('/').pop();
    const extractedName = fileName.replace('.clst', '');
    projectName.value = extractedName;
    
    const isValid = await ProjectService.ValidateProjectFile(selectedFilePath.value);
    if (!isValid) {
      fileError.value = true;
      fileErrorMessage.value = t('notifications.invalidProjectFile');
    }
  } catch (error) {
    fileError.value = true;
    fileErrorMessage.value = t('notifications.errorSelectingFile') + (error.message || error);
  }
};

// Uploads the project to the selected location.
const uploadProject = async () => {
  if (!selectedLocation.value) {
    notificationStore.addNotification(
      t('notifications.noLocationSelected'),
      t('notifications.selectOrAddLocation'),
      'error',
      false
    );
    return;
  }

  if (!workingDirectory.value) {
    notificationStore.addNotification(
      t('notifications.invalidWorkingDirectory'),
      t('notifications.workingDirectoryEmpty'),
      'error',
      false
    );
    return;
  }

  if (!selectedFilePath.value) {
    notificationStore.addNotification(
      t('notifications.noFilesSelectedUpload'),
      t('notifications.selectProjectFile'),
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
      t('notifications.projectUploadedSuccessfully'),
      t('notifications.triggerSyncToPush'),
      'success',
      false
    );

    closeModal();
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification(t('notifications.errorUploadingProject'), error);
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

// lifecycle
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
