<script setup>
import { TabsContent, TabsList, TabsRoot, TabsTrigger } from 'radix-vue'
import { cn } from '@/lib/utils'

defineProps({
  defaultValue: { type: String, default: undefined },
  modelValue: { type: String, default: undefined },
  class: { type: String, default: '' },
})

const emits = defineEmits(['update:modelValue'])
</script>

<template>
  <TabsRoot :default-value="defaultValue" :model-value="modelValue" :class="cn($props.class)" @update:model-value="emits('update:modelValue', $event)">
    <slot />
  </TabsRoot>
</template>
