<template>

  <div ref="modalContainer" class="modal-container" v-esc="closeModal" v-return="handleEnterKey" v-stop-propagation>

    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="getAppIcon(displayTypeCreator ? newTypeIcon : entityTypeIcon)" :showSearch="false" />
      <ActionButton v-if="isPreviewChanged && !displayTypeCreator" :icon="getAppIcon('revert')" :showLabel="false"
        v-tooltip="'Revert Cover Image'" :buttonFunction="revertCoverImage" />
      <ActionButton v-if="entityPreview && !displayTypeCreator" :icon="getAppIcon('trash')" :showLabel="false" v-tooltip="'Remove Cover Image'"
        :buttonFunction="removeCoverImage" />
      <ActionButton v-if="!entityPreview && !displayTypeCreator" :icon="getAppIcon('image-plus')" :showLabel="false" v-tooltip="'Add Cover Image'"
        :buttonFunction="addCoverImage" />
    </div>

    <div class="general-container">

      <!-- Collection Edit Context -->
      <template v-if="!displayTypeCreator">
        <span @click="addCoverImage" v-if="entityPreview" class="screenshot-preview">
          <img class="screenshot-thumb" :src="entityPreview">
        </span>

        <div class="input-section">
          <input v-model="entityName" class="input-short" type="text" placeholder="Collection Name" v-focus />
        </div>

        <div class="input-section">
          <div class="horizontal-flex">
            <div class="dropdown-wrapper">
              <DropDownBox :items="collectionStore.getCollectionTypesNames" :selectedItem="entityType" :onSelect="changeEntityType" />
            </div>
            <span @click="toggleTypeCreator" class="single-action-button" v-tooltip="'Add New Collection Type'">
              <img class="small-icons" :src="getAppIcon('plus-circle')">
            </span>
          </div>
        </div>

        <div class="horizontal-flex is-library-prompt">
          <ActionButton :isInactive="true" :icon="getAppIcon('library')" :label="'Library'" />
          <ToggleSwitch v-tooltip="isLibrary? 'Unmark as library' : 'Mark as a library'" @click="toggleIsLibrary" :switchValueProp="isLibrary" />
        </div>

        <div class="pop-up-actions">
          <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
          <GeneralButton :label="'Update'" :fullWidth="true" @click="updateEntity()" :isActive="isValueChanged" :loading="isAwaitingResponse" />
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

const collectionStore = useCollectionStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

// refs
const coverImagePath = ref('');
const displayTypeCreator = ref(false);
const entityName = ref('');
const entityPreview = ref(null);
const entityType = ref('');
const entityTypeIcon = ref('');
const entityTypeId = ref('');
const entityTypeName = ref('');
const isAwaitingResponse = ref(false);
const isLibrary = ref(null);
const modalContainer = ref(null);
const newTypeIcon = ref('generic');
const oldEntityName = ref('');
const oldEntityPreview = ref(null);
const oldEntityType = ref('');
const OldisLibrary = ref(null);
const selectedEntity = ref(null);
const typeFormRef = ref(null);

// constants
const title = computed(() => {
  if (displayTypeCreator.value) {
    return 'Add Collection Type';
  }
  return 'Collection Details';
});

// computed
// Returns the currently selected entity.
const entity = computed(() => {
  return collectionStore.selectedCollection;
});

// Returns whether the library flag has changed.
const isLibraryChanged = computed(() => {
  return OldisLibrary.value !== isLibrary.value;
});

// Returns whether the entity name has changed.
const isNameChanged = computed(() => {
  const restrictedEntries = [oldEntityName.value, ''];
  return !restrictedEntries.includes(entityName.value);
});

// Returns whether the preview image has changed.
const isPreviewChanged = computed(() => {
  return oldEntityPreview.value !== entityPreview.value;
});

// Returns whether the entity type has changed.
const isTypeChanged = computed(() => {
  return oldEntityType.value?.toLowerCase() !== entityType.value?.toLowerCase();
});

// Returns whether any form values have changed.
const isValueChanged = computed(() => {
  return isTypeChanged.value || isNameChanged.value || isPreviewChanged.value || isLibraryChanged.value;
});

