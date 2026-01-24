<template>

  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">
    
    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="taskTypeIcon"  />
    </div>


    <div class="general-container">
      <div class="input-section">
        <input v-model="taskTypeName" class="input-short" type="text" placeholder="Asset type Name" v-focus
          @keydown.enter="handleEnterKey" />
      </div>

      <IconGrid v-if="displayIconSelector" @iconSelected="setIcon"  :icons="icons"/>
      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Create'" :fullWidth="true" @click="createTaskType" :isActive="isValueChanged"
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
import { AssetService } from "@/services";

// stores
import { useAssetStore } from '@/stores/assets';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

const assetStore = useAssetStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

// refs
const displayIconSelector = ref(true);
const isAwaitingResponse = ref(false);
const modalContainer = ref(null);
const taskTypeIcon = ref('generic');
const taskTypeName = ref('');

// constants
const title = 'Add Asset type';

// computed
// Returns available icons excluding already used ones.
const icons = computed(() => {
  const allIcons = iconData.icons.filter((item) => item !== 'weblink');
  const allTaskTypeIcons = assetStore.assetTypes.map((item) => item.icon);
  return allIcons.filter((icon) => !allTaskTypeIcons.includes(icon));
});

// Checks if form is valid for submission.
const isValueChanged = computed(() => {
  return !!taskTypeName.value && taskTypeIcon.value !== 'generic';
});

// methods
// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility("addAssetTypeModal", false);
};

// Creates a new asset type.
const createTaskType = () => {
  AssetService.CreateAssetType(projectStore.activeProject.uri, taskTypeName.value, taskTypeIcon.value)
    .then((response) => {
      notificationStore.addNotification("Task Type Created", "", "success");
      assetStore.assetTypes.push(response);
      closeModal();
    })
    .catch((error) => {
      notificationStore.errorNotification("Error Deleting Task Type", error);
    });
};

// Handles enter key press.
const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    // createTaskType();
  }
};

// Sets the selected icon.
const setIcon = (icon) => {
  taskTypeIcon.value = icon;
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.input-short {
  flex: 1;
  width: 100%;
}
</style>


