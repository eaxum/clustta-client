<template>
  <span v-if="keys.length" class="shortcut-badges" aria-hidden="true">
    <kbd v-for="key in keys" :key="key" class="shortcut-key">{{ key }}</kbd>
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
