<template>
  <div v-if="!projectStore.getActiveProject" class="general-pane-header">
    <HeaderArea :title="'No project selected'" />
  </div>

  <div v-else class="general-pane-header">
    <HeaderArea v-if="isCustomIcon" :title="projectStore.getActiveProjectName"
      :customIcon="projectStore.activeProject.icon" />
    <HeaderArea v-else :title="projectStore.getActiveProjectName" :emoji="projectStore.activeProject.icon" />
    <ActionButton :icon="getAppIcon('parameters')" v-if="userStore.canDo('update_task')" :showLabel="false"
      v-tooltip="'Edit Project'" :buttonFunction="editProject" />
  </div>



  <div v-if="projectStore.getActiveProject" class="general-pane-root">

    <div class="general-pane-container">

      <div v-if="projectStore.activeProject.preview" class="entity-thumb-container">
        <div class="entity-thumb">
          <img v-if="projectStore.activeProject.preview" class="screenshot-thumb"
            :src="projectStore.activeProject.preview">
          <img v-else class="screenshot-thumb" src="/page-states/no_image.png">
        </div>
      </div>


    <div class="general-pane-content">

      <div class="action-bar">

        <!-- {{  isPinExceeded  }} -->
        <ActionButton v-if="isProjectPinned" :icon="getAppIcon('unpin')" :showLabel="true" :fullWidth="true"
          label="Unpin Project" :buttonFunction="unpinProject" />

        <ActionButton v-else-if="!isPinExceeded" :icon="getAppIcon('pin')" :showLabel="true" :fullWidth="true"
          label="Pin Project" :buttonFunction="pinProject" />

        <!-- Reveal in Explorer -->
        <span class="horizontal-flex">
          <ActionButton :icon="getAppIcon('folder-arrow-up-right')" :showLabel="true" :fullWidth="true" label="Show in Explorer"
            :buttonFunction="revealInExplorer" />
          <ActionButton :icon="getAppIcon('copy')" :showLabel="false" :fullWidth="false" @click="copyProjectPath()"
            v-tooltip="'Copy Path'" />
        </span>

        <!-- Locate Clustta file -->
        <ActionButton v-if="projectStore.getActiveProject.is_downloaded" :icon="getAppIcon('clustta')" :showLabel="true"
          :fullWidth="true" label="Locate Clustta File" :buttonFunction="locateClusttaFile" />

        <!-- Archive -->
        <ActionButton v-if="!projectStore.getActiveProject.is_closed && userStore.userCanCreateProject"
          :icon="getAppIcon('archive')" :showLabel="true" :fullWidth="true" label="Archive Project"
          :buttonFunction="prepCloseProjectPopUpModal" />


        <ActionButton v-else-if="userStore.userCanCreateProject" :icon="getAppIcon('unarchive')" :showLabel="true"
          :fullWidth="true" label="Unarchive Project" :buttonFunction="toggleCloseProject" />

        <!-- Rebuild -->
        <ActionButton v-if="projectStore.getActiveProject.is_downloaded && !projectStore.getActiveProject.is_closed"
          :icon="getAppIcon('jigsaw')" :showLabel="true" :fullWidth="true" label="Rebuild Project"
          :buttonFunction="rebuildAll" />

        <!-- Free space -->
        <ActionButton :icon="getAppIcon('broom')" :showLabel="true" :fullWidth="true" label="Free Up space"
          :buttonFunction="prepFreeUpSpacePopUpModal" />

        <!-- Delete project -->
        <ActionButton v-if="userStore.userCanCreateProject" :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true"
          label="Empty trash" :buttonFunction="prepEmptyTrashPopUpModal" />

      </div>

      <div class="project-stats">

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

</template>

<script setup>
// imports
import { computed, ref, onMounted, onBeforeUnmount, watch } from 'vue';
import { SettingsService, ProjectService, SyncService, TaskService } from "@/../bindings/clustta/services";
import { ClipboardService, FSService } from '@/../bindings/clustta/services/index';
import emitter from '@/lib/mitt';

// services
import { EntityService } from "@/../bindings/clustta/services";

// states/store imports
import { useTrayStates } from '@/stores/TrayStates';
import { useMenu } from '@/stores/menu';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useUserStore } from '@/stores/users';
import { useModalStore } from '@/stores/modals';
import { useEntityStore } from '@/stores/entity';
import { useTaskStore } from '@/stores/task';
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';
import { useProjectStore } from '@/stores/projects';

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
const entityStore = useEntityStore();
const taskStore = useTaskStore();
const projectStore = useProjectStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();



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
  // let entity = entityStore.selectedEntity;
  menu.hideContextMenu();
  notificationStore.cancleFunction = SyncService.CancelSync
  notificationStore.canCancel = true
  await EntityService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, "")
    .then((data) => {
      taskStore.refreshEntityFilesStatus("")
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
  await ClipboardService.WriteText(projectDir);
  menu.hideContextMenu();

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
  assetCount.value = await TaskService.GetTaskCount(project.uri);
}

const getCollectionCount = async() => {
  let project = projectStore.getActiveProject;
  collectionCount.value = await EntityService.GetEntityCount(project.uri);
}

const getProjectData = async () => {
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
  gap: .5rem;
  /* background-color: forestgreen; */
  /* padding-bottom: 1rem; */
}

.action-bar {
  position: relative;
  display: flex;
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
  /* font-weight: 300; */
  display: flex;
  flex-direction: column;
  background-color: var(--steel);
  width: 100%;
  height: min-content;
  min-height: min-content;
  align-items: center;
  padding: .5rem;
  gap: 5px;
  box-sizing: border-box;
  border-radius: var(--small-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  color: var(--white);
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