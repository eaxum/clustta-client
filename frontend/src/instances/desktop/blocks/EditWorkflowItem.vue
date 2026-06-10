<template>
    <div class="workflow-item">
        <div class="input-section">

            <div class="input-section drop-down-box-section">
                <input v-model="workflowName" class="input-short" type="text" :placeholder="$t('placeholders.workflowItemName')" v-focus
                    @keydown.enter="handleEnterKey" />
                <ActionButton :isDisabled="isWorkflowItemModified" :icon="getAppIcon('check')" v-tooltip="$t('common.confirm')"
                    @click="confirm()" />
                <ActionButton :icon="getAppIcon('close')" v-tooltip="$t('common.cancel')" @click="cancel()" />
            </div>
            <div class="input-section drop-down-box-section">
                <DropDownBox :items="itemTypes" :selectedItem="itemTypeLabel" :onSelect="changeItemType" />
                <DropDownBox v-if="itemType === 'Asset'" :items="assetTypeOptions" :selectedItem="assetType"
                    :onSelect="selectAssetType" />
                <DropDownBox v-else-if="itemType === 'Collection'" :items="collectionTypeOptions" :selectedItem="collectionType"
                    :onSelect="selectCollectionType" />
                <DropDownBox v-else-if="itemType === 'Workflow'" :items="projectWorkflowNames"
                    :selectedItem="selectedWorkflowName" :onSelect="selectWorkflow" />
                <DropDownBox v-if="itemType === 'Workflow'" :items="collectionTypeOptions" :selectedItem="collectionType"
                    :onSelect="selectCollectionType" />
            </div>
            <div v-if="itemType === 'Asset'" class="asset-options-container">
                <div class="input-section">
                    <Apps @templateSelected="selectTemplate" :selectedTemplateId="assetTemplateId" />
                </div>
            </div>

        </div>
    </div>
</template>

<script setup>


// imports
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import utils from '@/services/utils';
import { v4 as uuidv4 } from 'uuid';
import { useI18n } from 'vue-i18n';

// store imports
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useNotificationStore } from '@/stores/notifications';
import { useTemplateStore } from '@/stores/template';
import { useWorkflowStore } from '@/stores/workflow';
import { useIconStore } from '@/stores/icons';

// states imports
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const notificationStore = useNotificationStore();
const templateStore = useTemplateStore();
const workflowStore = useWorkflowStore();
const iconStore = useIconStore();

const { t } = useI18n();

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
const itemType = ref('selectType');

const workflowId = ref('');
const workflowName = ref('');

const assetTemplateId = ref('');
const assetTypeId = ref('');
const assetTypeIcon = ref('');

const collectionTypeId = ref('');
const collectionTypeIcon = ref('');

const workflowTemplateId = ref('');

// computed
const assetType = computed(() => {
    const allAssetTypes = assetStore.getAssetTypes;
    const selectedAssetType = allAssetTypes.find((item) => item.id === assetTypeId.value);
    return selectedAssetType ? selectedAssetType.name : t('blocks.selectAssetType')
});

