<template>
  <Switch :checked="switchValue" :disabled="disabled" :class="online ? 'data-[state=checked]:bg-green-500' : ''" />
</template>

<script setup>
import { computed } from 'vue';
import { Switch } from '@/components/ui/switch';

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
  disabled: {
    type: Boolean,
    default: false
  },
  switchValueFunc: {
    type: Function,
    default: () => false
  },
});

// Resolves the switch value from either a prop or a function.
const switchValue = computed(() => {
  if (props.useFunction) {
    return props.switchValueFunc;
  }
  return props.switchValueProp;
});
</script>


