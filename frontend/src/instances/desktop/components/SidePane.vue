<template>
	<div class="side-pane" ref="sidePane" :style="{ width: sidePaneWidth }">
		<div class="side-pane-content">
			<div class="project-section">
				<div class="project-list-items">
					<TabButton :icon="getAppIcon('home')" v-tooltip="'All Projects'" @click="goToProjects"
						:showLabel="sidePaneActive" :fullWidth="sidePaneActive" label="All Projects"
						:buttonFunction="doNothing" />
					<ProjectList />
				</div>
			</div>
			<div class="pane-options" :class="{ 'pane-options-minimized': !sidePaneActive }">
				<div class="menu-divider"></div>

				<div v-stop-propagation @click="showUserAccountMenu" class="user-avatar">
					<span v-tooltip="userFullName" class="single-action-button">
						<div class="profile-picture" :style="{ backgroundColor: profileColor(user?.id) }">
							<img v-if="userStore.user.photo" class="profile-img" :src="userStore.user.photo">
							<img v-else class="profile-img" :src="getAppIcon('person')">
						</div>
					</span>
				</div>

				<TabButton :icon="getAppIcon('cog')" v-tooltip="'Settings'"
					@click="stage.setStageVisibility('settings', true)" :isActive="stage.activeStage === 'settings'"
					:showLabel="sidePaneActive" :fullWidth="sidePaneActive" label="Settings" />
				<TabButton v-if="!userStore.getUserAuthentication" :icon="getAppIcon('login')" v-tooltip="'Login'"
					:showLabel="sidePaneIsOpen" :fullWidth="sidePaneIsOpen" :buttonFunction="logUserIn" />
				<!-- <TabButton v-else :icon="getAppIcon('logout')" v-tooltip="'Logout'" :showLabel="sidePaneIsOpen"
					:fullWidth="sidePaneIsOpen" :buttonFunction="logUserOut" /> -->
			</div>
		</div>
	</div>
</template>

<script setup>
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';

// services
import { AuthService } from "@/../bindings/clustta/services";

// stores
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';
import { useIconStore } from '@/stores/icons';
import { useStageStore } from '@/stores/stages';
import { usePaneStore } from '@/stores/panes';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useMenu } from '@/stores/menu';

// imports
const stage = useStageStore();
const panes = usePaneStore();
const trayStates = useTrayStates();
const projectStore = useProjectStore();
const userStore = useUserStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const iconStore = useIconStore();
const menu = useMenu();

// props
const props = defineProps({
	isWideScreen: Boolean
})

// components
import TabButton from '@/instances/desktop/components/TabButton.vue';
import ProjectList from '@/instances/desktop/components/ProjectList.vue'

// refs
const sidePane = ref(null);
const sidePaneIsOpen = ref(false);
const sidePaneIsPinned = ref(false);
const projectListClosed = ref(false);
const projectListPinned = ref(false);

const fullWidth = '240px';
const collapsedWidth = '50px';

const sidePaneWidth = ref(collapsedWidth);

// computed properties
const user = computed(() => { return userStore.user });

const sidePaneActive = computed(() => { return sidePaneIsOpen.value || sidePaneIsPinned.value });

// TODO
stage.sidePaneActive = computed(() => { return sidePaneActive.value });

const showPin = computed(() => { return !projectListClosed.value && projectListPinned.value });

// methods
const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

const profileColor = (uuid) => {
	let accountUser = user.value;
	if (!accountUser) {
		return 'black'
	}
	const parts = uuid.split('-');
	return '#' + parts[0];
};

const userFullName = computed(() => {
	let accountUser = user.value;
	if (!accountUser) {
		return 'No User'
	} else {
		return `${accountUser.first_name} ${accountUser.last_name}`;
	}
});

const doNothing = () => {
	// console.log('nothing');
};

const goToProjects = () => {
	stage.setStageVisibility('projects', true);
	panes.setPaneVisibility('projectDetails', true);
};

const showUserAccountMenu = (event) => {
	if(userStore.getUserAuthentication){
		// Show the account menu anchored to the right of the button
		menu.showContextMenu(event, 'accountMenu', true, { anchor: true, position: 'right' });
	}
};

const collapseSidePane = () => {
	sidePaneWidth.value = collapsedWidth;
	sidePaneIsOpen.value = !sidePaneIsOpen.value;
	sidePaneIsPinned.value = false;
}

const handleClickOutside = (event) => {
	if (!props.isWideScreen) {
		if (sidePaneIsOpen.value && event.target.closest('.active-project')) {
			collapseSidePane();
		};
	}
};

const logUserOut = async () => {
	await AuthService.Logout()
		.then((data) => {
			userStore.$reset()
			projectStore.$reset();
			trayStates.$reset();
			console.log('user logged out');
		}
		)
		.catch((error) => {
			notificationStore.errorNotification("Logout Failed", error)
		});
}

const logUserIn = () => {
	modals.setModalVisibility('loginModal', true);
};

