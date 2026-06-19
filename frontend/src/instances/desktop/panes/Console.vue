<template>
  <div class="general-pane-root">
    <div v-if="!isModal && isConsoleModalOpen" class="console-placeholder">
      <div class="empty-text">{{ $t('panes.consoleOpenInModal') }}</div>
      <div class="empty-subtext">{{ $t('panes.consoleOpenInModalSubtext') }}</div>
    </div>

    <div v-else class="console-container">
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
            <div class="msg-assistant-text">
              <template v-for="(segment, segIndex) in parseAssistantSegments(message.content)" :key="segIndex">
                <ConsoleChip v-if="segment.type === 'chip'" :type="segment.entityType" :entityId="segment.id" :fallbackLabel="segment.label" />

                <span v-else class="msg-assistant-segment" v-html="segment.html"></span>
              </template>
            </div>
          </div>
        </template>

        <div v-if="shouldShowRetry" class="console-retry-row">
          <ActionButton
            :icon="getAppIcon('refresh')"
            :showLabel="true"
            :label="$t('panes.retry')"
            :useOutline="true"
            :buttonFunction="retryLastMessage"
          />
        </div>

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
        <div class="console-input-wrapper" id="agent-console-drop-zone" data-file-drop-target>
          <div v-if="attachmentPath" class="console-attachment-row">
            <Chip :icon="getAppIcon('paper-clip')" :label="attachmentName" :onRemove="removeAttachment" />
          </div>

          <textarea ref="textareaRef" v-model="currentMessage" class="console-input" type="text" :placeholder="inputPlaceholder"
            spellcheck="false" @input="handleInput" @keydown.enter.exact.prevent="sendMessage" :disabled="isProcessing" />

          <div class="console-toolbar">
            <div class="console-toolbar-left">
              <ActionButton :icon="getAppIcon('paper-clip')" :showLabel="false" v-tooltip="$t('panes.attachFile')" :buttonFunction="selectAttachment" />

              <div v-if="isApiKeyConfigured && availableModels.length" class="console-model-select"
                v-tooltip="$t('panes.agentModel')">
                <DropDownBox :items="availableModels" :selectedItem="selectedModel" :onSelect="selectModel"
                  :fullWidth="true" :useFilter="false" :placeHolder="$t('panes.agentModel')" />
              </div>
            </div>

            <div class="console-toolbar-right">
              <ActionButton v-if="!isModal" :icon="getAppIcon('arrows-expand')" :showLabel="false"
                v-tooltip="$t('panes.expandConsole')" :buttonFunction="openConsoleModal" />

              <ActionButton :icon="getAppIcon('broom')" :showLabel="false" :isDisabled="!messages.length || isProcessing"
                v-tooltip="$t('panes.clearChat')" :buttonFunction="clearChat" />

              <ActionButton v-if="isProcessing" :icon="getAppIcon('close-circle')" :showLabel="false"
                v-tooltip="$t('panes.stopAgent')" :buttonFunction="stopAgent" />

              <ActionButton v-else :icon="getAppIcon('send')" :showLabel="false" :isDisabled="!currentMessage.trim()"
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
import { expandShortcut } from '@/lib/agentShortcuts';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import Chip from '@/instances/common/components/Chip.vue';
import ConsoleChip from '@/instances/desktop/components/ConsoleChip.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';

// services
import { AgentService, CollectionService, DialogService, FSService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useSettingsStore } from '@/stores/settings';
import { useStageStore } from '@/stores/stages';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const settings = useSettingsStore();
const stage = useStageStore();

// props
const props = defineProps({
  isModal: { type: Boolean, default: false }
});

const { t } = useI18n();

// refs
const attachmentPath = ref('');
const availableModels = ref([]);
const currentMessage = ref('');
const currentProvider = ref('');
const isApiKeyConfigured = ref(false);
const isProcessing = ref(false);
const messages = ref([]);
const messagesContainer = ref(null);
const selectedModel = ref('');
const textareaRef = ref(null);

