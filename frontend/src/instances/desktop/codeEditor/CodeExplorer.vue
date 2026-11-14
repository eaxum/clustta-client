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
        <img class="explorer-file-icon" :src="getCodeFileIcon(file)" alt="">
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
import { getCodeFileIcon } from '@/lib/codeFileIcons';

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
  padding: .35rem;
  background-color: var(--surface-2);
  border-radius: var(--normal-radius);
  color: var(--text);
  font-family: var(--font-family-interface);
}

.explorer-title {
  padding: .5rem .65rem .35rem;
  color: var(--text-muted);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.explorer-files {
  display: flex;
  flex-direction: column;
  gap: .15rem;
  overflow: auto;
}

.explorer-file {
  display: flex;
  width: 100%;
  align-items: center;
  gap: .5rem;
  padding: .42rem .6rem;
  border: 0;
  border-radius: var(--normal-radius);
  background: transparent;
  color: var(--text);
  cursor: pointer;
  font-family: inherit;
  font-size: var(--font-size-interface);
  text-align: left;
}

.explorer-file:hover,
.explorer-file.active {
  background-color: var(--selected-soft);
}

.explorer-file:focus-visible {
  outline: var(--solid-line);
  outline-offset: -1px;
}

.explorer-file-icon {
  flex: 0 0 auto;
  width: 17px;
  height: 17px;
}

.file-name {
  flex: 1;
  min-width: 0;
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
