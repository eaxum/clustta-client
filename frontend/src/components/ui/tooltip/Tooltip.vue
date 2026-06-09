<script setup>
import { TooltipContent, TooltipProvider, TooltipRoot, TooltipTrigger, TooltipPortal } from 'radix-vue'
import { cn } from '@/lib/utils'

defineProps({
  content: { type: String, default: '' },
  side: { type: String, default: 'top' },
  sideOffset: { type: Number, default: 4 },
  delayDuration: { type: Number, default: 200 },
  class: { type: String, default: '' },
})
</script>

<template>
  <TooltipProvider :delay-duration="delayDuration">
    <TooltipRoot>
      <TooltipTrigger as-child>
        <slot />
      </TooltipTrigger>

      <TooltipPortal>
        <TooltipContent :side="side" :side-offset="sideOffset" :class="cn('z-50 overflow-hidden rounded-md bg-primary px-3 py-1.5 text-xs text-primary-foreground animate-in fade-in-0 zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2', $props.class)">
          <slot name="content">{{ content }}</slot>
        </TooltipContent>
      </TooltipPortal>
    </TooltipRoot>
  </TooltipProvider>
</template>