// computed properties
const attachmentName = computed(() => {
  if (!attachmentPath.value) return '';
  return attachmentPath.value.split(/[/\\]/).pop();
});

const emptyStateSubtext = computed(() => t('panes.performOperation', { itemType: itemType.value }));

const emptyStateTitle = computed(() => t('panes.startConversation'));

const inputPlaceholder = computed(() => {
  if (isProcessing.value) return t('panes.agentWorking');
  if (!isApiKeyConfigured.value) return t('panes.configureLlmInSettings');
  return t('panes.askAnything');
});

const isConsoleModalOpen = computed(() => modals.modalStates.consoleModal);

const shouldShowRetry = computed(() => {
  return messages.value.length > 0 && !isProcessing.value;
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
  apply_workflow: 'flow-chart',
  batch_update_asset_types: 'brush',
  batch_update_collection_types: 'brush',
  bulk_change_asset_type: 'brush',
  change_asset_type: 'brush',
  change_collaborator_role: 'scale',
  list_roles: 'scale',
  list_workflows: 'flow-chart',
  reveal_asset_on_disk: 'folder-arrow-up-right',
  reveal_in_browser: 'file-search',
  search_project_text: 'file-search',
  update_asset_type: 'brush',
  update_collection_type: 'brush',
  update_role: 'scale',
};

// methods

const checkApiKeyStatus = async () => {
  try {
    const status = await AgentService.GetAPIKeyStatus();
    isApiKeyConfigured.value = status.configured;
    currentProvider.value = status.provider || '';
    if (isApiKeyConfigured.value && currentProvider.value) {
      await loadModelOptions();
    } else {
      availableModels.value = [];
      selectedModel.value = '';
    }
  } catch {
    isApiKeyConfigured.value = false;
    currentProvider.value = '';
    availableModels.value = [];
    selectedModel.value = '';
  }
};

const loadModelOptions = async () => {
  try {
    const [models, chosen] = await Promise.all([
      AgentService.GetAvailableModels(currentProvider.value),
      AgentService.GetSelectedModel(currentProvider.value),
    ]);
    availableModels.value = Array.isArray(models) ? models : [];
    selectedModel.value = chosen || '';
  } catch {
    availableModels.value = [];
    selectedModel.value = '';
  }
};

const selectModel = async (newModel) => {
  selectedModel.value = newModel || '';
  if (!currentProvider.value) return;
  try {
    await AgentService.SetSelectedModel(currentProvider.value, selectedModel.value);
  } catch (error) {
    console.error('AgentService.SetSelectedModel failed:', error);
  }
};

const openConsoleModal = () => {
  modals.setModalVisibility('consoleModal', true);
};

const clearChat = async () => {
  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath) return;
  try {
    await AgentService.ClearChatSession(projectPath);
    messages.value = [];
  } catch { /* ignore */ }
};

const escapeHtml = (text) => {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
};

const formatToolLabel = (toolName, count) => {
  const label = toolName.replace(/_/g, ' ');
  if (count > 1) return `${label} (${count})`;
  return label;
};

// Matches inline entity references in agent text, e.g. [[asset:abc-123|My Asset]].
const ENTITY_TOKEN_REGEX = /\[\[(asset|collection|user):([A-Za-z0-9_-]+)\|([\s\S]*?)\]\]/g;

