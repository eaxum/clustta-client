<template>
    <div class="info-bar-wrapper" :class="{ 'info-bar-maximized': maximized }" v-stop-propagation>
        <div v-if="activity.expanded" class="expandable-panel-container">
            <ActivityPanel :maximized="maximized" @toggle-maximize="maximized = !maximized" @restore="maximized = false" />
        </div>
        <div v-if="debugModeEnabled" class="expandable-panel-container">
            <DebugConsole :maximized="maximized" @toggle-maximize="maximized = !maximized" @close="toggleDebugConsole" />
        </div>

        <div class="info-bar-root" :style="{ backgroundColor : bgColor }">

        <button class="activity-toggle" :class="{ 'mini-progress': showActivityProgress }"
          :aria-expanded="activity.expanded" :aria-label="$t('activity.title')"
          v-tooltip="showActivityProgress ? currentActivity.title : $t('activity.title')" @click="toggleActivity">
          <img :src="getAppIcon('activity')" class="small-icons" alt="" />
          <template v-if="showActivityProgress">
            <span class="mini-progress-count">[{{ activity.batchIDs.indexOf(currentActivity.operation_id) + 1 }}/{{ activity.batchIDs.length }}]</span>
            <span class="mini-progress-text">{{ activityDisplayName(currentActivity) }}</span>
            <span class="mini-progress-bar" role="progressbar" :aria-label="currentActivity.title"
              :aria-valuenow="currentActivity.total > 0 ? currentActivity.percentage : undefined" aria-valuemin="0" aria-valuemax="100">
              <span class="mini-progress-fill" :style="{ width: currentActivity.percentage + '%' }"></span>
            </span>
          </template>
          <span v-else-if="activity.operations.length">{{ activity.operations.length }}</span>
        </button>

        <div v-if="currentPrompt" ref="promptItem" :class="['prompt-message', currentPrompt.type]">
            <span class="text-container" >{{ currentPrompt.message }}</span>
        </div>

        <div v-if="!currentActivity && progressRunning && progressMinimized"
             @click="restoreProgress"
             class="mini-progress"
             :class="{ 'write-operation': isWriteOperation }"
             v-tooltip="progressTooltip">
          <div class="mini-progress-content">
            <span class="mini-progress-count">[{{ progressCurrent }}/{{ progressTotal }}]</span>
            <span class="mini-progress-text">{{ sentenceCase(progressTitle) }} - {{ progressPercentage }}%</span>
          </div>
          <div class="mini-progress-bar">
            <div class="mini-progress-fill" :style="{ width: progressPercentage + '%' }"></div>
          </div>
        </div>

        <div class="spacer"></div>


        <div v-if="notification" ref="notificationItem" :class="['message', notification.type]" @mouseenter="stopTimer()"
            @mouseleave="showMessage(notification)">
            <img :src="getAppIcon(notificationIcon)" class="notification-icon" alt="">
            <span class="text-container" >{{ utils.capitalizeStr(notification.message) }}</span>
        </div>



        <ActionButton :icon="getAppIcon(bridgeEnabled ? 'brick-cancel' : 'brick')" v-tooltip="bridgeEnabled ? $t('components.infoBar.clickToStopBridge') : $t('components.infoBar.clickToStartBridge')" :buttonFunction="toggleBridge" />

        <ActionButton :icon="getAppIcon('console')" v-tooltip="debugModeEnabled ? $t('components.infoBar.closeConsole') : $t('components.infoBar.openConsole')" :buttonFunction="toggleDebugConsole" />

        <div class="version-info" :class="{ 'outdated': isOutdated, 'outdated-required': isUpdateRequired }" v-tooltip="isOutdated ? $t('components.infoBar.clickToUpdateTo', { version: latestVersion }) : ''" @click="isOutdated && handleUpdateClick()">
            <div>{{ isOutdated ? $t('components.infoBar.updateAvailable') : clusttaVersion }}</div>
        </div>

        </div>
    </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Events } from "@wailsio/runtime";
