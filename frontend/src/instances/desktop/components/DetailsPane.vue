<template>
  <div ref="detailsPaneRoot" class="details-pane-root" v-stop-propagation
    :class="{ 'details-pane-collapsed': !isVisible }">
    <div class="details-pane-inner">
      <div v-if="isMultipleItems" class="details-pane-content">

        <div v-if="itemsIsEntity" class="pane-parameter-detail">
          {{ itemCounts.entity + ' collections' }}
        </div>

        <div v-if="itemsIsTask" class="pane-parameter-detail">
          {{ itemCounts.task + ' Assets' }}
        </div>

        <div v-if="itemsIsUntracked" class="pane-parameter-detail">
          {{ (itemCounts.untracked_task + itemCounts.untracked_entity) + ' untracked items' }}
        </div>


        <div v-if="showTaskEntityActions || showEntityTaskActions" class="action-bar">
          <ActionButton v-if="activeIsTask" :icon="getAppIcon('dependency')" :label="'Make dependencies of active task'"
            :buttonFunction="makeDependenciesOfActive" />
          <ActionButton v-if="activeIsEntity" :icon="getAppIcon('folder-arrow-in')"
            :label="'Move into active collection'" :buttonFunction="moveIntoFolder" />
        </div>


        <div v-if="onlyTasks" class="action-bar">
          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('shapes')" :label="'Type'" />
            <DropDownBox :items="itemTypes" :selectedItem="''" :onSelect="toggleIsTask" :fixedWidth="true" />
          </div>

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('brush-plus')" :label="'Asset type'" />
            <DropDownBox :items="assetStore.getAssetTypesNames" :selectedItem="taskType" :onSelect="changeTaskType"
              :fixedWidth="true" />
          </div>

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('clock')" :label="'Status'" />
            <DropDownBox :items="projectStatuses" :selectedItem="defaultStatus" :onSelect="setMultipleStatus"
              :fixedWidth="true" />
          </div>
          
          <ActionButton v-if="tasksCanRebuild" :icon="getAppIcon('jigsaw')" :label="'Rebuild Assets'"
            :buttonFunction="revertAllChanges" />
          <ActionButton v-if="tasksModified" :noFilter="true" :icon="getAppIcon('layers-plus-alert')" :label="'Create Checkpoints'"
            :buttonFunction="prepAllCheckpointModal" />
          <ActionButton v-if="tasksModified" :noFilter="true" :icon="getAppIcon('revert-alert')" :label="'Revert Tasks'"
            :buttonFunction="prepResetPopUpModal" />
          <ActionButton :icon="getAppIcon('person-plus')" :label="'Assign assets'"
            @click="prepAssignTask($event)" />
          <ActionButton :icon="getAppIcon('person-minus')" :label="'Unassign assets'"
            :buttonFunction="unassignTasks" />
          <ActionButton v-if="tasksOnDisk" :icon="getAppIcon('broom')" :label="'Free up space'"
            :buttonFunction="prepFreeUpSpacePopUpModal" />
          <ActionButton :icon="getAppIcon('trash')" :label="'Delete Selected assets'"
            :buttonFunction="deleteMultipleTasks" />
        </div>

        <div v-else-if="onlyEntities" class="action-bar">
          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('folder')" :label="'Collection type'" />
            <DropDownBox :items="collectionStore.getCollectionTypesNames" :selectedItem="entityType"
              :onSelect="changeEntityType" :fixedWidth="true" />
          </div>

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('bookmark')" :label="'Library'" />
            <DropDownBox :items="collectionMode" :selectedItem="''" :onSelect="changeIsLibrary" :fixedWidth="true" />
          </div>

          <div class="vertical-flex assignees-search">
            <ActionButton :isInactive="true" :icon="getAppIcon('two-persons')" :label="'Assignees'" />
            <CollaboratorSuggestions :displayEmail="false" :placeholder="placeholder" :allItems="projectUsers"
              @tagAdded="assignCollections"/>
          </div>
          
          <ActionButton :icon="getAppIcon('person-minus')" :label="'Unassign collections'"
            :buttonFunction="unassignCollections" />
          <ActionButton :icon="getAppIcon('jigsaw')" :label="'Rebuild collections'" :buttonFunction="rebuildCollections" />
          <ActionButton :icon="getAppIcon('broom')" :label="'Free up space'"
            :buttonFunction="freeUpCollectionSpacePopUpModal" />
          <ActionButton :icon="getAppIcon('trash')" :label="'Delete collections'"
            :buttonFunction="deleteMultipleEntities" />
        </div>

        
        <div v-else-if="onlyUntrackedAssets || onlyUntrackedCollections" class="action-bar">
          <ActionButton v-if="userStore.canDo('create_task') && onlyUntrackedAssets" :icon="getAppIcon('layers-plus-danger')" :noFilter="true" :label="'Create Checkpoints'" :buttonFunction="prepAllCheckpointModal" />
          <ActionButton :icon="getAppIcon('file-watch')" :label="'Ignore Items'" :buttonFunction="ignoreItems" />
          <ActionButton :icon="getAppIcon('trash')" :label="'Delete Items'" :buttonFunction="deleteMultipleUntrackedTasks" />
        </div>

        <div v-else class="action-bar">
          <ActionButton :icon="getAppIcon('trash')" :label="'Delete Items'" :buttonFunction="deleteMultipleItems" />
        </div>

      </div>
      <div v-else class="details-pane-container absolute-pane">
        <div v-if="!noHeaders.includes(panes.activeModal)" class="pane-header-tabs">
          <PaneHeaderTabs :iconsOnly="false" :useSelected="true" :selectedTab="selectedSettingsContext" :dataTypes="settingsItems" @filter="filterList" />
					<div class="menu-divider"></div>
        </div>
        <component v-for="pane in visiblePanes" :key="pane.name" :is="pane.component" />
      </div>
    </div>
  </div>
