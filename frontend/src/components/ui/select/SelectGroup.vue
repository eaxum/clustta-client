<script setup>
import { SelectGroup, SelectLabel } from 'radix-vue'
import { cn } from '@/lib/utils'

defineProps({
  label: { type: String, default: '' },
  class: { type: String, default: '' },
})
</script>

<template>
  <SelectGroup>
    <SelectLabel v-if="label" :class="cn('px-2 py-1.5 text-sm font-semibold', $props.class)">{{ label }}</SelectLabel>
    <slot />
  </SelectGroup>
</template>
