<template>
    <div ref="breadcrumbRoot" class="breadcrumb-root">
        <ActionButton :icon="getAppIcon('home')"
            :allowDeactivate="true" v-tooltip="commonStore.navigatorMode ? 'Home' : ''" :label=" commonStore.navigatorMode ? '' : 'Home'" :buttonFunction="goHome" />
        <ActionButton v-if="commonStore.navigatorMode" :icon="getAppIcon('arrow-back-ramp')"
            :allowDeactivate="true" v-tooltip="'Up a level'" :buttonFunction="goUpALevel" />

        <div ref="breadcrumbWrapper" class="breadcrumb-wrapper">
          <div ref="breadcrumbContainer" class="breadcrumb-container">

            <nav class="nav" v-if="path" ref="breadcrumbContent">
              <ActionButton  @click="toggleOverflowListList()" v-if="showEllipsis" :icon="getAppIcon('dots')" :allowDeactivate="true" />
              <div v-for="(segment, index) in visibleSegments" :key="`${segment}-${index}`" class="breadcrumb-segment">
                <ActionButton v-if="path !== 'Home'" :icon="getAppIcon('forward-slash')" :allowDeactivate="true" :label="segment.split('/').pop()" @click="goToCollection(segment)" />
              </div>
            </nav>

          </div>
        </div>

        
      <ActionButton :icon="getAppIcon('copy')" :showLabel="false" :fullWidth="false" @click="copyDirectoryPath()"
        v-tooltip="'Copy Path'" />
      <ActionButton :icon="getAppIcon('folder-arrow-up-right')" :showLabel="false" :fullWidth="false" @click="revealInExplorer()"
        v-tooltip="'Show in Explorer'" />


    </div>

    <Teleport to="#app">
      <div v-if="displayOverflowItems" :style="{ top: listItemsAnchor + 'px', left: listItemsLeft + 'px' }"  class="breadcrumb-list-container">
        <div class="breadcrumb-instance-container">
          <div v-for="(path, index) in overflowPaths" :key="index" class="breadcrumb-instance" @click="goToCollection(path)">
            <div class="breadcrumb-instance-meta">
              <div>{{ path.split('/').pop() }}</div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

</template>

<script setup>
import { ref, computed, onUnmounted, onMounted, watch, nextTick, onBeforeUnmount } from 'vue';
import { EntityService, ClipboardService, FSService } from '@/../bindings/clustta/services';

const emit = defineEmits(['filter']);

import { useIconStore } from '@/stores/icons';
import { useCommonStore } from '@/stores/common';
import { useEntityStore } from '@/stores/entity';
import { useProjectStore } from '@/stores/projects';
import { useNotificationStore } from '@/stores/notifications';

const iconStore = useIconStore();
const commonStore = useCommonStore();
const entityStore = useEntityStore();
const projectStore = useProjectStore();
const notificationStore = useNotificationStore();

import ActionButton from '@/instances/desktop/components/ActionButton.vue'

const props = defineProps({
  dataTypes: { type: Array, default: () => [] },
  alertItems: { type: Function, default: () => {}},
  useFunctions: { type: Boolean, default: false },
  selectedTab: { type: String, default: '' },
});

const navigatedEntity = computed(() => {
  return entityStore.navigatedEntity;
});

const path = computed(() => {
  if(commonStore.navigatorMode){
    return navigatedEntity.value?.type === 'entity' 
      ? navigatedEntity.value?.entity_path
      : navigatedEntity.value?.item_path;
  } else{
    return 'Home'
  }
});

const segments = computed(() => {
  const pathParts = path.value.split('/').filter(segment => segment.trim() !== '')
  return pathParts.map((_, index) => pathParts.slice(0, index + 1).join('/'))
})

const breadcrumbRoot = ref(null)
const showEllipsis = ref(false)
const visibleSegments = ref([]);
const overflowPaths = ref([]);
const breadcrumbContent = ref(null);
const breadcrumbContainer = ref(null);
const displayOverflowItems = ref(false);

const listItemsAnchor = computed(() => {

  const breadCrumbsRootHeight = breadcrumbRoot.value.getBoundingClientRect().height;
  const breadCrumbsRootGlobalY = breadcrumbRoot.value.getBoundingClientRect().top;
  return breadCrumbsRootGlobalY + breadCrumbsRootHeight + 0;

})

const listItemsLeft = computed(() => {

  const breadcrumbsRootLeft = breadcrumbContent.value?.getBoundingClientRect().left;
  return breadcrumbsRootLeft;

});

const handleClickOutside = () => {
  if (displayOverflowItems.value) {
    displayOverflowItems.value = false;
  };
};

