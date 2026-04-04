<template>
  <div class="settings-component-root">
    <div class="settings-component-scroll">
    <div class="settings-component-container">

      <!-- Studio Info Card -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.studioInfo') }}</h2>
        </div>
        <div class="settings-section-card-content">

          <div class="settings-item" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('stall')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.studioName') }}</div>
              <div class="settings-body">{{ studioInfo.name }}</div>
            </div>
          </div>

          <div v-if="!isCloudHosted" class="settings-item" @click="launchUpdateStudioModal()" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('website')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.ipAddressUrl') }}</div>
              <div class="settings-body">{{ studioInfo.url }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('chevron-right')"></div>
          </div>

          <div v-if="!isCloudHosted && studioInfo?.alt_url" class="settings-item" @click="launchUpdateStudioModal()" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('website')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.alternateUrl') }}</div>
              <div class="settings-body">{{ studioInfo?.alt_url }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('chevron-right')"></div>
          </div>

          <div v-if="!isCloudHosted" class="settings-item" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('clustta')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.clusttaServerVersion') }}</div>
              <div class="settings-body">{{ serverVersion || $t('common.loading') + '...' }}</div>
            </div>
          </div>

          <div v-if="studioInfo?.hosting_mode" class="settings-item" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('storefront')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.hostingMode') }}</div>
              <div class="settings-body">{{ studioInfo.hosting_mode }}</div>
            </div>
          </div>

          <div v-if="studioInfo?.id" class="settings-item" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('key')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.studioId') }}</div>
              <div class="settings-body">{{ studioInfo.id }}</div>
            </div>
          </div>

          <div v-if="studioEntitlements" class="settings-item" @click="openBillingPortal" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('star')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.currentPlan') }}</div>
              <div class="settings-body">{{ planLabel }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('chevron-right')"></div>
          </div>

        </div>
      </div>

      <!-- Usage & Limits Card -->
      <div v-if="studioEntitlements" class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.usageAndLimits') }}</h2>
        </div>
        <div class="settings-section-card-content">

          <div class="settings-item" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('floppy-disk')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.storageUsed') }}</div>
              <div class="settings-body">{{ storageLabel }}</div>
              <div v-if="storagePercent > 0" class="progress-bar-track">
                <div class="progress-bar-fill" :class="{ 'near-quota': storagePercent >= 90 }" :style="{ width: storagePercent + '%' }"></div>
              </div>
            </div>
          </div>

          <div class="settings-item" @click="navigateToCollaborators" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('two-persons')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.collaborators') }}</div>
              <div class="settings-body">{{ collaboratorsLabel }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('chevron-right')"></div>
          </div>

          <div v-if="studioEntitlements.limits?.ai_credits_monthly > 0" class="settings-item" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('brain')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.aiCredits') }}</div>
              <div class="settings-body">{{ aiCreditsLabel }}</div>
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
import { ref, computed, onMounted } from "vue";
import { useI18n } from 'vue-i18n';
import { StudioService } from "@/services";
import { Browser } from "@wailsio/runtime";

// services
import utils from '@/services/utils';

// stores/state imports
import { useProjectStore } from '@/stores/projects';
import { useIconStore } from '@/stores/icons';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useEntitlementStore } from '@/stores/entitlements';
import { useSettingsStore } from '@/stores/settings';
import { useStudioStore } from '@/stores/studio';

// stores
const entitlementStore = useEntitlementStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const settings = useSettingsStore();
const studioStore = useStudioStore();
const { t } = useI18n();

// refs
const serverVersion = ref("");

// computed
// Returns a formatted AI credits usage label.
const aiCreditsLabel = computed(() => {
  if (!studioEntitlements.value) return '';
  const used = studioEntitlements.value.usage?.ai_credits_used || 0;
  const limit = studioEntitlements.value.limits?.ai_credits_monthly || 0;
  return t('settings.aiCreditsOf', { used, limit });
});

// Returns a formatted collaborators usage label.
const collaboratorsLabel = computed(() => {
  const used = studioStore.studioUsers?.length || 0;
  const limit = studioEntitlements.value?.limits?.max_collaborators;
  if (!limit || limit <= 0) return String(used);
  return t('settings.collaboratorsOf', { used, limit });
});

// Returns whether the studio is cloud-hosted.
const isCloudHosted = computed(() => {
  return studioInfo.value?.hosting_mode === 'cloud';
});

// Returns a formatted plan label from a snake_case plan name.
const planLabel = computed(() => {
  if (!studioEntitlements.value) return '';
  const plan = studioEntitlements.value.plan || 'free';
  const name = plan.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
  return name;
});

// Returns the storage usage as a percentage (0–100).
const storagePercent = computed(() => {
  if (!studioEntitlements.value) return 0;
  const limit = studioEntitlements.value.limits?.storage_bytes;
  if (!limit || limit <= 0) return 0;
  const used = studioEntitlements.value.usage?.storage_bytes || 0;
  return Math.min((used / limit) * 100, 100);
});

// Returns a formatted storage usage label.
const storageLabel = computed(() => {
  if (!studioEntitlements.value) return '';
  const used = utils.formatBytes(studioEntitlements.value.usage?.storage_bytes || 0, 2);
  const limit = studioEntitlements.value.limits?.storage_bytes;
  if (!limit || limit <= 0) return used;
  return t('settings.storageOf', { used, limit: utils.formatBytes(limit, 0) });
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

// Fetches the studio server version.
const fetchServerVersion = async () => {
  try {
    const studioUrl = studioInfo.value?.url;
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

// Navigates to the Studio Collaborators tab.
const navigateToCollaborators = () => {
  settings.setModalVisibility('studiocollaborators', true);
};

// Opens the Stripe billing portal in the system browser.
const openBillingPortal = async () => {
  const portalUrl = await entitlementStore.openBillingPortal();
  if (portalUrl) {
    Browser.OpenURL(portalUrl);
  }
};

// lifecycle hooks
onMounted(async () => {
  if (!isCloudHosted.value) {
    await fetchServerVersion();
  }
  const studioId = studioInfo.value?.id;
  if (studioId) {
    await entitlementStore.fetchStudioEntitlements(studioId);
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
  display: block;
  overflow-y: scroll;
  border-radius: var(--very-large-radius);
}

.settings-component-root::-webkit-scrollbar {
  width: 4px;
}

.settings-component-root::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
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
}

.progress-bar-track {
  width: 100%;
  height: 4px;
  background-color: var(--light-steel);
  border-radius: 2px;
  margin-top: 0.3rem;
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background-color: var(--accent);
  border-radius: 2px;
  transition: width 0.3s ease;
}

.progress-bar-fill.near-quota {
  background-color: var(--error-red, #e05252);
}
</style>

