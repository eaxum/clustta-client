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
            <div class="msg-user-content">
              <div class="msg-user-bubble">
                <div v-if="parseUserContext(message.content).context" class="msg-context-tag">{{ parseUserContext(message.content).context }}</div>
                <span v-html="formatContent(parseUserContext(message.content).body)"></span>
              </div>
              <div class="msg-message-actions">
                <ActionButton :icon="getAppIcon('copy')" :isDisabled="isProcessing"
                  v-tooltip="$t('common.copy')" :buttonFunction="() => copyMessage(message)" />
                <ActionButton v-if="isMostRecentUserMessage(message)" :icon="getAppIcon('edit')"
                  :isDisabled="isProcessing" v-tooltip="$t('common.edit')" :buttonFunction="() => editMessage(message)" />
                <ActionButton v-if="hasPersistedTurn(message)" :icon="getAppIcon('trash')"
                  :isDisabled="isProcessing" v-tooltip="$t('common.delete')" :buttonFunction="() => deleteMessage(message)" />
              </div>
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
              <div v-if="message.entities?.length && !hasInlineEntities(message.content)" class="msg-assistant-entities">
                <ConsoleChip
                  v-for="entity in visibleMessageEntities(message)"
                  :key="`${entity.type}:${entity.id}`"
                  :type="entity.type"
                  :entityId="entity.id"
                  :fallbackLabel="entity.name"
                />
                <span v-if="hiddenMessageEntityCount(message)" class="msg-entity-overflow">
                  +{{ hiddenMessageEntityCount(message) }} more
                </span>
              </div>
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
          <div class="empty-subtext">{{ $t('panes.agentHelpHintBefore') }} <button class="agent-inline-token empty-help-command" @click="insertHelpCommand">/help</button> {{ $t('panes.agentHelpHintAfter') }}</div>
        </div>

        <div v-if="!messages.length && !isApiKeyConfigured" class="console-empty">
          <div class="empty-text">{{ $t('panes.setupAiAgent') }}</div>
          <div class="empty-subtext">{{ $t('panes.configureLlmPre') }} <a class="console-link" @click="openAdvancedSettings">{{ $t('common.settings') }}</a> {{ $t('panes.configureLlmPost') }}</div>
        </div>
      </div>

      <div class="console-input-container">
        <div ref="inputWrapperRef" class="console-input-wrapper" id="agent-console-drop-zone" data-file-drop-target>
          <div v-if="attachmentPath" class="console-attachment-row">
            <Chip :icon="getAppIcon('paper-clip')" :label="attachmentName" :onRemove="removeAttachment" />
          </div>

          <div class="console-editor">
            <div ref="inputHighlightRef" class="console-input console-input-highlight" aria-hidden="true"
              v-html="composerHighlightHtml"></div>
            <textarea ref="textareaRef" v-model="currentMessage" class="console-input console-input-editor" type="text"
              :placeholder="inputPlaceholder" spellcheck="false" @input="handleInput" @click="updateComposerMenu"
              @keyup.stop="updateComposerMenu" @scroll="syncComposerScroll" @keydown.stop="handleComposerKeydown" @blur="closeComposerMenu"
              :disabled="isProcessing" />
          </div>

          <AgentComposerMenu :visible="composerMenuVisible" :title="composerMenuTitle" :items="composerMenuItems"
            :activeIndex="composerMenuIndex" :menuStyle="composerMenuStyle" :emptyText="composerMenuEmptyText"
            @select="selectComposerItem" />

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
import { agentShortcuts, expandShortcut, isAgentShortcut, restoreShortcutDisplay } from '@/lib/agentShortcuts';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import AgentComposerMenu from '@/instances/desktop/components/AgentComposerMenu.vue';
import Chip from '@/instances/common/components/Chip.vue';
import ConsoleChip from '@/instances/desktop/components/ConsoleChip.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';

