<template>
  <div class="modal-container approval-modal" v-stop-propagation>
    <HeaderArea :title="title" :icon="getAppIcon('alert')" :showSearch="false" />

    <div class="general-container">
      <div class="risk-banner risk-destructive">
        <img class="risk-icon small-icons" :src="getAppIcon('alert')" />
        <span>{{ $t('modals.agentApproval.destructiveBanner') }}</span>
      </div>

      <div v-if="affectedItems.length" class="items-block">
        <div class="items-header">
          <span>{{ $t('modals.agentApproval.affectedItems', { count: itemCount }) }}</span>
          <div class="selection-actions">
            <button type="button" @click="selectAll">Check all</button>
            <button type="button" @click="unselectAll">Uncheck all</button>
          </div>
        </div>

        <div class="items-list">
          <div v-for="item in paginatedItems" :key="itemKey(item)" class="item-row">
            <template v-if="isTypeChange(item)">
              <div class="item-type-icon-slot" v-tooltip="changeTypeName(item, 'before')">
                <img class="item-type-icon theme-icon" :src="getAppIcon(changeTypeIcon(item, 'before'))" />
              </div>
              <span class="change-arrow">→</span>
              <div class="item-type-icon-slot" v-tooltip="changeTypeName(item, 'after')">
                <img class="item-type-icon theme-icon" :src="getAppIcon(changeTypeIcon(item, 'after'))" />
              </div>
              <div class="item-entity-icon-slot">
                <img class="item-entity-icon" :class="{ 'no-filter': hasResolvedAssetIcon(item), 'theme-icon': !hasResolvedAssetIcon(item) }"
                  :src="entityIcon(item)" />
              </div>
              <div class="item-content">
                <div class="item-line">
                  <span class="item-name">{{ item.name || item.id || JSON.stringify(item) }}</span>
                  <span v-if="item.errors" class="item-error">{{ item.errors }}</span>
                </div>
              </div>
            </template>
            <template v-else>
              <div class="item-type-icon-slot">
                <img class="item-type-icon theme-icon" :src="getAppIcon(item.type_icon || (trackingLabel(item) === 'untracked' ? 'generic' : defaultTypeIcon(item)))" />
              </div>
              <div class="item-entity-icon-slot">
                <img class="item-entity-icon" :class="{ 'no-filter': hasResolvedAssetIcon(item), 'theme-icon': !hasResolvedAssetIcon(item) }"
                  :src="entityIcon(item)" />
              </div>
              <div class="item-content">
                <div class="item-line">
                  <template v-if="isRename(item)">
                    <span class="item-name old-name">{{ readableValue(item.before, item) }}</span>
                    <span class="change-arrow">→</span>
                    <span class="item-name">{{ readableValue(item.after, item) }}</span>
                  </template>
                  <template v-else-if="isAssignmentChange(item)">
                    <span class="item-name">{{ item.name || item.id || JSON.stringify(item) }}</span>
                    <span class="assignment-change">{{ isUnassign(item) ? 'Unassign' : 'Assign to' }}</span>
                    <AssigneeItem v-if="!isUnassign(item) && assignmentAssignee(item).id"
                      class="assignment-assignee" v-stop-propagation
                      :name="assignmentAssignee(item).name"
                      :assigneeId="assignmentAssignee(item).id"
                      avatarColor="var(--surface-4)" />
                  </template>
                  <template v-else>
                    <span class="item-name">{{ item.name || item.id || JSON.stringify(item) }}</span>
                    <span v-if="item.before || item.after" class="change-label">{{ changeLabel(item) }}</span>
                    <span v-if="item.before || item.after">{{ readableValue(item.before, item) }}</span>
                    <span v-if="item.before || item.after" class="change-arrow">→</span>
                    <span v-if="item.before || item.after">{{ readableValue(item.after, item) }}</span>
                  </template>
                  <span v-if="item.type_name && !isRename(item) && !isAssignmentChange(item)" class="item-type-name">{{ item.type_name }}</span>
                  <span v-if="item.errors" class="item-error">{{ item.errors }}</span>
                </div>
              </div>
            </template>
            <div class="item-badges">
              <span v-if="item.status && item.status !== 'skipped'" class="item-status">{{ item.status }}</span>
            </div>
            <span v-if="visibleWarning(item)" class="item-warning-indicator" tabindex="0"
              :aria-label="`Warning: ${visibleWarning(item)}`" v-tooltip="visibleWarning(item)">!</span>
            <CheckBox class="item-checkbox" :modelValue="isSelected(item)"
              :ariaLabel="`${isSelected(item) ? 'Exclude' : 'Include'} ${item.name || 'item'}`"
              @update:modelValue="setSelected(item, $event)" />
          </div>

        </div>

        <div v-if="totalPages > 1" class="pagination-controls">
          <button type="button" :disabled="currentPage === 1" @click="currentPage--">Previous</button>
          <span>Page {{ currentPage }} of {{ totalPages }}</span>
          <button type="button" :disabled="currentPage === totalPages" @click="currentPage++">Next</button>
        </div>
      </div>

      <div v-if="current?.preview?.notes && current.preview.notes.length" class="notes-block">
        <div v-for="(note, i) in current.preview.notes" :key="i" class="note-line">{{ note }}</div>
      </div>

      <div v-if="pendingCount > 1" class="queue-hint">{{ $t('modals.agentApproval.queueHint', { remaining: pendingCount - 1 }) }}</div>

      <div class="approval-actions">
        <GeneralButton :label="$t('common.deny')" :fullWidth="true" :buttonFunction="deny" :colored="false" />

        <GeneralButton :label="$t('common.approve')" :fullWidth="true" :buttonFunction="approve" :colored="true"
          :isActive="!current?.preview?.blocked && selectedCount > 0" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import CheckBox from '@/instances/common/components/CheckBox.vue';
