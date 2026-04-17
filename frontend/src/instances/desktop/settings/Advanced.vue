<template>
  <div class="settings-component-root">
    <div class="settings-component-scroll">
    <div class="settings-component-container">

      <!-- External Integrations Card -->
      <div v-if="entitlementStore.hasIntegrations" class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.externalIntegrations') }}</h2>
        </div>
        <div class="settings-section-card-content">

          <!-- Linked Integration -->
          <div v-if="linkedIntegration" v-stop-propagation class="settings-item" @click="openIntegrationLink">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon(linkedIntegration.integration_id)"></div>
            <div class="settings-content">
              <div class="settings-header">{{ linkedIntegration.external_project_name }}</div>
              <div class="settings-body">{{ $t('settings.linkedTo', { integration: linkedIntegration.integration_id }) }}</div>
            </div>
            <div class="settings-action" v-stop-propagation>
              <ActionButton :icon="getAppIcon('settings')" :label="$t('common.manage')" :buttonFunction="openIntegrationLink" />
            </div>
          </div>

          <!-- Directory Mapping (only when integration linked) -->
          <div v-if="linkedIntegration" v-stop-propagation class="settings-item" @click="openDirectoryMapping">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('file-path')"></div>
            <div class="settings-content">
              <div class="settings-header">Directory Mapping</div>
              <div class="settings-body">Configure folder structure for synced items</div>
            </div>
            <div class="settings-action" v-stop-propagation>
              <ActionButton :icon="getAppIcon('settings')" :label="'Configure'" :buttonFunction="openDirectoryMapping" />
            </div>
          </div>

          <!-- Asset Type Templates (only when integration linked) -->
          <div v-if="linkedIntegration" v-stop-propagation class="settings-item" @click="openAssetTypeMapping">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('extension')"></div>
            <div class="settings-content">
              <div class="settings-header">Asset Type Templates</div>
              <div class="settings-body">Map asset types to file templates</div>
            </div>
            <div class="settings-action" v-stop-propagation>
              <ActionButton :icon="getAppIcon('settings')" :label="'Configure'" :buttonFunction="openAssetTypeMapping" />
            </div>
          </div>

          <!-- Status Mapping (only when integration linked) -->
          <div v-if="linkedIntegration" v-stop-propagation class="settings-item" @click="openStatusMapping">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('clock')"></div>
            <div class="settings-content">
              <div class="settings-header">Status Mapping</div>
              <div class="settings-body">Map statuses to push on checkpoint</div>
            </div>
            <div class="settings-action" v-stop-propagation>
              <ActionButton :icon="getAppIcon('settings')" :label="'Configure'" :buttonFunction="openStatusMapping" />
            </div>
          </div>

          <!-- No Integration Linked -->
          <div v-else class="settings-item" v-stop-propagation @click="openIntegrationLink">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('plug')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.linkIntegration') }}</div>
              <div class="settings-body">{{ $t('settings.linkIntegrationDescription') }}</div>
            </div>
            <div class="settings-action">
              <ActionButton :icon="getAppIcon('link')" :label="$t('common.link')" :buttonFunction="openIntegrationLink" />
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

          <div class="settings-item" @click="toggleWriteThrough">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('arrow-big-up-lines')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.writeThroughSync') }}</div>
              <div class="settings-body">{{ $t('settings.writeThroughDescription') }}</div>
            </div>
            <div class="settings-action fixed-width">
              <ToggleSwitch :switchValueProp="writeThroughEnabled" />
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
import { useI18n } from 'vue-i18n';

// services
import { ProjectService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useEntitlementStore } from '@/stores/entitlements';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// refs
const desktopModals = useDesktopModalStore();
const entitlementStore = useEntitlementStore();
const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const writeThroughEnabled = ref(false);
const { t } = useI18n();

// computed
const linkedIntegration = computed(() => integrationStore.linkedIntegration);

// methods
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Opens the integration link modal to manage project integration.
const openIntegrationLink = () => {
  desktopModals.setModalVisibility('integrationLinkModal', true);
};

// Opens the directory mapping modal to configure folder structure.
const openDirectoryMapping = () => {
  desktopModals.setModalVisibility('directoryMappingModal', true);
};

// Opens the asset type mapping modal to configure template mappings.
const openAssetTypeMapping = () => {
  desktopModals.setModalVisibility('assetTypeMappingModal', true);
};

// Opens the status mapping modal to configure status sync.
const openStatusMapping = () => {
  desktopModals.setModalVisibility('statusMappingModal', true);
};

// Toggles the write-through sync experimental feature for the active project.
const toggleWriteThrough = () => {
  const projectUri = projectStore.activeProject?.uri;
  if (!projectUri) return;

  const newValue = !writeThroughEnabled.value;
  ProjectService.SetWriteThroughEnabled(projectUri, newValue).then(() => {
    writeThroughEnabled.value = newValue;
    notificationStore.addNotification(
      t('settings.writeThroughSync'),
      t('notifications.writeThroughToggled', { status: newValue ? 'enabled' : 'disabled' }),
      "success"
    );
  }).catch((error) => {
    console.log(error);
    notificationStore.addNotification(t('common.error'), t('notifications.failedToUpdateWriteThrough'), "error");
  });
};

// lifecycle hooks
onMounted(async () => {
  try {
    const projectUri = projectStore.activeProject?.uri;
    if (projectUri) {
      writeThroughEnabled.value = await ProjectService.GetWriteThroughEnabled(projectUri);
      await integrationStore.loadLinkedIntegration();
    }
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
  flex-direction: column;
  gap: 5px;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  display: block;
  overflow-y: scroll;
  border-radius: var(--very-large-radius);
}

.settings-component-root::-webkit-scrollbar {
  width: 6px;
}

.settings-component-root::-webkit-scrollbar-thumb {
  background-color: var(--midnight-steel);
  border-radius: 3px;
}

.settings-component-root::-webkit-scrollbar-track {
  background-color: var(--light-steel);
  border-radius: 3px;
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
  gap: 1.5rem;
  width: 100%;
  padding-right: .2rem;
  border-radius: var(--large-radius);
}

/* Settings item styling */
.settings-item {
  color: var(--white);
  box-sizing: border-box;
  overflow: hidden;
  min-height: 50px;
  display: flex;
  padding: .5rem 1rem;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: max-content;
  background-color: var(--dark-steel);
  cursor: pointer;
  transition: background-color 0.2s ease;
  border-bottom:  1px solid var(--light-steel);
}

.settings-item:hover {
  background-color: #ffffff15;
}

.settings-item:active {
  background-color: #00000013;
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
  color: var(--silver);
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
</style>
