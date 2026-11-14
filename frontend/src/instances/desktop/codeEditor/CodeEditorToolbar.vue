<template>
  <div class="code-editor-toolbar">
    <ActionButton
      :icon="panes.showDetailsPane ? getAppIcon('collapse-right') : getAppIcon('collapse-left')"
      :buttonFunction="toggleDetailsPane"
      v-tooltip="panes.showDetailsPane ? 'Hide details' : 'Show details'"
    />
    <div class="toolbar-divider"></div>
    <ActionButton :label="'File'" :showLabel="true" :showIcon="false" @click="showFileMenu" />
    <ActionButton :label="'Edit'" :showLabel="true" :showIcon="false" @click="showEditMenu" />
    <ActionButton :label="'View'" :showLabel="true" :showIcon="false" @click="showViewMenu" />
    <div class="toolbar-spacer"></div>
    <ActionButton
      :icon="getAppIcon('circle-check')"
      :isDisabled="!codeEditor.activeDocument?.dirty"
      :buttonFunction="saveActive"
      v-tooltip="'Save'"
    />
  </div>
</template>

<script setup>
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import { useCodeEditorStore } from '@/stores/codeEditor';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';

const codeEditor = useCodeEditorStore();
const iconStore = useIconStore();
const menu = useMenu();
const panes = usePaneStore();
const stage = useStageStore();

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

const showViewMenu = (event) => menu.showCodeEditorMenu(event, [
  { id: 'explorer', label: codeEditor.explorerVisible ? 'Hide Explorer' : 'Show Explorer', icon: 'collapse-left', action: () => { codeEditor.explorerVisible = !codeEditor.explorerVisible; } },
  { id: 'browser', label: 'Return to Browser', icon: 'four-squares', action: () => stage.setStageVisibility('browser', true) },
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
