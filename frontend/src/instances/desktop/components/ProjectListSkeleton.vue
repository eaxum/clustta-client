<template>
    <div class="asset-skeleton-container"  :class="{ 'asset-skeleton-container-cards': cardView }">

    <div v-for="(skeleton, index) in skeletonArray" :key="index" class="asset-skeleton-item" 
    :class="{ 'asset-skeleton-item-cards': cardView }"
    :style="{ animationDelay : `${(skeleton - 1) * 0.2}s` }">
        
        <!-- Preview image skeleton (only in card view) -->
        <div v-if="cardView" class="preview-skeleton-container">
          <div class="preview-skeleton"></div>
        </div>
        
        <!-- Footer section -->
        <div class="footer-skeleton">
          <!-- Thumbnail for list view -->
          <div v-if="!cardView" class="thumb-skeleton"></div>
          
          <!-- Project name skeleton -->
          <div class="name-skeleton"></div>
          
          <!-- Action buttons skeleton -->
          <div class="actions-skeleton">
            <div class="icon-skeleton"></div>
            <div class="icon-skeleton"></div>
            <div class="icon-skeleton"></div>
          </div>
        </div>
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

.asset-skeleton-container{
  padding: .5rem;
	box-sizing: border-box;
	height: max-content;
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(500px, 1fr));
	gap: 10px;
	width: 100%;
}
.asset-skeleton-container-cards{
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 10px;
  width: 100%;
}

.asset-skeleton-item{
    display: flex;
    flex-direction: column;
    gap: .2rem;
    color: hsl(var(--foreground));
    align-items: center;
    box-sizing: border-box;
    width: 100%;
    height: min-content;
    justify-content: flex-start;
    /* background-color: hsl(var(--muted)); */
    border-radius: var(--large-radius);
    overflow: hidden;
    min-width: 500px;
    min-height: 50px;
    opacity: 0;
    animation: fadeInFadeOut infinite  3s ease-in-out;
    box-sizing: border-box;
    border: 1px solid hsl(var(--border));
    
}

.asset-skeleton-item-cards{
    min-width: 300px;
    height: 200px;
    gap: 0;
}

.preview-skeleton-container{
    box-sizing: border-box;
    width: 100%;
    aspect-ratio: 16/9;
    padding: .3rem;
    height: 50%;
    height: 150px;
}

.preview-skeleton{
    box-sizing: border-box;
    border: 1px solid hsl(var(--border));
    
    width: 100%;
    aspect-ratio: 16/9;
    border-radius: var(--very-large-radius);
    height: 100%;
}

.footer-skeleton{
    display: flex;
    gap: .5rem;
    align-items: center;
    padding: .4rem;
    box-sizing: border-box;
    width: 100%;
    height: 50px;
    justify-content: space-between;
}

.thumb-skeleton{
    box-sizing: border-box;
    background-color: hsl(var(--accent));
    height: 100%;
    aspect-ratio: 16/9;
    border-radius: var(--very-large-radius);
    min-width: 60px;
}

/* Project name skeleton */
.name-skeleton{
    box-sizing: border-box;
    background-color: hsl(var(--accent));
    width: 100%;
    height: 30px;
    border-radius: var(--very-large-radius);
}

/* Action buttons container */
.actions-skeleton{
    display: flex;
    gap: .7rem;
    align-items: center;
    justify-content: space-between;
    min-width: max-content;
}

.icon-skeleton{
    box-sizing: border-box;
    background-color: hsl(var(--accent));
    width: 20px;
    height: 20px;
    border-radius: var(--normal-radius);
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
  
  
  

