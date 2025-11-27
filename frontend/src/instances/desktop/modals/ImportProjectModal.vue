<template>
  <div class="modal-container" v-stop-propagation>
    <HeaderArea :title="'Import Projects'" :icon="getAppIcon('arrow-down-ramp')" :showSearch="false" />
    <div class="general-container">

      <!-- File Selection Card -->
      <div v-if="!isImporting && !importComplete" class="settings-section-card">
        <div class="settings-section-card-header">
          <div class="header-content">
            <h2 class="settings-section-card-title">Select Clustta project archives to import</h2>
            <div class="card-description">
              Choose one or more .clst archives to import into your personal studio.
            </div>
          </div>
          <GeneralButton 
            :label="selectedFiles.length > 0 ? 'Add More' : 'Select Files'" 
            :buttonFunction="selectFiles"
            :fullWidth="false"
          />
        </div>
        <div v-if="selectedFiles.length > 0" class="settings-section-card-content">
          <div class="file-list">
            <div v-for="(file, index) in selectedFiles" :key="index" class="file-item">
              <div class="file-icon-wrapper">
                <img class="small-icons no-filter" :src="getAppIcon('clustta')">
              </div>
              <div class="file-info">
                <div class="file-name">{{ file.name }}</div>
                <div class="file-path">{{ file.path }}</div>
              </div>
              <div class="file-actions">
                <ActionButton 
                  :icon="getAppIcon('trash')" 
                  :useDanger="true"
                  :noFilter="true"
                  :buttonFunction="() => removeFile(index)"
                  v-tooltip="'Remove'"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Progress Display -->
      <div v-if="isImporting" class="settings-section-card">
        <div class="progress-section">
          <div class="progress-header">
            <span class="progress-title">{{ notificationStore.progress.title }}</span>
            <span class="progress-percentage">{{ Math.round(notificationStore.progress.percentage) }}%</span>
          </div>
          <div class="progress-message">{{ notificationStore.progress.message }}</div>
          <div class="progress-bar-wrapper">
            <ProgressBar :taskProgress="notificationStore.progress.percentage" />
          </div>
          <div class="progress-meta">
            <span>{{ notificationStore.progress.current }}/{{ notificationStore.progress.total }}</span>
          </div>
        </div>
      </div>

      <!-- Import Success Message -->
      <div v-if="importComplete" class="settings-section-card">
        <div class="settings-section-card-header">
          <div class="header-content">
            <h2 class="settings-section-card-title">✓ Import successful</h2>
            <div class="card-description">
              {{ importedFiles.length }} project{{ importedFiles.length > 1 ? 's' : '' }} imported successfully
            </div>
          </div>
        </div>
        <div class="settings-section-card-content">
          <div class="file-list">
            <div v-for="(filePath, index) in importedFiles" :key="index" class="file-item">
              <div class="file-icon-wrapper">
                <img class="small-icons no-filter" :src="getAppIcon('clustta')">
              </div>
              <div class="file-info">
                <div class="file-name">{{ getFileName(filePath) }}</div>
                <div class="file-path">{{ filePath }}</div>
              </div>
              <div class="file-actions">
                <ActionButton 
                  :icon="getAppIcon('folder-arrow-up-right')" 
                  :buttonFunction="() => locateFile(filePath)"
                  v-tooltip="'Locate in Explorer'"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Action Buttons -->
      <div v-if="!isImporting" class="pop-up-actions" :class="{ 'import-complete' : importComplete }" >
        <GeneralButton v-if="!importComplete"
          label="Cancel" 
          :buttonFunction="closeModal"
          :colored="false"
          :fullWidth="false"
        />
        <GeneralButton 
          v-if="!importComplete"
          label="Import" 
          :buttonFunction="importProjects"
          :fullWidth="false"
          :isActive="selectedFiles.length > 0"
        />
        <GeneralButton 
          v-else
          label="Close" 
          :buttonFunction="closeModal"
          :fullWidth="false"
        />
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';

// Components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import ProgressBar from '@/instances/common/components/ProgressBar.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

import { DialogService, FSService, SettingsService } from '@/../bindings/clustta/services/index';

// Stores
const iconStore = useIconStore();
const projectStore = useProjectStore();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const stage = useStageStore();

// Refs
const selectedFiles = ref([]);
const destinationDirectory = ref('');
const isImporting = ref(false);
const importComplete = ref(false);
const importedFiles = ref([]);

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon;
};

const getFileName = (path) => {
  return path.split('/').pop().split('\\').pop();
};

// Methods
const selectFiles = async () => {
  const result = await DialogService.SelectFilesDialog("Select .clst Files to Import", ".clst");
  
  if (result && result.length > 0) {
    const newFiles = [];
    
    result.forEach(filePath => {
      let normalizedPath = filePath.replace(/\\/g, '/');
      let fileName = normalizedPath.split('/').pop();
      
      // Check if file is already in the list
      const alreadySelected = selectedFiles.value.some(f => f.path === normalizedPath);
      if (!alreadySelected) {
        newFiles.push({
          path: normalizedPath,
          name: fileName
        });
      }
    });
    
    selectedFiles.value.push(...newFiles);
  }
};

