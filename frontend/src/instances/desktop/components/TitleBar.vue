<template>
  <div style="--wails-draggable:drag" @click="handleClickOutside" @dblclick="toggleMaximize" class="titlebar"
    :class="{ 'title-only': titleOnly }" v-stop-propagation>

    <div v-if="!titleOnly" class="titlebar-left" :class="{ 'titlebar-left-inactive': modalsActive }">

      <div v-if="os !== 'darwin'" class="titlebar-icon" :class="{ 'is-disabled': progressRunning }">
        <img src="/icons/clustta.png" alt="Clustta Icon" @click="displayAppInfo()" v-stop-propagation
          v-tooltip="'About Clustta'">
      </div>

      <div ref="studioTabsParent" class="studio-tabs-parent" v-if="userStore.user && projectStore.selectedStudio" 
      :class="{ 'is-disabled': progressRunning, 'mac-os': !isMacFullscreen && os === 'darwin' }">
        <div class="studio-tabs-container" @click="toggleStudioList()" v-stop-propagation>
          <span class="studio-tabs">
            <img v-if="projectStore.selectedStudio?.name == 'Personal'" class="large-icons"
              :src="getAppIcon('two-drives')">
            <img v-else-if="projectStore.selectedStudio" class="large-icons"
              :src="projectStore.useAltUrl ? getAppIcon('two-drives') : getAppIcon('website')">
            <div>{{ utils.capitalizeStr(projectStore.getSelectedStudioName) }} </div>
            <img v-if="studioList.length" class="small-icons chevron" :src="getAppIcon('chevron-down')">

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
        <ToggleSwitch :online="true" v-if="projectStore.selectedStudio?.alt_url"
          v-tooltip="projectStore.useAltUrl ? 'On Prem' : 'Remote'" :switchValueProp="!projectStore.useAltUrl"
          @click="toggleStudioRoute()" />

          <ActionButton v-if="userStore.userCanCreateProject && projectStore.selectedStudio?.name !== 'Personal'" :icon="getAppIcon('cog')" v-tooltip="'Studio Settings'" :buttonFunction="studioSettings" />
          <ActionButton :icon="getAppIcon('refresh')" v-tooltip="'Reload Studio'" :buttonFunction="reloadStudio" />
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


    <div v-if="os === 'darwin'" class="titlebar-icon" :class="{ 'is-disabled': progressRunning }">
        <img src="/icons/clustta.png" alt="Clustta Icon" @click="displayAppInfo()" v-stop-propagation
          v-tooltip="'About Clustta'">
      </div>

    <div v-else class="titlebar-buttons">
      <!-- <ToggleSwitch :switchValueProp="themeStore.isDarkMode" @click="toggleTheme()" /> -->
      <!-- {{  userStore.user?.id }} -->

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
          <img class="large-icons" :src="studio.name === 'Personal' ? getAppIcon('two-drives') : getAppIcon('website')">
          <div>{{ studio.name }}</div>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watchEffect } from 'vue';
import { AppService, SettingsService } from '@/../bindings/clustta/services/index';
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

import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import { useStudioStore } from '@/stores/studio';

const stage = useStageStore();
const userStore = useUserStore();
const studioStore = useStudioStore();
const iconStore = useIconStore();
const trayStates = useTrayStates();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const notificationStore = useNotificationStore();
const themeStore = useThemeStore();

const os = ref('');
const studioTabsParent = ref(null);

const parentLocation = computed(() => {
  if(!studioTabsParent.value) return
  return studioTabsParent.value.getBoundingClientRect()
});

const restrictedTitles = ref(['projects', 'studioSettings' ])

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

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const userCanCreateProject = () => {
  const user = userStore.user;
  const selectedStudio = projectStore.selectedStudio;

  if (!user || !selectedStudio) {
    userStore.userCanCreateProject = false
    return false
  }

  const activeUserId = user.id;
  const studioName = selectedStudio.name;

  if (studioName === 'Personal') {
    userStore.userCanCreateProject = true;
    return true
  } else {
    if (!userStore.getUserAuthentication) {
      userStore.userCanCreateProject = false
      return false
    } else {
      const userStudioRole = selectedStudio.Users?.find((item) => item.id === activeUserId)?.role_name;
      const isAdmin = userStudioRole === 'admin';
      userStore.userCanCreateProject = isAdmin;
      return userStudioRole === 'admin';
    }
  }
};

