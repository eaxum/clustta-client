<template>

  <div ref="modalContainer" class="modal-container" v-esc="closeModal" v-stop-propagation>
    <HeaderArea :title="title" :icon="typeIcon" :customIcon="customIcon" />

    <div class="general-container">

      <!-- Asset Edit Context -->
      <template v-if="!displayTypeCreator">
        <div class="input-section">
          <input v-model="assetName" class="input-short" type="text" :placeholder="$t('placeholders.assetName')" v-focus />
        </div>

        <div v-if="asset.is_link" class="input-section">
          <div class="horizontal-flex">
            <input v-model="assetWebLink" class="input-short" type="text" :placeholder="$t('placeholders.webLink')" ref="assetWebLinkInput" />
            <span @click="pasteWebLink" class="single-action-button" v-tooltip="$t('modals.pasteLink')">
              <img class="small-icons" :src="getAppIcon('clipboard')">
            </span>
          </div>
        </div>

        <div v-if="!asset.is_link" class="input-section drop-down-box-section">
          <div class="horizontal-flex">
            <div class="dropdown-wrapper">
              <DropDownBox :items="assetTypeOptions" :selectedItem="assetType" :onSelect="selectAssetType" />
            </div>
            <span @click="toggleTypeCreator" class="single-action-button" v-tooltip="$t('modals.addNewAssetType')">
              <img class="small-icons" :src="getAppIcon('plus-circle')">
            </span>
          </div>
        </div>

        <div class="pop-up-actions">
          <GeneralButton :label="$t('common.close')" :fullWidth="true" :buttonFunction="closeModal" :isActive="!isAwaitingResponse" :colored="false" />
          <GeneralButton :label="$t('common.confirm')" :fullWidth="true" @click="updateAsset()" :isActive="isValueChanged" :loading="isAwaitingResponse" />
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
import { computed, onMounted, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { isValidWeblink } from '@/lib/pointer';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';

// components
import AssetTypeForm from '@/instances/common/components/AssetTypeForm.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { AssetService, ClipboardService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

const { t } = useI18n();
const assetStore = useAssetStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

// refs
const displayTypeCreator = ref(false);
const isAwaitingResponse = ref(false);
const isResource = ref(false);
const modalContainer = ref(null);
const newTypeIcon = ref('generic');
const oldTags = ref([]);
const oldAssetName = ref('');
const oldAssetWebLink = ref('');
const tags = ref([]);
const assetName = ref('');
const assetType = ref('');
const assetTypeId = ref('');
const assetWebLink = ref('');
const typeFormRef = ref(null);

// computed
// Returns the custom icon path from the asset (used in edit context).
const customIcon = computed(() => {
  if (displayTypeCreator.value) {
    return null;
  }
  return asset.value.icon;
});

// Returns whether form values have changed.
const isValueChanged = computed(() => {
  const currentAsset = assetStore.selectedAsset;
  if (!currentAsset) {
    return false;
  }
  const restrictedEntries = [oldAssetName.value, ''];
  const isNameChanged = !restrictedEntries.includes(assetName.value);
  const isPointerChanged = isValidWeblink(assetWebLink.value) && (assetWebLink.value !== oldAssetWebLink.value) && !!assetWebLink.value.length;
  const isAssetTypeChanged = currentAsset.asset_type_id !== assetTypeId.value;
  const isTagsUpdated = tags.value.length === oldTags.value.length &&
    tags.value.every(tag => oldTags.value.includes(tag));
  return isNameChanged || isAssetTypeChanged || !isTagsUpdated || isPointerChanged;
});

// Returns the currently selected asset.
const asset = computed(() => {
  return assetStore.selectedAsset;
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

// Returns the modal title based on asset type or type creator.
const title = computed(() => {
  if (displayTypeCreator.value) {
    return t('modals.addAssetTypeTitle');
  }
  return asset.value.is_link ? t('modals.editLink') : t('modals.editAsset');
});

// Returns the type icon name (used in type creation context).
const typeIcon = computed(() => {
  if (displayTypeCreator.value) {
    return newTypeIcon.value || 'generic';
  }
  return null;
});

// methods
// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles successful type creation from the form.
const handleTypeCreated = (response) => {
  assetType.value = response.name;
  assetTypeId.value = response.id;
  displayTypeCreator.value = false;
};

// Handles icon change from the type form.
const handleTypeIconChange = (icon) => {
  newTypeIcon.value = icon;
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

// Selects a asset type from the dropdown.
const selectAssetType = (assetTypeName) => {
  const assetTypes = assetStore.getAssetTypes;
  const newAssetType = assetTypes.find((item) => item.name === assetTypeName);
  assetType.value = assetTypeName;
  assetTypeId.value = newAssetType.id;
  const allAssetTypeNames = assetTypeNames.value;
  const currentAssetName = assetName.value.toLowerCase();
  if (allAssetTypeNames.includes(currentAssetName)) {
    assetName.value = utils.capitalizeStr(assetTypeName);
  }
};

// Toggles the type creator context.
const toggleTypeCreator = () => {
  displayTypeCreator.value = !displayTypeCreator.value;
  if (!displayTypeCreator.value) {
    newTypeIcon.value = 'generic';
  }
};

// Updates the asset with the new values.
const updateAsset = async () => {
  isAwaitingResponse.value = true;
  const assetId = assetStore.selectedAsset.id;
  const currentAsset = assetStore.selectedAsset;
  const newAssetTags = tags.value;
  const assetTypes = assetStore.getAssetTypes;
  const newAssetType = assetTypes.find((item) => item.id === assetTypeId.value);
  if (assetName.value === '') {
    notificationStore.addNotification('Asset name cant be empty', 'Asset name cant be empty', 'error');
    return;
  }
  const typeChanged = currentAsset.asset_type_id !== assetTypeId.value;
  const tagsUnchanged = newAssetTags.length === currentAsset.tags.length &&
    newAssetTags.every(tag => currentAsset.tags.includes(tag));
  const onlyTypeChanged = typeChanged && assetName.value === currentAsset.name &&
    assetWebLink.value === currentAsset.pointer && tagsUnchanged;
  const update = onlyTypeChanged
    ? AssetService.ChangeAssetType(projectStore.activeProject.uri, assetId, assetTypeId.value)
    : AssetService.UpdateAsset(projectStore.activeProject.uri, assetId, assetName.value, assetTypeId.value, isResource.value, assetWebLink.value, newAssetTags)
      .then(() => ({ requires_sync: projectStore.activeProject?.has_remote === true }));

  await update
    .then((result) => {
      notificationStore.notifyMetadataUpdate(
        result,
        typeChanged ? t('notifications.assetTypeUpdated') : t('notifications.itemsUpdatedSuccessfully', 1),
        false
      );
      currentAsset.name = assetName.value;
      currentAsset.pointer = assetWebLink.value;
      currentAsset.is_resource = isResource.value;
      currentAsset.tags = newAssetTags;
      currentAsset.asset_type_name = newAssetType.name;
      currentAsset.asset_type_icon = newAssetType.icon;
      currentAsset.asset_type_id = newAssetType.id;
      const updateData = {
        itemId: assetId,
        updates: [
          { property: 'name', value: assetName.value },
          { property: 'pointer', value: assetWebLink.value },
          { property: 'is_resource', value: isResource.value },
          { property: 'tags', value: newAssetTags },
          { property: 'asset_type_name', value: newAssetType.name },
          { property: 'asset_type_icon', value: newAssetType.icon },
          { property: 'asset_type_id', value: newAssetType.id }
        ]
      };
      emitter.emit('update-root-data', updateData);
      isAwaitingResponse.value = false;
    })
    .catch((error) => {
      isAwaitingResponse.value = false;
      console.error('Error:', error);
    });
  closeModal();
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// lifecycle hooks
onMounted(() => {
  trayStates.tagSearchQuery = '';
  const currentAsset = assetStore.selectedAsset;
  assetName.value = currentAsset.name;
  assetWebLink.value = currentAsset.pointer;
  isResource.value = currentAsset.is_resource;
  assetType.value = currentAsset.asset_type_name;
  assetTypeId.value = currentAsset.asset_type_id;
  oldAssetName.value = currentAsset.name;
  oldAssetWebLink.value = currentAsset.pointer;
  tags.value = Array.from(currentAsset.tags);
  oldTags.value = Array.from(currentAsset.tags);
});
</script>


<style scoped>
@import "@/assets/desktop.css";

.dropdown-wrapper {
  flex: 1;
  width: 100%;
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
  color: var(--text);
  font-size: 14px;
  white-space: nowrap;
  flex: 1;
}
</style>




