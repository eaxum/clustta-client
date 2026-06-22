<template>
  <div ref="collectionMenu" class="filter-menu-container">

    <ActionButton :icon="getAppIcon('info')" :showLabel="true" :fullWidth="true"
      :label="$t('modals.projectDetails')" :buttonFunction="showProjectDetails" />

    <ActionButton :icon="getAppIcon('edit')" v-if="studioStore.canManageProject" :showLabel="true" :fullWidth="true" :label="$t('modals.renameProject')"
      :buttonFunction="renameProject" />

    <!-- Create -->
    <ActionButton :icon="getAppIcon('switches')" v-if="studioStore.canManageProject" :showLabel="true" :fullWidth="true"
      :label="$t('menus.editProject')" :buttonFunction="editProject" />

    <!-- {{  isPinExceeded  }} -->
    <ActionButton v-if="(projectStore.getActiveProject?.is_downloaded || platformStore.isWeb) && isProjectPinned" :icon="getAppIcon('unpin')" :showLabel="true" :fullWidth="true"
      :label="$t('menus.unpinProject')" :buttonFunction="unpinProject" />

    <ActionButton v-else-if="!isPinExceeded" :icon="getAppIcon('pin')" :showLabel="true" :fullWidth="true"
      :label="$t('menus.pinProject')" :buttonFunction="pinProject" />

    <span v-if="userStore.canDo('create_collection') && !platformStore.isWeb" class="menu-divider"></span>

    <!-- Reveal in Explorer -->
    <span v-if="!platformStore.isWeb && projectStore.getActiveProject?.is_downloaded" class="horizontal-flex">
      <ActionButton :icon="getAppIcon('folder-arrow-up-right')" :showLabel="true" :fullWidth="true" :label="showLabel"
        :buttonFunction="revealInExplorer" />
      <ActionButton :icon="getAppIcon('copy')" :showLabel="false" :fullWidth="false" @click="copyProjectPath()"
        v-tooltip="$t('common.copyPath')" />
    </span>

    <!-- Locate Clustta file -->
    <ActionButton v-if="!platformStore.isWeb && projectStore.getActiveProject.is_downloaded" :icon="getAppIcon('clustta')" :showLabel="true"
      :fullWidth="true" :label="$t('menus.locateClusttaFile')" :buttonFunction="locateClusttaFile" />

    <!-- Relocate Working Directory -->
    <ActionButton v-if="!platformStore.isWeb && projectStore.getActiveProject?.is_downloaded" :icon="getAppIcon('folder-arrow-in')" :showLabel="true" :fullWidth="true" :label="$t('menus.relocate')"
      :buttonFunction="relocateWorkingDirectory" />

    <span v-if="projectStore.getActiveProject?.is_downloaded || platformStore.isWeb" class="menu-divider"></span>

    <!-- Trim Project - only for remote projects that are synced -->
    <ActionButton v-if="!platformStore.isWeb && projectStore.getActiveProject?.has_remote && !projectStore.getActiveProject?.is_unsynced"
      :icon="getAppIcon('scissors')" :showLabel="true" :fullWidth="true" :label="$t('menus.trimProject')"
      :buttonFunction="prepTrimProjectPopUpModal" />
      
    <!-- Archive -->
    <ActionButton v-if="!projectStore.getActiveProject?.is_closed && studioStore.canManageProject"
      :icon="getAppIcon('archive')" :showLabel="true" :fullWidth="true" :label="$t('menus.archiveProject')"
      :buttonFunction="prepCloseProjectPopUpModal" />


    <ActionButton v-else-if="studioStore.canManageProject" :icon="getAppIcon('unarchive')" :showLabel="true"
      :fullWidth="true" :label="$t('menus.unarchiveProject')" :buttonFunction="toggleCloseProject" />

    <!-- Fetch -->
    <!-- <ActionButton v-if="!platformStore.isWeb && projectStore.getActiveProject?.is_downloaded && !projectStore.getActiveProject?.is_closed"
      :icon="getAppIcon('fetch')" :showLabel="true" :fullWidth="true" :label="$t('menus.fetchProject')"
      :buttonFunction="fetchAll" /> -->

    <!-- Leave project (collaborator removes themselves) -->
    <ActionButton v-if="isPersonalCollaborator" :icon="getAppIcon('logout')" :showLabel="true" :fullWidth="true" :label="$t('menus.leaveProject')"
      :buttonFunction="prepLeaveProjectPopUpModal" />

    <!-- Remove project (local copy only, remote stays) -->
    <ActionButton v-if="!platformStore.isWeb && projectStore.getActiveProject?.has_remote && projectStore.getActiveProject?.is_downloaded" :icon="getAppIcon('minus-circle')" :showLabel="true" :fullWidth="true" :label="$t('menus.removeProject')"
      :buttonFunction="prepRemovePopUpModal" />

    <!-- Delete project -->
    <ActionButton v-if="(projectStore.getActiveProject?.is_downloaded || platformStore.isWeb) && studioStore.canManageProject" :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true" :label="$t('menus.deleteProject')"
      :buttonFunction="prepDeletePopUpModal" />


  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { Clipboard } from '@wailsio/runtime';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// composables
