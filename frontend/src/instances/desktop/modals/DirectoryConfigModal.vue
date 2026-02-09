<template>
  <div class="modal-container">
    <HeaderArea :title="title" :icon="'explorer'" :showSearch="showSearch" />
    <div class="general-container">

      <div class="input-section">
        <span class="regular">Clustta local projects data</span>
        <div class="horizontal-flex">
          <input v-model="projectsDirectory" class="input-short" type="text" placeholder="Projects Directory"
            ref="projectsDirectoryInput" />
          <span @click="selectDirectoryPath('personal')" class="single-action-button" v-tooltip="'Browse Path'"><img
              class="small-icons" :src="getAppIcon('explorer')"></span>
        </div>
      </div>

      <div class="input-section">
        <span class="regular">Clustta shared projects data</span>
        <div class="horizontal-flex">
          <input v-model="sharedProjectsDirectory" class="input-short" type="text"
            placeholder="Shared Projects Directory" ref="sharedProjectsDirectoryInput"  />
          <span @click="selectDirectoryPath('shared')" class="single-action-button" v-tooltip="'Browse Path'"><img
              class="small-icons" :src="getAppIcon('explorer')"></span>
        </div>
      </div>

      <div class="menu-divider"></div>

      <div class="pop-up-actions">
        <button class="button default" @click="closeModal()" v-stop-propagation>Cancel</button>
        <button class="button colored" @click="saveChanges()" v-stop-propagation>Save</button>
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { onMounted, ref } from 'vue';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { DialogService, SettingsService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';

const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();

// refs
const projectLocations = ref([]);
const projectsDirectory = ref('');
const projectsDirectoryInput = ref(null);
const sharedProjectsDirectory = ref('');
const sharedProjectsDirectoryInput = ref(null);

// constants
const showSearch = false;
const title = 'Configure Directories';

// methods
// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Saves the directory configuration changes.
const saveChanges = async () => {
  try {
    await SettingsService.SetProjectDirectory(projectsDirectory.value);
    await SettingsService.SetSharedProjectDirectory(sharedProjectsDirectory.value);
    notificationStore.addNotification('Settings saved successfully', '', 'success', false);
    closeModal();
  } catch (error) {
    notificationStore.errorNotification('Error saving settings', error);
  }
};

// Opens a dialog to select a directory path.
const selectDirectoryPath = async (context) => {
  const result = await DialogService.SelectFolderDialog('Select Folder File');
  if (result) {
    const fileDir = result.replace(/\\/g, '/');
    if (context === 'shared') {
      sharedProjectsDirectory.value = fileDir;
      sharedProjectsDirectoryInput.value.focus();
    } else if (context === 'personal') {
      projectsDirectory.value = fileDir;
      projectsDirectoryInput.value.focus();
    }
  }
};

// lifecycle hooks
onMounted(async () => {
  try {
    projectsDirectory.value = await SettingsService.GetProjectDirectory();
    sharedProjectsDirectory.value = await SettingsService.GetSharedProjectDirectory();
    projectLocations.value = await SettingsService.GetAllLocationPaths();
  } catch (error) {
    notificationStore.addNotification('Error Loading Settings', error.message, 'error', false);
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.regular {
  padding-left: .5rem;
  color: var(--white);
  font-size: 14px;
}

.input-section {
  width: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  gap: .4px;
  color: white;
}

.general-container {
  gap: 1rem;
}

.input-short {
  flex: 1;
  width: 100%;
}

.input-label {
  font-family: Inter, sans-serif;
  color: var(--white);
  font-size: 14px;
  white-space: nowrap;
  flex: 1;
}
</style>


