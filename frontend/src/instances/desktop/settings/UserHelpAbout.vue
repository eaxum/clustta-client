<template>
  <div class="settings-component-root user-settings-page">
    <div class="settings-component-container">
      <SettingsCard :title="$t('settings.resourcesSupport')" :items="resourceItems">
        <template #action="{ item }">
          <ActionButton
            v-if="item.url"
            :icon="getAppIcon('square-arrow-right-up')"
            :buttonFunction="() => openResource(item.url)"
          />
          <ActionButton
            v-else
            :label="$t('common.send')"
            :showIcon="false"
            :buttonFunction="openDiagnosticsModal"
          />
        </template>
      </SettingsCard>

      <SettingsCard :title="$t('settings.about')" :items="aboutItems">
        <template #action="{ item }">
          <ActionButton
            :label="item.actionLabel"
            :showIcon="false"
            :buttonFunction="() => runAboutAction(item.id)"
          />
        </template>
      </SettingsCard>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { Browser } from '@wailsio/runtime';
import { useI18n } from 'vue-i18n';
import utils from '@/services/utils';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import SettingsCard from '@/instances/desktop/settings/components/SettingsCard.vue';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();
const modals = useDesktopModalStore();
const clusttaVersion = ref('');
const { t } = useI18n();
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

const resources = [
  { icon: 'book', labelKey: 'settings.documentation', descriptionKey: 'settings.documentationDescription', url: 'https://docs.clustta.com' },
  { icon: 'youtube', labelKey: 'settings.videoGuides', descriptionKey: 'settings.videoGuidesDescription', url: 'https://youtube.com/playlist?list=PLy9tuKQd1hzzuUktc6UVFUhhQxNQtkDqR&si=f2TQRtOYSHeqXma9' },
  { icon: 'help', labelKey: 'settings.communitySupport', descriptionKey: 'settings.communitySupportDescription', url: 'https://discord.gg/NuR4uAuTZd' },
  { icon: 'website', labelKey: 'settings.visitWebsite', descriptionKey: 'settings.visitWebsiteDescription', url: 'https://clustta.com/' },
];
const resourceItems = computed(() => [
  ...resources.map((resource) => ({
    id: resource.labelKey,
    icon: getAppIcon(resource.icon),
    label: t(resource.labelKey),
    description: t(resource.descriptionKey),
    url: resource.url,
  })),
  {
    id: 'feedback',
    icon: getAppIcon('megaphone'),
    label: t('settings.submitFeedback'),
    description: t('settings.submitFeedbackDescription'),
  },
]);
const aboutItems = computed(() => [
  {
    id: 'about-clustta',
    icon: getAppIcon('info'),
    label: t('settings.aboutClustta'),
    description: clusttaVersion.value,
    actionLabel: t('common.more'),
  },
  {
    id: 'open-source-licenses',
    icon: getAppIcon('scale'),
    label: t('settings.openSourceLicenses'),
    description: t('settings.openSourceLicensesDescription'),
    actionLabel: t('modals.viewFullNotices'),
  },
  {
    id: 'source-code',
    icon: getAppIcon('code-bracket'),
    label: t('settings.sourceCode'),
    description: t('settings.sourceCodeDescription'),
    actionLabel: t('common.open'),
  },
]);

const openResource = (url) => Browser.OpenURL(url);
const openDiagnosticsModal = () => modals.setModalVisibility('submitDiagnosticsModal', true);
const displayAppInfo = () => modals.setModalVisibility('appInfoModal', true);
const displayLicenses = () => modals.setModalVisibility('openSourceLicensesModal', true);

const runAboutAction = (itemId) => {
  if (itemId === 'about-clustta') {
    displayAppInfo();
    return;
  }
  if (itemId === 'open-source-licenses') {
    displayLicenses();
    return;
  }
  if (itemId === 'source-code') openResource('https://github.com/eaxum/clustta-client');
};

onMounted(async () => {
  clusttaVersion.value = await utils.getRawClusttaVersion();
});
</script>
