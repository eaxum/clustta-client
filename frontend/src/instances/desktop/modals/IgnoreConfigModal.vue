<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>


    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="'folder-arrow-in'" :showSearch="false" />

    </div>

    <div class="general-container">
      <HeaderTabs :dataTypes="settingsItems" @filter="filterList" :fullWidth="true" />

      <div class="config-group">
        <span class="config-item" @click="toggleField('tempFiles')">
          <!-- <img class="small-icons" :src="getStatusIcon(status)"> -->
          <div class="horizontal-flex">
            <div> Temporary files (.tmp, .temp) </div>
            <ToggleSwitch :switchValueProp="ignoreParameters['tempFiles']" />
          </div>
        </span>

        <span class="config-item" @click="toggleField('logFiles')">
          <!-- <img class="small-icons" :src="getStatusIcon(status)"> -->
          <div class="horizontal-flex">
            <div> Log files (.log) </div>
            <ToggleSwitch :switchValueProp="ignoreParameters['logFiles']" />
          </div>
        </span>

        <span class="config-item" @click="toggleField('backupFiles')">
          <!-- <img class="small-icons" :src="getStatusIcon(status)"> -->
          <div class="horizontal-flex">
            <div> Backup files (.bak, .backup) </div>
            <ToggleSwitch :switchValueProp="ignoreParameters['backupFiles']" />
          </div>
        </span>

        <span class="config-item" @click="toggleField('officeDocs')">
          <!-- <img class="small-icons" :src="getStatusIcon(status)"> -->
          <div class="horizontal-flex">
            <div> Office documents (.docx, .xlsx, .pptx) </div>
            <ToggleSwitch :switchValueProp="ignoreParameters['officeDocs']" />
          </div>
        </span>

        <span class="config-item" @click="toggleField('imageFiles')">
          <!-- <img class="small-icons" :src="getStatusIcon(status)"> -->
          <div class="horizontal-flex">
            <div> Image files (.jpg, .png, .gif) </div>
            <ToggleSwitch :switchValueProp="ignoreParameters['imageFiles']" />
          </div>
        </span>
        <span class="config-item" @click="toggleField('videosFiles')">
          <!-- <img class="small-icons" :src="getStatusIcon(status)"> -->
          <div class="horizontal-flex">
            <div> Videos files (.mov, .mp4, .av1) </div>
            <ToggleSwitch :switchValueProp="ignoreParameters['videosFiles']" />
          </div>
        </span>
      </div>

      <SearchSuggestions placeholder="Add Extension" :tags="extensions" :projectTags="projectTags" :showTags="true"
        :forSearch="false" @tagAdded="addExtension" @tagRemoved="removeExtension" />


      <div class="pop-up-actions" ref="popUpActions">
        <GeneralButton :label="'Cancel'" :fullWidth="false" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Confirm'" :fullWidth="false" @click="importItems()" :loading="isAwaitingResponse" />
      </div>

    </div>

  </div>

</template>

<script setup>
// imports
import { ref, watchEffect } from 'vue';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import HeaderTabs from '@/instances/common/components/HeaderTabs.vue';
import SearchSuggestions from '@/instances/common/components/SearchSuggestions.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';

const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();

// refs
const extensions = ref([]);
const ignoreParameters = ref({
  tempFiles: true,
  logFiles: true,
  backupFiles: true,
  officeDocs: true,
  imageFiles: false,
  videosFiles: false,
});
const isAwaitingResponse = ref(false);
const modalContainer = ref(null);
const popUpActions = ref(null);
const selectedTabName = ref('File Types');

// constants
const projectTags = [];
const settingsItems = [
  { name: 'File Types', icon: 'file' },
  { name: 'Folders', icon: 'folder' },
  { name: 'Common Presets', icon: 'cog' },
];
const title = 'Ignore Configuration';

// methods
// Adds an extension to the ignore list.
const addExtension = (extension) => {
  if (!extensions.value.includes(extension)) {
    extensions.value.push(extension);
  }
};

// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility('ignoreConfigModal', false);
};

// Filters the list based on selected tab.
const filterList = (selectedTab) => {
  console.log(selectedTab);
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Removes an extension from the ignore list.
const removeExtension = (extension) => {
  extensions.value = extensions.value.filter(t => t !== extension);
};

// Toggles a field in the ignore parameters.
const toggleField = (key) => {
  ignoreParameters.value[key] = !ignoreParameters.value[key];
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.config-group {
  display: flex;
  flex-direction: column;
  color: white;
  align-items: center;
  gap: .3rem;
  padding: .6rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  border-radius: var(--normal-radius);
  background-color: var(--dark-steel);
}

.config-item {
  overflow: hidden;
  background-color: transparent;
  text-align: center;
  font-size: 14px;
  line-height: 14px;
  color: white;
  position: relative;
  border-radius: var(--small-radius);
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  gap: 10px;
  align-items: center;
  padding-left: .3rem;
  width: 100%;
  min-width: max-content;
  min-height: max-content;
  transition: all 0.3s ease;
}

.general-container-wide {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  width: 90vw;
  min-width: 600px !important;
  max-width: 1000px;
  max-height: 80vh;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.modal-container {
  justify-content: flex-start;
  align-items: flex-start;
  max-height: 90vh;
}

.pop-up-actions {
  align-items: center;
  box-sizing: border-box;
}

.pop-up-prompt {
  gap: 10px;
  align-items: center;
  max-height: 400px;
}

.rules-toggle {
  display: flex;
  gap: .5rem;
  align-items: center;
  min-width: max-content;
}

.selected-folder {
  width: 100%;
  overflow: hidden;
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  padding: 10px 20px;
  box-sizing: border-box;
}

.selected-folder-container {
  display: flex;
  width: 100%;
  gap: .2rem;
  overflow: hidden;
  align-items: center;
}
</style>



