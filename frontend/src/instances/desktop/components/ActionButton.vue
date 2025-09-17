<template>
  <span v-stop-propagation @click="buttonFunction" :style="{ backgroundColor: color }" :class="{
    'button-background': useBackground, 'alert-background': isAlert, 'full-width': fullWidth, 'outline': useOutline, 'icon-after': iconAfter, 'centered':
      centered, 'button-active': isActive, 'is-inactive': isInactive, 'is-disabled': isDead, 'plain-background' : plainBackground,
  }" class="action-button">
    <img v-if="showIcon && !iconAfter" class="small-icons no-cursor" :class="{ 'no-filter' : noFilter }" :src="icon">
    <div v-if="showLabel || label" class="small-icons button-label no-cursor">{{ label }}</div>
    <img v-if="showIcon && iconAfter" class="small-icons no-cursor" :class="{ 'no-filter' : noFilter }" :src="icon">
  </span>
</template>

<script setup>
import { computed } from 'vue';
import { useStageStore } from '@/stores/stages';

const stage = useStageStore();


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
  isInactive: { type: Boolean, default: false },
  isDisabled: { type: Boolean, default: false },
  isActive: { type: Boolean, default: false },
  fullWidth: { type: Boolean, default: false },
  allowDeactivate: { type: Boolean, default: false },

});

const isDead = computed(() => {
  const notEnabled = props.isDisabled || stage.operationActive
  return notEnabled && !props.allowDeactivate
})

</script>

<style scoped>
@import "@/assets/desktop.css";


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

}

[data-theme="dark"] .action-button:hover{
  background-color: #ffffff15;
}

[data-theme="dark"] .action-button:active {
  background-color: #00000013;
}

.action-button:hover {
  background-color: rgba(0, 0, 0, 0.11);
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
  outline: var(--solid-line);
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

</style>