const removeFile = (index) => {
  selectedFiles.value.splice(index, 1);
};

const selectDestination = async () => {
  const result = await DialogService.SelectFolderDialog("Select Import Destination");

  if (result) {
    let fileDir = result.replace(/\\/g, '/');
    destinationDirectory.value = fileDir;
  }
};

const importProjects = async () => {
  if (selectedFiles.value.length === 0) {
    notificationStore.addNotification(
      'No files selected',
      'Please select at least one .clst file to import',
      'error',
      false
    );
    return;
  }

  try {
    isImporting.value = true;
    stage.operationActive = true;
    
    const sourcePaths = selectedFiles.value.map(file => file.path);
    const destination = destinationDirectory.value;
    
    const importedPaths = await FSService.ImportClusttaFiles(sourcePaths, destination);
    
    importedFiles.value = importedPaths;
    importComplete.value = true;
    
    notificationStore.addNotification(
      'Import successful',
      `${importedPaths.length} project${importedPaths.length > 1 ? 's' : ''} imported successfully`,
      'success',
      false
    );
    
  } catch (error) {
    notificationStore.errorNotification('Error importing projects', error);
  } finally {
    stage.operationActive = false;
    isImporting.value = false;
  }
};

const locateFile = (filePath) => {
  if (filePath) {
    FSService.RevealInExplorer(filePath);
  }
};

const closeModal = async () => {
    if (importComplete.value){
        projectStore.loadProjects();
    }
    modals.setModalVisibility('importProjectModal', false);
};

// Load default destination on mount
onMounted(async () => {
  try {
    const personalProjectsDir = await SettingsService.GetProjectDirectory();
    console.log(personalProjectsDir)
    destinationDirectory.value = personalProjectsDir.replace(/\\/g, '/');
  } catch (error) {
    console.error('Failed to get project directory:', error);
    notificationStore.errorNotification('Error loading destination', error);
  }
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.general-container {
  display: flex;
  flex-direction: column;
  width: 600px;
  max-width: 600px;
  max-height: 70vh;
  color: white;
  box-sizing: border-box;
  padding: 0;
  overflow-y: auto;
}

/* Settings Section Card Styles */
.settings-section-card {
  display: flex;
  flex-direction: column;
  background-color: var(--dark-steel);
  overflow: hidden;
  box-sizing: border-box;
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

.card-description {
  font-size: 13px;
  color: var(--silver);
  opacity: 0.9;
  line-height: 1.5;
}

/* File List Styles */
.file-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-height: 300px;
  overflow-y: auto;
  padding: 0.5rem;
  background-color: var(--light-steel);
  border-radius: var(--normal-radius);
}

.file-list::-webkit-scrollbar {
  width: 4px;
}

.file-list::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--dark-steel);
}

.file-list::-webkit-scrollbar-track {
  border-radius: 10px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0 0.75rem;
  background-color: var(--steel);
  border-radius: var(--normal-radius);
  min-height: 60px;
}

.file-icon-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.3rem;
  flex-shrink: 0;
}

.file-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
  overflow: hidden;
}

.file-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--white);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-path {
  font-size: 12px;
  color: var(--silver);
  opacity: 0.8;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

/* Location Items */
.location-item {
  display: flex;
  align-items: center;
  overflow: hidden;
  box-sizing: border-box;
  min-height: 60px;
  height: max-content;
  padding: 0.75rem 1rem;
  gap: 0.75rem;
  background-color: var(--light-steel);
}

.location-item-single {
  border-radius: var(--normal-radius);
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

/* Progress Section */
.progress-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1rem;
  background-color: var(--steel);
  border-radius: var(--normal-radius);
  width: 100%;
  box-sizing: border-box;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: var(--white);
}

.progress-title {
  font-size: 15px;
  font-weight: 500;
}

.progress-percentage {
  font-size: 14px;
  font-weight: 600;
  color: rgb(67, 210, 67);
}

.progress-message {
  font-size: 13px;
  color: var(--silver);
  opacity: 0.9;
}

.progress-bar-wrapper {
  position: relative;
  width: 100%;
  height: 0.2rem;
  border-radius: 999px;
  overflow: hidden;
  background-color: var(--dark-steel);
}

.progress-meta {
  display: flex;
  justify-content: flex-end;
  font-size: 12px;
  color: var(--silver);
  opacity: 0.8;
}

/* Action Buttons */
.pop-up-actions {
  display: flex;
  justify-content: space-between;
  gap: 0.5rem;
}


.import-complete{
    justify-content: flex-end;
}
</style>
