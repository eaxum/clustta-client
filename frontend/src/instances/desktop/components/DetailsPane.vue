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
            :buttonFunction="makeDependenciesOfActive" v-tooltip="'Add selected items as dependencies of the active task'" />
          <ActionButton v-if="activeIsEntity" :icon="getAppIcon('folder-arrow-in')"
            :label="'Move into active collection'" :buttonFunction="moveIntoFolder" v-tooltip="'Move selected items into the active collection'" />
        </div>


        <div v-if="onlyTasks" class="action-bar">
          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('shapes')" :label="'Type'" />
            <DropDownBox :items="itemTypes" :selectedItem="''" :onSelect="toggleIsTask" :fixedWidth="true" />
          </div>

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('file-plus')" :label="'Asset type'" />
            <DropDownBox :items="assetStore.getAssetTypesNames" :selectedItem="taskType" :onSelect="changeTaskType"
              :fixedWidth="true" />
          </div>

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('clock')" :label="'Status'" />
            <DropDownBox :items="projectStatuses" :selectedItem="defaultStatus" :onSelect="setMultipleStatus"
              :fixedWidth="true" />
          </div>
          
          <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('folder-arrow-in')" :label="'Move to Collection'"
            @click="prepMoveToCollection($event)" v-tooltip="'Move selected assets to a different collection'" />
          <ActionButton v-if="!platformStore.isWeb && tasksCanRebuild" :icon="getAppIcon('jigsaw')" :label="'Rebuild Assets'"
            :buttonFunction="revertAllChanges" v-tooltip="'Download and restore selected assets'" />
          <ActionButton v-if="tasksModified" :noFilter="true" :icon="getAppIcon('layers-plus')" :useAlert="true" :label="'Create Checkpoints'"
            :buttonFunction="prepAllCheckpointModal" v-tooltip="'Save current state of selected assets'" />
          <ActionButton v-if="!platformStore.isWeb && tasksModified" :noFilter="true" :icon="getAppIcon('revert')" :useAlert="true" :label="'Revert Tasks'"
            :buttonFunction="prepResetPopUpModal" v-tooltip="'Revert selected assets to their last checkpoint'" />
          <ActionButton :icon="getAppIcon('person-plus')" :label="'Assign assets'"
            @click="prepAssignTask($event)" v-tooltip="'Assign a collaborator to selected assets'" />
          <ActionButton :icon="getAppIcon('person-minus')" :label="'Unassign assets'"
            :buttonFunction="unassignTasks" v-tooltip="'Remove all assignees from selected assets'" />
          <ActionButton v-if="!platformStore.isWeb && tasksOnDisk" :icon="getAppIcon('broom')" :label="'Free up space'"
            :buttonFunction="prepFreeUpSpacePopUpModal" v-tooltip="'Delete working files to free disk space'" />
          <ActionButton :icon="getAppIcon('trash')" :label="'Delete Selected assets'"
            :buttonFunction="deleteMultipleTasks" v-tooltip="'Move selected assets to trash'" />
        </div>

        <div v-else-if="onlyEntities" class="action-bar">
          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('folder')" :label="'Collection type'" />
            <DropDownBox :items="collectionStore.getCollectionTypesNames" :selectedItem="entityType"
              :onSelect="changeEntityType" :fixedWidth="true" />
          </div>

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('library')" :label="'Library'" />
            <DropDownBox :items="collectionMode" :selectedItem="''" :onSelect="changeIsLibrary" :fixedWidth="true" />
          </div>

          <div class="vertical-flex assignees-search">
            <ActionButton :isInactive="true" :icon="getAppIcon('two-persons')" :label="'Assignees'" />
            <CollaboratorSuggestions :displayEmail="false" :placeholder="placeholder" :allItems="projectUsers"
              @tagAdded="assignCollections"/>
          </div>
          
          <ActionButton :icon="getAppIcon('person-minus')" :label="'Unassign collections'"
            :buttonFunction="unassignCollections" v-tooltip="'Remove all assignees from selected collections'" />
          <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('jigsaw')" :label="'Rebuild collections'" :buttonFunction="rebuildCollections" v-tooltip="'Download and restore all assets in selected collections'" />
          <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('broom')" :label="'Free up space'"
            :buttonFunction="freeUpCollectionSpacePopUpModal" v-tooltip="'Delete collection contents to free disk space'" />
          <ActionButton :icon="getAppIcon('trash')" :label="'Delete collections'"
            :buttonFunction="deleteMultipleEntities" v-tooltip="'Move selected collections to trash'" />
        </div>

        
        <div v-else-if="onlyUntrackedAssets || onlyUntrackedCollections" class="action-bar">
          <ActionButton v-if="userStore.canDo('create_task') && onlyUntrackedAssets" :icon="getAppIcon('layers-plus')" :useDanger="true" :noFilter="true" :label="'Create Checkpoints'" :buttonFunction="prepAllCheckpointModal" v-tooltip="'Track and create checkpoints for selected items'" />
          <ActionButton :icon="getAppIcon('file-watch')" :label="'Ignore Items'" :buttonFunction="ignoreItems" v-tooltip="'Add items to the ignore list'" />
          <ActionButton :icon="getAppIcon('trash')" :label="'Delete Items'" :buttonFunction="deleteMultipleUntrackedTasks" v-tooltip="'Permanently delete selected untracked items'" />
        </div>

        <div v-else class="action-bar">
          <ActionButton :icon="getAppIcon('trash')" :label="'Delete Items'" :buttonFunction="deleteMultipleItems" v-tooltip="'Delete all selected items'" />
        </div>

      </div>
      <div v-else class="details-pane-container absolute-pane">
        <div v-if="!noHeaders.includes(panes.activeModal)" class="pane-header-tabs">
          <PaneHeaderTabs :iconsOnly="false" :useSelected="true" :selectedTab="selectedSettingsContext" :dataTypes="settingsItems" @filter="filterList" />
					<div class="menu-divider"></div>
        </div>
        <component v-for="pane in visiblePanes" :key="pane.name" :is="pane.component" />
      </div>

      <Clipboard />
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, nextTick, onMounted, onUnmounted, ref, watch, watchEffect } from 'vue';
import emitter from '@/lib/mitt';
import { getRelativePath } from '@/lib/pathlib';
import { addIgnoredItem } from '@/lib/untracked';
import utils from "@/services/utils";

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import AssetDetails from "@/instances/desktop/panes/AssetDetails.vue";
import Checkpoints from "@/instances/desktop/panes/Checkpoints.vue";
import Clipboard from '@/instances/desktop/components/Clipboard.vue';
import CollaboratorSuggestions from '@/instances/common/components/CollaboratorSuggestions.vue';
import Collaborators from "@/instances/desktop/panes/Collaborators.vue";
import CollectionDetails from "@/instances/desktop/panes/CollectionDetails.vue";
import Console from "@/instances/desktop/panes/Console.vue";
import Dependencies from "@/instances/desktop/panes/Dependencies.vue";
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import PaneHeaderTabs from '@/instances/common/components/PaneHeaderTabs.vue';
import ProjectCheckpoints from "@/instances/desktop/panes/ProjectCheckpoints.vue";
import ProjectDetails from "@/instances/desktop/panes/ProjectDetails.vue";
import UntrackedItemDetails from "@/instances/desktop/panes/UntrackedItemDetails.vue";

