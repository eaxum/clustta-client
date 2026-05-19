<template>
  <div ref="tabRoot" class="header-tab-root" :class="{ 'fullwidth-header-tab-root': fullWidth, 'icon-header-tab-root': iconsOnly }">

    <div ref="measureContainer" class="tab-measure-container">
      <div v-for="(dataType, index) in dataTypes" :key="'m-' + (dataType.id || dataType.name)" class="tab-button tab-measure-item">
        <div class="tab-content">
          <div class="selected-tab-button-text">{{ dataType.nameKey ? $t(dataType.nameKey) : utils.capitalizeStr(dataType.name) }}</div>
          <img class="small-icons" :src="getAppIcon(dataType.icon)">
        </div>
      </div>
    </div>

    <div v-for="(dataType, index) in visibleTabs" :key="dataType.id || dataType.name"
      v-tooltip="((filterIndex !== index || iconsOnly || hideIcons) && useTooltip) ? (dataType.nameKey ? $t(dataType.nameKey) : utils.capitalizeStr(dataType.name)) : ''"
      @click="filterList(dataType._originalIndex, dataType.id || dataType.name)" class="tab-button"
      :class="{ 'selected-tab-button': useSelected ? selectedTab === (dataType.id || dataType.name) : filterIndex === dataType._originalIndex, 'fullwidth-tab-button': fullWidth }">
      <div v-if="useFunctions">
        <div class="alert-items" v-if="alertItems(dataType.id || dataType.name).value !== 0 && displayCount"
          :class="{ 'alert-items-with-text': alertItems(dataType.id || dataType.name).value !== 0, 'critical-items': criticalItems(dataType.id || dataType.name).value }">
          {{ alertItems(dataType.id || dataType.name) }}
        </div>
        <div class="alert-items" v-else-if="alertItems(dataType.id || dataType.name).value"
          :class="{ 'critical-items': criticalItems(dataType.id || dataType.name).value }">
        </div>
      </div>
      <div class="tab-content">
        <div v-if="!iconsOnly && (filterIndex === dataType._originalIndex || fullWidth)" class="selected-tab-button-text">{{ dataType.nameKey ? $t(dataType.nameKey) : utils.capitalizeStr(dataType.name) }}</div>
        <img v-if="!hideIcons" class="small-icons" :src="getAppIcon(dataType.icon)">
      </div>
    </div>

    <div v-if="overflowTabs.length" class="tab-button more-button" :class="{ 'selected-tab-button': isOverflowTabSelected }" @click.stop="toggleOverflow">
      <div class="tab-content">
        <div class="selected-tab-button-text">{{ $t('common.more') || 'More' }}</div>
        <img class="small-icons more-chevron" :class="{ 'more-chevron-open': showOverflow }" :src="getAppIcon('chevron-down')">
      </div>
    </div>

    <Teleport to="#app">
      <div v-if="showOverflow" class="overflow-backdrop" @click="showOverflow = false"></div>
      <div v-if="showOverflow" ref="overflowMenu" class="overflow-menu" :style="overflowMenuStyle">
        <div v-for="dataType in overflowTabs" :key="'o-' + (dataType.id || dataType.name)"
          @click="selectOverflowTab(dataType._originalIndex, dataType.id || dataType.name)" class="overflow-menu-item"
          :class="{ 'overflow-menu-item-selected': useSelected ? selectedTab === (dataType.id || dataType.name) : filterIndex === dataType._originalIndex }">
          <img class="small-icons" :src="getAppIcon(dataType.icon)">
          <span>{{ dataType.nameKey ? $t(dataType.nameKey) : utils.capitalizeStr(dataType.name) }}</span>
          <div v-if="useFunctions && alertItems(dataType.id || dataType.name).value" class="overflow-alert-dot"
            :class="{ 'critical-items': criticalItems(dataType.id || dataType.name).value }">
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<script setup>
// imports
import { ref, computed, onMounted, watch, nextTick, onBeforeUnmount } from 'vue';
import utils from '@/services/utils';

// stores
import { useIconStore } from '@/stores/icons';
import { useTrayStates } from '@/stores/TrayStates';
const iconStore = useIconStore();
const trayStates = useTrayStates();

const emit = defineEmits(['filter']);

