<template>
  <div class="project-item-root" v-right-click="openMenu" v-stop-propagation
    :class="{ 'project-item-container-selected': projectStore.activeProject?.id === project.id, 'project-item-root-cards': cardView }"
    @click="selectProject(project, $event)" @dblclick="launchProject(project)">
    
      <TabbedFolder v-if="cardView && !project.is_tracked">
        <div class="project-item-container-footer" :class="{ 'project-item-container-footer-cards': cardView }">

          <div v-if="!isEditing" class="project-item-content" :class="{ 'project-item-content-cards': cardView }">
            <div class="project-item-details">
              <span v-if="isCreatingProject">Adding {{ project.name }} to Clustta</span>
              <span v-else>{{ utils.capitalizeStr(project.name) }}</span>
            </div>
          </div>

          <RenameInput 
            v-else
            v-model="editableProjectName"
            :originalValue="project.name"
            placeholder="Project name"
            @confirm="confirmRename"
            @cancel="cancelRename"
          />

          <div v-if="!isEditing" class="project-item-actions">
            <ActionButton v-if="project.has_remote && project.is_unsynced" :icon="getAppIcon('dot-big')" :useAlert="true" :noFilter="true"
              v-tooltip="'Project not synced'" />
            <ActionButton v-if="!platformStore.isWeb && isProjectPinned && project.is_downloaded" :icon="getAppIcon('unpin')"
              v-tooltip="'Unpin Project'" @click="unpinProject" />
            <ActionButton v-if="project.is_downloaded || platformStore.isWeb" :icon="getAppIcon('launch')" v-tooltip="'Go to project'"
              @click="goToProject(project)" />
            <div v-if="isCreatingProject" class="loading-spinner">
              <img class="small-icons loading-project-icon" :src="getAppIcon('loading')">
            </div>
            <ActionButton v-if="!platformStore.isWeb && !project.has_remote" :icon="getAppIcon('folder-arrow-up-right')" v-tooltip="'Open folder'"
              @click="revealInExplorer" />
            <ActionButton v-else-if="!platformStore.isWeb && project.is_downloaded" :icon="getAppIcon('folder-arrow-up-right')"
              v-tooltip="project.is_downloaded ? 'Open folder' : 'Download Project'" @click="revealInExplorer" />
            <ActionButton v-else-if="!platformStore.isWeb" :icon="getAppIcon('cloud-down')" v-tooltip="'Download Project'"
              @click="cloneProject(project)" />
          </div>
        </div>
      </TabbedFolder>
    
    <div v-else class="project-item-container" :class="{ 'project-item-container-cards': cardView }">

      <div class="project-item-preview-container" :class="{ 'project-item-preview-container-cards': cardView }">
        <div class="project-item-preview-image">
          <img class="screenshot-thumb" :class="{'no-thumb' : !project.preview }" :src="project.preview ? project.preview : '/page-states/no_image.png'">
        </div>
      </div>

      <div class="project-item-container-footer" :class="{ 'project-item-container-footer-cards': cardView }">

        <div v-if="!isEditing" class="project-item-content" :class="{ 'project-item-content-cards': cardView }">
          <div class="project-item-details">
            <span v-if="isCreatingProject">Launching {{ project.name }}</span>
            <span v-else>{{ utils.capitalizeStr(project.name) }}</span>
          </div>
        </div>

        <RenameInput 
          v-else
          v-model="editableProjectName"
          :originalValue="project.name"
          placeholder="Project name"
          @confirm="confirmRename"
          @cancel="cancelRename"
        />

        <div v-if="!isEditing" class="project-item-actions">
          <ActionButton v-if="project.has_remote && project.is_unsynced" :icon="getAppIcon('dot-big')" :useAlert="true" :noFilter="true"
            v-tooltip="'Project not synced'" />
          <ActionButton v-if="!platformStore.isWeb && isProjectPinned && project.is_downloaded" :icon="getAppIcon('unpin')"
            v-tooltip="'Unpin Project'" @click="unpinProject" />
          <ActionButton v-if="project.is_downloaded || platformStore.isWeb && !isCreatingProject" :icon="getAppIcon('launch')" v-tooltip="'Go to project'"
            @click="goToProject(project)" />
          <ActionButton v-else :isLoading="true" :icon="getAppIcon('loading')"  
					v-tooltip="'Fetching data'" />
          <ActionButton v-if="!platformStore.isWeb && !project.has_remote" :icon="getAppIcon('folder-arrow-up-right')" v-tooltip="'Open folder'"
            @click="revealInExplorer" />
          <ActionButton v-else-if="!platformStore.isWeb && project.is_downloaded" :icon="getAppIcon('folder-arrow-up-right')"
            v-tooltip="project.is_downloaded ? 'Open folder' : 'Download Project'" @click="revealInExplorer" />
          <ActionButton v-else-if="!platformStore.isWeb" :icon="getAppIcon('cloud-down')" v-tooltip="'Download Project'"
            @click="cloneProject(project)" />
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
// imports
import { computed, ref, onMounted, onBeforeUnmount, watch } from 'vue';
import { Events } from "@wailsio/runtime";
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import RenameInput from '@/instances/desktop/components/RenameInput.vue';
import TabbedFolder from '@/instances/desktop/components/TabbedFolder.vue';