// Splits assistant text into a list of formatted text segments and entity chip
// segments, removing any leftover markdown emphasis around the chip boundaries.
const parseAssistantSegments = (text) => {
  if (!text) return [];
  const stripTrailingEmphasis = (s) => s.replace(/(\*{1,3}|_{1,3}|`)$/, '');
  const stripLeadingEmphasis = (s) => s.replace(/^(\*{1,3}|_{1,3}|`)/, '');

  const rawChunks = [];
  let lastIndex = 0;
  const regex = new RegExp(ENTITY_TOKEN_REGEX.source, 'g');
  let match;
  while ((match = regex.exec(text)) !== null) {
    if (match.index > lastIndex) {
      rawChunks.push({ type: 'text', raw: text.slice(lastIndex, match.index) });
    }
    rawChunks.push({ type: 'chip', entityType: match[1], id: match[2], label: match[3] });
    lastIndex = match.index + match[0].length;
  }
  if (lastIndex < text.length) {
    rawChunks.push({ type: 'text', raw: text.slice(lastIndex) });
  }

  const segments = [];
  for (let i = 0; i < rawChunks.length; i++) {
    const chunk = rawChunks[i];
    if (chunk.type === 'chip') {
      segments.push(chunk);
      continue;
    }
    let raw = chunk.raw;
    if (i > 0 && rawChunks[i - 1].type === 'chip') raw = stripLeadingEmphasis(raw);
    if (i < rawChunks.length - 1 && rawChunks[i + 1].type === 'chip') raw = stripTrailingEmphasis(raw);
    if (raw) segments.push({ type: 'text', html: formatContent(raw) });
  }
  return segments;
};