import AssigneeItem from '@/instances/common/components/AssigneeItem.vue';

// services
import { AgentService } from '@/services';

// stores
import { useAgentApprovalStore } from '@/stores/agentApproval';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';

const { t } = useI18n();
const approvalStore = useAgentApprovalStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();

// refs
const resolvedAssetIcons = ref({});
const selectedKeys = ref(new Set());
const currentPage = ref(1);
const pageSize = 100;

// computed
const current = computed(() => approvalStore.current);

const pendingCount = computed(() => approvalStore.pendingCount);

const affectedItems = computed(() => {
  const items = current.value?.preview?.items;
  return Array.isArray(items) ? items.filter(item => !isNoopRename(item)) : [];
});

const itemCount = computed(() => {
  return affectedItems.value.length;
});

const selectedCount = computed(() => selectedKeys.value.size);
const totalPages = computed(() => Math.max(1, Math.ceil(itemCount.value / pageSize)));
const paginatedItems = computed(() => {
  const start = (currentPage.value - 1) * pageSize;
  return affectedItems.value.slice(start, start + pageSize);
});

const title = computed(() => {
  if (!current.value) return t('common.confirm');

  const tool = String(current.value.tool || '').toLowerCase();
  const count = selectedCount.value;
  const operations = [
    ['rename', 'Rename', 'item'],
    ['assign', 'Assign', 'task'],
    ['status', 'Change status for', 'item'],
    ['type', 'Change type for', 'item'],
    ['move', 'Move', 'item'],
    ['tag', 'Update tags for', 'item'],
    ['delete', 'Delete', 'item'],
  ];
  const operation = operations.find(([key]) => tool.includes(key));
  const [, verb, noun] = operation || ['', 'Update', 'item'];
  return `Confirm: ${verb} ${count} ${noun}${count === 1 ? '' : 's'}`;
});

