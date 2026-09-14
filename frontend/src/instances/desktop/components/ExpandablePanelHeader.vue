<template>
  <header class="panel-header">
    <div class="panel-title">
      <img class="small-icons" :src="icons.getAppIcon(icon)" alt="" />
      <span>{{ title }}</span>
      <span v-if="count !== ''" class="panel-count">[{{ count }}]</span>
    </div>
    <div class="panel-actions">
      <SearchBar v-if="filterPlaceholder" :modelValue="modelValue" :placeholder="filterPlaceholder"
        @update:modelValue="$emit('update:modelValue', $event)" />
      <slot />
      <ActionButton v-if="showMaximize" :icon="icons.getAppIcon(maximized ? 'arrow-minimize' : 'arrow-maximize')"
        v-tooltip="$t(maximized ? 'activity.restorePanel' : 'activity.maximizePanel')"
        :aria-label="$t(maximized ? 'activity.restorePanel' : 'activity.maximizePanel')"
        role="button" tabindex="0" @keydown.enter.prevent="$emit('toggle-maximize')" @keydown.space.prevent="$emit('toggle-maximize')"
        :allowDeactivate="true" :buttonFunction="() => $emit('toggle-maximize')" />
      <ActionButton :icon="icons.getAppIcon(closeIcon)" v-tooltip="$t('activity.closePanel')"
        role="button" tabindex="0" @keydown.enter.prevent="$emit('close')" @keydown.space.prevent="$emit('close')"
        :aria-label="$t('activity.closePanel')" :allowDeactivate="true" :buttonFunction="() => $emit('close')" />
    </div>
  </header>
</template>

<script setup>
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import SearchBar from '@/instances/desktop/components/SearchBar.vue';
import { useIconStore } from '@/stores/icons';

defineProps({
  title: { type: String, required: true },
  icon: { type: String, default: 'fetch' },
  closeIcon: { type: String, default: 'close' },
  count: { type: [String, Number], default: '' },
  modelValue: { type: String, default: '' },
  filterPlaceholder: { type: String, default: '' },
  maximized: { type: Boolean, default: false },
  showMaximize: { type: Boolean, default: true },
});
defineEmits(['update:modelValue', 'toggle-maximize', 'close']);
const icons = useIconStore();
</script>

<style scoped>
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
  gap: .5rem;
  padding: .5rem;
  background-color: var(--bg);
  color: var(--text);
  user-select: none;
  border-radius: var(--normal-radius);
}
.panel-title {
  display: flex;
  align-items: center;
  gap: .5rem;
  min-width: 0;
  font-weight: 500;
  font-size: 13px;
}
.panel-count {
  color: var(--text-muted);
  font-weight: 400;
  white-space: nowrap;
}
.panel-actions {
  display: flex;
  align-items: center;
  gap: .25rem;
}
.panel-actions :deep(.searchbar-container) {
  height: 28px;
  min-height: 28px;
  width: 180px;
}
.panel-actions :deep(.searchbar-input) {
  font-size: 12px;
  padding: 6px 8px;
}
</style>
