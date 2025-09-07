<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>

    <HeaderArea :title="'Add Workflow'" :icon="'workflow-plus'" />

    <div class="general-container" :style="{ gap: showTaskOptions ? 10 + 'px' : 20 + 'px' }">

      <div v-if="!isMultiple" class="input-section">
        <div class="compound-input-section">
          <input v-model="workflowName" class="input-short" type="text" placeholder="Workflow Name" v-focus
            v-return="handleEnterKey" />
        </div>
      </div>

      <BatchGenerator v-else ref="batchGen" @updateData="onUpdateWorkflows" />

      <div class="input-section drop-down-box-section">
        <DropDownBox :items="projectWorkflowNames" :selectedItem="selectedWorkflowName"
          :onSelect="changeSelectedWorkflow" />
        <DropDownBox :items="collectionStore.getCollectionTypesNames" :selectedItem="entityType" :onSelect="selectEntityType" />
      </div>

      
      <div class="horizontal-flex">
        Generate Multiple Items
        <ToggleSwitch v-tooltip="isMultiple? 'Unmark as library' : 'Mark as a library'" @click="toggleIsMultiple" :switchValueProp="isMultiple" />
      </div>

      <div class="pop-up-actions" ref="popUpActions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Confirm'" :fullWidth="true" :buttonFunction="addWorkflows" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>

    </div>
  </div>

</template>

<script setup>
// imports
import { onMounted, watchEffect, ref, computed, onUnmounted  } from 'vue';

// services
import { EntityService } from "@/../bindings/clustta/services";

// state imports
import { useTrayStates } from '@/stores/TrayStates';
import emitter from '@/lib/mitt';

// store imports
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useCollectionStore } from '@/stores/collections';
import { useProjectStore } from '@/stores/projects';
import { useMenu } from '@/stores/menu';
import { useWorkflowStore } from '@/stores/workflow';
import { useStageStore } from '@/stores/stages';
import { useIconStore } from '@/stores/icons';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import BatchGenerator from '@/instances/desktop/components/BatchGenerator.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import { WorkflowService } from '@/../bindings/clustta/services/index';

// states
const trayStates = useTrayStates();

// stores
const projectStore = useProjectStore();
const workflowStore = useWorkflowStore();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const collectionStore = useCollectionStore();
const menu = useMenu();

// vars
let showSearch = false;

// refs
const modalContainer = ref(null);
const showTaskOptions = ref(true);
const popUpActions = ref(null);
const isAwaitingResponse = ref(false);
const entityType = ref(collectionStore.getCollectionTypesNames[0]);
const selectedWorkflowName = ref(workflowStore.selectedWorkflow.name);
const workflowName = ref(workflowStore.selectedWorkflow.name);
const stageStore = useStageStore();

const isMultiple = ref(false);
const batchGen = ref(null);

// computed props
const isValueChanged = computed(() => {
  if(isMultiple.value){
    return !batchGen.value?.invalidPattern
  } else {
    return workflowName.value !== '';
  }
});

const projectWorkflows = computed(() => {
  return workflowStore.workflows;
});

const projectWorkflowNames = computed(() => {
  return projectWorkflows.value?.map(workflow => workflow.name);
})

const toggleIsMultiple = () => {
  isMultiple.value = !isMultiple.value;
}

const handleEnterKey = (event) => {
  addWorkflows();
};

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const escape = () => {
  modals.setModalVisibility('configWorkflowModal', false);
};

const closeModal = () => {
  modals.setModalVisibility("configWorkflowModal", false);
};

const changeSelectedWorkflow = (workflowName) => {
  selectedWorkflowName.value = workflowName;
  workflowStore.selectedWorkflow = projectWorkflows.value?.find((workflow) => workflow.name === workflowName);
};

const selectEntityType = (entityTypeName) => {
  entityType.value = entityTypeName;
};

const addSingleWorkflow = async () => {

  let entityId = ""
  if(stageStore.selectedItem && stageStore.selectedItem.type === 'entity'){
    entityId = stageStore.markedItems[0]
  } else if (collectionStore.navigatedCollection) {
    entityId = collectionStore.navigatedCollection.id;
  } else {
    entityId = '';
  }
  

  let entityTypeData = collectionStore.collectionTypes.find((entityTypeData) => entityTypeData.name === entityType.value)
  await WorkflowService.AddWorkflow(
    projectStore.activeProject.uri, workflowStore.selectedWorkflow.id,
    workflowName.value, entityTypeData.id, entityId
  ).then(async (data) => {
  })
    .catch((error) => {
      console.log(error)
      notificationStore.errorNotification("Error adding workflow", error)
    });
};

const workflows = ref([]);

const onUpdateWorkflows = (allWorkflows) => {
  workflows.value = allWorkflows;
  console.log(allWorkflows);
}

const addMultipleWorkflows = async () => {
  const sequenceNames = workflows.value;
  for (let sequenceName of sequenceNames) {
    workflowName.value = sequenceName;
    await addSingleWorkflow()
  }
}

const addWorkflows = async () => {
  isAwaitingResponse.value = true;
  if(isMultiple.value){
    await addMultipleWorkflows();
  } else {
    await addSingleWorkflow();
  }
  isAwaitingResponse.value = false
  emitter.emit('refresh-browser');
  closeModal();
}

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

onMounted(() => {
  trayStates.listItemsBoundary = modalContainer.value;
  trayStates.tagSearchQuery = '';
});

onUnmounted(() => {
  stageStore.markedEntities = [];
  stageStore.selectedItem = null
})


</script>

<style scoped>
@import "@/assets/desktop.css";

.compound-input-section {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .4rem;
}

.task-options-container {
  position: relative;
  box-sizing: border-box;

  width: 100%;
  height: 0px;
  /* height: 80px; */
  overflow: hidden;
  transition-property: height;
  transition-duration: 0.2s;
  transition-timing-function: ease-in-out;
  transition: opacity .5s ease-in-out;
  /* transition: all .1s ease-in-out; */
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0;
  opacity: 1;
}

.task-options-container-closed {
  transition: all .2s ease-in-out;
  opacity: 0;
  height: 0px;
  padding: 0;
  overflow: hidden;
  /* margin-bottom: -1.5rem; */
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

.pop-up-prompt {
  gap: 10px;
  /* background-color: bisque; */
  align-items: center;
  /* justify-content: center; */
  max-height: 400px;
}
</style>