<template>

  <div class="modal-container" v-esc="closeModal" v-return="handleEnterKey">
    <div class="general-pane-header">
      <HeaderArea :notModal="true" v-if="isCustomIcon" :title="title" :customIcon="projectIcon" />
      <HeaderArea :notModal="true" v-else :title="title" :emoji="projectIcon" />
      <ActionButton v-if="displayEmojiSelector"  :icon="getAppIcon('arrow-left')" :showLabel="false" v-tooltip="$t('modals.backToDetails')"
        :buttonFunction="toggleEmojiSelector" />
      <ActionButton v-else  :icon="getAppIcon('face-plus')" :showLabel="false" v-tooltip="$t('modals.setProjectIcon')"
        :buttonFunction="toggleEmojiSelector" />
      <ActionButton v-if="isPreviewChanged" :icon="getAppIcon('revert')" :showLabel="false"
        v-tooltip="$t('modals.revertCoverImage')" :buttonFunction="revertCoverImage" />
      <ActionButton v-if="projectPreview && !displayEmojiSelector" :icon="getAppIcon('image-cancel')" :showLabel="false"
        v-tooltip="$t('modals.removeCoverImage')" :buttonFunction="removeCoverImage" />
      <ActionButton v-if="!projectPreview" :icon="getAppIcon('image-plus')" :showLabel="false" v-tooltip="$t('modals.addCoverImage')"
        :buttonFunction="addCoverImage" />
    </div>


    <div class="general-container">

      <span @click="addCoverImage" v-if="projectPreview && !displayEmojiSelector" class="screenshot-preview">
        <img class="screenshot-thumb" :src="projectPreview">
      </span>

      <div class="input-section">
        <input v-model="projectName" class="input-short" type="text" :placeholder="$t('placeholders.projectName')" v-focus />
      </div>

      <div v-if="displayEmojiSelector" class="header-tab-container">
        <div class="tab-button" :class="{ 'selected-tab-button': iconType === 'emoji', 'fullwidth-tab-button': true }"
          @click="changeIconType('emoji')">
          {{ $t('modals.emoji') }}
        </div>
        <div class="tab-button" :class="{ 'selected-tab-button': iconType === 'upload', 'fullwidth-tab-button': true }"
          @click="changeIconType('upload')">
          {{ $t('modals.upload') }}
        </div>
      </div>

      <EmojiPicker v-if="displayEmojiSelector && iconType == 'emoji'" @select="handleEmojiSelect" />
      <div v-if="displayEmojiSelector && iconType == 'upload'">
        
      <ActionButton  :icon="getAppIcon('image-plus')" :label="$t('modals.uploadAnImage')" :buttonFunction="selectIcon" />

      </div>

      <div v-if="showRemoteToggle" class="input-section">
        <div class="horizontal-flex toggle-row" @click="toggleRemote">
          <span class="input-label">{{ $t('modals.enableRemote') }}</span>
          <ToggleSwitch :switchValueProp="isRemoteEnabled" />
        </div>
        <p v-if="remoteWarning" class="remote-warning">{{ remoteWarning }}</p>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.close')" :fullWidth="true" :buttonFunction="closeModal" :isActive="!isAwaitingResponse" :colored="false" />
        <GeneralButton :label="$t('common.update')" :fullWidth="true" @click="updateProject()" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>


    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import utils from '@/services/utils';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import EmojiPicker from '@/instances/desktop/components/EmojiPicker.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// services
import { DialogService, FSService, ProjectService } from '@/services';

// stores
import { useAccountStore } from '@/stores/accounts';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useEntitlementStore } from '@/stores/entitlements';
import { refreshEntitlements } from '@/lib/sync';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';

const accountStore = useAccountStore();
const entitlementStore = useEntitlementStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const { t } = useI18n();

// refs
const coverImagePath = ref('');
const displayEmojiSelector = ref(false);
const fileIsSelected = ref(false);
const iconType = ref('emoji');
const isAwaitingResponse = ref(false);
const isRemoteEnabled = ref(false);
const oldProjectIcon = ref('');
const oldProjectName = ref('');
const oldProjectPreview = ref('');
const originalRemoteState = ref(false);
const projectIcon = ref('');
const projectName = ref('');
const projectPreview = ref('');
const projectsDirectory = ref('');
const selectedEmoji = ref('');

