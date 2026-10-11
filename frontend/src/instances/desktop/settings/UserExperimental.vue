<template>
  <div class="settings-component-root user-settings-page">
    <div class="settings-component-container">
      <SettingsCard :title="$t('settings.experimentalFeatures')" :items="experimentalItems">
        <template #action="{ item }">
          <ToggleSwitch
            :switchValueProp="item.enabled"
            @click="item.toggle"
          />
        </template>
      </SettingsCard>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import SettingsCard from '@/instances/desktop/settings/components/SettingsCard.vue';
import { SettingsService } from '@/services';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';

const iconStore = useIconStore();
const metadataOnlyStorage = ref(false);
const notificationStore = useNotificationStore();
const syncAfterCheckpoint = ref(false);
const useUpdateSync = ref(false);
const { t } = useI18n();

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

const updateSetting = async ({ currentValue, failureKey, labelKey, save, successKey }) => {
  const newValue = !currentValue.value;
  try {
    await save(newValue);
    currentValue.value = newValue;
    notificationStore.addNotification(t(labelKey), t(successKey, { status: newValue ? 'enabled' : 'disabled' }), 'success');
  } catch (error) {
    console.error(error);
    notificationStore.addNotification(t('common.error'), t(failureKey), 'error');
  }
};

const toggleSyncAfterCheckpoint = () => updateSetting({
  currentValue: syncAfterCheckpoint,
  failureKey: 'notifications.failedToUpdateSyncAfterCheckpoint',
  labelKey: 'settings.syncAfterCheckpoint',
  save: SettingsService.SetSyncAfterCheckpoint,
  successKey: 'notifications.syncAfterCheckpointToggled',
});

const toggleUseUpdateSync = () => updateSetting({
  currentValue: useUpdateSync,
  failureKey: 'notifications.failedToUpdateUseUpdateSync',
  labelKey: 'settings.useUpdateSync',
  save: SettingsService.SetUseUpdateSync,
  successKey: 'notifications.useUpdateSyncToggled',
});

const toggleMetadataOnlyStorage = () => updateSetting({
  currentValue: metadataOnlyStorage,
  failureKey: 'notifications.failedToUpdateMetadataOnlyStorage',
  labelKey: 'settings.metadataOnlyStorage',
  save: SettingsService.SetMetadataOnlyStorage,
  successKey: 'notifications.metadataOnlyStorageToggled',
});

const experimentalItems = computed(() => [
  {
    id: 'sync-after-checkpoint',
    icon: getAppIcon('refresh'),
    label: t('settings.syncAfterCheckpoint'),
    description: t('settings.syncAfterCheckpointDescription'),
    enabled: syncAfterCheckpoint.value,
    toggle: toggleSyncAfterCheckpoint,
  },
  {
    id: 'update-sync',
    icon: getAppIcon('cloud-down'),
    label: t('settings.useUpdateSync'),
    description: t('settings.useUpdateSyncDescription'),
    enabled: useUpdateSync.value,
    toggle: toggleUseUpdateSync,
  },
  {
    id: 'metadata-only-storage',
    icon: getAppIcon('database-sync'),
    label: t('settings.metadataOnlyStorage'),
    description: t('settings.metadataOnlyStorageDescription'),
    enabled: metadataOnlyStorage.value,
    toggle: toggleMetadataOnlyStorage,
  },
]);

onMounted(async () => {
  try {
    syncAfterCheckpoint.value = await SettingsService.GetSyncAfterCheckpoint();
    useUpdateSync.value = await SettingsService.GetUseUpdateSync();
    metadataOnlyStorage.value = await SettingsService.GetMetadataOnlyStorage();
  } catch (error) {
    console.error(error);
  }
});
</script>
