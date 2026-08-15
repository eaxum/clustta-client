<template>
  <div class="list-box-container" ref="listBoxContainer" :class="{ 
    'list-box-container-full' : fullWidth,
    'list-box-container-fixed' : fixedWidth
    }" >
    <div class="list-box-parent" :class="{ 'is-disabled': stage.operationActive || disabled, 'is-expanded': isExpanded}" ref="listBoxParent" @click="toggleList()">
      <div class="list-box-parent-content" @mouseenter="utils.handleHover($event)"
        @mouseleave="utils.resetScroll($event)">
        <div class="list-box-parent-text" :class="{ 'placeholder-text': isPlaceholder }" style="overflow: hidden; text-overflow: ellipsis; display: flex; align-items: center; gap: 0.5rem;">
          <img v-if="selectedItemIcon" :src="selectedItemIcon" class="listbox-icon small-icons" :class="selectedItemIconClass" v-tooltip="selectedItemIconTooltip" />
          {{ selectedListItem }}
        </div>
      </div>
      <span class="list-box-parent-chevron"><img class="small-icons chevron" src="/icons/chevron_down_white.svg"></span>
    </div>
    <Teleport to="#app">
      <div v-if="isExpanded" v-stop-propagation class="listbox-list-items-root"
        :style="{ top: listItemsAnchor + 'px', left: listItemsLeft + 'px', width: listItemsWidth + 'px', maxHeight: listItemMaxHeight + 'px' }">
        <div class="listbox-list-items">
          <div v-for="(item, index) in filteredItems" :key="getItemKey(item, index)" :value="getItemValue(item)" @click="selectItem(item)"
            class="listbox-item" :aria-disabled="isItemDisabled(item)" :class="{ 'listbox-item-closed': isUnique(getItemValue(item)) === true, 'listbox-item-selected': getItemValue(item) === props.selectedItem, 'listbox-item-disabled': isItemDisabled(item) }">
            <div class="listbox-item-text-mask" @mouseenter="startScrollText($event, index)"
              @mouseleave="stopScrollText($event)">
              <div class="listbox-item-text" :class="{ 'overflow-text': isHoveringIndex === index }" style="display: flex; align-items: center; gap: 0.5rem;">
                <img v-if="getItemIcon(item)" :src="getItemIcon(item)" class="listbox-icon small-icons" :class="getItemIconClass(item)" v-tooltip="getItemIconTooltip(item)" />
                {{ utils.capitalizeStr(getItemValue(item)) }}
              </div>
            </div>
            <div v-if="$slots.itemAction" class="listbox-item-action" @click.stop>
              <slot name="itemAction" :item="item" :value="getItemValue(item)" :close="closeList" />
            </div>
          </div>
        </div>
        <slot v-if="$slots.footer" name="footer" :close="closeList" />
      </div>
    </Teleport>
  </div>
</template>

<script setup>

// imports
import { computed, onMounted, watchEffect, ref, nextTick, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import utils from "@/services/utils";
import emitter from '@/lib/mitt';

// states
import { useStageStore } from '@/stores/stages';

import { useMenu } from '@/stores/menu';
const menu = useMenu();
const stage = useStageStore();
const { t } = useI18n();

// refs
const listBoxParent = ref(null);
const isHoveringIndex = ref(null);
const isExpanded = ref(false);

// Helper functions to handle both string arrays and object arrays
const isObjectArray = computed(() => {
  return props.items.length > 0 && typeof props.items[0] === 'object' && props.items[0] !== null;
});

const getItemValue = (item) => {
  if (typeof item === 'string') {
    return item;
  }
  return item.name || item.value || item.label || '';
};

const getItemIcon = (item) => {
  if (typeof item === 'string') {
    return null;
  }
  return item.icon || null;
};

const getItemIconTooltip = (item) => {
  if (typeof item === 'string') {
    return '';
  }
  return item.iconTooltip || '';
};

const getItemIconClass = (item) => {
  if (typeof item === 'string' || !item.iconTone) {
    return '';
  }
  return `listbox-icon-${item.iconTone}`;
};

const isItemDisabled = (item) => {
  return typeof item === 'object' && item !== null && item.disabled === true;
};

const getItemKey = (item, index) => {
  if (typeof item === 'string') {
    return item;
  }
  return item.id || item.name || index;
};

// computed properties
const listItemsBoundary = computed(() => menu.contextMenuBounds);
const listItemsAnchor = ref(0);
const listItemsLeft = ref(0);
const listItemsWidth = ref(0);
const listItemMaxHeight = ref(0);
const listItemsPaddingTop = ref(0);

// horizontal/vertical breathing room around the wrapped input
const WRAP_PAD_X = 6;
const WRAP_PAD_Y = 6;
const filteredItems = computed(() => {
  if (!props.items.length) return [];
  const selectedIdx = props.items.findIndex(item => getItemValue(item) === props.selectedItem);
  if (selectedIdx <= 0) return props.items;
  const reordered = props.items.slice();
  const [selected] = reordered.splice(selectedIdx, 1);
  reordered.unshift(selected);
  return reordered;
});

const isPlaceholder = computed(() => !props.selectedItem);

const selectedListItem = computed(() => { 
  return props.selectedItem ? utils.capitalizeStr(props.selectedItem) : props.placeHolder
});

const selectedItemIcon = computed(() => {
  if (!props.selectedItem || !isObjectArray.value) {
    return null;
  }
  const selectedObj = props.items.find(item => getItemValue(item) === props.selectedItem);
  return selectedObj ? getItemIcon(selectedObj) : null;
});

const selectedItemIconTooltip = computed(() => {
  if (!props.selectedItem || !isObjectArray.value) {
    return '';
  }
  const selectedObj = props.items.find(item => getItemValue(item) === props.selectedItem);
  return selectedObj ? getItemIconTooltip(selectedObj) : '';
});

const selectedItemIconClass = computed(() => {
  if (!props.selectedItem || !isObjectArray.value) {
    return '';
  }
  const selectedObj = props.items.find(item => getItemValue(item) === props.selectedItem);
  return selectedObj ? getItemIconClass(selectedObj) : '';
});

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
  disabled: { type: Boolean, default: false },
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
  if (props.disabled) return;

  const boundaryRect = listItemsBoundary.value ? listItemsBoundary.value.getBoundingClientRect() : null;
  const boundaryTop = boundaryRect ? boundaryRect.top : 0;
  const boundaryBottom = boundaryRect ? boundaryRect.bottom : window.innerHeight;
  const parentRect = listBoxParent.value.getBoundingClientRect();
  const listParentLeft = parentRect.left;
  const listParentTop = parentRect.top;
  const listParentBottom = parentRect.bottom;
  const listParentWidth = parentRect.width;

  const baseWidth = props.fullWidth ? listParentWidth : 200;
  const spaceBelow = boundaryBottom - listParentBottom;
  const spaceAbove = listParentTop - boundaryTop;
  // Flip up when there's clearly more room above than below (and below is too tight).
  const MIN_BELOW = 120;
  const flipUp = spaceBelow < MIN_BELOW && spaceAbove > spaceBelow;

  listItemsLeft.value = listParentLeft;
  listItemsWidth.value = baseWidth;
  listItemsPaddingTop.value = 0;

  if (flipUp) {
    const maxH = Math.max(80, spaceAbove + parentRect.height - 4);
    // Estimate actual list height so the panel hugs the parent and overlaps it
    // (mirrors the downward case where the panel anchors at listParentTop).
    const estimatedItemHeight = 32;
    const estimatedHeight = Math.min(maxH, filteredItems.value.length * estimatedItemHeight + 8);
    listItemMaxHeight.value = maxH;
    listItemsAnchor.value = listParentBottom - estimatedHeight;
  } else {
    listItemsAnchor.value = listParentTop;
    listItemMaxHeight.value = Math.max(80, boundaryBottom - listParentTop);
  }

  if (filteredItems.value.length) {
    isExpanded.value = !isExpanded.value;
  }
  else {
    isExpanded.value = false;
  }
};

