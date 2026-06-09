<template>
  <div class="modal-container">
    <HeaderArea :title="title" :icon="'explorer'" :showSearch="showSearch" />
    <div class="general-container">

      <div class="input-section">
        <span class="regular">{{ $t('settings.clusttaLocalProjectsData') }}</span>
        <div class="horizontal-flex">
          <input v-model="projectsDirectory" class="input-short" type="text" :placeholder="$t('placeholders.projectsDirectory')"
            ref="projectsDirectoryInput" />
          <span @click="selectDirectoryPath('personal')" class="single-action-button" v-tooltip="$t('settings.browsePath')"><img
              class="small-icons" :src="getAppIcon('explorer')"></span>
        </div>
      </div>

      <div class="input-section">
        <span class="regular">{{ $t('settings.clusttaSharedProjectsData') }}</span>
        <div class="horizontal-flex">
          <input v-model="sharedProjectsDirectory" class="input-short" type="text"
            :placeholder="$t('placeholders.sharedProjectsDirectory')" ref="sharedProjectsDirectoryInput"  />
          <span @click="selectDirectoryPath('shared')" class="single-action-button" v-tooltip="$t('settings.browsePath')"><img
              class="small-icons" :src="getAppIcon('explorer')"></span>
        </div>
      </div>

      <div class="menu-divider"></div>

      <div class="pop-up-actions">
        <button class="button default" @click="closeModal()" v-stop-propagation>{{ $t('common.cancel') }}</button>
        <button class="button colored" @click="saveChanges()" v-stop-propagation>{{ $t('common.save') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

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
const { t } = useI18n();

// refs
const projectLocations = ref([]);
const projectsDirectory = ref('');
const projectsDirectoryInput = ref(null);
const sharedProjectsDirectory = ref('');
const sharedProjectsDirectoryInput = ref(null);

// constants
const showSearch = false;
const title = t('modals.configureDirectories');

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
    notificationStore.addNotification(t('notifications.settingsSaved'), '', 'success', false);
    closeModal();
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorSavingSettings'), error);
  }
};

// Opens a dialog to select a directory path.
const selectDirectoryPath = async (context) => {
  const result = await DialogService.SelectFolderDialog(t('settings.selectFolder'));
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
    notificationStore.addNotification(t('notifications.errorLoadingSettings'), error.message, 'error', false);
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.regular {
  padding-left: .5rem;
  color: hsl(var(--foreground));
  font-size: 14px;
}

.input-section {
  width: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  gap: .4px;
  color: hsl(var(--foreground));
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
  color: hsl(var(--foreground));
  font-size: 14px;
  white-space: nowrap;
  flex: 1;
}
</style>


