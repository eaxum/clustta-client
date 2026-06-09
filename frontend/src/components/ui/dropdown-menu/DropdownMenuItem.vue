<script setup>
import { DropdownMenuItem } from 'radix-vue'
import { cn } from '@/lib/utils'

defineProps({
  disabled: { type: Boolean, default: false },
  class: { type: String, default: '' },
  inset: { type: Boolean, default: false },
})
</script>

<template>
  <DropdownMenuItem :disabled="disabled" :class="cn('relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none transition-colors focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50', inset && 'pl-8', $props.class)">
    <slot />
  </DropdownMenuItem>
</template>
