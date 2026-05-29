<template>
	<div class="project-list" ref="projectListRef" 
       :class="{ 
         'is-disabled': stage.operationActive, 
         'top-gradient': showTopGradient && !showCenterGradient, 
         'bottom-gradient': showBottomGradient && !showCenterGradient,
         'center-gradient': showCenterGradient 
       }" 
       @scroll="handleScroll"
       @mousemove="onDockMove" @mouseleave="onDockLeave">
		
	   <div v-if="projects.length" v-tooltip="pinnedIndicatorTooltip" 
			class="pinned-indicator" :class="{ 'clickable': isHoveringPinned }"
			@mouseenter="isHoveringPinned = true" @mouseleave="isHoveringPinned = false" @click="togglePinProject">
			<div class="menu-divider"></div>
			<img class="tiny-icons" :src="getAppIcon(pinnedIndicatorIcon)">
			<div class="menu-divider"></div>
		</div>

		<span class="compound-list-item" v-for="(project, index) in projects" :key="project.uri">


			<div :ref="setDockRef(`p-${project.uri}`)" class="project-avatar-item" v-tooltip="sidePaneActive ? '' : project.name"
				@click="projectStore.gotoProject(project)"
				:class="{ 'project-avatar-item-centered': !sidePaneActive, 'project-avatar-item-active': isActiveProject(project), 'dock-magnified-source': isMagnified(`p-${project.uri}`) }">

				<span v-if="project.icon.length < 10" class="project-icon">{{ decodeEmoji(project.icon) }}</span>
				<span v-else class="project-icon">
					<img class="screenshot-thumb" :src="project.icon">
				</span>

			</div>

		</span>

		<div v-if="recents.length" v-tooltip="isHoveringRecents ? $t('components.projectList.clearRecentProjects') : $t('components.projectList.recentProjects')" 
			 class="pinned-indicator" :class="{ 'clickable': isHoveringRecents }"
			 @mouseenter="isHoveringRecents = true" @mouseleave="isHoveringRecents = false" @click="clearRecents">
			<div class="menu-divider"></div>
			<img class="tiny-icons" :src="getAppIcon(isHoveringRecents ? 'broom' : 'clock')">
			<div class="menu-divider"></div>
		</div>

		<span class="compound-list-item" v-for="(project, index) in recents" :key="project.uri">


			<div :ref="setDockRef(`r-${project.uri}`, index)" class="project-avatar-item" v-tooltip="sidePaneActive ? '' : project.name"
				@click="projectStore.gotoProject(project)"
				:class="{ 'project-avatar-item-centered': !sidePaneActive, 'project-avatar-item-active': isActiveProject(project), 'dock-magnified-source': isMagnified(`r-${project.uri}`) }">

				<span v-if="project.icon.length < 10" class="project-icon">{{ decodeEmoji(project.icon) }}</span>
				<span v-else class="project-icon">
					<img class="screenshot-thumb" :src="project.icon">
				</span>

				<div v-if="project.is_remote && activeProjectIndex === index && !sidePaneActive && project.is_unsynced"
					class="critical-items" :style="{ top: anchor.top + 'px', left: anchor.left + 'px' }">
				</div>

				<div v-if="sidePaneActive && project.is_unsynced" class="critical-items critical-items-static">
				</div>

			</div>

		</span>
	</div>

	<Teleport to="body">
		<div v-if="dockMouseY !== null" class="dock-overlay">
			<div v-for="item in magnifiedItems" :key="item.key" class="dock-overlay-item" :style="item.style">
				<span v-if="item.project.icon.length < 10" class="project-icon">{{ decodeEmoji(item.project.icon) }}</span>
				<span v-else class="project-icon">
					<img class="screenshot-thumb" :src="item.project.icon">
				</span>
			</div>
		</div>
	</Teleport>
	
</template>

<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

