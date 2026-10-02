<template>
  <div class="dependency-graph-node" :class="{
    'dependency-graph-node-collection': isCollection,
    'dependency-graph-node-conflict': data.hasConflict,
  }">
    <Handle v-if="data.hasIncoming" class="dependency-node-handle dependency-node-handle-input"
      id="input" type="target" :position="Position.Left" />
    <Handle v-if="data.hasOutgoing" class="dependency-node-handle dependency-node-handle-output"
      id="output" type="source" :position="Position.Right" />

    <div class="dependency-node-primary">
      <img class="dependency-node-icon small-icons" :class="{ 'no-filter': hasResolvedIcon }" :src="displayIcon">
      <span class="dependency-node-name" v-tooltip="data.path || data.name">{{ data.name }}</span>
      <div class="dependency-node-version">
        <DependencySelector v-if="data.dependencyEdge" :edge="data.dependencyEdge"
          :ownerAssetId="data.dependencyEdge.asset_id" :editable="data.canEditSelector"
          @updated="emit('selectorUpdated')" />
        <span v-else-if="data.versionLabel" class="dependency-node-version-label">
          {{ data.versionLabel }}
        </span>
      </div>
      <ActionButton v-if="data.canAdd" :icon="getAppIcon('plus-circle')"
        v-tooltip="$t('components.virtualNode.addDependency')" @click="emit('add', data)" />
      <div v-if="isCollection || data.canRemove" class="dependency-node-actions">
        <ActionButton v-if="isCollection" :icon="getAppIcon('file-search')"
          v-tooltip="navigationTooltip" @click="emit('navigate', data)" />
        <ActionButton v-if="data.canRemove" :icon="getAppIcon('minus-circle')"
          v-tooltip="$t('components.virtualNode.remove')" @click="emit('remove', data)" />
      </div>
    </div>

    <div v-if="!isCollection" class="dependency-node-secondary">
      <button v-if="primaryAssignee" class="dependency-node-assignee" type="button"
        :disabled="!data.canAssign" v-tooltip="primaryAssigneeName" @click="emit('assign', data, $event)">
        <img v-if="primaryAssignee.photo" class="dependency-node-avatar" :src="primaryAssignee.photo">
        <img v-else class="dependency-node-avatar" :src="generateAvatar(primaryAssignee.id)">
      </button>
      <ActionButton v-else-if="data.canAssign" :icon="getAppIcon('person-plus')"
        v-tooltip="$t('blocks.assignAsset')" @click="emit('assign', data, $event)" />

      <span v-if="data.statusLabel" class="dependency-node-status"
        :style="{ backgroundColor: data.statusColor }">{{ data.statusLabel }}</span>

      <span v-if="data.warning" class="dependency-node-warning" v-tooltip="data.warning">
        {{ data.hasConflict ? $t('components.dependencyGraph.conflict') : $t('components.dependencyGraph.needsAttention') }}
      </span>

      <div class="dependency-node-actions">
        <ActionButton :icon="getAppIcon('file-search')" v-tooltip="navigationTooltip"
          @click="emit('navigate', data)" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Handle, Position } from '@vue-flow/core';
import { generateAvatar } from '@/lib/avatar';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DependencySelector from '@/instances/desktop/components/DependencySelector.vue';
import { useIconStore } from '@/stores/icons';
import { useUserStore } from '@/stores/users';

const props = defineProps({
  data: { type: Object, required: true },
});

const emit = defineEmits(['add', 'assign', 'navigate', 'remove', 'selectorUpdated']);
const iconStore = useIconStore();
const userStore = useUserStore();
const { t } = useI18n();
const resolvedExtensionIcon = ref('');

const getAppIcon = (name) => iconStore.getAppIcon(name);
const getUserName = (user) => `${user.first_name || ''} ${user.last_name || ''}`.trim() || user.email || t('common.unknown');

