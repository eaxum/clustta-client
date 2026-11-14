<template>
  <div class="code-search-panel">
    <SearchBar
      v-model="query"
      placeholder="Quick open files"
      @input="queueSearch"
      @clear="search"
    />
  </div>
</template>

<script setup>
import { onBeforeUnmount, ref } from 'vue';
import SearchBar from '@/instances/desktop/components/SearchBar.vue';
import { useCodeEditorStore } from '@/stores/codeEditor';
import { useProjectStore } from '@/stores/projects';

const codeEditor = useCodeEditorStore();
const projectStore = useProjectStore();
const query = ref('');
let searchTimer = null;

const search = () => {
  codeEditor.searchFiles(projectStore.activeProject.working_directory, query.value);
};

const queueSearch = () => {
  window.clearTimeout(searchTimer);
  searchTimer = window.setTimeout(search, 180);
};

onBeforeUnmount(() => window.clearTimeout(searchTimer));
</script>

<style scoped>
.code-search-panel {
  padding: .25rem;
  padding-bottom: .5rem;
}
</style>
