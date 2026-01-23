<template>
  <div v-if="!projectStore.getActiveProject" class="general-pane-header">
    <HeaderArea :title="'No project selected'" />
  </div>

  <div v-else class="general-pane-header">
    <HeaderArea v-if="isCustomIcon" :title="projectStore.getActiveProjectName"
      :customIcon="projectStore.activeProject.icon" />
    <HeaderArea v-else :title="projectStore.getActiveProjectName" :emoji="projectStore.activeProject.icon" />
    <ActionButton :icon="getAppIcon('switches')" v-if="userStore.canDo('update_task')" :showLabel="false"
      v-tooltip="'Edit Project'" :buttonFunction="editProject" />
  </div>



  <div v-if="projectStore.getActiveProject" class="general-pane-root">

    <div class="general-pane-container">

    <div class="general-pane-content">

      <div class="action-bar">

        <!-- {{  isPinExceeded  }} -->
        <ActionButton v-if="isProjectPinned" :icon="getAppIcon('unpin')" :showLabel="true" :fullWidth="true"
          label="Unpin Project" :buttonFunction="unpinProject" v-tooltip="'Remove project from pinned list'" />

        <ActionButton v-else-if="!isPinExceeded" :icon="getAppIcon('pin')" :showLabel="true" :fullWidth="true"
          label="Pin Project" :buttonFunction="pinProject" v-tooltip="'Pin project for quick access'"/>

        <!-- Reveal in Explorer -->
        <span v-if="!platformStore.isWeb" class="horizontal-flex">
          <ActionButton :icon="getAppIcon('folder-arrow-up-right')" :showLabel="true" :fullWidth="true" label="Show in Explorer"
            :buttonFunction="revealInExplorer" v-tooltip="'Open project folder in file explorer'" />
          <ActionButton :icon="getAppIcon('copy')" :showLabel="false" :fullWidth="false" @click="copyProjectPath()"
            v-tooltip="'Copy Path'" />
        </span>

        <!-- Locate Clustta file -->
        <ActionButton v-if="!platformStore.isWeb && projectStore.getActiveProject.is_downloaded" :icon="getAppIcon('clustta')" :showLabel="true"
          :fullWidth="true" label="Locate Clustta File" :buttonFunction="locateClusttaFile" v-tooltip="'Show the .clst archive in explorer'" />

        <!-- Relocate Working Directory -->
        <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('folder-arrow-in')" :showLabel="true" :fullWidth="true" label="Relocate"
          :buttonFunction="relocateWorkingDirectory" v-tooltip="'Change the working directory path'" />

        <!-- Backup Project -->
        <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('floppy-disk')" :showLabel="true" :fullWidth="true" label="Backup"
          :buttonFunction="backupProject" v-tooltip="'Create a backup of this project'" />

        <!-- Archive -->
        <ActionButton v-if="!projectStore.getActiveProject.is_closed && userStore.userCanCreateProject"
          :icon="getAppIcon('archive')" :showLabel="true" :fullWidth="true" label="Archive Project"
          :buttonFunction="prepCloseProjectPopUpModal" v-tooltip="'Archive project and free up space'" />


        <ActionButton v-else-if="userStore.userCanCreateProject" :icon="getAppIcon('unarchive')" :showLabel="true"
          :fullWidth="true" label="Unarchive Project" :buttonFunction="toggleCloseProject" v-tooltip="'Restore archived project'" />

        <!-- Rebuild -->
        <ActionButton v-if="!platformStore.isWeb && projectStore.getActiveProject.is_downloaded && !projectStore.getActiveProject.is_closed"
          :icon="getAppIcon('jigsaw')" :showLabel="true" :fullWidth="true" label="Rebuild Project"
          :buttonFunction="rebuildAll" v-tooltip="'Download and restore all project files'" />

        <!-- Free space -->
        <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('broom')" :showLabel="true" :fullWidth="true" label="Free Up space"
          :buttonFunction="prepFreeUpSpacePopUpModal" v-tooltip="'Delete working files to free disk space'" />

        <!-- Trim Project - only for remote projects that are synced -->
        <ActionButton v-if="!platformStore.isWeb && projectStore.getActiveProject.has_remote && !projectStore.getActiveProject.is_unsynced"
          :icon="getAppIcon('scissors')" :showLabel="true" :fullWidth="true" label="Trim Project"
          :buttonFunction="prepTrimProjectPopUpModal" v-tooltip="'Reduce project to contain only metadata'" />

        <!-- Delete project -->
        <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true"
          label="Empty trash" :buttonFunction="prepEmptyTrashPopUpModal" v-tooltip="'Permanently delete all items in trash'" />

      </div>

      <div v-if="!projectStore.isProjectStatsExpanded" class="project-stats project-stats-collapsed">

        <ActionButton :icon="getAppIcon('info')" :showLabel="true" :fullWidth="true"
          label="Project stats" :buttonFunction="toggleProjectStats" />

      </div>
      <div v-else class="project-stats project-stats-collapsed">

        <ActionButton :icon="getAppIcon('chevron-down')" :showLabel="true" :fullWidth="true"
          label="Project stats" :buttonFunction="toggleProjectStats" />

          <div class="project-stats-content">
            <div class="pane-parameter-detail">
              <div class="simple-text-key">
              Total Assets
              </div>
              <div class="simple-text-value">
              {{  assetsOnDiskCount }} / {{  assetCount }}
              </div>
            </div>

            <div class="pane-parameter-detail">
              <div class="simple-text-key">
              Total Collections
              </div>
              <div class="simple-text-value">
              {{  collectionsOnDiskCount }} / {{  collectionCount }}
              </div>
            </div>

            <div class="pane-parameter-detail">
              <div class="simple-text-key">
              Files on your computer 
              </div>
              <div class="simple-text-value">
                {{  projectSize }}
              </div>
            </div>

            <div class="pane-parameter-detail">
              <div class="simple-text-key">
              Clustta file size
              </div>
              <div class="simple-text-value">
              {{  clusttaSize }}
              </div>
            </div>
            </div>
      </div>

      </div>
    </div>

  </div>

