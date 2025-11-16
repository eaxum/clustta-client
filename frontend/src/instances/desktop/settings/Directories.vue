<template>
  <div class="settings-component-root">
    <div class="settings-component-container">
      
      <!-- Clustta Data Card -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">Clustta Data</h2>
        </div>
        <div class="settings-section-card-content">
          <!-- Local Projects -->
          <div class="location-item">
            <!-- Location Icon -->
            <div class="location-icon">
              <img class="small-icons" :src="getAppIcon('folder')">
            </div>
            
            <!-- Location Content -->
            <div class="location-content">
              <!-- Header -->
              <div class="location-header">
                <div class="location-name">Local projects</div>
              </div>
              
              <!-- Body (Path) -->
              <div class="location-body">
                {{ projectsDirectory }}
              </div>
            </div>
            
            <!-- Location Actions -->
            <div class="location-actions">
              <ActionButton 
                :icon="getAppIcon('explorer')" 
                :buttonFunction="() => selectDirectoryPath('personal')"
                v-tooltip="'Browse Path'"
              />
            </div>
          </div>

          <!-- Shared Projects -->
          <div class="location-item">
            <!-- Location Icon -->
            <div class="location-icon">
              <img class="small-icons" :src="getAppIcon('folder')">
            </div>
            
            <!-- Location Content -->
            <div class="location-content">
              <!-- Header -->
              <div class="location-header">
                <div class="location-name">Shared projects</div>
              </div>
              
              <!-- Body (Path) -->
              <div class="location-body">
                {{ sharedProjectsDirectory }}
              </div>
            </div>
            
            <!-- Location Actions -->
            <div class="location-actions">
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
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">Project folders</h2>
          <ActionButton 
            :icon="getAppIcon('plus-circle')" 
            label="Add Location" 
            :buttonFunction="addLocation"
            :showLabel="true"
          />
        </div>
        <div class="settings-section-card-content">
          <div class="locations-scroll-container">
            <div v-for="location in locations" :key="location.id" class="location-item">
              
              <!-- Location Icon - shows alert if path doesn't exist -->
              <div class="location-icon">
                <!-- <ActionButton 
                  v-if="locationHealthMap[location.id] && !locationHealthMap[location.id].exists"
                  :icon="getAppIcon('alert')" 
                  :useAlert="true"
                  :isInactive="true"
                  v-tooltip="'Path does not exist'"
                /> -->
                <img v-if="locationHealthMap[location.id] && !locationHealthMap[location.id].exists" class="small-icons" :src="getAppIcon('alert')">
                <img v-else class="small-icons" :src="getAppIcon('folder')">
              </div>
              
              <!-- Location Content -->
              <div class="location-content">
                <!-- Editing Mode - Header Only -->
                <div v-if="editingLocationId === location.id" class="location-header">
                  <RenameInput 
                    v-model="editableLocationName"
                    :originalValue="location.name"
                    placeholder="Location name"
                    @confirm="confirmEditLocationName(location)"
                    @cancel="cancelEditingLocation"
                  />
                </div>
                
                <!-- Normal Display Mode - Header -->
                <div v-else class="location-header">
                  <div class="location-name">{{ location.name }}</div>
                </div>
                
                <!-- Body (Path - always visible) -->
                <div class="location-body">
                  {{ location.path }}
                </div>
              </div>
              
              <!-- Location Actions -->
              <div class="location-actions">
                <!-- Star button - always visible for default location -->
                <ActionButton 
                  v-if="location.is_default"
                  :icon="getAppIcon('star')" 
                  :buttonFunction="() => setDefaultLocation(location.id)"
                  :disabled="true"
                  v-tooltip="'Default location'"
                />
                
                <!-- Other actions - visible on hover -->
                <template v-if="editingLocationId !== location.id">
                  <ActionButton 
                    :icon="getAppIcon('edit')" 
                    :buttonFunction="() => startEditingLocation(location)"
                    v-tooltip="'Edit name'"
                    class="hover-action"
                  />
                  <ActionButton 
                    v-if="!location.is_default"
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

/* Working Locations Section */
.locations-scroll-container {
  display: flex;
  flex-direction: column;
  /* gap: 0.5rem; */
  overflow-y: auto;
  /* padding-right: 0.25rem; */
  background-color: var(--dark-steel);
  border-radius: var(--normal-radius);
}

/* Location item - styled like settings-item */
.location-item {
  display: flex;
  /* border-radius: 8px; */
  align-items: center;
  /* background-color: var(--dark-steel); */
  overflow: hidden;
  box-sizing: border-box;
  min-height: 50px;
  height: max-content;
  padding: .5rem 1rem;
  gap: 0.5rem;
  cursor: pointer;
  transition: background-color 0.2s ease;
  border-bottom:  1px solid var(--light-steel);
}

.location-item:hover {
  background-color: #ffffff15;
}

.location-item:active {
  background-color: #00000013;
}

/* Location icon (optional) */
.location-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: .3rem;
  box-sizing: border-box;
}

/* Location content - like settings-content */
.location-content {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  flex: 1;
  overflow: hidden;
  padding: .2rem;
  box-sizing: border-box;
}

.location-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: .1rem;
}

.location-name {
  font-size: 14px;
  font-weight: 400;
  color: var(--white);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.location-body {
  color: var(--silver);
  font-size: 12px;
  opacity: .8;
  padding: .1rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  width: 100%;
}

/* Edit mode styling */
.rename-input {
  flex: 1;
  display: flex;
  align-items: center;
  width: 100%;
  box-sizing: border-box;
}

/* Location actions - like settings-action */
.location-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  box-sizing: border-box;
}

/* Hide hover actions by default, show on hover */
.hover-action {
  display: none;
}

.location-item:hover .hover-action {
  display: flex;
}

.disabled {
  opacity: 0.5;
  cursor: not-allowed !important;
  pointer-events: none;
}
</style>
