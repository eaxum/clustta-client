<template>
  <div class="project-item-root" v-right-click="openMenu" v-stop-propagation
    :class="{ 'project-item-container-selected': projectStore.activeProject?.id === project.id, 'project-item-root-cards': cardView }"
    @click="selectProject(project, $event)" @dblclick="launchProject(project)">

    <TabbedFolder v-if="cardView && !project.is_tracked">
      <div class="project-item-container-footer" :class="{ 'project-item-container-footer-cards': cardView }">
        <div  class="project-item-content" :class="{ 'project-item-content-cards': cardView }">
          <div class="project-item-details">
            <span v-if="isCreatingProject">{{ $t('notifications.addingToClustta', { name: project.name }) }}</span>
            <span v-else>{{ utils.capitalizeStr(project.name) }}</span>
          </div>
        </div>
        <div v-if="!isEditing" class="project-item-actions">
          <ActionButton class="hover-action" :isLoading="isCreatingProject" :icon="getAppIcon(isCreatingProject ? 'loading' : 'plus-circle')"
            v-tooltip="$t('notifications.startTrackingWithClustta')" @click="goToProject(project)" />
          <ActionButton class="hover-action" :icon="getAppIcon('folder-arrow-up-right')"
            v-tooltip="$t('notifications.openFolder')" @click="revealInExplorer" />
        </div>
      </div>
    </TabbedFolder>

    <div v-else-if="!cardView && !project.is_tracked" class="project-item-container untracked-list-item">
      <div class="project-item-icon">
        <img class="large-icons" :src="getAppIcon('folder')">
      </div>
      <div class="project-item-content">
        <div class="project-item-details">
          <span v-if="isCreatingProject">{{ $t('notifications.addingToClustta', { name: project.name }) }}</span>
          <span v-else>{{ utils.capitalizeStr(project.name) }}</span>
        </div>
        <div class="project-item-path">{{ project.working_directory }}</div>
      </div>
      <div class="project-item-actions">
        <ActionButton class="hover-action" :isLoading="isCreatingProject" :icon="getAppIcon(isCreatingProject ? 'loading' : 'plus-circle')"
          v-tooltip="$t('notifications.startTrackingWithClustta')" @click="goToProject(project)" />
        <ActionButton class="hover-action" :icon="getAppIcon('folder-arrow-up-right')"
          v-tooltip="$t('notifications.openFolder')" @click="revealInExplorer" />
      </div>
    </div>

    <div v-else class="project-item-container" :class="{ 'project-item-container-cards': cardView }">
      <div class="project-item-preview-container" :class="{ 'project-item-preview-container-cards': cardView }">
        <div class="project-item-preview-image">
          <img class="screenshot-thumb" :class="{ 'no-thumb': !project.preview }" :src="project.preview || '/page-states/no_image.png'">
        </div>
      </div>

      <div class="project-item-container-footer" :class="{ 'project-item-container-footer-cards': cardView }">
        <div v-if="!isEditing" class="project-item-content" :class="{ 'project-item-content-cards': cardView }">
          <div class="project-item-details">
            <span v-if="isCreatingProject">{{ $t('notifications.launchingProject', { name: project.name }) }}</span>
            <span v-else>{{ utils.capitalizeStr(project.name) }}</span>
          </div>
        </div>

        <RenameInput v-else v-model="editableProjectName" :originalValue="project.name" :placeholder="$t('placeholders.projectName')"
          @confirm="confirmRename" @cancel="cancelRename" />

        <div v-if="!isEditing" class="project-item-actions">
          <div class="project-item-actions-hover">
            <ActionButton v-if="!platformStore.isWeb && !project.has_remote" :icon="getAppIcon('folder-arrow-up-right')"
              v-tooltip="$t('notifications.openFolder')" @click="revealInExplorer" />
            <ActionButton v-else-if="!platformStore.isWeb && project.is_downloaded" :icon="getAppIcon('folder-arrow-up-right')"
              v-tooltip="$t('notifications.openFolder')" @click="revealInExplorer" />
          </div>
          <div class="project-item-actions-persistent">
            <ActionButton v-if="!platformStore.isWeb && isProjectPinned && project.is_downloaded" :icon="getAppIcon('pin')"
              v-tooltip="$t('blocks.unpinProject')" @click="unpinProject" />
            <ActionButton v-if="project.has_remote && project.is_unsynced" :icon="getAppIcon('dot-big')" :useAlert="true"
              :noFilter="true" v-tooltip="$t('notifications.projectNotSynced')" />
            <ActionButton v-if="!platformStore.isWeb && !project.is_downloaded" :icon="getAppIcon('cloud-down')" v-tooltip="$t('notifications.downloadProject')"
              @click="cloneProject(project)" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { Events } from "@wailsio/runtime";
