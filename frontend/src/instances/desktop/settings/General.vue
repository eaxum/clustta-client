<template>
  <div class="settings-component-root">
    <div class="settings-component-scroll">
    <div class="settings-component-container">

      <!-- Appearance Card -->
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
              <DropDownBox :items="iconStore.iconTypes" :onSelect="selectIconType"
                :selectedItem="iconStore.selectedIconType" :placeHolder="'None'" :fixedWidth="true" />
            </div>
          </div>

          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="themeStore.isDarkMode ? getAppIcon('moon') : getAppIcon('sun')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.theme') }}</div>
              <div class="settings-body">{{ $t('settings.themeDescription') }}</div>
            </div>
            <div class="settings-action fixed-width">
              <DropDownBox :items="themeStore.themes" :onSelect="selectTheme"
                :selectedItem="themeStore.currentTheme" :placeHolder="'None'" :fixedWidth="true" />
            </div>
          </div>

          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('translation')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.language') }}</div>
              <div class="settings-body">{{ $t('settings.languageDescription') }}</div>
            </div>
            <div class="settings-action fixed-width">
              <DropDownBox :items="availableLanguages" :onSelect="selectLanguage"
                :selectedItem="currentLanguageName" :placeHolder="'None'" :fixedWidth="true" />
            </div>
          </div>

          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon(defaultViewIcon)"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.defaultView') }}</div>
              <div class="settings-body">{{ $t('settings.defaultViewDescription') }}</div>
            </div>
            <div class="settings-action fixed-width">
              <DropDownBox :items="viewModeOptions" :onSelect="selectDefaultView"
                :selectedItem="currentViewLabel" :placeHolder="'None'" :fixedWidth="true" />
            </div>
          </div>
        </div>
      </div>

      <!-- Data Management Card -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.dataManagement') }}</h2>
        </div>
        <div class="settings-section-card-content">

          <div class="settings-item" @click="clearRecents">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('broom')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.clearRecents') }}</div>
              <div class="settings-body">{{ $t('settings.clearRecentsDescription') }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('chevron-right')"></div>
          </div>
          
        </div>
      </div>

      <!-- Resources & Support Card -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.resourcesSupport') }}</h2>
        </div>
        <div class="settings-section-card-content">
          <div class="settings-item" @click="Browser.OpenURL('https://docs.clustta.com')">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('book')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.documentation') }}</div>
              <div class="settings-body">{{ $t('settings.documentationDescription') }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('square-arrow-right-up')"></div>
          </div>

          <div class="settings-item" @click="Browser.OpenURL('https://youtube.com/playlist?list=PLy9tuKQd1hzzuUktc6UVFUhhQxNQtkDqR&si=f2TQRtOYSHeqXma9')">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('youtube')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.videoGuides') }}</div>
              <div class="settings-body">{{ $t('settings.videoGuidesDescription') }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('square-arrow-right-up')"></div>
          </div>

          <div class="settings-item" @click="Browser.OpenURL('https://discord.gg/NuR4uAuTZd')">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('help')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.communitySupport') }}</div>
              <div class="settings-body">{{ $t('settings.communitySupportDescription') }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('square-arrow-right-up')"></div>
          </div>

          <div class="settings-item" @click="Browser.OpenURL('https://clustta.com/')">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('website')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.visitWebsite') }}</div>
              <div class="settings-body">{{ $t('settings.visitWebsiteDescription') }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('square-arrow-right-up')"></div>
          </div>

          <div class="settings-item" @click="openDiagnosticsModal" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('megaphone')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.submitFeedback') }}</div>
              <div class="settings-body">{{ $t('settings.submitFeedbackDescription') }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('chevron-right')"></div>
          </div>
        </div>
      </div>

      <!-- About Card -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.about') }}</h2>
        </div>
        <div class="settings-section-card-content">
          <div class="settings-item" @click="displayAppInfo()" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('info')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.aboutClustta') }}</div>
              <div class="settings-body">{{ clusttaVersion }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('chevron-right')"></div>
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
import { SettingsService } from "@/services";
import { useLocale } from '@/composables/useLocale';

// services
import utils from '@/services/utils';

// stores/state imports
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { Browser } from "@wailsio/runtime";
import { useIconStore } from '@/stores/icons';
import { useThemeStore } from '@/stores/theme';
import { useCommonStore } from '@/stores/common';

// components
import DropDownBox from '@/instances/common/components/DropDownBox.vue';

// refs
const { t, currentLanguage, languageNames, setLocale, getLocaleCode } = useLocale();
const trayStates = useTrayStates();
const projectStore = useProjectStore();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const userStore = useUserStore();
const iconStore = useIconStore();
const themeStore = useThemeStore();
const commonStore = useCommonStore();
const autoStart = ref(trayStates.autoStart);
const clusttaVersion = ref("");

// computed
// Returns available languages for the dropdown derived from supported languages.
const availableLanguages = computed(() => {
  return Object.values(languageNames);
});

// Returns the current language name for display.
const currentLanguageName = computed(() => {
  return currentLanguage.value;
});

// Returns the display label for the current default view mode.
const currentViewLabel = computed(() => {
  const mode = commonStore.viewMode;
  if (mode === 'dense') return t('settings.compact');
  if (mode === 'grid') return t('settings.grid');
  return t('settings.list');
});

// Returns the icon for the current default view mode.
const defaultViewIcon = computed(() => {
  const mode = commonStore.viewMode;
  if (mode === 'dense') return 'list-compact';
  if (mode === 'grid') return 'four-squares';
  return 'list';
});

// Returns the available view mode options for the dropdown.
const viewModeOptions = computed(() => [
  t('settings.list'),
  t('settings.compact'),
  t('settings.grid'),
]);

// methods

const selectIconType = (iconType) => {
  SettingsService.SetIconScheme(iconType).then(() => {
    iconStore.selectedIconType = iconType;
  })

};

// Selects and applies the language preference.
const selectLanguage = async (languageName) => {
  const languageCode = getLocaleCode(languageName);
  
  const success = await setLocale(languageCode);
  if (success) {
    notificationStore.addNotification(
      t('notifications.languageUpdated'),
      t('notifications.languageChanged', { language: languageName }),
      "success"
    );
  } else {
    notificationStore.addNotification(
      t('common.error'),
      t('notifications.errorOccurred'),
      "error"
    );
  }
};

const selectTheme = (theme) => {
  SettingsService.SetTheme(theme).then(() => {
    themeStore.selectedTheme = theme;
    themeStore.isDarkMode = theme === 'dark'
    themeStore.applyTheme();
  })
};

// Selects and applies the default view mode.
const selectDefaultView = (viewLabel) => {
  const listLabel = t('settings.list');
  const compactLabel = t('settings.compact');
  let viewMode = 'compact';
  if (viewLabel === compactLabel) viewMode = 'dense';
  else if (viewLabel !== listLabel) viewMode = 'grid';

  SettingsService.SetDefaultViewMode(viewMode).then(() => {
    if (viewMode === 'compact') commonStore.setCompactView();
    else if (viewMode === 'dense') commonStore.setDenseView();
    else commonStore.setGridView();
    notificationStore.addNotification(
      t('notifications.defaultViewUpdated'),
      t('notifications.defaultViewSet', { viewType: viewLabel }),
      "success"
    );
  }).catch((error) => {
    console.log(error);
    notificationStore.addNotification(t('common.error'), t('notifications.failedToUpdate'), "error");
  });
};

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const displayAppInfo = () => {
  modals.setModalVisibility('appInfoModal', true);
};

const openDiagnosticsModal = () => {
  modals.setModalVisibility('submitDiagnosticsModal', true);
};

const launchDirConfigModal = () => {
  // modals.setModalVisibility('dirOnboardModal', true);
  modals.setModalVisibility('directoryConfigModal', true);
};

const clearRecents = () => {
  SettingsService.ClearRecentProject().then(() => {
    projectStore.recentProjects = []
    notificationStore.addNotification(
      t('notifications.recentProjectsCleared'),
      t('notifications.recentProjectsCleared'),
      "success"
    )
  })
};

// lifecycle hooks
onMounted(async () => {
  let user = userStore.user
  // trayStates.autoStart = await clusttaSettings.isAutoStart(user.username);
  // console.log(user);
  autoStart.value = trayStates.autoStart;
  clusttaVersion.value = await utils.getRawClusttaVersion();
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

