<template>
  <div v-if="projectStore.getActiveProject" class="general-pane-root">

    <div class="general-pane-container">

    <div class="general-pane-content">

      <div class="action-bar">

        <ActionButton v-if="studioStore.canManageProject" :icon="getAppIcon('switches')" :showLabel="true" :fullWidth="true"
          :label="$t('panes.editProject')" :buttonFunction="editProject" v-tooltip="$t('panes.editProjectTooltip')" />

        <ActionButton v-if="isProjectPinned" :icon="getAppIcon('unpin')" :showLabel="true" :fullWidth="true"
          :label="$t('panes.unpinProject')" :buttonFunction="unpinProject" v-tooltip="$t('panes.unpinProjectTooltip')" />

        <ActionButton v-else-if="!isPinExceeded" :icon="getAppIcon('pin')" :showLabel="true" :fullWidth="true"
          :label="$t('panes.pinProject')" :buttonFunction="pinProject" v-tooltip="$t('panes.pinProjectTooltip')"/>

        <span v-if="!platformStore.isWeb" class="menu-divider"></span>

        <!-- Reveal in Explorer -->
        <span v-if="!platformStore.isWeb" class="horizontal-flex">
          <ActionButton :icon="getAppIcon('folder-arrow-up-right')" :showLabel="true" :fullWidth="true" :label="$t('panes.showInExplorer')"
            :buttonFunction="revealInExplorer" v-tooltip="$t('panes.showInExplorerTooltip')" />
          <ActionButton :icon="getAppIcon('copy')" :showLabel="false" :fullWidth="false" @click="copyProjectPath()"
            v-tooltip="$t('common.copyPath')" />
        </span>

        <!-- Locate Clustta file -->
        <ActionButton v-if="!platformStore.isWeb && projectStore.getActiveProject.is_downloaded" :icon="getAppIcon('clustta')" :showLabel="true"
          :fullWidth="true" :label="$t('panes.locateClusttaFile')" :buttonFunction="locateClusttaFile" v-tooltip="$t('panes.locateClusttaFileTooltip')" />

        <!-- Relocate Working Directory -->
        <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('folder-arrow-in')" :showLabel="true" :fullWidth="true" :label="$t('panes.relocate')"
          :buttonFunction="relocateWorkingDirectory" v-tooltip="$t('panes.relocateTooltip')" />

        <!-- Backup Project -->
        <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('floppy-disk')" :showLabel="true" :fullWidth="true" :label="$t('panes.backup')"
          :buttonFunction="backupProject" v-tooltip="$t('panes.backupTooltip')" />

        <span class="menu-divider"></span>

        <!-- Archive -->
        <ActionButton v-if="!projectStore.getActiveProject.is_closed && studioStore.canManageProject"
          :icon="getAppIcon('archive')" :showLabel="true" :fullWidth="true" :label="$t('panes.archiveProject')"
          :buttonFunction="prepCloseProjectPopUpModal" v-tooltip="$t('panes.archiveProjectTooltip')" />

        <ActionButton v-else-if="studioStore.canManageProject" :icon="getAppIcon('unarchive')" :showLabel="true"
          :fullWidth="true" :label="$t('panes.unarchiveProject')" :buttonFunction="toggleCloseProject" v-tooltip="$t('panes.unarchiveProjectTooltip')" />

        <!-- Rebuild -->
        <ActionButton v-if="!platformStore.isWeb && projectStore.getActiveProject.is_downloaded && !projectStore.getActiveProject.is_closed"
          :icon="getAppIcon('jigsaw')" :showLabel="true" :fullWidth="true" :label="$t('panes.rebuildProject')"
          :buttonFunction="rebuildAll" v-tooltip="$t('panes.rebuildProjectTooltip')" />

        <span v-if="!platformStore.isWeb" class="menu-divider"></span>

        <!-- Free space -->
        <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('broom')" :showLabel="true" :fullWidth="true" :label="$t('panes.freeUpSpace')"
          :buttonFunction="prepFreeUpSpacePopUpModal" v-tooltip="$t('panes.freeUpSpaceTooltip')" />

        <!-- Trim Project - only for remote projects that are synced -->
        <ActionButton v-if="!platformStore.isWeb && projectStore.getActiveProject.has_remote && !projectStore.getActiveProject.is_unsynced"
          :icon="getAppIcon('scissors')" :showLabel="true" :fullWidth="true" :label="$t('panes.trimProject')"
          :buttonFunction="prepTrimProjectPopUpModal" v-tooltip="$t('panes.trimProjectTooltip')" />

        <!-- Delete project -->
        <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true"
          :label="$t('panes.emptyTrash')" :buttonFunction="prepEmptyTrashPopUpModal" v-tooltip="$t('panes.emptyTrashTooltip')" />

      </div>

      <div v-if="!projectStore.isProjectStatsExpanded" class="project-stats project-stats-collapsed">

        <ActionButton :icon="getAppIcon('info')" :showLabel="true" :fullWidth="true"
          :label="$t('panes.projectStats')" :buttonFunction="toggleProjectStats" />

      </div>
      <div v-else class="project-stats project-stats-collapsed">

        <ActionButton :icon="getAppIcon('chevron-down')" :showLabel="true" :fullWidth="true"
          :label="$t('panes.projectStats')" :buttonFunction="toggleProjectStats" />

          <div class="project-stats-content">
            <div class="pane-parameter-detail">
              <div class="simple-text-key">
              {{ $t('panes.totalAssets') }}
              </div>
              <div class="simple-text-value">
              {{  assetsOnDiskCount }} / {{  assetCount }}
              </div>
            </div>

            <div class="pane-parameter-detail">
              <div class="simple-text-key">
              {{ $t('panes.totalCollections') }}
              </div>
              <div class="simple-text-value">
              {{  collectionsOnDiskCount }} / {{  collectionCount }}
              </div>
            </div>

            <div class="pane-parameter-detail">
              <div class="simple-text-key">
              {{ $t('panes.filesOnComputer') }}
              </div>
              <div class="simple-text-value">
                {{  projectSize }}
              </div>
            </div>

            <div class="pane-parameter-detail">
              <div class="simple-text-key">
              {{ $t('panes.clusttaFileSize') }}
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
import { useI18n } from 'vue-i18n';
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
import { useStudioStore } from '@/stores/studio';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'

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
const studioStore = useStudioStore();

