<template>
  <div class="settings-component-root">
    <div class="settings-component-scroll">
    <div class="settings-component-container">
      
      <!-- Clustta Data Card -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.clusttaData') }}</h2>
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
                <div class="location-name">{{ $t('settings.localProjects') }}</div>
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
                v-tooltip="$t('settings.browsePath')"
              />
            </div>
          </div>

          <!-- Shared Projects -->
          <div class="location-item" v-if="!accountStore.isOfflineMode">
            <!-- Location Icon -->
            <div class="location-icon">
              <img class="small-icons" :src="getAppIcon('folder')">
            </div>
            
            <!-- Location Content -->
            <div class="location-content">
              <!-- Header -->
              <div class="location-header">
                <div class="location-name">{{ $t('settings.sharedProjects') }}</div>
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
                v-tooltip="$t('settings.browsePath')"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- Working Folder Locations Card -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.projectFolders') }}</h2>
          <ActionButton 
            :icon="getAppIcon('plus-circle')" 
            :label="$t('settings.addLocation')" 
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
                    :placeholder="$t('placeholders.locationName')"
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
                  v-tooltip="$t('settings.defaultLocation')"
                />
                
                <!-- Other actions - visible on hover -->
                <template v-if="editingLocationId !== location.id">
                  <ActionButton 
                    :icon="getAppIcon('edit')" 
                    :buttonFunction="() => startEditingLocation(location)"
                    v-tooltip="$t('settings.editName')"
                    class="hover-action"
                  />
                  <ActionButton 
                    v-if="!location.is_default"
                    :icon="getAppIcon('star')" 
                    :buttonFunction="() => setDefaultLocation(location.id)"
                    v-tooltip="$t('settings.setAsDefault')"
                    class="hover-action"
                  />
                  <ActionButton 
                    :icon="getAppIcon('explorer')" 
                    :buttonFunction="() => selectPath(location)"
                    v-tooltip="$t('settings.changeLocation')"
                    class="hover-action"
                  />
                  <ActionButton 
                    :icon="getAppIcon('trash')" 
                    :buttonFunction="() => removeLocation(location.id)"
                    :isDisabled="!canDeleteLocation(location.id)"
                    v-tooltip="canDeleteLocation(location.id) ? $t('settings.removeLocation') : $t('settings.cannotRemoveLocation')"
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
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useAccountStore } from '@/stores/accounts';
import { SettingsService, DialogService } from '@/services';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import RenameInput from '@/instances/desktop/components/RenameInput.vue';

const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const accountStore = useAccountStore();
const { t } = useI18n();

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
  const result = await DialogService.SelectFolderDialog(t('settings.selectFolder'));
  if (!result) return;
  
  const fileDir = result.replace(/\\/g, '/');
  
  try {
    if (context === 'shared') {
      await SettingsService.SetSharedProjectDirectory(fileDir);
      sharedProjectsDirectory.value = fileDir;
      notificationStore.addNotification(t('notifications.sharedDirectoryUpdated'), '', 'success', false);
    } else if (context === 'personal') {
      await SettingsService.SetProjectDirectory(fileDir);
      projectsDirectory.value = fileDir;
      notificationStore.addNotification(t('notifications.projectsDirectoryUpdated'), '', 'success', false);
    }
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorUpdatingDirectory'), error);
  }
};

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
  
  const result = await DialogService.SelectFolderDialog(t('settings.selectFolder'));
  if (!result) return;
  
  const path = result.replace(/\\/g, '/');
  
  try {
    const newLocation = await SettingsService.AddProjectLocation(newLocationName, path);
    locations.value.push(newLocation);
    
    await checkAllLocationHealth();
    await loadLocationUsage();
    
    notificationStore.addNotification(t('notifications.locationAdded'), '', 'success', false);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorAddingLocation'), error);
  }
};

const selectPath = async (location) => {
  const result = await DialogService.SelectFolderDialog(t('settings.selectFolder'));
  if (!result) return;
  
  const path = result.replace(/\\/g, '/');
  
  try {
    await SettingsService.UpdateProjectLocation(location.id, location.name, path);
    location.path = path;
    
    await checkAllLocationHealth();
    
    notificationStore.addNotification(t('notifications.locationUpdated'), '', 'success', false);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorUpdatingLocation'), error);
  }
};

const removeLocation = async (locationId) => {
  if (!canDeleteLocation(locationId)) {
    notificationStore.addNotification(
      t('notifications.cannotRemoveLocation'),
      t('notifications.locationInUse'),
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
    
    notificationStore.addNotification(t('notifications.locationRemoved'), '', 'success', false);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorRemovingLocation'), error);
  }
};

const setDefaultLocation = async (locationId) => {
  try {
    await SettingsService.SetDefaultLocation(locationId);
    
    // Update local state to reflect the change
    locations.value.forEach(loc => {
      loc.is_default = loc.id === locationId;
    });
    
    notificationStore.addNotification(t('notifications.defaultLocationUpdated'), '', 'success', false);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorSettingDefaultLocation'), error);
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
    
    notificationStore.addNotification(t('notifications.locationNameUpdated'), '', 'success', false);
    cancelEditingLocation();
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorUpdatingLocationName'), error);
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
      t('notifications.errorLoadingSettings'),
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
  box-sizing: border-box;
  display: block;
  overflow-y: auto;
}


.settings-component-root::-webkit-scrollbar {
  width: 4px;
}

.settings-component-root::-webkit-scrollbar-thumb {
  background-color: hsl(var(--border));
  border-radius: var(--small-radius);
}

.settings-component-root::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.settings-component-scroll {
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.settings-component-container {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  height: 100%;
  width: 100%;
  gap: 1rem;
  padding-right: .2rem;
}


/* Working Locations Section */
.locations-scroll-container {
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  border-radius: calc(var(--radius) - 2px);
}

/* Location item - styled like settings-item */
.location-item {
  display: flex;
  align-items: center;
  overflow: hidden;
  box-sizing: border-box;
  min-height: 44px;
  height: max-content;
  padding: .5rem 0.75rem;
  gap: 0.5rem;
  cursor: pointer;
  transition: background-color 0.15s ease;
  border-radius: calc(var(--radius) - 2px);
}

.location-item:hover {
  background-color: hsl(var(--accent));
}

.location-item:active {
  background-color: hsl(var(--accent));
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
  color: hsl(var(--foreground));
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.location-body {
  color: hsl(var(--muted-foreground));
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
