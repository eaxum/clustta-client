<template>
  <div class="project-item-root" v-right-click="openMenu" v-stop-propagation
    :class="{ 'project-item-container-selected': projectStore.activeProject.id === project.id, 'project-item-root-cards': cardView }"
    @click="selectProject(project, $event)" @dblclick="goToProject(project)">
    <div class="project-item-container" :class="{ 'project-item-container-cards': cardView }">

      <div class="project-item-preview-container" :class="{ 'project-item-preview-container-cards': cardView }">
        <div class="project-item-preview-image">
          <img class="screenshot-thumb" :class="{'no-thumb' : !project.preview }" :src="project.preview ? project.preview : '/page-states/no_image.png'">
        </div>
      </div>

      <div class="project-item-container-footer" :class="{ 'project-item-container-footer-cards': cardView }">

        <div v-if="!isEditing" class="project-item-content" :class="{ 'project-item-content-cards': cardView }">
          <div class="project-item-details">
            <span v-if="isCreatingProject">Adding {{ project.name }} to Clustta</span>
            <span v-else>{{ utils.capitalizeStr(project.name) }}</span>
          </div>
        </div>
        <div v-else class="rename-input">
          <input spellcheck="false" ref="renameInput" v-model="editableProjectName" class="input-short" type="text"
            placeholder="Project name" v-focus @keydown.enter="handleEnterKey" @keydown.esc="handleEscKey" />
          <ActionButton :isDisabled="!isNameChanged" :icon="getAppIcon('check')" v-tooltip="'Confirm'"
            @click="confirmRename" />
          <ActionButton :icon="getAppIcon('close')" v-tooltip="'Cancel'" @click="cancelRename" />
        </div>

        <div v-if="!isEditing" class="project-item-actions">
          <ActionButton v-if="!project.is_tracked" :icon="getAppIcon('dot-big')" :useDanger="true" :noFilter="true"
            v-tooltip="'Project not Tracked'" />
          <ActionButton v-if="project.has_remote && project.is_unsynced" :icon="getAppIcon('dot-big')" :useAlert="true" :noFilter="true"
            v-tooltip="'Project not synced'" />
          <ActionButton v-if="isProjectPinned && project.is_downloaded" :icon="getAppIcon('unpin')"
            v-tooltip="'Unpin Project'" @click="unpinProject" />
          <ActionButton v-if="project.is_downloaded" :icon="getAppIcon('launch')" v-tooltip="'Go to project'"
            @click="goToProject(project)" />
          <div v-if="isCreatingProject" class="loading-spinner">
            <img class="small-icons loading-project-icon" :src="getAppIcon('loading')">
          </div>
          <ActionButton v-if="!project.has_remote" :icon="getAppIcon('folder-arrow-up-right')" v-tooltip="'Open folder'"
            @click="revealInExplorer" />
          <ActionButton v-else-if="project.is_downloaded" :icon="getAppIcon('folder-arrow-up-right')"
            v-tooltip="project.is_downloaded ? 'Open folder' : 'Download Project'" @click="revealInExplorer" />
          <ActionButton v-else :icon="getAppIcon('cloud-down')" v-tooltip="'Download Project'"
            @click="cloneProject(project)" />
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
// imports
import { computed, ref, onMounted, onBeforeUnmount, watch } from 'vue';
import utils from '@/services/utils';
import { SettingsService } from "@/../bindings/clustta/services";
import { Events } from "@wailsio/runtime";
import emitter from '@/lib/mitt';

// states/store imports
import { useTrayStates } from '@/stores/TrayStates';
import { useMenu } from '@/stores/menu';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { useUserStore } from '@/stores/users';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useProjectStore } from '@/stores/projects';
import { useAssetStore } from '@/stores/assets';
import { useIconStore } from '@/stores/icons';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import { DialogService, FSService, SyncService, ProjectService } from '@/../bindings/clustta/services/index';

// states/stores
const userStore = useUserStore();
const trayStates = useTrayStates();
const menu = useMenu();
const panes = usePaneStore();
const stage = useStageStore();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const assetStore = useAssetStore();
const notificationStore = useNotificationStore();
const commonStore = useCommonStore();
const collectionStore = useCollectionStore();
const iconStore = useIconStore();

// props
const props = defineProps({
  project: Object,
  index: Number,
  cardView: { type: Boolean, default: true },
});

const renameInput = ref(null);
const isCreatingProject = ref(false);
const operationsActive = computed(() => {
  return stage.operationActive || !!modals.activeModal || !!menu.activeMenu || isEditing.value || stage.activeStage !== 'projects'
});

const isEditing = ref(false);
const editableProjectName = ref(props.project.name);

const isProjectInFocus = computed(() => {
  return projectStore.activeProject.id === props.project.id
});

Events.On('edit-item', async () => {
  if (operationsActive.value) return
  if (isProjectInFocus.value && userStore.userCanCreateProject) {
    modals.setModalVisibility('editProjectModal', true);
  }
});

