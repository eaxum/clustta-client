<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>
    <HeaderArea :title="'New Workflow'" :icon="getAppIcon(entityTypeIcon)" />

    <div class="general-container general-container-wide">
      <div class="input-section">
        <div class="input-section drop-down-box-section">
          <input v-model="workflowName" class="input-short" type="text" placeholder="Workflow Name" v-focus
            @keydown.enter="handleEnterKey" />
        </div>
      </div>

      <div class="workflow-items-container">
        <div class="workflow-item" v-for="workflow in workflowLinks">
          <WorkflowItem v-if="!isEditing(workflow.id)" @edit="editWorkflow" @delete="deleteWorkflowItem"
            :entity="workflow" :isParent="true" />
          <EditWorkflowItem v-else :isUpdate="true" :workflowItemData="workflow" @update="update" @cancel="cancel" />
        </div>

        <div class="workflow-item" v-for="workflow in workflowEntities">
          <WorkflowItem v-if="!isEditing(workflow.id)" @edit="editWorkflow" @delete="deleteWorkflowItem"
            :entity="workflow" :isParent="true" />
          <EditWorkflowItem v-else :isUpdate="true" :workflowItemData="workflow" @update="update" @cancel="cancel" />
        </div>

        <div class="workflow-item" v-for="workflow in workflowTasks">
          <WorkflowItem v-if="!isEditing(workflow.id)" @edit="editWorkflow" @delete="deleteWorkflowItem"
            :entity="workflow" :isParent="true" />
          <EditWorkflowItem v-else :isUpdate="true" :workflowItemData="workflow" @update="update" @cancel="cancel" />
        </div>

        <EditWorkflowItem v-if="isAdding && !editableWorkflowId" @confirm="confirm" @cancel="cancel" />

        <div v-else class="workflow-items-action">
          <ActionButton :label="'Add Item'" :icon="getAppIcon('plus-circle')" v-tooltip="'Confirm'"
            @click="addItem()" />
        </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="isUpdate ? 'Update' : 'Create'" :fullWidth="true" @click="createWorkflow(false)"
          :isActive="isValueChanged" :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref, watchEffect } from 'vue';
import { v4 as uuidv4 } from 'uuid';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import EditWorkflowItem from '@/instances/desktop/blocks/EditWorkflowItem.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import WorkflowItem from '@/instances/desktop/blocks/WorkflowItem.vue';

// services
import { WorkflowService } from "@/services";

// stores
const collectionStore = useCollectionStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const trayStates = useTrayStates();
const workflowStore = useWorkflowStore();

import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useWorkflowStore } from '@/stores/workflow';

// refs
const editableWorkflowId = ref('');
const entityType = ref('folder type');
const isAdding = ref(false);
const isAwaitingResponse = ref(false);
const isUpdate = ref(false);
const modalContainer = ref(null);
const workflowEntities = ref([]);
const workflowIcon = ref('');
const workflowId = ref('');
const workflowLinks = ref([]);
const workflowName = ref('');
const workflowTasks = ref([]);

// computed
const entityTypeIcon = computed(() => {
  const selectedEntityType = collectionStore.getCollectionTypes.find((item) => item.name === entityType.value);
  if (!selectedEntityType) {
    return 'folder';
  }
  return selectedEntityType.icon;
});

const isValueChanged = computed(() => {
  return workflowName.value !== ''
    && workflowId.value !== ''
    && editableWorkflowId.value === ''
    && !isAdding.value
    && !!(workflowTasks.value.length || workflowEntities.value.length || workflowLinks.value.length);
});

// methods
// Opens the add workflow item form.
const addItem = () => {
  isAdding.value = true;
  editableWorkflowId.value = '';
};

// Cancels the current add/edit operation.
const cancel = () => {
  isAdding.value = false;
  editableWorkflowId.value = '';
};

// Closes the compose workflow modal.
const closeModal = () => {
  modals.setModalVisibility("composeWorkflowModal", false);
};

// Adds a new workflow item to the appropriate list.
const confirm = (workflowData) => {
  isAdding.value = false;
  editableWorkflowId.value = '';

  let type = workflowData.type;
  if (type === 'Task') {
    workflowTasks.value.push(workflowData);
  } else if (type === 'Entity') {
    workflowEntities.value.push(workflowData);
  } else if (type === 'Workflow') {
    workflowLinks.value.push(workflowData);
  }
};

