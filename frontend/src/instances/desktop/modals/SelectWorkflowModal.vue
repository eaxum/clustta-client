<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>

    <HeaderArea :title="title" :icon="'workflow-arrow'" :showSearch="showSearch" />

    <div class="general-container general-container-wide" :style="{ gap: showTaskOptions ? 10 + 'px' : 20 + 'px' }">

      <div v-if="!workflowStore.workflows.length" class="page-state-container">
        <PageState :message="'This project has no Workflow templates'" :illustration="'/page-states/workflow.png'" />
      </div>

      <div v-else class="workflow-template-list">
        <WorkflowItem v-for="workflow in workflowStore.workflows" @expand="expandWorkflowItem"
             :entity="workflow" @select="selectWorkflowTemplate" :selectable="true"
            :isExpanded="isExpanded(workflow.id)" :isParent="true" />
      </div>

      <div class="pop-up-actions" ref="popUpActions">
        <ActionButton v-if="userStore.canDo('create_template')" :icon="getAppIcon('workflow-plus')" :label="'Manage workflows'"
          :buttonFunction="manageTemplates" />
        <!-- <GeneralButton :label="'Cancel'" :fullWidth="false" :buttonFunction="closeModal" :colored="false" /> -->
      </div>

    </div>

  </div>

</template>

<script setup>
// imports
import { onMounted, onUnmounted, ref } from 'vue';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import PageState from '@/instances/common/components/PageState.vue';
import WorkflowItem from '@/instances/desktop/blocks/WorkflowItem.vue';

// stores
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const settings = useSettingsStore();
const stage = useStageStore();
const userStore = useUserStore();
const workflowStore = useWorkflowStore();

import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useSettingsStore } from '@/stores/settings';
import { useStageStore } from '@/stores/stages';
import { useUserStore } from '@/stores/users';
import { useWorkflowStore } from '@/stores/workflow';

// constants
const showSearch = false;
const title = 'Select Workflow template';

// refs
const expandedWorkflowId = ref('');
const modalContainer = ref(null);
const popUpActions = ref(null);
const showTaskOptions = ref(true);

// methods
// Closes the select workflow modal.
const closeModal = () => {
  modals.setModalVisibility("selectWorkflowModal", false);
};

// Toggles workflow item expansion.
const expandWorkflowItem = (workflowId) => {
  if (expandedWorkflowId.value === workflowId) {
    expandedWorkflowId.value = '';
  } else {
    expandedWorkflowId.value = workflowId;
  }
};

// Returns icon path from icon store.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Checks if a workflow item is expanded.
const isExpanded = (id) => {
  return expandedWorkflowId.value === id;
};

// Opens project settings to manage workflows.
const manageTemplates = () => {
  modals.disableAllModals();
  settings.activeModalName = 'Workflows';
  stage.setStageVisibility('projectSettings', true);
  settings.setModalVisibility('workflows', true);
};

// Selects a workflow template and opens config modal.
const selectWorkflowTemplate = async (workflowId) => {
  const selectedWorkflow = workflowStore.workflows.find((workflow) => workflow.id === workflowId);
  workflowStore.selectedWorkflow = selectedWorkflow;
  console.log(workflowStore.selectedWorkflow);
  closeModal();
  modals.setModalVisibility('configWorkflowModal', true);
};

// lifecycle
onMounted(async () => {
});

onUnmounted(() => {
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.pop-up-actions {
  justify-content: center;
}

.modal-container {
  justify-content: flex-start;
  align-items: flex-start;
  max-height: 90vh;
}

.general-container-wide {
  overflow: hidden;
  max-width: 50vw;
  max-height: 80vh;
}

.page-state-container {
  height: 300px;
}

.workflow-template-list {
  width: 100%;
  padding: .2rem;
  overflow: hidden;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  box-sizing: border-box;
  gap: 10px;
}
</style>

