<template>
  <div class="settings-component-root">
    <div class="settings-component-container">

      <ActionBar :itemType="'Add Asset type'" :addFunction="addTaskType" />

      <ScrollList v-if="projectTaskTypes.length" :items="projectTaskTypes" :useIcons="true" :useItemId="true"
        :wrapItems="true" :editItems="true" :editListItem="prepEditTaskType" :deleteItems="true"
        :deleteListItem="deleteTaskType" />

      <PageState v-else :message="message()" :illustration="illustration()" :secondaryIcon="getAppIcon('plus-circle')"
        :secondaryActionMessage="secondaryActionMessage()" :secondaryActionFunction="secondaryActionFunction" />

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

// imports
import { onMounted, computed, ref } from 'vue';
import utils from '@/services/utils';

// services
import { TaskService } from "@/../bindings/clustta/services";

// store imports
import { useAssetStore } from '@/stores/assets';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useProjectStore } from '@/stores/projects';


// components
import ScrollList from '@/instances/desktop/components/ScrollList.vue';
import ActionBar from '@/instances/desktop/components/ActionBar.vue';
import PageState from '@/instances/common/components/PageState.vue';

// states
const assetStore = useAssetStore();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();

const projectTaskTypes = computed(() => {
  
  let taskTypes = assetStore.getTaskTypes;
  console.log(taskTypes)
  let viewTaskTypeIds = [];
  let tasks = assetStore.tasks;

  for (const task of tasks){
    let taskTypeId = task.task_type_id;
    if(!viewTaskTypeIds.includes(taskTypeId)){
      viewTaskTypeIds.push(taskTypeId)
    }
  }

  const allTypes =  taskTypes.map(type => ({
    ...type,
    can_delete: !viewTaskTypeIds.includes(type.id),
    can_edit: type.name !== 'generic',
  }))

  console.log(allTypes)

return allTypes

});

// methods
const message = () => {
  return 'You have no task types';
};

const illustration = () => {
  return '/page-states/resources.png';
};

const secondaryActionMessage = () => {
  return 'Add Task type'
};

const secondaryActionFunction = () => {
  addTaskType();
};

const addTaskType = () => {
  modals.setModalVisibility('addAssetTypeModal', true);
};

const prepEditTaskType = (selectedTaskTypeId) => {
  console.log(selectedTaskTypeId)
  assetStore.selectedTaskType = assetStore.getTaskTypes.find((item) => item.id === selectedTaskTypeId)
  modals.setModalVisibility('editAssetTypeModal', true);

};

const replaceSymbols = (name) => {
  return name.replace(/_/g, " ").toLowerCase().replace(/(^\w|\s\w)/g, match => match.toUpperCase());
};

const deleteTaskType = async (taskTypeId) => {
  TaskService.DeleteTaskType(projectStore.activeProject.uri, taskTypeId)
    .then((response) => {
      notificationStore.addNotification("Task Type Deleted", "", "success");
      const index = assetStore.taskTypes.findIndex(taskType => taskType.id === taskTypeId);
      assetStore.taskTypes.splice(index, 1);
    })
    .catch((error) => {
      notificationStore.errorNotification("Error Deleting Task Type", error);
    });
};

// onMounted hook
onMounted(async () => {

});
</script>


<style scoped>
.input-short {
  flex: 1;
  width: 100%;
}

.settings-component-root{
  width:100%;
  height:100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 5px;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.settings-component-container{
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
  width: 96%;
  gap: .5rem;
  align-items: center;
  color: white;
  justify-content: space-between;
  border-radius: var(--large-radius);
  padding: 1rem;
  background-color: crimson;
  background-color: var(--black-steel);
}
</style>