Events.On('rename-item', async () => {
  if (operationsActive.value) return
  if (isProjectInFocus.value && userStore.userCanCreateProject) {
    startRename();
  }
});

const toggleEditMode = (event) => {
  isEditing.value = !isEditing.value;
};

const cancelRename = () => {
  editableProjectName.value = props.project.name;
  toggleEditMode();
};

const startRename = () => {
  toggleEditMode();
};

const confirmRename = async () => {
  // isAwaitingResponse.value = true;
  await updateProjectName();
  toggleEditMode();
};

const menuRename = () => {
  if (isProjectInFocus.value && userStore.userCanCreateProject) {
    startRename();
  }
}

const handleEnterKey = () => {
  const inputElement = document.querySelector('.input-short');
  if (!isNameChanged.value) {
    if (inputElement) {
      inputElement.classList.add('shake');
      setTimeout(() => {
        inputElement.classList.remove('shake');
      }, 300);
    }
  } else {
    confirmRename();
  }
};

const handleEscKey = () => {
  console.log('canceled');
  if (isEditing.value) {
    cancelRename();
  }
};

const updateProjectName = async () => {

  if (projectStore.activeProject.has_remote) {
    ProjectService.Rename(projectStore.getActiveProjectUrl, projectStore.selectedStudio.name, editableProjectName.value)
      .then((data) => {
        projectStore.activeProject.name = editableProjectName.value
      }).catch(error => {
        console.log(error)
      })
  } else {
    ProjectService.Rename(projectStore.activeProject.uri, projectStore.selectedStudio.name, editableProjectName.value)
      .then((data) => {
        projectStore.activeProject.name = editableProjectName.value
      }).catch(error => {
        console.log(error)
      })
  }

};

const isNameChanged = computed(() => {
  const restrictedEntries = [props.project.name, ''];

  const lowerCaseEditableName = editableProjectName.value.toLowerCase();
  const lowerCaseRestrictedEntries = restrictedEntries.map(entry =>
    typeof entry === 'string' ? entry.toLowerCase() : entry
  );

  return !lowerCaseRestrictedEntries.includes(lowerCaseEditableName);
});


// computed properties
const projectUrl = computed(() => {
  return stage.showProjectCheckboxes || !stage.allProjectsMarked
});
const isProjectPinned = computed(() => {
  const projectId = props.project.id;
  const pinnedProjects = projectStore.pinnedProjects;
  return pinnedProjects?.includes(projectId);
});

// methods
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const unpinProject = async () => {
  const studioName = projectStore.getSelectedStudioName;
  const projectId = props.project.id;

  await SettingsService.UnpinProject(studioName, projectId).then((response) => {
    projectStore.pinnedProjects = projectStore.pinnedProjects.filter((item) => item !== projectId)
  }).catch((error) => {
    console.log(error)
  })

};

const revealInExplorer = async () => {
  let project = projectStore.getActiveProject;
  await FSService.MakeDirs(project.working_directory)
  FSService.RevealInExplorer(project.working_directory)
  menu.hideContextMenu();
};

const cloneProject = async (project) => {
  await projectStore.setActiveProject(project);
  modals.setModalVisibility('cloneProjectModal', true);
}

const goToProject = async (project) => {
  // Check if this is an untracked project
  if (!project.is_tracked) {
    try {
      // Set loading state
      isCreatingProject.value = true;
      
      // Get the projects directory to construct the .clst file path
      const projectsDir = await SettingsService.GetProjectDirectory();
      const projectUri = `${projectsDir}/${project.name}.clst`;
      const studioName = projectStore.selectedStudio.name;
      const workingDir = project.working_directory;
      const templateName = ""; // or "No Template"
      
      // Create the project
      const createdProject = await ProjectService.CreateProject(
        projectUri,
        studioName,
        workingDir,
        templateName
      );
      
      createdProject.is_tracked = true;
      
      // Update the project in the store
      const projectIndex = projectStore.projects.findIndex(p => p.name === project.name);
      if (projectIndex !== -1) {
        projectStore.projects[projectIndex] = createdProject;
      }
            
      // Launch the newly created project
      projectStore.gotoProject(createdProject);
      
      // Clear loading state
      isCreatingProject.value = false;
    } catch (error) {
      console.error('Error creating project from untracked directory:', error);
      notificationStore.errorNotification('Error creating project', error);
      // Clear loading state on error
      isCreatingProject.value = false;
    }
  } else {
    // Normal tracked project flow
    projectStore.gotoProject(project);
  }
};

const openMenu = (event) => {
  if(!props.project.is_tracked){
    // Show popup modal for untracked projects
    trayStates.popUpModalTitle = `Add "${props.project.name}" to Clustta?`;
    trayStates.popUpModalMessage = "This project is not yet in Clustta. Click CONFIRM to add it and start tracking your work.";
    trayStates.popUpModalFunction = async () => {
      modals.setModalVisibility('popUpModal', false);
      await goToProject(props.project);
    };
    trayStates.popUpModalIcon = 'briefcase-plus';
    modals.setModalVisibility('popUpModal', true);
    return
  }
  const project = props.project;
  projectStore.setActiveProject(project);
  menu.showContextMenu(event, 'projectItemMenu', true);
  // prepProjectMenu(index, project, event);
};