const readableValue = (value, item) => {
  if (!value) return item?.name || '';
  let parsed = value;
  try {
    if (typeof value === 'string') parsed = JSON.parse(value);
  } catch (_) {
    return value;
  }
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return String(parsed ?? '');

  // Approval rows should describe the entity change, not expose its full
  // filesystem or database representation. Raw values remain available in
  // "View raw arguments".
  for (const key of ['name', 'status', 'asset_type', 'collection_type', 'assignment']) {
    if (parsed[key] !== undefined && parsed[key] !== '') return String(parsed[key]);
  }
  if (Object.prototype.hasOwnProperty.call(parsed, 'assignee')) {
    return parsed.assignee ? String(parsed.assignee) : 'Unassigned';
  }
  if (Array.isArray(parsed.tags)) return parsed.tags.length ? parsed.tags.join(', ') : 'No tags';
  if (Object.prototype.hasOwnProperty.call(parsed, 'is_resource')) {
    return parsed.is_resource ? 'Resource' : 'Task';
  }
  if (parsed.path) return String(parsed.path).split(/[/\\]/).filter(Boolean).pop() || item?.name || '';
  if (parsed.deleted === true) return 'Deleted';
  if (parsed.parent_id !== undefined) return parsed.parent_id ? 'New collection' : 'Project root';
  return item?.name || '';
};

const changeLabel = (item) => {
  const action = String(item?.action || '').toLowerCase();
  if (action.includes('rename')) return 'Name';
  if (action.includes('assign')) return 'Assignment';
  if (action.includes('status')) return 'Status';
  if (action.includes('type')) return 'Type';
  if (action.includes('move')) return 'Location';
  if (action.includes('tag')) return 'Tags';
  if (action.includes('resource') || action.includes('task')) return 'Kind';
  if (action.includes('depend')) return 'Dependency';
  if (action.includes('delete')) return 'State';
  return 'Change';
};

const isRename = (item) => String(item?.action || '').toLowerCase().includes('rename');
const isTypeChange = (item) => String(item?.action || '').toLowerCase().includes('type');
const isAssignmentChange = (item) => String(item?.action || '').toLowerCase().includes('assign');

const changeObject = (value) => {
  if (!value) return {};
  if (typeof value === 'object' && !Array.isArray(value)) return value;
  try {
    const parsed = JSON.parse(value);
    return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed : {};
  } catch {
    return {};
  }
};

const changeTypeName = (item, side) => {
  const value = changeObject(item?.[side]);
  const key = item?.kind === 'collection' ? 'collection_type' : 'asset_type';
  return String(value[key] || (side === 'before' ? item?.type_name : '') || 'Generic');
};

const changeTypeIcon = (item, side) => {
  const value = changeObject(item?.[side]);
  const key = item?.kind === 'collection' ? 'collection_type_icon' : 'asset_type_icon';
  return String(value[key] || (side === 'before' ? item?.type_icon : '') || defaultTypeIcon(item));
};

const isUnassign = (item) => String(item?.action || '').toLowerCase().includes('unassign');

const assignmentAssignee = (item) => {
  const after = changeObject(item?.after);
  return {
    id: String(after.assignee_id || '').trim(),
    name: String(after.assignee || after.assignment || '').trim(),
  };
};

const isNoopRename = (item) => {
  if (!isRename(item)) return false;
  return String(item?.warnings || '').toLowerCase().includes('already in the requested format');
};

const visibleWarning = (item) => String(item?.warnings || '')
  .split(';')
  .map(warning => warning.trim())
  .filter(warning => warning && warning.toLowerCase() !== 'case-only rename will use a temporary path')
  .join('; ');

const itemKey = (item) => `${item?.type || ''}:${item?.id || ''}`;

const isSelected = (item) => selectedKeys.value.has(itemKey(item));

const setSelected = (item, checked) => {
  const next = new Set(selectedKeys.value);
  if (checked) next.add(itemKey(item));
  else next.delete(itemKey(item));
  selectedKeys.value = next;
};

const selectAll = () => {
  selectedKeys.value = new Set(affectedItems.value.map(itemKey));
};

const unselectAll = () => {
  selectedKeys.value = new Set();
};

// methods
// getAppIcon resolves an icon path via the icon store.
const getAppIcon = (name) => iconStore.getAppIcon(name);

const trackingLabel = (item) => String(item?.type || '').startsWith('untracked_') ? 'untracked' : 'tracked';

const entityIcon = (item) => {
  if (item?.kind === 'collection' || String(item?.type || '').includes('collection')) {
    return getAppIcon('folder');
  }
  return resolvedAssetIcons.value[`${item?.type}:${item?.id}`] || getAppIcon('file');
};

