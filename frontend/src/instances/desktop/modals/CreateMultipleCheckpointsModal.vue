<template>
  <div class="modal-container" v-stop-propagation>
    <HeaderArea :title="'Create Checkpoints'" :icon="getAppIcon('layers-plus')" />
    <div class="general-container">
      <textarea v-model="message" class="desktop-input-long" type="text" placeholder="write a comment..." v-focus
        @keydown.enter="handleEnterKey" />

      <InputAlert :show="!isValueChanged" :message="validationMessage" />

    </div>

    <div v-if="assetStore.loadingAssetStates" class="horizontal-flex input-alert loading-items-count">
      <ActionButton :isLoading="true" :icon="getAppIcon('loading')"  
					v-tooltip="'Loading collection states'" />

      <div class="refresh-label">
        refreshing modified items...
      </div>
    </div>

    <div v-else class="horizontal-flex input-alert modified-items-count" 
      :class="{ 'modified-items-count-expanded' : showCheckpointItems}" 
      @click="toggleShowCheckpointItems()">
      
      {{ currentModifiedDisplayPaths.length + currentUntrackedPaths.length }} item(s) modified

      <ActionButton :isInactive="true" :label="showCheckpointItems ? 'Hide' : 'Show'"
        :icon="getAppIcon(showCheckpointItems ? 'eye-cancel' : 'eye')" />
    </div>


    <div v-if="showCheckpointItems" class="modified-items">

      <div v-for="assetState in currentModifiedDisplayPaths" class="modified-item" :key="assetState.task_path">
        <ActionButton :icon="getAppIcon('dot-big')" :useAlert="true" :noFilter="true" v-tooltip="'Modified Asset'" />
        <div class="modified-item-name">
          {{ assetState.display_path }}
        </div>
        <span class="single-action-button" @click="removeItem(assetState.task_path)" v-tooltip="'Remove'">
          <img class="small-icons" src="/icons/close.svg">
        </span>
      </div>

      <div v-for="taskPath in currentUntrackedPaths" class="modified-item">
        <ActionButton :icon="getAppIcon('dot-big')" :useDanger="true" :noFilter="true" v-tooltip="'Untracked Asset'" />
        <div class="modified-item-name">
          {{ taskPath }}
        </div>
        <span class="single-action-button" @click="removeItem(taskPath)" v-tooltip="'Remove'">
          <img class="small-icons" src="/icons/close.svg">
        </span>
      </div>
      
    </div>

    <div class="pop-up-actions">
      <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
      <GeneralButton :label="'Confirm'" :fullWidth="true" @click="createCheckPoints" :isActive="isValueChanged"
        :loading="isAwaitingResponse" />
    </div>
  </div>
</template>

<script setup>
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';

import { useDesktopModalStore } from '@/stores/desktopModals';
import { useTrayStates } from '@/stores/TrayStates';
import { CheckpointService } from "@/services";
import { onMounted, ref, computed, onBeforeUnmount } from 'vue';
import { useNotificationStore } from '@/stores/notifications';
import { useAssetStore } from '@/stores/assets';
import { useStageStore } from '@/stores/stages';
import { useProjectStore } from '@/stores/projects';
import { useIconStore } from '@/stores/icons';
import { useCollectionStore } from '@/stores/collections';
import { v4 as uuidv4 } from 'uuid';
import emitter from '@/lib/mitt';
import ActionButton from '../components/ActionButton.vue';
import utils from '@/services/utils';

const trayStates = useTrayStates();
const assetStore = useAssetStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const useImageAsCover = ref(true);
const stage = useStageStore();
const projectStore = useProjectStore();
const iconStore = useIconStore();
const collectionStore = useCollectionStore();

const showCheckpointItems = ref(false);

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const toggleShowCheckpointItems = () => {
  showCheckpointItems.value = !showCheckpointItems.value;
};

const message = ref('');
const isAwaitingResponse = ref(false);
const removedPaths = ref([]);

const forbiddenComments = [
  'wip',
  'wfa',
  'retake',
  'retook',
  'todo',
  'fmf'
];

const isValueChanged = computed(() => {
  const messageWords = message.value.toLowerCase().split(/\s+/);
  const hasForbiddenWord = forbiddenComments.some(comment =>
    messageWords.includes(comment.toLowerCase())
  );
  
  return !assetStore.loadingAssetStates && message.value.trim().length > 6 && !hasForbiddenWord;
});

