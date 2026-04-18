<template>
  <div ref="detailsPaneRoot" class="details-pane-root" v-stop-propagation
    :class="{ 'details-pane-collapsed': !isVisible }">


    <div class="details-pane-inner">
      <div v-if="isMultipleItems" class="details-pane-content">

        <div v-if="itemsIsCollection" class="pane-parameter-detail">
          {{ itemCounts.collection + ' ' + $t('components.detailsPane.collections') }}
        </div>

        <div v-if="itemsIsAsset" class="pane-parameter-detail">
          {{ itemCounts.asset + ' ' + $t('components.detailsPane.assets') }}
        </div>

        <div v-if="itemsIsUntracked" class="pane-parameter-detail">
          {{ (itemCounts.untracked_asset + itemCounts.untracked_collection) + ' ' + $t('components.detailsPane.untrackedItems') }}
        </div>


        <div v-if="showAssetCollectionActions || showCollectionAssetActions" class="action-bar">
          <ActionButton v-if="activeIsAsset" :icon="getAppIcon('dependency')" :label="$t('components.detailsPane.makeDependencies')"
            :buttonFunction="makeDependenciesOfActive" v-tooltip="$t('components.detailsPane.makeDependenciesTooltip')" />
          <ActionButton v-if="activeIsCollection" :icon="getAppIcon('folder-arrow-in')"
            :label="$t('components.detailsPane.moveIntoCollection')" :buttonFunction="moveIntoFolder" v-tooltip="$t('components.detailsPane.moveIntoCollectionTooltip')" />
        </div>


        <div v-if="onlyAssets" class="action-bar">
          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('shapes')" :label="$t('components.detailsPane.type')" />
            <DropDownBox :items="itemTypes" :selectedItem="''" :onSelect="toggleIsAsset" :fixedWidth="true" />
          </div>

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('file-plus')" :label="$t('components.detailsPane.assetType')" />
            <DropDownBox :items="assetStore.getAssetTypesNames" :selectedItem="assetType" :onSelect="changeAssetType"
              :fixedWidth="true" />
          </div>

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('clock')" :label="$t('components.detailsPane.status')" />
            <DropDownBox :items="projectStatuses" :selectedItem="defaultStatus" :onSelect="setMultipleStatus"
              :fixedWidth="true" />
          </div>
          
          <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('folder-arrow-in')" :label="$t('components.detailsPane.moveToCollection')"
            @click="prepMoveToCollection($event)" v-tooltip="$t('components.detailsPane.moveToCollectionTooltip')" />
          <ActionButton v-if="!platformStore.isWeb && assetsCanRebuild" :icon="getAppIcon('jigsaw')" :label="$t('components.detailsPane.rebuildAssets')"
            :buttonFunction="revertAllChanges" v-tooltip="$t('components.detailsPane.rebuildAssetsTooltip')" />
          <ActionButton v-if="assetsModified" :noFilter="true" :icon="getAppIcon('plus-stone')" :useAlert="true" :label="$t('components.detailsPane.createCheckpoints')"
            :buttonFunction="prepAllCheckpointModal" v-tooltip="$t('components.detailsPane.createCheckpointsTooltip')" />
          <ActionButton v-if="!platformStore.isWeb && assetsModified" :noFilter="true" :icon="getAppIcon('revert')" :useAlert="true" :label="$t('components.detailsPane.revertAssets')"
            :buttonFunction="prepResetPopUpModal" v-tooltip="$t('components.detailsPane.revertAssetsTooltip')" />
          <ActionButton :icon="getAppIcon('person-plus')" :label="$t('components.detailsPane.assignAssets')"
            @click="prepAssignAsset($event)" v-tooltip="$t('components.detailsPane.assignAssetsTooltip')" />
          <ActionButton :icon="getAppIcon('person-minus')" :label="$t('components.detailsPane.unassignAssets')"
            :buttonFunction="unassignAssets" v-tooltip="$t('components.detailsPane.unassignAssetsTooltip')" />
          <ActionButton v-if="!platformStore.isWeb && assetsOnDisk" :icon="getAppIcon('broom')" :label="$t('components.detailsPane.freeUpSpace')"
            :buttonFunction="prepFreeUpSpacePopUpModal" v-tooltip="$t('components.detailsPane.freeUpSpaceAssetTooltip')" />
          <ActionButton :icon="getAppIcon('trash')" :label="$t('components.detailsPane.deleteSelectedAssets')"
            :buttonFunction="deleteMultipleAssets" v-tooltip="$t('components.detailsPane.deleteSelectedAssetsTooltip')" />
        </div>

        <div v-else-if="onlyCollections" class="action-bar">
          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('folder')" :label="$t('components.detailsPane.collectionType')" />
            <DropDownBox :items="collectionStore.getCollectionTypesNames" :selectedItem="collectionType"
              :onSelect="changeCollectionType" :fixedWidth="true" />
          </div>

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('shared')" :label="$t('components.detailsPane.shared')" />
            <DropDownBox :items="collectionMode" :selectedItem="''" :onSelect="changeIsShared" :fixedWidth="true" />
          </div>

          <div class="vertical-flex assignees-search">
            <ActionButton :isInactive="true" :icon="getAppIcon('two-persons')" :label="$t('components.detailsPane.assignees')" />
            <CollaboratorSuggestions :displayEmail="false" :placeholder="placeholder" :allItems="projectUsers"
              @tagAdded="assignCollections"/>
          </div>
          
          <ActionButton :icon="getAppIcon('person-minus')" :label="$t('components.detailsPane.unassignCollections')"
            :buttonFunction="unassignCollections" v-tooltip="$t('components.detailsPane.unassignCollectionsTooltip')" />
          <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('jigsaw')" :label="$t('components.detailsPane.rebuildCollections')" :buttonFunction="rebuildCollections" v-tooltip="$t('components.detailsPane.rebuildCollectionsTooltip')" />
          <ActionButton v-if="!platformStore.isWeb" :icon="getAppIcon('broom')" :label="$t('components.detailsPane.freeUpSpace')"
            :buttonFunction="freeUpCollectionSpacePopUpModal" v-tooltip="$t('components.detailsPane.freeUpSpaceCollectionTooltip')" />
          <ActionButton :icon="getAppIcon('trash')" :label="$t('components.detailsPane.deleteCollections')"
            :buttonFunction="deleteMultipleCollections" v-tooltip="$t('components.detailsPane.deleteCollectionsTooltip')" />
        </div>

        
        <div v-else-if="onlyUntrackedAssets || onlyUntrackedCollections" class="action-bar">
          <ActionButton v-if="userStore.canDo('create_asset') && onlyUntrackedAssets" :icon="getAppIcon('plus-stone')" :useDanger="true" :noFilter="true" :label="$t('components.detailsPane.createCheckpoints')" :buttonFunction="prepAllCheckpointModal" v-tooltip="$t('components.detailsPane.createCheckpointsUntrackedTooltip')" />
          <ActionButton v-if="squashEnabled" :icon="getAppIcon('squash')" :label="$t('components.detailsPane.squashAssets')" :buttonFunction="prepSquashModal" v-tooltip="$t('components.detailsPane.squashAssetsTooltip')" />
          <ActionButton :icon="getAppIcon('file-watch')" :label="$t('components.detailsPane.ignoreItems')" :buttonFunction="ignoreItems" v-tooltip="$t('components.detailsPane.ignoreItemsTooltip')" />
          <ActionButton :icon="getAppIcon('trash')" :label="$t('components.detailsPane.deleteItems')" :buttonFunction="deleteMultipleUntrackedAssets" v-tooltip="$t('components.detailsPane.deleteItemsTooltip')" />
        </div>

        <div v-else class="action-bar">
          <ActionButton :icon="getAppIcon('trash')" :label="$t('components.detailsPane.deleteItems')" :buttonFunction="deleteMultipleItems" v-tooltip="$t('components.detailsPane.deleteAllItemsTooltip')" />
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
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import { getRelativePath } from '@/lib/pathlib';
import { addIgnoredItem } from '@/lib/untracked';
import { canSquash } from '@/utils/squash';
import utils from "@/services/utils";

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import AssetDetails from "@/instances/desktop/panes/AssetDetails.vue";
import ChangeLog from "@/instances/desktop/panes/ChangeLog.vue";
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
const { t } = useI18n();

