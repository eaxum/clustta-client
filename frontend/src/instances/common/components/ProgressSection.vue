<template>
  <div class="progress-section">
    <div class="progress-header">
      <span class="progress-title">{{ displayTitle }}</span>
      <span class="progress-percentage" :class="{ 'success': variant === 'success' }">{{ Math.round(progress.percentage) }}%</span>
    </div>
    <div class="progress-message">{{ progress.message }}</div>
    <div class="progress-bar-wrapper">
      <ProgressBar :assetProgress="progress.percentage" />
    </div>
    <div class="progress-meta">
      <span>{{ progress.current }}/{{ progress.total }}</span>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed } from 'vue';

// components
import ProgressBar from '@/instances/common/components/ProgressBar.vue';

// stores
import { useNotificationStore } from '@/stores/notifications';

const notificationStore = useNotificationStore();

// props
const props = defineProps({
  title: {
    type: String,
    default: '',
  },
  variant: {
    type: String,
    default: 'default',
    validator: (value) => ['default', 'success'].includes(value),
  },
});

// computed
// Returns the progress data from notification store.
const progress = computed(() => notificationStore.progress);

// Returns the title to display, using prop override or store title.
const displayTitle = computed(() => props.title || progress.value.title);
</script>

<style scoped>
.progress-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1rem;
  /* background-color: hsl(var(--accent)); */
  /* border-radius: var(--normal-radius); */
  width: 100%;
  box-sizing: border-box;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: hsl(var(--foreground));
}

.progress-title {
  font-size: 15px;
  font-weight: 500;
}

.progress-percentage {
  font-size: 14px;
  font-weight: 600;
  color: hsl(var(--foreground));
}

.progress-percentage.success {
  color: rgb(67, 210, 67);
}

.progress-message {
  font-size: 13px;
  color: hsl(var(--muted-foreground));
  opacity: 0.9;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.progress-bar-wrapper {
  position: relative;
  width: 100%;
  height: 0.2rem;
  border-radius: 999px;
  overflow: hidden;
  background-color: hsl(var(--muted));
}

.progress-meta {
  display: flex;
  justify-content: flex-end;
  font-size: 12px;
  color: hsl(var(--muted-foreground));
  opacity: 0.8;
}
</style>
