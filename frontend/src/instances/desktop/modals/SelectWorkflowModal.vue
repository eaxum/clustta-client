<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>

    <HeaderArea :title="title" :icon="'workflow-arrow'" :showSearch="showSearch" />

    <div class="general-container general-container-wide" :style="{ gap: showTaskOptions ? 10 + 'px' : 20 + 'px' }">

      <div class="workflow-template-list">
        <!-- <WorkflowTemplate v-for="workflowTemplate in workflowTemplates" :workflowTemplate="workflowTemplate" /> -->
        
        <WorkflowItem v-for="workflow in workflowStore.workflows" @expand="expandWorkflowItem"
             :entity="workflow" @select="selectWorkflowTemplate" :selectable="true"
            :isExpanded="isExpanded(workflow.id)" :isParent="true" />
      </div>

      <div class="pop-up-actions" ref="popUpActions">
        <ActionButton v-if="userStore.canDo('create_template')" :icon="getAppIcon('squares-plus')" :label="'Manage workflows'"
          :buttonFunction="manageTemplates" />
        <!-- <GeneralButton :label="'Cancel'" :fullWidth="false" :buttonFunction="closeModal" :colored="false" /> -->
      </div>

    </div>

  </div>

</template>

<script setup>

// imports
import { onMounted, onUnmounted, ref, computed } from 'vue';


// store imports
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useDataStore } from '@/stores/data';
import { useWorkflowStore } from '@/stores/workflow';
import { useStageStore } from '@/stores/stages';
import { useUserStore } from '@/stores/users';
import { useSettingsStore } from '@/stores/settings';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import WorkflowItem from '@/instances/desktop/blocks/WorkflowItem.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue'

// stores
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const workflowStore = useWorkflowStore();
const stage = useStageStore();
const userStore = useUserStore();
const settings = useSettingsStore();

// vars
let showSearch = false;
let title = 'Select Workflow template';


// refs
const modalContainer = ref(null);
const showTaskOptions = ref(true);
const popUpActions = ref(null);
const expandedWorkflowId = ref('');

// methods

const isExpanded = (id) => {
  return expandedWorkflowId.value === id
};

// methods
const manageTemplates = () => {
  modals.disableAllModals();
  settings.activeModalName = 'Workflows';
  stage.setStageVisibility('projectSettings', true);
	settings.setModalVisibility('workflows', true);
};

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const expandWorkflowItem = (workflowId) => {
  if (expandedWorkflowId.value === workflowId) {
    expandedWorkflowId.value = ''
  } else {
    expandedWorkflowId.value = workflowId;
  }
};

const selectWorkflowTemplate = async (workflowId) => {
  const selectedWorkflow = workflowStore.workflows.find((workflow) => workflow.id === workflowId)
  workflowStore.selectedWorkflow = selectedWorkflow;
  console.log(workflowStore.selectedWorkflow);
  closeModal();
  modals.setModalVisibility('configWorkflowModal', true);
  return
};

const handleEnterKey = (event) => {
};

const escape = () => {
  modals.setModalVisibility('selectWorkflowModal', false);
};

const closeModal = () => {
  modals.setModalVisibility("selectWorkflowModal", false);
};

onMounted(async () => {
});

onUnmounted(() => {
})


</script>

<style scoped>
@import "@/assets/desktop.css";

.pop-up-actions {
  justify-content: flex-end;
  justify-content: center;
}

.modal-container {
  justify-content: flex-start;
  align-items: flex-start;
  max-height: 90vh;
}

.general-container-wide {
  /* background-color: firebrick; */
  overflow: hidden;
  /* width: 40vw; */
  /* min-width: 600px !important; */
  max-width: 50vw;
  max-height: 80vh;

}

.folder-path {
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
}

.rules-toggle {
  display: flex;
  /* background-color: red; */
  gap: .5rem;
  align-items: center;
  min-width: max-content
}

.workflow-template-list {
  /* background-color: crimson; */
  width: 100%;
  padding: .2rem;
  overflow: hidden;
  display: flex;
  align-items: center;
  /* gap: 10px; */
  flex-wrap: wrap;
  /* padding: 10px 20px; */
  box-sizing: border-box;
  gap: 10px;
}

.workflow-template-list-container {
  display: flex;
  /* background-color: firebrick; */
  width: 100%;
  gap: .2rem;
  overflow: hidden;
  display: flex;
  align-items: center;
}

.regex-instances {
  width: 100%;
  display: flex;
  max-height: 50vh;
  flex-direction: column;
  gap: 10px;
  /* background-color: green; */
  /* padding-right: 5px; */
  overflow: hidden;
  /* overflow-y: scroll; */
  box-sizing: border-box;
}

.regex-instances-scroll {
  box-sizing: border-box;
  width: 100%;
  display: flex;
  height: min-content;
  flex-direction: column;
  gap: 10px;
  /* background-color: green; */
}

.regex-instances::-webkit-scrollbar {
  width: 8px;
}

.regex-instances::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--light-steel);
}

.regex-instances::-webkit-scrollbar-track {
  border-radius: 10px;
}

.attachment-area {
  box-sizing: border-box;
  /* margin-top: 1rem; */
  display: flex;
  align-items: center;
  flex-direction: column;
  /* justify-content: space-between; */
  gap: 1rem;
  /* background-color: darkkhaki; */
  width: 98%;
}

.attachment-list {
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
  max-height: 150px;
  overflow: hidden;
  overflow-y: scroll;
  width: 100%;
  border-radius: 10px;
}

.attachment-list::-webkit-scrollbar {
  width: 4px;
}

.attachment-list::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgba(255, 255, 255, 0.295);
}

.attachment-list::-webkit-scrollbar-track {
  border-radius: 10px;
  /* background-color: rgba(0, 0, 0, 0.295); */
}

.attachment {
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

