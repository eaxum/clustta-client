<template>
  <div class="tabs-container">
    <div class="tab-bar">
      <TransitionGroup name="tab-list">
      <div v-for="(dataType, index) in dataTypes" :key="dataType.name" class="tab"
        :class="{ 'active': useSelected ? selectedTab === dataType.name : filterIndex === index, 'right-tab-split': rightTabPosition(index), 'left-tab-split': leftTabPosition(index) }"
        @click="filterList(index, dataType.name)"
        v-tooltip="((filterIndex !== index || iconsOnly) && useTooltip) ? utils.capitalizeStr(dataType.name) : ''">
        <div v-if="useSelected ? selectedTab === dataType.name : filterIndex === index" class="tab-gradient"></div>
        
        <div class="header-tab-meta" :class="{ 'icons-only': iconsOnly }" >
          <img v-if="dataType.icon" :src="getAppIcon(dataType.icon)" class="tab-favicon" alt="tab-icon">
          <span v-if="!iconsOnly && (filterIndex === index || fullWidth)" class="tab-title">{{ utils.capitalizeStr(dataType.name) }}</span>
          
          <!-- Alert indicators from HeaderTabs functionality -->
          <div v-if="useFunctions" class="alert-section">
            <div class="alert-items" v-if="alertItems(dataType.name).value !== 0 && displayCount"
              :class="{ 'alert-items-with-text': alertItems(dataType.name).value !== 0, 'critical-items': criticalItems(dataType.name).value }">
              {{ alertItems(dataType.name) }}
            </div>
            <div class="alert-items" v-else-if="alertItems(dataType.name).value"
              :class="{ 'critical-items': criticalItems(dataType.name).value }">
            </div>
          </div>
        </div>
      </div>
    </TransitionGroup>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useIconStore } from '@/stores/icons';
import { useTrayStates } from '@/stores/TrayStates';
import utils from '@/services/utils';

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

const filterIndex = ref(0);

const getAppIcon = (iconName) => {
  const formattedIconName = getIconName(iconName)
  const icon = iconStore.getAppIcon(formattedIconName);
  return icon
};

const getIconName = (path) => {
  if (!path.includes('/') && !path.includes('.svg')) {
    return path;
  }
  return path.split('/').pop().replace('.svg', '');
};

const filterList = (index, dataType) => {
  highlightFilter(index);
  emit('filter', dataType);
};

const highlightFilter = (index) => {
  filterIndex.value = index;
};

// Helper functions for tab positioning (similar to WorkspaceTabs)
const activeTabIndex = computed(() => {
  if (props.useSelected) {
    const selectedIndex = props.dataTypes.findIndex(item => item.name === props.selectedTab);
    return selectedIndex >= 0 ? selectedIndex : 0;
  }
  return filterIndex.value;
});

const rightTabPosition = (index) => {
  return activeTabIndex.value + 1 < index;
};

