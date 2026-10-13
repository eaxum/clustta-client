<template>
  <label class="toggle-container">
    <input :checked="switchValue" class="toggle-input" type="checkbox" disabled />
    <span
      class="toggle-switch"
      :class="{
        'use-online': online,
        'toggle-switch-animated': animationReady,
        'toggle-switch-on': switchValue,
        'toggle-switch-off': !switchValue,
      }"
      @animationend="finishAnimation"
    ></span>
  </label>
</template>

<script setup>
import { computed, ref, watch } from 'vue';

const props = defineProps({
  switchValueProp: {
    type: Boolean,
    default: false,
  },
  useFunction: {
    type: Boolean,
    default: false,
  },
  online: {
    type: Boolean,
    default: false,
  },
  switchValueFunc: {
    type: Function,
    default: () => false,
  },
});

const animationReady = ref(false);
const switchValue = computed(() => {
  if (props.useFunction) return props.switchValueFunc;
  return props.switchValueProp;
});

watch(switchValue, () => {
  animationReady.value = true;
});

const finishAnimation = () => {
  animationReady.value = false;
};
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
  background-color: var(--selected);
}

.toggle-input:checked + .use-online {
  background-color: var(--online);
}

.toggle-input:checked + .toggle-switch::before {
  border-color: var(--primaryDarkestColor);
  transform: translateX(var(--switch-travel));
  transform-origin: right center;
  background-color: white;
}

.toggle-input:focus:checked + .toggle-switch::before {
  border-color: var(--primaryDarkestColor);
}

.toggle-switch {
  --switch-container-width: 36px;
  --switch-size: calc(var(--switch-container-width) / 1.7);
  --switch-travel: calc(var(--switch-container-width) - var(--switch-size));
  --switch-motion-trailing-distance: 5px;
  --switch-motion-overshoot: 2px;
  display: flex;
  align-items: center;
  position: relative;
  height: var(--switch-size);
  flex-basis: var(--switch-container-width);
  border-radius: 12px;
  background-color: var(--primaryDarkestColor);
  flex-shrink: 0;
  transition:
    background-color var(--motion-duration-standard) ease-in-out,
    border-radius var(--motion-duration-standard) ease-in-out;
}

.toggle-switch::before {
  content: "";
  position: absolute;
  left: 4px;
  height: calc(var(--switch-size) - 8px);
  width: calc(var(--switch-size) - 8px);
  border-radius: 12px;
  background-color: rgba(255, 255, 255, 0.274);
  transition:
    transform var(--motion-duration-standard) ease-in-out,
    border-radius var(--motion-duration-standard) ease-in-out;
  transform-origin: left center;
  will-change: transform;
}

.toggle-container:hover .toggle-switch {
  border-radius: var(--small-radius);
}

.toggle-container:hover .toggle-switch::before {
  border-radius: var(--tiny-radius);
  transform: scale(0.72, 0.92);
}

.toggle-container:hover .toggle-input:checked + .toggle-switch::before {
  transform: translateX(var(--switch-travel)) scale(0.72, 0.92);
}

.toggle-switch-animated::before {
  transition: none;
}

.toggle-switch-animated.toggle-switch-on::before {
  animation: toggle-thumb-on var(--motion-duration-slow) var(--motion-easing-bounce) both;
}

.toggle-switch-animated.toggle-switch-off::before {
  animation: toggle-thumb-off var(--motion-duration-slow) var(--motion-easing-bounce) both;
}

@keyframes toggle-thumb-on {
  0% {
    transform: translateX(0) scale(0.72, 0.92);
  }

  42% {
    transform: translateX(calc(var(--switch-travel) - var(--switch-motion-trailing-distance))) scale(1.35, 0.78);
  }

  70% {
    transform: translateX(calc(var(--switch-travel) + var(--switch-motion-overshoot))) scale(0.86, 1.18);
  }

  100% {
    transform: translateX(var(--switch-travel)) scale(1);
  }
}

@keyframes toggle-thumb-off {
  0% {
    transform: translateX(var(--switch-travel)) scale(0.72, 0.92);
  }

  42% {
    transform: translateX(var(--switch-motion-trailing-distance)) scale(1.35, 0.78);
  }

  70% {
    transform: translateX(calc(0px - var(--switch-motion-overshoot))) scale(0.86, 1.18);
  }

  100% {
    transform: translateX(0) scale(1);
  }
}
</style>


