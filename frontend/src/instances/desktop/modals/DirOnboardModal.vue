<template>
  <div class="modal-container">
    <HeaderArea :title="'Select Clustta directory'" :icon="'explorer'" :showSearch="showSearch" />
    <div class="general-container">

        <div class="onboard-item-container">

            <img class="onboard-image" src="/page-states/project_cropped.png">

            <div class="pop-up-text-container">
                <div class="pop-up-body">
                Select the folder where Clustta will create and save your projects. You can change this later in the app settings.
                </div>
            </div>

            <div v-if="selectedClusttaDirectory" class="input-section">
                <div class="horizontal-flex">
                <input v-model="selectedClusttaDirectory" class="input-short read-only-input" type="text"
                    placeholder="Projects Directory" ref="projectsDirectoryInput" />
                <span @click="selectDirectory()" class="single-action-button" v-tooltip="'Browse Path'"><img
                    class="small-icons" :src="getAppIcon('explorer')"></span>
                </div>
            </div>

        </div>

      <div class="pop-up-actions" :class="{ 'spaced-out' : firstStage}" >
        <button v-if="!selectedClusttaDirectory" class="button colored" @click="selectDirectory()" v-stop-propagation>Select 'Clustta' Folder</button>
        <button v-else class="button colored" @click="saveChanges()" v-stop-propagation>Save</button>
      </div>

    </div>
  </div>
</template>

<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};


// imports
import { ref, onMounted, computed } from 'vue';

//stores
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';

//components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

import { v4 as uuidv4 } from "uuid";
import { ClipboardService, SettingsService, DialogService, FSService } from '@/../bindings/clustta/services/index';


//refs
const trayStates = useTrayStates();
const projectStore = useProjectStore();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();

const personalDataDirectory = ref('');
const sharedDataDirectory = ref('');

const projectsDirectory = ref('');
const projectsDirectoryInput = ref(null);

const defaultClusttaDirectory = ref('');
const selectedClusttaDirectory = ref('');

const userName = ref('');

const firstStage = ref(true);

//header vars
let showSearch = false;


//methods

const pasteDirectoryPath = async (context) => {
  ClipboardService.ReadText()
    .then(path => {
      if (context === 'dataDir') {
        selectedClusttaDirectory.value = path;
      }else if (context === 'projectsDir') {
        projectsDirectory.value = path;
      }
    }).catch(err => {
      console.error("Failed to paste from clipboard:", err);
    });
};



const selectDirectory = async () => {

  let title = 'Clustta Directory';
  let directory;
  let pathExists = false;

  try {
      pathExists = await FSService.DirExists(defaultClusttaDirectory.value);
  } catch (error) {
      pathExists = false;
  }
  
  directory = pathExists && !selectedClusttaDirectory.value ? defaultClusttaDirectory.value : selectedClusttaDirectory.value;

  const result = await DialogService.SelectSpecificFolderDialog(title, directory);

  if (result) {

      let fileDir = result.replace(/\\/g, '/');
      selectedClusttaDirectory.value = fileDir.replace(/\/clustta/g, '') + '/clustta';

      projectsDirectory.value = selectedClusttaDirectory.value + '/mnt';
      personalDataDirectory.value = selectedClusttaDirectory.value + '/projects';
      sharedDataDirectory.value = selectedClusttaDirectory.value + '/shared_projects';

  };

};

const saveChanges = async () => {
    await SettingsService.SetWorkingDirectory(projectsDirectory.value)
    await SettingsService.SetProjectDirectory(personalDataDirectory.value)
    await SettingsService.SetSharedProjectDirectory(sharedDataDirectory.value)
    await loadProjects();
    closeModal();
};

const loadProjects = async () => {
      await projectStore.loadProjects();
      trayStates.refreshData();
};

const closeModal = () => {
  modals.disableAllModals();
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    saveChanges();
  }
};



