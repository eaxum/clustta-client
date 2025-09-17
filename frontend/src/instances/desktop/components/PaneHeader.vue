<template>
  <div v-if="!trayStates.showTraySearch" class="page-header-area-container">
    <div class="page-header-area-container-title">
      <span><img class="large-icons header-icons" :src="icon"></span>
      <span class="page-header-title">{{ title }}</span>
    </div>
    <span v-if="showMeta" class="action-button" @click="toggleMeta" v-tooltip="'Toggle Metadata'"><img
        class="large-icons" src="/icons/info.svg"></span>
    <span v-if="showSearch" class="action-button" @click="toggleSearch" v-tooltip="'Toggle Search'"><img
        class="large-icons" src="/icons/search.svg"></span>
    <span v-if="showPin" class="action-button" :class="{ 'action-button-pressed': trayStates.keepModalOpen === true }"
      @click="trayStates.toggleKeepModal" v-tooltip="trayStates.keepModalOpen ? 'Unpin' : 'Pin'">
      <img v-if="!trayStates.keepModalOpen" class="large-icons" src="/icons/pin.svg">
      <img v-else class="large-icons" src="/icons/unpin.svg">
    </span>
  </div>
  <div v-if="trayStates.showTraySearch" class="page-header-area-container">
    <!-- <span><img class="large-icons" :src="icon"></span> -->
    <input v-model="query" class="input-short search-bar" type="text" :placeholder="placeholder" @input="updateSearch"
      v-focus />
    <span class="action-button" @click="toggleSearch" v-tooltip="'Toggle Search'"><img class="large-icons"
        src="/icons/search.svg"></span>
  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue';
import { useTrayStates } from '@/stores/TrayStates';

const emits = defineEmits(["update-search", "toggle-search"]);

const trayStates = useTrayStates();
const props = defineProps({

  title: String,
  icon: String,
  showSearch: Boolean,
  showMeta: Boolean,
  showPin: Boolean,
  searchQuery: {
    type: String,
    default: ''
  },
  placeholder: String,

});

const query = ref('');

watch(() => props.searchQuery, (value) => {
  query.value = value;
});

const updateSearch = () => {
  emits('update-search', query.value);
}

const toggleSearch = () => {
  trayStates.changeSearchVisibility();
  emits('toggle-search');
};

const toggleMeta = () => {
  trayStates.changeMetaVisibility();
  // emits('toggle-search');
}

onMounted(() => {
  trayStates.showTraySearch = false;

})


</script>

<style scoped>
@import "@/assets/tray.css";

.header-icons {
  padding: 2px;
  height: 30px;
  min-width: 30px;
  overflow: hidden;
  display: flex;
  align-items: center;
}

.page-header-area-container {
  display: flex;
  align-items: center;
  flex-direction: row;
  box-sizing: border-box;
  width: 98%;
  height: 60px;
  gap: .5rem;
  /* background-color: blue; */
  /* padding: 0 .5rem; */
  overflow: hidden;
  justify-content: space-between;
  /* background-color: transparent; */
}

.page-header-title {
  font-family: Inter, sans-serif;
  color: white;
  font-size: 20px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>

