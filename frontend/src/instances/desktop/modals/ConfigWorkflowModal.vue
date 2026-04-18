<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>

    <HeaderArea :title="$t('modals.addWorkflow')" :icon="headerIcon" />
    <div class="general-container" :style="{ gap: showAssetOptions ? 10 + 'px' : 20 + 'px' }">

      <div v-if="!isMultiple" class="input-section">
        <div class="compound-input-section">
          <input v-model="workflowName" class="input-short" type="text" :placeholder="$t('placeholders.workflowName')" v-focus
            v-return="handleEnterKey" />
        </div>
      </div>

      <BatchGenerator v-else ref="batchGen" @updateData="onUpdateWorkflows" />

      <div class="input-section drop-down-box-section">
        <DropDownBox :items="projectWorkflowNames" :selectedItem="selectedWorkflowName"
          :onSelect="changeSelectedWorkflow" />
        <DropDownBox :items="collectionStore.getCollectionTypesNames" :selectedItem="collectionType" :onSelect="selectCollectionType" />
      </div>

      
      <div class="horizontal-flex">
        {{ $t('modals.generateMultipleItems') }}
        <ToggleSwitch v-tooltip="isMultiple? 'Disable batch mode' : 'Enable batch mode'" @click="toggleIsMultiple" :switchValueProp="isMultiple" />
      </div>

      <div class="pop-up-actions" ref="popUpActions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.confirm')" :fullWidth="true" :buttonFunction="addWorkflows" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>

    </div>
  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, onUnmounted, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';

// components
import BatchGenerator from '@/instances/desktop/components/BatchGenerator.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// services
import { CollectionService, WorkflowService } from "@/services";

// stores
const { t } = useI18n();
const collectionStore = useCollectionStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stageStore = useStageStore();
const trayStates = useTrayStates();
const workflowStore = useWorkflowStore();

import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTrayStates } from '@/stores/TrayStates';
import { useWorkflowStore } from '@/stores/workflow';

// refs
const batchGen = ref(null);
const collectionType = ref(collectionStore.getCollectionTypesNames[0]);
const isAwaitingResponse = ref(false);
const isMultiple = ref(false);
const modalContainer = ref(null);
const popUpActions = ref(null);
const selectedWorkflowName = ref(workflowStore.selectedWorkflow.name);
const showAssetOptions = ref(true);
const workflowName = ref(workflowStore.selectedWorkflow.name);
const workflows = ref([]);

// computed
const collectionId = computed(() => {
  if (stageStore.selectedItem && stageStore.selectedItem.type === 'collection') {
    return stageStore.selectedItem?.id;
  } else if (collectionStore.navigatedCollection) {
    return collectionStore.navigatedCollection.id;
  }
  return '';
});

const headerIcon = computed(() => {
  const selectedType = collectionStore.collectionTypes.find(item => item.name === collectionType.value);
  return selectedType?.icon || 'workflow-plus';
});

const isValueChanged = computed(() => {
  if (isMultiple.value) {
    return !batchGen.value?.invalidPattern;
  }
  return workflowName.value !== '';
});

const projectWorkflows = computed(() => {
  return workflowStore.workflows;
});

const projectWorkflowNames = computed(() => {
  return projectWorkflows.value?.map(workflow => workflow.name);
});

// methods
// Adds multiple workflows using batch generation.
const addMultipleWorkflows = async () => {
  const sequenceNames = workflows.value;
  for (let sequenceName of sequenceNames) {
    workflowName.value = sequenceName;
    await addSingleWorkflow();
  }
};

// Adds a single workflow instance.
const addSingleWorkflow = async () => {
  let collectionTypeData = collectionStore.collectionTypes.find((collectionTypeData) => collectionTypeData.name === collectionType.value);
  await WorkflowService.AddWorkflow(
    projectStore.activeProject.uri, workflowStore.selectedWorkflow.id,
    workflowName.value, collectionTypeData.id, collectionId.value
  ).then(async (data) => {
  }).catch((error) => {
    console.log(error);
    notificationStore.errorNotification(t('notifications.errorAddingWorkflow'), error);
  });
};

// Adds workflows (single or multiple based on mode).
const addWorkflows = async () => {
  isAwaitingResponse.value = true;
  if (isMultiple.value) {
    await addMultipleWorkflows();
  } else {
    await addSingleWorkflow();
  }
  isAwaitingResponse.value = false;
  emitter.emit('refresh-browser');
  closeModal();
};

// Changes the selected workflow template.
const changeSelectedWorkflow = (workflowName) => {
  selectedWorkflowName.value = workflowName;
  workflowStore.selectedWorkflow = projectWorkflows.value?.find((workflow) => workflow.name === workflowName);
};

// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility("configWorkflowModal", false);
};

// Returns icon path from icon store.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles enter key press to trigger workflow addition.
const handleEnterKey = (event) => {
  addWorkflows();
};

// Updates workflows from batch generator.
const onUpdateWorkflows = (allWorkflows) => {
  workflows.value = allWorkflows;
  console.log(allWorkflows);
};

// Selects an collection type.
const selectCollectionType = (collectionTypeName) => {
  collectionType.value = collectionTypeName;
};

// Toggles multiple workflow mode.
const toggleIsMultiple = () => {
  isMultiple.value = !isMultiple.value;
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// lifecycle
onMounted(() => {
  trayStates.listItemsBoundary = modalContainer.value;
  trayStates.tagSearchQuery = '';
});

onUnmounted(() => {
  stageStore.markedCollections = [];
  stageStore.selectedItem = null;
});

onBeforeUnmount(() => {
  workflowStore.selectedWorkflow = null;
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.compound-input-section {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .4rem;
}

.input-short {
  flex: 1;
  width: 100%;
}
</style>