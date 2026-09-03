<template>
  <div class="changelog-item-container" @mouseenter="isHovered = true" @mouseleave="isHovered = false">
    <div class="changelog-item">
      <div class="changelog-item-meta">
        <img class="changelog-item-icon small-icons" :class="{ 'no-filter': resolvedIcon }" :src="itemIcon" />
        <div class="changelog-item-label">
          <div class="changelog-item-label-text">{{ displayText }}</div>
        </div>
        <span v-if="!isExpanded" class="changelog-change-badge" :class="'badge-' + item.change_type">{{ item.change_type }}</span>
      </div>

      <div class="changelog-item-actions">
        <ActionButton v-if="hasChildren && item.change_type === 'deleted'" :icon="getAppIcon('undo')" v-tooltip="$t('components.changeItem.restore')" :buttonFunction="() => $emit('restore', item.id)" :isDisabled="isLoading" />
        <ActionButton v-if="item.change_type !== 'deleted' && itemType !== 'other'" :icon="getAppIcon('file-search')" v-tooltip="$t('components.changeItem.goToItem')" :buttonFunction="() => $emit('find', item.id)" :isDisabled="isLoading" />
        <ActionButton v-if="hasChildren && item.change_type !== 'deleted' && item.change_type !== 'unchanged'" :icon="getAppIcon('undo')" v-tooltip="$t('components.changeItem.discard')" :buttonFunction="() => $emit('discard', item.id)" :isDisabled="isLoading" />
        <ActionButton v-if="!hasChildren && item.change_type === 'deleted'" :icon="getAppIcon('undo')" v-tooltip="$t('components.changeItem.restore')" :buttonFunction="() => $emit('restore', item.id)" :isDisabled="isLoading" />
        <ActionButton v-if="!hasChildren && item.change_type !== 'deleted' && item.change_type !== 'unchanged'" :icon="getAppIcon('undo')" v-tooltip="$t('components.changeItem.discard')" :buttonFunction="() => $emit('discard', item.id)" :isDisabled="isLoading" />
      </div>
      <ActionButton v-if="hasChildren" :icon="getAppIcon('chevron-right')" :class="{ 'chevron-expanded': isExpanded }" :buttonFunction="toggleChildren" />

    </div>

    <transition name="expand" appear>
      <div v-if="hasChildren" v-show="isExpanded" class="changelog-children">
        <template v-for="(child, index) in item.children" :key="child.id">
          <div v-if="index > 0" class="changelog-child-divider"></div>
          <div class="changelog-child">
            <img class="changelog-child-icon small-icons" :src="childIcon(child.source)" />
            <div class="changelog-child-label">{{ childDescription(child) }}</div>
            <span class="changelog-change-badge badge-child" :class="'badge-' + child.change_type">{{ child.change_type }}</span>
            <ActionButton :icon="getAppIcon('undo')" v-tooltip="child.change_type === 'deleted' ? $t('components.changeItem.restore') : $t('components.changeItem.discard')" :buttonFunction="() => $emit('undoChild', child)" :isDisabled="isLoading" />
          </div>
        </template>
      </div>
    </transition>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import utils from '@/services/utils';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// stores
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();
const { locale } = useI18n();

// props
const props = defineProps({
  item: { type: Object, required: true },
  itemType: { type: String, required: true },
  isLoading: { type: Boolean, default: false },
});

// emits
const emit = defineEmits(['find', 'discard', 'restore', 'undoChild']);

// refs
const isExpanded = ref(false);
const isHovered = ref(false);
const resolvedIcon = ref('');

// computed properties
const displayText = computed(() => {
  return props.item.name || props.item.id;
});

const hasChildren = computed(() => {
  return props.item.children && props.item.children.length > 0;
});

const itemIcon = computed(() => {
  if (resolvedIcon.value) return resolvedIcon.value;
  if (props.item.icon) return getAppIcon(props.item.icon);
  if (props.item.source === 'tag') return getAppIcon('tag');
  if (props.itemType === 'collection') return getAppIcon('folder');
  return getAppIcon('generic');
});

// methods
// Returns the formatted description for a child item.
const childDescription = (child) => {
  if (child.source === 'asset_checkpoint') return utils.formatDate(child.description, locale.value);
  return child.description;
};

