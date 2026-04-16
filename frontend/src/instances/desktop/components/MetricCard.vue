<template>
  <div class="metric-card" :class="{ 'metric-card-warning': warning }">
    <div class="metric-card-header">
      <img class="metric-card-icon small-icons" :src="icon">
      <span class="metric-card-title">{{ title }}</span>
    </div>

    <div class="metric-card-value">{{ value }}</div>

    <div v-if="subtitle" class="metric-card-subtitle">{{ subtitle }}</div>

    <div v-if="percent >= 0" class="metric-card-progress">
      <div class="progress-bar-track">
        <div class="progress-bar-fill" :class="{ 'near-quota': percent >= 90 }" :style="{ width: percent + '%' }"></div>
      </div>
    </div>

    <ActionButton v-if="actionLabel" :icon="actionIcon" :label="actionLabel" @click="actionFunction" :useBackground="true" />
  </div>
</template>

<script setup>
// components
import ActionButton from './ActionButton.vue';

defineProps({
  title: { type: String, required: true },
  value: { type: [String, Number], required: true },
  subtitle: { type: String, default: '' },
  icon: { type: String, required: true },
  actionIcon: { type: String, default: '' },
  actionLabel: { type: String, default: '' },
  actionFunction: { type: Function, default: () => {} },
  percent: { type: Number, default: -1 },
  warning: { type: Boolean, default: false },
});
</script>

<style scoped>
.metric-card {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  flex: 1;
  min-width: 0;
  padding: 1.2rem;
  background-color: var(--black-steel);
  border-radius: var(--very-large-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  box-sizing: border-box;
}

.metric-card-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.metric-card-icon {
  width: 16px;
  height: 16px;
  opacity: 0.7;
}

.metric-card-title {
  font-size: 0.8rem;
  color: var(--silver);
  font-weight: 400;
}

.metric-card-value {
  font-size: 1.6rem;
  font-weight: 600;
  color: var(--white);
  line-height: 1.2;
}

.metric-card-subtitle {
  font-size: 0.75rem;
  color: var(--silver);
  opacity: 0.8;
}

.metric-card-progress {
  margin-top: 0.2rem;
}

.progress-bar-track {
  width: 100%;
  height: 4px;
  background-color: var(--light-steel);
  border-radius: 2px;
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background-color: var(--accent);
  border-radius: 2px;
  transition: width 0.3s ease;
}

.progress-bar-fill.near-quota {
  background-color: var(--error-red, #e05252);
}

.metric-card-warning {
  outline-color: var(--error-red, #e05252);
}
</style>
