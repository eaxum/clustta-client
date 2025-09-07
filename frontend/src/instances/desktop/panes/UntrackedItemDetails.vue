<template>
  <div v-if="debugging" class="general-pane-header">
    <HeaderArea :title="utils.capitalizeStr(projectStore.selectedUntrackedItem?.name)" :icon="getAppIcon(itemIcon)" />
    <!-- <ActionButton v-if="userStore.canDo('update_entity')" :icon="getAppIcon('edit')" 
        v-tooltip="'Rename Entity'" :buttonFunction="editEntity" /> -->
  </div>

  <div v-if="debugging" class="general-pane-root">
    <div class="general-pane-container">

      <div v-if="projectStore.selectedUntrackedItem.preview" class="entity-thumb-container">
        <div class="entity-thumb">
          <img v-if="projectStore.selectedUntrackedItem.preview" class="screenshot-thumb"
            :src="projectStore.selectedUntrackedItem.preview">
          <img v-else class="screenshot-thumb" src="/page-states/no_image.png">
        </div>
      </div>

      <div class="pane-parameter-section">
        <div class="action-bar">

          <div class="action-bar-section">
            <ActionButton @click="deleteItem" :icon="getAppIcon('trash')" :label="'Delete item'" />
          </div>
        </div>

        <div class="task-details">

          <div v-if="projectStore.selectedUntrackedItem.type === 'untracked_task'" class="pane-parameter-detail">
            <div class="simple-text-key">
              Extension
            </div>
            <div class="simple-text-value">
              {{ projectStore.selectedUntrackedItem.extension }}
            </div>
          </div>

          <div v-if="projectStore.selectedUntrackedItem.type === 'untracked_task'" class="pane-parameter-detail">
            <div class="simple-text-key">
            Size 
            </div>
            <div class="simple-text-value">
              {{  itemSize }}
            </div>
          </div>

          <div v-if="projectStore.selectedUntrackedItem.type === 'untracked_entity'" class="pane-parameter-detail">
            <div class="simple-text-key">
            Size 
            </div>
            <div class="simple-text-value">
              {{  collectionSize }}
            </div>
        </div>

        </div>

      </div>
    </div>
  </div>
</template>

<script setup>


import { FSService, TaskService, EntityService } from "@/../bindings/clustta/services";

// imports
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// store imports
import { useIconStore } from '@/stores/icons';
import { useUserStore } from '@/stores/users';
import { useModalStore } from '@/stores/modals';
import { useEntityStore } from '@/stores/entity';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useStageStore } from '@/stores/stages';
import { useProjectStore } from '@/stores/projects';
import { useDndStore } from '@/stores/dnd';
import { usePaneStore } from '@/stores/panes';
import { useAssetStore } from '@/stores/assets';

// state imports
import { useTrayStates } from '@/stores/TrayStates';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// states
const trayStates = useTrayStates();

// stores
const iconStore = useIconStore();
const userStore = useUserStore();
const modalStore = useModalStore();
const entityStore = useEntityStore();
const modals = useDesktopModalStore();
const stage = useStageStore();
const projectStore = useProjectStore();
const dndStore = useDndStore();
const panes = usePaneStore();
const assetStore = useAssetStore();

// vars
const debugging = ref(true);

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const parentName = computed(() => {
  const parentId = projectStore.selectedUntrackedItem.parent_id
  const parent = entityStore.getEntities.find((item) => item.id === parentId)
  return parent ? parent.name : 'None'
});

const itemIcon = computed(() => {
  const item = projectStore.selectedUntrackedItem;
  if (item.type === 'untracked_task') {
    return 'file'
  } else if (item.type === 'untracked_entity') {
    return 'folder'
  }
});

const untrackedItem = computed(() => {
  return projectStore.selectedUntrackedItem
});

const deleteItem = () => {
  prepDeleteUntrackedItemPopUpModal();
};

const deleteUntrackedFolder = () => {
  FSService.DeleteFolder(untrackedItem.value.file_path);
  projectStore.removeUntrackedEntity(untrackedItem.value.id);
  panes.setPaneVisibility('projectDetails', true)
  entityStore.selectedEntity = null;
  stage.markedItems = [];
  emitter.emit('refresh-browser')
  modals.disableAllModals();
};

const deleteUntrackedFile = () => {
  FSService.DeleteFile(untrackedItem.value.file_path);
  projectStore.removeUntrackedTask(untrackedItem.value.id);
  panes.setPaneVisibility('projectDetails', true)
  assetStore.selectedTask = null;
  stage.markedItems = [];
  emitter.emit('refresh-browser')
  modals.disableAllModals();
};