const props = defineProps({
  dataTypes: { type: Array, default: () => [] },
  alertItems: {
    type: Function,
    default: (dummy) => {
      return dummy = { item: 0 }
    }
  },
  criticalItems: {
    type: Function,
    default: (dummy) => {
      return dummy = { item: 0 }
    }
  },
  useFunctions: { type: Boolean, default: false },
  forWorkspace: { type: Boolean, default: false },
  displayCount: { type: Boolean, default: true },
  fullWidth: { type: Boolean, default: false },
  iconsOnly: { type: Boolean, default: false },
  useSelected: { type: Boolean, default: false },
  useTooltip: { type: Boolean, default: false },
  selectedTab: { type: String, default: '' },
});

// refs
const filterIndex = ref(0);
const hideIcons = ref(false);
const measureContainer = ref(null);
const moreButtonRef = ref(null);
const overflowMenu = ref(null);
const showOverflow = ref(false);
const splitIndex = ref(-1);
const tabRoot = ref(null);

// computed properties
const isOverflowTabSelected = computed(() => {
  return overflowTabs.value.some(dt => {
    const key = dt.id || dt.name;
    return props.useSelected ? props.selectedTab === key : filterIndex.value === dt._originalIndex;
  });
});

// Adds _originalIndex to each dataType so we can track the real index after splitting.
const indexedDataTypes = computed(() => {
  return props.dataTypes.map((dt, i) => ({ ...dt, _originalIndex: i }));
});

const overflowMenuStyle = computed(() => {
  if (!tabRoot.value) return {};
  const root = tabRoot.value.getBoundingClientRect();
  return {
    position: 'fixed',
    top: `${root.bottom + 4}px`,
    right: `${window.innerWidth - root.right}px`,
  };
});

const overflowTabs = computed(() => {
  if (splitIndex.value < 0 || splitIndex.value >= indexedDataTypes.value.length) return [];
  return indexedDataTypes.value.slice(splitIndex.value);
});

const visibleTabs = computed(() => {
  if (splitIndex.value < 0) return indexedDataTypes.value;
  return indexedDataTypes.value.slice(0, splitIndex.value);
});

// methods
const calculateOverflow = () => {
  if (!tabRoot.value || !measureContainer.value) return;

  const rootWidth = tabRoot.value.getBoundingClientRect().width;
  const measureItems = measureContainer.value.children;
  if (!measureItems.length) return;

  const gap = 8;
  const moreButtonWidth = 80;
  let totalWidth = 0;
  let firstHideIcons = -1;
  let firstOverflow = -1;

  // Measure each tab at full size (with text + icon).
  for (let i = 0; i < measureItems.length; i++) {
    totalWidth += measureItems[i].getBoundingClientRect().width + gap;
  }

  // Phase 1: Everything fits — no changes needed.
  if (totalWidth <= rootWidth) {
    splitIndex.value = -1;
    hideIcons.value = false;
    return;
  }

  // Phase 2: Try hiding icons to reclaim space.
  // Each icon is roughly 24px (20px + 4px padding). Estimate savings.
  const iconSavings = 24;
  let noIconTotal = 0;
  for (let i = 0; i < measureItems.length; i++) {
    noIconTotal += (measureItems[i].getBoundingClientRect().width - iconSavings) + gap;
  }

  if (noIconTotal <= rootWidth) {
    splitIndex.value = -1;
    hideIcons.value = true;
    return;
  }

  // Phase 3: Even without icons it doesn't fit — need overflow menu.
  hideIcons.value = true;
  let usedWidth = moreButtonWidth;
  firstOverflow = -1;

  for (let i = 0; i < measureItems.length; i++) {
    const itemWidth = (measureItems[i].getBoundingClientRect().width - iconSavings) + gap;
    if (usedWidth + itemWidth > rootWidth) {
      firstOverflow = i;
      break;
    }
    usedWidth += itemWidth;
  }

  splitIndex.value = firstOverflow !== -1 ? firstOverflow : measureItems.length;
};

// Filters the tab list and emits the selected tab id.
const filterList = (index, dataType) => {
  highlightFilter(index);
  emit('filter', dataType);
};

const getAppIcon = (iconName) => {
  const formattedIconName = getIconName(iconName);
  return iconStore.getAppIcon(formattedIconName);
};

const getIconName = (path) => {
  if (!path.includes('/') && !path.includes('.svg')) return path;
  return path.split('/').pop().replace('.svg', '');
};

// Sets the visual highlight to the given tab index.
const highlightFilter = (index) => {
  filterIndex.value = index;
};

// Selects a tab from the overflow dropdown.
const selectOverflowTab = (index, dataType) => {
  showOverflow.value = false;
  filterList(index, dataType);
};

// Toggles the overflow dropdown open/closed.
const toggleOverflow = () => {
  showOverflow.value = !showOverflow.value;
};

