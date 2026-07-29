<template>
  <div class="settings-component-root">
    <div class="settings-component-scroll">
    <div class="settings-component-container">

      <!-- Studio Info -->
      <ProfileCard :title="$t('settings.studioInfo')" :showEditButton="!isCloudHosted" @toggleEdit="launchUpdateStudioModal()">
        <div class="header-layout">
          <div class="studio-avatar">
            <img class="studio-avatar-icon large-icons" :src="getAppIcon('stall')">
          </div>

          <div class="header-info">
            <div class="profile-name-row">
              <div class="profile-name">{{ studioInfo?.name }}</div>
            </div>

            <div v-if="studioEntitlements || studioInfo?.hosting_mode" class="meta-info">
              <span v-if="studioEntitlements" class="meta-badge">{{ planLabel }}</span>
              <span v-if="studioEntitlements && studioInfo?.hosting_mode" class="meta-dot">•</span>
              <span v-if="studioInfo?.hosting_mode" class="meta-badge">{{ studioInfo.hosting_mode }}</span>
            </div>

            <div class="studio-details">
              <div v-if="studioInfo?.id" class="info-item" @click="copyStudioId">
                <span>{{ studioInfo.id }}</span>
                <img class="info-icon small-icons" :src="getAppIcon(idCopied ? 'check-circle' : 'copy')">
              </div>

              <div v-if="!isCloudHosted" class="info-item">
                <img class="info-icon small-icons" :src="getAppIcon('website')">
                <span>{{ studioInfo?.url }}</span>
              </div>

              <div v-if="!isCloudHosted" class="info-item">
                <img class="info-icon small-icons" :src="getAppIcon('clustta')">
                <span>{{ serverVersion || $t('common.loading') + '...' }}</span>
              </div>
            </div>
          </div>
        </div>
      </ProfileCard>

      <!-- Metrics Row -->
      <div v-if="studioEntitlements" class="metrics-row">
        <MetricCard :title="$t('settings.collaborators')" :value="collaboratorsValue" :subtitle="collaboratorsLabel" :icon="getAppIcon('two-persons')" />

        <MetricCard :title="$t('settings.remoteProjects')" :value="projectsValue" :subtitle="projectsLabel" :icon="getAppIcon('briefcase')" />

        <MetricCard :title="$t('settings.storageUsed')" :value="storageValue" :subtitle="storageLabel" :icon="getAppIcon('floppy-disk')" :percent="storagePercent" :warning="storagePercent >= 90" />

        <MetricCard v-if="studioEntitlements.limits?.ai_credits_monthly > 0" :title="$t('settings.aiCredits')" :value="aiCreditsValue" :subtitle="aiCreditsLabel" :icon="getAppIcon('brain')" />
      </div>

      <!-- Administration Card -->
      <ProfileCard :title="$t('settings.administration')">
        <div class="admin-list">
          <div v-if="isCloudHosted" class="settings-item" @click="openBillingPortal">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('credit-card')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.billing') }}</div>
              <div class="settings-body">{{ $t('settings.manageBilling') }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('square-arrow-right-up')"></div>
          </div>

          <div class="settings-item disabled">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('file')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.auditLogs') }}</div>
              <div class="settings-body">{{ $t('settings.comingSoon') }}</div>
            </div>
          </div>
        </div>
      </ProfileCard>

      <ProfileCard v-if="showStorageConversions" :title="$t('settings.projectStorage')">
        <div class="storage-conversions">
          <p class="storage-description">{{ $t('settings.projectStorageDescription') }}</p>
          <div v-if="storageConversions.length === 0" class="storage-empty">{{ $t('settings.noStorageConversions') }}</div>
          <div v-for="conversion in storageConversions" :key="conversion.project_name" class="storage-project">
            <div class="storage-project-info">
              <div class="storage-project-title">
                <span>{{ conversion.project_name }}</span>
                <StatusBadge v-if="showStorageStatus(conversion)" :text="storageStatusLabel(conversion)" />
              </div>
              <div class="storage-project-meta">
                {{ $t('settings.currentStorageMode', { mode: storageModeLabel(conversion.current_mode) }) }}
                <span v-if="conversion.required_bytes > 0"> · {{ $t('settings.storageRequired', { size: utils.formatBytes(conversion.required_bytes, 1) }) }}</span>
              </div>
              <div v-if="conversion.status === 'running'" class="storage-progress-track">
                <div class="storage-progress-value" :style="{ width: storageProgress(conversion) + '%' }"></div>
              </div>
              <div v-if="conversion.status === 'running'" class="storage-progress-label">
                {{ storageProgress(conversion).toFixed(0) }}%
              </div>
              <div v-if="conversion.error" class="storage-error">{{ conversion.error }}</div>
            </div>
            <ActionButton
              :label="storageActionLabel(conversion)"
              :icon="getAppIcon('refresh')"
              :isDisabled="!canConvertStorage(conversion)"
              :buttonFunction="() => confirmStorageConversion(conversion)"
            />
          </div>
        </div>
      </ProfileCard>

      <!-- Danger Zone -->
      <ProfileCard :title="$t('stages.dangerZone')">
        <div class="danger-zone">
          <p class="danger-message">{{ $t('settings.deleteStudioWarning') }}</p>
          <ActionButton :iconAfter="true" :icon="getAppIcon('trash')" :label="$t('settings.deleteStudio')" @click="prepDeleteStudio()" :color="'crimson'" :useBackground="true" />
        </div>
      </ProfileCard>

    </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import { useI18n } from 'vue-i18n';
import { StudioService } from "@/services";
import { Browser } from "@wailsio/runtime";

// services
import utils from '@/services/utils';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import MetricCard from '@/instances/desktop/components/MetricCard.vue';
import ProfileCard from '@/instances/desktop/components/ProfileCard.vue';
import StatusBadge from '@/instances/common/components/StatusBadge.vue';

// stores
const entitlementStore = useEntitlementStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const settings = useSettingsStore();
const studioStore = useStudioStore();
const trayStates = useTrayStates();
const { t } = useI18n();

// stores/state imports
import { useProjectStore } from '@/stores/projects';
import { useIconStore } from '@/stores/icons';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useEntitlementStore } from '@/stores/entitlements';
import { useNotificationStore } from '@/stores/notifications';
import { useSettingsStore } from '@/stores/settings';
import { useStudioStore } from '@/stores/studio';
import { useTrayStates } from '@/stores/TrayStates';

