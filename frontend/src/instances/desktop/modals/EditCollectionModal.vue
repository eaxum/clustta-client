<template>

  <div ref="modalContainer" class="modal-container" v-esc="closeModal" v-return="handleEnterKey" v-stop-propagation>

    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="getAppIcon(entityTypeIcon)" :showSearch="false" />
      <ActionButton v-if="isPreviewChanged" :icon="getAppIcon('revert')" :showLabel="false"
        v-tooltip="'Revert Cover Image'" :buttonFunction="revertCoverImage" />
      <ActionButton v-if="entityPreview" :icon="getAppIcon('trash')" :showLabel="false" v-tooltip="'Remove Cover Image'"
        :buttonFunction="removeCoverImage" />
      <ActionButton v-else :icon="getAppIcon('image-plus')" :showLabel="false" v-tooltip="'Add Cover Image'"
        :buttonFunction="addCoverImage" />
    </div>

    <div class="general-container">

      <span @click="addCoverImage" v-if="entityPreview" class="screenshot-preview">
        <img class="screenshot-thumb" :src="entityPreview">
      </span>

      <div class="input-section">
        <input v-model="entityName" class="input-short" type="text" placeholder="Collection Name" v-focus />
      </div>

      <div class="input-section">
        <DropDownBox :items="entityStore.getEntityTypesNames" :selectedItem="entityType" :onSelect="changeEntityType" />
      </div>

      <div class="horizontal-flex is-library-prompt">
        
      <ActionButton :isInactive="true" :icon="getAppIcon('bookmark')" :label="'Library'" />
      <ToggleSwitch v-tooltip="isLibrary? 'Unmark as library' : 'Mark as a library'" @click="toggleIsLibrary" :switchValueProp="isLibrary" />

    </div>

      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Update'" :fullWidth="true" @click="updateEntity()" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>


    </div>
  </div>
</template>

<script setup>


// imports
import { ref, watchEffect, computed, onMounted } from 'vue';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// services
import { DialogService } from '@/../bindings/clustta/services/index';
import { EntityService } from "@/../bindings/clustta/services";

// state imports
import { useNotificationStore } from '@/stores/notifications';
import { useEntityStore } from '@/stores/entity';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useProjectStore } from '@/stores/projects';
import { useMenu } from '@/stores/menu';
import { useIconStore } from '@/stores/icons';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// states
const entityStore = useEntityStore();
const projectStore = useProjectStore();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const menu = useMenu();

// vars
let title = 'Collection Details';

// refs
const entityName = ref('');
const oldEntityName = ref('');
const entityType = ref('');
const oldEntityType = ref('');
const entityTypeName = ref('');
const entityTypeId = ref('');
const entityTypeIcon = ref('');
const isAwaitingResponse = ref(false);
const modalContainer = ref(null);
const selectedEntity = ref(null);
const entityPreview = ref(null);
const oldEntityPreview = ref(null);
const coverImagePath = ref('');
const isLibrary = ref(null);
const OldisLibrary = ref(null);

// computed properties
const entity = computed(() => {
  return entityStore.selectedEntity;
});

const isNameChanged = computed(() => {
  const restrictedEntries = [oldEntityName.value, ''];
  return !restrictedEntries.includes(entityName.value);
});

const isTypeChanged = computed(() => {
  return oldEntityType.value?.toLowerCase() !== entityType.value?.toLowerCase()
});

const isPreviewChanged = computed(() => {
  return oldEntityPreview.value !== entityPreview.value;
});

const isLibraryChanged = computed(() => {
  return OldisLibrary.value !== isLibrary.value;
});

const isValueChanged = computed(() => {
  return isTypeChanged.value || isNameChanged.value || isPreviewChanged.value || isLibraryChanged.value;
});


const icon = computed(() => {
  return '/types-icons/' + entityTypeIcon.value + '.svg'
});



// methods
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    updateEntity();
  }
};

