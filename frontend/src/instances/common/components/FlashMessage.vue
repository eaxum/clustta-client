<template>
  <div v-if="progressRunning && 
             !stageStore.operationActive && 
             !notificationStore.progress.isMinimized" 
       v-stop-propagation 
       class="desktop-overlay-mask">
  </div>

  <div v-if="progressRunning && !stageStore.operationActive && !notificationStore.progress.isMinimized" 
       class="flash-area" 
       ref="flashArea"
       :class="{ 'flash-area-desktop': isDesktop, 'flash-area-desktop-progress': progressRunning }">
    
    <div class="progress-bar">

      <div class="progress-bar-header">
        <div class="header-title">
          <img :src="getAppIcon(progressIcon)" class="header-icon small-icons" />
          <span class="header-text">{{ notificationStore.getProgress.title }}</span>
        </div>
        <button @click="minimizeProgress" class="minimize-button single-action-button" v-tooltip="$t('components.flashMessage.minimize')">
          <img :src="getAppIcon('chevron-down')" class="minimize-icon small-icons" />
        </button>
      </div>

      <div class="progress-bar-meta">
        <span class="progress-bar-message">{{ notificationStore.getProgress.message }}</span>
        <div class="progress-bar-total">{{ notificationStore.getProgress.current }}/{{
          notificationStore.getProgress.total }}</div>
      </div>

      <div class="progress-bar-loader tint">
        <ProgressBar :assetProgress="notificationStore.getProgress.percentage" v-stop-propagation />
      </div>

      <div class="pop-up-actions">

        <div v-if="notificationStore.getProgress.extra_message" class="pop-up-info">
          <div class="pop-up-stats" :style="{ color : message.color} "> 
            {{ throttledExtraMessage }}
          </div>
        </div>
        <GeneralButton v-if="notificationStore.canCancel" 
          :label="isAwaitingResponse ? $t('components.flashMessage.cancelling') : $t('components.flashMessage.cancel')" 
          :buttonFunction="cancelOperation" 
          :loading="isAwaitingResponse"
          :colored="false"
          :fullWidth="false" />
      </div>


    </div>

  </div>



</template>

<script setup>
import { watch, ref, computed, onMounted, onBeforeUnmount } from 'vue';
import { useTrayStates } from '@/stores/TrayStates';
import ProgressBar from '@/instances/common/components/ProgressBar.vue';
import { useNotificationStore } from '@/stores/notifications';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import { Events } from "@wailsio/runtime";
import emitter from '@/lib/mitt';
import { useStageStore } from '@/stores/stages';
import { useIconStore } from '@/stores/icons';
import { usePlatformStore } from '@/stores/platform';

const platformStore = usePlatformStore();

const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon;
};

const props = defineProps({
  isDesktop: {
    type: Boolean,
    default: false
  }
});


// refs
const isAwaitingResponse = ref(false);

const trayStates = useTrayStates();
const notificationStore = useNotificationStore();
const stageStore = useStageStore();

const notificationItem = ref(null);
const progressRunning = ref(false);
const files = ref(214);

// timers
const progressDelay = 100;
let progressTimer = null;

const progressIcon = computed(() => {
  const message = notificationStore.getProgress.message?.toLowerCase() || '';
  
  if (message.includes('download') || message.includes('receiving')) {
    return 'download';
  } else if (message.includes('revert')) {
    return 'revert';
  } else if (message.includes('rebuild')) {
    return 'jigsaw';
  } else if (message.includes('upload') || message.includes('sending')) {
    return 'cloud-up';
  } else if (message.includes('sync')) {
    return 'cloud-up';
  } else if (message.includes('checkpoint')) {
    return 'checkpoint-stone';
  } else if (message.includes('trim') || message.includes('compact')) {
    return 'scissors';
  } else if (message.includes('delete') || message.includes('trash')) {
    return 'trash';
  }
  
  return 'download';
});

const throttledExtraMessage = ref(notificationStore.getProgress.extra_message);
let extraMessageTimeout = null;
let lastExtraMessage = notificationStore.getProgress.extra_message;