const collectionType = computed(() => {
    const allCollectionTypes = collectionStore.getCollectionTypes;
    const selectedCollectionType = allCollectionTypes.find((item) => item.id === collectionTypeId.value);
    return selectedCollectionType ? selectedCollectionType.name : t('blocks.selectCollectionType')
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

const assetTypeNames = computed(() => {
    return assetStore.getAssetTypesNames;
});

const assetTypeOptions = computed(() => {
    return assetStore.getAssetTypes.map((type) => ({
        id: type.id,
        name: type.name,
        icon: type.icon ? getAppIcon(type.icon) : null,
    }));
});

const assetTemplates = computed(() => {
    return templateStore.getTemplates;
});

const collectionTypeNames = computed(() => {
    return collectionStore.getCollectionTypesNames;
});

const collectionTypeOptions = computed(() => {
    return collectionStore.getCollectionTypes.map((type) => ({
        id: type.id,
        name: type.name,
        icon: type.icon ? getAppIcon(type.icon) : null,
    }));
});

const itemTypeLabels = computed(() => ({
    selectType: t('blocks.selectType'),
    Asset: t('blocks.asset'),
    Collection: t('blocks.collection'),
    Workflow: t('blocks.workflow'),
}));

const itemTypeLabel = computed(() => itemTypeLabels.value[itemType.value] || itemTypeLabels.value.selectType);

const itemTypes = computed(() => {
    const keys = projectWorkflows.value.length
        ? ['Asset', 'Collection', 'Workflow']
        : ['Asset', 'Collection'];
    return keys.map((key) => itemTypeLabels.value[key]);
});

const newWorkflowItemData = computed(() => {
    const itemTypeName = itemType.value;
    let data = {};
    if (itemTypeName === 'Asset') {

        const allAssetTypes = assetStore.getAssetTypes;
        const firstAssetType = allAssetTypes[0];
        assetTypeIcon.value = assetTypeIcon.value ? assetTypeIcon.value : firstAssetType.icon;
        assetTypeId.value = assetTypeId.value ? assetTypeId.value : firstAssetType.id;

        const allAssetTemplates = templateStore.getTemplates;
        const firstAssetTemplate = allAssetTemplates[0];
        assetTemplateId.value = assetTemplateId.value ? assetTemplateId.value : firstAssetTemplate.id;

        data = {
            id: workflowId.value,
            name: workflowName.value,
            template_id: assetTemplateId.value,
            asset_type_id: assetTypeId.value,
            asset_type_icon: assetTypeIcon.value,
            type: 'Asset',
        };
    } else if (itemTypeName === 'Collection') {

        const allCollectionTypes = collectionStore.getCollectionTypes;
        const firstCollectionType = allCollectionTypes[0];
        collectionTypeIcon.value = collectionTypeIcon.value ? collectionTypeIcon.value : firstCollectionType.icon;
        collectionTypeId.value = collectionTypeId.value ? collectionTypeId.value : firstCollectionType.id;

        data = {
            id: workflowId.value,
            name: workflowName.value,
            collection_type_id: collectionTypeId.value,
            collection_type_icon: collectionTypeIcon.value,
            type: 'Collection',
        };
    } else if (itemTypeName === 'Workflow') {

        const allWorkflowTemplates = projectWorkflows.value;
        const firstWorkflowTemplate = allWorkflowTemplates[0];
        workflowTemplateId.value = workflowTemplateId.value ? workflowTemplateId.value : firstWorkflowTemplate.id;
        // workflowName.value = selectedWorkflowName.value;
        data = {
            id: workflowId.value,
            name: workflowName.value,
            collection_type_id: collectionTypeId.value,
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
    if (itemType.value === 'selectType') {
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

const selectAssetType = (assetTypeName) => {
    const allAssetTypes = assetStore.getAssetTypes;
    const selectedAssetType = allAssetTypes.find((item) => item.name === assetTypeName);

    assetTypeId.value = selectedAssetType.id;
    assetTypeIcon.value = selectedAssetType.icon;
};

const selectCollectionType = (collectionTypeName) => {
    const allCollectionTypes = collectionStore.getCollectionTypes;
    const selectedCollectionType = allCollectionTypes.find((item) => item.name === collectionTypeName);

    collectionTypeId.value = selectedCollectionType.id;
    collectionTypeIcon.value = selectedCollectionType.icon;
};

const selectTemplate = (assetTemplate) => {
    assetTemplateId.value = assetTemplate.id;
};

const selectWorkflow = (workflowName) => {
    selectedWorkflowName.value = workflowName;
    const allWorkflows = projectWorkflows.value;
    const selectedWorkflow = allWorkflows.find((item) => item.name === workflowName);

    workflowTemplateId.value = selectedWorkflow.id;
};

const changeItemType = (newItemTypeLabel) => {
    const entry = Object.entries(itemTypeLabels.value).find(([, label]) => label === newItemTypeLabel);
    const newItemTypeKey = entry ? entry[0] : newItemTypeLabel;
    if (newItemTypeKey === 'Asset') {
        if (!assetTemplates.value.length) {
            notificationStore.addNotification(
                t('notifications.noAssetTemplates'),
                t('notifications.noAssetTemplatesLong'),
                'warning',
            );
            return;
        }
        assetTemplateId.value = assetTemplates.value[0]?.id;
    }
    itemType.value = newItemTypeKey;
};

onMounted(() => {
    if (props.isUpdate) {

        workflowId.value = props.workflowItemData.id || '';
        workflowName.value = props.workflowItemData.name || '';

        if (props.workflowItemData.asset_type_id) {
            itemType.value = 'Asset';
            assetTypeId.value = props.workflowItemData.asset_type_id;
            assetTypeIcon.value = props.workflowItemData.asset_type_icon || '';
            assetTemplateId.value = props.workflowItemData.template_id || '';
        } else if (props.workflowItemData.collection_type_id) {
            itemType.value = 'Collection';
            collectionTypeId.value = props.workflowItemData.collection_type_id;
            collectionTypeIcon.value = props.workflowItemData.collection_type_icon || '';
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

.asset-options-container {
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
    background-color: var(--surface-2);
    border-radius: var(--normal-radius);
    padding: .5rem;
    outline: var(--transparent-line);
    outline-offset: -1px;
}
</style>