<template>
  <div class="virtua-scroll-container" ref="containerRef" >
    <VirtuaList
      :items="props.items"
      :isRoot="true"
      :containerHeight="containerHeight"
      :itemHeight="commonStore.listItemHeight"
    />
  </div>
</template>

<script setup>

// imports
import { ref, onMounted, onUnmounted, watchEffect, provide, computed } from 'vue';

import { useScrollStore } from '@/stores/scroll';
import { useStageStore } from '@/stores/stages';
import { useMenu } from '@/stores/menu';
import { useCommonStore } from '@/stores/common';

import VirtuaList from '@/instances/common/components/VirtuaList.vue';

const scrollStore = useScrollStore();
const stage = useStageStore();
const menu = useMenu();
const commonStore = useCommonStore();

// refs
let animationFrame = null;
const containerRef = ref(null);

provide('parentScrollContainer', containerRef);
provide('rootScrollContainer', containerRef);

// props
const props = defineProps({
  items: {
    type: Array,
    default: []
  }
});

// computed props
const containerHeight = computed(() => {
  const height = containerRef.value?.getBoundingClientRect().height;
  scrollStore.scrollRootHeight = height;
  return height ? height : 500
});

// methods
const onScroll = (e) => {
  menu.disableAllMenus();
  if (animationFrame) {
    cancelAnimationFrame(animationFrame);
  }
  animationFrame = requestAnimationFrame(() => {
    scrollStore.setScrollTop(e.target.scrollTop);
  });
};

const scrollToPosition = (position, smooth = false) => {
  if (containerRef.value) {
    containerRef.value.scrollTo({
      top: position,
      behavior: smooth ? 'smooth' : 'auto'  
    });
  }
};

scrollStore.$subscribe((mutation, state) => {
  if (state.requestedScrollPosition !== null) {
    scrollToPosition(state.requestedScrollPosition, scrollStore.smoothScroll);
    scrollStore.clearRequestedScrollPosition();
    scrollStore.smoothScroll = false;
  }
});

onMounted(() => {
  const scrollContainer = containerRef.value;
  scrollStore.scrollRoot = containerRef.value;
  scrollStore.setScrollTop(scrollContainer.scrollTop);
  scrollContainer.addEventListener('scroll', onScroll);
});

onUnmounted(() => {
  if (containerRef.value) {
    containerRef.value.removeEventListener('scroll', onScroll);
  }
  if (animationFrame) {
    cancelAnimationFrame(animationFrame);
  }
});
</script>

<style scoped>
.virtua-scroll-container{
  box-sizing: border-box;
  height: 600px; 
  width: 100%;
  height: 100%;
  overflow: auto;
  padding-right: 5px;
  /* background-color: crimson; */
  /* border-top: 1px solid royalblue; */
}

.virtua-scroll-container::-webkit-scrollbar {
  width: 6px;
}

.virtua-scroll-tray::-webkit-scrollbar {
  width: 4px;
}

.virtua-scroll-container::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgba(255, 255, 255, 0.295);

}

.virtua-scroll-container::-webkit-scrollbar-track {
  border-radius: 10px;
}
</style>