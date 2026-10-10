<template>
  <div class="settings-component-root">
    <div class="settings-component-scroll">
      <div class="settings-component-container">
        <div
          v-if="entitlementStore.hasIntegrations && userStore.canDo('manage_integrations')"
          class="settings-section-card"
        >
          <div class="settings-section-card-header integration-card-header">
            <div class="integration-card-title-group">
              <h2 class="settings-section-card-title">{{ $t('settings.externalIntegrations') }}</h2>
              <div class="settings-body integration-card-help">
                {{ $t('settings.linkIntegrationDescription') }}
              </div>
            </div>
            <ActionButton
              :icon="getAppIcon('link')"
              :label="$t('common.link')"
              :buttonFunction="openIntegrationLink"
            />
          </div>

          <div v-if="linkedIntegration" class="settings-section-card-content">
            <div class="settings-item">
              <div class="settings-icon">
                <img class="small-icons" :src="getAppIcon(linkedIntegration.integration_id)" alt="">
              </div>
              <div class="settings-content">
                <div class="settings-header">{{ linkedIntegration.external_project_name }}</div>
                <div class="settings-body">
                  {{ $t('settings.linkedTo', { integration: linkedIntegration.integration_id }) }}
                </div>
              </div>
              <div class="settings-action" v-stop-propagation>
                <ActionButton
                  :label="$t('common.manage')"
                  :showIcon="false"
                  :buttonFunction="openIntegrationLink"
                />
              </div>
            </div>

            <IntegrationSettingsRow
              icon="file-path"
              label="Directory Mapping"
              description="Configure folder structure for synced items"
              :open="openDirectoryMapping"
            />
            <IntegrationSettingsRow
              icon="extension"
              label="Asset Type Mapping"
              description="Map asset types to file templates"
              :open="openAssetTypeMapping"
            />
            <IntegrationSettingsRow
              icon="clock"
              label="Status Mapping"
              description="Map statuses to push on checkpoint"
              :open="openStatusMapping"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import IntegrationSettingsRow from '@/instances/desktop/settings/components/IntegrationSettingsRow.vue';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useEntitlementStore } from '@/stores/entitlements';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useUserStore } from '@/stores/users';

const desktopModals = useDesktopModalStore();
const entitlementStore = useEntitlementStore();
const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const userStore = useUserStore();

const linkedIntegration = computed(() => integrationStore.linkedIntegration);
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

const openIntegrationLink = () => {
  desktopModals.setModalVisibility('integrationLinkModal', true);
};

const openDirectoryMapping = () => {
  desktopModals.setModalVisibility('directoryMappingModal', true);
};

const openAssetTypeMapping = () => {
  desktopModals.setModalVisibility('assetTypeMappingModal', true);
};

const openStatusMapping = () => {
  desktopModals.setModalVisibility('statusMappingModal', true);
};

onMounted(async () => {
  try {
    await integrationStore.loadLinkedIntegration();
  } catch (error) {
    console.error('Unable to load project integration:', error);
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

.integration-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.integration-card-title-group {
  min-width: 0;
}

.integration-card-help {
  padding: .15rem 0 0;
}

.settings-item {
  display: flex;
  align-items: center;
  width: 100%;
  min-height: 50px;
  padding: .5rem 1rem;
  border-bottom: 1px solid var(--surface-4);
  color: var(--text);
  background-color: var(--surface-2);
  cursor: default;
  box-sizing: border-box;
}

.settings-icon {
  display: flex;
  padding: .3rem;
}

.settings-content {
  display: flex;
  flex: 1;
  flex-direction: column;
  padding: .4rem .2rem;
}

.settings-header {
  padding: .1rem;
  font-size: 14px;
  font-weight: 400;
}

.settings-body {
  padding: .1rem;
  color: var(--text-muted);
  font-size: 12px;
}

.settings-action {
  display: flex;
  align-items: center;
  justify-content: flex-end;
}
</style>
