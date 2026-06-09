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
const chevronIcon = iconStore.getAppIcon('chevron-right');
const collapseIcon = iconStore.getAppIcon('chevron-up');
const expandIcon = iconStore.getAppIcon('chevron-down');
const loadingIcon = iconStore.getAppIcon('loading');

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
  border-radius: var(--radius);
  background-color: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  cursor: pointer;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
  overflow: hidden;
}

.option-card:hover {
  border-color: hsl(var(--ring));
  box-shadow: 0 1px 3px rgb(0 0 0 / 0.1);
}

.option-card-selected {
  border-color: hsl(var(--primary));
  box-shadow: 0 0 0 1px hsl(var(--primary));
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
  border: 2px solid hsl(var(--border));
  display: flex;
  align-items: center;
  justify-content: center;
  transition: border-color 0.15s ease;
}

.option-radio-selected {
  border-color: hsl(var(--primary));
}

.option-radio-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: hsl(var(--primary));
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
  gap: 0.25rem;
  min-width: 0;
}

.option-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: hsl(var(--foreground));
  line-height: 1.25;
}

.option-description {
  font-size: 0.8rem;
  color: hsl(var(--muted-foreground));
  font-weight: 400;
  line-height: 1.35;
}

.option-action {
  display: flex;
  align-items: center;
  opacity: 0;
  transition: opacity .3s ease;
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