</template>

<script setup>
// imports
import { computed, ref, onMounted, onBeforeUnmount, watch } from 'vue';
import { SettingsService, ProjectService, SyncService, AssetService } from "@/services";
import { FSService, DialogService } from '@/services';
import { Clipboard } from '@wailsio/runtime';
import emitter from '@/lib/mitt';

// services
import { CollectionService } from "@/services";

// states/store imports
import { useTrayStates } from '@/stores/TrayStates';
import { useMenu } from '@/stores/menu';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useUserStore } from '@/stores/users';
import { useModalStore } from '@/stores/modals';
import { useCollectionStore } from '@/stores/collections';
import { useAssetStore } from '@/stores/assets';
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';
import { useProjectStore } from '@/stores/projects';
import { usePlatformStore } from '@/stores/platform';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// states/stores
const trayStates = useTrayStates();
const userStore = useUserStore();
const menu = useMenu();
const panes = usePaneStore();
const stage = useStageStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const collectionStore = useCollectionStore();
const assetStore = useAssetStore();
const projectStore = useProjectStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const platformStore = usePlatformStore();



const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const isCustomIcon = computed(() => projectStore.activeProject?.icon?.length > 10);

const isProjectPinned = computed(() => {
  const projectId = projectStore.getActiveProject.id;
  const pinnedProjects = projectStore.pinnedProjects;
  return pinnedProjects?.includes(projectId);
});

const isPinExceeded = computed(() => {
  return false
  const pinnedProjects = projectStore.pinnedProjects;
  return pinnedProjects.length > 10

})

// refs
const collectionMenu = ref(null);

const editProject = () => {
  modals.setModalVisibility('editProjectModal', true);
  menu.hideContextMenu();
};

