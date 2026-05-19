<template>
  <div class="squash-preview-item" :class="{ 'squash-preview-first': index === 0 }" :style="{ animationDelay: index < 10 ? `${index * 0.04}s` : '0s' }">
    <div class="squash-preview-badge" :class="{ 'badge-first': index === 0 }">{{ label }}</div>

    <div class="squash-preview-content">
      <div class="squash-preview-name">
        {{ item.name }}{{ item.extension }}
      </div>
      <div class="squash-preview-meta">{{ formattedSize }}</div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed } from 'vue';

// props
const props = defineProps({
  index: { type: Number, required: true },
  item: { type: Object, required: true },
  label: { type: String, required: true },
});

// computed
// Formats the file size for display.
const formattedSize = computed(() => {
  const size = props.item.file_size || 0;
  if (size === 0) return '';
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
});
</script>

<style scoped>
.squash-preview-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0.75rem;
  border-radius: var(--large-radius);
  background-color: var(--surface-2);
  opacity: 0;
  animation: fadeIn 0.3s ease-in-out forwards;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.squash-preview-item:hover {
  background-color: var(--surface-3);
}

.squash-preview-badge {
  min-width: 36px;
  max-width: 180px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--small-radius);
  background-color: var(--surface-3);
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  flex-shrink: 0;
  padding: 0 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.badge-first {
  background-color: var(--selected);
  color: var(--text);
}

.squash-preview-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow: hidden;
  flex: 1;
}

.squash-preview-name {
  font-size: 13px;
  font-weight: 400;
  color: var(--text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.squash-preview-meta {
  font-size: 11px;
  color: var(--text-muted);
  opacity: 0.8;
}

.squash-preview-first {
  outline: var(--transparent-line);
  outline-offset: -1px;
}
</style>
