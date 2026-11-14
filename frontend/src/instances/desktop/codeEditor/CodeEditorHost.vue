<template>
  <section class="code-editor-host" :class="{ 'code-editor-host-expanded' :!panes.showDetailsPane}">
    <CodeEditorToolbar />
    <ExternalFileChangeBanner />
    <MonacoEditor v-if="codeEditor.activeDocument" />
    <div v-else class="editor-empty">Open a text file from the browser or explorer.</div>
  </section>
</template>

<script setup>
import CodeEditorToolbar from '@/instances/desktop/codeEditor/CodeEditorToolbar.vue';
import MonacoEditor from '@/instances/desktop/codeEditor/MonacoEditor.vue';
import ExternalFileChangeBanner from '@/instances/desktop/codeEditor/ExternalFileChangeBanner.vue';
import { useCodeEditorStore } from '@/stores/codeEditor';
import { usePaneStore } from '@/stores/panes';

const codeEditor = useCodeEditorStore();
const panes = usePaneStore();
</script>

<style scoped>
.code-editor-host {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
  background-color: var(--surface-1);
  border-radius: var(--normal-radius);
}

.code-editor-host-expanded {
  border-top-right-radius: var(--very-large-radius);
}

.editor-empty {
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: center;
  color: var(--text-soft);
}
</style>
