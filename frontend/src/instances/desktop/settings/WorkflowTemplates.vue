<template>
  <div class="settings-component-root">
    <div class="settings-component-container">

      <ActionBar :itemType="'Add workflow'" :addFunction="composeWorkflow" />

      <div v-if="workflowStore.workflows.length" class="settings-component-body">
        <div class="workflow-items-container">
          <WorkflowItem v-for="workflow in workflowStore.workflows" @expand="expandWorkflowItem"
            @delete="deleteWorkflowItem" @edit="editWorkflowItem" :entity="workflow"
            :isExpanded="isExpanded(workflow.id)" :isParent="true" />
        </div>
      </div>

      <PageState v-else :message="message()" :illustration="illustration()"
        :secondaryIcon="getAppIcon('plus-circle')" />

    </div>
  </div>
</template>

<script setup>

// imports
import { onMounted, computed, ref } from 'vue';

// store imports
import { useEntityStore } from '@/stores/entity';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useIconStore } from '@/stores/icons';
import { useWorkflowStore } from '@/stores/workflow';

// components
import ActionBar from '@/instances/desktop/components/ActionBar.vue';
import WorkflowItem from '@/instances/desktop/blocks/WorkflowItem.vue';
import PageState from '@/instances/common/components/PageState.vue';

// states
const modals = useDesktopModalStore();
const iconStore = useIconStore();
const workflowStore = useWorkflowStore();

const expandedWorkflowId = ref('');

// computed props

const isExpanded = (id) => {
  return expandedWorkflowId.value === id
};

// methods
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

const editWorkflowItem = (workflowTemplateId) => {
  workflowStore.selectedWorkflow = workflowStore.workflows.find((workflowTemplate) => workflowTemplate.id === workflowTemplateId)
  composeWorkflow();
};

const deleteWorkflowItem = (workflowId) => {
  const allWorkflows = workflowStore.workflows;
  workflowStore.workflows = allWorkflows.filter((workflow) => workflow.id !== workflowId)
};

const message = () => {
  return 'You have no workflows';
};

const illustration = () => {
  return '/page-states/workflow.png';
};

const composeWorkflow = () => {
  modals.setModalVisibility('composeWorkflowModal', true);
};

// onMounted hook
onMounted(async () => {
  console.log(workflowStore.workflows);

});
</script>


<style scoped>
.input-short {
  flex: 1;
  width: 100%;
}

.settings-component-root {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 5px;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.settings-component-container {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
  width: 96%;
  gap: .5rem;
  align-items: center;
  color: var(--white);
  justify-content: space-between;
  border-radius: var(--large-radius);
  padding: 1rem;
  background-color: crimson;
  background-color: var(--black-steel);
}

.workflow-items-container {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
  height: 100%;
  overflow: hidden;
  overflow-y: scroll;
  padding-right: .5rem;
}

.workflow-items-container::-webkit-scrollbar {
  width: 6px;
}

.workflow-items-container::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--steel);
}

.workflow-items-container::-webkit-scrollbar-track {
  border-radius: 10px;
}

.settings-component-body {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
  height: 100%;
  /* background-color: royalblue; */
  overflow: hidden;
}
</style>