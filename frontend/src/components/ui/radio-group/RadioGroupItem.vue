<script setup>
import { RadioGroupIndicator, RadioGroupItem } from 'radix-vue'
import { cn } from '@/lib/utils'

defineProps({
  value: { type: String, required: true },
  disabled: { type: Boolean, default: false },
  id: { type: String, default: undefined },
  class: { type: String, default: '' },
})
</script>

<template>
  <RadioGroupItem :value="value" :disabled="disabled" :id="id" :class="cn('aspect-square h-4 w-4 rounded-full border border-primary text-primary shadow focus:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50', $props.class)">
    <RadioGroupIndicator class="flex items-center justify-center">
      <svg class="h-2.5 w-2.5 fill-primary" viewBox="0 0 24 24"><circle cx="12" cy="12" r="12" /></svg>
    </RadioGroupIndicator>
  </RadioGroupItem>
</template>
