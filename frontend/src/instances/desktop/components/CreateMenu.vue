<template>
	<div class="create-menu">
		<ActionButton :icon="getAppIcon('file-plus')" :isDisabled="props.disabled || kanbanView || !canCreateAsset"
			@click="createAsset" v-tooltip="{ text: $t('components.createMenu.addAsset'), shortcut: 'newAsset' }" />
		<ActionButton :icon="getAppIcon('folder-plus')" :isDisabled="props.disabled || kanbanView || !canCreateCollection"
			@click="createCollection" v-tooltip="{ text: $t('components.createMenu.addCollection'), shortcut: 'newCollection' }" />
		<ActionButton :icon="getAppIcon('data-download')" v-if="!(platformStore.isWeb || kanbanView)"  :isDisabled="props.disabled || !canCreateAsset"
			@click="importItemsWithPermission" v-tooltip="$t('components.createMenu.importItems')" />
		<ActionButton :icon="getAppIcon('data-upload')" v-if="!(platformStore.isWeb || kanbanView)"
			@click="openExport" v-tooltip="$t('components.createMenu.exportItems')" />
		<ActionButton :icon="getAppIcon('workflow-plus')" :isDisabled="props.disabled || kanbanView || !canCreateWorkflow"
			@click="createWorkflow" v-tooltip="$t('components.createMenu.addWorkflow')" />
		<ActionButton :icon="getAppIcon('web-plus')" :isDisabled="props.disabled || kanbanView || !canCreateAsset"
			@click="createWebLink" v-tooltip="{ text: $t('components.createMenu.addWeblink'), shortcut: 'newLink' }" />
		<ActionButton v-if="integrationStore.linkedIntegration" :icon="getAppIcon('kitsu')" :isDisabled="props.disabled || kanbanView"
			v-tooltip="$t('kitsu.importFromKitsu')" :buttonFunction="openSyncModal" />
		<!-- <ActionButton :icon="getAppIcon('arrow-down-ramp')" :isDisabled="platformStore.isWeb || kanbanView || !canCreateCollection"
			@click="importItems" v-tooltip="'Import Items'" /> -->
	</div>
</template>

<script setup>
// imports
import { computed } from 'vue';
import { canCreateAssetHere, canCreateCollectionHere } from '@/lib/permissions';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useExportStore } from '@/stores/exports';
import { usePlatformStore } from '@/stores/platform';
import { useStageStore } from '@/stores/stages';

const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const exportStore = useExportStore();
const modals = useDesktopModalStore();
const platformStore = usePlatformStore();
const stage = useStageStore();

// props
const props = defineProps({
	disabled: { type: Boolean, default: false },
	importItems: { type: Function, required: true },
	kanbanView: { type: Boolean, default: false },
});

// computed properties
const canCreateAsset = computed(() => canCreateAssetHere());

const canCreateCollection = computed(() => canCreateCollectionHere());

const canCreateWorkflow = computed(() => canCreateAsset.value && canCreateCollection.value);
const creationBlocked = computed(() => props.disabled || props.kanbanView);

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
const createAsset = () => {
	if (creationBlocked.value || !canCreateAsset.value) return;
	clearSelection();
	modals.setModalVisibility('selectAppModal', true);
};

// Opens the create collection modal.
const createCollection = () => {
	if (creationBlocked.value || !canCreateCollection.value) return;
	if (!stage.groupItems) clearSelection();
	modals.setModalVisibility('createCollectionModal', true);
};

// Opens the add web link modal.
const createWebLink = () => {
	if (creationBlocked.value || !canCreateAsset.value) return;
	clearSelection();
	modals.setModalVisibility('addWebLinkModal', true);
};

// Opens the workflow selection modal.
const createWorkflow = () => {
	if (creationBlocked.value || !canCreateWorkflow.value) return;
	clearSelection();
	modals.setModalVisibility('selectWorkflowModal', true);
};

// Imports untracked items only when asset creation is allowed here.
const importItemsWithPermission = () => {
	if (props.disabled || !canCreateAsset.value) return;
	return props.importItems();
};

// Opens an export preview for the assets currently displayed in the browser.
const openExport = () => exportStore.open('selection');

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Opens the integration sync modal.
const openSyncModal = () => {
	if (creationBlocked.value) return;
	modals.setModalVisibility('integrationSyncModal', true);
};

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
