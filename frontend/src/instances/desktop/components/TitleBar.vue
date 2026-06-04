<template>
  <div style="--wails-draggable:drag" @click="handleClickOutside" @dblclick="toggleMaximize" class="titlebar"
    :class="{ 'title-only': titleOnly, 'titlebar-unsynced': showUnsyncedBar, 'titlebar-inactive': studioInactive || locationsStale }"
    v-stop-propagation>

    <div v-if="!titleOnly" class="titlebar-left" :class="{ 'titlebar-left-inactive': modalsActive }">

      <ClusttaLogo v-if="os !== 'darwin'" :boldText="true" :showText="false" :colored="true" size="small" @click="displayAppInfo()" v-stop-propagation v-tooltip="$t('components.titleBar.aboutClustta')" :class="{ 'is-disabled': progressRunning }" />

      <div ref="studioTabsParent" class="studio-tabs-parent" v-if="userStore.user && projectStore.selectedStudio && !accountStore.isOfflineMode && stage.selectedStage !== 'settings'" 
      :class="{ 'is-disabled': progressRunning, 'mac-os': !isMacFullscreen && os === 'darwin' }">
        <div class="studio-tabs-container" @click="!accountStore.isStudioAuth && toggleStudioList()" v-stop-propagation>
          <span class="studio-tabs">
            <div class="studio-name-with-status">
              <span class="online-indicator" :class="studioStore.appOnline ? 'online' : 'offline'" v-tooltip="studioStore.appOnline ? $t('components.titleBar.connected') : $t('components.titleBar.offline')"></span>
              {{ utils.capitalizeStr(projectStore.getSelectedStudioName) }}
            </div>
            <img v-if="!accountStore.isStudioAuth" class="small-icons chevron" :src="getAppIcon('chevron-down')">

            <div v-if="displayStudioList" class="studio-list-container" :style="{ left: parentLocation?.left + 'px', top: parentLocation?.top + parentLocation?.height + 'px' }">
              <div class="studio-instance-container">

                <div v-for="(studio, index) in studioList" :key="index" class="studio-instance" @click="selectStudio(studio)">
                  <div class="studio-instance-meta">
                    <img class="large-icons" :src="studio.name === 'Personal' ? getAppIcon('two-drives') : getAppIcon('website')">
                    <div>{{ studio.name }}</div>
                  </div>
                </div>
              </div>

            </div>
          </span>
        </div>

          <ActionButton v-if="studioStore.isStudioAdmin && projectStore.selectedStudio?.name !== 'Personal'" :icon="getAppIcon('stall-cog')" v-tooltip="$t('components.titleBar.studioSettings')" :buttonFunction="studioSettings" />
          <ActionButton :icon="getAppIcon('refresh')" v-tooltip="$t('components.titleBar.reloadStudio')" :buttonFunction="reloadStudio" />
      </div>

    </div>

    <div v-if="projectStore.getActiveProject && !restrictedTitles.includes(stage.selectedStage)" style="--wails-draggable:drag"
      class="project-name-container">
      <div class="project-name-text">
        {{ projectStore.getActiveProjectName }}
      </div>
      <div v-if="progressRunning" class="operation-message-text">
        {{ operationMessage }}
      </div>
    </div>

    <div v-else-if="!isAuthPage" style="--wails-draggable:drag"
      class="project-name-container">
      <div class="project-name-text">
        Clustta
      </div>
    </div>

    <div v-if="os === 'darwin'" class="titlebar-buttons">
      <div v-if="locationsStale" class="location-stale-pill" @click="fixStaleLocations" v-stop-propagation v-tooltip="$t('components.titleBar.locationStale')">
        <img class="small-icons" :src="getAppIcon('alert')">
        <span>{{ $t('components.titleBar.locationStalePill') }}</span>
      </div>
      <PlanInfo />
      <ClusttaLogo  :showText="false" :colored="true" size="small" @click="displayAppInfo()" v-stop-propagation v-tooltip="$t('components.titleBar.aboutClustta')" :class="{ 'is-disabled': progressRunning }" />
    </div>

    <!-- Web mode auth buttons (only when not logged in) -->
    <div v-else-if="platformStore.isWeb && !userStore.isUserAuthenticated" class="titlebar-auth-buttons">
      <ActionButton :icon="getAppIcon('launch')" :label="$t('components.titleBar.signUp')" color="var(--grape)" forceIconColor="light" :buttonFunction="goToSignUp" v-tooltip="isWideScreen ? '' : $t('components.titleBar.signUp')" />
      <ActionButton :icon="getAppIcon('login')" :label="isWideScreen ? $t('components.titleBar.login') : ''" :useOutline="true" :buttonFunction="goToLogin" v-tooltip="isWideScreen ? '' : $t('components.titleBar.login')" />
    </div>

    <div v-else-if="!platformStore.isWeb" class="titlebar-buttons">
      <div v-if="locationsStale" class="location-stale-pill" @click="fixStaleLocations" v-stop-propagation v-tooltip="$t('components.titleBar.locationStale')">
        <img class="small-icons" :src="getAppIcon('alert')">
        <span>{{ $t('components.titleBar.locationStalePill') }}</span>
      </div>
      <PlanInfo />

      <div class="titlebar-button minimize" @click="minimizeWindow">
        <img class="small-icons" :src="getAppIcon('collapse-window')" alt="Minimize">
      </div>
      <div class="titlebar-button maximize" @click="toggleMaximize">
        <img class="small-icons" :src="isWindowMaximized ? getAppIcon('minimize-window') : getAppIcon('maximize-window')" alt="Maximize">
      </div>
      <div class="titlebar-button close" @click="closeWindow">
        <img class="small-icons" :src="getAppIcon('close')" alt="Close">
      </div>
    </div>



    <div :class="{ 'loader-bar-container-visible': stage.operationActive }" class="loader-bar-container">
      <div class="loaderBar"></div>
    </div>

  </div>

  <div v-if="displayStudioList" class="studio-list-container" :style="{ left: parentLocation?.left + 'px', top: parentLocation?.top + parentLocation?.height + 'px' }">
    <div class="studio-instance-container">

      <div v-for="(studio, index) in studioList" :key="index" class="studio-instance" @click="selectStudio(studio)">
        <div class="studio-instance-meta">
          <div>{{ studio.name }}</div>
        </div>
      </div>

      <div v-if="studioList.length" class="menu-divider"></div>

      <div class="studio-instance" @click="createStudio()" v-stop-propagation >
        <div class="studio-instance-meta">
          <img class="large-icons" :src="getAppIcon('stall')">
          <div>{{ $t('components.titleBar.newStudio') }}</div>
        </div>
      </div>

    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { AppService, SettingsService } from '@/services';
