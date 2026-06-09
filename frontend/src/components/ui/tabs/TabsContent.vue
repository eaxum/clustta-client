<script setup>
import { TabsContent } from 'radix-vue'
import { cn } from '@/lib/utils'

defineProps({
  value: { type: String, required: true },
  class: { type: String, default: '' },
})
</script>

<template>
  <TabsContent :value="value" :class="cn('mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2', $props.class)">
    <slot />
  </TabsContent>
</template>