// refs
const idCopied = ref(false);
const serverVersion = ref("");
const storageConversions = ref([]);
let storagePoll = null;
let storagePollingEnabled = true;
const activeStoragePollInterval = 1000;
const idleStoragePollInterval = 5000;

// computed

// Returns the AI credits usage count.
const aiCreditsValue = computed(() => {
  if (!studioEntitlements.value) return '0';
  return String(studioEntitlements.value.usage?.ai_credits_used || 0);
});

// Returns a formatted AI credits usage label.
const aiCreditsLabel = computed(() => {
  if (!studioEntitlements.value) return '';
  const limit = studioEntitlements.value.limits?.ai_credits_monthly || 0;
  return t('settings.aiCreditsOf', { used: aiCreditsValue.value, limit });
});

// Returns the collaborators count.
const collaboratorsValue = computed(() => {
  return String(studioStore.studioUsers?.length || 0);
});

// Returns a formatted collaborators usage label.
const collaboratorsLabel = computed(() => {
  const limit = studioEntitlements.value?.limits?.max_collaborators;
  if (!limit || limit <= 0) return '';
  return t('settings.collaboratorsOf', { used: collaboratorsValue.value, limit });
});

// Returns whether the studio is cloud-hosted.
const isCloudHosted = computed(() => {
  return studioInfo.value?.hosting_mode === 'cloud';
});