// watchers
watch(() => props.dataTypes, () => {
  nextTick(calculateOverflow);
}, { deep: true });

let resizeObserver = null;

// lifecycle hooks
onMounted(() => {
  nextTick(() => {
    calculateOverflow();
    // Fallback: recalculate after fonts/icons may have loaded.
    setTimeout(calculateOverflow, 100);
  });
  resizeObserver = new ResizeObserver(() => {
    calculateOverflow();
  });
  if (tabRoot.value) resizeObserver.observe(tabRoot.value);
});

onBeforeUnmount(() => {
  if (resizeObserver) resizeObserver.disconnect();
});
</script>

<style scoped>
.header-tab-root {
  display: flex;
  flex-direction: row;
  box-sizing: border-box;
  width: 100%;
  justify-content: flex-start;
  align-items: center;
  overflow: visible;
  height: max-content;
  gap: .5rem;
  padding: .3rem .3rem;
  color: var(--text);
  position: relative;
}

.fullwidth-header-tab-root {
  /* overflow logic handles layout — root stays width: 100% */
}

.icon-header-tab-root {
  width: min-content;
}

.tab-measure-container {
  position: fixed;
  top: -9999px;
  left: -9999px;
  visibility: hidden;
  pointer-events: none;
  display: flex;
  gap: .5rem;
  white-space: nowrap;
}

.selected-tab-button-text {
  padding: .2rem .1rem;
  font-weight: 250;
}

.tab-button {
  position: relative;
  border-radius: 8px;
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  align-items: center;
  height: max-content;
  opacity: .5;
  justify-content: center;
  padding: 5px .5rem;
  border-radius: var(--large-radius);
  transition: all .2s ease-in-out;
  white-space: nowrap;
  flex-shrink: 0;
}

.tab-button:hover {
  background-color: var(--surface-4);
  opacity: 1;
}

.tab-button:active {
  opacity: 1;
}

.tab-button-pressed {
  box-sizing: border-box;
  background-color: rgba(0, 0, 0, 0.216);
  outline: solid 1px var(--border-strong);
  outline-offset: -1px;
}

.selected-tab-button {
  background-color: var(--surface-1);
  outline: var(--transparent-line);
  outline-offset: -1px;
  opacity: 1;
}

.fullwidth-tab-button {
  width: max-content;
}

.selected-tab-button:hover {
  background-color: var(--surface-1);
}

.tab-content {
  display: flex;
  gap: .5rem;
  white-space: nowrap;
  align-items: center;
  font-size: 14px;
}

.more-button {
  flex-shrink: 0;
  opacity: .7;
}

.more-button:hover {
  opacity: 1;
}

.more-chevron {
  transition: transform .2s ease;
}

.more-chevron-open {
  transform: rotate(180deg);
}

.overflow-backdrop {
  position: fixed;
  inset: 0;
  z-index: 99998;
}

.overflow-menu {
  z-index: 99999;
  display: flex;
  flex-direction: column;
  gap: .2rem;
  padding: .4rem;
  min-width: 160px;
  border-radius: var(--large-radius);
  background-color: var(--surface-2);
  outline: var(--transparent-line);
  outline-offset: -1px;
  backdrop-filter: blur(55px);
  animation: overflowFadeIn .15s ease-out;
}

@keyframes overflowFadeIn {
  from {
    opacity: 0;
    transform: translateY(-6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.overflow-menu-item {
  display: flex;
  align-items: center;
  gap: .5rem;
  padding: .4rem .6rem;
  border-radius: var(--normal-radius);
  cursor: pointer;
  font-size: 14px;
  white-space: nowrap;
  transition: background-color .15s ease;
  color: var(--text);
}

.overflow-menu-item:hover {
  background-color: var(--surface-4);
}

.overflow-menu-item-selected {
  background-color: var(--surface-1);
}

.overflow-alert-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: #ecb603;
  margin-left: auto;
}

.alert-items {
  overflow: hidden;
  width: 3px;
  height: 3px;
  background-color: #ecb603;
  border-radius: 5px;
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: center;
  top: -3px;
  right: -3px;
  border-radius: 10px;
  padding: 3px;
  font-size: 12px;
  color: var(--text);
}

.alert-items-with-text {
  outline: solid 1px rgb(236, 182, 3);
  width: 10px;
  height: 10px;
  top: -5px;
  right: -5px;
}

.critical-items {
  outline: solid 1px #bd2d2d;
  background-color: #bd2d2d;
}
</style>


