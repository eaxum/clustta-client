<template>
  <div class="settings-component-root user-settings-page">
    <div class="settings-component-container">
      <SettingsCard :title="$t('settings.behaviour')" :items="behaviorItems">
        <template #action="{ item }">
          <DropDownBox
            v-if="item.id === 'default-view'"
            :items="viewModeOptions"
            :onSelect="selectDefaultView"
            :selectedItem="currentViewLabel"
            :placeHolder="'None'"
            :fixedWidth="true"
          />
          <DropDownBox
            v-else-if="item.id === 'interface-motion'"
            :items="motionOptions"
            :onSelect="selectMotionPreference"
            :selectedItem="motionPreference"
            :placeHolder="$t('settings.motionSystem')"
            :fixedWidth="true"
          />
          <ToggleSwitch
            v-else-if="item.id === 'minimize-on-close'"
            :switchValueProp="minimizeOnClose"
            @click="toggleMinimizeOnClose"
          />
          <ToggleSwitch
            v-else-if="item.id === 'overwrite-dropped-files'"
            :switchValueProp="overwriteDroppedFiles"
            @click="toggleOverwriteDroppedFiles"
          />
        </template>
      </SettingsCard>

      <SettingsCard :title="$t('settings.dataManagement')" :items="dataManagementItems">
        <template #action>
          <ActionButton :label="$t('common.clear')" :showIcon="false" :buttonFunction="clearRecents" />
        </template>
      </SettingsCard>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import SettingsCard from '@/instances/desktop/settings/components/SettingsCard.vue';
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

const MOTION_SYSTEM = 'system';
const MOTION_ON = 'on';
const MOTION_OFF = 'off';

const minimizeOnClose = computed(() => settingsStore.minimizeOnClose);
const overwriteDroppedFiles = computed(() => settingsStore.overwriteDroppedFiles);
const motionPreference = computed(() => settingsStore.motionPreference);
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
const motionOptions = computed(() => [
  { id: MOTION_SYSTEM, name: t('settings.motionSystem'), selectionValue: MOTION_SYSTEM },
  { id: MOTION_ON, name: t('settings.motionOn'), selectionValue: MOTION_ON },
  { id: MOTION_OFF, name: t('settings.motionOff'), selectionValue: MOTION_OFF },
]);
const behaviorItems = computed(() => [
  {
    id: 'default-view',
    icon: getAppIcon(defaultViewIcon.value),
    label: t('settings.defaultView'),
    description: t('settings.defaultViewDescription'),
  },
  {
    id: 'interface-motion',
    icon: getAppIcon('sparkles'),
    label: t('settings.interfaceMotion'),
    description: t('settings.interfaceMotionDescription'),
  },
  {
    id: 'minimize-on-close',
    icon: getAppIcon('minimize'),
    label: t('settings.minimizeOnClose'),
    description: t('settings.minimizeOnCloseDescription'),
  },
  {
    id: 'overwrite-dropped-files',
    icon: getAppIcon('data-download'),
    label: t('settings.overwriteDroppedFiles'),
    description: t('settings.overwriteDroppedFilesDescription'),
  },
]);
const dataManagementItems = computed(() => [
  {
    id: 'clear-recents',
    icon: getAppIcon('broom'),
    label: t('settings.clearRecents'),
    description: t('settings.clearRecentsDescription'),
  },
]);

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

const selectMotionPreference = async (preference) => {
  try {
    await settingsStore.setMotionPreference(preference);
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
