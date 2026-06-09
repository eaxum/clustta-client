<script setup>
import { AvatarFallback, AvatarImage, AvatarRoot } from 'radix-vue'
import { cn } from '@/lib/utils'

defineProps({
  src: { type: String, default: '' },
  alt: { type: String, default: '' },
  fallback: { type: String, default: '' },
  class: { type: String, default: '' },
  size: { type: String, default: 'default' },
})

const sizeClasses = {
  sm: 'h-8 w-8 text-xs',
  default: 'h-10 w-10 text-sm',
  lg: 'h-12 w-12 text-base',
}
</script>

<template>
  <AvatarRoot :class="cn('relative flex shrink-0 overflow-hidden rounded-full', sizeClasses[size] || sizeClasses.default, $props.class)">
    <AvatarImage v-if="src" :src="src" :alt="alt" class="aspect-square h-full w-full" />
    <AvatarFallback :class="cn('flex h-full w-full items-center justify-center rounded-full bg-muted text-muted-foreground')">
      <slot>{{ fallback }}</slot>
    </AvatarFallback>
  </AvatarRoot>
</template>
