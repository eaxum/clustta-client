<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>

    <HeaderArea :title="title" :icon="headerIcon"/>
    <div class="general-container" :style="{ gap: showTaskOptions ? 10 + 'px' : 20 + 'px' }">

      <div v-if="!isMultiple" class="input-section">
        <div class="compound-input-section">
          <input v-model="entityName" class="input-short" type="text" placeholder="Collection Name" v-focus
            v-return="handleEnterKey" />
        </div>
      </div>

      <BatchGenerator v-else ref="batchGen" @updateData="onUpdateCollections" />

      <div class="input-section">
        <DropDownBox :items="collectionStore.getCollectionTypesNames" :selectedItem="entityType" :onSelect="selectEntityType" />
      </div>

      <div v-if="!stage.groupItems" class="horizontal-flex">
        Generate Multiple Items
        <ToggleSwitch v-tooltip="isMultiple? 'Unmark as library' : 'Mark as a library'" @click="toggleIsMultiple" :switchValueProp="isMultiple" />
      </div>

      <div class="horizontal-flex">
        <ActionButton :isInactive="true" :icon="getAppIcon('library')" :label="'Library'" />
        <ToggleSwitch v-tooltip="isLibrary? 'Unmark as library' : 'Mark as a library'" @click="toggleIsLibrary" :switchValueProp="isLibrary" />
      </div>

      <div class="pop-up-actions" ref="popUpActions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Confirm'" :fullWidth="true" :buttonFunction="createCollections" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>

    </div>
  </div>

</template>

<script setup>
// imports
import { computed, onMounted, onUnmounted, ref, watchEffect } from 'vue';
import { getRelativePath } from '@/lib/pathlib';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import BatchGenerator from '@/instances/desktop/components/BatchGenerator.vue';
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

// refs
const batchGen = ref(null);
const collections = ref([]);
const entityName = ref('');
const entityType = ref(collectionStore.getCollectionTypesNames[0]);
const isAwaitingResponse = ref(false);
const isLibrary = ref(false);
const isMultiple = ref(false);
const itemsToGroup = ref([]);
const modalContainer = ref(null);
const popUpActions = ref(null);
const showTaskOptions = ref(true);

// computed
// Returns the header icon based on selected entity type.
const headerIcon = computed(() => {
  const selectedType = collectionStore.collectionTypes.find(item => item.name === entityType.value);
  return selectedType?.icon || 'folder-plus';
});

// Returns whether the form is valid for submission.
const isValueChanged = computed(() => {
  if (isMultiple.value) {
    return !batchGen.value?.invalidPattern;
  } else {
    return entityName.value !== '';
  }
});

// Returns the parent ID for the new collection.
const parentId = computed(() => {
  if (stage.selectedItem && stage.selectedItem.type === 'entity') {
    return stage.selectedItem?.id;
  } else if (collectionStore.navigatedCollection) {
    return collectionStore.navigatedCollection.id;
  } else {
    return '';
  }
});

// Returns the selected entity type ID.
const selectedEntityTypeId = computed(() => {
  const selectedEntityType = collectionStore.collectionTypes.find(item => item.name === entityType.value);
  return selectedEntityType?.id;
});

// Returns the modal title based on current mode.
const title = computed(() => {
  if (stage.groupItems) {
    return 'Move into new Collection';
  } else {
    return isMultiple.value ? 'Create Multiple Collections' : 'Create Collection';
  }
});

// methods
// Changes the parent of one or more entities.
const changeEntityParent = async (entityIds, parentId) => {
  await CollectionService.ChangeCollectionParent(projectStore.activeProject.uri, entityIds, parentId)
    .then(() => {
      notificationStore.addNotification('Moved successfully.', '', 'success');
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification('Error changing entity parent', error);
    });
};