// constants
const MAX_PREVIEW_BYTES = 2 * 1024 * 1024;
const title = t('modals.projectDetails');

// computed
// Returns whether the project icon is a custom image.
const isCustomIcon = computed(() => projectIcon.value.length > 10);

// Returns whether the project name has changed.
const isNameChanged = computed(() => {
  const restrictedEntries = [oldProjectName.value, ''];
  return !restrictedEntries.includes(projectName.value);
});

// Returns whether the preview image has changed.
const isPreviewChanged = computed(() => {
  return oldProjectPreview.value !== projectPreview.value;
});

// Returns whether the project icon has changed.
const isProjectIconChanged = computed(() => {
  return oldProjectIcon.value !== projectIcon.value;
});

// Returns whether the remote toggle has changed from its original state.
const isRemoteChanged = computed(() => {
  return isRemoteEnabled.value !== originalRemoteState.value;
});

// Returns whether any form values have changed.
const isValueChanged = computed(() => {
  const restrictedEntries = [oldProjectName.value, ''];
  return !restrictedEntries.includes(projectName.value) || isPreviewChanged.value || isProjectIconChanged.value || isRemoteChanged.value;
});

// Returns a warning message when the user is about to disable remote.
const remoteWarning = computed(() => {
  if (!isRemoteChanged.value) return '';
  if (!isRemoteEnabled.value) {
    return t('modals.remoteDisableWarning');
  }
  return '';
});

// Returns whether the remote toggle should be shown.
const showRemoteToggle = computed(() => {
  if (!accountStore.canUseRemoteFeatures) return false;
  if (projectStore.selectedStudio?.name !== 'Personal') return false;
  if (isRemoteEnabled.value) return true;
  return entitlementStore.canCreateRemoteProject;
});

// methods
// Opens a dialog to select a cover image.
const addCoverImage = async () => {
  const result = await DialogService.SelectFileDialog(t('modals.selectImageFile'), '*.png;*.jpg;*.jpeg;*.gif;*.bmp;*.webp');
  if (result) {
    const filePath = result.replace(/\\/g, '/');
    const stat = await FSService.FileStat(filePath);
    if (stat.size > MAX_PREVIEW_BYTES) {
      notificationStore.errorNotification(t('notifications.imageTooLarge', { limit: '2 MB' }), '');
      return;
    }
    projectPreview.value = await utils.base64FromFile(filePath);
    coverImagePath.value = filePath;
    fileIsSelected.value = true;
  }
};

// Changes the icon type between emoji and upload.
const changeIconType = (type) => {
  iconType.value = type;
};

// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles emoji selection from picker.
const handleEmojiSelect = (emojiData) => {
  selectedEmoji.value = emojiData;
  projectIcon.value = selectedEmoji.value.collection;
  displayEmojiSelector.value = false;
};

// Handles enter key press to submit form.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    updateProject();
  }
};

// Removes the current cover image.
const removeCoverImage = () => {
  projectPreview.value = '';
};

// Reverts to the original cover image.
const revertCoverImage = () => {
  projectPreview.value = oldProjectPreview.value;
};

// Opens a dialog to select a custom icon.
const selectIcon = async () => {
  try {
    const result = await DialogService.SelectIconDialog();
    if (result) {
      projectIcon.value = 'data:image/png;base64,' + result;
    }
  } catch (error) {
    notificationStore.errorNotification(t('notifications.iconTooLarge', { limit: '512 KB' }), '');
  }
};

// Toggles the emoji selector visibility.
const toggleEmojiSelector = () => {
  displayEmojiSelector.value = !displayEmojiSelector.value;
};

// Toggles the local remote toggle state.
const toggleRemote = () => {
  isRemoteEnabled.value = !isRemoteEnabled.value;
};

// Applies the remote state change by making the project remote or removing it from remote.
const updateRemoteState = async () => {
  const project = projectStore.activeProject;
  stage.operationActive = true;
  try {
    if (isRemoteEnabled.value) {
      await ProjectService.MakeProjectRemote(project.uri);
    } else {
      await ProjectService.RemoveProjectFromRemote(project.uri);
    }
    const updatedInfo = await ProjectService.ProjectInfo(project.uri);
    await projectStore.refreshProjects();
    const updatedProject = projectStore.projects.find(p => p.name === project.name);
    if (updatedProject) {
      updatedProject.remote = updatedInfo.remote;
      updatedProject.has_remote = updatedInfo.has_remote;
      projectStore.activeProject = updatedProject;
    }
    refreshEntitlements();
  } catch (error) {
    console.error(error);
    const errorKey = isRemoteEnabled.value ? 'errorMakingProjectRemote' : 'errorRemovingProjectFromRemote';
    notificationStore.errorNotification(t(`notifications.${errorKey}`), error);
  } finally {
    stage.operationActive = false;
  }
};

