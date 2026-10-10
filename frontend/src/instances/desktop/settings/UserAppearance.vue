<template>
  <div class="settings-component-root user-settings-page">
    <div class="settings-component-container">
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.appearance') }}</h2>
        </div>
        <div class="settings-section-card-content">
          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('palette')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.iconScheme') }}</div>
              <div class="settings-body">{{ $t('settings.iconSchemeDescription') }}</div>
            </div>
            <div class="settings-action fixed-width">
              <DropDownBox :items="iconStore.iconTypes" :onSelect="selectIconType" :selectedItem="iconStore.selectedIconType" :placeHolder="'None'" :fixedWidth="true" />
            </div>
          </div>

          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="themeIcon"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.theme') }}</div>
              <div class="settings-body">{{ $t('settings.themeDescription') }}</div>
            </div>
            <div class="settings-action fixed-width">
              <DropDownBox :items="themeStore.availableModes" :onSelect="selectTheme" :selectedItem="themeStore.mode" :placeHolder="'None'" :fixedWidth="true" />
            </div>
          </div>

          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('palette')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.accent') }}</div>
              <div class="settings-body">{{ $t('settings.accentDescription') }}</div>
            </div>
            <div class="settings-action tint-swatches">
              <button v-for="tint in themeStore.availableTints" :key="tint" class="tint-swatch" :class="{ 'tint-swatch-active': themeStore.tint === tint }" :style="tintSwatchStyle(tint)" v-tooltip="$t(`settings.tints.${tint}`)" @click="selectTint(tint)"></button>
            </div>
          </div>

          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('translation')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.language') }}</div>
              <div class="settings-body">{{ $t('settings.languageDescription') }}</div>
            </div>
            <div class="settings-action fixed-width">
              <DropDownBox :items="availableLanguages" :onSelect="selectLanguage" :selectedItem="currentLanguageName" :placeHolder="'None'" :fixedWidth="true" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import { useLocale } from '@/composables/useLocale';
import { SettingsService } from '@/services';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useThemeStore } from '@/stores/theme';
import { TINTS } from '@/theme/palette';

const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const themeStore = useThemeStore();
const { t, currentLanguage, languageNames, setLocale, getLocaleCode } = useLocale();

const availableLanguages = computed(() => Object.values(languageNames));
const currentLanguageName = computed(() => currentLanguage.value);
const themeIcon = computed(() => {
  if (themeStore.mode === 'system') return getAppIcon('palette');
  return themeStore.resolvedMode === 'dark' ? getAppIcon('moon') : getAppIcon('sun');
});

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

const selectIconType = async (iconType) => {
  await SettingsService.SetIconScheme(iconType);
  iconStore.selectedIconType = iconType;
};

const selectLanguage = async (languageName) => {
  const success = await setLocale(getLocaleCode(languageName));
  if (success) {
    notificationStore.addNotification(t('notifications.languageUpdated'), t('notifications.languageChanged', { language: languageName }), 'success');
    return;
  }
  notificationStore.addNotification(t('common.error'), t('notifications.errorOccurred'), 'error');
};

const selectTheme = (theme) => themeStore.setMode(theme);
const selectTint = (tint) => themeStore.setTint(tint);
const tintSwatchStyle = (tint) => {
  const config = TINTS[tint];
  if (!config) return {};
  const lightness = themeStore.resolvedMode === 'dark' ? 0.66 : 0.52;
  return { background: `oklch(${lightness} ${config.accentChroma} ${config.accentHue})` };
};
</script>

<style scoped>
.fixed-width {
  min-width: 200px;
}

.tint-swatches {
  display: flex;
  align-items: center;
  gap: .4rem;
  width: min-content;
  padding: .5rem;
}

.tint-swatch {
  width: 22px;
  height: 22px;
  padding: 0;
  border: 1px solid var(--border);
  border-radius: 40%;
  outline: 0 solid var(--accent);
  outline-offset: 0;
  cursor: pointer;
  transition: transform .1s ease, outline-offset .1s ease;
}

.tint-swatch:hover {
  transform: scale(1.1);
}

.tint-swatch-active {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}
</style>
