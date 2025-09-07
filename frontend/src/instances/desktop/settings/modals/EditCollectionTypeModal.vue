<template>

  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">
    <!-- <HeaderArea :title="title" :icon="'folder'" :showSearch="showSearch" /> -->

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
        <GeneralButton :label="'Update'" :fullWidth="true" @click="updateEntityType" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>


    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { EntityService } from "@/../bindings/clustta/services";
import { useCollectionStore } from '@/stores/collections';
import { useProjectStore } from '@/stores/projects';
import iconData from "@/data/iconData.json";

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import IconGrid from '@/instances/desktop/components/IconGrid.vue';

const icons = computed(() => {
  const allIcons = iconData.icons;
  const allEntityTypeIcons = collectionStore.getCollectionTypes.map((item) => item.icon);
  return allIcons.filter((icon) => !allEntityTypeIcons.includes(icon))
})

// vars
let title = 'Edit Collection type';
const isAwaitingResponse = ref(false);

// stores
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const collectionStore = useCollectionStore();
const projectStore = useProjectStore();

const displayIconSelector = ref(true);

const entityTypeName = ref('');
const entityTypeIcon = ref('');

const setIcon = (icon) => {
  entityTypeIcon.value = icon;
};

const isValueChanged = computed(() => {
  return !!entityTypeName.value && entityTypeIcon.value !== 'generic'
});

const closeModal = () => {
  modals.setModalVisibility("editEntityTypeModal", false);
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    // updateEntityType();
  }
};

const updateEntityType = () => {
  EntityService.UpdateEntityType(projectStore.activeProject.uri, collectionStore.selectedCollectionType.id, entityTypeName.value, entityTypeIcon.value)
    .then((response) => {
      notificationStore.addNotification("Collection type Updated", "", "success");
      const index = collectionStore.collectionTypes.findIndex(entityType => entityType.id === collectionStore.selectedCollectionType.id);
      collectionStore.collectionTypes[index] = response
      closeModal();
    })
    .catch((error) => {
      notificationStore.errorNotification("Error updating folder Type", error);
    });
};

onMounted(() => {
  entityTypeName.value = collectionStore.selectedCollectionType.name;
  entityTypeIcon.value = collectionStore.selectedCollectionType.icon;
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