// services
import { AgentService, ClipboardService, CollectionService, DialogService, FSService } from '@/services';

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
import { useAgentEntityCacheStore } from '@/stores/agentEntityCache';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const settings = useSettingsStore();
const stage = useStageStore();
const agentEntityCache = useAgentEntityCacheStore();

// props
const props = defineProps({
  isModal: { type: Boolean, default: false }
});

const { t } = useI18n();

const AGENT_DRAFT_STORAGE_PREFIX = 'clustta_agent_draft:';

// refs
const attachmentPath = ref('');
const availableModels = ref([]);
const currentMessage = ref('');
const currentProvider = ref('');
const isApiKeyConfigured = ref(false);
const isProcessing = ref(false);
const messages = ref([]);
const messagesContainer = ref(null);
const inputHighlightRef = ref(null);
const inputWrapperRef = ref(null);
const composerMenuIndex = ref(0);
const composerMenuStyle = ref({});
const composerTrigger = ref(null);
const recentResultEntities = ref([]);
const scriptReferences = ref([]);
const scriptReferencesLoadedFor = ref('');
const scriptReferencesLoadingFor = ref('');
const selectedScriptReferences = ref(new Map());
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

const composerHighlightHtml = computed(() => formatComposerText(currentMessage.value));

const composerMenuItems = computed(() => {
  const trigger = composerTrigger.value;
  if (!trigger) return [];
  const query = trigger.query.toLowerCase();
  if (trigger.character === '/') {
    return agentShortcuts
      .filter((shortcut) => shortcut.command.slice(1).includes(query))
      .map((shortcut) => ({
        key: shortcut.command,
        label: shortcut.command,
        meta: shortcut.args,
        description: shortcut.description,
        value: shortcut.command,
        type: 'command',
      }));
  }
  return scriptReferences.value
    .filter((reference) => reference.name.toLowerCase().includes(query))
    .map((reference) => ({
      key: reference.path,
      label: reference.name.toLowerCase().endsWith(reference.extension.toLowerCase())
        ? reference.name
        : `${reference.name}${reference.extension}`,
      reference,
      type: 'script',
    }));
});

const composerMenuEmptyText = computed(() => {
  if (composerTrigger.value?.character === '~' && !scriptReferences.value.length) {
    return t('panes.noConfiguredScripts');
  }
  return t('panes.noMatches');
});

const composerMenuTitle = computed(() => composerTrigger.value?.character === '~'
  ? t('panes.scripts')
  : t('panes.quickCommands'));
const composerMenuVisible = computed(() => !!composerTrigger.value && !isProcessing.value);

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