import { activityDisplayName, sentenceCase } from '@/lib/activity';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';

const { t } = useI18n();

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import { useActivityStore } from '@/stores/activity';
import ActivityPanel from '@/instances/desktop/components/ActivityPanel.vue';
import DebugConsole from '@/instances/desktop/components/DebugConsole.vue';

// stores
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { usePlatformStore } from '@/stores/platform';
import { useSettingsStore } from '@/stores/settings';
import { useUpdateStore } from '@/stores/updates';

const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const platformStore = usePlatformStore();
const settingsStore = useSettingsStore();
const updateStore = useUpdateStore();
const activity = useActivityStore();
let stopActivityEvents;

// props
const props = defineProps({
    bgColor: { type: String, default: '' },
});

// refs
const clusttaVersion = ref('');
const currentPrompt = ref(null);
const debugModeEnabled = ref(false);
const maximized = ref(false);
const emit = defineEmits(['maximize-changed']);
watch(maximized, (value) => emit('maximize-changed', value));
watch(() => activity.expanded || debugModeEnabled.value, (visible) => { if (!visible) maximized.value = false });
watch(() => activity.expanded, (expanded) => { if (expanded) debugModeEnabled.value = false });
const currentActivity = computed(() => activity.running[0]);
const showActivityProgress = computed(() => !!currentActivity.value && !activity.expanded);
const notification = ref(false);
const notificationItem = ref(null);
const timer = ref(null);

// Restricted messages that should not trigger notifications
const restrictedMessages = [
  'no active account set',
  'no active account',
];

// computed properties
const bridgeEnabled = computed(() => settingsStore.bridgeEnabled);

const isOutdated = computed(() => updateStore.isUpdateAvailable);

const isUpdateRequired = computed(() => updateStore.isUpdateRequired);

const latestVersion = computed(() => updateStore.latestVersion);

const notificationIcon = computed(() => {
  const icons = { error: 'close-circle', warning: 'alert', success: 'check-circle', info: 'info' };
  return icons[notification.value?.type] || 'info';
});

const isWriteOperation = computed(() => notificationStore.progress.operationType === 'write');

const progressCurrent = computed(() => notificationStore.progress.current || 0);

const progressMinimized = computed(() => notificationStore.progress.isMinimized);

const progressPercentage = computed(() => Math.round(notificationStore.progress.percentage) || 0);

const progressRunning = computed(() => notificationStore.progress.running);

const progressTitle = computed(() => notificationStore.progress.title || '');

const progressTooltip = computed(() => {
  return t('components.infoBar.clickToRestore', { title: progressTitle.value });
});

const progressTotal = computed(() => notificationStore.progress.total || 0);

// event handlers
const handleAddMessage = (payload) => {
  let notificationData;
  if (typeof payload === 'string' || payload instanceof String) {
    notificationData = JSON.parse(payload);
  } else {
    notificationData = payload;
  }
  showMessage(notificationData);
};

const handleAddPrompt = (payload) => {
  let promptData;
  if (typeof payload === 'string' || payload instanceof String) {
    promptData = JSON.parse(payload);
  } else {
    promptData = payload;
  }
  showPrompt(promptData);
};

const handleClearPrompt = () => { currentPrompt.value = null };

// Register event listeners based on platform
if (platformStore.isWeb) {
  emitter.on('add_message', handleAddMessage);
  emitter.on('add_prompt', handleAddPrompt);
  emitter.on('clear_prompt', handleClearPrompt);
} else {
  Events.On("add_message", async (message) => {
    handleAddMessage(message.data);
  });

  Events.On("add_prompt", async (prompt) => {
    handleAddPrompt(prompt.data);
  });

  Events.On("clear_prompt", async () => {
    handleClearPrompt();
  });
}