// props
const props = defineProps({
  isVisible: Boolean
});

// constants
const collectionMode = ['basic', 'shared'];
const itemTypes = ['asset', 'resource'];
const noHeaders = [];
const placeholder = computed(() => t('components.detailsPane.searchCollaborators'));

const assetDetailPanes = [
  { name: "Details", nameKey: "panes.detailsTab", tab_name: "assetDetails", icon: "info" },
  { name: "Checkpoints", nameKey: "panes.checkpointsTab", tab_name: "checkpoints", icon: "checkpoint-stone" },
  { name: "Dependencies", nameKey: "panes.dependenciesTab", tab_name: "dependencies", icon: "dependency" },
  { name: "Console", nameKey: "panes.consoleTab", tab_name: "console", icon: "console" },
];

const collectionDetailPanes = [
  { name: "Details", nameKey: "panes.detailsTab", tab_name: "collectionDetails", icon: "info" },
  { name: "Console", nameKey: "panes.consoleTab", tab_name: "console", icon: "console" }
];

const linkDetailPanes = [
  { name: "Details", nameKey: "panes.detailsTab", tab_name: "assetDetails", icon: "info" },
  { name: "Dependencies", nameKey: "panes.dependenciesTab", tab_name: "dependencies", icon: "dependency" },
];