import { Window, Events } from "@wailsio/runtime";
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

import { useUserStore } from '@/stores/users';
import { useIconStore } from '@/stores/icons';
import { useStageStore } from '@/stores/stages';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useThemeStore } from '@/stores/theme';
import { useCollectionStore } from '@/stores/collections';
import { useSettingsStore } from '@/stores/settings';

import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';
import PlanInfo from '@/instances/common/components/PlanInfo.vue';
import { useStudioStore } from '@/stores/studio';
import { usePlatformStore } from '@/stores/platform';
import { useAccountStore } from '@/stores/accounts';
import { useEntitlementStore } from '@/stores/entitlements';
import { refreshEntitlements } from '@/lib/sync';

const stage = useStageStore();
const userStore = useUserStore();
const studioStore = useStudioStore();
const iconStore = useIconStore();
const trayStates = useTrayStates();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const notificationStore = useNotificationStore();
const themeStore = useThemeStore();
const collectionStore = useCollectionStore();
const settingsStore = useSettingsStore();
const platformStore = usePlatformStore();
const accountStore = useAccountStore();
const entitlementStore = useEntitlementStore();
const route = useRoute();
const router = useRouter();

const { t } = useI18n();

const goToLogin = () => {
  router.push('/auth/login');
};

const goToSignUp = () => {
  router.push('/auth/signup');
};

const os = ref('');
const studioTabsParent = ref(null);
const screenWidth = ref(window.innerWidth);

// Responsive breakpoints
const isWideScreen = computed(() => screenWidth.value >= 400);

const updateScreenWidth = () => {
  screenWidth.value = window.innerWidth;
};

const parentLocation = computed(() => {
  if(!studioTabsParent.value) return
  return studioTabsParent.value.getBoundingClientRect()
});

const isAuthPage = computed(() => route.path.startsWith('/auth'));
const restrictedTitles = ref(['projects', 'settings' ])

const isMacFullscreen = ref(false);

Events.On('fullscreen-enter', async () => {
	isMacFullscreen.value = true;
  console.log('Fullscreen mode entered');
});

Events.On('fullscreen-exit', async () => {
	isMacFullscreen.value = false;
});

const toggleMacFullscreen = () => {
  isMacFullscreen.value = !isMacFullscreen.value;
};

const props = defineProps({
  titleOnly: { type: Boolean, default: false }
});