const pinProject = async () => {
  menu.hideContextMenu();
  const studioName = projectStore.getSelectedStudioName;
  const projectId = projectStore.getActiveProject.id;
  await SettingsService.PinProject(studioName, projectId).then((response) => {
    console.log(response)
    projectStore.pinnedProjects.push(projectId);
  }).catch((error) => {
    console.log(error)
  })


};

const unpinProject = async () => {

  menu.hideContextMenu();
  const studioName = projectStore.getSelectedStudioName;
  const projectId = projectStore.getActiveProject.id;
  await SettingsService.UnpinProject(studioName, projectId);
  projectStore.pinnedProjects = projectStore.pinnedProjects.filter((item) => item !== projectId)

};

const revealInExplorer = () => {
  let project = projectStore.getActiveProject;
  FSService.RevealInExplorer(project.working_directory)
  menu.hideContextMenu();
};

const locateClusttaFile = () => {
  let project = projectStore.getActiveProject;
  FSService.RevealInExplorer(project.uri)
  menu.hideContextMenu();
};

const relocateWorkingDirectory = async () => {
  const project = projectStore.getActiveProject;
  const currentWorkingDir = project.working_directory;
  
  try {
    const result = await DialogService.SelectFolderDialog("Select New Working Directory");
    
    if (!result) {
      return;
    }
    
    let newWorkingDir = result.replace(/\\/g, '/');
    
    trayStates.popUpModalTitle = 'Relocate Working Directory?';
    trayStates.popUpModalMessage = `Change working directory from:\n${currentWorkingDir}\n\nTo:\n${newWorkingDir}\n\nNote: Files will NOT be moved. Only the path will be updated.`;
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
        
        notificationStore.addNotification(
          'Working directory updated',
          `New location: ${newWorkingDir}`,
          'success',
          false
        );
        
      } catch (error) {
        notificationStore.errorNotification('Error updating working directory', error);
      } finally {
        stage.operationActive = false;
        modals.setModalVisibility('popUpModal', false);
        emitter.emit('refresh-browser');
      }
    };
    
    modals.setModalVisibility('popUpModal', true);
    
  } catch (error) {
    notificationStore.errorNotification('Error selecting directory', error);
  }
};

const backupProject = () => {
  modals.setModalVisibility('backUpProjectModal', true);
};

const deleteProjectWorkData = async () => {
  let project = projectStore.getActiveProject;
  await FSService.DeleteFolder(project.working_directory)
    .then((response) => {
      projectStore.refreshProjects();
      getProjectData()
      if (projectStore.activeProject.id == project.id) {
        trayStates.$reset();
      }
    })
    .catch((error) => {
      console.error(error);
    });

  modals.setModalVisibility('popUpModal', false);
};

const deleteProject = async () => {
  await FSService.DeleteFile(projectStore.activeProject.uri).then((data) => {
    projectStore.loadProjects()
    getProjectData()
  }).catch(error => {
    console.log(error)
  })
  modals.setModalVisibility('popUpModal', false);
};

const rebuildAll = async () => {
  // let entity = collectionStore.selectedCollection;
  menu.hideContextMenu();
  notificationStore.cancleFunction = SyncService.CancelSync
  notificationStore.canCancel = true
  await CollectionService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, "")
    .then((data) => {
      assetStore.refreshEntityFilesStatus("")
      getProjectData()
    }).catch(error => {
      console.log(error)
    })
  menu.hideContextMenu();
};

const copyProjectPath = async () => {
  let project = projectStore.getActiveProject;
  let projectDir = project.working_directory;
  projectDir = projectDir.replace(/\\/g, '/');
  FSService.MakeDirs(projectDir);
  await Clipboard.SetText(projectDir);
  menu.hideContextMenu();
};

const toggleProjectStats = () => {
  projectStore.isProjectStatsExpanded = !projectStore.isProjectStatsExpanded;
  if(projectStore.isProjectStatsExpanded) getProjectData()
};

