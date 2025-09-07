<template>
<div class="virtua-skeleton-container" :class="{ 'indent-style' : useIndent  }">

  <div v-if="useIndent" class="indent-guide-skeleton" :style="{ height: `${height}px`}" >
  </div>

  <div v-for="(skeleton, index) in skeletonArray" :key="index" class="virtua-skeleton-wrapper" 
  :style="{ 
    animationDelay : `${(skeleton - 1) * 0.2}s`, 
    height: `${itemHeight}px`,
    animationDuration: `${animationDuration}s`
  }" >
    <div class="virtua-skeleton-item" >
    </div>
  </div>
</div>
</template>

<script setup>

import { ref, onMounted, computed  } from 'vue';

const props = defineProps({
  forModal: { type: Boolean, default: false },
  height: { type: Number, default: 50 },
  itemHeight: { type: Number, default: 50 },
  depth: { type: Number, default: 0 },
})

const skeletonArray = computed(() => {
  const itemCount = Math.round(props.height / props.itemHeight);
  return Array.from({ length: Math.max(1, itemCount) }, (_, i) => i + 1);
});

const animationDuration = computed(() => {
  // Base duration of 1s + 0.3s per skeleton item
  return Math.max(1, skeletonArray.value.length * 0.3);
});

const useIndent = computed(() => {
  return props.depth > 0
})


onMounted(async () => {
    console.log()
});

</script>
  
<style scoped>

.indent-guide-skeleton {
  position: absolute;
  /* height: 30px; */
  width: 100%;
  box-sizing: border-box;
  border-left: var(--transparent-line);
  left: 15px;
}

.indent-style{
  padding-left: 30px;
  /* background-color: crimson; */
}

.virtua-skeleton-container{
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 0px;
  padding-bottom: 5px;
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
  /* background-color: var(--dark-steel); */
  /* background-color: crimson; */
  overflow: hidden;
  min-height: 50px;
  opacity: 0;
  animation: fadeInFadeOut infinite ease-in-out;
  /* padding-right: 10px; */
}

.virtua-skeleton-item{
  display: flex;
  color: white;
  align-items: center;
  padding-left: .5rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: space-between;
  align-items: center;
  background-color: var(--dark-steel);
  border-radius: 10px;
  overflow: hidden;
  padding: .5rem;
  height: 90%;
  /* min-height: 50px; */
  /* height: 100%; */
}

.virtua-skeleton-item-launcher{
    box-sizing: border-box;
    background-color: white;
    opacity: .1;
    width: 100%;
    height: 30px;
    height: 60%;
    border-radius: 12px;
}

.virtua-skeleton-spacer{
    box-sizing: border-box;
    /* background-color: white; */
    opacity: .1;
    width: 50px;
    height: 30px;
    height: 60%;
    border-radius: 12px;
}

.thumb-skeleton{
    box-sizing: border-box;
    background-color: white;
    opacity: .1;
    height: 80%;
    aspect-ratio: 16/9;
    border-radius: 8px;
}

.status-pill{
    box-sizing: border-box;
    background-color: white;
    opacity: .1;
    width: 5rem;
    height: 30px;
    height: 60%;
    border-radius: 12px;
}

.icon-skeleton{
    box-sizing: border-box;
    background-color: white;
    opacity: .1;
    /* width: 5rem; */
    height: 50%;
    aspect-ratio: 1/1;
    border-radius: 12px;
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
  
  
  