// methods

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Opens the appropriate update destination for the current channel.
const handleUpdateClick = () => { updateStore.handleUpdateClick() };

// Restores the progress indicator from minimized state.
const restoreProgress = () => { notificationStore.restoreProgress() };

// Displays a notification message with auto-dismiss timer.
const showMessage = async (data) => {
  const messageText = data.message?.toLowerCase() || '';
  const isRestricted = restrictedMessages.some(restricted =>
    messageText.includes(restricted.toLowerCase())
  );

  if (isRestricted) {
    return;
  }

  notification.value = data;
  clearTimeout(timer.value);
  timer.value = setTimeout(() => {
    notification.value = null;
  }, 6000);
};

// Displays a prompt message.
const showPrompt = async (data) => { currentPrompt.value = data };

// Stops the notification auto-dismiss timer.
const stopTimer = () => { clearTimeout(timer.value) };

// Toggles the bridge HTTP server on or off.
const toggleBridge = async () => {
  try {
    await settingsStore.toggleBridge();
  } catch (error) {
    console.log(error);
  }
};

// Toggles the Activity panel.
const toggleActivity = () => {
  if (activity.expanded) {
    activity.expanded = false;
    return;
  }
  debugModeEnabled.value = false;
  notificationStore.minimizeProgress();
  activity.open();
};

const toggleDebugConsole = () => {
  activity.expanded = false;
  debugModeEnabled.value = !debugModeEnabled.value;
};

// lifecycle hooks
onMounted(async () => {
  if (!platformStore.isWeb) stopActivityEvents = await activity.initialize();
  await updateStore.initialize();
  clusttaVersion.value = updateStore.currentVersion || await utils.getRawClusttaVersion();
  updateStore.startAutoCheck();
  await settingsStore.initializeBridge();
});

onBeforeUnmount(() => {
  stopActivityEvents?.();
  updateStore.stopAutoCheck();
});


</script>

<style scoped>
@import "@/assets/desktop.css";

.activity-toggle {
  display: flex;
  align-items: center;
  gap: .5rem;
  background: transparent;
  color: var(--text);
  border: 0;
  cursor: pointer;
  font-size: 12px;
}

.activity-toggle .small-icons, .activity-toggle .mini-progress-count {
  flex-shrink: 0;
}
.activity-toggle .mini-progress-text {
  font-weight: 600;
  max-width: 240px;
}
.activity-toggle .mini-progress-bar {
  width: 100px;
}
.activity-toggle .mini-progress-fill {
  display: block;
}

.info-bar-wrapper {
  display: flex;
  flex-direction: column;
  width: 100%;
  background-color: var(--surface-3);
  box-sizing: border-box;
  /* z-index: 9; */
}

.expandable-panel-container {
  padding: .4rem ;
  padding-bottom: 0;
}

.info-bar-maximized {
  flex: 1;
  min-height: 0;
  height: 100%;
  border-radius: 24px 24px 0px 0px;
}
.info-bar-maximized .expandable-panel-container {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}
.info-bar-root {
  flex-shrink: 0;
}

.info-bar-root{
    width: 100%;
    height: 100%;
    height: 30px;
    display: flex;
    overflow: hidden;
    box-sizing: border-box;
    align-items: center;
    justify-content: space-between;
    color: var(--text);
    padding: 0 .8rem;
    font-size: 13px;
    font-weight: 300;
  }

.version-info {
  gap: .5rem;
  width: 100%;
  width: max-content;
  min-width: max-content;
  display: flex;
  padding: .2rem;
  padding: .3rem .5rem;
  align-items: center;
  border-radius: var(--small-radius);
  transition: background-color 0.2s ease, border-radius 0.2s ease;
  height: 70%;
  box-sizing: border-box;
}

.version-info.outdated {
  cursor: pointer;
  border-radius: 6px;
  padding: .3rem .3rem;
  background-color: var(--warning);
  color: var(--text-inverse);
  font-weight: 500;
}

