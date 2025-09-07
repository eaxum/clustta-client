<template>
  <div v-if="!displaySearchBar" class="page-header-area-container">
    <div class="page-header-area-container-title">

      <span v-if="emoji" class="project-icon" v-html="emoji"></span>

      <span v-else-if="icon"><img class="large-icons header-icons" :class="{ 'no-filter' : useIconBlob}" :src="getAppIcon(icon)"></span>
      <span v-else-if="customIcon"><img class="large-icons header-icons no-filter" :src="customIcon"></span>

      <div v-if="profileImage" class="large-icons header-icons profile-picture"
        :style="{ backgroundColor: backgroundColor }">
        <img class="profile-img" :src="profileImage">
      </div>

      <span class="page-header-title" :class="{ 'mini-display': miniDisplay }">{{ title }}</span>
    </div>
    <span v-if="showMeta" class="single-action-button" @click="toggleMeta" v-tooltip="'Toggle Metadata'"><img
        class="large-icons" src="/icons/info.svg"></span>
    <span v-if="showSearch" class="single-action-button" @click="toggleSearch" v-tooltip="'Toggle Search'"><img
        class="large-icons" src="/icons/search.svg"></span>
    <span v-if="showPin" class="single-action-button"
      :class="{ 'single-action-button-pressed': trayStates.keepModalOpen === true }" @click="trayStates.toggleKeepModal"
      v-tooltip="trayStates.keepModalOpen ? 'Unpin' : 'Pin'">
      <img v-if="!trayStates.keepModalOpen" class="large-icons" src="/icons/pin.svg">
      <img v-else class="large-icons" src="/icons/unpin.svg">
    </span>
  </div>
  <div v-if="displaySearchBar" class="page-header-area-container">
    <!-- <span><img class="large-icons" :src="icon"></span> -->
    <input v-model="query" class="input-short search-bar" type="text" :placeholder="placeholder" @input="updateSearch"
      v-focus />
    <span class="single-action-button" @click="toggleSearch" v-tooltip="'Toggle Search'"><img class="large-icons"
        src="/icons/search.svg"></span>
  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue';
import { useTrayStates } from '@/stores/TrayStates';
import { useIconStore } from '@/stores/icons';

const emits = defineEmits(["update-search", "toggle-search", "clear-search"]);

const trayStates = useTrayStates();
const iconStore = useIconStore();
const props = defineProps({

  title: String,
  icon: String,
  emoji: String,
  customIcon: String,
  backgroundColor: String,
  profileImage: String,
  showSearch: Boolean,
  showMeta: Boolean,
  showPin: Boolean,
  miniDisplay: { type: Boolean, default: false },
  useIconBlob: { type: Boolean, default: false },
  searchQuery: {
    type: String,
    default: ''
  },
  placeholder: String,

});



const query = ref('');
const displaySearchBar = ref(false);

// methods
const getAppIcon = (iconName) => {

  if (props.useIconBlob) {
    return iconName
  }

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

watch(() => props.searchQuery, (value) => {
  query.value = value;
});

const updateSearch = () => {
  emits('update-search', query.value);
}

const toggleSearch = () => {
  displaySearchBar.value = !displaySearchBar.value;
  if(!displaySearchBar.value){
    emits('clear-search');
    query.value = ''
  }
  emits('toggle-search');
};

const toggleMeta = () => {
  trayStates.changeMetaVisibility();
}

onMounted(() => {
  displaySearchBar.value = false;

})


</script>

<style scoped>

.header-icons {
  padding: 2px;
  height: 30px;
  min-width: 30px;
  overflow: hidden;
  display: flex;
  align-items: center;
}

.project-icon {
  font-size: x-large;

  font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI Emoji",
    "Segoe UI", "Apple Color Emoji", sans-serif;
  font-size: 2rem;
}

.page-header-area-container {
  display: flex;
  align-items: center;
  flex-direction: row;
  box-sizing: border-box;
  width: 98%;
  height: 60px;
  gap: .5rem;
  overflow: hidden;
  justify-content: space-between;
}

.page-header-title {
  font-family: Inter, sans-serif;
  color: var(--white);
  font-size: 20px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.mini-display {
  font-size: 16px;
}
.input-short{
  width: 100%;
}
</style>

