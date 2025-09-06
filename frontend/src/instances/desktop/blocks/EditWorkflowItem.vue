<template>
    <div class="workflow-item">
        <div class="input-section">

            <div class="input-section drop-down-box-section">
                <input v-model="workflowName" class="input-short" type="text" placeholder="Workflow item Name" v-focus
                    @keydown.enter="handleEnterKey" />
                <ActionButton :isDisabled="isWorkflowItemModified" :icon="getAppIcon('check')" v-tooltip="'Confirm'"
                    @click="confirm()" />
                <ActionButton :icon="getAppIcon('close')" v-tooltip="'Cancel'" @click="cancel()" />
            </div>
            <div class="input-section drop-down-box-section">
                <DropDownBox :items="itemTypes" :selectedItem="itemType" :onSelect="changeItemType" />
                <DropDownBox v-if="itemType === 'Task'" :items="taskTypeNames" :selectedItem="taskType"
                    :onSelect="selectTaskType" />
                <DropDownBox v-else-if="itemType === 'Collection'" :items="entityTypeNames" :selectedItem="entityType"
                    :onSelect="selectEntityType" />
                <DropDownBox v-else-if="itemType === 'Workflow'" :items="projectWorkflowNames"
                    :selectedItem="selectedWorkflowName" :onSelect="selectWorkflow" />
                <DropDownBox v-if="itemType === 'Workflow'" :items="entityTypeNames" :selectedItem="entityType"
                    :onSelect="selectEntityType" />
            </div>
            <div v-if="itemType === 'Task'" class="task-options-container">
                <div class="input-section">
                    <Apps @templateSelected="selectTemplate" :selectedTemplateId="taskTemplateId" />
                </div>
            </div>

        </div>
    </div>
</template>

<script setup>


// imports
import { computed, ref, onMounted, onBeforeUnmount, reactive, watch } from 'vue';
import utils from '@/services/utils';
import { v4 as uuidv4 } from 'uuid';

// store imports
import { useTaskStore } from '@/stores/task';
import { useEntityStore } from '@/stores/entity';
import { useTemplateStore } from '@/stores/template';
import { useWorkflowStore } from '@/stores/workflow';
import { useIconStore } from '@/stores/icons';

// states imports
const taskStore = useTaskStore();
const entityStore = useEntityStore();
const templateStore = useTemplateStore();
const workflowStore = useWorkflowStore();
const iconStore = useIconStore();

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import Apps from '@/instances/common/components/Apps.vue';

// emits
const emit = defineEmits(['update', 'confirm', 'cancel']);

// props
const props = defineProps({
    index: Number,
    workflowItemData: { type: Object, default: {} },
    isUpdate: { type: Boolean, default: false }
});

// refs
const itemType = ref('select type');

const workflowId = ref('');
const workflowName = ref('');

const taskTemplateId = ref('');
const taskTypeId = ref('');
const taskTypeIcon = ref('');

const entityTypeId = ref('');
const entityTypeIcon = ref('');

const workflowTemplateId = ref('');

// computed
const taskType = computed(() => {
    const allTaskTypes = taskStore.getTaskTypes;
    const selectedTaskType = allTaskTypes.find((item) => item.id === taskTypeId.value);
    return selectedTaskType ? selectedTaskType.name : 'Select task type'
});

const entityType = computed(() => {
    const allEntityTypes = entityStore.getEntityTypes;
    const selectedEntityType = allEntityTypes.find((item) => item.id === entityTypeId.value);
    return selectedEntityType ? selectedEntityType.name : 'Select collection type'
});

const projectWorkflows = computed(() => {

    let selectedWorkflowId;
    let workflowsWithMatchingItemId;
    let unavailableWorkflowIds = []
    if (workflowStore.selectedWorkflow) {

        selectedWorkflowId = workflowStore.selectedWorkflow.id;
        workflowsWithMatchingItemId = workflowStore.workflows
            .filter(workflow => workflow.items?.some(item => item.workflow_template_id === selectedWorkflowId))
            .map(workflow => workflow.id);

        unavailableWorkflowIds = [selectedWorkflowId, ...workflowsWithMatchingItemId]
    };

    const availableWorkflows = workflowStore.workflows.filter((workflow) => !unavailableWorkflowIds.includes(workflow.id))
    return availableWorkflows;
});

const projectWorkflowNames = computed(() => {
    return projectWorkflows.value?.map((item) => item.name);
});

const selectedWorkflowName = ref(projectWorkflowNames.value[0]);

const taskTypeNames = computed(() => {
    return taskStore.getTaskTypesNames;
});

const taskTemplates = computed(() => {
    return templateStore.getTemplates;
});

const entityTypeNames = computed(() => {
    return entityStore.getEntityTypesNames;
});

const itemTypes = computed(() => {
    let allItemTypes;
    if (projectWorkflows.value.length) {
        allItemTypes = ['Task', 'Collection', 'Workflow'];
    } else {
        allItemTypes = ['Task', 'Collection'];
    }
    return allItemTypes.filter((item) => item !== itemType.value?.toLowerCase());
});