watch(() => notificationStore.getProgress.running, (isRunning) => {
  if (isRunning) {
    progressTimer = setTimeout(() => {
      progressRunning.value = true;
    }, progressDelay);
  } else {
    clearTimeout(progressTimer);
    progressRunning.value = false;
  }
});

watch(
  () => notificationStore.getProgress.extra_message,
  (newVal) => {
    lastExtraMessage = newVal;
    if (!extraMessageTimeout) {
      throttledExtraMessage.value = newVal;
      extraMessageTimeout = setTimeout(() => {
        throttledExtraMessage.value = lastExtraMessage;
        extraMessageTimeout = null;
      }, 500);
    }
  }
);

const handleAddMessage = (payload) => {
  let notificationData;
  if (typeof payload === 'string' || payload instanceof String) {
    notificationData = JSON.parse(payload);
  } else {
    notificationData = payload;
  }
  showMessage(notificationData);
};

const handleProgressUpdate = (progressData) => {
  notificationStore.updateProgress(progressData);
};

// Register event listeners based on platform
if (platformStore.isWeb) {
  emitter.on('add_message', handleAddMessage);
  emitter.on('progress-update', handleProgressUpdate);
} else {
  Events.On("add_message", async (message) => {
    handleAddMessage(message.data);
  });
  Events.On("progress-update", async (message) => {
    handleProgressUpdate(message.data);
  });
}

const notification = ref(false);

const timer = ref(null);

const message = ref({
  stats: 'Saved 24MB',
  color: 'var(--online)'
})

async function showMessage(data) {
  notification.value = data
  clearTimeout(timer.value);
  timer.value = setTimeout(() => {
    notification.value = null;
  }, 3000);
}
const handleClickOutside = (event) => {
  if (notification.value && (event.target !== notificationItem.value)) {
    notification.value = null;
  }
};

const cancelOperation = async () => {
  isAwaitingResponse.value = true;
  await notificationStore.cancleFunction()
  notificationStore.resetProgress()
  notificationStore.cancleFunction = null
  notificationStore.canCancel = false
  isAwaitingResponse.value = false;
};

const minimizeProgress = () => {
  notificationStore.minimizeProgress();
};

onMounted(() => {
  document.addEventListener('click', handleClickOutside);

});

onBeforeUnmount(() => {
  clearTimeout(progressTimer);
  document.removeEventListener('click', handleClickOutside);
});
</script>

<style scoped>
/* @import "@/assets/tray.css"; */

/* button */
.pop-up-actions {
  box-sizing: border-box;
  /* background-color: red; */
  justify-content: flex-end;
  gap: .5rem;
}

.pop-up-info{
  width: 100%;
  display: flex;
  color: var(--text);
  /* background-color: #20A41C; */
  height: 100%;
  min-height: 100%;
  font-weight: 500;
}

.flash-overlay-mask {
  z-index: 9998;
  cursor: not-allowed;
  cursor: wait;
}

.desktop-overlay-mask {
  position: absolute;
  z-index: 3;
  width: 100%;
  height: 100%;
  display: flex;
  transition: opacity 0.3s ease;
  background-color: rgba(0, 0, 0, 0.5);
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(3px);
  box-sizing: border-box;
  cursor: not-allowed;
  cursor: wait;
}

.flash-area {
  position: fixed;
  bottom: 2%;
  z-index: 9999;
  font-size: 14px;
  width: 96%;

  display: flex;
  align-items: center;
  border-radius: 8px;
  box-shadow: 0px 2px 5px rgba(0, 0, 0, 0.1);
  justify-content: space-between;
  white-space: nowrap;
  overflow: hidden;
  display: flex;
  align-items: center;
  box-sizing: border-box;
  height: max-content;
  justify-content: center;
  flex-direction: column;
  backdrop-filter: blur(10px);
  gap: .5rem;
}

.flash-area-desktop {
  position: fixed;
  width: 50%;
  max-width: 500px;
  min-width: 350px;
}