const prepFreeUpSpacePopUpModal = () => {
  let project = projectStore.getActiveProject;
  trayStates.popUpModalTitle = `Delete \"${project.name}\" Working Data? `;
  trayStates.popUpModalMessage = "This will irreversibly delete all unsynced data on this project.";
  trayStates.popUpModalFunction = deleteProjectWorkData;
  trayStates.popUpModalIcon = 'broom';
  modals.setModalVisibility('popUpModal', true);
  menu.hideContextMenu();
};

const prepEmptyTrashPopUpModal = () => {
  menu.hideContextMenu();
	trayStates.popUpModalIcon = 'trash'
	trayStates.popUpModalTitle = "Empty Trash";
	trayStates.popUpModalMessage = "This will irreversibly delete all items in trash. Continue?";
	trayStates.popUpModalFunction = emptyTrash;
	modals.setModalVisibility('popUpModal', true);
};

const emptyTrash = async () => {
	await ProjectService.Purge(projectStore.activeProject.uri)
		.then(() => {
			trayStates.trashables = [];
			modals.disableAllModals();
		}).catch((error) => {
			console.error(error.message)
			notificationStore.addNotification(
				"Error Syncing Data",
				error.message,
				"error",
				false
			)
			modals.disableAllModals();
		})
};

const prepTrimProjectPopUpModal = () => {
  menu.hideContextMenu();
  let project = projectStore.getActiveProject;
  trayStates.popUpModalIcon = 'scissors';
  trayStates.popUpModalTitle = `Trim \"${project.name}\"`;
  trayStates.popUpModalMessage = "This will remove cached file data from the project archive and delete the working directory to reduce disk usage. The data can be re-downloaded from the remote when needed. Continue?";
  trayStates.popUpModalFunction = trimProject;
  modals.setModalVisibility('popUpModal', true);
};

const trimProject = async () => {
  let project = projectStore.getActiveProject;
  
  try {
    // First, trim the project database (clear chunks and previews)
    await ProjectService.TrimProject(project.uri);
    
    // Then, delete the working directory (like "Free Up Space")
    await FSService.DeleteFolder(project.working_directory);
    
    projectStore.refreshProjects();
    getProjectData();
    
    if (projectStore.activeProject.id == project.id) {
      trayStates.$reset();
    }
    
    notificationStore.addNotification(
      "Project Trimmed",
      "Cached data and working files have been cleared.",
      "success",
      false
    );
  } catch (error) {
    console.error(error.message || error);
    notificationStore.addNotification(
      "Error Trimming Project",
      error.message || "An error occurred",
      "error",
      false
    );
  } finally {
    modals.disableAllModals();
  }
};


const prepCloseProjectPopUpModal = () => {
  let project = projectStore.getActiveProject;
  trayStates.popUpModalTitle = `Archive \"${project.name}\"`;
  trayStates.popUpModalMessage = "Archiving this project will also free up space in the working directory. Any untracked items will be lost. Proceed?";
  trayStates.popUpModalFunction = toggleCloseProject;
  trayStates.popUpModalIcon = 'archive';
  modals.setModalVisibility('popUpModal', true);
  menu.hideContextMenu();
};

const toggleCloseProject = async () => {
  let projectUri
  if (projectStore.activeProject.has_remote) {
    projectUri = projectStore.getActiveProjectUrl
  } else {
    projectUri = projectStore.activeProject.uri
  }

  await ProjectService.ToggleCloseProject(projectUri, projectStore.selectedStudio.name)
    .then((result) => {
      console.log(result)
      projectStore.activeProject.is_closed = !projectStore.activeProject.is_closed
    }).catch((error) => {
      console.error(error.message)
      notificationStore.addNotification(
        "Error closing project",
        error.message,
        "error",
        false
      )
    });
  modals.setModalVisibility('popUpModal', false);
  menu.hideContextMenu();

};


