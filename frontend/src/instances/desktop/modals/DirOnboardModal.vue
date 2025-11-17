<template>
  <div class="modal-container">
    <HeaderArea :title="'Setup Clustta'" :icon="getAppIcon('clustta')" :showSearch="false" />
    <div class="general-container">

      <!-- Card 1: Select data storage location -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <div class="header-content">
            <h2 class="settings-section-card-title">Select data location</h2>
            <div class="card-description">
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
      <div class="pop-up-actions">
        <GeneralButton 
          v-if="selectedClusttaDirectory" 
          label="Continue" 
          :buttonFunction="saveChanges"
          :fullWidth="false"
        />
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
import { ref, onMounted } from 'vue';

//stores
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';

//components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';

import { SettingsService, DialogService, FSService } from '@/../bindings/clustta/services/index';

//refs
const trayStates = useTrayStates();
const projectStore = useProjectStore();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();

const personalDataDirectory = ref('');
const sharedDataDirectory = ref('');

const defaultClusttaDirectory = ref('');
const selectedClusttaDirectory = ref('');

const userName = ref('');

// Location management refs
const locations = ref([]);
const locationHealthMap = ref({});
const locationUsageMap = ref({});


//methods

const selectDirectory = async () => {
  let title = 'Clustta Directory';
  let directory;
  let pathExists = false;

  try {
    pathExists = await FSService.DirExists(defaultClusttaDirectory.value);
  } catch (error) {
    pathExists = false;
  }
  
  directory = pathExists && !selectedClusttaDirectory.value ? defaultClusttaDirectory.value : selectedClusttaDirectory.value;

  const result = await DialogService.SelectSpecificFolderDialog(title, directory);

  if (result) {
    let fileDir = result.replace(/\\/g, '/');
    selectedClusttaDirectory.value = fileDir.replace(/\/clustta/g, '') + '/clustta';

    personalDataDirectory.value = selectedClusttaDirectory.value + '/projects';
    sharedDataDirectory.value = selectedClusttaDirectory.value + '/shared_projects';

    // Initialize default location if not already present
    if (locations.value.length === 0) {
      locations.value = [{
        id: '1',
        name: 'Default',
        path: selectedClusttaDirectory.value + '/mnt',
        is_default: true,
        project_ids: []
      }];
      await checkAllLocationHealth();
    }
  }
};

// Location management methods
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

const getProjectCount = (locationId) => {
  return locationUsageMap.value[locationId] || 0;
};

const canDeleteLocation = (locationId) => {
  if (locations.value.length <= 1) return false;
  return getProjectCount(locationId) === 0;
};

const addLocation = async () => {
  const newLocationName = `Location ${locations.value.length + 1}`;
  
  const result = await DialogService.SelectFolderDialog("Select Location Folder");
  if (!result) return;
  
  const path = result.replace(/\\/g, '/');
  
  // Add to local array (will be saved when user clicks Continue)
  const newLocation = {
    id: `${locations.value.length + 1}`,
    name: newLocationName,
    path: path,
    is_default: false,
    project_ids: []
  };
  
  locations.value.push(newLocation);
  await checkAllLocationHealth();
  
  notificationStore.addNotification('Location added', '', 'success', false);
};

const selectPath = async (location) => {
  const result = await DialogService.SelectFolderDialog("Select Location Folder");
  if (!result) return;
  
  const path = result.replace(/\\/g, '/');
  location.path = path;
  
  await checkAllLocationHealth();
  
  notificationStore.addNotification('Location updated', '', 'success', false);
};

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

const setDefaultLocation = (locationId) => {
  locations.value.forEach(loc => {
    loc.is_default = loc.id === locationId;
  });
  
  notificationStore.addNotification('Default location updated', '', 'success', false);
};

const saveChanges = async () => {
  try {
    // Save data directories
    await SettingsService.SetProjectDirectory(personalDataDirectory.value);
    await SettingsService.SetSharedProjectDirectory(sharedDataDirectory.value);
    
    // Save all project locations
    for (const location of locations.value) {
      try {
        // Check if location already exists
        const existingLocations = await SettingsService.GetAllLocationPaths();
        const exists = existingLocations.some(loc => loc.id === location.id);
        
        if (!exists) {
          const savedLocation = await SettingsService.AddProjectLocation(location.name, location.path);
          // Set as default if needed
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

const loadProjects = async () => {
      await projectStore.loadProjects();
      trayStates.refreshData();
};

const closeModal = () => {
  modals.disableAllModals();
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    saveChanges();
  }
};



onMounted(async () => {
  await SettingsService.GetUserDirectory()
    .then((response) => {
      defaultClusttaDirectory.value = `${response}clustta`;
      console.log(defaultClusttaDirectory.value)
    })
    .catch((error) => {
      notificationStore.addNotification(
        "Error Loading Settings",
        error.message,
        "error",
        false
      )
    });
  await SettingsService.GetUsername()
    .then((response) => {
      userName.value = response
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

.general-container {
  display: flex;
  flex-direction: column;
  /* gap: 1rem; */
  width: 700px;
  max-width: 700px;
  /* padding: 1rem; */
  color: white;
  box-sizing: border-box;
  /* background-color: crimson; */
  padding: 0;
}

/* Settings Section Card Styles */
.settings-section-card {
  display: flex;
  flex-direction: column;
  background-color: var(--dark-steel);
  /* border-radius: var(--normal-radius); */
  overflow: hidden;
  box-sizing: border-box;
  padding: 0;
  /* background-color: forestgreen; */
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

.header-content {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
}

.settings-section-card-title {
  font-size: 16px;
  font-weight: 400;
  color: var(--white);
  margin: 0;
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

.card-description {
  font-size: 13px;
  color: var(--silver);
  opacity: 0.9;
  line-height: 1.5;
}

/* Location Items */
.locations-scroll-container {
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  background-color: var(--midnight-steel);
  border-radius: var(--normal-radius);
  background-color: var(--light-steel);
  max-height: 300px;
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
  /* background-color: crimson; */
  background-color: var(--light-steel);
}

.location-item-single{
  border-radius: var(--normal-radius);
}

.location-item:last-child {
  border-bottom: none;
}

.location-item:hover {
  background-color: #ffffff15;
}

.location-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.3rem;
  box-sizing: border-box;
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

.location-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--white);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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

.default-badge {
  font-size: 11px;
  padding: 2px 8px;
  background-color: var(--accent-color);
  color: white;
  border-radius: 4px;
  font-weight: 500;
}

.location-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  box-sizing: border-box;
}

.hover-action {
  display: none;
}

.location-item:hover .hover-action {
  display: flex;
}

.pop-up-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  /* padding-top: 0.5rem; */
  /* background-color: forestgreen; */
}
</style>


