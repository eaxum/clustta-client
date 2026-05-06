<template>
  <div class="modal-container" v-stop-propagation>
    <HeaderArea :title="$t('modals.createCheckpoints')" :icon="CiPlusStone" />
    
    <div class="general-container">
      <textarea v-model="message" class="desktop-input-long" type="text" :placeholder="$t('placeholders.writeAComment')" v-focus
        @keydown.enter="handleEnterKey" />

      <InputAlert :show="!isValueChanged" :message="validationMessage" />

      
    <div v-if="assetStore.loadingAssetStates" class="horizontal-flex input-alert loading-items-count">
      <ActionButton :isLoading="true" :icon="CiLoading"  
					v-tooltip="$t('modals.loadingCollectionStates')" />

      <div class="refresh-label">
        {{ $t('modals.refreshingModifiedItems') }}
      </div>
    </div>

    <div v-else class="horizontal-flex input-alert modified-items-count" 
      :class="{ 'modified-items-count-expanded' : showCheckpointItems}" 
      @click="toggleShowCheckpointItems()">
      
      {{ $t('modals.itemsModified', { count: currentModifiedDisplayPaths.length + currentUntrackedPaths.length }) }}

      <ActionButton :isInactive="true" :label="showCheckpointItems ? $t('common.hide') : $t('common.show')"
        :icon="resolveIcon(showCheckpointItems ? 'eye-cancel' : 'eye')" />
    </div>


    <div v-if="showCheckpointItems" class="modified-items">

      <div v-for="assetState in currentModifiedDisplayPaths" class="modified-item" :key="assetState.asset_path">
        <ActionButton :icon="CiDotBig" :useAlert="true" :noFilter="true" v-tooltip="$t('modals.modifiedAsset')" />
        <div class="modified-item-name">
          {{ assetState.display_path }}
        </div>
        <span class="single-action-button" @click="removeItem(assetState.asset_path)" v-tooltip="$t('common.remove')">
          <img class="small-icons" src="/icons/close.svg">
        </span>
      </div>

      <div v-for="assetPath in currentUntrackedPaths" class="modified-item">
        <ActionButton :icon="CiDotBig" :useDanger="true" :noFilter="true" v-tooltip="$t('modals.untrackedAsset')" />
        <div class="modified-item-name">
          {{ assetPath }}
        </div>
        <span class="single-action-button" @click="removeItem(assetPath)" v-tooltip="$t('common.remove')">
          <img class="small-icons" src="/icons/close.svg">
        </span>
      </div>
      
    </div>

    <div class="pop-up-actions">
      <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
      <GeneralButton :label="$t('common.confirm')" :fullWidth="true" @click="createCheckPoints" :isActive="isValueChanged"
        :loading="isAwaitingResponse" />
    </div>
    
    </div>


  </div>
</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { v4 as uuidv4 } from 'uuid';
import emitter from '@/lib/mitt';
import { CiDotBig, CiLoading, CiPlusStone } from '@clustta/icons-vue';
import { resolveIcon } from '@/lib/icon-map';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';

// services
import { CheckpointService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTrayStates } from '@/stores/TrayStates';

const { t } = useI18n();
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const trayStates = useTrayStates();

// refs
const isAwaitingResponse = ref(false);
const message = ref('');
const removedPaths = ref([]);
const showCheckpointItems = ref(false);
const useImageAsCover = ref(true);

// constants
const forbiddenComments = ['wip', 'wfa', 'retake', 'retook', 'todo', 'fmf'];

// computed
// Returns modified asset display paths after filtering.
const currentModifiedDisplayPaths = computed(() => {
  let filteredAssets = assetStore.modifiedAssets.modified || [];
  filteredAssets = filteredAssets.filter((assetState) => !removedPaths.value.includes(assetState.asset_path));
  if (trayStates.createMultipleCheckpointsCollectionPath) {
    filteredAssets = filteredAssets.filter((assetState) => assetState.asset_path.startsWith(trayStates.createMultipleCheckpointsCollectionPath));
  }
  if (!trayStates.createMultipleCheckpoints) {
    filteredAssets = filteredAssets.filter((assetState) => stage.markedItems.includes(assetState.asset_id));
  }
  return filteredAssets;
});