const primaryAssignee = computed(() => {
  if (!props.data.assigneeId) return null;
  return userStore.getUserData(props.data.assigneeId) || {
    id: props.data.assigneeId,
    first_name: props.data.assigneeName,
  };
});
const primaryAssigneeName = computed(() => primaryAssignee.value ? getUserName(primaryAssignee.value) : '');
const isCollection = computed(() => props.data.entityType === 'collection');
const navigationTooltip = computed(() => t(props.data.entityType === 'collection' ? 'menus.goToCollection' : 'menus.goToAsset'));
const displayIcon = computed(() => resolvedExtensionIcon.value || props.data.icon || getAppIcon('file'));
const hasResolvedIcon = computed(() => !!resolvedExtensionIcon.value);

const loadExtensionIcon = async () => {
  resolvedExtensionIcon.value = '';
  if (props.data.entityType === 'collection' || !props.data.extension) return;
  const extension = String(props.data.extension).toLowerCase().replace(/^\./, '');
  resolvedExtensionIcon.value = await iconStore.getIcon(extension) || '';
};

watch(() => `${props.data.entityId}:${props.data.extension}`, loadExtensionIcon, { immediate: true });
</script>

<style scoped>
@import "@/assets/desktop.css";

.dependency-graph-node {
  position: relative;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: .45rem;
  min-width: 290px;
  max-width: 340px;
  padding: .65rem;
  border-radius: var(--large-radius);
  background-color: var(--surface-2);
  color: var(--text);
  outline: var(--transparent-line);
  outline-offset: -1px;
  transition: border-radius .2s ease-out, background-color .2s ease-out, outline-color .2s ease-out;
}

.dependency-graph-node:hover {
  border-radius: var(--small-radius);
  background-color: var(--surface-3);
  outline: 1px solid var(--surface-4);
}

.dependency-graph-node-conflict {
  outline-color: var(--danger);
}

.dependency-graph-node-collection {
  gap: 0;
  padding-top: .55rem;
  padding-bottom: .55rem;
}

:deep(.dependency-node-handle) {
  width: 7px;
  height: 7px;
  border: 1px solid var(--surface-1);
  border-radius: 50%;
  background-color: var(--text-muted);
}

:deep(.dependency-node-handle-input) {
  left: -16px;
}

:deep(.dependency-node-handle-output) {
  right: -16px;
}

.dependency-node-primary,
.dependency-node-secondary {
  display: flex;
  align-items: center;
  gap: .45rem;
  min-width: 0;
}

.dependency-node-icon {
  width: 22px;
  height: 22px;
  flex-shrink: 0;
  object-fit: contain;
}

.dependency-node-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  font-size: 14px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dependency-node-version {
  display: flex;
  flex-shrink: 0;
  max-width: 130px;
  overflow: hidden;
}

.dependency-node-version-label,
.dependency-node-warning {
  padding: .18rem .4rem;
  border-radius: var(--small-radius);
  background-color: var(--surface-3);
  color: var(--text-muted);
  font-size: 10px;
  font-weight: 600;
  white-space: nowrap;
}

.dependency-node-status {
  min-width: 48px;
  padding: .28rem .4rem;
  border-radius: var(--normal-radius);
  color: black;
  font-size: 10px;
  font-weight: 700;
  text-align: center;
  text-transform: uppercase;
  white-space: nowrap;
}

.dependency-node-warning {
  color: var(--danger);
}

.dependency-node-secondary {
  min-height: 24px;
}

.dependency-node-assignee {
  display: flex;
  align-items: center;
  gap: .3rem;
  min-width: 0;
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  font: inherit;
  font-size: 11px;
}

.dependency-node-assignee:disabled {
  cursor: default;
}

.dependency-node-avatar {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  object-fit: cover;
}

.dependency-node-actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: .25rem;
  max-width: 0;
  margin-left: auto;
  overflow: hidden;
  opacity: 0;
  transform: translateX(.5rem);
  transition: max-width .2s ease-in-out, opacity .2s ease-out, transform .2s ease-out;
}

.dependency-graph-node:hover .dependency-node-actions {
  width: max-content;
  max-width: none;
  overflow: visible;
  opacity: 1;
  transform: translateX(0);
}
</style>
