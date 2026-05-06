<template>
  <div class="assignee-list-item" :class="{ 'is-loading': isLoading }" @click="triggerAction">
    <ProfilePhoto :assigneeId="assigneeId" :userPhoto="userPhoto" :avatarColor="avatarColor" />
    <div class="assignee-list-item-name">{{ name }}</div>
    <div class="assignee-list-item-actions">
      <slot name="actions"></slot>
    </div>
    <div v-if="isLoading" class="assignee-loading-indicator">
      <CiLoading class="small-icons loading-icon" :size="20" />
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import ProfilePhoto from '@/instances/common/components/ProfilePhoto.vue'
import { CiLoading } from '@clustta/icons-vue';

const props = defineProps({
  name: {
    type: String,
    required: true
  },
  assigneeId: {
    type: String,
    required: true
  },
  userPhoto: {
    type: String,
    default: ''
  },
  avatarColor: {
    type: String,
    required: true
  },
  isLoading: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['click']);

const triggerAction = () => {
  if (props.isLoading) return;
  emit('click');
}
</script>

<style scoped>
.profile-picture {
  height: 24px;
  min-width: 24px;
  overflow: hidden;
  display: flex;
  align-items: center;
  border-radius: 24px;
}

.profile-img {
  width: 100%;
  height: 100%;
}

.assignee-list-item-name {
  font-family: 'Inter', sans-serif;
  /* font-weight: 100; */
  color: var(--white);
  font-size: 14px;
  display: flex;
  flex: 1;
  height: 100%;
  align-items: center;
  justify-content: flex-start;
}

.assignee-list-item {
  color: var(--white);
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

.assignee-list-item.is-loading {
  pointer-events: none;
  opacity: 0.7;
}

.assignee-loading-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
}

.loading-icon {
  animation: loadingRotate 0.8s linear infinite;
}

@keyframes loadingRotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>