const hasResolvedAssetIcon = (item) => Boolean(resolvedAssetIcons.value[`${item?.type}:${item?.id}`]);

const defaultTypeIcon = (item) => {
  return item?.kind === 'collection' || String(item?.type || '').includes('collection') ? 'folder' : 'generic';
};

// approve sends an allow decision to the backend and advances the queue.
const approve = async () => {
  const req = current.value;
  if (!req || req.preview?.blocked || selectedCount.value === 0) return;
  approvalStore.dequeueCurrent();
  try {
    await AgentService.ApproveToolCall(req.id, true, Array.from(selectedKeys.value));
  } catch (error) {
    console.error('AgentService.ApproveToolCall(approve) failed:', error);
  }
  closeIfEmpty();
};

// deny sends a refusal to the backend and advances the queue.
const deny = async () => {
  const req = current.value;
  if (!req) return;
  approvalStore.dequeueCurrent();
  try {
    await AgentService.ApproveToolCall(req.id, false, []);
  } catch (error) {
    console.error('AgentService.ApproveToolCall(deny) failed:', error);
  }
  closeIfEmpty();
};

// closeIfEmpty hides the modal once the approval queue is empty.
const closeIfEmpty = () => {
  if (approvalStore.pendingCount === 0) {
    modals.setModalVisibility('agentApprovalModal', false);
  }
};

// watchers
// Auto-close the modal when the queue drains externally (e.g. agent cancel).
watch(() => approvalStore.pendingCount, (count) => {
  if (count === 0) modals.setModalVisibility('agentApprovalModal', false);
});

watch(
  () => [current.value?.id, ...affectedItems.value.map(itemKey)],
  () => {
    currentPage.value = 1;
    selectAll();
  },
  { immediate: true },
);

watch(paginatedItems, async (items) => {
  const icons = {};
  await Promise.all((items || []).map(async (item) => {
    if (item?.kind !== 'asset' && !String(item?.type || '').includes('asset')) return;
    const extension = String(item.extension || '').toLowerCase().replace(/^\./, '');
    if (!extension) return;
    icons[`${item.type}:${item.id}`] = (await iconStore.getIcon(extension)) || '';
  }));
  resolvedAssetIcons.value = icons;
}, { immediate: true });

</script>

<style scoped>
.approval-modal {
  width: 840px;
  max-width: 90vw;
}

.general-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
  color: var(--text);
}

.risk-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: var(--small-radius);
  font-size: 13px;
}

.risk-destructive {
  background-color: color-mix(in oklch, var(--danger) 14%, transparent);
  color: var(--danger);
}

.risk-icon {
  width: 16px;
  height: 16px;
}

.notes-block {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 8px 12px;
  box-sizing: border-box;
  background-color: var(--surface-2);
  border-radius: var(--small-radius);
  font-size: 12px;
  color: var(--text-muted);
  width: 100%;
}

.note-line {
  /* font-family: var(--mono-font, monospace); */
  word-break: break-all;
}

.items-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
  align-self: stretch;
  box-sizing: border-box;
  width: 100%;
  min-width: 0;
  padding: 8px;
  border-radius: var(--large-radius);
  background-color: var(--surface-1);
  outline: 1px solid var(--surface-4);
  outline-offset: -1px;
  overflow: hidden;
}

.items-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.4px;
}

.selection-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.selection-actions button {
  border: 0;
  padding: 2px 6px;
  color: var(--text-muted);
  background: transparent;
  font: inherit;
  text-transform: none;
  letter-spacing: normal;
  cursor: pointer;
}

.selection-actions button:hover {
  color: var(--text);
  background-color: var(--hover);
  border-radius: var(--tiny-radius);
}

.selection-actions button:focus-visible {
  color: var(--text);
  outline: 2px solid var(--accent);
  outline-offset: 1px;
}

.items-list {
  max-height: 180px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
  box-sizing: border-box;
}

.items-list::-webkit-scrollbar {
  width: 4px;
}

.items-list::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-4);
}

