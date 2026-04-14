<template>
  <div class="general-pane-root">
    <div class="console-container">
      <div class="console-messages" ref="messagesContainer">
        <template v-for="(message, index) in messages" :key="index">
          <div v-if="message.type === 'user'" class="msg-user">
            <div class="msg-user-bubble">
              <div v-if="parseUserContext(message.content).context" class="msg-context-tag">{{ parseUserContext(message.content).context }}</div>
              {{ parseUserContext(message.content).body }}
            </div>
          </div>

          <div v-else-if="message.type === 'tool-group'" class="msg-tool">
            <img class="msg-tool-icon small-icons" :src="getAppIcon(getToolIcon(message.toolName))">
            <span class="msg-tool-label">{{ formatToolLabel(message.toolName, message.count) }}</span>
          </div>

          <div v-else-if="message.type === 'status'" class="msg-status">
            <div class="msg-status-dot"></div>
            <span>{{ message.content }}</span>
          </div>

          <div v-else-if="message.type === 'error'" class="msg-error">
            <div class="msg-error-text" v-html="formatContent(message.content)"></div>
          </div>

          <div v-else class="msg-assistant">
            <div class="msg-assistant-text" v-html="formatContent(message.content)"></div>
          </div>
        </template>

        <div v-if="!messages.length && isApiKeyConfigured" class="console-empty">
          <div class="empty-text">{{ emptyStateTitle }}</div>
          <div class="empty-subtext">{{ emptyStateSubtext }}</div>
        </div>

        <div v-if="!messages.length && !isApiKeyConfigured" class="console-empty">
          <div class="empty-text">{{ $t('panes.setupAiAgent') }}</div>
          <div class="empty-subtext">{{ $t('panes.configureLlmPre') }} <a class="console-link" @click="openAdvancedSettings">{{ $t('common.settings') }}</a> {{ $t('panes.configureLlmPost') }}</div>
        </div>
      </div>

      <div class="console-input-container">
        <div class="console-input-wrapper">
          <div v-if="attachmentPath" class="console-attachment-row">
            <Chip :icon="getAppIcon('paper-clip')" :label="attachmentName" :onRemove="removeAttachment" />
          </div>

          <textarea ref="textareaRef" v-model="currentMessage" class="console-input" type="text" :placeholder="inputPlaceholder"
            spellcheck="false" @input="handleInput" @keydown.enter.exact.prevent="sendMessage" :disabled="isProcessing" />

          <div class="console-toolbar">
            <div class="console-toolbar-left">
              <ActionButton :icon="getAppIcon('paper-clip')" :showLabel="false" v-tooltip="$t('panes.attachFile')" :buttonFunction="selectAttachment" />

              <PaneHeaderTabs :iconsOnly="false" :useSelected="true" :selectedTab="selectedConsoleTab"
                :dataTypes="consoleTabs" @filter="handleConsoleTabClick" />
            </div>

            <div class="console-toolbar-right">
              <ActionButton :icon="getAppIcon('broom')" :showLabel="false" :isDisabled="!messages.length || isProcessing"
                v-tooltip="$t('panes.clearChat')" :buttonFunction="clearChat" />

              <ActionButton :icon="getAppIcon('send')" :showLabel="false" :isDisabled="!currentMessage.trim() || isProcessing"
                v-tooltip="$t('panes.sendMessage')" :buttonFunction="sendMessage" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, nextTick, onActivated, onMounted, onUnmounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Events } from '@wailsio/runtime';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import Chip from '@/instances/common/components/Chip.vue';
import PaneHeaderTabs from '@/instances/common/components/PaneHeaderTabs.vue';

// services
import { AgentService, DialogService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useIconStore } from '@/stores/icons';
import { useProjectStore } from '@/stores/projects';
import { useSettingsStore } from '@/stores/settings';
import { useStageStore } from '@/stores/stages';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const iconStore = useIconStore();
const projectStore = useProjectStore();
const settings = useSettingsStore();
const stage = useStageStore();

const { t } = useI18n();

// refs
const attachmentPath = ref('');
const consoleTabs = ref([
  { name: "Agent", nameKey: "panes.agent", icon: "brain" },
  { name: "Bash", nameKey: "panes.bash", icon: "console" }
]);
const currentMessage = ref('');
const isApiKeyConfigured = ref(false);
const isProcessing = ref(false);
const messages = ref([]);
const messagesContainer = ref(null);
const selectedConsoleTab = ref('Agent');
const textareaRef = ref(null);

