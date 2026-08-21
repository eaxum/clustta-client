<template>
  <div class="preview-collection" :class="{ 'preview-collection-virtual': isVirtual }">
    <div class="collection-spacer" :class="{ 'collection-spacer-inactive': !hasChildren }">
      <span @click.stop="toggleExpand" class="single-action-button">
        <img class="small-icons collection-chevron" :class="{ 'collection-expanded': isExpanded }" :src="chevronIcon">
      </span>
    </div>

    <div class="preview-content">
      <span class="preview-name">{{ collection.name }}</span>
    </div>

    <div class="preview-meta">
      <span class="preview-type-label">{{ isVirtual ? 'Folder' : typeName }}</span>
    </div>

    <span v-if="!isVirtual" class="action-badge" :class="`action-${collection.action}`">{{ actionLabel }}</span>
    <CheckBox v-if="!isVirtual" :modelValue="isSelected" :disabled="!isActionable || !isSelectable"
      :ariaLabel="`${isSelected ? 'Exclude' : 'Include'} ${collection.name}`"
      @update:modelValue="emit('toggle-selection')" />

    <!-- <div v-if="childCount > 0" class="preview-meta">
      <span class="preview-count">{{ childCount }}</span>
    </div> -->
  </div>
</template>

<script setup>
// imports
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import utils from '@/services/utils';
import CheckBox from '@/instances/common/components/CheckBox.vue';

// stores
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();
const { t } = useI18n();

// props
const props = defineProps({
  childCount: { type: Number, default: 0 },
  collection: { type: Object, required: true },
  hasChildren: { type: Boolean, default: false },
  isExpanded: { type: Boolean, default: false },
  isSelected: { type: Boolean, default: false },
  isSelectable: { type: Boolean, default: true },
});

// emits
const emit = defineEmits(['toggle', 'toggle-selection']);

// computed
// Returns the chevron icon.
const chevronIcon = computed(() => {
  return iconStore.getAppIcon('chevron-down');
});

// Returns true if this is a virtual folder (directory structure node, not a Kitsu collection).
const isVirtual = computed(() => {
  return props.collection.action === 'virtual';
});

// Returns the type icon path.
const typeIcon = computed(() => {
  const iconName = props.collection.collection_type_icon || 'folder';
  return iconStore.getAppIcon(iconName);
});

// Returns the capitalized external type name.
const typeName = computed(() => {
  return utils.capitalizeStr(props.collection.external_type || 'Collection');
});

const isActionable = computed(() => props.collection.action === 'create' || props.collection.action === 'link');

const actionLabel = computed(() => {
  const labels = {
    create: t('kitsu.actionNew'),
    link: t('kitsu.actionLink'),
    skip: t('kitsu.actionNone'),
  };
  return labels[props.collection.action] || t('kitsu.actionNone');
});

// methods
// Toggles the expand state.
const toggleExpand = () => {
  if (props.hasChildren) {
    emit('toggle');
  }
};
</script>

<style scoped>
.preview-collection {
  display: flex;
  gap: 0.2rem;
  color: var(--text);
  align-items: center;
  padding: 0 0.5rem;
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  border-radius: var(--large-radius);
  background-color: var(--surface-2);
  outline: var(--transparent-line);
  outline-offset: -1px;
  transition: all 0.2s ease-out;
}

.preview-collection:hover {
  background-color: var(--surface-3);
  border-radius: var(--small-radius);
  outline: 1px solid var(--surface-4);
}

.preview-collection-virtual {
  opacity: 0.7;
}

.collection-spacer {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 24px;
}

.collection-spacer-inactive {
  opacity: 0.2;
  pointer-events: none;
}

.collection-chevron {
  transform: rotate(-90deg);
  transition: transform 0.2s ease-out;
}

.collection-expanded {
  transform: rotate(0deg);
}

.preview-type-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: min-content;
}

.preview-content {
  display: flex;
  align-items: center;
  flex: 1;
  overflow: hidden;
}

.preview-name {
  font-size: 13px;
  font-weight: 400;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.preview-meta {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 0.25rem;
}

.preview-count {
  font-size: 11px;
  font-weight: 500;
  color: var(--text-tertiary);
}

.preview-type-label {
  padding: 0.1rem 0.4rem;
  font-size: 11px;
  font-weight: 500;
  color: var(--text-tertiary);
  background-color: var(--surface-1);
  border-radius: var(--tiny-radius);
  text-transform: capitalize;
}

.action-badge {
  flex: 0 0 auto;
  padding: 2px 6px;
  border-radius: var(--tiny-radius);
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.action-create {
  color: #4ade80;
  background-color: rgba(34, 197, 94, 0.15);
}

.action-link {
  color: #60a5fa;
  background-color: rgba(59, 130, 246, 0.15);
}

.action-skip {
  color: var(--text-muted);
  background-color: var(--surface-3);
}
</style>
