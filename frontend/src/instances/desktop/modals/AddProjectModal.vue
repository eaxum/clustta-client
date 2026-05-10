<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>

    <HeaderArea :title="title" :icon="getAppIcon('briefcase-plus')" :showSearch="false" />

    <div class="general-container">

      <!-- Clone Progress Display -->
      <div v-if="isCloning" class="settings-section-card">
        <ProgressSection variant="success" />
      </div>

      <template v-else>
      <div class="input-section">
        <div class="horizontal-flex">
          <input v-model="projectName" class="input-short" type="text" :placeholder="$t('placeholders.projectName')" ref="projectNameInput"
            @keydown.enter="handleEnterKey" v-focus />
          <span @click="addExistingFolder" class="single-action-button" v-tooltip="$t('modals.addExistingFolder')">
            <img class="small-icons" :src="getAppIcon('folder-arrow-in')">
          </span>
        </div>
        <InputAlert :show="!projectIsCreated && projectNameInUse" :message="$t('modals.projectNameExists')" />
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
          <span @click="addNewLocation" class="single-action-button" v-tooltip="$t('modals.addNewLocation')">
            <img class="small-icons" :src="getAppIcon('plus-circle')">
          </span>
        </div>
      </div>

      <div v-if="projectTemplateStore.projectTemplates.length" class="input-section drop-down-box-section">
        <DropDownBox :items="projectTemplateNames" :selectedItem="selectedProjectTemplate"
          :onSelect="selectProjectTemplate" />
      </div>

      <div v-if="showRemoteToggle" class="input-section" @click="makeRemote = !makeRemote">
        <div class="horizontal-flex toggle-row">
          <span class="input-label">{{ $t('modals.enableRemote') }}</span>
          <ToggleSwitch :switchValueProp="makeRemote" :online="makeRemote" />
        </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :isActive="!isAwaitingResponse"  :colored="false" />
        <GeneralButton :label="$t('common.create')" :fullWidth="true" @click="createProject" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>
      </template>
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
import ProgressSection from '@/instances/common/components/ProgressSection.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// services
import { DialogService, ProjectService, SettingsService, SyncService } from '@/services';

// stores
const accountStore = useAccountStore();
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const projectTemplateStore = useProjectTemplateStore();
const entitlementStore = useEntitlementStore();
const stage = useStageStore();
const { t } = useI18n();

import { useAccountStore } from '@/stores/accounts';
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
import { useEntitlementStore } from '@/stores/entitlements';
import { refreshEntitlements } from '@/lib/sync';

// refs
const existingFolderPath = ref(null);
const isAwaitingResponse = ref(false);
const isCloning = ref(false);
const isLoadingLocations = ref(false);
const makeRemote = ref(false);
const modalContainer = ref(null);
const projectIsCreated = ref(false);
const projectLocations = ref([]);
const projectName = ref('');
const projectNameInput = ref(null);
const selectedLocation = ref(null);
const selectedProjectTemplate = ref(t('modals.noTemplate'));

// constants
const title = computed(() => t('modals.addProject'));

// computed
// Returns whether form values are valid for submission.
const isValueChanged = computed(() => {
  return !projectNameEmpty.value && !projectNameInUse.value;
});

// Returns display names for location dropdown.
const locationDisplayNames = computed(() => {
  return projectLocations.value.map(loc => `${loc.name}`);
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
  return [t('modals.noTemplate'), ...projectTemplateStore.projectTemplateNames];
});

// Returns list of restricted project names (lowercase).
const restrictedNames = computed(() => {
  return projectStore.projects.map((project) => project.name.toLowerCase());
});

// Returns display string for currently selected location.
const selectedLocationDisplay = computed(() => {
  if (!selectedLocation.value) return '';
  return `${selectedLocation.value.name}`;
});

// Returns whether the remote toggle should be shown.
const showRemoteToggle = computed(() => {
  return accountStore.canUseRemoteFeatures && projectStore.selectedStudio?.name === 'Personal' && entitlementStore.canCreateRemoteProject;
});

// Returns the computed working directory path.
const workingDirectory = computed(() => {
  if (existingFolderPath.value) return existingFolderPath.value;
  if (!selectedLocation.value || !projectName.value) return '';
  const studioName = projectStore.selectedStudio.name;
  if (studioName === 'Personal') {
    return `${selectedLocation.value.path}/${projectName.value}`;
  }
  return `${selectedLocation.value.path}/${studioName}/${projectName.value}`;
});

// methods
// Opens a folder dialog to select an existing project folder.
const addExistingFolder = async () => {
  const userDirectory = await SettingsService.GetUserDirectory();
  const documentsPath = userDirectory + 'Documents';
  const result = await DialogService.SelectSpecificFolderDialog("Select Existing Project Folder", documentsPath);
  if (!result) return;
  const path = result.replace(/\\/g, '/');
  const pathParts = path.split('/');
  const folderName = pathParts[pathParts.length - 1] || 'Project';
  const parentPath = pathParts.slice(0, -1).join('/');
  const parentName = pathParts[pathParts.length - 2] || `Location ${projectLocations.value.length + 1}`;
  existingFolderPath.value = path;
  projectName.value = folderName;
  const existingLocation = projectLocations.value.find(loc => loc.path === parentPath);
  if (existingLocation) {
    selectedLocation.value = existingLocation;
  } else {
    try {
      const newLocation = await SettingsService.AddProjectLocation(parentName, parentPath);
      projectLocations.value.push(newLocation);
      selectedLocation.value = newLocation;
    } catch (error) {
      notificationStore.errorNotification(t('notifications.errorAddingLocation'), error);
    }
  }
};

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
    notificationStore.errorNotification(t('notifications.errorLoadingLocations'), error);
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

