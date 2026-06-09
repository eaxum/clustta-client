<template>
  <div class="grid-skeleton-container" ref="containerRef" :style="gridStyles">
    <div v-for="(skeleton, index) in skeletonArray" :key="index" class="grid-skeleton-wrapper animate-pulse"
      :style="{ animationDelay: `${(skeleton - 1) * 0.2}s`, minWidth: `${gridSize}px`, height: `${gridSize}px` }">
      <div class="grid-skeleton-item">
        <div class="grid-skeleton-icon-area">
          <div class="rounded-md bg-primary/10 w-1/2 h-1/2"></div>
          <div class="absolute top-2 right-2 rounded-full bg-primary/10 w-5 h-5"></div>
        </div>
        <div class="grid-skeleton-bottom-bar">
          <div class="rounded-full bg-primary/10 w-5 h-5 shrink-0"></div>
          <div class="rounded-md bg-primary/10 flex-1 h-5 mx-1"></div>
          <div class="rounded-full bg-primary/10 w-5 h-5 shrink-0"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, nextTick } from 'vue';
import { useCommonStore } from '@/stores/common';

const commonStore = useCommonStore();

const props = defineProps({
  height: { type: Number, default: null },
  count: { type: Number, default: 12 },
});

const containerRef = ref(null);
const containerHeight = ref(500);

const gridSize = computed(() => commonStore.gridSize);

const gridStyles = computed(() => ({
  display: 'grid',
  boxSizing: 'border-box',
  gridTemplateColumns: `repeat(auto-fill, minmax(${gridSize.value}px, 1fr))`,
  gap: '10px',
  width: '100%',
}));

// Calculates the number of skeleton items based on container dimensions.
const skeletonArray = computed(() => {
  if (props.height && containerHeight.value && gridSize.value) {
    const itemsPerRow = Math.floor((containerRef.value?.clientWidth || 800) / (gridSize.value + 10));
    const rowsNeeded = Math.ceil((props.height || containerHeight.value) / (gridSize.value + 10));
    const itemCount = Math.max(1, itemsPerRow * rowsNeeded);
    return Array.from({ length: itemCount }, (_, i) => i + 1);
  }
  return Array.from({ length: props.count }, (_, i) => i + 1);
});

onMounted(async () => {
  if (!props.height) {
    await nextTick();
    if (containerRef.value) {
      containerHeight.value = containerRef.value.clientHeight;
    }
  }
});
</script>

<style scoped>
.grid-skeleton-container {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
}

.grid-skeleton-wrapper {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  justify-content: flex-end;
  overflow: hidden;
}

.grid-skeleton-item {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  overflow: hidden;
  padding: .5rem;
  box-sizing: border-box;
  border-radius: var(--large-radius);
  border: 1px solid hsl(var(--border));
}

.grid-skeleton-icon-area {
  position: relative;
  display: flex;
  overflow: hidden;
  height: 100%;
  width: 100%;
  border-radius: var(--normal-radius);
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
  border: 1px solid hsl(var(--border));
}

.grid-skeleton-bottom-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .2rem;
  min-height: 32px;
  box-sizing: border-box;
}
</style>