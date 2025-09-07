<template>
  <div ref="popUpMenu" class="filter-menu-container">

    <ActionButton v-if="userStore.canDo('pull_chunk')" :icon="getAppIcon('launch')" :showLabel="true" :fullWidth="true"
      label="Open With" :buttonFunction="launchTaskWithCommand" />

    <span v-if="userStore.canDo('pull_chunk')" class="menu-divider"></span>

    <ActionButton v-if="userStore.canDo('update_task')" :icon="getAppIcon('edit')" :showLabel="true" :fullWidth="true"
      label="Rename Asset" :buttonFunction="renameTask" />

    <ActionButton v-if="userStore.canDo('update_task')" :icon="getAppIcon('parameters')" :showLabel="true"
      :fullWidth="true" label="Edit Asset" :buttonFunction="editTask" />

    <ActionButton v-if="userStore.canDo('create_task')" :icon="getAppIcon('copy')" :showLabel="true"
      :fullWidth="true" label="Duplicate Asset" :buttonFunction="duplicateTask" />

    <ActionButton v-if="task.dependencies.length || task.entity_dependencies.length" :icon="getAppIcon('jigsaw')" :showLabel="true"
      :fullWidth="true" label="Build with dependencies" :buttonFunction="buildWithDependencies" />

    <ActionButton v-if="userStore.canDo('manage_dependencies')" :icon="getAppIcon('dependency')" :showLabel="true"
      :fullWidth="true" label="Dependency Graph" :buttonFunction="goToDependencyGraph" />

    <!-- Go to Location -->
    <ActionButton v-if="commonStore.viewSearchQuery || filtersActive" :icon="getAppIcon('magnifying-glass')" :showLabel="true" :fullWidth="true"
      label="Go to Asset" :buttonFunction="goToLocation" />

    <!-- Reveal in Explorer -->
    <span class="horizontal-flex">
      <ActionButton :icon="getAppIcon('folder-arrow-up-right')" :showLabel="true" :fullWidth="true" label="Show in Explorer"
        :buttonFunction="revealInExplorer" />
      <ActionButton :icon="getAppIcon('copy')" :showLabel="false" :fullWidth="false" @click="copyTaskPath('task')"
        v-tooltip="'Copy Path'" />
    </span>

    <!-- Checkpoints -->
    <ActionButton v-if="isTaskModified" :noFilter="true" :icon="getAppIcon('revert-alert')" :showLabel="true" :fullWidth="true"
      label="Revert File" :buttonFunction="revertTask" />

    <span v-if="userStore.canDo('update_task')" class="menu-divider"></span>

    <!-- Free space -->
    <ActionButton :icon="getAppIcon('broom')" v-if="!isNotOnDisk" :showLabel="true" :fullWidth="true"
      label="Free Up space" :buttonFunction="prepFreeUpSpacePopUpModal" />

    <!-- Delete Task -->
    <ActionButton :icon="getAppIcon('trash')" :showLabel="true" :fullWidth="true" label="Delete Asset"
      v-if="userStore.canDo('delete_task')" :buttonFunction="deleteTask" />

  </div>

</template>

<script setup>
import { isValidWeblink } from '@/lib/pointer';
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();
// imports
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';

// services
import { AssetService, CheckpointService, CollectionService } from "@/../bindings/clustta/services";
import { TrashService } from "@/../bindings/clustta/services";

// states/store imports
import { useTrayStates } from '@/stores/TrayStates';
import { useMenu } from '@/stores/menu';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useUserStore } from '@/stores/users';
import { useModalStore } from '@/stores/modals';
import { useAssetStore } from '@/stores/assets';
import { useProjectStore } from '@/stores/projects';
import { useCommonStore } from '@/stores/common';
import { useCollectionStore } from '@/stores/collections';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import { ClipboardService, FSService } from '@/../bindings/clustta/services/index';

// states/stores
const trayStates = useTrayStates();
const userStore = useUserStore();
const menu = useMenu();
const panes = usePaneStore();
const stage = useStageStore();
const modals = useDesktopModalStore();
const modalStore = useModalStore();
const notificationStore = useNotificationStore();
const assetStore = useAssetStore();
const projectStore = useProjectStore();
const commonStore = useCommonStore();
const collectionStore = useCollectionStore();

// refs
const popUpMenu = ref(null);

// computed properties
const task = computed(() => { return assetStore.selectedAsset });
const isNotOnDisk = computed(() => { return task.value?.file_status === 'rebuildable' });
const isTaskModified = computed(() => { return assetStore.selectedAsset.file_status === 'modified' });

