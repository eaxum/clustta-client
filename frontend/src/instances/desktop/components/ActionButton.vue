<template>
  <span v-stop-propagation @click="buttonFunction" :style="{ backgroundColor: color }" :class="{
    'button-background': useBackground, 'alert-background': isAlert, 'full-width': fullWidth, 'outline': useOutline, 'icon-after': iconAfter, 'centered':
      centered, 'button-active': isActive, 'is-inactive': isInactive, 'is-disabled': isDead, 'plain-background' : plainBackground, 'use-alert': useAlert, 'use-danger': useDanger,
  }" class="action-button">
    <img v-if="showIcon && !iconAfter" class="small-icons no-cursor" :class="{ 'no-filter' : noFilter, 'loading-icon' : isLoading }" :src="icon">
    <div v-if="showLabel || label" class="small-icons button-label no-cursor">{{ label }}</div>
    <img v-if="showIcon && iconAfter" class="small-icons no-cursor" :class="{ 'no-filter' : noFilter }" :src="icon">
  </span>
</template>

<script setup>
import { computed } from 'vue';
import { useStageStore } from '@/stores/stages';
import { useNotificationStore } from '@/stores/notifications';

const stage = useStageStore();
const notificationStore = useNotificationStore();


const props = defineProps({
  icon: String,
  label: String,
  color: String,
  buttonFunction: Function,
  noFilter: { type: Boolean, default: false },
  showIcon: { type: Boolean, default: true },
  smallIcons: { type: Boolean, default: false },
  showLabel: { type: Boolean, default: false },
  iconAfter: { type: Boolean, default: false },
  plainBackground: { type: Boolean, default: false },
  useOutline: { type: Boolean, default: false },
  centered: { type: Boolean, default: false },
  useBackground: { type: Boolean, default: false },
  isAlert: { type: Boolean, default: false },
  isLoading: { type: Boolean, default: false },
  useAlert: { type: Boolean, default: false },
  useDanger: { type: Boolean, default: false },
  isInactive: { type: Boolean, default: false },
  isDisabled: { type: Boolean, default: false },
  isActive: { type: Boolean, default: false },
  fullWidth: { type: Boolean, default: false },
  allowDeactivate: { type: Boolean, default: false },

});

const isDead = computed(() => {
  // Check if write operation is active
  const writeOperationActive = notificationStore.progress.running && 
                                notificationStore.progress.operationType === 'write';
  
  // Button is disabled if:
  // 1. Explicitly disabled via prop, OR
  // 2. Stage operation is active, OR  
  // 3. Write operation is running (unless allowDeactivate is true)
  const notEnabled = props.isDisabled || 
                     stage.operationActive || 
                     writeOperationActive;
  
  return notEnabled && !props.allowDeactivate;
})

</script>

<style scoped>
@import "@/assets/desktop.css";

@keyframes loadingRotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.loading-icon {
  width: 24px;
  height: 24px;
  padding: 0;
  overflow: hidden;
  animation: loadingRotate .5s linear infinite;
}

.chevron {
  pointer-events: none;
  height: 10px;
  min-width: 10px;
  transition: all .1s ease-in;
}

.action-button {
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
  align-items: center;
  padding: .3rem;
  height: max-content;
  width: max-content;
  min-width: max-content;
  min-height: max-content;
  transition: all 0.3s ease;
  opacity: 1;
  border-radius: var(--normal-radius);
  /* background-color: crimson; */
}

/* [data-theme="dark"] .action-button:hover{
  background-color: #ffffff15;
  background-color: var(--hover);
} */

[data-theme="dark"] .action-button:active {
  background-color: #00000013;
}

.action-button:hover {
  /* background-color: #09ff09bc; */
  background-color: var(--hover);
}

.action-button:active {
  background-color: rgba(0, 0, 0, 0.11);
}

.action-button-pressed {
  box-sizing: border-box;
  background-color: rgba(0, 0, 0, 0.216);
  outline: solid 1px var(--white);
  outline-offset: -1px;
}

.button-background {
  background-color: rgba(0, 0, 0, 0.216);
  border-radius: var(--small-radius);
  background-color: rgb(44, 117, 226);
  /* padding: .3rem 1rem; */
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.plain-background {
  background-color: var(--steel);
  background-color: rgba(0, 0, 0, 0.562);
  outline: 0px;
}

.button-background:hover {
  background-color: rgb(78, 137, 226);
}

.alert-background {
  background-color: rgb(238, 92, 8);
}

.alert-background:hover {
  background-color: rgb(238, 101, 21);
}

.full-width {
  /* background-color: red; */
  /* padding: 1.2rem .5rem; */
  width: 100%;
}

.outline {
  outline: var(--transparent-line);
  outline-offset: -1px;
  padding-right: .5rem;
  padding-left: .5rem;

}

.icon-after {
  justify-content: space-between;
}

.centered {
  justify-content: space-around;
}

.button-active {
  background-color: var(--midnight-steel);
  outline: var(--transparent-line);
}

.no-cursor {
  pointer-events: none;
}

.button-label{
  font-weight: 350 ;
}

[data-theme="dark"] .button-label{
  font-weight: 200 ;
}

.is-inactive {
  pointer-events: none;
}

.is-disabled{
  opacity: .5;
}

[data-theme="dark"] .use-alert img {
  filter: brightness(0) saturate(100%) invert(88%) sepia(45%) saturate(566%) hue-rotate(359deg) brightness(97%) contrast(92%);
}

.use-alert img {
  filter: brightness(0) saturate(100%) invert(60%) sepia(72%) saturate(489%) hue-rotate(1deg) brightness(92%) contrast(90%);
}

[data-theme="dark"] .use-danger img {
  filter: brightness(0) saturate(100%) invert(18%) sepia(95%) saturate(7471%) hue-rotate(347deg) brightness(88%) contrast(93%);
}

.use-danger img {
  filter: brightness(0) saturate(100%) invert(18%) sepia(95%) saturate(7471%) hue-rotate(347deg) brightness(88%) contrast(93%);
}

</style>

