<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>
    <HeaderArea :title="title" :icon="icon" :customIcon="customIcon" :showSearch="showSearch" :showPin="true" />
    <div class="general-container">

      <!-- Asset Creation Context -->
      <template v-if="!displayTypeCreator">
        <div class="input-section">
          <div class="compound-input-section">
            <input v-model="assetName" class="input-short" type="text" :placeholder="$t('placeholders.assetName')" v-focus @keydown.enter="handleEnterKey" />
            <ActionButton :icon="getAppIcon('switches')" :buttonFunction="toggleOptions" :isActive="exposeParams" v-tooltip="'Show Options'" />
          </div>
        </div>

        <div class="input-section drop-down-box-section">
          <div class="horizontal-flex">
            <div class="dropdown-wrapper">
              <DropDownBox :items="assetTypeOptions" :selectedItem="assetType" :onSelect="selectAssetType" :useFilter="false" :placeHolder="$t('placeholders.assetType')" />
            </div>
            <span @click="toggleTypeCreator" class="single-action-button" v-tooltip="$t('modals.addNewAssetType')">
              <img class="small-icons" :src="getAppIcon('plus-circle')">
            </span>
          </div>
        </div>

        <div class="asset-options-container" :class="{ 'asset-options-container-closed': showAssetOptions === true }">
          <div class="input-section">
            <Apps />
          </div>
        </div>

        <div class="pop-up-actions">
          <GeneralButton :label="$t('common.close')" :fullWidth="true" :buttonFunction="closeModal" :isActive="!isAwaitingResponse" :colored="false" />
          <GeneralButton :label="$t('common.create')" :fullWidth="true" @click="createAsset(false)" :isActive="isValueChanged" :loading="isAwaitingResponse" />
        </div>
      </template>

      <!-- Asset Type Creation Context -->
      <template v-else>
        <AssetTypeForm ref="typeFormRef" mode="create" @created="handleTypeCreated" @cancel="toggleTypeCreator" @iconChange="handleTypeIconChange" />
      </template>

    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, onUnmounted, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import Apps from '@/instances/common/components/Apps.vue';
import AssetTypeForm from '@/instances/common/components/AssetTypeForm.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { AssetService, FSService } from '@/services';

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
import { useTemplateStore } from '@/stores/template';
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
const templateStore = useTemplateStore();
const trayStates = useTrayStates();
const { t } = useI18n();

// refs
const displayTypeCreator = ref(false);
const exposeParams = ref(false);
const isAwaitingResponse = ref(false);
const isResource = ref(true);
const modalContainer = ref(null);
const newTypeIcon = ref('generic');
const selectedTemplate = ref('');
const showAssetOptions = ref(true);
const tags = ref([]);
const assetName = ref('');
const assetType = ref('');
const typeFormRef = ref(null);

// constants
const showSearch = false;

// computed
// Returns the custom icon path from tray states (used in asset creation context).
const customIcon = computed(() => {
  if (displayTypeCreator.value) {
    return null;
  }
  return trayStates.popUpModalIcon;
});

// Returns the type icon name (used in type creation context).
const icon = computed(() => {
  if (displayTypeCreator.value) {
    return newTypeIcon.value || 'generic';
  }
  return null;
});

// Returns whether the form is valid for submission.
const isValueChanged = computed(() => {
  return assetName.value !== '' && assetType.value !== '';
});

// Returns the list of asset type names.
const assetTypeNames = computed(() => {
  return assetStore.getAssetTypesNames;
});

// Returns the list of asset type options with icons for the dropdown.
const assetTypeOptions = computed(() => {
  return assetStore.getAssetTypes.map((type) => ({
    id: type.id,
    name: type.name,
    icon: type.icon ? getAppIcon(type.icon) : null,
  }));
});

// Returns the modal title from tray states or type creator title.
const title = computed(() => {
  if (displayTypeCreator.value) {
    return t('modals.addAssetTypeTitle');
  }
  return trayStates.popUpModalTitle;
});

// methods
// Closes the modal.
const closeModal = () => {
  trayStates.searchTags = false;
  modals.setModalVisibility('createAssetModal', false);
};