const filtersActive = computed(() => {
	const assigneeFilters = commonStore.hasAssignees || commonStore.noAssignees;
	const entityFilters = commonStore.entityFilters.length > 0;
	const taskFilters = commonStore.taskFilters.length > 0;
	const resourceFilters = commonStore.resourceFilters.length > 0;
	const generalFilter = isFilterActive('general');
	return assigneeFilters || entityFilters || taskFilters || resourceFilters || generalFilter
});

const isFilterActive = (filter) => {
	if (filter.includes('general')) {
		const isActive = commonStore.showEntities && commonStore.showTasks
			&& commonStore.showResources && commonStore.showChildEntities
			&& commonStore.showChildTasks && commonStore.showDependencies && !commonStore.onlyAssets;
		return !isActive;
	} else
		if (filter.includes('entity')) {
			return commonStore.entityFilters.some((item) => item.type === filter);
		} else if (filter.includes('assignation')) {
			const assigneeFilters = commonStore.hasAssignees || commonStore.noAssignees;
			const assignationFilters = commonStore.taskFilters.some((item) => item.type === filter);
			return assigneeFilters || assignationFilters;
		} else {
			return commonStore.taskFilters.some((item) => item.type === filter);
		}
};

// props
const props = defineProps({
  top: Number,
  left: Number,
});

const emit = defineEmits(['clicked']);

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const launchTaskWithCommand = async () => {
  let task = assetStore.selectedAsset
  if (task.is_link && isValidWeblink(task.pointer)) {
    Browser.OpenURL(task.pointer)
  } else {
    let file_path = task.pointer ? task.pointer : task.file_path
    if (await FSService.Exists(file_path)) {
      FSService.LaunchFileWith(file_path)
    } else {
      CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [task.id])
        .then(async (response) => {
          let fileStatus = await assetStore.getAssetFileStatus(task)
          task.file_status = fileStatus
          FSService.LaunchFileWith(file_path)
        })
        .catch((error) => {
          console.log(error)
          notificationStore.errorNotification("Error Rebuilding Task", error)
        });
    }
  }
  menu.hideContextMenu();
};

const editTask = () => {
  modals.setModalVisibility('editAssetModal', true);
  menu.hideContextMenu();
};

const duplicateTask = async () => {
  menu.hideContextMenu();
  
  try {

    stage.operationActive = true;
    let selectedTask = assetStore.selectedAsset;
    
		await AssetService.DuplicateTask(projectStore.activeProject.uri, selectedTask.id)
		.then((duplicatedTask) => {
			emitter.emit('refresh-browser')
			assetStore.selectAsset(duplicatedTask);
			stage.selectedItem = duplicatedTask;
			stage.markedItems = [duplicatedTask.id];
			stage.lastSelectedItemId = "";
			stage.firstSelectedItemId = duplicatedTask.id;
      
      stage.operationActive = false;

      notificationStore.addNotification(
        'Asset Duplicated', 
        `Asset duplicated`, 
        'success'
      );
		})
    
  } catch (error) {
    console.error('Error duplicating task:', error);
    notificationStore.errorNotification('Failed to Duplicate Asset', error);
  }
};

const renameTask = () => {
  emitter.emit('renameTask');
  menu.hideContextMenu();
};

const buildWithDependencies = async () => {
  menu.hideContextMenu();
  let taskIds = [task.value.id, ...task.value.dependencies];
  for (let entityId of task.value.entity_dependencies) {
    let entityTasks = assetStore.getEntityAssets(entityId, true);
    for (let entityTask of entityTasks) {
      if (!taskIds.includes(entityTask.id)) {
        taskIds.push(entityTask.id);
      }
    }
  }
  await CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, taskIds)
    .then((response) => {
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      notificationStore.errorNotification("Error Revering Tasks", error);
      console.error(error);
    });
};

const goToDependencyGraph = () => {
  stage.setStageVisibility('dependencies', true);
  menu.hideContextMenu();
};

const goToLocation = async () => {

  commonStore.activeWorkspace = 'Default'
  menu.hideContextMenu();
  
  // Reset search query and filters
  commonStore.viewSearchQuery = '';
  commonStore.resetFilters();
  
  try {
    // Enable navigator mode
    commonStore.navigatorMode = true;
    
    // Get the parent entity of the selected task
    const selectedTask = assetStore.selectedAsset;
    if (selectedTask && selectedTask.entity_id) {
      const parentEntity = await CollectionService.GetEntityByID(projectStore.activeProject.uri, selectedTask.entity_id);
      if (parentEntity) {
        collectionStore.navigatedCollection = parentEntity;
        collectionStore.selectedCollection = parentEntity;
      }
    }
  } catch (error) {
    console.error('Error navigating to location:', error);
    notificationStore.errorNotification('Failed to navigate to location', error);
  }
};