// state imports
import { useTrayStates } from '@/stores/TrayStates';
import { useMenu } from '@/stores/menu';
import { useStageStore } from '@/stores/stages';
import { useUserStore } from '@/stores/users';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useIconStore } from '@/stores/icons';
import { usePlatformStore } from '@/stores/platform';
import { FSService, ProjectService, SettingsService, SyncService } from '@/services';

// states/stores
const userStore = useUserStore();
const trayStates = useTrayStates();
const menu = useMenu();
const stage = useStageStore();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const notificationStore = useNotificationStore();
const iconStore = useIconStore();
const platformStore = usePlatformStore();

// props
const props = defineProps({
  project: Object,
  index: Number,
  cardView: { type: Boolean, default: true }
});

// refs
const isCreatingProject = ref(false);
const isEditing = ref(false);
const editableProjectName = ref(props.project.name);

// computed props
// Check if any operations are currently active
const operationsActive = computed(() => {
  return stage.operationActive || !!modals.activeModal || !!menu.activeMenu || isEditing.value || stage.activeStage !== 'projects';
});

// Check if current project is the active one
const isProjectInFocus = computed(() => {
  return projectStore.activeProject?.id === props.project.id;
});

// Check if project is pinned (not applicable in web mode)
const isProjectPinned = computed(() => {
  if (platformStore.isWeb) return false;
  const projectId = props.project.id;
  const pinnedProjects = projectStore.pinnedProjects;
  return pinnedProjects?.includes(projectId);
});


// functions
// Get icon from icon store
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon;
};

// Toggle edit mode for renaming
const toggleEditMode = (event) => {
  isEditing.value = !isEditing.value;
};

// Start rename operation
const startRename = () => {
  toggleEditMode();
};

// Cancel rename and restore original name
const cancelRename = () => {
  editableProjectName.value = props.project.name;
  toggleEditMode();
};

// Confirm rename and update project name
const confirmRename = async () => {
  await updateProjectName();
  toggleEditMode();
};

// Initiate rename from menu if project is in focus
const menuRename = () => {
  if (isProjectInFocus.value && userStore.userCanCreateProject) {
    startRename();
  }
};

// Update project name in backend and store
const updateProjectName = async () => {
  let project = projectStore.activeProject;
  const oldUri = project.uri;
  const oldUrl = projectStore.getActiveProjectUrl;
  const projectId = project.id;
  const oldName = project.name; 
  
  if (project.has_remote) {
    ProjectService.Rename(oldUrl, projectStore.selectedStudio.name, editableProjectName.value)
      .then((data) => {
        projectStore.updateProjectName(projectId, editableProjectName.value, oldName);
        selectProject(project);
      }).catch(error => {
        console.log(error);
        notificationStore.addNotification("Error", "Failed to rename project", "error");
      });
  } else {
    ProjectService.Rename(oldUri, projectStore.selectedStudio.name, editableProjectName.value)
      .then((data) => {
        projectStore.updateProjectName(projectId, editableProjectName.value, oldName);
        selectProject(project);
      }).catch(error => {
        console.log(error);
        notificationStore.addNotification("Error", "Failed to rename project", "error");
      });
  }
};

// Unpin project from pinned list
const unpinProject = async () => {
  const studioName = projectStore.getSelectedStudioName;
  const projectId = props.project.id;

  await SettingsService.UnpinProject(studioName, projectId).then((response) => {
    projectStore.pinnedProjects = projectStore.pinnedProjects.filter((item) => item !== projectId);
  }).catch((error) => {
    console.log(error);
  });
};

// Reveal project directory in file explorer
const revealInExplorer = async () => {
  let project = projectStore.getActiveProject;
  await FSService.MakeDirs(project.working_directory);
  FSService.RevealInExplorer(project.working_directory);
  menu.hideContextMenu();
};

// Launch project - clone if not downloaded, otherwise navigate to it
const launchProject = async (project) => {
  if (!platformStore.isWeb && !project.is_downloaded && project.is_tracked) {
    cloneProject(project);
  } else {
    goToProject(project);
  }
};

// Show clone project modal
const cloneProject = async (project) => {
  await projectStore.setActiveProject(project);
  modals.setModalVisibility('cloneProjectModal', true);
};

