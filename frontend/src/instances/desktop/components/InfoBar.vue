<template>
    <div class="info-bar-root" :style="{ backgroundColor : bgColor }" >

        <div v-if="notification" ref="notificationItem" :class="['message', notification.type]" @mouseenter="stopTimer()"
            @mouseleave="showMessage(notification)">
            <span class="text-container" >{{ notification.message }}</span>
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
import utils from '@/services/utils';

import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();

const props = defineProps({
    bgColor : { type: String, default: ''},
});

// vars
const notification = ref(false);
const notificationItem = ref(null);
const clusttaVersion = ref('');
const timer = ref(null);

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

// events
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

// methods
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const showMessage = async (data) => {
  // console.log(data)
  notification.value = data
  clearTimeout(timer.value);
  timer.value = setTimeout(() => {
    notification.value = null;
  }, 6000);
}

const clearNotification = () => {
  notification.value = null;
  timer.value = null
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
    justify-content: flex-end;
    color: var(--white);
    padding: 0 .8rem;
    font-size: 12px;
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
</style>

