<template>
  <div class="settings-component-root">
    <div class="settings-component-scroll">
    <div class="settings-component-container">
      <SettingsCard :title="$t('settings.clusttaData')" :items="systemDirectoryItems">
        <template #action="{ item }">
          <ActionButton
            :icon="getAppIcon('explorer')"
            :buttonFunction="() => selectDirectoryPath(item.context)"
            v-tooltip="$t('settings.browsePath')"
          />
        </template>
      </SettingsCard>

      <SettingsCard :title="$t('settings.projectFolders')" :items="locationItems">
        <template #header-action>
          <ActionButton
            :icon="getAppIcon('plus-circle')" 
            :label="$t('settings.addLocation')" 
            :buttonFunction="addLocation"
            :showLabel="true"
          />
        </template>

        <template #label="{ item }">
          <RenameInput
            v-if="editingLocationId === item.location.id"
            v-model="editableLocationName"
            :originalValue="item.location.name"
            :placeholder="$t('placeholders.locationName')"
            @confirm="confirmEditLocationName(item.location)"
            @cancel="cancelEditingLocation"
          />
          <template v-else>{{ item.label }}</template>
        </template>

        <template #action="{ item }">
          <ActionButton
            v-if="item.location.is_default && editingLocationId !== item.location.id"
            :icon="getAppIcon('star')"
            :buttonFunction="() => setDefaultLocation(item.location.id)"
            :isDisabled="true"
            v-tooltip="$t('settings.defaultLocation')"
          />
          <template v-if="editingLocationId !== item.location.id">
            <ActionButton
              :icon="getAppIcon('edit')"
              :buttonFunction="() => startEditingLocation(item.location)"
              v-tooltip="$t('settings.editName')"
            />
            <ActionButton
              v-if="!item.location.is_default"
              :icon="getAppIcon('star')"
              :buttonFunction="() => setDefaultLocation(item.location.id)"
              v-tooltip="$t('settings.setAsDefault')"
            />
            <ActionButton
              :icon="getAppIcon('explorer')"
              :buttonFunction="() => selectPath(item.location)"
              v-tooltip="$t('settings.changeLocation')"
            />
            <ActionButton
              :icon="getAppIcon('trash')"
              :buttonFunction="() => removeLocation(item.location.id)"
              :isDisabled="!canDeleteLocation(item.location.id)"
              v-tooltip="canDeleteLocation(item.location.id) ? $t('settings.removeLocation') : $t('settings.cannotRemoveLocation')"
            />
          </template>
        </template>
      </SettingsCard>

    </div>
  </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useAccountStore } from '@/stores/accounts';
import { useSettingsStore } from '@/stores/settings';
import { SettingsService, DialogService } from '@/services';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import RenameInput from '@/instances/desktop/components/RenameInput.vue';
import SettingsCard from '@/instances/desktop/settings/components/SettingsCard.vue';

const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const accountStore = useAccountStore();
const settingsStore = useSettingsStore();
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
const getLocationHealthTooltip = (health) => {
  if (health?.stale) return t('settings.locationStale');
  if (health && !health.exists) return t('settings.locationMissing');
  return '';
};
const systemDirectoryItems = computed(() => {
  const items = [
    {
      id: 'local-projects',
      context: 'personal',
      icon: getAppIcon(settingsStore.systemBookmarksHealth?.projects_dir_stale ? 'alert' : 'folder'),
      iconTooltip: settingsStore.systemBookmarksHealth?.projects_dir_stale
        ? t('settings.locationStale')
        : '',
      label: t('settings.localProjects'),
      description: projectsDirectory.value,
    },
  ];
  if (!accountStore.isOfflineMode) {
    items.push({
      id: 'shared-projects',
      context: 'shared',
      icon: getAppIcon(settingsStore.systemBookmarksHealth?.shared_projects_dir_stale ? 'alert' : 'folder'),
      iconTooltip: settingsStore.systemBookmarksHealth?.shared_projects_dir_stale
        ? t('settings.locationStale')
        : '',
      label: t('settings.sharedProjects'),
      description: sharedProjectsDirectory.value,
    });
  }
  return items;
});
const locationItems = computed(() => {
  return locations.value.map((location) => {
    const health = locationHealthMap.value[location.id];
    return {
      id: location.id,
      icon: getAppIcon(health && (!health.exists || health.stale) ? 'alert' : 'folder'),
      iconTooltip: getLocationHealthTooltip(health),
      label: location.name,
      description: location.path,
      location,
    };
  });
});

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
    await checkAllLocationHealth();
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
    const systemHealth = await SettingsService.CheckSystemBookmarksHealth();
    settingsStore.systemBookmarksHealth = systemHealth;
    const systemStale = systemHealth.projects_dir_stale || systemHealth.shared_projects_dir_stale;
    settingsStore.locationsStale = healthStatuses.some(h => h.stale) || systemStale;
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
  flex-direction: column;
  gap: 5px;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  display: block;
  overflow-y: auto;
}


.settings-component-root::-webkit-scrollbar {
  width: 6px;
}

.settings-component-root::-webkit-scrollbar-thumb {
  background-color: var(--bg);
  border-radius: 3px;
}

.settings-component-root::-webkit-scrollbar-track {
  background-color: var(--surface-4);
  border-radius: 3px;
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
  gap: 1.5rem;
  padding-right: .2rem;
  border-radius: var(--large-radius);
}


.rename-input {
  flex: 1;
  display: flex;
  align-items: center;
  width: 100%;
  box-sizing: border-box;
}

.disabled {
  opacity: 0.5;
  cursor: not-allowed !important;
  pointer-events: none;
}
</style>
