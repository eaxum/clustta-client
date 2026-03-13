<template>
	<div class="create-menu">
		<ActionButton :icon="getAppIcon('file-plus')" :isDisabled="kanbanView || !(canCreateAsset || canModifyCollection)"
			@click="createAsset" v-tooltip="$t('components.createMenu.addAsset')" />
		<ActionButton :icon="getAppIcon('folder-plus')" :isDisabled="kanbanView || !(canCreateCollection || canModifyCollection)"
			@click="createCollection" v-tooltip="$t('components.createMenu.addCollection')" />
		<ActionButton :icon="getAppIcon('arrow-down-on-square-stack')" v-if="!(platformStore.isWeb || kanbanView)"  :isDisabled="!(canCreateCollection || canModifyCollection)"
			@click="importItems" v-tooltip="$t('components.createMenu.importItems')" />
		<ActionButton :icon="getAppIcon('workflow-plus')" :isDisabled="kanbanView || !(canCreateCollection || canModifyCollection)"
			@click="createWorkflow" v-tooltip="$t('components.createMenu.addWorkflow')" />
		<ActionButton :icon="getAppIcon('web-plus')" :isDisabled="kanbanView || !(canCreateAsset || canModifyCollection)"
			@click="createWebLink" v-tooltip="$t('components.createMenu.addWeblink')" />
		<ActionButton v-if="integrationStore.linkedIntegration" :icon="getAppIcon('kitsu')"  :isDisabled="kanbanView || !(canCreateAsset || canModifyCollection)"
			v-tooltip="'Sync Now'" :buttonFunction="openSyncModal" />
		<!-- <ActionButton :icon="getAppIcon('arrow-down-ramp')" :isDisabled="platformStore.isWeb || kanbanView || !canCreateCollection"
			@click="importItems" v-tooltip="'Import Items'" /> -->
	</div>
</template>

<script setup>
// imports
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';

const { t } = useI18n();

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// services
import { CollectionService } from '@/services';

// stores
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useUserStore } from '@/stores/users';
import { useWorkflowStore } from '@/stores/workflow';

const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const modals = useDesktopModalStore();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const userStore = useUserStore();
const workflowStore = useWorkflowStore();

// props
const props = defineProps({
	importItems: { type: Function, required: true },
	kanbanView: { type: Boolean, default: false },
});

// computed properties
const canCreateCollection = computed(() => userStore.canDo('create_collection'));

const canCreateAsset = computed(() => userStore.canDo('create_asset'));

// refs
const canModifyCollection = ref(false);

// methods

// Checks whether the current user is assigned to the navigated collection or any ancestor.
const checkModifyPermission = async () => {
	const collection = collectionStore.navigatedCollection;
	if (!commonStore.navigatorMode || !collection) {
		canModifyCollection.value = false;
		return;
	}
	const userId = userStore.user?.id;
	if (!userId || !projectStore.activeProject) {
		canModifyCollection.value = false;
		return;
	}
	try {
		canModifyCollection.value = await CollectionService.IsUserAssignedToCollectionOrAncestor(
			projectStore.activeProject.uri, collection.id, userId
		);
	} catch {
		canModifyCollection.value = false;
	}
};

// watchers
watch(() => collectionStore.navigatedCollection, checkModifyPermission, { immediate: true });

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
	emitter.on('refresh-browser', checkModifyPermission);
	await integrationStore.loadLinkedIntegration();
});

onUnmounted(() => { emitter.off('refresh-browser', checkModifyPermission); });
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
