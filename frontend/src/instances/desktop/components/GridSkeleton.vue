<template>
  <div class="grid-skeleton-container" ref="containerRef" :style="gridStyles">
    <div v-for="(skeleton, index) in skeletonArray" :key="index" class="grid-skeleton-wrapper"
      :style="{
        animationDelay: `${(skeleton - 1) * 0.2}s`,
        animationDuration: `${animationDuration}s`,
        minWidth: `${gridSize}px`,
        height: `${gridSize}px`
      }">
      <div class="grid-skeleton-item">
        <!-- Main icon area -->
        <div class="grid-skeleton-icon-area">
          <div class="grid-skeleton-main-icon"></div>
          <!-- Top right assignee placeholder -->
          <div class="grid-skeleton-assignee-top-right"></div>
        </div>
        
        <!-- Bottom bar -->
        <div class="grid-skeleton-bottom-bar">
          <div class="grid-skeleton-type-icon"></div>
          <div class="grid-skeleton-name"></div>
          <div class="grid-skeleton-status"></div>
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
  count: { type: Number, default: 12 }, // Default number of skeleton items
});

const containerRef = ref(null);
const containerHeight = ref(500); // fallback value

const gridSize = computed(() => commonStore.gridSize);

const gridStyles = computed(() => ({
  display: 'grid',
  boxSizing: 'border-box',
  gridTemplateColumns: `repeat(auto-fill, minmax(${gridSize.value}px, 1fr))`,
  gap: '10px',
  width: '100%',
}));

const skeletonArray = computed(() => {
  if (props.height && containerHeight.value && gridSize.value) {
    // Calculate how many items can fit in the container
    const itemsPerRow = Math.floor((containerRef.value?.clientWidth || 800) / (gridSize.value + 10));
    const rowsNeeded = Math.ceil((props.height || containerHeight.value) / (gridSize.value + 10));
    const itemCount = Math.max(1, itemsPerRow * rowsNeeded);
    return Array.from({ length: itemCount }, (_, i) => i + 1);
  }
  return Array.from({ length: props.count }, (_, i) => i + 1);
});

const animationDuration = computed(() => {
  // Base duration of 1s + 0.3s per skeleton item, capped at reasonable maximum
  return Math.max(3, Math.min(skeletonArray.value.length * 0.1, 3));
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
@import "@/assets/desktop.css";

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
  color: white;
  align-items: stretch;
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  justify-content: flex-end;
  overflow: hidden;
  opacity: 0;
  animation: fadeInFadeOut infinite ease-in-out;
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
  /* background-color: var(--surface-2); */
  outline: var(--transparent-line);
  outline-offset: -1.5px;
}

.grid-skeleton-icon-area {
  position: relative;
  display: flex;
  overflow: hidden;
  height: 100%;
  width: 100%;
  /* background-color: rgba(0, 0, 0, 0.2); */
  border-radius: 8px;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
  outline: var(--transparent-line);
}

.grid-skeleton-main-icon {
  box-sizing: border-box;
  background-color: var(--surface-3);
  width: 50%;
  height: 50%;
  border-radius: 8px;
}

.grid-skeleton-assignee-top-right {
  position: absolute;
  top: 8px;
  right: 8px;
  box-sizing: border-box;
  background-color: var(--surface-3);
  width: 20px;
  height: 20px;
  border-radius: 50%;
}

.grid-skeleton-bottom-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .2rem;
  min-height: 32px;
  box-sizing: border-box;
}

.grid-skeleton-type-icon {
  box-sizing: border-box;
  background-color: var(--surface-3);
  width: 20px;
  height: 20px;
  border-radius: 50%;
  flex-shrink: 0;
}

.grid-skeleton-name {
  box-sizing: border-box;
  background-color: var(--surface-3);
  flex: 1;
  height: 20px;
  border-radius: 8px;
  margin-left: .3rem;
  margin-right: .3rem;
}

.grid-skeleton-status {
  box-sizing: border-box;
  background-color: var(--surface-3);
  width: 20px;
  height: 20px;
  border-radius: 50%;
  flex-shrink: 0;
}

@keyframes fadeInFadeOut {
  from {
    opacity: 0;
  }
  20% {
    opacity: 1;
  }
  70% {
    opacity: 1;
  }
  80% {
    opacity: 0;
  }
  to {
    opacity: 0;
  }
}
</style>