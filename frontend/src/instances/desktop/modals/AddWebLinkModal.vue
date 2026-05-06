<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>
    <HeaderArea :title="title" :icon="CiWebPlus" :showSearch="showSearch" />
    <div class="general-container">

      <div class="input-section">
        <div class="compound-input-section">
          <input v-model="assetName" class="input-short" type="text" :placeholder="$t('placeholders.assetName')" v-focus
            @keydown.enter="handleEnterKey" />
        </div>
      </div>

      <div class="input-section">
        <div class="horizontal-flex">
          <input v-model="assetWebLink" class="input-short" type="text" :placeholder="$t('placeholders.webLink')" ref="assetWebLinkInput" @keydown.enter="handleEnterKey"/>
          <span @click="pasteWebLink" class="single-action-button" v-tooltip="$t('modals.pasteLink')"><CiClipboard class="small-icons" :size="20" /></span>
        </div>
        <InputAlert :show="!isValidWeblink(assetWebLink) && assetWebLink !== 'https://'" :message="$t('modals.invalidWebLink')" />
      </div>



      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.create')" :fullWidth="true" @click="createWebLink(false)" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { CiClipboard, CiWebPlus } from '@clustta/icons-vue';
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
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTrayStates } from '@/stores/TrayStates';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stageStore = useStageStore();
const trayStates = useTrayStates();
const { t } = useI18n();

// refs
const isAwaitingResponse = ref(false);
const isResource = ref(false);
const modalContainer = ref(null);
const tags = ref([]);
const assetName = ref('');
const assetWebLink = ref('https://');
const assetWebLinkInput = ref(null);

// constants
const showSearch = false;
const title = computed(() => t('modals.addWebLink'));

// computed
// Returns whether the form values are valid for submission.
const isValueChanged = computed(() => {
  return assetName.value !== '' && isValidWeblink(assetWebLink.value);
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
  let selectedAssetType;
  try {
    selectedAssetType = await ensureWeblinkAssetType();
  } catch (error) {
    isAwaitingResponse.value = false;
    return;
  }
  const collections = stageStore.markedCollections;
  const isNested = commonStore.navigatorMode && !!collectionStore.navigatedCollection;
  if (collections.length <= 1) {
    let collectionId = '';
    if (isNested) {
      collectionId = collectionStore.navigatedCollection.id;
    } else if (collections.length > 0) {
      collectionId = collections[0];
    }
    await AssetService.CreateAsset(
      projectStore.activeProject.uri,
      assetName.value,
      '',
      selectedAssetType.id,
      collectionId,
      isResource.value,
      '',
      '',
      assetWebLink.value,
      true,
      tags.value,
      '',
      comment
    )
      .then(async (data) => {
        notificationStore.addNotification(t('notifications.creatingItem', { name: assetName.value }), '', 'success');
        if (!trayStates.keepModalOpen) {
          closeModal();
        } else {
          assetName.value = '';
          tags.value = [];
        }
        isAwaitingResponse.value = false;
        notificationStore.addNotification(t('notifications.createdItem', { name: assetName.value }), '', 'success');
        if (launch) {
          FSService.LaunchFile(data.file_path);
        }
        emitter.emit('refresh-browser');
      })
      .catch((error) => {
        console.log(error);
        notificationStore.errorNotification(t('notifications.errorCreatingAsset'), error);
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
      notificationStore.errorNotification(t('notifications.errorCreatingWeblink'), error);
      throw error;
    }
  }
  return weblinkType;
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
        assetWebLink.value = link;
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