// Navigate to project or create if untracked
const goToProject = async (project) => {
  console.log(project)
  if (!project.is_tracked) {
    try {
      isCreatingProject.value = true;
      
      const projectsDir = await SettingsService.GetProjectDirectory();
      const projectUri = `${projectsDir}/${project.name}.clst`;
      const studioName = projectStore.selectedStudio.name;
      const workingDir = project.working_directory;
      const templateName = "";
      
      const createdProject = await ProjectService.CreateProject(
        projectUri,
        studioName,
        workingDir,
        templateName
      );
      
      createdProject.is_tracked = true;
      
      const projectIndex = projectStore.projects.findIndex(p => p.name === project.name);
      if (projectIndex !== -1) {
        projectStore.projects[projectIndex] = createdProject;
      }
            
      projectStore.gotoProject(createdProject);
      
      isCreatingProject.value = false;
    } catch (error) {
      console.error('Error creating project from untracked directory:', error);
      notificationStore.errorNotification('Error creating project', error);
      isCreatingProject.value = false;
    }
  } else {
    if (!platformStore.isWeb && !project.is_downloaded) {
      return;
    }
    
    // In web mode, sync project data before navigating
    if (platformStore.isWeb) {
      try {
        isCreatingProject.value = true;
        const studioUrl = projectStore.studioUrl;
        await SyncService.SyncData(project.name, studioUrl, false, {});
        isCreatingProject.value = false;
      } catch (error) {
        console.error('Error syncing project data:', error);
        notificationStore.errorNotification('Error loading project', error);
        isCreatingProject.value = false;
        return;
      }
    }
    
    projectStore.gotoProject(project);
  }
};

// Open context menu for project
const openMenu = (event) => {
  if (!props.project.is_downloaded && !platformStore.isWeb) return;
  if (!props.project.is_tracked) {
    trayStates.popUpModalTitle = `Add "${props.project.name}" to Clustta?`;
    trayStates.popUpModalMessage = "This project is not yet in Clustta. Click CONFIRM to add it and start tracking your work.";
    trayStates.popUpModalFunction = async () => {
      modals.setModalVisibility('popUpModal', false);
      await goToProject(props.project);
    };
    trayStates.popUpModalIcon = 'briefcase-plus';
    modals.setModalVisibility('popUpModal', true);
    return;
  }
  const project = props.project;
  projectStore.setActiveProject(project);
  menu.showContextMenu(event, 'projectItemMenu', true);
};

// Select project and update active state
const selectProject = (project, event) => {
  handleClickOutside(event);
  menu.disableAllMenus();
  projectStore.setActiveProject(project);
  const id = project.id;
  stage.selectdProject = id;
};

// Handle clicks outside rename input
const handleClickOutside = (event) => {
  if (isEditing.value) {
    cancelRename();
  }
};

// watchers
// Reset editing state when project focus changes
watch(() => isProjectInFocus.value, (newItems, oldItems) => {
  if (isEditing.value) {
    isEditing.value = false;
    editableProjectName.value = props.project.name;
  }
}, { deep: true });

// lifecycle hooks
Events.On('edit-item', async () => {
  if (operationsActive.value) return;
  if (isProjectInFocus.value && userStore.userCanCreateProject) {
    modals.setModalVisibility('editProjectModal', true);
  }
});

Events.On('rename-item', async () => {
  if (operationsActive.value) return;
  if (isProjectInFocus.value && userStore.userCanCreateProject) {
    startRename();
  }
});

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

.project-item-root {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: var(--white);
  align-items: center;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-start;
  background-color: var(--dark-steel);
  border-radius: var(--large-radius);
  overflow: hidden;
  min-width: 500px;
  outline: var(--transparent-line);
  outline-offset: -1px;
  transition: all .2s ease-out;
}

.project-item-root-cards {
  min-width: 300px;
  height: 200px;
  border-radius: var(--large-radius);
}

.project-item-root:hover {
  background-color: var(--steel);
  border-radius: var(--small-radius);
}

.project-item-root:hover :deep(.folder) {
  background-color: var(--light-steel);
  border-radius: 0px;
}

.project-item-root:hover :deep(.tab) {
  background-color: var(--light-steel);
  border-radius: 24px 16px 0 0;
}

.project-item-root:hover :deep(.tab::after) {
  box-shadow: -25px 0 0 0 var(--light-steel);
}

.project-item-root:hover  .project-item-preview-container{
  border-radius: var(--tiny-radius);
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
}

.project-item-container-cards {
  height: 100%;
  flex-direction: column;
}

.project-item-container-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  background-color: var(--black-steel);
  outline-offset: -1px;
  background-color: var(--project-item-selected);
}

.project-item-container-selected :deep(.folder) {
  background-color: var(--blue-steel);
}

.project-item-container-selected:hover :deep(.folder) {
  background-color: var(--blue-steel);
}

.project-item-container-selected:hover {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  background-color: var(--black-steel);
  outline-offset: -1px;
  background-color: var(--project-item-selected);
}

.project-item-preview-container {

  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  min-width: 60px;
  height: 100%;
  aspect-ratio: 16 / 9;
  border-radius: 12px;
  transition: all .2s ease-out;
}

.project-item-preview-container-cards {
  border-radius: var(--very-large-radius);
  border-radius: 12px;
  width: 100%;
  aspect-ratio: 16 / 9;

}

.project-item-preview-image {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  height: 100%;
  background-color: var(--black-steel);
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
  font-size: 14px;
}

[data-theme="dark"] .project-item-details{
  font-weight: 200 ;
}

.project-item-container-footer {
  align-items: center;
  display: flex;
  width: min-content;
  box-sizing: border-box;
  width: 100%;
  height: 40px;
  min-height: 40px;
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