// Selects a project template from the dropdown.
const selectProjectTemplate = (selectedTemplateName) => {
  selectedProjectTemplate.value = selectedTemplateName;
};

const createProject = async () => {

  // Validate that a location is selected
  if (!selectedLocation.value) {
    notificationStore.addNotification(
      t('notifications.noLocationSelected'),
      t('notifications.selectOrAddLocation'),
      'error',
      false
    );
    return;
  }

  // Validate working directory is not empty
  if (!workingDirectory.value) {
    notificationStore.addNotification(
      t('notifications.invalidWorkingDirectory'),
      t('notifications.workingDirectoryEmpty'),
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
  
  const templateName = selectedProjectTemplate.value === t('modals.noTemplate') ? 'No Template' : selectedProjectTemplate.value;

  ProjectService.CreateProject(path, studio.name, workingDirectory.value, templateName, studio.hosting_mode || '', studio.id || '').then(async (project) => {

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

    if (studio.name !== 'Personal') {
      // Studio projects need the full reload so the in-list entry has the local .clst URI
      // (CreateProject returns the remote URL as `uri`). cloneProject relies on this to
      // pick up the correct local path when it finishes.
      await projectStore.loadProjects();
      projectStore.activeProject = projectStore.projects.find(p => p.id === project.id) || project;
      await cloneProject();
    } else {
      project.is_tracked = true;
      project.is_downloaded = true;
      projectStore.addProjectToList(project);
      projectStore.activeProject = project;

      if (makeRemote.value) {
        await makeProjectRemote(project);
      } else {
        closeModal();
      }
    }

    isAwaitingResponse.value = false;

  }).catch((error) => {
    isAwaitingResponse.value = false
    console.log(error)
    notificationStore.errorNotification(t('notifications.errorCreatingProject'), error);
  });

};

// Clones the project from server after creation.
const cloneProject = async () => {
  isCloning.value = true;
  stage.operationActive = true;
  const project = projectStore.activeProject;
  const studioDisplayName = projectStore.selectedStudio.name;
  const projectName = project.name;
  const projectUrl = project.remote || (projectStore.getStudioUrl + '/' + projectName);
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
    await projectStore.refreshProjects();
    const updatedProject = projectStore.projects.find(p => p.name === projectName);
    if (updatedProject) {
      projectStore.activeProject = updatedProject;
    }
    if (selectedProjectTemplate.value && selectedProjectTemplate.value !== t('modals.noTemplate')) {
      const localProjectPath = projectStore.activeProject.uri;
      try {
        await ProjectService.ApplyTemplate(localProjectPath, selectedProjectTemplate.value);
        const templateSyncOptions = {
          only_latest_checkpoints: false,
          asset_dependencies: false,
          assets: false,
          templates: false,
        };
        try {
          await SyncService.SyncData(localProjectPath, projectUrl, false, templateSyncOptions);
        } catch (error) {
          console.error('Failed to sync template changes:', error);
          notificationStore.addNotification(t('notifications.templateAppliedSyncFailed'), 'warning');
        }
      } catch (error) {
        console.error('Failed to apply template:', error);
        notificationStore.errorNotification(t('notifications.failedToApplyTemplate'), error);
      }
    }
    await projectStore.refreshProjectsPreview();
    refreshEntitlements();
    closeModal();
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification(t('notifications.errorCloningProject'), error);
  } finally {
    stage.operationActive = false;
    isCloning.value = false;
  }
};

// Makes a newly created project remote by uploading to server.
const makeProjectRemote = async (project) => {
  isCloning.value = true;
  stage.operationActive = true;
  try {
    await ProjectService.MakeProjectRemote(project.uri);
    const updatedInfo = await ProjectService.ProjectInfo(project.uri);
    await projectStore.refreshProjects();
    const updatedProject = projectStore.projects.find(p => p.name === project.name);
    if (updatedProject) {
      updatedProject.remote = updatedInfo.remote;
      updatedProject.has_remote = updatedInfo.has_remote;
      projectStore.activeProject = updatedProject;
    }
    closeModal();
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification(t('notifications.errorMakingProjectRemote'), error);
  } finally {
    stage.operationActive = false;
    isCloning.value = false;
    refreshEntitlements();
  }
};

// Resets project data in stores after creation.
const resetProjectData = () => {
  commonStore.activeWorkspace = 'Default';
  commonStore.resetFilters();
  collectionStore.collections = [];
  assetStore.assets = [];
  collectionStore.selectedCollection = null;
  assetStore.selectedAsset = null;
  stage.expandedCollections = {};
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

.settings-section-card{
  background-color: transparent;
  outline: 0px;
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

.input-short {
  flex: 1;
  width: 100%;
}

.location-dropdown-wrapper {
  flex: 1;
  width: 100%;
}

.toggle-row {
  cursor: pointer;
  align-items: center;
  justify-content: space-between;
}

</style>






