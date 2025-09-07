<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>

    <HeaderArea :title="title" :icon="'folder-plus'" :showSearch="showSearch" />

    <div class="general-container" :style="{ gap: showTaskOptions ? 10 + 'px' : 20 + 'px' }">

      <div v-if="!isMultiple" class="input-section">
        <div class="compound-input-section">
          <input v-model="entityName" class="input-short" type="text" placeholder="Collection Name" v-focus
            v-return="handleEnterKey" />
        </div>
      </div>

      <BatchGenerator v-else ref="batchGen" @updateData="onUpdateCollections" />

      <div class="input-section">
        <DropDownBox :items="entityStore.getEntityTypesNames" :selectedItem="entityType" :onSelect="selectEntityType" />
      </div>

      <div class="horizontal-flex">
        Generate Multiple Items
        <ToggleSwitch v-tooltip="isMultiple? 'Unmark as library' : 'Mark as a library'" @click="toggleIsMultiple" :switchValueProp="isMultiple" />
      </div>

      <div class="horizontal-flex">
        <ActionButton :isInactive="true" :icon="getAppIcon('bookmark')" :label="'Library'" />
        <ToggleSwitch v-tooltip="isLibrary? 'Unmark as library' : 'Mark as a library'" @click="toggleIsLibrary" :switchValueProp="isLibrary" />
      </div>

      <div class="pop-up-actions" ref="popUpActions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Confirm'" :fullWidth="true" :buttonFunction="createCollections" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>

    </div>
  </div>

</template>

<script setup>
// imports
import { onMounted, watchEffect, ref, computed, onUnmounted } from 'vue';
import emitter from '@/lib/mitt';

// services
import { CheckpointService, TaskService, EntityService } from "@/../bindings/clustta/services";
import { FSService } from '@/../bindings/clustta/services/index';

// state imports
import { useTrayStates } from '@/stores/TrayStates';

// store imports
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useTaskStore } from '@/stores/task';
import { useEntityStore } from '@/stores/entity';
import { useStageStore } from '@/stores/stages';
import { useProjectStore } from '@/stores/projects';
import { useMenu } from '@/stores/menu';
import { getRelativePath } from '@/lib/pathlib';
import { useIconStore } from '@/stores/icons';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import BatchGenerator from '@/instances/desktop/components/BatchGenerator.vue';

// states
const trayStates = useTrayStates();

// stores
const taskStore = useTaskStore();
const iconStore = useIconStore();
const projectStore = useProjectStore();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const entityStore = useEntityStore();
const stage = useStageStore();
const menu = useMenu();

// vars
let showSearch = false;

// refs
const modalContainer = ref(null);
const entityName = ref('');
const showTaskOptions = ref(true);
const popUpActions = ref(null);
const isAwaitingResponse = ref(false);
const entityType = ref(entityStore.getEntityTypesNames[0]);
const isLibrary = ref(false)
const isMultiple = ref(false);
const batchGen = ref(null);

// computed props

const title = computed(() => {
  return isMultiple.value ? 'Create Multiple Collections' : 'Create Collection'
})

const isValueChanged = computed(() => {
  if(isMultiple.value){
    return !batchGen.value?.invalidPattern
  } else {
    return entityName.value !== '';
  }
});

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const toggleIsMultiple = () => {
  isMultiple.value = !isMultiple.value;
}

const handleEnterKey = (event) => {
  createCollections();
};

const escape = () => {
  modals.setModalVisibility('createCollectionModal', false);
};

const closeModal = () => {
  modals.setModalVisibility("createCollectionModal", false);
};

const selectEntityType = (entityTypeName) => {
  entityType.value = entityTypeName;
}

const collections = ref([]);

const onUpdateCollections = (allCollections) => {
  collections.value = allCollections;
  console.log(allCollections);
}

const createCollections = async () => {

  isAwaitingResponse.value = true;
  if (stage.groupItems && !isMultiple.value) {
    await createEntityAndMove();
  } else if (isMultiple.value) {
    await createMultipleEntities();
    const successMessage = collections.value.length + ' collections created';
    console.log('created multiple entities');
    notificationStore.addNotification(successMessage, "", "success");
  } else {
    await createSingleEntity();
  }
  isAwaitingResponse.value = false;
  emitter.emit('refresh-browser');
  closeModal();
};

const createEntityAndMove = async () => {

  const selectedItem = stage.selectedItem;
  const type = stage.selectedItem?.type;
	let project = projectStore.activeProject;

  const allEntities = await EntityService.GetEntities(project.uri);

  let parent;

  if (type === 'task') {
    parent = allEntities.find((entity) => entity.id === selectedItem.entity_id)
  } else if (type === 'entity') {
    parent = allEntities.find((entity) => entity.id === selectedItem.parent_id)
  } else {
    const entityPath = selectedItem.entity_path;
    const parentEntity = allEntities.find((entity) => entity.entity_path === entityPath)
    parent = parentEntity
  }

  console.log(parent)

  let parentId = parent?.id

  isAwaitingResponse.value = true;
  let selectedEntityType = entityStore.entityTypes.find(item => item.name === entityType.value);


  EntityService.CreateEntity(projectStore.activeProject.uri, entityName.value, "", selectedEntityType.id, parentId, "", isLibrary.value)
    .then(async data => {
      const successMessage = entityName.value + ' collection created';
      const newEntity = data;
      entityStore.selectedEntity = newEntity;
      isAwaitingResponse.value = false;
      await moveIntoFolder(newEntity.id);
      closeModal();
      notificationStore.addNotification(successMessage, "", "success");
      if (parentId) {
        if (!parentId in stage.expandedEntities) {
          stage.expandEntity(parent);
        }
      }
      await trayStates.refreshData();
      stage.firstSelectedItemId = newEntity.id;
      stage.markedItems = [newEntity.id];
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      console.log(error)
      notificationStore.errorNotification("Error creating entity", error)
    });

}

