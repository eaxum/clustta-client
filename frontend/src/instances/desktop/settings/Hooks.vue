<template>
  <div class="settings-component-root">
    <div class="settings-component-container">
      <ActionBar :itemType="$t('settings.addLaunchHook')" :addFunction="addLaunchHook" />

      <div v-if="settingsStore.preLaunchHooks.length" class="hooks-list-wrapper">
        <div class="hooks-list">
          <PreLaunchHookItem v-for="hook in settingsStore.preLaunchHooks" :key="hook.id" :hook="hook"
            :onEdit="editLaunchHook" :onDelete="confirmDeleteLaunchHook" :onToggle="toggleLaunchHook" />
        </div>
      </div>

      <PageState v-else :message="$t('settings.noLaunchHooks')" illustration="/page-states/resources.png"
        :secondaryIcon="getAppIcon('plus-circle')" :secondaryActionMessage="$t('settings.addLaunchHook')"
        :secondaryActionFunction="addLaunchHook" />
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue';
import { useI18n } from 'vue-i18n';

import PageState from '@/instances/common/components/PageState.vue';
import ActionBar from '@/instances/desktop/components/ActionBar.vue';
import PreLaunchHookItem from '@/instances/desktop/components/PreLaunchHookItem.vue';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useSettingsStore } from '@/stores/settings';
import { useTrayStates } from '@/stores/TrayStates';

const desktopModals = useDesktopModalStore();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const settingsStore = useSettingsStore();
const trayStates = useTrayStates();
const { t } = useI18n();

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

const addLaunchHook = () => {
  settingsStore.selectedPreLaunchHook = null;
  desktopModals.setModalVisibility('preLaunchHookModal', true);
};

const editLaunchHook = (hookId) => {
  settingsStore.selectedPreLaunchHook = settingsStore.preLaunchHooks.find((hook) => hook.id === hookId) || null;
  if (settingsStore.selectedPreLaunchHook) {
    desktopModals.setModalVisibility('preLaunchHookModal', true);
  }
};

const toggleLaunchHook = async (hookId) => {
  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath) return;
  const hooks = settingsStore.preLaunchHooks.map((hook) => (
    hook.id === hookId ? { ...hook, enabled: !hook.enabled } : hook
  ));
  try {
    await settingsStore.savePreLaunchHooks(projectPath, hooks);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorSavingLaunchHook'), error);
  }
};

const deleteLaunchHook = async (hookId) => {
  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath) return;
  const hooks = settingsStore.preLaunchHooks.filter((hook) => hook.id !== hookId);
  try {
    await settingsStore.savePreLaunchHooks(projectPath, hooks);
    notificationStore.addNotification(t('notifications.launchHookDeleted'), '', 'success');
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorDeletingLaunchHook'), error);
    throw error;
  }
};

const confirmDeleteLaunchHook = (hookId) => {
  const hook = settingsStore.preLaunchHooks.find((item) => item.id === hookId);
  if (!hook) return;
  trayStates.dangerousActionTitle = t('settings.deleteLaunchHookTitle', { name: hook.name });
  trayStates.dangerousActionMessage = t('settings.deleteLaunchHookMessage');
  trayStates.dangerousActionIcon = 'trash';
  trayStates.dangerousActionConfirmLabel = t('common.delete');
  trayStates.dangerousActionConfirmText = '';
  trayStates.dangerousActionShowInput = false;
  trayStates.dangerousActionShowToggle = false;
  trayStates.dangerousActionFunction = () => deleteLaunchHook(hookId);
  desktopModals.setModalVisibility('confirmDangerousActionModal', true);
};

onMounted(async () => {
  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath) return;
  try {
    await settingsStore.loadPreLaunchHooks(projectPath);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorLoadingLaunchHooks'), error);
  }
});
</script>

<style scoped>
.settings-component-root {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
}

.settings-component-container {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  height: 100%;
  overflow: hidden;
  width: 96%;
  gap: .5rem;
  align-items: center;
  justify-content: space-between;
  padding: 1rem;
  color: var(--text);
  background-color: var(--surface-1);
  border-radius: var(--very-large-radius);
}

.hooks-list-wrapper {
  width: 100%;
  min-height: 0;
  flex: 1;
  overflow-y: auto;
}

.hooks-list-wrapper::-webkit-scrollbar {
  width: 4px;
}

.hooks-list-wrapper::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-4);
}

.hooks-list {
  display: flex;
  flex-direction: column;
  gap: .5rem;
  width: 100%;
  box-sizing: border-box;
}
</style>
