<template>
  <div class="changelog-item-container" @mouseenter="isHovered = true" @mouseleave="isHovered = false">
    <div class="changelog-item">
      <div class="changelog-item-meta">
        <img class="changelog-item-icon small-icons" :src="itemIcon" />
        <div class="changelog-item-label">
          <div class="changelog-item-label-text">{{ item.name || item.id }}</div>
        </div>
        <span class="changelog-change-badge" :class="'badge-' + item.change_type">{{ item.change_type }}</span>
      </div>

      <div class="changelog-item-actions">
        <ActionButton v-if="item.change_type !== 'deleted'" :icon="getAppIcon('file-search')" v-tooltip="'Go to Item'" :buttonFunction="() => $emit('find', item.id)" :isDisabled="isLoading" />
        <ActionButton :icon="getAppIcon('revert')" v-tooltip="'Discard'" :buttonFunction="() => $emit('discard', item.id)" :isDisabled="isLoading" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref } from 'vue';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// stores
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();

// props
const props = defineProps({
  item: { type: Object, required: true },
  itemType: { type: String, required: true },
  isLoading: { type: Boolean, default: false },
});

// emits
const emit = defineEmits(['find', 'discard']);

// refs
const isHovered = ref(false);

// computed properties
const itemIcon = computed(() => {
  if (props.itemType === 'entity') return getAppIcon('folder');
  return getAppIcon('generic');
});

// methods
// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);
</script>

<style scoped>
@import "@/assets/desktop.css";

.changelog-item-container {
  position: relative;
  cursor: auto;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  color: var(--white);
  border-radius: var(--large-radius);
  overflow: hidden;
  min-height: max-content;
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--dark-steel);
  transition: all .2s ease-in-out;
}

.changelog-item-container:hover {
  border-radius: var(--normal-radius);
  background-color: var(--steel);
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
  overflow: hidden;
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
  color: var(--white);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.changelog-change-badge {
  font-size: 10px;
  font-weight: 500;
  padding: 1px 5px;
  border-radius: 4px;
  text-transform: uppercase;
  white-space: nowrap;
  flex-shrink: 0;
}

.badge-deleted {
  background-color: rgba(220, 50, 50, 0.15);
  color: #f87171;
}

.badge-modified {
  background-color: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
}

.changelog-item-actions {
  display: none;
}

.changelog-item-container:hover .changelog-item-actions {
  display: flex;
}
</style>