const protectAgentTokens = (text) => {
  const tokens = [];
  const protectedText = `${text || ''}`.replace(/(^|[\s(,])((?:\/[A-Za-z][\w-]*)|(?:~(?:"[^"\n]+"|[^\s,\n]+)))/g, (match, prefix, token) => {
    if (token.startsWith('/') && !isAgentShortcut(token)) return match;
    const index = tokens.push(token) - 1;
    return `${prefix}\uE000${index}\uE001`;
  });
  return { protectedText, tokens };
};

const restoreAgentTokens = (html, tokens) => html.replace(/\uE000(\d+)\uE001/g, (match, index) => {
  const token = tokens[Number(index)];
  const kind = token?.startsWith('/') ? 'command' : 'script';
  return `<span class="agent-inline-token agent-inline-token-${kind}">${escapeHtml(token)}</span>`;
});

const formatComposerText = (text) => {
  const { protectedText, tokens } = protectAgentTokens(text);
  const highlighted = restoreAgentTokens(escapeHtml(protectedText), tokens);
  return `${highlighted || ''}\n`;
};

const formatToolLabel = (toolName, count) => {
  const label = toolName.replace(/_/g, ' ');
  if (count > 1) return `${label} (${count})`;
  return label;
};

// Matches inline entity references in agent text, e.g. [[asset:abc-123|My Asset]].
const ENTITY_TOKEN_REGEX = /\[\[(asset|collection|untracked_asset|untracked_collection|user):([A-Za-z0-9_-]+)\|([\s\S]*?)\]\]/g;
const MAX_RESPONSE_ENTITIES = 3;

const hasInlineEntities = (text) => {
  if (!text) return false;
  return new RegExp(ENTITY_TOKEN_REGEX.source).test(text);
};

const visibleMessageEntities = (message) => (message?.entities || []).slice(0, MAX_RESPONSE_ENTITIES);

const hiddenMessageEntityCount = (message) => Math.max(
  0,
  (message?.entities?.length || 0) - MAX_RESPONSE_ENTITIES,
);

// Splits assistant text into a list of formatted text segments and entity chip
// segments, removing any leftover markdown emphasis around the chip boundaries.
const parseAssistantSegments = (text) => {
  if (!text) return [];
  const stripTrailingEmphasis = (s) => s.replace(/(\*{1,3}|_{1,3}|`)$/, '');
  const stripLeadingEmphasis = (s) => s.replace(/^(\*{1,3}|_{1,3}|`)/, '');

  const rawChunks = [];
  let entityCount = 0;
  let hiddenEntityCount = 0;
  let lastIndex = 0;
  const regex = new RegExp(ENTITY_TOKEN_REGEX.source, 'g');
  let match;
  while ((match = regex.exec(text)) !== null) {
    if (match.index > lastIndex) {
      rawChunks.push({ type: 'text', raw: text.slice(lastIndex, match.index) });
    }
    if (entityCount < MAX_RESPONSE_ENTITIES) {
      rawChunks.push({ type: 'chip', entityType: match[1], id: match[2], label: match[3] });
      entityCount++;
    } else {
      hiddenEntityCount++;
    }
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
  if (hiddenEntityCount > 0) {
    segments.push({ type: 'text', html: `<span class="msg-entity-overflow">+${hiddenEntityCount} more</span>` });
  }
  return segments;
};

// Renders a small subset of markdown: ISO dates, code blocks, inline code, bold, line breaks.
const formatContent = (text) => {
  if (!text) return '';
  const { protectedText, tokens } = protectAgentTokens(text);
  let escaped = escapeHtml(protectedText);
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
  return restoreAgentTokens(escaped, tokens);
};

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Pulls a leading `[Context: ...]` block off a user message and returns it separately from the body.
const parseUserContext = (content) => {
  if (!content) return { context: '', body: content };
  const jsonContext = content.match(/^\[Context JSON\]\n[^\n]*\n?/);
  if (jsonContext) return { context: '', body: content.slice(jsonContext[0].length) };
  const match = content.match(/^\[Context:\s*(.+?)\]\n?/);
  if (match) {
    const display = match[1].replace(/\s*\([^)]*\)/g, '').replace(/"/g, '');
    return { context: display, body: content.slice(match[0].length) };
  }
  return { context: '', body: content };
};

const hasPersistedTurn = (message) => Number.isInteger(message?.turnIndex);

const mostRecentTurnIndex = computed(() => messages.value.reduce((highest, message) => {
  return message.type === 'user' && Number.isInteger(message.turnIndex)
    ? Math.max(highest, message.turnIndex)
    : highest;
}, -1));

const isMostRecentUserMessage = (message) => {
  return hasPersistedTurn(message) && message.turnIndex === mostRecentTurnIndex.value;
};

const nextTurnIndex = () => messages.value.reduce((highest, message) => {
  return Number.isInteger(message?.turnIndex) ? Math.max(highest, message.turnIndex) : highest;
}, -1) + 1;

const copyMessage = async (message) => {
  try {
    await ClipboardService.WriteText(parseUserContext(message.content).body || '');
  } catch (error) {
    console.error('ClipboardService.WriteText failed:', error);
  }
};

const editMessage = async (message) => {
  if (!isMostRecentUserMessage(message) || isProcessing.value) return;
  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath) return;
  currentMessage.value = parseUserContext(message.content).body || '';
  try {
    await AgentService.DeleteChatTurn(projectPath, message.turnIndex);
    await loadChatHistory();
  } catch (error) {
    console.error('AgentService.DeleteChatTurn(edit) failed:', error);
  }
  await nextTick();
  if (textareaRef.value) {
    textareaRef.value?.focus();
    handleInput();
  }
};

const deleteMessage = async (message) => {
  if (!hasPersistedTurn(message) || isProcessing.value) return;
  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath) return;
  try {
    await AgentService.DeleteChatTurn(projectPath, message.turnIndex);
    await loadChatHistory();
  } catch (error) {
    console.error('AgentService.DeleteChatTurn failed:', error);
  }
};

const openAdvancedSettings = () => {
  settings.pendingTab = 'advanced';
  stage.setStageVisibility('settings', true);
};

const getToolIcon = (toolName) => toolIconMap[toolName] || 'cog';

const insertHelpCommand = async () => {
  currentMessage.value = '/help';
  await nextTick();
  textareaRef.value?.focus();
  textareaRef.value?.setSelectionRange(currentMessage.value.length, currentMessage.value.length);
  handleInput();
};

const draftStorageKey = (projectPath) => `${AGENT_DRAFT_STORAGE_PREFIX}${encodeURIComponent(projectPath)}`;

const saveComposerDraft = (projectPath = projectStore.activeProject?.uri) => {
  if (!projectPath) return;
  try {
    const key = draftStorageKey(projectPath);
    if (!currentMessage.value) {
      localStorage.removeItem(key);
      return;
    }
    const scriptReferences = [...selectedScriptReferences.value.values()]
      .filter((reference) => currentMessage.value.includes(reference.token));
    localStorage.setItem(key, JSON.stringify({ message: currentMessage.value, scriptReferences }));
  } catch (error) {
    console.warn('Failed to save agent draft:', error);
  }
};

const restoreComposerDraft = (projectPath) => {
  currentMessage.value = '';
  selectedScriptReferences.value = new Map();
  if (!projectPath) return;
  try {
    const savedDraft = localStorage.getItem(draftStorageKey(projectPath));
    if (!savedDraft) return;
    const draft = JSON.parse(savedDraft);
    currentMessage.value = typeof draft?.message === 'string' ? draft.message : '';
    const references = Array.isArray(draft?.scriptReferences) ? draft.scriptReferences : [];
    selectedScriptReferences.value = new Map(references
      .filter((reference) => reference?.token && currentMessage.value.includes(reference.token))
      .map((reference) => [reference.token, reference]));
  } catch (error) {
    console.warn('Failed to restore agent draft:', error);
  }
};

const loadScriptReferences = async () => {
  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath || scriptReferencesLoadedFor.value === projectPath || scriptReferencesLoadingFor.value === projectPath) return;
  scriptReferencesLoadingFor.value = projectPath;
  try {
    scriptReferences.value = await AgentService.ListScriptReferences(projectPath) || [];
    scriptReferencesLoadedFor.value = projectPath;
  } catch (error) {
    scriptReferences.value = [];
    scriptReferencesLoadedFor.value = '';
    console.error('AgentService.ListScriptReferences failed:', error);
  } finally {
    scriptReferencesLoadingFor.value = '';
  }
};

