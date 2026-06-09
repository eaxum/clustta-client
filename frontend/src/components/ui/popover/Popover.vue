<script setup>
import { PopoverContent, PopoverPortal, PopoverRoot, PopoverTrigger } from 'radix-vue'
import { cn } from '@/lib/utils'

defineProps({
  class: { type: String, default: '' },
  sideOffset: { type: Number, default: 4 },
  align: { type: String, default: 'center' },
  side: { type: String, default: 'bottom' },
})
</script>

<template>
  <PopoverRoot>
    <PopoverTrigger as-child>
      <slot name="trigger" />
    </PopoverTrigger>

    <PopoverPortal>
      <PopoverContent :side-offset="sideOffset" :align="align" :side="side" :class="cn('z-50 w-72 rounded-md border bg-popover p-4 text-popover-foreground shadow-md outline-none data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2', $props.class)">
        <slot />
      </PopoverContent>
    </PopoverPortal>
  </PopoverRoot>
</template>