const toggleOverflowListList = () => {
  displayOverflowItems.value = !displayOverflowItems.value;
};

const checkOverflow = async () => {

  if(!path.value) return
  await nextTick()
  
  if (!breadcrumbContainer.value || segments.value.length === 0) return

  const nav = breadcrumbContainer.value.querySelector('.nav')
  const container = breadcrumbContainer.value

  // Helper function to test if a number of segments fits
  const testFit = async (numSegments) => {
    const testSegments = segments.value.slice(-numSegments);

    const hiddenSegments = segments.value.slice(0, segments.value.length - numSegments)
    const needsEllipsis = numSegments < segments.value.length
    
    overflowPaths.value = hiddenSegments
    visibleSegments.value = testSegments
    showEllipsis.value = needsEllipsis
    
    await nextTick()
    return nav.scrollWidth <= container.clientWidth
  }

  // First, test if all segments fit (no ellipsis needed)
  if (await testFit(segments.value.length)) {
    // All segments fit, we're done
    return
  }

  // Binary search to find the maximum that fits
  let left = 1
  let right = segments.value.length - 1 // We know all don't fit
  let bestFit = 1

  while (left <= right) {
    const mid = Math.floor((left + right) / 2)
    
    if (await testFit(mid)) {
      bestFit = mid
      left = mid + 1
    } else {
      right = mid - 1
    }
  }

  // Apply the best fit we found
  await testFit(bestFit)
}

const copyDirectoryPath = async () => {

  const isNavigated = commonStore.navigatorMode;
  let project = projectStore.getActiveProject;


  if(!isNavigated){
    let projectDir = project.working_directory;
    projectDir = projectDir.replace(/\\/g, '/');
    await ClipboardService.WriteText(projectDir);
  } else {

    let path = entityStore.navigatedEntity?.type === 'entity' 
      ? entityStore.navigatedEntity.entity_path
      : entityStore.navigatedEntity.item_path;

    let explorerPath = `${project.working_directory}${path}`
    explorerPath = explorerPath.replace(/\\/g, '/');
    await ClipboardService.WriteText(explorerPath);
  }

  const message = 'Path copied to clipboard';
  notificationStore.addNotification(message, "", "success");
};

const revealInExplorer = async () => {

  const isNavigated = commonStore.navigatorMode;
  let project = projectStore.getActiveProject;

  if(!isNavigated){
    await FSService.MakeDirs(project.working_directory)
    FSService.RevealInExplorer(project.working_directory)
  } else {
    let path = entityStore.navigatedEntity?.type === 'entity' 
      ? entityStore.navigatedEntity.entity_path
      : entityStore.navigatedEntity.item_path;

    const trimmedPath = path.endsWith('/') ? path.slice(0, -1) : path;
    let explorerPath = `${project.working_directory}${trimmedPath}`
    explorerPath = explorerPath.replace(/\\/g, '/');

    await FSService.MakeDirs(explorerPath)
    FSService.RevealInExplorer(explorerPath)
  }
};


let resizeObserver = null

onMounted(() => {

  checkOverflow()
  
  // Use ResizeObserver for better performance
  if (breadcrumbRoot.value) {
    resizeObserver = new ResizeObserver(() => {
      checkOverflow()
    })
    resizeObserver.observe(breadcrumbRoot.value)
  }
  document.addEventListener('click', handleClickOutside);
})

onUnmounted(() => {
  if (resizeObserver) {
    resizeObserver.disconnect()
  }
  document.removeEventListener('click', handleClickOutside);
})

watch(() => path.value, () => {
  if(path.value === 'Home'){
    showEllipsis.value = false;
  }
  checkOverflow()
})

const goToCollection = async (selectedPath) => {

  displayOverflowItems.value = false;

  const clickedPath = `/${selectedPath}/`

  displayOverflowItems.value = false;

  if(clickedPath === path.value) return 
    
    const allUntrackedFolders = projectStore.untrackedFolders;
    
    let currentEntity;
    const navigatedEntity = entityStore.navigatedEntity;
    const navigatedEntityType = navigatedEntity.type;
    if(navigatedEntityType === 'entity'){
      currentEntity = await EntityService.GetEntityByID(projectStore.activeProject.uri, navigatedEntity.parent_id);
    } else {
      console.log(allUntrackedFolders)
      currentEntity = allUntrackedFolders.find((collection) => collection.item_path === clickedPath)
    }
    console.log(currentEntity)

    entityStore.navigatedEntity = currentEntity;
    entityStore.selectedEntity = currentEntity;
};

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const goHome = () => {
	commonStore.navigatorMode = false;
	entityStore.navigatedEntity = null;
};

