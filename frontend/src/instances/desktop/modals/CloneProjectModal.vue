<template>
  <div ref="modalContainer" class="modal-container">

      <HeaderArea :title="title" :icon="getAppIcon('briefcase-plus')" :showSearch="false" />

    <div class="general-container">

      <!-- Clone Progress Display -->
      <div v-if="isCloning" class="settings-section-card">
        <ProgressSection variant="success" />
      </div>

      <span v-if="!isCloning && projectStore.activeProjectCover" class="screenshot-preview">
        <img class="screenshot-thumb" :src="projectStore.activeProjectCover">
      </span>
      
      <div v-if="!isCloning" class="input-section">
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
      </div>

      <div v-if="!isCloning" class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :isActive="!isAwaitingResponse" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.download')" :fullWidth="true" @click="cloneProject" :isActive="isValueChanged"
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
import ProgressSection from '@/instances/common/components/ProgressSection.vue';

// services
import { DialogService, SettingsService, SyncService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';

const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const { t } = useI18n();

// refs
const isAwaitingResponse = ref(false);
const isCloning = ref(false);
const isLoadingLocations = ref(false);
const modalContainer = ref(null);
const projectLocations = ref([]);
const selectedLocation = ref(null);

// constants
const title = t('modals.downloadProject', { name: projectStore.activeProject.name });

// computed
// Returns whether the form is valid for submission.
const isValueChanged = computed(() => {
  return selectedLocation.value !== null && workingDirectory.value !== '';
});

// Returns display names for location dropdown.
const locationDisplayNames = computed(() => {
  return projectLocations.value.map(loc => `${loc.name}`);
});

// Returns display string for currently selected location.
const selectedLocationDisplay = computed(() => {
  if (!selectedLocation.value) return '';
  return `${selectedLocation.value.name}`;
});

// Returns the computed working directory path.
const workingDirectory = computed(() => {
  if (!selectedLocation.value) return '';
  const studioName = projectStore.selectedStudio.name;
  const projectName = projectStore.activeProject.name;
  if (studioName === 'Personal') {
    return `${selectedLocation.value.path}/${projectName}`;
  }
  return `${selectedLocation.value.path}/${studioName}/${projectName}`;
});

// methods
// Adds a new project location via folder dialog.
const addNewLocation = async () => {
  const userDirectory = await SettingsService.GetUserDirectory();
  const documentsPath = userDirectory + 'Documents';
  const result = await DialogService.SelectSpecificFolderDialog('Select New Location Folder', documentsPath);
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

// Clones the project from the server to the selected location.
const cloneProject = async () => {
  if (!selectedLocation.value) {
    notificationStore.addNotification(t('notifications.noLocationSelected'), t('notifications.selectOrAddLocation'), 'error', false);
    return;
  }
  if (!workingDirectory.value) {
    notificationStore.addNotification(t('notifications.invalidWorkingDirectory'), t('notifications.workingDirectoryEmpty'), 'error', false);
    return;
  }
  isAwaitingResponse.value = true;
  isCloning.value = true;
  stage.operationActive = true;
  const project = projectStore.activeProject;
  const studioDisplayName = projectStore.selectedStudio.name;
  const projectName = project.name;
  const projectUrl = (project.has_remote && project.remote) ? project.remote : projectStore.getStudioUrl + '/' + projectName;
  const syncOptions = {
    only_latest_checkpoints: true,
    asset_dependencies: true,
    assets: false,
    templates: true,
  };
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  try {
    await SyncService.CloneProject(projectUrl, studioDisplayName, workingDirectory.value, syncOptions);
    projectStore.projects.find(p => p.name === projectName).working_directory = workingDirectory.value;
    projectStore.activeProject.working_directory = workingDirectory.value;
    if (selectedLocation.value) {
      try {
        await SettingsService.AssignProjectToLocation(project.id, selectedLocation.value.id);
      } catch (error) {
        console.error('Error assigning project to location:', error);
      }
    }
    await projectStore.refreshProjects();
    await projectStore.refreshProjectsPreview();
    closeModal();
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification(t('notifications.errorCloningProject'), error);
  } finally {
    stage.operationActive = false;
    isCloning.value = false;
    isAwaitingResponse.value = false;
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
    `${loc.name}` === displayName
  );
  if (location) {
    selectedLocation.value = location;
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

.general-container {
  gap: 1rem;
}

.input-label {
  font-family: Inter, sans-serif;
  color: var(--white);
  white-space: nowrap;
  flex: 1;
}

.input-section {
  width: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  gap: .5rem;
  color: var(--white);
}

.settings-section-card{
  background-color: transparent;
  outline: 0px;
}

.location-dropdown-wrapper {
  flex: 1;
  width: 100%;
}
</style>