// services
import { AssetService, CheckpointService, CollectionService, FSService, SyncService, TrashService } from "@/services";

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useDependencyStore } from '@/stores/dependency';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { usePaneStore } from '@/stores/panes';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';
import { useSettingsStore } from '@/stores/settings';
import { useStageStore } from '@/stores/stages';
import { useStatusStore } from '@/stores/status';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const dependencyStore = useDependencyStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const panes = usePaneStore();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const settings = useSettingsStore();
const stage = useStageStore();
const statusStore = useStatusStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

// props
const props = defineProps({
  isVisible: Boolean
});

// constants
const collectionMode = ['basic', 'library'];
const itemTypes = ['task', 'resource'];
const noHeaders = [];
const placeholder = 'Search collaborators';

const assetDetailPanes = [
  { name: "Details", tab_name: "assetDetails", icon: "info" },
  { name: "Checkpoints", tab_name: "checkpoints", icon: "layers" },
  { name: "Dependencies", tab_name: "dependencies", icon: "dependency" },
  { name: "Console", tab_name: "console", icon: "console" },
];

const collectionDetailPanes = [
  { name: "Details", tab_name: "collectionDetails", icon: "info" },
  { name: "Console", tab_name: "console", icon: "console" }
];

const linkDetailPanes = [
  { name: "Details", tab_name: "assetDetails", icon: "info" },
  { name: "Dependencies", tab_name: "dependencies", icon: "dependency" },
];

