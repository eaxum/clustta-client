<template>

  <div ref="modalContainer" class="modal-container" v-esc="closeModal" v-return="handleEnterKey" v-stop-propagation>

      <HeaderArea :title="title" :icon="getAppIcon(displayTypeCreator ? newTypeIcon : collectionTypeIcon)" :showSearch="false" />

    <div class="general-container">

      <!-- Collection Edit Context -->
      <template v-if="!displayTypeCreator">
        <span @click="addCoverImage" v-if="collectionPreview" class="screenshot-preview">
          <img class="screenshot-thumb" :src="collectionPreview">
        </span>

        <div class="input-section">
          <input v-model="collectionName" class="input-short" type="text" :placeholder="$t('placeholders.collectionName')" v-focus />
        </div>

        <div class="input-section">
          <div class="horizontal-flex">
            <div class="dropdown-wrapper">
              <DropDownBox :items="collectionTypeOptions" :selectedItem="collectionType" :onSelect="changeCollectionType" />
            </div>
            <span @click="toggleTypeCreator" class="single-action-button" v-tooltip="$t('modals.addNewCollectionType')">
              <img class="small-icons" :src="getAppIcon('plus-circle')">
            </span>
          </div>
        </div>

        <div v-if="projectStore.activeProject?.has_remote" class="horizontal-flex is-shared-prompt">
          <ActionButton :isInactive="true" :icon="getAppIcon('shared')" :label="$t('common.shared')" />
          <ToggleSwitch v-tooltip="isShared? $t('panes.unmarkAsShared') : $t('panes.markAsShared')" @click="toggleIsShared" :switchValueProp="isShared" />
        </div>

        <div class="pop-up-actions">
          <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
          <GeneralButton :label="$t('common.update')" :fullWidth="true" @click="updateCollection()" :isActive="isValueChanged" :loading="isAwaitingResponse" />
        </div>
      </template>

      <!-- Collection Type Creation Context -->
      <template v-else>
        <CollectionTypeForm ref="typeFormRef" mode="create" @created="handleTypeCreated" @cancel="toggleTypeCreator" @iconChange="handleTypeIconChange" />
      </template>

    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import CollectionTypeForm from '@/instances/common/components/CollectionTypeForm.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// services
import { CollectionService, DialogService } from '@/services';

// stores
import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

const { t } = useI18n();
const collectionStore = useCollectionStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

// refs
const coverImagePath = ref('');
const displayTypeCreator = ref(false);
const collectionName = ref('');
const collectionPreview = ref(null);
const collectionType = ref('');
const collectionTypeIcon = ref('');
const collectionTypeId = ref('');
const collectionTypeName = ref('');
const isAwaitingResponse = ref(false);
const isShared = ref(null);
const modalContainer = ref(null);
const newTypeIcon = ref('generic');
const oldCollectionName = ref('');
const oldCollectionPreview = ref(null);
const oldCollectionType = ref('');
const oldIsShared = ref(null);
const selectedCollection = ref(null);
const typeFormRef = ref(null);

// constants
const title = computed(() => {
  if (displayTypeCreator.value) {
    return t('modals.addCollectionTypeTitle');
  }
  return t('modals.collectionDetails');
});

// computed
// Returns the currently selected collection.
const collection = computed(() => {
  return collectionStore.selectedCollection;
});

// Returns the list of collection type options with icons for the dropdown.
const collectionTypeOptions = computed(() => {
  return collectionStore.getCollectionTypes.map((type) => ({
    id: type.id,
    name: type.name,
    icon: type.icon ? getAppIcon(type.icon) : null,
  }));
});

// Returns whether the shared flag has changed.
const isSharedChanged = computed(() => {
  return oldIsShared.value !== isShared.value;
});

// Returns whether the collection name has changed.
const isNameChanged = computed(() => {
  const restrictedEntries = [oldCollectionName.value, ''];
  return !restrictedEntries.includes(collectionName.value);
});

// Returns whether the preview image has changed.
const isPreviewChanged = computed(() => {
  return oldCollectionPreview.value !== collectionPreview.value;
});

// Returns whether the collection type has changed.
const isTypeChanged = computed(() => {
  return oldCollectionType.value?.toLowerCase() !== collectionType.value?.toLowerCase();
});

// Returns whether any form values have changed.
const isValueChanged = computed(() => {
  return isTypeChanged.value || isNameChanged.value || isPreviewChanged.value || isSharedChanged.value;
});

