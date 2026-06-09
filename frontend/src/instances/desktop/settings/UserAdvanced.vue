<template>
  <div class="settings-component-root">
    <div class="settings-component-scroll">
    <div class="settings-component-container">

      <!-- AI Agent Card -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.aiAgent') }}</h2>
        </div>
        <div class="settings-section-card-content">

          <div class="settings-item" v-stop-propagation @click="openAgentConfig">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('brain')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.llmProvider') }}</div>
              <div class="settings-body">{{ agentKeyConfigured ? $t('settings.providerConfigured') : $t('settings.configureProvider') }}</div>
            </div>
            <div class="settings-action">
              <img class="small-icons" :src="getAppIcon('chevron-right')">
            </div>
          </div>

        </div>
      </div>

      <!-- Behaviour Card -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.behaviour') }}</h2>
        </div>
        <div class="settings-section-card-content">

          <div class="settings-item" @click="toggleMinimizeOnClose">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('minimize')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.minimizeOnClose') }}</div>
              <div class="settings-body">{{ $t('settings.minimizeOnCloseDescription') }}</div>
            </div>
            <div class="settings-action fixed-width">
              <ToggleSwitch :switchValueProp="minimizeOnClose" />
            </div>
          </div>

        </div>
      </div>

      <!-- Integrations Card -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.integrations') }}</h2>
        </div>
        <div class="settings-section-card-content">

          <!-- Connected Integrations List -->
          <div v-for="integration in connectedIntegrations" :key="integration.id" class="settings-item" @click="openIntegrationAuth">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon(integration.icon)"></div>
            <div class="settings-content">
              <div class="settings-header">{{ integration.name }}</div>
              <div class="settings-body">{{ $t('settings.connected') }}</div>
            </div>
            <div class="settings-action">
              <span class="connected-badge">{{ $t('common.connected') }}</span>
            </div>
          </div>

          <!-- Connect New Integration -->
          <div class="settings-item" v-stop-propagation @click="openIntegrationAuth">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('plug')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.connectIntegration') }}</div>
              <div class="settings-body">{{ $t('settings.connectIntegrationDescription') }}</div>
            </div>
            <div class="settings-action">
              <img class="small-icons" :src="getAppIcon('chevron-right')">
            </div>
          </div>

        </div>
      </div>

      <!-- Plugins Card -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.plugins') }}</h2>
        </div>
        <div class="settings-section-card-content">

          <div class="settings-item" @click="toggleBridgeEnabled">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('brick')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ bridgeEnabled ? $t('settings.disableBridge') : $t('settings.enableBridge') }}</div>
              <div class="settings-body">{{ $t('settings.bridgeEnabledDescription') }}</div>
            </div>
            <div class="settings-action fixed-width">
              <ToggleSwitch :switchValueProp="bridgeEnabled" />
            </div>
          </div>

          <div class="settings-item" @click="Browser.OpenURL('https://www.clustta.com/plugins')">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('download')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.downloadPlugins') }}</div>
              <div class="settings-body">{{ $t('settings.downloadPluginsDescription') }}</div>
            </div>
            <div class="settings-action">
              <img class="small-icons" :src="getAppIcon('chevron-right')">
            </div>
          </div>

        </div>
      </div>

      <!-- Experimental Features Card -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.experimentalFeatures') }}</h2>
        </div>
        <div class="settings-section-card-content">

          <div class="settings-item" @click="toggleSyncAfterCheckpoint">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('refresh')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.syncAfterCheckpoint') }}</div>
              <div class="settings-body">{{ $t('settings.syncAfterCheckpointDescription') }}</div>
            </div>
            <div class="settings-action fixed-width">
              <ToggleSwitch :switchValueProp="syncAfterCheckpoint" />
            </div>
          </div>

        </div>
      </div>

    </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref, onMounted } from 'vue';
import { Browser } from '@wailsio/runtime';
import { useI18n } from 'vue-i18n';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// services
import { AgentService, SettingsService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useNotificationStore } from '@/stores/notifications';
import { useSettingsStore } from '@/stores/settings';

// refs
const agentKeyConfigured = ref(false);
const desktopModals = useDesktopModalStore();
const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const notificationStore = useNotificationStore();
const settingsStore = useSettingsStore();
const syncAfterCheckpoint = ref(false);
const { t } = useI18n();

// computed
// Returns the current bridge enabled state from the shared store.
const bridgeEnabled = computed(() => settingsStore.bridgeEnabled);
// Returns the current minimize on close state from the shared store.
const minimizeOnClose = computed(() => settingsStore.minimizeOnClose);
// Returns list of integrations user has authenticated with.
const connectedIntegrations = computed(() => {
  return integrationStore.availableIntegrations.filter(i => integrationStore.isAuthenticated(i.id));
});

