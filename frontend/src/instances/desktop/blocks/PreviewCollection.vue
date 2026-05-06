<template>
  <div class="preview-collection" :class="{ 'preview-collection-selected': isSelected, 'preview-collection-virtual': isVirtual }">
    <div class="collection-spacer" :class="{ 'collection-spacer-inactive': !hasChildren }">
      <span @click.stop="toggleExpand" class="single-action-button">
        <img class="small-icons collection-chevron" :class="{ 'collection-expanded': isExpanded }" :src="chevronIcon">
      </span>
    </div>

    <div class="preview-type-icon">
      <img class="small-icons" :src="typeIcon" v-tooltip="typeName">
    </div>

    <div class="preview-content" @click="console.log(collection)" >
      <span class="preview-name">{{ collection.name }}</span>
    </div>

    <div class="preview-meta">
      <span class="preview-type-label">{{ isVirtual ? 'Folder' : typeName }}</span>
    </div>

    <!-- <div v-if="childCount > 0" class="preview-meta">
      <span class="preview-count">{{ childCount }}</span>
    </div> -->
  </div>
</template>

<script setup>
// imports
import { computed } from 'vue';
import utils from '@/services/utils';
import { CiChevronDown } from '@clustta/icons-vue';
import { resolveIcon } from '@/lib/icon-map';

// stores
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();

// props
const props = defineProps({
  childCount: { type: Number, default: 0 },
  collection: { type: Object, required: true },
  hasChildren: { type: Boolean, default: false },
  isExpanded: { type: Boolean, default: false },
  isSelected: { type: Boolean, default: false },
});

// emits
const emit = defineEmits(['toggle']);

// computed
// Returns the chevron icon.
const chevronIcon = computed(() => {
  return iconStore.CiChevronDown;
});

// Returns true if this is a virtual folder (directory structure node, not a Kitsu collection).
const isVirtual = computed(() => {
  return props.collection.action === 'virtual';
});

// Returns the type icon path.
const typeIcon = computed(() => {
  const iconName = props.collection.collection_type_icon || 'folder';
  return iconStore.resolveIcon(iconName);
});

// Returns the capitalized external type name.
const typeName = computed(() => {
  return utils.capitalizeStr(props.collection.external_type || 'Collection');
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
  color: var(--white);
  align-items: center;
  padding: 0 0.5rem;
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  border-radius: var(--large-radius);
  background-color: var(--dark-steel);
  outline: var(--transparent-line);
  outline-offset: -1px;
  transition: all 0.2s ease-out;
}

.preview-collection:hover {
  background-color: var(--steel);
  border-radius: var(--small-radius);
  outline: 1px solid var(--light-steel);
}

.preview-collection-selected {
  background-color: var(--blue-steel);
}

.preview-collection-selected:hover {
  background-color: var(--solid-blue-steel);
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
  padding: 0.1rem 0.4rem;
  background-color: var(--black-steel);
  border-radius: var(--tiny-radius);
  margin-right: 0.25rem;
}

.preview-count {
  font-size: 11px;
  font-weight: 500;
  color: var(--text-tertiary);
}

.preview-type-label {
  font-size: 11px;
  font-weight: 500;
  color: var(--text-tertiary);
  text-transform: capitalize;
}
</style>
