<template>
  <div
    class="column-item"
    :class="{
      'column-item-editing': editing,
      'column-item-selected': selected,
    }"
    :aria-selected="selected"
    @click="handleClick"
    @dblclick="handleDoubleClick"
  >
    <RenameInput
      v-if="editing"
      v-model="editableName"
      :originalValue="item.name || ''"
      :placeholder="renamePlaceholder"
      @click.stop
      @confirm="emit('rename', $event)"
      @cancel="emit('cancelRename')"
    />

    <template v-else>
      <div class="column-item-meta">
        <img
          class="column-item-icon small-icons"
          :class="{ 'no-filter': hasResolvedIcon }"
          :src="displayIcon"
          @error="handleIconError"
        />
        <span class="column-item-name" v-tooltip="tooltip">{{ displayName }}</span>
      </div>

      <div class="column-item-actions">
        <ActionButton
          v-if="assetStateIcon"
          class="column-item-state-icon"
          :icon="getAppIcon(assetStateIcon)"
          :isInactive="!isAssetStateActionable"
          :noFilter="item.file_status === 'normal'"
          :useAlert="isAlertState"
          :useDanger="item.type === 'untracked_asset'"
          :buttonFunction="handleStateAction"
          v-tooltip="item.file_status || item.type"
        />
        <img
          v-if="isCollection"
          class="small-icons column-item-chevron"
          :src="getAppIcon('chevron-right')"
        />
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import RenameInput from '@/instances/desktop/components/RenameInput.vue';
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';
import { usePlatformStore } from '@/stores/platform';

const props = defineProps({
  editing: { type: Boolean, default: false },
  item: { type: Object, required: true },
  renaming: { type: Boolean, default: false },
  selected: { type: Boolean, default: false },
});

const emit = defineEmits(['cancelRename', 'dblclick', 'rename', 'select', 'stateAction']);

const commonStore = useCommonStore();
const iconStore = useIconStore();
const platformStore = usePlatformStore();
const { t } = useI18n();
const editableName = ref(props.item.name || '');
const resolvedIcon = ref('');
let iconLoadVersion = 0;

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);
const itemType = computed(() => props.item.type || props.item.item_type || '');
const isCollection = computed(() => itemType.value.includes('collection') || !!props.item.collection_type_id);
const extension = computed(() => isCollection.value ? '' : props.item.extension || '');
const explicitIcon = computed(() => {
  if (isCollection.value) {
    return getAppIcon(props.item.collection_type_icon || props.item.icon || 'folder');
  }
  return props.item.icon || '';
});
const fallbackIcon = computed(() => {
  return getAppIcon(isCollection.value ? 'folder' : props.item.asset_type_icon || 'file');
});
const displayIcon = computed(() => resolvedIcon.value || explicitIcon.value || fallbackIcon.value);
const hasResolvedIcon = computed(() => !!(resolvedIcon.value || explicitIcon.value));
const tooltip = computed(() => {
  return props.item.file_path || props.item.collection_path || props.item.asset_path || props.item.name || '';
});
const displayName = computed(() => {
  if (commonStore.showFullPath) return tooltip.value;
  if (isCollection.value || commonStore.hideExtensions || !extension.value) return props.item.name || '';
  if ((props.item.name || '').toLowerCase().endsWith(extension.value.toLowerCase())) return props.item.name;
  return `${props.item.name || ''}${extension.value}`;
});
const renamePlaceholder = computed(() => {
  return isCollection.value ? t('placeholders.collectionName') : t('placeholders.assetName');
});
const assetStateIcon = computed(() => {
  if (isCollection.value) return '';
  if (props.item.type === 'untracked_asset') return 'plus-stone';
  if (platformStore.isWeb) return 'arrow-down-ramp';

  const stateIcons = {
    fetchable: 'fetch',
    missing: 'alert',
    modified: 'plus-stone',
    normal: 'circle-check-go',
    outdated: 'circle-check',
  };
  return stateIcons[props.item.file_status] || '';
});
const isAlertState = computed(() => {
  return props.item.file_status === 'modified' || props.item.file_status === 'outdated';
});
const isAssetStateActionable = computed(() => {
  return (platformStore.isWeb && props.item.type === 'asset')
    || props.item.type === 'untracked_asset'
    || props.item.file_status === 'modified'
    || props.item.file_status === 'outdated'
    || props.item.file_status === 'fetchable';
});

const handleClick = () => {
  if (props.editing || props.renaming) return;
  emit('select', props.item);
};

const handleDoubleClick = () => {
  if (props.editing || props.renaming) return;
  emit('dblclick', props.item);
};

const handleStateAction = () => {
  if (!isAssetStateActionable.value) return;
  emit('stateAction', props.item);
};

const handleIconError = (event) => {
  event.target.src = '/file-icons/default.svg';
};

const loadIcon = async () => {
  const loadVersion = ++iconLoadVersion;
  resolvedIcon.value = '';
  if (explicitIcon.value || !extension.value) return;

  const iconPath = await iconStore.getIcon(extension.value.toLowerCase().replace(/^\./, ''));
  if (loadVersion === iconLoadVersion) {
    resolvedIcon.value = iconPath || '/file-icons/default.svg';
  }
};

watch(() => props.editing, (editing) => {
  if (editing) editableName.value = props.item.name || '';
});
watch(() => `${extension.value}:${explicitIcon.value}`, loadIcon, { immediate: true });
</script>

<style scoped>
@import "@/assets/desktop.css";

.column-item {
  display: flex;
  width: 100%;
  min-height: 40px;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  box-sizing: border-box;
  padding: 0 .5rem;
  overflow: hidden;
  cursor: pointer;
  color: var(--text);
  background-color: var(--surface-2);
  border-radius: var(--large-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  transition: background-color .2s ease-out, border-radius .2s ease-out;
}

.column-item:hover {
  background-color: var(--surface-3);
  border-radius: var(--small-radius);
}

.column-item-selected,
.column-item-selected:hover {
  background-color: var(--selected-soft);
}

.column-item-editing {
  cursor: text;
}

.column-item-meta {
  display: flex;
  min-width: 0;
  flex: 1;
  align-items: center;
  gap: .5rem;
  overflow: hidden;
}

.column-item-icon {
  width: 20px;
  height: 20px;
  min-width: 20px;
  object-fit: contain;
}

.column-item-name {
  overflow: hidden;
  font-size: 13px;
  font-weight: 300;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.column-item-actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: .25rem;
}

.column-item-chevron {
  flex: 0 0 auto;
  opacity: .65;
}

.column-item-state-icon {
  flex: 0 0 auto;
  padding: 0;
  opacity: 1;
}
</style>
