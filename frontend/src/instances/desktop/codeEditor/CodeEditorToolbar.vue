<template>
  <div class="code-editor-toolbar">
    <ActionButton
      :icon="getAppIcon(fileState.icon)"
      :isInactive="true"
      :useAlert="fileState.alert"
      :useDanger="fileState.danger"
      :noFilter="fileState.noFilter"
      v-tooltip="fileState.label"
    />
    <div class="toolbar-divider"></div>
    <ActionButton :label="'File'" :showLabel="true" :showIcon="false" @click="showFileMenu" />
    <ActionButton :label="'Edit'" :showLabel="true" :showIcon="false" @click="showEditMenu" />
    <div class="toolbar-spacer"></div>
    <ActionButton
      :icon="panes.showDetailsPane ? getAppIcon('collapse-right') : getAppIcon('collapse-left')"
      :buttonFunction="toggleDetailsPane"
      v-tooltip="panes.showDetailsPane ? 'Hide details' : 'Show details'"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import { useCodeEditorStore } from '@/stores/codeEditor';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { usePaneStore } from '@/stores/panes';

const codeEditor = useCodeEditorStore();
const iconStore = useIconStore();
const menu = useMenu();
const panes = usePaneStore();

const fileState = computed(() => {
  const asset = codeEditor.activeDocument?.asset;
  if (!asset) return { icon: 'checkpoint-stone', label: 'Checkpoint state unavailable' };
  if (asset.type === 'untracked_asset') {
    return { icon: 'plus-stone', label: 'Not checkpointed', danger: true, noFilter: true };
  }
  const states = {
    fetchable: { icon: 'fetch', label: 'Checkpoint available to fetch' },
    missing: { icon: 'alert', label: 'File missing', danger: true, noFilter: true },
    modified: { icon: 'plus-stone', label: 'File has uncheckpointed changes', alert: true, noFilter: true },
    normal: { icon: 'circle-check-go', label: 'Latest checkpoint', noFilter: true },
    outdated: { icon: 'circle-check', label: 'File is not at the latest checkpoint', alert: true, noFilter: true },
  };
  return states[asset.file_status] || { icon: 'checkpoint-stone', label: 'Checkpoint state unavailable' };
});

const getAppIcon = (name) => iconStore.getAppIcon(name);
const toggleDetailsPane = () => { panes.showDetailsPane = !panes.showDetailsPane; };
const saveActive = async () => {
  try {
    await codeEditor.saveDocument();
  } catch (error) {
    console.error(error);
  }
};

const showFileMenu = (event) => menu.showCodeEditorMenu(event, [
  { id: 'save', label: 'Save', icon: 'circle-check', disabled: !codeEditor.activeDocument?.dirty, action: saveActive },
  { id: 'save-all', label: 'Save All', icon: 'circle-check', action: codeEditor.saveAll },
  { id: 'divider-1', divider: true },
  { id: 'close', label: 'Close', icon: 'close', disabled: codeEditor.activeDocument?.dirty, action: () => codeEditor.closeDocument(codeEditor.activePath) },
], { anchor: true, position: 'bottom' });

const showEditMenu = (event) => menu.showCodeEditorMenu(event, [
  { id: 'undo', label: 'Undo', icon: 'undo', action: () => window.dispatchEvent(new CustomEvent('code-editor-command', { detail: 'undo' })) },
  { id: 'redo', label: 'Redo', icon: 'redo', action: () => window.dispatchEvent(new CustomEvent('code-editor-command', { detail: 'redo' })) },
  { id: 'divider-1', divider: true },
  { id: 'find', label: 'Find', icon: 'search', action: () => window.dispatchEvent(new CustomEvent('code-editor-command', { detail: 'find' })) },
], { anchor: true, position: 'bottom' });

</script>

<style scoped>
.code-editor-toolbar {
  display: flex;
  align-items: center;
  min-height: 38px;
  padding: 0 .45rem;
  border-bottom: var(--transparent-line);
  background-color: var(--surface-2);
}

.toolbar-divider {
  width: 1px;
  height: 18px;
  margin: 0 .35rem;
  background-color: var(--surface-4);
}

.toolbar-spacer {
  flex: 1;
}
</style>