onMounted(async () => {
  await SettingsService.GetUserDirectory()
    .then((response) => {
      defaultClusttaDirectory.value = `${response}clustta`;
      console.log(defaultClusttaDirectory.value)
    })
    .catch((error) => {
      notificationStore.addNotification(
        "Error Loading Settings",
        error.message,
        "error",
        false
      )
    });
  await SettingsService.GetUsername()
    .then((response) => {
      userName.value = response
    })
    .catch((error) => {
      notificationStore.addNotification(
        "Error Loading Settings",
        error.message,
        "error",
        false
      )
    });

});

</script>

<style scoped>
@import "@/assets/desktop.css";

.pop-up-text-container {
  /* background-color: crimson; */
  display: flex;
  flex-direction: column;
  gap: .5rem;
  padding: .5rem;
}

.pop-up-info {
  padding: 0rem 1rem;
  /* margin-bottom: .75rem; */
}

.pop-up-body {
  font-size: 14px;
  color: var(--white);
}

.tip{
    font-style: italic;
    font-weight: 600;
}

.regular{
  color: var(--white);
}
.spaced-out{
    justify-content: flex-end;
}

.input-section {
  width: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  justify-content: flex-start;
  gap: .4px;
  color: var(--white);
}

.general-container {
  /* gap: 1rem; */
}

.onboard-image{
    width: 100%;
    width: 300px;
    padding: .5rem;
    /* aspect-ratio: 1/1; */
    /* background-color: red; */
}

.onboard-item-container{
    width: 100%;
    /* background-color: hotpink; */
    display: flex;
    box-sizing: border-box;
    height: 100%;
    align-items: center;
    flex-direction: column;
    gap: .5rem;
}
.modal-info {
  display: flex;
  flex-direction: column;
  max-width: 100%;
  justify-content: flex-start;
  align-self: stretch;
  width: 464px;
  align-items: flex-start;
  box-sizing: border-box;

}

.modal-text-container {
  display: flex;
  flex-direction: column;
  max-width: 100%;
  justify-content: flex-start;
  align-self: stretch;
  width: 464px;
  align-items: flex-start;
  /* margin-top: 20px; */
}

.modal-title {
  max-width: 100%;
  align-self: stretch;
  width: 464px;
  color: rgba(16, 24, 40, 1);
  color: white;
  font-size: 18px;
  line-height: 28px;
  letter-spacing: 0%;
  text-align: left;
}

.input-header {
  /* background-color: lightblue; */
  width: 100%;
  display: flex;
  align-items: center;
  margin: 10px 0px;
}

.input-count {
  background-color: none;
  font-size: 14px;
  color: white;
}

.modal-subtitle {
  /* background-color: beige; */
  /* max-width: 100%; */
  align-self: stretch;
  width: 464px;
  color: rgba(16, 24, 40, 1);
  color: white;
  font-size: 14px;
  /* line-height: 28px; */
  letter-spacing: 0%;
  text-align: left;
}



.modal-body {
  box-sizing: border-box;
  max-width: 100%;
  align-self: stretch;
  width: 464px;
  margin: 8px 0px;
  font-size: 14px;
  color: rgba(16, 24, 40, 1);
  line-height: 20px;
  letter-spacing: 0%;
  text-align: left;
}

.modal-actions {
  box-sizing: border-box;
  padding: 1rem 2rem;
  gap: 2rem;
  display: flex;
  flex-direction: row;
  max-width: 100%;
  align-self: stretch;
  align-items: center;
  justify-content: space-evenly;
  width: 464px;
  margin-top: 32px;
}

.div-10 {
  display: flex;
}

.task-options-container {
  position: relative;
  box-sizing: border-box;
  width: 100%;
  height: max-content;
  height: 40px;
  transition: all .2s ease-in-out;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0;
}

.task-options-container-closed {
  height: 0px;
  padding: 0;
  margin-bottom: -1rem;
}

.input-short {
  flex: 1;
  width: 100%;
  font-size: 14px;
}

.read-only-input{
    pointer-events: none;
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
  align-items: center;
  justify-content: center;
}
</style>


