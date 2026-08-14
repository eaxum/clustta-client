<template>
  <div class="modal-container approval-modal" v-stop-propagation>
    <HeaderArea :title="title" :icon="getAppIcon('alert')" :showSearch="false" />

    <div class="general-container">
      <div class="risk-banner" :class="riskBannerClass">
        <img class="risk-icon small-icons" :src="getAppIcon('alert')" />
        <span>{{ riskBannerText }}</span>
      </div>

      <div v-if="summaryText" class="preview-summary">{{ summaryText }}</div>

      <div v-if="affectedItems.length" class="items-block">
        <div class="items-header">
          <span>{{ itemSectionTitle }}</span>
          <div v-if="isSelectable" class="selection-actions">
            <button type="button" @click="selectAll">{{ $t('modals.agentApproval.checkAll') }}</button>
            <button type="button" @click="unselectAll">{{ $t('modals.agentApproval.uncheckAll') }}</button>
          </div>
        </div>

        <div class="items-list">
          <div v-for="item in paginatedItems" :key="itemKey(item)" class="item-row">
            <template v-if="isTypeChange(item) && hasBeforeAfter(item)">
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
              <div v-if="operation === operationUpdate && !isCompactAction(item)" class="item-type-icon-slot">
                <img class="item-type-icon theme-icon" :src="getAppIcon(item.type_icon || (trackingLabel(item) === 'untracked' ? 'generic' : defaultTypeIcon(item)))" />
              </div>
              <div class="item-entity-icon-slot">
                <img class="item-entity-icon" :class="{ 'no-filter': hasResolvedAssetIcon(item), 'theme-icon': !hasResolvedAssetIcon(item) }"
                  :src="entityIcon(item)" />
              </div>
              <div class="item-content">
                <div class="item-line">
                  <template v-if="isCompactAction(item)">
                    <span class="item-name">{{ item.name || item.id || JSON.stringify(item) }}</span>
                    <span v-if="compactDetail(item)" class="compact-detail">{{ compactDetail(item) }}</span>
                  </template>
                  <template v-else-if="isRename(item)">
                    <span class="item-name old-name">{{ readableValue(item.before, item) }}</span>
                    <span class="change-arrow">→</span>
                    <span class="item-name">{{ readableValue(item.after, item) }}</span>
                  </template>
                  <template v-else-if="isAssignmentChange(item)">
                    <span class="item-name">{{ item.name || item.id || JSON.stringify(item) }}</span>
                    <span class="assignment-change">{{ isUnassign(item) ? $t('modals.agentApproval.actions.unassign') : $t('modals.agentApproval.actions.assignTo') }}</span>
                    <AssigneeItem v-if="!isUnassign(item) && assignmentAssignee(item).id"
                      class="assignment-assignee" v-stop-propagation
                      :name="assignmentAssignee(item).name"
                      :assigneeId="assignmentAssignee(item).id"
                      avatarColor="var(--surface-4)" />
                  </template>
                  <template v-else>
                    <span class="item-name">{{ item.name || item.id || JSON.stringify(item) }}</span>
                    <span v-if="hasBeforeAfter(item)" class="change-label">{{ changeLabel(item) }}</span>
                    <span v-if="hasBeforeAfter(item)">{{ readableValue(item.before, item) }}</span>
                    <span v-if="hasBeforeAfter(item)" class="change-arrow">→</span>
                    <span v-if="hasBeforeAfter(item)">{{ readableValue(item.after, item) }}</span>
                  </template>
                  <span v-if="item.type_name && !isRename(item) && !isAssignmentChange(item)" class="item-type-name">{{ item.type_name }}</span>
                  <span v-if="item.errors" class="item-error">{{ item.errors }}</span>
                </div>
              </div>
            </template>
            <div class="item-badges">
              <span class="operation-badge" :class="`badge-${operation}`">{{ operationBadge(item) }}</span>
              <span v-if="item.status && item.status !== 'skipped'" class="item-status">{{ item.status }}</span>
            </div>
            <span v-if="visibleWarning(item)" class="item-warning-indicator" tabindex="0"
              :aria-label="`Warning: ${visibleWarning(item)}`" v-tooltip="visibleWarning(item)">!</span>
            <CheckBox v-if="isSelectable" class="item-checkbox" :modelValue="isSelected(item)"
              :ariaLabel="`${isSelected(item) ? 'Exclude' : 'Include'} ${item.name || 'item'}`"
              @update:modelValue="setSelected(item, $event)" />
          </div>

        </div>

        <div v-if="totalPages > 1" class="pagination-controls">
          <button type="button" :disabled="currentPage === 1" @click="currentPage--">{{ $t('modals.agentApproval.previous') }}</button>
          <span>{{ $t('modals.agentApproval.page', { current: currentPage, total: totalPages }) }}</span>
          <button type="button" :disabled="currentPage === totalPages" @click="currentPage++">{{ $t('modals.agentApproval.next') }}</button>
        </div>
      </div>

      <div v-if="current?.preview?.notes && current.preview.notes.length" class="notes-block">
        <div v-for="(note, i) in current.preview.notes" :key="i" class="note-line">{{ note }}</div>
      </div>

      <div v-if="pendingCount > 1" class="queue-hint">{{ $t('modals.agentApproval.queueHint', { remaining: pendingCount - 1 }) }}</div>

      <div class="approval-actions">
        <GeneralButton :label="$t('common.deny')" :fullWidth="true" :buttonFunction="deny" :colored="false" />

        <GeneralButton :label="approveLabel" :fullWidth="true" :buttonFunction="approve" :colored="true"
          :isActive="canApprove" />
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
const operationCreate = 'create';
const operationUpdate = 'update';
const operationDelete = 'delete';
const operationExecute = 'execute';
const operationMembership = 'membership';

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
const operation = computed(() => current.value?.preview?.operation || operationUpdate);
const isSelectable = computed(() => Boolean(current.value?.preview?.selectable && itemCount.value > 0));
const approvalCount = computed(() => {
  if (isSelectable.value) return selectedCount.value;
  return Number(current.value?.preview?.counts?.changes || itemCount.value || 1);
});
const totalPages = computed(() => Math.max(1, Math.ceil(itemCount.value / pageSize)));
const paginatedItems = computed(() => {
  const start = (currentPage.value - 1) * pageSize;
  return affectedItems.value.slice(start, start + pageSize);
});

