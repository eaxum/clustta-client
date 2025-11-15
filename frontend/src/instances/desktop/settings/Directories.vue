<template>
  <div class="settings-component-root">
    <div class="settings-component-container">
      
      <!-- Clustta Data Card -->
      <div class="directory-card">
        <div class="card-header">
          <h2 class="card-title">Clustta Data</h2>
        </div>
        <div class="card-content">
          <div class="location-item">
            <div class="location-header">
              <div class="location-name">Local projects</div>
            </div>

            <div class="actions-divider" ></div>
            
            <div class="horizontal-flex">
              <div class="location-path">{{ projectsDirectory }}</div>
              <ActionButton 
                :icon="getAppIcon('explorer')" 
                :buttonFunction="() => selectDirectoryPath('personal')"
                v-tooltip="'Browse Path'"
              />
            </div>
          </div>

          <div class="location-item">
            <div class="location-header">
              <div class="location-name">Shared projects</div>
            </div>

            <div class="actions-divider" ></div>

            <div class="horizontal-flex">
              <div class="location-path">{{ sharedProjectsDirectory }}</div>
              <ActionButton 
                :icon="getAppIcon('explorer')" 
                :buttonFunction="() => selectDirectoryPath('shared')"
                v-tooltip="'Browse Path'"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- Working Folder Locations Card -->
      <div class="directory-card">
        <div class="card-header">
          <h2 class="card-title">Project folders</h2>
          <ActionButton 
            :icon="getAppIcon('plus-circle')" 
            label="Add Location" 
            :buttonFunction="addLocation"
            :showLabel="true"
          />
        </div>
        <div class="card-content">
          <div class="locations-scroll-container">
            <div v-for="location in locations" :key="location.id" class="location-item">
            
            <!-- Editing Mode -->
            <div v-if="editingLocationId === location.id" class="location-header">
              <RenameInput 
                v-model="editableLocationName"
                :originalValue="location.name"
                placeholder="Location name"
                @confirm="confirmEditLocationName(location)"
                @cancel="cancelEditingLocation"
              />
            </div>
            
            <!-- Normal Display Mode -->
            <template v-else>
              <div class="location-header">
                <div class="location-name">
                  {{ location.name }}
                </div>
                <ActionButton 
                  v-if="locationHealthMap[location.id] && !locationHealthMap[location.id].exists"
                  :icon="getAppIcon('alert')" 
                  :useAlert="true"
                  :isInactive="true"
                  v-tooltip="'Path does not exist'"
                />
              </div>
              
              <div class="actions-divider" ></div>

              <div class="horizontal-flex">
                <div class="location-path">
                  {{ location.path }}
                </div>
                
                <!-- Star button - always visible for default location -->
                <ActionButton 
                  v-if="location.is_default"
                  :icon="getAppIcon('star')" 
                  :buttonFunction="() => setDefaultLocation(location.id)"
                  :disabled="true"
                  v-tooltip="'Default location'"
                />
                
                <!-- Actions - visible on hover -->
                <div class="location-actions">
                  <ActionButton 
                    :icon="getAppIcon('edit')" 
                    :buttonFunction="() => startEditingLocation(location)"
                    v-tooltip="'Edit name'"
                  />
                  <ActionButton 
                    v-if="!location.is_default"
                    :icon="getAppIcon('star')" 
                    :buttonFunction="() => setDefaultLocation(location.id)"
                    v-tooltip="'Set as default'"
                  />
                  <ActionButton 
                    :icon="getAppIcon('explorer')" 
                    :buttonFunction="() => selectPath(location)"
                    v-tooltip="'Change Location'"
                  />
                  <ActionButton 
                    :icon="getAppIcon('trash')" 
                    :buttonFunction="() => removeLocation(location.id)"
                    :isDisabled="!canDeleteLocation(location.id)"
                    v-tooltip="canDeleteLocation(location.id) ? 'Remove Location' : 'Cannot remove: projects are using this location'"
                  />
                </div>
              </div>
            </template>
            
          </div>
        </div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { SettingsService, DialogService } from '@/../bindings/clustta/services';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import RenameInput from '@/instances/desktop/components/RenameInput.vue';

