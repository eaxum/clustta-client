<template>
  <div ref="previewItemRef" class="preview-virtua-item" :style="{ '--depth': depth }">
    <div class="preview-item-header" :style="{ height: `${itemHeight}px` }">
      <PreviewCollection v-if="item.type === 'collection'" :collection="item" :isSelected="isItemSelected" 
        :hasChildren="hasChildren" :isExpanded="isExpanded" :childCount="childCount" @toggle="toggleExpand" 
        @toggle-selection="handleToggleSelection" />
      <PreviewAsset v-else :asset="item" :isSelected="isItemSelected" @toggle-selection="handleToggleSelection" />
    </div>
    <template v-if="isExpanded && hasChildren">
      <div class="preview-item-children">
        <div class="indent-guide" :style="{ height: `${childrenHeight}px` }"></div>
        <div ref="childrenContainerRef" class="preview-children-container">
          <PreviewVirtuaItem v-for="(child, index) in itemChildren" :key="child.id" :item="child" 
            :depth="depth + 1" :itemHeight="itemHeight" :expandedItems="expandedItems" 
            :selectedItems="selectedItems" @toggle-expand="$emit('toggle-expand', $event)" 
            @toggle-selection="$emit('toggle-selection', $event)" />
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
// imports
import { computed, ref, watch, nextTick } from 'vue';

// components
import PreviewAsset from '@/instances/desktop/blocks/PreviewAsset.vue';
import PreviewCollection from '@/instances/desktop/blocks/PreviewCollection.vue';

// props
const props = defineProps({
  depth: { type: Number, default: 0 },
  expandedItems: { type: Set, default: () => new Set() },
  item: { type: Object, required: true },
  itemHeight: { type: Number, default: 36 },
  selectedItems: { type: Set, default: () => new Set() },
});

// emits
const emit = defineEmits(['toggle-expand', 'toggle-selection']);

// refs
const childrenContainerRef = ref(null);
const childrenHeight = ref(0);
const previewItemRef = ref(null);

// computed
// Returns the count of children.
const childCount = computed(() => {
  return props.item.children?.length || 0;
});

// Returns whether this item has children.
const hasChildren = computed(() => {
  return childCount.value > 0;
});

// Returns whether this item is expanded.
const isExpanded = computed(() => {
  return props.expandedItems.has(props.item.id);
});

// Returns whether this item is selected.
const isItemSelected = computed(() => {
  return props.selectedItems.has(props.item.id);
});

// Returns the children of this item.
const itemChildren = computed(() => {
  return props.item.children || [];
});

// methods
// Handles toggle selection.
const handleToggleSelection = (id) => {
  emit('toggle-selection', id);
};

// Toggles the expand state.
const toggleExpand = () => {
  emit('toggle-expand', props.item.id);
};

// Updates the children height for indent guide.
const updateChildrenHeight = async () => {
  await nextTick();
  if (childrenContainerRef.value) {
    childrenHeight.value = childrenContainerRef.value.getBoundingClientRect().height;
  }
};

// watchers
watch(() => isExpanded.value, async () => {
  if (isExpanded.value) {
    await updateChildrenHeight();
  }
});

watch(() => itemChildren.value, async () => {
  if (isExpanded.value) {
    await updateChildrenHeight();
  }
}, { deep: true });
</script>

<style scoped>
.preview-virtua-item {
  display: flex;
  flex-direction: column;
}

.preview-item-header {
  display: flex;
  align-items: center;
  width: 100%;
}

.preview-item-children {
  display: flex;
  flex-direction: row;
  position: relative;
}

.indent-guide {
  position: absolute;
  left: 8px;
  top: 0;
  width: 1px;
  background-color: hsl(var(--border));
  opacity: 0.3;
}

.preview-children-container {
  display: flex;
  flex-direction: column;
  gap: 2px;
  width: 100%;
  padding-top: 2px;
  padding-left: 16px;
}
</style>
