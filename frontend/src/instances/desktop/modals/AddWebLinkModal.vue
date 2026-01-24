<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>
    <HeaderArea :title="title" :icon="getAppIcon('web-plus')" :showSearch="showSearch" />
    <div class="general-container">

      <div class="input-section">
        <div class="compound-input-section">
          <input v-model="taskName" class="input-short" type="text" placeholder="Task Name" v-focus
            @keydown.enter="handleEnterKey" />
        </div>
      </div>

      <div class="input-section">
        <div class="horizontal-flex">
          <input v-model="taskWebLink" class="input-short" type="text" placeholder="Web link" ref="taskWebLinkInput" @keydown.enter="handleEnterKey"/>
          <span @click="pasteWebLink" class="single-action-button" v-tooltip="'Paste link'"><img class="small-icons"
              :src="getAppIcon('clipboard')"></span>
        </div>
        <InputAlert :show="!isValidWeblink(taskWebLink) && taskWebLink !== 'https://'" message="Invalid web link. Must start with 'http://' or 'https://'" />
      </div>



      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Create'" :fullWidth="true" @click="createWebLink(false)" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, watchEffect } from 'vue';
import { isValidWeblink } from '@/lib/pointer';
import emitter from '@/lib/mitt';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';

// services
import { AssetService, ClipboardService, FSService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTrayStates } from '@/stores/TrayStates';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stageStore = useStageStore();
const trayStates = useTrayStates();

// refs
const isAwaitingResponse = ref(false);
const isResource = ref(false);
const modalContainer = ref(null);
const tags = ref([]);
const taskName = ref('');
const taskWebLink = ref('https://');
const taskWebLinkInput = ref(null);

// constants
const showSearch = false;
const title = 'Add web link';

// computed
// Returns whether the form values are valid for submission.
const isValueChanged = computed(() => {
  return taskName.value !== '' && isValidWeblink(taskWebLink.value);
});

// methods
// Closes the modal and resets tag search state.
const closeModal = () => {
  trayStates.searchTags = false;
  modals.setModalVisibility('addWebLinkModal', false);
};

// Creates a weblink asset in the current project.
const createWebLink = async (launch = false, comment = 'Asset created') => {
  isAwaitingResponse.value = true;
  let selectedTaskType;
  try {
    selectedTaskType = await ensureWeblinkAssetType();
  } catch (error) {
    isAwaitingResponse.value = false;
    return;
  }
  const entities = stageStore.markedEntities;
  const isNested = commonStore.navigatorMode && !!collectionStore.navigatedCollection;
  if (entities.length <= 1) {
    let entityId = '';
    if (isNested) {
      entityId = collectionStore.navigatedCollection.id;
    } else if (entities.length > 0) {
      entityId = entities[0];
    }
    await AssetService.CreateAsset(
      projectStore.activeProject.uri,
      taskName.value,
      '',
      selectedTaskType.id,
      entityId,
      isResource.value,
      '',
      '',
      taskWebLink.value,
      true,
      tags.value,
      '',
      comment
    )
      .then(async (data) => {
        notificationStore.addNotification('Creating ' + taskName.value + '...', '', 'success');
        if (!trayStates.keepModalOpen) {
          closeModal();
        } else {
          taskName.value = '';
          tags.value = [];
        }
        isAwaitingResponse.value = false;
        notificationStore.addNotification('Created ' + taskName.value + ' successfully.', '', 'success');
        if (launch) {
          FSService.LaunchFile(data.file_path);
        }
        emitter.emit('refresh-browser');
      })
      .catch((error) => {
        console.log(error);
        notificationStore.errorNotification('Error creating task', error);
      });
  }
};

// Ensures the weblink asset type exists, creating it if necessary.
const ensureWeblinkAssetType = async () => {
  let weblinkType = assetStore.assetTypes.find(item => item.name === 'weblink');
  if (!weblinkType) {
    try {
      weblinkType = await AssetService.CreateAssetType(
        projectStore.activeProject.uri,
        'weblink',
        'website'
      );
      assetStore.assetTypes.push(weblinkType);
    } catch (error) {
      notificationStore.errorNotification('Error creating weblink asset type', error);
      throw error;
    }
  }
  return weblinkType;
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles enter key press to submit form.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    createWebLink(false);
  }
};

// Pastes a web link from clipboard if valid.
const pasteWebLink = async () => {
  ClipboardService.ReadText()
    .then(link => {
      if (isValidWeblink(link)) {
        taskWebLink.value = link;
      }
    })
    .catch(err => {
      console.error('Failed to paste from clipboard:', err);
    });
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// lifecycle hooks
onMounted(() => {
  menu.clickOutsideMask = null;
  trayStates.listItemsBoundary = modalContainer.value;
  trayStates.tagSearchQuery = '';
  trayStates.itemTags = [];
});
</script>


<style scoped>
@import "@/assets/desktop.css";

.compound-input-section {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .4rem;
}

.general-container {
  gap: 1rem;
}

.input-short {
  width: 100%;
}

.pop-up-actions {
  padding: 0px;
  margin-top: 0;
}
</style>





