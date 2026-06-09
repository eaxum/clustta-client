<script setup>
import { CheckboxIndicator, CheckboxRoot } from 'radix-vue'
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps({
  defaultChecked: { type: Boolean, default: undefined },
  checked: { type: [Boolean, String], default: undefined },
  disabled: { type: Boolean, default: false },
  required: { type: Boolean, default: false },
  name: { type: String, default: undefined },
  value: { type: String, default: 'on' },
  id: { type: String, default: undefined },
  class: { type: String, default: '' },
})

const emits = defineEmits(['update:checked'])

const modelValue = computed({
  get: () => props.checked,
  set: (val) => emits('update:checked', val),
})
</script>

<template>
  <CheckboxRoot v-model:checked="modelValue" :disabled="disabled" :required="required" :name="name" :value="value" :id="id" :class="cn('peer h-4 w-4 shrink-0 rounded-sm border border-primary shadow focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground', props.class)">
    <CheckboxIndicator class="flex items-center justify-center text-current">
      <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12" /></svg>
    </CheckboxIndicator>
  </CheckboxRoot>
</template>
