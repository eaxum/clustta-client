<template>
  <aside class="code-explorer">
    <CodeSearchPanel />
    <div class="explorer-title">Files</div>
    <div class="explorer-files app-scrollbar">
      <button
        v-for="file in codeEditor.searchResults"
        :key="file.path"
        class="explorer-file"
        :class="{ active: codeEditor.activePath === file.path }"
        type="button"
        @click="codeEditor.openDocument(file.path, null, true)"
        @dblclick="openPinned(file)"
      >
        <span class="file-name">{{ file.name }}</span>
        <span v-if="file.line" class="file-line">{{ file.line }}</span>
      </button>
    </div>
  </aside>
</template>

<script setup>
import { onMounted } from 'vue';
import CodeSearchPanel from '@/instances/desktop/codeEditor/CodeSearchPanel.vue';
import { useCodeEditorStore } from '@/stores/codeEditor';
import { useProjectStore } from '@/stores/projects';

const codeEditor = useCodeEditorStore();
const projectStore = useProjectStore();

const openPinned = async (file) => {
  await codeEditor.openDocument(file.path, null, false);
};

onMounted(() => {
  codeEditor.searchFiles(projectStore.activeProject.working_directory, '');
});
</script>

<style scoped>
.code-explorer {
  display: flex;
  flex: 0 0 240px;
  flex-direction: column;
  min-width: 180px;
  background-color: var(--surface-1);
  border-radius: var(--normal-radius);
  border-top-left-radius: var(--very-large-radius);
}

.explorer-title {
  padding: .35rem .75rem;
  color: var(--text-soft);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.explorer-files {
  overflow: auto;
}

.explorer-file {
  display: flex;
  width: 100%;
  padding: .38rem .75rem;
  border: 0;
  background: transparent;
  color: var(--text);
  cursor: pointer;
  text-align: left;
}

.explorer-file:hover,
.explorer-file.active {
  background-color: var(--selected-soft);
}

.file-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-line {
  margin-left: auto;
  color: var(--text-soft);
  font-size: 11px;
}
</style>
