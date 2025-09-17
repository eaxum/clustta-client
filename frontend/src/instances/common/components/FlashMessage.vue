<template>
  <div v-if="progressRunning && !stageStore.operationActive" v-stop-propagation class="desktop-overlay-mask"></div>

  <div v-if="progressRunning && !stageStore.operationActive" class="flash-area" ref="flashArea"
    :class="{ 'flash-area-desktop': isDesktop, 'flash-area-desktop-progress': progressRunning }">

    <div class="progress-bar">

      <div class="progress-bar-header">
        <HeaderArea :title="notificationStore.getProgress.title" :icon="'info'" :showSearch="showSearch" />
      </div>

      <div class="progress-bar-meta">
        <span class="progress-bar-message">{{ notificationStore.getProgress.message }}</span>
        <div class="progress-bar-total">{{ notificationStore.getProgress.current }}/{{
          notificationStore.getProgress.total }}</div>
      </div>

      <div class="progress-bar-loader tint">
        <ProgressBar :taskProgress="notificationStore.getProgress.percentage" v-stop-propagation />
      </div>

      <div class="pop-up-actions">

        <div v-if="notificationStore.getProgress.extra_message" class="pop-up-info">
          <div class="pop-up-stats" :style="{ color : message.color} "> 
            {{ throttledExtraMessage }}
          </div>
        </div>
        <button v-if="notificationStore.canCancel" class="button" @click="cancelOperation" v-stop-propagation>
          <div class="login-button-text">
            {{ isAwaitingResponse ? 'Cancelling...' : 'Cancel' }}
          </div>
          <div v-if="isAwaitingResponse" class="login-button-icon loading-icon">
            <img src="/icons/loading.svg" />
          </div>
        </button>
      </div>


    </div>

  </div>



</template>

<script setup>
import { watch, ref, computed, onMounted, onBeforeUnmount } from 'vue';
import { useTrayStates } from '@/stores/TrayStates';
import ProgressBar from '@/instances/common/components/ProgressBar.vue';
import { useNotificationStore } from '@/stores/notifications';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import { Events } from "@wailsio/runtime";
import { useStageStore } from '@/stores/stages';

// vars
let showSearch = false;

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
const progressRunning = computed(() => { return notificationStore.getProgress.running })
const files = ref(214);

const throttledExtraMessage = ref(notificationStore.getProgress.extra_message);
let extraMessageTimeout = null;
let lastExtraMessage = notificationStore.getProgress.extra_message;

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

const convertCount = (value) => {
  if (value < 0 || value > 100) {
    throw new Error("Value must be between 0 and 100");
  }
  const newValue = (value / 100) * files.value;
  return Math.round(newValue);
};

Events.On("add_message", async (message) => {
  const payload = message.data;
  let notificationData
  if (typeof payload === 'string' || payload instanceof String) {
    notificationData = JSON.parse(payload);
  } else {
    notificationData = payload;
  }
  showMessage(notificationData)
});

// const notification = ref(null);
const notification = ref(false);
// const notification = ref({
//     "hasUndo": false,
//     "longMessage": "",
//     "message": "Item Restored",
//     "read": false,
//     "type": "success"
// });


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
async function clearNotification() {
  notification.value = null;
  timer.value = null
}
async function undo() {
  trayStates.undoFunction()
  notification.value = null;
  timer.value = null
}

function stopTimer() {
  clearTimeout(timer.value);
}

const handleClickOutside = (event) => {
  if (notification.value && (event.target !== notificationItem.value)) {
    // //console.log('outside');
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

onMounted(() => {
  // //console.log(listBoxParent.value.getBoundingClientRect().width);
  document.addEventListener('click', handleClickOutside);

});

onBeforeUnmount(() => {
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
  color: var(--white);
  /* background-color: #20A41C; */
  height: 100%;
  min-height: 100%;
  font-weight: 500;
}

.login-button-text {
  font-size: 14px;
}

.login-button-icon {
  filter: invert(100%);
  width: 20px;
  height: 20px;
}

.button {
  display: flex;
  flex-direction: row;
  justify-content: center;
  align-items: center;
  border-radius: 16px;
  padding: 10px 18px;
  border-color: rgba(234, 236, 240, 1);
  border-width: 1px;
  border-style: solid;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  height: 35px;
  overflow: hidden;
  width: max-content;
  min-width: 100px;

  width: max-content;
  min-width: 80px;
  padding: .2rem .8rem;
  gap: .2rem;
  height: 40px;
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
  /* margin: auto; */
  display: flex;
  transition: opacity 0.3s ease;
  background-color: rgba(0, 0, 0, 0.5);
  /* background-color: var(--dark-steel); */
  /* background-color: red; */
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
  top: 110px;
  right: 30px;
  width: 40%;
  max-width: 400px;
  /* min-width: 250px; */
  /* width: 1000px; */
  /* background-color: red; */
}

.flash-area-desktop-progress {
  top: 0px;
  right: 0px;
  /* background-color: red; */
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
  background-color: var(--black);
  border-radius: var(--large-radius);

}

.progress-bar-meta {
  color: var(--white);
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
  color: var(--white);
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
  color: var(--white);
  display: flex;
  align-items: center;
  font-size: 16px;
  justify-content: space-between;
  /* background-color: #FF3333; */
  width: 100%;
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
  outline: solid 1px var(--white);
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


