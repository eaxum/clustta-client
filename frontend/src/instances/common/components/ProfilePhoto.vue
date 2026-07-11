<template>
    <div class="profile-picture" :style="{ backgroundColor: displayString ? avatarColor : profileColor(assigneeId) }">
      <span v-if="displayString" class="profile-string">{{ displayString }}</span>
      <img v-else-if="userPhoto" class="profile-img" :src="userPhoto">
      <img v-else class="profile-img" :src="fallbackAvatar">
    </div>
  </template>
  
<script setup>
import { computed } from 'vue';
import { generateAvatar } from '@/lib/avatar';
import { useUserStore } from '@/stores/users';

const userStore = useUserStore();
    
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
    },
    displayString: {
    type: String,
    default: ''
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
  
  <style scoped>
  .profile-picture {
    height: 24px;
    min-width: 24px;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 24px;
    opacity: 1;
  }
  
  .profile-img {
    width: 100%;
    height: 100%;
  }

  .profile-string {
    color: var(--text);
    font-size: 11px;
    font-weight: 600;
    line-height: 1;
  }
  
  .assignee-list-item-name {
    font-family: 'Inter', sans-serif;
    font-weight: 100;
    color: white;
    font-size: 14px;
    display: flex;
    flex: 1;
    height: 100%;
    align-items: center;
    justify-content: flex-start;
  }
  
  .assignee-list-item {
    color: white;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: .5rem;
    width: 96%;
    height: 40px;
    overflow: hidden;
    padding: 0 .2rem;
  }
  
  .assignee-list-item:hover {
    background-color: #ffffff15;
    border-radius: 10px;
  }
  
  .assignee-list-item:last-child {
    border-bottom: 0px;
  }
  
  .assignee-list-item:hover>*:last-child {
    opacity: 1;
    visibility: visible;
    transition: opacity 0.2s ease-in-out;
    display: flex;
  }
  
  .assignee-list-item-actions {
    display: none;
    opacity: 0;
    visibility: hidden;
  }
  </style>

