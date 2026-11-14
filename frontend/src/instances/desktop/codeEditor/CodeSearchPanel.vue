<template>
  <div class="code-search-panel">
    <input
      v-model="query"
      class="code-search-input"
      type="search"
      placeholder="Quick open files"
      @input="queueSearch"
    />
  </div>
</template>

<script setup>
import { onBeforeUnmount, ref } from 'vue';
import { useCodeEditorStore } from '@/stores/codeEditor';
import { useProjectStore } from '@/stores/projects';

const codeEditor = useCodeEditorStore();
const projectStore = useProjectStore();
const query = ref('');
let searchTimer = null;

const queueSearch = () => {
  window.clearTimeout(searchTimer);
  searchTimer = window.setTimeout(() => {
    codeEditor.searchFiles(projectStore.activeProject.working_directory, query.value);
  }, 180);
};

onBeforeUnmount(() => window.clearTimeout(searchTimer));
</script>

<style scoped>
.code-search-panel {
  padding: .5rem;
}

.code-search-input {
  box-sizing: border-box;
  width: 100%;
  padding: .45rem .65rem;
  border: var(--transparent-line);
  border-radius: var(--large-radius);
  outline: none;
  background-color: var(--surface-1);
  color: var(--text);
  font-family: var(--font-family-interface);
}

.code-search-input:focus {
  outline: var(--solid-line);
}
</style>