const storageCapabilities = computed(() => studioInfo.value?.capabilities?.project_storage || null);
const showStorageConversions = computed(() => {
  return !isCloudHosted.value && studioStore.isStudioAdmin && storageCapabilities.value?.conversion_supported === true;
});
const deflatedAvailable = computed(() => storageCapabilities.value?.available_modes?.includes('deflated'));

// Returns the project count.
const projectsValue = computed(() => {
  if (!studioEntitlements.value) return '0';
  if (studioEntitlements.value.usage_unavailable) return t('settings.unavailable');
  return String(studioEntitlements.value.usage?.project_count || 0);
});

// Returns a formatted projects usage label.
const projectsLabel = computed(() => {
  const limit = studioEntitlements.value?.limits?.max_remote_projects;
  if (!limit || limit <= 0) return '';
  return t('settings.projectsOf', { used: projectsValue.value, limit });
});

// Returns a formatted plan label from a snake_case plan name.
const planLabel = computed(() => {
  if (!studioEntitlements.value) return '';
  const plan = studioEntitlements.value.plan || 'free';
  return plan.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
});

// Returns the storage usage as a percentage (0–100).
const storagePercent = computed(() => {
  if (!studioEntitlements.value) return -1;
  if (studioEntitlements.value.usage_unavailable) return -1;
  const limit = studioEntitlements.value.limits?.storage_bytes;
  if (!limit || limit <= 0) return -1;
  const used = studioEntitlements.value.usage?.storage_bytes || 0;
  return Math.min((used / limit) * 100, 100);
});

// Returns the formatted storage usage value.
const storageValue = computed(() => {
  if (!studioEntitlements.value) return '0 B';
  if (studioEntitlements.value.usage_unavailable) return t('settings.unavailable');
  return utils.formatBytes(studioEntitlements.value.usage?.storage_bytes || 0, 2);
});

// Returns a formatted storage usage label.
const storageLabel = computed(() => {
  if (!studioEntitlements.value) return '';
  if (studioEntitlements.value.usage_unavailable) return '';
  const limit = studioEntitlements.value.limits?.storage_bytes;
  if (!limit || limit <= 0) return '';
  const used = studioEntitlements.value.usage?.storage_bytes || 0;
  const available = studioEntitlements.value.usage?.storage_available_bytes || Math.max(limit - used, 0);
  return `${utils.formatBytes(available, 0)} available`;
});

// Returns the cached entitlement bundle for the selected studio.
const studioEntitlements = computed(() => {
  const id = studioInfo.value?.id;
  if (!id) return null;
  return entitlementStore.studioEntitlements[id] || null;
});

// Returns the selected studio metadata.
const studioInfo = computed(() => {
  return projectStore.selectedStudio;
});

// methods

// Copies the studio ID to the clipboard.
const copyStudioId = async () => {
  if (!studioInfo.value?.id) return;
  await navigator.clipboard.writeText(studioInfo.value.id);
  notificationStore.addNotification(t('common.copied'), t('settings.studioIdCopied'), 'success');
  idCopied.value = true;
  setTimeout(() => { idCopied.value = false; }, 2000);
};

// Fetches the studio server version.
const fetchServerVersion = async () => {
  try {
    const studioUrl = await projectStore.resolveStudioUrl();
    if (!studioUrl) {
      serverVersion.value = t('settings.noStudioConnected');
      return;
    }
    const version = await StudioService.GetServerVersion(studioUrl);
    serverVersion.value = version || t('settings.unknown');
  } catch (error) {
    serverVersion.value = t('settings.unavailable');
  }
};

// Returns the resolved icon path for an icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Opens the update studio modal.
const launchUpdateStudioModal = () => {
  modals.setModalVisibility('updateStudioModal', true);
};

// Opens the Stripe billing portal in the system browser.
const openBillingPortal = async () => {
  const portalUrl = await entitlementStore.openBillingPortal();
  if (portalUrl) {
    Browser.OpenURL(portalUrl);
  }
};

// Prepares the delete studio confirmation.
const prepDeleteStudio = () => {
  modals.setModalVisibility('deleteStudioModal', true);
};