const title = computed(() => {
  if (!current.value) return t('common.confirm');

  const count = approvalCount.value;
  const subjectKey = String(current.value.preview?.subject || 'item').replaceAll(' ', '');
  const subject = t(`modals.agentApproval.subjects.${subjectKey}`, count);
  return t('modals.agentApproval.confirmItems', {
    action: t(`modals.agentApproval.actions.${toolActionKey()}`),
    count,
    subject,
  });
});

const itemSectionTitle = computed(() => {
  const labels = {
    [operationCreate]: 'itemsToCreate',
    [operationUpdate]: 'affectedItems',
    [operationDelete]: 'itemsToDelete',
    [operationExecute]: 'targetItems',
    [operationMembership]: 'affectedAccess',
  };
  return t(`modals.agentApproval.${labels[operation.value] || 'affectedItems'}`, { count: itemCount.value });
});

const riskBannerText = computed(() => {
  const keys = {
    [operationCreate]: 'createBanner',
    [operationUpdate]: 'updateBanner',
    [operationDelete]: 'deleteBanner',
    [operationExecute]: 'executeBanner',
    [operationMembership]: 'membershipBanner',
  };
  return t(`modals.agentApproval.${keys[operation.value] || 'updateBanner'}`);
});

const riskBannerClass = computed(() => `risk-${operation.value}`);
const summaryText = computed(() => {
  if (itemCount.value > 0 && operation.value !== operationMembership) return '';
  const summary = String(current.value?.preview?.summary || '');
  const args = current.value?.preview?.args || {};
  const contextKeys = ['email', 'role', 'role_name', 'user_id', 'studio_id', 'asset_type_id', 'collection_type_id'];
  const details = contextKeys
    .filter(key => args[key])
    .map(key => `${key.replaceAll('_', ' ')}: ${args[key]}`);
  return [summary, details.join(' · ')].filter(Boolean).join(' ');
});
const approveLabel = computed(() => {
  const keys = {
    [operationCreate]: 'createAction',
    [operationUpdate]: 'updateAction',
    [operationDelete]: 'deleteAction',
    [operationExecute]: 'executeAction',
    [operationMembership]: 'membershipAction',
  };
  return t(`modals.agentApproval.${keys[operation.value] || 'updateAction'}`);
});
const canApprove = computed(() => {
  if (current.value?.preview?.blocked) return false;
  return !isSelectable.value || selectedCount.value > 0;
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
  for (const key of ['name', 'status', 'asset_type', 'collection_type', 'assignment', 'tag_name', 'tag_id', 'dependency_name', 'dependency_id']) {
    if (parsed[key] !== undefined && parsed[key] !== '') return String(parsed[key]);
  }
  if (Object.prototype.hasOwnProperty.call(parsed, 'assignee')) {
    return parsed.assignee ? String(parsed.assignee) : 'Unassigned';
  }
  if (Array.isArray(parsed.tags)) return parsed.tags.length ? parsed.tags.join(', ') : 'No tags';
  if (Object.prototype.hasOwnProperty.call(parsed, 'is_resource')) {
    return parsed.is_resource ? 'Resource' : 'Task';
  }
  if (Object.prototype.hasOwnProperty.call(parsed, 'is_task')) {
    return parsed.is_task ? 'Task' : 'Resource';
  }
  if (parsed.path) return String(parsed.path).split(/[/\\]/).filter(Boolean).pop() || item?.name || '';
  if (parsed.deleted === true) return 'Deleted';
  if (parsed.parent_id !== undefined) return parsed.parent_id ? 'New collection' : 'Project root';
  return item?.name || '';
};

const changeLabel = (item) => {
  const action = String(item?.action || '').toLowerCase();
  if (action.includes('rename')) return 'Name';
  if (action.includes('assign') || action.includes('distribute')) return 'Assignment';
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
const isAssignmentChange = (item) => {
  const action = String(item?.action || '').toLowerCase();
  return action.includes('assign') || action.includes('distribute');
};

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

const hasChangeValue = (value) => {
  if (value === undefined || value === null || value === '') return false;
  if (typeof value !== 'string') return true;
  try {
    return JSON.parse(value) !== null;
  } catch {
    return true;
  }
};

const hasBeforeAfter = (item) => hasChangeValue(item?.before) && hasChangeValue(item?.after);

const isCompactAction = (item) => {
  if (operation.value !== operationUpdate) return true;
  return !hasBeforeAfter(item);
};

const toolActionKey = () => {
  const toolActions = {
    batch_rename: 'rename',
    batch_assign: 'assign',
    batch_unassign: 'unassign',
    batch_distribute: 'distribute',
    batch_change_status: 'changeStatus',
    batch_change_type: 'changeType',
    batch_move: 'move',
    batch_add_tags: 'addTags',
    batch_remove_tags: 'removeTags',
    batch_toggle_task_resource: 'changeKind',
    batch_add_dependency: 'addDependency',
    batch_remove_dependency: 'removeDependency',
    batch_update_asset_types: 'updateAssetType',
    batch_update_collection_types: 'updateCollectionType',
    dcc_open: 'open',
    dcc_render: 'render',
    dcc_export: 'export',
    dcc_run_script: 'runScript',
    dcc_run_python: 'runPython',
    dcc_set_settings: 'changeSettings',
    dcc_link_dependencies: 'linkDependencies',
    add_project_collaborator: 'addAccess',
    remove_project_collaborator: 'removeAccess',
    add_studio_collaborator: 'addAccess',
    change_studio_collaborator_role: 'changeRole',
    remove_studio_collaborator: 'removeAccess',
    remove_user: 'removeAccess',
  };
  const operationActions = {
    [operationCreate]: 'create',
    [operationUpdate]: 'update',
    [operationDelete]: 'delete',
    [operationExecute]: 'run',
    [operationMembership]: 'updateAccess',
  };
  return toolActions[current.value?.tool] || operationActions[operation.value] || 'update';
};

const operationBadge = (item) => {
  const action = String(item?.action || '').toLowerCase();
  const itemActions = [
    ['rename', 'rename'],
    ['unassign', 'unassign'],
    ['assign', 'assign'],
    ['distribute', 'distribute'],
    ['status', 'changeStatus'],
    ['type', 'changeType'],
    ['move', 'move'],
    ['tag', action.includes('remove') ? 'removeTags' : 'addTags'],
    ['depend', action.includes('remove') ? 'removeDependency' : 'addDependency'],
  ];
  const matchedAction = itemActions.find(([term]) => action.includes(term));
  const key = matchedAction?.[1] || toolActionKey();
  return t(`modals.agentApproval.actions.${key}`);
};

const compactDetail = (item) => {
  const after = changeObject(item?.after);
  if (operation.value === operationCreate) {
    return String(after.parent_name || after.collection_name || after.task_type_name || current.value?.preview?.subject || '');
  }
  if (operation.value !== operationUpdate) return '';
  for (const key of ['tag_name', 'tag_id', 'dependency_name', 'dependency_id']) {
    if (after[key]) return String(after[key]);
  }
  if (Object.prototype.hasOwnProperty.call(after, 'is_task')) return after.is_task ? 'Task' : 'Resource';
  return '';
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
  if (!req || !canApprove.value) return;
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
  min-width: min(840px, 90vw);
  max-width: 90vw;
}

.approval-modal > .general-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
  min-width: 0;
  max-width: none;
  color: var(--text);
}

.risk-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: var(--small-radius);
  font-size: 13px;
  width: 96%;
}