const displayStudioList = ref(false);
const progressRunning = computed(() => { return stage.operationActive || notificationStore.getProgress.running })

const projectStages = ['browser', 'trash', 'projectSettings'];
const showUnsyncedBar = computed(() => { return projectStore.getActiveProject?.has_remote && projectStore.getActiveProject?.is_unsynced && projectStages.includes(stage.activeStage) });

const studioInactive = computed(() => !entitlementStore.isStudioActive);

const locationsStale = computed(() => settingsStore.locationsStale);

// Navigates to Settings > Directories so the user can re-select stale folders.
const fixStaleLocations = () => {
  settingsStore.pendingTab = 'directories';
  stage.setStageVisibility('settings', true);
};

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};



const studioList = computed(() => { return projectStore.studios.filter(item => item.id !== projectStore.selectedStudio.id && (item.url || item.hosting_mode) ) });

const operationMessage = computed(() => {
  return ' - ' + t('components.titleBar.working');
});

const modalsActive = computed(() => {
  return !!modals.activeModal;
});

const toggleStudioList = () => {
  if (accountStore.isStudioAuth) return;
  displayStudioList.value = !displayStudioList.value;
};

const toggleTheme = () => {
  // Cycle: system -> light -> dark -> system
  const order = ['system', 'light', 'dark'];
  const next = order[(order.indexOf(themeStore.mode) + 1) % order.length];
  themeStore.setMode(next);
};

const displayAppInfo = () => {
  modals.setModalVisibility('appInfoModal', true);
};

const studioSettings = () => {
  stage.setStageVisibility('studioSettings', true)
};

const reloadStudio = async () => {
  stage.operationActive = true;
  let oldSelectedStudio = projectStore.selectedStudio
  await projectStore.loadStudios().then(() => {
    stage.operationActive = false;
  }).catch((error) => {
    console.error('Error:', error);
    stage.operationActive = false;
  });
  let studio = projectStore.studios.find((item) => item.name === oldSelectedStudio.name)
  if (studio) {
    projectStore.selectedStudio = studio
  } else {
    projectStore.selectedStudio = projectStore.studios[0]
  }
  await projectStore.loadProjects();
  refreshEntitlements();
  displayStudioList.value = false;
}

const selectStudio = async (studio) => {

  if(stage.activeStage !== 'studioSettings'){
    stage.setStageVisibility('projects', true);
  }

  displayStudioList.value = false;
  projectStore.activeProject = null
  projectStore.projects = []
  projectStore.untrackedFiles = []
  projectStore.untrackedFolders = []
  studioStore.studioUsers = []

  await projectStore.selectStudio(studio);

  if(projectStore.selectedStudio?.name !== 'Personal'){
    await studioStore.getStudioUsers();
  }

  refreshEntitlements();
  
  if (projectStore.projects.length && projectStore.activeProject && projectStore.activeProject.is_downloaded) {
    await trayStates.refreshData();
  }

}

const createStudio = () => {
  displayStudioList.value = false;

  const inactive = projectStore.studios.find((s) => s.active === false && s.hosting_mode === 'cloud');
  if (inactive) {
    notificationStore.addNotification(
      'Finish setting up your studio',
      `Complete checkout for "${inactive.name}" or delete it before creating another studio.`,
      'error',
      false,
    );
    return;
  }

  modals.setModalVisibility('configClusttaCloudStudioModal', true);
};

const handleClickOutside = (event) => {
  if (displayStudioList.value) {
    displayStudioList.value = false;
  };
};

const isWindowMaximized = ref(false);

const toggleFullscreen = () => {
  isWindowMaximized.value = !isWindowMaximized.value;
};

const isFullscreen = async () => {
  try {
    const isFullscreen = await Window.IsMaximised();
    isWindowMaximized.value = !isFullscreen;
  } catch (error) {
    // console.error('Error checking fullscreen status:', error);
  }
};


function closeWindow() {
  if (settingsStore.minimizeOnClose) {
    AppService.Hide()
  } else {
    AppService.Quit()
  }
}

function minimizeWindow() {
  AppService.Minimize()
}

function toggleMaximize() {
  isFullscreen();
  Window.ToggleMaximise()
}

const frontendReady = async () => {
  Events.Emit("frontend-ready", true);
};

onMounted( async () => {
  os.value = await AppService.GetOS();
  frontendReady();
  document.addEventListener('click', handleClickOutside);
  window.addEventListener('resize', updateScreenWidth);

	emitter.on('window-fullscreen', toggleFullscreen);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside);
  window.removeEventListener('resize', updateScreenWidth);
	emitter.off('window-unfullscreen', toggleFullscreen);

});


