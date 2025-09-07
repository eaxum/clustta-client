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
import { ref, computed, watchEffect, onMounted, onBeforeUnmount } from 'vue';
import utils from '@/services/utils';
import { v4 as uuidv4 } from 'uuid';

import { TaskService, FSService, WorkflowService } from "@/../bindings/clustta/services";

// state imports
import { useTrayStates } from '@/stores/TrayStates';
import { useMenu } from '@/stores/menu';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useCollectionStore } from '@/stores/collections';
import { useWorkflowStore } from '@/stores/workflow';
import { useProjectStore } from '@/stores/projects';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import WorkflowItem from '@/instances/desktop/blocks/WorkflowItem.vue';
import EditWorkflowItem from '@/instances/desktop/blocks/EditWorkflowItem.vue';

// states
const trayStates = useTrayStates();
const iconStore = useIconStore();
const collectionStore = useCollectionStore();
const workflowStore = useWorkflowStore();
const projectStore = useProjectStore();

// stores
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const menu = useMenu();

// refs
const workflowName = ref('');
const workflowId = ref('');
const workflowIcon = ref('');
const isUpdate = ref(false);

const modalContainer = ref(null);
const isAwaitingResponse = ref(false);
const isResource = ref(false);
const editableWorkflowId = ref('');
const isAdding = ref(false);
const entityType = ref('folder type');

const workflowTasks = ref([]);
const workflowEntities = ref([]);
const workflowLinks = ref([]);

// computed properties

// const isDataUnmodified = computed(() => {
//     const current = JSON.stringify(newWorkflowItemData.value);
//     const original = JSON.stringify(props.workflowItemData);
//     return current === original;
// });

const isEditing = (id) => {
  return editableWorkflowId.value === id
};

const editWorkflow = (workflowId) => {
  isAdding.value = false;
  if (editableWorkflowId.value === workflowId) {
    editableWorkflowId.value = ''
  } else {
    editableWorkflowId.value = workflowId;
  }
};

const deleteWorkflowItem = (workflowId) => {
  isAdding.value = false;
  editableWorkflowId.value = ''

  // let type = workflowData.type;
  // if (type === 'Task') {
  //   workflowTasks.value = workflowTasks.value.filter((item) => item.id !== workflowId)
  // } else if (type === 'Entity') {
  //   workflowEntities.value = workflowEntities.value.filter((item) => item.id !== workflowId)
  // } else if (type === 'Workflow') {
  //   workflowLinks.value = workflowLinks.value.filter((item) => item.id !== workflowId)
  // }
  workflowTasks.value = workflowTasks.value.filter((item) => item.id !== workflowId)
  workflowEntities.value = workflowEntities.value.filter((item) => item.id !== workflowId)
  workflowLinks.value = workflowLinks.value.filter((item) => item.id !== workflowId)
};

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

const update = (workflowData) => {

  isAdding.value = false;
  editableWorkflowId.value = '';

  let type = workflowData.type;
  if (type === 'Task') {
    workflowTasks.value = workflowTasks.value.filter((workflowItem) => workflowItem.id !== workflowData.id)
    workflowTasks.value.push(workflowData);
  } else if (type === 'Entity') {
    workflowEntities.value = workflowEntities.value.filter((workflowItem) => workflowItem.id !== workflowData.id)
    workflowEntities.value.push(workflowData);
  } else if (type === 'Workflow') {
    workflowLinks.value = workflowLinks.value.filter((workflowItem) => workflowItem.id !== workflowData.id)
    workflowLinks.value.push(workflowData);
  }
};

const cancel = () => {
  isAdding.value = false;
  editableWorkflowId.value = '';
};

const isValueChanged = computed(() => {
  return workflowName.value !== ''
    && workflowId.value !== ''
    && editableWorkflowId.value === ''
    && !isAdding.value
    //   && entityType.value !== 'folder type'
    && !!(workflowTasks.value.length || workflowEntities.value.length || workflowLinks.value.length)
});

const entityTypeIcon = computed(() => {
  const selectedEntityType = collectionStore.getCollectionTypes.find((item) => item.name === entityType.value);
  if (!selectedEntityType) {
    return 'folder'
  } else {
    return selectedEntityType.icon;
  }
});


// methods
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};


const addItem = () => {
  isAdding.value = true;
  editableWorkflowId.value = ''
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    createWorkflow(false);
  }
};

const closeModal = () => {
  modals.setModalVisibility("composeWorkflowModal", false);
};

const createWorkflow = async () => {
  isAwaitingResponse.value = true;
  let workflow = {
    name: workflowName.value,
    id: workflowId.value,
    icon: workflowIcon.value,
    tasks: workflowTasks.value,
    entities: workflowEntities.value,
    links: workflowLinks.value
  }
  if (isUpdate.value) {
    WorkflowService.UpdateWorkflow(projectStore.activeProject.uri, workflowId.value, workflowName.value, workflowTasks.value, workflowEntities.value, workflowLinks.value)
      .then((response) => {
        workflowStore.workflows = workflowStore.workflows.filter((workflowItem) => workflowItem.id !== workflow.id)
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

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// onMounted hook
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

.workflow-item-edit {
  background-color: var(--dark-steel);
  border-radius: var(--normal-radius);
  padding: .5rem;
  outline: var(--transparent-line);
  outline-offset: -1px;
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
  /* overflow: hidden;
  width: 40vw;
  max-width: 50vw;
  max-height: 80vh; */
  min-width: 500px !important;
}

.task-options-container {
  position: relative;
  box-sizing: border-box;
  width: 100%;
  height: max-content;
  height: 60px;
  transition: all .2s ease-in-out;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0;
  /* background-color: chocolate; */
}

.task-options-container-closed {
  height: 0px;
  padding: 0;
  margin-bottom: -1.5rem;
}

.input-short {
  width: 100%;
}

.listbox-short {
  width: 130px;
}

.input-label {
  font-family: Inter, sans-serif;
  color: white;
  font-size: 16px;
  white-space: nowrap;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 1rem;

}

.compound-input-section {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .4rem;
}

.pop-up-prompt {
  gap: 10px;
  align-items: center;
  justify-content: space-between;
}


.pop-up-actions {
  padding: 0px;
  margin-top: 0;
}
</style>