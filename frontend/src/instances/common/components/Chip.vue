<template>
  <span class="chip" :class="{ 'readonly': readonly }">
    <div v-if="useImage" class="chip-icon-container">
      <img class="chip-logo no-filter" :src="icon">
    </div>
    <ActionButton
      v-else
      :icon="icon"
      :isInactive="true"
      :showIcon="true"
      :showLabel="false"
    />
    <span class="chip-name">{{ label }}</span>
    <ActionButton
      v-if="!readonly"
      :icon="closeIcon"
      :buttonFunction="onRemove"
      :showIcon="true"
      :showLabel="false"
    />
  </span>
</template>

<script setup>
import { computed } from 'vue';
import { useIconStore } from '@/stores/icons';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

const iconStore = useIconStore();

const props = defineProps({
  icon: {
    type: String,
    required: true
  },
  label: {
    type: String,
    required: true
  },
  onRemove: {
    type: Function,
    default: () => {}
  },
  useImage: {
    type: Boolean,
    default: false
  },
  readonly: {
    type: Boolean,
    default: false
  }
});

const closeIcon = computed(() => iconStore.getAppIcon('close'));
</script>

<style scoped>
.chip {
  display: inline-flex;
  align-items: center;
  /* gap: 0.375rem; */
  background-color: var(--steel);
  border-radius: var(--large-radius);
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--white);
  transition: background-color 0.2s;
  padding: 0px;
  overflow: hidden;
}

.chip:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.chip-name {
  user-select: none;
}

.chip.readonly {
  padding-right: 0.75rem;
}

.chip-icon-container {
  padding: 0.3rem;
  align-items: center;
  justify-content: center;
  display: flex;
  min-width: 24px;
  min-height: 24px;
}

.chip-logo {
  width: 24px;
  height: 24px;
  object-fit: contain;
}
</style>