import emitter from '@/lib/mitt';
import { useI18n } from 'vue-i18n';
import utils from '@/services/utils';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import RenameInput from '@/instances/desktop/components/RenameInput.vue';
import TabbedFolder from '@/instances/desktop/components/TabbedFolder.vue';

// services
import { FSService, ProjectService, SettingsService, SyncService } from '@/services';

// store imports
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

// stores
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

const { t } = useI18n();

// props
const props = defineProps({
  project: Object,
  index: Number,
  cardView: { type: Boolean, default: true }
});

// refs
const editableProjectName = ref(props.project.name);
const isCreatingProject = ref(false);
const isEditing = ref(false);

// computed properties
const isProjectInFocus = computed(() => projectStore.activeProject?.id === props.project.id);

const isProjectPinned = computed(() => {
  if (platformStore.isWeb) return false;
  return projectStore.pinnedProjects?.includes(props.project.id);
});

const operationsActive = computed(() => {
  return stage.operationActive || !!modals.activeModal || !!menu.activeMenu || isEditing.value || stage.activeStage !== 'projects';
});

// methods

// Cancels the rename operation and restores the original name.
const cancelRename = () => {
  editableProjectName.value = props.project.name;
  isEditing.value = false;
};

// Shows the clone project modal.
const cloneProject = async (project) => {
  await projectStore.setActiveProject(project);
  modals.setModalVisibility('cloneProjectModal', true);
};

// Confirms the rename and updates the project name.
const confirmRename = async () => {
  await updateProjectName();
  isEditing.value = false;
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Navigates to the project or creates it if untracked.
const goToProject = async (project) => {
  if (!project.is_tracked) {
    try {
      isCreatingProject.value = true;
      const projectsDir = await SettingsService.GetProjectDirectory();
      const projectUri = `${projectsDir}/${project.name}.clst`;
      const createdProject = await ProjectService.CreateProject(projectUri, projectStore.selectedStudio.name, project.working_directory, "");
      createdProject.is_tracked = true;
      const projectIndex = projectStore.projects.findIndex(p => p.name === project.name);
      if (projectIndex !== -1) projectStore.projects[projectIndex] = createdProject;
      projectStore.gotoProject(createdProject);
    } catch (error) {
      console.error('Error creating project from untracked directory:', error);
      notificationStore.errorNotification(t('notifications.errorCreatingProject'), error);
    } finally {
      isCreatingProject.value = false;
    }
    return;
  }
  if (!platformStore.isWeb && !project.is_downloaded) return;
  if (platformStore.isWeb) {
    try {
      isCreatingProject.value = true;
      await SyncService.SyncData(project.name, projectStore.studioUrl, false, {});
    } catch (error) {
      console.error('Error syncing project data:', error);
      notificationStore.errorNotification(t('notifications.errorLoadingProject'), error);
      return;
    } finally {
      isCreatingProject.value = false;
    }
  }
  projectStore.gotoProject(project);
};

// Handles clicks outside the rename input to cancel editing.
const handleClickOutside = () => {
  if (isEditing.value) cancelRename();
};

// Launches the project - clones if not downloaded, otherwise navigates to it.
const launchProject = async (project) => {
  if (!platformStore.isWeb && !project.is_downloaded && project.is_tracked) cloneProject(project);
  else goToProject(project);
};

// Initiates rename from menu if the project is in focus.
const menuRename = () => {
  if (isProjectInFocus.value && userStore.userCanCreateProject) isEditing.value = true;
};

// Opens the context menu for the project item.
const openMenu = (event) => {
  if (!props.project.is_downloaded && !platformStore.isWeb) return;
  if (!props.project.is_tracked) {
    trayStates.popUpModalTitle = t('notifications.addToClustta', { name: props.project.name });
    trayStates.popUpModalMessage = t('notifications.addToClusttaMessage');
    trayStates.popUpModalIcon = 'briefcase-plus';
    trayStates.popUpModalFunction = async () => {
      modals.setModalVisibility('popUpModal', false);
      await goToProject(props.project);
    };
    modals.setModalVisibility('popUpModal', true);
    return;
  }
  projectStore.setActiveProject(props.project);
  menu.showContextMenu(event, 'projectItemMenu', true);
};

// Reveals the project directory in the file explorer.
const revealInExplorer = async () => {
  const project = projectStore.getActiveProject;
  await FSService.MakeDirs(project.working_directory);
  FSService.RevealInExplorer(project.working_directory);
  menu.hideContextMenu();
};

// Selects the project and updates the active state.
const selectProject = (project, event) => {
  handleClickOutside();
  menu.disableAllMenus();
  projectStore.setActiveProject(project);
  stage.selectdProject = project.id;
};

// Unpins the project from the pinned list.
const unpinProject = async () => {
  try {
    await SettingsService.UnpinProject(projectStore.getSelectedStudioName, props.project.id);
    projectStore.pinnedProjects = projectStore.pinnedProjects.filter((item) => item !== props.project.id);
  } catch (error) {
    console.error('Error unpinning project:', error);
  }
};

// Updates the project name in the backend and store.
const updateProjectName = async () => {
  const project = projectStore.activeProject;
  const projectPath = project.has_remote ? projectStore.getActiveProjectUrl : project.uri;
  try {
    await ProjectService.Rename(projectPath, projectStore.selectedStudio.name, editableProjectName.value);
    projectStore.updateProjectName(project.id, editableProjectName.value, project.name);
    selectProject(project);
  } catch (error) {
    notificationStore.addNotification(t('common.error'), t('notifications.failedToRenameProject'), "error");
  }
};

// watchers
watch(() => isProjectInFocus.value, () => {
  if (isEditing.value) {
    isEditing.value = false;
    editableProjectName.value = props.project.name;
  }
});

// events
Events.On('edit-item', async () => {
  if (operationsActive.value) return;
  if (isProjectInFocus.value && userStore.userCanCreateProject) modals.setModalVisibility('editProjectModal', true);
});

Events.On('rename-item', async () => {
  if (operationsActive.value || !props.project.is_tracked) return;
  if (isProjectInFocus.value && userStore.userCanCreateProject) isEditing.value = true;
});

// lifecycle hooks
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

.project-item-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: .4rem;
  width: max-content;
  min-width: max-content;
  padding: .2rem;
  box-sizing: border-box;
}