// methods
// Opens a dialog to select a cover image.
const addCoverImage = async () => {
  const result = await DialogService.SelectFileDialog('Select Image File', '*.png;*.jpg;*.jpeg;*.gif;*.bmp;*.webp');
  if (result) {
    const filePath = result.replace(/\\/g, '/');
    const base64Image = await utils.base64FromFile(filePath);
    coverImagePath.value = filePath;
    collectionPreview.value = base64Image;
  }
};

// Changes the collection type.
const changeCollectionType = (newCollectionTypeName) => {
  const collectionTypes = collectionStore.getCollectionTypes;
  const newCollectionType = collectionTypes.find((item) => item.name === newCollectionTypeName);
  collectionType.value = newCollectionType.name;
  collectionTypeIcon.value = newCollectionType.icon;
  collectionTypeId.value = newCollectionType.id;
};

// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles enter key press to submit form.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    updateCollection();
  }
};

// Handles successful type creation from the form.
const handleTypeCreated = (response) => {
  collectionType.value = response.name;
  collectionTypeIcon.value = response.icon;
  collectionTypeId.value = response.id;
  displayTypeCreator.value = false;
};

// Handles icon change from the type form.
const handleTypeIconChange = (icon) => {
  newTypeIcon.value = icon;
};

// Removes the current cover image.
const removeCoverImage = () => {
  collectionPreview.value = null;
};

// Reverts to the original cover image.
const revertCoverImage = () => {
  collectionPreview.value = oldCollectionPreview.value;
};

// Toggles the shared flag.
const toggleIsShared = () => {
  isShared.value = !isShared.value;
};

// Toggles the type creator context.
const toggleTypeCreator = () => {
  displayTypeCreator.value = !displayTypeCreator.value;
  if (!displayTypeCreator.value) {
    newTypeIcon.value = 'generic';
  }
};

// Updates the collection with all changed values.
const updateCollection = async () => {
  isAwaitingResponse.value = true;
  if (isTypeChanged.value || isNameChanged.value || isSharedChanged.value) {
    await updateCollectionMeta();
  }
  if (isPreviewChanged.value) {
    await updateCollectionCover();
  }
  await collectionStore.reloadCollections();
  isAwaitingResponse.value = false;
  closeModal();
};

// Updates the collection cover image.
const updateCollectionCover = async () => {
  const collectionId = collectionStore.selectedCollection.id;
  const currentCollection = collectionStore.findCollection(collectionId);
  const filePath = coverImagePath.value;
  await CollectionService.UpdatePreview(projectStore.activeProject.uri, collectionId, filePath)
    .then(() => {
      currentCollection.preview = collectionPreview.value;
    })
    .catch((error) => {
      console.error(error);
      isAwaitingResponse.value = false;
      notificationStore.addNotification('Error Updating Image', error, 'error', false);
    });
};

// Updates the collection metadata (name, type, shared flag).
const updateCollectionMeta = async () => {
  const collectionId = collectionStore.selectedCollection.id;
  const currentCollection = collectionStore.selectedCollection;
  if (currentCollection.name != collectionName.value) {
    await CollectionService.RenameCollection(projectStore.activeProject.uri, collectionId, collectionName.value)
      .then(() => {
        currentCollection.name = collectionName.value;
        emitter.emit('refresh-browser');
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        console.error('Error:', error);
      });
  }
  if (currentCollection.collectionTypeId != collectionTypeId.value) {
    await CollectionService.ChangeType(projectStore.activeProject.uri, collectionId, collectionTypeId.value)
      .then(() => {
        currentCollection.collection_type_name = collectionType.value;
        currentCollection.collection_type_icon = collectionTypeIcon.value;
        currentCollection.collection_type_id = collectionTypeId.value;
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        console.error('Error:', error);
      });
  }
  if (currentCollection.isShared != isShared.value) {
    await CollectionService.ChangeIsShared(projectStore.activeProject.uri, collectionId, isShared.value)
      .then(() => {
        currentCollection.isShared = isShared.value;
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        console.error('Error:', error);
      });
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
  const currentCollection = collectionStore.selectedCollection;
  selectedCollection.value = currentCollection;
  collectionName.value = currentCollection.name;
  oldCollectionName.value = currentCollection.name;
  collectionPreview.value = currentCollection.preview;
  oldCollectionPreview.value = currentCollection.preview;
  collectionType.value = currentCollection.collection_type_name;
  oldCollectionType.value = currentCollection.collection_type_name;
  collectionTypeName.value = currentCollection.collection_type_name;
  collectionTypeIcon.value = currentCollection.collection_type_icon;
  oldIsShared.value = currentCollection.is_shared;
  isShared.value = currentCollection.is_shared;
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
</style>