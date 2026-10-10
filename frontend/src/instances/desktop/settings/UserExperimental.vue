<template>
  <div class="settings-component-root user-settings-page">
    <div class="settings-component-container">
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.experimentalFeatures') }}</h2>
        </div>
        <div class="settings-section-card-content">
          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('refresh')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.syncAfterCheckpoint') }}</div>
              <div class="settings-body">{{ $t('settings.syncAfterCheckpointDescription') }}</div>
            </div>
            <div class="settings-action fixed-width"><ToggleSwitch :switchValueProp="syncAfterCheckpoint" @click="toggleSyncAfterCheckpoint" /></div>
          </div>

          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('cloud-down')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.useUpdateSync') }}</div>
              <div class="settings-body">{{ $t('settings.useUpdateSyncDescription') }}</div>
            </div>
            <div class="settings-action fixed-width"><ToggleSwitch :switchValueProp="useUpdateSync" @click="toggleUseUpdateSync" /></div>
          </div>

          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('database-sync')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.metadataOnlyStorage') }}</div>
              <div class="settings-body">{{ $t('settings.metadataOnlyStorageDescription') }}</div>
            </div>
            <div class="settings-action fixed-width"><ToggleSwitch :switchValueProp="metadataOnlyStorage" @click="toggleMetadataOnlyStorage" /></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
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

<style scoped>
.fixed-width {
  min-width: 200px;
}
</style>