.items-list::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.item-row {
  display: flex;
  align-items: center;
  /* gap: 10px; */
  width: 100%;
  box-sizing: border-box;
  padding: 0 8px;
  min-height: 50px;
  font-size: 12px;
  color: var(--text);
  background-color: var(--surface-2);
  outline: var(--transparent-line);
  outline-offset: -1px;
  border-radius: var(--large-radius);
  transition: all .2s ease-out;
  overflow: hidden;
}

.item-row:hover {
  background-color: var(--surface-3);
  border-radius: var(--small-radius);
  outline: 1px solid var(--surface-4);
}

.item-line {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
}

.item-status {
  padding: 1px 5px;
  border-radius: var(--small-radius);
  background: var(--surface-3);
  color: var(--text-muted);
  font-size: 10px;
}

.item-status {
  text-transform: lowercase;
}

.item-type-icon-slot,
.item-entity-icon-slot {
  display: grid;
  place-items: center;
  flex: 0 0 28px;
  width: 28px;
  height: 100%;
}

.item-entity-icon-slot {
  flex-basis: 30px;
  width: 30px;
}

.item-entity-icon {
  width: 20px;
  height: 20px;
}

.item-type-icon {
  width: 16px;
  height: 16px;
}

.theme-icon {
  opacity: .8;
  filter: invert(100%);
}

[data-theme="dark"] .theme-icon {
  opacity: 1;
  filter: none;
}

.item-content {
  flex: 1 1 auto;
  min-width: 0;
}

.item-type-name {
  color: var(--text-muted);
  font-size: 10px;
}

.item-badges {
  align-self: flex-start;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.item-checkbox {
  flex: 0 0 24px;
  margin-left: 8px;
}

.item-warning-indicator {
  display: inline-grid;
  place-items: center;
  flex: 0 0 18px;
  width: 18px;
  height: 18px;
  margin-left: 6px;
  border: 1px solid color-mix(in oklch, var(--warning) 70%, var(--border));
  border-radius: 50%;
  color: var(--warning);
  background-color: color-mix(in oklch, var(--warning) 12%, transparent);
  font-size: 11px;
  font-weight: 700;
  cursor: help;
}

.item-warning-indicator:focus-visible {
  outline: 2px solid var(--warning);
  outline-offset: 2px;
}

.change-arrow {
  flex: 0 0 auto;
  color: var(--text-muted);
}

.change-label {
  color: var(--text);
  font-family: inherit;
  font-weight: 600;
  min-width: 68px;
}

.assignment-change {
  color: var(--text-muted);
}

.assignment-assignee {
  flex: 0 1 auto;
  min-width: 0;
  width: auto;
  height: 28px;
  padding: 0;
  gap: 6px;
  background: transparent;
}

.assignment-assignee :deep(.assignee-list-item-name) {
  font-size: 12px;
  white-space: nowrap;
}

.item-error {
  font-size: 11px;
}

.item-error {
  color: var(--danger);
}

.item-name {
  overflow: hidden;
  text-overflow: ellipsis;
}

.old-name {
  opacity: .7;
  font-style: italic;
  text-decoration: line-through;
}

.pagination-controls {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  color: var(--text-muted);
  font-size: 11px;
}

.pagination-controls button {
  border: 1px solid var(--border);
  border-radius: var(--tiny-radius);
  padding: 3px 8px;
  color: var(--text);
  background-color: var(--surface-2);
  cursor: pointer;
}

.pagination-controls button:hover:not(:disabled) {
  border-color: var(--border-strong);
  background-color: var(--surface-3);
}

.pagination-controls button:disabled {
  opacity: .45;
  cursor: default;
}

.queue-hint {
  font-size: 12px;
  color: var(--text-muted);
  font-style: italic;
}

.approval-actions {
  display: flex;
  gap: 8px;
  width: 100%;
  justify-content: space-between;
}

.approval-actions :deep(.general-button) {
  color: var(--text);
  background-color: var(--surface-3);
}

.approval-actions :deep(.general-button:hover) {
  background-color: var(--surface-4);
}

.approval-actions :deep(.general-button.colored) {
  color: var(--accent-fg);
  background-color: var(--accent);
}

.approval-actions :deep(.general-button.colored:hover) {
  background-color: var(--accent-hover);
}

</style>
