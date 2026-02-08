<template>
  <div class="modal-container" v-stop-propagation>
    <HeaderArea :title="'Backup Project'" :icon="getAppIcon('clustta')" :showSearch="false" />
    <div class="general-container">

      <!-- Project Info Display -->
      <div v-if="!isBackingUp && !isSyncing && !backupComplete" class="settings-section-card">
        <div class="settings-section-card-header">
          <div class="header-content">
            <h2 class="settings-section-card-title">Project to backup</h2>
            <div class="card-description">
              {{ projectStore.getActiveProjectName }}
            </div>
          </div>
        </div>
        <div class="settings-section-card-content">
          <!-- Source File Display -->
          <div class="location-item location-item-single">
            <div class="location-icon">
              <img class="small-icons" :src="getAppIcon('clustta')">
            </div>
            <div class="location-content">
              <div class="location-header">
                <div class="location-name">Source File</div>
              </div>
              <div class="location-body">
                {{ projectStore.getActiveProject.uri }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Backup Destination Card -->
      <div v-if="!isBackingUp && !isSyncing && !backupComplete" class="settings-section-card">
        <div class="settings-section-card-header">
          <div class="header-content">
            <h2 class="settings-section-card-title">Backup destination</h2>
            <div class="card-description">
              Select where you want to save the backup copy of your .clst file.
            </div>
          </div>
          <GeneralButton 
            :label="selectedBackupDirectory ? 'Change' : 'Select'" 
            :buttonFunction="selectBackupDirectory"
            :fullWidth="false"
          />
        </div>
        <div class="settings-section-card-content" :class="{ 'no-padding': !selectedBackupDirectory }">
          
          <!-- Selected Backup Path Display -->
          <div v-if="selectedBackupDirectory" class="location-item location-item-single">
            <div class="location-icon">
              <img class="small-icons" :src="getAppIcon('folder')">
            </div>
            <div class="location-content">
              <div class="location-header">
                <div class="location-name">Backup Location</div>
              </div>
              <div class="location-body">
                {{ selectedBackupDirectory }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Full Sync Option Card -->
      <div v-if="!isBackingUp && !isSyncing && projectStore.activeProject.has_remote && !backupComplete" class="settings-section-card">
        <div class="settings-section-card-header">
          <div class="header-content">
            <h2 class="settings-section-card-title">Sync before backup</h2>
            <div class="card-description">
              Optionally perform a full sync to ensure all checkpoints are included in the backup.
            </div>
          </div>
          <GeneralButton 
            label="Full Sync" 
            :buttonFunction="performFullSync"
            :fullWidth="false"
            :isDisabled="isSyncing"
          />
        </div>
      </div>

      <!-- Progress Display -->
      <div v-if="isBackingUp || isSyncing" class="settings-section-card">
      <div class="progress-section">
        <div class="progress-header">
          <span class="progress-title">{{ isSyncing ? 'Performing full sync' : notificationStore.progress.title }}</span>
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

      <!-- Backup Success Message -->
      <div v-if="backupComplete" class="settings-section-card">
        <div class="settings-section-card-header">
          <div class="header-content">
            <h2 class="settings-section-card-title">✓ Backup successful</h2>
            <div class="card-description">
              {{ backupDestinationPath }}
            </div>
          </div>
        </div>
        <div class="settings-section-card-content">
          <div class="location-item location-item-single">
            <div class="location-icon">
              <img class="small-icons" :src="getAppIcon('folder')">
            </div>
            <div class="location-content">
              <div class="location-header">
                <div class="location-name">Backup File</div>
              </div>
              <div class="location-body">
                {{ backupDestinationPath }}
              </div>
            </div>
            <div class="location-actions">
              <ActionButton 
                :icon="getAppIcon('folder-arrow-up-right')" 
                :buttonFunction="() => locateBackupFile()"
                v-tooltip="'Locate in Explorer'"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- Action Buttons -->
      <div v-if="!isBackingUp" class="pop-up-actions" :class="{ 'pop-up-actions-syncing' : isSyncing || backupComplete }" >
        <GeneralButton 
          v-if="!backupComplete && !isSyncing"
          label="Cancel" 
          :buttonFunction="closeModal"
          :fullWidth="false"
        />
        <GeneralButton 
          v-if="!backupComplete && isSyncing"
          label="Cancel Sync"
          :colored="false" 
          :buttonFunction="cancelOperation"
          :fullWidth="false"
        />
        <GeneralButton 
          v-else-if="!backupComplete"
          label="Backup" 
          :buttonFunction="backupProject"
          :fullWidth="false"
          :isActive="selectedBackupDirectory !== ''"
        />
        <GeneralButton 
          v-if="backupComplete"
          label="Close" 
          :buttonFunction="closeModal"
          :fullWidth="false"
        />
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { ref } from 'vue';
import { syncFullData } from '@/lib/sync';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ProgressBar from '@/instances/common/components/ProgressBar.vue';

// services
import { DialogService, FSService, SyncService } from '@/services';

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

// refs
const backupComplete = ref(false);
const backupDestinationPath = ref('');
const isAwaitingResponse = ref(false);
const isBackingUp = ref(false);
const isSyncing = ref(false);
const selectedBackupDirectory = ref('');

// methods
// Creates a backup of the project to the selected directory.
const backupProject = async () => {
  if (!selectedBackupDirectory.value) {
    notificationStore.addNotification('No destination selected', 'Please select a backup location', 'error', false);
    return;
  }

  try {
    isBackingUp.value = true;
    stage.operationActive = true;
    const project = projectStore.getActiveProject;
    const sourceFile = project.uri;
    const destinationDirectory = selectedBackupDirectory.value;
    const destinationPath = await FSService.BackupFile(sourceFile, destinationDirectory);
    backupDestinationPath.value = destinationPath;
    backupComplete.value = true;
    notificationStore.addNotification('Backup successful', `Project backed up to: ${destinationPath}`, 'success', false);
  } catch (error) {
    notificationStore.errorNotification('Error backing up project', error);
  } finally {
    stage.operationActive = false;
    isBackingUp.value = false;
  }
};

// Cancels the current sync operation.
const cancelOperation = async () => {
  isAwaitingResponse.value = true;
  await notificationStore.cancleFunction();
  notificationStore.resetProgress();
  notificationStore.cancleFunction = null;
  notificationStore.canCancel = false;
  isAwaitingResponse.value = false;
};

// Closes the modal and resets state.
const closeModal = () => {
  selectedBackupDirectory.value = '';
  isBackingUp.value = false;
  backupComplete.value = false;
  backupDestinationPath.value = '';
  modals.setModalVisibility('backUpProjectModal', false);
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Opens the backup file location in file explorer.
const locateBackupFile = () => {
  if (backupDestinationPath.value) {
    FSService.RevealInExplorer(backupDestinationPath.value);
  }
};

// Performs a full sync before backup.
const performFullSync = async () => {
  try {
    isSyncing.value = true;
    stage.operationActive = true;
    notificationStore.cancleFunction = SyncService.CancelSync;
    notificationStore.canCancel = true;
    await syncFullData();
    notificationStore.addNotification('Sync complete', 'All checkpoints have been synced successfully', 'success', false);
  } catch (error) {
    notificationStore.errorNotification('Sync failed', error);
  } finally {
    stage.operationActive = false;
    isSyncing.value = false;
  }
};

// Opens a dialog to select the backup destination directory.
const selectBackupDirectory = async () => {
  const result = await DialogService.SelectFolderDialog('Select Backup Location');
  if (result) {
    const fileDir = result.replace(/\\/g, '/');
    selectedBackupDirectory.value = fileDir;
  }
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.general-container {
  display: flex;
  flex-direction: column;
  width: 600px;
  max-width: 600px;
  box-sizing: border-box;
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

.location-item-single{
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

.location-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  box-sizing: border-box;
}

.pop-up-actions {
  display: flex;
  justify-content: space-between;
  gap: 0.5rem;
}

.pop-up-actions-syncing {
  justify-content: flex-end;
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
  border-radius: 999px;
  overflow: hidden;
  background-color: var(--dark-steel);

  position: relative;
  width: 100%;
  height: .2rem;
  border-radius: 999px;
  /* background-color: white; */

}

.progress-meta {
  display: flex;
  justify-content: flex-end;
  font-size: 12px;
  color: var(--silver);
  opacity: 0.8;
}
</style>
