<template>
  <div class="modal-container" v-stop-propagation>
    <HeaderArea :title="$t('modals.importProjects')" :icon="getAppIcon('arrow-down-ramp')" :showSearch="false" />
    <div class="general-container">

      <!-- File Selection Card -->
      <div v-if="!isImporting && !importComplete" class="settings-section-card">
        <div class="settings-section-card-header">
          <div class="header-content">
            <h2 class="settings-section-card-title">{{ $t('modals.selectClusttaArchives') }}</h2>
            <div class="card-description">
              {{ $t('modals.chooseArchives') }}
            </div>
          </div>
          <GeneralButton 
            :label="selectedFiles.length > 0 ? $t('modals.addMore') : $t('modals.selectFiles')" 
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
                  v-tooltip="$t('common.remove')"
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
            <h2 class="settings-section-card-title">{{ $t('modals.importSuccessCheck') }}</h2>
            <div class="card-description">
              {{ $t('modals.projectsImportedCount', { count: importedFiles.length }) }}
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
                  v-tooltip="$t('modals.locateInExplorer')"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Action Buttons -->
      <div v-if="!isImporting" class="pop-up-actions" :class="{ 'import-complete' : importComplete }" >
        <GeneralButton v-if="!importComplete"
          :label="$t('common.cancel')" 
          :buttonFunction="closeModal"
          :colored="false"
          :fullWidth="false"
        />
        <GeneralButton 
          v-if="!importComplete"
          :label="$t('common.import')" 
          :buttonFunction="importProjects"
          :fullWidth="false"
          :isActive="selectedFiles.length > 0"
        />
        <GeneralButton 
          v-else
          :label="$t('common.close')" 
          :buttonFunction="closeModal"
          :fullWidth="false"
        />
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ProgressBar from '@/instances/common/components/ProgressBar.vue';

// services
import { DialogService, FSService, SettingsService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';

const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const { t } = useI18n();

// refs
const destinationDirectory = ref('');
const importComplete = ref(false);
const importedFiles = ref([]);
const isImporting = ref(false);
const selectedFiles = ref([]);

// methods
// Closes the modal and optionally reloads projects.
const closeModal = async () => {
  if (importComplete.value) {
    projectStore.loadProjects();
  }
  modals.setModalVisibility('importProjectModal', false);
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Extracts the filename from a path.
const getFileName = (path) => {
  return path.split('/').pop().split('\\').pop();
};

// Imports the selected project files.
const importProjects = async () => {
  if (selectedFiles.value.length === 0) {
    notificationStore.addNotification(t('notifications.noFilesSelected'), t('notifications.selectAtLeastOneFile'), 'error', false);
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
      t('notifications.importSuccessful'),
      t('notifications.projectsImportedCount', { count: importedPaths.length }),
      'success',
      false
    );
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorImportingProjects'), error);
  } finally {
    stage.operationActive = false;
    isImporting.value = false;
  }
};

// Opens the file location in file explorer.
const locateFile = (filePath) => {
  if (filePath) {
    FSService.RevealInExplorer(filePath);
  }
};

// Removes a file from the selection.
const removeFile = (index) => {
  selectedFiles.value.splice(index, 1);
};

// Opens a dialog to select project files to import.
const selectFiles = async () => {
  const result = await DialogService.SelectFilesDialog('Select .clst Files to Import', '.clst');
  if (result && result.length > 0) {
    const newFiles = [];
    result.forEach(filePath => {
      const normalizedPath = filePath.replace(/\\/g, '/');
      const fileName = normalizedPath.split('/').pop();
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

// lifecycle hooks
onMounted(async () => {
  try {
    const personalProjectsDir = await SettingsService.GetProjectDirectory();
    destinationDirectory.value = personalProjectsDir.replace(/\\/g, '/');
  } catch (error) {
    console.error('Failed to get project directory:', error);
    notificationStore.errorNotification(t('notifications.errorLoadingDestination'), error);
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
  box-sizing: border-box;
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
