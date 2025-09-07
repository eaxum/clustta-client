<template>
<div class="task-skeleton-container" >
  <div v-for="(skeleton, index) in skeletonArray" :key="index" class="task-skeleton-item" 
  :style="{ animationDelay : `${(skeleton - 1) * 0.2}s` }">
    <!-- <div class="task-skeleton-spacer"> </div>
    <div class="thumb-skeleton"> </div>
    <div class="icon-skeleton" ></div>
    <div class="task-skeleton-item-launcher"> </div>
    <div class="icon-skeleton" ></div>
    <div class="status-pill"></div>
    <div class="icon-skeleton" ></div>
    <div class="icon-skeleton" ></div> -->
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

.task-skeleton-container{
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.task-skeleton-item{
  display: flex;
  gap: .5rem;
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
  min-height: 60px;
  opacity: 0;
  animation: fadeInFadeOut infinite  3s ease-in-out;
  
}

.task-skeleton-item-launcher{
    box-sizing: border-box;
    background-color: white;
    opacity: .1;
    width: 100%;
    height: 30px;
    height: 60%;
    border-radius: 12px;
}

.task-skeleton-spacer{
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
  
  
  

