<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>
    <HeaderArea :title="$t('modals.newWorkflow')" :icon="getAppIcon(collectionTypeIcon)" />

    <div class="general-container general-container-wide">
      <div class="input-section">
        <div class="input-section drop-down-box-section">
          <input v-model="workflowName" class="input-short" type="text" :placeholder="$t('placeholders.workflowName')" v-focus
            @keydown.enter="handleEnterKey" />
        </div>
      </div>

      <div class="workflow-items-container">
        <div class="workflow-item" v-for="workflow in workflowLinks">
          <WorkflowItem v-if="!isEditing(workflow.id)" @edit="editWorkflow" @delete="deleteWorkflowItem"
            :collection="workflow" :isParent="true" />
          <EditWorkflowItem v-else :isUpdate="true" :workflowItemData="workflow" @update="update" @cancel="cancel" />
        </div>

        <div class="workflow-item" v-for="workflow in workflowCollections">
          <WorkflowItem v-if="!isEditing(workflow.id)" @edit="editWorkflow" @delete="deleteWorkflowItem"
            :collection="workflow" :isParent="true" />
          <EditWorkflowItem v-else :isUpdate="true" :workflowItemData="workflow" @update="update" @cancel="cancel" />
        </div>

        <div class="workflow-item" v-for="workflow in workflowAssets">
          <WorkflowItem v-if="!isEditing(workflow.id)" @edit="editWorkflow" @delete="deleteWorkflowItem"
            :collection="workflow" :isParent="true" />
          <EditWorkflowItem v-else :isUpdate="true" :workflowItemData="workflow" @update="update" @cancel="cancel" />
        </div>

        <EditWorkflowItem v-if="isAdding && !editableWorkflowId" @confirm="confirm" @cancel="cancel" />

        <div v-else class="workflow-items-action">
          <ActionButton :label="$t('modals.addItem')" :icon="getAppIcon('plus-circle')" v-tooltip="$t('common.confirm')"
            @click="addItem()" />
        </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.close')" :fullWidth="true" :buttonFunction="closeModal" :isActive="!isAwaitingResponse" :colored="false" />
        <GeneralButton :label="isUpdate ? $t('common.update') : $t('common.create')" :fullWidth="true" @click="createWorkflow(false)"
          :isActive="isValueChanged" :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
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
const { t } = useI18n();
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
const collectionType = ref('folder type');
const isAdding = ref(false);
const isAwaitingResponse = ref(false);
const isUpdate = ref(false);
const modalContainer = ref(null);
const workflowCollections = ref([]);
const workflowIcon = ref('');
const workflowId = ref('');
const workflowLinks = ref([]);
const workflowName = ref('');
const workflowAssets = ref([]);

// computed
const collectionTypeIcon = computed(() => {
  const selectedCollectionType = collectionStore.getCollectionTypes.find((item) => item.name === collectionType.value);
  if (!selectedCollectionType) {
    return 'folder';
  }
  return selectedCollectionType.icon;
});

const isValueChanged = computed(() => {
  return workflowName.value !== ''
    && workflowId.value !== ''
    && editableWorkflowId.value === ''
    && !isAdding.value
    && !!(workflowAssets.value.length || workflowCollections.value.length || workflowLinks.value.length);
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
  if (type === 'Asset') {
    workflowAssets.value.push(workflowData);
  } else if (type === 'Collection') {
    workflowCollections.value.push(workflowData);
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
    assets: workflowAssets.value,
    collections: workflowCollections.value,
    links: workflowLinks.value
  };
  if (isUpdate.value) {
    WorkflowService.UpdateWorkflow(projectStore.activeProject.uri, workflowId.value, workflowName.value, workflowAssets.value, workflowCollections.value, workflowLinks.value)
      .then((response) => {
        workflowStore.workflows = workflowStore.workflows.filter((workflowItem) => workflowItem.id !== workflow.id);
        workflowStore.workflows.push(response);
      })
      .catch((error) => {
        console.error(error);
        notificationStore.errorNotification(t('notifications.errorUpdatingWorkflow'), error);
      });
  } else {
    WorkflowService.CreateWorkflow(projectStore.activeProject.uri, workflowName.value, workflowAssets.value, workflowCollections.value, workflowLinks.value)
      .then((response) => {
        workflowStore.workflows.push(response);
      })
      .catch((error) => {
        console.error(error);
        notificationStore.errorNotification(t('notifications.errorCreatingWorkflow'), error);
      });
  }
  isAwaitingResponse.value = false;
  closeModal();
};

// Removes a workflow item from all lists.
const deleteWorkflowItem = (workflowId) => {
  isAdding.value = false;
  editableWorkflowId.value = '';
  workflowAssets.value = workflowAssets.value.filter((item) => item.id !== workflowId);
  workflowCollections.value = workflowCollections.value.filter((item) => item.id !== workflowId);
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
  if (type === 'Asset') {
    workflowAssets.value = workflowAssets.value.filter((workflowItem) => workflowItem.id !== workflowData.id);
    workflowAssets.value.push(workflowData);
  } else if (type === 'Collection') {
    workflowCollections.value = workflowCollections.value.filter((workflowItem) => workflowItem.id !== workflowData.id);
    workflowCollections.value.push(workflowData);
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
    workflowAssets.value = selectedWorkflow.assets;
    workflowCollections.value = selectedWorkflow.collections;
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