<template>
  <div class="list-box-container" ref="listBoxContainer" :class="{ 
    'list-box-container-full' : fullWidth,
    'list-box-container-fixed' : fixedWidth
    }" >
    <div class="list-box-parent" :class="{ 'is-disabled': stage.operationActive}" ref="listBoxParent" @click="toggleList()">
      <div class="list-box-parent-content" @mouseenter="utils.handleHover($event)"
        @mouseleave="utils.resetScroll($event)">
        <div class="list-box-parent-text" style="overflow: hidden; text-overflow: ellipsis;">
          {{ selectedListItem }}
        </div>
      </div>
      <span class="list-box-parent-chevron"><img class="small-icons chevron" src="/icons/chevron_down_white.svg"></span>
    </div>
    <Teleport to="#app">
      <div v-if="isExpanded" v-stop-propagation class="listbox-list-items-root"
        :style="{ top: listItemsAnchor + 'px', left: listItemsLeft + 'px', width: listItemsWidth + 'px', maxHeight: listItemMaxHeight + 'px' }">
        <div class="listbox-list-items">
          <div v-for="(item, index) in filteredItems" :key="item" :value="item" @click="selectItem(item, items)"
            class="listbox-item" :class="{ 'listbox-item-closed': isUnique(item) === true }">
            <div class="listbox-item-text-mask" @mouseenter="startScrollText($event, index)"
              @mouseleave="stopScrollText($event)">
              <div class="listbox-item-text" :class="{ 'overflow-text': isHoveringIndex === index }">
                {{ utils.capitalizeStr(item) }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>

// imports
import { computed, onMounted, watchEffect, ref, nextTick, onUnmounted } from 'vue';
import utils from "@/services/utils";
import emitter from '@/lib/mitt';

// states
import { useStageStore } from '@/stores/stages';

import { useMenu } from '@/stores/menu';
const menu = useMenu();
const stage = useStageStore();

// refs
const listBoxParent = ref(null);
const isHoveringIndex = ref(null);
const isExpanded = ref(false);

// computed properties
const listItemsBoundary = computed(() => menu.contextMenuBounds);
const listItemsAnchor = ref(0);
const listItemsLeft = ref(0);
const listItemsWidth = ref(0);
const listItemMaxHeight = ref(0);
const filteredItems = computed(() => {
  if (props.useFilter) {
    return props.items.length ? props.items.filter(item => item !== props.selectedItem) : ''
  } else {
    return props.items.length ? props.items : ''
  }
});
const selectedListItem = computed(() => { return props.selectedItem ? utils.capitalizeStr(props.selectedItem) : 'Select' })
// props
const props = defineProps({
  isUnique: {
    type: Function,
    default: () => {
      return false
    }
  },
  useFilter: { type: Boolean, default: true },
  fullWidth: { type: Boolean, default: true },
  fixedWidth: { type: Boolean, default: false },
  items: Array,
  onSelect: Function,
  selectedItem: String,
  extraData: {
    type: Object,
    default: {}
  },
  placeHolder: {
    type: String,
    default: 'No Item'
  }
});

// methods
const startScrollText = (event, index) => {
  isHoveringIndex.value = index;
  nextTick(() => {
    utils.handleHover(event);
  });
};

const stopScrollText = (event) => {
  isHoveringIndex.value = null;
  nextTick(() => {
    utils.resetScroll(event);
  });
};

const toggleList = () => {
  const containerHeight = listItemsBoundary.value ? listItemsBoundary.value.getBoundingClientRect().height : 200;
  const listParentLeft = listBoxParent.value.getBoundingClientRect().left;
  const listParentGlobalY = listBoxParent.value.getBoundingClientRect().top;
  const listParentHeight = listBoxParent.value.getBoundingClientRect().height;

  listItemsLeft.value = listParentLeft;
  listItemsAnchor.value = listParentGlobalY + listParentHeight + 5;
  listItemsWidth.value = props.fullWidth ? listBoxParent.value.getBoundingClientRect().width : 200;
  listItemMaxHeight.value = containerHeight - listParentHeight - listParentGlobalY;
  // listItemMaxHeight.value = 200;

  if (filteredItems.value.length) {
    isExpanded.value = !isExpanded.value;
  }
  else {
    isExpanded.value = false;
  }
};

const selectItem = (item, items) => {
  props.onSelect(item, props.extraData);
  isExpanded.value = false;
};

const hideListContent = (event) => {
  if (isExpanded.value && (event.target !== listBoxParent.value)) {
    isExpanded.value = false;
  }
};

const disableListBoxOnScroll = (event) => {
  if(isExpanded.value){
    isExpanded.value = false
  }
};

watchEffect(() => {
	if (menu.clickOutsideMask) {
    menu.clickOutsideMask.addEventListener('click', hideListContent);
	}
});

// onMounted hook
onMounted(() => {
  emitter.on('disableListBoxOnScroll', disableListBoxOnScroll);
});

onUnmounted(() => {
  emitter.off('disableListBoxOnScroll', disableListBoxOnScroll);
	if (menu.clickOutsideMask) {
    menu.clickOutsideMask.removeEventListener('click', hideListContent);
	}
});

</script>

<style scoped>
@import "@/assets/tray.css";

.listbox-list-items-root {
  opacity: 0;
  animation: fadeIn .1s ease-in-out forwards;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.list-box-container {
  box-sizing: border-box;
  display: flex;
  min-width: 100px;
  width: min-content;
  flex-direction: column;
  /* flex: 1; */
  
}

.list-box-container-full {
  width: 100%;
  /* max-width: 200px; */
}

.list-box-container-fixed {
  width: 100%;
  max-width: 200px;
}

.list-box-parent {
  position: relative;
  box-sizing: border-box;
  color: black;
  color: var(--white);
  display: flex;
  flex-direction: row;
  width: 100%;
  border-radius: 10px;
  height: 35px;
  align-items: center;
  padding: 4px;
  overflow: hidden;
  font-family: Inter, sans-serif;
  font-size: 16px;
  white-space: nowrap;
  justify-content: space-between;
  gap: .5rem;
  flex: 1;
  min-height: 35px;
  background-color: var(--midnight-steel);
  /* background-color: crimson; */
}

.list-box-parent:hover {
    outline: var(--transparent-line);
    background-color: var(--dark-steel);
}

.list-box-parent-content {
  position: relative;
  box-sizing: border-box;
  height: 20px;
  flex: 1;
  width: 100%;
  white-space: nowrap;
  pointer-events: none;
  overflow: hidden;
}

.list-box-parent-text {
  font-family: 'Inter', sans-serif;
  position: absolute;
  pointer-events: none;
  white-space: nowrap;
  padding-left: .5rem;
  font-size: 14px;
  width: 100%;
}

.list-box-parent-chevron {
  pointer-events: none;
}

.chevron {
  height: 10px;
  min-width: 10px;
}

.listbox-list-items-root {
  color: black;
  color: var(--white);
  box-sizing: border-box;
  z-index: 100000;
  border-radius: 8px;
  min-height: 32px;
  line-height: 1.4 !important;
  background-color: var(--dark-steel);
  overflow: hidden;
  overflow-y: auto;
  max-height: 300px;
  text-align: left;
  flex-direction: column;
  flex-wrap: nowrap;
  gap: .2rem;
  padding: .3rem .3rem;
  outline-offset: -1px;
  position: absolute;
  outline: var(--transparent-line);
  width: min-content;
}


.listbox-list-items-root::-webkit-scrollbar {
  border-radius: 4px;
  width: 4px;
}

.listbox-list-items-root::-webkit-scrollbar-thumb {
  border-radius: 4px;
  background-color: rgba(255, 255, 255, 0.295);
}

.listbox-list-items-root::-webkit-scrollbar-track {
    margin: 10px;
    border-radius: 4px;
    background-color: rgba(0, 0, 0, 0.295);
}

.listbox-list-items {
  color: var(--white);
  box-sizing: border-box;
  border-radius: 8px;
  min-height: 32px;
  /* background-color: #2e2e2e; */
  overflow: hidden;
  overflow-y: auto;
  display: flex;
  /* flex: 1; */
  flex-direction: column;
  flex-wrap: nowrap;
  height: max-content;
  gap: .2rem;
  /* padding: .3rem .3rem;  */
  /* background-color: blue; */
}

.listbox-item {
  box-sizing: border-box;
  list-style: none;
  cursor: pointer;
  background-color: transparent;
  transition: background-color 0.2s ease-in-out;
  border-radius: 8px;
  width: max-content;
  width: 100%;
  /* height: 50px; */
  display: flex;
  /* background-color: red; */
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.listbox-item:hover {
  background-color: var(--light-steel);
}

.listbox-item-closed {
  opacity: .5;
}

.listbox-item-text-mask {
  padding: .1rem .1rem;
}

.listbox-item-text {
  padding: 0.2rem .4rem;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 14px;

}

.overflow-text {
  overflow: unset;
}

.list-box {
  box-sizing: border-box;
  background-color: var(--white);
  width: 100%;
  border-radius: 8px;
  height: 35px;
  padding-right: 8px;
  overflow: hidden;
}
</style>
