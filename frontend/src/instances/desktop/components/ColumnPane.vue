<template>
  <section class="column-pane" :aria-label="title">
    <header class="column-pane-header" v-tooltip="title">
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
        <AssetItem
          :item="item"
          :icon="getItemIcon(item)"
          :fallbackIcon="getFallbackIcon(item)"
          :hideExtension="commonStore.hideExtensions"
          :showFullPath="commonStore.showFullPath"
          :selected="selectedKey === getBrowserItemKey(item)"
          :tooltip="getTooltip(item)"
          :showBadge="false"
          :clickable="true"
          variant="compact"
          @click="emit('select', item)"
          @dblclick="emit('open', item)"
        >
          <template #persistent>
            <ActionButton
              v-if="getAssetStateIcon(item)"
              class="column-pane-state-icon"
              :icon="getAppIcon(getAssetStateIcon(item))"
              :isInactive="!isAssetStateActionable(item)"
              :noFilter="item.file_status === 'normal'"
              :useAlert="isAlertState(item)"
              :useDanger="item.type === 'untracked_asset'"
              :buttonFunction="() => emit('stateAction', item)"
              v-tooltip="item.file_status || item.type"
            />
            <img v-if="isCollection(item)" class="small-icons column-pane-chevron" :src="getAppIcon('chevron-right')" />
          </template>
        </AssetItem>
      </div>
    </div>
  </section>
</template>

<script setup>
import AssetItem from '@/instances/desktop/components/AssetItem.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ListSkeleton from '@/instances/desktop/components/ListSkeleton.vue';
import { getBrowserItemKey } from '@/lib/browserTree';
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';
import { usePlatformStore } from '@/stores/platform';

defineProps({
  error: { type: [Error, String, Object], default: null },
  items: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
  selectedKey: { type: String, default: '' },
  title: { type: String, default: '' },
});

const emit = defineEmits(['contextmenu', 'open', 'retry', 'select', 'stateAction']);

const commonStore = useCommonStore();
const iconStore = useIconStore();
const platformStore = usePlatformStore();

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

const isCollection = (item) => {
  const itemType = item.type || item.item_type || '';
  return itemType.includes('collection') || !!item.collection_type_id;
};

const getFallbackIcon = (item) => {
  if (isCollection(item)) return getAppIcon(item.collection_type_icon || 'folder');
  return getAppIcon(item.asset_type_icon || 'file');
};

const getItemIcon = (item) => {
  if (isCollection(item)) return getAppIcon(item.collection_type_icon || item.icon || 'folder');
  return item.icon || '';
};

const getTooltip = (item) => {
  return item.file_path || item.collection_path || item.asset_path || item.name || '';
};

const getAssetStateIcon = (item) => {
  if (isCollection(item)) return '';
  if (item.type === 'untracked_asset') return 'plus-stone';
  if (platformStore.isWeb) return 'arrow-down-ramp';

  const stateIcons = {
    fetchable: 'fetch',
    missing: 'alert',
    modified: 'plus-stone',
    normal: 'circle-check-go',
    outdated: 'circle-check',
  };
  return stateIcons[item.file_status] || '';
};

const isAlertState = (item) => {
  return item.file_status === 'modified' || item.file_status === 'outdated';
};

const isAssetStateActionable = (item) => {
  return (platformStore.isWeb && item.type === 'asset')
    || item.type === 'untracked_asset'
    || item.file_status === 'modified'
    || item.file_status === 'outdated'
    || item.file_status === 'fetchable';
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
  min-height: 40px;
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

.column-pane-chevron {
  flex: 0 0 auto;
  opacity: .65;
}

.column-pane-state-icon {
  flex: 0 0 auto;
  padding: 0;
  opacity: 1;
}
</style>
