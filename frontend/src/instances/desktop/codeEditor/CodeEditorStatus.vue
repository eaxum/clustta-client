<template>
  <div v-if="document" class="code-editor-status">
    <span>{{ document.classification.language_id }}</span>
    <span>{{ document.classification.encoding }}</span>
    <span>{{ document.classification.line_ending }}</span>
    <span>{{ document.dirty ? 'Modified' : 'Saved' }}</span>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useCodeEditorStore } from '@/stores/codeEditor';

const codeEditor = useCodeEditorStore();
const document = computed(() => codeEditor.activeDocument);
</script>

<style scoped>
.code-editor-status {
  display: flex;
  gap: 1rem;
  min-width: max-content;
  color: var(--text-soft);
  text-transform: uppercase;
}
</style>
