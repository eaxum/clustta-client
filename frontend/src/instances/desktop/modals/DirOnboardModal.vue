<template>
  <div class="modal-container">
    <HeaderArea :title="title" :icon="'explorer'" :showSearch="showSearch" />
    <!-- {{  userDirectory }}
    {{  userName }} -->
    <div class="general-container">

        <div v-if="firstStage" class="onboard-item-container">

            <img class="onboard-image" src="/page-states/project_cropped.png">

            <div class="pop-up-text-container">
                <div class="pop-up-body">
                Select the folder where you would like to create your projects. If you already have a location, you can select it and Clustta will display your older projects as well.
                </div>
                <!-- <div class="pop-up-body tip">
                e.g Users/username/Documents/...
                </div> -->
            </div>

            
            <div v-if="projectsDirectory" class="input-section">
                <div class="horizontal-flex">
                <input v-model="projectsDirectory" class="input-short read-only-input" type="text"
                    placeholder="Projects Directory" ref="projectsDirectoryInput" />
                <!-- <span @click="pasteDirectoryPath('projectsDir', 'Projects Directory')" class="single-action-button" v-tooltip="'Paste link'"><img
                    class="small-icons" :src="getAppIcon('clipboard')"></span> -->
                <span @click="selectDirectory('projectsDir')" class="single-action-button" v-tooltip="'Browse Path'"><img
                    class="small-icons" :src="getAppIcon('explorer')"></span>
                </div>
            </div>

        </div>

         <div v-else class="onboard-item-container">

            <img class="onboard-image" src="/page-states/data.png">

            <div class="pop-up-text-container">
                <div class="pop-up-body">
                Now let's give Clustta permission to save data and backups of your projects. In the dialogue that appears, click Select Folder.
                </div>
                <!-- <div class="pop-up-body tip">
                Select Users/username/
                </div> -->
            </div>

            <div v-if="rootDataDirectory" class="input-section">
                <div class="horizontal-flex">
                <input v-model="fullDataDirectory" class="input-short read-only-input" type="text" placeholder="Clustta Data Directory"
                    ref="rootDataDirectoryInput" />
                <!-- <span @click="pasteDirectoryPath('dataDir')" class="single-action-button" v-tooltip="'Paste link'"><img
                    class="small-icons" :src="getAppIcon('clipboard')"></span> -->
                <span @click="selectDirectory('dataDir', 'Clustta Data Directory')" class="single-action-button" v-tooltip="'Browse Path'"><img
                    class="small-icons" :src="getAppIcon('explorer')"></span>
                </div>
            </div>

        </div>

      <div class="pop-up-actions" :class="{ 'spaced-out' : firstStage}" >
        <button v-if="!firstStage" class="button default" @click="toggleStage()" v-stop-propagation>Back</button>
        <button v-if="!projectsDirectory" class="button colored" @click="selectDirectory('projectsDir', 'Projects Directory')" v-stop-propagation>Select Projects Folder</button>
        <button v-else-if="!firstStage && !rootDataDirectory" class="button colored" @click="selectDirectory('dataDir', 'Clustta Data Directory')" v-stop-propagation>Select Data Folder</button>
        <button v-else class="button colored" @click="proceed()" v-stop-propagation>{{ firstStage ? 'Next' : 'Save' }}</button>
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
import { ClipboardService, SettingsService, DialogService } from '@/../bindings/clustta/services/index';


//refs
const trayStates = useTrayStates();
const projectStore = useProjectStore();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();

const personalDataDirectory = ref('');
const sharedDataDirectory = ref('');

const projectsDirectory = ref('');
const projectsDirectoryInput = ref(null);

const rootDataDirectory = ref('');
const rootDataDirectoryInput = ref(null);

const userDirectory = ref('');
const userName = ref('');

const firstStage = ref(true);

//header vars
let showSearch = false;


//methods
const title = computed(() => {
    const path = 'Configure Directories';
    const step = firstStage.value ? 1 : 2;
    return 'Configure Directories' + ` ${step} / 2`
});

const fullDataDirectory = computed(() => {
    const path = rootDataDirectory.value;
    return rootDataDirectory.value ? path.replace(/\/clustta/g, '') + '/clustta' : ''
});

const proceed = () => {
    if(firstStage.value){
        toggleStage()
    } else {
        saveChanges()
    }
};

const toggleStage = () => {
    console.log('toggled')
    firstStage.value = !firstStage.value 
};

const pasteDirectoryPath = async (context) => {
  ClipboardService.ReadText()
    .then(path => {
      if (context === 'dataDir') {
        rootDataDirectory.value = path;
      }else if (context === 'projectsDir') {
        projectsDirectory.value = path;
      }
    }).catch(err => {
      console.error("Failed to paste from clipboard:", err);
    });
};

const selectDirectory = async (context, title) => {

    let directory;

    if(context === 'projectsDir'){
        directory = projectsDirectory.value ? projectsDirectory.value : `${userDirectory.value}/Documents`;
    }else if(context === 'dataDir'){
        directory = rootDataDirectory.value ? rootDataDirectory.value : userDirectory.value;
    }

    const result = await DialogService.SelectSpecificFolderDialog(title, directory);
    if (result) {

        let fileDir = result.replace(/\\/g, '/');

        if (context === 'projectsDir') {
            projectsDirectory.value = fileDir;
        } else if (context === 'dataDir') {
            rootDataDirectory.value = fileDir;
            personalDataDirectory.value = fullDataDirectory.value + '/projects';
            sharedDataDirectory.value = fullDataDirectory.value + '/shared_projects';
        } 
    }
};

const saveChanges = async () => {
    console.log(projectsDirectory.value, personalDataDirectory.value, sharedDataDirectory.value)
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
      userDirectory.value = response
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

  await SettingsService.GetWorkingDirectory()
    .then((response) => {
    //   projectsDirectory.value = response
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