const projectSize = ref(0);
const clusttaSize = ref(0);
const assetCount = ref(0);
const assetsOnDiskCount = ref(0);
const collectionCount = ref(0);
const collectionsOnDiskCount = ref(0);

const getProjectSize = async() => {
  let project = projectStore.getActiveProject;
  const size = await FSService.FolderSize(project.working_directory);
  projectSize.value = size;
}

const getItemsCount = async() => {
  let project = projectStore.getActiveProject;
  assetsOnDiskCount.value = await FSService.FileCount(project.working_directory);
  collectionsOnDiskCount.value = await FSService.FolderCount(project.working_directory);
}

const getClusttaSize = async() => {
  let project = projectStore.getActiveProject;
  const size = await FSService.FileStat(project.uri);
  clusttaSize.value = size.formattedSize;
}

const getAssetCount = async() => {
  let project = projectStore.getActiveProject;
  assetCount.value = await AssetService.GetAssetCount(project.uri);
}

const getCollectionCount = async() => {
  let project = projectStore.getActiveProject;
  collectionCount.value = await CollectionService.GetCollectionCount(project.uri);
}

const getProjectData = async () => {
  if(!projectStore.isProjectStatsExpanded) return 
  let project = projectStore.getActiveProject;
  if (!await FSService.Exists(project.uri)) return
  getItemsCount();
  getProjectSize();
  getClusttaSize();
  getAssetCount();
  getCollectionCount();
}

watch(() => projectStore.getActiveProject.uri, () => {
  projectSize.value = 0;
  clusttaSize.value = 0;
  assetCount.value = 0;
  collectionCount.value = 0;
  assetsOnDiskCount.value = 0;
  collectionsOnDiskCount.value = 0;
  getProjectData();
});

// onMounted hook
onMounted( async() => {
  getProjectData();
	emitter.on('get-project-data', getProjectData);
});

onBeforeUnmount(() => {
	emitter.off('get-project-data', getProjectData);
});



</script>
<style scoped>
@import "@/assets/desktop.css";

.horizontal-flex{
  padding: 0;
}

.general-pane-root{
  padding-bottom: 1rem;
  box-sizing: border-box;
  position: relative;
}

.general-pane-container{
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
  position: relative;
}

.general-pane-content{
  position: relative;
  width: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-sizing: border-box;
  justify-content: space-between;
  height: 100%;
  /* min-height: 100%; */
  gap: .5rem;
  /* background-color: forestgreen; */
  /* padding-bottom: 1rem; */
}

.action-bar {
  position: relative;
  /* display: flex; */
  flex-direction: column;
  align-items: center;
  gap: .3rem;
  width: max-content;
  width: 100%;
  overflow: hidden;
  overflow-y: scroll;
  box-sizing: border-box;
  height: 100%;
  padding: .2rem;
  align-items: flex-start;
}

.action-bar::-webkit-scrollbar {
	width: 6px;
}

.action-bar::-webkit-scrollbar-thumb {
	border-radius: 10px;
	background-color: var(--steel);
}

.action-bar::-webkit-scrollbar-track {
	border-radius: 10px;
}

.project-stats{
  font-size: 14px;
  display: flex;
  flex-direction: column;
  background-color: var(--steel);
  width: 100%;
  height: min-content;
  min-height: min-content;
  align-items: center;
  box-sizing: border-box;
  border-radius: var(--small-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  border-radius: var(--large-radius);
}

.project-stats-content{
  font-size: 14px;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: min-content;
  min-height: min-content;
  align-items: center;
  padding: .5rem;
  gap: 5px;
  box-sizing: border-box;
  color: var(--white);
}

.project-stats-collapsed{
  padding: 0px;
}

.pane-parameter-detail {
  display: flex;
  font-size: 14px;
  height: max-content;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  height: 30px;
  height: min-content;
}

.simple-text-key {
  white-space: nowrap;
  /* font-weight: 300; */
  font-size: 13px;
}

.simple-text-value {
  text-overflow: ellipsis;
  font-size: 13px;
}
</style>






