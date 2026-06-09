<template>
  <div class="list-box-container" ref="listBoxContainer" :class="{ 
    'list-box-container-full' : fullWidth,
    'list-box-container-fixed' : fixedWidth
    }" >
    <div class="list-box-parent" :class="{ 'is-disabled': stage.operationActive}" ref="listBoxParent" @click="toggleList()">
      <div class="list-box-parent-content" @mouseenter="utils.handleHover($event)"
        @mouseleave="utils.resetScroll($event)">
        <div class="list-box-parent-text" :class="{ 'placeholder-text': isPlaceholder }" style="overflow: hidden; text-overflow: ellipsis; display: flex; align-items: center; gap: 0.5rem;">
          <img v-if="selectedItemIcon" :src="selectedItemIcon" class="listbox-icon" />
          {{ selectedListItem }}
        </div>
      </div>
      <span class="list-box-parent-chevron"><img class="small-icons chevron" src="/icons/chevron_down_white.svg"></span>
    </div>
    <Teleport to="#app">
      <div v-if="isExpanded" v-stop-propagation class="listbox-list-items-root"
        :style="{ top: listItemsAnchor + 'px', left: listItemsLeft + 'px', width: listItemsWidth + 'px', maxHeight: listItemMaxHeight + 'px' }">
        <div class="listbox-list-items">
          <div v-for="(item, index) in filteredItems" :key="getItemKey(item, index)" :value="getItemValue(item)" @click="selectItem(item, items)"
            class="listbox-item" :class="{ 'listbox-item-closed': isUnique(getItemValue(item)) === true }">
            <div class="listbox-item-text-mask" @mouseenter="startScrollText($event, index)"
              @mouseleave="stopScrollText($event)">
              <div class="listbox-item-text" :class="{ 'overflow-text': isHoveringIndex === index }" style="display: flex; align-items: center; gap: 0.5rem;">
                <img v-if="getItemIcon(item)" :src="getItemIcon(item)" class="listbox-icon" />
                {{ utils.capitalizeStr(getItemValue(item)) }}
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
const filteredItems = computed(() => {
  if (props.useFilter) {
    return props.items.length ? props.items.filter(item => {
      const itemValue = getItemValue(item);
      return itemValue !== props.selectedItem;
    }) : []
  } else {
    return props.items.length ? props.items : []
  }
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
  const itemValue = getItemValue(item);
  props.onSelect(itemValue, props.extraData);
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
  animation: fadeIn .15s ease-in-out forwards;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-4px);
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
}

.list-box-container-full {
  width: 100%;
}

.list-box-container-fixed {
  width: 100%;
  max-width: 200px;
}

.list-box-parent {
  position: relative;
  box-sizing: border-box;
  color: hsl(var(--foreground));
  display: flex;
  flex-direction: row;
  width: 100%;
  border-radius: calc(var(--radius) - 2px);
  height: 36px;
  align-items: center;
  padding: 0.5rem 0.75rem;
  overflow: hidden;
  font-family: Inter, sans-serif;
  font-size: 0.875rem;
  white-space: nowrap;
  justify-content: space-between;
  gap: .5rem;
  flex: 1;
  min-height: 36px;
  border: 1px solid hsl(var(--input));
  background-color: transparent;
  transition: border-color 0.15s ease;
  cursor: pointer;
}

.list-box-parent:hover {
  border-color: hsl(var(--ring));
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
  font-size: 0.875rem;
  width: 100%;
  font-weight: 400;
}

.list-box-parent-chevron {
  pointer-events: none;
  opacity: 0.5;
}

.chevron {
  height: 10px;
  min-width: 10px;
}

.listbox-list-items-root {
  color: hsl(var(--popover-foreground));
  box-sizing: border-box;
  z-index: 100000;
  border-radius: calc(var(--radius) - 2px);
  min-height: 32px;
  line-height: 1.4 !important;
  background-color: hsl(var(--popover));
  border: 1px solid hsl(var(--border));
  overflow: hidden;
  overflow-y: auto;
  max-height: 300px;
  text-align: left;
  flex-direction: column;
  flex-wrap: nowrap;
  gap: 0.25rem;
  padding: 0.25rem;
  position: absolute;
  width: min-content;
  box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
}

.listbox-list-items-root::-webkit-scrollbar {
  border-radius: var(--tiny-radius);
  width: 4px;
}

.listbox-list-items-root::-webkit-scrollbar-thumb {
  border-radius: var(--tiny-radius);
  background-color: hsl(var(--border));
}

.listbox-list-items-root::-webkit-scrollbar-track {
  margin: 10px;
  border-radius: var(--tiny-radius);
}

.listbox-list-items {
  color: hsl(var(--popover-foreground));
  box-sizing: border-box;
  border-radius: calc(var(--radius) - 4px);
  min-height: min-content;
  overflow: hidden;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  flex-wrap: nowrap;
  height: max-content;
  gap: 0.125rem;
}

.listbox-item {
  box-sizing: border-box;
  list-style: none;
  cursor: pointer;
  background-color: transparent;
  transition: background-color 0.15s ease;
  border-radius: calc(var(--radius) - 4px);
  width: 100%;
  display: flex;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.listbox-item:hover {
  background-color: hsl(var(--accent));
  color: hsl(var(--accent-foreground));
}

.listbox-item-closed {
  opacity: .5;
}

.listbox-item-text-mask {
  padding: 0.125rem;
}

.listbox-item-text {
  padding: 0.25rem 0.5rem;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 0.875rem;
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

.placeholder-text {
  color: hsl(var(--muted-foreground));
  font-style: italic;
  opacity: 1;
}

.list-box {
  box-sizing: border-box;
  background-color: hsl(var(--background));
  width: 100%;
  border-radius: calc(var(--radius) - 2px);
  height: 36px;
  padding-right: 8px;
  overflow: hidden;
}
</style>