const goUpALevel = async () => {
	const entity = entityStore.navigatedEntity;
  let parentEntityId = entity.parent_id;

  if(!parentEntityId){
    commonStore.navigatorMode = false;
    return
  }

  const parentEntity = await EntityService.GetEntityByID(projectStore.activeProject.uri, parentEntityId);

	if(parentEntity){
		entityStore.navigatedEntity = parentEntity;
		entityStore.selectedEntity = parentEntity;
	} else {
		commonStore.navigatorMode = false;
	}
};

watch(() => projectStore.activeProject?.uri, async () => {
	commonStore.navigatorMode = false;
  entityStore.navigatedEntity = null;
  entityStore.selectedEntity = null;
  
});

</script>

<style scoped>

.breadcrumb-instance-meta {
  display: flex;
  align-items: center;
  box-sizing: border-box;
  width: 100%;
  height: 100px;
  height: 30px;
  padding: .2rem .5rem;
  gap: 10px;
  /* background-color: palevioletred; */
}

.breadcrumb-list-container {
  position: absolute;
  z-index: 10000;
  top: 46px;
  min-width: 160px;
  width: max-content;
  height: max-content;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  /* background-color: royalblue; */
  /* right: 50%; */
  left: 50px;
  /* transform: translateX(50%); */
  border-radius: var(--small-radius);
  background-color: var(--black);
  color: var(--white);
  outline: var(--transparent-line);
  outline-offset: -1px;
  overflow: hidden;
  box-sizing: border-box;
  /* padding: .1rem; */
}

.breadcrumb-instance-container {
  width: 100%;
  height: 100%;
  height: max-content;
  gap: .5rem;
  display: flex;
  padding: .3rem;
  box-sizing: border-box;
  overflow: hidden;
  flex-direction: column;
  /* background-color: rebeccapurple; */
}

.breadcrumb-instance {
  overflow: hidden;
  background-color: transparent;
  text-align: center;
  font-size: 14px;
  line-height: 14px;
  background-color: transparent;
  color: var(--white);
  position: relative;
  border-radius: 8px;
  border-radius: var(--small-radius);
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  gap: 5px;
  align-items: center;
  padding: .1rem;
  height: max-content;
  width: max-content;
  min-width: max-content;
  min-height: max-content;
  transition: all 0.3s ease;
  justify-content: space-between;
  width: 100%;
}

.breadcrumb-instance:hover {
  background-color: rgb(121, 121, 121);
  background-color: #ffffff15;
}

.breadcrumb-instance:active {
  background-color: rgb(70, 70, 70);
  background-color: #00000013;
}

.breadcrumb-instance-pressed {
  box-sizing: border-box;
  background-color: rgba(0, 0, 0, 0.216);
  outline: solid 1px var(--white);
  outline-offset: -1px;
}

.breadcrumb-root {
  display: flex;
  align-items: center;
  font-size: 0.875rem;
  font-weight: 500;
  background-color: var(--black-steel);
  /* background-color: forestgreen; */
  border-radius: var(--normal-radius);
  /* width: 100%; */
  /* width: max-content; */
  overflow: hidden;
  /* flex: 1 1 50%; */
  /* flex: 0 1 auto; */
  max-width: 70%;
  min-width: 65%;
  /* gap: .1rem; */
  padding: .2rem;
  overflow: hidden;
  width: 100%;
  /* max-width: 100% */
}

.breadcrumb-root::-webkit-scrollbar {
    display: none;
}

.breadcrumb-wrapper {
  display: flex;
  align-items: center;
  width: 100%;
  overflow: hidden;
}

.breadcrumb-container {
  display: flex;
  align-items: center;
  width: min-content;
  overflow-x: auto;
  scrollbar-width: none;
  border-radius: var(--small-radius);
  justify-content: flex-end;
}

.nav {
  display: flex;
  align-items: center;
  /* gap: 4px; */
  font-size: 14px;
  white-space: nowrap;
}

.breadcrumb-segment {
  display: flex;
  align-items: center;
  width: min-content;
}

.breadcrumb-item {
  padding: 0.25rem 0;
}

.breadcrumb-link {
  background: none;
  border: none;
  padding: 0;
  color: #2563eb;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
}

.breadcrumb-link:hover {
  color: #1d4ed8;
  text-decoration: underline;
}

.chevron {
  display: flex;
  align-items: center;
  color: #9ca3af;
  margin: 0 0.5rem;
}

.home {
  color: #6b7280;
}

</style>


