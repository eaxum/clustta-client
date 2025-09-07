<template>
  <div class="tabs-container">
    <div class="tab-bar">
      <TransitionGroup name="tab-list">
        <div v-for="(workspace, index) in workspaceTabs" :key="workspace.name" class="tab"
          :class="{ 'active': isActiveTab(workspace), 'dragging': draggedTabIndex === index, 'right-tab-split': rightTabPosition(index), 'left-tab-split': leftTabPosition(index) }"
          @mousedown="startDrag($event, index)" @click="setWorkspace(workspace.name)" :style="getTabStyle(index)">

          <div class="workspace-tab-meta">
            <img v-if="workspace.icon" :src="workspace.icon" class="tab-favicon" alt="favicon">
            <span class="tab-title">{{ workspace.name }}</span>
            <span v-if="index > 1" v-stop-propagation v-tooltip="'Delete Workspace'"
              @click.stop="deleteWorkspace(workspace.name)" class="workspace-tab-button">
              <img class="small-icons no-cursor" :src="getAppIcon('close')">
            </span>
          </div>

        </div>
      </TransitionGroup>
      
      <span class="workspace-tab-button">
          <ActionButton v-if="addWorkspaceVisible" :icon="getAppIcon('plus-circle')" 
          v-tooltip="withinLimits ? 'Save Workspace' : 'Unable to add new workspaces'"
              :buttonFunction="addWorkspace" />
      </span>
      
      
    </div>
  </div>
</template>

<script setup>
// imports
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { SettingsService } from "@/../bindings/clustta/services";
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'

// state imports
import { useIconStore } from '@/stores/icons';
import { useEntityStore } from '@/stores/entity';
import { useCommonStore } from '@/stores/common';
import { useProjectStore } from '@/stores/projects';
import { useDesktopModalStore } from '@/stores/desktopModals';



// states/stores
const iconStore = useIconStore();
const entityStore = useEntityStore();
const commonStore = useCommonStore();
const projectStore = useProjectStore();
const modals = useDesktopModalStore();

// computed props
const workspaceTabs = computed(() => { return commonStore.workspaces });
const activeWorkspace = computed(() => { return commonStore.activeWorkspace });

const activeWorkspaceIndex = computed(() => {
  const allWorkspaces = commonStore.workspaces;
  const activeWorkspaceName = commonStore.activeWorkspace;
  const activeWorkspace = commonStore.workspaces.find((item) => item.name === activeWorkspaceName);
  const workspaceIndex = allWorkspaces.indexOf(activeWorkspace);
  return workspaceIndex
});

const addWorkspaceVisible = computed(() => {

  const taskFilters = commonStore.taskFilters.length;
  const entityFilters = commonStore.entityFilters.length;
  const resourceFilters = commonStore.resourceFilters.length;
  const isActive = commonStore.showEntities && commonStore.showTasks
    && commonStore.showResources && commonStore.showChildEntities
    && commonStore.showChildTasks && commonStore.showDependencies && !commonStore.onlyAssets;
  const defaultWorkspace = commonStore.activeWorkspace === 'Default' ? true : false;

  const showButton = taskFilters || entityFilters || resourceFilters || !isActive;
  return showButton && defaultWorkspace
});

const withinLimits = computed(() => {
  return workspaceTabs.value.length < 5
})

const getAppIcon = (iconName) => {
    const icon = iconStore.getAppIcon(iconName);
    return icon
};

const isActiveTab = (workspace) => {
  return activeWorkspace.value === workspace.name
};

const rightTabPosition = (index) => {
  return activeWorkspaceIndex.value + 1 < index;
};

const leftTabPosition = (index) => {
  return activeWorkspaceIndex.value > index + 1;
};

const activeTabIndex = ref(0)
const draggedTabIndex = ref(null)
const mouseOffset = ref({ x: 0, y: 0 })
const dragPosition = ref({ x: 0, y: 0 })
const tabRects = ref([]);