</template>

<script setup>
// services
import { CheckpointService, TaskService, EntityService, SyncService } from "@/../bindings/clustta/services";
import { TrashService } from "@/../bindings/clustta/services";
import { FSService } from '@/../bindings/clustta/services/index';
import emitter from '@/lib/mitt';

// imports
import { computed, ref, onMounted, onUnmounted, watchEffect, watch, nextTick } from 'vue';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useProjectStore } from '@/stores/projects';
import { useStatusStore } from '@/stores/status';
import { useUserStore } from '@/stores/users';
import { useTrayStates } from '@/stores/TrayStates';
import { useMenu } from '@/stores/menu';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useDependencyStore } from '@/stores/dependency';
import { useCommonStore } from '@/stores/common';
import { getRelativePath } from '@/lib/pathlib';
import { useIconStore } from '@/stores/icons';
import { useSettingsStore } from '@/stores/settings';

// components
import PaneHeaderTabs from '@/instances/common/components/PaneHeaderTabs.vue';
import ProjectDetails from "@/instances/desktop/panes/ProjectDetails.vue";
import Dependencies from "@/instances/desktop/panes/Dependencies.vue";
import Collaborators from "@/instances/desktop/panes/Collaborators.vue";
import CollectionDetails from "@/instances/desktop/panes/CollectionDetails.vue";
import UntrackedItemDetails from "@/instances/desktop/panes/UntrackedItemDetails.vue";
import AssetDetails from "@/instances/desktop/panes/AssetDetails.vue";
import Checkpoints from "@/instances/desktop/panes/Checkpoints.vue";
import ProjectCheckpoints from "@/instances/desktop/panes/ProjectCheckpoints.vue";
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import CollaboratorSuggestions from '@/instances/common/components/CollaboratorSuggestions.vue';
import utils from "@/services/utils";
import { addIgnoredItem } from '@/lib/untracked';

const panes = usePaneStore();
const iconStore = useIconStore();
const stage = useStageStore();
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const statusStore = useStatusStore();
const projectStore = useProjectStore();
const userStore = useUserStore();
const trayStates = useTrayStates();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const dependencyStore = useDependencyStore();
const commonStore = useCommonStore();
const settings = useSettingsStore();

const paneComponents = {
  projectDetails: ProjectDetails,
  dependencies: Dependencies,
  collaborators: Collaborators,
  collectionDetails: CollectionDetails,
  assetDetails: AssetDetails,
  checkpoints: Checkpoints,
  projectCheckpoints: ProjectCheckpoints,
  untrackedItemDetails: UntrackedItemDetails
};

const props = defineProps({
  isVisible: Boolean
});


let placeholder = 'Search collaborators';

const projectUsers = computed(() => {
  const availableUsers = userStore.getProjectCollaborators
    .map(user => ({
      ...user,
      full_name: `${user.first_name} ${user.last_name}`,
      avatarColor: userStore.userProfileColor(user.id)
    }))
    console.log(availableUsers)
  return availableUsers;
});