// imports
import { computed, ref, reactive, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';

// services
import { SettingsService } from '@/services';
import { decodeEmoji } from '@/services/utils';

// stores/state imports
import { useStageStore } from '@/stores/stages';
import { useProjectStore } from '@/stores/projects';
import { usePlatformStore } from '@/stores/platform';
import { useNotificationStore } from '@/stores/notifications';

// refs
const stage = useStageStore();
const projectStore = useProjectStore();
const platformStore = usePlatformStore();
const notificationStore = useNotificationStore();

const { t } = useI18n();
const listItem = ref([]);
const isHoveringPinned = ref(false);
const isHoveringRecents = ref(false);

// --- dock-style magnification ---------------------------------------------
// Tunables for the vertical sidebar dock effect.
const DOCK_BASE = 35;     // base icon size (matches .project-avatar-item)
const DOCK_MAX = 43.75;   // max size at cursor center (1.25× base)
const DOCK_RANGE = 140;   // px radius of influence along the Y axis

const dockMouseY = ref(null);
const dockScrollTick = ref(0);
const dockItemRefs = reactive({});

// Collect refs for each dock item. When `recentsIndex` is provided, also
// populate the legacy listItem array used by the `anchor` computed.
const setDockRef = (key, recentsIndex = null) => (el) => {
	if (el) {
		dockItemRefs[key] = el;
		if (recentsIndex !== null) listItem.value[recentsIndex] = el;
	} else {
		delete dockItemRefs[key];
	}
};

const onDockMove = (e) => { dockMouseY.value = e.clientY; };
const onDockLeave = () => { dockMouseY.value = null; };

// Cosine-bell falloff for smooth, symmetric dock magnification.
const dockScaleFor = (key) => {
	if (dockMouseY.value == null) return 1;
	const el = dockItemRefs[key];
	if (!el) return 1;
	const rect = el.getBoundingClientRect();
	const center = rect.top + rect.height / 2;
	const dist = Math.abs(dockMouseY.value - center);
	if (dist >= DOCK_RANGE) return 1;
	const t = dist / DOCK_RANGE;
	const falloff = 0.5 * (1 + Math.cos(Math.PI * t));
	const maxScale = DOCK_MAX / DOCK_BASE;
	return 1 + (maxScale - 1) * falloff;
};

const isMagnified = (key) => dockScaleFor(key) > 1.001;

// Floating magnified copies teleported to <body> so they can grow outside the
// scroll container instead of being clipped by it.
const magnifiedItems = computed(() => {
	// Touch dockScrollTick so the overlay recomputes while scrolling.
	void dockScrollTick.value;
	if (dockMouseY.value == null) return [];
	const out = [];
	const collect = (list, prefix) => {
		for (const project of list) {
			const key = `${prefix}-${project.uri}`;
			const el = dockItemRefs[key];
			if (!el) continue;
			const rect = el.getBoundingClientRect();
			const s = dockScaleFor(key);
			if (s <= 1.001) continue;
			out.push({
				key,
				project,
				style: {
					left: `${rect.left}px`,
					top: `${rect.top}px`,
					width: `${rect.width}px`,
					height: `${rect.height}px`,
					transform: `scale(${s})`,
					transformOrigin: 'left center',
				},
			});
		}
	};
	collect(projects.value, 'p');
	collect(recents.value, 'r');
	return out;
});

const props = defineProps({
	sidePaneActive: Boolean,
})

// computed properties
const activeProjectIsPinned = computed(() => {
	const activeProject = projectStore.getActiveProject;
	if (!activeProject) return false;
	return projectStore.pinnedProjects?.includes(activeProject.id);
});

const activeProjectIndex = computed(() => {
	const projects = projectStore.projects;
	const activeProject = projectStore.getActiveProject;
	return projects.indexOf(activeProject);
});

const projects = computed(() => {
	const pinnedProjects = projectStore.pinnedProjects;
	return projectStore.projects.filter((project) => (project.is_downloaded || platformStore.isWeb) && pinnedProjects?.includes(project.id));
});

const recents = computed(() => {
	const pinnedProjects = projectStore.pinnedProjects;
	const recentProjects = projectStore.recentProjects;
	let recentProjectsAvailable = projectStore.projects.filter((project) => (project.is_downloaded || platformStore.isWeb) && recentProjects?.includes(project.id) && !pinnedProjects?.includes(project.id));

	// sort recent projects by most recent
	recentProjectsAvailable.sort((a, b) => {
		return recentProjects.indexOf(a.id) - recentProjects.indexOf(b.id);
	});

	// Calculate how many recent projects we can show
	const pinnedCount = projects.value.length || 0;
	const maxRecentCount = 10 - pinnedCount;

	return recentProjectsAvailable.slice(0, Math.max(0, maxRecentCount));
});

const pinnedIndicatorIcon = computed(() => {
	if (!isHoveringPinned.value) return 'pin';
	return activeProjectIsPinned.value ? 'unpin' : 'new-pin';
});

const pinnedIndicatorTooltip = computed(() => {
	if (!isHoveringPinned.value) return t('components.projectList.pinnedProjects');
	return activeProjectIsPinned.value ? t('components.projectList.unpinProject') : t('components.projectList.pinProject');
});

const criticalItemsDot = computed(() => {
	return projectStore.getActiveProject.has_remote && projectStore.getActiveProject.is_unsynced && !props.sidePaneActive
});

const anchor = computed(() => {
	if (listItem.value) {
		const anchor = listItem.value[activeProjectIndex.value].getBoundingClientRect();
		const anchorData = { top: (anchor.top - 3), left: (anchor.right - 7) }
		return anchorData
	} else {
		return { top: 0, left: 0 }
	}
});

// methods
const isActiveProject = (project) => {
	if (!project) {
		return false
	}
	return projectStore.getActiveProjectName === project.name
};

const dynamicName = (string) => {
	return props.sidePaneActive ? string : string[0].toUpperCase();
};

// Toggles pin state for the active project.
const togglePinProject = async () => {
	const activeProject = projectStore.getActiveProject;
	if (!activeProject) return;
	const studioName = projectStore.getSelectedStudioName;
	const projectId = activeProject.id;
	if (activeProjectIsPinned.value) {
		await SettingsService.UnpinProject(studioName, projectId);
		projectStore.pinnedProjects = projectStore.pinnedProjects.filter((item) => item !== projectId);
	} else {
		await SettingsService.PinProject(studioName, projectId);
		projectStore.pinnedProjects.push(projectId);
	}
};

const clearRecents = () => {
	SettingsService.ClearRecentProject().then(() => {
		projectStore.recentProjects = [];
		notificationStore.addNotification(t('components.projectList.recentProjectsCleared'), t('components.projectList.recentProjectsCleared'), "success");
	});
};

const projectListRef = ref(null);
const showTopGradient = ref(false);
const showBottomGradient = ref(false);
const showCenterGradient = ref(false);


onMounted(() => {
  handleScroll();
});

const handleScroll = () => {
  if (projectListRef.value) {
    const element = projectListRef.value;
    const hasTopOverflow = element.scrollTop > 0;
    
    const hasBottomOverflow = element.scrollTop + element.clientHeight < element.scrollHeight - 10;
    
    showTopGradient.value = hasTopOverflow;
    showBottomGradient.value = hasBottomOverflow;
    
    showCenterGradient.value = hasTopOverflow && hasBottomOverflow;

    dockScrollTick.value++;
  }
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.pinned-indicator{
	display: flex;
	align-items: center;
	/* background-color: hotpink; */
}

.pinned-indicator.clickable {
	cursor: pointer;
}

.pinned-indicator.clickable .tiny-icons {
	opacity: .7;
}

.tiny-icons{
	padding-top: .3rem;
	height: 20px;
	opacity: .3;
}

.alert-items {
	z-index: 10000;
	overflow: hidden;
	width: 10px;
	height: 10px;
	background-color: #ecb603;
	border-radius: 5px;
	position: absolute;
	position: fixed;
	display: flex;
	align-items: center;
	justify-content: center;
	top: 15px;
	right: 15px;
	border-radius: 10px;
	padding: 3px;
	font-size: 12px;
	color: white;
	/* outline-offset: -1px; */
	outline: solid 1px rgb(236, 182, 3);
}

.critical-items {
	/* z-index: 10000; */
	overflow: hidden;
	width: min-content;
	max-width: 30px;
	min-width: 5px;
	height: 5px;
	background-color: #ecb603;
	border-radius: 5px;
	position: absolute;
	position: fixed;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 10px;
	padding: 3px;
	font-size: 12px;
	color: white;
	outline: solid 1px #bd2d2d;
	background-color: #bd2d2d;
}

.critical-items-static {
	position: relative;
}


.project-avatar-item {
	border-radius: 8px;
	box-sizing: border-box;
	cursor: pointer;
	display: flex;
	gap: 10px;
	justify-content: flex-start;
	align-items: center;
	padding: 3px 3px;
	width: 100%;
	color: #fff;
	overflow: hidden;
	text-wrap: nowrap;
	box-sizing: border-box;
	height: 35px;
	width: 35px;
	aspect-ratio: 1;
	position: relative;
	justify-content: space-between;
	transform-origin: left center;
	transition: transform 0.18s cubic-bezier(0.34, 1.56, 0.64, 1);
	will-change: transform;
}

.project-avatar-item:active {
	transform: scale(1.1);
	transition: transform 0.08s ease-out;
}

/* While an icon is mirrored in the floating overlay, hide the in-flow glyph so
   the clipped original doesn't show behind the un-clipped overlay copy. */
.dock-magnified-source .project-icon {
	visibility: hidden;
}

.project-avatar-item-text {
	overflow: hidden;
	/* background-color: chocolate; */
	text-overflow: ellipsis;
}

.project-avatar-item {
	/* flex-direction: row; */
}

.project-avatar-item-centered {
	justify-content: center;
}

.project-item-preview-image {
	display: flex;
	box-sizing: border-box;
	align-items: center;
	justify-content: center;
	overflow: hidden;
	height: 100%;
	width: 100%;
	aspect-ratio: 1;
	/* background-color: var(--surface-4); */
	border-radius: 5px;
	filter: blur(0px);
}

.project-icon {
	font-size: x-large;
	font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI Emoji",
		"Segoe UI", "Apple Color Emoji", sans-serif;
	font-size: 1.8rem;
}

.project-avatar-item-active {
	background-color: rgb(173, 173, 173);
	background-color: var(--surface-4);
	color: var(--surface-1);
	outline: var(--transparent-line);
	outline-offset: -1px;
}

.project-avatar-item-active:hover {
	color: var(--surface-1);
}

.project-list {
	display: flex;
	box-sizing: border-box;
	overflow: hidden;
	flex-direction: column;
	gap: 1rem;
	padding: 2px 2px;
	width: 100%;
	height: max-content;
	/* background-color: var(--surface-1); */
	border-radius: 10px;
	transition: all 0.3s cubic-bezier(0.6, 0.05, 0.01, 0.99);
	height: 100%;
	overflow-y: scroll;
	position: relative; 
}

.project-list::-webkit-scrollbar {
	width: 0px;
}

/* Gradient overlay elements */
.top-gradient {
	background: linear-gradient(to bottom, rgba(255, 255, 255, 0.1) 0%, transparent 10%);
}

.bottom-gradient {
	background: linear-gradient(to top, rgba(255, 255, 255, 0.1) 0%, transparent 10%);
}

.center-gradient {
	background: 
        linear-gradient(to bottom, rgba(255, 255, 255, 0.1) 0%, transparent 10%),
        linear-gradient(to top, rgba(255, 255, 255, 0.1) 0%, transparent 10%);

}

.gradient-overlay.visible {
	opacity: 1;
}

.project-list-closed {
	height: 0px;
	padding: 0px .2rem;
}

.project-list-minimized {
	overflow: hidden;

	align-items: center;
}

.compound-list-item {
	/* background-color: hotpink; */
	align-items: center;
	justify-content: center;
}

.menu-divider{
	height: 5px;
	margin-top: 10px;
	margin-bottom: 10px;
}

.chevron-container {
	display: flex;
	box-sizing: border-box;
	overflow: hidden;
	/* flex-direction: column; */
	gap: .4rem;
	align-items: center;
	justify-content: center;
	padding: .2rem;
	width: 100%;
	height: max-content;
	height: 44px;
	/* background-color: darkorange; */
	transition: all 0.3s cubic-bezier(0.6, 0.05, 0.01, 0.99);
}

.chevron-icon {
	display: flex;
	box-sizing: border-box;
	overflow: hidden;
	width: min-content;
	height: min-content;
	/* background-color: teal; */
	transition: all 0.3s ease;
}

/* Floating magnified copies rendered outside the scroll container via Teleport
   so they can grow over the main content instead of being clipped. */
.dock-overlay {
	position: fixed;
	inset: 0;
	z-index: 9999;
	pointer-events: none;
}

.dock-overlay-item {
	position: fixed;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 8px;
	transform-origin: left center;
	will-change: transform;
	pointer-events: none;
}

.dock-overlay-item .project-icon {
	font-size: 1.8rem;
}

.dock-overlay-item .screenshot-thumb {
	height: 100%;
	width: 100%;
	object-fit: cover;
	border-radius: 5px;
}
</style>

