<template>
  <div class="settings-component-root user-settings-page">
    <div class="settings-component-container">
      <SettingsCard :title="$t('settings.appearance')" :items="appearanceItems">
        <template #action="{ item }">
          <DropDownBox
            v-if="item.id === 'icon-scheme'"
            :items="iconSchemeOptions"
            :onSelect="selectIconType"
            :selectedItem="iconStore.selectedIconType"
            :placeHolder="'None'"
            :fixedWidth="true"
          />
          <DropDownBox
            v-else-if="item.id === 'theme'"
            :items="themeOptions"
            :onSelect="selectTheme"
            :selectedItem="themeStore.mode"
            :placeHolder="'None'"
            :fixedWidth="true"
          />
          <div v-else-if="item.id === 'accent'" class="tint-swatches">
            <button
              v-for="tint in themeStore.availableTints"
              :key="tint"
              class="tint-swatch"
              :class="{ 'tint-swatch-active': themeStore.tint === tint }"
              :style="tintSwatchStyle(tint)"
              v-tooltip="$t(`settings.tints.${tint}`)"
              @click="selectTint(tint)"
            ></button>
          </div>
          <DropDownBox
            v-else-if="item.id === 'language'"
            :items="availableLanguages"
            :onSelect="selectLanguage"
            :selectedItem="currentLanguageName"
            :placeHolder="'None'"
            :fixedWidth="true"
          />
        </template>
      </SettingsCard>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import { useLocale } from '@/composables/useLocale';
import SettingsCard from '@/instances/desktop/settings/components/SettingsCard.vue';
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
const iconSchemeOptions = computed(() => iconStore.iconTypes.map((iconType) => ({
  id: iconType,
  name: t(`settings.iconSchemes.${iconType}`),
  selectionValue: iconType,
})));
const themeOptions = computed(() => themeStore.availableModes.map((mode) => ({
  id: mode,
  name: t(`settings.themeModes.${mode}`),
  selectionValue: mode,
})));
const themeIcon = computed(() => {
  if (themeStore.mode === 'system') return getAppIcon('palette');
  return themeStore.resolvedMode === 'dark' ? getAppIcon('moon') : getAppIcon('sun');
});
const appearanceItems = computed(() => [
  {
    id: 'icon-scheme',
    icon: getAppIcon('palette'),
    label: t('settings.iconScheme'),
    description: t('settings.iconSchemeDescription'),
  },
  {
    id: 'theme',
    icon: themeIcon.value,
    label: t('settings.theme'),
    description: t('settings.themeDescription'),
  },
  {
    id: 'accent',
    icon: getAppIcon('palette'),
    label: t('settings.accent'),
    description: t('settings.accentDescription'),
  },
  {
    id: 'language',
    icon: getAppIcon('translation'),
    label: t('settings.language'),
    description: t('settings.languageDescription'),
  },
]);

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
.tint-swatches {
  display: flex;
  align-items: center;
  gap: .4rem;
  width: min-content;
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