const assignCollections = async (user) => {
  stage.operationActive = true;
  const entityIds = stage.markedItems;
  const userId = user.id;

  for (let entityId of entityIds) {
    // Check if the user is already assigned to this entity
    const item = stage.selectedItems.find(item => item.id === entityId);
    if (item && item.assignee_ids && item.assignee_ids.includes(userId)) {
      continue;
    }

    await EntityService.Assign(projectStore.activeProject.uri, entityId, userId)
      .then((data) => {
        // Update the specific item in stage.selectedItems to reflect the assignment
        const itemIndex = stage.selectedItems.findIndex(item => item.id === entityId);
        if (itemIndex !== -1) {
          // Add the user ID to assignee_ids if it's not already there
          if (!stage.selectedItems[itemIndex].assignee_ids.includes(userId)) {
            stage.selectedItems[itemIndex].assignee_ids.push(userId);
          }
        }
        projectStore.refreshActiveProject();
      })
      .catch((error) => {
        notificationStore.errorNotification('Error adding user', error);
        console.error('Error adding user:', error);
      });
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;
}

const noHeaders = ['projectCheckpoints', 'projectDetails']

const isMultipleItems = computed(() => {
  return stage.markedItems.length > 1
});

const numberOfSelectedTasks = computed(() => {
  return stage.markedItems.length;
});

const singleTask = computed(() => {
  numberOfSelectedTasks.value = stage.markedItems.length;
  const isSingleTask = stage.markedItems.length <= 1 && assetStore.selectedTask;
  return isSingleTask
});

// Helper computed properties to count items by type
const itemCounts = computed(() => {
  const counts = {
    entity: 0,
    task: 0,
    untracked_task: 0,
    untracked_entity: 0,
    resource: 0
  };
  
  stage.selectedItems.forEach(item => {
    if (item.type in counts) {
      counts[item.type]++;
    }
  });
  
  return counts;
});

const itemsIsEntity = computed(() => {
  return itemCounts.value.entity > 0
});

const itemsIsTask = computed(() => {
  return itemCounts.value.task > 0
});

const onlyEntities = computed(() => {
  return stage.selectedItems.every((item) => item.type === 'entity')
});

const onlyTasks = computed(() => {
  return stage.selectedItems.every((item) => item.type === 'task')
});

const onlyUntrackedAssets = computed(() => {
  return stage.selectedItems.every((item) => item.type === 'untracked_task')
});

const onlyUntrackedCollections = computed(() => {
  return stage.selectedItems.every((item) => item.type === 'untracked_entity')
});

const onlyUntracked = computed(() => {
  return onlyUntrackedAssets.value || onlyUntrackedCollections.value
});

const itemsIsUntracked = computed(() => {
  return stage.selectedItems.some((item) => item.type === 'untracked_task' || item.type === 'untracked_entity')
});


const assetDetailPanes = ref([
      { name: "Details", tab_name: "assetDetails", icon: "info" },
      { name: "Checkpoints", tab_name: "checkpoints", icon: "layers" },
      { name: "Dependencies", tab_name: "dependencies", icon: "dependency" },
]);

const collectionDetailPanes = ref([
      { name: "Details", tab_name: "collectionDetails", icon: "info" },
      { name: "Collaborators", tab_name: "collaborators", icon: "person" }
]);

const untrackedDetailPanes = ref([
      { name: "Details", tab_name: "untrackedItemDetails", icon: "info" },
]);

const settingsItems = computed(() => {

  const itemType = stage.selectedItem?.type;
  
  if(itemType === 'task'){
    return assetDetailPanes.value
  }else if(itemType === 'entity'){
    return collectionDetailPanes.value;
  } 
  else return untrackedDetailPanes.value; 

});

const selectedSettingsContext = computed(() => {

  let index = activeTabIndex.value;
    if(index < 0){
      index = 0;
    }

	const activePaneContext = settingsItems.value?.find((item) => item.tab_name === panes.activeModal)
  settingsItems.value[index]?.tab_name
	return activePaneContext?.name;

})

const activeTabIndex = ref(0);

const filterList = (selectedTab) => {
	const activePaneContext = settingsItems.value.find((item) => item.name === selectedTab)
  const currentIndex = settingsItems.value.indexOf(activePaneContext);
  activeTabIndex.value = currentIndex;
};

watch(() => settingsItems.value, async () => {
  activeTabIndex.value = 0;
});

const viewCheckpoints = () => {
  filterList('Checkpoints');
}

/////////////////////////////////////////////////

const visiblePanes = computed(() => {
  if (stage.activeStage === 'browser') {
    if (!collectionStore.selectedEntity && !assetStore.selectedTask && !projectStore.selectedUntrackedItem) {
      if (!stage.markedItems.length) {
        if(panes.activeModal !== 'projectCheckpoints'){
          panes.setPaneVisibility('projectDetails', true);
        }
      }
    } else if(stage.selectedItem && stage.markedItems.length === 1){
      
      let index = activeTabIndex.value;
      if(index < 0){
        index = 0;
      }

      panes.setPaneVisibility(settingsItems.value[index]?.tab_name, true);

    }
  }

  return Object.entries(panes.detailPanes)
    .filter(([name, isVisible]) => isVisible)
    .map(([name]) => ({
      name,
      component: paneComponents[name],
    }));
});


// refs
const defaultStatus = ref('TODO');
const multiStatusChange = ref(false);
const multiTypeChange = ref(false);
const itemTypes = ref(['task', 'resource']);
const collectionMode = ref(['basic', 'library']);
const itemType = ref(itemTypes.value[0]);
const taskType = ref(assetStore.getAssetTypesNames[0]);
const entityType = ref(collectionStore.getCollectionTypesNames[0]);
const detailsPaneRoot = ref(null);


watchEffect(() => {
  if (detailsPaneRoot.value) {
    menu.clickOutsideMask = detailsPaneRoot.value;
  }
});

// computed properties
const projectStatuses = computed(() => {
  
  const allStatuses = statusStore.statuses;

  if (!userStore.canDo('set_done_task')) {
    const limitedStatus = ['done', 'retake']
    const availableStatus = allStatuses.filter((item) => !limitedStatus.includes(item.short_name))
    return availableStatus.map((status) => toSentenceCase(status.short_name))
  } else {
    return allStatuses.map((status) => toSentenceCase(status.short_name))
  }
});

const tasksModified = computed(() => {
  const markedItems = stage.markedItems;
  const modifiedAssetsState = assetStore.getModifiedDisplayPaths;
  
  // Check if any of the marked items exist in the modified assets state
  return modifiedAssetsState.some((assetState) => markedItems.includes(assetState.task_id));
});

const tasksOnDisk = computed(() => {
  const displayedTasks = stage.selectedItems.filter((item) => item.type === 'task');
  return displayedTasks.some((item) => item.file_status !== 'rebuildable')
});

const tasksCanRebuild = computed(() => {
  const displayedTasks = stage.selectedItems.filter((item) => item.type === 'task');
  return displayedTasks.some((item) => item.file_status === 'rebuildable')
});

const activeIsTask = computed(() => {
  const activeTaskId = stage.lastSelectedItemId;
  const activeTask = stage.selectedItems.find((item) => item.id === activeTaskId);
  return activeTask?.type === 'task';
});

const activeIsEntity = computed(() => {
  const activeEntityId = stage.lastSelectedItemId;
  const activeEntity = stage.selectedItems.find((item) => item.id === activeEntityId);
  return activeEntity?.type === 'entity';
});

const showTaskEntityActions = computed(() => {
  const hasTasksOrEntities = stage.selectedItems.some(item => 
    item.type === 'task' || item.type === 'entity'
  );
  const lastSelectedIsTask = activeIsTask.value;
  return hasTasksOrEntities && lastSelectedIsTask;
});

const showEntityTaskActions = computed(() => {
  const hasTasksOrEntities = stage.selectedItems.some(item => 
    item.type === 'task' || item.type === 'entity'
  );
  const lastSelectedIsEntity = activeIsEntity.value;
  return hasTasksOrEntities && lastSelectedIsEntity;
});


// methods
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

// Helper function to emit task data updates
const emitTaskUpdates = (taskId, updates) => {
  // Update the corresponding item in stage.selectedItems
  const selectedItemIndex = stage.selectedItems.findIndex(item => item.id === taskId);
  if (selectedItemIndex !== -1) {
    // Apply each update to the selected item
    if (typeof updates === 'object' && !Array.isArray(updates)) {
      // Single update object
      Object.assign(stage.selectedItems[selectedItemIndex], updates);
    } else if (Array.isArray(updates)) {
      // Array of updates
      updates.forEach(update => {
        if (update.property && update.value !== undefined) {
          stage.selectedItems[selectedItemIndex][update.property] = update.value;
        }
      });
    }
  }
  
  const updateData = { itemId: taskId, updates };
  
  // Emit to both Browser and VirtuaItem components
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

const closeModals = () => {
  modals.disableAllModals();
};

const prepAssignTask = (event) => {
  menu.showContextMenu(event, 'assignMenu', true);
};

const setMultipleStatus = async (statusName) => {
  stage.operationActive = true;

  const status = statusStore.statuses.find(item => item.short_name === statusName.toLowerCase())
  // return
  const taskIds = stage.markedItems;
  await assetStore.setMultipleStatus(status, taskIds);
  defaultStatus.value = statusName.toUpperCase();
  stage.operationActive = false;
};

const toSentenceCase = (str) => {
  // Handle empty string
  if (!str) return str;

  // Convert entire string to lowercase first
  const lowercase = str.toLowerCase();

  // Capitalize first letter
  return lowercase.charAt(0).toUpperCase() + lowercase.slice(1);
}

const deleteMultipleItems = async () => {
  panes.setPaneVisibility('projectDetails', true);
  await deleteMultipleEntities();
  await deleteMultipleTasks();
  stage.markedItems = [];
  collectionStore.selectedEntity = null;
};

const deleteMultipleEntities = async () => {
  stage.operationActive = true;
  const entityIds = stage.markedItems;

  for (let entityId of entityIds) {
    await EntityService.DeleteEntity(projectStore.activeProject.uri, entityId, true)
      .then(async (response) => {
        if(onlyEntities.value){
          stage.markedItems = [];
          collectionStore.selectedEntity = null;
        }
      })
      .catch((error) => {
        console.log(error)
        notificationStore.errorNotification("Entities failed to delete.", error)
      });
  };
  clearSelection();
  notificationStore.addNotification("Collections moved to trash.", '', "success", false);
  emitter.emit('refresh-browser')
  stage.operationActive = false;
};

const deleteMultipleUntrackedTasks = async () => {
  stage.operationActive = true;
  
  try {
    for (let untrackedItem of stage.selectedItems) {
      if (untrackedItem.type === 'untracked_task') {
        await FSService.DeleteFile(untrackedItem.file_path);
        projectStore.removeUntrackedTask(untrackedItem.id);
      } else if (untrackedItem.type === 'untracked_entity') {
        await FSService.DeleteFolder(untrackedItem.file_path);
        projectStore.removeUntrackedEntity(untrackedItem.id);
      }
    }
    
    if (onlyUntracked.value) {
      stage.markedItems = [];
      projectStore.selectedUntrackedItem = null;
    }
    
    emitter.emit('refresh-browser');
    notificationStore.addNotification("Untracked items deleted.", '', "success", false);
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification("Failed to delete untracked items.", error);
  }
  
  stage.operationActive = false;
};

const ignoreItems = async () => {
  stage.operationActive = true;
  try {
    for (let untrackedItem of stage.selectedItems) {
      if (untrackedItem.type == "untracked_task") {
        await addIgnoredItem(untrackedItem.task_path)
      } else {
        const untrackedEntity = removeLastSlash(untrackedItem.item_path)
        await addIgnoredItem(untrackedEntity)
      }
    }
    panes.setPaneVisibility('projectDetails', true)
	  clearSelection();
    emitter.emit('refresh-browser');
    notificationStore.addNotification( "Updated ignore list", '', "success", false );
  } catch (error) {
    stage.operationActive = false;
    notificationStore.addNotification(
            "Failed to update ignore list",
            "error"
        );
  }
  stage.operationActive = false;
};

const clearSelection = () => {
	stage.markedItems = [];
	stage.selectedItems = [];
	stage.firstSelectedItemId = '';
	stage.lastSelectedItemId = '';
	assetStore.selectedTask = null;
	collectionStore.selectedEntity = null;
}

const deleteMultipleTasks = async () => {
  stage.operationActive = true;
  const taskIds = stage.markedItems;
  for (let taskId of taskIds) {
    await TaskService.DeleteTask(projectStore.activeProject.uri, taskId, true)
      .then(async (response) => {
        emitter.emit('refresh-browser');
        notificationStore.addNotification("Assets moved to Trash.", '', "success", false);
      })
      .catch((error) => {
        if(onlyTasks.value){
          console.log(error)
          notificationStore.errorNotification("Assets failed to delete.", error)
        }
      });
  }
  stage.operationActive = false;
};

const unassignTasks = async () => {
  let taskIds = stage.markedItems;
  for (const taskId of taskIds) {
    await TaskService.UnassignTask(projectStore.activeProject.uri, taskId)
      .then(async (data) => {
      })
      .catch((error) => {
        console.log(error)
        notificationStore.errorNotification("Error Assigning Task", error)
      });
  }
  emitter.emit('refresh-browser');
  notificationStore.addNotification("Tasks Unssigned Successfully.", "", "success");

};

const freeUpSpace = async () => {
  const selectedTasks = stage.selectedItems.filter(item => item.type === 'task');
  const fileStatus = ['missing', 'rebuildable'];
  const tasksToProcess = selectedTasks.filter(task => !fileStatus.includes(task.file_status));
  
  for (const task of tasksToProcess) {
    let taskPath = task.file_path.replace(/\\/g, '/')
    await FSService.DeleteFile(taskPath)
      .then((response) => {
        task.file_status = 'rebuildable'; 
        assetStore.rebuildableTasksPath.push(task.task_path)
        assetStore.outdatedTasksPath = assetStore.outdatedTasksPath.filter(taskPath => taskPath !== task.task_path)
        assetStore.modifiedTasksPath = assetStore.modifiedTasksPath.filter(taskPath => taskPath !== task.task_path);
        
        // Emit task updates to notify components of file state changes
        emitTaskUpdates(task.id, { file_status: 'rebuildable' });
      })
      .catch((error) => {
        console.error(error); 
      });
  }
  closeModals();
};

const freeUpCollectionSpace = async () => {
  const selectedCollections = stage.selectedItems.filter(item => item.type === 'entity');
  
  for (const collection of selectedCollections) {
    let collectionPath = collection.file_path.replace(/\\/g, '/')
    await FSService.DeleteFolder(collectionPath)
      .then((response) => {
      })
      .catch((error) => {
        console.error(error);
        notificationStore.errorNotification('Error freeing collection space', error);
      });
  }
  closeModals();
  emitter.emit('refresh-browser');
};

const moveIntoFolder = async () => {

  stage.operationActive = true;

  const activeItemId = stage.lastSelectedItemId;
  const selectedItems = stage.selectedItems.filter((item) => item.id !== activeItemId);

  for (const item of selectedItems) {

    if (item.type === 'entity') {
      const entityId = item.id;
      const parentId = activeItemId;
      await changeEntityParent(entityId, parentId);
    } else if (item.type === 'task') {
      const taskId = item.id;
      const entityId = activeItemId;
      await changeTaskEntity(taskId, entityId);
    } else {

      let entity = collectionStore.findCollection(activeItemId)

      await FSService.MakeDirs(entity.file_path)
      let newPath = await FSService.JoinPath(entity.file_path, item.name)
      const untrackedPath = newPath.replace(/^\/+|\/+$/g, "").replace(/\\/g, "/");
      const workingDir = projectStore.activeProject.working_directory.replace(/^\/+|\/+$/g, "").replace(/\\/g, "/");
      console.log(item)
      const itemPath = getRelativePath(workingDir, untrackedPath)

      let entityPath = "";
      const itemPathEntities = itemPath.split("/");
      if (itemPathEntities.length > 1) {
        // Take all elements except the last one
        const pathWithoutLast = itemPathEntities.slice(0, -1);
        entityPath = pathWithoutLast.join("/");
      }

      FSService.Rename(item.file_path, newPath)
        .then(() => {
          if (item.type == "untracked_task") {
            let untrackedTaskIndex = projectStore.untrackedFilesIndex[item.id];
            projectStore.untrackedFiles[untrackedTaskIndex].item_path = itemPath;
            projectStore.untrackedFiles[untrackedTaskIndex].file_path = newPath;
            projectStore.untrackedFiles[untrackedTaskIndex].entity_path = entityPath;
          } else if (item.type == "untracked_entity") {
            let untrackedFolderIndex = projectStore.untrackedFoldersIndex[item.id];
            projectStore.untrackedFolders[untrackedFolderIndex].item_path = itemPath;
            projectStore.untrackedFolders[untrackedFolderIndex].file_path = newPath;
            projectStore.untrackedFolders[untrackedFolderIndex].entity_path = entityPath;
          }
        })

    }
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;
};

const changeEntityParent = async (entityId, parentId) => {

  await EntityService.ChangeEntityParent(projectStore.activeProject.uri, entityId, parentId)
    .then((response) => {
      const successMessage = 'Moved successfully.'
      notificationStore.addNotification(successMessage, "", "success")
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification("Error changing entity parent", error)
    });
};

const changeTaskEntity = async (taskId, entityId) => {
  await TaskService.changeAssetEntity(projectStore.activeProject.uri, taskId, entityId)
    .then((response) => {
      const successMessage = 'Moved successfully.'
      notificationStore.addNotification(successMessage, "", "success")
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification("Error changing task entity", error)
    });
};

const makeDependenciesOfActive = async () => {

  stage.operationActive = true;
  const activeItemId = stage.lastSelectedItemId;
  const selectedItems = stage.selectedItems.filter((item) => item.id !== activeItemId);

  const task = stage.selectedItems.find((item) => item.id === activeItemId);

  for (const item of selectedItems) {
    const itemType = item.type;
    if (itemType === 'entity') {
      const entityId = item.id;
      if (entityId !== task.entity_id) {
        const entityDependencyExists = task.entity_dependencies?.includes(entityId);
        if (!entityDependencyExists) {
          await addEntityDependency(task, entityId);
        }
      }
    } else {
      const dependencyId = item.id;
      const taskDependencyExists = task.dependencies?.includes(dependencyId);
      if (!taskDependencyExists) {
        await addTaskDependency(task, dependencyId);
      }
    }
  }
  emitTaskUpdates(task.id, {
    dependencies: task.dependencies,
    entity_dependencies: task.entity_dependencies
  });

  stage.operationActive = false;
};

const addTaskDependency = async (task, dependencyId) => {

  let dependencyTypeID = dependencyStore.dependency_types.find(item => item.name === "linked").id;

  await TaskService.AddTaskDependency(projectStore.activeProject.uri, task.id, dependencyId, dependencyTypeID)
    .then((response) => {
      if (!task.dependencies) {
        task.dependencies = [];
      }
      task.dependencies.push(dependencyId);
      notificationStore.addNotification("Dependency Added", "", "success");
    })
    .catch((error) => {
      console.log(error)
      notificationStore.errorNotification("Error adding dependencies", error);
    });

};

const addEntityDependency = async (task, dependencyId) => {

  let dependencyTypeID = dependencyStore.dependency_types.find(item => item.name === "linked").id;

  await TaskService.AddEntityDependency(projectStore.activeProject.uri, task.id, dependencyId, dependencyTypeID)
    .then((response) => {
      // Update the local task object with the new entity dependency
      if (!task.entity_dependencies) {
        task.entity_dependencies = [];
      }
      task.entity_dependencies.push(dependencyId);
      notificationStore.addNotification("Dependency Added", "", "success");
    })
    .catch((error) => {
      console.log(error)
      notificationStore.errorNotification("Error adding dependencies", error);
    });

};

const isAllResource = ref('false')

const isAssetTypeSimilar = computed(() => {

  if(!onlyTasks.value) return false

  const selectedTasks = stage.selectedItems;

  const firstValue = selectedTasks[0].is_resource;
  return selectedTasks.every( task => task.is_resource === firstValue);

})

const allAssetTypes = computed(() => {

  if(!onlyTasks.value) return false

  const selectedTasks = stage.selectedItems;

  return !selectedTasks[0].is_resource

})

const toggleIsTask = async (newAssetType) => {

  stage.operationActive = true;
  
  let isResource = newAssetType === 'task';
  const projectPath = projectStore.activeProject.uri;
  const selectedTaskIds = stage.markedItems;

  for (const taskId of selectedTaskIds) {
    await TaskService.ToggleIsTask(projectPath, taskId, isResource)
      .then((data) => {
      })
      .catch((error) => {
        console.error('Error:', error);
      });
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;


};

const changeTaskType = async (taskTypeName) => {
  stage.operationActive = true;

  let newTaskType;
  const taskTypes = assetStore.getAssetTypes;
  newTaskType = taskTypes.find((item) => item.name === taskTypeName);
  taskType.value = taskTypeName;

  const projectPath = projectStore.activeProject.uri;
  const selectedTasksIds = stage.markedItems;

  for (const taskId of selectedTasksIds) {
    await TaskService.ChangeTaskType(projectPath, taskId, newTaskType.id)
      .then((data) => {
      })
      .catch((error) => {
        console.error('Error:', error);
      });
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;

};

const changeEntityType = async (entityTypeName) => {
  stage.operationActive = true;

  let newEntityType;
  const entityTypes = collectionStore.getCollectionTypes;
  newEntityType = entityTypes.find((item) => item.name === entityTypeName);
  entityType.value = entityTypeName;

  const projectPath = projectStore.activeProject.uri;
  const selectedEntitiesId = stage.markedItems

  for (const entityId of selectedEntitiesId) {
    await EntityService.ChangeType(projectPath, entityId, newEntityType.id)
      .then((data) => {
      })
      .catch((error) => {
        console.error('Error:', error);
      });
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;

};

const changeIsLibrary = async (collectionMode) => {

  stage.operationActive = true;
  
  let isLibrary = collectionMode === 'library';
  const projectPath = projectStore.activeProject.uri;
  const selectedCollectionIds = stage.markedItems;

  for (const collectionId of selectedCollectionIds) {
    await EntityService.ChangeIsLibrary(projectPath, collectionId, isLibrary)
      .then((data) => {

      })
      .catch((error) => {
        console.error('Error:', error);
      });
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;

};

const unassignCollections = async () => {

  stage.operationActive = true;

  for (const collection of stage.selectedItems) {
    const currentAssigneeIds = collection.assignee_ids || [];
    
    for (const assigneeId of currentAssigneeIds) {
      
        await EntityService.Unassign(projectStore.activeProject.uri, collection.id, assigneeId)
          .then((data) => {
            const itemIndex = stage.selectedItems.findIndex(item => item.id === collection.id);
            if (itemIndex !== -1) {
              stage.selectedItems[itemIndex].assignee_ids = stage.selectedItems[itemIndex].assignee_ids.filter(id => id !== assigneeId);
            }
          })
          .catch((error) => {
            notificationStore.errorNotification('Error removing user', error);
            console.error('Error removing user:', error);
          });
    }
  }

  emitter.emit('refresh-browser');
  stage.operationActive = false;

};

const revertAllChanges = async () => {
  modals.setModalVisibility('popUpModal', false);
  await CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, stage.markedItems)
    .then((response) => {
      // Update each reverted task in stage.selectedItems and emit updates
      const revertedTasks = stage.selectedItems.filter(item => item.type === 'task');
      for (const task of revertedTasks) {
        // Update the task's file status to reflect the revert
        task.file_status = 'normal';
        
        // Emit task updates to notify components of the state change
        emitTaskUpdates(task.id, { file_status: 'normal' });
      }
    })
    .catch((error) => {
      notificationStore.errorNotification("Error Revering Tasks", error)
      console.error(error);
    });
};

const rebuildCollections = async () => {
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  
  try {
    // Get all selected collection entity IDs and join them with commas for batch processing
    const entityIds = stage.markedItems;
    const entityIdsString = entityIds.join(',');
    
    await EntityService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, entityIdsString)
      .then((data) => {
        assetStore.refreshEntityFilesStatus();
        notificationStore.addNotification(`${entityIds.length} collection(s) rebuilt successfully`, '', "success", false);
      })
      .catch((error) => {
        console.error('Error rebuilding collections:', error);
        notificationStore.errorNotification('Error rebuilding collections', error);
      });
  } catch (error) {
    console.error('Error rebuilding collections:', error);
    notificationStore.errorNotification('Error rebuilding collections', error);
  } finally {
    emitter.emit('refresh-browser');
    notificationStore.canCancel = false;
  }
};

const prepAllCheckpointModal = () => {
  trayStates.createMultipleCheckpoints = false;
  modals.setModalVisibility('createMultipleCheckpointsModal', true);
};

const prepFreeUpSpacePopUpModal = () => {
  trayStates.popUpModalTitle = "Free Up Task Space";
  trayStates.popUpModalMessage = "Are you sure you want to delete these task working files? This will permanently remove all uncheckpointed resources and all task outputs. Please confirm if you wish to proceed.";
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpSpace;
  modals.setModalVisibility('popUpModal', true);
};

const freeUpCollectionSpacePopUpModal = () => {
  trayStates.popUpModalTitle = "Free Up Collection Space";
  trayStates.popUpModalMessage = "Are you sure you want to delete these Collections? This will permanently remove all untracked files and contents. Please confirm if you wish to proceed.";
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpCollectionSpace;
  modals.setModalVisibility('popUpModal', true);
};

const prepResetPopUpModal = () => {
  trayStates.popUpModalIcon = 'revert'
  trayStates.popUpModalTitle = "Revert Selected tasks";
  trayStates.popUpModalMessage = "Modified tasks will be reverted to their last saved state. Are you sure you want to continue?";
  trayStates.popUpModalFunction = revertAllChanges;
  modals.setModalVisibility('popUpModal', true);
};


onMounted(() => {
  panes.setPaneVisibility('projectDetails', true);
  emitter.on('view-checkpoints', viewCheckpoints);
});

onUnmounted(() => {
  panes.setPaneVisibility('projectDetails', true);
  emitter.off('view-checkpoints', viewCheckpoints);
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.pane-header-tabs{
  padding: .5rem 0;
  padding-bottom: 0;
  width: 100%;
  /* background-color: crimson; */
  display: flex;
  align-items: center;
  justify-content: flex-end;
  flex-direction: column;
  gap: .5rem;
}

.debugger {
  color: var(--white);
}

.details-pane-root {
  padding: .5rem;
  padding-top: .9rem;
  position: relative;
  height: 100%;
  max-width: 600px;
  min-width: 350px;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex: 1 1 50%;
  background-color: var(--steel);
  padding-bottom: 0;
}

.details-pane-inner {
  padding: 1rem;
  color: var(--white);
  position: relative;
  height: 100%;
  max-width: 600px;
  min-width: 250px;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex: 1 1 50%;
  background-color: var(--black);
  border-radius: var(--large-radius)
}

.details-pane-content {
  display: flex;
  width: 100%;
  height: 100%;
  box-sizing: border-box;
  background-color: var(--black);
  flex-direction: column;
  padding: 1rem;
}

.details-pane-collapsed {
  padding: 0px;
  min-width: 0px;
  width: 0px;
  flex: 0 0 0%;
}

.action-bar {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: .6rem;
  width: max-content;
  width: 100%;
  height: max-content;
  padding: .2rem;
  align-items: flex-start;
  box-sizing: border-box;
  overflow: hidden;
}

.action-bar-section {
  display: flex;
  align-items: center;
  gap: .5rem;
  justify-content: space-between;
  width: 100%;
}

.assignees-search {
  width: 100%;
  display: flex;
  gap: .5rem;
  align-items: center;
  justify-content: flex-start;
}

.multi-assign {
  /* background-color: tomato; */
  width: 100%;
  display: flex;
  align-items: flex-start;
  flex-direction: column;
}
</style>