const paneComponents = {
  assetDetails: AssetDetails,
  changeLog: ChangeLog,
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
  { name: "Details", nameKey: "panes.detailsTab", tab_name: "projectDetails", icon: "info" },
  { name: "Checkpoints", nameKey: "panes.checkpointsTab", tab_name: "projectCheckpoints", icon: "checkpoint-stone" },
  { name: "Change Log", nameKey: "panes.changeLogTab", tab_name: "changeLog", icon: "revert" },
  { name: "Collaborators", nameKey: "panes.collaboratorsTab", tab_name: "collaborators", icon: "person" },
  { name: "Console", nameKey: "panes.consoleTab", tab_name: "console", icon: "console" }
];

const remoteOnlyTabs = ['changeLog', 'collaborators', 'dependencies'];

const untrackedDetailPanes = [
  { name: "Details", nameKey: "panes.detailsTab", tab_name: "untrackedItemDetails", icon: "info" },
];

// refs
const activeTabIndex = ref(0);
const defaultStatus = ref('TODO');
const detailsPaneRoot = ref(null);
const collectionType = ref(collectionStore.getCollectionTypesNames[0]);
const assetType = ref(assetStore.getAssetTypesNames[0]);

// computed properties
const activeIsCollection = computed(() => {
  const activeCollection = stage.selectedItems.find((item) => item.id === stage.lastSelectedItemId);
  return activeCollection?.type === 'collection';
});

const activeIsAsset = computed(() => {
  const activeAsset = stage.selectedItems.find((item) => item.id === stage.lastSelectedItemId);
  return activeAsset?.type === 'asset';
});

const isMultipleItems = computed(() => stage.markedItems.length > 1);

const itemCounts = computed(() => {
  const counts = { collection: 0, asset: 0, untracked_asset: 0, untracked_collection: 0, resource: 0 };
  stage.selectedItems.forEach(item => { if (item.type in counts) counts[item.type]++; });
  return counts;
});

const itemsIsCollection = computed(() => itemCounts.value.collection > 0);

const itemsIsAsset = computed(() => itemCounts.value.asset > 0);

const itemsIsUntracked = computed(() => stage.selectedItems.some((item) => item.type === 'untracked_asset' || item.type === 'untracked_collection'));

const onlyCollections = computed(() => stage.selectedItems.every((item) => item.type === 'collection'));

const onlyAssets = computed(() => stage.selectedItems.every((item) => item.type === 'asset'));

const onlyUntracked = computed(() => onlyUntrackedAssets.value || onlyUntrackedCollections.value);

// Determines whether the "Move to Collection" button should be shown.
const hasCollections = computed(() => {
  const collectionsExist = collectionStore.collections.length > 0;
  const selectedAssetsHaveParent = stage.selectedItems.some(item => item.type === 'asset' && item.parent_id);
  return collectionsExist || selectedAssetsHaveParent;
});

const onlyUntrackedAssets = computed(() => stage.selectedItems.every((item) => item.type === 'untracked_asset'));

const onlyUntrackedCollections = computed(() => stage.selectedItems.every((item) => item.type === 'untracked_collection'));