// methods
// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Opens the AI agent configuration modal.
const openAgentConfig = () => {
  desktopModals.setModalVisibility('configAgentModal', true);
};

// Opens the integration authentication modal.
const openIntegrationAuth = () => {
  desktopModals.setModalVisibility('integrationAuthModal', true);
};

// Toggles the bridge HTTP server for DCC plugin integrations.
const toggleBridgeEnabled = () => {
  settingsStore.toggleBridge().then(() => {
    notificationStore.addNotification(
      t('settings.bridgeEnabled'),
      t('notifications.bridgeToggled', { status: settingsStore.bridgeEnabled ? 'enabled' : 'disabled' }),
      "success"
    );
  }).catch((error) => {
    console.log(error);
    notificationStore.addNotification(t('common.error'), t('notifications.failedToUpdateBridge'), "error");
  });
};

// Toggles the minimize-on-close behaviour.
const toggleMinimizeOnClose = () => {
  settingsStore.toggleMinimizeOnClose().then(() => {
    notificationStore.addNotification(
      t('settings.minimizeOnClose'),
      t('notifications.minimizeOnCloseToggled', { status: settingsStore.minimizeOnClose ? 'enabled' : 'disabled' }),
      "success"
    );
  }).catch((error) => {
    console.log(error);
    notificationStore.addNotification(t('common.error'), t('notifications.failedToUpdateMinimizeOnClose'), "error");
  });
};

// Toggles the sync-after-checkpoint default for the current user.
const toggleSyncAfterCheckpoint = () => {
  const newValue = !syncAfterCheckpoint.value;
  SettingsService.SetSyncAfterCheckpoint(newValue).then(() => {
    syncAfterCheckpoint.value = newValue;
    notificationStore.addNotification(
      t('settings.syncAfterCheckpoint'),
      t('notifications.syncAfterCheckpointToggled', { status: newValue ? 'enabled' : 'disabled' }),
      "success"
    );
  }).catch((error) => {
    console.log(error);
    notificationStore.addNotification(t('common.error'), t('notifications.failedToUpdateSyncAfterCheckpoint'), "error");
  });
};

// lifecycle hooks
onMounted(async () => {
  try {
    await settingsStore.initializeBridge();
    await settingsStore.initializeMinimizeOnClose();
    syncAfterCheckpoint.value = await SettingsService.GetSyncAfterCheckpoint();
    await integrationStore.initialize();
    const status = await AgentService.GetAPIKeyStatus();
    agentKeyConfigured.value = status.configured;
  } catch (error) {
    console.log(error);
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.settings-component-root {
  width: 100%;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
  display: block;
  overflow-y: scroll;
}

.settings-component-root::-webkit-scrollbar {
  width: 4px;
}

.settings-component-root::-webkit-scrollbar-thumb {
  background-color: hsl(var(--border));
  border-radius: var(--small-radius);
}

.settings-component-root::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.settings-component-scroll {
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.settings-component-container {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  height: 100%;
  gap: 1rem;
  width: 100%;
  padding-right: .2rem;
}

.settings-item {
  color: hsl(var(--foreground));
  box-sizing: border-box;
  overflow: hidden;
  min-height: 44px;
  display: flex;
  padding: .5rem 0.75rem;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: max-content;
  background-color: transparent;
  cursor: pointer;
  transition: background-color 0.15s ease;
  border-radius: calc(var(--radius) - 2px);
}

.settings-item:hover {
  background-color: hsl(var(--accent));
}

.settings-item:active {
  background-color: hsl(var(--accent));
}

.settings-icon {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  overflow: hidden;
  height: 100%;
  padding: .3rem;
  width: max-content;
}

.settings-content {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  overflow: hidden;
  height: 100%;
  padding: .4rem .2rem;
  flex: 1;
}

.settings-header {
  padding: .1rem;
  font-size: 14px;
  font-weight: 400;
}

.settings-body {
  color: hsl(var(--muted-foreground));
  padding: .1rem;
  font-size: 12px;
  opacity: .8;
}

.settings-action {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  overflow: hidden;
  height: 100%;
  width: max-content;
}

.fixed-width {
  min-width: 200px;
}

.connected-badge {
  color: var(--accent-primary);
  font-size: 12px;
  font-weight: 500;
  padding: 4px 8px;
  border-radius: var(--small-radius);
  background-color: rgba(var(--accent-primary-rgb), 0.15);
}
</style>
