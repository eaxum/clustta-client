<script setup>
import { cn } from '@/lib/utils'

const props = defineProps({
  modelValue: { type: String, default: undefined },
  class: { type: String, default: '' },
  placeholder: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
})

const emits = defineEmits(['update:modelValue'])

function handleInput(event) {
  emits('update:modelValue', event.target.value)
}
</script>

<template>
  <textarea :value="modelValue" :placeholder="placeholder" :disabled="disabled" :class="cn('flex min-h-[60px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50', props.class)" @input="handleInput" />
</template>