const leftTabPosition = (index) => {
  return activeTabIndex.value > index + 1;
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.tabs-container {
  width: 100%;
  height: 100%;
  overflow: hidden;
  position: relative;
  height: 40px;
  margin-top: .2rem;
  margin-bottom: .2rem;
}

.tabs-container::after {
  /* content: "";
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 1.5px;
  background-color: var(--surface-inverse);
  z-index: 0; */
}

.tab-bar {
  display: flex;
  padding: 0 .5rem;
  gap: 6px;
  /* padding: 0 20px; */
  height: 100%;
  box-sizing: border-box;
  position: relative;
  align-items: flex-end;
  align-items: center;
}

.tab {
  color: var(--text);
  display: flex;
  align-items: center;
  /* width: 200px; */
  /* min-width: 100px; */
  /* max-width: 200px; */
  height: 40px;
  box-sizing: border-box;
  user-select: none;
  position: relative;
  cursor: pointer;
  opacity: 0.5;
  /* padding-bottom: .2rem; */
}

.tab:hover {
  opacity: 1;
}

/* after active tab */
.right-tab-split::after {
  content: "";
  position: absolute;
  background-color: var(--surface-4);
  left: -4px;
  height: 16px;
  width: 1.5px;
}

.right-tab-split:hover+.tab::after {
  content: "";
  position: absolute;
  background-color: transparent;
  left: -4px;
  height: 16px;
  width: 1.5px;
}

.right-tab-split:hover::after {
  content: "";
  position: absolute;
  background-color: transparent;
  left: -4px;
  height: 16px;
  width: 1.5px;
}

/* before active tab */
.left-tab-split::before {
  content: "";
  position: absolute;
  background-color: var(--surface-4);
  right: -4px;
  height: 16px;
  width: 1.5px;
}

.tab:has(+ .left-tab-split:hover)::before {
  content: "";
  position: absolute;
  background-color: transparent;
  right: -4px;
  height: 16px;
  width: 1.5px;
}

.left-tab-split:hover::before {
  content: "";
  position: absolute;
  background-color: transparent;
  right: -4px;
  height: 16px;
  width: 1.5px;
}

.tab:hover {
  opacity: 1;
  color: var(--text);
  background: var(--selected-soft);
  border-radius: var(--normal-radius);
  border: 0px;
  height: 32px;
}

.header-tab-meta {
  display: flex;
  width: 100%;
  width: min-content;
  box-sizing: border-box;
  height: 100%;
  /* position: absolute; */
  z-index: 1;
  padding: .2rem .5rem;
  padding-left: 1rem;
  align-items: center;
  justify-content: space-between;
  /* left: -50%; */
  /* transform: translateX(50%); */
  /* background-color: crimson; */
}

.tab.active {
  color: var(--text);
  border-bottom: none;
  border-radius: 12px 12px 0px 0px;
  height: 100%;
  width: 100%;
  /* width: min-content; */
  /* height: 40px; */
  position: relative;
  /* background-color: var(--surface-3); */
  border-bottom: 0px;
  box-sizing: border-box;
  z-index: 2;
  opacity: 1;
}

.icons-only {
  width: 100%;
}

.tab.active:hover {
  background-color: transparent;
}

.tab-gradient {
  width: 100%;
  height: 60%;
  position: absolute;
  top: 0;
  right: 50%;
  transform: translateX(50%);
  /* background: linear-gradient(to bottom, var(--surface-2), transparent); */
  /* background-color: crimson; */
  border-radius: 16px 16px 0px 0px;
  /* box-sizing: border-box; */
  border: var(--medium-transparent-line);
  border-bottom: 0px;
  box-sizing: content-box;
}

/* Remove left border and left border radius for first tab */
.tab:first-child .tab-gradient {
  border-left: none;
  border-radius: 0px 16px 0px 0px;
}

/* Remove right border and right border radius for last tab */
.tab:last-child .tab-gradient {
  border-right: none;
  border-radius: 16px 0px 0px 0px;
}

.tab.active:not(:first-child)::before {
  content: "";
  position: absolute;
  background-color: transparent;
  /* background-color: crimson; */
  right: 100%;
  bottom: 0px;
  height: 20px;
  width: 100vw;
  border-bottom-right-radius: 16px;
  border-right: var(--medium-transparent-line);
  border-bottom: var(--medium-transparent-line);
  box-sizing: border-box;
  z-index: 1;
  pointer-events: none;
}

.tab.active:not(:last-child)::after {
  content: "";
  position: absolute;
  background-color: transparent;
  left: 100%;
  bottom: 0px;
  height: 20px;
  width: 100vw;
  border-bottom-left-radius: 12px;
  border-left: var(--medium-transparent-line);
  border-bottom: var(--medium-transparent-line);
  box-sizing: border-box;
  z-index: 1;
  pointer-events: none;
}

.tab-favicon {
  width: 16px;
  height: 16px;
  margin-right: 8px;
  pointer-events: none;
}

.tab-title {
  flex: 1;
  font-size: 16px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  pointer-events: none;
}

.alert-section {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
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
  top: -8px;
  right: -8px;
  border-radius: 10px;
  padding: 3px;
  font-size: 12px;
  color: var(--text);
}

.alert-items-with-text {
  outline: solid 1px rgb(236, 182, 3);
  width: 10px;
  height: 10px;
  top: -10px;
  right: -10px;
}

.critical-items {
  outline: solid 1px #bd2d2d;
  background-color: #bd2d2d;
}

.tab-list-move {
  /* transition: transform 0.2s ease; */
}
</style>