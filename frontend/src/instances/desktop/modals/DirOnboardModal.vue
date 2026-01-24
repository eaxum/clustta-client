<template>
  <div class="modal-container" v-stop-propagation>
    <HeaderArea :title="'Setup Clustta'" :icon="getAppIcon('clustta')" :showSearch="false" />
    <div class="general-container">

      <!-- Card 1: Select data storage location -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <div class="header-content">
            <h2 class="settings-section-card-title">Grant permission</h2>
            <div class="card-description">
              Grant clustta permission to save data to your computer.
              This is where Clustta will store data for your project checkpoints as well as shared projects.
            </div>
          </div>
          <GeneralButton 
            :label="selectedClusttaDirectory ? 'Change' : 'Select'" 
            :buttonFunction="selectDirectory"
            :fullWidth="false"
          />
        </div>
        <div class="settings-section-card-content" :class="{ 'no-padding': !selectedClusttaDirectory }">
          
          <!-- Selected Path Display -->
          <div v-if="selectedClusttaDirectory" class="location-item location-item-single">
            <div class="location-icon">
              <img class="small-icons" :src="getAppIcon('folder')">
            </div>
            <div class="location-content">
              <div class="location-header">
                <div class="location-name">Clustta Data</div>
              </div>
              <div class="location-body">
                {{ selectedClusttaDirectory }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Card 2: Project folders (only visible after selecting storage location) -->
      <div v-if="selectedClusttaDirectory" class="settings-section-card">
        <div class="settings-section-card-header">
          <div class="header-content">
            <h2 class="settings-section-card-title">Project folders</h2>
            <div class="card-description">
              Add any folders where you currently save your project files. E.g 'Documents', 'Projects' or even an external drive likle E://Projects.
            </div>
          </div>
          <GeneralButton 
            label="Add" 
            :buttonFunction="addLocation"
            :fullWidth="false"
          />
        </div>
        <div class="settings-section-card-content">
          
          <!-- Locations List -->
          <div class="locations-scroll-container">
            <div v-for="location in locations" :key="location.id" class="location-item">
              <!-- Location Icon -->
              <div class="location-icon">
                <img v-if="locationHealthMap[location.id] && !locationHealthMap[location.id].exists" 
                     class="small-icons" 
                     :src="getAppIcon('alert')">
                <img v-else class="small-icons" :src="getAppIcon('folder')">
              </div>
              
              <!-- Location Content -->
              <div class="location-content">
                <div class="location-header">
                  <div class="location-name">{{ location.name }}</div>
                </div>
                <div class="location-body">
                  {{ location.path }}
                </div>
              </div>
              
              <!-- Location Actions -->
              <div class="location-actions">
                <ActionButton 
                  v-if="location.is_default"
                  :icon="getAppIcon('star')" 
                  :buttonFunction="() => setDefaultLocation(location.id)"
                  :disabled="true"
                  v-tooltip="'Default location'"
                />
                
                <template v-else>
                  <ActionButton 
                    :icon="getAppIcon('star')" 
                    :buttonFunction="() => setDefaultLocation(location.id)"
                    v-tooltip="'Set as default'"
                    class="hover-action"
                  />
                  <ActionButton 
                    :icon="getAppIcon('explorer')" 
                    :buttonFunction="() => selectPath(location)"
                    v-tooltip="'Change Location'"
                    class="hover-action"
                  />
                  <ActionButton 
                    :icon="getAppIcon('trash')" 
                    :buttonFunction="() => removeLocation(location.id)"
                    :isDisabled="!canDeleteLocation(location.id)"
                    v-tooltip="canDeleteLocation(location.id) ? 'Remove Location' : 'Cannot remove: projects are using this location'"
                    class="hover-action"
                  />
                </template>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Action Buttons -->
      <div v-if="selectedClusttaDirectory" class="pop-up-actions">
        <GeneralButton 
          label="Continue" 
          :buttonFunction="saveChanges"
          :fullWidth="false"
          :isActive="hasDefaultLocation"
        />
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { DialogService, FSService, SettingsService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';

const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const trayStates = useTrayStates();

// refs
const defaultClusttaDirectory = ref('');
const locationHealthMap = ref({});
const locations = ref([]);
const locationUsageMap = ref({});
const personalDataDirectory = ref('');
const selectedClusttaDirectory = ref('');
const sharedDataDirectory = ref('');
const userBaseDirectory = ref('');
const userName = ref('');

// computed
// Returns true if at least one location is set as default.
const hasDefaultLocation = computed(() => {
  return locations.value.some(loc => loc.is_default === true);
});

// methods
// Adds a new project location via folder dialog.
const addLocation = async () => {
  const documentsPath = userBaseDirectory.value + 'Documents';
  const result = await DialogService.SelectSpecificFolderDialog("Select Location Folder", documentsPath);
  if (!result) return;
  
  const path = result.replace(/\\/g, '/');
  const pathParts = path.split('/');
  const folderName = pathParts[pathParts.length - 1] || `Location ${locations.value.length + 1}`;
  
  const newLocation = {
    id: `${locations.value.length + 1}`,
    name: folderName,
    path: path,
    is_default: locations.value.length === 0,
    project_ids: []
  };
  
  locations.value.push(newLocation);
  await checkAllLocationHealth();
  
  notificationStore.addNotification('Location added', '', 'success', false);
};

// Determines if a location can be deleted.
const canDeleteLocation = (locationId) => {
  if (locations.value.length <= 1) return false;
  return getProjectCount(locationId) === 0;
};

// Checks health status of all locations.
const checkAllLocationHealth = async () => {
  try {
    const healthStatuses = await SettingsService.CheckAllLocationsHealth();
    healthStatuses.forEach(h => {
      locationHealthMap.value[h.id] = h;
    });
  } catch (error) {
    console.error('Error checking location health:', error);
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

// Returns the project count for a location.
const getProjectCount = (locationId) => {
  return locationUsageMap.value[locationId] || 0;
};

// Loads project usage for all locations.
const loadLocationUsage = async () => {
  for (const location of locations.value) {
    try {
      const count = await SettingsService.GetLocationUsage(location.id);
      locationUsageMap.value[location.id] = count;
    } catch (error) {
      locationUsageMap.value[location.id] = 0;
    }
  }
};

// Loads projects and refreshes tray states.
const loadProjects = async () => {
  await projectStore.loadProjects();
  trayStates.refreshData();
};

// Removes a location from the list.
const removeLocation = (locationId) => {
  if (!canDeleteLocation(locationId)) {
    notificationStore.addNotification(
      'Cannot remove location',
      'Projects are using this location or it is the last location',
      'error',
      false
    );
    return;
  }
  
  const index = locations.value.findIndex(loc => loc.id === locationId);
  if (index !== -1) {
    locations.value.splice(index, 1);
    notificationStore.addNotification('Location removed', '', 'success', false);
  }
};

// Saves all changes and closes the modal.
const saveChanges = async () => {
  try {
    await SettingsService.SetProjectDirectory(personalDataDirectory.value);
    await SettingsService.SetSharedProjectDirectory(sharedDataDirectory.value);
    
    await projectStore.loadStudios();
    
    for (const location of locations.value) {
      try {
        const existingLocations = await SettingsService.GetAllLocationPaths();
        const exists = existingLocations.some(loc => loc.id === location.id);
        
        if (!exists) {
          const savedLocation = await SettingsService.AddProjectLocation(location.name, location.path);
          if (location.is_default) {
            await SettingsService.SetDefaultLocation(savedLocation.id);
          }
        }
      } catch (error) {
        console.error('Error saving location:', error);
      }
    }
    
    await loadProjects();
    closeModal();
  } catch (error) {
    notificationStore.errorNotification('Error saving settings', error);
  }
};

// Opens dialog to select the Clustta directory.
const selectDirectory = async () => {
  let title = 'Clustta Directory';
  let directory = userBaseDirectory.value;

  const result = await DialogService.SelectSpecificFolderDialog(title, directory);

  if (result) {
    let fileDir = result.replace(/\\/g, '/');
    selectedClusttaDirectory.value = fileDir.replace(/\/clustta/g, '') + '/clustta';

    personalDataDirectory.value = selectedClusttaDirectory.value + '/projects';
    sharedDataDirectory.value = selectedClusttaDirectory.value + '/shared_projects';

    if (locations.value.length === 0) {
      const mntPath = selectedClusttaDirectory.value + '/mnt';
      try {
        const mntExists = await FSService.DirExists(mntPath);
        if (mntExists) {
          locations.value = [{
            id: '1',
            name: 'mnt',
            path: mntPath,
            is_default: true,
            project_ids: []
          }];
          await checkAllLocationHealth();
        }
      } catch (error) {
        console.error('Error checking /mnt folder:', error);
      }
    }
  }
};

// Opens dialog to change the path of an existing location.
const selectPath = async (location) => {
  const result = await DialogService.SelectFolderDialog("Select Location Folder");
  if (!result) return;
  
  const path = result.replace(/\\/g, '/');
  location.path = path;
  
  await checkAllLocationHealth();
  
  notificationStore.addNotification('Location updated', '', 'success', false);
};

// Sets a location as the default.
const setDefaultLocation = (locationId) => {
  locations.value.forEach(loc => {
    loc.is_default = loc.id === locationId;
  });
  
  notificationStore.addNotification('Default location updated', '', 'success', false);
};

// lifecycle
onMounted(async () => {
  try {
    const response = await SettingsService.GetUserDirectory();
    userBaseDirectory.value = response;
    defaultClusttaDirectory.value = `${response}clustta`;
  } catch (error) {
    notificationStore.addNotification(
      "Error Loading Settings",
      error.message,
      "error",
      false
    );
  }
  
  try {
    const response = await SettingsService.GetUsername();
    userName.value = response;
  } catch (error) {
    notificationStore.addNotification(
      "Error Loading Settings",
      error.message,
      "error",
      false
    );
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.card-description {
  font-size: 13px;
  color: var(--silver);
  opacity: 0.9;
  line-height: 1.5;
}

.general-container {
  display: flex;
  flex-direction: column;
  width: 600px;
  max-width: 600px;
  color: white;
  box-sizing: border-box;
  padding: 0;
}

.header-content {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
}

.hover-action {
  display: none;
}

.location-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  box-sizing: border-box;
}

.location-body {
  color: var(--silver);
  font-size: 12px;
  opacity: 0.8;
  padding: 0.1rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  width: 100%;
}

.location-content {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  flex: 1;
  overflow: hidden;
  padding: 0.2rem;
  box-sizing: border-box;
}

.location-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.1rem;
}

.location-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.3rem;
  box-sizing: border-box;
}

.location-item {
  display: flex;
  align-items: center;
  overflow: hidden;
  box-sizing: border-box;
  min-height: 60px;
  height: max-content;
  padding: 0.75rem 1rem;
  gap: 0.75rem;
  cursor: pointer;
  transition: background-color 0.2s ease;
  border-bottom: 1px solid var(--dark-steel);
  background-color: var(--light-steel);
}

.location-item:hover {
  background-color: #ffffff15;
}

.location-item:hover .hover-action {
  display: flex;
}

.location-item:last-child {
  border-bottom: none;
}

.location-item-single {
  border-radius: var(--normal-radius);
}

.location-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--white);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.locations-scroll-container {
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  border-radius: var(--normal-radius);
  background-color: var(--light-steel);
  max-height: 300px;
}

.pop-up-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.settings-section-card {
  display: flex;
  flex-direction: column;
  background-color: var(--dark-steel);
  overflow: hidden;
  box-sizing: border-box;
  padding: 0;
}

.settings-section-card-content {
  display: flex;
  flex-direction: column;
  padding: 1rem;
  gap: 1rem;
}

.settings-section-card-content.no-padding {
  padding: 0;
}

.settings-section-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  background-color: var(--midnight-steel);
  border-radius: var(--normal-radius);
  margin: 0;
}

.settings-section-card-title {
  font-size: 16px;
  font-weight: 400;
  color: var(--white);
  margin: 0;
}
</style>


