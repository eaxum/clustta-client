<template>
	<div class="create-menu">
		<ActionButton :icon="getAppIcon('file-plus')" :isDisabled="props.disabled || kanbanView || !canCreateAsset"
			@click="createAsset" v-tooltip="$t('components.createMenu.addAsset')" />
		<ActionButton :icon="getAppIcon('folder-plus')" :isDisabled="props.disabled || kanbanView || !canCreateCollection"
			@click="createCollection" v-tooltip="$t('components.createMenu.addCollection')" />
		<ActionButton :icon="getAppIcon('data-download')" v-if="!(platformStore.isWeb || kanbanView)"  :isDisabled="props.disabled || !canCreateAsset"
			@click="importItems" v-tooltip="$t('components.createMenu.importItems')" />
		<ActionButton :icon="getAppIcon('workflow-plus')" :isDisabled="props.disabled || kanbanView || !canCreateWorkflow"
			@click="createWorkflow" v-tooltip="$t('components.createMenu.addWorkflow')" />
		<ActionButton :icon="getAppIcon('web-plus')" :isDisabled="props.disabled || kanbanView || !canCreateAsset"
			@click="createWebLink" v-tooltip="$t('components.createMenu.addWeblink')" />
		<ActionButton v-if="integrationStore.linkedIntegration" :icon="getAppIcon('kitsu')"  :isDisabled="props.disabled || kanbanView || !canCreateAsset"
			v-tooltip="'Sync Now'" :buttonFunction="openSyncModal" />
		<!-- <ActionButton :icon="getAppIcon('arrow-down-ramp')" :isDisabled="platformStore.isWeb || kanbanView || !canCreateCollection"
			@click="importItems" v-tooltip="'Import Items'" /> -->
	</div>
</template>

<script setup>
// imports
import { computed, onMounted } from 'vue';
import { canCreateAssetHere, canCreateCollectionHere } from '@/lib/permissions';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { usePlatformStore } from '@/stores/platform';
import { useStageStore } from '@/stores/stages';

const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
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
const createCollection = () => { if (!stage.groupItems) clearSelection(); modals.setModalVisibility('createCollectionModal', true); };

// Opens the add web link modal.
const createWebLink = () => { clearSelection(); modals.setModalVisibility('addWebLinkModal', true); };

// Opens the workflow selection modal.
const createWorkflow = () => { clearSelection(); modals.setModalVisibility('selectWorkflowModal', true); };

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Opens the integration sync modal.
const openSyncModal = () => { modals.setModalVisibility('integrationSyncModal', true); };

// lifecycle hooks
onMounted(async () => {
	await integrationStore.loadLinkedIntegration();
});
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