onMounted(() => {
	document.addEventListener('click', handleClickOutside);
});

onBeforeUnmount(() => {
	document.removeEventListener('click', handleClickOutside);
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.menu-divider{
	height: 5px;
	margin-top: 10px;
}

.sidebar-tabs {
	overflow: hidden;
	background-color: transparent;
	text-align: center;
	font-size: 14px;
	line-height: 14px;
	background-color: transparent;
	color: var(--white);
	position: relative;
	border-radius: 8px;
	box-sizing: border-box;
	cursor: pointer;
	display: flex;
	gap: 10px;
	align-items: center;
	padding: .3rem;
	height: max-content;
	width: max-content;
	transition: all 0.3s ease;
	/* background-color: red; */
	opacity: .4;

}

.sidebar-tabs:hover {
	background-color: rgb(121, 121, 121);
	background-color: #ffffff15;
}

.sidebar-tabs:active {
	background-color: rgb(70, 70, 70);
	background-color: #00000013;
}

.sidebar-tabs-pressed {
	box-sizing: border-box;
	background-color: rgba(0, 0, 0, 0.216);
	outline: solid 1px var(--white);
	outline-offset: -1px;
}

.sidebar-tabs-active {
	opacity: 1;
}

.side-pane-open {
	transform: rotate(-180deg);
}

.project-list-header-closed {
	transform: rotate(90deg);
}

.side-pane {
	border-right: var(--transparent-line);
	color: var(--white);
	display: flex;
	flex-direction: column;
	justify-content: space-between;
	box-sizing: border-box;
	overflow: hidden;
	padding: .1rem;
	align-items: center;
	width: 100%;
	max-width: 240px;
	height: 100%;
	transition: all 0.1s cubic-bezier(0.6, 0.05, 0.01, 0.99);
	background-color: var(--black);
	/* background-color: crimson; */

	/* flex: 1 1 100%; */
}

.side-pane-content {
	/* background-color: greenyellow; */
	display: flex;
	flex-direction: column;
	justify-content: space-between;
	box-sizing: border-box;
	overflow: hidden;
	/* padding: 1rem .1rem; */
	/* padding-bottom: 1rem; */
	width: 100%;
	height: 100%;
	transition: all 0.1s cubic-bezier(0.6, 0.05, 0.01, 0.99);
	/* flex: 1 1 100%; */
}

.pane-options {
	display: flex;
	box-sizing: border-box;
	overflow: hidden;
	flex-direction: column;
	gap: 1rem;
	align-items: flex-start;
	justify-content: flex-start;
	/* padding: .2rem; */
	padding: 1rem 0;
	padding-top: .5rem;
	width: 100%;
	height: max-content;
	min-height: max-content;
	overflow: hidden;
	/* background-color: darkorange; */
}

.pane-options-minimized {
	justify-content: center;
	align-items: center;
	/* background-color: rebeccapurple; */
	/* justify-content: flex-start; */
	align-items: center;
}

.project-section-header {
	color: var(--white);
	display: flex;
	box-sizing: border-box;
	overflow: hidden;
	flex-direction: row;
	gap: .4rem;
	align-items: center;
	justify-content: space-between;
	/* padding: .2rem; */
	width: 100%;
	height: max-content;
	/* background-color: darkslategrey; */
}

.hide-project-section-header {
	/* visibility: hidden; */
	justify-content: center;
	/* background-color: forestgreen; */
}

.project-section {
	display: flex;
	box-sizing: border-box;
	overflow: hidden;
	flex-direction: column;
	/* align-items: flex-start;
	justify-content: flex-start; */
	width: 100%;
	height: max-content;
	/* background-color: blue; */
	flex: 1;
}

.project-list-items {

	display: flex;
	box-sizing: border-box;
	overflow: hidden;
	flex-direction: column;
	gap: 1rem;
	align-items: center;
	justify-content: flex-start;
	width: 100%;
	height: max-content;
	flex: 1;
	padding-top: .5rem;
	/* background-color: hotpink; */
}

.chevron-container {
	display: flex;
	box-sizing: border-box;
	overflow: hidden;
	/* flex-direction: column; */
	gap: .4rem;
	align-items: center;
	justify-content: center;
	/* padding: .2rem; */
	width: 100%;
	height: max-content;
	height: 44px;
	/* background-color: darkorange; */
	/* transition: all 0.3s cubic-bezier(0.6, 0.05, 0.01, 0.99); */
}

.chevron-icon {
	display: flex;
	box-sizing: border-box;
	overflow: hidden;
	width: min-content;
	height: min-content;
	/* background-color: teal; */
	align-items: center;
	justify-content: center;
	/* transition: all 0.3s ease; */
}

.user-avatar {
	overflow: hidden;
	/* background-color: firebrick; */
	color: var(--white);
	position: relative;
	border-radius: 8px;
	box-sizing: border-box;
	cursor: pointer;
	display: flex;
	aspect-ratio: 1;
	width: 100%;
	align-items: center;
	align-items: center;
	justify-content: center;
}
</style>