const newWorkflowItemData = computed(() => {
    const itemTypeName = itemType.value;
    let data = {};
    if (itemTypeName === 'Task') {

        const allTaskTypes = taskStore.getTaskTypes;
        const firstTaskType = allTaskTypes[0];
        taskTypeIcon.value = taskTypeIcon.value ? taskTypeIcon.value : firstTaskType.icon;
        taskTypeId.value = taskTypeId.value ? taskTypeId.value : firstTaskType.id;

        const allTaskTemplates = templateStore.getTemplates;
        const firstTaskTemplate = allTaskTemplates[0];
        taskTemplateId.value = taskTemplateId.value ? taskTemplateId.value : firstTaskTemplate.id;

        data = {
            id: workflowId.value,
            name: workflowName.value,
            template_id: taskTemplateId.value,
            task_type_id: taskTypeId.value,
            task_type_icon: taskTypeIcon.value,
            type: 'Task',
        };
    } else if (itemTypeName === 'Collection') {

        const allEntityTypes = entityStore.getEntityTypes;
        const firstEntityType = allEntityTypes[0];
        entityTypeIcon.value = entityTypeIcon.value ? entityTypeIcon.value : firstEntityType.icon;
        entityTypeId.value = entityTypeId.value ? entityTypeId.value : firstEntityType.id;

        data = {
            id: workflowId.value,
            name: workflowName.value,
            entity_type_id: entityTypeId.value,
            entity_type_icon: entityTypeIcon.value,
            type: 'Entity',
        };
    } else if (itemTypeName === 'Workflow') {

        const allWorkflowTemplates = projectWorkflows.value;
        const firstWorkflowTemplate = allWorkflowTemplates[0];
        workflowTemplateId.value = workflowTemplateId.value ? workflowTemplateId.value : firstWorkflowTemplate.id;
        // workflowName.value = selectedWorkflowName.value;
        data = {
            id: workflowId.value,
            name: workflowName.value,
            entity_type_id: entityTypeId.value,
            workflow_id: workflowId.value,
            linked_workflow_id: workflowTemplateId.value,
            type: 'Workflow',
        };

    } else {
        data = {
            id: workflowId.value,
            name: workflowName.value,
            workflow_template_id: workflowTemplateId.value,
        };
    }
    return data
});

const isDataInComplete = computed(() => {
    const data = newWorkflowItemData.value;
    return Object.values(data).some(value => value === '');
});

const isDataUnmodified = computed(() => {
    const current = JSON.stringify(newWorkflowItemData.value);
    const original = JSON.stringify(props.workflowItemData);
    return current === original;
});

const isWorkflowItemModified = computed(() => {
    if (itemType.value === 'select type') {
        return true
    }

    if (props.isUpdate) {
        return isDataUnmodified.value
    } else {
        return isDataInComplete.value
    }
});

// methods
const confirm = () => {
    if (props.isUpdate) {
        emit('update', newWorkflowItemData.value);
    } else {
        emit('confirm', newWorkflowItemData.value);
    }
};

const cancel = () => {
    emit('cancel');
};

const getAppIcon = (iconName) => {
    const icon = iconStore.getAppIcon(iconName);
    return icon
};

const selectTaskType = (taskTypeName) => {
    const allTaskTypes = taskStore.getTaskTypes;
    const selectedTaskType = allTaskTypes.find((item) => item.name === taskTypeName);

    taskTypeId.value = selectedTaskType.id;
    taskTypeIcon.value = selectedTaskType.icon;
};

const selectEntityType = (entityTypeName) => {
    const allEntityTypes = entityStore.getEntityTypes;
    const selectedEntityType = allEntityTypes.find((item) => item.name === entityTypeName);

    entityTypeId.value = selectedEntityType.id;
    entityTypeIcon.value = selectedEntityType.icon;
};

const selectTemplate = (taskTemplate) => {
    taskTemplateId.value = taskTemplate.id;
};

const selectWorkflow = (workflowName) => {
    selectedWorkflowName.value = workflowName;
    const allWorkflows = projectWorkflows.value;
    const selectedWorkflow = allWorkflows.find((item) => item.name === workflowName);

    workflowTemplateId.value = selectedWorkflow.id;
};

const changeItemType = (newItemTypeName) => {
    itemType.value = newItemTypeName;
    if (newItemTypeName === 'Task') {
        taskTemplateId.value = taskTemplates.value[0]?.id;
    }
};

onMounted(() => {
    if (props.isUpdate) {

        workflowId.value = props.workflowItemData.id || '';
        workflowName.value = props.workflowItemData.name || '';

        if (props.workflowItemData.task_type_id) {
            itemType.value = 'Task';
            taskTypeId.value = props.workflowItemData.task_type_id;
            taskTypeIcon.value = props.workflowItemData.task_type_icon || '';
            taskTemplateId.value = props.workflowItemData.template_id || '';
        } else if (props.workflowItemData.entity_type_id) {
            itemType.value = 'Collection';
            entityTypeId.value = props.workflowItemData.entity_type_id;
            entityTypeIcon.value = props.workflowItemData.entity_type_icon || '';
        } else if (props.workflowItemData.workflow_template_id) {
            itemType.value = 'Workflow';
            workflowTemplateId.value = props.workflowItemData.workflow_template_id;
            const workflow = projectWorkflows.value.find(wf => wf.id === workflowTemplateId.value);
            if (workflow) {
                selectedWorkflowName.value = workflow.name;
            }
        }
    } else {
        workflowId.value = uuidv4();
    }
});

onBeforeUnmount(() => {
});

</script>

<style scoped>
@import "@/assets/desktop.css";

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

.input-short {
    width: 100%;
}

.uneditable {
    /* background-color: crimson; */
    /* pointer-events: none; */
    cursor: not-allowed;
    opacity: .5;
    font-style: italic;
}

.workflow-item {
    box-sizing: border-box;
    background-color: var(--dark-steel);
    border-radius: var(--normal-radius);
    padding: .5rem;
    outline: var(--transparent-line);
    outline-offset: -1px;
}
</style>