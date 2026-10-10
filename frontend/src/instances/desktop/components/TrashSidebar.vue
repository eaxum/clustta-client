<template>
	<div class="trash-sidebar">
		<SidebarHeader :title="$t('components.headerBar.trash')" :backFunction="goBack">
			<ActionButton
				v-if="trayStates.trashables.length"
				:icon="getAppIcon('trash')"
				:buttonFunction="confirmEmptyTrash"
				:useDanger="true"
				v-tooltip="$t('components.headerBar.emptyTrash')"
			/>
		</SidebarHeader>

		<SidebarNavigationList
			:activePageId="trayStates.trashTypeFilter"
			:ariaLabel="$t('components.headerBar.trash')"
			:groups="navigationGroups"
			@select="selectFilter"
		/>
	</div>
</template>

<script setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import SidebarHeader from '@/instances/desktop/components/SidebarHeader.vue';
import SidebarNavigationList from '@/instances/desktop/components/SidebarNavigationList.vue';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTrayStates } from '@/stores/TrayStates';
import { ProjectService } from '@/services';
import utils from '@/services/utils';

const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const trayStates = useTrayStates();
const { t } = useI18n();

const navigationGroups = computed(() => [{
	id: 'trash',
	name: '',
	pages: trayStates.trashTypes
		.filter((trashType) => !trashType.name.includes('checkpoint'))
		.map((trashType) => ({
			id: trashType.name,
			icon: trashType.icon,
			name: utils.capitalizeStr(trashType.name),
		})),
}]);

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

const goBack = () => {
	if (projectStore.activeProjectCanQuery) {
		stage.setStageVisibility('browser', true);
		return;
	}
	stage.setStageVisibility('projects', true);
};

const selectFilter = (filterId) => {
	trayStates.trashTypeFilter = filterId;
};

const emptyTrash = async () => {
	try {
		await ProjectService.Purge(projectStore.activeProject.uri);
		trayStates.trashables = [];
		modals.disableAllModals();
	} catch (error) {
		console.error(error.message);
		notificationStore.addNotification(
			t('components.headerBar.errorSyncingData'),
			error.message,
			'error',
			false,
		);
		modals.disableAllModals();
	}
};

const confirmEmptyTrash = () => {
	trayStates.popUpModalIcon = 'trash';
	trayStates.popUpModalTitle = t('components.headerBar.emptyTrashTitle');
	trayStates.popUpModalMessage = t('components.headerBar.emptyTrashMessage');
	trayStates.popUpModalFunction = emptyTrash;
	modals.setModalVisibility('popUpModal', true);
};
</script>

<style scoped>
.trash-sidebar {
	display: flex;
	flex-direction: column;
	width: 240px;
	min-width: 240px;
	height: 100%;
	min-height: 0;
	padding: 1rem .45rem .25rem;
	color: var(--text);
	background-color: var(--surface-1);
	box-sizing: border-box;
}
</style>