const findComposerTrigger = () => {
  const input = textareaRef.value;
  if (!input) return null;
  const cursor = input.selectionStart;
  const prefix = currentMessage.value.slice(0, cursor);
  const match = prefix.match(/(^|[\s,(])([/~])([^\s,]*)$/);
  if (!match) return null;
  return {
    character: match[2],
    query: match[3].replace(/^"/, ''),
    start: cursor - match[2].length - match[3].length,
    end: cursor,
  };
};

const updateComposerMenuAnchor = () => {
  const wrapper = inputWrapperRef.value;
  if (!wrapper || !composerTrigger.value) return;
  const rect = wrapper.getBoundingClientRect();
  const viewportPadding = 8;
  const menuGap = 6;
  composerMenuStyle.value = {
    left: `${Math.max(viewportPadding, rect.left)}px`,
    bottom: `${window.innerHeight - rect.top + menuGap}px`,
    width: `${Math.min(rect.width, window.innerWidth - (viewportPadding * 2))}px`,
    maxHeight: `${Math.max(100, Math.min(320, rect.top - menuGap - viewportPadding))}px`,
  };
};

const updateComposerMenu = async () => {
  composerTrigger.value = findComposerTrigger();
  if (!composerTrigger.value) return;
  if (composerTrigger.value.character === '~') await loadScriptReferences();
  composerMenuIndex.value = Math.min(composerMenuIndex.value, Math.max(0, composerMenuItems.value.length - 1));
  await nextTick();
  updateComposerMenuAnchor();
};

const closeComposerMenu = () => {
  window.setTimeout(() => { composerTrigger.value = null; }, 0);
};

const scriptToken = (reference) => reference.name.includes(' ') ? `~"${reference.name}"` : `~${reference.name}`;

const selectComposerItem = async (item) => {
  const trigger = composerTrigger.value;
  if (!trigger) return;
  const inserted = item.type === 'script' ? scriptToken(item.reference) : item.value;
  currentMessage.value = `${currentMessage.value.slice(0, trigger.start)}${inserted} ${currentMessage.value.slice(trigger.end)}`;
  if (item.type === 'script') {
    const selected = new Map(selectedScriptReferences.value);
    selected.set(inserted, { ...item.reference, token: inserted });
    selectedScriptReferences.value = selected;
  }
  composerTrigger.value = null;
  await nextTick();
  const cursor = trigger.start + inserted.length + 1;
  textareaRef.value?.focus();
  textareaRef.value?.setSelectionRange(cursor, cursor);
  handleInput();
};

const handleComposerKeydown = (event) => {
  if (composerMenuVisible.value) {
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault();
      event.stopPropagation();
      const direction = event.key === 'ArrowDown' ? 1 : -1;
      const length = composerMenuItems.value.length;
      if (length) composerMenuIndex.value = (composerMenuIndex.value + direction + length) % length;
      return;
    }
    if ((event.key === 'Enter' || event.key === 'Tab') && composerMenuItems.value.length) {
      event.preventDefault();
      event.stopPropagation();
      selectComposerItem(composerMenuItems.value[composerMenuIndex.value]);
      return;
    }
    if (event.key === 'Escape') {
      event.preventDefault();
      event.stopPropagation();
      composerTrigger.value = null;
      return;
    }
  }
  if (event.key === 'Enter' && !event.shiftKey && !event.ctrlKey && !event.altKey && !event.metaKey) {
    event.preventDefault();
    sendMessage();
  }
};

const syncComposerScroll = () => {
  if (!textareaRef.value || !inputHighlightRef.value) return;
  inputHighlightRef.value.scrollTop = textareaRef.value.scrollTop;
  inputHighlightRef.value.scrollLeft = textareaRef.value.scrollLeft;
  updateComposerMenuAnchor();
};

const clearComposer = () => {
  currentMessage.value = '';
  composerTrigger.value = null;
  selectedScriptReferences.value = new Map();
  if (textareaRef.value) textareaRef.value.style.height = 'auto';
};

const activeScriptReferences = () => [...selectedScriptReferences.value.entries()]
  .filter(([token]) => currentMessage.value.includes(token))
  .map(([, reference]) => reference);

// Grows or shrinks the input and refreshes token/menu state.
const handleInput = () => {
  if (!textareaRef.value) return;
  textareaRef.value.style.height = 'auto';
  textareaRef.value.style.height = textareaRef.value.scrollHeight + 'px';
  const selected = new Map([...selectedScriptReferences.value].filter(([token]) => currentMessage.value.includes(token)));
  selectedScriptReferences.value = selected;
  nextTick(() => {
    syncComposerScroll();
    updateComposerMenu();
  });
};

const loadChatHistory = async () => {
  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath) return;
  try {
    const history = await AgentService.GetChatHistory(projectPath);
    messages.value = (history || []).map((message, index) => {
      const parsed = message.type === 'user' ? parseUserContext(message.content) : null;
      const restoredShortcut = parsed ? restoreShortcutDisplay(parsed.body) : '';
      return {
        id: index + 1,
        ...message,
        content: restoredShortcut || message.content,
      };
    });
    scrollToBottom();
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
const buildSelectionContext = (referencedScripts = []) => {
  const context = {
    current_location: null,
    here_scope: { source: 'here', recursive: !!commonStore.onlyAssets },
    selection: [],
    active_view: stage.activeStage || '',
    script_references: referencedScripts.map(({ token, type, id, name, path, extension, tracked }) => ({
      token, type, id, name, path, extension, tracked,
    })),
  };

  // "Here" means the collection currently open in the browser. A selected
  // collection may merely be highlighted inside that location, so navigation
  // takes precedence and project root is represented by null.
  const currentCollection = commonStore.navigatorMode
    ? collectionStore.navigatedCollection
    : (collectionStore.navigatedCollection || collectionStore.selectedCollection);
  if (currentCollection) {
    const c = currentCollection;
    context.current_location = {
      type: c.type || 'collection',
      id: c.id,
      entity_id: c.id,
      name: c.name,
      path: c.collection_path || c.item_path || c.file_path || '',
      parent_id: c.parent_id || '',
    };
    context.here_scope = {
      source: 'here',
      entity_id: c.id,
      path: c.collection_path || c.item_path || c.file_path || '',
      recursive: !!commonStore.onlyAssets,
    };
  }

  const selected = stage.selectedItems?.length ? stage.selectedItems : (stage.selectedItem ? [stage.selectedItem] : []);
  context.selection = selected.map((item) => ({
    type: item.type,
    id: item.id,
    name: item.name,
    path: item.file_path || item.collection_path || '',
    parent_id: item.parent_id || '',
    parent_path: item.collection_path || '',
    collection_id: item.collection_id || '',
    extension: item.extension || '',
    metadata: {
      status_id: item.status_id || '',
      status: item.status_short_name || '',
      asset_type_id: item.asset_type_id || '',
      asset_type: item.asset_type_name || '',
      collection_type_id: item.collection_type_id || '',
      collection_type: item.collection_type_name || '',
      assignee_id: item.assignee_id || '',
      assignee: item.assignee_name || '',
      assignee_ids: item.assignee_ids || [],
      is_resource: !!item.is_resource,
    },
  }));

  if (!context.current_location && !context.selection.length && !context.active_view && !context.script_references.length) return '';
  return `[Context JSON]\n${JSON.stringify(context)}\n`;
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
    clearComposer();
    return;
  }
  if (shortcut?.localReply) {
    addMessage('user', rawInput);
    addMessage('assistant', shortcut.localReply);
    clearComposer();
    return;
  }
  const expanded = shortcut?.prompt ?? rawInput;
  const displayMessage = shortcut?.prompt ? rawInput : '';

  const turnIndex = nextTurnIndex();
  addMessage('user', rawInput, { turnIndex });
  const context = buildSelectionContext(activeScriptReferences());
  const messageContent = context + expanded;
  clearComposer();
  isProcessing.value = true;

  try {
    await AgentService.SendMessage(projectPath, messageContent, displayMessage, attachmentPath.value);
    attachmentPath.value = '';
  } catch (err) {
    addMessage('error', `${err}`);
    isProcessing.value = false;
    await loadChatHistory();
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
  addMessage('user', content, { turnIndex: lastUserMsg.turnIndex });
  isProcessing.value = true;

  try {
    await AgentService.RetryLastTurn(projectPath);
  } catch (err) {
    addMessage('error', `${err}`);
    isProcessing.value = false;
  }
};

const addMessage = (type, content, details = {}) => {
  messages.value.push({ id: messages.value.length + 1, type, content, ...details });
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
  messages.value.push({
    id: messages.value.length + 1,
    type: 'assistant',
    content: event.data,
    entities: recentResultEntities.value,
  });
  recentResultEntities.value = [];
  scrollToBottom();
};

const onAgentToolResult = (event) => {
  const data = event?.data?.data;
  if (!data) return;
  agentEntityCache.rememberCommandResult(data);
  const found = new Map(recentResultEntities.value.map((entity) => [`${entity.type}:${entity.id}`, entity]));
  const visit = (value) => {
    if (!value) return;
    if (Array.isArray(value)) return value.forEach(visit);
    if (typeof value !== 'object') return;
    const entity = value.entity && typeof value.entity === 'object' ? value.entity : value;
    if (entity.id && entity.type && entity.name && ['asset', 'collection', 'untracked_asset', 'untracked_collection'].includes(entity.type)) {
      found.set(`${entity.type}:${entity.id}`, { type: entity.type, id: entity.id, name: entity.name });
    }
    Object.values(value).forEach(visit);
  };
  visit(data);
  recentResultEntities.value = [...found.values()];
};

const onAgentCommandProgress = (event) => {
  const progress = event?.data;
  if (!progress) return;
  emitter.emit('progress-update', {
    title: 'Agent batch operation',
    message: progress.message || 'Applying local changes',
    percentage: progress.percentage || 0,
    current: progress.current || 0,
    total: progress.total || 1,
  });
};

const onAgentLocalChanges = (event) => {
  if (!event?.data?.requires_sync || !projectStore.activeProject) return;
  if (projectStore.activeProject.has_remote) {
    projectStore.activeProject.is_unsynced = true;
  }
  notificationStore.addNotification(
    'Agent changes saved locally',
    'Manual sync is required to share these changes.',
    'warning'
  );
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
const onAgentApplyFilter = async (event) => {
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
  if (typeof payload.show_untracked === 'boolean') await commonStore.setUntrackedVisibility(payload.show_untracked);
  if (typeof payload.show_tasks === 'boolean') commonStore.showTasks = payload.show_tasks;
  if (typeof payload.show_resources === 'boolean') commonStore.showResources = payload.show_resources;
  if (typeof payload.only_assets === 'boolean') commonStore.onlyAssets = payload.only_assets;
  if (typeof payload.only_collections === 'boolean') commonStore.onlyCollections = payload.only_collections;
  if (commonStore.onlyAssets) commonStore.onlyCollections = false;
  if (commonStore.onlyCollections) commonStore.onlyAssets = false;
  emitter.emit('refresh-browser');
};

// Clears all browser filters when the agent issues clear_browser_filter.
const onAgentClearFilter = () => {
  commonStore.resetFilters();
  emitter.emit('refresh-browser');
};

// watchers
watch(() => projectStore.activeProject?.uri, async (projectPath, previousProjectPath) => {
  saveComposerDraft(previousProjectPath);
  messages.value = [];
  scriptReferences.value = [];
  scriptReferencesLoadedFor.value = '';
  scriptReferencesLoadingFor.value = '';
  restoreComposerDraft(projectPath);
  await checkApiKeyStatus();
  await loadChatHistory();
  await nextTick();
  handleInput();
});

watch(composerMenuItems, () => { composerMenuIndex.value = 0; });
watch(currentMessage, () => saveComposerDraft());

// lifecycle hooks
onActivated(async () => {
  await checkApiKeyStatus();
});

onMounted(async () => {
  restoreComposerDraft(projectStore.activeProject?.uri);
  await checkApiKeyStatus();
  await loadChatHistory();
  await nextTick();
  handleInput();

  Events.On('agent-status', onAgentStatus);
  Events.On('agent-tool-start', onAgentToolStart);
  Events.On('agent-tool-result', onAgentToolResult);
  Events.On('agent-command-progress', onAgentCommandProgress);
  Events.On('agent-local-changes', onAgentLocalChanges);
  Events.On('agent-response', onAgentResponse);
  Events.On('agent-error', onAgentError);
  Events.On('agent-done', onAgentDone);
  Events.On('agent-cancelled', onAgentCancelled);
  Events.On('ignore-list-updated', onIgnoreListUpdated);
  Events.On('agent-reveal-in-browser', onAgentRevealInBrowser);
  Events.On('agent-apply-filter', onAgentApplyFilter);
  Events.On('agent-clear-filter', onAgentClearFilter);
  Events.On('files-dropped', onFilesDropped);
  window.addEventListener('resize', updateComposerMenuAnchor);
});

onUnmounted(() => {
  saveComposerDraft();
  Events.Off('agent-status');
  Events.Off('agent-tool-start');
  Events.Off('agent-tool-result');
  Events.Off('agent-command-progress');
  Events.Off('agent-local-changes');
  Events.Off('agent-response');
  Events.Off('agent-error');
  Events.Off('agent-done');
  Events.Off('agent-cancelled');
  Events.Off('ignore-list-updated');
  Events.Off('agent-reveal-in-browser');
  Events.Off('agent-apply-filter');
  Events.Off('agent-clear-filter');
  Events.Off('files-dropped');
  window.removeEventListener('resize', updateComposerMenuAnchor);
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

.console-editor {
  display: grid;
  width: 100%;
  min-height: 50px;
}

.console-editor > .console-input {
  grid-area: 1 / 1;
}

.console-input-highlight {
  pointer-events: none;
  white-space: pre-wrap;
  overflow-wrap: break-word;
  overflow: hidden;
}

.console-input-editor {
  position: relative;
  caret-color: var(--text);
  color: transparent;
  -webkit-text-fill-color: transparent;
}

.console-input-editor::placeholder {
  color: var(--text-muted);
  -webkit-text-fill-color: var(--text-muted);
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
  /* gap: 0.25rem; */
  min-height: 0;
  padding: 0.75rem 0.5rem;
  overflow-y: auto;
  /* background-color: crimson; */
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

.empty-help-command {
  border: 0;
  cursor: pointer;
}

:deep(.agent-inline-token),
.agent-inline-token {
  display: inline;
  padding: 0.08rem 0.35rem;
  border-radius: 6px;
  color: var(--text);
  background: var(--surface-4);
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 0.92em;
  font-weight: 600;
  line-height: inherit;
  box-decoration-break: clone;
  -webkit-box-decoration-break: clone;
}

:deep(.agent-inline-token-script),
.agent-inline-token-script {
  color: var(--accent-primary);
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

.msg-user-content {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  max-width: 80%;
  min-width: 0;
}

.msg-user-bubble {
  max-width: 100%;
  box-sizing: border-box;
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

.msg-message-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.25rem;
  min-height: 28px;
  padding-top: 0.125rem;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.12s ease-out;
}

.msg-user-content:hover .msg-message-actions,
.msg-user-content:focus-within .msg-message-actions {
  opacity: 0.75;
  pointer-events: auto;
}

.msg-message-actions:hover {
  opacity: 1;
}

.msg-message-actions :deep(.is-mini img) {
  width: 24px;
  height: 24px;
  min-width: 24px;
  min-height: 24px;
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

.msg-assistant-entities {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-top: 0.5rem;
}

.msg-entity-overflow,
.msg-assistant-text :deep(.msg-entity-overflow) {
  display: inline-flex;
  align-items: center;
  padding: 0.15rem 0.55rem;
  border-radius: var(--large-radius);
  color: var(--text-muted);
  background-color: var(--surface-2);
  line-height: 1.2;
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
