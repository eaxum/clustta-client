<template>
  <div class="settings-component-root user-settings-page">
    <div class="settings-component-container">
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.resourcesSupport') }}</h2>
        </div>
        <div class="settings-section-card-content">
          <div v-for="resource in resources" :key="resource.labelKey" class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon(resource.icon)"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t(resource.labelKey) }}</div>
              <div class="settings-body">{{ $t(resource.descriptionKey) }}</div>
            </div>
            <div class="settings-action">
              <ActionButton :icon="getAppIcon('square-arrow-right-up')" :buttonFunction="() => openResource(resource.url)" />
            </div>
          </div>

          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('megaphone')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.submitFeedback') }}</div>
              <div class="settings-body">{{ $t('settings.submitFeedbackDescription') }}</div>
            </div>
            <div class="settings-action">
              <ActionButton :label="$t('common.send')" :showIcon="false" :buttonFunction="openDiagnosticsModal" />
            </div>
          </div>
        </div>
      </div>

      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.about') }}</h2>
        </div>
        <div class="settings-section-card-content">
          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('info')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.aboutClustta') }}</div>
              <div class="settings-body">{{ clusttaVersion }}</div>
            </div>
            <div class="settings-action">
              <ActionButton :label="$t('common.more')" :showIcon="false" :buttonFunction="displayAppInfo" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue';
import { Browser } from '@wailsio/runtime';
import utils from '@/services/utils';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();
const modals = useDesktopModalStore();
const clusttaVersion = ref('');

const resources = [
  { icon: 'book', labelKey: 'settings.documentation', descriptionKey: 'settings.documentationDescription', url: 'https://docs.clustta.com' },
  { icon: 'youtube', labelKey: 'settings.videoGuides', descriptionKey: 'settings.videoGuidesDescription', url: 'https://youtube.com/playlist?list=PLy9tuKQd1hzzuUktc6UVFUhhQxNQtkDqR&si=f2TQRtOYSHeqXma9' },
  { icon: 'help', labelKey: 'settings.communitySupport', descriptionKey: 'settings.communitySupportDescription', url: 'https://discord.gg/NuR4uAuTZd' },
  { icon: 'website', labelKey: 'settings.visitWebsite', descriptionKey: 'settings.visitWebsiteDescription', url: 'https://clustta.com/' },
];

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);
const openResource = (url) => Browser.OpenURL(url);
const openDiagnosticsModal = () => modals.setModalVisibility('submitDiagnosticsModal', true);
const displayAppInfo = () => modals.setModalVisibility('appInfoModal', true);

onMounted(async () => {
  clusttaVersion.value = await utils.getRawClusttaVersion();
});
</script>
