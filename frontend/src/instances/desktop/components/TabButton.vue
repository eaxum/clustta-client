<template>
  <span v-stop-propagation @click="buttonFunction" 
  :class="{'button-background': useBackground, 'full-width' : fullWidth, 'active-button' : isActive, 
  'icon-after' : iconAfter, 'centered' : centered, 'is-disabled': stage.operationActive }" class="action-button">
    <img v-if="!iconAfter" class="extra-large-icons" :src="icon">
    <div v-if="showLabel">{{ label }}</div>
    <img v-if="iconAfter" class="extra-large-icons" :src="icon">
  </span>
</template>
  
<script setup>

import { useModalStore } from '@/stores/modals';
import { useTrayStates } from '@/stores/TrayStates';
import { useStageStore } from '@/stores/stages';

const stage = useStageStore();
const modalStore = useModalStore();

const editParams = (itemType) => {
  modalStore.setModalVisibility(itemType, true);
};

const props = defineProps({
  icon: String,
  label: String,
  buttonFunction: Function,
  showLabel: { type: Boolean, default: false},
  iconAfter: { type: Boolean, default: false},
  centered: { type: Boolean, default: false},
  useBackground: { type: Boolean, default: false},
  fullWidth: { type: Boolean, default: false},
  isActive: { type: Boolean, default: false},

});
 
</script>
  
<style scoped>
  @import "@/assets/desktop.css";

.action-button{
  overflow: hidden;
  background-color: transparent;
  text-align: center;
  font-size: 14px;
  line-height: 14px;
  color: var(--white);
  position: relative;
  border-radius: 8px;
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  gap: 10px;
  align-items: center;
  padding: 4px;
  min-height: max-content;
  /* height: 30px; */
  min-width: max-content;
  opacity: .4;
  transition: all 0.3s ease;
  /* background-color: firebrick; */

}

.active-button{
  opacity: 1;
}

.action-button:hover{
  opacity: 7;
  /* background-color: #ffffff15; */
}
.action-button:active{
  opacity: 1;
  /* background-color: #00000013; */
}

.action-button-pressed {
  box-sizing: border-box;
  background-color: rgba(0, 0, 0, 0.216);
  outline: solid 1px var(--white);
  outline-offset: -1px;
}

.button-background{
  background-color: rgba(0, 0, 0, 0.216);
  border-radius: var(--small-radius);
  padding: .3rem 1rem;
}
.full-width{
  /* background-color: red; */
  /* padding: 1.2rem .5rem; */
  width: 100%;
}

.icon-after{
  justify-content: space-between;
}

.centered{
  justify-content: space-around;
}
</style>
  
  
  