const setWorkspace = (workspaceName) => {
  if (workspaceName === 'Default') {
    setDefaultWorkspace();
    emitter.emit('refresh-browser');
    return
  }
  const workspace = commonStore.workspaces.find((item) => item.name === workspaceName);
	commonStore.navigatorMode = false;
	entityStore.navigatedEntity = null;
  commonStore.setActiveWorkspace(workspace);
  console.log(commonStore.taskFilters)
  emitter.emit('refresh-browser');
};

const deleteWorkspace = async (workspaceName) => {
  const allWorkspaces = commonStore.workspaces;
  const workspace = commonStore.workspaces.find((item) => item.name === workspaceName);
  const workspaceIndex = allWorkspaces.indexOf(workspace);
  const newWorkspaceIndex = workspaceIndex - 1;
  const newWorkspaceName = allWorkspaces[newWorkspaceIndex].name;
  await SettingsService.RemoveProjectWorkspace(projectStore.getActiveProject.id, workspaceName);
  commonStore.workspaces = commonStore.workspaces.filter((item) => item !== workspace);
  if (activeWorkspace.value === workspaceName) {
    setWorkspace(newWorkspaceName);
  }
};

const setDefaultWorkspace = () => {
  commonStore.activeWorkspace = 'Default';
  commonStore.resetFilters();
};

// Cache initial tab positions
const updateTabRects = () => {
  const tabElements = document.querySelectorAll('.tab')
  tabRects.value = Array.from(tabElements).map(tab => {
    const rect = tab.getBoundingClientRect()
    return {
      left: rect.left,
      width: rect.width
    }
  })
}

onMounted(() => {
  updateTabRects()
  window.addEventListener('resize', updateTabRects);
})

onUnmounted(() => {
  window.removeEventListener('resize', updateTabRects)
});

const addWorkspace = () => {
  if (!withinLimits.value) return
  modals.setModalVisibility('addWorkspaceModal', true)
};

const startDrag = (event, index) => {
  return
  if (event.button !== 0) return

  const tabElement = event.currentTarget
  const tabRect = tabElement.getBoundingClientRect()

  // Calculate offset from the cursor to the tab's left edge
  mouseOffset.value = {
    x: event.clientX - tabRect.left,
    y: event.clientY - tabRect.top
  }

  draggedTabIndex.value = index
  dragPosition.value = {
    x: event.clientX - mouseOffset.value.x,
    y: tabRect.top
  }

  updateTabRects()

  window.addEventListener('mousemove', handleDrag)
  window.addEventListener('mouseup', stopDrag)

  event.preventDefault()
}

const handleDrag = (event) => {
  if (draggedTabIndex.value === null) return

  // Update position to keep tab anchored to cursor
  dragPosition.value = {
    x: event.clientX - mouseOffset.value.x,
    y: dragPosition.value.y
  }

  // Find the tab position where we should move
  const draggedTabWidth = tabRects.value[draggedTabIndex.value].width
  const dragCenterX = event.clientX

  let newIndex = draggedTabIndex.value

  tabRects.value.forEach((rect, index) => {
    if (index === draggedTabIndex.value) return

    const tabCenter = rect.left + rect.width / 2

    if (index < draggedTabIndex.value && dragCenterX < tabCenter) {
      newIndex = index
    } else if (index > draggedTabIndex.value && dragCenterX > tabCenter) {
      newIndex = index
    }
  })

  if (newIndex !== draggedTabIndex.value) {
    // Reorder tabs
    const tabsCopy = [...tabs.value]
    const [draggedTab] = tabsCopy.splice(draggedTabIndex.value, 1)
    tabsCopy.splice(newIndex, 0, draggedTab)
    tabs.value = tabsCopy

    if (activeTabIndex.value === draggedTabIndex.value) {
      activeTabIndex.value = newIndex
    }
    draggedTabIndex.value = newIndex

    // Update cached positions after reordering
    setTimeout(updateTabRects, 0)
  }
}

const stopDrag = () => {
  draggedTabIndex.value = null
  window.removeEventListener('mousemove', handleDrag)
  window.removeEventListener('mouseup', stopDrag)
  updateTabRects()
}

