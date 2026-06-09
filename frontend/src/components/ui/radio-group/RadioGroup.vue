<script setup>
import { RadioGroupIndicator, RadioGroupItem, RadioGroupRoot } from 'radix-vue'
import { cn } from '@/lib/utils'

defineProps({
  modelValue: { type: String, default: undefined },
  defaultValue: { type: String, default: undefined },
  disabled: { type: Boolean, default: false },
  orientation: { type: String, default: 'vertical' },
  class: { type: String, default: '' },
})

const emits = defineEmits(['update:modelValue'])
</script>

<template>
  <RadioGroupRoot :model-value="modelValue" :default-value="defaultValue" :disabled="disabled" :orientation="orientation" :class="cn('grid gap-2', $props.class)" @update:model-value="emits('update:modelValue', $event)">
    <slot />
  </RadioGroupRoot>
</template>
