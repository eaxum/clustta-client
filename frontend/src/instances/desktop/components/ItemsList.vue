<template>
  <div ref="scrollContainer" class="virtual-scroll-container" @scroll="handleScroll" >
    <div :style="{ height: totalHeight + 'px' }" class="scroll-height">
      <div :style="{ transform: `translateY(${startOffset}px)` }">
        <div v-for="(item, index) in visibleItems" :key="firstVisibleIndex + index"
          :data-index="firstVisibleIndex + index">
          <VirtualNode :isDependency="isDependency" :forList="forList" :showAdd="showAdd" :showRemove="showRemove" :data="item" class="virtual-scroll-item" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue';


import VirtualNode from '@/instances/desktop/components/VirtualNode.vue'

import { useMenu } from '@/stores/menu';

const menu = useMenu();

const props = defineProps({
  items: { type: Array, required: true },

  defaultHeight: { type: Number, default: 50 },
  bufferSize: { type: Number, default: 20 },
  
  isDependency: { type: Boolean, default: false },
  
  forList: { type: Boolean, default: false },
  showAdd: { type: Boolean, default: false },
  showRemove: { type: Boolean, default: false },
})

const scrollContainer = ref(null)
const itemRefs = ref({});
const scrollTop = ref(0)
const containerHeight = ref(0);

// Map to store heights of all items
const itemHeights = ref(new Map())
const heightCache = ref(new Map())

// ResizeObserver instance
let resizeObserver = null

// Initialize or update ResizeObserver
const setupResizeObserver = () => {
  if (resizeObserver) {
    resizeObserver.disconnect()
  }

  resizeObserver = new ResizeObserver((entries) => {
    let heightsUpdated = false

    for (const entry of entries) {
      const index = parseInt(entry.target.getAttribute('data-index'))
      const height = entry.contentRect.height

      if (heightCache.value.get(index) !== height) {
        heightCache.value.set(index, height)
        itemHeights.value.set(index, height)
        heightsUpdated = true
      }
    }

    if (heightsUpdated) {
      updatePositionCache()
    }
  })
}

// Cache of cumulative heights for faster calculations
const positionCache = ref(new Map())

// Update position cache when heights change
const updatePositionCache = () => {
  let accumHeight = 0
  positionCache.value.clear()

  for (let i = 0; i < props.items.length; i++) {
    positionCache.value.set(i, accumHeight)
    accumHeight += itemHeights.value.get(i) || props.defaultHeight
  }

  totalHeight.value = accumHeight
}

// Calculate total height
const totalHeight = ref(0)

// Find the index of the first visible item based on scroll position
const findFirstVisibleIndex = (scrollTop) => {
  let low = 0
  let high = props.items.length - 1

  while (low <= high) {
    const mid = Math.floor((low + high) / 2)
    const position = positionCache.value.get(mid) || 0

    if (position <= scrollTop && scrollTop < (position + (itemHeights.value.get(mid) || props.defaultHeight))) {
      return mid
    } else if (position <= scrollTop) {
      low = mid + 1
    } else {
      high = mid - 1
    }
  }

  return low
}

// Calculate visible range
const visibleRange = computed(() => {
  const start = findFirstVisibleIndex(scrollTop.value)
  let end = start
  let currentHeight = 0

  while (end < props.items.length && currentHeight < containerHeight.value) {
    currentHeight += itemHeights.value.get(end) || props.defaultHeight
    end++
  }

  // Add buffer items before and after
  const rangeStart = Math.max(0, start - props.bufferSize)
  const rangeEnd = Math.min(props.items.length, end + props.bufferSize)

  return {
    start: rangeStart,
    end: rangeEnd
  }
})

// Get visible items
const visibleItems = computed(() => {
  const { start, end } = visibleRange.value
  return props.items.slice(start, end)
})

// Calculate start offset for visible items
const startOffset = computed(() => {
  return positionCache.value.get(visibleRange.value.start) || 0
})

const firstVisibleIndex = computed(() => visibleRange.value.start)

// Handle scroll events
const handleScroll = (event) => {
  menu.disableAllMenus()
  scrollTop.value = event.target.scrollTop
}

// Watch for changes in visible items to update observers
watch(() => visibleItems.value, async () => {
  await nextTick()

  // Remove all existing observers
  if (resizeObserver) {
    resizeObserver.disconnect()
  }

  // Observe new items
  if (itemRefs.value) {

    Object.values(itemRefs.value).forEach(el => {
      if (el) resizeObserver.observe(el)
    })
  }
}, { flush: 'post' })

// Initialize on mount
onMounted(() => {
  if (scrollContainer.value) {
    containerHeight.value = scrollContainer.value.clientHeight

    // Observe container size
    const containerObserver = new ResizeObserver((entries) => {
      containerHeight.value = entries[0].contentRect.height
    })

    containerObserver.observe(scrollContainer.value)

    // Setup item observers
    setupResizeObserver()

    // Initialize height cache with default heights
    props.items.forEach((_, index) => {
      itemHeights.value.set(index, props.defaultHeight)
    })

    updatePositionCache()
  }
})

// Watch for changes in items array
watch(() => props.items.length, () => {
  props.items.forEach((_, index) => {
    if (!itemHeights.value.has(index)) {
      itemHeights.value.set(index, props.defaultHeight)
    }
  })

  updatePositionCache()
})
</script>

<style scoped>
@import "@/assets/desktop.css";

.virtual-scroll-container {
  height: 100%;
  width: 100%;
  overflow-y: auto;
  position: relative;
  border-radius: var(--large-radius);
  padding: .2rem;
  box-sizing: border-box;
}

.virtual-scroll-container::-webkit-scrollbar {
  width: 6px;
}

.virtual-scroll-tray::-webkit-scrollbar {
  width: 4px;
}

.virtual-scroll-container::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: white;
  background-color: rgba(255, 255, 255, 0.295);

}

.virtual-scroll-container::-webkit-scrollbar-track {
  border-radius: 10px;
}

.scroll-height {
  position: relative;
}

.virtual-scroll-item {
  box-sizing: border-box;
  height: min-content;
  margin-bottom: 8px;
  overflow: hidden;
  /* opacity: 0; */
  animation: fadeIn .3s ease-in-out forwards;
}

.column-card {
  border-radius: 8px;
  cursor: grab;
}

.column-card.dragging {
  color: transparent;
  background: none;
  box-shadow: none;
  user-select: none;
  -moz-user-select: none;
  -webkit-user-select: none;
  cursor: grabbing;
  opacity: .2;
}

#ghost-card {
  position: fixed;
  z-index: 100;
  user-select: none;
  pointer-events: none;
  opacity: 0;
  transform-origin: center;
  transform: scale(1) rotate(0);
  box-shadow: 0 1px 0 rgba(9, 30, 66, 0.25);
  transition: transform 0.04s ease-in-out;
}

#ghost-card.animate {
  transition: box-shadow 0.1s ease-in-out;
  transition: transform 0.05s ease-in-out;
}

#ghost-card.active {
  outline: 1.5px solid rgb(255, 255, 255);
  outline-offset: -1.5px;
  opacity: 1;
  /* box-shadow: 0 12px 24px -6px rgba(9, 30, 66, 0.25), 0 0 0 1px rgba(9, 30, 66, 0.08); */
}

#ghost-card.leaving {
  transition: all 0.1s ease;
  box-shadow: 0 1px 0 rgba(9, 30, 66, 0.25);
}
</style>