const revealInExplorer = async () => {

  menu.hideContextMenu();
  const taskId = assetStore.selectedAsset.id;

  if(assetStore.selectedAsset.file_status == "rebuildable"){
    await CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [taskId])
    .then( async (response) => {
      assetStore.rebuildableAssetsPath = assetStore.rebuildableAssetsPath.filter(taskPath => taskPath !== task.task_path)
      assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(taskPath => taskPath !== task.task_path);
      emitter.emit('get-project-data')
    })
    .catch((error) => {
      notificationStore.errorNotification("Error downloading Task", error);
      console.error(error);
    });

  } 
  AssetService.RevealTask(projectStore.activeProject.uri, assetStore.selectedAsset.id);
};

const revertTask = async () => {
  let taskId = assetStore.selectedAsset.id;
  CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [taskId])
    .then((response) => {
      assetStore.selectedAsset.file_status = "normal"
    })
    .catch((error) => {
      notificationStore.errorNotification("Failed to Revert Task", error)
    });
  menu.hideContextMenu();
};

// methods
const copyTaskPath = async (pathType) => {
  let task = assetStore.selectedAsset;
  console.log(task)
  let taskPath = task.file_path;
  taskPath = taskPath.replace(/\\/g, '/');
  let taskDir = taskPath.split('/').slice(0, -1).join('/');
  let resourcesFolder = taskDir + '/resources';
  let outputPath = taskDir + '/output';
  if (pathType === 'resources') {
    taskPath = resourcesFolder;
  } else if (pathType === 'output') {
    taskPath = outputPath;
  }
  await ClipboardService.WriteText(taskPath);
  const message = 'Path copied to clipboard';
  notificationStore.addNotification(message, "", "success");
  menu.hideContextMenu();
};

const deleteTask = async () => {
  let taskId = assetStore.selectedAsset.id;
  let longMessage = `Asset of name: ${assetStore.selectedAsset.name} was moved to Trash.`
  panes.setPaneVisibility('projectDetails', true);
  menu.hideContextMenu();
  assetStore.selectedAsset = null;
  AssetService.DeleteTask(projectStore.activeProject.uri, taskId, true)
    .then(async (response) => {
      trayStates.undoItemId = taskId;
      trayStates.undoFunction = undoTaskDelete;
      assetStore.selectedAsset = null;
      stage.markedItems = [];
      projectStore.refreshActiveProject()
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      notificationStore.errorNotification("Asset failed to delete.", error)
    });
  notificationStore.addNotification("Asset moved to Trash.", longMessage, "success", true)

};

const undoTaskDelete = async () => {
  TrashService.Restore(projectStore.activeProject.uri, trayStates.undoItemId, "task")
    .then(async (response) => {
      assetStore.unmarkAssetAsDeleted(trayStates.undoItemId)
    })
    .catch((error) => {
      notificationStore.errorNotification("Error Restoring Item", error)
    });
}

const freeUpSpace = async () => {
  let task = assetStore.selectedAsset;
  let taskDir = task.file_path.replace(/\\/g, '/');
  await FSService.DeleteFile(taskDir)
    .then((response) => {
      task.file_status = 'rebuildable';
      assetStore.rebuildableAssetsPath.push(task.task_path)
      assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(taskPath => taskPath !== task.task_path)
      assetStore.modifiedAssetsPath = assetStore.modifiedAssetsPath.filter(taskPath => taskPath !== task.task_path);
      emitter.emit('refresh-browser')
    })
    .catch((error) => {
      console.error(error);
    });
  modals.disableAllModals();
};

const prepFreeUpSpacePopUpModal = () => {
  trayStates.popUpModalTitle = "Free Up Task Space";
  trayStates.popUpModalMessage = "Are you sure you want to delete this task working files? This will permanently remove all uncheckpointed resources and all task outputs. Please confirm if you wish to proceed.";
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpSpace;
  modals.setModalVisibility('popUpModal', true);
  menu.hideContextMenu();
};

const viewCheckPoints = () => {
  modalStore.triggerMenuItem('popUpMenu', 'CheckPoints');
};



// onMounted hook
onMounted(() => {
  menu.popUpMenuWidth = popUpMenu.value.getBoundingClientRect().width;
  menu.popUpMenu = popUpMenu.value;
});

onBeforeUnmount(() => {
  menu.popUpMenuWidth = popUpMenu.value.getBoundingClientRect().width;
  menu.popUpMenuHeight = popUpMenu.value.getBoundingClientRect().height;

});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/menu.css";

.task-item-menu-container {
  z-index: 10;
  display: flex;
  top: 0;
  left: 0;
  flex-direction: column;
  color: white;
  align-items: center;
  gap: .3rem;
  padding: .6rem;
  box-sizing: border-box;
  width: max-content;
  width: 250px;
  height: max-content;
  border-radius: 16px;
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--light-steel);

}

.task-item-menu-visible {
  /* display: flex; */
  opacity: 1;
  visibility: visible;
}
</style>







