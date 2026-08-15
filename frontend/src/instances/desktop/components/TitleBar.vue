<template>
  <div style="--wails-draggable:drag" @dblclick="toggleMaximize" class="titlebar"
    :class="{ 'title-only': titleOnly, 'titlebar-darwin': os === 'darwin', 'titlebar-unsynced': showUnsyncedBar, 'titlebar-inactive': studioInactive || locationsStale }"
    v-stop-propagation>

    <div v-if="!titleOnly" class="titlebar-left" :class="{ 'titlebar-left-inactive': modalsActive }">

      <div v-if="os !== 'darwin'" class="clustta-logo-left" >
        <ClusttaLogo :boldText="true" :showText="false" :colored="true" size="small" @click="displayAppInfo()" v-stop-propagation v-tooltip="$t('components.titleBar.aboutClustta')" :class="{ 'is-disabled': progressRunning }" />
      </div>

      <div class="studio-tabs-parent" v-if="userStore.user && projectStore.selectedStudio && !accountStore.isOfflineMode && stage.selectedStage !== 'settings'"
      :class="{ 'is-disabled': progressRunning, 'mac-os': !isMacFullscreen && os === 'darwin' }">
        <div class="studio-selector" v-stop-propagation>
          <DropDownBox
            :items="studioOptions"
            :selectedItem="projectStore.getSelectedStudioName"
            :onSelect="selectStudioByName"
            :disabled="accountStore.isStudioAuth"
            :fullWidth="false"
            :useFilter="false"
          >
            <template #footer="{ close }">
              <div class="studio-dropdown-divider"></div>
              <button class="studio-create-action" type="button" @click="createStudio(close)">
                <img class="small-icons" :src="getAppIcon('stall')">
                <span>{{ $t('components.titleBar.newStudio') }}</span>
              </button>
            </template>
          </DropDownBox>
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

</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { AppService } from '@/services';
import { Window, Events } from "@wailsio/runtime";
import emitter from '@/lib/mitt';

import { useUserStore } from '@/stores/users';
import { useIconStore } from '@/stores/icons';
import { useStageStore } from '@/stores/stages';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useSettingsStore } from '@/stores/settings';

import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import PlanInfo from '@/instances/common/components/PlanInfo.vue';
import { useStudioStore } from '@/stores/studio';
import { usePlatformStore } from '@/stores/platform';
import { useAccountStore } from '@/stores/accounts';
import { useEntitlementStore } from '@/stores/entitlements';
import { useMenu } from '@/stores/menu';
import { refreshEntitlements } from '@/lib/sync';

const stage = useStageStore();
const userStore = useUserStore();
const studioStore = useStudioStore();
const iconStore = useIconStore();
const trayStates = useTrayStates();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const notificationStore = useNotificationStore();
const settingsStore = useSettingsStore();
const platformStore = usePlatformStore();
const accountStore = useAccountStore();
const entitlementStore = useEntitlementStore();
const menu = useMenu();
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
const screenWidth = ref(window.innerWidth);

// Responsive breakpoints
const isWideScreen = computed(() => screenWidth.value >= 400);

const updateScreenWidth = () => {
  screenWidth.value = window.innerWidth;
};

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



const studioOptions = computed(() => {
  return projectStore.studios
    .filter((studio) => studio.id === projectStore.selectedStudio?.id || studio.url || studio.hosting_mode)
    .map((studio) => {
      if (studio.id !== projectStore.selectedStudio?.id) {
        return {
          ...studio,
          icon: null,
          iconTone: '',
          iconTooltip: '',
        };
      }

      const statusTooltip = studioStore.appOnline
        ? t('components.titleBar.connected')
        : t('components.titleBar.offline');

      return {
        ...studio,
        icon: getAppIcon('dot-big'),
        iconTone: studioStore.appOnline ? 'go' : 'alert',
        iconTooltip: statusTooltip,
      };
    });
});

const operationMessage = computed(() => {
  return ' - ' + t('components.titleBar.working');
});

const modalsActive = computed(() => {
  return !!modals.activeModal;
});

const displayAppInfo = () => {
  modals.setModalVisibility('appInfoModal', true);
};

const studioSettings = () => {
  menu.disableAllMenus();
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
}

const selectStudio = async (studio) => {

  if(stage.activeStage !== 'studioSettings'){
    stage.setStageVisibility('projects', true);
  }

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

const selectStudioByName = async (studioName) => {
  const studio = projectStore.studios.find((item) => item.name === studioName);
  if (!studio || studio.id === projectStore.selectedStudio?.id) return;

  await selectStudio(studio);
};

const createStudio = (closeDropdown) => {
  closeDropdown();

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
  window.addEventListener('resize', updateScreenWidth);

	emitter.on('window-fullscreen', toggleFullscreen);
});

onBeforeUnmount(() => {
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

.studio-tabs-parent {
  display: flex;
  box-sizing: border-box;
  width: max-content;
  height: 100%;
  gap: .5rem;
  align-items: center;
  position: relative;
  transition: all 0.1s ease;
}

.studio-selector {
  width: 200px;
}

.studio-selector :deep(.list-box-container) {
  width: 100%;
}

.studio-dropdown-divider {
  border-top: var(--transparent-line);
  margin: .3rem .1rem;
}

.studio-create-action {
  display: flex;
  align-items: center;
  gap: .5rem;
  width: 100%;
  padding: .3rem .5rem;
  border: 0;
  border-radius: var(--normal-radius);
  color: var(--text);
  background: transparent;
  font: inherit;
  font-size: 14px;
  text-align: left;
  cursor: pointer;
}

.studio-create-action:hover {
  background-color: var(--surface-4);
}

.clustta-logo-left{
  min-width: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
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
  min-height: 44px;
  color: var(--text);
  overflow: hidden;
  /* border-bottom: var(--transparent-line); */
  background-color: var(--surface-1);
  z-index: 999999999;
  /* padding-left: .2rem; */
  transition: background-color 0.3s ease;
}

.titlebar-darwin {
  padding-left: .2rem;
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

</style>


