<template>
  <section class="column-pane" :aria-label="title">
    <header class="column-pane-header" :class="{ 'column-pane-header-active': active }" v-tooltip="title">
      <span>{{ title }}</span>
    </header>

    <div class="column-pane-scroll" role="listbox">
      <ListSkeleton v-if="loading" :itemHeight="40" />

      <button v-else-if="error" class="column-pane-state column-pane-retry" type="button" @click="emit('retry')">
        <span>{{ $t('columns.loadFailed') }}</span>
        <span class="column-pane-retry-label">{{ $t('columns.retry') }}</span>
      </button>

      <div v-else-if="!items.length" class="column-pane-state">
        {{ $t('columns.empty') }}
      </div>

      <div
        v-for="item in items"
        v-else
        :key="getBrowserItemKey(item)"
        class="column-pane-row"
        :class="itemMainClass(item)"
        role="option"
        @click.stop
        @dblclick.stop
        @contextmenu.stop.prevent="selectForContextMenu($event, item)"
      >
        <ColumnItem
          :item="item"
          :editing="editingKey === getBrowserItemKey(item)"
          :renaming="renamingKey === getBrowserItemKey(item)"
          :selected="selectedKey === getBrowserItemKey(item)"
          @cancelRename="emit('cancelRename')"
          @dblclick="emit('open', item)"
          @rename="emit('rename', item, $event)"
          @select="emit('select', item)"
          @stateAction="emit('stateAction', item)"
        />
      </div>
    </div>
  </section>
</template>

<script setup>
import ColumnItem from '@/instances/desktop/components/ColumnItem.vue';
import ListSkeleton from '@/instances/desktop/components/ListSkeleton.vue';
import { getBrowserItemKey } from '@/lib/browserTree';

defineProps({
  active: { type: Boolean, default: false },
  editingKey: { type: String, default: '' },
  error: { type: [Error, String, Object], default: null },
  items: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
  renamingKey: { type: String, default: '' },
  selectedKey: { type: String, default: '' },
  title: { type: String, default: '' },
});

const emit = defineEmits(['cancelRename', 'contextmenu', 'open', 'rename', 'retry', 'select', 'stateAction']);

const isCollection = (item) => {
  const itemType = item.type || item.item_type || '';
  return itemType.includes('collection') || !!item.collection_type_id;
};

const selectForContextMenu = (event, item) => {
  emit('select', item);
  emit('contextmenu', event, item);
};

const itemMainClass = (item) => {
  return isCollection(item) ? 'collection-item-main' : 'asset-item-main';
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.column-pane {
  display: flex;
  min-width: 0;
  height: 100%;
  box-sizing: border-box;
  flex-direction: column;
  overflow: hidden;
  padding: .35rem;
  background-color: color-mix(in oklch, var(--surface-1) 82%, transparent);
  border-radius: var(--very-large-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.column-pane-header {
  display: flex;
  min-height: 32px;
  align-items: center;
  box-sizing: border-box;
  padding: 0 .8rem;
  overflow: hidden;
  color: var(--text-muted);
  background-color: var(--surface-2);
  border-radius: var(--very-large-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  font-size: var(--font-size-compact);
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.column-pane-header-active {
  color: var(--text);
  background-color: var(--selected-soft);
}

.column-pane-header span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.column-pane-scroll {
  display: flex;
  min-height: 0;
  flex: 1;
  flex-direction: column;
  gap: .35rem;
  overflow-y: auto;
  padding: .5rem 0;
  scrollbar-color: var(--surface-4) transparent;
  scrollbar-width: thin;
}

.column-pane-scroll :deep(.virtua-skeleton-container) {
  flex: 1;
}

.column-pane-scroll::-webkit-scrollbar {
  width: 6px;
}

.column-pane-scroll::-webkit-scrollbar-thumb {
  background-color: var(--surface-4);
  border-radius: var(--large-radius);
}

.column-pane-row {
  width: 100%;
}

.column-pane-state {
  display: flex;
  min-height: 5rem;
  flex: 1;
  align-items: center;
  justify-content: center;
  gap: .4rem;
  padding: 1rem;
  color: var(--text-muted);
  background: transparent;
  border: 0;
  font: inherit;
  text-align: center;
}

.column-pane-retry {
  cursor: pointer;
  flex-direction: column;
}

.column-pane-retry-label {
  color: var(--accent);
}

</style>
