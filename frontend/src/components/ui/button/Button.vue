<script setup>
import { computed } from 'vue'
import { buttonVariants } from '.'
import { cn } from '@/lib/utils'

const props = defineProps({
  variant: { type: String, default: 'default' },
  size: { type: String, default: 'default' },
  as: { type: String, default: 'button' },
  asChild: { type: Boolean, default: false },
  class: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
})

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props
  return delegated
})
</script>

<template>
  <component :is="as" :class="cn(buttonVariants({ variant, size }), props.class)" :disabled="disabled">
    <slot />
  </component>
</template>