// Returns untracked file paths after filtering.
const currentUntrackedPaths = computed(() => {
  let filteredAssets = assetStore.modifiedAssets.untracked || [];
  filteredAssets = filteredAssets.filter((untrackedAssetPath) => !removedPaths.value.includes(untrackedAssetPath));
  if (trayStates.createMultipleCheckpointsCollectionPath) {
    filteredAssets = filteredAssets.filter((untrackedAssetPath) => untrackedAssetPath.startsWith(trayStates.createMultipleCheckpointsCollectionPath));
  }
  if (trayStates.createMultipleCheckpoints) {
    return filteredAssets;
  } else {
    const selectedUntrackedAssets = stage.selectedItems
      .filter(item => item.type === 'untracked_asset')
      .map(item => item.asset_path)
      .filter(path => path && filteredAssets.includes(path));
    return selectedUntrackedAssets;
  }
});

// Returns whether the message is valid for submission.
const isValueChanged = computed(() => {
  const messageWords = message.value.toLowerCase().split(/\s+/);
  const hasForbiddenWord = forbiddenComments.some(comment =>
    messageWords.includes(comment.toLowerCase())
  );
  return !assetStore.loadingAssetStates && message.value.trim().length > 6 && !hasForbiddenWord;
});

// Returns the validation message for the comment field.
const validationMessage = computed(() => {
  if (message.value.trim().length <= 6) {
    return t('notifications.messageTooShort');
  }
  const messageWords = message.value.toLowerCase().split(/\s+/);
  const foundForbidden = forbiddenComments.find(comment =>
    messageWords.includes(comment.toLowerCase())
  );
  if (foundForbidden) {
    return t('notifications.avoidForbiddenWord', { word: foundForbidden.toUpperCase() });
  }
  return '';
});

// methods
// Closes the modal.
const closeModal = () => {
  trayStates.createMultipleCheckpoints = true;
  modals.disableAllModals();
};

// Creates checkpoints for all modified items.
const createCheckPoints = async () => {
  const startTime = performance.now();
  isAwaitingResponse.value = true;
  const comment = message.value;
  const previewPath = '';
  const groupId = uuidv4();
  const assetPathsForCheckpoints = currentModifiedDisplayPaths.value.map(assetState => assetState.asset_path);
  await CheckpointService.AddCheckpoint(projectStore.activeProject.uri, assetPathsForCheckpoints, comment, previewPath, groupId, useImageAsCover.value, false)
    .then(() => {
      assetStore.modifiedAssets.modified = assetStore.modifiedAssets.modified.filter(
        (item) => !assetPathsForCheckpoints.includes(item.asset_path)
      );
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification(t('notifications.failedToCreateCheckpoints'), error);
      isAwaitingResponse.value = false;
    });
  const untracked = currentUntrackedPaths.value;
  try {
    for (let i = 0; i < untracked.length; i += 100) {
      const batch = untracked.slice(i, i + 100);
      await CheckpointService.AddUntrackedAsset(projectStore.activeProject.uri, projectStore.activeProject.working_directory, batch, i, untracked.length, comment, previewPath, groupId);
    }
  } catch (error) {
    isAwaitingResponse.value = false;
    notificationStore.errorNotification(t('notifications.errorCreatingCheckpoint'), error);
  }
  assetStore.modifiedAssets.untracked = assetStore.modifiedAssets.untracked.filter(
    (untrackedAssetPath) => !currentUntrackedPaths.value.includes(untrackedAssetPath)
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

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.resolveIcon(iconName);
};

// Handles enter key press to submit form.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    createCheckPoints();
  }
};

// Removes an item from the checkpoint list.
const removeItem = (itemPath) => {
  removedPaths.value.push(itemPath);
  if (currentModifiedDisplayPaths.value.length + currentUntrackedPaths.value.length < 1) {
    closeModal();
  }
};

// Toggles the visibility of checkpoint items list.
const toggleShowCheckpointItems = () => {
  showCheckpointItems.value = !showCheckpointItems.value;
};

// lifecycle hooks
onMounted(async () => {
  if (trayStates.createMultipleCheckpoints) {
    let collectionId = null;
    let targetPath = null;
    const selectedItem = stage.selectedItem;
    let selectedCollection;
    if (selectedItem?.type?.includes('collection')) {
      selectedCollection = selectedItem;
    } else {
      selectedCollection = collectionStore.navigatedCollection;
    }
    if (selectedCollection) {
      if (selectedCollection.type === 'collection') {
        collectionId = selectedCollection.id;
      } else if (selectedCollection.type === 'untracked_collection') {
        targetPath = selectedCollection.file_path;
      }
    }
    await collectionStore.reloadItemsForCheckpoint(collectionId, targetPath);
  }
  trayStates.screenshot = null;
  trayStates.previewFile = '';
  trayStates.previewFullPath = '';
});

onBeforeUnmount(() => {
  trayStates.createMultipleCheckpoints = true;
  trayStates.createMultipleCheckpointsCollectionPath = '';
});
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
  margin-top: 0px;
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
  /* margin-top: 20px; */
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