const iconStore = useIconStore();
const notificationStore = useNotificationStore();

const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Refs
const projectsDirectory = ref('');
const sharedProjectsDirectory = ref('');
const locations = ref([]);
const locationHealthMap = ref({});
const locationUsageMap = ref({});
const editingLocationId = ref(null);
const editableLocationName = ref('');

// Methods for project/shared directories
const selectDirectoryPath = async (context) => {
  const result = await DialogService.SelectFolderDialog("Select Folder");
  if (!result) return;
  
  const fileDir = result.replace(/\\/g, '/');
  
  try {
    if (context === 'shared') {
      await SettingsService.SetSharedProjectDirectory(fileDir);
      sharedProjectsDirectory.value = fileDir;
      notificationStore.addNotification('Shared directory updated', '', 'success', false);
    } else if (context === 'personal') {
      await SettingsService.SetProjectDirectory(fileDir);
      projectsDirectory.value = fileDir;
      notificationStore.addNotification('Projects directory updated', '', 'success', false);
    }
  } catch (error) {
    notificationStore.errorNotification('Error updating directory', error);
  }
};

// Methods for working locations (merged from ProjectLocationManager)
const loadLocations = async () => {
  try {
    const fetchedLocations = await SettingsService.GetAllLocationPaths();
    locations.value = fetchedLocations;
  } catch (error) {
    console.error('Error loading locations:', error);
    locations.value = [];
  }
};

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
  
  try {
    const newLocation = await SettingsService.AddProjectLocation(newLocationName, path);
    locations.value.push(newLocation);
    
    await checkAllLocationHealth();
    await loadLocationUsage();
    
    notificationStore.addNotification('Location added successfully', '', 'success', false);
  } catch (error) {
    notificationStore.errorNotification('Error adding location', error);
  }
};

const selectPath = async (location) => {
  const result = await DialogService.SelectFolderDialog("Select Location Folder");
  if (!result) return;
  
  const path = result.replace(/\\/g, '/');
  
  try {
    await SettingsService.UpdateProjectLocation(location.id, location.name, path);
    location.path = path;
    
    await checkAllLocationHealth();
    
    notificationStore.addNotification('Location updated successfully', '', 'success', false);
  } catch (error) {
    notificationStore.errorNotification('Error updating location', error);
  }
};

const removeLocation = async (locationId) => {
  if (!canDeleteLocation(locationId)) {
    notificationStore.addNotification(
      'Cannot remove location',
      'Projects are using this location or it is the last location',
      'error',
      false
    );
    return;
  }
  
  try {
    await SettingsService.RemoveProjectLocation(locationId);
    const index = locations.value.findIndex(loc => loc.id === locationId);
    if (index !== -1) {
      locations.value.splice(index, 1);
    }
    
    notificationStore.addNotification('Location removed successfully', '', 'success', false);
  } catch (error) {
    notificationStore.errorNotification('Error removing location', error);
  }
};

const setDefaultLocation = async (locationId) => {
  try {
    await SettingsService.SetDefaultLocation(locationId);
    
    // Update local state to reflect the change
    locations.value.forEach(loc => {
      loc.is_default = loc.id === locationId;
    });
    
    notificationStore.addNotification('Default location updated', '', 'success', false);
  } catch (error) {
    notificationStore.errorNotification('Error setting default location', error);
  }
};

// Inline editing functions
const startEditingLocation = async (location) => {
  editingLocationId.value = location.id;
  editableLocationName.value = location.name;
  
  // Focus the input after it's rendered
  await nextTick();
  const input = document.querySelector('.rename-input .input-short');
  if (input) {
    input.focus();
    input.select();
  }
};

