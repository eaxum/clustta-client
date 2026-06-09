<template>
  <Avatar :src="userPhoto || fallbackAvatar" :alt="assigneeId" size="sm" :style="{ backgroundColor: profileColor(assigneeId) }" />
</template>

<script setup>
import { computed } from 'vue';
import { Avatar } from '@/components/ui/avatar';
import { generateAvatar } from '@/lib/avatar';

const props = defineProps({
  userPhoto: {
    type: String,
    default: ''
  },
  assigneeId: {
    type: String,
    default: ''
  },
  avatarColor: {
    type: String,
    default: 'grey'
  }
});

const emit = defineEmits(['click']);

// Generates a DiceBear avatar based on assigneeId as fallback.
const fallbackAvatar = computed(() => {
  return generateAvatar(props.assigneeId);
});

// Generates profile background color from UUID.
const profileColor = (uuid) => {
  if (!uuid) return '#cccccc';
  const parts = uuid.split('-');
  return '#' + parts[0];
};
</script>