const paneComponents = {
  assetDetails: AssetDetails,
  checkpoints: Checkpoints,
  collaborators: Collaborators,
  collectionDetails: CollectionDetails,
  console: Console,
  dependencies: Dependencies,
  projectCheckpoints: ProjectCheckpoints,
  projectDetails: ProjectDetails,
  untrackedItemDetails: UntrackedItemDetails
};

const projectDetailPanes = [
  { name: "Details", tab_name: "projectDetails", icon: "info" },
  { name: "Checkpoints", tab_name: "projectCheckpoints", icon: "layers" },
  { name: "Collaborators", tab_name: "collaborators", icon: "person" },
  { name: "Console", tab_name: "console", icon: "console" }
];

const untrackedDetailPanes = [
  { name: "Details", tab_name: "untrackedItemDetails", icon: "info" },
];

// refs
const activeTabIndex = ref(0);
const defaultStatus = ref('TODO');
const detailsPaneRoot = ref(null);
const entityType = ref(collectionStore.getCollectionTypesNames[0]);
const taskType = ref(assetStore.getAssetTypesNames[0]);

// computed properties
const activeIsEntity = computed(() => {
  const activeEntity = stage.selectedItems.find((item) => item.id === stage.lastSelectedItemId);
  return activeEntity?.type === 'entity';
});

const activeIsTask = computed(() => {
  const activeTask = stage.selectedItems.find((item) => item.id === stage.lastSelectedItemId);
  return activeTask?.type === 'task';
});

const isMultipleItems = computed(() => stage.markedItems.length > 1);

const itemCounts = computed(() => {
  const counts = { entity: 0, task: 0, untracked_task: 0, untracked_entity: 0, resource: 0 };
  stage.selectedItems.forEach(item => { if (item.type in counts) counts[item.type]++; });
  return counts;
});

const itemsIsEntity = computed(() => itemCounts.value.entity > 0);

const itemsIsTask = computed(() => itemCounts.value.task > 0);

const itemsIsUntracked = computed(() => stage.selectedItems.some((item) => item.type === 'untracked_task' || item.type === 'untracked_entity'));

const onlyEntities = computed(() => stage.selectedItems.every((item) => item.type === 'entity'));

const onlyTasks = computed(() => stage.selectedItems.every((item) => item.type === 'task'));

const onlyUntracked = computed(() => onlyUntrackedAssets.value || onlyUntrackedCollections.value);

// Determines whether the "Move to Collection" button should be shown.
const hasCollections = computed(() => {
  const collectionsExist = collectionStore.collections.length > 0;
  const selectedTasksHaveParent = stage.selectedItems.some(item => item.type === 'task' && item.parent_id);
  return collectionsExist || selectedTasksHaveParent;
});

const onlyUntrackedAssets = computed(() => stage.selectedItems.every((item) => item.type === 'untracked_task'));

const onlyUntrackedCollections = computed(() => stage.selectedItems.every((item) => item.type === 'untracked_entity'));

const projectStatuses = computed(() => {
  const allStatuses = statusStore.statuses;
  if (!userStore.canDo('set_done_task')) {
    const limitedStatus = ['done', 'retake'];
    return allStatuses.filter((item) => !limitedStatus.includes(item.short_name)).map((status) => toSentenceCase(status.short_name));
  }
  return allStatuses.map((status) => toSentenceCase(status.short_name));
});

const projectUsers = computed(() => {
  return userStore.getProjectCollaborators.map(user => ({
    ...user,
    full_name: `${user.first_name} ${user.last_name}`,
    avatarColor: userStore.userProfileColor(user.id)
  }));
});

const selectedSettingsContext = computed(() => {
  let index = activeTabIndex.value < 0 ? 0 : activeTabIndex.value;
  const activePaneContext = settingsItems.value?.find((item) => item.tab_name === panes.activeModal);
  return activePaneContext?.name;
});

const settingsItems = computed(() => {
  const itemType = stage.selectedItem?.type;
  if (!stage.markedItems.length) return projectDetailPanes;
  if (itemType === 'task') return stage.selectedItem?.is_link ? linkDetailPanes : assetDetailPanes;
  if (itemType === 'entity') return collectionDetailPanes;
  return untrackedDetailPanes;
});