// Creates a new asset/asset in the project.
const createAsset = async (launch = false, comment = 'Asset created') => {
  isAwaitingResponse.value = true;
  const selectedAssetType = assetStore.assetTypes.find(item => item.name === assetType.value);
  const collections = stageStore.markedCollections;
  const template = templateStore.templates.find(template => template.name === templateStore.selectedTemplateName);
  templateStore.lastUsedTemplate = template.name;
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
      template.id,
      '',
      '',
      false,
      tags.value,
      '',
      comment
    )
      .then(async (data) => {
        const newAsset = data;
        notificationStore.addNotification(t('notifications.creatingItem', { name: assetName.value }), '', 'success');
        if (!trayStates.keepModalOpen) {
          closeModal();
        } else {
          assetName.value = '';
          tags.value = [];
        }
        isAwaitingResponse.value = false;
        stageStore.selectedItem = newAsset;
        assetStore.selectedAsset = newAsset;
        stageStore.firstSelectedItemId = newAsset.id;
        stageStore.markedItems = [newAsset.id];
        notificationStore.addNotification(t('notifications.createdItem', { name: assetName.value }), '', 'success');
        emitter.emit('refresh-browser');
        if (launch) {
          FSService.LaunchFile(newAsset.file_path);
        }
      })
      .catch((error) => {
        console.log(error);
        notificationStore.errorNotification(t('notifications.errorCreatingAsset'), error);
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
    createAsset(false);
  }
};

// Handles successful type creation from the form.
const handleTypeCreated = (response) => {
  assetType.value = response.name;
  displayTypeCreator.value = false;
};

// Handles icon change from the type form.
const handleTypeIconChange = (icon) => {
  newTypeIcon.value = icon;
};

// Scrolls the selected app icon into view.
const scrollAppIntoView = () => {
  const selectedIcon = document.querySelector('.apps-flex-item-selected');
  const appsCenter = trayStates.appsContainer.offsetWidth / 2;
  const iconCenter = selectedIcon.offsetWidth / 2;
  const scrollPosition = selectedIcon.offsetLeft - (appsCenter - iconCenter);
  trayStates.appsContainer.scrollTo({
    left: scrollPosition,
    behavior: 'smooth'
  });
};

// Selects a asset type from the dropdown.
const selectAssetType = (assetTypeName) => {
  assetType.value = assetTypeName;
  const allAssetTypeNames = assetTypeNames.value;
  const currentAssetName = assetName.value.toLowerCase();
  if (allAssetTypeNames.includes(currentAssetName)) {
    assetName.value = utils.capitalizeStr(assetTypeName);
  }
};

// Toggles the visibility of asset options.
const toggleOptions = () => {
  showAssetOptions.value = !showAssetOptions.value;
  exposeParams.value = !exposeParams.value;
  scrollAppIntoView();
};

// Toggles the type creator context.
const toggleTypeCreator = () => {
  displayTypeCreator.value = !displayTypeCreator.value;
  if (!displayTypeCreator.value) {
    newTypeIcon.value = 'generic';
  }
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
  assetName.value = utils.capitalizeStr(assetStore.getAssetTypesNames[0]);
  trayStates.listItemsBoundary = modalContainer.value;
  trayStates.tagSearchQuery = '';
  trayStates.itemTags = [];
  if (templateStore.lastUsedTemplate) {
    selectedTemplate.value = templateStore.lastUsedTemplate;
  } else {
    selectedTemplate.value = templateStore.templates[0].name;
  }
});

onUnmounted(() => {
  stageStore.markedCollections = [];
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

.dropdown-wrapper {
  flex: 1;
  width: 100%;
}

.general-container {
  gap: 20px;
}

.input-short {
  width: 100%;
}

.pop-up-actions {
  padding: 0px;
  margin-top: 0;
}

.asset-options-container {
  position: relative;
  box-sizing: border-box;
  width: 100%;
  height: 60px;
  transition: all .2s ease-in-out;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0;
}

.asset-options-container-closed {
  height: 0px;
  padding: 0;
  margin-bottom: -1.5rem;
}
</style>