// i18n
const { t } = useI18n();

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

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
    const result = await DialogService.SelectFolderDialog(t('panes.selectNewWorkingDirectory'));
    
    if (!result) {
      return;
    }
    
    let newWorkingDir = result.replace(/\\/g, '/');
    
    trayStates.popUpModalTitle = t('panes.relocateWorkingDirectory');
    trayStates.popUpModalMessage = t('confirmations.relocateWorkingDirectoryMessage', { from: currentWorkingDir, to: newWorkingDir });
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
          t('notifications.workingDirUpdated'),
          `New location: ${newWorkingDir}`,
          'success',
          false
        );
        
      } catch (error) {
        notificationStore.errorNotification(t('notifications.errorUpdatingWorkingDirectory'), error);
      } finally {
        stage.operationActive = false;
        modals.setModalVisibility('popUpModal', false);
        emitter.emit('refresh-browser');
      }
    };
    
    modals.setModalVisibility('popUpModal', true);
    
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorSelectingDirectory'), error);
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
  // let collection = collectionStore.selectedCollection;
  menu.hideContextMenu();
  notificationStore.cancleFunction = SyncService.CancelSync
  notificationStore.canCancel = true
  await CollectionService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, "")
    .then((data) => {
      assetStore.refreshCollectionFilesStatus("")
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
  trayStates.popUpModalTitle = t('panes.deleteWorkingData', { name: project.name });
  trayStates.popUpModalMessage = t('confirmations.deleteWorkingData');
  trayStates.popUpModalFunction = deleteProjectWorkData;
  trayStates.popUpModalIcon = 'broom';
  modals.setModalVisibility('popUpModal', true);
  menu.hideContextMenu();
};

const prepEmptyTrashPopUpModal = () => {
  menu.hideContextMenu();
	trayStates.popUpModalIcon = 'trash'
	trayStates.popUpModalTitle = t('panes.emptyTrashTitle');
	trayStates.popUpModalMessage = t('confirmations.emptyTrash');
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
				t('notifications.errorSyncingData'),
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
  trayStates.popUpModalTitle = t('panes.trimProjectTitle', { name: project.name });
  trayStates.popUpModalMessage = t('confirmations.trimProjectMessage');
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
      t('notifications.projectTrimmed'),
      t('notifications.projectTrimmedDesc'),
      "success",
      false
    );
  } catch (error) {
    console.error(error.message || error);
    notificationStore.addNotification(
      t('notifications.errorTrimmingProject'),
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
  trayStates.popUpModalTitle = t('panes.archiveProjectTitle', { name: project.name });
  trayStates.popUpModalMessage = t('confirmations.archiveProjectMessage');
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
        t('notifications.errorClosingProject'),
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

watch(() => projectStore.getActiveProject?.uri, () => {
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

.menu-divider {
  display: block;
  width: 100%;
  margin: .2rem 0;
}

.general-pane-root{
  padding-top: .5rem;
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