const showEntityTaskActions = computed(() => {
  const hasTasksOrEntities = stage.selectedItems.some(item => item.type === 'task' || item.type === 'entity');
  return hasTasksOrEntities && activeIsEntity.value;
});

const showTaskEntityActions = computed(() => {
  const hasTasksOrEntities = stage.selectedItems.some(item => item.type === 'task' || item.type === 'entity');
  return hasTasksOrEntities && activeIsTask.value;
});

const tasksCanRebuild = computed(() => stage.selectedItems.filter((item) => item.type === 'task').some((item) => item.file_status === 'rebuildable'));

const tasksModified = computed(() => {
  const modifiedAssetsState = assetStore.getModifiedDisplayPaths;
  return modifiedAssetsState.some((assetState) => stage.markedItems.includes(assetState.task_id));
});

const tasksOnDisk = computed(() => stage.selectedItems.filter((item) => item.type === 'task').some((item) => item.file_status !== 'rebuildable'));

const visiblePanes = computed(() => {
  if (stage.activeStage === 'browser') {
    if (!collectionStore.selectedCollection && !assetStore.selectedAsset && !projectStore.selectedUntrackedItem) {
      if (!stage.markedItems.length) {
        let index = activeTabIndex.value < 0 ? 0 : activeTabIndex.value;
        const activePane = settingsItems.value[index]?.tab_name || 'projectDetails';
        panes.setPaneVisibility(activePane, true);
      } 
    } else if (stage.selectedItem && stage.markedItems.length === 1) {
      let index = activeTabIndex.value < 0 ? 0 : activeTabIndex.value;
      panes.setPaneVisibility(settingsItems.value[index]?.tab_name, true);
    }
  }
  return Object.entries(panes.detailPanes)
    .filter(([name, isVisible]) => isVisible)
    .map(([name]) => ({ name, component: paneComponents[name] }));
});

// methods

// Adds an entity dependency to a task.
const addEntityDependency = async (task, dependencyId) => {
  let dependencyTypeID = dependencyStore.dependency_types.find(item => item.name === "linked").id;
  await AssetService.AddEntityDependency(projectStore.activeProject.uri, task.id, dependencyId, dependencyTypeID)
    .then(() => {
      if (!task.entity_dependencies) task.entity_dependencies = [];
      task.entity_dependencies.push(dependencyId);
      notificationStore.addNotification("Dependency Added", "", "success");
    })
    .catch((error) => { console.log(error); notificationStore.errorNotification("Error adding dependencies", error); });
};

// Adds a task dependency to a task.
const addTaskDependency = async (task, dependencyId) => {
  let dependencyTypeID = dependencyStore.dependency_types.find(item => item.name === "linked").id;
  await AssetService.AddAssetDependency(projectStore.activeProject.uri, task.id, dependencyId, dependencyTypeID)
    .then(() => {
      if (!task.dependencies) task.dependencies = [];
      task.dependencies.push(dependencyId);
      notificationStore.addNotification("Dependency Added", "", "success");
    })
    .catch((error) => { console.log(error); notificationStore.errorNotification("Error adding dependencies", error); });
};

