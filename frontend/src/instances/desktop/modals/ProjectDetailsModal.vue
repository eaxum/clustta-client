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
      <ActionButton v-if="projectPreview && !displayEmojiSelector" :icon="getAppIcon('image-cancel')" :showLabel="false"
        v-tooltip="'Remove Cover Image'" :buttonFunction="removeCoverImage" />
      <ActionButton v-if="!projectPreview" :icon="getAppIcon('image-plus')" :showLabel="false" v-tooltip="'Add Cover Image'"
        :buttonFunction="addCoverImage" />
    </div>


    <div class="general-container">

      <span @click="addCoverImage" v-if="projectPreview && !displayEmojiSelector" v-tooltip="'Click to change'" class="screenshot-preview">
        <img class="screenshot-thumb" :src="projectPreview">
      </span>

      <div class="input-section">
        <div v-if="!isEditingName" class="project-name-display">
          <span class="project-name-text">{{ projectName }}</span>
          <ActionButton :icon="getAppIcon('edit')" v-tooltip="'Rename Project'" :buttonFunction="toggleEditName" />
        </div>
        <RenameInput 
          v-else
          v-model="projectName" 
          :originalValue="oldProjectName" 
          placeholder="Project Name"
          @confirm="confirmRename"
          @cancel="cancelRename"
        />
      </div>

      <div v-if="!displayEmojiSelector" class="project-stats-section">
        <div class="pane-parameter-detail">
          <div class="simple-text-key">Total Assets</div>
          <div class="simple-text-value">{{ assetsOnDiskCount }} / {{ assetCount }}</div>
        </div>
        <div class="pane-parameter-detail">
          <div class="simple-text-key">Total Collections</div>
          <div class="simple-text-value">{{ collectionsOnDiskCount }} / {{ collectionCount }}</div>
        </div>
        <div class="pane-parameter-detail">
          <div class="simple-text-key">Collaborators</div>
          <div class="simple-text-value">{{ collaboratorCount }}</div>
        </div>
        <div class="pane-parameter-detail">
          <div class="simple-text-key">Files on disk</div>
          <div class="simple-text-value">{{ projectSize }}</div>
        </div>
        <div class="pane-parameter-detail">
          <div class="simple-text-key">Clustta file size</div>
          <div class="simple-text-value">{{ clusttaSize }}</div>
        </div>
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
import { DialogService, ProjectService, FSService, AssetService, CollectionService, UserService } from '@/services';
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
import RenameInput from '@/instances/desktop/components/RenameInput.vue'

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

// Project stats refs
const projectSize = ref(0);
const clusttaSize = ref(0);
const assetCount = ref(0);
const assetsOnDiskCount = ref(0);
const collectionCount = ref(0);
const collectionsOnDiskCount = ref(0);
const collaboratorCount = ref(0);

const displayEmojiSelector = ref(false);
const isEditingName = ref(false);

const toggleEmojiSelector = () => {
  displayEmojiSelector.value = !displayEmojiSelector.value
};

const toggleEditName = () => {
  isEditingName.value = !isEditingName.value;
};

const confirmRename = () => {
  isEditingName.value = false;
};

const cancelRename = () => {
  projectName.value = oldProjectName.value;
  isEditingName.value = false;
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

// Project stats functions
const getProjectSize = async () => {
  let project = projectStore.activeProject;
  const size = await FSService.FolderSize(project.working_directory);
  projectSize.value = size;
};

const getItemsCount = async () => {
  let project = projectStore.activeProject;
  assetsOnDiskCount.value = await FSService.FileCount(project.working_directory);
  collectionsOnDiskCount.value = await FSService.FolderCount(project.working_directory);
};

const getClusttaSize = async () => {
  let project = projectStore.activeProject;
  const size = await FSService.FileStat(project.uri);
  clusttaSize.value = size.formattedSize;
};

const getAssetCount = async () => {
  let project = projectStore.activeProject;
  assetCount.value = await AssetService.GetAssetCount(project.uri);
};

const getCollectionCount = async () => {
  let project = projectStore.activeProject;
  collectionCount.value = await CollectionService.GetCollectionCount(project.uri);
};

const getCollaboratorCount = async () => {
  let project = projectStore.activeProject;
  const users = await UserService.GetUsers(project.uri);
  collaboratorCount.value = users?.length || 0;
};

const getProjectData = async () => {
  let project = projectStore.activeProject;
  if (!project?.uri) return;
  getItemsCount();
  getProjectSize();
  getClusttaSize();
  getAssetCount();
  getCollectionCount();
  getCollaboratorCount();
};

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

  // Fetch project stats
  getProjectData();
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

.input-section {
  width: 100%;
}

.project-name-display {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  width: 100%;
  padding-left: .8rem;
  box-sizing: border-box;
}

.project-name-text {
  flex: 1;
  color: var(--white);
}

.project-stats-section {
  display: flex;
  flex-direction: column;
  width: 100%;
  gap: 5px;
  padding: 0.5rem 0.8rem;
  box-sizing: border-box;
  background: var(--midnight-steel);
  border-radius: var(--large-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.pane-parameter-detail {
  display: flex;
  font-size: 14px;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  color: var(--white);
}

.simple-text-key {
  white-space: nowrap;
  font-size: 13px;
  opacity: 0.7;
}

.simple-text-value {
  text-overflow: ellipsis;
  font-size: 13px;
  font-size: 14px;
  font-family: Inter, sans-serif;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  color: var(--white);
  font-size: 14px;
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
