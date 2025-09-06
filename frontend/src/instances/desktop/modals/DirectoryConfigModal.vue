<template>
  <div class="modal-container">
    <HeaderArea :title="title" :icon="'explorer'" :showSearch="showSearch" />
    <div class="general-container">

      

      <div class="input-section">
        <span class="regular">Projects Folder</span>
        <div class="horizontal-flex">
          <input v-model="workingDirectory" class="input-short" type="text"
            placeholder="Working Directory" ref="sharedProjectsDirectoryInput" />
          <span @click="pasteDirectoryPath('workingDir')" class="single-action-button" v-tooltip="'Paste link'"><img
              class="small-icons" :src="getAppIcon('clipboard')"></span>
          <span @click="selectDirectoryPath('workingDir')" class="single-action-button" v-tooltip="'Browse Path'"><img
              class="small-icons" :src="getAppIcon('explorer')"></span>
        </div>
      </div>

      <div class="menu-divider" ></div>

      <div class="input-section">
        <span class="regular">Clustta local projects data</span>
        <div class="horizontal-flex">
          <input v-model="projectsDirectory" class="input-short" type="text" placeholder="Projects Directory"
            ref="projectsDirectoryInput" />
          <span @click="pasteDirectoryPath('personal')" class="single-action-button" v-tooltip="'Paste link'"><img
              class="small-icons" :src="getAppIcon('clipboard')"></span>
          <span @click="selectDirectoryPath('personal')" class="single-action-button" v-tooltip="'Browse Path'"><img
              class="small-icons" :src="getAppIcon('explorer')"></span>
        </div>
      </div>

      <div class="input-section">
        <span class="regular">Clustta shared projects data</span>
        <div class="horizontal-flex">
          <input v-model="sharedProjectsDirectory" class="input-short" type="text"
            placeholder="Shared Projects Directory" ref="sharedProjectsDirectoryInput"  />
          <span @click="pasteDirectoryPath('shared')" class="single-action-button" v-tooltip="'Paste link'"><img
              class="small-icons" :src="getAppIcon('clipboard')"></span>
          <span @click="selectDirectoryPath('shared')" class="single-action-button" v-tooltip="'Browse Path'"><img
              class="small-icons" :src="getAppIcon('explorer')"></span>
        </div>
      </div>

      <div class="pop-up-actions">
        <button class="button default" @click="closeModal()" v-stop-propagation>Cancel</button>
        <button class="button colored" @click="saveChanges()" v-stop-propagation>Save</button>
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
import { ref, onMounted } from 'vue';

//stores
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';

//components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

import { ClipboardService, SettingsService, DialogService } from '@/../bindings/clustta/services/index';

//header vars
let title = 'Configure Directories';
let showSearch = false;

//refs
const notificationStore = useNotificationStore();
const sharedProjectsDirectory = ref('');
const sharedProjectsDirectoryInput = ref(null);
const projectsDirectory = ref('');
const projectsDirectoryInput = ref(null);
const workingDirectory = ref('');
const workingDirectoryInput = ref(null);
const modals = useDesktopModalStore();

//methods

const pasteDirectoryPath = async (context) => {
  ClipboardService.ReadText()
    .then(path => {
      if (context === 'shared') {
        sharedProjectsDirectory.value = path;
      } else if (context === 'personal') {
        projectsDirectory.value = path;
      }else if (context === 'workingDir') {
        workingDirectory.value = path;
      }
    }).catch(err => {
      console.error("Failed to paste from clipboard:", err);
    });
};

const selectDirectoryPath = async (context) => {
  const result = await DialogService.SelectFolderDialog("Select Folder File");
  if (result) {
    let fileDir = result.replace(/\\/g, '/');

    if (context === 'shared') {
      sharedProjectsDirectory.value = fileDir;
      sharedProjectsDirectoryInput.value.focus();
    } else if (context === 'personal') {
      projectsDirectory.value = fileDir;
      projectsDirectoryInput.value.focus();
    } else if (context === 'workingDir') {
      workingDirectory.value = fileDir;
      workingDirectoryInput.value.focus();
    }
  }
};

const saveChanges = async () => {
  await SettingsService.SetProjectDirectory(projectsDirectory.value)
  await SettingsService.SetWorkingDirectory(workingDirectory.value)
  await SettingsService.SetSharedProjectDirectory(sharedProjectsDirectory.value)
  closeModal();
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
  await SettingsService.GetProjectDirectory()
    .then((response) => {
      projectsDirectory.value = response
    })
    .catch((error) => {
      notificationStore.addNotification(
        "Error Loading Settings",
        error.message,
        "error",
        false
      )
    });
  await SettingsService.GetSharedProjectDirectory()
    .then((response) => {
      sharedProjectsDirectory.value = response
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
      workingDirectory.value = response
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

.regular{
  padding-left: .5rem;
  color: var(--white);
  font-size: 14px;
}

.input-section {
  width: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  justify-content: flex-start;
  gap: .4px;
  color: white;
}

.general-container {
  gap: 1rem;
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
