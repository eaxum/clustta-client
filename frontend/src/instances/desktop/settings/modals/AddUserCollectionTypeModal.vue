<template>

  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">

    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="entityTypeIcon" />
    </div>

    <div class="general-container">
      <div class="input-section">
        <input v-model="entityTypeName" class="input-short" type="text" placeholder="Collection type Name" v-focus
          @keydown.enter="handleEnterKey" />
      </div>

      <IconGrid v-if="displayIconSelector" @iconSelected="setIcon" :icons="icons" />

      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Create'" :fullWidth="true" @click="createEntityType" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>


    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref } from 'vue';
import iconData from "@/data/iconData.json";

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import IconGrid from '@/instances/desktop/components/IconGrid.vue';

// services
import { CollectionService } from "@/services";

// stores
const collectionStore = useCollectionStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectTemplateStore = useProjectTemplateStore();

import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectTemplateStore } from '@/stores/project_template';

// constants
const title = 'Add Collection type';

// refs
const displayIconSelector = ref(true);
const entityTypeIcon = ref('generic');
const entityTypeName = ref('');
const isAwaitingResponse = ref(false);

// computed
const icons = computed(() => {
  const allIcons = iconData.icons;
  const allEntityTypeIcons = collectionStore.getCollectionTypes.map((item) => item.icon);
  return allIcons.filter((icon) => !allEntityTypeIcons.includes(icon));
});

const isValueChanged = computed(() => {
  return !!entityTypeName.value && entityTypeIcon.value !== 'generic';
});

// methods

// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility("addUserCollectionTypeModal", false);
};

// Creates a new collection type.
const createEntityType = () => {
  CollectionService.CreateCollectionType(projectTemplateStore.activeProjectTemplate.uri, entityTypeName.value, entityTypeIcon.value)
    .then(() => {
      notificationStore.addNotification("Collection type created", "", "success");
      projectTemplateStore.reloadProjectTemplate();
      closeModal();
    })
    .catch((error) => {
      notificationStore.errorNotification("Error deleting folder Type", error);
    });
};

// Handles enter key press.
const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    // createEntityType();
  }
};

// Sets the selected icon.
const setIcon = (icon) => {
  entityTypeIcon.value = icon;
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.input-short {
  flex: 1;
  width: 100%;
}
</style>