</script>

<style scoped>
@import "@/assets/desktop.css";

.loaderBar {
  position: absolute;
  top: 0;
  right: 100%;
  bottom: 0;
  left: 0;
  background: linear-gradient(to right, transparent, var(--accent), transparent);
  width: 0;
  animation: borealisBar 1.2s linear infinite;
  z-index: 1;
}

@keyframes borealisBar {
  0% {
    left: 0%;
    right: 100%;
    width: 0%;
  }

  25% {
    left: 0%;
    right: 75%;
    width: 40%;
  }

  50% {
    left: 0%;
    right: 50%;
    width: 90%;
  }

  75% {
    right: 0%;
    left: 75%;
    width: 40%;
  }

  100% {
    left: 100%;
    right: 0%;
    width: 0%;
  }
}

.loader-bar-container {
  position: absolute;
  bottom: 0;
  left: 0;
  left: 0;
  width: 100%;
  height: 2px;
  overflow: hidden;
  cursor: not-allowed !important;
  box-sizing: border-box;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.loader-bar-container-visible {
  opacity: 1;
}

.project-name-container {
  box-sizing: border-box;
  display: flex;
  overflow: hidden;
  width: max-content;
  height: max-content;
  max-width: 40%;
  width: max-content;
  padding: .3rem .8rem;
  align-items: center;
  position: absolute;
  right: 50%;
  transform: translateX(50%);
  border-radius: var(--small-radius);
  gap: 5px;
  box-sizing: border-box;
}

.project-name-text {
  box-sizing: border-box;
  overflow: hidden;
  align-items: center;
  text-overflow: ellipsis;
  font-weight: 400;
}

.operation-message-text {
  box-sizing: border-box;
  font-style: italic;
  color: var(--alert);
  overflow: hidden;
  align-items: center;
  text-overflow: ellipsis;
  font-weight: 300;
  font-size: small;
  display: flex;
  padding: .1rem .3rem;
}

.divider {
  width: 96%;
  border-bottom: 1px solid rgba(255, 255, 255, 0.096);
}

.menu-divider{
	height: 5px;
	margin-top: 5px;
	margin-bottom: 5px;
}

.studio-list-container {
  position: absolute;
  z-index: 10000;
  /* top: 206px; */
  min-width: 160px;
  width: max-content;
  height: max-content;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  gap: 1rem;
  border-radius: var(--large-radius);
  color: var(--text);

  overflow: hidden;
  box-sizing: border-box;
  max-height: 500px;
  overflow-y: scroll;

  /* background-color: hotpink; */
  
  border-radius: var(--very-large-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  backdrop-filter: blur(35px);
}


.studio-list-container::-webkit-scrollbar {
  width: 4px;
}

.studio-list-container::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--border-strong);
}

.studio-list-container::-webkit-scrollbar-track {
  margin: 20px;
  border-radius: 10px;
}

.chevron {
  pointer-events: none;
  transition: all .1s ease-in;
}

.studio-tabs-container {
  width: max-content;
  height: 80%;
  gap: .5rem;
  display: flex;
  /* padding: .3rem; */
  box-sizing: border-box;
  overflow: hidden;
  flex-direction: column;
  border-radius: var(--small-radius);
  background-color: hsla(0, 0%, 0%, 0.2);
  border-radius: var(--normal-radius);
}

.studio-tabs-container:hover {
  outline: var(--transparent-line);
  background-color: hsla(0, 0%, 0%, 0.15);
}

[data-theme="dark"] .studio-tabs-container {
  background-color: hsla(0, 0%, 0%, 0.8);
}

[data-theme="dark"] .studio-tabs-container:hover {
  background-color: hsla(0, 0%, 0%, 0.2);
}

.studio-tabs-parent {
  display: flex;
  box-sizing: border-box;
  width: max-content;
  height: 100%;
  gap: .5rem;
  align-items: center;
  /* background-color: crimson; */
  position: relative;
  transition: all 0.1s ease;
}

.studio-tabs {
  width: max-content;
  height: 100%;
  display: flex;
  box-sizing: border-box;
  overflow: hidden;
  align-items: center;
  gap: .8rem;
  padding: .3rem .8rem;
}

.studio-instance-container {
  width: 100%;
  height: 100%;
  height: max-content;
  /* gap: .2rem; */
  display: flex;
  padding: .3rem;
  box-sizing: border-box;
  overflow: hidden;
  flex-direction: column;
}

