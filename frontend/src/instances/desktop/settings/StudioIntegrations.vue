<template>
  <div class="settings-component-root">
    <div class="settings-component-scroll">
      <div class="settings-component-container">

        <div class="settings-section-card">
          <div class="settings-section-card-header">
            <h2 class="settings-section-card-title">{{ $t('settings.studioIntegrations') }}</h2>
          </div>

          <div class="settings-section-card-content">

            <div v-for="integration in studioIntegrations" :key="integration.id" class="settings-item" @click="openIntegration(integration.id)" v-stop-propagation>
              <div class="settings-icon"><img class="small-icons" :src="getAppIcon(integration.icon)"></div>

              <div class="settings-content">
                <div class="settings-header">{{ integration.name }}</div>
                <div class="settings-body">{{ rowBody(integration) }}</div>
                <div v-if="viewFor(integration.id)?.warning" class="integration-warning">{{ viewFor(integration.id).warning }}</div>
              </div>

              <div class="settings-action">
                <ActionButton v-if="viewFor(integration.id)?.configured" class="row-delete" :icon="getAppIcon('trash')" :label="$t('common.delete')" :showLabel="true" :buttonFunction="() => onDelete(integration.id)" useDanger useOutline v-stop-propagation />

                <div v-stop-propagation>
                  <ToggleSwitch v-if="viewFor(integration.id)?.configured" :switchValueProp="viewFor(integration.id)?.enabled" @click="toggleEnabled(integration.id)" />
                </div>

                <img class="small-icons chevron" :src="getAppIcon('chevron-right')">
              </div>
            </div>

            <div v-if="studioIntegrations.length === 0" class="empty-state">{{ $t('common.loading') }}...</div>

          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';

const desktopModals = useDesktopModalStore();
const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const projectStore = useProjectStore();
const trayStates = useTrayStates();
const { t } = useI18n();

// refs
const SUPPORTED_STUDIO_INTEGRATIONS = ['kitsu'];

// computed
const studioId = computed(() => projectStore.selectedStudio?.id || '');

const studioIntegrations = computed(() => {
  return SUPPORTED_STUDIO_INTEGRATIONS
    .map((id) => integrationStore.getIntegration(id))
    .filter((i) => !!i);
});

// methods
// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Returns the cached studio config view for the given integration id.
const viewFor = (id) => integrationStore.studioConfig[id] || null;

// Returns the badge class for a row based on its current status.
const statusClass = (id) => {
  const v = viewFor(id);
  if (!v?.configured) return 'status-stopped';
  if (v.last_error) return 'status-error';
  if (v.status === 'running') return 'status-running';
  return 'status-stopped';
};

// Returns the badge label for a row based on its current status.
const statusLabel = (id) => {
  const v = viewFor(id);
  if (!v?.configured) return t('settings.statusNotConfigured');
  if (v.last_error) return t('settings.statusError');
  return v.status === 'running' ? t('settings.statusRunning') : t('settings.statusStopped');
};

// Returns the descriptive body line for an integration row.
// Shows the live status when configured (Running/Stopped/Error), otherwise the integration description.
const rowBody = (integration) => {
  const v = viewFor(integration.id);
  if (v?.configured) return statusLabel(integration.id);
  return integration.description || '';
};

// Opens the studio integration config modal for the given integration id.
const openIntegration = (id) => {
  console.log('clicked')
  integrationStore.setActiveStudioIntegration(id);
  desktopModals.setModalVisibility('studioIntegrationConfigModal', true);
};

// Toggles the enabled flag of an already-configured integration.
const toggleEnabled = async (id) => {
  const v = viewFor(id);
  if (!v?.configured || !studioId.value) return;
  await integrationStore.setStudioIntegrationEnabled(studioId.value, id, !v.enabled);
};

// Opens the PopUpModal to confirm removal of an integration's stored credentials.
const onDelete = (id) => {
  if (!studioId.value) return;
  const integration = integrationStore.getIntegration(id);
  const name = integration?.name || id;
  trayStates.popUpModalTitle = t('common.delete') + ` "${name}"?`;
  trayStates.popUpModalMessage = t('settings.confirmDeleteIntegration');
  trayStates.popUpModalIcon = 'trash';
  trayStates.popUpModalFunction = async () => {
    try {
      await integrationStore.deleteStudioIntegrationConfig(studioId.value, id);
    } finally {
      desktopModals.setModalVisibility('popUpModal', false);
    }
  };
  desktopModals.setModalVisibility('popUpModal', true);
};

// Refreshes the cached studio config for all supported integrations.
const refresh = async () => {
  if (!studioId.value) return;
  await integrationStore.loadAvailableIntegrations();
  await Promise.all(
    SUPPORTED_STUDIO_INTEGRATIONS.map((id) =>
      integrationStore.loadStudioIntegrationConfig(studioId.value, id)
    )
  );
};

// watchers
watch(studioId, async (id) => {
  if (id) await refresh();
});

// lifecycle hooks
onMounted(async () => {
  await refresh();
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.settings-component-root {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: block;
  overflow-y: scroll;
  border-radius: var(--very-large-radius);
  box-sizing: border-box;
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
  border-bottom: 1px solid var(--light-steel);
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
  gap: .5rem;
}

.row-delete {
  opacity: 0;
  transition: opacity 0.15s ease;
}

.settings-item:hover .row-delete {
  opacity: 1;
}

.chevron {
  opacity: .6;
}

.status-badge {
  font-size: .7rem;
  padding: .25rem .6rem;
  border-radius: 999px;
  text-transform: uppercase;
  letter-spacing: .05em;
  white-space: nowrap;
}

.status-running { background-color: rgba(36, 129, 30, .25); color: #6ee07a; }
.status-stopped { background-color: rgba(255, 255, 255, .08); color: var(--white-half); }
.status-error { background-color: rgba(255, 80, 80, .15); color: #ff6b6b; }

.integration-warning {
  margin-top: .25rem;
  padding: .25rem .5rem;
  border-radius: var(--small-radius);
  background-color: rgba(255, 176, 32, .12);
  color: #ffb020;
  font-size: 11px;
  line-height: 1.3;
}

.empty-state {
  padding: 1rem;
  color: var(--white-half);
  font-size: .85rem;
  text-align: center;
}
</style>