const getTabStyle = (index) => {
  if (draggedTabIndex.value === index) {
    return {
      position: 'fixed',
      left: `${dragPosition.value.x}px`,
      top: `${dragPosition.value.y}px`,
      width: `${tabRects.value[index]?.width}px`,
      zIndex: 1000,
    }
  }
  return {}
}
</script>

<style scoped>
@import "@/assets/desktop.css";

.tabs-container {
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.tab-bar {
  display: flex;
  padding: 0 .5rem;
  gap: 6px;
  padding: 0 20px;
  height: 100%;
  box-sizing: border-box;
  position: relative;
  align-items: flex-end;
  align-items: center;
}

.tab {
  color: var(--white);
  display: flex;
  align-items: center;
  width: 200px;
  min-width: 100px;
  max-width: 200px;
  height: 32px;
  box-sizing: border-box;
  user-select: none;
  position: relative;
}

/* after active tab */

.right-tab-split::after {
  content: "";
  position: absolute;
  background-color: var(--light-steel);
  left: -4px;
  height: 16px;
  width: 1.5px;
}

.right-tab-split:hover+.tab::after {
  content: "";
  position: absolute;
  background-color: transparent;
  left: -4px;
  height: 16px;
  width: 1.5px;
}

.right-tab-split:hover::after {
  content: "";
  position: absolute;
  background-color: transparent;
  left: -4px;
  height: 16px;
  width: 1.5px;
}

/* before active tab */

.left-tab-split::before {
  content: "";
  position: absolute;
  background-color: var(--light-steel);
  right: -4px;
  height: 16px;
  width: 1.5px;
}

.tab:has(+ .left-tab-split:hover)::before {
  content: "";
  position: absolute;
  background-color: transparent;
  right: -4px;
  height: 16px;
  width: 1.5px;
}

.left-tab-split:hover::before {
  content: "";
  position: absolute;
  background-color: transparent;
  right: -4px;
  height: 16px;
  width: 1.5px;
}

/* end */

.tab:hover {
  color: var(--white);
  background: var(--blue-steel);
  border-radius: var(--normal-radius);
  border: 0px;
}

.workspace-tab-meta {
  display: flex;
  width: 100%;
  box-sizing: border-box;
  height: 100%;
  position: absolute;
  z-index: 1;
  padding: .2rem .5rem;
  padding-left: 1rem;
  align-items: center;
  justify-content: space-between;
  left: -50%;
  transform: translateX(50%);

}

.tab.active {
  color: var(--white);
  border-bottom: none;
  border-radius: 12px 12px 0px 0px;
  height: 100%;
  position: relative;
  background-color: var(--steel);
}

.tab.active::before {
  content: "";
  position: absolute;
  background-color: transparent;
  left: -50px;
  bottom: 0px;
  height: 25px;
  width: 50px;
  border-bottom-right-radius: 16px;
  box-shadow: 25px 0 0 0 var(--steel);
}

.tab.active::after {
  content: "";
  position: absolute;
  background-color: transparent;
  right: -50px;
  bottom: 0px;
  height: 25px;
  width: 50px;
  border-bottom-left-radius: 16px;
  box-shadow: -25px 0 0 0 var(--steel);
}

.tab.dragging {
  cursor: grab;
  opacity: 0.9;
  cursor: grabbing;
  position: fixed;
  pointer-events: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.tab-spacer {
  width: 4px;
  height: 30px;
  background-color: rgb(255, 255, 255);
}


.tab-favicon {
  width: 16px;
  height: 16px;
  margin-right: 8px;
  pointer-events: none;
}

.tab-title {
  flex: 1;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  pointer-events: none;
}

.workspace-tab-button {
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  opacity: 0.3;
  color: var(--white);
  color: black
}

.workspace-tab-button:hover {
  background: rgba(0, 0, 0, 0.1);
  opacity: 1;
}


.workspace-tab-button-disabled:hover {
  opacity: 0.3;
}

.new-tab-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  border-radius: 50%;
  margin-left: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
  color: var(--white);
  opacity: 0.5;
}

.new-tab-btn:hover {
  background: rgba(0, 0, 0, 0.1);
  opacity: 1;
}

.tab-content {
  padding: 16px;
}

.tab-list-move {
  transition: transform 0.2s ease;
}
</style>

