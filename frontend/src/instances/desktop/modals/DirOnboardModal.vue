<template>
  <div class="modal-container" v-stop-propagation>
    <HeaderArea :title="$t('modals.setupClustta')" :icon="getAppIcon('clustta')" :showSearch="false" />
    <div class="general-container">
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <div class="header-content">
            <h2 class="settings-section-card-title">{{ $t('modals.clusttaDataLocation') }}</h2>
            <div class="card-description">{{ $t('modals.clusttaDataLocationDesc') }}</div>
          </div>
        </div>
        <div class="settings-section-card-content">
          <div class="path-row">
            <FormInput
              v-model="dataDirectory"
              :placeholder="$t('modals.clusttaDataLocation')"
              :info="isMacAppStore ? $t('modals.macFolderPermissionDesc') : ''"
              @input="handleDataDirectoryInput"
            />
            <ActionButton
              :icon="getAppIcon('explorer')"
              :buttonFunction="selectDataDirectory"
              :showLabel="false"
              v-tooltip="$t('common.browse')"
            />
          </div>
        </div>
      </div>

      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <div class="header-content">
            <h2 class="settings-section-card-title">{{ $t('modals.workingProjectsLocation') }}</h2>
            <div class="card-description">{{ $t('modals.workingProjectsLocationDesc') }}</div>
          </div>
        </div>
        <div class="settings-section-card-content">
          <div class="path-row">
            <FormInput
              v-model="workingProjectsDirectory"
              :placeholder="$t('modals.workingProjectsLocation')"
              :info="isMacAppStore ? $t('modals.macFolderPermissionDesc') : ''"
              @input="handleWorkingProjectsInput"
            />
            <ActionButton
              :icon="getAppIcon('explorer')"
              :buttonFunction="selectWorkingProjectsDirectory"
              :showLabel="false"
              v-tooltip="$t('common.browse')"
            />
          </div>
        </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton
          :label="$t('common.cancel')"
          :buttonFunction="cancelOnboarding"
          :colored="false"
          :fullWidth="false"
          :isActive="!isAwaitingResponse"
          :loading="isCancelling"
        />
        <GeneralButton
          :label="$t('common.continue')"
          :buttonFunction="saveChanges"
          :fullWidth="false"
          :isActive="canContinue"
          :loading="isAwaitingResponse"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { resetStoreInitialization } from '@/router';

import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import FormInput from '@/instances/desktop/components/FormInput.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

import { DialogService, FSService, SettingsService } from '@/services';

import { useAccountStore } from '@/stores/accounts';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useEntitlementStore } from '@/stores/entitlements';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

const DEFAULT_DATA_FOLDER = 'clustta';
const DEFAULT_PROJECTS_FOLDER = 'Documents/Projects';
const PERSONAL_PROJECTS_FOLDER = 'projects';
const SHARED_PROJECTS_FOLDER = 'shared_projects';

const accountStore = useAccountStore();
const entitlementStore = useEntitlementStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const router = useRouter();
const trayStates = useTrayStates();
const userStore = useUserStore();
const { t } = useI18n();

const dataDirectory = ref('');
const dataPermissionPath = ref('');
const isAwaitingResponse = ref(false);
const isCancelling = ref(false);
const workingProjectsDirectory = ref('');
const workingProjectsPermissionPath = ref('');

const isMacAppStore = computed(() => platformStore.isMacAppStore);
const dataPermissionGranted = computed(() =>
  pathsMatch(dataDirectory.value, dataPermissionPath.value)
);
const projectsPermissionGranted = computed(() =>
  pathsMatch(workingProjectsDirectory.value, workingProjectsPermissionPath.value)
);
const canContinue = computed(() =>
  Boolean(dataDirectory.value.trim() && workingProjectsDirectory.value.trim()) &&
  !isAwaitingResponse.value &&
  !isCancelling.value
);

const normalizePath = (path) => {
  const normalized = path.trim().replace(/\\/g, '/');
  if (/^[A-Za-z]:\/$/.test(normalized) || normalized === '/') return normalized;
  return normalized.replace(/\/+$/, '');
};

const pathsMatch = (firstPath, secondPath) => {
  if (!firstPath || !secondPath) return false;
  const first = normalizePath(firstPath);
  const second = normalizePath(secondPath);
  return platformStore.isWindows
    ? first.toLowerCase() === second.toLowerCase()
    : first === second;
};

const joinPath = (basePath, childPath) => {
  const base = normalizePath(basePath);
  const separator = base.endsWith('/') ? '' : '/';
  return `${base}${separator}${childPath.replace(/^\/+/, '')}`;
};

const parentPath = (path) => {
  const normalized = normalizePath(path);
  const separatorIndex = normalized.lastIndexOf('/');
  if (separatorIndex <= 0) return normalized;
  return normalized.slice(0, separatorIndex);
};

const dialogStartPath = async (path) => {
  const normalized = normalizePath(path);
  if (await FSService.DirExists(normalized)) return normalized;
  return parentPath(normalized);
};

const selectFolder = async (title, currentPath) => {
  const startPath = await dialogStartPath(currentPath);
  const result = await DialogService.SelectSpecificFolderDialog(title, startPath);
  return result ? normalizePath(result) : '';
};

const selectDataDirectory = async () => {
  const selectedPath = await selectFolder(t('modals.selectClusttaDataFolder'), dataDirectory.value);
  if (!selectedPath) return false;
  dataDirectory.value = selectedPath;
  if (isMacAppStore.value) dataPermissionPath.value = selectedPath;
  return true;
};

