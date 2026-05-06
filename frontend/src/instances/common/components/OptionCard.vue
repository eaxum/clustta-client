<template>
  <div class="option-card" :class="{ 'option-card-selected': selected, 'option-card-expanded': isExpanded }" @click="handleClick">
    <div class="option-card-main">

      <div v-if="selectable" class="option-radio" :class="{ 'option-radio-selected': selected }">
        <div v-if="selected" class="option-radio-dot"></div>
      </div>

      <div v-if="icon" class="option-icon-container">
        <img :src="icon" class="large-icons option-icon" />
      </div>

      <div class="option-content">
        <div class="option-title display-font">{{ title }}</div>
        <div class="option-description">{{ description }}</div>
      </div>

      <div class="option-action">
        <ActionButton v-if="loading" :icon="loadingIcon" :isLoading="true" :showLabel="false" :noFilter="true" />
        <div v-else-if="expandable && selected" @click.stop="isExpanded = !isExpanded">
          <ActionButton :isInactive="true" :icon="isExpanded ? expandIcon : collapseIcon" />
        </div>
        <ActionButton v-else :isInactive="true" :icon="chevronIcon" />
      </div>

    </div>

    <div v-if="expandable && isExpanded && selected" class="option-expand" @click.stop>
      <slot name="details"></slot>
    </div>
  </div>
</template>

<script setup>
// imports
import { ref, watch } from 'vue';
import { CiChevronDown, CiChevronRight, CiChevronUp, CiLoading } from '@clustta/icons-vue';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// stores
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

// props
const props = defineProps({
  icon: { type: String, default: '' },
  title: { type: String, required: true },
  description: { type: String, default: '' },
  selectable: { type: Boolean, default: false },
  selected: { type: Boolean, default: false },
  expandable: { type: Boolean, default: false },
  loading: { type: Boolean, default: false },
});

// emits
const emit = defineEmits(['select']);

// refs
const isExpanded = ref(false);

// computed icons
const chevronIcon = iconStore.CiChevronRight;
const collapseIcon = iconStore.CiChevronUp;
const expandIcon = iconStore.CiChevronDown;
const loadingIcon = iconStore.CiLoading;

// methods

// Handles clicking the card.
const handleClick = () => {
  if (props.selectable) {
    if (!props.selected) {
      isExpanded.value = false;
      emit('select');
    }
  } else {
    emit('select');
  }
};

// watchers
watch(() => props.selected, (val) => {
  if (!val) isExpanded.value = false;
});
</script>

<style scoped>
.option-card {
  display: flex;
  flex-direction: column;
  border-radius: var(--very-large-radius);
  background-color: var(--midnight-steel);
  outline: var(--transparent-line);
  outline-offset: -1px;
  cursor: pointer;
  transition: border-color 0.2s, background-color 0.2s, border-radius 0.2s;
  overflow: hidden;
}

.option-card:hover {
  border-color: var(--grape);
  border-radius: var(--large-radius);
  box-shadow: 0 0px 4px rgba(0, 0, 0, 0.1);
}

.option-card-selected {
  outline: 1px solid var(--grape);
  border-radius: var(--large-radius);
}

.option-card-main {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.25rem;
}

.option-radio {
  width: 16px;
  height: 16px;
  min-width: 16px;
  border-radius: 50%;
  border: 2px solid var(--light-steel);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: border-color 0.2s;
}

.option-radio-selected {
  border-color: var(--grape);
}

.option-radio-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: var(--grape);
}

.option-icon-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 40px;
  min-height: 40px;
  max-width: 40px;
  max-height: 40px;
}

.option-icon {
  width: 32px;
  height: 32px;
  opacity: .5;
}

.option-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  min-width: 0;
}

.option-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--white);
  line-height: 120%;
}

.option-description {
  font-size: 0.8rem;
  color: var(--white);
  opacity: 0.55;
  font-weight: 300;
  line-height: 130%;
}

.option-action {
  display: flex;
  align-items: center;
  opacity: 0;
  transition: opacity .5s;
}

.option-card:hover .option-action {
  opacity: 1;
}

.option-card-selected .option-action {
  opacity: 1;
}

.option-expand {
  padding: 0 1.25rem 1rem;
  cursor: default;
}
</style>
