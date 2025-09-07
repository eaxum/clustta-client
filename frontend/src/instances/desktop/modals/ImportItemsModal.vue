<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>


    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="'folder-arrow-in'" :showSearch="false" />
      <!-- <ActionButton :icon="getAppIcon('edit')" :showLabel="false" :isActive="dndStore.importEditMode"
        v-tooltip="'Edit Items'" :buttonFunction="toggleEditItems" /> -->
    </div>

    <div class="general-container general-container-wide">

      <div class="selected-folder">
        <ImportPreview />
      </div>

      <div class="pop-up-actions" ref="popUpActions">
        <GeneralButton :label="'Cancel'" :fullWidth="false" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Confirm'" :fullWidth="false" @click="importItems()" :isActive="storeHasData"
          :loading="isAwaitingResponse" />
      </div>

    </div>

  </div>

</template>

<script setup>
import { useIconStore } from '@/stores/icons';
import { v4 as uuidv4 } from 'uuid'
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};
// imports
import { onMounted, watchEffect, onUnmounted, ref, computed } from 'vue';

// services
import { ImportService } from "@/../bindings/clustta/services";

// state imports
import { useTrayStates } from '@/stores/TrayStates';

// store imports
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useStageStore } from '@/stores/stages';
import { useAssetStore } from '@/stores/assets';
import { useEntityStore } from '@/stores/entity';
import { useStatusStore } from '@/stores/status';
import { useProjectStore } from '@/stores/projects';
import { useImportStore } from '@/stores/import';
import { useDndStore } from '@/stores/dnd';
import { useMenu } from '@/stores/menu';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import ImportPreview from '@/instances/desktop/components/ImportPreview.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// states
const trayStates = useTrayStates();
const stage = useStageStore();
const menu = useMenu();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const statusStore = useStatusStore();
const projectStore = useProjectStore();
const entityStore = useEntityStore();
const assetStore = useAssetStore();
const dndStore = useDndStore();

// vars
let title = 'Import Items';

// refs
const modalContainer = ref(null);
const popUpActions = ref(null);
const isAwaitingResponse = ref(false);

// computed
const storeHasData = computed(() => {
  const rawData = dndStore.previewData
  return Object.values(rawData).some(arr => arr.length > 0);
});

function getPathParent(path) {
  // If there's no slash, it's a root item, return empty string
  if (!path.includes('/')) {
    return "";
  }

  // Find the last slash and return everything before it
  const lastSlashIndex = path.lastIndexOf('/');
  return path.substring(0, lastSlashIndex);
}

// methods
const previewImportItems = async () => {

  isAwaitingResponse.value = true;
  let folders = dndStore.droppedFolders;
  let files = dndStore.droppedFiles;
  let parentId = dndStore.targetItemId;
  let parentPath = dndStore.targetItemPath;
  let entitiesId = {}
  entitiesId[parentPath] = parentId

  if (parentId === undefined) {
    parentId = ""
  }
  let trackedParentData = []
  let untrackedParentData = []
  let workingDir = projectStore.activeProject.working_directory
  await ImportService.ImportFolder(projectStore.activeProject.uri, parentId, folders, files, workingDir, projectStore.activeProject.ignore_list)
    .then((response) => {
      if (dndStore.trackedParents.length + dndStore.untrackedParents.length > 0) {
        let entityTypeId = entityStore.entityTypes.find((item) => item.name === "generic")?.id;
        for (let trackedParent of dndStore.trackedParents) {
          let entityData = entityStore.entities.find((item) => item.entity_path === trackedParent);
          entityData.is_tracked_parent = true
          entityData.is_expanded = true
          trackedParentData.push(entityData)
          entitiesId[trackedParent] = entityData.id
        }
        for (let untrackedParent of dndStore.untrackedParents) {
          let untrackedParentPath = getPathParent(untrackedParent)
          let untrackedParentId = entitiesId[untrackedParentPath]
          let name = untrackedParent.split('/').pop();
          let uid = uuidv4()
          let data = {
            id: uid,
            created_at: "",
            description: "",
            entity_path: "",
            entity_type_icon: "generic",
            entity_type_id: entityTypeId,
            entity_type_name: "generic",
            file_path: workingDir + "/" + untrackedParent,
            is_dependency: false,
            mtime: 0,
            name: name,
            parent_id: untrackedParentId,
            preview: "",
            preview_extension: "",
            preview_id: "",
            synced: false,
            trashed: false,
            is_expanded: true,
          }
          untrackedParentData.push(data)
          entitiesId[untrackedParent] = uid
        }

        response.entities.forEach(entity => {
          let entityParentPath = getPathParent(entity.entity_path)
          let entityParentId = entitiesId[entityParentPath]
          if (entityParentId) {
            entity.parent_id = entityParentId
          }
        });

        response.tasks.forEach(task => {
          let taskParentPath = getPathParent(task.task_path)
          let taskParentId = entitiesId[taskParentPath]
          if (taskParentId) {
            task.entity_id = taskParentId
          }
        });

        response.entities = [...trackedParentData, ...untrackedParentData, ...response.entities]
      }
      dndStore.previewData = response;
      isAwaitingResponse.value = false;
    })
    .catch((error) => {
      isAwaitingResponse.value = false;
      console.log(error)
      notificationStore.errorNotification("Error generating previews", error)
    });
};

