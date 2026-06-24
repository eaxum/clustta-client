<template>
  <div class="header-tab-root"
    :class="{ 'fullwidth-header-tab-root': fullWidth, 'icon-header-tab-root': iconsOnly }">

    <div v-for="(dataType, index) in dataTypes" :key="dataType.id || dataType.name"
      v-tooltip="((filterIndex !== index || iconsOnly)) ? (dataType.nameKey ? $t(dataType.nameKey) : utils.capitalizeStr(dataType.name)) : ''"
      @click="filterList(index, dataType.id || dataType.name)" class="tab-button"
      :class="{ 'selected-tab-button': selectedTab === (dataType.id || dataType.name), 'fullwidth-tab-button': fullWidth }">
      <div class="tab-content">
        <img class="small-icons" :class="dataType.iconClass" :src="getAppIcon(dataType.icon)">
        <div v-if="!iconsOnly && (selectedTab === (dataType.id || dataType.name) || fullWidth)" class="selected-tab-button-text"> {{
          dataType.nameKey ? $t(dataType.nameKey) : utils.capitalizeStr(dataType.name) }}</div>
      </div>
    </div>
    
  </div>
</template>

<script setup>

import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

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

import { ref, computed, onMounted, watch, nextTick, onBeforeUnmount } from 'vue';
import { useTrayStates } from '@/stores/TrayStates';
import utils from '@/services/utils';
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

const filterList = (index, dataType) => {
  highlightFilter(index);
  emit('filter', dataType);
};

const highlightFilter = (index) => {
  filterIndex.value = index;
};

onMounted(() => {
});
</script>

<style scoped>
.header-tab-root {
  display: flex;
  flex-direction: row;
  box-sizing: border-box;
  width: 100%;
  justify-content: space-between;
  align-items: center;
  overflow: hidden;
  height: 34px;
  gap: .2rem;
  padding: .3rem 0;
  color: var(--text);
}

.fullwidth-header-tab-root {
  width: max-content;
}

.icon-header-tab-root {
  width: min-content;
}

.header-tab-container {
  align-items: center;
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  flex-wrap: nowrap;
  box-sizing: border-box;
  /* width: 100%; */
  justify-content: space-evenly;
  overflow: hidden;
  border-radius: 8px;
  color: var(--text);
  padding: 1rem .3rem;
  gap: .5rem;
  overflow: hidden;
  background-color: goldenrod;
}

.fullwidth-header-tab-container {
  width: max-content;
}

.selected-tab-button-text {
  padding: .2rem .1rem;
  font-weight: 250;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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
  justify-content: flex-start;
  padding: 5px .5rem;
  transition: background-color 0.2s ease-out, opacity 0.2s ease-out;
  border-radius: var(--large-radius);
  flex-shrink: 0;
}

.tab-button:hover {
  background-color: var(--surface-4);
  opacity: 1;
}

.tab-button:active {
  /* background-color: #00000013; */
  opacity: 1;
}

.tab-button-pressed {
  box-sizing: border-box;
  background-color: rgba(0, 0, 0, 0.216);
  outline: solid 1px var(--border-strong);
  outline-offset: -1px;
}

.selected-tab-button {
  outline-offset: -1px;
  width: 100%;
  background-color: var(--surface-3);
  opacity: 1;
  transition: background-color 0.2s ease-out, opacity 0.2s ease-out;
  min-width: 0;
  flex-shrink: 1;
  overflow: hidden;
}

.fullwidth-tab-button {
  width: max-content;
}

.selected-tab-button:hover {
  background-color: var(--surface-4);
  border-radius: var(--normal-radius);
}

.tab-content {
  display: flex;
  gap: .5rem;
  font-size: 14px;
  align-items: center;
  overflow: hidden;
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
  /* outline-offset: -1px; */
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