const prepDeleteUntrackedItemPopUpModal = () => {
  const untrackedItemType = untrackedItem.value?.type;
  trayStates.popUpModalTitle = "Delete";
  trayStates.popUpModalMessage = "Are you sure you want to delete this item? This will permanently remove this item. Please confirm if you wish to proceed.";
  trayStates.popUpModalIcon = 'trash';
  trayStates.popUpModalFunction = untrackedItemType === 'untracked_entity' ? deleteUntrackedFolder : deleteUntrackedFile;
  modals.setModalVisibility('popUpModal', true);
};

const importItem = () => {
  if (untrackedItem.value.type === 'untracked_task') {
    importTask();
  } else if (untrackedItem.value.type === 'untracked_entity') {
    importFolder();
  }
};

const importFolder = () => {
  const inRoot = untrackedItem.value.entity_path === ""
  const folderPath = untrackedItem.value.file_path;
  let parentId = ""
  let untrackedParents = []
  let parentPaths = utils.getParentPaths(untrackedItem.value.entity_path)
  if (!inRoot) {
    for (let parent of parentPaths) {
      parentId = entityStore.entities.find((item) => item.entity_path === parent)?.id;
      if (parentId !== undefined) {
        break
      }
      untrackedParents.unshift(parent)
    }
  }

  dndStore.untrackedParents = untrackedParents
  dndStore.targetItemId = parentId;
  dndStore.droppedFolders.push(folderPath);
  panes.setPaneVisibility('projectDetails', true);
  modals.setModalVisibility('importItemsModal', true);
};

const importTask = () => {
  const inRoot = untrackedItem.value.entity_path === ""
  const taskPath = untrackedItem.value.file_path;
  let parentId = ""
  let untrackedParents = []
  let parentPaths = utils.getParentPaths(untrackedItem.value.entity_path)
  if (!inRoot) {
    for (let parent of parentPaths) {
      parentId = entityStore.entities.find((item) => item.entity_path === parent)?.id;
      if (parentId !== undefined) {
        break
      }
      untrackedParents.unshift(parent)
    }
  }
  dndStore.untrackedParents = untrackedParents
  dndStore.targetItemId = parentId;
  dndStore.droppedFiles.push(taskPath);
  panes.setPaneVisibility('projectDetails', true);
  modals.setModalVisibility('importItemsModal', true);
};

const itemSize = ref(0);
const collectionSize = ref(0);

const itemType = computed(() => {
  return projectStore.selectedUntrackedItem?.type;
})

const itemPath = computed(() => {
  const path = projectStore.selectedUntrackedItem?.file_path;
  return path.replace(/\\/g, '/')
});

const getItemSize = async() => {
  const size = await FSService.FileStat(itemPath.value);
  itemSize.value = size.formattedSize;
}

const getCollectionSize = async() => {
  const size = await FSService.FolderSize(itemPath.value);
  collectionSize.value = size;
}

const getProjectData = async () => {
  if(itemType.value === 'untracked_task'){
    if (!await FSService.Exists(itemPath.value)){
      itemSize.value = 'Not on disk'
      return
    }
  }
  getItemSize();
  getCollectionSize();
}

watch(() => projectStore.selectedUntrackedItem, () => {
  itemSize.value = 0;
  collectionSize.value = 0;
  getProjectData();
});


// onMounted
onMounted(() => {
  stage.markedTasks = [];
  
  getProjectData();
	emitter.on('get-project-data', getProjectData);
});

onBeforeUnmount(() => {
	emitter.off('get-project-data', getProjectData);
});


</script>
<style scoped>
@import "@/assets/desktop.css";

.compound-input-section {
  /* background-color: royalblue; */
  /* flex: 1; */
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .4rem;
  width: 100%;
  justify-content: space-between;
  justify-content: space-around;
}

.pane-parameter-section {
  flex: 1;
  height: 200px;
}

.task-details{
  overflow: hidden;
  overflow-y: scroll;
  padding-right: .5rem;
}

.task-details::-webkit-scrollbar {
  width: 4px;
}

.task-details::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--light-steel);
}

.task-details::-webkit-scrollbar-track {
  border-radius: 10px;
}
.pane-parameter-detail {
  display: flex;
  font-size: 14px;
  height: max-content;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  height: 30px;
  border-bottom: var(--transparent-line);
}

.simple-text-key {
  white-space: nowrap;
}

.action-bar {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: .6rem;
  width: max-content;
  width: 100%;
  /* justify-content: space-around; */
  height: max-content;
  padding: .2rem;
  /* background-color: black; */
  /* background-color: tomato; */
  align-items: flex-start;
  box-sizing: border-box;
}

.action-bar-section {
  display: flex;
  align-items: center;
  gap: .5rem;
  justify-content: space-between;
  width: 100%;
}
</style>


