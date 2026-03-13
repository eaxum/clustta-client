<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>
    <HeaderArea :title="$t('modals.createCheckpoint')" :icon="'layers-plus'" />
    <div class="general-container">

      <textarea v-model="message" class="desktop-input-long" type="text" :placeholder="$t('placeholders.makeAComment')" v-focus
        @keydown.enter="handleEnterKey" />

      <InputAlert :show="!isValueChanged" :message="validationMessage" />

      <div v-if="!statusMenuDisplayed" class="attachment-area">
        <div class="asset-item-status-container" v-stop-propagation>
          <div class="asset-item-status" @click="toggleDisplayStatusMenu()"
            :style="{ backgroundColor: assetStatus.color }">
            {{ assetStatus.short_name }}
          </div>
        </div>
        <ActionButton :icon="getAppIcon('paperclip')" v-tooltip="$t('modals.attachSnapshot')" v-stop-propagation
          :buttonFunction="selectPreviewFile" />
        <ActionButton :icon="getAppIcon('clipboard')" v-tooltip="$t('modals.pasteSnapshot')" v-stop-propagation
          :buttonFunction="addImageFromClipBoard" />
        <ActionButton v-if="trayStates.screenshot" :icon="getAppIcon('trash')" v-tooltip="$t('modals.deleteSnapshot')"
          v-stop-propagation :buttonFunction="removePreveiw" />
      </div>

      <div v-else class="status-section">
        <div class="asset-item-status-container status-displayed">
          <StatusMenu @statusSelected="closeStatusMenu" />
        </div>
      </div>

      <span v-if="trayStates.screenshot" class="screenshot-preview">
        <img class="screenshot-thumb" :src="trayStates.screenshot">
      </span>

      <div v-if="trayStates.screenshot" class="horizontal-flex">
        <div class="input-label"> {{ $t('modals.useImageAsThumbnail') }}</div>
        <ToggleSwitch :switchValueProp="useImageAsCover" @click="useAsCover()" />
      </div>

      <div v-if="isRemoteProject" class="horizontal-flex">
        <div class="input-label"> {{ $t('modals.syncAfterCheckpoint') }}</div>
        <ToggleSwitch :switchValueProp="syncAfterCheckpointEnabled" @click="toggleSyncAfterCheckpoint()" />
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.create')" :fullWidth="true" @click="createCheckPoint" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>

    </div>


  </div>
</template>

<script setup>
// imports
import { computed, onMounted, onUnmounted, ref, watchEffect } from 'vue';
import { v4 as uuidv4 } from 'uuid';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';
import StatusMenu from '@/instances/desktop/menus/StatusMenu.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// services
import { CheckpointService, ClipboardService, DialogService, SettingsService, SyncService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStatusStore } from '@/stores/status';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

const assetStore = useAssetStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const statusStore = useStatusStore();
const trayStates = useTrayStates();
const userStore = useUserStore();
const { t } = useI18n();

// refs
const displayStatusMenu = ref(false);
const isAwaitingResponse = ref(false);
const message = ref('');
const modalContainer = ref(null);
const syncAfterCheckpointEnabled = ref(false);
const useImageAsCover = ref(true);

// constants
const forbiddenComments = ['wip', 'wfa', 'retake', 'retook', 'todo', 'fmf'];

// computed
// Returns whether the message is valid for submission.
const isValueChanged = computed(() => {
  const messageWords = message.value.toLowerCase().split(/\s+/);
  const hasForbiddenWord = forbiddenComments.some(comment =>
    messageWords.includes(comment.toLowerCase())
  );
  return message.value.trim().length > 6 && !hasForbiddenWord;
});

// Returns whether the status menu should be displayed.
const isRemoteProject = computed(() => {
  return projectStore.activeProject?.has_remote;
});

const statusMenuDisplayed = computed(() => {
  return assetStore.selectedAsset.type !== 'untracked_asset' && displayStatusMenu.value;
});