const validationMessage = computed(() => {
  if (message.value.trim().length <= 6) {
    return 'Your message is too short.';
  }
  
  const messageWords = message.value.toLowerCase().split(/\s+/);
  const foundForbidden = forbiddenComments.find(comment =>
    messageWords.includes(comment.toLowerCase())
  );
  
  if (foundForbidden) {
    return `Please avoid using "${foundForbidden.toUpperCase()}" in your message. Be more descriptive.`;
  }
  
  return '';
});


const closeModal = () => {
  trayStates.createMultipleCheckpoints = true;
  modals.disableAllModals();
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    createCheckPoints();
  }
};

const removeItem = (itemPath) => {
  removedPaths.value.push(itemPath);
  if (currentModifiedDisplayPaths.value.length + currentUntrackedPaths.value.length < 1) {
    closeModal()
  }
}

const currentModifiedDisplayPaths = computed(() => {
  // Get modified assets from new structure
  let filteredAssets = assetStore.modifiedAssets.modified || [];
  
  // Filter out removed items
  filteredAssets = filteredAssets.filter((assetState) => !removedPaths.value.includes(assetState.task_path));
  
  // Apply entity path filter if set
  if (trayStates.createMultipleCheckpointsEntityPath) {
    filteredAssets = filteredAssets.filter((assetState) => assetState.task_path.startsWith(trayStates.createMultipleCheckpointsEntityPath));
  }
  
  // If not creating multiple checkpoints, filter by marked items
  if (!trayStates.createMultipleCheckpoints) {
    filteredAssets = filteredAssets.filter((assetState) => stage.markedItems.includes(assetState.task_id));
  }

  return filteredAssets;
});

const currentUntrackedPaths = computed(() => {
  // Get untracked files from new structure
  let filteredTasks = assetStore.modifiedAssets.untracked || [];
  
  // Filter out removed items
  filteredTasks = filteredTasks.filter((untrackedTaskPath) => !removedPaths.value.includes(untrackedTaskPath));

  if (trayStates.createMultipleCheckpointsEntityPath) {
    filteredTasks = filteredTasks.filter((untrackedTaskPath) => untrackedTaskPath.startsWith(trayStates.createMultipleCheckpointsEntityPath));
  }

  if (trayStates.createMultipleCheckpoints) {
    return filteredTasks;
  } else {
    // Use selectedItems to get the paths of selected untracked items
    const selectedUntrackedTasks = stage.selectedItems
      .filter(item => item.type === 'untracked_task')
      .map(item => item.task_path)
      .filter(path => path && filteredTasks.includes(path));
    
    return selectedUntrackedTasks;
  }
});

const createCheckPoints = async () => {
  const startTime = performance.now();
  isAwaitingResponse.value = true;
  let comment = message.value
  let previewPath = '';
  let groupId = uuidv4()
  
  // Extract task paths (without extensions) for checkpoint creation
  const taskPathsForCheckpoints = currentModifiedDisplayPaths.value.map(assetState => assetState.task_path);
  
  await CheckpointService.AddCheckpoint(projectStore.activeProject.uri, taskPathsForCheckpoints, comment, previewPath, groupId, useImageAsCover.value)
    .then((response) => {
      // Remove created checkpoints from modifiedAssets
      assetStore.modifiedAssets.modified = assetStore.modifiedAssets.modified.filter(
        (item) => !taskPathsForCheckpoints.includes(item.task_path)
      );
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification(
        "Failed to Create Checkpoints",
        error
      );
      isAwaitingResponse.value = false;
    });

  let untracked = currentUntrackedPaths.value
  try {
    for (let i = 0; i < untracked.length; i += 100) {
      const batch = untracked.slice(i, i + 100);
      await CheckpointService.AddUntrackedTask(projectStore.activeProject.uri, projectStore.activeProject.working_directory, batch, i, untracked.length, comment, previewPath, groupId)
    }
  } catch (error) {
    isAwaitingResponse.value = false;
    notificationStore.errorNotification("Error Creating Checkpoint", error)
  }
  // Remove created untracked items from modifiedAssets
  assetStore.modifiedAssets.untracked = assetStore.modifiedAssets.untracked.filter(
    (untrackedTaskPath) => !currentUntrackedPaths.value.includes(untrackedTaskPath)
  );


  emitter.emit('refresh-browser');
  isAwaitingResponse.value = false;
  modals.disableAllModals();


  const endTime = performance.now();
  const executionTime = endTime - startTime;
  const minutes = Math.floor(executionTime / 60000);
  const seconds = Math.floor((executionTime % 60000) / 1000);
  console.log(`createCheckPoints completed in: ${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`);
};