const selectedEntityTypeId = computed(() => {
  const selectedEntityType = entityStore.entityTypes.find(item => item.name === entityType.value);
  return selectedEntityType?.id;
})

const createSingleEntity = async () => {

  let parentId ;
  if(stage.selectedItem && stage.selectedItem.type === 'entity'){
    parentId = stage.markedItems[0]
  } else if (entityStore.navigatedEntity) {
    parentId = entityStore.navigatedEntity.id;
  } else {
    parentId = '';
  }

  await EntityService.CreateEntity(projectStore.activeProject.uri, entityName.value, "", selectedEntityTypeId.value, parentId, "", isLibrary.value)
    .then(async data => {
      if(!isMultiple){
      const newEntity = data;
      entityStore.selectedEntity = newEntity;
      const successMessage = entityName.value + ' collection created';
      notificationStore.addNotification(successMessage, "", "success");
      stage.firstSelectedItemId = newEntity.id;
      stage.markedItems = [newEntity.id];
      }
    })
    .catch((error) => {
      console.log(error)
      notificationStore.errorNotification("Error creating entity", error)
    });

}

const createMultipleEntities = async () => {
  const collectionNames = collections.value;
  for (let collectionName of collectionNames) {
    entityName.value = collectionName;
    await createSingleEntity()
  }
};

const itemsToGroup = ref([]);

const allProjectItems = computed(() => {
  const allTasks = taskStore.getTasks;
  const allEntities = entityStore.getEntities;
  const alluntrackedFiles = projectStore.untrackedFiles;
  const alluntrackedFolders = projectStore.untrackedFolders;

  const projectItems = [...allTasks, ...allEntities, ...alluntrackedFiles, ...alluntrackedFolders];
  return projectItems;
});

const moveIntoFolder = async (activeItemId) => {

  const itemIds = itemsToGroup.value;

  const selectedItems = allProjectItems.value?.filter((item) => itemIds.includes(item.id));

  for (const item of selectedItems) {

    if (item.type === 'entity') {
      const entityId = item.id;
      const parentId = activeItemId;
      await changeEntityParent(entityId, parentId);
    } else if (item.type === 'task') {
      const taskId = item.id;
      const entityId = activeItemId;
      await changeTaskEntity(taskId, entityId);
    } else {

      let entity = entityStore.selectedEntity
      // console.log(entity)
      // return
      await FSService.MakeDirs(entity.file_path)
      let newPath = await FSService.JoinPath(entity.file_path, item.name)
      const untrackedPath = newPath.replace(/^\/+|\/+$/g, "").replace(/\\/g, "/");
      const workingDir = projectStore.activeProject.working_directory.replace(/^\/+|\/+$/g, "").replace(/\\/g, "/");
      console.log(item)
      const itemPath = getRelativePath(workingDir, untrackedPath)

      let entityPath = "";
      const itemPathEntities = itemPath.split("/");
      if (itemPathEntities.length > 1) {
        // Take all elements except the last one
        const pathWithoutLast = itemPathEntities.slice(0, -1);
        entityPath = pathWithoutLast.join("/");
      }

      FSService.Rename(item.file_path, newPath)
        .then(() => {
          if (item.type == "untracked_task") {
            let untrackedTaskIndex = projectStore.untrackedFilesIndex[item.id];
            projectStore.untrackedFiles[untrackedTaskIndex].item_path = itemPath;
            projectStore.untrackedFiles[untrackedTaskIndex].file_path = newPath;
            projectStore.untrackedFiles[untrackedTaskIndex].entity_path = entityPath;
          } else if (item.type == "untracked_entity") {
            let untrackedFolderIndex = projectStore.untrackedFoldersIndex[item.id];
            projectStore.untrackedFolders[untrackedFolderIndex].item_path = itemPath;
            projectStore.untrackedFolders[untrackedFolderIndex].file_path = newPath;
            projectStore.untrackedFolders[untrackedFolderIndex].entity_path = entityPath;
          }
        })

    }
  }
  // stage.operationActive = false;
};

const changeEntityParent = async (entityId, parentId) => {

  await EntityService.ChangeEntityParent(projectStore.activeProject.uri, entityId, parentId)
    .then((response) => {
      entityStore.changeEntityParent(entityId, parentId);
      const successMessage = 'Moved successfully.'
      notificationStore.addNotification(successMessage, "", "success")
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification("Error changing entity parent", error)
    });
};

const changeTaskEntity = async (taskId, entityId) => {
  await TaskService.ChangeTaskEntity(projectStore.activeProject.uri, taskId, entityId)
    .then((response) => {
      taskStore.changeTaskEntity(taskId, entityId);
      const successMessage = 'Moved successfully.'
      notificationStore.addNotification(successMessage, "", "success")
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification("Error changing task entity", error)
    });
};

const toggleIsLibrary = () => {
  isLibrary.value = !isLibrary.value;
};
// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

onMounted(() => {

  if (stage.groupItems) {
    itemsToGroup.value = stage.markedItems;
    console.log(itemsToGroup.value)
  };

  trayStates.listItemsBoundary = modalContainer.value;
  trayStates.tagSearchQuery = '';
});

onUnmounted(() => {
  stage.groupItems = false;
  stage.markedEntities = [];
  stage.selectedItem = null;

})


</script>

<style scoped>
@import "@/assets/desktop.css";

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
  color: var(--white);
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