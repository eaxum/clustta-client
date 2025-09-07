<template>
  <div class="header-tab-root"
    :class="{ 'fullwidth-header-tab-root': fullWidth, 'icon-header-tab-root': iconsOnly }">
    <div v-for="(dataType, index) in dataTypes" :key="dataType.name"
      v-tooltip="((filterIndex !== index || iconsOnly)) ? utils.capitalizeStr(dataType.name) : ''"
      @click="filterList(index, dataType.name)" class="tab-button"
      :class="{ 'selected-tab-button': selectedTab === dataType.name, 'fullwidth-tab-button': fullWidth }">
      <div class="tab-content">
        <img class="small-icons" :src="getAppIcon(dataType.icon)">
        <div v-if="!iconsOnly && (selectedTab === dataType.name || fullWidth)" class="selected-tab-button-text"> {{
          utils.capitalizeStr(dataType.name) }}</div>
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
  height: 100%;
  justify-content: space-between;
  align-items: center;
  overflow: hidden;
  height: max-content;
  gap: .5rem;
  padding: .3rem .3rem;
  color: var(--white);
  /* background-color: firebrick; */
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
  color: var(--white);
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
  ;
}

.tab-button {
  position: relative;
  border-radius: 8px;
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  align-items: center;
  height: max-content;
  /* transition: all .1s ease-in-out; */
  /* background-color: var(--black-steel); */
  opacity: .5;
  justify-content: flex-start;
  padding: 5px .5rem;
  /* gap: .5rem; */
}

.tab-button:hover {
  background-color: #ffffff15;
  /* border-bottom: solid 2px var(--white); */
  background-color: var(--light-steel);
  opacity: 1;
}

.tab-button:active {
  /* background-color: #00000013; */
  opacity: 1;
}

.tab-button-pressed {
  box-sizing: border-box;
  background-color: rgba(0, 0, 0, 0.216);
  outline: solid 1px var(--white);
  outline-offset: -1px;
}

.selected-tab-button {
  /* transition: all .2s ease-in-out; */
  /* background-color: rgba(0, 0, 0, 0.216); */

  /* border-bottom: solid 2px var(--white); */
  
  /* outline: var(--transparent-line); */
  outline-offset: -1px;
  width: 100%;
  /* background-color: var(--white); */
  background-color: var(--steel);
  /* color: black; */
  opacity: 1;
}

.fullwidth-tab-button {
  width: max-content;
}

.selected-tab-button:hover {
  background-color: var(--light-steel);

}

.tab-content {
  display: flex;
  gap: .5rem;
  /* background-color: rebeccapurple; */
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
  color: var(--white);
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


