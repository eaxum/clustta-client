<template>
	<div class="settings-sidebar">
		<SidebarHeader :title="$t(titleKey)" :backFunction="goBack" />

		<SidebarNavigationList
			:activePageId="settings.activeModal || ''"
			:ariaLabel="$t(titleKey)"
			:groups="navigationGroups"
			@select="selectPage"
		/>
	</div>
</template>

<script setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import SidebarHeader from '@/instances/desktop/components/SidebarHeader.vue';
import SidebarNavigationList from '@/instances/desktop/components/SidebarNavigationList.vue';
import {
	getAvailableProjectSettingsPages,
	getAvailableStudioSettingsPages,
	projectSettingsGroups,
	studioSettingsGroups,
	userSettingsGroups,
	userSettingsPages,
} from '@/instances/desktop/settings/settingsNavigation';
import { useEntitlementStore } from '@/stores/entitlements';
import { useMenu } from '@/stores/menu';
import { useProjectStore } from '@/stores/projects';
import { useSettingsStore } from '@/stores/settings';
import { useStageStore } from '@/stores/stages';

const entitlementStore = useEntitlementStore();
const menu = useMenu();
const projectStore = useProjectStore();
const settings = useSettingsStore();
const stage = useStageStore();
const { t } = useI18n();

const navigationConfig = computed(() => {
	if (stage.activeStage === 'studioSettings') {
		return {
			groups: studioSettingsGroups,
			pages: getAvailableStudioSettingsPages({
				canCollaborate: entitlementStore.canCollaborate,
				isCloudHosted: projectStore.isCloudHosted,
			}),
			titleKey: 'components.headerBar.studioSettings',
		};
	}

	if (stage.activeStage === 'projectSettings') {
		return {
			groups: projectSettingsGroups,
			pages: getAvailableProjectSettingsPages({
				canCollaborate: entitlementStore.canCollaborate,
				hasCustomRoles: entitlementStore.hasCustomRoles,
				hasIntegrations: entitlementStore.hasIntegrations,
				isRemoteProject: !!projectStore.activeProject?.has_remote,
			}),
			titleKey: 'components.headerBar.projectSettings',
		};
	}

	return {
		groups: userSettingsGroups,
		pages: userSettingsPages,
		titleKey: 'settings.title',
	};
});

const titleKey = computed(() => navigationConfig.value.titleKey);

const navigationGroups = computed(() => {
	return navigationConfig.value.groups
		.map((group) => ({
			...group,
			name: t(group.nameKey),
			pages: navigationConfig.value.pages
				.filter((page) => page.group === group.id)
				.map((page) => ({
					...page,
					badge: page.badgeKey ? t(page.badgeKey) : '',
					name: t(page.nameKey),
				})),
		}))
		.filter((group) => group.pages.length);
});

const goBack = () => {
	if (stage.activeStage === 'projectSettings' && projectStore.activeProjectCanQuery) {
		stage.setStageVisibility('browser', true);
		return;
	}
	stage.setStageVisibility('projects', true);
};

const selectPage = (pageId) => {
	const pageIsAvailable = navigationConfig.value.pages.some((page) => page.id === pageId);
	if (!pageIsAvailable) return;
	menu.disableAllMenus();
	settings.activeModalName = pageId;
	settings.setModalVisibility(pageId, true);
};
</script>

<style scoped>
.settings-sidebar {
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