onMounted(
  async () => {

    if(trayStates.createMultipleCheckpoints){
      let collectionId = null;
      let targetPath = null;
      
      let selectedItem = stage.selectedItem;
      let selectedCollection;
      if (selectedItem?.type?.includes('entity')){
        selectedCollection = selectedItem;
      } else {
        selectedCollection = collectionStore.navigatedCollection;
      }
      
      if (selectedCollection) {
        if (selectedCollection.type === 'entity') {
          collectionId = selectedCollection.id;
        } else if (selectedCollection.type === 'untracked_entity') {
          targetPath = selectedCollection.file_path;
        }
      }
      
      await collectionStore.reloadItemsForCheckpoint(collectionId, targetPath);
    }
    trayStates.screenshot = null
    trayStates.previewFile = ""
    trayStates.previewFullPath = ""
  }
)

onBeforeUnmount(() => {
  trayStates.createMultipleCheckpoints = true;
  trayStates.createMultipleCheckpointsEntityPath = '';
})

</script>

<style scoped>
@import "@/assets/desktop.css";

.modified-items {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: .3rem;
  font-size: medium;
  width: 100%;
  height: min-content;
  box-sizing: border-box;
  padding: 0rem .5rem;
  border-radius: var(--small-radius);
  max-height: 40vh;
  overflow: hidden;
  overflow-y: scroll
}

.modified-items::-webkit-scrollbar {
  width: 4px;
}

.modified-items::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--light-steel);
}

.modified-items::-webkit-scrollbar-track {
  border-radius: 10px;
}

.modified-item {
  display: flex;
  align-items: center;
  /* gap: .5rem; */
  font-size: medium;
  width: 100%;
  justify-content: space-between;
  height: min-content;
  padding: .1rem .1rem .1rem .1rem;
  background-color: var(--steel);
  border-radius: var(--small-radius);
}

.ignored-folder {
  background-color: rgb(0, 161, 86);
}

.modified-item-name {
  /* font-weight: 250; */
  height: min-content;
  flex: 1 1 auto;
  display: flex;
  font-size: 14px;
  align-items: center;
  text-wrap: nowrap;
  overflow: hidden;
  color: var(--white);
  /* background-color: royalblue; */
}

.modal-container {
  max-width: 500px;
}

.general-container {
  gap: .5rem;
  padding-bottom: 1rem;
}

.desktop-input-long {
  margin-top: 20px;
  font-weight: 200;
  color: var(--white);
}

.modified-items-count {
  padding-left: .5rem;
  color: var(--white);
  /* background-color: forestgreen; */
  font-weight: 200;
  height: min-content;
  overflow: hidden;
  box-sizing: border-box;
  height: 30px;
  border-radius: var(--small-radius);
}

.modified-items-count-expanded {
  margin-bottom: 1rem;
}

[data-theme="dark"] .modified-items-count:hover{
  background-color: #ffffff15;
}

.modified-items-count:hover {
  background-color: rgba(0, 0, 0, 0.11);
}

.loading-items-count {
  padding-left: .5rem;
  color: var(--white);
  justify-content: flex-start;
}

.import-prompt {
  padding: 1rem .5rem;
}

@keyframes loadingRotate {
  from {
      transform: rotate(0deg);
  }
  to {
      transform: rotate(360deg);
  }
}

.single-action-button{
  align-content: center;
  justify-content: center;
}

.desktop-input-long {
  margin-top: 20px;
  font-weight: 200;
  color: var(--white);
}

.loading-children-icon {
  width: 20px;
  height: 20px;
  overflow: hidden;
  padding: 0px;
  animation: loadingRotate .5s linear infinite;
}

.refresh-label{
  font-style: italic;
  font-size: 14px;
  color: var(--white);
  opacity: 0.7;
}
</style>






