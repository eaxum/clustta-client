<template>
	<div class="create-menu">
		<ActionButton :icon="getAppIcon('file-plus')" v-if="!kanbanView && (canCreateTask || canModifyEntity)"
			@click="createAsset" v-tooltip="'Add Asset'" />
		<ActionButton :icon="getAppIcon('folder-plus')" v-if="!kanbanView && canCreateEntity || canModifyEntity"
			@click="createEntity" v-tooltip="'Add Collection'" />
		<ActionButton :icon="getAppIcon('workflow-plus')" v-if="!kanbanView && hasWorkflows && canCreateEntity"
			@click="createWorkflow" v-tooltip="'Add Workflow'" />
		<ActionButton :icon="getAppIcon('web-plus')" v-if="!kanbanView && canCreateTask"
			@click="createWebLink" v-tooltip="'Add Weblink'" />
		<ActionButton :icon="getAppIcon('arrow-down-ramp')" v-if="!platformStore.isWeb && !kanbanView && canCreateEntity"
			@click="importItems" v-tooltip="'Import Items'" />
	</div>
</template>

<script setup>
// imports
import { computed } from 'vue';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// stores
import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { usePlatformStore } from '@/stores/platform';
import { useStageStore } from '@/stores/stages';
import { useUserStore } from '@/stores/users';
import { useWorkflowStore } from '@/stores/workflow';

const collectionStore = useCollectionStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const platformStore = usePlatformStore();
const stage = useStageStore();
const userStore = useUserStore();
const workflowStore = useWorkflowStore();

// props
const props = defineProps({
	importItems: { type: Function, required: true },
	kanbanView: { type: Boolean, default: false },
});

// computed properties
const canCreateEntity = computed(() => userStore.canDo('create_entity'));

const canCreateTask = computed(() => userStore.canDo('create_task'));

const canModifyEntity = computed(() => {
	if (!collectionStore.selectedCollection) return false;
	let selectedIsMarked = stage.markedItems.includes(collectionStore.selectedCollection.id);
	if (selectedIsMarked && stage.markedItems.length === 1) return collectionStore.selectedCollection.can_modify;
	return false;
});

const hasWorkflows = computed(() => !!workflowStore.workflows.length);

// methods

// Clears all item selections and resets selection state.
const clearSelection = () => {
	stage.markedItems = [];
	stage.selectedItem = [];
	stage.selectedItems = [];
	stage.firstSelectedItemId = '';
	stage.lastSelectedItemId = '';
};

// Opens the application selection modal to create a new asset.
const createAsset = () => { clearSelection(); modals.setModalVisibility('selectAppModal', true); };

// Opens the create collection modal.
const createEntity = () => { if (!stage.groupItems) clearSelection(); modals.setModalVisibility('createCollectionModal', true); };

// Opens the add web link modal.
const createWebLink = () => { clearSelection(); modals.setModalVisibility('addWebLinkModal', true); };

// Opens the workflow selection modal.
const createWorkflow = () => { clearSelection(); modals.setModalVisibility('selectWorkflowModal', true); };

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);
</script>

<style scoped>
.create-menu {
	position: relative;
	display: flex;
	align-items: center;
	gap: .4rem;
	width: max-content;
	height: max-content;
}
</style>
