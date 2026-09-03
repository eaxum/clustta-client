<template>
  <button
    type="button"
    class="titlebar-status-pill"
    v-stop-propagation
    v-tooltip="tooltip"
    @click="$emit('click')"
  >
    <img class="small-icons titlebar-status-pill-icon" :src="icon" alt="">
    <span class="titlebar-status-pill-label">{{ label }}</span>
  </button>
</template>

<script setup>
defineProps({
  icon: { type: String, required: true },
  label: { type: String, required: true },
  tooltip: { type: String, default: '' },
});

defineEmits(['click']);
</script>

<style scoped>
.titlebar-status-pill {
  --wails-draggable: no-drag;
  display: flex;
  align-items: center;
  gap: 6px;
  max-width: 240px;
  height: 22px;
  margin-right: 6px;
  padding: 3px 10px 3px 5px;
  border: 1px solid rgba(255, 255, 255, 0.35);
  border-radius: var(--large-radius);
  color: white;
  background-color: transparent;
  font: inherit;
  font-size: 11px;
  cursor: pointer;
  transition: opacity 0.15s ease;
}

.titlebar-status-pill:hover {
  opacity: 0.8;
}

.titlebar-status-pill-icon {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.titlebar-status-pill-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
