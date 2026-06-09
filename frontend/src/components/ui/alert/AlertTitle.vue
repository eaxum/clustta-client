<script setup>
import { cn } from '@/lib/utils'

defineProps({
  class: { type: String, default: '' },
})
</script>

<template>
  <h5 :class="cn('mb-1 font-medium leading-none tracking-tight', $props.class)">
    <slot />
  </h5>
</template>
