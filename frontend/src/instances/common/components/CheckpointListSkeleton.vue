<template>
<div class="checkpoint-skeleton-container" ref="containerRef">
    <div class="checkpoint-item-skeleton" v-for="(skeleton, index) in skeletonArray" :key="index"  :style="{ animationDelay : `${(skeleton - 1) * 0.2}s` }">
        <div class="checkpoint-item-thumb-skeleton"  >
        </div>
        <div class="checkpoint-item-content-skeleton">
        </div>
    </div>
</div>
</template>

<script setup>

import { ref, onMounted, computed, nextTick } from 'vue';

const containerRef = ref(null);
const containerHeight = ref(500); // fallback value

const skeletonArray = computed(() => {
  const itemCount = Math.round(containerHeight.value / 50);
  return Array.from({ length: Math.max(1, itemCount) }, (_, i) => i + 1);
});

onMounted(async () => {
  await nextTick();
  if (containerRef.value) {
    containerHeight.value = containerRef.value.clientHeight;
  }
});

</script>
  

<style scoped>
  @import "@/assets/tray.css";

.checkpoint-item-skeleton{
box-sizing: border-box;
display: flex;
flex-direction: row;
align-items: center;
justify-content: center;
width: 100%;
height: 50px;
overflow: hidden;
padding:.5rem;
opacity: 0;
gap: .5rem;
animation: fadeInFadeOut infinite  4s ease-in-out;
outline: var(--transparent-line);
border-radius: 12px;
border-radius: var(--large-radius);
}

.checkpoint-item-thumb-skeleton{
box-sizing: border-box;
position: relative;
height: 80%;
aspect-ratio: 1/1;
overflow: hidden;
transition: all 0.2s ease-in;
border-radius: 50%;
background-color: var(--steel);
}

.checkpoint-item-content-skeleton{
box-sizing: border-box;
height: 100%;
overflow: hidden;
display: flex;
flex-direction: column;
align-items: center;
flex: 1;
gap: .2rem;
background-color: var(--steel);
border-radius: 10px;
}

.checkpoint-skeleton-container{
  width: 98%;
  height: 98%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: .5rem;
  padding: .2rem;
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
  
  
  

