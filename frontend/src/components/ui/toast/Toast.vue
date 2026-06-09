<script setup>
import { ref, computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps({
  variant: { type: String, default: 'default' },
  title: { type: String, default: '' },
  description: { type: String, default: '' },
  class: { type: String, default: '' },
  duration: { type: Number, default: 5000 },
  open: { type: Boolean, default: true },
})

const emits = defineEmits(['update:open', 'close'])

const variantClasses = {
  default: 'border bg-background text-foreground',
  destructive: 'border-destructive bg-destructive text-destructive-foreground',
  success: 'border-online bg-background text-foreground',
}

function close() {
  emits('update:open', false)
  emits('close')
}
</script>

<template>
  <div v-if="open" :class="cn('pointer-events-auto relative flex w-full items-center justify-between space-x-2 overflow-hidden rounded-md border p-4 shadow-lg transition-all', variantClasses[variant] || variantClasses.default, props.class)">
    <div class="grid gap-1">
      <div v-if="title" class="text-sm font-semibold">{{ title }}</div>
      <div v-if="description" class="text-sm opacity-90">{{ description }}</div>
      <slot />
    </div>

    <button class="absolute right-1 top-1 rounded-md p-1 opacity-0 transition-opacity hover:opacity-100 group-hover:opacity-100 focus:opacity-100" @click="close">
      <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" /></svg>
    </button>
  </div>
</template>