const fetchStorageConversions = async () => {
  if (!showStorageConversions.value) return;
  try {
    const studioUrl = await projectStore.resolveStudioUrl();
    storageConversions.value = await StudioService.GetStorageConversions(studioUrl) || [];
  } catch (error) {
    console.error('Failed to load project storage conversions:', error);
  } finally {
    scheduleStoragePoll();
  }
};

const storageModeLabel = (mode) => mode === 'deflated' ? t('settings.deflatedMode') : t('settings.compactMode');
const storageStatusLabel = (conversion) => {
  const labels = {
    idle: 'settings.storageReady',
    running: 'settings.storageConverting',
    failed: 'settings.storageFailed',
    cleanup_failed: 'settings.storageCleanupFailed',
    completed: 'settings.storageCompleted',
  };
  return t(labels[conversion.status] || 'settings.storageReady');
};
const showStorageStatus = (conversion) => {
  return ['running', 'failed', 'cleanup_failed'].includes(conversion.status);
};
const storageTargetMode = (conversion) => {
  if (conversion.status === 'failed' || conversion.status === 'cleanup_failed') return conversion.target_mode;
  return conversion.current_mode === 'compact' ? 'deflated' : 'compact';
};
const storageActionLabel = (conversion) => {
  if (conversion.status === 'running') return t('settings.storageConverting');
  if (conversion.status === 'failed' || conversion.status === 'cleanup_failed') return t('common.retry');
  return t('settings.convertStorageTo', { mode: storageModeLabel(storageTargetMode(conversion)) });
};
const canConvertStorage = (conversion) => {
  if (conversion.status === 'running') return false;
  const target = storageTargetMode(conversion);
  return target === 'compact' || deflatedAvailable.value;
};
const storageProgress = (conversion) => {
  if (!conversion.total_chunks) return 0;
  return Math.min(100, (conversion.processed_chunks / conversion.total_chunks) * 100);
};
const startStorageConversion = async (conversion) => {
  if (!canConvertStorage(conversion)) return;
  const target = storageTargetMode(conversion);
  const studioUrl = await projectStore.resolveStudioUrl();
  await StudioService.StartStorageConversion(
    studioUrl,
    conversion.project_name,
    target,
    '',
  );
  await fetchStorageConversions();
};
const confirmStorageConversion = (conversion) => {
  if (!canConvertStorage(conversion)) return;
  const target = storageTargetMode(conversion);
  trayStates.dangerousActionTitle = t('settings.storageConversionTitle', {
    project: conversion.project_name,
  });
  trayStates.dangerousActionMessage = t('settings.confirmStorageConversion', {
    project: conversion.project_name,
    source: storageModeLabel(conversion.current_mode),
    target: storageModeLabel(target),
  });
  trayStates.dangerousActionIcon = 'refresh';
  trayStates.dangerousActionConfirmText = conversion.project_name;
  trayStates.dangerousActionShowInput = true;
  trayStates.dangerousActionInputSecret = false;
  trayStates.dangerousActionRequireExactInput = true;
  trayStates.dangerousActionShowToggle = false;
  trayStates.dangerousActionFunction = () => startStorageConversion(conversion);
  modals.setModalVisibility('confirmDangerousActionModal', true);
};
const scheduleStoragePoll = () => {
  if (!storagePollingEnabled) return;
  if (storagePoll) window.clearTimeout(storagePoll);
  const hasActiveConversion = storageConversions.value.some(({ status }) => status === 'running');
  const interval = hasActiveConversion ? activeStoragePollInterval : idleStoragePollInterval;
  storagePoll = window.setTimeout(fetchStorageConversions, interval);
};

// lifecycle hooks
onMounted(async () => {
  if (!isCloudHosted.value && studioInfo.value?.name !== 'Personal') {
    try {
      await projectStore.ensureStudioCapabilities();
    } catch (error) {
      console.warn('Could not load Studio capabilities:', error);
    }
  }
  if (!isCloudHosted.value) {
    await fetchServerVersion();
  }
  const studioId = studioInfo.value?.id;
  if (studioId) {
    await studioStore.getStudioUsers();
    if (isCloudHosted.value) {
      await entitlementStore.fetchStudioEntitlements(studioId);
    } else {
      await entitlementStore.fetchPrivateStudioUsage(studioInfo.value);
    }
  }
  await fetchStorageConversions();
});