// Returns the icon for a child item based on its source type.
const childIcon = (source) => {
  if (source === 'asset_checkpoint') return getAppIcon('checkpoint-stone');
  if (source === 'asset_dependency' || source === 'collection_dependency') return getAppIcon('dependency');
  if (source === 'asset_tag' || source === 'asset_checkpoint_tag') return getAppIcon('tag');
  if (source === 'collection_assignee') return getAppIcon('person');
  return getAppIcon('generic');
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Toggles the expanded state of child items.
const toggleChildren = () => {
  isExpanded.value = !isExpanded.value;
};

// Resolves the file icon for assets with an extension.
const resolveIcon = async () => {
  if (props.item.extension) {
    const ext = props.item.extension.startsWith('.') ? props.item.extension.substring(1) : props.item.extension;
    resolvedIcon.value = await iconStore.getIcon(ext) || '/file-icons/default.svg';
  } else {
    resolvedIcon.value = '';
  }
};

// watchers
watch(() => props.item.extension, resolveIcon);

// lifecycle hooks
onMounted(resolveIcon);
</script>

<style scoped>
@import "@/assets/desktop.css";

.badge-child {
  font-size: 9px;
}

.badge-added {
  background-color: rgba(34, 197, 94, 0.15);
  color: #4ade80;
}

.badge-deleted,
.badge-removed {
  background-color: rgba(220, 50, 50, 0.15);
  color: #f87171;
}

.badge-modified {
  background-color: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
}

.badge-new {
  background-color: rgba(34, 197, 94, 0.15);
  color: #4ade80;
}

.badge-unchanged {
  background-color: rgba(148, 163, 184, 0.15);
  color: #94a3b8;
}

.changelog-change-badge {
  font-size: 10px;
  font-weight: 500;
  padding: 1px 5px;
  border-radius: 4px;
  text-transform: uppercase;
  white-space: nowrap;
  flex-shrink: 0;
  margin-left: auto;
}

.changelog-child {
  display: flex;
  align-items: center;
  gap: .5rem;
  padding: .3rem .5rem .3rem 1.5rem;
  min-height: 28px;
}

.changelog-child-divider {
  height: 1px;
  margin: 0 .75rem 0 1.5rem;
  background-color: rgba(255, 255, 255, 0.06);
}

.changelog-child-icon {
  width: 16px;
  height: 16px;
  min-width: 16px;
  object-fit: contain;
  opacity: .5;
}

.changelog-child-label {
  font-size: 12px;
  font-weight: 300;
  color: var(--text);
  opacity: .7;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.changelog-children {
  width: 100%;
  /* background-color: rgba(255, 255, 255, 0.03); */
  /* border-bottom-left-radius: var(--large-radius);
  border-bottom-right-radius: var(--large-radius); */
  overflow: hidden;
}

.changelog-item {
  position: relative;
  cursor: auto;
  box-sizing: border-box;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0 .5rem;
  background-color: var(--surface-2);
  overflow: hidden;
}

.changelog-item-actions {
  display: flex;
  align-items: center;
  gap: .25rem;
  max-width: 0;
  opacity: 0;
  overflow: hidden;
  transform: translateX(.5rem);
  transition: max-width .2s ease-in-out, opacity .2s ease-out, transform .2s ease-out;
  /* background-color: #f87171; */
}

.changelog-item-container {
  position: relative;
  cursor: auto;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  color: var(--text);
  border-radius: var(--large-radius);
  overflow: hidden;
  min-height: max-content;
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--surface-3);
  transition: all .2s ease-in-out;
}

.changelog-item-container:hover {
  border-radius: var(--small-radius);
  background-color: var(--surface-3);
}

.changelog-item-container:hover .changelog-item-actions {
  /* max-width: 96px; */
  min-width: 68px;
  padding: .2rem;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  box-sizing: border-box;
  opacity: 1;
  transform: translateX(0);
}

.changelog-item-icon {
  width: 20px;
  height: 20px;
  min-width: 20px;
  object-fit: contain;
}

.changelog-item-label {
  overflow: hidden;
  width: 100%;
  display: flex;
  white-space: nowrap;
}

.changelog-item-label-text {
  font-size: 13px;
  font-weight: 300;
  color: var(--text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.changelog-item-meta {
  padding-left: .2rem;
  box-sizing: border-box;
  overflow: hidden;
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .5rem;
  width: 100%;
  min-height: 40px;
  min-width: 0;
}

.chevron-expanded {
  transform: rotate(90deg);
}

.expand-enter-active,
.expand-leave-active {
  transition: all .2s ease-in-out;
  max-height: 300px;
}

.expand-enter-from,
.expand-leave-to {
  max-height: 0;
  opacity: 0;
}
</style>
