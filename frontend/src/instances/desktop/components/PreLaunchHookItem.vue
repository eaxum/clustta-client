<template>
  <div class="hook-item-main">
    <div class="hook-item-spacer">
      <img class="hook-item-icon small-icons" :src="getAppIcon('launch')" />
    </div>

    <div class="hook-item-root">
      <div class="hook-item-container">
        <div class="hook-item-details">
          <div class="hook-item-heading">
            <span class="hook-item-name">{{ hook.name }}</span>
          </div>
          <div class="hook-item-summary">{{ summary }}</div>
        </div>

        <div class="hook-item-controls">
          <div class="hook-item-actions">
            <ActionButton :icon="getAppIcon('edit')" @click="onEdit(hook.id)" v-tooltip="$t('common.edit')" />
            <ActionButton :icon="getAppIcon('trash')" @click="onDelete(hook.id)" v-tooltip="$t('common.delete')" />
          </div>
          <div class="hook-item-toggle" @click.stop="onToggle(hook.id)">
            <ToggleSwitch :switchValueProp="hook.enabled" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import { useIconStore } from '@/stores/icons';

const props = defineProps({
  hook: { type: Object, required: true },
  onDelete: { type: Function, default: () => {} },
  onEdit: { type: Function, default: () => {} },
  onToggle: { type: Function, default: () => {} },
});

const iconStore = useIconStore();
const { t } = useI18n();

const summary = computed(() => {
  const details = [
    (props.hook.extensions || []).join(', '),
  ];
  const variableCount = props.hook.environment_variables?.length || 0;
  if (variableCount) details.push(t('settings.hookVariableCount', { count: variableCount }));
  details.push(props.hook.failure_policy === 'warn' ? t('settings.hookWarns') : t('settings.hookBlocks'));
  return details.filter(Boolean).join(' - ');
});

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);
</script>

<style scoped>
.hook-item-main {
  display: flex;
  gap: .2rem;
  width: 100%;
  box-sizing: border-box;
  align-items: flex-start;
  padding-left: .5rem;
  color: var(--text);
  background-color: var(--surface-2);
  border-radius: var(--large-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  overflow: hidden;
  transition: all .2s ease-out;
}

.hook-item-main:hover {
  background-color: var(--surface-3);
  border-radius: var(--small-radius);
  outline: 1px solid var(--surface-4);
}

.hook-item-spacer {
  width: 36px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.hook-item-icon {
  width: 20px;
  height: 20px;
  opacity: .6;
}

.hook-item-root {
  width: 100%;
  box-sizing: border-box;
  padding: .3rem 0 .3rem .3rem;
}

.hook-item-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 50px;
  gap: .5rem;
  padding: .2rem .4rem;
}

.hook-item-details {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: .15rem;
  padding: .2rem;
}

.hook-item-controls {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: .5rem;
  margin-left: auto;
  min-width: max-content;
}

.hook-item-heading {
  display: flex;
  align-items: center;
  gap: .5rem;
}

.hook-item-name {
  font-size: 14px;
  font-weight: 400;
}

.hook-item-summary {
  overflow: hidden;
  color: var(--text);
  font-size: 12px;
  font-weight: 300;
  opacity: .5;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hook-item-actions {
  display: flex;
  align-items: center;
  gap: .5rem;
  min-width: max-content;
  opacity: 0;
  pointer-events: none;
  transition: opacity .15s ease-out;
}

.hook-item-main:hover .hook-item-actions,
.hook-item-main:focus-within .hook-item-actions {
  opacity: 1;
  pointer-events: auto;
}

.hook-item-toggle {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  min-width: 52px;
}
</style>
