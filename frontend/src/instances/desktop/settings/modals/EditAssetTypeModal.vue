<template>

  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">

    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="taskTypeIcon" />
    </div>


    <div class="general-container">
      <div class="input-section">
        <input v-model="taskTypeName" class="input-short" type="text" placeholder="Asset type Name" v-focus
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
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};


import { ref, onMounted, computed } from 'vue';
import { useProjectStore } from '@/stores/projects';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { TaskService } from "@/../bindings/clustta/services";
import { useAssetStore } from '@/stores/assets';
import iconData from "@/data/iconData.json";

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import IconGrid from '@/instances/desktop/components/IconGrid.vue';

// vars
let title = 'Edit asset type';
const isAwaitingResponse = ref(false);

const icons = computed(() => {
  const allIcons = iconData.icons;
  const allTaskTypeIcons = assetStore.assetTypes.map((item) => item.icon);
  return allIcons.filter((icon) => !allTaskTypeIcons.includes(icon))
})

// stores
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const assetStore = useAssetStore();

const taskTypeName = ref('');
const taskTypeIcon = ref('generic');

const displayIconSelector = ref(true);

const toggleIconSelector = () => {
  displayIconSelector.value = !displayIconSelector.value
};

const setIcon = (icon) => {
  taskTypeIcon.value = icon;
};

const isValueChanged = computed(() => {
  return !!taskTypeName.value && taskTypeIcon.value !== 'generic'
});

const closeModal = () => {
  modals.setModalVisibility("editTaskTypeModal", false);
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    // updateTaskType();
  }
};

const updateTaskType = () => {
  TaskService.UpdateTaskType(projectStore.activeProject.uri, assetStore.selectedAssetType.id, taskTypeName.value, taskTypeIcon.value)
    .then((response) => {
      notificationStore.addNotification("Task Type Updated", "", "success");
      const index = assetStore.assetTypes.findIndex(taskType => taskType.id === assetStore.selectedAssetType.id);
      assetStore.assetTypes[index] = response
      closeModal();
    })
    .catch((error) => {
      notificationStore.errorNotification("Error Updating Task Type", error);
    });
};

onMounted(() => {
  taskTypeName.value = assetStore.selectedAssetType.name;
  taskTypeIcon.value = assetStore.selectedAssetType.icon;
})


</script>

<style scoped>
@import "@/assets/desktop.css";

.add-category {

  display: flex;
  gap: .5rem;
  flex-direction: row;
  /* background-color: chocolate; */
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
  color: white;
  font-size: 16px;
  white-space: nowrap;
  flex: 1;

}

/* .general-container {
  gap: 10px;
  box-sizing: border-box;
  width: 300px;
  color: #fff;
  display: flex;
  align-items: center;
  flex-direction: column;
  align-items: center;
} */

.category-area {
  box-sizing: border-box;
  /* margin-top: 1rem; */
  display: flex;
  /* align-items: center; */
  flex-direction: column;
  /* justify-content: space-between; */
  gap: 1rem;
  /* background-color: darkkhaki; */
  color: white;
  width: 98%;
}

.category-list {
  box-sizing: border-box;
  /* margin-top: 1rem; */
  display: flex;
  padding: .5rem;
  align-items: center;
  flex-direction: column;
  /* justify-content: space-between; */
  gap: .2rem;
  /* background-color: rgb(57, 122, 108); */

  background-color: rgba(0, 0, 0, 0.144);
  height: 290px;
  /* height: 90%; */
  overflow: hidden;
  overflow-y: scroll;
  width: 100%;
  border-radius: 10px;
}

.category-list::-webkit-scrollbar {
  width: 4px;
}

.category-list::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgba(255, 255, 255, 0.295);
}

.category-list::-webkit-scrollbar-track {
  border-radius: 10px;
  /* background-color: rgba(0, 0, 0, 0.295); */
}

.category-item {
  color: white;
  /* margin-top: 1rem; */
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  width: 100%;
  border-bottom: 1px solid rgba(255, 255, 255, 0.096);
  height: max-content;
  padding: .2rem;

  /* background-color: greenyellow; */
}

.category-item-actions {

  /* background-color: red; */
  display: flex;

}
</style>


