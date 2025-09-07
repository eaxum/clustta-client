<template>
    <label class="toggle-container"> 
      <input v-model="switchValue" class="toggle-input" type="checkbox" disabled/>
      <span class="toggle-switch" :class="{'use-online' : online }"></span>
    </label>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";
import { useTrayStates } from '@/stores/TrayStates';
const trayStates = useTrayStates();

const switchValue = computed(() => { 
  if(props.useFunction){
    return props.switchValueFunc 
  } else {
    return props.switchValueProp 
  }
});

const props = defineProps({
  switchValueProp: {
    type: Boolean,
    default: false
  },
  useFunction: {
    type: Boolean,
    default: false
  },
  online: {
    type: Boolean,
    default: false
  },
  switchValueFunc: {
    type: Function,
    default: () => {
      return false
    }
  },
});


</script>

<style scoped>
.toggle-container {
  --primaryDarkestColor: rgba(0, 0, 0, 0.384);
  --secondaryLighterColor: #e0e0e0ff;
  height: max-content;
  position: relative;
  cursor: pointer;
  display: flex;
  align-items: center;
  width: 40px;
  height: 100%;
  /* height: 40px; */
  /* background-color: chartreuse; */
}

.toggle-input {
  pointer-events: none;
  position: absolute;
  width: 10px;
  height: 10px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}

.toggle-input:checked + .toggle-switch {
  background-color: var(--task-item-selected);
}

.toggle-input:checked +  .use-online {
  background-color: var(--online);
}


.toggle-input:checked + .toggle-switch::before {
  border-color: var(--primaryDarkestColor);
  transform: translateX(calc(var(--switch-container-width) - var(--switch-size)));
  background-color: white;
}

.toggle-input:focus:checked + .toggle-switch::before {
  border-color: var(--primaryDarkestColor);
}

.toggle-switch {
  --switch-container-width: 36px;
  --switch-size: calc(var(--switch-container-width) / 1.7);
  display: flex;
  align-items: center;
  position: relative;
  height: var(--switch-size);
  flex-basis: var(--switch-container-width);
  border-radius: 8px;
  background-color: var(--primaryDarkestColor);
  flex-shrink: 0;
  transition: background-color 0.15s ease-in-out;
}

.toggle-switch::before {
  content: "";
  position: absolute;
  left: 4px;
  height: calc(var(--switch-size) - 8px);
  width: calc(var(--switch-size) - 8px);
  border-radius: 5px;
  background-color: rgba(255, 255, 255, 0.274);
  transition: transform 0.15s ease-in-out;
}
</style>