const selectWorkingProjectsDirectory = async () => {
  const selectedPath = await selectFolder(
    t('modals.selectWorkingProjectsFolder'),
    workingProjectsDirectory.value
  );
  if (!selectedPath) return false;
  workingProjectsDirectory.value = selectedPath;
  if (isMacAppStore.value) workingProjectsPermissionPath.value = selectedPath;
  return true;
};

const handleDataDirectoryInput = () => {
  if (!dataPermissionGranted.value) dataPermissionPath.value = '';
};

const handleWorkingProjectsInput = () => {
  if (!projectsPermissionGranted.value) workingProjectsPermissionPath.value = '';
};

const ensureMacFolderPermissions = async () => {
  if (!isMacAppStore.value) return true;

  if (!dataPermissionGranted.value && !await selectDataDirectory()) {
    notificationStore.addNotification(
      t('modals.folderPermissionRequired'),
      t('modals.macFolderPermissionCancelled'),
      'error',
      false
    );
    return false;
  }

  if (!projectsPermissionGranted.value && !await selectWorkingProjectsDirectory()) {
    notificationStore.addNotification(
      t('modals.folderPermissionRequired'),
      t('modals.macFolderPermissionCancelled'),
      'error',
      false
    );
    return false;
  }

  return true;
};

const saveProjectLocation = async () => {
  const selectedPath = normalizePath(workingProjectsDirectory.value);
  const existingLocations = await SettingsService.GetAllLocationPaths();
  const existingLocation = existingLocations.find((location) =>
    pathsMatch(location.path, selectedPath)
  );
  const location = existingLocation || await SettingsService.AddProjectLocation(
    selectedPath.split('/').pop() || t('settings.defaultLocation'),
    selectedPath
  );
  await SettingsService.SetDefaultLocation(location.id);
};

const returnToWelcome = async () => {
  userStore.$reset();
  projectStore.$reset();
  trayStates.$reset();
  entitlementStore.reset();
  accountStore.$reset();
  resetStoreInitialization();
  await router.replace('/auth/welcome');
};

const cancelOnboarding = async () => {
  if (isAwaitingResponse.value || isCancelling.value) return;

  const currentAccountId = userStore.user?.id ||
    accountStore.activeAccount?.user?.id ||
    accountStore.currentAccount?.id;
  if (!currentAccountId) {
    modals.disableAllModals();
    await returnToWelcome();
    return;
  }

  isCancelling.value = true;
  try {
    const accountRemoved = await accountStore.removeAccount(currentAccountId);
    if (!accountRemoved) {
      throw new Error(t('notifications.unableToSignOut'));
    }

    const nextAccountId = accountStore.activeAccount?.user?.id ||
      accountStore.currentAccount?.id ||
      accountStore.accounts[0]?.id;

    modals.disableAllModals();
    if (nextAccountId) {
      await accountStore.switchToAccount(nextAccountId);
      return;
    }

    await returnToWelcome();
  } catch (error) {
    notificationStore.errorNotification(t('notifications.signOutFailed'), error);
  } finally {
    isCancelling.value = false;
  }
};

const saveChanges = async () => {
  if (!canContinue.value) return;

  isAwaitingResponse.value = true;
  try {
    if (!await ensureMacFolderPermissions()) return;

    const selectedDataDirectory = normalizePath(dataDirectory.value);
    await SettingsService.SetProjectDirectory(
      joinPath(selectedDataDirectory, PERSONAL_PROJECTS_FOLDER)
    );
    await SettingsService.SetSharedProjectDirectory(
      joinPath(selectedDataDirectory, SHARED_PROJECTS_FOLDER)
    );
    await saveProjectLocation();
    await projectStore.loadStudios();
    await projectStore.loadProjects();
    trayStates.refreshData();
    modals.disableAllModals();
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorSavingSettings'), error);
  } finally {
    isAwaitingResponse.value = false;
  }
};

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

onMounted(async () => {
  try {
    await platformStore.initialize();
    const userDirectory = normalizePath(await SettingsService.GetUserDirectory());
    dataDirectory.value = joinPath(userDirectory, DEFAULT_DATA_FOLDER);
    workingProjectsDirectory.value = joinPath(userDirectory, DEFAULT_PROJECTS_FOLDER);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorLoadingSettings'), error);
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.card-description {
  color: var(--text-muted);
  font-size: 13px;
  line-height: 1.5;
  opacity: 0.9;
}

.general-container {
  display: flex;
  flex-direction: column;
  width: 600px;
  max-width: 600px;
  color: var(--text);
  box-sizing: border-box;
}

.header-content {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 0.25rem;
}

.path-row {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
}

.path-row :deep(.form-group) {
  margin-bottom: 0;
}

.path-row :deep(.input-alert-info) {
  color: var(--warning);
  opacity: 1;
}

.path-row :deep(.action-button) {
  margin-top: 2px;
}

.pop-up-actions {
  display: flex;
  justify-content: space-between;
  gap: 0.5rem;
}

.settings-section-card {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-sizing: border-box;
  background-color: var(--surface-2);
}

.settings-section-card-content {
  display: flex;
  flex-direction: column;
  padding: 1rem 1.5rem;
}

.settings-section-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  margin: 0;
  border-radius: var(--normal-radius);
  background-color: var(--bg);
}

.settings-section-card-title {
  margin: 0;
  color: var(--text);
  font-size: 16px;
  font-weight: 400;
}
</style>
