<template>
  <div class="virtua-scroll-container" ref="containerRef" >
    <ListSkeleton v-if="!assetStore.assetsLoaded" 
      :height="containerHeight" 
      :itemHeight="commonStore.listItemHeight" />
    <VirtuaList
      v-else
      :items="props.items"
      :isRoot="true"
      :containerHeight="containerHeight"
      :itemHeight="commonStore.listItemHeight"
      :renderAhead="40"
    />
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, onUnmounted, provide, ref } from 'vue';

// components
import ListSkeleton from '@/instances/desktop/components/ListSkeleton.vue';
import VirtuaList from '@/instances/common/components/VirtuaList.vue';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCommonStore } from '@/stores/common';
import { useMenu } from '@/stores/menu';
import { useScrollStore } from '@/stores/scroll';

const assetStore = useAssetStore();
const commonStore = useCommonStore();
const menu = useMenu();
const scrollStore = useScrollStore();

// props
const props = defineProps({
  items: { type: Array, default: [] }
});

// refs
let animationFrame = null;
const containerRef = ref(null);

// provides
provide('parentScrollContainer', containerRef);
provide('rootScrollContainer', containerRef);

// computed
const containerHeight = computed(() => {
  const height = containerRef.value?.getBoundingClientRect().height;
  scrollStore.scrollRootHeight = height;
  return height ? height : 500;
});

// methods
// Handles scroll events with debouncing via requestAnimationFrame.
const onScroll = (e) => {
  menu.disableAllMenus();
  if (animationFrame) {
    cancelAnimationFrame(animationFrame);
  }
  animationFrame = requestAnimationFrame(() => {
    scrollStore.setScrollTop(e.target.scrollTop);
  });
};

// Scrolls the container to a specified position.
const scrollToPosition = (position, smooth = false) => {
  if (containerRef.value) {
    containerRef.value.scrollTo({
      top: position,
      behavior: smooth ? 'smooth' : 'auto'
    });
  }
};

// subscriptions
scrollStore.$subscribe((mutation, state) => {
  if (state.requestedScrollPosition !== null) {
    scrollToPosition(state.requestedScrollPosition, scrollStore.smoothScroll);
    scrollStore.clearRequestedScrollPosition();
    scrollStore.smoothScroll = false;
  }
});

// lifecycle hooks
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
.virtua-scroll-container {
  box-sizing: border-box;
  height: 100%;
  overflow: auto;
  padding-right: 5px;
  width: 100%;
}

.virtua-scroll-container::-webkit-scrollbar {
  width: 6px;
}

.virtua-scroll-container::-webkit-scrollbar-thumb {
  background-color: var(--surface-4);
  border-radius: 10px;
}

.virtua-scroll-container::-webkit-scrollbar-track {
  border-radius: 10px;
}
</style>