const importItems = async (comment = "new file") => {
  let entities = dndStore.previewData.entities.filter(entity => !entity.is_tracked_parent)
  let tasks = dndStore.previewData.tasks
  let success = false;
  let errorMessage
  try {
    for (let i = 0; i < entities.length; i += 100) {
      const batch = entities.slice(i, i + 100);
      await ImportService.CreateEntities(projectStore.activeProject.uri, batch, i, entities.length)
    }
    for (let i = 0; i < tasks.length; i += 100) {
      const batch = tasks.slice(i, i + 100);
      await ImportService.CreateTasks(projectStore.activeProject.uri, batch, i, tasks.length, comment)
    }
    success = true;
  } catch (error) {
    console.error("Error caught:", error);
    errorMessage = error
  }

  if (success) {
    dndStore.previewData = {};
    isAwaitingResponse.value = false;
    refresh();
    closeModal();
  } else {
    isAwaitingResponse.value = false;
    notificationStore.resetProgress()
    notificationStore.errorNotification("Error creating items", errorMessage)
    closeModal()
  }
  // await ImportService.CreateItems(
  //   projectStore.activeProject.uri,
  //   dndStore.previewData.entities,
  //   dndStore.previewData.tasks).then((response) => {
  //     dndStore.previewData = {};
  //     isAwaitingResponse.value = false;
  //     refresh();
  //     closeModal();
  //   }).catch((error) => {
  //     isAwaitingResponse.value = false;
  //     notificationStore.errorNotification("Error creating items", error)
  //   }
  //   );
};

const toggleEditItems = () => {
  dndStore.importEditMode = !dndStore.importEditMode;
  if (!dndStore.importEditMode) {
    dndStore.previewDataActiveItem = null;
    dndStore.previewDataSelectedItems = {};
    stage.markedItems = [];
  }
};

const resetDndValues = () => {
  setTimeout(() => {
    dndStore.resetValues();
  }, 100);
};

const refresh = async () => {
  assetStore.tasksLoaded = false;
  await projectStore.refreshActiveProject()
  await statusStore.reloadStatuses();
  await entityStore.reloadEntities();
  await assetStore.reloadTasks();
  projectStore.getUntrackedItems()
  assetStore.tasksLoaded = true;
};

const handleEnterKey = (event) => {
  importItems();
};

const escape = () => {
  modals.setModalVisibility('importItemsModal', false);
};