// Moves one or more tasks to a different collection.
const changeTaskEntity = async (taskIds, entityId) => {
  await AssetService.ChangeAssetCollection(projectStore.activeProject.uri, taskIds, entityId)
    .then(() => {
      notificationStore.addNotification('Moved successfully.', '', 'success');
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification('Error changing task entity', error);
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
    await createEntityAndMove();
  } else if (isMultiple.value) {
    await createMultipleEntities();
    const successMessage = collections.value.length + ' collections created';
    notificationStore.addNotification(successMessage, '', 'success');
  } else {
    await createSingleEntity();
  }
  isAwaitingResponse.value = false;
  emitter.emit('refresh-browser');
  closeModal();
};

// Creates an entity and moves selected items into it.
const createEntityAndMove = async () => {
  const referenceItem = stage.selectedItems.at(-1);
  const type = referenceItem.type;
  const project = projectStore.activeProject;
  let parent;
  if (type === 'task') {
    parent = await CollectionService.GetCollectionByID(project.uri, referenceItem.entity_id);
  } else {
    parent = await CollectionService.GetCollectionByID(project.uri, referenceItem.parent_id);
  }
  if (!parent) return;
  const parentIdValue = parent?.id;
  isAwaitingResponse.value = true;
  const selectedEntityType = collectionStore.collectionTypes.find(item => item.name === entityType.value);
  CollectionService.CreateCollection(projectStore.activeProject.uri, entityName.value, '', selectedEntityType.id, parentIdValue, '', isLibrary.value)
    .then(async data => {
      const newEntity = data;
      collectionStore.selectedCollection = newEntity;
      isAwaitingResponse.value = false;
      await moveIntoFolder(newEntity.id);
      closeModal();
      notificationStore.addNotification(entityName.value + ' collection created', '', 'success');
      if (parentIdValue && !(parentIdValue in stage.expandedEntities)) {
        stage.expandEntity(parent);
      }
      stage.firstSelectedItemId = newEntity.id;
      stage.markedItems = [newEntity.id];
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      console.log(error);
      notificationStore.errorNotification('Error creating entity', error);
    });
};

// Creates multiple entities from batch generator.
const createMultipleEntities = async () => {
  const collectionNames = collections.value;
  for (const collectionName of collectionNames) {
    entityName.value = collectionName;
    await createSingleEntity();
  }
};

// Creates a single entity.
const createSingleEntity = async () => {
  await CollectionService.CreateCollection(projectStore.activeProject.uri, entityName.value, '', selectedEntityTypeId.value, parentId.value, '', isLibrary.value)
    .then(async data => {
      if (!isMultiple.value) {
        const newEntity = data;
        collectionStore.selectedCollection = newEntity;
        stage.selectedItem = newEntity;
        notificationStore.addNotification(entityName.value + ' collection created', '', 'success');
        stage.firstSelectedItemId = newEntity.id;
        stage.markedItems = [newEntity.id];
      }
    })
    .catch((error) => {
      console.log(error);
      notificationStore.errorNotification('Error creating entity', error);
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

// Moves selected items into the specified folder.
const moveIntoFolder = async (activeItemId) => {
  const selectedItems = stage.selectedItems;

  // Collect items by type for batch operations
  const entityIds = [];
  const taskIds = [];
  const untrackedItems = [];

  for (const item of selectedItems) {
    if (item.type === 'entity') entityIds.push(item.id);
    else if (item.type === 'task') taskIds.push(item.id);
    else untrackedItems.push(item);
  }

  // Execute batch operations for tracked items
  if (entityIds.length) await changeEntityParent(entityIds, activeItemId);
  if (taskIds.length) await changeTaskEntity(taskIds, activeItemId);

  // Handle untracked items
  if (untrackedItems.length) {
    const entity = collectionStore.selectedCollection;
    await FSService.MakeDirs(entity.file_path);
    const renameOperations = [];
    for (const item of untrackedItems) {
      const newPath = await FSService.JoinPath(entity.file_path, item.name);
      renameOperations.push({ oldPath: item.file_path, newPath });
    }
    await FSService.RenameBatch(renameOperations);
  }
};

// Updates collections from batch generator.
const onUpdateCollections = (allCollections) => {
  collections.value = allCollections;
};

// Selects an entity type from the dropdown.
const selectEntityType = (entityTypeName) => {
  entityType.value = entityTypeName;
};

// Toggles the library flag.
const toggleIsLibrary = () => {
  isLibrary.value = !isLibrary.value;
};

// Toggles the multiple mode.
const toggleIsMultiple = () => {
  isMultiple.value = !isMultiple.value;
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
  stage.markedEntities = [];
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

.input-short {
  flex: 1;
  width: 100%;
}
</style>