import { useRevealLabel } from '@/composables/useRevealLabel';

// services
import { CollectionService, DialogService, FSService, ProjectService, SettingsService, SyncService } from "@/services";

// stores
import { useAssetStore } from '@/stores/assets';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useEntitlementStore } from '@/stores/entitlements';
import { refreshEntitlements } from '@/lib/sync';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';
import { useStudioStore } from '@/stores/studio';

const { t } = useI18n();
const { showLabel } = useRevealLabel();
const assetStore = useAssetStore();
const entitlementStore = useEntitlementStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const trayStates = useTrayStates();
const userStore = useUserStore();
const studioStore = useStudioStore();

// refs
const collectionMenu = ref(null);

// computed
// Checks if the user is a collaborator (not owner) on a Personal remote project.
const isPersonalCollaborator = computed(() => {
  const project = projectStore.getActiveProject;
  return projectStore.selectedStudio?.name === 'Personal' && project?.has_remote && !studioStore.canManageProject;
});

// Checks if the project is pinned.
const isProjectPinned = computed(() => {
  const projectId = projectStore.getActiveProject.id;
  const pinnedProjects = projectStore.pinnedProjects;
  return pinnedProjects?.includes(projectId);
});

// Checks if the pin limit has been exceeded.
const isPinExceeded = computed(() => {
  return false;
});

// methods
// Copies the project working directory path to clipboard.
const copyProjectPath = async () => {
  let project = projectStore.getActiveProject;
  let projectDir = project.working_directory;
  projectDir = projectDir.replace(/\\/g, '/');
  await Clipboard.SetText(projectDir);
  notificationStore.addNotification('Path copied', '', 'success')
  menu.hideContextMenu();
};

// Deletes the project from the local database.
const deleteProject = async ({ deleteWorkingFiles } = {}) => {
  const project = projectStore.getActiveProject;
  await FSService.DeleteFile(project.uri);

  if (deleteWorkingFiles && project.working_directory) {
    await FSService.DeleteFolder(project.working_directory);
  }

  projectStore.removeProjectFromList(project.uri);
};

// Deletes a remote project from the studio server.
const deleteRemoteProject = async ({ deleteWorkingFiles } = {}) => {
  const project = projectStore.getActiveProject;
  
  // Delete from server first
  await ProjectService.DeleteRemoteProject(
    projectStore.getActiveProjectUrl,
    projectStore.selectedStudio.name
  );
  
  // Then delete local file if it exists
  if (project.uri && project.is_downloaded) {
    await FSService.DeleteFile(project.uri);
  }

  // Delete working files if toggled on
  if (deleteWorkingFiles && project.working_directory) {
    await FSService.DeleteFolder(project.working_directory);
  }
  
  projectStore.removeProjectFromList(project.uri, { force: true });
  refreshEntitlements();
  notificationStore.addNotification(
    t('notifications.projectDeleted'),
    t('notifications.projectDeletedDesc', { name: project.name }),
    'success',
    false
  );
};

