<script setup>
import { DialogClose, DialogContent, DialogDescription, DialogOverlay, DialogPortal, DialogRoot, DialogTitle, DialogTrigger } from 'radix-vue'
import { cn } from '@/lib/utils'

defineProps({
  open: { type: Boolean, default: undefined },
  defaultOpen: { type: Boolean, default: false },
})

const emits = defineEmits(['update:open'])
</script>

<template>
  <DialogRoot :open="open" :default-open="defaultOpen" @update:open="emits('update:open', $event)">
    <slot />
  </DialogRoot>
</template>
