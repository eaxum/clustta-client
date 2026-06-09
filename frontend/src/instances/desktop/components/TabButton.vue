<template>
  <span v-stop-propagation @click="buttonFunction" 
  :class="{'button-background': useBackground, 'full-width' : fullWidth, 'active-button' : isActive, 
  'icon-after' : iconAfter, 'centered' : centered, 'is-disabled': stage.operationActive }" class="action-button">
    <img v-if="!iconAfter && smallIcons" class="small-icons" :src="icon">
    <img v-else-if="!smallIcons" class="extra-large-icons" :src="icon">
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
  smallIcons: { type: Boolean, default: false},

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
  color: hsl(var(--foreground));
  position: relative;
  border-radius: var(--normal-radius);
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
  /* background-color: hsl(var(--destructive)); */

}

.active-button{
  opacity: 1;
}

.action-button:hover{
  opacity: 7;
  /* background-color: hsl(var(--accent)); */
}
.action-button:active{
  opacity: 1;
  /* background-color: hsl(var(--accent)); */
}

.action-button-pressed {
  box-sizing: border-box;
  background-color: hsl(var(--muted));
  border: 1px solid hsl(var(--border));
  
}

.button-background{
  background-color: hsl(var(--muted));
  border-radius: var(--small-radius);
  padding: .3rem 1rem;
}
.full-width{
  /* background-color: hsl(var(--destructive)); */
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
  
  
  

