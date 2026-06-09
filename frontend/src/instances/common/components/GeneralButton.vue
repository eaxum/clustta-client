<template>
  <Button :variant="colored ? 'default' : 'outline'" :disabled="!isActive || loading" :class="['general-button', { 'w-full': fullWidth, 'rounded-full aspect-square p-0': loading }]"
    @click="buttonFunction" v-stop-propagation>
    <template v-if="!loading">{{ label }}</template>
    <img v-else class="h-5 w-5 animate-spin" :class="{ 'invert': !colored }" src="/icons/loading.svg" />
  </Button>
</template>

<script setup>
import { Button } from '@/components/ui/button';

const props = defineProps({
  icon: String,
  buttonFunction: { type: Function, default: () => {} },
  label: { type: String, default: 'Button' },
  colored: { type: Boolean, default: true },
  loading: { type: Boolean, default: false },
  isActive: { type: Boolean, default: true },
  showLabel: { type: Boolean, default: false },
  fullWidth: { type: Boolean, default: true },
});
</script>

<style scoped>
.general-button {
  min-width: 100px;
}
</style>
  
  
  

