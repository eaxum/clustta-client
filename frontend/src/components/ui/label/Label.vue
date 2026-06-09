<script setup>
import { LabelRoot } from 'radix-vue'
import { cn } from '@/lib/utils'

defineProps({
  for: { type: String, default: undefined },
  class: { type: String, default: '' },
})
</script>

<template>
  <LabelRoot :for="$props.for" :class="cn('text-sm font-medium leading-none text-foreground peer-disabled:cursor-not-allowed peer-disabled:opacity-70', $props.class)">
    <slot />
  </LabelRoot>
</template>
