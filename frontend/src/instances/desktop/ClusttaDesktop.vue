<template>
	<div style="--wails-drop-target-active" id="clustta-desktop"
		:class="['desktop-root', { 'no-radius': isMaximized }]">
		<TitleBar v-if="titleBarVisible" />
		<FlashMessage :isDesktop="true" />
		<ModalView v-if="modals.activeModal" />
		<div class="desktop-container">
			<div ref="desktopBody" id="desktop-body" class="desktop-body tray-root">
				<SidePane v-if="panes.enabledPanes.includes(stage.selectedStage)" :isWideScreen="isWideScreen" />
				<div class="active-project">
					<HeaderBar v-if="stage.activeStage !== 'projects'" />
					<div ref="mainAreaContainer" class="main-area">
						<CenterStage />
					</div>
        			<InfoBar v-if="!platformStore.isWeb" />
				</div>
			</div>
		</div>
	</div>
</template>

<script setup>
import { computed, watchEffect, ref, onMounted, onBeforeUnmount } from 'vue';


// components
import TitleBar from '@/instances/desktop/components/TitleBar.vue'
import InfoBar from '@/instances/desktop/components/InfoBar.vue'
import ModalView from '@/instances/desktop/components/ModalView.vue'
import SidePane from "@/instances/desktop/components/SidePane.vue";
import HeaderBar from "@/instances/desktop/components/HeaderBar.vue";
import CenterStage from "@/instances/desktop/components/CenterStage.vue";
import FlashMessage from '@/instances/common/components/FlashMessage.vue';

// state imports
import { useMenu } from '@/stores/menu';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useUserStore } from '@/stores/users';
import { useProjectStore } from '@/stores/projects';
import { useDndStore } from '@/stores/dnd';
import { usePlatformStore } from '@/stores/platform';
import { Events } from "@wailsio/runtime";

// refs
const desktopBody = ref(null);
const isMaximized = ref(false);
const mainAreaContainer = ref(null);
// states/stores
const menu = useMenu();
const panes = usePaneStore();
const stage = useStageStore();
const modals = useDesktopModalStore();
const userStore = useUserStore();
const dndStore = useDndStore();
const platformStore = usePlatformStore();

// computed properties
const isUserActivated = computed(() => userStore.user !== null);
const isWideScreen = ref(false);
const titleBarVisible = computed(() => {
	if (platformStore.isWeb) {
		return stage.activeStage === 'projects';
	}
	return true;
});(() => userStore.user !== null);

// methods

Events.On(Events.Types.Windows.WindowDragLeave, async (data) => {
	dndStore.isDragging = false;
	dndStore.isDropHovering = false;
});

const hideContextMenu = (event) => {
	if (menu.contextMenuVisible) {
	}
};

// watchers
watchEffect(() => {
	if (mainAreaContainer.value) {
		menu.contextMenuBounds = mainAreaContainer.value;
	}
	if (isUserActivated.value) {
		if (userStore.canDo('create_asset')) {
			dndStore.userCanDrag = true;
		} else {
			dndStore.userCanDrag = false;
		}
	}
});

// lifecycle hooks
onMounted(async () => {
	document.addEventListener('click', hideContextMenu);
});

onBeforeUnmount(async () => {
	document.removeEventListener('click', hideContextMenu);
});


</script>


<style scoped>
@import "@/assets/desktop.css";

.operation-mask {
  position: absolute;
  z-index: 3;
  width: 100%;
  height: 100%;
  display: flex;
  transition: opacity 0.3s ease;
  background-color: rgba(0, 0, 0, 0.5);
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  cursor: not-allowed;
  cursor: wait;
}

.holder {
	color: white;
	display: flex;
	align-items: center;
	justify-content: center;
}

.desktop-root {
	position: fixed;
	width: 100%;
	height: 100%;
	right: 50%;
	transform: translateX(50%);
	overflow: hidden;
	display: flex;
	box-sizing: border-box;
	flex-direction: column;
}

.desktop-container {
	position: relative;
	flex-direction: column;
	width: 100%;
	height: 100%;
	overflow: hidden;
	display: flex;
	box-sizing: border-box;
	min-width: 1000px;
}

.no-radius {
	border-radius: 0;
	outline: 0px;
	width: 100.001%;
	border: 0px;
}

.desktop-body {
	display: flex;
	height: 100vh;
	width: 100vw;
	background-color: var(--black);
	overflow: hidden;
	box-sizing: border-box;
}

.active-project {
	display: flex;
	flex: 1;
	flex-direction: column;
	width: 100%;
	height: 100%;
	overflow: hidden;
	box-sizing: border-box;
}

.main-area {
	position: relative;
	box-sizing: border-box;
	display: flex;
	width: 100%;
	height: 100%;
	overflow: hidden;
	padding: .4rem;
	padding-bottom: 0;
	padding-top: 0;
	background-color: forestgreen;
	background-color: var(--shadow-steel);
}
</style>

