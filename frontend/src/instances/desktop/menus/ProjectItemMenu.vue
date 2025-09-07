<template>
  <div ref="collectionMenu" class="filter-menu-container">

    <ActionButton :icon="getAppIcon('edit')" v-if="userStore.userCanCreateProject" :showLabel="true" :fullWidth="true" label="Rename Project"
      :buttonFunction="renameProject" />

    <!-- Create -->
    <ActionButton :icon="getAppIcon('parameters')" v-if="userStore.userCanCreateProject" :showLabel="true" :fullWidth="true"
      label="Edit Project" :buttonFunction="editProject" />

    <!-- {{  isPinExceeded  }} -->
    <ActionButton v-if="isProjectPinned" :icon="getAppIcon('unpin')" :showLabel="true" :fullWidth="true"
      label="Unpin Project" :buttonFunction="unpinProject" />

    <ActionButton v-else-if="!isPinExceeded" :icon="getAppIcon('pin')" :showLabel="true" :fullWidth="true"
      label="Pin Project" :buttonFunction="pinProject" />

    <span v-if="userStore.canDo('create_entity')" class="menu-divider"></span>

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

    <span class="menu-divider"></span>

    <!-- Archive -->
    <ActionButton v-if="!projectStore.getActiveProject.is_closed && userStore.userCanCreateProject"
      :icon="getAppIcon('archive')" :showLabel="true" :fullWidth="true" label="Archive Project"
      :buttonFunction="prepCloseProjectPopUpModal" />


    <ActionButton v-else-if="userStore.userCanCreateProject" :icon="getAppIcon('unarchive')" :showLabel="true"
      :fullWidth="true" label="Unarchive Project" :buttonFunction="toggleCloseProject" />

    <!-- Rebuild -->
    <ActionButton v-if="projectStore.getActiveProject?.is_downloaded && !projectStore.getActiveProject?.is_closed"
      :icon="getAppIcon('jigsaw')" :showLabel="true" :fullWidth="true" label="Rebuild Project"
      :buttonFunction="rebuildAll" />

    <!-- Delete project -->
    <ActionButton :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true" label="Remove Project"
      :buttonFunction="prepDeletePopUpModal" />


  </div>

</template>

<script setup>
// imports
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import { SettingsService, ProjectService, SyncService } from "@/../bindings/clustta/services";
import { ClipboardService, FSService } from '@/../bindings/clustta/services/index';
import emitter from '@/lib/mitt';

// services
import { EntityService } from "@/../bindings/clustta/services";

// states/store imports
import { useTrayStates } from '@/stores/TrayStates';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useUserStore } from '@/stores/users';
import { useAssetStore } from '@/stores/assets';
import { useIconStore } from '@/stores/icons';
import { useProjectStore } from '@/stores/projects';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import { syncData } from '@/lib/sync';

// states/stores
const trayStates = useTrayStates();
const userStore = useUserStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const assetStore = useAssetStore();
const projectStore = useProjectStore();
const iconStore = useIconStore();

// computed

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

const renameProject = () => {
  emitter.emit('renameProject');
  menu.hideContextMenu();
};

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

const revealInExplorer = async () => {
  let project = projectStore.getActiveProject;
  await FSService.MakeDirs(project.working_directory)
  FSService.RevealInExplorer(project.working_directory)
  menu.hideContextMenu();
};

const locateClusttaFile = () => {
  let project = projectStore.getActiveProject;
  FSService.RevealInExplorer(project.uri)
  menu.hideContextMenu();
};

// methods
const syncDataFunc = async () => {
  menu.hideContextMenu();
  syncData()
};

const deleteProjectWorkData = async () => {
  let project = projectStore.getActiveProject;
  await FSService.DeleteFolder(project.working_directory)
    .then((response) => {
      projectStore.refreshProjects()
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
      assetStore.refreshEntityFilesStatus("")
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

const prepDeletePopUpModal = () => {
  let project = projectStore.getActiveProject;
  trayStates.popUpModalTitle = `Remove \"${project.name}\"`;
  trayStates.popUpModalMessage = "This will irreversibly delete this project from your computer!";
  trayStates.popUpModalFunction = deleteProject;
  trayStates.popUpModalIcon = 'trash';
  modals.setModalVisibility('popUpModal', true);
  menu.hideContextMenu();
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
      console.error(error)
      notificationStore.addNotification(
        "Error closing project",
        error,
        "error",
        false
      )
    });
  modals.setModalVisibility('popUpModal', false);
  menu.hideContextMenu();

};

// onMounted hook
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


.entity-item-menu-container {
  z-index: 10;
  display: flex;
  /* opacity: 0;
  visibility : hidden;
  position: absolute; */
  top: 0;
  left: 0;
  flex-direction: column;
  color: white;
  align-items: center;
  gap: .3rem;
  padding: .6rem;
  box-sizing: border-box;
  width: max-content;
  width: 250px;
  height: max-content;
  border-radius: 16px;
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--light-steel);

}

.entity-item-menu-visible {
  /* display: flex; */
  opacity: 1;
  visibility: visible;
}
</style>


