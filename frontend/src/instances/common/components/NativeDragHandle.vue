<template>
  <span v-if="visible" class="native-drag-handle" :class="{ 'native-drag-disabled': !enabled }"
    :draggable="enabled" :title="enabled ? $t('blocks.dragFilesOut') : $t('blocks.dragFilesUnavailable')"
    :aria-label="$t('blocks.dragFilesOut')" :aria-disabled="!enabled"
    @mousedown.stop @mouseup.stop @pointerdown.stop @click.stop @dblclick.stop
    @contextmenu.stop.prevent @dragstart.stop="startDrag">
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true">
      <path d="M13 5h6v6M19 5l-9 9M10 5H5v14h14v-5" />
    </svg>
  </span>
</template>

<script setup>
import { toRef } from 'vue';
import { useNativeDrag } from '@/composables/useNativeDrag';

const props = defineProps({ asset: { type: Object, required: true } });
const { visible, enabled, startDrag } = useNativeDrag(toRef(props, 'asset'));
</script>

<style scoped>
.native-drag-handle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 6px;
  border-radius: 6px;
  cursor: grab;
  flex-shrink: 0;
  user-select: none;
}
.native-drag-handle:hover { background: rgba(128, 128, 128, 0.15); }
.native-drag-handle:active { cursor: grabbing; }
.native-drag-disabled { opacity: 0.35; cursor: default; }
</style>