// Deletes the project working directory data.
const deleteProjectWorkData = async () => {
  let project = projectStore.getActiveProject;
  await FSService.DeleteFolder(project.working_directory)
    .then(() => {
      projectStore.refreshProjects();
      if (projectStore.activeProject.id == project.id) {
        trayStates.$reset();
      }
    })
    .catch((error) => {
      console.error(error);
    });
  modals.setModalVisibility('popUpModal', false);
};

// Opens the edit project modal.
const editProject = () => {
  modals.setModalVisibility('editProjectModal', true);
  menu.hideContextMenu();
};

// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Opens the Clustta file location in the file explorer.
const locateClusttaFile = () => {
  let project = projectStore.getActiveProject;
  FSService.RevealInExplorer(project.uri);
  menu.hideContextMenu();
};

// Pins the current project.
const pinProject = async () => {
  menu.hideContextMenu();
  const studioName = projectStore.getSelectedStudioName;
  const projectId = projectStore.getActiveProject.id;
  await SettingsService.PinProject(studioName, projectId)
    .then(() => {
      projectStore.pinnedProjects.push(projectId);
    })
    .catch((error) => {
      console.error(error);
    });
};

// Prepares and shows the archive project confirmation modal.
const prepCloseProjectPopUpModal = () => {
  let project = projectStore.getActiveProject;
  trayStates.popUpModalTitle = t('menus.archiveProjectTitle', { name: project.name });
  trayStates.popUpModalMessage = t('confirmations.archiveProject');
  trayStates.popUpModalFunction = toggleCloseProject;
  trayStates.popUpModalIcon = 'archive';
  modals.setModalVisibility('popUpModal', true);
  menu.hideContextMenu();
};

// Prepares and shows the delete project confirmation modal.
const prepDeletePopUpModal = () => {
  let project = projectStore.getActiveProject;
  const isPersonal = projectStore.selectedStudio?.name === 'Personal';

  // Build contextual message
  let message = '';
  if (project.has_remote && studioStore.canManageProject) {
    message = t('confirmations.deleteRemoteProject', { name: project.name });
  } else {
    message = t('confirmations.deleteProjectLocal');
    message += ' ' + (isPersonal
      ? t('confirmations.deleteProjectPersonalSuffix')
      : t('confirmations.deleteProjectTeamSuffix'));
  }

  trayStates.dangerousActionTitle = t('menus.deleteProjectTitle', { name: project.name });
  trayStates.dangerousActionMessage = message;
  trayStates.dangerousActionIcon = 'trash';
  trayStates.dangerousActionConfirmText = project.name;
  trayStates.dangerousActionShowInput = true;
  trayStates.dangerousActionFunction = project.has_remote && studioStore.canManageProject
    ? deleteRemoteProject
    : deleteProject;
  trayStates.dangerousActionShowToggle = true;
  trayStates.dangerousActionToggleLabel = t('modals.confirmDangerousAction.deleteWorkingFiles');
  trayStates.dangerousActionToggleOffHint = t('modals.confirmDangerousAction.deleteWorkingFilesOff');
  trayStates.dangerousActionToggleOnHint = t('modals.confirmDangerousAction.deleteWorkingFilesOn');
  modals.setModalVisibility('confirmDangerousActionModal', true);
  menu.hideContextMenu();
};

