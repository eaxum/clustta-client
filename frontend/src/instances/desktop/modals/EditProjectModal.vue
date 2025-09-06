<template>

  <div class="modal-container" v-esc="closeModal" v-return="handleEnterKey">
    <div class="general-pane-header">
      <HeaderArea v-if="isCustomIcon" :title="title" :customIcon="projectIcon" />
      <HeaderArea v-else :title="title" :emoji="projectIcon" />
      <ActionButton v-if="displayEmojiSelector"  :icon="getAppIcon('arrow-left')" :showLabel="false" v-tooltip="'Back to details'"
        :buttonFunction="toggleEmojiSelector" />
      <ActionButton v-else  :icon="getAppIcon('face-plus')" :showLabel="false" v-tooltip="'Set project Icon'"
        :buttonFunction="toggleEmojiSelector" />
      <ActionButton v-if="isPreviewChanged" :icon="getAppIcon('revert')" :showLabel="false"
        v-tooltip="'Revert Cover Image'" :buttonFunction="revertCoverImage" />
      <ActionButton v-if="projectPreview && !displayEmojiSelector" :icon="getAppIcon('trash')" :showLabel="false"
        v-tooltip="'Remove Cover Image'" :buttonFunction="removeCoverImage" />
      <ActionButton v-else-if="projectPreview && !displayEmojiSelector" :icon="getAppIcon('image-plus')" :showLabel="false" v-tooltip="'Add Cover Image'"
        :buttonFunction="addCoverImage" />
    </div>


    <div class="general-container">

      <span @click="addCoverImage" v-if="projectPreview && !displayEmojiSelector" class="screenshot-preview">
        <img class="screenshot-thumb" :src="projectPreview">
      </span>

      <div class="input-section">
        <input v-model="projectName" class="input-short" type="text" placeholder="Project Name" v-focus />
      </div>

      <div v-if="displayEmojiSelector" class="header-tab-container">
        <div class="tab-button" :class="{ 'selected-tab-button': iconType === 'emoji', 'fullwidth-tab-button': true }"
          @click="changeIconType('emoji')">
          Emoji
        </div>
        <div class="tab-button" :class="{ 'selected-tab-button': iconType === 'upload', 'fullwidth-tab-button': true }"
          @click="changeIconType('upload')">
          Upload
        </div>
      </div>

      <EmojiPicker v-if="displayEmojiSelector && iconType == 'emoji'" @select="handleEmojiSelect" />
      <div v-if="displayEmojiSelector && iconType == 'upload'">
        
      <ActionButton  :icon="getAppIcon('image-plus')" :label="'Upload an image'" :buttonFunction="selectIcon" />

      </div>
      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Update'" :fullWidth="true" @click="updateProject()" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
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
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { DialogService, ProjectService } from '@/../bindings/clustta/services/index';
import utils from '@/services/utils';

// store imports
import { useUserStore } from '@/stores/users';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import EmojiPicker from '@/instances/desktop/components/EmojiPicker.vue'

// states
const projectStore = useProjectStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();

// vars
let title = 'Project Details';

// refs
const projectName = ref('');
const oldProjectName = ref('');
const projectIcon = ref('');
const oldProjectIcon = ref('');
const projectsDirectory = ref('');
const projectsDirectoryInput = ref(null);
const isAwaitingResponse = ref(false);

const fileIsSelected = ref(false);
const projectPreview = ref('');
const oldProjectPreview = ref('');
const coverImageName = ref('');
const coverImageFullName = ref('');
const coverImagePath = ref("");

const displayEmojiSelector = ref(false);

const toggleEmojiSelector = () => {
  displayEmojiSelector.value = !displayEmojiSelector.value
};

const selectedEmoji = ref('');

const handleEmojiSelect = (emojiData) => {
  selectedEmoji.value = emojiData;
  projectIcon.value = selectedEmoji.value.entity;
  displayEmojiSelector.value = false;
};

const iconType = ref("emoji")

const changeIconType = (type) => {
  iconType.value = type
};

// computed properties
const isNameChanged = computed(() => {
  const restrictedEntries = [oldProjectName.value, '']
  return !restrictedEntries.includes(projectName.value);
});
const isPreviewChanged = computed(() => {
  return oldProjectPreview.value !== projectPreview.value;
});
const isProjectIconChanged = computed(() => {
  return oldProjectIcon.value !== projectIcon.value;
});

const isValueChanged = computed(() => {
  const restrictedEntries = [oldProjectName.value, '']
  return !restrictedEntries.includes(projectName.value) || isPreviewChanged.value || isProjectIconChanged.value
});

const selectDirectoryPath = async () => {

  // if possible, this should open to the current project working directory

  // const result = await open({
  //   multiple: false,
  //   directory: true
  // });

  if (result) {
    let fileDir = result.replace(/\\/g, '/');
    projectsDirectory.value = fileDir;
    projectsDirectoryInput.value.focus();
  }
};

const setCoverImage = () => {
  coverImagePath.value = '';
  coverImageFullName.value = '';
  projectPreview.value = null;
};

