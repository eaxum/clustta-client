<template>
  <div class="settings-component-root">
    <div class="settings-component-scroll">
      <div class="settings-component-container">
        <SettingsCard :title="$t('settings.integrations')" :items="integrationItems">
          <template #action="{ item }">
            <span v-if="item.connected" class="connected-badge">{{ $t('common.connected') }}</span>
            <ActionButton
              :label="item.connected ? $t('common.manage') : $t('common.connect')"
              :showIcon="false"
              :buttonFunction="openIntegrationAuth"
            />
          </template>
        </SettingsCard>

        <SettingsCard :title="$t('settings.aiAgent')" :items="agentItems">
          <template #action>
            <ActionButton
              :label="$t('common.configure')"
              :showIcon="false"
              :buttonFunction="openAgentConfig"
            />
          </template>
        </SettingsCard>

        <SettingsCard :title="$t('settings.plugins')" :items="pluginItems">
          <template #action="{ item }">
            <ToggleSwitch
              v-if="item.id === 'bridge'"
              :switchValueProp="bridgeEnabled"
              @click="toggleBridgeEnabled"
            />
            <ActionButton
              v-else
              :label="$t('common.download')"
              :showIcon="false"
              :buttonFunction="openPluginsPage"
            />
          </template>
        </SettingsCard>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { Browser } from '@wailsio/runtime';
import { useI18n } from 'vue-i18n';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import SettingsCard from '@/instances/desktop/settings/components/SettingsCard.vue';
import { AgentService } from '@/services';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useNotificationStore } from '@/stores/notifications';
import { useSettingsStore } from '@/stores/settings';

const desktopModals = useDesktopModalStore();
const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const notificationStore = useNotificationStore();
const settingsStore = useSettingsStore();
const agentKeyConfigured = ref(false);
const { t } = useI18n();

const connectedIntegrations = computed(() => {
  return integrationStore.availableIntegrations.filter((integration) => {
    return integrationStore.isAuthenticated(integration.id);
  });
});
const bridgeEnabled = computed(() => settingsStore.bridgeEnabled);

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);
const integrationItems = computed(() => [
  ...connectedIntegrations.value.map((integration) => ({
    id: integration.id,
    icon: getAppIcon(integration.icon),
    label: integration.name,
    description: t('settings.connected'),
    connected: true,
  })),
  {
    id: 'connect-integration',
    icon: getAppIcon('plug'),
    label: t('settings.connectIntegration'),
    description: t('settings.connectIntegrationDescription'),
    connected: false,
  },
]);
const agentItems = computed(() => [
  {
    id: 'llm-provider',
    icon: getAppIcon('brain'),
    label: t('settings.llmProvider'),
    description: agentKeyConfigured.value
      ? t('settings.providerConfigured')
      : t('settings.configureProvider'),
  },
]);
const pluginItems = computed(() => [
  {
    id: 'bridge',
    icon: getAppIcon('brick'),
    label: bridgeEnabled.value ? t('settings.disableBridge') : t('settings.enableBridge'),
    description: t('settings.bridgeEnabledDescription'),
  },
  {
    id: 'download-plugins',
    icon: getAppIcon('download'),
    label: t('settings.downloadPlugins'),
    description: t('settings.downloadPluginsDescription'),
  },
]);

const openIntegrationAuth = () => {
  desktopModals.setModalVisibility('integrationAuthModal', true);
};

const openAgentConfig = () => desktopModals.setModalVisibility('configAgentModal', true);
const openPluginsPage = () => Browser.OpenURL('https://www.clustta.com/plugins');

const toggleBridgeEnabled = async () => {
  try {
    await settingsStore.toggleBridge();
    notificationStore.addNotification(
      t('settings.bridgeEnabled'),
      t('notifications.bridgeToggled', { status: settingsStore.bridgeEnabled ? 'enabled' : 'disabled' }),
      'success',
    );
  } catch (error) {
    console.error(error);
    notificationStore.addNotification(t('common.error'), t('notifications.failedToUpdateBridge'), 'error');
  }
};

onMounted(async () => {
  try {
    await integrationStore.initialize();
    await settingsStore.initializeBridge();
    const status = await AgentService.GetAPIKeyStatus();
    agentKeyConfigured.value = status.configured;
  } catch (error) {
    console.error('Unable to load user integrations:', error);
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.settings-component-root {
  display: block;
  width: 100%;
  height: 100%;
  overflow-y: auto;
  border-radius: var(--very-large-radius);
  box-sizing: border-box;
}

.settings-component-scroll {
  display: flex;
  justify-content: center;
}

.settings-component-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  gap: 1.5rem;
  padding-right: .2rem;
  box-sizing: border-box;
}

.connected-badge {
  padding: 4px 8px;
  border-radius: var(--small-radius);
  color: var(--accent-primary);
  background-color: rgba(var(--accent-primary-rgb), .15);
  font-size: 12px;
  font-weight: 500;
}
</style>