.version-info.outdated-required {
  background-color: var(--alert);
}

.version-info.outdated:hover {
  background-color: var(--alert);
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

.message {
  width: 100%;
  width: min-content;
  display: flex;
  align-items: center;
  white-space: nowrap;
  overflow: hidden;
  display: flex;
  align-items: center;
  gap: .3rem;
  border-radius: 6px;
  box-sizing: border-box;
  height: 70%;
  justify-content: flex-end;
  background-color: crimson;
  padding: .3rem .3rem;
  z-index: 99999;
}

.notification-icon {
  width: 18px;
  height: 18px;
  min-width: 18px;
  /* filter: invert(100%); */
}

.text-container{
  color: white;
    text-overflow: ellipsis;
    overflow: hidden;
}

.error {
  outline: solid 1px #FF3333;
  background-color:  #FF3333;
}

.warning {
  outline: solid 1px #F5A623;
  background-color: #F5A623;
}

.message.info {
  background-color: #2C75E2;
}

.success {
  background-color:  #20A41C;
}

.prompt-message {
  display: flex;
  align-items: center;
  white-space: nowrap;
  overflow: hidden;
  gap: .3rem;
  border-radius: 6px;
  box-sizing: border-box;
  height: 70%;
  padding: .3rem .5rem;
  margin-right: .5rem;
  background-color: #2C75E2;
  z-index: 99999;
}

.prompt-message.error {
  background-color: #FF3333;
}

.spacer {
  flex: 1;
}

/* Mini Progress Indicator */
.mini-progress {
  display: flex;
  /* flex-direction: column; */
  gap: .5rem;
  padding: .2rem .6rem;
  border-radius: 4px;
  /* background-color: rgba(44, 117, 226, 0.15); */
  cursor: pointer;
  transition: all 0.2s ease;
  min-width: 200px;
  /* max-width: 300px; */
  align-items: center;
  /* height: 10px; */
  height: min-content;
  overflow: hidden;
}

.mini-progress:hover {
  background-color: rgba(44, 117, 226, 0.25);
  background-color: var(--surface-4);
}

.mini-progress.write-operation:hover {
  background-color: rgba(238, 92, 8, 0.25);
}

.mini-progress-content {
  display: flex;
  align-items: center;
  gap: .5rem;
  color: var(--text);
}

.mini-progress-icon {
  width: 14px;
  height: 14px;
  filter: invert(100%);
}

.mini-progress-text {
  font-weight: 400;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.mini-progress-count {
  font-weight: 300;
  opacity: 0.8;
}

.mini-progress-bar {
  width: 100%;
  min-width: 100px;
  height: 4px;
  background-color: rgba(255, 255, 255, 0.2);
  border-radius: 999px;
  overflow: hidden;
  background-color: var(--surface-4);
}

.mini-progress-fill {
  height: 100%;
  background-color: rgb(44, 117, 226);
  background-color: rgb(67, 210, 67);
  border-radius: 999px;
  transition: width 0.3s ease;
}

.write-operation .mini-progress-fill {
  background-color: rgb(238, 92, 8);
}

.prompt-message {
  width: max-content;
  display: flex;
  align-items: center;
  white-space: nowrap;
  overflow: hidden;
  gap: 1rem;
  border-radius: 6px;
  box-sizing: border-box;
  height: 70%;
  justify-content: flex-start;
  /* background-color: var(--surface-2); */
  padding: .3rem .5rem;
  z-index: 99999;
}

.prompt-message.info {
  background-color: transparent;
  /* background-color: #4A90E2;
  outline: solid 1px #4A90E2; */
}

.prompt-message.warning {
  background-color: #F5A623;
  outline: solid 1px #F5A623;
}

.prompt-message.error {
  background-color: #FF3333;
  outline: solid 1px #FF3333;
}

.prompt-message.success {
  background-color: #20A41C;
  outline: solid 1px #20A41C;
}

</style>