const selectItem = (item) => {
  if (isItemDisabled(item)) return;
  const itemValue = getItemValue(item);
  props.onSelect(itemValue, props.extraData);
  isExpanded.value = false;
};

const closeList = () => {
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
  z-index: 1000;
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
  color: var(--text);
  display: flex;
  flex-direction: row;
  width: 100%;
  border-radius: var(--large-radius);
  height: 35px;
  align-items: center;
  padding: 6px;
  overflow: hidden;
  font-family: Inter, sans-serif;
  font-size: 16px;
  white-space: nowrap;
  justify-content: space-between;
  gap: .5rem;
  flex: 1;
  min-height: 35px;
  background-color: var(--bg);
  outline-offset: -1px;
  /* background-color: crimson; */
}

.list-box-parent:hover {
    outline: var(--transparent-line);
    background-color: var(--surface-2);
}

.list-box-parent.is-expanded {
    position: relative;
    z-index: 100001;
    background-color: transparent;
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
  font-weight: 400;
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
  color: var(--text);
  box-sizing: border-box;
  z-index: 100000;
  border-radius: var(--large-radius);
  min-height: 32px;
  line-height: 1.4 !important;
  background-color: var(--surface-2);
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
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.35);
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
  color: var(--text);
  box-sizing: border-box;
  border-radius: var(--normal-radius);
  min-height: min-content;
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
  border-radius: var(--normal-radius);
  width: max-content;
  width: 100%;
  /* height: 50px; */
  display: flex;
  align-items: center;
  justify-content: space-between;
  /* background-color: red; */
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.listbox-item-action {
  display: flex;
  align-items: center;
  padding: 0 .25rem;
  flex-shrink: 0;
}

.listbox-item:hover {
  background-color: var(--surface-4);
}

.listbox-item-selected {
  background-color: var(--surface-4);
}

.listbox-item-selected .listbox-item-text {
  font-weight: 500;
}

.listbox-item-closed {
  opacity: .5;
}

.listbox-item-disabled {
  cursor: default;
  opacity: .5;
}

.listbox-item-disabled:hover {
  background-color: transparent;
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

.listbox-icon {
  width: 16px;
  height: 16px;
  min-width: 16px;
  object-fit: contain;
  flex-shrink: 0;
}

img.listbox-icon.listbox-icon-alert {
  filter: brightness(0) saturate(100%) invert(60%) sepia(72%) saturate(489%) hue-rotate(1deg) brightness(92%) contrast(90%);
}

[data-theme="dark"] img.listbox-icon.listbox-icon-alert {
  filter: brightness(0) saturate(100%) invert(88%) sepia(45%) saturate(566%) hue-rotate(359deg) brightness(97%) contrast(92%);
}

img.listbox-icon.listbox-icon-go {
  filter: brightness(0) saturate(100%) invert(50%) sepia(74%) saturate(486%) hue-rotate(75deg) brightness(96%) contrast(87%);
}

.placeholder-text {
  font-style: italic;
  opacity: 0.6;
}

.list-box {
  box-sizing: border-box;
  background-color: var(--surface-inverse);
  width: 100%;
  border-radius: 8px;
  height: 35px;
  padding-right: 8px;
  overflow: hidden;
}
</style>