.studio-instance {
  overflow: hidden;
  background-color: transparent;
  text-align: center;
  font-size: 14px;
  line-height: 14px;
  background-color: transparent;
  color: var(--text);
  position: relative;
  border-radius: var(--large-radius);
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  gap: 5px;
  align-items: center;
  padding: .1rem;
  height: 20px;
  width: max-content;
  min-width: max-content;
  min-height: 35px;
  transition: all 0..1s ease;
  justify-content: space-between;
  width: 100%;
}

.studio-instance:hover {
  background-color: var(--hover);
}

.studio-instance:active {
  background-color: rgb(70, 70, 70);
  background-color: #00000013;
}

.studio-instance-pressed {
  box-sizing: border-box;
  background-color: rgba(0, 0, 0, 0.216);
  outline: solid 1px var(--border-strong);
  outline-offset: -1px;
}

.studio-instance-actions {
  display: flex;
  box-sizing: border-box;
  width: max-content;
}

.studio-instance-meta {
  display: flex;
  align-items: center;
  box-sizing: border-box;
  width: 100%;
  height: 100px;
  height: 40px;
  padding: .2rem .5rem;
  gap: 10px;
}

.titlebar-left {
  display: flex;
  box-sizing: border-box;
  width: max-content;
  height: 100%;
  align-items: center;
}

.titlebar-left-inactive {
  opacity: .5;
  pointer-events: none;
}

.titlebar {
  position: relative;
  display: flex;
  box-sizing: border-box;
  width: 100%;
  justify-content: space-between;
  align-items: center;
  min-height: 36px;
  color: var(--text);
  overflow: hidden;
  border-bottom: var(--transparent-line);
  background-color: var(--surface-1);
  z-index: 999999999;
  padding-left: .2rem;
  transition: background-color 0.3s ease;
}

.titlebar-unsynced {
  background-color: #d99a22;
}

[data-theme="dark"] .titlebar-unsynced {
  background-color: hsl(49, 74%, 35%);
}

.titlebar-inactive,
[data-theme="dark"] .titlebar-inactive {
  background-color: var(--danger);
}

.location-stale-pill {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px 3px 5px;
  border-radius: var(--large-radius);
  cursor: pointer;
  font-size: 11px;
  color: white;
  background-color: var(--danger);
  transition: background-color 0.15s ease;
  white-space: nowrap;
  height: 22px;
  margin-right: 6px;
}

.location-stale-pill:hover {
  background-color: hsl(0, 70%, 45%);
}

.title-only {
  border-bottom: 0px;
  justify-content: flex-end;
}

.mac-os{
  margin-left: 76px;
}

.full-screen {
  padding-left: 0px;
}

.titlebar-icon {
  display: flex;
  width: 50px;
  height: 100%;
  overflow: hidden;
  justify-content: flex-start;
  align-items: center;
  justify-content: center;
}

.titlebar-icon img {
  height: 24px;
  width: 24px;
}

.titlebar-title {
  font-size: 14px;
  font-weight: bold;
  justify-content: flex-start;
  flex: auto;
}

.titlebar-buttons {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  -webkit-app-region: no-drag;
  height: 100%;
}

.titlebar-right {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  -webkit-app-region: no-drag;
  height: 100%;
}

.version-info {
  gap: .5rem;
  font-size: 12px;
  width: 100%;
  display: flex;
  font-weight: 100;
  font-weight: 300;
  padding: .2rem;
  padding: .3rem .5rem;
  align-items: center;
  border-radius: var(--small-radius);
}

.oudated:hover{
  background-color: var(--surface-2);
}

.titlebar-button {
  cursor: pointer;
  height: 100%;
  aspect-ratio: 1/1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.titlebar-button.close img {
  width: 18px;
  height: 18px;
}

.titlebar-button.close:hover {
  background-color: crimson;
}

.titlebar-button.minimize img {
  width: 18px;
  height: 18px;
}

.titlebar-button.minimize:hover {
  background-color: #6d6d6d;
}

.titlebar-button.maximize img {
  width: 18px;
  height: 18px;
}

.titlebar-button.maximize:hover {
  background-color: #6d6d6d;
}

.outdated-icon-button {
  cursor: pointer;
  height: 100%;
  aspect-ratio: 1/1;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
}

/* Web mode auth buttons */
.titlebar-auth-buttons {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding-right: 1rem;
  height: 100%;
}

.studio-name-with-status {
  display: flex;
  align-items: center;
  gap: 6px;
}

.online-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  transition: background-color 0.3s ease;
}

.online-indicator.online {
  background-color: #22c55e;
}

.online-indicator.offline {
  background-color: #f59e0b;
}
</style>


