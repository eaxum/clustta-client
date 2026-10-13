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
            v-else-if="item.id === 'interface-font'"
            :items="interfaceFontOptions"
            :onSelect="selectInterfaceFont"
            :selectedItem="settingsStore.interfaceFont"
            :placeHolder="DEFAULT_INTERFACE_FONT_NAME"
            :fixedWidth="true"
          />
          <DropDownBox
            v-else-if="item.id === 'text-size'"
            :items="textSizeOptions"
            :onSelect="selectTextSize"
            :selectedItem="String(settingsStore.interfaceTextSize)"
            :placeHolder="`${DEFAULT_INTERFACE_TEXT_SIZE} px`"
            :prioritizeSelected="false"
            :fixedWidth="true"
          />
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
import { useSettingsStore } from '@/stores/settings';
import { useThemeStore } from '@/stores/theme';
import { TINTS } from '@/theme/palette';

const MINIMUM_INTERFACE_TEXT_SIZE = 13;
const MEDIUM_INTERFACE_TEXT_SIZE = 14;
const MAXIMUM_INTERFACE_TEXT_SIZE = 15;
const DEFAULT_INTERFACE_TEXT_SIZE = MINIMUM_INTERFACE_TEXT_SIZE;
const DEFAULT_INTERFACE_FONT_NAME = 'Sora';
const INTERFACE_TEXT_SIZES = [
  MINIMUM_INTERFACE_TEXT_SIZE,
  MEDIUM_INTERFACE_TEXT_SIZE,
  MAXIMUM_INTERFACE_TEXT_SIZE,
];
const INTERFACE_FONTS = [
  { id: 'sora', name: 'Sora', selectionValue: 'sora' },
  { id: 'inter', name: 'Inter', selectionValue: 'inter' },
  { id: 'sour-gummy', name: 'Sour Gummy', selectionValue: 'sour-gummy' },
];

const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const settingsStore = useSettingsStore();
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
const textSizeOptions = INTERFACE_TEXT_SIZES.map((size) => ({
  id: String(size),
  name: `${size} px`,
  selectionValue: String(size),
}));
const interfaceFontOptions = INTERFACE_FONTS;
const themeIcon = computed(() => {
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
    id: 'interface-font',
    icon: getAppIcon('alphabets'),
    label: t('settings.interfaceFont'),
    description: t('settings.interfaceFontDescription'),
  },
  {
    id: 'text-size',
    icon: getAppIcon('text-size'),
    label: t('settings.textSize'),
    description: t('settings.textSizeDescription'),
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
const selectInterfaceFont = async (font) => {
  try {
    await settingsStore.setInterfaceFont(font);
  } catch (error) {
    console.error(error);
    notificationStore.addNotification(t('common.error'), t('notifications.failedToUpdate'), 'error');
  }
};
const selectTextSize = async (size) => {
  try {
    await settingsStore.setInterfaceTextSize(size);
  } catch (error) {
    console.error(error);
    notificationStore.addNotification(t('common.error'), t('notifications.failedToUpdate'), 'error');
  }
};
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
