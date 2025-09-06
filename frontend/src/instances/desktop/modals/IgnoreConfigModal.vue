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
import { useIconStore } from '@/stores/icons';
import { v4 as uuidv4 } from 'uuid'
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};
// imports
import { onMounted, watchEffect, onUnmounted, ref, computed } from 'vue';

// state imports
import { useTrayStates } from '@/stores/TrayStates';

// store imports
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useStageStore } from '@/stores/stages';
import { useTaskStore } from '@/stores/task';
import { useEntityStore } from '@/stores/entity';
import { useStatusStore } from '@/stores/status';
import { useProjectStore } from '@/stores/projects';
import { useImportStore } from '@/stores/import';
import { useDndStore } from '@/stores/dnd';
import { useMenu } from '@/stores/menu';

// components
import HeaderTabs from '@/instances/common/components/HeaderTabs.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import SearchSuggestions from '@/instances/common/components/SearchSuggestions.vue';

// states
const trayStates = useTrayStates();
const stage = useStageStore();
const menu = useMenu();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const statusStore = useStatusStore();
const projectStore = useProjectStore();
const entityStore = useEntityStore();
const taskStore = useTaskStore();
const importStore = useImportStore();
const dndStore = useDndStore();

let title = 'Igore Configuration';

// refs
const modalContainer = ref(null);
const popUpActions = ref(null);
const isAwaitingResponse = ref(false);
const selectedTabName = ref("File Types");

const ignoreParameters = ref({
  tempFiles: true,
  logFiles: true,
  backupFiles: true,
  officeDocs: true,
  imageFiles: false,
  videosFiles: false,
});

const extensions = ref([]);

const settingsItems = [
  { name: "File Types", icon: "file" },
  { name: "Folders", icon: "folder" },
  { name: "Common Presets", icon: "cog" },
]

const toggleField = (key) => {
  ignoreParameters.value[key] = !ignoreParameters.value[key]
}

const filterList = (selectedTab) => {
  console.log(selectedTab);
  // let modalName;
  // if(selectedTab === 'Folder types'){
  // 	modalName = 'EntityTypes'
  // } else if(selectedTab === 'Task types'){
  // 	modalName = 'TaskTypes'
  // } else {
  // 	modalName = selectedTab;
  // }
  // const selectedTabName = modalName.toLowerCase();
  // settings.setModalVisibility(selectedTabName, true);
};

const removeExtension = (extension) => {
  extensions.value = extensions.value.filter(t => t !== extension);
};

const addExtension = (extension) => {
  if (extensions.value.includes(extension)) {
    return
  }
  else {
    extensions.value.push(extension);
  }
  console.log(extensions.value)
};

const closeModal = () => {
  modals.setModalVisibility("ignoreConfigModal", false);
};



// onMounted(async () => {
// });

// onUnmounted(() => {
// })


</script>

<style scoped>
@import "@/assets/desktop.css";

.pop-up-actions {
  /* width: min-content; */
  align-items: center;
  /* background-color: forestgreen; */
  /* justify-content: space-around; */
  box-sizing: border-box;
}

.modal-container {
  justify-content: flex-start;
  align-items: flex-start;
  max-height: 90vh;
}

.general-container-wide {
  display: flex;
  flex-direction: column;
  /* background-color: firebrick; */
  overflow: hidden;
  width: 90vw;
  min-width: 600px !important;
  max-width: 1000px;
  max-height: 80vh;

  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.folder-path {
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
}

.rules-toggle {
  display: flex;
  /* background-color: red; */
  gap: .5rem;
  align-items: center;
  min-width: max-content
}

.selected-folder {
  /* background-color: darkblue; */
  width: 100%;
  padding: .2rem;
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
  /* background-color: firebrick; */
  width: 100%;
  gap: .2rem;
  overflow: hidden;
  display: flex;
  align-items: center;
}

.compound-input-section {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .4rem;
}

.task-options-container {
  position: relative;
  box-sizing: border-box;

  width: 100%;
  height: 0px;
  /* height: 80px; */
  overflow: hidden;
  transition-property: height;
  transition-duration: 0.2s;
  transition-timing-function: ease-in-out;
  transition: opacity .5s ease-in-out;
  /* transition: all .1s ease-in-out; */
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0;
  opacity: 1;
}

.task-options-container-closed {
  transition: all .2s ease-in-out;
  opacity: 0;
  height: 0px;
  padding: 0;
  overflow: hidden;
  /* margin-bottom: -1.5rem; */
}


.input-short {
  flex: 1;
  width: 100%;
}

.listbox-short {

  flex: 1;
  width: 130px;
}

.input-label {

  font-family: Inter, sans-serif;
  color: white;
  font-size: 16px;
  white-space: nowrap;
  flex: 1;

}

.pop-up-prompt {
  gap: 10px;
  /* background-color: bisque; */
  align-items: center;
  /* justify-content: center; */
  max-height: 400px;
}

.config-group {
  display: flex;
  flex-direction: column;
  /* flex-wrap: wrap; */
  color: white;
  align-items: center;
  gap: .3rem;
  padding: .6rem;
  box-sizing: border-box;
  width: max-content;
  width: 100%;
  height: min-content;
  border-radius: var(--normal-radius);
  background-color: var(--dark-steel);
  /* background-color: crimson; */
  /* overflow: hidden; */
}

.config-item {
  overflow: hidden;
  background-color: transparent;
  text-align: center;
  font-size: 14px;
  line-height: 14px;
  background-color: transparent;
  color: white;
  position: relative;
  border-radius: 8px;
  border-radius: var(--small-radius);
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  gap: 10px;
  align-items: center;
  padding-left: .3rem;
  height: max-content;
  width: max-content;
  min-width: max-content;
  min-height: max-content;
  width: 100%;
  /* aspect-ratio: 1/1; */
  transition: all 0.3s ease;
  /* background-color: rebeccapurple; */
}
</style>