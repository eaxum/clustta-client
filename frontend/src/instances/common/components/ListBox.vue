<template>
  <div class="list-box-container" ref="listBoxContainer">
    <div class="list-box-parent" ref="listBoxParent" @click="toggleList()">
      <div class="list-box-parent-content" @mouseenter="trayStates.handleHover($event)"
        @mouseleave="trayStates.resetScroll($event)">
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
import { useTrayStates } from '@/stores/TrayStates';
import { computed, onMounted, ref, nextTick, onBeforeUnmount } from 'vue';
import utils from "@/services/utils";

// states
const trayStates = useTrayStates();

// refs
const listBoxParent = ref(null);
const isHoveringIndex = ref(null);
const isExpanded = ref(false);

// computed properties
const listItemsBoundary = computed(() => trayStates.listItemsBoundary);
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
const selectedListItem = computed(() => { return props.selectedItem ? utils.capitalizeStr(props.selectedItem) : 'Nothing Selected' })
// props
const props = defineProps({
  isUnique: {
    type: Function,
    default: () => {
      return false
    }
  },
  useFilter: { type: Boolean, default: true },
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
    trayStates.handleHover(event);
  });
};

const stopScrollText = (event) => {
  isHoveringIndex.value = null;
  nextTick(() => {
    trayStates.resetScroll(event);
  });
};

const toggleList = () => {
  const containerHeight = listItemsBoundary.value ? listItemsBoundary.value.getBoundingClientRect().height : 200;
  const listParentLeft = listBoxParent.value.getBoundingClientRect().left;
  const listParentGlobalY = listBoxParent.value.getBoundingClientRect().top;
  const listParentHeight = listBoxParent.value.getBoundingClientRect().height;

  listItemsLeft.value = listParentLeft;
  listItemsAnchor.value = listParentGlobalY + listParentHeight + 5;
  listItemsWidth.value = listBoxParent.value.getBoundingClientRect().width;
  listItemMaxHeight.value = containerHeight - listParentHeight - listParentGlobalY;

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

const handleClickOutside = (event) => {
  if (isExpanded.value && (event.target !== listBoxParent.value)) {
    isExpanded.value = false;
  }
};

// onMounted hook
onMounted(() => {
  document.addEventListener('click', handleClickOutside);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside);
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
  width: 100%;
  flex-direction: column;
  flex: 1;
}

.list-box-parent {
  position: relative;
  box-sizing: border-box;
  color: black;
  color: var(--white);
  display: flex;
  flex-direction: row;
  width: 100%;
  border-radius: 8px;
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
  background-color: var(--light-steel)
}

.list-box-parent:hover {
  background-color: var(--steel)
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
  color: var(--white);
  box-sizing: border-box;
  z-index: 100000;
  border-radius: 8px;
  min-height: 32px;
  line-height: 1.4 !important;
  background-color: var(--steel);
  overflow: hidden;
  overflow-y: auto;
  max-height: 300px;
  text-align: left;
  flex-direction: column;
  flex-wrap: nowrap;
  gap: .2rem;
  padding: .3rem .3rem;
  outline: var(--transparent-line);
  outline-offset: -1px;
  position: absolute;
}


.listbox-list-items-root::-webkit-scrollbar {
  border-radius: 8px;
  width: 8px;
}

.listbox-list-items-root::-webkit-scrollbar-thumb {
  border-radius: 8px;
  background-color: rgba(255, 255, 255, 0.295);
}

.listbox-list-items-root::-webkit-scrollbar-track {
  border-radius: 8px;
  background-color: rgba(0, 0, 0, 0.295);
}

.listbox-list-items {
  color: var(--white);
  box-sizing: border-box;
  border-radius: 8px;
  min-height: 32px;
  background-color: var(--steel);
  overflow: hidden;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  flex-wrap: nowrap;
  height: max-content;
  gap: .2rem;
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
  display: flex;
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