.project-item-actions-hover {
  display: flex;
  align-items: center;
  gap: .4rem;
  opacity: 0;
  transition: opacity .15s ease-out;
}

.project-item-actions-persistent {
  display: flex;
  align-items: center;
  gap: .4rem;
}

.project-item-root:hover .project-item-actions-hover,
.project-item-root:hover .hover-action {
  opacity: 1;
}

.hover-action {
  opacity: 0;
  transition: opacity .15s ease-out;
}

.project-item-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  width: 100%;
  height: 50px;
  padding: .4rem;
  box-sizing: border-box;
  color: var(--white);
}

.project-item-container-cards {
  height: 100%;
  flex-direction: column;
}

.project-item-container-footer {
  display: flex;
  align-items: center;
  width: 100%;
  height: 40px;
  min-height: 40px;
  box-sizing: border-box;
}

.project-item-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .4rem;
  width: 100%;
  height: 100%;
  box-sizing: border-box;
  overflow: hidden;
}

.project-item-content-cards {
  height: max-content;
}

.project-item-details {
  height: min-content;
  padding: .2rem;
  box-sizing: border-box;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-size: 14px;
  font-weight: 350;
}

[data-theme="dark"] .project-item-details {
  font-weight: 200;
}

.project-item-preview-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 60px;
  height: 100%;
  aspect-ratio: 16 / 9;
  box-sizing: border-box;
  overflow: hidden;
  border-radius: 12px;
  transition: all .2s ease-out;
}

.project-item-preview-container-cards {
  width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: 12px;
}

.project-item-preview-image {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  box-sizing: border-box;
  overflow: hidden;
  background-color: var(--black-steel);
}

.project-item-root {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  gap: .2rem;
  width: 100%;
  height: min-content;
  min-width: 500px;
  box-sizing: border-box;
  overflow: hidden;
  color: var(--white);
  background-color: var(--dark-steel);
  border-radius: var(--large-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  transition: all .2s ease-out;
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

.project-item-root:hover .project-item-preview-container {
  border-radius: var(--tiny-radius);
}


.project-item-container-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--project-item-selected);
}

.project-item-container-selected :deep(.folder),
.project-item-container-selected:hover :deep(.folder) {
  background-color: var(--blue-steel);
}

.project-item-container-selected:hover {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--project-item-selected);
}

.project-item-root-cards {
  min-width: 300px;
  height: 200px;
  border-radius: var(--large-radius);
}

.screenshot-thumb {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* Untracked list item styles */
.untracked-list-item {
  gap: 0.75rem;
  padding: 0.5rem 0.75rem;
}

.project-item-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  min-width: 32px;
  box-sizing: border-box;
}

.untracked-list-item .project-item-content {
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  gap: 0.1rem;
}

.project-item-path {
  font-size: 12px;
  color: var(--silver);
  opacity: 0.8;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  width: 100%;
  padding: 0 0.2rem;
  box-sizing: border-box;
}
</style>