const changeEntityType = (entityTypeName) => {

  let newEntityType;
  const entityTypes = entityStore.getEntityTypes;
  newEntityType = entityTypes.find((item) => item.name === entityTypeName);

  entityType.value = newEntityType.name;
  entityTypeIcon.value = newEntityType.icon;
  entityTypeId.value = newEntityType.id;

};

const addCoverImage = async () => {
  const result = await DialogService.SelectFileDialog("Select Image File", "*.png;*.jpg;*.jpeg;*.gif;*.bmp;*.webp");
  if (result) {
    let filePath = result.replace(/\\/g, '/');
    let base64Image = await utils.base64FromFile(filePath)
    coverImagePath.value = filePath;
    entityPreview.value = base64Image;
  }
};

const removeCoverImage = () => {
  entityPreview.value = null;
  // coverImagePath.value = null;
};

const revertCoverImage = () => {
  entityPreview.value = oldEntityPreview.value;
};

const closeModal = (all) => {
  modals.disableAllModals()
};

const updateEntityMeta = async () => {

  let entityId = entityStore.selectedEntity.id;
  let newEntityTypeId = entityTypeId.value;
  let entity = entityStore.selectedEntity;
  if (entity.name != entityName.value) {
    await EntityService.RenameEntity(projectStore.activeProject.uri, entityId, entityName.value)
      .then((data) => {
        entity.name = entityName.value;
        emitter.emit('refresh-browser');
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        console.error('Error:', error);
      });
  }
  if (entity.entityTypeId != newEntityTypeId) {
    await EntityService.ChangeType(projectStore.activeProject.uri, entityId, newEntityTypeId)
      .then((data) => {
        entity.entity_type_name = entityType.value;
        entity.entity_type_icon = entityTypeIcon.value;
        entity.entity_type_id = entityTypeId.value;
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        console.error('Error:', error);
      });
  }
  if (entity.isLibrary != isLibrary.value) {
    await EntityService.ChangeIsLibrary(projectStore.activeProject.uri, entityId, isLibrary.value)
    .then((data) => {
      entity.isLibrary = isLibrary.value;
    })
    .catch((error) => {
        isAwaitingResponse.value = false;
        console.error('Error:', error);
    });
  }
}

const updateEntityCover = async () => {

  let entityId = entityStore.selectedEntity.id;
  let entity = entityStore.findEntity(entityId);

  const filePath = coverImagePath.value;
  console.log(filePath)
  await EntityService.UpdatePreview(projectStore.activeProject.uri, entityId, filePath).then(() => {
    entity.preview = entityPreview.value;
    console.log('removed')
  }).catch((error) => {
    console.error(error)
    isAwaitingResponse.value = false;
    notificationStore.addNotification(
      "Error Updating Image",
      error,
      "error",
      false
    )
  });

}

const updateEntity = async () => {

  isAwaitingResponse.value = true;

  if (isTypeChanged.value || isNameChanged.value || isLibraryChanged.value) {
    console.log('meta changed');
    await updateEntityMeta();
  }
  if (isPreviewChanged.value) {
    console.log('image changed');
    await updateEntityCover();
  }

  await entityStore.reloadEntities();
  isAwaitingResponse.value = false;
  closeModal();

}

const toggleIsLibrary = () => {
  isLibrary.value = !isLibrary.value;
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// onMounted
onMounted(() => {
  let entity = entityStore.selectedEntity;
  selectedEntity.value = entityStore.selectedEntity;
  entityName.value = entity.name;
  oldEntityName.value = entity.name;
  entityPreview.value = entity.preview;
  oldEntityPreview.value = entity.preview;
  entityType.value = entity.entity_type_name;
  oldEntityType.value = entity.entity_type_name;
  entityTypeName.value = entity.entity_type_name;
  entityTypeIcon.value = entity.entity_type_icon;
  OldisLibrary.value = entity.is_library;
  isLibrary.value = entity.is_library;


});


</script>


<style scoped>
@import "@/assets/desktop.css";

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
  color: white;
  font-size: 16px;
  white-space: nowrap;
  flex: 1;
}
.is-library-prompt {
  /* padding: 1rem .5rem; */
}
</style>