watchEffect(() => {
  if (projectStore.selectedStudio) {
    userCanCreateProject()
  }
});

const studioList = computed(() => { return projectStore.studios.filter(item => item.id !== projectStore.selectedStudio.id && item.url ) });

const operationMessage = computed(() => {
  return ' - syncing';
});

const modalsActive = computed(() => {
  return !!modals.activeModal;
});

const toggleStudioList = () => {
  if (!studioList.value.length) return;
  displayStudioList.value = !displayStudioList.value;
};

const toggleStudioRoute = () => {
  SettingsService.SetUseAltUrl(!projectStore.useAltUrl).then(() => {
    projectStore.useAltUrl = !projectStore.useAltUrl;
  })
};

const toggleTheme = () => {
  themeStore.isDarkMode = !themeStore.isDarkMode
  let theme;
  if(themeStore.isDarkMode){
    theme = 'dark'
  } else {
    theme = 'light'
  }
  SettingsService.SetTheme(theme).then(() => {
    themeStore.selectedTheme = theme;
    themeStore.applyTheme();
  })
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

  if (projectStore.projects.length && projectStore.activeProject && projectStore.activeProject.is_downloaded) {
    await trayStates.refreshData();
  }

  userCanCreateProject();
}

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
  AppService.Quit()
  return
}

function minimizeWindow() {
  AppService.Minimize()
}

function toggleMaximize() {
  isFullscreen();
  Window.ToggleMaximise()
}

const frontendReady = async () => {
  let eventData = new Events.WailsEvent("frontend-ready", true);
  Events.Emit(eventData);
};

onMounted( async () => {
  os.value = await AppService.GetOS();
  frontendReady();
  document.addEventListener('click', handleClickOutside);

	emitter.on('window-fullscreen', toggleFullscreen);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside);
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
  background: crimson;
  width: 0;
  animation: borealisBar 2s linear infinite;
  z-index: 1;
  background-color: rgb(255, 229, 201);
  background-color: rgb(31, 163, 60);
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
  height: 3px;
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
  padding: .5rem 1rem;
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
  background-color: royalblue;
  /* left: 50px; */
  border-radius: var(--small-radius);
  background-color: var(--black);
  color: var(--white);
  outline: var(--transparent-line);
  outline-offset: -1px;
  overflow: hidden;
  box-sizing: border-box;
  max-height: 500px;
  overflow-y: scroll;
}


.studio-list-container::-webkit-scrollbar {
  width: 8px;
}

.studio-list-container::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--light-steel);
}

.studio-list-container::-webkit-scrollbar-track {
  margin: 10px;
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
  padding: .3rem;
  box-sizing: border-box;
  overflow: hidden;
  flex-direction: column;
  border-radius: var(--small-radius);
  background-color: var(--steel);
}

.studio-tabs-container:hover {
  outline: var(--transparent-line);
  background-color: var(--black-steel);
}

[data-theme="dark"] .studio-tabs-container {
  background-color: var(--midnight-steel);
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
  padding: .5rem 1rem;
}

.studio-instance-container {
  width: 100%;
  height: 100%;
  height: max-content;
  gap: .2rem;
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
  color: var(--white);
  position: relative;
  border-radius: 8px;
  border-radius: var(--small-radius);
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  gap: 5px;
  align-items: center;
  padding: .1rem;
  height: max-content;
  width: max-content;
  min-width: max-content;
  min-height: max-content;
  transition: all 0..1s ease;
  justify-content: space-between;
  width: 100%;
}

.studio-instance:hover {
  background-color: var(--steel);
}

.studio-instance:active {
  background-color: rgb(70, 70, 70);
  background-color: #00000013;
}

.studio-instance-pressed {
  box-sizing: border-box;
  background-color: rgba(0, 0, 0, 0.216);
  outline: solid 1px var(--white);
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
  padding: .2rem 1rem;
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
  height: 46px;
  color: var(--white);
  overflow: hidden;
  border-bottom: var(--transparent-line);
  background-color: var(--black);
  z-index: 999999999;
  padding-left: .2rem;
  transition: all 0..1s ease;
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
  background-color: var(--dark-steel);
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
</style>