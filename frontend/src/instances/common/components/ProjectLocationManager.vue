<template>
  <div class="location-manager">
    <div class="header-row">
      <span class="section-label">Working Folder Locations</span>
      <ActionButton 
        :icon="getAppIcon('plus-circle')" 
        label="Add Location" 
        :buttonFunction="addLocation"
        :showLabel="true"
      />
    </div>
    
    <div class="locations-scroll-container">
      <div v-for="location in locations" :key="location.id" class="location-item">
      <div class="location-header">
        <h3 class="location-name">
          {{ location.name }}
        </h3>
        <span v-if="location.is_default" class="default-badge">Default</span>
        <span 
          v-if="locationHealthMap[location.id] && !locationHealthMap[location.id].exists" 
          class="warning-badge" 
          v-tooltip="'Path does not exist'"
        >⚠️</span>
      </div>
      
      <div class="horizontal-flex">
        <span class="location-path">
          {{ location.path }}
        </span>
        <span 
          @click="selectPath(location)" 
          class="single-action-button" 
          v-tooltip="'Browse Path'"
        >
          <img class="small-icons" :src="getAppIcon('explorer')">
        </span>
        <span 
          @click="removeLocation(location.id)" 
          class="single-action-button" 
          :class="{ disabled: !canDeleteLocation(location.id) }"
          v-tooltip="canDeleteLocation(location.id) ? 'Remove Location' : 'Cannot remove: projects are using this location'"
        >
          <img class="small-icons" :src="getAppIcon('trash')">
        </span>
      </div>
      
      <div class="location-info">
        {{ getProjectCount(location.id) }} project(s) using this location
      </div>
    </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, defineProps, defineEmits } from 'vue';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { DialogService, SettingsService } from '@/../bindings/clustta/services';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

const iconStore = useIconStore();
const notificationStore = useNotificationStore();

const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

const props = defineProps({
  modelValue: {
    type: Array,
    default: () => []
  }
});

const emit = defineEmits(['update:modelValue', 'location-changed']);

const locations = ref([]);
const locationHealthMap = ref({});
const locationUsageMap = ref({});

// Load locations from backend or props
const loadLocations = async () => {
  try {
    // Try to load from backend if props are empty
    if (!props.modelValue || props.modelValue.length === 0) {
      const fetchedLocations = await SettingsService.GetAllLocationPaths();
      locations.value = fetchedLocations;
      emit('update:modelValue', fetchedLocations);
    } else {
      locations.value = props.modelValue;
    }
  } catch (error) {
    console.error('Error loading locations:', error);
    locations.value = props.modelValue || [];
  }
};

// Check health of all locations
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

// Get usage count for each location
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
  // Can't delete if only one location
  if (locations.value.length <= 1) return false;
  
  // Can't delete if projects are using it
  return getProjectCount(locationId) === 0;
};

const addLocation = async () => {
  const newLocationName = `Location ${locations.value.length + 1}`;
  
  // Open folder picker
  const result = await DialogService.SelectFolderDialog("Select Location Folder");
  if (!result) return;
  
  const path = result.replace(/\\/g, '/');
  
  try {
    const newLocation = await SettingsService.AddProjectLocation(newLocationName, path);
    locations.value.push(newLocation);
    emit('update:modelValue', locations.value);
    emit('location-changed');
    
    // Reload health and usage
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
    emit('update:modelValue', locations.value);
    emit('location-changed');
    
    // Reload health
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
    emit('update:modelValue', locations.value);
    emit('location-changed');
    
    notificationStore.addNotification('Location removed successfully', '', 'success', false);
  } catch (error) {
    notificationStore.errorNotification('Error removing location', error);
  }
};

// Watch for prop changes
watch(() => props.modelValue, (newValue) => {
  if (newValue && newValue.length > 0) {
    locations.value = newValue;
  }
}, { immediate: true, deep: true });

onMounted(async () => {
  await loadLocations();
  await checkAllLocationHealth();
  await loadLocationUsage();
});
</script>

<style scoped>
.location-manager {
  display: flex;
  flex-direction: column;
  gap: .5rem;
  width: 100%;
}

.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.section-label {
  font-family: Inter, sans-serif;
  color: var(--white);
  font-size: 14px;
  font-weight: 600;
  padding-left: 0.5rem;
}

.locations-scroll-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-height: 150px;
  overflow-y: auto;
  padding-right: 0.25rem;
}

.location-item {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  /* padding: 1rem; */
  background: var(--input-bg);
  border-radius: 8px;
  border: 1px solid var(--input-border);
}

.location-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.location-name {
  flex: 1;
  margin: 0;
  padding: 0.5rem;
  color: var(--white);
  font-size: 14px;
  font-weight: 600;
  font-family: Inter, sans-serif;
}

.location-path {
  flex: 1;
  padding: 0.5rem;
  color: var(--text-secondary);
  font-size: 12px;
  font-family: 'Courier New', monospace;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.default-badge {
  padding: 0.25rem 0.5rem;
  background: var(--primary-color);
  color: var(--white);
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.warning-badge {
  font-size: 16px;
  cursor: help;
}

.location-info {
  font-size: 12px;
  color: var(--text-secondary);
  padding-left: 0.5rem;
}

.disabled {
  opacity: 0.5;
  cursor: not-allowed !important;
  pointer-events: none;
}
</style>