// Prepares and shows the remove project confirmation modal.
const prepRemovePopUpModal = () => {
  let project = projectStore.getActiveProject;
  let message = t('confirmations.deleteProjectLocal') + ' ' + t('confirmations.deleteProjectTeamSuffix');

  trayStates.dangerousActionTitle = t('menus.removeProjectTitle', { name: project.name });
  trayStates.dangerousActionMessage = message;
  trayStates.dangerousActionIcon = 'minus-circle';
  trayStates.dangerousActionConfirmText = '';
  trayStates.dangerousActionShowInput = false;
  trayStates.dangerousActionFunction = removeProject;
  trayStates.dangerousActionShowToggle = true;
  trayStates.dangerousActionToggleLabel = t('modals.confirmDangerousAction.deleteWorkingFiles');
  trayStates.dangerousActionToggleOffHint = t('modals.confirmDangerousAction.deleteWorkingFilesOff');
  trayStates.dangerousActionToggleOnHint = t('modals.confirmDangerousAction.deleteWorkingFilesOn');
  modals.setModalVisibility('confirmDangerousActionModal', true);
  menu.hideContextMenu();
};

// Prepares and shows the trim project confirmation modal.
const prepTrimProjectPopUpModal = () => {
  menu.hideContextMenu();
  let project = projectStore.getActiveProject;
  trayStates.popUpModalIcon = 'scissors';
  trayStates.popUpModalTitle = t('menus.trimProjectTitle', { name: project.name });
  trayStates.popUpModalMessage = t('confirmations.trimProject');
  trayStates.popUpModalFunction = trimProject;
  modals.setModalVisibility('popUpModal', true);
};

// Fetches all assets in the project.
const fetchAll = async () => {
  menu.hideContextMenu();
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  await CollectionService.Fetch(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, "")
    .then(() => {
      assetStore.refreshCollectionFilesStatus("");
    })
    .catch((error) => {
      console.error(error);
    });
  menu.hideContextMenu();
};

// Opens the relocate working directory dialog.
const relocateWorkingDirectory = async () => {
  menu.hideContextMenu();
  
  const project = projectStore.getActiveProject;
  const currentWorkingDir = project.working_directory;
  
  try {
    const result = await DialogService.SelectFolderDialog(t('menus.selectNewWorkingDirectory'));
    if (!result) return;
    
    let newWorkingDir = result.replace(/\\/g, '/');
    
    trayStates.popUpModalTitle = t('menus.relocateWorkingDirectory');
    trayStates.popUpModalMessage = t('confirmations.relocateWorkingDirectory', { from: currentWorkingDir, to: newWorkingDir });
    trayStates.popUpModalIcon = 'folder';
    trayStates.popUpModalFunction = async () => {
      try {
        stage.operationActive = true;
        await ProjectService.UpdateWorkingDirectory(
          project.has_remote ? projectStore.getActiveProjectUrl : project.uri,
          projectStore.selectedStudio.name,
          newWorkingDir
        );
        project.working_directory = newWorkingDir;
        await projectStore.refreshProjects();
        notificationStore.addNotification(t('notifications.workingDirUpdated'), `New location: ${newWorkingDir}`, 'success', false);
      } catch (error) {
        notificationStore.errorNotification(t('notifications.errorUpdatingDirectory'), error);
      } finally {
        stage.operationActive = false;
        modals.setModalVisibility('popUpModal', false);
      }
    };
    
    modals.setModalVisibility('popUpModal', true);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorSelectingDirectory'), error);
  }
};

// Leaves a project the user is collaborating on.
const leaveProject = async () => {
  const project = projectStore.getActiveProject;
  trayStates.popUpModalLoading = true;
  try {
    await ProjectService.LeaveProject(project.remote);
    if (project.is_downloaded) {
      await FSService.DeleteFile(project.uri);
    }
    projectStore.removeProjectFromList(project.uri, { force: true });
    notificationStore.addNotification(
      t('notifications.leftProject'),
      t('notifications.leftProjectDesc', { name: project.name }),
      'success',
      false
    );
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification(t('notifications.errorLeavingProject'), error);
  } finally {
    trayStates.popUpModalLoading = false;
  }
  modals.disableAllModals();
};

