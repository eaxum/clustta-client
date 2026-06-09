<template>
  <Badge variant="secondary" :class="['gap-1 pl-1 pr-1', { 'pr-2.5': readonly || isStatic }]" :style="chipStyle">
    <div v-if="useImage && !isStatic" class="flex items-center justify-center min-w-[24px] min-h-[24px] p-0.5">
      <img class="w-6 h-6 object-contain" :src="icon">
    </div>
    <ActionButton v-else-if="!isStatic" :icon="icon" :isInactive="true" :showIcon="true" :showLabel="false" />
    <span class="font-light select-none overflow-hidden text-ellipsis whitespace-nowrap min-w-0" :class="{ 'px-2 py-1': isStatic }">{{ label }}</span>
    <ActionButton v-if="!readonly && !isStatic" :icon="closeIcon" :buttonFunction="onRemove" :showIcon="true" :showLabel="false" />
  </Badge>
</template>

<script setup>
import { computed } from 'vue';
import { useIconStore } from '@/stores/icons';
import { Badge } from '@/components/ui/badge';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

const iconStore = useIconStore();

const props = defineProps({
  icon: {
    type: String,
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
const closeIcon = computed(() => iconStore.getAppIcon('close'));
</script>