// Creates or updates a workflow.
const createWorkflow = async () => {
  isAwaitingResponse.value = true;
  let workflow = {
    name: workflowName.value,
    id: workflowId.value,
    icon: workflowIcon.value,
    tasks: workflowTasks.value,
    entities: workflowEntities.value,
    links: workflowLinks.value
  };
  if (isUpdate.value) {
    WorkflowService.UpdateWorkflow(projectStore.activeProject.uri, workflowId.value, workflowName.value, workflowTasks.value, workflowEntities.value, workflowLinks.value)
      .then((response) => {
        workflowStore.workflows = workflowStore.workflows.filter((workflowItem) => workflowItem.id !== workflow.id);
        workflowStore.workflows.push(response);
      })
      .catch((error) => {
        console.error(error);
        notificationStore.errorNotification('Error updating workflow', error);
      });
  } else {
    WorkflowService.CreateWorkflow(projectStore.activeProject.uri, workflowName.value, workflowTasks.value, workflowEntities.value, workflowLinks.value)
      .then((response) => {
        workflowStore.workflows.push(response);
      })
      .catch((error) => {
        console.error(error);
        notificationStore.errorNotification('Error creating workflow', error);
      });
  }
  isAwaitingResponse.value = false;
  closeModal();
};

// Removes a workflow item from all lists.
const deleteWorkflowItem = (workflowId) => {
  isAdding.value = false;
  editableWorkflowId.value = '';
  workflowTasks.value = workflowTasks.value.filter((item) => item.id !== workflowId);
  workflowEntities.value = workflowEntities.value.filter((item) => item.id !== workflowId);
  workflowLinks.value = workflowLinks.value.filter((item) => item.id !== workflowId);
};

// Toggles edit mode for a workflow item.
const editWorkflow = (workflowId) => {
  isAdding.value = false;
  if (editableWorkflowId.value === workflowId) {
    editableWorkflowId.value = '';
  } else {
    editableWorkflowId.value = workflowId;
  }
};

// Returns icon path from icon store.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles enter key press to trigger workflow creation.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    createWorkflow(false);
  }
};

// Checks if a workflow item is being edited.
const isEditing = (id) => {
  return editableWorkflowId.value === id;
};

// Updates an existing workflow item.
const update = (workflowData) => {
  isAdding.value = false;
  editableWorkflowId.value = '';

  let type = workflowData.type;
  if (type === 'Task') {
    workflowTasks.value = workflowTasks.value.filter((workflowItem) => workflowItem.id !== workflowData.id);
    workflowTasks.value.push(workflowData);
  } else if (type === 'Entity') {
    workflowEntities.value = workflowEntities.value.filter((workflowItem) => workflowItem.id !== workflowData.id);
    workflowEntities.value.push(workflowData);
  } else if (type === 'Workflow') {
    workflowLinks.value = workflowLinks.value.filter((workflowItem) => workflowItem.id !== workflowData.id);
    workflowLinks.value.push(workflowData);
  }
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// lifecycle
onMounted(() => {
  menu.clickOutsideMask = null;
  trayStates.listItemsBoundary = modalContainer.value;

  const selectedWorkflow = workflowStore.selectedWorkflow;

  if (selectedWorkflow) {
    isUpdate.value = true;
    workflowName.value = selectedWorkflow.name;
    workflowId.value = selectedWorkflow.id;
    workflowIcon.value = selectedWorkflow.icon;
    workflowTasks.value = selectedWorkflow.tasks;
    workflowEntities.value = selectedWorkflow.entities;
    workflowLinks.value = selectedWorkflow.links;
  } else {
    workflowId.value = uuidv4();
  }
});

onBeforeUnmount(() => {
  workflowStore.selectedWorkflow = null;
});
</script>


<style scoped>
@import "@/assets/desktop.css";

.workflow-items-container {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
  height: 100%;
}

.workflow-items-action {
  width: 100%;
  display: flex;
  height: 100%;
  justify-content: flex-end;
}

.general-container {
  gap: 20px;
}

.modal-container {
  justify-content: flex-start;
  align-items: flex-start;
  max-height: 90vh;
}

.general-container-wide {
  min-width: 500px !important;
}

.input-short {
  width: 100%;
}

.compound-input-section {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .4rem;
}

.pop-up-actions {
  padding: 0px;
  margin-top: 0;
}
</style>