// Updates the project with all changed values.
const updateProject = async () => {
  isAwaitingResponse.value = true;
  if (isRemoteChanged.value) {
    await updateRemoteState();
  }
  if (isPreviewChanged.value) {
    await updateProjectCover();
  }
  if (isProjectIconChanged.value) {
    await updateProjectIcon();
  }
  if (isNameChanged.value) {
    await updateProjectMeta();
  }
  isAwaitingResponse.value = false;
  closeModal();
};

// Updates the project cover image.
const updateProjectCover = async () => {
  await ProjectService.UpdatePreview(projectStore.activeProject.uri, coverImagePath.value)
    .then(() => {
      projectStore.refreshProjectPreview(projectStore.activeProject.id);
    })
    .catch((error) => {
      console.error(error);
      notificationStore.addNotification(t('notifications.errorUpdatingImage'), error, 'error', false);
    });
};

// Updates the project icon.
const updateProjectIcon = async () => {
  if (projectStore.activeProject.has_remote) {
    await ProjectService.UpdateIcon(projectStore.getActiveProjectUrl, projectStore.selectedStudio.name, projectIcon.value)
      .then(() => {
        projectStore.activeProject.icon = projectIcon.value;
        const index = projectStore.projects.findIndex(project => project.id === projectStore.activeProject.id);
        projectStore.projects[index].icon = projectIcon.value;
      })
      .catch((error) => {
        console.error(error);
        notificationStore.addNotification(t('notifications.errorUpdatingIcon'), error, 'error', false);
      });
  } else {
    await ProjectService.UpdateIcon(projectStore.activeProject.uri, projectStore.selectedStudio.name, projectIcon.value)
      .then(() => {
        projectStore.activeProject.icon = projectIcon.value;
        const index = projectStore.projects.findIndex(project => project.id === projectStore.activeProject.id);
        projectStore.projects[index].icon = projectIcon.value;
      })
      .catch((error) => {
        console.error(error);
        notificationStore.addNotification(t('notifications.errorUpdatingIcon'), error, 'error', false);
      });
  }
};

// Updates the project metadata (name).
const updateProjectMeta = async () => {
  if (projectStore.activeProject.has_remote) {
    ProjectService.Rename(projectStore.getActiveProjectUrl, projectStore.selectedStudio.name, projectName.value)
      .then(() => {
        projectStore.activeProject.name = projectName.value;
      })
      .catch(error => {
        console.log(error);
      });
  } else {
    ProjectService.Rename(projectStore.activeProject.uri, projectStore.selectedStudio.name, projectName.value)
      .then(() => {
        projectStore.activeProject.name = projectName.value;
      })
      .catch(error => {
        console.log(error);
      });
  }
};

// lifecycle hooks
onMounted(() => {
  const project = projectStore.activeProject;
  projectsDirectory.value = project.working_directory;
  projectName.value = project.name;
  oldProjectName.value = project.name;
  projectIcon.value = project.icon;
  oldProjectIcon.value = project.icon;
  projectPreview.value = project.preview;
  oldProjectPreview.value = project.preview;
  isRemoteEnabled.value = !!project.has_remote;
  originalRemoteState.value = !!project.has_remote;
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

.general-pane-header{
  box-sizing: border-box;
  background-color: var(--midnight-steel);
  border-radius: var(--small-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  gap: .5rem;
  padding: 0 1rem;
  padding-left: 0px;
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

.remote-warning {
  font-size: 13px;
  color: var(--danger);
  line-height: 1.4;
  margin: 0;
  padding: 4px;
  padding-bottom: 0;
  font-weight: 400;
}

.toggle-row {
  cursor: pointer;
  align-items: center;
  justify-content: space-between;
}

.input-section {
  width: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  gap: .4px;
  color: var(--white);
}
</style>

