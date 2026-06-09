<script setup>
import { cn } from '@/lib/utils'

defineProps({
  class: { type: String, default: '' },
})
</script>

<template>
  <div :class="cn('animate-pulse rounded-md bg-primary/10', $props.class)" />
</template>
