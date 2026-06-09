<script setup>
import { cn } from '@/lib/utils'

const props = defineProps({
  defaultValue: { type: [String, Number], default: undefined },
  modelValue: { type: [String, Number], default: undefined },
  class: { type: String, default: '' },
  type: { type: String, default: 'text' },
  placeholder: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
})

const emits = defineEmits(['update:modelValue'])

function handleInput(event) {
  emits('update:modelValue', event.target.value)
}
</script>

<template>
  <input :type="type" :value="modelValue" :placeholder="placeholder" :disabled="disabled" :class="cn('flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50', props.class)" @input="handleInput" />
</template>