const addCoverImage = async () => {
  const result = await DialogService.SelectFileDialog("Select Image File", "*.png;*.jpg;*.jpeg;*.gif;*.bmp;*.webp");
  if (result) {
    let filePath = result.replace(/\\/g, '/');
    let fileName = filePath.split('/').pop();

    projectPreview.value = await utils.base64FromFile(filePath);

    coverImagePath.value = filePath;
    coverImageFullName.value = fileName
    if (!coverImageName.value) {
      coverImageName.value = fileName.split('.').slice(0, -1).join('.');
    }
    fileIsSelected.value = true;
  }
};

const removeCoverImage = () => {
  projectPreview.value = '';
};

const revertCoverImage = () => {
  projectPreview.value = oldProjectPreview.value;
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    updateProject();
  }
};

const closeModal = (all) => {
  modals.disableAllModals()
};


const updateProject = async () => {

  isAwaitingResponse.value = true;

  if (isPreviewChanged.value) {
    console.log('image changed');
    await updateProjectCover();
  }

  if (isProjectIconChanged.value) {
    console.log('icon changed');
    await updateProjectIcon();
  }

  if (isNameChanged.value) {
    console.log('meta changed');
    await updateProjectMeta();
  }

  isAwaitingResponse.value = false;
  closeModal();

}

// TODO rename project
const updateProjectMeta = async () => {

  if (projectStore.activeProject.has_remote) {
    ProjectService.Rename(projectStore.getActiveProjectUrl, projectStore.selectedStudio.name, projectName.value)
      .then((data) => {
        projectStore.activeProject.name = projectName.value
      }).catch(error => {
        console.log(error)
      })
  } else {
    ProjectService.Rename(projectStore.activeProject.uri, projectStore.selectedStudio.name, projectName.value)
      .then((data) => {
        projectStore.activeProject.name = projectName.value
      }).catch(error => {
        console.log(error)
      })
  }

}

const updateProjectCover = async () => {
  await ProjectService.UpdatePreview(projectStore.activeProject.uri, coverImagePath.value).then(() => {
    projectStore.refreshProjectPreview(projectStore.activeProject.id);
  }).catch((error) => {
    console.error(error)
    notificationStore.addNotification(
      "Error Updating Image",
      error,
      "error",
      false
    )
  })
};

const isCustomIcon = computed(() => projectIcon.value.length > 10);

const selectIcon = async () => {
  // TODO this is broken
  const result = await DialogService.SelectIconDialog();
  if (result) {
    const image = "data:image/png;base64," + result
    projectIcon.value = image
  }
};

const updateProjectIcon = async () => {

  if (projectStore.activeProject.has_remote) {
    await ProjectService.UpdateIcon(projectStore.getActiveProjectUrl, projectStore.selectedStudio.name, projectIcon.value).then(() => {
      projectStore.activeProject.icon = projectIcon.value
      const index = projectStore.projects.findIndex(project => project.id === projectStore.activeProject.id);
      projectStore.projects[index].icon = projectIcon.value
    }).catch((error) => {
      console.error(error)
      notificationStore.addNotification(
        "Error Updating Icon",
        error,
        "error",
        false
      )
    })
  } else {
    await ProjectService.UpdateIcon(projectStore.activeProject.uri, projectStore.selectedStudio.name, projectIcon.value).then(() => {
      projectStore.activeProject.icon = projectIcon.value
      const index = projectStore.projects.findIndex(project => project.id === projectStore.activeProject.id);
      projectStore.projects[index].icon = projectIcon.value
    }).catch((error) => {
      console.error(error)
      notificationStore.addNotification(
        "Error Updating Icon",
        error,
        "error",
        false
      )
    })
  }
}



// onMounted
onMounted(() => {
  let project = projectStore.activeProject;
  projectsDirectory.value = project.working_directory;

  projectName.value = project.name;
  oldProjectName.value = project.name;

  projectIcon.value = project.icon;
  oldProjectIcon.value = project.icon;

  projectPreview.value = project.preview;
  oldProjectPreview.value = project.preview;
});

onUnmounted(() => {
  projectPreview.value = null;
});


</script>


<style scoped>
@import "@/assets/desktop.css";

.general-container {
  gap: 1rem;
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
  padding: .3rem;
  gap: .5rem;
  overflow: hidden;
  /* background-color: red; */
}

.selected-tab-button-text {
  padding: .2rem .1rem;
  font-weight: 250;
  ;
}

.tab-button {
  position: relative;
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  align-items: center;
  height: max-content;
  opacity: .5;
  justify-content: center;
  padding: 5px .5rem;
}

.tab-button:hover {
  background-color: #ffffff15;
  background-color: var(--light-steel);
  opacity: 1;
}

.tab-button:active {
  opacity: 1;
}

.tab-button-pressed {
  box-sizing: border-box;
  background-color: rgba(0, 0, 0, 0.216);
  outline: solid 1px var(--white);
  outline-offset: -1px;
}

.selected-tab-button {
  border-bottom: solid 2px var(--white);
  width: 100%;
  opacity: 1;
}

.fullwidth-tab-button {
  width: max-content;
}

.selected-tab-button:hover {
  background-color: var(--black-steel);

}

.tab-content {
  display: flex;
  gap: .5rem;
}

.upload-image {
  background-color: rgb(82, 81, 81);
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  padding: 14px 90px;
  border-radius: 6px;
  color: var(--white);
  flex-direction: row;
  font-weight: 500;
  gap: 10px;
}

.upload-image:hover {
  transform: scale(1.02);
}
</style>