// Prepares and shows the leave project confirmation modal.
const prepLeaveProjectPopUpModal = () => {
  const project = projectStore.getActiveProject;
  trayStates.popUpModalTitle = t('menus.leaveProjectTitle', { name: project.name });
  trayStates.popUpModalMessage = t('confirmations.leaveProject');
  trayStates.popUpModalIcon = 'logout';
  trayStates.popUpModalFunction = leaveProject;
  modals.setModalVisibility('popUpModal', true);
  menu.hideContextMenu();
};

// Emits event to rename the project.
const renameProject = () => {
  emitter.emit('renameProject');
  menu.hideContextMenu();
};

// Removes the local .clst file only (keeps server copy).
const removeProject = async ({ deleteWorkingFiles } = {}) => {
  const project = projectStore.getActiveProject;
  await FSService.DeleteFile(project.uri);

  if (deleteWorkingFiles && project.working_directory) {
    await FSService.DeleteFolder(project.working_directory);
  }

  projectStore.removeProjectFromList(project.uri);
};

// Reveals the project directory in the file explorer.
const revealInExplorer = async () => {
  let project = projectStore.getActiveProject;
  await FSService.MakeDirs(project.working_directory);
  FSService.RevealInExplorer(project.working_directory);
  menu.hideContextMenu();
};

// Opens the project details modal.
const showProjectDetails = () => {
  modals.setModalVisibility('projectDetailsModal', true);
  menu.hideContextMenu();
};

// Toggles the closed/archived state of the project.
const toggleCloseProject = async () => {
  let projectUri;
  if (projectStore.activeProject.has_remote) {
    projectUri = projectStore.getActiveProjectUrl;
  } else {
    projectUri = projectStore.activeProject.uri;
  }

  await ProjectService.ToggleCloseProject(projectUri, projectStore.selectedStudio.name)
    .then(() => {
      projectStore.activeProject.is_closed = !projectStore.activeProject.is_closed;
    })
    .catch((error) => {
      console.error(error);
      notificationStore.addNotification(t('notifications.errorClosingProject'), error, "error", false);
    });
  modals.setModalVisibility('popUpModal', false);
  menu.hideContextMenu();
};

// Trims the project by clearing cached data and working files.
const trimProject = async () => {
  let project = projectStore.getActiveProject;
  
  try {
    await ProjectService.TrimProject(project.uri);
    await FSService.DeleteFolder(project.working_directory);
    projectStore.refreshProjects();
    
    if (projectStore.activeProject.id == project.id) {
      trayStates.$reset();
    }
    
    notificationStore.addNotification(t('notifications.projectTrimmed'), t('notifications.projectTrimmedDesc'), "success", false);
  } catch (error) {
    console.error(error.message || error);
    notificationStore.addNotification(t('notifications.errorTrimmingProject'), error.message || t('notifications.errorOccurred'), "error", false);
  } finally {
    modals.disableAllModals();
  }
};

// Unpins the current project.
const unpinProject = async () => {
  menu.hideContextMenu();
  const studioName = projectStore.getSelectedStudioName;
  const projectId = projectStore.getActiveProject.id;
  await SettingsService.UnpinProject(studioName, projectId);
  projectStore.pinnedProjects = projectStore.pinnedProjects.filter((item) => item !== projectId);
};

// lifecycle hooks
onMounted(() => {
  menu.assetMenuWidth = collectionMenu.value.getBoundingClientRect().width;
  menu.collectionMenu = collectionMenu.value;
});

onBeforeUnmount(() => {
  menu.assetMenuWidth = collectionMenu.value.getBoundingClientRect().width;
  menu.assetMenuHeight = collectionMenu.value.getBoundingClientRect().height;
});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/menu.css";

.horizontal-flex {
  padding: 0;
}
</style>




