<template>
    <div class="task-skeleton-container"  :class="{ 'task-skeleton-container-cards': cardView }">

    <div v-for="(skeleton, index) in skeletonArray" :key="index" class="task-skeleton-item" 
    :class="{ 'task-skeleton-item-cards': cardView }"
    :style="{ animationDelay : `${(skeleton - 1) * 0.2}s` }">
        <!-- <div class="task-skeleton-spacer"> </div>
        <div class="thumb-skeleton"> </div>
        <div class="icon-skeleton" ></div>
        <div class="task-skeleton-item-launcher"> </div>
        <div class="status-pill"></div>
        <div class="icon-skeleton" ></div>
        <div class="icon-skeleton" ></div>
        <div class="icon-skeleton" ></div> -->
    </div>
    
    </div>
</template>

<script setup>

import { ref, onMounted  } from 'vue';

import { useTrayStates } from '@/stores/TrayStates';

const trayStates = useTrayStates();
const skeletonArray = ref([1,2,3,4,5,6,7,8,9,10,11,12,13]);

const props = defineProps({
    cardView: { type: Boolean, default: false },
});

onMounted(async () => {

});

</script>
  



<style scoped>

.task-skeleton-container{
	box-sizing: border-box;
	height: max-content;
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(500px, 1fr));
	gap: 10px;
	width: 100%;
}
.task-skeleton-container-cards{
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 10px;
  width: 100%;
}

.task-skeleton-item{
    display: flex;
    gap: .2rem;
    color: white;
    align-items: center;
    padding: .3rem;
    box-sizing: border-box;
    width: 100%;
    height: min-content;
    justify-content: flex-start;
    background-color: var(--dark-steel);
    border-radius: var(--normal-radius);
    overflow: hidden;
    min-width: 500px;
    min-height: 60px;
    opacity: 0;
    animation: fadeInFadeOut infinite  3s ease-in-out;
}

.task-skeleton-item-cards{
    min-width: 300px;
    height: 300px;
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
  
  
  