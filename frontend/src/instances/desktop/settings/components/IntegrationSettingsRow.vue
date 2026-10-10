<template>
  <div class="integration-settings-row">
    <div class="integration-settings-icon">
      <img class="small-icons" :src="getAppIcon(icon)" alt="">
    </div>
    <div class="integration-settings-content">
      <div class="integration-settings-label">{{ label }}</div>
      <div class="integration-settings-description">{{ description }}</div>
    </div>
    <div v-stop-propagation>
      <ActionButton :label="$t('common.configure')" :showIcon="false" :buttonFunction="open" />
    </div>
  </div>
</template>

<script setup>
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import { useIconStore } from '@/stores/icons';

defineProps({
  description: {
    type: String,
    required: true,
  },
  icon: {
    type: String,
    required: true,
  },
  label: {
    type: String,
    required: true,
  },
  open: {
    type: Function,
    required: true,
  },
});

const iconStore = useIconStore();
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);
</script>

<style scoped>
.integration-settings-row {
  display: flex;
  align-items: center;
  width: 100%;
  min-height: 50px;
  padding: .5rem 1rem;
  border-bottom: 1px solid var(--surface-4);
  color: var(--text);
  background-color: var(--surface-2);
  cursor: default;
  box-sizing: border-box;
}

.integration-settings-icon {
  display: flex;
  padding: .3rem;
}

.integration-settings-content {
  display: flex;
  flex: 1;
  flex-direction: column;
  padding: .4rem .2rem;
}

.integration-settings-label {
  padding: .1rem;
  font-size: 14px;
  font-weight: 400;
}

.integration-settings-description {
  padding: .1rem;
  color: var(--text-muted);
  font-size: 12px;
}
</style>