// computed properties
const attachmentName = computed(() => {
  if (!attachmentPath.value) return '';
  return attachmentPath.value.split(/[/\\]/).pop();
});

const emptyStateSubtext = computed(() => {
  if (selectedConsoleTab.value === 'Bash') return t('panes.executeTerminalCommands', { itemType: itemType.value });
  return t('panes.performOperation', { itemType: itemType.value });
});

const emptyStateTitle = computed(() => {
  if (selectedConsoleTab.value === 'Bash') return t('panes.terminalReady');
  return t('panes.startConversation');
});

const inputPlaceholder = computed(() => {
  if (isProcessing.value) return t('panes.agentWorking');
  if (!isApiKeyConfigured.value) return t('panes.configureLlmInSettings');
  return t('panes.askAnything');
});

const itemType = computed(() => {
  if (assetStore.selectedAsset) return 'asset';
  if (collectionStore.selectedCollection) return 'collection';
  if (!stage.markedItems.length) return 'project';
  return 'item';
});

// constants
const toolIconMap = {
  add_dependency: 'link',
  add_ignore_pattern: 'file-watch',
  add_tag_to_asset: 'tag',
  assign_asset: 'person-plus',
  batch_add_tags: 'tag',
  batch_create_assets: 'file-plus',
  batch_create_collections: 'folder-plus',
  blender_export: 'console',
  blender_link: 'link',
  blender_render: 'console',
  blender_run_python: 'console',
  blender_run_script: 'console',
  blender_set_settings: 'cog',
  bulk_assign: 'person-plus',
  bulk_change_status: 'workflow-arrow',
  bulk_delete_assets: 'trash',
  change_asset_status: 'workflow-arrow',
  create_asset: 'file-plus',
  create_asset_type: 'brush',
  create_collection: 'folder-plus',
  create_collection_type: 'brush',
  create_tag: 'tag',
  delete_asset: 'trash',
  delete_asset_type: 'trash',
  delete_collection: 'trash',
  delete_collection_type: 'trash',
  generate_script: 'console',
  get_asset_details: 'file-search',
  get_asset_tags: 'tag',
  get_my_permissions: 'scale',
  get_project_summary: 'four-squares',
  get_user_activity: 'person',
  list_assets_in_collection: 'file',
  list_checkpoints: 'history',
  list_collection_types: 'folder',
  list_collections: 'folder',
  list_dependencies: 'link',
  list_dependency_types: 'link',
  list_ignore_patterns: 'eye-off',
  list_statuses: 'workflow-arrow',
  list_tags: 'tag',
  list_task_types: 'brush',
  list_templates: 'file',
  list_users: 'person',
  move_assets: 'folder-arrow-in',
  open_in_dcc: 'external-link',
  random_assign: 'person-plus',
  remove_dependency: 'link',
  remove_ignore_pattern: 'eye-off',
  remove_tag_from_asset: 'tag',
  remove_user: 'person-minus',
  rename_asset: 'edit',
  rename_collection: 'edit',
  run_terminal_command: 'console',
  search_assets: 'file-search',
  search_knowledge: 'brain',
  setup_project_types: 'brush',
  unassign_all_assets: 'person-minus',
  unassign_asset: 'person-minus',
};

// methods

// Checks agent API key status on mount.
const checkApiKeyStatus = async () => {
  try {
    const status = await AgentService.GetAPIKeyStatus();
    isApiKeyConfigured.value = status.configured;
  } catch {
    isApiKeyConfigured.value = false;
  }
};

// Clears the chat history for the current project.
const clearChat = async () => {
  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath) return;
  try {
    await AgentService.ClearChatSession(projectPath);
    messages.value = [];
  } catch { /* ignore */ }
};

// Escapes HTML to prevent injection when rendering message content.
const escapeHtml = (text) => {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
};

// Formats a tool name and count into a readable label.
const formatToolLabel = (toolName, count) => {
  const label = toolName.replace(/_/g, ' ');
  if (count > 1) return `${label} (${count})`;
  return label;
};