.flash-area-desktop-progress {
  top: 0px;
  right: 0px;
  right: 50%;
  top: 45%;
  transform: translateX(50%);
}

.progress-bar {
  position: relative;
  /* right: 50%;
  transform: translateX(50%); */
  z-index: 9999;
  font-size: 14px;
  width: 100%;
  display: flex;
  align-items: center;
  padding: 1rem;
  border-radius: 8px;
  /* box-shadow: 0px 2px 5px rgba(0, 0, 0, 0.1); */
  justify-content: space-between;
  white-space: nowrap;
  overflow: hidden;
  display: flex;
  align-items: center;
  gap: 1rem;
  box-sizing: border-box;
  /* height: min-content; */
  /* height: 1rem; */
  background-color: rgba(0, 0, 0, 0.216);
  background-color: rgb(46, 46, 46);
  /* background-color: red; */
  justify-content: space-between;
  flex-direction: column;
  outline: solid 1px rgb(151, 151, 151);
  /* outline: var(--transparent-line); */
  outline-offset: -1px;
  outline-offset: -1px;
  outline: var(--transparent-line);
  background-color: var(--surface-1);
  border-radius: var(--very-large-radius);

}

.progress-bar-meta {
  color: var(--text);
  display: flex;
  align-items: center;
  justify-content: space-between;
  /* background-color: #FF3333; */
  width: 100%;
  gap: .5rem;
  box-sizing: border-box;
  overflow: hidden;
}

.progress-bar-message {
  /* flex: 2; */
  overflow: hidden;
  /* background-color: #FF3333; */
  align-items: center;
  justify-content: flex-start;
  flex: 1;
  text-overflow: ellipsis;
  color: var(--text);
}

.progress-bar-total {
  /* flex: 1; */
  display: flex;
  /* background-color: mediumseagreen; */
  align-items: center;
  justify-content: flex-end;
  overflow: hidden;
  /* width: 200px; */
  width: max-content;
}

.progress-bar-header {
  color: var(--text);
  display: flex;
  align-items: center;
  font-size: 16px;
  justify-content: space-between;
  width: 100%;
  gap: .5rem;
}

.header-title {
  display: flex;
  align-items: center;
  gap: .5rem;
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.header-icon {
  width: 22px;
  height: 22px;
  flex-shrink: 0;
}

.header-text {
  font-family: Inter, sans-serif;
  color: var(--text);
  font-size: 18px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.minimize-button {
  background: transparent;
  border: none;
  cursor: pointer;
  padding: .3rem;
  border-radius: var(--small-radius);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.minimize-button:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.minimize-icon {
  width: 18px;
  height: 18px;
  /* filter: invert(100%); */
}

.progress-bar-loader {
  position: relative;
  width: 100%;
  height: .2rem;
  border-radius: 999px;
  /* background-color: white; */

}

.action-section {
  display: flex;
  gap: .5rem;
  /* background-color: red; */
}

.cancel-button {
  max-width: 20px;
  padding: 0px;
  color: rgb(255, 255, 255);

}

.message {
  position: relative;
  /* right: 50%;
  transform: translateX(50%); */
  z-index: 9999;
  font-size: 14px;
  width: 100%;
  display: flex;
  align-items: center;
  padding: 1rem;
  border-radius: 8px;
  box-shadow: 0px 2px 5px rgba(0, 0, 0, 0.1);
  justify-content: space-between;
  outline: solid 1px var(--border-strong);
  white-space: nowrap;
  overflow: hidden;
  display: flex;
  align-items: center;
  gap: 1rem;
  box-sizing: border-box;
  min-height: 3rem;
  height: max-content;
  background-color: rgba(0, 0, 0, 0.216);
  background-color: rgb(46, 46, 46);
  /* background-color: green; */
  justify-content: space-between;
  opacity: .7;
  outline-offset: -1px;
  color: rgb(255, 255, 255);
}

.error {
  outline: solid 1px #FF3333;
}

.success {
  outline: solid 1px #20A41C;
}
</style>