onBeforeUnmount(() => {
  storagePollingEnabled = false;
  if (storagePoll) window.clearTimeout(storagePoll);
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
  width: 4px;
}

.settings-component-root::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-4);
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
  gap: 1.5rem;
  width: 100%;
  padding-right: .2rem;
  border-radius: var(--large-radius);
}

/* Studio Info Header — matches UserProfile header-layout */
.header-layout {
  display: flex;
  flex-direction: row;
  gap: 1.5rem;
  align-items: flex-start;
}

.studio-avatar {
  width: 80px;
  height: 80px;
  min-width: 80px;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.05);
  display: flex;
  align-items: center;
  justify-content: center;
}

.studio-avatar-icon {
  width: 44px;
  height: 44px;
  opacity: 0.9;
}

.header-info {
  flex: 1;
  width: 100%;
}

.profile-name-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.profile-name {
  font-size: 2rem;
  font-weight: 500;
  margin: 0 0 0.25rem 0;
  color: var(--text);
}

.meta-info {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
  margin-bottom: 0.25rem;
}

.meta-badge {
  font-size: 0.875rem;
  color: var(--text-muted);
  text-transform: capitalize;
}

.meta-dot {
  font-size: 0.875rem;
  color: var(--text-muted);
  opacity: 0.6;
}

.studio-details {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--text-muted);
  font-size: 0.875rem;
  cursor: pointer;
}

.info-icon {
  width: 16px;
  height: 16px;
  opacity: 0.7;
}

/* Metrics Row */
.metrics-row {
  display: flex;
  gap: 1rem;
}

/* Administration */
.admin-list {
  display: flex;
  flex-direction: column;
  margin: -1.5rem;
}

.settings-item {
  color: var(--text);
  box-sizing: border-box;
  overflow: hidden;
  min-height: 50px;
  display: flex;
  padding: .5rem 1rem;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: max-content;
  cursor: pointer;
  transition: background-color 0.2s ease;
  border-bottom: 1px solid var(--surface-4);
}

.settings-item:last-child {
  border-bottom: none;
}

.settings-item:hover {
  background-color: #ffffff15;
}

.settings-item:active {
  background-color: #00000013;
}

.settings-item.disabled {
  cursor: default;
  opacity: 0.6;
}

.settings-item.disabled:hover {
  background-color: transparent;
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
  color: var(--text-muted);
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

.storage-conversions { display: flex; flex-direction: column; gap: .75rem; }
.storage-description, .storage-empty { margin: 0; color: var(--text-muted); font-size: .8rem; }
.storage-project { display: flex; align-items: center; gap: 1rem; padding: .75rem; border-radius: var(--normal-radius); background: var(--surface-3); }
.storage-project-info { flex: 1; min-width: 0; }
.storage-project-title { display: flex; align-items: center; gap: .5rem; font-size: .9rem; }
.storage-project-meta { margin-top: .25rem; color: var(--text-muted); font-size: .75rem; }
.storage-progress-track { height: 4px; margin-top: .6rem; overflow: hidden; border-radius: 4px; background: var(--surface-4); }
.storage-progress-value { height: 100%; background: var(--accent); transition: width .2s ease; }
.storage-progress-label { margin-top: .25rem; color: var(--text-muted); font-size: .7rem; text-align: right; }
.storage-error { margin-top: .4rem; color: var(--warning); font-size: .72rem; }

/* Danger Zone */
.danger-zone {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1rem;
  background-color: rgba(220, 38, 38, 0.1);
  border-radius: var(--normal-radius);
  border: 1px solid rgba(220, 38, 38, 0.3);
}

.danger-message {
  color: var(--text);
  margin: 0;
  font-size: 0.875rem;
}

</style>
