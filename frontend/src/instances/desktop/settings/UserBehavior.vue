<template>
  <div class="settings-component-root user-settings-page">
    <div class="settings-component-container">
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.behaviour') }}</h2>
        </div>
        <div class="settings-section-card-content">
          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon(defaultViewIcon)"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.defaultView') }}</div>
              <div class="settings-body">{{ $t('settings.defaultViewDescription') }}</div>
            </div>
            <div class="settings-action fixed-width">
              <DropDownBox :items="viewModeOptions" :onSelect="selectDefaultView" :selectedItem="currentViewLabel" :placeHolder="'None'" :fixedWidth="true" />
            </div>
          </div>

          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('minimize')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.minimizeOnClose') }}</div>
              <div class="settings-body">{{ $t('settings.minimizeOnCloseDescription') }}</div>
            </div>
            <div class="settings-action fixed-width"><ToggleSwitch :switchValueProp="minimizeOnClose" @click="toggleMinimizeOnClose" /></div>
          </div>

          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('data-download')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.overwriteDroppedFiles') }}</div>
              <div class="settings-body">{{ $t('settings.overwriteDroppedFilesDescription') }}</div>
            </div>
            <div class="settings-action fixed-width"><ToggleSwitch :switchValueProp="overwriteDroppedFiles" @click="toggleOverwriteDroppedFiles" /></div>
          </div>
        </div>
      </div>

      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.dataManagement') }}</h2>
        </div>
        <div class="settings-section-card-content">
          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('broom')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.clearRecents') }}</div>
              <div class="settings-body">{{ $t('settings.clearRecentsDescription') }}</div>
            </div>
            <div class="settings-action">
              <ActionButton :label="$t('common.clear')" :showIcon="false" :buttonFunction="clearRecents" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import { SettingsService } from '@/services';
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useSettingsStore } from '@/stores/settings';

const commonStore = useCommonStore();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const settingsStore = useSettingsStore();
const { t } = useI18n();

const minimizeOnClose = computed(() => settingsStore.minimizeOnClose);
const overwriteDroppedFiles = computed(() => settingsStore.overwriteDroppedFiles);
const currentViewLabel = computed(() => {
  if (commonStore.defaultViewMode === 'grid') return t('settings.grid');
  if (commonStore.defaultViewMode === 'kanban') return t('settings.kanban');
  return t('settings.list');
});
const defaultViewIcon = computed(() => {
  if (commonStore.defaultViewMode === 'grid') return 'four-squares';
  if (commonStore.defaultViewMode === 'kanban') return 'kanban';
  return 'list';
});
const viewModeOptions = computed(() => [t('settings.list'), t('settings.grid'), t('settings.kanban')]);

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

const selectDefaultView = async (viewLabel) => {
  const modes = new Map([
    [t('settings.list'), 'list'],
    [t('settings.grid'), 'grid'],
    [t('settings.kanban'), 'kanban'],
  ]);
  try {
    await commonStore.setDefaultViewMode(modes.get(viewLabel) || 'list');
    notificationStore.addNotification(t('notifications.defaultViewUpdated'), t('notifications.defaultViewSet', { viewType: viewLabel }), 'success');
  } catch (error) {
    console.error(error);
    notificationStore.addNotification(t('common.error'), t('notifications.failedToUpdate'), 'error');
  }
};

const toggleMinimizeOnClose = async () => {
  try {
    await settingsStore.toggleMinimizeOnClose();
    notificationStore.addNotification(t('settings.minimizeOnClose'), t('notifications.minimizeOnCloseToggled', { status: settingsStore.minimizeOnClose ? 'enabled' : 'disabled' }), 'success');
  } catch (error) {
    console.error(error);
    notificationStore.addNotification(t('common.error'), t('notifications.failedToUpdateMinimizeOnClose'), 'error');
  }
};

const toggleOverwriteDroppedFiles = async () => {
  try {
    await settingsStore.toggleOverwriteDroppedFiles();
    notificationStore.addNotification(t('settings.overwriteDroppedFiles'), t('notifications.overwriteDroppedFilesToggled', { status: settingsStore.overwriteDroppedFiles ? 'enabled' : 'disabled' }), 'success');
  } catch (error) {
    console.error(error);
    notificationStore.addNotification(t('common.error'), t('notifications.failedToUpdateOverwriteDroppedFiles'), 'error');
  }
};

const clearRecents = async () => {
  await SettingsService.ClearRecentProject();
  projectStore.recentProjects = [];
  notificationStore.addNotification(t('notifications.recentProjectsCleared'), t('notifications.recentProjectsCleared'), 'success');
};

onMounted(async () => {
  await settingsStore.initializeMinimizeOnClose();
  await settingsStore.initializeOverwriteDroppedFiles();
});
</script>

<style scoped>
.fixed-width {
  min-width: 200px;
}
</style>