const projectStatuses = computed(() => {
  const allStatuses = statusStore.statuses;
  if (!userStore.canDo('set_done_asset')) {
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
  const isRemote = projectStore.activeProject?.has_remote;
  if (!stage.markedItems.length) {
    return isRemote ? projectDetailPanes : projectDetailPanes.filter(p => !remoteOnlyTabs.includes(p.tab_name));
  }
  if (itemType === 'asset') {
    const panes = stage.selectedItem?.is_link ? linkDetailPanes : assetDetailPanes;
    return isRemote ? panes : panes.filter(p => !remoteOnlyTabs.includes(p.tab_name));
  }
  if (itemType === 'collection') return collectionDetailPanes;
  return untrackedDetailPanes;
});

const showCollectionAssetActions = computed(() => {
  const hasAssetsOrCollections = stage.selectedItems.some(item => item.type === 'asset' || item.type === 'collection');
  return hasAssetsOrCollections && activeIsCollection.value;
});

// Determines whether the squash button should be shown.
const squashEnabled = computed(() => {
  if (!userStore.canDo('create_asset')) return false;
  return canSquash(stage.selectedItems).valid;
});

const showAssetCollectionActions = computed(() => {
  const hasAssetsOrCollections = stage.selectedItems.some(item => item.type === 'asset' || item.type === 'collection');
  return hasAssetsOrCollections && activeIsAsset.value;
});

const assetsCanRebuild = computed(() => stage.selectedItems.filter((item) => item.type === 'asset').some((item) => item.file_status === 'rebuildable'));

const assetsModified = computed(() => {
  const modifiedAssetsState = assetStore.getModifiedDisplayPaths;
  return modifiedAssetsState.some((assetState) => stage.markedItems.includes(assetState.asset_id));
});

const assetsOnDisk = computed(() => stage.selectedItems.filter((item) => item.type === 'asset').some((item) => item.file_status !== 'rebuildable'));

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

// Adds an collection dependency to a asset.
const addCollectionDependency = async (asset, dependencyId) => {
  let dependencyTypeID = dependencyStore.dependency_types.find(item => item.name === "linked").id;
  await AssetService.AddCollectionDependency(projectStore.activeProject.uri, asset.id, dependencyId, dependencyTypeID)
    .then(() => {
      if (!asset.collection_dependencies) asset.collection_dependencies = [];
      asset.collection_dependencies.push(dependencyId);
      notificationStore.addNotification(t('components.detailsPane.dependencyAdded'), "", "success");
    })
    .catch((error) => { console.log(error); notificationStore.errorNotification(t('components.detailsPane.errorAddingDependencies'), error); });
};

// Adds a asset dependency to a asset.
const addAssetDependency = async (asset, dependencyId) => {
  let dependencyTypeID = dependencyStore.dependency_types.find(item => item.name === "linked").id;
  await AssetService.AddAssetDependency(projectStore.activeProject.uri, asset.id, dependencyId, dependencyTypeID)
    .then(() => {
      if (!asset.dependencies) asset.dependencies = [];
      asset.dependencies.push(dependencyId);
      notificationStore.addNotification(t('components.detailsPane.dependencyAdded'), "", "success");
    })
    .catch((error) => { console.log(error); notificationStore.errorNotification(t('components.detailsPane.errorAddingDependencies'), error); });
};

// Assigns collections to a user.
const assignCollections = async (user) => {
  stage.operationActive = true;
  const collectionIds = stage.markedItems;
  const userId = user.id;
  for (let collectionId of collectionIds) {
    const item = stage.selectedItems.find(item => item.id === collectionId);
    if (item && item.assignee_ids && item.assignee_ids.includes(userId)) continue;
    await CollectionService.Assign(projectStore.activeProject.uri, collectionId, userId)
      .then(() => {
        const itemIndex = stage.selectedItems.findIndex(item => item.id === collectionId);
        if (itemIndex !== -1 && !stage.selectedItems[itemIndex].assignee_ids.includes(userId)) {
          stage.selectedItems[itemIndex].assignee_ids.push(userId);
        }
        projectStore.refreshActiveProject();
      })
      .catch((error) => { notificationStore.errorNotification(t('components.detailsPane.errorAddingUser'), error); console.error('Error adding user:', error); });
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;
};

// Changes the parent collection of one or more collections.
const changeCollectionParent = async (collectionIds, parentId) => {
  await CollectionService.ChangeCollectionParent(projectStore.activeProject.uri, collectionIds, parentId)
    .then(() => notificationStore.addNotification(t('components.detailsPane.movedSuccessfully'), "", "success"))
    .catch((error) => { console.error(error); notificationStore.errorNotification(t('components.detailsPane.errorChangingParent'), error); });
};

// Changes the type of multiple collections.
const changeCollectionType = async (collectionTypeName) => {
  stage.operationActive = true;
  const newCollectionType = collectionStore.getCollectionTypes.find((item) => item.name === collectionTypeName);
  collectionType.value = collectionTypeName;
  for (const collectionId of stage.markedItems) {
    await CollectionService.ChangeType(projectStore.activeProject.uri, collectionId, newCollectionType.id).catch((error) => console.error('Error:', error));
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;
};

// Changes the shared mode of multiple collections.
const changeIsShared = async (mode) => {
  stage.operationActive = true;
  let isShared = mode === 'shared';
  for (const collectionId of stage.markedItems) {
    await CollectionService.ChangeIsShared(projectStore.activeProject.uri, collectionId, isShared).catch((error) => console.error('Error:', error));
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;
};

// Moves one or more assets to a different collection.
const changeAssetCollection = async (assetIds, collectionId) => {
  await AssetService.ChangeAssetCollection(projectStore.activeProject.uri, assetIds, collectionId)
    .then(() => notificationStore.addNotification(t('components.detailsPane.movedSuccessfully'), "", "success"))
    .catch((error) => { console.error(error); notificationStore.errorNotification(t('components.detailsPane.errorChangingCollection'), error); });
};

// Changes the type of multiple assets.
const changeAssetType = async (assetTypeName) => {
  stage.operationActive = true;
  const newAssetType = assetStore.getAssetTypes.find((item) => item.name === assetTypeName);
  assetType.value = assetTypeName;
  for (const assetId of stage.markedItems) {
    await AssetService.ChangeAssetType(projectStore.activeProject.uri, assetId, newAssetType.id).catch((error) => console.error('Error:', error));
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

// Deletes multiple collections.
const deleteMultipleCollections = async () => {
  stage.operationActive = true;
  for (let collectionId of stage.markedItems) {
    await CollectionService.DeleteCollection(projectStore.activeProject.uri, collectionId, true)
      .then(() => { if (onlyCollections.value) { stage.markedItems = []; collectionStore.selectedCollection = null; } })
      .catch((error) => { console.log(error); notificationStore.errorNotification(t('components.detailsPane.collectionsFailedToDelete'), error); });
  }
  clearSelection();
  notificationStore.addNotification(t('components.detailsPane.collectionsMovedToTrash'), '', "success", false);
  emitter.emit('refresh-browser');
  stage.operationActive = false;
};

// Deletes multiple items (assets and collections).
const deleteMultipleItems = async () => {
  panes.setPaneVisibility('projectDetails', true);
  await deleteMultipleCollections();
  await deleteMultipleAssets();
  stage.markedItems = [];
  collectionStore.selectedCollection = null;
};

// Deletes multiple assets.
const deleteMultipleAssets = async () => {
  stage.operationActive = true;
  for (let assetId of stage.markedItems) {
    await AssetService.DeleteAsset(projectStore.activeProject.uri, assetId, true)
      .then(() => { emitter.emit('refresh-browser'); notificationStore.addNotification(t('components.detailsPane.assetsMovedToTrash'), '', "success", false); })
      .catch((error) => { if (onlyAssets.value) { console.log(error); notificationStore.errorNotification(t('components.detailsPane.assetsFailedToDelete'), error); } });
  }
  stage.operationActive = false;
};

// Deletes multiple untracked items.
const deleteMultipleUntrackedAssets = async () => {
  stage.operationActive = true;
  try {
    for (let untrackedItem of stage.selectedItems) {
      if (untrackedItem.type === 'untracked_asset') {
        await FSService.DeleteFile(untrackedItem.file_path);
        projectStore.removeUntrackedAsset(untrackedItem.id);
      } else if (untrackedItem.type === 'untracked_collection') {
        await FSService.DeleteFolder(untrackedItem.file_path);
        projectStore.removeUntrackedCollection(untrackedItem.id);
      }
    }
    if (onlyUntracked.value) { stage.markedItems = []; projectStore.selectedUntrackedItem = null; }
    emitter.emit('refresh-browser');
    notificationStore.addNotification(t('components.detailsPane.untrackedItemsDeleted'), '', "success", false);
  } catch (error) { console.error(error); notificationStore.errorNotification(t('components.detailsPane.failedToDeleteUntracked'), error); }
  stage.operationActive = false;
};

// Emits item data updates to notify components.
const emitItemUpdates = (assetId, updates) => {
  const selectedItemIndex = stage.selectedItems.findIndex(item => item.id === assetId);
  if (selectedItemIndex !== -1) {
    if (typeof updates === 'object' && !Array.isArray(updates)) {
      Object.assign(stage.selectedItems[selectedItemIndex], updates);
    } else if (Array.isArray(updates)) {
      updates.forEach(update => { if (update.property && update.value !== undefined) stage.selectedItems[selectedItemIndex][update.property] = update.value; });
    }
  }
  const updateData = { itemId: assetId, updates };
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

// Filters the detail pane tabs.
const filterList = (selectedTab) => {
  const activePaneContext = settingsItems.value.find((item) => item.name === selectedTab);
  activeTabIndex.value = settingsItems.value.indexOf(activePaneContext);
};

// Frees up collection space by deleting contents.
const freeUpCollectionSpace = async () => {
  const selectedCollections = stage.selectedItems.filter(item => item.type === 'collection');
  for (const collection of selectedCollections) {
    let collectionPath = collection.file_path.replace(/\\/g, '/');
    await FSService.DeleteFolder(collectionPath).catch((error) => { console.error(error); notificationStore.errorNotification('Error freeing collection space', error); });
  }
  closeModals();
  emitter.emit('refresh-browser');
};

// Shows the free up collection space confirmation modal.
const freeUpCollectionSpacePopUpModal = () => {
  trayStates.popUpModalTitle = t('components.detailsPane.freeUpCollectionSpaceTitle');
  trayStates.popUpModalMessage = t('components.detailsPane.freeUpCollectionSpaceMessage');
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpCollectionSpace;
  modals.setModalVisibility('popUpModal', true);
};

// Frees up asset space by deleting working files.
const freeUpSpace = async () => {
  const selectedAssets = stage.selectedItems.filter(item => item.type === 'asset');
  const fileStatus = ['missing', 'rebuildable'];
  const assetsToProcess = selectedAssets.filter(asset => !fileStatus.includes(asset.file_status));
  for (const asset of assetsToProcess) {
    let assetPath = asset.file_path.replace(/\\/g, '/');
    await FSService.DeleteFile(assetPath)
      .then(() => {
        asset.file_status = 'rebuildable';
        assetStore.rebuildableAssetsPath.push(asset.asset_path);
        assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(assetPath => assetPath !== asset.asset_path);
        assetStore.modifiedAssetsPath = assetStore.modifiedAssetsPath.filter(assetPath => assetPath !== asset.asset_path);
        emitItemUpdates(asset.id, [{ property: 'file_status', value: 'rebuildable' }]);
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
      if (untrackedItem.type == "untracked_asset") {
        await addIgnoredItem(untrackedItem.asset_path);
      } else {
        const untrackedCollection = removeLastSlash(untrackedItem.item_path);
        await addIgnoredItem(untrackedCollection);
      }
    }
    panes.setPaneVisibility('projectDetails', true);
    clearSelection();
    emitter.emit('refresh-browser');
    notificationStore.addNotification(t('components.detailsPane.updatedIgnoreList'), '', "success", false);
  } catch (error) {
    notificationStore.addNotification(t('components.detailsPane.failedToUpdateIgnoreList'), "error");
  }
  stage.operationActive = false;
};

// Makes selected items dependencies of the active asset.
const makeDependenciesOfActive = async () => {
  stage.operationActive = true;
  const activeItemId = stage.lastSelectedItemId;
  const selectedItems = stage.selectedItems.filter((item) => item.id !== activeItemId);
  const asset = stage.selectedItems.find((item) => item.id === activeItemId);
  for (const item of selectedItems) {
    if (item.type === 'collection') {
      if (item.id !== asset.collection_id && !asset.collection_dependencies?.includes(item.id)) {
        await addCollectionDependency(asset, item.id);
      }
    } else {
      if (!asset.dependencies?.includes(item.id)) {
        await addAssetDependency(asset, item.id);
      }
    }
  }
  emitItemUpdates(asset.id, [
    { property: 'dependencies', value: asset.dependencies },
    { property: 'collection_dependencies', value: asset.collection_dependencies }
  ]);
  stage.operationActive = false;
};

// Moves selected items into the active collection.
const moveIntoFolder = async () => {
  stage.operationActive = true;
  const activeItemId = stage.lastSelectedItemId;
  const selectedItems = stage.selectedItems.filter((item) => item.id !== activeItemId);

  // Collect items by type for batch operations
  const collectionIds = [];
  const assetIds = [];
  const untrackedItems = [];

  for (const item of selectedItems) {
    if (item.type === 'collection') collectionIds.push(item.id);
    else if (item.type === 'asset') assetIds.push(item.id);
    else untrackedItems.push(item);
  }

  // Execute batch operations for tracked items
  if (collectionIds.length) await changeCollectionParent(collectionIds, activeItemId);
  if (assetIds.length) await changeAssetCollection(assetIds, activeItemId);

  // Handle untracked items (need path computation for each)
  if (untrackedItems.length) {
    let collection = collectionStore.findCollection(activeItemId);
    await FSService.MakeDirs(collection.file_path);
    const renameOperations = [];
    const itemUpdates = [];

    for (const item of untrackedItems) {
      let newPath = await FSService.JoinPath(collection.file_path, item.name);
      const untrackedPath = newPath.replace(/^\/+|\/+$/g, "").replace(/\\/g, "/");
      const workingDir = projectStore.activeProject.working_directory.replace(/^\/+|\/+$/g, "").replace(/\\/g, "/");
      const itemPath = getRelativePath(workingDir, untrackedPath);
      let collectionPath = "";
      const itemPathCollections = itemPath.split("/");
      if (itemPathCollections.length > 1) collectionPath = itemPathCollections.slice(0, -1).join("/");
      renameOperations.push({ oldPath: item.file_path, newPath });
      itemUpdates.push({ item, itemPath, newPath, collectionPath });
    }

    await FSService.RenameBatch(JSON.stringify(renameOperations));

    // Update local state after successful rename
    for (const { item, itemPath, newPath, collectionPath } of itemUpdates) {
      if (item.type == "untracked_asset") {
        let untrackedAssetIndex = projectStore.untrackedFilesIndex[item.id];
        projectStore.untrackedFiles[untrackedAssetIndex].item_path = itemPath;
        projectStore.untrackedFiles[untrackedAssetIndex].file_path = newPath;
        projectStore.untrackedFiles[untrackedAssetIndex].collection_path = collectionPath;
      } else if (item.type == "untracked_collection") {
        let untrackedFolderIndex = projectStore.untrackedFoldersIndex[item.id];
        projectStore.untrackedFolders[untrackedFolderIndex].item_path = itemPath;
        projectStore.untrackedFolders[untrackedFolderIndex].file_path = newPath;
        projectStore.untrackedFolders[untrackedFolderIndex].collection_path = collectionPath;
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

// Shows the squash modal.
const prepSquashModal = () => {
  modals.setModalVisibility('squashModal', true);
};

// Opens the assign menu.
const prepAssignAsset = (event) => menu.showContextMenu(event, 'assignMenu', true);

// Shows the free up asset space confirmation modal.
const prepFreeUpSpacePopUpModal = () => {
  trayStates.popUpModalTitle = t('components.detailsPane.freeUpAssetSpaceTitle');
  trayStates.popUpModalMessage = t('components.detailsPane.freeUpAssetSpaceMessage');
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpSpace;
  modals.setModalVisibility('popUpModal', true);
};

// Shows the revert assets confirmation modal.
const prepResetPopUpModal = () => {
  trayStates.popUpModalIcon = 'revert';
  trayStates.popUpModalTitle = t('components.detailsPane.revertSelectedAssetsTitle');
  trayStates.popUpModalMessage = t('components.detailsPane.revertSelectedAssetsMessage');
  trayStates.popUpModalFunction = revertAllChanges;
  modals.setModalVisibility('popUpModal', true);
};

// Rebuilds multiple collections.
const rebuildCollections = async () => {
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  try {
    const collectionIdsString = stage.markedItems.join(',');
    await CollectionService.Rebuild(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, collectionIdsString)
      .then(() => {
        assetStore.refreshCollectionFilesStatus();
        notificationStore.addNotification(t('components.detailsPane.collectionsRebuiltSuccessfully', { count: stage.markedItems.length }), '', "success", false);
      })
      .catch((error) => { console.error('Error rebuilding collections:', error); notificationStore.errorNotification(t('components.detailsPane.errorRebuildingCollections'), error); });
  } catch (error) {
    console.error('Error rebuilding collections:', error);
    notificationStore.errorNotification(t('components.detailsPane.errorRebuildingCollections'), error);
  } finally {
    emitter.emit('refresh-browser');
    notificationStore.canCancel = false;
  }
};

// Removes the trailing slash from a path.
const removeLastSlash = (path) => path.replace(/\/+$/, '');

// Prepares and opens the Move to Collection sub-menu for multi-selection.
const prepMoveToCollection = (event) => {
  const selectedAssetIds = stage.markedItems.filter(id => stage.selectedItems.find(item => item.id === id && item.type === 'asset'));
  if (selectedAssetIds.length === 0) return;
  const firstAsset = stage.selectedItems.find(item => item.id === selectedAssetIds[0] && item.type === 'asset');
  menu.subMenuState.selectedAssetIds = selectedAssetIds;
  menu.subMenuState.startingCollectionId = firstAsset?.parent_id || '';
  menu.position = { x: event.clientX, y: event.clientY };
  menu.showSubMenu('move-to-collection');
};

// Reverts all selected assets to their last checkpoint.
const revertAllChanges = async () => {
  modals.setModalVisibility('popUpModal', false);
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  await CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, stage.markedItems)
    .then(() => {
      const revertedAssets = stage.selectedItems.filter(item => item.type === 'asset');
      for (const asset of revertedAssets) {
        asset.file_status = 'normal';
        emitItemUpdates(asset.id, [{ property: 'file_status', value: 'normal' }]);
      }
    })
    .catch((error) => { notificationStore.errorNotification(t('components.detailsPane.errorRevertingAssets'), error); console.error(error); });
};

// Sets the status of multiple assets.
const setMultipleStatus = async (statusName) => {
  stage.operationActive = true;
  const status = statusStore.statuses.find(item => item.short_name === statusName.toLowerCase());
  await assetStore.setMultipleStatus(status, stage.markedItems);
  defaultStatus.value = statusName.toUpperCase();
  stage.operationActive = false;
};

// Toggles between asset and resource type.
const toggleIsAsset = async (newAssetType) => {
  stage.operationActive = true;
  let isResource = newAssetType === 'asset';
  for (const assetId of stage.markedItems) {
    await AssetService.ToggleIsAsset(projectStore.activeProject.uri, assetId, isResource).catch((error) => console.error('Error:', error));
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
        .catch((error) => { notificationStore.errorNotification(t('components.detailsPane.errorRemovingUser'), error); console.error('Error removing user:', error); });
    }
  }
  emitter.emit('refresh-browser');
  stage.operationActive = false;
};

// Unassigns all collaborators from assets.
const unassignAssets = async () => {
  for (const assetId of stage.markedItems) {
    await AssetService.UnassignAsset(projectStore.activeProject.uri, assetId).catch((error) => { console.log(error); notificationStore.errorNotification(t('components.detailsPane.errorAssigningAsset'), error); });
  }
  emitter.emit('refresh-browser');
  notificationStore.addNotification(t('components.detailsPane.assetsUnassignedSuccessfully'), "", "success");
};

// Switches to the checkpoints tab.
const viewCheckpoints = () => filterList('Checkpoints');

// Switches to the change log tab.
const viewChangeLog = () => filterList('Change Log');

// watchers
watch(() => settingsItems.value, () => { activeTabIndex.value = 0; });

watchEffect(() => { if (detailsPaneRoot.value) menu.clickOutsideMask = detailsPaneRoot.value; });

// lifecycle hooks
onMounted(() => {
  panes.setPaneVisibility('projectDetails', true);
  emitter.on('view-checkpoints', viewCheckpoints);
  emitter.on('view-changelog', viewChangeLog);
});

onUnmounted(() => {
  panes.setPaneVisibility('projectDetails', true);
  emitter.off('view-checkpoints', viewCheckpoints);
  emitter.off('view-changelog', viewChangeLog);
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