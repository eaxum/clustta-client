<template>
  <div class="settings-component-root">
    <div class="settings-component-scroll">
      <div class="settings-component-container">
        <SettingsCard
          v-if="entitlementStore.hasIntegrations && userStore.canDo('manage_integrations')"
          :title="$t('settings.externalIntegrations')"
          :description="$t('settings.linkIntegrationDescription')"
          :items="integrationItems"
          :showContent="!!linkedIntegration"
        >
          <template #header-action>
            <ActionButton
              :icon="getAppIcon('link')"
              :label="$t('common.link')"
              :buttonFunction="openIntegrationLink"
            />
          </template>

          <template #action="{ item }">
            <ActionButton
              :label="item.id === 'linked-project' ? $t('common.manage') : $t('common.configure')"
              :showIcon="false"
              :buttonFunction="item.open"
              v-stop-propagation
            />
          </template>
        </SettingsCard>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import SettingsCard from '@/instances/desktop/settings/components/SettingsCard.vue';
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
const { t } = useI18n();

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
const integrationItems = computed(() => {
  if (!linkedIntegration.value) return [];
  return [
    {
      id: 'linked-project',
      icon: getAppIcon(linkedIntegration.value.integration_id),
      label: linkedIntegration.value.external_project_name,
      description: t('settings.linkedTo', {
        integration: linkedIntegration.value.integration_id,
      }),
      open: openIntegrationLink,
    },
    {
      id: 'directory-mapping',
      icon: getAppIcon('file-path'),
      label: 'Directory Mapping',
      description: 'Configure folder structure for synced items',
      open: openDirectoryMapping,
    },
    {
      id: 'asset-type-mapping',
      icon: getAppIcon('extension'),
      label: 'Asset Type Mapping',
      description: 'Map asset types to file templates',
      open: openAssetTypeMapping,
    },
    {
      id: 'status-mapping',
      icon: getAppIcon('clock'),
      label: 'Status Mapping',
      description: 'Map statuses to push on checkpoint',
      open: openStatusMapping,
    },
  ];
});

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

</style>
