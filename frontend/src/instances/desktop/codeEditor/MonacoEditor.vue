<template>
  <div ref="editorRoot" class="monaco-editor-root" @contextmenu.prevent="showEditorMenu"></div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
import * as monaco from 'monaco-editor/editor/editor.api';
import EditorWorker from 'monaco-editor/editor/editor.worker?worker';
import 'monaco-editor/languages/definitions/bat/register';
import 'monaco-editor/languages/definitions/cpp/register';
import 'monaco-editor/languages/definitions/csharp/register';
import 'monaco-editor/languages/definitions/css/register';
import 'monaco-editor/languages/definitions/dockerfile/register';
import 'monaco-editor/languages/definitions/go/register';
import 'monaco-editor/languages/definitions/html/register';
import 'monaco-editor/languages/definitions/ini/register';
import 'monaco-editor/languages/definitions/java/register';
import 'monaco-editor/languages/definitions/javascript/register';
import 'monaco-editor/languages/definitions/less/register';
import 'monaco-editor/languages/definitions/lua/register';
import 'monaco-editor/languages/definitions/markdown/register';
import 'monaco-editor/languages/definitions/powershell/register';
import 'monaco-editor/languages/definitions/python/register';
import 'monaco-editor/languages/definitions/rust/register';
import 'monaco-editor/languages/definitions/scss/register';
import 'monaco-editor/languages/definitions/shell/register';
import 'monaco-editor/languages/definitions/typescript/register';
import 'monaco-editor/languages/definitions/xml/register';
import 'monaco-editor/languages/definitions/yaml/register';
import { useCodeEditorStore } from '@/stores/codeEditor';
import { useMenu } from '@/stores/menu';

const codeEditor = useCodeEditorStore();
const menu = useMenu();
const editorRoot = ref(null);
let editor = null;
let model = null;
let changeSubscription = null;

self.MonacoEnvironment = {
  getWorker() {
    return new EditorWorker();
  },
};

const createModel = () => {
  const document = codeEditor.activeDocument;
  if (!document || !editor) return;
  changeSubscription?.dispose();
  model?.dispose();
  model = monaco.editor.createModel(
    document.content,
    document.classification.language_id || 'plaintext',
    monaco.Uri.file(document.path),
  );
  editor.setModel(model);
  if (document.classification.line_ending === 'CRLF') {
    model.setEOL(monaco.editor.EndOfLineSequence.CRLF);
  }
  editor.updateOptions({ readOnly: !document.classification.editable });
  changeSubscription = model.onDidChangeContent(() => {
    codeEditor.updateContent(model.getValue());
  });
};

const runCommand = (event) => {
  const commands = {
    undo: 'undo',
    redo: 'redo',
    find: 'actions.find',
  };
  const command = commands[event.detail];
  if (command) editor?.trigger('clustta', command, null);
};

const showEditorMenu = (event) => menu.showCodeEditorMenu(event, [
  { id: 'cut', label: 'Cut', icon: 'scissors', action: () => editor?.trigger('clustta', 'editor.action.clipboardCutAction', null) },
  { id: 'copy', label: 'Copy', icon: 'copy', action: () => editor?.trigger('clustta', 'editor.action.clipboardCopyAction', null) },
  { id: 'paste', label: 'Paste', icon: 'clipboard', action: () => editor?.trigger('clustta', 'editor.action.clipboardPasteAction', null) },
  { id: 'divider-1', divider: true },
  { id: 'find', label: 'Find', icon: 'search', action: () => editor?.trigger('clustta', 'actions.find', null) },
]);

watch(() => [codeEditor.activePath, codeEditor.activeDocument?.content_identity], createModel);

onMounted(() => {
  editor = monaco.editor.create(editorRoot.value, {
    automaticLayout: true,
    contextmenu: false,
    fontFamily: 'JetBrains Mono, monospace',
    minimap: { enabled: true },
    padding: { top: 12 },
    scrollBeyondLastLine: false,
    theme: document.documentElement.dataset.theme === 'light' ? 'vs' : 'vs-dark',
  });
  editor.addAction({
    id: 'clustta-save-file',
    label: 'Save',
    keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS],
    run: () => codeEditor.saveDocument(),
  });
  editor.addAction({
    id: 'clustta-save-all',
    label: 'Save All',
    keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.KeyS],
    run: () => codeEditor.saveAll(),
  });
  createModel();
  window.addEventListener('code-editor-command', runCommand);
});

onBeforeUnmount(() => {
  window.removeEventListener('code-editor-command', runCommand);
  changeSubscription?.dispose();
  model?.dispose();
  editor?.dispose();
});
</script>

<style scoped>
.monaco-editor-root {
  width: 100%;
  height: 100%;
  overflow: hidden;
  background-color: var(--surface-1);
}
</style>
