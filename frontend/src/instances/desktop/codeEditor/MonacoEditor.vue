<template>
  <div ref="editorRoot" class="monaco-editor-root" @contextmenu.prevent="showEditorMenu"></div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
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
let themeObserver = null;

self.MonacoEnvironment = {
  getWorker() {
    return new EditorWorker();
  },
};

const saveViewState = (path) => {
  if (!path || !editor || !model) return;
  const document = codeEditor.documents.find((item) => item.path === path);
  if (document) document.viewState = editor.saveViewState();
};

const createModel = (path, previousPath) => {
  const document = codeEditor.activeDocument;
  if (!document || !editor) return;
  saveViewState(previousPath);
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
  if (document.viewState) editor.restoreViewState(document.viewState);
  changeSubscription = model.onDidChangeContent(() => {
    codeEditor.updateContent(model.getValue());
  });
};

const cssColorToHex = (value) => {
  const canvas = document.createElement('canvas');
  canvas.width = 1;
  canvas.height = 1;
  const context = canvas.getContext('2d');
  context.clearRect(0, 0, 1, 1);
  context.fillStyle = value;
  context.fillRect(0, 0, 1, 1);
  const [red, green, blue, alpha] = context.getImageData(0, 0, 1, 1).data;
  const channels = [red, green, blue, alpha]
    .map((channel) => channel.toString(16).padStart(2, '0'));
  return `#${channels.join('')}`;
};

const getThemeColor = (name) => {
  const value = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  return cssColorToHex(value);
};

const applyClusttaTheme = () => {
  if (!editor) return;
  const mode = document.documentElement.dataset.mode;
  monaco.editor.defineTheme('clustta', {
    base: mode === 'dark' ? 'vs-dark' : 'vs',
    inherit: true,
    rules: [],
    colors: {
      'editor.background': getThemeColor('--surface-1'),
      'editor.foreground': getThemeColor('--text'),
      'editorLineNumber.foreground': getThemeColor('--text-muted'),
      'editorLineNumber.activeForeground': getThemeColor('--text'),
      'editor.selectionBackground': getThemeColor('--selected'),
      'editor.inactiveSelectionBackground': getThemeColor('--selected-soft'),
      'editor.lineHighlightBackground': getThemeColor('--hover'),
      'editorCursor.foreground': getThemeColor('--accent'),
      'editorGutter.background': getThemeColor('--surface-1'),
      'minimap.background': getThemeColor('--surface-1'),
      'scrollbarSlider.background': getThemeColor('--surface-4'),
      'scrollbarSlider.hoverBackground': getThemeColor('--surface-5'),
    },
  });
  monaco.editor.setTheme('clustta');
};

const syncDocumentContent = () => {
  const document = codeEditor.activeDocument;
  if (!document || !model || model.uri.fsPath !== monaco.Uri.file(document.path).fsPath) return;
  if (model.getValue() === document.content) return;
  const viewState = editor.saveViewState();
  model.setValue(document.content);
  if (viewState) editor.restoreViewState(viewState);
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

watch(() => codeEditor.activePath, createModel);
watch(
  () => [codeEditor.activeDocument?.content, codeEditor.activeDocument?.content_identity],
  syncDocumentContent,
);

onMounted(() => {
  editor = monaco.editor.create(editorRoot.value, {
    automaticLayout: true,
    contextmenu: false,
    disableMonospaceOptimizations: true,
    fontFamily: 'JetBrains Mono, monospace',
    fontLigatures: false,
    minimap: { enabled: false },
    padding: { top: 12 },
    scrollBeyondLastLine: false,
    theme: 'clustta',
  });
  applyClusttaTheme();
  themeObserver = new MutationObserver(() => nextTick(applyClusttaTheme));
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['data-mode', 'data-tint', 'style'],
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
  document.fonts.ready.then(() => {
    monaco.editor.remeasureFonts();
    editor?.layout();
  });
  window.addEventListener('code-editor-command', runCommand);
});

onBeforeUnmount(() => {
  window.removeEventListener('code-editor-command', runCommand);
  saveViewState(codeEditor.activePath);
  themeObserver?.disconnect();
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
