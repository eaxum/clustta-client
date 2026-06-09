<template>
<div class="asset-skeleton-container" >
  <div v-for="(skeleton, index) in skeletonArray" :key="index" class="asset-skeleton-item" 
  :style="{ animationDelay : `${(skeleton - 1) * 0.2}s` }">
  </div>
  
</div>
</template>

<script setup>

import { ref, onMounted, computed  } from 'vue';

const props = defineProps({
  forModal: { type: Boolean, default: false },
})

const skeletonArrayBrowser = ref(Array.from({ length: 13 }, (_, i) => i + 1));
const skeletonArrayModal = ref(Array.from({ length: 13 }, (_, i) => i + 1));

const skeletonArray = computed(() => {
  return props.forModal ? skeletonArrayModal.value : skeletonArrayBrowser.value;
})


onMounted(async () => {

});

</script>
  



<style scoped>
  /* @import "@/assets/tray.css"; */

.asset-skeleton-container{
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.asset-skeleton-item{
  display: flex;
  gap: .5rem;
  color: hsl(var(--foreground));
  align-items: center;
  padding-left: .5rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: space-between;
  align-items: center;
  background-color: hsl(var(--muted));
  border-radius: var(--large-radius);
  overflow: hidden;
  padding: .5rem;
  min-height: 60px;
  opacity: 0;
  animation: fadeInFadeOut infinite  3s ease-in-out;
  
}

.asset-skeleton-item-launcher{
    box-sizing: border-box;
    background-color: hsl(var(--foreground));
    opacity: .1;
    width: 100%;
    height: 30px;
    height: 60%;
    border-radius: var(--very-large-radius);
}

.asset-skeleton-spacer{
    box-sizing: border-box;
    /* background-color: hsl(var(--foreground)); */
    opacity: .1;
    width: 50px;
    height: 30px;
    height: 60%;
    border-radius: var(--very-large-radius);
}

.thumb-skeleton{
    box-sizing: border-box;
    background-color: hsl(var(--foreground));
    opacity: .1;
    height: 80%;
    aspect-ratio: 16/9;
    border-radius: var(--normal-radius);
}

.status-pill{
    box-sizing: border-box;
    background-color: hsl(var(--foreground));
    opacity: .1;
    width: 5rem;
    height: 30px;
    height: 60%;
    border-radius: var(--very-large-radius);
}

.icon-skeleton{
    box-sizing: border-box;
    background-color: hsl(var(--foreground));
    opacity: .1;
    /* width: 5rem; */
    height: 50%;
    aspect-ratio: 1/1;
    border-radius: var(--very-large-radius);
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
  
  
  