const cancelEditingLocation = () => {
  editableLocationName.value = '';
  editingLocationId.value = null;
};

const confirmEditLocationName = async (location) => {
  try {
    await SettingsService.UpdateProjectLocation(location.id, editableLocationName.value, location.path);
    location.name = editableLocationName.value;
    
    notificationStore.addNotification('Location name updated', '', 'success', false);
    cancelEditingLocation();
  } catch (error) {
    notificationStore.errorNotification('Error updating location name', error);
  }
};

// Load initial data
onMounted(async () => {
  try {
    projectsDirectory.value = await SettingsService.GetProjectDirectory();
    sharedProjectsDirectory.value = await SettingsService.GetSharedProjectDirectory();
    
    await loadLocations();
    await checkAllLocationHealth();
    await loadLocationUsage();
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

.settings-component-root {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 5px;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.settings-component-container {
    display: flex;
    flex-direction: column;
    box-sizing: border-box;
    height: 100%;
    overflow-y: auto;
    overflow-x: hidden;
    width: 96%;
    gap: 1.5rem;
    color: white;
    padding: 1rem;
    /* background-color: var(--black-steel); */
    border-radius: var(--large-radius);
}

.settings-component-container::-webkit-scrollbar {
  width: 6px;
}

.settings-component-container::-webkit-scrollbar-thumb {
  background-color: var(--midnight-steel);
  border-radius: 3px;
}

.settings-component-container::-webkit-scrollbar-track {
  background-color: var(--light-steel);
  border-radius: 3px;
}

/* ProfileCard-style cards */
.directory-card {
  background-color: var(--black-steel);
  border-radius: 24px;
  padding: 1.5rem;
  box-sizing: border-box;
  width: 100%;
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.card-header {
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: var(--transparent-line);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.card-title {
  font-size: 1rem;
  font-weight: 300;
  color: var(--white);
  margin: 0;
  flex: 1;
}

.card-content {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.actions-divider {
	display: flex;
	background-color: var(--light-steel);
	height: 16px;
	width: 1.5px;
}

.location-path {
  flex: 1;
  padding: 0.5rem;
  color: var(--text-secondary);
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  border-radius: 4px;
  /* background-color: crimson; */
  box-sizing: border-box;
}

/* Working Locations Section */
.locations-scroll-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  /* max-height: 150px; */
  overflow-y: auto;
  padding-right: 0.25rem;
}


.location-item {
  display: flex;
  /* flex-direction: column; */
  /* gap: 0.5rem; */
  border-radius: 8px;
  align-items: center;
  background-color: hotpink;
  background-color: var(--dark-steel);
  /* border-radius: var(--normal-radius); */
  overflow: hidden;
  box-sizing: border-box;
  height: 45px;
  min-height: 45px;
}

.location-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  /* background-color: forestgreen; */
  text-wrap: nowrap;
  max-width: 150px;
  width: 150px;
  text-overflow: ellipsis;
}

.location-item:has(.rename-input) .location-header {
  flex: 1;
  max-width: 100%;
  width: 100%;
}

.location-name {
  flex: 1;
  margin: 0;
  padding: 0.5rem;
  color: var(--white);
  font-size: 14px;
  font-weight: 400;
  font-family: Inter, sans-serif;
  /* background-color: royalblue; */
}

.rename-input {
  flex: 1;
  display: flex;
  align-items: center;
  width: 100%;
  padding: 0.5rem;
  box-sizing: border-box;
}

.location-actions {
  display: none;
  align-items: center;
  gap: 0.5rem;
}

.location-item:hover .location-actions {
  display: flex;
}

.location-info {
  font-size: 12px;
  color: var(--text-secondary);
  padding: 0.5rem;
  background-color: goldenrod;
}

.disabled {
  opacity: 0.5;
  cursor: not-allowed !important;
  pointer-events: none;
}
</style>
