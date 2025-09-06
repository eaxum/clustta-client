<template>
    <div class="accordion">
      <div class="accordion-header">
        <div class="header-content">
        </div>
      </div>
      <div class="accordion-content" :class="{ 'is-open': isOpen }">
        <div class="accordion-items" ref="itemsContainer">
          <!-- Top spacer -->
          <div 
            v-if="topSpacerHeight > 0"
            class="accordion-spacer"
            :style="{ 
              height: `${topSpacerHeight}px`,
              marginBottom: '5px'
            }"
          ></div>
  
          <!-- Visible items -->
          <div 
            v-for="(item, index) in entityData" 
            :key="index"
            v-show="isInRange(index)"
            class="accordion-item"
            :ref="el => setItemRef(el, index)"
          >
            {{ item }}
          </div>
  
          <!-- Bottom spacer -->
          <div 
            v-if="bottomSpacerHeight > 0"
            class="accordion-spacer"
            :style="{ 
              height: `${bottomSpacerHeight}px`,
              marginTop: '-5px'
            }"
          ></div>
        </div>
      </div>
    </div>
  </template>
  
  <script setup>
  import { ref, computed, onMounted, onUpdated } from 'vue'
  
  const isOpen = ref(true);
  const itemsContainer = ref(null);
  const itemRefs = ref(new Map());
  const itemHeights = ref(new Map());
  
  // props
  const props = defineProps({
    entityData: Array,
  });
  
  const entityData = ref(props.entityData);


  const fromIndex = ref(3);
  const toIndex = ref(5);


  const ITEM_MARGIN = 5;  // Margin between items
  
  // Function to store refs for measuring
  const setItemRef = (el, index) => {
    if (el) {
      itemRefs.value.set(index, el);
    } else {
      itemRefs.value.delete(index);
    }
  };
  
  // Measure and store heights of visible items
  const measureVisibleItems = () => {
    itemRefs.value.forEach((el, index) => {
      if (el && isInRange(index)) {
        const height = el.getBoundingClientRect().height;
        itemHeights.value.set(index, height);
      }
    });
  };
  
  // Calculate average height for items that haven't been measured
  const getAverageHeight = () => {
    if (itemHeights.value.size === 0) return 50; // Default fallback height
    const totalHeight = Array.from(itemHeights.value.values()).reduce((sum, height) => sum + height, 0);
    return totalHeight / itemHeights.value.size;
  };
  
  const calculateHeight = (startIdx, count) => {
    if (count <= 0) return 0;
    
    let totalHeight = 0;
    for (let i = startIdx; i < startIdx + count; i++) {
      const height = itemHeights.value.get(i) || getAverageHeight();
      totalHeight += height + ITEM_MARGIN;
    }
    return totalHeight - ITEM_MARGIN; // Subtract last margin
  };
  
  const isInRange = (index) => {
    const from = parseInt(fromIndex.value);
    const to = parseInt(toIndex.value);
  
    // Handle invalid inputs
    if (isNaN(from) || isNaN(to)) return true;
    if (from < 0 || to < 0) return true;
    if (from > entityData.value.length - 1 || to > entityData.value.length - 1) return true;
  
    return index >= from && index <= to;
  };
  
  const topSpacerHeight = computed(() => {
    const from = parseInt(fromIndex.value);
    if (isNaN(from) || from <= 0) return 0;
    return calculateHeight(0, from);
  });
  
  const bottomSpacerHeight = computed(() => {
    const to = parseInt(toIndex.value);
    if (isNaN(to) || to >= entityData.value.length - 1) return 0;
    const hiddenCount = entityData.value.length - 1 - to;
    return calculateHeight(to + 1, hiddenCount);
  });
  
  // Lifecycle hooks for measurements
  onMounted(() => {
    measureVisibleItems();
  });
  
  onUpdated(() => {
    measureVisibleItems();
  });
  </script>
  
  <style scoped>
  .accordion {
    width: 100%;
    border: 1px solid #ddd;
    border-radius: 4px;
    box-sizing: border-box;
  }
  
  .accordion-header {
    height: 50px;
    padding: 0 16px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    background-color: forestgreen;
    user-select: none;
    box-sizing: border-box;
  }
  
  .header-content {
    display: flex;
    align-items: center;
    gap: 20px;
    flex: 1;
  }
  
  .range-inputs {
    display: flex;
    gap: 10px;
  }
  
  .input-group {
    display: flex;
    align-items: center;
    gap: 5px;
  }
  
  .input-group input {
    width: 60px;
    height: 30px;
    padding: 0 8px;
    border: 1px solid #ddd;
    border-radius: 4px;
  }
  
  .accordion-icon {
    transition: transform 0.3s ease;
    font-size: 12px;
    cursor: pointer;
  }
  
  .accordion-icon.is-open {
    transform: rotate(180deg);
  }
  
  .accordion-content {
    height: 0;
    overflow: hidden;
    transition: height 0.3s ease;
    box-sizing: border-box;
  }
  
  .accordion-content.is-open {
    height: auto;
  }
  
  .accordion-items {
    padding: 5px;
    box-sizing: border-box;
  }
  
  .accordion-item {
    min-height: 50px; /* Minimum height, can grow based on content */
    padding: 16px;
    display: flex;
    align-items: center;
    background-color: crimson;
    border: 1px solid #eee;
    border-radius: 4px;
    margin-bottom: 5px;
    box-sizing: border-box;
  }
  
  .accordion-item:last-child {
    margin-bottom: 0;
  }
  
  .accordion-spacer {
    background-color: hotpink;
    box-sizing: border-box;
  }
  
  label {
    font-size: 14px;
  }
  </style>