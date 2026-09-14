<template>
  <aside ref="details" tabindex="-1" class="activity-details" :aria-label="$t('activity.details')"
    @keydown.esc.stop="$emit('close')">
    <ExpandablePanelHeader closeIcon="chevron-right" :title="$t('activity.requestedAssets', { count: operation.assets.length })"
      icon="info" :showMaximize="false" @close="$emit('close')" />
    <div class="activity-details-body app-scrollbar">
      <AssetItem v-for="asset in operation.assets" :key="asset.id" :item="asset"
        :hideExtension="commonStore.hideExtensions" :showNavigate="true"
        :navigateTooltip="$t('menus.goToAsset')" variant="compact"
        tabindex="0" @keydown.enter.prevent="$emit('navigate', operation, asset)"
        @navigate="$emit('navigate', operation, asset)" />
      <p v-if="error" role="alert">{{ error }}</p>
    </div>
  </aside>
</template>

<script setup>
import { onMounted, ref } from 'vue';
import { useCommonStore } from '@/stores/common';
import AssetItem from '@/instances/desktop/components/AssetItem.vue';
import ExpandablePanelHeader from '@/instances/desktop/components/ExpandablePanelHeader.vue';

defineProps({
  operation: { type: Object, required: true },
  error: { type: String, default: '' },
});
defineEmits(['close', 'navigate']);
const commonStore = useCommonStore();
const details = ref(null);
onMounted(() => details.value.focus());
</script>

<style scoped>

.activity-details {
  display: flex;
  flex-direction: column;
  width: min(400px, 100%);
  height: 100%;
  padding: .4rem;
  box-sizing: border-box;
  border: 1px solid var(--surface-4);
  border-radius: var(--large-radius);
  background: var(--surface-1);
  color: var(--text);
  box-shadow: 0 12px 40px #0006;
  overflow: hidden;
  animation: details-slide-in .18s ease-out;
}
.activity-details-body {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  min-height: 0;
  overflow-y: auto;
  padding: .5rem;
}
.activity-details-body > * {
  flex-shrink: 0;
}
p[role='alert'] {
  color: var(--red);
  overflow-wrap: anywhere;
}
@keyframes details-slide-in {
  from {
    opacity: 0;
    transform: translateX(2rem);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}
@media (prefers-reduced-motion: reduce) {
  .activity-details {
    animation: none;
  }
}
</style>