// Renders a small subset of markdown: ISO dates, code blocks, inline code, bold, line breaks.
const formatContent = (text) => {
  if (!text) return '';
  let escaped = escapeHtml(text);
  escaped = escaped.replace(/(\d{4}-\d{2}-\d{2})[T ](?:at )?((\d{2}):(\d{2})(?::\d{2})?Z?)/g, (match, datePart, timePart, hours, minutes) => {
    try {
      const date = new Date(datePart + 'T' + timePart.replace('at ', '') + (timePart.endsWith('Z') ? '' : 'Z'));
      if (isNaN(date)) return match;
      return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' }) + ' at ' + date.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' });
    } catch { return match; }
  });
  escaped = escaped.replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code class="code-block">$2</code></pre>');
  escaped = escaped.replace(/`([^`]+)`/g, '<code class="inline-code">$1</code>');
  escaped = escaped.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
  escaped = escaped.replace(/\n/g, '<br>');
  return escaped;
};

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Pulls a leading `[Context: ...]` block off a user message and returns it separately from the body.
const parseUserContext = (content) => {
  if (!content) return { context: '', body: content };
  const match = content.match(/^\[Context:\s*(.+?)\]\n?/);
  if (match) {
    const display = match[1].replace(/\s*\([^)]*\)/g, '').replace(/"/g, '');
    return { context: display, body: content.slice(match[0].length) };
  }
  return { context: '', body: content };
};

const openAdvancedSettings = () => {
  settings.pendingTab = 'advanced';
  stage.setStageVisibility('settings', true);
};

const getToolIcon = (toolName) => toolIconMap[toolName] || 'cog';

// Grows or shrinks the input textarea to match its content.
const handleInput = () => {
  if (!textareaRef.value) return;
  textareaRef.value.style.height = 'auto';
  textareaRef.value.style.height = textareaRef.value.scrollHeight + 'px';
};

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

const removeAttachment = () => {
  attachmentPath.value = '';
};

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
  });
};

// Allowed extensions and max size for agent attachments. Kept in sync with
// what agent.ReadAttachment can usefully parse (text + PDF).
const ATTACHMENT_ALLOWED_EXTENSIONS = [
  'txt', 'md', 'markdown', 'rst', 'log',
  'json', 'yml', 'yaml', 'toml', 'xml', 'csv', 'tsv', 'ini', 'env',
  'html', 'htm', 'css', 'scss',
  'js', 'jsx', 'ts', 'tsx', 'vue', 'svelte',
  'py', 'go', 'rs', 'java', 'kt', 'swift', 'c', 'h', 'cpp', 'hpp', 'cs',
  'rb', 'php', 'sh', 'bash', 'ps1', 'bat',
  'sql', 'graphql', 'proto',
  'pdf',
];
const ATTACHMENT_MAX_BYTES = 5 * 1024 * 1024;
const ATTACHMENT_DIALOG_FILTER = ATTACHMENT_ALLOWED_EXTENSIONS.map((e) => '*.' + e).join(';');

// Validates an attachment path against the allowed extension list and size cap.
// Returns true when the path can be used as an attachment; otherwise surfaces a
// notification and returns false.
const validateAttachment = async (path) => {
  if (!path) return false;
  const name = path.split(/[/\\]/).pop() || path;
  const ext = (name.includes('.') ? name.split('.').pop() : '').toLowerCase();
  if (!ext || !ATTACHMENT_ALLOWED_EXTENSIONS.includes(ext)) {
    notificationStore.errorNotification(
      t('panes.attachmentTypeNotSupported', { name }),
      ATTACHMENT_ALLOWED_EXTENSIONS.join(', ')
    );
    return false;
  }
  try {
    const info = await FSService.FileStat(path);
    if (info.IsDir) {
      notificationStore.errorNotification(t('panes.attachmentMustBeFile', { name }), '');
      return false;
    }
    if (typeof info.Size === 'number' && info.Size > ATTACHMENT_MAX_BYTES) {
      notificationStore.errorNotification(
        t('panes.attachmentTooLarge', { name }),
        info.FormattedSize || ''
      );
      return false;
    }
  } catch (err) {
    notificationStore.errorNotification(t('panes.attachmentReadFailed', { name }), `${err}`);
    return false;
  }
  return true;
};

const selectAttachment = async () => {
  try {
    const path = await DialogService.SelectFileDialog(t('panes.attachFile'), ATTACHMENT_DIALOG_FILTER);
    if (!path) return;
    if (await validateAttachment(path)) attachmentPath.value = path;
  } catch { /* user cancelled */ }
};

// Accepts files dropped onto the console area as the message attachment.
// Only the first file is used; folders and additional files are ignored.
const onFilesDropped = async (event) => {
  const details = event?.data?.details;
  if (details?.id !== 'agent-console-drop-zone') return;
  const files = event?.data?.files;
  if (!files || !files.length) return;
  const path = files[0];
  if (await validateAttachment(path)) attachmentPath.value = path;
};

// Describes the user's current selection (collection, items, active stage) for the agent.
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

const sendMessage = async () => {
  if (!currentMessage.value.trim() || isProcessing.value) return;
  if (!isApiKeyConfigured.value) return;

  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath) {
    addMessage('error', 'No project is currently open. Please open a project first.');
    return;
  }

  const rawInput = currentMessage.value.trim();
  const shortcut = expandShortcut(rawInput);
  if (shortcut?.error) {
    addMessage('user', rawInput);
    addMessage('error', shortcut.error);
    currentMessage.value = '';
    if (textareaRef.value) textareaRef.value.style.height = 'auto';
    return;
  }
  if (shortcut?.localReply) {
    addMessage('user', rawInput);
    addMessage('assistant', shortcut.localReply);
    currentMessage.value = '';
    if (textareaRef.value) textareaRef.value.style.height = 'auto';
    return;
  }
  const expanded = shortcut?.prompt ?? rawInput;

  addMessage('user', rawInput);
  const context = buildSelectionContext();
  const messageContent = context + expanded;
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

const retryLastMessage = async () => {
  if (isProcessing.value) return;
  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath) return;

  const lastUserIdx = messages.value.map(m => m.type).lastIndexOf('user');
  if (lastUserIdx === -1) return;

  const lastUserMsg = messages.value[lastUserIdx];
  const content = lastUserMsg.content;

  messages.value = messages.value.slice(0, lastUserIdx);
  addMessage('user', content);
  isProcessing.value = true;

  try {
    await AgentService.RetryLastTurn(projectPath);
  } catch (err) {
    addMessage('error', `${err}`);
    isProcessing.value = false;
  }
};

const addMessage = (type, content) => {
  messages.value.push({ id: messages.value.length + 1, type, content });
  scrollToBottom();
};

// --- Wails event handlers ---

const onAgentStatus = (event) => {
  const lastMsg = messages.value[messages.value.length - 1];
  if (lastMsg && lastMsg.type === 'status') {
    lastMsg.content = event.data;
  } else {
    messages.value.push({ id: messages.value.length + 1, type: 'status', content: event.data });
  }
  scrollToBottom();
};

// Folds repeated calls of the same tool into one row with a count.
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

// Appends the agent's reply and clears any pending status row above it.
const onAgentResponse = (event) => {
  if (messages.value.length && messages.value[messages.value.length - 1].type === 'status') {
    messages.value.pop();
  }
  addMessage('assistant', event.data);
};

const onAgentError = (event) => {
  if (messages.value.length && messages.value[messages.value.length - 1].type === 'status') {
    messages.value.pop();
  }
  addMessage('error', event.data);
  isProcessing.value = false;
};

const onAgentDone = (event) => {
  if (messages.value.length && messages.value[messages.value.length - 1].type === 'status') {
    messages.value.pop();
  }
  isProcessing.value = false;
  if (event?.data?.mutated) {
    emitter.emit('refresh-browser');
    notificationStore.addNotification(t('panes.projectUpdatedByAgent'), '', 'success');
  }
};

const onAgentCancelled = (event) => {
  if (messages.value.length && messages.value[messages.value.length - 1].type === 'status') {
    messages.value.pop();
  }
  isProcessing.value = false;
  if (event?.data?.mutated) {
    emitter.emit('refresh-browser');
  }
};

const stopAgent = async () => {
  if (!projectStore.activeProject) return;
  try {
    await AgentService.CancelRun(projectStore.activeProject.uri);
  } catch (error) {
    console.error('AgentService.CancelRun failed:', error);
  }
};

// Refreshes the project's ignore_list when the agent edits it server-side.
const onIgnoreListUpdated = (event) => {
  if (projectStore.activeProject && event.data) {
    projectStore.activeProject.ignore_list = event.data;
  }
};

// Jumps the browser to an asset or collection the agent points the user at.
const onAgentRevealInBrowser = async (event) => {
  const data = event.data;
  if (!data || !projectStore.activeProject) return;
  try {
    commonStore.activeWorkspace = 'Default';
    commonStore.viewSearchQuery = '';
    commonStore.resetFilters();
    commonStore.navigatorMode = true;

    const projectUri = projectStore.activeProject.uri;
    if (data.kind === 'asset') {
      if (data.collection_id) {
        const parent = await CollectionService.GetCollectionByID(projectUri, data.collection_id);
        if (parent) {
          collectionStore.navigatedCollection = parent;
          collectionStore.selectedCollection = parent;
        }
      } else {
        collectionStore.navigatedCollection = null;
        collectionStore.selectedCollection = null;
      }
      assetStore.selectedAsset = { id: data.asset_id, name: data.name, collection_id: data.collection_id || '' };
      emitter.emit('refresh-browser');
    } else if (data.kind === 'collection') {
      const collection = await CollectionService.GetCollectionByID(projectUri, data.collection_id);
      if (collection) {
        collectionStore.navigatedCollection = collection;
        collectionStore.selectedCollection = collection;
        emitter.emit('refresh-browser');
      }
    }
  } catch (error) {
    console.error('agent-reveal-in-browser failed:', error);
  }
};

// Applies a filter payload the agent built so the browser narrows to the
// requested items. Mirrors the field assignments the filter menus do, so the
// existing FilterBar Clear button and "modified workspace" indicator both work.
const onAgentApplyFilter = (event) => {
  const payload = event?.data?.applied;
  if (!payload) return;
  if (Array.isArray(payload.asset_filters)) commonStore.assetFilters = payload.asset_filters;
  if (Array.isArray(payload.collection_filters)) commonStore.collectionFilters = payload.collection_filters;
  if (Array.isArray(payload.resource_filters)) commonStore.resourceFilters = payload.resource_filters;
  if (typeof payload.has_assignees === 'boolean') commonStore.hasAssignees = payload.has_assignees;
  if (typeof payload.no_assignees === 'boolean') commonStore.noAssignees = payload.no_assignees;
  if (typeof payload.use_deep === 'boolean') commonStore.useDeep = payload.use_deep;
  if (typeof payload.view_search_query === 'string') commonStore.viewSearchQuery = payload.view_search_query;
  if (typeof payload.show_collections === 'boolean') commonStore.showCollections = payload.show_collections;
  if (typeof payload.show_assets === 'boolean') commonStore.showAssets = payload.show_assets;
  if (typeof payload.show_resources === 'boolean') commonStore.showResources = payload.show_resources;
  if (typeof payload.only_assets === 'boolean') commonStore.onlyAssets = payload.only_assets;
  emitter.emit('refresh-browser');
};

// Clears all browser filters when the agent issues clear_browser_filter.
const onAgentClearFilter = () => {
  commonStore.resetFilters();
  emitter.emit('refresh-browser');
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
  Events.On('agent-cancelled', onAgentCancelled);
  Events.On('ignore-list-updated', onIgnoreListUpdated);
  Events.On('agent-reveal-in-browser', onAgentRevealInBrowser);
  Events.On('agent-apply-filter', onAgentApplyFilter);
  Events.On('agent-clear-filter', onAgentClearFilter);
  Events.On('files-dropped', onFilesDropped);
});

onUnmounted(() => {
  Events.Off('agent-status');
  Events.Off('agent-tool-start');
  Events.Off('agent-tool-result');
  Events.Off('agent-response');
  Events.Off('agent-error');
  Events.Off('agent-done');
  Events.Off('agent-cancelled');
  Events.Off('ignore-list-updated');
  Events.Off('agent-reveal-in-browser');
  Events.Off('agent-apply-filter');
  Events.Off('agent-clear-filter');
  Events.Off('files-dropped');
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.console-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  text-align: center;
  color: var(--text);
  gap: 0.5rem;
}

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
  color: var(--text);
  gap: 0.5rem;
}

.console-input {
  font-family: 'Inter', sans-serif;
  font-weight: 300;
  font-size: 13px;
  box-sizing: border-box;
  width: 100%;
  min-height: 50px;
  height: auto;
  max-height: 30vh;
  border-width: 0px;
  outline: none;
  resize: none;
  background-color: transparent;
  color: var(--text);
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
}

.console-input-wrapper {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  background-color: var(--surface-3);
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

.console-model-select {
  margin-left: 0.25rem;
  width: 200px;
  min-width: 0;
  flex-shrink: 1;
}

.console-model-select :deep(.list-box-parent-text) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
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
  background-color: var(--surface-4);
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
  color: var(--text-muted);
}

.general-pane-root {
  /* padding: .5rem 0; */
  box-sizing: border-box;
}

.console-retry-row {
  display: flex;
  justify-content: flex-start;
  padding: 0.25rem 0.25rem 0.5rem;
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
  background-color: var(--bg);
  color: var(--text);
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
  color: var(--text-muted);
  background-color: var(--surface-1);
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
  line-height: 2;
  color: var(--text);
  font-weight: 400;
  word-wrap: break-word;
}

.msg-assistant-segment {
  vertical-align: middle;
}

.msg-assistant-text :deep(strong) {
  color: var(--text);
  font-weight: 600;
}

.msg-assistant-text :deep(.code-block) {
  display: block;
  padding: 0.625rem 0.75rem;
  margin: 0.375rem 0;
  background-color: var(--surface-1);
  border-radius: var(--small-radius);
  border: 1px solid var(--surface-4);
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  overflow-x: auto;
  white-space: pre;
  color: var(--text-muted);
}

.msg-assistant-text :deep(.inline-code) {
  padding: 0.1rem 0.35rem;
  background-color: var(--surface-1);
  border-radius: 3px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  color: var(--text-muted);
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
  color: var(--text);
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
  color: var(--text-muted);
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
  color: var(--text-muted);
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