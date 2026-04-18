<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>

    <HeaderArea :title="title" :icon="headerIcon"/>

    <div class="general-container" :style="{ gap: showAssetOptions ? 10 + 'px' : 20 + 'px' }">

      <!-- Collection Creation Context -->
      <template v-if="!displayTypeCreator">
        <div v-if="!isMultiple" class="input-section">
          <div class="compound-input-section">
            <input v-model="collectionName" class="input-short" type="text" :placeholder="$t('placeholders.collectionName')" v-focus v-return="handleEnterKey" />
          </div>
        </div>

        <BatchGenerator v-else ref="batchGen" @updateData="onUpdateCollections" />

        <div class="input-section">
          <div class="horizontal-flex">
            <div class="dropdown-wrapper">
              <DropDownBox :items="collectionStore.getCollectionTypesNames" :selectedItem="collectionType" :onSelect="selectCollectionType" :useFilter="false" :placeHolder="$t('placeholders.collectionType')" />
            </div>
            <span @click="toggleTypeCreator" class="single-action-button" v-tooltip="$t('modals.addCollectionTypeTitle')">
              <img class="small-icons" :src="getAppIcon('plus-circle')">
            </span>
          </div>
        </div>

        <div v-if="!stage.groupItems" class="horizontal-flex">
          <div class="batch-text">
            {{ $t('modals.generateMultipleItems') }}
          </div>
          <ToggleSwitch v-tooltip="isMultiple? 'Unmark as library' : 'Mark as a library'" @click="toggleIsMultiple" :switchValueProp="isMultiple" />
        </div>

        <div v-if="projectStore.activeProject?.has_remote" class="horizontal-flex">
          <ActionButton :isInactive="true" :icon="getAppIcon('shared')" :label="$t('common.shared')" />
          <ToggleSwitch v-tooltip="isShared? $t('panes.unmarkAsShared') : $t('panes.markAsShared')" @click="toggleIsShared" :switchValueProp="isShared" />
        </div>

        <div class="pop-up-actions" ref="popUpActions">
          <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
          <GeneralButton :label="$t('common.create')" :fullWidth="true" :buttonFunction="createCollections" :isActive="isValueChanged" :loading="isAwaitingResponse" />
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
import { computed, onMounted, onUnmounted, ref, watchEffect } from 'vue';
import { getRelativePath } from '@/lib/pathlib';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import BatchGenerator from '@/instances/desktop/components/BatchGenerator.vue';
import CollectionTypeForm from '@/instances/common/components/CollectionTypeForm.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// services
import { AssetService, CollectionService, FSService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTrayStates } from '@/stores/TrayStates';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const trayStates = useTrayStates();
const { t } = useI18n();

// refs
const batchGen = ref(null);
const collections = ref([]);
const displayTypeCreator = ref(false);
const collectionName = ref('');
const collectionType = ref('');
const isAwaitingResponse = ref(false);
const isShared = ref(false);
const isMultiple = ref(false);
const itemsToGroup = ref([]);
const modalContainer = ref(null);
const newTypeIcon = ref('generic');
const popUpActions = ref(null);
const showAssetOptions = ref(true);
const typeFormRef = ref(null);

// computed
// Returns the header icon based on selected collection type or new type icon.
const headerIcon = computed(() => {
  if (displayTypeCreator.value) {
    return newTypeIcon.value || 'folder-plus';
  }
  const selectedType = collectionStore.collectionTypes.find(item => item.name === collectionType.value);
  return selectedType?.icon || 'folder-plus';
});

// Returns whether the form is valid for submission.
const isValueChanged = computed(() => {
  if (!collectionType.value) return false;
  if (isMultiple.value) {
    return !batchGen.value?.invalidPattern;
  } else {
    return collectionName.value !== '';
  }
});

// Returns the parent ID for the new collection.
const parentId = computed(() => {
  if (stage.selectedItem && stage.selectedItem.type === 'collection') {
    return stage.selectedItem?.id;
  } else if (collectionStore.navigatedCollection) {
    return collectionStore.navigatedCollection.id;
  } else {
    return '';
  }
});

// Returns the selected collection type ID.
const selectedCollectionTypeId = computed(() => {
  const selectedCollectionType = collectionStore.collectionTypes.find(item => item.name === collectionType.value);
  return selectedCollectionType?.id;
});

// Returns the modal title based on current mode.
const title = computed(() => {
  if (displayTypeCreator.value) {
    return t('modals.addCollectionTypeTitle');
  } else if (stage.groupItems) {
    return t('modals.moveIntoNewCollection');
  } else {
    return isMultiple.value ? t('modals.createMultipleCollections') : t('modals.createCollection');
  }
});

// methods
// Changes the parent of one or more collections.
const changeCollectionParent = async (collectionIds, parentId) => {
  await CollectionService.ChangeCollectionParent(projectStore.activeProject.uri, collectionIds, parentId)
    .then(() => {
      notificationStore.addNotification(t('notifications.movedSuccessfully'), '', 'success');
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification(t('notifications.errorChangingCollectionParent'), error);
    });
};

// Moves one or more assets to a different collection.
const changeAssetCollection = async (assetIds, collectionId) => {
  await AssetService.ChangeAssetCollection(projectStore.activeProject.uri, assetIds, collectionId)
    .then(() => {
      notificationStore.addNotification(t('notifications.movedSuccessfully'), '', 'success');
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification(t('notifications.errorChangingAssetCollection'), error);
    });
};

// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility('createCollectionModal', false);
};

// Creates collections based on current mode.
const createCollections = async () => {
  isAwaitingResponse.value = true;
  if (stage.groupItems) {
    await createCollectionAndMove();
  } else if (isMultiple.value) {
    await createMultipleCollections();
    const successMessage = t('notifications.collectionsCreated', { count: collections.value.length });
    notificationStore.addNotification(successMessage, '', 'success');
  } else {
    await createSingleCollection();
  }
  isAwaitingResponse.value = false;
  emitter.emit('refresh-browser');
  closeModal();
};

// Creates an collection and moves selected items into it.
const createCollectionAndMove = async () => {
  const referenceItem = stage.selectedItems.at(-1);
  const type = referenceItem.type;
  const project = projectStore.activeProject;
  let parent;
  if (type === 'asset') {
    parent = await CollectionService.GetCollectionByID(project.uri, referenceItem.collection_id);
  } else {
    parent = await CollectionService.GetCollectionByID(project.uri, referenceItem.parent_id);
  }
  if (!parent) return;
  const parentIdValue = parent?.id;
  isAwaitingResponse.value = true;
  const selectedCollectionType = collectionStore.collectionTypes.find(item => item.name === collectionType.value);
  CollectionService.CreateCollection(projectStore.activeProject.uri, collectionName.value, '', selectedCollectionType.id, parentIdValue, '', isShared.value)
    .then(async data => {
      const newCollection = data;
      collectionStore.selectedCollection = newCollection;
      isAwaitingResponse.value = false;
      await moveIntoFolder(newCollection.id);
      closeModal();
      notificationStore.addNotification(t('notifications.collectionCreated', { name: collectionName.value }), '', 'success');
      if (parentIdValue && !(parentIdValue in stage.expandedCollections)) {
        stage.expandCollection(parent);
      }
      stage.firstSelectedItemId = newCollection.id;
      stage.markedItems = [newCollection.id];
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      console.log(error);
      notificationStore.errorNotification(t('notifications.errorCreatingCollection'), error);
    });
};

// Creates multiple collections from batch generator.
const createMultipleCollections = async () => {
  const collectionNames = collections.value;
  for (const collectionName of collectionNames) {
    collectionName.value = collectionName;
    await createSingleCollection();
  }
};

// Creates a single collection.
const createSingleCollection = async () => {
  await CollectionService.CreateCollection(projectStore.activeProject.uri, collectionName.value, '', selectedCollectionTypeId.value, parentId.value, '', isShared.value)
    .then(async data => {
      if (!isMultiple.value) {
        const newCollection = data;
        collectionStore.selectedCollection = newCollection;
        stage.selectedItem = newCollection;
        notificationStore.addNotification(t('notifications.collectionCreated', { name: collectionName.value }), '', 'success');
        stage.firstSelectedItemId = newCollection.id;
        stage.markedItems = [newCollection.id];
      }
    })
    .catch((error) => {
      console.log(error);
      notificationStore.errorNotification(t('notifications.errorCreatingCollection'), error);
    });
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles enter key press to submit form.
const handleEnterKey = () => {
  createCollections();
};

// Handles successful type creation from the form.
const handleTypeCreated = (response) => {
  collectionType.value = response.name;
  displayTypeCreator.value = false;
};

// Handles icon change from the type form.
const handleTypeIconChange = (icon) => {
  newTypeIcon.value = icon;
};

// Moves selected items into the specified folder.
const moveIntoFolder = async (activeItemId) => {
  const selectedItems = stage.selectedItems;

  // Collect items by type for batch operations
  const collectionIds = [];
  const assetIds = [];
  const untrackedItems = [];

  for (const item of selectedItems) {
    if (item.type === 'collection') collectionIds.push(item.id);
    else if (item.type === 'asset') assetIds.push(item.id);
    else untrackedItems.push(item);
  }

  // Execute batch operations for tracked items
  if (collectionIds.length) await changeCollectionParent(collectionIds, activeItemId);
  if (assetIds.length) await changeAssetCollection(assetIds, activeItemId);

  // Handle untracked items
  if (untrackedItems.length) {
    const collection = collectionStore.selectedCollection;
    await FSService.MakeDirs(collection.file_path);
    const renameOperations = [];
    for (const item of untrackedItems) {
      const newPath = await FSService.JoinPath(collection.file_path, item.name);
      renameOperations.push({ oldPath: item.file_path, newPath });
    }
    await FSService.RenameBatch(JSON.stringify(renameOperations));
  }
};

// Updates collections from batch generator.
const onUpdateCollections = (allCollections) => {
  collections.value = allCollections;
};

// Selects an collection type from the dropdown.
const selectCollectionType = (collectionTypeName) => {
  collectionType.value = collectionTypeName;
};

// Toggles the shared flag.
const toggleIsShared = () => {
  isShared.value = !isShared.value;
};

// Toggles the multiple mode.
const toggleIsMultiple = () => {
  isMultiple.value = !isMultiple.value;
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
  if (stage.groupItems) {
    itemsToGroup.value = stage.markedItems;
  }
  trayStates.listItemsBoundary = modalContainer.value;
  trayStates.tagSearchQuery = '';
});

onUnmounted(() => {
  stage.groupItems = false;
  stage.markedCollections = [];
  stage.selectedItem = null;
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

.batch-text{
  font-size: 14px;
  padding-left: .5rem;
}

.input-short {
  flex: 1;
  width: 100%;
}
</style>