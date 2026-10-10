<template>
	<div ref="pageListRoot" class="page-list-root absolute-pane">
		<SettingsShell
			:pages="availablePages"
			:activePageId="activePageId"
		/>
	</div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch, watchEffect } from 'vue';
import SettingsShell from '@/instances/desktop/settings/components/SettingsShell.vue';
import {
	getAvailableStudioSettingsPages,
} from '@/instances/desktop/settings/settingsNavigation';
import { useEntitlementStore } from '@/stores/entitlements';
import { useMenu } from '@/stores/menu';
import { useProjectStore } from '@/stores/projects';
import { useSettingsStore } from '@/stores/settings';

const DEFAULT_PAGE_ID = 'studio';

const entitlementStore = useEntitlementStore();
const menu = useMenu();
const projectStore = useProjectStore();
const settings = useSettingsStore();
const pageListRoot = ref(null);
const activePageId = ref(DEFAULT_PAGE_ID);

const availablePages = computed(() => {
	return getAvailableStudioSettingsPages({
		canCollaborate: entitlementStore.canCollaborate,
		isCloudHosted: projectStore.isCloudHosted,
	});
});

const isAvailablePage = (pageId) => {
	return availablePages.value.some((page) => page.id === pageId);
};

const selectPage = (pageId) => {
	if (!isAvailablePage(pageId)) return;
	activePageId.value = pageId;
	settings.activeModalName = pageId;
	settings.setModalVisibility(pageId, true);
};

const ensureAvailablePage = () => {
	if (isAvailablePage(settings.activeModal)) {
		activePageId.value = settings.activeModal;
		return;
	}
	selectPage(DEFAULT_PAGE_ID);
};

watch(
	() => settings.activeModal,
	(pageId) => {
		if (isAvailablePage(pageId)) activePageId.value = pageId;
	},
);

watch(
	() => availablePages.value.map((page) => page.id).join('|'),
	ensureAvailablePage,
);

watchEffect(() => {
	if (pageListRoot.value) menu.clickOutsideMask = pageListRoot.value;
});

onMounted(ensureAvailablePage);

onUnmounted(() => {
	settings.disableAllModals();
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.page-list-root {
	box-sizing: border-box;
	padding: 0;
	display: flex;
	align-items: center;
	justify-content: center;
	color: white;
	/* background-color: sienna; */
}

.settings-stage-root {
	display: flex;
	flex-direction: column;
	box-sizing: border-box;
	/* background-color: firebrick; */
	width: 100%;
	height: 100%;
	gap: .5rem;
}

.settings-stage-header {
	width: 100%;
	display: flex;
	box-sizing: border-box;
	/* background-color: rebeccapurple; */
	align-items: flex-start;
	justify-content: flex-start;
}

.settings-stage-body {
	width: 100%;
	/* max-width: 1200px; */
	height: 100%;
	display: flex;
	box-sizing: border-box;
	/* background-color: teal; */
	align-items: flex-start;
	justify-content: center;
	overflow: hidden;
	padding: .5rem;
}

.settings-stage-body-container {
	width: 100%;
	max-width: 960px;
	height: 100%;
	display: flex;
	box-sizing: border-box;
	/* background-color: tomato; */
	align-items: center;
	justify-content: center;
	overflow: hidden;
	padding: .5rem;
}

.page-list {
	background-color: rgb(125, 192, 59);
	padding: .4rem;
	display: flex;
	gap: .4rem;
	flex-direction: column;
	box-sizing: border-box;
	height: 100%;
	width: 100%;
	/* overflow: hidden; */
	height: max-content;
}

.page-list-container {
	display: flex;
	padding-right: .4rem;
	overflow: hidden;
	overflow-y: scroll;
	height: 100%;
	width: 100%;
	box-sizing: border-box;
	/* max-width: 600px; */
	min-width: 300px;
}

.page-list-container::-webkit-scrollbar {
	width: 8px;
}

.page-list-container::-webkit-scrollbar-thumb {
	border-radius: 10px;
	background-color: rgb(36, 49, 59);
}

.page-list-container::-webkit-scrollbar-track {
	border-radius: 10px;
}

.page-header {
	position: relative;
	display: flex;
	width: 100%;
	align-items: center;
	height: max-content;
	gap: 1rem;
	justify-content: space-between;
	background-color: khaki;
	padding: .2rem;
	box-sizing: border-box;
	min-width: max-content;
}
</style>

