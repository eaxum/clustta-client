<template>
  <span v-if="visible && exportable" class="export-drag-handle"
    :draggable="enabled" v-tooltip="$t('blocks.dragIntoOtherApp')"
    :aria-label="$t('blocks.dragIntoOtherApp')" :aria-disabled="!enabled"
    @mousedown.stop @mouseup.stop @pointerdown.stop @click.stop @dblclick.stop
    @contextmenu.stop.prevent @dragstart.stop="startExportDrag">
    <img class="small-icons" :src="iconStore.getAppIcon('grip-vertical')" draggable="false" alt="">
  </span>
</template>

<script setup>
import { toRef } from 'vue';
import { useExportDrag } from '@/composables/useExportDrag';
import { useIconStore } from '@/stores/icons';

const props = defineProps({
  assets: { type: Array, required: true },
  exportable: { type: Boolean, default: false },
});
const iconStore = useIconStore();
const { visible, enabled, startExportDrag } = useExportDrag(toRef(props, 'assets'), toRef(props, 'exportable'));
</script>

<style scoped>
.export-drag-handle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  opacity: 0.5;
  cursor: grab;
  user-select: none;
}
.export-drag-handle:active { cursor: grabbing; }
</style>