// Returns the current asset status.
const assetStatus = computed(() => {
  if (assetStore.selectedAsset.type === 'untracked_asset') {
    return statusStore.statuses.find((item) => item.name === 'todo');
  }
  return assetStore.selectedAsset.status;
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
// Adds an image from clipboard as preview.
const addImageFromClipBoard = () => {
  ClipboardService.ReadImageBase64()
    .then(async (base64Img) => {
      const imageStr = `data:image/png;base64, ${base64Img}`;
      trayStates.screenshot = imageStr;
      trayStates.previewFullPath = await utils.base64ToFile(base64Img);
    })
    .catch((err) => {
      console.log(err);
    });
};

// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Closes the status menu.
const closeStatusMenu = () => {
  displayStatusMenu.value = false;
};

// Creates a checkpoint for the selected asset.
const createCheckPoint = async () => {
  isAwaitingResponse.value = true;
  const assetPath = assetStore.selectedAsset.asset_path;
  const comment = message.value;
  const previewPath = trayStates.previewFullPath;
  const groupId = uuidv4();
  if (assetStore.selectedAsset.type === 'asset') {
    CheckpointService.AddCheckpoint(projectStore.activeProject.uri, [assetPath], comment, previewPath, groupId, useImageAsCover.value)
      .then(() => {
        emitter.emit('refresh-browser');
        emitter.emit('update-checkpoints');
        assetStore.modifiedAssetsPath = assetStore.modifiedAssetsPath.filter((modifiedAssetPath) => modifiedAssetPath !== assetPath);
        assetStore.selectedAsset.file_status = 'normal';
        projectStore.refreshProjects();
        isAwaitingResponse.value = false;
        closeModal();
        if (syncAfterCheckpointEnabled.value) {
          syncAsset();
        }
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        notificationStore.errorNotification(t('notifications.errorCreatingCheckpoint'), error);
      });
  } else {
    await CheckpointService.AddUntrackedAsset(projectStore.activeProject.uri, projectStore.activeProject.working_directory, [assetPath], 0, 1, comment, previewPath, groupId)
      .then(() => {
        assetStore.untrackedAssetsPath = assetStore.untrackedAssetsPath.filter((path) => path !== assetPath);
        emitter.emit('refresh-browser');
        emitter.emit('update-checkpoints');
        projectStore.refreshProjects();
        isAwaitingResponse.value = false;
        closeModal();
        if (syncAfterCheckpointEnabled.value) {
          syncAsset();
        }
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        notificationStore.errorNotification(t('notifications.errorCreatingCheckpoint'), error);
      });
  }
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles enter key press to submit form.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    createCheckPoint();
  }
};

// Removes the current preview image.
const removePreveiw = () => {
  trayStates.screenshot = '';
  trayStates.previewFile = '';
  trayStates.previewFullPath = '';
};

// Syncs the current asset to the remote server.
const syncAsset = () => {
  const assetId = assetStore.selectedAsset.id;
  if (!assetId || !isRemoteProject.value) return;
  SyncService.SyncAsset(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, assetId)
    .then(() => {
      notificationStore.addNotification(t('common.sync'), t('notifications.assetSyncedSuccessfully'), 'success');
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification(t('notifications.errorSyncingAsset'), error);
    });
};

// Toggles the sync after checkpoint toggle.
const toggleSyncAfterCheckpoint = () => {
  syncAfterCheckpointEnabled.value = !syncAfterCheckpointEnabled.value;
};

// Opens a dialog to select a preview file.
const selectPreviewFile = async () => {
  if (!trayStates.userPin) {
    await trayStates.togglePin();
  }
  const result = await DialogService.SelectFileDialog('Select Image File', '*.png; *.jpg; *.jpeg; *.gif; *.bmp; *.tiff; *.webp');
  if (result) {
    const filePath = result.replace(/\\/g, '/');
    const fileName = filePath.split('/').pop();
    const base64Image = await utils.base64FromFile(filePath);
    trayStates.previewFile = fileName;
    trayStates.previewFullPath = filePath;
    trayStates.screenshot = base64Image;
  }
  if (!trayStates.userPin) {
    await trayStates.togglePin();
  }
};

// Toggles the status menu visibility.
const toggleDisplayStatusMenu = () => {
  if (!userStore.canDo('change_status')) {
    return;
  }
  assetStore.isAssetAssetStatus = true;
  displayStatusMenu.value = true;
};

// Toggles whether to use the image as cover.
const useAsCover = () => {
  useImageAsCover.value = !useImageAsCover.value;
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    modalContainer.value.addEventListener('click', closeStatusMenu);
  }
});

// lifecycle hooks
onMounted(async () => {
  trayStates.screenshot = null;
  trayStates.previewFile = '';
  trayStates.previewFullPath = '';
  try {
    syncAfterCheckpointEnabled.value = await SettingsService.GetSyncAfterCheckpoint();
  } catch (error) {
    console.log(error);
  }
});

onUnmounted(() => {
  if (modalContainer.value) {
    modalContainer.value.removeEventListener('click', closeStatusMenu);
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.attachment-area {
  display: flex;
  align-items: center;
  gap: .5rem;
  width: 100%;
  padding: .1rem .3rem;
  /* background-color: deepskyblue; */
  box-sizing: border-box;
}

.status-section {
  display: flex;
  /* background-color: forestgreen; */
  width: 100%;
}

.asset-item-status-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  width: 100%;
  padding: .4rem .2rem;
  height: 3.2rem;
  overflow: hidden;
  /* background-color: darkorange; */
  /* flex: 1; */
}

.status-displayed {
  justify-content: center;
  /* background-color: crimson; */

}

.asset-item-status {
  display: flex;
  border-radius: var(--normal-radius);
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  width: 60px;
  padding: .4rem .4rem;
  height: max-content;
  background-color: firebrick;
  font-size: 12px;
  text-transform: uppercase;
  font-weight: 700;
  color: black;
  transition: all 0.2s ease-out;
}

.asset-item-status:hover {
  border-radius: 10px;
  transform: scale(1.03);
}

.general-container {
  gap: .5rem;
  align-items: center;
  justify-content: flex-start;
}

.desktop-input-long {
  margin-top: 0px;
  font-weight: 200;
  color: var(--white);
}

.input-label {
  font-family: Inter, sans-serif;
  font-size: 14px;
  white-space: nowrap;
  flex: 1;
}
</style>