const closeModal = () => {
  dndStore.targetItemId = '';
  dndStore.trackedParents = []
  dndStore.untrackedParents = []
  dndStore.droppedFolders = [];
  dndStore.droppedFiles = [];

  modals.setModalVisibility("importItemsModal", false);
  resetDndValues();
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

onMounted(async () => {
  await previewImportItems();
  trayStates.listItemsBoundary = modalContainer.value;
  trayStates.tagSearchQuery = '';
});

onUnmounted(() => {
  dndStore.previewDataSelectedItems = {};
  menu.clickOutsideMask = null;
  dndStore.droppedItems = '';
  stage.markedItems = [];
  stage.firstSelectedItemId = '';
  stage.lastSelectedItemId = '';
  dndStore.importEditMode = false;
})


</script>

<style scoped>
@import "@/assets/desktop.css";

.pop-up-actions {
  /* width: min-content; */
  align-items: center;
  /* background-color: forestgreen; */
  /* justify-content: space-around; */
  box-sizing: border-box;
}

.modal-container {
  justify-content: flex-start;
  align-items: flex-start;
  max-height: 90vh;
}

.general-container-wide {
  display: flex;
  flex-direction: column;
  /* background-color: firebrick; */
  overflow: hidden;
  width: 50vw;
  min-width: 600px !important;
  max-width: 1000px;
  max-height: 80vh;

  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.folder-path {
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
}

.rules-toggle {
  display: flex;
  /* background-color: red; */
  gap: .5rem;
  align-items: center;
  min-width: max-content
}

.selected-folder {
  /* background-color: darkblue; */
  width: 100%;
  padding: .2rem;
  overflow: hidden;
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  padding: 10px 20px;
  box-sizing: border-box;
}

.selected-folder-container {
  display: flex;
  /* background-color: firebrick; */
  width: 100%;
  gap: .2rem;
  overflow: hidden;
  display: flex;
  align-items: center;
}

.regex-instances {
  width: 100%;
  display: flex;
  max-height: 50vh;
  flex-direction: column;
  gap: 10px;
  /* background-color: green; */
  /* padding-right: 5px; */
  overflow: hidden;
  /* overflow-y: scroll; */
  box-sizing: border-box;
}

.regex-instances-scroll {
  box-sizing: border-box;
  width: 100%;
  display: flex;
  height: min-content;
  flex-direction: column;
  gap: 10px;
  /* background-color: green; */
}

.regex-instances::-webkit-scrollbar {
  width: 8px;
}

.regex-instances::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--light-steel);
}

.regex-instances::-webkit-scrollbar-track {
  border-radius: 10px;
}

.attachment-area {
  box-sizing: border-box;
  /* margin-top: 1rem; */
  display: flex;
  align-items: center;
  flex-direction: column;
  /* justify-content: space-between; */
  gap: 1rem;
  /* background-color: darkkhaki; */
  width: 98%;
}

.attachment-list {
  box-sizing: border-box;
  /* margin-top: 1rem; */
  display: flex;
  padding: .5rem;
  align-items: center;
  flex-direction: column;
  /* justify-content: space-between; */
  gap: .2rem;
  /* background-color: rgb(57, 122, 108); */

  background-color: rgba(0, 0, 0, 0.144);
  max-height: 150px;
  overflow: hidden;
  overflow-y: scroll;
  width: 100%;
  border-radius: 10px;
}

.attachment-list::-webkit-scrollbar {
  width: 4px;
}

.attachment-list::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgba(255, 255, 255, 0.295);
}

.attachment-list::-webkit-scrollbar-track {
  border-radius: 10px;
  /* background-color: rgba(0, 0, 0, 0.295); */
}

.attachment {
  /* margin-top: 1rem; */
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  width: 100%;
  border-bottom: 1px solid rgba(255, 255, 255, 0.096);
  height: max-content;
  padding: .2rem;

  /* background-color: greenyellow; */
}

.compound-input-section {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .4rem;
}

.task-options-container {
  position: relative;
  box-sizing: border-box;

  width: 100%;
  height: 0px;
  /* height: 80px; */
  overflow: hidden;
  transition-property: height;
  transition-duration: 0.2s;
  transition-timing-function: ease-in-out;
  transition: opacity .5s ease-in-out;
  /* transition: all .1s ease-in-out; */
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0;
  opacity: 1;
}

.task-options-container-closed {
  transition: all .2s ease-in-out;
  opacity: 0;
  height: 0px;
  padding: 0;
  overflow: hidden;
  /* margin-bottom: -1.5rem; */
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
  color: white;
  font-size: 16px;
  white-space: nowrap;
  flex: 1;

}

.pop-up-prompt {
  gap: 10px;
  /* background-color: bisque; */
  align-items: center;
  /* justify-content: center; */
  max-height: 400px;
}
</style>


