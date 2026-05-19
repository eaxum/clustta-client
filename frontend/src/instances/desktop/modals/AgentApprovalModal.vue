<template>
  <div class="modal-container approval-modal" v-stop-propagation>
    <HeaderArea :title="title" :icon="getAppIcon('alert')" :showSearch="false" />

    <div class="general-container">
      <div class="risk-banner risk-destructive">
        <img class="risk-icon small-icons" :src="getAppIcon('alert')" />
        <span>{{ $t('modals.agentApproval.destructiveBanner') }}</span>
      </div>

      <div class="summary-block">
        <div class="tool-name">{{ current?.tool }}</div>
        <div v-if="current?.preview?.summary" class="summary-text">{{ current.preview.summary }}</div>
      </div>

      <div v-if="current?.preview?.notes && current.preview.notes.length" class="notes-block">
        <div v-for="(note, i) in current.preview.notes" :key="i" class="note-line">{{ note }}</div>
      </div>

      <div v-if="affectedItems.length" class="items-block">
        <div class="items-header">{{ $t('modals.agentApproval.affectedItems', { count: itemCount }) }}</div>

        <div class="items-list">
          <div v-for="(item, i) in affectedItems.slice(0, 50)" :key="i" class="item-row">
            <span class="item-name">{{ item.name || item.id || JSON.stringify(item) }}</span>
          </div>

          <div v-if="affectedItems.length > 50" class="item-overflow">+{{ affectedItems.length - 50 }} more…</div>
        </div>
      </div>

      <details class="args-details">
        <summary>{{ $t('modals.agentApproval.viewArguments') }}</summary>
        <pre class="args-json">{{ formattedArgs }}</pre>
      </details>

      <div v-if="pendingCount > 1" class="queue-hint">{{ $t('modals.agentApproval.queueHint', { remaining: pendingCount - 1 }) }}</div>

      

      <div class="always-allow-row" @click="toggleAlwaysAllow">
        <ToggleSwitch :switchValueProp="alwaysAllow" />
        <span class="always-allow-label">{{ $t('modals.agentApproval.alwaysAllow') }}</span>
      </div>

      <div class="approval-actions">
        <GeneralButton :label="$t('common.deny')" :fullWidth="true" :buttonFunction="deny" :colored="false" />

        <GeneralButton :label="$t('common.approve')" :fullWidth="true" :buttonFunction="approve" :colored="true" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

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
const alwaysAllow = ref(false);

// computed
const current = computed(() => approvalStore.current);

const pendingCount = computed(() => approvalStore.pendingCount);

const title = computed(() => current.value ? t('modals.agentApproval.title', { tool: current.value.tool }) : t('common.confirm'));

const affectedItems = computed(() => {
  const items = current.value?.preview?.items;
  return Array.isArray(items) ? items : [];
});

const itemCount = computed(() => {
  const counts = current.value?.preview?.counts;
  if (counts && typeof counts === 'object') {
    return Object.values(counts).reduce((a, b) => a + (typeof b === 'number' ? b : 0), 0);
  }
  return affectedItems.value.length;
});

const formattedArgs = computed(() => {
  try {
    return JSON.stringify(current.value?.args ?? {}, null, 2);
  } catch {
    return '{}';
  }
});

// methods
// getAppIcon resolves an icon path via the icon store.
const getAppIcon = (name) => iconStore.getAppIcon(name);

// approve sends an allow decision to the backend and advances the queue.
const approve = async () => {
  const req = current.value;
  if (!req) return;
  approvalStore.dequeueCurrent();
  try {
    await AgentService.ApproveToolCall(req.id, true);
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
    await AgentService.ApproveToolCall(req.id, false);
  } catch (error) {
    console.error('AgentService.ApproveToolCall(deny) failed:', error);
  }
  closeIfEmpty();
};

// toggleAlwaysAllow flips the auto-approve preference and persists it.
const toggleAlwaysAllow = async () => {
  alwaysAllow.value = !alwaysAllow.value;
  try {
    await AgentService.SetAutoApproveDestructive(alwaysAllow.value);
  } catch (error) {
    console.error('AgentService.SetAutoApproveDestructive failed:', error);
    alwaysAllow.value = !alwaysAllow.value;
  }
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

// lifecycle
onMounted(async () => {
  try {
    alwaysAllow.value = await AgentService.GetAutoApproveDestructive();
  } catch {
    alwaysAllow.value = false;
  }
});
</script>

<style scoped>
.approval-modal {
  width: 540px;
  max-width: 90vw;
}

.general-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
  /* padding: 16px; */
  width: 100%;
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
  background-color: var(--danger-soft, rgba(220, 53, 69, 0.12));
  color: var(--danger, #dc3545);
}

.risk-icon {
  width: 16px;
  height: 16px;
}

.summary-block {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.tool-name {
  font-family: var(--mono-font, monospace);
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color);
}

.summary-text {
  color: var(--text-color);
  font-size: 14px;
  line-height: 1.4;
}

.notes-block {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 8px 12px;
  background-color: var(--light-steel-bg, rgba(0, 0, 0, 0.04));
  border-radius: var(--small-radius);
  font-size: 12px;
  color: var(--text-secondary);
}

.note-line {
  font-family: var(--mono-font, monospace);
  word-break: break-all;
}

.items-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.items-header {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.4px;
}

.items-list {
  max-height: 180px;
  overflow-y: auto;
  border: 1px solid var(--surface-4);
  border-radius: var(--small-radius);
  padding: 6px;
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
  padding: 3px 6px;
  font-size: 12px;
  color: var(--text-color);
}

.item-name {
  word-break: break-all;
}

.item-overflow {
  padding: 4px 6px;
  font-size: 11px;
  font-style: italic;
  color: var(--text-secondary);
}

.args-details {
  font-size: 12px;
  color: var(--text-secondary);
}

.args-details summary {
  cursor: pointer;
  user-select: none;
  padding: 4px 0;
}

.args-json {
  margin: 6px 0 0 0;
  padding: 8px;
  background-color: var(--light-steel-bg, rgba(0, 0, 0, 0.04));
  border-radius: var(--small-radius);
  font-family: var(--mono-font, monospace);
  font-size: 11px;
  max-height: 160px;
  overflow: auto;
  white-space: pre;
}

.args-json::-webkit-scrollbar {
  width: 4px;
}

.args-json::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-4);
}

.args-json::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.queue-hint {
  font-size: 12px;
  color: var(--text-secondary);
  font-style: italic;
}

.approval-actions {
  display: flex;
  gap: 8px;
  width: 100%;
  justify-content: space-between;
}

.always-allow-row {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
  padding: 4px 0;
}

.always-allow-label {
  font-size: 12px;
  color: var(--text-secondary);
}
</style>