// Formats message content with basic markdown-like rendering.
const formatContent = (text) => {
  if (!text) return '';
  let escaped = escapeHtml(text);
  // Format ISO dates (e.g. 2025-04-17T23:36:15Z or 2025-04-17 at 23:36:15Z)
  escaped = escaped.replace(/(\d{4}-\d{2}-\d{2})[T ](?:at )?((\d{2}):(\d{2})(?::\d{2})?Z?)/g, (match, datePart, timePart, hours, minutes) => {
    try {
      const date = new Date(datePart + 'T' + timePart.replace('at ', '') + (timePart.endsWith('Z') ? '' : 'Z'));
      if (isNaN(date)) return match;
      return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' }) + ' at ' + date.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' });
    } catch { return match; }
  });
  // Code blocks
  escaped = escaped.replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code class="code-block">$2</code></pre>');
  // Inline code
  escaped = escaped.replace(/`([^`]+)`/g, '<code class="inline-code">$1</code>');
  // Bold
  escaped = escaped.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
  // Line breaks
  escaped = escaped.replace(/\n/g, '<br>');
  return escaped;
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Extracts [Context: ...] prefix from a user message into separate display tag and body.
const parseUserContext = (content) => {
  if (!content) return { context: '', body: content };
  const match = content.match(/^\[Context:\s*(.+?)\]\n?/);
  if (match) {
    // Strip parenthetical metadata (IDs, types) for display — those are for the agent, not the user
    const display = match[1].replace(/\s*\([^)]*\)/g, '').replace(/"/g, '');
    return { context: display, body: content.slice(match[0].length) };
  }
  return { context: '', body: content };
};

// Opens settings page directly to the Advanced tab.
const openAdvancedSettings = () => {
  settings.pendingTab = 'advanced';
  stage.setStageVisibility('settings', true);
};

// Returns the icon name for a given tool function name.
const getToolIcon = (toolName) => toolIconMap[toolName] || 'cog';

// Switches the active console tab mode.
const handleConsoleTabClick = (tabName) => {
  selectedConsoleTab.value = tabName;
};

// Auto-resizes the textarea based on content.
const handleInput = () => {
  if (!textareaRef.value) return;
  textareaRef.value.style.height = 'auto';
  textareaRef.value.style.height = textareaRef.value.scrollHeight + 'px';
};

// Loads persisted chat history from the backend and renders it.
const loadChatHistory = async () => {
  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath) return;
  try {
    const history = await AgentService.GetChatHistory(projectPath);
    if (history && history.length) {
      messages.value = history.map((msg, i) => ({ id: i + 1, ...msg }));
      scrollToBottom();
    }
  } catch { /* no history available */ }
};

// Removes the attached file.
const removeAttachment = () => {
  attachmentPath.value = '';
};

// Scrolls the messages container to the bottom.
const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
  });
};

// Opens a file dialog to select an attachment.
const selectAttachment = async () => {
  try {
    const path = await DialogService.SelectFileDialog('Select file to attach', '');
    if (path) attachmentPath.value = path;
  } catch { /* user cancelled */ }
};

// Builds a selection context string to inform the agent of what the user is currently viewing.
const buildSelectionContext = () => {
  const parts = [];

  if (collectionStore.selectedCollection) {
    const c = collectionStore.selectedCollection;
    parts.push(`Viewing collection: "${c.name}" (ID: ${c.id}, type: ${c.collection_type_name || 'default'})`);
  }

  if (stage.selectedItems.length > 1) {
    const items = stage.selectedItems.map(i => `"${i.name}" (ID: ${i.id}, type: ${i.type})`).join(', ');
    parts.push(`Selected items: ${items}`);
  } else if (stage.selectedItem) {
    const i = stage.selectedItem;
    const details = [`ID: ${i.id}`, `type: ${i.type}`];
    if (i.extension) details.push(`extension: ${i.extension}`);
    if (i.asset_type_name) details.push(`asset type: ${i.asset_type_name}`);
    if (i.status_short_name) details.push(`status: ${i.status_short_name}`);
    if (i.assignee_name) details.push(`assignee: ${i.assignee_name}`);
    parts.push(`Selected item: "${i.name}" (${details.join(', ')})`);
  }

  if (stage.activeStage) {
    parts.push(`Active view: ${stage.activeStage}`);
  }

  if (!parts.length) return '';
  return `[Context: ${parts.join(' | ')}]\n`;
};

// Sends the current message to the agent backend.
const sendMessage = async () => {
  if (!currentMessage.value.trim() || isProcessing.value) return;
  if (!isApiKeyConfigured.value) return;

  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath) {
    addMessage('error', 'No project is currently open. Please open a project first.');
    return;
  }

  addMessage('user', currentMessage.value.trim());
  const context = buildSelectionContext();
  const messageContent = context + currentMessage.value.trim();
  currentMessage.value = '';
  if (textareaRef.value) textareaRef.value.style.height = 'auto';
  isProcessing.value = true;

  try {
    await AgentService.SendMessage(projectPath, messageContent, attachmentPath.value);
    attachmentPath.value = '';
  } catch (err) {
    addMessage('error', `${err}`);
    isProcessing.value = false;
  }
};

// Adds a message to the messages list.
const addMessage = (type, content) => {
  messages.value.push({ id: messages.value.length + 1, type, content });
  scrollToBottom();
};

// --- Wails event handlers ---

// Handles agent status updates (e.g., "Thinking...").
const onAgentStatus = (event) => {
  const lastMsg = messages.value[messages.value.length - 1];
  if (lastMsg && lastMsg.type === 'status') {
    lastMsg.content = event.data;
  } else {
    messages.value.push({ id: messages.value.length + 1, type: 'status', content: event.data });
  }
  scrollToBottom();
};

// Handles tool execution start events, grouping consecutive calls of the same tool.
const onAgentToolStart = (event) => {
  const data = event.data;
  const toolName = data.tool;
  if (messages.value.length && messages.value[messages.value.length - 1].type === 'status') {
    messages.value.pop();
  }
  const lastMsg = messages.value[messages.value.length - 1];
  if (lastMsg && lastMsg.type === 'tool-group' && lastMsg.toolName === toolName) {
    lastMsg.count++;
  } else {
    messages.value.push({ id: messages.value.length + 1, type: 'tool-group', toolName, count: 1 });
  }
  scrollToBottom();
};

// Handles the final agent text response.
const onAgentResponse = (event) => {
  // Remove any trailing status message
  if (messages.value.length && messages.value[messages.value.length - 1].type === 'status') {
    messages.value.pop();
  }
  addMessage('assistant', event.data);
};

// Handles agent errors.
const onAgentError = (event) => {
  if (messages.value.length && messages.value[messages.value.length - 1].type === 'status') {
    messages.value.pop();
  }
  addMessage('error', event.data);
  isProcessing.value = false;
};

// Handles agent completion signal.
const onAgentDone = () => {
  if (messages.value.length && messages.value[messages.value.length - 1].type === 'status') {
    messages.value.pop();
  }
  isProcessing.value = false;
  emitter.emit('refresh-browser');
};

// Syncs the ignore list from the agent's DB update to the in-memory project store.
const onIgnoreListUpdated = (event) => {
  if (projectStore.activeProject && event.data) {
    projectStore.activeProject.ignore_list = event.data;
  }
};

// watchers
watch(() => projectStore.activeProject, async () => {
  messages.value = [];
  await checkApiKeyStatus();
  await loadChatHistory();
});

// lifecycle hooks
onActivated(async () => {
  await checkApiKeyStatus();
});

onMounted(async () => {
  await checkApiKeyStatus();
  await loadChatHistory();

  Events.On('agent-status', onAgentStatus);
  Events.On('agent-tool-start', onAgentToolStart);
  Events.On('agent-tool-result', () => {}); // silently consumed
  Events.On('agent-response', onAgentResponse);
  Events.On('agent-error', onAgentError);
  Events.On('agent-done', onAgentDone);
  Events.On('ignore-list-updated', onIgnoreListUpdated);
});

onUnmounted(() => {
  Events.Off('agent-status');
  Events.Off('agent-tool-start');
  Events.Off('agent-tool-result');
  Events.Off('agent-response');
  Events.Off('agent-error');
  Events.Off('agent-done');
  Events.Off('ignore-list-updated');
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.console-attachment-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
  padding: 0.375rem 0.5rem 0;
}

.console-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  box-sizing: border-box;
  overflow: hidden;
}

.console-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  text-align: center;
  color: var(--white);
  gap: 0.5rem;
}

.console-input {
  font-family: 'Inter', sans-serif;
  font-weight: 300;
  font-size: 14px;
  box-sizing: border-box;
  width: 100%;
  min-height: 36px;
  height: auto;
  max-height: 30vh;
  border-width: 0px;
  outline: none;
  resize: none;
  background-color: transparent;
  color: var(--white);
  padding: 0.5rem 0.5rem 0;
  overflow-y: auto;
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.console-input::-webkit-scrollbar {
  display: none;
}

.console-input-container {
  height: min-content;
  padding-bottom: 0.4rem;
  box-sizing: border-box;
  /* background-color: forestgreen; */
}

.console-input-wrapper {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  background-color: var(--steel);
  border-radius: var(--large-radius);
  transition: border-color 0.15s;
  outline-offset: -1px;
}

.console-input-wrapper:focus-within {
  outline: var(--transparent-line);
}

.console-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.125rem 0.25rem;
}

.console-toolbar-left {
  display: flex;
  align-items: center;
  gap: 0.125rem;
  overflow: hidden;
}

.console-toolbar-right {
  display: flex;
  align-items: center;
  gap: 0.125rem;
}

.console-messages {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-height: 0;
  padding: 0.75rem 0.5rem;
  overflow-y: auto;
}

.console-messages::-webkit-scrollbar {
  width: 4px;
}

.console-messages::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.console-messages::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.console-link {
  color: var(--accent-color);
  cursor: pointer;
  text-decoration: underline;
}

.empty-subtext {
  font-size: 12px;
  max-width: 250px;
}

.empty-text {
  font-size: 1.125rem;
  font-weight: 500;
  color: var(--silver);
}

.general-pane-root {
  /* padding: .5rem 0; */
  box-sizing: border-box;
}

/* User message — compact right-aligned bubble */
.msg-user {
  display: flex;
  justify-content: flex-end;
  margin: 0.5rem 0;
  font-weight: 400;
}

.msg-user-bubble {
  max-width: 80%;
  padding: 0.5rem 0.75rem;
  background-color: var(--midnight-steel);
  color: var(--white);
  border-radius: 12px 12px 2px 12px;
  font-size: 13px;
  line-height: 1.45;
  word-wrap: break-word;
  white-space: pre-wrap;
  flex-direction: column;
}

.msg-context-tag {
  /* display: inline-block; */
  max-width: 100%;
  font-size: 10px;
  font-weight: 500;
  color: var(--silver);
  background-color: var(--black-steel);
  padding: 0.125rem 0.375rem;
  border-radius: 4px;
  margin-bottom: 0.25rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  letter-spacing: 0.02em;
  box-sizing: border-box;
  width: 100%;
}

/* Assistant message — no bubble, plain left-aligned text */
.msg-assistant {
  padding: 0.375rem 0.25rem;
}

.msg-assistant-text {
  font-size: 13px;
  line-height: 1.6;
  color: var(--white);
  font-weight: 400;
  word-wrap: break-word;
}

.msg-assistant-text :deep(strong) {
  color: var(--white);
  font-weight: 600;
}

.msg-assistant-text :deep(.code-block) {
  display: block;
  padding: 0.625rem 0.75rem;
  margin: 0.375rem 0;
  background-color: var(--black-steel);
  border-radius: var(--small-radius);
  border: 1px solid var(--light-steel);
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  overflow-x: auto;
  white-space: pre;
  color: var(--silver);
}

.msg-assistant-text :deep(.inline-code) {
  padding: 0.1rem 0.35rem;
  background-color: var(--black-steel);
  border-radius: 3px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  color: var(--silver);
}

/* Tool call indicator — collapsible inline row */
.msg-tool {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.5rem;
  margin: 0.125rem 0;
  border-radius: var(--small-radius);
  cursor: pointer;
  user-select: none;
  transition: background-color 0.15s;
  color: var(--white);
  font-weight: 400;
}

.msg-tool:hover {
  background-color: var(--hover);
}

.msg-tool-icon {
  width: 14px;
  height: 14px;
  opacity: 0.5;
}

.msg-tool-label {
  font-size: 12px;
  color: var(--silver);
  text-transform: capitalize;
}

/* Status indicator — animated dot + text */
.msg-status {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.25rem;
  margin: 0.125rem 0;
}

.msg-status span {
  font-size: 12px;
  color: var(--silver);
  font-style: italic;
}

.msg-status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: var(--accent-color);
  animation: pulse 1.2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 0.4; transform: scale(0.9); }
  50% { opacity: 1; transform: scale(1.1); }
}

/* Error message */
.msg-error {
  padding: 0.375rem 0.5rem;
  margin: 0.125rem 0;
  border-left: 2px solid var(--danger);
  background-color: rgba(231, 76, 60, 0.06);
  border-radius: 0 var(--small-radius) var(--small-radius) 0;
}

.msg-error-text {
  font-size: 13px;
  line-height: 1.5;
  color: var(--danger);
  word-wrap: break-word;
}
</style>