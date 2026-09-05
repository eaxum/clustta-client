<template>
  <div class="progress-section">
    <div class="progress-header">
      <span v-if="showTitle" class="progress-title">{{ displayTitle }}</span>
      <span v-else class="progress-message" role="status">{{ displayMessage }}</span>
      <span class="progress-percentage" :class="{ 'success': variant === 'success' }">{{ Math.round(displayPercentage) }}%</span>
    </div>
    <div v-if="showTitle && displayMessage" class="progress-message" role="status">{{ displayMessage }}</div>
    <div class="progress-bar-wrapper">
      <ProgressBar :assetProgress="displayPercentage" />
    </div>
    <div v-if="showCount" class="progress-meta">
      <span>{{ progress.current }}/{{ progress.total }}</span>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import ProgressBar from '@/instances/common/components/ProgressBar.vue';

// stores
import { useNotificationStore } from '@/stores/notifications';

const notificationStore = useNotificationStore();
const { t } = useI18n();
const downloadPhaseKeys = {
  preparing: 'progress.projectDownload.preparing',
  receiving: 'progress.projectDownload.receiving',
  completed: 'progress.projectDownload.completed',
};

// props
const props = defineProps({
  showTitle: { type: Boolean, default: true },
  showCount: { type: Boolean, default: true },
  message: { type: String, default: '' },
  percentage: { type: Number, default: null },
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
const displayPercentage = computed(() => props.percentage ?? progress.value.percentage);

// Returns the title to display, using prop override or store title.
const displayTitle = computed(() => props.title || progress.value.title);
const phaseKey = computed(() => progress.value.operation === 'project-download'
  ? downloadPhaseKeys[progress.value.phase] : null);
const displayMessage = computed(() => {
  if (props.message) return props.message;
  if (!phaseKey.value) return progress.value.message;
  if (progress.value.phase !== 'receiving' || !progress.value.message?.trim()) return t(phaseKey.value);

  // Translate existing download callback messages until transfer details are structured.
  const previews = progress.value.message.match(/^Receiving previews (\d+)\/(\d+)$/);
  if (previews) return t('progress.projectDownload.previews', { current: previews[1], total: previews[2] });
  const transfer = progress.value.message.match(/^Receiving (.+)\/(.+)$/);
  if (transfer) return t('progress.projectDownload.transfer', { current: transfer[1], total: transfer[2] });
  return progress.value.message;
});
</script>

<style scoped>
.progress-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1rem;
  /* background-color: var(--surface-3); */
  /* border-radius: var(--normal-radius); */
  width: 100%;
  box-sizing: border-box;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: var(--text);
}

.progress-title {
  font-size: 15px;
  font-weight: 500;
}

.progress-percentage {
  font-size: 14px;
  font-weight: 600;
  color: var(--text);
}

.progress-percentage.success {
  color: rgb(67, 210, 67);
}

.progress-message {
  font-size: 13px;
  color: var(--text-muted);
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
  background-color: var(--surface-2);
}

.progress-meta {
  display: flex;
  justify-content: flex-end;
  font-size: 12px;
  color: var(--text-muted);
  opacity: 0.8;
}
</style>
