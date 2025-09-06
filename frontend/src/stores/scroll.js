import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useScrollStore = defineStore('scroll', () => {
  const scrollTop = ref(0);
  const requestedScrollPosition = ref(null);
  const scrollRoot = ref(null);
  const scrollRootHeight = ref(0);
  const smoothScroll = ref(false);
  
  const setScrollTop = (value) => {
    scrollTop.value = value;
  }

  const requestScroll = (position) => {
    requestedScrollPosition.value = position;
  }

  const clearRequestedScrollPosition = () => {
    requestedScrollPosition.value = null;
  }

  return {
    scrollRoot,
    scrollRootHeight,
    scrollTop,
    requestedScrollPosition,
    setScrollTop,
    requestScroll,
    clearRequestedScrollPosition
  };
});