// methods
// Opens a dialog to select a cover image.
const addCoverImage = async () => {
  const result = await DialogService.SelectFileDialog('Select Image File', '*.png;*.jpg;*.jpeg;*.gif;*.bmp;*.webp');
  if (result) {
    const filePath = result.replace(/\\/g, '/');
    const base64Image = await utils.base64FromFile(filePath);
    coverImagePath.value = filePath;
    entityPreview.value = base64Image;
  }
};

// Changes the entity type.
const changeEntityType = (newEntityTypeName) => {
  const entityTypes = collectionStore.getCollectionTypes;
  const newEntityType = entityTypes.find((item) => item.name === newEntityTypeName);
  entityType.value = newEntityType.name;
  entityTypeIcon.value = newEntityType.icon;
  entityTypeId.value = newEntityType.id;
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
    updateEntity();
  }
};

// Handles successful type creation from the form.
const handleTypeCreated = (response) => {
  entityType.value = response.name;
  entityTypeIcon.value = response.icon;
  entityTypeId.value = response.id;
  displayTypeCreator.value = false;
};

// Handles icon change from the type form.
const handleTypeIconChange = (icon) => {
  newTypeIcon.value = icon;
};

// Removes the current cover image.
const removeCoverImage = () => {
  entityPreview.value = null;
};

// Reverts to the original cover image.
const revertCoverImage = () => {
  entityPreview.value = oldEntityPreview.value;
};

// Toggles the library flag.
const toggleIsLibrary = () => {
  isLibrary.value = !isLibrary.value;
};

// Toggles the type creator context.
const toggleTypeCreator = () => {
  displayTypeCreator.value = !displayTypeCreator.value;
  if (!displayTypeCreator.value) {
    newTypeIcon.value = 'generic';
  }
};

// Updates the entity with all changed values.
const updateEntity = async () => {
  isAwaitingResponse.value = true;
  if (isTypeChanged.value || isNameChanged.value || isLibraryChanged.value) {
    await updateEntityMeta();
  }
  if (isPreviewChanged.value) {
    await updateEntityCover();
  }
  await collectionStore.reloadCollections();
  isAwaitingResponse.value = false;
  closeModal();
};

// Updates the entity cover image.
const updateEntityCover = async () => {
  const entityId = collectionStore.selectedCollection.id;
  const currentEntity = collectionStore.findCollection(entityId);
  const filePath = coverImagePath.value;
  await CollectionService.UpdatePreview(projectStore.activeProject.uri, entityId, filePath)
    .then(() => {
      currentEntity.preview = entityPreview.value;
    })
    .catch((error) => {
      console.error(error);
      isAwaitingResponse.value = false;
      notificationStore.addNotification('Error Updating Image', error, 'error', false);
    });
};

// Updates the entity metadata (name, type, library flag).
const updateEntityMeta = async () => {
  const entityId = collectionStore.selectedCollection.id;
  const currentEntity = collectionStore.selectedCollection;
  if (currentEntity.name != entityName.value) {
    await CollectionService.RenameCollection(projectStore.activeProject.uri, entityId, entityName.value)
      .then(() => {
        currentEntity.name = entityName.value;
        emitter.emit('refresh-browser');
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        console.error('Error:', error);
      });
  }
  if (currentEntity.entityTypeId != entityTypeId.value) {
    await CollectionService.ChangeType(projectStore.activeProject.uri, entityId, entityTypeId.value)
      .then(() => {
        currentEntity.entity_type_name = entityType.value;
        currentEntity.entity_type_icon = entityTypeIcon.value;
        currentEntity.entity_type_id = entityTypeId.value;
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        console.error('Error:', error);
      });
  }
  if (currentEntity.isLibrary != isLibrary.value) {
    await CollectionService.ChangeIsLibrary(projectStore.activeProject.uri, entityId, isLibrary.value)
      .then(() => {
        currentEntity.isLibrary = isLibrary.value;
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
  const currentEntity = collectionStore.selectedCollection;
  selectedEntity.value = currentEntity;
  entityName.value = currentEntity.name;
  oldEntityName.value = currentEntity.name;
  entityPreview.value = currentEntity.preview;
  oldEntityPreview.value = currentEntity.preview;
  entityType.value = currentEntity.entity_type_name;
  oldEntityType.value = currentEntity.entity_type_name;
  entityTypeName.value = currentEntity.entity_type_name;
  entityTypeIcon.value = currentEntity.entity_type_icon;
  OldisLibrary.value = currentEntity.is_library;
  isLibrary.value = currentEntity.is_library;
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