const selectProject = (project, event) => {
  handleClickOutside(event);
  menu.disableAllMenus();
  projectStore.setActiveProject(project);
  const id = project.id;
  stage.selectdProject = id;
};

watch(() => isProjectInFocus.value, (newItems, oldItems) => {
  if (isEditing.value) {
    isEditing.value = false;
    editableProjectName.value = props.project.name;
  }
}, { deep: true });

const handleClickOutside = (event) => {
  if (renameInput.value && !renameInput.value.contains(event.target)) {
    cancelRename();
  }

};

onMounted(() => {
  emitter.on('renameProject', menuRename);
  document.addEventListener('click', handleClickOutside);
});

onBeforeUnmount(() => {
  emitter.off('renameProject', menuRename);
  document.removeEventListener('click', handleClickOutside);
});


</script>

<style scoped>
@import "@/assets/desktop.css";

.single-action-button-disabled {
  pointer-events: none;
}

.project-item-root {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: var(--white);
  align-items: center;
  padding: .3rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-start;
  background-color: var(--dark-steel);
  border-radius: var(--normal-radius);
  overflow: hidden;
  min-width: 500px;

  outline: var(--transparent-line);
  outline-offset: -1px;
}

.project-item-root-cards {
  min-width: 300px;
  height: 300px;
}

.project-item-root:hover {
  outline: var(--transparent-line);
  outline: var(--solid-line);
  outline-offset: -1.5px;
  /* background-color: var(--light-steel); */
}

.project-item-container {
  display: flex;
  gap: .5rem;
  color: var(--white);
  align-items: center;
  padding: .4rem;
  box-sizing: border-box;
  width: 100%;
  height: 50px;
  justify-content: space-between;
  /* background-color: crimson; */
  /* border-bottom: var(--transparent-line); */

  /* transition: all .3s ease-out; */
}

.project-item-container-cards {
  /* justify-content: space-around; */
  height: 100%;
  flex-direction: column;
  /* background-color: goldenrod; */
}

.project-item-container-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  background-color: var(--black-steel);
  outline-offset: -1px;
  background-color: var(--project-item-selected);
}

.project-item-container-selected:hover {
  outline: 1px solid rgb(255, 255, 255);
  background-color: var(--black-steel);
  outline-offset: -1px;
  background-color: var(--project-item-selected);
}

.project-item-preview-container {

  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  padding: .1rem;
  overflow: hidden;
  min-width: 60px;
  height: 100%;
  /* background-color: hotpink; */
  aspect-ratio: 16 / 9;
}

.project-item-preview-container-cards {
  /* background-color: firebrick; */
  width: 100%;
  /* height: 60%; */
  aspect-ratio: 16 / 9;
  /* height: 100%; */

}

.project-item-preview-image {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  height: 100%;
  /* aspect-ratio: 16 / 9; */
  background-color: var(--black-steel);
  border-radius: 5px;
  width: 100%;
}

.screenshot-thumb{
  object-fit: cover;
  width: 100%;
  height: 100%;
}

.project-item-content {
  gap: .4rem;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  /* background-color: indigo; */
  width: 100%;
  overflow: hidden;
}

.project-item-content-cards {
  height: max-content;
}

.input-short {
  width: 100%;
  height: 100%;
}

.project-item-details {
  padding: .2rem;
  flex-wrap: nowrap;
  overflow: hidden;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  height: min-content;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-weight: 350;
}

[data-theme="dark"] .project-item-details{
  font-weight: 200 ;
}

.project-item-container-footer {
  align-items: center;
  display: flex;
  width: min-content;
  box-sizing: border-box;
  padding: .2rem;
  width: 100%;
}

.project-item-container-footer-cards {
  /* justify-content: flex-end; */

}

.project-item-actions {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: max-content;
  min-width: max-content;
  gap: .7rem;
  padding: .2rem;
}

.project-item-status-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  padding: .2rem;
  height: 100%;
  /* background-color: darkorange; */
  /* flex: 1; */
}

.project-item-status {
  display: flex;
  border-radius: var(--normal-radius);
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  width: 80px;
  padding: .4rem .4rem;
  height: max-content;
  /* background-color: firebrick; */
  font-size: 12px;
  text-transform: uppercase;
  font-weight: 700;
  color: black;
}

.loading-spinner {
  display: flex;
  align-items: center;
  justify-content: center;
  width: max-content;
  height: max-content;
}

.loading-project-icon {
  width: 20px;
  height: 20px;
  overflow: hidden;
  animation: loadingRotate .5s linear infinite;
}

@keyframes loadingRotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>




