<template>
  <span v-if="keys.length" class="shortcut-badges" aria-hidden="true">
    <kbd class="shortcut-label">{{ keys.join('+') }}</kbd>
  </span>
</template>

<script setup>
import { computed } from 'vue';
import { getShortcutKeys } from '@/lib/shortcuts';
import { usePlatformStore } from '@/stores/platform';

const platformStore = usePlatformStore();

const props = defineProps({
  shortcut: {
    type: [String, Array],
    required: true,
  },
});

const keys = computed(() => getShortcutKeys(props.shortcut, platformStore.isMac));
</script>

<style scoped>
.shortcut-label {
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--text-secondary, var(--text));
  font-family: inherit;
  font-size: 12px;
  font-weight: 400;
  line-height: 1rem;
  white-space: nowrap;
}
</style>
