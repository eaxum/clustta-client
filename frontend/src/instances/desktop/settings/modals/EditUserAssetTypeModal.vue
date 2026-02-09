<template>

  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">

    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="taskTypeIcon" />
    </div>


    <div class="general-container">
      <div class="input-section">
        <input v-model="taskTypeName" class="input-short" type="text" placeholder="Task type Name" v-focus
          @keydown.enter="handleEnterKey" />
      </div>

      <IconGrid v-if="displayIconSelector" @iconSelected="setIcon" :icons="icons" />
      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Update'" :fullWidth="true" @click="updateTaskType" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import iconData from "@/data/iconData.json";

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import IconGrid from '@/instances/desktop/components/IconGrid.vue';

// services
import { AssetService } from "@/services";

// stores
const assetStore = useAssetStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

import { useAssetStore } from '@/stores/assets';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

// constants
const title = 'Edit task type';

// refs
const displayIconSelector = ref(true);
const isAwaitingResponse = ref(false);
const taskTypeIcon = ref('generic');
const taskTypeName = ref('');

// computed
const icons = computed(() => {
  const allIcons = iconData.icons;
  const allTaskTypeIcons = assetStore.assetTypes.map((item) => item.icon);
  return allIcons.filter((icon) => !allTaskTypeIcons.includes(icon));
});

const isValueChanged = computed(() => {
  return !!taskTypeName.value && taskTypeIcon.value !== 'generic';
});

// methods

// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility("editUserAssetTypeModal", false);
};

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles enter key press.
const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    // updateTaskType();
  }
};

// Sets the selected icon.
const setIcon = (icon) => {
  taskTypeIcon.value = icon;
};

// Updates the task type with the new values.
const updateTaskType = () => {
  AssetService.UpdateAssetType(projectStore.activeProject.uri, assetStore.selectedAssetType.id, taskTypeName.value, taskTypeIcon.value)
    .then((response) => {
      notificationStore.addNotification("Task Type Updated", "", "success");
      const index = assetStore.assetTypes.findIndex(taskType => taskType.id === assetStore.selectedAssetType.id);
      assetStore.assetTypes[index] = response;
      closeModal();
    })
    .catch((error) => {
      notificationStore.errorNotification("Error Updating Task Type", error);
    });
};

// lifecycle
onMounted(() => {
  // taskTypeName.value = assetStore.selectedAssetType.name;
  // taskTypeIcon.value = assetStore.selectedAssetType.icon;
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.input-short {
  flex: 1;
  width: 100%;
}
</style>


