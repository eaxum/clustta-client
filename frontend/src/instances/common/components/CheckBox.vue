<template>
  <Checkbox :checked="modelValue" :disabled="disabled" @update:checked="handleChange" />
</template>

<script setup>
import { Checkbox } from '@/components/ui/checkbox';

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  disabled: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['update:modelValue', 'change']);

// Handles checkbox state changes and emits both v-model and change events.
const handleChange = (value) => {
  emit('update:modelValue', value);
  emit('change', value);
};
</script>
