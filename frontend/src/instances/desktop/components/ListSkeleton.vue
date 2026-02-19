<template>
<div class="virtua-skeleton-container" ref="containerRef" :class="{ 'indent-style' : useIndent  }">

  <div v-if="useIndent" class="indent-guide-skeleton" :style="{ height: `${indentHeight}px`}" >
  </div>

  <div v-for="(skeleton, index) in skeletonArray" :key="index" class="virtua-skeleton-wrapper" 
  :style="{ 
    animationDelay : `${(skeleton - 1) * 0.2}s`, 
    height: `${itemHeight}px`,
    animationDuration: `${animationDuration}s`
  }" >
    <div class="virtua-skeleton-item" >
      <div class="icon-skeleton"></div>
      <div class="virtua-skeleton-item-launcher"></div>
      <div class="icon-skeleton"></div>
      <div class="status-pill"></div>
      <div class="icon-skeleton"></div>
    </div>
  </div>
</div>
</template>

<script setup>

import { ref, onMounted, computed, nextTick } from 'vue';

const props = defineProps({
  forModal: { type: Boolean, default: false },
  height: { type: Number, default: null },
  itemHeight: { type: Number, default: 50 },
  depth: { type: Number, default: 0 },
})

const containerRef = ref(null);
const containerHeight = ref(500); // fallback value

const skeletonArray = computed(() => {
  const heightToUse = props.height ? props.height : containerHeight.value;
  const itemCount = Math.round(heightToUse / props.itemHeight);
  return Array.from({ length: Math.max(1, itemCount) }, (_, i) => i + 1);
});

const animationDuration = computed(() => {
  // Base duration of 1s + 0.3s per skeleton item
  return Math.max(1, skeletonArray.value.length * 0.3);
});

const indentHeight = computed(() => {
  return props.height ? props.height : containerHeight.value;
});

const useIndent = computed(() => {
  return props.depth > 0
})

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
  @import "@/assets/tray.css";

.indent-guide-skeleton {
  position: absolute;
  width: 100%;
  box-sizing: border-box;
  border-left: var(--transparent-line);
  left: 15px;
}

.indent-style{
  padding-left: 30px;
}

.virtua-skeleton-container{
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-sizing: border-box;
}

.virtua-skeleton-wrapper{
  display: flex;
  color: white;
  align-items: center;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: space-between;
  align-items: center;
  overflow: hidden;
  min-height: 34px;
  opacity: 0;
  animation: fadeInFadeOut infinite ease-in-out;
}

.virtua-skeleton-item{
  display: flex;
  gap: .5rem;
  color: var(--white);
  align-items: center;
  padding-left: .5rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: space-between;
  align-items: center;
  border-radius: 12px;
  overflow: hidden;
  padding: .5rem 1rem;
  height: 90%;
  outline: var(--transparent-line);
  
  outline-offset: -1px;
  border-radius: var(--large-radius);
}

.virtua-skeleton-item-launcher{
    box-sizing: border-box;
    background-color: var(--steel);
    width: 100%;
    height: 30px;
    height: 60%;
    border-radius: 8px;
}

.status-pill{
    box-sizing: border-box;
    background-color: var(--steel);
    width: 5rem;
    height: 30px;
    height: 60%;
    border-radius: 12px;
}

.icon-skeleton{
    box-sizing: border-box;
    background-color: var(--steel);
    height: 60%;
    aspect-ratio: 1/1;
    border-radius: 50%;
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
  
  
  

