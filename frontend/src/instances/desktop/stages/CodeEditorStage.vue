<template>
  <div class="code-editor-stage">
    <CodeExplorer v-if="codeEditor.explorerVisible" />
    <CodeEditorHost />
    <DetailsPane v-if="panes.showDetailsPane" />
  </div>
</template>

<script setup>
import CodeEditorHost from '@/instances/desktop/codeEditor/CodeEditorHost.vue';
import CodeExplorer from '@/instances/desktop/codeEditor/CodeExplorer.vue';
import DetailsPane from '@/instances/desktop/components/DetailsPane.vue';
import { useCodeEditorStore } from '@/stores/codeEditor';
import { usePaneStore } from '@/stores/panes';
import { useProjectStore } from '@/stores/projects';
import { useFsWatch } from '@/composables/useFsWatch';

const codeEditor = useCodeEditorStore();
const panes = usePaneStore();
const projectStore = useProjectStore();

useFsWatch(
  () => projectStore.activeProject?.working_directory,
  (changedPath) => codeEditor.handleExternalChange(changedPath),
  { debounce: 150 },
);
</script>

<style scoped>
.code-editor-stage {
  display: flex;
  width: 100%;
  height: 100%;
  overflow: hidden;
  border-radius: var(--large-radius);
  box-sizing: border-box;
  gap: 5px;
}
</style>
