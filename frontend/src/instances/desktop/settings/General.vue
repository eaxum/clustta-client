<template>
  <div class="settings-component-root">
    <div class="settings-component-container">

      <!-- Appearance Card -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">Appearance</h2>
        </div>
        <div class="settings-section-card-content">
          
          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('palette')"></div>
            <div class="settings-content">
              <div class="settings-header">Icon scheme</div>
              <div class="settings-body">Toggle between different icon styles for the user interface.</div>
            </div>
            <div class="settings-action fixed-width">
              <DropDownBox :items="iconStore.iconTypes" :onSelect="selectIconType"
                :selectedItem="iconStore.selectedIconType" :placeHolder="'None'" :fixedWidth="true" />
            </div>
          </div>

          <div class="settings-item">
            <div class="settings-icon"><img class="small-icons" :src="themeStore.isDarkMode ? getAppIcon('moon') : getAppIcon('sun')"></div>
            <div class="settings-content">
              <div class="settings-header">Theme</div>
              <div class="settings-body">Light or Dark mode.</div>
            </div>
            <div class="settings-action fixed-width">
              <DropDownBox :items="themeStore.themes" :onSelect="selectTheme"
                :selectedItem="themeStore.currentTheme" :placeHolder="'None'" :fixedWidth="true" />
            </div>
          </div>

          <div class="settings-item" @click="toggleUseGrid">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon(commonStore.useGrid ? 'four-squares' : 'list-compact')"></div>
            <div class="settings-content">
              <div class="settings-header">Default View</div>
              <div class="settings-body">Choose between grid or list view as default.</div>
            </div>
            <div class="settings-action fixed-width">
              <ToggleSwitch :switchValueProp="commonStore.useGrid" />
            </div>
          </div>
        </div>
      </div>

      <!-- Data Management Card -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">Data Management</h2>
        </div>
        <div class="settings-section-card-content">
          <div class="settings-item" @click="clearRecents">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('broom')"></div>
            <div class="settings-content">
              <div class="settings-header">Clear recents</div>
              <div class="settings-body">Clear recent projects from the side pane.</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('chevron-right')"></div>
          </div>
        </div>
      </div>

      <!-- Resources & Support Card -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">Resources & Support</h2>
        </div>
        <div class="settings-section-card-content">
          <div class="settings-item" @click="Browser.OpenURL('https://docs.clustta.com')">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('help')"></div>
            <div class="settings-content">
              <div class="settings-header">Help</div>
              <div class="settings-body">Access the community and help docs.</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('square-arrow-right-up')"></div>
          </div>

          <div class="settings-item" @click="Browser.OpenURL('https://clustta.com/')">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('website')"></div>
            <div class="settings-content">
              <div class="settings-header">Visit Website</div>
              <div class="settings-body">Go to Clustta's website</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('square-arrow-right-up')"></div>
          </div>
        </div>
      </div>

      <!-- About Card -->
      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">About</h2>
        </div>
        <div class="settings-section-card-content">
          <div class="settings-item" @click="displayAppInfo()" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('info')"></div>
            <div class="settings-content">
              <div class="settings-header">About Clustta</div>
              <div class="settings-body">{{ clusttaVersion }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('chevron-right')"></div>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { ref, computed, onMounted } from "vue";
import { SettingsService } from "@/../bindings/clustta/services/index";

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
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// refs
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

const selectIconType = (iconType) => {
  SettingsService.SetIconScheme(iconType).then(() => {
    iconStore.selectedIconType = iconType;
  })

};

const selectTheme = (theme) => {
  SettingsService.SetTheme(theme).then(() => {
    themeStore.selectedTheme = theme;
    themeStore.isDarkMode = theme === 'dark'
    themeStore.applyTheme();
  })
};

const toggleUseGrid = () => {
  const newUseGrid = !commonStore.useGrid;
  SettingsService.SetUseGrid(newUseGrid).then(() => {
    commonStore.useGrid = newUseGrid;
    // Update viewMode to match
    if (newUseGrid) {
      commonStore.setGridView();
    } else {
      commonStore.setCompactView();
    }
    notificationStore.addNotification("Default View Updated", `Default view set to ${newUseGrid ? 'Grid' : 'List'}`, "success");
  }).catch((error) => {
    console.log(error);
    notificationStore.addNotification("Error", "Failed to update default view", "error");
  });
};

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const displayAppInfo = () => {
  modals.setModalVisibility('appInfoModal', true);
};

const launchDirConfigModal = () => {
  // modals.setModalVisibility('dirOnboardModal', true);
  modals.setModalVisibility('directoryConfigModal', true);
};

const clearRecents = () => {
  SettingsService.ClearRecentProject().then(() => {
    projectStore.recentProjects = []
    notificationStore.addNotification("Recent Projects Cleared", "Recent Projects Cleared", "success")
  })
};

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
  display: flex;
  flex-direction: column;
  gap: 5px;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.settings-component-container {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  height: 100%;
  overflow-y: auto;
  overflow-x: hidden;
  width: 96%;
  gap: 1.5rem;
  color: white;
  padding: 1rem;
  border-radius: var(--large-radius);
}

.settings-component-container::-webkit-scrollbar {
  width: 6px;
}

.settings-component-container::-webkit-scrollbar-thumb {
  background-color: var(--midnight-steel);
  border-radius: 3px;
}

.settings-component-container::-webkit-scrollbar-track {
  background-color: var(--light-steel);
  border-radius: 3px;
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
  /* border-radius: 8px; */
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
  justify-content: center;
  overflow: hidden;
  height: 100%;
  width: max-content;
}

.fixed-width {
  min-width: 200px;
}
</style>

