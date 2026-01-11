<template>
    <div class="info-bar-root" :style="{ backgroundColor : bgColor }" >

        <!-- <div v-if="currentPrompt" ref="promptItem" :class="['prompt-message', currentPrompt.type]">
            <span class="text-container" >{{ currentPrompt.message }}</span>
        </div> -->

        <div v-if="progressRunning && progressMinimized" 
             @click="restoreProgress" 
             class="mini-progress"
             :class="{ 'write-operation': isWriteOperation }"
             v-tooltip="progressTooltip">
          <div class="mini-progress-content">
            <span class="mini-progress-count">[{{ progressCurrent }}/{{ progressTotal }}]</span>
            <span class="mini-progress-text">{{ progressTitle }} - {{ progressPercentage }}%</span>
          </div>
          <div class="mini-progress-bar">
            <div class="mini-progress-fill" :style="{ width: progressPercentage + '%' }"></div>
          </div>
        </div>

        <div class="spacer"></div>


        <div v-if="notification" ref="notificationItem" :class="['message', notification.type]" @mouseenter="stopTimer()"
            @mouseleave="showMessage(notification)">
            <span class="text-container" >{{ utils.capitalizeStr(notification.message) }}</span>
        </div>

        <div class="version-info" :class="{ 'oudated' : isOutdated}" v-tooltip=" isOutdated ? 'Click to update' : '' ">
            <div v-if="isOutdated" class="outdated-icon-button">
            <img :src="getAppIcon('info-triangle')" alt="Maximize">
            </div>
            <div> {{ clusttaVersion }} </div>
        </div>
    </div>

</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watchEffect } from 'vue';
import { Events } from "@wailsio/runtime";
import emitter from '@/lib/mitt';
import utils from '@/services/utils';

import { useIconStore } from '@/stores/icons';
import { usePromptStore } from '@/stores/prompts';
import { useNotificationStore } from '@/stores/notifications';
import { usePlatformStore } from '@/stores/platform';

const platformStore = usePlatformStore();

const iconStore = useIconStore();
const promptStore = usePromptStore();
const notificationStore = useNotificationStore();

const props = defineProps({
    bgColor : { type: String, default: ''},
});

// vars
const notification = ref(false);
const notificationItem = ref(null);
const currentPrompt = ref(null);
const promptItem = ref(null);
const clusttaVersion = ref('');
const timer = ref(null);

// Restricted messages that should not trigger notifications
const restrictedMessages = [
  'no active account set',
  'no active account',
];

// const notification = ref({
//     "hasUndo": false,
//     "longMessage": "",
//     // "message": "Item Restored",
//     "message": "this is a very very very long message that has problems wrapping around itself this is a very very very long message that has problems wrapping around itself",
//     "read": false,
//     "type": "success"
// });

// computed props
const isOutdated = computed(() => {
  return false
});

const progressRunning = computed(() => {
  return notificationStore.progress.running;
});

const progressMinimized = computed(() => {
  return notificationStore.progress.isMinimized;
});

const isWriteOperation = computed(() => {
  return notificationStore.progress.operationType === 'write';
});

const progressTitle = computed(() => {
  return notificationStore.progress.title || '';
});

const progressPercentage = computed(() => {
  return Math.round(notificationStore.progress.percentage) || 0;
});

const progressCurrent = computed(() => {
  return notificationStore.progress.current || 0;
});

const progressTotal = computed(() => {
  return notificationStore.progress.total || 0;
});

const progressTooltip = computed(() => {
  const type = isWriteOperation.value ? 'Write' : 'Read';
  return `${progressTitle.value} (Click to restore)`;
});

// events
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

const handleUpdatePrompt = (payload) => {
  let promptData;
  if (typeof payload === 'string' || payload instanceof String) {
    promptData = JSON.parse(payload);
  } else {
    promptData = payload;
  }
  showPrompt(promptData);
};

const handleClearPrompt = () => {
  currentPrompt.value = null;
};

const handleClearAllPrompts = () => {
  currentPrompt.value = null;
};

// Register event listeners based on platform
if (platformStore.isWeb) {
  // Web mode: use mitt emitter (payload comes directly)
  emitter.on('add_message', handleAddMessage);
  emitter.on('add_prompt', handleAddPrompt);
  emitter.on('update_prompt', handleUpdatePrompt);
  emitter.on('clear_prompt', handleClearPrompt);
  emitter.on('clear_all_prompts', handleClearAllPrompts);
} else {
  // Desktop mode: use Wails Events (payload is in message.data)
  Events.On("add_message", async (message) => {
    handleAddMessage(message.data);
  });
  
  Events.On("add_prompt", async (prompt) => {
    handleAddPrompt(prompt.data);
  });
  
  Events.On("update_prompt", async (prompt) => {
    handleUpdatePrompt(prompt.data);
  });
  
  Events.On("clear_prompt", async () => {
    handleClearPrompt();
  });
  
  Events.On("clear_all_prompts", async () => {
    handleClearAllPrompts();
  });
}

// methods
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const getProgressIcon = () => {
  if (isWriteOperation.value) {
    return getAppIcon('info-triangle'); // Orange warning for write operations
  }
  return getAppIcon('info'); // Blue info for read operations
};

const restoreProgress = () => {
  notificationStore.restoreProgress();
};

const showMessage = async (data) => {

  // Check if message is restricted
  const messageText = data.message?.toLowerCase() || '';
  const isRestricted = restrictedMessages.some(restricted => 
    messageText.includes(restricted.toLowerCase())
  );
  
  if (isRestricted) {
    return;
  }
  
  notification.value = data
  clearTimeout(timer.value);
  timer.value = setTimeout(() => {
    notification.value = null;
  }, 6000);
}

const showPrompt = async (data) => {
  currentPrompt.value = data;
}

const clearNotification = () => {
  notification.value = null;
  timer.value = null
}

const clearPrompt = () => {
  currentPrompt.value = null;
}

const stopTimer = () => {
  clearTimeout(timer.value);
}

onMounted(async () => {
  clusttaVersion.value = await utils.getRawClusttaVersion();
});


</script>

<style scoped>
@import "@/assets/desktop.css";

.info-bar-root{
    width: 100%;
    height: 100%;
    height: 30px;
    display: flex;
    overflow: hidden;
    box-sizing: border-box;
    align-items: center;
    justify-content: space-between;
    color: var(--white);
    padding: 0 .8rem;
    font-size: 13px;
    font-weight: 300;
  }
  
.solid-background{
  background-color: var(--steel);
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
}

.oudated:hover{
  background-color: var(--dark-steel);
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
  gap: 1rem;
  border-radius: 3px;
  box-sizing: border-box;
  height: 70%;
  justify-content: flex-end;
  background-color: crimson;
  padding: .3rem .5rem;
  z-index: 99999;
}

.text-container{
    text-overflow: ellipsis;
    overflow: hidden;
}

.error {
  outline: solid 1px #FF3333;
  background-color:  #FF3333;
}

.success {
  background-color:  #20A41C;
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
  background-color: var(--light-steel);
}

.mini-progress.write-operation {
  /* background-color: rgba(238, 92, 8, 0.15); */
}

.mini-progress.write-operation:hover {
  background-color: rgba(238, 92, 8, 0.25);
}

.mini-progress-content {
  display: flex;
  align-items: center;
  gap: .5rem;
  color: var(--white);
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
  background-color: var(--light-steel);
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
  border-radius: 3px;
  box-sizing: border-box;
  height: 70%;
  justify-content: flex-start;
  /* background-color: var(--dark-steel); */
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

