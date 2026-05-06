<template>
  <span class="chip" :class="{ 'readonly': readonly || isStatic }" :style="chipStyle">
    <div v-if="useImage && !isStatic" class="chip-icon-container">
      <img class="chip-logo no-filter" :src="icon">
    </div>
    <ActionButton
      v-else-if="!isStatic"
      :icon="icon"
      :isInactive="true"
      :showIcon="true"
      :showLabel="false"
    />
    <span class="chip-name" :class="{ 'chip-name-static': isStatic }">{{ label }}</span>
    <ActionButton
      v-if="!readonly && !isStatic"
      :icon="CiClose"
      :buttonFunction="onRemove"
      :showIcon="true"
      :showLabel="false"
    />
  </span>
</template>

<script setup>
import { computed } from 'vue';
import { CiClose } from '@clustta/icons-vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

const props = defineProps({
  icon: {
    type: [String, Object, Function],
    default: ''
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
  },
  isStatic: {
    type: Boolean,
    default: false
  },
  color: {
    type: String,
    default: ''
  }
});

const chipStyle = computed(() => props.color ? { 'background-color': props.color } : {});
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
  max-width: 100%;
}

.chip:hover {
  /* background-color: rgba(255, 255, 255, 0.1); */
  background-color: var(--light-steel);
  outline: var(--transparent-line);
}

.chip-name {
  font-weight: 300;
  user-select: none;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.chip.readonly {
  padding-right: 0.75rem;
}

.chip-name-static {
  padding: 0.25rem 0.5rem;
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