.risk-delete {
  background-color: color-mix(in oklch, var(--danger) 14%, transparent);
  color: var(--danger);
}

.risk-create,
.risk-update,
.risk-membership {
  background-color: color-mix(in oklch, var(--accent) 10%, transparent);
  color: var(--text);
}

.risk-execute {
  background-color: color-mix(in oklch, var(--warning) 12%, transparent);
  color: var(--warning);
}

.risk-icon {
  width: 16px;
  height: 16px;
}

.preview-summary {
  padding: 0 4px;
  color: var(--text-muted);
  font-size: 12px;
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
  min-width: 500px;
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
  align-self: center;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.operation-badge {
  flex: 0 0 auto;
  padding: 1px 5px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 500;
  text-transform: uppercase;
  white-space: nowrap;
}

.badge-create {
  background-color: rgba(34, 197, 94, 0.15);
  color: #4ade80;
}

.badge-delete {
  background-color: rgba(220, 50, 50, 0.15);
  color: #f87171;
}

.badge-update,
.badge-membership {
  background-color: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
}

.badge-execute {
  background-color: color-mix(in oklch, var(--warning) 15%, transparent);
  color: var(--warning);
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

.compact-detail {
  overflow: hidden;
  color: var(--text-muted);
  text-overflow: ellipsis;
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