// Assigns collections to a user.
const assignCollections = async (user) => {
  stage.operationActive = true;
  const entityIds = stage.markedItems;
  const userId = user.id;
  for (let entityId of entityIds) {
    const item = stage.selectedItems.find(item => item.id === entityId);
    if (item && item.assignee_ids && item.assignee_ids.includes(userId)) continue;
    await CollectionService.Assign(projectStore.activeProject.uri, entityId, userId)
      .then(() => {
        const itemIndex = stage.selectedItems.findIndex(item => item.id === entityId);
        if (itemIndex !== -1 && !stage.selectedItems[itemIndex].assignee_ids.includes(userId)) {
          stage.selectedItems[itemIndex].assignee_ids.push(userId);
        }
        projectStore.refreshActiveProject();
      })
      .catch((error) => { notificationStore.errorNotification('Error adding user', error); console.error('Error adding user:', error); });
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;
};

// Changes the parent collection of one or more entities.
const changeEntityParent = async (entityIds, parentId) => {
  await CollectionService.ChangeCollectionParent(projectStore.activeProject.uri, entityIds, parentId)
    .then(() => notificationStore.addNotification('Moved successfully.', "", "success"))
    .catch((error) => { console.error(error); notificationStore.errorNotification("Error changing entity parent", error); });
};

// Changes the type of multiple entities.
const changeEntityType = async (entityTypeName) => {
  stage.operationActive = true;
  const newEntityType = collectionStore.getCollectionTypes.find((item) => item.name === entityTypeName);
  entityType.value = entityTypeName;
  for (const entityId of stage.markedItems) {
    await CollectionService.ChangeType(projectStore.activeProject.uri, entityId, newEntityType.id).catch((error) => console.error('Error:', error));
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;
};

// Changes the library mode of multiple collections.
const changeIsLibrary = async (mode) => {
  stage.operationActive = true;
  let isLibrary = mode === 'library';
  for (const collectionId of stage.markedItems) {
    await CollectionService.ChangeIsLibrary(projectStore.activeProject.uri, collectionId, isLibrary).catch((error) => console.error('Error:', error));
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;
};

// Moves one or more tasks to a different collection.
const changeTaskEntity = async (taskIds, entityId) => {
  await AssetService.ChangeAssetCollection(projectStore.activeProject.uri, taskIds, entityId)
    .then(() => notificationStore.addNotification('Moved successfully.', "", "success"))
    .catch((error) => { console.error(error); notificationStore.errorNotification("Error changing task entity", error); });
};

// Changes the type of multiple tasks.
const changeTaskType = async (taskTypeName) => {
  stage.operationActive = true;
  const newTaskType = assetStore.getAssetTypes.find((item) => item.name === taskTypeName);
  taskType.value = taskTypeName;
  for (const taskId of stage.markedItems) {
    await AssetService.ChangeAssetType(projectStore.activeProject.uri, taskId, newTaskType.id).catch((error) => console.error('Error:', error));
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;
};

// Clears all item selections.
const clearSelection = () => {
  stage.markedItems = [];
  stage.selectedItems = [];
  stage.firstSelectedItemId = '';
  stage.lastSelectedItemId = '';
  assetStore.selectedAsset = null;
  collectionStore.selectedCollection = null;
};

// Closes all modals.
const closeModals = () => modals.disableAllModals();

// Deletes multiple entities.
const deleteMultipleEntities = async () => {
  stage.operationActive = true;
  for (let entityId of stage.markedItems) {
    await CollectionService.DeleteCollection(projectStore.activeProject.uri, entityId, true)
      .then(() => { if (onlyEntities.value) { stage.markedItems = []; collectionStore.selectedCollection = null; } })
      .catch((error) => { console.log(error); notificationStore.errorNotification("Entities failed to delete.", error); });
  }
  clearSelection();
  notificationStore.addNotification("Collections moved to trash.", '', "success", false);
  emitter.emit('refresh-browser');
  stage.operationActive = false;
};

// Deletes multiple items (tasks and entities).
const deleteMultipleItems = async () => {
  panes.setPaneVisibility('projectDetails', true);
  await deleteMultipleEntities();
  await deleteMultipleTasks();
  stage.markedItems = [];
  collectionStore.selectedCollection = null;
};

// Deletes multiple tasks.
const deleteMultipleTasks = async () => {
  stage.operationActive = true;
  for (let taskId of stage.markedItems) {
    await AssetService.DeleteAsset(projectStore.activeProject.uri, taskId, true)
      .then(() => { emitter.emit('refresh-browser'); notificationStore.addNotification("Assets moved to Trash.", '', "success", false); })
      .catch((error) => { if (onlyTasks.value) { console.log(error); notificationStore.errorNotification("Assets failed to delete.", error); } });
  }
  stage.operationActive = false;
};

// Deletes multiple untracked items.
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
    if (onlyUntracked.value) { stage.markedItems = []; projectStore.selectedUntrackedItem = null; }
    emitter.emit('refresh-browser');
    notificationStore.addNotification("Untracked items deleted.", '', "success", false);
  } catch (error) { console.error(error); notificationStore.errorNotification("Failed to delete untracked items.", error); }
  stage.operationActive = false;
};

// Emits item data updates to notify components.
const emitItemUpdates = (taskId, updates) => {
  const selectedItemIndex = stage.selectedItems.findIndex(item => item.id === taskId);
  if (selectedItemIndex !== -1) {
    if (typeof updates === 'object' && !Array.isArray(updates)) {
      Object.assign(stage.selectedItems[selectedItemIndex], updates);
    } else if (Array.isArray(updates)) {
      updates.forEach(update => { if (update.property && update.value !== undefined) stage.selectedItems[selectedItemIndex][update.property] = update.value; });
    }
  }
  const updateData = { itemId: taskId, updates };
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

// Filters the detail pane tabs.
const filterList = (selectedTab) => {
  const activePaneContext = settingsItems.value.find((item) => item.name === selectedTab);
  activeTabIndex.value = settingsItems.value.indexOf(activePaneContext);
  console.log(activeTabIndex.value)  
};

// Frees up collection space by deleting contents.
const freeUpCollectionSpace = async () => {
  const selectedCollections = stage.selectedItems.filter(item => item.type === 'entity');
  for (const collection of selectedCollections) {
    let collectionPath = collection.file_path.replace(/\\/g, '/');
    await FSService.DeleteFolder(collectionPath).catch((error) => { console.error(error); notificationStore.errorNotification('Error freeing collection space', error); });
  }
  closeModals();
  emitter.emit('refresh-browser');
};

// Shows the free up collection space confirmation modal.
const freeUpCollectionSpacePopUpModal = () => {
  trayStates.popUpModalTitle = "Free Up Collection Space";
  trayStates.popUpModalMessage = "Are you sure you want to delete these Collections? This will permanently remove all untracked files and contents. Please confirm if you wish to proceed.";
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpCollectionSpace;
  modals.setModalVisibility('popUpModal', true);
};

// Frees up task space by deleting working files.
const freeUpSpace = async () => {
  const selectedTasks = stage.selectedItems.filter(item => item.type === 'task');
  const fileStatus = ['missing', 'rebuildable'];
  const tasksToProcess = selectedTasks.filter(task => !fileStatus.includes(task.file_status));
  for (const task of tasksToProcess) {
    let taskPath = task.file_path.replace(/\\/g, '/');
    await FSService.DeleteFile(taskPath)
      .then(() => {
        task.file_status = 'rebuildable';
        assetStore.rebuildableAssetsPath.push(task.task_path);
        assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(taskPath => taskPath !== task.task_path);
        assetStore.modifiedAssetsPath = assetStore.modifiedAssetsPath.filter(taskPath => taskPath !== task.task_path);
        emitItemUpdates(task.id, [{ property: 'file_status', value: 'rebuildable' }]);
      })
      .catch((error) => console.error(error));
  }
  closeModals();
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Adds items to the ignore list.
const ignoreItems = async () => {
  stage.operationActive = true;
  try {
    for (let untrackedItem of stage.selectedItems) {
      if (untrackedItem.type == "untracked_task") {
        await addIgnoredItem(untrackedItem.task_path);
      } else {
        const untrackedEntity = removeLastSlash(untrackedItem.item_path);
        await addIgnoredItem(untrackedEntity);
      }
    }
    panes.setPaneVisibility('projectDetails', true);
    clearSelection();
    emitter.emit('refresh-browser');
    notificationStore.addNotification("Updated ignore list", '', "success", false);
  } catch (error) {
    notificationStore.addNotification("Failed to update ignore list", "error");
  }
  stage.operationActive = false;
};

// Makes selected items dependencies of the active task.
const makeDependenciesOfActive = async () => {
  stage.operationActive = true;
  const activeItemId = stage.lastSelectedItemId;
  const selectedItems = stage.selectedItems.filter((item) => item.id !== activeItemId);
  const task = stage.selectedItems.find((item) => item.id === activeItemId);
  for (const item of selectedItems) {
    if (item.type === 'entity') {
      if (item.id !== task.entity_id && !task.entity_dependencies?.includes(item.id)) {
        await addEntityDependency(task, item.id);
      }
    } else {
      if (!task.dependencies?.includes(item.id)) {
        await addTaskDependency(task, item.id);
      }
    }
  }
  emitItemUpdates(task.id, [
    { property: 'dependencies', value: task.dependencies },
    { property: 'entity_dependencies', value: task.entity_dependencies }
  ]);
  stage.operationActive = false;
};

// Moves selected items into the active collection.
const moveIntoFolder = async () => {
  stage.operationActive = true;
  const activeItemId = stage.lastSelectedItemId;
  const selectedItems = stage.selectedItems.filter((item) => item.id !== activeItemId);

  // Collect items by type for batch operations
  const entityIds = [];
  const taskIds = [];
  const untrackedItems = [];

  for (const item of selectedItems) {
    if (item.type === 'entity') entityIds.push(item.id);
    else if (item.type === 'task') taskIds.push(item.id);
    else untrackedItems.push(item);
  }

  // Execute batch operations for tracked items
  if (entityIds.length) await changeEntityParent(entityIds, activeItemId);
  if (taskIds.length) await changeTaskEntity(taskIds, activeItemId);

  // Handle untracked items (need path computation for each)
  if (untrackedItems.length) {
    let entity = collectionStore.findCollection(activeItemId);
    await FSService.MakeDirs(entity.file_path);
    const renameOperations = [];
    const itemUpdates = [];

    for (const item of untrackedItems) {
      let newPath = await FSService.JoinPath(entity.file_path, item.name);
      const untrackedPath = newPath.replace(/^\/+|\/+$/g, "").replace(/\\/g, "/");
      const workingDir = projectStore.activeProject.working_directory.replace(/^\/+|\/+$/g, "").replace(/\\/g, "/");
      const itemPath = getRelativePath(workingDir, untrackedPath);
      let entityPath = "";
      const itemPathEntities = itemPath.split("/");
      if (itemPathEntities.length > 1) entityPath = itemPathEntities.slice(0, -1).join("/");
      renameOperations.push({ oldPath: item.file_path, newPath });
      itemUpdates.push({ item, itemPath, newPath, entityPath });
    }

    await FSService.RenameBatch(JSON.stringify(renameOperations));

    // Update local state after successful rename
    for (const { item, itemPath, newPath, entityPath } of itemUpdates) {
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
    }
  }

  emitter.emit('refresh-browser');
  stage.operationActive = false;
};

// Shows the create checkpoints modal.
const prepAllCheckpointModal = () => {
  trayStates.createMultipleCheckpoints = false;
  modals.setModalVisibility('createMultipleCheckpointsModal', true);
};

// Opens the assign menu.
const prepAssignTask = (event) => menu.showContextMenu(event, 'assignMenu', true);

// Shows the free up task space confirmation modal.
const prepFreeUpSpacePopUpModal = () => {
  trayStates.popUpModalTitle = "Free Up Task Space";
  trayStates.popUpModalMessage = "Are you sure you want to delete these task working files? This will permanently remove all uncheckpointed resources and all task outputs. Please confirm if you wish to proceed.";
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpSpace;
  modals.setModalVisibility('popUpModal', true);
};

// Shows the revert tasks confirmation modal.
const prepResetPopUpModal = () => {
  trayStates.popUpModalIcon = 'revert';
  trayStates.popUpModalTitle = "Revert Selected tasks";
  trayStates.popUpModalMessage = "Modified tasks will be reverted to their last saved state. Are you sure you want to continue?";
  trayStates.popUpModalFunction = revertAllChanges;
  modals.setModalVisibility('popUpModal', true);
};

// Rebuilds multiple collections.
const rebuildCollections = async () => {
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  try {
    const entityIdsString = stage.markedItems.join(',');
    await CollectionService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, entityIdsString)
      .then(() => {
        assetStore.refreshEntityFilesStatus();
        notificationStore.addNotification(`${stage.markedItems.length} collection(s) rebuilt successfully`, '', "success", false);
      })
      .catch((error) => { console.error('Error rebuilding collections:', error); notificationStore.errorNotification('Error rebuilding collections', error); });
  } catch (error) {
    console.error('Error rebuilding collections:', error);
    notificationStore.errorNotification('Error rebuilding collections', error);
  } finally {
    emitter.emit('refresh-browser');
    notificationStore.canCancel = false;
  }
};

// Removes the trailing slash from a path.
const removeLastSlash = (path) => path.replace(/\/+$/, '');

// Prepares and opens the Move to Collection sub-menu for multi-selection.
const prepMoveToCollection = (event) => {
  const selectedTaskIds = stage.markedItems.filter(id => stage.selectedItems.find(item => item.id === id && item.type === 'task'));
  if (selectedTaskIds.length === 0) return;
  const firstTask = stage.selectedItems.find(item => item.id === selectedTaskIds[0] && item.type === 'task');
  menu.subMenuState.selectedAssetIds = selectedTaskIds;
  menu.subMenuState.startingEntityId = firstTask?.parent_id || '';
  menu.position = { x: event.clientX, y: event.clientY };
  menu.showSubMenu('move-to-collection');
};

// Reverts all selected tasks to their last checkpoint.
const revertAllChanges = async () => {
  modals.setModalVisibility('popUpModal', false);
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  await CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, stage.markedItems)
    .then(() => {
      const revertedTasks = stage.selectedItems.filter(item => item.type === 'task');
      for (const task of revertedTasks) {
        task.file_status = 'normal';
        emitItemUpdates(task.id, [{ property: 'file_status', value: 'normal' }]);
      }
    })
    .catch((error) => { notificationStore.errorNotification("Error Revering Tasks", error); console.error(error); });
};

// Sets the status of multiple tasks.
const setMultipleStatus = async (statusName) => {
  stage.operationActive = true;
  const status = statusStore.statuses.find(item => item.short_name === statusName.toLowerCase());
  await assetStore.setMultipleStatus(status, stage.markedItems);
  defaultStatus.value = statusName.toUpperCase();
  stage.operationActive = false;
};

// Toggles between task and resource type.
const toggleIsTask = async (newAssetType) => {
  stage.operationActive = true;
  let isResource = newAssetType === 'task';
  for (const taskId of stage.markedItems) {
    await AssetService.ToggleIsTask(projectStore.activeProject.uri, taskId, isResource).catch((error) => console.error('Error:', error));
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;
};

// Converts a string to sentence case.
const toSentenceCase = (str) => {
  if (!str) return str;
  const lowercase = str.toLowerCase();
  return lowercase.charAt(0).toUpperCase() + lowercase.slice(1);
};

// Unassigns all collaborators from collections.
const unassignCollections = async () => {
  stage.operationActive = true;
  for (const collection of stage.selectedItems) {
    const currentAssigneeIds = collection.assignee_ids || [];
    for (const assigneeId of currentAssigneeIds) {
      await CollectionService.Unassign(projectStore.activeProject.uri, collection.id, assigneeId)
        .then(() => {
          const itemIndex = stage.selectedItems.findIndex(item => item.id === collection.id);
          if (itemIndex !== -1) stage.selectedItems[itemIndex].assignee_ids = stage.selectedItems[itemIndex].assignee_ids.filter(id => id !== assigneeId);
        })
        .catch((error) => { notificationStore.errorNotification('Error removing user', error); console.error('Error removing user:', error); });
    }
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;
};

// Unassigns all collaborators from tasks.
const unassignTasks = async () => {
  for (const taskId of stage.markedItems) {
    await AssetService.UnassignAsset(projectStore.activeProject.uri, taskId).catch((error) => { console.log(error); notificationStore.errorNotification("Error Assigning Task", error); });
  }
  emitter.emit('refresh-browser');
  notificationStore.addNotification("Tasks Unssigned Successfully.", "", "success");
};

// Switches to the checkpoints tab.
const viewCheckpoints = () => filterList('Checkpoints');

// watchers
watch(() => settingsItems.value, () => { activeTabIndex.value = 0; });

watchEffect(() => { if (detailsPaneRoot.value) menu.clickOutsideMask = detailsPaneRoot.value; });

// lifecycle hooks
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

.action-bar {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: .6rem;
  width: 100%;
  height: max-content;
  padding: .2rem;
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

.details-pane-content {
  display: flex;
  width: 100%;
  height: 100%;
  box-sizing: border-box;
  flex-direction: column;
  /* padding: 1rem; */
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
  background-color: var(--black-steel);
  border-radius: var(--large-radius);
}

.details-pane-root {
  position: relative;
  height: 100%;
  max-width: 400px;
  min-width: 350px;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex: 1 1 50%;
  border-radius: var(--very-large-radius);
}

.details-pane-collapsed {
  padding: 0px;
  min-width: 0px;
  width: 0px;
  flex: 0 0 0%;
}

.pane-header-tabs {
  padding: .5rem 0;
  padding-bottom: 0;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  flex-direction: column;
  gap: .5rem;
}
</style>