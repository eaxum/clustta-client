<script setup>
import { SeparatorRoot } from 'radix-vue'
import { cn } from '@/lib/utils'

defineProps({
  orientation: { type: String, default: 'horizontal' },
  decorative: { type: Boolean, default: true },
  class: { type: String, default: '' },
})
</script>

<template>
  <SeparatorRoot :orientation="orientation" :decorative="decorative" :class="cn('shrink-0 bg-border', orientation === 'horizontal' ? 'h-px w-full' : 'h-full w-px', $props.class)" />
</template>
